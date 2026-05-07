# Implementation Plan — `0.12.0` (Auth / Token Lifecycle)

> **Status**: ⬜ open — noch nicht aktiviert. Dieser Plan darf erst nach
> explizitem Move nach `docs/planning/in-progress/` umgesetzt werden.
> Vorgänger ist voraussichtlich `0.11.0`; Aktivierung erst nach dessen
> Release-Closeout.
>
> **Release-Typ**: Minor-Release mit Lastenheft-Patch, neuer RAK-
> Gruppe, RAK-Verifikationsmatrix und Tag `v0.12.0`.
>
> **Ziel**: Die Token- und Policy-Erweiterungen `F-111`..`F-113`
> werden als zusammenhängender Security-/Auth-Scope geplant:
> serverseitig signierte Session Tokens, rotierbare Project Tokens und
> tenant-spezifische Ingest Policies.
>
> **Bezug**:
> [`spec/lastenheft.md`](../../../spec/lastenheft.md) F-106..F-113,
> NF-16, NF-24..NF-29, NF-30..NF-37, NF-38..NF-44.
>
> **Nachfolger**: offen.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Security- oder Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

In Scope:

- `F-111`: serverseitig signierte Session Tokens.
- `F-112`: rotierbare Project Tokens mit Migrationspfad.
- `F-113`: tenant-spezifische Ingest Policies, soweit ohne
  vollständige Multi-Tenant-Control-Plane möglich.
- Security-Doku, Threat-Model-Notizen und Contract-Tests für Token-
  Fehlerfälle.

Out of scope:

- Keine vollständige User-/Org-Verwaltung.
- Kein OAuth/OIDC-Login.
- Keine Admin-UI.
- Keine echte Multi-Tenant-SaaS-Plattform.
- Keine zentrale Secret-Management-Integration.

### 0.2 Offene Vorentscheidungen

- Ob signierte Session Tokens vor oder nach Project-Token-Rotation
  umgesetzt werden.
- Ob Token-Metadaten SQLite-persistiert oder konfigurationsbasiert
  gehalten werden.
- Ob tenant-spezifische Policies schon API-Pfadänderungen brauchen
  oder über bestehende Project-Konfiguration abgebildet werden.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung, Lastenheft-Patch, RAK-Gruppe und Threat Model | ⬜ |
| 1 | Signierte Session Tokens (`F-111`) | ⬜ |
| 2 | Project-Token-Rotation (`F-112`) | ⬜ |
| 3 | Ingest Policies (`F-113`) | ⬜ |
| 4 | SDK/API-Kompatibilität, Security-Doku und Smokes | ⬜ |
| 5 | Gates, RAK-Verifikationsmatrix, Versions-Bump, Closeout und Tag | ⬜ |

## 2. Tranche 0 — Aktivierung und Threat Model

DoD:

- [ ] Plan von `docs/planning/open/plan-0.12.0.md` nach
  `docs/planning/in-progress/plan-0.12.0.md` verschoben.
- [ ] `git status --short` vor erster Änderung dokumentiert.
- [ ] Lastenheft-Patch mit neuer RAK-Gruppe für `F-111`..`F-113`
  ergänzt.
- [ ] Threat-Model-Notiz ergänzt: öffentliche Browser-Tokens,
  Rotation, Replay, Origin-Bindung und Logging-Grenzen.
- [ ] Roadmap-Status und Release-Übersicht auf `0.12.0` als aktive
  Folgephase umgestellt.

## 3. Tranche 1 — Signierte Session Tokens

DoD:

- [ ] Token-Format und Signaturstrategie dokumentiert.
- [ ] Server validiert Ablaufzeit, Signatur und Project-Bindung.
- [ ] Fehlercodes sind stabil und getestet.
- [ ] Keine Token-Geheimnisse werden im Browser-Bundle oder in Logs
  offengelegt.

## 4. Tranche 2 — Project-Token-Rotation

DoD:

- [ ] Project Tokens können parallel gültige Generationen abbilden.
- [ ] Rotation hat klaren Deaktivierungs-/Grace-Pfad.
- [ ] Bestehende SDK- und Demo-Flows bleiben abwärtskompatibel.
- [ ] Contract-Tests decken alte, neue, abgelaufene und unbekannte
  Tokens ab.

## 5. Tranche 3 — Ingest Policies

DoD:

- [ ] Policy-Modell für erlaubte Origins, Methoden, Header und
  Ingest-Grenzen definiert.
- [ ] Tenant-spezifische Policies werden als Project-/Policy-Scope
  abgebildet, ohne Multi-Tenant-Produktversprechen.
- [ ] Tests decken Policy-Miss, Origin-Miss und Preflight-Fälle ab.

## 6. Tranche 4 — Doku, Compatibility und Smokes

DoD:

- [ ] Security-Doku beschreibt Token-Lifecycle und Meldeweg.
- [ ] Local-Development-Doku zeigt sichere Beispielwerte.
- [ ] SDK-Doku beschreibt kompatible Token-Nutzung.
- [ ] Relevante Smokes oder API-Tests sind im Release-Pfad verankert.

## 7. Tranche 5 — Release-Closeout

DoD:

- [ ] RAK-Verifikationsmatrix vollständig ausgefüllt.
- [ ] `make docs-check` grün.
- [ ] `make build` grün.
- [ ] `make gates` grün.
- [ ] `make security-gates` grün oder CI-Job `Security gates` grün
  dokumentiert.
- [ ] Wave-2-Quality-Gates vor dem Tag geprüft.
- [ ] Vollständiger Versions-Bump auf `0.12.0`.
- [ ] `CHANGELOG.md` mit `[0.12.0] - YYYY-MM-DD` aktualisiert.
- [ ] Plan nach `docs/planning/done/plan-0.12.0.md` verschoben und
  Status auf ✅ released aktualisiert.
- [ ] Annotierter Tag `v0.12.0` erstellt.
