# Implementation Plan — `0.12.5` (Auth-/Ingest-Adapter-Minor)

> **Status**: ⬜ open — noch nicht aktiviert. Dieser Plan darf erst
> nach explizitem Move nach `docs/planning/in-progress/` umgesetzt
> werden, und erst nachdem `plan-0.12.1.md` (Trigger-Re-Eval +
> Operator-Doku) released wurde.
>
> **Release-Typ**: **Minor-Release** (`0.12.5`) gemäß
> [`docs/user/releasing.md`](../../user/releasing.md) §3.1 — neue
> User-Surface (Auth-/Ingest-Adapter), neuer Lastenheft-Patch,
> neue RAK-Verifikationsmatrix.
>
> **Ziel**: Die Adapter-/Wire-Pfade liefern, die in `0.12.0` als
> Folge-Scope definiert und in `0.12.1` als „Code-Pfad in 0.12.5"
> markiert wurden:
> - **R-17** Multi-Replica-fähiger Issuance-Limiter (Shared-State).
> - **R-18** Multi-Key-Rotation: ENV-Schema, dynamischer Resolver,
>   Smoke.
> - **R-20** KMS-/Vault-/Cloud-Secret-Manager-Adapter.
> - **R-21** Browser-Ingest-Policy auf `/api/ingest/*`
>   (RAK-74-Scope-Cut aufheben unter klarer Ergebnis-Bedingung).
> - Optional: **R-14** Auth-Bridge MediaMTX/SRS,
>   **R-16** Outbound-Webhook für Stream-Lifecycle.
>
> **Bezug**:
> [`done/plan-0.12.0.md`](../done/plan-0.12.0.md) §10 Folge-Scope,
> §13.14 Lastenheft-RAK-71..RAK-76;
> [`open/plan-0.12.1.md`](./plan-0.12.1.md) §1 Trigger-Re-Eval-Stand;
> [`risks-backlog.md`](../in-progress/risks-backlog.md) §1.1 R-14,
> R-16, R-17, R-18, R-20, R-21;
> [`spec/lastenheft.md`](../../../spec/lastenheft.md) §13.14 (vor
> Aktivierung extendieren um RAK-77..).
>
> **Nachfolger**: [`plan-0.13.0.md`](./plan-0.13.0.md)
> (Production / Ops Backends, MVP-40..MVP-44).

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
  Default: aktueller ENV-basierter In-Memory-Adapter. Erstes externes
  Backend: KMS-Adapter-Skelett (AWS-KMS, Provider-API als Interface)
  oder Vault-Adapter — Auswahl in §1.3 entscheiden.
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
   externe KMS-/Vault-Adapter unter einem Interface zusammen.

**Bestehende Driven-Ports erweitert:**

- `SigningKeyResolver` (R-18) — heute `StaticSigningKeyResolver`
  mit fixem `kid`+Verify-Set; wird zu einem ENV-getriebenen
  `MultiKeySigningResolver` mit aktiver `kid`-Rotation.

**Persistenz:**

- SQLite-Schema bleibt; ggf. Migration `V5` für ein
  `auth_issuance_counters`-Tabelle, wenn der SQLite-Limiter-Backend
  Default-Adapter wird (T-Tranche entscheidet).

## 1. Lastenheft-Patch

`spec/lastenheft.md` bekommt einen Patch-Block — Versionsnummer
`1.1.16` (oder höher, je nach aktuellem Patch-Stand). Neue
RAK-Gruppe in §13.15 (Anschluss an §13.14 mit RAK-71..RAK-76):

| RAK | Bereich | Anforderung |
|---|---|---|
| RAK-77 | Auth/Issuance | Shared-State-Issuance-Limiter ist als Adapter-Pfad verfügbar; Selektion per ENV; das bisherige In-Process-Verhalten bleibt Default-Pfad. |
| RAK-78 | Auth/Signing | Multi-Key-Rotation: ENV-getriebener Resolver mit aktiver `kid`-Auswahl; alte Keys verifizieren weiter bis explizite Entfernung; Restart bleibt ohne Token-Invalidierung stabil (Erweiterung von RAK-72/RAK-73). |
| RAK-79 | Auth/Secret-Backend | Driven-Port für Secret-Bezug; ENV-Adapter bleibt Default; mindestens ein externes Backend-Adapter-Skelett (KMS oder Vault) liegt vor. |
| RAK-80 | Ingest/Browser-Policy | `/api/ingest/*` ist optional aus Browser-Origin nutzbar; Project-Policy steuert Allowlist; ohne Policy gilt RAK-74-Scope-Cut weiter. |
| RAK-81 (optional) | Ingest/Auth-Bridge | MediaMTX-/SRS-Auth-Hook konsumiert signierte Publish-Tokens; Lab-/Produktiv-Trennung bleibt. |
| RAK-82 (optional) | Lifecycle/Webhooks | Stream-Lifecycle-Events können an externen Konsumenten zugestellt werden; HMAC-signierte Payload, definierter Retry-/Timeout-Pfad. |

Tranche 0 erstellt den Lastenheft-Patch; jeder Tranche schließt mit
`make lastenheft-check` und `make rak-trace`.

## 2. RAK-Verifikationsmatrix

| RAK | Code-Pfad | Test-/Smoke-Pfad |
|---|---|---|
| RAK-77 | `apps/api/adapters/driven/issuance/sqlite_limiter.go` (oder analog) | `apps/api/adapters/driven/issuance/*_test.go` + `make smoke-issuance-replica` |
| RAK-78 | `apps/api/adapters/driven/auth/signing/multi_key_resolver.go` | `apps/api/adapters/driven/auth/signing/*_test.go` + `make smoke-key-rotation` |
| RAK-79 | `apps/api/adapters/driven/secrets/{env,kms,vault}_backend.go` | `apps/api/adapters/driven/secrets/*_test.go` |
| RAK-80 | `apps/api/adapters/driving/http/ingest_browser.go` (oder Policy-Filter) | `apps/api/adapters/driving/http/ingest_browser_test.go` + Smoke |
| RAK-81 (opt) | `apps/api/adapters/driven/ingest/mediamtx_authbridge.go` | Lab-Smoke gegen echtes MediaMTX |
| RAK-82 (opt) | `apps/api/adapters/driven/webhooks/outbound_dispatcher.go` | Adapter-Test + Mock-Konsument |

## 3. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| --- | --- | --- |
| 0 | Plan-Aktivierung, Lastenheft-Patch, RAK-Matrix-Skelett, Roadmap-Insert | ⬜ |
| 1 | R-18 Multi-Key-Rotation (Code) | ⬜ |
| 2 | R-17 Shared-Issuance-Limiter (SQLite als erster Shared-State-Adapter, opt-in; globaler Default bleibt `memory`) | ⬜ |
| 3 | R-20 Secret-Backend-Port + KMS-/Vault-Adapter-Skelett | ⬜ |
| 4 | R-21 Browser-Ingest-Policy + RAK-74-Scope-Cut-Aufhebung | ⬜ |
| 5 (optional) | R-14 Auth-Bridge MediaMTX/SRS und/oder R-16 Outbound-Webhook | ⬜ |
| 6 | Closeout: Versions-Bump, CHANGELOG, Plan-Move, Tag, Wave-2-Verdict | ⬜ |

---

## 4. Tranche 0 — Aktivierung

Ziel: Lastenheft-Patch + RAK-Matrix vor erster Code-Lieferung.

DoD:

- [ ] Plan von `docs/planning/open/plan-0.12.5.md` nach
  `docs/planning/in-progress/plan-0.12.5.md` verschoben.
- [ ] Lastenheft-Patch `spec/lastenheft.md` §13.15 mit RAK-77..
  RAK-82 ergänzt (RAK-81/82 als optional gekennzeichnet).
- [ ] `make lastenheft-check` und `make rak-trace` grün.
- [ ] Roadmap-Insert: §1 Phase auf `0.12.5` aktiv; §2 Schritt 47.6
  ergänzt; §3 Release-Übersicht-Zeile `0.12.5`.
- [ ] Vorgänger-Gate verifiziert: `git tag --list v0.12.0 v0.12.1`.

## 5. Tranche 1 — R-18 Multi-Key-Rotation (Code)

Ziel: ENV-getriebener Multi-Key-Resolver ersetzt den heutigen
`StaticSigningKeyResolver` als Default-Pfad.

DoD:

- [ ] ENV-Schema `MTRACE_AUTH_SIGNING_KEYS` (Kommagetrennte
  `kid:base64-secret`-Paare) plus `MTRACE_AUTH_SIGNING_ACTIVE_KID`
  parsen; Validierung gegen Duplikate, leere KIDs, ungültige
  Base64.
- [ ] `MultiKeySigningResolver` baut Verify-Set aus allen Keys,
  Sign-Key aus aktivem `kid`. Restart-stabil (analog vorhandener
  `TestHMACSigner_RestartStableAcrossKeyResolverReinitialization`).
- [ ] Backwards-Compat: einzelner Key über alten ENV-Pfad bleibt
  unterstützt — alter Pfad ist degenerierter Multi-Key-Resolver
  mit `len(keys)==1`.
- [ ] `make smoke-key-rotation`: erstellt Tokens unter `kid=A`,
  rotiert auf `kid=B` (active-kid-Umschaltung), prüft dass alte
  Tokens weiterhin verifizieren bis sie ablaufen.
- [ ] Operator-Runbook in `auth.md` §5.3 (aus `0.12.1` Doku-Stand)
  gegen die jetzt funktionsfähige Implementierung ergänzt: Beispiel-
  ENV, exakter Rotation-Befehl.
- [ ] Risks-Backlog R-18: Trigger als „aufgelöst durch Code in
  0.12.5" markieren, mit Commit-Hash und Migrations-Notiz.

## 6. Tranche 2 — R-17 Shared-Issuance-Limiter

Ziel: Multi-Replica-API teilt sich einen Issuance-Counter.

DoD:

- [ ] `IssuanceLimiterPort` definiert (Interface in
  `apps/api/core/ports/issuance.go` oder analog), heutiger
  In-Process-Limiter implementiert das Interface.
- [ ] SQLite-Backend-Adapter implementiert: `auth_issuance_counters`-
  Tabelle (Migration `V5`), atomare Counter-Erhöhung über UPSERT
  oder Lock, TTL-getriebenes Cleanup.
- [ ] ENV-Selektion: `MTRACE_AUTH_ISSUANCE_LIMITER=memory|sqlite`.
  **Globaler Default bleibt `memory`** (Backwards-Compat); `sqlite`
  ist opt-in. Andere Werte (`redis`, `memcached`, …) lehnt der
  Boot-Validator mit klarem Fehler ab — sie sind in `0.12.5` nicht
  implementiert.
- [ ] `make smoke-issuance-replica`: zwei API-Replicas auf
  demselben Host mit gemounteter SQLite-DB (Compose-Volume oder
  Bind-Mount) teilen den Counter; Limit über beide Replicas hinweg
  durchgesetzt. Operator-Doku zur Topologie-Voraussetzung
  (Single-Host + Shared-Volume) in `auth.md` §5.4 ergänzt.
- [ ] RAK-74-Scope-Cut: Limiter bleibt **nicht** vor
  `/api/ingest/*` (Doku in `auth.md` und Lastenheft).
- [ ] Risks-Backlog R-17: Status auf „**teilweise gelöst** —
  Code-Pfad für Single-Host-Shared-Volume verfügbar (SQLite-
  Adapter in 0.12.5)" setzen; Resttrigger „Multi-Host-Topologie
  oder Network-Backend-Bedarf" bleibt offen für späteren Adapter
  (Redis/Memcached als Folge-Item).

## 7. Tranche 3 — R-20 Secret-Backend-Adapter

Ziel: Externer Secret-Bezug ist als Adapter-Pfad verfügbar; das
heutige ENV-Lookup bleibt Default.

DoD:

- [ ] `SecretBackendPort` definiert; ENV-Adapter implementiert das
  Interface 1:1 zum heutigen Verhalten.
- [ ] Mindestens ein externes Backend-Skelett implementiert (Auswahl
  in T0): KMS-Adapter (AWS-KMS-Provider-API als Interface ohne
  AWS-SDK-Pflicht-Dependency, falls zu groß) oder Vault-Adapter.
  Lab-Test-Pfad mit Mock-Backend, kein Pflicht-Lab gegen echte
  KMS/Vault-Instanz.
- [ ] ENV-Selektion: `MTRACE_AUTH_SECRET_BACKEND=env|kms|vault`.
  Default `env`.
- [ ] Lifecycle-Verhalten dokumentiert: Caching, Refresh-TTL,
  Failure-Modus (fail-closed bei externem Backend-Outage).
- [ ] Risks-Backlog R-20: Status auf „**teilweise gelöst** —
  Driven-Port und Adapter-Skelett (KMS oder Vault) in 0.12.5
  verfügbar, ENV-Adapter bleibt Default" setzen. **Resttrigger
  bleibt offen**: produktive Backend-Anbindung (Operator
  konfiguriert tatsächliches KMS/Vault), Compliance-Audit
  (PCI/SOC2). Skelett-Lieferung allein erfüllt den Resttrigger
  nicht.

## 8. Tranche 4 — R-21 Browser-Ingest-Policy

Ziel: `/api/ingest/*` ist kontrolliert browser-tauglich; RAK-74-
Scope-Cut wird durch RAK-80 aufgehoben.

DoD:

- [ ] Project-Policy-Schema erweitert: `browser_ingest_policy.{enabled,
  cors_allowlist[],csrf_required,origin_pin}` (oder ähnlich,
  finalisiert beim Schema-Diff).
- [ ] `/api/ingest/*`-CORS-Pfad: Mit aktivem Project-Policy-Eintrag
  Origin gegen Allowlist prüfen; ohne Eintrag gilt RAK-74-Scope-Cut
  weiter (heutiges 204-Verhalten).
- [ ] CSRF-Schutz oder Origin-Pin als Default-Vorgabe für aktivierten
  Browser-Pfad.
- [ ] Tests: Browser-Origin mit Allowlist-Match → akzeptiert;
  ohne Match → 204 (RAK-74-Scope-Cut bleibt); ohne aktivierte
  Policy → 204.
- [ ] `make smoke-browser-ingest` als Lab-Pfad.
- [ ] Risks-Backlog R-21: Trigger als „aufgelöst durch RAK-80 in
  0.12.5" markieren.

## 9. Tranche 5 — Optionale R-14 / R-16

Ziel: Optional Auth-Bridge MediaMTX/SRS und/oder Outbound-Webhook.
Tranche-Go nur bei zusätzlicher Bandbreite und Test-Pipeline-
Verfügbarkeit.

DoD (R-14, falls aktiviert):

- [ ] Adapter `MediaMTXAuthBridge` mit signierten Publish-Tokens.
- [ ] Lab-Smoke gegen echtes MediaMTX-Container-Setup.
- [ ] Operator-Runbook für die Aktivierung.

DoD (R-16, falls aktiviert):

- [ ] Outbound-Webhook-Dispatcher (Driven-Port) mit Retry-/Timeout-
  Schema (Exponential-Backoff, Max-Attempts, Dead-Letter-Pfad).
- [ ] HMAC-signierte Payload mit Project-Token-abgeleitetem Secret.
- [ ] Adapter-Test plus Mock-Konsument-Smoke.

Falls eine oder beide Tranchen nicht aktiviert werden: dokumentierte
Defer-Entscheidung in `risks-backlog.md` mit Folge-Plan-Verweis.

## 10. Tranche 6 — Release-Closeout

DoD:

- [ ] `make docs-check` grün.
- [ ] `make gates` grün.
- [ ] `make lastenheft-check` grün.
- [ ] `make rak-trace` grün; alle RAK-77..RAK-80 (plus optional
  RAK-81/82) verlinken auf Code- und Test-Pfade.
- [ ] `make generated-drift-check` grün.
- [ ] Wave-2-Quality-Gates dokumentiert (`releasing.md` §3.1):
  `gh run list --workflow benchmark.yml --limit 1`,
  `gh run list --workflow fuzz.yml --limit 1`,
  `gh issue list --label fuzz --state open`,
  `gh run list --workflow mutation.yml --limit 3`.
- [ ] Versions-Bump auf `0.12.5` an allen Stellen aus
  `releasing.md` §3.1 (5× `package.json` + `main.go`
  `serviceVersion` + `version.ts` + `contracts/sdk-compat.json`
  + 20+20 Analyzer-Fixtures + Test-Fixtures mit hartkodiertem
  Versions-String).
- [ ] `CHANGELOG.md` mit `[0.12.5] - YYYY-MM-DD`-Block:
  `### Added` (neue Adapter-Pfade, neue ENV-Variablen),
  `### Changed` (RAK-74-Scope-Cut bei aktiver Browser-Policy),
  `### Security` (neue Secret-Backend-Adapter, Multi-Key-Rotation).
- [ ] Roadmap-Status aktualisiert: §1 Phase auf released, §2
  Schritt 47.6 ✅, §3-Zeile `0.12.5` ✅.
- [ ] Plan nach `docs/planning/done/plan-0.12.5.md` verschoben;
  Status-Header auf ✅ released; Tranchen-Übersicht §3 alle ✅.
- [ ] Annotierter Tag `v0.12.5` mit Lieferzusammenfassung.
- [ ] GitHub-Release `m-trace 0.12.5` mit Notes-File aus dem
  CHANGELOG-Block.

## 11. Folge-Scope nach `0.12.5`

- [`plan-0.13.0.md`](./plan-0.13.0.md): Production / Ops Backends
  (`MVP-40`..`MVP-44`) — Postgres, ClickHouse, K8s, Devcontainer.
- Optional `plan-0.13.x` für Externe Provisionierung (R-15), falls
  Trigger nach Operator-Bedarf in `0.12.5`-Closeout greift.
- Später `plan-0.14.x` o. ä. für OAuth/OIDC/SSO + User-Verwaltung,
  falls konkreter Bedarf entsteht.

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
