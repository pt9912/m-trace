# Slice 025: Regelwerk-Baseline v6.8.0 → v6.13.0

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle — Closure-Bedingung
ist die DoD dieses Slices, keine Wellen-Grenze.

**Bezug:** `.harness/baseline/`, `harness/conventions.md` §Baseline,
`AGENTS.md` §1/§3.6, MR-002–MR-009 (nur Link-Versionen),
`.harness/skills/reviewer.md`, `.claude/agents/architect.md`,
`docs/reviews/README.md`, `docs/plan/carveouts/README.md`,
`docs/plan/planning/observations/README.md`, `.d-check.yml`
(ignore-refs-Tombstone für MR-001).

**Autor:** Owner-Antrag „auf das neueste Regelwerk umstellen". **Datum:** 2026-09-29.

---

## 1. Ziel und Abgrenzung

**Ziel:** Die vendierte Baseline von `v6.8.0` (Kurs-Welle 135) auf `v6.13.0`
(Kurs-Welle 153) heben: Vendoring mit Provenienz-Verifikation,
Baseline-Abschnitt und aktive Referenzen umstellen, Alt-Stand entfernen.

**Ausdrücklich NICHT in diesem Slice:**

- **Neue Baseline-Mechanik übernehmen** (modul-05 „Ein Slice, dessen Gegenstand
  ein anderer übernimmt", Hard-Rule-Trigger-Audit modul-06, RB-Reihe modul-03,
  file.max-lines/`harness/rules/`-Split) — Append-only (modul-02): bestehende
  Form-Instanzen werden nicht rückwirkend umgeschrieben; die Mechanik gilt für
  neue Instanzen ab diesem Stand. Ausnahme: die AGENTS.md-§3.6-Ergänzung
  (Carveout-Nuance), weil sie die dort zitierte Hard Rule direkt präzisiert.
- **RB-Familie einführen** — das Lastenheft trägt keine Randbedingungen;
  MR-003 ersetzt das ID-Schema ohnehin als Ganzes (Kanon-Schweigen: kein
  MR nötig).
- **Historische Artefakte anfassen** (`done/`-Pläne, Review-Reports,
  Roadmap-Closure-Log) — bleiben in ihrer damaligen Form.

## 2. Definition of Done

- [x] `.harness/baseline/v6.13.0/{regelwerk,templates}/` + `SHA256SUMS`
      committet; Provenienz verifiziert (Zip-Digest `b5151e…` gegen das
      Release-SHA256SUMS-Asset, Extraktion diff-identisch zum liegenden
      Bestand, `sha256sum -c` → 54× OK).
- [x] `harness/conventions.md` §Baseline auf `v6.13.0` (Stand, Datum, URLs,
      Pfad, Digest, Kurs-Welle 153) und ohne den „Frühere Stände bleiben"-
      Satz (Owner-Entscheid, Präzedenz `b254801`).
- [x] `.harness/baseline/v6.8.0/` entfernt.
- [x] Aktive Referenzen auf `v6.13.0` gehoben: `AGENTS.md` (Pfade + §3.6),
      MR-002–MR-009 (Link-Versionen, Einträge unverändert),
      `.harness/skills/reviewer.md`, `.claude/agents/architect.md`,
      `docs/reviews/README.md`, `docs/plan/carveouts/README.md`,
      `docs/plan/planning/observations/README.md`.
- [x] `ignore-refs`-Tombstone für `MR-001` (done/, immutable) gegen die
      entfernten v6.8.0-Ziele.
- [x] `make docs-check` grün (156 Dateien, 0 Befunde).
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.harness/baseline/v6.13.0/**` | neu | aus Release-Asset entpackt (liegt vor, Provenienz geprüft) |
| `.harness/baseline/v6.8.0/**` | gelöscht | Owner-Entscheid; Präzedenz `b254801` |
| `harness/conventions.md` | update | §Baseline inkl. „Frühere Stände"-Satz |
| `AGENTS.md` | update | Pfade §1; §3.6 Carveout-Nuance (modul-09-Ziel-Form v6.13.0) |
| `harness/conventions/MR-002..009` | update | Link-Versionen; Append-only für Einträge |
| `.harness/skills/reviewer.md`, `.claude/agents/architect.md`, `docs/reviews/README.md`, `docs/plan/carveouts/README.md`, `docs/plan/planning/observations/README.md` | update | v6.8.0-Pins → v6.13.0 |
| `.d-check.yml` | update | ignore-refs-Tombstone MR-001 → `.harness/baseline/v6.8.0/**` |

## 4. Trigger

- **Start (`next` → `in-progress`):** sofort — Owner-Antrag, keine Vorgänger.
- **Rückführungen:** keine erwartet (Umfang begrenzt, `make docs-check` als
  Sensor).

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **Tote Links nach dem Alt-Entfernen** (historische Doku zitiert oder
  linkt v6.8.0-Pfade) — **Ausgang:** eingetreten (MR-001 linkt, alle übrigen
  `done/`-Nennungen zitieren nur in Backticks) → ignore-refs-Tombstone,
  Scope `harness/conventions/done/MR-001-*`.
- **Anchor-Drift der MR-Links** — **Ausgang:** entfallen (Ziel-Headings gegen
  v6.13.0 geprüft: alle bestehenden Anker stehen; einziges neues Heading
  „Nachzug ist keine Überschreibung" ist additiv).

## 7. Closure-Notiz

Upgrade `v6.8.0` (Kurs-Welle 135) → `v6.13.0` (Kurs-Welle 153) vollzogen:
Vendoring mit vollständiger Provenienz-Kette (Release-SHA256SUMS-Asset
trägt den Zip-Digest `b5151e…`, Extraktion diff-identisch zum liegenden
Bestand, `sha256sum -c` → 54× OK), §Baseline umgestellt, Alt-Stand per
Owner-Entscheid entfernt (Präzedenz `b254801`), aktive Referenzen gehoben.
Inhalte der MR-Einträge unverändert (Append-only, modul-02).

**Was hat funktioniert:** Headings-Abgleich zwischen beiden Baseline-Ständen
vor dem MR-Link-Bump — alle Ziel-Anker bestanden, kein Anchor-Drift. Der
liegende (schon entpackte) Bestand machte den Vendor-Commit zum
Verifikations-Diff statt zum Download-Risiko.

**Was ging anders als geplant:** Zwei docs-check-Befunde nach den
mechanischen Edits: (1) Der pauschale `s/v6.8.0/v6.13.0/g`-Sed griff auch
den welle-02-**Dateinamen** in MR-008 (Artefakt-Adresse, kein Pin) —
zurückgeschrieben. (2) planning-drift: Der Ruhe-Marker „Keine aktive
Welle" widerspricht jedem Slice in `in-progress/`, auch dem wellenlosen —
der Marker muss weichen, nicht ergänzt werden.

**Steering-Loop-Eintrag:** Guide geschärft: Versions-Bumps per Sed dürfen
nur auf Pins zielen — Dateinamen, die eine Versionsnummer tragen
(`welle-02-regelwerk-v6.8.0-migration.md`), ausnehmen und repo-weit
gegenzählen. (Gezählt, nicht verkörpert — kein Sensor nötig, `make
docs-check` fing beide Fälle.)

**Beobachtungs-Register (`../observations/`):** keine Beobachtung angefallen.

**Folge-Slices:** keine.

**Risiken aus §6:** beide mit Ausgang (Tote Links → Tombstone;
Anchor-Drift → entfallen, siehe §6).

**Nicht übernommen aus v6.13.0 (Append-only, gilt ab diesem Stand):**
modul-05 „Ein Slice, dessen Gegenstand ein anderer übernimmt", Hard-Rule-
Trigger-Audit (modul-06), RB-Reihe (modul-03), file.max-lines/
`harness/rules/`-Split. Einzige Zielform-Anpassung der verkörperten Form:
AGENTS.md §3.6 Carveout-Nuance (modul-09).

## 8. Sub-Area-Prüfungen und Modus-Begründung

Reine Harness-/Doku-Nachführung, kein Produktcode. Berührte Sub-Area:
Baseline-Vendor + Konventions-Dateien (Brownfield, grandfathered Formen
bleiben; jede Änderung folgt der v6.13.0-Ziel-Form). Kein ADR — kein Gate
gesenkt oder geschaffen; die §3.6-Ergänzung präzisiert die bestehende
Schwellen-Regel, sie senkt nichts.
