# Slice 007: Review-Report-Praxis scharf schalten

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle (Prozess/Harness).

**Bezug:** `.harness/skills/reviewer.md`, `docs/reviews/README.md`,
`.harness/baseline/v3.5.0/templates/docs/reviews/review-report.template.md`,
Regelwerk Modul 8/10 (Review-Harness), ADR-0010 (Closure-Note-Muster als
Präzedenz für „geschriebenes Artefakt statt ad-hoc"). Auslöser:
Sekundärbefund aus slice-006 (Closure-Notiz §Sekundärbefund) — das
Review-Harness ist eingerichtet (Template + zwei Skills + `docs/reviews/`),
aber ungenutzt.

**Autor:** Review-Harness-Adoption. **Datum:** 2026-07-24.

---

## 1. Ziel

Die Review-Praxis von m-trace produziert ab diesem Slice **Kanon-Review-Reports
unter `docs/reviews/`** statt Findings ad-hoc in Commit-Messages/Memory zu
verstreuen — belegt durch den ersten echten Modul-8/10-Handoff-Report und
verankert durch eine geschärfte normative Regel. **Keine neue Struktur**
(Template, Skills, Verzeichnis, README-Konvention existieren seit W3); dieser
Slice *adoptiert* das Vorhandene.

## 2. Definition of Done

- [x] Erster echter Handoff-Report aus dem vendored Template ausgefüllt:
      `docs/reviews/2026-07-24-slice-004.md` (Code-Review des slice-004-Diffs
      `bd7e3a8` gegen Plan + Konventionen; Findings + Negativbefunde + Verdikt).
- [x] AGENTS.md §5 von deskriptiv („Reports landen in `docs/reviews/`") auf
      **normativ** geschärft: ein Review-Lauf *produziert* einen Report aus dem
      Template; ad-hoc-Findings in Commit/Memory ersetzen ihn nicht.
- [x] `docs/reviews/README.md` um eine **„Wann entsteht ein Report"**-Sektion
      ergänzt (die Konvention beschrieb bisher nur das *Format*, nicht den
      Auslöser).
- [x] **Kein neues automatisiertes Gate** (Owner-Entscheidung 2026-07-24 —
      „Regel + Exemplar", nicht „+ Gate"); die Adoption trägt sich über Regel +
      gelebten Präzedenzfall, nicht über eine Fitness Function.
- [x] `make gates` grün (Doku-weite Änderung; `ids`/Link-Netz erfasst
      `in-progress/` + `docs/plan/**`).
- [x] Closure-Notiz mit Steering-Loop-Eintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `docs/reviews/2026-07-24-slice-004.md` | neu | erster echter Handoff-Report — Exemplar, das künftige Reviews nachahmen |
| `AGENTS.md` §5 | update | Review-Report-Zeile deskriptiv → normativ (Report ersetzt ad-hoc) |
| `docs/reviews/README.md` | update | neue Sektion „Wann entsteht ein Report" (Auslöser statt nur Format) |
| `docs/plan/planning/in-progress/roadmap.md` | update (bei Closure) | Slice-Kandidat „Review-Harness" → abgeschlossener Slice ohne Welle |

**Bewusst NICHT im Scope:** kein `make`-Review-Target (m-trace nutzt den
built-in `/code-review`-Skill, siehe `reviewer.md`); kein Closure-Note-artiges
Gate (Owner-Wahl); kein Backfill historischer ad-hoc-Reviews (geringer
Audit-Wert bei gemergtem, grünem Bestand).

## 4. Trigger

- **`in-progress`:** Sekundärbefund aus slice-006 steht + Owner-Schnitt
  (2026-07-24). Erfüllt.
- **Rückführung `in-progress` → `next`:** falls die Regel-Schärfung eine
  eigene Diskussion über einen automatisierten Gate auslöst → Gate-Frage als
  eigenen Slice abspalten, dieser bleibt bei Regel + Exemplar.
- **Rückführung `in-progress` → `open`:** nicht zu erwarten (kein Blocker,
  keine Carveout-Abhängigkeit).

## 5. Closure-Trigger

DoD grün + `make gates` grün + Closure-Notiz; `git mv` des Slice nach `done/`.

## 6. Risiken und offene Punkte

- **Regel ohne Gate driftet.** Ohne Fitness Function hängt die Adoption an
  Disziplin. Bewusst akzeptiert (Owner): erst der gelebte Präzedenzfall zeigt,
  ob ein Gate nötig wird. Steering-Loop-Trigger: wenn nach diesem Slice erneut
  Reviews ad-hoc landen, ist der Gate-Slice fällig (siehe Closure-Notiz).
- **Link-Netz.** Der Report liegt in `docs/reviews/` (außerhalb des
  `ids`-Scope — nackte Kennungen dort erlaubt); dieser Slice-Plan liegt in
  `docs/plan/planning/` (im `ids`-Scope — daher keine nackten
  Requirement-Kennungen im Fließtext). `make gates` verifiziert beides.

## 7. Closure-Notiz

Der Slice war klein und ging wie geplant: das Harness stand vollständig, es
fehlte nur die *Adoption*. Zwei Hebel statt Infrastruktur — ein echter
Report als Präzedenzfall und eine geschärfte Regel. Die Lehre aus slice-006
([[feedback_kanon_schweigen_keine_luecke]]) trug direkt: **erst prüfen, was
Kanon + Repo schon vorsehen**, dann adoptieren — hier existierten Template,
beide Skills, `docs/reviews/` samt README-Konvention und sogar eine (schwache)
AGENTS.md-Zeile; nichts musste erfunden werden.

**Was anders war als erwartet:** AGENTS.md §5 hatte bereits eine Review-Report-
Zeile — nur deskriptiv formuliert („Reports landen in `docs/reviews/`"), was
genau erklärt, *warum* Reviews trotzdem ad-hoc landeten. Der Fix war Schärfung,
nicht Neuschrift. Beim Schreiben des Exemplars deckte das Review selbst zwei
INFO-Befunde am slice-004-Diff auf (a-check-`_test.go`-Blind-Spot; „Wire-Format"-
Frame in `domain`) und **entkräftete einen HIGH-Verdacht** (vermeintlicher
Split-Commit ohne `port/driving`-Definitionen — die `--stat`-Ausgabe war nur
abgeschnitten, `git ls-tree` belegte die Definitionen im selben Commit). Das
demonstriert die Negativbefund-Disziplin an echtem Material.

**Steering-Loop-Eintrag:** Der Guide (AGENTS.md §5) wurde geschärft
(deskriptiv → normativ) und die `docs/reviews/README.md`-Konvention um den
*Auslöser* ergänzt. **Bewusst kein neuer Sensor** (Owner-Wahl): der
Gate-Kandidat (Slice, der nach `done/` wandert, braucht einen zugeordneten
Review-Report — analog zum ADR-0010-Closure-Note-Gate) bleibt als benannter
Folge-Punkt, falls die reine Regel driftet.

**Folge-Slices (neue `open/`-Kandidaten):** (1) **Review-Report-Gate** —
automatisierte Kopplung „Slice → Report" analog `verify-closure-notes`, nur
schneiden, wenn die Regel nachweislich driftet. (2) **Verifier-/Validator-
Skills** (Modul 8/11) — die Schwester-Rollen zum Reviewer haben noch keine
`.harness/skills/`-Datei; kein aktueller Bedarf, als Beobachtungspunkt notiert.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Review-Harness (Prozess/Doku)

- **Modus:** Brownfield (Bestand: Harness eingerichtet W3, aber ungenutzt).
- **Konventionen-Dichte:** hoch — Modul 8/10, zwei Skills, README-Konvention,
  vendored Template, ADR-0010 als Präzedenz-Muster.
- **Phase-Reife:** reifer Bestand; nur die gelebte Praxis fehlte.
- **Evidenz-/Diskrepanz-Risiko:** niedrig — reine Doku/Prozess, `docs-check`
  + `ids` verifizieren das Link-/Kennungs-Netz vollständig; das Exemplar wurde
  gegen den echten Diff per `grep`/`git ls-tree` belegt.
- **Reconciliation-Aufwand:** dieser Slice.
