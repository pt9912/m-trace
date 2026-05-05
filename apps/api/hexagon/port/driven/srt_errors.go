package driven

import "errors"

// Sentinel-Fehler für SrtSource-Adapter (plan-0.6.0 §4 Sub-3.6).
// Adapter (z. B. `adapters/driven/srt/mediamtxclient`) wrappen
// HTTP-/Parse-Fehler mit diesen Sentinels, damit der Use Case
// (`SrtHealthCollector`) sie via `errors.Is` auf
// `SourceStatus`/`SourceErrorCode` aus
// `spec/telemetry-model.md` §7.5 mappen kann, ohne das konkrete
// Adapter-Paket zu importieren (Hexagon-Boundary).
var (
	ErrSrtSourceUnauthorized = errors.New("srt source: unauthorized")
	ErrSrtSourceUnavailable  = errors.New("srt source: unavailable")
	ErrSrtSourceParseError   = errors.New("srt source: parse error")
)
