# MR-009 — ADR-Re-Evaluierungs-Trigger-Grandfathering

- **Datum:** 2026-09-13
- **Geltungsbereich:** `docs/plan/adr/0001-*.md` bis `docs/plan/adr/0008-*.md`
- **Ersetzt-Baseline-Regel:** [`modul-04-adrs.md` §Kernidee (Modul
  4)](../../.harness/baseline/v6.8.0/regelwerk/modul-04-adrs.md#kernidee-modul-4)
  — dort gilt: „Jede ADR trägt einen Re-Evaluierungs-Trigger … oder
  ausdrücklich *permanent*", umgesetzt als eigener
  `## Re-Evaluierungs-Trigger`-Abschnitt (Ziel-Form ADR/MADR).
- **Adaption:** ADR-0001 bis ADR-0008 tragen keinen
  `## Re-Evaluierungs-Trigger`-Abschnitt — sie entstanden vor dessen
  Einführung in den Templates. ADR-0009, -0010 und -0011 haben den
  Abschnitt bereits (entstanden nach Adoption der `v3.5.0`-Baseline, die
  ihn schon vorsah). Die acht Bestands-ADRs werden **pauschal
  grandfathered**, analog [MR-002](../conventions.md#mr-002)
  (ADR-Pfad-Grandfathering): Kein Trigger-Abschnitt wird nachgetragen — das
  wäre ein inhaltlicher Edit an `Accepted`-ADRs (AGENTS.md §3.5, Hard Rule
  für Accepted-ADRs) allein zum Nachrüsten der Konvention. Alle **neuen**
  ADRs ab diesem Slice tragen den Abschnitt verpflichtend.
- **Begründung:** Ein rückwirkender Trigger-Nachtrag würde entweder die
  Hard Rule verletzen (Edit an `Accepted`-Inhalt) oder acht neue Trigger
  erfinden, die zum Zeitpunkt der ursprünglichen Entscheidung nicht
  formuliert wurden — beides ohne echten Auditierbarkeits-Gewinn
  gegenüber einer klaren Registry-Ausnahme.
- **Auflösungs-Trigger:** Permanente historische Ausnahme für ADR-0001..
  0008. Neue ADRs erhalten keine Ausnahme; der ADR-Immutabilitäts-Sensor
  (`make docs-immutable`) bleibt unverändert scharf.
