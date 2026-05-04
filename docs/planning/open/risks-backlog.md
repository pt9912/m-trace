# Risiken-Backlog

> **Stand**: 2026-04-29  
> **Bezug**: `docs/adr/0001-backend-stack.md` §5 (Bewertungsraster, Zeile
> *Absehbare Phase-2-Risiken*), §8 (Konsequenzen),
> `spec/lastenheft.md` §4.3, §10.1; `docs/planning/in-progress/roadmap.md` §4
> (Folge-ADRs).

Dieses Dokument verfolgt absehbare technische Risiken, die mit der
Backend-Stack-Entscheidung (Go) eingegangen oder nicht aufgelöst
worden sind. Folge-ADRs, die ein Risiko verbindlich entscheiden,
stehen in `docs/planning/in-progress/roadmap.md` §4; hier wird das Risiko selbst geführt,
inklusive Status, Ziel-Phase und Mitigationspfad.

Wartungsregel: ein Eintrag bleibt im Backlog, bis er durch einen
Folge-ADR aufgelöst, durch eine Architekturentscheidung neutralisiert
oder als nicht eingetreten markiert worden ist. Aufgelöste Einträge
werden mit Verweis auf den auflösenden ADR oder Commit gestrichen.

Statusspalte: 🟢 aufgelöst · 🟡 in Arbeit · ⬜ offen · ⬛ nicht
eingetreten.

---

## 1. Risiken aus ADR-0001 (Phase-2-Risiken)

| Kennung | Risiko | Quelle | Ziel-Phase | Status | Mitigation / Folge-ADR |
|---|---|---|---|---|---|
| R-1 | Hexagon-Boundaries werden nur über Disziplin und Code-Review erzwungen — Go im Single-Modul-Layout kennt keine Compile-Time-Boundary-Checks (anders als Gradle-Multi-Modul mit `implementation`-Dependencies). | ADR-0001 §5, §8 *Bleibt* | on demand | 🟢 | Aufgelöst durch alternative Mitigation statt Multi-Modul-ADR: `apps/api/scripts/check-architecture.sh` (`go list`-Imports-Diff gegen die vier Boundary-Regeln aus `spec/architecture.md` §3.2/§3.4) plus `apps/api/.golangci.yml` `depguard`-Rules (deckungsgleich). Beide laufen in `make gates` (`make arch-check` und `make lint`); Compile-Time-Check ist durch zwei unabhängige CI-Time-Checks ersetzt. Folge-ADR „`apps/api` Multi-Modul-Aufteilung (`go.work`)" damit nicht mehr erforderlich; siehe Roadmap §4. Wieder-Eröffnung, falls die Static-Analysis-Tools bei einem zukünftigen Refactor nicht mehr greifen. |
| R-2 | CGO-basierte SRT-Bindings könnten das `distroless-static`-Pattern brechen. `apps/api`-Image müsste auf eine `glibc`-Runtime wechseln (z. B. `gcr.io/distroless/base-debian12`), Cold-Start- und Image-Größen-Vorteile aus dem Spike (10,2 MB / 9 ms) wären teilweise weg. | ADR-0001 §5, §8 *Neu*; Lastenheft §4.3 | `0.6.0` (SRT Health) | ⬜ | Folge-ADR „SRT-Binding-Stack" (Roadmap §4); cgo-freie SRT-Library prüfen, alternativ alternative Runtime evaluieren. |
| R-3 | WebSocket-Ökosystem in Go ist fragmentiert (`gorilla/websocket`, `nhooyr.io/websocket`, `coder/websocket`) — kein klarer stdlib-Default. Zusätzlich vergrößert ein WebSocket-Pfad die Surface gegenüber HTTP-only-Endpunkten. | ADR-0001 §5 | `0.4.0` (Live-Updates) | ✅ | ADR-0003 entscheidet SSE mit Polling-Fallback; WebSocket bleibt für `0.4.0` aus dem Scope. |
| R-4 | In-Memory-Persistenz verliert Sessions und Events bei Restart; Cursor werden über `process_instance_id` absichtlich invalidiert und sind nicht durable. | OE-3, MVP-16, Architektur §11 | `0.4.0` | 🟢 | ADR-0002 entscheidet SQLite als lokalen Durable-Store; SQLite-Adapter, Cursor-v2 und Restart-Stabilität sind in `plan-0.4.0.md` Tranche 1 (§2.1–§2.6) geliefert. |
| R-5 | Time-Skew-Persistenz auf Event-Ebene fehlt: `0.4.0` setzt `mtrace.time.skew_warning=true` als Span-Attribut (siehe `spec/telemetry-model.md` §2.5/§5.3), aber die Schema-Spalte und Dashboard-Anzeige sind explizit deferred. Folge: skew-betroffene Events sind im Read-Pfad (Dashboard ohne Tempo) nicht sichtbar markiert; Operator muss in Tempo schauen. | `plan-0.4.0.md` §3.1 | `0.5.0` oder Folge-Tranche bei Bedarf | ⬜ | Schema-Spalte (`time_skew_warning BOOLEAN`) ergänzen + EventRepository füllt sie + Dashboard-Timeline blendet sie ein. Triggerschwelle: ≥ 5 Spans mit `mtrace.time.skew_warning=true` außerhalb von Synthetik-Tests innerhalb einer Lab-Woche, oder ein Operator-Report, in dem skew-betroffene Events nicht über Tempo auffindbar gemacht werden konnten. |
| R-6 | `correlation_id`-Race bei konkurrenter Erstanlage derselben `session_id`: zwei parallele Use-Case-Aufrufe rufen beide `sessions.Get` → `ErrSessionNotFound`, generieren je eine eigene UUIDv4 und schreiben sie auf ihre Events. Der SQLite-Adapter wendet `INSERT ... ON CONFLICT DO NOTHING` an, sodass der zweite Insert keinen 5xx erzeugt — aber `playback_events.correlation_id` der Verlust-Race-Events trägt nicht den `stream_sessions.correlation_id`. Wahrscheinlichkeit gering (Player schickt sequentiell, session_id ist 26-Char-ULID), Konsistenz-Verstoß aber real. | `plan-0.4.0.md` §3.2 (Review K-1) | offen | 🟢 | Technisch geschlossen mit `plan-0.4.0.md` §4.2 C2 (`949a265`): `UpsertFromEvents` liefert die DB-finale `correlation_id` jeder Session als Map-Rückgabe (`[sessionID]canonicalCID`), der Use-Case enricht damit die Events vor `EventRepository.Append`. Der SQLite-Adapter prüft `RowsAffected()` nach dem `ON CONFLICT (project_id, session_id) DO NOTHING`-Insert und liest die Sieger-CID nach, falls der eigene Insert zur No-op wurde. Race-Test `TestUpsertFromEvents_RaceCanonicalCorrelationID` (8 Goroutines, gleiche `(project, session)`, unterschiedliche Kandidat-CIDs) zeigt: alle Aufrufe liefern dieselbe Sieger-CID, eine Zeile in `stream_sessions`, kein Mismatch zwischen `playback_events.correlation_id` und `stream_sessions.correlation_id`. Wieder-Eröffnung, falls vor Release-Bump (`plan-0.4.0.md` Tranche 8) erneut ein Mismatch beobachtet wird. |
| R-7 | `SessionsService.ListSessions` lädt `network_signal_absent[]` pro Session-Page-Eintrag einzeln (`ListBoundariesForSession` N+1). Bei Hard-Cap 1000 Sessions pro Page sind das im Worst Case 1000 SQL-Round-Trips ohne gemeinsamen Tx-Snapshot (jede Query öffnet eine eigene Tx-Boundary). Schreibpfad (`POST /api/playback-events`) und Detail-Read (`GET /api/stream-sessions/{id}`) sind nicht betroffen. Wahrscheinlichkeit moderat (Lab-typisch wenige Sessions, Production unbekannt); Auswirkung: spürbare List-Latenz, kein funktionaler Bug. | `plan-0.4.0.md` §4.4 D3 (Review-N-1) | `0.5.0` oder Folge-Tranche bei Bedarf | ⬜ | Bulk-Read-Port `ListBoundariesForSessions(ctx, projectID, sessionIDs []string) (map[string][]SessionBoundary, error)` als additive Erweiterung des `driven.SessionRepository`. SQLite löst das mit `SELECT ... WHERE project_id = ? AND session_id IN (?, ?, ...)` in einer Query, InMemory iteriert über die session-IDs. HTTP-Adapter konsumiert die Map indexbasiert, kein Vertragsbruch. Triggerschwelle: ≥ 200 ms p95-Latenz auf `GET /api/stream-sessions` im Lab-Smoke, oder Operator-Report über spürbare List-Latenz nach erstem produktiven Deploy. |

---

## 2. Wartung

- Bei einem neuen Folge-ADR, der ein Risiko verbindlich auflöst,
  wird der Eintrag mit Status 🟢 markiert und zeitnah entfernt
  (Verweis auf ADR-Nummer und Commit im Commit-Body).
- Neue Risiken, die im Verlauf der Implementierung auftauchen und
  nicht direkt durch einen Folge-ADR adressiert werden können,
  bekommen eine fortlaufende `R-N`-Kennung.
- Status-Änderungen folgen demselben Statusset wie `docs/planning/in-progress/roadmap.md`
  §2/§3, ergänzt um ⬛ für Risiken, die sich nicht materialisiert
  haben.
