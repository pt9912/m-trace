# Reviewer-Skill — m-trace

* Status: Accepted
* Bezug: `AGENTS.md` §3 (Harte Regeln), `docs/plan/adr/` (aktive ADRs),
  `harness/conventions.md` (MR-001..MR-004) · <!-- d-check:ignore (ADR-/Skill-Referenzen; Anker gelten repo-lokal) -->
* Gilt für: Code-/Plan-/Design-Review-Läufe. m-trace hat **kein** dediziertes
  `make`-Review-Target; Einstieg ist der `/code-review`-Skill (Working-Diff)
  bzw. `/code-review ultra` (Cloud, PR/Branch). Gate-Bezug: `make gates`.

## Kontext-Eingang (Pflicht)

Was der Reviewer *immer* mitbringt, bevor er den Diff liest:

- Diff des PR / der Änderung
- `spec/lastenheft.md` (für referenzierte `F-*`/`NF-*`/`MVP-*`/`AK-*`/`RAK-*`-IDs)
- ADRs unter `docs/plan/adr/`, deren ID im PR oder in der Commit-Message vorkommt
- `AGENTS.md` §3 (Harte Regeln)
- vorherige Findings am gleichen Modul (letzte ~5 PRs)

Ohne diesen Block sieht der Reviewer den Code, aber nicht *die Verträge, gegen
die er prüft*.

## Klassifikation

Jeder Anker HIGH/MEDIUM/LOW hat eine *konkrete* Liste. INFO ist bewusst kurz
(Ergänzungs-Kanal, nicht Hauptkanal).

**HIGH** — eines der folgenden:

- ADR-Verstoß (Hexagon-Layer-Import, Tool, Hard Rule) — Layer-Regel prüft
  `make arch-check` (`scripts/check-architecture.sh`)
- Sicherheits-Anti-Pattern (Injection, fehlende Auth-Prüfung, ungültig
  validierte XFF/`client_ip`-Ableitung)
- Korrektheitsfehler im *kritischen* Pfad: Ingest-Sequencer/Event-Persistenz,
  Auth-Token-Validierung, Rate-Limiter (Fairness/fail-open)
- Suppression eines Gates (`//nolint`, `eslint-disable`, ad-hoc CVE-Ignore
  außerhalb `.security/vulnignore.yaml`) ohne ADR — Hard Rule 3.2
- **`git mv` + Inhaltsänderung in einem Commit** statt zwei — Hard Rule 3.3
  (bricht die Rename-Detection)
- **Inhaltliche Änderung an einer Accepted-ADR** statt neuer ADR mit
  `Supersedes` — Hard Rule 3.5 (`make docs-immutable` fängt den Kern)

**MEDIUM** — eines der folgenden:

- unklare Fehlerbehandlung am Rand des Spec-Bereichs
- fehlende Negativtests bei neuem öffentlichem Vertrag (API-Kontrakt, SDK)
- Variante-B-Drift in Kommentaren/Specs (Plan-/Tranche-/§-Verweis statt
  Kennung/Link) — Hard Rule 3.7, `make lint-variante-b`
- Wiederholung eines Musters, das schon zweimal LOW war

**LOW** — stilistisch unschön ohne semantische Auswirkung, einmalige Tippfehler,
unbenutzte Imports.

**INFO** — Hinweis ohne erwartete Aktion (z. B. „hier gäbe es ein passendes
Smoke-Target, das der Diff nicht nutzt").

## Was dieser Skill NICHT macht

- Keine Lösungsvorschläge („schreib das so") — Reviewer kategorisiert,
  Implementer entscheidet.
- Kein Refactoring-Vorschlag, der über den Diff hinausgeht.
- Keine Verifikation gegen DoD — das ist Verifier-Aufgabe.
- Keine Closure-Note-Inhaltsprüfung — dafür der Schwester-Skill
  `.harness/skills/closure-note-reviewer.md` (ADR-0010).

Wenn etwas auffällt, das in diese Kategorien gehört: ein INFO-Finding mit
Verweis auf die zuständige Rolle.

## Output-Schema

Jedes Finding:

- `kategorie`: HIGH | MEDIUM | LOW | INFO
- `quelle`: ADR-ID, `RAK-*`/`R-*`-ID, Hard-Rule-Nummer oder „Maintainability"
- `pfad`: `Datei:Zeile`
- `befund`: 1–2 Sätze, beobachtbar, ohne Lösungsvorschlag
- `verifizierbar`: ja/nein — gibt es einen Gate-Lauf, der es bestätigen würde?

Zusätzlich am Ende: eine Zeile „geprüft, ohne Befund" pro betrachtetem
Verzeichnis (Negativbefund-Zeile — sonst ist „keine Findings" nicht von „nicht
geprüft" unterscheidbar). Report-Gerüst für den ganzen Lauf:
`.harness/baseline/v3.5.1/templates/docs/reviews/review-report.template.md`
kopiert-und-ausgefüllt nach `docs/reviews/`, ein Report pro Lauf, Folgeläufe als
neue Datei statt Überschreibung.

## Pflege (Steering-Loop)

Bei dreimaligem Auftreten desselben Findings:

- ist die Kategorie noch richtig? → Klassifikation schärfen
- gibt es einen ADR/`AGENTS.md`-Eintrag, der das verhindert hätte?
  → Folge-ADR oder `AGENTS.md`-Update
- gibt es eine Fitness Function, die das prüfen würde? → Gate hinzufügen

Diese Skill-Datei wird **nicht** überschrieben, sondern versioniert.
