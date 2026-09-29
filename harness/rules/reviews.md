# Review-Läufe

Review-Läufe (Code/Plan/Design) folgen den Skills unter
[`.harness/skills/`](../../.harness/skills/)
(`reviewer.md`, `closure-note-reviewer.md`) und **produzieren einen
Report** aus dem vendored Template
([`review-report.template.md`](../../.harness/baseline/v6.13.0/templates/docs/reviews/review-report.template.md))
unter [`docs/reviews/`](../../docs/reviews/) — ein Report pro Lauf, Folgeläufe
als neue Datei. Ad-hoc-Findings in Commit-Message oder Notizen **ersetzen
den Report nicht** (Auditierbarkeit). Wann ein Report fällig ist, steht in
[`docs/reviews/README.md`](../../docs/reviews/README.md).
