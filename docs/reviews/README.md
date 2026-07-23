# Reviews

Dieses Verzeichnis hält **Review-Reports** — das Übergabe-Artefakt
Reviewer → Implementation (v3.5.0-Regelwerk Modul 8/10).

## Konvention

- **Ein Report pro Lauf.** Folgeläufe bekommen eine neue Datei, keine
  Überschreibung (Auditierbarkeit).
- **Namensschema:** `<YYYY-MM-DD>-<slice-oder-diff-ref>.md`.
- **Gerüst:** das vendored Template
  `.harness/baseline/v3.5.0/templates/docs/reviews/review-report.template.md`
  wird kopiert-und-ausgefüllt (nicht frei formuliert).
- **Skill:** die Findings folgen dem Output-Schema von
  `.harness/skills/reviewer.md` (allgemein) bzw.
  `.harness/skills/closure-note-reviewer.md` (Closure-Notes, ADR-0010) — dort
  liegt die verbindliche Single Source of Truth für Kategorien und Felder.

Ein Review-Report ohne den Eingangs-Kontext (die Verträge, gegen die geprüft
wurde) ist nicht reproduzierbar; das Gerüst erzwingt diese Liste.

## Nebenklasse: Security-Audit-Re-Reviews

Neben dem Modul-8/10-Handoff-Report hält dieses Verzeichnis auch datierte
**Security-Audit-Re-Reviews** (periodische Re-Evaluation von CVE-Suppressions,
Form: Kontext · Eingaben · Kommandos · Ergebnis · Entscheidung). Sie sind kein
Reviewer→Implementation-Übergabeartefakt, teilen aber Namensschema
(`<YYYY-MM-DD>-<ref>.md`) und die „ein Report pro Lauf"-Auditierbarkeit — daher
hier statt in einem slice-reservierten Lifecycle-Verzeichnis (slice-006).
