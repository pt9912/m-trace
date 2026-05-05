# DASH-Beispiel — Multi-Protocol Lab

> **Variante**: Eigenes Compose. Project-Name `mtrace-dash`, Compose-
> Datei `examples/dash/compose.yaml` (siehe
> [`examples/README.md`](../README.md) Sektion „Compose-Form"
> Variante „Eigenes Compose").
>
> Bezug: Lastenheft §7.6 F-58, §7.7 F-73, §7.8 F-82..F-84, §12.3
> MVP-37, RAK-38;
> [`plan-0.5.0.md`](../../docs/planning/in-progress/plan-0.5.0.md) §5
> (Tranche 4).
>
> Quickref aller Multi-Protocol-Lab-Beispiele:
> [`docs/user/local-development.md`](../../docs/user/local-development.md)
> §2.7.

## Zweck

Lokal nachvollziehbarer DASH-Pfad: ein FFmpeg-Container generiert
einen synthetischen DASH-Live-Output (`manifest.mpd` + Init-/Media-
Segmente in CMAF-/fMP4-Form) in ein shared Volume; ein nginx-Container
serviert das Volume statisch auf Host-Port `8891`. Damit ist die
DASH-Auslieferung lokal testbar — ohne externe CDNs oder
Internet-Zugriff.

**Wichtig**: `0.5.0`-DASH liefert nur das Ausspielungsbeispiel. Es
liefert **nicht**:

- vollständige DASH-Manifestanalyse in `@npm9912/stream-analyzer`
  (MVP-37, Kann-Folge-Scope; Plan §0.1 DASH-Zeile),
- einen `dash.js`-Adapter im `@npm9912/player-sdk`,
- eine DASH-Erweiterung von `POST /api/analyze` (bleibt HLS-only;
  Lastenheft §7.6 F-58 Folge-Scope).

## Voraussetzungen

- Docker Engine ≥ 24.0, Docker Compose v2.20.
- Freier Host-Port `8891/tcp` (DASH-HTTP-Server). Der Port ist
  absichtlich gegen Core-Lab (`8888`) und SRT-Beispiel (`8889`/
  `8890`/`9998`) verschoben, damit alle Stacks parallel laufen
  können.
- Aufruf aus dem Repo-Root.

## Start

```bash
docker compose -p mtrace-dash -f examples/dash/compose.yaml up -d --build
```

Project-Name `mtrace-dash` ist Pflicht (siehe
[`examples/README.md`](../README.md) Sektion „Project-Name-Pflicht
für eigenes Compose").

Nach ca. 10–20 s hat FFmpeg das erste Manifest geschrieben und nginx
servisiert es. Der DASH-Live-Window deckt 5 Segmente à 4 s ab (≈ 20 s)
plus 2 Extra-Segmente für Late-Joiner.

## Verifikation

```bash
make smoke-dash
```

Smoke-Skript: [`scripts/smoke-dash.sh`](../../scripts/smoke-dash.sh).
Verifiziert mit bounded Wait + Diagnose:

1. **MPD erreichbar** unter `http://localhost:8891/manifest.mpd`
   (HTTP 200 mit bounded Polling, Default 45 s).
2. **MPD-Body ist sinnvoll** — enthält das `<MPD`-Root-Element
   (gültige DASH-XML).
3. **Init-Segment erreichbar** — `http://localhost:8891/init-stream0.m4s`
   liefert `200 OK` (FFmpeg-Standard-Template-Pfad). Damit ist der
   gesamte Pfad belegt: FFmpeg-Generator schreibt, nginx serviert,
   Player könnte das Stream initialisieren.

`make smoke-dash` startet den `mtrace-dash`-Stack selbst und beendet
ihn nach Abschluss inklusive Volume-Reset (`down --volumes`) — der
DASH-Output-Volume `dash-output` wird beim Cleanup gelöscht, damit
ein erneuter Smoke einen sauberen Neustart hat.

Manuelle Verifikation:

```bash
# MPD direkt
curl -L http://localhost:8891/manifest.mpd

# Init-Segment HEAD-Check
curl -I http://localhost:8891/init-stream0.m4s

# Generator-Logs
docker compose -p mtrace-dash logs dash-generator | tail -20

# nginx-Logs
docker compose -p mtrace-dash logs dash-server | tail -20
```

DASH-Player-Test im Browser (manuell, nicht Teil des Smokes): die
[Shaka-Player-Demo](https://shaka-player-demo.appspot.com/) oder
[dash.js-Reference-Player](https://reference.dashif.org/dash.js/latest/samples/dash-if-reference-player/index.html)
mit MPD-URL `http://localhost:8891/manifest.mpd` lädt typischerweise
nach ein paar Sekunden den Live-Stream (CORS ist in `nginx.conf`
freigegeben).

## Stop / Reset

```bash
docker compose -p mtrace-dash -f examples/dash/compose.yaml down --volumes
```

`--volumes` ist bei Stop empfohlen, weil das DASH-Output-Volume
mit jedem Generator-Neustart sowieso komplett überschrieben wird.
Greift nur den `mtrace-dash`-Project-Namen — Core-Lab und andere
Beispiele bleiben unangetastet.

## Troubleshooting

- **MPD 404 / Smoke-Timeout**: FFmpeg-Generator hat das erste
  Manifest noch nicht geschrieben. Default-`WAIT_SECONDS=45` reicht
  in den meisten Fällen; bei langsameren Maschinen
  `WAIT_SECONDS=90 make smoke-dash`.
- **MPD 200, Init-Segment 404**: FFmpeg-Template hat einen anderen
  `init-`-Pfad gewählt (selten). Manifest direkt prüfen
  (`curl http://localhost:8891/manifest.mpd`); `initialization=`-
  Attribut zeigt den realen Pfad.
- **Port `8891` belegt**: anderer Prozess oder ein früherer
  `mtrace-dash`-Stack blockt. `docker compose -p mtrace-dash down
  --volumes` beenden, dann `ss -ltnp | grep 8891`.
- **Browser-Player zeigt CORS-Fehler**: das Beispiel setzt
  `Access-Control-Allow-Origin: *` in `nginx.conf`. Falls der Fehler
  trotzdem auftritt, Browser-DevTools-Console für den genauen
  Header-Fehler prüfen — ggf. eine Browser-Extension blockt.

## Bekannte Grenzen

- `@npm9912/stream-analyzer` analysiert DASH in `0.5.0` nicht
  produktiv: `analyzerKind: "hls"` ist die einzige produktive
  Variante; `dash`-/`cmaf`-Erweiterung ist Folge-Scope (MVP-37).
  Wer aus dem Beispiel heraus eine MPD an `POST /api/analyze`
  schickt, bekommt einen `not_hls`-Fehler — das ist **erwartetes**
  Verhalten in `0.5.0`.
- Kein `dash.js`-Adapter im `@npm9912/player-sdk`. Der Demo-Player
  in `apps/dashboard/src/routes/demo/` nutzt weiterhin `hls.js` mit
  dem Core-Lab-`teststream`; eine DASH-Demo-Route ist nicht Teil
  von `0.5.0`.
- DASH-Live-Window ist auf 20 s + 2 Extra-Segmente begrenzt; ältere
  Segmente löscht FFmpeg automatisch (`window_size 5`,
  `extra_window_size 2`). Late-Joiner haben < 30 s Catch-Up-Window
  — das ist Lab-typisch und nicht produktiv.
- nginx ohne TLS: das Beispiel servisiert HTTP, nicht HTTPS. Lokale
  Browser-Player akzeptieren das; produktive DASH-Setups würden
  HTTPS terminieren.
