# Implementation Plan — `0.12.0` (Auth / Token Lifecycle)

> **Status**: ⬜ open — noch nicht aktiviert. Dieser Plan darf erst nach
> explizitem Move nach `docs/planning/in-progress/` umgesetzt werden.
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
> [`spec/lastenheft.md`](../../../spec/lastenheft.md) §7.11
> `F-106`..`F-113`, §8.3 `NF-16`, §8.5 `NF-24`..`NF-29`,
> §8.5 CORS/CSP `NF-30`..`NF-37`, §8.6 Datenschutz/GDPR
> `NF-38`..`NF-44`, §15.1 Risiko „Project Token im Browser-Code wird
> zweckentfremdet";
> [`spec/backend-api-contract.md`](../../../spec/backend-api-contract.md)
> §4 Authentifizierung und §5 Validierungsreihenfolge;
> [`docs/planning/done/plan-0.11.0.md`](../done/plan-0.11.0.md) §10
> Folge-Scope nach `0.11.0`;
> [`docs/planning/in-progress/risks-backlog.md`](../in-progress/risks-backlog.md)
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
    Rate-Limit-Buckets pro Project/Origin/IP-Bereich und optionale
    Ingest-Grenzen für den `0.11.0`-Stream-Control-Pfad.
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
| `POST` | `/api/ingest/*` | bleibt Project-tokenpflichtig; Policy-Grenzen werden zusätzlich geprüft |

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
| 9 | Issuance-/Auth-Rate-Limit überschritten | `429 auth_issuance_rate_limited` oder endpoint-spezifischer Rate-Limit-Code |

Diese Präzedenz verhindert stillen Fallback: ein ungültiger
höher-priorisierter Token neben einem gültigen niedriger-priorisierten
Token liefert weiter `401 auth_token_invalid`. Project-Mismatch wird
nur geprüft, wenn alle präsentierten Tokens vorher syntaktisch,
kryptografisch und zeitlich gültig sind. Fremde `Authorization`-
Header ohne `mtr_st_`-Bearer-Token sind kein m-trace Auth-Versuch und
blockieren den Legacy-Project-Token-Pfad nicht.

**CORS-Preflight-Modell.**

Browser-Preflights enthalten in der Praxis kein Project- oder Session-
Token, das der Server verlässlich validieren kann. Deshalb nutzt
`0.12.0` für `OPTIONS` eine globale, konservative und
informationsarme Allowlist:

- erlaubte Methoden maximal `POST, OPTIONS`;
- erlaubte Header maximal `Content-Type`, `Authorization`,
  `X-MTrace-Token`, `X-MTrace-Session-Token`, `X-MTrace-Project`,
  `traceparent`;
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
definieren, aber keine höhere. Fehlt `ttl_seconds`, verwendet der
Server `min(project_max_ttl_seconds, 900)`. Ist `ttl_seconds` kleiner
oder gleich 0 oder größer als die wirksame Project-Grenze, liefert die
API `422 auth_token_ttl_too_large`; es gibt keinen stillen Clamp.

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
  Missbrauchspfad. Auch mit gültigem Project Token gelten globale,
  Project-, Origin- und IP-nahe Quoten; Überschreitung liefert
  `429 auth_issuance_rate_limited`. Token-Issuance darf keine
  unbegrenzten Logs, DB-Zeilen oder teuren Signaturpfade erzeugen.
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
- **Rate Limits**: Policy-Grenzen gelten mindestens pro Project; wo
  technisch vorhanden zusätzlich pro Origin/IP-Bereich.
- **Abgrenzung**: keine Production-Secret-Verwaltung, kein OAuth/OIDC,
  kein Media-Server-Auth-Hook-Versprechen.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| --- | --- | --- |
| 0 | Plan-Aktivierung, Lastenheft-Patch, RAK-Gruppe, Architektur-/Persistenzentscheidung und Threat Model | ⬜ |
| 1 | Auth-Domainmodell, Token-Generationen und Project Policies | ⬜ |
| 2 | Signierte Session Tokens (`F-111`) und Auth-Wire-Vertrag | ⬜ |
| 3 | Project-Token-Rotation (`F-112`) mit SQLite-/InMemory-Persistenz | ⬜ |
| 4 | Ingest Policies (`F-113`), CORS/Preflight und Rate-Limit-Integration | ⬜ |
| 5 | SDK/API-Kompatibilität, Doku, Contract-Fixtures und Smokes | ⬜ |
| 6 | Gates, RAK-Verifikationsmatrix, Versions-Bump, Closeout und Tag | ⬜ |

---

## 2. Tranche 0 — Aktivierung, Patch und Entscheidungen

Ziel: Der Security-Scope ist normativ geschlossen, bevor Auth-Code
geändert wird.

DoD:

- [ ] Plan von `docs/planning/open/plan-0.12.0.md` nach
  `docs/planning/in-progress/plan-0.12.0.md` verschoben.
- [ ] `git status --short` vor erster Änderung dokumentiert.
- [ ] Lastenheft-Header auf `1.1.15` erhöht.
- [ ] `F-111`..`F-113` im Lastenheft für den begrenzten `0.12.0`-
  Scope nachvollziehbar von Kann-Historie auf Release-Muss
  abgebildet.
- [ ] RAK-71..RAK-76 im Lastenheft ergänzt.
- [ ] Patch-Log im aktivierten `plan-0.12.0.md` geführt und im
  Closeout über Roadmap, `CHANGELOG.md` und
  `docs/planning/done/plan-0.12.0.md` tracebar gemacht. Historische
  Pläne wie `plan-0.1.0.md` werden nicht als Pflichtziel für diesen
  Release-Patch verwendet.
- [ ] Architekturentscheidung dokumentiert: Auth bleibt in `apps/api`
  oder explizit begründete Abweichung.
- [ ] Persistenzentscheidung dokumentiert: SQLite + InMemory-Testpfad
  oder explizit begründete Config-only-Variante.
- [ ] Threat-Model-Notiz mit Public Browser Token, Replay, Leakage,
  Privacy, Origin-Bindung, Signatur-Key-Rotation und Logging-Grenzen
  ergänzt.
- [ ] `spec/backend-api-contract.md` bekommt einen `0.12.0`-
  Abschnitt mit Auth-Matrix, Token-Issuance-Wire, Fehlercodes und
  Validierungsreihenfolge.
- [ ] Roadmap-Status und Release-Übersicht auf `0.12.0` als aktive
  Folgephase umgestellt.
- [ ] Risiken R-14..R-16 im `risks-backlog.md` gegen den finalen
  `0.12.0`-Scope überprüft und bei Bedarf aktualisiert.

## 3. Tranche 1 — Auth-Domainmodell und Policy-Grundlage

Ziel: Token- und Policy-Regeln sind frameworkfrei testbar, bevor HTTP
oder Persistenz angebunden werden.

DoD:

- [ ] Domainobjekte für `ProjectTokenGeneration`,
  `SessionTokenClaims`, `SessionSigningKey`, `ProjectPolicy`,
  `OriginPolicy`, `RateLimitPolicy` und Auth-Fehler definiert.
- [ ] Project-Token-Material trennt Klartext, Hash und Fingerprint;
  Persistenzsicht enthält nie den Klartext.
- [ ] Session-Token-Claims enthalten mindestens `iss`, `sub`
  (`project_id`), `aud`, `iat`, `nbf`, `exp`, `jti`, optional
  `session_id` und `origin`.
- [ ] `token_id` ist der öffentliche Wire-/Log-Name des `jti`-Claims;
  beide Werte sind identisch. Tests pinnen, dass `jti` nur in
  signierten Claims und `token_id` in Response, Logs und Contract-
  Fixtures verwendet wird.
- [ ] Token-Zeitvalidierung nutzt injizierbare Clock für Tests.
- [ ] Signatur-/Hash-Vergleiche laufen konstantzeitnah, soweit die
  verwendete Primitive das erlaubt.
- [ ] Fehlerkonstanten sind stabil, z. B.
  `ErrAuthTokenMissing`, `ErrAuthTokenInvalid`,
  `ErrAuthTokenExpired`, `ErrAuthTokenNotYetValid`,
  `ErrAuthTokenRevoked`, `ErrAuthProjectMismatch`,
  `ErrAuthPolicyDenied`, `ErrAuthSessionScopeDenied`.
- [ ] Domain-Tests decken Expiry, `nbf`, unbekanntes `kid`,
  Audience-Mismatch, Project-Mismatch, Origin-Policy-Miss,
  Token-Hash-Fingerprint-Trennung und Redaktionsregeln ab.
- [ ] Domain-Schicht importiert keine HTTP-, JSON-, SQLite-,
  Prometheus-, OTel- oder Framework-Pakete.

## 4. Tranche 2 — Signierte Session Tokens

Ziel: Browser- und SDK-Flows können kurzlebige Tokens nutzen, ohne das
Project Token dauerhaft pro Event zu senden.

DoD:

- [ ] Application-Service stellt Session Tokens aus, wenn ein gültiges
  Project Token präsentiert wird und die Project Policy den Request
  erlaubt.
- [ ] Session-Token-Audience wird serverseitig gegen die Project-
  Policy-Allowlist geprüft; `0.12.0` erlaubt als Muss-Pfad nur
  `playback-events`. Nicht erlaubte oder unbekannte Audiences liefern
  `403 auth_session_scope_denied`.
- [ ] Token-Issuance-Endpoint
  `POST /api/auth/session-tokens` implementiert; `project_id` im Body
  ist optionaler Konsistenzcheck zum Project Token. Fehlt
  `project_id`, wird das Project aus dem Token abgeleitet und in der
  Response zurückgegeben; ein gesetzter Mismatch liefert
  `401 auth_project_mismatch`.
- [ ] `ttl_seconds` ist deterministisch: maximale Pflichtgrenze 900
  Sekunden, Project Policies dürfen nur niedriger begrenzen, fehlende
  Werte nutzen die wirksame Project-Grenze, und Werte `<= 0` oder
  oberhalb der wirksamen Grenze liefern `422
  auth_token_ttl_too_large` ohne stilles Clamping.
- [ ] Issuance-Endpoint hat eigene Abuse-Grenzen: mindestens globale
  und Project-Quote, nach Möglichkeit zusätzlich Origin/IP-nahe Quote.
  Überschreitungen liefern `429 auth_issuance_rate_limited`; Policy-
  Ablehnungen liefern `403 auth_policy_denied`.
- [ ] `POST /api/playback-events` akzeptiert zusätzlich zu
  `X-MTrace-Token` ein Session Token über `Authorization: Bearer` oder
  `X-MTrace-Session-Token`.
- [ ] Auth-Priorität ist für alle Kombinationen dokumentiert und
  getestet: `Authorization: Bearer`, `X-MTrace-Session-Token` und
  `X-MTrace-Token`; widersprüchliche Project-Bindungen liefern
  `401 auth_project_mismatch`, ungültige zusätzlich präsentierte
  Tokens liefern `401 auth_token_invalid`, und es gibt keinen stillen
  Fallback auf niedriger priorisierte gültige Tokens.
- [ ] Fremde `Authorization`-Header ohne `Bearer mtr_st_*` werden als
  nicht-m-trace Auth ignoriert, wenn ein gültiger m-trace Header
  vorhanden ist; ohne gültigen m-trace Header liefern sie
  `401 auth_token_missing`. Malformed `Bearer mtr_st_*` bleibt
  `401 auth_token_invalid` und blockiert Fallback.
- [ ] Auth-Fehlerpräzedenz ist als Entscheidungstabelle im API-
  Kontrakt gepinnt und getestet: malformed/invalid vor revoked vor
  expired vor not-yet-valid vor Project-Mismatch vor Scope-/Policy-
  Denial vor Rate-Limit.
- [ ] Session Token validiert Signatur, `kid`, `exp`, `nbf`,
  `audience`, `project_id`, optional `session_id` und `origin`.
- [ ] Signing-Key-Ring ist restart-stabil: aktive Signing-Keys und
  alte Verify-Keys sind über ENV/File-Konfiguration reproduzierbar.
  Tests pinnen, dass ein vor Key-Switch ausgestellter, noch nicht
  abgelaufener Session Token nach Wechsel des aktiven `kid` und nach
  Reinitialisierung weiterhin validiert wird.
- [ ] Abgelaufene, manipulierte, falsch gebundene und mit unbekanntem
  `kid` signierte Tokens liefern stabile Fehlercodes ohne
  Identifier-Leak.
- [ ] Issuance-Response gibt den Klartext-Session-Token genau einmal
  zurück; Logs/Traces/Metriken enthalten höchstens `token_id` oder
  Fingerprint.
- [ ] Contract-Tests pinnen Issuance-Happy-Path, Expired,
  Signature-Mismatch, Project-Mismatch, Audience-Mismatch, Origin-
  Mismatch, fehlenden Auth-Header und
  `auth_issuance_rate_limited`.
- [ ] Bestehende Project-Token-Tests für Playback Events bleiben grün.

## 5. Tranche 3 — Project-Token-Rotation

Ziel: Project Tokens können ohne harte Downtime rotiert werden und
haben einen nachvollziehbaren Migrationspfad weg von der hardcodierten
Map.

DoD:

- [ ] Driven-Port für Project-Token-Generationen ergänzt oder
  bestehender `ProjectResolver` so erweitert, dass mehrere aktive/
  grace/deaktivierte Generationen pro Project möglich sind.
- [ ] InMemory- und SQLite-Implementierungen speichern Hash,
  Fingerprint, `token_id`, `project_id`, Status, `not_before`,
  `grace_until?`, `expires_at?`, `revoked_at?`, `created_at`,
  `rotated_from?`.
- [ ] Migration ist versioniert und Restart-Test weist nach, dass
  Token-Generationen inklusive `grace_until` persistent bleiben.
- [ ] Static-Resolver aus der bisherigen hardcodierten Map bleibt als
  Dev-/Test-Fallback oder wird über einen klaren Kompatibilitäts-
  Adapter abgelöst.
- [ ] Rotation erzeugt eine neue Generation, lässt alte Generationen
  optional bis zum persistierten `grace_until` gültig und lehnt
  widerrufene/abgelaufene Generationen deterministisch ab. `grace_until`
  darf nicht aus volatilem Prozesszustand oder aus `rotated_from`
  rekonstruiert werden.
- [ ] Keine Persistenz, Fixtures oder Logs enthalten Klartext-Project-
  Tokens.
- [ ] Tests decken aktive, neue, grace, abgelaufene, widerrufene,
  unbekannte und kollidierende Token-Generationen ab.
- [ ] Doku beschreibt den Operator-Migrationspfad für `demo-token` und
  lokale `.env`-Beispielwerte.

## 6. Tranche 4 — Ingest Policies, CORS und Rate Limits

Ziel: Project Policies steuern Browser- und Ingest-Zugriffe
deterministisch, ohne eine vollständige Tenant-Control-Plane zu
behaupten.

DoD:

- [ ] Policy-Modell definiert erlaubte Origins, Methoden, Header,
  Session-Token-Audiences, maximale Session-Token-TTL und Rate-Limit-
  Parameter inklusive separater Issuance-Quote.
- [ ] CORS-Preflight nutzt eine dokumentierte globale, konservative
  Allowlist, weil `OPTIONS` ohne validierbares Project-/Session-Token
  kein deterministisches Project-Enforcement erlaubt. Project-
  spezifische Policy-Erzwingung erfolgt beim tatsächlichen `POST`.
  `Access-Control-Allow-Origin` wird nie `*` für tokenpflichtige
  Browser-Telemetrie. Unbekannte Origins erhalten keine Origin-Liste,
  keine Project-Hinweise und keine diagnostischen Bodies.
- [ ] Preflight-Antworten sind exakt gepinnt: bekannte Origins liefern
  `204` mit leerem Body, gespiegeltem `Access-Control-Allow-Origin`,
  erlaubten Methoden/Headern, `Access-Control-Max-Age: 600`,
  `Vary: Origin` und `Cache-Control: no-store`; unbekannte Origins
  liefern `204` mit leerem Body, ohne Allow-Origin/Methods/Headers,
  aber ebenfalls mit `Vary: Origin` und `Cache-Control: no-store`.
- [ ] `Origin`-Validierung unterscheidet Browser-Pfad und CLI/curl-
  Pfad ohne `Origin` gemäß API-Kontrakt.
- [ ] `POST /api/playback-events` und `/api/ingest/*` prüfen
  relevante Policies vor Use-Case-Seiteneffekten.
- [ ] Rate-Limit-Key wird mindestens um Origin/IP-Bereich erweitert,
  soweit die vorhandene Rate-Limit-Infrastruktur das ohne größere
  Architekturänderung trägt; andernfalls wird `[!]` mit Folge-Scope
  dokumentiert.
- [ ] Tests decken erlaubte globale Preflight-Origin, unbekannte
  Preflight-Origin, falsche Preflight-Methode, nicht erlaubten
  Preflight-Header, exakte Header-Sets für bekannte und unbekannte
  Origins, minimierte Signalisierung ohne Project-/Origin-
  Enumeration, project-spezifischen POST-Origin-Miss, leeren Origin im
  CLI-Pfad und Policy-Denial mit `403` ab.
- [ ] `0.11.0` Ingest-Control-Validate bleibt explizit kein
  produktiver Media-Server-Auth-Pfad; Doku und Tests verhindern diese
  Verwechslung.

## 7. Tranche 5 — Doku, Compatibility und Smokes

Ziel: Nutzer können von Public Project Token auf Session Tokens und
rotierbare Token-Generationen migrieren, ohne bestehende Lab-Flows zu
brechen.

DoD:

- [ ] `docs/user/demo-integration.md` und/oder neue Auth-Doku
  beschreiben Session-Token-Issuance, Browser-Nutzung, TTL, CSP,
  CORS und Token-Rotation.
- [ ] `docs/user/local-development.md` enthält sichere lokale
  Beispielwerte und erklärt, welche Tokens Demo-/Lab-Werte sind.
- [ ] `README.md` grenzt Auth-Härtung weiterhin gegen Production-
  Identity, OAuth/OIDC, KMS/Vault und SaaS-Tenant-Management ab.
- [ ] Datenschutz-/GDPR-Doku beschreibt, welche Auth-Metadaten
  zusätzlich entstehen, welche Claims pseudonym sind und dass
  IP-/User-Agent-Speicherung nicht ausgeweitet wird.
- [ ] SDK-Doku beschreibt kompatible Token-Nutzung: Project Token
  direkt bleibt erlaubt, Session Token wird bevorzugter Browser-Pfad.
- [ ] Doku und SDK-Beispiele beschreiben Session-Token-Caching:
  Token bis kurz vor `expires_at` wiederverwenden, Refresh-Fenster
  nutzen, keine Speicherung in `localStorage`/persistenten Browser-
  Stores empfehlen.
- [ ] Player-SDK braucht keine Secret-Verwaltung; falls API-Optionen
  ergänzt werden, bleiben ESM/CJS/IIFE und Public-API-Snapshot stabil
  oder bewusst additiv aktualisiert.
- [ ] Contract-Fixtures unter `spec/contract-fixtures/api/` pinnen
  mindestens Session-Token-Issuance, Playback mit Session Token,
  Expired Token, Policy Denied und Project-Token-Rotation.
- [ ] `make sync-contract-fixtures` kopiert neue API-Fixtures in
  `apps/api/adapters/driving/http/testdata/`.
- [ ] `make generated-drift-check` deckt neue Fixtures ab.
- [ ] Relevante Smokes sind dokumentiert; falls ein Runtime-Smoke
  eine laufende API braucht, bleibt er opt-in und wird nicht in
  `make gates` gezogen.

## 8. Tranche 6 — Release-Closeout

DoD:

- [ ] RAK-Verifikationsmatrix in §9 vollständig ausgefüllt.
- [ ] `make docs-check` grün.
- [ ] `make api-test` grün.
- [ ] `make build` grün.
- [ ] `make gates` grün.
- [ ] `make security-gates` grün oder CI-Job `Security gates` grün
  dokumentiert.
- [ ] `make generated-drift-check` grün.
- [ ] `pnpm --filter @npm9912/player-sdk run pack:smoke` grün, falls
  SDK-Public-API oder Paketmetadaten berührt wurden.
- [ ] Wave-2-Quality-Gates (`benchmark-smoke`, `fuzz-check`,
  `mutation-report`) bewertet: laufen lassen oder `[!]` mit
  Begründung dokumentieren.
- [ ] Vollständiger Versions-Bump auf `0.12.0` (Root
  `package.json`, Workspace-Packages, relevante Test-Fixtures).
- [ ] `CHANGELOG.md` mit `[0.12.0] - YYYY-MM-DD` aktualisiert.
- [ ] Roadmap auf released `0.12.0` und Folgephase `0.13.0`
  umgestellt.
- [ ] Plan nach `docs/planning/done/plan-0.12.0.md` verschoben und
  Status auf ✅ released aktualisiert.
- [ ] Annotierter Tag `v0.12.0` erstellt.

## 9. RAK-Verifikationsmatrix

Wird während der Umsetzung gepflegt. Jede Zeile braucht vor Closeout
Datei-, Test- und Doku-Nachweis.

| RAK | Priorität | Nachweis | Status |
| --- | --- | --- | --- |
| RAK-71 | Muss | Lastenheft-Patch, Scope-Grenze, README-/Doku-Abgrenzung gegen OAuth/OIDC, Admin-UI, KMS/Vault und SaaS-Tenant-Management. | [ ] |
| RAK-72 | Muss | Session-Token-Issuance, serverseitige Audience-Allowlist, harte TTL-Grenze ohne Clamp, Issuance-Abuse-Limits, Signaturvalidierung, Claims-Validierung, `token_id`/`jti`-Mapping, restart-stabiler Signing-Key-Ring, Fehlerpräzedenz, Fehlercodes und Contract-Tests. | [ ] |
| RAK-73 | Muss | Project-Token-Generationen, Rotation/Grace/Revocation, persistiertes `grace_until`, Persistenz ohne Klartext, Restart- und Repository-Tests. | [ ] |
| RAK-74 | Muss | Project Policies für Origins/Methoden/Header/Rate-Limits, separate Issuance-Quoten, globale konservative Preflight-Regeln mit exakt gepinnten `204`-Antworten und minimierter Signalisierung, project-spezifisches POST-Enforcement und Policy-Denial-Tests. | [ ] |
| RAK-75 | Muss | Kompatibilität mit bestehenden Project-Token-Flows inklusive fremder `Authorization`-Header, SDK/Demo/Analyze/Session/Ingest-Tests und Migrationsdoku. | [ ] |
| RAK-76 | Muss | Security-Doku, Threat Model, Datenschutz-/GDPR-Grenzen, CSP-Beispiele, Contract-Fixtures, Smokes und Drift-Check. | [ ] |

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
