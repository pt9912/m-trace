# Implementation Plan — `0.9.6` (Lastenheft-Konvergenz + Repo-Artefakte)

> **Status**: ✅ released — Tag `v0.9.6` am 2026-05-08; Plan
> archiviert in [`done/plan-0.9.6.md`](./plan-0.9.6.md).
> Vorgänger `0.9.5` ist released; Plan archiviert in
> [`done/plan-0.9.5.md`](./plan-0.9.5.md).
>
> **Release-Typ**: Patch-Release nach `0.9.5` mit vollständigem
> Versions-Bump und Tag `v0.9.6` analog
> [`docs/user/releasing.md`](../../../user/releasing.md) mit der
> Überschrift „Patch-Release-Konvention". Ziel ist
> Lastenheft-Konvergenz und fehlende Repository-Artefakte, keine neue
> Produktfunktion.
> Patch-Release-Konvention bleibt anwendbar: Der Lastenheft-Patch ist
> hier erlaubt, weil er keine neuen RAKs, keine User-Surface und keine
> Wire-Verträge einführt, sondern bestehende Scope-Widersprüche
> und Lieferstands-Unschärfen bereinigt.
>
> **Lastenheft-Status**: Patch `1.1.12` erforderlich. `1.1.11` ist
> technisch bis RAK-59 ausgeliefert, enthält aber noch Scope-
> Unschärfen, offene Folge-Muss-Lücken und Pflichtdokument-/Struktur-Lücken, die beim Audit
> nach `0.9.5` sichtbar wurden.
>
> **Bezug**:
> [`spec/lastenheft.md`](../../../../spec/lastenheft.md) F-7, F-131
> (neu), NF-13, NF-18, NF-25, NF-29, MVP-19..MVP-26,
> MVP-40..MVP-42;
> [`docs/planning/in-progress/roadmap.md`](../in-progress/roadmap.md)
> Release-Übersicht und Folge-ADRs;
> [`docs/planning/in-progress/risks-backlog.md`](../risks-backlog.md)
> R-9/R-13;
> [`README.md`](../../../../README.md) mit der Überschrift
> „Was m-trace nicht ist".
>
> **Nachfolger**: [`plan-0.10.0.md`](../done/plan-0.10.0.md) für NF-13 /
> CMAF-Analyse vorbereitet.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz oder offene
  Entscheidung.
- 🟡 in Arbeit.

### 0.1 Zielbild

`0.9.6` beantwortet die Frage „Ist alles aus dem Lastenheft umgesetzt?"
präzise und maschinen-/reviewbar:

1. Kleine, unstrittige Muss-Lücken werden als Repo-Artefakte ergänzt
   (`CONTRIBUTING.md`, `SECURITY.md`, `.env.example`, `deploy/`-
   Struktur).
2. Stale oder missverständliche Lastenheft-Aussagen werden mit Patch
   `1.1.12` in einen auditierbaren Lieferstands- und Folgeplan-Status
   überführt.
3. Noch bewusst offene Folge-Themen werden nicht still als erledigt
   markiert, sondern in Roadmap/Risiken klar als Trigger- oder
   Folge-Plan-Themen verankert.

### 0.1.1 Rückverfolgbarkeit

Bis der Lastenheft-Patch `1.1.12` umgesetzt ist, bleibt der bisherige
Herkunftsblock für Pflichtdokumente der unnummerierte
Pflichtdokumente-Block im Lastenheft. `F-131` ist in diesem offenen
Plan die geplante Zielkennung für denselben Anforderungsblock. Audit-
und DoD-Einträge müssen deshalb den Herkunftsblock nur beschreibend
und die Zielkennung als eigentliche Kennung nennen:

- vor Patch `1.1.12`: Pflichtdokumente-Block ohne Kennung;
- nach Patch `1.1.12`: `F-131`.

### 0.2 Out-of-Scope-Klauseln

- Keine Implementierung von CMAF-Analyse. `F-73` bleibt der bereits
  gelieferte Vorbereitungsschritt; `NF-13` bleibt eine offene Muss-
  Vollimplementierung für einen Folge-Feature-Plan mit eigener RAK und
  eigener Akzeptanzmatrix.
- Kein Kubernetes-Produktionsbetrieb und keine K8s-Smoke-Stage.
  `NF-18`/`MVP-42` werden auf optionalen bzw. Folge-Scope
  harmonisiert.
- Kein Multi-Tenant-Betrieb, keine Postgres-/ClickHouse-/Mimir-
  Integration, keine Control-Plane.
- Keine neue API- oder SDK-Wire-Vertragsänderung.
- Keine Verschiebung bestehender Doku-Ordner. Wo das Lastenheft alte
  Pfade nennt, wird bevorzugt das Lastenheft auf die aktuelle
  `docs/user/...`-Struktur korrigiert.

### 0.3 Sequenzierung und harte Gates

1. Tranche 0 (Audit-Snapshot + Plan-Aktivierung) ist Pflicht.
2. Tranche 1 (fehlende Repo-Artefakte) kann unabhängig von Tranche 2
   umgesetzt werden.
3. Tranche 2 (Lastenheft-Patch `1.1.12`) muss vor Tranche 3
   abgeschlossen sein, damit Roadmap/Risiken auf den finalen Scope
   verweisen.
4. Tranche 4 (Closeout) erst nach grünen Gates.

### 0.4 Prüfstrategie

Minimal verpflichtend:

- `make docs-check`
- `make build`
- `make gates`
- `make security-gates` oder grüner CI-Job `Security gates`

Zusätzliche Smokes nur, wenn eine Änderung den jeweiligen Bereich
berührt:

- `make smoke-cli` bei Stream-Analyzer-Doku/Pfadänderungen.
- `make smoke-observability` bei Änderungen an Security-, Env- oder
  Observability-Doku, die Operator-Konfiguration betreffen.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Audit-Snapshot + Plan-Aktivierung + auditierbare Befundliste aus Lastenheft vs. Repo | ⬜ |
| 1 | Fehlende Muss-Artefakte ergänzen (`CONTRIBUTING.md`, `SECURITY.md`, `.env.example`, `deploy/`) | ⬜ |
| 2 | Lastenheft-Patch `1.1.12`: Lieferstands-Unschärfen und stale Pfade bereinigen | ⬜ |
| 3 | Roadmap/Risiken/README mit Patch-Status synchronisieren | ⬜ |
| 4 | Gates, CHANGELOG/Version/Closeout, Plan nach `done/`, Tag `v0.9.6` | ⬜ |

---

## 2. Tranche 0 — Audit-Snapshot + Plan-Aktivierung

Ziel: die bekannten Befunde festhalten, bevor Dateien geändert werden.

DoD:

- [x] `git status --short` vor erster Änderung dokumentiert
  (Vor-Aktivierungs-Snapshot 2026-05-08: clean, HEAD `7d64381`).
  Damit sind keine unrelated User-Änderungen mit dieser Tranche
  vermischt.
- [x] Plan von `docs/planning/open/plan-0.9.6.md` nach
  `docs/planning/in-progress/plan-0.9.6.md` verschoben (`git mv`,
  Status `R`; Tranche-3 wird Roadmap/Risiken erst nach
  Lastenheft-Patch nachziehen).
- [x] Roadmap-Status und Release-Übersicht auf `0.9.6` als aktive
  Folgephase umgestellt (`roadmap.md` Stand 2026-05-08, §1
  Header + §1.2 + §2 Schritt 44 + §3 Release-Übersicht).
- [x] Audit-Snapshot im Plan ergänzt: Tabelle unten mit
  Lastenheft-Kennung, Ist-Zustand, Entscheidung
  (`implementieren` / `Lastenheft patchen` / `Folge-Plan`) und
  Ziel-Tranche.

Audit-Snapshot 2026-05-08:

| Kennung | Befund / Ist-Zustand | Entscheidung | Ziel-Tranche |
| ------- | -------------------- | ------------ | ------------ |
| F-7 | `deploy/` wird als Muss-Struktur gefordert; im Repo nicht vorhanden (`ls deploy/` → not found). | implementieren (Struktur + README) + Lastenheft patchen (Status-Präzisierung: Compose-Root bleibt primärer Pfad). | Tranche 1 (Struktur), Tranche 2 (Patch). |
| Pflichtdokumente-Block → F-131 (Zielkennung) | Block hat keine eigene Kennung; `CONTRIBUTING.md` und `SECURITY.md` fehlen; `docs/stream-analyzer.md` ist stale (nur `docs/user/stream-analyzer.md` existiert). | implementieren (Artefakte) + Lastenheft patchen (`F-131` einführen, Pfade korrigieren). | Tranche 1 (Artefakte), Tranche 2 (`F-131`). |
| NF-25 | `.env.example` als Muss; im Repo nicht vorhanden. | implementieren. | Tranche 1. |
| NF-29 | `SECURITY.md` als Muss; im Repo nicht vorhanden. | implementieren. | Tranche 1. |
| NF-13 | CMAF-Analyse als Muss; `F-73` deckt nur die Erweiterbarkeit, nicht die Vollanalyse. | Folge-Plan: in [`plan-0.10.0.md`](../done/plan-0.10.0.md) mit neuer RAK verankern; Lastenheft patchen (Verweis auf Folge-Plan, kein Down-Grade). | Tranche 2 (Verweis), Folge-Plan `0.10.0`. |
| NF-18 / MVP-42 | K8s-Deployment als Muss steht im Widerspruch zur README-Abgrenzung („kein Production-K8s") und zum `MVP-42`-Kann-Scope. | Lastenheft patchen: K8s Production bleibt out of scope; `MVP-42` als optional/Folge-Plan; R-9 bleibt Trigger-Risiko. | Tranche 2 (Patch), Tranche 3 (Risk-Header). |
| MVP-19..MVP-26 | „Nicht im `0.1.0`-MVP"-Liste enthält historisch ausgelieferte Einträge (`MVP-24`, `MVP-25`) und einen später auf Muss hochgezogenen Eintrag (`MVP-37`); zusätzlich offene Kann-Themen (`MVP-40`..`MVP-42`). | Lastenheft patchen: redaktionell so klären, dass die historische Nicht-`0.1.0`-Liste nicht als heutige offene Muss-Lücke missverstanden wird. | Tranche 2. |

---

## 3. Tranche 1 — Fehlende Muss-Artefakte

Ziel: kleine, nicht kontroverse Lastenheft-Lücken schließen, ohne neue
Produktfunktionen zu bauen.

DoD:

- [x] [`CONTRIBUTING.md`](../../../../CONTRIBUTING.md) angelegt mit:
  - lokalem Setup-Verweis auf [`docs/user/local-development.md`](../../../user/local-development.md),
  - Build/Test-Hinweis über `make` (Single-Entry-Point, keine
    direkten `pnpm`/`docker`-Aufrufe),
  - Commit-/Release-Konventionsverweis auf [`docs/user/releasing.md`](../../../user/releasing.md),
  - Erwartung an Issues/PRs (Lastenheft-Bezug) und
    Security-Meldungen über `SECURITY.md`.
- [x] [`SECURITY.md`](../../../../SECURITY.md) angelegt mit:
  - unterstützten Versionen (`0.9.x` aktiv, `0.8.x` und älter
    nicht mehr),
  - vertraulichem Meldeweg (GitHub Security Advisories bzw.
    Maintainer-Mail), Bestätigung in 7 Tagen,
  - Hinweis, keine Secrets/Exploits öffentlich in Issues zu
    posten,
  - Bezug auf Security-Gates aus `0.8.5` (`vuln-check`,
    `audit-ts`, `image-scan`, Generated-Drift-Gate),
  - Disclosure-Pfad über Patch-Release-Konvention und
    `CHANGELOG`.
- [x] [`.env.example`](../../../../.env.example) angelegt mit
  dokumentierten, nicht geheimen Beispielwerten für API,
  Analyzer, Dashboard, Observability (alle aktuell in
  `docker-compose.yml` referenzierten `MTRACE_*`/`ANALYZER_*`/
  `PUBLIC_*`/`OTEL_*`/`MEDIAMTX_*`-Vars). Keine realen Tokens,
  keine privaten URLs; SRT-Health-Vars als Kommentar-Block
  optional.
- [x] `deploy/`-Struktur angelegt:
  - [`deploy/README.md`](../../../../deploy/README.md) macht
    expliziten Status: `docker-compose.yml` im Repo-Root bleibt
    der unterstützte lokale Pfad; `deploy/k8s/` ist Folge-Scope
    (`MVP-42`) und kein Production-Ready-K8s.
  - Unterordner `deploy/compose/`, `deploy/docker/`,
    `deploy/k8s/` jeweils mit `.gitkeep` belegt.
- [x] README ergänzt: `cp .env.example .env`-Hinweis im
  Lokal-Setup, neuer Abschnitt „Mitarbeit und
  Sicherheitsmeldungen" mit Links auf `CONTRIBUTING.md`,
  `SECURITY.md`, `.env.example`, `deploy/README.md`.

---

## 4. Tranche 2 — Lastenheft-Patch `1.1.12`

Ziel: Anforderungen nicht kleiner reden, sondern ihre Priorität und
Phase korrekt ausdrücken.

DoD:

- [x] Header von [`spec/lastenheft.md`](../../../../spec/lastenheft.md)
  auf `1.1.12` erhöht.
- [x] Patch-Notiz ergänzt unmittelbar nach dem Header-Frontmatter
  („Lastenheft-Konvergenz nach `0.9.5`; keine neue
  Produktfunktion"; verweist auf den Patch-Log-Eintrag in
  `plan-0.1.0.md` §4a.15).
- [x] [`plan-0.1.0.md`](./plan-0.1.0.md) Tranche 0c §4a.15
  ergänzt: Begründung + DoD; die fortlaufende Patch-Übersicht ist
  damit von `1.1.11` (§4a.14) auf `1.1.12` nachgezogen.
- [x] `F-7` präzisiert (§7.1): `deploy/` ist Struktur-Anker für
  spätere Deployment-Artefakte; `docker-compose.yml` im Repo-Root
  bleibt aktuell der primäre unterstützte Pfad; `deploy/k8s/` ist
  ausdrücklich Folge-Scope (`MVP-42`).
- [x] Pflichtdokumente-Block (§7.12) erhält die neue Kennung
  `F-131` (`Muss`: Pflichtdokumente müssen vorhanden und auf
  aktuelle Repository-Pfade harmonisiert sein).
- [x] `F-131` auf reale Pfade harmonisiert: `docs/user/stream-analyzer.md`
  statt `docs/stream-analyzer.md`; `CONTRIBUTING.md`,
  `SECURITY.md` (mit Tranche 1 angelegt) und `.env.example` (mit
  Tranche 1 angelegt) sind in der Pflichtdokumente-Tabelle
  vorhanden bzw. werden vom Patch implizit referenziert.
- [x] `NF-13` präzisiert: `F-73` ist nur die vorbereitete
  Erweiterbarkeit; die CMAF-Vollanalyse bleibt als offene Muss-
  Anforderung bestehen und wird nicht durch `0.9.6` geschlossen.
  Für die Umsetzung verweist das Lastenheft auf
  [`plan-0.10.0.md`](../done/plan-0.10.0.md) mit neuer RAK und
  eigener Akzeptanzmatrix.
- [x] `NF-18` präzisiert: Kubernetes Production ist nicht
  Bestandteil der ersten Projektphase; `MVP-42` bleibt
  Kann/Folge-Plan; R-9 bleibt Trigger-Risiko für eine künftige
  K8s-Smoke-Stage; Strukturanker `deploy/k8s/` mit `0.9.6`
  angelegt, aber leer.
- [x] `MVP-19`..`MVP-26` und `MVP-37` redaktionell bereinigt
  (§12.1, §12.3): „Nicht im `0.1.0`-MVP"-Tabelle bekam eine
  zusätzliche Spalte „Status (Patch `1.1.12`)" mit verbindlicher
  Aussage pro Item (erfüllt anders / erfüllt / out of scope /
  bewusst gegenteilig entschieden); die historische „Prioritaet"-
  Spalte bleibt zur Audit-Nachvollziehbarkeit erhalten. Erledigte
  spätere Muss-Hebungen (`MVP-24` `0.7.0`/`0.8.0`/`0.9.0`,
  `MVP-25` `0.6.0`, `MVP-37` `0.9.0`) sind mit RAK-Verweis
  markiert; `MVP-40`..`MVP-42` bleiben Folge-Scope.
- [x] Kein neuer Scope-Widerspruch zwischen README-Abgrenzung,
  `NF-13`, `NF-18`, `MVP-19`..`MVP-26` und `MVP-40`..`MVP-42`:
  alle drei Stellen verweisen auf denselben Lieferstand
  (CMAF/K8s/Multi-Tenant out of scope für die erste Phase, mit
  Folge-Plan- bzw. Folge-ADR-Trigger).

---

## 5. Tranche 3 — Roadmap/Risiken/README synchronisieren

Ziel: nach dem Patch ist klar, was erledigt ist und was bewusst offen
bleibt.

DoD:

- [x] [`roadmap.md`](../in-progress/roadmap.md) Statusblock (§1 Header, §1.2)
  und Release-Übersicht (§3) ergänzt: `0.9.6` als
  Lastenheft-Konvergenz-Patch; §2 Schritt 44 (🟡 in Arbeit) mit
  Inhaltsbeschreibung.
- [x] [`roadmap.md`](../in-progress/roadmap.md) Folge-ADRs unverändert
  (Postgres/Multi-Tenant-Folge-ADRs aus Roadmap §4 — Patch
  `1.1.12` erzeugt keine neue Folge-ADR-Entscheidung).
- [x] [`risks-backlog.md`](../risks-backlog.md) Header auf
  `0.9.6`-Stand aktualisiert; R-9/R-13 inhaltlich unverändert
  (der `1.1.12`-NF-18-Patch bestätigt nur den bestehenden
  „K8s-Smoke-Einführung"-Trigger, ohne ihn zu verschieben).
- [x] [`README.md`](../../../../README.md) Statusblock auf `0.9.6`
  in Arbeit umgestellt; Abgrenzungsabschnitt „Was m-trace nicht
  ist" unverändert (kein Production-K8s, keine Multi-Tenant-SaaS-
  Plattform, kein Production-Grade-Storage-Backend);
  „Leitende Dokumente"-Verweis auf Lastenheft `1.1.12` nachgezogen.
- [x] [`CHANGELOG.md`](../../../../CHANGELOG.md) erhält einen
  `[Unreleased]`-Eintrag mit „0.9.6 in Arbeit"-Block, Added- und
  Changed-Listen für Repo-Artefakte und Lastenheft-Patch.

---

## 6. Tranche 4 — Gates, Release-Closeout und Tag

Ziel: Patch sauber abschließen.

DoD:

- [x] `make docs-check` grün (mehrfach im Verlauf der Tranchen
  0–3 verifiziert; finaler Lauf in Tranche 4 unten).
- [x] `make build` grün — implizit Bestandteil von `make gates`
  (`build` lädt alle Workspace-Artefakte und ist
  `gates`-Voraussetzung).
- [x] `make gates` grün — Lauf v2 nach Versions-Bump-Commit
  `a891672`, exit 0, alle Stages durch (api-race, ts-test, lint,
  coverage-gate, arch-check, schema-validate,
  generated-drift-check, sdk-pack-smoke, sdk-performance-smoke,
  docs-check); Log unter `/tmp/gates_096_v2.log`.
- [x] `make security-gates` grün — initialer Lauf rot wegen vier
  neuer Go-Stdlib-CVEs aus `go1.26.2` (GO-2026-4982, GO-2026-4980,
  GO-2026-4971, GO-2026-4918), alle „Fixed in: go1.26.3";
  `golang:1.26`-Image-Tag refresht (nun `go1.26.3`), zusätzlich
  Build-Pin explizit auf `golang:1.26.3` in `apps/api/Dockerfile`,
  `apps/api/Makefile` und Root-`Makefile` (`vuln-check`-Target);
  Re-Lauf grün (Log `/tmp/sec_096_v2.log`). Stdlib-Bump folgt der
  `0.8.5`-Präzedenz (OTel-Stack-Bump als `GO-2026-4394`-Fix in
  einem Quality-Gates-Patch).
- [x] `make smoke-cli` grün — alle CLI-Pfade (Master, DASH-VOD,
  unsupported, missing file, no-args, URL-loader, `.bin`-Konsument)
  durch; Log `/tmp/smoke_cli_096.log`.
- [x] `make smoke-observability` nicht erneut nötig:
  `.env.example` dokumentiert ausschließlich die in
  `docker-compose.yml` und `apps/*` bereits gesetzten Defaults
  (keine neuen Env-Vars, kein neuer Cardinality-/Allowlist-
  Eintrag); Operator-Konfiguration unverändert. `make
  smoke-observability` aus dem `0.9.5`-Tag deckt diesen
  Allowlist-Stand ab.
- [x] Wave-2-Quality-Gates: keine performance-/fuzz-/mutation-
  relevante Änderung in `0.9.6`. `make benchmark-smoke` (Hot-Path-
  Code unverändert), `fuzz.yml`-Crash-Issue-Liste (Fuzz-Targets
  unverändert) und Mutation-Score-Trend (Pilot-Module unverändert)
  bleiben auf dem `0.9.5`-Stand; ein neuer Beobachtungs-Run ist
  nicht nötig.
- [x] Vollständiger `0.9.5` → `0.9.6`-Versions-Bump in allen
  versionsführenden Stellen analog Patch-Release-Konvention
  ([`docs/user/releasing.md`](../../../user/releasing.md)
  §3.1) — Commit `a891672`: 39 Files, 85 Inserts/Deletes
  (Root-/Workspace-`package.json`, `apps/api/cmd/api/main.go`
  `serviceVersion`, `packages/player-sdk/src/version.ts`,
  `pack-smoke.mjs`, `contracts/sdk-compat.json`,
  `spec/contract-fixtures/analyzer/*.json` plus Go-Testdata-Kopien,
  hartkodierte SDK-/Analyzer-Version-Strings in Tests). Plan-
  Verweis-Kommentare (`// plan-0.9.5 §X`) bewusst unverändert.
- [x] `CHANGELOG.md`: `[Unreleased]` in `[0.9.6] - 2026-05-08`
  überführt; neuer Security-Block dokumentiert die vier
  Stdlib-CVE-Fixes.
- [x] Plan nach `docs/planning/done/plan-0.9.6.md` verschoben
  (`git mv`); relative Pfade an die neue Lage angepasst
  (`../in-progress/roadmap.md`, `../in-progress/risks-backlog.md`,
  `./plan-0.1.0.md`, `./plan-0.9.5.md`); Status-Header auf
  ✅ released aktualisiert.
- [x] Annotierter Tag `v0.9.6` erstellt (siehe `git show v0.9.6`).

## 7. Nicht-Ziele für Review

Review-Kommentare zu den folgenden Themen sollen in Folge-Pläne, nicht
in diesen Patch:

- „Jetzt auch CMAF implementieren."
- „Kubernetes-Deployment produktionsreif machen."
- „Postgres/ClickHouse/Mimir anbinden."
- „Control-Plane oder Multi-Tenant-Betrieb starten."
- „Stream Analyzer API umbauen oder Wire-Vertrag ändern."
