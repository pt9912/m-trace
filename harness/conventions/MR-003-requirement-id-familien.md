# MR-003 — Requirement-ID-Familien

- **Datum:** 2026-07-14
- **Geltungsbereich:** Contract (`spec/lastenheft.md`), Pläne, Commits und
  Reviews
- **Ersetzt-Baseline-Regel:** [`grundlagen-source-precedence.md` §ID-Schema
  als Klammer](../../.harness/baseline/v6.8.0/regelwerk/grundlagen-source-precedence.md#id-schema-als-klammer)
  — dort ist `LH-FA-<NN>`/`LH-QA-<NN>` die vorgeführte Beispielfamilie mit
  frei wählbarem Vertrags-Präfix.
- **Adaption:** m-trace datiert vor dieser Beispielfamilie und nutzt
  `F-*`, `NF-*`, `MVP-*`, `AK-*`, `RAK-*` und `R-*` als etablierte,
  produktspezifische Requirement-Familien statt eines einzelnen
  `<PREFIX>-FA-*`/`<PREFIX>-QA-*`-Paars.
- **Begründung:** Die Kennungen sind Teil des etablierten Contracts und der
  Release-Historie; eine Umstellung auf das Beispielschema würde die
  bestehende Traceability-Kette (RTM, Commits, Reviews) brechen, ohne einen
  Mehrwert zu liefern.
- **Auflösungs-Trigger:** Permanent. Neue Requirement-Familien müssen hier
  vor Nutzung deklariert werden.
