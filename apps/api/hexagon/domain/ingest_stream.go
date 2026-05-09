package domain

import (
	"errors"
	"strings"
	"time"
)

// Ingest-Control-Domain (`0.11.0`, NF-13 / RAK-65..RAK-70).
//
// Modul lebt in `apps/api/hexagon/domain/` analog zur restlichen
// Domain-Schicht (Variante B aus `docs/planning/in-progress/
// plan-0.11.0.md` §0.3). Persistenz, HTTP, Konfigurations-Generierung
// und Lifecycle-Events liegen in den Adapter-Schichten und
// konsumieren nur diese Typen plus `stream_key.go`.
//
// Sicherheitsgrenzen aus Plan §0.7:
//   - Klartext-Keys leben nur in Memory zwischen `GenerateStreamKey`
//     und Adapter-Output; sie werden in `IngestStream` nicht
//     gespeichert. Der Hash und der redigierte Fingerprint sitzen in
//     `StreamKey` und sind die einzigen persistenz- und
//     log-tauglichen Werte.
//   - Lifecycle-Events tragen höchstens den Fingerprint, niemals den
//     Klartext.

// IngestProtocol ist die in `0.11.0` zulässige Allowlist für
// Ingest-Endpunkt-Protokolle (RAK-67). Weitere Protokolle (z. B.
// WebRTC-Ingest, SRT-Listener mit Auth-Token) bleiben Folge-Scope
// und werden additiv ergänzt — bestehende Werte bleiben stabil.
type IngestProtocol string

// IngestProtocol-Werte aus der `0.11.0`-Allowlist (RAK-67).
const (
	IngestProtocolSRT  IngestProtocol = "srt"
	IngestProtocolRTMP IngestProtocol = "rtmp"
)

// IsKnown prüft, ob ein Protocol-Wert in der `0.11.0`-Allowlist
// vertreten ist. Der HTTP-Adapter mappt unbekannte Werte auf
// `400 invalid_request`; der Generator-Pfad lehnt sie strukturell
// ab. Die Funktion ist absichtlich case-sensitive, damit
// Wire-Vertrag und CLI-Smokes deterministisch sind.
func (p IngestProtocol) IsKnown() bool {
	switch p {
	case IngestProtocolSRT, IngestProtocolRTMP:
		return true
	default:
		return false
	}
}

// IngestStreamStatus klassifiziert den aktuellen Lifecycle-Zustand
// eines lokalen `IngestStream`. Übergänge sind monoton im Lab-Smoke-
// Pfad: `ready` → `live` → `ended`. `disabled` ist ein Operator-
// Override (z. B. wenn die Routing-Regel manuell deaktiviert wurde).
type IngestStreamStatus string

// IngestStreamStatus-Werte für den Lifecycle-Zustand eines lokalen
// Streams. Übergänge laut Plan §0.6 (`ready → live → ended`); `disabled`
// ist ein Operator-Override.
const (
	IngestStreamStatusReady    IngestStreamStatus = "ready"
	IngestStreamStatusLive     IngestStreamStatus = "live"
	IngestStreamStatusEnded    IngestStreamStatus = "ended"
	IngestStreamStatusDisabled IngestStreamStatus = "disabled"
)

// MediaServerKind benennt das Konfigurations-Profil eines
// `MediaServerTarget`. `0.11.0` deklariert MediaMTX als normativen
// Zielserver (RAK-68); SRS bleibt Kompatibilitäts-/Dokuhintergrund
// und ist daher als `MediaServerKindSRS` vertreten, aber kein
// Pflicht-Target.
type MediaServerKind string

// MediaServerKind-Werte für `MediaServerTarget.Kind` (RAK-68).
const (
	MediaServerKindMediaMTX MediaServerKind = "mediamtx"
	MediaServerKindSRS      MediaServerKind = "srs"
)

// IngestStream ist die Aggregat-Identität eines lokalen
// Stream-Control-Eintrags. Klartext-Keys leben **nicht** in dieser
// Struktur — der Adapter bekommt sie nur transient zurück.
type IngestStream struct {
	ID             string
	ProjectID      string
	DisplayName    string
	Protocol       IngestProtocol
	EndpointID     string
	TargetID       string
	RoutingRuleID  string
	Status         IngestStreamStatus
	Key            StreamKey
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// IngestEndpoint beschreibt einen lokal/lab-nahen Ingest-Pfad.
// `LabStack` benennt den `examples/`-Compose-Stack (z. B.
// `mtrace-srt`), `PublicURLHint` ist eine reine Doku-Zeile für
// `make smoke-ingest-control` und das Dashboard.
type IngestEndpoint struct {
	ID            string
	Protocol      IngestProtocol
	ListenHost    string
	ListenPort    int
	PathTemplate  string
	LabStack      string
	PublicURLHint string
}

// MediaServerTarget bündelt das Lab-/Demo-Ziel eines Streams. Das
// `ConfigPath`-Feld zeigt auf das vom Generator erzeugte Artefakt
// (z. B. `examples/ingest-control/mediamtx.generated.yml`).
// `ControlAPIURL` ist optional und bleibt für `0.11.0` rein
// dokumentarisch — der Plan schließt produktive Provisionierung über
// MediaMTX-Control-API explizit aus.
type MediaServerTarget struct {
	ID              string
	Kind            MediaServerKind
	ConfigPath      string
	HLSURLTemplate  string
	ControlAPIURL   string
}

// RoutingRuleMode unterscheidet die zulässigen Routing-Modi. Plan
// §0.1 schließt Fan-out, Failover und Load-Balancing für `0.11.0`
// aus; `single` ist deshalb der einzige Wert, ist aber als String-
// Enum modelliert, damit Folge-Releases additiv erweitern können.
type RoutingRuleMode string

// RoutingRuleMode-Werte aus dem `0.11.0`-Scope. Fan-out/Failover/
// Load-Balancing sind Folge-Scope.
const (
	RoutingRuleModeSingle RoutingRuleMode = "single"
)

// RoutingRule verbindet einen `IngestStream` mit genau einem
// `MediaServerTarget`. Eine deaktivierte Regel (`Enabled=false`)
// blockiert Lifecycle-Events deterministisch (`409
// routing_rule_disabled` im HTTP-Adapter, RAK-70).
type RoutingRule struct {
	ID       string
	StreamID string
	TargetID string
	Mode     RoutingRuleMode
	Enabled  bool
}

// StreamLifecycleEventKind benennt die in `0.11.0` ausgelieferten
// Lifecycle-Events (RAK-69). Produktive ausgehende Webhook-
// Zustellung an externe Systeme bleibt Folge-Scope (siehe R-16 im
// `risks-backlog.md`).
type StreamLifecycleEventKind string

// StreamLifecycleEventKind-Werte aus RAK-69 (`stream_started` /
// `stream_ended`).
const (
	StreamLifecycleEventStarted StreamLifecycleEventKind = "stream_started"
	StreamLifecycleEventEnded   StreamLifecycleEventKind = "stream_ended"
)

// StreamLifecycleEventSource benennt den Auslöser eines Events.
// `smoke` markiert ein lokal über
// `POST /api/ingest/hooks/stream-{started,ended}` eingespeistes
// Event; `mediamtx-hook` reserviert den späteren Adapter-Pfad, ist
// in `0.11.0` aber nicht aktiv.
type StreamLifecycleEventSource string

// StreamLifecycleEventSource-Werte: `smoke` für lokal eingespeiste
// Events, `mediamtx-hook` reserviert den späteren Adapter-Pfad.
const (
	StreamLifecycleSourceSmoke        StreamLifecycleEventSource = "smoke"
	StreamLifecycleSourceMediaMTXHook StreamLifecycleEventSource = "mediamtx-hook"
)

// StreamLifecycleEvent ist das normative Eventmodell. `KeyFingerprint`
// ist der **einzige** Key-bezogene Wert, der hier auftaucht — der
// Klartext darf weder in Logs noch in Persistenz noch in Webhook-
// Payloads erscheinen (RAK-66/RAK-69).
type StreamLifecycleEvent struct {
	Kind           StreamLifecycleEventKind
	StreamID       string
	ProjectID      string
	OccurredAt     time.Time
	Source         StreamLifecycleEventSource
	KeyFingerprint string
}

// Validation-/Domain-Errors für Ingest-Control. HTTP-Mapping in
// `spec/backend-api-contract.md` §3.8 dokumentiert.
var (
	ErrIngestProtocolUnknown      = errors.New("ingest protocol must be one of: srt, rtmp")
	ErrIngestStreamNotFound       = errors.New("ingest stream not found in project")
	ErrIngestStreamNameConflict   = errors.New("active ingest stream with same display_name already exists in project")
	ErrIngestEndpointNotFound     = errors.New("ingest endpoint not found")
	ErrIngestTargetNotFound       = errors.New("media-server target not found")
	ErrIngestRoutingRuleDisabled  = errors.New("routing rule is disabled")
	ErrIngestProjectIDMismatch    = errors.New("request project_id does not match resolved token")
	ErrIngestKeyInvalid           = errors.New("stream key validation failed")
	ErrIngestMediaServerConfigUnavailable = errors.New("media-server configuration could not be generated")
)

// ValidateIngestProtocol normalisiert (Whitespace, Lowercase) und
// prüft die Allowlist. Der Adapter sollte den raw-Wert für die
// Fehlermeldung behalten, damit der Aufrufer den verstoßenden Wert
// im Body wiedererkennt.
func ValidateIngestProtocol(raw string) (IngestProtocol, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	candidate := IngestProtocol(normalized)
	if !candidate.IsKnown() {
		return "", ErrIngestProtocolUnknown
	}
	return candidate, nil
}

// ValidateProjectIDConsistency prüft die in `spec/backend-api-
// contract.md` §3.8 normierte Project-ID-Konsistenz: ein optionaler
// Request-`project_id`-Wert darf nur als Konsistenzcheck dienen und
// muss zum Token passen. Leere Request-Werte sind erlaubt; der
// Aufrufer setzt dann den Token-Project-ID-Wert ein.
func ValidateProjectIDConsistency(requestProjectID, resolvedProjectID string) error {
	requestTrim := strings.TrimSpace(requestProjectID)
	if requestTrim == "" {
		return nil
	}
	if requestTrim != resolvedProjectID {
		return ErrIngestProjectIDMismatch
	}
	return nil
}

// FilterStreamForProject ist der Wire-Vertrag-konforme „Existenz-
// Schutz": Streams aus fremden Projekten werden wie nicht-existent
// behandelt (§3.8 — kein Cross-Project-Leak). Der Aufrufer kann den
// Returnwert direkt als `domain.ErrIngestStreamNotFound`-Marker
// verwenden.
func FilterStreamForProject(stream *IngestStream, resolvedProjectID string) (*IngestStream, error) {
	if stream == nil || stream.ProjectID != resolvedProjectID {
		return nil, ErrIngestStreamNotFound
	}
	return stream, nil
}
