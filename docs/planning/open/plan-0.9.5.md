# Implementation Plan — `0.9.5` (Quality-Gates Wave 2: Benchmarks + Fuzzing + Mutation)

> **Status**: ⬜ geplant (Plan-Skelett, liegt unter
> `docs/planning/open/`). Vorgänger `v0.9.0` muss released sein,
> bevor dieser Plan in `in-progress/` ziehen darf — siehe §0.1.
> Tranche 0 aktiviert die Phase. Plan wandert dann atomar nach
> `docs/planning/in-progress/`.
>
> **Release-Typ**: Patch-Release nach `0.9.0` (Konvention aus
> `plan-0.8.5.md` §0.6 / `docs/user/releasing.md` §3).
>
> **Lastenheft-Status**: kein Lastenheft-Patch nötig (Quality-Gates,
> keine User-Surface).
>
> **Bezug**: [`extra-gates.md`](./extra-gates.md) §3.2 (Benchmark-
> Smoke), §3.3 (Nightly-`benchstat`-Regressionen), §3.5 (Selektives
> Fuzzing / Property Tests), §3.6 (Mutation Testing) — die vier
> statistisch-langlaufenden Wave-2-Gates aus dem Master-Backlog;
> [`plan-0.8.5.md`](../done/plan-0.8.5.md) Wave 1 (Security + Generated-
> Drift) als Vorlage für die Patch-Release-Mechanik;
> [`done/plan-0.4.0.md`](../done/plan-0.4.0.md) Tranche 1 (Cursor
> v2 — bevorzugter Fuzz-Kandidat aus §3.5);
> [`packages/stream-analyzer/`](../../../packages/stream-analyzer/)
> (Manifest-Parser — bevorzugter Property-Test-Kandidat).
>
> **Nachfolger**: offen — kein `plan-0.10.0.md` vorbereitet.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog
[`done/plan-0.1.0.md`](../done/plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz oder offene
  Entscheidung.
- 🟡 in Arbeit.

Scope-Grenze: dieser Plan liefert die vier statistisch-/langlaufenden
Quality-Gates aus `extra-gates.md` (Wave 2). Im Gegensatz zu Wave 1
(`plan-0.8.5.md`) sind diese Gates **nicht** alle PR-blockierend —
einige laufen Nightly oder Release-only. Die Trennung folgt der
Benchmarking-Policy aus `extra-gates.md` §4.

### 0.1 Vorgänger-Gate

Voraussetzungen, bevor dieser Plan in `in-progress/` gezogen werden
kann:

- **`0.9.0` ist released** (Tag `v0.9.0`). Der Drift-Smoke aus
  [`plan-0.9.0.md`](../done/plan-0.9.0.md) Tranche 1 ist live;
  SRS-Lab und DASH-Analyse sind ausgeliefert. `0.9.5` ist das
  Quality-Hardening-Patch **nach** dem Feature-Release.
- `0.8.5` ist released (Tag `v0.8.5`); Wave-1-Gates (Security +
  Generated-Drift) sind aktiv und PR-blockierend.
- `extra-gates.md` §6 (offene Entscheidungen) ist bei Tranche 0
  geklärt: Baseline-Pfad (Git-Repo, Actions-Artefakt oder Release-
  Asset), initiale Performance-Budgets pro Modul, Quarantäne-
  Policy für laute Benchmarks.

### 0.2 Out-of-Scope-Klauseln (durchgängig)

- Kein `govulncheck`/Container-Scan und kein Generated-Drift-Gate —
  diese sind Wave 1 (`plan-0.8.5.md`).
- Keine Wire-Vertrags-Änderung; reine CI-/Tooling-Lieferung.
- Keine produktiven Telemetrie-Pfade (kein neuer
  `mtrace_*`-Counter); Benchmark-Ergebnisse fließen über
  CI-Artefakte, nicht über die m-trace-API.
- Kein eigenständiger Lastenheft-Patch.

### 0.3 Sequenzierung und harte Gates

1. Tranche 0 (Plan-Aktivierung + Baseline-Entscheidungen) ist
   Pflicht.
2. Tranche 1 (Benchmark-Smoke) ist Voraussetzung für Tranche 2
   (Nightly-`benchstat`-Regressionen) — `benchstat` braucht die
   Benchmarks als Quelle. Reihenfolge ist erzwungen.
3. Tranche 3 (Fuzzing) und Tranche 4 (Mutation Testing) sind
   **unabhängig** voneinander und von Tranche 1+2.
4. Tranche 5 (Closeout) erst nach allen vier inhaltlichen Tranchen.

### 0.4 Implementierungsleitplanken

**Benchmark-Smoke (Tranche 1)**: Bevorzugte Form ist eine
go-Benchmark-Suite (`go test -bench=. -benchmem`) für die in
`extra-gates.md` §3.2 gelisteten API-Kandidaten plus eine
TypeScript-Benchmark-Suite (vitest-bench oder Tinybench) für die
Stream-Analyzer-Kandidaten. PR-Budgets sind absolute Schwellen,
keine Vergleiche. Erste Beobachtungsläufe sind nicht-blockierend
(N=3-5 grüne Runs), bevor das Budget PR-blockierend wird.

**Nightly-`benchstat` (Tranche 2)**: Bevorzugte Form ist ein
neuer GitHub-Actions-Workflow `.github/workflows/benchmark.yml`
(`on: schedule: cron`). Baseline ist ein Git-Branch
`benchmark-baseline` oder ein dedizierter Release-Asset (in
Tranche 0 entschieden). Bei Regression > 15 % auf einem statistisch
signifikanten Benchmark wird ein Issue auto-erstellt.

**Fuzzing (Tranche 3)**: Bevorzugte Form ist Go-Fuzzing
(`go test -fuzz=...`) für die in `extra-gates.md` §3.5 gelisteten
Go-Kandidaten (Cursor, HTTP-Validation, Event-Metadaten,
SRT-Health-Mapping); für TypeScript Property-Tests via
`fast-check`. Fuzzing läuft **nicht** PR-blockierend (zu lang); ein
opt-in `make fuzz-check`-Target läuft mit kurzem `-fuzztime` (z. B.
`30s`), Nightly mit längerem Budget.

**Mutation Testing (Tranche 4)**: Bevorzugte Form ist
`go-mutesting` für ein bis zwei kritische Module (Vorschlag:
`apps/api/hexagon/application/event_meta_validation.go` und
`packages/player-sdk/src/adapters/webrtc/sampling.ts` über
StrykerJS). Initial **nicht-blockierend**; Output als Nightly-
Report-Artefakt. PR-Blockierung erst, wenn die Reports zeigen, dass
die Mutation-Score stabil > 80 % ist.

### 0.5 Test-Fixture-Versions-Drift bei Patch-Release

Identisch zu `plan-0.8.5.md` §0.5: `xargs sed -i` über
Test-Fixture-Files. Folge-Backlog-Item bleibt offen (Tests-aus-
package.json-lesen).

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung + Baseline-Entscheidungen aus `extra-gates.md` §6 (Baseline-Pfad, initiale Budgets, Quarantäne-Policy) | ⬜ |
| 1 | Benchmark-Smoke für API + Stream-Analyzer mit konservativen Budgets, opt-in PR-blockierend nach N grünen Beobachtungsläufen | ⬜ |
| 2 | Nightly-`benchstat`-Regressionen mit Baseline-Vergleich; CI-Workflow `benchmark.yml` (cron) | ⬜ |
| 3 | Selektives Fuzzing (Go) + Property Tests (TypeScript) für Cursor/Parser/URL-Klassifizierung | ⬜ |
| 4 | Mutation Testing als nicht-blockierender Nightly-Report für ein bis zwei kritische Module | ⬜ |
| 5 | Release-Doku, Versions-Bump 0.9.0 → 0.9.5, Plan nach `done/`, Tag `v0.9.5` | ⬜ |

---

## 1a. Tranche 0 — Plan-Aktivierung + Baseline-Entscheidungen

Bezug: `extra-gates.md` §6 (offene Entscheidungen).

DoD:

- [ ] Plan-Skelett von `docs/planning/open/plan-0.9.5.md` nach
  `docs/planning/in-progress/plan-0.9.5.md` verschoben.
- [ ] Baseline-Pfad entschieden: Git-Branch `benchmark-baseline`,
  GitHub-Actions-Artefakt mit Retention oder Release-Asset.
  Default-Empfehlung: Git-Branch (deterministische Historie, kein
  Retention-Limit).
- [ ] Initiale Performance-Budgets pro Modul dokumentiert (Tabelle
  im Plan oder in einem `docs/perf/budgets.md`); Werte sind
  bewusst großzügig (Faktor 2-3 über aktueller Messung).
- [ ] Quarantäne-Policy für laute Benchmarks formuliert: ein
  Benchmark darf max. 30 Tage in Quarantäne, danach entweder
  fix oder entfernt.

---

## 2. Tranche 1 — Benchmark-Smoke für API + Stream-Analyzer

Bezug: `extra-gates.md` §3.2.

Ziel: PR-blockierende Budget-Smokes für die kritischen Hot-Paths.
Budgets sind absolute Schwellen, nicht Diffs.

DoD:

- [ ] Go-Benchmark-Suite in `apps/api/...` für mindestens
  `RegisterPlaybackEventBatch` (typische + maximale Batch-Größe),
  SQLite-Ingest, Session-Listing/Pagination, Cursor Encode/Decode.
- [ ] TypeScript-Benchmark-Suite in `packages/stream-analyzer/...`
  für mindestens HLS-Master/Media-Manifest klein/groß plus
  SSRF-/URL-Klassifizierung; DASH-Manifest-Benchmarks ergänzt,
  sobald `0.9.0` Tranche 3 die DASH-Analyse produktiv hat.
- [ ] `make api-benchmark-smoke` und `make analyzer-benchmark-smoke`-
  Targets im Root-`Makefile`; Wrapper `make benchmark-smoke`.
- [ ] Output enthält Laufzeit, Allokationen oder Durchsatz lesbar;
  Budget-Verletzung erzeugt eindeutige Fehlermeldung mit Ist/Soll.
- [ ] Fixtures sind stabil und versioniert (keine
  Netzwerkabhängigkeit).
- [ ] Beobachtungsphase: erste 3-5 grüne CI-Läufe sind nicht-
  blockierend; danach wird `make benchmark-smoke` PR-blockierend
  (in `make gates` aufgenommen).
- [ ] Jeder Lauf druckt Runner-OS, CPU-Modell und Runtime-Versionen
  (Go, Node), damit Budget-Failures einordenbar bleiben.

---

## 3. Tranche 2 — Nightly-`benchstat`-Regressionen

Bezug: `extra-gates.md` §3.3.

Ziel: Statistisch belastbare Trend-Erkennung mit Baseline-Vergleich.
Nicht im PR-Pfad; Nightly + Release-blockierend.

DoD:

- [ ] CI-Workflow `.github/workflows/benchmark.yml` (`on: schedule:
  cron`) führt die Go-Benchmarks mit `-count=10` aus, lädt die
  Baseline aus dem Pfad aus Tranche 0 und vergleicht via
  `benchstat`.
- [ ] Regressions-Schwelle ist explizit (z. B. > 15 % bei
  statistisch signifikantem Ergebnis); Schwelle ist im Workflow
  oder in `docs/perf/budgets.md` dokumentiert.
- [ ] `benchstat`-Output wird als Workflow-Artefakt gespeichert
  (Retention ≥ 30 Tage).
- [ ] Bei Regression: Auto-Issue oder Slack-/Mail-Alert (in
  Tranche 0 entschieden).
- [ ] Release-Gate in `releasing.md` referenziert den letzten
  grünen Nightly-Run vor Release-Tag als Pflicht-Voraussetzung
  für Minor-Releases.
- [ ] Quarantäne-Mechanik: ein Benchmark kann via
  `// bench:quarantine`-Tag aus dem Vergleich genommen werden;
  Tag verfällt nach 30 Tagen automatisch (Workflow-Skript prüft
  Tag-Alter).

---

## 4. Tranche 3 — Selektives Fuzzing + Property Tests

Bezug: `extra-gates.md` §3.5.

Ziel: Edge-Case-Robustheit für Parser/Decoder/Validierung. Nicht
PR-blockierend; opt-in `make fuzz-check` mit kurzem Budget,
Nightly mit längerem Budget.

DoD:

- [ ] Go-Fuzz-Targets für mindestens: Cursor Encode/Decode (aus
  ADR-0004), HTTP-Validation für Playback-Event-Batches,
  Event-Meta-Validation (`webrtc.*`-Allowlist aus `0.8.0`),
  SRT-Health-Mapping.
- [ ] TypeScript-Property-Tests via `fast-check` für mindestens:
  HLS-Manifest-Parser, DASH-Manifest-Parser (sobald `0.9.0`
  Tranche 3 das produktiv hat), URL-Redaction.
- [ ] `make fuzz-check`-Target im Root-`Makefile` mit kurzem
  `-fuzztime` (Default `30s`); CI-Stage opt-in (manueller Trigger
  oder Nightly).
- [ ] Nightly-Workflow erweitert: längeres Fuzz-Budget (z. B.
  10 min pro Target); gefundene Regressions werden als Issue
  auto-erstellt mit Repro-Test.
- [ ] Doku in `docs/dev/fuzzing.md` (oder ähnlich): Liste der
  aktiven Fuzz-Targets, lokale Reproduktion, Sample-Korpus-Pfad.

---

## 5. Tranche 4 — Mutation Testing (Nightly-Report, nicht-blockierend)

Bezug: `extra-gates.md` §3.6.

Ziel: Test-Qualität messen über Coverage hinaus. Initial nicht-
blockierend; nur Reporting.

DoD:

- [ ] Mutation-Tool entschieden: `go-mutesting` für Go (Modul-
  Auswahl: `apps/api/hexagon/application/event_meta_validation.go`
  als Erst-Kandidat); StrykerJS für TypeScript
  (`packages/player-sdk/src/adapters/webrtc/sampling.ts` als
  Erst-Kandidat).
- [ ] `make mutation-report`-Target im Root-`Makefile`; läuft auf
  einem Modul gleichzeitig (nicht repo-weit).
- [ ] Nightly-Workflow erweitert: führt das Target aus, lädt den
  HTML-Report als Artefakt hoch.
- [ ] Score-Schwelle dokumentiert (z. B. > 70 % Mutation-Score als
  Wunsch-Ziel; PR-Blockierung erst, wenn das Modul die Schwelle
  drei Beobachtungsläufe in Folge erreicht).
- [ ] Doku in `docs/dev/mutation-testing.md`: Liste der Module,
  aktueller Score, lokale Reproduktion.

---

## 6. Tranche 5 — Release-Doku, Closeout

Bezug: `plan-0.8.5.md` §6 als Vorlage.

DoD:

- [ ] `docs/user/releasing.md` §3 referenziert Wave-2-Gates
  (`make benchmark-smoke` PR-blockierend, `make fuzz-check` und
  `make mutation-report` opt-in/Nightly); Release-Voraussetzung
  ist „letzter Nightly-Benchmark grün".
- [ ] `README.md` Status-Block erwähnt `0.9.5` als Patch-Release
  mit Quality-Gates Wave 2.
- [ ] Versions-Bump 0.9.0 → 0.9.5 (analog `plan-0.8.5.md` §4
  Closeout-Mechanik).
- [ ] CHANGELOG: [Unreleased]-Block in `[0.9.5] - YYYY-MM-DD`
  umgewandelt.
- [ ] `make docs-check` grün; `make gates` grün;
  `make benchmark-smoke` grün; Nightly-Workflow läuft sauber
  durch.
- [ ] `plan-0.9.5.md` von `docs/planning/in-progress/` nach
  `docs/planning/done/` verschoben; Cross-Refs angepasst;
  Roadmap §3 zeigt `0.9.5` ✅.
- [ ] Tag `v0.9.5` annotiert; Push opt-in; GitHub-Release.

---

## 7. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`,
  Commit-Hash anhängen.
- `extra-gates.md` ist Master-Backlog für die sechs Gates. Wenn
  ein neues Gate hinzukommt (z. B. License-Compliance), wird es
  zuerst dort erfasst, dann in einem Folge-Plan konkretisiert.
- Quarantäne-Tags in Benchmarks sind self-expiring nach 30 Tagen;
  Tranche-2-Workflow setzt das automatisch um.
- Beobachtungsphase für PR-Blockierung von Wave-2-Gates: jeder
  neue Gate startet nicht-blockierend, wird erst nach
  3-5 grünen Läufen blockierend. Diese Übergänge werden im Plan
  als Wartungs-Eintrag vermerkt.
