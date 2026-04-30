# Implementation Plan — `0.1.2` (Observability-Stack)

> **Status**: ✅ implementiert. Observability-Profil, RAK-9 und RAK-10 sind umgesetzt und lokal verifiziert.  
> **Bezug**: [Lastenheft `1.1.6`](./lastenheft.md) §13.3 (RAK-9, RAK-10), §18 (MVP-DoD-Anteil); [Roadmap](./roadmap.md) §3; [Architektur (Zielbild)](./architecture.md); [API-Kontrakt](./spike/backend-api-contract.md); [Risiken-Backlog](./risks-backlog.md).
> **Vorgänger-Gate (Stand zum `0.1.2`-Start, nicht zum heutigen Zeitpunkt)**:
>
> - [`plan-0.1.1.md`](./plan-0.1.1.md) (Player-SDK + Dashboard) muss vollständig (`[x]`) sein, inklusive Release-Akzeptanzkriterien `0.1.1` (§5).
> - [`plan-0.1.0.md`](./plan-0.1.0.md) Tranche 0b §4.3 (Telemetry-Driven-Port + OTel-Counter + Request-Span + autoexport) muss vollständig (`[x]`) sein. **Harte technische Voraussetzung** — Tranche 1 dieses Plans (Pflicht-Anteile in `apps/api`) baut darauf auf; F-91 ist hier nicht erneut implementiert, sondern „Voraussetzung erfüllt durch §4.3".
> - **Tranche 0c (Lastenheft-Patches)** in `plan-0.1.0.md`: konstruktionsbedingt fortlaufend offen; das Gate verlangt nur, dass alle bis zum `0.1.2`-Start eingetragenen §4a.x-Items entweder `[x]` oder explizit als nicht-blockierend markiert sind — nicht den Abschluss der Tranche selbst.
>
> Startkonsequenz: das Gate wurde in §1a verifiziert; offene Punkte liegen ab jetzt in Tranche 3.

## 0. Konvention

DoD-Checkboxen analog [`plan-0.1.0.md`](./plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz.
- 🟡 in Arbeit.

Tranchen 0/0a/0b/0c werden in `plan-0.1.0.md` gepflegt — neue Lastenheft-Patches in der `0.1.2`-Phase landen ebenfalls dort als §4a.x-Eintrag.

---

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
|---|---|---|
| 0 | Vorgänger-Gate-Verifikation | ✅ |
| 1 | Pflicht-Anteile (F-89..F-93, Mindestmetriken) — Code in `apps/api` plus Repo-weite Prometheus-Konfiguration unter `observability/prometheus/` | ✅ |
| 2 | Soll-Anteile im `observability`-Compose-Profil (F-94/MVP-28 Grafana, MVP-29 OTel-Collector) | ✅ |
| 3 | Release-Akzeptanzkriterien `0.1.2` | ✅ |

Tempo bleibt explizit Nicht-MVP (MVP-22).

---

## 1a. Tranche 0 — Vorgänger-Gate-Verifikation

Konvertiert die narrative Vorgänger-Gate-Beschreibung aus §0 in prüfbare DoD-Items. Gate ist in zwei Kategorien geteilt: **harte Voraussetzungen** (alle `[x]`) und **weiche Voraussetzungen** (offen erlaubt, wenn explizit als nicht-blockierend markiert). Tranche ist „erfüllt", wenn alle harten und alle blockierenden weichen Items `[x]` sind.

DoD — **harte Voraussetzungen, technisch zwingend** (Pflicht `[x]` vor `0.1.2`-Start):

- [x] `plan-0.1.1.md` Tranche 1 (Player-SDK) abgeschlossen — alle DoD-Items `[x]` (`55ccac4`, `966699f`).
- [x] `plan-0.1.1.md` Tranche 2 (Dashboard) abgeschlossen (`55ccac4`, `966699f`).
- [x] `plan-0.1.1.md` Tranche 3 (Compose-Lab-Erweiterung um `dashboard`) abgeschlossen (`1d58f19`, `966699f`).
- [x] `plan-0.1.1.md` Tranche 4 (Release-Akzeptanzkriterien `0.1.1` — RAK-2, RAK-5, RAK-7) abgeschlossen (`9b26e4f`, `55ccac4`, `966699f`).
- [x] `plan-0.1.0.md` Tranche 0b §4.3 (Telemetry-Driven-Port + OTel-Counter + Request-Span + autoexport) abgeschlossen — harte technische Voraussetzung für F-91 (`51b3812`).
- [x] `plan-0.1.0.md` §3.5 telemetry-model.md, **Pflicht-Anteile für `0.1.2`** — OTel-Modell (F-91, F-92) und Cardinality-Regeln (F-95..F-105): das Observability-Stack-Setup braucht diese Spezifikationen für Konfiguration und Verifikation (`e532e1e`, `51b3812`).

DoD — **weiche Voraussetzungen, Dokumentations-/Aufräumarbeiten** (offen erlaubt; nicht Gate-blockierend):

- [x] `plan-0.1.0.md` §3.5 telemetry-model.md, **nicht-Pflicht-Anteile für `0.1.2`** — Wire-Format §1, Backpressure §4, Time-Stempel §5, Schema-Versionierung §6 (sind primär für `0.1.1` SDK relevant; bis `0.1.2`-Start erwartet `[x]`, aber nicht harter Gate-Blocker) (`e532e1e`, `51b3812`).
- [x] `plan-0.1.0.md` §3.6 local-development.md: nicht direkt für `0.1.2`-Implementierung erforderlich (`2eede43`, `504e4c9`, `beabcd5`).
- [x] `plan-0.1.0.md` Tranche 0c §4a.x-Items werden vor dem `0.1.2`-Start einzeln eingestuft: jedes offene Item ist entweder
    - **blockierend** → muss `[x]` sein, **oder**
    - **nicht-blockierend** → offen erlaubt, mit `(nicht-blockierend für 0.1.2)`-Vermerk. Bis Patch `1.1.6` sind keine blockierenden §4a.x-Items offen (`966699f`, `beabcd5`).
- [x] Vorgänger-Gate-Verifikations-Commit dokumentiert die Einstufung pro offenem Item nachvollziehbar (`beabcd5`).

---

## 2. Tranche 1 — Observability Pflicht-Anteile (`apps/api` und `observability/prometheus/`)

Bezug: MVP-10 (Muss), MVP-15 (Muss); F-89..F-93 (Muss); Mindestmetriken laut Lastenheft §7.9.

DoD:

- [x] **F-89** Strukturierte Logs in `apps/api` (`log/slog` + JSON-Handler ist bereits aus dem Spike vorhanden; im Compose-Stack stdout-fähig konfiguriert; Verifikation per `docker compose logs api`) (`504e4c9`, verifiziert in `beabcd5`).
- [x] **F-90** Health Check `/api/health` ist bereits aus dem Spike vorhanden — Verifikation, dass der Endpoint im Compose-Stack `200` liefert (Bezug RAK-3 aus `0.1.0`) (`504e4c9`, verifiziert in `beabcd5`).
- [x] **F-91** OpenTelemetry-Unterstützung — **Voraussetzung aus Tranche-0b §4.3** in `plan-0.1.0.md` (`Telemetry`-Port + OTLP-Anbindung via `autoexport`). Vor Beginn von Tranche 1 muss §4.3 vollständig (`[x]`) sein; `0.1.2` baut darauf auf, implementiert F-91 aber nicht erneut (`51b3812`, `beabcd5`).
- [x] **F-92** Playback-Events sind als Metriken oder Traces exportierbar — über den `Telemetry`-Port-Counter (Metriken) sowie HTTP-Adapter-Spans (Traces). Aktivierung erfolgt über `OTEL_*`-Env-Vars; im Core-Stack ohne observability-Profil bleiben sie silent (`51b3812`, `beabcd5`).
- [x] **F-93** Prometheus-Konfiguration unter `observability/prometheus/` mit Scrape-Job für den `api`-Compose-Service (`targets: ["api:8080"]`, `metrics_path: "/api/metrics"`). Prometheus selbst läuft im observability-Profil (Tranche 2) (`beabcd5`).
- [x] Mindestmetriken aus Lastenheft §7.9 in `apps/api` instrumentiert: bereits vorhanden sind die vier API-Kontrakt-Counter (`mtrace_playback_events_total`, `mtrace_invalid_events_total`, `mtrace_rate_limited_events_total`, `mtrace_dropped_events_total`); ergänzend für `0.1.2`: `mtrace_active_sessions`, `mtrace_api_requests_total`, `mtrace_playback_errors_total`, `mtrace_rebuffer_events_total`, `mtrace_startup_time_ms`. Cardinality-Regeln aus Lastenheft §7.10 sind einzuhalten (`beabcd5`).
- [x] **RAK-9-Seed-Skript** `scripts/seed-rak9.sh` (oder gleichwertiges `make seed-rak9`-Target) erzeugt reproduzierbar mindestens 50 Events in 5 Sessions an `/api/playback-events`. Nutzt `curl` gegen einen laufenden Compose-Stack mit `make dev-observability`; verschiedene `session_id`/`event_name`-Muster für Cardinality-Spot-Check. Voraussetzung für RAK-9-Smoke-Test (§4) und spätere CI-Verifikation; ohne dieses Skript bleibt RAK-9 von manueller Lastaufbereitung abhängig (`beabcd5`).
- [x] **Seed-Skript-Contract**: das Skript akzeptiert Pflicht- und Optional-Parameter, damit es deterministisch in frischen Checkouts und in CI läuft (`beabcd5`):
    - `--base-url` (Default `http://localhost:8080`).
    - `--project-id` (Default `demo`); muss in der API-`StaticProjectResolver`-Konfiguration als gültige Project-ID hinterlegt sein.
    - `--token` (Default `demo-token`); muss zu `--project-id` passen.
    - `--origin` (Default leer = kein `Origin`-Header gesendet, CLI/Lab-Pfad); bei gesetztem Wert wird der Origin in den HTTP-Headern mitgesendet — der Wert muss in den Allowed-Origins des Projekts sein, sonst antwortet die API mit 403.
    - `--sessions` (Default 5), `--events-per-session` (Default 10).
    - `--skip-auth` (Bool, Default false): überspringt Auth-Header für lokale isolierte Smoke-Variante (z. B. wenn Backend `apps/api` ohne Auth-Konfiguration läuft) — in CI niemals nutzen.
    - Compose-Default-Fixtures: `services/`-Konfigurationen liefern eine `demo`-Project-ID mit Token `demo-token` und Allowed-Origin `http://localhost:5173`, damit das Seed-Skript out-of-the-box läuft.

---

## 3. Tranche 2 — Observability Soll-Anteile (`observability`-Compose-Profil)

Bezug: MVP-28 (Soll Grafana), MVP-29 (Soll OTel-Collector); F-94 (Soll, harmonisiert mit MVP-28 in Lastenheft `1.0.1`, siehe `plan-0.1.0.md` Tranche 0c §4a.1); F-87/F-88 (siehe Patch `1.0.2`, `plan-0.1.0.md` §4a.2).

Soll-Komponenten leben im `observability`-Compose-Profil und werden über `make dev-observability` (oder `docker compose --profile observability up`) ergänzend zum Core-Stack gestartet.

DoD:

- [x] Compose-Erweiterung: `prometheus`, `grafana`, `otel-collector` mit `profiles: ["observability"]` — additiv und opt-in (`beabcd5`).
- [x] `make dev-observability` (Makefile-Target) aktiviert das observability-Profil zusätzlich zum Core (`beabcd5`).
- [x] **MVP-29** OTel-Collector unter `observability/otel-collector/`; nimmt OTLP von `apps/api` entgegen und exportiert Metriken zu Prometheus. Trace-Backend (z. B. Jaeger) ist Bonus, **kein** Pflicht-Bestandteil — Tempo ist per MVP-22 Nicht-MVP (`beabcd5`).
- [x] **F-94 + MVP-28** Grafana-Container im observability-Profil mit einem einfachen Beispiel-Dashboard unter `observability/grafana/`. Dashboard zeigt mindestens die vier API-Kontrakt-Counter; weitere Mindestmetriken aus §7.9 als Bonus (`beabcd5`).
- [x] System-Status-Ansicht im Dashboard (`apps/dashboard`, `0.1.1` §3) erkennt das aktive observability-Profil und zeigt Prometheus/Grafana/OTel-Collector als „connected" statt „inaktiv" an (`beabcd5`).

---

## 4. Tranche 3 — Release-Akzeptanzkriterien `0.1.2` (Lastenheft §13.3)

DoD:

- [x] **RAK-9** Prometheus enthält nur aggregierte Metriken — Smoke-Test über `make dev-observability` (`make smoke-observability`, `beabcd5`):
    - **Setup-Pflicht**: vor dem Smoke-Test muss eine Mindestdaten-Lage im Prometheus erzeugt sein, sonst geben die Queries leer zurück und der Cardinality-Check besteht trivial (false positive). Konkret: Compose-Stack läuft, mindestens 5 Player-Sessions mit jeweils ≥ 10 Events; mindestens ein Prometheus-Scrape-Intervall (Default 15 s) ist vergangen.
    - **Seed-Skript**: `scripts/seed-rak9.sh` (oder `make seed-rak9`-Target) erzeugt die Mindestdaten-Lage reproduzierbar via `curl`-Aufrufe gegen `/api/playback-events` (50 Events in 5 Sessions mit unterschiedlichen `session_id`/`event_name`-Mustern). Der Smoke-Test ruft das Skript als ersten Schritt auf, danach laufen die unten genannten Queries gegen Prometheus. Eine spätere CI-Erweiterung ruft denselben Pfad auf — das vermeidet manuelle Lastaufbereitung und „false confidence". Demo-SDK-basierte Lastaufbereitung ist eine Bonus-Variante für interaktive Lab-Sessions, kein DoD-Pfad.
    - **Label-Name-Check (verbotene Labels)**: `make smoke-observability` nutzt Prometheus `api/v1/series` mit URL-encodiertem `match[]={__name__=~"mtrace_.+"}` und listet alle `mtrace_*`-Series. Erwartet: Liste ist **nicht-leer** (Setup-Voraussetzung greift) und keine Series enthält die verbotenen Labels `session_id`, `user_agent`, `segment_url`, `client_ip`.
    - **Cardinality-Wert-Check (PromQL)**: zusätzlich zur Label-Name-Liste werden konkrete Kardinalitäts-Limits per PromQL geprüft, weil `api/v1/labels` nur Label-Namen, nicht ihre Wert-Verteilung liefert:
        - Pro Mindestmetrik: `curl 'http://localhost:9090/api/v1/query?query=count(count by (__name__) (mtrace_playback_events_total))'` (analog für die anderen Pflicht-Counter) — Erwartung: Ergebnis ≤ 1 für die Pflicht-Counter mit nur Aggregat-Labels.
        - Generisch für alle `mtrace_*`-Metriken: `curl 'http://localhost:9090/api/v1/query?query=count(count by (instance, job, __name__) ({__name__=~"mtrace_.+"}))'` — Erwartung: Anzahl Series ≤ kleine Konstante (z. B. < 50, abhängig von Mindestmetriken-Anzahl × erlaubte Aggregat-Labels). Eine plötzliche Explosion auf > 100 deutet auf eine Cardinality-Verletzung hin (z. B. wurde session_id versehentlich als Label aufgenommen).
        - Sanity-Check: `mtrace_playback_events_total` muss einen Wert > 0 haben (`curl 'http://localhost:9090/api/v1/query?query=mtrace_playback_events_total'`); bestätigt, dass der Counter aktiv inkrementiert wurde.
    - Der frühere `api/v1/label/session_id/values`-Endpoint ist zu schwach (globaler Discovery-Endpoint, hängt von der Datenmenge ab) und wird nicht mehr verwendet.
- [x] **RAK-10 (Soll)** Player-Session-Traces sind vorbereitet oder exemplarisch sichtbar. Variante A ist über `make smoke-rak10-console` reproduzierbar geprüft (`79be435`):
    - **Variante A — OTel-Spans im Backend (Console-Exporter)**: `apps/api` erzeugt mindestens einen Request-Span pro `POST /api/playback-events` (abgedeckt durch Tranche-0b §4.3 in `plan-0.1.0.md`). Verifikation läuft mit deterministischem Console-Exporter, damit der Test reproduzierbar ist und nicht von Trace-Backend-Bonus-Komponenten (Jaeger u. a. — sind im MVP nicht Pflicht) abhängt: das Smoke-Skript setzt `OTEL_TRACES_EXPORTER=console` für den `api`-Service (entweder via Compose-Override oder `make seed-rak9` mit Env-Var-Injection), führt `seed-rak9.sh` aus, dann prüft `docker compose logs api | grep '"name":"http.handler POST'` mindestens einen Span-Eintrag. Console-Exporter ist immer verfügbar (Teil der OTel-SDK-Distribution); kein Trace-Backend nötig.
    - **Variante B — Dashboard-Trace-Ansicht (constraint)**: die eingebaute Session-/Trace-Ansicht aus `plan-0.1.1.md` §3 (MVP-14) zählt **nur dann** als RAK-10-Erfüllung, wenn sie tatsächlich Trace-äquivalente Daten zeigt — d. h. mindestens (a) Span-äquivalente Hierarchie pro Session (z. B. Request-Span umschließt Event-Spans) **und** (b) Span-Attribute wie Dauer/Latenz pro Sub-Vorgang aus den persistierten Daten ableitbar. Eine reine Event-Liste ohne diese Trace-Semantik gilt als „best effort", nicht als RAK-10-DoD — in dem Fall muss Variante A erfüllt sein. Der Smoke-Test-Commit dokumentiert konkret, welche Variante (A oder B) genutzt wurde und wie sie geprüft wurde.
    - Tempo bleibt **explizit Nicht-MVP** (MVP-22).

### 4.1 Übergreifende DoD `0.1.2` (Lastenheft §18, `0.1.2`-Anteil)

- [x] `CHANGELOG.md` enthält Eintrag für `0.1.2` (`79be435`).
- [x] README/`docs/local-development.md` ergänzt um die `make dev-observability`-Variante und die Service-URLs (Prometheus, Grafana, OTel-Collector) — RAK-8-Refinement (`beabcd5`).

---

## 5. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`, Commit-Hash anhängen.
- Neue Findings in `0.1.2`-Phase landen entweder in dieser Datei oder in `risks-backlog.md`.
- Lastenheft-Patches während `0.1.2` werden in `plan-0.1.0.md` Tranche 0c als neue §4a.x-Einträge ergänzt (zentrale Patch-Historie).
- Mit `0.1.2`-Release ist die `0.1.x`-Phase abgeschlossen; nächster Plan ist `plan-0.2.0.md` (publizierbares Player-SDK).
