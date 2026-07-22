# ADR-Index — <Projektname>

> **Template-Hinweis.** Vorlage für `docs/plan/adr/README.md`. Kopiere nach
> `docs/plan/adr/README.md`, ersetze `<Platzhalter>` und lösche diesen Block.
> **Derivativ:** Quelle der Wahrheit sind die ADR-Dateien; dieser Index ist
> eine Bequemlichkeits-Sicht — bei jedem neuen/akzeptierten ADR mitziehen.

| ID | Titel | Status | Bezug |
|---|---|---|---|
| [<NNNN>](<NNNN>-<titel>.md) | <Titel der Entscheidung> | Proposed \| Accepted | `<LH-FA-NN>` |

## Konventionen

- ADRs sind nach `Accepted` **immutable** (siehe [Kurs Modul 4](https://github.com/pt9912/ai-harness-course/blob/v3.5.0/kurs/de/01-spec-und-architektur/modul-04-architektur-adrs.md)).
- Schärfungen entstehen als neue ADR mit `Supersedes ADR-NNNN`.
- Bei `Accepted`: diesen Index aktualisieren (Status, Datum).
- Jede ADR deklariert im `**Schärft:**`-Feld *aufwärts*, welche Spec-Stelle
  sie verbindlich macht ([Kurs §Referenz-Richtung](https://github.com/pt9912/ai-harness-course/blob/v3.5.0/kurs/de/grundlagen/konventionen.md#referenz-richtung-sdp-wer-darf-wen-referenzieren)).
  Prozess-ADRs ohne Spec-Stratum tragen `—`.
