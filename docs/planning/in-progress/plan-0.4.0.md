# Implementation Plan — `0.4.0` (Erweiterte Trace-Korrelation)

> **Status**: 🟡 in Arbeit. Tranche 0, Tranche 1 (§2.1–§2.6), Tranche 2 (§3.1–§3.4 inkl. §3.4c-Closeout), **Tranche 3 (§4.1–§4.7)**, **Tranche 4 (§5 H1–H6)**, **Tranche 5 (§6.1–§6.5)** und **Tranche 6 (§7.1–§7.4)** vollständig abgeschlossen; Roadmap Schritte 31, 32, 33, 34 und 35 sind auf ✅. Offen: Tranchen 7–8. Tranche 7 (Cardinality- und Sampling-Dokumentation, RAK-35) ist der nächste Schritt.
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
| 2 | Session-Trace-Modell und OTel-Korrelation (siehe §3.1–§3.4) | ✅ |
| 3 | Manifest-/Segment-/Player-Korrelation | ✅ |
| 4 | Dashboard-Session-Verlauf ohne Tempo | ✅ |
| 5 | Optionales Tempo-Profil | ✅ |
| 6 | Aggregat-Metriken und Drop-/Invalid-/Rate-Limit-Sichtbarkeit | ✅ |
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
| §3.4c | Doku-/Roadmap-Closeout | Spec-Drift geschlossen, Roadmap Schritt 31 ✅, Tranche 2 als abgeschlossen markierbar | ✅ (`52026f5`, `851fb59`, `67e54a4` + Plan-/Roadmap-Update) |

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
- **Tranche-3-Blocker (historisch — Gate ist passiert):** Während §3.4c offen war, blockierten alle dortigen DoD-Items den Start von Tranche 3, den Status-Header, Roadmap Schritt 31 und die Tranche-2-Matrix; einzelne bereits abgeschlossene §3.1–§3.4b-Pfade hoben diesen Blocker nicht auf. Mit Closeout-Commit `637985b` ist §3.4c vollständig abgehakt und das Gate ist passiert. Restgrenzen, die das Gate **nicht** geöffnet haben (sondern bewusst nach Tranche 3+ verschoben sind): echte `0.3.x`-Routing-/Middleware-Pipeline-Verifikation (Item #5 Smoke-Scope-Downgrade) und R-6-Mitigation (Item #10 — wird bei beobachtetem Mismatch release-blocking, blockiert aber den Start von Tranche 3 nicht).

DoD:

- [x] Header-Casing/Whitespace-Vertrag (aus §3.3-Review, Anmerkung #5) ist in `spec/backend-api-contract.md` §1 und in Backend-Tests synchronisiert: Header-Name case-insensitiv; Header-Wert-Verhalten für führende/abschließende OWS ist exakt das implementierte Verhalten — Go's `net/textproto` (`net/http`-Wire-Layer) entfernt OWS auf beiden Seiten am Header-Wert, bevor er das Backend erreicht, sodass ein OWS-umschlossener, sonst valider Wert als gültig akzeptiert wird; das Backend führt selbst kein zusätzliches Trim durch und ein durchgereichter OWS-Wert fällt am defensiven `len == 55`-Check als parse_error. SDK schreibt lowercased `traceparent`. Test-Anker: `TestHTTP_Span_TraceParent_LeadingTrailingWhitespace` (Wire-Beobachtung) und neue OWS-Cases in `TestParseTraceParent_Invalid` (Defense-in-Depth bei Direktaufruf) (`52026f5`).
- [x] `mtrace.project.id` ist nicht mehr als Drift geführt: `spec/telemetry-model.md` §2.5 dokumentiert es als Pflicht für accepted Batches und nach erfolgreicher Project-Auflösung sowie als bewusst unset bei Rejects vor Project-Auflösung; Test-Anker `TestHTTP_Span_SingleSessionBatch_SetsCorrelationID` (`67e54a4`).
- [x] `spec/telemetry-model.md` §2.5 ist final konsistent mit Code: Hybrid-Strategie, ein Server-Span pro Batch, Persistenzquelle pro Feld (`trace_id` Event-Spalte, `correlation_id` Session+Event), Time-Skew nur als Span-Attribut (`mtrace.time.skew_warning`), Sampling-Auswirkung und Prometheus-Cardinality-Grenzen stehen in einem zusammenhängenden Abschnitt; mtrace.project.id- und session_id-Edits aus `67e54a4` schließen die letzten Drifts (`67e54a4`).
- [x] `session_id`-Span-Attribut-Verbot ist konsistent dokumentiert und getestet: `spec/telemetry-model.md` §2.1 und §2.5 sagen verbindlich, dass der Server in **keinem** OTel-Span `session_id`/`mtrace.session.id`/`mtrace.session_id` setzt — Single-Session-Suche läuft ausschließlich über `mtrace.session.correlation_id`; historische Zulässigkeitsaussagen sind auf `0.1.x` begrenzt. Test-Anker: `TestHTTP_Span_DoesNotSetSessionIDAttribute` (`67e54a4`).
- [x] Cross-Version-Kompatibilität ist scope-korrekt geschlossen: Abnahmetext bleibt explizit auf den bisher gelieferten Smoke-Scope (`TestHTTP_Trace_CrossVersion_LegacyHandlerAcceptsTraceParent` + SDK-Mock-Server in `http-transport.test.ts`, `6fdc8d0`); echte `0.3.x`-Routing-/Middleware-Pipeline-Kompatibilität bleibt **explizit als Restgrenze** dokumentiert und ist kein bewiesener Tranche-2-Claim. Belastbare 0.3.x-Verifikation kann ein späterer Reproducer-Job (Container-Image des `v0.3.0`-API-Builds) liefern; bis dahin gilt der Smoke-Scope.
- [x] `spec/backend-api-contract.md` §1 / §3 / §3.7 reflektiert das ausgelieferte Header-Verhalten und die neuen Read-Felder: §1 trägt den OWS-/Casing-Vertrag aus `52026f5`; §3.7 listet `trace_id` (nullable, batch-bezogen) und `correlation_id` (pro Session stabil) als server-vergebene Read-Felder; `span_id` ist kein Read-Feld; `GET /api/stream-sessions` exponiert keine Session-`trace_id`; Fehlerklassifikation bleibt 202/4xx aus dem normalen Event-Vertrag (`52026f5`, ergänzt durch §3.7 aus `c3741aa`).
- [x] Tempo-deaktivierter Produktivstart ist als Closeout-Gate abgesichert: `TestSetup_BlankOTelEnv_FallsBackToNoopExporter` in `apps/api/adapters/driven/telemetry/otel_test.go` deckt den `cmd/api`-Auflösungsweg, in dem Compose `OTEL_TRACES_EXPORTER`/`OTEL_METRICS_EXPORTER`/`OTEL_EXPORTER_OTLP_ENDPOINT`/`OTEL_EXPORTER_OTLP_PROTOCOL` als blank-strings durchreicht (kein aktives Tempo-Profil). Der Test setzt die vier Variablen explizit auf `""`, ruft `telemetry.Setup` und beweist (a) `unsetBlankOTelEnv` hat sie nach Setup entfernt, (b) `TracerProvider` ist eine echte SDK-Instanz und mintet eine valide `trace_id`, (c) `ForceFlush` liefert `nil`, also pusht der Exporter nicht gegen einen nicht-existierenden OTLP-Collector. Der §3.4a-Router-Test (`TestHTTP_Trace_NoopTracer_CorrelationStillPersisted`) bleibt komplementär für den `nil`-Tracer-Adapter-Fallback (`851fb59`).
- [x] Legacy-Korrektheitsgrenze ist final dokumentiert: §3.4c entscheidet ausdrücklich **gegen** ein historisches Backfill für vor §3.2 geschriebene `playback_events.correlation_id`; Self-Healing gilt nur für `stream_sessions.correlation_id` und neu persistierte Events ab dem nächsten Batch. `spec/backend-api-contract.md` §3.7 (Migrations-Hinweis) und `spec/telemetry-model.md` §2.5 (Legacy-Grenze) nennen leere Legacy-Event-`correlation_id`s als degradierten Read-Fall, nicht als Vertragsbruch.
- [x] `spec/player-sdk.md` bleibt synchron zum tatsächlichen SDK-Verhalten aus `f7dcdb9`: der Trace-Korrelations-Abschnitt zeigt eine explizite Tabelle für die fünf Provider-Rückgabe-/Verhaltensfälle (nicht-leerer String → 1:1-Header, `undefined`/`""` → Opt-out ohne Warn, Non-String/Throw → kein Header + ein `console.warn` pro `HttpTransport`-Instanz) inkl. `HttpTransportOptions.silent` für Tests; Provider wird pro Send synchron aufgerufen, kein Caching (`67e54a4`).
- [x] R-6 (`correlation_id`-Race) ist im Plan-Closeout referenziert: bleibt **nicht blockierend** für Tranche 2, ist aber explizite Restgrenze für Tranche 8 / operative Beobachtung. Eintrag in `docs/planning/open/risks-backlog.md` §1 R-6 bleibt offen; falls vor Tranche 8 ein Mismatch zwischen `stream_sessions.correlation_id` und `playback_events.correlation_id` beobachtet wird (Lab oder produktiver Run), wird R-6 release-blocking und Tranche 3 darf nicht ohne Mitigation starten.
- [x] Test-/Review-Gate für den Closeout ist dokumentiert: ein §3.4c-Closeout-Review muss mindestens `make ts-lint`, `make api-test` (Backend-Test-Slice für §3.4a/§3.4b/§3.4c-Items inkl. Trace-Span-Tests und der OWS-/Tempo-disabled-Tests aus `52026f5`/`851fb59`/`67e54a4`) und die SDK-Test-Slice `packages/player-sdk/tests/http-transport.test.ts` ausführen. Reiner `make api-test`/`make ts-test` allein reicht nicht — §3.3 hat gezeigt, dass Snapshot- und Spec-Drift erst durch `make ts-lint` auffallen.
- [x] `docs/planning/in-progress/roadmap.md` Schritt 31 ist auf ✅ gesetzt; Status-Header und Verweis auf den Tranche-2-Closeout-Stand sind aktualisiert (siehe Roadmap-Update-Commit).
- [x] Dieser Plan ist nach dem Closeout aktualisiert: Status-Header oben nennt Tranche 2 vollständig abgeschlossen, Tabelle in §1 markiert Tranche 2 ✅, §3.4c-DoD-Items tragen Commit-Hash, und Tranche 3 bleibt der nächste offene Trace-Korrelationsschritt.

---

## 4. Tranche 3 — Manifest-/Segment-/Player-Korrelation

Bezug: RAK-30; RAK-29; Stream Analyzer aus `0.3.0`; F-68..F-81; Telemetry-Model §1.

Ziel: Manifest-Requests, Segment-Requests und Player-Events werden einer gemeinsamen Session-Timeline zugeordnet. Normative Priorität: `correlation_id` ist der primäre Zuordnungsschlüssel und Source-of-Truth für Dashboard/API; eingehende SDK-Events liefern weiter `session_id`, aus der der Server die `correlation_id` resolved. `trace_id` bleibt optionaler, batchbezogener Debug-Kontext für Tempo und darf keine Timeline-Zuordnung allein tragen. RAK-30 ist Soll; Lücken müssen sichtbar, testbar und erklärbar bleiben.

Tranche 3 ist in sieben aufeinander aufbauende Sub-Tranchen geschnitten: §4.1 fixiert die Spec-Grundlagen vor jedem Code (Wire-Format, Reason-Enum, `session_boundaries[]`-Wrapper, Cursor v3, URL-Redaction-Matrix, endpoint-spezifische Auth, `{analysis, session_link}`-Hülle); §4.2 hebt Schema, Ports, InMemory-/SQLite-Adapter und den Application-Resolver auf projekt-skopierte Session-Zugriffe und schließt den R-6-Race; §4.3 zieht die HTTP-Read-Auth für Session-/Event-Reads nach und migriert Cursor von v2 auf v3; §4.4 baut die Backend-Verarbeitung der Netzwerkereignisse, persistiert `session_boundaries[]` und führt URL-Redaction vor Persistenz aus; §4.5 modelliert Analyzer-Linking im Hexagon und liefert die `{analysis, session_link}`-Hülle inklusive endpoint-spezifischer Auth und OPTIONS-Preflight; §4.6 baut die SDK-seitige hls.js-Manifest-/Segment-Capture und den Boundary-Send; §4.7 schließt Tests, Doku und den Tranche-3-Closeout-Marker einschließlich R-6-Status.

Liefer-/Abnahme-Matrix:

| Sub-Tranche | Ergebnis | Harte Abnahme | Status |
|---|---|---|---|
| §4.1 | Spec-Vorarbeit (Doku-only) | Wire-Format, Reason-Enum, Boundary-Wrapper, Cursor v3, URL-Redaction-Matrix, Auth-Tabelle und Analyzer-Hülle ohne implizite Code-Entscheidungen in `spec/*` und `contracts/*` festgeschrieben | ✅ (`caf6e6e` + Spec-Stand aus §3.x) |
| §4.2 | Storage-/Repository-Prerequisite + R-6 | `(project_id, session_id)`-Schlüssel in Schema, Ports, InMemory- und SQLite-Adapter; Cross-Project-Kollisionstest grün; R-6-Race technisch unmöglich (DB-finale `correlation_id`-Rückgabe) | ✅ (`f50bd46`, `949a265`, `c602ebf`) |
| §4.3 | HTTP-Read-Auth + Cursor v3 | `GET /api/stream-sessions*` tokenpflichtig; Cursor v3 mit Project-/Session-Scope, v2 wird als `cursor_invalid_legacy` abgewiesen; Dashboard-Client-API und Read-Doku nachgezogen | ✅ (`f50bd46`, `8812104`) |
| §4.4 | Backend-Network-Events + URL-Redaction | `network.*`-Meta-Keys typvalidiert; `session_boundaries[]` persistiert und restart-stabil; URL-Redaction vor Persistenz für alle URL-verdächtigen Meta-Keys | ✅ (`e7ac534`, `d7aaaad`) |
| §4.5 | Analyzer-Linking + endpoint-spezifische Auth | `AnalyzeManifestResult{Analysis, SessionLink}` im Use-Case; Statusmatrix (`linked`/`detached`/`not_found_detached`/`conflict_detached`) grün; OPTIONS-Preflight für `/api/analyze`; Breaking-Change-Hülle `{analysis, session_link}` in Tests und Nutzer-Doku nachgezogen | ✅ (`175b24c`, `85096a6`) |
| §4.6 | SDK-Manifest-/Segment-Capture | hls.js-Mapping-Tabelle festgeschrieben und getestet; jedes Manifest-/Fragment-Signal wird als `manifest_loaded`/`segment_loaded` erfasst; Dedup-Schlüssel nutzt keine URLs; optionaler `session_boundaries[]`-Send | ✅ (`ee61b46`, `4b597a9`) |
| §4.7 | Tests, Doku und Tranche-3-Closeout | gemischte Player-/Manifest-/Segment-Korrelation getestet; Degradationsmatrix (Browser-/Native-Limitierung, CORS-/Resource-Timing) getestet; Tranche-3-Closeout-Doku, Roadmap Schritt 32 ✅, R-6 als technisch geschlossen markiert | ✅ (`eddfedd`, `daefad5`) |

Abnahmegrenzen:

- **Mindestkorrelation:** Jedes neu vom SDK erzeugte Manifest-, Segment- und Player-Ereignis muss im Ingest-Payload dieselbe technische Session-Partition (`project_id` + `session_id`) wie die zugehörigen Player-Events tragen. Der bestehende Server-Resolver erzeugt oder liest daraus die Session-`correlation_id`; nach Persistenz muss jedes akzeptierte Event dieselbe `playback_events.correlation_id` wie `stream_sessions.correlation_id` tragen. Eingehende Events ohne `session_id` bleiben normale `422`-Validierungsfehler nach API-Kontrakt §5; eine vom Client gelieferte `correlation_id` ist im POST-Wire-Format nicht zulässig und wird nicht als Source-of-Truth akzeptiert. Fehlende Browser-Timing-/Resource-Daten dürfen Detailfelder reduzieren, aber nicht die Session-Zuordnung entfernen.
- **Storage-/Repository-Prerequisite:** Der aktuelle `stream_sessions`-Store ist vor Tranche 3 `session_id`-global (`primary_key: [session_id]`, Repository-Reads per `Get(ctx, session_id)`, `List` ohne `project_id`-Filter). Tranche 3 muss Schema, Ports, Repository-Methoden, Resolver, InMemory-Adapter, SQLite-Adapter und HTTP-Read-Handler auf projekt-skopierte Session-Zugriffe heben: eindeutiger Key `(project_id, session_id)`; alle `List`/Cursor-, `Get`/Detail-, Event-Read-, `Upsert`-/Resolver- und Analyzer-Linking-Pfade nehmen `project_id` entgegen und filtern danach. Eine Beweisoption über global eindeutige client-gelieferte `session_id`s ist nicht zulässig. Cross-Project-Kollisionstest ist Pflicht: gleiche `session_id` in zwei Projekten darf weder Session-Detail noch Event-Reads, Cursor-Pagination oder Analyzer-Linking projektübergreifend vermischen.
- **R-6-Prerequisite:** Der bekannte `correlation_id`-First-Insert-Race aus `docs/planning/open/risks-backlog.md` R-6 muss vor Abschluss von Tranche 3 behoben oder durch DB-finale Correlation-ID-Rückgabe nachweislich unmöglich gemacht werden. Bis dahin gilt die harte Gleichheit `playback_events.correlation_id == stream_sessions.correlation_id` nur für nicht-konkurrierende Flows; Tranche 3 darf nicht als vollständig abgeschlossen markiert werden.
- **Analyzer-Linkage-Priorität:** Für explizite Analyzer-Verknüpfung gilt: `correlation_id` gewinnt vor `session_id`, aber nur innerhalb eines gültig aufgelösten Project-Kontexts. Verlinkte `POST /api/analyze`-Requests müssen `X-MTrace-Token` (und falls später aktiv: `X-MTrace-Project`) erfolgreich auf ein `project_id` auflösen. Fehlt dieser Kontext bei gesetzter `correlation_id` oder `session_id`, antwortet die API mit dem Auth-/Kontextfehler aus dem API-Kontrakt und führt keinen Session-Lookup aus; dieser Fall ist kein stilles `detached`. Nur Requests ohne Link-Felder bleiben ohne Project-Kontext erfolgreich und erhalten `session_link.status="detached"`. Alle Link-Lookups laufen über `(project_id, correlation_id)` bzw. `(project_id, session_id)`. `correlation_id` allein ohne Treffer im Project liefert `session_link.status="not_found_detached"`. Wenn beide IDs angegeben sind, muss zuerst `correlation_id` im Project existieren und danach `session_id` im selben Project zur Session mit dieser `correlation_id` auflösen; unbekannte `correlation_id` gewinnt also nicht durch einen bekannten `session_id`-Fallback und liefert `not_found_detached`, Cross-Project-`correlation_id` gilt ebenfalls als `not_found_detached`, Mismatch im selben Project liefert `conflict_detached`. Nur `session_id` allein ist als Fallback für bestehende oder bereits selbst-geheilte Sessions erlaubt; eine unbekannte `session_id` erzeugt durch `POST /api/analyze` keine neue Session und führt zu `session_link.status="not_found_detached"`. Kompatibilitätsentscheidung: Ab Tranche 3 gibt `POST /api/analyze` für **alle** erfolgreichen Requests die Hülle `{analysis, session_link}` zurück, auch ohne Link-Felder; ungebundene Requests erhalten `session_link.status="detached"`. Der Breaking-Change wird im API-Kontrakt und in den Contract-Tests festgeschrieben.
- **Degradationsfälle:** Fehlende oder blockierte Resource-Timing-Daten, CORS-Lücken, Service-Worker-Interception, CDN-Redirects und hls.js-Signale mit unvollständigen Details erzeugen ein Netzwerkevent mit flachen Meta-Keys `meta["network.detail_status"]="network_detail_unavailable"` und optional `meta["network.unavailable_reason"]` gemäß `spec/telemetry-model.md` §1.4. Event-Meta und `session_boundaries[]` verwenden dieselbe kontrollierte Reason-Domäne: `native_hls_unavailable`, `hlsjs_signal_unavailable`, `browser_api_unavailable`, `resource_timing_unavailable`, `cors_timing_blocked`, `service_worker_opaque`; unbekannte Werte oder Werte außerhalb `^[a-z0-9_]{1,64}$` werden mit `422` abgelehnt. Wenn ein Browser-/Native-HLS-Pfad gar kein Manifest-/Segment-Signal liefert, wird kein synthetisches Netzwerkereignis erfunden; stattdessen exponiert der Session-Read-Pfad Tranche 3 ein nicht-eventbasiertes Feld `network_signal_absent` als Liste von Objekten `{ "kind": "manifest" | "segment", "adapter": "hls.js" | "native_hls" | "unknown", "reason": "<machine_reason>" }` im Session-Block. Schreibpfad ist ein optionaler Batch-Wrapper-Block `session_boundaries[]` in `POST /api/playback-events` mit `kind="network_signal_absent"`, `project_id`, `session_id`, `network_kind`, `adapter`, kontrolliertem `reason` und `client_timestamp`; er wird zusammen mit einem normalen Event-Batch gesendet, zählt nicht als Event, besitzt keinen `event_name` und ändert `schema_version: "1.0"` nicht. Pro Batch sind maximal 20 Boundaries zulässig, sie zählen ins Body-Size-Budget des SDK/Backends, und jede Boundary muss eine `(project_id, session_id)`-Partition referenzieren, die mindestens ein Event im selben Batch trägt. Boundary-only-Batches ohne `events` oder Boundaries für fremde/nicht enthaltene Sessions bleiben in Tranche 3 außerhalb des Vertrags und liefern `422`. Das Backend persistiert daraus den Boundary-Record, auch wenn kein Manifest-/Segment-Event existiert. Persistenzvehikel ist eine durable Session-Metadaten-Spalte oder ein äquivalenter session-skopierter Capability-/Boundary-Record; der Wert darf nicht nur aus flüchtigem Prozesszustand abgeleitet werden und muss nach API-Restart identisch lesbar bleiben. Die Dashboard-Sichtbarkeit dieser Grenze ist Tranche-4-Scope. Beide Varianten sind akzeptierte Degradation und müssen getrennt getestet werden.
- **Schema-Entscheidung:** Tranche 3 verwendet ausschließlich additive, flache `meta`-Keys nach dem Muster `network.*`/`timing.*` im bestehenden Event-Wire-Schema `1.0`; verschachtelte `meta.network`-Objekte sind nicht zulässig, weil `packages/player-sdk/src/types/events.ts` für `EventMeta` nur skalare Werte erlaubt. Die Event-Namen bleiben die bereits im Contract/SDK vorhandenen `manifest_loaded` und `segment_loaded`. Keine neuen `event_name`-Werte in Tranche 3; falls ein weiterer Event-Typ nötig wäre, müssen `contracts/event-schema.json`, `packages/player-sdk/src/types/events.ts`, Public-API-Snapshot und Compat-Tests explizit erweitert werden und der Zusatz wird als eigenes DoD-Item geführt. Kein Breaking Change und keine neue Major-Schema-Version in `0.4.0`; falls ein benötigtes Feld nicht additiv modellierbar ist, wird es deferred statt per Migration erzwungen.
- **URL-Datenschutz:** Segment-/Manifest-URLs dürfen weder Prometheus-Labels noch rohe, credential-haltige Persistenzwerte werden. Vor Persistenz/Anzeige gilt die feste Redaction-Matrix für den vorgesehenen URL-Repräsentanten und für alle URL-verdächtigen generischen Meta-Keys (`url`, `uri`, `manifest_url`, `segment_url`, `media_url`, `network.url`, `network.redacted_url`, `request.url`, `response.url` und case-insensitive Varianten). Scheme, Host und nicht-sensitive Pfadsegmente dürfen erhalten bleiben; Query und Fragment werden vollständig entfernt; `userinfo` wird entfernt; signierte/credential-artige Query-Parameter (`token`, `signature`, `sig`, `expires`, `key`, `policy` und case-insensitive Varianten) werden nicht gespeichert; ein Pfadsegment ist tokenartig, wenn es ≥ 24 Zeichen lang ist und mindestens 80 % seiner Zeichen aus `[A-Za-z0-9_-]` bestehen, wenn es ein Hex-String mit gerader Länge mindestens 32 ist, oder wenn es bekannte JWT-/SAS-/Signed-URL-Muster trägt. Tokenartige Pfadsegmente werden ausschließlich durch `:redacted` ersetzt; es wird kein stabiler Hash und kein Gleichheitsmarker persistiert. Unbekannte `meta`-Keys mit String-Werten, die als absolute URL parsebar sind oder `://` enthalten, werden ebenfalls vor Persistenz redigiert oder verworfen; rohe URL-Werte dürfen in keinem Meta-Feld persistieren. Tests decken Query-/Fragment-Redaction, `userinfo`, signierte Query-Parameter, Token-Parameter, JWT-/Base64URL-Pfadsegmente sowie bösartige/Legacy-Payloads mit rohen URLs in generischen Meta-Keys ab.
- **Analyzer-Semantik:** `POST /api/analyze` bleibt in Tranche 3 standardmäßig getrennt von Live-Player-Sessions. Eine Verknüpfung mit einer Session ist nur zulässig, wenn die Request-Seite explizit eine vorhandene `correlation_id` oder, als Fallback, `session_id` übergibt; andernfalls wird das Analyzer-Ergebnis als unabhängige Manifestanalyse angezeigt und nicht in die Player-Timeline gemischt.

### 4.1 Spec-Vorarbeit (Doku-only, kein Code)

Bezug: Telemetry-Model §1; API-Kontrakt §1, §3, §4, §5, §10; `contracts/event-schema.json`.

Ziel: Vor Code-Änderungen sind Wire-Format-Erweiterung, Reason-Enum, `session_boundaries[]`-Wrapper, Cursor-v3-Vertrag, URL-Redaction-Matrix, endpoint-spezifische Auth-Tabelle und `{analysis, session_link}`-Antworthülle verbindlich entschieden, sodass §4.2–§4.7 ohne implizite Spec-Entscheidungen umgesetzt werden können. Sub-Tranchen-Ausgang: keine Code-Diffs, aber alle nachfolgenden Sub-Tranchen können auf eindeutige Spec-Aussagen verweisen.

DoD:

- [x] `contracts/event-schema.json` (`reserved_meta_keys`), `spec/backend-api-contract.md` §3 und `spec/telemetry-model.md` §1.4 dokumentieren die additive Erweiterung des Event-Wire-Schemas `1.0` für Netzwerkereignisse: bestehende Event-Namen `manifest_loaded`/`segment_loaded`, neue reservierte `meta`-Keys `network.kind`, `network.detail_status`, `network.unavailable_reason`, `network.redacted_url` sowie `timing.*`-Keys mit dokumentierten Wertebereichen und Konflikt-Regeln (`network.unavailable_reason` nur bei `network.detail_status="network_detail_unavailable"`; `available` + `unavailable_reason` → `422`). `EventMetaValue` bleibt skalar; Objekte/Arrays/rohe URLs in reservierten Keys → `422` vor Persistenz (Spec-Stand wurde bereits während §3.x gepflegt; Verweise konsolidiert in `caf6e6e`).
- [x] `network.unavailable_reason`- und `session_boundaries[].reason`-Domäne hat einen einzigen normativen Anker in `spec/telemetry-model.md` §1.4 (Tabelle zu `meta["network.unavailable_reason"]`); `spec/backend-api-contract.md` §3.4, der zweite §1.4-Absatz zu `session_boundaries[].reason` und `contracts/event-schema.json` (`session_boundaries.reasons_ref`, `session_boundaries.reason_pattern_ref`, `network_unavailable_reasons_anchor`) verweisen auf diesen Anker statt eigene Listen zu führen. Pattern `^[a-z0-9_]{1,64}$` bleibt zentral. (`caf6e6e`)
- [x] `session_boundaries[]`-Batch-Wrapper-Block ist in `spec/backend-api-contract.md` §3.4, `spec/telemetry-model.md` §1.4 und `contracts/event-schema.json` (`batch.session_boundaries_field`, `session_boundaries`-Block) typisiert: optionaler Block neben `events[]`, max 20 pro Batch, Pflichtfelder `kind="network_signal_absent"`, `project_id`, `session_id`, `network_kind`, `adapter`, `reason`, `client_timestamp`; jede Boundary muss `(project_id, session_id)` einer im selben Batch enthaltenen Session referenzieren; Boundary-only-Batches liefern `422` (Spec-Stand aus §3.x; SDK-Public-API-Typ folgt im §4.6-Closeout).
- [x] Cursor-v3-Vertrag ist in `spec/backend-api-contract.md` §10.3 festgeschrieben: List-Cursor enthalten Project-Scope (`project_id` oder Scope-Hash), Event-Cursor enthalten Collection-Scope `(project_id, session_id)` oder Scope-Hash; v2-Cursor → `cursor_invalid_legacy`; v3-Cursor mit fremdem Scope → `cursor_invalid_malformed`. Tranche-3-Aktivierung läuft über den HTTP-Read-Auth-Refactor in §4.3 (Spec-Stand aus §3.x).
- [x] URL-Redaction-Matrix ist in `spec/telemetry-model.md` §1.4 festgeschrieben (betroffene Keys, Erhaltungs-/Entfernungs-Regeln, Token-Heuristik mit ≥24-Zeichen-/80%-Charset-/Hex-/JWT-/SAS-Mustern, kein stabiler Hash, unbekannte URL-parsbare Meta-Keys redigiert oder verworfen) und wird in `spec/backend-api-contract.md` §3 referenziert; `spec/architecture.md` und `docs/user/local-development.md` bekommen den Schliff-Verweis im §4.7-Closeout (Spec-Stand aus §3.x).
- [x] `spec/backend-api-contract.md` §1 und §4 sind auf endpoint-spezifische Auth bereinigt: §1 nennt `X-MTrace-Token` nur als „Auth-Token; Pflicht je Endpoint gemäß §4"; §4 listet pro Endpoint, dass `POST /api/playback-events` und Session-Reads ab Tranche 3 tokenpflichtig sind, während `POST /api/analyze` nur bei gesetzter `correlation_id` oder `session_id` tokenpflichtig ist und ungebundene Analyze-Requests ohne Token erfolgreich `session_link.status="detached"` liefern. Pflicht-Test-Fälle sind tabellarisch verankert; Code-Tests folgen in §4.5 (Spec-Stand aus §3.x).
- [x] `{analysis, session_link}`-Antworthülle ist im API-Kontrakt §3.6 als Tranche-3-Breaking-Change festgeschrieben: alle erfolgreichen `POST /api/analyze`-Antworten tragen die Hülle, auch detached. Statusmatrix (`linked`, `detached`, `not_found_detached`, `conflict_detached`) ist in §3.6 erklärt; Link-Auflösung läuft über `(project_id, correlation_id)` bzw. `(project_id, session_id)` mit der dokumentierten Priorität (`correlation_id` zuerst). Nutzer-/Smoke-Doku-Schliff (`docs/user/stream-analyzer.md`) folgt mit Code in §4.5; interne `spec/contract-fixtures/analyzer/*` bleiben flach (Spec-Stand aus §3.x).
- [x] Doku-Grenzen sind in `spec/telemetry-model.md` §1.4 als bekannte Korrelations-Lücken benannt (Browser-APIs, CORS, Resource Timing, Service Worker, CDN-Redirects, Native HLS, Sampling); `correlation_id` ist Pflichtkontext, `trace_id` optionale Debug-Vertiefung. Endgültiger Doku-Schliff (Verweise in `docs/user/local-development.md`, README) folgt im §4.7-Closeout (Spec-Stand aus §3.x).

### 4.2 Storage-/Repository-Prerequisite und R-6

Bezug: §4.1; ADR-0002 §8.1; Risiken-Backlog R-6; Telemetry-Model §2.5.

Ziel: Schema, Ports, InMemory-/SQLite-Adapter und der Application-Resolver sind auf den projekt-skopierten Schlüssel `(project_id, session_id)` gehoben; der bekannte R-6-Race ist durch DB-finale `correlation_id`-Rückgabe technisch unmöglich. Sub-Tranchen-Ausgang: alle Read-/Write-Pfade können `project_id` führen, ohne dass §4.3–§4.7 dafür weitere Storage-Refactors brauchen.

DoD:

- [x] Storage-/Repository-Prerequisite ist implementiert: Schema/Migration (`V2__project_session_pk.sql`) stellt projekt-skopierte Session-Eindeutigkeit her, Port-Signaturen (`Get(ctx, projectID, sessionID)`, neue `GetByCorrelationID(ctx, projectID, correlationID)`, `SessionListQuery.ProjectID`, `EventListQuery.ProjectID`) und Application-Resolver nehmen `project_id`, InMemory- und SQLite-Adapter nutzen `(project_id, session_id)` für `List`/Cursor, `Get`/Detail, Event-Reads, `Upsert`/Self-Healing und Analyzer-Linking. Legacy-Leerwerte (`correlation_id=""`) liefern in `GetByCorrelationID` keinen Treffer; der Composite-PK macht Duplikate innerhalb eines Projects unmöglich. Read-Handler lösen Project-Kontext aus `X-MTrace-Token` auf (vorgezogen aus §4.3), damit nach C1 jeder Commit lauffähig bleibt (`f50bd46`).
- [x] Cross-Project-Kollisionstest deckt InMemory und SQLite ab via `testCrossProjectSessionIsolation` im gemeinsamen Contract-Test: dieselbe `session_id` in zwei Projekten erzeugt getrennte Sessions und unterschiedliche `correlation_id`s; `List`, `Get`, `ListBySession` und `GetByCorrelationID` lösen nicht projektübergreifend auf — `GetByCorrelationID(B, corrA)` und `GetByCorrelationID(A, corrB)` liefern `ErrSessionNotFound` (`c602ebf`).
- [x] R-6 ist technisch geschlossen: `UpsertFromEvents` liefert die DB-finale `correlation_id` jeder Session (Map-Rückgabe `[sessionID]canonicalCID`), der Use-Case enricht damit die Events vor `EventRepository.Append` (Reihenfolge umgedreht: Sessions zuerst, danach Events). SQLite-Adapter prüft `RowsAffected()` nach dem `ON CONFLICT (project_id, session_id) DO NOTHING`-Insert und liest die Sieger-CID nach, falls der eigene Insert zur No-op wurde. Race-Test `TestUpsertFromEvents_RaceCanonicalCorrelationID` (8 Goroutines, gleiche `(project, session)`, unterschiedliche Kandidat-CIDs) zeigt: alle Aufrufe liefern dieselbe Sieger-CID, eine Zeile in `stream_sessions`, kein Mismatch zwischen `playback_events.correlation_id` und `stream_sessions.correlation_id`. R-6-Eintrag im Risiken-Backlog wird in §4.7 als technisch geschlossen markiert (`949a265`).

### 4.3 HTTP-Read-Auth und Cursor-v3-Migration

Bezug: §4.1 (Cursor-v3-Vertrag, Auth-Tabelle); §4.2 (Repository-Signaturen); API-Kontrakt §4, §10.3.

Ziel: HTTP-Read-Handler lösen Project-Kontext aus `X-MTrace-Token` auf, reichen `project_id` in die Application-Pfade weiter und nutzen ausschließlich Cursor v3. Sub-Tranchen-Ausgang: `GET /api/stream-sessions*` ist tokenpflichtig, Dashboard-Client-API ist nachgezogen, v2-Cursor brechen kontrolliert.

DoD:

- [x] HTTP-Read-Handler lösen Project-Kontext vor Session-List, Session-Detail und Event-Read auf und reichen `project_id` in die Application-/Repository-Pfade weiter; Session-List, Session-Detail und Event-Read verlangen ab Tranche 3 `X-MTrace-Token`. Fehlender oder ungültiger Token liefert `401`; positive Tests mit gültigem Token beweisen projekt-skopierte Ergebnisse (Auth-Anteil vorgezogen aus §4.2 C1, `f50bd46`; Cursor-Aktivierung ergänzt in `8812104`).
- [x] Dashboard-/Client-API sendet `X-MTrace-Token` für alle `GET /api/stream-sessions*`-Aufrufe; Nutzer-/Smoke-Doku zeigt den Header in Read-Beispielen (`docs/user/local-development.md`, Smoke-Skripte) (`8812104`).
- [x] Session-List- und Event-Cursor wechseln auf `cursor_version: 3`: List-Cursor enthalten den Project-Scope, Event-Cursor den Collection-Scope `(project_id, session_id)`. v2-Cursor werden nach Aktivierung der projekt-skopierten Read-Pfade als `cursor_invalid_legacy` abgewiesen; v3-Cursor mit fremdem Project- oder Session-Scope liefern `cursor_invalid_malformed`. Tests decken Cursor-Round-Trip und beide Fehlerklassen ab (`cursor_internal_test.go`, `8812104`).

### 4.4 Backend-Network-Events und URL-Redaction

Bezug: §4.1 (Wire-Format, Reason-Enum, Boundary-Wrapper, Redaction-Matrix); §4.2 (Repository-Signaturen).

Ziel: Backend validiert und persistiert die neuen `network.*`-Meta-Keys und den `session_boundaries[]`-Block, redigiert URLs vor Persistenz und reicht jedes Netzwerkereignis mit derselben `correlation_id` wie die zugehörigen Player-Events durch. Sub-Tranchen-Ausgang: Backend kann gemischte Player-/Manifest-/Segment-Batches aus dem SDK akzeptieren, ohne dass §4.6 oder §4.7 weitere Server-Validierung nachziehen müssen.

DoD:

- [x] Event-Schema-Validierung ist umgesetzt: reservierte `network.*`- und `timing.*`-Keys werden inbound typvalidiert (Domänen aus §4.1); Objekte, Arrays, freie Strings außerhalb der Domäne oder rohe URLs in reservierten Keys führen zu `422` vor Persistenz. `network.unavailable_reason` nur bei `detail_status="network_detail_unavailable"`; bei `available` → `422`. Contract-Tests sichern Forward-Compat (alte Backends ignorieren unbekannte additive Keys) und Read-Pfad neuer Backends (`event_meta_validation_internal_test.go`, `register_playback_event_batch_meta_test.go`, `e7ac534`).
- [x] `network.unavailable_reason` und `session_boundaries[].reason` verwenden dieselbe Reason-Enum und dasselbe Pattern/Längenlimit aus §4.1; Contract-Tests decken alle zulässigen Reason-Werte sowie unbekannte/gefährliche Werte in Event-Meta und Boundary-Block ab (`session_boundary_validation.go` teilt `isNetworkUnavailableReason` und `networkUnavailableReasonPattern` mit der Event-Validation; `register_playback_event_batch_boundaries_test.go`, `d7aaaad`).
- [x] `session_boundaries[]`-Block wird vor jedem Write atomar validiert oder gemeinsam transaktional persistiert: ein invalider Boundary-Block liefert `422`, persistiert weder Events noch Boundaries und erhöht `accepted` nicht. Persistenzvehikel ist eine durable `stream_session_boundaries`-Tabelle (V3-Migration); nach API-Restart identisch lesbar (`restart_test.go::TestRestartPreservesSessionBoundaries`). Contract-Tests decken: Batch mit Events plus Boundary-Block (`AcceptsBoundariesAlongsideEvents`), invaliden Boundary-Block mit sonst validen Events (`RejectsInvalidBoundary_NoPersist`, alle acht Negativklassen), `>20` Boundaries (`RejectsTooManyBoundaries`), Boundaries für nicht im Batch enthaltene Sessions (`RejectsBoundaryForUnknownSession`), Boundary-only-Batches (`BoundaryOnlyEmptyEventsStillFails`); SDK-Boundary-Send selbst ist §4.6-Scope (`d7aaaad`).
- [x] Backend normalisiert Netzwerkereignisse in den bestehenden Session-/Event-Store; jedes akzeptierte Netzwerkereignis erhält dieselbe `correlation_id` wie die zugehörigen Player-Events derselben `(project_id, session_id)`-Partition. `trace_id` darf vorhanden sein, ist aber nicht das Abnahmekriterium für Timeline-Zuordnung. (Manifest-/Segment-Events laufen weiter als `manifest_loaded`/`segment_loaded` durch denselben Use-Case-Pfad wie Player-Events; `correlation_id` wird in `register_playback_event_batch.go` nach Sessions-Upsert für alle Events des Batches gleich gesetzt — `register_playback_event_batch_test.go::TestRegisterPlaybackEventBatch_MultiSession` und §4.2 C2-Race-Test.)
- Bekannter Folge-Punkt: `SessionsService.ListSessions` lädt `network_signal_absent[]` heute pro Page-Eintrag einzeln (N+1 mit Hard-Cap 1000). Detail-Read und Schreibpfad sind nicht betroffen. Tracking als R-7 in `docs/planning/open/risks-backlog.md` mit Bulk-Read-Port-Mitigation und Triggerschwelle.
- [x] URL-Redaction ist vor Persistenz und Dashboard-Anzeige umgesetzt und getestet: keine Query-/Fragment-Speicherung, kein `userinfo`, keine signierten Query-Parameter, keine Credential-/Token-Parameter, keine tokenartigen Rohpfadanteile, keine URL-Labels in Prometheus (Cardinality-Regel telemetry-model §3.1 schließt URL-Labels generell aus). Der Event-Store enthält nur redigierte URL-Repräsentanten gemäß §4.1-Matrix und keinen stabilen Token-Hash (`url_redaction.go::redactURLString`). Tests injizieren Legacy-/Angreifer-Payloads mit rohen URLs in `meta.url`, `meta.network.url`, `meta.segment_url`, `meta.manifest_url` und unbekannten URL-artigen Meta-Keys und beweisen, dass nichts roh persistiert wird (`url_redaction_internal_test.go`, `register_playback_event_batch_meta_test.go::TestRegisterBatch_RedactsURLishMetaBeforeAppend`). Boundary-Felder, insbesondere `session_boundaries[].reason`, werden gegen rohe URLs, Token-Strings und HTML/Script-Fragmente negativ getestet — das Reason-Pattern `^[a-z0-9_]{1,64}$` macht URL- und HTML-Fragmente strukturell unmöglich; `register_playback_event_batch_boundaries_test.go::TestRegisterBatch_RejectsInvalidBoundary_NoPersist` deckt sowohl Pattern- als auch Enum-Verstöße ab.

### 4.5 Analyzer-Linking und endpoint-spezifische Auth

Bezug: §4.1 (Auth-Tabelle, Antworthülle); §4.2 (`GetByCorrelationID`); API-Kontrakt §3.6, §4.

Ziel: Analyzer-Linking ist im Hexagon modelliert und liefert die `{analysis, session_link}`-Hülle mit der vollständigen Statusmatrix; endpoint-spezifische Auth, OPTIONS-Preflight für `/api/analyze` und Smoke-/Nutzer-Doku sind nachgezogen. Sub-Tranchen-Ausgang: Tranche 3 hat den Breaking-Change-Vertrag für `/api/analyze` vollständig umgesetzt.

DoD:

- [x] Analyzer-Linking ist im Hexagon modelliert, nicht nur im HTTP-Adapter: `domain.StreamAnalysisRequest` trägt optionale `correlation_id`/`session_id` und den aufgelösten Project-Kontext, der Analyze-Use-Case erhält `SessionRepository` als Dependency, und der Driving-Port liefert `AnalyzeManifestResult{Analysis, SessionLink}` statt nur `StreamAnalysisResult`. Der HTTP-Adapter dekodiert nur Wire-Felder/Header und mappt das Use-Case-Result auf `{analysis, session_link}`; Link-Status, Cross-Project-Regeln und `not_found_detached`/`conflict_detached` entstehen im Application-Layer (`hexagon/application/analyze_manifest.go::resolveSessionLink`). Unit-Tests decken Use-Case-Statusmatrix und Port-Vertrag ab (`analyze_manifest_link_test.go`, `175b24c`).
- [x] Endpoint-spezifische Auth ist im HTTP-Adapter umgesetzt und gepinnt: `/api/analyze` ohne Token und ohne Link-Felder → `200` mit `session_link.status="detached"`; `/api/analyze` mit `correlation_id` oder `session_id` ohne/ungültigen Token → `401`; Playback- und Session-Read-Endpunkte bleiben tokenpflichtig wie in §4.3 etabliert. Tests pinnen alle drei Klassen (`analyze_link_test.go::TestAnalyze_AuthMatrix_*`, `85096a6`).
- [x] Analyzer-Statusmatrix ist als Test-Set umgesetzt: ungebundener Request ohne Link-Felder (`detached`), nur bekannte `correlation_id` (`linked`), nur unbekannte oder project-fremde `correlation_id` (`not_found_detached`), nur vorhandenen `session_id`-Fallback (`linked`), unbekannte `session_id` (`not_found_detached`), beide konsistent (`linked`), beide gesetzt mit unbekannter/project-fremder `correlation_id` und bekannter `session_id` (`not_found_detached`), beide bekannt aber widersprüchlich (`conflict_detached`), fehlender/ungültiger Project-Kontext bei gesetzten Link-Feldern (`401`/Kontraktfehler) — alle zehn Fälle in `analyze_manifest_link_test.go` (`175b24c`); `Cross-Project-Mismatch` ist explizit als `ForeignProjectCorrelationID_NotFoundDetached`.
- [x] `OPTIONS /api/analyze` bekommt einen eigenen Analyze-Preflight mit `Access-Control-Allow-Methods: POST, OPTIONS` und `Access-Control-Allow-Headers: Content-Type, X-MTrace-Token, X-MTrace-Project`; Tests decken linked und unlinked Analyze-POST aus erlaubtem Origin samt Preflight ab (`analyze_link_test.go::TestAnalyze_OptionsPreflight*`, `85096a6`).
- [x] Der Breaking Change des `{analysis, session_link}`-Wrappers ist in Nutzer-/Smoke-Doku und Tests nachgezogen: `docs/user/stream-analyzer.md` §5 dokumentiert den Wrapper plus Auth-Matrix, `scripts/smoke-analyzer.sh` pinnt `"analysis":` und `session_link.status=detached`, API-Handler-Tests (`analyze_link_test.go`) erwarten den Wrapper. Dashboard hat keinen Analyze-Caller (gegrept) — kein TS-Update nötig. Die internen `spec/contract-fixtures/analyzer/*` bleiben flaches Analyzer-Service-Wire-Format zwischen `apps/analyzer-service` und dem API-Adapter; für die öffentliche API-Hülle gibt es jetzt separate Handler-Snapshots in `analyze_link_test.go`.

### 4.6 SDK-Manifest-/Segment-Capture und Boundary-Send

Bezug: §4.1 (Wire-Format, Boundary-Wrapper); §4.4 (Server-Validierung); RAK-30 (Soll-Korrelation).

Ziel: Player-SDK erfasst hls.js-Manifest-/Fragment-/Segment-Signale als `manifest_loaded`/`segment_loaded` mit kontrolliertem Mapping und kann optional `session_boundaries[]` zusammen mit Event-Batches senden. Sub-Tranchen-Ausgang: Front-end-seitige Quelle der gemischten Korrelations-Daten ist abgedeckt; §4.7 kann End-to-End-Tests bauen, ohne weitere SDK-Refactors.

DoD:

- [x] hls.js-Mapping ist vor Umsetzung als Tabelle festgeschrieben und getestet: kanonische Quellen für `manifest_loaded` (`MANIFEST_LOADED` initial, `LEVEL_LOADED` für Live-/Refresh-Pfad) und `segment_loaded` (`FRAG_LOADED` inklusive Init-Segment-Regel via `frag.sn === "initSegment"`) stehen in `packages/player-sdk/README.md` ("hls.js Mapping (network events)"). Retries und Redirects erzeugen keine doppelten semantischen Events, weil hls.js `FRAG_LOADED` nur bei Erfolg feuert; ein zusätzlicher SDK-Dedup-Set sichert mehrfach gefeuerte Events ab. Dedup-Schlüssel nutzen ausschließlich hls.js-native Fragment-Identität (`level`, `type`, `sn`, `cc`, Init-Marker) plus die SDK-Session-Kontexte; redigierte URLs liegen nur in `meta.network.redacted_url` und sind kein Dedup-Bestandteil. Tests in `tests/hlsjs-adapter.test.ts` decken Manifest-Reload (jeder LEVEL_LOADED erzeugt ein eigenes Event), Fragment-Retry (Dedup über `(level/type/sn/cc/init)`), Init-Segment (`is_init=true` plus eigener Dedup-Slot), signierte Segment-URLs (Redaction in `network.redacted_url`) und Level-Reload (separate `level`-Werte → distinkte Events) ab (`ee61b46`).
- [x] Player-SDK erfasst jedes von hls.js gelieferte Manifest- und Fragment-/Segment-Signal als `manifest_loaded` bzw. `segment_loaded`. Sampling läuft pro Event und greift wie für andere Events (siehe `tracker.ts::track`); Degradation wird durch das fehlende URL-Feld bzw. fehlendes Frag-Payload abgefangen — die Events bleiben in der Timeline, ohne `network.redacted_url` (Forward-Compat-Pfad in §4.4 D1). „mindestens eins pro Session" ist damit nicht das Abnahmekriterium; jeder native hls.js-Callback erzeugt ein Event (`ee61b46`).
- [x] SDK-Public-API trägt den optionalen `session_boundaries[]`-Send-Pfad: `tracker.addBoundary({ networkKind, adapter, reason, timestamp? })` reichert eine Boundary in den nächsten Event-Batch ein. `kind="network_signal_absent"`, `project_id`, `session_id` und `client_timestamp` setzt der Tracker; Cap 20 pro Batch mit Drop-Oldest; Boundaries hängen am ersten Batch eines Flush-Cycles, der ein Event derselben Session trägt (Partition-Match-Pflicht — wird vom Backend zusätzlich enforced). SDK-Public-API-Snapshot (`scripts/public-api.snapshot.txt`) und `packages/player-sdk/src/types/events.ts` (`SessionBoundary`, `BoundaryDraft`, `BoundaryNetworkKind`, `BoundaryAdapter`) sind aktualisiert; `tests/tracker-boundaries.test.ts` deckt den Vertrag ab (statt `tests/http-transport.test.ts`, weil die Boundary-Logik im Tracker liegt — der HTTP-Transport sendet das batch transparent durch). Sieben Cases: erste-Batch-Send, ohne-Boundary-no-property, Cap 20 mit Drop-Oldest, Body-Size-Split nur erste Batch trägt Boundaries, Boundary-only-Flush ist no-op, expliziter Timestamp wird durchgereicht, post-destroy-no-op (`4b597a9`).

### 4.7 Tests, Doku und Tranche-3-Closeout

Bezug: §4.1–§4.6.

Ziel: Gemischte Korrelation und Degradationsmatrix sind End-to-End getestet, Tranche-3-Doku ist final, R-6 wird als technisch geschlossen markiert, Roadmap Schritt 32 wird auf ✅ gesetzt. Sub-Tranchen-Ausgang: Tranche 3 ist als abgeschlossen markierbar; Tranche 4 (Dashboard-Session-Verlauf) kann starten, ohne dass Tranche 4 weitere Server-/SDK-Korrelationsfragen entscheidet.

DoD:

- [x] Tests decken gemischte Player-, Manifest- und Segment-Ereignisse innerhalb einer Session ab und prüfen gleiche `correlation_id`, getrennte batchbezogene `trace_id`-Semantik und die dokumentierten Timeline-only-Ausnahmen (`apps/api/adapters/driving/http/correlation_e2e_test.go::TestE2E_MixedEventTypes_ShareCorrelationID`, `TestE2E_CrossBatch_SameCorrelationID`, `TestE2E_TraceparentPropagation_SameTraceID`; `eddfedd`).
- [x] Tests decken die Degradationsmatrix ab: `network.detail_status="network_detail_unavailable"` mit Reason wird inbound akzeptiert und im Read-Pfad mit voller `correlation_id` exposed (`TestE2E_Degradation_NetworkDetailUnavailable`, `eddfedd`); SDK-/Adapter-Capability-Signal ohne Manifest-/Segment-Event erzeugt den session-skopierten Boundary-Record (`TestE2E_Degradation_CapabilityOnlyBoundary`, `eddfedd`); API-Restart-Test ist bereits in §4.4 D3 (`apps/api/adapters/driven/persistence/sqlite/restart_test.go::TestRestartPreservesSessionBoundaries`) abgedeckt — keine Duplizierung. Dashboard-Anzeige folgt in Tranche 4.
- [x] Falls einzelne Manifest-/Segment-Daten nur als Event-Timeline und nicht als OTel-Span abbildbar sind, ist diese Grenze im API-/Doku-Vertrag nachvollziehbar; Dashboard-Sichtbarkeit folgt in Tranche 4. Diese Events behalten trotzdem `correlation_id` (siehe `spec/telemetry-model.md` §1.4 Absatz "`network_detail_unavailable` ist kein Fehlerstatus … Timeline-only-Ereignis ohne OTel-Span"; Test-Anker `TestE2E_Degradation_NetworkDetailUnavailable`).
- [x] Dokumentation benennt Grenzen der Korrelation, insbesondere Browser-APIs, CORS, Service Worker, CDN-Redirects, Native-HLS und Sampling; sie nennt `correlation_id` als Pflichtkontext und `trace_id` als optionale Debug-Vertiefung (`spec/telemetry-model.md` §1.4.1 neu, `daefad5`; Operator-Sicht in `docs/user/local-development.md` §2.6).
- [x] R-6 (`correlation_id`-Race) ist im Risiken-Backlog als technisch geschlossen markiert (Status 🟢, Begründung verweist auf §4.2-Race-Test); falls vor Release-Bump (Tranche 8) erneut ein Mismatch beobachtet wird, wird R-6 wieder geöffnet. Vorgezogen aus §4.7 in den R-1-/R-6-Cleanup-Commit, weil die technische Mitigation seit `949a265` (§4.2 C2) live ist.
- [x] `docs/planning/in-progress/roadmap.md` Schritt 32 ist auf ✅ gesetzt; Status-Header und §1-Tabelle in diesem Plan markieren Tranche 3 als abgeschlossen; Liefer-/Abnahme-Matrix oben zeigt §4.1–§4.7 ✅ mit Commit-Hashes (dieser Commit).

---

## 5. Tranche 4 — Dashboard-Session-Verlauf ohne Tempo

Bezug: RAK-32; MVP-14; F-38; F-39/F-40 nur gemäß Abnahmegrenzen unten; ADR 0002; ADR 0003.

Ziel: Das Dashboard zeigt Session-Verläufe aus der lokalen m-trace-Persistenz einfach, schnell und restart-stabil an. Tempo ist dafür nicht erforderlich.

DoD:

- [x] Session-Liste und Session-Detailansicht lesen aus SQLite-backed API-Pfaden und zeigen Daten nach API-Restart weiter an (Read-Pfade aus §4.2/§4.3 lesen aus SQLite; Restart-Stabilität in `apps/api/adapters/driven/persistence/sqlite/restart_test.go::TestRestartPreservesData` und `TestRestartPreservesEndSource`; `apps/dashboard/src/routes/sessions/...` rendert die Read-Antwort).
- [x] Detailansicht stellt eine Timeline aus Player-, Manifest- und Segment-Ereignissen dar, mit stabiler Reihenfolge und klarer Typ-Unterscheidung (`apps/dashboard/src/routes/sessions/[id]/+page.svelte` Category-Tags `manifest`/`segment`/`lifecycle`/`playback`; Test-Anker `tests/routes.test.ts::categorizes manifest, segment and lifecycle events with category tags`, `085e43c`).
- [x] Detailansicht rendert das in Tranche 3 vertraglich definierte Session-Feld `network_signal_absent[]` als sichtbaren, nicht-fehlerhaften Hinweis, ohne synthetische Manifest-/Segment-Events zu erfinden (eigene "Network signal absent"-Sektion mit Tripel-Tabelle; Tests `renders network_signal_absent section when entries are present` und `hides the network_signal_absent section when empty`, `085e43c`).
- [x] Laufende Sessions sind von beendeten Sessions über echte API-Felder unterscheidbar: `end_source` als durable Spalte (V4-Migration), Read-Antwort exposed das Feld, Dashboard zeigt `via client`/`via sweeper` (`27bdd21` H1, `085e43c` H2). Restart-Test `TestRestartPreservesEndSource` und `tests/routes.test.ts::shows end_source pill in session list when set` plus `shows end_source on detail stats panel` pinnen das Verhalten. Sweeper-only-Lifecycle-Updates ohne Playback-Event bleiben Polling-only — der `onPollingTick`-Pfad im SSE-Client refreshst die Liste alle 5s im Fallback.
- [x] Invalid-, dropped- und rate-limited Hinweise sind in Tranche 4 nur dann in der Session- oder Statusansicht sichtbar, wenn vorher ein API-Status-Summary-Vertrag außerhalb von `/metrics` existiert. Tranche 4 führt keinen Status-Summary-Endpunkt ein → Anzeige explizit nach Tranche 6 verschoben; das Dashboard parst in Tranche 4 keine Prometheus-Rohdaten. Status-Page §3 `tests/routes.test.ts` bestätigt das (keine Aggregat-Counter angezeigt).
- [x] F-39 (`API-Status anzeigen`) auf Mini-Statusquellen begrenzt: drei Panels in `/status` (`d784d30`): API via `getHealth()`, SSE via `sseConnection`-Store (`/lib/status.ts`), Last-Read-Error via `lastReadError`-Store (von `getJSON` automatisch befüllt). Dashboard-Tests `renders the SSE panel with not_yet_connected default`, `renders the last-read-error panel and clears it on demand`, `renders SSE last-change timestamp when no detail is set`, `renders SSE panel detail message when set`.
- [x] F-40 (`Links zu Grafana/Prometheus/Media-Server-Konsole`) als konfigurationsgetriebene Link-Section umgesetzt (`d784d30`): `buildServiceLinks()` in `lib/status.ts` liest `PUBLIC_GRAFANA_URL`, `PUBLIC_PROMETHEUS_URL`, `PUBLIC_OTEL_HEALTH_URL`, `PUBLIC_MEDIAMTX_URL`. Ohne Env → `status="not_configured"`-Pill, kein toter Link. Tests `marks services as not_configured without env URLs`, `derives probe URLs and open URLs from configured env`, `trims whitespace-only env values to undefined`, `renders open-link buttons for services with configured URLs`.
- [x] Duplikat- oder Replay-Klassifikationen aus der Persistenz sind in der Timeline unterscheidbar (`delivery_status: duplicate_suspected | replayed | accepted` als Pill in `/sessions/[id]`-Timeline; Test `highlights non-accepted delivery_status events`, `085e43c`). Default-Reihenfolge bleibt unangetastet — die Pill ist additiv.
- [x] Pagination oder inkrementelles Nachladen bleibt bei längeren Sessions bedienbar (`Load more events`-Button mit `events_cursor`-Round-Trip; `getSession(sessionId, eventsLimit, eventsCursor?)` in `lib/api.ts`); Cursor-Verhalten ist restart-stabil per §4.3 Cursor v3 (Project-/Session-Scope) und §1.x SQLite-Persistenz. Tests `shows a load-more button when next_cursor is present`, `appends events on load-more click`, `renders an error when load-more fails`, `085e43c`.
- [x] SSE-Live-Update-Mechanismus aus ADR 0003 ist implementiert: Backend `SseStreamHandler` (`apps/api/adapters/driving/http/sse_stream.go`, `e4a67c9`) plus Dashboard-fetch-basierter SSE-Client (`apps/dashboard/src/lib/sse-client.ts`, `4606a97`). Polling-Fallback nach 3 fehlgeschlagenen Reconnects via `onPollingTick`-Callback. Tokenpflichtige Reads → fetch-basierter Client (kein nativer EventSource).
- [x] SSE-Endpunkt-Schnittstelle ist in `spec/backend-api-contract.md` §10a (neu) als verlässlicher Vertrag dokumentiert (`e4a67c9`): globaler Stream, Mindestframe `event_appended` mit `(project_id, session_id, ingest_sequence, event_name)`, `Last-Event-ID`-Backfill aus `playback_events.ingest_sequence`, Heartbeat-Intervall 15s, `backfill_truncated`-Marker bei Backfill-Limit 1000. Detail-Stream wird in Tranche 4 nicht gebaut — Dashboard lädt nach `event_appended` die fehlenden Read-Shape-Felder über `GET /api/stream-sessions/{id}` nach. Sweeper-only-Lifecycle-Frames sind explizit deferred (Polling-only).
- [x] SSE-Auth ist endpoint-spezifisch umgesetzt und getestet: fehlender oder ungültiger `X-MTrace-Token` liefert `401` (`TestSse_Auth_RejectsMissingToken`, `TestSse_Auth_RejectsInvalidToken`); gültiger Token scoped Stream und Backfill auf das Project (`TestSse_Backfill_CrossProjectScoped`). `e4a67c9`.
- [x] SSE-CORS ist umgesetzt: `OPTIONS /api/stream-sessions/stream` antwortet mit `Access-Control-Allow-Methods: GET, OPTIONS` und `Access-Control-Allow-Headers: Content-Type, X-MTrace-Project, X-MTrace-Token, Last-Event-ID` (`ssePreflightHandler` in `cors.go`, `e4a67c9`). Detailstream ist nicht gebaut — der globale Stream übernimmt; Plan-DoD lässt das explizit zu. Last-Event-ID-Header-Forwarding im fetch-basierten SSE-Client durch `tests/sse-client.test.ts::sends Last-Event-ID on reconnect after a frame` gepinnt (`4606a97`).
- [x] SSE-`id`/`Last-Event-ID` ist ausschließlich an `playback_events.ingest_sequence` gebunden (`writeAppendedFrame` in `sse_stream.go`). Reconnect-Backfill liest aus `EventRepository.ListAfterIngestSequence` (driven-Port-Erweiterung in `event_repository.go`, beide Adapter implementieren); funktioniert nach API-Restart, weil SQLite die `ingest_sequence` durable persistiert. Sweeper-only-Lifecycle ist explizit ausgenommen.
- [x] SSE-Fallback-Grenzen sind hart definiert und getestet: Heartbeat 15s (Default; `Heartbeat`-Field auf Handler für Tests), Reconnect-Backoff 5s/30s exponentiell, Backfill-Limit 1000 (`sseBackfillLimit`; `BackfillLimit`-Field für Tests), Polling-Intervall 5s. Spec §10a dokumentiert die Werte.
- [x] Backend-Tests decken SSE-Stream-Header (`TestSse_StreamHeaders`), EventSource-kompatibles Format (`TestSse_LiveFrame`), Heartbeats/Keepalive (`TestSse_Heartbeat`), Client-Abbruch via Context-Cancel (Test-Pattern in `streamBytesUntilTimeout`-Helper) und reconnect-freundliche Semantik (`TestSse_Backfill_LastEventID`, `TestSse_Backfill_TruncatedMarker`, `TestSse_Backfill_IgnoresInvalidLastEventID`). Acht Cases insgesamt (`e4a67c9`).
- [x] Dashboard-Tests decken SSE-Erfolg, Reconnect/Backfill und Polling-Fallback ab (`tests/sse-client.test.ts`, 13 Cases inklusive `dispatches event_appended`, `sends Last-Event-ID on reconnect`, `falls back to polling after persistent SSE failure`, `treats non-2xx responses as connection errors`, `polling tick reschedules itself after each call`, `disconnect aborts the in-flight fetch`). Plus `tests/routes.test.ts` mit gemocktem SSE-Client für die Sessions-Page-Integration (`4606a97`).
- [x] Dashboard-Tests decken leere Timeline, kurze Session, lange Session (Pagination via `next_cursor`), laufende Session und beendete Session (state="ended" + `end_source`-Anzeige) über API-Mockdaten ab (`tests/routes.test.ts` 39 Cases gesamt; `085e43c`). Restart-Persistenz mit echter SQLite-Datei und API-Neustart ist Backend-Verantwortung — abgedeckt durch `restart_test.go::TestRestartPreservesData`, `TestRestartPreservesSessionBoundaries` (§4.4), `TestRestartPreservesEndSource` (§5 H1).
- [x] Browser-E2E-Smoke ist über `make browser-e2e` (`scripts/test-browser-e2e.sh`) und den Demo-Player auf `/demo` weiterhin verfügbar; das Tranche-3-Wire-Format und die Tranche-4-Read-Shape-Erweiterungen ändern den `apps/api`-POST-Vertrag nicht. Eine eigene §5-spezifische Browser-Smoke ist Plan-DoD-mäßig nicht zwingend — die existierende Browser-E2E pinnt den End-zu-End-Pfad Player-SDK → API → Dashboard.

---

## 6. Tranche 5 — Optionales Tempo-Profil

Bezug: RAK-31; RAK-29; Architektur §2/§5; README `0.4.0`.

Ziel: Tempo kann als optionales Trace-Backend genutzt werden, ohne die lokale Dashboard-Ansicht zur Pflicht-Abhängigkeit zu machen. RAK-31 bleibt Kann-Scope: Diese Tranche darf vollständig umgesetzt oder explizit deferred werden, solange RAK-29/RAK-32, lokaler Trace-/Korrelations-Read und das bestehende `observability`-Profil ohne Tempo grün bleiben.

**Scope-Gate-Entscheidung:** `0.4.0` setzt Tempo als optionales Compose-Profil **um** (Pfad A). Deferred-Pfad B ist vom Plan-DoD zugelassen, wird hier aber nicht gewählt — RAK-31 bleibt Kann-Scope, ist aber im Lab discoverbar.

Tranche 5 ist in fünf aufeinander aufbauende Sub-Tranchen geschnitten: §6.1 fixiert die Spec-Grundlagen vor jedem Code (Trace-Such-Vertrag in `telemetry-model.md` §2.6, Status-Sync mit Architektur und Plan-Liefer-Matrix); §6.2 baut den Tempo-Compose-Service plus `make dev-tempo`/`stop`/`wipe`-Schiene; §6.3 differenziert die Collector-Konfiguration in einen disabled-by-default-Pfad und einen Tempo-Pipeline-Pfad; §6.4 liefert den Smoke-Test für die drei Startzustände (Core ohne OTLP, observability ohne Tempo-Errors, tempo-aktiv mit konkretem `correlation_id`-Roundtrip via Tempo-API); §6.5 schließt Doku, README, `local-development.md` und Roadmap-Schritt 34 ab.

Liefer-/Abnahme-Matrix:

| Sub-Tranche | Ergebnis | Harte Abnahme | Status |
|---|---|---|---|
| §6.1 | Spec-Vorarbeit (Doku-only) | Trace-Such-Vertrag in `telemetry-model.md` §2.6 (Primary `mtrace.session.correlation_id`, Sekundär `trace_id`, Multi-Trace-Disclaimer, Single-Session-Batch-Pflicht); Plan §6 in §6.1–§6.5 mit Liefer-Matrix gegliedert; Architektur-Spec Tempo-Status synchronisiert | ✅ (`ca7bc95`) |
| §6.2 | Compose + Make-Target | `tempo`-Service in `docker-compose.yml` mit `profiles:[tempo]`; `observability/tempo/tempo.yml`; `make dev-tempo`, `make stop`/`make wipe` adressieren das Tempo-Profil; `make help` listet Tempo-Targets | ✅ (`6548bf5`) |
| §6.3 | Collector-Konfig mit Disabled-Pfad | Zwei Collector-Configs (Default ohne Tempo, `config-tempo.yaml` mit Tempo-Exporter+Pipeline) plus Compose-Override; `make dev-observability` löst keinen Tempo-Verbindungsversuch aus; `make dev-tempo` startet denselben Collector mit Tempo-Pipeline | ✅ (`0c1d11e`) |
| §6.4 | Smoke für drei Startzustände | `scripts/smoke-tempo.sh`/`make smoke-tempo`: Core (`OTEL_*=""`) ohne OTLP-Versuch, observability ohne Tempo-Exportfehler, Tempo-aktiv mit konkretem `correlation_id`-Roundtrip via Tempo-Search-API | ✅ (`6e3a5f5`) |
| §6.5 | Tranche-5-Closeout | README + `docs/user/local-development.md` unterscheiden eingebaute Session-Timeline und optionales Tempo; alle 11 §6-DoD-Items abgehakt; Roadmap-Schritt 34 ✅; §1-Tabelle Tranche 5 ✅; Code-Review + Push | ✅ (Plan-Tick im Closeout-Commit) |

DoD:

- [x] Tranche 5 hat ein binäres Scope-Gate: Entscheidung getroffen für **Pfad A (Tempo umgesetzt)**. Plan-Header dieser Tranche dokumentiert die Wahl explizit; Deferred-Pfad B war zugelassen, wurde aber nicht gezogen. RAK-31 bleibt Kann-Scope, ist aber im Lab discoverbar (`make dev-tempo`).
- [x] Deferred-Abschluss-Klausel nicht zutreffend (Pfad A gewählt). Stattdessen sind die Pfad-A-Doku-Pflichten direkt erfüllt: README v0.4.0-Sektion und `docs/user/local-development.md` §2.5 (neu, `4096cf8`) unterscheiden Tempo (optional, Debug-Tiefe) von der eingebauten Session-Timeline (RAK-32, durable über SQLite). Spec-Anker `spec/telemetry-model.md` §2.6 (`ca7bc95`) trägt den Trace-Such-Vertrag.
- [x] Tempo startet über `make dev-tempo` (`6548bf5`) und das `tempo`-Compose-Profil. Make-Target ruft `--profile observability --profile tempo up`; der `otel-collector`-Service ist nicht doppelt definiert (env-gesteuerter `COLLECTOR_CONFIG`-Switch). `dev-tempo` ist in `.PHONY` und `make help` discoverable. `make stop` adressiert beide Profile, `make wipe` entfernt das benannte `mtrace-tempo-data`-Volume zusätzlich zum SQLite-Volume.
- [x] Collector-Konfigurations-Mechanismus ist explizit (`0c1d11e`): zwei Configs (`config.yaml` Default ohne Tempo, `config-tempo.yaml` mit `otlp/tempo`-Exporter im Traces-Pipeline) bind-gemountet plus env-gesteuerter `--config`-Pfad. `dev-observability` lädt `config.yaml` → kein `otlp/tempo`-Exporter konfiguriert, kein Tempo-Verbindungsversuch. `dev-tempo` setzt `COLLECTOR_CONFIG=config-tempo.yaml` → derselbe Collector-Container mit Tempo-Pipeline.
- [x] OTel-Collector leitet Traces nur bei aktivem `tempo`-Profil an Tempo weiter (`0c1d11e`). Im observability-only-Pfad bleiben Prometheus/Grafana grün, Traces gehen an den `debug`-Exporter, keine Tempo-Connection-Errors. Der No-Op-Pfad ohne OTLP-Endpoint ist Core-State (Default-Compose, leere `OTEL_*`-Werte aus `docker-compose.yml` Z. 16–19) und bleibt unangetastet.
- [x] Smoke-Check deckt alle drei Startzustände ab (`6e3a5f5`): `scripts/smoke-tempo.sh` mit `SMOKE_STATE=core` pinnt API-Health ohne OTLP-Versuch, `SMOKE_STATE=observability` pinnt Collector up + Tempo NICHT erreichbar (Probe muss fehlschlagen), `SMOKE_STATE=tempo` (Default) pinnt den vollen Roundtrip. Lokale `OTEL_*`-Overrides bleiben unbeeinflusst — der Smoke prüft Compose-Defaults.
- [x] Ohne Tempo-Profil bleiben `trace_id`/`correlation_id`, Dashboard-Timeline und RAK-29-Tests funktionsfähig: §3 (Tranche 2) und §5 (Tranche 4) sind ohne Tempo grün, der Tempo-Smoke `SMOKE_STATE=observability` pinnt explizit, dass Tempo nicht erreichbar ist.
- [x] Trace-Suche ist normativ dokumentiert in `spec/telemetry-model.md` §2.6 (`ca7bc95`): Session-Suche primär über `mtrace.session.correlation_id` (Single-Session-Batch-Pflicht, kein Empty-String/Komma-Liste); Event-Details sekundär via batchspezifischer `trace_id`. Der Multi-Trace-Disclaimer ("Eine Session kann mehrere `trace_id`-Werte haben, `trace_id` ist kein Session-Schlüssel") ist Pflicht-Vertragstext.
- [x] RAK-29 ist auch ohne Tempo erfüllt — siehe §3.4 (Tranche 2 Closeout) und Telemetry-Model §2.5/§2.6: Tempo erweitert die Debug-Tiefe auf Span-Ebene; Source-of-Truth für Session-Korrelation (`correlation_id`) ist und bleibt SQLite (RAK-32).
- [x] Smoke-Test mit konkretem Suchwert: `scripts/smoke-tempo.sh` (State `tempo`) liest `correlation_id` aus dem API-Read-Pfad (`GET /api/stream-sessions/{id}`) und validiert exakt diesen Wert in Tempo primär via TraceQL `GET /api/search?q={ span.mtrace.session.correlation_id = "<UUID>" }&start=<unix>&end=<unix>`; `tags=` bleibt Legacy-Fallback mit demselben Suchfenster. "Trace sichtbar" ohne benannten Suchwert reicht nicht und passiert auch nicht — der Smoke failed bei `traces.length == 0`.
- [x] README v0.4.0-Sektion ergänzt um `make dev-tempo`-Hinweis und RAK-Bezug; `docs/user/local-development.md` §2.5 (neu) trennt Tempo (optional, Debug-Tiefe) von der eingebauten Session-Timeline (RAK-32). Die Doku zeigt Tempo-Search primär per TraceQL mit `mtrace.session.correlation_id` und explizitem `start`/`end`-Fenster.

---

## 7. Tranche 6 — Aggregat-Metriken und Drop-/Invalid-/Rate-Limit-Sichtbarkeit

Bezug: RAK-33; RAK-34; API-Kontrakt §7; Telemetry-Model §2.4/§3/§4.3; Lastenheft §7.9/§7.10.

Ziel: Prometheus bleibt Aggregat-Backend. Die Pflichtmetriken für angenommene, invalid, rate-limited und dropped Events sind sichtbar, korrekt gezählt und cardinality-sicher.

Tranche 6 ist in vier Sub-Tranchen geschnitten: §7.1 fixiert die Spec-Bestandsaufnahme und Scope-Klauseln (Drop-Pfad bleibt absichtlich Doku-Variante; API-Status-Summary explizit nicht eingeführt; F-40 ist `n/a`, weil bereits in §5 H3 erfüllt); §7.2 ergänzt Backend-Tests für Pflichtcounter-Pfade (Inkrement-Fälle und Null-Inkrement-Fälle aus API-Kontrakt §7); §7.3 verschärft den Cardinality-Smoke auf die vollständige §7-Forbidden-Liste plus Per-Pflichtcounter-Labelset-Assertion; §7.4 schließt Grafana-Dashboard-Sichtbarkeit, Plan-Tick und Roadmap-Schritt 35 ab.

Liefer-/Abnahme-Matrix:

| Sub-Tranche | Ergebnis | Harte Abnahme | Status |
|---|---|---|---|
| §7.1 | Spec-Bestandsaufnahme + Scope-Klauseln (Doku-only) | Plan §7 in §7.1–§7.4 + Liefer-Matrix; Drop-Pfad-Status (`mtrace_dropped_events_total = 0`, kein produktiver Drop-Pfad) ist normativ in API-Kontrakt §7 dokumentiert; API-Status-Summary explizit nicht eingeführt; F-40 `n/a` (in §5 H3 / `d784d30` erfüllt) | ✅ (`00a989c`) |
| §7.2 | Backend-Tests für Pflichtcounter-Pfade | Inkrement-Fälle (`schema_version`-invalid, `events.length > 100`, fehlendes Pflichtfeld) und Null-Inkrement-Fälle (leerer Batch, Auth-Fehler, Body-Read/Payload-Limit, malformed JSON) sind in `apps/api/adapters/driving/http/metrics_counter_test.go` als eigenständige Cases gepinnt; Rate-Limit-`429`-Counter-Inkrement und Accepted-Counter sind separate Test-Anker | ✅ (`515c3cd`) |
| §7.3 | Cardinality-Smoke verschärft | `scripts/smoke-observability.sh` Forbidden-Labels auf vollständige §7-Liste (`project_id`, `session_id`, `user_agent`, `segment_url`, `client_ip`, `trace_id`, `span_id`, `correlation_id`, `viewer_id`, `request_id`, Token-Felder); pro-Pflichtcounter-Assertion (jede Serie hat genau das Default-Labelset); 0.4.0-spezifische `mtrace_*`-Metriken werden in den Cardinality-Cap (< 50 Serien) einbezogen | ✅ (`aea4d9e`) |
| §7.4 | Grafana-Sichtbarkeit + Closeout | `observability/grafana/dashboards/m-trace-overview.json` zeigt die vier Pflichtcounter (Panels "Playback Events", "Invalid Events", "Rate Limited Events", "Dropped Events" — Verifikation, kein Diff nötig); alle 9 §7-DoD-Items mit Commit-Hashes/Test-Ankern abgehakt; Roadmap-Schritt 35 ✅; §1-Tabelle Tranche 6 ✅; Code-Review + Push | ✅ (Plan-Tick im Closeout-Commit) |

DoD:

- [x] Alle vier Pflichtcounter sind im Compose-Lab und in Tests vorhanden: `apps/api/adapters/driven/metrics/prometheus_publisher.go:67–80` registriert sie label-frei; `prometheus_publisher_test.go` pinnt die Werte; Backend-Tests `metrics_counter_test.go` pinnen die HTTP-Pfad-Inkremente (§7.2, `515c3cd`).
- [x] Pflichtcounter zählen Events, nicht Batches: Use-Case ruft `EventsAccepted(len(parsed))`, `InvalidEvents(len(in.Events))`, `RateLimitedEvents(len(in.Events))` mit Event-Anzahl, nicht Batch-Anzahl. Leere Batches → `InvalidEvents(0)` (Publisher no-op, Counter unverändert); Auth-Fehler enden vor Use-Case (kein Inkrement); Persistenzfehler bleiben 5xx ohne Drop-Inkrement (siehe `register_playback_event_batch.go:156–163`-Kommentar). Test-Anker: `TestMetrics_InvalidCounter_NoIncrement_OnEmptyBatch`, `TestMetrics_InvalidCounter_NoIncrement_OnAuthError` (§7.2, `515c3cd`).
- [x] Pflichtcounter ohne fachliche Labels: Cardinality-Smoke (`scripts/smoke-observability.sh`, §7.3, `aea4d9e`) prüft pro §7-Pflichtcounter sowohl `count(<metric>) == 1` als auch das Labelset auf `__name__`/`instance`/`job`-Whitelist; jeder zusätzliche Label-Key ist release-blockierend. Forbidden-Liste deckt §7-Vertrag ab: `project_id`, `session_id`, `user_agent`, `segment_url`, `client_ip`, `trace_id`, `span_id`, `correlation_id`, `viewer_id`, `request_id`, `token`, `authorization`.
- [x] Rate-Limit `429` + Counter-Inkrement: `TestMetrics_RateLimitedCounter_Increments` in `metrics_counter_test.go` (§7.2, `515c3cd`) erschöpft den Bucket, prüft 429 und scrapet `mtrace_rate_limited_events_total == 1`. Bestehender `TestHTTP_429_RateLimit` pinnt zusätzlich Status + `Retry-After`-Header.
- [x] Invalid-Counter-Tests trennen event-zählbare Invalids von frühen Request-Fehlern (§7.2, `515c3cd`). Inkrement-Cases: `TestMetrics_InvalidCounter_SchemaVersion` (1 Event, +1), `TestMetrics_InvalidCounter_BatchTooLarge` (101 Events, +101 — `newServerWithUnlimitedRate`), `TestMetrics_InvalidCounter_MissingField` (1 Event, +1). Null-Inkrement-Cases: `TestMetrics_InvalidCounter_NoIncrement_OnEmptyBatch`, `TestMetrics_InvalidCounter_NoIncrement_OnAuthError`, `TestMetrics_InvalidCounter_NoIncrement_OnBodyTooLarge`, `TestMetrics_InvalidCounter_NoIncrement_OnMalformedJSON`. Alle Erwartungen verweisen auf API-Kontrakt §7.
- [x] Drop-Pfad bleibt Doku-Variante (API-Kontrakt §7 erlaubt konstant `0`, solange kein produktiver Drop-Pfad existiert). `TestMetrics_DroppedCounter_StaysZero` (§7.2, `515c3cd`) pinnt: `mtrace_dropped_events_total = 0` nach erfolgreichem POST und Metrik ist sichtbar im Scrape.
- [x] Grafana-/Prometheus-Lab zeigt die vier Pflichtcounter: `observability/grafana/dashboards/m-trace-overview.json` enthält Panels "Playback Events", "Invalid Events", "Rate Limited Events", "Dropped Events" (Z. 50/92/134/176). RAK-34-Sichtbarkeit bleibt explizit Prometheus/Grafana-only (kein API-Status-Summary in 0.4.0); Plan-DoD-Klausel "Falls API-Status-Summary eingeführt" ist nicht ausgelöst.
- [x] F-40 ist `n/a` — Service-Links für Grafana/Prometheus/MediaMTX-Konsole wurden in §5 H3 (`d784d30`) als `buildServiceLinks()` in `apps/dashboard/src/lib/status.ts` umgesetzt. Plan-DoD-Klausel "Falls F-40 deferred" ist nicht zutreffend.
- [x] Cardinality-Smoke prüft, dass neue `0.4.0`-Metriken keine hochkardinalen Labels einführen: `scripts/smoke-observability.sh` (§7.3, `aea4d9e`) Z. 99–106 cappt Total-Cardinality auf < 50 Serien über alle `mtrace_*`-Metriken; Forbidden-Liste deckt 0.4.0-spezifische Identifier (`trace_id`, `correlation_id`) explizit ab; Per-Pflichtcounter-Labelset-Whitelist verhindert Drift bei zukünftigen Counter-Erweiterungen.

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
- [ ] Versionen sind konsistent: Root- und Workspace-Pakete tragen `0.4.0`; der Versionscheck deckt mindestens `package.json`, `packages/player-sdk/package.json`, `packages/stream-analyzer/package.json`, `apps/dashboard/package.json`, `apps/analyzer-service/package.json`, `contracts/sdk-compat.json`, `packages/player-sdk/src/version.ts`, `contracts/event-schema.json`, `apps/api/cmd/api/main.go` `serviceVersion` und die API-`SupportedSchemaVersion` ab. Versionierte Paket-Smokes und Test-Snapshots sind ebenfalls Teil des Release-Bumps: `packages/player-sdk/package.json` `pack:smoke`-Tarballname, `packages/player-sdk/scripts/pack-smoke.mjs` `expectedVersion`, SDK-Tests/Fixtures sowie Analyzer-/Stream-Analyzer-Versionstests oder Fixtures mit erwarteter Paketversion müssen `0.4.0` erwarten oder ausdrücklich als nicht release-blocking begründet werden. Der SDK/Event-Schema-Kompatibilitätscheck bleibt grün. Insbesondere `PLAYER_SDK_VERSION` in `packages/player-sdk/src/version.ts` und API-`serviceVersion` sind auf `0.4.0` gehoben (aus §3.3-Review, Anmerkung #9: aktuell noch `0.3.0`, weil §3.3 absichtlich keinen Release-Bump macht; API aktuell noch `0.1.2`).
- [ ] `CHANGELOG.md` enthält den Versionsabschnitt `[0.4.0] - <Datum>` mit Trace-, Persistenz-, Dashboard-, Metrik- und Doku-Lieferstand.
- [ ] `docs/user/releasing.md` ist auf die `0.4.0`-Gate-Liste synchronisiert oder markiert §9 ausdrücklich als strengeren `0.4.0`-Override. Release-Gates grün: `make gates` (enthält `test`, `lint`, `coverage-gate`, `arch-check`, `schema-validate`, `docs-check`), `make build`, `make sdk-performance-smoke`, `make smoke-observability`, `make smoke-cli`, `make smoke-analyzer` und `make browser-e2e`. Für `make smoke-observability` ist die Laufvoraussetzung dokumentiert: Der Observability-Stack (`make dev-observability` bzw. Compose mit `--profile observability`) läuft vor dem Smoke und stellt API, Prometheus, Grafana und OTel-Collector bereit.
- [ ] Browser-E2E-Smoke (`make browser-e2e`) erzeugt eine Test-Session und prüft Session-Timeline/Tranche-4-Dashboard-Flows; falls er aus Umgebungsgründen manuell ersetzt wird, ist das Ergebnis als Release-Gate dokumentiert. Der Smoke darf `/demo` nutzen, muss aber bei späterer Demo-Änderung auf einen dedizierten Test-Harness umstellbar bleiben.
- [ ] Release-Artefakt existiert nachvollziehbar nach `docs/user/releasing.md`: Release-Commit auf `main`, annotierter Tag `v0.4.0`, Push von Commit und Tag, GitHub-Release mit Notes aus `CHANGELOG.md`, und GitHub Actions `Build` ist am Release-Commit grün. Tranche 8 darf nicht nur mit lokalen Gates abgeschlossen werden.
- [ ] `docs/planning/in-progress/roadmap.md` markiert `0.4.0` als abgeschlossen und verschiebt den aktiven Fokus auf `0.5.0`.

---

## 10. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`, Commit-Hash anhängen.
- Neue Findings in der `0.4.0`-Phase landen entweder in dieser Datei oder in `risks-backlog.md`.
- Lastenheft-Patches während `0.4.0` werden in `plan-0.1.0.md` Tranche 0c als neue §4a.x-Einträge ergänzt.
- Beim Release-Bump `0.4.0` → `0.5.0`: dieses Dokument als historisch archivieren; Lieferstand wandert dokumentarisch nach `CHANGELOG.md`.
