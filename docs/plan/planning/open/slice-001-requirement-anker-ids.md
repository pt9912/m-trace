# Slice 001: Requirement-Anker + `ids`-Aktivierung

**Lifecycle:** Zustand = Verzeichnis (`open/` → `next/` → `in-progress/` →
`done/`), Wechsel nur per `git mv`.

**Welle:** ohne Welle (Harness-/Konvergenz-Arbeit; Anschluss an v3.5.0-Migration W7).

**Bezug:** `conventions.md` §Requirement-Link-Konvergenz, MR-003 (ID-Familien),
MR-007 (Planning-Form), ADR-0010-Umfeld. Erster Slice in der Kanon-Slice-Form.

**Autor:** Harness-Migration. **Datum:** 2026-07-23.

---

## 1. Ziel

Requirement-Kennungen (`F-`/`NF-`/`MVP-`/`AK-`/`RAK-` in `spec/lastenheft.md`,
`R-` in `risks-backlog.md`) einzeln adressierbar machen und das opt-in-Modul
`ids` mit **verankerten** Links scharfschalten — so, dass nackte Kennungen im
Fließtext auf ihre *Definition* zeigen, nicht auf den Datei-Anfang.

## 2. Definition of Done

- [ ] Jede der **403 Definitionen** (372 im Lastenheft, 31 im Risiko-Register)
      trägt einen inline-Anker `<a id="<kennung-kleingeschrieben>"></a>` in ihrer
      ID-Zelle (Tabellenform unverändert, RTM-`format: table` intakt).
- [ ] Alle nackten Kennungs-Mentions in **aktiver** Doku (spec, docs/user|dev|ops|
      perf, Planning-`in-progress/`, examples) sind verankerte Links
      `[ID](ziel#slug)` — Ziel je Familie (F/NF/MVP/AK/RAK → lastenheft, R →
      risks-backlog).
- [ ] `ids` in `.d-check.yml`-`modules` + per-Familie-`patterns` mit `target` und
      `exempt-paths`; `make docs-check`/`make gates` grün, **keine** Falschbefunde.
- [ ] `anchors`-Modul grün (alle `#slug` lösen auf).
- [ ] `conventions.md` §Requirement-Link-Konvergenz auf „aufgelöst/aktiviert"
      nachgezogen; Modus-Zeile „Requirement-Links" graduiert.
- [ ] `trace.slices.file-pattern` um die `slice-*`-Form erweitert (MR-007-Folge-
      Punkt — dieser Slice ist das erste `slice-*`-Artefakt).
- [ ] Closure-Notiz mit Steering-Loop-Eintrag.

## 3. Plan (vor Code)

**Ansatz A — inline-Zellen-Anker (verifiziert):** d-checks `anchors`-Modul
erkennt `<a id="x"></a>` in einer Tabellenzelle als Link-Ziel (Spike grün). Kein
generiertes Heading-Doc, kein neuer Drift-Sensor; das Lastenheft behält seine
normative Tabellenform.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `spec/lastenheft.md` | update (scripted) | 372 ID-Zellen um `<a id="…">` ergänzen |
| `docs/plan/planning/in-progress/risks-backlog.md` | update (scripted) | 31 R-ID-Zellen verankern |
| aktive Doku (spec/, docs/user\|dev\|ops\|perf/, Planning-`in-progress/`, examples/) | update (scripted) | ~635 nackte Mentions → verankerte Links |
| `.d-check.yml` | update | `ids` in `modules` + per-Familie-`patterns` (target+exempt); `trace.slices.file-pattern` erweitern |
| `harness/conventions.md` | update | §Requirement-Link-Konvergenz + Modus-Zeile graduieren |
| `scripts/` | neu (Migrations-Skript) | deterministische Anker- + Link-Transformation, reproduzierbar |

**Exempt (kein Link-Retrofit):**
- **`docs/plan/adr/**` — immutable Accepted-ADRs** (Body-Edit = Immutabilitätsbruch,
  wie W5/MR-002). ADR-Mentions bleiben nackt, `ids` nimmt ADRs aus.
- `docs/plan/planning/done/**` + `CHANGELOG.md` — historischer Audit-Bestand.

## 4. Trigger

- **`open` → `next`/`in-progress`:** Owner-Freigabe des Ansatzes (dieser Slice).
- **`in-progress` → `next`:** falls der Umfang (~1000 Edits) sich als zu grob für
  einen Slice erweist → in Teil-Slices (Anker / Links / Config) zerlegen.
- **`in-progress` → `open`:** blockiert, falls `ids`/`anchors` unerwartete
  Falschbefunde in nicht-exemptbaren Dateien liefern.

## 5. Closure-Trigger

DoD vollständig + `make gates` grün + Closure-Notiz geschrieben; `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **Skalierung/Churn:** ~403 Anker + ~635 Links über ~25 Dateien inkl. der
  normativen Spec-Straten. Mitigation: deterministisches, reviewbares Skript;
  Anker-/Link-Transformation getrennt committen; `make gates` nach jedem Schritt.
- **Anker-Slug-Kollisionen:** Kennungen sind eindeutig → Slugs (`rak-131`)
  eindeutig; Skript prüft auf Duplikate.
- **Mentions in Code-Spans / bereits verlinkt:** `ids` fasst nur nackte Prosa an
  (Inline-Code/bestehende Links bleiben) — Skript spiegelt dieselbe Regel.
- **Contract-Berührung:** Anker in Lastenheft-Zellen sind additiv (kein
  Text-/Modalitäts-Edit) → normativer Inhalt unverändert, nur adressierbar.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Requirement-Links

- **Modus:** Brownfield (Bestand vor der Link-Pflicht; `conventions.md` §Modi
  „Requirement-Links — Brownfield").
- **Konventionen-Dichte:** hoch — `conventions.md` §Requirement-Link-Konvergenz
  beschreibt die Graduierungs-Vorbedingung; MR-003 fixiert die ID-Familien.
- **Phase-Reife:** reifer Bestand (post-`0.25.0`); die Anker-Infrastruktur fehlte
  nur mechanisch.
- **Evidenz-/Diskrepanz-Risiko:** niedrig — additive Anker + Link-Wrapping, kein
  Semantik-Edit; `anchors`/`ids`/`links` verifizieren maschinell.
- **Reconciliation-Aufwand:** dieser Slice; Graduation = `ids` grün in `make gates`.
