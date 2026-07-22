# Review-Report: <slice-NN | PR-Ref> — <YYYY-MM-DD>

> **Template-Hinweis.** Vorlage für einen Review-Report (das
> Übergabe-Artefakt Reviewer → Implementation, Modul 8/10). Kopiere
> nach `docs/reviews/<YYYY-MM-DD>-<slice-oder-diff-ref>.md`, ersetze
> `<Platzhalter>` und lösche diesen Block. Ein Report pro Lauf —
> Folgeläufe bekommen eine neue Datei, keine Überschreibung
> (Auditierbarkeit).

**Review-Art:** Plan | Design | Code — *wogegen* geprüft wird:
Plan-Review gegen Spec/ADR, Design-Review gegen Architektur,
Code-Review gegen Plan + Konventionen (Modul 10 §Drei Review-Arten).

**Gegenstand:** <Slice-ID / Diff-Range / Commit-Hash>

**Skill:** `.harness/skills/reviewer.md` @ <Version/Commit> · <!-- d-check:ignore (Adopter-spezifischer Skill-Pfad, existiert im Ziel-Repo ggf. nicht) -->
**Modell:** <Modell-ID> · **Datum:** <YYYY-MM-DD>

**Eingangs-Kontext** (die Verträge, gegen die geprüft wurde — ohne
diese Liste ist der Lauf nicht reproduzierbar):

- <Slice-Plan / Plan-Dokument>
- <aktive ADRs, z. B. ADR-<NNNN>>
- <berührte `LH-*`-IDs>
- `AGENTS.md` (Hard Rules)

---

## Findings

Jedes Finding folgt dem **§Output-Schema des Reviewer-Skills** — der
verbindlichen Single Source of Truth. Die Felder unten sind nur
**gespiegelt** (Bequemlichkeit beim Ausfüllen), nicht neu definiert; bei
Abweichung gilt der Skill bzw. dessen Quelle
[Kurs Modul 10 §Output-Schema](https://github.com/pt9912/ai-harness-course/blob/v3.5.0/kurs/de/04-qualitaet/modul-10-review-harness.md#worked-example-eine-reviewer-skill-datei-schreiben).

<!-- Kein Fließtext, kein Lösungsvorschlag im Befund. -->

### F-1 — <Kurztitel>

- `kategorie`: HIGH | MEDIUM | LOW | INFO
- `quelle`: <ADR-ID, LH-ID, Hard-Rule-Name oder "Maintainability">
- `pfad`: <Datei:Zeile>
- `befund`: <1–2 Sätze, beobachtbar, ohne Lösungsvorschlag>
- `verifizierbar`: ja/nein — <welcher Gate-Lauf würde es bestätigen?>

## Negativbefunde

<!--
Eine Zeile pro betrachtetem Bereich. Ohne diesen Block ist "keine
Findings" nicht von "nicht geprüft" unterscheidbar (Modul 10
§Reviewer berichtet auch, was er nicht gefunden hat).
-->

- geprüft, ohne Befund: <Verzeichnis/Bereich>
- geprüft, ohne Befund: <Verzeichnis/Bereich>

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | <n> |
| MEDIUM | <n> |
| LOW | <n> |
| INFO | <n> |

## Verdikt

**Merge-blockierend:** ja | nein — HIGH und MEDIUM blockieren
typischerweise; eine Abweichung davon wird hier begründet, nicht
still entschieden.

**Übergabe:** Findings gehen an die Implementation (Rückkante
Review → Plan bei Plan-Defekt). Der Report ersetzt keine
Verifikation — DoD-/Spec-Konformität prüft der Verifier separat
(Modul 11; anderes Prüf-Artefakt, anderer Eingabe-Kontext).
