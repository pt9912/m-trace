# 0009 — Regelwerk-Baseline auf ai-harness-course v3.5.0: strukturelle Adoption

> **Status**: **Proposed** (2026-07-21)
> **Datum**: 2026-07-21 (Proposed)
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)
> **Bezug**: [`harness/conventions.md`](../../harness/conventions.md) §Baseline / §Adaptations
> (MR-001..MR-004); [`harness/README.md`](../../harness/README.md) §Source
> precedence; RAK-95 (Release-Automatisierung, normativ im Lastenheft) als
> Prozess-Anker. Kein neuer normativer Lastenheft-Bezug — dies ist ein
> **Prozess-/Harness-ADR** ohne Spec-Stratum-Schärfung. (Das aus der
> v3.5.0-Vorlage stammende `Schärft:`-Kopffeld wird erst mit dem
> Template-Landing in W1/W2 eingeführt und hier noch nicht geführt.)

## Kontext

m-trace hat den ai-harness-Harness **teilweise** adoptiert: `harness/README.md`,
`harness/conventions.md`, die Spec-Stratifizierung (Lastenheft +
vier Technik-Contracts + Architektur-Sicht), die ADR-Reihe unter `docs/adr/`,
den Planning-Lifecycle unter `docs/planning/` und Wellen in der Roadmap sind da.

Die adoptierte **Baseline ist jedoch commit-gepinnt** auf `grundlagen-konventionen.md`
@ `d2f60da` (abgerufen 2026-07-14) — **nur die eine Grundlagen-Datei**, nicht das
ganze Regelwerk, und **ohne Templates**. Seither ist ai-harness-course **v3.5.0**
(2026-07-19) als self-contained `lab-regelwerk.zip` erschienen (sha256
`123e3383261102e6be6465e1f4bade08a474c00edc4fff89f5c4b11bd640f8ff`): 17 Module +
3 Grundlagen + der komplette Template-Satz.

v3.5.0 macht drei Dinge normativ, die m-traces Teil-Adoption nicht abbildet:

1. **Vendored Baseline.** Regelwerk *und* Templates werden committet unter
   .harness/baseline/&lt;tag&gt;/{regelwerk,templates}/ + SHA256SUMS geführt
   (netzlos, integritäts-geprüft). Die Templates sind die **Referenz-„Ziel-Form"**,
   auf die die `../templates/…`-Verweise des Regelwerks netzlos auflösen (weil
   `regelwerk/` und `templates/` parallel liegen). m-trace verlinkt stattdessen nur
   eine Commit-URL in [`harness/conventions.md`](../../harness/conventions.md) —
   die Ziel-Formen lösen nirgends auf, und der Pin driftet still bei jedem
   Kurs-Release.
2. **AGENTS.md als primärer Agent-Einstieg.** m-trace hat **keine**. Im v3.5.0-Kanon
   ist AGENTS.md das Onboarding-Briefing (Source Precedence, Hard Rules, Quality-Gate-
   Tabelle, 8-Schritt-Workflow); das Regelwerk-README benennt sie ausdrücklich als
   Einstieg des Adopter-Repos.
3. **Kanonisches Verzeichnis-Layout.** Der Kanon ist docs/plan/adr/,
   docs/plan/planning/{open,next,in-progress,done}/, docs/plan/carveouts/ und
   docs/reviews/. Die **Templates *und* die AGENTS-Vorlage hartkodieren** diese
   Pfade in Source-Precedence, Hard Rule 3.4 („die zeitliche Schicht lebt in
   docs/plan/planning/") und in allen `../templates/…`-Cross-Links.

**Der entscheidende Punkt.** MR-001 (`docs/adr/` statt docs/plan/adr/,
`docs/planning/` statt docs/plan/planning/) war gegen die alte *Grundlagen-nur*-
Baseline eine saubere Ein-Zeilen-Adaption. Gegen v3.5.0 — mit vendored Templates
und einer AGENTS.md, deren Inhalt durchgehend docs/plan/… referenziert — wird
MR-001 zu einer **fortlaufenden Divergenz-Steuer**: jede aus einem Template
kopierte Gründungs-Datei (AGENTS.md, `harness/README.md` §Source precedence) und
jeder „Ziel-Form"-Verweis müsste um-gepfadet werden, und ein Leser, der der
Referenz-Form folgt, bekäme kanonische Pfade, die das Repo nicht hat. Die
v3.5.0-Adoption ist damit **kein reiner Versions-Bump, sondern ein Struktur-Umbau.**

Zusätzlich weichen zwei m-trace-Praktiken vom Kanon ab:

- **Slice-Form:** m-trace nutzt release-gebundene `plan-<version>.md` statt
  `slice-NNN`/`welle-NN`. **Owner-Entscheidung (2026-07-21): auf die kanonische
  Wellen/Slices-Form migrieren** (neue Arbeit `slice-NNN`/`welle-NN`, Bestand
  grandfathered — der Umgang mit dem Bestand ist offene Owner-Ratifizierung, Plan §4);
  `roadmap.md` wird auf den v3.5.0-Kanon (5 Abschnitte) reformatiert.
- **Vertagte Arbeit:** m-trace führt `docs/planning/in-progress/risks-backlog.md`
  (R-N mit Triggern) — das Regelwerk kennt kein „risks-backlog", sondern Carveouts
  (Gate-Ausnahmen), BF-Sub-Area-Markierungen und Roadmap-Kandidaten. Die R-N werden
  gegen die Werkzeug-Triade triagiert (Plan W4).

## Entscheidung

> **Entscheidung (Proposed 2026-07-21):** **Strukturelle Adoption von
> ai-harness-course v3.5.0.** m-trace übernimmt das kanonische Layout und den
> vendored-Baseline-Mechanismus, statt die Pfad-Divergenz als permanente Adaption
> fortzuschreiben.

Bestandteile:

1. **Baseline vendoren:** Regelwerk + Templates aus `lab-regelwerk.zip` (sha256
   `123e3383…`) nach .harness/baseline/v3.5.0/{regelwerk,templates}/ + SHA256SUMS
   entpacken (netzlos, committet). [`harness/conventions.md`](../../harness/conventions.md)
   §Baseline zeigt künftig auf diesen vendored Stand statt auf die Commit-URL.
2. **Layout auf Kanon migrieren:** `docs/adr/` → docs/plan/adr/,
   `docs/planning/` → docs/plan/planning/ (inkl. neuem next/), neu
   docs/plan/carveouts/ und docs/reviews/. Umzug per reinem `git mv`
   (Rename-Detection), Inhalts-/Link-Anpassung als zweiter Commit (AGENTS-Hard-Rule
   „git mv + Inhaltsänderung = zwei Commits").
3. **AGENTS.md anlegen** (aus Template kopiert-und-ausgefüllt): Source Precedence
   auf m-traces Spec-Straten, Hard Rules (Docker-only, Suppression-Verbot,
   git-mv-Zweischritt, Architektur meilensteinfrei, ADR-Immutabilität, Gate-Senkung
   nur per ADR), reale Gate-Tabelle (keine halluzinierten Gates), 8-Schritt-Workflow.
4. **conventions.md + Planungs-Form nachziehen:** **MR-001 zurückziehen** (Layout ist
   danach kanonisch → Eintrag nach done-Historie mit „aufgelöst"); MR-002/003/004
   bleiben. **Kanonische Wellen/Slices-Form** (Owner-Entscheidung 2026-07-21): neue
   Arbeit als `slice-NNN`/`welle-NN`, Bestand `plan-<version>` grandfathered (Umgang
   mit dem Bestand: offene Owner-Ratifizierung, Plan §4); **frische, altlastenfreie
   `roadmap.md`** im v3.5.0-Kanon anlegen und den historienlastigen Alt-Stand nach
   `done/` archivieren (Owner-Entscheidung 2026-07-21). Ggf. neue MR nur für nach der
   Triage verbleibende risks-backlog-Restpraxis.
5. **risks-backlog triagieren** gegen die Werkzeug-Triade: echte Gate-Ausnahmen →
   Carveout (docs/plan/carveouts/CO-NNN); Architektur-Dauerentscheidungen → ADR;
   Rest (Risiko-/Folge-Trigger) → Roadmap-Kandidat/Slice bzw. als deklarierte
   m-trace-Adaption belassen.
6. **Tooling nachziehen:** `.d-check.yml` (trace-`slices`-Dir, matrix-Klassen-Pfade,
   codepaths-scope), `harness/README.md`, `Makefile`/`d-check.mk`-Kommentare,
   `README.md` und veröffentlichte „stable links" auf die neuen Pfade.

Die Ausführung ist **kein Ein-Schritt-Commit**, sondern eine sequenzierte
Migrations-Welle — Sequenz, Wellen und Slice-Schnitt stehen im zugehörigen
Migrationsplan (in der Roadmap §1.2 gelistet). Die Referenz-Richtung bleibt
**Plan → ADR** (SDP): der Plan verweist aufwärts auf dieses ADR, ein
ADR → Planning-Link ist `matrix-forbidden` und daher bewusst weggelassen —
die Discovery läuft über die Roadmap. Der Link-Churn (genau der Grund für
MR-001) wird kontrolliert und gate-grün abgearbeitet.

## Konsequenzen

**Positiv:**

- Kanonisches Layout; AGENTS.md als deklarierter Agent-Einstieg; netzlos vendored,
  integritäts-geprüfte Baseline (reproduzierbar über den `v3.5.0`-Tag).
- Die `../templates/…`-„Ziel-Form"-Verweise des Regelwerks lösen im Repo auf;
  neue Artefakte (ADR, Slice/Welle, Carveout, Review-Report) werden aus der vendored
  Referenz-Form kopiert statt frei formuliert.
- Die MR-Liste schrumpft um die schwerste Divergenz (Pfade) und dokumentiert die
  verbleibenden bewussten Abweichungen sauber.
- v3.5.0-Werkzeuge werden anschlussfähig: `docs/reviews/`, Wellen-Closure-Prozedur,
  Carveout-Disziplin, und die neuen d-check-Module (citations/sources/versions/ids)
  aus [ADR-0008](0008-benchmark-mutation-execution-in-docker.md)-Nachbarschaft
  (d-check-Pin) stehen als Folge-Schritte bereit.

**Kosten / Grenzen (ehrlich benannt):**

- **Großer Link-Churn.** MR-001 existierte wegen „etabliertem öffentlichem Layout mit
  umfangreichen stabilen Links". Der Umzug berührt alle internen Verweise auf
  `docs/adr/`/`docs/planning/`, die `.d-check.yml`-Pfade, README/Handbuch und ggf.
  extern zitierte URLs. Das ist die eigentliche Umbau-Arbeit und trägt Regressions-
  Risiko (tote Links) — abgesichert durch `make docs-check` nach jedem Schritt.
- **Accepted-ADRs sind immutabel.** Der Move betrifft **Pfad, nicht Inhalt**:
  ADR-0001..0008 wandern per `git mv`, ihr Inhalt bleibt unberührt; der
  Immutabilitäts-Sensor (`make docs-immutable`) darf durch den reinen Move nicht
  anschlagen — Move-Commit ohne Inhaltsänderung. Davon zu trennen ist das
  **inhaltliche** MR-002-Grandfathering: es gilt nur für die historisch
  abweichenden 0001–0007; ADR-0008 wurde bereits konform verfasst und braucht es
  nicht. Der Pfad-Move gilt für alle acht.
- **Zwischenzustände.** Während der Welle sind Pfade teils alt/teils neu; die Gates
  müssen pro Schritt grün bleiben (kein „großer Bang"-Commit).
- **Erweiterter Scope = mehr Arbeit (Owner-Entscheidung 2026-07-21).** Die Umstellung
  der Planungs-Form auf `slice-NNN`/`welle-NN` samt `roadmap.md`-Reformat (Plan W6) und
  die Aktivierung der opt-in-d-check-Module (Plan W7) sind **Teil** der Migration —
  zwei zusätzliche Content-Wellen nach dem Layout-Move, nicht vertagte
  Folge-Entscheidungen. Einzige offene Owner-Ratifizierung: Umgang mit dem
  `plan-<version>`-Bestand (grandfathern vs. umbenennen, Plan §4).

## Alternativen

- **B — Reiner Baseline-Versions-Bump, MR-001 behalten.** Baseline-Zeile auf `v3.5.0`
  heben, Layout lassen. **Verworfen:** perpetuiert die Divergenz-Steuer — AGENTS.md und
  `harness/README.md` müssten dauerhaft um-gepfadet werden, und die vendored
  „Ziel-Form"-Templates zeigten auf Pfade, die das Repo nicht hat. (Dies war die
  zunächst vertretene, dann verworfene Einschätzung.)
- **C — Hybrid: AGENTS.md + vendored Baseline + carveouts + docs/reviews jetzt,
  Layout-Move als separat reviewter Folge-Schritt (MR-001 befristet halten).**
  Plausibel als **Sequenzierung innerhalb** der Adoption (frühe Wellen liefern das
  Netzlose/Additive, die Layout-Welle folgt), **nicht** als Endzustand. Im
  Migrationsplan als Wellen-Reihenfolge aufgegriffen; der Zielzustand bleibt die
  volle strukturelle Adoption dieses ADR.
- **D — Nicht migrieren, bei Commit-Pin bleiben.** **Verworfen:** der Pin driftet
  still gegen Kurs-Releases, die Ziel-Formen lösen nie auf, und reale v3.5.0-Werkzeuge
  (vendored-Baseline-Doktrin, Carveout-/Wellen-Prozedur, neue d-check-Module) bleiben
  ungenutzt.

## Re-Evaluierungs-Trigger

- Ein Kurs-Release nach v3.5.0 mit strukturellen Template-/Layout-Änderungen (dann
  Baseline-Re-Vendoring als eigener Review, kein stiller Auto-Bump).
- Sollte der Layout-Umzug (Bestandteil 2) einen unvertretbaren externen
  Link-Bruch verursachen, fällt die Entscheidung auf Variante C zurück (Layout-Welle
  vertagt, MR-001 befristet reaktiviert) — dokumentiert als Folge-ADR.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-07-21 | Proposed | ADR-0009 |
