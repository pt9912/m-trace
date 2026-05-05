# SRT-Beispiel — Multi-Protocol Lab

> **Variante**: Eigenes Compose. Project-Name `mtrace-srt`, Compose-
> Datei `examples/srt/compose.yaml` (siehe
> [`examples/README.md`](../README.md) Sektion „Compose-Form"
> Variante „Eigenes Compose").
>
> Bezug: Lastenheft §7.8 F-82..F-84, RAK-37;
> [`plan-0.5.0.md`](../../docs/planning/in-progress/plan-0.5.0.md) §4
> (Tranche 3).
>
> Quickref aller Multi-Protocol-Lab-Beispiele:
> [`docs/user/local-development.md`](../../docs/user/local-development.md)
> §2.7.

## Zweck

Reproduzierbarer SRT-Pfad: ein FFmpeg-Container publishet einen
synthetischen Test-Stream per SRT in einen MediaMTX-Container, der
ihn als HLS wieder ausspielt. Damit ist sowohl SRT-Ingress als auch
HLS-Egress am selben Server-Setup verifizierbar.

**Wichtig**: `0.5.0`-SRT bedeutet Beispiel/Smoke. Es bedeutet
**nicht** SRT-Health-View (RAK-41..RAK-46, Folge-Scope `0.6.0`),
keinen SRT-Metrikimport in `apps/api` und keine CGO-Bindings
(Risiken-Backlog R-2). Das `apps/api`-Image bleibt
`distroless-static` ohne CGO.

## Voraussetzungen

- Docker Engine ≥ 24.0, Docker Compose v2.20.
- Freie Host-Ports: `8889/tcp` (HLS-Out), `8890/udp` (SRT-Listener),
  `9998/tcp` (MediaMTX-Control-API, Diagnose). Die Ports sind
  absichtlich gegen das Core-Lab (`8888`/`9997`) verschoben, damit
  beide Stacks parallel laufen können.
- Aufruf aus dem Repo-Root.

## Start

```bash
docker compose -p mtrace-srt -f examples/srt/compose.yaml up -d --build
```

Project-Name `mtrace-srt` ist Pflicht (siehe
[`examples/README.md`](../README.md) Sektion „Project-Name-Pflicht
für eigenes Compose"). Damit kollidieren Container, Netzwerk und
Volumes nicht mit `mtrace`-Default des Core-Labs.

Nach ca. 10–25 s hat FFmpeg den ersten SRT-Connection-Setup
abgeschlossen und MediaMTX hat das erste HLS-Segment fertig.

## Verifikation

```bash
make smoke-srt
```

Smoke-Skript: [`scripts/smoke-srt.sh`](../../scripts/smoke-srt.sh).
Verifiziert mit bounded Wait + Diagnose:

1. **HLS-Manifest erreichbar** unter `http://localhost:8889/srt-test/index.m3u8`
   (HTTP 200 mit bounded Polling, Default 45 s).
2. **Manifest-Body ist sinnvoll** — Body beginnt mit `#EXTM3U` und
   enthält Media-Referenzen (`.m3u8`-Substreams oder
   `.ts`/`.m4s`/`.aac`-Segmente). Beides bedeutet: SRT-Publisher
   verbunden, MediaMTX servisiert echte Inhalte.

`make smoke-srt` startet den `mtrace-srt`-Stack selbst (`up -d --build`)
und beendet ihn nach Abschluss (`down`) — Operator muss kein Compose
händisch hochfahren. Für manuelle Verifikation ohne Auto-Start:

```bash
docker compose -p mtrace-srt -f examples/srt/compose.yaml up -d --build
SMOKE_SRT_AUTOSTART=0 bash scripts/smoke-srt.sh
docker compose -p mtrace-srt -f examples/srt/compose.yaml down
```

Manueller Schnell-Check ohne Smoke:

```bash
# HLS-Manifest direkt
curl -L http://localhost:8889/srt-test/index.m3u8

# SRT-Publisher-Logs (FFmpeg)
docker compose -p mtrace-srt logs srt-publisher | tail -20

# MediaMTX-Logs (Server)
docker compose -p mtrace-srt logs mediamtx | tail -20
```

## Stop / Reset

```bash
docker compose -p mtrace-srt -f examples/srt/compose.yaml down
```

Greift nur den `mtrace-srt`-Project-Namen — Core-Lab und andere
Beispiele bleiben unangetastet. Der Stack hat keine persistenten
Volumes; `down` reicht für vollständigen Reset.

## Troubleshooting

- **HLS-Manifest 404 / Smoke-Timeout**: SRT-Publisher braucht typisch
  10–25 s, bevor MediaMTX das erste HLS-Segment serviert. `WAIT_SECONDS`
  in `make smoke-srt` ist auf 45 s gesetzt; bei langsameren Maschinen
  kann der Wert gehoben werden:
  `WAIT_SECONDS=90 make smoke-srt`.
- **`8889`/`8890`/`9998` Port belegt**: ein anderer Prozess (oder ein
  früherer `mtrace-srt`-Stack) blockiert die Ports. `docker compose
  -p mtrace-srt down` beenden, dann nachschauen mit
  `ss -ltnp` / `ss -lunp` für TCP/UDP-Konflikte.
- **`srt-publisher`-Container loggt FFmpeg-Reconnect-Schleife**: das
  `ffmpeg-srt-loop.sh`-Script versucht alle 2 s neu zu verbinden, falls
  MediaMTX noch nicht bereit ist. Bei dauerhaftem Reconnect
  `docker compose -p mtrace-srt logs mediamtx` prüfen — der MediaMTX-
  Container muss den SRT-Listener auf `:8890/udp` geöffnet haben.
- **MediaMTX-Control-API auf `:9998` antwortet `401 Unauthorized`**:
  ab MediaMTX 1.14+ Default. Für lokale Operator-Diagnose
  `authInternalUsers` in `examples/srt/mediamtx.yml` ergänzen — der
  Smoke-Pfad braucht das nicht (HLS-Pfad ist Auth-frei).

## Bekannte Grenzen

- Keine SRT-Health-Metriken in `apps/api` — kein Metrikimport, keine
  Aggregat-Counter, keine Dashboard-Sichtbarkeit. Das ist
  Folge-Scope `0.6.0` (RAK-41..RAK-46) und Risiken-Backlog R-2.
- Kein CGO-Binding in `apps/api`: das Lab nutzt SRT ausschließlich
  innerhalb von Container (MediaMTX hat eigene SRT-Implementation).
  Das `apps/api`-`distroless-static`-Pattern bleibt unverändert; der
  Image-Größen-/Cold-Start-Vorteil aus ADR-0001 wird nicht riskiert.
- Kein Multi-Publisher- oder Authentifizierungs-Setup: das Beispiel
  zeigt einen einzelnen `streamid=publish:srt-test`-Pfad. Mehrere
  Publisher, SRT-Passphrase oder PUSH-/PULL-Mode-Tests sind out of
  scope.
- Kein Player-SDK-/Dashboard-Integrations-Test: der Demo-Player im
  Core-Lab nutzt weiterhin den `teststream` aus dem Core-Compose,
  nicht `srt-test` aus diesem Beispiel.
