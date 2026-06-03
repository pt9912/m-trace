package driving

import (
	"context"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// IngestControlInbound ist der Driving-Port für den Ingest-Control-
// Pfad (NF-13 / RAK-65..RAK-67). Der HTTP-Adapter
// dekodiert die Anfrage, leitet sie als `Create*Input`-Struktur an
// den Application-Layer weiter und mappt das Ergebnis auf das
// JSON-Schema aus `spec/backend-api-contract.md` §3.8. Domain- und
// Application-Layer bleiben frei von HTTP- und SQLite-Annahmen.
//
// Sicherheitsprofil:
//  - `CreateStreamResult` und `RotateKeyResult` enthalten den
//  **Klartext-Stream-Key** ausschließlich auf dem direkten
//  Use-Case-Output; die API-Antwort liefert ihn genau einmal an
//  den Aufrufer und persistiert ihn nirgends weiter.
//  - List-, Detail-, Validate- und Lifecycle-Pfade liefern
//  keinerlei Klartext-Werte.
type IngestControlInbound interface {
	CreateStream(ctx context.Context, req CreateStreamRequest) (CreateStreamResult, error)
	ListStreams(ctx context.Context, projectID string) ([]domain.IngestStream, error)
	GetStreamDetail(ctx context.Context, projectID, streamID string) (StreamDetail, error)
	RotateKey(ctx context.Context, projectID, streamID string) (RotateKeyResult, error)
	ValidateKey(ctx context.Context, projectID, streamID, candidateKey string) (ValidateKeyResult, error)
	RecordLifecycleEvent(ctx context.Context, req LifecycleHookRequest) (LifecycleEventResult, error)
	GetMediaServerConfig(ctx context.Context, projectID, targetID string) (MediaServerConfigResult, error)
}

// LifecycleHookRequest bündelt die Driving-Port-Eingabe für
// `POST /api/ingest/hooks/stream-{started,ended}` (`0.11.0` Tranche
// 4 / RAK-69). `Kind` wird vom HTTP-Adapter aus dem URL-Suffix
// abgeleitet (nicht aus dem Body), damit ein POST auf den Stop-
// Endpoint nicht mit einem manipulierten `type:"stream_started"`-
// Feld einen falschen Lifecycle einspeisen kann.
type LifecycleHookRequest struct {
	ResolvedProjectID string
	StreamID          string
	Kind              domain.StreamLifecycleEventKind
	ObservedAt        time.Time
	Source            domain.StreamLifecycleEventSource
	ConnectionID      string
	Reason            string
}

// LifecycleEventResult ist die Use-Case-Antwort für
// Lifecycle-Hook-Calls. Der HTTP-Adapter echo't den `EventID`-Wert
// im Acknowledgement; Klartext-Keys oder Hash-Werte erscheinen
// nicht.
type LifecycleEventResult struct {
	EventID    string
	StreamID   string
	Kind       domain.StreamLifecycleEventKind
	ObservedAt time.Time
}

// MediaServerConfigResult ist die Antwort auf
// `GET /api/ingest/media-server-config` (
// RAK-68). `ConfigYAML` enthält das deterministisch generierte
// MediaMTX-Artefakt; Klartext-Stream-Keys erscheinen niemals im
// Output (siehe Plan §0.7 + RAK-66). `Warnings` listet
// non-fatal Hinweise (z. B. übersprungene Streams mit
// nicht-konformem `display_name`).
type MediaServerConfigResult struct {
	TargetID   string
	Kind       domain.MediaServerKind
	ConfigPath string
	ConfigYAML string
	Warnings   []string
}

// CreateStreamRequest ist die Driving-Port-Eingabe für
// `POST /api/ingest/streams`. `RequestProjectID` ist der optionale
// Wire-Vertrag-Wert (Konsistenzcheck zum Token); der Use-Case nutzt
// `ResolvedProjectID` als kanonischen Wert.
//
// `Provision` schaltet ab (R-15) den optionalen
// Media-Server-I/O-Pfad. Default `false`: byte-stabil zum
// Format — kein Server-I/O, kein neues Response-Feld.
// `true` triggert den `MediaServerProvisioner.Apply`-Aufruf nach
// erfolgreicher API-State-Anlage; das Ergebnis fließt in
// `CreateStreamResult.MediaServerState`.
type CreateStreamRequest struct {
	ResolvedProjectID string
	RequestProjectID  string
	DisplayName       string
	Protocol          string
	EndpointID        string
	TargetID          string
	Provision         bool
}

// CreateStreamResult bündelt den frisch angelegten Stream und das
// **transiente** Klartext-Key-Material. Der HTTP-Adapter reicht
// `Material.Value` genau einmal an den Aufrufer durch; alles andere
// ist persistierbar.
//
// `MediaServerState` (`0.12.6` T9 / R-15) ist nur gesetzt, wenn der
// Request `Provision=true` trug. Bei `Provision=false` bleibt der
// Wert leer; der HTTP-Adapter mappt das auf ein **fehlendes**
// `media_server_state`-Feld im Wire-Body (byte-stabil zum
// Format).
type CreateStreamResult struct {
	Stream           domain.IngestStream
	Material         domain.StreamKeyMaterial
	MediaServerState string // "" | "disabled" | "applied" | "partial" | "failed"
	MediaServerHint  string // operator-sichtbarer Detail-String, leer wenn nicht nötig
}

// StreamDetail ergänzt den Stream um die referenzierten Endpoint-,
// Target- und Routing-Rule-Daten für die Detail-Antwort aus
// §3.8.
type StreamDetail struct {
	Stream      domain.IngestStream
	Endpoint    domain.IngestEndpoint
	Target      domain.MediaServerTarget
	RoutingRule domain.RoutingRule
}

// RotateKeyResult liefert den aktualisierten Stream und das neue
// Klartext-Key-Material. Klartext lebt nur in der Antwort.
type RotateKeyResult struct {
	Stream   domain.IngestStream
	Material domain.StreamKeyMaterial
}

// ValidateKeyResult ist die Antwort auf
// `POST /api/ingest/streams/{id}/validate-key`. `Valid:false` enthält
// keinen Stream-ID-Hinweis (Cross-Project-Leak-Schutz, §3.8).
type ValidateKeyResult struct {
	Valid          bool
	StreamID       string
	KeyFingerprint string
}
