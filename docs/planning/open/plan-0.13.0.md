# Implementation Plan — `0.13.0` (Production / Ops Backends)

> **Status**: ⬜ open — noch nicht aktiviert. Dieser Plan darf erst nach
> explizitem Move nach `docs/planning/in-progress/` umgesetzt werden.
> Vorgänger ist voraussichtlich `0.12.0`; Aktivierung erst nach dessen
> Release-Closeout.
>
> **Release-Typ**: Minor-Release mit Lastenheft-Patch, neuer RAK-
> Gruppe, RAK-Verifikationsmatrix und Tag `v0.13.0`.
>
> **Ziel**: Production-/Ops-nahe Folgepunkte werden in einen
> entscheidbaren Scope überführt: `MVP-40` Postgres, `MVP-41`
> ClickHouse/VictoriaMetrics, `MVP-42` Kubernetes-Manifeste,
> `MVP-43` Devcontainer und `MVP-44` Release-Automatisierung. `NF-18`
> wird dabei mit `MVP-42` harmonisiert.
>
> **Bezug**:
> [`spec/lastenheft.md`](../../../spec/lastenheft.md) NF-18,
> MVP-40..MVP-44; [`docs/planning/in-progress/roadmap.md`](../in-progress/roadmap.md)
> Folge-ADRs; [`docs/planning/in-progress/risks-backlog.md`](../in-progress/risks-backlog.md)
> R-9.
>
> **Nachfolger**: offen.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR- oder Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

Dieser Plan ist ein Decision-and-Seed-Release für Production-/Ops-
Backends. Er darf einzelne Artefakte liefern, muss aber nicht alle
Backends produktionsreif implementieren.

In Scope:

- `MVP-40`: Postgres als produktionsnaher Store bewerten und ggf.
  ersten Adapter-/Migration-Slice planen.
- `MVP-41`: ClickHouse oder VictoriaMetrics als hochvolumiges Event-
  Backend bewerten.
- `MVP-42` / `NF-18`: Kubernetes-Manifeste als optionalen Folgepfad
  konkretisieren oder weiterhin begründet deferred lassen.
- `MVP-43`: Devcontainer für reproduzierbare Entwicklung bewerten.
- `MVP-44`: Release-Automatisierung bewerten und erste sichere
  Automationsschritte definieren.
- R-9 prüfen: K8s-Smoke-Stage würde Observability-Label-Allowlists
  beeinflussen.

Out of scope:

- Kein vollständiger Production-Kubernetes-Betrieb.
- Kein Managed-Cloud-Betrieb.
- Kein Multi-Tenant-SaaS-Produkt.
- Keine ClickHouse-/VictoriaMetrics-Pflicht, solange kein konkreter
  Analysebedarf besteht.
- Keine automatische Releases ohne manuelle Freigabe.

### 0.2 Offene Vorentscheidungen

- Ob `0.13.0` ein reiner ADR-/Seed-Release bleibt oder einen ersten
  Postgres-Adapter-Slice liefert.
- Ob K8s-Manifeste als Beispielartefakte oder nur als Folgeplan
  behandelt werden.
- Ob Devcontainer und Release-Automatisierung zusammen in diesem Plan
  bleiben oder vor Aktivierung getrennt werden.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung, Lastenheft-Patch, ADR-Schnitt und Scope-Gates | ⬜ |
| 1 | Postgres-Entscheidung und ggf. Adapter-Seed (`MVP-40`) | ⬜ |
| 2 | Analytics-Backend-Entscheidung (`MVP-41`) | ⬜ |
| 3 | Kubernetes-/Devcontainer-Scope (`MVP-42`, `MVP-43`, `NF-18`, R-9) | ⬜ |
| 4 | Release-Automatisierung (`MVP-44`) | ⬜ |
| 5 | Gates, RAK-Verifikationsmatrix, Versions-Bump, Closeout und Tag | ⬜ |

## 2. Tranche 0 — Aktivierung und Scope-Gates

DoD:

- [ ] Plan von `docs/planning/open/plan-0.13.0.md` nach
  `docs/planning/in-progress/plan-0.13.0.md` verschoben.
- [ ] `git status --short` vor erster Änderung dokumentiert.
- [ ] Lastenheft-Patch mit neuer RAK-Gruppe für `MVP-40`..`MVP-44`
  und `NF-18` ergänzt.
- [ ] ADR-Schnitt entschieden: welche Entscheidungen werden als ADR
  geschrieben, welche bleiben Plan-DoD.
- [ ] Roadmap-Status und Release-Übersicht auf `0.13.0` als aktive
  Folgephase umgestellt.

## 3. Tranche 1 — Postgres

DoD:

- [ ] Entscheidung dokumentiert: Postgres jetzt implementieren,
  Adapter-Seed liefern oder weiter deferieren.
- [ ] Falls implementiert: Migration-/Repository-Scope ist begrenzt
  und Contract-Tests pinnen SQLite-Kompatibilität.
- [ ] Falls deferred: Trigger für spätere Umsetzung ist konkret
  dokumentiert.

## 4. Tranche 2 — Analytics-Backend

DoD:

- [ ] ClickHouse/VictoriaMetrics/Mimir-Bedarf ist anhand aktueller
  Datenpfade bewertet.
- [ ] Entscheidung dokumentiert: kein Backend, ein bevorzugter
  Kandidat oder Folge-ADR.
- [ ] Keine neue Pflichtabhängigkeit im lokalen Standard-Setup.

## 5. Tranche 3 — Kubernetes und Devcontainer

DoD:

- [ ] `NF-18` und `MVP-42` sind harmonisiert: optionaler K8s-Scope
  oder konkrete Beispielmanifeste.
- [ ] R-9-Auswirkung auf Observability-Smokes geprüft.
- [ ] Devcontainer-Scope für `MVP-43` entschieden und dokumentiert.
- [ ] README-Abgrenzung bleibt konsistent: kein Production-Ready-K8s.

## 6. Tranche 4 — Release-Automatisierung

DoD:

- [ ] Sicherer Automationsumfang für `MVP-44` entschieden.
- [ ] Keine automatische Veröffentlichung ohne explizite menschliche
  Freigabe.
- [ ] Release-Doku bleibt Source of Truth für manuelle Schritte.
- [ ] Automations- oder Dry-Run-Tests sind definiert.

## 7. Tranche 5 — Release-Closeout

DoD:

- [ ] RAK-Verifikationsmatrix vollständig ausgefüllt.
- [ ] `make docs-check` grün.
- [ ] `make build` grün.
- [ ] `make gates` grün.
- [ ] `make security-gates` grün oder CI-Job `Security gates` grün
  dokumentiert.
- [ ] Wave-2-Quality-Gates vor dem Tag geprüft.
- [ ] Vollständiger Versions-Bump auf `0.13.0`.
- [ ] `CHANGELOG.md` mit `[0.13.0] - YYYY-MM-DD` aktualisiert.
- [ ] Plan nach `docs/planning/done/plan-0.13.0.md` verschoben und
  Status auf ✅ released aktualisiert.
- [ ] Annotierter Tag `v0.13.0` erstellt.
