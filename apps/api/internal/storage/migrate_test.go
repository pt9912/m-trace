package storage

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"sort"
	"testing"
	"testing/fstest"
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
		"stream_sessions",
	}
	if got := tableNames(t, db); !equalSlices(got, wantTables) {
		t.Errorf("tables = %v, want superset of %v", got, wantTables)
	}

	rows, err := allMigrationRows(ctx, db)
	if err != nil {
		t.Fatalf("read schema_migrations: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("schema_migrations rows = %d, want 1", len(rows))
	}
	if rows[0].version != 1 || rows[0].dirty != 0 {
		t.Errorf("row = %+v, want version=1 dirty=0", rows[0])
	}
}

// TestOpen_ReRunIsNoop prüft, dass ein zweites Open() gegen dieselbe
// Datei keinen Fehler wirft und schema_migrations unverändert bleibt.
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
		"0001_broken.sql": &fstest.MapFile{
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
		"0008_after_dirty.sql": &fstest.MapFile{
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
	defer rows.Close()
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
	defer rows.Close()
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
