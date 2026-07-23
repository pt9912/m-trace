# Harness

## Purpose

Dieses Verzeichnis hält die Repository-weite Präzedenz, die Konventionen und die
realen Quality-Sensoren fest, die von Menschen und Automation genutzt werden. Es
verweist auf die autoritativen Quellen und dupliziert deren Verträge nicht.

## Source precedence

Bei Konflikt zweier Quellen gewinnt die höherrangige, und die niederrangige muss
korrigiert werden:

1. `spec/lastenheft.md` (Contract)
2. Technik-Spezifikationen unter `spec/`
3. `spec/architecture.md` (abgeleitete Architektursicht)
4. akzeptierte Records unter `docs/plan/adr/`
5. `docs/plan/planning/in-progress/roadmap.md`
6. Anwender- und Betriebs-Dokumentation unter `docs/user/` und `docs/ops/`
7. Root-README-Dateien
8. diese Harness-Dokumentation

Die genaue Klassifikation und die repository-spezifischen Adaptionen sind in
[`conventions.md`](conventions.md) deklariert.

## Guides (Feedforward-Quellen)

| Guide | Rolle |
|---|---|
| [`spec/lastenheft.md`](../spec/lastenheft.md) | Contract: was m-trace bereitstellen muss |
| [`spec/architecture.md`](../spec/architecture.md) | Abgeleitete Komponenten- und Abhängigkeitssicht |
| [`docs/plan/adr/`](../docs/plan/adr/) | Begründung akzeptierter Architektur-Entscheidungen |
| [`docs/plan/planning/in-progress/roadmap.md`](../docs/plan/planning/in-progress/roadmap.md) | Lieferstatus und Sequenzierung |

## Sensors (Feedback-Gates)

| Target | Prüft |
|---|---|
| `make docs-check` | Markdown-Referenzen, Spans, tracked Targets, Code-Pfade und Dokument-Richtung |
| `make doc-trace` | Advisory-Requirements-Matrix aus den nativen Lastenheft-Tabellen und Planning-Referenzen |
| `make docs-immutable STAGED=1` | Accepted-ADR-Kern gegen den gestagten Diff |
| `make docs-commits RANGE=base..head` | Commit-Message-Traceability über einen Pull-Request-Bereich |
| `make verify-closure-notes` | Struktureller Closure-Note-Gate für neue `done/`-Pläne (ADR-0010; standalone) |
| `make gates` | Verpflichtende Repository-Quality-Gates |
| `make build` | Baubare Release-Artefakte |

Nur Befehle, die im Makefile existieren, sind hier gelistet. Der Ausführungs-
Status gehört zu CI und wird in diesem Dokument nicht festgehalten.

## Traceability rules

Requirements nutzen die bestehenden Familien `F-*`, `NF-*`, `MVP-*`, `AK-*`,
`RAK-*` und `R-*`. Architektur-Entscheidungen nutzen `ADR-NNNN`. Neue normative
Verweise zeigen von volatilen zu stabilen Quellen; Abwärts-Provenienz bleibt auf
ausgewiesene History-Abschnitte beschränkt.

Die Commit-Message-Durchsetzung gilt für Pull-Request-Bereiche. Dokumentations-,
Test-, Build-, CI- und Wartungs-Commits sind exempt; Feature- und Fix-Commits
tragen eine Requirement-, Entscheidungs- oder Plan-Kennung.

## Safety and scope boundaries

- Niemals einen höherrangigen Contract abschwächen, um einer Implementierungs-
  Drift zu entsprechen.
- Niemals ein Gate behaupten, das kein ausführbares Target hat.
- Bestehende akzeptierte ADRs sind historische Records; Änderungen erfordern den
  dokumentierten Entscheidungsprozess.

## Minimal agent workflow

1. Die höchstrangige für die Aufgabe relevante Quelle lesen.
2. Bestehenden Code und Tests prüfen, bevor Verhalten geändert wird.
3. Die Änderung an ein bestehendes Requirement, eine Entscheidung, einen Test
   oder ein Gate binden.
4. Die kleinste kohärente Änderung umsetzen.
5. Fokussierte Tests laufen lassen.
6. `make docs-check` bei Dokumentations-Änderungen laufen lassen.
7. Den proportionalen Aggregat-Gate laufen lassen.
8. Niederrangigen Status oder Anwender-Dokumentation aktualisieren, ohne den
   höherrangigen Contract implizit zu ändern.
