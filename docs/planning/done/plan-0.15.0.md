# Implementation Plan — `0.15.0` (Product Scope / Analyzer Boundary)

> **Status**: ✅ released 2026-05-12 — Tranchen 0–5
> geschlossen; Release-Tag `v0.15.0`.
>
> **Vorgänger**: `0.14.0` (Ops Backend Follow-up), released
> 2026-05-12; Plan in
> [`../done/plan-0.14.0.md`](../done/plan-0.14.0.md).
>
> **Release-Typ**: Minor-Release mit Lastenheft-Patch `1.1.20`,
> neuer RAK-Gruppe `RAK-101`..`RAK-105` und Tag `v0.15.0`.
>
> **Ziel**: `0.15.0` trifft die naechsten Produkt- und
> Architekturgrenzen, bevor neue grosse Backend- oder Plattformpfade
> implementiert werden. Der Release soll Zielgruppe, externe
> Analyzer-Grenze, Control-Plane-Scope, naechsten Analyzer-Folge-Slice
> und Ops-Backend-Trigger zusammenfuehren.
>
> **Bezug**:
> [`../in-progress/roadmap.md`](../in-progress/roadmap.md),
> [`../in-progress/risks-backlog.md`](../in-progress/risks-backlog.md),
> [`../../../spec/lastenheft.md`](../../../spec/lastenheft.md)
> §16.1, `MVP-20`, `F-132`, `NF-13`, `MVP-40`, `MVP-41`.
>
> **Nachfolger**: vorbereiteter Folgeplan
> [`../open/plan-0.16.0.md`](../open/plan-0.16.0.md).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR- oder Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.15.0` ist ein Product-Scope- und Boundary-Decision-Release. Er
liefert bewusst keine neue Pflichtplattform, keinen Postgres-Default
und kein Hochvolumen-Analytics-Backend. Der Plan bereitet nur die
naechsten sinnvollen Implementierungsslices vor.

Vorlaeufig in Scope:

- Zielgruppenentscheidung aus Lastenheft §16.1 formalisieren:
  Selbsthoster/kleine Teams/Broadcaster-Labs bleiben Primaerziel oder
  ein Plattform-Betreiber-Pfad wird explizit als spaeterer Scope
  vorbereitet.
- `MVP-20` bewerten: externe `apps/analyzer-api` gegen bestehenden
  internen `apps/analyzer-service` abgrenzen, inklusive Job-Modell,
  Auth, Rate Limits und Ergebnisabruf als Entscheidungsoptionen.
- `F-132` (`apps/control-plane`) nur als Decision-Record zuschneiden:
  Aufgaben, Nicht-Ziele, Trigger, Abhaengigkeiten und minimale
  RACI/Owner klaeren, aber keine Control-Plane bauen.
- `NF-13` Folge-Scope zuschneiden: Low-Latency-CMAF, HTTP-Range/
  Byte-Range, vollstaendigere Segmentabdeckung oder Codec-Decoding
  gegeneinander bewerten und hoechstens einen kleinen Folge-Slice
  fuer einen spaeteren Plan empfehlen.
- `MVP-40`/`MVP-41` Trigger-Re-Eval nach `0.14.0`: pruefen, ob
  Postgres- oder Analytics-Trigger tatsaechlich erreicht wurden.

Vorlaeufig out of scope:

- Keine Implementierung von `apps/control-plane`.
- Keine neue externe `apps/analyzer-api`, solange Tranche 2 keinen
  klaren Go-Scope liefert.
- Kein Postgres-Runtime-Adapter, kein Postgres-Default und keine
  automatische SQLite-zu-Postgres-Migration ohne ausgelösten Trigger.
- Kein ClickHouse-/VictoriaMetrics-/Mimir-Pflichtbackend und kein POC
  ohne konkreten Workload-Trigger.
- Kein vollstaendiger Production-Kubernetes-Betrieb.
- Kein Multi-Tenant-SaaS-Produkt.
- Kein OAuth/OIDC/SSO- oder User-/Org-Verwaltungs-Ausbau ohne
  ausdrueckliche Zielgruppenentscheidung.

### 0.2 Vorgänger-Gate

Vor Aktivierung von `0.15.0` muessen diese Bedingungen erfuellt sein:

- [x] `0.14.0` ist released und als
  `docs/planning/done/plan-0.14.0.md` archiviert.
- [x] Roadmap zeigt `0.15.0` als aktive Folgephase oder begruendet
  einen anderen Nachfolger.
- [x] RAK-96..RAK-100 sind in der `0.14.0`-Verifikationsmatrix
  geschlossen oder mit explizitem Defer-/Blockerstatus versehen.
- [x] `0.14.0` hat bestaetigt, dass K8s-/Devcontainer-/Release-Guard-
  Hardening abgeschlossen oder bewusst deferred ist.
- [x] Keine Postgres-/Analytics-/K8s-Trigger wurden stillschweigend
  als Implementierungsfreigabe interpretiert.

Uebergangsausnahme bei Release-Freeze oder blockiertem Vorgaenger:
Tranche 0 darf `0.15.0` nur als **draft/in-progress Planning** starten,
wenn Roadmap und Blocker-Log den fehlenden `0.14.0`-Closeout
begründen, alle aus `0.14.0` importierten RAKs als `[!]` blockiert
markiert sind und kein Tag/Release fuer `0.15.0` erstellt wird, bevor
`0.14.0` archiviert oder ausdruecklich durch einen neuen Plan ersetzt
ist.

### 0.3 Lastenheft-Patch `1.1.20`

Die RAK-Gruppe ist bei Aktivierung final vergeben:

| Kennung | Thema | Bedingung |
| --- | --- | --- |
| RAK-101 | Zielgruppenentscheidung | Immer aktiv; Lastenheft §16.1 muss geschlossen oder bewusst als Produkt-ADR deferred werden. |
| RAK-102 | Analyzer-API-Boundary (`MVP-20`) | Bewertung externe API vs. interner Service, mit Go/No-Go und Triggern. |
| RAK-103 | Control-Plane-Scope (`F-132`) | Nur Decision-Record; keine Implementierung ohne spaeteren Plan. |
| RAK-104 | Analyzer-Folge-Slice (`NF-13`) | Auswahl oder Defer eines kleinen naechsten Analyzer-Slices. |
| RAK-105 | Ops-Trigger-Re-Eval (`MVP-40`/`MVP-41`) | Postgres-/Analytics-Trigger pruefen und sichtbar deferred oder aktiviert halten. |

### 0.4 Qualitätsregeln für `0.15.0`

- Jede Entscheidung enthaelt mindestens `Entscheidung`, `Begruendung`,
  `Nicht entschieden`, `Trigger` und `Naechster Planpfad`.
- Keine neue Plattform-App wird gebaut, bevor Zielgruppe und
  Betriebsmodell entschieden sind.
- Analyzer-Scope muss zwischen internem Service, externer API und
  Library/CLI sauber trennen.
- Neue Backends bleiben opt-in und duerfen lokale Entwicklung nicht
  belasten.
- Decision-only-Tranchen liefern trotzdem pruefbare Artefakte:
  Plan-Notiz, ADR, Lastenheft-Patch, Roadmap-Update oder
  Risks-Backlog-Update.
- Jede aktive Tranche endet mit einem `What aendert sich` /
  `What bleibt unveraendert`-Block.

### 0.5 Aktivierungsszenarien

Tranche 0 waehlt genau eines dieser Aktivierungsszenarien:

| Szenario | Inhalt | Release-Charakter | Go-Bedingung |
| --- | --- | --- | --- |
| A | Zielgruppe + Analyzer-Boundary | fokussierter Product-/Analyzer-Decision-Release | `0.14.0` schliesst Ops-Hardening ohne neue Backend-Trigger |
| B | Zielgruppe + Control-Plane-Decision | Product-/Platform-Decision-Release | Stakeholder- oder Operator-Bedarf fuer Admin-/Multi-Project-Scope liegt vor |
| C | Analyzer-Folge-Slice + Zielgruppe | kleiner Analyzer-Feature-Prep-Release | `NF-13`-Folgeslice ist klar begrenzt und ohne neue Plattformabhaengigkeit |
| D | Trigger-/Defer-Release | rein dokumentarisch | Alle Implementierungsoptionen bleiben ohne Trigger |

Mehr als ein grosser Implementierungspfad in `0.15.0` ist No-Go,
ausser Tranche 0 dokumentiert explizit Ressourcen, Risikoentkopplung
und getrennte Gates.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang | Status |
| --- | --- | --- | --- | --- | --- |
| 0 | Aktivierung, RAK-Zuschnitt und `0.14.0`-Closeout-Import | Finaler `0.15.0`-Scope | `0.14.0` released | Szenario A | ✅ |
| 1 | Zielgruppenentscheidung | Primaerziel und spaetere Plattformgrenze entschieden | Lastenheft §16.1 | Product-Decision-Notiz in Plan + Lastenheft | ✅ |
| 2 | Analyzer-API-Boundary (`MVP-20`) | Externe Analyzer-API: Go, Defer oder Triggerpflege | `apps/analyzer-service` Bestand | Defer-Decision + Trigger | ✅ |
| 3 | Control-Plane-Scope (`F-132`) | Control-Plane bleibt deferred oder erhaelt klaren spaeteren Planpfad | Zielgruppenentscheidung | Defer-Decision + Trigger | ✅ |
| 4 | Analyzer-Folge-Slice (`NF-13`) | Naechster Analyzer-Slice gewaehlt oder bewusst deferred | CMAF-Folge-Scope | Slice-Zuschnitt: HTTP-Range-/Byte-Range-Loader | ✅ |
| 5 | Ops-Trigger-Re-Eval und Closeout | Postgres/Analytics-Triggerstatus, RAK-Matrix, Release | `MVP-40`/`MVP-41` Trigger | Tag `v0.15.0` | ✅ |

## 2. Tranche 0 — Aktivierung und Scope-Härtung

Ziel: `0.15.0` wird nach dem `0.14.0`-Closeout aktiviert, ohne die
gerade gehaerteten Ops-Pfade erneut als Implementierungsvorwand zu
nutzen.

DoD:

- [x] Plan von `docs/planning/open/plan-0.15.0.md` nach
  `docs/planning/in-progress/plan-0.15.0.md` verschoben.
- [x] Ausgangszustand von `git status --short` dokumentiert.
- [x] `0.14.0`-Closeout gelesen und Folgepunkte explizit importiert.
- [x] Aktivierungsszenario A/B/C/D ausgewaehlt und begruendet.
- [x] Lastenheft-Patch mit finaler RAK-Range ergaenzt.
- [x] Roadmap auf `0.15.0` als aktive Folgephase umgestellt.
- [x] Risiken-Backlog aktualisiert, falls neue Trigger oder Risiken
  entstehen.
- [x] Aktivierungsnotiz enthaelt `What aendert sich` /
  `What bleibt unveraendert` und benennt nicht aktivierte Pfade.
- [x] No-Go-Liste geprueft:
  - Control-Plane-Implementierung ohne Zielgruppenentscheidung,
  - externe Analyzer-API ohne klaren Konsumenten,
  - Postgres/Analytics-Implementation ohne Trigger,
  - Multi-Tenant-/OAuth-Scope ohne Produktentscheidung.

### 2.1 Aktivierungsnotiz

| Feld | Wert |
| --- | --- |
| Aktivierungsdatum | 2026-05-12 |
| Ausgangs-Commit | `956c82a` (`v0.14.0`, `origin/main`) |
| Ausgangszustand | `git status --short --branch`: `## main...origin/main` |
| Gewaehltes Szenario | **A — Zielgruppe + Analyzer-Boundary** |
| Begruendung | `0.14.0` hat Ops-Hardening abgeschlossen, ohne Postgres-, Analytics- oder K8s-Implementierungstrigger auszulösen. Ein Control-Plane-Go liegt nicht vor; ein `NF-13`-Folgeslice braucht zuerst die Analyzer-Boundary. |
| Uebernommene 0.14-Folgepunkte | Postgres bleibt `defer-with-migration-seed`; Analytics bleibt `defer`; K8s bleibt optional und clusterfrei validiert; Devcontainer ist Zusatzpfad; Release-Guard bleibt freigabepflichtig. |
| Explizit deferred | Control-Plane-Implementierung, externe Analyzer-API-Implementierung, Postgres-Runtime-/Default-Pfad, Analytics-Backend-POC, Production-K8s, OAuth/OIDC/SSO, Multi-Tenant-SaaS. |
| Blocker | Keine Blocker fuer Tranche 0. Tranche 1 muss Zielgruppe entscheiden; Tranche 2 bleibt ohne externen Analyzer-Konsumenten ein Defer-Kandidat; Tranche 5 aktiviert Postgres/Analytics nur bei echten Triggern. |
| Required Gates | Tranche 0 ist Doku-only: `make docs-check`; Build-/Security-Gates erst bei Code-, Wire-, Script- oder Release-Aenderungen. |

### 2.2 Aktivierungsentscheid

**What aendert sich**

- `0.15.0` ist aktive Folgephase in `docs/planning/in-progress/`.
- Lastenheft-Patch `1.1.20` reserviert `RAK-101`..`RAK-105`.
- Der Release startet mit Szenario A: Zielgruppe und Analyzer-Boundary
  werden vor Control-Plane-, Backend- oder Plattform-Implementierungen
  entschieden.
- Roadmap und Risiken-Backlog verweisen auf den aktiven Plan und
  markieren Postgres/Analytics/K8s weiterhin als Trigger-/Defer-Pfade.

**What bleibt unveraendert**

- SQLite und Compose bleiben Standardpfade.
- `apps/analyzer-service` bleibt der bestehende interne
  Analyzer-HTTP-Wrapper; eine externe `apps/analyzer-api` entsteht
  nicht durch Tranche 0.
- `apps/control-plane`, OAuth/OIDC/SSO, Multi-Tenant-Betrieb,
  Postgres-Default, Analytics-Pflichtbackend und Production-K8s
  bleiben ohne eigenen Folgeplan out of scope.
- Fuer `0.15.0` wird noch kein Tag und kein GitHub-Release erstellt.

## 3. Tranche 1 — Zielgruppenentscheidung

Ziel: Die bisher offene Produktfrage aus Lastenheft §16.1 wird so weit
entschieden, dass Folgeplaene nicht gleichzeitig Selbsthoster-Lab,
kleine Teams und grosse Plattformbetreiber bedienen muessen.

DoD:

- [x] Primaerziel fuer die naechsten Minor-Releases entschieden.
- [x] Plattform-Betreiber-Scope entweder deferred oder als eigener
  spaeterer Planpfad mit Triggern beschrieben.
- [x] Auswirkungen auf Storage, Sampling, Cardinality, Multi-Tenant,
  Betriebsmodell, Dashboard-Komplexitaet und Alerting dokumentiert.
- [x] Lastenheft §16.1 aktualisiert oder durch ADR/Plan-Decision
  referenziert.
- [x] Out-of-Scope-Liste fuer `0.15.0` nachgezogen.
- [x] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Go/No-Go:

- **Go:** klare Primaerzielgruppe, explizite Nicht-Ziele, kein
  impliziter Plattform-Scope.
- **No-Go:** gleichzeitige Optimierung fuer Solo-Selbsthoster und
  hunderte parallele Plattformstreams ohne Ressourcenschnitt.

Vorlaeufige Artefakte:

- Product-Scope-Decision in diesem Plan oder `docs/adr/`.
- Roadmap-Update mit Zielgruppenstand.
- Optionaler Lastenheft-Patch fuer §16.1.

### 3.1 Product-Scope-Decision

| Feld | Entscheidung |
| --- | --- |
| Datum | 2026-05-12 |
| Entscheidung | Für die nächsten Minor-Releases bleibt m-trace auf Selbsthoster, kleine bis mittlere Streaming-Teams, Broadcaster-Labs und technische Media-/DevOps-Teams ausgerichtet. |
| Begründung | Diese Zielgruppe passt zum bestehenden lokalen Compose-/SQLite-/OpenTelemetry-Scope, zum internen Analyzer-Service und zu den gelieferten Lab-Smokes. Große Plattformbetreiber mit hunderten parallelen Streams würden sofort andere Annahmen für Multi-Tenant, Storage, Analytics, Control-Plane und Betriebs-SLOs erzwingen. |
| Nicht entschieden | Kein SaaS-Produkt, kein generisches Multi-Tenant-Control-Plane-Modell, keine Operator-UI für hunderte Projekte, kein verbindlicher Postgres-/Analytics-/K8s-Production-Pfad. |
| Trigger für Plattform-Betreiber-Scope | Konkreter Stakeholder-/Operator-Bedarf mit mindestens einem benannten Betreiberprofil, erwarteter Stream-/Event-Größenordnung, Multi-Tenant-/Auth-Anforderungen, Betriebs-SLO, Owner und Folgeplan. |
| Naechster Planpfad | Tranche 2 bewertet die Analyzer-API-Boundary unter dieser Zielgruppe; Tranche 3 beschreibt Control-Plane nur als deferred oder bei späterem Trigger als eigenen Folgepfad. |

### 3.2 Auswirkungen

| Bereich | Entscheidung / Auswirkung |
| --- | --- |
| Storage | SQLite bleibt lokaler Standard. Postgres bleibt Trigger-/Folgepfad aus ADR 0005 und wird nicht wegen hypothetischer Plattformbetreiber aktiviert. |
| Sampling | Sampling bleibt auf technische Diagnose und kontrollierte Cardinality ausgelegt; keine Plattform-weite Billing-/Audience-Analytics-Semantik. |
| Cardinality | Prometheus bleibt aggregiert und ohne Session-/Viewer-Labels; hochkardinale Plattformmetriken bleiben out of scope. |
| Multi-Tenant | Project-/Token-Grenzen bleiben technische Isolation für Lab- und kleine Team-Setups, kein SaaS-Tenant-Modell mit User-/Org-Verwaltung. |
| Betriebsmodell | Docker Compose und opt-in Beispielpfade bleiben Default. K8s bleibt Beispiel-/Triggerpfad, nicht Production-Betriebsversprechen. |
| Dashboard-Komplexitaet | Dashboard bleibt Diagnose- und Lab-Oberfläche, keine Admin-/Billing-/Fleet-Management-Control-Plane. |
| Alerting | Alerts bleiben technisch und operatornah; keine kunden-/mandantenbezogenen Alerting-Workflows. |

### 3.3 What aendert sich

- Lastenheft §16.1 ist nicht mehr offen: Die Primärzielgruppe ist
  entschieden und als Patch-`1.1.20`-Decision dokumentiert.
- Tranche 2/3 dürfen externe Analyzer-API oder Control-Plane nicht
  aus großem Plattform-Scope ableiten, sondern brauchen konkrete
  Konsumenten, Trigger und Folgepläne.
- Roadmap und RAK-Matrix markieren `RAK-101` als erledigt.

### 3.4 What bleibt unveraendert

- Die bestehenden Zielgruppen in Lastenheft §5 bleiben kompatibel;
  `0.15.0` schärft nur das Primärziel für die nächsten Minor-Releases.
- SQLite, Compose, interne Analyzer-Service-Nutzung,
  aggregierte Metriken und lokale Lab-Smokes bleiben Standard.
- Große Plattformbetreiber bleiben späterer Scope und werden nicht
  still über Storage-, Analytics-, K8s- oder Control-Plane-Pfade
  in `0.15.0` hineingezogen.

## 4. Tranche 2 — Analyzer-API-Boundary (`MVP-20`)

Ziel: Die historische externe `apps/analyzer-api`-Anforderung wird
gegen den realen internen `apps/analyzer-service` bewertet.

DoD:

- [x] Bestehender interner Analyzer-Service-Pfad beschrieben:
  Verantwortlichkeiten, Grenzen, Verbraucher, Failure-Modes.
- [x] Externe Analyzer-API-Optionen bewertet:
  synchroner HTTP-Pfad, async Job API, nur Library/CLI, oder Defer.
- [x] Auth, Rate Limits, SSRF-Schutz, Ergebnisabruf, Retention und
  Contract-Fixtures als Pflichtfragen fuer eine externe API
  dokumentiert.
- [x] Entscheidung `proceed`, `POC`, `defer` oder `drop/anders
  erfuellt` getroffen.
- [x] Wenn `proceed` oder `POC`: minimaler Folgeslice ohne Control-
  Plane-Abhaengigkeit definiert.
- [x] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Go/No-Go:

- **Go:** konkreter externer Konsument, klare API-Grenze, keine
  Dopplung zum internen Service.
- **No-Go:** API nur aus historischer `MVP-20`-Formulierung bauen,
  ohne Nutzer oder Betriebsmodell.

Vorlaeufige Artefakte:

- Boundary-Decision-Notiz.
- Optionaler ADR fuer externe Analyzer-API.
- Triggerliste fuer spaeteren Implementierungsplan.

### 4.1 Bestehender interner Analyzer-Pfad

| Bereich | Stand |
| --- | --- |
| Verantwortlichkeit | `apps/api` exposes `POST /api/analyze` als m-trace API-Surface und delegiert über den Driven-Port `StreamAnalyzer` an den internen `apps/analyzer-service`. |
| Analyzer-Service | `apps/analyzer-service` ist ein interner Node-HTTP-Wrapper um `@npm9912/stream-analyzer` mit `POST /analyze` und `GET /health`; Compose verbindet `apps/api` über `ANALYZER_BASE_URL=http://analyzer-service:7000`. |
| Library/CLI | `packages/stream-analyzer` liefert Public API und CLI (`pnpm m-trace check`) für direkte technische Nutzung ohne m-trace API-Surface. |
| Verbraucher | Heute: `apps/api` und Operator/Entwickler über Library/CLI. Kein dokumentierter externer API-Konsument außerhalb des m-trace Backend-Pfads. |
| Grenzen | Der interne Service ist nicht als öffentliches Produkt-API gedacht; externe Exposure braucht eigene Auth-, Rate-Limit-, SSRF-, Retention- und Contract-Entscheidung. |
| Failure-Modes | API-Validation liefert 400/413/415; Analyzer-Domain-Fehler liefern strukturierte Codes wie `invalid_input`, `fetch_blocked`, `manifest_not_hls`, `fetch_failed`, `manifest_too_large`; Transport-/Verfügbarkeitsfehler werden als `analyzer_unavailable` gemappt. |
| Bestehende Schutzgrenzen | API-Body-Limit 1 MiB, Analyzer-Adapter-Timeout 30 s, Antwortlimit 4 MiB, Loader-SSRF-Schutz in `@npm9912/stream-analyzer`, Contract-Fixtures unter `spec/contract-fixtures/analyzer/`. |

### 4.2 Optionen

| Option | Bewertung | Entscheidung |
| --- | --- | --- |
| Synchroner externer HTTP-Pfad | Würde die heutige interne API duplizieren und sofort Public-Auth, Rate Limits, Abuse-Schutz, Versionierung und Betriebs-SLOs erzwingen. Kein externer Konsument belegt. | `defer` |
| Async Job API | Sinnvoll erst bei langen Analysen, Queue/Retention, Ergebnisabruf, Cancellation, Quotas und eigener Persistenz. Diese Voraussetzungen liegen nicht vor und würden Control-Plane-/Storage-Scope vorziehen. | `defer` |
| Nur Library/CLI plus interner Service | Passt zur entschiedenen Zielgruppe: technische Teams können die Library/CLI direkt nutzen, m-trace API bleibt der integrierte Produktpfad. | `keep` |
| Externe API komplett streichen | Zu hart: spätere externe Konsumenten oder schwere Analysejobs können einen POC rechtfertigen. | `no` |

### 4.3 Pflichtfragen fuer eine spaetere externe Analyzer-API

| Thema | Mindestentscheidung vor `proceed` oder `POC` |
| --- | --- |
| Auth | Token-/Project-Scope, optional Session-Bindung, klare Fehlerpräzedenz und keine anonyme öffentliche URL-Fetch-API. |
| Rate Limits | Origin-/Project-Quotas, URL-Fetch-Budget, Concurrency-Limit und Fail-Mode. |
| SSRF-Schutz | Bestehende Loader-Sperren bleiben Mindestniveau; zusätzlich Egress-Policy, DNS-Rebinding-Grenzen, Redirect-Policy und private Netzwerkfreigabe nur explizit. |
| Ergebnisabruf | Synchrones Resultat vs. Job-ID, Pagination/Download, Error-Retention und stabile Response-Shapes. |
| Retention | Keine implizite Speicherung; falls Jobs persistiert werden, braucht es TTL, Löschpfad, Project-Scope und Datenschutzgrenzen. |
| Contract-Fixtures | Public API braucht versionierte Request-/Response-Fixtures, Error-Fixtures, Backwards-Compat-Regeln und CI-Drift-Gate. |
| Betriebsmodell | Owner, SLO, Observability, Abuse-Runbook, Ressourcenbudget und Abbruchkriterien. |

### 4.4 Boundary-Decision

| Feld | Entscheidung |
| --- | --- |
| Datum | 2026-05-12 |
| Entscheidung | `MVP-20` bleibt für eine eigenständige externe `apps/analyzer-api` **deferred**. Der bestehende interne `apps/analyzer-service` plus Library/CLI erfüllt den aktuellen Zielgruppen-Scope anders. |
| Begründung | Nach Tranche 1 ist der Primärscope technisch/labnah. Es gibt keinen externen Analyzer-Konsumenten, keine Job-Retention-Anforderung und keinen Betreiberbedarf, der eine öffentliche API rechtfertigt. Eine externe API würde heute mehr Plattform-, Auth- und Betriebs-Scope einführen als sie Produktnutzen liefert. |
| Nicht entschieden | Kein Public Analyzer API Contract, keine async Job Queue, keine Ergebnis-Retention, keine dedizierte Analyzer-Auth-Schicht, keine Control-Plane-Abhängigkeit. |
| Reaktivierungs-Trigger | Konkreter externer Konsument, Analysen mit Laufzeit/Größe jenseits synchroner API-Grenzen, Bedarf an isolierter Fetch-Sandbox, oder Operator-Wunsch nach Job-/Batch-Workflow mit Owner und SLO. |
| Naechster Planpfad | Falls ein Trigger eintritt: eigener Folgeplan oder `0.16.0`-Szenario A mit POC-Scope. Ohne Trigger bleibt Tranche 4 auf internen Analyzer-Folge-Slices (`NF-13`) ohne externe API-Kopplung. |

### 4.5 What aendert sich

- `RAK-102` ist entschieden: externe `apps/analyzer-api` wird nicht
  in `0.15.0` implementiert.
- Lastenheft `MVP-20` wird auf den Stand geschärft:
  interner `apps/analyzer-service` plus Library/CLI erfüllt den
  aktuellen Scope anders; externe API bleibt triggerbasierter
  Folge-Scope.
- Tranche 4 darf Analyzer-Folge-Slices nur gegen Library, CLI oder
  internen Service planen, solange kein externer API-Trigger eintritt.

### 4.6 What bleibt unveraendert

- `POST /api/analyze` bleibt die m-trace API-Surface und delegiert
  intern an `apps/analyzer-service`, wenn `ANALYZER_BASE_URL` gesetzt
  ist.
- `@npm9912/stream-analyzer` bleibt die wiederverwendbare Library/CLI
  für technische Nutzer.
- Bestehende Contract-Fixtures und SSRF-/Timeout-/Größenlimits bleiben
  Mindestschutz; sie werden nicht als Freigabe für eine öffentliche
  Fetch-API interpretiert.
- Keine neue App, Queue, Persistenz, Auth-Schicht oder Control-Plane
  entsteht durch Tranche 2.

## 5. Tranche 3 — Control-Plane-Scope (`F-132`)

Ziel: `apps/control-plane` bekommt eine belastbare Kennung und
Entscheidungsgrenze, bleibt aber ohne gesonderten Folgeplan
unimplementiert.

DoD:

- [x] `F-132`-Scope aus Lastenheft §7.5.6 importiert.
- [x] Aufgaben, Nicht-Ziele und Abhaengigkeiten dokumentiert:
  Multi-Project, Teams, API-Keys, Audit, Admin-UI, User-/Org-Modell,
  OAuth/OIDC, Betriebsprofile.
- [x] Trigger definiert, wann eine Control-Plane geplant werden darf.
- [x] Abhaengigkeit zur Zielgruppenentscheidung aus Tranche 1
  explizit beschrieben.
- [x] Entscheidung `defer`, `POC` oder `proceed in later plan`
  getroffen.
- [x] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Go/No-Go:

- **Go:** Control-Plane bleibt klarer spaeterer Planpfad mit
  Triggern und Nicht-Zielen.
- **No-Go:** UI/API-Geruest bauen, bevor Zielgruppe, Auth-Modell und
  Betriebsmodell entschieden sind.

Vorlaeufige Artefakte:

- `F-132` Decision-Record.
- Roadmap- oder Risks-Backlog-Update.
- Optionaler ADR-Draft fuer Plattform-/Admin-Scope.

### 5.1 `F-132` Scope-Import

| Bereich | Stand aus Lastenheft §7.5.6 |
| --- | --- |
| Charakter | `apps/control-plane` ist eine spätere Verwaltungsanwendung für produktionsnahe m-trace-Installationen. |
| Status | Nicht Bestandteil des MVP; nur vorbereitet. |
| Potenzielle Aufgaben | Konfiguration mehrerer m-trace-Instanzen, Media-Server-Verwaltung, Stream-Profile, Teams/Projects, Audit-Log, API-Keys, Integrationen, spätere Benutzerverwaltung. |
| Architektur | `apps/control-plane/` oder getrennt `apps/control-plane-api` + `apps/control-plane-ui`; finale Aufteilung erst bei echten Mehrbenutzer-/Admin-Anforderungen. |
| Bestehende Abgrenzung | `0.11.0` Ingest-Control und `0.12.0` Auth sind lokale/API-nahe Pfade und ausdrücklich keine mandantenfähige Control-Plane. |

### 5.2 Aufgaben, Nicht-Ziele und Abhaengigkeiten

| Thema | Entscheidung / Grenze |
| --- | --- |
| Multi-Project | Heute existieren Project-/Token-Grenzen als technische Isolation. Eine UI/API für viele Projects, Owner-Wechsel oder Fleet-Administration bleibt deferred. |
| Teams | Kein Team-/Mitgliedschaftsmodell in `0.15.0`; spätere Teams brauchen User-/Org- und Rollenmodell. |
| API-Keys | Project-Token-Generationen bleiben der vorhandene Pfad. Key-Management-UI, Delegation, Scopes und Audit-Trail sind Control-Plane-Folge-Scope. |
| Audit | Kein produktiver Audit-Log-Scope. Ein späterer Scope braucht Ereignismodell, Retention, Export und Datenschutzbewertung. |
| Admin-UI | Kein Admin-/Billing-/Fleet-Management-Frontend in `0.15.0`. Dashboard bleibt Diagnose- und Lab-Oberfläche. |
| User-/Org-Modell | RAK-71-Out-of-Scope bleibt normativ: keine User-/Org-Verwaltung, keine Rollen, kein SaaS-Tenant-Modell. |
| OAuth/OIDC/SSO | Bleibt out of scope; ein späterer Control-Plane-Pfad muss Identity-Provider, Token-Lifecycle, Rollen und Threat Model vorab entscheiden. |
| Betriebsprofile | Kein Production-Control-Plane-Betrieb, kein K8s-Operator und kein Managed-Cloud-Versprechen. |

### 5.3 Control-Plane-Decision

| Feld | Entscheidung |
| --- | --- |
| Datum | 2026-05-12 |
| Entscheidung | `F-132` bleibt **deferred**. `0.15.0` baut keine `apps/control-plane` und gibt auch keinen POC frei. |
| Begründung | Tranche 1 entschied den labnahen/technischen Primärscope; Tranche 2 deferred externe Plattform-Surface. Ohne konkreten Betreiberbedarf würde eine Control-Plane User-/Org-/Auth-/Audit-/Multi-Tenant-Scope vorziehen und den aktuellen Diagnosefokus verwässern. |
| Nicht entschieden | Kein User-/Org-/Team-Modell, keine Rollen/RBAC, kein OAuth/OIDC/SSO, keine Admin-UI, kein Audit-Log, keine Billing-/Fleet-Verwaltung, keine Control-Plane-API. |
| Trigger für Reaktivierung | Konkreter Operator-/Stakeholder-Bedarf mit mindestens zwei administrierten m-trace-Instanzen oder Projects, benanntem Betreiberprofil, User-/Org-/Auth-Anforderungen, Audit-/Compliance-Bedarf, Owner, SLO und eigenem Folgeplan. |
| Abhaengigkeiten | Zielgruppenentscheidung aus Tranche 1, RAK-71-Out-of-Scope, Project-/Token-Modell aus `0.12.x`, mögliche Postgres-/Analytics-Trigger aus ADR 0005 und Datenschutz-/Telemetry-Grenzen. |
| Naechster Planpfad | Kein Control-Plane-Pfad in `0.16.0`, solange der Trigger nicht eintritt. Wenn er eintritt: separater Decision-/POC-Plan vor jeder Implementierung. |

### 5.4 What aendert sich

- `RAK-103` ist entschieden: Control-Plane bleibt deferred.
- `F-132` ist nicht nur "später", sondern mit konkreten
  Reaktivierungsbedingungen und Nicht-Zielen versehen.
- `0.16.0` darf keinen Control-Plane-POC importieren, solange kein
  Trigger mit Owner, Auth-/Tenant-Entscheidung und Folgeplan vorliegt.

### 5.5 What bleibt unveraendert

- `apps/dashboard` bleibt Diagnose- und Lab-Oberfläche.
- `apps/api` bleibt der Ort für vorhandene Project-Token-, Session-
  Token- und Ingest-Control-Pfade; diese werden nicht zu einer
  mandantenfähigen Admin-Plattform erweitert.
- OAuth/OIDC/SSO, User-/Org-Verwaltung, Rollenmodell, Audit-Log,
  Admin-UI, Billing, Fleet-Management und Production-Control-Plane
  bleiben out of scope.

## 6. Tranche 4 — Analyzer-Folge-Slice (`NF-13`)

Ziel: Der bewusst ausgegrenzte CMAF-/Analyzer-Folge-Scope wird in
eine priorisierte Reihenfolge gebracht, ohne den Analyzer unbounded
zu erweitern.

DoD:

- [x] Optionen verglichen:
  - Low-Latency-CMAF (`#EXT-X-PART`, chunked CMAF),
  - HTTP-Range-/Byte-Range-Loader,
  - vollstaendigere Segmentset-Abdeckung,
  - Codec-Decoding,
  - Player-SDK-CMAF-Laufzeitpfad.
- [x] Nutzen, Risiko, Testbarkeit, Datenvolumen und Security-Grenzen
  pro Option dokumentiert.
- [x] Hoestens ein kleiner Folgeslice empfohlen oder alles deferred.
- [x] SSRF-/Fetch-Grenzen und Contract-Fixture-Auswirkungen benannt.
- [x] Kein Analyzer-Slice wird an eine externe Analyzer-API gekoppelt,
  solange Tranche 2 diese nicht freigibt.
- [x] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Go/No-Go:

- **Go:** ein eng begrenzter Analyzer-Slice mit Test- und Fixture-Plan.
- **No-Go:** vollstaendige CMAF-Vollanalyse, Codec-Decoding oder
  Player-Laufzeitpfad ohne getrennten Plan.

Vorlaeufige Artefakte:

- Analyzer-Folge-Scope-Matrix.
- POC-/Defer-Notiz.
- Optionaler Planvorschlag fuer `0.16.0` oder spaeter.

### 6.1 Baseline aus `0.10.0`

`NF-13` ist seit `0.10.0` im Stream-Analyzer-Scope erfüllt:
manifestbasierte HLS-/DASH-CMAF-Signale plus begrenzte binäre
CMAF-Konformitätsprüfung ausgewählter Init-/Media-Segmente. Explizit
offen blieben Low-Latency-CMAF, vollständige Segmentset-Abdeckung,
Codec-Decoding, Player-SDK-CMAF-Playback und Byte-Range-/Range-Fetches.
Tranche 4 bewertet nur diese Rest-Slices und koppelt sie nicht an eine
externe `apps/analyzer-api`.

### 6.2 Analyzer-Folge-Scope-Matrix

| Option | Nutzen | Risiko / Datenvolumen | Testbarkeit | Security-Grenze | Entscheidung |
| --- | --- | --- | --- | --- | --- |
| Low-Latency-CMAF (`#EXT-X-PART`, chunked CMAF) | Erhöht Signalqualität für LL-HLS-/CMAF-Streams. | Viele Partial-Segmente, CDN-/Cache-Verhalten, chunked Transfer und Zeitfensterlogik würden Loader und Fixtures deutlich verbreitern. | Nur mit eigenem LL-HLS-Fixture-Stack oder aufgezeichneten Teilsegmenten belastbar. | Mehr Fetches, mehr Redirect-/Timeout-Flächen, potenziell lange Live-Fenster. | `defer` |
| HTTP-Range-/Byte-Range-Loader | Schließt eine konkrete Lücke aus `0.10.0`: `EXT-X-MAP`/`#EXT-X-BYTERANGE` und DASH-Range-Initialisierungen können bounded geprüft werden. | Begrenztes Zusatzrisiko, wenn Range-Anfragen nur auf manifest-referenzierte URLs, kleine Bytebereiche und bestehende Segmentlimits beschränkt bleiben. | Gut mit lokalen Fixtures und Contract-Fixtures testbar; keine echte Live-Infrastruktur nötig. | Bestehende SSRF-/Scheme-/Redirect-/Size-/Timeout-Grenzen bleiben Pflicht; Range-Header muss strikt aus Manifestwerten kommen. | **Empfohlen als einziger Folgeslice** |
| Vollständigere Segmentset-Abdeckung | Würde mehr Media-Segmente prüfen und Fehler später im Stream erkennen. | Schnell unbounded: viele Segmente, hohe Kosten, Retention-/Sampling-Fragen. | Nur mit großem Fixture-Set sinnvoll; hohe CI-Laufzeit. | Fetch-Budget, Sampling-Policy und Abbruchregeln müssten neu entschieden werden. | `defer` |
| Codec-Decoding | Liefert tiefere Validierung von Audio-/Video-Bitstreams. | Bringt Native-/WASM-/FFmpeg-Abhängigkeiten, CPU-Kosten und Sicherheitsfläche. | Schwer deterministisch und plattformübergreifend in CI. | Decoder-Sandbox, Ressourcenlimits und CVE-Pflege nötig. | `defer` |
| Player-SDK-CMAF-Laufzeitpfad | Würde Analyse und Playback näher verbinden. | Vermischt Analyzer- und Player-SDK-Scope; Browser-/hls.js-/MSE-Compat wird eigener Produktpfad. | Browser-E2E nötig, nicht nur Analyzer-Fixtures. | Kein Fetch-Analyzer-Scope mehr; betrifft Runtime, CORS und Playback-Telemetrie. | `defer` |

### 6.3 RAK-104 Decision

| Feld | Entscheidung |
| --- | --- |
| Datum | 2026-05-12 |
| Entscheidung | `RAK-104` empfiehlt als einzigen kleinen Folge-Slice einen **HTTP-Range-/Byte-Range-Loader** für manifest-referenzierte CMAF-Init- und Media-Segmente. |
| Begründung | Der Slice erweitert den bestehenden `0.10.0`-Analyzer gezielt, bleibt innerhalb von Library/CLI und internem Analyzer-Service, ist fixture-basiert testbar und benötigt weder externe Analyzer-API noch Control-Plane, Storage, Codec-Decoder oder Player-Laufzeitpfad. |
| Nicht entschieden | Kein Low-Latency-CMAF, keine vollständige Segmentset-Abdeckung, kein Codec-Decoding, kein Player-SDK-CMAF-Playback, keine neue `apps/analyzer-api`, kein Job-/Retention-Modell. |
| Reaktivierungs-Trigger fuer deferred Optionen | LL-CMAF: konkreter LL-HLS-Operator-Stream und Fixture-Plan. Vollständige Segmentsets: nachgewiesene False-Negatives durch erstes-Segment-Sampling. Codec-Decoding: benannter Decoder-Scope mit Sandbox-/CVE-Plan. Player-Laufzeit: eigener Player-SDK-Plan mit Browser-Matrix. |
| Naechster Planpfad | `0.16.0` darf Szenario B wählen und den Range-/Byte-Range-Slice spezifizieren. Falls Tranche 5 keinen Ops-Trigger findet, ist dieser Slice der bevorzugte Implementierungskandidat fuer `0.16.0`. |

### 6.4 Test-, Fixture- und Fetch-Grenzen

- Der Slice darf nur URLs laden, die aus einem bereits akzeptierten
  Manifest stammen; keine frei eingegebenen Zusatz-URLs.
- Range-Werte muessen aus `EXT-X-MAP`-/`#EXT-X-BYTERANGE`-Attributen
  oder DASH-Range-Feldern stammen und vor dem Request auf nichtnegative,
  endliche Bytebereiche validiert werden.
- Bestehende Loader-Grenzen bleiben Mindestniveau: erlaubte Schemes,
  SSRF-/private-network-Blocklist, Redirect-Policy, Timeout,
  Content-Type-Prüfung und `maxSegmentBytes`/`maxBinarySegments`.
- Contract-Fixtures brauchen mindestens HLS-`EXT-X-MAP` mit
  `BYTERANGE`, HLS-`#EXT-X-BYTERANGE`-Media-Segment, DASH-Range-
  Initialization und Reject-Fixtures fuer malformed/oversized Ranges.
- Kein neuer Top-Level-`analyzerKind`; Ergebnisse bleiben unter
  `details.cmaf.binary` oder einem additiven Unterfeld im bestehenden
  HLS-/DASH-Detailschema.

### 6.5 What aendert sich

- `RAK-104` ist entschieden: Ein kleiner Analyzer-Folge-Slice ist
  priorisiert.
- `0.16.0` bekommt mit Szenario B einen konkreten Kandidaten:
  HTTP-Range-/Byte-Range-Loader fuer CMAF-Init-/Media-Segmente.
- `spec/lastenheft.md` §13.19 vermerkt die Tranche-4-Entscheidung.

### 6.6 What bleibt unveraendert

- Externe `apps/analyzer-api` bleibt deferred; der Slice gehoert zu
  `@npm9912/stream-analyzer`, CLI und internem `apps/analyzer-service`.
- Low-Latency-CMAF, vollständige Segmentset-Abdeckung,
  Codec-Decoding und Player-SDK-CMAF-Laufzeitpfade bleiben deferred.
- Keine Control-Plane-, Storage-, Job-Queue- oder Retention-
  Abhaengigkeit entsteht durch Tranche 4.

## 7. Tranche 5 — Ops-Trigger-Re-Eval und Release-Closeout

Ziel: `MVP-40` und `MVP-41` bleiben sichtbar getrackt, werden aber
nur bei echten Triggern in Implementierung ueberfuehrt.

DoD:

- [x] Postgres-Trigger aus ADR 0005 geprueft:
  Multi-Replica-Store, Recovery-SLO, Retention-/Query-Last.
- [x] Analytics-Trigger geprueft:
  > 50 Mio. Events/Tag, Ad-hoc-Analysebedarf, Owner + Abbruchdatum.
- [x] Falls kein Trigger erreicht ist: Defer-Status mit Datum und
  Owner aktualisiert.
- [x] Falls Trigger erreicht ist: Folgeplan statt stiller Umsetzung
  angelegt oder als Blocker fuer `0.15.0` markiert.
- [x] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.
- [x] RAK-Verifikationsmatrix vollstaendig ausgefuellt.
- [x] `make docs-check` gruen.
- [x] Bei dokumentarischem Szenario ohne Code-/Release-Skript-
  Aenderungen: Code-, Build- und Security-Gates sind als `n/a` mit
  Begruendung im Closeout markiert.
- [x] Bei codebezogenen Aenderungen: `make build` und `make gates`
  gruen.
- [x] Bei codebezogenen/Release-Aenderungen: `make security-gates`
  gruen oder CI-Job `Security gates` gruen dokumentiert.
- [x] Versions-Bump auf `0.15.0` vollstaendig durchgefuehrt.
- [x] `CHANGELOG.md` mit `[0.15.0] - 2026-05-12` aktualisiert.
- [x] Roadmap auf released `0.15.0` und naechste Folgephase
  umgestellt.
- [x] Plan nach `docs/planning/done/plan-0.15.0.md` verschoben,
  Status auf `✅ released`.
- [x] Annotierter Tag `v0.15.0` erstellt.

### 7.1 Ops-Trigger-Re-Eval

| Pfad | Trigger aus ADR 0005 / `0.14.0` | Stand 2026-05-12 | Entscheidung |
| --- | --- | --- | --- |
| Postgres (`MVP-40`) | Multi-Replica-Store, Recovery-SLO, Retention-/Query-Last, automatischer SQLite-Export oder DSN-Pflichtpfad | Kein neuer Betreiberbedarf, kein Multi-Replica-SLO, keine Recovery-/Retention-Anforderung und kein Migrations-/Rollback-Owner dokumentiert. | `defer-with-migration-seed` bleibt bestehen; kein Runtime-Adapter, kein Default und keine automatische Migration in `0.15.0`. |
| Analytics (`MVP-41`) | > 50 Mio. Events/Tag, Ad-hoc-Analysebedarf, Owner, Kosten-/Abbruchdatum | Kein Hochvolumen-Workload, kein dedizierter Analytics-Owner und keine Abbruch-/Kostenannahme dokumentiert. | `defer` bleibt bestehen; kein ClickHouse-/VictoriaMetrics-/Mimir-POC in `0.15.0`. |

Kein Trigger ist erreicht. Daher entsteht kein Folgeplan aus Tranche 5;
der naechste bevorzugte Implementierungskandidat bleibt der in Tranche 4
zugeschnittene HTTP-Range-/Byte-Range-Slice fuer `0.16.0` Szenario B.

### 7.2 What aendert sich

- `RAK-105` ist geschlossen: Postgres und Analytics bleiben sichtbar
  deferred.
- `0.15.0` ist ein Decision-/Scope-Release ohne neue Runtime-Pflicht,
  ohne neuen Backend-Store und ohne neue externe Analyzer-/Control-
  Plane-App.
- Versionen, Changelog, Roadmap und Release-Plan sind auf `0.15.0`
  umgestellt.

### 7.3 What bleibt unveraendert

- SQLite bleibt lokaler Default; Postgres bleibt Folge-Scope nach ADR
  0005-Trigger.
- Analytics-Backends bleiben optionaler spaeterer POC-Pfad ohne
  Standardpflicht.
- K8s, Devcontainer und Release-Guard bleiben die in `0.14.0`
  validierten Zusatzpfade; `0.15.0` baut keinen Production-K8s-Pfad.

### 7.4 Verifikation

| Gate | Ergebnis |
| --- | --- |
| `make docs-check` | ✅ gruen |
| `make build` | ✅ gruen |
| `make gates` | ✅ gruen |
| `make security-gates` | ✅ gruen |
| `make release-guard-test` | ✅ gruen |
| `MTRACE_RELEASE_APPROVED=1 make release-guard VER=0.15.0` | ✅ gruen |
| Wave-2-Gates | n/a lokal nicht erneut via GitHub Actions geprüft; `0.15.0` ist ein Decision-/Scope-Release ohne Code-Slice. |

## 8. RAK-Verifikationsmatrix

Die finalen RAK-IDs fuer `0.15.0` sind mit Lastenheft-Patch `1.1.20`
vergeben. Status bleibt offen, bis die jeweilige Tranche ihren
Nachweis liefert.

| RAK | Prioritaet | Nachweis | Akzeptanz | Status |
| --- | --- | --- | --- | --- |
| RAK-101 | Muss | Zielgruppen-Decision, Lastenheft §16.1 oder ADR | Primaerzielgruppe ist entschieden; Plattform-Scope ist explizit deferred oder als spaeterer Planpfad beschrieben | [x] |
| RAK-102 | Muss | Analyzer-Boundary-Decision zu `MVP-20` | Externe Analyzer-API ist `proceed`, `POC`, `defer` oder `anders erfuellt`; Trigger sind messbar | [x] |
| RAK-103 | Muss | `F-132` Control-Plane-Decision | Control-Plane hat Scope, Nicht-Ziele und Trigger; keine Implementierung ohne Folgeplan | [x] |
| RAK-104 | Muss | `NF-13` Folge-Scope-Matrix | Naechster Analyzer-Slice ist eng zugeschnitten oder bewusst deferred | [x] |
| RAK-105 | Muss | Postgres-/Analytics-Trigger-Re-Eval | `MVP-40`/`MVP-41` bleiben deferred oder bekommen einen separaten Folgeplan bei erreichtem Trigger | [x] |

Sofort nutzbares Verifikationsmapping:

| RAK | Primaere Datei(en) | Datum | Owner | Status |
| --- | --- | --- | --- | --- |
| RAK-101 | `docs/planning/done/plan-0.15.0.md`, `spec/lastenheft.md` §13.19/§16.1 | 2026-05-12 | Product/PM | ✅ |
| RAK-102 | `docs/planning/done/plan-0.15.0.md`, `spec/lastenheft.md` §7.5.5/§12.1/§13.19, `apps/analyzer-service` Bestand, `spec/backend-api-contract.md` §3.6 | 2026-05-12 | Platform/Analyzer | ✅ |
| RAK-103 | `docs/planning/done/plan-0.15.0.md`, `spec/lastenheft.md` §7.5.6/§13.19, `spec/backend-api-contract.md` §3.8/§3.9, `docs/user/ingest-control.md` §5 | 2026-05-12 | Platform/Product | ✅ |
| RAK-104 | `docs/planning/done/plan-0.15.0.md`, `spec/lastenheft.md` §8.3/§13.19, `docs/planning/open/plan-0.16.0.md` Szenario B | 2026-05-12 | Platform/QA | ✅ |
| RAK-105 | `docs/planning/done/plan-0.15.0.md`, `docs/adr/0005-production-ops-backends.md`, `docs/planning/in-progress/risks-backlog.md` | 2026-05-12 | Platform/Ops | ✅ |

## 8.1 Blocker-Log

| Blocker | Betroffene Tranche | Status / Erwartete Aufloesung |
| --- | --- | --- |
| `0.14.0` noch nicht released | alle | ✅ geschlossen: `v0.14.0` released und Plan archiviert. |
| RAK-Range noch offen | Tranche 0/5 | ✅ geschlossen: `RAK-101`..`RAK-105` in Lastenheft `1.1.20`. |
| Zielgruppe nicht entschieden | Tranche 2/3 | ✅ geschlossen: Primärzielgruppe in Tranche 1 und Lastenheft §16.1 entschieden. |
| Kein externer Analyzer-Konsument | Tranche 2 | ✅ geschlossen: externe Analyzer-API deferred; Reaktivierungs-Trigger in §4.4 definiert. |
| Keine Ops-Trigger erreicht | Tranche 5 | ✅ geschlossen: Postgres bleibt `defer-with-migration-seed`, Analytics bleibt `defer`; kein Runtime-Scope in `0.15.0`. |

## 9. Folge-Scope nach `0.15.0`

- Spaeter: externer Analyzer-API-Slice, falls `MVP-20` auf `proceed`
  oder `POC` gesetzt wird.
- Spaeter: Control-Plane-Plan, falls `F-132` konkrete Betreiber- oder
  Multi-Project-Anforderungen bekommt.
- Spaeter: begrenzter Analyzer-Folge-Slice aus `NF-13`: Tranche 4
  empfiehlt HTTP-Range-/Byte-Range-Loader fuer manifest-referenzierte
  CMAF-Init-/Media-Segmente.
- Spaeter: Postgres- oder Analytics-Backend-Plan, falls die in ADR 0005
  dokumentierten Trigger erreicht werden.
