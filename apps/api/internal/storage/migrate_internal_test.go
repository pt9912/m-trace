package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"testing/fstest"
	"time"
)

// TestOpen_FreshStart prüft, dass Open() gegen eine leere Datei das
// volle Schema anlegt und schema_migrations einen Eintrag mit dirty=0
// für die einzige Migration enthält.
func TestOpen_FreshStart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")

	db, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	wantTables := []string{
		"playback_events",
		"projects",
		"schema_migrations",
		"stream_session_boundaries",
		"stream_sessions",
	}
	if got := tableNames(t, db); !equalSlices(got, wantTables) {
		t.Errorf("tables = %v, want superset of %v", got, wantTables)
	}

	rows, err := allMigrationRows(ctx, db)
	if err != nil {
		t.Fatalf("read schema_migrations: %v", err)
	}
	// Ab plan-0.4.0 §4.4 D2 drei Migrations-Files (V1 baseline + V2
	// project-scoped session PK + V3 session_boundaries); ein
	// Fresh-Start läuft alle drei an.
	if len(rows) != 3 {
		t.Fatalf("schema_migrations rows = %d, want 3", len(rows))
	}
	if rows[0].version != 1 || rows[0].dirty != 0 {
		t.Errorf("row[0] = %+v, want version=1 dirty=0", rows[0])
	}
	if rows[1].version != 2 || rows[1].dirty != 0 {
		t.Errorf("row[1] = %+v, want version=2 dirty=0", rows[1])
	}
	if rows[2].version != 3 || rows[2].dirty != 0 {
		t.Errorf("row[2] = %+v, want version=3 dirty=0", rows[2])
	}
}

// TestOpen_ReRunIsNoop prüft, dass ein zweites Open() gegen dieselbe
// Datei keinen Fehler wirft, schema_migrations unverändert bleibt
// **und** das Schema nach Re-Run weiter benutzbar ist (FK,
// CHECK-Constraint, AUTOINCREMENT).
func TestOpen_ReRunIsNoop(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")

	first, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("Open #1: %v", err)
	}
	rowsBefore, err := allMigrationRows(ctx, first)
	if err != nil {
		t.Fatalf("read schema_migrations #1: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close #1: %v", err)
	}

	second, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("Open #2: %v", err)
	}
	t.Cleanup(func() { _ = second.Close() })
	rowsAfter, err := allMigrationRows(ctx, second)
	if err != nil {
		t.Fatalf("read schema_migrations #2: %v", err)
	}
	if !equalRows(rowsBefore, rowsAfter) {
		t.Errorf("rows changed: before=%v after=%v", rowsBefore, rowsAfter)
	}

	// Schema-Nutzbarkeit: INSERT auf projects + playback_events.
	if _, err := second.ExecContext(ctx,
		"INSERT INTO projects(project_id) VALUES (?)", "p1"); err != nil {
		t.Fatalf("insert project: %v", err)
	}
	if _, err := second.ExecContext(ctx,
		"INSERT INTO playback_events(project_id, session_id, event_name, "+
			"client_timestamp, server_received_at, sdk_name, sdk_version, "+
			"schema_version) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"p1", "s1", "rebuffer_started",
		"2026-05-02T10:00:00Z", "2026-05-02T10:00:01Z",
		"@npm9912/player-sdk", "0.4.0", "1.0"); err != nil {
		t.Fatalf("insert event: %v", err)
	}

	// AUTOINCREMENT: erste Zeile bekommt ingest_sequence = 1.
	var ing int64
	if err := second.QueryRowContext(ctx,
		"SELECT ingest_sequence FROM playback_events WHERE session_id = ?",
		"s1").Scan(&ing); err != nil {
		t.Fatalf("query ingest_sequence: %v", err)
	}
	if ing != 1 {
		t.Errorf("ingest_sequence = %d, want 1 (AUTOINCREMENT)", ing)
	}

	// CHECK-Constraint: ungültiger delivery_status muss abgelehnt werden.
	if _, err := second.ExecContext(ctx,
		"INSERT INTO playback_events(project_id, session_id, event_name, "+
			"client_timestamp, server_received_at, sdk_name, sdk_version, "+
			"schema_version, delivery_status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"p1", "s1", "x",
		"2026-05-02T10:00:00Z", "2026-05-02T10:00:01Z",
		"@npm9912/player-sdk", "0.4.0", "1.0", "bogus"); err == nil {
		t.Error("expected CHECK violation for delivery_status='bogus', got nil")
	}

	// FK-Constraint: stream_sessions referenziert nicht-existentes Projekt.
	if _, err := second.ExecContext(ctx,
		"INSERT INTO stream_sessions(session_id, project_id, started_at, "+
			"last_seen_at) VALUES (?, ?, ?, ?)",
		"s2", "missing_project",
		"2026-05-02T10:00:00Z", "2026-05-02T10:00:00Z"); err == nil {
		t.Error("expected FK violation for missing project_id, got nil")
	}
}

// TestApply_FailureMarksDirty prüft, dass eine kaputte Migration
// schema_migrations.dirty=1 setzt und der Fehler weitergereicht wird.
func TestApply_FailureMarksDirty(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")

	db, err := openBareDB(ctx, path)
	if err != nil {
		t.Fatalf("openBareDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	files := fstest.MapFS{
		"V1__broken.sql": &fstest.MapFile{
			Data: []byte("THIS IS NOT VALID SQL;"),
		},
	}

	err = Apply(ctx, db, files)
	if err == nil {
		t.Fatal("Apply: expected error, got nil")
	}

	rows, err := allMigrationRows(ctx, db)
	if err != nil {
		t.Fatalf("read schema_migrations: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].version != 1 || rows[0].dirty != 1 {
		t.Errorf("row = %+v, want version=1 dirty=1", rows[0])
	}
}

// TestApply_DirtyStateRefuses prüft, dass ein nachfolgender Apply()
// gegen eine dirty=1-DB ErrSchemaDirty zurückgibt.
func TestApply_DirtyStateRefuses(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")

	db, err := openBareDB(ctx, path)
	if err != nil {
		t.Fatalf("openBareDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// dirty=1-Eintrag manuell setzen.
	if err := ensureSchemaMigrationsTable(ctx, db); err != nil {
		t.Fatalf("ensureSchemaMigrationsTable: %v", err)
	}
	if err := markDirty(ctx, db, 7); err != nil {
		t.Fatalf("markDirty: %v", err)
	}

	files := fstest.MapFS{
		"V8__after_dirty.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE later (id INTEGER);"),
		},
	}

	err = Apply(ctx, db, files)
	if !errors.Is(err, ErrSchemaDirty) {
		t.Fatalf("Apply: error = %v, want ErrSchemaDirty", err)
	}

	// Sicherstellen, dass keine Folge-Migration angewandt wurde.
	if names := tableNames(t, db); contains(names, "later") {
		t.Errorf("later table created despite dirty refuse-to-start")
	}
}

// TestApply_MultiStatementRollback prüft, dass ein Failure mitten in
// einer Multi-Statement-Migration die ganze Migration rolled-back —
// das erste CREATE TABLE darf NICHT übrig bleiben, wenn das zweite
// scheitert. Andernfalls wäre `tx.Rollback()` in `applyOne` wirkungslos.
func TestApply_MultiStatementRollback(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")

	db, err := openBareDB(ctx, path)
	if err != nil {
		t.Fatalf("openBareDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// Erstes CREATE wäre erfolgreich; zweites scheitert (Tabellenname
	// `a` existiert bereits aus Statement 1 derselben Tx).
	files := fstest.MapFS{
		"V1__multi.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE a(id INTEGER); CREATE TABLE a(id INTEGER);"),
		},
	}

	if err := Apply(ctx, db, files); err == nil {
		t.Fatal("Apply: expected error, got nil")
	}
	if contains(tableNames(t, db), "a") {
		t.Error("table 'a' was created despite multi-statement rollback")
	}
	rows, err := allMigrationRows(ctx, db)
	if err != nil {
		t.Fatalf("read schema_migrations: %v", err)
	}
	if len(rows) != 1 || rows[0].version != 1 || rows[0].dirty != 1 {
		t.Errorf("rows = %+v, want one row with version=1 dirty=1", rows)
	}
}

// TestApply_ConcurrentWritersDoNotDeadlock startet zwei Goroutinen,
// die parallel je eine Schreib-Tx (`db.BeginTx` → INSERT → Commit)
// gegen dieselbe DB ausführen. Mit `_txlock=immediate` aus
// `buildDSN` serialisiert SQLite die Writer per DB-Lock; ohne diese
// DSN-Konfig würde der Test mit `database is locked`-Fehlern flaky.
// Beweis ist nicht "parallel exakt", sondern "kein Deadlock, beide
// Tx committen".
func TestApply_ConcurrentWritersDoNotDeadlock(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")

	db, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	const writers = 2
	var (
		wg      sync.WaitGroup
		barrier = make(chan struct{})
		errs    = make(chan error, writers)
	)
	for i := 0; i < writers; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-barrier
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				errs <- fmt.Errorf("begin %d: %w", i, err)
				return
			}
			// Kurze Pause, damit beide Goroutinen ihre Tx parallel
			// gestartet haben, bevor der Insert ausgeführt wird.
			time.Sleep(10 * time.Millisecond)
			if _, err := tx.ExecContext(ctx,
				"INSERT INTO projects(project_id) VALUES (?)",
				fmt.Sprintf("p%d", i)); err != nil {
				_ = tx.Rollback()
				errs <- fmt.Errorf("insert %d: %w", i, err)
				return
			}
			if err := tx.Commit(); err != nil {
				errs <- fmt.Errorf("commit %d: %w", i, err)
				return
			}
			errs <- nil
		}()
	}
	close(barrier)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Errorf("concurrent writer: %v", err)
		}
	}

	var count int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM projects").Scan(&count); err != nil {
		t.Fatalf("count projects: %v", err)
	}
	if count != writers {
		t.Errorf("projects count = %d, want %d", count, writers)
	}
}

// openBareDB öffnet die SQLite-Datei mit der gleichen DSN wie Open(),
// wendet aber keine Migrationen an. Nur für Test-Setup.
func openBareDB(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open(driverName, buildDSN(path))
	if err != nil {
		return nil, err
	}
	if err := setPragmas(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

type migrationRow struct {
	version int64
	dirty   int64
}

func allMigrationRows(ctx context.Context, db *sql.DB) ([]migrationRow, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT version, dirty FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []migrationRow
	for rows.Next() {
		var r migrationRow
		if err := rows.Scan(&r.version, &r.dirty); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func tableNames(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query(
		"SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		t.Fatalf("query tables: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	sort.Strings(out)
	return out
}

func equalSlices(got, wantSubset []string) bool {
	have := map[string]bool{}
	for _, s := range got {
		have[s] = true
	}
	for _, s := range wantSubset {
		if !have[s] {
			return false
		}
	}
	return true
}

func equalRows(a, b []migrationRow) bool {
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

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
