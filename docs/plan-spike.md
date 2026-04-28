# Implementierungsplan: Backend-Spike

> **Milestone**: 0.1.0 — OTel-native Local Demo (Vorstufe)
> **Phase**: Backend-Technologie-Entscheidung
> **Status**: Geplant
> **Referenz**: `docs/spike/0001-backend-stack.md`;
> `docs/lastenheft.md` §9.1, §10.1, §16.2, §17 Schritt 0;
> `docs/adr/0001-backend-stack.md` (entsteht);
> `README.md`.

---

## 1. Ziel

Der Backend-Spike liefert die technische Entscheidung zwischen Go und Micronaut
für `apps/api` durch zwei zeitlich begrenzte Prototypen mit identischem
Funktionsumfang. Er produziert keinen Produktionscode und kein
MVP-Skelett, sondern genug Vergleichsmaterial für eine **belastbare**
Entscheidung — auf Basis eigener Erfahrung, nicht weiterer Recherche.

Dieser Plan operationalisiert die Spike-Spezifikation aus
`docs/spike/0001-backend-stack.md`. Er bricht sie in datierte Arbeitspakete,
Abnahmekriterien, Verifikationsschritte und eine Definition of Done auf.

Konkrete Ergebnisse nach Abschluss (mit verbindlichen Pfaden):

- API-Kontrakt-Datei `docs/spike/backend-api-contract.md`
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

## 2. Ausgangslage

### 2.1 Lastenheft hält Backend-Entscheidung bewusst offen

`docs/lastenheft.md` §9.1 nennt zwei zulässige Optionen — Go oder Micronaut —
und bindet die Wahl explizit an einen technischen Spike. §10.1 listet die
Mindestanforderungen unabhängig vom Stack. §16.2 führt die endgültige
Backend-Technologie als offene Entscheidung. §17 Schritt 0 fordert den Spike
als Voraussetzung jeder MVP-Implementierung.

### 2.2 Spike-Spezifikation ist normativ

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

### 2.3 Repository ist noch leer

Aktueller Stand:

- `README.md` mit Projekt-Vision
- `docs/lastenheft.md`
- `docs/spike/0001-backend-stack.md`
- `docs/plan-spike.md` (dieses Dokument)
- kein `apps/`, kein `packages/`, kein `services/`
- keine CI, kein `Makefile`, kein `docker-compose.yml`

Verbindliche Folge:

- Spike-Branches arbeiten gegen einen leeren Working Tree
- der Plan darf keine bestehenden Komponenten als gegeben annehmen
- Mono-Repo-Tooling (pnpm-Workspace, Gradle, Go-Modules) wird erst im
  jeweiligen Prototyp eingeführt, nicht im Spike-Vorlauf

---

## 3. Scope für den Spike

### 3.1 In Scope

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

### 3.2 Bewusst nicht Teil des Spikes

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

## 4. Leitentscheidungen

### 4.1 API-Kontrakt vor jedem Implementierungsstart

Vor `git checkout -b spike/go-api` wird `docs/spike/backend-api-contract.md`
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

### 4.2 Harte Zeitgrenze, keine dritte Spike-Runde

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

### 4.3 Identischer Scope, keine Coexistenz danach

Beide Prototypen liefern denselben Muss-Scope. Erweiterungen "weil es
schnell geht" oder Auslassungen "weil es im anderen Stack einfacher ist"
sind verboten. Bonusfunktionen dürfen den Muss-Scope nicht gefährden.

Verbindliche Folge:

- Sieger-Branch wird Basis für `apps/api`
- unterlegener Branch wird nicht weiterentwickelt
- unterlegener Branch wird gelöscht oder als Tag
  `spike/backend-stack-loser-YYYYMMDD` archiviert
- der finale Commit-Hash beider Prototypen wird im ADR referenziert

### 4.4 Reihenfolge-Bias wird dokumentiert, nicht ausgeglichen

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

### 4.5 Hexagon nur dort, wo Fachlogik entsteht

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

### 4.6 Eine Entscheidung, keine "beide Optionen sind gut"

Die Entscheidungsregel aus Spec §10 ist verbindlich:

1. ≥ 10 gewichtete Prozentpunkte Vorsprung → Sieger
2. < 10 Prozentpunkte → `Contributor-Fit` entscheidet
3. Gleichstand bei `Contributor-Fit` → subjektive Wartbarkeit nach 2 Tagen
4. weiterhin gleich → der Stack mit weniger Infrastruktur-/Build-Komplexität
5. keine weitere Spike-Runde

Das ADR muss eine *gewählte* Option benennen.

---

## 5. Zielarchitektur

### 5.1 Branch-Layout im Repository

| Branch | Zweck | Lebensdauer |
|---|---|---|
| `main` | Lastenheft, Spike-Doc, Plan, README | dauerhaft |
| `spike/go-api` | Go-Prototyp | bis Auswertung; danach Sieger oder Archiv |
| `spike/micronaut-api` | Micronaut-Prototyp | bis Auswertung; danach Sieger oder Archiv |
| `spike/backend-stack-loser-YYYYMMDD` | optionales Archiv | dauerhaft, kein aktiver Branch |

`main` enthält keine Prototyp-Sourcen. Die API-Kontrakt-Datei
`docs/spike/backend-api-contract.md` wird vor dem ersten Implementierungs-
Branch nach `main` gemerged.

### 5.2 Mindeststruktur pro Prototyp

Verbindlich ist die *logische* Schichtung, nicht der konkrete Dateipfad.
Beide Prototypen müssen folgende Schichten erkennbar voneinander trennen:

- `hexagon/domain/` — frameworkfreie Fachobjekte
- `hexagon/port/in/` — Eingangs-Ports (Use-Case-Schnittstellen)
- `hexagon/port/out/` — Ausgangs-Ports (Repository, Publisher)
- `hexagon/application/` — Use Cases / Application Services
- `adapters/in/http/` — HTTP-Controller
- `adapters/out/persistence/` — Event-Repository
- `adapters/out/telemetry/` — OTel-Setup
- `adapters/out/metrics/` — Prometheus-Publisher

Die Abbildung auf Dateipfade folgt der jeweiligen Sprach-Konvention:

- **Go**: nutzt `apps/api/internal/hexagon/...` und
  `apps/api/internal/adapters/...`; Entry-Point unter
  `apps/api/cmd/api/main.go`. Konkreter Tree in §12.1.
- **Micronaut/Kotlin**: nutzt `apps/api/src/main/kotlin/dev/mtrace/api/hexagon/...`
  und `.../adapters/...`. Konkreter Tree in §12.2.

Stack-spezifische Builddateien (`go.mod`, `build.gradle.kts`,
`gradle/wrapper/`, `Makefile`) und `Dockerfile`/`README.md` liegen
ergänzend im jeweiligen `apps/api/`-Verzeichnis.

Verbindliche Folge:

- Vergleichende LoC-Messungen (siehe §7.2) zählen die *logischen*
  Schichten, nicht ein bestimmtes Verzeichnis. `cloc` läuft pro Prototyp
  gegen den jeweiligen Domain- bzw. Adapter-Pfad gemäß §12.
- Eine zusätzliche `internal/` oder `src/main/kotlin/...`-Ebene gilt nicht
  als Hexagon-Verletzung, solange die Schichten klar bleiben.

### 5.3 Domain-Objekte

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

### 5.4 Metriken-Vertrag

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

### 5.5 OTel-Mindestumfang

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

## 6. Arbeitspakete

### 6.1 AP-0: API-Kontrakt fixieren

Maximalbudget: 0,5 Tag.

Aufgaben:

- Datei `docs/spike/backend-api-contract.md` erstellen
- Endpunkte, Pfade, Methoden, Statuscodes festschreiben
- Beispielpayload für `POST /api/playback-events` festschreiben
- Header-Definitionen (`X-MTrace-Token`, `Retry-After`) festschreiben
- Metriknamen festschreiben (siehe §5.4)
- Validierungsregeln und Fehlerfälle festschreiben
- Pflichttests (Happy Path und Fehlerfälle) festschreiben
- Datei nach `main` mergen

Abnahme:

- Datei existiert auf `main`
- jeder Endpunkt hat Beispielpayload, Statuscode-Liste und mindestens einen
  Testfall
- alle Metriken der Spec sind genannt
- nach Merge nach `main` werden keine Vertragsänderungen mehr einseitig
  vorgenommen

### 6.2 AP-1: Go-Prototyp

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
- `make test` und `docker run -p 8080:8080 m-trace-api-spike:go` müssen laufen
- `make lint` ruft `golangci-lint run ./...` auf (Soll-Aufgabe; siehe §14.9)
- Spike-Notizen pro Bewertungskategorie in
  `docs/spike/backend-stack-results.md` führen

Abnahme:

- alle Pflicht-Endpunkte liefern erwartete Statuscodes
- alle Pflichttests grün
- Docker-Image baut und startet, `/api/health` liefert HTTP 200
- alle Pflichtmetriken sichtbar an `/api/metrics`
- OTel ist mindestens einmal im Code berührt
- Notizen zu Bewertungskategorien sind in
  `docs/spike/backend-stack-results.md` festgehalten

### 6.3 AP-2: Micronaut-Prototyp

Maximalbudget: 2 Tage. Branch: `spike/micronaut-api`.

Stack:

- Micronaut 4.x
- Kotlin 2.x auf JDK 21 (siehe §14.6 für Begründung)
- OTel via Micronaut-OpenTelemetry-Modul oder direkte SDK-Nutzung
- Tests mit `@MicronautTest` und JUnit 5 (alternativ Kotest)
- Logging mit Logback, optional Logstash-Encoder
- Linting: `detekt` (Soll, siehe §14.9)

Aufgaben:

- Branch von `main` ziehen (nicht von `spike/go-api`)
- Gradle-Wrapper, `build.gradle.kts`, Micronaut-Application initialisieren
- Domain, Use Case, Ports, Adapter gemäß §5.2/§5.3 anlegen
- HTTP-Controller, In-Memory-Repository, Prometheus-Publisher, In-Memory-
  Rate-Limiter implementieren
- minimaler OTel-Setup integrieren
- strukturierte JSON-Logs einrichten
- Dockerfile mit Multi-Stage-Build (`eclipse-temurin:21-jre-alpine` oder
  Distroless Java)
- Pflichttests gemäß Spec §6.12 schreiben
- Image als `m-trace-api-spike:micronaut` taggen
- `./gradlew test` und `docker run -p 8080:8080 m-trace-api-spike:micronaut`
  müssen laufen
- `make lint` ruft `./gradlew detekt` auf (Soll-Aufgabe; siehe §14.9)
- Spike-Notizen pro Bewertungskategorie in
  `docs/spike/backend-stack-results.md` ergänzen
- Spike-Notizen pro Bewertungskategorie führen

Abnahme:

- identisch zu AP-1, mit `./gradlew test` statt `go test ./...`

### 6.4 AP-3: Auswertung und ADR

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

## 7. Verifikationsstrategie

### 7.1 Pflichttests pro Prototyp

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

oder stack-spezifisch (`go test ./...` bzw. `./gradlew test`).

### 7.2 Mess-Punkte

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

Globale Caches (`~/go/pkg/mod`, `~/.gradle`) sind durch andere Projekte
verfälscht und nicht vergleichbar. Der Spike misst pro Prototyp einen
isolierten Cache aus einem leeren Branch-Clone:

- **Go** (aus `apps/api/`):  
  `cd apps/api && GOMODCACHE="$PWD/.gomodcache" go build ./... && du -sh .gomodcache`
- **Gradle** (aus `apps/api/`):  
  `cd apps/api && ./gradlew --gradle-user-home "$PWD/.gradle-user-home" build && du -sh .gradle-user-home`

Ergebnis liegt damit unter `apps/api/.gomodcache/` bzw.
`apps/api/.gradle-user-home/`. Beide Pfade sind im `.gitignore` gemappt
(Pattern ohne führenden Slash matcht in jedem Unterverzeichnis).

Vor der Messung muss der jeweilige Cache leer sein (frischer Branch-Clone
gilt als sauber). `.gomodcache/` und `.gradle-user-home/` sind im
`.gitignore` ausgenommen.

### 7.3 Bewertungsraster

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

### 7.4 Keine Lasttests, keine Produktionssimulation

Performance-Optimierung über "läuft mit Tests durch" hinaus ist nicht Teil
des Spikes. Lasttests, Profiling-Sessions und JVM-Tuning sind ausgeschlossen.

---

## 8. Betroffene Codebasis

Voraussichtlich erzeugt:

- `docs/spike/backend-api-contract.md` (neu, auf `main`)
- `docs/adr/0001-backend-stack.md` (neu, auf `main`)
- `apps/api/**` im Branch `spike/go-api` (Go-Prototyp)
- `apps/api/**` im Branch `spike/micronaut-api` (Micronaut-Prototyp)
- `Makefile`, `Dockerfile`, Build-Konfiguration jeweils pro Branch

Bewusst nicht betroffen:

- `apps/dashboard`, `packages/player-sdk`, `services/`, `examples/`,
  `observability/`, `deploy/` — werden erst im MVP `0.1.0` aufgebaut
- `docs/lastenheft.md` — Inhalt wird erst durch ADR auf `1.0.0` gehoben,
  nicht durch den Spike selbst
- `README.md` — Tech-Overview wird nach ADR aktualisiert, nicht währenddessen

---

## 9. Risiken und Gegenmaßnahmen

### 9.1 Scope-Creep im Erststack

Risiko:

- Der Erstprototyp wird "schnell noch um Persistenz / Auth / Multi-Tenant
  erweitert", weil es im jeweiligen Stack einfach scheint.
- Der Vergleich wird verzerrt.

Gegenmaßnahme:

- Spec §8 ("explizit nicht zum Scope") wird vor jedem Commit konsultiert.
- Pull-Request-Beschreibung pro Prototyp listet umgesetzten Muss-Scope.
- Bonus-Punkte werden separat im Bewertungsbogen ausgewiesen.

### 9.2 Erfahrungs-Bias zugunsten eines Stacks

Risiko:

- Geübterer Stack erzeugt schnelleren, eleganteren Code; Vergleich
  wirkt einseitig.

Gegenmaßnahme:

- Bias wird ausdrücklich im ADR notiert, nicht ausgeglichen.
- Reihenfolge darf bewusst gedreht werden (§4.4).
- `Contributor-Fit` (10%) zielt auf das OSS-Umfeld, nicht auf Eigen-Skill.

### 9.3 Unklarer Sieger nach 5 Tagen

Risiko:

- Bewertung liefert kein klares Ergebnis; Versuchung, Spike zu verlängern.

Gegenmaßnahme:

- Entscheidungsregel §4.6 erzwingt eine Wahl.
- Tabu: "noch ein halber Tag", "ich brauche keinen Test", "OTel später".
- Wenn die Regel nicht greift, gewinnt der Stack mit weniger
  Infrastruktur-/Build-Komplexität.

### 9.4 OTel-Integration wird "später sauber gemacht"

Risiko:

- OTel-Ergonomie ist eine der wichtigsten Bewertungskategorien (15%).
  Wird sie übersprungen, bleibt der Spike ohne Aussagekraft.

Gegenmaßnahme:

- minimaler OTel-Setup ist Muss-Scope, nicht Bonus.
- Bewertung von OTel ohne Code-Berührung ist ungültig.

### 9.5 Coexistenz beider Prototypen "für später"

Risiko:

- "Ich behalte beide Stacks und entscheide nach dem MVP."
- Architektur driftet, Vertragsdoppelpflege entsteht.

Gegenmaßnahme:

- §4.3 verbietet Coexistenz.
- Sieger-Branch wird Basis für `apps/api`, unterlegener Branch wird
  archiviert oder gelöscht.

### 9.6 Spec-Drift während des Spikes

Risiko:

- Während der Implementierung entstehen Erkenntnisse, die nur im
  zweiten Stack einfließen.

Gegenmaßnahme:

- Vertragsänderungen müssen in beiden Prototypen identisch landen.
- Änderungen werden im Spike-Protokoll begründet.
- Datum des letzten API-Kontrakt-Commits steht im ADR.

---

## 10. Definition of Done für den Spike

Der Spike ist abgeschlossen, wenn:

- `docs/spike/backend-api-contract.md` auf `main` existiert
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

## 11. Anschluss an MVP `0.1.0`

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

---

## 12. Verbindliche Modul- und Paketstruktur

### 12.1 Go-Prototyp

```text
apps/api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── hexagon/
│   │   ├── domain/
│   │   │   ├── playback_event.go
│   │   │   ├── stream_session.go
│   │   │   ├── project.go
│   │   │   └── project_token.go
│   │   ├── port/
│   │   │   ├── in/
│   │   │   │   └── playback_event_inbound.go
│   │   │   └── out/
│   │   │       ├── event_repository.go
│   │   │       └── metrics_publisher.go
│   │   └── application/
│   │       └── register_playback_event_batch.go
│   └── adapters/
│       ├── in/
│       │   └── http/
│       │       ├── handler.go
│       │       ├── auth.go
│       │       └── rate_limit.go
│       └── out/
│           ├── persistence/
│           │   └── inmemory_event_repository.go
│           ├── telemetry/
│           │   └── otel.go
│           └── metrics/
│               └── prometheus_publisher.go
├── go.mod
├── Dockerfile
└── Makefile
```

Module-Path-Konvention: `github.com/<owner>/m-trace/apps/api` —
finaler Owner-Pfad bleibt offen bis zur Repo-Erstellung auf GitHub.

### 12.2 Micronaut-Prototyp (Kotlin)

```text
apps/api/
├── src/
│   ├── main/
│   │   ├── kotlin/
│   │   │   └── dev/mtrace/api/
│   │   │       ├── Application.kt
│   │   │       ├── hexagon/
│   │   │       │   ├── domain/
│   │   │       │   ├── port/
│   │   │       │   │   ├── in/
│   │   │       │   │   └── out/
│   │   │       │   └── application/
│   │   │       └── adapters/
│   │   │           ├── in/
│   │   │           │   └── http/
│   │   │           └── out/
│   │   │               ├── persistence/
│   │   │               ├── telemetry/
│   │   │               └── metrics/
│   │   └── resources/
│   │       ├── application.yml
│   │       └── logback.xml
│   └── test/
│       └── kotlin/
│           └── dev/mtrace/api/
├── build.gradle.kts
├── detekt.yml
├── gradle/
│   └── wrapper/
├── gradlew
├── Dockerfile
└── Makefile
```

Kotlin-Package-Konvention: `dev.mtrace.api.*`. Group-Id im Gradle-Build:
`dev.mtrace`. `build.gradle.kts` aktiviert die Plugins
`org.jetbrains.kotlin.jvm`, `io.micronaut.application` und
`io.gitlab.arturbosch.detekt`. `detekt.yml` enthält die Lint-Konfiguration.

### 12.3 Gemeinsame Identifier-Konventionen

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

Diese Konventionen sind im API-Kontrakt (`docs/spike/backend-api-contract.md`)
zu fixieren und dürfen zwischen den Prototypen nicht abweichen.

---

## 13. Arbeitspaket-Abhängigkeiten und Test-Tabelle

### 13.1 Reihenfolge

```text
AP-0  ──►  AP-1  ──►  AP-2  ──►  AP-3
```

- AP-1 startet erst, wenn AP-0 nach `main` gemerged ist.
- AP-2 startet erst, wenn AP-1 abgeschlossen ist (Bewertungs-Notizen liegen
  vor).
- AP-3 startet erst, wenn AP-2 abgeschlossen ist.

Parallelarbeit ist nicht vorgesehen — der Spike ist Solo-Aufwand.

### 13.2 Pflichttests pro Arbeitspaket

| AP | Pflichttest |
|---|---|
| AP-0 | keine Code-Tests; Schema-Beispielpayload muss valides JSON sein |
| AP-1 | alle Tests aus §7.1 grün; `docker run` liefert HTTP 200 auf `/api/health` |
| AP-2 | identisch zu AP-1 |
| AP-3 | ADR-Datei existiert; Bewertungs-/Messwertbogen ausgefüllt |

---

## 14. Offene Entscheidungen mit Default-Empfehlung

### 14.1 Reihenfolge der Prototypen

- **Default**: Go zuerst, Micronaut zweit.
- Drehen erlaubt, wenn Micronaut-Erfahrung sonst zu großen Vorteil hätte.
- Festlegung im ADR-Abschnitt "Reihenfolge und Bias".

### 14.2 Routing-Library im Go-Prototyp

- **Default**: Standard-`net/http`.
- Wechsel zu `chi` zulässig, wenn Standard für die Pflicht-Endpunkte
  spürbar Boilerplate erzeugt.
- `gorilla/mux`, `echo`, `gin` bewusst ausgeschlossen — der Spike soll
  keinen Framework-Vergleich aufmachen, der die OTel-Bewertung verwässert.

### 14.3 Logging-Format im Micronaut-Prototyp

- **Default**: Logback mit JSON-Encoder (Logstash-Encoder oder
  `LogstashEncoder`-äquivalent).
- Strukturierte JSON-Logs mit `project_id`, `session_id`, `status_code`,
  `error_type` sind Pflicht.

### 14.4 ADR-Pfad

- **Default**: `docs/adr/0001-backend-stack.md`.
- Konsistent mit Lastenheft §17 Schritt 0 und Spike-Spec §6.
- Kein Wechsel zu alternativen ADR-Tools (adr-tools-Templates ja,
  Toolchain-Lock-In nein).

### 14.5 Archivierungsmodus für unterlegenen Branch

- **Default**: Tag `spike/backend-stack-loser-YYYYMMDD` setzen, danach
  Branch löschen.
- Reine Branch-Löschung erlaubt, wenn Disk-Space oder Repo-Hygiene das
  rechtfertigen — finale Commit-Hash steht im ADR und reicht für spätere
  Referenz.

### 14.6 Sprache und JDK-Version im Micronaut-Prototyp

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

### 14.7 Container-Basisimage

- **Go-Default**: `gcr.io/distroless/static-debian12`.
- **Micronaut-Default**: `eclipse-temurin:21-jre-alpine`; Distroless Java
  als Bonus.
- Image-Größe ist Bewertungskategorie, nicht Optimierungsziel des Spikes.

### 14.8 GitHub-Owner und Modul-Pfade

- **Default**: offen bis zur GitHub-Repo-Erstellung.
- Im Go-Prototyp wird `go.mod` zunächst mit Platzhalter-Path
  (`github.com/example/m-trace/apps/api`) initialisiert; Anpassung beim
  Übergang zu `main`.
- Im Micronaut-Prototyp ist Group-Id `dev.mtrace` direkt finalisierbar.

### 14.9 Linting

- **Soll-Vorgabe** (nicht Spec-Muss): jeder Prototyp stellt einen
  `make lint`-Befehl bereit, der einen Standard-Linter mit
  Default-Regelsatz ausführt.
- **Go**: `golangci-lint run ./...` mit den Default-Lintern (`govet`,
  `errcheck`, `staticcheck`, `unused`, `ineffassign`).
- **Kotlin**: `./gradlew detekt` mit der mitgelieferten Default-
  Konfiguration (`detekt.yml` aus dem Plugin-Default).
- Beide Linter sind **Soll**, nicht Muss. Sie dürfen nicht den
  Pflicht-Test-Scope vergrößern. Ein roter Lint-Run blockiert weder
  AP-1 noch AP-2 vom DoD, fließt aber in die Bewertungskategorien
  `Build-Komplexität` und `Test-Velocity` ein.
- Begründung: ohne Linter-Pendant pro Stack wäre detekt ein
  einseitiger JVM-Bonus; mit Linter pro Stack wird Build-Ergonomie
  symmetrisch messbar.
- Custom Lint-Regeln, Suppressions oder Tooling-Ausbau sind im Spike
  ausgeschlossen — Default-Profil pur.
