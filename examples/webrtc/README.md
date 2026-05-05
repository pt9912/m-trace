# WebRTC-Beispiel — Multi-Protocol Lab (Vorbereitungspfad)

> **Status**: Skelett, Vorbereitungspfad. Inhalt liefert
> [`plan-0.5.0.md`](../../docs/planning/in-progress/plan-0.5.0.md) §6
> (Tranche 5, RAK-39).

## Zweck

Dokumentiert einen vorbereiteten WebRTC-Beispielplatz mit Ports und
Grenzen. **Kein produktives WebRTC-Monitoring**, kein Signaling-
Service und keine `getStats()`-Normalisierung in `0.5.0` (siehe Plan
§0.1 WebRTC-Zeile).

Ein `make smoke-webrtc-prep`-Target wird **nur** ergänzt, wenn der
Headless-Pfad in der CI stabil reproduzierbar ist; ansonsten bleibt
WebRTC ausschließlich als Doku-Pfad in `0.5.0`. Bezug: Lastenheft
§7.8, RAK-39.

## Voraussetzungen

_Liefert Tranche 5._ Bisher offen: Docker Engine ≥ 24.0, Compose
v2.20, ggf. STUN/TURN-Container, Browser mit WebRTC-Unterstützung
(Chromium/Firefox empfohlen, Safari als documented limitation).

## Start

```bash
docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml up -d --build
```

Project-Name `mtrace-webrtc` ist Pflicht (siehe
[`examples/README.md`](../README.md)).

_Compose-Datei liefert Tranche 5, falls überhaupt produktive
Container nötig sind. Andernfalls liefert die Tranche eine reine
Doku-Sektion ohne `compose.yaml`._

## Verifikation

```bash
make smoke-webrtc-prep    # ggf. — siehe Plan §6
```

_Status: Smoke-Target wird nur ergänzt, wenn der Pfad in der CI
zuverlässig läuft._ Andernfalls bleibt die Verifikation manuell:
Browser öffnet eine Demo-Seite, `getStats()`-Output ist im Browser
sichtbar — keine automatisierte Pipeline.

## Stop / Reset

```bash
docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml down
```

(Falls die Tranche keine Container ergänzt, ist dieser Block ein
No-op; die README dokumentiert das dann explizit.)

## Troubleshooting

_Liefert Tranche 5._ Erwartete Einträge: STUN/TURN-Erreichbarkeit,
Browser-`getStats()`-Schema-Drift zwischen Browser-Versionen,
Headless-Chrome-Restriktionen für WebRTC.

## Bekannte Grenzen

- Keine produktive WebRTC-Observability in `0.5.0`: kein
  Signaling-Service in `apps/api`, keine `getStats()`-Sammlung im
  Player-SDK, keine Aggregat-Metriken in Prometheus.
- `0.5.0` zeigt höchstens einen vorbereiteten Container-Pfad und
  Doku-Stellen, die spätere WebRTC-Releases referenzieren können.
- Kein `dashboard`-Hook für WebRTC-Sessions; die Session-Timeline
  bleibt auf hls.js-Quellen aus `0.4.0` Tranche 4 beschränkt.
