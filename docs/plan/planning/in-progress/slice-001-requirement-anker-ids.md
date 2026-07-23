# Slice 001: Requirement-Anker (Lastenheft) + `ids` auf Spec-Straten

**Lifecycle:** Zustand = Verzeichnis (`open/` → `next/` → `in-progress/` →
`done/`), Wechsel nur per `git mv`.

**Welle:** `welle-01-requirement-link-konvergenz`.

**Bezug:** `conventions.md` §Requirement-Link-Konvergenz, MR-003 (ID-Familien),
MR-007 (Planning-Form). Erster Slice in der Kanon-Slice-Form.

**Autor:** Harness-Migration. **Datum:** 2026-07-23.

---

## 1. Ziel

Die `F-`/`NF-`/`MVP-`/`AK-`/`RAK-`-Definitionen im Lastenheft einzeln
adressierbar machen und `ids` **auf die Spec-Straten** (`scope.roots: [spec]`)
mit **verankerten** Links scharfschalten — der erste, in sich geschlossene
Konvergenz-Schritt auf dem normativen Kern.

## 2. Definition of Done

- [ ] Alle **372** F/NF/MVP/AK/RAK-Definitionen in `spec/lastenheft.md` tragen
      einen inline-Anker `<a id="<kennung-klein>"></a>` in ihrer ID-Zelle
      (Tabellenform + RTM-`format: table` unverändert).
- [ ] Alle nackten F/NF/MVP/AK/RAK-Mentions in `spec/**` sind verankerte Links
      `[ID](lastenheft.md#slug)` (relativ je Quelldatei).
- [ ] `ids` in `.d-check.yml`-`modules` + `scope.roots: [spec]` + `patterns`
      (F/NF/MVP/AK/RAK → `spec/lastenheft.md`); `make gates` grün, keine
      Falschbefunde; `anchors` grün.
- [x] `trace.slices.file-pattern` erweitert (MR-007-Folgepunkt): `id-prefix` ist
      optional; Voll-Capture `^((?:plan|slice|welle)-.+)\.md$` mischt grandfathered
      `plan-*` (byte-identische Erkennung, verifiziert) mit `slice-*`/`welle-*`.
- [x] **Regressions-Fix:** Anker sitzen in der **Anforderungs-Zelle**, NICHT in der
      Kennungs-Zelle — sonst passt die ID-Zelle nicht „vollständig" aufs id-pattern
      und `--trace`/`doc-complete` erkennt 0 Anforderungen. (Kosmetik: der RTM-Titel
      zeigt den `<a id>`-Span; advisory, tolerierbar.)
- [ ] `conventions.md` §Requirement-Link-Konvergenz auf „Spec-Straten graduiert,
      Rest = `slice-002`" nachgezogen.
- [ ] Closure-Notiz.

## 3. Plan (vor Code)

**Ansatz A — inline-Zellen-Anker (verifiziert):** `anchors` erkennt
`<a id="x"></a>` in Tabellenzellen (Spike grün); Lastenheft behält Tabellenform,
kein generiertes Heading-Doc.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `spec/lastenheft.md` | update (Skript) | 372 ID-Zellen um `<a id>` ergänzen |
| `spec/*.md` (telemetry-model, architecture, backend-api-contract, browser-support, player-sdk, lastenheft-intern) | update (Skript) | nackte F/NF/MVP/AK/RAK-Mentions → verankerte Links |
| `.d-check.yml` | update | `ids` (modules + `scope.roots: [spec]` + patterns); `trace.slices.file-pattern` |
| `harness/conventions.md` | update | §Requirement-Link-Konvergenz teil-graduieren |
| `scripts/` | neu | deterministisches Anker-/Link-Skript, reproduzierbar |

**Bewusst NICHT in slice-001** (→ `slice-002`): `R-`-Familie + `risks-backlog.md`,
docs/user, examples, Planning-`in-progress/`. **Immer exempt:** immutable ADRs
(`docs/plan/adr/**`), `done/`, `CHANGELOG.md`.

## 4. Trigger

- **`open` → `in-progress`:** Owner-Freigabe (erfolgt), Welle-01 aktiv.
- **`in-progress` → `next`:** falls der Spec-Scope noch zu groß → in Anker-/Link-
  Teil-Slices zerlegen.
- **`in-progress` → `open`:** blockiert bei unerwarteten `ids`/`anchors`-
  Falschbefunden.

## 5. Closure-Trigger

DoD vollständig + `make gates` grün + Closure-Notiz; `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Contract-Berührung:** Anker in Lastenheft-Zellen sind additiv (kein Text-/
  Modalitäts-Edit) → normativer Inhalt unverändert, nur adressierbar.
- **Skript-Disziplin:** nur nackte Prosa-Mentions wrappen (Inline-Code/bestehende
  Links unangetastet, wie `ids` selbst); Anker- und Link-Schritt getrennt
  committen, `make gates` nach jedem.
- **Slug-Eindeutigkeit:** Kennungen eindeutig → Slugs eindeutig; Skript prüft.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Requirement-Links

- **Modus:** Brownfield (`conventions.md` §Modi „Requirement-Links").
- **Konventionen-Dichte:** hoch — §Requirement-Link-Konvergenz + MR-003.
- **Phase-Reife:** reifer Bestand (post-`0.25.0`); nur die Anker-Infrastruktur
  fehlte mechanisch.
- **Evidenz-/Diskrepanz-Risiko:** niedrig — additive Anker + Link-Wrapping, kein
  Semantik-Edit; `anchors`/`ids`/`links` verifizieren maschinell.
- **Reconciliation-Aufwand:** dieser Slice (Spec-Straten); Rest = `slice-002`.
