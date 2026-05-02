package driven

// IngestSequencer liefert den serverseitigen ingest_sequence-Wert für
// jedes vom Use Case akzeptierte Event. Werte sind monoton steigend
// und global eindeutig (ADR-0002 §8.1).
//
// Implementierungen:
//   - InMemory zählt einen atomaren int64 — Restart = Reset.
//   - SQLite initialisiert den Counter aus
//     `SELECT MAX(ingest_sequence) FROM playback_events` und setzt
//     fortlaufend in derselben durable Sequenz auf.
//
// Cursor-v2 (ADR-0004 §5) trägt `ingest_sequence` als finalen
// Tie-Breaker; eine Restart-stabile Pagination verlässt sich auf die
// SQLite-Variante.
//
// Implementierungen müssen für nebenläufige Aufrufe sicher sein.
type IngestSequencer interface {
	Next() int64
}
