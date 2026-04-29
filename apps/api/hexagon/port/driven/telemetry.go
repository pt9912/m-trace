package driven

import "context"

// Telemetry kapselt OTel-Aufrufe hinter einer frameworkneutralen
// Schnittstelle. Damit darf hexagon/ keinen direkten OTel-Import
// tragen — siehe docs/architecture.md §3.4 und der Boundary-Test
// scripts/check-architecture.sh.
//
// Der Use Case ruft BatchReceived am Eintritt jedes
// RegisterPlaybackEventBatch-Aufrufs (vor Step 3 Auth-Token), damit
// der Counter auch auth-fehlgeschlagene Requests zählt — mtrace.api.
// batches.received misst received, nicht validated.
type Telemetry interface {
	BatchReceived(ctx context.Context, size int)
}
