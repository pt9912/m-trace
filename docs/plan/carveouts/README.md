# Carveouts — m-trace

Aktive Carveouts mit Auflösungs-Trigger (v3.5.0-Regelwerk Modul 7). Ein Carveout
ist eine **einzelne, temporäre Gate-Senkung** mit messbarem Auflösungs-Trigger und
Folge-Slice. Aufgelöste Carveouts wandern nach `done/` (reiner `git mv`).

- **Gerüst:** das vendored Template
  [`carveout.template.md`](../../../.harness/baseline/v3.5.0/templates/docs/plan/carveouts/carveout.template.md)
  wird kopiert-und-ausgefüllt nach `docs/plan/carveouts/CO-<NNN>-<kurztitel>.md`.
- **Werkzeug-Wahl vor dem Anlegen:** erst den Modul-7-Trichter prüfen (Carveout
  vs. BF-Sub-Area-Markierung vs. ADR). Ein Diskrepanz-**Cluster** im selben
  Geltungsbereich ist eine BF-Markierung, **kein** Carveout-je-Fall.

## Aktive Carveouts

_Keine aktiven generischen Carveouts._

Die W4-Werkzeug-Triage
([`risks-backlog-werkzeug-triage.md`](../planning/risks-backlog-werkzeug-triage.md))
fand keine einzelne generische Gate-Senkung, die hier zu materialisieren wäre:
Die aktiven offenen Risiken sind entweder **Roadmap-Kandidaten** (bleiben im
Risiko-Register) oder gehören zum **Security-Gate-Suppression-Cluster**, der über
die reichere domänenspezifische Registry geführt wird — siehe unten.

## Security-Gate-Suppressions (eigene Registry)

Der `image-scan`/`vuln-check`-Gate wird für einen **Cluster** transitiver
OS-CVEs der `node:22-trixie-slim`-Base gesenkt. Dieser Cluster liegt **nicht**
als CO-`NNN`-Kaskade hier, sondern in der etablierten, reicheren Registry
`.security/vulnignore.yaml` (per-CVE `reason` + `expires` + `scope`,
deterministisch nach `.trivyignore` gerendert, Nightly-Audit-Re-Eval). Modul-7-
Begründung (Cluster im selben Geltungsbereich → BF-Sub-Area-Markierung) und
Geltungsbereich stehen als [MR-006](../../../harness/conventions.md) in
`harness/conventions.md`.

## Aufgelöste Carveouts

_(noch keine)_

## Konventionen

- Jeder aktive Carveout braucht: Trigger, Folge-Slice, letztes Prüf-Datum.
- Bei Welle-Closure: Carveout-Audit zwingend — welche gültig, welche aufgelöst?
- Siehe [Kurs Modul 7](https://github.com/pt9912/ai-harness-course/blob/v3.5.0/kurs/de/02-planung/modul-07-carveouts.md).
