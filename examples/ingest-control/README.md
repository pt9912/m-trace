# Ingest-Control-Beispiel — Multi-Protocol Lab

> **Variante**: Eigenes Compose. Project-Name `mtrace-ingest-control`,
> Compose-Datei `examples/ingest-control/compose.yaml` (siehe
> [`examples/README.md`](../README.md) Sektion „Compose-Form"
> Variante „Eigenes Compose").
>
> Bezug: Lastenheft §7.5.4 `F-46`..`F-51`, §13.13 `RAK-65`..`RAK-70`
> (Patch `1.1.13`); [`plan-0.11.0.md`](../../docs/planning/in-progress/plan-0.11.0.md)
> Tranche 3.
>
> Quickref aller Multi-Protocol-Lab-Beispiele:
> [`docs/user/local-development.md`](../../docs/user/local-development.md)
> §2.7.

## Zweck

Reproduzierbarer **Stream-Control-Pfad**: ein lokal laufender
MediaMTX-Container empfängt SRT-/RTMP-Publisher und stellt das
generierte Manifest als HLS bereit. Anders als das `examples/srt/`-
Beispiel demonstriert dieser Stack die **Stream-Catalog-Sicht** aus
der `apps/api`-Ingest-Control-Domain (`POST /api/ingest/streams`,
`GET /api/ingest/media-server-config`) — die `mediamtx.generated.yml`
in diesem Verzeichnis ist genau die Form, die `apps/api` per
deterministischer YAML-Generierung produziert.

**Wichtig**: `0.11.0`-Ingest-Control ist Lab-/Demo-Pfad, **nicht**
produktive Ingest-Control-Plane. Es bedeutet:

- **kein** mandantenfähiger SaaS-Pfad (Tenant-Auth → `0.12.0`),
- **kein** produktiver Media-Server-Auth-Hook (`validate-key` ist
  diagnostisch, nicht ein Auth-Replacement; siehe `risks-backlog.md`
  R-14),
- **keine** automatische Provisionierung externer MediaMTX-Server
  (`media-server-config` schreibt nur lokale Artefakte; R-15),
- **keine** produktive ausgehende Webhook-Zustellung (R-16).

Produktionsdeployments setzen Auth-Plugins, IP-Allowlists und/oder
HMAC-signierte Stream-Keys auf MediaMTX-Seite ein — der vorliegende
Compose-Stack tut das **nicht**.

## Kurzanleitung

```bash
# Lab starten (eigene Project-ID, parallel zum Core-Lab nutzbar):
docker compose -p mtrace-ingest-control \
  -f examples/ingest-control/compose.yaml up -d --build

# SRT-Publisher gegen den lokalen MediaMTX-Listener:
ffmpeg -re -f lavfi -i testsrc=size=320x240:rate=10 \
  -f mpegts "srt://localhost:8891?streamid=publish:lab-srt&pkt_size=1316"

# RTMP-Publisher analog:
ffmpeg -re -f lavfi -i testsrc=size=320x240:rate=10 \
  -c:v libx264 -tune zerolatency -f flv \
  "rtmp://localhost:1936/lab-rtmp"

# HLS-Auslieferung prüfen:
curl http://localhost:8892/lab-srt/index.m3u8

# Lab beenden:
docker compose -p mtrace-ingest-control \
  -f examples/ingest-control/compose.yaml down -v
```

## Port-Verteilung

Ports auf dem Host sind verschoben, damit das Beispiel parallel zu
`make dev` (Core-Lab MediaMTX auf 8888/9997) und parallel zu
`examples/srt/` (8889/8890/9998) startbar bleibt:

| Container-Port | Host-Port | Zweck                                     |
| -------------- | --------- | ----------------------------------------- |
| 8888 (TCP)     | 8892      | MediaMTX HLS-Egress                       |
| 8890 (UDP)     | 8891      | MediaMTX SRT-Listener (Ingest)            |
| 1935 (TCP)     | 1936      | MediaMTX RTMP-Listener (Ingest)           |
| 9997 (TCP)     | 9999      | MediaMTX Control-API                      |

## Generierte vs. eingecheckte Konfiguration

`mediamtx.generated.yml` in diesem Verzeichnis ist die **deterministische
Referenz-Form**, die `apps/api` über
`GET /api/ingest/media-server-config?target_id=mediamtx-local`
ausspielt — Header-Comment, `srt`/`rtmp`/`hls`-Toggles, `paths:`-
Block sortiert nach `display_name`.

In Produktion liest MediaMTX diese Datei direkt; in diesem Lab-Stack
mountet `compose.yaml` sie read-only als
`/mediamtx.yml`.

## Wartung

- **Verifikation der Generator-Form**: `make api-test` läuft
  `application.GenerateMediaMTXConfig`-Tests, die deterministischen
  Output und Klartext-Key-Schutz pinnen.
- **Smoke**: `make smoke-ingest-control` (opt-in, NICHT Teil von
  `make gates`) verifiziert den Lifecycle-Hook-Pfad reproduzierbar:
  Stream anlegen → `stream-started` einspeisen → `stream-ended`
  einspeisen, jeweils `202 accepted:true` und unterschiedliche
  `event_id`. Default-API-URL `http://localhost:8080`,
  Default-Token `demo-token`; beides via Env-Vars überschreibbar.
  Das Script `smoke-lifecycle.sh` lebt direkt in diesem Verzeichnis.
- **Compose-Konvention**: Project-Name `mtrace-ingest-control`,
  eigene Volumes, Host-Port-Schnitt kollisionsfrei zu Core-Lab.
- **Bestehende Beispiele bleiben unangetastet**: dieses Verzeichnis
  ist additiv; `examples/srt/`, `examples/mediamtx/`,
  `examples/srs/` sind nicht betroffen.

## Was dieses Beispiel **nicht** ist

- **Kein** produktiver Stream-Control-Pfad. Die Lab-Konfiguration hat
  keine Auth, keine IP-Allowlist, kein Replay-Schutz.
- **Kein** Performance-Test. FFmpeg-`testsrc` ist synthetisch und
  steht für funktionalen Smoke, nicht für Lastmessung.
- **Kein** Webhook-Endpoint nach außen. Eingehende Lifecycle-Events
  (`/api/ingest/hooks/stream-{started,ended}`) sind in `0.11.0`
  Tranche 4 ausgeliefert und lassen sich aus diesem Lab heraus per
  `make smoke-ingest-control` an `apps/api` einspeisen — ausgehende
  produktive Webhook-Zustellung an externe Systeme bleibt
  Folge-Scope.
