# Implementation Plan — `0.20.0` (Package Publishing)

> **Status**: ✅ abgeschlossen — Plan archiviert in `done/`.
> Release-Tag `v0.20.0` und erster GitHub-Packages-Publish sind
> Ziel des Closeouts.
>
> **Vorgänger**: `0.19.0` ist als Decision-only-Plan archiviert in
> [`plan-0.19.0.md`](./plan-0.19.0.md). Letzter veröffentlichter
> Release vor dieser Welle: `0.18.0`.
>
> **Auslöser**: GitHub Releases existieren, aber unter
> <https://github.com/pt9912?tab=packages&repo_name=m-trace> sind noch
> keine Packages veröffentlicht. Der bisherige npm-Scope `@npm9912`
> passt nicht zum GitHub-Owner `pt9912`.

## 0. Scope

In Scope:

- publishbare npm-Pakete auf `@pt9912/...` umstellen.
- GitHub-Packages-Publish für `@pt9912/player-sdk` und
  `@pt9912/stream-analyzer` einführen.
- Apps als private Workspace-Pakete behalten.
- Release-Doku, Lastenheft, Changelog und Roadmap aktualisieren.

Nicht in Scope:

- Container-Image-Publishing.
- npmjs.org-Veröffentlichung.
- neue Runtime-, Wire-, API-, Persistenz- oder Analyzer-Schema-
  Funktionalität.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Ergebnis |
| --- | --- | --- |
| 0 | Aktivierung | Scope `@pt9912` und GitHub Packages bestätigt |
| 1 | Package-Metadaten | Workspace, SDK-/Analyzer-Referenzen und Versionen synchron |
| 2 | Publish-Automation | Workflow für Dry-Run und produktiven Publish vorhanden |
| 3 | Doku/Release | `releasing.md`, Lastenheft, Changelog, Roadmap aktualisiert |
| 4 | Gates/Closeout | Checks grün, Commit/Tag/Release/Packages veröffentlicht |

## 2. Tranche 0 — Aktivierung

DoD:

- [x] Ausgangslage geprüft: keine Packages veröffentlicht.
- [x] GitHub-Packages-Ziel `pt9912` gewählt.
- [x] Entscheidung getroffen, den ersten Publish nicht rückwirkend als
  `0.18.0`, sondern als eigenen `0.20.0`-Release auszuliefern.

## 3. Tranche 1 — Package-Metadaten

DoD:

- [x] `@npm9912/player-sdk` → `@pt9912/player-sdk`.
- [x] `@npm9912/stream-analyzer` →
  `@pt9912/stream-analyzer`.
- [x] Private Workspace-Pakete auf `@pt9912/...` synchronisiert.
- [x] `publishConfig.registry=https://npm.pkg.github.com` in den
  beiden publishbaren Paketen gesetzt.
- [x] `.npmrc` mappt `@pt9912` auf GitHub Packages.
- [x] Versionstragende Artefakte auf `0.20.0` synchronisiert.

## 4. Tranche 2 — Publish-Automation

DoD:

- [x] `.github/workflows/publish-packages.yml` ergänzt.
- [x] Workflow nutzt `GITHUB_TOKEN` mit `packages: write`.
- [x] `workflow_dispatch` unterstützt `ref` und `dry_run`.
- [x] `release.published` veröffentlicht den Release-Tag.
- [x] Build vor Publish umfasst beide publishbaren Pakete.

## 5. Tranche 3 — Doku / Normativer Stand

DoD:

- [x] `docs/user/releasing.md` beschreibt Package-Publish und Rollback-
  Grenzen.
- [x] `spec/lastenheft.md` auf Version `1.1.23` mit RAK-116..RAK-120
  aktualisiert.
- [x] `CHANGELOG.md` enthält `0.20.0`.
- [x] Roadmap beschreibt `0.20.0` als Package-Publishing-Release.

## 6. Tranche 4 — Gates / Closeout

Vor Release:

- [x] `pnpm install --lockfile-only`
- [x] `make sync-contract-fixtures`
- [x] `make ts-test`
- [x] `make lint`
- [x] `make build`
- [x] `make package-publish-dry-run`
- [x] `make generated-drift-check`
- [x] `MTRACE_RELEASE_APPROVED=1 make release-guard VER=0.20.0`
- [x] Git commit `release: prepare 0.20.0`
- [x] Tag `v0.20.0`
- [x] GitHub Release `m-trace 0.20.0`
- [x] GitHub-Packages-Publish erfolgreich:
  [`publish-packages.yml` run `25802649951`](https://github.com/pt9912/m-trace/actions/runs/25802649951)
  veröffentlichte `@pt9912/player-sdk@0.20.0` und
  `@pt9912/stream-analyzer@0.20.0`.

Closeout-Verdict:

- RAK-116: ✅ Owner-Scope-Konsistenz umgesetzt.
- RAK-117: ✅ Publishbare Pakete auf zwei Artefakte begrenzt.
- RAK-118: ✅ GitHub-Packages-Workflow vorhanden.
- RAK-119: ✅ Release-Dokumentation aktualisiert.
- RAK-120: ✅ Tag, GitHub Release und erster GitHub-Packages-Publish
  abgeschlossen.
