package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// IngestStreamRepository ist die durable SQLite-Variante des
// `driven.IngestStreamRepository`-Ports (`0.11.0` Tranche 2).
// Tabellen aus V2__ingest.sql.
//
// Sicherheitsprofil:
//   - Klartext-Stream-Keys werden nicht persistiert. `stream_keys`
//     speichert ausschließlich `key_hash` und `fingerprint`.
//   - `idx_stream_keys_active_unique` erzwingt einen einzigen aktiven
//     Key pro `(project_id, key_hash)`; rotierte Keys bleiben für
//     Audit liegen.
//   - Cross-Project-Lookups liefern `domain.ErrIngestStreamNotFound`
//     ohne Hinweis auf Existenz.
type IngestStreamRepository struct {
	db *sql.DB
}

// NewIngestStreamRepository konstruiert den Adapter.
func NewIngestStreamRepository(db *sql.DB) *IngestStreamRepository {
	return &IngestStreamRepository{db: db}
}

// CreateStream legt Project, Stream, Routing-Regel und initialen Key
// in einer Transaktion an.
func (r *IngestStreamRepository) CreateStream(ctx context.Context, input driven.CreateStreamInput) (*domain.IngestStream, error) {
	if _, err := r.GetEndpointByID(ctx, input.EndpointID); err != nil {
		return nil, err
	}
	if _, err := r.GetTargetByID(ctx, input.TargetID); err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ingest: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `INSERT INTO projects(project_id) VALUES (?) ON CONFLICT(project_id) DO NOTHING`, input.ProjectID); err != nil {
		return nil, fmt.Errorf("ingest: ensure project: %w", err)
	}

	streamID := newIngestID("ing_")
	ruleID := newIngestID("route_")
	keyID := newIngestID("key_")
	createdAt := input.CreatedAt.UTC().Format(time.RFC3339Nano)

	_, err = tx.ExecContext(ctx, `
INSERT INTO ingest_streams(stream_id, project_id, display_name, protocol, endpoint_id, target_id, routing_rule_id, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		streamID, input.ProjectID, input.DisplayName, string(input.Protocol),
		input.EndpointID, input.TargetID, ruleID, string(domain.IngestStreamStatusReady),
		createdAt, createdAt)
	if err != nil {
		return nil, mapIngestStreamCreateError(err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO ingest_routing_rules(rule_id, project_id, stream_id, target_id, mode, enabled)
VALUES (?, ?, ?, ?, ?, 1)`,
		ruleID, input.ProjectID, streamID, input.TargetID, string(domain.RoutingRuleModeSingle)); err != nil {
		return nil, fmt.Errorf("ingest: insert routing rule: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO stream_keys(key_id, project_id, stream_id, key_hash, fingerprint, created_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		keyID, input.ProjectID, streamID, input.InitialKey.Hash, input.InitialKey.Fingerprint, createdAt); err != nil {
		return nil, fmt.Errorf("ingest: insert initial key: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("ingest: commit: %w", err)
	}

	return &domain.IngestStream{
		ID:            streamID,
		ProjectID:     input.ProjectID,
		DisplayName:   input.DisplayName,
		Protocol:      input.Protocol,
		EndpointID:    input.EndpointID,
		TargetID:      input.TargetID,
		RoutingRuleID: ruleID,
		Status:        domain.IngestStreamStatusReady,
		Key:           input.InitialKey,
		CreatedAt:     input.CreatedAt,
		UpdatedAt:     input.CreatedAt,
	}, nil
}

func mapIngestStreamCreateError(err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "idx_ingest_streams_active_display_name"),
		strings.Contains(msg, "ingest_streams.display_name"),
		strings.Contains(msg, "ingest_streams.project_id, ingest_streams.display_name"):
		return domain.ErrIngestStreamNameConflict
	case strings.Contains(msg, "FOREIGN KEY") && strings.Contains(msg, "endpoint"):
		return domain.ErrIngestEndpointNotFound
	case strings.Contains(msg, "FOREIGN KEY") && strings.Contains(msg, "target"):
		return domain.ErrIngestTargetNotFound
	default:
		return fmt.Errorf("ingest: insert stream: %w", err)
	}
}

// GetStreamByID liefert einen Stream inkl. aktivem Key. Cross-Project-
// Lookups (anderes `projectID`) liefern `domain.ErrIngestStreamNotFound`.
func (r *IngestStreamRepository) GetStreamByID(ctx context.Context, projectID, streamID string) (*domain.IngestStream, error) {
	stream, err := r.scanStream(ctx, r.db, projectID, streamID)
	if err != nil {
		return nil, err
	}
	key, err := r.FindActiveStreamKey(ctx, projectID, streamID)
	if err != nil && !errors.Is(err, domain.ErrIngestKeyInvalid) {
		return nil, err
	}
	stream.Key = key
	return &stream, nil
}

// ListByProject liefert alle Streams eines Projects mit aktivem Key.
// Sortierung: `created_at desc, stream_id asc`.
func (r *IngestStreamRepository) ListByProject(ctx context.Context, projectID string) ([]domain.IngestStream, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT stream_id, project_id, display_name, protocol, endpoint_id, target_id, routing_rule_id, status, created_at, updated_at
FROM ingest_streams
WHERE project_id = ?
ORDER BY created_at DESC, stream_id ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("ingest: list streams: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []domain.IngestStream
	for rows.Next() {
		stream, err := scanIngestStream(rows)
		if err != nil {
			return nil, err
		}
		key, err := r.FindActiveStreamKey(ctx, projectID, stream.ID)
		if err != nil && !errors.Is(err, domain.ErrIngestKeyInvalid) {
			return nil, err
		}
		stream.Key = key
		out = append(out, stream)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ingest: list streams iter: %w", err)
	}
	return out, nil
}

// RotateKey deaktiviert den bisher aktiven Stream-Key (UPDATE
// `deactivated_at`) und fügt den neuen Key ein. Atomare Tx über die
// drei Schritte.
func (r *IngestStreamRepository) RotateKey(ctx context.Context, projectID, streamID string, newKey domain.StreamKey) (*domain.IngestStream, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ingest: rotate begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stream, err := r.scanStream(ctx, tx, projectID, streamID)
	if err != nil {
		return nil, err
	}

	deactivatedAt := newKey.CreatedAt.UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `
UPDATE stream_keys
SET deactivated_at = ?
WHERE project_id = ? AND stream_id = ? AND deactivated_at IS NULL`,
		deactivatedAt, projectID, streamID); err != nil {
		return nil, fmt.Errorf("ingest: deactivate old key: %w", err)
	}

	keyID := newIngestID("key_")
	if _, err := tx.ExecContext(ctx, `
INSERT INTO stream_keys(key_id, project_id, stream_id, key_hash, fingerprint, created_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		keyID, projectID, streamID, newKey.Hash, newKey.Fingerprint, deactivatedAt); err != nil {
		return nil, fmt.Errorf("ingest: insert rotated key: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE ingest_streams
SET updated_at = ?
WHERE project_id = ? AND stream_id = ?`,
		deactivatedAt, projectID, streamID); err != nil {
		return nil, fmt.Errorf("ingest: bump updated_at: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("ingest: commit rotate: %w", err)
	}
	stream.Key = newKey
	stream.UpdatedAt = newKey.CreatedAt
	return &stream, nil
}

// FindActiveStreamKey liefert den aktiven Key (`deactivated_at IS
// NULL`). `domain.ErrIngestKeyInvalid` bei keinem Treffer.
func (r *IngestStreamRepository) FindActiveStreamKey(ctx context.Context, projectID, streamID string) (domain.StreamKey, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT key_hash, fingerprint, created_at
FROM stream_keys
WHERE project_id = ? AND stream_id = ? AND deactivated_at IS NULL
LIMIT 1`, projectID, streamID)
	var hash, fingerprint, createdAtRaw string
	if err := row.Scan(&hash, &fingerprint, &createdAtRaw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.StreamKey{}, domain.ErrIngestKeyInvalid
		}
		return domain.StreamKey{}, fmt.Errorf("ingest: find active key: %w", err)
	}
	createdAt, _ := time.Parse(time.RFC3339Nano, createdAtRaw)
	return domain.StreamKey{
		Hash:        hash,
		Fingerprint: fingerprint,
		CreatedAt:   createdAt,
	}, nil
}

// GetEndpointByID liest einen Endpoint aus `ingest_endpoints`.
func (r *IngestStreamRepository) GetEndpointByID(ctx context.Context, endpointID string) (*domain.IngestEndpoint, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT endpoint_id, protocol, listen_host, listen_port, path_template, lab_stack, public_url_hint
FROM ingest_endpoints WHERE endpoint_id = ?`, endpointID)
	var protocol string
	endpoint := domain.IngestEndpoint{}
	if err := row.Scan(&endpoint.ID, &protocol, &endpoint.ListenHost, &endpoint.ListenPort, &endpoint.PathTemplate, &endpoint.LabStack, &endpoint.PublicURLHint); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrIngestEndpointNotFound
		}
		return nil, fmt.Errorf("ingest: get endpoint: %w", err)
	}
	endpoint.Protocol = domain.IngestProtocol(protocol)
	return &endpoint, nil
}

// GetTargetByID liest ein MediaServerTarget aus `media_server_targets`.
func (r *IngestStreamRepository) GetTargetByID(ctx context.Context, targetID string) (*domain.MediaServerTarget, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT target_id, kind, config_path, hls_url_template, control_api_url
FROM media_server_targets WHERE target_id = ?`, targetID)
	var kind string
	target := domain.MediaServerTarget{}
	if err := row.Scan(&target.ID, &kind, &target.ConfigPath, &target.HLSURLTemplate, &target.ControlAPIURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrIngestTargetNotFound
		}
		return nil, fmt.Errorf("ingest: get target: %w", err)
	}
	target.Kind = domain.MediaServerKind(kind)
	return &target, nil
}

// GetRoutingRuleByID liest die Routing-Regel des Streams.
func (r *IngestStreamRepository) GetRoutingRuleByID(ctx context.Context, projectID, streamID string) (*domain.RoutingRule, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT rule_id, stream_id, target_id, mode, enabled
FROM ingest_routing_rules WHERE project_id = ? AND stream_id = ?`, projectID, streamID)
	var enabled int
	var mode string
	rule := domain.RoutingRule{}
	if err := row.Scan(&rule.ID, &rule.StreamID, &rule.TargetID, &mode, &enabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrIngestStreamNotFound
		}
		return nil, fmt.Errorf("ingest: get routing rule: %w", err)
	}
	rule.Mode = domain.RoutingRuleMode(mode)
	rule.Enabled = enabled == 1
	return &rule, nil
}

// AppendLifecycleEvent persistiert ein Lifecycle-Event in
// `stream_lifecycle_events` (append-only).
func (r *IngestStreamRepository) AppendLifecycleEvent(ctx context.Context, event domain.StreamLifecycleEvent) error {
	if _, err := r.scanStream(ctx, r.db, event.ProjectID, event.StreamID); err != nil {
		return err
	}
	occurredAt := event.OccurredAt.UTC().Format(time.RFC3339Nano)
	receivedAt := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := r.db.ExecContext(ctx, `
INSERT INTO stream_lifecycle_events(project_id, stream_id, kind, occurred_at, received_at, source, key_fingerprint)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.ProjectID, event.StreamID, string(event.Kind), occurredAt, receivedAt, string(event.Source), event.KeyFingerprint); err != nil {
		return fmt.Errorf("ingest: insert lifecycle event: %w", err)
	}
	return nil
}

// SeedEndpoint / SeedTarget sind Boot-/Test-Helfer (Endpoint- und
// Target-Tabellen sind im `0.11.0`-Scope statisch konfiguriert; ein
// produktives Verwaltungs-API ist Folge-Scope).
func (r *IngestStreamRepository) SeedEndpoint(ctx context.Context, endpoint domain.IngestEndpoint) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO ingest_endpoints(endpoint_id, protocol, listen_host, listen_port, path_template, lab_stack, public_url_hint)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(endpoint_id) DO UPDATE SET
    protocol = excluded.protocol,
    listen_host = excluded.listen_host,
    listen_port = excluded.listen_port,
    path_template = excluded.path_template,
    lab_stack = excluded.lab_stack,
    public_url_hint = excluded.public_url_hint`,
		endpoint.ID, string(endpoint.Protocol), endpoint.ListenHost, endpoint.ListenPort,
		endpoint.PathTemplate, endpoint.LabStack, endpoint.PublicURLHint)
	if err != nil {
		return fmt.Errorf("ingest: seed endpoint: %w", err)
	}
	return nil
}

// SeedTarget legt ein MediaServerTarget an oder aktualisiert es. Boot-/
// Test-Helfer; produktive Verwaltung ist Folge-Scope.
func (r *IngestStreamRepository) SeedTarget(ctx context.Context, target domain.MediaServerTarget) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO media_server_targets(target_id, kind, config_path, hls_url_template, control_api_url)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(target_id) DO UPDATE SET
    kind = excluded.kind,
    config_path = excluded.config_path,
    hls_url_template = excluded.hls_url_template,
    control_api_url = excluded.control_api_url`,
		target.ID, string(target.Kind), target.ConfigPath, target.HLSURLTemplate, target.ControlAPIURL)
	if err != nil {
		return fmt.Errorf("ingest: seed target: %w", err)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanIngestStream(rs rowScanner) (domain.IngestStream, error) {
	var stream domain.IngestStream
	var protocol, status string
	var createdAtRaw, updatedAtRaw string
	if err := rs.Scan(&stream.ID, &stream.ProjectID, &stream.DisplayName, &protocol,
		&stream.EndpointID, &stream.TargetID, &stream.RoutingRuleID,
		&status, &createdAtRaw, &updatedAtRaw); err != nil {
		return domain.IngestStream{}, err
	}
	stream.Protocol = domain.IngestProtocol(protocol)
	stream.Status = domain.IngestStreamStatus(status)
	stream.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAtRaw)
	stream.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAtRaw)
	return stream, nil
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func (r *IngestStreamRepository) scanStream(ctx context.Context, q queryRower, projectID, streamID string) (domain.IngestStream, error) {
	row := q.QueryRowContext(ctx, `
SELECT stream_id, project_id, display_name, protocol, endpoint_id, target_id, routing_rule_id, status, created_at, updated_at
FROM ingest_streams
WHERE project_id = ? AND stream_id = ?`, projectID, streamID)
	stream, err := scanIngestStream(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.IngestStream{}, domain.ErrIngestStreamNotFound
		}
		return domain.IngestStream{}, fmt.Errorf("ingest: scan stream: %w", err)
	}
	return stream, nil
}

func newIngestID(prefix string) string {
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic("sqlite: ingest-control id rng failed: " + err.Error())
	}
	return prefix + hex.EncodeToString(raw[:])
}
