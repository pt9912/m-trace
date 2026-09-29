#!/usr/bin/env bash
# tools/harness/selbstpruefung.sh — Selbstpruefung der emittierten
# Durchsetzungsschicht, emittiert von ai-harness-init. EIN KOMMANDO, KEIN GATE.
#
# WAS SIE BELEGT. Dieses Repo klont sich selbst in ein Wegwerf-Verzeichnis,
# stellt fest, dass der frische Klon in seiner LOKALEN Konfiguration keinen
# core.hooksPath traegt, aktiviert dort den Traeger der Commit-Kennung und
# faehrt danach ZWEI Commit-Versuche: einer OHNE Kennung faellt und entsteht
# nicht, einer MIT Kennung geht durch. Zuletzt laeuft das Gate-Kommando im
# Klon. BEIDE AUSGAENGE STEHEN IN EINEM LAUF — ein Lauf, der nur den
# durchgelassenen Commit beobachtet, belegt nicht, dass der Traeger ueberhaupt
# etwas aufhaelt.
#
# WARUM EIN KLON UND NICHT DIESES ARBEITSVERZEICHNIS. Die Aktivierung ist
# lokale Konfiguration (core.hooksPath) und reist nicht mit dem Klon; genau
# dieser Zustand ist der Pruefgegenstand. Und die zwei Commit-Versuche
# schreiben Historie — im Klon ist sie nach dem Lauf weg.
#
# FUENF MARKER, JE MIT BELEGUNG. Sie sind Variablen, keine Platzhalter zum
# Suchen-und-Ersetzen:
#   SELBSTPRUEFUNG_TRAEGER      der Hook, den der Aktivierungsschritt in Betrieb nimmt
#   SELBSTPRUEFUNG_AKTIVIERUNG  der Schritt, der ihn in Betrieb nimmt
#   SELBSTPRUEFUNG_GATE         das Kommando, das im Klon gruen laufen muss
#   SELBSTPRUEFUNG_MSG_ROT      die Message, die am Traeger fallen muss
#   SELBSTPRUEFUNG_MSG_GRUEN    die Message, die durchgehen muss
# Der Lauf nennt in seiner ersten Zeile die Werte, mit denen er faehrt, und er
# prueft den Traeger-Marker gegen den Hook, den git nach der Aktivierung
# wirklich ruft — ein Marker, der eine andere Datei nennt, bricht den Lauf ab.
#
# DIESE DATEI IST KONVERGENT (ADR-0007 Festlegung 3): jeder Lauf des Werkzeugs
# schreibt sie kanonisch neu. Ein Edit an ihr ist danach still weg — deshalb
# sind die fuenf Stellen oben Variablen. DASSELBE GILT FUER DAS ROOT-Makefile:
# es ist der generierte Aggregator und wird ebenso kanonisch neu geschrieben;
# eine Vorgabe darin ueberlebt den naechsten Bootstrap nicht.
#
# WO EINE DAUERHAFTE VORGABE TRAEGT: in einem EIGENEN Fragment unter
# harness/mk/ mit einem Namen, den dieses Werkzeug nicht schreibt — der
# Aggregator bindet es ueber `include harness/mk/*.mk` mit ein, und kein Lauf
# entfernt, was er nicht selbst angelegt hat. Vorgeschlagen:
# harness/mk/vorgaben.mk mit einer Zeile je Marker, in der Form
#   SELBSTPRUEFUNG_GATE = <kommando>
# Das einfache `=` gewinnt unabhaengig von der Include-Reihenfolge: steht es
# vorher, greift das `?=` daneben nicht mehr; steht es nachher, ueberschreibt
# es dessen Belegung. Wer keinen dauerhaften Ort braucht, setzt am Aufruf.
# Der Traeger unter .githooks/ ist der eine Pfad, den dieses Werkzeug an einen
# belegten Ort nicht schreibt (skip-if-present, ADR-0054): er gehoert dem Repo.
#
# DIE GRENZE. Geprueft sind der Traeger und die zwei Commit-Ausgaenge. NICHT
# geprueft ist, ob der Traeger jeden Commit-Pfad erreicht: `git commit
# --no-verify` umgeht ihn, Commits aus Repo-Werkzeugen tragen ihre eigenen
# Messages, und Aufrufformen ausserhalb der aktivierten Traeger-Form bleiben
# ausserhalb der Zusage. Geprueft ist ferner die ANWESENHEIT einer Kennung in
# der Message, nicht ihre Wahrheit.
#
# EIN REPO MIT EIGENEM TRAEGER passt die zwei Message-Marker an. Die Belegung
# unten gilt dem Traeger, den dieses Werkzeug mitbringt, und der Kennungs-Menge
# der Pruefung daneben (Zeile `patterns=`). Fuehrt das Repo an .githooks/ seinen
# eigenen Traeger — der Zustand, den ADR-0054 Festlegung 1 ihm freistellt —,
# nennt es mit SELBSTPRUEFUNG_MSG_ROT und SELBSTPRUEFUNG_MSG_GRUEN je eine
# Message, die sein Traeger aufhaelt bzw. durchlaesst.
#
# VORAUSSETZUNG. Ein git-Repo mit mindestens einem Commit. Geklont wird die
# Wurzel, die `git rev-parse --show-toplevel` von hier aus liefert — liegt
# dieses Repo als Unterverzeichnis in einem umgebenden, ist das die AEUSSERE
# Wurzel. Der Klon traegt den Stand von HEAD; nicht committete Aenderungen des
# Arbeitsverzeichnisses sind darin nicht enthalten. Die Identitaet der zwei
# Commit-Versuche bringt der Lauf selbst mit (git -c user.email/user.name),
# damit die Pruefung auch auf einem Rechner ohne konfigurierte Identitaet
# laeuft.
#
# ABHAENGIGKEIT. git, make und coreutils. Kein Netz, kein Paketmanager, kein
# zweites Bild — was das Gate-Kommando seinerseits braucht, bringt es selbst
# mit.
set -euo pipefail

# Der Traeger, den der Klon fuehrt und den der Aktivierungsschritt in Betrieb
# nimmt. Der Name gehoert git, das Verzeichnis dem Repo.
SELBSTPRUEFUNG_TRAEGER="${SELBSTPRUEFUNG_TRAEGER:-.githooks/commit-msg}"
# Der eine Schritt zwischen liegendem Traeger und wirksamem Traeger.
SELBSTPRUEFUNG_AKTIVIERUNG="${SELBSTPRUEFUNG_AKTIVIERUNG:-make hooks-install}"
# Das Kommando, das im Klon gruen laufen muss. Ein Ziel mit anderem Bau- oder
# Sprachmodell setzt hier seines.
SELBSTPRUEFUNG_GATE="${SELBSTPRUEFUNG_GATE:-make gates}"
# Die zwei Commit-Messages. Sie gehoeren zum Traeger, nicht zum Lauf: ein
# eigener Traeger bringt eine eigene Kennungs-Menge mit.
SELBSTPRUEFUNG_MSG_ROT="${SELBSTPRUEFUNG_MSG_ROT:-Selbstpruefung ohne Kennung}"
SELBSTPRUEFUNG_MSG_GRUEN="${SELBSTPRUEFUNG_MSG_GRUEN:-Selbstpruefung mit Kennung LH-FA-01}"

fehler() {
	echo "selbstpruefung: FEHLER — $1" >&2
	exit 1
}

# e2e_abdeckung <Kennungen> <Kurzbeschreibung> <Anker> — DIE DEKLARATION EINER STUFE.
#
# WOZU: Der Aufruf gibt der Beziehung ANFORDERUNG -> STUFE einen Ort. Er steht IN der
# Stufe, die er deklariert, und der Anker ist ein woertlicher Ausschnitt aus einer
# ANDEREN Zeile dieser Stufe. Aus diesen drei Argumenten schreibt
# `make e2e-abdeckung` (tools/harness/e2e-abdeckung.sh) die Abdeckungs-Sicht dieses
# Repos.
#
# DIE STUFEN-MENGE IST EIN KRITERIUM, KEINE AUFZAEHLUNG: eine Zeile
# `echo "selbstpruefung: … ..."` eroeffnet eine Stufe, ihre Region reicht bis zur
# naechsten solchen. Ein eigenes E2E dieses Repos deklariert seine Stufen genauso und
# nennt sich dem Erzeuger ueber die Marker E2E_ABDECKUNG_QUELLE und
# E2E_ABDECKUNG_PRAEFIX.
#
# DAS ERSTE ARGUMENT IST HIER EIN GEDANKENSTRICH: diese Stufe kommt mit dem Werkzeug,
# und welche Anforderung DIESES Repos sie traegt, weiss das Werkzeug nicht. Eine
# geratene Kennung loeste vielleicht auf und behauptete trotzdem eine Zuordnung, die
# niemand getroffen hat; der Gedankenstrich sagt, dass keine getroffen ist. Der
# Erzeuger traegt ihn in die Zelle und nennt die Stufe.
#
# WAS DER AUFRUF PRUEFT: nichts. Die zwei Luecken-Richtungen — Stufe ohne Deklaration,
# Deklaration ohne aufloesenden Anker — prueft der Erzeuger ueber dem TEXT dieser Datei,
# und nur er sieht auch eine Stufe, die gar keinen Aufruf mehr fuehrt.
e2e_abdeckung() {
	echo "selbstpruefung: Abdeckung dieser Stufe — Kennung(en): $1 — $2 (Anker: $3)"
}

command -v git >/dev/null 2>&1 || fehler "git liegt nicht im Pfad — ohne git gibt es keinen Klon, auf dem ein Traeger etwas aufhalten koennte."

quelle="$(git rev-parse --show-toplevel 2>/dev/null)" || fehler "dieses Verzeichnis liegt in keinem git-Repo — die Pruefung klont das Repo, in dem sie laeuft."
git -C "$quelle" rev-parse --verify --quiet 'HEAD^{commit}' >/dev/null 2>&1 ||
	fehler "$quelle traegt keinen Commit — ein Klon haette keinen Baum, auf dem ein Commit-Versuch etwas bedeutete. Ein erster Commit macht die Pruefung fahrbar."

echo "selbstpruefung: Marker — Traeger=[$SELBSTPRUEFUNG_TRAEGER] Aktivierung=[$SELBSTPRUEFUNG_AKTIVIERUNG] Gate=[$SELBSTPRUEFUNG_GATE] MsgRot=[$SELBSTPRUEFUNG_MSG_ROT] MsgGruen=[$SELBSTPRUEFUNG_MSG_GRUEN]"
echo "selbstpruefung: Quelle=[$quelle] (der Klon traegt den Stand von HEAD)"

echo "selbstpruefung: Traeger, Aktivierung, die zwei Commit-Ausgaenge und das Gate-Kommando im frischen Klon ..."
e2e_abdeckung "—" "Der Traeger der Commit-Kennung greift im frischen Klon, und das Gate-Kommando laeuft dort gruen" "der frische Klon traegt lokal keinen core.hooksPath"

arbeit="$(mktemp -d -p "${TMPDIR:-/tmp}")"
trap 'rm -rf "$arbeit"' EXIT
# mktemp -d liefert 0700. Laeuft das Gate-Kommando in einem Container als
# Nicht-Root, kann er einen 0700-Pfad nicht traversieren.
chmod 755 "$arbeit"
klon="$arbeit/klon"

# `file://` ist Pflicht: ueber einen blossen Pfad legt git einen Klon mit
# geteilten Objekten an, und der haengt am Quell-Repo statt fuer sich zu stehen.
git clone -q "file://$quelle" "$klon" || fehler "der Klon von $quelle ist nicht entstanden."
chmod 755 "$klon"

# (1) DER FRISCHE KLON IST UNGEPRUEFT. Gelesen wird die LOKALE Konfiguration
# des Klons — sie ist es, die ein Klon nicht erbt. `--get` ohne Scope naehme
# auch eine globale oder systemweite Setzung mit und meldete einen Zustand des
# Rechners als einen des Klons.
vorher_hooks="$(git -C "$klon" config --get --local core.hooksPath || true)"
if [ -n "$vorher_hooks" ]; then
	fehler "der frische Klon traegt in seiner lokalen Konfiguration bereits core.hooksPath=$vorher_hooks — dann misst der Lauf nicht, was die Aktivierung bewirkt."
fi
echo "selbstpruefung: der frische Klon traegt lokal keinen core.hooksPath — der Traeger reist mit, seine Aktivierung nicht."

if [ ! -f "$klon/$SELBSTPRUEFUNG_TRAEGER" ]; then
	fehler "im Klon liegt kein Traeger unter [$SELBSTPRUEFUNG_TRAEGER] — er ist entweder nicht committet oder liegt woanders; der Marker SELBSTPRUEFUNG_TRAEGER nennt den Pfad."
fi

# (2) DIE AKTIVIERUNG. Gelesen wird der Wert, den git danach lokal fuehrt.
akt_out=""
if ! akt_out="$(cd "$klon" && bash -c "$SELBSTPRUEFUNG_AKTIVIERUNG" 2>&1)"; then
	printf '%s\n' "$akt_out" >&2
	fehler "der Aktivierungsschritt [$SELBSTPRUEFUNG_AKTIVIERUNG] endet nicht mit Exit 0 — der Traeger ist im Klon nicht in Betrieb zu nehmen."
fi
printf '%s\n' "$akt_out" | sed 's/^/selbstpruefung:   /'
nachher_hooks="$(git -C "$klon" config --get --local core.hooksPath || true)"
if [ -z "$nachher_hooks" ]; then
	fehler "nach [$SELBSTPRUEFUNG_AKTIVIERUNG] ist core.hooksPath im Klon weiter leer — der Schritt meldet Erfolg, ohne einen zu haben."
fi
echo "selbstpruefung: aktiviert — core.hooksPath=$nachher_hooks"

# (2b) DER TRAEGER-MARKER NENNT DIE DATEI, DIE AUFHAELT — nicht irgendeine.
# git ruft einen Hook ueber seinen nackten Namen aus core.hooksPath; welche
# Datei das ist, steht damit fest. Ohne diesen Abgleich lenkte der Marker nur
# eine Existenzpruefung, und die Schluss-Zeile schriebe ihm eine Wirkung zu,
# die eine andere Datei hatte.
hook_name="${SELBSTPRUEFUNG_TRAEGER##*/}"
case "$nachher_hooks" in
/*) gerufen="$nachher_hooks/$hook_name" ;;
*) gerufen="$klon/$nachher_hooks/$hook_name" ;;
esac
if [ ! -e "$gerufen" ] || [ ! "$gerufen" -ef "$klon/$SELBSTPRUEFUNG_TRAEGER" ]; then
	fehler "[$SELBSTPRUEFUNG_TRAEGER] ist nicht der Traeger, den der Aktivierungsschritt in Betrieb nimmt: [$SELBSTPRUEFUNG_AKTIVIERUNG] setzt core.hooksPath=$nachher_hooks, und git ruft daraus $gerufen. Der Marker nennt damit eine andere Datei, als aufhaelt."
fi
echo "selbstpruefung: in Betrieb ist [$SELBSTPRUEFUNG_TRAEGER] — git ruft ihn als $nachher_hooks/$hook_name."

# (3) DIE ZWEI COMMIT-AUSGAENGE, in einem Lauf. `--allow-empty` haelt jeden
# Versuch ohne Baum-Aenderung; gelesen werden Exit-Code UND die Lage von HEAD,
# denn eine Meldung allein belegt keinen Abbruch.
vorher_head="$(git -C "$klon" rev-parse HEAD)"
for fall in rot gruen; do
	case "$fall" in
	rot) msg="$SELBSTPRUEFUNG_MSG_ROT" ;;
	gruen) msg="$SELBSTPRUEFUNG_MSG_GRUEN" ;;
	*) fehler "unbekannter Fall [$fall]." ;;
	esac
	out=""
	rc=0
	out="$(git -C "$klon" -c user.email=selbstpruefung@example.invalid -c user.name=selbstpruefung \
		commit -q --allow-empty -m "$msg" 2>&1)" || rc=$?
	if [ "$fall" = rot ]; then
		if [ "$rc" -eq 0 ]; then
			fehler "der Commit '$msg' geht durch — er sollte am Traeger [$SELBSTPRUEFUNG_TRAEGER] fallen. Fuehrt dieses Repo dort seinen EIGENEN Traeger (skip-if-present, ADR-0054 Festlegung 1), nennt SELBSTPRUEFUNG_MSG_ROT eine Message, die jener aufhaelt."
		fi
		nun_head="$(git -C "$klon" rev-parse HEAD)"
		if [ "$nun_head" != "$vorher_head" ]; then
			fehler "der Traeger meldet den Abbruch, der Commit ist aber entstanden (HEAD bewegt sich von $vorher_head nach $nun_head)."
		fi
		echo "selbstpruefung: ROT — die Message [$msg] faellt am Traeger (Exit $rc) und HEAD steht unveraendert. Meldung des Traegers:"
		printf '%s\n' "$out" | sed -n '1,3p' | sed 's/^/selbstpruefung:   /'
	else
		if [ "$rc" -ne 0 ]; then
			printf '%s\n' "$out" >&2
			fehler "der Commit '$msg' faellt am Traeger (Exit $rc) — erwartet war ein Durchgang. Fuehrt dieses Repo an [$SELBSTPRUEFUNG_TRAEGER] seinen EIGENEN Traeger, nennt SELBSTPRUEFUNG_MSG_GRUEN eine Message, die jener durchlaesst."
		fi
		betreff="$(git -C "$klon" log -1 --format=%s)"
		if [ "$betreff" != "$msg" ]; then
			fehler "der Commit '$msg' meldet Erfolg, HEAD traegt aber '$betreff' — der Durchgang ist nicht belegt."
		fi
		echo "selbstpruefung: GRUEN — die Message [$msg] geht durch, HEAD traegt: $betreff"
	fi
done

# (4) DAS GATE-KOMMANDO IM KLON. Es laeuft zuletzt, weil die Schritte davor
# Historie schreiben und der Klon danach wegfaellt.
gate_rc=0
gate_out="$(cd "$klon" && bash -c "$SELBSTPRUEFUNG_GATE" 2>&1)" || gate_rc=$?
if [ "$gate_rc" -ne 0 ]; then
	printf '%s\n' "$gate_out" >&2
	fehler "[$SELBSTPRUEFUNG_GATE] im Klon endet mit Exit $gate_rc — der Klon traegt den Stand von HEAD, und der ist damit nicht gruen."
fi
# DIE GANZE AUSGABE, nicht ihre letzte Zeile: was ein Kommando hinterlaesst, ist
# das einzige, woran von aussen abzulesen ist, WELCHES lief. Eine einzelne Zeile
# stammt aus einem fremden Werkzeug und traegt diese Unterscheidung nicht. Der
# Voll-E2E des Werkzeugs greift dafuer in seiner Selbstpruefungs-Stufe zwei Spuren
# aus zwei Kommandos der Kette, die auf verschiedenen Zeilen stehen.
echo "selbstpruefung: GATE — [$SELBSTPRUEFUNG_GATE] im Klon ist Exit 0. Ausgabe:"
printf '%s\n' "$gate_out" | sed 's/^/selbstpruefung:   /'
echo "selbstpruefung: OK — der Traeger [$SELBSTPRUEFUNG_TRAEGER] reist mit dem Klon, seine Aktivierung nicht; [$SELBSTPRUEFUNG_AKTIVIERUNG] nimmt genau ihn in Betrieb, danach faellt '$SELBSTPRUEFUNG_MSG_ROT' und geht '$SELBSTPRUEFUNG_MSG_GRUEN' durch, und [$SELBSTPRUEFUNG_GATE] laeuft im Klon gruen. Nicht geprueft: ob der Traeger jeden Commit-Pfad erreicht (--no-verify, Werkzeug-Commits, andere Aufrufformen)."
