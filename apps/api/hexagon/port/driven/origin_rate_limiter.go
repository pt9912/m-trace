package driven

import "context"

// OriginRateLimiter begrenzt Aufrufe pro Client-Origin (Default:
// `r.RemoteAddr`) oder, bei aktivem Reverse-Proxy-Trust, pro
// `X-Forwarded-For`-Client-IP. Spec-Anker: /
// R-22 (Backend-Strategie §8).
//
// Im Gegensatz zum `IssuanceRateLimiter` (Project-skopiert,
// Two-Bucket: global + per-project) ist der Origin-Limiter
// **single-bucket pro Key** — der Aufrufer entscheidet das
// Key-Schema (`ip:<addr>` bzw. `origin:<hash>`). Das passt zur
// Defense-in-Depth-Rolle: der Limiter sitzt **vor** dem
// Project-Limit und schützt vor Brute-Force-Bursts aus einer
// einzelnen Quelle, bevor der Project-Layer überhaupt geprüft wird.
//
// `Allow=false` mappt der HTTP-Adapter auf
// `429 origin_rate_limited` (siehe HTTP-Handler-Integration).
// `key=""` ist No-Op (`true, nil`) — fehlende RemoteAddr-/XFF-
// Information darf den Pfad nicht blockieren.
type OriginRateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}
