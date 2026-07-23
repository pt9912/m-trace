# Implementation Plan — `0.22.4` (x/net-Security-Patch + Trivy-Pin + Ingest-Rate-Limit-ENV)

> **Status**: ✅ released 2026-06-23 — Tag `v0.22.4`, GHCR- + npm-Publish
> grün (publish-images
> [`28007976324`](https://github.com/pt9912/m-trace/actions/runs/28007976324),
> publish-packages
> [`28007976457`](https://github.com/pt9912/m-trace/actions/runs/28007976457)).
> Archiviert in `done/`.
>
> **Vorgänger**: `0.22.3` ist als Security-/CI-Sammel-Patch veröffentlicht
> und archiviert in
> [`./plan-0.22.3-webrtc-drift.md`](./plan-0.22.3-webrtc-drift.md).
>
> **Auslöser**: Fünfter echter Treffer des in `0.22.1` eingeführten
> Nightly-`security-audit.yml`-Workflows (Run
> [`27996614696`](https://github.com/pt9912/m-trace/actions/runs/27996614696),
> 2026-06-23, Issue
> [#9](https://github.com/pt9912/m-trace/issues/9)). Der
> Trivy-Image-Scan meldet sechs HIGH-CVEs in `golang.org/x/net v0.53.0`
> im `usr/local/bin/api`-gobinary; `govulncheck` und `pnpm audit`
> (nur low/moderate) waren grün.

## 0. Scope

Patch-Release ohne neue User-Surface (`docs/user/releasing.md` §3.1):
kein Lastenheft-Patch, keine RAK-Verifikationsmatrix, normativer Stand
bleibt `1.1.24`.

Inhalt:

1. **`golang.org/x/net 0.53.0 → 0.56.0`** in `apps/api/go.mod` (transitiv
   `golang.org/x/sys 0.43.0 → 0.46.0`, `golang.org/x/text 0.36.0 →
   0.38.0`). Behebt die sechs vom Trivy-Image-Scan gemeldeten HIGH-CVEs
   `CVE-2026-25680`/`-25681`/`-27136`/`-39821`/`-42502`/`-42506`
   (HTML-Parsing bzw. `idna`-Punycode; upstream gefixt in `0.55.0`).
   `govulncheck` war grün, weil der Call-Graph die verwundbaren Pfade
   nicht erreicht; Trivy scannt den im Binary eingebetteten Modulgraphen
   unabhängig davon — daher der Gate-Fail.
2. **`undici`-`pnpm.overrides` `^7.28.0`** im Root-`package.json`
   (GHSA-vmh5-mc38-953g) — bereits vor diesem Patch auf `main`
   (Commit `073393f`), rollt hier mit ein.
3. **Trivy-Scanner-Image `0.71.0 → 0.71.2`** in `Makefile`
   (`TRIVY_IMAGE`) — übernimmt den vom Nightly gemeldeten
   Versions-Notice.
4. **ENV-konfigurierbarer Ingest-Rate-Limiter** (`MTRACE_RATE_LIMIT_CAPACITY`/
   `-REFILL`, Commit `14f3e64`) — macht den zuvor hart auf `100/100`
   codierten Token-Bucket überschreibbar; Default unverändert `100/100`,
   kein Verhaltensbruch, keine Änderung der Performance-Charakteristik
   (Patch-konform, §2.5).
5. **Load-/Soak-Smoke-Readback** zählt persistierte Events per direktem
   SQLite-`COUNT(*)` statt O(N)-HTTP-Pagination (`dd37c3d`, Test/CI-only).

## 1. Tranchen-Übersicht

| # | Tranche | Status |
| - | ------- | ------ |
| 1 | x/net-Bump + `make vuln-check`/`make image-scan` grün (Issue #9) | ✅ |
| 2 | Versions-Bump 0.22.3 → 0.22.4 (alle §3.1-Stellen) + Fixtures-Sync | ✅ |
| 3 | CHANGELOG + Roadmap + Plan-Closeout | ✅ |
| 4 | `make gates` grün + Release-Commit | ✅ |
| 5 | Tag + GitHub-Release + GHCR-/npm-Publish | ✅ |

## 2. Verifikation

- `make vuln-check` → „No vulnerabilities found." (govulncheck im
  `golang:1.26.4`-Container).
- `make image-scan` → `usr/local/bin/api`-gobinary **0 Findings**, alle
  drei Runtime-Images grün, exit 0 (Trivy `0.71.2`).
- `make gates` → CI-äquivalenter Komplettcheck grün.
- Bump-Pattern-Sweep: keine verbleibenden `0.22.3`-Quellanker (nur ein
  bewusst historischer Kommentar in
  `packages/stream-analyzer/vitest.config.ts`).

## 3. Quality-Gates-Verdict (für Tag-Annotation, §3.1)

Letzte Nightly-Beobachtungs-Gates vor dem Tag (alle auf `main`,
2026-06-23):

- **benchmark.yml** Run
  [`27995729054`](https://github.com/pt9912/m-trace/actions/runs/27995729054)
  ✅ (für Patch nicht hart erforderlich, dennoch grün).
- **fuzz.yml** Run
  [`27995968019`](https://github.com/pt9912/m-trace/actions/runs/27995968019)
  ✅; keine offenen Issues mit Label `fuzz` → Tag nicht blockiert.
- **mutation.yml** letzte drei Läufe
  ([`27996021298`](https://github.com/pt9912/m-trace/actions/runs/27996021298),
  [`27925206926`](https://github.com/pt9912/m-trace/actions/runs/27925206926),
  [`27890458779`](https://github.com/pt9912/m-trace/actions/runs/27890458779))
  ✅, Score-Trend stabil (nicht-blockierend).

## 4. Release-Closeout (ausgeführt 2026-06-23)

Ausgeführte Schritte (alle grün):

1. `MTRACE_RELEASE_APPROVED=1 make release-guard VER=0.22.4` → „dry-run ok".
2. `make package-publish-dry-run` + `make image-publish-dry-run VER=0.22.4`.
3. Pre-Tag-Smokes: `make build`, `make smoke-cli` (11 Probes),
   `make smoke-analyzer`. Compose-/Browser-Smokes
   (`smoke-observability`/`browser-e2e`/`smoke-mediamtx`/`-srt`/`-dash`/
   `-webrtc-prep`) nicht gefahren — von einem Dep-Bump unberührte Pfade.
4. `git tag -a v0.22.4` (Verdict in Annotation) + `git push origin main`
   + `git push origin v0.22.4` (Commit `2a4f67e`).
5. `gh release create v0.22.4` → `release.published` triggert
   `publish-images.yml` (`28007976324`) + `publish-packages.yml`
   (`28007976457`), beide ✅.

Veröffentlicht: `ghcr.io/pt9912/m-trace-{api,dashboard,analyzer-service}:0.22.4`,
`@pt9912/{player-sdk,stream-analyzer}@0.22.4`. Issue #9 war bereits
geschlossen (Fix verifiziert); `## [Unreleased]` bleibt für die nächste
Tranche offen.
