# Slice 013: `harness/conventions.md` von Adaptions-Container zu Index umbauen

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](../welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 2 von 10).

**Bezug:** `harness/conventions.md`, neu `harness/conventions/MR-<NNN>-*.md`
(+ `harness/conventions/done/`), `.harness/baseline/v6.8.0/templates/harness/conventions.template.md`,
`.harness/baseline/v6.8.0/templates/harness/conventions/MR-NNN-titel.template.md`.

**Autor:** Nachgang Templates-Diff v3.5.1→v6.8.0. **Datum:** 2026-09-13.

---

## 1. Ziel

`harness/conventions.md` von einem Container mit sieben inline ausformulierten
MR-Einträgen (230 Zeilen) zu einem **reinen Index** umbauen — jede Adaption
eine eigene Datei unter `harness/conventions/MR-<NNN>-<slug>.md`, aufgelöste
Adaptionen per `git mv` nach `harness/conventions/done/`. Grund laut neuem
Regelwerk: `harness/conventions.md` liest **jeder** Agentenlauf — aufgelöste
Adaptionen gehören nicht in diesen Lesepfad.

**Blockiert Tranche 10** (d-check-Sensor `vcs` prüft MR-Datei-Immutability —
ohne Einzeldateien nichts zu prüfen) und ist Voraussetzung für einen sauberen
Tranche-3/4-Umbau (beide verweisen auf die neue Konventions-Form).

## 2. Definition of Done

- [ ] `harness/conventions.md` enthält nur noch: Zweck, Baseline-Block (mit
      `**Stand:**` als **Version**, nicht Datum — Pflicht für den
      `versions`-Sensor aus Tranche 10), MR-000-Adoptionserklärung (bleibt
      inline, ist keine Adaption), zwei Tabellen „Aktive Adaptionen" /
      „Aufgelöste Adaptionen" mit `<a id="mr-<NNN>">`-Ankern **in der
      Index-Zeile** (nicht in der Einzeldatei — der Anker muss den `git mv`
      nach `done/` überleben).
- [ ] Sieben Einzeldateien `harness/conventions/MR-001-*.md` … `MR-007-*.md`
      angelegt, Inhalt aus dem bisherigen `### MR-NNN`-Abschnitt übernommen
      (Pflichtfelder: Datum, Geltungsbereich, Ersetzt-Baseline-Regel,
      Adaption, Begründung, Auflösungs-Trigger — bereits heute vorhanden,
      nur Feldnamen ggf. angleichen).
- [ ] `MR-001` (bereits inline als „AUFGELÖST 2026-07-23" markiert) liegt in
      `harness/conventions/done/MR-001-*.md`; die Index-Tabelle „Aufgelöste
      Adaptionen" führt die Zeile (siehe §3, offene Frage zur
      „aufgelöst durch"-Spalte).
- [ ] Alle Querverweise auf `harness/conventions.md#mr-NNN`
      (`grep -rn "conventions.md#mr-" `) bleiben gültig — die Anker wandern
      nicht, weil sie im Index bleiben.
- [ ] `make docs-check` grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/conventions.md` | umschreiben | Container → Index (Struktur unten) |
| `harness/conventions/MR-001-repository-pfade.md` | neu, dann `git mv` nach `done/` | aufgelöst seit 2026-07-23 |
| `harness/conventions/MR-002-accepted-adr-grandfathering.md` | neu | aktiv, permanent |
| `harness/conventions/MR-003-requirement-id-familien.md` | neu | aktiv, permanent |
| `harness/conventions/MR-004-wsl-host-pfad-beispiele.md` | neu | aktiv, permanent |
| `harness/conventions/MR-005-nicht-slice-register-flache-platzierung.md` | neu | aktiv, permanent |
| `harness/conventions/MR-006-security-gate-carveout-registry.md` | neu | aktiv, je-CVE-Trigger |
| `harness/conventions/MR-007-planning-artefakt-form.md` | neu | aktiv, permanent für Bestand |

**Neue Index-Struktur** (aus dem `v6.8.0`-Template `harness/conventions.template.md`
direkt übernommen, nicht nacherzählt):

```
## Baseline
- **Stand:** v6.8.0        <!-- Version, kein Datum -->
- **Regelwerk + Templates:** committet vendored unter .harness/baseline/v6.8.0/ …
- **Datum der Adoption:** …

## Adaptionen

### MR-000 — Baseline-Aussage
<bleibt inline: keine Adaption, sondern Adoptionserklärung>

### Aktive Adaptionen

| MR | Titel | Geltungsbereich | Ersetzt-Baseline-Regel |
|---|---|---|---|
| [002](conventions/MR-002-accepted-adr-grandfathering.md) <a id="mr-002"></a> | Accepted-ADR-Grandfathering | docs/plan/adr/0001-0007 | … |
| … (003–007) | … | … | … |

### Aufgelöste Adaptionen

| MR | aufgelöst durch |
|---|---|
| [001](conventions/done/MR-001-repository-pfade.md) <a id="mr-001"></a> | slice-006 (v3.5.0-Migration W5, kein Nachfolger-MR) |
```

**Bereits geklärt (Recherche vor Schnitt, 2026-09-13):** Template-Diff
`.harness/baseline/v3.5.1/templates/harness/conventions.template.md` gegen
`v6.8.0` direkt gezogen (nicht nur aus Fork-Zusammenfassung übernommen).
Neue ID-Zusätze im Template (`SPEC-<NNN>`, `ARC-<NNN>`, `BEO-<NNN>`,
`RC-<NNN>`, `slice-<Kennung>` statt `slice-<NNN>`) werden hier **nicht**
übernommen — das ist Tranche 5/7-Territorium (Spec-Straten-MR bzw.
Beobachtungs-Register), dieser Slice bewegt nur die Container-Form.
`Kürzel`-Spalte (für Mehr-Schreiber-Zählräume) entfällt — m-trace hat einen
schreibenden Menschen plus Agent, kein Bereichssegment nötig.

**Offen für die Implementierung:** Die Template-Tabellenform
„aufgelöst durch" nimmt an, dass eine Adaption durch eine **andere MR**
abgelöst wird (Retirement-Kopfmarker-Muster „ÜBERHOLT: … → MR-NNN", beobachtet
im Referenz-Repo `ai-harness-init`). `MR-001` wurde aber durch **Slice-Arbeit**
(die v3.5.0-Migration), nicht durch eine Nachfolger-MR aufgelöst — die Spalte
braucht hier einen Verweis auf `slice-006`/die Migration statt auf eine MR.
Passt das Template-Feld dafür, oder ist eine abweichende Formulierung nötig?
Beim Umsetzen entscheiden, keine Vorab-Annahme.

## 4. Trigger

- **`in-progress`:** nach `slice-012` (Baseline muss vendort sein, bevor
  diese Tranche den neuen Wortlaut zitiert).
- **Rückführung:** keine erwartet — reine Umstrukturierung bestehenden
  Inhalts, kein neuer Adaptions-Entscheid.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + alle `#mr-NNN`-Querverweise
verifiziert + Closure-Notiz + `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Anker-Disziplin ist der ganze Punkt dieses Slices.** Wandert ein Anker
  versehentlich in die Einzeldatei statt in die Index-Zeile, bricht jeder
  externe Verweis beim nächsten `done/`-Move. Vor Abschluss: `grep -rn
  "conventions.md#mr-"` gegen den gesamten Baum, jede Fundstelle einzeln
  gegenprüfen.
- **Kein ADR nötig.** Reine Formumstellung bestehender, bereits akzeptierter
  Adaptionen — keine neue inhaltliche Abweichung.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Harness-Konventionen (Werkzeug/Prozess)

Reine Struktur-Migration bestehenden Doku-Inhalts, kein Produktcode, keine
Spec, kein Requirement berührt. Ohne ADR — keine neue Adaption, nur eine
andere Ablageform für bestehende.
