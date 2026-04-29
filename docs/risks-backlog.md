# Risiken-Backlog

> **Stand**: 2026-04-29  
> **Bezug**: `docs/adr/0001-backend-stack.md` §5 (Bewertungsraster, Zeile
> *Absehbare Phase-2-Risiken*), §8 (Konsequenzen),
> `docs/lastenheft.md` §4.3, §10.1; `docs/roadmap.md` §4
> (Folge-ADRs).

Dieses Dokument verfolgt absehbare technische Risiken, die mit der
Backend-Stack-Entscheidung (Go) eingegangen oder nicht aufgelöst
worden sind. Folge-ADRs, die ein Risiko verbindlich entscheiden,
stehen in `docs/roadmap.md` §4; hier wird das Risiko selbst geführt,
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
| R-3 | WebSocket-Ökosystem in Go ist fragmentiert (`gorilla/websocket`, `nhooyr.io/websocket`, `coder/websocket`) — kein klarer stdlib-Default. Zusätzlich vergrößert ein WebSocket-Pfad die Surface gegenüber HTTP-only-Endpunkten. | ADR-0001 §5 | `0.4.0` (Live-Updates) | ⬜ | Folge-ADR „WebSocket vs. SSE" (Roadmap §4); SSE oder Polling als Fallback bewusst behalten. |

---

## 2. Wartung

- Bei einem neuen Folge-ADR, der ein Risiko verbindlich auflöst,
  wird der Eintrag mit Status 🟢 markiert und zeitnah entfernt
  (Verweis auf ADR-Nummer und Commit im Commit-Body).
- Neue Risiken, die im Verlauf der Implementierung auftauchen und
  nicht direkt durch einen Folge-ADR adressiert werden können,
  bekommen eine fortlaufende `R-N`-Kennung.
- Status-Änderungen folgen demselben Statusset wie `docs/roadmap.md`
  §2/§3, ergänzt um ⬛ für Risiken, die sich nicht materialisiert
  haben.
