#!/bin/sh
# plan-0.5.0 §5 Tranche 4 — FFmpeg-Loop für das DASH-Beispiel.
#
# Erzeugt einen synthetischen Test-Stream und schreibt einen
# DASH-Live-Output (manifest.mpd + Init-/Media-Segmente) in das
# shared Volume `/dash/output`. Der nginx-Container serviert das
# Verzeichnis statisch — externer Internet-Zugriff ist nicht nötig.

set -eu

DASH_OUT="${DASH_OUT:-/dash/output}"

mkdir -p "$DASH_OUT"

# DASH-Live-Muxer-Optionen:
#   -seg_duration 4         — 4s pro Segment
#   -window_size 5          — 5 aktive Segmente (Live-Window 20s)
#   -extra_window_size 2    — 2 zusätzliche Segmente außerhalb des
#                              aktiven Window für Late-Joiner
#   -remove_at_exit 0       — Segmente beim FFmpeg-Stop nicht löschen
#                              (Smoke kann nach Process-Stop weiter lesen)
#   -use_template 1, -use_timeline 1 — modernes Template-/Timeline-Modell
while :; do
  ffmpeg -hide_banner -loglevel warning \
    -re \
    -f lavfi -i "testsrc2=size=1280x720:rate=30" \
    -f lavfi -i "sine=frequency=1000:sample_rate=48000" \
    -c:v libx264 -preset veryfast -tune zerolatency -g 60 -keyint_min 60 -b:v 1M \
    -c:a aac -b:a 96k -ar 48000 -ac 2 \
    -f dash \
    -seg_duration 4 \
    -window_size 5 \
    -extra_window_size 2 \
    -remove_at_exit 0 \
    -use_template 1 \
    -use_timeline 1 \
    -dash_segment_type mp4 \
    "$DASH_OUT/manifest.mpd" \
    || sleep 2
done
