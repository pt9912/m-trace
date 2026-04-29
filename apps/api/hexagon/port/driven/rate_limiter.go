package driven

import "context"

// RateLimiter consumes n event-tokens für jeden gesetzten Schlüssel
// in RateLimitKey. Returns domain.ErrRateLimited wenn auch nur eine
// Dimension das Budget überschritten hat; nil otherwise.
//
// Per plan-0.1.0.md §5.1 (F-110-Item) prüft die 0.1.0-Implementierung
// mindestens (project_id, client_ip); origin ist optional. Die
// Implementierung ist für nebenläufige Aufrufe sicher, prüft alle
// Dimensionen unter einer Sperre („all-or-nothing"-Commit) und
// verbraucht keine Tokens, wenn auch nur eine Dimension blockiert.
type RateLimiter interface {
	Allow(ctx context.Context, key RateLimitKey, n int) error
}

// RateLimitKey trägt die drei Dimensionen, gegen die der Limiter
// ratenbeschränkt. Leere Felder werden vom Adapter übersprungen — ein
// CLI/curl-Pfad ohne Origin reicht es z. B., dass die Project- und
// Client-IP-Dimension belegt sind.
type RateLimitKey struct {
	ProjectID string
	ClientIP  string
	Origin    string
}
