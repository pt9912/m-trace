package domain

// EventAppendedFrame ist das Mindest-Wire-Format für SSE-Live-Updates
// (spec/backend-api-contract.md). Konsumenten laden den vollen Event-/
// Session-Read-Shape per REST nach. `TimeSkewWarning` (R-5) wird
// mitgeschickt, damit das Dashboard den Indikator schon im Live-Update
// setzen kann, ohne den vollen Detail-Read nachzuziehen.
//
// Domain-Typ (slice-004), damit der SSE-Driving-Adapter und der
// Ingest-Use-Case dieselbe Form ohne Application-Import teilen.
type EventAppendedFrame struct {
	ProjectID       string
	SessionID       string
	EventName       string
	IngestSequence  int64
	TimeSkewWarning bool
}
