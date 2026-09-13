# MR-007 — Planning-Artefakt-Form (Slice/Welle vs. `plan-<version>`)

- **Datum:** 2026-07-23
- **Geltungsbereich:** Planning-Artefakte unter `docs/plan/planning/`
- **Ersetzt-Baseline-Regel:** [`modul-05-planning-harness.md` §Ziel-Form:
  Slice](../../.harness/baseline/v6.8.0/regelwerk/modul-05-planning-harness.md#ziel-form-slice)
  — dort ist `slice-<Kennung>.md` (Zustand = Lifecycle-Verzeichnis) die
  einzige vorgesehene Form für neue Arbeit.
- **Adaption:** m-traces Bestand nutzt release-gekoppelte
  `plan-<version>.md`-Dateien in `done/`. **Entscheidung (Owner
  2026-07-21):** **Neue** Arbeit folgt der kanonischen Slice/Welle-Form (aus
  den vendored Templates
  `.harness/baseline/v6.8.0/templates/docs/plan/planning/{slice,welle}.template.md`).
  Der **Bestand `plan-<version>.md` wird grandfathered** (Variante A):
  historische Release-Records bleiben unverändert, keine
  Massen-Umbenennung, die Release-Versions-Kopplung bleibt für die Alt-Form.
  Brownfield-konsistent (analog [MR-002](../conventions.md#mr-002)-Grandfathering).
- **Begründung:** Ein rückwirkendes Umbenennen aller `plan-<version>.md`-Records
  auf `slice-<Kennung>.md` würde die Release-Versions-Kopplung der
  historischen Records zerstören, ohne einen praktischen Nutzen zu liefern —
  die Records sind abgeschlossen und werden nicht mehr bearbeitet.
- **Auflösungs-Trigger:** Permanent für den Bestand. `trace.slices.file-pattern`
  (`.d-check.yml`) und der Closure-Note-Glob führen additiv sowohl die
  `plan-*`- als auch die `slice-*`/`welle-*-results`-Form.
