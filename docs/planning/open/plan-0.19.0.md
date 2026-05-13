# Implementation Plan — `0.19.0` (Follow-up: offene Roadmap-ADR-Trigger + triggerfreie Folgefragen)

> **Status**: 🟡 in der Vorbereitung — aktiv in `open/`, noch nicht archiviert.
>
> **Vorgänger**: `0.17.0` ist released und in
> [`done/plan-0.17.0.md`](../done/plan-0.17.0.md) archiviert.
> `0.18.0` ist als Decision-Closeout in
> [`done/plan-0.18.0.md`](../done/plan-0.18.0.md) archiviert; die drei
> Folge-Risiken `R-9`, `R-12` und `R-13` bleiben im
> [`in-progress/risks-backlog.md`](../in-progress/risks-backlog.md)
> offen.
>
> **Auslöser**: Offene Trigger in der Roadmap (§4 „Erwartete ADRs"):
> - Postgres als produktionsnaher Store (**MVP-40**)
> - Strengere CORS-Preflight-Project-Isolation (**Variante A**)
>
> **Zielbild**: Diese Planwelle entscheidet die beiden offenen ADR-Trigger,
> dokumentiert die Architekturentscheidung inkl. Migrations- und Kompatibilitäts
> Konsequenzen, und legt die minimale Umsetzungs- oder Defer-Route fest.

## 0. Konvention

- `[x]` erledigt
- `[ ]` offen
- `[!]` durch ADR-/Scope-Entscheidung blockiert
- `🟡` in Arbeit

## 0.1 Scope-Definition

In Scope:

- `MVP-40`: Produktionsnaher Store (`Postgres`) für `apps/api`.
- `CORS-Preflight-Project-Isolation (Variante A)`: Projektscharfe Preflight-Muster,
  ausgelöst durch reale Multi-Tenant-Anforderungen.
- `triggerfreie` Follow-ups (`RAK-102` / `RAK-103`): Entscheidungen für
  `apps/analyzer-api` und `apps/control-plane` auf Entscheidungsniveau schärfen.

Nicht in Scope:

- Umsetzungsinhalte aus `0.18.0` (`R-9`, `R-12`, `R-13`).
- Neue Produktfeatures ohne klaren ADR-Trigger.
- Produktions-Analyzer-/Control-Plane-/Analytics-Ausbau.

## 0.2 Vorgänger-Gate (informativ)

- [x] `0.17.0` abgeschlossen und archiviert.
- [x] Offene Roadmap-Trigger sind klar identifiziert.
- [ ] Auslöser-Matrix (`MVP-40`, Variante A) ist noch nicht in `0.19.0` entschieden.
- [ ] Triggerfreie Punkte (`apps/analyzer-api`, `apps/control-plane`) sind im Plan als
  Decision-Track aufgenommen und mit Owner/Abbruchkriterien versehen.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang |
| --- | --- | --- | --- | --- |
| 0 | Aktivierung und Zielabgrenzung | Triggerscope fixiert, Entscheidungen für beide Punkte vorbereitet | Roadmap zeigt offene Trigger | Freigabe für Tranche-1-Entscheidungscheck |
| 1 | ADR-/Entscheidungs-Work | Für beide Punkte: **Implement** oder **Deferred** dokumentiert | Trigger-Analyse abgeschlossen | Klare ADR-Entscheidungs-DoD je Punkt |
| 2 | Zielgerichtete Umsetzung (bei Implement) | Architektur-/Code-/Migrationsartefakte umgesetzt oder Defer sauber begründet | Tranche-1-Entscheid | DoD je Pfad erfüllt |
| 3 | Gates und Risiken | Gate-Nachweise oder dokumentierte Defer-Konditionen | Umsetzungsartefakte vorhanden | Entscheidung final dokumentiert |
| 4 | Closeout | Roadmap/Backlog-/ADR-Konsistenz abgeschlossen | Tranche 3 grün | Plan in `done/` oder in `open/` fortgeführt |

## 2. Tranche 0 — Aktivierung

DoD:

- [x] Plan im `open/`-Pfad angelegt.
- [x] Zwei offene Roadmap-Trigger als eindeutiger Scope aufgenommen.
- [ ] Entscheidungspfade (Implementierung vs. Defer) je Punkt formell beschrieben.

### 2.1 Auslöser-Matrix (initial)

| Trigger | Baseline | Standardannahme |
| --- | --- | --- |
| MVP-40 Postgres | offen (Multi-Instance/Multi-Tenant) | Defer mit klarer Wiederaufnahme bei Bedarf |
| Variante A CORS-Preflight | offen (Multi-Tenant) | Defer; Variante B bleibt bis klarer Multi-Tenant-Bedarf aktiv |

## 3. Tranche 1 — Entscheidungs-Work

### 3.1 Tranche-1-Aktionsliste `MVP-40`

- [ ] Prüfen, ob belastbarer Multi-Instance/Multi-Tenant-Trigger vorliegt:
  - realer Betreiber-Bedarf, SLO/Backups/Recovery-Trigger,
  - Persistenz-Replikations-/HA-Anforderungen,
  - Migration-/Rollback-Pfad im Team-Betrieb.
- [ ] Entscheidung dokumentieren:
  - `implemented`: Roadmap und Plan-Entscheid auf `Postgres` als produktionsnaher Store setzen,
  - `deferred`: Triggertext in ADR/Plan präzisieren und ohne Produkt-Änderung schließen.

### 3.2 Tranche-1-Aktionsliste `Variante A (striktere CORS-Preflight)`

- [ ] Prüfen, ob echter Multi-Tenant- oder projektscharfer Preflight-Bedarf besteht
  (z. B. API-Nutzung jenseits der bisherigen Variante B).
- [ ] Entscheidung dokumentieren:
  - `implemented`: Projektscharfen Preflight-Pfad als Option/Standard definieren,
  - `deferred`: Variante B beibehalten und klarer Triggertext hinterlegen.

### 3.3 Tranche-1-C — Triggerfreie Decision-Track

- [ ] `apps/analyzer-api` (`RAK-102`) als follow-up dokumentieren:
  - Konkrete Bedingungen für `proceed` oder `POC` formalisieren (externer Konsument,
    Auth-/Rate/SSRF-/Retention-Stand, Contract-Abnahme).
  - Entscheidungsmatrix im Plan: `proceed`/`POC`/`defer`/`anders erfüllt`.
- [ ] `apps/control-plane` (`F-132`, `RAK-103`) als follow-up dokumentieren:
  - Konkrete Bedingungen für `proceed`/`POC`/`defer` (mehr als eine Instanz/Project,
    Betreiberprofil, Audit-/Tenant-/Owner-Anforderung).
  - Entscheidungstransfer: kein POC/Code in `0.19.0` ohne separaten Folgeplan.

## 4. Tranche 2 — Umsetzung / Follow-up-Artefakte

## 4.1 Wenn `MVP-40` implementiert wird

- [ ] Architektur-ADR-Entwurf auf Basis `MVP-40` finalisieren.
- [ ] Migrationspfad (SQLite → Postgres) inkl. Datenhaltung, Backfill,
  Downtime-/Rollback-Ansatz und `spec/lastenheft.md`-Anpassung festhalten.
- [ ] Deployment-/Betriebserwartung (Backups, HA, Monitoring) dokumentieren.

## 4.2 Wenn Variante A implementiert wird

- [ ] Preflight-Vertragsmodell auf Projekt-Isolation umstellen oder ergänzen.
- [ ] API-Routen-/Contract-Verlauf und Client-Kompatibilität prüfen.
- [ ] Doku- und ADR-Abgleich (`docs/`, `spec/lastenheft.md`, ggf. Plan-Abschnitte).

## 4.3 Defer-Pfade

- [ ] Wenn ein Punkt deferred bleibt, den Triggertext in der Roadmap und im Plan
  präzisieren (z. B. konkrete Metrik/Betreibersicht statt allgemeiner Notiz).

## 5. Tranche 3 — Gate-Phase

- [ ] `make docs-check` grün.
- [ ] Für implementierte Punkte:
  - Zielbezogene Migrations-/Kompatibilitätsnachweise vorhanden.
  - Gate-Nachweise (sofern eingeführt) grün.
- [ ] Für deferred Punkte:
  - Nachweis, warum kein Defer-Verstoß vorliegt,
  - Triggertext ist eindeutig und reproduzierbar.

## 6. Tranche 4 — Closeout

- [ ] Beide Punkte mit Abschlussstatus in der Roadmap (§4) aktualisiert.
- [ ] `docs/planning/in-progress/roadmap.md` reflektiert den Folgeplan-Stand.
- [ ] `RAK-102` (`apps/analyzer-api`) und `RAK-103` (`apps/control-plane`)
  sind als `Decision-Record` mit klarem Proceed/Defer-Trigger, Owner und
  nächstem Auslöser abgeschlossen.
- [ ] `docs/planning/open/plan-0.19.0.md` in `done/` verschieben, falls vollständig abgeschlossen.
- [ ] Bei Restoffen den Nachfolgepfad benennen (z. B. `0.20.0`), ohne Scope-Drift.

## 7. Triggerfreie Folgefragen ohne harte Roadmap-Konditionen

Ziel: zwei Themen bleiben als operator- oder stakeholdergetriebener Follow-up auf
Decisionschiene, ohne dass sie als Implementations-Work in `0.19.0` codeseitig umgesetzt
werden müssen.

### 7.1 Scope

- `apps/analyzer-api` (`RAK-102`, `MVP-20`) bleibt ohne Trigger-ID in `defer`, solange kein
  konkreter externer Konsument, Auth-/Rate-Limit-/SSRF-/Retention-Nachweis, Contract-Finalisierung
  und Folgeplan vorliegt.
- `apps/control-plane` (`F-132`, `RAK-103`) bleibt ohne Trigger-ID in `defer`, solange kein klarer
  Betreiberbedarf mit Operator-/Tenant-/Audit-Nachweis, SLO und Produkt-Owner-Freigabe vorliegt.

### 7.2 DoD

- [ ] Für beide Themen existiert ein kurzer, verbindlicher Decision-Record-Abschnitt in
  `docs/planning/open/plan-0.19.0.md` oder `docs/adr/`.
- [ ] Die Aufnahmekriterien für den nächsten Folgeplan sind dokumentiert:
  - `apps/analyzer-api`: externer Nutzenfall, Sicherheitsmodell, Retention, Owner.
  - `apps/control-plane`: Zielgruppenfit (beyond MVP), Admin-/Operator-Scope, Audit-/Compliance-Ausblick.
- [ ] Keine Runtime-/Schema-/API-Änderung in `0.19.0` für diese beiden Punkte.
