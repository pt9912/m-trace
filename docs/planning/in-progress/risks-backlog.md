# Risiken-Backlog

> **Stand**: 2026-05-13 — `0.18.0` (Offene Risiken /
> Trigger-Re-Eval) released 2026-05-13 (Tag `v0.18.0`); kein
> Lastenheft-Patch, normativer Stand bleibt `1.1.22`. `0.17.0`
> released 2026-05-13 (Tag `v0.17.0`); `0.16.0` released
> 2026-05-12 (Tag `v0.16.0`); Vorgänger `0.15.0` released
> 2026-05-12 (Tag `v0.15.0`);
> `0.14.0`
> released 2026-05-12 (Tag `v0.14.0`);
> `0.13.0` released 2026-05-12 (Tag `v0.13.0`);
> `0.12.6` released 2026-05-12 (Tag `v0.12.6`);
> `0.12.5` released 2026-05-11 (Tag `v0.12.5`);
> `0.12.1` released 2026-05-10 (Tag `v0.12.1`);
> `0.12.0` released 2026-05-10 (Tag `v0.12.0` auf `e3457fa`).
>
> ### Aktuelle Phase
>
> `0.19.0` ist als Roadmap-Trigger-Nacharbeit abgeschlossen und in
> [`done/plan-0.19.0.md`](../done/plan-0.19.0.md) archiviert. Die
> Decision-Welle erzeugt kein neues R-N-Item: Postgres (`MVP-40`) und
> strengere CORS-Preflight-Project-Isolation (Variante A) bleiben ohne
> belegten Betreiber-/Multi-Tenant-Trigger deferred; Analyzer-API und
> Control-Plane sind als triggerfreie Decision-Tracks mit Proceed-/
> POC-/Defer-Kriterien dokumentiert. Kein Release-Tag, kein Versions-
> Bump.
>
> `0.18.0` ist released und archiviert in
> [`done/plan-0.18.0.md`](../done/plan-0.18.0.md). Die Trigger-Re-Eval
> fuer `R-9`, `R-12` und `R-13` hat keinen Implementierungs-Trigger
> belegt; alle drei Risiken bleiben mit praezisierten Folge-Triggern
> offen. Es entsteht weiterhin kein neues R-N-Item.
>
> `0.17.0` ist released und archiviert in
> [`done/plan-0.17.0.md`](../done/plan-0.17.0.md); Lastenheft-Patch
> `1.1.22` mit RAK-111..RAK-115 bleibt der normative Stand.
>
> `0.16.0` ist released und archiviert in
> [`done/plan-0.16.0.md`](../done/plan-0.16.0.md). Der Release
> importiert aus `0.15.0` genau Szenario B: HTTP-Range-/Byte-Range-
> Loader fuer manifest-referenzierte CMAF-Init-/Media-Segmente.
> Geliefert wurde ein begrenzter HLS-CMAF-Byte-Range-Fetch fuer
> explizite `EXT-X-MAP:BYTERANGE`-/`#EXT-X-BYTERANGE`-Offsets ohne
> neues Public-Schema. RAK-106..RAK-110 sind geschlossen; es entsteht
> kein neues R-N-Item. Die Risiken sind ueber RAK-109 kontrolliert und
> im Release-Closeout mit Artefakt- und Gate-Nachweis geschlossen.
>
> `0.15.0` ist released und archiviert in
> [`done/plan-0.15.0.md`](../done/plan-0.15.0.md). Tranche 0
> wählte Szenario A: Zielgruppe + Analyzer-Boundary. Aus `0.14.0`
> importiert:
> Postgres bleibt `defer-with-migration-seed`, Analytics bleibt
> `defer`, und **R-9** bleibt ein K8s-Smoke-/Allowlist-Trigger:
> optionale K8s-Beispiele liegen unter `deploy/k8s/`, aber keine
> K8s-Smoke-Pflicht und keine Erweiterung der Compose-Lab-Label-
> Allowlist ohne separaten K8s-Allowlist-Modus. Die Aktivierung von
> `0.15.0` erzeugt kein neues R-N-Item; Zielgruppen-, Analyzer-,
> Control-Plane- und Ops-Trigger werden über RAK-101..RAK-105
> entschieden. Tranche 1 schließt RAK-101: Primärziel bleiben
> Selbsthoster, kleine bis mittlere Streaming-Teams, Broadcaster-Labs
> und technische Media-/DevOps-Teams; große Plattformbetreiber bleiben
> triggerbasierter Folge-Scope. Tranche 2 schließt RAK-102:
> eine externe `apps/analyzer-api` bleibt deferred, solange kein
> konkreter externer Konsument, kein Job-/Retention-Bedarf und kein
> Auth-/Rate-Limit-/SSRF-/Contract-Nachweis vorliegen. Tranche 3
> schließt RAK-103: `apps/control-plane` bleibt deferred und bekommt
> keinen POC, solange kein konkreter Betreiberbedarf mit Auth-/Tenant-/
> Audit-Nachweis, Owner, SLO und Folgeplan vorliegt. Tranche 4
> schließt RAK-104: HTTP-Range-/Byte-Range-Loader ist der einzige
> empfohlene kleine `NF-13`-Folgeslice; LL-CMAF, vollständige
> Segmentsets, Codec-Decoding und Player-Laufzeitpfade bleiben
> deferred. Tranche 5 schließt RAK-105: Postgres bleibt
> `defer-with-migration-seed`, Analytics bleibt `defer`; kein Backend-
> Runtime-Scope in `0.15.0`.
>
> `0.12.6` adressierte zuvor alle neun nach `0.12.5` offenen
> R-N-Items in Tranchen 1–9 (Plan in
> [`plan-0.12.6.md`](../done/plan-0.12.6.md)). T0-Entscheidung Option A
> (alle neun aktiv) sequenziell vor `0.13.0`:
>
> - **R-5** Time-Skew-Persistenz + Dashboard-Marker (Tranche 3, RAK-83).
> - **R-7** `ListSessions` Bulk-Read-Port (Tranche 5, RAK-84).
> - **R-10** Sampling-Marker mit Integer-ppm (Tranche 4, RAK-85).
> - **R-11** SRT-Cursor-Pagination via `samples_cursor`/`next_cursor`
>   + `400 cursor_invalid` gemäß §7a.3/§7a.4 (Tranche 2, RAK-86).
> - **R-13** Trivy-Re-Review 2026-08-04 (Tranche 1, Wartung ohne RAK).
> - **R-15** MediaMTX-Provisionierung mit `provision=true` (Tranche 9, RAK-87).
> - **R-17** Multi-Host-Issuance-Limiter via Redis-Backend
>   (Tranche 7, RAK-88; Resttrigger aus `0.12.5` Tranche 2).
> - **R-20** Produktive Vault-AppRole + KMS (Tranche 8, RAK-89;
>   Resttrigger aus `0.12.5` Tranche 3).
> - **R-22** Origin-/IP-Rate-Limiter via Redis (Tranche 6, RAK-90;
>   gemeinsamer Redis-Server mit RAK-88).
>
> **`0.12.5` Auslieferungen** (Plan archiviert in
> [`done/plan-0.12.5.md`](../done/plan-0.12.5.md)): R-14/R-16/R-18/
> R-21 🟢 aufgelöst; R-17/R-20 ⬜ teilweise gelöst (Resttrigger
> werden in `0.12.6` Tranche 7/8 final geschlossen). RAK-77..RAK-82
> in §13.15 persistiert.
>
> Lastenheft-Stand: `1.1.22` (`0.17.0`-Patch persistiert in T0).
> `0.17.0` belegt RAK-111..RAK-115 in §13.21; `0.16.0` belegt
> RAK-106..RAK-110 in §13.20; `0.15.0` belegt
> RAK-101..RAK-105 in §13.19.
>
> ### Trigger-Re-Eval-Stand (2026-05-10, `0.12.1` Tranche 1)
>
> Alle aktiven R-N-Items haben einen Trigger-Stand-Eintrag in der
> Mitigation-Spalte (R-5/R-7/R-9/R-10/R-11/R-12/R-13/R-14/R-15/
> R-16/R-17/R-18/R-20/R-21/R-22 — durchgängig „nicht ausgelöst"
> zum 2026-05-10).
>
> ### OS-Schärfung in `0.12.1` Tranche 1
>
> - `OS-1`..`OS-5` aus `done/plan-0.12.0.md` §10 als ⬛ Duplikate
>   bzw. strukturell-nicht-trackbar in §1.2 abgelegt.
> - `OS-6` (Origin-/IP-Rate-Limiting) zu **R-22** in §1.1
>   konvertiert; Auflösungspfad `plan-0.13.x`.
> - `R-19` als ⬛ historischer Marker (README-Risiko-Grundlage in
>   Commit `5798473` entfernt).
>
> ### Persistente Out-of-Scope-Stände (normativ im Lastenheft)
>
> - OAuth/OIDC/User-Verwaltung bleibt normativ in §13.14 RAK-71
>   Out-of-Scope — **kein eigenes R-N-Item**, weil ohne konkreten
>   Bedarf kein Tracking-Trigger sinnvoll ist; Reaktivierung in
>   `0.15.0` oder einem Folgeplan bei Stakeholder-Bedarf.
> - Konsequenz für **R-14/R-15/R-16** aus `0.12.0`: keines der drei
>   Items wird strukturell aufgelöst — `validate-key` bleibt
>   Diagnose-/Smoke-Funktion (kein Media-Server-Auth-Ersatz),
>   externe Provisionierung bleibt Folge-Scope, ausgehende
>   Webhook-Zustellung bleibt Folge-Scope (optionale Tranche 5
>   in `0.12.5`).
>
> ### Risiko-Status-Markierungen aus Vorgänger-Releases
>
> - **R-12** WebRTC-`getStats()`-Drift seit `0.9.0` Tranche 1 von
>   „release-blockierend ab nächstem Browser-Major-Bump" auf
>   „automatisiert detektiert, Drift bricht den Drift-Smoke"
>   (Nightly-Workflow `webrtc-drift.yml`).
> - **R-13** drei Trixie-OS-CVEs ohne Upstream-Fix — seit `0.8.5`
>   Tranche 1 als Trivy-Ignore-Wartungsregel aktiv, `expires`
>   2026-08-04.
> - **R-5/-7/-9/-10/-11** aus `0.7.0`/`0.8.0` inhaltlich
>   unverändert; Triggerschwellen unausgelöst.
>
> ### Bezug
>
> - `docs/adr/0001-backend-stack.md` §5 (Bewertungsraster,
>   *Absehbare Phase-2-Risiken*), §8 (Konsequenzen)
> - `docs/adr/0002-persistence-store.md` (löst R-4),
>   `docs/adr/0003-live-updates.md` (löst R-3)
> - `spec/lastenheft.md` §4.3, §10.1; §13.14 RAK-71..RAK-76
>   (`0.12.0` Auth/Token); §13.15 RAK-77..RAK-82 (`0.12.5`
>   Auth-/Ingest-Adapter); §13.17 RAK-91..RAK-95 (`0.13.0`
>   Production / Ops Backends); §13.18 RAK-96..RAK-100 (`0.14.0`
>   Ops Backend Follow-up); §13.19 RAK-101..RAK-105 (`0.15.0`
>   Product Scope / Analyzer Boundary); §13.20 RAK-106..RAK-110
>   (`0.16.0` Selected Product Slice / Analyzer Range Fetch);
>   §13.21 RAK-111..RAK-115 (`0.17.0` Hardening / Evidence Review)
> - `docs/planning/in-progress/roadmap.md` §4
> - `docs/planning/done/plan-0.12.0.md` (Folge-ADRs);
>   `docs/planning/done/plan-0.12.1.md` (Patch: Trigger-Re-Eval +
>   Operator-Doku);
>   `docs/planning/done/plan-0.12.5.md` (Minor:
>   Auth-/Ingest-Adapter R-14/R-17/R-18/R-20/R-21, optional R-16);
>   `docs/planning/done/plan-0.13.0.md` (Production / Ops
>   Backends, R-9 Tranche-3-Scope);
>   `docs/planning/done/plan-0.14.0.md` (Ops Backend Follow-up,
>   R-9 Hardening-Scope);
>   `docs/planning/done/plan-0.15.0.md` (Product Scope /
>   Analyzer Boundary, RAK-101..RAK-105);
>   `docs/planning/done/plan-0.16.0.md` (Selected Product
>   Slice / Analyzer Range Fetch, RAK-106..RAK-110);
>   `docs/planning/done/plan-0.17.0.md` (Hardening /
>   Evidence Review, RAK-111..RAK-115)

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
| R-5 | Time-Skew-Persistenz auf Event-Ebene fehlte: `0.4.0` setzte `mtrace.time.skew_warning=true` als Span-Attribut (siehe `spec/telemetry-model.md` §2.5/§5.3), aber die Schema-Spalte und Dashboard-Anzeige waren explizit deferred. Folge: skew-betroffene Events waren im Read-Pfad (Dashboard ohne Tempo) nicht sichtbar markiert. | `plan-0.4.0.md` §3.1; Migration V6 (`time_skew_warning`); `spec/telemetry-model.md` §2.5/§5.3 | `0.12.6` Tranche 3 (Code) ✅ | 🟢 | **gelöst** — Pro-Event-Persistenz geliefert in `0.12.6` Tranche 3: Migration V6 fügt `playback_events.time_skew_warning INTEGER NOT NULL DEFAULT 0` an; Ingest-Use-Case setzt das Bit pro Event basierend auf der bestehenden 60-s-Schwelle (`TimeSkewThreshold`); Read-Pfad (`ListSessions`, `GetSessionDetail`, SSE-`event_appended`-Frame) echo't `time_skew_warning` als JSON-Feld (`omitempty`); Dashboard-Timeline zeigt `⏱ skew`-Pin pro betroffenes Event mit Tooltip auf die Schwelle. Span-Attribut auf Batch-Ebene bleibt unverändert — beide Pfade tragen jetzt denselben Beleg. Tests: Use-Case-Per-Event-Flag (`TestRegisterPlaybackEventBatch_TimeSkewPerEvent`), Adapter-Restart-Persistenz (`TestRestartPreservesTimeSkewWarning`), E2E-Wire (`TestE2E_TimeSkewPersistedPerEvent`). Doku-Update in `spec/telemetry-model.md` §2.5/§5.3. **Wieder-Eröffnungs-Trigger**: Operator-Report über fehlende Skew-Sichtbarkeit im Read-Pfad oder strukturelle Skew-Schwellen-Re-Eval (z. B. konfigurierbare Schwelle pro Project). |
| R-7 | `SessionsService.ListSessions` lud `network_signal_absent[]` pro Session-Page-Eintrag einzeln (`ListBoundariesForSession` N+1). Bei Hard-Cap 1000 Sessions pro Page waren das im Worst Case 1000 SQL-Round-Trips ohne gemeinsamen Tx-Snapshot. | `plan-0.4.0.md` §4.4 D3 (Review-N-1); `apps/api/hexagon/application/sessions_service.go`; `apps/api/adapters/driven/persistence/sqlite/session_repository.go` | `0.12.6` Tranche 5 (Code) ✅ | 🟢 | **gelöst** — Bulk-Read-Port geliefert in `0.12.6` Tranche 5: neue Port-Methode `SessionRepository.ListBoundariesForSessions(ctx, projectID, sessionIDs []string) (map[string][]Boundary, error)` (Driven-Port, project-skopiert); SQLite-Adapter mit dynamischer `IN (?, ?, ?)`-Clause (eine Query pro Page statt N), InMemory-Adapter mit map-lookup. `SessionsService.ListSessions` ruft die neue Methode statt der N+1-Schleife auf; Reihenfolge bleibt parallel zu `page.Sessions`, Default für eine Session ohne Boundaries ist ein leerer Slice (Map-Miss). Bench `BenchmarkSessionsService_ListSessions_MaxPage_BulkBoundaries` (1000 Sessions in einer Page) im Budget < 200 ms p95 (check-bench-budgets.mjs); Adapter-Tests `TestListBoundariesForSessions_BulkReadAndScopeIsolation` + `_EmptyInput` pinnen Bulk-Sortierung, Map-Miss-Verhalten und Cross-Project-Isolation. **Wieder-Eröffnungs-Trigger**: List-Latenz ≥ 200 ms p95 unter realistischer Last (Multi-Tenant-Produktion, hochfrequente `network_signal_absent`-Pfade) → Folge-Item für Detail-Read und Cursor-Stream-Pfad. |
| R-9 | Die Observability-Smoke-Whitelist für Infrastruktur-Labels (`__name__`, `instance`, `job`) ist bewusst Compose-Lab-spezifisch. Eine spätere K8s-Smoke-Stage würde übliche Labels wie `pod`, `namespace` oder `container` wahrscheinlich als False Positive werten. | `plan-0.4.0.md` §7.4 (Review F-2); `plan-0.13.0.md` Tranche 3; `plan-0.14.0.md` Tranche 3; `deploy/k8s/README.md`; ADR 0005 | Folge-Trigger bei K8s-Smoke-Gate | ⬜ | **Entscheidung 2026-05-12 (`0.13.0` Tranche 3)**: K8s bleibt optionaler Beispielpfad (`deploy/k8s/`) und ist nicht production-ready. Keine Änderung am Compose-Lab-Schema und keine K8s-Smoke-Pflicht in `0.13.0`. **Hardening 2026-05-12 (`0.14.0` Tranche 3)**: `make k8s-validate` prüft die Beispielmanifeste clusterfrei, blockiert Infrastruktur-Labelkeys `pod`/`namespace`/`container`, erzwingt Single-Replica-Beispiele und hält Production-Kinds (`Ingress`, `HorizontalPodAutoscaler`, `NetworkPolicy`, `PodDisruptionBudget`) aus dem Seed heraus. **Trigger-Re-Eval 2026-05-13 (`0.18.0` Tranche 1)**: kein Roadmap-Trigger fuer K8s-Smoke als PR-/Release-Gate; `make k8s-validate` deckte eine bestehende Seed-Version-Drift (`0.14.0` statt `0.17.0`) auf. Die Manifeste wurden zunaechst auf den damaligen Release-Stand und mit dem `v0.18.0`-Release auf `0.18.0` synchronisiert. Gegenmaßnahmen für spätere Reaktivierung: (a) separater K8s-Allowlist-Modus für Infrastruktur-Labels (`pod`, `namespace`, `container`, ggf. `service`) statt Erweiterung des Compose-Defaults; (b) Smoke-Scope-Trennung Compose vs. K8s mit explizitem Profil/ENV-Gate; (c) README-/Deploy-Doku markiert Beispielmanifeste als nicht production-ready. **Wieder-Eröffnungs-Trigger**: Ein K8s-Smoke soll PR-/Release-Gate werden oder K8s-Observability-Manifeste landen mit Prometheus-Scrape/Label-Policy. |
| R-10 | Sampling-Vollständigkeitsnachweis für `sampleRate < 1` fehlte: gesampelte Events trugen keine SDK-seitige Markierung, deshalb konnte der Server eine Sampling-Lücke nicht strukturell von einem echten Verlust unterscheiden. | `plan-0.4.0.md` §8.3; Migration V7 (`stream_sessions.sample_rate_ppm`); `spec/telemetry-model.md` §8; `contracts/event-schema.json#reserved_meta_keys["session_sample_rate"]` | `0.12.6` Tranche 4 (Code) ✅ | 🟢 | **gelöst** — Sampling-Markierung auf Session-Ebene geliefert in `0.12.6` Tranche 4: Migration V7 ergänzt `sample_rate_ppm INTEGER NOT NULL DEFAULT 1000000` an `stream_sessions`; SDK-Pflicht-Feld `meta.session_sample_rate` (Float `(0, 1]`) wird server-seitig auf Integer-ppm normalisiert (`domain.SampleRatePPMFromFloat`) und durch Immutability-Set (`UPDATE … WHERE sample_rate_ppm = SampleRateFull`) genau einmal pro Session persistiert; spätere Drift wird via `mtrace_sample_rate_drift_total{project_id}` mit ±100ppm-Toleranz gezählt, aber nicht überschrieben. Read-Pfad echo't `sample_rate_ppm` (Integer, raw) und `sample_rate` (Float, Display-Hilfe) `omitempty` auf der Default-Marke. Dashboard zeigt einen Banner „Sampled session (X.XX %)" mit `data-testid="sampled-banner"`. Doku: spec/telemetry-model.md §8 (neue Section: Wire-Vertrag, Persistenz/Immutability, Drift-Counter, Read-Pfad, Lücken-Heuristik als Folge-Item). Tests: Domain-Helper (Range-Check + Round-Half-Behavior); Use-Case Immutability-First-Set, No-Op-Default, Drift-Counted, Within-Tolerance. **Wieder-Eröffnungs-Trigger**: Operator-Bedarf nach konkreter Lücken-Heuristik (z. B. „erwartete vs. tatsächliche Events" als Read-Pfad-Marker `possible_loss`) — bleibt als Folge-Item für `0.13.0+`. |
| R-11 | SRT-Health-Detail-Pagination war in `0.6.0` Sub-3.3 als ErrNotImplemented gestubbed. Der HTTP-Pfad `GET /api/srt/health/{stream_id}` lieferte nur `samples_limit` (default 100, max 1000) ohne Cursor; eine länger laufende Lab-/Production-Instanz mit > 1000 persistierten Samples pro Stream konnte das ältere Drittel nicht mehr abfragen. | `plan-0.6.0.md` §4 Sub-3.3 (alter Adapter-Stub); `spec/backend-api-contract.md` §7a.3/§7a.4 (Wire-Format) | `0.12.6` Tranche 2 (Code) ✅ | 🟢 | **gelöst** — Cursor-Pagination geliefert in `0.12.6` Tranche 2: SQLite-Adapter implementiert Keyset-Pagination über `(ingested_at, id)` (Index `idx_srt_health_samples_stream_ingested`), Application-Service reicht den Cursor durch (`HistoryByStream(ctx, projectID, streamID, limit, after)`), HTTP-Handler akzeptiert `samples_cursor` und liefert `next_cursor`. Wire-Codec v3 mit Collection-Scope `(project_id, stream_id)` analog §10.3-Event-Cursor — Cross-Project- und Cross-Stream-Cursor werden als `400 cursor_invalid_malformed` rejected. Reject-Klassen-Trennung folgt §10.3-Tabelle: `cursor_invalid_legacy` (v=1/2 oder fehlendes v-Feld) vs. `cursor_invalid_malformed` (alles andere). Spec-Konsistenz-Fix §7a.3/§7a.4 — Pre-§4.3-Wording (`process_instance_id` + monolithische `cursor_invalid`-Klasse) durch §10.3-konformes v3-Wording ersetzt. Tests: Adapter-Pagination-Walk über 1500 Samples + Cross-Stream-Scope-Isolation; HTTP-Roundtrip + sechs `cursor_invalid_*`-Pfade; Contract-Fixtures `srt-health-cursor-invalid-legacy.json` + `srt-health-cursor-invalid-malformed.json`. Operator-Smoke `make smoke-srt-health-pagination` deckt Cursor-Wandern + malformed-Reject. **Wieder-Eröffnungs-Trigger**: Operator-Report über Inkonsistenz im Cursor-Wandern oder neuer Schema-Drift gegen §7a.3/§10.3. |
| R-13 | Drei OS-CVEs ohne Upstream-Fix in `node:22-trixie-slim`, geteilt zwischen Dashboard- und Analyzer-Service-Image: `CVE-2025-69720` (ncurses Buffer Overflow in `libtinfo6`/`ncurses-base`/`ncurses-bin`), `CVE-2026-29111` (systemd Arbitrary Code Exec via IPC in `libsystemd0`/`libudev1`) und `CVE-2026-4878` (libcap TOCTOU-Race in `libcap2`). Keine Ausnutzbarkeit im m-trace-Runtime sichtbar (Container ohne TTY/IPC-Mounts, USER node ohne setcap-Pfad), aber statisch HIGH. Hinweis: Analyzer-Service wechselte mit plan-0.8.5 Tranche 1 Closeout von `node:22-alpine` auf `node:22-trixie-slim`, weil musl-libc bei multi-threaded Workloads (libuv-Worker-Pool, V8-GC/JIT) gegenüber glibc spürbar pessimisiert ist; die einheitliche Trixie-Basis vermeidet das. | `plan-0.8.5.md` §2 Tranche 1 Closeout; `.security/vulnignore.yaml`; `apps/dashboard/Dockerfile`, `apps/analyzer-service/Dockerfile`; `r13-trivy-rereview-2026-05-13.md` | bei Triggerschwelle | ⬜ | **Mitigation aktuell**: dokumentierte Trivy-Ignores in `.security/vulnignore.yaml` mit 90-Tage-`expires` (**aktuell `2026-11-02`**, verlängert in `0.12.6` Tranche 1 von zuvor `2026-08-04`); Generator `scripts/render-trivyignore.sh` bricht ab, sobald ein `expires` überschritten ist (erzwungene Re-Review). **Folge-Trigger**: (a) Trixie-Point-Release liefert Fixes — Eintrag entfernen; (b) `expires`-Schwelle erreicht ohne Fix — Re-Review entscheidet zwischen Verlängerung und Base-Image-Wechsel; (c) Distroless-Variante wird vor 1.0 als Folge-Plan evaluiert (`gcr.io/distroless/nodejs22-debian12` bringt nur Node-Binär ohne npm/Debian-CLI-Surface, eliminiert die drei CVEs strukturell, behält glibc). Aktuell Lab-/Demo-Pfad — keine Production-Kontroll-Plane. **Trigger-Stand 2026-05-11 (`0.12.6` T1 Re-Review)**: nicht ausgelöst — Trivy-Scan zeigt 6 HIGH-Findings (genau die drei bekannten CVEs verteilt auf 6 Packages), `Fixed Version`-Spalte leer in allen Treffern → kein Upstream-Fix in Trixie; `expires` um 90 Tage auf `2026-11-02` verlängert mit Re-Review-Kommentar-Block pro CVE in `.security/vulnignore.yaml`. **Trigger-Re-Eval 2026-05-13 (`0.18.0` Tranche 1)**: `make image-scan` gruen; `mtrace-api:scan` 0 HIGH/CRITICAL ohne Ignores, Dashboard und Analyzer-Service jeweils 3 gescopte Ignores und keine unignorierten HIGH-/CRITICAL-Findings. Re-Review-Artefakt: `docs/planning/in-progress/r13-trivy-rereview-2026-05-13.md`. Distroless-ADR-Draft bleibt deferred. |
| R-23 | OpenSSL Heap-Use-after-free `CVE-2026-45447` (`PKCS7_verify()`, CISA-ADP CVSS 9.8 CRITICAL / Trivy HIGH) im `libssl3t64`/`openssl` der `node:22-trixie-slim`-Base von Dashboard- und Analyzer-Service-Image. **Anders als R-13 existiert ein Upstream-Fix** (`openssl 3.5.6-1~deb13u2`, DSA-6335-1, 2026-06-09); die Images tragen die verwundbare `…deb13u1`, weil das floating `node:22-trixie-slim` zum Build-Zeitpunkt noch `deb13u1` liefert und die Dockerfiles bewusst kein `apt-get upgrade` fahren. Der Befund wurde erst durch den Trivy-Bump `0.59.1 → 0.71.0` (frische Vuln-DB) sichtbar — der Nightly mit `0.59.1` hatte den am 2026-06-09 publizierten CVE noch nicht. Vektor im Runtime nicht erreichbar (laut OpenSSL-Advisory nicht über Standard-TLS; Node-Runtime verarbeitet kein PKCS#7/S-MIME, grep-bestätigt), aber statisch CRITICAL/HIGH. | `.security/vulnignore.yaml` (Eintrag `CVE-2026-45447`); `apps/dashboard/Dockerfile`, `apps/analyzer-service/Dockerfile`; `Makefile` (`TRIVY_IMAGE` → `aquasec/trivy:0.71.0`); OpenSSL-Advisory 2026-06-09 / DSA-6335-1 | bei `expires`/Base-Update | ⬜ | **Mitigation aktuell**: Trivy-Ignore in `.security/vulnignore.yaml` mit **kurzem** `expires` `2026-07-10` (statt der 90-Tage-Default der no-fix-Cohorts), weil der Fix bereits existiert und nur der floating Base-Tag + das Trivy-DB-Lag (leere `Fixed Version`-Spalte) den Gate röten. **Folge-Trigger**: (a) `node:22-trixie-slim` liefert `deb13u2` → `docker build --pull` zieht die gepatchte Lib, Eintrag **ersatzlos entfernen**; (b) Trivy-Debian-Feed kennt die Fixed-Version → Befund verschwindet ohnehin, Eintrag entfernen; (c) `expires`-Schwelle `2026-07-10` ohne (a)/(b) erreicht → Re-Review entscheidet zwischen gezieltem `apt-get install --only-upgrade openssl libssl3` in beiden Runtime-Stages und Verlängerung. **Trigger-Stand 2026-06-10**: ausgelöst durch Nightly-Audit-Fail (Issue #4) + Trivy-Bump; Suppression gesetzt, `make image-scan` mit `0.71.0` anschließend grün. |
| R-15 | Externe Media-Server-Provisionierung war im `0.11.0`-Scope ausgeklammert: `POST /api/ingest/streams` produzierte kein I/O auf laufenden externen MediaMTX-/SRS-Servern, und `GET /api/ingest/media-server-config` schrieb das Konfigurations-Artefakt nur lokal. | `apps/api/hexagon/port/driven/media_server_provisioner.go`; `apps/api/adapters/driven/mediaserver/mediamtx_provisioner.go`; `apps/api/hexagon/application/ingest_control_service.go` (Provision-Pfad); `spec/backend-api-contract.md` §3.8 (additive `provision=true`-Erweiterung); `docs/user/ingest-control.md` §2.2 (Rollback-Pfad) | `0.12.6` Tranche 9 (Code) ✅ | 🟢 | **gelöst** — Optionaler Media-Server-Provisioner geliefert in `0.12.6` Tranche 9: neuer Driven-Port `driven.MediaServerProvisioner` mit `Apply(ctx, MediaServerApplyInput)` und `Rollback(ctx, projectID, streamID)`. Bounded State-Enum: `applied`, `partial`, `failed`, `disabled`. MediaMTX-Adapter (`adapters/driven/mediaserver/mediamtx_provisioner.go`) gegen `/v3/config/paths/add/<stream_id>`-API mit optionalem Bearer-Token (`MTRACE_MEDIASERVER_PROVISION_URL`/`_TOKEN`); idempotent über 409-Conflict-Mapping auf `applied`; bounded ErrorCodes (`unreachable`, `auth_failure`, `server_status_<N>`). Wire-Erweiterung **strikt additiv** in `POST /api/ingest/streams?provision=true`: Default-Pfad (`provision=false` ODER fehlt) bleibt **byte-stabil** zum `0.11.0`-Format — kein `media_server_state`-Feld, kein I/O gegen den externen Server, `MTRACE_MEDIASERVER_PROVISION_URL` wird in diesem Pfad **nicht** gelesen. `provision=true` immer mit `media_server_state.state`-Wert; `disabled`-Pfad liefert Operator-Hint. HTTP-Status bleibt in allen Fällen `201 Created` — **kein** automatisches API-State-Rollback bei `failed`/`partial`; Operator-Rollback-Pfad in `docs/user/ingest-control.md` §2.2 dokumentiert. Tests: 4 Use-Case-Tests (Provision-False-leaves-empty, Provision-True-Disabled, Adapter-Applied-flows-through, Adapter-Error-is-failed-without-Rollback), 4 HTTP-Handler-Tests (Backwards-Compat, applied, disabled mit hint, failed mit 201), 8 MediaMTX-Adapter-Tests (happy, idempotent 409, auth_failure 401, server_status_500, unreachable, Auth-Token-Header, Rollback-Happy, Rollback-NotFound-OK). Operator-Smoke `make smoke-mediaserver-provision`. **Wieder-Eröffnungs-Trigger**: (a) Operator-Bedarf nach SRS-Adapter (Folge-Item; MediaMTX-Default abdeckt den Lab-Pfad); (b) Folge-Endpoint `POST /api/ingest/streams/{id}/provision` für nachträglichen Server-Sync (Operator-Workflow ohne Stream-Löschen); (c) DELETE-Endpoint für API-Stream + automatischen Rollback (heute Operator-manual via SQLite). |
| R-17 | `InMemoryIssuanceRateLimiter` (`0.12.0`) hielt Token-Bucket-State pro API-Prozess. Ein Multi-Instance-Setup (≥ 2 API-Replicas hinter Load Balancer) würde die Issuance-Quote pro Replica messen — die effektive globale Issuance-Rate wäre bis zu N× höher als konfiguriert. | `apps/api/adapters/driven/auth/redis_issuance_rate_limiter.go`; `apps/api/adapters/driven/auth/sqlite_issuance_rate_limiter.go`; `docs/user/auth.md` §5.4; Plan-0.12.0 §0.6 Threat Model „Issuance-Abuse" | `0.12.5` Tranche 2 (SQLite-Single-Host); `0.12.6` Tranche 7 (Redis-Multi-Host) ✅ | 🟢 | **gelöst** — Multi-Host-Backend geliefert in `0.12.6` Tranche 7: `RedisIssuanceRateLimiter` implementiert `driven.IssuanceRateLimiter` über ein atomares Lua-`EVAL`-Script, das beide Buckets (global + project) plus Refund bei project-deny in einem einzigen Redis-Roundtrip behandelt. Keys `mtrace:issuance:global` / `mtrace:issuance:project:<id>`, Default-TTL 24 h. ENV-Selektor `MTRACE_AUTH_ISSUANCE_LIMITER=memory|sqlite|redis` mit Pflicht-ENV `MTRACE_REDIS_ADDR` (+ optional `_AUTH`/`_DB`); `memcached` bleibt Folge-Item gemeinsam mit R-22-Memcached. Fail-Mode default fail-closed (Outage → 429); opt-in fail-open via `MTRACE_AUTH_ISSUANCE_FAIL_OPEN=1` mit lokalem In-Memory-Fallback (Schalter gilt gemeinsam mit R-22, kein halb-fail-closed Pfad möglich). Sieben Adapter-Tests gegen `miniredis` (Cross-Instance-Sharing, Project-Isolation, globaler Refund-Pfad, Fail-Closed/Fail-Open bei Outage, Nil-Client, Context-Cancel). **Operator-Smoke**: `make smoke-issuance-multi-host` (opt-in). **Wieder-Eröffnungs-Trigger**: (a) Operator-Bedarf nach Memcached-Backend (Folge-Item gemeinsam mit R-22 reaktivieren); (b) Operator-Report über Token-Mint-Welle trotz fail-closed-Default (Indiz auf Redis-Auth-/Network-Setup-Bug). |
| R-22 | Origin-/IP-nahes Rate-Limiting fehlte strukturell. `0.12.0`/`0.12.5`-Limiter waren beide **Project-Token-skopiert** — IP-basierte Bursts konnten den Project-Layer umgehen. | `apps/api/hexagon/port/driven/origin_rate_limiter.go`; `apps/api/adapters/driven/auth/in_memory_origin_rate_limiter.go`; `apps/api/adapters/driven/auth/redis_origin_rate_limiter.go`; `apps/api/adapters/driving/http/origin_rate_limit.go`; `docs/user/auth.md` §5.9 | `0.12.6` Tranche 6 (Memory); `0.12.6` Tranche 7 (Redis) ✅ | 🟢 | **gelöst** — Memory- und Redis-Backend geliefert in `0.12.6` Tranche 6 + 7. Memory: `InMemoryOriginRateLimiter` (Single-Bucket pro Key, Cap 20 / Refill 5/s, opportunistische Idle-Eviction). Redis: `RedisOriginRateLimiter` (atomares Lua-Script auf `mtrace:origin:<key>`, TTL 10 min; teilt sich den Redis-Server mit dem Issuance-Limiter, eigener Key-Prefix). HTTP-Middleware `originRateLimitMiddleware` vor `POST /api/auth/session-tokens` und `POST /api/playback-events`; Reihenfolge `Origin-Limit → Project-Limit`. ENV-Selektor `MTRACE_ORIGIN_RATE_LIMITER=disabled|memory|redis` mit Boot-Validator-Rejects für `sqlite` (Multi-Host-unsicher) und `memcached` (gemeinsames Folge-Item mit R-17). `MTRACE_TRUST_FORWARDED_FOR`-Opt-in für Reverse-Proxy. Fail-Mode-Schalter `MTRACE_AUTH_ISSUANCE_FAIL_OPEN=1` gilt gemeinsam mit dem Issuance-Limiter (kein halb-fail-closed Pfad). Reject-Body `{"error":"origin_rate_limited"}` (429). Sieben Adapter-Tests (Memory: Burst+Refill, Key-Isolation, Empty-Key-No-Op, Disabled-Bucket, Nil-Receiver, Eviction, Context-Cancel) plus sechs Redis-Adapter-Tests gegen `miniredis` (Cross-Instance-Sharing, Empty-Key, Fail-Closed/Fail-Open, Nil-Client, Context-Cancel). HTTP-Roundtrip-Tests (Burst-429, Disabled-Pass-Through, XFF-Trust). `make smoke-origin-rate-limit` (Memory) + `make smoke-issuance-multi-host` (Redis) als Operator-Smokes. **Wieder-Eröffnungs-Trigger**: Operator-Bedarf nach Memcached-Backend (Folge-Item gemeinsam mit R-17). |
| R-20 | `0.12.0` Signing-/Project-Token-Secrets lagen im **lokalen ENV/File-Konfigurations-Pfad** (`MTRACE_AUTH_SIGNING_KEY` plus SQLite-`project_token_generations`). RAK-71 schloss KMS-/Vault-/Cloud-Secret-Manager-Integration als Muss-Pfad explizit aus — Compliance-Pfade (PCI/SOC2) und Multi-Instance-Setups fehlten dadurch. | `apps/api/hexagon/port/driven/auth_secret_backend.go`; `apps/api/adapters/driven/auth/env_secret_backend.go`; `apps/api/adapters/driven/auth/vault_secret_backend.go`; `apps/api/adapters/driven/auth/kms_secret_backend.go`; `apps/api/cmd/api/main.go` (`buildAuthSecretBackend` + `buildKMSDecrypter`); `docs/user/auth.md` §5.5 (Compliance-Block); `spec/lastenheft.md` §13.15 RAK-79 | `0.12.5` Tranche 3 (Skelett); `0.12.6` Tranche 8 (Production-Pfade) ✅ | 🟢 | **gelöst** — Production-Auth-Pfade geliefert in `0.12.6` Tranche 8: Vault-Adapter um drei Auth-Methoden erweitert: `token` (Backwards-Compat), `approle` (zwei-Phasen-Login `/v1/auth/<mount>/login` mit `role_id`+`secret_id`), `kubernetes` (Pod-ServiceAccount-Token aus `/var/run/secrets/kubernetes.io/serviceaccount/token` + `/v1/auth/<mount>/login`). ENV-Selektor `MTRACE_AUTH_VAULT_AUTH_METHOD=token|approle|kubernetes`; pro Methode konfigurierbare Pflicht-ENV (siehe `auth.md` §5.5). Zusätzlich neuer `KMSSecretBackend` mit `KMSDecrypter`-Interface als vendor-neutrale Abstraktion; ENV-Selektor `MTRACE_AUTH_SECRET_BACKEND=kms` aktiviert den Pfad. Production-AWS-SDK-v2-Wiring ist Folge-Item (operator-injected Decrypter); für Lab-Smokes existiert `MTRACE_AUTH_KMS_LAB_MODE=1` als Pass-Through-Decrypter (NICHT für Production, Boot-Log markiert das explizit). `MTRACE_AUTH_SECRET_BACKEND_REFRESH_SECONDS`-ENV gelesen + Boot-Log nennt Status; periodischer Refresh-Loop ist Folge-Item (Werte > 0 sind heute no-op mit Operator-Warning). **Compliance-Doku** in `auth.md` §5.5: PCI/SOC2-Hinweise zu Key-Material-Sichtbarkeit (keine Logs), Fail-closed-Vertrag, TLS-Anforderung, Audit-Trail-Konvention. Tests: 11 Tests in `secret_backend_t8_test.go` (AppRole-HappyPath/MissingSecret/WrongSecret, K8s-HappyPath/MissingJWTFile/EmptyJWT, Unsupported-Auth-Method, KMS-HappyPath/DecryptError/LabPassThrough/MissingConfig/NilDecrypter/EncryptedKeysFromPath). Operator-Smokes `make smoke-vault-approle` + `make smoke-kms-skeleton`. **Wieder-Eröffnungs-Trigger**: (a) Operator-Bedarf nach produktiver AWS-SDK-v2-KMS-Anbindung (Folge-Item); (b) Refresh-Loop für automatische Schlüsselrotation ohne Restart (Folge-Item); (c) PCI/SOC2-Audit gegen ein konkretes Operator-Setup (Lieferung allein erfüllt den Audit nicht). |
| R-12 | WebRTC-`getStats()`-Schema-Drift zwischen Browser-Major-Versionen (Chromium/Firefox/Safari) kann die in `spec/telemetry-model.md` §3.5 spezifizierte WebRTC-Aggregat-Allowlist (`connection_state`, `ice_state`, `dtls_state`) und das `getStats()`-Subset (Muss-/Soll-Felder pro `RTCStatsType`-Gruppe) ungültig machen. **Stand `0.9.0` Tranche 1**: produktive WebRTC-Telemetrie ist seit `0.8.0` ausgeliefert; mit `0.9.0` Tranche 1 ist der Browser-Drift-Smoke `tests/e2e/webrtc-stats-drift.spec.ts` plus `make smoke-webrtc-stats-drift` plus Nightly-Workflow `.github/workflows/webrtc-drift.yml` aktiv und detektiert einen Schema-Drift automatisiert. | `docs/planning/done/plan-0.8.0.md` §4 Tranche 3 (RAK-53/RAK-54); `docs/planning/done/plan-0.9.0.md` §2 Tranche 1 (RAK-56); `spec/telemetry-model.md` §3.5; `tests/e2e/webrtc-stats-drift.spec.ts`; `scripts/smoke-webrtc-stats-drift.sh`; `.github/workflows/webrtc-drift.yml`; `packages/player-sdk/src/adapters/webrtc/sampling.ts`; `apps/api/adapters/driven/metrics/webrtc_metrics.go` | automatisiert detektiert, Drift bricht den Drift-Smoke | ⬜ | **Detektion**: Nightly-Workflow `webrtc-drift.yml` (cron `30 3 * * *`, plus `workflow_dispatch`) führt `make smoke-webrtc-stats-drift` gegen Chromium und Firefox aus dem Playwright-Bundle aus; bei Schema-Drift (fehlendes Muss-Feld in `RTCStatsType`-Gruppe oder unbekannter Enum-Wert für `connectionState`/`iceConnectionState`/`dtlsState`) bricht der Smoke release-blockierend, und ein Issue wird optional via `secrets.DRIFT_AUTO_ISSUE=1` automatisch erstellt. **Reaktionspfad** bei Befund: `webrtc.*`-Allowlist in `spec/telemetry-model.md` §1.4 + §3.5.2, `contracts/event-schema.json#reserved_meta_keys` und `packages/player-sdk/src/adapters/webrtc/sampling.ts#collectAggregate` synchron aktualisieren; danach `make smoke-webrtc-stats-drift` lokal grün ziehen. **Manuell-Review entfällt** seit Tranche 1; vor jedem Release-Tag genügt ein Blick auf den letzten Nightly-Lauf. **WebKit/Safari** ist opt-in via `MTRACE_WEBRTC_DRIFT_BROWSERS=chromium,firefox,webkit` — Trigger zur Aktivierung wäre ein Safari-Major-Release oder ein Operator-Bedarf an Safari-Drift-Detektion. Wieder-Reduktion auf „bei Triggerschwelle" nur, wenn der WebRTC-Adapter wieder deaktiviert wird. **Trigger-Stand 2026-05-10 (`0.12.1` T1)**: Drift-Smoke unverändert grün; vor `0.12.1`-Tag erneuter Check des letzten `webrtc-drift.yml`-Nightly-Runs als Teil der Wave-2-Verdict-Pflicht aus `releasing.md` §3.1. **Trigger-Re-Eval 2026-05-13 (`0.18.0` Tranche 1)**: letzte drei `webrtc-drift.yml`-Runs gruen; neuester gepruefter Run `25769902117` (`2026-05-13T00:13:10Z`, Head `2f75331bc5bd1cee37983972b919c98444770d1d`). Kein Safari-/WebKit-Pflicht- oder Operator-Trigger gefunden; Status bleibt `deferred` mit Nightly-Detector. |
| R-24 | Load-Smoke-Nightly (`load-smoke.yml`) ist nicht-blockierend; ein Verdict-Step färbt den Job rot nur bei Hard-FAIL (Smoke-Exit 1: stiller Verlust / SLO-Bruch / Fehlerquote), INCONCLUSIVE (Exit 3) bleibt grün. Damit ist die Sichtbarkeitslücke des reinen `continue-on-error` geschlossen — aber eine **einzelne** lastabhängig-flaky SLO-Verfehlung färbt den Job sofort rot (Noise), und es gibt kein Debounce / keine debounced Auto-Issue. | `.github/workflows/load-smoke.yml` (Run+Verdict-Step); `scripts/smoke-load.sh` (Exit-Codes 1/2/3); `docs/planning/in-progress/extra-gates.md` §3.9 | Folge-Item | ⬜ | **Mitigation aktuell**: Verdict-Step trennt Hard-FAIL (rot) von INCONCLUSIVE/Flake (grün); Verdikt zusätzlich in Job-Summary + Artefakt. **Eröffnungs-Trigger**: wiederholte Flake-Rot-Läufe ohne echte Regression → Debounce (Job rot erst nach N Hard-FAILs in Folge, analog Mutation-Promote-Schwelle) oder debounced Auto-Issue nach N Hard-FAILs. Bis dahin liest der Operator Summary/Artefakt. |
| R-25 | Die Readback-Reconciliation in `smoke-load.sh` (Gate „kein stiller Verlust": `persisted >= accepted`) zählt die persistierten Events, indem sie den Detail-Endpoint **seitenweise zu je 1000 Events über HTTP paginiert** — über alle Sessions. Das ist **O(N) über das Gesamt-Eventvolumen**: Beim 4h-Dispatch-Soak (`27628293077`, 45,68 Mio akzeptierte Events / 20 Sessions) waren das ~45,7k sequentielle Reads → **~2h reine Readback-Zeit ohne jede Log-Ausgabe** (Command-Substitution `persisted="$(…)"`, druckt erst am Ende). Zusammen mit den 4h Last sprengte der „Run load smoke"-Step den **GitHub-6h-Job-Cap** → Job gecancelt, **Verdict-Step skipped, kein Load-Readiness-Verdict** (ADR-0005-Trigger #3 ungemessen). Für kurze Smokes (≤ einige 100k Events) unkritisch; nur der lange Soak-Pfad ist betroffen — der Nightly-SLO-Smoke (`27661025846`, open-loop kurz) bleibt grün. | `scripts/smoke-load.sh` (Readback-Block Z. 211–244); `.github/workflows/load-smoke.yml` (6h-Job-Cap); Run `27628293077` (cancelled, 6h); `docs/planning/in-progress/plan-0.22.5-load-smoke.md` §5 Soak-Dispatch-Log | Tranche 4 (blockiert Verdict) | ⬜ | **Fix-Optionen**: (a) im Autostart-Pfad persistierte Events per **direktem `SELECT count(*)` gegen die Container-SQLite** statt HTTP-Pagination — O(1), authentischer Persistenz-Beleg (deckt sich mit der Skript-Absicht „die TATSÄCHLICH in `playback_events` liegenden Events"), HTTP-Pagination als Fallback für `SMOKE_LOAD_AUTOSTART=0`; (b) Readback-Walk mit Fortschritts-`log()` + harter Zeitobergrenze, damit ein Hang sichtbar wird statt stiller 2h; (c) Soak-`DURATION` so wählen, dass Last + Readback < 6h bleiben. **Eröffnungs-Trigger**: jetzt — blockiert das Tranche-4-Verdict, bis ein Soak die ≥-10-Mio-Schwelle innerhalb der 6h durchläuft. Bis dahin liefert ein Lauf nur die k6-Ingest-Leg (kein Reconciliation-/Retention-Urteil). |

### 1.2 Aufgelöste / nicht eingetretene Risiken

Sichtbar als Historie. Spalten kompakt (Risiko, Auflösung, Verweis).

| Kennung | Risiko (Kurzform) | Auflösung | Verweis |
|---|---|---|---|
| OS-1 ⬛ | OAuth/OIDC/SSO + Rollenmodell + User-/Org-/Admin-Verwaltung | `0.12.1` Tranche 1: strukturell nicht trackbar ohne konkreten Bedarf — RAK-71-Out-of-Scope-Stand bleibt normativ in Lastenheft §13.14; ein Tracking-Trigger („Multi-Tenant-/Regulated-Requirement") ist zu vage für ein R-N-Item. Reaktivierung in `0.15.0` oder einem Folgeplan, wenn ein Operator/Stakeholder konkreten Bedarf signalisiert. | `done/plan-0.12.0.md` §10; `in-progress/plan-0.12.1.md` §2 OS-Schärfung |
| OS-2 ⬛ | Produktive Admin-/Operator-UI für Auth/Policy-Management | `0.12.1` Tranche 1: implizit mit OS-1 verzahnt; ohne OS-1-Kontext kein eigenständiges Risiko. Mit OS-1 streichen. | `done/plan-0.12.0.md` §10; `in-progress/plan-0.12.1.md` §2 |
| OS-3 ⬛ | Produktive MediaMTX-/SRS-Auth-Hook-Brücke (signierte Publish-Tokens) | `0.12.1` Tranche 1: Duplikat zu **R-14** (validate-key-vs-Auth-Bridge-Risiko). R-14 ist die getrackte Form mit operator-observablem Trigger; OS-3 als Folge-Scope-Notiz aus `done/plan-0.12.0.md` §10 abgelöst. | R-14 in §1.1; `done/plan-0.12.0.md` §10 |
| OS-4 ⬛ | KMS/Vault/Cloud-Secret-Manager-Backend | `0.12.1` Tranche 1: Duplikat zu **R-20** (Production-Secret-Backend). R-20 ist die getrackte Form, `plan-0.12.5.md` Tranche 3 (RAK-79) liefert das Driven-Port-Skelett. OS-4 abgelöst. | R-20 in §1.1; `done/plan-0.12.0.md` §10; `done/plan-0.12.5.md` Tranche 3 |
| OS-5 ⬛ | Multi-Deployment / Multi-Replica Secret- und Issuance-Mechanik | `0.12.1` Tranche 1: Duplikat zu **R-17** (Multi-Replica-Issuance-Limiter) und **R-18** (Multi-Key-Rotation). Beide Items sind in §1.1 mit operator-observablem Trigger; `plan-0.12.5.md` Tranche 1+2 liefert die Code-Pfade. OS-5 abgelöst. | R-17 + R-18 in §1.1; `done/plan-0.12.0.md` §10; `done/plan-0.12.5.md` Tranchen 1+2 |
| R-19 ⬛ | War als ein README-formuliertes Risiko angelegt; in `5798473` (`0.12.0` T5 Phase A) wurde der „Was m-trace nicht ist"-README-Block entfernt. | `0.12.1` Tranche 1: nicht eingetreten — Risiko-Grundlage existiert nicht mehr im Repo. Lücke in der R-N-Nummerierung bewusst beibehalten als historischer Marker. | Commit `5798473` |
| R-14 🟢 | `validate-key` ist kein produktiver Media-Server-Auth-Pfad — Operator könnte ihn fälschlich als Auth-Ersatz konfigurieren. | `0.12.5` Tranche 5 (RAK-81): `MediaMTXAuthHookHandler` bedient `POST /api/ingest/auth-hook` als MediaMTX-`externalAuth`-Bridge; mapped `user`→Project, `password`→Stream-Key Klartext, `path`→Stream-ID. Trust-Boundary über Netzwerk-Isolation (Operator-Verantwortung). Wieder-Eröffnung bei Bedarf an Read-Auth, eigenem Publish-Token-Lifecycle oder mTLS. | `apps/api/adapters/driving/http/mediamtx_auth_hook.go`; `docs/user/auth.md` §5.7; `done/plan-0.12.5.md` Tranche 5 |
| R-16 🟢 | Lokales Lifecycle-Eventmodell aus `0.11.0` ohne ausgehende Webhook-Zustellung an externe Konsumenten. | `0.12.5` Tranche 5 (RAK-82): Driven-Port `OutboundWebhookDispatcher` + `webhooks.HTTPDispatcher`-Adapter mit HMAC-SHA-256, Exponential-Backoff-Retry (3×: 100/200/400 ms), Dead-Letter via `ErrOutboundWebhookExhausted`. ENV-Selektor `MTRACE_OUTBOUND_WEBHOOK_URL/_SECRET`; leere URL → No-Op. Integration im `IngestControlService.RecordLifecycleEvent`. Wieder-Eröffnung bei Bedarf an Project-spezifischen URLs, persistenter DLQ oder ENV-konfigurierbaren Retry-Parametern. | `apps/api/hexagon/port/driven/outbound_webhook.go`; `apps/api/adapters/driven/webhooks/http_dispatcher.go`; `docs/user/auth.md` §5.8; `done/plan-0.12.5.md` Tranche 5 |
| R-18 🟢 | `MTRACE_AUTH_SIGNING_KEY`-Rotation ohne dokumentierten Operator-Workflow / Code-Pfad. | `0.12.5` Tranche 1 (RAK-78): `MultiKeySigningResolver` (Rename des `0.12.0`-`StaticSigningKeyResolver`) + `ParseSigningKeysEnv` mit ENV-Schema `MTRACE_AUTH_SIGNING_KEYS=kid_a:b64,kid_b:b64` + `_ACTIVE_KID`. Single-Key-Backwards-Compat bleibt als degenerierter `len(keys)==1`-Pfad. Operator-Runbook in `auth.md` §5.3 (Doku-Stand aus `0.12.1` Tranche 2). `make smoke-key-rotation` wrapt End-to-End-Test. Wieder-Eröffnung bei Operator-Bug-Report im Multi-Replica-Pfad. | `apps/api/adapters/driven/auth/multi_key_signing_resolver.go`; `apps/api/adapters/driven/auth/signing_keys_env.go`; `docs/user/auth.md` §5.3; `done/plan-0.12.5.md` Tranche 1 |
| R-21 🟢 | `/api/ingest/*` strikt operator-only (RAK-74-Scope-Cut) — keine kontrollierte Browser-Konsumenten-Aufhebung. | `0.12.5` Tranche 4 (RAK-80): `domain.BrowserIngestPolicy` (`Enabled`/`CORSAllowlist`/`CSRFRequired`/`OriginPin`) als Feld in `ProjectPolicy`; `browserIngestPreflightHandler` für Origin-Match-Routing zwischen RAK-74 (strikt) und RAK-80 (kontrolliert offen); `browserIngestEnforcement`-Middleware vor `/api/ingest/*`-POSTs mit `ingest_browser_origin_not_allowed`/`_pin_mismatch`/`_csrf_missing`-403s. CSRF-Skelett (Header-Anwesenheits-Check) — produktive Anti-CSRF-Token-Library bleibt potenzielles Wieder-Eröffnungs-Item. | `apps/api/hexagon/domain/auth_project_policy.go`; `apps/api/adapters/driving/http/browser_ingest_enforcement.go`; `docs/user/auth.md` §5.6; `done/plan-0.12.5.md` Tranche 4 |
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
- **Teilweise Lösungen**: Liefert ein Release einen Code-Pfad,
  ohne den Triggerkern vollständig aufzulösen (z. B. Skelett-
  Adapter, Single-Host-Variante eines Multi-Host-Risikos), bleibt
  der Eintrag in §1.1 mit **Status `⬜ offen`**. Die Mitigations-
  Spalte beginnt dann mit dem Präfix `**teilweise gelöst** —`
  und nennt explizit (a) was geliefert wurde, (b) den
  **Resttrigger**, der das Item endgültig in §1.2 verschieben
  würde, und (c) den Folge-Plan-Verweis. Erst wenn der Resttrigger
  ausgelöst und vollständig adressiert ist, wandert der Eintrag
  mit 🟢 in §1.2.
- Done-Pläne (`docs/planning/done/plan-X.Y.Z.md`) sind nach Tag
  Release-Historie und werden **nicht** im Nachhinein editiert,
  um OS-Items oder Trigger-Schärfungen widerzuspiegeln. Solche
  Folge-Aktionen passieren ausschließlich in diesem Backlog
  und/oder in einem Folge-Plan; der Done-Plan bleibt als
  Snapshot des damaligen Folge-Scope-Stands sichtbar.
- Stand-Datum im Header wird beim nächsten substanziellen Update
  einer §1.1-Zeile mitgepflegt.
