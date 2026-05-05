package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// plan-0.4.0 §4.5 E1 — Statusmatrix-Tests für den Session-Link-
// Resolver im AnalyzeManifestUseCase. Spec-Anker:
// `spec/backend-api-contract.md` §3.6 + plan-0.4.0 §4.5.

// linkSessionRepo ist ein minimaler driven.SessionRepository-Stub für
// die Statusmatrix. Get/GetByCorrelationID werden gegen vor-konfigurierte
// Maps aufgelöst; alle anderen Methoden sind no-op.
type linkSessionRepo struct {
	byCID    map[linkKey]domain.StreamSession
	bySID    map[linkKey]domain.StreamSession
	getCalls int
	cidCalls int
}

type linkKey struct{ projectID, value string }

func (r *linkSessionRepo) Get(_ context.Context, projectID, sessionID string) (domain.StreamSession, error) {
	r.getCalls++
	if s, ok := r.bySID[linkKey{projectID, sessionID}]; ok {
		return s, nil
	}
	return domain.StreamSession{}, domain.ErrSessionNotFound
}

func (r *linkSessionRepo) GetByCorrelationID(_ context.Context, projectID, correlationID string) (domain.StreamSession, error) {
	r.cidCalls++
	if correlationID == "" {
		return domain.StreamSession{}, domain.ErrSessionNotFound
	}
	if s, ok := r.byCID[linkKey{projectID, correlationID}]; ok {
		return s, nil
	}
	return domain.StreamSession{}, domain.ErrSessionNotFound
}

func (r *linkSessionRepo) UpsertFromEvents(_ context.Context, _ []domain.PlaybackEvent) (map[string]string, error) {
	return map[string]string{}, nil
}
func (r *linkSessionRepo) List(_ context.Context, _ driven.SessionListQuery) (driven.SessionPage, error) {
	return driven.SessionPage{}, nil
}
func (r *linkSessionRepo) Sweep(_ context.Context, _ time.Time, _, _ time.Duration) error {
	return nil
}
func (r *linkSessionRepo) CountByState(_ context.Context, _ domain.SessionState) (int64, error) {
	return 0, nil
}
func (r *linkSessionRepo) AppendBoundaries(_ context.Context, _ []domain.SessionBoundary) error {
	return nil
}
func (r *linkSessionRepo) ListBoundariesForSession(_ context.Context, _, _ string) ([]domain.SessionBoundary, error) {
	return nil, nil
}

func newLinkRepo(sessions ...domain.StreamSession) *linkSessionRepo {
	repo := &linkSessionRepo{
		byCID: make(map[linkKey]domain.StreamSession),
		bySID: make(map[linkKey]domain.StreamSession),
	}
	for _, s := range sessions {
		repo.bySID[linkKey{s.ProjectID, s.ID}] = s
		if s.CorrelationID != "" {
			repo.byCID[linkKey{s.ProjectID, s.CorrelationID}] = s
		}
	}
	return repo
}

func validAnalyzeRequest() domain.StreamAnalysisRequest {
	return domain.StreamAnalysisRequest{ManifestText: "#EXTM3U\n"}
}

func okAnalyzer() *fakeAnalyzer {
	return &fakeAnalyzer{result: domain.StreamAnalysisResult{
		AnalyzerVersion: "0.4.0", PlaylistType: domain.PlaylistTypeMedia,
	}}
}

const (
	demoProject = "demo"
	demoCID     = "2f6f1a3c-9fb9-4c0b-a78f-2f41d8f6e1e7"
	demoSID     = "01J7K9X4Z2QHB6V3WS5R8Y4D1F"
)

func demoSession() domain.StreamSession {
	return domain.StreamSession{
		ID: demoSID, ProjectID: demoProject, CorrelationID: demoCID,
		State: domain.SessionStateActive,
	}
}

func TestAnalyzeManifestLink_NoLinkFields_Detached(t *testing.T) {
	t.Parallel()
	repo := newLinkRepo()
	uc := application.NewAnalyzeManifestUseCase(okAnalyzer(), repo)
	got, err := uc.AnalyzeManifest(context.Background(), validAnalyzeRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.SessionLink.Status != domain.SessionLinkStatusDetached {
		t.Errorf("status = %q want detached", got.SessionLink.Status)
	}
	if repo.getCalls != 0 || repo.cidCalls != 0 {
		t.Errorf("no link fields → no repo calls expected; got Get=%d, GetByCID=%d", repo.getCalls, repo.cidCalls)
	}
}

func TestAnalyzeManifestLink_KnownCorrelationID_Linked(t *testing.T) {
	t.Parallel()
	repo := newLinkRepo(demoSession())
	uc := application.NewAnalyzeManifestUseCase(okAnalyzer(), repo)
	req := validAnalyzeRequest()
	req.ProjectID = demoProject
	req.CorrelationID = demoCID
	got, err := uc.AnalyzeManifest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.SessionLink.Status != domain.SessionLinkStatusLinked {
		t.Errorf("status = %q want linked", got.SessionLink.Status)
	}
	if got.SessionLink.SessionID != demoSID || got.SessionLink.CorrelationID != demoCID || got.SessionLink.ProjectID != demoProject {
		t.Errorf("link payload mismatch: %+v", got.SessionLink)
	}
}

func TestAnalyzeManifestLink_UnknownCorrelationID_NotFoundDetached(t *testing.T) {
	t.Parallel()
	repo := newLinkRepo() // empty
	uc := application.NewAnalyzeManifestUseCase(okAnalyzer(), repo)
	req := validAnalyzeRequest()
	req.ProjectID = demoProject
	req.CorrelationID = demoCID
	got, _ := uc.AnalyzeManifest(context.Background(), req)
	if got.SessionLink.Status != domain.SessionLinkStatusNotFoundDetached {
		t.Errorf("status = %q want not_found_detached", got.SessionLink.Status)
	}
}

func TestAnalyzeManifestLink_ForeignProjectCorrelationID_NotFoundDetached(t *testing.T) {
	t.Parallel()
	// CorrelationID existiert in Project "other", aber Token resolved
	// auf "demo" — Cross-Project-Lookup darf nicht greifen.
	other := demoSession()
	other.ProjectID = "other"
	repo := newLinkRepo(other)
	uc := application.NewAnalyzeManifestUseCase(okAnalyzer(), repo)
	req := validAnalyzeRequest()
	req.ProjectID = demoProject
	req.CorrelationID = demoCID
	got, _ := uc.AnalyzeManifest(context.Background(), req)
	if got.SessionLink.Status != domain.SessionLinkStatusNotFoundDetached {
		t.Errorf("status = %q want not_found_detached (cross-project)", got.SessionLink.Status)
	}
}

func TestAnalyzeManifestLink_KnownSessionIDAlone_Linked(t *testing.T) {
	t.Parallel()
	repo := newLinkRepo(demoSession())
	uc := application.NewAnalyzeManifestUseCase(okAnalyzer(), repo)
	req := validAnalyzeRequest()
	req.ProjectID = demoProject
	req.SessionID = demoSID
	got, _ := uc.AnalyzeManifest(context.Background(), req)
	if got.SessionLink.Status != domain.SessionLinkStatusLinked {
		t.Errorf("status = %q want linked", got.SessionLink.Status)
	}
	if got.SessionLink.SessionID != demoSID || got.SessionLink.CorrelationID != demoCID {
		t.Errorf("link payload mismatch: %+v", got.SessionLink)
	}
}

func TestAnalyzeManifestLink_UnknownSessionIDAlone_NotFoundDetached(t *testing.T) {
	t.Parallel()
	repo := newLinkRepo()
	uc := application.NewAnalyzeManifestUseCase(okAnalyzer(), repo)
	req := validAnalyzeRequest()
	req.ProjectID = demoProject
	req.SessionID = demoSID
	got, _ := uc.AnalyzeManifest(context.Background(), req)
	if got.SessionLink.Status != domain.SessionLinkStatusNotFoundDetached {
		t.Errorf("status = %q want not_found_detached", got.SessionLink.Status)
	}
}

func TestAnalyzeManifestLink_BothConsistent_Linked(t *testing.T) {
	t.Parallel()
	repo := newLinkRepo(demoSession())
	uc := application.NewAnalyzeManifestUseCase(okAnalyzer(), repo)
	req := validAnalyzeRequest()
	req.ProjectID = demoProject
	req.CorrelationID = demoCID
	req.SessionID = demoSID
	got, _ := uc.AnalyzeManifest(context.Background(), req)
	if got.SessionLink.Status != domain.SessionLinkStatusLinked {
		t.Errorf("status = %q want linked", got.SessionLink.Status)
	}
}

func TestAnalyzeManifestLink_UnknownCorrelationIDWithKnownSessionID_NotFoundDetached(t *testing.T) {
	t.Parallel()
	// session_id ist im Project bekannt, aber correlation_id ist
	// project-fremd/unbekannt — eine bekannte session_id darf eine
	// unbekannte correlation_id nicht "retten". Spec §3.6.
	repo := newLinkRepo(demoSession())
	uc := application.NewAnalyzeManifestUseCase(okAnalyzer(), repo)
	req := validAnalyzeRequest()
	req.ProjectID = demoProject
	req.CorrelationID = "00000000-0000-0000-0000-000000000000" // unbekannt
	req.SessionID = demoSID                                     // bekannt
	got, _ := uc.AnalyzeManifest(context.Background(), req)
	if got.SessionLink.Status != domain.SessionLinkStatusNotFoundDetached {
		t.Errorf("status = %q want not_found_detached (unknown cid wins over known sid)", got.SessionLink.Status)
	}
}

// TestAnalyzeManifestLink_ForeignProjectCorrelationIDPlusKnownSessionID_NotFoundDetached
// pinnt den Spec §3.6-Halbsatz "eine bekannte session_id darf eine
// unbekannte ODER project-fremde correlation_id nicht retten" für
// den Mix-Fall: cid existiert nur in einem fremden Project, sid
// existiert im eigenen Project.
func TestAnalyzeManifestLink_ForeignProjectCorrelationIDPlusKnownSessionID_NotFoundDetached(t *testing.T) {
	t.Parallel()
	// Eine Session mit derselben cid existiert in Project "other";
	// die session_id existiert in Project "demo". Resolver auf
	// "demo" → cid-Lookup auf "demo" findet nichts, Verhalten muss
	// not_found_detached sein, ohne durch sid-Fallback "gerettet"
	// zu werden.
	other := demoSession()
	other.ProjectID = "other"
	demo := demoSession()
	demo.CorrelationID = "" // demo-sid existiert, aber ohne cid
	repo := newLinkRepo(other, demo)
	uc := application.NewAnalyzeManifestUseCase(okAnalyzer(), repo)
	req := validAnalyzeRequest()
	req.ProjectID = demoProject
	req.CorrelationID = demoCID // existiert nur in "other"
	req.SessionID = demoSID     // existiert in "demo"
	got, _ := uc.AnalyzeManifest(context.Background(), req)
	if got.SessionLink.Status != domain.SessionLinkStatusNotFoundDetached {
		t.Errorf("status = %q want not_found_detached (foreign cid wins over known sid)", got.SessionLink.Status)
	}
}

// TestAnalyzeManifestLink_LegacySessionWithoutCorrelationID pinnt
// den Sub-Fall "sid-Treffer auf Legacy-Session (Pre-§3.2-Closeout,
// CorrelationID leer)" → linked, aber `link.CorrelationID` bleibt
// leer. Der HTTP-Adapter darf darauf vertrauen, dass `omitempty` das
// Wire-Feld weglässt; Konsumenten lesen `status="linked"` plus
// `session_id`/`project_id` ohne `correlation_id`.
func TestAnalyzeManifestLink_LegacySessionWithoutCorrelationID(t *testing.T) {
	t.Parallel()
	legacy := demoSession()
	legacy.CorrelationID = "" // Pre-§3.2 Legacy-Session
	repo := newLinkRepo(legacy)
	uc := application.NewAnalyzeManifestUseCase(okAnalyzer(), repo)
	req := validAnalyzeRequest()
	req.ProjectID = demoProject
	req.SessionID = demoSID // sid-Pfad, kein cid
	got, _ := uc.AnalyzeManifest(context.Background(), req)
	if got.SessionLink.Status != domain.SessionLinkStatusLinked {
		t.Fatalf("status = %q want linked", got.SessionLink.Status)
	}
	if got.SessionLink.SessionID != demoSID {
		t.Errorf("session_id = %q want %q", got.SessionLink.SessionID, demoSID)
	}
	if got.SessionLink.CorrelationID != "" {
		t.Errorf("correlation_id must be empty for legacy session, got %q", got.SessionLink.CorrelationID)
	}
}

func TestAnalyzeManifestLink_BothKnownButMismatch_ConflictDetached(t *testing.T) {
	t.Parallel()
	// correlation_id zeigt auf eine andere Session als die übergebene
	// session_id — Spec §3.6: conflict_detached.
	other := domain.StreamSession{
		ID: "other-sid", ProjectID: demoProject,
		CorrelationID: "ffffffff-9fb9-4c0b-a78f-2f41d8f6e1e7",
		State:         domain.SessionStateActive,
	}
	repo := newLinkRepo(demoSession(), other)
	uc := application.NewAnalyzeManifestUseCase(okAnalyzer(), repo)
	req := validAnalyzeRequest()
	req.ProjectID = demoProject
	req.CorrelationID = other.CorrelationID
	req.SessionID = demoSID // andere Session als die der cid
	got, _ := uc.AnalyzeManifest(context.Background(), req)
	if got.SessionLink.Status != domain.SessionLinkStatusConflictDetached {
		t.Errorf("status = %q want conflict_detached", got.SessionLink.Status)
	}
}

func TestAnalyzeManifestLink_LinkFieldsWithoutProjectContext_NotFoundDetached(t *testing.T) {
	t.Parallel()
	// Adapter-Fehler-Pfad: Link-Felder gesetzt, aber ProjectID leer
	// (z. B. Token nicht aufgelöst). HTTP-Adapter SOLL das schon vor
	// dem Use-Case-Aufruf mit 401 abfangen; Use-Case ist defensiv
	// und liefert not_found_detached statt zu paniken.
	repo := newLinkRepo(demoSession())
	uc := application.NewAnalyzeManifestUseCase(okAnalyzer(), repo)
	req := validAnalyzeRequest()
	req.CorrelationID = demoCID // ProjectID absichtlich leer
	got, _ := uc.AnalyzeManifest(context.Background(), req)
	if got.SessionLink.Status != domain.SessionLinkStatusNotFoundDetached {
		t.Errorf("status = %q want not_found_detached (defensive without project)", got.SessionLink.Status)
	}
	if repo.cidCalls != 0 {
		t.Errorf("repo must not be called without project context; got %d cid calls", repo.cidCalls)
	}
}

func TestAnalyzeManifestLink_AnalyzerErrorBlocksLinkResolution(t *testing.T) {
	t.Parallel()
	// Wenn der Analyzer fehlschlägt, wird kein Link-Resolver-Aufruf
	// gemacht — der Fehler propagiert direkt nach oben.
	stub := &fakeAnalyzer{err: errors.New("analyzer boom")}
	repo := newLinkRepo(demoSession())
	uc := application.NewAnalyzeManifestUseCase(stub, repo)
	req := validAnalyzeRequest()
	req.ProjectID = demoProject
	req.CorrelationID = demoCID
	_, err := uc.AnalyzeManifest(context.Background(), req)
	if err == nil {
		t.Fatalf("expected propagated analyzer error")
	}
	if repo.cidCalls != 0 {
		t.Errorf("link resolution must not run on analyzer error; got %d calls", repo.cidCalls)
	}
}
