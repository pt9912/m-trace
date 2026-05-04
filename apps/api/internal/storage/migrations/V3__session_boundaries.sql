-- Tranche 3 §4.4 D2: durable Persistenz für `session_boundaries[]`
-- (plan-0.4.0 §4.4; spec/telemetry-model.md §1.4; API-Kontrakt §3.4).
-- Boundary-Records sind kein Event-Stream — sie haben kein
-- `event_name`, zählen nicht in `accepted` und ändern die Batch-
-- `schema_version` nicht. Read-Pfad (§3.7.1) liefert das Tripel
-- `(kind, adapter, reason)` als `network_signal_absent[]` im Session-
-- Block; Doppel-Tripel werden dort dedupliziert.
--
-- PK = `(project_id, session_id, kind, network_kind, adapter, reason)`
-- erlaubt Insert-Or-Refresh-Idempotenz (Mehrfach-Sends desselben
-- Tripels führen nur zur Aktualisierung der Timestamps); FK auf
-- `stream_sessions(project_id, session_id) ON DELETE CASCADE` sorgt
-- dafür, dass beim späteren Löschen einer Session keine verwaisten
-- Boundaries zurückbleiben (ADR-0002 §8.4 Reset-Pfad).

CREATE TABLE "stream_session_boundaries" (
    "project_id"         TEXT NOT NULL,
    "session_id"         TEXT NOT NULL,
    "kind"               TEXT NOT NULL,
    "network_kind"       TEXT NOT NULL,
    "adapter"            TEXT NOT NULL,
    "reason"             TEXT NOT NULL,
    "client_timestamp"   TEXT NOT NULL,
    "server_received_at" TEXT NOT NULL,
    CONSTRAINT "chk_stream_session_boundaries_kind"
        CHECK (kind IN ('network_signal_absent')),
    CONSTRAINT "chk_stream_session_boundaries_network_kind"
        CHECK (network_kind IN ('manifest', 'segment')),
    CONSTRAINT "chk_stream_session_boundaries_adapter"
        CHECK (adapter IN ('hls.js', 'native_hls', 'unknown')),
    PRIMARY KEY ("project_id", "session_id", "kind", "network_kind", "adapter", "reason"),
    FOREIGN KEY ("project_id", "session_id")
        REFERENCES "stream_sessions"("project_id", "session_id")
        ON DELETE CASCADE
);

CREATE INDEX "idx_stream_session_boundaries_session"
    ON "stream_session_boundaries" ("project_id", "session_id");
