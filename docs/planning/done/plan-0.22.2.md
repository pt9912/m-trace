# Implementation Plan — `0.22.2` (Go-Stdlib-Security-Patch + Trivy-Ignore-Erweiterung)

> **Status**: ✅ released und archiviert in `done/`.
>
> **Vorgänger**: `0.22.1` ist als devalue-Security-Patch + Nightly-Audit-Mirror
> veröffentlicht und archiviert in
> [`./plan-0.22.1.md`](./plan-0.22.1.md).
>
> **Auslöser**: Erster echter Treffer des in `0.22.1` eingeführten
> Nightly-`security-audit.yml`-Workflows (Run
> [`26859599263`](https://github.com/pt9912/m-trace/actions/runs/26859599263),
> 2026-06-03 02:14 UTC). `govulncheck` meldet zwei Go-Stdlib-Findings,
> beide fixed in `go1.26.4`:
>
> - **GO-2026-5039** — `net/textproto`: ungescaptes Input-Echo in Errors;
>   verwundbarer Aufrufpfad
>   `auth.VaultSecretBackend.LoadSigningKeys → io.ReadAll →
>   textproto.Reader.ReadMIMEHeader`.
> - **GO-2026-5037** — `crypto/x509`: ineffizientes Hostname-Kandidaten-
>   Parsing; verwundbarer Aufrufpfad
>   `auth.NewRedisIssuanceRateLimiter → x509.HostnameError.Error` plus
>   `coverage.main → fmt.Fprintf → x509.Certificate.Verify`.
>
> Issue #3 (Labels `security`, `audit`, `plan-0.8.5`) wurde durch
> `scripts/open-security-audit-issue.sh` automatisch eröffnet — der mit
> `0.22.1` geschlossene Push-Audit-Gap hat sein erstes Finding gemeldet.
>
> **Release-Typ**: Patch-/Tooling-Release ohne Lastenheft-Patch,
> ohne Runtime-, Wire-, Public-API-, Persistenz- oder
> Analyzer-Schema-Änderung. Versionstragend, weil
> `image-publish`/`package-publish` die neue Go-Stdlib-Closure
> (gepatchte `mtrace-api`-Layer) als neues Image-Tag veröffentlicht.

## 0. Scope

In Scope:

- `golang:1.26.3 → 1.26.4` an sechs Build-/Test-Image-Stellen:
  `apps/api/Dockerfile` (Multistage-`deps`-Stage), `Makefile` (`vuln-check`),
  `apps/api/Makefile` (`arch-check`, `benchmark-smoke`, `fuzz-check`,
  `mutation-report`).
- `.security/vulnignore.yaml`: fünf `perl-base`-CVEs aus dem
  `debian:13.5`-Base-Image der `mtrace-dashboard`- und
  `mtrace-analyzer-service`-Images dokumentiert ignoriert
  (CVE-2026-42496, CVE-2026-42497, CVE-2026-8376, CVE-2026-9538,
  CVE-2026-48962); bereits seit Commits `97d8a29`, `04a6419`,
  `76ba007` auf `main`. Jeweils mit Begründung und `expires`-Termin.
- README-Sprach-Split (Commit `c39b74a`): `README.md` ist jetzt englisch,
  Deutsche Doku verschoben nach `README.de.md`.
- Doku: `CHANGELOG.md`, Roadmap-Eintrag.

Nicht in Scope:

- Go-Modul-Upgrades (das `go.mod`-`go 1.26.0`-Pin bleibt; nur das
  Build-Image bestimmt die Stdlib-Patch-Version).
- Runtime-/Analyzer-Funktionalität, Lastenheft-Patch.
- Behebung des `webrtc-drift.yml`-Firefox-Findings vom 2026-06-03
  (siehe §6 Beobachtung) — separater Folge-Plan.
- Anpassung der Benchmark-Budgets oder neue Bench-Targets.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Ergebnis |
| --- | --- | --- |
| 1 | Trivy-Ignore-Erweiterung perl-base + README-Split | Commits `97d8a29`, `04a6419`, `76ba007`, `c39b74a` bereits auf `main` |
| 2 | Go-Stdlib-Bump `1.26.3 → 1.26.4` an sechs Image-Referenzen | `make vuln-check` lokal grün: "No vulnerabilities found." |
| 3 | Release-Closeout | Versions-Sweep auf `0.22.2`, Plan archiviert, Tag/Release, Issue #3 geschlossen |

## 2. Tranche 1 — Trivy-Ignore + README-Split (bereits auf main)

Commit-IDs: `97d8a29` (perl-base CVE-2026-42496/8376/9538 für dashboard +
analyzer-service), `04a6419` (CVE-2026-42497 Hardlink-Variante),
`76ba007` (CVE-2026-48962 IO::Compress-Glob), `c39b74a` (README-Split).

Hintergrund perl-base: Die fünf Findings tauchten in den
Trivy-DB-Updates vom 2026-05-28 und 2026-05-31 auf. In den
`mtrace-dashboard`- und `mtrace-analyzer-service`-Runtime-Images ist
`perl-base` als Debian-Base-Paket vorhanden, aber im Runtime nie
aufgerufen (kein Perl-Pfad, kein User-controlled Tar/Regex/Glob).
Trixie hat (noch) keinen Upstream-Fix; Forky liefert die Patches.
Pro Eintrag dokumentierte `expires`-Termine (90-Tage-Default).

DoD:

- [x] `.security/vulnignore.yaml` enthält fünf neue `perl-base`-Einträge
  mit Begründung und `expires`-Termin (siehe Diff `v0.22.1..HEAD`).
- [x] `make image-scan` rendert pro Image die `.trivyignore`-Liste
  automatisch (vgl. `scripts/render-trivyignore.sh`); Verifikation
  durch Nightly-Run `26859599263` (Trivy-Stage `success`).
- [x] README-Sprach-Split: GitHub zeigt englisches `README.md` als
  Top-Level, deutsche Übersicht verlinkt unter `README.de.md`.

## 3. Tranche 2 — Go-Stdlib-Bump

Hintergrund: `golang:1.26.3` enthält die zwei Stdlib-CVEs aus dem
Issue-Header. Bump auf `golang:1.26.4` zieht die gepatchte Stdlib in
alle Build-/Test-Container; `go.mod` bleibt auf `go 1.26.0` (Pin
beschreibt die Sprach-Toolchain-Mindeststelle, nicht die Patch-
Version — die kommt aus dem Image).

DoD:

- [x] `apps/api/Dockerfile:24` `FROM golang:1.26.3 AS deps`
  → `golang:1.26.4`.
- [x] `Makefile:605` `vuln-check`-Target nutzt `golang:1.26.4`.
- [x] `apps/api/Makefile` `arch-check`, `benchmark-smoke`,
  `fuzz-check`, `mutation-report` nutzen `golang:1.26.4`.
- [x] Verifikation lokal: `make vuln-check` zeigt
  `"No vulnerabilities found."`; Aufruf gegen `golang:1.26.4` zieht
  die gepatchte Stdlib (Image-Digest
  `sha256:68cb6d68bed024785b69195b89af7ac7a444f27791435f98647edff595aa0479`).
- [ ] Bestätigung in CI nach Tag: nächster `build.yml::Security
  gates`-Lauf zeigt `vuln-check` `success`, kein Re-Eröffnen von
  Issue #3 im darauffolgenden `security-audit.yml`-Nightly.

## 4. Tranche 3 — Release-Closeout

DoD:

- [x] Versions-Sweep `0.22.1` → `0.22.2` an allen 5× `package.json`,
  `apps/api/cmd/api/main.go::serviceVersion`,
  `packages/player-sdk/src/version.ts::PLAYER_SDK_VERSION`,
  `contracts/sdk-compat.json::sdk_version`, 21 Analyzer-Fixtures
  (`spec/contract-fixtures/analyzer/*.json`), 20 testdata-Kopien
  (`make sync-contract-fixtures`), hartkodierte Test-Strings
  (`apps/api/adapters/driven/streamanalyzer/{contract_test.go,
  http_test.go}`, `packages/player-sdk/tests/tracker.test.ts`,
  `packages/stream-analyzer/tests/version.test.ts`).
- [x] `CHANGELOG.md`: neuer `[0.22.2] - 2026-06-03`-Abschnitt mit
  `Security`-Block (Go-Bump + perl-base Trivy-Ignores) und
  `Changed`-Block (README-Split).
- [x] `docs/planning/in-progress/roadmap.md`: 0.22.2-Eintrag im
  Lieferstand, im Phase-Block und in der Release-Tabelle; §1.1+§1.2
  nach Bump neu geschrieben (releasing.md §1 Wartungsregel).
- [x] `make gates` lokal: Pre-Commit-Lauf grün außer am inhärenten
  `generated-drift-check` (Working-Tree-Bumps vs. HEAD); Post-Commit-
  Re-Lauf gegen den `chore: release 0.22.2`-HEAD ist clean.
- [x] `MTRACE_RELEASE_APPROVED=1 make release-guard VER=0.22.2`.
- [x] Plan direkt in `done/plan-0.22.2.md` archiviert, Status auf
  „released" gesetzt.
- [x] Tag `v0.22.2`.
- [x] Issue #3 mit Verweis auf Release-Tag geschlossen.

## 5. Wave-2-Quality-Gates-Verdict (für Tag-Annotation)

Verifikation der jüngsten Nightly-Runs vor dem Release-Tag, laut
[`docs/user/releasing.md`](../../user/releasing.md) §3.1:

| Gate | Run-ID | Status |
| --- | --- | --- |
| Benchmark regression (latest schedule) | `26859227837` | success (Patch-Pfad ist `make benchmark-smoke`; Nightly ist Bonus) |
| Fuzz nightly (latest schedule) | `26859493259` | success (kein offenes `fuzz`-Issue) |
| Mutation nightly (latest 3) | `26859551494`, `26793776028`, `26731258932` | success × 3 (Score-Trend stabil) |
| Security audit nightly (Auslöser) | `26859599263` | failure (Issue #3, dieser Release schließt es) |

**Verdict**: Wave-2-Gates erfüllt — fuzz hat kein offenes Crash-Issue,
mutation-Trend stabil, benchmark wird über `make benchmark-smoke`
in `make gates` abgesichert (Patch-Pfad).

## 6. Beobachtung: webrtc-drift Firefox-Failure (separates Folge-Item)

Der `webrtc-drift.yml`-Nightly vom 2026-06-03 (Run
[`26858728018`](https://github.com/pt9912/m-trace/actions/runs/26858728018))
ist auf **Firefox** failed; Chromium passt. Reproduzierbarer Drift in
`getStats()`-Sollfeldern (`roundTripTime`, `availableOutgoingBitrate`,
`framesDecoded`, `framesPerSecond`) — keine `0.22.2`-relevante
Änderung, sondern Browser-Drift gegen Spec-Allowlist
([`spec/telemetry-model.md`](../../../spec/telemetry-model.md) §1.4 + §3.5.2).

Releasing.md §3.1 listet webrtc-drift nicht als Patch-Release-Blocker
(nur fuzz blockt Patches); der Failure wird in einem separaten
Folge-Plan adressiert (Drift-Review + Sampling-Update). Kein
auto-Issue weil `DRIFT_AUTO_ISSUE` nicht gesetzt; manueller
Folge-Eintrag in der Roadmap nach diesem Release.
