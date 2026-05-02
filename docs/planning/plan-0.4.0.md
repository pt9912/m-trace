# Implementation Plan — `0.4.0` (Erweiterte Trace-Korrelation)

> **Status**: 🟡 in Arbeit. Tranche 0 + Tranche 1 (§2.1–§2.6) abgeschlossen; Tranchen 2–8 offen.
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

Ziel: Player-Sessions werden konsistent als Trace-Konzept modelliert. OTel-Spans und gespeicherte Events teilen stabile Korrelations-IDs, ohne Prometheus-Cardinality-Regeln zu verletzen.

Tranche 2 ist in vier aufeinander aufbauende Sub-Tranchen geschnitten: §3.1 fixiert die Spec-Grundlagen vor jedem Code, §3.2 baut die Server-Korrelation (Spans, Persistenz, Validierung), §3.3 die SDK-Wire-Format-Erweiterung, §3.4 die Tests und finalisiert die Doku.

### 3.1 Spec-Vorarbeit (Doku-only, kein Code)

Bezug: RAK-29; RAK-35; Telemetry-Model §2/§3/§5; API-Kontrakt §3, §5, §8.

Ziel: Vor Code-Änderungen sind Trace-ID-Strategie, Span-Modell, `correlation_id`-Vertrag, SDK-Wire-Format-Erweiterung und Validierungsregeln verbindlich entschieden, sodass §3.2–§3.4 ohne implizite Spec-Entscheidungen umgesetzt werden können. Sub-Tranchen-Ausgang: keine Code-Diffs, aber alle nachfolgenden Sub-Tranchen können auf eindeutige Spec-Aussagen verweisen.

Verbindliche Entscheidungen (gehören in `spec/telemetry-model.md` und `spec/backend-api-contract.md`):

- **Trace-ID-Quelle: Hybrid.** Player-SDK propagiert optional einen W3C-`traceparent`-Header (Format laut [W3C Trace Context](https://www.w3.org/TR/trace-context/)). Ist der Header gültig, übernimmt der Server `trace_id` und `parent_span_id` aus dem Header und erzeugt einen Child-Span. Fehlt der Header oder ist er ungültig, generiert der Server einen Root-Span mit eigener W3C-konformer `trace_id`. SDK-Wert hat Vorrang gegenüber Server-Generierung.
- **`trace_id` ≠ `correlation_id`.** Beide sind getrennte Konzepte mit klarer Verantwortung:
  - `trace_id` (TEXT, nullable, 32 Hex-Zeichen): W3C-Trace-ID — vom SDK propagiert oder server-generiert; primär für Tempo (RAK-31, optional).
  - `correlation_id` (TEXT, immer pro Session gesetzt): server-generierte, durable Source-of-Truth für die Dashboard-Korrelation. Wird beim allerersten Event einer Session erzeugt (UUIDv4 oder vergleichbar), in `stream_sessions.correlation_id` persistiert und für **alle** Folge-Events derselben Session konstant gehalten.
  - Dashboard-Timeline (RAK-32) nutzt `correlation_id` — Tempo-unabhängig. Tempo (RAK-31) nutzt `trace_id`, wenn das Profil aktiv ist.
- **Span-Modell: ein HTTP-Request-Span pro Batch.** Keine Child-Spans pro Event (Cardinality-Risiko). Pflicht-Attribute am Server-Span:
  - `mtrace.project.id` (kontrolliert; Allowlist aus dem Use-Case-Resolver)
  - `mtrace.batch.size` (int)
  - `mtrace.batch.outcome` (`accepted` / `invalid` / `rate_limited` / `auth_error` etc.)
  - **Bei Single-Session-Batch (alle Events teilen `session_id`)** zusätzlich `mtrace.session.correlation_id` (und nur dieser Wert — `session_id` selbst ist Prometheus-tabu, im Span-Attribut aber zulässig, weil Spans sampled/short-lived sind).
  - `mtrace.batch.session_count` (int) — bei Multi-Session-Batches > 1; das einzelne Event trägt seine `correlation_id` aus der Persistenz, nicht der Span.
  - `mtrace.trace.parse_error=true` falls eingehender `traceparent` ungültig war (siehe Validierungsregel unten).
  - `mtrace.time.skew_warning=true` falls für mindestens ein Event im Batch `|client_timestamp - server_received_at| > 60s` (Schwelle aus `telemetry-model.md` §5.3); Persistenz des Skew-Flags auf Event-Ebene ist deferred (siehe unten).
- **Defensive Validierung des `traceparent`-Headers.** Ein ungültiger oder formal kaputter Header führt **nicht** zu 4xx und nicht zum Absturz. Stattdessen: Span-Attribut `mtrace.trace.parse_error=true`, Server-Fallback erzeugt eine eigene `trace_id`, Event wird normal verarbeitet. Die Pflicht-Validierungs-Reihenfolge aus API-Kontrakt §5 wird dadurch nicht verändert.
- **Cardinality-Regel.** Weder `trace_id`, `correlation_id` noch `span_id` werden als Prometheus-Labels verwendet — Span-Attribute (kontrolliert), Event-Persistenz-Spalten (durable) und Wire-Format-Felder (optional) sind die einzigen Konsumenten.
- **Time-Skew-Handling: nur Span-Attribut in `0.4.0`.** Span-Attribut `mtrace.time.skew_warning=true` aus `telemetry-model.md` §5.3 ist Pflicht in §3.2. Persistenz-Spalte und Dashboard-Anzeige sind explizit deferred — ein Folge-Tranchen-Item wird sie ergänzen, sobald Bedarf entsteht.

DoD:

- [x] `spec/telemetry-model.md` §2.5 (neu) dokumentiert die Hybrid-Trace-ID-Strategie, das `trace_id`/`correlation_id`-Verhältnis (mit Persistenz-Quelle pro Feld), das Span-Modell mit Pflicht-Attribut-Tabelle und die Sampling-Auswirkung für `0.4.0` (`5a8ab19`).
- [x] `spec/backend-api-contract.md` §1 dokumentiert `traceparent` als optionalen HTTP-Header auf `POST /api/playback-events` mit defensiver Server-Validierung (kein 4xx bei kaputtem Header). §3.7 ergänzt `correlation_id` (immer gesetzt) und `trace_id` (nullable) als server-vergebene Read-Felder ab `0.4.0`-§3.2-Closeout (`5a8ab19`).
- [x] Cardinality-Regel ist in `spec/telemetry-model.md` §2.5 festgehalten: `trace_id`/`correlation_id`/`span_id` sind Prometheus-tabu — Span-Attribute, Persistenz-Spalten und Wire-Format-Felder sind die einzigen Konsumenten; Verstöße sind release-blocking via Cardinality-Smoke (`5a8ab19`).
- [x] Folge-Item für persistenten Time-Skew-Flag ist als R-5 in `docs/planning/risks-backlog.md` aufgenommen; §5.3 in `telemetry-model.md` verweist explizit auf den Backlog-Eintrag und markiert die Persistenz-Spalte als deferred (`5a8ab19`).

### 3.2 Server-Korrelation

Bezug: §3.1; Telemetry-Model §5.4; API-Kontrakt §3; ADR-0002 §8.1 (Schema-Spalten reserviert in §2.3).

Ziel: Backend liest `traceparent` (wenn vorhanden), erzeugt einen Server-Span pro Batch, generiert/liest `correlation_id` pro Session, persistiert `trace_id`/`span_id`/`correlation_id` auf jedem Event und auf der Session, validiert defensiv. Sub-Tranchen-Ausgang: ein Batch-POST hinterlässt einen kompletten Trace mit korrelations-fähigen Persistenz-Daten — auch ohne SDK-`traceparent`.

DoD:

- [x] HTTP-Adapter parst `traceparent`-Header gemäß W3C-Spec (`apps/api/adapters/driving/http/traceparent.go`); bei valider `trace_id`/`parent_span_id` wird der Server-Span als Child gestartet (`withTraceParent`), sonst als Root mit Server-`trace_id` und Span-Attribut `mtrace.trace.parse_error=true` (`c3741aa`).
- [x] HTTP-Request-Span für `POST /api/playback-events` trägt die in §3.1 spezifizierten Attribute. `mtrace.project.id` ist in 0.4.0 noch nicht gesetzt (Use-Case-Resolver-Wert nur bei Erfolg verfügbar — als Folgepunkt im Backlog vermerken oder in §3.4 nachziehen, falls nötig); alle übrigen Pflicht-Attribute (`http.method/route/status_code`, `batch.size/outcome/session_count`, optional `mtrace.session.correlation_id`, `mtrace.trace.parse_error`, `mtrace.time.skew_warning`) sind gesetzt (`c3741aa`).
- [x] `domain.PlaybackEvent` und `domain.StreamSession` sind um `TraceID`/`SpanID`/`CorrelationID` (Event) bzw. `CorrelationID` (Session) erweitert; Application- und Adapter-Code füllt sie konsistent (`c3741aa`).
- [x] `correlation_id` wird beim allerersten Event einer Session in `RegisterPlaybackEventBatch.resolveCorrelationIDs` erzeugt (UUIDv4 via `crypto/rand`); existing Sessions liefern sie aus dem Repository; existing Sessions ohne `correlation_id` (Legacy von vor §3.2-Closeout) bekommen via Self-Healing eine neue. SessionRepository persistiert sie in `stream_sessions.correlation_id` (`c3741aa`).
- [x] SQLite-Adapter und InMemory-Adapter schreiben und lesen die drei neuen Spalten korrekt; gemeinsamer Contract-Test `testTraceFieldsRoundTrip` deckt beide Backends ab (`c3741aa`).
- [x] Defensive Validierung: `parseTraceParent` lehnt jeden Formatfehler ab (Längen, Hex, Version, all-zero); `withTraceParent` mappt das auf `mtrace.trace.parse_error=true` ohne 4xx (`c3741aa`).
- [x] Time-Skew-Detection mit Konstante `TimeSkewThreshold = 60 * time.Second` im Use-Case; bei Treffer `BatchResult.TimeSkewWarning=true`, der HTTP-Adapter setzt das Span-Attribut (`c3741aa`).
- [x] Adapter-Contract-Tests in `persistence/contract` erweitert um `testTraceFieldsRoundTrip`; läuft identisch gegen InMemory und SQLite. Use-Case-Test-Suite erweitert um fünf Cases (neue Session, existing Session mit/ohne CorrelationID, Multi-Session, Time-Skew, Trace-Context-Durchreiche) plus `parseTraceParent`-Unit-Tests (`c3741aa`).

### 3.3 SDK-Wire-Format-Erweiterung

Bezug: §3.1; §3.2 (Server-Pfad muss bereit sein, bevor SDK Header schickt); RAK-29 (kein Breaking Change).

Ziel: Player-SDK propagiert optional einen W3C-`traceparent`-Header, wenn der Browser-Pfad einen aktiven Span hat oder das SDK selbst eine `trace_id` führen kann. Wenn kein Trace-Kontext da ist, schickt das SDK den Header **nicht** — Server-Fallback erzeugt eine eigene `trace_id`. Schema-Version bleibt `1.0`. Sub-Tranchen-Ausgang: SDK 0.4.0 kann den Header optional schicken; Server toleriert sowohl SDKs mit als auch ohne Header.

DoD:

- [x] `@npm9912/player-sdk` HTTP-Transport setzt `traceparent`-Header optional über die neue `PlayerSDKConfig.traceparent`-Provider-Funktion; ohne Provider oder bei Provider-Return `undefined`/`""` bleibt der Header weg. Provider-Throws werden im SDK still gefangen — Tracing darf den Event-Pfad nicht sabotieren (`8f3011c`).
- [x] Abwärtskompatibilität: kein Wire-Format-Bruch (Header ist additiv, Payload unverändert); ältere Backends ignorieren unbekannte Header per HTTP-Standard. SDK-Doku verweist explizit darauf (`8f3011c`).
- [x] SDK-Tests in `packages/player-sdk/tests/http-transport.test.ts` decken: Provider-Wert → Header gesetzt; Provider-`undefined` → kein Header; Provider-`""` → kein Header; kein Provider konfiguriert → kein Header; Provider-Throw → still verworfen; Provider wird pro Send aufgerufen, nicht gecached (sechs Cases) (`8f3011c`).
- [x] Schema-Version bleibt `1.0`; SDK↔Backend-Kompatibilitätscheck (CI `make gates`) bleibt grün — `EVENT_SCHEMA_VERSION` und `PLAYER_SDK_VERSION` sind unverändert (`8f3011c`).
- [x] `spec/player-sdk.md` neue Sektion „Trace-Korrelation (optional, ab `0.4.0`)" zeigt das Provider-Pattern (Beispielcode mit OpenTelemetry-Bridge), nennt die Backwards-Compat-Garantie und verweist auf den Vertrag in `spec/telemetry-model.md` §2.5 (`8f3011c`).

Closeout-Notiz: Der §3.3-Code-Commit `8f3011c` hat eine zweite Lint-Regression hinterlassen (`packages/player-sdk/scripts/public-api.snapshot.txt` enthielt `TraceParentProvider` nicht, obwohl `src/index.ts` ihn exportiert). Das §3.3-Review hat den Snapshot-Drift nicht gefangen; `make workspace-lint` wäre auf `8f3011c` rot gewesen. Geheilt im Followup-Commit `7bbea4d` (Should-fix #3+#4, Anmerkungen #6+#7+#8 aus dem §3.3-Review plus Drive-by-Snapshot-Fix). Im Review-of-Review-Commit `f7dcdb9` zusätzlich: parallele Spec-Drift in `spec/player-sdk.md` (Public-API-Bulletliste) geheilt, einmaliger `console.warn` pro `HttpTransport`-Instanz für Non-String-Returns und Provider-Throws ergänzt (Observabilität ohne Send-Pfad-Sabotage), `HttpTransportOptions.silent` für Tests, JSDoc-`@see`-Verlinkung statt Doppeltexte. Lehre für Tranche 8 / Folge-Reviews: Reviewer muss `make workspace-lint` und nicht nur `make workspace-test` laufen lassen; Snapshot-Drift ist nicht durch `tsc --noEmit` abgedeckt. Die §3.3-DoD-Items oben pinen weiterhin auf `8f3011c` als historisches Lieferdatum; das beobachtbare Schluck-Verhalten reflektiert den Stand ab `f7dcdb9` und ist in `spec/player-sdk.md` festgeschrieben.

### 3.4 Tests und Doku-Closeout

Bezug: §3.1–§3.3.

Ziel: Trace-Konsistenz ist auf allen Ebenen abgesichert (mehrere Batches einer Session teilen `correlation_id`; ungültiger Trace-Kontext führt zu sauberem Fallback; Tempo-deaktivierter Pfad funktioniert ungestört). Doku spiegelt den ausgelieferten Stand. Sub-Tranchen-Ausgang: Roadmap §2 Schritt 29 ist auf ✅ aktualisierbar.

§3.4 ist in drei Sub-Tranchen geschnitten: §3.4a sichert das Server-Verhalten aus §3.2 mit Backend-Tests ab (rein server-seitig, nutzt Use-Case + HTTP-Adapter + tracetest.SpanRecorder); §3.4b deckt zwei Cross-Cutting-Pfade zwischen SDK und Server ab, die §3.3-Review als Should-fix #1/#2 markiert hat (Cross-Version-Kompat, E2E-Garbage); §3.4c finalisiert die Spec-Texte und schließt Roadmap Schritt 29.

#### 3.4a Backend-Tests Trace-Konsistenz

Bezug: §3.2 (Server-Pfad mit Spans, `correlation_id`-Resolver, `parseTraceParent`, Time-Skew); ADR-0002 §8.1; Telemetry-Model §2.5.

Ziel: Das in §3.2 ausgelieferte Server-Verhalten ist durch wiederholbare Backend-Tests gegen Use-Case + HTTP-Adapter abgesichert; jeder spezifizierte Pfad (Multi-Batch-Konsistenz, fehlender Kontext, ungültiger Kontext, Session-Ende, Time-Skew, Tempo-deaktiviert) hat einen eigenen Test mit klaren Assertions auf `trace_id`, `correlation_id` und Span-Attributen. Sub-Tranchen-Ausgang: Reviewer kann §3.2-Lieferung gegen §3.4a-Tests nachvollziehen, ohne externes Trace-Backend zu brauchen.

DoD:

- [x] Backend-Test deckt Trace-Konsistenz über mehrere Batches einer Session: drei aufeinanderfolgende Batches mit gleicher `session_id` produzieren drei verschiedene `trace_id`-Werte (jeder Batch ein Trace), aber **dieselbe** `correlation_id` an allen Events und der Session — `TestHTTP_Trace_MultiBatchSameSessionConsistency` (`f329d5f`).
- [x] Backend-Test deckt fehlenden Client-Kontext: Batch ohne `traceparent` → Server generiert `trace_id`, `mtrace.trace.parse_error` ist nicht gesetzt — `TestHTTP_Trace_MissingTraceparent_ServerGeneratesTrace` (`f329d5f`); persistiert auch die server-generierte `trace_id` aufs Event.
- [x] Backend-Test deckt ungültigen Client-Kontext: Batch mit kaputtem `traceparent` → 202 Accepted, Span-Attribut `mtrace.trace.parse_error=true`, `trace_id` ist server-generiert — bereits im §3.2-Bestand `TestHTTP_Span_TraceParent_InvalidSetsParseError` (`c3741aa`); §3.4a-Header-Comment in `trace_consistency_test.go` mappt das DoD-Item explizit dorthin (`f329d5f`).
- [x] Backend-Test deckt Session-Ende: `session_ended`-Event innerhalb eines Batches behält die `correlation_id` der Session bei und schließt den State; nachfolgende Events in derselben Session-ID erhalten dieselbe `correlation_id` (Reihenfolge ist Tranche-1-Verhalten) — `TestHTTP_Trace_SessionEnded_PreservesCorrelationID` (`f329d5f`).
- [x] Backend-Test verifiziert Time-Skew-Span-Attribut bei `|client_timestamp - server_received_at| > 60s` — bereits im §3.2-Bestand `TestHTTP_Span_TimeSkew_SetsWarning` (`c3741aa`); §3.4a-Header-Comment mappt das DoD-Item dorthin (`f329d5f`).
- [x] Backend-Test verifiziert Trace-Konsistenz **mit NoOp-`TracerProvider`** (Stand-in für Tempo-deaktivierten Pfad ohne `OTEL_TRACES_EXPORTER`): `correlation_id` bleibt gesetzt, Dashboard-Timeline ist nutzbar. Test darf kein externes Trace-Backend voraussetzen — Realisierung über den `tracenoop`-Fallback in `router.go`, ausgelöst durch `nil`-Tracer-Argument. Die Config-Resolution (welche `cmd/api`-Config einen `nil`-Tracer ergibt) ist explizit nicht Teil dieses Tests — `TestHTTP_Trace_NoopTracer_CorrelationStillPersisted` (`f329d5f`).

#### 3.4b Cross-Cutting-Tests SDK ↔ Server

Bezug: §3.3-Review (Should-fix #1/#2); §3.2; §3.3.

Ziel: Zwei Pfade, die SDK und Server überspannen, sind explizit getestet — Vorwärtskompat zu Pre-§3.2-Backends und das Garbage-Traceparent-Ende-zu-Ende-Verhalten. Sub-Tranchen-Ausgang: das §3.3-Review hat keine Test-Lücken mehr offen.

DoD:

- [ ] Cross-Version-Vertragstest (aus §3.3-Review, Should-fix #1): SDK `0.4.0` mit konfiguriertem `traceparent`-Provider gegen einen Server-Handler auf `0.3.x`-Verhaltensstand (kein Header-Lesen, keine `correlation_id`-Persistenz) liefert weiterhin `202 Accepted`; der Header darf nicht zu Validierungs-/Parser-Fehlern führen. Realisierung als Adapter-Test mit minimal-konfiguriertem Handler oder Snapshot des Pre-§3.2-Verhaltens.
- [ ] E2E-Test mit kaputtem `traceparent` (aus §3.3-Review, Should-fix #2): SDK-`HttpTransport` sendet einen Provider-gelieferten Garbage-String; Server akzeptiert den Batch (`202`) und setzt `mtrace.trace.parse_error=true`; SDK-Pfad bleibt unverändert (keine Drop-, Retry-, `console.warn`-Effekte — Garbage-String ist `typeof === "string"` und triggert deshalb nicht den §3.3-Followup-Warn).

#### 3.4c Doku-Closeout und Roadmap-Marker

Bezug: §3.1–§3.3; §3.4a–§3.4b.

Ziel: Spec-Texte sind final mit dem Code synchronisiert; Roadmap Schritt 29 ist als ✅ markierbar; offene Items aus dem §3.3-Review (Anmerkung #5 Header-Casing) sind eingearbeitet. Sub-Tranchen-Ausgang: Tranche 2 ist abgeschlossen, Tranche 3 kann starten.

DoD:

- [ ] Header-Casing/Whitespace-Kommentar (aus §3.3-Review, Anmerkung #5): kurze Notiz in `spec/backend-api-contract.md` §1, dass der Server `traceparent` case-insensitiv liest (HTTP-Header-Standard) und führende/abschließende Whitespaces toleriert; SDK schreibt lowercased `traceparent`.
- [ ] `spec/telemetry-model.md` ist final konsistent mit Code (Hybrid-Strategie, Span-Attribute, Time-Skew, Sampling); §3.1-Entscheidungen sind festgeschrieben.
- [ ] `spec/backend-api-contract.md` §3 / §3.7 reflektiert das `traceparent`-Header-Verhalten und die neuen Read-Felder `trace_id`/`correlation_id`.
- [ ] `docs/planning/roadmap.md` Schritt 29 ist auf ✅ gesetzt; Status-Header und Verweis auf den Tranche-2-Closeout-Stand sind aktualisiert.

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
- [ ] Versionen sind konsistent: Root- und Workspace-Pakete tragen `0.4.0`; SDK/Event-Schema-Kompatibilitätscheck bleibt grün. Insbesondere `PLAYER_SDK_VERSION` in `packages/player-sdk/src/version.ts` ist auf `0.4.0` gehoben (aus §3.3-Review, Anmerkung #9: aktuell noch `0.3.0`, weil §3.3 absichtlich keinen Release-Bump macht).
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
