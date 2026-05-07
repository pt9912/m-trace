#!/bin/sh
# plan-0.9.0 §3 Tranche 2 — FFmpeg-Loop-Publisher für das SRS-Beispiel.
#
# Erzeugt einen synthetischen Test-Stream (Test-Pattern + Sine-Audio)
# und publishet ihn per RTMP in den lokalen SRS-Container. Das ist
# das RTMP-Pendant zum services/stream-generator/ffmpeg-loop.sh aus
# dem Core-Lab und zu examples/srt/ffmpeg-srt-loop.sh; die FFmpeg-
# Optionen sind absichtlich minimal gehalten.

set -eu

RTMP_URL="${RTMP_URL:-rtmp://srs:1935/live/srs-test}"

while :; do
  ffmpeg -hide_banner -loglevel warning \
    -re \
    -f lavfi -i "testsrc2=size=1280x720:rate=30" \
    -f lavfi -i "sine=frequency=1000:sample_rate=48000" \
    -c:v libx264 -preset veryfast -tune zerolatency -g 60 -keyint_min 60 -b:v 1M \
    -c:a aac -b:a 96k -ar 48000 -ac 2 \
    -f flv "$RTMP_URL" \
    || sleep 2
done
