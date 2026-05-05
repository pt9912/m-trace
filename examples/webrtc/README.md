# WebRTC-Beispiel — Multi-Protocol Lab (Vorbereitungspfad)

> **Variante**: Doku-only Vorbereitungspfad. Tranche 5 liefert
> bewusst **keine** `examples/webrtc/compose.yaml`, **kein**
> `make smoke-webrtc-prep`-Target und **keinen** Browser-Handcheck.
> Diese Stelle dokumentiert die Konfigurations-Grenze für eine
> zukünftige Lab-Erweiterung — produktives WebRTC-Monitoring ist
> Folge-Scope.
>
> Bezug: Lastenheft §7.6 F-62, §8.3 NF-14, §12.1 MVP-24, RAK-39;
> [`plan-0.5.0.md`](../../docs/planning/in-progress/plan-0.5.0.md) §6
> (Tranche 5) und §0.1 Tabellen-Zeile „WebRTC".
>
> Quickref aller Multi-Protocol-Lab-Beispiele:
> [`docs/user/local-development.md`](../../docs/user/local-development.md)
> §2.7.

## Zweck

Ein vorbereiteter Beispielplatz für eine zukünftige WebRTC-
Lab-Erweiterung. `0.5.0` dokumentiert hier ausschließlich:

- die voraussichtlich nötigen Ports und NAT-/ICE-Grenzen,
- die Out-of-Scope-Klauseln (kein Signaling-Server, keine
  `getStats()`-Erfassung, keine Dashboard-Sichtbarkeit),
- die Entscheidungsmarke, dass `0.5.0` keinen Smoke ergänzt.

So ist nachvollziehbar, was eine spätere Tranche oder ein folgendes
Release zu liefern hätte, ohne dass `0.5.0` die Surface schon
einbaut.

## Voraussetzungen (geplant)

Für eine spätere produktive Tranche wäre voraussichtlich nötig:

- Docker Engine ≥ 24.0, Compose v2.20.
- Browser mit WebRTC-Unterstützung (Chromium/Firefox empfohlen,
  Safari als documented limitation).
- TCP-Port `8889` für MediaMTX-WebRTC-WHIP-/WHEP-Endpoint
  (vorgesehen) — kollidiert in `0.5.0` mit dem SRT-Beispiel
  (`examples/srt/`, Host-Port `8889`); eine Folge-Tranche muss
  Port-Mapping pro Beispiel neu schneiden, falls SRT und WebRTC
  parallel laufen sollen.
- Optional ein STUN/TURN-Container (z. B. `coturn`) für ICE-
  Negotiation, falls der Lab-Pfad nicht nur localhost abdeckt.

## Start

`0.5.0` liefert keinen Startpfad. Für die zukünftige Tranche ist
vorgesehen:

```bash
# Geplant — nicht in 0.5.0 enthalten:
# docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml up -d --build
```

Project-Name `mtrace-webrtc` ist in der Konvention
([`examples/README.md`](../README.md) Sektion „Project-Name-Pflicht
für eigenes Compose") reserviert.

## Verifikation

`0.5.0` liefert keinen Smoke. Begründung:

- **Headless-Browser-WebRTC ist instabil in CI.** Chromium/Firefox-
  Headless-Modi haben Browser-Versions-spezifisches Verhalten bei
  ICE-Aushandlung und Codec-Negotiation. Ein Smoke, der lokal grün
  läuft und in CI rot wird, würde mehr Lärm als Nutzen erzeugen.
- **`getStats()` ist Browser-spezifisch.** Eine seriöse Smoke-
  Verifikation müsste Schema-Drifts zwischen Chromium- und Firefox-
  Versionen handhaben; das ist ein eigenes Folge-Thema, nicht
  Multi-Protocol-Lab-Scope.
- **Lab-Wert ohne Smoke ist begrenzt.** Solange WebRTC nicht
  produktiv im Dashboard-/`apps/api`-Pfad sichtbar ist, fehlt der
  konkrete Operator-Use-Case für ein laufendes Beispiel.

Wenn eine Folge-Tranche WebRTC produktiv macht, ist der Smoke-Name
[`make smoke-webrtc-prep`](../../Makefile) für „Vorbereitungs-
Smoke" (Port-/Konfig-Check) reserviert (siehe
[`examples/README.md`](../README.md) Sektion „Smoke-Targets").

## Stop / Reset

`0.5.0` startet keinen Stack — kein Stop nötig. Geplant für eine
Folge-Tranche:

```bash
# Geplant — nicht in 0.5.0 enthalten:
# docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml down
```

## Troubleshooting

`0.5.0` hat keinen aktiven Pfad zum Troubleshooten. Erwartete
Folgepunkte einer späteren Tranche:

- ICE-Negotiation-Fehler durch fehlenden STUN/TURN.
- Browser-Versions-Drift bei `getStats()`-Schema (Chromium vs.
  Firefox vs. Safari).
- Headless-Chrome-Restriktionen für WebRTC (insbesondere
  `--use-fake-ui-for-media-stream` und Codec-Allowlist).
- Port-Konflikt zwischen `examples/srt/` (Host `8889`) und einem
  geplanten WebRTC-Port; spätere Tranche muss Ports neu schneiden.

## Bekannte Grenzen

- **Kein Signaling-Server in `apps/api`.** WebRTC-WHIP/-WHEP-
  Endpoints sind nicht im m-trace-Backend implementiert. Die
  zukünftige Tranche müsste entweder MediaMTX-WHIP nutzen oder
  einen eigenen, klar separierten Signaling-Pfad einführen.
- **Keine `getStats()`-Sammlung im `@npm9912/player-sdk`.** Public-
  API bleibt unverändert; ein WebRTC-Adapter-Pfad würde additiv
  ergänzt, ohne den `hls.js`-Pfad zu brechen.
- **Keine WebRTC-Aggregat-Metriken in Prometheus.** Forbidden-Liste
  aus [`spec/telemetry-model.md`](../../spec/telemetry-model.md)
  §3.1 gilt unverändert; ein Folge-Pfad müsste WebRTC-spezifische
  bounded Aggregat-Labels (z. B. `connection_state`, `ice_state`)
  vorab in §3.2 freischalten.
- **Kein Dashboard-Hook für WebRTC-Sessions.** Die Session-Timeline
  bleibt in `0.5.0` auf hls.js-Quellen aus `0.4.0` Tranche 4
  beschränkt.

## Folge-Pfad (jenseits `0.5.0`)

Was eine spätere Phase aus diesem Vorbereitungspfad eine produktive
WebRTC-Lab-Erweiterung machen würde, in grober Reihenfolge:

1. **Lab-Compose** unter `examples/webrtc/compose.yaml` mit
   MediaMTX-WHIP-/WHEP-Endpoint und `coturn`-Container (falls
   nicht-localhost-Pfade getestet werden sollen).
2. **README-Konkretisierung** mit Operator-Befehlen, Port-Schnitt
   gegen die anderen Lab-Beispiele und Browser-Handcheck-Anleitung.
3. **`make smoke-webrtc-prep`-Target**, das ausschließlich
   Vorbereitungsgrenzen prüft (Compose hochgefahren, Endpoints
   antworten, kein Playback-/`getStats()`-Anspruch). Reservierter
   Target-Name siehe `examples/README.md` Sektion „Smoke-Targets".
4. **Bewertung WebRTC-Telemetrie** in einem eigenen Plan-Schnitt:
   bounded Allowlist-Labels, `getStats()`-Subset-Auswahl, Schema-
   Drift-Strategie. Erst danach würde ein produktiver Pfad ins
   `apps/api`/Dashboard-System kommen.

Diese Schritte sind ausdrücklich **nicht** Teil von `0.5.0` und
nicht in der `0.5.0`-Roadmap als Pflicht aufgenommen — das
Multi-Protocol-Lab fokussiert MediaMTX, SRT und DASH.
