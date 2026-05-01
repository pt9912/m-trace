# Analyzer-Wire-Format Contract Fixtures

Geteilte Wahrheit für das JSON-Wire-Format zwischen
`apps/analyzer-service` (TypeScript-Producer) und
`apps/api/adapters/driven/streamanalyzer` (Go-Consumer). Beide Seiten
testen gegen dieselben Dateien:

- `success-master.json` — Erfolgsfall mit `playlistType: "master"`,
  einem Variant und einer Rendition. Pinnt das volle Envelope-Schema
  inklusive `analyzerKind`, `findings`-Form und `details`-Struktur.
- `error-fetch-blocked.json` — Fehlerfall mit `status: "error"`,
  `code: "fetch_blocked"`. Pinnt die Error-Envelope-Form.

## Tests

- TypeScript: `packages/stream-analyzer/tests/contract.test.ts` —
  speist eine bekannte Manifest-Eingabe in `analyzeHlsManifest`,
  serialisiert das Result und vergleicht byte-equal gegen
  `success-master.json`. Jede TS-Output-Drift bricht den Test.
- Go: `apps/api/adapters/driven/streamanalyzer/contract_test.go` —
  liest beide Dateien per `go:embed`, parst sie via
  `parseSuccessResponse` / `parseDomainError`, und prüft die
  resultierenden Domain-Strukturen feldgenau. Jede Drift, die das
  Go-Decoding bricht, fällt auf.

## Updates

Wenn das Format absichtlich erweitert wird:

1. Code-Änderung committen.
2. TS-Test zeigt den Diff — die Fixture mit `vitest -u` (oder
   manuell) aktualisieren.
3. Go-Test gegen die neue Fixture prüfen.
4. Drift in einem Pflichtfeld (z. B. neuer `analyzerKind`-Wert)
   bedingt synchrone Anpassung beider Seiten — das ist der ganze
   Sinn dieser Fixtures.
