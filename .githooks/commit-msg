#!/usr/bin/env bash
# commit-msg — der git-eigene Traeger der Traceability-Zusage dieses Repos.
# git uebergibt den Pfad der vorgeschlagenen Message-Datei als $1; Exit ungleich
# 0 bricht den Commit ab.
#
# WARUM ER AM COMMIT HAENGT UND NICHT AM AGENTEN: er sieht jede Commit-Klasse,
# die git selbst erzeugt — auch die aus einem Repo-Werkzeug und die `-m`-Form,
# die kein Kommando-Zeilen-Matcher zuverlaessig erkennt. Die Pruefung liegt in
# tools/harness/commit-msg-traceability.sh; dieser Aufruf reicht sie weiter.
#
# AKTIVIERUNG. `core.hooksPath` ist lokale Konfiguration und reist nicht mit dem
# Klon; `make hooks-install` setzt sie. Ein Klon ohne diesen Aufruf ist
# ungeprueft, und `git commit --no-verify` umgeht diesen Traeger auch danach.
# Beides steht in harness/mk/hooks-install.mk.
#
# GEBAUT WIE DER COMMIT-PFAD ES VERLANGT: bash und coreutils, kein Docker, kein
# Netz — ein Commit ohne laufenden Daemon bricht nicht.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec bash "$here/../tools/harness/commit-msg-traceability.sh" "$@"
