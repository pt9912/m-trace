-- V4: Rotierbare Project-Token-Generationen für plan-0.12.0 Tranche 3
-- (RAK-73). Hand-gepflegt analog V2/V3; d-migrate `schema-generate`
-- berührt ausschließlich V1__m_trace.sql.
--
-- Sicherheitsprofil:
--   - Persistenz speichert ausschließlich SHA-256-Hex-Hash plus
--     redigierten Fingerprint plus Lifecycle-Metadaten — nie den
--     Klartext-Token.
--   - `key_hash` ist repository-weit unique, damit ein Token nicht
--     zwei Projects gleichzeitig gehören kann.
--   - Lifecycle-Felder `not_before`, `grace_until`, `expires_at`,
--     `revoked_at` und `created_at` sind RFC3339-Nano-Strings (UTC);
--     `grace_until` und `expires_at` sind nullable.
--   - `rotated_from` ist optional und referenziert die Vorgänger-
--     Generation nur dokumentarisch — die Grace-Entscheidung läuft
--     ausschließlich über `grace_until`, nicht über
--     `rotated_from`-Lookups.
--
-- Migrations-/Rollback-Regeln (RAK-73):
--   - Rollback auf eine vorherige Config darf keine bereits
--     widerrufene Generation reaktivieren — `revoked_at` bleibt im
--     Audit-Pfad sichtbar; ein „Rollback" muss eine neue Generation
--     anlegen.
--   - Eine in der Vorgänger-Schema-Version persistierte
--     `mtr_pt_*`-Generation bleibt nach diesem Migrate-Schritt
--     gültig; die Spalte `created_at` ist Pflicht und wird vom
--     Repository immer gesetzt.

CREATE TABLE "project_token_generations" (
    "token_id" TEXT NOT NULL,
    "project_id" TEXT NOT NULL REFERENCES "projects"("project_id") ON DELETE CASCADE,
    "key_hash" TEXT NOT NULL,
    "fingerprint" TEXT NOT NULL,
    "not_before" TEXT NOT NULL,
    "grace_until" TEXT,
    "expires_at" TEXT,
    "revoked_at" TEXT,
    "created_at" TEXT NOT NULL,
    "rotated_from" TEXT,
    PRIMARY KEY ("token_id"),
    UNIQUE ("key_hash")
);

CREATE INDEX "idx_project_token_generations_project" ON "project_token_generations" ("project_id", "created_at");
