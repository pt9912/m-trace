# Planning — <Projektname>

> **Template-Hinweis.** Vorlage für `docs/plan/planning/README.md`. Kopiere
> nach `docs/plan/planning/README.md`, ersetze `<Platzhalter>` und lösche
> diesen Block. **Derivativ:** dokumentiert die Konvention; Quelle der
> Wahrheit sind die Dateien in den Verzeichnissen selbst.

Slice-Lifecycle: `open/` → `next/` → `in-progress/` → `done/`.

Reine `git mv`-Commits beim Wechsel zwischen Verzeichnissen — siehe Hard
Rule „git mv + Inhaltsänderung = zwei Commits" in
[`../../../AGENTS.md`](../../../AGENTS.md).

## Lifecycle-Bedeutungen

| Verzeichnis | Bedeutung |
|---|---|
| `open/` | Geplant, noch nicht priorisiert. Keine Garantie auf Umsetzung. |
| `next/` | Als Nächstes priorisiert. Verantwortlicher zugeordnet. |
| `in-progress/` | Branch / PR existiert. |
| `done/` | DoD erfüllt, gemerged, Closure-Notiz vorhanden. |

## Slices vs. Wellen — zwei Status-Mechanismen

- **Slices** tragen ihren Status über das **Verzeichnis** (open → … → done).
- Eine **Welle** (Bündel von Slices) wird **in der Roadmap** geführt
  ([`in-progress/roadmap.md`](in-progress/roadmap.md): Meilensteine, Wellen,
  aktive Welle); ihr Status lebt im `Status:`-Feld, nicht im Verzeichnis. Ein
  optionaler Welle-Plan liegt **flach** in `planning/` (`<welle-id>.md`) — die
  Lifecycle-Verzeichnisse sind **slice-reserviert**. Welle-Closure: Lerneintrag
  in `done/<welle-id>-results.md`.

## Aktueller Stand

Nicht als Snapshot hier eintragen — der Stand ergibt sich aus den
`open/`/`next/`/`in-progress/`/`done/`-Verzeichnissen (optional ein
`plan-status`-Target wie im Kurs-`lab/example`), sonst driftet die Tabelle.

## Roadmap

Siehe [`in-progress/roadmap.md`](in-progress/roadmap.md).
