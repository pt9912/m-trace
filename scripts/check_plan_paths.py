#!/usr/bin/env python3
"""check_plan_paths.py — Plan-§3-Pfade gegen den realen Slice-Diff (BEO-PLAN-Sensor).

Je Slice werden die in Plan §3 deklarierten Pfade gegen die real geänderten
Dateien der Slice-Range geprüft. Zwei Befund-Richtungen:

  PHANTOM      §3 deklariert einen Pfad, der in der Range nie berührt wurde.
  UNDECLARED   die Range ändert eine Datei, die §3 nicht deklariert.

Lifecycle-Ausnahmen (nie Befund): die Slice-Datei selbst (alle vier
Lifecycle-Orte), `docs/plan/planning/in-progress/roadmap.md` (Ruhe-Marker-
Tanz), `docs/reviews/**` (Review-Report-Artefakte),
`docs/plan/planning/observations/**` (Evidence-Dateien).

Aufruf: ohne Argument werden alle Slices in `in-progress/` geprüft (die
Gate-Position); mit Kennung (`slice-0NN`) der genannte done-Slice
(Kalibrierung über historische Stände). Exit 1 bei Befunden, Exit 2 bei
Verwendungsfehlern, Exit 0 sauber.

Erstvorkommen der Klasse: Review slice-032 F-1 (BEO-PLAN, 5. Auftreten der
Familie „Plan-Angabe vs. realer Diff"); Ausschöpfungs-Entscheid
slice-029-Closure (Prosa ausgeschöpft, Sensor modul-06-gefordert).
"""

import fnmatch
import re
import subprocess
import sys

PLANNING = "docs/plan/planning"
LIFECYCLE = ["open", "next", "in-progress", "done"]
# Lifecycle-Ausnahmen: Dateien, die je Slice-Range sinnvoll geändert werden,
# ohne in Plan §3 zu stehen (Marker-Tanz, Review-Report, Evidence).
EXCEPT_PREFIXES = [
    "docs/plan/planning/in-progress/roadmap.md",
    "docs/reviews/",
    "docs/plan/planning/observations/",
]


def git(*args):
    return subprocess.run(
        ["git", *args], capture_output=True, text=True, check=False
    )


def slice_file(kennung):
    """Liefert den Pfad der Slice-Datei über alle Lifecycle-Orte (0 oder 1)."""
    treffer = []
    for d in LIFECYCLE:
        out = git("ls-files", "-z", "--", f"{PLANNING}/{d}/{kennung}*.md")
        treffer += [p for p in out.stdout.split("\0") if p]
    return treffer


def plan_pfade(slice_path):
    """Extrahiert die deklarierten Pfade aus Plan §3 (Backtick-Token, Zelle 1)."""
    text = open(slice_path, encoding="utf-8").read()
    pfade = set()
    in_drei = False
    for zeile in text.splitlines():
        if zeile.startswith("## "):
            in_drei = zeile.startswith("## 3.")
            continue
        if in_drei and zeile.startswith("|"):
            zellen = [z.strip() for z in zeile.strip("|").split("|")]
            if zellen and zellen[0] and zellen[0] != "Datei / Komponente":
                pfade.update(re.findall(r"`([^`]+)`", zellen[0]))
    return pfade


def range_und_diff(slice_path):
    """Range (Anlage^ .. letzte Berührung) und geänderte Dateien je Slice."""
    # NB: `--follow` kombiniert nicht mit `--reverse` (git liefert newest-first,
    # --reverse bleibt still unwirksam) — die Ordnung wird hier explizit gedreht.
    commits = git("log", "--follow", "--format=%H", "--", slice_path).stdout.split()
    if len(commits) < 1:
        return None, None, []
    erster, letzter = commits[-1], commits[0]
    diff = git(
        "diff", "--name-status", "-M", f"{erster}^..{letzter}"
    ).stdout.splitlines()
    geaendert = set()
    for zeile in diff:
        teile = zeile.split("\t")
        status = teile[0][0]  # R100 -> R: Rename-Quelle und -Ziel zählen beide
        rest = teile[1:]
        if status == "R":
            geaendert.add(rest[0])
            geaendert.add(rest[1])
        elif status == "D":
            geaendert.add(rest[0])
        else:
            geaendert.add(rest[-1])
    return erster, letzter, sorted(geaendert)


def ist_ausnahme(pfad, slice_path):
    if pfad == slice_path:
        return True
    if pfad.rsplit("/", 1)[-1] == slice_path.rsplit("/", 1)[-1]:
        return True
    for praefix in EXCEPT_PREFIXES:
        if pfad.startswith(praefix):
            return True
    return False


def pruefe(kennung, slice_path):
    _, _, geaendert = range_und_diff(slice_path)
    deklariert = plan_pfade(slice_path)
    befunde = []

    for pfad in sorted(geaendert):
        if ist_ausnahme(pfad, slice_path):
            continue
        if not any(fnmatch.fnmatch(pfad, m) for m in deklariert):
            befunde.append(
                f"UNDECLARED: {pfad} wurde in der Range geändert, steht nicht in Plan §3"
            )

    for muster in sorted(deklariert):
        if not any(fnmatch.fnmatch(p, muster) for p in geaendert):
            befunde.append(
                f"PHANTOM: {muster} steht in Plan §3, wurde in der Range nie berührt"
            )

    return befunde


def main():
    args = sys.argv[1:]
    if len(args) > 1:
        print("Aufruf: check_plan_paths.py [slice-kennung]", file=sys.stderr)
        return 2

    if args:
        kennung = args[0]
        dateien = slice_file(kennung)
        if not dateien:
            print(f"plan-paths: kein Slice '{kennung}' gefunden.", file=sys.stderr)
            return 2
        prueflinge = [(kennung, dateien[0])]
    else:
        out = git("ls-files", "-z", "--", f"{PLANNING}/in-progress/slice-*.md")
        prueflinge = [
            (p.split("/")[-1].removesuffix(".md"), p)
            for p in out.stdout.split("\0")
            if p
        ]

    if not prueflinge:
        print("plan-paths: keine in-progress-Slices — nichts zu prüfen.")
        return 0

    befunde_gesamt = []
    for kennung, pfad in prueflinge:
        befunde = pruefe(kennung, pfad)
        print(f"plan-paths: {kennung} — {len(befunde)} Befund(e)")
        befunde_gesamt += [(kennung, b) for b in befunde]

    for kennung, befund in befunde_gesamt:
        print(f"plan-paths [{kennung}]: {befund}")

    return 1 if befunde_gesamt else 0


if __name__ == "__main__":
    sys.exit(main())
