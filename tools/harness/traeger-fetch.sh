#!/usr/bin/env bash
# traeger-fetch.sh — legt den Traeger per Fetch aus dem gepinnten Release ab
# (ADR-0058 Festlegung 1 und 3; Verifizierung gegen die SHA256SUMS des Releases,
# ADR-0059 Festlegung 1). EIN KOMMANDO, KEIN GATE: es prueft nichts am Baum,
# haengt an keiner gates-Kette und steht in keiner Prerequisite-Kette — der Fehlt-Fall
# der Konsumenten (archive-welle, span-report u. a.) bleibt unangetastet
# (ADR-0058 Festlegung 3 und Folgepflicht 5). Es braucht Netz an genau diesem Aufruf
# (MR-007-Muster) und laeuft nur auf ausdruecklichem Aufruf, nie nebenbei.
#
# DER TRANSPORT LAEUFT IM GEPINNTEN BILD, NICHT AUF DEM HOST (ADR-0058 Festlegung 4):
# Download und Digest-Verifizierung fahren in einem Container mit digest-gepinntem
# Image — der Host braucht weiter nur git, docker, make (LH-QA-03); ein curl/wget in
# der Befehlsposition waere eine vierte Abhaengigkeit.
#
# FAIL-CLOSED, alle VOR der Ablage: fehlt der Release-Pin, fehlt bei TEILWEISE
# exportierten Digest-Pins der eigene der Plattform (die Kopplung hat eine Stelle
# stehen lassen — dieselbe Klasse wie test/sources-pin.bats), fehlt das Asset im
# Manifest, oder weicht der Digest des geholten Assets ab, bricht der Lauf, ohne den
# Traeger abzulegen (LH-QA-02).
#
# WOHER DER ERWARTETE DIGEST KOMMT (ADR-0059 Festlegung 1 und 3): sind
# TRAEGER_SHA256_*-Variablen exportiert, verifiziert der Lauf gegen den eigenen
# Makefile-Pin — zwei Kanaele, das Makefile reist in git, das Asset ueber den
# Release-Kanal (Dogfood-Haeifte von ADR-0058 Festlegung 1). Ist KEINE gesetzt,
# laedt der Lauf die SHA256SUMS desselben Releases und verifiziert gegen den
# Manifest-Eintrag des gewaehlten Assets — Manifest und Asset reisen ueber denselben
# Kanal (die Grenze: ein gemeinsamer Ersatz beider geht durch, kein Signier-Schritt).
# Die Teilweise-exportierte Kopplung bricht, statt still in den Manifest-Kanal zu
# fallen. Dieses Skript fuehrt keinen zweiten Bestand der Pins;
# test/traeger-fetch.bats haelt beide Stellen gegen das Makefile-Paar.
set -euo pipefail

# Transport-Bild, digest-gepinnt (LH-QA-02). Der Einzige Pin dieses Skripts: seine
# Konsumenten sind hier und im emittierten Zwilling (byte-gleich, test/traeger-fetch.bats).
TRAEGER_IMAGE="${TRAEGER_IMAGE:-curlimages/curl@sha256:463eaf6072688fe96ac64fa623fe73e1dbe25d8ad6c34404a669ad3ce1f104b6}"
# Ablageort des Traegers — derselbe gitignorierte Zustands-Bereich, den placeCarrier
# des Werkzeugs beschreibt (carrierDir/carrierName, internal/emit/enforce.go).
TRAEGER_CARRIER="${TRAEGER_CARRIER:-.harness/state/bin/ai-harness-init}"

case "$TRAEGER_IMAGE" in
*@sha256:*) ;;
*)
	echo "traeger-fetch: TRAEGER_IMAGE ist nicht digest-gepinnt ($TRAEGER_IMAGE) — der Transport laeuft im gepinnten Bild (ADR-0058 Festlegung 4, LH-QA-02)." >&2
	exit 2
	;;
esac

# Plattform des HOSTS, nicht des Bilds: das Asset muss zur Maschine passen, die den
# Traeger startet. Closed Set je Achse — die Asset-Matrix der Release fuehrt
# linux/darwin/windows x amd64/arm64 (LH-QA-04); alles andere bricht laut, statt ein
# Asset zu raten.
os="${TRAEGER_OS:-$(uname -s)}"
arch="${TRAEGER_ARCH:-$(uname -m)}"
case "$os" in
Linux) plat="linux" ;;
Darwin) plat="darwin" ;;
MINGW* | MSYS* | CYGWIN*) plat="windows" ;;
*)
	echo "traeger-fetch: unbekannte Plattform (uname -s: $os) — die Asset-Matrix fuehrt linux, darwin, windows (LH-QA-04)." >&2
	exit 2
	;;
esac
case "$arch" in
x86_64 | amd64) a="amd64" ;;
aarch64 | arm64) a="arm64" ;;
*)
	echo "traeger-fetch: unbekannte Architektur (uname -m: $arch) — die Asset-Matrix fuehrt amd64, arm64 (LH-QA-04)." >&2
	exit 2
	;;
esac
plat_u="$(printf '%s' "$plat" | tr '[:lower:]' '[:upper:]')"
arch_u="$(printf '%s' "$a" | tr '[:lower:]' '[:upper:]')"

asset="ai-harness-init-${plat}-${a}"
carrier="$TRAEGER_CARRIER"
if [ "$plat" = "windows" ]; then
	asset="${asset}.exe"
	case "$carrier" in
	*.exe) ;;
	*) carrier="${carrier}.exe" ;;
	esac
fi

# Der erwartete Digest: eigener Pin der Plattform, oder Manifest-Eintrag. Sind
# Digest-Pins TEILWEISE exportiert und fehlt der eigene, ist eine Pin-Stelle stehen
# geblieben — der Lauf bricht, BEVOR er anfaengt, statt still gegen das Manifest zu
# verifizieren (ADR-0059 Festlegung 3: die Dogfood-Haelfte bleibt am Makefile-Pin).
sha_var="TRAEGER_SHA256_${plat_u}_${arch_u}"
sha="${!sha_var:-}"
teilweise=""
for p in LINUX_AMD64 LINUX_ARM64 DARWIN_AMD64 DARWIN_ARM64 WINDOWS_AMD64 WINDOWS_ARM64; do
	name="TRAEGER_SHA256_$p"
	if [ -n "${!name:-}" ]; then teilweise=ja; fi
done
if [ -z "$sha" ] && [ -n "$teilweise" ]; then
	echo "traeger-fetch: $sha_var ist nicht gesetzt, aber andere Digest-Pins sind es — eine Pin-Stelle der Kopplung fehlt; der Lauf bricht, statt still gegen das Manifest zu verifizieren (ADR-0058 Festlegung 1, ADR-0059 Festlegung 3, LH-QA-02)." >&2
	exit 2
fi
tag="${TRAEGER_TAG:-}"
if [ -z "$tag" ]; then
	echo "traeger-fetch: TRAEGER_TAG ist nicht gesetzt — der Release-Pin fehlt (ADR-0058 Festlegung 1, LH-QA-02)." >&2
	exit 2
fi
url="https://github.com/pt9912/ai-harness-init/releases/download/${tag}/${asset}"
sums_url="https://github.com/pt9912/ai-harness-init/releases/download/${tag}/SHA256SUMS"

mkdir -p "$(dirname "$carrier")"
carrier_abs="$(cd "$(dirname "$carrier")" && pwd)/$(basename "$carrier")"

# DER TRANSPORT IM BILD. Das Bild mountet den Ablage-Ordner an DERSELBEN absoluten
# Adresse, unter der der Host ihn liest — das Payload referenziert genau diesen Pfad.
# --user haelt die Ablage im Besitz des Host-Nutzers; das Bild laeuft sonst als
# eigener Nutzer und legte fremde Dateien an. Die Plattform-Erkennung steht bewusst
# AUSSERHALB des Bilds: uname im Container nennt die Plattform des Bilds, nicht die
# des Hosts.
#
# Verifiziert wird VOR der Ablage: eine Digest-Abweichung bricht ab, ohne den Traeger
# zu legen — unter der geschwächten Zusicherung (Abweichung bricht, Traeger bleibt
# liegen) bleibt der Negative-Fall von test/traeger-fetch.bats rot. Im Manifest-Modus
# laedt das Payload SHA256SUMS und den Asset-Eintrag daraus; ein fehlender Eintrag
# bricht laut und legt nichts ab (ADR-0059 Festlegung 1).
payload="$(cat <<'ENDE'
set -eu
erwartet="$TRAEGER_SHA256"
quelle="Pin"
if [ -z "$erwartet" ]; then
	curl -fsSL -o /tmp/traeger-sums "$TRAEGER_SUMS_URL"
	erwartet="$(awk -v a="$TRAEGER_ASSET" '$2 == a { print $1; exit }' /tmp/traeger-sums)"
	if [ -z "$erwartet" ]; then
		echo "traeger-fetch: $TRAEGER_ASSET fehlt im Manifest (SHA256SUMS) — der Manifest-Eintrag seiner Plattform fehlt; der Lauf legt nichts ab (ADR-0059 Festlegung 1, LH-QA-04)." >&2
		exit 1
	fi
	quelle="SHA256SUMS"
fi
curl -fsSL -o /tmp/traeger-asset "$TRAEGER_URL"
ist="$(sha256sum /tmp/traeger-asset)" && ist="${ist%% *}"
if [ "$ist" != "$erwartet" ]; then
	echo "traeger-fetch: Digest-Abweichung — ist $ist, erwartet $erwartet (aus $quelle). Der Traeger wird nicht abgelegt." >&2
	exit 1
fi
cp /tmp/traeger-asset "$TRAEGER_CARRIER_ABS"
chmod 0755 "$TRAEGER_CARRIER_ABS"
ENDE
)"

docker run --rm \
	--user "$(id -u):$(id -g)" \
	-e TRAEGER_URL="$url" \
	-e TRAEGER_SHA256="$sha" \
	-e TRAEGER_SUMS_URL="$sums_url" \
	-e TRAEGER_ASSET="$asset" \
	-e TRAEGER_CARRIER_ABS="$carrier_abs" \
	-v "$(dirname "$carrier_abs"):$(dirname "$carrier_abs")" \
	"$TRAEGER_IMAGE" sh -c "$payload"

echo "traeger-fetch: Traeger abgelegt ($carrier_abs) — $asset aus Release $tag, Digest verifiziert."

