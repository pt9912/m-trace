# Welle <welle-id>: <Titel>

> **Template-Hinweis.** Vorlage für eine Welle (Bündel von Slices, das
> gemeinsam geplant und abgeschlossen wird, siehe
> [Kurs Modul 5](https://github.com/pt9912/ai-harness-course/blob/v3.5.1/kurs/de/02-planung/modul-05-planning-harness.md)
> und [Modul 6](https://github.com/pt9912/ai-harness-course/blob/v3.5.1/kurs/de/02-planung/modul-06-roadmap.md)).
> Kopiere nach `docs/plan/planning/<welle-id>.md` und ersetze
> Platzhalter. Lösche diesen Block.

**Lifecycle:** Die aktive Welle liegt flach unter `docs/plan/planning/`; bei
Closure wandert diese Datei per `git mv` nach `done/` (neben ihre
`welle-<NN>-results.md`). Der Zustand ist die Verzeichnis-Position — kein
Status-Feld. Ob eine flache Welle *aktuell* oder *geplant* ist, sagt die Roadmap.

**Zielmeilenstein:** M<NN> oder "kein Meilenstein-Bezug".

**Verantwortlich:** <Name>. **Datum:** YYYY-MM-DD.

---

## 1. Welle-Ziel

<!--
Was liefert die Welle? Eine Aussage, die sich an einem Lasttest oder
Akzeptanzkriterium spiegelt.
-->

<…>

## 2. Trigger (Welle startet)

<!--
Was muss vorher passiert sein? Verweise auf vorangegangene Wellen
oder externe Ereignisse.
-->

- <z.B. Welle <welle-vorher-id> done.>
- <z.B. ADR-<NNNN> accepted.>

## 3. Closure-Trigger (Welle schließt)

<!--
Was muss erreicht sein, damit die Welle done ist? Aktion, nicht
Termin.
-->

- <z.B. Alle Slices done.>
- <z.B. `make fullbuild` grün.>
- <z.B. Replay-Lauf gegen Golden Set durchläuft.>
- <z.B. Closure-Notiz in `welle-<NN>-results.md`.>

## 4. Slices in dieser Welle

<!-- Zustand jedes Slice = sein Lifecycle-Verzeichnis (open/next/in-progress/
done), hier NICHT gespiegelt — eine Status-Spalte driftete gegen die
Verzeichnisse (dieselbe zweite Wahrheit, die beim Slice retired wurde). -->

| Slice | Titel | Bezug |
|---|---|---|
| slice-<NN-A> | <…> | LH-FA-<NN> |
| slice-<NN-B> | <…> | LH-FA-<NN> |

## 5. Abhängigkeiten

<!--
Welche Wellen kommen *nach* dieser? Falls jemand sie ändert, was
bricht?
-->

- Blockiert: Welle <welle-id> (wegen <Vertragspunkt>).
- Wird blockiert von: Welle <welle-id>.

## 6. Out-of-Scope für diese Welle

<!--
Explizite Nicht-Inhalte. Schützt vor Scope-Creep.
-->

- <…>

## 7. Closure-Notiz

<!-- Erst nach Welle-Abschluss füllen. Verweis auf welle-<NN>-results.md. -->
