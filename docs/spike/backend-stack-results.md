# Backend-Stack-Spike — Spike-Protokoll

> **Living document auf `main`.** Notizen während AP-1 (Go) und AP-2
> (Kotlin/Micronaut) werden direkt hierher committed; AP-3 verdichtet sie
> zum ADR `docs/adr/0001-backend-stack.md`.
>
> **Bezug**: `docs/plan-spike.md` (§4.1, §6.2, §6.3, §7.3, §10),
> `docs/spike/0001-backend-stack.md` (§9 Bewertungsraster).

---

## 1. Vertrags-/Spec-Änderungen

> Nach dem Merge von `docs/spike/backend-api-contract.md` nach `main` ist
> der Kontrakt eingefroren. Jede Änderung **muss** hier dokumentiert sein
> (Plan §4.1) und in beiden Prototypen identisch landen.

| Datum | Änderung | Begründung | Commit-Hash |
|---|---|---|---|
| _–_ | keine bisher | _–_ | _–_ |

**Beobachtung (keine Änderung)**: Bei der AP-1-Implementierung fiel
auf, dass §5 step-3 (Rate-Limit) und step-5 (Batch-Size > 100)
einander maskieren können: ein 101-Event-Batch verbraucht 101 Tokens
aus einem 100-Token-Bucket und triggert dadurch step-3 (`429`), bevor
step-5 (`422`) greifen kann. Der Test `TestHTTP_422_TooManyEvents` im
Go-Prototyp umgeht das mit einem `unlimitedLimiter` als Test-Fixture
(Kommentar im Testcode). Der Vertrag bleibt unverändert; AP-2 sollte
denselben Test-Workaround nutzen, damit beide Prototypen die §11-
Pflichttests symmetrisch grün bekommen.

---

## 2. AP-1: Go-Prototyp

**Branch**: `spike/go-api`  
**Start**: 2026-04-28  
**Ende**: 2026-04-28  
**Final-Commit**: `7148a8d` (HTTP integration tests)

### 2.1 Notizen pro Bewertungskategorie

#### Time to Running Endpoint
Bootstrap (`/api/health` → 200 via Docker) lief ab dem leeren Branch
in zwei Iterationen: erstes `docker build --target compile` brauchte
nur `go.mod` + `cmd/api/main.go` mit `net/http`-Default und einem
`HealthHandler`. Die `golangci/golangci-lint`-Stage musste eine kleine
Korrektur am `deps`-Stage haben (`mkdir -p $GOMODCACHE`), damit der
`COPY --from=deps`-Schritt nicht failt, wenn keine externen Deps
existieren. Danach glatt.

#### OTel-Integration-Ergonomie
Sehr leichter Setup (`go.opentelemetry.io/otel`, `sdk/metric`,
`semconv/v1.26.0`). Eine Stolperfalle: `semconv/v1.26.0` ist als
Sub-Pfad im selben Modul unterwegs, *darf nicht* separat im
`require`-Block stehen — `go mod tidy` wirft sonst „malformed module
path". Sobald der direkte `require` entfernt war, hat `tidy` alles
sauber aufgelöst. Kein Exporter konfiguriert (Spec §6.7 erlaubt das);
ein OTLP-Wechsel wäre eine Konfigurationszeile. Die globale
`otel.SetMeterProvider`-API fühlt sich für eine kleine Anwendung
natürlich an.

#### Hexagon-Fit
Sehr gut — Go-Pakete als Schicht-Boundary funktionieren idiomatisch:
`hexagon/domain` hat null externe Imports, `hexagon/port/{driving,
driven}` definiert die Verträge, `hexagon/application` orchestriert,
und alle Adapter unter `adapters/{driving,driven}/...` importieren
nur Ports + Domain. Compile-Time-Enforcement ist *nicht* möglich
(Single-Modul), aber der Lint-Run plus Code-Review reicht für den
Spike-Scope. Die `var _ Interface = (*Impl)(nil)`-Compile-Time-
Checks an jedem Adapter machen die Port-Implementierung explizit.

#### Test-Velocity
Stdlib `testing` + `httptest` sind exzellent für diesen Stack. 24
Tests in **0,5 s** Wallclock (10 Application-Unit + 4 RateLimiter +
10 HTTP-Integration). `t.Parallel()` funktioniert ohne Reibung,
Test-Helpers (`newTestServer`, `postEvents`) bleiben kurz. Kein
externes Mock-Framework nötig — kleine Stubs/Spies pro Testfall
reichen.

#### Docker Image Size
Final runtime-Image: **10,2 MB** (`gcr.io/distroless/static-debian12:
nonroot`). Für einen Spike sehr sauber; produktiv tauglich.

#### Cold Start
Erster `200 OK` auf `/api/health` nach `docker run`: **9 ms**
(via Curl-Loop mit 50 ms Polling-Intervall). Das ist Distroless +
statisch gelinkter Go-Binary, ohne JVM-Warmup. Für Latenz-kritische
Path-Bewertung auffällig gut.

#### Build-Komplexität
Eine `Dockerfile` mit sechs Stages (`deps → compile → lint → test →
build → runtime`); `Makefile` mit fünf Targets (`test`, `lint`,
`build`, `deps`, `compile`, `run`, `clean`). `.dockerignore` sperrt
`.git/` und IDE-State. `go.mod` mit fünf direkten Deps. Keine
Wrapper-Skripte, kein Plugin-Stack zu pflegen.

#### Subjektiver Spaß
Hoch. Go 1.22 method-aware Mux (`mux.Handle("POST /api/...", h)`)
hält Routing übersichtlich; `log/slog` JSON-Logs sind out-of-the-box
strukturiert; `http.MaxBytesReader` macht den 256-KB-Body-Check zum
Einzeiler. Wenig Zauberei, klarer Datenfluss.

#### Contributor-Fit
Sehr hoch im Streaming-/Observability-Umfeld. MediaMTX, OTel
Collector, Prometheus, Grafana, viele HLS/WebRTC-Tools sind in Go.
Standard-Toolchain (`go test`, `golangci-lint`) ist universell
bekannt, Onboarding für externe Mitwirkende ist niedrigschwellig.

#### Absehbare Phase-2-Risiken
- **Persistenz-Wechsel**: In-Memory → DB (z. B. SQLite/PostgreSQL)
  braucht eine echte `EventRepository`-Implementation; Port-Vertrag
  scheint stabil, aber Transaktions-/Batch-Semantik wird Detail.
- **Hexagon-Boundaries**: Ohne Multi-Modul-Enforcement gibt es kein
  Compile-Time-Schutz gegen falsche Imports. Bei wachsendem Team
  reicht Lint-Konvention vielleicht nicht. Workspace-Aufteilung mit
  `internal/`-Unterprojekten oder `go.work` ist denkbar.
- **WebSocket/SSE für Live-Updates**: Stdlib `net/http` hat SSE-
  Grundlagen, WebSocket via `gorilla/websocket` oder `nhooyr/websocket`
  — Ökosystem leicht fragmentiert.
- **SRT/Streaming-Protokoll-Integration**: Go-Bindings für libsrt
  existieren (`srtgo`), aber CGO-Pflicht bricht das Distroless-Setup.

### 2.2 Bonus-Scope (umgesetzt?)

- [ ] `GET /api/stream-sessions`
- [ ] `GET /api/stream-sessions/{id}`
- [ ] Origin-Allowlist pro Project
- [ ] expliziter Session-Lifecycle `ended`
- [ ] OTel-Counter zusätzlich zu Prometheus
- [ ] `trace_id` in Logs
- [ ] vollständige Hexagon-Schichtung bis ins Detail

Keiner der Bonus-Punkte umgesetzt — Muss-Scope hatte Vorrang im
Zeitbudget.

### 2.3 Stolpersteine / Beobachtungen

- **`semconv/v1.26.0`-require**: siehe oben unter OTel-Ergonomie. War
  der einzige nicht-triviale Tooling-Stolperer.
- **`COPY --from=deps /go/pkg/mod`** im Lint-Stage scheitert, wenn
  `go mod download` keine externen Deps gefunden hat (Verzeichnis
  existiert dann nicht). Fix: `mkdir -p "$GOMODCACHE"` vor dem
  `download` im `deps`-Stage.
- **§5 step-3 vs. step-5-Maskierung**: 101-Event-Batch greift step-3
  Rate-Limit, bevor step-5 (Batch-Size) prüft. Test-Workaround
  dokumentiert (`unlimitedLimiter`); Vertrag unverändert (§1).

### 2.4 Mess-Werte (Stand `7148a8d`)

| Metrik | Wert |
|---|---|
| Wallclock bis erster grüner Test | ~5 min nach Bootstrap |
| Wallclock bis erstes `docker run` mit 200 OK | ~5 min nach Bootstrap |
| LoC `hexagon/domain/` | 77 |
| LoC `hexagon/application/` + `port/` | 255 |
| LoC `adapters/` (ohne Tests) | 508 |
| LoC Tests | 520 |
| LoC Total Production | 942 |
| Final Docker Image Size (`runtime`) | 10,2 MB |
| Cold Start (`/api/health` → 200) | 9 ms |
| Direkte Dependencies (`go.mod` `require` ohne `// indirect`) | 5 |
| Testlaufzeit (`go test ./...`) | ~0,5 s |
| Konfigurationsdateien | `Dockerfile`, `Makefile`, `.dockerignore`, `go.mod` |

---

## 3. AP-2: Kotlin-/Micronaut-Prototyp

**Branch**: `spike/micronaut-api`  
**Start**: _YYYY-MM-DD_  
**Ende**: _YYYY-MM-DD_  
**Final-Commit**: _<hash>_

### 3.1 Notizen pro Bewertungskategorie

#### Time to Running Endpoint
_–_

#### OTel-Integration-Ergonomie
_–_

#### Hexagon-Fit
_–_

#### Test-Velocity
_–_

#### Docker Image Size
_–_

#### Cold Start
_–_

#### Build-Komplexität
_–_

#### Subjektiver Spaß
_–_

#### Contributor-Fit
_–_

#### Absehbare Phase-2-Risiken
_–_

### 3.2 Bonus-Scope (umgesetzt?)

- [ ] `GET /api/stream-sessions`
- [ ] `GET /api/stream-sessions/{id}`
- [ ] Origin-Allowlist pro Project
- [ ] expliziter Session-Lifecycle `ended`
- [ ] OTel-Counter zusätzlich zu Prometheus
- [ ] `trace_id` in Logs
- [ ] vollständige Hexagon-Schichtung bis ins Detail

### 3.3 Stolpersteine / Beobachtungen

_–_

---

## 4. Reihenfolge und Bias (Plan §4.4)

- Tatsächliche Reihenfolge: _Go zuerst / Micronaut zuerst_
- Beobachteter Bias: _–_
- Korrekturmaßnahmen während des Spikes: _–_

---

## 5. Subjektive Gesamteindrücke

_Freier Block für übergreifende Notizen, Aha-Momente, Frustpunkte. Wird
in AP-3 zu den ADR-Abschnitten "Bewertung" und "Konsequenzen" verdichtet._

---

## 6. Übergang zu AP-3

- Bewertungsbogen (Spec §16) ausgefüllt im ADR: ☐
- Messwertbogen (Spec §17) ausgefüllt im ADR: ☐
- Reihenfolge-Bias im ADR notiert: ☐
- Sieger-Branch markiert: ☐
- Unterlegener Branch gelöscht oder als Tag
  `spike/backend-stack-loser-YYYYMMDD` archiviert: ☐
- Finale Commit-Hashes beider Prototypen im ADR: ☐
