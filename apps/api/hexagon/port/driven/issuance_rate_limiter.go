package driven

import "context"

// IssuanceRateLimiter begrenzt die Issuance-Aufrufe von
// `POST /api/auth/session-tokens` (`0.12.0`, RAK-72). Der Limiter
// kennt zwei Buckets:
//
//   - global: ein Bucket für die gesamte Instanz, schützt vor
//     unbegrenztem Token-Mint über alle Projects hinweg.
//   - per-project: ein Bucket pro `project_id`, schützt vor einem
//     einzelnen kompromittierten Project Token, das die Issuance-
//     Quote für andere Projects auffressen würde.
//
// Beide Buckets werden bei jedem Aufruf in derselben Reihenfolge
// (global → project) geprüft; ein Verstoß auf einer Stufe verhindert
// die nachgelagerte Ausführung. Der `ctx`-Wert erlaubt Cancellation,
// nicht Storage-Lookups — der Adapter ist im `0.12.0`-Pflichtpfad
// in-memory.
type IssuanceRateLimiter interface {
	Allow(ctx context.Context, projectID string) (bool, error)
}
