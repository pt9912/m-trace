# Implementation Plan — `0.18.0` (Follow-up: offene Risiken)

> **Status**: 🟡 in der Vorbereitung — aktiv in `open/`, noch nicht
> archiviert.
>
> **Vorgänger**: `0.17.0` (Hardening / Evidence Review), archiviert in
> [`../done/plan-0.17.0.md`](../done/plan-0.17.0.md).
>
> **Auslöser**: Drei aktive Punkte in
> [`risks-backlog.md`](../in-progress/risks-backlog.md): `R-9`, `R-12`,
> `R-13`.
>
> **Release-Typ**: Follow-up-Patch oder -Minor (abhängig von der finalen
> Umsetzungstiefe).
>
> **Zielbild**: Die drei offenen Risiken werden entweder vollständig
> entschärft oder auf einen klaren Folge-Trigger reduziert, ohne neue
> produktive Scope-Flächen zu eröffnen.

## 0. Konvention

- `[x]` erledigt
- `[ ]` offen
- `[!]` durch ADR-/Scope-Entscheidung blockiert
- `🟡` in Arbeit

## 0.1 Scope-Definition

In Scope:

- `R-9`: Follow-up für K8s-Observability-Label- und Allowlist-Handling.
- `R-13`: Trivy-Re-Review-Aktionspfad und mögliche End-of-Ignore-Entscheidung
  für `node:22-trixie-slim`.
- `R-12`: Nachführung der WebRTC-`getStats`-Drift-Reaktion zu einem
  klaren Release-Entscheidungsprozess.

Nicht in Scope:

- Neue Produktfeatures außerhalb der drei Risiken.
- Produktiv-Analyzer-/Control-Plane-/Postgres-/Analytics-Ausbau.
- Neue Metrik- oder Schemaänderungen ohne unmittelbaren Bezug zu
  `R-9`, `R-12`, `R-13`.

## 0.2 Vorgänger-Gate (informativ)

- [x] `0.17.0` ist abgeschlossen und in `docs/planning/done/` archiviert.
- [x] Die offenen Risiken im Backlog sind eindeutig identifiziert.
- [ ] Release-Entscheidungs-Kriterien für `0.18.0` sind final festgelegt.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang |
| --- | --- | --- | --- | --- |
| 0 | Aktivierung + Scope-Fixierung | Zielkonflikte aufgelöst, DoD-Logik gesetzt | Risiken sind als offen markiert | Follow-up-Scope und Priorisierung gesetzt |
| 1 | Trigger- und Entscheidungs-Work | Entscheidungsstatus pro Risiko (Umsetzen vs. Deferred) | Plan aktiv | Für jedes Risiko: Implement- oder Defer-Entscheidung |
| 2 | Risiken gezielt bearbeiten | Artefakte/ADR-Nachweise pro Risiko erstellt | Tranche-1-Entscheide | DoD pro Risiko teilweise oder vollständig erfüllt |
| 3 | Gate- und Nachweis-Phase | Relevante Gates/Smokes im Standardfluss | Risiko-Artefakte vorhanden | Entscheidungsstand dokumentiert |
| 4 | Closeout | Backlog- und Roadmap-Abgleich, optionales Releasecloseout | Tranche 3 grün | Plan in `done/` oder Weiterführung in `open/` |

## 2. Tranche 0 — Aktivierung + Entscheidungsvorlage

DoD:

- [x] Plan im `open/`-Pfad angelegt.
- [x] `R-9`, `R-12`, `R-13` als einziges Folge-Agenda gesetzt.
- [ ] Umfang zwischen kleinem Doku-/Ops-Artefakt und Code-Umsetzung auf einen klaren Satz reduziert.

### 2.1 Entscheidungsvorlage (zu finalisieren)

| Risiko | Baseline-Status | Kandidat A | Kandidat B | Vorauswahl |
| --- | --- | --- | --- | --- |
| R-9 | offen (`⬜`) | K8s-Observability bleibt deferred, keine neuen K8s-Smokes | eigener K8s-Label-Allowlist-Modus + Smoke-Profiltrennung | `TBD` |
| R-13 | offen (`⬜`) | Trivy-Ignores mit strikter Re-Review-Fortführung | Distroless-Umstieg evaluieren und ggf. einführen | `TBD` |
| R-12 | offen (`⬜`) | Status-quo fortführen (Nightly + optional Safari) | `make smoke-webrtc-stats-drift` in Release-Gates integrieren | `TBD` |

### 2.2 Harte Exit-Kriterien je Risiko

Jeder Risikopfad gilt nur dann als abgeschlossen, wenn für die
entscheidende Variante mindestens ein kompletter Satz harter Kriterien
erfüllt ist.

#### R-9 (K8s-Observability)

Implementiert:

- [ ] `make k8s-validate` inkl. K8s-spezifischer Labelpolicy ist definiert
  und im Gate-Flow reproduzierbar dokumentiert.
- [ ] `R-9` ist in
  `docs/planning/in-progress/risks-backlog.md` mit `🟢` und
  nachvollziehbaren Abschluss-Referenz dokumentiert.

Defer:

- [ ] `R-9` bleibt offen (`⬜`) mit explizitem Triggertext
  (z. B. „bei K8s-Smoke-Gate“), ohne dass K8s-spezifische
  produktive Änderungen eingeführt wurden.

#### R-13 (CVE-Re-Review)

Implementiert:

- [ ] Nachweis, dass mindestens eine Maßnahme entschieden und ausgeführt
  wurde: neuer Trivy-Durchlauf, Ignored-Update mit nachvollziehbaren Gründen
  **oder** Distroless-Basis-Entscheid inkl. Build-/Runtime-Nachweis.
- [ ] `R-13` ist in
  `docs/planning/in-progress/risks-backlog.md` mit `🟢` oder explizitem
  `⬛` und Abschlussreferenz dokumentiert.

Defer:

- [ ] `R-13` bleibt offen (`⬜`) mit exakt dokumentiertem Triggertext
  (z. B. „kein Upstream-Fix bis `expires`“), inklusive dem letzten
  Re-Review-Datum in der Ignore-Historie.

#### R-12 (WebRTC-Drift)

Implementiert:

- [ ] Der WebRTC-Drift-Pfad ist im ausgewählten Release-Gate verankert
  und scheitert bei echter Release-blockierender Schemaabweichung.
- [ ] `R-12` ist in
  `docs/planning/in-progress/risks-backlog.md` mit `🟢` und
  Abschlussreferenz dokumentiert.

Defer:

- [ ] `R-12` bleibt offen (`⬜`) mit explizitem Triggertext
  (z. B. Safari als verpflichtender Zielbrowser / Supportanforderung),
  inklusive grüner Drift-Nachtschicht-Historie.

## 3. Tranche 1 — Trigger-Re-Eval

### 3.1 R-9: K8s-Observability-Trigger

- [ ] Prüfen, ob K8s-Smoke in der Roadmap als Gate geplant ist.
- [ ] Prüfen, ob Compose- und K8s-Layers sauber trennbar sind
  (`smoke scope`, Labels, README).
- [ ] Entscheidung dokumentieren:
  - `deferred`: kein weiteres Artefakt, Offene-Punkt bleibt im Backlog
  - `implemented`: separater K8s-Allowlist-Modus + Smoke-Profiltrennung

### 3.2 R-13: CVE-Re-Review

- [ ] Prüfen, ob Upstream-Fixes für die drei CVEs verfügbar sind.
- [ ] Prüfen, ob die aktuelle `expires`-Schwelle erreicht ist.
- [ ] Entscheidung dokumentieren:
  - `continued`: Ignore erneuern + harte Re-Review-Logik dokumentieren
  - `migrated`: distroless als produktive Basis-Entscheidung vorbereiten

### 3.3 R-12: WebRTC-Drift-Kontur

- [ ] Prüfen, ob Nightly-Drift-Smoke stabil grün ist.
- [ ] Prüfen, ob Safari/WebKit als verbindlicher Produktiv- oder
  Support-Trigger aufgenommen wurde.
- [ ] Entscheidung dokumentieren:
  - `deferred`: weiterhin „automatisch detektiert“
  - `migrated`: Browserumfang und Release-Kontrolle erweitern

## 4. Tranche 2 — Umsetzung / Follow-up-Artefakte

DoD je Risiko-Entscheid:

### 4.1 Wenn `R-9` implementiert wird

- [ ] K8s-Labelprofil (`pod`, `namespace`, `container`, optional
  `service`) als separater Modus implementieren.
- [ ] Smoke-Profil-Trennung Compose vs. K8s im Skript-/Env-Flow setzen.
- [ ] Dokumentation anpassen: `deploy/k8s/*` bleibt optional und nicht
  production-ready ohne Trigger.

### 4.2 Wenn `R-13` implementiert wird

- [ ] Distroless-Basis (`gcr.io/distroless/nodejs22-debian12`) als PoC
  für Dashboard/Analyzer evaluieren.
- [ ] Build- und Tooling-Pfade auf Nicht-Bedarf außerhalb CI prüfen.
- [ ] Migrations-/Rollback-Nachweis und Security-Auswirkung dokumentieren.

### 4.3 Wenn `R-12` implementiert wird

- [ ] Release-Gate-Mechanik prüfen/erweitern, damit Drift-Ergebnis vor
  jedem Gate explizit bewertet wird.
- [ ] Optional Safari/WebKit als regulären Browser im Drift-Loop aufnehmen,
  falls die entsprechende Unterstützung bindend wird.

## 5. Tranche 3 — Gate-Phase

- [ ] `make docs-check` grün.
- [ ] Relevante Tranche-2-Artefakte grün:
  - `make smoke-webrtc-stats-drift` (R-12-Pfad)
  - `make k8s-validate` nur bei `R-9`-Implementierung (sonst als
    dokumentierter Defer-Entscheid ausgeschlossen)
  - Trivy- und Security-Nachweise für den CVE-Pfad (R-13)
- [ ] `docs/planning/in-progress/risks-backlog.md` auf
  `0.18.0`-Konsequenz aktualisieren.

## 6. Tranche 4 — Closeout

- [ ] Für `R-9`, `R-12`, `R-13` einen Abschlussstatus setzen
  (`🟢`/`⬜`/`⬛`) und Begründung dokumentieren.
- [ ] Bei vollständiger Umsetzung: Plan nach
  `docs/planning/done/plan-0.18.0.md` verschieben.
- [ ] Bei Restoffen: Auslöser/Resttrigger präzisieren und Weiterführung
  explizit in Backlog + Roadmap verankern.
