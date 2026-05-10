# Implementation Plan — `0.12.1` (Auth-/Risk-Folge: Trigger-Re-Eval + Operator-Doku)

> **Status**: 🟡 Tranche 0 aktiv (aktiviert 2026-05-10). Vorgänger
> ist `0.12.0` (`v0.12.0`, Auth / Token Lifecycle; Plan in
> [`done/plan-0.12.0.md`](../done/plan-0.12.0.md)).
>
> **Release-Typ**: **Patch-Release** (`0.12.1`) gemäß
> [`docs/user/releasing.md`](../../user/releasing.md) §3.1 — kein
> Lastenheft-Patch, keine RAK-Verifikationsmatrix, keine neue
> User-Surface oder Wire-Verträge. Inhaltlich: Trigger-Re-Eval für
> alle aktiven `R-N`-Items, Operator-Runbooks und Doku-Schärfungen.
>
> **Ziel**: Den Risiken-Backlog nach `0.12.0`-Release konsistent
> halten — getriggerte Items bekommen einen sauberen Stand-Eintrag,
> Out-of-Scope-Items aus `done/plan-0.12.0.md` §10 bekommen klare
> Triggerschwellen mit operator-observablem Signal. Adapter-/Wire-
> Implementierungen (Multi-Replica-Limiter, Multi-Key-Rotation-
> Code, KMS/Vault-Backend, Webhook-Outbound, Ingest-Browser-Policy)
> wandern in [`plan-0.12.5.md`](../open/plan-0.12.5.md) als Minor-Release.
>
> **Bezug**:
> [`done/plan-0.12.0.md`](../done/plan-0.12.0.md) §10 Folge-Scope;
> [`risks-backlog.md`](./risks-backlog.md) §1.1
> R-5/R-7/R-9/R-10/R-11/R-12/R-13/R-14/R-15/R-16/R-17/R-18/R-20/R-21;
> [`docs/user/releasing.md`](../../user/releasing.md) §3.1
> Patch-Release-Konvention.
>
> **Nachfolger**: `plan-0.12.5.md` (Adapter-Minor) und `plan-0.13.0.md`
> (Production / Ops Backends).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR-/Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.12.1` ist ein **Patch-Release** im Sinne von `releasing.md` §3.1
— CI-/Tooling-/Doku-Lieferungen ohne neue User-Surface, ohne neue
Wire-Verträge, ohne neuen Lab-/Adapter-Pfad.

In Scope:

- **Trigger-Re-Eval** für alle aktiven R-N-Items im Backlog
  (R-5, R-7, R-9, R-10, R-11, R-12, R-13, R-14, R-15, R-16, R-17,
  R-18, R-20, R-21). Pro Item: ist der Trigger ausgelöst? Beleg
  (Operator-Report, Smoke-Failure, Lab-Beobachtung). Falls
  ausgelöst → Auflösungs-Item in `0.12.5` oder `0.13.0` planen.
  Falls nicht ausgelöst → Eintrag bleibt unverändert, Stand-Datum
  und Beobachtungs-Notiz aktualisieren.
- **Operator-Runbook für R-18** (Multi-Key-Signing-Rotation): die
  Code-Pfade existieren bereits restart-stabil (
  `StaticSigningKeyResolver` mit aktivem `kid` plus Verify-Keys,
  Test `TestHMACSigner_RestartStableAcrossKeyResolverReinitialization`).
  Was fehlt ist die operator-side Doku: Multi-Key-ENV-Schema,
  Rotation-Order, Smoke-Befehl. Reines Doku-Item — der
  ENV-Schema-Code wandert nach `0.12.5`.
- **Trigger-Schärfung für OS-1..OS-6** aus
  `done/plan-0.12.0.md` §10: jedes OS-Item bekommt ein konkretes
  operator-observables Signal (analog R-13 `expires`-Datum,
  R-17 „zweite API-Replica"). Vage Trigger („Multi-Tenant-/
  Regulated-Requirement") werden geschärft oder fallen weg.
  Die Schärfung erfolgt **ausschließlich im aktuellen
  `risks-backlog.md`** (Konvertierung zu R-N oder Streichung mit
  Begründung); der Done-Plan `done/plan-0.12.0.md` bleibt
  unverändert (Release-Historie).
- **Doku-Schärfungen** aus `0.12.0`-Audit-Findings: `auth.md` §6
  Wire-Code-vs-Metric-Klärung war im T5-Review-Fix bereits drin;
  `releasing.md` §3.1 Bump-Pattern-Sweep + Wave-2-`gh run list`-
  Pflicht ist im Post-Release-Audit (`e958737`) gelandet — beide
  rückwirkend in den `0.12.1`-DoD aufnehmen.
- **CHANGELOG- und Roadmap-Pflege** nach Patch-Konvention:
  `[0.12.1] - YYYY-MM-DD`-Block, Roadmap §2 Schritt 47.5 + §3-Zeile
  ergänzt.

Out of Scope (wandert in `plan-0.12.5.md` oder später):

- Adapter-Implementierungen jeder Art: Shared-State-Limiter
  (`R-17`), Multi-Key-Rotation-Code (`R-18`), KMS/Vault-Backend
  (`R-20`), Browser-Ingest-Policy-Lift (`R-21`),
  Auth-Hook-Bridge (`R-14`), Outbound-Webhook (`R-16`), externe
  Provisionierung (`R-15`).
- Lastenheft-Patch oder neue RAK-Gruppe.
- Neue Wire-Verträge oder Endpoint-Surface.
- Versions-Bump im Sinne von Public-API-Änderung; bumpt werden nur
  die Standard-Stellen aus `releasing.md` §3.1.

### 0.2 Vorgänger-Gate

- `0.12.0` ist released; Roadmap zeigt `0.12.0` als ✅ und
  `0.13.0` als ⬜ aktivierbar.
- Bestehende Auth-/Lab-Grenzen aus `0.12.0` bleiben verbindlich:
  RAK-71 normativer Auth-Scope, RAK-74-Scope-Cut für `/api/ingest/*`,
  R-14/R-17/R-18/R-20/R-21 als getrackte Folge-Items.
- Wave-2-Quality-Gates-Voraussetzung (`releasing.md` §3.1) wird
  vor dem Tag dokumentiert (Lehre aus `0.12.0`-Audit).

### 0.3 Architektur-/Persistenzentscheidung

**Keine** in `0.12.1` — der Plan ist explizit doku-/runbook-only.
Adapter-/Persistenz-Entscheidungen wandern in `plan-0.12.5.md` §0.3.

## 1. Trigger-Re-Eval-Matrix

Pro Risiko ein Eintrag „Trigger ausgelöst? ja/nein/Beleg" plus
Folge-Aktion.

| R-N | Trigger laut Backlog | Ausgelöst? | Folge-Aktion in `0.12.1` |
|---|---|---|---|
| R-5 | ≥ 5 Spans `mtrace.time.skew_warning=true` außerhalb Synthetik in einer Lab-Woche, oder Operator-Report | t.b.d. (DoD-Eintrag mit Beleg) | Stand-Notiz aktualisieren oder Folge-Item für `0.13.0` |
| R-7 | List-Latenz ≥ 200 ms p95 oder Operator-Report | t.b.d. | analog R-5 |
| R-9 | K8s-Smoke-Stage wird eingeführt | t.b.d. (Plan-0.13.0 ist K8s-relevant) | Wenn ja: Folge-Item in `plan-0.13.0.md` koordinieren |
| R-10 | Sampling-Lücke außerhalb Player-Pfad | t.b.d. | Stand-Notiz |
| R-11 | ≥ 1000 persistierte SRT-Health-Samples pro Stream oder Operator-Report | t.b.d. | Stand-Notiz oder Folge-Item für `0.12.5`/`0.13.0` |
| R-12 | Drift-Smoke-Failure | automatisch detektiert | Stand-Notiz: Smoke läuft, kein Befund |
| R-13 | `expires` 2026-08-04 oder Trixie-Point-Release mit CVE-Fix | t.b.d. (Eval-Datum vor Tag prüfen) | Re-Review oder Verlängerung dokumentieren |
| R-14 | Operator-Bug-Report über `validate-key`-Missverständnis | t.b.d. | Doku-Schärfung in `auth.md`/`ingest-control.md` |
| R-15 | Lab-Operator-Bedarf nach automatischer Provisionierung | t.b.d. | Stand-Notiz |
| R-16 | Externer Konsument für Stream-Lifecycle-Webhooks | t.b.d. | Stand-Notiz; Code-Pfad in `plan-0.12.5.md` |
| R-17 | Zweite API-Replica im selben Compose-/K8s-Setup | t.b.d. | Stand-Notiz; Code-Pfad in `plan-0.12.5.md` |
| R-18 | Erstes Key-Rotation-Event in Lab/Staging | t.b.d. | **Operator-Runbook in `0.12.1` liefern** (reine Doku); Code-Pfad in `plan-0.12.5.md` |
| R-20 | Multi-Instance-Setup oder Compliance-Audit (PCI/SOC2) | t.b.d. | Stand-Notiz; Code-Pfad in `plan-0.12.5.md` |
| R-21 | Erster Browser-Pfad gegen `/api/ingest/*` | t.b.d. | Stand-Notiz; Code-Pfad in `plan-0.12.5.md` |

## 2. OS-1..OS-6 Trigger-Schärfung

`done/plan-0.12.0.md` §10 listet sechs Folge-Themen mit teils vagen
Triggern. `0.12.1` schärft sie auf konkrete operator-observables
Signale. Ergebnis ist eine `risks-backlog.md`-§1.1-Zeile pro Item
oder eine begründete Streichung („nicht trackbar ohne konkreten
Bedarf").

| OS | Folge-Item | Aktueller Trigger | Schärfung in `0.12.1` |
|---|---|---|---|
| OS-1 | OAuth/OIDC/SSO + User-/Org-Verwaltung | „Multi-Tenant-/Regulated-Requirement" | Bleibt RAK-71-Out-of-Scope ohne R-N (kein konkreter Bedarf trackbar); Streichung dokumentieren |
| OS-2 | Admin-UI / Auth-Management-UI | implizit, mit OS-1 verzahnt | Streichung dokumentieren |
| OS-3 | Produktive MediaMTX-/SRS-Auth-Hook-Brücke | implizit über R-14 | R-14 ist die getrackte Form; OS-3 als Duplikat streichen |
| OS-4 | KMS/Vault/Cloud-Secret-Manager | implizit über R-20 | R-20 ist die getrackte Form; OS-4 als Duplikat streichen |
| OS-5 | Multi-Replica Secret-/Issuance-Mechanik | implizit über R-17 + R-18 | Streichung als Duplikat |
| OS-6 | Origin-/IP-nahe Rate-Limit-Buckets | „Pflicht-/Opt-in-Implementierung inkl. Messgröße" | `0.12.5` deckt Project-Token-basiertes Rate-Limiting ab (R-17/RAK-77), aber **kein** Origin-/IP-Rate-Limiting — die OS-6-Bedingung „Folge-Item nötig" greift. Tranche 1 legt **R-22** in `risks-backlog.md` §1.1 an mit operator-observablem Trigger („IP-basiertes Last-/Replay-Pattern im Operator-Report" oder „Issuance-Abuse trotz aktivem R-17-Limiter") und Auflösungspfad „[`plan-0.13.0.md`](../open/plan-0.13.0.md) bzw. `plan-0.13.x`, sobald Trigger ausgelöst". Alternative: Streichung mit Begründung „Project-Token-Rotation + R-17 decken Missbrauch ausreichend ab" — Tranche-1-Entscheidung. |

## 3. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| --- | --- | --- |
| 0 | Plan-Aktivierung, Roadmap-Insert, Trigger-Re-Eval-Matrix vorbereiten | ✅ |
| 1 | Trigger-Re-Eval pro R-N + OS-Schärfung in `risks-backlog.md` umsetzen | 🟡 |
| 2 | Operator-Runbook für R-18 + Doku-Schärfungen | ⬜ |
| 3 | Closeout: Versions-Bump, CHANGELOG, Plan-Move, Tag | ⬜ |

---

## 4. Tranche 0 — Aktivierung

Ziel: Patch-Scope ist sauber, bevor R-N-Re-Evals laufen.

DoD:

- [x] Plan von `docs/planning/open/plan-0.12.1.md` nach
  `docs/planning/in-progress/plan-0.12.1.md` verschoben (Commit
  folgt am Ende von Tranche 0).
- [x] `git status --short` vor erster Änderung dokumentiert: clean
  nach `63a7fa0` (drei Review-Fixes); Aktivierung beginnt auf
  einem aufgeräumten Tree.
- [x] Roadmap-Insert: §1 Phase-Header auf 🟡 `0.12.1` in Arbeit;
  §1.2 „aktivierbar" → „aktiviert 2026-05-10"; §2 Schritt 47.5
  Status auf 🟡; §3 Release-Übersicht-Zeile `0.12.1` auf 🟡.
- [x] Patch-Release-Konvention bestätigt: kein Lastenheft-Patch,
  keine RAK-Matrix, keine neue Wire-Surface — siehe §0.1 Scope-
  Definition oben.

## 5. Tranche 1 — Trigger-Re-Eval und OS-Schärfung

Ziel: Risiken-Backlog ist nach `0.12.0`-Release konsistent.

DoD:

- [ ] Pro R-N (R-5, R-7, R-9, R-10, R-11, R-12, R-13, R-14, R-15,
  R-16, R-17, R-18, R-20, R-21): „Trigger ausgelöst? ja/nein, Beleg"
  als ein-Zeilen-Notiz in der Mitigation-Spalte ergänzt; Stand-Datum
  in der jeweiligen Zeile aktualisiert. Falls ausgelöst: Folge-Item
  in `plan-0.12.5.md` oder `plan-0.13.0.md` referenziert.
- [ ] R-13 spezifisch: `expires`-Datum 2026-08-04 vs. aktuelles
  Datum prüfen; Trixie-Point-Release-Status checken; Re-Review
  oder Verlängerung mit Begründung dokumentieren.
- [ ] R-12 spezifisch: letzten `webrtc-drift.yml`-Nightly-Run
  zitieren als Beleg (Status grün → kein Trigger).
- [ ] OS-1..OS-6 ausschließlich im `risks-backlog.md` aufgelöst
  (jedes Item entweder zu R-N umgewidmet — z. B. `OS-6` → neuer
  R-22 mit Trigger und Auflösungs-Plan-Verweis — oder als
  Duplikat/strukturell-nicht-trackbar mit Streichungs-Begründung
  dokumentiert). **Done-Plan `done/plan-0.12.0.md` bleibt
  unverändert** — Release-Historie ist immutabel; OS-1..OS-6
  bleiben dort als ursprünglicher Folge-Scope-Stand sichtbar,
  ihre Schärfung/Streichung wird im aktuellen Backlog
  nachgezogen.
- [ ] R-19-Lücke (im T5-Cleanup gestrichen) als historische Notiz
  im Risks-Backlog erhalten („n/a, war auf README-Aussage gegründet,
  die in `5798473` entfernt wurde").

## 6. Tranche 2 — Operator-Runbook + Doku-Schärfungen

Ziel: Operator hat einen funktionierenden Multi-Key-Rotation-Pfad
(Doku-Stand) und die Audit-Findings aus dem `0.12.0`-Post-Release
sind konsolidiert.

DoD:

- [ ] **Operator-Runbook für Signing-Key-Rotation** in
  `docs/user/auth.md` §5.3 erweitert: konkrete Rotation-Reihenfolge
  (neuen Key generieren → in ENV-Liste aufnehmen → aktiven KID
  umschalten → alten Key nach max-TTL-Fenster + Reservezeit aus
  ENV entfernen). Multi-Key-ENV-Schema bleibt heute nicht
  implementiert — der Runbook beschreibt den **Soll-Workflow**;
  ENV-Parser-Code-Pfad wandert in `plan-0.12.5.md`. Verweis auf
  R-18 setzen.
- [ ] `auth.md` §6 Wire-Code-vs-Metric-Wording wurde im T5-Review-
  Fix (`0ebeed5`) bereits korrigiert — als bestehender Stand im
  `0.12.1`-DoD nachweisen, kein neuer Edit.
- [ ] `releasing.md` §3.1 Bump-Pattern-Sweep + Wave-2-`gh run list`-
  Pflicht wurde im Post-Audit (`e958737`) bereits committed —
  als bestehender Stand nachweisen, kein neuer Edit.
- [ ] CHANGELOG-Format-Hygiene aus dem Audit-Finding (`e958737`)
  ist drin; im `0.12.1`-Block bleibt das Format Keep-a-Changelog.

## 7. Tranche 3 — Release-Closeout

DoD:

- [ ] `make docs-check` grün.
- [ ] `make gates` grün (zur Vollständigkeit, auch wenn der Patch
  doku-only ist — Test-Drift-Schutz).
- [ ] `make generated-drift-check` grün.
- [ ] Wave-2-Quality-Gates dokumentiert per `releasing.md` §3.1:
  `gh run list --workflow benchmark.yml --limit 1`,
  `gh run list --workflow fuzz.yml --limit 1`,
  `gh issue list --label fuzz --state open`,
  `gh run list --workflow mutation.yml --limit 3`. Verdict im
  Closeout-Log oder Tag-Annotation festgehalten.
- [ ] Versions-Bump auf `0.12.1` an allen Stellen aus
  `releasing.md` §3.1 (5× `package.json` + `main.go`
  `serviceVersion` + `version.ts` + `contracts/sdk-compat.json`
  + 20+20 Analyzer-Fixtures + Test-Fixtures mit hartkodiertem
  Versions-String; Pattern-Sweep mit den drei `grep`-Patterns
  aus §3.1).
- [ ] `CHANGELOG.md` mit `[0.12.1] - YYYY-MM-DD`-Block aktualisiert
  (Keep-a-Changelog: `### Changed` für Trigger-Re-Evals und Doku;
  ggf. `### Removed` für gestrichene OS-Items).
- [ ] Roadmap-Status aktualisiert: §1 Phase auf released, §2
  Schritt 47.5 ✅, §3-Zeile `0.12.1` ✅.
- [ ] Plan nach `docs/planning/done/plan-0.12.1.md` verschoben;
  Status-Header auf ✅ released; Tranchen-Übersicht §3 alle ✅.
- [ ] Annotierter Tag `v0.12.1` erstellt mit Lieferzusammenfassung
  (Trigger-Re-Eval-Stand, Operator-Runbook, Doku-Schärfungen) plus
  Verweis auf `plan-0.12.5.md` als Folge-Minor.
- [ ] GitHub-Release `m-trace 0.12.1` mit Notes-File aus dem
  CHANGELOG-Block.

## 8. Folge-Scope nach `0.12.1`

- [`plan-0.12.5.md`](../open/plan-0.12.5.md): Auth-/Ingest-Adapter-Minor
  (Multi-Replica-Limiter R-17, Multi-Key-Rotation-Code R-18,
  KMS/Vault R-20, Browser-Ingest-Policy R-21, ggf. Auth-Bridge
  R-14, Webhook-Outbound R-16). Eigener Lastenheft-Patch `1.1.16`,
  neue RAK-Gruppe.
- [`plan-0.13.0.md`](../open/plan-0.13.0.md): Production / Ops Backends
  (`MVP-40`..`MVP-44`).
- Später: `plan-0.14.x` o. ä. für OAuth/OIDC/SSO/User-Verwaltung,
  falls konkreter Bedarf entsteht (RAK-71-Out-of-Scope-Stand bleibt
  bis dahin).

## 9. Qualitätsregeln für `0.12.1`

- Keine Code-Adapter-Lieferung. Wenn ein R-N-Trigger ausgelöst ist
  und nur Code es löst, wandert das Item in `plan-0.12.5.md`.
- Jede `defer`-Entscheidung enthält: konkretes operator-observables
  Signal als Trigger, Auflösungspfad und Folge-Plan-Verweis.
- Jede Doku-Änderung verweist auf das normative Lastenheft / API-
  Kontrakt / Plan, nicht auf eine README-Sektion (Memory-Lehre
  `feedback_lastenheft_normativ.md`).
- Wave-2-Verdict vor Tag dokumentieren (Lehre aus `0.12.0`-Audit).
