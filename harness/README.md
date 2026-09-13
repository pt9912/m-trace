# Harness

## Purpose

Dieser Harness verbindet bestehende Spezifikationen, ADRs, Planning-Dokumente
und Gates. Er ist **kein Ersatz** für `spec/` oder `docs/`, sondern ein
**Einstiegspunkt** für Menschen und AI-Code-Agenten.

Wenn diese Datei einer kanonischen Quelle widerspricht, **gewinnt die
kanonische Quelle**, und diese Datei wird angepasst.

Strukturregeln (ID-Schemata, Modus-Deklarationen pro Sub-Area, Zusatzklassen
für Sensors-Bindung) sowie Adaptionen ggü. der adoptierten Baseline leben in
[`conventions.md`](conventions.md). Diese Datei dupliziert sie nicht.

## Source precedence

| Rang | Datei | Charakter |
|---|---|---|
| 1 | [`spec/lastenheft.md`](../spec/lastenheft.md) | vertraglich abnahmebindend |
| 2 | [`spec/backend-api-contract.md`](../spec/backend-api-contract.md), [`spec/browser-support.md`](../spec/browser-support.md), [`spec/player-sdk.md`](../spec/player-sdk.md), [`spec/telemetry-model.md`](../spec/telemetry-model.md) | technisch fortschreibbar |
| 3 | [`spec/architecture.md`](../spec/architecture.md) | Komponenten/Sequenzen, meilensteinfrei |
| 4 | [`docs/plan/adr/`](../docs/plan/adr/) | Architekturentscheidungen |
| 5 | [`docs/plan/planning/in-progress/roadmap.md`](../docs/plan/planning/in-progress/roadmap.md) | Wellen-Sequenz |
| 6 | [`docs/user/`](../docs/user/) | Anwenderhandbuch, Betrieb, Qualität, Releasing |
| 7 | [`README.md`](../README.md) | Projekt-Überblick |
| 8 | [`AGENTS.md`](../AGENTS.md) | Agent-Briefing |
| 9 | diese Datei | Harness-Einstieg |

Rang 2 führt vier Dateien statt einer `spec/spezifikation.md` — deklariert
als [MR-008](conventions.md#mr-008).

## Guides (Feedforward-Quellen)

| Quelle | Inhalt |
|---|---|
| [`spec/lastenheft.md`](../spec/lastenheft.md) | Anforderungen, IDs, Akzeptanzkriterien |
| [`spec/backend-api-contract.md`](../spec/backend-api-contract.md), [`spec/browser-support.md`](../spec/browser-support.md), [`spec/player-sdk.md`](../spec/player-sdk.md), [`spec/telemetry-model.md`](../spec/telemetry-model.md) | technische Details, Defaults |
| [`spec/architecture.md`](../spec/architecture.md) | Komponenten, Schichten, Constraints |
| [`docs/plan/adr/`](../docs/plan/adr/) | Architekturentscheidungen |
| [`docs/plan/planning/`](../docs/plan/planning/) | Slice-Pläne und Roadmap |
| [`docs/plan/planning/observations/`](../docs/plan/planning/observations/) | Beobachtungs-Register: wiederkehrende Diskrepanz-/Konventions-Funde je Sub-Area (`BEO-<KUERZEL>/<slug>`), Steering-Loop-Zähler |
| [`AGENTS.md`](../AGENTS.md) | Hard Rules, Source Precedence, Workflow |
| [`conventions.md`](conventions.md) | ID-Schemata, Adaptions-Block (`MR-*`), Modus-Deklarationen |
| [`.harness/skills/reviewer.md`](../.harness/skills/reviewer.md) | Reviewer-Skill: HIGH-Liste, Kategorien-Regeln, Negativbefund-Pflicht, Output-Schema — nächste Rolle nach Schritt 8 des Minimal Agent Workflow, nicht Teil der Implementer-Eingabe |
| `.harness/baseline/<tag>/regelwerk/` (vendored; `README.md` = Index) | adoptiertes Betriebsregelwerk in Agenten-Kurzform — präsente nachschlagbare Vertiefung, pro Entscheidung abschnittsweise (siehe [`AGENTS.md`](../AGENTS.md) §1); Stand/Tag siehe [`conventions.md`](conventions.md) §Baseline |
| `.harness/baseline/<tag>/templates/` (vendored, parallel) | Referenz-Form der Skelette, auf die das Regelwerk mit `../templates/…` als „Ziel-Form" verweist; Vorlagen zum Kopieren-und-Ausfüllen |

## Sensors (Feedback-Gates)

**DIES IST DER EINZIGE GATE-INDEX.** Kommt ein Target dazu, wird es hier
eingetragen, nirgends sonst; `AGENTS.md` §4 trägt die Regel und den Zeiger
hierher.

**Referenz-Regel für einfrierende Artefakte** (Review-Reports,
Closure-Notizen, Accepted-ADRs, geschlossene Slices — ab hier auch
unsere eigenen): Ein Gate wird über sein `make <target>`-Token zitiert,
**nicht** über den Pfad zu seiner Sensor-Datei — die Datei kann
verschwinden oder wandern, das Token bleibt die stabile Adresse. Gilt
nicht rückwirkend für bestehende Closure-Notizen (`slice-009`…`012`).

| Target | Vertrag | Bindung |
|---|---|---|
| `make test` | Go- und TypeScript-Tests (`api-test`/`api-race` + `ts-test`) | — |
| `make lint` | Go- und TypeScript-Linter (`api-lint` + `ts-lint`) | — |
| `make lint-variante-b` | Variante-B-Cross-Reference-Disziplin (AGENTS.md §3.8) | — |
| `make arch-check` | Hexagonal-Architektur-Abhängigkeitsregeln (`a-check` via `.a-check.yml`) | — |
| `make coverage-gate` | Go- und TypeScript-Coverage-Schwellen | Schwelle 90 % |
| `make docs-check` | Markdown-Referenzen, Spans, tracked Targets, Code-/Host-Pfade, ID-Verankerung, Planning-Lifecycle (d-check, Module s. `.d-check.yml`) | — |
| `make docs-immutable` | Accepted-ADR-Kern gegen den gestagten Diff (Aufruf: `STAGED=1`) | — |
| `make docs-commits` | Commit-Message-Traceability über einen Pull-Request-Bereich (Aufruf: `RANGE=base..head`) | — |
| `make verify-closure-notes` | Struktureller Closure-Note-Gate für neue `done/`-Pläne | ADR-0010 |
| `make build` | Baubare Release-Artefakte (`api-build` + `ts-build`) | — |
| [`make gates`](sensors/gates.md) | Alle inneren Quality-Gates gebündelt; Grenze und Zusammensetzung in der verlinkten Datei | — |
| `make security-gates` | `govulncheck` + `pnpm audit` + Trivy-Image-Scan + Image-Start-Smoke (separater CI-Job, nicht in `make gates`) | — |
| `make ci` | CI-äquivalent (`gates` + `build`) | — |
| `make fullbuild` | Volle Closure (`install` + `ci`), vor einem Welle-Merge | — |

**Werkzeuge — genannt, weil der Lauf sie braucht, aber kein Gate:**

| Target | Tut was | Bindung |
|---|---|---|
| `make doc-trace` | druckt die Advisory-Requirements-Matrix aus den nativen Lastenheft-Tabellen und Planning-Referenzen auf stdout | kein Gate |
| `make coverage-report` | druckt den Coverage-Report ohne Schwellen-Urteil | kein Gate |
| `make host-deps` | installiert lokale `node_modules` für Nicht-Docker-Targets (frozen-lockfile) | kein Gate |

**Aktueller Lauf-Status:** CI-Badge bzw. lokal `make help` / `make gates`.
**Rote Gates:** Begründung im verlinkten `docs/plan/carveouts/CO-<NNN>` —
aktuell keine aktiven generischen Carveouts (siehe
[`docs/plan/carveouts/README.md`](../docs/plan/carveouts/README.md); der
domänenspezifische Security-Suppression-Cluster läuft separat über
[`.security/vulnignore.yaml`](../.security/vulnignore.yaml)).

## Traceability rules

Requirements nutzen die bestehenden Familien `F-*`, `NF-*`, `MVP-*`, `AK-*`,
`RAK-*` und `R-*`. Architektur-Entscheidungen nutzen `ADR-NNNN`. Neue normative
Verweise zeigen von volatilen zu stabilen Quellen; Abwärts-Provenienz bleibt auf
ausgewiesene History-Abschnitte beschränkt.

Die Commit-Message-Durchsetzung gilt für Pull-Request-Bereiche. Dokumentations-,
Test-, Build-, CI- und Wartungs-Commits sind exempt; Feature- und Fix-Commits
tragen eine Requirement-, Entscheidungs- oder Plan-Kennung.

## Safety and scope boundaries

- Niemals einen höherrangigen Contract abschwächen, um einer Implementierungs-
  Drift zu entsprechen.
- Niemals ein Gate behaupten, das kein ausführbares Target hat.
- Bestehende akzeptierte ADRs sind historische Records; Änderungen erfordern den
  dokumentierten Entscheidungsprozess.

## Minimal agent workflow

1. Die höchstrangige für die Aufgabe relevante Quelle lesen.
2. Bestehenden Code und Tests prüfen, bevor Verhalten geändert wird.
3. Die Änderung an ein bestehendes Requirement, eine Entscheidung, einen Test
   oder ein Gate binden.
4. Die kleinste kohärente Änderung umsetzen.
5. Fokussierte Tests laufen lassen.
6. `make docs-check` bei Dokumentations-Änderungen laufen lassen.
7. Den proportionalen Aggregat-Gate laufen lassen.
8. Niederrangigen Status oder Anwender-Dokumentation aktualisieren, ohne den
   höherrangigen Contract implizit zu ändern.

Dieser Workflow deckt ausschließlich die Implementer-Rolle ab. Schritt 8 ist
der Rollenwechsel, kein Abschluss: Bericht → Handoff an Reviewer
([`.harness/skills/reviewer.md`](../.harness/skills/reviewer.md), siehe
§Guides) → Verifier. Kein Self-Review — anderer Kontext findet andere
Findings, derselbe Kontext dieselben blinden Flecken.

## Leseordnung

1. [`AGENTS.md`](../AGENTS.md) §Hard Rules — Pflichtregeln vor jeder Änderung.
2. [`spec/lastenheft.md`](../spec/lastenheft.md) — Contract, was m-trace
   bereitstellen muss.
3. [`conventions.md`](conventions.md) bei Bedarf — ID-Schemata, Adaptionen
   (`MR-*`), Modi.
