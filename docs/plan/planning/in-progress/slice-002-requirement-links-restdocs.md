# Slice 002: `ids`-Scope-Ausweitung (R-Familie + übrige aktive Doku)

**Lifecycle:** Zustand = Verzeichnis. **Welle:** `welle-01-requirement-link-konvergenz`.

**Bezug:** `conventions.md` §Requirement-Link-Konvergenz (Stufe 2), MR-003, MR-007.
Baut auf `slice-001` (Anker-Infrastruktur + `ids`-Config-Muster).

**Autor:** Harness-Migration. **Datum:** 2026-07-23.

---

## 1. Ziel

`ids` von den Spec-Straten auf die **übrige aktive Doku** ausweiten und die
**R-Familie** (Risiko-Register) aufnehmen — damit ist `ids` repo-weit scharf und
`welle-01` schließt.

## 2. Definition of Done

- [ ] 31 R-Definitionen in `risks-backlog.md` verankert (Anker in der letzten
      Zelle, wie im Lastenheft — RTM-neutral).
- [ ] `ids`-Scope ausgeweitet: `scope.roots` = spec + docs/user|dev|ops|perf +
      examples + docs/plan/planning + docs/plan/carveouts; R-Familie →
      `risks-backlog.md`.
- [ ] Alle nackten Mentions in diesem Scope verankerte Links; `make gates` grün,
      keine Falschbefunde (inkl. `matrix`), `anchors`/`links` grün, `--trace` ok.
- [ ] `conventions.md` §Requirement-Link-Konvergenz voll graduiert; Modus-Zeile
      „Requirement-Links" auf Greenfield.
- [ ] Closure-Notiz; danach `welle-01`-Closure (`welle-01-results.md`).

## 3. Plan (vor Code)

Gleiches Muster wie slice-001: `add_requirement_anchors.py` für R-Anker, dann
d-check `--repair` (datei-level) → Upgrade-Skript auf verankerte Links.

**Matrix-Constraint (verifiziert vor dem Verlinken):** R-Mentions in `spec/**`
(Contract/Technical/View) dürfen **nicht** nach `risks-backlog.md` (Planning)
verlinkt werden — das wäre `from: contract/technical/view, to: planning` =
`matrix-forbidden`. Daher trägt die **R-Familie `exempt-paths: [spec/**]`**; die
R-Mentions im Lastenheft/Spec bleiben bewusst nackt (Schicht-Richtung: der
Vertrag verweist nicht abwärts aufs Risiko-Register).

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `risks-backlog.md` | update (Skript) | 31 R-Anker |
| docs/user, examples, docs/plan/planning\|carveouts, docs/perf | update (Skript) | nackte Mentions → verankerte Links |
| `.d-check.yml` | update | `ids.scope.roots` weiten; R-Familie + `exempt-paths` |
| `harness/conventions.md` | update | §Requirement-Link-Konvergenz voll graduieren |

**Dauerhaft exempt:** immutable ADRs (`docs/plan/adr/**`, nicht in scope.roots),
`done/`, `CHANGELOG.md`, Root-Übersichts-`*.md` (nicht in scope.roots).

## 4. Trigger

- **`in-progress`:** slice-001 grün (erfüllt), welle-01 aktiv.
- **Rückführung:** bei Matrix-/Scope-Falschbefunden zurück zu `next` (Scope enger).

## 5. Closure-Trigger

DoD grün + `make gates` + `--trace` ok + Closure-Notiz; `git mv` nach `done/`;
danach welle-01-Closure.

## 6. Risiken und offene Punkte

- **Matrix-Richtung:** siehe §3 — R-Familie exempt in `spec/**`.
- **Churn in Planning-Docs:** roadmap/risks-backlog/migration-plan/triage werden
  aktiv editiert; verankerte Links sind ab jetzt Pflicht (das ist der Zweck).
- **RTM-Titel-Kosmetik:** R-Anker in risks-backlog-Zellen erscheinen im
  `--trace`-Titel (risks-backlog ist keine RTM-Quelle → nur kosmetisch).

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Requirement-Links

- **Modus:** Brownfield → Greenfield (Abschluss der Graduierung).
- **Konventionen-Dichte:** hoch (§Requirement-Link-Konvergenz, MR-003).
- **Phase-Reife:** reifer Bestand; Anker-Infrastruktur aus slice-001 vorhanden.
- **Evidenz-/Diskrepanz-Risiko:** niedrig (additiv, maschinell verifiziert) —
  einziger Sonderfall die Matrix-Richtung (§3, vorab geprüft).
- **Reconciliation-Aufwand:** dieser Slice; danach ist die Sub-Area Greenfield.
