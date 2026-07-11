package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/postgres"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// TestEventRepository_R27Watermark_PgLab ist der adversariale R-27-Test:
// ein spät-committendes Früh-Event darf die Keyset-Pagination nicht
// "zerreißen". Szenario: E1(sra=1), E3(sra=3), E5(sra=5) sind committed;
// nach Page 1 (die W erfasst) wird B(sra=2) committed — B sortiert direkt
// hinter dem Cursor (Keyset WÜRDE B liefern), aber commit_ts(B) > W.
//
// Beweis, dass W wirkt (distinktiv, nicht nur Sichtbarkeit): Page 2 der
// W-Session liefert E3, NICHT B (das Wasserzeichen hält B aktiv zurück,
// obwohl B sichtbar+sortier-nächst ist). Die W-Session ist damit ein
// konsistenter Snapshot [E1,E3,E5]. Eine frische Session (W'>commit_ts(B))
// liefert B an korrekter Position [E1,B,E3,E5] → B wird nie übersprungen.
// Braucht track_commit_timestamp=on (smoke-pg-lab-Harness). Gated über
// MTRACE_PG_LAB_DSN.
func TestEventRepository_R27Watermark_PgLab(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, `TRUNCATE playback_events, projects RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO projects(project_id) VALUES ('r27') ON CONFLICT DO NOTHING`); err != nil {
		t.Fatalf("seed project: %v", err)
	}

	repo := postgres.NewEventRepository(db)
	const proj, sess = "r27", "s1"
	base := time.Date(2026, 7, 11, 15, 0, 0, 0, time.UTC)
	mk := func(ing int64, sraSec int) domain.PlaybackEvent {
		return domain.PlaybackEvent{
			EventName: "playback_progress", ProjectID: proj, SessionID: sess,
			ClientTimestamp: base, ServerReceivedAt: base.Add(time.Duration(sraSec) * time.Second),
			IngestSequence: ing, SDK: domain.SDKInfo{Name: "js", Version: "1.0.0"},
		}
	}
	// E1/E3/E5 committed VOR Page 1.
	if err := repo.Append(ctx, []domain.PlaybackEvent{mk(10, 1), mk(30, 3), mk(50, 5)}); err != nil {
		t.Fatalf("Append E1/E3/E5: %v", err)
	}

	// Page 1 (limit 1) → [E1]; erfasst W und trägt es im Cursor.
	p1, err := repo.ListBySession(ctx, driven.EventListQuery{ProjectID: proj, SessionID: sess, Limit: 1})
	if err != nil {
		t.Fatalf("page1: %v", err)
	}
	if len(p1.Events) != 1 || p1.Events[0].IngestSequence != 10 {
		t.Fatalf("page1 = %v, want [E1(10)]", ingSeqs(p1.Events))
	}
	if p1.NextAfter == nil || p1.NextAfter.Watermark == nil {
		t.Fatalf("page1 NextAfter/Watermark = %+v, want gesetzt", p1.NextAfter)
	}

	// B(sra=2) committed NACH Page 1 → commit_ts(B) > W. B sortiert direkt
	// hinter dem Cursor (E1, sra=1): ohne Wasserzeichen käme B auf Page 2.
	if err := repo.Append(ctx, []domain.PlaybackEvent{mk(20, 2)}); err != nil {
		t.Fatalf("Append B: %v", err)
	}

	// KERN-ASSERTION: Page 2 der W-Session liefert E3, NICHT B.
	p2, err := repo.ListBySession(ctx, driven.EventListQuery{ProjectID: proj, SessionID: sess, Limit: 1, After: p1.NextAfter})
	if err != nil {
		t.Fatalf("page2: %v", err)
	}
	if len(p2.Events) != 1 || p2.Events[0].IngestSequence != 30 {
		t.Fatalf("page2 = %v, want [E3(30)] — B(20) müsste vom Wasserzeichen zurückgehalten werden", ingSeqs(p2.Events))
	}

	// Volle W-Session ab Page 1 einsammeln → konsistenter Snapshot [10,30,50], kein B.
	wSession := append([]int64{10}, collectPagesR27(ctx, t, repo, proj, sess, p1.NextAfter)...)
	if !equalInt64s(wSession, []int64{10, 30, 50}) {
		t.Errorf("W-Session = %v, want [10 30 50] (B zurückgehalten)", wSession)
	}

	// Frische Session (neues W' > commit_ts(B)) → B an korrekter Position.
	fresh, err := repo.ListBySession(ctx, driven.EventListQuery{ProjectID: proj, SessionID: sess, Limit: 10})
	if err != nil {
		t.Fatalf("fresh: %v", err)
	}
	if got := ingSeqs(fresh.Events); !equalInt64s(got, []int64{10, 20, 30, 50}) {
		t.Errorf("frische Session = %v, want [10 20 30 50] (B an sra=2 sichtbar → nicht übersprungen)", got)
	}
}

// collectPagesR27 blättert ab `after` (mit dessen getragenem W) alle
// weiteren Pages zu je 1 Event durch und liefert die ingest_sequences.
func collectPagesR27(ctx context.Context, t *testing.T, repo *postgres.EventRepository, proj, sess string, after *driven.EventCursorPosition) []int64 {
	t.Helper()
	var out []int64
	for after != nil {
		p, err := repo.ListBySession(ctx, driven.EventListQuery{ProjectID: proj, SessionID: sess, Limit: 1, After: after})
		if err != nil {
			t.Fatalf("collect page: %v", err)
		}
		out = append(out, ingSeqs(p.Events)...)
		after = p.NextAfter
	}
	return out
}

func ingSeqs(events []domain.PlaybackEvent) []int64 {
	out := make([]int64, len(events))
	for i, e := range events {
		out[i] = e.IngestSequence
	}
	return out
}

func equalInt64s(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
