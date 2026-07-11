package postgres_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/postgres"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// TestSessionRepository_PgLab deckt den Postgres-session-Adapter gegen
// echte PG ab: UpsertFromEvents-Lifecycle (new/tick/ended), den R-6-
// Race (zwei parallele Upserts auf dieselbe neue Session → genau eine
// Sieger-CorrelationID, beide Aufrufe bekommen sie zurück), List mit
// Keyset-Pagination, Boundaries (inkl. Bulk-IN), SetSampleRate-
// Immutability und Sweep. Gated über MTRACE_PG_LAB_DSN.
func TestSessionRepository_PgLab(t *testing.T) {
	dsn := os.Getenv("MTRACE_PG_LAB_DSN")
	if dsn == "" {
		t.Skip("MTRACE_PG_LAB_DSN nicht gesetzt — PG-Lab-Integrationstest übersprungen")
	}
	ctx := context.Background()
	db, err := storage.OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("OpenPostgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := postgres.NewSessionRepository(db)
	base := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)

	t.Run("UpsertFromEvents: new → tick → ended (idempotent)", func(t *testing.T) {
		const proj, sess = "sess-lab-lc", "s-lc-1"
		// Erstes Event: neue Session, Active, EventCount=1, CID gesetzt.
		cids, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
			sessionEvent(proj, sess, "playback_started", "cid-lc", base),
		})
		if err != nil {
			t.Fatalf("Upsert new: %v", err)
		}
		if cids[sess] != "cid-lc" {
			t.Fatalf("canonical CID = %q, want cid-lc", cids[sess])
		}
		got, err := repo.Get(ctx, proj, sess)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.State != domain.SessionStateActive || got.EventCount != 1 {
			t.Fatalf("nach new: state=%s count=%d, want active/1", got.State, got.EventCount)
		}

		// Tick: bekanntes Event, EventCount hoch, last_seen aktualisiert.
		if _, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
			sessionEvent(proj, sess, "playback_progress", "cid-ignored", base.Add(time.Second)),
		}); err != nil {
			t.Fatalf("Upsert tick: %v", err)
		}
		got, _ = repo.Get(ctx, proj, sess)
		if got.EventCount != 2 || !got.LastEventAt.Equal(base.Add(time.Second)) {
			t.Fatalf("nach tick: count=%d last=%v, want 2 / +1s", got.EventCount, got.LastEventAt)
		}

		// session_ended: State → Ended, EndedAt gesetzt, EndSource=client.
		endAt := base.Add(2 * time.Second)
		if _, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
			sessionEvent(proj, sess, "session_ended", "cid-ignored", endAt),
		}); err != nil {
			t.Fatalf("Upsert ended: %v", err)
		}
		got, _ = repo.Get(ctx, proj, sess)
		if got.State != domain.SessionStateEnded || got.EndedAt == nil || !got.EndedAt.Equal(endAt) {
			t.Fatalf("nach ended: state=%s endedAt=%v, want ended/%v", got.State, got.EndedAt, endAt)
		}
		if got.EndSource != domain.SessionEndSourceClient {
			t.Errorf("EndSource = %q, want client", got.EndSource)
		}

		// Zweites session_ended: EndedAt bleibt (idempotent), Count zählt weiter.
		later := base.Add(time.Hour)
		if _, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
			sessionEvent(proj, sess, "session_ended", "cid-ignored", later),
		}); err != nil {
			t.Fatalf("Upsert ended#2: %v", err)
		}
		got, _ = repo.Get(ctx, proj, sess)
		if !got.EndedAt.Equal(endAt) {
			t.Errorf("EndedAt nach 2. session_ended = %v, want unverändert %v (idempotent)", got.EndedAt, endAt)
		}
		if got.EventCount != 4 {
			t.Errorf("EventCount = %d, want 4 (verspätete Events zählen weiter)", got.EventCount)
		}
	})

	t.Run("R-6: paralleler Upsert auf neue Session → genau eine Sieger-CID", func(t *testing.T) {
		const proj, sess = "sess-lab-race", "s-race-1"
		var wg sync.WaitGroup
		results := make([]string, 2)
		errs := make([]error, 2)
		candidates := []string{"cid-A", "cid-B"}
		wg.Add(2)
		for i := 0; i < 2; i++ {
			go func(idx int) {
				defer wg.Done()
				cids, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
					sessionEvent(proj, sess, "playback_started", candidates[idx], base),
				})
				errs[idx] = err
				if err == nil {
					results[idx] = cids[sess]
				}
			}(i)
		}
		wg.Wait()

		for i, err := range errs {
			if err != nil {
				t.Fatalf("parallel Upsert #%d: %v", i, err)
			}
		}
		// Beide Aufrufe müssen dieselbe kanonische CID zurückbekommen.
		if results[0] != results[1] {
			t.Fatalf("R-6 verletzt: Aufrufe bekamen unterschiedliche CIDs (%q vs %q)", results[0], results[1])
		}
		winner := results[0]
		if winner != "cid-A" && winner != "cid-B" {
			t.Fatalf("Sieger-CID = %q, want eine der Kandidaten cid-A/cid-B", winner)
		}
		// Die DB-Zeile trägt exakt die Sieger-CID (genau eine Session).
		got, err := repo.Get(ctx, proj, sess)
		if err != nil {
			t.Fatalf("Get nach Race: %v", err)
		}
		if got.CorrelationID != winner {
			t.Errorf("DB correlation_id = %q, want Sieger %q", got.CorrelationID, winner)
		}
		// Über GetByCorrelationID auffindbar; die Verlust-CID nicht.
		if _, err := repo.GetByCorrelationID(ctx, proj, winner); err != nil {
			t.Errorf("GetByCorrelationID(winner) = %v, want Treffer", err)
		}
		loser := "cid-A"
		if winner == "cid-A" {
			loser = "cid-B"
		}
		if _, err := repo.GetByCorrelationID(ctx, proj, loser); !errors.Is(err, domain.ErrSessionNotFound) {
			t.Errorf("GetByCorrelationID(loser %q) = %v, want ErrSessionNotFound", loser, err)
		}
	})

	t.Run("Get/GetByCorrelationID: Cross-Project + leere CID", func(t *testing.T) {
		const proj, sess = "sess-lab-scope", "s-scope-1"
		if _, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
			sessionEvent(proj, sess, "playback_started", "cid-scope", base),
		}); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
		if _, err := repo.Get(ctx, "other-proj", sess); !errors.Is(err, domain.ErrSessionNotFound) {
			t.Errorf("Get(cross-project) = %v, want ErrSessionNotFound", err)
		}
		if _, err := repo.GetByCorrelationID(ctx, proj, ""); !errors.Is(err, domain.ErrSessionNotFound) {
			t.Errorf("GetByCorrelationID(empty) = %v, want ErrSessionNotFound", err)
		}
		if _, err := repo.GetByCorrelationID(ctx, "other-proj", "cid-scope"); !errors.Is(err, domain.ErrSessionNotFound) {
			t.Errorf("GetByCorrelationID(cross-project) = %v, want ErrSessionNotFound", err)
		}
	})

	t.Run("List Keyset-Pagination (started_at desc, session_id asc)", func(t *testing.T) {
		const proj = "sess-lab-list"
		// Drei Sessions mit gestaffelten started_at.
		for i, sfx := range []string{"a", "b", "c"} {
			if _, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
				sessionEvent(proj, "s-"+sfx, "playback_started", "cid-"+sfx, base.Add(time.Duration(i)*time.Minute)),
			}); err != nil {
				t.Fatalf("Upsert %s: %v", sfx, err)
			}
		}
		// Page 1: Limit 2 → neueste zwei (c, b), plus NextAfter.
		p1, err := repo.List(ctx, driven.SessionListQuery{ProjectID: proj, Limit: 2})
		if err != nil {
			t.Fatalf("List page1: %v", err)
		}
		if len(p1.Sessions) != 2 || p1.NextAfter == nil {
			t.Fatalf("page1 = %d Sessions, NextAfter=%v; want 2 + Cursor", len(p1.Sessions), p1.NextAfter)
		}
		if p1.Sessions[0].ID != "s-c" || p1.Sessions[1].ID != "s-b" {
			t.Errorf("page1 order = [%s %s], want [s-c s-b]", p1.Sessions[0].ID, p1.Sessions[1].ID)
		}
		// Page 2: Cursor fortsetzen → letzte Session (a), kein NextAfter.
		p2, err := repo.List(ctx, driven.SessionListQuery{ProjectID: proj, Limit: 2, After: p1.NextAfter})
		if err != nil {
			t.Fatalf("List page2: %v", err)
		}
		if len(p2.Sessions) != 1 || p2.Sessions[0].ID != "s-a" || p2.NextAfter != nil {
			t.Errorf("page2 = %+v (NextAfter=%v), want genau [s-a] ohne Cursor", ids(p2.Sessions), p2.NextAfter)
		}
	})

	t.Run("Boundaries: Append idempotent + Einzel- und Bulk-Read", func(t *testing.T) {
		const proj = "sess-lab-bnd"
		for _, sfx := range []string{"x", "y"} {
			if _, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
				sessionEvent(proj, "b-"+sfx, "playback_started", "cid-"+sfx, base),
			}); err != nil {
				t.Fatalf("Upsert %s: %v", sfx, err)
			}
		}
		bnd := mkBoundary(proj, "b-x", "manifest", "hls.js", "gap")
		// Zweimaliges Append derselben Tripel = idempotent (Insert-or-Refresh).
		if err := repo.AppendBoundaries(ctx, []domain.SessionBoundary{bnd, bnd}); err != nil {
			t.Fatalf("AppendBoundaries: %v", err)
		}
		single, err := repo.ListBoundariesForSession(ctx, proj, "b-x")
		if err != nil {
			t.Fatalf("ListBoundariesForSession: %v", err)
		}
		if len(single) != 1 {
			t.Errorf("ListBoundariesForSession = %d, want 1 (Tripel-dedupe)", len(single))
		}
		// Bulk-IN über beide Sessions; b-y hat keine Boundary → fehlt in der Map.
		bulk, err := repo.ListBoundariesForSessions(ctx, proj, []string{"b-x", "b-y"})
		if err != nil {
			t.Fatalf("ListBoundariesForSessions: %v", err)
		}
		if len(bulk["b-x"]) != 1 {
			t.Errorf("bulk[b-x] = %d, want 1", len(bulk["b-x"]))
		}
		if _, ok := bulk["b-y"]; ok {
			t.Errorf("bulk[b-y] vorhanden, want fehlend (keine Boundaries)")
		}
	})

	t.Run("SetSessionSampleRatePPMIfDefault: Immutability", func(t *testing.T) {
		const proj, sess = "sess-lab-rate", "s-rate-1"
		if _, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
			sessionEvent(proj, sess, "playback_started", "cid-rate", base),
		}); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
		// Erstsetzung: applied=true, Rückgabe == Eingabe.
		got, applied, err := repo.SetSessionSampleRatePPMIfDefault(ctx, proj, sess, 250000)
		if err != nil || !applied || got != 250000 {
			t.Fatalf("erste Setzung = (%d, %v, %v), want (250000, true, nil)", got, applied, err)
		}
		// Zweitsetzung: immutable → applied=false, bestehender Wert zurück.
		got, applied, err = repo.SetSessionSampleRatePPMIfDefault(ctx, proj, sess, 500000)
		if err != nil || applied || got != 250000 {
			t.Fatalf("zweite Setzung = (%d, %v, %v), want (250000, false, nil)", got, applied, err)
		}
	})

	t.Run("Sweep: active → stalled → ended nach Schwellen", func(t *testing.T) {
		const proj, sess = "sess-lab-sweep", "s-sweep-1"
		// last_seen liegt weit in der Vergangenheit.
		past := base.Add(-time.Hour)
		if _, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
			sessionEvent(proj, sess, "playback_started", "cid-sweep", past),
		}); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
		// stalledAfter=1m, endedAfter=90m: bei now=base ist last_seen (−60m)
		// älter als 1m (→ stalled), aber nicht älter als 90m (→ nicht ended).
		if err := repo.Sweep(ctx, base, time.Minute, 90*time.Minute); err != nil {
			t.Fatalf("Sweep#1: %v", err)
		}
		got, _ := repo.Get(ctx, proj, sess)
		if got.State != domain.SessionStateStalled {
			t.Fatalf("nach Sweep#1: state=%s, want stalled", got.State)
		}
		// Jetzt endedAfter=30m: −60m ist älter → ended.
		if err := repo.Sweep(ctx, base, time.Minute, 30*time.Minute); err != nil {
			t.Fatalf("Sweep#2: %v", err)
		}
		got, _ = repo.Get(ctx, proj, sess)
		if got.State != domain.SessionStateEnded || got.EndSource != domain.SessionEndSourceSweeper {
			t.Errorf("nach Sweep#2: state=%s source=%s, want ended/sweeper", got.State, got.EndSource)
		}
	})

	t.Run("CountByState: Delta über frische Active-Sessions", func(t *testing.T) {
		const proj = "sess-lab-count"
		before, err := repo.CountByState(ctx, domain.SessionStateActive)
		if err != nil {
			t.Fatalf("CountByState before: %v", err)
		}
		for _, sfx := range []string{"c1", "c2"} {
			if _, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
				sessionEvent(proj, sfx, "playback_started", "cid-"+sfx, base),
			}); err != nil {
				t.Fatalf("Upsert %s: %v", sfx, err)
			}
		}
		after, err := repo.CountByState(ctx, domain.SessionStateActive)
		if err != nil {
			t.Fatalf("CountByState after: %v", err)
		}
		if after-before != 2 {
			t.Errorf("Active-Delta = %d, want 2", after-before)
		}
	})
}

func sessionEvent(proj, sess, name, cid string, serverAt time.Time) domain.PlaybackEvent {
	return domain.PlaybackEvent{
		EventName:        name,
		ProjectID:        proj,
		SessionID:        sess,
		ClientTimestamp:  serverAt,
		ServerReceivedAt: serverAt,
		CorrelationID:    cid,
	}
}

func mkBoundary(proj, sess, networkKind, adapter, reason string) domain.SessionBoundary {
	return domain.SessionBoundary{
		Kind:             "network_signal_absent",
		ProjectID:        proj,
		SessionID:        sess,
		NetworkKind:      networkKind,
		Adapter:          adapter,
		Reason:           reason,
		ClientTimestamp:  time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC),
		ServerReceivedAt: time.Date(2026, 7, 11, 12, 0, 1, 0, time.UTC),
	}
}

func ids(sessions []domain.StreamSession) []string {
	out := make([]string, len(sessions))
	for i, s := range sessions {
		out[i] = s.ID
	}
	return out
}
