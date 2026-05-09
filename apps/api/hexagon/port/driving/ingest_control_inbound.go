package driving

import (
	"context"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// IngestControlInbound ist der Driving-Port für den Ingest-Control-
// Pfad (`0.11.0` Tranche 2, NF-13 / RAK-65..RAK-67). Der HTTP-Adapter
// dekodiert die Anfrage, leitet sie als `Create*Input`-Struktur an
// den Application-Layer weiter und mappt das Ergebnis auf das
// JSON-Schema aus `spec/backend-api-contract.md` §3.8. Domain- und
// Application-Layer bleiben frei von HTTP- und SQLite-Annahmen.
//
// Sicherheitsprofil:
//   - `CreateStreamResult` und `RotateKeyResult` enthalten den
//     **Klartext-Stream-Key** ausschließlich auf dem direkten
//     Use-Case-Output; die API-Antwort liefert ihn genau einmal an
//     den Aufrufer und persistiert ihn nirgends weiter.
//   - List-, Detail-, Validate- und Lifecycle-Pfade liefern
//     keinerlei Klartext-Werte.
type IngestControlInbound interface {
	CreateStream(ctx context.Context, req CreateStreamRequest) (CreateStreamResult, error)
	ListStreams(ctx context.Context, projectID string) ([]domain.IngestStream, error)
	GetStreamDetail(ctx context.Context, projectID, streamID string) (StreamDetail, error)
	RotateKey(ctx context.Context, projectID, streamID string) (RotateKeyResult, error)
	ValidateKey(ctx context.Context, projectID, streamID, candidateKey string) (ValidateKeyResult, error)
	RecordLifecycleEvent(ctx context.Context, projectID, streamID string, kind domain.StreamLifecycleEventKind, occurredAt time.Time, source domain.StreamLifecycleEventSource) error
}

// CreateStreamRequest ist die Driving-Port-Eingabe für
// `POST /api/ingest/streams`. `RequestProjectID` ist der optionale
// Wire-Vertrag-Wert (Konsistenzcheck zum Token); der Use-Case nutzt
// `ResolvedProjectID` als kanonischen Wert.
type CreateStreamRequest struct {
	ResolvedProjectID string
	RequestProjectID  string
	DisplayName       string
	Protocol          string
	EndpointID        string
	TargetID          string
}

// CreateStreamResult bündelt den frisch angelegten Stream und das
// **transiente** Klartext-Key-Material. Der HTTP-Adapter reicht
// `Material.Value` genau einmal an den Aufrufer durch; alles andere
// ist persistierbar.
type CreateStreamResult struct {
	Stream   domain.IngestStream
	Material domain.StreamKeyMaterial
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
