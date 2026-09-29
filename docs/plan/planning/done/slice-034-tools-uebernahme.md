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

- [x] `tools/harness/{slice-mv,baseline-verify,history-range-guard}.sh`
      übernommen — `slice-mv`/`history-range-guard` byte-identisch;
      `baseline-verify` mit deklarierter Abweichung (§6).
- [x] `harness/mk/{slice-mv,baseline}.mk` übernommen (GATE_CHECKS-Zeilen
      inert — m-trace hat keinen Aggregator-Verbraucher).
- [x] `Makefile`: `history-range-guard`-Target + Vorbedingung an
      `doc-immutable`/`doc-commits`.
- [x] Verifikation: `make baseline-verify` grün (54 Dateien);
      `make slice-mv` ohne Argument → Usage/Exit 2;
      `make history-range-guard RANGE=<gültige Range>` Exit 0.
- [x] `make docs-check` grün (191 Dateien, 0 Befunde).
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag (§7).

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
- **`baseline-verify.sh` weicht um genau einen sed-Filter ab** —
  Normalisierung `./`-präfixierter SHA256SUMS-Zeilen auf der Soll-Seite des
  Vollständigkeits-Vergleichs (Bootstrap slice-025 listete mit Präfix,
  Emission ohne — ohne Filter falsch-rot über alle 54 Zeilen). Owner-Entscheid:
  `.harness/baseline/**` bleibt unangetastet; die Digest-Prüfung (`sha256sum
  -c`) läuft weiter gegen die committete SHA256SUMS. **Ausgang:** deklariert
  ([Review-Report slice-034, F-4](../../../reviews/2026-09-29-slice-034.md);
  Verifier Item 1 bestätigt exakt diese eine Differenz).
- **Roadmap-Fix (+2/−1) war in §3 nicht deklariert** — planning-drift
  („Keine aktive Welle" bei liegendem Slice) erzwang die Zeile. **Ausgang:**
  eingetreten, in §7 nachgeführt.

## 7. Closure-Notiz

**Datum:** 2026-09-29. **Belege:** Review
[2026-09-29-slice-034.md](../../../reviews/2026-09-29-slice-034.md) (Erstlauf
merge-blockierend, Nachfolge `577502e` entkräftet); Verifier
[2026-09-29-slice-034-verifier.md](../../../reviews/2026-09-29-slice-034-verifier.md)
(„DoD bestätigt: ja", alle Läufe selbst gefahren, Negativ-Proben belegen
fail-closed).

- **Geliefert:** `slice-mv` + `history-range-guard` byte-identisch,
  `baseline-verify` mit deklariertem Einzeiler (§6), Fragmente unter
  `harness/mk/` (GATE_CHECKS inert), Makefile-Wiring als Vorbedingung vor
  `doc-immutable`/`doc-commits`. Die Durchsetzungsschicht (Rollen-Agenten,
  Commands, commit-msg-Träger, Selbstprüfung) ging als Owner-Entscheid
  neben dem Slice (`8856c40`) — die §1-Grenze (Stop-Hook, Command-Guard,
  Gate-Nachweis nicht adoptiert) ist eingehalten.
- **Validator:** n/a — interne Wartung, kein End-Nutzer-Wert; ausdrücklich
  übersprungen statt still.
- **Risiken:** [R-32](../risks-backlog.md#r-32) (commit-msg-Muster ohne
  MR-003-Familien) weiter offen, Triggerschwelle „vor `make hooks-install`".
  Übrige §6: entfallen bzw. deklariert.
- **Beobachtungs-Register:** keine Beobachtung angefallen.
- **Steering-Loop-Lerneintrag (geschärfte Regel, `· seit slice-034`):**
  Adoptions-Abgleich prüft Tool UND Datenseite. „Byte-identisch" blieb für
  `slice-mv`/`history-range-guard` erfüllt, während `baseline-verify` am
  Bootstrap-Datenformat scheiterte (`SHA256SUMS` mit `./`-Präfix). Die
  Reparatur liegt am Tool als deklarierte Abweichung, nie am Datenanker
  `.harness/baseline/**` — er trägt die Integrität und wird nicht
  regeneriert.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Werkzeug-Übernahme (Brownfield, byte-identisch): die Emission ist
geprüft (Review slice-031-Nachlauf über die aih-Tools), die Anpassung
beschränkt sich auf Fragment-Wiring und Makefile-Verdrahtung. Kein ADR —
kein Gate gesenkt (baseline-verify/history-range-guard verschärfen).
