-- plan-0.6.0 §4 Tranche 3 Sub-3.3 (RAK-42/RAK-46): durable SRT-Health-
-- Samples. Spec-Anker:
--   spec/backend-api-contract.md §10.6 (Tabellenform, Dedupe-Regel,
--   Retention),
--   spec/telemetry-model.md §7 (Datenmodell, Enums, Cardinality-
--   Vertrag).
--
-- Dedupe-Regel: ein Sample ist eindeutig über
--   (project_id, stream_id, connection_id,
--    COALESCE(source_observed_at, source_sequence))
-- Adapter erzwingt die Regel mit einem Vorab-Lookup auf den Dedupe-
-- Index plus ON CONFLICT DO NOTHING — `collected_at` allein ist
-- kein stabiler Schlüssel.
--
-- Retention: in 0.6.0 unbegrenzt analog der bestehenden SQLite-Demo-
-- Daten-Politik; bounded Snapshot-Historie mit Reset-/Prune-Pfad ist
-- Folge-Scope (plan-0.6.0 §4.3 + spec §10.6).

CREATE TABLE "srt_health_samples" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "project_id" TEXT NOT NULL REFERENCES "projects"("project_id") ON DELETE RESTRICT,
    "stream_id" TEXT NOT NULL,
    "connection_id" TEXT NOT NULL,
    "source_observed_at" TEXT,
    "source_sequence" TEXT,
    "collected_at" TEXT NOT NULL,
    "ingested_at" TEXT NOT NULL,
    "rtt_ms" REAL NOT NULL,
    "packet_loss_total" INTEGER NOT NULL,
    "packet_loss_rate" REAL,
    "retransmissions_total" INTEGER NOT NULL,
    "available_bandwidth_bps" INTEGER NOT NULL,
    "throughput_bps" INTEGER,
    "required_bandwidth_bps" INTEGER,
    "sample_window_ms" INTEGER,
    "source_status" TEXT NOT NULL,
    "source_error_code" TEXT NOT NULL,
    "connection_state" TEXT NOT NULL,
    "health_state" TEXT NOT NULL,
    CONSTRAINT "chk_srt_health_samples_source_status"
        CHECK (source_status IN ('ok', 'unavailable', 'partial', 'stale', 'no_active_connection')),
    CONSTRAINT "chk_srt_health_samples_source_error_code"
        CHECK (source_error_code IN ('none', 'source_unavailable', 'no_active_connection', 'partial_sample', 'stale_sample', 'parse_error')),
    CONSTRAINT "chk_srt_health_samples_connection_state"
        CHECK (connection_state IN ('connected', 'no_active_connection', 'unknown')),
    CONSTRAINT "chk_srt_health_samples_health_state"
        CHECK (health_state IN ('healthy', 'degraded', 'critical', 'unknown'))
);

CREATE INDEX "idx_srt_health_samples_stream_ingested"
    ON "srt_health_samples" ("project_id", "stream_id", "ingested_at", "id");

CREATE INDEX "idx_srt_health_samples_dedupe"
    ON "srt_health_samples" ("project_id", "stream_id", "connection_id", "source_observed_at", "source_sequence");
