# MediaMTX-Beispiel — Multi-Protocol Lab

> **Status**: Skelett. Inhalt liefert
> [`plan-0.5.0.md`](../../docs/planning/in-progress/plan-0.5.0.md) §3
> (Tranche 2, RAK-36).

## Zweck

Zeigt einen reproduzierbaren MediaMTX-Pfad: RTSP/RTMP-Ingest, HLS-
Ausspielung, Status-API. Bezug: Lastenheft §7.8 F-82..F-84,
RAK-36.

## Voraussetzungen

_Liefert Tranche 2._ Bisher offen: Docker Engine ≥ 24.0, Docker
Compose v2.20, freie Ports analog Core-Lab (`8888` HLS, `9997`
Status, plus RTSP/RTMP-Eingangsports).

## Start

```bash
docker compose -p mtrace-mediamtx -f examples/mediamtx/compose.yaml up -d --build
```

Project-Name `mtrace-mediamtx` ist Pflicht (siehe
[`examples/README.md`](../README.md) — Project-Name-Konvention).

_Compose-Datei und Detail-Konfiguration liefert Tranche 2._

## Verifikation

```bash
make smoke-mediamtx
```

_Smoke-Skript liefert Tranche 2._ Erwartete Endpoints und ihre
Status-Codes werden dort dokumentiert.

## Stop / Reset

```bash
docker compose -p mtrace-mediamtx -f examples/mediamtx/compose.yaml down
```

Volumes bleiben standardmäßig erhalten (Lab-Defaults). Hard-Reset:
`docker compose -p mtrace-mediamtx ... down --volumes` — anschließende
`up`-Befehle bauen Container/Volumes neu.

## Troubleshooting

_Liefert Tranche 2._ Erwartete Einträge: Port-Kollision mit Core-Lab,
fehlende Codec-Unterstützung im Player, Encoder-Latenz.

## Bekannte Grenzen

- `0.5.0` liefert nur das Lab-Beispiel; produktive Ingest-Verwaltung,
  Multi-Tenant-Setup oder TLS-Terminierung sind out of scope.
- SRT-Health-Metriken sind kein MediaMTX-Beispiel-Scope; siehe
  [`examples/srt/`](../srt/) und Folgepfad `0.6.0`.
