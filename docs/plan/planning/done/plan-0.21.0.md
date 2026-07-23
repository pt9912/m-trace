# Implementation Plan — `0.21.0` (OCI Image Publishing)

> **Status**: ✅ released — Plan archiviert in `done/`.
> Release-Tag `v0.21.0`, npm-Package-Publish und erster GHCR-Image-
> Publish sind abgeschlossen.
>
> **Vorgänger**: `0.20.0` ist als Package-Publishing-Release
> veröffentlicht und archiviert in
> [`../done/plan-0.20.0.md`](../done/plan-0.20.0.md).
>
> **Auslöser**: `0.20.0` hat GitHub Packages für npm-Artefakte
> aktiviert, Container-Image-Publishing aber ausdrücklich aus dem Scope
> genommen. Der alte Release-Text „in einem späteren Release" hatte
> keinen eigenen Trigger. `0.21.0` macht diesen Trigger explizit:
> Release-Artefakte sollen ohne lokalen Docker-Build nutzbar sein.
>
> **Release-Typ**: Minor-Release mit Lastenheft-Patch `1.1.24`
> (`RAK-121`..`RAK-125`), ohne Runtime-, Wire-, Public-API-,
> Persistenz- oder Analyzer-Schema-Änderung.

## 0. Scope

In Scope:

- versionierte GHCR-Images für `api`, `dashboard` und
  `analyzer-service`.
- lokale Make-Targets für Image-Build, Dry-Run und Publish.
- GitHub-Actions-Workflow für manuellen Dry-Run und
  `release.published`-Publish.
- Release-Doku, Lastenheft, Changelog und Roadmap aktualisieren.

Nicht in Scope:

- `latest`-Tag.
- Multi-Arch-Builds oder Signierung/Attestations.
- Production-K8s-Pflichtbetrieb.
- neue Runtime-, Wire-, API-, Persistenz- oder Analyzer-Schema-
  Funktionalität.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Ergebnis |
| --- | --- | --- |
| 0 | Aktivierung | OCI-Publishing-Trigger und Scope fixiert |
| 1 | Make-Targets | lokale Build-/Dry-Run-/Publish-Pfade vorhanden |
| 2 | Publish-Automation | GHCR-Workflow für Dry-Run und Release-Publish vorhanden |
| 3 | Doku/Normativer Stand | Releasing, Lastenheft, Changelog, Roadmap aktualisiert |
| 4 | Gates/Closeout | Checks grün, Versionen bumpbar, Tag/Release/Images publizierbar |

## 2. Tranche 0 — Aktivierung

DoD:

- [x] Container-Image-Publishing als eigener `0.21.0`-Track angelegt.
- [x] Trigger dokumentiert: GitHub-Packages-Publish ist live, aber
  Runtime-Images sind weiterhin nicht öffentlich als m-trace-Pakete
  verfügbar.
- [x] `latest`, Multi-Arch, Signierung und Production-K8s bleiben
  bewusst außerhalb des ersten OCI-Slices.

## 3. Tranche 1 — Make-Targets

DoD:

- [x] `make image-build VER=X.Y.Z` baut die drei Runtime-Images mit
  GHCR-kompatiblen Namen.
- [x] `make image-publish-dry-run VER=X.Y.Z` baut und inspiziert die
  Images ohne Push.
- [x] `MTRACE_IMAGE_PUBLISH_APPROVED=1 make image-publish VER=X.Y.Z`
  pusht nur nach expliziter Freigabe.
- [x] Der Publish-Pfad nutzt nur versionierte Tags; `latest` bleibt
  ausgeschlossen.

## 4. Tranche 2 — Publish-Automation

DoD:

- [x] `.github/workflows/publish-images.yml` ergänzt.
- [x] Workflow nutzt `GITHUB_TOKEN` mit `packages: write`.
- [x] `workflow_dispatch` unterstützt `ref`, optionalen `image_tag`
  und `dry_run`.
- [x] `release.published` veröffentlicht den Release-Tag.

## 5. Tranche 3 — Doku / Normativer Stand

DoD:

- [x] `docs/user/releasing.md` beschreibt GHCR-Image-Dry-Run,
  produktiven Publish und Rollback-Grenzen.
- [x] `spec/lastenheft.md` ist für `1.1.24` / RAK-121..RAK-125
  vorbereitet.
- [x] `CHANGELOG.md` enthält den `0.21.0`-Versionsabschnitt.
- [x] Roadmap beschreibt `0.21.0` als aktiven OCI-Publishing-Track.

## 6. Tranche 4 — Gates / Closeout

Vor Release:

- [x] `make image-publish-dry-run VER=0.21.0`
- [x] `make image-scan`
- [x] `make package-publish-dry-run`
- [x] `make generated-drift-check`
- [x] `MTRACE_RELEASE_APPROVED=1 make release-guard VER=0.21.0`
- [x] Versionstragende Artefakte auf `0.21.0` synchronisieren.
- [x] `CHANGELOG.md` von `Unreleased` nach `0.21.0` datieren.
- [x] Plan nach `docs/planning/done/` archivieren.
- [x] Tag `v0.21.0`, GitHub Release und GHCR-Publish abschließen.

Closeout-Verdict:

- RAK-121: ✅ GHCR-Namensschema final.
- RAK-122: ✅ Make-Targets verifiziert.
- RAK-123: ✅ Publish-Workflow erfolgreich:
  [`publish-images.yml` run `25804508692`](https://github.com/pt9912/m-trace/actions/runs/25804508692).
- RAK-124: ✅ Release-Dokumentation final.
- RAK-125: ✅ Tag, GitHub Release und erster Image-Publish
  abgeschlossen.

Post-Release-Nachweis:

- GitHub Release:
  <https://github.com/pt9912/m-trace/releases/tag/v0.21.0>
- npm GitHub Packages:
  [`publish-packages.yml` run `25804508709`](https://github.com/pt9912/m-trace/actions/runs/25804508709)
  veröffentlichte `@pt9912/player-sdk@0.21.0` und
  `@pt9912/stream-analyzer@0.21.0`.
- GHCR Container Packages:
  [`publish-images.yml` run `25804508692`](https://github.com/pt9912/m-trace/actions/runs/25804508692)
  veröffentlichte `ghcr.io/pt9912/m-trace-api:0.21.0`,
  `ghcr.io/pt9912/m-trace-dashboard:0.21.0` und
  `ghcr.io/pt9912/m-trace-analyzer-service:0.21.0`.
