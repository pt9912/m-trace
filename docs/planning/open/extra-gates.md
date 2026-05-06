# Extra Quality Gates

> **Status**: Vorschlag / Backlog-Plan. Dieses Dokument ist noch kein
> aktiver Implementierungsplan und fuehrt keine neuen Gates ein. Es
> priorisiert zusaetzliche Quality-Gates, die spaeter in eine konkrete
> Release-Tranche oder ADR ueberfuehrt werden koennen.

## 1. Zweck

m-trace hat bereits harte Gates fuer Tests, Coverage, SOLID-nahe
Lint-Regeln, Architekturgrenzen, Schema-Validierung, Contract-Drift,
SDK-Packaging, SDK-Performance und mehrere Smoke-Pfade. Diese Notiz
beschreibt die naechste sinnvolle Gate-Schicht:

- Security-Risiken frueher sichtbar machen.
- Performance-Regressionen messbar machen, ohne PRs durch Runner-
  Rauschen unzuverlaessig zu blockieren.
- Generierte Artefakte synchron halten.
- Testqualitaet an kritischen Stellen ueber Coverage hinaus pruefen.

Nicht-Ziel: jedes Gate sofort in `make gates` aufnehmen. Einige
Pruefungen sind bewusst Nightly- oder Release-Kandidaten.

## 2. Gate-Matrix

| Prio | Gate | Zielpfad | Blockierend | Hauptnutzen |
| ---: | ---- | -------- | ----------- | ----------- |
| 1 | `govulncheck` + Container-Scan | PR + Release | ja | bekannte CVEs frueh stoppen |
| 2 | Benchmark-Smoke mit festen Budgets | PR | ja | grobe Performance-Regressionen stoppen |
| 3 | Nightly `benchstat`-Regressionen | Nightly + Release | ja fuer Release | statistisch belastbare Trends erkennen |
| 4 | Generated-Artifact-Drift-Gate | PR | ja | Quellen und generierte Dateien synchron halten |
| 5 | Selektives Fuzzing / Property Tests | PR kurz, Nightly lang | ja fuer Seed-Corpus | Parser-/Validation-Robustheit erhoehen |
| 6 | Mutation Testing auf kritischen Modulen | Nightly / manuell | nein initial | Teststaerke sichtbar machen |

## 3. Priorisierung

### 3.1 `govulncheck` + Container-Scan

**Entscheidung:** zuerst einfuehren. Der Nutzen ist hoch, die
Interpretation klar und das Gate passt gut zum bestehenden
Docker-/CI-Modell.

Scope:

- Go-Abhaengigkeiten unter `apps/api` mit `govulncheck ./...`.
- Runtime-Images fuer API, Dashboard und Analyzer-Service mit einem
  Image-Scanner wie Trivy oder Grype.
- Scan-Policy fuer Container: `CRITICAL` und `HIGH` blockieren;
  `MEDIUM` wird berichtet, aber nicht sofort blockierend.

Erste Ziel-Targets:

- `make vuln-check`
- `make image-scan`
- spaeter optional: `make security-gates`

DoD:

- Tool-Versionen sind gepinnt oder reproduzierbar bezogen.
- CI gibt maschinenlesbare Scan-Artefakte aus.
- False-Positive-/Ignore-Regeln liegen versioniert mit Begruendung vor.
- `make image-scan` baut die zu scannenden Runtime-Images im selben
  Lauf oder konsumiert eindeutig benannte Image-Tags aus einem
  vorangegangenen CI-Step. Dashboard und Analyzer-Service duerfen
  nicht implizit als vorhanden angenommen werden, weil `make build`
  aktuell nur API-Image plus TypeScript-Artefakte baut.
- Wenn Security-Gates PR-blockierend werden, ist der GitHub-Actions-
  Workflow explizit erweitert oder auf ein zentrales Target umgestellt,
  das diese Gates enthaelt. Ein neues Make-Target allein reicht nicht.

### 3.2 Benchmark-Smoke fuer API und Stream-Analyzer

**Entscheidung:** als Budget-Smoke in den PR-Pfad aufnehmen, aber
nicht als Mikrobenchmark-Baselinevergleich. PRs sollen nur bei klaren
Budget-Verletzungen scheitern.

API-Kandidaten:

- Event-Ingest: `RegisterPlaybackEventBatch` mit typischer Batch-
  Groesse und maximal erlaubter Batch-Groesse.
- SQLite-Ingest: persistierter Event-Batch inklusive Sequenzvergabe.
- Session-Listing/Pagination: typische Seiten und grosse Session-
  Menge.
- Cursor: Encode/Decode und Fehlerpfade.

Stream-Analyzer-Kandidaten:

- HLS Master-Manifest klein/gross.
- HLS Media-Manifest mit vielen Segmenten.
- DASH-MPD, sobald DASH-Analyse produktiv ist.
- SSRF-/URL-Klassifizierung mit typischen erlaubten und geblockten
  Eingaben.

Policy:

- Budgets sind absolute Obergrenzen, nicht Vergleich gegen letzten
  Commit.
- Budgets starten grosszuegig und werden erst nach Messhistorie
  geschaerft.
- Budget-Smokes messen Hot Paths mit synthetischen, repo-lokalen
  Fixtures; keine Netzwerkabhaengigkeit.
- PR-Budgets gelten zunaechst fuer GitHub Actions auf `ubuntu-24.04`.
  Jeder Lauf druckt mindestens Runner-OS, CPU-Modell und relevante
  Runtime-Versionen, damit Budget-Failures einordenbar bleiben.
- Neue oder geaenderte Budgets laufen erst fuer mehrere gruene CI-
  Beobachtungslaeufe nicht-blockierend mit, bevor sie PRs blockieren.

Erste Ziel-Targets:

- `make api-benchmark-smoke`
- `make analyzer-benchmark-smoke`
- `make benchmark-smoke`

DoD:

- Jeder Smoke druckt Laufzeit, Allokationen oder Durchsatz lesbar aus.
- Budgetverletzung erzeugt eine eindeutige Fehlermeldung mit Ist/Soll.
- Die Fixtures sind stabil und versioniert.
- Wenn `make benchmark-smoke` PR-blockierend wird, ist der
  GitHub-Actions-Workflow explizit erweitert oder auf ein zentrales
  Target umgestellt, das den Smoke enthaelt.

### 3.3 Nightly `benchstat`-Regressionen

**Entscheidung:** nicht in den PR-Critical-Path. Benchmarks mit
Baselinevergleich brauchen Wiederholungen und stabile Auswertung.

Scope:

- Go-Benchmarks mit `go test -bench=. -benchmem -count=N`.
- Auswertung mit `benchstat` gegen eine gespeicherte Baseline.
- Initial nur API-Benchmarks; TypeScript-Benchmarks erst nach
  stabilen Budget-Smokes.

Regression Policy:

- Nightly markiert auffaellige Regressionen als CI-Failure.
- Release-Gate blockiert bei bestaetigter Regression.
- Einzelne laute Benchmarks duerfen quarantined werden, muessen aber
  im Plan vermerkt bleiben.

DoD:

- Baseline-Herkunft ist dokumentiert.
- `benchstat`-Output wird als CI-Artefakt gespeichert.
- Schwellen fuer Blocker sind explizit, z. B. >15 % Regression bei
  stabil signifikantem Ergebnis.

### 3.4 Generated-Artifact-Drift-Gate

**Entscheidung:** frueh einfuehren. Dieses Gate ist deterministisch und
passt gut zu PRs.

Kandidaten:

- Schema-DDL aus `apps/api/internal/storage/schema.yaml`.
- Contract-Fixtures aus `spec/contract-fixtures`.
- Public-API-Snapshots der TypeScript-Pakete.
- `contracts/sdk-compat.json` und versionsfuehrende Fixtures.

Bevorzugtes Muster:

1. Generierungs-/Sync-Targets ausfuehren.
2. `git diff --exit-code` auf die erwarteten Pfade.
3. Bei Diff mit klarer Meldung abbrechen.

Erste Ziel-Targets:

- `make generated-drift-check`
- optional paketweise Untertargets fuer Schema, Fixtures und
  Public-API.

DoD:

- Das Gate laeuft ohne Netzwerk.
- Es prueft nur deterministische Artefakte.
- Der Fehlertext nennt den konkreten Regenerierungsbefehl.
- Wenn `make generated-drift-check` PR-blockierend wird, ist der
  GitHub-Actions-Workflow explizit erweitert oder auf ein zentrales
  Target umgestellt, das den Check enthaelt.

### 3.5 Selektives Fuzzing / Property Tests

**Entscheidung:** selektiv, nicht repo-weit. Ziel sind Parser,
Decoder und Validierungslogik, bei denen Edge-Cases realistisch sind.

Go-Kandidaten:

- Cursor Encode/Decode.
- HTTP-Validation von Playback-Event-Batches.
- Event-Metadaten und Redaction.
- SRT-Health-Mapping.

TypeScript-Kandidaten:

- HLS Attribute Parser.
- HLS Master-/Media-Parser.
- SSRF-/URL-Klassifizierung.
- Result-Stability-Invarianten.

Policy:

- PR-Gate: kurzer Seed-Corpus-Lauf oder Property Tests mit begrenztem
  Case-Budget.
- Nightly: laengerer Fuzz-Lauf mit Artefakt fuer gefundene Crasher.
- Crasher werden als Regression-Fixtures eingecheckt.

DoD:

- Fuzz-Ziele sind klein und deterministisch.
- Seeds liegen im Repo.
- Gefundene Minimalfaelle werden in normale Unit-Tests ueberfuehrt.

### 3.6 Mutation Testing auf kritischen Modulen

**Entscheidung:** spaeter und nur punktuell. Mutation Testing ist
nuetzlich, aber teuer und anfangs oft noisy.

Kandidaten:

- Event-Validierung.
- Cursor-Logik.
- HLS-Parser; DASH-Parser erst, sobald DASH-Analyse produktiv ist.
- SRT-Health-Mapping.
- Security-relevante URL-/SSRF-Pruefung.

Policy:

- Kein repo-weites Gate.
- Start als Nightly- oder manueller Report.
- Erst blockierend machen, wenn Mutant-Survival-Rate stabil und
  False-Positive-Rate niedrig ist.

DoD:

- Kritische Module sind explizit allowlisted.
- Ergebnis wird als Trend verfolgt, nicht nur als Einzelwert.
- Threshold-Senkungen oder Scope-Aenderungen sind begruendungspflichtig.

## 4. Benchmarking-Policy

Benchmarking bleibt zweigeteilt:

- **PR-Gate:** robuste Budget-Smokes mit festen, grosszuegigen
  Schwellwerten.
- **Nightly/Release:** statistische Regressionserkennung mit mehreren
  Wiederholungen und Baseline-Vergleich.

Ein PR sollte nicht daran scheitern, dass ein Mikrobenchmark auf einem
Shared Runner leicht schwankt. Blockierend im PR-Pfad sind nur klare
Budget-Verletzungen. Statistische Trends gehoeren in Nightly- und
Release-Gates.

## 5. Vorgeschlagene Einfuehrungsreihenfolge

1. Security-Gates (`govulncheck`, Container-Scan) als eigene Targets
   einfuehren und danach in CI verdrahten.
2. Generated-Artifact-Drift-Gate einfuehren, weil es deterministisch
   und schnell ist.
3. Benchmark-Smoke fuer API und Stream-Analyzer mit konservativen
   Budgets ergaenzen.
4. Nightly-Benchmark-Workflow mit `benchstat` und Baseline-Artefakt
   aufsetzen.
5. Property-/Fuzz-Ziele fuer Cursor, Parser und URL-Klassifizierung
   aufnehmen.
6. Mutation Testing als nicht-blockierenden Nightly-Report fuer wenige
   kritische Module starten.

## 6. Offene Entscheidungen

- Welcher Container-Scanner wird Standard: Trivy oder Grype?
- Werden Security-Gates direkt Teil von `make gates` oder zuerst ein
  separates `make security-gates`?
- Wo lebt die Benchmark-Baseline: Git-Repo, GitHub Actions Artefakt
  oder Release-Asset?
- Welche Performance-Budgets gelten initial fuer API und
  Stream-Analyzer?
- Ab welcher Regression blockiert `benchstat` ein Release?
