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

- [ ] Sensor-Target existiert und ist an die Gate-Kette angebunden (Target-
      Name im Closure-Entscheid; d-check-Modul oder Skript+Target).
- [ ] Sensor meldet über die bestehenden done-Slices (025–032) die
      bekannten Drift-Fälle reproduzierbar (Verifikation gegen die
      Review-Fundliste).
- [ ] Lifecycle-Ausnahmen deklariert und begrenzt: Marker-Tanz (Roadmap),
      Slice-Dokument selbst, Review-Report-Dateien.
- [ ] `make gates` grün mit dem neuen Sensor in der Kette.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag; BEO-PLAN `state.md` auf
      `verkörpert` (Zielort + `seit slice-033`).

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| Sensor (Ablageort im Closure-Entscheid: d-check-Modul oder `scripts/`) | neu | Plan-§3-Pfade gegen Slice-Range-Diff |
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

*(füllt bei Closure)*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Sensor-Bau (Greenfield): der Mechanismus ist aus den 5 Funden abgeleitet
(Familie BEO-PLAN), die Deklarations-Form folgt den done-Slice-Gegebenheiten.
Kein ADR — kein Gate gesenkt (der neue Sensor verschärft).
