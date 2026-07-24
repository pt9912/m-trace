# Harness-Konventionen

## Zweck

Diese Datei deklariert m-trace-spezifische Strukturregeln und Adaptionen
gegenüber der adoptierten Harness-Baseline. Sie supplementiert die Baseline,
ohne sie zu kopieren.

## Baseline

m-trace adoptiert die ai-harness-course-Baseline; die **aktive Version ist
v3.5.1** (Kurs-Welle 33, 2026-07-23), vendored netzlos unter
`.harness/baseline/v3.5.1/regelwerk/` (17 Module + 3 Grundlagen) und
`.harness/baseline/v3.5.1/templates/` (die Referenz-Ziel-Formen). Das
Release-Archiv `lab-regelwerk.zip` trägt sha256
`7268a8e6f36476c98d5cf0547d16deacec70fcddcf23df38f87d029e967cb10d`; die
Per-Datei-Integrität ist in `.harness/baseline/v3.5.1/SHA256SUMS` gepinnt und
wird mit `sha256sum -c` verifiziert. Vendored 2026-07-24 gemäß
[ADR-0011](../docs/plan/adr/0011-harness-baseline-v3.5.1-bump.md) (nicht-
struktureller Bump: Kurs-URL-Stempel + eine Template-Klarstellung, dass `done/`
neben Slices auch Nicht-Slice-Records hält — bestätigt MR-005/slice-006).

Die **strukturelle Adoption** (kanonisches Layout, vendored-Baseline-Mechanismus,
AGENTS.md-Einstieg) traf ADR-0009 mit v3.5.0; ADR-0011 schreibt nur den Pin fort.
Die vendored **v3.5.0-Baseline bleibt** unter `.harness/baseline/v3.5.0/`
zusätzlich liegen (Owner-Entscheidung 2026-07-24), damit die
`.harness/baseline/v3.5.0/…`-Verweise der historischen `done/`-Records und der
immutablen ADRs netzlos auflösbar bleiben — sie ist Audit-Referenzform, nicht
der aktive Stand. Der Baseline-Mechanismus löste seinerzeit den früheren
Commit-Pin ab (nur `grundlagen-konventionen.md`, bei ai-harness-course
`d2f60da`, abgerufen 2026-07-14): seither sind das gesamte Regelwerk und der
vollständige Template-Satz präsent und integritäts-geprüft, sodass die
`../templates/…`-Referenz-Formen lokal auflösen, statt gegen einen sich
bewegenden Kurs-Head zu driften.

## Spec-Straten

| Stratum | Dateien | Rolle |
|---|---|---|
| Contract | `spec/lastenheft.md` | Bindende Anforderungen und Akzeptanzkriterien |
| Technical | `spec/backend-api-contract.md`, `spec/browser-support.md`, `spec/player-sdk.md`, `spec/telemetry-model.md` | Bindendes technisches Detail, das den Contract verfeinert |
| View | `spec/architecture.md` | Abgeleitete Komponenten-, Abhängigkeits- und Datenfluss-Sicht |

Die normative Referenz-Stabilität ist `Contract > Technical > View > ADR >
Planning`. Verweise zeigen aufwärts. Spec-Dokumente nutzen keine ADR- oder
Planning-Artefakte als normative Quelle. Abwärts-Provenienz ist nur in
ausgewiesenen History-Abschnitten erlaubt.

## Adaptionen

### MR-001 - Repository-Pfade — AUFGELÖST 2026-07-23

- **Datum:** 2026-07-14 (angelegt), 2026-07-23 (aufgelöst)
- **Scope:** ADR- und Planning-Verzeichnisse
- **Baseline-Unterschied (historisch):** m-trace nutzte `docs/adr/` statt
  `docs/plan/adr/` und `docs/planning/` statt `docs/plan/planning/`.
- **Grund (historisch):** Etabliertes öffentliches Repository-Layout mit
  umfangreichen stabilen Links.
- **Auflösung:** Die v3.5.0-Migration W5 (Layout-Move,
  [`plan-harness-v3.5.0-migration.md`](../docs/plan/planning/done/plan-harness-v3.5.0-migration.md))
  hat das Repo auf das Kanon-Layout gehoben: `docs/plan/adr/`,
  `docs/plan/planning/`, `docs/plan/carveouts/`. Die immutablen Accepted-ADRs
  blieben unangetastet — ihre Pre-Move-Verweise sind per `ignore-refs`-Tombstone
  in `.d-check.yml` grandfathered (Frozen-Doc-Refactoring). Damit ist die
  Pfad-Divergenz beseitigt; diese Adaption ist geschlossen.

### MR-002 - Accepted-ADR-Grandfathering

- **Datum:** 2026-07-14
- **Scope:** `docs/plan/adr/0001-*.md` bis `docs/plan/adr/0007-*.md`
- **Baseline-Unterschied:** Diese akzeptierten Brownfield-Records enthalten
  historische Plan-Provenienz außerhalb eines ausgewiesenen History-Abschnitts.
- **Grund:** Akzeptierte ADRs sind unter der adoptierten Baseline immutable und
  werden nicht allein zum Nachrüsten der Konvention umgeschrieben.
- **Auflösungs-Trigger:** Permanente historische Ausnahme. Neue ADRs erhalten
  keine Ausnahme; künftige Accepted-ADR-Änderungen werden vom
  ADR-Immutabilitäts-Sensor blockiert.

### MR-003 - Requirement-ID-Familien

- **Datum:** 2026-07-14
- **Scope:** Contract, Pläne, Commits und Reviews
- **Baseline-Unterschied:** m-trace datiert vor der `LH-*`-Beispielfamilie und
  nutzt `F-*`, `NF-*`, `MVP-*`, `AK-*`, `RAK-*` und `R-*`.
- **Grund:** Die Kennungen sind Teil des etablierten Contracts und der
  Release-Historie.
- **Auflösungs-Trigger:** Permanent. Neue Requirement-Familien müssen hier vor
  Nutzung deklariert werden.

### MR-004 - WSL-Host-Pfad-Beispiele

- **Datum:** 2026-07-14
- **Scope:** Drei WSL-Troubleshooting-Beispiele in
  `docs/user/local-development.md`
- **Baseline-Unterschied:** Der `hostpaths`-Sensor lehnt host-lokale absolute
  Pfade normalerweise ab.
- **Grund:** Diese Pfade sind der Gegenstand der Operator-Guidance und können
  nicht durch repository-relative Pfade ersetzt werden, ohne die Diagnose zu
  verlieren.
- **Auflösungs-Trigger:** Permanent, solange WSL2 unterstützt wird. Der
  `hostpaths`-Sensor gated daher `/Users` und `/Development`; `/mnt` und
  `/home` liegen bewusst außerhalb seines konfigurierten Präfix-Satzes.

### MR-005 - Nicht-Slice-Register: flache Platzierung in `planning/`

- **Datum:** 2026-07-23 (angelegt), 2026-07-23 (zurückgebaut, slice-006)
- **Scope:** `docs/plan/planning/risks-backlog.md` (`R-*`-Familie),
  `docs/plan/planning/extra-gates.md` (Quality-Gate-Backlog) samt Companion
  `docs/plan/planning/risks-backlog-werkzeug-triage.md`.
- **Repo-lokale Strukturregel:** m-trace führt stehende Discovery-Register
  (Risiko-Register mit Re-Eval-Triggern, RAK-gekoppelt an die Release-Historie;
  Quality-Gate-Backlog) samt zugehöriger Analysen. Das sind Nicht-Slice-
  Artefakte; sie liegen **flach in `planning/`** — dasselbe Muster wie der
  kanonische Welle-Plan, während die Lifecycle-Verzeichnisse
  (`open/next/in-progress/done`) **slice-reserviert** bleiben.
- **Keine Kanon-Abweichung:** Der v3.5.0-Kanon *schweigt* über
  Nicht-Slice-Artefakte (er verbietet sie nicht); die flache Ablage füllt keine
  „Lücke" und sanktioniert keine neue Artefaktklasse — sie folgt dem vorhandenen
  Flach-in-`planning/`-Muster. Die frühere Fassung führte die Register in
  `in-progress/` und rechtfertigte das mit „Kanon kennt kein Äquivalent" — beides
  in slice-006 zurückgebaut (die W4-Triage ordnet R-9/R-12/R-28/R-30 als
  Roadmap-Kandidaten ein, die im Register bleiben; Security-Suppressions
  graduieren in ihr Gate-Werkzeug → MR-006).
- **Auflösungs-Trigger:** Permanent, solange die Register geführt werden.

### MR-006 - Security-Gate-Carveout-Registry

- **Datum:** 2026-07-23
- **Scope:** `image-scan`/`vuln-check`-Gate; OS-CVE-Ausnahmen der
  `node:22-trixie-slim`-Base (`mtrace-dashboard`, `mtrace-analyzer-service`),
  geführt in `.security/vulnignore.yaml`
- **Baseline-Unterschied:** Der Kanon (Modul 7) führt einzelne, temporäre
  Gate-Senkungen als `docs/plan/carveouts/CO-NNN`. m-trace senkt den Security-Gate
  für einen **Cluster** transitiver OS-CVEs (kein Runtime-Pfad, oft ohne
  Upstream-Fix) über die domänenspezifische Registry `.security/vulnignore.yaml`
  (per-CVE `reason` + `expires` + `scope`, deterministisch nach `.trivyignore`
  gerendert, Nightly-Audit-Re-Eval).
- **Grund:** Modul-7-Werkzeug-Wahl Frage 1 (Granularität): ein Cluster im selben
  Geltungsbereich ist eine BF-Sub-Area-Markierung, **keine** Carveout-Kaskade
  (ein CO-File je CVE ist der explizit gewarnte Anti-Pattern). Die vorhandene
  Registry ist reicher als das generische CO-Template und die Single Source of
  Truth; ein CO-File je CVE würde sie duplizieren.
- **Auflösungs-Trigger:** Je CVE der eigene `expires`/Upstream-Fix-Trigger in
  `vulnignore.yaml`; als Sub-Area permanent, solange die `trixie-slim`-Base
  transitive OS-CVEs ohne Runtime-Exponierung trägt. Das generische
  `docs/plan/carveouts/` bleibt für künftige einzelne, nicht-Security
  Gate-Senkungen reserviert.

### MR-007 - Planning-Artefakt-Form (Slice/Welle vs. plan-&lt;version&gt;)

- **Datum:** 2026-07-23
- **Scope:** Planning-Artefakte unter `docs/plan/planning/`
- **Baseline-Unterschied:** Der v3.5.0-Kanon (Modul 5/6) führt Arbeit als
  **Slices** (`open/…/done/slice-<NNN>-<titel>.md`, Zustand = Verzeichnis) und
  **Wellen** (`welle-<NN>.md` flach → `done/` neben `welle-<NN>-results.md`).
  m-traces Bestand nutzt release-gekoppelte `plan-<version>.md`-Dateien in `done/`.
- **Entscheidung (Owner 2026-07-21):** **Neue** Arbeit folgt der kanonischen
  Slice/Welle-Form (aus den vendored Templates
  `.harness/baseline/v3.5.1/templates/docs/plan/planning/{slice,welle}.template.md`).
  Der **Bestand `plan-<version>.md` wird grandfathered** (Variante A): historische
  Release-Records bleiben unverändert, keine Massen-Umbenennung, die
  Release-Versions-Kopplung bleibt für die Alt-Form. Brownfield-konsistent
  (analog MR-002-Grandfathering).
- **Auflösungs-Trigger:** Permanent für den Bestand. **Folge-Punkt (bei erstem
  Slice/Welle-Artefakt):** `trace.slices.file-pattern` (`.d-check.yml`, heute
  `^plan-(.+)\.md$`) und der Closure-Note-Glob (`check_closure_notes.py --glob`,
  heute `plan-*.md`) werden dann additiv um die `slice-*`/`welle-*-results`-Form
  erweitert — solange keine solche Datei existiert, ist keine Config-Änderung
  nötig (netzlos).

## Sensor-Bindungsklassen

m-trace nutzt derzeit Requirement-Bindung (`F-*`, `NF-*`, `MVP-*`, `AK-*`,
`RAK-*`, `R-*`), ADR-Bindung (`ADR-NNNN`) und Reproduzierbarkeits-Bindung über
immutable Image-Digests.

## Modi

| Sub-Area | Modus | Graduierungs-Bedingung |
|---|---|---|
| Spec-Referenz-Richtung | Greenfield | Durchgesetzt von `make docs-check`; keine offenen Reconciliation-Befunde |
| Bestehende akzeptierte ADRs | Brownfield, grandfathered | Historische Dateien bleiben immutable; jede neue ADR folgt der Baseline |
| Commit-Traceability | Greenfield für neue Pull Requests | PR-Bereiche bestehen `make docs-commits`; Vor-Adoptions-Historie bleibt unverändert |
| Requirement-Coverage | Brownfield, observable | Jedes geforderte Requirement hat einen Slice oder kuratierten Coverage-Verweis und `make doc-complete` besteht |
| Requirement-Links | Greenfield | `ids` repo-weit über alle aktiven Doc-Dirs aktiv (welle-01: slice-001 Spec-Straten, slice-002 Rest + R-Familie); verankerte Links gg. inline-`<a id>`-Anker. Durchgesetzt in `make gates`. Exempt: immutable ADRs, `done/`, Root-Übersicht; R-Familie in `spec/**` (matrix-Richtung) |
| Security-Gate-Suppressions (`image-scan`) | Brownfield, observable | Jede Suppression trägt Begründung + `expires` + Scope in `.security/vulnignore.yaml`; Nightly-Audit re-evaluiert; aufgelöst, sobald die `trixie-slim`-Base keine transitiven OS-CVEs ohne Runtime-Pfad mehr trägt (MR-006) |

## Requirement-Coverage-Konvergenz

d-check v0.43.0 liest die bestehenden `Kennung`/`Prioritaet`/`Anforderung`- und
`Akzeptanzkriterium`-Tabellen nativ. `make doc-trace` ist der Advisory-Sensor.
Die historische `RAK-51`-Redefinition nutzt die explizite
`duplicate-ids: last`-Policy, weil die spätere Zeile ihre Modalität von Kann auf
Muss anhebt.

`make doc-complete` ist nicht in CI gebunden, solange geforderte Requirements
ohne Slice- oder kuratierten Coverage-Verweis verbleiben. Die Graduierung
erfordert das Triagieren dieser Einträge, das Hinzufügen wahrheitsgemäßer
Aufwärts-Coverage-Verweise oder expliziter kuratierter Coverage und das
Erreichen eines bestehenden Gates ohne Schwächung der Modalitäts-Policy.

## Requirement-Link-Konvergenz

Das d-check-`ids`-Modul ist **repo-weit aktiv** (welle-01, `make gates`) über alle
aktiven Doc-Dirs. Requirements leben in Markdown-Tabellenzeilen; damit d-check auf
die *einzelne Definition* auflöst statt auf den Datei-Anfang (link-förmige
Mehrdeutigkeit, die m-trace nicht als Konvergenz akzeptiert), trägt jede
Definitionszeile einen inline-Anker `<a id="<kennung-klein>"></a>` **in der
letzten Zelle** (die Kennungs-Zelle bleibt rein — sonst passt sie nicht
„vollständig" aufs RTM-`id-pattern` und `--trace`/`doc-complete` erkennt 0
Anforderungen), und Mentions sind verankerte Links `[ID](…#slug)`.

**Klarstellung (slice-001):** d-check selbst verlangt den Anker *nicht* — sein
`--repair` erzeugt datei-level Links; der Anker-Zwang ist m-traces Qualitäts-
Policy (Verweis muss auf die Definition zeigen). Der Weg zu den Ankern folgt der
Handbuch-„Brownfield-Migration für tabellarische Lastenhefte" (inline-Anker statt
Heading-Umbau, Tabellenform bleibt).

**Stand der Graduierung — abgeschlossen (welle-01):**

1. ✅ `slice-001` — Spec-Straten (`spec/**`, F/NF/MVP/AK/RAK), 372 Lastenheft-
   Anker, 213 Mentions verankert.
2. ✅ `slice-002` — Rest der aktiven Doku (docs/user, examples, docs/perf,
   Planning-`in-progress/`, carveouts) + `R-`-Familie (31 Anker in
   `risks-backlog.md`). `\b`-Wortgrenze verhindert Über-Matches (`R-` in
   `MR-`/`ADR-`).
3. **Dauerhaft außerhalb** (`ids`-Scope/`exempt-paths`): immutable Accepted-ADRs
   (`docs/plan/adr/**`, MR-002), `done/`, `CHANGELOG`, Root-Übersichts-Docs; die
   `R-`-Familie zusätzlich in `spec/**` (der Vertrag verweist nicht abwärts aufs
   Risiko-Register — `matrix`-Richtung).
