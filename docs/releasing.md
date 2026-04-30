# Releasing — m-trace

> **Status**: Skeleton. CI-Verifikation ist konkretisiert; Asset-Liste
> und Branching-Modell werden befüllt, sobald
> [`OE-7`](./roadmap.md) (Release-Konvention) entschieden ist.
> Bezug: AK-11, DoD §18 (Lastenheft).

## 0. Zweck

Dieses Dokument beschreibt den minimalen, reproduzierbaren
Release-Ablauf für m-trace. Der Ablauf ist versionsunabhängig
formuliert und verwendet Platzhalter der Form `X.Y.Z`. Er gilt für
alle Releases aus dem Release-Plan (Lastenheft §13, Roadmap §3 —
RAK-1..RAK-46).

## 1. Vorbereitung

```bash
VER="X.Y.Z"
TAG="v$VER"
```

Vor jedem Release:

- noch nicht veröffentlichte Änderungen stehen unter `## [Unreleased]`
  in `CHANGELOG.md`; ein datierter Versionsabschnitt entsteht erst mit
  dem Release-Commit.
- `CHANGELOG.md` auf den Zielstand bringen.
- betroffene Plan-, Status- und Nutzungsdokumente aktualisieren
  (`docs/roadmap.md`, `docs/architecture.md`, `apps/api/README.md`).
- Roadmap §1.1 und §1.2 nach dem Release-Bump neu schreiben (siehe
  `docs/roadmap.md` §7 Wartungsregel).
- offene `OE-X` und `R-X` durchsehen — Einträge, die mit dem Release
  aufgelöst werden, aus den Tabellen entfernen.

## 2. Verifikation

Vor Tag und GitHub-Release müssen die Root-Targets grün sein:

```bash
make test
make lint
make coverage-gate
make arch-check
make build
```

Erfolgskriterien:

- alle fünf Targets exit code 0.
- `golangci-lint`-Stage liefert keine Findings.
- `go test ./...` deckt mindestens die Pflichttests aus
  `docs/spike/backend-api-contract.md` §11 ab.
- Coverage-Gate liegt bei mindestens 90 %.
- Architektur-Grenzen bleiben laut `make arch-check` intakt.

CI-Zielplattform für `0.1.0` ist GitHub Actions auf `ubuntu-24.04`.
Workflow-Name: `build`.

```bash
gh run watch --workflow build.yml
```

## 3. Release-Commit und Tag

> **TODO bis OE-7**: Branching-Modell festlegen (Trunk-based auf
> `main` vs. GitFlow `develop` → `main`). Davon hängt ab, ob der
> Release-Commit direkt auf `main` landet oder über einen Merge aus
> `develop`. Tag-Format ebenfalls fixieren — vorgeschlagen: `vX.Y.Z`
> (SemVer, kein Pre-Release-Suffix für Hauptreleases).

```bash
git commit -m "chore(release): vX.Y.Z"
git tag -a "$TAG" -m "Release X.Y.Z"
git push origin main
git push origin "$TAG"
```

## 4. GitHub-Release

> **TODO bis OE-7**: Asset-Liste, Source-Bundle, Container-
> Image-Pfad (z. B. `ghcr.io/pt9912/m-trace-api:X.Y.Z`) und Pretty-
> Print der Release-Notes festlegen. Vorlage analog cmake-xray /
> d-migrate-Pattern.

Mindestumfang:

- Release-Notes aus dem `CHANGELOG.md`-Versionsabschnitt extrahieren.
- Release-Titel: `m-trace X.Y.Z`.
- Tag: `vX.Y.Z`.

```bash
gh release create "$TAG" \
    --title "m-trace $VER" \
    --notes-file <changelog-extract>
```

## 5. Post-Release

- `CHANGELOG.md` öffnet einen neuen `## [Unreleased]`-Abschnitt.
- `docs/roadmap.md` §3 (Release-Übersicht) aktualisiert den Status
  des veröffentlichten Releases (`⬜ → ✅`).
- Folge-ADRs, die mit dem Release entstehen oder fällig werden,
  in `docs/roadmap.md` §4 ergänzen.

## 6. Rollback

> **TODO bis OE-7**: Rollback-Szenarien dokumentieren analog
> d-migrate `releasing.md` §6 (Tag noch lokal, Tag bereits gepusht,
> GitHub-Release zurückziehen, CI-Build nach Release fehlschlägt).

## 7. Referenzen

- Lastenheft §14 — Akzeptanzkriterien (AK-11).
- Lastenheft §18 — Definition of Done für den MVP.
- `docs/roadmap.md` §3 — Release-Übersicht und RAK-Akzeptanzkriterien.
- `docs/roadmap.md` §5 — Offene Entscheidungen (OE-7).
- `CHANGELOG.md` — Versionsverlauf.
