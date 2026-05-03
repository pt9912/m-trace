package application_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// stubProjectResolver returns a single demo project for "demo-token",
// nothing else (matches the contract's hardcoded map). allowedOrigins
// trägt die CORS-Variante-B-Allowlist für origin-bezogene Tests; leer
// lässt IsOriginAllowed für jeden non-empty Origin false werden,
// während Origin="" (CLI/curl) immer durchgewinkt wird.
type stubProjectResolver struct {
	allowedOrigins []string
}

func (s stubProjectResolver) ResolveByToken(_ context.Context, token string) (domain.Project, error) {
	if token == "demo-token" {
		return domain.Project{
			ID:             "demo",
			Token:          domain.ProjectToken("demo-token"),
			AllowedOrigins: append([]string(nil), s.allowedOrigins...),
		}, nil
	}
	return domain.Project{}, domain.ErrUnauthorized
}

type stubLimiter struct {
	deny bool
}

func (s *stubLimiter) Allow(_ context.Context, _ driven.RateLimitKey, _ int) error {
	if s.deny {
		return domain.ErrRateLimited
	}
	return nil
}

type stubRepo struct {
	appended []domain.PlaybackEvent
	failNext bool
}

func (s *stubRepo) Append(_ context.Context, events []domain.PlaybackEvent) error {
	if s.failNext {
		s.failNext = false
		return errors.New("repo failure")
	}
	s.appended = append(s.appended, events...)
	return nil
}

// ListBySession ist no-op — der Use Case in dieser Test-Suite ruft den
// Read-Pfad nicht. Stub-Methode existiert nur, damit *stubRepo den
// driven.EventRepository-Vertrag erfüllt.
func (s *stubRepo) ListBySession(_ context.Context, _ driven.EventListQuery) (driven.EventPage, error) {
	return driven.EventPage{}, nil
}

// stubSessionRepo zeichnet UpsertFromEvents-Aufrufe auf und kann eine
// einmalige Failure simulieren. List/Get/Sweep sind no-ops — der Use
// Case in dieser Test-Suite ruft sie nicht.
type stubSessionRepo struct {
	upserts  [][]domain.PlaybackEvent
	failNext bool
	// existing erlaubt Tests, ein Repository-Get-Resultat pro
	// session_id vorzubereiten — z. B. um den
	// "existing-CorrelationID-übernehmen"-Pfad in
	// resolveCorrelationIDs abzudecken.
	existing map[string]domain.StreamSession
	// getError lässt Tests einen DB-Fehler-Pfad simulieren (Get
	// returnt einen anderen Fehler als domain.ErrSessionNotFound).
	getError error
}

func (s *stubSessionRepo) UpsertFromEvents(_ context.Context, events []domain.PlaybackEvent) (map[string]string, error) {
	if s.failNext {
		s.failNext = false
		return nil, errors.New("session repo failure")
	}
	dup := make([]domain.PlaybackEvent, len(events))
	copy(dup, events)
	s.upserts = append(s.upserts, dup)
	canonical := make(map[string]string, len(events))
	for _, e := range events {
		// Stub liefert die Eingabe-CorrelationID als „canonical" zurück
		// (kein Race-Mischen). Tests, die R-6-spezifisches Race-Verhalten
		// brauchen, gehen direkt gegen den SQLite-Adapter.
		canonical[e.SessionID] = e.CorrelationID
	}
	return canonical, nil
}

func (s *stubSessionRepo) List(_ context.Context, _ driven.SessionListQuery) (driven.SessionPage, error) {
	return driven.SessionPage{}, nil
}

func (s *stubSessionRepo) Get(_ context.Context, _ string, sessionID string) (domain.StreamSession, error) {
	if s.getError != nil {
		return domain.StreamSession{}, s.getError
	}
	if sess, ok := s.existing[sessionID]; ok {
		return sess, nil
	}
	return domain.StreamSession{}, domain.ErrSessionNotFound
}

func (s *stubSessionRepo) GetByCorrelationID(_ context.Context, _ string, _ string) (domain.StreamSession, error) {
	return domain.StreamSession{}, domain.ErrSessionNotFound
}

func (s *stubSessionRepo) Sweep(_ context.Context, _ time.Time, _, _ time.Duration) error {
	return nil
}

func (s *stubSessionRepo) CountByState(_ context.Context, _ domain.SessionState) (int64, error) {
	return 0, nil
}

type spyMetrics struct {
	accepted, invalid, rateLimited, dropped int
	playbackErrors, rebufferEvents          int
	startupTimes                            []float64
}

func (s *spyMetrics) EventsAccepted(n int)    { s.accepted += n }
func (s *spyMetrics) InvalidEvents(n int)     { s.invalid += n }
func (s *spyMetrics) RateLimitedEvents(n int) { s.rateLimited += n }
func (s *spyMetrics) DroppedEvents(n int)     { s.dropped += n }
func (s *spyMetrics) PlaybackErrors(n int)    { s.playbackErrors += n }
func (s *spyMetrics) RebufferEvents(n int)    { s.rebufferEvents += n }
func (s *spyMetrics) StartupTimeMS(ms float64) {
	s.startupTimes = append(s.startupTimes, ms)
}

// stubTelemetry zählt BatchReceived-Aufrufe. Pro Aufruf wird die
// gemeldete Batch-Größe addiert; calls misst die reine Aufrufzahl.
type stubTelemetry struct {
	calls     int
	totalSize int
	lastSize  int
}

func (s *stubTelemetry) BatchReceived(_ context.Context, size int) {
	s.calls++
	s.totalSize += size
	s.lastSize = size
}

// stubAnalyzer zählt AnalyzeBatch-Aufrufe. Im 0.1.0-Use-Case ruft
// das System ihn nicht produktiv auf — der Slot existiert ausschließlich
// als F-22-Architektur-Vorbereitung (siehe plan-0.1.0.md §5.1 F-22).
// calls bleibt damit in allen Tests 0; das ist die DoD-Bedingung.
//
// AnalyzeManifest ist ab 0.3.0 Teil des Ports (plan-0.3.0 §2 Tranche 1)
// und vom Batch-Use-Case ebenfalls nicht aufgerufen; der Stub
// implementiert die Methode no-op, um den Port-Vertrag zu erfüllen.
type stubAnalyzer struct {
	calls int
}

func (s *stubAnalyzer) AnalyzeBatch(_ context.Context, _ []domain.PlaybackEvent) error {
	s.calls++
	return nil
}

func (*stubAnalyzer) AnalyzeManifest(_ context.Context, _ domain.StreamAnalysisRequest) (domain.StreamAnalysisResult, error) {
	return domain.StreamAnalysisResult{}, nil
}

// stubSequencer liefert deterministische, monoton steigende
// ingest_sequence-Werte ab 1 für die Use-Case-Tests.
type stubSequencer struct {
	last int64
}

func (s *stubSequencer) Next() int64 {
	s.last++
	return s.last
}

func validBatch() driving.BatchInput {
	return driving.BatchInput{
		SchemaVersion: application.SupportedSchemaVersion,
		AuthToken:     "demo-token",
		Events: []driving.EventInput{
			{
				EventName:       "rebuffer_started",
				ProjectID:       "demo",
				SessionID:       "01J7K9X4Z2QHB6V3WS5R8Y4D1F",
				ClientTimestamp: "2026-04-28T12:00:00.000Z",
				SDK:             driving.SDKInput{Name: "@npm9912/player-sdk", Version: "0.2.0"},
			},
		},
	}
}

func newUseCase() (*application.RegisterPlaybackEventBatchUseCase, *stubLimiter, *stubRepo, *stubSessionRepo, *spyMetrics, *stubTelemetry, *stubAnalyzer, *stubSequencer) {
	return newUseCaseWithOrigins(nil)
}

// newUseCaseWithOrigins erlaubt die Origin-Allowlist des Stub-Project-
// Resolvers zu konfigurieren — Voraussetzung für die CORS-Variante-B-
// Tests aus plan-0.1.0.md §5.1 Sub-Item 6.
func newUseCaseWithOrigins(origins []string) (*application.RegisterPlaybackEventBatchUseCase, *stubLimiter, *stubRepo, *stubSessionRepo, *spyMetrics, *stubTelemetry, *stubAnalyzer, *stubSequencer) {
	limiter := &stubLimiter{}
	repo := &stubRepo{}
	sessions := &stubSessionRepo{}
	metrics := &spyMetrics{}
	telemetry := &stubTelemetry{}
	analyzer := &stubAnalyzer{}
	sequencer := &stubSequencer{}
	uc := application.NewRegisterPlaybackEventBatchUseCase(
		stubProjectResolver{allowedOrigins: origins}, limiter, repo, sessions, metrics, telemetry, analyzer, sequencer,
		func() time.Time { return time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC) },
	)
	return uc, limiter, repo, sessions, metrics, telemetry, analyzer, sequencer
}

func TestHappyPath(t *testing.T) {
	t.Parallel()
	uc, _, repo, sessions, metrics, telemetry, analyzer, _ := newUseCase()
	res, err := uc.RegisterPlaybackEventBatch(context.Background(), validBatch())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Accepted != 1 {
		t.Errorf("expected 1 accepted, got %d", res.Accepted)
	}
	if len(repo.appended) != 1 {
		t.Errorf("expected 1 appended event, got %d", len(repo.appended))
	}
	if metrics.accepted != 1 {
		t.Errorf("expected EventsAccepted=1, got %d", metrics.accepted)
	}
	if telemetry.calls != 1 {
		t.Errorf("expected Telemetry.BatchReceived calls=1, got %d", telemetry.calls)
	}
	if telemetry.lastSize != 1 {
		t.Errorf("expected Telemetry.lastSize=1, got %d", telemetry.lastSize)
	}
	// F-22-Architektur-Vorbereitung: in 0.1.0 darf der Use Case den
	// StreamAnalyzer noch nicht produktiv aufrufen (Slot-only).
	if analyzer.calls != 0 {
		t.Errorf("expected StreamAnalyzer.AnalyzeBatch=0 in 0.1.0 (slot-only), got %d", analyzer.calls)
	}
	// IngestSequence: erstes Event muss den ersten Sequencer-Wert (1) tragen.
	if got := repo.appended[0].IngestSequence; got != 1 {
		t.Errorf("expected IngestSequence=1 for first event, got %d", got)
	}
	// SessionRepository: ein Upsert pro Batch.
	if got := len(sessions.upserts); got != 1 {
		t.Errorf("expected 1 SessionRepository.UpsertFromEvents call, got %d", got)
	}
	if got := len(sessions.upserts[0]); got != 1 {
		t.Errorf("expected upsert with 1 event, got %d", got)
	}
}

// TestIngestSequenceMonotonic verifiziert, dass mehrere Events einer
// Batch jeweils einen monoton steigenden ingest_sequence-Wert tragen.
// plan-0.1.0.md §5.1: "monoton steigender Counter pro apps/api-Prozess".
func TestIngestSequenceMonotonic(t *testing.T) {
	t.Parallel()
	uc, _, repo, _, _, _, _, _ := newUseCase()
	in := validBatch()
	template := in.Events[0]
	in.Events = []driving.EventInput{template, template, template}
	if _, err := uc.RegisterPlaybackEventBatch(context.Background(), in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.appended) != 3 {
		t.Fatalf("expected 3 appended events, got %d", len(repo.appended))
	}
	for i, ev := range repo.appended {
		if got, want := ev.IngestSequence, int64(i+1); got != want {
			t.Errorf("event[%d].IngestSequence=%d want %d", i, got, want)
		}
	}
}

func TestPlaybackAggregateMetrics(t *testing.T) {
	t.Parallel()
	uc, _, repo, _, metrics, _, _, _ := newUseCase()
	in := validBatch()
	template := in.Events[0]
	playbackError := template
	playbackError.EventName = "playback_error"
	startup := template
	startup.EventName = "startup_time_measured"
	startup.Meta = map[string]any{"duration_ms": float64(1234)}
	in.Events = []driving.EventInput{template, playbackError, startup}

	if _, err := uc.RegisterPlaybackEventBatch(context.Background(), in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := len(repo.appended); got != 3 {
		t.Fatalf("expected 3 appended events, got %d", got)
	}
	if metrics.rebufferEvents != 1 {
		t.Errorf("expected RebufferEvents=1, got %d", metrics.rebufferEvents)
	}
	if metrics.playbackErrors != 1 {
		t.Errorf("expected PlaybackErrors=1, got %d", metrics.playbackErrors)
	}
	if len(metrics.startupTimes) != 1 || metrics.startupTimes[0] != 1234 {
		t.Errorf("expected StartupTimeMS=[1234], got %#v", metrics.startupTimes)
	}
	if got := repo.appended[2].Meta["duration_ms"]; got != float64(1234) {
		t.Errorf("expected persisted meta duration_ms=1234, got %#v", got)
	}
}

// TestTelemetryReceivedBeforeAuth verifiziert, dass BatchReceived auch
// bei fehlgeschlagener Auth gerufen wird (Counter misst received,
// nicht validated — siehe Telemetry-Port-Doc).
func TestTelemetryReceivedBeforeAuth(t *testing.T) {
	t.Parallel()
	uc, _, _, _, _, telemetry, _, _ := newUseCase()
	in := validBatch()
	in.AuthToken = "wrong-token"
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
	if telemetry.calls != 1 {
		t.Errorf("expected Telemetry.BatchReceived calls=1 (received zählt vor Auth), got %d", telemetry.calls)
	}
}

func TestUnauthorizedToken(t *testing.T) {
	t.Parallel()
	uc, _, _, _, _, _, _, _ := newUseCase()
	in := validBatch()
	in.AuthToken = "wrong-token"
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestSchemaVersionMismatch(t *testing.T) {
	t.Parallel()
	uc, _, _, _, metrics, _, _, _ := newUseCase()
	in := validBatch()
	in.SchemaVersion = "2.0"
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrSchemaVersionMismatch) {
		t.Errorf("expected ErrSchemaVersionMismatch, got %v", err)
	}
	if metrics.invalid != 1 {
		t.Errorf("expected InvalidEvents=1, got %d", metrics.invalid)
	}
}

func TestEmptyBatch(t *testing.T) {
	t.Parallel()
	uc, _, _, _, metrics, _, _, _ := newUseCase()
	in := validBatch()
	in.Events = nil
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrBatchEmpty) {
		t.Errorf("expected ErrBatchEmpty, got %v", err)
	}
	// Counter zählt Events, nicht Batches — bei n=0 kein Increment
	// (Lastenheft 1.1.2 §7.9).
	if metrics.invalid != 0 {
		t.Errorf("expected InvalidEvents=0 (empty batch counts no events), got %d", metrics.invalid)
	}
}

func TestBatchTooLarge(t *testing.T) {
	t.Parallel()
	uc, _, _, _, metrics, _, _, _ := newUseCase()
	in := validBatch()
	template := in.Events[0]
	in.Events = make([]driving.EventInput, application.MaxBatchSize+1)
	for i := range in.Events {
		in.Events[i] = template
	}
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrBatchTooLarge) {
		t.Errorf("expected ErrBatchTooLarge, got %v", err)
	}
	if metrics.invalid != application.MaxBatchSize+1 {
		t.Errorf("expected InvalidEvents=%d, got %d", application.MaxBatchSize+1, metrics.invalid)
	}
}

func TestInvalidEventMissingField(t *testing.T) {
	t.Parallel()
	uc, _, _, _, _, _, _, _ := newUseCase()
	in := validBatch()
	in.Events[0].EventName = ""
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrInvalidEvent) {
		t.Errorf("expected ErrInvalidEvent, got %v", err)
	}
}

func TestInvalidEventBadTimestamp(t *testing.T) {
	t.Parallel()
	uc, _, _, _, _, _, _, _ := newUseCase()
	in := validBatch()
	in.Events[0].ClientTimestamp = "not-a-timestamp"
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrInvalidEvent) {
		t.Errorf("expected ErrInvalidEvent, got %v", err)
	}
}

func TestProjectIDTokenMismatch(t *testing.T) {
	t.Parallel()
	uc, _, _, _, metrics, _, _, _ := newUseCase()
	in := validBatch()
	in.Events[0].ProjectID = "other-project"
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
	// Auth-Fehler (401) zählen nicht in invalid_events (API-Kontrakt §7).
	if metrics.invalid != 0 {
		t.Errorf("expected InvalidEvents=0 (auth-Fehler zählen nicht), got %d", metrics.invalid)
	}
}

// TestOriginNotAllowed_NoSideEffects verifiziert plan-0.1.0.md §5.1
// CORS Variante B: ein Origin, der nicht in der Allowlist des Project
// steht, gibt ErrOriginNotAllowed zurück, ohne den Rate-Limiter zu
// belasten oder Events anzulegen — die Origin-Validierung läuft vor
// Step 4.
func TestOriginNotAllowed_NoSideEffects(t *testing.T) {
	t.Parallel()
	uc, limiter, repo, sessions, metrics, _, _, _ := newUseCaseWithOrigins([]string{"http://allowed.example"})

	in := validBatch()
	in.Origin = "http://attacker.example"

	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrOriginNotAllowed) {
		t.Fatalf("expected ErrOriginNotAllowed, got %v", err)
	}
	if limiter.deny {
		t.Errorf("limiter wasn't reached but state changed unexpectedly")
	}
	if len(repo.appended) != 0 {
		t.Errorf("expected 0 appended, got %d (origin reject must not persist)", len(repo.appended))
	}
	if len(sessions.upserts) != 0 {
		t.Errorf("expected 0 session upserts, got %d", len(sessions.upserts))
	}
	if metrics.invalid != 0 || metrics.rateLimited != 0 || metrics.accepted != 0 || metrics.dropped != 0 {
		t.Errorf("origin reject must not touch metrics (got accepted=%d invalid=%d rl=%d dropped=%d)",
			metrics.accepted, metrics.invalid, metrics.rateLimited, metrics.dropped)
	}
}

// TestOriginEmpty_BypassesProjectBinding verifiziert den CLI/curl-Pfad
// (plan-0.1.0.md §5.1): kein Origin-Header → keine Project-Bindung,
// kein 403.
func TestOriginEmpty_BypassesProjectBinding(t *testing.T) {
	t.Parallel()
	uc, _, repo, _, _, _, _, _ := newUseCaseWithOrigins([]string{"http://allowed.example"})
	in := validBatch()
	in.Origin = ""
	if _, err := uc.RegisterPlaybackEventBatch(context.Background(), in); err != nil {
		t.Fatalf("expected accepted, got %v", err)
	}
	if len(repo.appended) != 1 {
		t.Errorf("expected 1 appended, got %d", len(repo.appended))
	}
}

func TestRateLimited(t *testing.T) {
	t.Parallel()
	uc, limiter, _, _, metrics, _, _, _ := newUseCase()
	limiter.deny = true
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), validBatch())
	if !errors.Is(err, domain.ErrRateLimited) {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
	if metrics.rateLimited != 1 {
		t.Errorf("expected RateLimitedEvents=1, got %d", metrics.rateLimited)
	}
}

func TestRepoFailureDoesNotCountAsDropped(t *testing.T) {
	t.Parallel()
	uc, _, repo, sessions, metrics, _, _, _ := newUseCase()
	repo.failNext = true
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), validBatch())
	if err == nil {
		t.Fatal("expected an error")
	}
	// Reihenfolge ab plan-0.4.0 §4.2 C2: Session-Upsert läuft VOR dem
	// Event-Append, damit der Use-Case die DB-finale CorrelationID
	// kennt, bevor Events persistiert werden (R-6-Fix). Bei einem
	// Append-Fehler ist die Session-Zeile damit bereits angelegt; das
	// ist die dokumentierte schwächere Divergenz im Vergleich zur
	// R-6-Inkonsistenz, die sie ersetzt. Test pinnt deshalb den
	// **einen** UpsertFromEvents-Call.
	if got := len(sessions.upserts); got != 1 {
		t.Errorf("expected 1 SessionRepository.UpsertFromEvents call (sessions persisted before append, see C2 reorder), got %d", got)
	}
	// Synchron fehlgeschlagenes Append ist kein Backpressure-Drop;
	// dropped_events bleibt unverändert (API-Kontrakt §7,
	// Lastenheft 1.1.2 §7.9 nach Plan §4.2).
	if metrics.dropped != 0 {
		t.Errorf("expected DroppedEvents=0 (synchron fehlgeschlagenes Append ist kein Backpressure-Drop), got %d", metrics.dropped)
	}
	if metrics.accepted != 0 {
		t.Errorf("expected EventsAccepted=0 on repo failure, got %d", metrics.accepted)
	}
}

// --- Trace-Korrelation und Time-Skew (plan-0.4.0 §3.2) ---------------

const uuidPattern = `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`

// TestRegisterPlaybackEventBatch_NewSession_GeneratesCorrelationID
// verifiziert: Session noch nicht im Repository → Use-Case generiert
// eine neue UUIDv4 als CorrelationID, schreibt sie auf jedes Event und
// reicht sie an UpsertFromEvents weiter.
func TestRegisterPlaybackEventBatch_NewSession_GeneratesCorrelationID(t *testing.T) {
	t.Parallel()
	uc, _, repo, sessions, _, _, _, _ := newUseCase()

	res, err := uc.RegisterPlaybackEventBatch(context.Background(), validBatch())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if res.SessionCount != 1 {
		t.Errorf("SessionCount = %d, want 1", res.SessionCount)
	}
	if res.SessionCorrelationID == "" {
		t.Fatalf("SessionCorrelationID empty (Single-Session-Batch erwartet)")
	}
	if matched, _ := regexp.MatchString(uuidPattern, res.SessionCorrelationID); !matched {
		t.Errorf("SessionCorrelationID %q is not UUIDv4-shape", res.SessionCorrelationID)
	}
	if len(repo.appended) != 1 || repo.appended[0].CorrelationID != res.SessionCorrelationID {
		t.Errorf("event CorrelationID does not match BatchResult: events=%+v result=%q",
			repo.appended, res.SessionCorrelationID)
	}
	if len(sessions.upserts) != 1 || sessions.upserts[0][0].CorrelationID != res.SessionCorrelationID {
		t.Errorf("session UpsertFromEvents did not see the assigned CorrelationID")
	}
}

// TestRegisterPlaybackEventBatch_ExistingSession_ReusesCorrelationID
// verifiziert: Session bereits bekannt mit CorrelationID → Use-Case
// übernimmt sie, generiert keine neue.
func TestRegisterPlaybackEventBatch_ExistingSession_ReusesCorrelationID(t *testing.T) {
	t.Parallel()
	uc, _, repo, sessions, _, _, _, _ := newUseCase()
	const existingCorr = "11111111-2222-4333-8444-555555555555"
	sessions.existing = map[string]domain.StreamSession{
		"01J7K9X4Z2QHB6V3WS5R8Y4D1F": {
			ID:            "01J7K9X4Z2QHB6V3WS5R8Y4D1F",
			ProjectID:     "demo",
			State:         domain.SessionStateActive,
			CorrelationID: existingCorr,
		},
	}

	res, err := uc.RegisterPlaybackEventBatch(context.Background(), validBatch())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if res.SessionCorrelationID != existingCorr {
		t.Errorf("SessionCorrelationID = %q, want existing %q", res.SessionCorrelationID, existingCorr)
	}
	if repo.appended[0].CorrelationID != existingCorr {
		t.Errorf("event CorrelationID = %q, want %q (must reuse existing session value)",
			repo.appended[0].CorrelationID, existingCorr)
	}
}

// TestRegisterPlaybackEventBatch_MultiSession verifiziert: zwei
// distincte session_ids im Batch → SessionCount=2,
// SessionCorrelationID="" (Span-Attribut wird nicht gesetzt).
func TestRegisterPlaybackEventBatch_MultiSession(t *testing.T) {
	t.Parallel()
	uc, _, _, _, _, _, _, _ := newUseCase()

	in := validBatch()
	second := in.Events[0]
	second.SessionID = "second-session"
	in.Events = append(in.Events, second)

	res, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if res.SessionCount != 2 {
		t.Errorf("SessionCount = %d, want 2", res.SessionCount)
	}
	if res.SessionCorrelationID != "" {
		t.Errorf("SessionCorrelationID = %q, want empty (multi-session batch)",
			res.SessionCorrelationID)
	}
}

// TestRegisterPlaybackEventBatch_TimeSkew verifiziert die drei
// Schwellwert-Fälle aus telemetry-model §5.3 / §3.1: Skew exakt 60 s
// → kein Warning (strict greater); 60 s + 1 ns → Warning; 120 s →
// Warning. Server now() im Stub ist 2026-04-28T12:00:00.000Z.
func TestRegisterPlaybackEventBatch_TimeSkew(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		clientStamp  string
		wantWarning  bool
	}{
		{"exactly 60s skew → no warning", "2026-04-28T11:59:00.000Z", false},
		{"60s + 1ns → warning", "2026-04-28T11:58:59.999999999Z", true},
		{"2min skew → warning", "2026-04-28T11:58:00.000Z", true},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			uc, _, _, _, _, _, _, _ := newUseCase()
			in := validBatch()
			in.Events[0].ClientTimestamp = c.clientStamp

			res, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
			if err != nil {
				t.Fatalf("unexpected: %v", err)
			}
			if res.TimeSkewWarning != c.wantWarning {
				t.Errorf("TimeSkewWarning = %v, want %v (boundary case)",
					res.TimeSkewWarning, c.wantWarning)
			}
		})
	}
}

// TestRegisterPlaybackEventBatch_LegacySessionWithoutCorrelationID
// verifiziert das Self-Healing aus resolveCorrelationIDs: existing
// Session aus Vor-§3.2-Daten hat CorrelationID="" — Use-Case
// generiert eine neue UUIDv4 und schreibt sie aufs Event.
func TestRegisterPlaybackEventBatch_LegacySessionWithoutCorrelationID(t *testing.T) {
	t.Parallel()
	uc, _, repo, sessions, _, _, _, _ := newUseCase()
	sessions.existing = map[string]domain.StreamSession{
		"01J7K9X4Z2QHB6V3WS5R8Y4D1F": {
			ID:            "01J7K9X4Z2QHB6V3WS5R8Y4D1F",
			ProjectID:     "demo",
			State:         domain.SessionStateActive,
			CorrelationID: "", // Legacy-Daten von vor §3.2-Closeout
		},
	}

	res, err := uc.RegisterPlaybackEventBatch(context.Background(), validBatch())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if res.SessionCorrelationID == "" {
		t.Fatal("expected self-healed CorrelationID, got empty")
	}
	if matched, _ := regexp.MatchString(uuidPattern, res.SessionCorrelationID); !matched {
		t.Errorf("self-healed CorrelationID %q is not UUIDv4-shape", res.SessionCorrelationID)
	}
	if repo.appended[0].CorrelationID != res.SessionCorrelationID {
		t.Errorf("event CorrelationID does not match self-healed value")
	}
}

// TestRegisterPlaybackEventBatch_SessionRepoGetError verifiziert,
// dass ein nicht-ErrSessionNotFound-Fehler aus Get propagiert wird —
// wir wollen keine stillschweigend weiterlaufende Session-Anlage,
// wenn das Repository einen DB-Fehler signalisiert.
func TestRegisterPlaybackEventBatch_SessionRepoGetError(t *testing.T) {
	t.Parallel()
	uc, _, _, sessions, _, _, _, _ := newUseCase()
	sessions.getError = errors.New("simulated repo failure")

	_, err := uc.RegisterPlaybackEventBatch(context.Background(), validBatch())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sessions.getError) && err.Error() != sessions.getError.Error() {
		t.Errorf("expected repo failure to propagate, got %v", err)
	}
}

// TestRegisterPlaybackEventBatch_TraceContextPropagated verifiziert,
// dass BatchInput.Trace.TraceID und .SpanID auf jedem persistierten
// Event landen — Voraussetzung für Tempo-Korrelation und für die
// Read-Antwort aus API-Kontrakt §3.7.
func TestRegisterPlaybackEventBatch_TraceContextPropagated(t *testing.T) {
	t.Parallel()
	uc, _, repo, _, _, _, _, _ := newUseCase()

	in := validBatch()
	in.Trace = driving.BatchTraceContext{
		TraceID: "0af7651916cd43dd8448eb211c80319c",
		SpanID:  "b7ad6b7169203331",
	}

	if _, err := uc.RegisterPlaybackEventBatch(context.Background(), in); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got := repo.appended[0].TraceID; got != in.Trace.TraceID {
		t.Errorf("event.TraceID = %q, want %q", got, in.Trace.TraceID)
	}
	if got := repo.appended[0].SpanID; got != in.Trace.SpanID {
		t.Errorf("event.SpanID = %q, want %q", got, in.Trace.SpanID)
	}
}
