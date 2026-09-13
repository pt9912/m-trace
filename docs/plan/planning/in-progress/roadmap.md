# Roadmap

**Format-Regel:** Diese Roadmap ist eine Reihenfolge von **Wellen**, keine
Reihenfolge von Terminen (v3.5.0-Regelwerk Modul 6). Ein Trigger ist eine
*beobachtbare Bedingung* (nicht ein Datum); Termine erscheinen höchstens als
Schätzung, treiben aber nie eine Welle.

> **Historie vor v3.5.0.** Die vollständige Release- und Entscheidungshistorie
> bis `0.25.0` (25 Release-/Patch-Pläne, Trigger-Re-Evals, Lessons-learned) steht
> im Archiv [`../done/roadmap-pre-v3.5.0.md`](../done/roadmap-pre-v3.5.0.md) und
> in den einzelnen [`done/plan-*.md`](../done/)-Records. Diese Datei ist bewusst
> altlastenfrei und forward-looking.

---

## Offene Wellen

Geöffnet ist [`welle-02`](../welle-02-regelwerk-v6.8.0-migration.md)
(Harness-Regelwerk-Baseline v3.5.1 → v6.8.0, elf Tranchen, additiv zuerst).
Aktuell `in-progress`:
[`slice-016`](slice-016-roadmap-offene-wellen-format.md) (Tranche 8,
dieses Roadmap-Format-Update selbst). Fünf weitere Tranchen sind geschnitten
und warten in [`open/`](../open/) (`slice-012`–`015`); vier hängen an
Owner-Entscheidungen (Welle-Datei §8) oder Abhängigkeiten zu anderen
Tranchen. Kein Produktcode betroffen, reines Harness/Prozess-Territorium,
läuft unabhängig von einer künftigen Produkt-Folgewelle (die weiterhin
**nicht geschnitten** ist, siehe *Nächste Wellen*).

## Slices ohne Welle

Wartung/Architektur, Modul 5 „ohne Welle", alle in [`done/`](../done/), seit
Abschluss der ersten Welle: [`slice-003`](../done/slice-003-a-check-arch-gate.md)
(a-check ersetzt `check-architecture.sh`),
[`slice-004`](../done/slice-004-driving-port-boundary.md) (Driving-Port-Boundary
säubern), [`slice-005`](../done/slice-005-closure-gate-slice-welle.md)
(Closure-Gate greift auf slice/welle), [`slice-006`](../done/slice-006-planning-layout-nicht-slices-flach.md)
(Planning-Layout: Nicht-Slices flach aus `in-progress/`),
[`slice-007`](../done/slice-007-review-report-praxis.md) (Review-Report-Praxis
scharf geschaltet — erster echter Handoff-Report + normative Regel),
[`slice-008`](../done/slice-008-baseline-v3.5.1-bump.md) (Regelwerk-Baseline
v3.5.0 → v3.5.1, nicht-struktureller Re-Vendor — [ADR-0011](../../adr/0011-harness-baseline-v3.5.1-bump.md), Accepted),
[`slice-009`](../done/slice-009-image-start-gate.md) (Runtime-Images im Gate
starten, nicht nur bauen und scannen), [`slice-010`](../done/slice-010-dcheck-planning-waves.md)
(d-check `planning.waves` — Wellen-Register-Invariante) und
[`slice-011`](../done/slice-011-dcheck-links-resolve-from.md) (d-check
`links.resolve-from` — ortsfeste Verweise im Planning-Lifecycle).

## Nächste Wellen

| Welle | Trigger (beobachtbar) | Wichtigste Slices | Aufwand |
|---|---|---|---|
| Produkt-Folgewelle (noch **ungeschnitten**) | Migration done **und** Owner schneidet Tranche | Kandidaten aus dem Risiko-Register: `R-13` (perl-freies Runtime-Base für Dashboard/Analyzer — **einziger Kandidat mit Frist**: am `2026-11-02` laufen beide Suppression-Cluster ab und `make security-gates` bricht am `expires`-Check), `R-30` (SSE-Backfill-Skip Multi-Replica), `R-24` (Load-Smoke-Debounce), policy-getriebene Per-Projekt-Limiter-Buckets ([`RAK-74`](../../../../spec/lastenheft.md#rak-74)-Anschluss), Redis-Cluster-Tauglichkeit der Lua-Limiter, Durchsatz jenseits Single-Postgres (`budgets.md` §8) | S–L (je Schnitt) |

Es liegt **keine geschnittene Produkt-Tranche** vor. Die Kandidaten sind im
[Risiko-Register](../risks-backlog.md) mit Triggern geführt (Roadmap-Discovery,
MR-005; Werkzeug-Einordnung in der
[Triage](../risks-backlog-werkzeug-triage.md)) — keiner ist ein aktiver Blocker.
`R-13` wird am `2026-11-02` zu einem, falls bis dahin weder geschnitten noch
erneut verlängert wird: der `expires`-Check in `scripts/render-trivyignore.sh`
bricht dann von selbst, ohne dass ein neues Advisory dazukommen muss. Anders als
bei den übrigen Kandidaten hilft Abwarten hier nicht — für vier der neun
perl-CVEs existiert auch in `sid` kein Fix, der Backport-Trigger kann also nicht
feuern (Re-Review: [`2026-08-16`](../../../reviews/2026-08-16-perl-base-cohort-rereview.md)).
Mutation-Gate-Blockierung bleibt deferred, bis echte >70 %-Score-Reihen vorliegen.

**Review-Harness — adoptiert (`slice-007`, 2026-07-24).** Das Harness (Modul
8/10, `docs/reviews/`) war eingerichtet, aber ungenutzt; `slice-007` hat die
Praxis scharf geschaltet — erster echter Handoff-Report
([`2026-07-24-slice-004.md`](../../../reviews/2026-07-24-slice-004.md)) + normative
Regel (`AGENTS.md` §5, `docs/reviews/README.md` §„Wann entsteht ein Report").
**Folge-Kandidaten (open, nur bei Bedarf):** (1) Review-Report-Gate —
automatisierte Kopplung „Slice → Report" analog `verify-closure-notes`, nur
schneiden, wenn die reine Regel driftet; (2) Verifier-/Validator-Skills (Modul
8/11) — Schwester-Rollen ohne `.harness/skills/`-Datei, Beobachtungspunkt.

## Meilensteine

| Meilenstein | Welle(n) | Trigger (extern) | Status |
|---|---|---|---|
| `0.25.0` released | Multi-Tenant-Fairness + Cutover | Tag `v0.25.0` + GHCR/npm-Publish (2026-07-13) | erreicht 2026-07-13 ([`CHANGELOG.md#0250---2026-07-13`](../../../../CHANGELOG.md#0250---2026-07-13)) |
| v3.5.0-Harness-Migration abgeschlossen | W1–W7 | W7 done, `make gates` grün (2026-07-23) | erreicht 2026-07-23 ([Plan](../done/plan-harness-v3.5.0-migration.md)) |
| `0.25.1` released | ohne Welle (Wartung/Security) | Tag `v0.25.1` + GHCR/npm-Publish (2026-08-21) | erreicht 2026-08-21 ([`CHANGELOG.md#0251---2026-08-21`](../../../../CHANGELOG.md#0251---2026-08-21)) |

## Abhängigkeitsgraph

```mermaid
flowchart LR
    W1[W1 Baseline] --> W2[W2 AGENTS.md]
    W2 --> W3[W3 Review-/Closure-Harness]
    W3 --> W4[W4 Carveouts + Triage]
    W4 --> W5[W5 Layout-Move]
    W5 --> W6[W6 Planning-Form + Roadmap]
    W6 --> W7[W7 opt-in-Module]
    W7 --> WL1[welle-01 Requirement-Link-Konvergenz]
    WL1 --> P[Produkt-Folgewelle]
```

## Abgeschlossene Wellen

| Welle | Abschluss | Closure-/Ergebnis-Record |
|---|---|---|
| Produkt-Releases `0.1.0`–`0.25.0` (25 Pläne) | bis 2026-07-13 | [`../done/plan-*.md`](../done/) · Übersicht: [`roadmap-pre-v3.5.0.md`](../done/roadmap-pre-v3.5.0.md) |
| Migration W1 — Vendored Baseline | 2026-07-22 | [Plan §2](../done/plan-harness-v3.5.0-migration.md) (`.harness/baseline/v3.5.0/` + SHA256SUMS) |
| Migration W2 — AGENTS.md | 2026-07-22 | [Plan §2](../done/plan-harness-v3.5.0-migration.md) |
| Migration W3 — Review-/Closure-Harness | 2026-07-22 | [Plan §2](../done/plan-harness-v3.5.0-migration.md) ([ADR-0010](../../adr/0010-closure-note-pflicht.md)) |
| Migration W4 — Carveouts + Werkzeug-Triage | 2026-07-23 | [Triage](../risks-backlog-werkzeug-triage.md), MR-005/006 |
| Migration W5 — Layout-Move (`docs/plan/…`) | 2026-07-23 | [Plan §3/§5](../done/plan-harness-v3.5.0-migration.md), MR-001 aufgelöst |
| Migration W6 — Planning-Form + Roadmap-Reformat | 2026-07-23 | [Plan §2](../done/plan-harness-v3.5.0-migration.md), MR-007 |
| Migration W7 — `version.md` + `versions`-Modul | 2026-07-23 | [Plan §2](../done/plan-harness-v3.5.0-migration.md); `ids` → `welle-01` |
| **welle-01 — Requirement-Link-Konvergenz** | 2026-07-23 | [`welle-01-results.md`](../done/welle-01-results.md) (slice-001 + slice-002; `ids` repo-weit) |

> Der Bestand ist grandfathered (MR-007): abgeschlossene Produktarbeit liegt als
> `plan-<version>.md` in `done/`, nicht als `welle-<NN>-results.md`. Neue Wellen
> ab der Produkt-Folgewelle folgen der kanonischen Form.

## Historische Trigger-Verschiebungen

| Datum | Was wurde geändert? | Warum? |
|---|---|---|
| 2026-07-21 | Adoption der vollen v3.5.0-Form (Wellen/Slices für neue Arbeit, Roadmap-Reformat, opt-in-Module als W7) | Owner-Entscheidung; strukturelle statt bloß versionierter Adoption |
| 2026-07-23 | Diese Roadmap neu angelegt (Kanon-Form); historienlastige Fassung nach `done/roadmap-pre-v3.5.0.md` archiviert | W6-Reformat — forward-looking Roadmap, Historie als Audit-Bestand erhalten |

> Die vollständigen Pre-v3.5.0-Verschiebungen (Trigger-Re-Evals `0.12.1`/`0.18.0`/
> `0.19.0`, Szenario-Entscheidungen `0.15.0`–`0.17.0` u. a.) stehen im Archiv
> [`roadmap-pre-v3.5.0.md`](../done/roadmap-pre-v3.5.0.md).
