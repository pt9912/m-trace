# Risiken-Backlog

> **Stand**: 2026-04-29  
> **Bezug**: `docs/adr/0001-backend-stack.md` §5 (Bewertungsraster, Zeile
> *Absehbare Phase-2-Risiken*), §8 (Konsequenzen),
> `spec/lastenheft.md` §4.3, §10.1; `docs/planning/roadmap.md` §4
> (Folge-ADRs).

Dieses Dokument verfolgt absehbare technische Risiken, die mit der
Backend-Stack-Entscheidung (Go) eingegangen oder nicht aufgelöst
worden sind. Folge-ADRs, die ein Risiko verbindlich entscheiden,
stehen in `docs/planning/roadmap.md` §4; hier wird das Risiko selbst geführt,
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
| R-1 | Hexagon-Boundaries werden nur über Disziplin und Code-Review erzwungen — Go im Single-Modul-Layout kennt keine Compile-Time-Boundary-Checks (anders als Gradle-Multi-Modul mit `implementation`-Dependencies). | ADR-0001 §5, §8 *Bleibt* | on demand | ⬜ | Folge-ADR „`apps/api` Multi-Modul-Aufteilung (`go.work`)" (Roadmap §4); auslösendes Signal: Boundary-Verstöße im Review oder bei wachsender Codebase. |
| R-2 | CGO-basierte SRT-Bindings könnten das `distroless-static`-Pattern brechen. `apps/api`-Image müsste auf eine `glibc`-Runtime wechseln (z. B. `gcr.io/distroless/base-debian12`), Cold-Start- und Image-Größen-Vorteile aus dem Spike (10,2 MB / 9 ms) wären teilweise weg. | ADR-0001 §5, §8 *Neu*; Lastenheft §4.3 | `0.6.0` (SRT Health) | ⬜ | Folge-ADR „SRT-Binding-Stack" (Roadmap §4); cgo-freie SRT-Library prüfen, alternativ alternative Runtime evaluieren. |
| R-3 | WebSocket-Ökosystem in Go ist fragmentiert (`gorilla/websocket`, `nhooyr.io/websocket`, `coder/websocket`) — kein klarer stdlib-Default. Zusätzlich vergrößert ein WebSocket-Pfad die Surface gegenüber HTTP-only-Endpunkten. | ADR-0001 §5 | `0.4.0` (Live-Updates) | ✅ | ADR-0003 entscheidet SSE mit Polling-Fallback; WebSocket bleibt für `0.4.0` aus dem Scope. |
| R-4 | In-Memory-Persistenz verliert Sessions und Events bei Restart; Cursor werden über `process_instance_id` absichtlich invalidiert und sind nicht durable. | OE-3, MVP-16, Architektur §11 | `0.4.0` | 🟢 | ADR-0002 entscheidet SQLite als lokalen Durable-Store; SQLite-Adapter, Cursor-v2 und Restart-Stabilität sind in `plan-0.4.0.md` Tranche 1 (§2.1–§2.6) geliefert. |
| R-5 | Time-Skew-Persistenz auf Event-Ebene fehlt: `0.4.0` setzt `mtrace.time.skew_warning=true` als Span-Attribut (siehe `spec/telemetry-model.md` §2.5/§5.3), aber die Schema-Spalte und Dashboard-Anzeige sind explizit deferred. Folge: skew-betroffene Events sind im Read-Pfad (Dashboard ohne Tempo) nicht sichtbar markiert; Operator muss in Tempo schauen. | `plan-0.4.0.md` §3.1 | `0.5.0` oder Folge-Tranche bei Bedarf | ⬜ | Schema-Spalte (`time_skew_warning BOOLEAN`) ergänzen + EventRepository füllt sie + Dashboard-Timeline blendet sie ein. Triggerschwelle: ≥ 5 Spans mit `mtrace.time.skew_warning=true` außerhalb von Synthetik-Tests innerhalb einer Lab-Woche, oder ein Operator-Report, in dem skew-betroffene Events nicht über Tempo auffindbar gemacht werden konnten. |

---

## 2. Wartung

- Bei einem neuen Folge-ADR, der ein Risiko verbindlich auflöst,
  wird der Eintrag mit Status 🟢 markiert und zeitnah entfernt
  (Verweis auf ADR-Nummer und Commit im Commit-Body).
- Neue Risiken, die im Verlauf der Implementierung auftauchen und
  nicht direkt durch einen Folge-ADR adressiert werden können,
  bekommen eine fortlaufende `R-N`-Kennung.
- Status-Änderungen folgen demselben Statusset wie `docs/planning/roadmap.md`
  §2/§3, ergänzt um ⬛ für Risiken, die sich nicht materialisiert
  haben.
