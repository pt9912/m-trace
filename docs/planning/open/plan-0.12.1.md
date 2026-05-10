# Implementation Plan — `0.12.1` (Folgearbeiten nach `0.12.0`)

> **Status**: ⬜ open — noch nicht aktiviert. Dieser Plan darf erst nach
> expliziter Aktivierung umgesetzt werden.
>
> **Release-Typ**: Patch-Release mit Fokus auf Folgeaufgaben aus
> `docs/planning/in-progress/risks-backlog.md` (`0.12.0`-Nachlauf).
>
> **Ziel**: Alle noch offenen Folgepunkte aus dem Risiko-Backlog auf
> einen klaren Entscheidungs-/Umsetzungsweg bringen, entweder:
>
> - in `0.12.1` vollständig bearbeiten oder
> - mit Trigger-/Entscheidungsprotokoll sauber in `0.13.0` / spätere
>   Pläne verschieben.
>
> **Bezug**:
> [`docs/planning/in-progress/risks-backlog.md`](../in-progress/risks-backlog.md),
> [`docs/planning/done/plan-0.12.0.md`](../done/plan-0.12.0.md),
> [`docs/planning/in-progress/roadmap.md`](../in-progress/roadmap.md).
>
> **Nachfolger**: offen.

## 0. Konvention

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR-/Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.12.1` ist ein gebündelter Follow-up-Patch. Primärziele:

- Auth-/Secret- und Ingest-Folgepunkte aus `R-14`, `R-15`, `R-16`,
  `R-17`, `R-18`, `R-20`, `R-21`.
- Operative/Produktionsnähe-Optimierungen aus noch offenen Risks:
  `R-5`, `R-7`, `R-9`, `R-10`, `R-11`, `R-12`, `R-13`.
- Abschlussarbeit im Risiko-Backlog: Risiken entweder aufgelöst, verschoben
  oder als kontrollierte Defer-Entscheidung dokumentiert.

Zusätzlich explicit Folgerungen aus `docs/planning/done/plan-0.12.0.md`:

- OAuth/OIDC/SSO, Rollenmodell, User-/Org-/Admin-Verwaltung als spätere
  Ausbaustufen.
- Produktive MediaMTX-/SRS-Auth-Hook-Brücke.
- Signaturs-/Projekt-Secret-Management über `KMS`/`Vault`/`Cloud-Secret-Manager`
  als strukturierte Folgeentscheidung.
- Produktive Token-/Issuance-Hardwarepfade (Multi-Deployment-Rate-Limit/Rotation)
  als gesonderter Entscheidungsstrang.

Nicht im Umfang:

- Neuaufnahme von `0.13.0`-MVP-41/42-/MVP-44-Decision-Work (`Postgres`/`ClickHouse`/`K8s`/`Devcontainer`)
  über das hinaus, was klar als Folgearbeit aus den unten gelisteten Risiken
  nötig ist.
- Neue produktive Tenant-/User-/Admin-Featurefläche.

### 0.2 Vorgänger-Gate

- `0.12.0` ist in `docs/planning/done/plan-0.12.0.md` abgeschlossen.
- `docs/planning/in-progress/roadmap.md` muss bei Aktivierung auf `0.12.1`
  als aktive Folgephase gesetzt werden.
- Die bestehende Trennung Lab/Produktion in `0.12.0` bleibt bis zu
  expliziten neuen ADR-/RAK-Entscheidungen in Kraft.

### 0.3 Entscheidungsprinzip

- Jedes Risiko bekommt pro Tranche ein explizites Ergebnis: `implementiert`,
  `teilweise implementiert`, `defer` oder `entbehrlich`.
- Jeder Punkt aus dem 0.12.0-Folge-Scope bekommt denselben Entscheidungs- und
  Trigger-Mechanismus wie Risiken in dieser Matrix.
- Bei `defer` muss ein scharf formulierter Trigger mit Verantwortlichem
  hinterlegt werden (Datum + Auslösebedingung + erwartete Messgröße).
- Nach jedem Risiko-Ergebnis muss `docs/planning/in-progress/risks-backlog.md`
  konsistent nachgezogen werden (Status/Notiz/Absatz in §1.1 bzw.
  §1.2).

## 1. Offene Risiken aus 0.12.0 und Vorgänger-Scopes

| Risiko | Statusziel in `0.12.1` | Minimales Ergebnis |
|---|---|---|
| R-5 | Entscheiden | Cursor-/Anzeige-Path für Time-Skew-Metadaten finalisieren oder als Produktiv-Nichtbedarf begründen |
| R-7 | Implementieren | `network_signal_absent[]`-Auflösung auf bulk-/batch-fähigen Pfad prüfen/implementieren |
| R-9 | Definieren | K8s-Labelling-Regeln präzisieren; entweder implementieren oder im Plan als harte Defer-Trigger festlegen |
| R-10 | Entscheiden | Sample-Lückenerkennung für `sampleRate < 1` in nicht-Player-Pfaden verifizieren oder Defer mit externer Metrikgrenze |
| R-11 | Implementieren | Cursor-Pagination für `GET /api/srt/health/{stream_id}` inkl. Token-Shape, Tests und Migrationstoleranz |
| R-12 | Fortführen | WebKit/Safari-Drift-Pfad via CI/Smoke aktivieren oder formell als opt-in Defer mit Trigger dokumentieren |
| R-13 | Entscheiden | Trivy-Ignore-Strategie evaluieren: CVE-Expirationspunkt, Base-Image-Option (`distroless`) und Re-Review-Pfad |
| R-14 | Implementieren | Verhindern, dass `validate-key` als produktive Auth-Kette interpretiert wird: harte Doku + Guardrails im Operator-Workflow |
| R-15 | Implementieren | Klare Folgeentscheidung für externe Media-Server-Provisionierung inkl. minimaler opt-in-Adapter-Option |
| R-16 | Entscheiden | Ausgehende Webhook-Zustellung als separaten optionalen Lieferstrang oder als neuer `0.13.x`-Auslieferungspfad schließen |
| R-17 | Implementieren | Mehrinstanzfähige Rate-Limit-Variante (Redis/Memcached/SQLite) technisch freischalten oder harte Defer-Grenze definieren |
| R-18 | Implementieren | Multi-Key-Signing-Rotation operativ nutzbar machen (Env-Parsing + Runbook + Test / Smoke) |
| R-20 | Implementieren | KMS/Vault/Cloud-Secret-Manager Adapterpfad optionaler Standardfall oder als dokumentierte Defer-Bedingung |
| R-21 | Implementieren | `/api/ingest/*` Scope-Cut nach klarer Trigger-Boundary aufheben oder Release-bedingt in 0.13.x überführen |

## 1a. Direkte Folge-Items aus `plan-0.12.0.md` (jetzt in `0.12.1` führen)

Diese Punkte wurden in `plan-0.12.0.md` als Folge-Scope festgehalten und
werden in `0.12.1` nicht nur referenziert, sondern explizit mit
Entscheidungs- oder Umsetzungsverpflichtung geführt.

| ID | Folge-Item | Bereich | Statusziel in `0.12.1` | Minimales Ergebnis |
|---|---|---|---|---|
| OS-1 | OAuth/OIDC/SSO und Rollen-/User-/Org-/Admin-Verwaltung | Auth/Sicherheit | defer | Trigger-Entscheidung (Multi-Tenant-/Regulated-Requirement), konkrete Scope-Grenze und Release-Ziel für 0.14+ |
| OS-2 | Admin-UI / Produktive Auth-Management-UI | Auth/Sicherheit | defer | Begründete Defer-Entscheidung plus Minimalvoraussetzung für Reaktivierung |
| OS-3 | Produktive MediaMTX-/SRS-Auth-Hook-Brücke (signierte Publish-Tokens) | Ingest/Media-Server | defer | Entscheidung inkl. Adapter-Interface-Schnittstelle und Trigger-Metrik (Lab/Produktiv-Bedarf) |
| OS-4 | KMS/Vault/Cloud-Secret-Management für Signing-/Project-Token-Secrets | Secret-Management | defer/implementieren | Entweder Seed-Adapter + Runbook oder harte Triggerschwelle + Fallback-Doku |
| OS-5 | Multi-Deployment / Multi-Replica Secret- und Issuance-Mechanik als Standardfall | Skalierung/Operations | defer/implementieren | Konkreter Scope: nur Adapter-Pfad oder vollständige Standardfunktion |
| OS-6 | Origin-/IP-nahe Rate-Limit-Buckets | Rate Limiting/OPS | defer | Entscheidung über Pflicht-/Opt-in-Implementierung inkl. Messgröße |

## 2. Tranchen-Übersicht

## 2. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang | Status |
|---|---|---|---|---|---|
| 0 | Aktivierung, Scope-Präzisierung, Risiko-Map | Geklärte Ein-/Ausschlussliste und Verantwortlichkeiten | Tranche-Ready | Freigabe, Notizblock, Roadmap-Anpassung | ⬜ |
| 1 | Auth & Secret Follow-ups (`R-14`, `R-15`, `R-16`, `R-17`, `R-18`, `R-20`) | Konkrete Umsetzungen oder harte Defer-Trigger | Tranche 0 | Risikozustand aktualisiert + Nachweise | ⬜ |
| 2 | Telemetry/Read-Pfad Follow-ups (`R-5`, `R-7`, `R-9`, `R-10`, `R-11`, `R-12`, `R-13`) | Entscheidungen mit minimalen Implementierungen oder strukturierten Defer-Entscheidungen | Tranche 1 | Risiko-Entscheidungen + Tests + API/Docs-Nachweis | ⬜ |
| 3 | 0.12.0-Folge-Scope-Entscheidungen (`OS-1`..`OS-6`) | Defer-/Implementierungsentscheidungen in `plan-0.12.1` abschließen | Tranche 2 | 0.12.0-Ausnahmeumfang vollständig dokumentiert | ⬜ |
| 4 | Closeout und Roadmap-Konsolidierung | Alle offenen Risiken inkl. `OS-*` konsistent in `risks-backlog.md` nachgeführt; Plan-Status in Roadmap gesetzt | Tranche 3 | Geprüfte Abschlussmatrix + Release-Candidate-Entscheidung | ⬜ |

## 3. Tranche 0 — Aktivierung und Zielbild

DoD:

- [ ] Plan-Status auf `open` aktivieren und `roadmap.md`-Follow-up auf
  `0.12.1` setzen.
- [ ] Risikobesitz klären (Owner je Risiko), inkl. Eskalationspfad bei Triggern.
- [ ] Jede Risiken-Spalte auf "Implementieren/Defer/Entbehrlich" vorläufig
  mappen.
- [ ] Ausprägungs-Definition pro Tranche dokumentieren:
  Architekturentwurf, Doku-Anpassung, Testumfang, Rollback.

Go/No-Go nach Tranche 0:

- **Go**: Entweder 100% eindeutige Risk-Zuweisung oder schriftlich begründete
  Defer-Entscheidung für jedes Item.
- **No-Go**: Unklare Verantwortlichkeit oder fehlender Trigger bei `defer`.

## 3a. Tranche 1 — Auth- und Ingest-Folgearbeiten (`R-14`..`R-21`)

DoD:

- [ ] `R-14`: Operator-sichere Warnpfade dokumentieren und implementieren
  (klarer Schutzhinweis, Fail-Safety gegen Missbrauch im Ingest-Flow).
- [ ] `R-15`: Entscheidung „opt-in Externer Provisionierungs-Adapter“ trifft
  zwischen Seed-Implementierung oder klarer Triggerschwelle.
- [ ] `R-16`: Outbound-Webhooks als Mini-MVP/Follow-up mit Retry-/Timeout-Schema oder explizit als 0.13+ Scope dokumentieren.
- [ ] `R-17`: Shared-Rate-Limiter-Adapter (`redis`/`sqlite`/`memcached`) als
  implementierbar markiert oder als harte Triggerschwelle inkl. Migrationspfad dokumentiert.
- [ ] `R-18`: Multi-Key-Rotation für Signer in Doku + Smoke + Testablauf.
- [ ] `R-20`: Secret-Backend-Adapter-Contract skizzieren (API, Rotation,
  Fehler-Modell) und in 0.12.1 als implementierbar oder Defer fixieren.

## 4. Tranche 2 — Telemetry-/Read-Pfad-Folgearbeiten (`R-5`,`R-7`,`R-9`,`R-10`,`R-11`,`R-12`,`R-13`)

DoD:

- [ ] `R-11`: Cursor-Pagination implementieren oder mit Entscheidungsprotokoll auf
  `0.13.x` schieben.
- [ ] `R-7`: Effektive Batch- oder Prefetch-Lösung gegen N+1-Latenz definieren
  (Proof-of-concept + Grenzfälle).
- [ ] `R-5`: Sichtbarkeit von Time-Skew im Read-Pfad entweder
  implementiert oder Trigger-Entscheidung dokumentiert.
- [ ] `R-9` / `R-12`: K8s-Drift-/Label-Pfade als Option mit konkretem Trigger.
- [ ] `R-10`: Sampling-Gaps für relevante Pfade als messbare Entwurfsentscheidung
  (Implementieren vs. Defer).
- [ ] `R-13`: Nach Ablaufdatum und CVE-Entwicklung klare Re-Review-Regel aktiv.

## 5. Tranche 3 — Closeout / Release-Closeout

DoD:

- [ ] Alle betroffenen Risiken im `risks-backlog.md` auf den neuen
  Abschlusszustand gebracht (resolved / defer / offen).
- [ ] `docs/planning/in-progress/roadmap.md` und ggf. Lastenheft/ADR-Verweise
  auf die neuen Scope-Entscheidungen angepasst.
- [ ] `docs/planning/open/plan-0.12.1.md` enthält Commit-Hooks als Closeout-Liste
  (Datei-/Test-/Doku-Nachweis je Risiko-Cluster).
- [ ] Release-Entscheidung inkl. Tag-Strategie dokumentiert und Reviewer-Check
  gegen Roadmap- und Risiko-Führung durchgeführt.

## 6. Qualitätsregeln für 0.12.1

- Keine unkontrollierten Scope-Verdoppelungen mit `0.13.0`.
- Jede Entscheidung enthält "Was ändert sich / Was bleibt unverändert".
- Kein Risiko darf ohne dokumentierten Trigger in "defer" wandern.
- Kein neues Pflicht-Tooling ohne Relevanzbegründung.
- Jede abgeschlossene Maßnahme: mindestens 1 Testnachweis + 1 Doku-Nachweis.
