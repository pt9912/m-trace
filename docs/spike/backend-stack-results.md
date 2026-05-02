# Backend-Stack-Spike — Spike-Protokoll

> **Living document auf `main`.** Notizen während AP-1 (Go) und AP-2
> (Kotlin/Micronaut) werden direkt hierher committed; AP-3 verdichtet sie
> zum ADR `docs/adr/0001-backend-stack.md`.
>
> **Bezug**: `docs/planning/done/plan-spike.md` (§4.1, §6.2, §6.3, §7.3, §10),
> `docs/spike/0001-backend-stack.md` (§9 Bewertungsraster).

---

## 1. Vertrags-/Spec-Änderungen

> Zum Spike-Zeitpunkt war der damalige API-Kontrakt nach dem Merge nach
> `main` eingefroren. Jede Änderung musste hier dokumentiert sein
> (Plan §4.1) und in beiden Prototypen identisch landen. Der laufende
> API-Kontrakt wird heute unter `spec/backend-api-contract.md` gepflegt.

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
**Start**: 2026-04-28  
**Ende**: 2026-04-28  
**Final-Commit**: `7c8bc44` (HTTP integration tests)

### 3.1 Notizen pro Bewertungskategorie

#### Time to Running Endpoint
Bootstrap (`/api/health` → 200 via Docker) lief in zwei Iterationen: erst
ein leeres Kotlin/Gradle-Setup ohne Micronaut, dann die volle
Micronaut-Application-Plugin-Konfiguration. Hauptzeit-Kostenpunkt war das
erste Auflösen des Micronaut-BOM und der Plugin-Cache (~90 s im
deps-Stage). Iterations-Builds danach ~30–40 s pro `compile`-Stage.

#### OTel-Integration-Ergonomie
Solid, aber mit kleineren Stolpersteinen: `OpenTelemetrySdk.builder()` +
`SdkMeterProvider.builder()` + `Resource.merge()` ist mehr Boilerplate
als `otel.SetMeterProvider(mp)` in Go. Micronaut-Factory-Wiring (`@Factory
class OtelFactory`) macht den Lifecycle aber sauber. Der erste Versuch
mit `@Bean(preDestroy = "close")` schlug fehl, weil das `OpenTelemetry`-
Interface keine `close`-Methode hat (nur `OpenTelemetrySdk` über
`AutoCloseable`); Lösung war Spike-pragmatisch: kein preDestroy-Hook,
JVM-Shutdown reicht. Kein Exporter konfiguriert (Spec §6.7).

#### Hexagon-Fit
Sehr gut, mit einem Knackpunkt: Die ursprüngliche Lösung `@Singleton class
RegisterPlaybackEventBatchUseCase(...)` warf `jakarta.inject` in das
Inner-Hexagon. Auf User-Hinweis refaktoriert zu `object
RegisterPlaybackEventBatch { fun execute(input, ports..., clock) }` —
stateless function-module, keine DI-Annotation in `hexagon/`. Driving-
Adapter (Controller) injiziert die vier Driven-Ports und ruft die
Funktion direkt auf. Idiomatischer Kotlin-Stil, der das Go-`var _
Interface = (*Impl)(nil)`-Compile-Time-Checks-Muster funktional spiegelt.

#### Test-Velocity
- Inner-Hexagon-Unit-Tests (Kotest StringSpec): 14 Tests in ~80 ms.
- HTTP-Integrationstests (JUnit 5 + `@MicronautTest`): 10 Tests in
  ~910 ms (volle Server-Bootstrap pro Test-Klasse).
- Gesamtlauf inkl. Gradle-Overhead: ~22 s.

Stilbruch zwischen Kotest (Unit) und JUnit 5 (Integration), weil
Kotest5 + `@MicronautTest` `SpecInstantiationException` warf
(Spec-Konstruktor mit `@Inject @Client(...)` wird vom Kotest-Runner
nicht aufgelöst). Reine Kotest- oder reine JUnit5-Suite wäre einheitlicher;
Kotest behalten weil bessere DSL für Domain-Tests.

#### Docker Image Size
Final runtime-Image: **231 MB** (`eclipse-temurin:21-jre-alpine` plus
Micronaut-Application-Distribution unter `/opt/api`). Distroless Java
würde kleiner ausfallen, ist aber Bonus-Scope.

#### Cold Start
Erster `200 OK` auf `/api/health` nach `docker run`: **1.613 ms**
(via Curl-Loop mit 50-ms-Polling). Klassisches Micronaut-Verhalten:
Compile-Time-DI hilft, JVM-Class-Load + Netty-Bootstrap kosten ~1,6 s.
ZGC-Variante (Spec-Off-Scope) wurde nicht gemessen (siehe Plan §7.2 zur
Bewertungstrennung).

#### Build-Komplexität
Höher als Go: ein `Dockerfile`, ein `Makefile`, ein `build.gradle.kts`,
ein `gradle.properties`, ein `detekt.yml`, eine `application.yml` und
ein `logback.xml` — sieben Konfigurationsdateien. Plugins:
`kotlin("jvm")`, `kotlin("plugin.allopen")`, `com.google.devtools.ksp`,
`io.micronaut.application`, `io.gitlab.arturbosch.detekt`.

Spike-spezifische Stolpersteine im Build, jeder einzelne kostete
~15 min:
- `resolveAllDependencies` muss Konfigurationen explizit listen
  (compileClasspath, runtimeClasspath, ksp*…), sonst greift es
  Micronaut-BOM-unaufgelöste Versionen.
- `detekt` mit custom `srcDirs` braucht explizites `setSource(files(...))`.
- `detekt`-Regel-Kategorien: `SpreadOperator` lebt in `performance`,
  nicht `style` (zwei Fehlversuche).
- snakeyaml ist `runtimeOnly`-Pflicht, Micronaut wirft sonst beim
  `inspectRuntimeClasspath`.
- `PrometheusMeterRegistry`-Import: in Micrometer 1.13+ liegt der Class
  unter `io.micrometer.prometheusmetrics`, nicht mehr `prometheus`.
- Kotlin-Default-Werte werden vom KSP-Processor nicht respektiert;
  Konstruktoren mit `Int=100` / `Map=mapOf(...)` brauchen einen
  `@Factory` für die Bean-Bereitstellung.
- `@Replaces` ist application-context-global; eine globale
  `UnlimitedLimiter` für nur einen Test-Block braucht
  `@Requires(env=...)` plus `@MicronautTest(environments=[...])`.

#### Subjektiver Spaß
Kotlin-Sprachfeatures sind angenehm: sealed-class result types
(`RegisterBatchResult`), `data object`, value classes (`ProjectToken`),
exhaustive when-matching im Controller. Kotest StringSpec mit
`shouldBe`/`shouldThrow` liest sich klar. `data class` mit
`@JsonProperty` macht Wire-DTOs kurz.

Gegenkonto: jeder Build-Stolperstein (siehe oben) kostet
Konzentrations-Reset. Iterationsdauer im Spike (~30–40 s `compile` vs.
~5 s in Go) bremst spürbar, wenn man eng nachjustieren muss.

#### Contributor-Fit
Streaming-/Observability-OSS ist überwiegend Go (MediaMTX, OTel
Collector, Prometheus, Grafana). Kotlin/Micronaut hat eine kleinere
Schnittmenge in dem Umfeld; wer m-trace mitwartet, kommt eher aus
JVM-Backend-Welt als aus Streaming-Tooling. Vorteil: Konsistenz mit
dem schon existierenden Kotlin-Projekt `d-migrate` (gleicher Stack,
gleiche Konventionen — Cross-Project-Wartung leichter).

#### Absehbare Phase-2-Risiken
- **Persistenz-Wechsel**: Micronaut Data + JDBC oder Exposed sind
  beide gut verfügbar; Migration vom In-Memory-Repo ist mechanisch.
- **DI-Annotation-Drift in inner hexagon**: bei wachsendem Code
  könnte `jakarta.inject` doch in `hexagon/` rutschen, wenn Use Cases
  Klasseninstanzen mit DI-Lebenszyklus brauchen. Disziplin halten,
  oder pro Use Case einen `object` plus expliziten Adapter-Factory.
- **Cold-Start im Container-Restart-Szenario**: 1,6 s ist akzeptabel
  für Long-Running-Server, aber unangenehm bei häufigen K8s-Pod-
  Restarts oder bei Scale-from-zero. Native-Image (GraalVM) via
  Micronaut AOT würde das auf ~50 ms drücken — ist Phase-2.
- **Multi-Modul-Aufteilung**: `apps/api/` als ein Gradle-Modul
  funktioniert; bei `core / application / ports / adapters/...` als
  Sub-Module (Pattern aus d-migrate) wird der Build deutlich länger,
  KSP-Annotationen müssen pro Modul konfiguriert sein.

### 3.2 Bonus-Scope (umgesetzt?)

- [ ] `GET /api/stream-sessions`
- [ ] `GET /api/stream-sessions/{id}`
- [ ] Origin-Allowlist pro Project
- [ ] expliziter Session-Lifecycle `ended`
- [ ] OTel-Counter zusätzlich zu Prometheus
- [ ] `trace_id` in Logs
- [ ] vollständige Hexagon-Schichtung bis ins Detail

Keiner der Bonus-Punkte umgesetzt — Muss-Scope hatte Priorität im
Zeitbudget.

### 3.3 Stolpersteine / Beobachtungen

Übergreifend (Details auch oben in §3.1):

- **`@Singleton` im Inner-Hexagon**: erst auf User-Hinweis korrigiert
  (object-Refactor). Ohne den Hinweis hätte die "Hexagon-Fit"-Note
  schwächer ausgesehen.
- **DI-Defaults vs. KSP**: das Auseinanderdriften zwischen Kotlin-
  Sprachsemantik (Default-Parameter) und Micronaut-DI-Annotation-
  Processor war die mit Abstand teuerste Build-Iteration im Spike.
- **`@Replaces` global**: hat mich beim ersten Lauf ~15 min gekostet.
  Klares Lessons-learned: bei Micronaut-Tests immer scope-isolierende
  `@Requires(env=...)` mitdenken.
- **Kotest5 + `@MicronautTest`**: nicht ohne weiteres kombinierbar.
  Mixing Kotest (Unit) + JUnit5 (Integration) ist akzeptabel, sollte
  aber für 0.1.0+ vereinheitlicht werden — entweder beides Kotest5
  mit korrigierter Listener-Konfig, oder beides JUnit5.

### 3.4 Mess-Werte (Stand `7c8bc44`)

| Metrik | Wert |
|---|---|
| Wallclock bis erster grüner Test | ~10 min nach Bootstrap |
| Wallclock bis erstes `docker run` mit 200 OK | ~10 min nach Bootstrap |
| LoC `hexagon/domain/` | 59 |
| LoC `hexagon/application/` + `port/` | 246 |
| LoC `adapters/` (ohne Tests) | 393 |
| LoC Tests | 416 |
| LoC Total Production | 698 |
| Final Docker Image Size (`runtime`) | 231 MB |
| Cold Start (`/api/health` → 200) | 1.613 ms |
| Direkte Dependencies (`build.gradle.kts` deklariert) | 20 |
| Testlaufzeit (Gradle inkl. Suite-Setup) | ~22 s |
| Konfigurationsdateien | `Dockerfile`, `Makefile`, `.dockerignore`, `build.gradle.kts`, `gradle.properties`, `settings.gradle.kts`, `detekt.yml`, `resources/application.yml`, `resources/logback.xml` (9) |

---

## 4. Reihenfolge und Bias (Plan §4.4)

- Tatsächliche Reihenfolge: **Go zuerst, Kotlin/Micronaut danach**
  (Default aus Plan §4.4).
- Beobachteter Bias: erwarteter Vorteil für den zweiten Stack
  (Schema/Statuscodes/Edge Cases bereits durchdacht) hat **nicht**
  durchgeschlagen. Mehraufwand bei Kotlin/Micronaut war
  tooling-environment-spezifisch (Micronaut-BOM, KSP-Defaults,
  detekt-srcDirs, Micrometer-1.13-Package-Move, `@Replaces`-Scope,
  Kotest5+`@MicronautTest`-Inkompatibilität), nicht API-Spec-bezogen.
- Korrekturmaßnahmen während des Spikes: keine — der Bias hätte das
  Ergebnis nicht in Richtung Kotlin verschoben (siehe
  `docs/adr/0001-backend-stack.md` §4).

---

## 5. Subjektive Gesamteindrücke

Im ADR `docs/adr/0001-backend-stack.md` §5 (Bewertung) und §8
(Konsequenzen) verdichtet. Kernpunkte:

- Go fühlt sich für einen kleinen HTTP-Service mit OTel + Prometheus
  *kürzer und glatter* an als Kotlin/Micronaut. Stdlib `net/http`,
  `log/slog`, `prometheus/client_golang` reichen vollständig; keine
  Plugin-Stack-Pflege.
- Kotlin-Sprachfeatures (sealed result types, `data object`, value
  classes, exhaustive when) sind angenehm zu schreiben — gehen aber
  im Spike-Kontext gegen Build-Zyklus von 30–40 s pro `compile`
  (Go: ~5 s) verloren.
- Hexagon ist in beiden Stacks gleich gut umsetzbar (nach
  `object`-Refactor in Kotlin), aber Go ist näher an der
  Streaming-/Observability-OSS-Welt und erleichtert Contributor-Onboarding.

---

## 6. Übergang zu AP-3

- Bewertungsbogen (Spec §16) ausgefüllt im ADR: ☑
- Messwertbogen (Spec §17) ausgefüllt im ADR: ☑
- Reihenfolge-Bias im ADR notiert: ☑
- Sieger-Branch markiert: ☑ (`spike/go-api`,
  Final-Commit `7148a8d`)
- Unterlegener Branch gelöscht oder als Tag
  `spike/backend-stack-loser-2026-04-28` archiviert: ☑ (Tag gesetzt
  auf `7c8bc44`, Branch `spike/micronaut-api` gelöscht)
- Finale Commit-Hashes beider Prototypen im ADR: ☑

---

## 7. ADR

`docs/adr/0001-backend-stack.md` (Status: **Accepted**, Datum
2026-04-28). **Gewählt: Go.** Vorsprung 1,55 gewichtete Punkte
(38,75 pp), klar über der 10-pp-Schwelle aus Plan §4.6.
