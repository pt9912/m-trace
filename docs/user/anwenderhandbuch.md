# Benutzerhandbuch: m-trace

Handbuch-Version: 1.1<br>
Software-Version: 0.25.0<br>
Stand: 2026-07-13<br>
Gültigkeitsbereich: Self-hosted-Betrieb (Compose-Lab und eigene Deployments)

> Dieses Handbuch ist aufgabenorientiert: Es beschreibt, wie Sie Ihre
> Aufgabe erledigen — nicht, was jede Funktion kann. Vertiefende
> Betriebs- und Integrationsdetails stehen in den verlinkten
> Spezialdokumenten; dieses Handbuch ist der Einstiegspunkt.

## 1. Einleitung

### Zweck der Software

m-trace misst die **Wiedergabe-Qualität von Video-Streams aus Sicht der
Zuschauer**: Player (HLS, DASH, WebRTC) senden Telemetrie-Ereignisse
(Start, Rebuffering, Fehler, Qualitätswechsel) an eine API; ein Dashboard
zeigt Sessions und deren Zeitverlauf. Zusätzlich analysiert m-trace
Stream-Manifeste auf Konformität und überwacht die Gesundheit von
SRT-Zubringern.

### Zielgruppe

Dieses Handbuch richtet sich an **Selbsthoster, kleine bis mittlere
Teams, Broadcaster-Labs und technische Media-Teams**: Sie betreiben
m-trace selbst und binden das Player-SDK in eigene Seiten ein.
Grundkenntnisse in Terminal, Docker und Web-Deployments werden
vorausgesetzt; m-trace-interne Begriffe erklärt das Glossar (§9).

### Voraussetzungen

- Linux oder macOS mit **Docker** (Compose v2) und **make**.
- Für SDK-/CLI-Aufgaben: **Node.js 22** und **pnpm**.
- Freie lokale Ports **8080** (API) und **5173** (Dashboard).

## 2. Erste Schritte

### Das Lab starten

1. Repository klonen und ins Verzeichnis wechseln.
2. Optional: `cp .env.example .env` und Werte anpassen
   (die Datei dokumentiert die Lab- und Betriebs-Variablen mit
   Beispielwerten).
3. `make dev` ausführen.

**Ergebnis:** API und Dashboard laufen; ein Beispiel-HLS-Stream wird
erzeugt. Prüfen Sie:

- Dashboard: <http://localhost:5173>
- API-Health: <http://localhost:8080/api/health>

Die vollständige Ersteinrichtung (Profile, Observability-Stack,
Troubleshooting beim Start) beschreibt
[`local-development.md`](local-development.md).

### Die Demo ansehen

Öffnen Sie
<http://localhost:5173/demo?session_id=readme-demo&autostart=1> —
ein Demo-Player spielt den Lab-Stream und sendet dabei echte
Telemetrie. Die entstehende Session erscheint unter
<http://localhost:5173/sessions>.

### Überblick über die Oberfläche

Das Dashboard hat folgende Bereiche (Routen):

| Route | Aufgabe |
|---|---|
| `/sessions` | Wiedergabe-Sessions auflisten; Detailansicht mit Ereignis-Zeitverlauf |
| `/events` | Ereignisse einsehen |
| `/errors` | Fehler-Ereignisse einsehen |
| `/srt-health` | SRT-Zubringer-Gesundheit (wenn aktiviert, §3.5) |
| `/status` | System-/Aggregatstatus |
| `/demo`, `/demo-webrtc` | eingebaute Demo-Player (HLS bzw. WebRTC) |

### Grundlegende Bedienkonzepte

- **Projekt + Token**: Jede Telemetrie gehört zu einem Projekt und wird
  mit einem Token authentifiziert. Das Lab bringt das Projekt `demo`
  mit dem Token `demo-token` mit — **nur für lokale Labs**, nicht für
  produktive Deployments (§5).
- **Session**: Eine Wiedergabe-Sitzung eines Players. Alle Ereignisse
  einer Session teilen dieselbe `session_id`.
- **Alles ist API-first**: Was das Dashboard zeigt, liefert die API
  unter `/api/...`; Sie können jede Ansicht auch per `curl` prüfen.

## 3. Aufgaben ausführen

### 3.1 Player-Telemetrie aus der eigenen Anwendung senden

#### Voraussetzung

Ihre Web-Anwendung nutzt hls.js (oder WebRTC/WHEP); die m-trace-API ist
von den Zuschauer-Browsern erreichbar; Sie kennen Projekt-ID und Token.

#### Vorgehen

1. SDK installieren: `pnpm add @pt9912/player-sdk hls.js`
2. Tracker erstellen und an den Player hängen:

   ```ts
   import Hls from "hls.js";
   import { attachHlsJs, createTracker } from "@pt9912/player-sdk";

   const video = document.querySelector("video");
   const tracker = createTracker({
     endpoint: "https://ihre-mtrace-api.example/api/playback-events",
     token: "<ihr-token>",
     projectId: "<ihre-projekt-id>",
   });
   const hls = new Hls();
   attachHlsJs(video, hls, tracker);
   ```

3. Für WebRTC-Wiedergabe steht additiv `attachWebRtc(...)` bereit
   (Details: `packages/player-sdk/README.md`).

#### Ergebnis

Wiedergabe-Ereignisse laufen als Batches gegen
`POST /api/playback-events`; die Session erscheint im Dashboard unter
`/sessions`.

#### Hinweise

- Für Browser-Clients sollten Sie statt des langlebigen Projekt-Tokens
  **kurzlebige Session-Tokens** ausstellen (§3.6) und die erlaubten
  Origins des Projekts pflegen — sonst antwortet die API mit `403`.
- Eine vollständige Integrations-Referenz inkl. CORS/CSP:
  [`demo-integration.md`](demo-integration.md) und
  [`auth.md`](auth.md).

### 3.2 Sessions und Ereignisse ansehen

#### Vorgehen

1. Öffnen Sie `/sessions` im Dashboard (die Liste aktualisiert sich
   live über Server-Sent-Events, mit Polling-Fallback).
2. Wählen Sie eine Session für die Detailansicht mit Zeitverlauf
   (per Refresh-Knopf aktualisieren; lange Verläufe über
   „Load more" nachladen).

Alternativ per API:

```bash
curl -H "X-MTrace-Token: <token>" \
  "http://localhost:8080/api/stream-sessions?limit=20"
curl -H "X-MTrace-Token: <token>" \
  "http://localhost:8080/api/stream-sessions/<session-id>?events_limit=100"
```

#### Ergebnis

Sie sehen Sessions Ihres Projekts mit Ereignissen in stabiler
Reihenfolge; lange Listen sind cursor-paginiert (`next_cursor` folgen).

### 3.3 Ein Stream-Manifest analysieren

#### Vorgehen

Per CLI (ohne laufende API):

```bash
pnpm m-trace check https://cdn.example.test/manifest.m3u8
```

Oder gegen die API:

```bash
curl -X POST http://localhost:8080/api/analyze \
  -H "Content-Type: application/json" \
  -d '{"kind":"url","url":"https://cdn.example.test/manifest.m3u8"}'
```

> **Betriebshinweis:** `POST /api/analyze` ist für ungebundene Aufrufe
> (wie oben) bewusst authentifizierungsfrei und für den Betrieb im
> **internen Netz** gedacht; wird das Ergebnis per
> `correlation_id`/`session_id` an eine Session gebunden, ist das
> Projekt-Token Pflicht. Exponieren Sie die API öffentlich, gehört
> dieser Endpunkt hinter eine Egress-/Zugriffs-Beschränkung
> ([`../../spec/backend-api-contract.md`](../../spec/backend-api-contract.md) §3.6).

#### Ergebnis

Ein Analyse-Resultat mit erkanntem Typ (HLS oder DASH), Struktur-Details
und ggf. CMAF-Konformitätsbefunden. Nicht unterstützte Inhalte werden
sauber mit `manifest_not_supported` abgelehnt (HTTP `422`).

#### Hinweise

Formate, Grenzen (Fetch-Limits, Segment-Prüfung) und die volle
CLI-Referenz: [`stream-analyzer.md`](stream-analyzer.md).

### 3.4 Ingest-Streams anlegen und verwalten

Für lokale/lab-nahe Zubringer-Verwaltung (SRT/RTMP-Endpunkte,
Stream-Keys, MediaMTX-Konfiguration) stellt die API `/api/ingest/...`
bereit. Aufgaben, Schritt-für-Schritt-Abläufe und der Smoke dazu:
[`ingest-control.md`](ingest-control.md).

### 3.5 SRT-Zubringer überwachen

#### Voraussetzung

Ein MediaMTX (oder kompatibler) SRT-Endpunkt mit erreichbarer API.

#### Vorgehen

1. Auf dem API-Service `MTRACE_SRT_SOURCE_URL` auf die
   MediaMTX-API setzen (z. B. `http://mediamtx:9997`) und den Service
   neu starten. Das Log meldet „srt-health collector enabled".
2. Dashboard-Route `/srt-health` öffnen.

#### Ergebnis

Eine Tabelle je Stream mit Gesundheitszustand (u. a. RTT, Paketverlust,
Bandbreite); die Detailansicht zeigt die Sample-Historie. Bleiben
frische Samples aus, wechselt der Zustand sichtbar auf „stale".

#### Hinweise

Schwellwerte, Wire-Format und Betriebsdetails:
[`srt-health.md`](srt-health.md).

### 3.6 Kurzlebige Session-Tokens für Browser ausstellen

#### Vorgehen

```bash
curl -X POST http://localhost:8080/api/auth/session-tokens \
  -H "Content-Type: application/json" \
  -H "X-MTrace-Token: <projekt-token>" \
  -d '{"audience":"playback-events"}'
```

Das zurückgegebene Token (`mtr_st_...`) gibt der Browser-Client als
`Authorization: Bearer ...` oder `X-MTrace-Session-Token` mit.

#### Ergebnis

Der Browser sendet Telemetrie, ohne das langlebige Projekt-Token zu
kennen; das Session-Token läuft automatisch ab (TTL ≤ 900 s).

#### Hinweise

Rotation der Projekt-Tokens, Signing-Keys, Policies und die komplette
Fehler-Präzedenz: [`auth.md`](auth.md).

### 3.7 Rate-Limits konfigurieren

Der Ingest-Pfad drosselt pro Projekt, Client-IP und Origin
(Token-Bucket, Default 100 Events/s, Antwort `429` mit `Retry-After`).

- Budget anpassen: `MTRACE_RATE_LIMIT_CAPACITY` /
  `MTRACE_RATE_LIMIT_REFILL`.
- **Mehrere API-Replicas** hinter einem Load-Balancer: mit
  `MTRACE_RATE_LIMIT_BACKEND=redis` teilen sich alle Replicas **ein**
  Budget pro Projekt (sonst vervielfacht sich das Limit mit der
  Replica-Zahl). Hinter Proxy/LB zusätzlich
  `MTRACE_TRUST_FORWARDED_FOR=1` setzen, damit die Client-IP-Dimension
  echte Clients statt der Proxy-IP zählt — nur mit vertrauenswürdigem,
  XFF-setzendem Proxy aktivieren.

Alle Varianten, Fail-Verhalten und Sicherheits-Randbedingungen:
[`auth.md`](auth.md) §5.10 (Ingest), §5.9 (Origin/IP), §5.4 (Issuance).

### 3.8 Auf Postgres wechseln (mehrere Replicas)

SQLite ist der Standard-Store und für Einzel-Instanzen richtig. Für
≥ 2 API-Replicas auf einem geteilten Store:

1. `MTRACE_PERSISTENCE=postgres` und `MTRACE_POSTGRES_DSN` setzen.
2. Beim ersten Start legt die API das Schema selbst an
   (`track_commit_timestamp=on` muss auf dem Postgres aktiv sein —
   die API bricht sonst beim Start mit klarer Meldung ab).

**Bestehende SQLite-Daten mitnehmen:** Die Migration der Historie ist
ein geführter, verifizierter Ablauf (`make cutover ...`) — folgen Sie
dem Runbook [`../ops/postgres-cutover.md`](../ops/postgres-cutover.md).

## 4. Einstellungen

Die kommentierte Referenz der Lab- und Betriebs-Variablen ist
[`.env.example`](../../.env.example) im Repo-Wurzelverzeichnis; die
Auth-Spezialvariablen (Limiter-Backends, Signing-Keys, Policies)
dokumentiert [`auth.md`](auth.md) §5. Die wichtigsten:

| Variable | Wirkung | Default |
|---|---|---|
| `MTRACE_PERSISTENCE` | Store: `sqlite` \| `postgres` \| `inmemory` | `sqlite` |
| `MTRACE_SQLITE_PATH` | Pfad der SQLite-Datei | `/var/lib/mtrace/m-trace.db` |
| `MTRACE_RATE_LIMIT_CAPACITY` / `_REFILL` | Ingest-Budget je Dimension | `100` / `100` |
| `MTRACE_RATE_LIMIT_BACKEND` | Limiter: `memory` \| `redis` (geteiltes Budget) | `memory` |
| `MTRACE_TRUST_FORWARDED_FOR` | Client-IP aus `X-Forwarded-For` (nur hinter vertrautem Proxy) | aus |
| `MTRACE_SRT_SOURCE_URL` | SRT-Health-Quelle (leer = Pfad inaktiv) | leer |
| `MTRACE_REDIS_ADDR` / `_AUTH` / `_DB` | Redis-Server für die Redis-Backends | leer |

## 5. Rollen und Rechte

m-trace hat keine Benutzerkonten im Dashboard; das Zugriffsmodell sind
**Projekt-Tokens**:

- Ein Token gehört genau einem Projekt und sieht nur dessen Daten.
- **Projekt-Token** (`mtr_pt_...` bzw. Lab-Token): langlebig, für
  Server-zu-Server und zum Ausstellen von Session-Tokens; rotierbar
  mit Übergangsfrist ([`auth.md`](auth.md) §5.1–§5.2).
- **Session-Token** (`mtr_st_...`): kurzlebig, für Browser (§3.6).
- Das Lab-Projekt `demo`/`demo-token` und env-geseedete Lab-Projekte
  (`MTRACE_LAB_PROJECTS`) haben **vorhersagbare Tokens — niemals in
  produktiven Umgebungen verwenden** (beim Seeden zusätzlicher
  Lab-Projekte warnt die API im Log; `demo` ist im Lab immer da).

## 6. Betrieb: Software beziehen und aktualisieren

- **Container-Images** (empfohlen für den Betrieb):
  `ghcr.io/pt9912/m-trace-api`, `ghcr.io/pt9912/m-trace-dashboard`,
  `ghcr.io/pt9912/m-trace-analyzer-service` — jeweils mit konkretem
  Versions-Tag (z. B. `0.25.0`); ein `latest`-Tag wird bewusst nicht
  gepflegt. Kubernetes-Beispiele: `deploy/k8s/`.
- **npm-Pakete**: `@pt9912/player-sdk` (Browser-SDK) und
  `@pt9912/stream-analyzer` (Library + CLI) über GitHub Packages.
- Versionsstände und Änderungen: [`CHANGELOG.md`](../../CHANGELOG.md).

## 7. Fehlerbehebung

### Fehler: API antwortet `401 Unauthorized`

**Ursache:** Token fehlt, ist unbekannt, abgelaufen oder gehört zu
einem anderen Projekt (der `error`-Code im Body nennt die Ursache,
z. B. `auth_token_invalid`, `auth_token_expired`).

**Lösung:** Header prüfen (`X-MTrace-Token` bzw.
`Authorization: Bearer mtr_st_...`); bei Session-Tokens neu ausstellen
(§3.6); Fehler-Präzedenz in [`auth.md`](auth.md).

### Fehler: `403 Forbidden` aus dem Browser

**Ursache:** Der `Origin` Ihrer Seite steht nicht in den erlaubten
Origins des Projekts, oder die Projekt-Policy verbietet den Zugriff.

**Lösung:** Erlaubte Origins des Projekts um Ihre Seiten-URL ergänzen
(exakter Scheme+Host+Port-Vergleich); siehe [`auth.md`](auth.md) §5.6.

### Fehler: `429 Too Many Requests`

**Ursache:** Das Rate-Limit (pro Projekt, Client-IP oder Origin) ist
erschöpft; bei `{"error":"origin_rate_limited"}` hat der vorgelagerte
Origin-/IP-Limiter gegriffen.

**Lösung:** Erneut senden — beim Event-Ingest nennt `Retry-After` die
Wartezeit; der Origin-/IP-Limiter (`origin_rate_limited`) sendet
keinen `Retry-After`-Header, hier mit Backoff wiederholen (das SDK
batcht und wiederholt in beiden Fällen selbst). Dauerhaft: Budget
erhöhen bzw. bei mehreren Replicas das geteilte Redis-Backend
aktivieren (§3.7). Wichtig: `429` verwirft client-bestätigte Daten
nicht still — abgelehnte Events wurden nie angenommen.

### Fehler: `422` mit `manifest_not_supported`

**Ursache:** Die angefragte URL liefert kein unterstütztes
HLS-/DASH-Manifest (z. B. eine HTML-Seite).

**Lösung:** URL im Browser prüfen; sie muss das Manifest selbst
liefern, nicht eine Player-Seite.

### Problem: Dashboard zeigt keine Sessions

**Ursachen und Lösungen:**

1. Falsches Projekt/Token — mit `curl` gegen
   `/api/stream-sessions` (§3.2) gegenprüfen.
2. Player sendet nicht — Browser-Netzwerk-Tab auf
   `POST /api/playback-events` (Status `202`) prüfen.
3. API nicht erreichbar/healthy — `curl http://localhost:8080/api/health`.

### Problem: Port 8080 oder 5173 ist belegt

**Lösung:** Belegenden Prozess stoppen oder Port-Mapping in
`docker-compose.yml` anpassen; danach `make stop && make dev`.

### Problem: Lab-Daten zurücksetzen

**Lösung:** `make wipe` stoppt die Services und löscht **destruktiv**
das Daten-Volume (alle Sessions/Events). `make stop` allein behält die
Daten.

## 8. FAQ

**Bleiben meine Daten bei einem Neustart erhalten?**
Ja — Standard ist SQLite in einem Docker-Volume; `make stop`/`make dev`
behalten es. Nur `make wipe` löscht.

**Kann ich m-trace ohne Docker nutzen?**
Der unterstützte Weg ist das Compose-Lab bzw. die GHCR-Images. Die
Analyzer-CLI (`pnpm m-trace check`) läuft auch ohne laufende API.

**Wie viele Zuschauer verkraftet eine Instanz?**
Eine Einzel-Instanz verarbeitet im Referenz-Lab mehrere tausend Events
pro Sekunde; belastbare Messwerte und Budgets stehen in
[`../perf/budgets.md`](../perf/budgets.md).

**Verliert m-trace Events unter Last?**
Angenommene Events (`202`) sind persistiert („kein stiller Verlust" ist
Release-Gate); bei Überlast lehnt die API sichtbar mit `429` ab.

**Werden personenbezogene Daten gespeichert?**
m-trace speichert technische Wiedergabe-Telemetrie (Session-IDs,
Ereignisse, Zeitstempel) — keine Konten oder Klartext-Tokens in Logs.
Bewertung und Betroffenenrechte für Ihr Deployment liegen bei Ihnen als
Betreiber; Details zum Datenmodell: [`auth.md`](auth.md) §6.

## 9. Glossar

| Begriff | Bedeutung |
|---|---|
| Projekt | Mandant/Namensraum; alle Daten und Tokens hängen an genau einem Projekt |
| Projekt-Token | Langlebiges Zugangs-Token eines Projekts (`X-MTrace-Token`) |
| Session-Token | Kurzlebiges Browser-Token (`mtr_st_...`, TTL ≤ 900 s) |
| Session | Eine Wiedergabe-Sitzung eines Players (eigene `session_id`) |
| Rebuffering | Wiedergabe-Unterbrechung, weil der Puffer leerläuft |
| Manifest | Beschreibungsdatei eines Streams (HLS: `.m3u8`, DASH: `.mpd`) |
| HLS / DASH | HTTP-basierte Streaming-Protokolle (Apple bzw. MPEG) |
| SRT | Transportprotokoll für Zubringer-Streams (Contribution) |
| WebRTC/WHEP | Latenzarmes Browser-Streaming; WHEP = Wiedergabe-Handshake |
| CMAF | Gemeinsames Segmentformat für HLS/DASH; m-trace prüft Konformität |
| Ingest | Zubringer-Seite: Streams, Keys, Endpunkte (`/api/ingest/...`) |
| Rate-Limit | Ereignis-Budget je Projekt/Client-IP/Origin (Token-Bucket) |

## 10. Support und Kontakt

- Fragen und Fehlerberichte: GitHub-Issues im Repository
  (<https://github.com/pt9912/m-trace/issues>).
- Sicherheitsrelevante Funde: bitte gemäß
  [`SECURITY.md`](../../SECURITY.md) melden, nicht als öffentliches Issue.
- Mitwirken: [`CONTRIBUTING.md`](../../CONTRIBUTING.md).

## 11. Änderungshistorie

| Handbuch-Version | Software-Version | Datum | Änderung |
|---|---|---|---|
| 1.0 | 0.25.0 | 2026-07-13 | Erstausgabe (aufgabenbasiert nach [`benutzerhandbuch-standard.md`](benutzerhandbuch-standard.md)) |
| 1.1 | 0.25.0 | 2026-07-13 | Review-Korrekturen: Live-Update gilt für die Sessions-Liste (Detail = manueller Refresh); `Retry-After` nur beim Event-Ingest-429; Seeding-Warnung präzisiert (nur `MTRACE_LAB_PROJECTS`); Node 22 statt „22+"; `/api/analyze`-Auth bei Session-Bindung; ENV-Referenz-Wortlaut (`.env.example` um Rate-Limit-/Redis-/Postgres-Block ergänzt) |
