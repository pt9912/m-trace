# Qualität

## Zweck

Dieses Dokument beschreibt die verbindlichen Qualitäts- und Mess­pfade
für m-trace. Es dokumentiert keine Review-Historie und keine
einmaligen Befunde, sondern die aktuell gültigen, **reproduzierbaren
Docker-Prüfwege**. Aktuelle Werte (Image-Größe, Cold-Start, Test-
Anzahl) leben in `docs/spike/backend-stack-results.md`; dieses Doc
fixiert *wie* gemessen wird, nicht *was die letzte Messung ergab*.

Bezug: `docs/planning/done/plan-spike.md` §14.9 (Linting), §14.11 (Dockerfile-
Stages); `docs/spike/0001-backend-stack.md` §6.6, §6.11–§6.12; ADR
`docs/adr/0001-backend-stack.md` (Coverage-Konsequenzen).

---

## 1. Statische Analyse (`golangci-lint`)

Statische Analyse läuft Docker-basiert über die `lint`-Stage des
`apps/api/Dockerfile`:

```bash
cd apps/api
docker build --target lint -t m-trace-api-spike:lint .
```

Die Stage führt `golangci-lint run ./...` mit Default-Lintern und dem
im Lastenheft definierten SOLID-nahen Zusatzprofil aus:

| Linter        | Zweck                                            |
| ------------- | ------------------------------------------------ |
| `govet`       | semantische Korrektheit (z. B. printf-Argumente) |
| `errcheck`    | Fehlerwerte werden nicht ignoriert               |
| `staticcheck` | Klassische Bug-Patterns + Style                  |
| `unused`      | toter Code                                       |
| `ineffassign` | unwirksame Zuweisungen                           |

Die vollständige Konfiguration (5 Defaults + 24 SOLID-nahe Linter
aus §1.2) lebt in `apps/api/.golangci.yml`. „Pflicht: Ja" in
Lastenheft §10.1 fixiert die Profil-Mitgliedschaft, nicht den
Scope: ein Linter kann pro Pfad-Klasse (Tests, internes Tooling)
designseitig anders sinnvoll sein. `//nolint`-Suppressions bleiben
ausgeschlossen — falls ein Linter auf einem Pfad designseitig
keinen Sinn ergibt (z. B. `noctx` in Tests gegen `httptest.Server`),
wird der Pfad per `issues.exclude-rules` mit Begründung als Kommentar
ausgenommen; dort dokumentierte Scope-Definitionen sind keine
Suppressions, sondern bewusste Profil-Entscheidungen. Verstöße
brechen den Build.

Make-Target (Soll, Plan §14.9):

```bash
make lint    # = docker build --target lint
```

### 1.1 TypeScript/Svelte-Analyse

Die TypeScript- und Svelte-Pakete laufen über den Root-Workspace-Lint:

```bash
pnpm run lint
```

Der aktuelle Pfad ist bewusst leichtgewichtig:

| Paket/App | Gate | SOLID-naher Anteil |
|---|---|---|
| `packages/player-sdk` | `tsc --noEmit`, Boundary-Check, Public-API-Snapshot | `core/` darf nicht von Browser-Adaptern oder `hls.js` abhängen |
| `packages/stream-analyzer` | `tsc --noEmit`, Boundary-Check, Public-API-Snapshot | Public Modules dürfen nicht direkt aus `internal/` importieren |
| `apps/dashboard` | `svelte-check` | noch kein SOLID-nahes Zusatzprofil |
| `apps/analyzer-service` | `tsc --noEmit` | noch kein SOLID-nahes Zusatzprofil |

Soll-Ausbau: alle TypeScript-/Svelte-Pakete erhalten ein gemeinsames
SOLID-nahes Zusatzprofil für Import-Boundaries, verbotene Deep Imports,
Komplexität, Funktionslänge, Verschachtelung und stabile Public APIs.
Das Profil muss über `make lint` laufen und darf keine lokalen Editor-
Plugins voraussetzen.

### 1.2 Go: SOLID-nahe Linter

Die folgenden `golangci-lint`-Linter sind keine offizielle
SOLID-Kategorie. Sie sind die verbindliche Projektauswahl für
SOLID-nahe Designsignale: geringe Komplexität und kleine
Verantwortlichkeiten (SRP), schlanke Interfaces (ISP), stabile
Import-/Modulgrenzen (DIP) oder reduzierte globale Kopplung.

| Linter | Kurzbeschreibung | Teil von SOLID |
|---|---|---|
| `containedctx` | `context.Context` nicht in Structs speichern | Y |
| `contextcheck` | Context korrekt weiterreichen | Y |
| `cyclop` | Zyklomatische Komplexität | Y |
| `depguard` | Import-Regeln/Layer-Grenzen | Y |
| `dupl` | Code-Duplikate | Y |
| `fatcontext` | Context in Loops/Closures | Y |
| `forbidigo` | Verbotene Identifier/APIs | Y |
| `funlen` | Zu lange Funktionen | Y |
| `gochecknoglobals` | Keine globalen Variablen | Y |
| `gochecknoinits` | Keine `init()`-Funktionen | Y |
| `gocognit` | Kognitive Komplexität | Y |
| `gocyclo` | Zyklomatische Komplexität | Y |
| `gomodguard` | Modul-Allow-/Blocklist | Y |
| `iface` | Interface-Pollution vermeiden | Y |
| `inamedparam` | Interface-Parameter benennen | Y |
| `interfacebloat` | Zu große Interfaces | Y |
| `ireturn` | Interfaces annehmen, konkrete Typen zurückgeben | Y |
| `maintidx` | Maintainability Index | Y |
| `nestif` | Tiefe `if`-Verschachtelung | Y |
| `noctx` | HTTP-Aufrufe ohne Context | Y |
| `reassign` | Package-Variablen nicht neu zuweisen | Y |
| `revive` | Konfigurierbarer Stil-/Design-Linter | Y |
| `testpackage` | Externe `_test`-Packages | Y |
| `unparam` | Ungenutzte Parameter | Y |

---

## 2. Tests

Tests laufen Docker-basiert über die `test`-Stage:

```bash
cd apps/api
docker build --target test -t m-trace-api-spike:test .
```

Die Stage führt `go test ./...` aus. Aktuell deckt der Spike-
Prototyp:

- Unit-Tests: `RegisterPlaybackEventBatch` (Application-Layer),
  `TokenBucketRateLimiter` (Driven-Adapter).
- Integrationstests: `POST /api/playback-events` mit
  `httptest.NewServer` für alle 10 §11-Pflichttests aus
  `spec/backend-api-contract.md`.

Make-Target:

```bash
make test    # = docker build --target test
```

Kein Tag-System. Keine Test-Categorization über
Build-Argumente. Single-Modul, alle Tests laufen in einem Lauf.
Test-Output und Reports leben in den jeweiligen
`build/test-results/`-Verzeichnissen im Container.

---

## 3. Coverage

Coverage ist pro Workspace-Bereich definiert. Harte Coverage-Gates
existieren aktuell für `apps/api`, `packages/player-sdk` und
`apps/dashboard`.

### 3.1 API (`apps/api`)

Die API-Coverage-Messung läuft Docker-basiert über die
`coverage`-Stage des `apps/api/Dockerfile`. Das Gate ist Pflicht-
Bestandteil des Build-Prozesses und bricht den Build, wenn die
Total-Line-Coverage den Threshold unterschreitet.

```bash
cd apps/api
make coverage-gate                 # Pflicht-Gate, Default-Threshold 90 %
make coverage-gate THRESHOLD=92    # Threshold ad-hoc anheben
make coverage-report               # Profil + HTML in build/coverage/ extrahieren
```

Der Coverage-Range ist bewusst auf `hexagon/` und `adapters/`
beschränkt (`-coverpkg=./hexagon/...,./adapters/...`); `cmd/api`
besteht aus Wiring, Signal-Handling und OTel-Setup und ist nicht
sinnvoll testbar — Tests dort wären Doppel-Verifikationen der
Standardbibliothek.

| Artefakt | Quelle | Form |
|---|---|---|
| `build/coverage/coverage.out`      | Go-Coverage-Profil (`-coverprofile`) | binär, von `go tool cover` lesbar |
| `build/coverage/coverage-func.txt` | per-Funktion-Report                  | Plain-Text, Last-Line trägt das Total |
| `build/coverage/coverage.html`     | gerenderte HTML-Übersicht            | öffnen mit Browser; CI-Artifact |

Der Threshold steht aktuell auf **90 %**, Ziel ist **>= 95 %**.
Begründung: ein Threshold, der mit der Realität gleichzieht, statt sie
zu führen, wird typischerweise gesenkt — also wird er von Anfang an
hoch gehalten, damit jede neue Code-Zeile mit ihrem Test einkommt.
Override jederzeit per `make coverage-gate THRESHOLD=…` möglich;
Senkung des Defaults ist eine ADR-pflichtige Entscheidung.

`docker build --no-cache-filter coverage` erzwingt die Re-Evaluation
der Coverage-Stage, ohne die `deps`-Layer zu verwerfen — das verhindert
das Stale-Cache-Maskieren von Test-Failures, das im Spike beobachtet
wurde. Make-Targets `test`, `lint`, `coverage-gate`, `coverage-report`
ziehen den Filter automatisch.

Skript: `apps/api/scripts/coverage-gate.sh <func-file> [<threshold>]`.
Lesen die Last-Line des `go tool cover -func`-Outputs und exitet
1 bei Unterschreitung, 2 bei Eingabe-Fehler.

### 3.2 Player-SDK (`packages/player-sdk`)

Das Player-SDK nutzt Vitest für Unit-Tests und `@vitest/coverage-v8`
für das reproduzierbare Coverage-Gate:

```bash
pnpm --filter @npm9912/player-sdk run test
pnpm --filter @npm9912/player-sdk run test:coverage
pnpm --filter @npm9912/player-sdk run performance:smoke
```

Der Coverage-Scope ist das produktive SDK unter
`packages/player-sdk/src/`; Testcode unter `packages/player-sdk/tests/`,
Build-Artefakte unter `packages/player-sdk/dist/` und Hilfsskripte
unter `packages/player-sdk/scripts/` gehören nicht in den Nenner.

Die Konfiguration steht in `packages/player-sdk/vitest.config.ts`.
Artefakte landen unter `packages/player-sdk/coverage/`:

| Artefakt | Form |
|---|---|
| `coverage-summary.json` | maschinenlesbare Summary |
| `lcov.info` | LCOV-Profil |
| `index.html` | HTML-Report |

Der Threshold ist verbindlich: Statements 90 %, Lines 90 %, Functions
90 %, Branches 90 %. Senkungen nach Einführung sind
begründungspflichtig.

Der Performance-Smoke ist kein Coverage-Gate. Er baut das SDK und prüft
das normative `0.2.0`-Budget: Bundle < 30 KiB gzip ohne hls.js,
Event-Verarbeitung < 5 ms pro Event, kein synchroner Netzwerkaufruf im
Hot Path sowie Queue-/Retry-Grenzen.

### 3.3 Dashboard (`apps/dashboard`)

Das Dashboard ist eine SvelteKit-App. Der aktuelle Qualitätspfad ist
Build, Svelte-Type-Check, Vitest-jsdom-Tests und Coverage:

```bash
pnpm --filter @npm9912/m-trace-dashboard run build
pnpm --filter @npm9912/m-trace-dashboard run check
pnpm --filter @npm9912/m-trace-dashboard run test
pnpm --filter @npm9912/m-trace-dashboard run test:coverage
```

Die Konfiguration steht in `apps/dashboard/vite.config.ts`. Der
Coverage-Scope ist `apps/dashboard/src/**/*.{ts,svelte}`; Testcode,
generierte SvelteKit-Dateien, Build-Artefakte und statische
Framework-Ausgabe gehören nicht in den Nenner.

Der Threshold ist verbindlich: Statements 90 %, Lines 90 %, Functions
90 %, Branches 90 %. Senkungen nach Einführung sind
begründungspflichtig.

### 3.4 Monorepo-Gate-Abgrenzung

`make coverage-gate` umfasst ab `0.2.0` das API-Gate, das
Player-SDK-Gate und das Dashboard-Gate. Die CI ruft weiterhin das
Root-Target auf und bekommt damit alle Coverage-Pfade.

---

## 4. Runtime-Image

Das Runtime-Image baut die `runtime`-Stage des Dockerfile. Es ist
Spike-gerecht gehärtet:

- Final-Image: `gcr.io/distroless/static-debian12:nonroot`
- statisch gelinktes Go-Binary (`CGO_ENABLED=0`)
- läuft als nicht-root User (`nonroot`)
- exposed Port: `8080`

Smoke-Test für Release- und Abnahmepfade (Plan §10 DoD,
**AK-1**, **AK-2**):

```bash
cd apps/api
docker build --target runtime -t m-trace-api-spike:go .
docker run --rm -p 8080:8080 m-trace-api-spike:go &
PID=$!
sleep 1
curl -fs http://localhost:8080/api/health
kill $PID
```

Erwartet: `HTTP 200` und `{"status":"ok"}` (Spec §6.1).

---

## 5. Validierungs- und Endpunkt-Verträge

Verbindlich aus `spec/backend-api-contract.md`:

| Vertrag                                          | Ort   | Gate                                                                                                   |
| ------------------------------------------------ | ----- | ------------------------------------------------------------------------------------------------------ |
| Endpunkt-Pfad/-Methode/-Statuscode               | §2    | HTTP-Integrationstests pro Statuscode (§11)                                                            |
| Wire-Format `events`-Batch                       | §3    | Unit-Test `RegisterPlaybackEventBatch` (Schema-Version, Pflichtfelder)                                 |
| Auth-Reihenfolge `X-MTrace-Token` ↔ `project_id` | §4–§5 | Drei `401`-Integrationstests (fehlend, falsch, mismatch)                                               |
| 9-Schritt-Validierungsreihenfolge                | §5    | Test-Pärchen mit `unlimitedLimiter`-Fixture für 422-too-many-events trotz step-3 Rate-Limit-Maskierung |
| Pflicht-Metriken `mtrace_*`                      | §7    | Smoke-Curl in §4 oben + automatischer Test (`go test`)                                                 |
| Rate-Limit + `Retry-After`                       | §6    | dedizierter 429-Integrationstest                                                                       |

Vertragsänderungen sind nur synchron in beiden Implementierungen
zulässig — bis `0.1.0` ist nur Go-Implementierung relevant — und
müssen im Spike-Protokoll
(`docs/spike/backend-stack-results.md` §1) eingetragen sein.

---

## 6. CI-Pipeline

`0.1.0` nutzt GitHub Actions auf `ubuntu-24.04` als CI-Zielplattform.
Workflow: `.github/workflows/build.yml`.

Der Workflow läuft auf Pull Requests und Pushes nach `main`:

- `make test`
- `make lint`
- `make coverage-gate`
- `make arch-check`
- `make build`

Die Root-Targets delegieren in `apps/api/` und nutzen denselben
Docker-only-Pfad wie die lokale Entwicklung. Test-, Coverage- und
Lint-Reports als Workflow-Artefakte bleiben eine spätere
Verbesserung; `0.1.0` erzwingt zunächst die Gates.

---

## 7. Release-Pipeline-Gates (Platzhalter)

Keine Release-Pipeline definiert. Wird mit dem ersten Tag
(`v0.1.0`) konkret. Erwartete Gates analog zum cmake-xray-Pattern:

- Tag-Validator (semver)
- reproduzierbares Linux-Archiv mit `SOURCE_DATE_EPOCH`
- OCI-Image-Idempotenz über `docker buildx imagetools inspect`
- Drei-Wege-Versionscheck (Tag ↔ Build-eingebrannte Version ↔
  Release-Asset-Metadaten)
- Asset-Allowlist gegen versehentliche Asset-Drift

Die konkrete Ausgestaltung kommt mit einer Folge-ADR vor `v0.1.0`.

---

## 8. Hinweise

- Dieses Dokument soll den **reproduzierbaren Ist-Stand**
  beschreiben. Veraltete Review-Notizen, Placeholder-Befunde
  einzelner Commits und einmalige Zwischenstaende gehören nicht
  hier hinein.
- Aktuelle Zahlen (LoC, Image-Größe, Cold-Start, Testlaufzeit)
  liegen in `docs/spike/backend-stack-results.md`. Dieses Doc
  fixiert *wie* gemessen wird, nicht *welcher Wert beim letzten
  Lauf herauskam*.
- Coverage-, CI- und Release-Gates sind explizit
  Platzhalter-Sektionen und werden mit dem `0.1.0`-Cycle gefüllt.
  Ergänzungen werden hier eingetragen, sobald die jeweilige
  Folge-ADR existiert.
