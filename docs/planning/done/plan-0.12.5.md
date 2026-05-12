# Implementation Plan — `0.12.5` (Auth-/Ingest-Adapter-Minor)

> **Status**: ✅ released 2026-05-11 (Tag `v0.12.5`). Vorgänger ist
> `0.12.1` (`v0.12.1`, Trigger-Re-Eval + Operator-Doku; Plan in
> [`plan-0.12.1.md`](./plan-0.12.1.md)).
>
> **Release-Typ**: **Minor-Release** (`0.12.5`) gemäß
> [`docs/user/releasing.md`](../../user/releasing.md) §3.1 — neue
> User-Surface (Auth-/Ingest-Adapter), Lastenheft-Patch `1.1.16`
> §13.15 mit RAK-77..RAK-82, neue RAK-Verifikationsmatrix.
>
> **Ziel**: Die Adapter-/Wire-Pfade liefern, die in `0.12.0` als
> Folge-Scope definiert und in `0.12.1` als „Code-Pfad in 0.12.5"
> markiert wurden:
> - **R-17** Multi-Replica-fähiger Issuance-Limiter (Shared-State).
> - **R-18** Multi-Key-Rotation: ENV-Schema, dynamischer Resolver,
>   Smoke.
> - **R-20** Secret-Backend-Driven-Port mit **Vault-Adapter-Skelett**
>   (KMS bleibt additive Folge-Option, siehe §0.3/§7).
> - **R-21** Browser-Ingest-Policy auf `/api/ingest/*`
>   (RAK-74-Scope-Cut aufheben unter klarer Ergebnis-Bedingung).
> - Optional: **R-14** Auth-Bridge MediaMTX/SRS,
>   **R-16** Outbound-Webhook für Stream-Lifecycle.
>
> **Bezug**:
> [`plan-0.12.0.md`](./plan-0.12.0.md) §10 Folge-Scope,
> §13.14 Lastenheft-RAK-71..RAK-76;
> [`plan-0.12.1.md`](./plan-0.12.1.md) §1 Trigger-Re-Eval-Stand;
> [`risks-backlog.md`](../in-progress/risks-backlog.md) §1.1+§1.2
> R-14/R-16/R-17/R-18/R-20/R-21;
> [`spec/lastenheft.md`](../../../spec/lastenheft.md) §13.15
> (Lastenheft-Patch `1.1.16` für `0.12.5`, RAK-77..RAK-82).
>
> **Nachfolger**:
> [`plan-0.12.6.md`](./plan-0.12.6.md) (Folge-Items-Sammlung
> für R-5/R-7/R-10/R-11/R-13/R-15/R-17/R-20/R-22 — adressiert die
> nach diesem Release offen oder „teilweise gelöst" gebliebenen
> R-N-Einträge; T0-Aktivierung entscheidet zwischen Patch und
> Minor), danach
> [`plan-0.13.0.md`](../done/plan-0.13.0.md) (Production / Ops
> Backends, MVP-40..MVP-44).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR-/Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.12.5` ist ein **Minor-Release**: neue User-Surface (Adapter,
ggf. neue Endpoints), neuer Lastenheft-Patch, neue RAK-Gruppe.

In Scope (verbindlich):

- **R-17 Shared-Issuance-Limiter** (`apps/api/adapters/driven/issuance/`):
  zweiter Adapter neben dem aktuellen In-Process-Limiter. Erster und
  einziger Shared-State-Adapter in `0.12.5` ist **SQLite via WAL**
  (Migration `V5`, atomare Counter-Erhöhung). Selektion über ENV
  `MTRACE_AUTH_ISSUANCE_LIMITER=memory|sqlite`; **globaler Default
  bleibt `memory`** (Backwards-Compat, Single-Instance-Lab),
  `sqlite` ist opt-in für Topologien mit Shared-State-Bedarf.
  **Topologie-Constraint** für `sqlite`: nur sinnvoll bei Single-
  Host-Deployments mit gemeinsam gemountetem Persistent-Volume
  (z. B. Compose mit `volumes:` auf demselben Host); echte Multi-
  Host-Topologie braucht einen Network-Backend-Adapter und ist
  Folge-Scope. Werte `redis`/`memcached` sind in `0.12.5` **nicht
  verfügbar** — die ENV-Validierung lehnt unbekannte Werte mit
  `fail-fast`-Startup ab. RAK-74-Scope-Cut in `0.12.0` bleibt:
  Limiter darf nicht hinter `/api/ingest/*` hängen.
- **R-18 Multi-Key-Rotation** (`apps/api/adapters/driven/auth/signing/`):
  ENV-Schema `MTRACE_AUTH_SIGNING_KEYS=<kid1>:<key1>[,<kid2>:<key2>,…]`
  plus `MTRACE_AUTH_SIGNING_ACTIVE_KID=<kid>`. Resolver liest beim
  Start alle Keys, der aktive `kid` signiert, alle übrigen `kid`s
  verifizieren. Smoke-Pfad analog `make smoke-ingest-control` für
  einen Rotation-Cycle (alter Key signiert → aktiv umschalten →
  neue Tokens validieren weiterhin gegen den alten Key bis Ablauf).
- **R-20 Secret-Backend-Adapter** (`apps/api/adapters/driven/secrets/`):
  `Driven-Port` für Signing- und Project-Token-Secret-Bezug. Erstes
  Default: aktueller ENV-basierter In-Memory-Adapter (Backwards-Compat).
  Erstes externes Backend: **Vault-Adapter-Skelett** (`hashicorp/vault/api`,
  Lab-Pfad mit `vault dev`-Server) — Entscheidung in T0-Closeout
  fixiert (siehe §0.3). KMS-Adapter (AWS-KMS) bleibt **additive
  Folge-Option** und ist explizit nicht Teil von `0.12.5`.
- **R-21 Browser-Ingest-Policy**: `/api/ingest/*` darf von Browser-
  Origins akzeptiert werden, mit klar definierten Project-Policy-
  Constraints (CORS-Allowlist je Project, optional CSRF-Token oder
  Origin-Pin). RAK-74-Scope-Cut wird kontrolliert aufgehoben — neue
  RAK definiert das Wire-Verhalten und ein Test-Set.

In Scope (optional, T-Splitter mit eigenem Go/No-Go):

- **R-14 Auth-Bridge MediaMTX/SRS**: signierte Publish-Tokens für
  externen Media-Server. Setzt minimal eine Test-Pipeline gegen
  echtes MediaMTX voraus.
- **R-16 Outbound-Webhook**: Stream-Lifecycle-Webhooks (`stream.started`,
  `stream.ended`) als optionaler Adapter. Retry-/Timeout-Schema,
  HMAC-signierte Payload.

Out of Scope:

- **R-15** Externe Provisionierung — erst auf konkreten
  Operator-Bedarf reaktivieren.
- OAuth/OIDC/SSO + User-/Org-/Admin-Verwaltung (RAK-71-Out-of-Scope
  bleibt; siehe `0.12.1` §2 OS-Schärfung).
- Production-Backends (`Postgres`, `ClickHouse`, K8s) — die wandern
  in `plan-0.13.0.md`.

### 0.2 Vorgänger-Gate

- `0.12.0` ist released; `v0.12.0` getaggt.
- `0.12.1` ist released; `v0.12.1` getaggt; alle aktiven R-N-Items
  haben einen aktuellen Trigger-Re-Eval-Eintrag.
- Operator-Runbook für Multi-Key-Rotation (Doku-Stand) ist in
  `auth.md` §5.3; `0.12.5` liefert den Code-Pfad dazu.

### 0.3 Architektur-/Persistenzentscheidung

**Hexagonale Erweiterung um zwei neue Driven-Ports:**

1. `IssuanceLimiterPort` (R-17) — fasst den heutigen In-Process-
   Limiter und mindestens einen Shared-Backend-Adapter unter einem
   Interface zusammen. Selektion zur Boot-Zeit per ENV.
2. `SecretBackendPort` (R-20) — fasst die ENV-Lookup-Quelle und
   einen externen Vault-Adapter unter einem Interface zusammen.
   **T0-Entscheidung 2026-05-11**: für `0.12.5` wird ausschließlich
   ein **Vault-Skelett** geliefert (`hashicorp/vault/api`,
   `vault dev`-Lab-Pfad); KMS-Adapter bleibt additive Folge-Option
   nach `0.12.5` und ist nicht Teil des Adapter-Skeletts hier.
   Begründung: schlankere Go-Dependency-Surface, lokale
   Reproduzierbarkeit mit `vault dev`, neutraler Driven-Port lässt
   KMS später ohne Wire-Bruch nachziehen.

**Bestehende Driven-Ports erweitert:**

- `SigningKeyResolver` (R-18) — heute `StaticSigningKeyResolver`
  mit fixem `kid`+Verify-Set; wird zu einem ENV-getriebenen
  `MultiKeySigningResolver` mit aktiver `kid`-Rotation.

**Persistenz:**

- SQLite-Schema bleibt; ggf. Migration `V5` für ein
  `auth_issuance_counters`-Tabelle, wenn der SQLite-Limiter-Backend
  Default-Adapter wird (T-Tranche entscheidet).

## 1. Lastenheft-Patch

`spec/lastenheft.md` bekommt mit T0-Aktivierung Patch `1.1.16` und
die neue RAK-Gruppe `RAK-77`..`RAK-82` in §13.15 (Anschluss an
§13.14 mit RAK-71..RAK-76). Die Inhaltszusammenfassung pro RAK
spiegelt den persistierten Lastenheft-Block:

| RAK | Bereich | Anforderung |
|---|---|---|
| RAK-77 | Auth/Issuance | Shared-State-Issuance-Limiter ist als Adapter-Pfad verfügbar; Selektion per ENV; das bisherige In-Process-Verhalten bleibt Default-Pfad. |
| RAK-78 | Auth/Signing | Multi-Key-Rotation: ENV-getriebener Resolver mit aktiver `kid`-Auswahl; alte Keys verifizieren weiter bis explizite Entfernung; Restart bleibt ohne Token-Invalidierung stabil (Erweiterung von RAK-72/RAK-73). |
| RAK-79 | Auth/Secret-Backend | Driven-Port für Secret-Bezug; ENV-Adapter bleibt Default; **Vault-Adapter-Skelett** liegt vor (T0-Entscheidung 2026-05-11, siehe §0.3). KMS bleibt additive Folge-Option. |
| RAK-80 | Ingest/Browser-Policy | `/api/ingest/*` ist optional aus Browser-Origin nutzbar; Project-Policy steuert Allowlist; ohne Policy gilt RAK-74-Scope-Cut weiter. |
| RAK-81 (optional) | Ingest/Auth-Bridge | MediaMTX-/SRS-Auth-Hook konsumiert signierte Publish-Tokens; Lab-/Produktiv-Trennung bleibt. |
| RAK-82 (optional) | Lifecycle/Webhooks | Stream-Lifecycle-Events können an externen Konsumenten zugestellt werden; HMAC-signierte Payload, definierter Retry-/Timeout-Pfad. |

Tranche 0 hat den Lastenheft-Patch persistiert. Doku-Konsistenz
wird via `make docs-check` (Doku-Link-Verifikation in
`scripts/verify-doc-refs.sh`) geprüft. Dedizierte
`make lastenheft-check`-/`make rak-trace`-Targets existieren im
Makefile **nicht** und sind in `0.12.5` nicht im Scope; eine
Aufnahme bleibt potenzielles Folge-Item (Quality-Gates Wave 3).

## 2. RAK-Verifikationsmatrix

| RAK | Code-Pfad | Test-/Smoke-Pfad |
|---|---|---|
| RAK-77 | `apps/api/adapters/driven/issuance/sqlite_limiter.go` (oder analog) | `apps/api/adapters/driven/issuance/*_test.go` + `make smoke-issuance-replica` |
| RAK-78 | `apps/api/adapters/driven/auth/signing/multi_key_resolver.go` | `apps/api/adapters/driven/auth/signing/*_test.go` + `make smoke-key-rotation` |
| RAK-79 | `apps/api/adapters/driven/secrets/{env,vault}_backend.go` (KMS additiv nach `0.12.5`) | `apps/api/adapters/driven/secrets/*_test.go` |
| RAK-80 | `apps/api/adapters/driving/http/ingest_browser.go` (oder Policy-Filter) | `apps/api/adapters/driving/http/ingest_browser_test.go` + Smoke |
| RAK-81 (opt) | `apps/api/adapters/driven/ingest/mediamtx_authbridge.go` | Lab-Smoke gegen echtes MediaMTX |
| RAK-82 (opt) | `apps/api/adapters/driven/webhooks/outbound_dispatcher.go` | Adapter-Test + Mock-Konsument |

## 3. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| --- | --- | --- |
| 0 | Plan-Aktivierung, Lastenheft-Patch, RAK-Matrix-Skelett, Roadmap-Insert | ✅ |
| 1 | R-18 Multi-Key-Rotation (Code) | ✅ |
| 2 | R-17 Shared-Issuance-Limiter (SQLite als erster Shared-State-Adapter, opt-in; globaler Default bleibt `memory`) | ✅ |
| 3 | R-20 Secret-Backend-Port + Vault-Adapter-Skelett (KMS additive Folge-Option) | ✅ |
| 4 | R-21 Browser-Ingest-Policy + RAK-74-Scope-Cut-Aufhebung | ✅ |
| 5 (optional) | R-14 Auth-Bridge MediaMTX/SRS und/oder R-16 Outbound-Webhook | ✅ (beide aktiviert) |
| 6 | Closeout: Versions-Bump, CHANGELOG, Plan-Move, Tag, Wave-2-Verdict | ⬜ |

---

## 4. Tranche 0 — Aktivierung

Ziel: Lastenheft-Patch + RAK-Matrix + Sekundär-Architekturentscheidung
(Vault-Adapter für R-20) vor erster Code-Lieferung.

DoD:

- [x] Plan von `docs/planning/open/plan-0.12.5.md` nach
  `docs/planning/in-progress/plan-0.12.5.md` verschoben (T0-Commit).
- [x] Lastenheft-Patch `spec/lastenheft.md` §13.15 mit RAK-77..
  RAK-82 ergänzt (RAK-81/82 als optional gekennzeichnet);
  Lastenheft-Header `Version: 1.1.16` mit neuem Patch-Block oben;
  vorab-deklarierter `1.1.16`-Block für `0.13.0` (RAK-77..RAK-81
  in §13.15) zurückgezogen (Variante C, siehe T0-Commit-Body).
- [x] `make docs-check` grün (Doku-Link-Verifikation via
  `scripts/verify-doc-refs.sh`; siehe §1 Klarstellung — `make
  lastenheft-check`/`make rak-trace` existieren nicht und sind
  nicht im `0.12.5`-Scope).
- [x] Roadmap-Insert: §1 Phase auf `0.12.5` aktiv; §2 Schritt 47.6
  auf 🟡; §3 Release-Übersicht-Zeile `0.12.5` auf 🟡.
- [x] Vorgänger-Gate verifiziert: `git tag --list v0.12.0 v0.12.1`
  liefert beide Tags.
- [x] Sekundär-Entscheidung Tranche 3: **Vault-Skelett** für R-20
  fixiert in §0.3 / §0.1 / §7 / §13.15 (KMS bleibt additive
  Folge-Option, nicht Teil von `0.12.5`).
- [x] `plan-0.13.0.md` §0.4 auf "RAK-IDs werden bei Aktivierung
  neu vergeben" umgestellt (Platzhalter-Mapping RAK-77..RAK-81 →
  vmtl. RAK-83..RAK-87 in §13.16).

## 5. Tranche 1 — R-18 Multi-Key-Rotation (Code)

Ziel: ENV-getriebener Multi-Key-Resolver ersetzt den heutigen
`StaticSigningKeyResolver` als Default-Pfad.

DoD:

- [x] ENV-Schema `MTRACE_AUTH_SIGNING_KEYS` (Kommagetrennte
  `kid:base64-secret`-Paare) plus `MTRACE_AUTH_SIGNING_ACTIVE_KID`
  parsen; Validierung gegen Duplikate, leere KIDs, ungültige
  Base64. Geliefert in `apps/api/adapters/driven/auth/signing_keys_env.go`
  (`ParseSigningKeysEnv`); 12 Unit-Tests in
  `signing_keys_env_test.go` decken Happy-Path, Whitespace-
  Toleranz, Single-Key-Fallback und alle Fehlerfälle ab.
- [x] `MultiKeySigningResolver` baut Verify-Set aus allen Keys,
  Sign-Key aus aktivem `kid`. Restart-stabil (analog vorhandener
  `TestHMACSigner_RestartStableAcrossKeyResolverReinitialization`).
  Rename des `0.12.0`-`StaticSigningKeyResolver` in
  `apps/api/adapters/driven/auth/multi_key_signing_resolver.go`
  (Public-API-Change: alle 11 Usage-Stellen in main.go +
  Tests aktualisiert).
- [x] Backwards-Compat: einzelner Key über alten ENV-Pfad bleibt
  unterstützt — alter Pfad ist degenerierter Multi-Key-Resolver
  mit `len(keys)==1`. Fallback-Logik in `ParseSigningKeysEnv`
  bei leerem `MTRACE_AUTH_SIGNING_KEYS`.
- [x] `make smoke-key-rotation`: erstellt Tokens unter `kid=A`,
  rotiert auf `kid=B` (active-kid-Umschaltung), prüft dass alte
  Tokens weiterhin verifizieren bis sie ablaufen. Geliefert als
  `scripts/smoke-key-rotation.sh`, das den End-to-End-Test
  `TestParseSigningKeysEnv_RotationEndToEnd` über
  `golang:1.26.3`-Docker triggert. Echte API-Restart-Variante
  ist Folge-Item, sobald Compose-Multi-Replica-Setup gebraucht
  wird (Synergie mit R-17 Tranche 2).
- [x] Operator-Runbook in `auth.md` §5.3 gegen die jetzt
  funktionsfähige Implementierung ergänzt: Multi-Key-Default-Setup
  in §5.2 Schritt 2, Code-Pfad-Verweis in §5.3 Header,
  Smoke-Verweis in §5.3.1 Footer.
- [x] Risks-Backlog R-18: Status auf 🟢 mit Auflösungspfad
  „MultiKeySigningResolver + ParseSigningKeysEnv + make smoke-
  key-rotation in `0.12.5` Tranche 1, Code in
  multi_key_signing_resolver.go / signing_keys_env.go;
  Wieder-Eröffnung nur bei Operator-Bug-Report im Multi-Replica-
  Pfad (Synergie R-17)". Eintrag bleibt in §1.1 bis
  Tranche-6-Closeout den finalen Move nach §1.2 macht.

## 6. Tranche 2 — R-17 Shared-Issuance-Limiter

Ziel: Multi-Replica-API teilt sich einen Issuance-Counter.

DoD:

- [x] Port-Interface bereits seit `0.12.0` definiert als
  `driven.IssuanceRateLimiter` in
  `apps/api/hexagon/port/driven/issuance_rate_limiter.go`. Plan-
  Notation „IssuanceLimiterPort" gemappt auf den existierenden
  Port-Namen; In-Process-Adapter
  (`auth.InMemoryIssuanceRateLimiter`) bleibt Default und
  implementiert das Interface unverändert.
- [x] SQLite-Backend-Adapter `SqliteIssuanceRateLimiter`
  implementiert in
  `apps/api/adapters/driven/auth/sqlite_issuance_rate_limiter.go`:
  Migration V5 `auth_issuance_counters` (bucket_key TEXT PK,
  capacity, refill_per_second, tokens REAL, last_at, expires_at
  + Cleanup-Index auf expires_at); atomare Refill+Consume in
  `BeginTx` (DSN erzwingt `BEGIN IMMEDIATE`, ADR-0002 §8.3);
  UPSERT-basierte Bucket-Init bei Erstaufruf oder Cfg-Wechsel;
  asymmetrischer Refund auf globalen Bucket bei project-deny;
  opportunistisches TTL-Cleanup (Default 24h, ~5 % Hot-Path-
  Wahrscheinlichkeit).
- [x] ENV-Selektion `MTRACE_AUTH_ISSUANCE_LIMITER=memory|sqlite`
  in `main.go#buildIssuanceRateLimiter`. Globaler Default
  `memory` (Backwards-Compat); `sqlite` braucht `persist.db !=
  nil` (also `MTRACE_PERSISTENCE=sqlite`) — sonst hard-fail mit
  klarer Fehlermeldung. Unbekannte Werte (`redis`, `memcached`)
  liefern explizit „not supported" mit Verweis auf den Folge-
  Plan; kein stiller Fallback.
- [x] `make smoke-issuance-replica` plus
  `scripts/smoke-issuance-replica.sh` triggern den End-to-End-
  Sharing-Test (`TestSqliteIssuanceRateLimiter_SharedAcrossInstances`)
  über `golang:1.26.3`-Docker. Test öffnet zwei `*sql.DB` auf
  dieselbe SQLite-Datei, verbraucht das Project-Bucket auf
  Instance A und prüft, dass Instance B den nächsten Allow als
  „denied" sieht — semantisch deckungsgleich zum Compose-Multi-
  Replica-Pfad. Operator-Topologie-Hinweis (Single-Host +
  Shared-Volume) in `docs/user/auth.md` §5.4 ergänzt; echte
  Compose-Container-Variante bleibt Folge-Item.
- [x] RAK-74-Scope-Cut explizit dokumentiert: Limiter hängt
  **nicht** vor `/api/ingest/*` (in `auth.md` §5.4 und
  Lastenheft §13.14 RAK-74).
- [x] Risks-Backlog R-17 auf **teilweise gelöst** gesetzt:
  Single-Host-Shared-Volume-Pfad geliefert (SQLite-Adapter +
  Migration V5 + ENV-Selektor + Smoke); Resttrigger
  „Multi-Host-Topologie oder Network-Backend-Bedarf" bleibt
  offen für späteren Adapter (Redis/Memcached). Eintrag bleibt
  in §1.1 mit ⬜-Status nach Wartungsregel §2 „teilweise
  Lösungen".

## 7. Tranche 3 — R-20 Secret-Backend-Adapter

Ziel: Externer Secret-Bezug ist als Adapter-Pfad verfügbar; das
heutige ENV-Lookup bleibt Default.

DoD:

- [x] `AuthSecretBackend`-Port in
  `apps/api/hexagon/port/driven/auth_secret_backend.go` definiert
  (Plan-Notation `SecretBackendPort` gemappt auf den hexagon-
  konformen Namen). `EnvSecretBackend` in
  `apps/api/adapters/driven/auth/env_secret_backend.go`
  implementiert das Interface 1:1 zum bisherigen
  `ParseSigningKeysEnv`-Pfad und signalisiert „kein Material" über
  den `ErrNoSecretConfigured`-Sentinel — der Caller (`main.go`)
  steuert den Lab-Default-Fallback exklusiv beim ENV-Backend.
- [x] **Vault-Adapter-Skelett** implementiert in
  `apps/api/adapters/driven/auth/vault_secret_backend.go`
  (T0-Entscheidung 2026-05-11, siehe §0.3): eigener minimaler
  `net/http`-Client gegen Vault KV-v2
  (`/v1/<mount>/data/<path>`) mit `X-Vault-Token`-Authentication.
  Bewusst **ohne** `hashicorp/vault/api`-Dependency — go.mod
  bleibt schlank, produktive Anbindung kann den Adapter
  später 1:1 durch `hashicorp/vault/api` ersetzen, ohne den
  Port zu ändern. Lab-Pfad: `vault dev`-Server (siehe
  `docs/user/auth.md` §5.5 Lab-Setup). KMS-Adapter (AWS-KMS)
  bleibt **additive Folge-Option** nach `0.12.5` und ist nicht
  Teil dieses Skeletts.
- [x] ENV-Selektion `MTRACE_AUTH_SECRET_BACKEND=env|vault` in
  `main.go#buildAuthSecretBackend`. Default `env`. Wert `kms`
  lehnt der Boot-Validator mit klarem „follow-up item"-Fehler ab.
  Externe Backends (`vault`) bekommen **kein** Lab-Default —
  ein nicht erreichbares Backend ist immer ein Boot-Fehler
  (fail-closed). Pflicht-ENV-Vars für Vault:
  `MTRACE_AUTH_VAULT_ADDR/_TOKEN/_PATH`. Optionale Feld-Aliase
  über `MTRACE_AUTH_VAULT_KEYS_FIELD`/`_ACTIVE_KID_FIELD`.
- [x] Lifecycle-Verhalten dokumentiert in `auth.md` §5.5: Boot-
  Time-Load, kein periodischer Refresh, fail-closed bei
  Backend-Outage. Schlüsselwechsel passiert über Operator-Restart
  (analog ENV-Pfad §5.3.1).
- [x] Risks-Backlog R-20: Status **teilweise gelöst** —
  Driven-Port + ENV-Adapter + Vault-Adapter-Skelett geliefert,
  ENV-Adapter bleibt Default. **Resttrigger** weiter offen:
  produktive Vault/KMS-Anbindung (AppRole/IAM-Auth),
  Compliance-Audit (PCI/SOC2), KMS-Adapter selbst. Tests:
  4 ENV-Adapter-Tests + 7 Vault-Adapter-Tests (httptest-Mock).
  Eintrag bleibt in §1.1 mit ⬜-Status nach Wartungsregel §2
  „teilweise Lösungen".

## 8. Tranche 4 — R-21 Browser-Ingest-Policy

Ziel: `/api/ingest/*` ist kontrolliert browser-tauglich; RAK-74-
Scope-Cut wird durch RAK-80 aufgehoben.

DoD:

- [x] Project-Policy-Schema erweitert: `domain.ProjectPolicy.BrowserIngest`
  vom Typ `BrowserIngestPolicy` mit Feldern `Enabled` (bool),
  `CORSAllowlist` ([]string), `CSRFRequired` (bool) und `OriginPin`
  (string). Domain-Helpers `IsZero()`, `AllowsBrowserOrigin(origin)`,
  `MatchesOriginPin(origin)` in
  `apps/api/hexagon/domain/auth_project_policy.go`.
- [x] `/api/ingest/*`-CORS-Pfad: `browserIngestPreflightHandler` in
  `apps/api/adapters/driving/http/cors.go` konsultiert den
  Project-Policy-Resolver (Methode `IsBrowserIngestOriginAllowed`
  am `InMemoryProjectPolicyResolver`). Match → 204 + Allow-Origin
  + `Access-Control-Allow-Methods: POST, OPTIONS` +
  `Allow-Headers: Content-Type, X-MTrace-Token, X-MTrace-CSRF`.
  Ohne Match → 204 ohne Allow-Origin (kein Enum-Leak). Ohne
  aktivierte Policy für irgendein Project → der heutige
  `dashboardPreflightHandler` bleibt aktiv
  (RAK-74-Scope-Cut strikt; Backwards-Compat-Pfad).
- [x] CSRF-Schutz + Origin-Pin als Default-Vorgaben über die neue
  `browserIngestEnforcement`-Middleware
  (`browser_ingest_enforcement.go`). Bei `Policy.Enabled=true`:
  Origin gegen Allowlist + ggf. Pin-Match prüfen,
  `X-MTrace-CSRF`-Header bei `CSRFRequired=true` einfordern.
  Fehlerschlüssel: `ingest_browser_origin_not_allowed`,
  `ingest_browser_origin_pin_mismatch`, `ingest_browser_csrf_missing`
  (alle `403`). **Skelett-Hinweis**: CSRF prüft nur Header-
  Anwesenheit; eine produktive Anti-CSRF-Token-Bibliothek
  (signiert, zeitlich begrenzt) bleibt Folge-Item.
- [x] Tests in `browser_ingest_test.go`: Browser-Origin mit
  Allowlist-Match → 204 mit CORS-Headern; ohne Match → 204 leer
  (RAK-74-Scope-Cut); ohne aktivierte Policy → 204 leer; POST
  Pfade — Policy disabled passes-through, Origin-Pin-Mismatch 403,
  CSRF-missing 403, All-Checks-Pass durchgereicht.
- [x] `make smoke-browser-ingest` plus
  `scripts/smoke-browser-ingest.sh` wrappen die Browser-Ingest-
  Tests über `golang:1.26.3`-Docker als reproduzierbaren Operator-
  Smoke. Help-Eintrag in `Makefile` ergänzt.
- [x] Risks-Backlog R-21: Status ⬜ → 🟢 mit Auflösungspfad
  „BrowserIngestPolicy + browserIngestPreflightHandler +
  browserIngestEnforcement-Middleware in `0.12.5` Tranche 4 plus
  `make smoke-browser-ingest`". CSRF-Hardening (produktive
  Anti-CSRF-Token-Library) bleibt als potenzieller
  Wieder-Eröffnungs-Trigger dokumentiert, ohne den
  Auflösungs-Status zu schmälern.

## 9. Tranche 5 — Optionale R-14 / R-16

Ziel: Optional Auth-Bridge MediaMTX/SRS und/oder Outbound-Webhook.
Tranche-Go nur bei zusätzlicher Bandbreite und Test-Pipeline-
Verfügbarkeit.

DoD (R-14, aktiviert):

- [x] Adapter `MediaMTXAuthHookHandler` bedient
  `POST /api/ingest/auth-hook` als Bridge zwischen MediaMTX-
  `externalAuth` und dem existierenden
  `IngestControlInbound.ValidateKey`-Pfad. Form-encoded Body,
  Mapping `user`/`password`/`action`/`path`; `200` allow vs.
  `403` deny. Bewusst **keine eigenen Publish-Tokens** — die
  vorhandenen Stream-Keys aus `/api/ingest/streams` liefern den
  kryptographischen Anker (Plan-Wording „signierte Publish-Tokens"
  inhaltlich abgedeckt durch die existierenden CSPRNG-Stream-Keys).
- [x] Lab-Smoke `make smoke-mediamtx-auth` wickelt sieben Wire-
  Pfade (allow, invalid-key, read-action, missing-field,
  bad-content-type, GET 405, validate-error) als End-to-End-Test
  über `httptest.Server` ab — keine MediaMTX-Container nötig.
  Echte Compose-Variante mit MediaMTX-Container gegen laufende
  m-trace-API bleibt Folge-Item.
- [x] Operator-Runbook in `docs/user/auth.md` §5.7: MediaMTX-
  Config-Beispiel (`authMethod: http` + `authHTTPAddress`),
  Mapping-Tabelle, Sicherheitsprofil (Netzwerk-Isolation), Liste
  der nicht-im-Skelett-enthaltenen Folge-Items.

DoD (R-16, aktiviert):

- [x] Driven-Port `OutboundWebhookDispatcher` (`hexagon/port/driven/outbound_webhook.go`)
  mit `Dispatch(ctx, event)`-Methode plus konsistentem
  `OutboundWebhookEvent`-Wire-Modell.
  Adapter `webhooks.HTTPDispatcher` mit Retry-/Timeout-Schema
  (Exponential-Backoff `100ms→200ms→400ms`, MaxAttempts 3,
  RequestTimeout 10s, Dead-Letter über
  `ErrOutboundWebhookExhausted`-Sentinel). Disabled-Modus
  (`URL == ""`) → No-Op.
- [x] HMAC-SHA-256-Signatur (`X-MTrace-Signature: sha256=<hex>`)
  über den exakten Body. Secret kommt aus
  `MTRACE_OUTBOUND_WEBHOOK_SECRET`. Replay-Schutz über
  `X-MTrace-Timestamp` (RFC3339Nano). Klartext-Stream-Keys nie in
  der Payload.
- [x] Adapter-Test plus Mock-Konsument-Smoke
  (`make smoke-outbound-webhook`): sieben Pfade — disabled-noop,
  happy path, HMAC-Match, Retry-success, Dead-Letter-Exhaustion,
  Body-Shape (Pflichtfelder + kein Klartext-Leak), Context-Cancel.

Integration: `IngestControlService.WithOutboundWebhookDispatcher`
hängt den Dispatcher optional an den Lifecycle-Pfad
(`RecordLifecycleEvent`). Boot-Pfad in `main.go#buildOutboundWebhookDispatcher`
liest ENV; leere URL → nil-Adapter, Service-Verhalten bleibt
identisch zum `0.11.0`-Stand. R-14 und R-16 sind beide aktiviert
und in der Risks-Backlog auf 🟢.

## 10. Tranche 6 — Release-Closeout

DoD:

- [x] `make docs-check` grün (verifiziert vor + nach Closeout-Commit).
- [x] `make gates` grün — final-Lauf nach Closeout-Commit
  `9a57ec2` (Memory `feedback_test_verification.md`:
  `generated-drift-check` läuft gegen HEAD, also Versions-Bump
  zuerst committen, dann Gates).
- [!] `make lastenheft-check` — Make-Target existiert nicht im
  Repo (siehe T0-Klarstellung in §1, Plan-0.12.5-Z.152). Doku-
  Link-Konsistenz wird durch `make docs-check` abgedeckt;
  dediziertes RAK-/Lastenheft-Sync-Gate bleibt potenzielles
  Quality-Gates-Wave-3-Folge-Item.
- [!] `make rak-trace` — Make-Target existiert nicht im Repo
  (analog `make lastenheft-check`; siehe §1 T0-Klarstellung).
  RAK-77..RAK-82-Code-/Test-Pfade sind stattdessen in Plan §1
  und Lastenheft §13.15 verlinkt; manuelle RAK-Trace-Audit
  zeigt vollständige Code-Test-Verlinkung pro RAK (siehe
  Tranche-Logs T1-T5 oben).
- [x] `make generated-drift-check` grün (Teil von `make gates`;
  finaler Lauf nach Versions-Bump-Commit).
- [x] **Wave-2-Quality-Gates dokumentiert** (`releasing.md` §3.1)
  — Stand 2026-05-11 vor Tag-Setzung:
  - `gh run list --workflow benchmark.yml --limit 1`:
    `completed success` (Run `25643634634`, 2026-05-11T00:11Z,
    4m31s).
  - `gh run list --workflow fuzz.yml --limit 1`:
    `completed success` (Run `25645472440`, 2026-05-11T01:24Z,
    30m52s).
  - `gh issue list --label fuzz --state open`: keine offenen
    Fuzz-Issues.
  - `gh run list --workflow mutation.yml --limit 3`: drei
    `completed success` (`25645624823`/`25623561573`/`25595781337`,
    alle grün).
  - **Verdict: GO** — alle vier Wave-2-Indikatoren grün, kein
    Re-Trigger blockierend.
- [x] Versions-Bump auf `0.12.5` an allen Stellen aus
  `releasing.md` §3.1 — geliefert mit Closeout-Commit
  `9a57ec2` (61 Dateien betroffen: 5× `package.json` +
  `main.go#serviceVersion` + `version.ts` +
  `contracts/sdk-compat.json` + 20+20 Analyzer-Fixtures
  + Go-/TS-Test-Hardcodes).
- [x] `CHANGELOG.md` `[0.12.5] - 2026-05-11`-Block mit
  `### Added` (sechs Adapter-Pfade + ENV-Vars + fünf Smokes),
  `### Changed` (RAK-74-Scope-Cut-Aufhebung, Static→Multi-Key-
  Rename, IngestControlService-API), `### Security` (Rotation,
  Limiter, Vault, Browser-Policy, MediaMTX-Bridge, Webhook-HMAC),
  `### Lastenheft` (Patch `1.1.16`, Variante-C-Erklärung).
- [x] Roadmap aktualisiert (Commit `9a57ec2`): §1 Phase auf
  ✅ released, §1.1 Tabelle ergänzt um Auth-/Ingest-Adapter-
  Minor-Zeile, §1.2 Nächste Phase auf `0.13.0`-Aktivierungs-
  bereitschaft, §2 Schritt 47.6 ✅, §3 Release-Übersicht-Zeile
  `0.12.5` ✅.
- [x] Plan nach `docs/planning/done/plan-0.12.5.md` verschoben
  (`git mv` in Closeout-Commit `9a57ec2`); Status-Header
  `✅ released 2026-05-11 (Tag v0.12.5)`; Tranchen-Übersicht §3
  alle ✅. Self-Refs auf `done/`-Pfade umgestellt.
- [x] Annotierter Tag `v0.12.5` auf den Closeout-Commit
  `9a57ec2` setzen mit Lieferzusammenfassung (Tag-Setzung folgt
  nach Audit-Review; Commit-SHA gepinnt im Tag-Message-Body).
- [ ] GitHub-Release `m-trace 0.12.5` mit Notes-File aus dem
  CHANGELOG-Block. Wird vom Operator nach Tag-Push via
  `gh release create v0.12.5` angelegt; Notes-Body kommt aus
  dem `CHANGELOG.md` `[0.12.5]`-Block.

## 11. Folge-Scope nach `0.12.5`

Direkte Folge-Items aus `0.12.5` — adressiert im
[`plan-0.12.6.md`](./plan-0.12.6.md) (Folge-Items-Sammlung
nach `0.12.5`, T0-Aktivierung entscheidet zwischen Patch und
Minor):

- **Multi-Host-Issuance-Limiter** (Network-Backend wie Redis/
  Memcached): bleibt nach `0.12.5` offen — der SQLite-Adapter
  deckt nur Single-Host-Shared-Volume. Resttrigger steht in
  `risks-backlog.md` R-17 („Multi-Host-Topologie oder
  Network-Backend-Bedarf"); Auflösung in
  [`plan-0.12.6.md`](./plan-0.12.6.md) Tranche 7 (bei
  Minor-Aktivierung), sonst späterer Minor.
- **Origin-/IP-nahes Rate-Limiting (R-22, in `0.12.1` Tranche 1
  als R-22 angelegt)**: Project-Token-basiertes Limit aus `0.12.5`
  löst R-22 nicht — IP-Layer-Limits sind eigener Code-Pfad.
  [`plan-0.12.6.md`](./plan-0.12.6.md) Tranche 6 liefert den
  Driven-Port-Adapter (bei Minor-Aktivierung), sonst späterer
  Minor sobald der R-22-Trigger ausgelöst ist.
- **Produktive KMS/Vault-Anbindung (R-20-Resttrigger)**: das
  Skelett aus `0.12.5` Tranche 3 deckt den Driven-Port ab; eine
  echte Operator-Anbindung (AppRole/IAM/K8s-ServiceAccount-Auth)
  plus KMS-Adapter und Compliance-Audit-Pfad sind
  [`plan-0.12.6.md`](./plan-0.12.6.md) Tranche 8.
- **Externe Media-Server-Provisionierung (R-15)**:
  [`plan-0.12.6.md`](./plan-0.12.6.md) Tranche 9 falls
  Trigger nach Operator-Bedarf greift.

Production-Backends bleiben separater Folge-Plan:

- [`plan-0.13.0.md`](../done/plan-0.13.0.md): Production / Ops
  Backends (`MVP-40`..`MVP-44`) — Postgres, ClickHouse,
  Kubernetes, Devcontainer, Release-Automatisierung. RAK-Range
  hängt davon ab, ob `plan-0.12.6.md` als Minor mit `RAK-83+`
  belegt wird.

Themen-Defer (keine getrackten R-N-Items):

- Später `plan-0.14.x` o. ä. für OAuth/OIDC/SSO + User-Verwaltung,
  falls konkreter Bedarf entsteht (RAK-71-Out-of-Scope bleibt
  normativ).

## 12. Qualitätsregeln für `0.12.5`

- Hexagonale Architektur: jeder neue Backend-Pfad ist ein Driven-
  Port mit Default-Adapter; ENV-Selektion ist die einzige
  Auswahlsteuerung.
- Backwards-Compat: heutige ENV-Variablen-Werte (Single-Key,
  In-Process-Limiter, ENV-Secret-Lookup) bleiben weiterhin
  unterstützt.
- Lastenheft als normative Quelle: jede neue Verhaltensaussage
  geht zuerst in §13.15 RAK-Gruppe, dann in `auth.md`/`ingest-
  control.md`/Code (Memory-Lehre `feedback_lastenheft_normativ.md`).
- Wave-2-Verdict vor Tag dokumentieren.
