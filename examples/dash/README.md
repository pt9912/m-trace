# DASH-Beispiel — Multi-Protocol Lab

> **Status**: Skelett. Inhalt liefert
> [`plan-0.5.0.md`](../../docs/planning/in-progress/plan-0.5.0.md) §5
> (Tranche 4, RAK-38).

## Zweck

Zeigt einen reproduzierbaren DASH-Pfad: lokale MPD-Ausspielung mit
einer minimalen Player-Test-Seite. Bezug: Lastenheft §7.8 F-82..F-84,
RAK-38.

**Wichtig**: `0.5.0`-DASH liefert ein lokal erreichbares MPD-/DASH-
Ausspielungsbeispiel. Es liefert **keine** vollständige DASH-
Manifestanalyse im `@npm9912/stream-analyzer` und **keinen**
dash.js-Adapter im Player-SDK. Diese Erweiterungen bleiben Folge-
Scope (siehe Plan §0.1 DASH-Zeile).

## Voraussetzungen

_Liefert Tranche 4._ Bisher offen: Docker Engine ≥ 24.0, Compose
v2.20, freier Port für DASH-HTTP-Auslieferung; Browser mit dash.js-
Demo-Player (CDN-Pfad oder lokal eingebunden).

## Start

```bash
docker compose -p mtrace-dash -f examples/dash/compose.yaml up -d --build
```

Project-Name `mtrace-dash` ist Pflicht (siehe
[`examples/README.md`](../README.md)).

_Compose-Datei und Detail-Konfiguration liefert Tranche 4._

## Verifikation

```bash
make smoke-dash
```

_Smoke-Skript liefert Tranche 4._ Pfad: Container generiert oder
serviert eine Demo-MPD; Smoke verifiziert, dass die MPD-URL einen
gültigen DASH-Manifest-Body liefert (Content-Type
`application/dash+xml` oder vergleichbar) und mindestens eine
Segment-URL auflösbar ist.

## Stop / Reset

```bash
docker compose -p mtrace-dash -f examples/dash/compose.yaml down
```

Hard-Reset: `down --volumes`. Greift nur den `mtrace-dash`-Project-
Namen.

## Troubleshooting

_Liefert Tranche 4._ Erwartete Einträge: Browser-CORS, MPD-Fetch-
Fehler durch fehlenden `Access-Control-Allow-Origin`-Header,
Codec-Inkompatibilität (z. B. AC-3 in Chromium ohne System-
Codec).

## Bekannte Grenzen

- `@npm9912/stream-analyzer` analysiert DASH in `0.5.0` nicht
  produktiv — die `hls.js`-Pfade aus `0.3.0` bleiben unverändert.
  DASH-Manifestanalyse ist Folge-Scope (siehe Plan §0.1 DASH-Zeile).
- Kein dash.js-Adapter im `@npm9912/player-sdk`. Der Demo-Player ist
  ein eigenständiges Lab-Setup, nicht Teil der Player-SDK-Bibliothek.
