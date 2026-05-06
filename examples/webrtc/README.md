# WebRTC-Beispiel — Multi-Protocol Lab

> **Status**: Lab-Compose ab `0.7.0` Tranche 1 (RAK-47). Liefert einen
> lokal startbaren WHIP-/WHEP-Lab-Pfad mit FFmpeg-Publisher (RTSP-
> Push) und Browser-Handcheck. Endpoint-/compose-only Smoke
> (`make smoke-webrtc-prep`) folgt in Tranche 3.
>
> Bezug: Lastenheft `1.1.9` §7.6 F-62, §7.8 F-82..F-84, §8.3 NF-14,
> §12.1 MVP-24, §13.9 RAK-47..RAK-50;
> [`docs/planning/done/plan-0.7.0.md`](../../docs/planning/done/plan-0.7.0.md)
> §2 Tranche 1.
>
> Quickref aller Multi-Protocol-Lab-Beispiele:
> [`docs/user/local-development.md`](../../docs/user/local-development.md)
> §2.7.

## Zweck

Ein lokal startbarer WebRTC-Lab-Stack, der zeigt, wie ein WHIP-/WHEP-
Pfad gegen einen MediaMTX-basierten Lab-Server aussieht. Die
Publisher-Seite läuft als FFmpeg-Container, der per RTSP in MediaMTX
pushed (F-84 Muss); MediaMTX exposed denselben Stream zusätzlich als
WHIP-Publish- und WHEP-Read-Endpoint für einen Browser-Handcheck
(RAK-50, Browser-WHIP-Push optional manueller Pfad).

Das Beispiel ersetzt **nicht** den Core-Lab-HLS-Pfad — WebRTC ist
keine produktive Telemetrie- oder Dashboard-Quelle in `0.7.0`. Siehe
„Bekannte Grenzen".

## Voraussetzungen

- Docker Engine ≥ 24.0, Compose v2.20.
- Browser mit WebRTC-Unterstützung für den Handcheck: Chromium 120+
  oder Firefox 120+. Safari als Best-Effort (Codec-/ICE-Verhalten
  abweichend, nicht Pflicht-Browser für RAK-50).
- Freie Host-Ports: `8892/tcp` (WHIP/WHEP-HTTP), `8189/udp` (WebRTC-
  ICE-Media), `9999/tcp` (MediaMTX-Control-API). Kollisionsfrei zu
  Core-Lab (`8888`/`9997`), `mtrace-srt` (`8889`/`8890`/`9998`) und
  `mtrace-dash` (`8891`).
- Kein TLS-/Public-Internet-Setup. localhost ist der Pflichtpfad;
  STUN/TURN bleibt opt-in (siehe „Troubleshooting").

## Start

```bash
docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml up -d --build
```

Project-Name `mtrace-webrtc` ist Pflicht ([`examples/README.md`](../README.md)
Sektion „Project-Name-Pflicht für eigenes Compose").

Nach dem Hochfahren erzeugt der `webrtc-publisher`-Container einen
synthetischen Test-Stream (`testsrc2` 1280×720 @ 30 fps + `sine` 1 kHz)
und pushed ihn per RTSP in MediaMTX. MediaMTX re-published den Stream
unter dem WebRTC-Pfad `webrtc-test`.

WHIP-/WHEP-URL-Form (lokales Lab; gepinntes Image
`bluenviron/mediamtx:1`, getestet mit MediaMTX `1.18.1`):

| Pfad | URL |
|---|---|
| WHIP (Publish) | `http://localhost:8892/webrtc-test/whip` |
| WHEP (Read)    | `http://localhost:8892/webrtc-test/whep` |
| MediaMTX-API   | `http://localhost:9999/v3/paths/list` |

## Verifikation

### Endpoint-Probe (browserfrei, Tranche-3-Smoke-Vorbereitung)

```bash
# 1. Stream ist registriert (MediaMTX-API):
curl -sS -u any: http://localhost:9999/v3/paths/list | grep -q '"webrtc-test"'

# 2. WHEP-Endpoint antwortet (aktiver Pfad → 204):
curl -sS -o /dev/null -w "%{http_code}\n" -X OPTIONS http://localhost:8892/webrtc-test/whep
# → 204

# 3. WHIP-Endpoint antwortet (aktiver Pfad → 204):
curl -sS -o /dev/null -w "%{http_code}\n" -X OPTIONS http://localhost:8892/webrtc-test/whip
# → 204
```

Erwarteter Statussatz für die spätere Smoke-Implementierung
([`plan-0.7.0.md`](../../docs/planning/done/plan-0.7.0.md) Tranche 3):

| Methode | Pfad | Bedingung | Status |
|---|---|---|---|
| `OPTIONS` | `/webrtc-test/whep` | Compose oben + Stream aktiv | `204` |
| `OPTIONS` | `/webrtc-test/whip` | Compose oben + Stream aktiv | `204` |
| `OPTIONS` | `/missing/whep`     | Compose oben, Pfad unbekannt | `500` |
| `GET`/`HEAD` | beide | Endpoint-Existenz, falsche Methode | `405` |
| `POST` | beide | ohne SDP-Body | `400` |
| beliebig | beliebig | Compose nicht oben | `Connection refused` |

Die Probe weist nach, dass MediaMTX läuft, der Stream registriert
ist und die WHIP-/WHEP-Listener bedient sind. Sie weist **nicht**
nach:

- dass ein realer Browser einen Stream **abspielen** kann
  (Playback-Qualität),
- dass die ICE-Aushandlung mit einem Browser **erfolgreich** ist
  (ICE-Erfolgsquote),
- dass `getStats()` zwischen Browser-Versionen **stabile** Felder
  liefert (Schema-Drift, siehe Tranche 4).

Diese drei Aspekte deckt nur der Browser-Handcheck ab.

### Browser-Handcheck (RAK-50, manuell)

MediaMTX bringt eine eingebaute WebRTC-Read-Demo-Seite mit:

```text
http://localhost:8892/webrtc-test
```

In Chromium oder Firefox aufrufen → Video- und Audio-Spur sollten
spielen. Erwartung: Test-Pattern + 1 kHz Sinuston, latenzarm.

`getStats()`-Inspektion über `chrome://webrtc-internals` (Chromium)
oder `about:webrtc` (Firefox) zeigt aktive `RTCPeerConnection` mit
`connection_state=connected`, `ice_state=connected`,
`dtls_state=connected`. Schema-Drift zwischen Browser-Versionen ist
in [`spec/telemetry-model.md`](../../spec/telemetry-model.md) §3.2
beschrieben (siehe `0.7.0` Tranche 4).

## Stop / Reset

```bash
# Stack stoppen, Volumes behalten:
docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml down

# Stack stoppen + alle eigenen Volumes/Netze entfernen:
docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml down --volumes
```

Greift nur das `mtrace-webrtc`-Project. Core-Lab (`mtrace`),
`mtrace-srt` und `mtrace-dash` bleiben unangetastet.

## Troubleshooting

- **`bind: address already in use` auf 8892/8189/9999** — anderer
  Prozess belegt einen der Host-Ports. `ss -tulpn | grep -E ':(8892|8189|9999)'`
  zeigt den Konflikt; entweder den Prozess beenden oder einen
  Override-Compose mit alternativen Ports anlegen.
- **`OPTIONS /webrtc-test/whep → 500`** — der Stream ist (noch)
  nicht in MediaMTX angekommen. Prüfen mit
  `docker logs mtrace-webrtc-webrtc-publisher-1` und
  `curl -u any: http://localhost:9999/v3/paths/list`. FFmpeg
  braucht ~3-5 s nach dem `up -d`, bis der erste Frame durch ist.
- **Browser zeigt schwarzes Bild oder kein Audio** — ICE-Negotiation
  schlägt fehl, weil der Browser die advertised Kandidaten nicht
  erreicht. `webrtcAdditionalHosts` in `mediamtx.yml` listet
  `127.0.0.1` und `mediamtx`; bei Zugriff vom Host muss der
  `127.0.0.1`-Kandidat genommen werden. `chrome://webrtc-internals`
  zeigt die ICE-Pair-Kandidaten und Auswahl.
- **WebRTC läuft nur über localhost, nicht über LAN** — bewusst.
  Für LAN-Pfade müsste ein zusätzlicher `coturn`-Container die
  STUN-/TURN-Resolution liefern und `webrtcICEServers2` in der
  MediaMTX-Konfig konfiguriert werden. Das ist Folge-Scope, nicht
  `0.7.0`.
- **FFmpeg-Publisher loggt `encoder 'opus' is experimental`** — das
  Skript nutzt `libopus` als Encoder; falls eine alternative
  FFmpeg-Image-Variante das nicht hat, `-strict experimental`
  ergänzen oder ein Image mit libopus-Build wählen.
- **MediaMTX-API liefert `401 unauthorized`** — der Lab-Auth-Block
  in `mediamtx.yml` erlaubt `user any` mit leerem Passwort. `curl
  -u any: …` (Doppelpunkt nach dem User für leeres Passwort)
  reicht; ohne `-u` lehnt MediaMTX 1.14+ default ab.

## Bekannte Grenzen

- **Kein produktiver `apps/api`-WebRTC-Ingress.** Das Beispiel ist
  ein Lab-Pfad ohne `mtrace_webrtc_*`-Metriken. Tranche 4 erweitert
  `spec/telemetry-model.md` §3.2 um die bounded WebRTC-Aggregat-
  Allowlist; eine produktive Telemetrie-Anbindung braucht einen
  eigenen Folgeplan.
- **Kein Player-SDK-WebRTC-Adapter.** RAK-51 ist deferred (siehe
  `plan-0.7.0.md` §7); der `@npm9912/player-sdk` bleibt auf
  `hls.js`-only, ohne Codepfad-Vermischung.
- **Kein TLS, kein Public-Internet, kein NAT-Traversal über LAN
  hinaus.** Der Compose-Stack zielt auf `localhost`-Tests; STUN/TURN
  und HTTPS sind Folge-Scope.
- **Kein Dashboard-Hook für WebRTC-Sessions.** Die Session-Timeline
  bleibt in `0.7.0` auf hls.js-Quellen aus `0.4.0` Tranche 4
  beschränkt.
- **Headless-Browser-Smoke ist optional und nicht release-blockierend
  in `0.7.0`.** Tranche 3 liefert ein endpoint-/compose-only
  `make smoke-webrtc-prep` ohne Browser. Headless-Chrome- oder
  Playwright-Erweiterungen können additiv ergänzt werden, kippen
  aber nicht den Muss-Pfad.
