# Contributing to m-trace

Danke für dein Interesse an m-trace. Dieser Leitfaden beschreibt das
Minimum, das ein PR braucht, um reviewbar zu sein.

## 1. Lokales Setup

- Repository klonen, dann
  [`docs/user/local-development.md`](docs/user/local-development.md)
  folgen (Toolchain-Versionen, Compose-Lab, Smoke-Targets,
  Persistenz- und Observability-Profile, Troubleshooting).
- Es ist kein lokales Go/Node erforderlich — alle Build-/Test-/
  Lint-Schritte laufen Docker-/Compose-basiert über das `Makefile`.

## 2. Build, Test, Lint

Das `Makefile` ist der Single-Entry-Point. Direkte
`pnpm --filter` / `docker build` / `go test` Aufrufe sind
nicht der unterstützte Pfad.

| Aufgabe | Kommando |
| ------- | -------- |
| CI-äquivalenter Komplettcheck (test + lint + coverage + arch + smokes) | `make gates` |
| Schneller Doku-Link- und Frontmatter-Check | `make docs-check` |
| Nur API-Tests (Go, Race-Detector) | `make api-race` |
| Nur TypeScript/Workspace-Tests | `make ts-test` |
| Bauartefakte | `make build` |
| Lab hochziehen (Compose, Default-Profil) | `make dev` |
| Lab + Observability (Prometheus + Grafana + OTel) | `make dev-observability` |

Vor jedem PR muss `make gates` lokal grün sein. CI fährt dieselben
Targets zusätzlich mit Security-Gates (siehe `0.8.5`).

## 3. Commits, Branches, Releases

- Commit-Stil: kurze Zusammenfassung (≤ 72 Zeichen) plus optionaler
  Body, der das *Warum* erklärt. Konventionelle Präfixe
  (`feat:`/`fix:`/`docs:`/`chore:`) sind willkommen, aber nicht
  erzwungen.
- PRs werden gegen `main` gestellt; CI muss grün sein, bevor
  reviewt wird.
- Release- und Tag-Verfahren: siehe
  [`docs/user/releasing.md`](docs/user/releasing.md). Patch-Releases
  folgen der dort verankerten Patch-Release-Konvention (§3.1).

## 4. Issues & Pull Requests

- Bug-Reports beschreiben das beobachtete vs. erwartete Verhalten,
  betroffene Version (Tag oder Commit-SHA) und einen minimalen
  Reproduktionspfad (idealerweise ein `make`-Target).
- Feature-Vorschläge beziehen sich auf einen Lastenheft-Eintrag
  (`F-`/`NF-`/`MVP-`/`RAK-`) oder schlagen einen neuen vor.
- PRs verweisen auf das zugehörige Issue oder den jeweiligen Plan
  in [`docs/planning/`](docs/plan/planning/) und beschreiben die
  durchgeführten Smokes/Gates.

## 5. Sicherheitsmeldungen

Sicherheitslücken **nicht** in öffentlichen Issues oder PRs
melden. Meldeweg und Scope siehe [`SECURITY.md`](SECURITY.md).
