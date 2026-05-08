# Implementation Plan — `0.9.6` (Lastenheft-Konvergenz + Repo-Artefakte)

> **Status**: 🟡 in Arbeit — am 2026-05-08 nach `in-progress/`
> aktiviert. Vorgänger `0.9.5` ist released; Plan archiviert in
> [`done/plan-0.9.5.md`](../done/plan-0.9.5.md).
>
> **Release-Typ**: Patch-Release nach `0.9.5` mit vollständigem
> Versions-Bump und Tag `v0.9.6` analog
> [`docs/user/releasing.md`](../../user/releasing.md) mit der
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
> [`spec/lastenheft.md`](../../../spec/lastenheft.md) F-7, F-131
> (neu), NF-13, NF-18, NF-25, NF-29, MVP-19..MVP-26,
> MVP-40..MVP-42;
> [`docs/planning/in-progress/roadmap.md`](./roadmap.md)
> Release-Übersicht und Folge-ADRs;
> [`docs/planning/in-progress/risks-backlog.md`](./risks-backlog.md)
> R-9/R-13;
> [`README.md`](../../../README.md) mit der Überschrift
> „Was m-trace nicht ist".
>
> **Nachfolger**: [`plan-0.10.0.md`](../open/plan-0.10.0.md) für NF-13 /
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
| NF-13 | CMAF-Analyse als Muss; `F-73` deckt nur die Erweiterbarkeit, nicht die Vollanalyse. | Folge-Plan: in [`plan-0.10.0.md`](../open/plan-0.10.0.md) mit neuer RAK verankern; Lastenheft patchen (Verweis auf Folge-Plan, kein Down-Grade). | Tranche 2 (Verweis), Folge-Plan `0.10.0`. |
| NF-18 / MVP-42 | K8s-Deployment als Muss steht im Widerspruch zur README-Abgrenzung („kein Production-K8s") und zum `MVP-42`-Kann-Scope. | Lastenheft patchen: K8s Production bleibt out of scope; `MVP-42` als optional/Folge-Plan; R-9 bleibt Trigger-Risiko. | Tranche 2 (Patch), Tranche 3 (Risk-Header). |
| MVP-19..MVP-26 | „Nicht im `0.1.0`-MVP"-Liste enthält historisch ausgelieferte Einträge (`MVP-24`, `MVP-25`) und einen später auf Muss hochgezogenen Eintrag (`MVP-37`); zusätzlich offene Kann-Themen (`MVP-40`..`MVP-42`). | Lastenheft patchen: redaktionell so klären, dass die historische Nicht-`0.1.0`-Liste nicht als heutige offene Muss-Lücke missverstanden wird. | Tranche 2. |

---

## 3. Tranche 1 — Fehlende Muss-Artefakte

Ziel: kleine, nicht kontroverse Lastenheft-Lücken schließen, ohne neue
Produktfunktionen zu bauen.

DoD:

- [x] [`CONTRIBUTING.md`](../../../CONTRIBUTING.md) angelegt mit:
  - lokalem Setup-Verweis auf [`docs/user/local-development.md`](../../user/local-development.md),
  - Build/Test-Hinweis über `make` (Single-Entry-Point, keine
    direkten `pnpm`/`docker`-Aufrufe),
  - Commit-/Release-Konventionsverweis auf [`docs/user/releasing.md`](../../user/releasing.md),
  - Erwartung an Issues/PRs (Lastenheft-Bezug) und
    Security-Meldungen über `SECURITY.md`.
- [x] [`SECURITY.md`](../../../SECURITY.md) angelegt mit:
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
- [x] [`.env.example`](../../../.env.example) angelegt mit
  dokumentierten, nicht geheimen Beispielwerten für API,
  Analyzer, Dashboard, Observability (alle aktuell in
  `docker-compose.yml` referenzierten `MTRACE_*`/`ANALYZER_*`/
  `PUBLIC_*`/`OTEL_*`/`MEDIAMTX_*`-Vars). Keine realen Tokens,
  keine privaten URLs; SRT-Health-Vars als Kommentar-Block
  optional.
- [x] `deploy/`-Struktur angelegt:
  - [`deploy/README.md`](../../../deploy/README.md) macht
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

- [ ] Header von [`spec/lastenheft.md`](../../../spec/lastenheft.md)
  auf `1.1.12` erhöht.
- [ ] Patch-Notiz ergänzt: „Lastenheft-Konvergenz nach `0.9.5`;
  keine neue Produktfunktion".
- [ ] [`plan-0.1.0.md`](../done/plan-0.1.0.md) Tranche 0c um
  `4a.15 Patch 1.1.12` ergänzt und die fortlaufende Patch-Übersicht
  von `1.1.11` auf `1.1.12` nachgezogen.
- [ ] `F-7` präzisiert: `deploy/` ist Struktur-Anker für spätere
  Deployment-Artefakte; lokale Compose-Root-Datei bleibt aktuell der
  primäre unterstützte Pfad.
- [ ] Pflichtdokumente-Block erhält die neue Kennung `F-131`
  (`Muss`: Pflichtdokumente müssen vorhanden und auf aktuelle
  Repository-Pfade harmonisiert sein).
- [ ] `F-131` auf reale Pfade harmonisiert:
  `docs/user/stream-analyzer.md` statt `docs/stream-analyzer.md`;
  `CONTRIBUTING.md`, `SECURITY.md`, `.env.example` als vorhanden
  referenziert.
- [ ] `NF-13` präzisiert: `F-73` ist nur die vorbereitete
  Erweiterbarkeit; die CMAF-Vollanalyse bleibt als offene Muss-
  Anforderung bestehen und wird nicht durch `0.9.6` geschlossen. Für
  die Umsetzung verweist das Lastenheft auf
  [`plan-0.10.0.md`](../open/plan-0.10.0.md) mit neuer RAK und eigener
  Akzeptanzmatrix.
- [ ] `NF-18` präzisiert: Kubernetes Production ist nicht
  Bestandteil der ersten Projektphase; `MVP-42` bleibt Kann/Folge-
  Plan. R-9 bleibt Trigger-Risiko für eine künftige K8s-Smoke-Stage.
- [ ] `MVP-19`..`MVP-26` und `MVP-37` redaktionell bereinigt:
  historische „Nicht im `0.1.0`-MVP"-Einträge dürfen nicht mehr als
  offene `0.9.x`-Muss-Lücken missverstanden werden. Erledigte spätere
  Muss-Hebungen (`MVP-24`, `MVP-25`, `MVP-37`) bleiben mit Verweis auf
  die jeweiligen RAKs markiert; weiterhin optionale Themen (`MVP-40`..
  `MVP-42`) bleiben Folge-Scope.
- [ ] Kein neuer Scope-Widerspruch zwischen README-Abgrenzung,
  `NF-13`, `NF-18`, `MVP-19`..`MVP-26` und `MVP-40`..`MVP-42`.

---

## 5. Tranche 3 — Roadmap/Risiken/README synchronisieren

Ziel: nach dem Patch ist klar, was erledigt ist und was bewusst offen
bleibt.

DoD:

- [ ] [`roadmap.md`](./roadmap.md) Statusblock und
  Release-Übersicht ergänzen
  `0.9.6` als Lastenheft-Konvergenz-Patch.
- [ ] [`roadmap.md`](./roadmap.md) Folge-ADRs bleiben bei
  Postgres/Multi-Tenant-Folge-ADRs, ergänzt aber nur neue Folge-ADRs,
  falls Tranche 2 wirklich eine neue Entscheidung erzeugt.
- [ ] [`risks-backlog.md`](./risks-backlog.md) Header
  auf `0.9.6`-Stand aktualisiert; R-9/R-13 nur dann inhaltlich
  geändert, wenn der Lastenheft-Patch ihre Trigger beeinflusst.
- [ ] [`README.md`](../../../README.md) Statusblock und
  Abgrenzungsabschnitt bleiben konsistent: kein Production-K8s,
  keine Multi-Tenant-SaaS-Plattform, kein Production-Grade-
  Storage-Backend.
- [ ] `CHANGELOG.md` erhält einen `[Unreleased]`-Eintrag für
  Lastenheft-/Repo-Artefakt-Konvergenz.

---

## 6. Tranche 4 — Gates, Release-Closeout und Tag

Ziel: Patch sauber abschließen.

DoD:

- [ ] `make docs-check` grün.
- [ ] `make build` grün.
- [ ] `make gates` grün.
- [ ] `make security-gates` grün oder CI-Job `Security gates` grün
  dokumentiert.
- [ ] `make smoke-cli` grün, weil Tranche 2 den
  Stream-Analyzer-Pflichtdokumentpfad korrigiert.
- [ ] Wenn Tranche 1 `.env.example` oder Deployment-Doku mit
  Operator-Relevanz ergänzt: `make smoke-observability` geprüft oder
  begründet nicht nötig.
- [ ] Wave-2-Quality-Gates laut
  [`docs/user/releasing.md`](../../user/releasing.md) mit der
  Überschrift „Patch-Release-Konvention" vor dem Tag
  geprüft:
  - letzter `benchmark.yml`-Nightly geprüft; für Patch reicht
    `make benchmark-smoke` im PR-Pfad, falls keine Performance-
    relevante Änderung vorliegt.
  - kein offenes Crash-Issue mit Label `fuzz` aus dem letzten
    `fuzz.yml`-Nightly.
  - Mutation-Score-Trend aus den letzten drei `mutation.yml`-
    Nightly-Artefakten geprüft; Score-Senkung begründet.
- [ ] Vollständiger `0.9.5` → `0.9.6`-Versions-Bump in allen
  versionsführenden Stellen analog Patch-Release-Konvention aus
  [`docs/user/releasing.md`](../../user/releasing.md):
  Root-/Package-Manifeste, `serviceVersion`, `version.ts`,
  `pack-smoke.mjs`, `contracts/sdk-compat.json` und Test-/Contract-
  Fixtures mit hartkodierten Versionsstrings.
- [ ] `CHANGELOG.md`: `[Unreleased]` in `[0.9.6] - YYYY-MM-DD`
  überführt.
- [ ] Plan nach `docs/planning/done/plan-0.9.6.md` verschoben und
  Status auf ✅ released aktualisiert.
- [ ] Annotierter Tag `v0.9.6` erstellt.

## 7. Nicht-Ziele für Review

Review-Kommentare zu den folgenden Themen sollen in Folge-Pläne, nicht
in diesen Patch:

- „Jetzt auch CMAF implementieren."
- „Kubernetes-Deployment produktionsreif machen."
- „Postgres/ClickHouse/Mimir anbinden."
- „Control-Plane oder Multi-Tenant-Betrieb starten."
- „Stream Analyzer API umbauen oder Wire-Vertrag ändern."
