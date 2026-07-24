# <Projektname>

> **Template-Hinweis.** Vorlage für das Projekt-Root-`README.md`. Kopiere
> nach `README.md` deines Repos, ersetze `<Platzhalter>` und lösche diesen
> Block. Das README ist **Rang 7** der Source Precedence (Projekt-Überblick)
> — es *verweist* auf die kanonischen Quellen, es *dupliziert* sie nicht.
> Tipp: oft zuletzt in Phase 1 füllen, wenn die verlinkten Artefakte stehen.
> Hintergrund: [Kurs Modul 2 / Harness-Bootstrap](https://github.com/pt9912/ai-harness-course/blob/v3.5.1/kurs/de/01-spec-und-architektur/modul-02-harness-bootstrap.md).

## Was ist <Projektname>?

<!-- 2–3 Sätze: was leistet es, für wen, gegen welche Annahme. Überblick,
nicht Implementierung (die lebt in spec/). -->

<…>

## Was kann ich heute tun?

<!-- Ehrlicher Ist-Stand — was JETZT läuft, nicht was geplant ist. Konkrete
Befehle/Fähigkeiten. Phase-bewusst: keine Erfolgsmeldung ohne lauffähigen
Beleg. -->

- <z. B. `make gates` läuft grün>
- <z. B. Befehl X liefert Y>

## Warum <Projektname>?

<!-- Welche Lücke / welcher Schmerz? Warum existiert es, was wäre die
Alternative? Ein Absatz. -->

<…>

## Kerngedanke

<!-- Die eine Leitidee / das Designprinzip in 1–2 Sätzen. Woran sich jede
Entscheidung messen lässt. -->

<…>

## Was macht es vertrauenswürdig?

<!-- Die Harness-Signale, auf die sich Mensch und Agent verlassen. Pointer
auf die kanonischen Quellen — Inhalt nicht wiederholen. -->

- **Prozess:** [`AGENTS.md`](AGENTS.md) (Hard Rules), [`harness/README.md`](harness/README.md) (Source Precedence, Gates).
- **Verträge:** [`spec/lastenheft.md`](spec/lastenheft.md) (`LH-*`-IDs mit Akzeptanzkriterien).
- **Gates:** <welche Sensors laufen — nur existierende nennen (keine halluzinierten Gates, Modul 13)>.
- **Auditierbarkeit:** Entscheidungen in `docs/plan/adr/`, Planung in `docs/plan/planning/`.
