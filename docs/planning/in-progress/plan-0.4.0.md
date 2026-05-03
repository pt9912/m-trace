# Implementation Plan — `0.4.0` (Erweiterte Trace-Korrelation)

> **Status**: 🟡 in Arbeit. Tranche 0, Tranche 1 (§2.1–§2.6) und Tranche 2 §3.1–§3.4b abgeschlossen; offen: Tranche 2 §3.4c (Doku-Closeout + Roadmap Schritt 31) sowie Tranchen 3–8.
> **Bezug**: [Lastenheft `1.1.8`](../../../spec/lastenheft.md) §13.6 (RAK-29..RAK-35), §7.9, §7.10, §7.11; [Roadmap](./roadmap.md) §1.2/§3/§4/§5; [Architektur](../../../spec/architecture.md); [Telemetry-Model](../../../spec/telemetry-model.md); [API-Kontrakt](../../../spec/backend-api-contract.md); [ADR 0002 Persistenz-Store](../../adr/0002-persistence-store.md); [ADR 0003 Live-Updates](../../adr/0003-live-updates.md); [Risiken-Backlog](../open/risks-backlog.md).
> **Vorgänger-Gate (Stand zum `0.4.0`-Start)**:
>
> - [`plan-0.3.0.md`](../done/plan-0.3.0.md) ist vollständig (`[x]`) und `v0.3.0` ist veröffentlicht.
> - GitHub Actions `Build` ist für den Release-Commit `v0.3.0` grün.
> - ADR 0002 ist `Accepted`: SQLite ist der lokale Durable-Store für Sessions, Playback-Events und Ingest-Sequenzen.
> - OE-5 ist durch [ADR 0003](../../adr/0003-live-updates.md) entschieden:
>   Dashboard-Live-Updates nutzen SSE mit Polling-Fallback; WebSocket ist
>   nicht Teil von `0.4.0`.
>
> **Nachfolger**: `plan-0.5.0.md` (Multi-Protocol Lab).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog [`plan-0.1.0.md`](../done/plan-0.1.0.md) §0:

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
| 1 | SQLite-Persistenz und durable Cursor (siehe §2.1–§2.6) | ✅ |
| 2 | Session-Trace-Modell und OTel-Korrelation (siehe §3.1–§3.4) | ⬜ |
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
- [x] `docs/planning/in-progress/roadmap.md` führt `0.4.0` als aktiv geplantes Release und verweist auf dieses Dokument (Roadmap §3 + Schritt 27 ✅).
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

- [x] `spec/architecture.md` beschreibt den Storage-Stand: §3.1-Mermaid und §3.4-Adapter-Tabelle reflektieren das Sub-Paket-Layout (`inmemory/`/`sqlite/`/`contract/`); §4.2-Tree zeigt `internal/storage` für den Apply-Runner. SQLite ist als Default ab `0.4.0` markiert; ADR-0002 §8.1 bleibt die normative Schema-Quelle (`9dbbc52`).
- [x] `spec/backend-api-contract.md` ist final konsistent mit dem Code: §3.7 (Server-Read-Felder), §10.1–§10.5 (Storage, Idempotenz, Cursor-Matrix v2, Sortierung, Retention) wurden in §2.1/§2.5 aktualisiert; veralteter Implementierungs-Status-Callout in §10.3 ist nach §2.5-Closeout entfernt.
- [x] `docs/user/local-development.md` §3.4 beschreibt SQLite-Pfad, env-var-Konfiguration, `make wipe` als einzigen Reset-Pfad, Cursor-Recovery-Mapping und Migrations-Verhalten (`dirty`-Flag) (`9dbbc52`).
- [x] Persistenztest-Suite ist auf die folgenden Test-Files verteilt:
  - `apps/api/internal/storage/migrate_test.go` — Frischstart, Re-Run-no-op, Migrations-Fehler-Pfad mit `dirty`-Flag, Refuse-to-Start, Multi-Statement-Rollback, Concurrent-Writers (sechs Tests).
  - `apps/api/adapters/driven/persistence/{inmemory,sqlite}/contract_test.go` plus `contract/contract.go` — gemeinsame Adapter-Suite (Event-Ordering, Cursor-Pagination, Session-Ended-Idempotenz, Sweep-Lifecycle inkl. Single-Sweep-Active→Ended, Sequencer-Monotonie, Session-List-Tie-Breaker, ended-as-first-event, Meta-Roundtrip mit verschachtelten Werten, CountByState).
  - `apps/api/adapters/driven/persistence/sqlite/restart_test.go` — Restart-Stabilität (Daten + Sequencer + Cursor) gegen echte SQLite-Datei.
  - `apps/api/adapters/driven/persistence/sqlite/dedup_test.go` — Klassifikation (`accepted`/`duplicate_suspected`) plus Concurrent-Writer-Race über `BEGIN IMMEDIATE`.
  - `apps/api/adapters/driving/http/cursor_test.go` und `cursor_error_mapping_test.go` — alle vier Matrix-Klassen (`accepted`, `cursor_invalid_legacy`, `cursor_invalid_malformed`, `cursor_expired`), Encode-/Decode-Stufen, HTTP-Status-Mapping. Retention selbst wird in `0.4.0` nicht automatisch ausgeführt (siehe API-Kontrakt §10.5); der `cursor_expired`-Mapping-Pfad ist über `cursor_error_mapping_test.go` gesichert.
- [x] Coverage-Strategie für SQL-Pakete: Status-quo bleibt — `apps/api/internal/storage/`, `apps/api/adapters/driven/persistence/sqlite/`, `persistence/contract/` (und proaktiv `postgres/`/`mysql/`) sind im Dockerfile-coverpkg-Filter ausgenommen, weil defensive Error-Pfade ohne SQL-/FS-Mocks unerreichbar sind. Restdeckung über Contract-Tests gegen echte SQLite-Dateien hält die Adapter-Logik abgesichert. Re-Evaluation, sobald ein zweites SQL-Backend (Postgres) hinzukommt.
- [x] Multi-Tenant-Sichtbarkeit für `mtrace_active_sessions`: Status-quo bleibt — Gauge zählt projekt-übergreifend. Per-Project-Aufschlüsselung wird mit dem Postgres-Folge-ADR (Multi-Instance/Multi-Tenant) gemeinsam entschieden, weil Cardinality-Erhöhung dort erst real wird; bis dahin reicht der globale Gauge.
- [x] Roadmap §2 Schritt 28 ist auf ✅ aktualisiert; Tranche 1 (§2.1–§2.6) abgeschlossen.

---

## 3. Tranche 2 — Session-Trace-Modell und OTel-Korrelation

Bezug: RAK-29; RAK-35; Lastenheft §7.10/§7.11; Telemetry-Model §2/§3/§5; API-Kontrakt §8.

Ziel: Player-Sessions werden konsistent als Trace-Konzept modelliert. OTel-Spans und gespeicherte Events teilen stabile Korrelations-IDs, ohne Prometheus-Cardinality-Regeln zu verletzen. Das Pflichtziel ist **Tempo-unabhängig**: Dashboard- und API-Read-Pfade müssen auch ohne aktives Trace-Backend eine Session über `correlation_id` rekonstruieren können. OTel/Tempo ist in dieser Tranche nur Debug-Vertiefung über `trace_id`, nicht Source-of-Truth.

Tranche 2 ist in vier aufeinander aufbauende Sub-Tranchen geschnitten: §3.1 fixiert die Spec-Grundlagen vor jedem Code, §3.2 baut die Server-Korrelation (Spans, Persistenz, Validierung), §3.3 die SDK-Wire-Format-Erweiterung, §3.4 die Tests und finalisiert die Doku.

Abnahmegrenzen für die gesamte Tranche:

- **Source-of-Truth:** `stream_sessions.correlation_id` ist der stabile Session-Zusammenhang; `playback_events.correlation_id` muss für ab §3.2 verarbeitete Events dazu passen. `trace_id` und `span_id` sind batch-/span-bezogen, werden nur auf Events persistiert und dürfen zwischen Batches derselben Session wechseln.
- **Legacy-Grenze:** Für vor §3.2 angelegte Sessions/Events gibt es in Tranche 2 kein historisches Event-Backfill. Self-Healing setzt beim nächsten Event einer Legacy-Session die Session-`correlation_id` und alle neu geschriebenen Event-`correlation_id`s; ältere `playback_events.correlation_id`-Leerwerte bleiben ein dokumentierter degradierter Read-Fall.
- **Wire-Kompatibilität:** Payload-Schema bleibt `1.0`; der optionale `traceparent`-Header ist additiv. SDKs ohne Header, SDKs mit gültigem Header und SDKs mit kaputtem Header müssen denselben Event-Annahme-Pfad behalten.
- **Observability-Grenze:** `trace_id`, `span_id`, `correlation_id`, `session_id`, URLs und User-Agent bleiben Prometheus-Label-tabu. Falls sie auftauchen, ist das ein Release-Blocker und muss vor Tranche 8 behoben werden.
- **Tempo-Unabhängigkeit:** Tests dürfen kein externes Tempo/OTLP-Backend benötigen. Tempo-Integration wird erst in Tranche 5 optional verdrahtet. Der Tranche-2-Claim ist erst vollständig erfüllt, wenn §3.4c zusätzlich den produktiven `cmd/api`-Config-Auflösungsweg ohne aktives Trace-Backend abdeckt; §3.4a allein ist nur Adapter-/Use-Case-Nachweis.
- **Rest-Risiko:** Der bekannte `correlation_id`-Race bei paralleler Erstanlage derselben `session_id` ist als R-6 im Risiken-Backlog geführt. Er blockiert Tranche 2 nicht, solange §3.4c die Doku-Grenze klar benennt, der Mitigationspfad im Backlog bleibt und Tranche 8 keine beobachtete Inkonsistenz findet. R-6 darf nicht aus dem Abschluss-Text verschwinden.

Liefer-/Abnahme-Matrix:

| Sub-Tranche | Ergebnis | Harte Abnahme | Status |
|---|---|---|---|
| §3.1 | Normativer Vertrag für Trace-ID, `correlation_id`, Span-Attribute und Defensive Parsing | Spec enthält keine impliziten Entscheidungen mehr für §3.2/§3.3 | ✅ |
| §3.2 | Server erzeugt/persistiert batchbezogene `trace_id`/`span_id` auf Events, Session-`correlation_id` auf Session und Events sowie Span-Attribute | Backend- und Adapter-Tests laufen ohne externes Trace-Backend; `mtrace.project.id` ist auf accepted Batches gesetzt | ✅ |
| §3.3 | SDK kann optional `traceparent` senden | Kein Payload-/Schema-Bruch; Provider-Fehler sabotieren `send` nicht | ✅ mit nachgezogenem Snapshot-/Spec-Fix |
| §3.4a | Backend-Trace-Konsistenz abgesichert | Multi-Batch, Missing/Invalid Header, Session-Ende, Time-Skew, NoOp-Tracer getestet | ✅ |
| §3.4b | SDK↔Server-Cross-Cutting-Lücken geschlossen | Pre-§3.2-Backend-Kompat und Garbage-Traceparent-Pfad getestet | ✅ |
| §3.4c | Doku-/Roadmap-Closeout | Spec-Drift geschlossen, Roadmap Schritt 31 ✅, Tranche 2 als abgeschlossen markierbar | ⬜ |

### 3.1 Spec-Vorarbeit (Doku-only, kein Code)

Bezug: RAK-29; RAK-35; Telemetry-Model §2/§3/§5; API-Kontrakt §3, §5, §8.

Ziel: Vor Code-Änderungen sind Trace-ID-Strategie, Span-Modell, `correlation_id`-Vertrag, SDK-Wire-Format-Erweiterung und Validierungsregeln verbindlich entschieden, sodass §3.2–§3.4 ohne implizite Spec-Entscheidungen umgesetzt werden können. Sub-Tranchen-Ausgang: keine Code-Diffs, aber alle nachfolgenden Sub-Tranchen können auf eindeutige Spec-Aussagen verweisen.

Verbindliche Entscheidungen (gehören in `spec/telemetry-model.md` und `spec/backend-api-contract.md`):

- **Trace-ID-Quelle: Hybrid.** Player-SDK propagiert optional einen W3C-`traceparent`-Header (Format laut [W3C Trace Context](https://www.w3.org/TR/trace-context/)). Ist der Header gültig, übernimmt der Server `trace_id` und `parent_span_id` aus dem Header und erzeugt einen Child-Span. Fehlt der Header oder ist er ungültig, generiert der Server einen Root-Span mit eigener W3C-konformer `trace_id`. SDK-Wert hat Vorrang gegenüber Server-Generierung.
- **`trace_id` ≠ `correlation_id`.** Beide sind getrennte Konzepte mit klarer Verantwortung:
  - `trace_id` (TEXT, nullable, 32 Hex-Zeichen): W3C-Trace-ID — vom SDK propagiert oder server-generiert; primär für Tempo (RAK-31, optional).
  - `correlation_id` (TEXT, ab §3.2 pro Session gesetzt; historische Leerwerte möglich): server-generierte, durable Source-of-Truth für die Dashboard-Korrelation. Wird beim allerersten Event einer neuen Session erzeugt (UUIDv4 oder vergleichbar), in `stream_sessions.correlation_id` persistiert und für **alle ab §3.2 verarbeiteten** Folge-Events derselben Session konstant gehalten. Legacy-Sessions ohne `correlation_id` werden beim nächsten Event per Self-Healing auf Session und neu persistierten Events nachgezogen; ältere Events werden nicht backfilled.
  - Dashboard-Timeline (RAK-32) nutzt `correlation_id` — Tempo-unabhängig. Tempo (RAK-31) nutzt `trace_id`, wenn das Profil aktiv ist.
- **Span-Modell: ein HTTP-Request-Span pro Batch.** Keine Child-Spans pro Event (Cardinality-Risiko). Pflicht-Attribute am Server-Span:
  - `mtrace.project.id` (kontrolliert; Allowlist aus dem Use-Case-Resolver) ist Pflicht für accepted Batches und für alle Pfade, in denen der Use Case ein Project erfolgreich aufgelöst hat; bei Rejects vor Project-Auflösung bleibt das Attribut bewusst unset.
  - `mtrace.batch.size` (int)
  - `mtrace.batch.outcome` (`accepted` / `invalid` / `rate_limited` / `auth_error` etc.)
  - **Bei Single-Session-Batch (alle Events teilen `session_id`)** zusätzlich `mtrace.session.correlation_id` (und nur dieser Wert). `session_id` selbst wird ab `0.4.0` in keinem Span-Attribut gesetzt; die frühere §2.1-Erlaubnis gilt nur für den historischen `0.1.x`-Kontext und ist für Tranche 2 durch `correlation_id` ersetzt.
  - `mtrace.batch.session_count` (int) — bei Multi-Session-Batches > 1; das einzelne Event trägt seine `correlation_id` aus der Persistenz, nicht der Span.
  - `mtrace.trace.parse_error=true` falls eingehender `traceparent` ungültig war (siehe Validierungsregel unten).
  - `mtrace.time.skew_warning=true` falls für mindestens ein Event im Batch `|client_timestamp - server_received_at| > 60s` (Schwelle aus `telemetry-model.md` §5.3); Persistenz des Skew-Flags auf Event-Ebene ist deferred (siehe unten).
- **Defensive Validierung des `traceparent`-Headers.** Ein ungültiger oder formal kaputter Header führt **nicht** zu 4xx und nicht zum Absturz. Stattdessen: Span-Attribut `mtrace.trace.parse_error=true`, Server-Fallback erzeugt eine eigene `trace_id`, Event wird normal verarbeitet. Die Pflicht-Validierungs-Reihenfolge aus API-Kontrakt §5 wird dadurch nicht verändert.
- **Header-Normalisierung.** Der Header-Name ist HTTP-konform case-insensitiv (`Traceparent`, `traceparent`, `TRACEPARENT` sind derselbe Header). Der Header-Wert ist ein einzelner W3C-`traceparent`-Wert; ob führende/abschließende OWS technisch getrimmt oder als Parse-Error behandelt wird, muss in §3.4c exakt mit Code und Tests synchronisiert werden.
- **Cardinality-Regel.** Weder `trace_id`, `correlation_id` noch `span_id` werden als Prometheus-Labels verwendet — Span-Attribute (kontrolliert), Event-Persistenz-Spalten (durable) und Wire-Format-Felder (optional) sind die einzigen Konsumenten.
- **Time-Skew-Handling: nur Span-Attribut in `0.4.0`.** Span-Attribut `mtrace.time.skew_warning=true` aus `telemetry-model.md` §5.3 ist Pflicht in §3.2. Persistenz-Spalte und Dashboard-Anzeige sind explizit deferred — ein Folge-Tranchen-Item wird sie ergänzen, sobald Bedarf entsteht.

DoD:

- [x] `spec/telemetry-model.md` §2.5 (neu) dokumentiert die Hybrid-Trace-ID-Strategie, das `trace_id`/`correlation_id`-Verhältnis (mit Persistenz-Quelle pro Feld), das Span-Modell mit Pflicht-Attribut-Tabelle und die Sampling-Auswirkung für `0.4.0` (`5a8ab19`).
- [x] `spec/backend-api-contract.md` §1 dokumentiert `traceparent` als optionalen HTTP-Header auf `POST /api/playback-events` mit defensiver Server-Validierung (kein 4xx bei kaputtem Header). §3.7 ergänzt `correlation_id` (ab §3.2 gesetzt; historische Leerwerte möglich) und `trace_id` (nullable) als server-vergebene Read-Felder ab `0.4.0`-§3.2-Closeout (`5a8ab19`).
- [x] Cardinality-Regel ist in `spec/telemetry-model.md` §2.5 festgehalten: `trace_id`/`correlation_id`/`span_id` sind Prometheus-tabu — Span-Attribute, Persistenz-Spalten und Wire-Format-Felder sind die einzigen Konsumenten; Verstöße sind release-blocking via Cardinality-Smoke (`5a8ab19`).
- [x] Folge-Item für persistenten Time-Skew-Flag ist als R-5 in `docs/planning/open/risks-backlog.md` aufgenommen; §5.3 in `telemetry-model.md` verweist explizit auf den Backlog-Eintrag und markiert die Persistenz-Spalte als deferred (`5a8ab19`).

### 3.2 Server-Korrelation

Bezug: §3.1; Telemetry-Model §5.4; API-Kontrakt §3; ADR-0002 §8.1 (Schema-Spalten reserviert in §2.3).

Ziel: Backend liest `traceparent` (wenn vorhanden), erzeugt einen Server-Span pro Batch, generiert/liest `correlation_id` pro Session, persistiert `trace_id`/`span_id` als Event-/Batch-Felder auf jedem Event und persistiert `correlation_id` als Session-Scope-Feld auf `stream_sessions` sowie auf den zugehörigen Events, validiert defensiv. `StreamSession` speichert keine stabile Session-`trace_id` und keine Session-`span_id`. Sub-Tranchen-Ausgang: ein Batch-POST hinterlässt korrelations-fähige Persistenz-Daten — auch ohne SDK-`traceparent`.

DoD:

- [x] HTTP-Adapter parst `traceparent`-Header gemäß W3C-Spec (`apps/api/adapters/driving/http/traceparent.go`); bei valider `trace_id`/`parent_span_id` wird der Server-Span als Child gestartet (`withTraceParent`), sonst als Root mit Server-`trace_id` und Span-Attribut `mtrace.trace.parse_error=true` (`c3741aa`).
- [x] HTTP-Request-Span für `POST /api/playback-events` trägt die in §3.1 spezifizierten Attribute: `http.method/route/status_code`, `mtrace.batch.size`, `mtrace.batch.outcome`, `mtrace.batch.session_count`, `mtrace.project.id` auf accepted Batches, optional `mtrace.session.correlation_id`, `mtrace.trace.parse_error`, `mtrace.time.skew_warning`. Testnachweis für `mtrace.project.id` und Single-Session-Korrelation: `TestHTTP_Span_SingleSessionBatch_SetsCorrelationID` (`c3741aa`).
- [x] `domain.PlaybackEvent` ist um `TraceID`/`SpanID`/`CorrelationID` erweitert; `domain.StreamSession` trägt **nur** `CorrelationID` als Session-Scope-Feld. Es gibt keine stabile Session-`trace_id`: `trace_id`/`span_id` sind Event-/Batch-Felder und dürfen zwischen Batches derselben Session wechseln. Application- und Adapter-Code füllt die Felder konsistent (`c3741aa`).
- [x] `correlation_id` wird beim allerersten Event einer Session in `RegisterPlaybackEventBatch.resolveCorrelationIDs` erzeugt (UUIDv4 via `crypto/rand`); existing Sessions liefern sie aus dem Repository; existing Sessions ohne `correlation_id` (Legacy von vor §3.2-Closeout) bekommen via Self-Healing eine neue. SessionRepository persistiert sie in `stream_sessions.correlation_id` (`c3741aa`).
- [x] SQLite-Adapter und InMemory-Adapter schreiben und lesen die drei neuen Spalten korrekt; gemeinsamer Contract-Test `testTraceFieldsRoundTrip` deckt beide Backends ab (`c3741aa`).
- [x] Defensive Validierung: `parseTraceParent` lehnt jeden Formatfehler ab (Längen, Hex, Version, all-zero); `withTraceParent` mappt das auf `mtrace.trace.parse_error=true` ohne 4xx (`c3741aa`).
- [x] Time-Skew-Detection mit Konstante `TimeSkewThreshold = 60 * time.Second` im Use-Case; bei Treffer `BatchResult.TimeSkewWarning=true`, der HTTP-Adapter setzt das Span-Attribut (`c3741aa`).
- [x] Adapter-Contract-Tests in `persistence/contract` erweitert um `testTraceFieldsRoundTrip`; läuft identisch gegen InMemory und SQLite. Use-Case-Test-Suite erweitert um fünf Cases (neue Session, existing Session mit/ohne CorrelationID, Multi-Session, Time-Skew, Trace-Context-Durchreiche) plus `parseTraceParent`-Unit-Tests (`c3741aa`).

### 3.3 SDK-Wire-Format-Erweiterung

Bezug: §3.1; §3.2 (Server-Pfad muss bereit sein, bevor SDK Header schickt); RAK-29 (kein Breaking Change).

Ziel: Player-SDK propagiert optional einen W3C-`traceparent`-Header, wenn der Browser-Pfad einen aktiven Span hat oder das SDK selbst eine `trace_id` führen kann. Wenn kein Trace-Kontext da ist, schickt das SDK den Header **nicht** — Server-Fallback erzeugt eine eigene `trace_id`. Schema-Version bleibt `1.0`. Sub-Tranchen-Ausgang: SDK 0.4.0 kann den Header optional schicken; Server toleriert sowohl SDKs mit als auch ohne Header. Header-Casing ist mit §3.3 entschieden; das exakte Header-Wert-Verhalten für führende/abschließende OWS bleibt bis §3.4c bewusst offen und ist nicht Teil des §3.3-Done-Claims.

DoD:

- [x] `@npm9912/player-sdk` HTTP-Transport setzt `traceparent`-Header optional über die neue `PlayerSDKConfig.traceparent`-Provider-Funktion; ohne Provider oder bei Provider-Return `undefined`/`""` bleibt der Header weg. Provider-Throws werden im SDK still gefangen — Tracing darf den Event-Pfad nicht sabotieren (`8f3011c`).
- [x] Abwärtskompatibilität: kein Wire-Format-Bruch (Header ist additiv, Payload unverändert); ältere Backends ignorieren unbekannte Header per HTTP-Standard. SDK-Doku verweist explizit darauf (`8f3011c`).
- [x] SDK-Tests in `packages/player-sdk/tests/http-transport.test.ts` decken: Provider-Wert → Header gesetzt; Provider-`undefined` → kein Header; Provider-`""` → kein Header; kein Provider konfiguriert → kein Header; Provider-Throw → still verworfen; Provider wird pro Send aufgerufen, nicht gecached (sechs Cases) (`8f3011c`).
- [x] Schema-Version bleibt `1.0`; SDK↔Backend-Kompatibilitätscheck (CI `make gates`) bleibt grün — `EVENT_SCHEMA_VERSION` und `PLAYER_SDK_VERSION` sind unverändert (`8f3011c`).
- [x] `spec/player-sdk.md` neue Sektion „Trace-Korrelation (optional, ab `0.4.0`)" zeigt das Provider-Pattern (Beispielcode mit OpenTelemetry-Bridge), nennt die Backwards-Compat-Garantie und verweist auf den Vertrag in `spec/telemetry-model.md` §2.5 (`8f3011c`).

Closeout-Notiz: Der §3.3-Code-Commit `8f3011c` hat eine zweite Lint-Regression hinterlassen (`packages/player-sdk/scripts/public-api.snapshot.txt` enthielt `TraceParentProvider` nicht, obwohl `src/index.ts` ihn exportiert). Das §3.3-Review hat den Snapshot-Drift nicht gefangen; `make ts-lint` wäre auf `8f3011c` rot gewesen. Geheilt im Followup-Commit `7bbea4d` (Should-fix #3+#4, Anmerkungen #6+#7+#8 aus dem §3.3-Review plus Drive-by-Snapshot-Fix). Im Review-of-Review-Commit `f7dcdb9` zusätzlich: parallele Spec-Drift in `spec/player-sdk.md` (Public-API-Bulletliste) geheilt, einmaliger `console.warn` pro `HttpTransport`-Instanz für Non-String-Returns und Provider-Throws ergänzt (Observabilität ohne Send-Pfad-Sabotage), `HttpTransportOptions.silent` für Tests, JSDoc-`@see`-Verlinkung statt Doppeltexte. Lehre für Tranche 8 / Folge-Reviews: Reviewer muss `make ts-lint` und nicht nur `make ts-test` laufen lassen; Snapshot-Drift ist nicht durch `tsc --noEmit` abgedeckt. Die §3.3-DoD-Items oben pinen weiterhin auf `8f3011c` als historisches Lieferdatum; das beobachtbare Schluck-Verhalten reflektiert den Stand ab `f7dcdb9` und ist in `spec/player-sdk.md` festgeschrieben.

### 3.4 Tests und Doku-Closeout

Bezug: §3.1–§3.3.

Ziel: Trace-Konsistenz ist auf allen Ebenen abgesichert (mehrere Batches einer Session teilen `correlation_id`; ungültiger Trace-Kontext führt zu sauberem Fallback; Tempo-deaktivierter Pfad funktioniert ungestört). Doku spiegelt den ausgelieferten Stand. Sub-Tranchen-Ausgang: Roadmap §2 Schritt 31 ist auf ✅ aktualisierbar.

§3.4 ist in drei Sub-Tranchen geschnitten: §3.4a sichert das Server-Verhalten aus §3.2 mit Backend-Tests ab (rein server-seitig, nutzt Use-Case + HTTP-Adapter + tracetest.SpanRecorder); §3.4b deckt zwei Cross-Cutting-Pfade zwischen SDK und Server ab, die §3.3-Review als Should-fix #1/#2 markiert hat (Cross-Version-Kompat, E2E-Garbage); §3.4c finalisiert die Spec-Texte und schließt Roadmap Schritt 31.

#### 3.4a Backend-Tests Trace-Konsistenz

Bezug: §3.2 (Server-Pfad mit Spans, `correlation_id`-Resolver, `parseTraceParent`, Time-Skew); ADR-0002 §8.1; Telemetry-Model §2.5.

Ziel: Das in §3.2 ausgelieferte Server-Verhalten ist durch wiederholbare Backend-Tests gegen Use-Case + HTTP-Adapter abgesichert; jeder spezifizierte Pfad (Multi-Batch-Konsistenz, fehlender Kontext, ungültiger Kontext, Session-Ende, Time-Skew, Tempo-deaktiviert) hat einen eigenen Test mit klaren Assertions auf `trace_id`, `correlation_id` und Span-Attributen. Sub-Tranchen-Ausgang: Reviewer kann §3.2-Lieferung gegen §3.4a-Tests nachvollziehen, ohne externes Trace-Backend zu brauchen. Das OWS-Verhalten des `traceparent`-Header-Werts bleibt bis §3.4c eine explizite Doku-/Test-Abhängigkeit; §3.4a bestätigt nur fehlenden, validen und offensichtlich ungültigen Header-Kontext.

DoD:

- [x] Backend-Test deckt Trace-Konsistenz über mehrere Batches einer Session: drei aufeinanderfolgende Batches mit gleicher `session_id` produzieren drei verschiedene `trace_id`-Werte (jeder Batch ein Trace), aber **dieselbe** `correlation_id` an allen Events und der Session — `TestHTTP_Trace_MultiBatchSameSessionConsistency` (`f329d5f`).
- [x] Backend-Test deckt fehlenden Client-Kontext: Batch ohne `traceparent` → Server generiert `trace_id`, `mtrace.trace.parse_error` ist nicht gesetzt — `TestHTTP_Trace_MissingTraceparent_ServerGeneratesTrace` (`f329d5f`); persistiert auch die server-generierte `trace_id` aufs Event.
- [x] Backend-Test deckt ungültigen Client-Kontext: Batch mit kaputtem `traceparent` → 202 Accepted, Span-Attribut `mtrace.trace.parse_error=true`, `trace_id` ist server-generiert — bereits im §3.2-Bestand `TestHTTP_Span_TraceParent_InvalidSetsParseError` (`c3741aa`); §3.4a-Header-Comment in `trace_consistency_test.go` mappt das DoD-Item explizit dorthin (`f329d5f`).
- [x] Backend-Test deckt Session-Ende: `session_ended`-Event innerhalb eines Batches behält die `correlation_id` der Session bei und schließt den State; nachfolgende Events in derselben Session-ID erhalten dieselbe `correlation_id` (Reihenfolge ist Tranche-1-Verhalten) — `TestHTTP_Trace_SessionEnded_PreservesCorrelationID` (`f329d5f`).
- [x] Backend-Test verifiziert Time-Skew-Span-Attribut bei `|client_timestamp - server_received_at| > 60s` — bereits im §3.2-Bestand `TestHTTP_Span_TimeSkew_SetsWarning` (`c3741aa`); §3.4a-Header-Comment mappt das DoD-Item dorthin (`f329d5f`).
- [x] Backend-Test verifiziert Trace-Konsistenz **mit NoOp-`TracerProvider`** (Stand-in für Tempo-deaktivierten Pfad ohne `OTEL_TRACES_EXPORTER`): `correlation_id` bleibt gesetzt, Dashboard-Timeline ist nutzbar. Test darf kein externes Trace-Backend voraussetzen — Realisierung über den `tracenoop`-Fallback in `router.go`, ausgelöst durch `nil`-Tracer-Argument. Dieser §3.4a-Done-Claim gilt nur für Adapter-/Use-Case-Wiring; die produktive `cmd/api`-Config-Resolution ist ein §3.4c-Closeout-Gate und darf nicht als durch §3.4a abgedeckt gelesen werden — `TestHTTP_Trace_NoopTracer_CorrelationStillPersisted` (`f329d5f`).

#### 3.4b Cross-Cutting-Tests SDK ↔ Server

Bezug: §3.3-Review (Should-fix #1/#2); §3.2; §3.3.

Ziel: Zwei Pfade, die SDK und Server überspannen, sind explizit getestet — Vorwärtskompat zu Pre-§3.2-Backends und das Garbage-Traceparent-Ende-zu-Ende-Verhalten. Sub-Tranchen-Ausgang: das §3.3-Review hat keine Test-Lücken mehr offen. Der Done-Claim umfasst noch keine Entscheidung, ob führende/abschließende OWS im `traceparent`-Header-Wert getrimmt oder als Parse-Error behandelt wird; diese Abnahme bleibt §3.4c vorbehalten.

DoD:

- [x] Cross-Version-Vertragstest (aus §3.3-Review, Should-fix #1): SDK `0.4.0` mit
  konfiguriertem `traceparent`-Provider gegen einen Server-Handler auf
  `0.3.x`-Verhaltensstand (kein Header-Lesen, keine `correlation_id`-Persistenz)
  liefert weiterhin `202 Accepted`; der Header darf nicht zu Validierungs-/Parser-
  Fehlern führen. Realisierung in zwei Hälften:
  Server-seitig `TestHTTP_Trace_CrossVersion_LegacyHandlerAcceptsTraceParent`
  mit minimalem `legacyPlaybackHandler` (snapshotted ausschließlich
  „liest `traceparent` nicht"; keine weiteren 0.3.x-Verhaltensdetails)
  (`6fdc8d0`).
  SDK-seitig `cross-version against pre-§3.2 server`-Block in
  `packages/player-sdk/tests/http-transport.test.ts` mit
  `sends successfully against a 0.3.x-shaped mock that ignores the header`
  — `HttpTransport.send` läuft gegen einen 202-Mock-Server, der den Header
  nicht liest. Beide Hälften zusammen ergeben nur einen Smoke-Nachweis
  für „zusätzlicher Header sabotiert den SDK-Send-Pfad nicht" und
  „ein Handler, der `traceparent` ignoriert, bleibt 202-fähig". Sie
  sind **kein** vollständiger Beleg gegen die echte `0.3.x`-Routing-/
  Middleware-Pipeline. Der reale `0.3.x`-E2E-Pfad ist deshalb ein
  §3.4c-Closeout-Gate; bis dahin darf der Plan nicht behaupten, echte
  `0.3.x`-Kompatibilität sei vollständig bewiesen.
- [x] E2E-Test mit kaputtem `traceparent` (aus §3.3-Review, Should-fix #2):
  hybrider Schnitt.
  Server-Seite durch `TestHTTP_Span_TraceParent_InvalidSetsParseError`
  aus §3.2 abgedeckt (`c3741aa`, 202 + `mtrace.trace.parse_error=true`).
  SDK-Seite durch zwei vitest-Cases in
  `packages/player-sdk/tests/http-transport.test.ts` unter
  `describe("garbage traceparent string")` (`6fdc8d0`):
  Garbage-String wird 1:1 weitergereicht
  (`forwards the garbage string verbatim and keeps the SDK path quiet`);
  202 ist Erfolg ohne Retry — genau ein `fetch`, kein Sleep
  (`treats 202 as success and does not retry, even with a garbage traceparent`).
  Kein `console.warn` — Garbage ist `typeof === "string"` und triggert
  weder den Throw- noch den Non-String-Pfad aus `f7dcdb9`.

#### 3.4c Doku-Closeout und Roadmap-Marker

Bezug: §3.1–§3.3; §3.4a–§3.4b.

Ziel: Spec-Texte sind final mit dem Code synchronisiert; Roadmap Schritt 31 ist als ✅ markierbar; offene Items aus dem §3.3-Review (Anmerkung #5 Header-Casing/OWS) und aus dem §3.2-Review (R-6-Risikogrenze, Session-`trace_id`-Semantik) sind eingearbeitet. Sub-Tranchen-Ausgang: Tranche 2 ist abgeschlossen, Tranche 3 kann starten, ohne dass Tranche 3 Trace-Grundsatzfragen erneut entscheiden muss.

Closeout-Regeln:

- §3.4c ist Doku-/Plan-Closeout, kein Release-Bump. `PLAYER_SDK_VERSION`, Root-Versionen und `CHANGELOG.md` bleiben Tranche-8-Arbeit.
- Normative Aussagen stehen in `spec/telemetry-model.md`, `spec/backend-api-contract.md` und `spec/player-sdk.md`; dieser Plan referenziert nur Lieferstand, Commit und Restgrenzen.
- Wenn Code und Spec voneinander abweichen, muss §3.4c entweder den Code nachziehen oder die Spec bewusst korrigieren/deferieren. Eine bekannte Abweichung darf nicht nur im Fließtext stehen.
- **Tranche-3-Blocker:** Solange irgendein §3.4c-DoD-Item offen ist, bleiben Status-Header, Roadmap Schritt 31 und die Tranche-2-Matrix offen. Tranche 3 darf erst starten, wenn §3.4c vollständig abgehakt ist; einzelne bereits abgeschlossene §3.1–§3.4b-Pfade heben diesen Blocker nicht auf. Das gilt insbesondere für OWS-Verhalten, echten `0.3.x`-Cross-Version-Pfad und produktiven Tempo-disabled-Start.

DoD:

- [ ] Header-Casing/Whitespace-Vertrag (aus §3.3-Review, Anmerkung #5) ist in `spec/backend-api-contract.md` §1 und in Backend-Tests synchronisiert: Header-Name case-insensitiv; Header-Wert-Verhalten für führende/abschließende OWS ist exakt das implementierte Verhalten (entweder `strings.TrimSpace` + gültiger Parent oder parse_error-Fallback, aber nicht ungetesteter Fließtext); SDK schreibt lowercased `traceparent`. Solange dieser OWS-Case nicht mit mindestens einem Server-Test festgezurrt ist, bleibt §3.4c blockierend.
- [ ] `mtrace.project.id` ist nicht mehr als Drift geführt: `spec/telemetry-model.md` dokumentiert es als Pflichtattribut für accepted Batches bzw. nach erfolgreicher Project-Auflösung und als bewusst unset für Rejects vor Project-Auflösung; Plan verweist auf `TestHTTP_Span_SingleSessionBatch_SetsCorrelationID`.
- [ ] `spec/telemetry-model.md` ist final konsistent mit Code: Hybrid-Strategie, ein Server-Span pro Batch, Persistenzquelle pro Feld, Time-Skew nur als Span-Attribut, Sampling-Auswirkung und Prometheus-Cardinality-Grenzen sind in einem zusammenhängenden Abschnitt festgeschrieben.
- [ ] `session_id`-Span-Attribut-Verbot ist konsistent dokumentiert und getestet: ab `0.4.0` setzt der Server in keinem OTel-Span `session_id`; Single-Session-Suche läuft ausschließlich über `mtrace.session.correlation_id`. Historische Aussagen zur Zulässigkeit von `session_id` als Span-Attribut sind auf `0.1.x` begrenzt und nicht Teil des Tranche-2-Vertrags.
- [ ] Cross-Version-Kompatibilität ist scope-korrekt geschlossen: entweder läuft ein echter `0.3.x`-E2E-Pfad (alte Routing-/Middleware-Pipeline, nicht nur `legacyPlaybackHandler`/Mock) mit `traceparent`-Header auf `202`, oder der Abnahmetext wird explizit auf den bisher gelieferten Smoke-Scope downgraded und echte `0.3.x`-Kompatibilität bleibt als Restgrenze dokumentiert.
- [ ] `spec/backend-api-contract.md` §1 / §3 / §3.7 reflektiert das ausgelieferte Header-Verhalten und die neuen Read-Felder: `trace_id` nullable und batch-bezogen, `correlation_id` pro Session stabil, `span_id` nur als technisches Event-Feld falls im Read-Pfad offengelegt; `GET /api/stream-sessions` exponiert keine Session-`trace_id`; Fehlerklassifikation bleibt unverändert bei 202/4xx aus dem normalen Event-Vertrag.
- [ ] Tempo-deaktivierter Produktivstart ist als Closeout-Gate abgesichert: ein Test oder expliziter Config-Check deckt den `cmd/api`-Auflösungsweg ohne aktives Trace-Backend ab und zeigt, dass daraus der NoOp-/nil-Tracer-Pfad entsteht. Der §3.4a-Router-Test allein reicht für dieses Gate nicht.
- [ ] Legacy-Korrektheitsgrenze ist final dokumentiert: §3.4c entscheidet ausdrücklich gegen ein historisches Backfill für vor §3.2 geschriebene `playback_events.correlation_id`; Self-Healing gilt nur für `stream_sessions.correlation_id` und neu persistierte Events nach dem nächsten Batch. API-Kontrakt und Telemetry-Model nennen leere Legacy-Event-`correlation_id`s als degradierten Read-Fall, nicht als Vertragsbruch.
- [ ] `spec/player-sdk.md` bleibt synchron zum tatsächlichen SDK-Verhalten aus `f7dcdb9`: Provider wird pro Send aufgerufen, `undefined`/`""`/Non-String/Throw schicken keinen Header, Throw/Non-String warnen höchstens einmal pro `HttpTransport`-Instanz, Garbage-String wird als String unverändert weitergereicht.
- [ ] R-6 (`correlation_id`-Race) ist im Plan-Closeout referenziert: nicht blockierend für Tranche 2, aber als explizite Restgrenze für Tranche 8/operative Beobachtung geführt. Falls vor Tranche 8 ein Mismatch zwischen `stream_sessions.correlation_id` und `playback_events.correlation_id` beobachtet wird, wird R-6 release-blocking.
- [ ] Test-/Review-Gate für den Closeout ist dokumentiert: mindestens `make ts-lint` plus die relevanten Backend-/SDK-Test-Slices aus §3.4a/§3.4b. Grund: §3.3 hat gezeigt, dass Snapshot- und Spec-Drift nicht durch `make ts-test` allein auffallen.
- [ ] `docs/planning/in-progress/roadmap.md` Schritt 31 ist auf ✅ gesetzt; Status-Header und Verweis auf den Tranche-2-Closeout-Stand sind aktualisiert.
- [ ] Dieser Plan ist nach dem Closeout aktualisiert: Status-Header oben nennt Tranche 2 vollständig abgeschlossen, Tabelle in §1 markiert Tranche 2 ✅, §3.4c-DoD-Items tragen Commit-Hash, und Tranche 3 bleibt der nächste offene Trace-Korrelationsschritt.

---

## 4. Tranche 3 — Manifest-/Segment-/Player-Korrelation

Bezug: RAK-30; RAK-29; Stream Analyzer aus `0.3.0`; F-68..F-81; Telemetry-Model §1.

Ziel: Manifest-Requests, Segment-Requests und Player-Events werden einer gemeinsamen Session-Timeline zugeordnet. Normative Priorität: `correlation_id` ist der primäre Zuordnungsschlüssel und Source-of-Truth für Dashboard/API; eingehende SDK-Events liefern weiter `session_id`, aus der der Server die `correlation_id` resolved. `trace_id` bleibt optionaler, batchbezogener Debug-Kontext für Tempo und darf keine Timeline-Zuordnung allein tragen. RAK-30 ist Soll; Lücken müssen sichtbar, testbar und erklärbar bleiben.

Abnahmegrenzen:

- **Mindestkorrelation:** Jedes neu vom SDK erzeugte Manifest-, Segment- und Player-Ereignis muss im Ingest-Payload dieselbe technische Session-Partition (`project_id` + `session_id`) wie die zugehörigen Player-Events tragen. Der bestehende Server-Resolver erzeugt oder liest daraus die Session-`correlation_id`; nach Persistenz muss jedes akzeptierte Event dieselbe `playback_events.correlation_id` wie `stream_sessions.correlation_id` tragen. Eingehende Events ohne `session_id` bleiben normale `422`-Validierungsfehler nach API-Kontrakt §5; eine vom Client gelieferte `correlation_id` ist im POST-Wire-Format nicht zulässig und wird nicht als Source-of-Truth akzeptiert. Fehlende Browser-Timing-/Resource-Daten dürfen Detailfelder reduzieren, aber nicht die Session-Zuordnung entfernen.
- **Storage-/Repository-Prerequisite:** Der aktuelle `stream_sessions`-Store ist vor Tranche 3 `session_id`-global (`primary_key: [session_id]`, Repository-Reads per `Get(ctx, session_id)`, `List` ohne `project_id`-Filter). Tranche 3 muss Schema, Ports, Repository-Methoden, Resolver, InMemory-Adapter, SQLite-Adapter und HTTP-Read-Handler auf projekt-skopierte Session-Zugriffe heben: eindeutiger Key `(project_id, session_id)`; alle `List`/Cursor-, `Get`/Detail-, Event-Read-, `Upsert`-/Resolver- und Analyzer-Linking-Pfade nehmen `project_id` entgegen und filtern danach. Eine Beweisoption über global eindeutige client-gelieferte `session_id`s ist nicht zulässig. Cross-Project-Kollisionstest ist Pflicht: gleiche `session_id` in zwei Projekten darf weder Session-Detail noch Event-Reads, Cursor-Pagination oder Analyzer-Linking projektübergreifend vermischen.
- **R-6-Prerequisite:** Der bekannte `correlation_id`-First-Insert-Race aus `docs/planning/open/risks-backlog.md` R-6 muss vor Abschluss von Tranche 3 behoben oder durch DB-finale Correlation-ID-Rückgabe nachweislich unmöglich gemacht werden. Bis dahin gilt die harte Gleichheit `playback_events.correlation_id == stream_sessions.correlation_id` nur für nicht-konkurrierende Flows; Tranche 3 darf nicht als vollständig abgeschlossen markiert werden.
- **Analyzer-Linkage-Priorität:** Für explizite Analyzer-Verknüpfung gilt: `correlation_id` gewinnt vor `session_id`, aber nur innerhalb eines gültig aufgelösten Project-Kontexts. Verlinkte `POST /api/analyze`-Requests müssen `X-MTrace-Token` (und falls später aktiv: `X-MTrace-Project`) erfolgreich auf ein `project_id` auflösen. Fehlt dieser Kontext bei gesetzter `correlation_id` oder `session_id`, antwortet die API mit dem Auth-/Kontextfehler aus dem API-Kontrakt und führt keinen Session-Lookup aus; dieser Fall ist kein stilles `detached`. Nur Requests ohne Link-Felder bleiben ohne Project-Kontext erfolgreich und erhalten `session_link.status="detached"`. Alle Link-Lookups laufen über `(project_id, correlation_id)` bzw. `(project_id, session_id)`. `correlation_id` allein ohne Treffer im Project liefert `session_link.status="not_found_detached"`. Wenn beide IDs angegeben sind, muss zuerst `correlation_id` im Project existieren und danach `session_id` im selben Project zur Session mit dieser `correlation_id` auflösen; unbekannte `correlation_id` gewinnt also nicht durch einen bekannten `session_id`-Fallback und liefert `not_found_detached`, Cross-Project-`correlation_id` gilt ebenfalls als `not_found_detached`, Mismatch im selben Project liefert `conflict_detached`. Nur `session_id` allein ist als Fallback für bestehende oder bereits selbst-geheilte Sessions erlaubt; eine unbekannte `session_id` erzeugt durch `POST /api/analyze` keine neue Session und führt zu `session_link.status="not_found_detached"`. Kompatibilitätsentscheidung: Ab Tranche 3 gibt `POST /api/analyze` für **alle** erfolgreichen Requests die Hülle `{analysis, session_link}` zurück, auch ohne Link-Felder; ungebundene Requests erhalten `session_link.status="detached"`. Der Breaking-Change wird im API-Kontrakt und in den Contract-Tests festgeschrieben.
- **Degradationsfälle:** Fehlende oder blockierte Resource-Timing-Daten, CORS-Lücken, Service-Worker-Interception, CDN-Redirects und hls.js-Signale mit unvollständigen Details erzeugen ein Netzwerkevent mit flachen Meta-Keys `meta["network.detail_status"]="network_detail_unavailable"` und optional `meta["network.unavailable_reason"]` gemäß `spec/telemetry-model.md` §1.4. Event-Meta und `session_boundaries[]` verwenden dieselbe kontrollierte Reason-Domäne: `native_hls_unavailable`, `hlsjs_signal_unavailable`, `browser_api_unavailable`, `resource_timing_unavailable`, `cors_timing_blocked`, `service_worker_opaque`; unbekannte Werte oder Werte außerhalb `^[a-z0-9_]{1,64}$` werden mit `422` abgelehnt. Wenn ein Browser-/Native-HLS-Pfad gar kein Manifest-/Segment-Signal liefert, wird kein synthetisches Netzwerkereignis erfunden; stattdessen exponiert der Session-Read-Pfad Tranche 3 ein nicht-eventbasiertes Feld `network_signal_absent` als Liste von Objekten `{ "kind": "manifest" | "segment", "adapter": "hls.js" | "native_hls" | "unknown", "reason": "<machine_reason>" }` im Session-Block. Schreibpfad ist ein optionaler Batch-Wrapper-Block `session_boundaries[]` in `POST /api/playback-events` mit `kind="network_signal_absent"`, `project_id`, `session_id`, `network_kind`, `adapter`, kontrolliertem `reason` und `client_timestamp`; er wird zusammen mit einem normalen Event-Batch gesendet, zählt nicht als Event, besitzt keinen `event_name` und ändert `schema_version: "1.0"` nicht. Pro Batch sind maximal 20 Boundaries zulässig, sie zählen ins Body-Size-Budget des SDK/Backends, und jede Boundary muss eine `(project_id, session_id)`-Partition referenzieren, die mindestens ein Event im selben Batch trägt. Boundary-only-Batches ohne `events` oder Boundaries für fremde/nicht enthaltene Sessions bleiben in Tranche 3 außerhalb des Vertrags und liefern `422`. Das Backend persistiert daraus den Boundary-Record, auch wenn kein Manifest-/Segment-Event existiert. Persistenzvehikel ist eine durable Session-Metadaten-Spalte oder ein äquivalenter session-skopierter Capability-/Boundary-Record; der Wert darf nicht nur aus flüchtigem Prozesszustand abgeleitet werden und muss nach API-Restart identisch lesbar bleiben. Die Dashboard-Sichtbarkeit dieser Grenze ist Tranche-4-Scope. Beide Varianten sind akzeptierte Degradation und müssen getrennt getestet werden.
- **Schema-Entscheidung:** Tranche 3 verwendet ausschließlich additive, flache `meta`-Keys nach dem Muster `network.*`/`timing.*` im bestehenden Event-Wire-Schema `1.0`; verschachtelte `meta.network`-Objekte sind nicht zulässig, weil `packages/player-sdk/src/types/events.ts` für `EventMeta` nur skalare Werte erlaubt. Die Event-Namen bleiben die bereits im Contract/SDK vorhandenen `manifest_loaded` und `segment_loaded`. Keine neuen `event_name`-Werte in Tranche 3; falls ein weiterer Event-Typ nötig wäre, müssen `contracts/event-schema.json`, `packages/player-sdk/src/types/events.ts`, Public-API-Snapshot und Compat-Tests explizit erweitert werden und der Zusatz wird als eigenes DoD-Item geführt. Kein Breaking Change und keine neue Major-Schema-Version in `0.4.0`; falls ein benötigtes Feld nicht additiv modellierbar ist, wird es deferred statt per Migration erzwungen.
- **URL-Datenschutz:** Segment-/Manifest-URLs dürfen weder Prometheus-Labels noch rohe, credential-haltige Persistenzwerte werden. Vor Persistenz/Anzeige gilt die feste Redaction-Matrix für den vorgesehenen URL-Repräsentanten und für alle URL-verdächtigen generischen Meta-Keys (`url`, `uri`, `manifest_url`, `segment_url`, `media_url`, `network.url`, `network.redacted_url`, `request.url`, `response.url` und case-insensitive Varianten). Scheme, Host und nicht-sensitive Pfadsegmente dürfen erhalten bleiben; Query und Fragment werden vollständig entfernt; `userinfo` wird entfernt; signierte/credential-artige Query-Parameter (`token`, `signature`, `sig`, `expires`, `key`, `policy` und case-insensitive Varianten) werden nicht gespeichert; ein Pfadsegment ist tokenartig, wenn es ≥ 24 Zeichen lang ist und mindestens 80 % seiner Zeichen aus `[A-Za-z0-9_-]` bestehen, wenn es ein Hex-String mit gerader Länge mindestens 32 ist, oder wenn es bekannte JWT-/SAS-/Signed-URL-Muster trägt. Tokenartige Pfadsegmente werden ausschließlich durch `:redacted` ersetzt; es wird kein stabiler Hash und kein Gleichheitsmarker persistiert. Unbekannte `meta`-Keys mit String-Werten, die als absolute URL parsebar sind oder `://` enthalten, werden ebenfalls vor Persistenz redigiert oder verworfen; rohe URL-Werte dürfen in keinem Meta-Feld persistieren. Tests decken Query-/Fragment-Redaction, `userinfo`, signierte Query-Parameter, Token-Parameter, JWT-/Base64URL-Pfadsegmente sowie bösartige/Legacy-Payloads mit rohen URLs in generischen Meta-Keys ab.
- **Analyzer-Semantik:** `POST /api/analyze` bleibt in Tranche 3 standardmäßig getrennt von Live-Player-Sessions. Eine Verknüpfung mit einer Session ist nur zulässig, wenn die Request-Seite explizit eine vorhandene `correlation_id` oder, als Fallback, `session_id` übergibt; andernfalls wird das Analyzer-Ergebnis als unabhängige Manifestanalyse angezeigt und nicht in die Player-Timeline gemischt.

DoD:

- [ ] Player-SDK erfasst jedes von hls.js gelieferte Manifest- und Fragment-/Segment-Signal als `manifest_loaded` bzw. `segment_loaded`, außer Sampling oder Degradation ist ausdrücklich konfiguriert, dokumentiert und getestet; "mindestens eins pro Session" reicht nur als Smoke-Schwelle, nicht als RAK-30-Abnahme.
- [ ] hls.js-Mapping ist vor Umsetzung als Tabelle festgeschrieben und getestet: kanonische Quellen für `manifest_loaded` und `segment_loaded` sind benannt (z. B. `MANIFEST_LOADED`/`LEVEL_LOADED` vs. `FRAG_LOADED` inklusive Init-Segment-Regel), Retries und Redirects erzeugen keine doppelten semantischen Events, Dedup-Schlüssel nutzen ausschließlich hls.js-native Identität plus stabile Session-Kontexte (`project_id`, `session_id`, `sequence_number` sowie Frag-Felder wie `sn`, `cc`, `type`, `level`, Init-Segment-Marker); redigierte URLs sind nur persistierte Diagnose und dürfen nicht primärer Dedup-Schlüssel sein. Tests decken Manifest-Reload, Fragment-Retry, Init-Segment, signierte Segment-URLs mit gleicher Redaction und Level-Reload ab.
- [ ] Storage-/Repository-Prerequisite ist implementiert: Schema/Migration stellt projekt-skopierte Session-Eindeutigkeit her, Port-Signaturen und Application-Resolver nehmen `project_id`, InMemory- und SQLite-Adapter nutzen `(project_id, session_id)` für `List`/Cursor, `Get`/Detail, Event-Reads, `Upsert`/Self-Healing und bieten `GetByCorrelationID(project_id, correlation_id)` für Analyzer-Linking. Legacy-Leerwerte (`correlation_id=""`) liefern dabei keinen Treffer; Duplikate innerhalb eines Projects sind ein Datenfehler und blockieren Tranche 3.
- [ ] HTTP-Read-Handler lösen Project-Kontext vor Session-List, Session-Detail, Event-Read und Analyzer-Linking auf und reichen `project_id` in die Application-/Repository-Pfade weiter; Session-List, Session-Detail und Event-Read verlangen ab Tranche 3 `X-MTrace-Token`. Fehlender oder ungültiger Token liefert `401`; positive Tests mit gültigem Token beweisen projekt-skopierte Ergebnisse. Dashboard-/Client-API sendet den Header für alle `GET /api/stream-sessions*`-Aufrufe, und Nutzer-/Smoke-Doku zeigt den Header in Read-Beispielen. Session-List- und Event-Cursor wechseln auf `cursor_version: 3`: List-Cursor enthalten den Project-Scope (`project_id` oder Scope-Hash), Event-Cursor enthalten den Collection-Scope über `(project_id, session_id)` oder einen daraus abgeleiteten Scope-Hash. v2-Cursor ohne Scope werden nach Aktivierung der projekt-skopierten Read-Pfade als `cursor_invalid_legacy` abgewiesen; v3-Cursor mit fremdem Project- oder Session-Scope liefern `cursor_invalid_malformed`, sodass ein Event-Cursor aus Session A nicht für Session B im selben Project nutzbar ist.
- [ ] Cross-Project-Kollisionstest deckt InMemory und SQLite ab: dieselbe `session_id` in zwei Projekten erzeugt getrennte Sessions und unterschiedliche `correlation_id`s; List/Cursor, Detail, Event-Reads und Analyzer-Linking können nicht projektübergreifend auflösen.
- [ ] R-6 ist technisch geschlossen: konkurrierende Erst-Batches derselben `(project_id, session_id)` persistieren ausschließlich Events mit der DB-finalen `stream_sessions.correlation_id`; der Race-Test läuft mindestens gegen SQLite und beweist keinen Mismatch zwischen `playback_events.correlation_id` und `stream_sessions.correlation_id`.
- [ ] Event-Schema bleibt `1.0` und nutzt für Netzwerkereignisse die bestehenden Event-Namen `manifest_loaded`/`segment_loaded` plus additive, flache `meta`-Keys wie `network.kind`, `network.detail_status`, `network.unavailable_reason` und `network.redacted_url`. Reservierte `network.*`- und `timing.*`-Keys werden inbound typvalidiert: `network.kind`, `network.detail_status`, `network.unavailable_reason` und `network.redacted_url` sind Strings mit den dokumentierten Domänen/Redaction-Regeln; `network.unavailable_reason` ist nur erlaubt, wenn `network.detail_status="network_detail_unavailable"` ist, und bei `network.detail_status="available"` führt er immer zu `422`. Timing-Werte sind Zahlen oder dokumentierte RFC3339-Strings. Objekte, Arrays, freie Strings außerhalb der Domäne oder rohe URLs in reservierten Keys führen zu `422` vor Persistenz. `contracts/event-schema.json`, `packages/player-sdk/src/types/events.ts` und Public-API-Snapshot bleiben bezüglich Event-Name-Union und `EventMetaValue`-Skalartyp unverändert; Contract-Tests sichern, dass ältere Backends unbekannte additive Netzwerk-Meta-Keys weiter ignorieren und neue Backends die neuen Meta-Felder lesen.
- [ ] `network.unavailable_reason` und `session_boundaries[].reason` verwenden dieselbe Reason-Enum und dasselbe Pattern/Längenlimit; Contract-Tests decken alle zulässigen Reason-Werte sowie unbekannte/gefährliche Werte in Event-Meta und Boundary-Block ab.
- [ ] `session_boundaries[]` ist als optionaler Batch-Wrapper-Block in `spec/backend-api-contract.md`, `spec/telemetry-model.md`, `contracts/event-schema.json` und der SDK-Public-API typisiert. `reason` ist kontrolliert (Enum plus Pattern/Längenlimit) und kein freier persistierter Text. Der komplette Wrapper wird vor jedem Write validiert oder gemeinsam transaktional persistiert: ein invalider Boundary-Block liefert `422`, persistiert weder Events noch Boundaries und erhöht `accepted` nicht. Contract-Tests decken einen Batch mit Events plus Boundary-Block, einen invaliden Boundary-Block mit sonst validen Events, zu viele Boundaries (`>20`), Boundaries für nicht im selben Batch enthaltene `(project_id, session_id)`-Partitionen, Body-Size-Budget inklusive Boundaries und Legacy-Kompatibilität ab (alte Backends ignorieren den Block oder die SDK-Kompatibilitätsmatrix markiert ihn als erst ab 0.4.0 sendbar).
- [ ] Backend normalisiert Netzwerkereignisse in den bestehenden Session-/Event-Store; jedes akzeptierte Netzwerkereignis erhält dieselbe `correlation_id` wie die zugehörigen Player-Events derselben `project_id`+`session_id`-Partition. `trace_id` darf vorhanden sein, ist aber nicht das Abnahmekriterium für Timeline-Zuordnung.
- [ ] URL-Redaction ist vor Persistenz und Dashboard-Anzeige umgesetzt und getestet: keine Query-/Fragment-Speicherung, kein `userinfo`, keine signierten Query-Parameter, keine Credential-/Token-Parameter, keine tokenartigen Rohpfadanteile, keine URL-Labels in Prometheus. Der Event-Store enthält nur redigierte URL-Repräsentanten gemäß Abnahmegrenze und keinen stabilen Token-Hash. Tests injizieren auch Legacy-/Angreifer-Payloads mit rohen URLs in `meta.url`, `meta.network.url`, `meta.segment_url`, `meta.manifest_url` und unbekannten URL-artigen Meta-Keys und beweisen, dass nichts roh persistiert wird. Boundary-Felder, insbesondere `session_boundaries[].reason`, werden gegen rohe URLs, Token-Strings und HTML/Script-Fragmente negativ getestet.
- [ ] Falls einzelne Manifest-/Segment-Daten nur als Event-Timeline und nicht als OTel-Span abbildbar sind, ist diese Grenze im API-/Doku-Vertrag nachvollziehbar; Dashboard-Sichtbarkeit folgt in Tranche 4. Diese Events behalten trotzdem `correlation_id`.
- [ ] Tests decken gemischte Player-, Manifest- und Segment-Ereignisse innerhalb einer Session ab und prüfen gleiche `correlation_id`, getrennte batchbezogene `trace_id`-Semantik und die dokumentierten Timeline-only-Ausnahmen.
- [ ] Tests decken die Degradationsmatrix ab: fehlende SDK-Felder, blockierte Browser-Timings, CORS-/Resource-Timing-Lücken als `network_detail_unavailable`-Events sowie mindestens ein Native-/Browser-Limitierungsfall ohne Netzwerksignal als durable API-/Persistenzgrenze `network_signal_absent` im Session-Read-Shape werden als akzeptierte Degradation geprüft. Zusätzlich gibt es einen Implementierungstest für den Schreibpfad: SDK-/Adapter-Capability-Signal ohne Manifest-/Segment-Event erzeugt den session-skopierten Boundary-Record; ein API-Restart-Test beweist, dass `network_signal_absent` stabil erhalten bleibt. Dashboard-Anzeige folgt in Tranche 4.
- [ ] Analyzer-Linking ist im Hexagon modelliert, nicht nur im HTTP-Adapter: `domain.StreamAnalysisRequest` trägt optionale `correlation_id`/`session_id` und den aufgelösten Project-Kontext, der Analyze-Use-Case erhält ProjectResolver/SessionRepository bzw. explizite Link-Dependencies, und der Driving-Port liefert ein neues Resultmodell wie `AnalyzeManifestResult{Analysis, SessionLink}` statt nur `domain.StreamAnalysisResult`. Der HTTP-Adapter dekodiert nur Wire-Felder/Header und mappt das Use-Case-Result auf `{analysis, session_link}`; Link-Status, Cross-Project-Regeln und `not_found_detached`/`conflict_detached` entstehen im Application-Layer. Unit-Tests decken Use-Case-Statusmatrix und Port-Vertrag ab.
- [ ] `spec/backend-api-contract.md` §1 und §4 sind auf endpoint-spezifische Auth bereinigt: keine globale `X-MTrace-Token`-Pflichtformulierung bleibt übrig; `POST /api/playback-events` und Session-Reads sind tokenpflichtig; `POST /api/analyze` ist nur bei gesetzter `correlation_id` oder `session_id` tokenpflichtig, während ungebundene Analyze-Requests ohne Token erfolgreich `session_link.status="detached"` liefern. Tests pinnen: `/api/analyze` ohne Token und ohne Link-Felder -> `200` mit `detached`, `/api/analyze` mit `correlation_id` oder `session_id` ohne/ungültigen Token -> `401`, Playback- und Session-Read-Endpunkte bleiben tokenpflichtig.
- [ ] Analyzer-Ergebnisse aus `POST /api/analyze` werden ohne explizite Session-/`correlation_id`-Bindung nicht in die Player-Timeline gemischt; mit expliziter Bindung ist die Verknüpfung über `(project_id, correlation_id)` testbar und ein fehlender/ungültiger Project-Kontext liefert den API-Kontraktfehler statt `detached`. Tests decken die Response-Hülle `{analysis, session_link}` für ungebundenen Request ohne Link-Felder (`detached`), nur bekannte `correlation_id` (`linked`), nur unbekannte oder project-fremde `correlation_id` (`not_found_detached`), nur vorhandenen `session_id`-Fallback (`linked`), unbekannte `session_id` (`not_found_detached`), beide konsistent (`linked`), beide gesetzt mit unbekannter oder project-fremder `correlation_id` und bekannter `session_id` im Request-Project (`not_found_detached`), beide im selben Project bekannt aber widersprüchlich (`conflict_detached`), fehlenden/ungültigen Project-Kontext bei gesetzten Link-Feldern (`401`/Kontraktfehler) und Cross-Project-Mismatch ab. `OPTIONS /api/analyze` bekommt einen eigenen Analyze-Preflight mit `Access-Control-Allow-Methods: POST, OPTIONS` und `Access-Control-Allow-Headers: Content-Type, X-MTrace-Token, X-MTrace-Project`; Tests decken linked und unlinked Analyze-POST aus erlaubtem Origin samt Preflight ab. Alle Fälle sind dokumentiert.
- [ ] Der Breaking Change des `{analysis, session_link}`-Wrappers ist in Nutzer-/Smoke-Doku und Tests nachgezogen: `docs/user/stream-analyzer.md`, lokale Entwicklungs-/Smoke-Beispiele, API-Handler-Tests und Dashboard-/Client-API-Tests erwarten ab Tranche 3 nicht mehr das flache `AnalysisResult` als direkte `POST /api/analyze`-Antwort. Die internen `spec/contract-fixtures/analyzer/*` bleiben flaches Analyzer-Service-Wire-Format zwischen `apps/analyzer-service` und dem API-Adapter; für die öffentliche API-Hülle werden separate HTTP-/API-Fixtures oder Handler-Snapshots angelegt.
- [ ] Dokumentation benennt Grenzen der Korrelation, insbesondere Browser-APIs, CORS, Service Worker, CDN-Redirects, Native-HLS und Sampling; sie nennt `correlation_id` als Pflichtkontext und `trace_id` als optionale Debug-Vertiefung.

---

## 5. Tranche 4 — Dashboard-Session-Verlauf ohne Tempo

Bezug: RAK-32; MVP-14; F-38; F-39/F-40 nur gemäß Abnahmegrenzen unten; ADR 0002; ADR 0003.

Ziel: Das Dashboard zeigt Session-Verläufe aus der lokalen m-trace-Persistenz einfach, schnell und restart-stabil an. Tempo ist dafür nicht erforderlich.

DoD:

- [ ] Session-Liste und Session-Detailansicht lesen aus SQLite-backed API-Pfaden und zeigen Daten nach API-Restart weiter an.
- [ ] Detailansicht stellt eine Timeline aus Player-, Manifest- und Segment-Ereignissen dar, mit stabiler Reihenfolge und klarer Typ-Unterscheidung.
- [ ] Detailansicht rendert das in Tranche 3 vertraglich definierte Session-Feld `network_signal_absent[]` als sichtbaren, nicht-fehlerhaften Hinweis, ohne synthetische Manifest-/Segment-Events zu erfinden.
- [ ] Laufende Sessions sind von beendeten Sessions über echte API-Felder unterscheidbar: `spec/backend-api-contract.md` definiert im Session-Block mindestens `state`, `started_at`, `last_seen_at`, `ended_at`, `end_source`, `event_count`, `correlation_id` und `network_signal_absent[]`; Dashboard- und Backend-Tests lesen diese Felder gegen echte API-Antworten. `session_ended` und Sweeper-Ende werden nicht nur als `state="ended"`, sondern über `end_source="client"` bzw. `end_source="sweeper"` sichtbar. `end_source` ist in `stream_sessions` persistiert und per Schema-/Migration-DoD sowie Restart-Test abgesichert; ADR 0002 §8.1 bleibt die Storage-Quelle. Für Tranche 4 bleibt `Last-Event-ID` gemäß ADR 0003 die globale `playback_events.ingest_sequence`; Sweeper-only Session-Änderungen ohne Playback-Event sind deshalb bewusst Polling-only sichtbar und werden mit einem eigenen Fallback-Test abgesichert. Eine spätere SSE-Backfill-Unterstützung für Sweeper/Lifecycle erfordert eine separate `session_update_sequence` oder typisierte Lifecycle-Event-ID.
- [ ] Invalid-, dropped- und rate-limited Hinweise sind in Tranche 4 nur dann in der Session- oder Statusansicht sichtbar, wenn vorher ein API-Status-Summary-Vertrag außerhalb von `/metrics` existiert. Andernfalls wird die Anzeige explizit nach Tranche 6 verschoben; das Dashboard parst in Tranche 4 keine Prometheus-Rohdaten.
- [ ] F-39 (`API-Status anzeigen`) bleibt in Tranche 4 auf Mini-Statusquellen begrenzt: `GET /api/health`/API-Erreichbarkeit, SSE-Verbindungszustand des Dashboard-Clients und der letzte Session-Read-Fehler aus den realen Session-API-Aufrufen. Dashboard-Tests decken genau diese drei Quellen ab; aggregierte Invalid-/Dropped-/Rate-Limit-Zähler werden erst mit Tranche 6 sichtbar, sofern kein vorheriger API-Status-Summary-Vertrag eingeführt wird.
- [ ] F-40 (`Links zu Grafana, Prometheus und Media-Server-Konsole`) ist in Tranche 4 binär: entweder eine reine konfigurationsgetriebene Link-Section mit dokumentierten Config-Keys, Verhalten bei fehlenden URLs und Dashboard-Test wird umgesetzt, oder F-40 steht als explizites Deferred-/Nicht-Scope-Item mit Zieltranche Tranche 6/Observability im Plan. Es gibt keine stillschweigende F-40-Erfüllung ohne DoD-Test; bei Deferred muss Tranche 6 ein eigenes F-40-DoD für Grafana-, Prometheus- und Media-Server-Konsolenlinks enthalten.
- [ ] Duplikat- oder Replay-Klassifikationen aus der Persistenz sind in der Timeline unterscheidbar und beschädigen nicht die Default-Reihenfolge.
- [ ] Pagination oder inkrementelles Nachladen bleibt bei längeren Sessions bedienbar; Cursor-Verhalten ist restart-stabil.
- [ ] SSE-Live-Update-Mechanismus aus ADR 0003 ist implementiert; Polling bleibt Fallback für Stream-Abbruch oder nicht verfügbare SSE-Verbindung. Wegen tokenpflichtiger Session-Reads nutzt das Dashboard keinen nativen Browser-`EventSource` für geschützte Streams, sondern einen fetch-basierten SSE-Client/Polyfill, der `X-MTrace-Token` setzen kann.
- [ ] SSE-Endpunkt-Schnittstelle ist im `spec/backend-api-contract.md` als verlässlicher Vertrag dokumentiert: globaler Stream, optionaler Session-Detail-Stream, Payload-Schema, `Last-Event-ID`-/Backfill-Regel, Fehler-/Reconnect-Semantik und Polling-Fallback-Intervalle. Pro `playback_events.ingest_sequence` wird höchstens ein backfill-relevanter SSE-Frame gesendet. Der Mindestframe `event_appended` trägt die SSE-`id` der `ingest_sequence` und enthält nur Nachlade-/Invalidierungsdaten (`project_id`, `session_id`, `ingest_sequence`, `event_name`); Session-Header-Felder werden im SSE-Backfill nicht historisch rekonstruiert und sind nicht Teil des Mindestpayloads. Er ist nicht die vollständige Timeline-Zeile: Dashboard-Detailansichten laden nach `event_appended` die fehlenden Event-Read-Shape-Felder (`server_received_at`, `client_timestamp`, `sequence_number`, `delivery_status`, `meta`, `trace_id`, Event-`correlation_id` usw.) sowie den aktuellen Session-Header über `GET /api/stream-sessions/{id}` nach und rendern Timeline-Zeilen ausschließlich aus diesem REST-Read-Shape. Tranche 4 führt keinen neuen `after_ingest_sequence`-/Delta-Read-Parameter ein; bei langen Sessions nutzt das Dashboard die bestehenden `events_cursor`-/Pagination-Regeln oder lädt bewusst den Detail-Snapshot erneut, was als MVP-Grenze dokumentiert und getestet wird. Es gibt in Tranche 4 keinen separaten `session_updated`-Frame mit derselben `ingest_sequence`; Session-Änderungen ohne Playback-Event, insbesondere Sweeper-Ende, werden nicht als SSE-Event gesendet und sind Polling-only. Wenn kein Detailstream gebaut wird, muss der globale Stream diese Mindestpayloads liefern und Detailansichten zum REST-Nachladen auslösen; ein reiner Listen-Invalidierungsstream ohne `session_id`/`ingest_sequence` erfüllt die Tranche-4-SSE-DoD nicht.
- [ ] SSE-Auth ist endpoint-spezifisch dokumentiert und getestet: fehlender/ungültiger `X-MTrace-Token` liefert `401`, gültiger Token scoped globalen Stream und Backfill auf das Project, Detailstream und Backfill sind zusätzlich auf `(project_id, session_id)` begrenzt; Tests decken fehlenden/ungültigen Token, Cross-Project-Isolation und Project-Scope im Backfill ab.
- [ ] SSE-CORS ist für neue Stream-Routen explizit verdrahtet: `OPTIONS /api/stream-sessions/stream` und optional `OPTIONS /api/stream-sessions/{id}/events/stream` erlauben bei bekanntem Origin `GET, OPTIONS` und `Access-Control-Allow-Headers: X-MTrace-Token, X-MTrace-Project, Last-Event-ID`. Reconnect-Backfill nutzt beim fetch-basierten SSE-Client den `Last-Event-ID`-Header; Tests decken erlaubten Origin, unbekannten Origin, fehlenden Token nach erfolgreichem Preflight, `Last-Event-ID`-Preflight und Header-Forwarding im fetch-basierten SSE-Client ab.
- [ ] SSE-`id`/`Last-Event-ID` ist in Tranche 4 ausschließlich an `playback_events.ingest_sequence` gebunden. Reconnect-Backfill deckt Playback-Events ab; Session-Lifecycle-Updates ohne Playback-Event, insbesondere Sweeper-Ende, sind vom SSE-Backfill ausgenommen und werden nur über Polling sichtbar. Scope und Eindeutigkeit sind passend zum Stream-Typ definiert (globaler Stream braucht global eindeutige ID, Session-Stream mindestens session-eindeutige ID); Reconnect-Backfill liest ausschließlich aus SQLite und funktioniert nach API-Restart.
- [ ] SSE-Fallback-Grenzen sind hart definiert und getestet: Heartbeat-Intervall, Reconnect-Backoff, maximale Backfill-Lücke und Polling-Intervall haben konkrete Defaults sowie obere Grenzen im API-Kontrakt.
- [ ] Backend-Tests decken SSE-Stream-Header, EventSource-kompatibles Format, Heartbeats/Keepalive, Client-Abbruch und reconnect-freundliche Semantik ab.
- [ ] Dashboard-Tests decken SSE-Erfolg, Reconnect/Backfill und Polling-Fallback ab.
- [ ] Dashboard-Tests decken leere Timeline, kurze Session, lange Session, laufende Session und beendete Session über API-Mockdaten ab; Restart-Persistenz wird zusätzlich durch einen Integration-/E2E-Test mit echter SQLite-Datei und API-Neustart geprüft.
- [ ] Browser-E2E-Smoke erzeugt eine Session über einen stabilen Test-Harness (`/demo` oder dedizierte E2E-Seed-Route/API-Fixture) und prüft, dass der Session-Verlauf im Dashboard sichtbar ist; `/demo` ist nicht die einzige zulässige Datenquelle.

---

## 6. Tranche 5 — Optionales Tempo-Profil

Bezug: RAK-31; RAK-29; Architektur §2/§5; README `0.4.0`.

Ziel: Tempo kann als optionales Trace-Backend genutzt werden, ohne die lokale Dashboard-Ansicht zur Pflicht-Abhängigkeit zu machen. RAK-31 bleibt Kann-Scope: Diese Tranche darf vollständig umgesetzt oder explizit deferred werden, solange RAK-29/RAK-32, lokaler Trace-/Korrelations-Read und das bestehende `observability`-Profil ohne Tempo grün bleiben.

DoD:

- [ ] Tranche 5 hat ein binäres Scope-Gate: Entweder Tempo wird in `0.4.0` umgesetzt, oder §6 wird als Deferred mit Zielrelease markiert und darf keine Release-Gates für `0.4.0` blockieren. Deferred ist nur zulässig, wenn die Doku klar sagt, dass Tempo Debug-Tiefe ist und die Session-Timeline ohne Tempo vollständig nutzbar bleibt.
- [ ] Deferred-Abschluss ist eigenständig prüfbar: Plan/Roadmap nennen ein Zielrelease oder explizites Nicht-Scope, README und `docs/user/local-development.md` nennen Tempo als optional/deferred, normative Specs (`spec/telemetry-model.md`, `spec/backend-api-contract.md` und Architektur-Specs, falls sie Tempo erwähnen) beschreiben den tatsächlichen `0.4.0`-Status statt ein aktives Tempo-Profil zu behaupten, `make dev-observability` und `make smoke-observability` bleiben ohne Tempo grün, und alle Tempo-Implementierungsitems dieses Abschnitts sind als "nur wenn umgesetzt" markiert. Alternativ muss §8 diese Spec-Synchronisierung als blockierendes Cleanup-Gate führen.
- [ ] Nur wenn Tempo in `0.4.0` umgesetzt wird: Tempo startet über das feste zusätzliche Compose-Profil `tempo` und das Make-Target `make dev-tempo`, nicht automatisch durch das bestehende `observability`-Profil. `make dev-tempo` startet die vorhandene Collector-Schiene mit `--profile observability --profile tempo`, damit Prometheus/Grafana/OTel-Collector plus Tempo-Backend gemeinsam laufen; der Collector bleibt nicht doppelt definiert. `make dev-tempo` ist in `.PHONY` und `make help` discoverable; `make stop` beendet auch ein aktiviertes Tempo-Profil, und `make wipe` dokumentiert, ob Tempo-Daten/Volumes gelöscht oder bewusst erhalten werden. `make dev-observability` bleibt Prometheus/Grafana/OTel-Collector ohne Tempo-Backend.
- [ ] Nur wenn Tempo in `0.4.0` umgesetzt wird: Der Collector-Konfigurationsmechanismus für Tempo ist explizit und getestet: entweder ein Tempo-spezifischer Collector-Config-Mount/Compose-Override, env-gesteuerte Exporter-/Pipeline-Aktivierung mit geprüftem Disabled-Pfad oder eine gleichwertig dokumentierte Lösung. Reine Compose-Profil-Aktivierung ohne geänderte Collector-Konfiguration reicht nicht. `make dev-observability` darf keinen Tempo-Exporter konfigurieren oder Tempo-Verbindungsversuch auslösen; `make dev-tempo` muss denselben Collector-Service mit Tempo-Exporter/Pipeline starten.
- [ ] Nur wenn Tempo in `0.4.0` umgesetzt wird: OTel-Collector leitet Traces nur im Tempo-Profil an Tempo weiter. Ohne Tempo-Profil, aber mit aktivem `observability`-Profil, bleibt die bestehende Collector-Konfiguration gültig: OTLP vom API-Service ist erlaubt, Prometheus/Grafana bleiben grün, Traces gehen an den vorhandenen Collector-Debug-Exporter und erzeugen keine Tempo-Connection-Errors. Der No-Op-Pfad gehört nur zum Core-/Default-Compose-Zustand ohne aktivierten OTLP-Endpoint.
- [ ] Nur wenn Tempo in `0.4.0` umgesetzt wird: Regressionstest oder Smoke-Check deckt alle drei Startzustände ab: Core in der Repo-Default-Compose-Umgebung mit den leeren Compose-Defaults `OTEL_EXPORTER_OTLP_ENDPOINT=""`, `OTEL_EXPORTER_OTLP_PROTOCOL=""`, `OTEL_TRACES_EXPORTER=""` und `OTEL_METRICS_EXPORTER=""` (kein OTLP-Verbindungsversuch), `observability` ohne Tempo (OTLP zum Collector, keine Tempo-Exportfehler), Tempo-Profil aktiv (Collector exportiert Traces nach Tempo). Lokal gesetzte nicht-leere `OTEL_*`-Variablen dürfen erwartungsgemäß Exportversuche auslösen und sind kein Produktfehler.
- [ ] Ohne Tempo-Profil bleiben lokale `trace_id`/`correlation_id`, Dashboard-Timeline und RAK-29-Tests vollständig funktionsfähig.
- [ ] Nur wenn Tempo in `0.4.0` umgesetzt wird: Trace-Suche oder ein Link-Konzept ist dokumentiert, falls Dashboard und Tempo gemeinsam laufen: Session-Suche nutzt primär `mtrace.session.correlation_id`, wenn das Span-Attribut vorhanden ist; Event-Details können einzelne batchbezogene `trace_id`-Links anbieten. Die dokumentierte Grenze ist Pflicht: Eine Session kann mehrere `trace_id`-Werte haben, `trace_id` ist kein Session-Schlüssel, und `mtrace.session.correlation_id` darf nur bei Single-Session-Batches als Span-Attribut gesetzt werden.
- [ ] RAK-29 ist auch ohne Tempo erfüllt; Tempo erweitert nur Debug-Tiefe.
- [ ] Nur wenn Tempo in `0.4.0` umgesetzt wird: Lokaler Smoke-Test oder manuelle Release-Checkliste erzeugt eine Playback-Session, liest den Suchwert aus API/Dashboard/SQLite (`correlation_id` für Session-Suche oder eine konkrete Event-`trace_id`) und validiert genau diesen Wert in Tempo. "Trace sichtbar" ohne benannten Suchwert reicht nicht als Abnahme.
- [ ] README und `docs/user/local-development.md` unterscheiden klar zwischen eingebauter Session-Timeline und optionalem Tempo.

---

## 7. Tranche 6 — Aggregat-Metriken und Drop-/Invalid-/Rate-Limit-Sichtbarkeit

Bezug: RAK-33; RAK-34; API-Kontrakt §7; Telemetry-Model §2.4/§3/§4.3; Lastenheft §7.9/§7.10.

Ziel: Prometheus bleibt Aggregat-Backend. Die Pflichtmetriken für angenommene, invalid, rate-limited und dropped Events sind sichtbar, korrekt gezählt und cardinality-sicher.

DoD:

- [ ] `mtrace_playback_events_total`, `mtrace_invalid_events_total`, `mtrace_rate_limited_events_total` und `mtrace_dropped_events_total` existieren im Compose-Lab und in Tests.
- [ ] Alle Pflichtcounter zählen Events, nicht Batches; leere Batches, Auth-Fehler und Persistenzfehler folgen den Regeln aus API-Kontrakt §7.
- [ ] Die vier Pflichtcounter tragen keine fachlichen Labels; erlaubte Labels sind nur technische Prometheus-/Target-Labels außerhalb des Metric-Vektors. Insbesondere sind `project_id`, `session_id`, `user_agent`, `segment_url`, `client_ip`, `trace_id`, `correlation_id` und beliebige URL-/Token-Felder auf `mtrace_playback_events_total`, `mtrace_invalid_events_total`, `mtrace_rate_limited_events_total` und `mtrace_dropped_events_total` verboten. Der Cardinality-Smoke fragt jede der vier Serien ab und failt bei jedem nicht-leeren Labelset jenseits der Prometheus-Target-Metadaten; eine spätere `project_id`-Ausnahme bräuchte eine eigene feste Label-Allowlist und bounded-Allowlist-Test.
- [ ] Rate-Limit-Fälle sind mit `429` und Counter-Inkrement testbar.
- [ ] Invalid-Counter-Tests trennen event-zählbare Invalids von frühen Request-Fehlern: Inkrement-Fälle sind mindestens falsche `schema_version` bei nicht-leerem Batch, `events.length > 100` und fehlendes Event-Pflichtfeld in einem nicht-leeren Batch. Null-Inkrement-Fälle sind mindestens leerer Batch (`events.length == 0`), Auth-Fehler, Body-Read-/Payload-Limit-Fehler und malformed JSON, weil dort keine belastbare Event-Anzahl vorliegt oder der Pfad vor dem Use-Case endet. Alle Erwartungen verweisen auf API-Kontrakt §7.
- [ ] Drop-Pfad ist entweder real implementiert und testbar oder die Metrik existiert sichtbar mit `0` und der fehlende Drop-Pfad ist dokumentiert.
- [ ] Grafana-/Prometheus-Lab zeigt die vier Pflichtcounter oder eine dokumentierte Abfrage dafür. RAK-34-Sichtbarkeit ist in Tranche 6 bewusst Prometheus/Grafana-only: Ohne zusätzlich eingeführten API-Status-Summary-Vertrag gibt es kein neues Dashboard-/Session-Status-DoD für Invalid-/Dropped-/Rate-Limit-Zähler. Falls ein API-Status-Summary eingeführt wird, muss §7 ihn mit Response-Shape, Backend-Test und Dashboard-Test explizit aufnehmen; andernfalls bleibt der Tranche-4-Handoff aus §5 als Observability-Lab-Sichtbarkeit geschlossen, nicht als In-App-Statusanzeige.
- [ ] Falls F-40 in Tranche 4 deferred wurde, liefert Tranche 6 die Dashboard-Link-Section für Grafana, Prometheus und Media-Server-Konsole mit dokumentierten Config-Keys, Verhalten bei fehlenden URLs und Dashboard-Test; andernfalls bleibt dieses Item als nicht zutreffend markiert.
- [ ] Cardinality-Smoke prüft, dass neue `0.4.0`-Metriken keine hochkardinalen Labels einführen.

---

## 8. Tranche 7 — Cardinality- und Sampling-Dokumentation

Bezug: RAK-35; RAK-33; RAK-34; Lastenheft §7.10/§7.11; Telemetry-Model §3/§4.4.

Ziel: Nutzer verstehen, welche Daten in Prometheus, OTel/Tempo und SQLite landen, welche Sampling-Strategie gilt und welche Grenzen für produktionsnahe Nutzung bestehen.

DoD:

- [ ] Doku-Synchronisierung ist pro Artefakt zugeschnitten: `spec/backend-api-contract.md` §7 beschreibt nur Pflichtcounter-, Zähl- und Label-Semantik; `spec/telemetry-model.md` beschreibt Prometheus-vs-SQLite-vs-OTel/Tempo, Backend-Spans, Sampling-Auswirkung, allgemeine Cardinality-Regeln und in §4.4 die tatsächlichen SDK-Konfigurationsnamen `sampleRate`, `batchSize`, `flushIntervalMs` und `maxQueueEvents` statt historischer Namen wie `batchMaxEvents`/`batchMaxAgeMs`; `spec/player-sdk.md` und `packages/player-sdk/README.md` beschreiben dieselben SDK-Parameter, den optionalen `traceparent`-Provider und die Timeline-Nachweisgrenze. Falls SDK-Nutzerdoku bewusst nicht geändert wird, muss §8 das ausdrücklich als Nicht-Scope begründen.
- [ ] `docs/user/local-development.md` beschreibt lokale Storage-Retention, SQLite-Reset, Prometheus-Aggregate und den `0.4.0`-Tempo-Status: optionales Tempo-Profil, falls §6 / Tranche 5 umgesetzt wird, oder explizit deferred ohne aktives Profil.
- [ ] `docs/user/demo-integration.md` zeigt, wie eine Demo-Session inklusive Timeline reproduzierbar erzeugt wird.
- [ ] `README.md` aktualisiert den `0.4.0`-Abschnitt mit tatsächlichem Lieferstand.
- [ ] Doku enthält eine klare Tabelle: Prometheus = Aggregate, SQLite = Session-/Event-Historie, OTel/Tempo = Trace-Debugging.
- [ ] Sampling-Grenzen erklären die aktuelle Nachweisgrenze: Das SDK sampelt normale Events eventbasiert, gesampelte Events verbrauchen keine `sequence_number`, und der Server kann fehlende Events deshalb nicht automatisch als Lücke erkennen. Vollständige Timeline-Abnahme und E2E-Smokes laufen mit `sampleRate=1`. Für `sampleRate < 1` ist Vollständigkeit ohne neues session-/batch-skopiertes Sampling-Metadaten-Signal nicht beweisbar; entweder führt eine spätere Tranche ein solches Read-Shape-/Dashboard-Signal ein, oder das Dashboard markiert Sampled-Sessions nur über dokumentierte Konfiguration/Benutzerhinweis, nicht durch serverseitige Lückenerkennung.
- [ ] Datenschutz- und Cardinality-Hinweise trennen die Ebenen: Die vier Pflichtcounter aus `spec/backend-api-contract.md` §7 tragen keine fachlichen Metric-Vector-Labels. Andere `mtrace_*`-Metriken dürfen nur bounded Labels aus der Allowlist in `spec/telemetry-model.md` verwenden; die Allowlist muss bestehende bounded Labels wie `outcome` und `code` für `mtrace_analyze_requests_total{outcome,code}` sowie das bereits ausgelieferte OTel-Attribut `batch.size` auf `mtrace.api.batches.received` behandeln. Für Prometheus ist die normalisierte Label-Schreibweise separat zu dokumentieren (z. B. `batch_size` auf `mtrace_api_batches_received`, nicht `batch.size`). Für die Batchgrößen-Dimension muss §8 ausdrücklich entscheiden: entfernen, vor Export bucketen/clampen, oder erst nach Größenvalidierung als bounded Integerdomäne erlauben. Weil die Telemetrie aktuell vor `MaxBatchSize`-Validierung mit `len(in.Events)` läuft, darf eine Erlaubnis als bounded Label nur bestehen, wenn abgelehnte `events.length > 100` keine freie Labeldomäne erzeugen; Tests müssen den Reject-Pfad abdecken. Mindest-Verbote über alle Prometheus-Metriken hinweg sind `session_id`, URLs/URL-Teile, User-Agent, Client-IP, `viewer_id`, `request_id`, `trace_id`, `span_id`, `correlation_id`, beliebige `project_id`-Labels ohne bounded Allowlist und Token-/Credential-Felder; kürzere Beispiel-Listen reichen nicht als Abnahme.
- [ ] Release-Notes-Vorlage im `CHANGELOG.md`-Unreleased-Abschnitt enthält die neuen Trace-, Storage-, Dashboard-/Session-Timeline-, Metrik- und Doku-Punkte.

---

## 9. Tranche 8 — Release-Akzeptanzkriterien `0.4.0`

Bezug: RAK-29..RAK-35; `docs/user/releasing.md`.

DoD:

- [ ] **RAK-29** Player-Session-Traces werden konsistent und Tempo-unabhängig erzeugt: mehrere Batches einer Session teilen lokal persistierte Korrelationsdaten; Tests decken Erfolg, fehlenden Kontext und deaktiviertes Tempo-Profil ab.
- [ ] **RAK-30** Manifest-Requests, Segment-Requests und Player-Events werden über `correlation_id` einer gemeinsamen Session-Timeline zugeordnet; `trace_id` ist nur optionale Tempo-Vertiefung. Timeline-only- und Browser-Degradationsfälle sind je Ereignistyp begründet, sichtbar und getestet.
- [ ] **RAK-31** Tempo kann optional als Trace-Backend verwendet werden oder ist bewusst als Kann-Scope deferred, ohne Muss-Kriterien zu gefährden.
- [ ] **RAK-32** Dashboard kann Session-Verläufe ohne Tempo anzeigen; API-Restart verliert bestehende lokale Session-Historie nicht.
- [ ] **RAK-33** Prometheus bleibt auf aggregierte Metriken beschränkt; Cardinality-Smoke ist grün.
- [ ] **RAK-34** Dropped-, Rate-Limited- und Invalid-Event-Metriken sind sichtbar und testbar.
- [ ] **RAK-35** Dokumentation beschreibt Cardinality-Grenzen und Sampling-Strategie.
- [ ] Versionen sind konsistent: Root- und Workspace-Pakete tragen `0.4.0`; SDK/Event-Schema-Kompatibilitätscheck bleibt grün. Insbesondere `PLAYER_SDK_VERSION` in `packages/player-sdk/src/version.ts` ist auf `0.4.0` gehoben (aus §3.3-Review, Anmerkung #9: aktuell noch `0.3.0`, weil §3.3 absichtlich keinen Release-Bump macht).
- [ ] `CHANGELOG.md` enthält den Versionsabschnitt `[0.4.0] - <Datum>` mit Trace-, Persistenz-, Dashboard-, Metrik- und Doku-Lieferstand.
- [ ] Release-Gates grün: `make gates` (enthält `test`, `lint`, `coverage-gate`, `arch-check`, `schema-validate`, `docs-check`), `make build`, `make sdk-performance-smoke`, `make smoke-observability`, `make smoke-cli`, `make smoke-analyzer` und `make browser-e2e`.
- [ ] Browser-E2E-Smoke (`make browser-e2e`) erzeugt eine Test-Session und prüft Session-Timeline/Tranche-4-Dashboard-Flows; falls er aus Umgebungsgründen manuell ersetzt wird, ist das Ergebnis als Release-Gate dokumentiert. Der Smoke darf `/demo` nutzen, muss aber bei späterer Demo-Änderung auf einen dedizierten Test-Harness umstellbar bleiben.
- [ ] `docs/planning/in-progress/roadmap.md` markiert `0.4.0` als abgeschlossen und verschiebt den aktiven Fokus auf `0.5.0`.

---

## 10. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`, Commit-Hash anhängen.
- Neue Findings in der `0.4.0`-Phase landen entweder in dieser Datei oder in `risks-backlog.md`.
- Lastenheft-Patches während `0.4.0` werden in `plan-0.1.0.md` Tranche 0c als neue §4a.x-Einträge ergänzt.
- Beim Release-Bump `0.4.0` → `0.5.0`: dieses Dokument als historisch archivieren; Lieferstand wandert dokumentarisch nach `CHANGELOG.md`.
