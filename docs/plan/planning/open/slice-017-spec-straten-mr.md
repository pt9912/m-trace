# Slice 017: Spec-Straten — MR für m-traces Vier-Datei-Technical-Schicht

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](../welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 5 von 11).

**Bezug:** `harness/conventions.md` (neuer `MR-<NNN>`),
`spec/backend-api-contract.md`, `spec/browser-support.md`,
`spec/player-sdk.md`, `spec/telemetry-model.md` (unverändert — keine
Refactor), Regelwerk `modul-03-spec.md` §Ziel-Form: Spezifikation.

**Autor:** Nachgang Templates-Diff v3.5.1→v6.8.0, Owner-Entscheidung
`welle-02` §8 Punkt 1 (2026-09-13: „MR-Adaption deklarieren"). **Datum:**
2026-09-13.

---

## 1. Ziel

Die bestehende m-trace-Abweichung von der kanonischen Ein-Datei-Form der
Spezifikation-Schicht (`spec/spezifikation.md`) als `MR-<NNN>` in
`harness/conventions.md` deklarieren — reine Registry-Ergänzung, **kein**
Refactor der vier bestehenden Spec-Dateien.

## 2. Definition of Done

- [ ] **Neue `MR-<NNN>-spec-straten-multi-datei.md`** in
      `harness/conventions/` (aktiv, analog `MR-002`/`MR-003`/`MR-006`/
      `MR-007` aus `slice-013`): Datum, Geltungsbereich (`spec/`), **Ersetzt-
      Baseline-Regel** mit auflösbarem Anker auf
      `modul-03-spec.md#ziel-form-spezifikation` (Zitat: „ein Repo mit zwei
      Straten deklariert das als `MR-<NNN>`" — m-traces Fall ist kein
      *fehlendes* Stratum, sondern ein **mehrteiliges**: alle drei Straten
      sind vorhanden, das Technical-Stratum liegt nur auf vier Dateien statt
      einer), Adaption (vier Dateien statt einer `spezifikation.md` —
      Aufzählung), Begründung (vier fachlich getrennte Domänen —
      Backend-API, Browser-Support, Player-SDK, Telemetrie —, deren
      Zusammenführung in eine Datei ohne inhaltlichen Nutzen wäre und
      Cross-Ref-/RTM-Aufwand nach sich zöge), Auflösungs-Trigger (permanent,
      **außer** ein künftiger Slice führt die Dateien tatsächlich zusammen —
      dann löst dieser MR-Eintrag auf, nicht vorher).
- [ ] Index-Eintrag in `harness/conventions.md` §Adaptions-Block →
      **Aktive Adaptionen**-Tabelle (Anker `<a id="mr-NNN">` dort, nicht in
      der Einzeldatei — Muster aus `slice-013`).
- [ ] **Keine Änderung** an `spec/backend-api-contract.md`,
      `spec/browser-support.md`, `spec/player-sdk.md`,
      `spec/telemetry-model.md` selbst — reine Registry-Deklaration.
- [ ] `AGENTS.md` §2 Rang 2 und `harness/README.md` §Source precedence Rang 2
      verweisen auf den neuen MR (statt der bisherigen freien Prosa-Notiz
      „Baseline-Abweichung, noch nicht als `MR-<NNN>` deklariert").
- [ ] `make docs-check` grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/conventions/MR-<NNN>-spec-straten-multi-datei.md` | neu | Adaption deklarieren |
| `harness/conventions.md` §Adaptions-Block | Index-Zeile ergänzen | Aktive-Adaptionen-Tabelle |
| `AGENTS.md` §2, `harness/README.md` §Source precedence | Prosa-Notiz → MR-Verweis | Rang-2-Fußnote auf den MR ziehen statt „noch nicht deklariert" |

**Bereits geklärt:** Owner-Entscheidung liegt vor (`welle-02` §8 Punkt 1,
2026-09-13) — MR-Adaption, kein Zusammenführen der vier Dateien.

## 4. Trigger

- **`in-progress`:** nach `slice-013` (Konventions-Index/MR-Datei-Form muss
  stehen).
- **Rückführung:** keine erwartet — reine Registry-Ergänzung ohne offene
  Nebenfrage.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **Kein ADR nötig.** Reine Registry-Deklaration einer bereits bestehenden,
  gelebten Struktur — keine Architekturentscheidung, kein gesenktes Gate.
- **Keine Rückwirkung auf `spec/`-Inhalte.** Der MR dokumentiert nur die
  Datei-**Form**, nicht den Inhalt der vier Spec-Dateien.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

### Sub-Area: harness/conventions.md, AGENTS.md, harness/README.md (Werkzeug/Prozess)

Reine Registry-/Index-Ergänzung, kein Produktcode und keine Spec-Inhalte
berührt. Ohne ADR — keine Architekturentscheidung, kein Gate betroffen.
