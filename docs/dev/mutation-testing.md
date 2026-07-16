# Mutation Testing

> **Status**: Aktiv seit `plan-0.9.5` Tranche 4 (RAK-Wave-2 /
> [`extra-gates.md`](../planning/in-progress/extra-gates.md) §3.6).
> **Initial nicht-blockierend**: nur Nightly-Reporting; PR-
> Blockierung erst nach Beobachtungsphase (siehe §3 Score-Schwelle).
> Workflow:
> [`.github/workflows/mutation.yml`](../../.github/workflows/mutation.yml).

## 1. Tool-Auswahl

| Sprache | Tool | Kandidat in `extra-gates.md` §3.6 | Ausgeliefert |
|---|---|---|---|
| Go | `gremlins.tools` (`github.com/go-gremlins/gremlins`) | go-mutesting | **gremlins** (Substitution) |
| TypeScript | StrykerJS + `@stryker-mutator/vitest-runner` | StrykerJS | StrykerJS |

**Substitution Go: gremlins statt go-mutesting.** Plan-DoD §5-1
nennt `go-mutesting` als ursprünglichen Vorschlag. Begründung
für die Substitution:

- `go-mutesting` ist seit ~2022 unmaintained; AST-Brüche auf
  Go 1.21+ wurden gemeldet, ohne Fix.
- gremlins ist die aktiv gepflegte Alternative (Coverage-aware
  Mutator-Selektion, JSON-Output, Go-1.21+-Support).
- gremlins-Output ist Maschinen-lesbar (JSON), passt zum Trend-
  Tracking aus `extra-gates.md` §3.6 DoD ("Ergebnis als Trend").

`go-mutesting` bleibt als Fallback im Hinterkopf, falls gremlins
einen Bug zeigt; ein Wechsel ist eine Tool-Substitution mit
Plan-DoD-Update — nicht still im Make-Target.

## 2. Pilot-Module (Tranche-4-Stand)

| Sprache | File | Test-Surface |
|---|---|---|
| Go | `apps/api/hexagon/application/event_meta_validation.go` (343 LoC) | `event_meta_validation_internal_test.go` (337 LoC) + `event_meta_validation_fuzz_internal_test.go` (84 LoC, plan-0.9.5 §4) |
| TS | `packages/player-sdk/src/adapters/webrtc/sampling.ts` (248 LoC) | `tests/webrtc-adapter.test.ts` (823 LoC) + `tests/tracker.test.ts` (269 LoC) |

**Auswahl-Begründung**: beide Module sind sicherheits-relevant
(Event-Meta-Validation entscheidet, was als reservierter
Namespace durchgeht; Sampling extrahiert Telemetrie aus
`getStats()` und mappt direkt in den Wire-Vertrag). Beide haben
substantielle Test-Surfaces, sodass die Mutation-Score
aussagekräftig wird (nicht von trivial-niedriger Coverage
verzerrt).

`extra-gates.md` §3.6 listet weitere Kandidaten (Cursor-Logik,
HLS/DASH-Parser, SRT-Health-Mapping, SSRF-Prüfung). Diese sind
**nicht** Teil von Tranche 4 — Aufnahme erfolgt in Folge-Plänen
(`plan-0.10.x` o. ä.) nach Auswertung der ersten Beobachtungsläufe.

## 3. Score-Schwelle und PR-Blockierungs-Pfad

**Plan-DoD §5-4**: > 70 % Mutation-Score als Wunsch-Ziel; PR-
Blockierung erst, wenn das Modul die Schwelle drei
Beobachtungsläufe in Folge erreicht.

| Schwelle | Bedeutung |
|---|---|
| < 60 % | Test-Suite hat Lücken; Mutation-Killing-Quote zu niedrig. Folge-Backlog-Item. |
| 60 % – 70 % | Beobachtungsphase-Mittelfeld; weitere Tests gewünscht, aber keine Pflicht. |
| > 70 % | Modul reif für PR-Blockierung — nach **drei** Nightly-Runs in Folge mit > 70 %. |
| > 80 % | Stryker `thresholds.high`; gilt als „grün". |

**Übergangs-Mechanik** (analog zur Benchmark-Beobachtungsphase
aus `plan-0.9.5` §2 Tranche 1c):

Status `0.22.0`: Die TS-Nightly-Filterung wurde auf den aktuellen
GitHub-Packages-Scope `@pt9912/player-sdk` korrigiert. Frühere
grüne Workflow-Ergebnisse belegen deshalb nur, dass der nicht-
blockierende Workflow lief; sie belegen noch keine stabile
Mutation-Score-Reihe für den TS-Piloten.

1. Drei aufeinanderfolgende Nightly-`mutation.yml`-Runs zeigen
   > 70 % für ein Modul.
2. Folge-Commit nimmt das Modul aus `continue-on-error: true`
   raus und setzt eine `--threshold-break=70`-Variante (Workflow
   wird pro Modul rot, nicht ganzer Job-Failure).
3. Plan-DoD-Update im jeweiligen Folge-Plan markiert das Modul
   als „PR-Blockierungs-aktiv".

Score-Senkungen oder Scope-Änderungen sind begründungspflichtig
(`extra-gates.md` §3.6 DoD).

## 4. Lokale Reproduktion

### 4.1 Beide Module auf einmal

```bash
make mutation-report      # opt-in, NICHT Teil von make gates
```

Output:

- Go: `apps/api/.tmp/mutation/api-mutation-report.txt`
  + `.json`.
- TS:
  `packages/player-sdk/reports/mutation/mutation.html`
  + `mutation.json`.

### 4.2 Nur Go (gremlins)

```bash
make api-mutation-report
# oder mit alter gremlins-Version pinnen:
GREMLINS_VERSION=v0.5.0 make api-mutation-report
```

Läuft in einem `golang:1.26`-Container; das gremlins-CLI wird
per `go install` zur Laufzeit gezogen — kein Eintrag in `go.mod`.

### 4.3 Nur TS (StrykerJS)

```bash
make ts-mutation-report
```

Läuft in Docker
([ADR-0008](../adr/0008-benchmark-mutation-execution-in-docker.md),
analog zum Go-Pendant im `golang`-Container): der `mutation-ts`-Stage
(= `build` + `procps`) wird gebaut, `docker run` fährt StrykerJS, und
der Report wird per Bind-Mount auf den Host gespiegelt. StrykerJS +
`@stryker-mutator/vitest-runner` sind als **exakte devDeps** (`9.6.1`)
gepinnt (Vitest-Runner nutzt dieselbe Vitest-Version wie `make ts-test`).

> **Warum devDeps statt `pnpm dlx`**: unter pnpm 11 macht der
> isolierte Store-Linker Strykers `import('typescript')` UND das
> vitest-Runner-Plugin unauflösbar (`ERR_MODULE_NOT_FOUND`) — beide
> keine deklarierten stryker-core-Deps. Das legte den Gate lange still
> lahm (maskiert durch `continue-on-error` + `\|\| true` im Nightly).
> devDeps + `.pnpm`-Hoisting lösen `typescript`; `plugins: [...]` in
> `stryker.conf.cjs` deklariert den Runner explizit; `procps` im
> Container deckt Strykers `ps`-Worker-Verwaltung. Siehe `R-31`
> (aufgelöst) und ADR-0008.

HTML-Report öffnen:

```bash
xdg-open packages/player-sdk/reports/mutation/mutation.html
# oder im CI: Workflow-Run → Artifact `mutation-ts-<run-id>`
```

## 5. Wartung

- **Modul aufnehmen**: neue Pilot-Module werden in §2 dieser
  Doku eingetragen, plus passendes Make-Sub-Target. Auswahl
  folgt der `extra-gates.md` §3.6-Kandidatenliste.
- **Tool-Substitution**: jede Tool-Änderung (gremlins-Major,
  Stryker-Major, Wechsel auf go-mutesting o. ä.) ist ein Plan-
  DoD-Item, kein stiller Make-Target-Patch.
- **Trend-Tracking**: die Nightly-Reports (`mutation.json`) sind
  die Quelle für den Trend. Manuelle Auswertung über
  `gh run download <run-id> --name mutation-go-<run-id>` /
  `mutation-ts-<run-id>`. Automatisches Trend-Diagramm ist
  Folge-Backlog-Item.
- **Quarantäne-Pfad**: ein flaky Mutation-Test (z. B. Stryker
  Test-Runner-Race) wird **nicht** wie ein Benchmark in
  Quarantäne gelegt — Mutation-Tests sind nicht-blockierend, der
  flaky Lauf rauscht im Trend einfach durch. Sollte gremlins/
  Stryker selbst persistente Crashes zeigen, wandert das in
  `docs/planning/in-progress/risks-backlog.md` als R-N-Eintrag plus
  Tool-Update-Trigger.
