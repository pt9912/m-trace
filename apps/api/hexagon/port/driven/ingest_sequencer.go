package driven

// IngestSequencer liefert den serverseitigen ingest_sequence-Wert für
// jedes vom Use Case akzeptierte Event. Die produktive Implementierung
// (apps/api/adapters/driven/persistence/InMemoryIngestSequencer) zählt
// einen atomaren int64 hoch — pro Prozess monoton, beginnend bei 1.
//
// Eine SQLite-Migration (OE-3-Folge-ADR) wird die Sequence-Quelle in
// die Datenbank verlagern; bis dahin ist Restart = Sequence-Reset, was
// in Kombination mit der process_instance_id im Cursor (siehe
// plan-0.1.0.md §5.1) zur Cursor-Invalidierung führt.
//
// Implementierungen müssen für nebenläufige Aufrufe sicher sein.
type IngestSequencer interface {
	Next() int64
}
