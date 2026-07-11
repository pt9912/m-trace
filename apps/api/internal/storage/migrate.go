// Package storage embeds und applied Schema-Migrationen für m-trace.
//
// Der Apply-Runner ist absichtlich klein: er liest SQL-Files aus dem
// embed.FS, vergleicht mit der schema_migrations-Tabelle, und wendet
// offene Versionen in einer einzelnen Transaktion pro Migration an.
// Die schema_migrations-Tabelle wird hier verwaltet (ADR-0002)
// und erscheint nicht in der schema.yaml.
//
// Der Runner ist dialekt-parametrisiert (ADR-0006): derselbe
// Apply-Pfad läuft gegen SQLite (Default, `Open`) und Postgres
// (`OpenPostgres`). Die dialekt-spezifischen Teile bündelt `dialect`
// (Placeholder-Stil, Migrations-Unterordner); die Verbindungs-Specs
// (Treiber, DSN, PRAGMA-Bedarf) hängen an der jeweiligen Open-Funktion.
//
// Race-Schutz für konkurrierende Writer ergibt sich auf SQLite aus
// `_txlock=immediate` in der DSN: SQLite akquiriert beim Transaktions-
// start sofort den Write-Lock, alle anderen Writer blockieren bis zum
// Commit (ADR-0002). Auf Postgres übernimmt das der MVCC-Row-Lock der
// jeweiligen Migrations-Transaktion.
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

	// modernc.org/sqlite registriert per init den "sqlite"-
	// database/sql-Treiber. Der Blank-Import ist die idiomatische Form
	// (kein direktes Paket-Symbol nötig), siehe ADR-0002
	_ "modernc.org/sqlite"

	// pgx/v5 stellt den Postgres-Treiber für den optionalen
	// Postgres-Runtime-Adapter (ADR-0006). `stdlib.OpenDB` öffnet eine
	// database/sql-DB, `pgx.ParseConfig` erlaubt das Setzen des
	// Query-Exec-Modes (siehe OpenPostgres).
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql migrations/postgres/*.sql
var embeddedMigrations embed.FS

// driverName ist der database/sql-Treiber des SQLite-Default-Pfads.
// Postgres öffnet nicht über einen Treiber-Namen, sondern direkt via
// stdlib.OpenDB (siehe OpenPostgres).
const driverName = "sqlite"

// dialect bündelt die dialekt-spezifischen Teile des Apply-Runners.
// placeholder rendert den n-ten Bind-Parameter (`?` für SQLite,
// `$n` für Postgres); migrationsDir wählt den Unterordner im embed.FS.
type dialect struct {
	placeholder   func(n int) string
	migrationsDir string
	// bodyExecArgs werden dem Exec des (Multi-Statement-)Migrations-
	// Bodys vorangestellt. Für Postgres trägt es
	// pgx.QueryExecModeSimpleProtocol: nur der Migrations-Body braucht
	// das Simple-Protokoll (Extended erlaubt kein Multi-Statement), die
	// Verbindung selbst bleibt auf Extended → typisierte Parameter und
	// korrekte Typinferenz in den Runtime-Adaptern.
	bodyExecArgs []any
	// acquireMigrationLock serialisiert konkurrierende Migrations-Läufe
	// beim Multi-Replica-Boot (ADR-0006): booten N Replicas
	// gleichzeitig, laufen N Prozesse durch apply() und kollidieren sonst
	// auf Postgres' Katalog (`CREATE TABLE` ist nicht race-safe →
	// pg_type_typname_nsp_index-Dup, SQLSTATE 23505). Für Postgres ein
	// pg_advisory_lock (nur einer migriert, die anderen warten und sehen
	// dann pending=∅); für SQLite nil (single-writer, kein Cross-Prozess-
	// Race). Liefert eine release-Closure.
	acquireMigrationLock func(ctx context.Context, db *sql.DB) (func(), error)
}

// migrationLockKey ist der feste pg_advisory_lock-Schlüssel für den
// Migrations-Apply (bigint). Beliebig, aber projektstabil, damit alle
// Replicas denselben Lock nehmen.
const migrationLockKey int64 = 0x6D74726163650001

// sqliteDialect und postgresDialect liefern den jeweiligen Dialekt.
// Als Funktionen statt package-Globals gehalten (der Linter erlaubt
// keine nicht-exempten Globals), ohne Laufzeit-Kosten: die Dialekte
// werden nur beim Migrations-Apply (einmal beim Start) gebraucht.
func sqliteDialect() dialect {
	return dialect{
		placeholder:   func(int) string { return "?" },
		migrationsDir: "migrations",
	}
}

func postgresDialect() dialect {
	return dialect{
		placeholder:          func(n int) string { return "$" + strconv.Itoa(n) },
		migrationsDir:        "migrations/postgres",
		bodyExecArgs:         []any{pgx.QueryExecModeSimpleProtocol},
		acquireMigrationLock: acquirePostgresMigrationLock,
	}
}

// acquirePostgresMigrationLock nimmt einen Session-weiten pg_advisory_lock
// auf einer dedizierten Verbindung (die anderen Replicas blockieren hier,
// bis der Migrierende fertig ist). Der Lock MUSS auf derselben Connection
// gehalten und wieder freigegeben werden — daher ein gepinnter `*sql.Conn`
// statt des Pools. Die Migrations-Statements selbst laufen weiter über den
// Pool; der Lock ist reiner Mutex.
func acquirePostgresMigrationLock(ctx context.Context, db *sql.DB) (func(), error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage: migration-lock conn: %w", err)
	}
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", migrationLockKey); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("storage: pg_advisory_lock: %w", err)
	}
	return func() {
		_, _ = conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", migrationLockKey)
		_ = conn.Close()
	}, nil
}

// migrationNamePattern matcht das Flyway-File-Format
// `V<integer>__<description>.sql` (z. B. `V1__m_trace.sql`,
// `V2__add_trace_columns.sql`). Die Integer-Versionsnummer ist
// numerisch sortierbar; eine spätere Erweiterung auf semver-Versionen
// (z. B. `V1.2.3`) müsste Pattern + Sortier-Logik anpassen.
var migrationNamePattern = regexp.MustCompile(`^V(\d+)__.+\.sql$`)

// ErrSchemaDirty wird zurückgegeben, wenn schema_migrations einen
// Eintrag mit dirty=1 hat. Der API-Start refused dann; Reparatur ist
// in docs/user/local-development.md beschrieben.
var ErrSchemaDirty = errors.New("storage: schema is in dirty state")

// Open öffnet (oder erzeugt) die SQLite-Datei unter path, sichert
// Foreign-Keys und WAL-Mode, wendet alle offenen Migrationen aus dem
// eingebetteten migrations/-Verzeichnis an und gibt die fertige
// *sql.DB zurück.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	dsn := buildSQLiteDSN(path)
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: open %q: %w", path, err)
	}
	if err := setPragmas(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := applyDialect(ctx, db, sqliteDialect()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// OpenPostgres öffnet eine Postgres-Verbindung über dsn, wendet die
// eingebetteten Postgres-Migrationen (migrations/postgres/) an und gibt
// die fertige *sql.DB zurück (ADR-0006, optionaler Scale-out-Adapter).
//
// Die Verbindung nutzt das pgx-Default-**Extended**-Protokoll:
// typisierte Bind-Parameter → korrekte Typinferenz in den
// Runtime-Adaptern (das Simple-Protokoll interpolierte Parameter als
// Text-Literale und ließe PG bei mehrdeutigen `COALESCE($1,$2)`-Cursorn
// fälschlich `text` inferieren). Nur der Multi-Statement-Migrations-Body
// (V1 = 13 CREATE TABLE + Indizes in einem Exec) braucht das
// Simple-Protokoll — das trägt der postgresDialect() per bodyExecArgs
// gezielt an genau diesem Exec (applyOne), nicht an der ganzen
// Verbindung.
func OpenPostgres(ctx context.Context, dsn string) (*sql.DB, error) {
	connConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: parse postgres dsn: %w", err)
	}
	db := stdlib.OpenDB(*connConfig)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("storage: ping postgres: %w", err)
	}
	if err := requireCommitTimestampTracking(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := applyDialect(ctx, db, postgresDialect()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// requireCommitTimestampTracking erzwingt track_commit_timestamp=on beim
// Boot. Der Postgres-Adapter braucht pg_xact_commit_timestamp(xmin) für
// das R-27-Commit-Zeit-Wasserzeichen (ADR-0006) im Event-List-Pfad; ist
// das Setting aus, wirft die Funktion pro Request einen harten Fehler
// ("could not get commit timestamp data") statt NULL. Darum hier
// fail-loud beim Start (nicht als 500 im Read-Pfad) — dasselbe
// refuse-to-start-Prinzip wie bei dirty schema_migrations.
func requireCommitTimestampTracking(ctx context.Context, db *sql.DB) error {
	var setting string
	if err := db.QueryRowContext(ctx, "SHOW track_commit_timestamp").Scan(&setting); err != nil {
		return fmt.Errorf("storage: read track_commit_timestamp: %w", err)
	}
	if setting != "on" {
		return fmt.Errorf("storage: postgres requires track_commit_timestamp=on "+
			"(R-27 commit-time watermark for event pagination), got %q; set it in "+
			"postgresql.conf or start postgres with '-c track_commit_timestamp=on' and restart",
			setting)
	}
	return nil
}

// Apply wendet offene Migrationen aus files (fs.FS, gewurzelt am
// Verzeichnis mit den SQL-Files) auf die übergebene *sql.DB an, im
// SQLite-Dialekt. Test-freundlich: fstest.MapFS einsetzbar.
func Apply(ctx context.Context, db *sql.DB, files fs.FS) error {
	return apply(ctx, db, files, sqliteDialect())
}

// applyDialect wurzelt den embed.FS am dialekt-spezifischen
// Migrations-Unterordner und wendet die offenen Migrationen an.
func applyDialect(ctx context.Context, db *sql.DB, d dialect) error {
	sub, err := fs.Sub(embeddedMigrations, d.migrationsDir)
	if err != nil {
		return fmt.Errorf("storage: embed sub %q: %w", d.migrationsDir, err)
	}
	return apply(ctx, db, sub, d)
}

// buildSQLiteDSN setzt _txlock=immediate (BEGIN IMMEDIATE für Race-
// Schutz), _pragma=foreign_keys(ON) und WAL-Journal. Der Pfad bleibt
// clean (file:-URI nur bei Bedarf, hier Plain-Path).
func buildSQLiteDSN(path string) string {
	return path + "?_txlock=immediate" +
		"&_pragma=foreign_keys(ON)" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=busy_timeout(5000)"
}

// setPragmas verifiziert, dass die DSN-PRAGMAs gegriffen haben. Wir
// verlassen uns nicht blind auf den Driver — bei einer Driver-
// Konfigurationsabweichung würde die Test-Suite das sonst nicht sehen.
// Nur für den SQLite-Pfad relevant (Postgres kennt keine PRAGMAs).
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

func apply(ctx context.Context, db *sql.DB, files fs.FS, d dialect) error {
	// Multi-Replica-Boot: konkurrierende Migrations-Läufe serialisieren
	// (Postgres pg_advisory_lock; SQLite no-op). Der Lock umfasst
	// ensureSchemaMigrationsTable + alle applyOne, damit nur ein Replica
	// die CREATE-TABLE-DDL fährt; die anderen warten und sehen danach
	// pending=∅.
	if d.acquireMigrationLock != nil {
		release, err := d.acquireMigrationLock(ctx, db)
		if err != nil {
			return err
		}
		defer release()
	}
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
		if err := applyOne(ctx, db, m, d); err != nil {
			// dirty=1 separat persistieren (eigene Transaktion); ein
			// Fehler dabei darf den ursprünglichen Fehler nicht
			// verschlucken.
			if markErr := markDirty(ctx, db, m.version, d); markErr != nil {
				return fmt.Errorf("storage: apply %s: %w (failed to record dirty state: %v)",
					m.name, err, markErr)
			}
			return fmt.Errorf("storage: apply %s: %w", m.name, err)
		}
	}
	return nil
}

// schemaMigrationsDDL ist bewusst dialekt-portabel: `INTEGER PRIMARY
// KEY` (app-vergebene version), `TEXT`, `INTEGER NOT NULL DEFAULT 0`
// und `CREATE TABLE IF NOT EXISTS` sind auf SQLite wie auf Postgres
// gültig — daher braucht diese DDL keinen Dialekt-Zweig.
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
	defer func() { _ = rows.Close() }()
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

func applyOne(ctx context.Context, db *sql.DB, m migration, d dialect) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	if _, err := tx.ExecContext(ctx, m.body, d.bodyExecArgs...); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(ctx,
		fmt.Sprintf("INSERT INTO schema_migrations(version, applied_at, dirty) VALUES (%s, %s, 0)",
			d.placeholder(1), d.placeholder(2)),
		m.version, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func markDirty(ctx context.Context, db *sql.DB, version int64, d dialect) error {
	// ON CONFLICT ... DO UPDATE + excluded ist auf SQLite wie auf
	// Postgres gültig; nur die Bind-Platzhalter sind dialekt-spezifisch.
	_, err := db.ExecContext(ctx,
		fmt.Sprintf("INSERT INTO schema_migrations(version, applied_at, dirty) VALUES (%s, %s, 1) "+
			"ON CONFLICT(version) DO UPDATE SET dirty = 1, applied_at = excluded.applied_at",
			d.placeholder(1), d.placeholder(2)),
		version, time.Now().UTC().Format(time.RFC3339Nano))
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
