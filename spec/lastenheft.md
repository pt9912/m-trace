# Lastenheft: m-trace

**Projektname:** m-trace<br>
**Dokumenttyp:** Lastenheft<br>
**Version:** 1.1.24<br>
**Status:** Verbindlich<br>
**Lizenz:** MIT<br>
**Architekturstil:** Mono-Repo mit hexagonaler Architektur<br>
**Primärer Stack:** Go 1.22 (stdlib `net/http`, Prometheus, OpenTelemetry, Distroless-Runtime), SvelteKit, TypeScript, Docker — Backend-Stack entschieden in [`docs/adr/0001-backend-stack.md`](../docs/adr/0001-backend-stack.md).

> **Patch `1.1.24` (OCI Image Publishing für `0.21.0`)**:
> Aktiviert die erste GHCR-Veröffentlichung der drei Runtime-Images
> und führt die neue RAK-Gruppe `RAK-121`..`RAK-125` in §13.23 ein.
> Inhalt: versionierte Images `ghcr.io/pt9912/m-trace-api`,
> `ghcr.io/pt9912/m-trace-dashboard` und
> `ghcr.io/pt9912/m-trace-analyzer-service`, Make-Targets für
> Build/Dry-Run/Publish, ein approval-gated GitHub-Actions-Workflow
> und Release-Dokumentation mit Rollback-Grenzen. Kein `latest`-Tag,
> keine Production-K8s-Pflicht und keine Runtime-/Wire-/Schema-
> Änderung. Patch-Log siehe

>
> **Patch `1.1.23` (Package Publishing für `0.20.0`)**:
> Aktiviert die erste echte GitHub-Packages-Veröffentlichung und führt
> die neue RAK-Gruppe `RAK-116`..`RAK-120` in §13.22 ein. Inhalt:
> publishbare npm-Pakete werden auf den GitHub-Owner-Scope
> `@pt9912` umgestellt, `@pt9912/player-sdk` und
> `@pt9912/stream-analyzer` werden über GitHub Packages
> veröffentlicht, die Apps bleiben `private: true`, und der
> Release-Pfad dokumentiert Dry-Run, produktiven Publish und Rollback-
> Grenzen. Patch-Log siehe

>
> **Patch `1.1.22` (Hardening / Evidence Review für `0.17.0`)**:
> Aktiviert die Folgephase nach `0.16.0` und führt die neue
> RAK-Gruppe `RAK-111`..`RAK-115` in §13.21 ein. Inhalt:
> Szenario D wird als einziger Go-Pfad gewaehlt. `0.17.0` startet als
> Hardening-only-/Evidence-Review des gelieferten HLS-CMAF-Byte-
> Range-Fetch-Slice aus `0.16.0`; Productization, Next Slice oder
> Switch bleiben blockiert, bis  konkrete Belege,
> Testluecken oder Trigger nachweist. Externe Analyzer-API,
> Control-Plane, Postgres-Default, Analytics-Pflichtbackend,
> Production-K8s, weiterer CMAF-/DASH-/LL-CMAF-Scope,
> Codec-Decoding und Player-Laufzeitpfade bleiben deferred. Patch-Log
> 
>
> **Patch `1.1.21` (Selected Product Slice / Analyzer Range Fetch für
> `0.16.0`)**: Aktiviert die Folgephase nach `0.15.0` und führt die
> neue RAK-Gruppe `RAK-106`..`RAK-110` in §13.20 ein. Inhalt:
> Szenario B aus `0.15.0` RAK-104 wird als einziger Go-Pfad
> importiert; der HTTP-Range-/Byte-Range-Loader fuer
> manifest-referenzierte CMAF-Init-/Media-Segmente wird zum
> `0.16.0`-Slice. Der Release liefert HLS-CMAF-Byte-Range-Fetches fuer
> explizite Offsets additiv im bestehenden Analyzer-Binary-Pfad,
> schliesst RAK-106..RAK-110 und bleibt ohne neues Public-Schema.
> Externe Analyzer-API, Control-Plane, Postgres, Analytics,
> Production-K8s, Low-Latency-CMAF, vollstaendige Segmentsets,
> Codec-Decoding und Player-Laufzeitpfade bleiben deferred. Patch-Log
> 
>
> **Patch `1.1.20` (Product Scope / Analyzer Boundary für
> `0.15.0`)**: Aktiviert die Folgephase nach `0.14.0` und führt die
> neue RAK-Gruppe `RAK-101`..`RAK-105` in §13.19 ein. Inhalt:
> Zielgruppenentscheidung (§16.1) als Produktgrenze, Analyzer-API-
> Boundary gegen den bestehenden internen `apps/analyzer-service`
> (`MVP-20`), Control-Plane-Scope als Decision-Record ohne
> Implementierung (`F-132`), Analyzer-Folge-Slice-Zuschnitt für
> `NF-13` und erneute Ops-Trigger-Prüfung für Postgres/Analytics
> (`MVP-40`/`MVP-41`). Backwards-Compat: keine Code-, Wire-,
> Runtime- oder Default-Änderung durch die Aktivierung; SQLite,
> Compose und der interne Analyzer-Service bleiben Standardpfade.
> Externe Analyzer-API, Control-Plane, Postgres-Default,
> Analytics-Pflichtbackend, Production-K8s, OAuth/OIDC/SSO und
> Multi-Tenant-SaaS bleiben ohne eigenen Folgeplan out of scope.
> Patch-Log siehe

>
> **Patch `1.1.19` (Ops Backend Follow-up für `0.14.0`)**:
> Aktiviert die Folgephase nach `0.13.0` und führt die neue
> RAK-Gruppe `RAK-96`..`RAK-100` in §13.18 ein. Inhalt:
> Postgres bleibt nach ADR 0005 als `defer-with-migration-seed`
> geführt (`MVP-40`), Analytics-Backends bleiben triggerbasiert
> deferred (`MVP-41`), K8s-/NF-18-Seed-Hardening bleibt optional
> und nicht production-ready (`MVP-42`, R-9), Devcontainer wird als
> Zusatzpfad validiert (`MVP-43`) und der Release-Guard wird ohne
> automatische Veröffentlichung gehärtet (`MVP-44`). Backwards-
> Compat: SQLite und Compose bleiben Standardpfade; Postgres,
> ClickHouse/VictoriaMetrics/Mimir, Kubernetes und Devcontainer
> werden nicht zu lokalen Pflichtabhängigkeiten. Patch-Log siehe

>
>
> **Patch `1.1.18` (Production / Ops Backends für `0.13.0`)**:
> Aktiviert die Production-/Ops-nahe Minor-Phase nach `0.12.6`
> und führt die neue RAK-Gruppe `RAK-91`..`RAK-95` in §13.17
> ein. Inhalt: Postgres-Entscheidungspfad (`MVP-40`) mit
> Seed-/Defer-Boundaries, Analytics-Backend-Vergleich
> (`MVP-41`: ClickHouse/VictoriaMetrics/Mimir oder gleichwertig),
> K8s-/NF-18-Harmonisierung (`MVP-42`) als optionaler
> Option-Pfad ohne Production-Ready-Zusage, Devcontainer-
> Entscheidung (`MVP-43`) und Release-Automatisierung
> (`MVP-44`) mit verpflichtender manueller Freigabe. Backwards-
> Compat: SQLite bleibt lokaler Standard-Store; Postgres,
> Analytics-Backends, Kubernetes und Devcontainer werden nicht
> zur lokalen Pflichtabhängigkeit. R-9 wird in den `0.13.0`-
> Tranche-3-Scope gezogen, weil eine K8s-Smoke-Stage die
> heutige Compose-Lab-Observability-Label-Allowlist verändern
> kann. **Out of Scope**: vollständiger Production-Kubernetes-
> Betrieb, Managed-Cloud-Betrieb, Multi-Tenant-SaaS-Produkt,
> verpflichtendes Hochvolumen-Analytics-Backend und automatische
> Veröffentlichung ohne explizite Human Approval. Patch-Log siehe

>
>
> **Patch `1.1.17` (Auth-/Ingest-Folge-Items für `0.12.6`)**: Liefert
> die R-N-Folge-Items aus dem `0.12.5`-Closeout in einem
> Minor-Release und führt die neue RAK-Gruppe `RAK-83`..`RAK-90`
> in §13.16 ein. Inhalt: Time-Skew-Persistenz mit Dashboard-Marker
> (`R-5`/RAK-83), `ListSessions`-Bulk-Read-Port als
> N+1-Performance-Fix (`R-7`/RAK-84), Sampling-Vollständigkeits-
> Marker mit Integer-ppm-Persistenz und Immutability nach erstem
> gültigem Wert (`R-10`/RAK-85), SRT-Health-Detail-Cursor-
> Pagination via `samples_cursor`/`next_cursor` plus `400
> cursor_invalid` gemäß [`spec/backend-api-contract.md`](backend-api-contract.md)
> §7a.3/§7a.4 (`R-11`/RAK-86), MediaMTX-Provisionierungs-Adapter
> mit strikt-additivem `provision=true`-Query-Param
> (`R-15`/RAK-87), Redis-basierter Multi-Host-Issuance-Limiter als
> Network-Backend (`R-17`/RAK-88, gemeinsam mit `R-22`-Redis-
> Backend), produktive Vault-AppRole-/IAM-Auth plus KMS-Adapter-
> Skelett (`R-20`/RAK-89, hebt den `0.12.5`-Skelett-Stand auf
> produktive Anbindung), Origin-/IP-Rate-Limiter als Driven-Port
> mit Redis-Backend (`R-22`/RAK-90, gemeinsam mit R-17). Plus
> Trivy-Ignore-Re-Review (`R-13`-Wartung ohne RAK, Tranche-Inhalt
> ohne User-Surface). Architekturentscheidungen: hexagonale
> Erweiterung um zwei neue Driven-Ports (`OriginRateLimiter`,
> `MediaServerProvisioner`); bestehende `AuthSecretBackend`-,
> `IssuanceRateLimiter`-Ports bekommen neue Adapter
> (Vault-AppRole/KMS bzw. Redis-Network-Backend). Backwards-
> Compat: heutige ENV-Werte bleiben Default-Pfade; neue Adapter
> sind opt-in. **Out of Scope** und damit nicht durch diesen
> Patch erfüllt: Memcached-Backend für R-17/R-22 (wird gemeinsam
> als Folge-Item geliefert, sobald Operator-Bedarf entsteht),
> persistente Dead-Letter-Queue für Outbound-Webhooks aus
> `0.12.5`, Compliance-Audit-Zertifikate (PCI/SOC2 — `0.12.6`
> liefert nur die Konfigurations-Pfade), externe Provisionierung
> für SRS (`MediaServerProvisioner` ist auf MediaMTX-API
> spezialisiert), Production-Backends aus `0.13.0` (Postgres,
> ClickHouse, Kubernetes). Patch-Log siehe

>
>
> **Patch `1.1.16` (Auth-/Ingest-Adapter für `0.12.5`)**: Liefert
> die Adapter-/Wire-Pfade, die in `0.12.0` als Folge-Scope angelegt
> und in `0.12.1` als „Code-Pfad in 0.12.5" markiert wurden, und
> führt die neue RAK-Gruppe `RAK-77`..`RAK-82` in §13.15 ein. Inhalt:
> Shared-State-Issuance-Limiter (`R-17`) als Driven-Port
> (`IssuanceLimiterPort`) mit SQLite-Backend-Adapter als erstem
> opt-in Shared-State-Pfad (globaler Default bleibt In-Process-
> `memory`; Topologie-Constraint: Single-Host-Shared-Volume);
> Multi-Key-Signing-Resolver (`R-18`) mit ENV-Schema
> `MTRACE_AUTH_SIGNING_KEYS` und aktiver `kid`-Auswahl, restart-
> stabil; Secret-Backend-Driven-Port (`R-20`,
> `SecretBackendPort`) mit ENV-Default-Adapter plus
> **Vault-Adapter-Skelett** (KMS bleibt additive Folge-Option,
> nicht Teil dieses Patches); Browser-Ingest-Policy (`R-21`) als
> Project-Policy-gesteuerte, kontrollierte Aufhebung des
> RAK-74-Scope-Cuts auf `/api/ingest/*`. Optionale Adapter
> `RAK-81`/`RAK-82` (Auth-Bridge MediaMTX/SRS, Outbound-Webhook)
> bei verfügbarer Bandbreite. Architekturentscheidung:
> hexagonale Erweiterung um zwei neue Driven-Ports; bestehender
> `SigningKeyResolver` wird zum `MultiKeySigningResolver`.
> Backwards-Compat: heutige ENV-Werte (Single-Key,
> In-Process-Limiter, ENV-Secret-Lookup) bleiben Default-Pfad.
> **Out of Scope** und damit nicht durch diesen Patch erfüllt:
> KMS-Backend (nur Vault-Skelett wird geliefert), echte
> Multi-Host-Topologie für den Issuance-Limiter (Network-Backend
> wie Redis/Memcached bleibt Folge-Item), Externe Provisionierung
> (`R-15`), OAuth/OIDC/SSO, User-/Org-/Admin-Verwaltung
> (RAK-71-Out-of-Scope bleibt normativ), Origin-/IP-nahes
> Rate-Limiting (`R-22` getrennter Folge-Plan),
> Production-Backends aus `0.13.0` (Postgres, ClickHouse,
> Kubernetes). Patch-Log siehe

>
>
> **Patch `1.1.15` (Auth / Token Lifecycle für `0.12.0`)**: Hebt
> `F-111`..`F-113` (serverseitig signierte Session Tokens, rotierbare
> Project Tokens, tenant-spezifische Ingest Policies; bisher Kann-
> Erweiterungen) für den begrenzten `0.12.0`-Auth-/Security-Scope auf
> Release-Muss und führt die neue RAK-Gruppe `RAK-71`..`RAK-76` in
> §13.14 ein. Architekturentscheidung **Variante B** — Auth-Modul in
> `apps/api`, **kein** eigener Auth-Service in `0.12.0`. Persistenz:
> SQLite über bestehende API-Persistenz plus InMemory-Testadapter;
> Klartext-Tokens werden nicht persistiert. Bestehende
> `X-MTrace-Token`-Project-Token-Flows bleiben im
> `0.12.0`-Compatibility-Fenster gültig; SDK-/Demo-/Analyze-/
> Session-Link-/Ingest-Control-Pfade haben dokumentierte
> Kompatibilitätsnachweise. **Out of Scope** und damit nicht durch
> diesen Patch erfüllt: User-/Org-Verwaltung, OAuth/OIDC, SSO,
> Admin-UI, mandantenfähige SaaS-Control-Plane, KMS-/Vault-/Cloud-
> Secret-Manager, produktive MediaMTX-/SRS-Auth-Hook-Kopplung,
> globale Stream-Key-Rotation über mehrere Deployments,
> Production-Ops-Backends aus `0.13.0`, Cookies für Player-
> Telemetrie. Wire-Vertrag für `POST /api/auth/session-tokens` und
> die zusätzlichen Auth-Header für `POST /api/playback-events` lebt in
> [`spec/backend-api-contract.md`](backend-api-contract.md). Patch-Log siehe

>
>
> **Patch `1.1.14` (Ingest-Control-Scope für `0.11.0`)**: Hebt
> `F-46`..`F-51` (Ingest-Gateway / Stream Control, bisher Kann-
> Historie) für den begrenzten `0.11.0`-Lab-Control-Scope auf
> Release-Muss und führt die neue RAK-Gruppe `RAK-65`..`RAK-70` in
> §13.13 ein. `MVP-38` wird als lokaler SRT-/RTMP-Ingest-Control-
> Smoke für MediaMTX-nahe Lab-Artefakte präzisiert und für diesen
> Scope auf Release-Muss gezogen; die ältere Kann-Stufung bleibt
> auditierbar. **Out of Scope** und damit nicht durch diesen Patch
> erfüllt: mandantenfähige Control-Plane, KMS-/Vault-Secrets,
> globale Stream-Key-Rotation, produktive Media-Server-Auth-Hooks,
> automatische externe Provisionierung, Kubernetes-Operator,
> Auth-/Token-Lifecycle-Themen aus `0.12.0`, produktive ausgehende
> Webhook-Zustellung. Architekturentscheidung für `0.11.0`:
> Variante B — Ingest-Control als Modul in `apps/api`, **kein**
> eigener `apps/ingest-gateway`-Service. Eine spätere Ausgliederung
> bleibt möglich, ist aber Folge-Scope. Wire-Erweiterung in
> [`spec/backend-api-contract.md`](backend-api-contract.md) für `/api/ingest/*`. Patch-Log
> siehe

>
>
> **Patch `1.1.13` (CMAF-Analyse-Scope für `0.10.0`)**: Verankert den
> normativ begrenzten Analyzer-Scope für `NF-13` und führt die neue
> RAK-Gruppe `RAK-60`..`RAK-64` in §13.12 ein. `NF-13` ist nicht mehr
> als „CMAF-Vollanalyse" beschrieben, sondern als „CMAF-Analyse im
> Stream-Analyzer-Scope": manifestbasierte HLS-/DASH-Signale plus
> begrenzte binäre CMAF-Konformitätsprüfung ausgewählter Init-/
> Media-Segmente. Vollständige Segmentset-Abdeckung, Codec-Decoding,
> Low-Latency-CMAF und Player-Laufzeitpfade bleiben explizit
> Folge-Scope. Der Patch ändert keine Wire-Verträge — die
> Result-Schema-Erweiterungen (`details.cmaf`, `cmaf.binary.*`-
> Optionen) sind im Plan dokumentiert.
>
> **Patch `1.1.12` (Lastenheft-Konvergenz nach `0.9.5`)**: Keine
> neue Produktfunktion, keine User-Surface- oder Wire-Vertrags-
> änderung. Bereinigt die nach dem `0.9.5`-Audit sichtbaren
> Lieferstands-Unschärfen: `F-7` als Struktur-Anker präzisiert
> (Compose-Lab im Repo-Root bleibt der primäre Pfad), neue
> Pflichtdokumente-Kennung `F-131` mit harmonisierten Repo-Pfaden,
> `NF-13` (CMAF-Vollanalyse) bleibt offen und verweist auf den
> Folge-Plan `0.10.0`, `NF-18` (Kubernetes Production) als
> Folge-Scope abgegrenzt, `MVP-19`..`MVP-26`/`MVP-37`
> redaktionell so geschärft, dass die historische
> „Nicht im `0.1.0`-MVP"-Liste nicht als heutige offene Muss-
> Lücke missverstanden wird. Patch-Log siehe

>

---

## 1. Ziel des Projekts

m-trace ist ein Open-Source-Projekt zur lokalen und produktionsnahen Beobachtung, Analyse und Diagnose von Media-Streaming-Workflows.

Das Projekt soll Entwicklern, DevOps-Teams und Streaming-Betreibern ermöglichen, Live-Streams lokal und später auch in realen Umgebungen zu überwachen, Playback-Metriken zu erfassen, HLS-/DASH-Streams zu analysieren und Streaming-Probleme schneller einzugrenzen.

Der erste Fokus liegt auf einem reproduzierbaren lokalen Streaming-Labor mit Dashboard, Backend, Player-SDK, OpenTelemetry-Anbindung und Beispiel-Streaming-Server.

---

## 2. Ausgangssituation

Media-Streaming-Systeme bestehen häufig aus mehreren lose gekoppelten Komponenten:

- Encoder, z. B. OBS oder FFmpeg
- Ingest-Protokolle, z. B. RTMP oder SRT
- Media-Server, z. B. MediaMTX oder SRS
- Ausspielung über HLS, DASH, WebRTC oder ähnliche Protokolle
- Browser-Player
- Monitoring- und Logging-Systeme

In der Praxis ist die Fehlersuche oft schwierig, weil Informationen über Player-Verhalten, Stream-Zustand, Segment-Probleme, Latenz und Infrastrukturmetriken über mehrere Systeme verteilt sind.

m-trace soll diese Lücke schließen, indem es ein einfach startbares, erweiterbares und beobachtbares Streaming-Lab bereitstellt.

---

## 3. Projektvision

m-trace soll langfristig ein offenes Werkzeug für Streaming Observability und Stream-Diagnose werden.

Die langfristige Vision umfasst:

- lokale Streaming-Testumgebung per Docker Compose
- Browser-Player-SDK für Playback-Metriken
- API zur Annahme und Verarbeitung von Playback- und Stream-Events
- Dashboard für Live-Metriken und Sessions
- HLS-/DASH-/CMAF-Analyse
- OpenTelemetry-Export
- Prometheus- und Grafana-Integration
- SRT-, RTMP-, HLS-, DASH- und WebRTC-Beispiele
- erweiterbare Adapter für verschiedene Media-Server

---

## 4. Differenzierung und Marktpositionierung

Der Markt für Media-Streaming-Observability ist bereits gut besetzt. Kommerzielle Anbieter wie Mux Data, Bitmovin Analytics, NPAW/YOUBORA und Conviva decken viele klassische QoE- und Analytics-Anwendungsfälle ab.

m-trace soll sich deshalb nicht als allgemeines Video-Analytics-Produkt positionieren, sondern als offener, selbsthostbarer und OpenTelemetry-nativer Diagnose-Stack für Streaming-Infrastruktur.

### 4.1 Zentrale Differenzierung

Die zentrale Lücke liegt in der gemeinsamen Betrachtung von:

- Ingest
- Media Server / Origin
- Manifesten und Segmenten
- Player-Sessions
- Observability-Pipelines

Das Alleinstellungsmerkmal soll sein:

```text
OpenTelemetry-native streaming observability from ingest to player.
```

### 4.2 OpenTelemetry-native Ansatz

m-trace soll Player-Sessions, Stream-Ereignisse und Infrastrukturzustände so modellieren, dass sie in bestehende OpenTelemetry-Pipelines passen.

Ziel ist nicht ein weiteres isoliertes Monitoring-Silo, sondern Integration mit bestehenden Systemen wie:

- OpenTelemetry Collector
- Tempo
- Loki
- Mimir
- Prometheus
- Grafana
- ClickHouse oder VictoriaMetrics für hochvolumige Events

Ein wichtiges Zielbild ist die Modellierung einer Player-Session als Trace.

Beispielhafte Trace-Struktur:

```text
Player Session Trace
├── manifest_request
├── segment_request
├── segment_request
├── startup_time
├── bitrate_switch
├── rebuffer_event
└── playback_error
```

Damit wird eine spätere End-to-End-Korrelation zwischen Encoder, Ingest, Origin und Player möglich.

### 4.3 SRT als späterer starker Hebel

SRT ist für Contribution-Workflows, Broadcaster und Remote-Produktion besonders interessant.

m-trace soll später SRT-spezifische Metriken sichtbar machen, insbesondere:

- RTT
- Packet Loss
- Retransmissions
- verfügbare Bandbreite
- Send- und Receive-Buffer
- Verbindungsstabilität
- Link Health
- Failover-Zustände

Dieser Bereich ist für spätere Versionen ein hohes Differenzierungspotenzial, aber nicht Bestandteil des ersten MVP.

### 4.4 Manifest Analyzer als eigenständiger Wert

Der HLS-/DASH-Manifest-Analyzer soll als eigenständige Library und CLI betrachtet werden, nicht nur als internes Dashboard-Feature.

Besonders relevant sind:

- HLS-Compliance
- DASH-Compliance
- Segment-Drift
- Target-Duration-Verletzungen
- `EXT-X-DISCONTINUITY`-Plausibilität
- Varianten-/Rendition-Konsistenz
- Codec-/Container-Hinweise

Eine offene, gut diagnostizierende Alternative zu schwer zugänglichen oder proprietären Validatoren kann eigenständig wertvoll sein.

### 4.5 Bewusste Abgrenzung

m-trace soll im ersten MVP nicht versuchen, kommerzielle QoE-Plattformen vollständig zu ersetzen.

Nicht der Fokus im MVP:

- vollständige Business-Analytics
- Zuschauer-Tracking
- A/B-Testing
- DRM-Analytics
- Ad-Analytics
- WebRTC-Monitoring
- Multi-CDN-Kostenoptimierung
- umfangreiche Endgeräte-Kompatibilitätsmatrix

Der erste Fokus liegt auf technischer Diagnose und OpenTelemetry-Integration.

---


## 5. Zielgruppen

### 5.1 Primäre Zielgruppen

- Softwareentwickler im Media-Streaming-Umfeld
- DevOps- und Plattformteams
- Betreiber kleiner und mittlerer Streaming-Plattformen
- Entwickler von Playern, Streaming-Backends oder Video-Workflows
- Open-Source-Contributors mit Interesse an Media-Infrastruktur

> **Patch `1.1.20` (Tranche-Notiz):** Für die nächsten
> Minor-Releases ist das Primärziel auf Selbsthoster, kleine bis
> mittlere Streaming-Teams, Broadcaster-Labs und technische Media-/
> DevOps-Teams geschärft. Große Plattformbetreiber mit hunderten
> parallelen Streams bleiben späterer Scope und brauchen konkrete
> Trigger, Owner und Folgeplan.

### 5.2 Sekundäre Zielgruppen

- Vereine, Bildungseinrichtungen und Event-Teams mit Self-Hosted-Streaming
- Unternehmen mit internen Live-Streaming-Workflows
- Entwickler, die Streaming-Protokolle lernen oder testen möchten

---

## 6. Geltungsbereich

Dieses Lastenheft beschreibt die Anforderungen an die erste öffentliche Projektphase von m-trace.

Der Fokus liegt auf:

- Mono-Repo-Struktur
- hexagonaler Architektur
- lokaler Entwicklungsumgebung
- lauffähigem Docker-Compose-Setup
- Backend-API in Go (siehe [`docs/adr/0001-backend-stack.md`](../docs/adr/0001-backend-stack.md))
- SvelteKit Dashboard
- TypeScript Player-SDK
- einfachem Stream Analyzer
- OpenTelemetry-Grundlagen
- Dokumentation und Open-Source-Projektstruktur

Nicht Bestandteil der ersten Projektphase sind:

- vollständige Produktionsplattform
- Mandantenfähigkeit
- Abrechnungssystem
- DRM
- Benutzerverwaltung mit SSO
- Kubernetes-Produktionsbetrieb
- hochverfügbare Streaming-Infrastruktur
- kommerzielles CDN-Management

---

## 7. Funktionale Anforderungen

### 7.1 Mono-Repo

Das Projekt muss als Mono-Repo organisiert werden.

#### Anforderungen

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-1 | Muss | Das Repository muss alle Hauptbestandteile des Projekts enthalten. |
| F-2 | Muss | Anwendungen müssen unter `apps/` liegen. |
| F-3 | Muss | Wiederverwendbare Libraries müssen unter `packages/` liegen. |
| F-4 | Muss | Hilfsdienste müssen unter `services/` liegen. |
| F-5 | Muss | Beispiele müssen unter `examples/` liegen. |
| F-6 | Muss | Observability-Konfigurationen müssen unter `observability/` liegen. |
| F-7 | Muss | Deployment-Artefakte müssen unter `deploy/` liegen. **Patch `1.1.12` (Status-Präzisierung):** `deploy/` ist Struktur-Anker für reproduzierbare Deployment-Artefakte. Der primäre unterstützte lokale Deployment-Pfad bleibt das Compose-Lab über die Root-Datei `docker-compose.yml`. `deploy/k8s/` ist Folge-Scope (`MVP-42`, `Kann`) und ausdrücklich kein Production-Ready-K8s-Stand; siehe `deploy/README.md`. |
| F-8 | Muss | Dokumentation muss unter `docs/` liegen. |
| F-9 | Muss | Skripte müssen unter `scripts/` liegen. |

#### Zielstruktur

```text
m-trace/
├── apps/
│   ├── api/                    # Backend/API
│   ├── dashboard/              # SvelteKit Web UI
│   ├── ingest-gateway/         # optionaler Ingest-/Routing-Service
│   ├── analyzer-api/           # deferred Analyse-Service
│   ├── control-plane/          # spätere Verwaltungs-/Admin-App
│   └── demo-player/            # isolierte Player-Demo-App
├── packages/
│   ├── player-sdk/
│   ├── stream-analyzer/
│   ├── shared-types/
│   ├── ui/
│   └── config/
├── services/
│   ├── stream-generator/
│   ├── otel-collector/
│   └── media-server/
├── examples/
│   ├── srs/
│   ├── mediamtx/
│   ├── hls/
│   ├── dash/
│   ├── srt/
│   └── webrtc/
├── observability/
│   ├── prometheus/
│   ├── grafana/
│   └── otel/
├── deploy/
│   ├── compose/
│   ├── docker/
│   └── k8s/
├── docs/
├── scripts/
├── docker-compose.yml
├── Makefile
├── README.md
└── CHANGELOG.md
```

---

### 7.2 Hexagonale Architektur

Die fachlich relevanten Anwendungen und Libraries müssen nach hexagonaler Architektur strukturiert werden.

#### Anforderungen

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-10 | Muss | Fachlogik muss im Ordner `hexagon/` liegen. |
| F-11 | Muss | Technische Ein- und Ausgänge müssen im Ordner `adapters/` liegen. |
| F-12 | Muss | Abhängigkeiten müssen von außen nach innen zeigen. |
| F-13 | Muss | Die Domain darf keine Framework-, HTTP-, Datenbank- oder Docker-Abhängigkeiten enthalten. |
| F-14 | Muss | Ports müssen als Schnittstellen definiert werden. |
| F-15 | Muss | Adapter müssen Ports implementieren oder Use Cases aufrufen. |
| F-16 | Muss | DTOs dürfen nicht Teil der Domain sein. |

#### Standardstruktur

```text
src/
├── hexagon/
│   ├── domain/
│   ├── port/
│   │   ├── in/
│   │   └── out/
│   └── application/
└── adapters/
    ├── in/
    └── out/
```

#### Abhängigkeitsregel

```text
adapters → hexagon
```

Nicht erlaubt:

```text
hexagon → adapters
```

---

### 7.3 API-Anwendung

Die API-Anwendung muss unter `apps/api` liegen. Backend-Technologie ist Go gemäß [`docs/adr/0001-backend-stack.md`](../docs/adr/0001-backend-stack.md); Spec in §10.1.

#### Hauptaufgaben

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-17 | Muss | Annahme von Playback-Events |
| F-18 | Muss | Verwaltung von Stream-Sessions |
| F-19 | Muss | Bereitstellung von Metriken |
| F-20 | Muss | Weitergabe von Telemetrie an OpenTelemetry |
| F-21 | Muss | Bereitstellung von Daten für das Dashboard |
| F-22 | Muss | Architektur-Vorbereitung in `apps/api` für Stream Analyzer (Port-Hook); volle Integration ab Phase `0.3.0`. |

#### Mindest-Endpunkte für den MVP

| Methode | Pfad | Zweck |
|---|---|---|
| `POST` | `/api/playback-events` | Annahme eines Playback-Events |
| `GET` | `/api/stream-sessions` | Liste bekannter Stream-Sessions |
| `GET` | `/api/stream-sessions/{id}` | Details einer Stream-Session |
| `GET` | `/api/health` | Health Check |
| `GET` | `/api/metrics` | technische Metriken, sofern aktiviert |

#### Beispielhafte API-Domänenobjekte

- `Project`
- `ProjectId`
- `ProjectToken`
- `AllowedOrigin`
- `StreamSession`
- `StreamId`
- `PlaybackEvent`
- `PlaybackMetric`
- `PlaybackError`
- `StreamHealth`
- `LatencyMeasurement`

---

### 7.4 Dashboard

Das Dashboard muss unter `apps/dashboard` liegen und mit SvelteKit umgesetzt werden.

#### Hauptaufgaben

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-23 | Muss | Anzeige laufender Stream-Sessions |
| F-24 | Muss | Anzeige aktueller Playback-Metriken |
| F-25 | Muss | Anzeige von Fehlern und Warnungen |
| F-26 | Muss | Anzeige einfacher Stream-Health-Zustände |
| F-27 | Muss | Anzeige von Backend- und Telemetrie-Status |
| F-28 | Muss | Integration eines Test-Players |

#### Mindestansichten für den MVP

| Ansicht | Zweck |
|---|---|
| Startseite | Überblick über lokale Demo |
| Stream Sessions | Liste aktiver und vergangener Sessions |
| Session Details | Detailansicht zu Metriken und Events |
| Test Player | HLS-Testplayer mit eingebundenem Player-SDK |
| System Status | Status von API, Media Server und Observability |

#### Frontend-Architektur

Das Dashboard muss nicht zwingend vollständig hexagonal aufgebaut werden. Es soll eine pragmatische Feature-Struktur verwenden.

```text
apps/dashboard/src/
├── lib/
│   ├── api/
│   ├── components/
│   ├── features/
│   ├── stores/
│   └── types/
└── routes/
```

---

### 7.5 Weitere Anwendungen im Mono-Repo

Neben `apps/api` und `apps/dashboard` soll das Mono-Repo so vorbereitet werden, dass weitere Anwendungen sauber ergänzt werden können.

Die Detailarchitektur der Pflicht-Apps wird nur einmal verbindlich beschrieben. Spätere App-Beschreibungen dürfen diese Struktur nicht duplizieren, sondern nur Verantwortlichkeiten und Abgrenzungen ergänzen. Nicht jede App muss im ersten MVP vollständig implementiert sein, aber ihre fachliche Rolle, Abgrenzung und spätere Architektur sollen im Lastenheft definiert sein.

#### Grundregel

Jede Anwendung unter `apps/` ist eine eigenständig startbare Anwendung oder ein klar abgegrenzter Dienst mit eigenem Build, eigener Konfiguration und eigener Verantwortlichkeit.

Wiederverwendbare Fachlogik gehört nicht direkt in eine App, sondern in `packages/`.

---

#### 7.5.1 `apps/api`

`apps/api` ist die zentrale Backend-API für Playback-Events, Stream-Sessions, Dashboard-Daten und Telemetrie.

Status im MVP: **Muss**

Hauptaufgaben:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-29 | Muss | Playback-Events annehmen |
| F-30 | Muss | Stream-Sessions verwalten |
| F-31 | Muss | Metriken vorbereiten oder exportieren |
| F-32 | Muss | Daten für Dashboard bereitstellen |
| F-33 | Muss | Stream Analyzer anbinden |
| F-34 | Muss | Health Checks bereitstellen |

Architektur:

```text
apps/api/
├── src/
│   ├── hexagon/
│   │   ├── domain/
│   │   ├── port/
│   │   │   ├── in/
│   │   │   └── out/
│   │   └── application/
│   └── adapters/
│       ├── in/
│       │   ├── http/
│       │   └── websocket/
│       └── out/
│           ├── persistence/
│           ├── telemetry/
│           └── analyzer/
└── Dockerfile
```

---

#### 7.5.2 `apps/dashboard`

`apps/dashboard` ist die Weboberfläche für lokale Demo, Stream-Sessions, Playback-Events, Test-Player und Systemstatus.

Status im MVP: **Muss**

Hauptaufgaben:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-35 | Muss | Live-Übersicht anzeigen |
| F-36 | Muss | Test-Player bereitstellen |
| F-37 | Muss | Playback-Events anzeigen |
| F-38 | Muss | Stream-Sessions anzeigen |
| F-39 | Muss | API-Status anzeigen |
| F-40 | Muss | Links zu Grafana, Prometheus und Media-Server-Konsole anzeigen |

Architektur:

```text
apps/dashboard/src/
├── lib/
│   ├── api/
│   ├── components/
│   ├── features/
│   ├── stores/
│   └── types/
└── routes/
```

Hinweis: Das Dashboard muss nicht strikt hexagonal aufgebaut werden. Wenn später echte Fachlogik entsteht, kann innerhalb einzelner Features eine kleine Hexagon-Struktur eingeführt werden.

---

#### 7.5.3 `apps/demo-player`

`apps/demo-player` ist keine MVP-App.

Im MVP wird die Player-Demo als Route im Dashboard umgesetzt:

```text
apps/dashboard/src/routes/demo/
```

Eine separate App `apps/demo-player` wird erst sinnvoll, wenn der Player-SDK als eigenständiges Produktpaket demonstriert werden soll.

Status im MVP: **Nicht Bestandteil**

Spätere Hauptaufgaben:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-41 | Kann | HLS-Teststream abspielen |
| F-42 | Kann | Player-SDK isoliert integrieren |
| F-43 | Kann | erzeugte Events sichtbar machen |
| F-44 | Kann | SDK-Konfiguration testen |
| F-45 | Kann | als minimale Referenzintegration für externe Nutzer dienen |

Warum nicht im MVP:

Das Dashboard kann die Demo-Funktion zunächst ausreichend abdecken. Eine eigene App würde Build-, Deployment- und Dokumentationsaufwand erhöhen, ohne den ersten Nutzwert wesentlich zu steigern.


---

#### 7.5.4 `apps/ingest-gateway`

`apps/ingest-gateway` ist ein späterer Dienst zur Verwaltung von Ingest-Flows, Stream-Keys und Routing-Regeln.

Status im MVP: **Kann**

Hauptaufgaben:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-46 | Muss (`0.11.0`-Scope, Patch `1.1.14`) | Stream-Keys verwalten — lokale/lab-nahe CSPRNG-Erzeugung mit URL-sicherem Output, Persistenz speichert nur `key_hash` und redigierten `fingerprint`, Klartext-Keys nur bei Anlage/Rotation in Antworten. Historische Kann-Stufung gilt für `apps/ingest-gateway`-Ausgliederung; siehe RAK-66 in §13.13. |
| F-47 | Muss (`0.11.0`-Scope, Patch `1.1.14`) | Ingest-Endpunkte beschreiben — Protokoll-Allowlist `srt`/`rtmp`, Host/Port/Path plus Lab-/Egress-Hinweise. Externe Infrastruktur-Provisionierung bleibt Folge-Scope; siehe RAK-67 in §13.13. |
| F-48 | Muss (`0.11.0`-Scope, Patch `1.1.14`) | Routing-Regeln für Streams definieren — 1:1-Mapping `IngestStream` ↔ `MediaServerTarget`, deterministisch validierbar. Priorisierung, Fan-out und Failover bleiben Folge-Scope; siehe RAK-67. |
| F-49 | Muss (`0.11.0`-Scope, Patch `1.1.14`) | Stream-Lifecycle-Events — `stream_started`/`stream_ended` als stabiles Eventmodell, lokal reproduzierbar empfangbar/auslösbar; **keine** Klartext-Keys. Produktive ausgehende Webhook-Zustellung an externe Systeme bleibt Folge-Scope; siehe RAK-69 in §13.13. |
| F-50 | Muss (`0.11.0`-Scope, Patch `1.1.14`) | SRT-/RTMP-Konfigurationen vorbereiten — beschreibbare, reviewbare Lab-Artefakte, kein Direkt-Manipulation laufender externer Server; siehe RAK-68 in §13.13. |
| F-51 | Muss (`0.11.0`-Scope, Patch `1.1.14`) | Media-Server-Konfigurationen generieren oder validieren — normativer Zielserver MediaMTX im Lab-Scope; SRS bleibt Kompatibilitäts-/Dokuhintergrund. Bestehende `examples/`-Stacks bleiben grün; siehe RAK-68. |

Mögliche Endpunkte:

| Methode | Pfad | Zweck |
|---|---|---|
| `POST` | `/api/ingest/streams` | neuen Ingest-Stream registrieren |
| `GET` | `/api/ingest/streams` | Ingest-Streams listen |
| `POST` | `/api/ingest/streams/{id}/rotate-key` | Stream-Key erneuern |
| `POST` | `/api/ingest/hooks/stream-started` | Start-Webhook empfangen |
| `POST` | `/api/ingest/hooks/stream-ended` | Ende-Webhook empfangen |

Architektur:

```text
apps/ingest-gateway/
├── src/
│   ├── hexagon/
│   │   ├── domain/
│   │   │   ├── model/
│   │   │   └── service/
│   │   ├── port/
│   │   │   ├── in/
│   │   │   └── out/
│   │   └── application/
│   └── adapters/
│       ├── in/
│       │   ├── http/
│       │   └── webhook/
│       └── out/
│           ├── persistence/
│           ├── media_server/
│           └── telemetry/
└── Dockerfile
```

Mögliche Domain-Objekte:

- `IngestStream`
- `StreamKey`
- `IngestEndpoint`
- `RoutingRule`
- `MediaServerTarget`
- `IngestProtocol`
- `StreamLifecycleEvent`

---

#### 7.5.5 `apps/analyzer-api`

`apps/analyzer-api` ist ein historisch optionaler, seit Patch `1.1.20`
deferred geführter HTTP-Service für Stream-Analysen. Er würde
`packages/stream-analyzer` kapseln und Analysefunktionen über HTTP
bereitstellen, wird aber nur bei konkretem externem Konsumenten und
Folgeplan reaktiviert.

Status im MVP: **Deferred / Folge-Scope nur bei Trigger**

> **Patch `1.1.20` (Tranche-Notiz):** Eine eigenständige,
> nach außen exponierte `apps/analyzer-api` bleibt deferred. Der
> bestehende interne `apps/analyzer-service` ist der aktuelle
> HTTP-Wrapper für `apps/api`; technische Nutzer können
> `@pt9912/stream-analyzer` als Library oder CLI direkt verwenden.
> Eine externe API braucht vor `proceed` oder `POC` einen konkreten
> externen Konsumenten, Auth-/Rate-Limit-/SSRF-Grenzen,
> Ergebnisabruf-/Retention-Entscheidung, Contract-Fixtures, Owner und
> Folgeplan.

Warum deferred:

Im ersten MVP kann `apps/api` den Analyzer direkt als Library bzw. über
den internen `apps/analyzer-service` nutzen. Ein separater, nach außen
exponierter Analyse-Service lohnt sich erst, wenn Analysen schwerer
werden, unabhängig skaliert werden sollen, unsichere externe URLs
isoliert verarbeitet werden müssen oder ein externer Konsument den
API-/Job-Scope konkret macht.

Hauptaufgaben bei späterer Reaktivierung:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-52 | Deferred | HLS-URL entgegennehmen |
| F-53 | Deferred | Manifest analysieren |
| F-54 | Deferred | Analyseergebnis als JSON liefern |
| F-55 | Deferred | Fehler und Warnungen normalisieren |
| F-56 | Deferred | spätere DASH-/CMAF-Analyse anbieten |
| F-57 | Deferred | Sicherheitsgrenzen für externe URL-Abrufe schaffen |

Mögliche spätere Endpunkte:

| Methode | Pfad | Zweck |
|---|---|---|
| `POST` | `/api/analyze/hls` | HLS-Stream analysieren |
| `POST` | `/api/analyze/dash` | DASH-Stream analysieren, später |
| `GET` | `/api/analyze/jobs/{id}` | Analysejob abfragen, später |

Architektur:

```text
apps/analyzer-api/src/
├── hexagon/
│   ├── domain/
│   ├── port/
│   │   ├── in/
│   │   └── out/
│   └── application/
└── adapters/
    ├── in/
    │   └── http/
    └── out/
        ├── analyzer/
        ├── http_fetcher/
        └── telemetry/
```

Mögliche Domain-Objekte:

- `AnalysisJob`
- `StreamAnalysisRequest`
- `StreamAnalysisResult`
- `ManifestWarning`
- `ManifestError`
- `SegmentTimingIssue`

---

#### 7.5.6 `apps/control-plane`

`apps/control-plane` ist eine spätere Verwaltungsanwendung für produktionsnahe m-trace-Installationen.

Kennung: `F-132`

Status im MVP: **Nicht Bestandteil**, nur vorbereitet

> **Patch `1.1.20` (Tranche-Notiz):** `F-132` bleibt
> deferred. `0.15.0` baut keine `apps/control-plane` und gibt keinen
> POC frei. Reaktivierung braucht einen konkreten Operator-/
> Stakeholder-Bedarf mit mindestens zwei administrierten m-trace-
> Instanzen oder Projects, Betreiberprofil, User-/Org-/Auth-
> Anforderungen, Audit-/Compliance-Bedarf, Owner, SLO und eigenem
> Folgeplan. RAK-71-Out-of-Scope bleibt normativ: keine User-/Org-
> Verwaltung, kein OAuth/OIDC/SSO, keine Admin-UI und kein SaaS-
> Tenant-Modell in `0.15.0`.

Hauptaufgaben in späteren Versionen:

- Konfiguration mehrerer m-trace-Instanzen
- Verwaltung von Media-Servern
- Verwaltung von Stream-Profilen
- Verwaltung von Teams und Projekten
- Audit-Log
- API-Keys
- Integrationen
- spätere Benutzerverwaltung

Wichtige Abgrenzung:

`apps/control-plane` darf im MVP nicht gebaut werden. Sonst entsteht zu früh eine Plattform, bevor das eigentliche Streaming-Diagnoseproblem gelöst ist.

Mögliche spätere Architektur:

```text
apps/control-plane/
├── backend/
└── frontend/
```

Oder bei klarer Trennung:

```text
apps/control-plane-api/
apps/control-plane-ui/
```

Die finale Aufteilung ist erst sinnvoll, wenn echte Anforderungen für Mehrbenutzerbetrieb und Administration vorliegen.

---

#### 7.5.7 App-Übersicht nach Priorität

| App | Zweck | MVP-Status | Technologie |
|---|---|---|---|
| `apps/api` | zentrale Backend-API | Muss | Go (ADR-0001) |
| `apps/dashboard` | Web-Dashboard | Muss | SvelteKit |
| `apps/demo-player` | SDK-Referenz und Testplayer | Nicht MVP, zunächst `/demo`-Route | SvelteKit oder Vite |
| `apps/ingest-gateway` | Stream-Key, Ingest und Routing | Kann | Go (analog ADR-0001) |
| `apps/analyzer-api` | separater Analyse-Service | Deferred / Folge-Scope bei RAK-102-Trigger | Technologie offen (Go oder Node.js erst im Folgeplan) |
| `apps/control-plane` | spätere Verwaltungsplattform (`F-132`) | Später | offen |

---

#### 7.5.8 Empfehlung für die erste Umsetzung

Für den ersten lauffähigen Release sollen nur folgende Apps aktiv implementiert werden:

```text
apps/
├── api/
└── dashboard/
```

Der Demo-Player wird zunächst als Route im Dashboard umgesetzt:

```text
apps/dashboard/src/routes/demo/
```

Folgende Apps sollen zunächst höchstens als dokumentierte Platzhalter existieren:

```text
apps/
├── ingest-gateway/
├── analyzer-api/
└── control-plane/
```

Das verhindert Architektur-Overhead und hält den ersten Release realistisch.

---

### 7.6 Player-SDK

Das Player-SDK muss unter `packages/player-sdk` liegen und in TypeScript umgesetzt werden.

#### MVP-Abgrenzung

Im MVP unterstützt das Player-SDK nur `hls.js`.

Weitere Player-Adapter sind spätere Erweiterungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-58 | Kann | dash.js |
| F-59 | Kann | Shaka Player |
| F-60 | Kann | Video.js |
| F-61 | Kann | native Safari HLS |
| F-62 | Kann | WebRTC `getStats()`, separat in späterer Phase |

Ein Player-SDK von Grund auf ist ein eigenes Subprojekt und darf nicht unterschätzt werden. Unterschiedliche Player liefern unterschiedliche Events, Timing-Modelle und Metriken. Safari mit nativem HLS bietet besonders wenig Introspektion.

#### Browser-Support im MVP

Der MVP definiert bewusst eine enge Browser-Matrix, um den Testaufwand realistisch zu halten.

| Umgebung | Status im MVP |
|---|---|
| Chrome Desktop, aktuelle stabile Version | unterstützt |
| Firefox Desktop, aktuelle stabile Version | unterstützt |
| Safari Desktop, aktuelle stabile Version | eingeschränkt, nur Basis-Playback |
| Chromium-basierte Browser | best effort |
| iOS Safari | nicht verpflichtend im MVP |
| Android Chrome | nicht verpflichtend im MVP |
| Smart-TV Browser | explizit nicht im Scope |
| Embedded WebViews | explizit nicht im Scope |

Für den MVP gilt:

- hls.js ist der primäre Integrationspfad.
- Native Safari-HLS-Introspektion ist nicht Ziel von `0.1.0`.
- Mobile Browser werden später gezielt getestet.
- Smart-TV- und Set-Top-Box-Umgebungen sind vorerst ausgeschlossen.

#### Hauptaufgaben

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-63 | Muss | Anbindung an ein `HTMLVideoElement` |
| F-64 | Muss | Erfassung von Playback-Events |
| F-65 | Muss | Erfassung einfacher Metriken |
| F-66 | Muss | Versand der Events über OpenTelemetry Web SDK oder HTTP an die API |
| F-67 | Muss | Trennung von Browser-Adapter und fachlicher Tracking-Logik |

#### Zu erfassende Events im MVP

| Event | Beschreibung |
|---|---|
| `playback_started` | Wiedergabe wurde gestartet |
| `playback_paused` | Wiedergabe wurde pausiert |
| `playback_ended` | Wiedergabe wurde beendet |
| `startup_time_measured` | Startup-Zeit wurde gemessen |
| `rebuffer_started` | Buffering hat begonnen |
| `rebuffer_ended` | Buffering wurde beendet |
| `quality_changed` | Qualitäts-/Bitratenwechsel erkannt |
| `playback_error` | Player-Fehler erkannt |
| `metrics_sampled` | Regelmäßiger Metrik-Snapshot |

#### Zielstruktur im MVP

Das Player-SDK wird im MVP bewusst pragmatisch aufgebaut. Es nutzt keine vollständige Hexagon-Ceremony.

```text
packages/player-sdk/src/
├── core/
│   ├── session.ts
│   ├── event-buffer.ts
│   └── event-normalizer.ts
├── adapters/
│   └── hlsjs/
│       └── hlsjs-tracker.ts
├── transport/
│   ├── http-transport.ts
│   └── otel-transport.ts
├── types/
│   ├── events.ts
│   ├── config.ts
│   └── schema.ts
└── index.ts
```

Eine strengere Port-/Adapter-Struktur wird erst eingeführt, wenn mehr als ein Player-Adapter produktiv unterstützt wird.

---

### 7.7 Stream Analyzer

Der Stream Analyzer muss unter `packages/stream-analyzer` liegen und in TypeScript umgesetzt werden.

#### Hauptaufgaben

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-68 | Muss | Abruf von HLS-Manifesten |
| F-69 | Muss | Analyse einfacher Manifest-Eigenschaften |
| F-70 | Muss | Prüfung von Segment-Dauern |
| F-71 | Muss | Erkennung offensichtlicher Inkonsistenzen |
| F-72 | Muss | Bereitstellung einer API für Backend und CLI |
| F-73 | Muss | Vorbereitung für DASH- und CMAF-Analyse |

#### Mindestfunktionen für den MVP

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-74 | Muss | HLS Master Playlist erkennen |
| F-75 | Muss | HLS Media Playlist erkennen |
| F-76 | Muss | Varianten und Renditions extrahieren |
| F-77 | Muss | Segment-Anzahl bestimmen |
| F-78 | Muss | durchschnittliche Segment-Dauer berechnen |
| F-79 | Muss | Abweichungen bei Segment-Dauern erkennen |
| F-80 | Muss | einfache Live-Latenz-Schätzung |
| F-81 | Muss | Analyseergebnis als JSON liefern |

#### CLI-Ziel

```bash
pnpm m-trace check https://example.com/live/master.m3u8
```

---

### 7.8 Lokales Streaming-Lab

Das Projekt muss eine lokale Streaming-Testumgebung bereitstellen.

#### Anforderungen

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-82 | Muss | Start per Docker Compose |
| F-83 | Muss | Media Server für lokale Tests |
| F-84 | Muss | FFmpeg-basierter Teststream |
| F-85 | Muss | API erreichbar unter `localhost` |
| F-86 | Muss | Dashboard erreichbar unter `localhost` |
| F-87 | Muss | Prometheus und Grafana optional verfügbar |
| F-88 | Muss | OpenTelemetry Collector optional verfügbar |

#### Mindestdienste

Die Dienste sind in zwei Klassen gegliedert (harmonisiert mit F-87/F-88
und MVP-28/MVP-29 in Patch `1.0.2`):

**Pflicht (Muss, im Default-Compose-Profil):**

| Dienst | Zweck |
|---|---|
| `api` | Backend-API |
| `dashboard` | SvelteKit UI |
| `mediamtx` | lokaler Media Server |
| `stream-generator` | FFmpeg-Teststream |

**Soll (optional, im `observability`-Compose-Profil):**

| Dienst | Zweck | Bezug |
|---|---|---|
| `otel-collector` | OpenTelemetry Collector | F-88 (optional verfügbar), MVP-29 |
| `prometheus` | Metrikspeicherung | F-87 (optional verfügbar) |
| `grafana` | Visualisierung | F-87 (optional verfügbar), MVP-28 |

> **Hinweis (Patch `1.1.1`)**: Die obigen Tabellen beschreiben den
> **End-Zustand der `0.1.x`-Phase**. Mit der Sub-Release-Schneidung
> aus Patch `1.1.0` (§13.1–§13.3) werden die Pflicht-Mindestdienste
> stufenweise aktiviert: `0.1.0` startet `api`, `mediamtx`,
> `stream-generator`; `0.1.1` ergänzt `dashboard`; `0.1.2` aktiviert
> die Soll-Tabelle über das `observability`-Compose-Profil. Der
> jeweilige Sub-Release-Pflicht-Subset steht im zugehörigen
> Plan-Dokument .

#### Erwarteter Startbefehl

```bash
make dev
```

Oder direkt:

```bash
docker compose up --build
```

---

### 7.9 Observability

Das Projekt muss Observability von Beginn an berücksichtigen.

#### Anforderungen

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-89 | Muss | API muss strukturierte Logs erzeugen. |
| F-90 | Muss | API muss Health Checks bereitstellen. |
| F-91 | Muss | API soll OpenTelemetry unterstützen. |
| F-92 | Muss | Playback-Events sollen als Metriken oder Traces exportierbar sein. |
| F-93 | Muss | Prometheus soll technische Metriken erfassen können. |
| F-94 | Soll | Grafana kann mit einem einfachen Beispiel-Dashboard ausgeliefert werden (harmonisiert mit MVP-28). |

#### Mindestmetriken

| Metrik | Beschreibung |
|---|---|
| `mtrace_playback_events_total` | Anzahl empfangener Playback-Events |
| `mtrace_playback_errors_total` | Anzahl empfangener Playback-Fehler |
| `mtrace_active_sessions` | Anzahl aktiver Sessions |
| `mtrace_rebuffer_events_total` | Anzahl Buffering-Ereignisse |
| `mtrace_startup_time_ms` | gemessene Startup-Zeit |
| `mtrace_api_requests_total` | API Requests |
| `mtrace_dropped_events_total` | Anzahl verworfener Events |
| `mtrace_rate_limited_events_total` | Anzahl durch Rate Limits abgelehnter Events |
| `mtrace_invalid_events_total` | Anzahl wegen Schema-/Validierungsfehlern (`400`/`422`) abgelehnter Events; Auth-Fehler (`401`) zählen nicht (harmonisiert mit API-Kontrakt §7 in Patch `1.1.2`). |

---

### 7.10 Datenmodell, Cardinality und Storage

m-trace muss von Beginn an zwischen aggregierten Metriken, hochvolumigen Events und per-Session-Daten unterscheiden.

#### Problem

Prometheus ist nicht geeignet für hochkardinale Labels wie:

- `session_id`
- `viewer_id`
- `client_ip`
- `user_agent`
- `segment_url`
- `request_id`

Diese Labels können bei Player-Telemetrie sehr schnell zu unkontrollierbarer Cardinality führen.

#### Anforderungen

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-95 | Muss | Prometheus darf nur für aggregierte Metriken verwendet werden. |
| F-96 | Muss | `session_id` darf nicht als Prometheus-Label verwendet werden. |
| F-97 | Muss | Per-Session-Daten sollen als Traces oder Events modelliert werden. |
| F-98 | Muss | Für hochvolumige Eventdaten muss eine spätere Storage-Option vorgesehen werden. |
| F-99 | Muss | Das System muss Sampling vorbereiten. |
| F-100 | Muss | Das Telemetrie-Modell muss Datenschutz und Cardinality gemeinsam berücksichtigen. |

#### Empfohlene Zuordnung

| Datentyp | Geeigneter Speicher | Zweck |
|---|---|---|
| aggregierte technische Metriken | Prometheus / Mimir | Dashboards, Alerts |
| Player-Session-Verläufe | Tempo / Traces | Debugging einzelner Sessions |
| hochvolumige Events | ClickHouse / VictoriaMetrics / später | Analyse und Historie |
| Logs | Loki | technische Fehlersuche |
| Konfiguration | PostgreSQL / SQLite / später | persistente Projekt- und Streamdaten |

#### MVP-Entscheidung

Im ersten MVP sollen folgende Regeln gelten:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-101 | Muss | Prometheus nur für Aggregate |
| F-102 | Muss | Player-Sessions als OpenTelemetry-Traces vorbereiten |
| F-103 | Muss | In-Memory-Speicherung nur für lokale Demo |
| F-104 | Muss | keine produktive Langzeitspeicherung im MVP |
| F-105 | Muss | keine `session_id`-Labels in Prometheus |

---

### 7.11 Telemetry Ingest, Event-Schema und SDK-Budget

Die Telemetrie-Schnittstelle ist ein Kernbestandteil des Projekts und muss früh spezifiziert werden.

#### Authentifizierung von Player-Events

Das Browser-SDK darf nicht dauerhaft gegen einen vollständig offenen Ingest-Endpunkt senden.

MVP-Anforderungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-106 | Muss | Events enthalten eine `project_id`. |
| F-107 | Muss | Events werden mit einem öffentlichen Project Token oder einem kurzlebigen Ingest Token versehen. |
| F-108 | Muss | Das Backend validiert erlaubte Origins. |
| F-109 | Muss | Tokens dürfen keine Secrets mit hoher Kritikalität sein, da Browser-Code öffentlich ist. |
| F-110 | Muss | Rate Limits gelten pro Project, Origin und IP-Bereich. |

Erweiterungen, die mit Patch `1.1.15` für den `0.12.0`-Auth-/
Security-Scope auf Release-Muss gehoben sind (historische Kann-
Stufung bleibt bis Patch `1.1.14` auditierbar):

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-111 | Muss (`0.12.0`-Scope, Patch `1.1.15`) | serverseitig signierte Session Tokens — kurzlebig, an Project, Audience, Ablauf und optional Session/Origin gebunden; Signaturschlüssel werden nie im Browser-Bundle, in Fixtures, Logs, Traces oder Beispiel-Configs ausgeliefert. Klartext-Token erscheint ausschließlich in der Issuance-Antwort. Siehe RAK-72 in §13.14. |
| F-112 | Muss (`0.12.0`-Scope, Patch `1.1.15`) | rotierbare Project Tokens — Generationen mit Hash/Fingerprint, `not_before`, `grace_until?`, `expires_at?`, `revoked_at?`; Persistenz speichert nie Klartext, Grace-Phase ist persistiert und restart-stabil. Siehe RAK-73 in §13.14. |
| F-113 | Muss (`0.12.0`-Scope, Patch `1.1.15`) | tenant-spezifische Ingest Policies im Project-Scope — erlaubte Origins, Methoden, Header, Audiences, maximale Session-Token-TTL und Rate-Limit-Buckets pro Project (Origin-/IP-nahe Buckets sind optionaler Zusatz oder Folge-Scope). Bestehende CORS-Regeln aus `NF-30`..`NF-37` werden in den Policy-Vertrag überführt. Siehe RAK-74 in §13.14. |

#### Schema-Versionierung

Jedes Event muss eine Schema-Version enthalten.

Pflichtfelder im Wire-Format:

```json
{
  "schema_version": "1.0",
  "event_name": "rebuffer_started",
  "project_id": "demo",
  "session_id": "01J...",
  "client_timestamp": "2026-04-28T12:00:00.000Z",
  "sdk": {
    "name": "@pt9912/player-sdk",
    "version": "0.2.0"
  }
}
```

Anforderungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-114 | Muss | neue Felder müssen abwärtskompatibel sein |
| F-115 | Muss | unbekannte Felder dürfen nicht zum Fehler führen |
| F-116 | Muss | entfernte Felder müssen über mindestens eine Minor-Version toleriert werden |
| F-117 | Muss | Breaking Changes erfordern neue Major-Version der Event-Schemas |

#### Backpressure und Rate Limiting

Die Ingest-API muss Überlastung kontrolliert behandeln.

MVP-Anforderungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-118 | Muss | maximale Event-Batch-Größe definieren |
| F-119 | Muss | maximale Request-Rate pro Project definieren |
| F-120 | Muss | HTTP `429` bei Rate Limit |
| F-121 | Muss | HTTP `202` für angenommene Events |
| F-122 | Muss | Events dürfen bei lokaler Überlast verworfen werden, wenn dies als Dropped-Event-Metrik sichtbar wird |
| F-123 | Muss | SDK muss Sampling und Batch-Größe konfigurieren können |

#### Zeitstempel und Time Skew

Browser-Clocks sind unzuverlässig. Das Backend muss daher zwischen Client-Zeit und Server-Zeit unterscheiden.

Pflichtfelder:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-124 | Muss | `client_timestamp` |
| F-125 | Muss | `server_received_at` |
| F-126 | Muss | optional `client_time_origin` |
| F-127 | Muss | optional `sequence_number` |

Regeln:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-128 | Muss | Ordering innerhalb einer Session bevorzugt über `sequence_number` |
| F-129 | Muss | Latenzberechnungen niemals blind nur aus Client-Zeit ableiten |
| F-130 | Muss | Backend muss auffälligen Time Skew markieren können |

#### Performance-Budget für das Player-SDK

Das SDK darf Playback nicht stören.

MVP-Budget:

| Kennzahl | Ziel |
|---|---|
| Bundle-Größe | kleiner als 30 KB gzip ohne hls.js |
| Event-Verarbeitung | unter 5 ms pro Event im Normalfall |
| Hot Path | keine synchronen Netzwerkaufrufe |
| Transport | batchingfähig |
| Fehlerverhalten | niemals Playback abbrechen |
| Sampling | konfigurierbar |

#### OpenTelemetry Semantic Conventions

m-trace soll sich an bestehenden OpenTelemetry-Konventionen orientieren und eigene Media-Konventionen nur dort ergänzen, wo keine passende Konvention existiert.

Strategie:

- bestehende HTTP-, Client-, Browser- und Runtime-Konventionen nutzen
- eigene Attribute mit stabilem Prefix definieren, z. B. `mtrace.*`
- Media-spezifische Semantik dokumentieren
- spätere Kompatibilität mit entstehenden OTel-Media-Konventionen einplanen

---

### 7.12 Dokumentation

Das Projekt muss eine entwicklerfreundliche Dokumentation enthalten.

#### Pflichtdokumente

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-131 | Muss | Die folgenden Pflichtdokumente müssen vorhanden und auf die aktuellen Repository-Pfade harmonisiert sein. **Eingeführt in Patch `1.1.12`** (vorher trug der Block keine eigene Kennung); [`docs/user/stream-analyzer.md`](../docs/user/stream-analyzer.md) aus früheren Lastenheft-Ständen ist auf den realen Pfad [`docs/user/stream-analyzer.md`](../docs/user/stream-analyzer.md) korrigiert. |

| Datei | Zweck |
|---|---|
| [`README.md`](../README.md) | Einstieg und Schnellstart |
| [`CHANGELOG.md`](../CHANGELOG.md) | Änderungsverlauf |
| `CONTRIBUTING.md` | Beitragsregeln |
| `LICENSE` | Lizenz |
| `SECURITY.md` | Sicherheitsmeldungen |
| [`spec/architecture.md`](architecture.md) | Architekturüberblick |
| [[`docs/user/local-development.md`](../docs/user/local-development.md)](../docs/user/local-development.md) | lokale Entwicklung |
| [`spec/telemetry-model.md`](telemetry-model.md) | Telemetrie- und Eventmodell |
| [`spec/player-sdk.md`](player-sdk.md) | Player-SDK-Nutzung |
| [`docs/user/stream-analyzer.md`](../docs/user/stream-analyzer.md) | Stream Analyzer |
| [`docs/planning/in-progress/roadmap.md`](../docs/planning/in-progress/roadmap.md) | geplante Entwicklung |

---

## 8. Nichtfunktionale Anforderungen

### 8.1 Plattform

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-1 | Muss | Entwicklung muss unter Linux möglich sein. |
| NF-2 | Muss | Entwicklung muss mit VS Code kompatibel sein. |
| NF-3 | Muss | Lokaler Betrieb muss über Docker möglich sein. |
| NF-4 | Muss | Build-Prozesse müssen ohne proprietäre Dienste funktionieren. |

### 8.2 Wartbarkeit

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-5 | Muss | Fachlogik muss testbar sein, ohne externe Infrastruktur zu starten. |
| NF-6 | Muss | Domain-Klassen dürfen keine Framework-Abhängigkeiten enthalten. |
| NF-7 | Muss | Ports müssen klar benannt und dokumentiert sein. |
| NF-8 | Muss | Adapter müssen austauschbar sein. |
| NF-9 | Muss | Technische Implementierungen dürfen nicht in die Domain-Schicht lecken. |

### 8.3 Erweiterbarkeit

Das Projekt muss vorbereitet sein für spätere Erweiterungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-10 | Muss | MediaMTX-Adapter |
| NF-11 | Muss | SRT-Ingest-Metriken |
| NF-12 | Muss | DASH-Analyse |
| NF-13 | Muss | CMAF-Analyse im Stream-Analyzer-Scope. **Patch `1.1.20` (Tranche-Notiz):** Aus den offenen Folgeoptionen wird nur der HTTP-Range-/Byte-Range-Loader für manifest-referenzierte CMAF-Init-/Media-Segmente als kleiner Folge-Slice empfohlen; Low-Latency-CMAF, vollständige Segmentset-Abdeckung, Codec-Decoding und Player-SDK-CMAF-Laufzeitpfade bleiben deferred. **Patch `1.1.13` (Scope-Präzisierung):** `NF-13` wird nicht mehr als „CMAF-Vollanalyse" geführt; vollständig ist sie für den Analyzer-Scope aus `0.10.0` (manifestbasierte HLS-/DASH-Signale plus begrenzte binäre CMAF-Konformitätsprüfung ausgewählter Init-/Media-Segmente). Explizit nicht umfasst sind vollständige Segmentset-Abdeckung, Codec-Decoding, Low-Latency-CMAF (`#EXT-X-PART`, chunked CMAF) und Player-Laufzeitpfade — diese bleiben Folge-Scope. Pflicht-Brand-Allowlist: Init-`ftyp` `cmfc`/`cmf2`, Media-`styp` `cmfs`/`cmff`/`cmfc`/`cmf2`. Umsetzung mit RAK-60..RAK-64 (§13.12). **Patch `1.1.12` (historisch):** `F-73` aus §7.7 deckte nur die *vorbereitete Erweiterbarkeit*; vor `0.10.0` blieb die CMAF-Analyse offen. |
| NF-14 | Muss | WebRTC-Metriken |
| NF-15 | Muss | Datenbankpersistenz |
| NF-16 | Muss | Authentifizierung |
| NF-17 | Muss | Multi-Stream-Betrieb |
| NF-18 | Muss | Kubernetes Deployment. **Patch `1.1.12` (Scope-Präzisierung):** „Erweiterbarkeit für Kubernetes" — Production-K8s ist **nicht** Bestandteil der ersten Projektphase. Optionale K8s-Manifeste bleiben `MVP-42` (`Kann`/Folge-Plan); R-9 ([`docs/planning/in-progress/risks-backlog.md`](../docs/planning/in-progress/risks-backlog.md)) bleibt Trigger-Risiko für eine künftige K8s-Smoke-Stage. Strukturanker `deploy/k8s/` ist mit `0.9.6` angelegt, aber leer. **Patch `1.1.18` (`0.13.0`):** NF-18 wird mit `MVP-42` als optionaler Optionspfad harmonisiert. `0.13.0` darf Beispielmanifeste oder Entscheidungsnotizen liefern, verpflichtet aber nicht zu Production-Ready-Kubernetes, Cluster-Betrieb, Cloud-Provider-Integration oder K8s-Smoke-Stage als Standard-Gate. Wird eine K8s-Smoke-Stage aktiviert, muss R-9 vorher mit eigener Observability-Label-Allowlist oder dokumentierter Gegenmaßnahme entschieden sein. |
| NF-19 | Muss | CI-basierte Stream-Checks |

### 8.4 Performance

Für den MVP gelten einfache Anforderungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-20 | Muss | API muss lokale Demo-Last problemlos verarbeiten. |
| NF-21 | Muss | Player-SDK darf Playback nicht merklich beeinflussen. |
| NF-22 | Muss | Dashboard muss bei mehreren aktiven Sessions bedienbar bleiben. |
| NF-23 | Muss | Event-Erfassung muss asynchron oder leichtgewichtig erfolgen. |

### 8.5 Sicherheit

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-24 | Muss | Keine Secrets im Repository. |
| NF-25 | Muss | `.env.example` muss Beispielwerte enthalten. |
| NF-26 | Muss | Produktive Secrets müssen über Umgebungsvariablen gesetzt werden. |
| NF-27 | Muss | CORS muss im lokalen Setup kontrolliert konfiguriert sein. |
| NF-28 | Muss | Externe URLs für Stream-Analyse müssen später abgesichert werden, um SSRF-Risiken zu vermeiden. |
| NF-29 | Muss | Security-Meldungen müssen über `SECURITY.md` beschrieben werden. |

#### CORS- und CSP-Grundregeln für Player-Telemetrie

Für Browser-SDK-Telemetrie muss Cross-Origin-Kommunikation kontrolliert werden.

MVP-Anforderungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-30 | Muss | erlaubte Origins werden pro Project konfiguriert |
| NF-31 | Muss | SDK-Requests nutzen standardmäßig `credentials: "omit"` |
| NF-32 | Muss | keine Cookies für Player-Telemetrie im MVP |
| NF-33 | Muss | Preflight-fähige CORS-Konfiguration |
| NF-34 | Muss | `Access-Control-Allow-Origin` darf nicht pauschal `*` sein, sobald Project Tokens genutzt werden |
| NF-35 | Muss | erlaubte Methoden zunächst auf `POST` und `OPTIONS` begrenzen |
| NF-36 | Muss | erlaubte Header explizit definieren, z. B. `Content-Type`, `X-MTrace-Project`, `X-MTrace-Token` |
| NF-37 | Muss | CSP-Beispiele für `connect-src` müssen dokumentiert werden |

Beispiel-CSP für eine Demo-Integration:

```text
Content-Security-Policy: connect-src 'self' https://m-trace.example.com;
```


### 8.6 Datenschutz und GDPR

Player-Telemetrie kann personenbezogene oder personenbeziehbare Daten enthalten. Dazu gehören insbesondere IP-Adressen, User-Agents, Session-IDs und grobe Standortinformationen.

Anforderungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-38 | Muss | IP-Adressen dürfen im MVP nicht unnötig gespeichert werden. |
| NF-39 | Muss | User-Agent-Daten müssen reduzierbar oder anonymisierbar sein. |
| NF-40 | Muss | Session-IDs müssen pseudonym sein. |
| NF-41 | Muss | Ein konfigurierbarer Anonymisierungs-Layer im Collector soll vorbereitet werden. |
| NF-42 | Muss | Das Projekt muss dokumentieren, welche Telemetriedaten erhoben werden. |
| NF-43 | Muss | Datenschutzfreundliche Defaults haben Vorrang vor maximaler Analyse-Tiefe. |
| NF-44 | Muss | Für EU-Nutzung muss eine GDPR-freundliche Betriebsweise möglich sein. |

### 8.7 Qualität

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-45 | Muss | Automatisierte Tests für Domain- und Application-Schicht |
| NF-46 | Muss | Linting für TypeScript |
| NF-47 | Muss | Tests für zentrale Backend-Use-Cases |
| NF-48 | Muss | CI-Pipeline für Build und Test |
| NF-49 | Muss | klare Commit- und Release-Konventionen |
| NF-50 | Muss | CHANGELOG-Pflege ab dem ersten Release |

---

## 9. Technologie-Strategie und Architekturentscheidungen

Streaming-Observability-relevante Komponenten und Communities sind stark durch Go, Rust und TypeScript geprägt:

- Media-Server und Streaming-Infrastruktur häufig in Go
- OpenTelemetry Collector in Go
- Browser- und Player-Ökosystem stark in TypeScript
- performante Analyzer- und CLI-Werkzeuge häufig in Go oder Rust

### 9.1 Backend-Entscheidung

**Entschieden: Go.** Die Wahl ist in [`docs/adr/0001-backend-stack.md`](../docs/adr/0001-backend-stack.md) (Status: Accepted) festgehalten und beruht auf zwei Mini-Prototypen mit identischem Muss-Scope ([`spec/backend-api-contract.md`](backend-api-contract.md)); das Spike-Protokoll liegt in [[`docs/spike/backend-stack-results.md`](../docs/spike/backend-stack-results.md)](../docs/spike/backend-stack-results.md).

Historischer Tradeoff (Stand vor dem Spike):

| Option | Vorteil | Nachteil |
|---|---|---|
| **Go** ✅ | passt kulturell gut zu OTel, MediaMTX und Infrastruktur-Tools | — |
| JVM (Micronaut) | vertrauter JVM-Stack, gute DI, gute Testbarkeit | kleinerer Contributor-Pool im Streaming-OSS-Umfeld |

Konkrete Stack-Spezifikation in §10.1.

### 9.2 Hexagonale Architektur

Hexagonale Architektur soll nicht dogmatisch für alle Komponenten gelten.

Verbindliche Regel:

```text
Hexagonal nur dort, wo echte fachliche Anwendungslogik entsteht.
```

Empfohlene Anwendung:

| Komponente | Architektur |
|---|---|
| `apps/api` | hexagonal |
| `packages/stream-analyzer` | hexagonal oder klar geschichtete Library |
| `packages/player-sdk` | pragmatisch, keine vollständige Hexagon-Ceremony |
| `apps/dashboard` | Feature-Struktur |
| `apps/demo-player` | keine eigene App im MVP, höchstens Route im Dashboard |

Für das Player-SDK genügt eine leichte Adapter-Struktur:

```text
packages/player-sdk/src/
├── core/
├── adapters/
│   └── hlsjs/
├── transport/
└── types/
```

Ports und Use Cases sind dort erst nötig, wenn mehrere Player-Adapter tatsächlich implementiert werden.

### 9.3 Selbsthoster-first Konsequenz

Da der MVP auf Selbsthoster, kleine Plattformen, Broadcaster-Labs und technische Teams zielt, muss die Architektur zuerst einfach betreibbar sein.

Für den MVP bedeutet das:

- keine Mimir-Pflicht
- keine ClickHouse-Pflicht
- keine große Multi-Tenant-Architektur
- keine getrennte Demo-Player-App
- keine getrennte Analyzer-API
- bevorzugt lokale Speicherung mit SQLite oder In-Memory
- eingebaute Trace-/Session-Anzeige im Dashboard als Alternative zu Tempo
- Tempo, Mimir und ClickHouse nur als optionale spätere Integrationen

---

## 10. Technische Rahmenbedingungen

### 10.1 Backend

Backend-Technologie: **Go**, entschieden in [`docs/adr/0001-backend-stack.md`](../docs/adr/0001-backend-stack.md).

| Bereich | Festlegung |
|---|---|
| Sprache | Go 1.22 oder höher |
| HTTP | Standard-Library `net/http` |
| Metriken | `prometheus/client_golang` |
| Tracing | `go.opentelemetry.io/otel` |
| Logging | `log/slog`, JSON-Formatter |
| Build/Runtime | Distroless-static (`gcr.io/distroless/static-debian12:nonroot`) |
| Linting | `golangci-lint` mit Default-Lintern (`govet`, `errcheck`, `staticcheck`, `unused`, `ineffassign`) plus SOLID-nahem Zusatzprofil |
| Tests | `testing` + `httptest`, keine externen Frameworks |
| Workflow | Docker-only (`docker build --target {test,lint,build,runtime}`); lokales Go optional |
| Modulpfad | `github.com/pt9912/m-trace/apps/api` |

Mindestanforderungen an die Implementierung:

- HTTP API für Event-Ingest gemäß [`spec/backend-api-contract.md`](backend-api-contract.md)
- Health Check
- strukturierte Logs (`slog`)
- OpenTelemetry-kompatibles Eventmodell
- klare Trennung von Domain, Application und Adapters (Hexagon-Layout `hexagon/{domain,application,port/{driving,driven}}`, `adapters/{driving,driven}/...`)
- Containerisierung per Docker

Das SOLID-nahe Zusatzprofil ist keine offizielle `golangci-lint`-
Kategorie, sondern die verbindliche Projektauswahl für Designsignale:
geringe Komplexität und kleine Verantwortlichkeiten (SRP), schlanke
Interfaces (ISP), stabile Import-/Modulgrenzen (DIP) und reduzierte
globale Kopplung. Es umfasst:

| Linter | Pflicht |
|---|---|
| `containedctx` | Ja |
| `contextcheck` | Ja |
| `cyclop` | Ja |
| `depguard` | Ja |
| `dupl` | Ja |
| `fatcontext` | Ja |
| `forbidigo` | Ja |
| `funlen` | Ja |
| `gochecknoglobals` | Ja |
| `gochecknoinits` | Ja |
| `gocognit` | Ja |
| `gocyclo` | Ja |
| `gomodguard` | Ja |
| `iface` | Ja |
| `inamedparam` | Ja |
| `interfacebloat` | Ja |
| `ireturn` | Ja |
| `maintidx` | Ja |
| `nestif` | Ja |
| `noctx` | Ja |
| `reassign` | Ja |
| `revive` | Ja |
| `testpackage` | Ja |
| `unparam` | Ja |

Multi-Modul-Aufteilung über `go.work` ist nicht im MVP erforderlich; erst on demand bei wachsender Codebase (siehe [`docs/planning/in-progress/roadmap.md`](../docs/planning/in-progress/roadmap.md) Folge-ADR).


### 10.2 Frontend

- Sprache: TypeScript
- Framework: SvelteKit
- Package Manager: pnpm
- Styling: zunächst pragmatisch, später UI-Package möglich
- Kommunikation: REST, später WebSocket oder SSE
- Qualität: Typecheck, Svelte-Check, Tests und SOLID-nahe
  Boundary-/Komplexitätsregeln für UI-Module

### 10.3 Player-SDK

- Sprache: TypeScript
- Zielumgebung: Browser
- Build: pnpm
- Ausgabeformat: ESM
- Kernlogik frameworkfrei
- Adapter für Browser und HTTP
- Qualität: Typecheck, Public-API-Snapshot und SOLID-nahe
  Boundary-Regeln zwischen Kernlogik, Transport und Browser-Adaptern

### 10.4 Stream Analyzer

- Sprache: TypeScript
- Zielumgebung: Node.js
- HLS zuerst
- DASH später
- CLI später
- API-kompatible JSON-Ergebnisse
- Qualität: Typecheck, Public-API-Snapshot und SOLID-nahe
  Boundary-Regeln zwischen Public API, CLI und `internal/`

### 10.5 Infrastruktur

- Docker
- Docker Compose
- MediaMTX als erster Media Server
- FFmpeg als Teststream-Generator
- OpenTelemetry Collector
- Prometheus
- Grafana

---

## 11. Abgrenzung zu ähnlichen Projekten

m-trace soll kein Ersatz sein für:

- OBS
- FFmpeg
- SRS
- MediaMTX
- Wowza
- Mux Data
- Grafana
- Prometheus
- kommerzielle Streaming-Plattformen

m-trace soll diese Systeme ergänzen, indem es lokale Reproduzierbarkeit, Player-Metriken, Stream-Diagnose und Observability verbindet.

---

## 12. MVP-Umfang

Der erste funktionsfähige MVP muss folgende Bestandteile enthalten:

### 12.1 Muss-Anforderungen

Der MVP wird bewusst enger gefasst. Für eine Solo-Umsetzung ist der ursprüngliche Scope zu groß. Realistisch ist ein kleiner, durchgängiger Pfad.

MVP-Ziel:

```text
MediaMTX + hls.js Demo Route + Player Events + OTel-kompatibles Eventmodell + Dashboard-Anzeige
```

Muss-Anforderungen für die `0.1.x`-Phase (Gesamt-MVP). Die einzelnen
Items werden mit Patch `1.1.0` auf die drei Sub-Releases verteilt
(siehe §13.1–§13.3 für die RAK-Zuordnung):

- `0.1.0` Backend Core + Demo-Lab — Pflicht: MVP-1, MVP-2, MVP-7
  (Compose Core), MVP-8, MVP-9, MVP-11, MVP-16, MVP-17, MVP-18.
  Nicht in `0.1.0`: MVP-3 (Dashboard), MVP-4 (Demo-Route), MVP-5
  (Player-SDK), MVP-12..MVP-14 (Dashboard-Anzeigen), MVP-10/MVP-15
  (Observability).
- `0.1.1` Player-SDK + Dashboard — kommen MVP-3, MVP-4, MVP-5,
  MVP-6, MVP-12, MVP-13, MVP-14 hinzu.
- `0.1.2` Observability-Stack — kommen MVP-10, MVP-15
  (Pflicht-Anteile, siehe §7.9 Hinweis aus Patch `1.1.1`/`1.1.2`)
  hinzu; Soll-Anteile MVP-28, MVP-29 als observability-Profil.

Muss-Anforderungen für die Gesamt-`0.1.x`-Phase:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| MVP-1 | Muss | Mono-Repo-Struktur |
| MVP-2 | Muss | eine Backend-App unter `apps/api` |
| MVP-3 | Muss | eine Web-App unter `apps/dashboard` |
| MVP-4 | Muss | Demo-Player als `/demo`-Route im Dashboard, nicht als separate App |
| MVP-5 | Muss | `packages/player-sdk` mit hls.js-Adapter |
| MVP-6 | Muss | pragmatische SDK-Struktur ohne vollständige Hexagon-Ceremony |
| MVP-7 | Muss | Docker Compose Setup |
| MVP-8 | Muss | MediaMTX als erster Media Server |
| MVP-9 | Muss | FFmpeg-Teststream |
| MVP-10 | Muss | OpenTelemetry-kompatibles Eventmodell |
| MVP-11 | Muss | API-Endpunkt für Playback-Event-Batches |
| MVP-12 | Muss | einfache Session-Liste |
| MVP-13 | Muss | einfache Event-Anzeige |
| MVP-14 | Muss | einfache eingebaute Session-/Trace-Ansicht im Dashboard |
| MVP-15 | Muss | Prometheus nur für aggregierte Metriken |
| MVP-16 | Muss | lokale Speicherung per In-Memory oder SQLite |
| MVP-17 | Muss | README mit Schnellstart |
| MVP-18 | Muss | CHANGELOG mit initialem Eintrag |

Nicht im `0.1.0`-MVP:

> **Hinweis (Patch `1.1.12`)**: Die folgende Tabelle dokumentiert
> historisch *aus dem `0.1.0`-MVP herausgehaltene* Themen. Sie ist
> ausdrücklich **keine** offene Muss-Lücke der heutigen `0.9.x`-
> Reihe. Der jeweils aktuelle Lieferstand bzw. Scope-Status steht
> in der „Status (Patch `1.1.12`)"-Spalte; die ursprüngliche
> Spalte „Prioritaet" bleibt für Audit-Nachvollziehbarkeit
> erhalten. Verbindlich ist der Status-Eintrag.

| Kennung | Prioritaet | Anforderung | Status (Patch `1.1.12`) |
|---|---|---|---|
| MVP-19 | Muss (historisch) | separate `apps/demo-player` | **Anders entschieden:** Demo-Player läuft als `/demo`-Route im Dashboard (`MVP-4`, `Muss`). Eine separate `apps/demo-player`-App ist nicht geplant; wenn sie kommt, ist sie Folge-Scope und triggert ein eigenes `MVP-`-Item. |
| MVP-20 | Muss (historisch) | separate `apps/analyzer-api` | **Erfüllt anders, externe API deferred:** `apps/analyzer-service` ist seit `0.3.0` der interne HTTP-Wrapper; `@pt9912/stream-analyzer` deckt Library/CLI-Nutzung ab. **Patch `1.1.20` (Tranche-Notiz):** eine nach außen exponierte `apps/analyzer-api` wird erst bei konkretem externem Konsumenten, Auth-/Rate-Limit-/SSRF-/Retention-/Contract-Nachweis und eigenem Folgeplan reaktiviert. |
| MVP-21 | Muss (historisch) | `packages/stream-analyzer` als fertiges Paket | **Erfüllt:** Paket steht seit `0.3.0` (RAK-22..RAK-28); DASH-Manifest-Analyse ergänzt in `0.9.0` (RAK-58). |
| MVP-22 | Muss (historisch) | Tempo als Pflichtkomponente | **Bewusst gegenteilig entschieden:** Tempo ist `Kann` (`MVP-35`) und ein optionales Compose-Profil (ADR-0003); die produktive Trace-Korrelation läuft Tempo-unabhängig (RAK-32 in `0.4.0`). |
| MVP-23 | Muss (historisch) | Mimir oder ClickHouse | **Out of scope:** Production-Grade-Storage-Backends wie Mimir/ClickHouse sind nicht Bestandteil der ersten Projektphase. Bleibt höchstens Folge-ADR, wenn Multi-Tenant-/Skalierungs-Anforderung konkret wird (siehe Roadmap). |
| MVP-24 | Muss (historisch) | WebRTC | **Erfüllt:** WebRTC-Lab ab `0.7.0` (RAK-47..RAK-50), produktiver Player-SDK-Adapter ab `0.8.0` (RAK-51..RAK-55), Drift-Smoke ab `0.9.0` (RAK-56). |
| MVP-25 | Muss (historisch) | SRT-Health-View | **Erfüllt:** SRT-Health-View ab `0.6.0` (RAK-41..RAK-46). |
| MVP-26 | Muss (historisch) | Multi-Tenant-Betrieb | **Out of scope:** Multi-Tenant-SaaS-Betrieb ist nicht Bestandteil der ersten Projektphase. Bleibt Folge-ADR-Trigger; siehe Roadmap. |


### 12.2 Soll-Anforderungen

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| MVP-27 | Soll | SQLite-Persistenz statt reinem In-Memory |
| MVP-28 | Soll | Grafana-Dashboard für Aggregate |
| MVP-29 | Soll | einfache OTel-Collector-Konfiguration |
| MVP-30 | Soll | rudimentäre HLS-Manifest-Prüfung als interner Spike |
| MVP-31 | Soll | WebSocket oder SSE für Live-Updates |
| MVP-32 | Soll | CI mit GitHub Actions |


### 12.3 Kann-Anforderungen

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| MVP-33 | Kann | eigenständiger Stream Analyzer als `packages/stream-analyzer` |
| MVP-34 | Kann | CLI für Stream Analyzer |
| MVP-35 | Kann | Tempo-Integration |
| MVP-36 | Kann | SRS-Beispiel |
| MVP-37 | Muss | DASH-Analyse — **Hochstufung in Patch `1.1.11`** entsprechend NF-12 (DASH-Analyse, Muss). Die Kann-Stufung in dieser Tabelle bleibt als historischer Stand bis `1.1.10` erhalten; verbindlich ist die Muss-Stufung in §13.11 (RAK-58). **Patch `1.1.12` (Lieferstand-Vermerk):** in `0.9.0` ausgeliefert (DASH-MPD-Pfad im `@pt9912/stream-analyzer`); kein offener Folge-Scope. |
| MVP-38 | Muss (`0.11.0`-Scope, Patch `1.1.14`) | SRT-Ingest-Beispiel — präzisiert auf lokalen SRT-/RTMP-Ingest-Control-Smoke (`make smoke-ingest-control`) für MediaMTX-nahe Lab-Artefakte (Generierung/Validierung). Historische Kann-Stufung bleibt als auditierbarer Stand bis `1.1.13` erhalten; verbindlich ist die Muss-Stufung im `0.11.0`-Lab-Control-Scope (siehe RAK-68 in §13.13). |
| MVP-39 | Kann | SRT-Health-View |
| MVP-40 | Kann | Persistenz mit PostgreSQL. **Patch `1.1.18` (`0.13.0`):** Entscheidungspfad als Seed- oder Defer-Scope; SQLite bleibt lokaler Standard-Store und darf nicht implizit durch Postgres als Pflichtabhängigkeit ersetzt werden. |
| MVP-41 | Kann | ClickHouse- oder VictoriaMetrics-Anbindung. **Patch `1.1.18` (`0.13.0`):** Vergleichspfad umfasst ClickHouse, VictoriaMetrics und Mimir oder eine begründete gleichwertige Option; Ergebnis ist `proceed`, `defer` oder `POC` mit klaren Erfolgskriterien. |
| MVP-42 | Kann | Kubernetes-Manifeste. **Patch `1.1.12` (Status-Vermerk):** bleibt `Kann` und Folge-Plan; Strukturanker `deploy/k8s/` ist mit `0.9.6` angelegt, aber leer (kein Production-Ready-K8s, siehe `NF-18`-Patch und `deploy/README.md`). **Patch `1.1.18` (`0.13.0`):** optionale Beispielmanifeste oder ein K8s-Decision-Record sind zulässig; kein Production-Ready-K8s-Versprechen und keine K8s-Smoke-Pflicht ohne R-9-Entscheidung. |
| MVP-43 | Kann | Devcontainer. **Patch `1.1.18` (`0.13.0`):** Devcontainer wird als reproduzierbarer Entwicklungs-Seed geliefert oder mit Begründung und Reaktivierungs-Trigger deferred; er darf die lokale Standardentwicklung nicht ersetzen. |
| MVP-44 | Kann | Release-Automatisierung. **Patch `1.1.18` (`0.13.0`):** Automatisierung muss sichere Gates, Dry-Run-/Rollback-Regeln und explizite manuelle Freigabe enthalten; automatische Veröffentlichung ohne Human Approval ist ausgeschlossen. |


---

## 13. Release-Plan

Die `0.1.x`-Phase ist in drei Sub-Releases geschnitten (Patch `1.1.0`),
damit jeder Schritt einen demonstrierbaren Eigenwert hat und der
Gesamt-MVP-Scope nicht in einem einzelnen Release-Cycle landet:

- `0.1.0` Backend Core + Demo-Lab — curl-driven End-to-End.
- `0.1.1` Player-SDK + Dashboard — Browser-getrieben, UI sichtbar.
- `0.1.2` Observability-Stack — Aggregat-Metriken und Traces.

`0.2.0` schließt mit dem publizierbaren Player-SDK an wie ursprünglich.

### 13.1 Version 0.1.0: Backend Core + Demo-Lab

Ziel: Ein Entwickler kann das Repository klonen und ein lokales Lab mit MediaMTX-basiertem Teststream und Backend-API starten; Player-Events werden per `curl` (oder anderem HTTP-Client) an die API gesendet, Sessions sind über die API abfragbar.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-1 | Muss | `make dev` startet die in `0.1.0` erforderlichen Pflicht-Dienste (`api`, `mediamtx`, `stream-generator`). |
| RAK-3 | Muss | API ist erreichbar (`/api/health` liefert `200`, drei Pflicht-Endpoints aus dem Spike plus die zwei Stream-Sessions-Endpoints). |
| RAK-4 | Muss | Teststream läuft über MediaMTX. |
| RAK-6 | Muss | API nimmt Events an (`POST /api/playback-events` mit gültigem Token). |
| RAK-8 | Muss | README/Local-Development-Doku beschreibt den `0.1.0`-Quickstart reproduzierbar. |

### 13.2 Version 0.1.1: Player-SDK + Dashboard

Ziel: Browser-Player auf der Demo-Route sendet echte Playback-Events an die API; Dashboard zeigt Sessions, Events und Status.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-2 | Muss | Dashboard ist erreichbar; `make dev` startet zusätzlich den `dashboard`-Service. |
| RAK-5 | Muss | Player-SDK sendet hls.js-basierte Events. |
| RAK-7 | Muss | Dashboard zeigt empfangene Events und einfache Session-Zusammenhänge. |

### 13.3 Version 0.1.2: Observability-Stack

Ziel: Optionales `observability`-Compose-Profil bringt Prometheus, Grafana und OTel-Collector additiv zum Core-Stack; Aggregat-Metriken und Traces sind sichtbar.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-9 | Muss | Prometheus enthält nur aggregierte Metriken (Cardinality-Regeln aus §7.10 eingehalten). |
| RAK-10 | Soll | Player-Session-Traces sind vorbereitet oder exemplarisch sichtbar. |

RAK-8 wird in `0.1.2` ergänzt um die Doku zum `observability`-Profil.

---

### 13.4 Version 0.2.0: Publizierbares Player SDK

Ziel: Das Player-SDK wird vom MVP-Prototyp zu einem eigenständig nutzbaren und dokumentierten npm-Paket ausgebaut.

Abgrenzung zu `0.1.0`:

`0.1.0` beweist den End-to-End-Pfad mit hls.js-Adapter und Event-Ingest.  
`0.2.0` stabilisiert das SDK als wiederverwendbares Paket mit Public API, Tests, Dokumentation und Versionierungsstrategie.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-11 | Muss | SDK ist als npm-Paket baubar und lokal installierbar. |
| RAK-12 | Muss | Public API ist dokumentiert. |
| RAK-13 | Muss | Event-Schema ist versioniert. |
| RAK-14 | Muss | hls.js-Adapter ist getestet. |
| RAK-15 | Muss | HTTP-Transport ist getestet. |
| RAK-16 | Soll | OTel-Transport ist vorbereitet oder experimentell nutzbar. |
| RAK-17 | Muss | SDK unterstützt Batching, Sampling und Retry-Grenzen. |
| RAK-18 | Muss | SDK hält das definierte Performance-Budget ein. |
| RAK-19 | Muss | Browser-Support-Matrix ist dokumentiert. |
| RAK-20 | Muss | Beispielintegration in der Dashboard-Route `/demo` ist dokumentiert. |
| RAK-21 | Muss | Kompatibilität zwischen SDK-Version und Event-Schema wird in CI geprüft. |


---

### 13.5 Version 0.3.0: Stream Analyzer

Ziel: HLS-Streams können analysiert werden.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-22 | Muss | HLS Manifest kann geladen werden. |
| RAK-23 | Muss | Master Playlist kann erkannt werden. |
| RAK-24 | Muss | Media Playlist kann erkannt werden. |
| RAK-25 | Muss | Segment-Dauern werden geprüft. |
| RAK-26 | Muss | Ergebnis wird als JSON ausgegeben. |
| RAK-27 | Muss | API kann Analyzer nutzen. |
| RAK-28 | Muss | CLI-Grundlage existiert. |

---

### 13.6 Version 0.4.0: Erweiterte Trace-Korrelation

Ziel: Die in `0.1.0` vorbereitete OTel-Grundlage wird zu einer nutzbaren Korrelationsschicht ausgebaut.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-29 | Muss | Player-Session-Traces werden konsistent erzeugt. |
| RAK-30 | Soll | Manifest-Requests, Segment-Requests und Player-Events werden in einem Trace zusammengeführt, soweit technisch möglich. |
| RAK-31 | Kann | Tempo kann optional als Trace-Backend verwendet werden. |
| RAK-32 | Muss | Dashboard kann Session-Verläufe auch ohne Tempo einfach anzeigen. |
| RAK-33 | Muss | Prometheus bleibt auf aggregierte Metriken beschränkt. |
| RAK-34 | Muss | Dropped-, Rate-Limited- und Invalid-Event-Metriken sind sichtbar. |
| RAK-35 | Muss | Dokumentation beschreibt Cardinality-Grenzen und Sampling-Strategie. |


---

### 13.7 Version 0.5.0: Multi-Protocol Lab

Ziel: Das lokale Lab unterstützt weitere Streaming-Szenarien.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-36 | Muss | MediaMTX-Beispiel vorhanden. |
| RAK-37 | Muss | SRT-Beispiel vorhanden. |
| RAK-38 | Muss | DASH-Beispiel vorhanden. |
| RAK-39 | Soll | WebRTC-Beispiel vorbereitet. |
| RAK-40 | Muss | Beispiele sind dokumentiert. |

---

### 13.8 Version 0.6.0: SRT Health View

Ziel: SRT-Contribution-Workflows technisch sichtbar machen.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-41 | Muss | SRT-Testsetup vorhanden. |
| RAK-42 | Muss | SRT-Verbindungsmetriken werden erfasst oder importiert. |
| RAK-43 | Muss | RTT, Packet Loss, Retransmissions und Bandbreite werden angezeigt. |
| RAK-44 | Muss | Dashboard enthält eine SRT-Health-Ansicht. |
| RAK-45 | Muss | Dokumentation erklärt typische SRT-Fehlerbilder. |
| RAK-46 | Muss | SRT-Metriken werden OTel-kompatibel modelliert. |

### 13.9 Version 0.7.0: WebRTC-Lab-Erweiterung

Ziel: Den WebRTC-Vorbereitungspfad aus `0.5.0` (RAK-39, Lab-Skelett mit
Doku-only `examples/webrtc/`) zu einer produktiven Lab-Erweiterung
ausbauen — inklusive lokalem WHIP-/WHEP-Endpoint, opt-in Smoke und
kontrolliertem Telemetrie-Schnitt. Kein Vermischen mit dem
`hls.js`-Demo-Pfad in `apps/dashboard`; WebRTC bleibt eigenständig.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-47 | Muss | WebRTC-Lab-Setup mit lokalem WHIP-/WHEP-Endpoint und Compose-Stack `mtrace-webrtc` (analog `examples/srt/`/`examples/dash/`); Project-Name-Konvention aus `examples/README.md` ist eingehalten. |
| RAK-48 | Muss | `make smoke-webrtc-prep` prüft die Vorbereitungsgrenze (Endpoints antworten, Compose-Stack hochgefahren, kein Playback-Qualitäts- oder `getStats()`-Anspruch) und ist als opt-in Target dokumentiert (analog `make smoke-srt`/`make smoke-dash`). |
| RAK-49 | Soll | `getStats()`-Subset für produktive Telemetrie ist als bounded Allowlist in [`spec/telemetry-model.md`](telemetry-model.md) dokumentiert; eine Schema-Drift-Strategie zwischen Browser-Versionen (Chromium/Firefox/Safari) ist beschrieben. |
| RAK-50 | Kann | Browser-Handcheck ist in `examples/webrtc/README.md` als manueller Verifikationspfad dokumentiert; ergänzt den Smoke um eine Operator-sichtbare Stelle. |
| RAK-51 | Kann | `@pt9912/player-sdk` exposed einen optionalen WebRTC-Adapter-Pfad ohne Vermischung mit dem `hls.js`-Pfad; Public-API bleibt abwärtskompatibel. **Hinweis (Patch `1.1.10`):** RAK-51 ist in §13.10 für `0.8.0` zu „Muss" hochgestuft; §13.9 bleibt als historische Aussage für `0.7.0` bestehen. |

### 13.10 Version 0.8.0: Player-SDK-WebRTC-Adapter

Ziel: Den WebRTC-Lab-Pfad aus `0.7.0` (RAK-47..RAK-50, lokales WHIP-/
WHEP-Compose, Vorbereitungs-Smoke, Telemetrie-Spec §3.5) in einen
produktiven Player-SDK-Adapter überführen. RAK-51 wird aus dem
Kann-Status in §13.9 hochgestuft; vier neue RAK-52..RAK-55 sondern die
Sub-Items aus (Public-API + hls.js-Trennung, produktive WebRTC-
Telemetrie auf bounded Allowlist, Compat-Tests).

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-51 | Muss | `@pt9912/player-sdk` exposed einen produktiven WebRTC-Adapter-Pfad ohne Vermischung mit `hls.js`; Public-API bleibt abwärtskompatibel. **Hochstufung von §13.9 „Kann" auf „Muss".** |
| RAK-52 | Muss | Public-API für Adapter-Auswahl (z. B. `attachHlsJs(...)` / `attachWebRtc(...)`) ist dokumentiert; hls.js-Pfad bleibt Default und unverändert; opt-in pro Player-Instanz. Pack-Smoke und Browser-Support-Matrix erweitert. |
| RAK-53 | Soll | Produktive WebRTC-Telemetrie auf bounded Allowlist aus [`spec/telemetry-model.md`](telemetry-model.md) (`connection_state`, `ice_state`, `dtls_state`); `mtrace_webrtc_*`-Counter im API-Ingress; `scripts/smoke-observability.sh` spiegelt die WebRTC-Forbidden-Liste aus §3.1. |
| RAK-54 | Soll | `getStats()`-Sammlung im SDK aktiv; Muss-/Soll-Felder pro `RTCStatsType`-Gruppe aus [`spec/telemetry-model.md`](telemetry-model.md) werden geliefert. Die Muss-Felder sind per Contract plus Metrik-/Read-Pfad nachgewiesen; Schema-Drift-Strategie ist im Adapter-Code umgesetzt. R-12 ([[`docs/planning/in-progress/risks-backlog.md`](../docs/planning/in-progress/risks-backlog.md)](../docs/planning/in-progress/risks-backlog.md)) wird ab diesem Punkt release-blockierend. |
| RAK-55 | Kann | Browser-E2E-Smoke (Playwright) für den WebRTC-Adapter-Pfad gegen das `examples/webrtc/`-Lab; opt-in im CI-Workflow. |

### 13.11 Version 0.9.0: Drift-Smoke + SRS-Lab + DASH-Analyse

Ziel: Drei thematisch getrennte, einzeln zu kleine Liefergegenstände in einem Minor-Release bündeln. Erstens den Browser-`getStats()`-Schema-Drift aus R-12 von einem manuellen Drift-Review-Gate auf einen automatisierten Smoke heben. Zweitens das SRS-Lab analog zu `examples/srt/`/`examples/dash/`/`examples/webrtc/` als opt-in Multi-Protocol-Lab-Pfad ausliefern. Drittens die offene NF-12-Pflicht (DASH-Analyse, Muss) durch eine produktive DASH-Manifest-Analyse im `@pt9912/stream-analyzer` einlösen. MVP-37 wird entsprechend NF-12 von „Kann" auf „Muss" hochgezogen (siehe §12.3-Patch-Note); §13.5 (`0.3.0`-HLS-Analyzer) bleibt unverändert.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-56 | Soll | Browser-Drift-Smoke (Playwright, opt-in `make smoke-webrtc-stats-drift`) probt `getStats()` aus echten Browser-Versionen (Chromium, Firefox; Safari/WebKit opt-in über `MTRACE_WEBRTC_DRIFT_BROWSERS`) gegen das `examples/webrtc/`-Lab und vergleicht die Reports gegen die `webrtc.*`-Allowlist aus [`spec/telemetry-model.md`](telemetry-model.md).2. Treffer eines unbekannten Enum-Werts oder fehlender Muss-Felder bricht den Smoke; Soll-Felder werden geloggt aber nicht release-blockierend geprüft. Nightly-CI-Job (`schedule: cron`) führt den Smoke aus; bei Failure wird (opt-in über `secrets.DRIFT_AUTO_ISSUE`) ein Issue mit Browser-Version und Drift-Befund erstellt. Mit Auslieferung wandert R-12 im `risks-backlog.md` von „release-blockierend ab nächstem Browser-Major-Bump" auf „automatisiert detektiert, Drift bricht den Drift-Smoke"; das Manuell-Review entfällt. |
| RAK-57 | Kann | SRS-Lab-Beispiel `examples/srs/` (Project `mtrace-srs`, analog `examples/srt/`/`examples/dash/`/`examples/webrtc/`): eigenständiger Compose-Stack mit `ossrs/srs:5`-Image gepinnt, FFmpeg-Publisher analog `examples/srt/ffmpeg-srt-loop.sh`, Host-Port-Schnitt kollisionsfrei zu Core-Lab/`mtrace-srt`/`mtrace-dash`/`mtrace-webrtc`; `examples/srs/README.md` auf 7-Punkt-Standard analog der anderen Beispiele; opt-in `make smoke-srs` (endpoint-/compose-only, kein Playback-/Telemetrie-Anspruch). Hebt MVP-36 auf eingelöst, ohne MVP-Priorität zu ändern; [`docs/user/local-development.md`](../docs/user/local-development.md) §2.7 Port-Quickref nachgezogen. |
| RAK-58 | Muss | DASH-Manifest-Analyse im `@pt9912/stream-analyzer`: Auto-Detection von DASH-MPD-Eingaben (XML-Header `<?xml`/`<MPD`, Content-Type-Heuristik `application/dash+xml`); Manifest-Loader von HLS-only auf HLS+DASH generalisiert; MPD-Parser deckt `MPD/Period/AdaptationSet/Representation/SegmentTemplate`-Hierarchie für VOD-MPD und einfache Live-MPD ab; JSON-Result-Schema bekommt `analyzerKind: "dash"` als zweiten Wert (HLS bleibt unverändert, additiv); Mindest-Felder im Result: `playlistType: "dash"`, `summary.itemCount` (Anzahl Representations), `details.adaptationSets` (Array mit `mimeType`, `codecs`, `bandwidth`, `width`/`height`). Fehlercode `manifest_not_hls` bleibt nur für den HLS-Parser-/HLS-Kompat-Pfad; für Eingaben, die weder HLS noch DASH sind, kommt ein additiver Public-Code (z. B. `manifest_not_supported`) im Stream-Analyzer, der API-Domain (`apps/api/hexagon/domain/stream_analysis.go`), HTTP-Status-Mapping, API-Metrik-Allowlist und CLI/API-Tests durchgereicht; Fehlermeldungen dürfen nicht mehr behaupten, eine DASH-MPD sei „kein HLS-Manifest". Analyzer-Wire-Vertrag (`spec/contract-fixtures/analyzer/` plus Go-Testdata-Kopien) wird um zwei DASH-Beispiele erweitert. `apps/api`-Adapter reicht `analyzerKind` aus dem Analyzer-Result ins Domain-Modell durch; HLS-Pfad bleibt grün und unverändert. **Hochstufung von §12.3 MVP-37 „Kann" auf „Muss" entsprechend NF-12 (Erweiterbarkeit, Muss).** |
| RAK-59 | Kann | DASH-CLI-Pfad: `pnpm m-trace check <url-or-file.mpd>` detektiert MPD und liefert dasselbe JSON-Result wie der Library-Pfad (Dispatcher anhand Content-Type oder Datei-Endung); `make smoke-cli` zusätzlich um eine DASH-MPD-Probe erweitert; Tests in `packages/stream-analyzer/tests/cli.test.ts` decken HLS- und DASH-Pfad parallel. |

### 13.12 Version 0.10.0: CMAF-Analyse (NF-13)

Ziel: Die offene `NF-13`-Pflicht (CMAF-Analyse, Muss) im
Stream-Analyzer-Scope schließen — nicht über einen neuen
Manifesttyp, sondern über additive HLS-/DASH-CMAF-Signale plus
begrenzte binäre CMAF-Konformitätsprüfung ausgewählter Init-/
Media-Segmente. Der bisherige Lastenheft-Begriff
„CMAF-Vollanalyse" wird in §8.3 normativ präzisiert: vollständig
heißt vollständig für den Analyzer-Scope aus `0.10.0`, nicht
vollständige Prüfung aller Segmente, Codecs, Byte-Ranges oder
Player-Laufzeitpfade.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-60 | Muss | CMAF-Scope ist normativ begrenzt: manifestbasierte Signalanalyse plus begrenzte binäre Prüfung ausgewählter HLS-/DASH-Init- und Media-Segmente; das Lastenheft präzisiert „CMAF-Vollanalyse" als vollständige Erfüllung dieses Analyzer-Scopes, nicht als vollständige Segmentset-/Codec-/Player-Prüfung. Out of scope sind Low-Latency-CMAF (`#EXT-X-PART`, chunked CMAF), CDN-/Byte-Range-Vollprüfung und Player-SDK-CMAF-Playback. |
| RAK-61 | Muss | HLS-CMAF-Signale: `EXT-X-MAP`, fMP4-Segmentmuster (`.m4s`/`.cmfv`/`.cmfa`) und relevante Tags erzeugen stabile `details.cmaf.signals[]` mit Confidence-Semantik (`binary` > `manifest` > `inferred`) im Analyseergebnis. `EXT-X-MAP` und `#EXT-X-BYTERANGE` werden strukturiert extrahiert; HLS-Master-Summaries bleiben konservativ ohne `binary`-Objekt. |
| RAK-62 | Muss | DASH-CMAF-Signale: MPD-`mimeType` (`video/mp4`/`audio/mp4`/`application/mp4`), `codecs`, `SegmentTemplate`/`SegmentList` und Initialization-Informationen erzeugen stabile `details.cmaf.signals[]` mit Confidence-Semantik; MP4-MIME allein gilt nur als `confidence:"inferred"`-Indiz, nicht als CMAF-Konformitätsnachweis. Vererbung von `BaseURL`/`SegmentTemplate` auf `MPD`/`Period`/`AdaptationSet`/`Representation`-Ebene wird deterministisch aufgelöst. |
| RAK-63 | Muss | CLI, API-Adapter, Contract-Fixtures (`spec/contract-fixtures/analyzer/` plus Go-Testdata-Kopien) und User-Doku führen CMAF-Signale additiv durch; bestehende HLS-/DASH-Smokes bleiben unverändert grün. `details.cmaf` lebt unter den bestehenden HLS-/DASH-Detail-Objekten — kein neuer Top-Level-Envelope-Wert, kein neuer `analyzerKind`. `apps/api`-Adapter reicht `details.cmaf` über `EncodedDetails` unverändert durch. |
| RAK-64 | Muss | Binäre CMAF-Konformitätsprüfung: ISO-BMFF-Box-Parser validiert ausgewählte Init-/Media-Segmente bounded und meldet `details.cmaf.binary.status:"passed"|"failed"|"skipped"` mit nachvollziehbaren Box-/Segment-Nachweisen. Brand-Allowlist (`0.10.0`): Init-`ftyp` `cmfc`/`cmf2`, Media-`styp` `cmfs`/`cmff`/`cmfc`/`cmf2`. Pflicht-Boxen: Init `ftyp`+`moov`; Media-Fragment `styp`+`moof`+`traf`+`tfdt`+`mdat`; `sidx` optional. Bounded Segment-Loader mit Defaults `maxSegmentBytes=2_000_000`/`maxBinarySegments=6` nutzt dieselben SSRF-/Scheme-/Redirect-Regeln wie der Manifest-Loader. |

### 13.13 Version 0.11.0: Ingest-Gateway / Stream Control (F-46..F-51)

Ziel: Die historisch als Kann geführten Ingest-Gateway-Funktionen
`F-46`..`F-51` werden für einen begrenzten lokalen/lab-nahen
Stream-Control-Pfad auf Release-Muss gehoben — Stream-Keys (CSPRNG,
nur Hash persistiert), `srt`/`rtmp`-Endpunkte, einfache
1:1-Routing-Regeln, MediaMTX-nahe Konfigurationsartefakte und ein
lokal reproduzierbares Lifecycle-Eventmodell. Architektur:
**Variante B** — Ingest-Control als Modul in `apps/api`, **kein**
eigener `apps/ingest-gateway`-Service in `0.11.0`. Out of Scope
sind Multi-Tenant-Control-Plane, KMS-/Vault-Secrets, produktive
Auth-Hooks, externe Provisionierung, K8s-Operator und produktive
ausgehende Webhook-Zustellung — diese Themen wandern in `0.12.0`
oder spätere Releases.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-65 | Muss | Ingest-Control-Scope ist normativ begrenzt: lokale/lab-nahe Stream-Verwaltung in `apps/api` (Variante B), keine Multi-Tenant-Control-Plane, keine produktive Secret-Verwaltung (kein KMS/Vault), keine globale Stream-Key-Rotation über mehrere Deployments, keine produktive Media-Server-Auth-Kopplung, keine automatische externe Media-Server-Provisionierung, kein K8s-Operator. |
| RAK-66 | Muss | Stream-Key-Verwaltung: Streams können angelegt, gelistet, lokal validiert und rotiert werden. CSPRNG mit mindestens 256 Bit Entropie; URL-sicheres Output-Format dokumentiert; `key_hash` und redigierter `fingerprint` getrennt vom Klartext berechnet. Klartext-Keys erscheinen ausschließlich in den Antworten von `POST /api/ingest/streams` und `POST /api/ingest/streams/{id}/rotate-key` und nicht in Logs, Fixtures, Persistenz oder Lifecycle-Events. Validate-Endpoint nutzt den vollständigen `key_hash` als verifier; `fingerprint` ist nur Anzeige-/Audit-Hilfe. |
| RAK-67 | Muss | Ingest-Endpunkt- und Routing-Modell: `srt`/`rtmp`-Endpunkte (Host/Port/Path, Lab-Stack-Hinweis, optionaler Egress-Hinweis), Stream-Ziele (`MediaServerTarget`) und 1:1-Routing-Regeln sind validiert, dokumentiert und per API/Artefakt stabil beschreibbar. Priorisierung, Fan-out, Failover und dynamisches Load-Balancing bleiben Folge-Scope. |
| RAK-68 | Muss | Media-Server-Artefakte: MediaMTX-nahe Konfigurationen für SRT und RTMP im Lab-Scope können generiert oder validiert werden (`GET /api/ingest/media-server-config`). Bestehende Multi-Protocol-Lab-Beispiele (`examples/srt/`, `examples/mediamtx/`, `examples/srs/`) und Smokes bleiben unverändert grün; SRS bleibt Kompatibilitäts-/Dokuhintergrund, ist aber kein Pflicht-Target. |
| RAK-69 | Muss | Stream-Lifecycle-Events: `stream_started` und `stream_ended` haben ein stabiles Eventmodell und werden lokal reproduzierbar empfangen oder über `POST /api/ingest/hooks/stream-{started,ended}` exemplarisch ausgelöst. Events tragen **keine** Klartext-Keys (höchstens `key_fingerprint`). Produktive ausgehende Webhook-Zustellung an externe Systeme ist nicht Teil des `0.11.0`-Nachweises und darf nicht als erfüllt behauptet werden. |
| RAK-70 | Muss | Doku, API-/Contract-Tests und Release-Smokes (`make smoke-ingest-control`) beschreiben den lokalen Stream-Control-Workflow, die Sicherheitsgrenzen (Klartext-Key nur einmal, kein produktiver Auth-Pfad) und den Unterschied zu Auth-/Tenant-Folge-Scope `0.12.0`. Out-of-Scope-Liste in [`docs/user/ingest-control.md`](../docs/user/ingest-control.md) ist konsistent mit dem Lab-Control-Scope. |

### 13.14 Version 0.12.0: Auth / Token Lifecycle (F-111..F-113)

Ziel: Die historisch als Kann geführten Erweiterungen
`F-111`..`F-113` werden für einen begrenzten Auth-/Security-Scope auf
Release-Muss gehoben — kurzlebige serverseitig signierte Session Tokens
für Browser-Telemetrie, rotierbare Project-Token-Generationen mit
Grace-Pfad und Project-gebundene Ingest Policies für Origins, CORS und
Rate-Limit-Grenzen. Architektur: **Variante B** — Auth als Modul in
`apps/api`, **kein** eigener Auth-Service in `0.12.0`. Persistenz:
SQLite über bestehende API-Persistenz plus InMemory-Testadapter;
Klartext-Tokens werden nicht persistiert. Out of Scope sind
User-/Org-Verwaltung, OAuth/OIDC, SSO, Admin-UI, mandantenfähige SaaS-
Control-Plane, KMS-/Vault-/Cloud-Secret-Manager, produktive MediaMTX-/
SRS-Auth-Hook-Kopplung, globale Stream-Key-Rotation über mehrere
Deployments und Production-Ops-Backends — diese Themen wandern in
`0.13.0` oder spätere Releases.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-71 | Muss | Auth-Scope ist normativ begrenzt: kurzlebige Session Tokens, Project-Token-Generationen und Project-Policies; keine User-/Org-Verwaltung, kein OAuth/OIDC, keine Admin-UI, keine KMS-/Vault-Pflicht, keine produktive Media-Server-Auth-Kopplung als Muss-Scope. |
| RAK-72 | Muss | Signierte Session Tokens (`F-111`): API kann kurzlebige Tokens über `POST /api/auth/session-tokens` ausstellen und über `Authorization: Bearer mtr_st_*` oder `X-MTrace-Session-Token` validieren. Tokens sind an `project_id`, `audience` (Muss-Pfad `playback-events`), Ablaufzeit und optional `session_id`/`origin` gebunden. `ttl_seconds` hat eine harte globale Obergrenze von 900 Sekunden, fehlendes `project_max_ttl_seconds` defaultet auf exakt 900, Werte `<= 0` oder oberhalb der wirksamen Project-Grenze liefern `422 auth_token_ttl_too_large` ohne stillen Clamp. `kid` erlaubt restart-stabile Signatur-Key-Rotation; alte Verify-Keys bleiben geladen, bis alle damit signierten Tokens abgelaufen sind. Klartext-Token erscheint nur in der Issuance-Antwort; Logs/Traces/Metriken/Fixtures enthalten höchstens `token_id` oder Fingerprint. Issuance-Endpoint hat eigene globale und Project-Quoten; Überschreitung liefert `429 auth_issuance_rate_limited`. Ungültige, abgelaufene oder falsch gebundene Tokens liefern stabile `401`/`403`-Fehler nach gepinnter Präzedenz (missing → invalid → revoked → expired → not-yet-valid → project-mismatch → scope-denied → policy-denied → rate-limited) ohne Secret-Leak. |
| RAK-73 | Muss | Project-Token-Rotation (`F-112`): mehrere Generationen pro Project sind modelliert (`token_id`, `project_id`, Hash/Fingerprint, Status, `not_before`, `grace_until?`, `expires_at?`, `revoked_at?`, `created_at`, `rotated_from?`); aktive, Grace-, abgelaufene und widerrufene Tokens werden deterministisch validiert. `grace_until` ist das persistierte Feld der alten Generation und die Restart-stabile Quelle der Grace-Validierung; `revoked_at` beendet Grace sofort. Persistenz speichert nur Hash/Fingerprint und Metadaten — nie Klartext. Migrations-/Rollback-Kriterien sind dokumentiert: Rollback auf eine vorherige Config darf keine bereits widerrufene oder abgelaufene Generation reaktivieren. |
| RAK-74 | Muss | Ingest Policies (`F-113`): erlaubte Origins, Methoden, Header, Session-Token-Audiences, maximale Session-Token-TTL und Rate-Limit-Parameter (Muss-Pfad: globale und Project-Buckets; Origin-/IP-nahe Buckets sind optionaler Zusatz oder Folge-Scope) sind Project-gebunden konfigurierbar und werden in den Browser-/Telemetrie-Pfaden `POST /api/playback-events` und `POST /api/auth/session-tokens` erzwungen. **`/api/ingest/*` bleibt im `0.11.0`-Token-Validierungs-Pfad** (RAK-65: lokale/lab-nahe Stream-Verwaltung) und konsumiert keine Project-Policy — der Pfad ist `curl`-/Operator-driven, kein Browser-Konsument; Project-Policy-Enforcement für Ingest-Control wäre Defense-in-Depth ohne `0.12.0`-Anwendungsfall und bleibt Folge-Scope. CORS-Preflight nutzt eine dokumentierte globale, konservative Allowlist (Methoden maximal `POST, OPTIONS`; Header maximal `Content-Type`, `Authorization`, `X-MTrace-Token`, `X-MTrace-Session-Token`, `traceparent`); `Access-Control-Allow-Origin` wird nie `*` für tokenpflichtige Browser-Telemetrie. Bekannte Origins liefern `204` mit gespiegeltem Origin, erlaubten Methoden/Headern, `Access-Control-Max-Age: 600`, `Vary: Origin` und `Cache-Control: no-store`; unbekannte Origins liefern `204` mit leerem Body ohne Allow-Origin/Methods/Headers, mit `Vary: Origin` und `Cache-Control: no-store` — keine Project- oder Origin-Enumeration. Project-spezifische Policy-Erzwingung erfolgt beim tatsächlichen `POST` der gelisteten Browser-Pfade. |
| RAK-75 | Muss | Backward Compatibility: bestehende `X-MTrace-Token`-Project-Token-Flows (Demo, SDK, Analyze-/Session-Link-Auth, `0.11.0` Ingest-Control) bleiben im `0.12.0`-Compatibility-Fenster gültig oder haben dokumentierte Migrationstests. Fremde `Authorization`-Header ohne `Bearer mtr_st_*` werden als nicht-m-trace Auth ignoriert, wenn ein gültiger m-trace Header vorhanden ist; ohne gültigen m-trace Header liefern sie `401 auth_token_missing`. Es gibt keinen stillen Fallback von einem ungültigen höher priorisierten Token auf einen gültigen niedriger priorisierten Token. |
| RAK-76 | Muss | Security-Doku, Threat Model, Contract-Fixtures und Smokes beschreiben Token-Lifecycle, Rotation, Replay-/Leakage-Grenzen, CSP-/CORS-Beispiele, GDPR-/Datenschutzgrenzen (Auth-Metadaten erweitern IP-/User-Agent-Speicherung nicht) und den Unterschied zu Production-Secret-Backends aus Folge-Scope. SDK-Doku zeigt sicheres Session-Token-Caching bis kurz vor `expires_at`, ohne Speicherung in `localStorage`/persistenten Browser-Stores. `make sync-contract-fixtures` und `make generated-drift-check` decken neue Auth-Fixtures ab. |

### 13.15 Version 0.12.5: Auth-/Ingest-Adapter (RAK-77..RAK-82)

Ziel: Die Adapter-/Wire-Pfade, die `0.12.0` (Auth / Token Lifecycle)
als Folge-Scope angelegt und `0.12.1` (Patch) als „Code-Pfad in
0.12.5" markiert hat, werden in einen umsetzbaren Minor-Release
geschnitten. Architektur: hexagonale Erweiterung um zwei neue
Driven-Ports (`IssuanceLimiterPort`, `SecretBackendPort`) plus
ENV-getriebener Ersatz des heutigen `StaticSigningKeyResolver`
durch einen `MultiKeySigningResolver`. Backwards-Compat: heutige
ENV-Werte (Single-Key, In-Process-Limiter, ENV-Secret-Lookup)
bleiben Default-Pfad — opt-in via neuen ENV-Variablen.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-77 | Muss | **Shared-State-Issuance-Limiter**: `IssuanceLimiterPort` mit Default-In-Process-Adapter (1:1 zum heutigen `InMemoryIssuanceRateLimiter`) plus SQLite-Backend-Adapter als erstem Shared-State-Pfad (Migration `V5`, atomare Counter-Erhöhung, TTL-Cleanup). ENV-Selektion `MTRACE_AUTH_ISSUANCE_LIMITER=memory|sqlite`; **globaler Default bleibt `memory`** (Backwards-Compat). Andere Werte (`redis`, `memcached`, …) lehnt der Boot-Validator mit klarem Fehler ab. Topologie-Constraint des SQLite-Adapters: nur sinnvoll bei Single-Host-Deployments mit Shared-Persistent-Volume; echte Multi-Host-Topologie braucht einen Network-Backend-Adapter und bleibt Folge-Item. RAK-74-Scope-Cut bleibt aktiv: Limiter darf nicht vor `/api/ingest/*` hängen. |
| RAK-78 | Muss | **Multi-Key-Signing-Rotation**: `MultiKeySigningResolver` ersetzt den heutigen `StaticSigningKeyResolver` als Default-Pfad mit ENV-Schema `MTRACE_AUTH_SIGNING_KEYS=<kid1>:<base64_secret1>[,<kid2>:<base64_secret2>,…]` plus `MTRACE_AUTH_SIGNING_ACTIVE_KID=<kid>`. Verify-Set umfasst alle geladenen Keys, Sign-Key kommt aus dem aktiven `kid`; alte Tokens verifizieren weiter bis zum Ablauf der Token-TTL. Restart-stabil analog `TestHMACSigner_RestartStableAcrossKeyResolverReinitialization`. Backwards-Compat: einzelner Key über alten ENV-Pfad bleibt als degenerierter Multi-Key-Resolver mit `len(keys)==1` unterstützt. Erweitert RAK-72/RAK-73 um den ENV-getriebenen Rotation-Code-Pfad zur Doku-Stand-Lieferung aus `0.12.1` (`auth.md` §5.3.1). |
| RAK-79 | Muss | **Secret-Backend-Driven-Port**: `SecretBackendPort` als neuer Driven-Port für Signing- und Project-Token-Secret-Bezug; ENV-Lookup-Adapter implementiert den Port 1:1 zum heutigen Verhalten und bleibt Default-Selektion. Mindestens ein externes Backend-Adapter-Skelett wird geliefert — **Auswahl: Vault-Adapter** (`hashicorp/vault/api`, Lab-Pfad mit `vault dev`-Server). KMS-Adapter bleibt additive Folge-Option. ENV-Selektion `MTRACE_AUTH_SECRET_BACKEND=env|vault` (KMS-Wert wird als unbekannter Wert vom Boot-Validator abgelehnt, bis ein Adapter geliefert wird). Caching, Refresh-TTL und Failure-Modus (fail-closed bei Backend-Outage) sind dokumentiert. Resttrigger „erste Operator-Anbindung an produktives Vault/KMS" und „Compliance-Audit (PCI/SOC2)" bleiben offen — Skelett-Lieferung allein erfüllt diese nicht. |
| RAK-80 | Muss | **Browser-Ingest-Policy**: `/api/ingest/*` ist optional aus Browser-Origins nutzbar — Project-Policy-Schema wird um `browser_ingest_policy.{enabled, cors_allowlist[], csrf_required, origin_pin}` (oder finalisierte Schema-Variante) erweitert. Mit aktivem Project-Policy-Eintrag wird das Browser-Origin gegen die Allowlist geprüft, CSRF-Token oder Origin-Pin sind Default-Vorgabe für den aktivierten Pfad. Ohne aktivierte Policy gilt der RAK-74-Scope-Cut weiter (heutiges 204-Verhalten). Kontrollierte Aufhebung des RAK-74-Scope-Cuts — die Aussage „RAK-74-Scope-Cut bleibt strikt" aus §13.14 wird durch RAK-80 unter der Bedingung „aktive `browser_ingest_policy`" abgelöst. |
| RAK-81 | optional (Kann) | **Auth-Bridge MediaMTX/SRS**: `MediaMTXAuthBridge`-Adapter konsumiert signierte Publish-Tokens (eigener Issuer auf Basis des bestehenden Signing-Key-Rings aus RAK-72/RAK-78). Lab-Smoke gegen echtes MediaMTX-Container-Setup; Lab-/Produktiv-Trennung bleibt. Aktivierung nur bei verfügbarer Test-Pipeline gegen echtes MediaMTX. Löst R-14 (`validate-key` ist kein Auth-Ersatz) strukturell, sofern aktiviert. |
| RAK-82 | optional (Kann) | **Outbound-Webhook für Stream-Lifecycle**: Driven-Port `OutboundWebhookDispatcher` für `stream.started`/`stream.ended`-Events; HMAC-signierte Payload mit Project-Token-abgeleitetem Secret; definierter Retry-/Timeout-/Backoff-Pfad mit Dead-Letter-Schema. Adapter-Test plus Mock-Konsument-Smoke. Aktivierung nur bei zusätzlicher Bandbreite. Löst R-16 (lokales Lifecycle-Eventmodell ohne Outbound) strukturell, sofern aktiviert. |

Out-of-Scope-Bekräftigung (nicht durch `0.12.5` erfüllt):
KMS-Backend-Adapter (additive Folge-Option), echte Multi-Host-
Topologie für den Issuance-Limiter (Network-Backend-Adapter
bleibt Folge-Item), Externe Media-Server-Provisionierung
(`R-15`), Origin-/IP-nahes Rate-Limiting (`R-22`),
Production-Backends aus `0.13.0` (Postgres, ClickHouse,
Kubernetes), OAuth/OIDC/SSO + User-/Org-/Admin-Verwaltung
(RAK-71-Out-of-Scope-Stand bleibt normativ).

### 13.16 Version 0.12.6: Auth-/Ingest-Folge-Items (RAK-83..RAK-90)

Ziel: Die nach `0.12.5` offen oder „teilweise gelöst" gebliebenen
R-N-Items aus `risks-backlog.md` §1.1 werden in einem
Folge-Items-Minor adressiert — strukturell für die acht
Items mit User-Surface (RAK-83..RAK-90); R-13 (Trivy-Re-Review)
ist Wartung ohne neue Surface und wird ohne RAK in Plan-DoD
abgehandelt. Architekturentscheidungen: zwei neue Driven-Ports
(`OriginRateLimiter`, `MediaServerProvisioner`); bestehende
Ports (`AuthSecretBackend`, `IssuanceRateLimiter`) bekommen
zusätzliche Adapter (Vault-AppRole/IAM, KMS, Redis-Network-
Backend). Backwards-Compat: heutige ENV-Werte bleiben Default;
neue Adapter sind opt-in. Wire-Erweiterungen sind strikt
additiv — der `0.11.0`/`0.12.5`-Body-Stand bleibt byte-stabil
für alte Clients.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-83 | Muss | **Time-Skew-Persistenz auf Event-Ebene** (`R-5`): Events mit `mtrace.time.skew_warning=true` werden in einer persistenten Spalte markiert (Migration `V6`); Read-Pfad (`ListSessions`, `GetSessionDetail`, SSE-Frames) echo't den Marker; Dashboard-UI zeigt einen Indikator am betroffenen Event in der Timeline. Operator-Doku in [`spec/telemetry-model.md`](telemetry-model.md) entsprechend aktualisiert. |
| RAK-84 | Muss | **`ListSessions` Bulk-Read-Performance** (`R-7`): neuer Port-Methode `ListBoundariesForSessions(ctx, sessionIDs)` in `SessionRepository`; SQLite-Adapter implementiert mit einer `IN`-Clause statt N+1; Performance-Benchmark zeigt < 200 ms p95 für 1000 Sessions pro Page. Race-Test bleibt grün. |
| RAK-85 | Muss | **Sampling-Vollständigkeits-Marker** (`R-10`): SDK-Pflicht-Feld `meta.session_sample_rate` (Wire-Float, Range `(0, 1]`) bei `sampleRate < 1`; Server normalisiert via `round(x * 1_000_000)` auf Integer-ppm (`SAMPLE_RATE_FULL = 1_000_000`); Session-Spalte `sample_rate_ppm INTEGER NOT NULL DEFAULT 1000000` (Migration `V7`); Immutability nach erstem nicht-Default-Wert via exaktem Integer-Vergleich; Drift-Counter `mtrace_sample_rate_drift_total{project_id}` mit konfigurablem Toleranz-Band (±100 ppm). Read-API liefert beide Werte (ppm + abgeleiteter Float-Display). Schema-Eintrag in `contracts/event-schema.json`. |
| RAK-86 | Muss | **SRT-Health-Cursor-Pagination** (`R-11`): `GET /api/srt/health/{stream_id}` akzeptiert `samples_cursor`-Query-Param (gemäß §7a.3 — **nicht** `cursor`-Alias) und liefert `next_cursor`-Feld; Cursor-Token kapselt `process_instance_id + (ingested_at, id)`-Position analog §10.3; `400 cursor_invalid` bei `process_instance_id`-Mismatch oder malformed-Base64/Schema-Mismatch (§7a.4); Contract-Fixture für den `400`-Pfad. |
| RAK-87 | Muss | **MediaMTX-Provisionierung** (`R-15`): neuer Driven-Port `MediaServerProvisioner` mit `Apply`/`Rollback`-API; MediaMTX-Adapter über `/v3/config/`-Pfade; ENV `MTRACE_MEDIASERVER_PROVISION_URL/_TOKEN`. Wire-Erweiterung `POST /api/ingest/streams` mit `provision`-Query-Param: **strikt additiv** — `provision=false` (Default) ist byte-stabil zum `0.11.0`-Format und liest ENV **nicht**; `provision=true` macht I/O und liefert immer `media_server_state` (vier Werte: `applied`/`partial`/`failed`/`disabled`). Provisionier-Fehler löst keinen API-State-Rollback aus. SRS bleibt Folge-Item. |
| RAK-88 | Muss | **Multi-Host-Issuance-Limiter** (`R-17`-Resttrigger): neuer Redis-Backend-Adapter `RedisIssuanceRateLimiter` implementiert `driven.IssuanceRateLimiter` über atomare Token-Bucket-Operationen (`EVAL`-Lua mit Bucket-Key-Prefix `mtrace:issuance:`); ENV-Selektor `MTRACE_AUTH_ISSUANCE_LIMITER=memory|sqlite|redis` erweitert; Pflicht-ENV `MTRACE_REDIS_ADDR`/`_AUTH`. Fail-modus Default fail-closed (`429`); Opt-in fail-open via `MTRACE_AUTH_ISSUANCE_FAIL_OPEN=1`. Memcached bleibt Folge-Item (gemeinsam mit RAK-90). |
| RAK-89 | Muss | **Produktive Vault/KMS-Anbindung** (`R-20`-Resttrigger): `VaultSecretBackend` aus `0.12.5` um produktive Auth-Mechanismen erweitert (AppRole via `role_id`/`secret_id`-Login, optional Kubernetes-ServiceAccount-Auth); neuer `KMSSecretBackend`-Adapter (AWS-KMS, Provider-API als Interface). ENV-Selektor `MTRACE_AUTH_SECRET_BACKEND=env|vault|kms` erweitert (heute wird `kms` abgelehnt). Optionaler Refresh-TTL via `MTRACE_AUTH_SECRET_BACKEND_REFRESH_SECONDS` (Default 0 = Boot-Time-Load). Compliance-Audit-Doku zu PCI-/SOC2-relevanten Konfigurationspfaden in `auth.md` §5.5. |
| RAK-90 | Muss | **Origin-/IP-Rate-Limiter** (`R-22`): neuer Driven-Port `OriginRateLimiter` mit `Allow(ctx, key)`-Methode (key = `client_ip` oder `Origin`-Header-Hash); Adapter `InMemoryOriginRateLimiter` und `RedisOriginRateLimiter` (gemeinsamer Redis-Server mit RAK-88, Bucket-Key-Prefix `mtrace:origin:`). **Kein SQLite-Adapter** — nicht Multi-Host-tauglich. ENV-Selektor `MTRACE_ORIGIN_RATE_LIMITER=disabled|memory|redis` (Default `disabled`, Backwards-Compat); `sqlite`/`memcached`-Werte werden vom Boot-Validator mit präzisen Begründungs-Fehlermeldungen abgelehnt. Integration vor `POST /api/auth/session-tokens` und `POST /api/playback-events`. `X-Forwarded-For`-Trust-Boundary via `MTRACE_TRUST_FORWARDED_FOR=1` opt-in. |

Out-of-Scope-Bekräftigung (nicht durch `0.12.6` erfüllt):
Memcached-Backend für RAK-88/RAK-90 (Folge-Item bei
gemeinsamer Aktivierung), persistente Dead-Letter-Queue für
Outbound-Webhooks aus `0.12.5`, Compliance-Audit-Zertifikate
(PCI/SOC2, `0.12.6` liefert nur Konfigurations-Pfade), SRS-
Provisionierungs-Adapter (MediaMTX-only), Production-Backends
aus `0.13.0`.

### 13.17 Version 0.13.0: Production / Ops Backends (RAK-91..RAK-95)

Ziel: Production-/Ops-nahe Folgepunkte aus `MVP-40`..`MVP-44`
und `NF-18` werden in einen entscheidbaren Scope überführt. Dieses
Release ist ein Decision-and-Seed-Release: Es darf optionale
Artefakte liefern, aber es verpflichtet den Standardbetrieb nicht
auf Postgres, ClickHouse/VictoriaMetrics/Mimir, Kubernetes oder
Devcontainer. SQLite und das lokale Compose-Lab bleiben die
Standardpfade, solange keine spätere ADR oder ein Folge-Plan eine
Migration verbindlich macht.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-91 | Muss | **Postgres-/Ops-Scope (`MVP-40`)**: `0.13.0` dokumentiert eine verbindliche Entscheidung für Postgres als Seed-Slice, POC oder Defer. Die Entscheidung enthält mindestens zwei Alternativen, Begründung, Migrations-/Rollbackpfad oder messbare Reaktivierungs-Trigger. SQLite bleibt lokaler Standard und wird durch Postgres nicht als versteckte Pflichtabhängigkeit ersetzt. |
| RAK-92 | Muss | **Analytics-Backend-Entscheidung (`MVP-41`)**: ClickHouse, VictoriaMetrics und Mimir (oder eine begründete gleichwertige Option) werden anhand von Komplexität, Betriebskosten, Query-Fähigkeit, Integrationsaufwand, Relevanz der Workloads und Migrationsrisiko verglichen. Ergebnis ist `proceed`, `defer` oder `POC`; ein POC braucht Erfolgskriterien, Abbruchkriterien und Zeitgrenze. |
| RAK-93 | Muss | **Kubernetes-/NF-18-Harmonisierung (`MVP-42`, `NF-18`, R-9)**: K8s bleibt ein optionaler Option-Pfad. Beispielmanifeste oder Entscheidungsnotizen dürfen geliefert werden, aber ohne Production-Ready-Zusage. Eine K8s-Smoke-Stage darf erst verbindlich werden, wenn R-9 mit Observability-Label-Allowlist, Risiko-Matrix und mindestens zwei Gegenmaßnahmen entschieden ist. |
| RAK-94 | Muss | **Devcontainer-Scope (`MVP-43`)**: Devcontainer wird als reproduzierbarer Entwicklungs-Seed geliefert oder explizit deferred. In beiden Fällen sind Begründung, lokale Standardentwicklungs-Abgrenzung und Reaktivierungs-Trigger dokumentiert. |
| RAK-95 | Muss | **Release-Automatisierung (`MVP-44`)**: Release-Automatisierung definiert automatisierte und manuelle Schritte, Owner/RACI, Branch-/Tag-/Environment-Guards, Dry-Run-Tests, Rollback-Regeln und Closeout-Nachweise. Jede Veröffentlichung braucht explizite Human Approval; automatische Veröffentlichung ohne Freigabe bleibt ausgeschlossen. |

Out-of-Scope-Bekräftigung (nicht durch `0.13.0` erfüllt):
vollständiger Production-Kubernetes-Betrieb, Managed-Cloud-Betrieb,
Multi-Tenant-SaaS-Produkt, verpflichtendes Hochvolumen-Analytics-
Backend im Standardbetrieb, automatische Veröffentlichung ohne
Human Approval, Production-Identity-/Secret-Management-Vollausbau
jenseits der bereits gelieferten `0.12.x`-Pfade.

### 13.18 Version 0.14.0: Ops Backend Follow-up (RAK-96..RAK-100)

Ziel: Die in `0.13.0` getroffenen Ops-Backend-Entscheidungen werden
als begrenzter Folge-Scope umgesetzt. `0.14.0` aktiviert Szenario C
aus dem Plan: K8s-/DevEx-/Release-Guard-Hardening. Postgres und
Analytics werden nicht als Runtime-Pflichtpfade eingeführt; ihre
Trigger und Defer-Grenzen bleiben explizit sichtbar.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-96 | Muss | **Postgres-Triggerpflege (`MVP-40`)**: `0.14.0` importiert die `0.13.0`-Entscheidung `defer-with-migration-seed`. Ein Postgres-Runtime-Adapter, DSN-Pflichtpfad oder automatischer SQLite-Export darf nur bei dokumentiert ausgelöstem Trigger entstehen. Mindestens DDL-/Replay-/Rollback-Grenzen und Owner bleiben nachvollziehbar. |
| RAK-97 | Muss | **Analytics-Triggerpflege (`MVP-41`)**: `0.14.0` hält ClickHouse/VictoriaMetrics/Mimir als deferred, solange kein konkreter Workload-Trigger ausgelöst ist. Query-Workloads, Erfolgskriterien, Abbruchkriterien und Kostenannahmen bleiben messbar; kein Analytics-System wird Standard- oder Pflichtabhängigkeit. |
| RAK-98 | Muss | **K8s-/NF-18-Seed-Hardening (`MVP-42`, `NF-18`, R-9)**: Der optionale K8s-Pfad wird validiert oder gehärtet, ohne Production-Ready-Zusage. Eine K8s-Smoke-Stage bleibt optional oder deferred, bis R-9 mit separatem K8s-Allowlist-Modus, Smoke-Scope-Trennung und Label-Cardinality-Gegenmaßnahmen geschlossen ist. |
| RAK-99 | Muss | **Devcontainer-/DevEx-Validation (`MVP-43`)**: Der Devcontainer wird als reproduzierbarer Zusatzpfad validiert. Er darf Make-/Docker-Dokumentation und lokale Standardentwicklung nicht ersetzen und darf keine neue Pflichtumgebung einführen. |
| RAK-100 | Muss | **Release-Guard-Hardening (`MVP-44`)**: Release-Automation bleibt freigabepflichtig. Guard-Dry-Runs, Fehlerfälle, Branch-/Tag-/Dirty-Worktree-Regeln und Runbook-Konsistenz werden nachgewiesen; Commit, Tag, Push und GitHub-Release bleiben manuelle Schritte. |

Out-of-Scope-Bekräftigung (nicht durch `0.14.0` erfüllt):
vollständiger Production-Kubernetes-Betrieb, Managed-Cloud-Betrieb,
Multi-Tenant-SaaS-Produkt, verpflichtendes Hochvolumen-Analytics-
Backend im Standardbetrieb, Postgres als lokaler Default-Store,
automatische Veröffentlichung ohne Human Approval.

### 13.19 Version 0.15.0: Product Scope / Analyzer Boundary (RAK-101..RAK-105)

Ziel: Nach dem Ops-Hardening in `0.14.0` werden die nächsten
Produkt- und Architekturgrenzen entschieden, bevor neue Backend-,
Analyzer- oder Plattformpfade in Code überführt werden. `0.15.0`
ist ein Decision-Release: Er darf Folgeslices vorbereiten, aber keine
neue Pflichtplattform, keinen Postgres-Default und kein
Hochvolumen-Analytics-Backend einführen.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-101 | Muss | **Zielgruppenentscheidung (§16.1)**: Die Primärzielgruppe für die nächsten Minor-Releases ist entschieden oder bewusst als Produkt-ADR deferred. Selbsthoster, kleine Teams und Broadcaster-Labs bleiben Default-Empfehlung, solange kein konkreter Plattform-Betreiber-Trigger mit Betriebsmodell, Owner und Folgeplan dokumentiert ist. |
| RAK-102 | Muss | **Analyzer-API-Boundary (`MVP-20`)**: Die historische externe `apps/analyzer-api`-Anforderung wird gegen den bestehenden internen `apps/analyzer-service` bewertet. Ergebnis ist `proceed`, `POC`, `defer` oder `anders erfüllt`; ein Go braucht konkreten externen Konsumenten, API-Grenze, Auth-/Rate-Limit-/SSRF-Grenzen und Folgeplan. **Tranche-2-Entscheidung:** externe API `defer`; interner Service plus Library/CLI erfüllen den aktuellen Zielgruppen-Scope anders. |
| RAK-103 | Muss | **Control-Plane-Scope (`F-132`)**: `apps/control-plane` wird nur als Decision-Record zugeschnitten. Aufgaben, Nicht-Ziele, Trigger, Abhängigkeiten zu Zielgruppe/Auth/Multi-Tenant und späterer Planpfad sind dokumentiert; ohne eigene Folgerelease-Freigabe entsteht keine Control-Plane-Implementierung. **Tranche-3-Entscheidung:** `defer`, kein POC in `0.15.0`; Reaktivierung nur bei konkretem Betreiberbedarf mit Auth-/Tenant-/Audit-Nachweis und eigenem Folgeplan. |
| RAK-104 | Muss | **Analyzer-Folge-Slice (`NF-13`)**: Low-Latency-CMAF, HTTP-Range-/Byte-Range-Loader, vollständigere Segmentabdeckung, Codec-Decoding und Player-SDK-CMAF-Laufzeitpfad werden gegeneinander bewertet. Höchstens ein eng begrenzter Folge-Slice wird empfohlen; alles andere bleibt deferred. **Tranche-4-Entscheidung:** HTTP-Range-/Byte-Range-Loader für manifest-referenzierte CMAF-Init-/Media-Segmente wird als einziger kleiner Folge-Slice empfohlen; alle anderen Optionen bleiben deferred. |
| RAK-105 | Muss | **Ops-Trigger-Re-Eval (`MVP-40`/`MVP-41`)**: Postgres- und Analytics-Trigger aus ADR 0005 werden erneut geprüft. Ohne erreichte Trigger bleiben Postgres als `defer-with-migration-seed` und Analytics als `defer` geführt; bei erreichtem Trigger wird ein separater Folgeplan angelegt statt still umzusetzen. |

Out-of-Scope-Bekräftigung (nicht durch `0.15.0` erfüllt):
Control-Plane-Implementierung, externe Analyzer-API-Implementierung,
Postgres-Runtime-Adapter oder Postgres-Default, automatische
SQLite-zu-Postgres-Migration, ClickHouse-/VictoriaMetrics-/Mimir-
Pflichtbackend, vollständiger Production-Kubernetes-Betrieb,
OAuth/OIDC/SSO, User-/Org-Verwaltung, Multi-Tenant-SaaS-Produkt und
automatische Veröffentlichung ohne Human Approval.

### 13.20 Version 0.16.0: Selected Product Slice / Analyzer Range Fetch (RAK-106..RAK-110)

`0.16.0` importiert genau einen freigegebenen Folgepfad aus
`0.15.0`: Szenario B, den kleinen `NF-13`-Analyzer-Slice fuer
HTTP-Range-/Byte-Range-Fetches manifest-referenzierter CMAF-Init- und
Media-Segmente.  setzt den HLS-Teil dieses Scope additiv um:
explizite Byte-Ranges werden ueber den bestehenden geschuetzten
Segment-Loader geladen und binär validiert; offset-lose oder
ungueltige Byte-Ranges bleiben skipped.
Aktuell validiert den Pfad mit TS-, Doku-, Drift- und Security-
Gates;  schliesst den Release mit Tag `v0.16.0`, Roadmap-
Update und Folgeplan `0.17.0`.

Aktuell begrenzt den Lieferumfang auf HLS-CMAF-Byte-Ranges:
`#EXT-X-MAP` mit `BYTERANGE`-Attribut fuer Init-Segmente und
`#EXT-X-BYTERANGE` fuer das erste fMP4-Media-Segment. DASH-Range-/
SegmentBase-Ausbau, Multi-Range-Requests, LL-CMAF, vollstaendige
Segmentsets und Codec-Decoding bleiben ausgeschlossen.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-106 | Muss | **Import der `0.15.0`-Entscheidung**: Genau ein Folgepfad ist aktiv. Szenario B (`NF-13` HTTP-Range-/Byte-Range-Slice) wird gewaehlt; externe Analyzer-API (`RAK-102`), Control-Plane (`RAK-103`), Postgres/Analytics (`RAK-105`) und Production-K8s bleiben deferred. |
| RAK-107 | Muss | **HTTP-Range-/Byte-Range Analyzer-Slice (`NF-13`)**: Der Slice darf nur manifest-referenzierte CMAF-Init-/Media-Segmente adressieren. Low-Latency-CMAF, vollstaendige Segmentsets, Codec-Decoding, Player-SDK-CMAF-Laufzeitpfade und nicht manifest-referenzierte Fetches bleiben out of scope. |
| RAK-108 | Konditional Muss | **Contract-/Compatibility-Nachweis**: Sobald Analyzer-Code, Result-Schema, CLI/API-Doku oder Fixtures geaendert werden, muss der Nachweis additiv und rueckwaertskompatibel sein. Ohne Wire-/Schema-Aenderung ist die Unveraendertheit explizit zu dokumentieren. |
| RAK-109 | Muss | **Operational-/Security-Grenzen**: Vor Fetch-Code muessen SSRF-Schutz, Redirect-Regeln, Timeout, maximale Gesamtbytes, maximale Range-Anzahl, Status-/Content-Type-Grenzen und Laufzeitbudget festgelegt und getestet oder begruendet ausgeschlossen werden. Der Slice darf keine neue externe API, Control-Plane oder Backend-Pflicht erzeugen. |
| RAK-110 | Muss | **Closeout und Folge-Trigger**: Der Release-Closeout muss Release-Nachweis, RAK-Matrix, Changelog, Roadmap, Tag `v0.16.0`, offene Defer-Trigger und den naechsten Planpfad dokumentieren. Nicht gewaehlte Pfade bleiben sichtbar deferred. |

Out-of-Scope-Bekräftigung (nicht durch `0.16.0` erfüllt):
externe Analyzer-API, Control-Plane, Postgres-Default, Analytics-
Pflichtbackend, Production-Kubernetes, OAuth/OIDC/SSO,
Multi-Tenant-SaaS, Low-Latency-CMAF, vollstaendige Segmentset-
Abdeckung, Codec-Decoding und Player-SDK-CMAF-Laufzeitpfade.

### 13.21 Version 0.17.0: Hardening / Evidence Review (RAK-111..RAK-115)

`0.17.0` importiert den released Stand von `0.16.0` und waehlt
Szenario D: Hardening-only. Der Release startet nicht mit einer neuen
Product-Surface, sondern prueft den gelieferten HLS-CMAF-Byte-Range-
Fetch-Slice gegen Evidence, Testluecken, Compatibility und
Security-/Ops-Grenzen.  muss belegen, ob  konkrete
Hardening-Artefakte braucht oder ob Productization, Next Slice und
Switch weiter deferred bleiben.

Der Scope bleibt auf den bestehenden Analyzer-Pfad beschraenkt:
`@pt9912/stream-analyzer` Library/CLI und der interne
`apps/analyzer-service` bleiben Standard. Es entsteht durch den aktuellen Stand
kein neuer Endpoint, kein neues Result-Schema, kein neuer Runtime-
Default und kein neuer Pflichtdienst.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-111 | Muss | **Import des `0.16.0`-Ergebnisses**: `0.16.0` ist released und archiviert; Szenario D wird gewaehlt. Productization, Next Slice, Switch und Defer-Pfade werden sichtbar abgegrenzt. |
| RAK-112 | Muss | **Evidence Review und Hardening-Scope**:  prueft die `0.16.0`-Belege, Fixtures, Gate-Ergebnisse und Restgrenzen.  darf genau einen Hardening-Scope liefern oder den Folgepfad explizit deferred halten. |
| RAK-113 | Konditional Muss | **Betriebs-/Security-Haertung**: Falls Fetch-, Analyzer-, Fixture- oder Runtime-Artefakte geaendert werden, muessen SSRF-/Redirect-/Timeout-/Groessen-/Drift-/Security-Grenzen passend nachgewiesen werden. Ohne Code-/Runtime-Aenderung ist `n/a` zu begruenden. |
| RAK-114 | Konditional Muss | **Compatibility- und Migration-Gates**: Wire-, API-, Persistenz-, Contract- oder Runtime-Aenderungen brauchen Kompatibilitaets- und Drift-Nachweis. Ohne Aenderung muss die Unveraendertheit dokumentiert werden. |
| RAK-115 | Muss | **Closeout und Folgepfad**: Der Release-Closeout muss RAK-Matrix, Changelog, Roadmap, Tag `v0.17.0`, offene Trigger und den naechsten Planpfad dokumentieren. Nicht gewaehlte Pfade bleiben sichtbar deferred. |

Out-of-Scope-Bekräftigung (nicht durch `0.17.0`  erfüllt):
externe Analyzer-API, Control-Plane, Postgres-Default, Analytics-
Pflichtbackend, Production-Kubernetes, OAuth/OIDC/SSO,
Multi-Tenant-SaaS, weiterer CMAF-/DASH-/LL-CMAF-Scope,
vollstaendige Segmentset-Abdeckung, Codec-Decoding und Player-SDK-
CMAF-Laufzeitpfade.

### 13.22 Version 0.20.0: Package Publishing (RAK-116..RAK-120)

`0.20.0` schliesst die Lücke zwischen GitHub-Releases und tatsächlich
veröffentlichten npm-Artefakten. Der Release macht die beiden
Library-/CLI-Pakete erstmals über GitHub Packages installierbar:
`@pt9912/player-sdk` und `@pt9912/stream-analyzer`.

Der Scope-Wechsel von `@npm9912` auf `@pt9912` ist bewusst Teil dieses
Releases, weil GitHub Packages den Owner-/Namespace-Bezug für scoped
npm-Pakete erwartet. Der bisherige `@npm9912`-Stand wurde im Projekt
verwendet, aber nicht öffentlich veröffentlicht; es gibt deshalb keinen
externen Migrationsvertrag für bereits publizierte Pakete.

Nicht veröffentlicht werden die Anwendungen
`@pt9912/m-trace-dashboard` und `@pt9912/analyzer-service`; sie bleiben
`private: true` und werden weiterhin über Build-/Container-/Lab-Pfade
betrieben.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-116 | Muss | **Owner-Scope-Konsistenz**: Alle aktuellen Workspace-, SDK-, Analyzer-, Spec- und Test-Referenzen nutzen `@pt9912/...`; historische Done-Pläne dürfen ihren alten Kontext behalten. |
| RAK-117 | Muss | **Publishbare Pakete begrenzen**: Nur `@pt9912/player-sdk` und `@pt9912/stream-analyzer` haben `publishConfig.registry=https://npm.pkg.github.com`; Dashboard und Analyzer-Service bleiben `private: true`. |
| RAK-118 | Muss | **GitHub-Packages-Workflow**: Ein manueller Workflow kann gegen einen Git-Ref trocken oder produktiv publishen; `release.published` veröffentlicht den Release-Tag mit `GITHUB_TOKEN` und `packages: write`. |
| RAK-119 | Muss | **Release-Dokumentation**: [`docs/user/releasing.md`](../docs/user/releasing.md) beschreibt Dry-Run, produktiven Publish, automatischen Release-Hook und Package-Rollback-Grenzen. |
| RAK-120 | Muss | **Closeout und Erstveröffentlichung**: `0.20.0` bump, Changelog, Roadmap, Plan-Archiv, Tag `v0.20.0` und der erste erfolgreiche GitHub-Packages-Publish sind dokumentiert. |

### 13.23 Version 0.21.0: OCI Image Publishing (RAK-121..RAK-125)

`0.21.0` schliesst die Lücke zwischen lokalen Runtime-Image-Builds und
öffentlich pullbaren m-trace-Images. Die drei bestehenden Runtime-
Artefakte werden versioniert über GHCR veröffentlicht:
`ghcr.io/pt9912/m-trace-api`, `ghcr.io/pt9912/m-trace-dashboard` und
`ghcr.io/pt9912/m-trace-analyzer-service`.

Die Veröffentlichung ist bewusst ein Release-Artefakt, kein
Production-K8s-Go. Kubernetes bleibt gemäß `NF-18`/`MVP-42` optional;
`0.21.0` liefert nur pullbare Images und einen reproduzierbaren
Publish-Pfad. `latest`-Tags, Multi-Arch-Builds, Signierung und
Attestations bleiben Folge-Scope.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-121 | Muss | **GHCR-Namensschema**: Die drei Runtime-Images verwenden stabile, owner-konsistente Namen unter `ghcr.io/pt9912/` und werden nur mit expliziten Versions-Tags veröffentlicht; `latest` wird nicht gesetzt. |
| RAK-122 | Muss | **Make-Publish-Pfad**: Root-Targets bauen, prüfen und veröffentlichen die drei Images reproduzierbar. Der produktive Push verlangt `MTRACE_IMAGE_PUBLISH_APPROVED=1`; ohne Freigabe muss der Publish abbrechen. |
| RAK-123 | Muss | **GitHub-Actions-Workflow**: Ein manueller Workflow kann gegen einen Git-Ref trocken oder produktiv publishen; `release.published` veröffentlicht den Release-Tag mit `GITHUB_TOKEN` und `packages: write`. |
| RAK-124 | Muss | **Release-Dokumentation**: [`docs/user/releasing.md`](../docs/user/releasing.md) beschreibt Image-Dry-Run, produktiven GHCR-Publish, automatische Release-Hook-Ausführung und Rollback-Grenzen für teilweise oder fehlerhafte Image-Veröffentlichungen. |
| RAK-125 | Muss | **Closeout und Erstveröffentlichung**: `0.21.0` bump, Changelog, Roadmap, Plan-Archiv, Tag `v0.21.0` und der erste erfolgreiche GHCR-Publish der drei Runtime-Images sind dokumentiert. |

---

## 14. Akzeptanzkriterien für das Gesamtprojekt

Das Projekt gilt in der ersten Phase als erfolgreich, wenn folgende Punkte erfüllt sind:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| AK-1 | Muss | Ein neuer Entwickler kann das Projekt unter Linux lokal starten. |
| AK-2 | Muss | Die Startanleitung funktioniert ohne manuelle Sonderkonfiguration. |
| AK-3 | Muss | Die Architektur ist klar nachvollziehbar. |
| AK-4 | Muss | Die Domain-Schicht ist frameworkfrei. |
| AK-5 | Muss | Die Adapter sind technisch klar getrennt. |
| AK-6 | Muss | Mindestens ein Teststream kann abgespielt werden. |
| AK-7 | Muss | Playback-Events werden vom Browser an die API gesendet. |
| AK-8 | Muss | Events sind im Dashboard sichtbar. |
| AK-9 | Muss | Basis-Metriken sind über Observability-Komponenten sichtbar oder vorbereitet. |
| AK-10 | Muss | Das Repository ist Open-Source-tauglich dokumentiert. |
| AK-11 | Muss | Die erste Version ist als GitHub-Release veröffentlichbar. |

---

## 15. Risiken

### 15.1 Technische Risiken

| Risiko | Bewertung | Gegenmaßnahme |
|---|---|---|
| Projekt wird zu groß | Hoch | MVP strikt begrenzen |
| Streaming-Protokolle werden zu komplex | Mittel | HLS zuerst, andere später |
| Hexagonale Architektur wird übertrieben | Mittel | Nur dort einsetzen, wo Fachlogik existiert |
| Lokales Docker-Setup wird instabil | Mittel | einfache Defaults, klare Health Checks |
| Observability wird zu früh zu komplex | Mittel | erst minimale Metriken, später Ausbau |
| Browser-Verhalten unterscheidet sich stark | Hoch | MVP nur hls.js, weitere Adapter später |
| Prometheus-Cardinality explodiert | Hoch | keine Session-Labels, Traces für Per-Session-Daten |
| Player-SDK wird unterschätzt | Hoch | als eigenes Subprojekt mit Adapter-Schichten planen |
| WebRTC verwässert den MVP | Hoch | WebRTC aus Phase 1 entfernen |
| Datenschutz bremst Adoption | Mittel | Anonymisierung und sparsame Defaults früh vorsehen |
| Schema-Evolution bricht externe SDK-Versionen | Mittel | Schema-Versionierung, Contract-Tests und Kompatibilitätsprüfungen in CI |
| Project Token im Browser-Code wird zweckentfremdet | Mittel | niedrige Kritikalität, Origin-Pinning, Rate Limits und kurze Token-Rotation |

### 15.2 Projektbezogene Risiken

| Risiko | Bewertung | Gegenmaßnahme |
|---|---|---|
| Zu wenig sichtbarer Nutzen | Hoch | Demo-first Ansatz |
| README unklar | Hoch | Schnellstart prominent platzieren |
| Keine Contributor gewinnen | Mittel | gute Issues, Roadmap, klare Architektur |
| Zu viele unfertige Module | Mittel | Platzhalter reduzieren, Fokus auf lauffähigen Pfad |

---

## 16. Entscheidungen und offene Punkte

Dieser Abschnitt sammelt produkt- und architekturrelevante
Entscheidungen. Einträge können bereits entschieden sein (`resolved`)
oder als offene Folgefrage weitergeführt werden.

### 16.1 Zielgruppenentscheidung

Status: **resolved** seit Patch `1.1.20`.

Die wichtigste Produktentscheidung war:

```text
Selbsthoster und kleine Teams oder Plattform-Betreiber mit hunderten parallelen Streams?
```

**Entscheidung (Patch `1.1.20`):** Für die
nächsten Minor-Releases fokussiert m-trace Selbsthoster, kleine bis
mittlere Streaming-Teams, Broadcaster-Labs und technische
Media-/DevOps-Teams. Große Plattformbetreiber mit hunderten parallelen
Streams bleiben späterer Scope.

Diese Entscheidung beeinflusst:

- **Storage:** SQLite bleibt lokaler Standard; Postgres bleibt der
  in ADR 0005 dokumentierte Trigger-/Folgepfad.
- **Sampling:** Sampling bleibt Diagnose- und Cardinality-Schutz,
  keine Billing- oder Audience-Analytics-Semantik.
- **Cardinality:** Prometheus bleibt aggregiert und ohne Session-/
  Viewer-Labels; hochkardinale Plattformmetriken sind out of scope.
- **Multi-Tenant-Fähigkeit:** Project-/Token-Grenzen bleiben technische
  Isolation; kein SaaS-Tenant-Modell mit User-/Org-Verwaltung.
- **Betriebsmodell:** Compose und optionale Beispielpfade bleiben
  Default; K8s bleibt Beispiel-/Triggerpfad.
- **Dashboard-Komplexität:** Dashboard bleibt Diagnose- und
  Lab-Oberfläche, keine Admin-/Billing-/Fleet-Control-Plane.
- **Alerting:** Alerts bleiben technisch und operatornah, keine
  mandantenbezogenen Alerting-Workflows.

Empfehlung für den MVP:

```text
Fokus auf Selbsthoster, kleine Plattformen, Broadcaster-Labs und technische Teams.
```

Große Plattform-Betreiber werden erst adressiert, wenn ein konkreter
Stakeholder-/Operator-Bedarf mit Betreiberprofil, erwarteter Stream-/
Event-Größenordnung, Multi-Tenant-/Auth-Anforderungen, Betriebs-SLO,
Owner und Folgeplan vorliegt.

### 16.2 Weitere Entscheidungen

| Kennung | Status | Entscheidung |
|---|---|---|
| OE-1 | resolved | Projektlizenz: **MIT** (siehe `LICENSE`) |
| OE-2 | resolved | Backend-Technologie final: **Go** (siehe [`docs/adr/0001-backend-stack.md`](../docs/adr/0001-backend-stack.md)) |
| OE-3 | resolved | Datenhaltung im MVP: **SQLite als lokaler Durable-Store** ab `0.4.0` (siehe [`docs/adr/0002-persistence-store.md`](../docs/adr/0002-persistence-store.md)) |
| OE-4 | resolved | Frontend-Styling im MVP: **eigenes CSS ohne Tailwind/UI-Library** |
| OE-5 | resolved | Live-Updates: **Server-Sent Events mit Polling-Fallback**; WebSocket nicht in `0.4.0` (siehe [`docs/adr/0003-live-updates.md`](../docs/adr/0003-live-updates.md)) |
| OE-6 | resolved | CI-Zielplattform: **GitHub Actions `ubuntu-24.04`** |
| OE-7 | resolved | Release-Konvention: **trunk-based auf `main`, annotierte SemVer-Tags `vX.Y.Z`, GitHub Release aus [`CHANGELOG.md`](../CHANGELOG.md)** |
| OE-8 | resolved | npm-Paketname Player-SDK: **`@pt9912/player-sdk`** ab `0.20.0`; Lieferstände vor `0.20.0` wurden nicht öffentlich über GitHub Packages publiziert. Der historische interne Scope `@npm9912` hat daher keinen externen Migrationsvertrag. |
| OE-9 | resolved | Go Module Name final: **`github.com/pt9912/m-trace/apps/api`** |

---

## 17. Erste empfohlene Umsetzungsschritte

### Schritt 0: Backend-Technologie-Spike — abgeschlossen

Backend-Technologie wurde durch zwei lauffähige Mini-Prototypen (Go,
Micronaut) im identischen Muss-Scope entschieden. Dokumentation:

- Spike-Spezifikation: [`docs/spike/0001-backend-stack.md`](../docs/spike/0001-backend-stack.md)
- Implementierungsplan: [`docs/planning/done/plan-spike.md`](../docs/planning/done/plan-spike.md)
- API-Kontrakt: [`spec/backend-api-contract.md`](backend-api-contract.md)
- Spike-Protokoll: [`docs/spike/backend-stack-results.md`](../docs/spike/backend-stack-results.md)
- Entscheidung: [`docs/adr/0001-backend-stack.md`](../docs/adr/0001-backend-stack.md) (Status: Accepted) — **Go**

Sieger-Branch `spike/go-api` ist auf `main` als `apps/api` integriert
(siehe [`docs/planning/in-progress/roadmap.md`](../docs/planning/in-progress/roadmap.md)).

---

### Schritt 1: Repository initialisieren

- Mono-Repo-Struktur anlegen
- README.md erstellen
- CHANGELOG.md erstellen
- LICENSE hinzufügen
- Makefile hinzufügen
- `.env.example` hinzufügen

### Schritt 2: API-Grundgerüst

- Backend-App unter `apps/api` in Go (siehe [`docs/adr/0001-backend-stack.md`](../docs/adr/0001-backend-stack.md))
- Hexagon-Struktur anlegen
- Domain-Modelle für StreamSession und PlaybackEvent
- Use Case `RegisterPlaybackEventUseCase`
- In-Memory Repository
- HTTP Controller

### Schritt 3: Dashboard- und Demo-Player-Grundgerüst

- SvelteKit-App unter `apps/dashboard`
- Startseite
- Test-Player-Seite
- Stream-Sessions-Seite
- API Client
- Demo-Player-Route unter `apps/dashboard/src/routes/demo/`
- SDK-Referenzintegration innerhalb des Dashboards vorbereiten

### Schritt 4: Player-SDK

- TypeScript-Package unter `packages/player-sdk`
- HTMLVideoElement Adapter
- HTTP Event Publisher
- einfache Event-Erfassung

### Schritt 5: Docker Lab

- Docker Compose
- MediaMTX Service
- FFmpeg Teststream
- API Service
- Dashboard Service

### Schritt 6: Observability

- OTel Collector
- Prometheus
- Grafana
- erste Metriken
- Dokumentation

---

## 18. Definition of Done für den MVP

Der MVP ist fertig, wenn:

- `make dev` erfolgreich startet.
- Der Teststream lokal läuft.
- Das Dashboard im Browser erreichbar ist.
- Der Test-Player den Stream abspielen kann.
- Das Player-SDK Events erzeugt.
- Die API Events annimmt.
- Das Dashboard Events anzeigt.
- Die Architektur in [`spec/architecture.md`](architecture.md) beschrieben ist.
- Das Eventmodell in [`spec/telemetry-model.md`](telemetry-model.md) beschrieben ist.
- Tests für zentrale Use Cases vorhanden sind.
- CI mindestens Build und Tests ausführt.
- [`CHANGELOG.md`](../CHANGELOG.md) einen Eintrag pro Release enthält.

---

## 19. Glossar

| Begriff | Bedeutung |
|---|---|
| Adapter | Technische Implementierung eines Eingangs oder Ausgangs |
| DASH | MPEG-DASH, adaptives Streaming-Protokoll |
| Domain | Fachlicher Kern der Anwendung |
| HLS | HTTP Live Streaming |
| Hexagon | Architekturmodell mit Ports und Adapters |
| Inbound Adapter | Adapter, der die Anwendung von außen aufruft, z. B. HTTP Controller |
| Media Server | Server zur Annahme, Verarbeitung und Auslieferung von Streams |
| MediaMTX | Media Server mit Unterstützung für RTSP, RTMP, HLS, WebRTC und SRT |
| Mono-Repo | Repository, das mehrere Anwendungen und Pakete gemeinsam enthält |
| OpenTelemetry | Standard für Logs, Metriken und Traces |
| Outbound Adapter | Adapter, mit dem die Anwendung externe Systeme nutzt |
| Player-SDK | Browser-Bibliothek zur Erfassung von Playback-Metriken |
| Port | Schnittstelle zwischen Hexagon und Außenwelt |
| RTMP | Real-Time Messaging Protocol |
| SRS | Simple Realtime Server |
| SRT | Secure Reliable Transport |
| Stream Analyzer | Komponente zur Analyse von Streaming-Manifesten |
| Stream Session | zusammenhängende Betrachtung einer Wiedergabe- oder Streaming-Sitzung |
| Use Case | fachlicher Anwendungsfall |
| CMAF | Common Media Application Format, Container-/Segmentierungsstandard für adaptive Streaming-Workflows |
| LL-HLS | Low-Latency HLS, Variante von HLS für geringere Latenz |
| QoE | Quality of Experience, nutzerbezogene Qualitätswahrnehmung beim Playback |
| Cardinality | Anzahl unterschiedlicher Zeitreihen-Kombinationen durch Labels, besonders relevant für Prometheus |
| OTLP | OpenTelemetry Protocol für den Transport von Traces, Metriken und Logs |
| Time Skew | Abweichung zwischen Client-Uhr und Server-Uhr |

---

## 20. Zusammenfassung

m-trace soll als Open-Source-Mono-Repo ein praxisnahes Werkzeug für Media-Streaming-Observability werden.

Der entscheidende Erfolgsfaktor ist nicht maximale Funktionsbreite, sondern ein sofort nutzbarer lokaler Demo-Pfad:

```bash
git clone <repo>
cd m-trace
make dev
```

Danach soll ein Entwickler im Browser sehen können:

- ein laufender Teststream
- Player-Events
- Stream-Sessions
- erste Metriken
- technische Diagnoseinformationen

Die Architektur muss sauber genug sein, um langfristig wartbar zu bleiben, aber pragmatisch genug, damit der MVP schnell nutzbar wird.
