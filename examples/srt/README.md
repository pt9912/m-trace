# SRT-Beispiel — Multi-Protocol Lab

> **Variante**: Eigenes Compose. Project-Name `mtrace-srt`, Compose-
> Datei `examples/srt/compose.yaml` (siehe
> [`examples/README.md`](../README.md) Sektion „Compose-Form"
> Variante „Eigenes Compose").
>
> Bezug: Lastenheft §7.8 F-82..F-84, RAK-37 (`0.5.0`-Smoke);
> [`plan-0.5.0.md`](../../docs/planning/done/plan-0.5.0.md) §4
> (Tranche 3); für die SRT-Health-Erweiterung
> [`plan-0.6.0.md`](../../docs/planning/in-progress/plan-0.6.0.md)
> §3 (Tranche 2) plus Lastenheft §13.8 RAK-41..RAK-46.
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
(Risiken-Backlog R-2 in `0.6.0` Tranche 1 als CGO-frei aufgelöst —
Quelle ist die MediaMTX-Control-API über HTTP). Das `apps/api`-Image
bleibt `distroless-static` ohne CGO.

Ab `0.6.0` Tranche 2 erweitert dieser Stack das Lab um einen
**Health-Smoke** (`make smoke-srt-health`), der zusätzlich gegen
MediaMTX `/v3/srtconns/list` probt und die vier RAK-43-Pflichtwerte
(RTT, Packet Loss, Retransmissions, verfügbare Bandbreite) numerisch
verifiziert. Smoke-Baseline `make smoke-srt` (HLS-Pfad) bleibt
unverändert grün und ist weiterhin der `0.5.0`-RAK-37-Nachweis.

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

Zwei opt-in Smokes mit unterschiedlicher Tiefe:

### `make smoke-srt` (Baseline, RAK-37)

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

### `make smoke-srt-health` (Health-Erweiterung, plan-0.6.0 Tranche 2)

```bash
make smoke-srt-health
```

Smoke-Skript:
[`scripts/smoke-srt-health.sh`](../../scripts/smoke-srt-health.sh).
Erweitert die Baseline um eine API-Probe:

1. HLS-Manifest erreichbar (wie oben).
2. **MediaMTX-API antwortet `200`** auf
   `http://localhost:9998/v3/srtconns/list`.
3. **`items[]` enthält** mindestens einen Eintrag mit
   `path=srt-test` und `state=publish`.
4. **Vier RAK-43-Pflichtwerte numerisch gesetzt**: `msRTT`,
   `packetsReceivedLoss`, `packetsReceivedRetrans` (alle ≥ 0;
   gesundes Lab liefert 0), `mbpsLinkCapacity > 0` (sonst keine
   Bandbreiten-Schätzung verfügbar).

Datenfluss (Health-Pfad):
`srt-publisher (FFmpeg)` → `mediamtx :8890/udp` (SRT-Listener)
→ `mediamtx /v3/srtconns/list :9997 → host :9998` (Control-API)
→ `scripts/smoke-srt-health.sh` (curl + python3-Validierung).

Zusätzliche Dependency: `python3` (JSON-Schema-Check). Sonst
identisch zur Baseline.

### Manueller Aufruf ohne Auto-Start

`make smoke-srt` und `make smoke-srt-health` starten den
`mtrace-srt`-Stack selbst (`up -d --build`) und beenden ihn nach
Abschluss (`down`). Für manuelle Verifikation ohne Auto-Start:

```bash
docker compose -p mtrace-srt -f examples/srt/compose.yaml up -d --build
SMOKE_SRT_AUTOSTART=0 bash scripts/smoke-srt.sh         # Baseline
SMOKE_SRT_AUTOSTART=0 bash scripts/smoke-srt-health.sh  # Health
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
  ab MediaMTX 1.14+ Default. `examples/srt/mediamtx.yml` enthält ab
  `0.6.0` Tranche 2 einen `authInternalUsers`-Block, der
  `api`/`metrics`/`publish`/`read` für `any` mit leerem Passwort
  freischaltet — ausreichend für Lab. Falls die Datei lokal
  überschrieben oder älter ist, den Block ergänzen oder
  `make smoke-srt-health` zeigt das in der Diagnose explizit
  („MediaMTX-API unreachable… braucht authInternalUsers mit
  action=api").
- **`make smoke-srt-health` schlägt mit „missing dependency:
  python3" fehl**: der Health-Smoke nutzt python3 für die JSON-
  Schema-Validierung. Die Baseline `make smoke-srt` braucht
  python3 nicht.

## Bekannte Grenzen

- Keine SRT-Health-Metriken in `apps/api` — kein Metrikimport, keine
  Aggregat-Counter, keine Dashboard-Sichtbarkeit. Der Health-Smoke
  aus `make smoke-srt-health` validiert die **Quelle** (MediaMTX-
  API), nicht den `apps/api`-Read-Pfad. Das kommt in
  `plan-0.6.0` Tranche 3+ (RAK-42..RAK-46).
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
