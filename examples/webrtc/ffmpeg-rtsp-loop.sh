#!/bin/sh
#  — FFmpeg-Loop-Publisher für das WebRTC-Beispiel.
#
# Erzeugt einen synthetischen Test-Stream (Test-Pattern + Sine-Audio)
# und publishet ihn per RTSP in den lokalen MediaMTX-Container. MediaMTX
# re-published denselben Stream automatisch auf den WHEP-Read-Pfad —
# der Browser-Handcheck (RAK-50) liest dann via WebRTC. FFmpeg-WHIP
# ist nicht der Pflichtpfad.

set -eu

RTSP_URL="${RTSP_URL:-rtsp://mediamtx:8554/webrtc-test}"

while :; do
  ffmpeg -hide_banner -loglevel warning \
    -re \
    -f lavfi -i "testsrc2=size=1280x720:rate=30" \
    -f lavfi -i "sine=frequency=1000:sample_rate=48000" \
    -c:v libx264 -preset veryfast -tune zerolatency -g 60 -keyint_min 60 -b:v 1M \
    -c:a libopus -b:a 96k -ar 48000 -ac 2 \
    -f rtsp -rtsp_transport tcp "$RTSP_URL" \
    || sleep 2
done
