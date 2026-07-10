package storage_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// TestOpenPostgres_LiveSchema ist der PG-Lab-Integrationstest (ADR-0006):
// er wendet die eingecheckten Postgres-Migrationen (migrations/postgres/)
// über storage.OpenPostgres gegen eine frische PG-DB an und verifiziert,
// dass die DB alle Live-Tabellen/-Spalten trägt, die zwei 64-bit-PKs als
// bigint, die benannten CHECK-Constraints und einen benutzbaren Zustand.
//
// Übersprungen, wenn MTRACE_PG_LAB_DSN nicht gesetzt ist — der normale
// `go test`/`make test`-Lauf (Coverage-Gate) hat keine Postgres-DB.
// scripts/smoke-pg-lab.sh stellt Container + DSN bereit (`make smoke-pg-lab`).
func TestOpenPostgres_LiveSchema(t *testing.T) {
	dsn := os.Getenv("MTRACE_PG_LAB_DSN")
	if dsn == "" {
		t.Skip("MTRACE_PG_LAB_DSN nicht gesetzt — PG-Lab-Integrationstest übersprungen (siehe scripts/smoke-pg-lab.sh)")
	}
	ctx := context.Background()

	db, err := storage.OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("OpenPostgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	t.Run("schema_migrations baseline V1", func(t *testing.T) { assertBaseline(ctx, t, db) })
	t.Run("table and column inventory", func(t *testing.T) { assertInventory(ctx, t, db) })
	t.Run("64-bit bigint PKs", func(t *testing.T) { assertBigintPKs(ctx, t, db) })
	t.Run("18 named CHECK constraints", func(t *testing.T) { assertChecks(ctx, t, db) })
	t.Run("idempotent re-run", func(t *testing.T) { assertIdempotent(ctx, t, dsn) })
	t.Run("usable schema", func(t *testing.T) { assertUsable(ctx, t, db) })
}

// assertBaseline: schema_migrations trägt genau die PG-Baseline V1, dirty=0.
func assertBaseline(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()
	var version, dirty int64
	var n int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM schema_migrations").Scan(&n); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if n != 1 {
		t.Fatalf("schema_migrations rows = %d, want 1", n)
	}
	if err := db.QueryRowContext(ctx,
		"SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatalf("read schema_migrations: %v", err)
	}
	if version != 1 || dirty != 0 {
		t.Errorf("schema_migrations = {version:%d dirty:%d}, want {1 0}", version, dirty)
	}
}

// assertInventory: alle 13 Live-Tabellen (V1–V7) mit exakter Spaltenzahl,
// keine unerwarteten App-Tabellen. Bricht laut, wenn eine Migration
// Spalten ändert, ohne das PG-DDL nachzuziehen (Drift-Sensor).
func assertInventory(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()
	want := map[string]int{
		"auth_issuance_counters":    6,
		"ingest_endpoints":          7,
		"ingest_routing_rules":      6,
		"ingest_streams":            10,
		"media_server_targets":      5,
		"playback_events":           16,
		"project_token_generations": 10,
		"projects":                  1,
		"srt_health_samples":        20,
		"stream_keys":               7,
		"stream_lifecycle_events":   10,
		"stream_session_boundaries": 8,
		"stream_sessions":           10,
	}
	rows, err := db.QueryContext(ctx,
		"SELECT table_name, count(*) FROM information_schema.columns "+
			"WHERE table_schema = 'public' GROUP BY table_name")
	if err != nil {
		t.Fatalf("query column counts: %v", err)
	}
	defer func() { _ = rows.Close() }()
	got := map[string]int{}
	for rows.Next() {
		var name string
		var count int
		if err := rows.Scan(&name, &count); err != nil {
			t.Fatalf("scan column count: %v", err)
		}
		got[name] = count
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("column counts rows.Err: %v", err)
	}
	for tbl, wantN := range want {
		if got[tbl] != wantN {
			t.Errorf("Tabelle %q: %d Spalten, want %d", tbl, got[tbl], wantN)
		}
	}
	for tbl := range got {
		if _, ok := want[tbl]; !ok && tbl != "schema_migrations" {
			t.Errorf("unerwartete Tabelle in frischer PG-DB: %q", tbl)
		}
	}
}

// assertBigintPKs: die zwei 64-bit-AUTOINCREMENT-PKs sind bigint mit
// nextval-Default (BIGSERIAL, width=64), NICHT int32.
func assertBigintPKs(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()
	for _, c := range []struct{ table, column string }{
		{"playback_events", "ingest_sequence"},
		{"srt_health_samples", "id"},
	} {
		var dataType string
		var colDefault sql.NullString
		if err := db.QueryRowContext(ctx,
			"SELECT data_type, column_default FROM information_schema.columns "+
				"WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2",
			c.table, c.column).Scan(&dataType, &colDefault); err != nil {
			t.Fatalf("query column type %s.%s: %v", c.table, c.column, err)
		}
		if dataType != "bigint" {
			t.Errorf("%s.%s data_type = %q, want %q", c.table, c.column, dataType, "bigint")
		}
		if !strings.HasPrefix(colDefault.String, "nextval(") {
			t.Errorf("%s.%s column_default = %q, want nextval(...) (BIGSERIAL)",
				c.table, c.column, colDefault.String)
		}
	}
}

// assertChecks: alle 18 benannten CHECK-Constraints (chk_*) sind
// materialisiert (der d-migrate-0.9.9-Klammer-Fix trägt bis in die PG-DB).
func assertChecks(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()
	var checks int
	if err := db.QueryRowContext(ctx,
		"SELECT count(*) FROM pg_constraint WHERE contype = 'c' AND conname LIKE 'chk_%'").
		Scan(&checks); err != nil {
		t.Fatalf("count checks: %v", err)
	}
	if checks != 18 {
		t.Errorf("CHECK-Constraints = %d, want 18", checks)
	}
}

// assertIdempotent: ein zweiter OpenPostgres-Lauf ist ein No-op (keine
// offenen Migrationen, schema_migrations bleibt bei einer Zeile V1).
func assertIdempotent(ctx context.Context, t *testing.T, dsn string) {
	t.Helper()
	db, err := storage.OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("OpenPostgres #2: %v", err)
	}
	defer func() { _ = db.Close() }()
	var n int
	var version int64
	if err := db.QueryRowContext(ctx,
		"SELECT count(*), max(version) FROM schema_migrations").Scan(&n, &version); err != nil {
		t.Fatalf("read schema_migrations #2: %v", err)
	}
	if n != 1 || version != 1 {
		t.Errorf("after re-run: %d rows, max version %d; want 1 row version 1", n, version)
	}
}

// assertUsable: app-vergebene ingest_sequence (Pre-Assign-Flow),
// CHECK-Ablehnung, FK-Ablehnung.
func assertUsable(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.ExecContext(ctx,
		"INSERT INTO projects(project_id) VALUES ($1)", "p1"); err != nil {
		t.Fatalf("insert project: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		"INSERT INTO playback_events(ingest_sequence, project_id, session_id, "+
			"event_name, client_timestamp, server_received_at, sdk_name, "+
			"sdk_version, schema_version) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		int64(1), "p1", "s1", "rebuffer_started",
		"2026-07-10T10:00:00Z", "2026-07-10T10:00:01Z",
		"@pt9912/player-sdk", "0.4.0", "1.0"); err != nil {
		t.Fatalf("insert event (pre-assigned ingest_sequence): %v", err)
	}
	if _, err := db.ExecContext(ctx,
		"INSERT INTO playback_events(ingest_sequence, project_id, session_id, "+
			"event_name, client_timestamp, server_received_at, sdk_name, "+
			"sdk_version, schema_version, delivery_status) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		int64(2), "p1", "s1", "x",
		"2026-07-10T10:00:00Z", "2026-07-10T10:00:01Z",
		"@pt9912/player-sdk", "0.4.0", "1.0", "bogus"); err == nil {
		t.Error("expected CHECK violation for delivery_status='bogus', got nil")
	}
	if _, err := db.ExecContext(ctx,
		"INSERT INTO stream_sessions(session_id, project_id, started_at, last_seen_at) "+
			"VALUES ($1, $2, $3, $4)",
		"s2", "missing_project",
		"2026-07-10T10:00:00Z", "2026-07-10T10:00:00Z"); err == nil {
		t.Error("expected FK violation for missing project_id, got nil")
	}
}
