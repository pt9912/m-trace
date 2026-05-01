# Implementierungsplan: Backend-Spike

> **Milestone**: 0.1.0 — OTel-native Local Demo (Vorstufe)
> **Phase**: Backend-Technologie-Entscheidung
> **Status**: Geplant
> **Referenz**: `docs/spike/0001-backend-stack.md`;
> `spec/lastenheft.md` §9.1, §10.1, §16.2, §17 Schritt 0;
> `docs/adr/0001-backend-stack.md` (entsteht);
> `README.md`.

---

## 1. Ziel (SP-1)

Der Backend-Spike liefert die technische Entscheidung zwischen Go und Micronaut
für `apps/api` durch zwei zeitlich begrenzte Prototypen mit identischem
Funktionsumfang. Er produziert keinen Produktionscode und kein
MVP-Skelett, sondern genug Vergleichsmaterial für eine **belastbare**
Entscheidung — auf Basis eigener Erfahrung, nicht weiterer Recherche.

Dieser Plan operationalisiert die Spike-Spezifikation aus
`docs/spike/0001-backend-stack.md`. Er bricht sie in datierte Arbeitspakete,
Abnahmekriterien, Verifikationsschritte und eine Definition of Done auf.

Konkrete Ergebnisse nach Abschluss (mit verbindlichen Pfaden):

- API-Kontrakt-Datei `spec/backend-api-contract.md`
- Branch `spike/go-api` mit lauffähigem Go-Prototyp
- Branch `spike/micronaut-api` mit lauffähigem Micronaut-Prototyp
- Spike-Protokoll `docs/spike/backend-stack-results.md` mit
  Live-Notizen, Vertragsänderungen und subjektiven Eindrücken pro
  Bewertungskategorie
- ADR `docs/adr/0001-backend-stack.md` mit gewähltem Stack; enthält
  den ausgefüllten Bewertungsbogen (Spec §16) und Messwertbogen
  (Spec §17) für beide Prototypen
- Sieger-Branch als Basis für die spätere `apps/api`-Implementierung

Nicht Ergebnis dieses Plans:

- produktionsreife API
- vollständiges Mono-Repo-Skelett
- Frontend, Player-SDK, MediaMTX-Integration
- Persistenz auf Disk
- Multi-Tenant-Verwaltung

Wichtig: Phase-Bezeichnungen "Tag 0" bis "Tag 5" sind Maximalbudgets
gemäß `docs/spike/0001-backend-stack.md` §2, keine Kalendertage. Die
harte Gesamtgrenze beträgt 5 Arbeitstage (0,5 + 2 + 2 + 0,5).

---

## 2. Ausgangslage (SP-2)

### 2.1 Lastenheft hält Backend-Entscheidung bewusst offen (SP-3)

`spec/lastenheft.md` §9.1 nennt zwei zulässige Optionen — Go oder Micronaut —
und bindet die Wahl explizit an einen technischen Spike. §10.1 listet die
Mindestanforderungen unabhängig vom Stack. §16.2 führt die endgültige
Backend-Technologie als offene Entscheidung. §17 Schritt 0 fordert den Spike
als Voraussetzung jeder MVP-Implementierung.

### 2.2 Spike-Spezifikation ist normativ (SP-4)

`docs/spike/0001-backend-stack.md` ist verbindlich für:

- Muss- und Bonus-Scope der Prototypen
- API-Endpunkte, Event-Schema, Validierungsregeln
- Authentifizierung mit `X-MTrace-Token`
- Pflichtmetriken im Prometheus-Format
- Bewertungsraster mit Gewichtung
- Messwerttabelle für ADR
- Entscheidungsregel ohne Unentschieden

Dieser Plan verfeinert die Spec, widerspricht ihr nicht. Bei Konflikten
gilt die Spec.

### 2.3 Repository ist noch leer (SP-5)

Aktueller Stand:

- `README.md` mit Projekt-Vision
- `spec/lastenheft.md`
- `docs/spike/0001-backend-stack.md`
- `docs/planning/plan-spike.md` (dieses Dokument)
- kein `apps/`, kein `packages/`, kein `services/`
- keine CI, kein `Makefile`, kein `docker-compose.yml`

Verbindliche Folge:

- Spike-Branches arbeiten gegen einen leeren Working Tree
- der Plan darf keine bestehenden Komponenten als gegeben annehmen
- Mono-Repo-Tooling (pnpm-Workspace, Gradle, Go-Modules) wird erst im
  jeweiligen Prototyp eingeführt, nicht im Spike-Vorlauf

---

## 3. Scope für den Spike (SP-6)

### 3.1 In Scope (SP-7)

- API-Kontrakt-Datei mit Endpunkten, Beispielpayloads, Statuscodes,
  Header-Definitionen, Metriknamen und Testfällen
- Go-Prototyp gemäß Muss-Scope der Spec
- Micronaut-Prototyp gemäß Muss-Scope der Spec
- In-Memory-Event-Repository, Prometheus-Metrics-Publisher, In-Memory-Rate-
  Limiter pro Prototyp
- minimaler OTel-Setup zum Bewerten der Ergonomie
- strukturierte JSON-Logs
- Dockerfile mit Multi-Stage-Build pro Prototyp
- Pflichttests gemäß Spec §6.12
- ausgefüllter Bewertungs- und Messwertbogen
- ADR mit Stack-Entscheidung, Bias-Notiz und Konsequenzen

### 3.2 Bewusst nicht Teil des Spikes (SP-8)

Übernommen aus `docs/spike/0001-backend-stack.md` §8 und hier verbindlich:

- Persistenz auf Disk, SQLite, Redis
- dynamische Project-Verwaltung
- Authentication-Flows jenseits des hardcodierten Tokens
- WebSocket-Endpunkte
- Stream-Analyzer-Anbindung
- Frontend-Integration
- Migrations, Schemas, ORM
- Tempo, Loki, OTel Collector
- Kubernetes
- Performance-Optimierung jenseits "läuft mit Tests durch"
- Lasttests unter realer Produktionslast

Präzisierung:

Der Spike beantwortet "welcher Stack passt besser zu m-trace?". Er beantwortet
nicht "wie sieht apps/api in Produktion aus?" und nicht "welche Persistenz
nutzt der MVP?".

---

## 4. Leitentscheidungen (SP-9)

### 4.1 API-Kontrakt vor jedem Implementierungsstart (SP-10)

Vor `git checkout -b spike/go-api` wird `spec/backend-api-contract.md`
erstellt und committed. Inhalt: Endpunkte, Beispielpayloads, Statuscodes,
Header, Metriknamen, minimale Fehlerfälle, Testfälle.

Verbindliche Folge:

- der Kontrakt darf nach dem ersten Implementierungs-Branch nicht mehr
  einseitig geändert werden
- Änderungen müssen in beiden Prototypen identisch landen und im
  Spike-Protokoll (`docs/spike/backend-stack-results.md`) begründet sein
- der zweite Stack darf nicht davon profitieren, dass unklare
  Anforderungen erst im ersten Prototyp entdeckt wurden
- genehmigte Vertragsänderungen aus AP-1 werden **vor** AP-2-Start nach
  `main` gemerged; AP-2 (§6.3) zieht den Branch damit gegen den
  finalen Vertrag
- der AP-1-Branch wird mit dem geänderten Vertrag aktualisiert (Rebase
  oder Merge), damit beide Prototypen am Ende gegen denselben Vertrag
  bewertet werden
- das Spike-Protokoll `docs/spike/backend-stack-results.md` lebt
  durchgehend auf `main` und wird **nicht** in den Spike-Branches
  gepflegt. Jeder AP committet seine Notizen direkt auf `main` (eigener
  Commit, separat vom Prototypen-Code), damit AP-2 die AP-1-Notizen
  ohne Merge sieht und AP-3 in Konsequenz auf einer vollständigen
  Notiz-Historie aufsetzt.

### 4.2 Harte Zeitgrenze, keine dritte Spike-Runde (SP-11)

Maximalbudgets:

| Phase | Maximum |
|---|---:|
| API-Kontrakt fixieren | 0,5 Tag |
| Go-Prototyp | 2 Tage |
| Micronaut-Prototyp | 2 Tage |
| Auswertung und ADR | 0,5 Tage |

Verbindliche Folge:

- erreicht ein Prototyp den Muss-Scope nicht in 2 Tagen, ist *das* die
  Erkenntnis
- der Plan erlaubt keinen "halben Tag mehr"
- ein dritter Spike-Versuch ist ausgeschlossen

### 4.3 Identischer Scope, keine Coexistenz danach (SP-12)

Beide Prototypen liefern denselben Muss-Scope. Erweiterungen "weil es
schnell geht" oder Auslassungen "weil es im anderen Stack einfacher ist"
sind verboten. Bonusfunktionen dürfen den Muss-Scope nicht gefährden.

Verbindliche Folge:

- Sieger-Branch wird Basis für `apps/api`
- unterlegener Branch wird nicht weiterentwickelt
- unterlegener Branch wird gelöscht oder als Tag
  `spike/backend-stack-loser-YYYYMMDD` archiviert
- der finale Commit-Hash beider Prototypen wird im ADR referenziert

### 4.4 Reihenfolge-Bias wird dokumentiert, nicht ausgeglichen (SP-13)

Default-Reihenfolge: Go-Prototyp zuerst, Micronaut-Prototyp zweit.

Begründung:

Go ist im Streaming-/Observability-Umfeld kulturell näher (siehe Lastenheft
§9). Micronaut profitiert davon, dass Edge Cases bereits einmal durchdacht
wurden. Der Bias wird im ADR ausdrücklich notiert. Künstliches Ausgleichen
durch zusätzlichen Aufwand im Erststack ist nicht vorgesehen.

Verbindliche Folge:

- die Reihenfolge darf bewusst gedreht werden, wenn Micronaut sonst einen
  unfairen Erfahrungsvorteil hätte
- die tatsächliche Reihenfolge wird im ADR festgehalten

### 4.5 Hexagon nur dort, wo Fachlogik entsteht (SP-14)

Beide Prototypen müssen Domain-Logik frameworkfrei halten. Die
Standard-Schichtung ist:

```text
src/
├── hexagon/
│   ├── domain/
│   ├── port/
│   │   ├── in/
│   │   └── out/
│   └── application/
└── adapters/
    ├── in/
    │   └── http/
    └── out/
        ├── persistence/
        ├── telemetry/
        └── metrics/
```

Verbindliche Folge:

- die Domain darf keine HTTP-, DB-, Framework-, Docker- oder OTel-
  Implementierungstypen referenzieren
- DTOs liegen in den Adapters, nicht in der Domain
- Adapter implementieren Ports oder rufen Use Cases

### 4.6 Eine Entscheidung, keine "beide Optionen sind gut" (SP-15)

Die Entscheidungsregel aus Spec §10 ist verbindlich:

1. ≥ 10 gewichtete Prozentpunkte Vorsprung → Sieger
2. < 10 Prozentpunkte → `Contributor-Fit` entscheidet
3. Gleichstand bei `Contributor-Fit` → subjektive Wartbarkeit nach 2 Tagen
4. weiterhin gleich → der Stack mit weniger Infrastruktur-/Build-Komplexität
5. keine weitere Spike-Runde

Das ADR muss eine *gewählte* Option benennen.

---

## 5. Zielarchitektur (SP-16)

### 5.1 Branch-Layout im Repository (SP-17)

| Branch | Zweck | Lebensdauer |
|---|---|---|
| `main` | Lastenheft, Spike-Doc, Plan, README | dauerhaft |
| `spike/go-api` | Go-Prototyp | bis Auswertung; danach Sieger oder Archiv |
| `spike/micronaut-api` | Micronaut-Prototyp | bis Auswertung; danach Sieger oder Archiv |

Hinweis zum unterlegenen Stack: nach AP-3 wird der unterlegene
Spike-Branch entweder gelöscht oder als **Tag**
`spike/backend-stack-loser-YYYYMMDD` archiviert (kein paralleler
Branch). Details in §4.3 und §14.5.

`main` enthält keine Prototyp-Sourcen. Die API-Kontrakt-Datei
`spec/backend-api-contract.md` wird vor dem ersten Implementierungs-
Branch nach `main` gemerged.

### 5.2 Mindeststruktur pro Prototyp (SP-18)

Verbindlich ist die *logische* Schichtung, nicht der konkrete Dateipfad.
Beide Prototypen müssen folgende Schichten erkennbar voneinander trennen:

- `hexagon/domain/` — frameworkfreie Fachobjekte
- `hexagon/port/driving/` — Eingangs-Ports (Use-Case-Schnittstellen)
- `hexagon/port/driven/` — Ausgangs-Ports (Repository, Publisher)
- `hexagon/application/` — Use Cases / Application Services
- `adapters/driving/http/` — HTTP-Controller (inbound)
- `adapters/driven/persistence/` — Event-Repository (outbound)
- `adapters/driven/telemetry/` — OTel-Setup (outbound)
- `adapters/driven/metrics/` — Prometheus-Publisher (outbound)

Hexagon und Adapters liegen **direkt** unter `apps/api/` — bewusst
flach, damit die Architektur beim ersten `ls apps/api/` sichtbar ist
(kein `internal/`-Wrapper, keine zusätzliche
`src/main/kotlin/dev/mtrace/api/`-Ebene). Die Abbildung pro Stack:

- **Go**: `apps/api/hexagon/...` und `apps/api/adapters/...` direkt;
  Entry-Point unter `apps/api/cmd/api/main.go`. Konkreter Tree in §12.1.
- **Micronaut/Kotlin**: `apps/api/hexagon/...` und
  `apps/api/adapters/...` direkt, via custom Gradle-`srcDirs`
  (siehe §12.2). Package-Namen bleiben sauber
  (`dev.mtrace.api.hexagon.domain`, ...), Verzeichnisse sind flach.

Stack-spezifische Builddateien (`go.mod`/`go.sum` bzw.
`build.gradle.kts`/`gradle.properties`, `detekt.yml`) und
`Dockerfile`/`Makefile` liegen ergänzend direkt unter `apps/api/`.

Verbindliche Folge:

- Vergleichende LoC-Messungen (siehe §7.2) zählen die *logischen*
  Schichten, nicht ein bestimmtes Verzeichnis. `cloc` läuft pro Prototyp
  gegen den jeweiligen Domain- bzw. Adapter-Pfad gemäß §12.
- Hexagon-/Adapter-Verzeichnisse direkt unter `apps/api/` sind
  Pflicht — keine zusätzliche `internal/`-Ebene (Go) und keine
  zusätzliche `src/main/kotlin/dev/mtrace/api/`-Ebene (Kotlin).
  Architektur muss beim ersten Blick sichtbar sein.

### 5.3 Domain-Objekte (SP-19)

Beide Prototypen modellieren mindestens:

- `PlaybackEvent`
- `StreamSession`
- `Project`
- `ProjectToken`

Use Case:

- `RegisterPlaybackEventBatch`

Ports:

- `EventRepository` (out)
- `MetricsPublisher` (out)
- `PlaybackEventInbound` (in)

Adapter:

- `InMemoryEventRepository`
- `PrometheusMetricsPublisher`
- HTTP-Controller für `POST /api/playback-events`, `GET /api/health`,
  `GET /api/metrics`

`StreamSession` hat im Muss-Scope nur den Lifecycle `active` (automatisch
angelegt) und `ended` (Bonus). Ein expliziter Session-Ende-Endpunkt ist
nicht Pflicht.

### 5.4 Metriken-Vertrag (SP-20)

Pflichtmetriken im Prometheus-Format an `GET /api/metrics`:

| Metrik | Bedeutung |
|---|---|
| `mtrace_playback_events_total` | angenommene Playback-Events |
| `mtrace_invalid_events_total` | Schema-/Validierungsfehler |
| `mtrace_rate_limited_events_total` | Rate-Limit-Ablehnungen |
| `mtrace_dropped_events_total` | interne Drops |

Hochkardinale Labels (`session_id`, `user_agent`, `segment_url`) sind
verboten. `mtrace_dropped_events_total` darf in Phase 1 bei `0` bleiben,
wenn kein Drop-Pfad implementiert wird; die Metrik muss aber existieren.

### 5.5 OTel-Mindestumfang (SP-21)

Verbindlich pro Prototyp:

- minimaler OTel-Meter- oder Trace-Setup im Code
- Bewertung der Integration im Spike-Protokoll
- keine externe OTel-Infrastruktur erforderlich

Optional (Bonus):

- OTel-Counter analog zu Prometheus-Metriken
- vorbereiteter OTLP-Export
- `trace_id` in Logs

OTel ist *Ergonomie-Test*, nicht produktive Telemetrie-Pipeline.

---

## 6. Arbeitspakete (SP-22)

### 6.1 AP-0: API-Kontrakt fixieren (SP-23)

Maximalbudget: 0,5 Tag.

Aufgaben:

- Datei `spec/backend-api-contract.md` erstellen
- Endpunkte, Pfade, Methoden, Statuscodes festschreiben
- Beispielpayload für `POST /api/playback-events` festschreiben
- Header-Definitionen (`X-MTrace-Token`, `Retry-After`) festschreiben
- Metriknamen festschreiben (siehe §5.4)
- Validierungsregeln und Fehlerfälle festschreiben
- Pflichttests (Happy Path und Fehlerfälle) festschreiben
- Datei nach `main` mergen
- Spike-Protokoll-Skelett `docs/spike/backend-stack-results.md` auf
  `main` anlegen (Abschnitte: AP-1-Notizen, AP-2-Notizen,
  Vertragsänderungen, Subjektive Eindrücke). Datei lebt fortan auf
  `main` und wird von AP-1/AP-2 dort ergänzt — siehe §4.1.

Abnahme:

- Datei existiert auf `main`
- `docs/spike/backend-stack-results.md` existiert auf `main` mit
  leerem Skelett
- jeder Endpunkt hat Beispielpayload, Statuscode-Liste und mindestens einen
  Testfall
- alle Metriken der Spec sind genannt
- nach Merge nach `main` werden keine Vertragsänderungen mehr einseitig
  vorgenommen

### 6.2 AP-1: Go-Prototyp (SP-24)

Maximalbudget: 2 Tage. Branch: `spike/go-api`.

Stack:

- Go 1.22+
- Standard-Library für HTTP; `chi` falls Standard `net/http` zu spartanisch
- OTel via `go.opentelemetry.io/otel`
- Tests mit `testing` und `httptest`
- Logging mit `log/slog`
- Linting: `golangci-lint` (Soll, siehe §14.9)

Aufgaben:

- Branch von `main` ziehen
- `go.mod` initialisieren
- Domain, Use Case, Ports, Adapter gemäß §5.2/§5.3 anlegen
- HTTP-Controller, In-Memory-Repository, Prometheus-Publisher, In-Memory-
  Rate-Limiter implementieren
- minimaler OTel-Setup integrieren
- strukturierte JSON-Logs einrichten
- Dockerfile mit Multi-Stage-Build (`gcr.io/distroless/static-debian12`)
- Pflichttests gemäß Spec §6.12 schreiben
- Image als `m-trace-api-spike:go` taggen
- `make test` ruft `docker build --target test` auf und muss grün
  durchlaufen; `docker run -p 8080:8080 m-trace-api-spike:go` startet
  den Service
- `make lint` ruft `docker build --target lint` auf (Soll-Aufgabe;
  siehe §14.9). Stage führt intern `golangci-lint run ./...` aus.
- Spike-Notizen pro Bewertungskategorie direkt auf `main` in
  `docs/spike/backend-stack-results.md` committen (siehe §4.1) —
  **nicht** in `spike/go-api`

Abnahme:

- alle Pflicht-Endpunkte liefern erwartete Statuscodes
- alle Pflichttests grün
- Docker-Image baut und startet, `/api/health` liefert HTTP 200
- alle Pflichtmetriken sichtbar an `/api/metrics`
- OTel ist mindestens einmal im Code berührt
- AP-1-Notizen sind auf `main` in
  `docs/spike/backend-stack-results.md` festgehalten und vor AP-2-
  Start sichtbar

### 6.3 AP-2: Micronaut-Prototyp (SP-25)

Maximalbudget: 2 Tage. Branch: `spike/micronaut-api`.

Stack:

- Micronaut 4.x
- Kotlin 2.1.x auf JDK 21 (siehe §14.6 für Begründung)
- Versionen zentral in `apps/api/gradle.properties` (Pattern aus
  d-migrate `gradle.properties`), nicht als String-Literale verstreut
- OTel via Micronaut-OpenTelemetry-Modul oder direkte SDK-Nutzung
- Tests mit Kotest 6.x + MockK + `@MicronautTest`-Integration
  (siehe §14.10 für Default-Wahl)
- Logging mit Logback (1.5.x), optional Logstash-Encoder
- Linting: `detekt` 1.23+ (Soll, siehe §14.9)

Aufgaben:

- Branch von `main` ziehen (nicht von `spike/go-api`); damit sind
  AP-1-Notizen aus `docs/spike/backend-stack-results.md` schon
  sichtbar
- `build.gradle.kts`, `gradle.properties` und Micronaut-Application
  initialisieren. **Kein** `gradle-wrapper.jar` ins Repo — Build-Image
  bringt Gradle 8.12 mit (`gradle:8.12-jdk21`).
- Domain, Use Case, Ports, Adapter gemäß §5.2/§5.3 anlegen
- HTTP-Controller, In-Memory-Repository, Prometheus-Publisher, In-Memory-
  Rate-Limiter implementieren
- minimaler OTel-Setup integrieren
- strukturierte JSON-Logs einrichten
- Dockerfile mit Multi-Stage-Build:
  Build-Image `gradle:8.12-jdk21`, Final-Image
  `eclipse-temurin:21-jre-alpine` (alternativ Distroless Java)
- Pflichttests gemäß Spec §6.12 schreiben
- Image als `m-trace-api-spike:micronaut` taggen
- `make test` ruft `docker build --target test` auf und muss grün
  durchlaufen; `docker run -p 8080:8080 m-trace-api-spike:micronaut`
  startet den Service
- `make lint` ruft `docker build --target lint` auf (Soll-Aufgabe;
  siehe §14.9). Stage führt intern `gradle --no-daemon detekt` aus.
  Stage-Name ist über beide Stacks hinweg `lint` (siehe §14.11).
- AP-2-Notizen pro Bewertungskategorie direkt auf `main` in
  `docs/spike/backend-stack-results.md` committen (siehe §4.1) —
  **nicht** in `spike/micronaut-api`

Abnahme:

- identisch zu AP-1; `make test` ruft die Kotlin-Test-Stage des
  Dockerfiles auf (statt der Go-Stage)

### 6.4 AP-3: Auswertung und ADR (SP-26)

Maximalbudget: 0,5 Tag.

Aufgaben:

- Bewertungsraster (Vorlage Spec §16) für beide Prototypen direkt im
  ADR ausfüllen
- Messwertbogen (Vorlage Spec §17) für beide Prototypen direkt im ADR
  ausfüllen
- Reihenfolge-Bias und subjektive Eindrücke aus
  `docs/spike/backend-stack-results.md` ins ADR übernehmen
- Entscheidung gemäß Entscheidungsregel treffen (§4.6)
- ADR `docs/adr/0001-backend-stack.md` schreiben (Vorlage Spec §15)
- Sieger-Branch markieren; unterlegenen Branch löschen oder als Tag
  archivieren
- finale Commit-Hashes beider Prototypen im ADR referenzieren

Abnahme:

- Bewertungs- und Messwertbogen sind vollständig im ADR ausgefüllt
- ADR existiert auf `main` mit klarer Stack-Entscheidung
- Bias-Abschnitt im ADR ist nicht leer und referenziert das
  Spike-Protokoll
- unterlegener Branch ist nicht mehr aktiver Entwicklungspfad

---

## 7. Verifikationsstrategie (SP-27)

### 7.1 Pflichttests pro Prototyp (SP-28)

Jeder Prototyp muss ohne externe Dienste folgende Tests grün haben:

- Unit-Test `RegisterPlaybackEventBatch`: Happy Path
- Unit-Test zentrale Domain-Validierung: Pflichtfelder, Schema-Version
- Integrationstest `POST /api/playback-events`: Happy Path mit gültigem
  Token
- Integrationstest HTTP 400 bei abweichender `schema_version`
- Integrationstest HTTP 401 bei fehlendem oder falschem Token
- Integrationstest HTTP 401 bei gültigem Token, dessen `project_id` im
  Event nicht zum Token passt (Spec §6.4)
- Integrationstest HTTP 401 bei unbekanntem `project_id` (Spec §6.4)
- Integrationstest HTTP 413 bei Body über 256 KB
- Integrationstest HTTP 422 bei ungültigem Event (fehlendes Pflichtfeld)
- Integrationstest HTTP 422 bei leerem `events`-Array (`[]`)
- Integrationstest HTTP 422 bei fehlendem `events`-Feld
- Integrationstest HTTP 422 bei mehr als 100 Events im Batch
- Integrationstest HTTP 429 bei Rate-Limit-Überschreitung mit
  `Retry-After`-Header

Diese Tests decken sämtliche Validierungsregeln aus Spec §6.3 ab. Ein
Prototyp mit einem fehlenden oder fehlschlagenden Pflichttest ist nur
dann DoD-fähig, wenn das Scheitern gemäß §10 dokumentiert ist.

Bonus-Tests:

- `GET /api/stream-sessions` Liste
- `GET /api/stream-sessions/{id}` Detail

Ein einziger Testbefehl muss funktionieren:

```bash
make test
```

`make test` ruft intern `docker build --target test
-t m-trace-api-spike:<stack>-test .` auf. Lokale Toolchain-Befehle
(`go test ./...`, `./gradlew test`) sind kein Pflichtweg — der
Spike ist Docker-only (siehe §14.11 und Spec §6.11).

### 7.2 Mess-Punkte (SP-29)

Pro Prototyp objektiv festzuhalten:

| Metrik | Erfassung |
|---|---|
| Wallclock bis erster grüner Test | Stoppuhr |
| Wallclock bis erstes `docker run` mit 200 OK | Stoppuhr |
| LoC im Domain-Layer | `cloc` über die Domain-Pfade gemäß §12.1/§12.2 |
| LoC im Adapter-Layer | `cloc` über die Adapter-Pfade gemäß §12.1/§12.2 |
| Artefaktgröße | Go-Binary bzw. JAR/App |
| Final Docker Image Size | `docker images` |
| Cold Start bis erster 200 OK auf `/api/health` | `time` + Curl-Loop |
| Build-Zeit von Scratch | `time docker build --no-cache` |
| Größe des Dependency-Caches | isolierter Cache pro Prototyp (siehe Hinweis) |
| Anzahl direkter Dependencies | direkt deklariert in `apps/api/go.mod` (`require`-Block ohne `// indirect`) bzw. `apps/api/build.gradle.kts` (Einträge im `dependencies {}`-Block: `implementation`, `api`, `testImplementation`, ...). Transitive Dependencies werden nicht mitgezählt. |
| Testlaufzeit | `time make test` |
| Anzahl direkt geschriebener Konfigurationsdateien | manuell |

Hinweis zum Dependency-Cache:

Im Docker-only Workflow lebt der Dependency-Cache in einem
Docker-Layer der `deps`-Stage, nicht im Host-Filesystem. Gemessen
wird die Größe dieses Layers nach einem `--no-cache`-Build:

```bash
docker build --no-cache --target deps -t m-trace-api-spike:<stack>-deps .
docker history --no-trunc --format "{{.Size}}\t{{.CreatedBy}}" \
  m-trace-api-spike:<stack>-deps | head -5
```

Notiert wird:

- die Image-Größe der `deps`-Stage abzüglich der jeweiligen Base-
  Image-Größe (`docker image inspect ... --format '{{.Size}}'`)
- die Größe des Layers, der `gradle resolveAllDependencies` bzw.
  `go mod download` ausführt — das ist der eigentliche Dependency-
  Cache-Footprint

Globale Host-Caches (`~/go/pkg/mod`, `~/.gradle`) bleiben unberührt
und werden nicht gemessen. Damit ist der Wert Docker-reproduzierbar
und unabhängig von der lokalen Entwicklungs-Umgebung.

Hinweis zum JVM-Cold-Start (Micronaut-Variante):

Bewertet wird **ausschließlich** der Default-JVM-Cold-Start ohne
zusätzliche Tuning-Flags. JVM-Tuning ist gemäß §7.4 und Spec §8 aus
dem Spike-Scope ausgeschlossen, damit der Vergleich Go vs.
Kotlin/Micronaut nicht durch Runtime-Konfiguration verzerrt wird.

Optional darf eine **nicht-bewertete** Zusatznotiz im Messwertbogen
stehen: Cold-Start-Wert mit `-XX:+UseZGC -XX:+ZGenerational` (Pattern
aus d-migrate). Der Wert dient nur als Referenz für die
spätere `apps/api`-Konfiguration im MVP und fließt nicht in die
Bewertungskategorie `Cold Start` ein.

### 7.3 Bewertungsraster (SP-30)

Verbindlich aus Spec §9, identisch für beide Prototypen, Punkte 1–5:

| Kategorie | Gewicht |
|---|---:|
| Time to Running Endpoint | 10% |
| OTel-Integration-Ergonomie | 15% |
| Hexagon-Fit | 15% |
| Test-Velocity | 10% |
| Docker Image Size | 5% |
| Cold Start | 5% |
| Build-Komplexität | 10% |
| Subjektiver Spaß | 10% |
| Contributor-Fit | 10% |
| Absehbare Phase-2-Risiken | 10% |

Nicht direkt in der Bewertung:

- "welche Sprache ich besser kann" — fließt indirekt über Velocity ein
- Performance-Benchmarks unter Last
- theoretische Vorteile aus Blogposts

### 7.4 Keine Lasttests, keine Produktionssimulation (SP-31)

Performance-Optimierung über "läuft mit Tests durch" hinaus ist nicht Teil
des Spikes. Lasttests, Profiling-Sessions und JVM-Tuning sind ausgeschlossen.

---

## 8. Betroffene Codebasis (SP-32)

Voraussichtlich erzeugt:

- `spec/backend-api-contract.md` (neu, auf `main`)
- `docs/adr/0001-backend-stack.md` (neu, auf `main`)
- `apps/api/**` im Branch `spike/go-api` (Go-Prototyp)
- `apps/api/**` im Branch `spike/micronaut-api` (Micronaut-Prototyp)
- `Makefile`, `Dockerfile`, Build-Konfiguration jeweils pro Branch

Bewusst nicht betroffen:

- `apps/dashboard`, `packages/player-sdk`, `services/`, `examples/`,
  `observability/`, `deploy/` — werden erst im MVP `0.1.0` aufgebaut
- `spec/lastenheft.md` — Inhalt wird erst durch ADR auf `1.0.0` gehoben,
  nicht durch den Spike selbst
- `README.md` — Tech-Overview wird nach ADR aktualisiert, nicht währenddessen

---

## 9. Risiken und Gegenmaßnahmen (SP-33)

### 9.1 Scope-Creep im Erststack (SP-34)

Risiko:

- Der Erstprototyp wird "schnell noch um Persistenz / Auth / Multi-Tenant
  erweitert", weil es im jeweiligen Stack einfach scheint.
- Der Vergleich wird verzerrt.

Gegenmaßnahme:

- Spec §8 ("explizit nicht zum Scope") wird vor jedem Commit konsultiert.
- Pull-Request-Beschreibung pro Prototyp listet umgesetzten Muss-Scope.
- Bonus-Punkte werden separat im Bewertungsbogen ausgewiesen.

### 9.2 Erfahrungs-Bias zugunsten eines Stacks (SP-35)

Risiko:

- Geübterer Stack erzeugt schnelleren, eleganteren Code; Vergleich
  wirkt einseitig.

Gegenmaßnahme:

- Bias wird ausdrücklich im ADR notiert, nicht ausgeglichen.
- Reihenfolge darf bewusst gedreht werden (§4.4).
- `Contributor-Fit` (10%) zielt auf das OSS-Umfeld, nicht auf Eigen-Skill.

### 9.3 Unklarer Sieger nach 5 Tagen (SP-36)

Risiko:

- Bewertung liefert kein klares Ergebnis; Versuchung, Spike zu verlängern.

Gegenmaßnahme:

- Entscheidungsregel §4.6 erzwingt eine Wahl.
- Tabu: "noch ein halber Tag", "ich brauche keinen Test", "OTel später".
- Wenn die Regel nicht greift, gewinnt der Stack mit weniger
  Infrastruktur-/Build-Komplexität.

### 9.4 OTel-Integration wird "später sauber gemacht" (SP-37)

Risiko:

- OTel-Ergonomie ist eine der wichtigsten Bewertungskategorien (15%).
  Wird sie übersprungen, bleibt der Spike ohne Aussagekraft.

Gegenmaßnahme:

- minimaler OTel-Setup ist Muss-Scope, nicht Bonus.
- Bewertung von OTel ohne Code-Berührung ist ungültig.

### 9.5 Coexistenz beider Prototypen "für später" (SP-38)

Risiko:

- "Ich behalte beide Stacks und entscheide nach dem MVP."
- Architektur driftet, Vertragsdoppelpflege entsteht.

Gegenmaßnahme:

- §4.3 verbietet Coexistenz.
- Sieger-Branch wird Basis für `apps/api`, unterlegener Branch wird
  archiviert oder gelöscht.

### 9.6 Spec-Drift während des Spikes (SP-39)

Risiko:

- Während der Implementierung entstehen Erkenntnisse, die nur im
  zweiten Stack einfließen.

Gegenmaßnahme:

- Vertragsänderungen müssen in beiden Prototypen identisch landen.
- Änderungen werden im Spike-Protokoll begründet.
- Datum des letzten API-Kontrakt-Commits steht im ADR.

---

## 10. Definition of Done für den Spike (SP-40)

Der Spike ist abgeschlossen, wenn:

- `spec/backend-api-contract.md` auf `main` existiert
- Branch `spike/go-api` den Muss-Scope erfüllt und alle Pflichttests aus
  §7.1 grün hat **oder** das Scheitern inklusive fehlender bzw.
  fehlschlagender Tests im Spike-Protokoll dokumentiert ist
- Branch `spike/micronaut-api` den Muss-Scope erfüllt und alle Pflichttests
  aus §7.1 grün hat **oder** das Scheitern inklusive fehlender bzw.
  fehlschlagender Tests im Spike-Protokoll dokumentiert ist
- Bewertungsbogen (Spec §16) für beide Prototypen vollständig in
  `docs/adr/0001-backend-stack.md` ausgefüllt ist
- Messwertbogen (Spec §17) für beide Prototypen vollständig in
  `docs/adr/0001-backend-stack.md` ausgefüllt ist
- `docs/spike/backend-stack-results.md` enthält Live-Notizen pro
  Bewertungskategorie und protokollierte Vertragsänderungen
- Reihenfolge-Bias im ADR dokumentiert ist
- `docs/adr/0001-backend-stack.md` mit klarer Stack-Entscheidung,
  Begründung in 2–3 Sätzen und Konsequenzen-Abschnitt existiert
- finale Commit-Hashes beider Prototypen im ADR genannt sind
- unterlegener Branch ist gelöscht oder als Tag
  `spike/backend-stack-loser-YYYYMMDD` archiviert
- du **nicht** das Gefühl hast, "noch eine Runde recherchieren" zu müssen

Letzter Punkt ist der wichtigste. Ein bestandener Spike beseitigt das
Gefühl der Unsicherheit; tut er das nicht, war die Bewertung zu zögerlich
oder das Raster passt nicht zum Kontext. Das wird durch explizite
Neubewertung gelöst, nicht durch mehr Code.

---

## 11. Anschluss an MVP `0.1.0` (SP-41)

Nach abgeschlossenem Spike beginnt die MVP-Implementierung gemäß
Lastenheft §17 Schritt 2 ff.

Verbindliche Folge:

- der Sieger-Branch wird zum `apps/api`-Skelett ausgebaut, nicht neu
  geschrieben
- Lastenheft wird auf Version `1.0.0` gehoben:
  - Header und §10.1 mit getroffener Entscheidung füllen
  - §9.1 in den Vergangenheitsmodus setzen
  - Backend-Entscheidung aus den offenen Punkten in §16.2 entfernen
- README dokumentiert den gewählten Stack im Tech-Overview
- erst danach werden Dashboard (Schritt 3), Player-SDK (Schritt 4),
  Docker-Lab (Schritt 5) und Observability (Schritt 6) angegangen

Phase-1-Risiken aus Bewertungskategorie "Absehbare Phase-2-Risiken"
werden im ADR dokumentiert und in einen ersten Issue-Backlog überführt.

Lessons-learned-Hinweise aus dem Vergleich mit `d-migrate` (für
0.1.0+, nicht im Spike-Scope):

- Coverage mit Kover (Kotlin) bzw. `go test -cover` (Go) ab MVP
  einführen; `koverVerify` mit Mindestschwelle (d-migrate: 90% Root,
  modulweise abgestuft).
- Per-Modul `detekt-baseline.xml` beim Übergang zu Multi-Modul-
  Mono-Repo erzeugen — pragmatischer Pfad bei wachsendem Bestand.
- CI-Workflow uploadet Test-Results, Coverage-Reports und
  detekt-Reports als Artifacts (Pattern aus
  `d-migrate/.github/workflows/build.yml`).
- `outputs.cacheIf { false }` auf Test-Tasks setzen, damit Coverage-
  Counter nicht aus stalem Build-Cache kommen.
- Bei wachsender Codebase `apps/api/` von Single-Modul zu
  Gradle-Multi-Modul aufteilen: jedes `hexagon/<x>/` und
  `adapters/<driving|driven>/<x>/` wird ein eigenes Sub-Modul mit
  eigener `build.gradle.kts`. Vorteil: Compile-Time-Enforcement der
  Hexagon-Boundaries via Modul-Dependencies (Pattern aus
  `d-migrate/settings.gradle.kts`). Im Spike bewusst Single-Modul,
  damit das 2-Tage-Budget pro Prototyp realistisch bleibt.

---

## 12. Verbindliche Modul- und Paketstruktur (SP-42)

### 12.1 Go-Prototyp (SP-43)

```text
apps/api/
├── cmd/
│   └── api/
│       └── main.go
├── hexagon/
│   ├── domain/
│   │   ├── playback_event.go
│   │   ├── stream_session.go
│   │   ├── project.go
│   │   └── project_token.go
│   ├── port/
│   │   ├── driving/
│   │   │   └── playback_event_inbound.go
│   │   └── driven/
│   │       ├── event_repository.go
│   │       └── metrics_publisher.go
│   └── application/
│       └── register_playback_event_batch.go
├── adapters/
│   ├── driving/
│   │   └── http/
│   │       ├── handler.go
│   │       ├── auth.go
│   │       └── rate_limit.go
│   └── driven/
│       ├── persistence/
│       │   └── inmemory_event_repository.go
│       ├── telemetry/
│       │   └── otel.go
│       └── metrics/
│           └── prometheus_publisher.go
├── go.mod
├── Dockerfile
└── Makefile
```

`hexagon/` und `adapters/` liegen direkt unter `apps/api/`, kein
`internal/`-Wrapper. Architektur ist beim ersten `ls apps/api/`
sichtbar. Module-Path-Konvention:
`github.com/<owner>/m-trace/apps/api` — finaler Owner-Pfad bleibt
offen bis zur Repo-Erstellung auf GitHub.

### 12.2 Micronaut-Prototyp (Kotlin) (SP-44)

```text
apps/api/
├── hexagon/
│   ├── domain/
│   ├── port/
│   │   ├── driving/
│   │   └── driven/
│   └── application/
├── adapters/
│   ├── driving/
│   │   └── http/
│   │       └── Application.kt   # Micronaut-Bootstrap = Inbound-Adapter
│   └── driven/
│       ├── persistence/
│       ├── telemetry/
│       └── metrics/
├── resources/
│   ├── application.yml
│   └── logback.xml
├── test/
│   └── (Spiegelung der hexagon/-/adapters/-Schichten)
├── build.gradle.kts
├── gradle.properties
├── detekt.yml
├── Dockerfile
└── Makefile
```

`hexagon/` und `adapters/` liegen direkt unter `apps/api/`; kein
`src/main/kotlin/dev/mtrace/api/`-Wrapper. Architektur ist beim
ersten `ls apps/api/` sichtbar — symmetrisch zu §12.1 (Go).

Voraussetzung in `build.gradle.kts` (custom Source Sets):

```kotlin
sourceSets {
    main {
        kotlin {
            srcDirs("hexagon", "adapters")
        }
        resources {
            srcDirs("resources")
        }
    }
    test {
        kotlin {
            srcDirs("test")
        }
    }
}
```

Package-Namen bleiben d-migrate-konform: `dev.mtrace.api.hexagon.domain`,
`dev.mtrace.api.adapters.driving.http`, ... — die Sprache erkennt
Pakete unabhängig vom Verzeichnispfad, sobald `srcDirs` korrekt sind.

Application.kt liegt unter `adapters/driving/http/`, weil der
HTTP-Server der Inbound-Adapter ist (Hexagon-Konvention; analog
zu d-migrate `adapters/driving/cli/.../Main.kt`).

Kein `gradle-wrapper.jar` und kein `gradlew`-Script im Repo: der
Spike läuft Docker-only (siehe §14.11 und Spec §6.11). Das
Build-Image `gradle:8.12-jdk21` bringt Gradle direkt mit.

Kotlin-Package-Konvention: `dev.mtrace.api.*`. Group-Id im Gradle-Build:
`dev.mtrace`. `build.gradle.kts` aktiviert die Plugins
`org.jetbrains.kotlin.jvm` (2.1.x), `io.micronaut.application` und
`io.gitlab.arturbosch.detekt` (1.23+). `detekt.yml` enthält die
Lint-Konfiguration.

`gradle.properties` hält Versionen zentral (Pattern aus d-migrate):

```properties
kotlin.code.style=official
org.gradle.parallel=true
org.gradle.caching=true
org.gradle.jvmargs=-Xmx4g

kotlinVersion=2.1.20
micronautVersion=4.7.x
kotestVersion=6.1.x
mockkVersion=1.14.x
logbackVersion=1.5.x
detektVersion=1.23.8
```

Konkrete Patch-Levels werden beim Anlegen des Branches auf den dann
aktuellen Stand gesetzt; Spike-relevant sind die Major/Minor.

### 12.3 Gemeinsame Identifier-Konventionen (SP-45)

Verbindlich für beide Prototypen:

- HTTP-Header `X-MTrace-Token`: **Pflicht** im Spike-Muss-Scope. Trägt
  das Auth-Token gemäß Spec §6.4. Validiert in den Tests aus §7.1.
- HTTP-Header `X-MTrace-Project`: reservierter Konventionsname für
  CORS-Allowlist und spätere Multi-Tenant-Nutzung gemäß Lastenheft §8.5.
  **Im Spike optional** — die `project_id` kommt im Muss-Scope aus dem
  Event-Payload, nicht aus dem Header. Wird im API-Kontrakt als reserviert
  dokumentiert, aber nicht erzwungen.
- Prometheus-Metrik-Prefix: `mtrace_`
- OTel-Attribut-Prefix: `mtrace.*`
- Docker-Image-Tag: `m-trace-api-spike:<stack>` mit `<stack>` ∈
  {`go`, `micronaut`}. Beide Prototypen müssen koexistieren können,
  damit AP-3-Messungen (Image-Größe, Cold Start) reproduzierbar sind.
  Image-ID (`docker images --format '{{.ID}}'`) wird zusätzlich im
  Messwertbogen festgehalten.
- npm-Package-Name in Beispielen: `@m-trace/player-sdk`

Diese Konventionen sind im API-Kontrakt (`spec/backend-api-contract.md`)
zu fixieren und dürfen zwischen den Prototypen nicht abweichen.

---

## 13. Arbeitspaket-Abhängigkeiten und Test-Tabelle (SP-46)

### 13.1 Reihenfolge (SP-47)

```text
AP-0  ──►  AP-1  ──►  AP-2  ──►  AP-3
```

- AP-1 startet erst, wenn AP-0 nach `main` gemerged ist.
- AP-2 startet erst, wenn AP-1 abgeschlossen ist (Bewertungs-Notizen liegen
  vor).
- AP-3 startet erst, wenn AP-2 abgeschlossen ist.

Parallelarbeit ist nicht vorgesehen — der Spike ist Solo-Aufwand.

### 13.2 Pflichttests pro Arbeitspaket (SP-48)

| AP | Pflichttest |
|---|---|
| AP-0 | keine Code-Tests; Schema-Beispielpayload muss valides JSON sein |
| AP-1 | alle Tests aus §7.1 grün; `docker run` liefert HTTP 200 auf `/api/health` |
| AP-2 | identisch zu AP-1 |
| AP-3 | ADR-Datei existiert; Bewertungs-/Messwertbogen ausgefüllt |

---

## 14. Offene Entscheidungen mit Default-Empfehlung (SP-49)

### 14.1 Reihenfolge der Prototypen (SP-50)

- **Default**: Go zuerst, Micronaut zweit.
- Drehen erlaubt, wenn Micronaut-Erfahrung sonst zu großen Vorteil hätte.
- Festlegung im ADR-Abschnitt "Reihenfolge und Bias".

### 14.2 Routing-Library im Go-Prototyp (SP-51)

- **Default**: Standard-`net/http`.
- Wechsel zu `chi` zulässig, wenn Standard für die Pflicht-Endpunkte
  spürbar Boilerplate erzeugt.
- `gorilla/mux`, `echo`, `gin` bewusst ausgeschlossen — der Spike soll
  keinen Framework-Vergleich aufmachen, der die OTel-Bewertung verwässert.

### 14.3 Logging-Format im Micronaut-Prototyp (SP-52)

- **Default**: Logback mit JSON-Encoder (Logstash-Encoder oder
  `LogstashEncoder`-äquivalent).
- Strukturierte JSON-Logs mit `project_id`, `session_id`, `status_code`,
  `error_type` sind Pflicht.

### 14.4 ADR-Pfad (SP-53)

- **Default**: `docs/adr/0001-backend-stack.md`.
- Konsistent mit Lastenheft §17 Schritt 0 und Spike-Spec §6.
- Kein Wechsel zu alternativen ADR-Tools (adr-tools-Templates ja,
  Toolchain-Lock-In nein).

### 14.5 Archivierungsmodus für unterlegenen Branch (SP-54)

- **Default**: Tag `spike/backend-stack-loser-YYYYMMDD` setzen, danach
  Branch löschen.
- Reine Branch-Löschung erlaubt, wenn Disk-Space oder Repo-Hygiene das
  rechtfertigen — finale Commit-Hash steht im ADR und reicht für spätere
  Referenz.

### 14.6 Sprache und JDK-Version im Micronaut-Prototyp (SP-55)

- **Default**: Kotlin 2.x auf JDK 21 (LTS).
- Begründung: Die spätere `apps/api`-Implementierung ist in Kotlin
  geplant. Ein Java-Spike würde `Contributor-Fit` und subjektive
  Velocity falsch einordnen, weil produktive Wartung dann in einer
  anderen Sprache stattfände als die Spike-Bewertung.
- Konsequenz: Der Vergleich ist Go vs. Kotlin/Micronaut, nicht Go vs.
  Java/Micronaut. Diese bewusste Entscheidung wird im ADR notiert
  (Bewertungskategorien `Contributor-Fit` und `Subjektiver Spaß`).
- Java 21 als JVM-Plattform bleibt verbindlich (Spec §6.11). Kotlin
  kompiliert nach JVM-21-Bytecode; das Container-Basisimage
  (`eclipse-temurin:21-jre-alpine`) bleibt unverändert.
- Reine-Java-Variante: nur, wenn der Kotlin-Build im Spike
  unverhältnismäßig viel Reibung erzeugt (z. B. Gradle-Plugin-
  Inkompatibilität). Wechsel im ADR begründen.

### 14.7 Container-Basisimage (SP-56)

- **Go-Default**: `gcr.io/distroless/static-debian12`.
- **Micronaut-Default**: `eclipse-temurin:21-jre-alpine`; Distroless Java
  als Bonus.
- Image-Größe ist Bewertungskategorie, nicht Optimierungsziel des Spikes.

### 14.8 GitHub-Owner und Modul-Pfade (SP-57)

- **Default**: offen bis zur GitHub-Repo-Erstellung.
- Im Go-Prototyp wird `go.mod` zunächst mit Platzhalter-Path
  (`github.com/example/m-trace/apps/api`) initialisiert; Anpassung beim
  Übergang zu `main`.
- Im Micronaut-Prototyp ist Group-Id `dev.mtrace` direkt finalisierbar.

### 14.9 Linting (SP-58)

- **Soll-Vorgabe** (nicht Spec-Muss): jeder Prototyp stellt einen
  `make lint`-Befehl bereit, der einen Standard-Linter mit
  Default-Regelsatz ausführt.
- **Go**: `golangci-lint run ./...` mit den Default-Lintern (`govet`,
  `errcheck`, `staticcheck`, `unused`, `ineffassign`).
- **Kotlin**: `docker build --target lint` (intern
  `gradle --no-daemon detekt`) mit `buildUponDefaultConfig = true`
  und der mitgelieferten Default-Konfiguration als Startpunkt.
  Stage-Name ist `lint` (gemeinsam mit Go-Stack), nicht `detekt`.
- Empfohlene Gradle-Konfiguration für detekt (übernommen aus
  `d-migrate/build.gradle.kts`):
  - `tasks.named("check") { dependsOn("detekt") }` — `gradle build`
    bzw. `gradle check` führen detekt automatisch aus.
  - `tasks.withType<Test>().configureEach { dependsOn("detekt") }` —
    Lint läuft *vor* den Tests, damit Lint-Probleme schneller sichtbar
    werden als Test-Failures.
  - Reports: `html`, `xml`, `sarif` aktivieren. SARIF wird vom GitHub
    Security-Tab gelesen — null Aufwand, hoher Wert.
  - `parallel = true`, `allRules = false`, `ignoreFailures = false`.
- Beide Linter sind **Soll**, nicht Muss. Sie dürfen nicht den
  Pflicht-Test-Scope vergrößern. Ein roter Lint-Run blockiert weder
  AP-1 noch AP-2 vom DoD, fließt aber in die Bewertungskategorien
  `Build-Komplexität` und `Test-Velocity` ein.
- Begründung: ohne Linter-Pendant pro Stack wäre detekt ein
  einseitiger JVM-Bonus; mit Linter pro Stack wird Build-Ergonomie
  symmetrisch messbar.
- Custom Lint-Regeln, Suppressions oder Tooling-Ausbau sind im Spike
  ausgeschlossen — Default-Profil pur. Per-Modul `detekt-baseline.xml`
  ist im Spike (Greenfield) nicht nötig; in §11 als pragmatischer
  Migrationspfad bei Bestandsmodulen erwähnt.

### 14.10 Test-Framework im Micronaut-Prototyp (SP-59)

- **Default**: Kotest 6.x als Test-Runner, MockK als Mocking-Library,
  `@MicronautTest` für DI-Integration in Specs.
- Begründung:
  - Kotest ist Kotlin-idiomatischer (Spec-DSL, eingebautes
    Property-Testing) und reduziert Boilerplate gegenüber JUnit 5.
  - MockK ist für Kotlin-Klassen (z. B. `final` per Default) ohne
    Open-Workarounds nutzbar.
  - Konsistenz mit dem produktiven Kotlin-Stack `d-migrate`: gleicher
    Test-Stack erleichtert Cross-Project-Wartung und Contributor-Fit.
- JUnit 5 als Plattform bleibt eingebunden (Kotest läuft auf JUnit
  Platform), wird aber nicht direkt als Test-Stil verwendet.
- Falls `@MicronautTest` mit Kotest-Specs unverhältnismäßig viel
  Reibung erzeugt, ist ein Wechsel auf JUnit 5 + JUnit-Style erlaubt;
  Begründung im ADR.

### 14.11 Dockerfile-Struktur (verbindlich) (SP-60)

Der Spike ist Docker-only: alle Build-, Test- und Lint-Schritte
laufen über `docker build --target <stage>`. Der Dockerfile pro
Prototyp ist daher zentral und muss folgende Stages benennen:

- **`deps`**: kopiert nur Build-Metadaten und löst Abhängigkeiten auf.
  - Go: `go.mod`, `go.sum` → `RUN go mod download`
  - Kotlin: `build.gradle.kts`, `gradle.properties` → `RUN gradle
    --no-daemon resolveAllDependencies`. Der Task wird im Spike
    selbst angelegt (Snippet unten), Inspiration aus
    `d-migrate/build.gradle.kts:179`.
- **`compile`**: kopiert Sources und kompiliert.
  - Go: `RUN go build ./...`
  - Kotlin: `RUN gradle --no-daemon classes`
- **`lint`** (Soll, siehe §14.9):
  - Go: `RUN golangci-lint run ./...`
  - Kotlin: `RUN gradle --no-daemon detekt`
- **`test`**: führt die Pflichttests aus §7.1 aus.
  - Go: `RUN go test ./...`
  - Kotlin: `RUN gradle --no-daemon test`
- **`build`**: erzeugt das finale Artefakt.
  - Go: Binary
  - Kotlin: Distribution via `gradle installDist` oder ähnlich
- **`runtime`**: minimales Final-Image.
  - Go: `gcr.io/distroless/static-debian12`
  - Kotlin: `eclipse-temurin:21-jre-alpine` (alternativ Distroless Java)

Make-Targets:

```makefile
test:    ; docker build --target test -t m-trace-api-spike:<stack>-test .
lint:    ; docker build --target lint -t m-trace-api-spike:<stack>-lint .
build:   ; docker build --target runtime -t m-trace-api-spike:<stack> .
```

Damit funktioniert ein frischer Clone ohne lokales Go, JDK oder
Gradle. Die `gradle-wrapper.jar` und das `gradlew`-Script werden
**nicht** versioniert.

Custom-Task `resolveAllDependencies` (Kotlin), Snippet in
`apps/api/build.gradle.kts`:

```kotlin
tasks.register("resolveAllDependencies") {
    group = "build setup"
    description = "Resolve all configurations to warm the Gradle " +
        "dependency cache (used by the deps Dockerfile stage)."
    doLast {
        configurations
            .filter { it.isCanBeResolved }
            .forEach { it.resolve() }
    }
}
```

Das Snippet ist gegenüber `d-migrate/build.gradle.kts:179` bewusst
einfacher gehalten (Single-Modul: keine `allprojects`-Iteration).
Bei späterer Multi-Modul-Aufteilung gemäß §11 muss die
`allprojects { ... }`-Variante übernommen werden.
