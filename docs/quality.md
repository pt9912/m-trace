# Qualität

## Zweck

Dieses Dokument beschreibt die verbindlichen Qualitäts- und Mess­pfade
für m-trace. Es dokumentiert keine Review-Historie und keine
einmaligen Befunde, sondern die aktuell gültigen, **reproduzierbaren
Docker-Prüfwege**. Aktuelle Werte (Image-Größe, Cold-Start, Test-
Anzahl) leben in `docs/spike/backend-stack-results.md`; dieses Doc
fixiert *wie* gemessen wird, nicht *was die letzte Messung ergab*.

Bezug: `docs/plan-spike.md` §14.9 (Linting), §14.11 (Dockerfile-
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

Die Stage führt `golangci-lint run ./...` mit Default-Lintern aus:

| Linter        | Zweck                                            |
| ------------- | ------------------------------------------------ |
| `govet`       | semantische Korrektheit (z. B. printf-Argumente) |
| `errcheck`    | Fehlerwerte werden nicht ignoriert               |
| `staticcheck` | Klassische Bug-Patterns + Style                  |
| `unused`      | toter Code                                       |
| `ineffassign` | unwirksame Zuweisungen                           |

Custom-Regeln, Suppressions und Toolchain-Ausbau sind im Spike-Scope
ausgeschlossen — Default-Profil pur (Plan §14.9). Verstöße brechen
den Build.

Make-Target (Soll, Plan §14.9):

```bash
make lint    # = docker build --target lint
```

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
  `docs/spike/backend-api-contract.md`.

Make-Target:

```bash
make test    # = docker build --target test
```

Kein Tag-System. Keine Test-Categorization über
Build-Argumente. Single-Modul, alle Tests laufen in einem Lauf.
Test-Output und Reports leben in den jeweiligen
`build/test-results/`-Verzeichnissen im Container.

---

## 3. Coverage (Platzhalter)

Coverage-Instrumentierung ist nicht im aktuellen `Dockerfile` aktiviert.
Der ADR (`docs/adr/0001-backend-stack.md` §8) sieht für `0.1.0+` eine
Coverage-Strategie vor:

- `go test -cover` (Standardbibliothek-nativ)
- Aggregierter Threshold im Build-Gate (Vorschlag: 90% Line-
  Coverage auf `apps/api/hexagon/` und `apps/api/adapters/`)
- HTML-/JSON-Report als CI-Artifact

Konkrete Ausgestaltung wird in einer Folge-ADR festgelegt
(siehe `docs/roadmap.md` §4 — *Coverage-Tooling für Go*). Bis dahin
existiert dieser Pfad nicht; das Gate ist `Tests grün` über §2.

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

Verbindlich aus `docs/spike/backend-api-contract.md`:

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

## 6. CI-Pipeline (Platzhalter)

Eine CI-Pipeline existiert noch nicht. Geplant ab `0.1.0` (siehe
`docs/roadmap.md` §2 Schritt 1 ff.):

- GitHub Actions Workflow `build.yml` mit Steps für
  `make test`, `make lint`, `docker build --target runtime`
- Test-, Coverage- und Lint-Reports als Workflow-Artifacts
  (Pattern aus `d-migrate/.github/workflows/build.yml`)
- Cache: `gradle/actions/setup-*` analog für Go über
  `actions/cache` auf `~/go/pkg/mod` (oder Docker-Layer-Cache
  direkt)

Bis CI existiert ist der lokale Docker-Pfad maßgeblich und
identisch zu dem, was die spätere CI ausführen wird (Plan §14.11
Docker-only-Workflow).

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
