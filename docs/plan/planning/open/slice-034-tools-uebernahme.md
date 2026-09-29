# Slice 034: aih-Tools übernehmen — slice-mv, baseline-verify, history-range-guard

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle — Closure-Bedingung
ist die DoD dieses Slices, keine Wellen-Grenze.

**Bezug:** `tools/harness/` (neu: drei Shell-Tools, byte-identisch aus der
aih-v6.13.0-Emission), `harness/mk/{slice-mv,baseline}.mk` (Fragmente),
`Makefile` (history-range-guard-Verdrahtung), Quelle:
`/tmp/aih-v6.13.0/{tools/harness,harness/mk}/`.

**Autor:** Owner-Auftrag „mach weiter" nach Übernahme-Matrix. **Datum:** 2026-09-29.

---

## 1. Ziel und Abgrenzung

**Ziel:** Die drei direkten aih-Tools übernehmen und als Make-Targets
verdrahten: `slice-mv` (Lifecycle-Move inkl. Verweis-Nachzug und
getrennten Commits — Hard Rule 3.3 als Code), `baseline-verify`
(Integrität + Vollständigkeit der vendored Baseline, netzlos) und
`history-range-guard` (fail-closed gegen grün über leerer Range in
Shallow-Klons, Vorbedingung vor `doc-immutable`/`doc-commits`). Fragmente
unter `harness/mk/` nach aih-Vorlage; der Glob-Include aus slice-032 zieht
sie automatisch.

**Ausdrücklich NICHT in diesem Slice:**

- **commit-msg-Hook + Selbstprüfung** — Enforcement ist strenger als die
  §5-Exemption (Doku-/Test-/Build-Commits); separate Abwägung.
- **Stop-Hook-Erzwingsame** (record-gates + working-tree-hash) — m-trace-
  Gates dauern ~5 min, der Slice-Flow committet docs-only dazwischen;
  proportionaler Stamp wäre eigenes Design.
- **Emit-Läufe von ai-harness-init** — der Träger ist abgelegt (slice-032);
  Emit gegen m-traces verkörperte Form bleibt eigene Entscheidung.

## 2. Definition of Done

- [ ] `tools/harness/{slice-mv,baseline-verify,history-range-guard}.sh`
      byte-identisch aus der Emission übernommen.
- [ ] `harness/mk/{slice-mv,baseline}.mk` übernommen (GATE_CHECKS-Zeilen
      inert — m-trace hat keinen Aggregator-Verbraucher).
- [ ] `Makefile`: `history-range-guard`-Target + Vorbedingung an
      `doc-immutable`/`doc-commits`.
- [ ] Verifikation: `make baseline-verify` grün (54 Dateien);
      `make slice-mv` ohne Argument → Usage/Exit 2;
      `make history-range-guard RANGE=<gültige Range>` Exit 0.
- [ ] `make docs-check` grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tools/harness/{slice-mv,baseline-verify,history-range-guard}.sh` | neu | byte-identisch aus der v6.13.0-Emission |
| `harness/mk/{slice-mv,baseline}.mk` | neu | Fragment-Verdrahtung nach aih-Vorlage |
| `Makefile` | update | history-range-guard-Target + Vorbedingung |

## 4. Trigger

- **Start (`next` → `in-progress`):** sofort — Owner-Auftrag.
- **Rückführungen:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + Verifikationsläufe grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **slice-mv committet selbst** (Move + Verweis-Nachzug) — **Ausgang:**
  entfallen — sauberer Arbeitsbaum ist Vorbedingung des Skripts (bricht
  sonst ab), im eigenen Slice nur kontrolliert gegen Testbaums genutzt.
- **GATE_CHECKS-Zeilen in den Fragmenten sind inert** (m-trace hat keinen
  aih-Aggregator) — **Ausgang:** entfallen — verbatim-Übernahme laut
  Emission, Kommentar erklärt die Herkunft.

## 7. Closure-Notiz

*(füllt bei Closure)*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Werkzeug-Übernahme (Brownfield, byte-identisch): die Emission ist
geprüft (Review slice-031-Nachlauf über die aih-Tools), die Anpassung
beschränkt sich auf Fragment-Wiring und Makefile-Verdrahtung. Kein ADR —
kein Gate gesenkt (baseline-verify/history-range-guard verschärfen).
