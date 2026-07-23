# Closure-Note-Reviewer-Skill — m-trace

* Status: Accepted
* Bezug: ADR-0010 (Closure-Note-Pflicht), `scripts/check_closure_notes.py`,
  `AGENTS.md` §3 · <!-- d-check:ignore (ADR-/Tool-Referenzen; Pfade gelten repo-lokal) -->
* Gilt für: den *inferentiellen* Nachlauf zu `make verify-closure-notes` —
  greift dort, wo Struktur allein die Floskel nicht fängt.

## Kontext-Eingang (Pflicht)

Was der Reviewer *immer* mitbringt, bevor er urteilt:

- alle Closure-Sektionen der Pläne in `docs/plan/planning/done/`
- ADR-0010 §Entscheidung — die **drei Pflicht-Inhalte** (Lernsignal /
  Folge-Slice / Architektur-Beobachtung), gegen die geprüft wird
- ADR-0010 §Kontext — *warum* Closure-Notes existieren (Auditierbarkeit,
  Lernsignal)
- das Ergebnis von `make verify-closure-notes` für denselben Stand — was das
  Struktur-Gate bereits abgedeckt hat, wird **nicht** doppelt gemeldet
- die Grandfather-Liste `docs/plan/planning/.closure-grandfathered` — grandfatherte
  Altpläne tragen keine Pflicht und werden **nicht** geprüft

Ohne diesen Block prüft der Reviewer Text, aber nicht *gegen die drei
Pflicht-Inhalte, die eine Closure-Note tragen muss*.

## Prüf-Auftrag

Lies die Closure-Sektion jedes nicht-grandfatherten Plans in
`docs/plan/planning/done/`. Markiere alle, die *keinen* der folgenden Inhalte
tragen: (a) ein konkretes Lernsignal (z. B. „Test rot, *weil* X"), (b) ein
konkretes Folge-Slice, (c) eine konkrete Architektur-Beobachtung. Floskeln ohne
Inhalt sind ein HIGH-Finding.

Inferentiell, weil „Inhalt vs. Floskel" semantisch ist; das computational Gate
(`scripts/check_closure_notes.py`) deckt nur die Struktur (Closure-Heading,
Satzzahl außerhalb Code-Blöcken, Floskel-Blockliste).

## Klassifikation

**HIGH** — Floskel ohne Substanz: die Closure-Note ist syntaktisch vorhanden
(überlebt das Struktur-Gate), trägt aber *keinen* der drei Pflicht-Inhalte.
Beispiele: „war ganz okay, läuft jetzt", „Fertig.", „wie geplant umgesetzt".

**MEDIUM** — genau *einer* der drei Pflicht-Inhalte fehlt oder ist unkonkret:

- Lernsignal ohne das „weil X" (Behauptung statt Ursache)
- Folge-Slice benannt, aber ohne zugehörigen `open/`-Eintrag (sobald die
  Slice-Form aus W6 steht)
- Architektur-Beobachtung als Etikett statt als beobachtbare Aussage

**LOW** — alle drei Inhalte vorhanden, aber schwer nachvollziehbar formuliert
(Substanz da, Klarheit fehlt).

**INFO** — Hinweis ohne erwartete Aktion (z. B. „verweist auf ein Folge-Slice,
das noch nicht in `open/` liegt — Tracker-Nachtrag durch die Planning-Rolle").

## Was dieser Skill NICHT macht

- Keine Struktur-Prüfung (Heading vorhanden? >= 2 Sätze?) — das ist
  `scripts/check_closure_notes.py`; nicht doppeln.
- Keine Bewertung, ob der Plan *fachlich* korrekt abgeschlossen wurde —
  Verifier/Validator.
- Keine Umschreibung der Closure-Note — der Autor formuliert nach, der Reviewer
  kategorisiert nur.
- Keine Prüfung grandfatherter Altpläne oder von Plänen außerhalb `done/`
  (`open/`/in-progress tragen noch keine Pflicht-Closure).

## Output-Schema

Jedes Finding:

- `kategorie`: HIGH | MEDIUM | LOW | INFO
- `quelle`: `ADR-0010` | `Closure-Inhaltspflicht (a/b/c)`
- `pfad`: `docs/plan/planning/done/<plan>.md:<Zeile>`
- `befund`: *welcher* der drei Pflicht-Inhalte fehlt, 1–2 Sätze, beobachtbar,
  ohne Formulierungs-Vorschlag
- `verifizierbar`: nein — Floskel-Erkennung ist inferentiell;
  `check_closure_notes.py` bestätigt nur die Struktur, nicht den Inhalt

Zusätzlich am Ende: eine Zeile „geprüft, ohne Befund: `done/<Charge>`" pro
betrachteter Plan-Charge (Negativbefund — macht die Abdeckung sichtbar).

## Pflege (Steering-Loop)

Bei dreimaligem HIGH derselben Floskel-Art:

- Muster in ADR-0010 bzw. `AGENTS.md` §Closure als benanntes Anti-Pattern
  aufnehmen
- prüfen, ob `scripts/check_closure_notes.py` die Floskel *strukturell* fangen
  könnte (Floskel-Blockliste erweitern) → dann ins computational Gate heben; ein
  computational Feedforward-Marker ist billiger als inferentielles Nachlesen
- ADR-0010 §Entscheidung schärfen, falls die drei Pflicht-Inhalte nicht klar
  genug abgefragt werden

Diese Skill-Datei wird **nicht** überschrieben, sondern versioniert.
