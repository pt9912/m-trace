package application

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// Reservierte Meta-Keys aus spec/telemetry-model.md §1.4 und
// contracts/event-schema.json. Werden inbound vor Persistenz
// typvalidiert; Verstöße liefern 422 (domain.ErrInvalidEvent) und
// werden nicht persistiert.
const (
	metaKeyNetworkKind              = "network.kind"
	metaKeyNetworkDetailStatus      = "network.detail_status"
	metaKeyNetworkUnavailableReason = "network.unavailable_reason"
	metaKeyNetworkRedactedURL       = "network.redacted_url"
	metaKeyTimingPrefix             = "timing."

	networkDetailStatusUnavailable = "network_detail_unavailable"
)

// networkUnavailableReasonPattern entspricht
// contracts/event-schema.json#network_unavailable_reason_pattern.
// `regexp.MustCompile`-Globals werden vom Linter-Default
// gochecknoglobals ignoriert (kompilierte Regex ohne Mutationsfläche).
var networkUnavailableReasonPattern = regexp.MustCompile(`^[a-z0-9_]{1,64}$`)

// isReservedNetworkKindValue spiegelt
// contracts/event-schema.json#reserved_meta_keys["network.kind"].values.
func isReservedNetworkKindValue(s string) bool {
	switch s {
	case "manifest", "segment":
		return true
	default:
		return false
	}
}

// isReservedNetworkDetailStatus spiegelt
// contracts/event-schema.json#reserved_meta_keys["network.detail_status"].values.
func isReservedNetworkDetailStatus(s string) bool {
	switch s {
	case "available", networkDetailStatusUnavailable:
		return true
	default:
		return false
	}
}

// isNetworkUnavailableReason spiegelt
// contracts/event-schema.json#network_unavailable_reasons (gemeinsamer
// Reason-Enum mit session_boundaries[].reason; einziger normativer
// Anker in spec/telemetry-model.md §1.4).
func isNetworkUnavailableReason(s string) bool {
	switch s {
	case "native_hls_unavailable",
		"hlsjs_signal_unavailable",
		"browser_api_unavailable",
		"resource_timing_unavailable",
		"cors_timing_blocked",
		"service_worker_opaque":
		return true
	default:
		return false
	}
}

// validateReservedEventMeta prüft typ- und domänenspezifisch die
// reservierten `network.*`- und `timing.*`-Keys aus
// spec/telemetry-model.md §1.4. Verstöße werden vor Persistenz mit
// 422 abgewiesen. Nicht-reservierte Keys bleiben unangetastet
// (Vorwärtskompatibilität nach API-Kontrakt §3.4 — alte Backends
// ignorieren unbekannte additive Keys).
func validateReservedEventMeta(meta domain.EventMeta) error {
	if len(meta) == 0 {
		return nil
	}
	for k, v := range meta {
		if err := validateReservedKeyValue(k, v); err != nil {
			return err
		}
	}
	return validateReasonRequiresUnavailable(meta)
}

// validateReservedKeyValue prüft den Typ und die Domäne eines
// einzelnen reservierten Keys.
func validateReservedKeyValue(k string, v any) error {
	switch k {
	case metaKeyNetworkKind:
		return requireEnumString(k, v, isReservedNetworkKindValue)
	case metaKeyNetworkDetailStatus:
		return requireEnumString(k, v, isReservedNetworkDetailStatus)
	case metaKeyNetworkUnavailableReason:
		return validateUnavailableReason(k, v)
	case metaKeyNetworkRedactedURL:
		return validateRedactedURLValue(k, v)
	}
	if strings.HasPrefix(k, metaKeyTimingPrefix) {
		return validateTimingValue(k, v)
	}
	return nil
}

// validateReasonRequiresUnavailable prüft die `requires`-Beziehung aus
// contracts/event-schema.json#reserved_meta_keys["network.unavailable_reason"]:
// `unavailable_reason` ist nur zulässig, wenn `detail_status` exakt
// `network_detail_unavailable` ist. Bei `available` (oder fehlendem
// detail_status) ist ein gesetzter Reason-Wert ein semantischer
// Widerspruch und wird mit 422 abgelehnt.
func validateReasonRequiresUnavailable(meta domain.EventMeta) error {
	rawReason, ok := meta[metaKeyNetworkUnavailableReason]
	if !ok {
		return nil
	}
	reason, _ := rawReason.(string)
	if reason == "" {
		return nil
	}
	detailStatus, _ := meta[metaKeyNetworkDetailStatus].(string)
	if detailStatus != networkDetailStatusUnavailable {
		return fmt.Errorf(
			"%w: meta[%q] requires meta[%q]=%q",
			domain.ErrInvalidEvent,
			metaKeyNetworkUnavailableReason,
			metaKeyNetworkDetailStatus,
			networkDetailStatusUnavailable,
		)
	}
	return nil
}

// requireEnumString prüft, dass der Wert ein nicht-leerer String aus
// einer Enum-Domäne ist. Kein-String, Objekte/Arrays oder Werte
// außerhalb der Domäne führen zu domain.ErrInvalidEvent. Die Domäne
// wird über ein Prädikat statt einer Map übergeben, damit kein
// gochecknoglobals-Verstoß entsteht.
func requireEnumString(key string, v any, allowed func(string) bool) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w: meta[%q] must be string", domain.ErrInvalidEvent, key)
	}
	if !allowed(s) {
		return fmt.Errorf("%w: meta[%q] value not in reserved domain", domain.ErrInvalidEvent, key)
	}
	return nil
}

// validateUnavailableReason prüft den Reason gegen Enum + Pattern aus
// spec/telemetry-model.md §1.4.
func validateUnavailableReason(key string, v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w: meta[%q] must be string", domain.ErrInvalidEvent, key)
	}
	if !networkUnavailableReasonPattern.MatchString(s) {
		return fmt.Errorf("%w: meta[%q] violates ^[a-z0-9_]{1,64}$", domain.ErrInvalidEvent, key)
	}
	if !isNetworkUnavailableReason(s) {
		return fmt.Errorf("%w: meta[%q] not in reserved reason enum", domain.ErrInvalidEvent, key)
	}
	return nil
}

// validateRedactedURLValue prüft, dass der bereits redigierte URL-
// Repräsentant keine rohen Bestandteile (Query, Fragment, userinfo,
// tokenartige Pfadsegmente) enthält. Vor Persistenz, sodass ein
// fehlerhaft redigierter Wert das Event mit 422 abweist und nicht
// roh in den Store läuft.
func validateRedactedURLValue(key string, v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w: meta[%q] must be string", domain.ErrInvalidEvent, key)
	}
	if s == "" {
		return nil
	}
	if !isAlreadyRedactedURL(s) {
		return fmt.Errorf(
			"%w: meta[%q] must be already redacted (no query/fragment/userinfo/token segment)",
			domain.ErrInvalidEvent, key,
		)
	}
	return nil
}

// validateTimingValue prüft `timing.*`-Keys gegen number oder explizit
// dokumentierten RFC3339-String aus contracts/event-schema.json.
func validateTimingValue(key string, v any) error {
	switch typed := v.(type) {
	case float64, float32, int, int64:
		return nil
	case string:
		if typed == "" {
			return fmt.Errorf("%w: meta[%q] timing string is empty", domain.ErrInvalidEvent, key)
		}
		if _, err := time.Parse(time.RFC3339Nano, typed); err != nil {
			if _, err2 := time.Parse(time.RFC3339, typed); err2 != nil {
				return fmt.Errorf("%w: meta[%q] timing string not RFC3339", domain.ErrInvalidEvent, key)
			}
		}
		return nil
	default:
		return fmt.Errorf("%w: meta[%q] timing value must be number or RFC3339 string", domain.ErrInvalidEvent, key)
	}
}
