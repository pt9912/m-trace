# 0008 — Ausführungsort der Bench-/Mutation-Gates: Docker

> **Status**: **Accepted** (2026-07-16); **TS-Bench-Gate geliefert**,
> **TS-Mutation-Gate zurückgestellt** (R-31). `make analyzer-benchmark-smoke`
> läuft jetzt im `build`-Stage-Container (`docker run`), analog zum
> schon-containerisierten Go-Bench (`api-benchmark-smoke` in
> `golang:1.26.5`); die Go-Gates (`api-benchmark-smoke` /
> `api-mutation-report`) waren bereits Docker. Der `ts-mutation-report`-
> Umzug ist **zurückgestellt (R-31)**: StrykerJS via `pnpm dlx` löst im
> Container das Workspace-`typescript` nicht auf (`ERR_MODULE_NOT_FOUND`)
> — eigener Fix nötig, bleibt bis dahin host-seitig. Der Beschluss ist
> **mess-basiert**, nicht plausibilitäts-basiert: eine A/B-Messung Host
> vs. Docker (5× je Seite, interleaved) belegt, dass der
> Container-Overhead für die CPU-gebundenen Hot-Path-Benches innerhalb
> des Mess-Rauschens (≈ 0) liegt und alle Wall-Clock-Budgets aus
> [`docs/perf/budgets.md`](../perf/budgets.md) §4 halten.
> **Datum**: 2026-07-16 (Proposed + Accepted + geliefert)
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)
> **Bezug**: [`docs/perf/budgets.md`](../perf/budgets.md) §2/§3/§4
> (Wall-Clock-Budgets, `make gates`-blockierend);
> [`docs/dev/mutation-testing.md`](../dev/mutation-testing.md)
> (Mutation-Gates, opt-in); RAK-18 (Performance-Budget, normativ im
> Lastenheft); [ADR-0005](0005-production-ops-backends.md)-Trigger-#3
> (Bench als Evidenzquelle).

## Kontext

Das Repo hat vier Performance-/Qualitäts-Gates mit **uneinheitlichem
Ausführungsort**:

| Gate | Sprache | Ausführungsort (vorher) |
| --- | --- | --- |
| `api-benchmark-smoke` | Go | **Docker** (`golang:1.26.5`, gemounteter Source) |
| `api-mutation-report` | Go | **Docker** (`golang:1.26.5`, gremlins via `go install`) |
| `analyzer-benchmark-smoke` | TS | **Host** (`pnpm --filter … run bench`) |
| `ts-mutation-report` | TS | **Host** (`pnpm --filter … run mutation`, StrykerJS via `pnpm dlx`) |

Der Ausführungsort war **nirgends normativ** festgelegt — nur das
Budget selbst (RAK-18, [`docs/perf/budgets.md`](../perf/budgets.md)) ist
normativ. Die Host-Wahl der zwei TS-Gates stand nur implizit im
Makefile.

Die Host-Ausführung bringt zwei betriebliche Kosten:

1. Sie braucht **lokale `node_modules`** (`make host-deps`) und in CI
   einen `pnpm install`-Schritt — eine zweite Dependency-Installations-
   Achse neben den Docker-Layern.
2. Seit dem pnpm-10→11-Bump reißt der Host-`node_modules`-Zustand bei
   Versionswechsel (`.modules.yaml`-Purge-Falle,
   `ERR_PNPM_ABORTED_REMOVE_MODULES_DIR_NO_TTY`) — eine Sonderbehandlung,
   die die Docker-Layer nicht kennen.

„Alles in Docker" würde beide Kosten eliminieren und den Ausführungsort
über alle vier Gates vereinheitlichen. Die **offene Kernfrage** war
allein: halten die Wall-Clock-Budgets aus
[`docs/perf/budgets.md`](../perf/budgets.md) §4 auch **unter Docker**,
oder macht Container-Overhead/Noise die Budget-Checks flaky? Diese Frage
darf nicht durch Plausibilitätsannahme entschieden werden, sondern durch
Messung.

## Evidenz — A/B-Messung Host vs. Docker (2026-07-16)

Setup: `analyzer-benchmark-smoke` (7 Hot-Paths) je 5× auf dem Host und
5× im `build`-Stage-Container (`docker run m-trace-ts:build`),
**interleaved** (gegen Drift), sonst idle. Host `node v22.18.0` /
Container `node v22.23.1` (beide pnpm 11.13.0), 20-Kern i9-13900H.
`mean`/`p99` in ms; der Budget-Gate prüft `p99` (dann `mean`) je Lauf,
maßgeblich ist also der **worst-case p99 über die Läufe**.

| Hot-Path | Budget | Host mean~ | Docker mean~ | Host worst-p99 | Docker worst-p99 | D/H p99 (median) | Headroom |
| --- | --- | --- | --- | --- | --- | --- | --- |
| HLS Master klein | 5 ms | 0,009 | 0,009 | ~0,030 | 0,030 | 1,02× | 169× |
| HLS Master groß | 25 ms | 0,154 | 0,148 | ~0,34 | 0,373 | 0,79× | 67× |
| HLS Media (1k Seg) | 50 ms | 0,767 | 0,760 | 1,505 | 1,404 | 0,83× | 36× |
| DASH VOD | 5 ms | 0,030 | 0,031 | ~0,05 | 0,067 | 0,97× | 74× |
| DASH Live | 10 ms | 0,039 | 0,040 | ~0,06 | 0,094 | 1,16× | 106× |
| **Detector 256K** (engster) | **0,5 ms** | 0,124 | 0,123 | 0,266 | **0,268** | 0,86× | **1,9×** |
| SSRF (100 Calls) | 5 ms | 0,004 | 0,004 | ~0,007 | 0,007 | 0,86× | 725× |

**Befund:**

- **Kein Budget-Bruch** in irgendeinem der 5 Docker-Läufe. Engster
  Fall: Detector, Docker-worst-p99 **0,268 ms vs. 0,5 ms = 1,9×
  Headroom** — und host-seitig ist der Worst-Case mit 0,266 ms praktisch
  gleich.
- **Container-Overhead ≈ 0.** Die D/H-p99-Mediane liegen bei
  **0,79–1,16×**, streuen also um 1,0 — Docker ist mal minimal schneller,
  mal langsamer, alles innerhalb des Mess-Rauschens. Der p99-Jitter
  (GC/Scheduler) ist **nicht Docker-spezifisch**: der Detector-Worst-Case
  tritt host-seitig (0,266 ms) und im Container (0,268 ms) gleich auf.
- **Physikalisch erwartbar:** Container teilen den Host-Kernel;
  CPU-gebundene In-Memory-Parserei läuft nativ. Container-Overhead lebt
  in IO/Netzwerk/Syscalls — die haben diese Benches nicht (Fixtures sind
  In-Memory-Strings, kein IO).

Stützend: der **Go-Bench läuft bereits in Docker** und die §3-Budgets
halten in CI grün — Docker-Budget-Validierung ist für den Go-Pfad schon
Existenzbeweis. Die einzige Achse, die das A/B nicht abbildet, ist
Dev-Maschine vs. CI-Runner (`ubuntu-24.04`, schwächer); übertragbar ist
aber die **Delta**-Größe (Host vs. Docker, gleiche Maschine), und die ist
~0. Die absoluten Budgets sind ohnehin CI-kalibriert-großzügig
([`docs/perf/budgets.md`](../perf/budgets.md) §2).

## Entscheidung

> **Entscheidung (Accepted 2026-07-16):** Der **Zielzustand für den
> Ausführungsort der Bench-/Mutation-Gates ist Docker** — für alle vier
> Gates. Geliefert: `analyzer-benchmark-smoke` läuft im `build`-Stage-
> Image via `docker run` (die Go-Gates waren schon Docker).
> `ts-mutation-report` folgt demselben Zielzustand, ist aber
> zurückgestellt (R-31, s. u.). Der Budget-Check
> (`scripts/check-bench-budgets.mjs`, reines Node-builtins) bleibt auf
> dem Host, weil er die vom Container erzeugte Ausgabe liest.

Mechanik:

- **Bench** (`analyzer-benchmark-smoke`, geliefert): `docker run
  m-trace-ts:build` schreibt die vitest-Bench-Tabelle auf stdout; `tee`
  spiegelt sie nach `.tmp/bench/analyzer-bench.txt` (Host-Artefakt, das
  `benchmark-observation.yml` als Trend hochlädt); der Budget-Check läuft
  auf dem Host. **Exakt** das Muster des schon-Dockerisierten Go-Benches
  (`docker run golang:1.26.5 … | tee .tmp/bench/api-bench.txt`).
- **Mutation** (`ts-mutation-report`): **zurückgestellt (R-31).** Der
  geplante Weg — `docker run m-trace-ts:build` fährt StrykerJS, Report-
  Ordner per **Bind-Mount** (`-v …/reports:/workspace/…/reports`) auf den
  Host (kein stdout) — scheitert daran, dass StrykerJS via `pnpm dlx` das
  Workspace-`typescript` im Container nicht auflöst
  (`ts-config-preprocessor` → `ERR_MODULE_NOT_FOUND: Cannot find package
  'typescript'`). Kandidat-Fix: `typescript` in die `pnpm dlx
  --package`-Liste, plus Nachweis, dass Stryker im Container vollständig
  durchläuft UND der Host-Pfad heil bleibt (der könnte durch
  `continue-on-error` + `|| true` bislang stillschweigend fehlschlagen).
  `ts-mutation-report` bleibt bis dahin host-seitig.

**Kein dedizierter `bench`-Dockerfile-Stage.** Der `build`-Stage enthält
bereits alles, was Bench/Mutation brauchen (node_modules + gebautes
`dist` + `benchmarks/` + vitest); ein eigener `RUN`-Stage wäre entweder
ein inhaltsloser Alias von `build`, oder würde — mit Budget-Check *im*
Container — den Host-Artefakt zerstören, den der Beobachtungs-Nightly
hochlädt. `docker run` des `build`-Images erzeugt den Artefakt direkt und
bleibt symmetrisch zum Go-Bench.

## Konsequenzen

**Positiv:**

- Einheitlicher Ausführungsort für die Bench-Gates (Go + TS) und die
  Go-Mutation; TS-Mutation folgt mit R-31.
- `make analyzer-benchmark-smoke` braucht **keine Host-`node_modules`**
  mehr → in `benchmark-observation.yml` entfällt der `pnpm install`-
  Schritt; die pnpm-`.modules.yaml`-Purge-Falle beim Toolchain-Bump
  betrifft dieses Gate nicht mehr. `ts-mutation-report` bleibt bis zum
  R-31-Fix host-seitig (inkl. `pnpm install` in `mutation.yml`).
- Reproduzierbarkeit: der Bench läuft gegen die exakt gebaute Artefakt-
  Kette (`build`-Image) statt gegen einen möglicherweise driftenden
  Host-`node_modules`-Zustand.

**Kosten / Grenzen (ehrlich benannt):**

- Jeder Lauf baut zuerst den `build`-Stage (`docker build --target
  build`) — in `make gates` und CI sind die Layer ohnehin gecacht (andere
  TS-Stages bauen sie), der Zusatzaufwand ist damit ein Cache-Hit.
- Der Container-`node` (v22.23.1) weicht minimal vom Host-`node`
  (v22.18.0) ab; beide sind node 22, für CPU-gebundene Benches
  irrelevant (die A/B-Zahlen streuen um 1,0). `print-bench-runner-info.sh`
  druckt weiter Host-Info; die CPU ist geteilt (kernel-nah), also
  repräsentativ.
- **host-deps entfällt NICHT vollständig**: `package-publish-dry-run`
  (release-zeitig, opt-in, host-`tsup`) bleibt host-seitig — außerhalb
  des Scopes dieses ADR (Bench/Mutation-Gates). Ein späterer Umzug ist
  ein eigener Schritt.
- Beim späteren ts-mutation-Umzug (R-31) werden die Reports vom
  Container (root) via Bind-Mount geschrieben → root-owned auf dem Host,
  konsistent mit dem schon bestehenden Go-Mutation-Verhalten
  (`-v $(CURDIR):/src` in `golang`). Für den jetzt gelieferten Bench
  entfällt das (stdout/tee statt Verzeichnis-Output).

**Budget-Wartung unverändert:** die Schwellen aus
[`docs/perf/budgets.md`](../perf/budgets.md) bleiben, was sie sind —
**keine Anhebung nötig** (die Messung deckt die Docker-Ausführung ab).
§2 der Budget-Datei nennt Docker jetzt explizit als Ausführungsort.

## Alternativen

- **Status quo (TS-Gates host)** — verworfen: hält die uneinheitliche
  Ausführungsort-Landschaft und die Host-`node_modules`-Achse inkl.
  `.modules.yaml`-Purge-Falle am Leben.
- **Dedizierter `bench`-Dockerfile-Stage (`docker build --target
  bench`)** — verworfen: kein eigener Inhalt gegenüber `build`, und der
  Budget-Check im Container ließe die Bench-Zahlen nur im Build-Log statt
  im Host-Artefakt, das `benchmark-observation.yml` hochlädt. Extra
  Extraktions-Plumbing (`docker cp`/`--output`) ohne Gewinn.
- **Budgets pauschal anheben, um Docker-Noise abzufedern** — verworfen:
  die Messung zeigt, dass es keinen systematischen Docker-Noise gibt;
  eine Anhebung würde echte Regressionen maskieren.
