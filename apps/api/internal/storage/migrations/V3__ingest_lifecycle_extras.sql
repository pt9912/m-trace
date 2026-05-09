-- V3: Lifecycle-Hook-Felder für Tranche 4 (`0.11.0` / RAK-69).
--
-- Die V2-Variante von `stream_lifecycle_events` wurde nur als
-- Vorbereitung für T2 angelegt; T4 verfeinert das Schema in drei
-- Punkten:
--   1. `event_id` wird ein opaker, server-generierter String mit
--      Prefix `evt_`, kein Auto-Inkrement-Integer mehr — der
--      HTTP-Hook-Adapter echo't ihn als Acknowledgement an den
--      Aufrufer (Plan §3.8).
--   2. `connection_id` und `reason` aus dem Hook-Body werden
--      persistiert (optional, dokumentarisch); der Längenlimit-
--      Schutz lebt im Adapter, hier nur als TEXT.
--   3. Die Source-Allowlist wird auf `local-smoke` und
--      `mediamtx-hook` gesetzt — der bisherige Wert `smoke` wird in
--      `0.11.0` nicht ausgeliefert (Plan §0.11.0 Tranche 4 nennt
--      `local-smoke` als kanonische Wire-Source).
--
-- Da die Tabelle in keinem Release ausgeliefert wurde (V2 lebt nur
-- in der `0.11.0`-In-Progress-Phase), darf sie hier neu angelegt
-- werden. Existierende Lab-Reihen gehen verloren — das ist das
-- erwartete Verhalten für den Lab-Workflow.

DROP INDEX IF EXISTS "idx_stream_lifecycle_events_stream";
DROP TABLE IF EXISTS "stream_lifecycle_events";

CREATE TABLE "stream_lifecycle_events" (
    "event_id" TEXT NOT NULL,
    "project_id" TEXT NOT NULL,
    "stream_id" TEXT NOT NULL,
    "kind" TEXT NOT NULL,
    "occurred_at" TEXT NOT NULL,
    "received_at" TEXT NOT NULL,
    "source" TEXT NOT NULL,
    "key_fingerprint" TEXT NOT NULL DEFAULT '',
    "connection_id" TEXT NOT NULL DEFAULT '',
    "reason" TEXT NOT NULL DEFAULT '',
    PRIMARY KEY ("event_id"),
    CONSTRAINT "chk_stream_lifecycle_events_kind" CHECK (kind IN ('stream_started', 'stream_ended')),
    CONSTRAINT "chk_stream_lifecycle_events_source" CHECK (source IN ('local-smoke', 'mediamtx-hook')),
    CONSTRAINT "fk_stream_lifecycle_events_stream" FOREIGN KEY ("project_id", "stream_id")
        REFERENCES "ingest_streams" ("project_id", "stream_id") ON DELETE CASCADE
);

CREATE INDEX "idx_stream_lifecycle_events_stream" ON "stream_lifecycle_events" ("project_id", "stream_id", "occurred_at");
