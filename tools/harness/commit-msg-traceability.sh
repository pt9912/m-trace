#!/usr/bin/env bash
# commit-msg-traceability.sh — prueft eine Commit-Message-DATEI gegen die
# Traceability-Zusage des mitgelieferten Regelwerks. Der Aufrufer gibt den Pfad
# der vorgeschlagenen Message-Datei als $1; Exit ungleich 0 bricht den Commit ab.
#
# ZUSAGE. Exit 0, wenn die Message mindestens eine Kennung aus der Menge
# {ADR-, LH-, MR-, slice-} traegt oder ihr Betreff mit "Merge " bzw. "Revert "
# beginnt; Exit 1, wenn keines von beidem zutrifft; Exit 2, wenn die Datei fehlt
# oder nicht lesbar ist. Betreff ist die erste nicht-leere Zeile ohne
# Kommentarzeichen; die Kennung darf auch im Rumpf stehen, eine Kommentarzeile
# zaehlt dagegen nicht.
#
# GRENZE. Geprueft wird die ANWESENHEIT einer Kennung, nicht ihre Wahrheit: eine
# Message, die zusaetzlich einen nicht aufloesbaren Verweis nennt, geht mit
# derselben Kennung durch. Die Menge ist die des mitgelieferten Regelwerks und
# steht in der Zeile `patterns=` unten — als Muster steht sie nur dort: der Kopf
# nennt die Klassen in Worten, die Fehlermeldung zeigt auf die Zeile, und ein
# erneuter Bootstrap schreibt diese Datei kanonisch neu. Ein Repo mit einer
# eigenen Kennungs-Klasse setzt darum HOOKS_DIR auf sein eigenes Hook-Verzeichnis
# (harness/mk/hooks-install.mk) und fuehrt dort seinen Traeger.
# Als Kommentarzeile gilt die fuehrende Raute; ein abweichendes core.commentChar
# liest dieses Skript nicht.
#
# ABGRENZUNG. Kein Gate und kein `make`-Ziel: der Aufruf steht im Commit-Pfad und
# setzt darum nichts voraus, was der Host nicht hat — kein Docker, kein Netz, nur
# bash und coreutils. Der Aufrufer ist der git-eigene Traeger .githooks/commit-msg,
# aktiviert per `make hooks-install`.
set -euo pipefail

msg_file="${1:-}"
if [ -z "$msg_file" ]; then
  echo "commit-msg-traceability: keine Message-Datei uebergeben (Aufruf: $0 <datei>)." >&2
  exit 2
fi
if [ ! -r "$msg_file" ]; then
  echo "commit-msg-traceability: '$msg_file' ist nicht lesbar." >&2
  exit 2
fi

# Betreff: erste Zeile, die weder leer noch eine Kommentarzeile ist, ohne
# fuehrende Leerzeichen. An ihr haengt die Merge-/Revert-Ausnahme.
subject=""
while IFS= read -r line || [ -n "$line" ]; do
  trimmed="${line#"${line%%[![:space:]]*}"}"
  [ -n "$trimmed" ] || continue
  case "$trimmed" in
    '#'*) continue ;;
  esac
  subject="$trimmed"
  break
done < "$msg_file"

exempt='^(Merge |Revert )'
if [[ "$subject" =~ $exempt ]]; then
  exit 0
fi

# Kennung: ERE ueber die ganze Datei, Kommentarzeilen ausgenommen. Der Dialekt
# ist POSIX-ERE ([0-9] statt \d) — bash kennt die \d-Kurzform nicht.
patterns='(ADR-[0-9]{4}|LH-[A-Z]{2}-[0-9]{2}|MR-[0-9]{3}|slice-[0-9]+)'
while IFS= read -r line || [ -n "$line" ]; do
  trimmed="${line#"${line%%[![:space:]]*}"}"
  case "$trimmed" in
    '#'*) continue ;;
  esac
  if [[ "$line" =~ $patterns ]]; then
    exit 0
  fi
done < "$msg_file"

echo "commit-msg-traceability: keine Traceability-Kennung in der Commit-Message:" >&2
echo "            ${subject}" >&2
echo "            Erwartet wird eine Kennung aus der Menge in der Zeile \`patterns=\` dieser Pruefung." >&2
exit 1
