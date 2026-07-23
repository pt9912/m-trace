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
    # Definitionszeile: | <ID> ... | ...  (ID als erstes Token der ersten Zelle)
    row = re.compile(r"^(\|\s*)(" + id_regex + r")(\b)")
    lines = path.read_text().splitlines(keepends=True)
    count = 0
    for i, line in enumerate(lines):
        m = row.match(line)
        if not m:
            continue
        ident = m.group(2)
        # Rest der ersten Zelle bis zum nächsten '|' prüfen: schon verankert?
        cell_end = line.find("|", m.end())
        cell = line[m.end(): cell_end if cell_end != -1 else len(line)]
        if "<a id=" in cell:
            continue
        slug = ident.lower()
        anchor = f' <a id="{slug}"></a>'
        lines[i] = line[: m.end()] + anchor + line[m.end():]
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
