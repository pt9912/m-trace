package inmemory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// ProjectTokenRepository ist die in-memory-Variante des
// `driven.ProjectTokenRepository`-Ports (RAK-73). Lebt für
// Tests und den Lab-Inmemory-Modus; produktiver Pfad ist die SQLite-
// Variante.
type ProjectTokenRepository struct {
	mu          sync.RWMutex
	byTokenID   map[string]map[string]domain.ProjectTokenGeneration // projectID → tokenID → gen
	byHash      map[string]projectTokenAddress                       // hash → (projectID, tokenID)
}

type projectTokenAddress struct {
	ProjectID string
	TokenID   string
}

// NewProjectTokenRepository konstruiert das leere Repo.
func NewProjectTokenRepository() *ProjectTokenRepository {
	return &ProjectTokenRepository{
		byTokenID: make(map[string]map[string]domain.ProjectTokenGeneration),
		byHash:    make(map[string]projectTokenAddress),
	}
}

// Compile-time check.
var _ driven.ProjectTokenRepository = (*ProjectTokenRepository)(nil)

// Create persistiert eine neue Token-Generation. Der Hash muss
// repository-weit unique sein (analog zum SQLite-Constraint).
func (r *ProjectTokenRepository) Create(_ context.Context, gen domain.ProjectTokenGeneration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if gen.ProjectID == "" || gen.TokenID == "" || gen.KeyHash == "" {
		return domain.ErrAuthTokenInvalid
	}
	if _, exists := r.byHash[gen.KeyHash]; exists {
		return domain.ErrAuthTokenInvalid
	}
	bucket := r.byTokenID[gen.ProjectID]
	if bucket == nil {
		bucket = make(map[string]domain.ProjectTokenGeneration)
		r.byTokenID[gen.ProjectID] = bucket
	}
	if _, exists := bucket[gen.TokenID]; exists {
		return domain.ErrAuthTokenInvalid
	}
	bucket[gen.TokenID] = cloneGeneration(gen)
	r.byHash[gen.KeyHash] = projectTokenAddress{ProjectID: gen.ProjectID, TokenID: gen.TokenID}
	return nil
}

// ListByProject liefert alle Generationen eines Projects in stabiler
// Reihenfolge (nach `CreatedAt`, dann `TokenID`).
func (r *ProjectTokenRepository) ListByProject(_ context.Context, projectID string) ([]domain.ProjectTokenGeneration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bucket := r.byTokenID[projectID]
	out := make([]domain.ProjectTokenGeneration, 0, len(bucket))
	for _, g := range bucket {
		out = append(out, cloneGeneration(g))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].TokenID < out[j].TokenID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out, nil
}

// FindByHash liefert die Generation zum Hash. `ErrProjectTokenNotFound`
// für unbekannte Hashes.
func (r *ProjectTokenRepository) FindByHash(_ context.Context, keyHash string) (domain.ProjectTokenGeneration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	addr, ok := r.byHash[keyHash]
	if !ok {
		return domain.ProjectTokenGeneration{}, driven.ErrProjectTokenNotFound
	}
	bucket, ok := r.byTokenID[addr.ProjectID]
	if !ok {
		return domain.ProjectTokenGeneration{}, driven.ErrProjectTokenNotFound
	}
	gen, ok := bucket[addr.TokenID]
	if !ok {
		return domain.ProjectTokenGeneration{}, driven.ErrProjectTokenNotFound
	}
	return cloneGeneration(gen), nil
}

// SetGraceUntil aktualisiert das `GraceUntil`-Feld; idempotent.
func (r *ProjectTokenRepository) SetGraceUntil(_ context.Context, projectID, tokenID string, graceUntil time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bucket, ok := r.byTokenID[projectID]
	if !ok {
		return driven.ErrProjectTokenNotFound
	}
	gen, ok := bucket[tokenID]
	if !ok {
		return driven.ErrProjectTokenNotFound
	}
	gu := graceUntil
	gen.GraceUntil = &gu
	bucket[tokenID] = gen
	return nil
}

// Revoke setzt `RevokedAt`; idempotent.
func (r *ProjectTokenRepository) Revoke(_ context.Context, projectID, tokenID string, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bucket, ok := r.byTokenID[projectID]
	if !ok {
		return driven.ErrProjectTokenNotFound
	}
	gen, ok := bucket[tokenID]
	if !ok {
		return driven.ErrProjectTokenNotFound
	}
	rv := revokedAt
	gen.RevokedAt = &rv
	bucket[tokenID] = gen
	return nil
}

// cloneGeneration produziert eine defensive Kopie inklusive
// Time-Pointer, damit Caller-Mutationen den gespeicherten State nicht
// verändern.
func cloneGeneration(g domain.ProjectTokenGeneration) domain.ProjectTokenGeneration {
	out := g
	if g.GraceUntil != nil {
		v := *g.GraceUntil
		out.GraceUntil = &v
	}
	if g.ExpiresAt != nil {
		v := *g.ExpiresAt
		out.ExpiresAt = &v
	}
	if g.RevokedAt != nil {
		v := *g.RevokedAt
		out.RevokedAt = &v
	}
	if g.RotatedFrom != nil {
		v := *g.RotatedFrom
		out.RotatedFrom = &v
	}
	return out
}
