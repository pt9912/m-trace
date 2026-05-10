package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// ProjectTokenRepository ist die durable SQLite-Variante des
// `driven.ProjectTokenRepository`-Ports (`0.12.0`, RAK-73). Tabellen
// aus `V4__project_tokens.sql`.
//
// Sicherheitsprofil:
//   - Klartext-Tokens werden nicht persistiert. Persistente Sicht
//     enthält ausschließlich `key_hash`, `fingerprint` und Lifecycle-
//     Metadaten.
//   - `key_hash` ist repository-weit unique (UNIQUE-Constraint),
//     damit Cross-Project-Hash-Kollisionen früh sichtbar sind.
//   - `Find` liefert eine defensive Kopie inklusive Time-Pointer-
//     Klone, damit nachträgliche Mutationen am Caller den DB-State
//     nicht beeinflussen.
type ProjectTokenRepository struct {
	db *sql.DB
}

// NewProjectTokenRepository konstruiert den Adapter.
func NewProjectTokenRepository(db *sql.DB) *ProjectTokenRepository {
	return &ProjectTokenRepository{db: db}
}

// Compile-time check.
var _ driven.ProjectTokenRepository = (*ProjectTokenRepository)(nil)

// Create persistiert eine neue Generation. Setzt das Project bei
// Bedarf (analog `IngestStreamRepository.CreateStream`), damit der
// Foreign-Key auf `projects` greift.
func (r *ProjectTokenRepository) Create(ctx context.Context, gen domain.ProjectTokenGeneration) error {
	if gen.ProjectID == "" || gen.TokenID == "" || gen.KeyHash == "" {
		return domain.ErrAuthTokenInvalid
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("project token: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `INSERT INTO projects(project_id) VALUES (?) ON CONFLICT(project_id) DO NOTHING`, gen.ProjectID); err != nil {
		return fmt.Errorf("project token: ensure project: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO project_token_generations(
    token_id, project_id, key_hash, fingerprint,
    not_before, grace_until, expires_at, revoked_at,
    created_at, rotated_from)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		gen.TokenID, gen.ProjectID, gen.KeyHash, gen.Fingerprint,
		formatProjectTokenTime(gen.NotBefore),
		nullableProjectTokenTime(gen.GraceUntil),
		nullableProjectTokenTime(gen.ExpiresAt),
		nullableProjectTokenTime(gen.RevokedAt),
		formatProjectTokenTime(gen.CreatedAt),
		nullableProjectTokenString(gen.RotatedFrom),
	)
	if err != nil {
		return mapProjectTokenCreateError(err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("project token: commit: %w", err)
	}
	return nil
}

// ListByProject liefert alle Generationen des Projects, sortiert
// nach `created_at` aufsteigend.
func (r *ProjectTokenRepository) ListByProject(ctx context.Context, projectID string) ([]domain.ProjectTokenGeneration, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT token_id, project_id, key_hash, fingerprint, not_before, grace_until, expires_at, revoked_at, created_at, rotated_from
FROM project_token_generations
WHERE project_id = ?
ORDER BY created_at, token_id`, projectID)
	if err != nil {
		return nil, fmt.Errorf("project token: list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []domain.ProjectTokenGeneration
	for rows.Next() {
		gen, err := scanGeneration(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, gen)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("project token: list rows: %w", err)
	}
	return out, nil
}

// FindByHash liefert die Generation mit dem angegebenen Hash.
func (r *ProjectTokenRepository) FindByHash(ctx context.Context, keyHash string) (domain.ProjectTokenGeneration, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT token_id, project_id, key_hash, fingerprint, not_before, grace_until, expires_at, revoked_at, created_at, rotated_from
FROM project_token_generations
WHERE key_hash = ?`, keyHash)
	gen, err := scanGeneration(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ProjectTokenGeneration{}, driven.ErrProjectTokenNotFound
	}
	if err != nil {
		return domain.ProjectTokenGeneration{}, err
	}
	return gen, nil
}

// SetGraceUntil aktualisiert das `grace_until`-Feld; idempotent.
func (r *ProjectTokenRepository) SetGraceUntil(ctx context.Context, projectID, tokenID string, graceUntil time.Time) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE project_token_generations
SET grace_until = ?
WHERE project_id = ? AND token_id = ?`,
		formatProjectTokenTime(graceUntil), projectID, tokenID)
	if err != nil {
		return fmt.Errorf("project token: set grace: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return driven.ErrProjectTokenNotFound
	}
	return nil
}

// Revoke setzt `revoked_at`; idempotent.
func (r *ProjectTokenRepository) Revoke(ctx context.Context, projectID, tokenID string, revokedAt time.Time) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE project_token_generations
SET revoked_at = ?
WHERE project_id = ? AND token_id = ?`,
		formatProjectTokenTime(revokedAt), projectID, tokenID)
	if err != nil {
		return fmt.Errorf("project token: revoke: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return driven.ErrProjectTokenNotFound
	}
	return nil
}

// scanGeneration liest eine Row in `domain.ProjectTokenGeneration`.
// Nutzt das bestehende `rowScanner`-Interface aus
// `ingest_stream_repository.go` — beide Repos teilen das Interface,
// damit der Linter keine redundante Definition meldet.
func scanGeneration(s rowScanner) (domain.ProjectTokenGeneration, error) {
	var (
		tokenID, projectID, keyHash, fingerprint string
		notBefore, createdAt                     string
		graceUntil, expiresAt, revokedAt         sql.NullString
		rotatedFrom                              sql.NullString
	)
	if err := s.Scan(&tokenID, &projectID, &keyHash, &fingerprint, &notBefore, &graceUntil, &expiresAt, &revokedAt, &createdAt, &rotatedFrom); err != nil {
		return domain.ProjectTokenGeneration{}, err
	}
	gen := domain.ProjectTokenGeneration{
		TokenID:     tokenID,
		ProjectID:   projectID,
		KeyHash:     keyHash,
		Fingerprint: fingerprint,
	}
	t, err := time.Parse(time.RFC3339Nano, notBefore)
	if err != nil {
		return domain.ProjectTokenGeneration{}, fmt.Errorf("project token: parse not_before: %w", err)
	}
	gen.NotBefore = t
	c, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.ProjectTokenGeneration{}, fmt.Errorf("project token: parse created_at: %w", err)
	}
	gen.CreatedAt = c
	if graceUntil.Valid {
		t, err := time.Parse(time.RFC3339Nano, graceUntil.String)
		if err != nil {
			return domain.ProjectTokenGeneration{}, fmt.Errorf("project token: parse grace_until: %w", err)
		}
		gen.GraceUntil = &t
	}
	if expiresAt.Valid {
		t, err := time.Parse(time.RFC3339Nano, expiresAt.String)
		if err != nil {
			return domain.ProjectTokenGeneration{}, fmt.Errorf("project token: parse expires_at: %w", err)
		}
		gen.ExpiresAt = &t
	}
	if revokedAt.Valid {
		t, err := time.Parse(time.RFC3339Nano, revokedAt.String)
		if err != nil {
			return domain.ProjectTokenGeneration{}, fmt.Errorf("project token: parse revoked_at: %w", err)
		}
		gen.RevokedAt = &t
	}
	if rotatedFrom.Valid {
		v := rotatedFrom.String
		gen.RotatedFrom = &v
	}
	return gen, nil
}

// formatProjectTokenTime formatiert eine Zeit als RFC3339Nano in
// UTC. Pinnt den Persistenz-Format-Vertrag analog zu
// `ingest_stream_repository`.
func formatProjectTokenTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// nullableProjectTokenTime produziert sql.NullString für optionale
// Time-Pointer.
func nullableProjectTokenTime(t *time.Time) sql.NullString {
	if t == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: t.UTC().Format(time.RFC3339Nano), Valid: true}
}

// nullableProjectTokenString produziert sql.NullString für optionale
// String-Pointer.
func nullableProjectTokenString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// mapProjectTokenCreateError unterscheidet Unique-Verletzungen
// (Hash- oder Token-ID-Kollision) von echten DB-Fehlern.
func mapProjectTokenCreateError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "UNIQUE") || strings.Contains(msg, "PRIMARY KEY") {
		return domain.ErrAuthTokenInvalid
	}
	return fmt.Errorf("project token: create: %w", err)
}
