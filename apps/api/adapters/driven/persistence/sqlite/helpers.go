package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// nullSeqSentinel ist der int64-Wert, mit dem `sequence_number IS NULL`
// in COALESCE-basierten Vergleichen behandelt wird. Wert ist
// math.MinInt64; damit sortieren NULL-Events strikt VOR allen
// gesetzten sequence_number-Werten — passend zur kanonischen
// Sortier-Order (ADR-0002 §8.1, API-Kontrakt §10.4).
const nullSeqSentinel int64 = -1 << 63

// persistedSchemaVersion ist die einzige Wire-Format-Version, die
// dieses Release annimmt (siehe spec/backend-api-contract.md §3.4).
const persistedSchemaVersion = "1.0"

// formatTime serialisiert einen Zeitpunkt für die TEXT-Spalten der
// SQLite-Migration; UTC + RFC3339Nano sind die ISO-8601-Konvention,
// die d-migrate für `datetime`-Spalten in SQLite festlegt.
func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// parseTime ist das Pendant zu formatTime. Liefert immer UTC zurück.
func parseTime(raw string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("sqlite: parse time %q: %w", raw, err)
	}
	return t.UTC(), nil
}

// nullableInt64 wandelt einen optionalen *int64 in einen für
// database/sql passenden Wert: nil → SQL-NULL, sonst der Wert.
func nullableInt64(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}

// nullableString mappt Empty-String-Domain-Konvention auf SQL-NULL.
// Einsetzbar für nullable TEXT-Spalten wie trace_id/span_id/
// correlation_id, deren Domain-Pendants als Empty-String den
// "nicht gesetzt"-Zustand kodieren.
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
// Hilft, time.Time-Felder mit Zero-Wert auf SQL-NULL zu mappen, ohne
// dass der Aufrufer explizit auf IsZero prüfen muss.
func timePtrOrNil(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// nullableFloat64 wandelt einen optionalen *float64 in einen für
// database/sql passenden Wert: nil → SQL-NULL, sonst der Wert.
func nullableFloat64(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

// encodeMeta serialisiert die Domain-Meta-Map als JSON-String. nil →
// SQL-NULL, sonst kompaktes JSON. ADR-0002 §8.1 fordert App-Layer-
// Validierung statt DB-CHECK für Postgres-Portabilität.
func encodeMeta(m domain.EventMeta) (any, error) {
	if m == nil {
		return nil, nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("sqlite: encode meta: %w", err)
	}
	return string(raw), nil
}

// decodeMeta deserialisiert einen TEXT-meta-Wert zurück in die Domain-
// Map. SQL-NULL und Leerstring → nil-Map.
func decodeMeta(raw sql.NullString) (domain.EventMeta, error) {
	if !raw.Valid || raw.String == "" {
		return nil, nil
	}
	var m domain.EventMeta
	if err := json.Unmarshal([]byte(raw.String), &m); err != nil {
		return nil, fmt.Errorf("sqlite: decode meta: %w", err)
	}
	return m, nil
}
