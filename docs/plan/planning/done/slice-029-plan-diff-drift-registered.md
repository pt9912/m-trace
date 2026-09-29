# Slice 029: Plan-vs-Diff-Drift ins Beobachtungs-Register — BEO-PLAN

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle — Closure-Bedingung
ist die DoD dieses Slices, keine Wellen-Grenze.

**Bezug:** `docs/plan/planning/observations/`, `harness/conventions.md`
(§Modus-Deklaration), `AGENTS.md` §6, Review-Reports
`2026-09-29-slice-025.md` / `-slice-026.md` / `-slice-028.md`.

**Autor:** Owner-Auftrag nach Review-Familie „Plan-Angabe vs. realer Diff"
(3×). **Datum:** 2026-09-29.

---

## 1. Ziel und Abgrenzung

**Ziel:** Die drei Review-Fundmeldungen der Familie „Plan-Angabe vs. realer
Diff" (je erstes Finding in den Reports `2026-09-29-slice-025.md`,
`-slice-026.md`, `-slice-028.md`) als BEO-PLAN/plan-diff-drift ins
Beobachtungs-Register formalisieren (3×-Schwelle erreicht): Sub-Area `PLAN`
in der Modus-Deklaration ergänzen (Kürzel-Nachschlag-Pflicht), die
Gegenmaßnahme in `AGENTS.md` §6 verkörpern, `state.md` auf `verkörpert`
stellen.

**Ausdrücklich NICHT in diesem Slice:**

- **Mechanischen Sensor für den Drift** — die Prüfung „Plan §3 gegen realer
  Diff" ist Urteilsfall (Formulierung vs. Gebautes); Prosa-Form ist
  ausgeschöpft ab einem vierten Auftreten (modul-06).
- **Retro-Edit der drei Closed-Slice-Pläne** — done/ bleibt.

## 2. Definition of Done

- [x] `harness/conventions.md` §Modus-Deklaration um Sub-Area `PLAN` ergänzt
      (Kürzel-Nachschlag-Basis für den BEO-Pfad).
- [x] `AGENTS.md` §6, Schritt 4: Plan-Abgleich vor der Slice-Closure
      verkörpert.
- [x] `BEO-PLAN/plan-diff-drift/` angelegt: `observation.md`, `state.md`
      (Stand `verkörpert`), `evidence/slice-025.md`, `evidence/slice-026.md`,
      `evidence/slice-028.md`.
- [x] `make docs-check` grün.
- [x] Closure-Notiz.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/conventions.md` | update | Sub-Area-Zeile `PLAN` |
| `AGENTS.md` | update | §6 Schritt 4: Plan-Abgleich |
| `docs/plan/planning/observations/BEO-PLAN/plan-diff-drift/**` | neu | Register-Eintrag (3 Evidence-Dateien) |
| `docs/plan/planning/observations/README.md` | update | Eintrag sichtbar |

## 4. Trigger

- **Start (`next` → `in-progress`):** sofort — Owner-Auftrag.
- **Rückführungen:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **`state.md` `verkörpert` bei Erst-Anlage** — **Ausgang:** entfallen — die
  Verkörperung (AGENTS.md §6) wird im selben Vorgang geliefert; der
  Herkunfts-Anker ist `seit slice-029`.

## 7. Closure-Notiz

Das Beobachtungs-Register trägt seinen ersten Eintrag: `BEO-PLAN/plan-diff-
drift` (observation + state `verkörpert` + 3 Evidence-Dateien je
abgeschlossenem Vorgang). Die Gegenmaßnahme ist in `AGENTS.md` §6, Schritt 4
verkörpert (Plan-Abgleich vor der Slice-Closure); die Sub-Area `PLAN` steht
in der Modus-Deklaration als Kürzel-Nachschlag-Basis.

**Was hat funktioniert:** Die 3×-Lage war vollständig belegt — je Fund
existiert eine Evidence-Datei mit Vorgang + Fund-Satz; der Zähler
(Anzahl Evidence-Dateien) ergibt sich aus dem Dateisystem.

**Was ging anders als geplant:** Die Evidence-Dateien lagen im
`ids`-Scope — die Review-Finding-IDs (`F-1`) kollidieren dort mit der
Lastenheft-Requirement `F-1` (gleicher Token, andere Semantik) und wurden
als `id-unlinked` beanstandet. Formuliert auf „erstes Finding" statt
Token; die `F-N`-Tokens sind im `ids`-Scope nicht zu nutzen, wenn kein
Lastenheft-Bezug gemeint ist.

**Steering-Loop-Eintrag:** Guide geschärft: Evidence- und Beleg-Dateien in
`ids`-gescannten Dirs dürfen Lastenheft-Kennungs-Token (`F-N`) nur mit
Lastenheft-Bezug tragen. (Gezählt, nicht verkörpert — Erstvorkommen.)

**Beobachtungs-Register (`../observations/`):** `BEO-PLAN/plan-diff-drift`
neu angelegt, Belege `evidence/slice-025.md`, `evidence/slice-026.md`,
`evidence/slice-028.md`.

**Folge-Slices:** keine.

**Risiken aus §6:** entfallen (siehe §6 — Verkörperung im selben Vorgang).

## 8. Sub-Area-Prüfungen und Modus-Begründung

Register-Pflege (Brownfield, observable — Sub-Area `PLAN` wurde hier
angelegt). Kein ADR — keine Schwellen- oder Architekturentscheidung.
