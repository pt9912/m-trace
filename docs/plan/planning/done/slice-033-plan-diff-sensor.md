# Slice 033: Mechanischer Sensor — Plan-§3-Pfade gegen Slice-Diff

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle — Closure-Bedingung
ist die DoD dieses Slices, keine Wellen-Grenze.

**Bezug:** `BEO-PLAN/plan-diff-drift` (`state.md`: `geplant`), Review-Reports
`2026-09-29-slice-025.md` bis `-slice-032.md` (Familie „Plan-Angabe vs. realer Diff", 5×),
`.d-check.yml`, `tools/` bzw. `scripts/` (Sensor-Ablageort offen).

**Autor:** modul-06-Konsequenz nach 5. Familien-Auftreten (Review slice-032,
MEDIUM-Finding); Owner-Freigabe über BEO-PLAN-Folge. **Datum:** 2026-09-29.

---

## 1. Ziel und Abgrenzung

**Ziel:** Die Prosa-Ausschöpfung der BEO-PLAN-Familie („Plan-Angabe vs.
realer Diff", 5× — zuletzt slice-032: Pfad- und Fragment-Angaben im Plan
ohne Diff-Deckung) mit dem modul-06-geforderten mechanischen Sensor
adressieren: je done-Slice werden die in Plan §3 deklarierten Pfade gegen
die real geänderten Dateien der Slice-Range (`git diff --name-status`)
geprüft — Abweichungen (Datei geändert, aber nicht deklariert; deklariert,
aber nie berührt; falsche Änderungs-Art bei Neuanlagen) melden als Befund.
Ablageort und Anbindung im Closure-Entscheid festlegen (d-check-Modul
kandidiert; alternativ eigenes Skript neben `verify-closure-notes`).

**Ausdrücklich NICHT in diesem Slice:**

- **Retro-Triage der fünf Familie-Funde** — die Reports sind
  Lauf-Belege; der Zähler läuft über die Evidence-Dateien und bleibt.
- **Semantik-Prüfung** („Sagt der Plan-Text das Wahre?") — der Sensor prüft
  die Pfad-Menge und -Arten, nicht Formulierungen (die bleiben
  Urteilsfall, siehe Closure-Notiz slice-029).
- **Retro-Edit archivierter Plan-Dateien** — done/ bleibt.

## 2. Definition of Done

- [x] Sensor-Target existiert und ist an die Gate-Kette angebunden (Target-
      Name im Closure-Entscheid; d-check-Modul oder Skript+Target).
- [x] Sensor meldet über die bestehenden done-Slices (025–032) die
      bekannten Drift-Fälle reproduzierbar (Verifikation gegen die
      Review-Fundliste).
- [x] Lifecycle-Ausnahmen deklariert und begrenzt: Marker-Tanz (Roadmap),
      Slice-Dokument selbst, Review-Report-Dateien.
- [x] `make gates` grün mit dem neuen Sensor in der Kette.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag; BEO-PLAN `state.md` auf
      `verkörpert` (Zielort + `seit slice-033`).

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `scripts/check_plan_paths.py` | neu | Plan-§3-Pfade gegen Slice-Range-Diff |
| `Makefile` | update | Target-Anbindung |
| `docs/plan/planning/observations/BEO-PLAN/plan-diff-drift/state.md` | update | Ausgang `geplant` → `verkörpert` |

## 4. Trigger

- **Start (`next` → `in-progress`):** sofort — Owner-Freigabe über
  BEO-PLAN-Folge (Prosa-Ausschöpfung).
- **Rückführungen:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + Gates grün + Closure-Notiz + `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **False-Positives bei Lifecycle-Mustern** (Marker-Tanz, Slice-Dokument,
  Review-Reports) — **Ausgang:** Deklarationsliste im Sensor, im Slice
  033-Design festgelegt und am ersten Lauf kalibriert.
- **Range-Ermittlung je done-Slice** (Anlage-Commit → done-Commit) —
  **Ausgang:** offen bis Design; Git-Ancestry per `git log --follow` auf
  die Slice-Datei.

## 7. Closure-Notiz

Der mechanische Sensor ist verkörpert: `scripts/check_plan_paths.py` +
Target `verify-plan-paths` in der `gates`-Kette. Je Slice-Range
(Anlage^..letzte Berührung, `--follow`) werden die Plan-§3-Pfade gegen die
real geänderten Dateien geprüft — zwei Befund-Richtungen (PHANTOM /
UNDECLARED), Lifecycle-Ausnahmen deklarierbar. Kalibriert über alle
done-Slices 025–032: slice-032 reproduziert die Review-Funde 1:1 (3
Befunde), slice-025/028 zeigen die historischen Drifts (7 bzw. 2),
026/027/029/030 sauber. BEO-PLAN `state.md` → `verkörpert` (Zielort
`scripts/check_plan_paths.py` + `Makefile` Target `verify-plan-paths`,
`seit slice-033`).

**Was hat funktioniert:** Kalibrierung gegen die bekannten Funde vor dem
Ausbau — der Sensor meldete slice-032s Drift exakt (3 Befunde) und blieb
auf den sauberen Slices stumm.

**Was ging anders als geplant:** Git-Quirk: `--follow` kombiniert nicht
mit `--reverse` (git ignoriert `--reverse` still) — die erste Range-Version
war rückwärts und lief leer; der Fix dreht die Ordnung explizit im Code
(kommentiert im Skript).

**Steering-Loop-Eintrag:** Sensor verkörpert — die Familie „Plan-Angabe
vs. realer Diff" wird künftig beim ersten Verstoß (statt nach der
Ausschöpfung) gemeldet; `state.md` von BEO-PLAN trägt den Ausgang.
— liegt in `scripts/check_plan_paths.py` + `Makefile` Target
`verify-plan-paths`. Auslöser: `BEO-PLAN/plan-diff-drift`
(slice-025/026/028/032, 5×).

**Beobachtungs-Register (`../observations/`):** `BEO-PLAN/plan-diff-drift`
Ausgang vollzogen — `state.md` → `verkörpert` mit Zielort; Evidence-Dateien
bleiben als Historie.

**Folge-Slices:** keine.

**Risiken aus §6:** beide entfallen — False-Positives über die
deklarierte Ausnahmeliste begrenzt (erster Lauf kalibriert), Range-
Ermittlung über `git log --follow` gelöst.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Sensor-Bau (Greenfield): der Mechanismus ist aus den 5 Funden abgeleitet
(Familie BEO-PLAN), die Deklarations-Form folgt den done-Slice-Gegebenheiten.
Kein ADR — kein Gate gesenkt (der neue Sensor verschärft).
