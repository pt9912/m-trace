#!/usr/bin/env bash
# baseline-verify — prueft die vendored Baseline dieses Repos
# (.harness/baseline/<tag>/{regelwerk,templates}/ + SHA256SUMS) NETZLOS.
#
# Emittiert von ai-harness-init (LH-FA-09). Tool-als-Quelle: das Skript ist
# generiert, nicht aus dem Kurs kopiert — es gehoert zur selben Herkunftsklasse
# wie das Verzeichnis-Geruest (ADR-0005).
#
# ZWEI Pruefungen, beide noetig:
#   1. Integritaet     — sha256sum -c ueber SHA256SUMS: erkennt GEAENDERTE und
#                        GELOESCHTE Dateien.
#   2. Vollstaendigkeit — Dateibestand == SHA256SUMS-Liste: erkennt ZUSAETZLICH
#                        EINGELEGTE Dateien. Ohne diesen Schritt bliebe die
#                        Pruefung dabei GRUEN, denn `sha256sum -c` prueft nur,
#                        was gelistet ist. "Integritaet der Arbeitskopie" waere
#                        dann ueberdehnt — genau das stille Gruen, das eine
#                        Gate-Behauptung wertlos macht.
#
# Die Baseline ist committet: ein fehlendes Verzeichnis bedeutet einen kaputten
# Checkout, keinen ausstehenden Fetch. Der Bootstrap-Fetch laeuft EINMAL; danach
# ist dieses Repo offline pruefbar.
#
# Was dieses Skript NICHT leistet: Upstream-Drift. SHA256SUMS ist selbst erzeugt
# und belegt nur, dass der Baum sich seit dem Bootstrap nicht bewegt hat — NICHT
# seine Herkunft. Die haengt am sha256 des Release-Assets, gegen den beim
# Bootstrap VOR dem Entpacken geprueft wurde.
#
# Keine Fremd-Laufzeit (node/jq/python): bash + coreutils.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
base="$here/../../.harness/baseline"

# <tag>-Politik: ein Tag zur Zeit (Ersetzen), Historie liegt in git. Das
# Verzeichnis wird ENTDECKT, nicht geraten — so steht der Tag-String an genau
# einer Stelle. Mehr als ein <tag>-Verzeichnis verletzt die Setzung und schlaegt
# hier an, statt still zu passieren.
shopt -s nullglob
dirs=("$base"/*/)
shopt -u nullglob

if [ "${#dirs[@]}" -eq 0 ]; then
  echo "FEHLER: keine vendored Baseline unter .harness/baseline/<tag>/." >&2
  echo "  Die Baseline ist committet — ein leeres Verzeichnis bedeutet einen kaputten Checkout." >&2
  exit 1
fi
if [ "${#dirs[@]}" -gt 1 ]; then
  echo "FEHLER: mehr als ein <tag>-Verzeichnis unter .harness/baseline/:" >&2
  printf '  %s\n' "${dirs[@]}" >&2
  echo "  Setzung: ein Tag zur Zeit (Ersetzen), Historie liegt in git." >&2
  exit 1
fi

dir="${dirs[0]%/}"
tag="$(basename "$dir")"
cd "$dir"

if [ ! -f SHA256SUMS ]; then
  echo "FEHLER: $tag/SHA256SUMS fehlt — die Baseline ist ohne Pruefsummen nicht verifizierbar." >&2
  exit 1
fi

# 0) Format-Vorbedingung ZUERST: GNU sha256sum ESCAPT Dateinamen mit
# Backslash/Newline (fuehrender Backslash am Zeilenanfang, verdoppelte im Pfad).
# Der Vollstaendigkeits-Vergleich unten dekodiert das NICHT und wuerde eine
# solche Datei faelschlich als abweichend melden — Rot ohne Manipulation. Hier
# LAUT abbrechen, BEVOR ein Urteil faellt: ehrlich "kann ich nicht" schlaegt
# still "alles gut". Regex '^[\]' statt '^\\' (Letzteres loest SC1003 aus).
if grep -q '^[\]' SHA256SUMS; then
  printf '%s\n' "FEHLER: SHA256SUMS enthaelt GNU-escapte Pfade (fuehrender Backslash) — der Vollstaendigkeits-Check dekodiert die nicht und wuerde falsch-positiv melden. Baum manuell pruefen." >&2
  exit 1
fi

# 1) Integritaet der gelisteten Dateien (geaendert/geloescht). Kein --quiet
# (GNU-only) — stattdessen Output unterdruecken, nur der Exit-Code zaehlt.
if ! sha256sum -c SHA256SUMS >/dev/null 2>&1; then
  echo "FEHLER: Baseline $tag weicht von SHA256SUMS ab (geaenderte oder fehlende Datei)." >&2
  exit 1
fi

# 2) Vollstaendigkeit (zusaetzlich eingelegtes Artefakt). SHA256SUMS selbst ist
# ausgenommen — sie kann sich nicht selbst hashen; ihre Integritaet traegt git.
#
# `! -type d` statt `-type f`: ein eingelegter SYMLINK (oder jede andere
# Nicht-Regulaer-Datei) ist weder in SHA256SUMS gelistet noch von `-type f`
# sichtbar — beide Achsen blieben gruen, waehrend der Baum Fremdinhalt liefert.
# Genau das stille Gruen, das dieser Check verhindern soll. Mit `! -type d`
# taucht der Symlink im Ist-Bestand auf, fehlt in der Soll-Liste und schlaegt an.
#
# Soll-Seite normalisiert: dieses Repo listet in SHA256SUMS mit ./-Praefix
# (Bootstrap-Form), die Emission ohne. Der Praefix ist Form, kein Befund —
# die Digests selbst prueft `sha256sum -c` oben unangetastet gegen die
# committete SHA256SUMS.
listed="$(cut -d' ' -f3- SHA256SUMS | sed 's|^\./||' | LC_ALL=C sort)"
actual="$(find . ! -type d ! -path ./SHA256SUMS | sed 's|^\./||' | LC_ALL=C sort)"
if [ "$listed" != "$actual" ]; then
  echo "FEHLER: Dateibestand von $tag weicht von SHA256SUMS ab (ungelistete oder fehlende Pfade):" >&2
  diff <(printf '%s\n' "$listed") <(printf '%s\n' "$actual") >&2 || true
  exit 1
fi

echo "baseline-verify: $tag OK — $(wc -l < SHA256SUMS) Dateien (Integritaet + Vollstaendigkeit, netzlos)"
