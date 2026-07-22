# AGENTS.md — Briefing für AI-Coding-Agenten

## 1. Was diese Datei ist

Onboarding-Briefing für jede AI-Session, die in diesem Repo Code oder
Dokumentation ändert. Sie verweist auf die kanonischen Quellen und formuliert
die Hard Rules, die der Implementation-Agent immer einhalten muss.

**Bei Konflikt zwischen dieser Datei und einer kanonischen Quelle gilt die
kanonische Quelle** (Source Precedence — siehe
[`harness/README.md`](harness/README.md)).

Strukturregeln (ID-Schemata, Verzeichniskonventionen, Adaptionen gegenüber der
Baseline, Modus-Deklarationen pro Sub-Area, Sensor-Bindungsklassen) leben in
[`harness/conventions.md`](harness/conventions.md).

Das **Regelwerk der adoptierten Baseline** ist die **präsente, nachschlagbare
Vertiefung** zu diesem Briefing: ein self-navigierbares **Modul-Bundle**
(`README.md` = Index). Es ist committet vendored unter
`.harness/baseline/v3.5.0/{regelwerk,templates}/` (Regelwerk *und* Templates
parallel, netzlos materialisiert samt `SHA256SUMS`; Bootstrap-Verfahren im
Bundle unter `.harness/baseline/v3.5.0/regelwerk/modul-02-harness-bootstrap.md`
§Bootstrap; Quelle/Stand in
[`harness/conventions.md`](harness/conventions.md) §Baseline).

Die **verkörperte Form** (dieses Briefing, die Konventionen, deine ausgefüllten
Artefakte) **führt**; das Regelwerk wird **pro Entscheidung nachgeschlagen,
deren operative Detailtiefe das Briefing nicht trägt** — Trigger-Klassen,
Sub-Area-Qualifikation, Carveout-vs-Reconciliation, Modus-Diagnose. Dabei **nur
den benötigten Abschnitt** laden (README ist der Index), **nicht das ganze
Regelwerk im Kontext halten**. Breiterer Pflicht-Blick bleibt bei: Bootstrap,
Änderung an [`harness/conventions.md`](harness/conventions.md) (Adaptionen
`MR-001..MR-004`, Source-Precedence, ID-Schema), Drift-Audit gegen die Baseline.
Derivativ: bei Konflikt gelten die kanonischen Quellen.

Die **Skelett-Vorlagen** der Baseline liegen **vendored** unter
`.harness/baseline/v3.5.0/templates/` (aus demselben Baseline-Bundle) und tragen
zwei Rollen: als **Referenz-Form**, auf die das Regelwerk mit `../templates/…`
als „Ziel-Form" verweist (netzlos, weil parallel zu `regelwerk/` vendored), und
als Vorlage, die beim Anlegen neuer Artefakte (ADR, Plan/Slice/Welle, Carveout,
Review-Report) **kopiert und ausgefüllt** wird statt frei zu formulieren.

> **Pfad-Hinweis.** Die vendored Templates nutzen das Kanon-Layout
> (`docs/plan/adr/`, `docs/plan/planning/`). m-trace nutzt derzeit
> `docs/adr/` und `docs/planning/` (Adaption MR-001 in
> [`harness/conventions.md`](harness/conventions.md)); die Pfade unten sind
> m-traces reale, aktuelle Pfade. Der Layout-Umzug auf die Kanon-Form ist eine
> spätere, separat abgesicherte Migrations-Welle.

## 2. Kanonische Quellen (Source Precedence)

In dieser Reihenfolge, gemäß [`harness/conventions.md`](harness/conventions.md)
§Spec-Straten (`Contract > Technical > View > ADR > Planning`):

1. [`spec/lastenheft.md`](spec/lastenheft.md) — **Contract**: vertraglich
   abnahmebindende Anforderungen und Akzeptanzkriterien.
2. Technik-Contracts, die den Contract verfeinern:
   [`spec/backend-api-contract.md`](spec/backend-api-contract.md),
   [`spec/browser-support.md`](spec/browser-support.md),
   [`spec/player-sdk.md`](spec/player-sdk.md),
   [`spec/telemetry-model.md`](spec/telemetry-model.md).
3. [`spec/architecture.md`](spec/architecture.md) — abgeleitete Komponenten-,
   Abhängigkeits- und Sequenzsicht.
4. [`docs/adr/`](docs/adr/) — ADR-Verzeichnis und -Index.
5. [`docs/planning/in-progress/roadmap.md`](docs/planning/in-progress/roadmap.md)
   — aktuelle Welle und Lieferstatus.
6. [`README.md`](README.md) — Projekt-Überblick.
7. **AGENTS.md (diese Datei).**
8. [`harness/README.md`](harness/README.md) — Harness-Einstieg.

Verweise zeigen aufwärts (volatil zu stabil). Spec-Dokumente nutzen nie ADR-
oder Planning-Artefakte als normative Quelle; Abwärts-Provenienz bleibt auf
ausgewiesene History-Abschnitte beschränkt.

## 3. Harte Regeln

### 3.1 Docker-only

Kein lokales Toolchain-Install. Alles läuft über `make` (das Docker nutzt);
der Host braucht nur Docker und GNU `make`. Die wenigen Host-seitigen Targets
(bench/mutation) holen ihre Abhängigkeiten über `make host-deps`
(frozen-lockfile), nie über ein bloßes `pnpm install`.

**Falsch:** `go test ./...`, `pnpm --filter ... test`, `docker build …` direkt.
**Richtig:** `make test`, `make gates`.

**Begründung:** Toolchain-Reproduzierbarkeit + Supply-Chain-Defense.

### 3.2 Suppression-Verbot

Inline-Suppression (`//nolint`, `eslint-disable`, ad-hoc CVE-Ignores) bricht die
Gates. Bewusst akzeptierte Vulnerability-Ausnahmen leben zentral in
[`.security/vulnignore.yaml`](.security/vulnignore.yaml) mit Begründung und
verpflichtendem `expires`-Datum — nie einzeln im Code.

### 3.3 git mv + Inhaltsänderung = zwei Commits

Wenn eine Datei verschoben **und** ihr Inhalt umgeschrieben wird:

1. `git mv source target` → eigener Commit (reiner Move; Git erkennt den R-Rename).
2. Inhalt umschreiben → zweiter Commit.

**Begründung:** Sonst fällt die Rename-Detection unter die 50-%-Similarity-
Schwelle und `git log --follow` wird unzuverlässig.

### 3.4 Architektur ist sprach- und meilensteinfrei

[`spec/architecture.md`](spec/architecture.md) referenziert ADRs und
Modul-Pfade, aber **keine** Wellen, Slices, Commit-Hashes oder Closure-Daten.
Die zeitliche Schicht lebt in [`docs/planning/`](docs/planning/) und den
Closure-Notizen.

### 3.5 ADRs sind nach `Accepted` immutable

Eine ADR mit Status `Accepted` wird nicht inhaltlich überschrieben. Korrekturen
entstehen als neue ADR mit `Supersedes ADR-NNNN`. (Die Vor-Adoptions-Records
ADR-0001..0007 sind unter MR-002 grandfathered; neue ADRs erhalten keine
Ausnahme.)

### 3.6 Gates dürfen nicht ohne ADR gelockert werden

Jede Schwellen-Senkung (Coverage, Linter-Strenge, Architekturregel) ist ein
ADR, kein PR-Kommentar.

### 3.7 Variante-B-Cross-Reference-Disziplin

Specs, Releasing-Docs und Makefile-Kommentare stehen als Zielbild.
Cross-Doc-Verweise nutzen **Kennungen** (`F-*`, `NF-*`, `MVP-*`, `AK-*`,
`RAK-*`, `R-*`, `ADR-NNNN`) oder echte Markdown-Links — nie Plan-/Tranche-/
§-Verweise oder Audit-Trail-Stempel. Durchgesetzt von `make lint-variante-b`.

## 4. Quality Gates

Nur Targets, die im `Makefile` existieren, sind gelistet. Die vollständige
Sensor-Liste steht in [`harness/README.md`](harness/README.md) §Sensors.

| Target | Zweck |
|---|---|
| `make test` | Go- + TypeScript-Tests (`api-test` + `ts-test`) |
| `make lint` | Go- + TypeScript-Linter (`api-lint` + `ts-lint`) |
| `make lint-variante-b` | Variante-B-Cross-Reference-Disziplin (§3.7) |
| `make arch-check` | Hexagonal-Architektur-Abhängigkeitsregeln |
| `make coverage-gate` | Go- + TypeScript-Coverage-Schwellen |
| `make docs-check` | Markdown-Referenzen, Spans, tracked Targets, Code-Pfade, Richtung |
| `make verify-closure-notes` | Struktureller Closure-Note-Gate für neue `done/`-Pläne (ADR-0010; standalone, noch nicht in `make gates`) |
| `make gates` | Alle inneren Quality-Gates — mandatory vor einem Pull Request |
| `make security-gates` | `govulncheck` + `pnpm audit` + Trivy-Image-Scan (separater CI-Job, nicht in `make gates`) |
| `make ci` | CI-äquivalent (`gates` + `build`) |
| `make fullbuild` | Volle Closure (`install` + `ci`), vor einem Welle-Merge |

## 5. Dokumentations-Regeln

- Requirement- und ADR-IDs müssen in Pull Requests/Commits referenziert sein,
  nach dem in [`harness/conventions.md`](harness/conventions.md) deklarierten
  ID-Schema (`F-*`, `NF-*`, `MVP-*`, `AK-*`, `RAK-*`, `R-*`; ADR-Nummern über
  den ADR-Index) — nie ad hoc im PR vergeben. Dokumentations-, Test-, Build-,
  CI- und Wartungs-Commits sind exempt.
- Neue ADRs müssen den ADR-Index unter [`docs/adr/`](docs/adr/) aktualisieren.
- Roadmap und Status-Geschichte leben in [`docs/planning/`](docs/planning/),
  nicht in [`spec/architecture.md`](spec/architecture.md).
- Bewusst vertagte Tradeoffs werden als `R-N`-Einträge mit Triggerschwelle in
  [`docs/planning/in-progress/risks-backlog.md`](docs/planning/in-progress/risks-backlog.md)
  getrackt, nicht nur in einem Code-Kommentar.
- Neue (nicht grandfatherte) Pläne in `docs/planning/done/` tragen eine
  **Closure-Note** mit den drei Pflicht-Inhalten aus ADR-0010 (Lernsignal /
  Folge-Slice / Architektur-Beobachtung); Struktur prüft
  `make verify-closure-notes`, Inhalt der closure-note-reviewer-Skill.
- Review-Läufe (Code/Plan/Design) folgen den Skills unter
  [`.harness/skills/`](.harness/skills/)
  (`reviewer.md`, `closure-note-reviewer.md`); Reports landen in
  [`docs/reviews/`](docs/reviews/).
- Quality-Gate-Definitionen leben im `Makefile`; nie ein Gate behaupten, das
  kein ausführbares Target hat.

## 6. Minimal Agent Workflow

Pro Slice:

1. [`harness/README.md`](harness/README.md) lesen.
2. Relevante kanonische Quelle lesen (Source Precedence beachten).
3. Betroffene Requirement-/ADR-IDs identifizieren.
4. Kleinste sinnvolle Änderung planen.
5. Engsten nützlichen Sensor laufen lassen.
6. Vor Handoff den proportionalen Aggregat-Gate laufen lassen — `make docs-check`
   bei reinen Doku-Änderungen, `make gates` bei Code.
7. Doku/Indizes aktualisieren, falls ein öffentlicher Vertrag berührt wird.
8. Ausgeführte Sensors und verbleibende Risiken berichten — keine Erfolgsmeldung
   ohne Gate-Ausführung.
