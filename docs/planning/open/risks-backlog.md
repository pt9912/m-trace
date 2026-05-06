# Risiken-Backlog

> **Stand**: 2026-05-06 (post-`0.8.0`-Release; Tag `v0.8.0` auf dem Release-Gate-Fix nach `8df263a` + `plan-0.9.0`-Skelett unter `docs/planning/open/`). R-12 ist seit `0.8.0` Tranche 3 release-blockierend; der konkrete Drift-Smoke-Folgeplan steht als Tranche 1 in `plan-0.9.0.md`. R-5/-7/-9/-10/-11 sind aus `0.7.0`/`0.8.0` inhaltlich unverändert; Triggerschwellen unausgelöst.  
> **Bezug**: `docs/adr/0001-backend-stack.md` §5 (Bewertungsraster, Zeile
> *Absehbare Phase-2-Risiken*), §8 (Konsequenzen);
> `docs/adr/0002-persistence-store.md` (löst R-4);
> `docs/adr/0003-live-updates.md` (löst R-3);
> `spec/lastenheft.md` §4.3, §10.1; `docs/planning/in-progress/roadmap.md` §4
> (Folge-ADRs).

Dieses Dokument verfolgt absehbare technische Risiken, die mit der
Backend-Stack-Entscheidung (Go) eingegangen oder nicht aufgelöst
worden sind. Folge-ADRs, die ein Risiko verbindlich entscheiden,
stehen in `docs/planning/in-progress/roadmap.md` §4; hier wird das Risiko selbst geführt,
inklusive Status, Ziel-Phase und Mitigationspfad.

Wartungsregel: aktive Risiken stehen in §1.1, aufgelöste oder nicht
eingetretene in §1.2 (kompakte Historie mit Verweis auf den
auflösenden ADR/Commit). Aufgelöste Einträge bleiben sichtbar, damit
spätere Reviewer die Begründung schnell finden — sie wandern aber aus
der Aktiv-Tabelle, damit das aktive Risiko-Set lesbar bleibt.

Statusspalte: 🟢 aufgelöst · 🟡 in Arbeit · ⬜ offen · ⬛ nicht
eingetreten.

---

## 1. Risiken aus ADR-0001 (Phase-2-Risiken)

### 1.1 Aktive Risiken

| Kennung | Risiko | Quelle | Ziel-Phase | Status | Mitigation / Folge-ADR |
|---|---|---|---|---|---|
| R-5 | Time-Skew-Persistenz auf Event-Ebene fehlt: `0.4.0` setzt `mtrace.time.skew_warning=true` als Span-Attribut (siehe `spec/telemetry-model.md` §2.5/§5.3), aber die Schema-Spalte und Dashboard-Anzeige sind explizit deferred. Folge: skew-betroffene Events sind im Read-Pfad (Dashboard ohne Tempo) nicht sichtbar markiert; Operator muss in Tempo schauen. | `plan-0.4.0.md` §3.1 | bei Triggerschwelle | ⬜ | **Stand `0.8.0` Closeout / `0.9.0`-Plan offen**: Triggerschwelle weiterhin nicht ausgelöst — weder SRT-Health-View, WebRTC-Lab noch der WebRTC-Adapter haben einen neuen Player-Event-Skew-Bezug; bestehender Aktivierungspfad bleibt unverändert. Trigger wie zuvor: ≥ 5 Spans mit `mtrace.time.skew_warning=true` außerhalb von Synthetik-Tests innerhalb einer Lab-Woche, oder Operator-Report. |
| R-7 | `SessionsService.ListSessions` lädt `network_signal_absent[]` pro Session-Page-Eintrag einzeln (`ListBoundariesForSession` N+1). Bei Hard-Cap 1000 Sessions pro Page sind das im Worst Case 1000 SQL-Round-Trips ohne gemeinsamen Tx-Snapshot (jede Query öffnet eine eigene Tx-Boundary). Schreibpfad (`POST /api/playback-events`) und Detail-Read (`GET /api/stream-sessions/{id}`) sind nicht betroffen. Wahrscheinlichkeit moderat (Lab-typisch wenige Sessions, Production unbekannt); Auswirkung: spürbare List-Latenz, kein funktionaler Bug. | `plan-0.4.0.md` §4.4 D3 (Review-N-1) | bei Triggerschwelle | ⬜ | **Stand `0.8.0` Closeout / `0.9.0`-Plan offen**: Triggerschwelle nicht ausgelöst. WebRTC-Pfade (`metrics_sampled` mit `webrtc.*`-Meta in `0.8.0`) gehen nicht durch den `network_signal_absent`-Pfad; das Risiko bleibt ausschließlich auf den `stream-sessions`-List-Read beschränkt. Aktivierungspfad und Bulk-Read-Port-Vorschlag unverändert. |
| R-9 | Die Observability-Smoke-Whitelist für Infrastruktur-Labels (`__name__`, `instance`, `job`) ist bewusst Compose-Lab-spezifisch. Eine spätere K8s-Smoke-Stage würde übliche Labels wie `pod`, `namespace` oder `container` wahrscheinlich als False Positive werten. | `plan-0.4.0.md` §7.4 (Review F-2) | K8s-Smoke-Einführung | ⬜ | **Stand `0.8.0` Closeout / `0.9.0`-Plan offen**: keine Änderung am Compose-Lab-Schema, aber die Aggregat-Allowlist ist seit `0.6.0`/`0.8.0` gewachsen — `mtrace_srt_health_*` (`health_state`, `source_status`, `source_error_code`) plus jetzt `mtrace_webrtc_*` (`connection_state`, `ice_state`, `dtls_state` plus drei label-freie Counter). Eine künftige K8s-Smoke-Stage muss diese drei Themenblöcke gemeinsam adaptieren; Folgeplan zur K8s-Anbindung ist Voraussetzung (z. B. MVP-42-Folge-Plan), Trigger und Mitigation strukturell unverändert. |
| R-10 | Sampling-Vollständigkeitsnachweis für `sampleRate < 1` fehlt: gesampelte Events verbrauchen keine `sequence_number`, deshalb kann der Server eine Sampling-Lücke nicht von einem echten Verlust unterscheiden. Folge: vollständige Timeline-Abnahme und alle E2E-Smokes laufen mit `sampleRate = 1`; Sampled-Sessions sind in `0.4.0` ausschließlich über dokumentierte Konfiguration und Operator-Hinweis als „sampled" markiert, nicht über serverseitige Lückenerkennung. | `plan-0.4.0.md` §8.3 (Beschluss Variante (b)) | bei Triggerschwelle | ⬜ | **Stand `0.8.0` Closeout / `0.9.0`-Plan offen**: Triggerschwelle nicht ausgelöst. SRT-Health- und WebRTC-`metrics_sampled`-Pfade nutzen eigene Sample-Schlüssel (SRT: `bytesReceived`-Source-Sequence; WebRTC: `webrtc.sample_id` pro `peer_connection_run_id`) — R-10 bleibt auf den Player-Event-Sampling-Pfad beschränkt. Mitigation unverändert. |
| R-11 | SRT-Health-Detail-Pagination ist in `0.6.0` Sub-3.3 als ErrNotImplemented gestubbed. Der HTTP-Pfad `GET /api/srt/health/{stream_id}` liefert nur `samples_limit` (default 100, max 1000) ohne Cursor; eine länger laufende Lab-/Production-Instanz mit > 1000 persistierten Samples pro Stream kann das ältere Drittel nicht mehr abfragen. Auswirkung im `0.6.0`-Lab gering (Lab läuft typisch < 1 h und akkumuliert wenige Samples), wird aber bei dauerhaftem Betrieb relevant. | `plan-0.6.0.md` §4 Sub-3.3 (Adapter wirft `errors.New("cursor pagination not yet implemented")`); `spec/backend-api-contract.md` §7a.3 (Wire-Format-Definition steht bereits) | bei Triggerschwelle | ⬜ | **Stand `0.8.0` Closeout / `0.9.0`-Plan offen**: Triggerschwelle nicht ausgelöst — Lab-Läufe in `0.7.0`/`0.8.0` haben weiterhin < 1 h Akkumulation. Mitigation: Cursor-Adapter-Pfad implementieren analog `EventRepository`-Cursor (process_instance_id + (ingested_at, id)-Position als opaker Token); Wire-Format ist in §7a.3 schon spezifiziert, also keine Vertrags-Änderung. Triggerschwelle: ≥ 1000 persistierte Samples pro Stream in Lab/Production, ODER Operator-Report über fehlende ältere Samples im Detail-Verlauf. |
| R-13 | Drei OS-CVEs ohne Upstream-Fix in `node:22-trixie-slim`, geteilt zwischen Dashboard- und Analyzer-Service-Image: `CVE-2025-69720` (ncurses Buffer Overflow in `libtinfo6`/`ncurses-base`/`ncurses-bin`), `CVE-2026-29111` (systemd Arbitrary Code Exec via IPC in `libsystemd0`/`libudev1`) und `CVE-2026-4878` (libcap TOCTOU-Race in `libcap2`). Keine Ausnutzbarkeit im m-trace-Runtime sichtbar (Container ohne TTY/IPC-Mounts, USER node ohne setcap-Pfad), aber statisch HIGH. Hinweis: Analyzer-Service wechselte mit plan-0.8.5 Tranche 1 Closeout von `node:22-alpine` auf `node:22-trixie-slim`, weil musl-libc bei multi-threaded Workloads (libuv-Worker-Pool, V8-GC/JIT) gegenüber glibc spürbar pessimisiert ist; die einheitliche Trixie-Basis vermeidet das. | `plan-0.8.5.md` §2 Tranche 1 Closeout; `.security/vulnignore.yaml`; `apps/dashboard/Dockerfile`, `apps/analyzer-service/Dockerfile` | bei Triggerschwelle | ⬜ | **Mitigation aktuell**: dokumentierte Trivy-Ignores in `.security/vulnignore.yaml` mit 90-Tage-`expires` (2026-08-04); Generator `scripts/render-trivyignore.sh` bricht ab, sobald ein `expires` überschritten ist (erzwungene Re-Review). **Folge-Trigger**: (a) Trixie-Point-Release liefert Fixes — Eintrag entfernen; (b) `expires`-Schwelle erreicht ohne Fix — Re-Review entscheidet zwischen Verlängerung und Base-Image-Wechsel; (c) Distroless-Variante wird vor 1.0 als Folge-Plan evaluiert (`gcr.io/distroless/nodejs22-debian12` bringt nur Node-Binär ohne npm/Debian-CLI-Surface, eliminiert die drei CVEs strukturell, behält glibc). Aktuell Lab-/Demo-Pfad — keine Production-Kontroll-Plane. |
| R-12 | WebRTC-`getStats()`-Schema-Drift zwischen Browser-Major-Versionen (Chromium/Firefox/Safari) kann die in `spec/telemetry-model.md` §3.5 spezifizierte WebRTC-Aggregat-Allowlist (`connection_state`, `ice_state`, `dtls_state`) und das `getStats()`-Subset (Muss-/Soll-Felder pro `RTCStatsType`-Gruppe) ungültig machen. **Stand `0.8.0` Tranche 3**: produktive WebRTC-Telemetrie ist ausgeliefert (`mtrace_webrtc_*`-Counter im API-Ingress, `webrtc.*`-Wire-Allowlist in `contracts/event-schema.json`, Player-SDK sammelt `getStats()` und sendet `metrics_sampled`-Events); ein Schema-Drift wirkt jetzt direkt auf den Validierungspfad (422 bei unbekannten Keys) und auf die Counter-Aggregation. | `plan-0.8.0.md` §4 Tranche 3 (RAK-53/RAK-54); `spec/telemetry-model.md` §3.5; `packages/player-sdk/src/adapters/webrtc/sampling.ts`; `apps/api/adapters/driven/metrics/webrtc_metrics.go` | release-blockierend ab nächstem Browser-Major-Bump | ⬜ | **Drift-Review-Pflicht (manuell)**: vor jedem Release-Tag prüfen, ob ein Chromium/Firefox/Safari-Major das `RTCStatsType`-Schema oder die State-Enums (`connectionState`/`iceConnectionState`/`dtlsState`) geändert hat. Bei Änderung: `webrtc.*`-Allowlist in `spec/telemetry-model.md` §1.4 + §3.5.2, `contracts/event-schema.json#reserved_meta_keys` und `packages/player-sdk/src/adapters/webrtc/sampling.ts#collectAggregate` synchron aktualisieren. **Folge-Mitigation in Planung**: `docs/planning/open/plan-0.9.0.md` §2 Tranche 1 (RAK-56 Soll) liefert einen Browser-Drift-Smoke (Playwright gegen das `mtrace-webrtc`-Lab) mit Nightly-CI-Job; ab dessen Auslieferung wandert R-12 auf „automatisiert detektiert, Drift bricht den Drift-Smoke", Manuell-Review entfällt. Wieder-Reduktion auf „bei Triggerschwelle" nur, wenn der WebRTC-Adapter wieder deaktiviert wird. |

### 1.2 Aufgelöste / nicht eingetretene Risiken

Sichtbar als Historie. Spalten kompakt (Risiko, Auflösung, Verweis).

| Kennung | Risiko (Kurzform) | Auflösung | Verweis |
|---|---|---|---|
| R-1 🟢 | Hexagon-Boundaries ohne Compile-Time-Checks | `apps/api/scripts/check-architecture.sh` (`go list`-Imports-Diff) plus `apps/api/.golangci.yml` `depguard`-Rules; beide laufen in `make gates` (`make arch-check` und `make lint`). Folge-ADR „Multi-Modul-Aufteilung" damit nicht mehr nötig. Wieder-Eröffnung, falls Static-Analysis bei Refactor nicht mehr greift. | `make arch-check`, `make lint`; ADR-0001 §5/§8 |
| R-2 🟢 | CGO-basierte SRT-Bindings vs. `distroless-static`-Pattern | `0.6.0` Tranche 1 wählt MediaMTX-/Server-API über HTTP als SRT-Metrikquelle (Sub-1.2 Probe nachweist alle vier RAK-43-Pflichtwerte direkt; Sub-1.3 fixiert die Wahl). `apps/api` bleibt CGO-frei, `distroless-static`-Pattern unangefasst. Folge-ADR „SRT-Binding-Stack" damit obsolet — Roadmap §4 wird im 0.6.0-Closeout entsprechend bereinigt. Wieder-Eröffnung, falls die HTTP-Quelle in einer Folge-Phase nicht mehr trägt und libsrt-Binding doch nötig wird. | [`plan-0.6.0.md`](../done/plan-0.6.0.md) §2.4/§2.5; [`spec/contract-fixtures/srt/mediamtx-srtconns-list.json`](../../../spec/contract-fixtures/srt/mediamtx-srtconns-list.json) |
| R-3 🟢 | WebSocket-Ökosystem in Go fragmentiert | ADR-0003 entscheidet SSE mit Polling-Fallback; WebSocket bleibt für `0.4.0` aus dem Scope. | `docs/adr/0003-live-updates.md` |
| R-4 🟢 | In-Memory-Persistenz verliert Sessions/Events bei Restart | ADR-0002 entscheidet SQLite als lokalen Durable-Store; SQLite-Adapter, Cursor-v2 und Restart-Stabilität sind in `plan-0.4.0.md` Tranche 1 (§2.1–§2.6) geliefert. | `docs/adr/0002-persistence-store.md`; `done/plan-0.4.0.md` §2 |
| R-6 🟢 | `correlation_id`-Race bei konkurrenter Session-Anlage | `plan-0.4.0.md` §4.2 C2 (`949a265`): `UpsertFromEvents` liefert die DB-finale `correlation_id` jeder Session als Map-Rückgabe; SQLite prüft `RowsAffected()` und liest die Sieger-CID nach. Race-Test `TestUpsertFromEvents_RaceCanonicalCorrelationID` (8 Goroutines) zeigt: alle Aufrufe liefern dieselbe Sieger-CID. Wieder-Eröffnung bei erneutem Mismatch vor Release-Bump. | `done/plan-0.4.0.md` §4.2 C2; `949a265` |
| R-8 🟢 | Forbidden-Label-Smoke fängt neue Credential-/URL-Formen evtl. nicht | `plan-0.4.0.md` §8.2 (`59f66dc`): Forbidden-Pattern zentralisiert in `spec/telemetry-model.md` §3.1 als normative Quelle (`session_id`, URLs/URL-Teile, `user_agent`, `client_ip`, `viewer_id`, `request_id`, `trace_id`, `span_id`, `correlation_id`, `project_id` ohne Allowlist, Token-/Credential-Felder, `batch_size`); `scripts/smoke-observability.sh` spiegelt diese Liste deckungsgleich plus Suffix-Regeln (`_url`, `_uri`, `_token`, `_secret`); `policyProbe`-Self-Tests erweitert. Wieder-Eröffnung, falls neue Form weder durch Namen noch durch Suffix gefangen wird. | `spec/telemetry-model.md` §3.1; `done/plan-0.4.0.md` §8.2; `59f66dc` |

---

## 2. Wartung

- Bei einem neuen Folge-ADR, der ein Risiko verbindlich auflöst,
  wird der Eintrag in §1.2 verschoben (Status 🟢) mit Verweis auf
  ADR-Nummer und auflösenden Commit. Die ausführliche Mitigations-
  Beschreibung wird auf eine Kurzform reduziert; lebende Verweise
  bleiben.
- Neue Risiken, die im Verlauf der Implementierung auftauchen und
  nicht direkt durch einen Folge-ADR adressiert werden können,
  bekommen eine fortlaufende `R-N`-Kennung und landen in §1.1.
- Status-Änderungen folgen demselben Statusset wie `docs/planning/in-progress/roadmap.md`
  §2/§3, ergänzt um ⬛ für Risiken, die sich nicht materialisiert
  haben (landen ebenfalls in §1.2).
- Stand-Datum im Header wird beim nächsten substanziellen Update
  einer §1.1-Zeile mitgepflegt.
