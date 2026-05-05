package sqlite_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/sqlite"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// TestUpsertFromEvents_RaceCanonicalCorrelationID schließt R-6:
// konkurrente Erst-Batches derselben `(project_id, session_id)` müssen
// alle dieselbe DB-finale `stream_sessions.correlation_id`
// zurückbekommen (`UpsertFromEvents`-Rückgabewert), sodass der Use-
// Case Events vor `EventRepository.Append` mit dieser Sieger-CID
// enrichen kann. Vorher hat jeder Race-Aufruf eine eigene Kandidat-
// UUID generiert; Events des Verlust-Aufrufs trugen damit eine andere
// CID als `stream_sessions.correlation_id` (R-6 im risks-backlog).
//
// Test-Strategie: N Goroutines schreiben parallel ein Single-Event-
// Batch für dieselbe Session, jeweils mit einer eigenen Kandidat-CID.
// Erwartet:
//   - Alle Goroutines kommen ohne Fehler zurück (kein UNIQUE-Verstoß
//     auf dem Composite-PK).
//   - Alle zurückgegebenen `canonical[sessionID]`-Werte sind identisch
//     (genau eine Sieger-CID).
//   - In `stream_sessions` existiert genau eine Zeile mit
//     `correlation_id = <Sieger-CID>`.
//
// Der Test läuft gegen das echte SQLite-Backend, weil InMemory keinen
// `ON CONFLICT`-Pfad hat — die R-6-Garantie ist SQLite-spezifisch.
func TestUpsertFromEvents_RaceCanonicalCorrelationID(t *testing.T) {
	t.Parallel()

	const (
		concurrency = 8
		projectID   = "demo"
		sessionID   = "01J7K9X4Z2QHB6V3WS5R8Y4D1F"
	)

	ctx := context.Background()
	db := openRaceDB(ctx, t)
	repo := sqlite.NewSessionRepository(db)
	t0 := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)

	results := runRaceUpserts(ctx, t, repo, concurrency, projectID, sessionID, t0)
	winner := assertCanonicalWinner(t, results, concurrency)
	assertSingleSessionRow(ctx, t, db, projectID, sessionID)
	assertWinnerMatchesPersisted(ctx, t, repo, projectID, sessionID, winner)
}

// openRaceDB öffnet eine frische SQLite-Datei in t.TempDir().
func openRaceDB(ctx context.Context, t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "race.db")
	db, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// runRaceUpserts startet `concurrency` Goroutines, jede schickt einen
// Single-Event-Batch mit eigener Kandidat-CID. Liefert die je
// Goroutine zurückgegebene `canonical[sessionID]`-CID.
func runRaceUpserts(
	ctx context.Context, t *testing.T, repo *sqlite.SessionRepository,
	concurrency int, projectID, sessionID string, t0 time.Time,
) []string {
	t.Helper()
	results := make([]string, concurrency)
	errs := make([]error, concurrency)
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			canonical, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{{
				ProjectID:        projectID,
				SessionID:        sessionID,
				EventName:        "playback_started",
				ClientTimestamp:  t0,
				ServerReceivedAt: t0,
				IngestSequence:   int64(idx + 1),
				CorrelationID:    candidateCID(idx),
				SDK:              domain.SDKInfo{Name: "@npm9912/player-sdk", Version: "0.5.0"},
			}})
			if err != nil {
				errs[idx] = err
				return
			}
			results[idx] = canonical[sessionID]
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: %v", i, err)
		}
	}
	if t.Failed() {
		t.FailNow()
	}
	return results
}

// assertCanonicalWinner verifiziert, dass alle Goroutines dieselbe CID
// zurückbekommen und dass diese CID aus dem Kandidat-Pool stammt.
func assertCanonicalWinner(t *testing.T, results []string, concurrency int) string {
	t.Helper()
	winner := results[0]
	if winner == "" {
		t.Fatalf("goroutine 0 returned empty canonical CID")
	}
	for i, got := range results {
		if got != winner {
			t.Errorf("goroutine %d canonical=%q, want %q (split-brain)", i, got, winner)
		}
	}
	for i := 0; i < concurrency; i++ {
		if winner == candidateCID(i) {
			return winner
		}
	}
	t.Errorf("winner %q is not from the candidate pool", winner)
	return winner
}

// assertSingleSessionRow prüft direkt in der DB, dass für
// (project_id, session_id) genau eine Zeile existiert — weil
// `repo.Get` nur die erste matchende Zeile liefert und einen
// Composite-PK-Verstoß so nicht aufdecken würde.
func assertSingleSessionRow(ctx context.Context, t *testing.T, db *sql.DB, projectID, sessionID string) {
	t.Helper()
	var rowCount int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM stream_sessions WHERE project_id = ? AND session_id = ?",
		projectID, sessionID,
	).Scan(&rowCount); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if rowCount != 1 {
		t.Errorf("stream_sessions row count = %d, want 1", rowCount)
	}
}

// assertWinnerMatchesPersisted lädt die DB-finale Zeile via repo.Get
// und stellt sicher, dass `correlation_id` die Sieger-CID trägt.
func assertWinnerMatchesPersisted(ctx context.Context, t *testing.T, repo *sqlite.SessionRepository, projectID, sessionID, winner string) {
	t.Helper()
	got, err := repo.Get(ctx, projectID, sessionID)
	if err != nil {
		t.Fatalf("repo.Get after race: %v", err)
	}
	if got.CorrelationID != winner {
		t.Errorf("stream_sessions.correlation_id = %q, want %q (canonical mismatch)", got.CorrelationID, winner)
	}
}

// candidateCID erzeugt eine deterministische, je-Goroutine-eindeutige
// UUIDv4-formatierte CorrelationID. Test-only — nutzt keinen
// crypto/rand, sondern einen festen Index, damit der Test reproduzierbar
// bleibt und der Sieger-Check eindeutig ist.
func candidateCID(idx int) string {
	const suffix = "-0000-4000-8000-000000000000"
	// 8 Hex-Stellen = 32 bits — reicht für test-Concurrency-Range.
	return makeHex8(idx) + suffix
}

func makeHex8(idx int) string {
	const hexDigits = "0123456789abcdef"
	out := make([]byte, 8)
	for i := 0; i < 8; i++ {
		out[7-i] = hexDigits[idx&0x0f]
		idx >>= 4
	}
	return string(out)
}
