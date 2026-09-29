# harness/mk/traeger.mk — Fragment des Traeger-Fetch, emittiert von ai-harness-init.
# EIN KOMMANDO, KEIN GATE.
#
# Es liegt im Fragment-Verzeichnis wie die Gate-Fragmente, haengt aber NICHTS an
# GATE_CHECKS und steht in keiner Prerequisite-Kette — auch nicht an archive-welle:
# dessen Fehlt-Fall ist die Zusage "Exit 0, nennt das Fehlende, schreibt nichts", und
# ein Fetch davor wuerde sie still aendern (ADR-0058 Festlegung 3 und Folgepflicht 5).
# Der Fetch ist der AUSDRUECKLICHE Weg aus dem Zustand, den jene Meldung benennt.
#
# Der Traeger liegt im gitignorierten Zustands-Bereich: ein frischer Klon hat ihn
# nicht. Dieses Kommando holt ihn aus dem gepinnten Release nach — das Asset wird
# gegen den SHA256SUMS-Eintrag desselben Releases verifiziert VOR der Ablage, eine
# Abweichung bricht ab, ohne den Traeger zu legen (ADR-0059 Festlegung 1); der
# Transport laeuft im gepinnten Bild, nicht auf dem Host (LH-QA-03). Netz braucht
# er an genau diesem Aufruf (MR-007-Muster); er laeuft nur auf ausdruecklichem Aufruf.
.PHONY: traeger-fetch

# DER PIN (ADR-0058 Festlegung 1, Dogfood-Haeifte; ADR-0059 Festlegung 2): das
# Fragment fuehrt nur den Release-Tag — keinen Wert, der vom Bau-Ergebnis abhaengt
# (die Digest-Variablen und ihr Export sind mit ADR-0059 Folgepflicht 3 entfallen).
# Die Verifizierung liest der Fetch aus der SHA256SUMS desselben Releases; das
# Dogfood-Makefile fuehrt daneben die sechs Einzeldigests (zwei Kanaele,
# ADR-0059 Festlegung 3). Der Tag ist eine Release-Entscheidung: er wandert mit dem
# Release-Schnitt, nicht mit jedem Bau.
TRAEGER_TAG ?= v0.2.5
TRAEGER_CARRIER ?= .harness/state/bin/ai-harness-init

export TRAEGER_TAG TRAEGER_CARRIER

traeger-fetch: ## Traeger aus dem gepinnten Release nachholen (braucht Netz) — NICHT in gates
	@bash tools/harness/traeger-fetch.sh
