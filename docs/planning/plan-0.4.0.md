# Implementation Plan — `0.4.0` (Erweiterte Trace-Korrelation)

> **Status**: 🟡 in Arbeit. Tranche 0 abgeschlossen; Tranche 1 in §2.1–§2.5 ausgeliefert, §2.6 offen.
> **Bezug**: [Lastenheft `1.1.8`](../../spec/lastenheft.md) §13.6 (RAK-29..RAK-35), §7.9, §7.10, §7.11; [Roadmap](./roadmap.md) §1.2/§3/§4/§5; [Architektur](../../spec/architecture.md); [Telemetry-Model](../../spec/telemetry-model.md); [API-Kontrakt](../../spec/backend-api-contract.md); [ADR 0002 Persistenz-Store](../adr/0002-persistence-store.md); [ADR 0003 Live-Updates](../adr/0003-live-updates.md); [Risiken-Backlog](./risks-backlog.md).
> **Vorgänger-Gate (Stand zum `0.4.0`-Start)**:
>
> - [`plan-0.3.0.md`](./plan-0.3.0.md) ist vollständig (`[x]`) und `v0.3.0` ist veröffentlicht.
> - GitHub Actions `Build` ist für den Release-Commit `v0.3.0` grün.
> - ADR 0002 ist `Accepted`: SQLite ist der lokale Durable-Store für Sessions, Playback-Events und Ingest-Sequenzen.
> - OE-5 ist durch [ADR 0003](../adr/0003-live-updates.md) entschieden:
>   Dashboard-Live-Updates nutzen SSE mit Polling-Fallback; WebSocket ist
>   nicht Teil von `0.4.0`.
>
> **Nachfolger**: `plan-0.5.0.md` (Multi-Protocol Lab).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog [`plan-0.1.0.md`](./plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz oder offene Entscheidung.
- 🟡 in Arbeit.

Neue Lastenheft-Patches während `0.4.0` landen weiterhin zentral in `plan-0.1.0.md` Tranche 0c, weil sie projektweit gelten.

---

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
|---|---|---|
| 0 | Vorgänger-Gate und Scope-Entscheidungen | ✅ |
| 1 | SQLite-Persistenz und durable Cursor (siehe §2.1–§2.6) | 🟡 (§2.1–§2.5 ✅, §2.6 ⬜) |
| 2 | Session-Trace-Modell und OTel-Korrelation | ⬜ |
| 3 | Manifest-/Segment-/Player-Korrelation | ⬜ |
| 4 | Dashboard-Session-Verlauf ohne Tempo | ⬜ |
| 5 | Optionales Tempo-Profil | ⬜ |
| 6 | Aggregat-Metriken und Drop-/Invalid-/Rate-Limit-Sichtbarkeit | ⬜ |
| 7 | Cardinality- und Sampling-Dokumentation | ⬜ |
| 8 | Release-Akzeptanzkriterien `0.4.0` | ⬜ |

---

## 1a. Tranche 0 — Vorgänger-Gate und Scope-Entscheidungen

Bezug: Roadmap §1.2, §4, §5; ADR 0002; R-3; OE-5.

Ziel: Vor Implementierung ist klar, welche Entscheidungen `0.4.0` wirklich blockieren und welche bewusst als optionaler oder späterer Scope behandelt werden.

DoD:

- [x] `plan-0.3.0.md` ist vollständig (`[x]`), inklusive Release-Akzeptanzkriterien (Tranche 8 grün; verbleibende `[ ]`-Items in §9.1 sind non-blocking Folge-Issues, siehe Item unten).
- [x] Annotierter Release-Tag `v0.3.0` existiert und zeigt auf den finalen Release-Stand.
- [x] GitHub Actions `Build` ist für den Release-Commit grün (per `plan-0.3.0.md` §10 verifiziert).
- [x] `docs/planning/roadmap.md` führt `0.4.0` als aktiv geplantes Release und verweist auf dieses Dokument (Roadmap §3 + Schritt 27 ✅).
- [x] OE-5 ist entschieden: SSE mit Polling-Fallback ist für `0.4.0` gewählt; WebSocket bleibt deferred (ADR 0003).
- [x] Folge-ADR „Live-Updates via SSE" ist geschrieben und accepted (ADR 0003).
- [x] Folge-ADR „Dauerhaft konsistente Cursor-Strategie" ist geschrieben (ADR 0004, `1028688`).
- [x] Offene Folge-Issues aus `plan-0.3.0.md` §9.1 sind bewertet: beide non-blocking für `0.4.0`, deferred zu `0.3.x`-Fix — Contract-Test-Vollständigkeitsgrenze (kein Verhaltens-Bruch, nur Test-Härtung) und CI-Workflow-Bin-Symlink-Refactor (heute durch `make smoke-cli` workaround abgedeckt).
- [x] RAK-31 ist als optionaler Kann-Scope bestätigt: Tempo darf `0.4.0` nicht blockieren, solange RAK-29 und RAK-32 ohne Tempo erfüllt sind (Tranche 5 in §6).

---

## 2. Tranche 1 — SQLite-Persistenz und durable Cursor

Bezug: ADR 0002 §7/§8; RAK-32; F-18, F-30, F-38; MVP-14, MVP-16.

Ziel: Sessions, Playback-Events und Ingest-Sequenzen überleben API-Restarts. Die Dashboard-Session-Ansicht liest aus m-trace selbst und ist nicht von Tempo abhängig.

Tranche 1 ist in sechs aufeinander aufbauende Sub-Tranchen geschnitten: §2.1 fixiert die Spec-Grundlagen vor jedem Code, §2.2 liefert Schema und Migrationen, §2.3 die SQLite-Adapter, §2.4 das Wiring im Compose-Lab, §2.5 die Cursor-Migration im Code und §2.6 den Doku- und Test-Closeout. Die ursprüngliche, flache DoD-Liste aus früheren Plan-Ständen ist unverändert auf §2.1–§2.6 verteilt; pro Sub-Tranche werden nur die Items gelistet, die dort tatsächlich abgeschlossen werden.

### 2.1 Spec-Vorarbeit (Doku-only, kein Code)

Bezug: ADR 0002 §8; Folge-ADR „Dauerhaft konsistente Cursor-Strategie" (Roadmap §4); RAK-32; F-30; F-38.

Ziel: Vor jeder Codeänderung sind Cursor-Format, Schema-Skizze, Migrations-Tool-Wahl, Idempotenz-Regeln und kanonische Sortierung verbindlich entschieden, damit §2.2–§2.5 ohne implizite Spec-Entscheidungen umgesetzt werden können. Sub-Tranchen-Ausgang: keine Code-Diffs, aber alle nachfolgenden Sub-Tranchen können auf eindeutige Spec-Aussagen verweisen.

DoD:

- [x] Folge-ADR „Dauerhaft konsistente Cursor-Strategie" (`docs/adr/0004-cursor-strategy.md`) ist geschrieben und `Accepted`: definiert `cursor_version`, durable Token-Form (Storage-Position-Token mit `v`-Feld) und Recovery-Verhalten nach API-Restart (`1028688`).
- [x] Cursor-Kompatibilitätsmatrix ist in `spec/backend-api-contract.md` §10.3 festgeschrieben: `cursor_version`, erkannte Legacy-Formate (`process_instance_id`), Verhalten je Version (`accepted`, `cursor_invalid_legacy`, `cursor_invalid_malformed`, `cursor_expired`), HTTP-Status, Body-Schema und Client-Recovery sind eindeutig (`1028688`).
- [x] ADR 0002 §8 schließt die offenen Punkte: Tabellen-Skizze (`projects`, `stream_sessions`, `playback_events`, `schema_migrations`) mit globalem `ingest_sequence` und Migrations-Tool-Wahl (d-migrate für Schema-YAML und DDL-Generierung, eigener Go-Apply-Runner zur Laufzeit) ist verbindlich entschieden (`1028688`).
- [x] Idempotenz-Grenzen sind in ADR 0002 §8.3 und `spec/backend-api-contract.md` §10.2 festgelegt: Session-State-Updates sind idempotent; Event-Level-Dedup über Timeline-Klassifikation (`accepted`, `duplicate_suspected`, `replayed`) mit `(project_id, session_id, sequence_number)` als Dedup-Key und Dashboard-Anzeige (`1028688`).
- [x] Kanonische API-Event-Sortierung ist in `spec/backend-api-contract.md` §10.4 festgeschrieben: `server_received_at asc`, `sequence_number asc` (falls vorhanden), `ingest_sequence asc` als verpflichtender, durabler Tie-Breaker; `ingest_sequence` ist global eindeutig und monoton (`1028688`).
- [x] Retention-Defaults sind als „unlimited mit dokumentiertem Reset-Pfad" in ADR 0002 §8.4 und `spec/backend-api-contract.md` §10.5 verankert; Implementierungs- und Nutzerdoku folgen in §2.6 (`1028688`).

### 2.2 Schema und Migrationen

Bezug: §2.1; ADR 0002.

Ziel: Das in §2.1 entschiedene Schema existiert als versioniertes SQL und läuft beim API-Start deterministisch und idempotent. Sub-Tranchen-Ausgang: leeres Schema startet sauber, bestehender Schema-State bleibt bei Re-Run unverändert, Migrationsfehler hinterlässt erkennbaren State.

DoD:

- [x] SQLite-Schema für Projekte, Sessions, Playback-Events und Ingest-Sequenzen ist als versionierte Migration implementiert; Schema-Version ist getrennt vom Event-Wire-Schema versioniert. Schema-YAML in `apps/api/internal/storage/schema.yaml`, generierter SQLite-DDL im Flyway-Format als `migrations/V1__m_trace.sql` (`137a838`, `4db4a79`).
- [x] Migrationsmechanismus läuft beim lokalen API-Start deterministisch und idempotent; mehrfache Starts gegen denselben SQLite-State sind no-op (`4db4a79`).
- [x] Migrationsfehler-Pfad ist im Code abgefangen: Apply-Runner persistiert `dirty=1` in `schema_migrations` und weigert den Re-Start mit `ErrSchemaDirty`; Reparatur-Doku folgt in §2.6 (`4db4a79`).
- [x] Schema-/Migrationstests decken Frischstart, Re-Run gegen bestehenden State und simulierten Migrationsfehler ab (`TestOpen_FreshStart`, `TestOpen_ReRunIsNoop`, `TestApply_FailureMarksDirty`, `TestApply_DirtyStateRefuses` in `4db4a79`).

### 2.3 SQLite-Adapter

Bezug: §2.1; §2.2; ADR 0002.

Ziel: Drei Driven-Adapter hinter den bestehenden Ports machen Sessions, Playback-Events und Ingest-Sequenzen restart-stabil. Application- und Domain-Layer bleiben SQLite-frei. Sub-Tranchen-Ausgang: Adapter-Contract-Tests laufen identisch gegen In-Memory- und SQLite-Implementierung.

DoD:

- [x] Driven-Adapter für `SessionRepository`, `EventRepository` und `IngestSequencer` sind in `apps/api/adapters/driven/persistence/sqlite/` als SQLite-Implementierung umgesetzt; Application- und Domain-Layer importieren keine SQLite-Pakete (Sub-Paket-Refactor: `inmemory/`, `sqlite/`, `contract/`) (`11f6d85`).
- [x] Idempotenz aus §2.1 ist im Adapter implementiert: Session-State-Updates sind idempotent (zweimaliges `session_ended` ändert `ended_at` nicht); Event-Dedup via Timeline-Klassifikation (`accepted` / `duplicate_suspected`) auf Basis `(project_id, session_id, sequence_number)` über `BEGIN IMMEDIATE`-Serialisierung (`11f6d85`).
- [x] Kanonische Event-Sortierung aus §2.1 ist im Adapter durchgesetzt (COALESCE-basierter Filter mit `nullSeqSentinel`); `ingest_sequence` ist global eindeutig und durable persistiert (`11f6d85`).
- [x] In-Memory-Adapter bleiben für Tests und expliziten Dev-Fallback erhalten; Compose-Lab-Default-Wechsel selbst erfolgt in §2.4 (`11f6d85`).
- [x] Adapter-Contract-Tests laufen gegen In-Memory- und SQLite-Adapter über eine gemeinsame Suite (`persistence/contract`); Neustart-Simulation und Cursor-Stabilität sind in SQLite-spezifischen Restart-Tests abgedeckt (`11f6d85`).

### 2.4 Wiring und Compose

Bezug: §2.3; ADR 0002.

Ziel: API-Bootstrap wählt SQLite per Default im Compose-Lab; die Datei überlebt Container-Neustart und ist getrennt vom expliziten Reset-Pfad. Sub-Tranchen-Ausgang: `make stop` + erneuter Start zeigt vorherige Sessions weiter; Reset ist nur über einen dedizierten Pfad möglich.

DoD:

- [x] `apps/api/cmd/api/main.go` wählt den Persistenz-Adapter über `MTRACE_PERSISTENCE` (Default `sqlite` ab `0.4.0`, `inmemory` opt-in für Tests/Dev). Wahl-Logik in `newPersistence()` (`722f0ef`).
- [x] `MTRACE_SQLITE_PATH` setzt den expliziten SQLite-Pfad für lokale Entwicklung und CI; Default `/var/lib/mtrace/m-trace.db` matcht den Compose-Volume-Mountpoint (`722f0ef`).
- [x] SQLite-Datei liegt im Compose-Lab im benannten Volume `mtrace-data` des `api`-Service (`/var/lib/mtrace/m-trace.db`); `make stop` (`docker compose down`) entfernt das Volume nicht. `make wipe` ist als getrennter, dokumentierter Reset-Pfad eingeführt (`docker compose down --volumes`) (`722f0ef`).
- [x] Compose-Lab startet per Default mit SQLite-Adapter (`MTRACE_PERSISTENCE: sqlite` im `api`-Service); In-Memory ist nicht mehr Compose-Default — nur über Override `MTRACE_PERSISTENCE=inmemory` aktivierbar (`722f0ef`).

### 2.5 Cursor-Format im Code

Bezug: §2.1 Cursor-Kompatibilitätsmatrix; §2.3 SQLite-Adapter.

Ziel: Cursor-Format auf `cursor_version` umgestellt; Legacy-Verhalten entspricht der Matrix; kein `process_instance_id` mehr im Token-Inhalt. Sub-Tranchen-Ausgang: Cursor-Tests decken alle Matrix-Fälle ab und ein nach Restart fortgesetzter Cursor liefert keinen Datenverlust gegenüber In-Memory-Verhalten von `0.3.0`.

DoD:

- [x] `apps/api/adapters/driving/http/cursor.go` ist auf `cursor_version: 2` umgestellt (Pflicht-`v`-Feld, kein `pid` mehr); JSON-Decode mit `DisallowUnknownFields` lehnt Zusatzfelder als `cursor_invalid_malformed` ab; `Retry-After`-Header wird in keiner Fehlerklasse gesetzt (`1e41b85`).
- [x] Legacy-Detection: Cursor mit `pid`-Feld oder ohne `v`/`v:1` werden dauerhaft als `errCursorInvalidLegacy` abgewiesen — kein One-Shot-Grace-Pfad. `domain.ProcessInstanceID` und `domain.ErrCursorInvalid` sind aus dem Code entfernt; Application-Layer (SessionsService) trägt keine Prozess-ID mehr (`1e41b85`).
- [x] Recovery-Verhalten ist im Body-`reason`-Feld dokumentiert (`reload snapshot`); kein `Retry-After`. Vertrag steht in API-Kontrakt §10.3 (`1e41b85`).
- [x] Cursor-Tests decken alle Decode-Stufen: Round-Trip, Empty, Malformed (Base64-/JSON-/`v`-/Pflichtfelder/Extra-Felder), Legacy (PID, fehlendes `v`, `v:1`); Matrix-Klassen `accepted`, `cursor_invalid_legacy`, `cursor_invalid_malformed` sind abgedeckt. `cursor_expired` ist als Klasse spezifiziert; Code-Pfad existiert (`writeCursorError` mappt auf 410 Gone), aber bleibt in `0.4.0` ohne TTL nicht durch decode-Pfade triggerbar — Restart-stabile Cursor-Fortsetzung ist über die SQLite-Restart-Tests in §2.3 (`TestRestartCursorStability`) abgedeckt (`1e41b85`).
- [x] `cursor_test.go` und `sessions_handlers_test.go` sind auf die feiner aufgelösten Fehlerklassen umgestellt; alle `cursor_invalid`/`storage_restart`-Erwartungen entfernt (`1e41b85`).

### 2.6 Doku und Persistenztest-Closeout

Bezug: §2.1–§2.5.

Ziel: Spec-, Nutzer- und Architektur-Doku spiegeln den ausgelieferten Stand; Test-Suite ist über alle DoD-Aspekte hinweg grün. Sub-Tranchen-Ausgang: Roadmap-Schritt 28 ist auf ✅ aktualisierbar.

DoD:

- [ ] `spec/architecture.md` beschreibt den Storage-Stand (SQLite-Adapter, Volume, Retention) konsistent mit dem ausgelieferten Code.
- [ ] `spec/backend-api-contract.md` ist final konsistent mit dem Code (Cursor-Matrix, Sortier-Reihenfolge, Idempotenz-Regeln).
- [ ] `docs/user/local-development.md` beschreibt SQLite-Pfad, Volume-Reset/Wipe-Anleitung, Retention-Defaults und Recovery-Verhalten bei Cursor-Fehlern.
- [ ] Persistenztest-Suite deckt zusammenführend ab: Neustart-Simulation, Migration (Frischstart, Re-Run, Fehler), Cursor-Stabilität (alle Matrix-Fälle), Session-Ende-Idempotenz, Event-Ordering inkl. Tie-Breaker, Retention.
- [ ] Coverage-Strategie für SQL-Pakete ist entschieden: aktuell außerhalb des 90 %-Gates sind `apps/api/internal/storage/`, `apps/api/adapters/driven/persistence/sqlite/` und `apps/api/adapters/driven/persistence/contract/` (defensive SQL-/FS-Error-Pfade ohne Mocks nicht erreichbar, siehe `apps/api/Dockerfile` Coverage-Stage); §2.6 entscheidet, ob ein dediziertes Storage-Coverage-Setup mit niedrigerer Threshold lohnt oder Status quo bleibt.
- [ ] Multi-Tenant-Sichtbarkeit für `mtrace_active_sessions` ist entschieden: aktueller Gauge zählt projekt-übergreifend (`SELECT COUNT(*) FROM stream_sessions WHERE state = ?`); ein Per-Project-Label würde Cardinality erhöhen und braucht Allowlist-Schutz (siehe API-Kontrakt §7). Folge-Tranche entscheidet, ob Per-Project-Aufschlüsselung lohnt oder der Gauge global bleibt.
- [ ] Roadmap §2 Schritt 28 ist auf ✅ aktualisiert, sobald §2.1–§2.6 alle `[x]` sind.

---

## 3. Tranche 2 — Session-Trace-Modell und OTel-Korrelation

Bezug: RAK-29; RAK-35; Lastenheft §7.10/§7.11; Telemetry-Model §2/§3/§5; API-Kontrakt §8.

Ziel: Player-Sessions werden konsistent als Trace-Konzept modelliert. OTel-Spans und gespeicherte Events teilen stabile Korrelations-IDs, ohne Prometheus-Cardinality-Regeln zu verletzen.

DoD:

- [ ] Trace-ID-Strategie ist festgelegt: pro Player-Session existiert eine stabile Korrelation, die Backend-Spans und gespeicherte Events verbinden kann.
- [ ] RAK-29/RAK-32 sind Tempo-unabhängig erfüllbar: die lokale Persistenz speichert `trace_id` oder eine äquivalente `correlation_id` als Source of Truth; Tempo ist nur optionaler Export/Viewer und darf kein Pflichtpfad für Dashboard-Korrelation sein.
- [ ] `session_id` bleibt pseudonym und wird nicht als Prometheus-Label verwendet.
- [ ] HTTP-Request-Spans für `POST /api/playback-events` tragen kontrollierte Attribute für Project, Batch-Outcome, Event-Anzahl und bei Erfolg Session-Korrelationsdaten.
- [ ] Event-Persistenz speichert Trace-/Span-Kontext oder eine daraus abgeleitete Korrelations-ID so, dass die Dashboard-Ansicht ohne Tempo nutzbar bleibt.
- [ ] Player-SDK-Transport propagiert optionalen Trace-Kontext oder sendet die nötigen Korrelationsfelder ohne Breaking Change im Event-Wire-Format.
- [ ] Server validiert eingehende Korrelationsfelder defensiv; ungültige Trace-Kontexte führen nicht zum Absturz und werden dokumentiert behandelt.
- [ ] Time-Skew-Handling aus `spec/telemetry-model.md` §5.3 ist umgesetzt oder als explizit späterer Scope dokumentiert.
- [ ] Tests decken Trace-Konsistenz über mehrere Batches einer Session, fehlenden Client-Kontext, ungültigen Kontext und Session-Ende ab.
- [ ] Tests verifizieren Trace-/Korrelationskonsistenz bei deaktiviertem Tempo-Profil; dieselben Tests dürfen nicht von einem externen Trace-Backend abhängen.
- [ ] `spec/telemetry-model.md` dokumentiert die konkrete Span-Struktur, Attribute und Sampling-Auswirkung für `0.4.0`.

---

## 4. Tranche 3 — Manifest-/Segment-/Player-Korrelation

Bezug: RAK-30; RAK-29; Stream Analyzer aus `0.3.0`; F-68..F-81; Telemetry-Model §1.

Ziel: Manifest-Requests, Segment-Requests und Player-Events werden soweit technisch möglich einem gemeinsamen Session-Trace zugeordnet. RAK-30 ist Soll; Lücken müssen sichtbar und erklärbar bleiben.

DoD:

- [ ] Player-SDK erfasst Manifest- und Segment-nahe Ereignisse aus dem hls.js-Adapter, soweit hls.js sie zuverlässig liefert.
- [ ] Event-Schema erlaubt die Unterscheidung von Manifest-Request, Segment-Request und Player-Zustandsereignis ohne Breaking Change oder mit dokumentierter Schema-Migration.
- [ ] Segment- und Manifest-URLs werden nicht als Prometheus-Labels verwendet; Speicherung im Event-Store folgt den Datenschutz- und Retention-Regeln.
- [ ] Backend normalisiert die eingehenden Netzwerkereignisse in den bestehenden Session-/Event-Store.
- [ ] Manifest-, Segment- und Player-Events teilen denselben Trace- oder Korrelationskontext, wenn der Browser/SDK-Pfad die nötigen Signale liefert; Abweichungen werden pro Ereignistyp begründet.
- [ ] Falls einzelne Manifest-/Segment-Daten nur als Event-Timeline und nicht als OTel-Span abbildbar sind, ist diese Grenze explizit dokumentiert und im Dashboard sichtbar nachvollziehbar.
- [ ] Korrelation ist tolerant gegenüber fehlenden SDK-Feldern, blockierten Browser-Timings und CORS-/Resource-Timing-Lücken.
- [ ] Analyzer-Ergebnisse aus `POST /api/analyze` sind optional mit einer Session verknüpfbar oder bewusst getrennt dokumentiert, damit Manifestanalyse und Player-Timeline nicht inkonsistent vermischt werden.
- [ ] Tests decken gemischte Player-, Manifest- und Segment-Ereignisse innerhalb einer Session ab und prüfen den gemeinsamen Trace-/Korrelationskontext oder die dokumentierte Timeline-only-Ausnahme.
- [ ] Dokumentation benennt Grenzen der Korrelation, insbesondere Browser-APIs, CORS, Service Worker, CDN-Redirects und Sampling.

---

## 5. Tranche 4 — Dashboard-Session-Verlauf ohne Tempo

Bezug: RAK-32; MVP-14; F-38..F-40; ADR 0002.

Ziel: Das Dashboard zeigt Session-Verläufe aus der lokalen m-trace-Persistenz einfach, schnell und restart-stabil an. Tempo ist dafür nicht erforderlich.

DoD:

- [ ] Session-Liste und Session-Detailansicht lesen aus SQLite-backed API-Pfaden und zeigen Daten nach API-Restart weiter an.
- [ ] Detailansicht stellt eine Timeline aus Player-, Manifest- und Segment-Ereignissen dar, mit stabiler Reihenfolge und klarer Typ-Unterscheidung.
- [ ] Laufende Sessions sind von beendeten Sessions unterscheidbar; `session_ended` und Sweeper-Ende werden sichtbar.
- [ ] Invalid-, dropped- und rate-limited Hinweise sind in der Session- oder Statusansicht auffindbar, ohne Prometheus-Rohwissen vorauszusetzen.
- [ ] Duplikat- oder Replay-Klassifikationen aus der Persistenz sind in der Timeline unterscheidbar und beschädigen nicht die Default-Reihenfolge.
- [ ] Pagination oder inkrementelles Nachladen bleibt bei längeren Sessions bedienbar; Cursor-Verhalten ist restart-stabil.
- [ ] SSE-Live-Update-Mechanismus aus ADR 0003 ist implementiert; Polling bleibt Fallback für Stream-Abbruch oder nicht verfügbare SSE-Verbindung.
- [ ] SSE-Endpunkt-Schnittstelle ist im `spec/backend-api-contract.md` als verlässlicher Vertrag dokumentiert: globaler Stream, optionaler Session-Detail-Stream, Payload-Schema, `Last-Event-ID`-/Backfill-Regel, Fehler-/Reconnect-Semantik und Polling-Fallback-Intervalle.
- [ ] SSE-`id`/`Last-Event-ID` ist an ein dauerhaft persistiertes Event-Store-Feld gebunden, z. B. eine monotone Persistenz-ID oder `ingest_sequence`; Scope und Eindeutigkeit sind passend zum Stream-Typ definiert (globaler Stream braucht global eindeutige ID, Session-Stream mindestens session-eindeutige ID); Reconnect-Backfill liest ausschließlich aus SQLite und funktioniert nach API-Restart.
- [ ] SSE-Fallback-Grenzen sind hart definiert und getestet: Heartbeat-Intervall, Reconnect-Backoff, maximale Backfill-Lücke und Polling-Intervall haben konkrete Defaults sowie obere Grenzen im API-Kontrakt.
- [ ] Backend-Tests decken SSE-Stream-Header, EventSource-kompatibles Format, Heartbeats/Keepalive, Client-Abbruch und reconnect-freundliche Semantik ab.
- [ ] Dashboard-Tests decken SSE-Erfolg, Reconnect/Backfill und Polling-Fallback ab.
- [ ] Dashboard-Tests decken leere Timeline, kurze Session, lange Session, laufende Session und beendete Session über API-Mockdaten ab; Restart-Persistenz wird zusätzlich durch einen Integration-/E2E-Test mit echter SQLite-Datei und API-Neustart geprüft.
- [ ] Browser-E2E-Smoke erzeugt eine Session über einen stabilen Test-Harness (`/demo` oder dedizierte E2E-Seed-Route/API-Fixture) und prüft, dass der Session-Verlauf im Dashboard sichtbar ist; `/demo` ist nicht die einzige zulässige Datenquelle.

---

## 6. Tranche 5 — Optionales Tempo-Profil

Bezug: RAK-31; RAK-29; Architektur §2/§5; README `0.4.0`.

Ziel: Tempo kann als optionales Trace-Backend genutzt werden, ohne die lokale Dashboard-Ansicht zur Pflicht-Abhängigkeit zu machen.

DoD:

- [ ] Compose-Profil für Tempo ist optional und startet nur bei expliziter Aktivierung.
- [ ] OTel-Collector leitet Traces an Tempo weiter, wenn das Profil aktiv ist; ohne Profil bleibt der API-Start silent/no-op.
- [ ] Ohne Tempo-Profil bleiben lokale `trace_id`/`correlation_id`, Dashboard-Timeline und RAK-29-Tests vollständig funktionsfähig.
- [ ] Trace-Suche oder ein Link-Konzept ist dokumentiert, falls Dashboard und Tempo gemeinsam laufen.
- [ ] RAK-29 ist auch ohne Tempo erfüllt; Tempo erweitert nur Debug-Tiefe.
- [ ] Lokaler Smoke-Test oder manuelle Release-Checkliste beschreibt, wie ein Trace in Tempo sichtbar wird.
- [ ] README und `docs/user/local-development.md` unterscheiden klar zwischen eingebauter Session-Timeline und optionalem Tempo.

---

## 7. Tranche 6 — Aggregat-Metriken und Drop-/Invalid-/Rate-Limit-Sichtbarkeit

Bezug: RAK-33; RAK-34; API-Kontrakt §7; Telemetry-Model §2.4/§3/§4.3; Lastenheft §7.9/§7.10.

Ziel: Prometheus bleibt Aggregat-Backend. Die Pflichtmetriken für angenommene, invalid, rate-limited und dropped Events sind sichtbar, korrekt gezählt und cardinality-sicher.

DoD:

- [ ] `mtrace_playback_events_total`, `mtrace_invalid_events_total`, `mtrace_rate_limited_events_total` und `mtrace_dropped_events_total` existieren im Compose-Lab und in Tests.
- [ ] Alle Pflichtcounter zählen Events, nicht Batches; leere Batches, Auth-Fehler und Persistenzfehler folgen den Regeln aus API-Kontrakt §7.
- [ ] Es gibt keinen `session_id`-, `user_agent`-, `segment_url`-, `client_ip`- oder unbounded-`project_id`-Label auf `mtrace_*`-Metriken.
- [ ] Rate-Limit-Fälle sind mit `429` und Counter-Inkrement testbar.
- [ ] Invalid-Event-Fälle mit `400`/`422` sind mit Counter-Inkrement testbar.
- [ ] Drop-Pfad ist entweder real implementiert und testbar oder die Metrik existiert sichtbar mit `0` und der fehlende Drop-Pfad ist dokumentiert.
- [ ] Grafana-/Prometheus-Lab zeigt die vier Pflichtcounter oder eine dokumentierte Abfrage dafür.
- [ ] Cardinality-Smoke prüft, dass neue `0.4.0`-Metriken keine hochkardinalen Labels einführen.

---

## 8. Tranche 7 — Cardinality- und Sampling-Dokumentation

Bezug: RAK-35; RAK-33; RAK-34; Lastenheft §7.10/§7.11; Telemetry-Model §3/§4.4.

Ziel: Nutzer verstehen, welche Daten in Prometheus, OTel/Tempo und SQLite landen, welche Sampling-Strategie gilt und welche Grenzen für produktionsnahe Nutzung bestehen.

DoD:

- [ ] `spec/telemetry-model.md` beschreibt `0.4.0`-Sampling für SDK-Events, Backend-Spans und optionale Tempo-Nutzung.
- [ ] `docs/user/local-development.md` beschreibt lokale Storage-Retention, SQLite-Reset, Prometheus-Aggregate und optionales Tempo-Profil.
- [ ] `docs/user/demo-integration.md` zeigt, wie eine Demo-Session inklusive Timeline reproduzierbar erzeugt wird.
- [ ] `README.md` aktualisiert den `0.4.0`-Abschnitt mit tatsächlichem Lieferstand.
- [ ] Doku enthält eine klare Tabelle: Prometheus = Aggregate, SQLite = Session-/Event-Historie, OTel/Tempo = Trace-Debugging.
- [ ] Sampling-Grenzen erklären, wie unvollständige Timelines im Dashboard markiert werden.
- [ ] Datenschutz- und Cardinality-Hinweise nennen ausdrücklich `session_id`, URLs, User-Agent und Client-IP.
- [ ] Release-Notes-Vorlage im `CHANGELOG.md`-Unreleased-Abschnitt enthält die neuen Trace-, Storage-, Metrik- und Doku-Punkte.

---

## 9. Tranche 8 — Release-Akzeptanzkriterien `0.4.0`

Bezug: RAK-29..RAK-35; `docs/user/releasing.md`.

DoD:

- [ ] **RAK-29** Player-Session-Traces werden konsistent und Tempo-unabhängig erzeugt: mehrere Batches einer Session teilen lokal persistierte Korrelationsdaten; Tests decken Erfolg, fehlenden Kontext und deaktiviertes Tempo-Profil ab.
- [ ] **RAK-30** Manifest-Requests, Segment-Requests und Player-Events werden soweit technisch möglich in einem gemeinsamen Trace-/Korrelationskontext zusammengeführt; Timeline-only-Ausnahmen sind je Ereignistyp begründet und dokumentiert.
- [ ] **RAK-31** Tempo kann optional als Trace-Backend verwendet werden oder ist bewusst als Kann-Scope deferred, ohne Muss-Kriterien zu gefährden.
- [ ] **RAK-32** Dashboard kann Session-Verläufe ohne Tempo anzeigen; API-Restart verliert bestehende lokale Session-Historie nicht.
- [ ] **RAK-33** Prometheus bleibt auf aggregierte Metriken beschränkt; Cardinality-Smoke ist grün.
- [ ] **RAK-34** Dropped-, Rate-Limited- und Invalid-Event-Metriken sind sichtbar und testbar.
- [ ] **RAK-35** Dokumentation beschreibt Cardinality-Grenzen und Sampling-Strategie.
- [ ] Versionen sind konsistent: Root- und Workspace-Pakete tragen `0.4.0`; SDK/Event-Schema-Kompatibilitätscheck bleibt grün.
- [ ] `CHANGELOG.md` enthält den Versionsabschnitt `[0.4.0] - <Datum>` mit Trace-, Persistenz-, Dashboard-, Metrik- und Doku-Lieferstand.
- [ ] Release-Gates grün: `make test`, `make lint`, `make coverage-gate`, `make arch-check`, `make build`, `make sdk-performance-smoke`, `make smoke-observability` und Dashboard-Tests.
- [ ] Browser-E2E-Smoke für eine erzeugte Test-Session und Session-Timeline ist grün oder als manuelles Release-Gate mit Ergebnis dokumentiert; der Smoke darf `/demo` nutzen, muss aber bei späterer Demo-Änderung auf einen dedizierten Test-Harness umstellbar bleiben.
- [ ] `docs/planning/roadmap.md` markiert `0.4.0` als abgeschlossen und verschiebt den aktiven Fokus auf `0.5.0`.

---

## 10. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`, Commit-Hash anhängen.
- Neue Findings in der `0.4.0`-Phase landen entweder in dieser Datei oder in `risks-backlog.md`.
- Lastenheft-Patches während `0.4.0` werden in `plan-0.1.0.md` Tranche 0c als neue §4a.x-Einträge ergänzt.
- Beim Release-Bump `0.4.0` → `0.5.0`: dieses Dokument als historisch archivieren; Lieferstand wandert dokumentarisch nach `CHANGELOG.md`.
