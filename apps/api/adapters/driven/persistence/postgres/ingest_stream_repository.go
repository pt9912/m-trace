package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// pgForeignKeyViolation ist der SQLSTATE-Code für eine Foreign-Key-
// Verletzung (Pendant zur SQLite-Fehlerstring-Prüfung "FOREIGN KEY").
const pgForeignKeyViolation = "23503"

// IngestStreamRepository ist die Postgres-Variante des
// driven.IngestStreamRepository-Ports (ADR-0006, Dialekt-Spiegel des
// SQLite-Adapters). Query-Konstanten sind mit dem SQLite-Adapter
// identisch (dialekt-neutrale `?`-Platzhalter); rebind() übersetzt sie
// zur Laufzeit auf `$n`. Die Spalten-Typen sind dank reversiertem Schema
// gleich (datetime-als-TEXT/RFC3339Nano, enabled/listen_port als INTEGER).
//
// Sicherheitsprofil (identisch zum SQLite-Adapter):
//   - Klartext-Stream-Keys werden nicht persistiert. `stream_keys`
//     speichert ausschließlich `key_hash` und `fingerprint`.
//   - `idx_stream_keys_active_unique` erzwingt einen einzigen aktiven
//     Key pro `(project_id, key_hash)`; rotierte Keys bleiben für Audit.
//   - Cross-Project-Lookups liefern `domain.ErrIngestStreamNotFound`
//     ohne Hinweis auf Existenz.
//
// Einziger echter Dialekt-Unterschied ist die Constraint-Verletzungs-
// Erkennung über den SQLSTATE-Code (23505/23503 + ConstraintName) statt
// über den SQLite-Fehlerstring. Endpoint-/Target-Existenz wird — wie im
// SQLite-Adapter — über die Vorab-Lookups in CreateStream geprüft
// (`ingest_streams` trägt keine endpoint/target-FK); die 23503-Branches
// bleiben defensiv gespiegelt.
type IngestStreamRepository struct {
	db *sql.DB
}

// NewIngestStreamRepository konstruiert den Adapter.
func NewIngestStreamRepository(db *sql.DB) *IngestStreamRepository {
	return &IngestStreamRepository{db: db}
}

var _ driven.IngestStreamRepository = (*IngestStreamRepository)(nil)

// CreateStream legt Project, Stream, Routing-Regel und initialen Key in
// einer Transaktion an. Endpoint/Target werden vorab verifiziert
// (ErrIngestEndpointNotFound / ErrIngestTargetNotFound), bevor die Tx
// beginnt — identisch zum SQLite-Pfad.
func (r *IngestStreamRepository) CreateStream(ctx context.Context, input driven.CreateStreamInput) (*domain.IngestStream, error) {
	if _, err := r.GetEndpointByID(ctx, input.EndpointID); err != nil {
		return nil, err
	}
	if _, err := r.GetTargetByID(ctx, input.TargetID); err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ingest-postgres: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		rebind(`INSERT INTO projects(project_id) VALUES (?) ON CONFLICT(project_id) DO NOTHING`),
		input.ProjectID); err != nil {
		return nil, fmt.Errorf("ingest-postgres: ensure project: %w", err)
	}

	streamID := newIngestID("ing_")
	ruleID := newIngestID("route_")
	keyID := newIngestID("key_")
	createdAt := input.CreatedAt.UTC().Format(time.RFC3339Nano)

	_, err = tx.ExecContext(ctx, rebind(`
INSERT INTO ingest_streams(stream_id, project_id, display_name, protocol, endpoint_id, target_id, routing_rule_id, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
		streamID, input.ProjectID, input.DisplayName, string(input.Protocol),
		input.EndpointID, input.TargetID, ruleID, string(domain.IngestStreamStatusReady),
		createdAt, createdAt)
	if err != nil {
		return nil, mapIngestStreamCreateError(err)
	}

	if _, err := tx.ExecContext(ctx, rebind(`
INSERT INTO ingest_routing_rules(rule_id, project_id, stream_id, target_id, mode, enabled)
VALUES (?, ?, ?, ?, ?, 1)`),
		ruleID, input.ProjectID, streamID, input.TargetID, string(domain.RoutingRuleModeSingle)); err != nil {
		return nil, fmt.Errorf("ingest-postgres: insert routing rule: %w", err)
	}

	if _, err := tx.ExecContext(ctx, rebind(`
INSERT INTO stream_keys(key_id, project_id, stream_id, key_hash, fingerprint, created_at)
VALUES (?, ?, ?, ?, ?, ?)`),
		keyID, input.ProjectID, streamID, input.InitialKey.Hash, input.InitialKey.Fingerprint, createdAt); err != nil {
		return nil, fmt.Errorf("ingest-postgres: insert initial key: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("ingest-postgres: commit: %w", err)
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

// mapIngestStreamCreateError übersetzt Constraint-Verletzungen des
// ingest_streams-INSERT auf Domain-Fehler. Der aktive-display_name-
// Unique-Index (`idx_ingest_streams_active_display_name`) ist der einzige
// Unique-Constraint auf dieser Tabelle → 23505 = Name-Konflikt. Die
// 23503-Branches sind defensiv gespiegelt (heute existiert keine
// endpoint/target-FK; Existenz prüft der Vorab-Lookup in CreateStream).
func mapIngestStreamCreateError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			if strings.Contains(pgErr.ConstraintName, "display_name") {
				return domain.ErrIngestStreamNameConflict
			}
		case pgForeignKeyViolation:
			switch {
			case strings.Contains(pgErr.ConstraintName, "endpoint"):
				return domain.ErrIngestEndpointNotFound
			case strings.Contains(pgErr.ConstraintName, "target"):
				return domain.ErrIngestTargetNotFound
			}
		}
	}
	return fmt.Errorf("ingest-postgres: insert stream: %w", err)
}

// GetStreamByID liefert einen Stream inkl. aktivem Key. Cross-Project-
// Lookups liefern `domain.ErrIngestStreamNotFound`.
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
	rows, err := r.db.QueryContext(ctx, rebind(`
SELECT stream_id, project_id, display_name, protocol, endpoint_id, target_id, routing_rule_id, status, created_at, updated_at
FROM ingest_streams
WHERE project_id = ?
ORDER BY created_at DESC, stream_id ASC`), projectID)
	if err != nil {
		return nil, fmt.Errorf("ingest-postgres: list streams: %w", err)
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
		return nil, fmt.Errorf("ingest-postgres: list streams iter: %w", err)
	}
	return out, nil
}

// RotateKey deaktiviert den bisher aktiven Stream-Key und fügt den neuen
// Key ein. Atomare Tx über die drei Schritte.
func (r *IngestStreamRepository) RotateKey(ctx context.Context, projectID, streamID string, newKey domain.StreamKey) (*domain.IngestStream, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ingest-postgres: rotate begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stream, err := r.scanStream(ctx, tx, projectID, streamID)
	if err != nil {
		return nil, err
	}

	deactivatedAt := newKey.CreatedAt.UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, rebind(`
UPDATE stream_keys
SET deactivated_at = ?
WHERE project_id = ? AND stream_id = ? AND deactivated_at IS NULL`),
		deactivatedAt, projectID, streamID); err != nil {
		return nil, fmt.Errorf("ingest-postgres: deactivate old key: %w", err)
	}

	keyID := newIngestID("key_")
	if _, err := tx.ExecContext(ctx, rebind(`
INSERT INTO stream_keys(key_id, project_id, stream_id, key_hash, fingerprint, created_at)
VALUES (?, ?, ?, ?, ?, ?)`),
		keyID, projectID, streamID, newKey.Hash, newKey.Fingerprint, deactivatedAt); err != nil {
		return nil, fmt.Errorf("ingest-postgres: insert rotated key: %w", err)
	}

	if _, err := tx.ExecContext(ctx, rebind(`
UPDATE ingest_streams
SET updated_at = ?
WHERE project_id = ? AND stream_id = ?`),
		deactivatedAt, projectID, streamID); err != nil {
		return nil, fmt.Errorf("ingest-postgres: bump updated_at: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("ingest-postgres: commit rotate: %w", err)
	}
	stream.Key = newKey
	stream.UpdatedAt = newKey.CreatedAt
	return &stream, nil
}

// FindActiveStreamKey liefert den aktiven Key (`deactivated_at IS NULL`).
// `domain.ErrIngestKeyInvalid` bei keinem Treffer.
func (r *IngestStreamRepository) FindActiveStreamKey(ctx context.Context, projectID, streamID string) (domain.StreamKey, error) {
	row := r.db.QueryRowContext(ctx, rebind(`
SELECT key_hash, fingerprint, created_at
FROM stream_keys
WHERE project_id = ? AND stream_id = ? AND deactivated_at IS NULL
LIMIT 1`), projectID, streamID)
	var hash, fingerprint, createdAtRaw string
	if err := row.Scan(&hash, &fingerprint, &createdAtRaw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.StreamKey{}, domain.ErrIngestKeyInvalid
		}
		return domain.StreamKey{}, fmt.Errorf("ingest-postgres: find active key: %w", err)
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
	row := r.db.QueryRowContext(ctx, rebind(`
SELECT endpoint_id, protocol, listen_host, listen_port, path_template, lab_stack, public_url_hint
FROM ingest_endpoints WHERE endpoint_id = ?`), endpointID)
	var protocol string
	endpoint := domain.IngestEndpoint{}
	if err := row.Scan(&endpoint.ID, &protocol, &endpoint.ListenHost, &endpoint.ListenPort, &endpoint.PathTemplate, &endpoint.LabStack, &endpoint.PublicURLHint); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrIngestEndpointNotFound
		}
		return nil, fmt.Errorf("ingest-postgres: get endpoint: %w", err)
	}
	endpoint.Protocol = domain.IngestProtocol(protocol)
	return &endpoint, nil
}

// GetTargetByID liest ein MediaServerTarget aus `media_server_targets`.
func (r *IngestStreamRepository) GetTargetByID(ctx context.Context, targetID string) (*domain.MediaServerTarget, error) {
	row := r.db.QueryRowContext(ctx, rebind(`
SELECT target_id, kind, config_path, hls_url_template, control_api_url
FROM media_server_targets WHERE target_id = ?`), targetID)
	var kind string
	target := domain.MediaServerTarget{}
	if err := row.Scan(&target.ID, &kind, &target.ConfigPath, &target.HLSURLTemplate, &target.ControlAPIURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrIngestTargetNotFound
		}
		return nil, fmt.Errorf("ingest-postgres: get target: %w", err)
	}
	target.Kind = domain.MediaServerKind(kind)
	return &target, nil
}

// GetRoutingRuleByID liest die Routing-Regel des Streams.
func (r *IngestStreamRepository) GetRoutingRuleByID(ctx context.Context, projectID, streamID string) (*domain.RoutingRule, error) {
	row := r.db.QueryRowContext(ctx, rebind(`
SELECT rule_id, stream_id, target_id, mode, enabled
FROM ingest_routing_rules WHERE project_id = ? AND stream_id = ?`), projectID, streamID)
	var enabled int
	var mode string
	rule := domain.RoutingRule{}
	if err := row.Scan(&rule.ID, &rule.StreamID, &rule.TargetID, &mode, &enabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrIngestStreamNotFound
		}
		return nil, fmt.Errorf("ingest-postgres: get routing rule: %w", err)
	}
	rule.Mode = domain.RoutingRuleMode(mode)
	rule.Enabled = enabled == 1
	return &rule, nil
}

// AppendLifecycleEvent persistiert ein Lifecycle-Event in
// `stream_lifecycle_events` (append-only). Der Aufrufer muss `EventID`
// setzen — der Service generiert ihn beim Empfang des Hooks (RAK-69).
func (r *IngestStreamRepository) AppendLifecycleEvent(ctx context.Context, event domain.StreamLifecycleEvent) error {
	if _, err := r.scanStream(ctx, r.db, event.ProjectID, event.StreamID); err != nil {
		return err
	}
	if event.EventID == "" {
		return fmt.Errorf("ingest-postgres: append lifecycle event: empty event_id")
	}
	occurredAt := event.OccurredAt.UTC().Format(time.RFC3339Nano)
	receivedAt := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := r.db.ExecContext(ctx, rebind(`
INSERT INTO stream_lifecycle_events(event_id, project_id, stream_id, kind, occurred_at, received_at, source, key_fingerprint, connection_id, reason)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
		event.EventID, event.ProjectID, event.StreamID, string(event.Kind),
		occurredAt, receivedAt, string(event.Source), event.KeyFingerprint,
		event.ConnectionID, event.Reason); err != nil {
		return fmt.Errorf("ingest-postgres: insert lifecycle event: %w", err)
	}
	return nil
}

// SeedEndpoint / SeedTarget sind Boot-/Test-Helfer (Endpoint- und
// Target-Tabellen sind im Scope statisch konfiguriert; ein produktives
// Verwaltungs-API ist Folge-Scope). Nicht Teil des Port-Vertrags.
func (r *IngestStreamRepository) SeedEndpoint(ctx context.Context, endpoint domain.IngestEndpoint) error {
	_, err := r.db.ExecContext(ctx, rebind(`
INSERT INTO ingest_endpoints(endpoint_id, protocol, listen_host, listen_port, path_template, lab_stack, public_url_hint)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(endpoint_id) DO UPDATE SET
    protocol = excluded.protocol,
    listen_host = excluded.listen_host,
    listen_port = excluded.listen_port,
    path_template = excluded.path_template,
    lab_stack = excluded.lab_stack,
    public_url_hint = excluded.public_url_hint`),
		endpoint.ID, string(endpoint.Protocol), endpoint.ListenHost, endpoint.ListenPort,
		endpoint.PathTemplate, endpoint.LabStack, endpoint.PublicURLHint)
	if err != nil {
		return fmt.Errorf("ingest-postgres: seed endpoint: %w", err)
	}
	return nil
}

// SeedTarget legt ein MediaServerTarget an oder aktualisiert es. Boot-/
// Test-Helfer; produktive Verwaltung ist Folge-Scope. Nicht Teil des
// Port-Vertrags.
func (r *IngestStreamRepository) SeedTarget(ctx context.Context, target domain.MediaServerTarget) error {
	_, err := r.db.ExecContext(ctx, rebind(`
INSERT INTO media_server_targets(target_id, kind, config_path, hls_url_template, control_api_url)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(target_id) DO UPDATE SET
    kind = excluded.kind,
    config_path = excluded.config_path,
    hls_url_template = excluded.hls_url_template,
    control_api_url = excluded.control_api_url`),
		target.ID, string(target.Kind), target.ConfigPath, target.HLSURLTemplate, target.ControlAPIURL)
	if err != nil {
		return fmt.Errorf("ingest-postgres: seed target: %w", err)
	}
	return nil
}

// queryRower abstrahiert `*sql.DB` und `*sql.Tx` über ihre gemeinsame
// QueryRowContext-Signatur, damit scanStream beide Quellen bedient.
type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (r *IngestStreamRepository) scanStream(ctx context.Context, q queryRower, projectID, streamID string) (domain.IngestStream, error) {
	row := q.QueryRowContext(ctx, rebind(`
SELECT stream_id, project_id, display_name, protocol, endpoint_id, target_id, routing_rule_id, status, created_at, updated_at
FROM ingest_streams
WHERE project_id = ? AND stream_id = ?`), projectID, streamID)
	stream, err := scanIngestStream(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.IngestStream{}, domain.ErrIngestStreamNotFound
		}
		return domain.IngestStream{}, fmt.Errorf("ingest-postgres: scan stream: %w", err)
	}
	return stream, nil
}

// scanIngestStream liest eine Row in `domain.IngestStream` (ohne Key).
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

// newIngestID erzeugt eine präfixierte Ingest-Control-ID (12 Byte
// crypto/rand, hex). Panic bei RNG-Ausfall — ein nicht funktionsfähiger
// Zufallszahlengenerator ist ein nicht behebbarer Prozesszustand
// (identisch zum SQLite-Adapter).
func newIngestID(prefix string) string {
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic("postgres: ingest-control id rng failed: " + err.Error())
	}
	return prefix + hex.EncodeToString(raw[:])
}
