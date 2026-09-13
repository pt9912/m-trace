# MR-001 — Repository-Pfade

- **Datum:** 2026-07-14 (angelegt), 2026-07-23 (aufgelöst)
- **Geltungsbereich:** ADR- und Planning-Verzeichnisse
- **Ersetzt-Baseline-Regel (historisch):** das kanonische Layout
  `docs/plan/adr/`, `docs/plan/planning/` (siehe
  [`modul-05-planning-harness.md`](../../../.harness/baseline/v6.8.0/regelwerk/modul-05-planning-harness.md)
  und die ADR-Konvention in
  [`modul-04-adrs.md`](../../../.harness/baseline/v6.8.0/regelwerk/modul-04-adrs.md)).
- **Adaption (historisch):** m-trace nutzte `docs/adr/` statt
  `docs/plan/adr/` und `docs/planning/` statt `docs/plan/planning/`.
- **Begründung (historisch):** Etabliertes öffentliches Repository-Layout
  mit umfangreichen stabilen Links.
- **Auflösung:** Die v3.5.0-Migration W5 (Layout-Move,
  [`plan-harness-v3.5.0-migration.md`](../../../docs/plan/planning/done/plan-harness-v3.5.0-migration.md))
  hat das Repo auf das Kanon-Layout gehoben. Die immutablen Accepted-ADRs
  blieben unangetastet — ihre Pre-Move-Verweise sind per
  `ignore-refs`-Tombstone in `.d-check.yml` grandfathered
  (Frozen-Doc-Refactoring). Damit ist die Pfad-Divergenz beseitigt; diese
  Adaption ist geschlossen.
- **Auflösungs-Trigger:** Eingetreten (kein Nachfolger-MR — aufgelöst durch
  Slice-Arbeit, siehe `harness/conventions.md` §Aufgelöste Adaptionen).
