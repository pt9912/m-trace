# Backend-Stack-Spike — Spike-Protokoll

> **Living document auf `main`.** Notizen während AP-1 (Go) und AP-2
> (Kotlin/Micronaut) werden direkt hierher committed; AP-3 verdichtet sie
> zum ADR `docs/adr/0001-backend-stack.md`.
>
> **Bezug**: `docs/plan-spike.md` (§4.1, §6.2, §6.3, §7.3, §10),
> `docs/spike/0001-backend-stack.md` (§9 Bewertungsraster).

---

## 1. Vertrags-/Spec-Änderungen

> Nach dem Merge von `docs/spike/backend-api-contract.md` nach `main` ist
> der Kontrakt eingefroren. Jede Änderung **muss** hier dokumentiert sein
> (Plan §4.1) und in beiden Prototypen identisch landen.

| Datum | Änderung | Begründung | Commit-Hash |
|---|---|---|---|
| _–_ | _–_ | _–_ | _–_ |

---

## 2. AP-1: Go-Prototyp

**Branch**: `spike/go-api`  
**Start**: _YYYY-MM-DD_  
**Ende**: _YYYY-MM-DD_  
**Final-Commit**: _<hash>_

### 2.1 Notizen pro Bewertungskategorie

#### Time to Running Endpoint
_–_

#### OTel-Integration-Ergonomie
_–_

#### Hexagon-Fit
_–_

#### Test-Velocity
_–_

#### Docker Image Size
_–_

#### Cold Start
_–_

#### Build-Komplexität
_–_

#### Subjektiver Spaß
_–_

#### Contributor-Fit
_–_

#### Absehbare Phase-2-Risiken
_–_

### 2.2 Bonus-Scope (umgesetzt?)

- [ ] `GET /api/stream-sessions`
- [ ] `GET /api/stream-sessions/{id}`
- [ ] Origin-Allowlist pro Project
- [ ] expliziter Session-Lifecycle `ended`
- [ ] OTel-Counter zusätzlich zu Prometheus
- [ ] `trace_id` in Logs
- [ ] vollständige Hexagon-Schichtung bis ins Detail

### 2.3 Stolpersteine / Beobachtungen

_–_

---

## 3. AP-2: Kotlin-/Micronaut-Prototyp

**Branch**: `spike/micronaut-api`  
**Start**: _YYYY-MM-DD_  
**Ende**: _YYYY-MM-DD_  
**Final-Commit**: _<hash>_

### 3.1 Notizen pro Bewertungskategorie

#### Time to Running Endpoint
_–_

#### OTel-Integration-Ergonomie
_–_

#### Hexagon-Fit
_–_

#### Test-Velocity
_–_

#### Docker Image Size
_–_

#### Cold Start
_–_

#### Build-Komplexität
_–_

#### Subjektiver Spaß
_–_

#### Contributor-Fit
_–_

#### Absehbare Phase-2-Risiken
_–_

### 3.2 Bonus-Scope (umgesetzt?)

- [ ] `GET /api/stream-sessions`
- [ ] `GET /api/stream-sessions/{id}`
- [ ] Origin-Allowlist pro Project
- [ ] expliziter Session-Lifecycle `ended`
- [ ] OTel-Counter zusätzlich zu Prometheus
- [ ] `trace_id` in Logs
- [ ] vollständige Hexagon-Schichtung bis ins Detail

### 3.3 Stolpersteine / Beobachtungen

_–_

---

## 4. Reihenfolge und Bias (Plan §4.4)

- Tatsächliche Reihenfolge: _Go zuerst / Micronaut zuerst_
- Beobachteter Bias: _–_
- Korrekturmaßnahmen während des Spikes: _–_

---

## 5. Subjektive Gesamteindrücke

_Freier Block für übergreifende Notizen, Aha-Momente, Frustpunkte. Wird
in AP-3 zu den ADR-Abschnitten "Bewertung" und "Konsequenzen" verdichtet._

---

## 6. Übergang zu AP-3

- Bewertungsbogen (Spec §16) ausgefüllt im ADR: ☐
- Messwertbogen (Spec §17) ausgefüllt im ADR: ☐
- Reihenfolge-Bias im ADR notiert: ☐
- Sieger-Branch markiert: ☐
- Unterlegener Branch gelöscht oder als Tag
  `spike/backend-stack-loser-YYYYMMDD` archiviert: ☐
- Finale Commit-Hashes beider Prototypen im ADR: ☐
