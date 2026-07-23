#!/usr/bin/env python3
"""Structural closure-note gate for done/ plans.

Computational half of the closure-note enforcement introduced by ADR-0010.
It checks *structure* only — heading present, minimum sentence count outside
code fences, and a bare-phrase (Floskel) blocklist. The *semantic* half
(content vs. filler: does the note carry a real learning signal / follow-up /
architecture observation) is the `.harness/skills/closure-note-reviewer.md`
inferential reviewer; this script never tries to judge content quality.

Brownfield: every plan that predates the policy is exempt via the
grandfather list. A plan that is neither grandfathered nor compliant fails
the gate — that friction is the point: a new done/ plan must carry a note or
be explicitly grandfathered with justification.

Exit code 0 = clean, 1 = violations, 2 = usage/config error.
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

# A closure section is a heading whose text contains "Closure". The
# slice/welle template carries *two* such headings — "## 5. Closure-Trigger"
# (boilerplate) and "## 7. Closure-Notiz" (the real note) — so we prefer the
# note heading ("Closure-Not…"/"Closure Note") and fall back to any "Closure"
# heading for plans that only carry a generic closure section.
CLOSURE_HEADING = re.compile(r"^(#{2,})\s+.*Closure", re.IGNORECASE | re.MULTILINE)
CLOSURE_NOTE_HEADING = re.compile(
    r"^(#{2,})\s+.*Closure[-\s]?Not", re.IGNORECASE | re.MULTILINE
)
ANY_HEADING = re.compile(r"^#{1,6}\s+", re.MULTILINE)
FENCE = re.compile(r"^```.*?^```", re.DOTALL | re.MULTILINE)
HTML_COMMENT = re.compile(r"<!--.*?-->", re.DOTALL)
SENTENCE_END = re.compile(r"[.!?](?:\s|$)")

# Bare fillers: a note reduced to only these (after stripping) is not a note.
FLOSKEL = [
    "fertig",
    "erledigt",
    "done",
    "wie geplant umgesetzt",
    "wie geplant",
    "wie erwartet",
    "laeuft jetzt",
    "laeuft",
    "läuft jetzt",
    "läuft",
    "alles gut",
    "war ganz okay",
    "war okay",
    "passt",
]

MIN_SENTENCES = 2


def load_grandfather(path: Path) -> set[str]:
    if not path.exists():
        return set()
    names: set[str] = set()
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.split("#", 1)[0].strip()
        if line:
            names.add(line)
    return names


def closure_section_text(body: str) -> str | None:
    """Return the text of the closure section, or None if absent."""
    # Prefer the dedicated note heading ("Closure-Notiz"/"Closure Note"); fall
    # back to the first generic "Closure" heading so plans with only a single
    # closure section keep working.
    m = CLOSURE_NOTE_HEADING.search(body) or CLOSURE_HEADING.search(body)
    if not m:
        return None
    start = m.end()
    level = len(m.group(1))
    # Section runs until the next heading of the same or higher level.
    rest = body[start:]
    end = len(rest)
    for h in ANY_HEADING.finditer(rest):
        if len(h.group(0).split()[0]) <= level:
            end = h.start()
            break
    return rest[:end]


def is_substantive(section: str) -> tuple[bool, str]:
    text = FENCE.sub("", section)
    text = HTML_COMMENT.sub("", text)
    # Strip markdown scaffolding to judge the prose.
    stripped = re.sub(r"[>*#`\-|_\s]+", " ", text).strip()
    if not stripped:
        return False, "leer (kein Prosatext außerhalb von Code/Kommentaren)"
    lowered = stripped.lower()
    if lowered in FLOSKEL or all(
        part.strip() in FLOSKEL for part in re.split(r"[.!?]", lowered) if part.strip()
    ):
        return False, f"nur Floskel ohne Substanz: {stripped!r}"
    sentences = [s for s in SENTENCE_END.split(stripped) if s.strip()]
    if len(sentences) < MIN_SENTENCES:
        return False, f"unter {MIN_SENTENCES} Sätzen ({len(sentences)})"
    return True, ""


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--done-dir", default="docs/plan/planning/done", type=Path)
    ap.add_argument(
        "--grandfather-file",
        default="docs/plan/planning/.closure-grandfathered",
        type=Path,
    )
    ap.add_argument(
        "--glob",
        action="append",
        default=None,
        help=(
            "Datei-Glob(s) in done/; wiederholbar. Default deckt die drei "
            "Plan-Familien ab: plan-*.md, slice-*.md, welle-*.md."
        ),
    )
    args = ap.parse_args()

    if not args.done_dir.is_dir():
        print(f"check_closure_notes: done-dir nicht gefunden: {args.done_dir}", file=sys.stderr)
        return 2

    globs = args.glob or ["plan-*.md", "slice-*.md", "welle-*.md"]
    grandfathered = load_grandfather(args.grandfather_file)
    seen: set[Path] = set()
    plans: list[Path] = []
    for pattern in globs:
        for plan in sorted(args.done_dir.glob(pattern)):
            if plan not in seen:
                seen.add(plan)
                plans.append(plan)
    plans.sort()
    checked = 0
    exempt = 0
    violations: list[str] = []

    for plan in plans:
        if plan.name in grandfathered:
            exempt += 1
            continue
        checked += 1
        body = plan.read_text(encoding="utf-8")
        section = closure_section_text(body)
        if section is None:
            violations.append(f"{plan}: keine Closure-Sektion (## …Closure…)")
            continue
        ok, reason = is_substantive(section)
        if not ok:
            violations.append(f"{plan}: Closure-Sektion {reason}")

    if violations:
        print("check_closure_notes: FEHLER — Closure-Note-Pflicht (ADR-0010) verletzt:")
        for v in violations:
            print(f"  - {v}")
        print(
            f"\n{len(violations)} Verletzung(en); {checked} geprüft, "
            f"{exempt} grandfathered ({args.grandfather_file})."
        )
        return 1

    print(
        f"check_closure_notes: OK — {checked} Plan/Pläne mit gültiger Closure-Note, "
        f"{exempt} grandfathered."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
