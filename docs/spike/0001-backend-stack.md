# Spike: Backend-Technologie-Entscheidung

**Status:** Plan  
**Version:** 0.2.0  
**Bezug:** Lastenheft v0.7.0, §9.1, §10.1  
**Outcome:** ADR `docs/adr/0001-backend-stack.md`

---

## 1. Ziel

Entscheidung zwischen Go und Micronaut für `apps/api` durch zwei zeitlich begrenzte Mini-Prototypen mit identischem Funktionsumfang.

Das Spike soll keine produktionsreife API liefern. Es soll genug Code produzieren, um eine **belastbare** Entscheidung zu treffen — auf Basis eigener Erfahrung, nicht weiterer Recherche oder Lektüre.

Wenn am Ende des Spikes nicht klar ist, welcher Stack besser passt, war der Scope zu klein, die Bewertung zu zögerlich oder die Entscheidungskriterien waren nicht scharf genug. Ein dritter Spike-Versuch ist nicht vorgesehen.

---

## 2. Zeitbudget

| Phase | Maximum |
|---|---:|
| API-Kontrakt fixieren | 0,5 Tag |
| Go-Prototyp | 2 Tage |
| Micronaut-Prototyp | 2 Tage |
| Auswertung und ADR | 0,5 Tage |

**Harte Grenze.** Wenn ein Prototyp in 2 Tagen den definierten Muss-Scope nicht erreicht, ist *das* die Erkenntnis — nicht ein Hinweis, mehr Zeit zu investieren.

---

## 3. Vorbereitender API-Kontrakt

Vor dem ersten Implementierungs-Branch wird ein kleiner API-Kontrakt erstellt.

Datei:

```text
spec/backend-api-contract.md
```

Dieser Kontrakt enthält:

- Endpunkte
- Beispielpayloads
- Statuscodes
- Header
- Metriknamen
- minimale Fehlerfälle
- Testfälle

Nach Start der Implementierungen darf der Kontrakt nicht mehr nachverhandelt werden, außer beide Prototypen werden identisch angepasst und die Änderung wird im Spike-Protokoll begründet.

Ziel dieser Vorphase ist, Reihenfolge-Bias zu reduzieren. Der zweite Stack darf nicht davon profitieren, dass unklare Anforderungen erst im ersten Prototyp entdeckt wurden.

---

## 4. Reihenfolge und Bias

Die geplante Reihenfolge ist:

1. Go-Prototyp
2. Micronaut-Prototyp

Diese Reihenfolge ist pragmatisch, aber nicht neutral. Der zweite Prototyp profitiert davon, dass Schema, Statuscodes und Edge Cases bereits einmal durchdacht wurden.

Der Bias wird im ADR ausdrücklich notiert.

Optional kann die Reihenfolge bewusst gedreht werden, wenn Micronaut aufgrund vorhandener Erfahrung einen unfairen Vorteil hätte. Wichtig ist nicht die perfekte wissenschaftliche Neutralität, sondern die explizite Dokumentation des Bias.

---

## 5. Identischer Scope

Beide Prototypen müssen denselben Scope liefern. Keine Erweiterungen "weil es schnell geht". Keine Auslassungen "weil es im anderen Stack einfacher ist".

Der Scope ist in **Muss** und **Bonus** getrennt.

Ein Prototyp ist nur gültig, wenn der Muss-Scope umgesetzt ist. Bonuspunkte fließen in die Bewertung ein, sind aber nicht erforderlich, damit der Prototyp vergleichbar bleibt.

---

## 6. Muss-Scope

### 6.1 HTTP-Endpunkte

| Methode | Pfad | Verhalten |
|---|---|---|
| `POST` | `/api/playback-events` | Batch von 1–100 Events annehmen |
| `GET` | `/api/health` | Liveness Check |
| `GET` | `/api/metrics` | Prometheus-Format |

### 6.2 Event-Schema, Wire Format

```json
{
  "schema_version": "1.0",
  "events": [
    {
      "event_name": "rebuffer_started",
      "project_id": "demo",
      "session_id": "01J...",
      "client_timestamp": "2026-04-28T12:00:00.000Z",
      "sequence_number": 42,
      "sdk": {
        "name": "@m-trace/player-sdk",
        "version": "0.1.0"
      }
    }
  ]
}
```

### 6.3 Validierung

| Fall | Erwartetes Verhalten |
|---|---|
| `schema_version: "1.0"` | akzeptiert |
| andere `schema_version` | HTTP 400 |
| fehlender oder falscher Header `X-MTrace-Token` | HTTP 401 |
| Event ohne Pflichtfelder | gesamter Batch HTTP 422 |
| mehr als 100 Events im Batch | HTTP 422 |
| Body größer als 256 KB | HTTP 413 |
| Rate Limit überschritten | HTTP 429 mit `Retry-After` |

### 6.4 Authentifizierung

Der Header `X-MTrace-Token` wird gegen eine hardcodierte Map geprüft:

```json
{
  "demo": "demo-token"
}
```

Regel:

- `project_id` im Event muss zum Token passen.
- Falsches Token führt zu HTTP 401.
- Unbekanntes Project führt zu HTTP 401.
- Es gibt keine dynamische Project-Verwaltung im Spike.

### 6.5 Domain-Logik

Die Domain muss frameworkfrei sein.

Domain-Objekte:

- `PlaybackEvent`
- `StreamSession`
- `Project`
- `ProjectToken`

Use Case:

- `RegisterPlaybackEventBatch`

Ports:

- `EventRepository`
- `MetricsPublisher`

Adapter:

- In-Memory Event Repository
- Prometheus Metrics Publisher

`StreamSession` hat einen einfachen Lifecycle:

```text
active
ended
```

Im Muss-Scope reicht es, Sessions anhand von `session_id` automatisch anzulegen und als `active` zu führen. Ein explizites Session-Ende ist Bonus.

### 6.6 Observability: Prometheus

Prometheus ist im Muss-Scope.

`/api/metrics` muss Prometheus-Format exponieren.

Pflichtmetriken:

| Metrik | Bedeutung |
|---|---|
| `mtrace_playback_events_total` | angenommene Playback-Events |
| `mtrace_invalid_events_total` | wegen Schema oder Validierung abgelehnte Events |
| `mtrace_rate_limited_events_total` | durch Rate Limit abgelehnte Events |
| `mtrace_dropped_events_total` | intern verworfene Events |

Hinweis:

`mtrace_dropped_events_total` muss im Spike nicht durch komplexe Backpressure entstehen. Es reicht, die Metrik vorzusehen und bei einem definierten internen Drop-Pfad zu inkrementieren, falls dieser implementiert wird.

### 6.7 Observability: OpenTelemetry

OpenTelemetry ist im Muss-Scope nur als **Ergonomie-Test** erforderlich, nicht als vollständige produktive Telemetrie-Pipeline.

Muss:

- minimaler OTel-Meter- oder Trace-Setup im Code
- Bewertung der OTel-Integration dokumentieren
- keine externe OTel-Infrastruktur erforderlich

Soll:

- OTel-Counter analog zu Prometheus-Metriken modellieren
- OTLP-Export vorbereiten oder exemplarisch aktivieren

Nicht erforderlich:

- Tempo
- OTel Collector
- produktive Collector-Konfiguration
- vollständige Trace-Korrelation

### 6.8 Logging

Muss:

- strukturierte JSON-Logs
- mindestens folgende Felder, sofern im Kontext vorhanden:
  - `project_id`
  - `session_id`
  - `status_code`
  - `error_type`

Soll:

- `trace_id`
- `request_id`

Wenn `trace_id` in einem Stack unverhältnismäßig viel Aufwand erzeugt, wird das als OTel-Ergonomie-Befund notiert und nicht durch Workarounds kaschiert.

### 6.9 Rate Limiting

- 100 Events pro Sekunde pro `project_id`
- Token Bucket oder vergleichbarer Ansatz
- Bei Überschreitung: HTTP 429
- `Retry-After`-Header muss gesetzt werden
- `mtrace_rate_limited_events_total` zählt abgelehnte Events

Es reicht ein In-Memory-Rate-Limiter. Verteiltes Rate Limiting ist nicht Teil des Spikes.

### 6.10 Persistenz

- In-Memory-Repository
- keine Persistenz auf Disk
- kein SQLite
- kein Redis
- kein ORM
- Daten überleben keinen Neustart — beabsichtigt
- concurrent-safe, aber nicht performance-optimiert

### 6.11 Build und Deployment

- **Docker-only Workflow**: alle Build-, Test- und Lint-Schritte
  laufen über `docker build --target <stage>`. Lokale Toolchains (Go,
  JDK, Gradle-Wrapper) sind nicht erforderlich.
- `Dockerfile` mit Multi-Stage Build pro Prototyp
- Build-Stages (Detailstruktur in `docs/planning/plan-spike.md` §14.11):
  `deps` → `compile` → `test` → `runtime`. Kotlin ergänzt zusätzlich
  eine `detekt`-Stage.
- Build-Tooling im Docker-Image:
  - Go: `golang:1.22` (oder vergleichbar) als Build-Image
  - Micronaut: `gradle:8.12-jdk21` als Build-Image — bringt Gradle
    direkt mit, daher kein checked-in `gradle-wrapper.jar` nötig
- Final Image:
  - Go: bevorzugt `gcr.io/distroless/static-debian12` oder vergleichbar
  - Micronaut: bevorzugt `eclipse-temurin:21-jre-alpine` oder Distroless Java
- `docker build` läuft ohne Mounts und ohne BuildKit-Spezialfeatures
- Image wird stack-spezifisch getaggt (`m-trace-api-spike:go` bzw.
  `m-trace-api-spike:micronaut`), damit beide Prototypen für die
  AP-3-Messungen koexistieren
- `docker run -p 8080:8080 m-trace-api-spike:<stack>` startet den Service
- `/api/health` liefert nach Start HTTP 200

### 6.12 Tests

Tests laufen ohne externe Dienste.

Pflichttests (Detailaufstellung in `docs/planning/plan-spike.md` §7.1):

- Unit-Test für `RegisterPlaybackEventBatch`
- Unit-Test für zentrale Domain-Validierung
- Integrationstest `POST /api/playback-events` Happy Path
- Integrationstest HTTP 400 bei abweichender `schema_version`
- Integrationstest HTTP 401 bei fehlendem oder falschem Token
- Integrationstest HTTP 401 bei `project_id`/Token-Mismatch
- Integrationstest HTTP 401 bei unbekanntem `project_id`
- Integrationstest HTTP 413 bei Body über 256 KB
- Integrationstest HTTP 422 bei ungültigem Event
- Integrationstest HTTP 422 bei leerem oder fehlendem `events`-Feld
- Integrationstest HTTP 422 bei mehr als 100 Events im Batch
- Integrationstest HTTP 429 bei Rate Limit mit `Retry-After`-Header

Ein Testbefehl muss funktionieren:

```bash
make test
```

`make test` ruft intern `docker build --target test -t
m-trace-api-spike:<stack>-test .` auf. Lokale Toolchain-Befehle
(`go test ./...`, `./gradlew test`) sind im Spike kein Pflichtweg —
der Docker-Build ist die kanonische Test-Ausführung. Wer lokale
Iteration bevorzugt, kann sich Toolchains selbst einrichten; das
Repo committet weder `gradle-wrapper.jar` noch lokale Toolchain-
Files.

---

## 7. Bonus-Scope

Der Bonus-Scope wird nur umgesetzt, wenn nach dem Muss-Scope Zeit bleibt.

Bonusfunktionen:

| Funktion | Zweck |
|---|---|
| `GET /api/stream-sessions` | Liste aktiver Sessions |
| `GET /api/stream-sessions/{id}` | Session-Detail mit Events |
| Origin-Allowlist pro Project | CORS/Auth-Ergonomie prüfen |
| expliziter Session-Lifecycle `ended` | Domainmodell prüfen |
| OTel-Counter zusätzlich zu Prometheus | OTel-Ergonomie besser bewerten |
| `trace_id` in Logs | Trace-/Logging-Integration prüfen |
| vollständige Hexagon-Schichtung bis ins Detail | Architekturfit prüfen |

Wichtig:

Bonusfunktionen dürfen den Muss-Scope nicht gefährden. Wenn ein Prototyp Bonusfunktionen hat, der andere aber nicht, wird das als Bonus notiert, aber nicht als Muss-Versagen des anderen gewertet.

---

## 8. Was explizit NICHT zum Scope gehört

- Persistenz auf Disk
- SQLite
- Redis
- Authentication-Flows jenseits des hardcodierten Tokens
- dynamische Project-Verwaltung
- WebSocket-Endpunkte
- Stream-Analyzer-Anbindung
- Frontend-Integration
- Migrations
- Schemas
- ORM
- Tempo
- Loki
- OTel Collector
- Kubernetes
- Performance-Optimierung jenseits "läuft mit Tests durch"
- Lasttests unter realer Produktionslast

---

## 9. Bewertungsraster

Beide Prototypen werden anhand identischer Kategorien bewertet. Punkte 1–5, höher = besser.

| Kategorie | Gewicht | Was wird gemessen |
|---|---:|---|
| Time to Running Endpoint | 10% | Stunden von `git init` bis zum ersten erfolgreichen `curl POST` |
| OTel-Integration-Ergonomie | 15% | Boilerplate, Doku-Qualität, Stolperstellen |
| Hexagon-Fit | 15% | Wie natürlich fühlt sich Port/Adapter-Trennung an, ohne gegen das Framework zu kämpfen? |
| Test-Velocity | 10% | Test-Setup, Mock-Aufwand, Suite-Laufzeit |
| Docker Image Size | 5% | Gemessen, nicht geschätzt |
| Cold Start | 5% | Gemessen, nicht geschätzt |
| Build-Komplexität | 10% | Schritte für `make build`, CI-Aufwand |
| Subjektiver Spaß | 10% | Wie es sich nach 2 Tagen anfühlt |
| Contributor-Fit | 10% | Wie wahrscheinlich greift ein Streaming-OSS-Mensch das in PRs an? |
| Absehbare Phase-2-Risiken | 10% | Was wird vermutlich problematisch bei SRT, Stream-Analyzer-Integration, Tempo? |

Bewusst nicht direkt in der Bewertung:

- "Welche Sprache ich besser kann" — fließt indirekt über Velocity ein, soll aber nicht doppelt gewichtet werden
- Performance-Benchmarks unter Last — irrelevant für die zu erwartende MVP-Größenordnung
- Theoretische Vorteile aus Blogposts — nur eigene Erfahrung zählt

---

## 10. Entscheidungsregel

Es gibt kein Unentschieden.

Regeln:

1. Wenn ein Stack mindestens 10 gewichtete Prozentpunkte Vorsprung hat, gewinnt er.
2. Wenn der Abstand unter 10 Prozentpunkten liegt, entscheidet `Contributor-Fit`.
3. Wenn `Contributor-Fit` gleich bewertet wird, entscheidet subjektive Wartbarkeit nach 2 Tagen.
4. Wenn auch das nicht reicht, wird die Entscheidung bewusst zugunsten des Stacks getroffen, der weniger Infrastruktur- und Build-Komplexität erzeugt.
5. Eine weitere Spike-Runde ist nicht erlaubt.

Diese Regel ist absichtlich hart. Sie verhindert, dass das ADR zu "beide Optionen haben Vor- und Nachteile" verkommt.

---

## 11. Messpunkte

Diese Werte werden für jeden Prototyp objektiv festgehalten und in das ADR übernommen.

| Metrik | Erfassung |
|---|---|
| Wallclock bis erster grüner Test | Stoppuhr |
| Wallclock bis erstes `docker run` mit 200 OK | Stoppuhr |
| LoC im Domain-Layer | `cloc src/.../domain/` oder äquivalent |
| LoC im Adapter-Layer | `cloc src/.../adapters/` oder äquivalent |
| Artefaktgröße | Go Binary bzw. JVM App/JAR |
| Final Docker Image Size | `docker images` |
| Cold Start bis erster 200 OK auf `/api/health` | `time` + Curl-Loop |
| Build-Zeit von Scratch | `time docker build --no-cache` |
| Größe des Dependency-Caches | isolierter Cache pro Prototyp (siehe Hinweis in `docs/planning/plan-spike.md` §7.2) |
| Anzahl direkter Dependencies | direkt deklariert in `apps/api/go.mod` (`require`-Block ohne `// indirect`) bzw. `apps/api/build.gradle.kts` (`dependencies {}`-Block); transitive werden nicht mitgezählt |
| Testlaufzeit | `time make test` oder äquivalent |
| Anzahl direkt geschriebener Konfigurationsdateien | manuell zählen |

Diese Zahlen sollen subjektive Eindrücke verankern, nicht ersetzen.

---

## 12. Prozess

### 12.1 Tag 0: API-Kontrakt

- Datei `spec/backend-api-contract.md` erstellen
- Endpunkte fixieren
- Beispielpayloads fixieren
- Statuscodes fixieren
- Metriknamen fixieren
- Testfälle fixieren
- Danach keine einseitigen Scope-Änderungen mehr

### 12.2 Tag 1–2: Go-Prototyp

Branch:

```text
spike/go-api
```

Stack:

- Go 1.22+
- Standard Library für HTTP
- `chi` für Routing, falls Standard `net/http` zu spartanisch wirkt
- OTel via `go.opentelemetry.io/otel`
- Tests mit Standard `testing` und `httptest`
- Logging mit `log/slog`

### 12.3 Tag 3–4: Micronaut-Prototyp

Branch:

```text
spike/micronaut-api
```

Stack:

- Micronaut 4.x
- Kotlin 2.1.x auf JDK 21 (Default; Begründung in
  `docs/planning/plan-spike.md` §14.6)
- OTel via Micronaut-OpenTelemetry-Modul oder direkte OTel SDK-Nutzung
- Tests mit Kotest 6.x + MockK + `@MicronautTest`-Integration
  (Default; siehe `docs/planning/plan-spike.md` §14.10)
- Logging mit Logback und optional Logstash-Encoder
- Linting: `detekt` 1.23+ als Soll (siehe `docs/planning/plan-spike.md` §14.9)

### 12.4 Tag 5: Auswertung

1. Bewertungsraster für beide Prototypen ausfüllen
2. Messpunkte-Tabelle ausfüllen
3. Reihenfolge-Bias notieren
4. Subjektive Notizen sortieren
5. Entscheidung gemäß Entscheidungsregel treffen
6. ADR schreiben
7. Sieger-Branch wird Basis für `apps/api`
8. Unterlegener Branch wird archiviert, aber nicht weiterentwickelt

---

## 13. Branch- und Archivierungsregeln

Es gibt keine dauerhafte Coexistenz beider Backend-Stacks.

Regeln:

- Sieger-Branch wird in die Hauptentwicklung übernommen.
- Unterlegener Branch wird nicht weiterentwickelt.
- Unterlegener Branch wird entweder gelöscht oder als Tag archiviert.
- Wenn archiviert, dann mit eindeutigem Namen:

```text
spike/backend-stack-loser-YYYYMMDD
```

- Der finale Commit-Hash beider Prototypen wird im ADR referenziert.
- Das ADR ist die Quelle der Entscheidung, nicht der alte Branch.

Empfehlung:

Nicht stumpf löschen. Besser den unterlegenen Prototyp als Tag oder Commit-Hash erhalten, aber keine parallele Codebasis im aktiven Repository pflegen.

---

## 14. Anti-Pattern während des Spikes

- *"Ich habe Micronaut lange nicht benutzt, der Vergleich wird unfair."*  
  Genau das ist eine Erkenntnis. Notieren, nicht ausgleichen.

- *"Ich erweitere den Scope, weil X im anderen Stack so einfach war."*  
  Verfälscht den Vergleich. Strikt verboten.

- *"Ich brauche noch einen halben Tag."*  
  Tabu. Was in zwei Tagen nicht steht, sagt mehr als ein perfekter Prototyp am dritten Tag.

- *"Ich kann beide Prototypen behalten und später entscheiden."*  
  Keine Coexistenz. Ein Stack wird gewählt.

- *"Ich brauche keinen Test, der Endpunkt funktioniert ja im Browser."*  
  Test-Velocity ist eine Bewertungskategorie. Ohne Tests kein gültiger Vergleich.

- *"OTel richte ich später sauber ein."*  
  Nein. OTel-Ergonomie ist einer der wichtigsten Gründe für den Spike. Minimal muss sie im Code berührt werden.

- *"Prometheus reicht, OTel ist egal."*  
  Nein. m-trace positioniert sich OpenTelemetry-native. Der Spike muss zeigen, welcher Stack dabei weniger Reibung erzeugt.

---

## 15. ADR-Vorlage

Nach Abschluss des Spikes wird das Ergebnis als ADR commitet:

```markdown
# 0001 — Backend-Technologie für apps/api

Status:   Accepted
Datum:    YYYY-MM-DD
Beteiligt: <Name>

## Kontext

Lastenheft §9.1 hielt die Wahl zwischen Go und Micronaut bewusst offen.
Spike gemäß `docs/spike/0001-backend-stack.md` durchgeführt.

## Optionen

- Go mit Standard Library und optional Chi
- Kotlin 2.1.x auf JDK 21 mit Micronaut 4.x

## Scope

[Muss-Scope erfüllt? Bonus-Scope teilweise erfüllt? Abweichungen dokumentieren.]

## Reihenfolge und Bias

[Welcher Stack wurde zuerst gebaut? Welche Bias-Effekte wurden beobachtet?]

## Bewertung

[Tabelle aus §9 dieses Spike-Plans, mit Punkten je Stack und Kommentar pro Zeile.]

## Messwerte

[Tabelle aus §11 dieses Spike-Plans.]

## Entscheidung

Gewählt: <Stack>

Begründung in 2–3 Sätzen.

## Konsequenzen

- Welche Abschnitte des Lastenhefts werden konkret?
- Welche Anforderungen werden technisch anders umgesetzt?
- Welche Risiken entstehen, welche verschwinden?
- Welche Folge-ADRs werden absehbar nötig, z. B. Persistenz, Tracing-Backend, SDK-Transport?
- Welche Spike-Ergebnisse werden archiviert?
```

---

## 16. Bewertungsbogen für ADR

| Kategorie | Gewicht | Go Punkte | Go gewichtet | Micronaut Punkte | Micronaut gewichtet | Kommentar |
|---|---:|---:|---:|---:|---:|---|
| Time to Running Endpoint | 10% |  |  |  |  |  |
| OTel-Integration-Ergonomie | 15% |  |  |  |  |  |
| Hexagon-Fit | 15% |  |  |  |  |  |
| Test-Velocity | 10% |  |  |  |  |  |
| Docker Image Size | 5% |  |  |  |  |  |
| Cold Start | 5% |  |  |  |  |  |
| Build-Komplexität | 10% |  |  |  |  |  |
| Subjektiver Spaß | 10% |  |  |  |  |  |
| Contributor-Fit | 10% |  |  |  |  |  |
| Absehbare Phase-2-Risiken | 10% |  |  |  |  |  |
| **Gesamt** | **100%** |  |  |  |  |  |

---

## 17. Messwertbogen für ADR

| Metrik | Go | Micronaut | Kommentar |
|---|---:|---:|---|
| Wallclock bis erster grüner Test |  |  |  |
| Wallclock bis erstes `docker run` mit 200 OK |  |  |  |
| LoC im Domain-Layer |  |  |  |
| LoC im Adapter-Layer |  |  |  |
| Artefaktgröße |  |  |  |
| Final Docker Image Size |  |  |  |
| Cold Start bis erster 200 OK auf `/api/health` |  |  |  |
| Build-Zeit von Scratch |  |  |  |
| Größe des Dependency-Caches |  |  |  |
| Anzahl direkter Dependencies |  |  |  |
| Testlaufzeit |  |  |  |
| Anzahl direkt geschriebener Konfigurationsdateien |  |  |  |

---

## 18. Was nach dem Spike passiert

1. Sieger-Branch wird zum `apps/api`-Skelett ausgebaut.
2. Unterlegener Branch wird archiviert oder gelöscht, aber nicht weiterentwickelt.
3. ADR wird commitet als:

```text
docs/adr/0001-backend-stack.md
```

4. Lastenheft wird auf `1.0.0` gehoben:
   - Header und §10.1 mit getroffener Entscheidung füllen
   - §9.1 in den Vergangenheitsmodus setzen
   - offene Backend-Entscheidung aus den offenen Punkten entfernen

5. README dokumentiert den gewählten Stack im `Tech Overview`.

---

## 19. Erfolgskriterium des Spikes selbst

Der Spike ist erfolgreich, wenn nach 5 Arbeitstagen (Summe aus §2: 0,5 + 2 + 2 + 0,5):

- zwei lauffähige Prototypen existieren oder einer dokumentiert am Muss-Scope gescheitert ist,
- Bewertungsraster und Messpunkte für beide ausgefüllt sind,
- Reihenfolge-Bias dokumentiert ist,
- ein ADR mit klarer Entscheidung existiert,
- du **nicht** das Gefühl hast, "noch eine Runde recherchieren" zu müssen.

Letzter Punkt ist der wichtigste. Wenn der Spike das Gefühl der Unsicherheit nicht beseitigt, war die Bewertung zu zögerlich oder das Raster passt nicht zu deinem Kontext. Das wird dann durch explizite Neubewertung gelöst, nicht durch mehr Code.
