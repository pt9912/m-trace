# MR-002 — Accepted-ADR-Grandfathering

- **Datum:** 2026-07-14
- **Geltungsbereich:** `docs/plan/adr/0001-*.md` bis `docs/plan/adr/0007-*.md`
- **Ersetzt-Baseline-Regel:** [`modul-04-adrs.md` §Hard Rule für
  Accepted-ADRs](../../.harness/baseline/v6.8.0/regelwerk/modul-04-adrs.md#hard-rule-für-accepted-adrs)
  — dort gilt uneingeschränkt: „Eine ADR mit Status `Accepted` wird nicht
  inhaltlich überschrieben."
- **Adaption:** Die Hard Rule gilt für diese sieben Vor-Adoptions-Records
  nur eingeschränkt: Sie enthalten historische Plan-Provenienz außerhalb
  eines ausgewiesenen History-Abschnitts (eine Form, die die Baseline für
  neue ADRs nicht vorsieht) — diese Altlast wird nicht rückwirkend bereinigt.
  Die Immutabilität selbst bleibt uneingeschränkt: keine dieser sieben ADRs
  wird inhaltlich verändert.
- **Begründung:** Akzeptierte ADRs sind unter der adoptierten Baseline
  immutable und werden nicht allein zum Nachrüsten der Konvention
  umgeschrieben.
- **Auflösungs-Trigger:** Permanente historische Ausnahme. Neue ADRs
  erhalten keine Ausnahme; künftige Accepted-ADR-Änderungen werden vom
  ADR-Immutabilitäts-Sensor blockiert.
