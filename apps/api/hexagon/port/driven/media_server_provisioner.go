package driven

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// MediaServerProvisioner ist der optionale Adapter, der einen
// laufenden Media-Server (MediaMTX/SRS) konfiguriert, wenn der
// `POST /api/ingest/streams?provision=true`-Pfad aktiviert ist
// (R-15, spec/backend-api-contract.md).
//
// Adapter-Vertrag:
//
//  - `Apply(ctx, config)`: server-seitiges I/O — Endpoint und
//  Routing-Rule auf dem externen Server anlegen. Idempotenz ist
//  Adapter-Detail (MediaMTX-`/v3/config/`-Endpunkte sind
//  idempotent über PUT; SRS bleibt Folge-Item nach `0.12.6`).
//  Liefert `MediaServerStateApplied` bei vollem Erfolg,
//  `MediaServerStatePartial` bei teilweise erfolgter
//  Konfiguration (z. B. Endpoint angelegt, Routing-Rule
//  rejected), `MediaServerStateFailed` bei vollständigem
//  Fehlschlag plus `ErrorCode` für den Operator. Der Aufrufer
//  (Use-Case) macht **keinen** API-State-Rollback bei `failed`/
//  `partial` — der HTTP-Adapter signalisiert das im Response
//  mit `201 Created` plus `media_server_state="failed"`.
//
//  - `Rollback(ctx, ids)`: Best-Effort-Cleanup, wenn der Operator
//  einen Stream löscht oder das API-State-vs-Server-State-
//  Diff manuell synchronisieren möchte. Wird vom Use-Case in
//  `0.12.6` NICHT automatisch aufgerufen — `provision=true` ist
//  opt-in und Rollback ist Operator-Verantwortung über einen
//  Folge-Endpoint (post-`0.12.6`).
//
// **Disabled-Pfad**: ein nil-Provisioner (Boot-Wiring ohne
// `MTRACE_MEDIASERVER_PROVISION_URL`) signalisiert dem Use-Case,
// dass der Adapter deaktiviert ist. Der HTTP-Antwort-Block trägt
// dann `media_server_state="disabled"` plus Hinweis-Text.
type MediaServerProvisioner interface {
	Apply(ctx context.Context, cfg MediaServerApplyInput) (MediaServerApplyResult, error)
	Rollback(ctx context.Context, projectID, streamID string) error
}

// MediaServerApplyInput trägt die für die externe Provisionierung
// relevanten Felder des frisch angelegten Streams. Der Adapter
// projiziert diese auf die Wire-Form des Target-Servers.
type MediaServerApplyInput struct {
	ProjectID     string
	Stream        domain.IngestStream
	StreamKeyHash string // SHA-256-Hash; Klartext wird hier nicht durchgereicht.
}

// MediaServerApplyResult bündelt das Server-Resultat plus optionale
// Operator-Hinweise. Der HTTP-Adapter mappt `State` auf das
// `media_server_state`-Wire-Feld.
type MediaServerApplyResult struct {
	State     MediaServerState
	ErrorCode string // optional; z. B. "unreachable", "auth_failure", "rule_rejected"
	Detail    string // operator-sichtbarer Detail-String, < 256 chars
}

// MediaServerState ist der bounded Wire-Enum für die Response.
type MediaServerState string

// Wire-Enum-Werte gemäß spec/backend-api-contract.md
// (`0.12.6` T9).
const (
	// MediaServerStateDisabled: ENV nicht konfiguriert; Adapter nil
	// im Use-Case.
	MediaServerStateDisabled MediaServerState = "disabled"
	// MediaServerStateApplied: Endpoint + Routing-Rule erfolgreich
	// auf dem externen Server angelegt.
	MediaServerStateApplied MediaServerState = "applied"
	// MediaServerStatePartial: Endpoint angelegt, Routing-Rule oder
	// Folge-Operation rejected; Operator-Sync nötig.
	MediaServerStatePartial MediaServerState = "partial"
	// MediaServerStateFailed: Server unreachable, Auth-Failure oder
	// Server-side-Reject; lokaler API-State + Stream sind angelegt,
	// externer Server ist nicht synchron.
	MediaServerStateFailed MediaServerState = "failed"
)
