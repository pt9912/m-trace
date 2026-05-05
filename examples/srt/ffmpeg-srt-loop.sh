#!/bin/sh
# plan-0.5.0 §4 Tranche 3 — FFmpeg-Loop-Publisher für das SRT-Beispiel.
#
# Erzeugt einen synthetischen Test-Stream (Test-Pattern + Sine-Audio)
# und publishet ihn per SRT in den lokalen MediaMTX-Container. Das ist
# das SRT-Pendant zum services/stream-generator/ffmpeg-loop.sh aus dem
# Core-Lab; die FFmpeg-Optionen sind absichtlich minimal gehalten.

set -eu

SRT_URL="${SRT_URL:-srt://mediamtx:8890?streamid=publish:srt-test&pkt_size=1316}"

while :; do
  ffmpeg -hide_banner -loglevel warning \
    -re \
    -f lavfi -i "testsrc2=size=1280x720:rate=30" \
    -f lavfi -i "sine=frequency=1000:sample_rate=48000" \
    -c:v libx264 -preset veryfast -tune zerolatency -g 60 -keyint_min 60 -b:v 1M \
    -c:a aac -b:a 96k -ar 48000 -ac 2 \
    -f mpegts "$SRT_URL" \
    || sleep 2
done
