package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// nullSeqSentinel ist der int64-Wert, mit dem `sequence_number IS NULL`
// in COALESCE-basierten Cursor-Vergleichen behandelt wird
// (math.MinInt64) — NULL-Events sortieren strikt VOR gesetzten
// sequence_number-Werten (ADR-0002, API-Kontrakt).
const nullSeqSentinel int64 = -1 << 63

// persistedSchemaVersion ist die einzige Wire-Format-Version, die dieses
// Release annimmt (spec/backend-api-contract.md).
const persistedSchemaVersion = "1.0"

// rowScanner abstrahiert `*sql.Row` und `*sql.Rows` über ihre
// gemeinsame Scan-Signatur, damit scan-Helper beide Quellen bedienen.
type rowScanner interface {
	Scan(dest ...any) error
}

// rebind wandelt die dialekt-neutralen `?`-Platzhalter der aus dem
// SQLite-Adapter gespiegelten Query-Konstanten in Postgres' `$1,$2,…`.
// Die gespiegelten Statements tragen `?` ausschließlich als
// Bind-Platzhalter (nie in String-Literalen), daher genügt eine
// laufende Nummerierung — analog sqlx.Rebind für BindType Dollar.
func rebind(query string) string {
	var b strings.Builder
	b.Grow(len(query) + 8)
	n := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			n++
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
		} else {
			b.WriteByte(query[i])
		}
	}
	return b.String()
}

// formatTime serialisiert einen Zeitpunkt für die TEXT-Spalten. Das
// reversierte PG-Schema behielt SQLites datetime-als-TEXT (RFC3339Nano,
// UTC) — dieselbe ISO-8601-Konvention wie der SQLite-Adapter, damit die
// byte-lexikografische Sortierung (Keyset-Cursor) dialektgleich bleibt.
func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// parseTime ist das Pendant zu formatTime. Liefert immer UTC zurück.
func parseTime(raw string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("postgres: parse time %q: %w", raw, err)
	}
	return t.UTC(), nil
}

// nullableString mappt die Empty-String-Domain-Konvention auf SQL-NULL.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// stringFromNull ist das Lese-Pendant zu nullableString.
func stringFromNull(n sql.NullString) string {
	if !n.Valid {
		return ""
	}
	return n.String
}

// nullableTime wandelt einen optionalen Zeitpunkt in einen für
// database/sql passenden Wert: Nil-Pointer/Zero → SQL-NULL.
func nullableTime(p *time.Time) any {
	if p == nil || p.IsZero() {
		return nil
	}
	return formatTime(*p)
}

// timePtrOrNil liefert einen Pointer auf t, falls gesetzt; sonst nil.
func timePtrOrNil(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// nullableInt64 wandelt einen optionalen *int64 in einen für
// database/sql passenden Wert: nil → SQL-NULL, sonst der Wert.
func nullableInt64(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}

// nullableFloat64 wandelt einen optionalen *float64 in einen für
// database/sql passenden Wert: nil → SQL-NULL, sonst der Wert.
func nullableFloat64(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

// boolToInt mappt einen Go-bool auf den INTEGER-Wert (0/1). Das
// reversierte PG-Schema behielt SQLites boolean-als-INTEGER (z. B.
// time_skew_warning INTEGER), daher dieselbe Kodierung.
func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// encodeMeta serialisiert die Domain-Meta-Map als JSON-String. nil →
// SQL-NULL. Die meta-Spalte ist TEXT (reversiertes Schema), kein jsonb.
func encodeMeta(m domain.EventMeta) (any, error) {
	if m == nil {
		return nil, nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("postgres: encode meta: %w", err)
	}
	return string(raw), nil
}

// decodeMeta deserialisiert einen TEXT-meta-Wert zurück in die
// Domain-Map. SQL-NULL und Leerstring → nil-Map.
func decodeMeta(raw sql.NullString) (domain.EventMeta, error) {
	if !raw.Valid || raw.String == "" {
		return nil, nil
	}
	var m domain.EventMeta
	if err := json.Unmarshal([]byte(raw.String), &m); err != nil {
		return nil, fmt.Errorf("postgres: decode meta: %w", err)
	}
	return m, nil
}
