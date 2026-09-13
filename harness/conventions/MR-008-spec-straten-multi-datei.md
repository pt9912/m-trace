# MR-008 — Spec-Straten: Technical-Schicht auf vier Dateien

- **Datum:** 2026-09-13
- **Geltungsbereich:** `spec/backend-api-contract.md`, `spec/browser-support.md`,
  `spec/player-sdk.md`, `spec/telemetry-model.md`
- **Ersetzt-Baseline-Regel:** [`modul-03-spec.md` §Ziel-Form:
  Spezifikation](../../.harness/baseline/v6.8.0/regelwerk/modul-03-spec.md#ziel-form-spezifikation)
  — dort ist `spec/spezifikation.md` (eine Datei) die einzige vorgesehene
  Form für das Technical-Stratum; „ein Repo mit zwei Straten deklariert das
  als `MR-<NNN>`" behandelt das *Fehlen* eines Stratums, nicht dessen
  Aufteilung auf mehrere Dateien — m-traces Fall ist der letztere: alle drei
  Straten sind vorhanden, das Technical-Stratum liegt nur nicht in einer
  Datei.
- **Adaption:** Das Technical-Stratum liegt auf vier fachlich getrennten
  Dateien statt einer `spec/spezifikation.md`: `backend-api-contract.md`
  (API-Vertrag), `browser-support.md` (Browser-/Client-Matrix),
  `player-sdk.md` (SDK-Oberfläche), `telemetry-model.md`
  (Telemetrie-Schema). Alle vier folgen den operativen Regeln des Technical-
  Stratums unverändert (fortschreibbar, präzisieren-nie-erweitern, kein
  ADR-Verweis, `SPEC-<NNN>`-Kennungen für nicht-requirement-gebundene
  Festlegungen).
- **Begründung:** Die vier Domänen (Backend-API, Browser-Support,
  Player-SDK, Telemetrie) sind fachlich unabhängig genug, dass eine
  Zusammenführung in eine Datei keinen inhaltlichen Nutzen brächte, aber
  Cross-Referenzierung und RTM-Pflege erschweren würde. Ein Refactor auf
  eine Datei wäre ein erheblicher, produktnaher Eingriff (RTM, Cross-Refs,
  `matrix`-Modul) ohne erkennbaren Gegenwert (Owner-Entscheidung, siehe
  [`welle-02`](../../docs/plan/planning/welle-02-regelwerk-v6.8.0-migration.md)
  §8 Punkt 1).
- **Auflösungs-Trigger:** Permanent — außer ein künftiger Slice führt die
  vier Dateien tatsächlich zu einer `spezifikation.md` zusammen.
