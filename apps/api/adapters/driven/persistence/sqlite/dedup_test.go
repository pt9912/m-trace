package sqlite_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/sqlite"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// TestDedupClassification verifiziert ADR-0002 §8.3: Events mit
// gesetzter `sequence_number` werden anhand
// `(project_id, session_id, sequence_number)` gegen bereits
// `delivery_status='accepted'`-Einträge geprüft; Duplikate landen mit
// `delivery_status='duplicate_suspected'`. Pfad ist SQLite-spezifisch
// (InMemory kennt keine Klassifikation), deshalb nicht im Contract-
// Test sondern direkt am Adapter.
func TestDedupClassification(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")
	t0 := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	db, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	seq, err := sqlite.NewIngestSequencer(ctx, db)
	if err != nil {
		t.Fatalf("seq: %v", err)
	}
	sess := sqlite.NewSessionRepository(db)
	evt := sqlite.NewEventRepository(db)

	// Project-Row anlegen, damit FK auf stream_sessions hält. Beide
	// Events teilen (project_id, session_id, sequence_number).
	first := mkSeqEvent(seq, "demo", "s1", t0, 1)
	dup := mkSeqEvent(seq, "demo", "s1", t0.Add(time.Second), 1)
	noseq := mkSeqEvent(seq, "demo", "s1", t0.Add(2*time.Second), 0)
	noseq.SequenceNumber = nil

	if err := sess.UpsertFromEvents(ctx, []domain.PlaybackEvent{first, dup, noseq}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := evt.Append(ctx, []domain.PlaybackEvent{first, dup, noseq}); err != nil {
		t.Fatalf("append: %v", err)
	}

	// Direkter DB-Query, weil delivery_status nicht über das
	// driven.EventRepository-Interface exposed wird (Read-Vertrag in
	// API-Kontrakt §3.7 steht erst ab Tranche 4 / 0.4.0-Release).
	rows, err := db.QueryContext(ctx,
		"SELECT ingest_sequence, delivery_status FROM playback_events "+
			"ORDER BY ingest_sequence ASC")
	if err != nil {
		t.Fatalf("query delivery_status: %v", err)
	}
	defer rows.Close()

	var got []struct {
		ingest int64
		status string
	}
	for rows.Next() {
		var r struct {
			ingest int64
			status string
		}
		if err := rows.Scan(&r.ingest, &r.status); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("rows = %d, want 3", len(got))
	}
	if got[0].status != "accepted" {
		t.Errorf("first.status = %q, want accepted", got[0].status)
	}
	if got[1].status != "duplicate_suspected" {
		t.Errorf("dup.status = %q, want duplicate_suspected", got[1].status)
	}
	if got[2].status != "accepted" {
		t.Errorf("noseq.status = %q, want accepted (kein Dedup ohne sequence_number)",
			got[2].status)
	}

	// Sicherstellen, dass die DB tatsächlich genau eine 'accepted'-
	// Zeile pro Dedup-Tuple zurückliefert (Vertrag aus §8.3).
	var acceptedCount int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM playback_events "+
			"WHERE project_id='demo' AND session_id='s1' AND sequence_number=1 "+
			"AND delivery_status='accepted'").Scan(&acceptedCount); err != nil {
		t.Fatalf("count accepted: %v", err)
	}
	if acceptedCount != 1 {
		t.Errorf("accepted count for dedup-key = %d, want 1", acceptedCount)
	}
}

// mkSeqEvent ist ein lokales mkEvent-Pendant für Tests in diesem
// Paket. Setzt SequenceNumber explizit (0 → nil-Mapping über Caller).
func mkSeqEvent(s interface{ Next() int64 }, project, session string, recv time.Time, seqNum int64) domain.PlaybackEvent {
	var sp *int64
	if seqNum > 0 {
		sp = &seqNum
	}
	return domain.PlaybackEvent{
		EventName:        "playback_started",
		ProjectID:        project,
		SessionID:        session,
		ClientTimestamp:  recv,
		ServerReceivedAt: recv,
		IngestSequence:   s.Next(),
		SequenceNumber:   sp,
		SDK:              domain.SDKInfo{Name: "@npm9912/player-sdk", Version: "0.4.0"},
	}
}

// _ = sql.NullInt64{} ist ein bewusst stiller Import — der Adapter
// nutzt sql für seinen Test-Helper indirekt. Diese Zeile dient als
// kompilier-zeitige Garantie, dass `database/sql` verfügbar bleibt.
var _ = sql.ErrNoRows
