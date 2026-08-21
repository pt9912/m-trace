#!/usr/bin/env bash
# Prueft, dass die Runtime-Images ueberhaupt STARTEN.
#
# Luecke, die dieses Script schliesst: `make image-scan` baut die drei
# Images und scannt sie statisch, `make gates` fasst sie gar nicht an.
# Ein Image, das beim ersten `docker run` sofort stirbt, passiert damit
# jede Pruefung gruen. Genau so blieb unbemerkt, dass das
# analyzer-service-Image nicht lauffaehig war (`Cannot find module
# '@pt9912/stream-analyzer'`, behoben in 373db24) — aufgefallen ist es
# nur zufaellig beim Verifizieren einer anderen Aenderung.
#
# Kriterium ist bewusst schmal: Container starten, Karenz abwarten,
# feststellen dass er NOCH LAEUFT. Das trennt den Defektfall (Exit
# innerhalb von Millisekunden) sauber vom gesunden Fall und kommt ohne
# Health-Endpoint aus — das API-Image ist distroless und hat keine
# Shell, in der man proben koennte.
#
# Was dieses Script NICHT behauptet: dass ein Image fachlich
# funktioniert. Dafuer sind die Compose-Smokes da (`make
# smoke-analyzer` & Co.). Hier geht es nur um: startet es ueberhaupt.
set -euo pipefail

IMAGES="${IMAGES:-mtrace-api:scan mtrace-dashboard:scan mtrace-analyzer-service:scan}"
# Karenz in Sekunden. Der gejagte Fehlerfall stirbt in Millisekunden;
# 5s sind reichlich Reserve und halten den Gate-Lauf kurz.
GRACE="${GRACE:-5}"

fail=0
for img in $IMAGES; do
  if ! cid=$(docker run -d "$img" 2>/dev/null); then
    echo "[image-start-check] FAIL $img — docker run schlug fehl"
    fail=1
    continue
  fi

  sleep "$GRACE"

  status=$(docker inspect -f '{{.State.Status}}' "$cid" 2>/dev/null || echo "weg")
  if [ "$status" = "running" ]; then
    echo "[image-start-check] OK   $img — laeuft nach ${GRACE}s"
  else
    code=$(docker inspect -f '{{.State.ExitCode}}' "$cid" 2>/dev/null || echo "?")
    echo "[image-start-check] FAIL $img — beendet innerhalb ${GRACE}s (Status $status, Exit $code)"
    echo "[image-start-check]      letzte Container-Logs:"
    docker logs "$cid" 2>&1 | tail -8 | sed 's/^/[image-start-check]      | /'
    fail=1
  fi

  docker rm -f "$cid" >/dev/null 2>&1 || true
done

if [ "$fail" -ne 0 ]; then
  echo "[image-start-check] FEHLGESCHLAGEN — mindestens ein Runtime-Image startet nicht."
  exit 1
fi
echo "[image-start-check] alle Images gestartet."
