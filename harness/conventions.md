# Harness-Konventionen

## Zweck

Diese Datei deklariert m-trace-spezifische Strukturregeln und Adaptionen
gegenüber der adoptierten Harness-Baseline. Sie supplementiert die Baseline,
ohne sie zu kopieren.

## Baseline

m-trace adoptiert die ai-harness-course-Baseline; die **aktive Version ist
v6.8.0** (Kurs-Welle 135, 2026-09-13), vendored netzlos unter
`.harness/baseline/v6.8.0/regelwerk/` (17 Module + 8 Grundlagen-Dateien) und
`.harness/baseline/v6.8.0/templates/` (die Referenz-Ziel-Formen). Das
Release-Archiv `lab-regelwerk.zip` trägt sha256
`2c55e6d1b821ae15ff73f5a9b3dc2269843db0ffcf9845a4bd0df2cfebbdc6c7`; die
Per-Datei-Integrität ist in `.harness/baseline/v6.8.0/SHA256SUMS` gepinnt und
wird mit `sha256sum -c` verifiziert. Vendored 2026-09-13
([`slice-012`](../docs/plan/planning/done/slice-012-harness-baseline-v6.8.0-vendoring.md),
[`welle-02`](../docs/plan/planning/welle-02-regelwerk-v6.8.0-migration.md)
Tranche 1) — reines Vendoring + Zeiger-Umstellung; die inhaltliche Anpassung
von `AGENTS.md` und dieser Datei an den neuen Kanon folgt in den weiteren
`welle-02`-Tranchen. Die vorherige **v3.5.1-Baseline bleibt** unter
`.harness/baseline/v3.5.1/` zusätzlich liegen (Audit-Referenzform, analog zur
v3.5.0-Präzedenz unten).

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

Diese Sektion trägt den **Index**, nicht die Einträge — jede Adaption ist
eine eigene Datei unter `harness/conventions/`, kopiert aus
`harness/conventions/MR-NNN-titel.template.md` der vendored Baseline; ist ihr
Auflösungs-Trigger eingetreten, wandert sie per `git mv` nach
`conventions/done/`. Der Zustand ist die Verzeichnis-Position, kein
Status-Feld (Baseline-Regelwerk `grundlagen-harness-dateien.md`
§harness/conventions.md als Konventionsspeicher).

### Aktive Adaptionen

| MR | Titel | Geltungsbereich | Ersetzt-Baseline-Regel |
|---|---|---|---|
| [002](conventions/MR-002-accepted-adr-grandfathering.md) <a id="mr-002"></a> | Accepted-ADR-Grandfathering | `docs/plan/adr/0001-*.md`–`0007-*.md` | `modul-04-adrs.md` §Hard Rule für Accepted-ADRs |
| [003](conventions/MR-003-requirement-id-familien.md) <a id="mr-003"></a> | Requirement-ID-Familien | Contract, Pläne, Commits, Reviews | `grundlagen-source-precedence.md` §ID-Schema als Klammer |
| [006](conventions/MR-006-security-gate-carveout-registry.md) <a id="mr-006"></a> | Security-Gate-Carveout-Registry | `image-scan`/`vuln-check`-Gate | `modul-07-carveouts.md` §Ziel-Form: Carveout |
| [007](conventions/MR-007-planning-artefakt-form.md) <a id="mr-007"></a> | Planning-Artefakt-Form (Slice/Welle vs. `plan-<version>`) | `docs/plan/planning/` | `modul-05-planning-harness.md` §Ziel-Form: Slice |

### Aufgelöste Adaptionen

| MR | aufgelöst durch |
|---|---|
| [001](conventions/done/MR-001-repository-pfade.md) <a id="mr-001"></a> | slice-006 (v3.5.0-Migration W5 — kein Nachfolger-MR, Auflösung durch Slice-Arbeit) |

**`MR-004` und `MR-005` sind keine Adaptionen mehr in diesem Mechanismus**
(Nachzug `welle-02` Tranche 2, 2026-09-13): Das `v6.8.0`-Template verlangt für
jeden Eintrag *„Ersetzt-Baseline-Regel: genau eine Regel der Baseline"* — „ein
Eintrag, der keine benannte Regel ersetzt, ist ein Fork, keine Adaption." Für
beide fand sich **keine** ersetzte Baseline-Regel:

- **`MR-004`** (WSL-Host-Pfad-Beispiele) war nie eine Abweichung von einer
  Regelwerk-Regel, sondern eine Sensor-Konfiguration — jetzt als Kommentar
  direkt bei `hostpaths:` in `.d-check.yml`.
- **`MR-005`** (Nicht-Slice-Register) sagte selbst „keine Kanon-Abweichung,
  der Kanon schweigt" — passt strukturell nicht in einen Mechanismus für
  Baseline-*Abweichungen*. Steht jetzt unten als repo-lokale Strukturregel.

## Repo-lokale Strukturregeln

Regeln, die **keine** Baseline-Vorgabe ersetzen (das Regelwerk schweigt an
dieser Stelle), aber Konsistenz brauchen — kein Fork, keine Adaption, nur
eine Ergänzung, wo die Baseline keine Aussage trifft.

### Nicht-Slice-Register: flache Platzierung in `planning/` <a id="mr-005"></a>

*(vormals `MR-005`, Nummer erhalten für bestehende Verweise — siehe oben)*

- **Datum:** 2026-07-23 (angelegt), 2026-07-23 (zurückgebaut, slice-006)
- **Geltungsbereich:** `docs/plan/planning/risks-backlog.md` (`R-*`-Familie),
  `docs/plan/planning/extra-gates.md` (Quality-Gate-Backlog) samt Companion
  `docs/plan/planning/risks-backlog-werkzeug-triage.md`.
- **Baseline-Bezug:** keine ersetzte Regel — das Regelwerk (Modul 5/6)
  schweigt über Nicht-Slice-Artefakte (verbietet sie nicht).
- **Strukturregel:** m-trace führt stehende Discovery-Register
  (Risiko-Register mit Re-Eval-Triggern, RAK-gekoppelt an die
  Release-Historie; Quality-Gate-Backlog) samt zugehöriger Analysen. Das sind
  Nicht-Slice-Artefakte; sie liegen **flach in `planning/`** — dasselbe
  Muster wie der kanonische Welle-Plan, während die Lifecycle-Verzeichnisse
  (`open/next/in-progress/done`) **slice-reserviert** bleiben.
- **Begründung:** Die flache Ablage füllt keine „Lücke" und sanktioniert
  keine neue Artefaktklasse — sie folgt dem vorhandenen
  Flach-in-`planning/`-Muster. Die frühere Fassung führte die Register in
  `in-progress/` und rechtfertigte das mit „Kanon kennt kein Äquivalent" —
  beides in `slice-006` zurückgebaut (die W4-Triage ordnet
  R-9/R-12/R-28/R-30 als Roadmap-Kandidaten ein, die im Register bleiben;
  Security-Suppressions graduieren in ihr Gate-Werkzeug →
  [MR-006](#mr-006)).
- **Gültig, solange:** die Register geführt werden.

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
