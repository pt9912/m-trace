# Slice 023: d-check-Modul `reviews` aktivieren

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](../welle-02-regelwerk-v6.8.0-migration.md)
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

- [ ] **`.d-check.yml`**: `reviews:` Sektion ergänzt (`done-dir:
      docs/plan/planning/done`, `reviews-dir: docs/reviews`), `reviews`
      in die `modules:`-Liste aufgenommen.
- [ ] **Isoliert getestet** (vor dem Einbau, 2026-09-13): `--enable
      reviews` gegen den aktuellen Bestand — 0 Befunde. Bestehende
      Review-Praxis seit `slice-007` deckt sich bereits mit dem Sensor.
- [ ] **`make docs-check` grün** mit dem neuen Modul im regulären Lauf
      (nicht nur isoliert getestet).
- [ ] `harness/README.md` §Sensors: `make docs-check`-Zeile bleibt
      unverändert (Modul-Liste ist in `.d-check.yml` dokumentiert, nicht
      in der Sensors-Tabelle einzeln aufgeführt — Präzedenzfall: `planning`,
      `ids`, `matrix` etc. sind dort auch nicht einzeln genannt).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

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

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

### Sub-Area: .d-check.yml (Werkzeug/Prozess)

Reine Sensor-Konfiguration, kein Produktcode berührt. Ohne ADR — keine
Architekturentscheidung, kein Gate gesenkt.
