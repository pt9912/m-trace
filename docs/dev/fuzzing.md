# Fuzzing & Property Tests

> **Status**: Aktiv seit `plan-0.9.5` Tranche 3 (RAK-Wave-2 /
> [`extra-gates.md`](../plan/planning/extra-gates.md) §3.5).
> PR-Pfad: `make fuzz-check` mit kurzem Budget; Nightly-Pfad:
> [`.github/workflows/fuzz.yml`](../../.github/workflows/fuzz.yml)
> mit längerem Budget plus Auto-Issue bei Crash.

## 1. Aktive Fuzz-Targets

Sechs Go-Fuzz-Targets in vier Packages, plus drei TypeScript-
Property-Test-Suites in zwei Packages.

### 1.1 Go (`apps/api`)

| Target | Package | Was wird gefuzzt | Pflicht-Invariante |
|---|---|---|---|
| `FuzzDecodeListSessionsCursor` | `apps/api/adapters/driving/http` | random Strings → `decodeListSessionsCursor` | Erlaubte Errors: `errCursorInvalidLegacy`/`errCursorInvalidMalformed`/`errCursorExpired`. Panic = Bug. |
| `FuzzDecodeSessionEventsCursor` | `apps/api/adapters/driving/http` | random Strings → `decodeSessionEventsCursor` | wie oben; dekodierter Cursor darf nicht komplett leer sein. |
| `FuzzWireBatchDecode` | `apps/api/adapters/driving/http` | random JSON-Bytes → `wireBatch`-Decode + `to*Inputs`-Mapper | Längen-Stabilität (events.length == payload.events.length); kein Panic; Mapping deterministisch. |
| `FuzzValidateReservedEventMeta` | `apps/api/hexagon/application` | random Key+Value-Paare (string/int64/float64/bool) → `validateReservedEventMeta` | Reserved-Namespace-Keys (`network.*`/`timing.*`/`webrtc.*`) liefern entweder nil oder typed Validation-Error. Panic = Bug. |
| `FuzzValidateUnavailableReason` | `apps/api/hexagon/application` | random Strings → `validateUnavailableReason` | Reason-Enum aus `spec/telemetry-model.md` §1.4 + `^[a-z0-9_]{1,64}$`-Pattern. |
| `FuzzMapMediaMtxItem` | `apps/api/adapters/driven/srt/mediamtxclient` | random JSON-Bytes als MediaMTX-`/v3/srtconns/list` → `mapItem` | Domain-Felder bleiben non-negative auch bei Negativ-Source-Werten. **Erstfund**: `mbpsLinkCapacity=-1` produzierte `AvailableBandwidthBPS=-1_000_000` (Tranche-3a-Commit `53adbab` plus Fix in `mapping.go`). |

### 1.2 TypeScript (`packages/`)

| Test-File | Package | Was wird property-getestet |
|---|---|---|
| `tests/hls-parser.property.test.ts` | `@pt9912/stream-analyzer` | jede Eingabe mit `#EXTM3U`-Header → `analyzerKind:"hls"` + definierter `playlistType`; non-HLS/non-DASH-Bodies → `manifest_not_supported`. |
| `tests/dash-parser.property.test.ts` | `@pt9912/stream-analyzer` | jede `<?xml`/`<MPD`-Eingabe → Detector klassifiziert als `dash`; well-formed MPD → deterministisches Result mit `details.type`/`details.live`/`summary.itemCount`. |
| `tests/redact.property.test.ts` | `@pt9912/player-sdk` | bounded ASCII/Sentinel-Inputs → `redactUrl` throwed nicht; JWT-Shape-Pfadsegmente → `:redacted`; lange Hex-Pfadsegmente → `:redacted`. |

## 2. Lokale Reproduktion

### 2.1 Pre-Commit-Pfad: `make fuzz-check`

Default 30 s pro Go-Target plus die TS-Property-Tests:

```bash
make fuzz-check
```

Override `FUZZTIME` für längere Pre-Push-Läufe:

```bash
FUZZTIME=120s make api-fuzz-check
```

`make fuzz-check` ist **opt-in** und nicht Teil von `make gates`.

### 2.2 Einzel-Target-Reproduktion

Go:

```bash
cd apps/api
go test -run='^$' -fuzz='^FuzzMapMediaMtxItem$' -fuzztime=2m \
  ./adapters/driven/srt/mediamtxclient
```

TypeScript:

```bash
make ts-test    # property-tests laufen mit den Unit-Tests
```

### 2.3 Crash-Reproduktion aus dem Nightly

Wenn der Nightly-Workflow ein Issue eröffnet (Label `fuzz`):

1. **Artefakt herunterladen** —
   `gh run download <run-id> --name fuzz-nightly-<run-id>`. Im
   Verzeichnis `.tmp/fuzz/crashes/` liegen die Crash-Inputs auf
   demselben relativen Pfad wie im Repo
   (`apps/api/<package>/testdata/fuzz/<Target>/<id>`).
2. **In den Repo committen** — Crash-Files an die produktiven
   `testdata/fuzz/`-Pfade kopieren. Damit werden sie zu
   deterministischen Regression-Seeds, die jeder reguläre
   `make api-test`-Lauf mitläuft.
3. **Reproduzieren** —
   `go test -run='Fuzz<Target>/<id>' ./<package>` (Go nutzt das
   Slash-Format für Seed-Lookups).
4. **Fixen + Seed-File behalten** als permanenter Regression-Test.

## 3. Sample-Korpus / Seed-Pfade

Go-Fuzz hält zwei Korpus-Schichten:

- **In-File-Seeds** (`f.Add(...)` in den `_fuzz_test.go`-Files):
  versionskontrolliert, definieren die Initial-Inputs (Drift-
  Pfade, gültige Beispiele).
- **Generierte Korpus-Files** (`testdata/fuzz/<Target>/`):
  von `go test -fuzz` automatisch angelegt für Crash-Inputs;
  bleiben im Repo nur, wenn der Nightly-Workflow einen Crash
  findet und der Operator das Crash-File als Regression-Seed
  committet.

TypeScript-Property-Tests nutzen `fast-check`s Default-Seed plus
die Determinismus-Generators in den Test-Files (`fc.constantFrom`,
`fc.array` mit fixer Länge). Lehre aus Tranche 3b: **kein
`fc.webUrl(...)` und kein `fc.stringMatching(...).filter(...)`**
— das hat fast-check 4.4 in einen Discard-Loop geschickt
(vitest-Workers liefen 30+ min auf 97% CPU). Schutznetz:
`interruptAfterTimeLimit: 4_000` als shared `FC_OPTIONS` in den
Property-Test-Files; falls fast-check trotz fixer Generators wieder
in eine Falle läuft, fail der Test mit eindeutiger Timeout-Meldung
statt zu hängen.

## 4. Wartung

- **Neuer Fuzz-Target hinzufügen**: `_fuzz_test.go`-File neben dem
  produktiven Code anlegen (`<feature>_fuzz_test.go` oder
  `<feature>_fuzz_internal_test.go`, je nach Package-Sichtbarkeit).
  Make-Target greppt automatisch nach `^func Fuzz...` — keine
  Registry-Pflege nötig. In dieser Doku aktualisieren.
- **Crash-Triage-Pfad**: Issue-Label `fuzz` filtert.
  Sofort-Aktion: Crash-File ins Repo, dann fixen. Verlängerte
  Triage > 7 Tage produziert ein zusätzliches Risiko-Backlog-
  Item (`docs/plan/planning/in-progress/risks-backlog.md`).
- **Quarantäne-Pfad**: ein Fuzz-Target kann **nicht** wie ein
  Benchmark in Quarantäne gelegt werden — Plan-DoD §4 hat keine
  vergleichbare Mechanik definiert. Wenn ein Fuzz-Target
  permanent flaky wird (wäre echter Bug-Indikator), gehört es
  ins Backlog plus ein Skip mit Begründung im Test-File selbst.
