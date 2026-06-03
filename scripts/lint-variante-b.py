#!/usr/bin/env python3
"""
Variante-B Kommentar-Hygiene-Lint für Go-, Shell-, mjs- und Makefile-Kommentare.

Entfernt aus Kommentaren:
- plan-X.Y.Z-Refs (mit oder ohne §-Suffix)
- Tranche-N-Marker und Sub-X.Y-Suffixe
- Backticked Versions-Stempel `0.X.Y` in audit-trail-Kontexten
- "ab 0.X.Y", "seit 0.X.Y", "in 0.X.Y" Versions-Audit-Trail
- RAK-Wave-N Pseudo-Kennungen
- Sister-Doc-§-Suffixe (spec/X.md §Y.Z → spec/X.md)
- Lastenheft §-, API-Kontrakt §-, Architektur §-, Spike Spec §-Refs
- ADR-NNNN §X.Y → ADR-NNNN (preserves number!)
- DoD-Item §X-Y Marker

Behält:
- Echte Kennungen: ADR-NNNN, RAK-NN, R-NN, F-NN, NF-NN, MVP-NN, AK-NN
- Stabile Code-Anchor (apps/api/...)
- Echte Markdown-Links [text](url)

SICHERHEIT:
- Wirkt NUR auf Kommentar-Zeilen (// in Go/mjs, # in shell/Makefile, * in
  block-comments).
- Wirkt NIEMALS auf Code-Lines — Funktionsaufrufe `()` bleiben intakt.
- Spezialbehandlung für user-visible Runtime-Strings (Prometheus Help:,
  WriteString, Error-Messages mit plan-Refs).

Modes:
  check  — Report violations, exit 1 if any found (für CI)
  fix    — Apply cleanup in place
  diff   — Show what would change (dry-run)

Usage:
  python3 scripts/lint-variante-b.py check apps/api
  python3 scripts/lint-variante-b.py fix apps/api scripts Makefile
  python3 scripts/lint-variante-b.py diff apps/api/cmd/api/main.go
"""
import re
import sys
from pathlib import Path


# ============================================================
# COMMENT-LINE DETECTION (per file-type)
# ============================================================


def file_kind(path):
    """Return 'go' | 'mjs' | 'sh' | 'make' | None."""
    s = path.suffix
    if s == ".go":
        return "go"
    if s == ".mjs":
        return "mjs"
    if s == ".sh":
        return "sh"
    if path.name == "Makefile" or path.name.endswith(".mk"):
        return "make"
    return None


# ============================================================
# CLEANUP PATTERNS
# ============================================================


def clean_comment_text(t):
    """Apply Variante-B cleanup patterns to comment text only.

    Order matters: ADR-NNNN cleanup must happen BEFORE we drop bare § refs,
    because we need to preserve the ADR-NNNN number.
    """
    # 1. ADR-NNNN §X.Y[a-z]? — PRESERVE the NNNN!
    t = re.sub(r"(ADR-\d+)\s+§[0-9.]+[a-z]?", r"\1", t)

    # 2. plan-X.Y.Z §A.B[a-z]?(.C)?(/§D.E[a-z]?)?(?: Tranche N)?(?:[,/] RAK-NN)?
    #    preserve RAK-NN as standalone Kennung
    t = re.sub(
        r"plan-0\.\d+\.\d+(?:\.md)?\s+§[0-9.]+[a-z]?(?:/§[0-9.]+[a-z]?)?"
        r"(?:\s+Tranche\s+\d+(?:\s+Sub-[0-9.]+)?)?\s*[,/]\s*(RAK-\d+)",
        r"\1",
        t,
    )
    t = re.sub(
        r"plan-0\.\d+\.\d+(?:\.md)?\s+§[0-9.]+[a-z]?(?:/§[0-9.]+[a-z]?)?"
        r"(?:\s+Tranche\s+\d+(?:\s+Sub-[0-9.]+)?)?",
        "",
        t,
    )
    t = re.sub(
        r"plan-0\.\d+\.\d+(?:\.md)?\s+Tranche\s+\d+(?:\s+Sub-[0-9.]+)?\s*[,/]\s*(RAK-\d+)",
        r"\1",
        t,
    )
    t = re.sub(
        r"plan-0\.\d+\.\d+(?:\.md)?(?:\s+Tranche\s+\d+(?:\s+Sub-[0-9.]+)?)?",
        "",
        t,
    )

    # 3. Backticked versions `0.X.Y` Tranche N / RAK-NN
    t = re.sub(
        r"`0\.\d+\.\d+`\s+Tranche\s+\d+(?:\s+Sub-[0-9.]+)?\s*/\s*(RAK-\d+|R-\d+)",
        r"\1",
        t,
    )
    t = re.sub(
        r"`0\.\d+\.\d+`\s+Tranche\s+\d+(?:\s+Sub-[0-9.]+)?",
        "",
        t,
    )
    t = re.sub(r"\(`0\.\d+\.\d+`,\s*(RAK-\d+)\)", r"(\1)", t)
    t = re.sub(r"\(`0\.\d+\.\d+`\)", "", t)
    t = re.sub(r"`0\.\d+\.\d+`\s*/\s*(R-\d+)", r"\1", t)
    t = re.sub(r"`0\.\d+\.\d+`-Scope", "Scope", t)
    t = re.sub(r"`0\.\d+\.\d+`-", "", t)
    t = re.sub(r" ab `0\.\d+\.\d+`", "", t)
    t = re.sub(r" in `0\.\d+\.\d+`", "", t)
    t = re.sub(r" seit `0\.\d+\.\d+`", "", t)

    # 4. Bare "ab/seit/in 0.X.Y"
    t = re.sub(r" ab 0\.\d+\.\d+\s+Tranche \d+", "", t)
    t = re.sub(r" ab 0\.\d+\.\d+([ .,])", r"\1", t)
    t = re.sub(r" seit 0\.\d+\.\d+([ .,])", r"\1", t)
    t = re.sub(r"Default ab 0\.\d+\.\d+ ist", "Default ist", t)

    # 5. Stand-alone Tranche-Marker
    t = re.sub(r" \(Tranche \d+\)", "", t)
    t = re.sub(r" Tranche \d+ fixiert", " fixiert", t)
    t = re.sub(r"in Tranche \d+ liefert", "liefert", t)
    t = re.sub(r"Tranche \d+ liefert die", "Aktuell sind die", t)
    t = re.sub(r"Tranche \d+ nutzt das im", "Nutzt das im", t)
    t = re.sub(r"das Tranche-\d+-Modell", "das Modell", t)
    t = re.sub(r"Tranche \d+ fokussiert", "Aktuell fokussiert", t)
    t = re.sub(r"Tranche \d+ wird", "Aktuell wird", t)
    t = re.sub(r"Tranche \d+ ist", "Aktuell ist", t)
    t = re.sub(r"Tranche-\d+ ", "", t)
    t = re.sub(r"\s+Tranche\s+\d+(?=\W)", "", t)
    t = re.sub(r"^Tranche \d+\s*", "", t, flags=re.MULTILINE)

    # 6. RAK-Wave (pseudo-Kennung)
    t = re.sub(r"RAK-Wave-\d+\s*/?\s*", "", t)

    # 7. Sister-doc §-suffixes (preserve doc name)
    for doc in (
        "spec/telemetry-model.md",
        "spec/backend-api-contract.md",
        "spec/architecture.md",
        "spec/lastenheft.md",
        "spec/player-sdk.md",
        "backend-api-contract.md",
        "telemetry-model.md",
        "architecture.md",
        "auth.md",
        "extra-gates.md",
        "ingest-control.md",
        "releasing.md",
        "stream-analyzer.md",
        "local-development.md",
        "fuzzing.md",
        "mutation-testing.md",
        "perf/budgets.md",
        "docs/perf/budgets.md",
        "docs/user/auth.md",
        "docs/dev/fuzzing.md",
        "docs/dev/mutation-testing.md",
    ):
        doc_esc = re.escape(doc)
        t = re.sub(rf"({doc_esc})\s+§[0-9a-z.]+(?:/§[0-9a-z.]+)?", r"\1", t)

    # 8. Lastenheft §, Spike Spec §, API-Kontrakt §, Architektur §
    t = re.sub(r"Lastenheft\s+§[0-9.]+", "", t)
    t = re.sub(r"Spike Spec\s+§[0-9.]+", "Spike Spec", t)
    t = re.sub(r"API-Kontrakt\s+§[0-9a-z.]+", "API-Kontrakt", t)
    t = re.sub(r"Architektur\s+§[0-9.]+", "Architektur", t)

    # 9. DoD-Item §X-Y
    t = re.sub(r"\(DoD-Item §[0-9-]+\)\s*—\s*", "", t)
    t = re.sub(r"\s+\(DoD-Item §[0-9-]+\)", "", t)
    t = re.sub(r"DoD-Item §[0-9-]+", "DoD-Item", t)

    # 10. Cleanup artifacts in COMMENT TEXT (safe because we never touch code)
    t = re.sub(r"\(\s*\)", "", t)
    t = re.sub(r"\(\s*—\s*", "(", t)
    t = re.sub(r"\(\s*/\s*", "(", t)
    t = re.sub(r"\(\s*,\s*", "(", t)
    t = re.sub(r"  +", " ", t)
    t = re.sub(r" ([.,;:])", r"\1", t)

    return t


def clean_runtime_string(line):
    """Clean runtime-visible strings (Prometheus Help:, WriteString, error messages)."""
    line = re.sub(
        r'Help:\s*"Total accepted WebRTC samples grouped by RTCPeerConnectionState \(plan-0\.\d+\.\d+ §\d+ Tranche \d+\)\."',
        'Help: "Total accepted WebRTC samples grouped by RTCPeerConnectionState."',
        line,
    )
    line = re.sub(
        r" siehe plan-0\.\d+\.\d+ Tranche \d+ / (R-\d+)",
        r" siehe \1",
        line,
    )
    line = re.sub(
        r'WriteString\("# Generated by apps/api ingest-control \(plan-0\.\d+\.\d+ Tranche \d+, (RAK-\d+)\)\.\\n"\)',
        r'WriteString("# Generated by apps/api ingest-control (\1).\\n")',
        line,
    )
    line = re.sub(
        r"(volumes); see plan-0\.\d+\.\d+ §\d+ Backend-Strategie",
        r"\1",
        line,
    )
    line = re.sub(
        r"is a follow-up item after 0\.\d+\.\d+ Tranche \d+",
        "is a follow-up item",
        line,
    )
    return line


# ============================================================
# FILE PROCESSING
# ============================================================


def process_line_based(text, comment_starts, block_starts=None):
    """Process file line-by-line. Only cleanup lines identified as comments.

    comment_starts: list of strings that begin a single-line comment ("//", "#")
    block_starts: tuple of (start, end) for block comments ("/*", "*/"), or None
    """
    lines = text.split("\n")
    new_lines = []
    in_block = False
    changed = False

    for line in lines:
        original = line
        stripped = line.lstrip()
        is_comment = False
        is_runtime_string = False

        if in_block:
            is_comment = True
            if block_starts and block_starts[1] in line:
                in_block = False
        elif block_starts and stripped.startswith(block_starts[0]):
            is_comment = True
            if block_starts[1] not in line[line.index(block_starts[0]) + len(block_starts[0]):]:
                in_block = True
        elif any(stripped.startswith(cs) for cs in comment_starts):
            is_comment = True
        elif 'Help: "' in line or 'WriteString("' in line or 'sqlite is not supported' in line:
            is_runtime_string = True

        if is_comment:
            # Find the comment prefix
            m = None
            if block_starts:
                m = re.match(
                    rf"^(\s*(?:{re.escape(block_starts[0])}|{re.escape(block_starts[1])}|\*)\s?)(.*)$",
                    line,
                )
            if m is None:
                for cs in comment_starts:
                    m2 = re.match(rf"^(\s*{re.escape(cs)}\s?)(.*)$", line)
                    if m2:
                        m = m2
                        break
            if m:
                prefix, content = m.group(1), m.group(2)
                cleaned = clean_comment_text(content).rstrip()
                new_line = (prefix + cleaned).rstrip()
                if new_line != original:
                    changed = True
                new_lines.append(new_line)
            else:
                new_lines.append(line)
        elif is_runtime_string:
            new_line = clean_runtime_string(line)
            if new_line != original:
                changed = True
            new_lines.append(new_line)
        else:
            new_lines.append(line)

    return "\n".join(new_lines), changed


def process_file(path, fix=False):
    """Returns (changed_bool, original_text, new_text)."""
    kind = file_kind(path)
    if kind is None:
        return False, None, None

    text = path.read_text()
    if kind == "go" or kind == "mjs":
        new_text, changed = process_line_based(text, ["//"], block_starts=("/*", "*/"))
    elif kind == "sh":
        new_text, changed = process_line_based(text, ["#"])
    elif kind == "make":
        new_text, changed = process_line_based(text, ["#"])
    else:
        return False, None, None

    if changed and fix:
        path.write_text(new_text)
    return changed, text, new_text


# ============================================================
# CLI
# ============================================================


def iter_targets(paths):
    for p in paths:
        path = Path(p)
        if path.is_file():
            if file_kind(path):
                yield path
        elif path.is_dir():
            for sub in path.rglob("*"):
                if sub.is_file() and file_kind(sub):
                    yield sub


def main():
    if len(sys.argv) < 3:
        print(__doc__)
        sys.exit(2)
    mode, *targets = sys.argv[1:]
    if mode not in ("check", "fix", "diff"):
        print(f"unknown mode: {mode}")
        sys.exit(2)

    violators = 0
    changed_files = 0
    for path in iter_targets(targets):
        changed, original, new = process_file(path, fix=(mode == "fix"))
        if changed:
            violators += 1
            if mode == "check":
                # Report first divergent line
                for i, (oa, ob) in enumerate(
                    zip(original.split("\n"), new.split("\n")), 1
                ):
                    if oa != ob:
                        print(f"{path}:{i}: {oa.strip()[:120]}")
                        break
            elif mode == "fix":
                changed_files += 1
            elif mode == "diff":
                print(f"=== {path} ===")
                import difflib

                for line in difflib.unified_diff(
                    original.split("\n"),
                    new.split("\n"),
                    lineterm="",
                    n=1,
                ):
                    print(line)

    if mode == "check":
        if violators:
            print(f"\n{violators} file(s) have Variante-B violations. Run `fix` to clean.")
            sys.exit(1)
        print("✓ no Variante-B violations")
    elif mode == "fix":
        print(f"changed: {changed_files}")


if __name__ == "__main__":
    main()
