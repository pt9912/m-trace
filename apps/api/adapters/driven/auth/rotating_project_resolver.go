package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// RotatingProjectResolver kombiniert den `0.12.0`-Project-Token-
// Generationen-Pfad (RAK-73) mit dem Legacy-Static-Resolver. Tokens
// mit dem `mtr_pt_*`-Prefix laufen über das `ProjectTokenRepository`
// und werden gegen Lifecycle-Status (active/grace/expired/revoked/
// not_yet_valid) evaluiert; Tokens ohne diesen Prefix werden über
// die bisherige Static-Map aufgelöst (Backward-Compat für
// `demo-token` und vergleichbare Lab-Werte).
//
// Die Reihenfolge ist deterministisch:
//
//   1. Wenn `token` mit `mtr_pt_` startet → Repo-Lookup.
//   2. Sonst → Fallback auf den Static-Resolver.
//
// `domain.ErrAuthTokenInvalid` für unbekannte/malformed Tokens auf
// dem Repo-Pfad; auf dem Static-Pfad bleibt das bisherige
// `domain.ErrUnauthorized` als Fehler-Code.
type RotatingProjectResolver struct {
	Repo            driven.ProjectTokenRepository
	Fallback        driven.ProjectResolver
	ProjectsByID    ProjectByIDLookup
	Now             func() time.Time
}

// ProjectByIDLookup bridge auf den `0.12.0`-Bedarf, das aufgelöste
// Project anhand der ID aus dem Static-Resolver zu materialisieren.
// `StaticProjectResolver` implementiert das über `ResolveByID`.
type ProjectByIDLookup interface {
	ResolveByID(projectID string) (domain.Project, error)
}

// NewRotatingProjectResolver konstruiert den Resolver. `repo`,
// `fallback` und `projects` sind Pflicht — fehlt einer, fällt der
// Resolver für den jeweiligen Pfad auf den Fehlerfall zurück.
func NewRotatingProjectResolver(repo driven.ProjectTokenRepository, fallback driven.ProjectResolver, projects ProjectByIDLookup) *RotatingProjectResolver {
	return &RotatingProjectResolver{
		Repo:         repo,
		Fallback:     fallback,
		ProjectsByID: projects,
		Now:          func() time.Time { return time.Now().UTC() },
	}
}

// Compile-time check.
var _ driven.ProjectResolver = (*RotatingProjectResolver)(nil)

// ResolveByToken implementiert `driven.ProjectResolver`.
//
// Lifecycle-Fehler (`auth_token_revoked`/`_expired`/`_not_yet_valid`)
// werden 1:1 propagiert — der HTTP-Adapter mappt sie auf §3.9-Codes.
// Unbekannte Hashes oder fehlgeleitete Project-IDs liefern
// `domain.ErrAuthTokenInvalid`, ohne Hinweis auf Existenz.
func (r *RotatingProjectResolver) ResolveByToken(ctx context.Context, token string) (domain.Project, error) {
	if domain.HasProjectTokenPrefix(token) {
		return r.resolveRotatingToken(ctx, token)
	}
	if r.Fallback == nil {
		return domain.Project{}, domain.ErrUnauthorized
	}
	return r.Fallback.ResolveByToken(ctx, token)
}

func (r *RotatingProjectResolver) resolveRotatingToken(ctx context.Context, token string) (domain.Project, error) {
	if r.Repo == nil || r.ProjectsByID == nil {
		return domain.Project{}, domain.ErrAuthTokenInvalid
	}
	hash := hashRotatingToken(token)
	gen, err := r.Repo.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, driven.ErrProjectTokenNotFound) {
			return domain.Project{}, domain.ErrAuthTokenInvalid
		}
		return domain.Project{}, err
	}
	status := domain.EvaluateProjectTokenStatus(gen, r.now())
	if statusErr := domain.StatusToAuthError(status); statusErr != nil {
		return domain.Project{}, statusErr
	}
	project, err := r.ProjectsByID.ResolveByID(gen.ProjectID)
	if err != nil {
		// Project ist nicht (mehr) konfiguriert — Token zeigt auf ein
		// Ghost-Project. `auth_token_invalid` ist konservativer als
		// `_revoked`, weil es keinen Hinweis auf Existenz gibt.
		return domain.Project{}, domain.ErrAuthTokenInvalid
	}
	return project, nil
}

// hashRotatingToken berechnet den SHA-256-Hex-Hash analog zur Domain-
// Implementierung (`domain.GenerateProjectToken`).
func hashRotatingToken(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (r *RotatingProjectResolver) now() time.Time {
	if r.Now == nil {
		return time.Now().UTC()
	}
	return r.Now()
}
