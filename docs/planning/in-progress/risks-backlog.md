# Risiken-Backlog

> **Stand**: 2026-05-11 — `0.12.6` (Auth-/Ingest-Folge-Items-Minor)
> in **Tranche 0 aktiv** seit 2026-05-11; Lastenheft-Patch `1.1.17`
> mit RAK-83..RAK-90 in §13.16 persistiert. Vorgänger `0.12.5`
> released 2026-05-11 (Tag `v0.12.5`); `0.12.1` released
> 2026-05-10 (Tag `v0.12.1`); `0.12.0` released 2026-05-10
> (Tag `v0.12.0` auf `e3457fa`).
>
> ### Aktuelle Phase
>
> `0.12.6` adressiert alle neun nach `0.12.5` offenen R-N-Items in
> Tranchen 1–9 (Plan in
> [`plan-0.12.6.md`](./plan-0.12.6.md)). T0-Entscheidung Option A
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
> Lastenheft-Stand: `1.1.17` (`0.12.6`-Patch persistiert in T0).
> `0.13.0` bekommt voraussichtlich `1.1.18` und RAK-91..RAK-95
> in §13.17, da RAK-83..RAK-90 mit `0.12.6` belegt sind.
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
>   Bedarf kein Tracking-Trigger sinnvoll ist; Reaktivierung als
>   `plan-0.14.x` o. ä. bei Stakeholder-Bedarf.
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
>   Auth-/Ingest-Adapter)
> - `docs/planning/in-progress/roadmap.md` §4
> - `docs/planning/done/plan-0.12.0.md` (Folge-ADRs);
>   `docs/planning/done/plan-0.12.1.md` (Patch: Trigger-Re-Eval +
>   Operator-Doku);
>   `docs/planning/done/plan-0.12.5.md` (Minor:
>   Auth-/Ingest-Adapter R-14/R-17/R-18/R-20/R-21, optional R-16)

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
| R-5 | Time-Skew-Persistenz auf Event-Ebene fehlt: `0.4.0` setzt `mtrace.time.skew_warning=true` als Span-Attribut (siehe `spec/telemetry-model.md` §2.5/§5.3), aber die Schema-Spalte und Dashboard-Anzeige sind explizit deferred. Folge: skew-betroffene Events sind im Read-Pfad (Dashboard ohne Tempo) nicht sichtbar markiert; Operator muss in Tempo schauen. | `plan-0.4.0.md` §3.1 | bei Triggerschwelle | ⬜ | **Stand `0.8.0` Closeout / `0.9.0`-Plan offen**: Triggerschwelle weiterhin nicht ausgelöst — weder SRT-Health-View, WebRTC-Lab noch der WebRTC-Adapter haben einen neuen Player-Event-Skew-Bezug; bestehender Aktivierungspfad bleibt unverändert. Trigger wie zuvor: ≥ 5 Spans mit `mtrace.time.skew_warning=true` außerhalb von Synthetik-Tests innerhalb einer Lab-Woche, oder Operator-Report. **Trigger-Stand 2026-05-10 (`0.12.1` T1)**: nicht ausgelöst — kein neuer Skew-Beleg seit `0.8.0`-Closeout, kein Operator-Report. |
| R-7 | `SessionsService.ListSessions` lädt `network_signal_absent[]` pro Session-Page-Eintrag einzeln (`ListBoundariesForSession` N+1). Bei Hard-Cap 1000 Sessions pro Page sind das im Worst Case 1000 SQL-Round-Trips ohne gemeinsamen Tx-Snapshot (jede Query öffnet eine eigene Tx-Boundary). Schreibpfad (`POST /api/playback-events`) und Detail-Read (`GET /api/stream-sessions/{id}`) sind nicht betroffen. Wahrscheinlichkeit moderat (Lab-typisch wenige Sessions, Production unbekannt); Auswirkung: spürbare List-Latenz, kein funktionaler Bug. | `plan-0.4.0.md` §4.4 D3 (Review-N-1) | bei Triggerschwelle | ⬜ | **Stand `0.8.0` Closeout / `0.9.0`-Plan offen**: Triggerschwelle nicht ausgelöst. WebRTC-Pfade (`metrics_sampled` mit `webrtc.*`-Meta in `0.8.0`) gehen nicht durch den `network_signal_absent`-Pfad; das Risiko bleibt ausschließlich auf den `stream-sessions`-List-Read beschränkt. Aktivierungspfad und Bulk-Read-Port-Vorschlag unverändert. **Trigger-Stand 2026-05-10 (`0.12.1` T1)**: nicht ausgelöst — kein List-Latenz-Bericht ≥ 200 ms p95 seit letztem Stand. |
| R-9 | Die Observability-Smoke-Whitelist für Infrastruktur-Labels (`__name__`, `instance`, `job`) ist bewusst Compose-Lab-spezifisch. Eine spätere K8s-Smoke-Stage würde übliche Labels wie `pod`, `namespace` oder `container` wahrscheinlich als False Positive werten. | `plan-0.4.0.md` §7.4 (Review F-2) | K8s-Smoke-Einführung | ⬜ | **Stand `0.8.0` Closeout / `0.9.0`-Plan offen**: keine Änderung am Compose-Lab-Schema, aber die Aggregat-Allowlist ist seit `0.6.0`/`0.8.0` gewachsen — `mtrace_srt_health_*` (`health_state`, `source_status`, `source_error_code`) plus jetzt `mtrace_webrtc_*` (`connection_state`, `ice_state`, `dtls_state` plus drei label-freie Counter). Eine künftige K8s-Smoke-Stage muss diese drei Themenblöcke gemeinsam adaptieren; Folgeplan zur K8s-Anbindung ist Voraussetzung (z. B. MVP-42-Folge-Plan), Trigger und Mitigation strukturell unverändert. **Trigger-Stand 2026-05-10 (`0.12.1` T1)**: nicht ausgelöst — K8s-Smoke-Stage frühestens mit `plan-0.13.0.md` (MVP-42); R-9 wandert dann in dessen Tranche-Scope. |
| R-10 | Sampling-Vollständigkeitsnachweis für `sampleRate < 1` fehlt: gesampelte Events verbrauchen keine `sequence_number`, deshalb kann der Server eine Sampling-Lücke nicht von einem echten Verlust unterscheiden. Folge: vollständige Timeline-Abnahme und alle E2E-Smokes laufen mit `sampleRate = 1`; Sampled-Sessions sind in `0.4.0` ausschließlich über dokumentierte Konfiguration und Operator-Hinweis als „sampled" markiert, nicht über serverseitige Lückenerkennung. | `plan-0.4.0.md` §8.3 (Beschluss Variante (b)) | bei Triggerschwelle | ⬜ | **Stand `0.8.0` Closeout / `0.9.0`-Plan offen**: Triggerschwelle nicht ausgelöst. SRT-Health- und WebRTC-`metrics_sampled`-Pfade nutzen eigene Sample-Schlüssel (SRT: `bytesReceived`-Source-Sequence; WebRTC: `webrtc.sample_id` pro `peer_connection_run_id`) — R-10 bleibt auf den Player-Event-Sampling-Pfad beschränkt. Mitigation unverändert. **Trigger-Stand 2026-05-10 (`0.12.1` T1)**: nicht ausgelöst — kein neuer Sampling-Lücken-Beleg außerhalb Player-Pfad. |
| R-11 | SRT-Health-Detail-Pagination ist in `0.6.0` Sub-3.3 als ErrNotImplemented gestubbed. Der HTTP-Pfad `GET /api/srt/health/{stream_id}` liefert nur `samples_limit` (default 100, max 1000) ohne Cursor; eine länger laufende Lab-/Production-Instanz mit > 1000 persistierten Samples pro Stream kann das ältere Drittel nicht mehr abfragen. Auswirkung im `0.6.0`-Lab gering (Lab läuft typisch < 1 h und akkumuliert wenige Samples), wird aber bei dauerhaftem Betrieb relevant. | `plan-0.6.0.md` §4 Sub-3.3 (Adapter wirft `errors.New("cursor pagination not yet implemented")`); `spec/backend-api-contract.md` §7a.3 (Wire-Format-Definition steht bereits) | bei Triggerschwelle | ⬜ | **Stand `0.8.0` Closeout / `0.9.0`-Plan offen**: Triggerschwelle nicht ausgelöst — Lab-Läufe in `0.7.0`/`0.8.0` haben weiterhin < 1 h Akkumulation. Mitigation: Cursor-Adapter-Pfad implementieren analog `EventRepository`-Cursor (process_instance_id + (ingested_at, id)-Position als opaker Token); Wire-Format ist in §7a.3 schon spezifiziert, also keine Vertrags-Änderung. Triggerschwelle: ≥ 1000 persistierte Samples pro Stream in Lab/Production, ODER Operator-Report über fehlende ältere Samples im Detail-Verlauf. **Trigger-Stand 2026-05-10 (`0.12.1` T1)**: nicht ausgelöst — Lab-Akkumulation < 1000 Samples pro Stream; kein Operator-Report. |
| R-13 | Drei OS-CVEs ohne Upstream-Fix in `node:22-trixie-slim`, geteilt zwischen Dashboard- und Analyzer-Service-Image: `CVE-2025-69720` (ncurses Buffer Overflow in `libtinfo6`/`ncurses-base`/`ncurses-bin`), `CVE-2026-29111` (systemd Arbitrary Code Exec via IPC in `libsystemd0`/`libudev1`) und `CVE-2026-4878` (libcap TOCTOU-Race in `libcap2`). Keine Ausnutzbarkeit im m-trace-Runtime sichtbar (Container ohne TTY/IPC-Mounts, USER node ohne setcap-Pfad), aber statisch HIGH. Hinweis: Analyzer-Service wechselte mit plan-0.8.5 Tranche 1 Closeout von `node:22-alpine` auf `node:22-trixie-slim`, weil musl-libc bei multi-threaded Workloads (libuv-Worker-Pool, V8-GC/JIT) gegenüber glibc spürbar pessimisiert ist; die einheitliche Trixie-Basis vermeidet das. | `plan-0.8.5.md` §2 Tranche 1 Closeout; `.security/vulnignore.yaml`; `apps/dashboard/Dockerfile`, `apps/analyzer-service/Dockerfile` | bei Triggerschwelle | ⬜ | **Mitigation aktuell**: dokumentierte Trivy-Ignores in `.security/vulnignore.yaml` mit 90-Tage-`expires` (2026-08-04); Generator `scripts/render-trivyignore.sh` bricht ab, sobald ein `expires` überschritten ist (erzwungene Re-Review). **Folge-Trigger**: (a) Trixie-Point-Release liefert Fixes — Eintrag entfernen; (b) `expires`-Schwelle erreicht ohne Fix — Re-Review entscheidet zwischen Verlängerung und Base-Image-Wechsel; (c) Distroless-Variante wird vor 1.0 als Folge-Plan evaluiert (`gcr.io/distroless/nodejs22-debian12` bringt nur Node-Binär ohne npm/Debian-CLI-Surface, eliminiert die drei CVEs strukturell, behält glibc). Aktuell Lab-/Demo-Pfad — keine Production-Kontroll-Plane. **Trigger-Stand 2026-05-10 (`0.12.1` T1)**: nicht ausgelöst — `expires` 2026-08-04 ist 86 Tage entfernt; kein Trixie-Point-Release mit Fix; Re-Review weiterhin auf 2026-08-04 datiert. |
| R-15 | Externe Media-Server-Provisionierung ist im `0.11.0`-Scope explizit ausgeklammert: `POST /api/ingest/streams` produziert kein I/O auf laufenden externen MediaMTX-/SRS-Servern, und `GET /api/ingest/media-server-config` schreibt das Konfigurations-Artefakt nur lokal. Risiko: Operator erwartet, dass ein angelegter Stream automatisch in einer laufenden MediaMTX-Instanz aktiv wird, und sieht Diskrepanz zwischen API-State und Server-Config. | `docs/planning/done/plan-0.11.0.md` §0.1 + §0.6 (`media-server-config`-Klausel); `spec/lastenheft.md` §13.13 RAK-65/RAK-68; `docs/planning/done/plan-0.12.0.md` §0.1 (Out-of-Scope-Bekräftigung) | `0.12.1` Tranche 1 (Trigger-Re-Eval); `0.12.5` Out-of-Scope, später falls Trigger | ⬜ | **Mitigation aktuell**: Wire-Vertrag in `spec/backend-api-contract.md` §3.8 (`media_server_config_unavailable`-Fehler-Code; `config_path`-Hinweis als reines Artefakt-Output), `docs/user/ingest-control.md` und `examples/ingest-control/README.md`. **`0.12.0`-Review (2026-05-10)**: Plan-0.12.0 ändert nichts am `media-server-config`-Wire — externe Provisionierung ist nicht Teil des Auth-/Security-Scopes. R-15 bleibt strukturell unverändert offen. **`0.12.1` Tranche 1**: Trigger-Re-Eval; bei nicht-ausgelöstem Trigger Stand-Notiz, sonst Eskalation in `plan-0.13.x`. **`0.12.5` Out-of-Scope** (siehe [`plan-0.12.5.md`](../done/plan-0.12.5.md) §0.1 Out of Scope): Provisionierung wird erst auf konkreten Operator-Bedarf reaktiviert. **Triggerschwelle** unverändert: Lab-Operator-Bedarf nach automatischer Übernahme oder ein Bug-Report über stille State-Diskrepanz. **Trigger-Stand 2026-05-10 (`0.12.1` T1)**: nicht ausgelöst — kein Operator-Bedarf, kein Bug-Report. |
| R-17 | `InMemoryIssuanceRateLimiter` (`0.12.0`) hält Token-Bucket-State pro API-Prozess. Ein Multi-Instance-Setup (≥ 2 API-Replicas hinter Load Balancer) würde die Issuance-Quote pro Replica messen — die effektive globale Issuance-Rate wäre bis zu N× höher als konfiguriert. Das ist im Lab-/Single-Instance-Pfad kein Problem; bei Production-Scale wird das Bucket umgehbar. | `apps/api/adapters/driven/auth/in_memory_issuance_rate_limiter.go`; `apps/api/adapters/driven/auth/sqlite_issuance_rate_limiter.go`; `apps/api/internal/storage/migrations/V5__auth_issuance_counters.sql`; Plan-0.12.0 §0.6 Threat Model „Issuance-Abuse"; Code-Review (2026-05-10) Should-Address #5/Follow-up | `0.12.5` Tranche 2 (Code) ✅ | ⬜ | **teilweise gelöst** — Single-Host-Shared-Volume-Pfad geliefert in `0.12.5` Tranche 2: `SqliteIssuanceRateLimiter` implementiert `driven.IssuanceRateLimiter` über die `auth_issuance_counters`-Tabelle (Migration V5, atomare Refill+Consume via `BEGIN IMMEDIATE`-Tx, asymmetrischer Refund bei project-deny, opportunistisches TTL-Cleanup mit Default 24h). ENV-Selektor `MTRACE_AUTH_ISSUANCE_LIMITER=memory|sqlite` (globaler Default `memory`); unbekannte Werte (`redis`, `memcached`, …) lehnt der Boot-Validator mit klarer Fehlermeldung ab. `sqlite` braucht `MTRACE_PERSISTENCE=sqlite`; sonst hard-fail. Acht Unit-Tests (Capacity, Refill, Override, asymmetrischer Refund, disabled-Buckets, nil-Receiver, ctx-Cancel, **Cross-Instance-Sharing über zwei `*sql.DB`** auf derselben Datei) plus `make smoke-issuance-replica` als reproduzierbarer Operator-Smoke. RAK-74-Scope-Cut: Limiter hängt nicht vor `/api/ingest/*` (Doku `auth.md` §5.4). **Resttrigger** (weiter offen): (a) echte Multi-Host-Topologie ohne Shared-Volume — braucht Network-Backend-Adapter (Redis/Memcached) als Folge-Plan; (b) Operator-Bug-Report über zu hohe Issuance-Rate trotz aktivem SQLite-Limiter. Erst wenn der Resttrigger ausgelöst und vollständig adressiert ist, wandert R-17 nach §1.2. **Trigger-Stand 2026-05-11 (`0.12.5` T2)**: nicht ausgelöst — Single-Host-Lab-Pfad gedeckt, kein Multi-Host-Setup, kein Operator-Bug-Report. |
| R-22 | Origin-/IP-nahes Rate-Limiting fehlt strukturell. `0.12.0` (R-17) und `0.12.5` (RAK-77 SQLite-Limiter) sind beide **Project-Token-skopiert** — sie limitieren Token-Issuance pro Project, nicht IP-/Origin-skopiert. Ein Angreifer, der von einem Botnet aus identische `mtr_pt_*`-Tokens stiehlt oder Issuance-Endpoints flutet, kann den heutigen Limiter umgehen oder die Replay-Last über viele IPs verteilen. | `docs/planning/done/plan-0.12.0.md` §10 OS-6; `docs/planning/done/plan-0.12.1.md` §2 (`0.12.1` Tranche 1: OS-6 → R-22-Konvertierung) | bei Triggerschwelle | ⬜ | **Mitigation aktuell**: keine — Project-Token-Layer-Limiter (R-17) ist die einzige Bremse. **Auflösungspfad**: optionaler IP-/Origin-Bucket-Limiter als Driven-Port-Adapter (analog `IssuanceLimiterPort`-Konstruktion, aber pro `client_ip` oder `Origin`-Header-Hash). Ein Adapter-Skelett mit Test-Pfad würde sowohl Lab-Single-Instance- als auch das in `R-17` gedeckte Single-Host-Multi-Replica-Setup absichern. Folge-Plan: `plan-0.13.x` (Production / Ops Backends) — IP-Layer ist Production-Konzern; ggf. `plan-0.12.6` als Patch falls der Trigger früh greift. **Triggerschwelle**: (a) Operator-Report über IP-basiertes Last-/Replay-Pattern; (b) Issuance-Abuse-Symptome trotz aktivem `R-17`-Limiter (z. B. wiederholte 429-Bursts aus gleichen Subnets); (c) Compliance-/Audit-Anforderung an Origin-Allowlisting. **Trigger-Stand 2026-05-10 (`0.12.1` T1)**: nicht ausgelöst — kein Operator-Report, Lab-Pfad bleibt Single-Host. |
| R-20 | `0.12.0` Signing-/Project-Token-Secrets liegen im **lokalen ENV/File-Konfigurations-Pfad** (`MTRACE_AUTH_SIGNING_KEY` plus SQLite-`project_token_generations`). RAK-71 schließt KMS-/Vault-/Cloud-Secret-Manager-Integration als Muss-Pfad explizit aus. Risiko: ein produktiver Stack wird mit Lab-Defaults betrieben (Lab-Default-Hard-Fail mitigiert das, R-18 trackt Multi-Key-Rotation), aber ein zentrales Secret-Backend für Multi-Instance-Setups oder Compliance-Bedarf (PCI/SOC2) fehlt. | `apps/api/hexagon/port/driven/auth_secret_backend.go`; `apps/api/adapters/driven/auth/env_secret_backend.go`; `apps/api/adapters/driven/auth/vault_secret_backend.go`; `apps/api/cmd/api/main.go` (`buildAuthSecretBackend`); `docs/user/auth.md` §5.5; `spec/lastenheft.md` §13.15 RAK-79 | `0.12.5` Tranche 3 (Code) ✅ | ⬜ | **teilweise gelöst** — Driven-Port `AuthSecretBackend` (Hexagon-Port) und zwei Adapter-Implementierungen geliefert in `0.12.5` Tranche 3: (a) `EnvSecretBackend` als Default — Backwards-Compat-Wrapper um `ParseSigningKeysEnv`, signalisiert fehlende Konfiguration über `ErrNoSecretConfigured`-Sentinel. (b) `VaultSecretBackend` als externes Adapter-Skelett — minimaler `net/http`-Client gegen Vault KV-v2 (`/v1/<mount>/data/<path>`) mit `X-Vault-Token`-Authentication, bewusst ohne `hashicorp/vault/api`-Dependency. Pflicht-ENV `MTRACE_AUTH_VAULT_ADDR/_TOKEN/_PATH`; optionale Feld-Aliase `MTRACE_AUTH_VAULT_KEYS_FIELD`/`_ACTIVE_KID_FIELD`. Boot-Time-Load, kein periodischer Refresh, fail-closed bei Backend-Outage. ENV-Selektor `MTRACE_AUTH_SECRET_BACKEND=env|vault` in `main.go#buildAuthSecretBackend`; Default `env`. `kms` wird vom Boot-Validator mit klarer „follow-up item"-Meldung abgelehnt. Tests: 4 ENV-Adapter-Tests + 7 Vault-Adapter-Tests (httptest-Mock: happy path, missing config, ungültiger Pfad, 401, 404, malformed JSON, missing field, custom field aliases). **Resttrigger** weiter offen: (a) erste Operator-Anbindung an produktives Vault oder KMS (Lab-Pfad mit `vault dev` ist gedeckt, produktiver Pfad mit AppRole/IAM-Auth nicht); (b) Compliance-Audit (PCI/SOC2) — Skelett-Lieferung allein erfüllt das nicht; (c) KMS-Adapter (AWS-KMS) als additiver Folge-Adapter. **Trigger-Stand 2026-05-11 (`0.12.5` T3)**: nicht ausgelöst — kein Multi-Replica-Setup, kein Compliance-Audit, kein Operator-Bedarf für KMS. |
| R-12 | WebRTC-`getStats()`-Schema-Drift zwischen Browser-Major-Versionen (Chromium/Firefox/Safari) kann die in `spec/telemetry-model.md` §3.5 spezifizierte WebRTC-Aggregat-Allowlist (`connection_state`, `ice_state`, `dtls_state`) und das `getStats()`-Subset (Muss-/Soll-Felder pro `RTCStatsType`-Gruppe) ungültig machen. **Stand `0.9.0` Tranche 1**: produktive WebRTC-Telemetrie ist seit `0.8.0` ausgeliefert; mit `0.9.0` Tranche 1 ist der Browser-Drift-Smoke `tests/e2e/webrtc-stats-drift.spec.ts` plus `make smoke-webrtc-stats-drift` plus Nightly-Workflow `.github/workflows/webrtc-drift.yml` aktiv und detektiert einen Schema-Drift automatisiert. | `docs/planning/done/plan-0.8.0.md` §4 Tranche 3 (RAK-53/RAK-54); `docs/planning/done/plan-0.9.0.md` §2 Tranche 1 (RAK-56); `spec/telemetry-model.md` §3.5; `tests/e2e/webrtc-stats-drift.spec.ts`; `scripts/smoke-webrtc-stats-drift.sh`; `.github/workflows/webrtc-drift.yml`; `packages/player-sdk/src/adapters/webrtc/sampling.ts`; `apps/api/adapters/driven/metrics/webrtc_metrics.go` | automatisiert detektiert, Drift bricht den Drift-Smoke | ⬜ | **Detektion**: Nightly-Workflow `webrtc-drift.yml` (cron `30 3 * * *`, plus `workflow_dispatch`) führt `make smoke-webrtc-stats-drift` gegen Chromium und Firefox aus dem Playwright-Bundle aus; bei Schema-Drift (fehlendes Muss-Feld in `RTCStatsType`-Gruppe oder unbekannter Enum-Wert für `connectionState`/`iceConnectionState`/`dtlsState`) bricht der Smoke release-blockierend, und ein Issue wird optional via `secrets.DRIFT_AUTO_ISSUE=1` automatisch erstellt. **Reaktionspfad** bei Befund: `webrtc.*`-Allowlist in `spec/telemetry-model.md` §1.4 + §3.5.2, `contracts/event-schema.json#reserved_meta_keys` und `packages/player-sdk/src/adapters/webrtc/sampling.ts#collectAggregate` synchron aktualisieren; danach `make smoke-webrtc-stats-drift` lokal grün ziehen. **Manuell-Review entfällt** seit Tranche 1; vor jedem Release-Tag genügt ein Blick auf den letzten Nightly-Lauf. **WebKit/Safari** ist opt-in via `MTRACE_WEBRTC_DRIFT_BROWSERS=chromium,firefox,webkit` — Trigger zur Aktivierung wäre ein Safari-Major-Release oder ein Operator-Bedarf an Safari-Drift-Detektion. Wieder-Reduktion auf „bei Triggerschwelle" nur, wenn der WebRTC-Adapter wieder deaktiviert wird. **Trigger-Stand 2026-05-10 (`0.12.1` T1)**: Drift-Smoke unverändert grün; vor `0.12.1`-Tag erneuter Check des letzten `webrtc-drift.yml`-Nightly-Runs als Teil der Wave-2-Verdict-Pflicht aus `releasing.md` §3.1. |

### 1.2 Aufgelöste / nicht eingetretene Risiken

Sichtbar als Historie. Spalten kompakt (Risiko, Auflösung, Verweis).

| Kennung | Risiko (Kurzform) | Auflösung | Verweis |
|---|---|---|---|
| OS-1 ⬛ | OAuth/OIDC/SSO + Rollenmodell + User-/Org-/Admin-Verwaltung | `0.12.1` Tranche 1: strukturell nicht trackbar ohne konkreten Bedarf — RAK-71-Out-of-Scope-Stand bleibt normativ in Lastenheft §13.14; ein Tracking-Trigger („Multi-Tenant-/Regulated-Requirement") ist zu vage für ein R-N-Item. Reaktivierung als `plan-0.14.x` o. ä. wenn ein Operator/Stakeholder konkreten Bedarf signalisiert. | `done/plan-0.12.0.md` §10; `in-progress/plan-0.12.1.md` §2 OS-Schärfung |
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
