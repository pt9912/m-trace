-- Tranche 4 §5 H1: durable Spalte `end_source` in `stream_sessions`,
-- damit Dashboard- und API-Konsumenten zwischen explizitem
-- Session-Ende (`session_ended`-Event vom Client) und Sweeper-Ende
-- (zeitbasierter Lifecycle-Übergang) unterscheiden können.
-- Spec-Anker: spec/backend-api-contract.md §3.7.1, plan-0.4.0.md §5.
--
-- Werte (CHECK-Constraint):
--   - 'client'  — explizites `session_ended`-Event vom SDK
--   - 'sweeper' — zeitbasiertes Ende durch SessionsSweeper
--   - NULL      — Session ist noch active/stalled, oder Legacy-Eintrag
--                 vor V4 (Backfill auf NULL ist gewollt; der Operator-
--                 Doku-Anker dokumentiert den Legacy-Fall in §3.7.1).
--
-- ALTER TABLE in SQLite kann eine Spalte nur ohne FOREIGN KEY und
-- ohne PRIMARY-KEY-Bezug additiv anhängen — beides ist hier gegeben,
-- daher reicht der einfache ADD COLUMN-Pfad ohne 12-Step-Pattern
-- (anders als V2).

ALTER TABLE "stream_sessions"
    ADD COLUMN "end_source" TEXT
    CHECK (end_source IS NULL OR end_source IN ('client', 'sweeper'));
