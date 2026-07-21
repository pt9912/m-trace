# Plan: Harness-Baseline-Migration auf ai-harness-course v3.5.0

> **Status**: Proposed (open) — 2026-07-21. Umsetzung von
> [ADR-0009](../../adr/0009-harness-baseline-v3.5.0.md) (Proposed): strukturelle
> Adoption des v3.5.0-Kanons.
>
> **Bezug**: [ADR-0009](../../adr/0009-harness-baseline-v3.5.0.md);
> [`../../../harness/conventions.md`](../../../harness/conventions.md) (Baseline,
> MR-001..MR-004); [`../in-progress/roadmap.md`](../in-progress/roadmap.md).
> Prozess-Anker RAK-95 (Release-Automatisierung). Kein Lastenheft-Patch (Harness-/
> Prozess-Arbeit, keine User-Surface).

## 1. Ziel

Die commit-gepinnte Grundlagen-nur-Baseline (`d2f60da`, 2026-07-14) durch die
**vendored, integritäts-geprüfte v3.5.0-Baseline** ersetzen und das Repo auf das
kanonische Layout heben, sodass AGENTS.md, die vendored Templates („Ziel-Form")
und die Source-Precedence-Pfade **ohne Divergenz-Steuer** zusammenpassen. MR-001
(Pfad-Divergenz) wird dabei aufgelöst, nicht fortgeschrieben.

**Owner-Entscheidung (2026-07-21): volle v3.5.0-Form.** Es wird die kanonische
**Wellen/Slices-Planung** adoptiert (neue Arbeit als `slice-NNN`/`welle-NN`, **nicht**
weiter `plan-<version>`), und `roadmap.md` wird auf den 5-Abschnitt-Kanon (Aktuelle
Welle · Nächste Wellen · Meilensteine · Abgeschlossene Wellen · Historische
Trigger-Verschiebungen) reformatiert. Die opt-in-d-check-Module werden als konkreter
Slice **umgesetzt** (W7), nicht als Kandidat vertagt.

Zielzustand (Kanon): `docs/plan/adr/`, `docs/plan/planning/{open,next,in-progress,done}/`,
`docs/plan/carveouts/`, `docs/reviews/`, `AGENTS.md`,
`.harness/baseline/v3.5.0/{regelwerk,templates}/` + `SHA256SUMS`,
`.harness/skills/` (Reviewer-/Closure-Note-Skills, aus W3).

## 2. Wellen-Sequenz

Bewusst **additiv/netzlos zuerst, Layout-Move zuletzt** (ADR-0009 Variante C als
Reihenfolge, nicht als Endzustand): so liefert jede Welle einen eigenständigen Wert,
und der risikoreiche Link-Churn ist eine isolierte, gate-abgesicherte letzte Welle.

| Welle | Inhalt | Trigger (Start) | Closure-Trigger |
|---|---|---|---|
| **W1 — Vendored Baseline** | `lab-regelwerk.zip` (sha256 `123e3383…`) nach `.harness/baseline/v3.5.0/{regelwerk,templates}/` + `SHA256SUMS` entpacken; `harness/conventions.md` §Baseline auf vendored v3.5.0 umstellen (Commit-URL ersetzt) | ADR-0009 Accepted | Vendored Bundle committet, SHA256SUMS verifiziert, `make docs-check` grün |
| **W2 — AGENTS.md** | `AGENTS.md` aus vendored Template kopiert-und-ausgefüllt (Source Precedence auf m-trace-Spec-Straten, reale Gate-Tabelle, Hard Rules, 8-Schritt-Workflow); `harness/README.md` gegen die Template-Pflichtgliederung abgeglichen | W1 done | `AGENTS.md` vorhanden, nur reale Make-Targets behauptet, `make docs-check` grün |
| **W3 — docs/reviews/ + next/** | `docs/reviews/` (Review-Report-Template) + `docs/planning/next/` anlegen; Reviewer-/Closure-Note-Skills nach `.harness/skills/` (aus Template) | W2 done | Verzeichnisse + Templates da, Konvention in `AGENTS.md`/`harness/README.md` verlinkt, Gates grün |
| **W4 — Carveout-Mechanismus + risks-backlog-Triage** | `docs/plan/carveouts/` (Index + Template) anlegen; `risks-backlog.md` R-N gegen die Werkzeug-Triade triagieren (Carveout / ADR / Roadmap-Kandidat); ggf. MR nur für nach der Triage verbleibende risks-backlog-Restpraxis | W3 done | Triage-Ergebnis dokumentiert, `conventions.md` MR-Liste aktualisiert, Gates grün |
| **W5 — Layout-Move (der Struktur-Umbau)** | Inventar-Slice (repo-weiter Grep, §3) → `docs/adr/` → `docs/plan/adr/`, `docs/planning/` → `docs/plan/planning/`; `.d-check.yml` an **allen 8 Stellen** (inkl. `vcs.paths` + `matrix.exempt-paths`), Nicht-MD-Refs (`.go`/`.ts`-Kommentare, `.gitignore`, `scripts/*.sh`, `lastenheft.md`-Tabelle), MD-Links, `README.md`, Handbuch, „stable links"; **MR-001 auflösen** (nach done-Historie) | W4 done | `make docs-check` + `make gates` + **Grep-Rest-Check** grün auf frischem Klon, `docs-immutable` prüft >0 ADRs, MR-001 als aufgelöst markiert |
| **W6 — Kanonische Planning-Form** | MR: kanonische `slice-NNN`/`welle-NN`-Form für **neue** Arbeit deklarieren (Owner-entschieden, §1); Umgang mit dem Bestand `plan-<version>.md` — grandfathern (**nur dieser Punkt offen, Owner-Ratifizierung §4**); `roadmap.md` auf den 5-Abschnitt-Kanon reformatieren (im kanonischen Pfad aus W5; bisherige Lieferstand-Tabelle → „Abgeschlossene Wellen", nichts löschen) | W5 done | `roadmap.md` folgt dem Kanon, MR in `conventions.md`, `make gates` grün, kein Lieferstand-Verlust |
| **W7 — opt-in-d-check-Module aktivieren** | zwei Teile: **(7a)** `version.md#aktuell` nach Regelwerk §Versions-Register **anlegen** — es existiert heute **keine** `version.md`; sie konsolidiert die verstreuten Quellen (`packages/*/src/version.ts` `0.25.0`, `DCHECK_DIGEST` in `Makefile`, Go-`serviceVersion` `0.25.0`, `releasing.md`-Bump-Checkliste). **(7b)** Module aktivieren: `versions` (ghcr-Pins gg. `version.md#aktuell`), `ids` (nach Anker-Retrofit, `conventions.md` §Requirement-link convergence), `citations`/`sources` nach Bedarf — je Modul erst advisory-Lauf ohne Befund, dann in `.d-check.yml` + `make gates` binden | W6 done | `version.md` kanonisch (single source), aktivierte Module grün in `make gates`, keine Falschbefunde, DoD je Modul |

**Reihenfolge-Begründung:** W1–W4 sind additiv und netzlos — sie brechen keine
bestehenden Pfade. W5 trägt den gesamten Link-Churn und läuft zuletzt, wenn AGENTS.md
und die Konventionen bereits kanonisch formuliert sind, sodass der Move nur noch die
Verzeichnisse nachzieht. W6/W7 sind **Content-Wellen nach dem Move** (getrennt vom
Pfad-Move, §3-Disziplin): W6 reformatiert `roadmap.md` und etabliert die Wellen/Slices-
Form im schon kanonischen Pfad; W7 legt zuerst das kanonische `version.md` an (Teil 7a)
und aktiviert dann die opt-in-Module (Abhängigkeit nur vom W5-Layout; `version.md`
baut W7 selbst, keine Vorbedingung aus W6).

## 3. Slice-Schnitt (W5, der kritische)

**Vorab — Inventar-Slice (W5.0).** `make docs-check` (scan.roots `['.']`) prüft nur
**Markdown-Links** — Nicht-Markdown-Referenzen auf `docs/adr/`/`docs/planning/` sind
für das Gate **unsichtbar** und driften still (Beleg: `.gitignore` zeigt schon heute
auf das nach `done/` gewanderte `plan-spike.md`). Vor dem ersten Move ein repo-weiter
Grep als **committete Arbeitsliste** über alle Nicht-MD-Stellen:

- **~12 `.go`/`.ts`-Kommentare** (Layer-Regel-Verweise, z. B.
  `apps/api/hexagon/domain/playback_event.go` → `plan-spike.md`),
- **`.gitignore`** (Spike-Pfade — die stale Zeilen gleich mitfixen),
- **`scripts/*.sh`** (z. B. `open-security-audit-issue.sh` → `risks-backlog.md`),
- **`spec/lastenheft.md`** — ein `roadmap.md`-**Code-Span in einer Tabelle** (kein
  MD-Link → `links`-Modul greift nicht) **im normativen Vertrag**.

Der Move wird nach **Verzeichnis-Familie** geschnitten, jeder Slice einzeln gate-grün.
Der **Move-Commit ist *nicht* „nur `git mv`"**: ein reiner Verzeichnis-Move ließe die
Tooling-Config auf verschwundene Pfade zeigen → Gates **rot**. Er umfasst daher
`git mv` **plus** das Repathing der Tooling-Config (andere Dateien, **kein**
ADR-/Slice-*Body*-Edit → Content-Immutabilität *und* git-Rename-Detection bleiben
gewahrt); der **zweite Commit** trägt die Cross-Doc-Link-Fixes.

1. **`docs/adr/` → `docs/plan/adr/`** (git mv, reiner Rename der ADR-Dateien) +
   `.d-check.yml` an **allen** ADR-Stellen: `trace.adrs.dir`, `matrix` adr-`paths`,
   **`matrix.exempt-paths`** (die 7-ADR-Grandfather-Liste), **`vcs.paths`** (der
   `docs-immutable`-Sensor) und `codepaths.scope.roots`. **`vcs.paths` zwingend im
   selben Slice** — sonst scannt der Immutabilitäts-Sensor 0 Dateien und **passt
   vakuum** (die Garantie verschwindet still, genau beim Move, der ADR-Inhalt am
   ehesten anfasst). Plus ADR-Index.
2. **`docs/planning/` → `docs/plan/planning/`** (git mv) + `.d-check.yml`
   `trace.slices.dir` und `matrix` planning-`paths`; Planning-interne Links; `next/`
   ist aus W3 da.
3. **Nicht-MD-Referenzen** aus der Inventar-Arbeitsliste: `.go`/`.ts`-Kommentare,
   `.gitignore`, `scripts/*.sh`, der `lastenheft.md`-Tabellen-Code-Span.
4. **Root-/externe MD-Verweise:** `README.md`, `docs/user/*`, veröffentlichte
   „stable links".
5. **MR-001 auflösen** + `harness/conventions.md`/`harness/README.md` Source-Precedence
   auf die kanonischen Pfade.

## 4. Out-of-Scope / offene Owner-Entscheidung

> **Provenienz-Hinweis.** Frühere Fassungen dieses Plans führten „Slice-Form bleibt
> `plan-<version>`" und „opt-in-Module bleiben Kandidat" als „bewusst" out-of-scope —
> das war ein Scoping-Vorgriff des Autors ohne Owner-Beschluss. Owner-Entscheidung
> 2026-07-21 (§1): beides ist In-Scope (W6/W7). Es bleibt nur eine rule-backed
> Grenze und eine echte offene Entscheidung.

- **Inhaltliche Neufassung** bestehender Accepted-ADRs bleibt ausgeschlossen —
  gebunden durch die Hard Rule „ADRs immutabel nach `Accepted`" (Regelwerk Modul 4)
  **plus MR-002** (Owner-erfasst 2026-07-14). Der W5-Move betrifft nur den **Pfad**,
  nicht den Inhalt.
- **Offen (Owner-Entscheidung, W6):** Behandlung des **Bestands** `plan-<version>.md`
  (Dutzende Dateien in `done/`). *Vorschlag des Autors:* grandfathern (historische
  Records bleiben, nur neue Arbeit nutzt `slice-NNN`/`welle-NN`) — regelwerk-konsistent
  (Brownfield-Grandfathering wie MR-002), kein Massen-Churn, Release-Versions-Kopplung
  bleibt. *Alternative:* Vollumbenennung des Bestands. **Nicht vorentschieden.**

## 5. Risiken und offene Punkte

- **`make docs-check` ist Markdown-blind.** Es prüft nur MD-Links, **nicht**
  `.go`/`.ts`-Kommentare, `.gitignore`, Shell-Skripte oder Tabellen-Code-Spans →
  die W5-Closure braucht **zusätzlich** einen repo-weiten Grep-Rest-Check (Inventar
  §3), nicht nur `docs-check`. Beleg für die stille Drift: die bereits stale
  `.gitignore`-Zeile auf `plan-spike.md`.
- **`vcs.paths`-Vakuum.** Wird `vcs.paths` (`docs-immutable`) nicht im ADR-Move-Slice
  mitgepfadet, scannt der Sensor 0 Dateien und passt still — die Immutabilitäts-
  Garantie fällt weg, genau beim riskantesten Move. Zwingend im selben Slice (§3.1);
  Closure prüft „`docs-immutable` sieht >0 ADRs".
- **Content-Immutabilität** der Accepted-ADRs: Move-Commit = `git mv` +
  Tooling-Config-Repath, **kein** ADR-Body-Edit (§3) → Rename-Detection + Immutabilität
  gewahrt.
- **Toter externer Link** durch W5-Move (Grund für MR-001). Gegenmittel: W5 zuletzt,
  Grep + `docs-check` pro Slice, Redirect-Hinweis im `README.md` erwägen.
- **`roadmap.md`-Reformat (W6)** darf keinen Lieferstand verlieren: die bisherige
  §1.1-Lieferstand-Tabelle wird in „Abgeschlossene Wellen" überführt, nicht gelöscht;
  der 5-Abschnitt-Kanon trägt die Historie weiter.
- **`version.md` fehlt (W7a):** `versions` prüft gegen `version.md#aktuell`, das es
  heute **nicht** gibt und das keine Vor-Welle anlegt → W7 baut es zuerst selbst
  (konsolidiert die 4 verstreuten Versions-Quellen, §2 W7). Ohne den Schritt hätte
  `versions` kein Ziel.
- **W7-Falschbefunde:** `ids` würde heute 340 nackte Kennungen auf den Dateianfang
  linken (`conventions.md` §Requirement-link convergence) — Anker-Retrofit ist
  Vorbedingung, sonst bleibt `ids` aus. Je Modul erst advisory-Lauf ohne Befund.
- Ob W5 überhaupt gefahren wird, hängt an ADR-0009: bei unvertretbarem Link-Bruch
  greift ADR-0009 Variante C (Layout-Welle vertagt, MR-001 befristet).

## 6. Folge-Kandidaten (nach Migration)

- **Freshness-Audit** der vendored Baseline gegen künftige Kurs-Releases (v3.5.0
  §Modul 2): beobachtbarer Auslöser (neuer Kurs-Tag) → Re-Vendoring-Review, kein
  Auto-Bump.
- Sollte der Owner die **Vollumbenennung** des `plan-<version>`-Bestands (§4) wählen,
  ist das ein eigener, großer Slice nach W6.

## 7. Closure-Notiz

<!-- Erst nach Abschluss der Migration füllen: was lief, Steering-Loop-Eintrag,
     Folge-Slices. -->
