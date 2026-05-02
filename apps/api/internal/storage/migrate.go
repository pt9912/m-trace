// Package storage embeds und applied SQLite-Migrationen für m-trace.
//
// Der Apply-Runner ist absichtlich klein: er liest SQL-Files aus dem
// embed.FS, vergleicht mit der schema_migrations-Tabelle, und wendet
// offene Versionen in einer einzelnen Transaktion pro Migration an.
// Die schema_migrations-Tabelle wird hier verwaltet (ADR-0002 §8.1)
// und erscheint nicht in der schema.yaml.
//
// Race-Schutz für konkurrierende Writer ergibt sich aus
// `_txlock=immediate` in der DSN: SQLite akquiriert beim Transaktions-
// start sofort den Write-Lock, alle anderen Writer blockieren bis zum
// Commit (ADR-0002 §8.3).
package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

const driverName = "sqlite"

// nowFn liefert den aktuellen Zeitpunkt für `applied_at`-Einträge.
// In Tests überschreibbar, damit Reihenfolge-Assertions deterministisch
// werden; Default ist `time.Now`.
var nowFn = time.Now

// migrationNamePattern matcht "0001_initial.sql", "0002_xxx.sql" etc.
// Die ersten vier Ziffern sind die Versionsnummer.
var migrationNamePattern = regexp.MustCompile(`^(\d{4})_.+\.sql$`)

// ErrSchemaDirty wird zurückgegeben, wenn schema_migrations einen
// Eintrag mit dirty=1 hat. Der API-Start refused dann; Reparatur ist
// in docs/user/local-development.md beschrieben.
var ErrSchemaDirty = errors.New("storage: schema is in dirty state")

// Open öffnet (oder erzeugt) die SQLite-Datei unter path, sichert
// Foreign-Keys und WAL-Mode, wendet alle offenen Migrationen aus dem
// eingebetteten migrations/-Verzeichnis an und gibt die fertige
// *sql.DB zurück.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	dsn := buildDSN(path)
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: open %q: %w", path, err)
	}
	if err := setPragmas(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	sub, err := fs.Sub(embeddedMigrations, "migrations")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("storage: embed sub: %w", err)
	}
	if err := apply(ctx, db, sub); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// Apply wendet offene Migrationen aus files (fs.FS, gewurzelt am
// Verzeichnis mit den SQL-Files) auf die übergebene *sql.DB an.
// Test-freundlich: fstest.MapFS einsetzbar.
func Apply(ctx context.Context, db *sql.DB, files fs.FS) error {
	return apply(ctx, db, files)
}

// buildDSN setzt _txlock=immediate (BEGIN IMMEDIATE für Race-Schutz),
// _pragma=foreign_keys(ON) und WAL-Journal. Der Pfad bleibt clean
// (file:-URI nur bei Bedarf, hier Plain-Path).
func buildDSN(path string) string {
	return path + "?_txlock=immediate" +
		"&_pragma=foreign_keys(ON)" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=busy_timeout(5000)"
}

// setPragmas verifiziert, dass die DSN-PRAGMAs gegriffen haben. Wir
// verlassen uns nicht blind auf den Driver — bei einer Driver-
// Konfigurationsabweichung würde die Test-Suite das sonst nicht sehen.
func setPragmas(ctx context.Context, db *sql.DB) error {
	checks := []struct{ pragma, want string }{
		{"foreign_keys", "1"},
		{"journal_mode", "wal"},
	}
	for _, c := range checks {
		var got string
		if err := db.QueryRowContext(ctx, "PRAGMA "+c.pragma).Scan(&got); err != nil {
			return fmt.Errorf("storage: pragma %s: %w", c.pragma, err)
		}
		if !strings.EqualFold(got, c.want) {
			return fmt.Errorf("storage: pragma %s = %q, want %q", c.pragma, got, c.want)
		}
	}
	return nil
}

func apply(ctx context.Context, db *sql.DB, files fs.FS) error {
	if err := ensureSchemaMigrationsTable(ctx, db); err != nil {
		return err
	}
	if version, ok, err := dirtyVersion(ctx, db); err != nil {
		return err
	} else if ok {
		return fmt.Errorf("%w (version %d); manual repair required (see docs/user/local-development.md)",
			ErrSchemaDirty, version)
	}
	pending, err := pendingMigrations(ctx, db, files)
	if err != nil {
		return err
	}
	for _, m := range pending {
		if err := applyOne(ctx, db, m); err != nil {
			// dirty=1 separat persistieren (eigene Transaktion); ein
			// Fehler dabei darf den ursprünglichen Fehler nicht
			// verschlucken.
			if markErr := markDirty(ctx, db, m.version); markErr != nil {
				return fmt.Errorf("storage: apply %s: %w (failed to record dirty state: %v)",
					m.name, err, markErr)
			}
			return fmt.Errorf("storage: apply %s: %w", m.name, err)
		}
	}
	return nil
}

const schemaMigrationsDDL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    INTEGER PRIMARY KEY,
    applied_at TEXT    NOT NULL,
    dirty      INTEGER NOT NULL DEFAULT 0
)`

func ensureSchemaMigrationsTable(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schemaMigrationsDDL); err != nil {
		return fmt.Errorf("storage: create schema_migrations: %w", err)
	}
	return nil
}

type migration struct {
	version int64
	name    string
	body    string
}

func pendingMigrations(ctx context.Context, db *sql.DB, files fs.FS) ([]migration, error) {
	applied, err := appliedVersions(ctx, db)
	if err != nil {
		return nil, err
	}

	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return nil, fmt.Errorf("storage: read migrations dir: %w", err)
	}
	var out []migration
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		match := migrationNamePattern.FindStringSubmatch(e.Name())
		if match == nil {
			continue
		}
		v, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			continue
		}
		if applied[v] {
			continue
		}
		body, err := fs.ReadFile(files, e.Name())
		if err != nil {
			return nil, fmt.Errorf("storage: read %s: %w", e.Name(), err)
		}
		out = append(out, migration{version: v, name: e.Name(), body: string(body)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}

func appliedVersions(ctx context.Context, db *sql.DB) (map[int64]bool, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT version FROM schema_migrations WHERE dirty = 0")
	if err != nil {
		return nil, fmt.Errorf("storage: query schema_migrations: %w", err)
	}
	defer rows.Close()
	out := map[int64]bool{}
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("storage: scan schema_migrations: %w", err)
		}
		out[v] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: scan schema_migrations: %w", err)
	}
	return out, nil
}

func applyOne(ctx context.Context, db *sql.DB, m migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	if _, err := tx.ExecContext(ctx, m.body); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO schema_migrations(version, applied_at, dirty) VALUES (?, ?, 0)",
		m.version, nowFn().UTC().Format(time.RFC3339Nano)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func markDirty(ctx context.Context, db *sql.DB, version int64) error {
	_, err := db.ExecContext(ctx,
		"INSERT INTO schema_migrations(version, applied_at, dirty) VALUES (?, ?, 1) "+
			"ON CONFLICT(version) DO UPDATE SET dirty = 1, applied_at = excluded.applied_at",
		version, nowFn().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("storage: mark dirty: %w", err)
	}
	return nil
}

func dirtyVersion(ctx context.Context, db *sql.DB) (int64, bool, error) {
	var version int64
	err := db.QueryRowContext(ctx,
		"SELECT version FROM schema_migrations WHERE dirty = 1 ORDER BY version LIMIT 1").
		Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("storage: query dirty: %w", err)
	}
	return version, true, nil
}
