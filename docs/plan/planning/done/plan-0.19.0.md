# Implementation Plan — `0.19.0` (Follow-up: offene Roadmap-ADR-Trigger + triggerfreie Folgefragen)

> **Status**: ✅ abgeschlossen — Decision-only-Plan archiviert in
> `done/`. Kein Release-Tag, kein Versions-Bump, keine Runtime-,
> Schema-, Storage- oder Public-API-Änderung.
>
> **Vorgänger**: `0.18.0` ist released und als Decision-Closeout in
> [`plan-0.18.0.md`](./plan-0.18.0.md) archiviert; die drei
> Folge-Risiken `R-9`, `R-12` und `R-13` bleiben im
> [`risks-backlog.md`](../risks-backlog.md)
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

- [x] `0.18.0` released und archiviert.
- [x] Offene Roadmap-Trigger sind klar identifiziert.
- [x] Auslöser-Matrix (`MVP-40`, Variante A) ist in `0.19.0`
  bewertet.
- [x] Triggerfreie Punkte (`apps/analyzer-api`, `apps/control-plane`)
  sind im Plan als Decision-Track aufgenommen und mit Owner/
  Abbruchkriterien versehen.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang |
| --- | --- | --- | --- | --- |
| 0 | Aktivierung und Zielabgrenzung | Triggerscope fixiert, Entscheidungen für beide Punkte vorbereitet | Roadmap zeigt offene Trigger | Freigabe für Tranche-1-Entscheidungscheck |
| 1 | ADR-/Entscheidungs-Work | Für beide Punkte: **Implement** oder **Deferred** dokumentiert | Trigger-Analyse abgeschlossen | Klare ADR-Entscheidungs-DoD je Punkt |
| 2 | Zielgerichtete Umsetzung (bei Implement) | Architektur-/Code-/Migrationsartefakte umgesetzt oder Defer sauber begründet | Tranche-1-Entscheid | DoD je Pfad erfüllt |
| 3 | Gates und Risiken | Gate-Nachweise oder dokumentierte Defer-Konditionen | Umsetzungsartefakte vorhanden | Entscheidung final dokumentiert |
| 4 | Closeout | Roadmap/Backlog-/ADR-Konsistenz abgeschlossen | Tranche 3 grün | Plan in `done/` archiviert oder in `in-progress/` fortgeführt |

## 2. Tranche 0 — Aktivierung

DoD:

- [x] Plan im `open/`-Pfad angelegt und zur Bearbeitung nach
  `in-progress/` verschoben.
- [x] Zwei offene Roadmap-Trigger als eindeutiger Scope aufgenommen.
- [x] Entscheidungspfade (Implementierung vs. Defer) je Punkt formell
  beschrieben.

### 2.1 Auslöser-Matrix

| Trigger | Baseline | Bewertung 2026-05-13 | Standardannahme |
| --- | --- | --- | --- |
| `MVP-40` Postgres | ADR 0005: `defer-with-migration-seed`; `0.14.0` und `0.17.0` fanden keinen neuen Last-/Multi-Host-/Recovery-Trigger | Kein neuer Betreiber-, Multi-Replica-, Recovery- oder Retention-Beleg im Repo. `make gates`/Release-Pfade bleiben SQLite-default. | Deferred; Wiederaufnahme nur bei messbarer ADR-0005-Schwelle. |
| Variante A CORS-Preflight | Variante B ist produktiv: globale konservative Preflight-Union, project-spezifische Policy erst beim echten Request | Kein echter Multi-Tenant-/Project-in-URL-Bedarf und kein Client-Vertrag jenseits der bestehenden Browser-Ingest-Policy belegt. | Deferred; Variante B bleibt Standard. |

### 2.2 Decision-Modus

`0.19.0` ist eine Decision-only-Welle. Ohne belegten Trigger wird kein
Runtime-, Wire-, Schema-, Storage- oder Public-API-Pfad eingeführt.
Ein `implemented`-Pfad ist nur zulässig, wenn der jeweilige Trigger vor
Tranche 2 mit konkretem Betreiber-/Last-/Tenant-Nachweis belegt wird.

## 3. Tranche 1 — Entscheidungs-Work

### 3.1 Tranche-1-Aktionsliste `MVP-40`

- [x] Prüfen, ob belastbarer Multi-Instance/Multi-Tenant-Trigger vorliegt:
  - realer Betreiber-Bedarf, SLO/Backups/Recovery-Trigger,
  - Persistenz-Replikations-/HA-Anforderungen,
  - Migration-/Rollback-Pfad im Team-Betrieb.
- [x] Entscheidung dokumentieren:
  - `implemented`: Roadmap und Plan-Entscheid auf `Postgres` als produktionsnaher Store setzen,
  - `deferred`: Triggertext in ADR/Plan präzisieren und ohne Produkt-Änderung schließen.

#### 3.1.1 Entscheidung — `MVP-40` Postgres

**Entscheidung:** `deferred`, ADR 0005 bleibt unverändert gültig.
`0.19.0` liefert keinen Postgres-Runtime-Adapter, keinen DSN-Selector,
keinen Dual-Write, keinen automatischen SQLite-Export und keine
Postgres-Pflicht in Tests oder lokaler Entwicklung.

**Begründung:** Die bisherigen Defer-Nachweise aus `0.13.0`,
`0.14.0`, `0.15.0` und `0.17.0` wurden nicht durch neue Fakten
überholt. Es gibt keinen belegten Bedarf für zwei oder mehr API-
Replicas auf gemeinsamem Store, kein verbindliches Recovery-SLO und
keine Retention-/Read-Last, die SQLite plus bestehende API-Pfade
überschreitet.

**Wiederaufnahme-Trigger:** genau die ADR-0005-Schwellen bleiben
maßgeblich:

| Trigger | Schwelle | Owner |
| --- | --- | --- |
| Multi-Replica-Store | `>= 2` API-Replicas brauchen denselben Store ohne shared-volume SQLite | Platform/Ops |
| Recovery-SLO | `RPO <= 15 min` oder `RTO <= 30 min` wird verbindlich | Platform/Ops |
| Retention-/Query-Last | mehr als 10 Mio. Events mit p95-Read-Anforderung `< 2 s` | Platform/QA |

**Nicht entschieden:** Postgres-DDL, Adapter-Schnitt, Migration,
Backfill, Rollback und Backup/Restore bleiben Folgeplan-Scope nach
ausgelöstem Trigger.

### 3.2 Tranche-1-Aktionsliste `Variante A (striktere CORS-Preflight)`

- [x] Prüfen, ob echter Multi-Tenant- oder projektscharfer Preflight-Bedarf besteht
  (z. B. API-Nutzung jenseits der bisherigen Variante B).
- [x] Entscheidung dokumentieren:
  - `implemented`: Projektscharfen Preflight-Pfad als Option/Standard definieren,
  - `deferred`: Variante B beibehalten und klarer Triggertext hinterlegen.

#### 3.2.1 Entscheidung — CORS-Preflight-Project-Isolation Variante A

**Entscheidung:** `deferred`. Variante B bleibt Standard:
Preflight prüft eine globale konservative Origin-Union ohne Project-
Enumeration; project-spezifische Origin-/Audience-/Policy-Prüfung
passiert erst beim echten Request, wenn Project- oder Session-Kontext
belastbar vorliegt.

**Begründung:** Browser-Preflights tragen keinen verlässlichen Project-
oder Session-Token-Kontext. Eine Variante-A-Migration würde eine neue
URL-Struktur wie `/api/projects/{project_id}/...` oder einen
preflight-fähigen Project-Parameter als Public-API-Vertrag einführen.
Dafür liegt aktuell kein Multi-Tenant-Client, kein externer SDK-
Konsument und kein Betreiberbedarf vor.

**Wiederaufnahme-Trigger:**

| Trigger | Schwelle | Owner |
| --- | --- | --- |
| Multi-Tenant-Browser-Client | Ein realer Client braucht Preflight-Entscheidungen pro Project statt globaler Union | Product/API |
| Project-in-URL-Kontrakt | Ein Folgeplan führt `/api/projects/{project_id}/...` oder äquivalenten URL-skopierten Project-Kontext ein | API |
| Security-/Audit-Befund | Audit verlangt Preflight-Isolation vor dem eigentlichen Request und akzeptiert die notwendige API-Änderung | Security/API |

**Nicht entschieden:** Keine neue Route, kein neuer Query-Parameter,
keine Änderung an CORS-Headern, kein Lastenheft-Patch in Tranche 1.

### 3.3 Tranche-1-C — Triggerfreie Decision-Track

- [x] `apps/analyzer-api` (`RAK-102`) als follow-up dokumentieren:
  - Konkrete Bedingungen für `proceed` oder `POC` formalisieren (externer Konsument,
    Auth-/Rate/SSRF-/Retention-Stand, Contract-Abnahme).
  - Entscheidungsmatrix im Plan: `proceed`/`POC`/`defer`/`anders erfüllt`.
- [x] `apps/control-plane` (`F-132`, `RAK-103`) als follow-up dokumentieren:
  - Konkrete Bedingungen für `proceed`/`POC`/`defer` (mehr als eine Instanz/Project,
    Betreiberprofil, Audit-/Tenant-/Owner-Anforderung).
  - Entscheidungstransfer: kein POC/Code in `0.19.0` ohne separaten Folgeplan.

#### 3.3.1 Decision-Record — `apps/analyzer-api` (`RAK-102`)

**Entscheidung:** `defer`. Der bestehende interne
`apps/analyzer-service` plus `@npm9912/stream-analyzer` Library/CLI
bleibt Standard. Kein neues `apps/analyzer-api`-Artefakt in `0.19.0`.

| Option | Bedingung | Ergebnis |
| --- | --- | --- |
| `proceed` | Externer Konsument mit akzeptiertem Auth-/Rate-/SSRF-/Retention- und Contract-Modell | Nicht belegt |
| `POC` | Zeitbegrenzter Owner, Contract-Abnahmeziel und Abbruchdatum | Nicht belegt |
| `defer` | Kein externer Konsument oder Security-/Retention-Modell fehlt | Gewählt |
| `anders erfüllt` | Interner Service/Library/CLI decken den aktuellen Bedarf | Aktueller Standard |

**Nächster Auslöser:** benannter externer Konsument plus Folgeplan, der
Auth, Rate Limits, SSRF-Grenzen, Retention, Contract-Versionierung und
Abbruchkriterien vor Codebeginn akzeptiert.

#### 3.3.2 Decision-Record — `apps/control-plane` (`F-132`, `RAK-103`)

**Entscheidung:** `defer`. Kein `apps/control-plane`-POC und keine
Admin-/Operator-UI in `0.19.0`.

| Option | Bedingung | Ergebnis |
| --- | --- | --- |
| `proceed` | Mehr als eine produktnahe Instanz oder mehrere Projects brauchen Admin-/Tenant-/Audit-Workflows | Nicht belegt |
| `POC` | Betreiberprofil, Owner, SLO, Audit-/Compliance-Frage und Abbruchdatum sind benannt | Nicht belegt |
| `defer` | Kein konkreter Betreiberbedarf und keine Product-Owner-Freigabe | Gewählt |

**Nächster Auslöser:** Betreiberbedarf mit mindestens einem echten
Admin-Workflow, Tenant-/Owner-Modell, Audit-Anforderung, SLO und
separatem Folgeplan. Ohne diesen Nachweis bleibt die lokale Lab- und
Selfhoster-Zielgruppe ausreichend durch Doku, ENV und bestehende APIs
bedient.

## 4. Tranche 2 — Umsetzung / Follow-up-Artefakte

### 4.1 Wenn `MVP-40` implementiert wird

- [!] Architektur-ADR-Entwurf auf Basis `MVP-40` finalisieren.
- [!] Migrationspfad (SQLite → Postgres) inkl. Datenhaltung, Backfill,
  Downtime-/Rollback-Ansatz und `spec/lastenheft.md`-Anpassung festhalten.
- [!] Deployment-/Betriebserwartung (Backups, HA, Monitoring) dokumentieren.

Nicht gezogen, weil Tranche 1 keinen `implemented`-Entscheid getroffen
hat.

### 4.2 Wenn Variante A implementiert wird

- [!] Preflight-Vertragsmodell auf Projekt-Isolation umstellen oder ergänzen.
- [!] API-Routen-/Contract-Verlauf und Client-Kompatibilität prüfen.
- [!] Doku- und ADR-Abgleich (`docs/`, `spec/lastenheft.md`, ggf. Plan-Abschnitte).

Nicht gezogen, weil Tranche 1 keinen `implemented`-Entscheid getroffen
hat.

### 4.3 Defer-Pfade

- [x] Wenn ein Punkt deferred bleibt, den Triggertext in der Roadmap und im Plan
  präzisieren (z. B. konkrete Metrik/Betreibersicht statt allgemeiner Notiz).

### 4.4 Tranche-2-Ergebnis

Alle vier Tracks sind in Tranche 1 deferred. Damit gibt es in `0.19.0`
keine Runtime-/Schema-/API-/Storage-Umsetzung. Tranche 2 ist auf die
Dokumentation der Triggertexte und die Roadmap-Konsistenz begrenzt.

## 5. Tranche 3 — Gate-Phase

- [x] `make docs-check` grün (`2026-05-13`).
- [!] Für implementierte Punkte:
  - Zielbezogene Migrations-/Kompatibilitätsnachweise vorhanden.
  - Gate-Nachweise (sofern eingeführt) grün.
- [x] Für deferred Punkte:
  - Nachweis, warum kein Defer-Verstoß vorliegt,
  - Triggertext ist eindeutig und reproduzierbar.

## 6. Tranche 4 — Closeout

- [x] Beide Punkte mit Abschlussstatus in der Roadmap (§4) aktualisiert.
- [x] `docs/planning/in-progress/roadmap.md` reflektiert den Folgeplan-Stand.
- [x] `RAK-102` (`apps/analyzer-api`) und `RAK-103` (`apps/control-plane`)
  sind als `Decision-Record` mit klarem Proceed/Defer-Trigger, Owner und
  nächstem Auslöser abgeschlossen.
- [x] `docs/planning/in-progress/plan-0.19.0.md` nach
  `docs/planning/done/plan-0.19.0.md` verschieben,
  falls vollständig abgeschlossen.
- [x] Bei Restoffen den Nachfolgepfad benennen (z. B. `0.20.0`), ohne Scope-Drift.

### 6.1 Closeout-Ergebnis

`0.19.0` schließt ohne Implementierung ab:

- `MVP-40` Postgres bleibt offen bis zu einer ADR-0005-Schwelle.
- CORS-Preflight-Variante A bleibt offen bis zu einem echten
  Multi-Tenant-/Project-in-URL-/Audit-Trigger.
- `apps/analyzer-api` (`RAK-102`) bleibt deferred bis zu externem
  Konsumenten plus Security-/Retention-/Contract-Folgeplan.
- `apps/control-plane` (`F-132`, `RAK-103`) bleibt deferred bis zu
  Betreiber-, Tenant-, Audit-, Owner- und SLO-Nachweis.

Ein separater Folgeplan wird erst angelegt, wenn einer dieser Trigger
belegt ist; `0.20.0` wird nicht durch `0.19.0` vorreserviert.

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

- [x] Für beide Themen existiert ein kurzer, verbindlicher
  Decision-Record-Abschnitt in
  `docs/planning/done/plan-0.19.0.md`.
- [x] Die Aufnahmekriterien für den nächsten Folgeplan sind dokumentiert:
  - `apps/analyzer-api`: externer Nutzenfall, Sicherheitsmodell, Retention, Owner.
  - `apps/control-plane`: Zielgruppenfit (beyond MVP), Admin-/Operator-Scope, Audit-/Compliance-Ausblick.
- [x] Keine Runtime-/Schema-/API-Änderung in `0.19.0` für diese beiden Punkte.
