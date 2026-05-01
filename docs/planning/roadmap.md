# Roadmap

> **Stand**: 2026-05-01
> **Phase**: Post-`0.3.0`, vor `0.4.0` Erweiterte Trace-Korrelation
> **Bezug**: `spec/lastenheft.md` RAK-1..RAK-46 (Release-Plan, normativ),
> `spec/architecture.md` (Zielbild),
> Plan-Dokumente pro Release in `docs/planning/plan-X.Y.Z.md`,
> ADRs in `docs/adr/`.

Dieses Dokument ist die **Statusseite** des Projekts. Es duplikiert nicht
die Anforderungen pro Release (die stehen normativ im Release-Plan des
Lastenheft), sondern verfolgt: *Wo sind wir, was kommt als nächstes,
welche Risiken und Folge-Entscheidungen liegen vor uns.*

Wartungsregel: nach jedem Release-Bump und nach jedem Folge-ADR
aktualisieren.

---

## 1. Aktueller Stand (2026-05-01)

### 1.1 Was abgeschlossen ist

| Status | Bereich                  | Ergebnis                                                                                                                              | Verweise                                                                                          |
| ------ | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| ✅      | Lastenheft               | `v0.7.0` mit verbindlichem Release-Plan; aktuell `1.1.7`.                                                                             | `spec/lastenheft.md`                                                                              |
| ✅      | Architektur + ADRs       | `0001` Backend-Stack (Go) Accepted; `0002` Persistenz (Draft, OE-3 offen).                                                           | `docs/adr/0001-backend-stack.md`, `docs/adr/0002-persistence-store.md`                            |
| ✅      | Backend Core (`0.1.0`)    | API-Skelett, Compose-Lab, RAK-1/3/4/6/8.                                                                                              | [`plan-0.1.0.md`](./plan-0.1.0.md)                                                                |
| ✅      | Player-SDK + Dashboard (`0.1.1`) | Dashboard, Demo-Player, hls.js-Adapter, Session-Ansicht.                                                                       | [`plan-0.1.1.md`](./plan-0.1.1.md)                                                                |
| ✅      | Observability (`0.1.2`)   | Prometheus + Grafana + OTel-Collector als Profil; RAK-9, RAK-10.                                                                     | [`plan-0.1.2.md`](./plan-0.1.2.md)                                                                |
| ✅      | Publizierbares Player-SDK (`0.2.0`) | `@npm9912/player-sdk` mit ESM/CJS/IIFE, Pack-Smokes, Browser-Support-Matrix; RAK-11..RAK-21.                              | [`plan-0.2.0.md`](./plan-0.2.0.md)                                                                |
| ✅      | Stream-Analyzer (`0.3.0`) | `@npm9912/stream-analyzer` (Library + CLI), `analyzer-service` (interner HTTP-Wrapper), `POST /api/analyze`; RAK-22..RAK-28.        | [`plan-0.3.0.md`](./plan-0.3.0.md)                                                                |

### 1.2 Verbleibend für `0.4.0`-Scope-Cut

`0.3.0` ist veröffentlicht; `0.4.0` (Erweiterte Trace-Korrelation, RAK-29..RAK-35)
ist noch nicht geplant. Vor dem Scope-Cut:

| Reihenfolge | Status | Aufgabe                                                                                                | Trigger                | Verweis                                |
| ----------- | ------ | ------------------------------------------------------------------------------------------------------ | ---------------------- | -------------------------------------- |
| 1           | ⬜      | OE-3/Persistenz neu bewerten — `0.3.0` ist stateless, ADR-0002-Draft offen.                            | Vor `0.4.0`-Plan       | OE-3; MVP-16; ADR-0002                 |
| 2           | ⬜      | OE-5/Live-Updates: Polling vs. WebSocket vs. SSE — Folge-ADR für `0.4.0`.                              | Vor `0.4.0`-Plan       | OE-5                                   |
| 3           | ⬜      | `docs/planning/plan-0.4.0.md` anlegen und Scope in Tranchen schneiden.                                 | Nach OE-3/OE-5         | RAK-29..RAK-35                         |

---

## 2. Nächste Schritte

Statusspalte: ✅ abgeschlossen · 🟡 in Arbeit · ⬜ geplant.

Verweise nutzen die Lastenheft-Kennungen (`F-`, `NF-`, `MVP-`, `AK-`)
wo sie existieren; Plan- und ADR-Sektionsnummern werden behalten,
weil dort kein ID-System existiert. Granularer Lieferstand pro Release
steht in den jeweiligen Plan-Dateien mit DoD-Checkboxen und
Commit-Hashes, z. B. [`docs/planning/plan-0.3.0.md`](./plan-0.3.0.md).

| #   | Status | Schritt                                                                                                               | Trigger                                                         | Verweis                                                   |
| --- | ------ | --------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------- |
| 1   | ✅      | `spike/go-api` → `apps/api` auf `main` integrieren                                                                    | Sofort                                                          | MVP-2; OE-9; SP-41                                        |
| 2   | ✅      | Lastenheft auf `1.0.0` heben                                                                                          | Nach Schritt 1                                                  | OE-2; OE-9; SP-41                                         |
| 3   | ✅      | README Tech-Overview anpassen                                                                                         | Nach Schritt 2                                                  | MVP-17; SP-41                                             |
| 4   | ✅      | Phase-2-Risiken in `docs/planning/risks-backlog.md`                                                                            | Nach Schritt 3                                                  | SP-41                                                     |
| 5   | ✅      | `spec/architecture.md` schreiben                                                                                      | Vor `0.1.0`-DoD                                                 | AK-3, AK-10                                               |
| 6   | ✅      | `spec/telemetry-model.md` schreiben (Datenmodell, Wire-Format, Cardinality — kein Observability-Setup)                | Vor `0.1.0`-DoD                                                 | F-91, F-92, F-95..F-105, F-106..F-115, F-118..F-130, AK-9 |
| 7   | ✅      | `docs/user/local-development.md` schreiben                                                                                 | Vor `0.1.0`-DoD                                                 | AK-1, AK-2                                                |
| 8   | ✅      | Dashboard-App (`apps/dashboard`) anlegen — `0.1.1` (siehe `plan-0.1.1.md`)                                            | Nach `0.1.0`-Release                                            | MVP-3; F-23..F-28                                         |
| 9   | ✅      | Player-SDK (`packages/player-sdk`) anlegen — `0.1.1` (siehe `plan-0.1.1.md`)                                          | Nach `0.1.0`-Release                                            | MVP-5; F-63..F-67                                         |
| 10  | ✅      | Docker-Compose-Lab inkl. MediaMTX + FFmpeg (Core in `0.1.0`, `dashboard` in `0.1.1`, observability-Profil in `0.1.2`) | Core: vor `0.1.0`-DoD; Erweiterungen mit jeweiligem Sub-Release | MVP-7..MVP-9; F-82..F-88                                  |
| 11  | ✅      | Observability-Stack (Prometheus + optional Grafana, OTel-Collector) — `0.1.2` (siehe `plan-0.1.2.md`)                 | Nach `0.1.1`-Release                                            | MVP-10, MVP-15; F-89..F-94                                |
| 12  | ✅      | `docs/planning/plan-0.2.0.md` anlegen und `0.2.0`-Scope in umsetzbare Tranchen schneiden                                       | Nach `0.1.2`-Release                                            | RAK-11..RAK-21                                            |
| 13  | ✅      | Player-SDK-Paketierung und Public API stabilisieren                                                                   | Nach Schritt 12                                                 | RAK-11, RAK-12                                            |
| 14  | ✅      | Event-Schema-Versionierung und SDK↔Schema-Kompatibilitätscheck in CI planen                                           | Nach Schritt 12                                                 | RAK-13, RAK-21                                            |
| 15  | ✅      | hls.js-Adapter, HTTP-Transport sowie Batching/Sampling/Retry-Grenzen testbar absichern                                | Nach Schritt 12                                                 | RAK-14, RAK-15, RAK-17                                    |
| 16  | ✅      | OTel-Transport-Option bewerten und Performance-Budget nachweisen                                                      | Nach Schritt 15                                                 | RAK-16, RAK-18                                            |
| 17  | ✅      | Browser-Support-Matrix und Demo-Integrationsdoku erstellen                                                            | Nach Schritt 16                                                 | RAK-19, RAK-20                                            |
| 18  | ✅      | OE-3-Folge-ADR für Persistenz vorbereiten                                                                             | Parallel zu `0.2.0`-Planung                                     | OE-3; MVP-16                                              |
| 19  | ✅      | `docs/planning/plan-0.3.0.md` anlegen und `0.3.0`-Scope in umsetzbare Tranchen schneiden                                       | Nach `0.2.0`-Release                                            | RAK-22..RAK-28                                            |
| 20  | ✅      | Stream-Analyzer-Paket `packages/stream-analyzer` anlegen                                                              | Nach Schritt 19                                                 | RAK-22..RAK-26; MVP-33                                    |
| 21  | ✅      | HLS-Manifest laden und Master-/Media-Playlist-Erkennung umsetzen                                                      | Nach Schritt 20                                                 | RAK-22, RAK-23, RAK-24                                    |
| 22  | ✅      | Segment-Dauern prüfen und JSON-Ergebnisformat stabilisieren                                                           | Nach Schritt 21                                                 | RAK-25, RAK-26                                            |
| 23  | ✅      | API-Anbindung über bestehenden StreamAnalyzer-Port umsetzen                                                           | Nach Schritt 22                                                 | RAK-27; F-22, F-33                                        |
| 24  | ✅      | CLI-Grundlage für den Stream Analyzer schaffen                                                                        | Nach Schritt 22                                                 | RAK-28; MVP-34                                            |
| 25  | 🟡      | OE-3/Persistenz nach ADR-Draft neu bewerten — `0.3.0` ist stateless ausgeliefert; Entscheidung bleibt für `0.4.0`-Scope offen | Vor `0.4.0`-Scope-Cut                                          | OE-3; MVP-16                                              |

---

## 3. Release-Übersicht

Statusspalte: ✅ abgeschlossen · 🟡 in Arbeit · ⬜ geplant.

| Version | Titel                        | Status | Akzeptanzkriterien                                                                              |
| ------- | ---------------------------- | ------ | ----------------------------------------------------------------------------------------------- |
| `0.0.x` | Spike + Planungsphase        | ✅      | —                                                                                               |
| `0.1.0` | Backend Core + Demo-Lab      | ✅      | RAK-1, RAK-3, RAK-4, RAK-6, RAK-8 (initial); DoD-Tracking in [`plan-0.1.0.md`](./plan-0.1.0.md) |
| `0.1.1` | Player-SDK + Dashboard       | ✅      | RAK-2, RAK-5, RAK-7; DoD-Tracking in [`plan-0.1.1.md`](./plan-0.1.1.md)                         |
| `0.1.2` | Observability-Stack          | ✅      | RAK-9, RAK-10; DoD-Tracking in [`plan-0.1.2.md`](./plan-0.1.2.md)                               |
| `0.2.0` | Publizierbares Player SDK    | ✅      | RAK-11..RAK-21                                                                                  |
| `0.3.0` | Stream Analyzer              | ✅      | RAK-22..RAK-28; DoD-Tracking in [`plan-0.3.0.md`](./plan-0.3.0.md)                              |
| `0.4.0` | Erweiterte Trace-Korrelation | ⬜      | RAK-29..RAK-35                                                                                  |
| `0.5.0` | Multi-Protocol Lab           | ⬜      | RAK-36..RAK-40                                                                                  |
| `0.6.0` | SRT Health View              | ⬜      | RAK-41..RAK-46                                                                                  |

`0.1.x` ist seit Lastenheft-Patch `1.1.0` in drei Sub-Releases
geschnitten (Variante 2-A); RAK-1..RAK-10 sind dort verteilt.

DoD für die erste Phase ist über **AK-1..AK-11** abgedeckt
(Lastenheft-übergreifend, nicht Release-spezifisch). Detaillierter
Lieferstand pro Tranche steht in den drei `0.1.x`-Plan-Dokumenten;
Release-Vorgehen in [`docs/user/releasing.md`](../user/releasing.md).

---

## 4. Folge-ADRs

Aus `docs/adr/0001-backend-stack.md` §8 erwartete Folge-ADRs.
Alle sind ⬜ geplant; ADR-Nummer wird beim Schreiben vergeben. Die
zugehörigen Risiken stehen in `docs/planning/risks-backlog.md`.

| Erwartete ADR                                                 | Trigger-Release                                                | Begründung                                                                                                                                                                                                                                                                                                                                                                                             |
| ------------------------------------------------------------- | -------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Persistenz-Wechsel In-Memory → SQLite/PostgreSQL (**MVP-16**) | `0.1.0`–`0.2.0`                                                | Draft: [`docs/adr/0002-persistence-store.md`](../adr/0002-persistence-store.md). Spike-In-Memory ist nicht ausreichend, sobald Sessions persistiert werden sollen.                                                                                                                                                                                                                                      |
| WebSocket vs. SSE für Live-Updates                            | `0.4.0`                                                        | Live-Update-Mechanismus für Trace/Session-Ansicht.                                                                                                                                                                                                                                                                                                                                                     |
| SRT-Binding-Stack                                             | `0.6.0`                                                        | CGO-Bindings könnten das distroless-static-Pattern brechen.                                                                                                                                                                                                                                                                                                                                            |
| `apps/api` Multi-Modul-Aufteilung (`go.work`)                 | offen                                                          | Wird nur relevant, wenn Hexagon-Boundaries Disziplin-basiert nicht reichen.                                                                                                                                                                                                                                                                                                                            |
| Strengere CORS-Preflight-Project-Isolation (Variante A)       | offen, Trigger Multi-Tenant                                    | `0.1.0` setzt Variante B (globale Preflight-Allowlist + Project↔Origin-Validierung beim POST). Wenn echte Multi-Tenant-Projektion oder strengere Preflight-Isolation gebraucht wird, Migration auf Variante A — Project im Pfad (`/api/projects/{project_id}/...`) oder als URL-Parameter, damit der Preflight bereits projektscharf prüfen kann.                                                      |
| Durabel-konsistente Cursor-Strategie für Pagination           | offen, Trigger Horizontalskalierung oder Blue/Green-Deployment | `0.1.0` nutzt `process_instance_id` im Cursor; Restart bzw. Cross-Instance-Routing invalidiert den Cursor. Mit Multi-Instance-Setup wird das nicht mehr tragbar — Folge-ADR muss eine durabel-stabile Cursor-Form (z. B. opaker Token mit Storage-Token-ID, durable Sequence-Generator, server-side Snapshot) festlegen. Voraussetzung: Persistenz-Folge-ADR (OE-3) muss die Storage-Garantien klären. |

Neue Folge-ADRs werden hier ergänzt, sobald der Bedarf entsteht oder
ein Issue darauf hinweist.

---

## 5. Offene Entscheidungen

Verbleibende Lastenheft-`OE-X`; aufgelöste Einträge sind nach §7-Wartungsregel entfernt.

| Kennung | Entscheidung                                                                     | Wo wird sie getroffen             | Status                                                      |
| ------- | -------------------------------------------------------------------------------- | --------------------------------- | ----------------------------------------------------------- |
| OE-3    | Datenhaltung im MVP (In-Memory vs. SQLite/PostgreSQL) — verknüpft mit **MVP-16** | erste Folge-ADR (`0.1.0`–`0.2.0`) | in Arbeit — ADR-Draft [`0002`](../adr/0002-persistence-store.md) |
| OE-5    | Live-Updates: Polling / WebSocket / SSE                                          | Folge-ADR `0.4.0`                 | offen                                                       |

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
  `docs/planning/plan-spike.md` §14.11 wird beibehalten.
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

### 7.1 Source-of-Truth-Konvention bei Lastenheft-Widersprüchen

Lastenheft ist die normative Anforderungsquelle. Bei **interner**
Inkonsistenz zwischen einer F-Kennung (Anforderungs-Detail in §7) und
einer MVP-Kennung (Release-Prioritäts-Klassifikation in §12) gewinnt
**keine** Seite automatisch:

1. Plan-Dokumente (`plan-X.Y.Z.md`) markieren betroffene DoD-Items mit
   Status `[!]` (statt `[ ]` oder `[x]`) und beschreiben die
   Inkonsistenz in einem kurzen Hinweis.
2. Auflösung erfolgt durch einen **Lastenheft-Patch**: betroffene
   F- oder MVP-Kennung wird angepasst, Lastenheft-Header-Version
   bekommt einen Patch-Level-Bump (`1.0.0` → `1.0.1` → `1.0.2` …).
3. Der Patch wird im jeweiligen Plan-Dokument unter der dortigen
   Tranche „Lastenheft-Patches" (z. B. `plan-0.1.0.md` Tranche 0c)
   getrackt — mit Verweis auf die geänderten F-/MVP-Kennungen und
   den Begründungs-Pfad (Code-Review-Finding, ADR, Diskussion).
4. Bezug-Listen in den Soll-Dokumenten (`architecture.md`,
   `plan-X.Y.Z.md`, `README.md`) werden auf die neue Patch-Version
   gepinnt; historische Verweise (frühere Plan-Stände, ADRs,
   Spike-Doku) bleiben auf der ursprünglichen Version.

Diese Konvention verhindert, dass der Plan eigenmächtig zugunsten
einer der widersprüchlichen Quellen entscheidet und damit eine
normative Anforderung des Lastenhefts unterläuft.
