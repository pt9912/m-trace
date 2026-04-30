#!/bin/sh
set -eu

TARGET_URL="${STREAM_TARGET_URL:-rtsp://mediamtx:8554/teststream}"

while true; do
  ffmpeg -hide_banner -loglevel warning -re \
    -f lavfi -i testsrc=size=1280x720:rate=30 \
    -f lavfi -i sine=frequency=1000:sample_rate=48000 \
    -c:v libx264 -preset veryfast -tune zerolatency -pix_fmt yuv420p \
    -g 60 -b:v 2500k \
    -c:a aac -b:a 128k \
    -f rtsp -rtsp_transport tcp "$TARGET_URL" || true

  sleep 2
done
