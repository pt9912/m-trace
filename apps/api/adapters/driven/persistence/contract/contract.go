// Package contract bündelt die kanonische Test-Suite für alle
// driven-Persistence-Adapter (siehe `plan-0.4.0.md` §2.3 DoD).
// Factories aus `inmemory_test` und `sqlite_test` rufen `RunAll` auf
// und garantieren damit identisches Verhalten beider Implementierungen
// gegen denselben Test-Korpus. Das Paket importiert `testing` analog
// zu Stdlib-Helfern wie `testing/fstest`.
package contract

import (
	"context"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// Repos bündelt die drei Adapter eines Backends. Jede Test-Funktion
// erwartet ein frisches, leeres Tripel.
type Repos struct {
	Sessions  driven.SessionRepository
	Events    driven.EventRepository
	Sequencer driven.IngestSequencer
}

// Factory erzeugt für jeden Test-Run ein frisches Repos-Tripel.
// Test-Setups dürfen `t.Cleanup` nutzen, um Backend-spezifische
// Ressourcen (z. B. SQLite-Dateien) zu entsorgen.
type Factory func(t *testing.T) Repos

// RunAll führt die kanonische Test-Suite gegen die durch Factory
// bereitgestellte Implementierung aus. Aufrufer wählen über `t.Run`
// einen sprechenden Sub-Test-Namen (z. B. "inmemory" / "sqlite").
func RunAll(t *testing.T, factory Factory) {
	t.Helper()
	cases := []struct {
		name string
		run  func(*testing.T, Factory)
	}{
		{"event ordering", testEventOrdering},
		{"event cursor pagination", testEventCursorPagination},
		{"session upsert from first event", testSessionUpsertFirstEvent},
		{"session tick increments event_count", testSessionTickIncrements},
		{"session_ended is idempotent", testSessionEndedIdempotent},
		{"sweep transitions lifecycle", testSweepTransitions},
		{"sequencer is monotone and starts at one", testSequencerMonotonic},
		{"session list with cursor pagination", testSessionListPagination},
		{"event meta round trips", testEventMetaRoundTrip},
		{"session_ended as first event creates ended session", testSessionEndedAsFirstEvent},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			c.run(t, factory)
		})
	}
}

// --- Test cases --------------------------------------------------------

func testEventOrdering(t *testing.T, factory Factory) {
	ctx := context.Background()
	r := factory(t)
	t0 := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	events := []domain.PlaybackEvent{
		mkEvent(r.Sequencer, "demo", "s1", t0.Add(2*time.Second), seq(2)),
		mkEvent(r.Sequencer, "demo", "s1", t0.Add(1*time.Second), seq(1)),
		mkEvent(r.Sequencer, "demo", "s1", t0.Add(3*time.Second), nil),
	}
	if err := r.Sessions.UpsertFromEvents(ctx, events); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := r.Events.Append(ctx, events); err != nil {
		t.Fatalf("append: %v", err)
	}

	page, err := r.Events.ListBySession(ctx, driven.EventListQuery{
		SessionID: "s1", Limit: 10,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Events) != 3 {
		t.Fatalf("len(events) = %d, want 3", len(page.Events))
	}
	gotRcv := []time.Time{page.Events[0].ServerReceivedAt, page.Events[1].ServerReceivedAt, page.Events[2].ServerReceivedAt}
	wantRcv := []time.Time{t0.Add(1 * time.Second), t0.Add(2 * time.Second), t0.Add(3 * time.Second)}
	for i := range gotRcv {
		if !gotRcv[i].Equal(wantRcv[i]) {
			t.Errorf("event[%d] rcv = %v, want %v", i, gotRcv[i], wantRcv[i])
		}
	}
}

func testEventCursorPagination(t *testing.T, factory Factory) {
	ctx := context.Background()
	r := factory(t)
	t0 := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	events := make([]domain.PlaybackEvent, 0, 5)
	for i := 0; i < 5; i++ {
		s := int64(i + 1)
		events = append(events, mkEvent(r.Sequencer, "demo", "s1",
			t0.Add(time.Duration(i)*time.Second), &s))
	}
	if err := r.Sessions.UpsertFromEvents(ctx, events); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := r.Events.Append(ctx, events); err != nil {
		t.Fatalf("append: %v", err)
	}

	page1, err := r.Events.ListBySession(ctx, driven.EventListQuery{
		SessionID: "s1", Limit: 2,
	})
	if err != nil {
		t.Fatalf("list page1: %v", err)
	}
	if len(page1.Events) != 2 || page1.NextAfter == nil {
		t.Fatalf("page1 = %+v, want len=2 and NextAfter set", page1)
	}

	page2, err := r.Events.ListBySession(ctx, driven.EventListQuery{
		SessionID: "s1", Limit: 2, After: page1.NextAfter,
	})
	if err != nil {
		t.Fatalf("list page2: %v", err)
	}
	if len(page2.Events) != 2 || page2.NextAfter == nil {
		t.Fatalf("page2 = %+v, want len=2 and NextAfter set", page2)
	}

	page3, err := r.Events.ListBySession(ctx, driven.EventListQuery{
		SessionID: "s1", Limit: 2, After: page2.NextAfter,
	})
	if err != nil {
		t.Fatalf("list page3: %v", err)
	}
	if len(page3.Events) != 1 || page3.NextAfter != nil {
		t.Fatalf("page3 = %+v, want len=1 and no NextAfter", page3)
	}

	all := append(append(append([]domain.PlaybackEvent{}, page1.Events...), page2.Events...), page3.Events...)
	if len(all) != 5 {
		t.Fatalf("total = %d, want 5", len(all))
	}
	for i := 1; i < len(all); i++ {
		if !all[i].ServerReceivedAt.After(all[i-1].ServerReceivedAt) {
			t.Errorf("page boundary not sorted: [%d]=%v >= [%d]=%v",
				i-1, all[i-1].ServerReceivedAt, i, all[i].ServerReceivedAt)
		}
	}
}

func testSessionUpsertFirstEvent(t *testing.T, factory Factory) {
	ctx := context.Background()
	r := factory(t)
	t0 := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	e := mkEvent(r.Sequencer, "demo", "s1", t0, seq(1))
	if err := r.Sessions.UpsertFromEvents(ctx, []domain.PlaybackEvent{e}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	got, err := r.Sessions.Get(ctx, "s1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.State != domain.SessionStateActive {
		t.Errorf("state = %q, want %q", got.State, domain.SessionStateActive)
	}
	if got.EventCount != 1 {
		t.Errorf("event_count = %d, want 1", got.EventCount)
	}
	if !got.StartedAt.Equal(t0) || !got.LastEventAt.Equal(t0) {
		t.Errorf("times = %v / %v, want both %v", got.StartedAt, got.LastEventAt, t0)
	}
	if got.EndedAt != nil {
		t.Errorf("ended_at = %v, want nil", got.EndedAt)
	}
}

func testSessionTickIncrements(t *testing.T, factory Factory) {
	ctx := context.Background()
	r := factory(t)
	t0 := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	first := mkEvent(r.Sequencer, "demo", "s1", t0, seq(1))
	second := mkEvent(r.Sequencer, "demo", "s1", t0.Add(5*time.Second), seq(2))
	if err := r.Sessions.UpsertFromEvents(ctx, []domain.PlaybackEvent{first, second}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	got, err := r.Sessions.Get(ctx, "s1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.EventCount != 2 {
		t.Errorf("event_count = %d, want 2", got.EventCount)
	}
	if !got.StartedAt.Equal(t0) {
		t.Errorf("started_at = %v, want %v", got.StartedAt, t0)
	}
	if !got.LastEventAt.Equal(t0.Add(5 * time.Second)) {
		t.Errorf("last_event_at = %v, want %v", got.LastEventAt, t0.Add(5*time.Second))
	}
}

func testSessionEndedIdempotent(t *testing.T, factory Factory) {
	ctx := context.Background()
	r := factory(t)
	t0 := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	begin := mkEvent(r.Sequencer, "demo", "s1", t0, seq(1))
	endFirst := mkEvent(r.Sequencer, "demo", "s1", t0.Add(2*time.Second), seq(2))
	endFirst.EventName = persistence.SessionEndedEventName
	endSecond := mkEvent(r.Sequencer, "demo", "s1", t0.Add(5*time.Second), seq(3))
	endSecond.EventName = persistence.SessionEndedEventName

	if err := r.Sessions.UpsertFromEvents(ctx, []domain.PlaybackEvent{begin, endFirst, endSecond}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	got, err := r.Sessions.Get(ctx, "s1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.State != domain.SessionStateEnded {
		t.Errorf("state = %q, want %q", got.State, domain.SessionStateEnded)
	}
	if got.EndedAt == nil {
		t.Fatalf("ended_at = nil, want set")
	}
	if !got.EndedAt.Equal(t0.Add(2 * time.Second)) {
		t.Errorf("ended_at = %v, want %v (first session_ended wins)",
			got.EndedAt, t0.Add(2*time.Second))
	}
	// LastEventAt + EventCount sind weiter inkrementiert.
	if got.EventCount != 3 {
		t.Errorf("event_count = %d, want 3 (verspätete Events zählen weiter)", got.EventCount)
	}
}

func testSweepTransitions(t *testing.T, factory Factory) {
	ctx := context.Background()
	r := factory(t)
	t0 := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	e := mkEvent(r.Sequencer, "demo", "s1", t0, seq(1))
	if err := r.Sessions.UpsertFromEvents(ctx, []domain.PlaybackEvent{e}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Sweep mit kleinem now → noch active
	if err := r.Sessions.Sweep(ctx, t0.Add(10*time.Second), 60*time.Second, 600*time.Second); err != nil {
		t.Fatalf("sweep #1: %v", err)
	}
	got, err := r.Sessions.Get(ctx, "s1")
	if err != nil {
		t.Fatalf("get #1: %v", err)
	}
	if got.State != domain.SessionStateActive {
		t.Errorf("after sweep #1 state = %q, want %q", got.State, domain.SessionStateActive)
	}

	// Sweep weit später → stalled
	if err := r.Sessions.Sweep(ctx, t0.Add(2*time.Minute), 60*time.Second, 600*time.Second); err != nil {
		t.Fatalf("sweep #2: %v", err)
	}
	got, err = r.Sessions.Get(ctx, "s1")
	if err != nil {
		t.Fatalf("get #2: %v", err)
	}
	if got.State != domain.SessionStateStalled {
		t.Errorf("after sweep #2 state = %q, want %q", got.State, domain.SessionStateStalled)
	}

	// Sweep noch später → ended
	if err := r.Sessions.Sweep(ctx, t0.Add(20*time.Minute), 60*time.Second, 600*time.Second); err != nil {
		t.Fatalf("sweep #3: %v", err)
	}
	got, err = r.Sessions.Get(ctx, "s1")
	if err != nil {
		t.Fatalf("get #3: %v", err)
	}
	if got.State != domain.SessionStateEnded {
		t.Errorf("after sweep #3 state = %q, want %q", got.State, domain.SessionStateEnded)
	}
	if got.EndedAt == nil || !got.EndedAt.Equal(t0.Add(20*time.Minute)) {
		t.Errorf("ended_at = %v, want %v", got.EndedAt, t0.Add(20*time.Minute))
	}

	// Idempotenz: nochmal Sweep ändert ended_at nicht.
	if err := r.Sessions.Sweep(ctx, t0.Add(30*time.Minute), 60*time.Second, 600*time.Second); err != nil {
		t.Fatalf("sweep #4: %v", err)
	}
	got, err = r.Sessions.Get(ctx, "s1")
	if err != nil {
		t.Fatalf("get #4: %v", err)
	}
	if !got.EndedAt.Equal(t0.Add(20 * time.Minute)) {
		t.Errorf("after sweep #4 ended_at = %v, want %v (idempotent)",
			got.EndedAt, t0.Add(20*time.Minute))
	}
}

func testSequencerMonotonic(t *testing.T, factory Factory) {
	r := factory(t)
	got := []int64{r.Sequencer.Next(), r.Sequencer.Next(), r.Sequencer.Next()}
	for i, v := range got {
		if v != int64(i+1) {
			t.Errorf("sequencer[%d] = %d, want %d", i, v, i+1)
		}
	}
}

func testSessionListPagination(t *testing.T, factory Factory) {
	ctx := context.Background()
	r := factory(t)
	t0 := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	// Drei Sessions mit aufsteigendem started_at, damit DESC-Sort
	// definiert ist.
	for i, sess := range []string{"s-a", "s-b", "s-c"} {
		e := mkEvent(r.Sequencer, "demo", sess,
			t0.Add(time.Duration(i)*time.Second), seq(1))
		if err := r.Sessions.UpsertFromEvents(ctx, []domain.PlaybackEvent{e}); err != nil {
			t.Fatalf("upsert %s: %v", sess, err)
		}
	}

	page1, err := r.Sessions.List(ctx, driven.SessionListQuery{Limit: 2})
	if err != nil {
		t.Fatalf("list page1: %v", err)
	}
	if len(page1.Sessions) != 2 || page1.NextAfter == nil {
		t.Fatalf("page1 = %+v, want len=2 + cursor", page1)
	}
	// DESC-Sort: zuerst kommt s-c (jüngste), dann s-b.
	if page1.Sessions[0].ID != "s-c" || page1.Sessions[1].ID != "s-b" {
		t.Errorf("page1 ids = [%q, %q], want [s-c, s-b]",
			page1.Sessions[0].ID, page1.Sessions[1].ID)
	}

	page2, err := r.Sessions.List(ctx, driven.SessionListQuery{
		Limit: 10, After: page1.NextAfter,
	})
	if err != nil {
		t.Fatalf("list page2: %v", err)
	}
	if len(page2.Sessions) != 1 || page2.NextAfter != nil {
		t.Fatalf("page2 = %+v, want len=1 + no cursor", page2)
	}
	if page2.Sessions[0].ID != "s-a" {
		t.Errorf("page2 id = %q, want s-a", page2.Sessions[0].ID)
	}
}

func testSessionEndedAsFirstEvent(t *testing.T, factory Factory) {
	ctx := context.Background()
	r := factory(t)
	t0 := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	// Sonderfall: das allererste Event einer Session ist `session_ended`.
	// Adapter muss Session direkt im Ended-State anlegen, ohne den
	// Umweg über einen Active-State.
	e := mkEvent(r.Sequencer, "demo", "s1", t0, seq(1))
	e.EventName = persistence.SessionEndedEventName
	if err := r.Sessions.UpsertFromEvents(ctx, []domain.PlaybackEvent{e}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	got, err := r.Sessions.Get(ctx, "s1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.State != domain.SessionStateEnded {
		t.Errorf("state = %q, want %q", got.State, domain.SessionStateEnded)
	}
	if got.EndedAt == nil || !got.EndedAt.Equal(t0) {
		t.Errorf("ended_at = %v, want %v", got.EndedAt, t0)
	}
	if got.EventCount != 1 {
		t.Errorf("event_count = %d, want 1", got.EventCount)
	}
}

func testEventMetaRoundTrip(t *testing.T, factory Factory) {
	ctx := context.Background()
	r := factory(t)
	t0 := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	e := mkEvent(r.Sequencer, "demo", "s1", t0, seq(1))
	e.Meta = domain.EventMeta{
		"buffer_ms":   float64(1234), // JSON unmarshal als float64
		"is_live":     true,
		"description": "rebuffer at 12s",
	}
	if err := r.Sessions.UpsertFromEvents(ctx, []domain.PlaybackEvent{e}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := r.Events.Append(ctx, []domain.PlaybackEvent{e}); err != nil {
		t.Fatalf("append: %v", err)
	}

	page, err := r.Events.ListBySession(ctx, driven.EventListQuery{
		SessionID: "s1", Limit: 10,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Events) != 1 {
		t.Fatalf("len = %d, want 1", len(page.Events))
	}
	got := page.Events[0].Meta
	if got["buffer_ms"] != float64(1234) {
		t.Errorf("meta[buffer_ms] = %v (%T), want 1234 (float64)",
			got["buffer_ms"], got["buffer_ms"])
	}
	if got["is_live"] != true {
		t.Errorf("meta[is_live] = %v, want true", got["is_live"])
	}
	if got["description"] != "rebuffer at 12s" {
		t.Errorf("meta[description] = %v, want %q", got["description"], "rebuffer at 12s")
	}
}

// --- Helpers -----------------------------------------------------------

func mkEvent(seq driven.IngestSequencer, project, session string, recv time.Time, sequenceNumber *int64) domain.PlaybackEvent {
	return domain.PlaybackEvent{
		EventName:        "playback_started",
		ProjectID:        project,
		SessionID:        session,
		ClientTimestamp:  recv,
		ServerReceivedAt: recv,
		IngestSequence:   seq.Next(),
		SequenceNumber:   sequenceNumber,
		SDK:              domain.SDKInfo{Name: "@npm9912/player-sdk", Version: "0.4.0"},
	}
}

func seq(v int64) *int64 { return &v }
