# Roadmap

> **Stand**: 2026-04-28  
> **Phase**: Post-Spike, Pre-MVP `0.1.0`  
> **Bezug**: `docs/lastenheft.md` RAK-1..RAK-46 (Release-Plan, normativ),
> `docs/adr/0001-backend-stack.md` (Backend-Entscheidung),
> `docs/plan-spike.md` SP-41 (Anschluss an MVP),
> `docs/spike/backend-stack-results.md` (Spike-Protokoll).

Dieses Dokument ist die **Statusseite** des Projekts. Es duplikiert nicht
die Anforderungen pro Release (die stehen normativ im Release-Plan des
Lastenheft), sondern verfolgt: *Wo sind wir, was kommt als nächstes,
welche Risiken und Folge-Entscheidungen liegen vor uns.*

Wartungsregel: nach jedem Release-Bump (z. B. `0.0.x → 0.1.0`) und nach
jedem Folge-ADR aktualisieren.

---

## 1. Aktueller Stand (2026-04-28)

### 1.1 Was abgeschlossen ist

| Status | Bereich | Ergebnis | Verweise |
|---|---|---|---|
| ✅ | Lastenheft | `v0.7.0` mit Anforderungen nach IDs (`F-`, `NF-`, `MVP-`, `AK-`, `RAK-`, `OE-`) und Release-Plan vollständig versioniert. | `docs/lastenheft.md` |
| ✅ | Backend-Spike | Zwei Prototypen (Go, Micronaut) im identischen Muss-Scope abgeschlossen, Vergleich nach Plan-SP-30 (Bewertungskriterien) erfolgt, Sieger ist Go. | `docs/spike/0001-backend-stack.md`, `docs/spike/backend-stack-results.md`, `docs/plan-spike.md` (SP-30), `docs/plan-spike.md` (SP-41) |
| ✅ | API-Kontrakt | Spike-API-Kontrakt erstellt, dokumentiert und eingefroren (`frozen`). | `docs/spike/backend-api-contract.md` |
| ✅ | ADR | Backend-Stack-Entscheidung entschieden und als **Accepted** festgehalten. | `docs/adr/0001-backend-stack.md` |
| ✅ | Siegerbranch | `spike/go-api` finalisiert (Commit `7148a8d`) als Basis für `apps/api` in `0.1.0`. | `spike/go-api`, ADR |
| ✅ | Unterlegener Branch | Als Tag archiviert: `spike/backend-stack-loser-2026-04-28` (Commit `7c8bc44`), `spike/micronaut-api` gelöscht. | `spike/backend-stack-loser-2026-04-28` |

### 1.2 Was noch offen ist (vor MVP `0.1.0`)

Reihenfolge ist verbindlich (SP-41).

| Reihenfolge | Status | Aufgabe | Trigger | Verweis |
|---|---|---|---|---|
| 1 | ✅ | `spike/go-api` zum `apps/api`-Skelett auf `main` ausbauen (MVP-2). | Sofort | OE-9; SP-41 |
| 2 | ✅ | Lastenheft auf `1.0.0` heben: Backend-Entscheidung einarbeiten, offene Entscheidungen reduzieren. | Nach Schritt 1 | OE-2; OE-9; SP-41 |
| 3 | ✅ | `README.md` Tech-Overview auf den gewählten Stack anpassen (Go 1.22 + stdlib + Prometheus + OTel + distroless). | Nach Schritt 2 | MVP-17; SP-41 |
| 4 | ⬜ | Phase-2-Risiken aus ADR §8 in den Issue-Backlog überführen (Form: siehe §5). | Nach Schritt 3 | SP-41 |

Erst danach beginnt die eigentliche `0.1.0`-Implementierung:
Dashboard, Player-SDK, Docker-Lab und Observability.

---

## 2. Nächste Schritte

Statusspalte: ✅ abgeschlossen · 🟡 in Arbeit · ⬜ geplant.

Verweise nutzen die Lastenheft-Kennungen (`F-`, `NF-`, `MVP-`, `AK-`)
wo sie existieren; Plan- und ADR-Sektionsnummern werden behalten,
weil dort kein ID-System existiert.

| # | Status | Schritt | Trigger | Verweis |
|---|---|---|---|---|
| 1 | ✅ | `spike/go-api` → `apps/api` auf `main` integrieren | Sofort | MVP-2; OE-9; SP-41 |
| 2 | ✅ | Lastenheft auf `1.0.0` heben | Nach Schritt 1 | OE-2; OE-9; SP-41 |
| 3 | ✅ | README Tech-Overview anpassen | Nach Schritt 2 | MVP-17; SP-41 |
| 4 | ⬜ | Phase-2-Risiken in Issue-Backlog | Nach Schritt 3 | SP-41 |
| 5 | ⬜ | `docs/architecture.md` schreiben | Vor `0.1.0`-DoD | AK-3, AK-10 |
| 6 | ⬜ | `docs/telemetry-model.md` schreiben | Vor `0.1.0`-DoD | F-89..F-94, F-106..F-115, AK-9 |
| 7 | ⬜ | `docs/local-development.md` schreiben | Vor `0.1.0`-DoD | AK-1, AK-2 |
| 8 | ⬜ | Dashboard-App (`apps/dashboard`) anlegen | Nach Schritt 1 | MVP-3; F-23..F-28 |
| 9 | ⬜ | Player-SDK (`packages/player-sdk`) anlegen | Nach Schritt 1 | MVP-5; F-63..F-65 |
| 10 | ⬜ | Docker-Compose-Lab inkl. MediaMTX + FFmpeg | Nach Schritt 1 | MVP-7..MVP-9; F-82..F-88 |
| 11 | ⬜ | Observability-Stack (Prometheus, optional Grafana, OTel-Collector) | Nach Schritt 1 | MVP-10, MVP-15; F-89..F-94 |

---

## 3. Release-Übersicht

Statusspalte: ✅ abgeschlossen · 🟡 in Arbeit · ⬜ geplant.

| Version | Titel | Status | Akzeptanzkriterien |
|---|---|---|---|
| `0.0.x` | Spike + Planungsphase | ✅ | — |
| `0.1.0` | OTel-native Local Demo | ⬜ | RAK-1..RAK-10 |
| `0.2.0` | Publizierbares Player SDK | ⬜ | RAK-11..RAK-21 |
| `0.3.0` | Stream Analyzer | ⬜ | RAK-22..RAK-28 |
| `0.4.0` | Erweiterte Trace-Korrelation | ⬜ | RAK-29..RAK-35 |
| `0.5.0` | Multi-Protocol Lab | ⬜ | RAK-36..RAK-40 |
| `0.6.0` | SRT Health View | ⬜ | RAK-41..RAK-46 |

DoD für die erste Phase ist über **AK-1..AK-11** abgedeckt
(Lastenheft-übergreifend, nicht Release-spezifisch).

---

## 4. Folge-ADRs

Aus `docs/adr/0001-backend-stack.md` §8 erwartete Folge-ADRs.
Alle sind ⬜ geplant; ADR-Nummer wird beim Schreiben vergeben.

| Erwartete ADR | Trigger-Release | Begründung |
|---|---|---|
| Persistenz-Wechsel In-Memory → SQLite/PostgreSQL (**MVP-16**) | `0.1.0`–`0.2.0` | Spike-In-Memory ist nicht ausreichend, sobald Sessions persistiert werden sollen. |
| WebSocket vs. SSE für Live-Updates | `0.4.0` | Live-Update-Mechanismus für Trace/Session-Ansicht. |
| SRT-Binding-Stack | `0.6.0` | CGO-Bindings könnten das distroless-static-Pattern brechen. |
| Coverage-Tooling für Go (`go test -cover` + Threshold) | `0.1.0`+ | Coverage-Strategie analog zu d-migrate-Pattern. |
| `apps/api` Multi-Modul-Aufteilung (`go.work`) | offen | Wird nur relevant, wenn Hexagon-Boundaries Disziplin-basiert nicht reichen. |

Neue Folge-ADRs werden hier ergänzt, sobald der Bedarf entsteht oder
ein Issue darauf hinweist.

---

## 5. Offene Entscheidungen

Verbleibende Lastenheft-`OE-X` plus ein roadmap-spezifischer Punkt; aufgelöste Einträge sind nach §7-Wartungsregel entfernt.

| Kennung | Entscheidung | Wo wird sie getroffen | Status |
|---|---|---|---|
| — | Issue-Backlog-Form (GitHub Issues / Markdown-TODO / Linear / …) | mit Schritt 4 in §2 | offen, roadmap-spezifisch |
| OE-1 | Projektlizenz: MIT oder Apache-2.0 | vor `0.1.0` Public-Release | MIT bereits committed (`LICENSE`); Apache-2.0-Prüfung offen |
| OE-3 | Datenhaltung im MVP (In-Memory vs. SQLite/PostgreSQL) — verknüpft mit **MVP-16** | erste Folge-ADR (`0.1.0`–`0.2.0`) | offen |
| OE-4 | Frontend-Styling (eigenes CSS / Tailwind / UI-Library) | mit Schritt 8 in §2 | offen |
| OE-5 | Live-Updates: Polling / WebSocket / SSE | Folge-ADR `0.4.0` | offen |
| OE-6 | CI-Zielplattformen | mit Schritt 4 in §2 | offen |
| OE-7 | Release-Konvention | vor `0.1.0` Public-Release | offen |
| OE-8 | Paketnamen für npm | Schritt 9 in §2 | offen |

---

## 6. Lessons-learned aus Spike (Verdichtung)

Vollständige Notizen in `docs/spike/backend-stack-results.md`. Hier nur
die für `0.1.0`+ relevanten Punkte:

- **Hexagon ohne DI-Container-Druck**: Go braucht keine
  Annotation-Magie; `var _ Interface = (*Impl)(nil)`-Compile-Time-Checks
  pro Adapter reichen. Beibehalten.
- **Test-Stack einheitlich**: `testing` + `httptest` deckt Unit und
  Integration ab. Keine externen Test-Frameworks erforderlich.
- **Linting**: `golangci-lint` mit Default-Lintern
  (`govet`, `errcheck`, `staticcheck`, `unused`, `ineffassign`).
  `make lint` als Soll-Target im Dockerfile.
- **Docker-only-Workflow**: alle Build-/Test-/Lint-Schritte über
  `docker build --target ...`. Lokales Go ist optional. Pattern aus
  `docs/plan-spike.md` §14.11 wird beibehalten.
- **CI-Artifacts** (SP-41 Lessons-learned): Test-Results,
  Coverage-Reports, Lint-Reports beim CI-Setup hochladen — Pattern
  analog zu `d-migrate/.github/workflows/build.yml`.
- **Multi-Modul-Aufteilung erst on demand**: bei wachsender
  Codebase `apps/api/` per `go.work` oder Sub-Modul-Splits aufteilen.
  Im Spike bewusst Single-Modul für Übersicht.

---

## 7. Wartung dieses Dokuments

- Statusspalten in §2 und §3 nach jedem abgeschlossenen Schritt
  bzw. neuen Release-Tag aktualisieren (✅).
- Nach jedem neuen Folge-ADR Eintrag in §4 ergänzen oder erledigte
  ADRs aus §4 herausnehmen.
- Nach jeder gelösten offenen Entscheidung Eintrag in §5 entfernen
  und (falls strukturell) in das Lastenheft übernehmen.
- §1 Aktueller Stand wird nach jedem signifikanten Meilenstein neu
  geschrieben (nicht inkrementell — die Liste bleibt kurz).
