# Extra Quality Gates

## Zweck

Diese Notiz sammelt zusaetzliche Quality-Gates, die ueber die bereits
etablierten Coverage-, SOLID-/Lint-, Architektur-, Schema-, Smoke- und
SDK-Performance-Gates hinausgehen. Die Reihenfolge ist bewusst
priorisiert: zuerst Gates mit hohem Sicherheits- oder Regressionsnutzen
bei moderatem Aufwand, danach teurere Analysepfade.

## Priorisierung

1. **`govulncheck` + Container-Scan**
   - Go-Abhaengigkeiten mit `govulncheck` pruefen.
   - Runtime-Images fuer API, Dashboard und Analyzer-Service mit einem
     Image-Scanner wie Trivy oder Grype pruefen.
   - Ziel: bekannte CVEs frueh erkennen, bevor Release-Artefakte
     entstehen.

2. **Benchmark-Smoke fuer API und Stream-Analyzer mit festen Budgets**
   - API: Budget-Smokes fuer zentrale Hot Paths wie Event-Ingest,
     Session-Listing/Pagination, Cursor-Verarbeitung und SQLite-Zugriff.
   - Stream-Analyzer: Parser-Budgets fuer typische und grosse
     HLS-/DASH-Manifeste.
   - Ziel: grobe Performance-Regressionen im PR-Pfad blockieren, ohne
     auf fragile Mikrobenchmark-Vergleiche angewiesen zu sein.

3. **Nightly `benchstat`-Regressionen**
   - Go-Benchmarks mehrfach ausfuehren und mit `benchstat` gegen eine
     stabile Baseline vergleichen.
   - Als Nightly- oder Release-Gate fuehren, nicht als harter
     PR-Critical-Path.
   - Ziel: statistisch belastbare Performance-Regressionen erkennen.

4. **Generated-Artifact-Drift-Gate**
   - Generierte Artefakte neu erzeugen und danach einen sauberen
     Worktree erwarten.
   - Kandidaten: Schema-DDL, Contract-Fixtures, Public-API-Snapshots,
     SDK-Kompatibilitaetsdateien.
   - Ziel: manuelle Aenderungen an Quellen und generierten Artefakten
     synchron halten.

5. **Selektives Fuzzing / Property Tests**
   - Go-Fuzzing fuer Cursor, HTTP-Validation, Event-Metadaten und
     Serialisierung.
   - Property Tests fuer TypeScript-Parser und SSRF-/URL-Klassifizierung.
   - Kurze Seed-Corpus-Laeufe koennen in PRs laufen; laengere
     Fuzz-Laeufe gehoeren in Nightly-Jobs.

6. **Mutation Testing nur auf kritischen Modulen**
   - Nicht repo-weit einfuehren, da Laufzeit und False Positives schnell
     steigen.
   - Kandidaten: Event-Validierung, Cursor-Logik, HLS-/DASH-Parser,
     SRT-Health-Mapping und Security-relevante URL-Pruefung.
   - Ziel: Tests auf echte Fehlererkennung pruefen, nicht nur Coverage
     erhoehen.

## Benchmarking-Policy

Benchmarking sollte zweigeteilt bleiben:

- **PR-Gate:** robuste Budget-Smokes mit festen, grosszuegigen
  Schwellwerten.
- **Nightly/Release:** statistische Regressionserkennung mit mehreren
  Wiederholungen und Baseline-Vergleich.

Ein PR sollte nicht daran scheitern, dass ein Mikrobenchmark auf einem
Shared Runner leicht schwankt. Blockierend sollten nur klare
Budget-Verletzungen oder wiederholbare Regressionen sein.
