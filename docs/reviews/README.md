# Reviews

Dieses Verzeichnis hält **Review-Reports** — das Übergabe-Artefakt
Reviewer → Implementation (v3.5.1-Regelwerk Modul 8/10).

## Konvention

- **Ein Report pro Lauf.** Folgeläufe bekommen eine neue Datei, keine
  Überschreibung (Auditierbarkeit).
- **Namensschema:** `<YYYY-MM-DD>-<slice-oder-diff-ref>.md`.
- **Gerüst:** das vendored Template
  `.harness/baseline/v3.5.1/templates/docs/reviews/review-report.template.md`
  wird kopiert-und-ausgefüllt (nicht frei formuliert).
- **Skill:** die Findings folgen dem Output-Schema von
  `.harness/skills/reviewer.md` (allgemein) bzw.
  `.harness/skills/closure-note-reviewer.md` (Closure-Notes, ADR-0010) — dort
  liegt die verbindliche Single Source of Truth für Kategorien und Felder.

Ein Review-Report ohne den Eingangs-Kontext (die Verträge, gegen die geprüft
wurde) ist nicht reproduzierbar; das Gerüst erzwingt diese Liste.

## Wann entsteht ein Report

Diese Sektion nennt den **Auslöser** (die Konvention oben nennt nur die *Form*).
Grundregel (`AGENTS.md` §5, Modul 8/10): **ein Review-Lauf produziert einen
Report** — Findings ad-hoc in Commit-Message oder Notizen abzulegen ersetzt ihn
nicht.

- **Code-Review** eines Slice-Diffs / PR-Bereichs → Report vor dem Merge
  (bzw. als Retro-Report, wenn nachträglich gegen gemergten Code geprüft wird).
- **Plan-Review** eines Slice-Plans gegen Spec/ADR → Report *vor* der
  Implementation (Rückkante Review → Plan bei Plan-Defekt).
- **Design-Review** eines Lösungs-Schnitts gegen die Architektur → Report,
  bevor die Details festgezurrt sind.

Der Auslöser ist der **Lauf**, nicht das Ergebnis: auch ein Review ohne HIGH/
MEDIUM bekommt einen Report (die Negativbefund-Zeilen sind das auditierbare
„geprüft, ohne Befund"). Es gibt bewusst **kein** automatisiertes Gate, das den
Report erzwingt (Owner-Entscheidung 2026-07-24, slice-007) — die Praxis trägt
sich über die Regel und den ersten gelebten Präzedenzfall
([`2026-07-24-slice-004.md`](2026-07-24-slice-004.md)).

## Nebenklasse: Security-Audit-Re-Reviews

Neben dem Modul-8/10-Handoff-Report hält dieses Verzeichnis auch datierte
**Security-Audit-Re-Reviews** (periodische Re-Evaluation von CVE-Suppressions,
Form: Kontext · Eingaben · Kommandos · Ergebnis · Entscheidung). Sie sind kein
Reviewer→Implementation-Übergabeartefakt, teilen aber Namensschema
(`<YYYY-MM-DD>-<ref>.md`) und die „ein Report pro Lauf"-Auditierbarkeit — daher
hier statt in einem slice-reservierten Lifecycle-Verzeichnis (slice-006).
