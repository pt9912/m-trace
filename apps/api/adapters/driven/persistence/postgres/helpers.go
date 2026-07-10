package postgres

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
