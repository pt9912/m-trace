# Slice 018: ADR-Re-Evaluierungs-Trigger-Audit + Grandfathering (MR-009)

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](../welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 6 von 11).

**Bezug:** `docs/plan/adr/0001-*.md`–`0008-*.md` (unverändert — keine
Edits), neue `harness/conventions/MR-009-adr-trigger-grandfathering.md`,
Regelwerk `modul-04-adrs.md` §Kernidee (Modul 4).

**Autor:** Nachgang Templates-Diff v3.5.1→v6.8.0, Owner-Entscheidung
`welle-02` §8 Punkt 4 (2026-09-13: „pauschal grandfathern"). **Datum:**
2026-09-13.

---

## 1. Ziel

Audit, welche der elf bestehenden ADRs den seit `v3.5.0`-Templates
vorgesehenen `## Re-Evaluierungs-Trigger`-Abschnitt tragen, und die
fehlenden Fälle als `MR-009` pauschal grandfathern — **keine**
inhaltliche Änderung an den ADR-Bodies selbst.

## 2. Definition of Done

- [x] **Audit durchgeführt** (`grep -l "^## Re-Evaluierungs-Trigger"
      docs/plan/adr/*.md`): Ergebnis dokumentiert (welche ADRs haben den
      Abschnitt, welche nicht).
- [x] **Neue `MR-009-adr-trigger-grandfathering.md`** in
      `harness/conventions/` (aktiv): Geltungsbereich exakt die ADRs ohne
      Abschnitt (Audit-Ergebnis), Ersetzt-Baseline-Regel-Anker auf
      `modul-04-adrs.md#kernidee-modul-4`, Auflösungs-Trigger permanent
      für den Bestand.
- [x] Index-Eintrag in `harness/conventions.md` §Adaptions-Block →
      **Aktive Adaptionen**-Tabelle.
- [x] **Keine Änderung** an den ADR-Dateien selbst — reine
      Registry-Deklaration.
- [x] `make docs-check` grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/conventions/MR-009-adr-trigger-grandfathering.md` | neu | Adaption deklarieren |
| `harness/conventions.md` §Adaptions-Block | Index-Zeile ergänzen | Aktive-Adaptionen-Tabelle |

**Bereits geklärt:** Owner-Entscheidung liegt vor (`welle-02` §8 Punkt 4,
2026-09-13) — pauschal grandfathern, kein rückwirkender Trigger-Nachtrag.

## 4. Trigger

- **`in-progress`:** nach `slice-013` (Konventions-Index/MR-Datei-Form muss
  stehen).
- **Rückführung:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **Kein ADR nötig.** Reine Registry-Deklaration, keine
  Architekturentscheidung, kein gesenktes Gate, keine ADR-Body-Änderung.
- **Genauer Geltungsbereich erst nach Audit bekannt** — der in `welle-02`
  §8 vermerkte Umfang (ADR-0001..0011) war eine Vorab-Schätzung, nicht das
  Audit-Ergebnis; die MR-Datei trägt den tatsächlichen Befund.

## 7. Closure-Notiz (nach `done/`)

**Audit-Ergebnis** (`grep -l "^## Re-Evaluierungs-Trigger" docs/plan/adr/*.md`):
ADR-0009 (`harness-baseline-v3.5.0`), ADR-0010 (`closure-note-pflicht`) und
ADR-0011 (`harness-baseline-v3.5.1-bump`) tragen den Abschnitt bereits —
sie entstanden nach dessen Einführung in den `v3.5.0`-Templates. ADR-0001
bis ADR-0008 fehlt er. Der in `welle-02` §8 vermerkte Umfang
(„ADR-0001..0011") war eine Vorab-Schätzung ohne Audit-Grundlage — die
tatsächliche Grandfathering-Menge ist kleiner (acht statt elf ADRs).
`MR-009-adr-trigger-grandfathering.md` neu angelegt (analog `MR-002`),
Ersetzt-Baseline-Regel-Anker auf `modul-04-adrs.md#kernidee-modul-4`.
Index-Zeile in `harness/conventions.md` §Aktive Adaptionen ergänzt. Keine
Änderung an den ADR-Dateien selbst.

**Verifikation:** `make docs-check` — 0 Befunde (nach Marker-Rücksetzung).

**Steering-Loop-Lerneintrag:** Eine Vorab-Schätzung im Welle-Plan
(„ADR-0001..0011") deckte sich beim tatsächlichen Audit nicht mit dem
Befund — drei der elf ADRs erfüllten die neue Regel bereits, weil sie
zeitlich nach der Trigger-Einführung entstanden. Der Fehler wäre
unbemerkt geblieben, hätte die MR-Datei den geschätzten statt den
auditierten Geltungsbereich übernommen: Ein Grandfathering-Eintrag, der
mehr Dateien nennt als tatsächlich betroffen, würde später falsch lesen,
sobald jemand ADR-0009 als Positiv-Beispiel für einen vorhandenen
Trigger-Abschnitt sucht und ihn fälschlich als „grandfathered, also ohne
Abschnitt" einordnet.

**Folge-Slices:** keine unmittelbaren.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: harness/conventions.md (Werkzeug/Prozess)

Reine Registry-Ergänzung, kein Produktcode und keine ADR-Inhalte berührt.
Ohne ADR — keine Architekturentscheidung, kein Gate betroffen.
