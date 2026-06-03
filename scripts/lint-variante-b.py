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
    """Return 'go' | 'mjs' | 'ts' | 'svelte' | 'sh' | 'make' | None."""
    s = path.suffix
    if s == ".go":
        return "go"
    if s == ".mjs":
        return "mjs"
    if s == ".ts":
        return "ts"
    if s == ".tsx":
        return "ts"
    if s == ".js":
        return "mjs"
    if s == ".jsx":
        return "mjs"
    if s == ".svelte":
        return "svelte"
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

    # 1b. Backtick-wrapped plan-Refs `plan-X.Y.Z` → unwrap so the
    #     subsequent plan- patterns can strip cleanly (otherwise we
    #     leave empty `` behind).
    t = re.sub(r"`plan-(0\.\d+\.\d+(?:\.md)?)`", r"plan-\1", t)

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
        r"(?:\s+H\d+)?"  # plan-0.4.0 §5 H3 style
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

    # 4. Bare "ab/seit/in 0.X.Y" (numeric)
    t = re.sub(r" ab 0\.\d+\.\d+\s+Tranche \d+", "", t)
    t = re.sub(r" ab 0\.\d+\.\d+([ .,])", r"\1", t)
    t = re.sub(r" seit 0\.\d+\.\d+([ .,])", r"\1", t)
    t = re.sub(r"Default ab 0\.\d+\.\d+ ist", "Default ist", t)

    # 4b. "0.X.y" or "0.X.x" pattern with letter suffix (e.g. 0.3.x)
    t = re.sub(r"0\.\d+\.[xy]-(?:Backends?|Pflichten|Verhalten[a-z]*|Scope)", "Server", t)
    t = re.sub(r"reales 0\.\d+\.[xy]\b", "reales Verhalten", t)
    t = re.sub(r" 0\.\d+\.[xy]([ .,;:])", r"\1", t)
    t = re.sub(r"^0\.\d+\.[xy]([ .,;:])", r"\1", t)

    # 4e. "ab 0.X.Y §A.B-Closeout" / "seit 0.X.Y §A.B-Closeout" patterns
    t = re.sub(r"Server-vergeben ab 0\.\d+\.\d+ §[0-9.]+[a-z]?-Closeout", "Server-vergeben", t)
    t = re.sub(r" ab 0\.\d+\.\d+ §[0-9.]+[a-z]?-Closeout", "", t)
    t = re.sub(r" seit 0\.\d+\.\d+ §[0-9.]+[a-z]?-Closeout", "", t)
    # Bare §X.Y-Closeout artifacts (incl. dangling prepositions)
    t = re.sub(r" (?:vor|nach|bis|seit|ab) §[0-9.]+[a-z]?-Closeout", " historisch", t)
    t = re.sub(r" §[0-9.]+[a-z]?-Closeout", "", t)
    t = re.sub(r"vor dem Closeout", "historisch", t)
    # "ab plan-X §A.B/§C.D" already covered, but explicit "ab plan-X" form
    t = re.sub(r" ab plan-0\.\d+\.\d+(?:\.md)?\s+§[0-9.]+[a-z]?(?:/§[0-9.]+[a-z]?)?\s+", " ", t)
    t = re.sub(r" ab plan-0\.\d+\.\d+(?:\.md)?", "", t)
    # "ist Folge-Scope/in Tranche N"
    t = re.sub(r" Plan-DoD §[0-9-]+\s+verschiebt das auf Tranche \d+", " ist Folge-Scope", t)

    # 4c. Orphan "#N-" prefix (from §X.Yz-#N- pattern strip)
    t = re.sub(r"^#\d+-(Vertrag|Snapshot|Item)\b", r"\1", t)
    t = re.sub(r" #\d+-(Vertrag|Snapshot|Item)\b", r" \1", t)

    # 4d. RFC NNNN §X.Y — external standard reference, drop § suffix
    t = re.sub(r"(RFC \d+)\s+§[0-9.]+", r"\1", t)

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
    # Trailing "," or ", " inside parens before closing
    t = re.sub(r",\s*\)", ")", t)
    # Stand-alone "(H1)" or "(Hn)" — orphan from plan-X §Y Hn strip
    # Including "( H4). " patterns mid-sentence (drop the orphan + dot)
    t = re.sub(r"\s*\(\s*H\d+\s*\)\s*\.\s*", " ", t)
    t = re.sub(r"\(\s*H\d+\)", "", t)
    t = re.sub(r"\(\s*H\d+\s*[:.]", "(", t)
    # "§5 H5" / "Tranche 4 §5 H5" leftovers — orphan H-numbers
    t = re.sub(r" §\d+\s+H\d+(?=[\W])", "", t)
    t = re.sub(r" H\d+:\s*", " ", t)
    # Trailing artifacts like "Default `[]`; siehe API-Kontrakt, "
    # Drop trailing ", " before " */" only — never on bare comment lines
    # (those are legitimate prose continuations).
    t = re.sub(r",\s*\*\/", " */", t)
    # Collapse multiple spaces — but only in the MIDDLE of content, never at
    # the very start. Markdown-style list indents (`  - foo`) inside block
    # comments are meaningful structure and must survive cleanup.
    t = re.sub(r"(?<=\S)  +(?=\S)", " ", t)
    t = re.sub(r" ([.;:])", r"\1", t)
    # ", " followed by "(" can collapse but only safely inside parens
    # Already handled in `\(\s*,\s*` pattern earlier.

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

    # Multi-line orphan cleanup (PHASE 2): handle artifacts left by
    # single-line cleanup that span multiple comment lines.
    #
    # CRITICAL safety: every operation here verifies that BOTH the
    # prev-line AND the next-line are comment lines (never touches code).

    def is_cmt(s):
        """True if line is a comment per current syntax."""
        st = s.lstrip()
        if any(st.startswith(cs) for cs in comment_starts):
            return True
        if block_starts and (st.startswith(block_starts[0]) or st.startswith("*")):
            return True
        return False

    def cmt_prefix(line):
        """Return (prefix, content) tuple, or (None, None) if not a comment."""
        if block_starts:
            m = re.match(
                rf"^(\s*(?:{re.escape(block_starts[0])}|{re.escape(block_starts[1])}|\*)\s?)(.*)$",
                line,
            )
            if m:
                return m.group(1), m.group(2)
        for cs in comment_starts:
            m = re.match(rf"^(\s*{re.escape(cs)}\s?)(.*)$", line)
            if m:
                return m.group(1), m.group(2)
        return None, None

    # Phase 2a: orphan §-line starts (e.g. `// §5.1). Der Use Case ruft...`)
    out = list(new_lines)
    skip = set()
    for i, line in enumerate(out):
        if i in skip:
            continue
        if not is_cmt(line):
            continue
        prefix, content = cmt_prefix(line)
        if content is None:
            continue
        # Pattern: line starts with `§X.Y...` then `).` or `).` followed by text
        m = re.match(
            r"^§[0-9a-z.]+(?:\s*/\s*§[0-9a-z.]+)?\s*[).,;:-]\s*(.*)$",
            content,
        )
        if m:
            rest = m.group(1)
            if rest.strip():
                out[i] = prefix + rest
            else:
                skip.add(i)
            changed = True
            continue
        m2 = re.match(
            r"^§[0-9a-z.]+(?:\s*/\s*§[0-9a-z.]+)?\s+(.*)$",
            content,
        )
        if m2:
            rest = m2.group(1)
            if rest.strip():
                out[i] = prefix + rest
            else:
                skip.add(i)
            changed = True

    # Phase 2b: orphan opening parens (line ends with ` (`, next line is closing)
    for i in range(len(out) - 1):
        if i in skip or (i + 1) in skip:
            continue
        cur = out[i]
        nxt = out[i + 1]
        if not (is_cmt(cur) and is_cmt(nxt)):
            continue
        cur_rstripped = cur.rstrip()
        if not cur_rstripped.endswith("("):
            continue
        _, nxt_content = cmt_prefix(nxt)
        if nxt_content is None:
            continue
        new_prev = re.sub(r"\s*\(\s*$", "", cur).rstrip()
        # Case A: just `).` → drop next
        if re.match(r"^\)\s*\.?\s*$", nxt_content):
            out[i] = new_prev
            skip.add(i + 1)
            changed = True
            continue
        # Case B: continuation with Kennung — drop next
        if re.match(
            r"^[/]?\s*(R-\d+|RAK-\d+|F-\d+|NF-\d+|MVP-\d+|AK-\d+|Sub-[0-9.]+)",
            nxt_content,
        ):
            if nxt_content.rstrip().endswith((")", ").", ");", "),", "):", ")")):
                out[i] = new_prev
                skip.add(i + 1)
                changed = True
                continue
        # Case C: sister-doc ref — drop next
        if re.match(r"^[`]?(spec/|docs/)[a-z/0-9-]+\.md", nxt_content):
            if nxt_content.rstrip().endswith((")", ").", ");", "),", "):", ")")):
                out[i] = new_prev
                skip.add(i + 1)
                changed = True
                continue
        # Case D: any other content with closing paren → merge
        if ")" in nxt_content:
            merged = re.sub(r"^[/]?\s*", "", nxt_content)
            if merged.count("(") < merged.count(")"):
                merged = re.sub(r"\)\s*([.,;:]?)\s*$", r"\1", merged)
            if merged.strip():
                out[i] = new_prev + " " + merged.strip()
            else:
                out[i] = new_prev
            skip.add(i + 1)
            changed = True
            continue
        # Otherwise: just drop the trailing ` (`
        out[i] = new_prev
        changed = True

    new_lines = [l for j, l in enumerate(out) if j not in skip]
    return "\n".join(new_lines), changed


def process_file(path, fix=False):
    """Returns (changed_bool, original_text, new_text)."""
    kind = file_kind(path)
    if kind is None:
        return False, None, None

    text = path.read_text()
    if kind in ("go", "mjs", "ts", "svelte"):
        # TS/JS/Go/mjs/Svelte-script-blocks all use // and /* */
        # Svelte has additional HTML <!-- --> comments in templates,
        # but those are rare audit-trail vectors — script blocks are
        # the high-signal area.
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
