# MediaMTX-Beispiel — Multi-Protocol Lab

> **Variante**: Core-Lab-Beispiel — MediaMTX läuft bereits als
> Pflicht-Service im Root-`docker-compose.yml`. Dieses Beispiel
> dokumentiert den bestehenden Pfad und ergänzt nur den opt-in Smoke
> `make smoke-mediamtx`.
>
> Bezug: Lastenheft §7.8 F-82..F-84, RAK-36;
> [`plan-0.5.0.md`](../../docs/planning/done/plan-0.5.0.md) §3
> (Tranche 2);
> [`services/media-server/mediamtx.yml`](../../services/media-server/mediamtx.yml).
>
> Quickref aller Multi-Protocol-Lab-Beispiele:
> [`docs/user/local-development.md`](../../docs/user/local-development.md)
> §2.7.

## Zweck

Zeigt den bestehenden MediaMTX-Pfad als nachvollziehbares Beispiel:
RTSP-Ingest des FFmpeg-Teststreams, HLS-Ausspielung, MediaMTX-API/
Status-Endpoint. Kein separater Stack — das Core-Lab (`make dev`)
liefert den vollständigen Pfad bereits.

## Voraussetzungen

- Docker Engine ≥ 24.0, Docker Compose v2.20.
- Freie Ports: `8554` (RTSP), `8888` (HLS), `9997` (MediaMTX-API).
- Aufruf aus dem Repo-Root.

## Start

```bash
make dev
```

Startet das Core-Lab. MediaMTX läuft direkt aus
[`docker-compose.yml`](../../docker-compose.yml) mit der Konfiguration
[`services/media-server/mediamtx.yml`](../../services/media-server/mediamtx.yml).
Der `stream-generator`-Service publishedet einen FFmpeg-Loop-Stream
unter dem Pfad `teststream`.

Kein eigener `-p`-Flag nötig — Project-Name bleibt der Core-Lab-
Default `mtrace` (siehe [`examples/README.md`](../README.md) Sektion
„Compose-Form", Variante „Core-Lab-Beispiel").

## Verifikation

```bash
make smoke-mediamtx
```

Smoke-Skript: [`scripts/smoke-mediamtx.sh`](../../scripts/smoke-mediamtx.sh).
Verifiziert mit bounded Waits + Diagnose:

1. **HLS-Manifest erreichbar** — `GET http://localhost:8888/teststream/index.m3u8`
   liefert `200 OK` (FFmpeg-Publisher hat den Stream registriert,
   MediaMTX liefert das Manifest aus).
2. **Manifest-Body ist sinnvoll** — Body beginnt mit `#EXTM3U` und
   enthält mindestens eine Segment-Referenz (`.ts`/`.m4s`/`.aac`),
   d. h. der teststream publisht echte Segmente — eine leere/initiale
   Playlist reicht nicht.

Funktional ist HLS-200 + sinnvolles Manifest die Zielmetrik des
Beispiels: wenn HLS ausspielt, ist sowohl MediaMTX als auch der
FFmpeg-Publisher erreichbar. Die Control-API auf `:9997` ist ab
MediaMTX 1.14+ standardmäßig Auth-pflichtig und daher absichtlich
**nicht** Teil des Smoke-Pfads — sie bleibt eine optionale Operator-
Diagnose (siehe Troubleshooting).

Bei Timeout schreibt der Smoke einen aussagekräftigen Diagnose-Block
auf stderr (letzte HLS-Antwort plus Hinweise auf
`docker compose logs stream-generator|mediamtx`). Default-Wartezeit:
30 s, überschreibbar via `WAIT_SECONDS`-Env-Var.

Manuelle Schnellprüfung ohne Smoke-Skript:

```bash
# HLS-Manifest direkt
curl -L http://localhost:8888/teststream/index.m3u8

# Optional, MediaMTX-Logs
docker compose logs mediamtx | tail -20
docker compose logs stream-generator | tail -20
```

## Stop / Reset

```bash
make stop                  # docker compose down ohne --volumes
make wipe                  # Hard-Reset, entfernt mtrace-data und mtrace-tempo-data
```

`make stop` lässt das `mtrace-data`-Volume unangetastet und beendet
nur die Core-Lab-Services.

## Troubleshooting

- **HLS-Manifest 404 bzw. Smoke-Timeout**: FFmpeg-Publisher ist noch
  nicht fertig hochgefahren oder hat sich nicht gegen MediaMTX
  verbunden. Smoke-Skript wartet bounded (30 s, via `WAIT_SECONDS`
  überschreibbar); manuell prüfen:
  `docker compose logs stream-generator | tail -20` (Publisher-Log),
  `docker compose logs mediamtx | tail -20` (Server-Log).
- **`8888`/`9997` Port belegt**: Core-Lab kollidiert mit lokalem
  Service. Lokal stoppen oder per `docker-compose.override.yml`
  alternative Ports binden.
- **HLS-Manifest 200, aber Player liefert keine Frames**: MediaMTX-
  Konfiguration prüfen (`services/media-server/mediamtx.yml`); `paths.teststream.source: publisher` muss aktiv sein.
- **MediaMTX-API auf `:9997` antwortet `401 Unauthorized`**: das ist
  Default ab MediaMTX 1.14 (Control-API ist Auth-pflichtig). Für
  reine Operator-Diagnose im Lab können API-Auth-Credentials in
  `services/media-server/mediamtx.yml` (Felder `authInternalUsers`)
  gesetzt werden — der Smoke-Pfad braucht das nicht.

## Bekannte Grenzen

- Dieses Beispiel zeigt nur den bestehenden Core-Lab-Pfad. Eigene
  Encoder-Konfigurationen, Multi-Bitrate-Ladders oder TLS-
  Terminierung sind out of scope.
- SRT-Health-Metriken sind kein MediaMTX-Beispiel-Scope; siehe
  [`examples/srt/`](../srt/) und Folgepfad `0.6.0`.
- Die `analyzer-service`-Anbindung über `POST /api/analyze` und der
  Dashboard-Demo-Player nutzen weiterhin
  `http://mediamtx:8888/teststream/index.m3u8` aus
  [`docker-compose.yml`](../../docker-compose.yml) (`PUBLIC_HLS_URL`-
  Env-Var); dieses Beispiel ändert daran nichts.
