# 0010 — Closure-Note-Pflicht für abgeschlossene Pläne

> **Status**: **Proposed** (2026-07-22)
> **Datum**: 2026-07-22 (Proposed)
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)
> **Bezug**: [ADR-0009](0009-harness-baseline-v3.5.0.md) (v3.5.0-Baseline, W3
> zog den Review-/Closure-Harness vor); `harness/conventions.md` §Modi
> (Brownfield-Grandfathering, analog MR-002). Prozess-/Harness-ADR ohne
> Spec-Stratum-Schärfung.

## Kontext

Der v3.5.0-Regelwerk-Kanon (Modul 11 §Schritt 5) führt eine **Closure-Note**
pro abgeschlossenem Slice: ein kurzer Abschnitt, der festhält, *was der Slice
gelehrt hat*. Ohne diese Pflicht bleibt ein Slice-Abschluss ein reines
„erledigt" — das Lernsignal (warum ein Test rot war, welche Architektur-
Beobachtung auffiel, welches Folge-Slice daraus entsteht) versickert.

m-trace hat davon heute **nichts**: von 41 `plan-*.md` in `docs/planning/done/`
trägt **keiner** eine Closure-Sektion. Der Regelwerk-Skill-Satz (W3) liefert
einen `closure-note-reviewer`, dessen ganze Prämisse eine „semantische Schicht
**über** einem computational Gate" ist — dieses Gate, die Policy und die
Pflicht-Inhalte fehlen in m-trace. Owner-Entscheidung 2026-07-22: den Stack
**jetzt vollständig bauen** (gegen die aktuelle Repo-Form; der W5-Layout-Move
und die W6-Slice-Form ziehen Pfade/Verankerung später nach).

## Entscheidung

> **Entscheidung (Proposed 2026-07-22):** Jeder **neue** (nicht grandfatherte)
> abgeschlossene Plan in `docs/planning/done/` trägt eine **Closure-Note** mit
> drei Pflicht-Inhalten. Durchgesetzt zweischichtig — strukturell computational
> plus inhaltlich inferentiell.

**Drei Pflicht-Inhalte** (mindestens einer je Kategorie, alle drei erwünscht):

1. **Konkretes Lernsignal** — eine Ursache, nicht nur eine Behauptung
   („Test rot, *weil* X"), nicht „lief wie geplant".
2. **Konkretes Folge-Slice** — ein benannter Anschluss (mit `open/`-Eintrag,
   sobald die Slice-Form steht), nicht „vielleicht später".
3. **Konkrete Architektur-Beobachtung** — eine beobachtbare Aussage
   (z. B. „DB-autoritativer Sequencer verschiebt den Flaschenhals auf
   Single-PG"), kein Etikett.

Eine syntaktisch vorhandene, aber inhaltsleere Note (Floskel) ist ein
**HIGH-Verstoß**.

**Durchsetzung (zwei Schichten):**

- **Computational** — `scripts/check_closure_notes.py` hinter
  `make verify-closure-notes`: prüft *Struktur* (Closure-Heading vorhanden,
  >= 2 Sätze außerhalb Code-Blöcken, Floskel-Blockliste). Deterministisch,
  billig, netzlos (Muster wie `scripts/lint-variante-b.py`).
- **Inferentiell** — `.harness/skills/closure-note-reviewer.md`: prüft *Inhalt
  vs. Floskel* (die drei Pflicht-Inhalte), dort wo Struktur allein die Floskel
  nicht fängt. Semantische Schicht über dem computational Gate.

**Brownfield-Grandfathering.** Die 41 bestehenden `done/`-Pläne datieren vor
dieser Policy und sind über `docs/planning/.closure-grandfathered` (explizite
Enumeration, analog MR-002 für ADR-0001..0007) vom Gate ausgenommen. Ein
Rück-Backfill Dutzender historischer Pläne findet **nicht** statt. Neue Pläne
stehen nicht auf der Liste und müssen tragen; wandert künftig ein Plan nach
`done/`, ist die bewusste Reibung („Note schreiben oder mit Begründung
grandfathern") gewollt.

**Gate-Platzierung.** `make verify-closure-notes` läuft zunächst **standalone**,
**nicht** in `make gates` — solange dieses ADR `Proposed` ist. Graduierung in
`make gates` erfolgt nach Accept (Hard Rule „Gates nur per ADR", hier additiv).

## Konsequenzen

**Positiv:**

- Jeder künftige Slice-Abschluss trägt ein auditierbares Lernsignal; der
  `closure-note-reviewer`-Skill (W3) hat ein reales Prüfziel statt einer
  fingierten Infrastruktur.
- Zweischichtig: das computational Gate fängt Struktur billig, der Skill die
  semantische Floskel — kein Doppeln.
- Brownfield sauber: keine Massen-Backfill-Schuld, keine falsche „alles grün"-
  Behauptung über note-lose Altpläne.

**Kosten / Grenzen (ehrlich benannt):**

- **Gegen die aktuelle Form gebaut (Owner-Entscheidung).** Tool, Target und
  dieses ADR referenzieren `docs/planning/done/`; der W5-Layout-Move repathed
  sie auf die kanonische docs/plan/planning/done/-Form, und W6 verankert die
  drei Pflicht-Inhalte zusätzlich im Slice-Template. Das ist bewusst
  akzeptierter Migrations-Churn.
- **Immutabilität.** Nach Accept ist dieses ADR immutable; die hier genannten
  aktuellen Pfade werden dann zu einem historischen Stand (der W5-Move
  korrigiert Tooling/Config, nicht diesen ADR-Body — dieselbe Prosa-Drift wie
  bei ADR-0009 §Adaptations, gate-neutral).
- **Strukturheuristik.** Das computational Gate kann eine wortreiche Floskel
  nicht von Substanz unterscheiden — dafür ist der inferentielle Skill da; das
  Gate ist die billige erste Schicht, nicht die Wahrheit.

## Alternativen

- **A — Nur den Skill, kein Gate.** Verworfen: der Skill setzt laut Vorlage ein
  computational Gate voraus; ohne es wäre der Skill-Bezug eine Harness-Lüge, und
  die Pflicht wäre nicht deterministisch prüfbar.
- **B — Erst nach W5/W6 bauen (natürlicher Ort).** Sachlich sauberer (kanonische
  Pfade, Slice-Template steht), aber Owner wählte 2026-07-22 die sofortige
  Umsetzung gegen die aktuelle Form.
- **C — Alle 41 Altpläne backfillen statt grandfathern.** Verworfen: absurder
  Aufwand ohne Lernwert; Grandfathering ist der etablierte Brownfield-Weg
  (MR-002).

## Re-Evaluierungs-Trigger

- **Accept** schaltet die Graduierung von `make verify-closure-notes` in
  `make gates` frei.
- Der W5-Layout-Move (Repathing von `docs/planning/` auf die kanonische
  docs/plan/planning/-Form) und die W6-Slice-Form (Verankerung der drei Inhalte
  im Slice-Template).
- Dreimaliges HIGH derselben Floskel-Art → Floskel-Blockliste im computational
  Gate erweitern (billiger Feedforward statt inferentiellem Nachlesen).

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-07-22 | Proposed | ADR-0010 |
