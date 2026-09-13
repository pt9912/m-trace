# Slice 024: Review-Harness-Templates nachziehen

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](../welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 11 von 11 — letzte Tranche dieser Welle).

**Bezug:** `.harness/skills/reviewer.md`, `docs/reviews/README.md`,
Regelwerk `modul-10-review-harness.md`, Templates
(`docs/reviews/review-report.template.md`,
`.harness/skills/reviewer.template.md` — beide vendored unter
`.harness/baseline/v6.8.0/templates/`, kein lokales Duplikat).

**Autor:** Nachgang Templates-Diff v3.5.1→v6.8.0. **Datum:** 2026-09-13.

---

## 1. Ziel

`.harness/skills/reviewer.md` (m-traces einzige lokale Review-Artefakt-
Datei — der Report selbst nutzt das vendored Template direkt, kein
lokales Duplikat) gegen das `v6.8.0`-Template nachziehen: neue
`klasse`-Spalte im Output-Schema, zwei neue HIGH-Fundklassen, Zitier-
Form-Disziplin für Baseline-Zitate. `docs/reviews/README.md` auf
denselben Stand + dieselbe Zitier-Form-Disziplin gebracht.

## 2. Definition of Done

- [ ] **`.harness/skills/reviewer.md` HIGH-Liste** um zwei neue Klassen
      ergänzt: „Norm nur im Template-Kommentar" und „Kommentar oder
      Zustandsfeld trägt Chronik statt Zustand" (Hard Rule 3.7, AGENTS.md
      — m-trace zitiert die eigene, bereits codifizierte Hard Rule statt
      der generischen Baseline-Regel, da die Norm hier schon verkörpert
      ist).
- [ ] **Output-Schema** bekommt `klasse`-Feld (stabile Kurz-Bezeichnung,
      speist den Steering-Loop-Zähler — Beobachtungs-Register aus
      `slice-019`).
- [ ] **Zitier-Form-Disziplin** durchgezogen: Baseline-Zitate als `vX.Y.Z`
      · `regelwerk/<datei>.md` §<Abschnitt> in Inline-Code statt Link
      (Report-Gerüst-Verweis, `quelle`-Feld-Beispiel).
- [ ] **Stale Referenzen gefixt** (gefunden beim Diff-Abgleich, nicht Teil
      des Template-Deltas selbst): `MR-001..MR-004` → generisches
      `MR-<NNN>`; `scripts/check-architecture.sh`-Erwähnung entfernt
      (retired seit `slice-003`); `.harness/baseline/v3.5.1/` → `v6.8.0`
      in beiden Dateien; **Hard-Rule-Nummerierungs-Bug**: die
      Variante-B-MEDIUM-Zeile zitierte noch „Hard Rule 3.7" — seit
      `slice-014` ist Variante-B §3.8 (§3.7 ist jetzt die neue
      Kommentar-Regel); korrigiert.
- [ ] **`docs/reviews/README.md`**: `v3.5.1` → `v6.8.0`-Referenzen,
      Findings-Tabellenform + `Klasse`-Spalte + Zitier-Form-Disziplin
      ergänzt (§Konvention).
- [ ] `make docs-check` grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.harness/skills/reviewer.md` | HIGH-Liste + Output-Schema + Zitier-Form + Stale-Ref-Fixes | Template-Delta + gefundene Bugs |
| `docs/reviews/README.md` | Konvention nachgezogen | Findings-Form + Zitier-Form-Disziplin |

**Bereits geklärt:** `docs/reviews/review-report.template.md` und
`.harness/skills/reviewer.template.md` selbst brauchen **keine** lokale
Kopie-Pflege — m-trace nutzt beide direkt aus dem vendorten
`v6.8.0`-Baum (bereits durch Tranche 1 aktuell). Nur die aus dem Template
**abgeleitete** lokale Datei (`reviewer.md`) und der darauf verweisende
README-Text brauchen Nacharbeit.

## 4. Trigger

- **`in-progress`:** nach `slice-019` (Beobachtungs-Register muss stehen
  — die `klasse`-Spalte referenziert es).
- **Rückführung:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **Kein ADR nötig.** Reine Skill-/Konventions-Nachführung, kein Gate
  gesenkt oder neu geschaffen.
- **Bestehende Review-Reports werden nicht retrofitted** — die neue
  Findings-Form gilt ab diesem Slice für neue Reports, historische
  Reports (`2026-05-13`…`2026-08-16`) bleiben in ihrer damaligen Form
  (Auditierbarkeit, kein rückwirkendes Umschreiben).
- **Damit ist `welle-02` vollständig**: alle elf Tranchen `done`. Welle-
  Closure (§9, `welle-02-results.md`) ist der nächste, eigenständige
  Schritt — nicht Teil dieses Slices.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

### Sub-Area: .harness/skills/ (Werkzeug/Prozess)

Reine Skill-/Konventions-Nachführung, kein Produktcode berührt. Ohne
ADR — keine Architekturentscheidung, kein Gate gesenkt.
