# Slice 019: Beobachtungs-Register etablieren

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](../welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 7 von 11).

**Bezug:** neu `docs/plan/planning/README.md`,
`docs/plan/planning/observations/README.md`,
`harness/conventions.md` §Modus-Deklaration pro Sub-Area (Kürzel-Spalte),
`harness/README.md` §Guides, Regelwerk `modul-06-roadmap.md` §Das
Beobachtungs-Register, `modul-05-planning-harness.md` §Ziel-Form:
Sub-Area-Modus-Begründung (Beobachtungs-Register-Sichtung als
vorgelagerter Schritt).

**Autor:** Nachgang Templates-Diff v3.5.1→v6.8.0. **Datum:** 2026-09-13.

---

## 1. Ziel

Das Beobachtungs-Register als Ablage-Konvention etablieren
(`docs/plan/planning/observations/BEO-<KUERZEL>/<slug>/`, drei Dateien —
`observation.md`/`state.md`/`evidence/`), den dafür nötigen
Sub-Area-Kürzel-Zeiger in `harness/conventions.md` ergänzen, und die bisher
fehlende `docs/plan/planning/README.md` (Lifecycle-Konvention,
Slices-vs.-Wellen, Register-Zeiger) anlegen — sie dokumentiert die
Planning-Konvention als Ganzes, wovon das Register nur ein Teil ist.
**Kein** mechanischer `d-check`-Sensor in diesem Slice — der aktuell
gepinnte `d-check v0.75.0` kennt kein `planning.observations`-Opt-in
(Audit unten).

## 2. Definition of Done

- [ ] **Kürzel-Spalte** in `harness/conventions.md`
      §Modus-Deklaration ergänzt — **nur** für die
      `BEO-<KUERZEL>/<slug>`-Pfadform (Owner-Entscheidung
      `welle-02`-Nachtrag, 2026-09-13), die bestehende Kennungs-Zählung
      ohne Bereichssegment (`MR-000`) bleibt unverändert.
- [ ] **`docs/plan/planning/observations/README.md`** neu angelegt —
      Register-Erklärung, Zeiger auf
      `.harness/baseline/v6.8.0/templates/docs/plan/planning/observation.template.md`,
      Zustand „keine offenen Beobachtungen" (kein Verzeichnis unterhalb
      nötig, solange nichts beobachtet wurde).
- [ ] **`harness/README.md` §Guides** bekommt eine Zeile für
      `docs/plan/planning/observations/`.
- [ ] **`docs/plan/planning/README.md`** neu angelegt (fehlte bisher
      vollständig, gefunden beim Schneiden dieses Slices) — Lifecycle-
      Bedeutungen, Slices-vs.-Wellen, Register-Zeiger (Beobachtungs-Register
      **und** die Feststellung, dass m-trace kein `reconciliation.md`
      führt — kein Brownfield-Bootstrap-Rückbau durchlaufen), Aktueller-
      Stand-Hinweis (kein Snapshot), Roadmap-Zeiger.
- [ ] **`d-check`-Audit dokumentiert:** `--print-config` des gepinnten
      `v0.75.0` geprüft — `planning:` kennt `roadmap`/`closure`/`waves`,
      **kein** `observations`-Schlüssel. Mechanische Durchsetzung bleibt
      manuell (Slice-Closure §7/§8), bis ein d-check-Release das
      Opt-in liefert — **kein Blocker** dieses Slices, nur dokumentierte
      Grenze.
- [ ] `make docs-check` grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/conventions.md` §Modus-Deklaration | Kürzel-Spalte ergänzen | Beobachtungs-Pfadform braucht einen nachschlagbaren Kürzel |
| `docs/plan/planning/observations/README.md` | neu | Register-Ablage muss existieren, auch leer |
| `harness/README.md` §Guides | Zeile ergänzen | Register ist eine Feedforward-Quelle für §8 der Slice-Planung |
| `docs/plan/planning/README.md` | neu | Planning-Verzeichnis hatte nie eine eigene Konventions-Dokumentation |

**Bereits geklärt:** Owner-Entscheidung liegt vor (AskUserQuestion,
2026-09-13) — Kürzel-Spalte nur für Beobachtungs-Pfade, nicht für
ADR-/Slice-/Welle-Zählung.

## 4. Trigger

- **`in-progress`:** nach `slice-013` (Konventions-Index muss stehen).
- **Rückführung:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **Kein ADR nötig.** Reine Konventions-/Ablage-Ergänzung, keine
  Architekturentscheidung, kein Gate gesenkt oder neu geschaffen.
- **Mechanische Durchsetzung fehlt bewusst** — solange `d-check` kein
  `planning.observations`-Modul kennt, ist die Register-Pflege
  Disziplin, kein Gate. Folge-Punkt für einen künftigen Tool-Upgrade-Slice,
  kein Blocker hier.
- **Tranche 11 hängt an dieser Tranche** (Review-Report-`Klasse`-Spalte
  braucht das Register als Ziel für den Steering-Loop-Zähler) —
  unverändert gültig nach diesem Slice.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

### Sub-Area: harness/conventions.md, docs/plan/planning/ (Werkzeug/Prozess)

Reine Konventions-/Ablage-Ergänzung, kein Produktcode berührt. Ohne ADR —
keine Architekturentscheidung, kein Gate betroffen (die spätere
mechanische Durchsetzung, falls sie kommt, wäre ein eigener Slice mit
eigener Modus-Begründung).
