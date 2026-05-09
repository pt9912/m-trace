-- Hand-gepflegte Migration für plan-0.11.0 Tranche 2 (NF-13 / RAK-65..RAK-70).
-- d-migrate `schema-generate` berührt ausschließlich V1__m_trace.sql; V2+
-- sind hand-gepflegt laut apps/api/Makefile §schema-generate-Kommentar.
--
-- Sicherheitsprofil:
--   - `ingest_streams` und `stream_keys` sind project-scoped via
--     project_id + ON DELETE CASCADE.
--   - Klartext-Stream-Keys werden nicht persistiert. `stream_keys`
--     speichert nur den vollständigen SHA-256-Hex-Hash (Unique pro
--     Project), den redigierten Fingerprint und den Lifecycle-Stand
--     (active vs. deactivated_at). Validate-Endpoint nutzt den Hash
--     als verifier; rotierte Keys bleiben für Audit, sind aber
--     nicht mehr aktiv.
--   - Lifecycle-Events tragen höchstens den Fingerprint (`key_fingerprint`).

CREATE TABLE "ingest_streams" (
    "stream_id" TEXT NOT NULL,
    "project_id" TEXT NOT NULL REFERENCES "projects"("project_id") ON DELETE RESTRICT,
    "display_name" TEXT NOT NULL,
    "protocol" TEXT NOT NULL,
    "endpoint_id" TEXT NOT NULL,
    "target_id" TEXT NOT NULL,
    "routing_rule_id" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'ready',
    "created_at" TEXT NOT NULL,
    "updated_at" TEXT NOT NULL,
    CONSTRAINT "chk_ingest_streams_protocol" CHECK (protocol IN ('srt', 'rtmp')),
    CONSTRAINT "chk_ingest_streams_status" CHECK (status IN ('ready', 'live', 'ended', 'disabled')),
    PRIMARY KEY ("project_id", "stream_id")
);

CREATE TABLE "ingest_endpoints" (
    "endpoint_id" TEXT NOT NULL,
    "protocol" TEXT NOT NULL,
    "listen_host" TEXT NOT NULL,
    "listen_port" INTEGER NOT NULL,
    "path_template" TEXT NOT NULL,
    "lab_stack" TEXT NOT NULL,
    "public_url_hint" TEXT NOT NULL,
    CONSTRAINT "chk_ingest_endpoints_protocol" CHECK (protocol IN ('srt', 'rtmp')),
    PRIMARY KEY ("endpoint_id")
);

CREATE TABLE "media_server_targets" (
    "target_id" TEXT NOT NULL,
    "kind" TEXT NOT NULL,
    "config_path" TEXT NOT NULL,
    "hls_url_template" TEXT NOT NULL,
    "control_api_url" TEXT NOT NULL DEFAULT '',
    CONSTRAINT "chk_media_server_targets_kind" CHECK (kind IN ('mediamtx', 'srs')),
    PRIMARY KEY ("target_id")
);

CREATE TABLE "ingest_routing_rules" (
    "rule_id" TEXT NOT NULL,
    "project_id" TEXT NOT NULL,
    "stream_id" TEXT NOT NULL,
    "target_id" TEXT NOT NULL REFERENCES "media_server_targets"("target_id") ON DELETE RESTRICT,
    "mode" TEXT NOT NULL DEFAULT 'single',
    "enabled" INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT "chk_ingest_routing_rules_mode" CHECK (mode IN ('single')),
    CONSTRAINT "chk_ingest_routing_rules_enabled" CHECK (enabled IN (0, 1)),
    CONSTRAINT "fk_ingest_routing_rules_stream" FOREIGN KEY ("project_id", "stream_id")
        REFERENCES "ingest_streams" ("project_id", "stream_id") ON DELETE CASCADE,
    PRIMARY KEY ("rule_id")
);

CREATE TABLE "stream_keys" (
    "key_id" TEXT NOT NULL,
    "project_id" TEXT NOT NULL,
    "stream_id" TEXT NOT NULL,
    "key_hash" TEXT NOT NULL,
    "fingerprint" TEXT NOT NULL,
    "created_at" TEXT NOT NULL,
    "deactivated_at" TEXT,
    CONSTRAINT "fk_stream_keys_stream" FOREIGN KEY ("project_id", "stream_id")
        REFERENCES "ingest_streams" ("project_id", "stream_id") ON DELETE CASCADE,
    PRIMARY KEY ("key_id")
);

CREATE TABLE "stream_lifecycle_events" (
    "event_id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "project_id" TEXT NOT NULL,
    "stream_id" TEXT NOT NULL,
    "kind" TEXT NOT NULL,
    "occurred_at" TEXT NOT NULL,
    "received_at" TEXT NOT NULL,
    "source" TEXT NOT NULL,
    "key_fingerprint" TEXT NOT NULL DEFAULT '',
    CONSTRAINT "chk_stream_lifecycle_events_kind" CHECK (kind IN ('stream_started', 'stream_ended')),
    CONSTRAINT "chk_stream_lifecycle_events_source" CHECK (source IN ('smoke', 'mediamtx-hook')),
    CONSTRAINT "fk_stream_lifecycle_events_stream" FOREIGN KEY ("project_id", "stream_id")
        REFERENCES "ingest_streams" ("project_id", "stream_id") ON DELETE CASCADE
);

CREATE UNIQUE INDEX "idx_ingest_streams_active_display_name" ON "ingest_streams" ("project_id", "display_name")
    WHERE "status" != 'ended';

CREATE INDEX "idx_ingest_routing_rules_stream" ON "ingest_routing_rules" ("project_id", "stream_id");

CREATE UNIQUE INDEX "idx_stream_keys_active_unique" ON "stream_keys" ("project_id", "key_hash")
    WHERE "deactivated_at" IS NULL;

CREATE INDEX "idx_stream_keys_stream" ON "stream_keys" ("project_id", "stream_id", "deactivated_at");

CREATE INDEX "idx_stream_lifecycle_events_stream" ON "stream_lifecycle_events" ("project_id", "stream_id", "occurred_at");
