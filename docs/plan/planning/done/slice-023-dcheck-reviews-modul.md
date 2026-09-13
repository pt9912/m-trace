# Slice 023: d-check-Modul `reviews` aktivieren

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 10 von 11).

**Bezug:** `.d-check.yml` (`modules:`, neu `reviews:`), Regelwerk
Templates (`.d-check.yml`-Vorbild), Modul 10 (Review-Harness).

**Autor:** Nachgang Templates-Diff v3.5.1→v6.8.0, Owner-Entscheidung
„nur `reviews` aktivieren, `targets` zurückstellen" (AskUserQuestion,
2026-09-13). **Datum:** 2026-09-13.

---

## 1. Ziel

Das d-check-Modul `reviews` (Review-Report-Deckung: jeder `done/`-Slice
mit einem „Review"-DoD-Haken braucht einen passenden Report unter
`docs/reviews/`) in `.d-check.yml` aktivieren.

**Ausdrücklich NICHT in diesem Slice:** der `targets`-Sensor
(Deklarations-Konsistenz Doku↔Makefile) — zurückgestellt, siehe
`welle-02` §8 Punkt 5.

## 2. Definition of Done

- [x] **`.d-check.yml`**: `reviews:` Sektion ergänzt (`done-dir:
      docs/plan/planning/done`, `reviews-dir: docs/reviews`), `reviews`
      in die `modules:`-Liste aufgenommen.
- [x] **Isoliert getestet** (vor dem Einbau, 2026-09-13): `--enable
      reviews` gegen den aktuellen Bestand — 0 Befunde. Bestehende
      Review-Praxis seit `slice-007` deckt sich bereits mit dem Sensor.
- [x] **`make docs-check` grün** mit dem neuen Modul im regulären Lauf
      (nicht nur isoliert getestet).
- [x] `harness/README.md` §Sensors: `make docs-check`-Zeile bleibt
      unverändert (Modul-Liste ist in `.d-check.yml` dokumentiert, nicht
      in der Sensors-Tabelle einzeln aufgeführt — Präzedenzfall: `planning`,
      `ids`, `matrix` etc. sind dort auch nicht einzeln genannt).
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.d-check.yml` | `reviews:`-Sektion + `modules:`-Eintrag | Modul aktivieren |

**Bereits geklärt (2026-09-13):** Isolierter Test gegen den Bestand lief
mit 0 Befunden — Aktivierung ist risikoarm.

## 4. Trigger

- **`in-progress`:** sofort, keine Abhängigkeiten.
- **Rückführung:** falls der reguläre `make docs-check`-Lauf (anders als
  der isolierte Test) unerwartete Befunde zeigt — dann Ursache klären,
  bevor committet wird.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **Kein ADR nötig.** Reine Sensor-Aktivierung, kein Gate gesenkt (im
  Gegenteil: schärfer).
- **`targets`-Sensor bleibt offen** (siehe §1) — kein Blocker dieses
  Slices.

## 7. Closure-Notiz (nach `done/`)

`reviews` zur `modules:`-Liste in `.d-check.yml` hinzugefügt, `reviews:`
Sektion (`done-dir`/`reviews-dir`) ergänzt. Isolierter Test vor dem
Einbau (0 Befunde) und der reguläre `make docs-check`-Lauf danach
stimmen überein — kein Überraschungs-Fund. `harness/README.md` unverändert
(Modul-Liste steht in `.d-check.yml`, nicht in der Sensors-Tabelle
einzeln aufgeführt, wie bei den übrigen `modules:`-Einträgen auch).

**Verifikation:** `make docs-check` — 1 Befund, der erwartete transiente
`planning-drift` (dieser Slice liegt selbst in `in-progress/`).

**Steering-Loop-Lerneintrag:** Ein Sensor, der beim isolierten Test 0
Befunde zeigt, ist nicht automatisch risikofrei beim Einbau — der reale
Unterschied ist der volle Modul-Satz gleichzeitig (Wechselwirkungen
zwischen Modulen sind zwar unwahrscheinlich, aber ungetestet, bis der
reguläre Lauf tatsächlich läuft). Beide Läufe stimmten hier überein;
dasselbe Zwei-Stufen-Vorgehen (isoliert, dann regulär) lohnt sich auch
für künftige Modul-Aktivierungen.

**Folge-Slices:** `targets`-Sensor (zurückgestellt, `welle-02` §8 Punkt
5) — kein Slice geschnitten, da kein akuter Bedarf.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: .d-check.yml (Werkzeug/Prozess)

Reine Sensor-Konfiguration, kein Produktcode berührt. Ohne ADR — keine
Architekturentscheidung, kein Gate gesenkt.
