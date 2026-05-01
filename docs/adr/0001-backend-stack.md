# 0001 — Backend-Technologie für `apps/api`

> **Status**: Accepted  
> **Datum**: 2026-04-28  
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)  
> **Bezug**: `spec/lastenheft.md` §9.1, §10.1; `docs/plan-spike.md`;
> `docs/spike/0001-backend-stack.md`;
> `spec/backend-api-contract.md`;
> `docs/spike/backend-stack-results.md`.

---

## 1. Kontext

Lastenheft §9.1 hielt die Wahl zwischen Go und Micronaut bewusst offen.
Spike gemäß `docs/spike/0001-backend-stack.md` und Implementierungsplan
`docs/plan-spike.md` durchgeführt: zwei Prototypen mit identischem
Muss-Scope (`spec/backend-api-contract.md`), Docker-only-Workflow,
ein gemeinsames Spike-Protokoll auf `main`.

Die Stack-Auswahl entscheidet, in welcher Sprache und mit welcher
Toolchain `apps/api` für `0.1.0` ausgebaut wird (Lastenheft §17 Schritt
2 ff.).

## 2. Optionen

| Option | Kurzbeschreibung |
|---|---|
| **Go** | Go 1.22+ mit Standard-Library `net/http`, `prometheus/client_golang`, `go.opentelemetry.io/otel`, `log/slog`. Distroless-static-Runtime. |
| **Kotlin/Micronaut** | Kotlin 2.1.x auf JDK 21 mit Micronaut 4.7, KSP für Annotation Processing, Micrometer-Prometheus, OpenTelemetry SDK, Logback. `eclipse-temurin:21-jre-alpine`-Runtime. |

Eine reine Java/Micronaut-Variante wurde im Plan §14.6 ausdrücklich
ausgeschlossen, weil die spätere `apps/api`-Implementierung in Kotlin
geplant war — der Vergleich sollte den ehrlichen Contributor-Fit
abbilden.

## 3. Scope

Beide Prototypen erfüllen den Muss-Scope vollständig
(`spec/backend-api-contract.md` §11):

- 3 Pflicht-Endpunkte: `POST /api/playback-events`,
  `GET /api/health`, `GET /api/metrics`
- 9-Schritt-Validierungsreihenfolge aus §5 mit deterministischen
  Statuscodes (202 / 400 / 401 / 413 / 422 / 429)
- 4 Pflicht-Counter mit Prefix `mtrace_`
- Token-Bucket-Rate-Limit (100 Events/Sek/Project)
- In-Memory-Persistenz, hardcodiertes Auth-Token (`demo-token`)
- minimaler OTel-Setup ("wired but silent" per Spec §6.7)
- strukturierte Logs

**Bonus-Scope**: keiner der sieben Bonus-Punkte (Spec §7) wurde
umgesetzt. Beide Prototypen blieben bewusst minimal, damit das
2-Tage-Budget pro Stack eingehalten wird.

**Tests**: 24 Tests pro Prototyp, alle grün. Aufteilung:
- Go: 14 Unit + 10 HTTP-Integration via stdlib `testing` + `httptest`.
- Kotlin: 14 Kotest + 10 JUnit5 (`@MicronautTest`). Stilbruch wird in
  Konsequenzen §8 adressiert.

Eine **Vertrags-Beobachtung** (keine Vertragsänderung): §5 step-3
(Rate-Limit) maskiert step-5 (Batch-Size > 100), wenn der Bucket
exakt die Batch-Size-Grenze hat. Beide Prototypen lösen das mit einer
Test-Only-`unlimitedLimiter`-Fixture für das §11 422-Pflichttest. Im
Spike-Protokoll §1 dokumentiert.

## 4. Reihenfolge und Bias

**Reihenfolge tatsächlich gebaut**: Go zuerst, Kotlin/Micronaut
danach (Default aus Plan §4.4).

**Erwarteter Bias**: der zweite Stack profitiert davon, dass
Schema, Statuscodes und Edge Cases bereits einmal durchdacht
wurden. Konkret hat Kotlin/Micronaut den fertigen API-Kontrakt aus
AP-0 unverändert übernommen.

**Tatsächlich beobachtet**: der Reihenfolge-Bias zugunsten Kotlin
hat nicht durchgeschlagen. Die Mehrheit der Kotlin-Mehraufwände
(siehe §6) waren tooling-environment-spezifisch (Micronaut-BOM,
KSP-Defaults, detekt-srcDirs, Micrometer-1.13-Package-Move,
`@Replaces`-Scope, Kotest5+`@MicronautTest`-Inkompatibilität) und
nicht Spec-bezogen. Die Reihenfolge half mit *"was muss die API
können"*, nicht mit *"wie kombiniere ich diese fünf Plugins"*.

Korrekturmaßnahme während des Spikes: keine. Der Bias wird hier
notiert, aber er hätte das Ergebnis nicht in Richtung Kotlin
verschoben.

Final-Commit-Hashes:
- `spike/go-api`: `7148a8d AP-1: Go HTTP integration tests for all §11 Pflichttests`
- `spike/micronaut-api`: `7c8bc44 AP-2: HTTP integration tests (JUnit 5 + @MicronautTest)`

## 5. Bewertung

Bewertet auf einer 1–5-Skala (höher = besser) per Plan §7.3 / Spec
§16. Notizen pro Kategorie sind in
`docs/spike/backend-stack-results.md` ausführlich; hier steht die
Verdichtung mit Begründung.

| Kategorie | Gewicht | Go | Go gewichtet | Kotlin | Kotlin gewichtet | Kommentar |
|---|---:|---:|---:|---:|---:|---|
| Time to Running Endpoint | 10% | 5 | 0,50 | 3 | 0,30 | Go: ~5 min nach Bootstrap. Kotlin: ~10 min, Micronaut-BOM-Resolution + KSP-Setup. |
| OTel-Integration-Ergonomie | 15% | 4 | 0,60 | 3 | 0,45 | Go: 5 imports, 1 Setup-Call. `semconv/v1.26.0`-Sub-Path-Quirk. Kotlin: `OpenTelemetrySdk.builder()` + `SdkMeterProvider.builder()` + `Resource.merge()` ist mehr Boilerplate; preDestroy-Stolperer. |
| Hexagon-Fit | 15% | 4 | 0,60 | 4 | 0,60 | Go: Package-Boundaries + `var _ Interface = (*Impl)(nil)`-Compile-Time-Checks. Kotlin: nach `object`-Refactor genauso sauber — Inner-Hexagon ist DI-frei, function module statt `@Singleton`-class. |
| Test-Velocity | 10% | 5 | 0,50 | 2 | 0,20 | Go: 24 Tests in 0,5 s, stdlib `testing` + `httptest`. Kotlin: 24 Tests in ~22 s (Gradle-Overhead + Micronaut-Bootstrap pro Test-Klasse), Stilbruch Kotest+JUnit5. |
| Docker Image Size | 5% | 5 | 0,25 | 2 | 0,10 | Go: 10,2 MB distroless static. Kotlin: 231 MB JRE-alpine. Distroless Java möglich aber nicht Spike-Scope. |
| Cold Start | 5% | 5 | 0,25 | 2 | 0,10 | Go: 9 ms statisch gelinktes Binary. Kotlin: 1.613 ms (JVM-Class-Load + Netty-Bootstrap). ZGC-Variante absichtlich nicht bewertet (Plan §7.2). |
| Build-Komplexität | 10% | 5 | 0,50 | 2 | 0,20 | Go: 5 Konfig-Files, 5 direkte Deps. Kotlin: 9 Konfig-Files, 5 Plugins, 20 direkte Deps, 6 Spike-spezifische Build-Stolperer (~90 min Gesamtkosten). |
| Subjektiver Spaß | 10% | 4 | 0,40 | 3 | 0,30 | Go: stdlib reicht, `log/slog`, method-aware Mux klar. Kotlin: sealed-class result + `data object` + value classes + exhaustive when sind angenehm; jeder Build-Stolperer reisst aber Konzentration auf, ~30–40 s `compile` vs. ~5 s in Go bremst. |
| Contributor-Fit | 10% | 5 | 0,50 | 3 | 0,30 | Go: MediaMTX, OTel Collector, Prometheus, Grafana — Streaming-/Observability-OSS dominant in Go. Kotlin: kleinere Schnittmenge im Streaming-Umfeld; Vorteil ist die Konsistenz mit `d-migrate` (Cross-Project-Wartung). |
| Absehbare Phase-2-Risiken | 10% | 3 | 0,30 | 3 | 0,30 | Go-Risiken: keine Compile-Time-Hexagon-Enforcement, CGO bricht distroless (SRT-Binding), WebSocket-Ökosystem fragmentiert. Kotlin-Risiken: DI-Drift in inner hexagon, Cold-Start in K8s-Scale-from-zero, KSP-Komplexität bei Multi-Modul-Aufteilung. Beide ähnlich schwer. |
| **Gesamt** | **100%** | | **4,40** | | **2,85** | Differenz: **1,55 gewichtete Punkte**, entspricht **38,75 Prozentpunkten** auf der 0–100%-Skala. |

## 6. Messwerte

| Metrik | Go (`7148a8d`) | Kotlin (`7c8bc44`) | Kommentar |
|---|---:|---:|---|
| Wallclock bis erster grüner Test | ~5 min | ~10 min | jeweils nach Bootstrap-Commit |
| Wallclock bis erstes `docker run` mit 200 OK | ~5 min | ~10 min | jeweils nach Bootstrap-Commit |
| LoC `hexagon/domain/` | 77 | 59 | Kotlin kompakter durch `data class` |
| LoC `hexagon/application/` + `port/` | 255 | 246 | vergleichbar |
| LoC `adapters/` (ohne Tests) | 508 | 393 | Kotlin spart durch Micronaut-Annotationen + Kotlin-Idiomatik |
| LoC Tests | 520 | 416 | Kotest + JUnit5 zusammen |
| LoC Total Production | 942 | 698 | Kotlin ~25% weniger Code |
| Final Docker Image Size (`runtime`) | 10,2 MB | 231 MB | Distroless static vs. JRE-alpine |
| Cold Start (`/api/health` → 200) | 9 ms | 1.613 ms | gemessen mit Curl-Polling |
| Build-Zeit `--no-cache` (deps + compile) | ~10 s | ~125 s | Kotlin: BOM-Resolution dominiert |
| Direkte Dependencies | 5 | 20 | Go nutzt fast nur stdlib |
| Testlaufzeit (`go test ./...` bzw. `gradle test`) | 0,5 s | ~22 s | Faktor 44 |
| Konfigurationsdateien | 5 | 9 | siehe Bewertungstabelle |

Image- und Cold-Start-Werte stammen aus dem jeweils im
Final-Commit gebauten `runtime`-Stage (`docker build --target
runtime`, `docker run -p 8080:8080`).

## 7. Entscheidung

**Gewählt: Go.**

Begründung: Der Vorsprung von 1,55 gewichteten Punkten
(38,75 Prozentpunkten auf 0–100%-Skala) liegt deutlich über der
in Plan §4.6 geforderten 10-Prozentpunkte-Schwelle. Inhaltlich
ziehen Go-Punkte vor allem aus den objektiven Messwerten
(Image-Größe, Cold-Start, Build-Komplexität, Test-Velocity) und
aus dem `Contributor-Fit` im Streaming-Observability-Ökosystem.
`Hexagon-Fit` ist nach dem Kotlin-`object`-Refactor in beiden
Stacks gleichwertig — also kein Pluspunkt für Kotlin als
JVM-DI-Vorteil.

`apps/api` wird ab `0.1.0` in Go aufgebaut. Spike-Branch
`spike/go-api` ist die Basis.

## 8. Konsequenzen

**Welche Lastenheft-Stellen werden konkret:**

- §9.1 (Backend-Entscheidung): in den Vergangenheitsmodus
  setzen — Go ist gewählt.
- §10.1 (Backend-Tabelle): nur noch die Go-Zeile aktiv.
- §16.2 (offene Entscheidungen): "Backend-Technologie final" wird
  resolved und aus der Liste entfernt.

Lastenheft-Version-Bump auf `1.0.0` gemäß Plan §11.

**Welche Anforderungen werden technisch anders umgesetzt als
ursprünglich offen:**

- DI-Annotations-Strategie ist nicht mehr nötig — Go hat keine
  DI-Container-Idiome. `var _ Interface = (*Impl)(nil)`
  Compile-Time-Checks bleiben das primäre Hexagon-Enforcement.
- Coverage: Kover entfällt; `go test -cover` ist seit Pre-MVP
  `0.1.0` als Coverage-Werkzeug eingeführt — Pflicht-Gate mit
  90 %-Threshold und HTML-/Übersicht-Report (siehe
  `docs/quality.md` §3).
- detekt-Konfig fällt weg; `golangci-lint` mit Default-Lintern
  ist die Soll-Vorgabe.

**Welche Risiken entstehen, welche verschwinden:**

- *Verschwindet*: Cold-Start-Sorgen bei K8s-Scale-from-zero,
  Bootstrap-Tooling-Drift (BOM-Resolution, KSP-Defaults).
- *Bleibt*: Hexagon-Boundaries werden nur über Disziplin und
  Code-Review erzwungen (Go-Module-Splitting via `go.work` ist
  Phase-2 falls nötig).
- *Neu*: SRT-Binding-Integration könnte CGO erfordern und damit
  das distroless-static-Pattern brechen — siehe Lastenheft §4.3.
  Wird in der SRT-Health-Phase (Roadmap `0.6.0`) entschieden.

**Welche Folge-ADRs werden absehbar nötig:**

- Persistenz-Wechsel In-Memory → SQLite/PostgreSQL (Phase
  `0.1.0`/`0.2.0`).
- WebSocket vs. SSE für Live-Updates (Phase `0.4.0` Trace-Ansicht).
- SRT-Binding-Stack (Phase `0.6.0` SRT-Health, ggf. CGO).

**Welche Spike-Ergebnisse werden archiviert:**

- `spike/micronaut-api` wird gelöscht und als Tag
  `spike/backend-stack-loser-2026-04-28` archiviert. Final-Commit
  `7c8bc44` bleibt referenzierbar.
- `spike/go-api` wird Basis für `apps/api` in `main`. Branch wird
  *nicht* gelöscht, aber auch nicht parallel zu `main`
  weiterentwickelt — Folge-Arbeit landet auf `main`-Branches
  (`feat/...` o. ä.).

**Test-Stack-Vereinheitlichung für 0.1.0+ (entfällt mit Go):**

War für Kotlin notwendig (Kotest+JUnit5-Mix). Mit Go-Wahl
besteht das Problem nicht — `testing` + `httptest` deckt beide
Test-Sorten ab.

**Build-Image-Wahl:**

`golang:1.22` als Build-Image, `gcr.io/distroless/static-debian12`
als Runtime. Bleibt unverändert aus dem Spike (`apps/api/Dockerfile`).

---

## Anhang A — Bezüge

- Lastenheft (`spec/lastenheft.md`): §9.1, §10.1, §16.2.
- Spike-Spec (`docs/spike/0001-backend-stack.md`): §6, §7, §9, §11, §15.
- Spike-Plan (`docs/plan-spike.md`): §4.6 (Entscheidungsregel),
  §7.2 (Mess-Punkte), §7.3 (Bewertungsraster), §10 (DoD), §11
  (Anschluss an MVP).
- API-Kontrakt (`spec/backend-api-contract.md`): vollständig.
- Spike-Protokoll (`docs/spike/backend-stack-results.md`): §2 (Go),
  §3 (Kotlin/Micronaut).
