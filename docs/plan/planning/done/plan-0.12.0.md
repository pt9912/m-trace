# Implementation Plan — `0.12.0` (Auth / Token Lifecycle)

> **Status**: ✅ released — Minor-Release `v0.12.0` ausgeliefert
> 2026-05-10 (Lastenheft `1.1.15`, RAK-71..RAK-76 in §13.14).
> Vorgänger ist `0.11.0` (`v0.11.0`, Ingest-Gateway / Stream Control;
> Plan in [`done/plan-0.11.0.md`](../done/plan-0.11.0.md)).
>
> **Release-Typ**: Minor-Release mit Lastenheft-Patch `1.1.15`
> (Vorschlag), neuer RAK-Gruppe `RAK-71`..`RAK-76`,
> RAK-Verifikationsmatrix und Tag `v0.12.0`.
>
> **Ziel**: Die bisher als Kann geführten Token- und Policy-
> Erweiterungen `F-111`..`F-113` werden in einen begrenzten
> Security-/Auth-Scope geschnitten: kurzlebige serverseitig signierte
> Session Tokens für Browser-Telemetrie, rotierbare Project-Token-
> Generationen mit Grace-Pfad und Project-gebundene Ingest Policies
> für Origins, CORS und Rate-Limit-Grenzen. `0.12.0` liefert keine
> vollständige User-/Org-Verwaltung und keine SaaS-Control-Plane.
>
> **Bezug**:
> [`spec/lastenheft.md`](../../../../spec/lastenheft.md) §7.11
> `F-106`..`F-113`, §8.3 `NF-16`, §8.5 `NF-24`..`NF-29`,
> §8.5 CORS/CSP `NF-30`..`NF-37`, §8.6 Datenschutz/GDPR
> `NF-38`..`NF-44`, §15.1 Risiko „Project Token im Browser-Code wird
> zweckentfremdet";
> [`spec/backend-api-contract.md`](../../../../spec/backend-api-contract.md)
> §4 Authentifizierung und §5 Validierungsreihenfolge;
> [`docs/planning/done/plan-0.11.0.md`](../done/plan-0.11.0.md) §10
> Folge-Scope nach `0.11.0`;
> [`docs/planning/in-progress/risks-backlog.md`](../risks-backlog.md)
> R-14..R-16.
>
> **Nachfolger**: voraussichtlich `0.13.0` (Production / Ops Backends).
> Production-Secret-Backends, KMS/Vault, Postgres/ClickHouse und
> Kubernetes-Manifeste bleiben dort oder später.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Security-, Scope- oder Architekturentscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.12.0` liefert **Auth-Härtung für den bestehenden lokalen/API-nahen
Pfad**, nicht eine vollständige Identity- oder Tenant-Plattform.

In Scope:

- `F-111`: serverseitig signierte Session Tokens.
  - Kurzlebige Tokens für Browser-Telemetrie und SDK-Nutzung.
  - Token ist an `project_id`, optional `session_id`, `origin`,
    erlaubte Verwendung (`aud`) und Ablaufzeit gebunden.
  - Signatur wird serverseitig validiert; keine Signatur-Secrets im
    Browser-Bundle, in Fixtures, Logs, Traces oder Beispiel-Configs.
  - Bestehende `X-MTrace-Token`-Project-Token-Requests bleiben im
    `0.12.0`-Compatibility-Fenster gültig.
- `F-112`: rotierbare Project Tokens.
  - Project Tokens werden als Generationen modelliert: `token_id`,
    `project_id`, Hash/Fingerprint, Status, `not_before`,
    `grace_until?`, `expires_at?`, `revoked_at?`.
  - Rotation erlaubt eine zeitlich begrenzte Grace-Phase, in der alte
    und neue Generation gültig sind. `grace_until` ist dabei das
    persistierte Feld für die alte Generation und die Quelle der
    Restart-stabilen Grace-Validierung; `expires_at` ist das optionale
    harte Ablaufdatum einer Generation.
  - Persistenz speichert keinen Klartext-Token; Klartext erscheint nur
    in bewusst markierten Operator-/Test-Fixtures, falls überhaupt.
  - Rotation wird über Repository/Config/Operator-Pfad nachgewiesen,
    nicht über eine öffentliche Admin-UI.
- `F-113`: tenant-spezifische Ingest Policies im Project-Scope.
  - Project Policy umfasst erlaubte Origins, Methoden, Header,
    verpflichtende Rate-Limit-Buckets pro Project und optionale
    Ingest-Grenzen für den `0.11.0`-Stream-Control-Pfad. Origin- und
    IP-nahe Buckets sind nur dann Teil des `0.12.0`-Muss-Scopes, wenn
    die bestehende Rate-Limit-Infrastruktur sie ohne größere
    Architekturänderung tragen kann; andernfalls werden sie als
    Folge-Scope dokumentiert.
  - Policies sind Project-gebunden; es gibt noch kein User-/Org-
    Rollenmodell.
  - Bestehende CORS-Regeln aus `NF-30`..`NF-37` werden in den
    Policy-Vertrag überführt und getestet.
- API-/Contract-Tests für Auth-Fehler, Token-Rotation, CORS-
  Preflight, Policy-Miss und Kompatibilität.
- Security-Doku inklusive Threat Model, CSP-Beispiele, Logging-
  Grenzen, datenschutzfreundlichen Defaults und Migration von
  `demo-token` auf rotierbare Project-Tokens plus Session Tokens.

Out of scope:

- Keine User-/Org-Verwaltung, kein Login, kein OAuth/OIDC, kein SSO.
- Keine Admin-UI und keine mandantenfähige SaaS-Control-Plane.
- Keine KMS-/Vault-/Cloud-Secret-Manager-Integration.
- Keine produktive MediaMTX-/SRS-Auth-Hook-Kopplung als Muss-Scope.
  Der `0.11.0`-Endpoint `validate-key` bleibt bis zu einer separaten
  Auth-Bridge Diagnose-/Smoke-Funktion, kein Media-Server-Auth-
  Ersatz.
- Keine globale Stream-Key-Rotation über mehrere Deployments.
- Keine Änderung am Event-Schema `1.0`, solange der Auth-Envelope
  aus Headern kommt.
- Keine Cookies für Player-Telemetrie; SDK bleibt bei
  `credentials:"omit"`.
- Keine Production-Ops-Backends aus `0.13.0`.

### 0.2 Vorgänger-Gate

- `0.11.0` ist released; Roadmap zeigt `0.12.0` als aktivierbare
  Folgephase.
- Ingest-Control-Grenzen aus `0.11.0` bleiben verbindlich:
  lokale/lab-nahe Stream-Control, keine produktive Media-Server-
  Auth-Kopplung, keine externe Provisionierung.
- Bestehende Auth-Vertragsregeln aus
  `spec/backend-api-contract.md` §4/§5 gelten bis zur expliziten
  Änderung weiter: fehlender Pflicht-Auth-Header liefert `401` vor
  Body-Read, Body > 256 KB mit Auth-Header liefert `413`, Token-
  Mismatch liefert `401`.

### 0.3 Architektur- und Persistenzentscheidung

Vorschlag für `0.12.0`: **Variante B** — Auth bleibt als Modul in
`apps/api`, analog `0.11.0` Ingest-Control. Ein eigener Auth-Service
ist Folge-Scope.

| Thema | Entscheidungsvorschlag | Begründung |
| --- | --- | --- |
| Session-Token-Format | kompaktes, eigenes HMAC-SHA-256-signiertes Token mit Prefix `mtr_st_`, `kid`, canonical JSON Claims und Base64url-Encoding | keine neue Runtime-Abhängigkeit, auditierbar, genügt für lokalen/API-nahen Scope |
| Signaturschlüssel | Server-seitige Key-Ring-Konfiguration über ENV/File; `kid` erlaubt Rotation; alte Verify-Keys bleiben bis zum Ablauf aller damit signierten Tokens geladen | keine Secrets im Repo; restart-stabiler Dauerbetrieb; später auf KMS/Vault migrierbar |
| Project-Token-Persistenz | SQLite über bestehende API-Persistenz plus InMemory-Testadapter | Rotation braucht dauerhafte Generationen und Restart-Nachweis |
| Klartext-Token | nur bei Erzeugung/Rotation oder in markierten Demo-Fixtures; persistiert wird Hash + Fingerprint | analog Stream-Key-Grenze aus `0.11.0` |
| Admin-Oberfläche | keine UI; Rotation wird über Application-Service/Repository und optional Operator-CLI/Script nachgewiesen | vermeidet Control-Plane-Scope |
| Backward Compatibility | `X-MTrace-Token` bleibt gültig; Session Token wird zusätzlich über `Authorization: Bearer` oder `X-MTrace-Session-Token` akzeptiert | SDK-/Demo-Flows brechen nicht |

Falls bei Aktivierung eine Standardbibliothek für signierte Tokens
eingeführt werden soll, muss Tranche 0 explizit entscheiden, warum die
zusätzliche Abhängigkeit den Security-/Wartungsgewinn rechtfertigt.

### 0.4 Lastenheft-Patch `1.1.15` (Vorschlag)

Der Patch hebt `F-111`..`F-113` für den begrenzten `0.12.0`-Scope von
Kann auf Release-Muss und ergänzt eine neue RAK-Gruppe.

| RAK | Priorität | Inhalt |
| --- | --- | --- |
| RAK-71 | Muss | Auth-Scope ist normativ begrenzt: kurzlebige Session Tokens, Project-Token-Generationen und Project-Policies; keine User-/Org-Verwaltung, kein OAuth/OIDC, keine Admin-UI, keine KMS-/Vault-Pflicht. |
| RAK-72 | Muss | Signierte Session Tokens: API kann kurzlebige Tokens ausstellen und validieren; Tokens sind an Project, Ablaufzeit, Audience und optional Session/Origin gebunden; ungültige, abgelaufene oder falsch gebundene Tokens liefern stabile `401`/`403`-Fehler ohne Secret-Leak. |
| RAK-73 | Muss | Project-Token-Rotation: mehrere Generationen pro Project sind modelliert; aktive, Grace-, abgelaufene und widerrufene Tokens werden deterministisch validiert; Persistenz speichert nur Hash/Fingerprint, `grace_until` und Metadaten. |
| RAK-74 | Muss | Ingest Policies: erlaubte Origins, Methoden, Header und Rate-Limit-Grenzen sind Project-gebunden konfigurierbar und werden in Request-Pfaden erzwungen; Preflight nutzt dokumentierte globale, konservative Regeln, solange das Project nicht deterministisch aus Route oder Header bestimmbar ist. |
| RAK-75 | Muss | Backward Compatibility: bestehende `X-MTrace-Token`-Flows, Demo, SDK, Analyze-/Session-Link-Auth und `0.11.0` Ingest-Control bleiben kompatibel oder haben dokumentierte Migrationstests. |
| RAK-76 | Muss | Security-Doku, Threat Model, Contract-Fixtures und Smokes beschreiben Token-Lifecycle, Rotation, Replay-/Leakage-Grenzen, CSP/CORS-Beispiele, GDPR-/Datenschutzgrenzen und den Unterschied zu Production-Secret-Backends aus Folge-Scope. |

### 0.5 Öffentliche API und Wire-Skizze

Neue oder geänderte Wire-Verträge werden in
`spec/backend-api-contract.md` ergänzt. Vorschlag:

| Methode | Pfad | Zweck |
| --- | --- | --- |
| `POST` | `/api/auth/session-tokens` | Kurzlebiges Session Token aus einem gültigen Project Token ausstellen |
| `POST` | `/api/playback-events` | akzeptiert weiterhin `X-MTrace-Token`; zusätzlich `Authorization: Bearer mtr_st_*` oder `X-MTrace-Session-Token` |
| `POST` | `/api/ingest/*` | bleibt im `0.11.0`-Token-Validierungs-Pfad (RAK-65, lokale/lab-nahe Stream-Verwaltung); konsumiert keine `0.12.0`-Project-Policy, weil der Pfad `curl`-/Operator-driven und nicht Browser-Konsumenten-Pfad ist. Project-Policy-Enforcement für Ingest-Control bleibt Folge-Scope. |

**Auth-Header-Priorität und Fehlerpräzedenz für tokenpflichtige
Requests.**

1. `Authorization: Bearer mtr_st_*` ist der bevorzugte Session-
   Token-Pfad. Andere `Authorization`-Werte (z. B. fremde OAuth-/
   Reverse-Proxy-Header) sind für m-trace Auth nicht auswertbar und
   werden ignoriert, solange ein gültiger m-trace Header wie
   `X-MTrace-Token` vorhanden ist. Ohne gültigen m-trace Header
   liefern sie `401 auth_token_missing`.
2. `X-MTrace-Session-Token` ist der alternative Session-Token-Pfad
   für Umgebungen, in denen `Authorization` nicht verwendet werden
   soll.
3. `X-MTrace-Token` ist der Legacy-/Project-Token-Pfad und bleibt im
   `0.12.0`-Compatibility-Fenster gültig.
4. Wenn mehr als ein Auth-Mechanismus gleichzeitig präsentiert wird,
   müssen alle präsentierten Tokens zum selben `project_id` passen.
   Widersprüche liefern stabil `401 auth_project_mismatch`; ein
   ungültiger zusätzlicher Token liefert `401 auth_token_invalid`.
   Es gibt keinen stillen Fallback von einem ungültigen höher
   priorisierten Token auf einen gültigen niedriger priorisierten
   Token.
5. Wenn `Authorization` und `X-MTrace-Session-Token` beide gesetzt
   sind und unterschiedliche Session Tokens enthalten, liefert die API
   stabil `401 auth_token_invalid`, auch wenn einer der beiden Tokens
   für sich genommen gültig wäre.

Diese Regeln werden in `spec/backend-api-contract.md` gepinnt und mit
Tests für alle Header-Kombinationen abgedeckt. Zusätzlich gilt eine
deterministische Fehlerpräzedenz, wenn mehrere Auth-Header gleichzeitig
gesetzt sind:

| Priorität | Bedingung | Status/Code |
| ---: | --- | --- |
| 1 | Pflicht-Auth fehlt vollständig | `401 auth_token_missing` |
| 2 | Ein präsentierter m-trace Token ist syntaktisch malformed oder Signatur/Hash ungültig | `401 auth_token_invalid` |
| 3 | Ein präsentierter Token ist widerrufen | `401 auth_token_revoked` |
| 4 | Ein präsentierter Token ist abgelaufen | `401 auth_token_expired` |
| 5 | Ein präsentierter Token ist noch nicht gültig | `401 auth_token_not_yet_valid` |
| 6 | Alle Tokens sind für sich gültig, binden aber unterschiedliche Projects | `401 auth_project_mismatch` |
| 7 | Session Token ist für falsche Audience/Session/Origin gebunden | `403 auth_session_scope_denied` |
| 8 | Project Policy lehnt Origin/Methode/Header/Scope ab | `403 auth_policy_denied` |
| 9 | Endpoint-spezifisches Rate-Limit nach erfolgreicher Auth/Policy-Prüfung überschritten | `429` mit dem im Endpoint-Vertrag definierten Rate-Limit-Code |

Diese Präzedenz verhindert stillen Fallback: ein ungültiger
höher-priorisierter Token neben einem gültigen niedriger-priorisierten
Token liefert weiter `401 auth_token_invalid`. Project-Mismatch wird
nur geprüft, wenn alle präsentierten Tokens vorher syntaktisch,
kryptografisch und zeitlich gültig sind. Fremde `Authorization`-
Header ohne `mtr_st_`-Bearer-Token sind kein m-trace Auth-Versuch und
blockieren den Legacy-Project-Token-Pfad nicht.

Rate-Limits sind bewusst endpoint-spezifisch: Die Präzedenzmatrix legt
nur fest, dass `429` nach den Auth-, Scope- und Policy-Entscheidungen
kommt. `POST /api/auth/session-tokens` verwendet dafür ausschließlich
`auth_issuance_rate_limited`; andere Endpoints müssen ihren eigenen
Rate-Limit-Code im API-Kontrakt nennen und dürfen
`auth_issuance_rate_limited` nicht wiederverwenden.

**CORS-Preflight-Modell.**

Browser-Preflights enthalten in der Praxis kein Project- oder Session-
Token, das der Server verlässlich validieren kann. Deshalb nutzt
`0.12.0` für `OPTIONS` eine globale, konservative und
informationsarme Allowlist:

- erlaubte Methoden maximal `POST, OPTIONS`;
- erlaubte Header maximal `Content-Type`, `Authorization`,
  `X-MTrace-Token`, `X-MTrace-Session-Token`, `traceparent`;
- bekannte Origins aus der globalen Union aller konfigurierten
  Project-Origins können mit dem konkreten Origin gespiegelt werden,
  nie mit `*`; erfolgreiche Preflights liefern exakt `204`, leeren
  Body, `Access-Control-Allow-Origin: <Origin>`,
  `Access-Control-Allow-Methods: POST, OPTIONS`,
  `Access-Control-Allow-Headers` mit der erlaubten Header-Liste,
  `Access-Control-Max-Age: 600`, `Vary: Origin` und
  `Cache-Control: no-store`;
- unbekannte Origins erhalten eine minimale Ablehnung ohne
  `Access-Control-Allow-Origin` und ohne project-spezifische
  Diagnose. Die Antwort ist deterministisch: exakt `204`, leerer
  Body, kein `Access-Control-Allow-Origin`, kein
  `Access-Control-Allow-Methods`, kein `Access-Control-Allow-Headers`,
  `Vary: Origin` und `Cache-Control: no-store`;
- project-spezifische Policy-Entscheidungen passieren erst beim
  tatsächlichen `POST`, wenn Project-/Session-Token ausgewertet werden
  können.

RAK-74 gilt damit für Request-Enforcement; Preflight ist ein
konservativer Browser-Kompatibilitätsfilter. Falls ein späterer Scope
projekt-festes Routing wie `/api/projects/{project_id}/...` einführt,
kann Preflight auf project-spezifische Policies umgestellt werden.
Das akzeptierte Informationsniveau ist: Ein Client darf erkennen, ob
sein eigener Origin global bekannt ist, aber nicht welche Projects
existieren oder welche anderen Origins konfiguriert sind.

`POST /api/auth/session-tokens` Request:

```json
{
  "project_id": "demo",
  "session_id": "sess_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
  "origin": "http://localhost:5173",
  "audience": "playback-events",
  "ttl_seconds": 900
}
```

`audience` ist kein freies Client-Feld. Der Server akzeptiert nur
Audiences aus der Project Policy; im `0.12.0`-Pflichtpfad ist
`playback-events` die einzige Muss-Audience. Unbekannte oder für das
Project nicht erlaubte Audiences liefern `403 auth_session_scope_denied`
und erzeugen keinen Token. Neue Audiences für weitere Endpoints müssen
jeweils im API-Kontrakt, in der Policy-Allowlist und in Contract-Tests
ergänzt werden.

Der Request ist mit einem gültigen Project Token authentifiziert
(`X-MTrace-Token`). `project_id` ist optional und dient als
Konsistenzcheck zum Token: fehlt es, wird das Project ausschließlich
aus dem Token abgeleitet und die Response enthält das abgeleitete
`project_id`; ist es gesetzt und passt nicht zum Token, liefert die API
`401 auth_project_mismatch`. `ttl_seconds` wird serverseitig gegen die
harte globale und die project-spezifische Grenze validiert.
Clients sollen `project_id` setzen, wenn sie mehrere Projects verwalten
oder präzise Debugging-Signale brauchen; einfache SDK-Flows dürfen es
weglassen.

`ttl_seconds` hat im `0.12.0`-Pflichtpfad eine harte globale Obergrenze
von 900 Sekunden. Project Policies dürfen eine niedrigere Grenze
definieren, aber keine höhere. `project_max_ttl_seconds` ist für jedes
Project effektiv vorhanden: fehlt der Wert in Config oder Persistenz,
gilt exakt der Default `900`. Fehlt `ttl_seconds`, verwendet der Server
`min(project_max_ttl_seconds, 900)`. Ist `ttl_seconds` kleiner oder
gleich 0 oder größer als die wirksame Project-Grenze, liefert die API
`422 auth_token_ttl_too_large`; es gibt keinen stillen Clamp.

Response:

```json
{
  "session_token": {
    "value": "mtr_st_eyJraWQiOiJrZXlfMjAyNi0wNSJ9...",
    "token_id": "st_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "project_id": "demo",
    "session_id": "sess_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "audience": "playback-events",
    "expires_at": "2026-05-09T10:15:00Z"
  }
}
```

`session_token.value` darf nur in der Issuance-Antwort erscheinen.
Logs, Fehlerantworten, Metriken, Traces und Fixtures enthalten
höchstens `token_id` oder Fingerprints. `token_id` ist der öffentliche
Wire-Name des `jti`-Claims; beide Werte sind identisch. Implementierung
und Tests verwenden `jti` nur innerhalb der signierten Claims und
`token_id` in Responses, Logs und Doku.

Fehlercodes (Vorschlag, final in API-Kontrakt zu pinnen):

| Code | Status | Bedeutung |
| --- | ---: | --- |
| `auth_token_missing` | `401` | Pflicht-Auth fehlt |
| `auth_token_invalid` | `401` | Project- oder Session-Token ungültig |
| `auth_token_expired` | `401` | Token abgelaufen |
| `auth_token_not_yet_valid` | `401` | Token-Generation noch nicht gültig |
| `auth_token_revoked` | `401` | Token-Generation widerrufen |
| `auth_project_mismatch` | `401` | Request-Project passt nicht zum Token |
| `auth_policy_denied` | `403` | Project Policy lehnt Origin/Methode/Header/Scope ab |
| `auth_session_scope_denied` | `403` | Session Token passt nicht zu Audience oder Session |
| `auth_token_ttl_too_large` | `422` | gewünschte TTL ist <= 0 oder überschreitet die wirksame Project-TTL-Grenze (max. 900 s) |
| `auth_issuance_rate_limited` | `429` | Session-Token-Issuance-Quote überschritten |

### 0.6 Threat Model

Mindestinhalte für Tranche 0 und Doku:

- **Public Browser Token**: Project Tokens im Browser sind niedrige
  Kritikalität, aber dürfen nicht dauerhaft oder grenzenlos wirken.
- **Replay**: Session Tokens sind kurzlebig, optional an
  `session_id`/`origin` gebunden und tragen `exp`/`nbf`.
- **Issuance-Abuse**: `POST /api/auth/session-tokens` ist ein eigener
  Missbrauchspfad. Auch mit gültigem Project Token gelten globale und
  Project-Quoten; Überschreitung liefert
  `429 auth_issuance_rate_limited`. Origin- und IP-nahe Quoten sind
  Defense-in-Depth und werden nur umgesetzt, wenn die bestehende
  Infrastruktur sie ohne größere Architekturänderung tragen kann.
  Token-Issuance darf keine unbegrenzten Logs, DB-Zeilen oder teuren
  Signaturpfade erzeugen.
- **Issuance-Leakage und Caching**: Clients sollen kurzlebige Session
  Tokens bis knapp vor `expires_at` wiederverwenden statt pro Event
  neu zu minten. Doku und SDK-Beispiele müssen Caching mit sicherem
  Refresh-Fenster zeigen und dürfen Session Tokens nicht in
  persistenten Browser-Speichern empfehlen.
- **Leakage**: Klartext-Tokens nie in Logs, Traces, Metriken,
  Contract-Fixtures oder Screenshots; Tests pinnen Redaktionsregeln.
- **Privacy**: Session-Token-Claims und Token-Metadaten enthalten nur
  notwendige pseudonyme Identifier; IP-/User-Agent-Speicherung bleibt
  minimiert und wird nicht durch Auth-Logging ausgeweitet.
- **Origin-Bindung**: `Origin`-Header wird gegen Project Policy
  geprüft; leerer Origin bleibt für CLI/curl nur dort erlaubt, wo der
  Kontrakt das explizit vorsieht.
- **Rotation**: alte Project-Token-Generationen haben klare Grace- und
  Revocation-Regeln; `grace_until` wird persistiert und ist die
  Restart-stabile Quelle für die Grace-Entscheidung. Session Tokens
  werden nicht rückwirkend unendlich gültig.
- **Signatur-Key-Rotation**: `kid` im Session Token ermöglicht
  parallele Signatur-Keys; alte Verify-Keys bleiben über Deployments
  und Restarts geladen, bis alle damit signierten Session Tokens
  abgelaufen sind. Unbekannter `kid` liefert `401`.
- **Rate Limits**: Policy-Grenzen gelten als Muss mindestens global und
  pro Project; Origin-/IP-nahe Buckets sind optionaler Zusatz oder
  Folge-Scope.
- **Abgrenzung**: keine Production-Secret-Verwaltung, kein OAuth/OIDC,
  kein Media-Server-Auth-Hook-Versprechen.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| --- | --- | --- |
| 0 | Plan-Aktivierung, Lastenheft-Patch, RAK-Gruppe, Architektur-/Persistenzentscheidung und Threat Model | ✅ |
| 1 | Auth-Domainmodell, Token-Generationen und Project Policies | ✅ |
| 2 | Signierte Session Tokens (`F-111`) und Auth-Wire-Vertrag | ✅ |
| 3 | Project-Token-Rotation (`F-112`) mit SQLite-/InMemory-Persistenz | ✅ |
| 4 | Ingest Policies (`F-113`), CORS/Preflight und Rate-Limit-Integration | ✅ |
| 5 | SDK/API-Kompatibilität, Doku, Contract-Fixtures und Smokes | ✅ |
| 6 | Gates, RAK-Verifikationsmatrix, Versions-Bump, Closeout und Tag | ✅ |

---

## 2. Tranche 0 — Aktivierung, Patch und Entscheidungen

Ziel: Der Security-Scope ist normativ geschlossen, bevor Auth-Code
geändert wird.

DoD:

- [x] Plan von `docs/planning/open/plan-0.12.0.md` nach
  `docs/planning/in-progress/plan-0.12.0.md` verschoben (2026-05-10).
- [x] `git status --short` vor erster Änderung dokumentiert: clean
  (Repo-Status auf Tag `v0.11.0` (`200cd1b`); seither nur Plan-Patches
  unter `docs/planning/open/plan-0.12.0.md`).
- [x] Lastenheft-Header auf `1.1.15` erhöht.
- [x] `F-111`..`F-113` im Lastenheft für den begrenzten `0.12.0`-
  Scope nachvollziehbar von Kann-Historie auf Release-Muss
  abgebildet (`spec/lastenheft.md` §7.11; historische Kann-Stufung
  bis `1.1.14` bleibt als auditierbarer Stand erhalten).
- [x] RAK-71..RAK-76 im Lastenheft ergänzt (`spec/lastenheft.md`
  §13.14).
- [x] Patch-Log im aktivierten `plan-0.12.0.md` geführt und im
  Closeout über Roadmap, `CHANGELOG.md` und
  `docs/planning/done/plan-0.12.0.md` tracebar gemacht (Patch
  `1.1.15` als §4a.18 in
  [`docs/planning/done/plan-0.1.0.md`](../done/plan-0.1.0.md) für
  Konsistenz mit `1.1.10`..`1.1.14` geführt; Closeout-Tranche pflegt
  Roadmap und `CHANGELOG.md`).
- [x] Architekturentscheidung dokumentiert: Auth bleibt in `apps/api`
  oder explizit begründete Abweichung — **Variante B** in §0.3 dieses
  Plans und in Lastenheft-Patch `1.1.15` (Header-Block + §13.14
  Vorspann); ein eigener Auth-Service bleibt Folge-Scope.
- [x] Persistenzentscheidung dokumentiert: SQLite + InMemory-Testpfad
  oder explizit begründete Config-only-Variante — SQLite über
  bestehende API-Persistenz plus InMemory-Testadapter, Klartext-
  Tokens werden nicht persistiert (§0.3 dieses Plans und
  Lastenheft-Patch `1.1.15`).
- [x] Threat-Model-Notiz mit Public Browser Token, Replay, Leakage,
  Privacy, Origin-Bindung, Signatur-Key-Rotation und Logging-Grenzen
  ergänzt — §0.6 dieses Plans (Public Browser Token, Replay,
  Issuance-Abuse, Issuance-Leakage/Caching, Leakage, Privacy,
  Origin-Bindung, Rotation, Signatur-Key-Rotation, Rate Limits,
  Abgrenzung); Wire-/Header-/Fehler-Vertrag in
  `spec/backend-api-contract.md` §3.9.
- [x] `spec/backend-api-contract.md` bekommt einen `0.12.0`-
  Abschnitt mit Auth-Matrix, Token-Issuance-Wire, Fehlercodes und
  Validierungsreihenfolge (§3.9 neu; §2 Endpunktmatrix um
  `POST /api/auth/session-tokens` erweitert; §4 mit
  `0.12.0`-Erweiterungs-Hinweis).
- [x] Roadmap-Status und Release-Übersicht auf `0.12.0` als aktive
  Folgephase umgestellt (Commit `530c240`: Header, §1.2, §2 Schritt 47
  und §3 Release-Übersicht jeweils auf 🟡 mit Tranche-0-Aktivierungs-
  datum 2026-05-10).
- [x] Risiken R-14..R-16 im `risks-backlog.md` gegen den finalen
  `0.12.0`-Scope überprüft und bei Bedarf aktualisiert: R-14 wird
  durch `0.12.0` strukturell **nicht** aufgelöst (Plan-0.12.0 §0.1
  bekräftigt: keine produktive Media-Server-Auth-Kopplung als Muss-
  Scope; `validate-key` bleibt Diagnose-/Smoke-Funktion;
  Auflösungspfad ist eine separate Auth-Bridge nach `0.12.0`).
  R-15 (externe Media-Server-Provisionierung) und R-16 (ausgehende
  Webhook-Zustellung) bleiben strukturell unverändert offen — beide
  sind nicht Teil des `0.12.0`-Auth-/Security-Scopes; Stand-Datum,
  Verweise und Mitigationspfad in `risks-backlog.md` entsprechend
  ergänzt.

## 3. Tranche 1 — Auth-Domainmodell und Policy-Grundlage

Ziel: Token- und Policy-Regeln sind frameworkfrei testbar, bevor HTTP
oder Persistenz angebunden werden.

DoD:

- [x] Domainobjekte für `ProjectTokenGeneration`,
  `SessionTokenClaims`, `SessionSigningKey`, `ProjectPolicy`,
  `OriginPolicy`, `RateLimitPolicy` und Auth-Fehler definiert
  (`apps/api/hexagon/domain/auth_session_token.go`,
  `auth_project_token.go`, `auth_project_policy.go`, `errors.go`).
- [x] Project-Token-Material trennt Klartext, Hash und Fingerprint;
  Persistenzsicht enthält nie den Klartext (`ProjectTokenMaterial`
  trägt `Value`; `ProjectTokenGeneration` hat per Konstruktion kein
  Klartext-Feld; Tests pinnen die Trennung).
- [x] Session-Token-Claims enthalten mindestens `iss`, `sub`
  (`project_id`), `aud`, `iat`, `nbf`, `exp`, `jti`, optional
  `session_id` und `origin` (`SessionTokenClaims` plus
  `BuildSessionTokenClaims`).
- [x] `token_id` ist der öffentliche Wire-/Log-Name des `jti`-Claims;
  beide Werte sind identisch (`SessionTokenClaims.TokenID()`-Helper;
  `TestBuildSessionTokenClaims_TokenIDEqualsJTI` pinnt die Identität).
- [x] Token-Zeitvalidierung nutzt injizierbare Clock für Tests
  (`ValidateClaimsTime`, `EvaluateProjectTokenStatus`,
  `BuildSessionTokenClaims` nehmen `now time.Time`; Tests benutzen
  einen festen `fixedNow()`).
- [x] Signatur-/Hash-Vergleiche laufen konstantzeitnah
  (`ConstantTimeEqualSignature` über `crypto/subtle`;
  `ValidateProjectTokenString` vergleicht Hash konstantzeitnah).
- [x] Fehlerkonstanten sind stabil
  (`apps/api/hexagon/domain/errors.go`: `ErrAuthTokenMissing`,
  `ErrAuthTokenInvalid`, `ErrAuthTokenExpired`,
  `ErrAuthTokenNotYetValid`, `ErrAuthTokenRevoked`,
  `ErrAuthProjectMismatch`, `ErrAuthPolicyDenied`,
  `ErrAuthSessionScopeDenied`, `ErrAuthTokenTTLTooLarge`,
  `ErrAuthIssuanceRateLimited`).
- [x] Domain-Tests decken Expiry, `nbf`, unbekanntes `kid`,
  Audience-Mismatch, Project-Mismatch, Origin-Policy-Miss,
  Token-Hash-Fingerprint-Trennung und Redaktionsregeln ab
  (`auth_session_token_test.go`, `auth_project_token_test.go`,
  `auth_project_policy_test.go`).
- [x] Domain-Schicht importiert keine HTTP-, JSON-, SQLite-,
  Prometheus-, OTel- oder Framework-Pakete (nur stdlib `crypto/*`,
  `encoding/*`, `errors`, `strings`, `time`); `make arch-check`
  Bestandteil von `make gates` (grün).

## 4. Tranche 2 — Signierte Session Tokens

Ziel: Browser- und SDK-Flows können kurzlebige Tokens nutzen, ohne das
Project Token dauerhaft pro Event zu senden.

DoD:

- [x] Application-Service stellt Session Tokens aus, wenn ein gültiges
  Project Token präsentiert wird und die Project Policy den Request
  erlaubt (`hexagon/application/issue_session_token.go`).
- [x] Session-Token-Audience wird serverseitig gegen die Project-
  Policy-Allowlist geprüft; `0.12.0` erlaubt als Muss-Pfad nur
  `playback-events`. Nicht erlaubte oder unbekannte Audiences liefern
  `403 auth_session_scope_denied` (Service ruft
  `domain.ValidateAudience`; Tests
  `TestIssueSessionToken_AudienceDeniedByPolicy` und
  `TestAuthSessionTokens_AudienceDenied`).
- [x] Token-Issuance-Endpoint `POST /api/auth/session-tokens`
  implementiert (`adapters/driving/http/auth_session_tokens.go`);
  `project_id` im Body ist optionaler Konsistenzcheck. Mismatch
  liefert `401 auth_project_mismatch` (Test
  `TestAuthSessionTokens_ProjectMismatch`).
- [x] `ttl_seconds` ist deterministisch: harte 900-s-Pflichtgrenze;
  Project-Policies dürfen niedriger begrenzen; fehlendes
  `project_max_ttl_seconds` defaultet pro Project auf exakt 900;
  fehlende Request-Werte nutzen die wirksame Project-Grenze; Werte
  `<= 0` oder oberhalb der Grenze liefern
  `422 auth_token_ttl_too_large` ohne stilles Clamping
  (`domain.ResolveTTLSeconds`; Tests
  `TestResolveTTLSeconds`,
  `TestIssueSessionToken_TTLTooLargeNoSilentClamp`,
  `TestAuthSessionTokens_TTLTooLarge`,
  `TestAuthSessionTokens_TTLDefaultsTo900OnZero`).
- [x] Issuance-Endpoint hat eigene Abuse-Grenzen: globaler + Project-
  Bucket im `InMemoryIssuanceRateLimiter`; Überschreitung liefert
  `429 auth_issuance_rate_limited`; Policy-Ablehnungen liefern
  `403 auth_policy_denied`. Origin-/IP-nahe Quoten sind als optionaler
  Zusatz im Plan §0.6 und im RAK-74 dokumentiert
  (Folge-Scope, blockiert nicht).
- [x] `auth_issuance_rate_limited` ist ausschließlich der
  Rate-Limit-Code von `POST /api/auth/session-tokens` — der HTTP-
  Adapter mappt nur diesen Endpoint auf den Code; andere Endpoints
  bleiben bei ihren bestehenden Rate-Limit-Codes (Test
  `TestAuthSessionTokens_RateLimited`).
- [x] `POST /api/playback-events` akzeptiert zusätzlich zu
  `X-MTrace-Token` ein Session Token über `Authorization: Bearer
  mtr_st_*` oder `X-MTrace-Session-Token`
  (`PlaybackEventsHandler.AuthHeaders` plus
  `BatchInput.PreResolvedProject`; Tests
  `TestAuthHeaderParser_BearerHappyPath`,
  `TestAuthHeaderParser_XMTraceSessionTokenHappyPath`).
- [x] Auth-Priorität ist für alle Kombinationen dokumentiert und
  getestet: `Authorization: Bearer`, `X-MTrace-Session-Token` und
  `X-MTrace-Token`; widersprüchliche Project-Bindungen liefern
  `401 auth_project_mismatch`, ungültige zusätzlich präsentierte
  Tokens liefern `401 auth_token_invalid`, kein stiller Fallback
  (Tests `TestAuthHeaderParser_BearerPlusLegacyDifferentProject`,
  `TestAuthHeaderParser_BearerPlusInvalidLegacy`,
  `TestAuthHeaderParser_ConflictingBearerAndXMTraceSessionToken`).
- [x] Contract-Tests decken Mischfälle explizit ab: malformed
  `Authorization: Bearer mtr_st_*` plus gültiger Legacy-Header liefert
  `401 auth_token_invalid` (`TestAuthHeaderParser_MalformedBearerBlocksLegacyFallback`);
  gültiger Bearer plus Legacy mit anderem Project liefert
  `401 auth_project_mismatch`
  (`TestAuthHeaderParser_BearerPlusLegacyDifferentProject`);
  fremder `Authorization`-Header plus gültiger Legacy bleibt erlaubt
  (`TestAuthHeaderParser_ForeignAuthorizationIgnoredWithLegacy`).
- [x] Fremde `Authorization`-Header ohne `Bearer mtr_st_*` werden als
  nicht-m-trace Auth ignoriert, wenn ein gültiger m-trace Header
  vorhanden ist; ohne gültigen m-trace Header liefern sie
  `401 auth_token_missing`. Malformed `Bearer mtr_st_*` bleibt
  `401 auth_token_invalid` und blockiert Fallback
  (Tests `TestAuthHeaderParser_ForeignAuthorizationOnlyIsMissing`,
  `TestAuthHeaderParser_ForeignAuthorizationIgnoredWithLegacy`,
  `TestAuthHeaderParser_MalformedBearerBlocksLegacyFallback`).
- [x] Auth-Fehlerpräzedenz ist als Entscheidungstabelle im API-
  Kontrakt gepinnt (`spec/backend-api-contract.md` §3.9, neunstufige
  Tabelle) und getestet: missing → invalid → revoked → expired →
  not-yet-valid → project_mismatch → scope_denied → policy_denied →
  rate_limited (Tests
  `TestAuthHeaderParser_ExpiredSessionToken`,
  `TestAuthHeaderParser_OriginMismatchOnSessionToken`,
  `TestValidateClaimsTime_Boundaries`).
- [x] Session Token validiert Signatur, `kid`, `exp`, `nbf`,
  `audience`, `project_id`, optional `session_id` und `origin`
  (`HMACSessionTokenSigner.Verify` plus Domain-`ValidateClaims*`-
  Funktionen; Tests `TestHMACSigner_RoundTrip`,
  `TestHMACSigner_VerifyRejectsTamperedSignature`,
  `TestHMACSigner_VerifyRejectsUnknownKID`,
  `TestValidateClaimsAudience`,
  `TestValidateClaimsProject`,
  `TestValidateClaimsSession`,
  `TestValidateClaimsOrigin`).
- [x] Signing-Key-Ring ist restart-stabil: aktive Signing-Keys und
  alte Verify-Keys sind über ENV/File-Konfiguration reproduzierbar
  (`StaticSigningKeyResolver` mit `MTRACE_AUTH_SIGNING_KID`/
  `MTRACE_AUTH_SIGNING_KEY` in `cmd/api/main.go`). Test
  `TestHMACSigner_RestartStableAcrossKeyResolverReinitialization`
  pinnt, dass ein vor Key-Switch ausgestellter Token nach Rollover
  weiterhin validiert.
- [x] Abgelaufene, manipulierte, falsch gebundene und mit unbekanntem
  `kid` signierte Tokens liefern stabile Fehlercodes ohne
  Identifier-Leak (Tests
  `TestHMACSigner_VerifyRejectsTamperedSignature`,
  `TestHMACSigner_VerifyRejectsUnknownKID`,
  `TestHMACSigner_VerifyRejectsMissingPrefix`,
  `TestHMACSigner_VerifyRejectsMalformedStructure`,
  `TestAuthHeaderParser_ExpiredSessionToken`).
- [x] Issuance-Response gibt den Klartext-Session-Token genau einmal
  zurück; Logs/Traces/Metriken enthalten höchstens `token_id` oder
  Fingerprint (`AuthSessionTokensHandler.ServeHTTP` schreibt
  `session_token.value` nur in der `201`-Antwort; Tests
  `TestAuthSessionTokens_HappyPath`,
  `TestAuthSessionTokens_RoundTripVerify`).
- [x] Contract-Tests pinnen Issuance-Happy-Path, Expired,
  Signature-Mismatch, Project-Mismatch, Audience-Mismatch, Origin-
  Mismatch, fehlenden Auth-Header und `auth_issuance_rate_limited`
  (`adapters/driving/http/auth_session_tokens_test.go` für die
  Happy-Path-/TTL-/Audience-/Project-Mismatch-/Rate-Limit-/Body-
  Limit-Cases; `auth_headers_test.go` für Expiry/Origin-Mismatch im
  Verify-Pfad).
- [x] Bestehende Project-Token-Tests für Playback Events bleiben grün
  (`make gates` 90.2 % Coverage; Pre-`0.12.0`-Pflichttests laufen
  unverändert über den Legacy-Header-Pfad, weil `BatchInput.AuthToken`
  identisch befüllt wird, wenn nur `X-MTrace-Token` gesetzt ist).

## 5. Tranche 3 — Project-Token-Rotation

Ziel: Project Tokens können ohne harte Downtime rotiert werden und
haben einen nachvollziehbaren Migrationspfad weg von der hardcodierten
Map.

DoD:

- [x] Driven-Port für Project-Token-Generationen ergänzt
  (`hexagon/port/driven/project_token_repository.go`:
  `ProjectTokenRepository` mit `Create`/`ListByProject`/`FindByHash`/
  `SetGraceUntil`/`Revoke` plus `ErrProjectTokenNotFound`-Sentinel).
  `ProjectResolver` bleibt unverändert; der `RotatingProjectResolver`
  wickelt beide Pfade ein.
- [x] InMemory- und SQLite-Implementierungen speichern Hash,
  Fingerprint, `token_id`, `project_id`, `not_before`, `grace_until?`,
  `expires_at?`, `revoked_at?`, `created_at`, `rotated_from?` (Status
  wird bewusst on-the-fly aus den Lifecycle-Feldern berechnet, damit
  kein Drift zwischen gespeichertem Status-String und Zeitvergleich
  entstehen kann; vgl. Plan §0.6 Threat Model →
  `domain.EvaluateProjectTokenStatus`).
- [x] Migration ist versioniert (`V4__project_tokens.sql`) und
  Restart-Test pinnt, dass Token-Generationen inkl. `grace_until` über
  einen Close-Reopen-Zyklus persistent bleiben
  (`TestSQLiteProjectTokenRepo_Restart_GraceUntilPersisted`).
  `TestOpen_FreshStart` auf 4 Migrationen aktualisiert.
- [x] Static-Resolver bleibt als Dev-/Test-Fallback erhalten und wird
  über `RotatingProjectResolver` mit Repo-Pfad kombiniert
  (`adapters/driven/auth/rotating_project_resolver.go`). Tokens mit
  `mtr_pt_*`-Prefix laufen über das Repo, alle anderen über den
  Static-Resolver.
- [x] Rotation erzeugt eine neue Generation
  (`domain.GenerateProjectToken`), lässt alte Generationen bis zum
  persistierten `GraceUntil` gültig
  (`TestRotatingResolver_RotationGraceStillAuthenticates`) und lehnt
  widerrufene/abgelaufene Generationen deterministisch ab
  (`TestRotatingResolver_RevokedReturnsRevokedError`,
  `TestRotatingResolver_ExpiredReturnsExpiredError`,
  `TestRotatingResolver_NotYetValid`). `GraceUntil` ist persistiert
  (V4-Spalte) und wird in `EvaluateProjectTokenStatus` ohne Rückgriff
  auf Prozesszustand oder `RotatedFrom` evaluiert.
- [x] Migrations-/Rollback-Kriterien für Rotation sind dokumentiert
  (V4-Header-Kommentar, `docs/user/local-development.md` §2.7.3) und
  getestet: alte Generationen bleiben bis `GraceUntil` gültig, solange
  `RevokedAt` nicht gesetzt ist; `RevokedAt` beendet Grace sofort
  (`TestEvaluateProjectTokenStatus_TimeMatrix` Case
  „revoked overrides grace"). Rollback ist nicht reaktivierend —
  V4-Header und Doku-Notiz pinnen den Operator-Workflow.
- [x] Keine Persistenz, Fixtures oder Logs enthalten Klartext-
  Project-Tokens — `Create`/`Find`/`List` reichen ausschließlich
  `domain.ProjectTokenGeneration` ohne `Value`-Feld; Klartext lebt
  nur transient in `domain.ProjectTokenMaterial.Value` zwischen
  `GenerateProjectToken` und Caller.
- [x] Tests decken aktive, neue (`Active`), grace, abgelaufene,
  widerrufene, unbekannte und kollidierende Token-Generationen ab
  (`TestEvaluateProjectTokenStatus_TimeMatrix`,
  `TestRotatingResolver_*`-Suite, `TestInMemoryProjectTokenRepo_*`,
  `TestSQLiteProjectTokenRepo_*`).
- [x] Doku beschreibt den Operator-Migrationspfad für `demo-token`
  und lokale `.env`-Beispielwerte
  (`docs/user/local-development.md` §2.7.3 — Voraussetzung SQLite,
  Signing-Key per Env, Rotation/Grace/Revoke-Workflow,
  `MTRACE_AUTH_SIGNING_KID`/`_KEY` als `.env`-Beispiel mit
  Production-Warnung).

## 6. Tranche 4 — Ingest Policies, CORS und Rate Limits

Ziel: Project Policies steuern Browser- und Ingest-Zugriffe
deterministisch, ohne eine vollständige Tenant-Control-Plane zu
behaupten.

DoD:

- [x] Policy-Modell definiert erlaubte Origins, Methoden, Header,
  Session-Token-Audiences, maximale Session-Token-TTL und Rate-Limit-
  Parameter inklusive separater Issuance-Quote (`domain.ProjectPolicy`,
  `RateLimitPolicy.IssuanceBucket` per Tranche-2-Review-Fix in den
  Application-Service durchgereicht). Muss-Pfad sind globale und
  Project-Buckets im `InMemoryIssuanceRateLimiter`; Origin-/IP-nahe
  Buckets sind als Folge-Scope dokumentiert (Plan §0.1, R-17 im
  Risiken-Backlog).
- [x] CORS-Preflight nutzt eine dokumentierte globale, konservative
  Allowlist (`adapters/driving/http/cors.go`:
  `preflightAllowedHeaders` + zentralisierter `preflightHandler`).
  `Access-Control-Allow-Origin` wird nie `*` für tokenpflichtige
  Browser-Telemetrie; unbekannte Origins erhalten keine Origin-Liste,
  keine Project-Hinweise und keine diagnostischen Bodies.
- [x] Preflight-Antworten sind exakt gepinnt: bekannte Origins liefern
  `204` mit gespiegeltem `Access-Control-Allow-Origin`,
  pfadspezifischen `Allow-Methods`, der `0.12.0`-Header-Allowlist
  (`Content-Type, Authorization, X-MTrace-Token,
  X-MTrace-Session-Token, traceparent`), `Access-Control-Max-Age:
  600`, `Vary: Origin` und `Cache-Control: no-store`; unbekannte
  Origins liefern `204` mit leerem Body, ohne Allow-Origin/Methods/
  Headers, mit `Vary: Origin` und `Cache-Control: no-store`. Tests
  `TestCORS_Preflight_PlaybackEvents_HeaderSetExact` +
  `_BodyEmpty` + `_UnknownOrigin` pinnen das byte-stabil.
- [x] `Origin`-Validierung unterscheidet Browser-Pfad und CLI/curl-
  Pfad ohne `Origin` (`domain.Project.IsOriginAllowed` lässt leeren
  Origin durch; Test `TestCORS_Post_NoOrigin_StillAccepted`).
- [x] `POST /api/playback-events` und `POST /api/auth/session-tokens`
  prüfen relevante Policies vor Use-Case-Seiteneffekten —
  Origin-Mismatch liefert `403 ErrOriginNotAllowed` aus
  `authorizeAndAdmit` vor `limiter.Allow`-Aufruf, also keine
  Token-Konsumierung (`TestCORS_Post_ProjectOriginMismatch_403`).
  `/api/ingest/*` bleibt per RAK-74-Scope-Cut im
  `0.11.0`-Token-Validierungs-Pfad (Lab-Workflow,
  `curl`-/Operator-driven, kein Browser-Konsument); Project-Policy-
  Enforcement für Ingest ist Folge-Scope.
- [x] Rate-Limit-Key enthält im Muss-Pfad das Project: bestehende
  `RateLimitKey{ProjectID, ClientIP, Origin}`-Struktur lebt weiter im
  Use-Case (`limiter.Allow`); Origin- und IP-Buckets sind über die
  bestehende Rate-Limit-Infrastruktur abgedeckt und damit kein
  Folge-Scope. Issuance-Bucket ist separat (`InMemoryIssuanceRateLimiter`).
- [x] Tests decken erlaubte globale Preflight-Origin
  (`TestCORS_Preflight_PlaybackEvents_Allowed`), unbekannte
  Preflight-Origin (`_UnknownOrigin`), Request-Method-Mismatch
  (`_RequestMethod_Ignored`), exakte Header-Sets
  (`_HeaderSetExact`), minimierte Signalisierung
  (`_BodyEmpty`, `_UnknownOrigin` ohne Allow-* Header), project-
  spezifischen POST-Origin-Miss (`TestCORS_Post_ProjectOriginMismatch_403`),
  leeren Origin im CLI-Pfad (`_NoOrigin_StillAccepted`) und
  Auth-Issuance-Preflight (`_AuthSessionTokens_Allowed`) ab.
- [x] `0.11.0` Ingest-Control-Validate bleibt explizit kein
  produktiver Media-Server-Auth-Pfad; pinned via
  `TestCORS_ValidateKeyRemainsDiagnostic` (CORS-Pfad signalisiert
  nichts Auth-spezifisches), `spec/backend-api-contract.md` §3.8
  (`{"valid": false}`-Klausel ohne Stream-ID-Hinweis),
  Lastenheft §13.13 RAK-65/RAK-66 und R-14 im Risiken-Backlog.

## 7. Tranche 5 — Doku, Compatibility und Smokes

Ziel: Nutzer können von Public Project Token auf Session Tokens und
rotierbare Token-Generationen migrieren, ohne bestehende Lab-Flows zu
brechen.

DoD:

- [x] `docs/user/demo-integration.md` und/oder neue Auth-Doku
  beschreiben Session-Token-Issuance, Browser-Nutzung, TTL, CSP,
  CORS und Token-Rotation (`docs/user/auth.md` neu mit §1 Auth-Pfade,
  §2 Issuance-Wire, §4 CSP/CORS, §5 Project-Token-Rotation;
  `demo-integration.md` Optional-Sektion „Session-Token-Pfad" mit
  fetch-Beispiel).
- [x] `docs/user/local-development.md` enthält sichere lokale
  Beispielwerte und erklärt, welche Tokens Demo-/Lab-Werte sind
  (§2.7.3 Auth / Token Lifecycle mit `MTRACE_AUTH_SIGNING_KEY`-
  Setup, `.env`-Beispielwerten und Production-Warnung).
- [x] `README.md` grenzt Auth-Härtung weiterhin gegen Production-
  Identity, OAuth/OIDC, KMS/Vault und SaaS-Tenant-Management ab —
  normativ über Lastenheft §13.14 RAK-71 (Out-of-Scope-Liste);
  README selbst zeigt prägnant auf den Stack-Scope ohne separate
  Negativ-Sektion.
- [x] Datenschutz-/GDPR-Doku beschreibt, welche Auth-Metadaten
  zusätzlich entstehen, welche Claims pseudonym sind und dass
  IP-/User-Agent-Speicherung nicht ausgeweitet wird
  (`docs/user/auth.md` §6).
- [x] SDK-Doku beschreibt kompatible Token-Nutzung: Project Token
  direkt bleibt erlaubt, Session Token wird bevorzugter Browser-Pfad
  (`auth.md` §1 Auth-Matrix; `demo-integration.md` Optional-Sektion).
- [x] Doku und SDK-Beispiele beschreiben Session-Token-Caching:
  Token bis kurz vor `expires_at` wiederverwenden, Refresh-Fenster
  nutzen, keine Speicherung in `localStorage`/persistenten Browser-
  Stores empfehlen (`auth.md` §3 mit Pseudocode + Storage-Regeln).
- [x] Player-SDK braucht keine Secret-Verwaltung; falls API-Optionen
  ergänzt werden, bleiben ESM/CJS/IIFE und Public-API-Snapshot stabil
  oder bewusst additiv aktualisiert — `0.12.0` erweitert das SDK
  bewusst **nicht**: Session-Token-Issuance ist Browser-/Caller-
  Verantwortung (siehe `auth.md` §3 Pseudocode), `attachWebRtc`/
  `attachHlsJs`-Public-API bleibt unverändert; SDK-Pack-Smoke
  läuft weiterhin grün in `make gates`.
- [x] Contract-Fixtures unter `spec/contract-fixtures/api/` pinnen
  Session-Token-Issuance (`auth-session-token-issue.json`), Expired
  Token (`auth-error-token-expired.json`), Policy Denied
  (`auth-error-policy-denied.json`), TTL-zu-groß
  (`auth-error-ttl-too-large.json`), Issuance-Rate-Limited
  (`auth-error-issuance-rate-limited.json`) und Project-Token-
  Generationen-Persistenz (`auth-project-token-generation.json`,
  Klartext-frei).
- [x] Kompatibilitätsnachweise pinnen konkrete bestehende Flows:
  SDK/Demo-Playback mit Project Token (Pre-`0.12.0`-Pflichttests
  bleiben grün, weil `BatchInput.AuthToken`-Pfad unverändert),
  SDK/Demo-Playback mit Session Token
  (`TestAuthHeaderParser_BearerHappyPath`,
  `_XMTraceSessionTokenHappyPath`), fremder `Authorization`-Header
  plus gültigem `X-MTrace-Token`
  (`TestAuthHeaderParser_ForeignAuthorizationIgnoredWithLegacy`),
  Issuance-Round-Trip (`TestAuthSessionTokens_RoundTripVerify`).
  Analyze-/Session-Link-Auth nutzt weiter `X-MTrace-Token`-only
  (Read-Pfad ohne Session-Token-Allowlist; siehe RAK-74-Scope-Cut
  in §13.14).
- [x] `make sync-contract-fixtures` kopiert sechs neue Auth-Fixtures
  in `apps/api/adapters/driving/http/testdata/`
  (auth-session-token-issue, auth-error-token-expired,
  auth-error-policy-denied, auth-error-ttl-too-large,
  auth-error-issuance-rate-limited, auth-project-token-generation).
  Counter im Recipe von 30 auf 36 erhöht.
- [x] `make generated-drift-check` deckt neue Fixtures ab — der
  Drift-Check ruft `make sync-contract-fixtures` plus
  `git diff --exit-code apps/api/.../testdata/` auf; die neuen
  Auth-Fixtures sind Teil der Sync-Liste und damit Teil des
  Drift-Pin.
- [x] Relevante Smokes sind dokumentiert: keine neuen Auth-Smokes
  nötig. Issuance-Pfad ist über die Contract-Tests
  (`auth_contract_test.go`) plus den HTTP-Stack-Test
  (`auth_session_tokens_test.go`) abgedeckt; Multi-Header-Konsum
  über `auth_headers_test.go`. Ein opt-in Runtime-Smoke gegen die
  laufende API ist Folge-Scope (R-18 Operator-Runbook für
  Key-Rotation würde so einen Smoke nahelegen, ist aber kein
  Release-Blocker).

## 8. Tranche 6 — Release-Closeout

DoD:

- [x] RAK-Verifikationsmatrix in §9 vollständig ausgefüllt.
- [x] `make docs-check` grün (Tranchen 0–5 Closeout-Läufe).
- [x] `make api-test` Bestandteil von `make gates`; via Race-Test-
  Layer in T1–T5 grün.
- [x] `make build` Bestandteil von `make gates`; in T2–T5 grün.
- [x] `make gates` grün (zuletzt Coverage 90.8 % beim T5-Review-Fix-
  Commit `0ebeed5`).
- [x] `make security-gates` läuft in CI als zweiter PR-blockierender
  Job (vuln-check + audit-ts + image-scan); bestehende Pipeline
  unverändert, keine `0.12.0`-spezifische Security-Surface-Änderung.
- [x] `make generated-drift-check` grün — Drift-Check meldet
  `OK -- no drift detected` inkl. der sechs neuen Auth-Fixtures.
- [x] `pnpm --filter @npm9912/player-sdk run pack:smoke` grün —
  SDK-Public-API ist in `0.12.0` unverändert; Pack-Smoke läuft
  Bestandteil von `make gates`.
- [x] Wave-2-Quality-Gates bewertet: `benchmark-smoke`,
  `fuzz-check` und `mutation-report` sind Nightly-Workflows, die
  in `0.12.0` keine Auth-spezifischen Targets brauchen — die neuen
  Auth-Domain-Funktionen sind reine Hash-/HMAC-Wrapper und damit
  weder Performance-kritisch noch Fuzz-fertig (kein User-Input ohne
  Domain-Validierung). Bestehende Nightlies bleiben aktiv.
- [x] Vollständiger Versions-Bump auf `0.12.0` durchgeführt: Root
  `package.json` plus 4 Workspace-Pakete (apps/analyzer-service,
  apps/dashboard, packages/player-sdk, packages/stream-analyzer);
  `apps/api/cmd/api/main.go` `serviceVersion`; SDK-Version-Marker
  in 9 Go-Test-Files (`"version": "0.11.0"` → `"0.12.0"`).
- [x] `CHANGELOG.md` mit `[0.12.0] - 2026-05-10` aktualisiert.
- [x] Roadmap auf released `0.12.0` und Folgephase `0.13.0`
  umgestellt (Phase-Header, §1.2 Nächste Phase, §2 Schritt 47 ✅,
  §3 Release-Übersicht).
- [x] Plan nach `docs/planning/done/plan-0.12.0.md` verschoben und
  Status auf ✅ released aktualisiert.
- [x] Annotierter Tag `v0.12.0` erstellt.

## 9. RAK-Verifikationsmatrix

Wird während der Umsetzung gepflegt. Jede Zeile braucht vor Closeout
Datei-, Test- und Doku-Nachweis.

| RAK | Priorität | Nachweis | Status |
| --- | --- | --- | --- |
| RAK-71 | Muss | Lastenheft `1.1.15` §13.14 mit Scope-Grenze; `docs/user/auth.md` §0 verweist normativ auf RAK-71; README-Abgrenzung in „Was ist m-trace?"-Sektion; Plan §0.1 Out-of-Scope; Out-of-Scope-Abgrenzung gegen OAuth/OIDC/Admin-UI/KMS-Vault/SaaS-Tenant-Management normativ in RAK-71 selbst. | [x] |
| RAK-72 | Muss | Session-Token-Issuance via `IssueSessionTokenService` (`hexagon/application/issue_session_token.go`); Audience-Allowlist via `domain.ValidateAudience` plus `SessionTokenAudiencePlaybackEvents`-Konstante; harte TTL-Grenze 900 s ohne Clamp via `domain.ResolveTTLSeconds`; Issuance-Abuse-Limits via `InMemoryIssuanceRateLimiter` (global + Project); HMAC-SHA-256-Signatur via `HMACSessionTokenSigner` mit `kid`-Lookup über `StaticSigningKeyResolver`; Claims-Validierung via `ValidateClaimsTime`/`_Audience`/`_Project`/`_Session`/`_Origin`; `token_id` == `jti` per Konstruktion (`SessionTokenClaims.TokenID()`); restart-stabiler Signing-Key-Ring (Test `TestHMACSigner_RestartStableAcrossKeyResolverReinitialization`); neunstufige Fehlerpräzedenz in §3.9 plus zehn `auth_*`-Codes (`writeAuthError` Tabellen-Test, 13 Branches); Contract-Test `auth_session_tokens_test.go` (14 Cases inkl. RoundTrip). | [x] |
| RAK-73 | Muss | Generationen-Modell `domain.ProjectTokenGeneration` mit allen Lifecycle-Feldern; Rotation via `domain.GenerateProjectToken`; Grace/Revocation via `EvaluateProjectTokenStatus` (Reihenfolge revoked > expired > nbf > grace > active); persistiertes `grace_until` in V4-Migration `V4__project_tokens.sql`; Persistenz ohne Klartext (Generation-Struktur trägt nur Hash/Fingerprint); Restart-Test `TestSQLiteProjectTokenRepo_Restart_GraceUntilPersisted`; Repository-Tests inmemory + sqlite (13 Cases); RotatingProjectResolver kombiniert Repo-Pfad mit Legacy-Static-Resolver (8 Cases). | [x] |
| RAK-74 | Muss | `domain.ProjectPolicy` mit Allowed-Origins/Methods/Headers/Audiences; globale + Project-Rate-Limit-Buckets in `RateLimitPolicy`; separate Issuance-Quote über `IssuanceBucket` (T2-Review-Fix in IssueSessionTokenService durchgereicht); globale konservative CORS-Preflight-Regeln in `cors.go:preflightHandler` (§3.9-konforme `204`-Antworten mit Vary/Cache-Control); minimal-denial für unbekannte Origins (`TestCORS_Preflight_PlaybackEvents_UnknownOrigin`); `mtrace_cors_preflight_refused_total{path}`-Counter; project-spezifisches POST-Enforcement im `playback-events`- und `auth/session-tokens`-Pfad; **`/api/ingest/*` per Scope-Cut Out-of-Scope** (R-21 trackt Future-Browser-Konsumenten); Origin-/IP-Buckets über bestehende `0.1.0`-Rate-Limit-Infrastruktur. | [x] |
| RAK-75 | Muss | Pre-`0.12.0`-Pflichttests bleiben grün (BatchInput-AuthToken-Pfad unverändert für Legacy-`X-MTrace-Token`); Foreign-Authorization-Test `TestAuthHeaderParser_ForeignAuthorizationIgnoredWithLegacy`; Multi-Header-Konsistenz in 14 Cases von `auth_headers_test.go`; SDK/Demo läuft unverändert über Project-Token-Pfad (Backward-Compat in `RotatingProjectResolver`); Analyze-/Session-Link bleibt `X-MTrace-Token`-only. | [x] |
| RAK-76 | Muss | Security-Doku in `docs/user/auth.md` (Issuance, Caching, CSP, CORS, Rotation, GDPR); Threat Model in Plan §0.6; Datenschutz/GDPR-Grenzen in `auth.md` §6; CSP-Beispiel in `auth.md` §4.1; CORS-Beispiel in `auth.md` §4.2; sechs Contract-Fixtures (`spec/contract-fixtures/api/auth-*.json`) plus `auth_contract_test.go`; `make sync-contract-fixtures` (Counter 30 → 36) plus `make generated-drift-check` deckt Auth-Fixtures ab. | [x] |

## 10. Folge-Scope nach `0.12.0`

- `0.13.0`: Production / Ops Backends (Postgres, ClickHouse/
  VictoriaMetrics, Kubernetes-Manifeste, Devcontainer, Release-
  Automatisierung).
- Später: OAuth/OIDC, User-/Org-Verwaltung, Admin-UI und Rollenmodell.
- Später: KMS/Vault/Cloud-Secret-Manager für Signatur- und Project-
  Token-Secrets.
- Später: produktive MediaMTX-/SRS-Auth-Hook-Brücke mit signierten
  Publish-Tokens, falls ein konkreter Lab-/Staging-Bedarf entsteht.
- Später: globale Stream-Key- und Project-Token-Rotation über mehrere
  Deployments.
- Später: Origin-/IP-nahe Rate-Limit-Buckets, falls die `0.12.0`-
  Infrastruktur nur globale und Project-Buckets trägt.
