# Slice 008: Regelwerk-Baseline-Bump v3.5.0 → v3.5.1

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle (Harness-Wartung).

**Bezug:** [ADR-0011](../../adr/0011-harness-baseline-v3.5.1-bump.md)
(nicht-struktureller Re-Vendor; Aufwärts-Referenz), ADR-0009
(Re-Evaluierungs-Trigger), `harness/conventions.md` §Baseline. Auslöser:
Kurs-Release v3.5.1 (Welle 33, 2026-07-23).

**Autor:** Baseline-Wartung. **Datum:** 2026-07-24.

---

## 1. Ziel

Die aktive Regelwerk-Baseline steht auf **v3.5.1**: v3.5.1 vendored + integritäts-
geprüft, alle lebenden Zeiger umgehängt, die vendored v3.5.0-Baseline als
Audit-Referenzform behalten. Kein Layout-/Pfad-Umbau (nicht-struktureller Bump,
ADR-0011).

## 2. Definition of Done

- [x] `.harness/baseline/v3.5.1/{regelwerk,templates}/` vendored aus
      `lab-regelwerk.zip` (Archiv-sha256 `7268a8e6…`), erzeugtes `SHA256SUMS`
      (43 Dateien), `sha256sum -c` grün.
- [x] `.harness/baseline/v3.5.0/` **bleibt** liegen (Owner-Entscheidung
      2026-07-24) — die `done/`-Records + immutablen ADRs lösen darauf netzlos auf.
- [x] Lebende Zeiger auf v3.5.1: `harness/conventions.md` §Baseline (Version +
      Archiv-sha + Datum + Retention-Hinweis), `AGENTS.md`, `CLAUDE.md`,
      `docs/plan/carveouts/README.md`, `docs/reviews/README.md`,
      `.harness/skills/reviewer.md`.
- [x] Historische Eigennamen unberührt (`v3.5.0-Migration W5`, die datierten
      MR-Blöcke, `plan-harness-v3.5.0-migration.md`, done/-Records, immutable ADRs).
- [x] ADR-0011 als **Proposed** angelegt (wartet auf Owner-Accept).
- [x] `make gates` grün. Closure-Notiz.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.harness/baseline/v3.5.1/**` | neu (vendored) | aktive Baseline + `SHA256SUMS` |
| `docs/plan/adr/0011-…-v3.5.1-bump.md` | neu | Pin-Fortschreibung (ADR-0009 immutable) |
| `harness/conventions.md` §Baseline | update | aktive Version v3.5.1, Retention v3.5.0 |
| `AGENTS.md`, `CLAUDE.md` | update | Baseline-Pfad-/Versions-Zeiger auf v3.5.1 |
| `docs/plan/carveouts/README.md`, `docs/reviews/README.md`, `.harness/skills/reviewer.md` | update | Template-/Kanon-Zeiger auf v3.5.1 |

**Repath-Regel:** nur `.harness/baseline/v3.5.0/`-Pfade + der CLAUDE.md-Versions-
Satz + die Kanon-Zitate der ohnehin berührten Live-Docs wandern auf v3.5.1;
historische Eigennamen und die §Baseline-Retention-Prosa bleiben (v3.5.0 existiert
weiter, daher kein Link-Bruch). Verifikation: `make docs-check` (volles Link-Netz)
+ `grep` auf verbleibende Live-`baseline/v3.5.0/`-Pfade (soll leer).

## 4. Trigger

- **`in-progress`:** v3.5.1-Release steht + Delta assessed (ADR-0011). Erfüllt.
- **Rückführung `in-progress` → `next`:** falls das Delta doch strukturell wäre
  (Layout/Pfad) → als volle Migrations-Welle re-scopen (wie ADR-0009 W5). Trifft
  hier nicht zu (Assessment: nicht-strukturell).
- **Rückführung `in-progress` → `open`:** nicht zu erwarten (kein Blocker).

## 5. Closure-Trigger

DoD grün + `make gates` grün + Closure-Notiz; `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Zwei vendored Baselines** = ~124 KB Redundanz + geteilter Pin-Zustand
  (welche gilt, sagt nur `conventions.md` §Baseline + ADR-0011). Bewusst
  akzeptiert; Folge-Trigger beim nächsten Bump: v3.5.0 entfernen, sobald keine
  lebende `done/`-Referenz mehr darauf zeigt.
- **ADR-Status Proposed.** Der Bump ist umgesetzt, aber ADR-0011 wartet auf
  Owner-Accept — analog zum ADR-0010-Fluss (Proposed → Accepted separat).
- **docs-check-Scope.** Falls d-check `.harness/baseline/**` scannt, muss die
  neue v3.5.1-Baseline (interne `../templates/…`-Links + `blob/v3.5.1`-URLs) grün
  durchlaufen — dieselbe Struktur wie die grüne v3.5.0-Baseline.

## 7. Closure-Notiz

Der Bump war klein und ging wie geplant. Der wesentliche Erkenntnisschritt war
das **Delta-Assessment**: `diff -rq` meldete zunächst *alle* Files als verschieden
und der WebFetch-Compare *„3 Files"* — beides irreführend. Der echte Zeilen-Diff
löste den Widerspruch auf: der Bundle-Export stempelt in jeder Datei die
Quell-URLs (`blob/v3.5.0` → `blob/v3.5.1`) und den `Stand:`, substantiell änderte
sich **genau ein File** — die „slice-reserviert"-Korrektur im
Planning-README-Template (`done/` hält neben Slices auch Nicht-Slice-Records:
Welle-Results + aufgelöste Carveouts). Lehre: bei einem vendored Bundle ist
„Datei verschieden" fast bedeutungslos — nur der inhaltliche Zeilen-Diff trennt
den mechanischen Stempel vom echten Delta.

**Was anders war als der letzte Struktur-Umbau (ADR-0009):** kein Layout-Churn.
v3.5.1 ändert keine Pfade; da v3.5.0 zusätzlich liegen bleibt (Owner-Wahl),
brechen keine historischen Links, und der Repath beschränkt sich auf die lebenden
„aktueller-Stand"-Zeiger. Die eine substantielle Korrektur **bestätigt** rückwirkend
MR-005/slice-006 (flach-in-`planning/` + `done/`-Records) — die frühere
„slice-reserviert"-Spannung ([[feedback_kanon_schweigen_keine_luecke]]) ist im
Kanon selbst aufgelöst; kein m-trace-Edit nötig, weil das Repo nie ein eigenes
`planning/README.md` mit der alten Formulierung kopiert hatte.

**Steering-Loop-Eintrag:** ADR-0009 ist immutable und nennt Version + sha —
Konsequenz: Pin-Fortschreibungen brauchen einen **eigenen ADR** (ADR-0011), nicht
einen Edit des alten. Der Re-Evaluierungs-Trigger von ADR-0009 („kein stiller
Auto-Bump") hat genau wie vorgesehen gegriffen. Folge-Trigger im Register/ADR-0011:
beim nächsten Bump entscheiden, ob die dann zwei Versionen alte Baseline entfällt.

**Folge-Slices:** keine neuen `open/`-Einträge. Offener Prozess-Punkt bleibt der
Owner-Accept von ADR-0011.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Harness-Baseline (Vendoring/Doku)

- **Modus:** Brownfield (vendored Bestand v3.5.0; additiver Bump).
- **Konventionen-Dichte:** hoch — ADR-0009-Doktrin (vendored Baseline + sha),
  ADR-0011, `conventions.md` §Baseline.
- **Phase-Reife:** reifer Bestand; nur der Pin bewegt sich.
- **Evidenz-/Diskrepanz-Risiko:** niedrig — Zip-sha + `sha256sum -c` verifiziert
  die Integrität, `docs-check` das Link-Netz, `grep` den vollständigen Repath.
- **Reconciliation-Aufwand:** dieser Slice.
