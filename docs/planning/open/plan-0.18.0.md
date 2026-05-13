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
- [x] Release-Entscheidungs-Kriterien für `0.18.0` sind final festgelegt:
  `0.18.0` aktiviert nur dann Code, wenn Tranche 1 einen echten Trigger
  belegt; sonst bleibt der Release ein Doku-/Ops-Decision-Follow-up mit
  reproduzierbaren Nachweisen.

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
- [x] Umfang zwischen kleinem Doku-/Ops-Artefakt und Code-Umsetzung auf
  einen klaren Satz reduziert: Tranche 1 ist ein Trigger-Re-Eval; Tranche
  2 liefert nur dann Code, wenn Tranche 1 einen der Implementierungs-
  Trigger belegt. Ohne Trigger endet `0.18.0` als Doku-/Ops-Decision-
  Patch mit Backlog-/Roadmap-Fortschreibung.

### 2.0 Scope-Fixierung

`0.18.0` schneidet die drei aktiven Risiken nicht als pauschales
Hardening-Bundle, sondern als Entscheidungswelle:

- `R-9`: Kein K8s-Allowlist-Modus ohne neue K8s-Smoke-Stage oder
  Prometheus-Scrape-/Label-Policy im K8s-Pfad. `make k8s-validate`
  bleibt ein clusterfreier Manifest-Seed-Check, kein Runtime-Smoke.
- `R-13`: Kein Distroless-Umstieg allein aufgrund der bestehenden
  Ignore-Liste. Der naechste harte Schritt ist ein reproduzierbares
  Re-Review-Artefakt mit Trivy-Version, Scan-Kontext, bekannten CVEs,
  `expires`-Stand und Entscheidung `continued` vs. `migrated`.
- `R-12`: Kein Release-Gate-Ausbau ohne neue Browser-Pflicht. Der
  bestehende Nightly-Drift-Smoke bleibt der Default; Safari/WebKit wird
  erst verpflichtend, wenn ein Support- oder Operator-Trigger vorliegt.

Damit ist Tranche 2 standardmaessig dokumentarisch. Code-Scope entsteht
nur bei folgenden Nachweisen:

| Risiko | Code-Trigger |
| --- | --- |
| `R-9` | K8s-Smoke soll PR-/Release-Gate werden oder K8s-Observability-Manifeste fuehren konkrete Scrape-/Label-Policy ein. |
| `R-13` | Trivy-Re-Review zeigt verfuegbare Fixes, `expires` ist erreicht, oder Distroless wird als expliziter Pre-1.0-Basisentscheid freigegeben. |
| `R-12` | Safari/WebKit wird verbindlicher Zielbrowser oder ein Nightly-Drift-Befund verlangt Allowlist-/Gate-Aenderungen. |

### 2.1 Entscheidungsvorlage (zu finalisieren)

| Risiko | Baseline-Status | Kandidat A | Kandidat B | Vorauswahl |
| --- | --- | --- | --- | --- |
| R-9 | offen (`⬜`) | K8s-Observability bleibt deferred, keine neuen K8s-Smokes | eigener K8s-Label-Allowlist-Modus + Smoke-Profiltrennung | A: `deferred`, solange kein K8s-Smoke-/Scrape-Trigger belegt ist |
| R-13 | offen (`⬜`) | Trivy-Ignores mit strikter Re-Review-Fortführung | Distroless-Umstieg evaluieren und ggf. einführen | A: `continued`, aber nur mit reproduzierbarem Re-Review-Artefakt |
| R-12 | offen (`⬜`) | Status-quo fortführen (Nightly + optional Safari) | `make smoke-webrtc-stats-drift` in Release-Gates integrieren | A: `deferred`, solange Safari/WebKit nicht verpflichtend ist und Nightly gruen bleibt |

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

- [x] `R-9` bleibt offen (`⬜`) mit explizitem Triggertext
  (z. B. „bei K8s-Smoke-Gate“), ohne dass K8s-spezifische
  produktive Änderungen eingeführt wurden.

#### R-13 (CVE-Re-Review)

Implementiert:

- [ ] Nachweis, dass mindestens eine Maßnahme entschieden und ausgeführt
  wurde: neuer Trivy-Durchlauf, Ignored-Update mit nachvollziehbaren Gründen
  **oder** Distroless-Basis-Entscheid inkl. Build-/Runtime-Nachweis.
- [ ] Ein reproduzierbares Re-Review-Artefakt für `R-13` ist abgelegt
  (Scan-Output, `expires`-Werte, Trivy-Version, Commit/PR-Referenz) und
  im Plan verlinkt.
- [ ] `R-13` ist in
  `docs/planning/in-progress/risks-backlog.md` mit `🟢` oder explizitem
  `⬛` und Abschlussreferenz dokumentiert.

Defer:

- [x] `R-13` bleibt offen (`⬜`) mit exakt dokumentiertem Triggertext
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

- [x] `R-12` bleibt offen (`⬜`) mit explizitem Triggertext
  (z. B. Safari als verpflichtender Zielbrowser / Supportanforderung),
  inklusive grüner Drift-Nachtschicht-Historie.

## 3. Tranche 1 — Trigger-Re-Eval

### 3.1 R-9: K8s-Observability-Trigger

- [x] Prüfen, ob K8s-Smoke in der Roadmap als Gate geplant ist.
- [x] Prüfen, ob Compose- und K8s-Layers sauber trennbar sind
  (`smoke scope`, Labels, README).
- [x] Entscheidung dokumentieren:
  - `deferred`: kein weiteres Artefakt, Offene-Punkt bleibt im Backlog
  - `implemented`: separater K8s-Allowlist-Modus + Smoke-Profiltrennung

Ergebnis 2026-05-13: `deferred`. In der Roadmap ist kein K8s-Smoke als
PR-/Release-Gate geplant; `deploy/k8s/` bleibt ein optionaler,
clusterfreier Seed-Pfad. `make k8s-validate` fand eine bestehende
Version-Drift in den Beispiel-Images (`0.14.0` statt `0.17.0`), die als
Seed-/Release-Drift korrigiert wurde. Das ist kein neuer
K8s-Observability- oder Runtime-Smoke-Scope. Nach dem Fix ist
`make k8s-validate` gruen.

### 3.2 R-13: CVE-Re-Review

- [x] Prüfen, ob Upstream-Fixes für die drei CVEs verfügbar sind.
- [x] Prüfen, ob die aktuelle `expires`-Schwelle erreicht ist.
- [x] Entscheidung dokumentieren:
  - `continued`: Ignore erneuern + harte Re-Review-Logik dokumentieren
  - `migrated`: distroless als produktive Basis-Entscheidung vorbereiten

Ergebnis 2026-05-13: `continued`. Das aktuelle `expires` ist
`2026-11-02` und damit nicht erreicht. `make image-scan` ist gruen:
`mtrace-api:scan` hat 0 HIGH/CRITICAL ohne Ignores; Dashboard und
Analyzer-Service rendern jeweils die drei bekannten CVE-Ignores und
haben keine unignorierten HIGH-/CRITICAL-Findings. Debian Security
Tracker zeigt fuer `trixie` weiterhin `vulnerable`/`no-dsa` fuer alle
drei CVEs; Fixes liegen nicht in der Trixie-Slim-Basis. Re-Review-Artefakt:
[`../in-progress/r13-trivy-rereview-2026-05-13.md`](../in-progress/r13-trivy-rereview-2026-05-13.md).

### 3.3 R-12: WebRTC-Drift-Kontur

- [x] Prüfen, ob Nightly-Drift-Smoke stabil grün ist.
- [x] Prüfen, ob Safari/WebKit als verbindlicher Produktiv- oder
  Support-Trigger aufgenommen wurde.
- [x] Entscheidung dokumentieren:
  - `deferred`: weiterhin „automatisch detektiert“
  - `migrated`: Browserumfang und Release-Kontrolle erweitern

Ergebnis 2026-05-13: `deferred`. Die letzten drei `webrtc-drift.yml`-
Runs sind gruen; der neueste gepruefte Lauf ist `25769902117`
(`2026-05-13T00:13:10Z`, Head
`2f75331bc5bd1cee37983972b919c98444770d1d`). Es wurde kein neuer
Safari-/WebKit-Pflicht- oder Operator-Support-Trigger gefunden; der
Nightly-Detector bleibt der Reaktionspfad.

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

- [x] `make docs-check` grün.
- [x] Relevante Tranche-2-Artefakte grün:
  - R-12: letzter Nightly-Drift-Run geprueft, keine lokale
    Gate-Migration ohne Safari-/Support-Trigger.
  - R-9: `make k8s-validate` gruen nach K8s-Seed-Version-Fix.
  - R-13: `make image-scan` gruen; Re-Review-Artefakt (`trivy`-Scan,
    ignore-lifecycle, `expires`-Historie) liegt reproduzierbar vor.
- [x] `docs/planning/in-progress/risks-backlog.md` auf
  `0.18.0`-Konsequenz aktualisieren.

## 6. Tranche 4 — Closeout

- [ ] Für `R-9`, `R-12`, `R-13` einen Abschlussstatus setzen
  (`🟢`/`⬜`/`⬛`) und Begründung dokumentieren.
- [ ] Bei vollständiger Umsetzung: Plan nach
  `docs/planning/done/plan-0.18.0.md` verschieben.
- [ ] Bei Restoffen: Auslöser/Resttrigger präzisieren und Weiterführung
  explizit in Backlog + Roadmap verankern.
