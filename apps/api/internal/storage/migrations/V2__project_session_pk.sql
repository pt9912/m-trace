-- Tranche 3 §4.2: stream_sessions PK von (session_id) auf
-- (project_id, session_id) heben, damit dieselbe session_id in zwei
-- Projekten als zwei getrennte Sessions geführt wird (Cross-Project-
-- Kollisionsschutz). SQLite kann den PRIMARY KEY einer bestehenden
-- Tabelle nicht in-place ändern; das offizielle Pattern ist
-- "12-step ALTER" (siehe https://www.sqlite.org/lang_altertable.html
-- §7), hier in der minimalen Variante: neue Tabelle anlegen, Daten
-- kopieren, alte droppen, umbenennen, Indizes neu anlegen.
--
-- Der Apply-Runner umhüllt diese Migration in einer Transaktion
-- (siehe migrate.go Header), DDL ist in SQLite transaktional. WAL
-- bleibt aktiv. PRAGMA foreign_keys kann nicht innerhalb einer
-- Transaktion verändert werden; stream_sessions referenziert
-- projects(project_id) ON DELETE RESTRICT — die FK-Definition wird
-- beim CREATE TABLE neu deklariert.

CREATE TABLE "stream_sessions_v2" (
    "session_id" TEXT NOT NULL,
    "project_id" TEXT NOT NULL REFERENCES "projects"("project_id") ON DELETE RESTRICT,
    "state" TEXT NOT NULL DEFAULT 'active',
    "started_at" TEXT NOT NULL,
    "last_seen_at" TEXT NOT NULL,
    "ended_at" TEXT,
    "event_count" INTEGER NOT NULL DEFAULT 0,
    "correlation_id" TEXT,
    CONSTRAINT "chk_stream_sessions_state" CHECK (state IN ('active', 'stalled', 'ended')),
    PRIMARY KEY ("project_id", "session_id")
);

INSERT INTO "stream_sessions_v2" (
    "session_id", "project_id", "state", "started_at", "last_seen_at",
    "ended_at", "event_count", "correlation_id"
)
SELECT
    "session_id", "project_id", "state", "started_at", "last_seen_at",
    "ended_at", "event_count", "correlation_id"
FROM "stream_sessions";

DROP TABLE "stream_sessions";

ALTER TABLE "stream_sessions_v2" RENAME TO "stream_sessions";

CREATE INDEX "idx_stream_sessions_project_started" ON "stream_sessions" ("project_id", "started_at", "session_id");

CREATE INDEX "idx_stream_sessions_state" ON "stream_sessions" ("state");
