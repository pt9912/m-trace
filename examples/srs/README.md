# SRS-Beispiel — Multi-Protocol Lab

> **Variante**: Eigenes Compose. Project-Name `mtrace-srs`, Compose-
> Datei `examples/srs/compose.yaml` (siehe
> [`examples/README.md`](../README.md) Sektion „Compose-Form"
> Variante „Eigenes Compose").
>
> Bezug: Lastenheft §12.3 MVP-36 (SRS-Beispiel, „Kann"); RAK-57 aus
> Lastenheft §13.11; [`plan-0.9.0.md`](../../docs/planning/done/plan-0.9.0.md)
> §3 (Tranche 2). MVP-36 wird mit diesem Beispiel als eingelöst
> betrachtet, ohne die MVP-Priorität zu ändern.
>
> Quickref aller Multi-Protocol-Lab-Beispiele:
> [`docs/user/local-development.md`](../../docs/user/local-development.md)
> §2.7.

## Zweck

Reproduzierbarer SRS-Pfad: ein FFmpeg-Container publishet einen
synthetischen Test-Stream per RTMP in einen
[SRS](https://github.com/ossrs/srs)-Container, der ihn über die
HTTP-API auflistet und über HTTP-FLV wieder ausspielt. Damit ist
sowohl RTMP-Ingress als auch HTTP-FLV-Egress am selben Server-Setup
verifizierbar.

**Wichtig**: `0.9.0`-SRS bedeutet Beispiel/Smoke. Es bedeutet
**nicht** SRS-Telemetrie in `apps/api`, keinen `mtrace_srs_*`-
Metriken-Import und keinen Lastenheft-Wire-Vertrag. Der Pfad ist
streng analog zu `examples/srt/`/`examples/dash/`/`examples/webrtc/`
— eigenständiges Compose, opt-in Smoke, kein Player-SDK-Hookup.

## Voraussetzungen

- Docker Engine ≥ 24.0, Docker Compose v2.20.
- Freie Host-Ports: `1935/tcp` (RTMP-Ingest), `8088/tcp` (HTTP-FLV-
  Egress), `1985/tcp` (HTTP-API für Operator-Inspektion und Smoke-
  Probe). Die Ports sind absichtlich gewählt, damit das Beispiel
  parallel zum Core-Lab (`8888`/`9997`) und zu den anderen
  Multi-Protocol-Stacks (`mtrace-srt` `8889`/`8890`/`9998`,
  `mtrace-dash` `8891`, `mtrace-webrtc` `8892`/`8189`/`9999`)
  laufen kann.
- Aufruf aus dem Repo-Root.

## Start

```bash
docker compose -p mtrace-srs -f examples/srs/compose.yaml up -d --build
```

Project-Name `mtrace-srs` ist Pflicht (siehe
[`examples/README.md`](../README.md) Sektion „Project-Name-Pflicht
für eigenes Compose"). Damit kollidieren Container, Netzwerk und
Volumes nicht mit `mtrace`-Default des Core-Labs oder anderen
Multi-Protocol-Stacks.

Nach ca. 5–15 s hat FFmpeg den ersten RTMP-Connect mit SRS
ausgehandelt; der Stream `live/srs-test` ist dann in der
HTTP-API sichtbar.

## Verifikation

Opt-in Smoke (RAK-57, Kann; nicht in `make gates`):

```bash
make smoke-srs
```

Smoke-Skript: [`scripts/smoke-srs.sh`](../../scripts/smoke-srs.sh).
Verifiziert mit bounded Wait + Diagnose:

1. **HTTP-API antwortet `200`** auf
   `http://localhost:1985/api/v1/streams/`.
2. **`streams[]` enthält** mindestens einen Eintrag mit
   `app=live`, `name=srs-test` und `publish.active=true`.
3. **HTTP-FLV-Egress liefert** unter
   `http://localhost:8088/live/srs-test.flv` einen `200`-Response;
   ein paar Bytes vom Anfang des FLV-Streams werden gelesen
   (FLV-Header `FLV` als Magic), um sicherzustellen, dass HTTP-FLV-
   Remux nicht nur den Header hält.

### Manueller Aufruf ohne Auto-Start

`make smoke-srs` startet den `mtrace-srs`-Stack selbst
(`up -d --build`) und beendet ihn nach Abschluss (`down`). Für
manuelle Verifikation ohne Auto-Start:

```bash
docker compose -p mtrace-srs -f examples/srs/compose.yaml up -d --build
SMOKE_SRS_AUTOSTART=0 bash scripts/smoke-srs.sh
docker compose -p mtrace-srs -f examples/srs/compose.yaml down
```

Manueller Schnell-Check ohne Smoke:

```bash
# Stream-Liste direkt
curl -sS http://localhost:1985/api/v1/streams/ | python3 -m json.tool

# HTTP-FLV-Egress (Magic-Header `FLV`)
curl -sS -o /tmp/srs.flv -w '%{http_code}\n' --max-time 3 \
  http://localhost:8088/live/srs-test.flv || true
head -c 3 /tmp/srs.flv

# Publisher-Logs (FFmpeg)
docker compose -p mtrace-srs logs srs-publisher | tail -20

# Server-Logs (SRS)
docker compose -p mtrace-srs logs srs | tail -20
```

## Stop / Reset

```bash
docker compose -p mtrace-srs -f examples/srs/compose.yaml down
```

Greift nur den `mtrace-srs`-Project-Namen — Core-Lab und andere
Beispiele bleiben unangetastet. Der Stack hat keine persistenten
Volumes; `down` reicht für vollständigen Reset.

## Troubleshooting

- **HTTP-API `connection refused` / Smoke-Timeout**: SRS braucht
  typisch 2–5 s zum Start des HTTP-API-Listeners. `WAIT_SECONDS` in
  `make smoke-srs` ist auf 30 s gesetzt; bei langsameren Maschinen
  kann der Wert gehoben werden: `WAIT_SECONDS=60 make smoke-srs`.
- **`1935`/`1985`/`8088` Port belegt**: ein anderer Prozess (oder
  ein früherer `mtrace-srs`-Stack) blockiert die Ports. `docker
  compose -p mtrace-srs down` beenden, dann nachschauen mit
  `ss -ltnp` für TCP-Konflikte.
- **`srs-publisher`-Container loggt FFmpeg-Reconnect-Schleife**:
  das `ffmpeg-rtmp-loop.sh`-Skript versucht alle 2 s neu zu
  verbinden, falls SRS noch nicht bereit ist. Bei dauerhaftem
  Reconnect `docker compose -p mtrace-srs logs srs` prüfen — der
  SRS-Container muss den RTMP-Listener auf `:1935/tcp` geöffnet
  haben.
- **HTTP-FLV liefert `404`**: SRS hat den Stream noch nicht
  registriert (FFmpeg-Connect war noch nicht erfolgreich) oder die
  `http_remux`-Sektion in `srs.conf` fehlt. Smoke-Probe gegen
  `http://localhost:1985/api/v1/streams/` zeigt, ob `srs-test`
  existiert; wenn `streams[]` leer ist, ist FFmpeg das Problem.

## Bekannte Grenzen

- **Kein produktiver Telemetriepfad**: kein `mtrace_srs_*`-Counter,
  kein API-Hookup, keine Dashboard-Sichtbarkeit. Der Smoke
  validiert die Quelle (SRS HTTP-API + FLV-Egress), nicht einen
  `apps/api`-Read-Pfad. Eine SRS-Health-View analog `0.6.0`
  SRT-Health bleibt out of scope für `0.9.0` und ist Folge-Plan-
  Material.
- **Kein Player-SDK-/Dashboard-Integrations-Test**: der Demo-Player
  im Core-Lab nutzt weiterhin den `teststream` aus dem Core-Compose,
  nicht `srs-test` aus diesem Beispiel.
- **Kein Multi-Publisher- oder Authentifizierungs-Setup**: das
  Beispiel zeigt einen einzelnen `live/srs-test`-Pfad. Mehrere
  Publisher, RTMP-Auth oder Edge-Konfigurationen sind out of scope.
- **Kein HLS-/DASH-/WebRTC-Output**: SRS unterstützt diese Formate,
  aber dafür gibt es eigene Beispiele (`mtrace-srt` für HLS aus
  SRT-Ingest, `mtrace-dash` für DASH-Live, `mtrace-webrtc` für
  WHIP/WHEP). SRS-Lab fokussiert RTMP-Ingress + FLV-Egress als
  Operator-sichtbaren Verifikationspfad.
