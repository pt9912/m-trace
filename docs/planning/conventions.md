# Conventions

> **Status**: temporär. Aus dem Backend-Spike verdichtete Engineering-
> Conventions, die ab `0.1.0` verbindlich gelten. Inhalte wandern
> beim Schreiben von `spec/architecture.md` (roadmap §2 Schritt 5)
> und beim Ausbau von `docs/user/quality.md` schrittweise an die
> thematisch passenden Stellen — siehe Migration-Plan unten.
>
> **Quelle**: `docs/spike/backend-stack-results.md` (vollständige
> Notizen). Dieses Doc verdichtet nur die für `0.1.0+` relevanten
> Punkte.

---

## 1. Hexagon ohne DI-Container

Der Spike hat gezeigt: Go braucht keine Annotation-Magie, um
Hexagon-Boundaries zu enforcen. Pro Adapter gilt:

- Compile-Time-Check: `var _ Interface = (*Impl)(nil)` neben jeder
  Adapter-Implementierung. Bricht der Build, wenn die Implementation
  vom Port-Interface abweicht.
- Package-Boundaries: `hexagon/` darf keine HTTP-, DB-, Framework-,
  Docker- oder OTel-Implementierungstypen importieren.
- Reine Funktionen / `var`-Singletons im Inner-Hexagon — keine DI-
  Container-Annotations (das wäre Kotlin/Micronaut-Pattern).

**Beibehalten**.

---

## 2. Test-Stack einheitlich

`testing` + `httptest` aus der Standardbibliothek deckt sowohl Unit-
als auch Integration-Tests ab. Keine externen Test-Frameworks
erforderlich (Testify/Ginkgo/Gomega bleiben bewusst außen vor).

- Unit-Tests neben dem Code (gleicher Package), `_test.go`-Suffix.
- Integration-Tests via `httptest.NewServer` mit voller Wire.
- `t.Parallel()` für unabhängige Tests.
- Stubs/Spies kleinflächig pro Test, kein Mock-Framework.

---

## 3. Linting

`golangci-lint` mit Default-Lintern als Soll-Gate:

- `govet`, `errcheck`, `staticcheck`, `unused`, `ineffassign`.
- Custom-Regeln und Suppressions sind ausgeschlossen.
- `make lint` ruft `docker build --target lint` auf (siehe
  `docs/user/quality.md` §1).

---

## 4. Docker-only-Workflow

Alle Build-, Test- und Lint-Schritte laufen via
`docker build --target <stage>`. Lokales Go ist optional.

- `Dockerfile` mit Stages `deps`, `compile`, `lint`, `test`,
  `build`, `runtime` — Pattern aus `docs/planning/plan-spike.md` §14.11.
- `Makefile` mit `make test`, `make lint`, `make build`, `make run`,
  `make deps`, `make compile`, `make clean` — alle delegieren an
  Docker.
- Frischer Repo-Clone funktioniert ohne JDK/Go-Toolchain.

---

## 5. CI-Artifacts

Beim CI-Setup (roadmap §2 Schritt 1+) hochladen — Pattern analog
zu `d-migrate/.github/workflows/build.yml`:

- Test-Results (`build/test-results/test/*.xml` bzw.
  `apps/api/build/...`)
- Coverage-Reports (sobald `go test -cover` aktiv ist)
- Lint-Reports (`golangci-lint`-Output)

`outputs.cacheIf { false }` für Test-Tasks (verhindert stale
Coverage-Counter aus Build-Cache).

---

## 6. Multi-Modul erst on-demand

`apps/api/` bleibt im MVP **Single-Modul**. Aufteilung ist nur
gerechtfertigt, wenn:

- Hexagon-Boundaries durch Disziplin nicht mehr halten
  (häufiger falscher Import-Direction-Bruch).
- Buildzeit oder Test-Velocity unter Single-Modul leidet.

Dann: per `go.work` oder `internal/`-basierte Sub-Modul-Splits.
Alternative: separate Apps unter `apps/<name>/` für unabhängige
Services (siehe Lastenheft Mono-Repo-Struktur).

---

## 7. Migration-Plan

Wenn die Ziel-Docs existieren bzw. erweitert werden, wandern
diese Conventions hierher:

| Convention | Ziel-Doc | Trigger |
|---|---|---|
| §1 Hexagon ohne DI-Container | `spec/architecture.md` | roadmap §2 Schritt 5 |
| §2 Test-Stack | `docs/user/quality.md` (§2 erweitern) | beim ersten Update von quality.md |
| §3 Linting | `docs/user/quality.md` §1 (bereits dort) | bereits dort — §3 hier kann sofort entfernt werden, sobald migriert |
| §4 Docker-only-Workflow | `docs/user/quality.md` §1/§2/§4 (bereits dort) | bereits dort |
| §5 CI-Artifacts | `docs/user/quality.md` §6 (Platzhalter erweitern) | beim CI-Setup (`0.1.0`+) |
| §6 Multi-Modul on-demand | `spec/architecture.md` | roadmap §2 Schritt 5 |

Sobald alle sechs migriert sind, kann `conventions.md` ersatzlos
entfernt werden — der Eintrag in roadmap §3 (Releases) und der
spike-protocol-Quelle bleibt als Audit-Spur.
