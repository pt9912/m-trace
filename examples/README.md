# `examples/` — Multi-Protocol Lab Beispiele

> **Status**: `0.5.0`-Lieferung in Arbeit. Diese Stelle liefert
> reproduzierbare lokale Beispiele für die Streaming-Protokolle, die
> das Multi-Protocol Lab (Lastenheft §7.8 + RAK-36..RAK-39) abdeckt.
> Bezug: [`docs/planning/in-progress/plan-0.5.0.md`](../docs/planning/in-progress/plan-0.5.0.md).

## Zweck

`examples/` bündelt protokollspezifische Lab-Setups, die unabhängig vom
Core-Lab (`make dev`) starten. Jeder Eintrag ist ein eigenständiges
Sub-Verzeichnis mit kurzer README, einer eigenen `compose.yaml` (oder
einem klar benannten Compose-Profil) und — soweit sinnvoll — einem
opt-in Smoke-Target im Root-`Makefile`.

Das Core-Lab (`make dev`) bleibt unverändert; Beispiele laufen neben
oder anstatt des Core-Labs, niemals automatisch zusammen.

## Konventionen

### README-Mindeststruktur

Jede `examples/<name>/README.md` folgt derselben Reihenfolge, damit ein
neuer Operator den Pfad ohne Code-Lesen findet:

1. **Zweck** — was das Beispiel zeigt; Bezug zu RAK/Lastenheft.
2. **Voraussetzungen** — Tool-Versionen, Image-Pulls, Out-of-Scope-Hinweise.
3. **Start** — exakter Befehl plus Compose-Projektname (siehe unten).
4. **Verifikation** — `curl`/Browser-Check oder Smoke-Target-Aufruf,
   inklusive der erwarteten Ports/URLs.
5. **Stop / Reset** — Befehl, der nur dieses Beispiel beendet bzw.
   sein Volume entfernt; **darf keine fremden Volumes/Container
   anrühren**.
6. **Troubleshooting** — typische Fehlerbilder (Port-Konflikt,
   fehlende Codec-Unterstützung, …) mit konkretem Workaround.
7. **Bekannte Grenzen** — was das Beispiel ausdrücklich **nicht**
   liefert (Out-of-Scope-Folgepfade in spätere Releases).

### Compose-Form

Jedes Beispiel hat eine eigene `examples/<name>/compose.yaml`. Das
Root-`docker-compose.yml` bleibt das Core-Lab; es wird nicht mit
optionalen Beispiel-Profilen überfrachtet. (Beschluss
`plan-0.5.0.md` §0.1 Zeile „Compose-Form".)

**Project-Name-Pflicht**: weil eigene Compose-Dateien sonst per Default
denselben Project-Namen wie das Core-Lab nutzen würden, schreibt jede
README explizit `-p mtrace-<name>` im Startbefehl vor:

```bash
docker compose -p mtrace-<name> -f examples/<name>/compose.yaml up -d --build
```

Damit kollidieren `examples/<name>`-Container und -Volumes nicht mit
`mtrace-data`/`mtrace-tempo-data` aus dem Core-Lab.

### Smoke-Targets

Beispiel-Smokes sind opt-in im Root-`Makefile` und folgen dem
existierenden Naming `smoke-<name>` (analog `smoke-observability`,
`smoke-tempo`, `smoke-analyzer`, `smoke-cli`). Sie landen
**nicht** in `make gates`, weil sie zusätzliche Streaming-Images
oder lange Medien-Starts brauchen, die jeder PR-Run nicht tragen kann.
(Beschluss `plan-0.5.0.md` §0.1 Zeile „Smoke-Targets".)

Geplante Targets in `0.5.0`:

| Target                  | Tranche | Status                              |
|-------------------------|---------|-------------------------------------|
| `make smoke-mediamtx`   | 2       | ⬜                                  |
| `make smoke-srt`        | 3       | ⬜                                  |
| `make smoke-dash`       | 4       | ⬜                                  |
| `make smoke-webrtc-prep`| 5       | ⬜ — nur falls headless stabil      |

### Smoke-Skript-Konvention

Smoke-Skripte liegen unter `scripts/smoke-<name>.sh` (analog
`scripts/smoke-observability.sh`, `scripts/smoke-tempo.sh`).
Anforderungen:

- `set -euo pipefail` als erste Zeile.
- Klare Fehlermeldungen mit `[smoke-<name>]`-Präfix auf stderr.
- **Räumt keine fremden Container/Volumes auf** — Compose-Down nur
  für den eigenen Project-Namen (`docker compose -p mtrace-<name>
  down`), Volumes nur, wenn das Beispiel sie selbst angelegt hat.
- Exit-Code: 0 grün, ≥1 rot. Reproduzierbar lokal und im Lab.

## Beispiele

| Verzeichnis              | Tranche | Status                          |
|--------------------------|---------|---------------------------------|
| [`mediamtx/`](./mediamtx/) | 2     | Skelett (Konventions-konform)   |
| [`srt/`](./srt/)         | 3       | Skelett (Konventions-konform)   |
| [`dash/`](./dash/)       | 4       | Skelett (Konventions-konform)   |
| [`webrtc/`](./webrtc/)   | 5       | Skelett (Vorbereitungspfad)     |

Inhalte folgen in den jeweiligen Tranchen. Bis dahin ist jedes
Sub-Verzeichnis ein READMEs-only-Platzhalter, der die Mindeststruktur
oben belegt.
