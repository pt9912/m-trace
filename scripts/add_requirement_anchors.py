#!/usr/bin/env python3
"""slice-001: Inline-Anker für Requirement-Definitionen in einer Tabellen-Quelle.

Fügt in jede Tabellen-Definitionszeile (erste Zelle = bare Kennung) einen
`<a id="<kennung-klein>"></a>`-Anker direkt hinter der Kennung ein. Idempotent
(überspringt bereits verankerte Zellen). Additiv — kein Text-/Modalitäts-Edit.
"""
import re
import sys
from pathlib import Path

# Kennungs-Familien je Quelle (MR-003).
FAMILIES = {
    "spec/lastenheft.md": r"(?:F|NF|MVP|AK|RAK)-\d+",
    "docs/plan/planning/in-progress/risks-backlog.md": r"R-\d+",
}


def anchor_file(path: Path, id_regex: str) -> int:
    # Definitionszeile: | <ID> ... | ...  (ID als erstes Token der ersten Zelle).
    # Der Anker MUSS außerhalb der Kennungs-Zelle liegen — sonst passt die ID-Zelle
    # nicht mehr "vollständig" auf das id-pattern und die RTM (--trace/doc-complete)
    # erkennt 0 Anforderungen. Wir setzen ihn ans Ende der Zeile (vor das letzte '|').
    row = re.compile(r"^\|\s*(" + id_regex + r")\b")
    strip_anchor = re.compile(r'\s*<a id="[^"]*"></a>')
    lines = path.read_text().splitlines(keepends=True)
    count = 0
    for i, line in enumerate(lines):
        m = row.match(line)
        if not m:
            continue
        slug = m.group(1).lower()
        # Idempotent + Reparatur: bestehende Anker (auch fehlplatzierte) entfernen.
        eol = ""
        body = line
        while body and body[-1] in "\r\n":
            eol = body[-1] + eol
            body = body[:-1]
        body = strip_anchor.sub("", body).rstrip()
        if not body.endswith("|"):
            continue  # keine wohlgeformte Tabellenzeile
        inner = body[:-1].rstrip()  # alles vor dem schließenden '|'
        lines[i] = f'{inner} <a id="{slug}"></a> |' + eol
        count += 1
    path.write_text("".join(lines))
    return count


def main() -> int:
    root = Path(__file__).resolve().parent.parent
    total = 0
    for rel, regex in FAMILIES.items():
        if len(sys.argv) > 1 and rel not in sys.argv[1:]:
            continue
        p = root / rel
        n = anchor_file(p, regex)
        print(f"add_requirement_anchors: {rel} — {n} Anker gesetzt")
        total += n
    print(f"add_requirement_anchors: gesamt {total}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
