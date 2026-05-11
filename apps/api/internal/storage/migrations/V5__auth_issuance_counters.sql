-- V5: Shared-State-Issuance-Counter für plan-0.12.5 Tranche 2
-- (RAK-77). Hand-gepflegt analog V2/V3/V4; d-migrate `schema-generate`
-- berührt ausschließlich V1__m_trace.sql.
--
-- Zweck:
--   Multi-Replica-API-Setup (zwei API-Instances auf demselben Host
--   mit gemounteter SQLite-DB) teilt sich den Token-Bucket-Counter
--   für `POST /api/auth/session-tokens`. Ohne diesen Adapter
--   misst der In-Process-`InMemoryIssuanceRateLimiter` die Quote
--   pro Replica — die effektive globale Issuance-Rate wäre bis zu
--   N× höher als konfiguriert (R-17 im Risiken-Backlog).
--
-- Sicherheitsprofil:
--   - Persistente Sicht enthält ausschließlich Bucket-Identifier
--     (`global`, `project:<id>`), die wirksame Bucket-Konfiguration
--     (Capacity + RefillPerSecond) und den Zähler-State
--     (`tokens` als REAL, weil Token-Bucket-Refill float-genau ist).
--   - Keine Klartext-Secrets, keine PII, keine Cross-Project-IDs
--     außerhalb des `project:<id>`-Bucket-Keys.
--   - `expires_at` erlaubt opportunistisches TTL-Cleanup, damit
--     verwaiste Project-Buckets nach langem Stillstand nicht
--     unbegrenzt wachsen. Default-TTL ist Lab-konservativ (24h).
--
-- Topologie-Constraint (Operator-Doku `auth.md` §5.4):
--   - Sinnvoll nur bei Single-Host-Deployments mit Shared-Persistent-
--     Volume (Compose `volumes:` auf demselben Host, K8s `hostPath`).
--   - Echte Multi-Host-Topologie braucht einen Network-Backend-Adapter
--     (Redis/Memcached) als Folge-Item — bleibt nach `0.12.5` offen.
--
-- Migrations-/Rollback-Regeln:
--   - Rollback auf V4: DROP TABLE entfernt nur den Counter-State.
--     Ohne diese Tabelle fällt der `MTRACE_AUTH_ISSUANCE_LIMITER=
--     sqlite`-Pfad beim Boot mit klarer Fehlermeldung aus; die API
--     muss in dem Fall auf `MTRACE_AUTH_ISSUANCE_LIMITER=memory`
--     zurückgestellt werden.
--   - Counter-State ist nicht Audit-relevant — ein Rollback verliert
--     ausschließlich den aktuellen Bucket-Stand, was bei einer
--     frischen Token-Bucket-Implementierung mit Sekunden-genauem
--     Refill akzeptabel ist.

CREATE TABLE "auth_issuance_counters" (
    "bucket_key" TEXT NOT NULL PRIMARY KEY,
    "capacity" INTEGER NOT NULL,
    "refill_per_second" REAL NOT NULL,
    "tokens" REAL NOT NULL,
    "last_at" TEXT NOT NULL,
    "expires_at" TEXT NOT NULL
);

CREATE INDEX "idx_auth_issuance_counters_expires" ON "auth_issuance_counters" ("expires_at");
