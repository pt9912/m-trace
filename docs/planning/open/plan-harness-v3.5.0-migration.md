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
| **W4 — Carveout-Mechanismus + risks-backlog-Triage** | `docs/plan/carveouts/` (Index + Template) anlegen; `risks-backlog.md` R-N gegen die Werkzeug-Triade triagieren (Carveout / ADR / Roadmap-Kandidat); neue MR für die bewusst behaltene risks-backlog-Praxis + `plan-<version>`-Slice-Form | W3 done | Triage-Ergebnis dokumentiert, `conventions.md` MR-Liste aktualisiert, Gates grün |
| **W5 — Layout-Move (der Struktur-Umbau)** | Inventar-Slice (repo-weiter Grep, §3) → `docs/adr/` → `docs/plan/adr/`, `docs/planning/` → `docs/plan/planning/`; `.d-check.yml` an **allen 8 Stellen** (inkl. `vcs.paths` + `matrix.exempt-paths`), Nicht-MD-Refs (`.go`/`.ts`-Kommentare, `.gitignore`, `scripts/*.sh`, `lastenheft.md`-Tabelle), MD-Links, `README.md`, Handbuch, „stable links"; **MR-001 auflösen** (nach done-Historie) | W4 done | `make docs-check` + `make gates` + **Grep-Rest-Check** grün auf frischem Klon, `docs-immutable` prüft >0 ADRs, MR-001 als aufgelöst markiert |

**Reihenfolge-Begründung:** W1–W4 sind additiv und netzlos — sie brechen keine
bestehenden Pfade. W5 trägt den gesamten Link-Churn und läuft zuletzt, wenn AGENTS.md
und die Konventionen bereits kanonisch formuliert sind, sodass der Move nur noch die
Verzeichnisse nachzieht.

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

## 4. Out-of-Scope (bewusst nicht in dieser Welle)

- **Slice-Form-Umstellung** `plan-<version>` → `slice-NNN`/`welle-NN`: bleibt
  bewusst release-gebunden (neue MR in W4), keine Umbenennung des Bestands.
- **Aktivierung neuer opt-in-d-check-Module** (`ids`/`citations`/`sources`/`versions`):
  eigener Folge-Slice (Kandidat, siehe §6). Die opt-in-Modul-Frage wurde
  zwischenzeitlich als „R-32" erwogen, **bewusst nicht** in die risks-backlog
  aufgenommen (der Identifier existiert nirgends sonst) — sie ist ein
  Roadmap-Kandidat, kein Risiko.
- **Inhaltliche Neufassung** bestehender Accepted-ADRs (Immutabilität, MR-002).

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
- Ob W5 überhaupt gefahren wird, hängt an ADR-0009: bei unvertretbarem Link-Bruch
  greift ADR-0009 Variante C (Layout-Welle vertagt, MR-001 befristet).

## 6. Folge-Kandidaten (nach Migration)

- **d-check opt-in-Module aktivieren** (`ids` nach Anker-Retrofit laut
  `conventions.md` §Requirement-link convergence; `versions` für die ghcr-Image-Pins;
  `citations`/`sources` nach Bedarf) — Roadmap-Kandidat, wird zum Slice, wenn geschnitten.
- **Freshness-Audit** der vendored Baseline gegen künftige Kurs-Releases (v3.5.0 §Modul 2).

## 7. Closure-Notiz

<!-- Erst nach Abschluss der Migration füllen: was lief, Steering-Loop-Eintrag,
     Folge-Slices. -->
