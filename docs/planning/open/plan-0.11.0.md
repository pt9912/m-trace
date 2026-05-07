# Implementation Plan — `0.11.0` (Ingest-Gateway / Stream Control)

> **Status**: ⬜ open — noch nicht aktiviert. Dieser Plan darf erst nach
> explizitem Move nach `docs/planning/in-progress/` umgesetzt werden.
> Vorgänger ist voraussichtlich `0.10.0` (CMAF-Signal-Analyse /
> NF-13); Aktivierung erst nach dessen Release-Closeout.
>
> **Release-Typ**: Minor-Release mit Lastenheft-Patch, neuer RAK-
> Gruppe, RAK-Verifikationsmatrix und Tag `v0.11.0`.
>
> **Ziel**: Die bisher als Kann geführten Ingest-Gateway-Funktionen
> `F-46`..`F-51` werden in einen umsetzbaren Produkt-Scope geschnitten.
> Der Plan liefert noch keine Control-Plane oder Multi-Tenant-SaaS-
> Plattform, sondern einen lokal betreibbaren Stream-Control-Pfad.
>
> **Bezug**:
> [`spec/lastenheft.md`](../../../spec/lastenheft.md) F-46..F-51,
> MVP-38; [`README.md`](../../../README.md) mit der Überschrift
> „Was m-trace nicht ist".
>
> **Nachfolger**: offen.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

In Scope:

- `F-46`: Stream-Key-Verwaltung für lokale/lab-nahe Ingest-Flows.
- `F-47`: Ingest-Endpunkte beschreiben und dokumentieren.
- `F-48`: einfache Routing-Regeln für Streams modellieren.
- `F-49`: Webhook-Events bei Stream-Start und Stream-Ende vorbereiten
  oder exemplarisch auslösen.
- `F-50`: SRT-/RTMP-Konfigurationen als beschreibbare Artefakte
  vorbereiten.
- `F-51`: Media-Server-Konfigurationen generieren oder validieren,
  begrenzt auf den im Repo vorhandenen Lab-Stack.

Out of scope:

- Keine mandantenfähige Control-Plane.
- Keine produktive Secret-Verwaltung.
- Keine globale Stream-Key-Rotation über mehrere Deployments.
- Keine automatische Provisionierung externer Media-Server.
- Kein Kubernetes-Operator.

### 0.2 Offene Vorentscheidungen

- Ob das Ingest-Gateway als eigene App unter `apps/` entsteht oder als
  API-Modul startet.
- Ob Stream-Key-Daten zunächst SQLite nutzen oder rein konfigurativ
  bleiben.
- Ob Webhooks nur dokumentiert oder bereits im lokalen Lab ausgelöst
  werden.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung, Lastenheft-Patch, RAK-Gruppe und Architekturentscheidung | ⬜ |
| 1 | Stream-Key- und Ingest-Endpunkt-Modell | ⬜ |
| 2 | Routing-Regeln und Media-Server-Konfig-Artefakte | ⬜ |
| 3 | Webhook-Events und lokale Lab-Verifikation | ⬜ |
| 4 | Doku, API-/Contract-Tests, Smokes | ⬜ |
| 5 | Gates, RAK-Verifikationsmatrix, Versions-Bump, Closeout und Tag | ⬜ |

## 2. Tranche 0 — Aktivierung und RAK-Schnitt

DoD:

- [ ] Plan von `docs/planning/open/plan-0.11.0.md` nach
  `docs/planning/in-progress/plan-0.11.0.md` verschoben.
- [ ] `git status --short` vor erster Änderung dokumentiert.
- [ ] Lastenheft-Patch mit neuer RAK-Gruppe für `F-46`..`F-51`
  ergänzt.
- [ ] Architekturentscheidung dokumentiert: eigene App oder API-Modul.
- [ ] Roadmap-Status und Release-Übersicht auf `0.11.0` als aktive
  Folgephase umgestellt.

## 3. Tranche 1 — Stream-Key- und Endpunktmodell

DoD:

- [ ] Datenmodell für Stream-Key, Ingest-Endpunkt und Stream-Ziel
  definiert.
- [ ] Validierungsregeln für lokale Lab-Nutzung dokumentiert.
- [ ] Tests decken ungültige Keys, doppelte Stream-Namen und fehlende
  Endpunkte ab.
- [ ] Keine Secrets werden in Logs, Fixtures oder Dokumentation
  veröffentlicht.

## 4. Tranche 2 — Routing und Media-Server-Konfiguration

DoD:

- [ ] Routing-Regeln sind als stabile Konfiguration beschreibbar.
- [ ] MediaMTX-nahe Konfigurationsartefakte können generiert oder
  validiert werden.
- [ ] SRT-/RTMP-Vorbereitung bleibt klar auf Lab-Scope begrenzt.
- [ ] Bestehende `examples/`-Stacks bleiben unverändert grün.

## 5. Tranche 3 — Webhooks und Lab-Verifikation

DoD:

- [ ] Stream-Start- und Stream-Ende-Ereignisse sind als Eventmodell
  dokumentiert.
- [ ] Webhook-Zustellung ist entweder implementiert oder als expliziter
  Folge-Scope mit RAK-Status begründet.
- [ ] Lokaler Smoke verifiziert den gewählten Pfad reproduzierbar.

## 6. Tranche 4 — Doku, Contracts und Smokes

DoD:

- [ ] User-Doku beschreibt lokalen Stream-Control-Workflow.
- [ ] API-/Contract-Tests pinnen das öffentliche Verhalten.
- [ ] README grenzt den Scope weiter gegen Control-Plane und
  Multi-Tenant-Betrieb ab.
- [ ] Relevante Smokes sind im Makefile dokumentiert.

## 7. Tranche 5 — Release-Closeout

DoD:

- [ ] RAK-Verifikationsmatrix vollständig ausgefüllt.
- [ ] `make docs-check` grün.
- [ ] `make build` grün.
- [ ] `make gates` grün.
- [ ] `make security-gates` grün oder CI-Job `Security gates` grün
  dokumentiert.
- [ ] Wave-2-Quality-Gates vor dem Tag geprüft.
- [ ] Vollständiger Versions-Bump auf `0.11.0`.
- [ ] `CHANGELOG.md` mit `[0.11.0] - YYYY-MM-DD` aktualisiert.
- [ ] Plan nach `docs/planning/done/plan-0.11.0.md` verschoben und
  Status auf ✅ released aktualisiert.
- [ ] Annotierter Tag `v0.11.0` erstellt.
