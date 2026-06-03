package driven

import (
	"context"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// IngestStreamRepository persistiert die Ingest-Control-Domain
// (NF-13 / RAK-65..RAK-67). Implementierungen:
//  - `apps/api/adapters/driven/persistence/inmemory` für Tests und
//  Spike/Lab-Modus,
//  - `apps/api/adapters/driven/persistence/sqlite` für den
//  produktiv-bestätigten Persistenz-Pfad (Plan §0.4 SQLite-
//  Entscheidung).
//
// Sicherheitsprofil:
//  - Klartext-Stream-Keys werden **nicht** persistiert; der
//  Use-Case übergibt nur `domain.StreamKey` (Hash + Fingerprint).
//  - Cross-Project-Leak-Schutz ist Aufgabe des Repositorys: alle
//  Lookups MÜSSEN `projectID` filtern; `GetStreamByID` ohne
//  Treffer im Project liefert `domain.ErrIngestStreamNotFound`.
type IngestStreamRepository interface {
	// CreateStream legt einen Stream + Routing-Regel + initialen
	// Stream-Key in einer atomaren Transaktion an. Der Adapter
	// erzwingt die Pflicht-Constraints aus V2__ingest.sql
	// (`chk_ingest_streams_protocol`, `idx_ingest_streams_active_display_name`,
	// `idx_stream_keys_active_unique`) und gibt die folgenden
	// Domain-Fehler zurück:
	//  - `domain.ErrIngestEndpointNotFound` / `ErrIngestTargetNotFound`
	//  wenn der Endpoint/Target nicht existiert,
	//  - `domain.ErrIngestStreamNameConflict` bei doppeltem aktiven
	//  `display_name` im Project.
	CreateStream(ctx context.Context, input CreateStreamInput) (*domain.IngestStream, error)
	// GetStreamByID lädt einen Stream **inkl.** des aktiven
	// Stream-Keys (`stream.Key`). `projectID` ist Pflicht; ein
	// Stream eines fremden Projects darf wie nicht-existent
	// behandelt werden (`ErrIngestStreamNotFound`).
	GetStreamByID(ctx context.Context, projectID, streamID string) (*domain.IngestStream, error)
	// ListByProject liefert alle Streams eines Projects ohne
	// Klartext-Key — ausschließlich `stream.Key.Fingerprint` und
	// `stream.Key.CreatedAt`. Sortierung: `created_at desc,
	// stream_id asc` (deterministisch reproduzierbar; analog zur
	// bestehenden Stream-Session-Sortierung).
	ListByProject(ctx context.Context, projectID string) ([]domain.IngestStream, error)
	// RotateKey deaktiviert den aktuell aktiven Stream-Key und
	// fügt einen neuen `domain.StreamKey` ein. Der Aufrufer hat
	// den neuen Key bereits via `domain.GenerateStreamKey`
	// erzeugt; das Repository persistiert nur Hash + Fingerprint
	// und gibt den aktualisierten Stream zurück.
	RotateKey(ctx context.Context, projectID, streamID string, newKey domain.StreamKey) (*domain.IngestStream, error)
	// FindActiveStreamKey liefert den aktiven (`deactivated_at IS
	// NULL`) Stream-Key eines Streams für den Validate-Endpoint.
	// Der Use-Case nutzt das Ergebnis für `ValidateStreamKey`.
	FindActiveStreamKey(ctx context.Context, projectID, streamID string) (domain.StreamKey, error)
	// GetEndpointByID liefert einen `IngestEndpoint`. Liefert
	// `domain.ErrIngestEndpointNotFound`, wenn der Endpoint nicht
	// existiert.
	GetEndpointByID(ctx context.Context, endpointID string) (*domain.IngestEndpoint, error)
	// GetTargetByID liefert ein `MediaServerTarget`. Liefert
	// `domain.ErrIngestTargetNotFound` bei Nichtexistenz.
	GetTargetByID(ctx context.Context, targetID string) (*domain.MediaServerTarget, error)
	// GetRoutingRuleByID liefert die Routing-Regel zu einem
	// Stream. `domain.ErrIngestStreamNotFound` bei Nichtexistenz
	// (analog zur Stream-Sicht — kein Cross-Project-Leak).
	GetRoutingRuleByID(ctx context.Context, projectID, streamID string) (*domain.RoutingRule, error)
	// AppendLifecycleEvent persistiert ein `StreamLifecycleEvent`.
	// Die Tabelle ist append-only; Lebenszyklus-Aggregation lebt
	// in T4-Use-Cases und im Lab-Smoke. Klartext-Keys dürfen hier
	// niemals landen — der Aufrufer übergibt höchstens den
	// Fingerprint.
	AppendLifecycleEvent(ctx context.Context, event domain.StreamLifecycleEvent) error
}

// CreateStreamInput bündelt die Eingaben aus dem Use-Case-Pfad. Das
// Repository ist verantwortlich für ID-Vergabe (Stream-ID und
// Routing-Rule-ID), für Constraint-Verstoß-Mapping auf
// Domain-Fehler und für die atomare Transaktion über Stream +
// Routing-Regel + Initial-Key.
type CreateStreamInput struct {
	ProjectID    string
	DisplayName  string
	Protocol     domain.IngestProtocol
	EndpointID   string
	TargetID     string
	InitialKey   domain.StreamKey
	CreatedAt    time.Time
}
