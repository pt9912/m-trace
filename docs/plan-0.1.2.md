# Implementation Plan — `0.1.2` (Observability-Stack)

> **Status**: ⬜ offen. Beginnt nach Abschluss von `0.1.1` (Player-SDK + Dashboard).  
> **Bezug**: [Lastenheft `1.1.3`](./lastenheft.md) §13.3 (RAK-9, RAK-10), §18 (MVP-DoD-Anteil); [Roadmap](./roadmap.md) §3; [Architektur (Zielbild)](./architecture.md); [API-Kontrakt](./spike/backend-api-contract.md); [Risiken-Backlog](./risks-backlog.md).  
> **Vorgänger-Gate (Stand zum `0.1.2`-Start, nicht zum heutigen Zeitpunkt)**:
>
> - [`plan-0.1.1.md`](./plan-0.1.1.md) (Player-SDK + Dashboard) muss vollständig (`[x]`) sein, inklusive Release-Akzeptanzkriterien `0.1.1` (§5).
> - [`plan-0.1.0.md`](./plan-0.1.0.md) Tranche 0b §4.3 (Telemetry-Driven-Port + OTel-Counter + Request-Span + autoexport) muss vollständig (`[x]`) sein. **Harte technische Voraussetzung** — Tranche 1 dieses Plans (Pflicht-Anteile in `apps/api`) baut darauf auf; F-91 ist hier nicht erneut implementiert, sondern „Voraussetzung erfüllt durch §4.3".
> - **Tranche 0c (Lastenheft-Patches)** in `plan-0.1.0.md`: konstruktionsbedingt fortlaufend offen; das Gate verlangt nur, dass alle bis zum `0.1.2`-Start eingetragenen §4a.x-Items entweder `[x]` oder explizit als nicht-blockierend markiert sind — nicht den Abschluss der Tranche selbst.
>
> Konsequenz: solange `0.1.1` nicht released oder `plan-0.1.0` §4.3 nicht abgeschlossen ist, hat dieses Plan-Dokument Status ⬜ in Tranchen-Übersicht und Roadmap §3.

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
| 0 | Vorgänger-Gate-Verifikation | ⬜ |
| 1 | Pflicht-Anteile (F-89..F-93, Mindestmetriken) — Code in `apps/api` plus Repo-weite Prometheus-Konfiguration unter `observability/prometheus/` | ⬜ |
| 2 | Soll-Anteile im `observability`-Compose-Profil (F-94/MVP-28 Grafana, MVP-29 OTel-Collector) | ⬜ |
| 3 | Release-Akzeptanzkriterien `0.1.2` | ⬜ |

Tempo bleibt explizit Nicht-MVP (MVP-22).

---

## 1a. Tranche 0 — Vorgänger-Gate-Verifikation

Konvertiert die narrative Vorgänger-Gate-Beschreibung aus §0 in prüfbare DoD-Items. Tranche muss `[x]` sein, bevor Tranche 1 beginnt.

DoD:

- [ ] `plan-0.1.1.md` Tranche 1 (Player-SDK) abgeschlossen — alle DoD-Items `[x]`.
- [ ] `plan-0.1.1.md` Tranche 2 (Dashboard) abgeschlossen.
- [ ] `plan-0.1.1.md` Tranche 3 (Compose-Lab-Erweiterung um `dashboard`) abgeschlossen.
- [ ] `plan-0.1.1.md` Tranche 4 (Release-Akzeptanzkriterien `0.1.1` — RAK-2, RAK-5, RAK-7) abgeschlossen.
- [ ] `plan-0.1.0.md` Tranche 0b §4.3 (Telemetry-Driven-Port + OTel-Counter + Request-Span + autoexport) abgeschlossen — harte technische Voraussetzung für F-91.
- [ ] `plan-0.1.0.md` Tranche 0c §4a.x: alle bis zum `0.1.2`-Start eingetragenen Items entweder `[x]` oder explizit als nicht-blockierend markiert.

---

## 2. Tranche 1 — Observability Pflicht-Anteile (`apps/api` und `observability/prometheus/`)

Bezug: MVP-10 (Muss), MVP-15 (Muss); F-89..F-93 (Muss); Mindestmetriken laut Lastenheft §7.9.

DoD:

- [ ] **F-89** Strukturierte Logs in `apps/api` (`log/slog` + JSON-Handler ist bereits aus dem Spike vorhanden; im Compose-Stack stdout-fähig konfiguriert; Verifikation per `docker compose logs api`).
- [ ] **F-90** Health Check `/api/health` ist bereits aus dem Spike vorhanden — Verifikation, dass der Endpoint im Compose-Stack `200` liefert (Bezug RAK-3 aus `0.1.0`).
- [ ] **F-91** OpenTelemetry-Unterstützung — **Voraussetzung aus Tranche-0b §4.3** in `plan-0.1.0.md` (`Telemetry`-Port + OTLP-Anbindung via `autoexport`). Vor Beginn von Tranche 1 muss §4.3 vollständig (`[x]`) sein; `0.1.2` baut darauf auf, implementiert F-91 aber nicht erneut.
- [ ] **F-92** Playback-Events sind als Metriken oder Traces exportierbar — über den `Telemetry`-Port-Counter (Metriken) sowie HTTP-Adapter-Spans (Traces). Aktivierung erfolgt über `OTEL_*`-Env-Vars; im Core-Stack ohne observability-Profil bleiben sie silent.
- [ ] **F-93** Prometheus-Konfiguration unter `observability/prometheus/` mit Scrape-Job für den `api`-Compose-Service (`targets: ["api:8080"]`, `metrics_path: "/api/metrics"`). Prometheus selbst läuft im observability-Profil (Tranche 2).
- [ ] Mindestmetriken aus Lastenheft §7.9 in `apps/api` instrumentiert: bereits vorhanden sind die vier API-Kontrakt-Counter (`mtrace_playback_events_total`, `mtrace_invalid_events_total`, `mtrace_rate_limited_events_total`, `mtrace_dropped_events_total`); ergänzend für `0.1.2`: `mtrace_active_sessions`, `mtrace_api_requests_total`, `mtrace_playback_errors_total`, `mtrace_rebuffer_events_total`, `mtrace_startup_time_ms`. Cardinality-Regeln aus Lastenheft §7.10 sind einzuhalten.
- [ ] **RAK-9-Seed-Skript** `scripts/seed-rak9.sh` (oder gleichwertiges `make seed-rak9`-Target) erzeugt reproduzierbar mindestens 50 Events in 5 Sessions an `/api/playback-events`. Nutzt `curl` gegen einen laufenden Compose-Stack mit `make dev-observability`; verschiedene `session_id`/`event_name`-Muster für Cardinality-Spot-Check. Voraussetzung für RAK-9-Smoke-Test (§4) und CI-Verifikation (sobald OE-6 entschieden); ohne dieses Skript bleibt RAK-9 von manueller Lastaufbereitung abhängig.

---

## 3. Tranche 2 — Observability Soll-Anteile (`observability`-Compose-Profil)

Bezug: MVP-28 (Soll Grafana), MVP-29 (Soll OTel-Collector); F-94 (Soll, harmonisiert mit MVP-28 in Lastenheft `1.0.1`, siehe `plan-0.1.0.md` Tranche 0c §4a.1); F-87/F-88 (siehe Patch `1.0.2`, `plan-0.1.0.md` §4a.2).

Soll-Komponenten leben im `observability`-Compose-Profil und werden über `make dev-observability` (oder `docker compose --profile observability up`) ergänzend zum Core-Stack gestartet.

DoD:

- [ ] Compose-Erweiterung: `prometheus`, `grafana`, `otel-collector` mit `profiles: ["observability"]` — additiv und opt-in.
- [ ] `make dev-observability` (Makefile-Target) aktiviert das observability-Profil zusätzlich zum Core.
- [ ] **MVP-29** OTel-Collector unter `observability/otel-collector/`; nimmt OTLP von `apps/api` entgegen und exportiert Metriken zu Prometheus. Trace-Backend (z. B. Jaeger) ist Bonus, **kein** Pflicht-Bestandteil — Tempo ist per MVP-22 Nicht-MVP.
- [ ] **F-94 + MVP-28** Grafana-Container im observability-Profil mit einem einfachen Beispiel-Dashboard unter `observability/grafana/`. Dashboard zeigt mindestens die vier API-Kontrakt-Counter; weitere Mindestmetriken aus §7.9 als Bonus.
- [ ] System-Status-Ansicht im Dashboard (`apps/dashboard`, `0.1.1` §3) erkennt das aktive observability-Profil und zeigt Prometheus/Grafana/OTel-Collector als „connected" statt „inaktiv" an.

---

## 4. Tranche 3 — Release-Akzeptanzkriterien `0.1.2` (Lastenheft §13.3)

DoD:

- [ ] **RAK-9** Prometheus enthält nur aggregierte Metriken — Smoke-Test über `make dev-observability`:
    - **Setup-Pflicht**: vor dem Smoke-Test muss eine Mindestdaten-Lage im Prometheus erzeugt sein, sonst geben die Queries leer zurück und der Cardinality-Check besteht trivial (false positive). Konkret: Compose-Stack läuft, mindestens 5 Player-Sessions mit jeweils ≥ 10 Events; mindestens ein Prometheus-Scrape-Intervall (Default 15 s) ist vergangen.
    - **Seed-Skript**: `scripts/seed-rak9.sh` (oder `make seed-rak9`-Target) erzeugt die Mindestdaten-Lage reproduzierbar via `curl`-Aufrufe gegen `/api/playback-events` (50 Events in 5 Sessions mit unterschiedlichen `session_id`/`event_name`-Mustern). Der Smoke-Test ruft das Skript als ersten Schritt auf, danach laufen die unten genannten Queries gegen Prometheus. CI-Pipeline (sobald OE-6 entschieden) ruft denselben Pfad auf — das vermeidet manuelle Lastaufbereitung und „false confidence". Demo-SDK-basierte Lastaufbereitung ist eine Bonus-Variante für interaktive Lab-Sessions, kein DoD-Pfad.
    - `curl -g 'http://localhost:9090/api/v1/series?match[]={__name__=~"mtrace_.+"}'` listet alle `mtrace_*`-Series. Erwartet: Liste ist **nicht-leer** (Setup-Voraussetzung greift) und keine Series enthält die verbotenen Labels `session_id`, `user_agent`, `segment_url`, `client_ip`. Das `-g`-Flag deaktiviert curl-URL-Globbing, das eckige Klammern sonst als Range-Pattern interpretiert.
    - Pro Mindestmetrik aus Lastenheft §7.9: `curl -g 'http://localhost:9090/api/v1/labels?match[]={__name__="mtrace_playback_events_total"}'` (analog für die übrigen) listet die tatsächlich verwendeten Labels — Spot-Check gegen die Cardinality-Regeln aus Lastenheft §7.10. `match[]` benötigt einen vollständigen Vektor-Selector (`{__name__="..."}`); der nackte Metrikname ist zwischen Prometheus-Versionen nicht zuverlässig akzeptiert.
    - Zusätzlich: `mtrace_playback_events_total` muss einen Wert > 0 haben (`curl 'http://localhost:9090/api/v1/query?query=mtrace_playback_events_total'`); diente als Sanity-Check, dass der Counter überhaupt aktiv inkrementiert wurde.
    - Der frühere `api/v1/label/session_id/values`-Endpoint ist zu schwach (globaler Discovery-Endpoint, hängt von der Datenmenge ab) und wird nicht mehr verwendet.
- [ ] **RAK-10 (Soll)** Player-Session-Traces sind vorbereitet oder exemplarisch sichtbar — entweder als OTel-Span-Struktur in `apps/api` (mindestens ein Span pro Batch, abgedeckt durch Tranche-0b §4.3 in `plan-0.1.0.md`) oder über die eingebaute Session-/Trace-Ansicht im Dashboard (MVP-14, `plan-0.1.1.md` §3). Tempo bleibt **explizit Nicht-MVP** (MVP-22).

### 4.1 Übergreifende DoD `0.1.2` (Lastenheft §18, `0.1.2`-Anteil)

- [ ] `CHANGELOG.md` enthält Eintrag für `0.1.2`.
- [ ] README/`docs/local-development.md` ergänzt um die `make dev-observability`-Variante und die Service-URLs (Prometheus, Grafana, OTel-Collector) — RAK-8-Refinement.

---

## 5. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`, Commit-Hash anhängen.
- Neue Findings in `0.1.2`-Phase landen entweder in dieser Datei oder in `risks-backlog.md`.
- Lastenheft-Patches während `0.1.2` werden in `plan-0.1.0.md` Tranche 0c als neue §4a.x-Einträge ergänzt (zentrale Patch-Historie).
- Mit `0.1.2`-Release ist die `0.1.x`-Phase abgeschlossen; nächster Plan ist `plan-0.2.0.md` (publizierbares Player-SDK).
