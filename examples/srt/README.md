# SRT-Beispiel — Multi-Protocol Lab

> **Status**: Skelett. Inhalt liefert
> [`plan-0.5.0.md`](../../docs/planning/in-progress/plan-0.5.0.md) §4
> (Tranche 3, RAK-37).

## Zweck

Zeigt einen reproduzierbaren SRT-Pfad: SRT-Publish in einen Lab-
Container, lokale Wiedergabe. Bezug: Lastenheft §7.8 F-82..F-84,
RAK-37.

**Wichtig**: `0.5.0`-SRT bedeutet Beispiel/Smoke. Es bedeutet
**nicht** SRT-Health-View (RAK-Folge), keinen SRT-Metrikimport in
`apps/api` und keine CGO-Bindings (Risiken-Backlog R-2). Diese
Themen sind explizit Folge-Scope (`0.6.0`).

## Voraussetzungen

_Liefert Tranche 3._ Bisher offen: Docker Engine ≥ 24.0, Compose
v2.20, freier UDP-Port für SRT-Listener, FFmpeg auf der
Operator-Seite zum Test-Publish.

## Start

```bash
docker compose -p mtrace-srt -f examples/srt/compose.yaml up -d --build
```

Project-Name `mtrace-srt` ist Pflicht (siehe
[`examples/README.md`](../README.md)).

_Compose-Datei und Detail-Konfiguration liefert Tranche 3._

## Verifikation

```bash
make smoke-srt
```

_Smoke-Skript liefert Tranche 3._ Pfad: Test-Publisher (FFmpeg)
schickt einen kurzen Stream, Smoke verifiziert, dass der SRT-
Container den Stream empfängt und über die konfigurierte Endpunkt-
URL ausspielt.

## Stop / Reset

```bash
docker compose -p mtrace-srt -f examples/srt/compose.yaml down
```

Hard-Reset analog `examples/mediamtx/`: `down --volumes` für vollen
Reset. Greift nur den `mtrace-srt`-Project-Namen — Core-Lab und andere
Beispiele bleiben unangetastet.

## Troubleshooting

_Liefert Tranche 3._ Erwartete Einträge: UDP-Port-Konflikt, MTU-
Fragmentation, FFmpeg-/SRT-Library-Versionsdifferenzen.

## Bekannte Grenzen

- Keine SRT-Health-Metriken in `apps/api` (Folge-Scope `0.6.0`,
  Risiken-Backlog R-2).
- Kein CGO-Binding in `apps/api`: das Lab nutzt SRT ausschließlich
  innerhalb des Compose-Containers; das `apps/api`-`distroless-static`
  Image-Pattern bleibt unverändert.
- Kein Multi-Publisher- oder Authentifizierungs-Setup; `0.5.0` zeigt
  nur den minimalen Lab-Pfad.
