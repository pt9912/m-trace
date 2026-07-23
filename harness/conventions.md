# Harness-Konventionen

## Zweck

Diese Datei deklariert m-trace-spezifische Strukturregeln und Adaptionen
gegenüber der adoptierten Harness-Baseline. Sie supplementiert die Baseline,
ohne sie zu kopieren.

## Baseline

m-trace adoptiert die ai-harness-course-**v3.5.0**-Baseline (2026-07-19),
vendored netzlos unter `.harness/baseline/v3.5.0/regelwerk/` (17 Module +
3 Grundlagen) und `.harness/baseline/v3.5.0/templates/` (die Referenz-Ziel-
Formen). Das Release-Archiv `lab-regelwerk.zip` trägt sha256
`123e3383261102e6be6465e1f4bade08a474c00edc4fff89f5c4b11bd640f8ff`; die
Per-Datei-Integrität ist in `.harness/baseline/v3.5.0/SHA256SUMS` gepinnt und
wird mit `sha256sum -c` verifiziert. Vendored 2026-07-22 gemäß
[ADR-0009](../docs/plan/adr/0009-harness-baseline-v3.5.0.md).

Dies löst den vorherigen Commit-Pin ab (nur `grundlagen-konventionen.md`, bei
ai-harness-course `d2f60da`, abgerufen 2026-07-14): das gesamte Regelwerk und
der vollständige Template-Satz sind jetzt präsent und integritäts-geprüft,
sodass die `../templates/…`-Referenz-Formen lokal auflösen, statt gegen einen
sich bewegenden Kurs-Head zu driften.

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
  [`plan-harness-v3.5.0-migration.md`](../docs/plan/planning/in-progress/plan-harness-v3.5.0-migration.md))
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

### MR-005 - Risiko-Register als Roadmap-Discovery

- **Datum:** 2026-07-23
- **Scope:** `docs/plan/planning/in-progress/risks-backlog.md` (`R-*`-Familie)
- **Baseline-Unterschied:** m-trace führt absehbare technische Risiken mit
  Re-Eval-Triggern (RAK-gekoppelt an die Release-Historie) in einem eigenen
  Risiko-Register. Der v3.5.0-Kanon kennt kein direktes Äquivalent — er trennt
  Roadmap (aufgeschobene Arbeit), Carveout (temporäre Gate-Senkung) und ADR
  (Architekturentscheidung).
- **Grund:** Das Register ist die etablierte, release-gekoppelte
  Roadmap-Discovery-Praxis. Die W4-Werkzeug-Triage
  (`docs/plan/planning/in-progress/risks-backlog-werkzeug-triage.md`) ordnet die
  aktiven Einträge zu: R-9/R-12/R-28/R-30 sind Roadmap-Kandidaten und bleiben
  hier; genuine Gate-Senkungen graduieren in ihr Gate-Werkzeug
  (Security-Suppressions → MR-006).
- **Auflösungs-Trigger:** Permanent, solange die Wellen/Slices-Roadmap-Form (W6)
  die Roadmap-Kandidat-Klasse nicht selbst aufnimmt.

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
| Requirement-Links | Brownfield | Stabile Per-Requirement-Anker existieren und der ID-Link-Gate kann ohne mehrdeutige Root-Links aktiviert werden |
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

Das d-check-`ids`-Modul ist noch nicht aktiviert. Requirements leben derzeit in
Markdown-Tabellenzeilen, sodass eine automatische Reparatur 340 nackte
Kennungen in Links auf den Anfang von `spec/lastenheft.md` verwandeln würde,
nicht auf die einzelne Definition. Das erzeugt link-förmige Mehrdeutigkeit und
wird nicht als Konvergenz akzeptiert.

Die Graduierung erfordert diese Schritte:

1. Jedem normativen Requirement einen stabilen, direkt adressierbaren
   Definitions-Anker geben, ohne seine Kennung oder vertragliche Formulierung zu
   ändern.
2. Verifizieren, dass d-check jede konfigurierte ID-Familie auf diese Definition
   auflöst, nicht bloß auf die enthaltende Datei.
3. Aufwärts-Verweise in aktiver technischer, ADR-, Planning- und Anwender-
   Dokumentation migrieren; historische Release-Records bleiben unverändert, wo
   gefordert.
4. `ids` erst aktivieren, nachdem sein Advisory-Lauf kein mehrdeutiges Ziel
   meldet und die resultierenden Links `links` und `anchors` bestehen.
