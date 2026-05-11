# Auth / Token Lifecycle (`0.12.0`)

> **Stand**: `0.12.0` Tranche 5 (Lastenheft `1.1.15`, RAK-71..RAK-76).
> Siehe normativen Wire-Vertrag in
> [`spec/backend-api-contract.md`](../../spec/backend-api-contract.md)
> §3.9 und Plan-Scope in
> [`docs/planning/done/plan-0.12.0.md`](../planning/done/plan-0.12.0.md)
> §0.

## 0. Scope

`0.12.0` härtet den bestehenden lokalen/API-nahen Auth-Pfad: kurzlebige
signierte Session Tokens (`F-111`/RAK-72), rotierbare Project-Token-
Generationen (`F-112`/RAK-73), Project-gebundene Ingest Policies
(`F-113`/RAK-74). `X-MTrace-Token`-Legacy-Flows bleiben gültig
(RAK-75). Normativer Scope und Out-of-Scope-Liste in
[`spec/lastenheft.md`](../../spec/lastenheft.md) §13.14 (RAK-71) und
[`plan-0.12.0.md`](../planning/done/plan-0.12.0.md) §0.1.

---

## 1. Drei Auth-Pfade auf einen Blick

| Pfad | Header | Wann |
| --- | --- | --- |
| **Project Token** (Legacy) | `X-MTrace-Token: demo-token` oder `X-MTrace-Token: mtr_pt_…` | CLI/curl, Lab-Setup, alle Endpoints außer `/api/auth/session-tokens`-Issuance. Bleibt im `0.12.0`-Compat-Fenster gültig. |
| **Session Token (Bearer)** | `Authorization: Bearer mtr_st_…` | Bevorzugter Browser-Pfad ab `0.12.0`. Token wird vorher über `POST /api/auth/session-tokens` ausgestellt. |
| **Session Token (Header)** | `X-MTrace-Session-Token: mtr_st_…` | Alternative für Umgebungen, in denen `Authorization` schon belegt ist (z. B. Reverse-Proxy mit eigener Auth). |

Werden mehrere Tokens gleichzeitig präsentiert, müssen alle dasselbe
Project binden — sonst `401 auth_project_mismatch`. Ein zusätzlich
ungültiges Token blockiert den Fallback (`401 auth_token_invalid`,
kein silent downgrade auf einen gültigen niedriger priorisierten
Token). Die vollständige Header-Priorität und neunstufige
Fehlerpräzedenz steht in §3.9 des Wire-Vertrags.

---

## 2. Session-Token-Issuance

### 2.1 Wire-Vertrag

`POST /api/auth/session-tokens` mit Pflicht-`X-MTrace-Token` (Project
Token).

```bash
curl -X POST http://localhost:8080/api/auth/session-tokens \
  -H "Content-Type: application/json" \
  -H "X-MTrace-Token: demo-token" \
  -d '{
    "audience": "playback-events",
    "ttl_seconds": 900,
    "session_id": "sess_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "origin": "http://localhost:5173"
  }'
```

Response (`201 Created`):

```json
{
  "session_token": {
    "value": "mtr_st_eyJraWQiOiJsYWItMjAyNi0wNSJ9.eyJpc3MiOiJtLXRyYWNlIiwic3ViIjoiZGVtbyJ9.<sig>",
    "token_id": "st_01hzxj7a5k9v7w1e7btkj8v7n9",
    "project_id": "demo",
    "audience": "playback-events",
    "session_id": "sess_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "expires_at": "2026-05-09T10:15:00Z"
  }
}
```

`session_token.value` darf **nur in dieser Antwort** erscheinen. Logs,
Metriken, Traces, Persistenz und Fixtures enthalten höchstens
`token_id` oder Fingerprints. Die Spec-Tabelle der zehn
Auth-Fehler-Codes steht in §3.9.

### 2.2 Pflicht- und Optionalfelder

| Feld | Typ | Pflicht? | Bedeutung |
| --- | --- | --- | --- |
| `audience` | string | ✓ | Im `0.12.0`-Pflichtpfad nur `playback-events`. Andere Werte → `403 auth_session_scope_denied`. |
| `ttl_seconds` | int | optional | Harte globale Obergrenze 900. Project-Policies dürfen niedriger begrenzen, niemals höher. Fehlt → `min(project_max_ttl_seconds, 900)`. Werte ≤ 0 oder oberhalb der wirksamen Grenze → `422 auth_token_ttl_too_large` (kein silent clamp). |
| `session_id` | string | optional | Bindet den Token zusätzlich an eine Session. Wenn gesetzt, muss der konsumierende Request denselben Wert tragen — sonst `403 auth_session_scope_denied`. |
| `origin` | string | optional | Bindet den Token zusätzlich an einen Browser-Origin. Mismatch → `403 auth_session_scope_denied`. |
| `project_id` | string | optional | Konsistenzcheck zum Project Token. Mismatch → `401 auth_project_mismatch`. Fehlt → Project wird aus dem Token abgeleitet und in der Response zurückgegeben. |

### 2.3 Token konsumieren

```bash
# Bevorzugter Browser-Pfad
curl -X POST http://localhost:8080/api/playback-events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer mtr_st_…" \
  -d '{ "schema_version": "1.0", "events": [...] }'

# Alternative (z. B. wenn Authorization schon belegt ist)
curl -X POST http://localhost:8080/api/playback-events \
  -H "Content-Type: application/json" \
  -H "X-MTrace-Session-Token: mtr_st_…" \
  -d '{ "schema_version": "1.0", "events": [...] }'
```

---

## 3. SDK-Caching-Empfehlung

Der ausgestellte Session Token darf **bis kurz vor `expires_at`
wiederverwendet** werden — kein Pro-Event-Mint. Empfohlenes Muster:

```js
// Pseudocode — konkreter Caching-Adapter folgt im Player-SDK.
let cachedSessionToken = null;

async function getSessionToken() {
  const now = Date.now();
  // Refresh 30 s vor Ablauf, damit kein Inflight-Request mit
  // abgelaufenem Token reinrennt.
  if (cachedSessionToken && cachedSessionToken.expiresAtMs - now > 30_000) {
    return cachedSessionToken.value;
  }
  const resp = await fetch("/api/auth/session-tokens", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-MTrace-Token": projectToken,  // public, kommt aus dem Bundle
    },
    body: JSON.stringify({
      audience: "playback-events",
      ttl_seconds: 900,
      session_id: currentSessionId,
      origin: window.location.origin,
    }),
    credentials: "omit",  // niemals Cookies (NF-31/NF-32)
  });
  if (!resp.ok) {
    // Bei 422/429/403/401 nicht stillschweigend cachen — durch
    // den Error den Telemetrie-Pfad informieren und ggf.
    // exponentiell zurückziehen (analog Player-SDK-Retry-Budget).
    throw new Error(`session-token issuance failed: ${resp.status}`);
  }
  const body = await resp.json();
  cachedSessionToken = {
    value: body.session_token.value,
    expiresAtMs: Date.parse(body.session_token.expires_at),
  };
  return cachedSessionToken.value;
}
```

**Storage-Regeln:**

- **Niemals** in `localStorage`, `sessionStorage`, IndexedDB,
  Cookies oder anderen persistenten Browser-Stores ablegen — Session
  Tokens sind für die Dauer einer Page/Tab gedacht. In-Memory-Variable
  oder Service-Worker-Closure sind die akzeptierten Ablageorte.
- **Niemals** in URL-Query-Parametern, Referer-Headern oder
  `<img src>`-Beacons mitgeben — solche Pfade leaken in Logs und
  Browser-History.
- **Refresh-Fenster** wählen, das einen Inflight-Roundtrip abdecken
  kann — 30 s vor `expires_at` ist eine konservative Default-
  Grenze für 900-s-Tokens.
- **Issuance-Quote beachten**: jeder Mint zählt gegen das
  `auth_issuance_rate_limited`-Budget. Excessive-Mint deutet auf
  fehlerhaftes Caching hin und führt zu `429`-Fehlern.

---

## 4. CSP- und CORS-Empfehlungen

### 4.1 Content-Security-Policy

Wenn die einbettende Seite eine CSP setzt, muss sie den `connect-src`
für die m-trace-API freigeben:

```
Content-Security-Policy:
  default-src 'self';
  connect-src 'self' https://api.m-trace.example.com;
  script-src 'self' 'sha256-<player-sdk-bundle-hash>';
```

`'unsafe-inline'` und `'unsafe-eval'` sind **nicht** nötig — das
Player-SDK liefert keine Inline-Scripts und keine `eval`-Pfade.
`script-src 'self'` plus Bundle-Hash reicht.

### 4.2 CORS-Preflight

Die m-trace-API antwortet auf `OPTIONS /api/playback-events` und
`OPTIONS /api/auth/session-tokens` mit der globalen, konservativen
§3.9-Allowlist:

```
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: https://your.app
Access-Control-Allow-Methods: POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-MTrace-Token, X-MTrace-Session-Token, traceparent
Access-Control-Max-Age: 600
Vary: Origin, Access-Control-Request-Method, Access-Control-Request-Headers
Cache-Control: no-store
```

Unbekannte Origins erhalten **dieselbe `204`-Antwort, aber ohne
Allow-Origin/Methods/Headers** — keine Project- oder Origin-
Enumeration. Der Browser bricht dann den nachgelagerten POST
serverseitig nicht, sondern client-seitig (CORS-Check fehlschlägt).

`Access-Control-Allow-Origin: *` wird **nie** gesendet. `Vary: Origin`
ist immer gesetzt, damit shared Caches Origin-spezifische Antworten
nicht mischen.

---

## 5. Project-Token-Rotation

### 5.1 Generations-Modell

Project Tokens werden serverseitig als Generationen gehalten —
`token_id`, `key_hash`, `not_before`, `grace_until?`, `expires_at?`,
`revoked_at?`. Persistenz speichert nie den Klartext-Token. Klartext
erscheint nur:

- bei Erzeugung/Rotation in der Response,
- in markierten Operator-/Test-Fixtures (z. B. `demo-token`).

Lifecycle-Status wird **pro Validierungsaufruf** aus den Lifecycle-
Feldern plus aktueller Clock berechnet (`EvaluateProjectTokenStatus`),
damit kein Drift zwischen gespeichertem Status und Zeitvergleich
entsteht. Reihenfolge: `revoked > expired > not_yet_valid > grace >
active`.

### 5.2 Operator-Workflow

1. **SQLite-Persistenz** ist Voraussetzung — InMemory-Modus hat keine
   Generations-Persistenz. Default ab `0.4.0`:
   `MTRACE_PERSISTENCE=sqlite`.
2. **Signing-Key** per Env. Seit `0.12.5` (RAK-78) ist der
   Multi-Key-Pfad der Default:
   `MTRACE_AUTH_SIGNING_KEYS=<kid_a>:<base64url-secret>[,<kid_b>:<base64url-secret>,…]`
   plus `MTRACE_AUTH_SIGNING_ACTIVE_KID=<kid_a>`. Der Multi-Key-
   Pfad ist Voraussetzung für die in §5.3.1 dokumentierte
   Rotation. Als Backwards-Compat bleibt der Single-Key-Pfad aus
   `0.12.0` (`MTRACE_AUTH_SIGNING_KID` + `MTRACE_AUTH_SIGNING_KEY`)
   verfügbar; er ergibt einen degenerierten `len(keys)==1`-Resolver.
   Ohne jeden Key-ENV plus `MTRACE_AUTH_LAB_DEFAULT=1`-Opt-in failt
   der API-Start. Lab-Beispiel (Multi-Key):
   ```bash
   SECRET_A="$(openssl rand -base64 48 | tr '+/' '-_' | tr -d '=')"
   SECRET_B="$(openssl rand -base64 48 | tr '+/' '-_' | tr -d '=')"
   export MTRACE_AUTH_SIGNING_KEYS="lab-2026-03:${SECRET_A},lab-2026-05:${SECRET_B}"
   export MTRACE_AUTH_SIGNING_ACTIVE_KID="lab-2026-05"
   ```
3. **Frische `mtr_pt_*`-Generation anlegen** (heute über das
   `ProjectTokenRepository` direkt; ein dediziertes CRUD-API ist
   Folge-Scope). Solange noch keine Generation persistiert ist,
   akzeptiert der API-Pfad weiterhin `demo-token` über den Static-
   Resolver — Backward-Compat aus RAK-75.
4. **Rotation**: neue Generation anlegen, alte über
   `SetGraceUntil(now+grace)` markieren, später `Revoke(now)`.
   `grace_until` lebt persistent — ein API-Restart ändert die
   Grace-Entscheidung nicht.
5. **Rollback** ist nie reaktivierend: eine widerrufene oder
   abgelaufene Generation bleibt aus dem Audit-Pfad sichtbar; eine
   neue Generation muss frisch angelegt werden.

### 5.3 Signing-Key-Rotation

`kid` im Token-Header erlaubt parallele Signing-Keys. Alte Verify-
Keys bleiben über Deployments und Restarts geladen, bis alle damit
signierten Tokens abgelaufen sind. Der Code-Pfad wird seit `0.12.5`
(Tranche 1, RAK-78) vom `MultiKeySigningResolver` bedient — er
ersetzt den `0.12.0`-`StaticSigningKeyResolver` als Default-Pfad
und liest die Multi-Key-ENV-Konfiguration aus
`MTRACE_AUTH_SIGNING_KEYS=kid_a:base64,kid_b:base64` plus
`MTRACE_AUTH_SIGNING_ACTIVE_KID=kid_a`. Backwards-Compat: die
heutigen ENV-Variablen `MTRACE_AUTH_SIGNING_KEY` /
`MTRACE_AUTH_SIGNING_KID` bleiben als degenerierter Single-Key-
Pfad (`len(keys)==1`) unterstützt.

Validierung beim Boot (Plan-0.12.5 §5 DoD): leere KIDs, doppelte
KIDs, ungültige Base64-Werte und ein `MTRACE_AUTH_SIGNING_ACTIVE_KID`,
das nicht in der Liste vorkommt, brechen den Start mit klarer
Fehlermeldung — kein stiller Fallback. Reproduzierbarer Lab-Smoke:
`make smoke-key-rotation` (opt-in, nicht in `make gates`).

#### 5.3.1 Workflow Rotation (Code-Pfad in `0.12.5`, RAK-78)

Reihenfolge, ohne ablaufende Tokens zu invalidieren:

1. **Neuen Key generieren** — z. B.
   `openssl rand -base64 32`. Notieren als `<NEW_SECRET>` und
   neuen `kid` festlegen (z. B. `kid_2026_05`).
2. **Verify-Set erweitern** — den neuen Key zusätzlich zum alten
   in `MTRACE_AUTH_SIGNING_KEYS` aufnehmen, **ohne** den aktiven
   `kid` umzuschalten:
   ```
   MTRACE_AUTH_SIGNING_KEYS=kid_2026_03:<OLD_SECRET>,kid_2026_05:<NEW_SECRET>
   MTRACE_AUTH_SIGNING_ACTIVE_KID=kid_2026_03
   ```
   Restart der API-Instanzen. Ergebnis: alter Key signiert weiter,
   alter und neuer Key verifizieren.
3. **Aktiven `kid` umschalten** — Active-KID auf den neuen Wert
   setzen:
   ```
   MTRACE_AUTH_SIGNING_KEYS=kid_2026_03:<OLD_SECRET>,kid_2026_05:<NEW_SECRET>
   MTRACE_AUTH_SIGNING_ACTIVE_KID=kid_2026_05
   ```
   Restart. Ergebnis: neue Tokens werden mit `kid_2026_05`
   signiert; bereits ausgegebene Tokens unter `kid_2026_03`
   verifizieren weiterhin gegen den alten Key, bis sie ablaufen.
4. **Alten Key abbauen** — nach **max-Token-TTL plus Reservezeit**
   den alten Key aus `MTRACE_AUTH_SIGNING_KEYS` entfernen. Bei
   Default-TTL aus `MTRACE_AUTH_SESSION_TOKEN_TTL` (typisch
   ≤ 5 min) plus konservativer Reserve (z. B. 30 min) reicht
   nach Schritt 3 ein Wartefenster von etwa einer Stunde, bevor
   der alte Key entfernt wird:
   ```
   MTRACE_AUTH_SIGNING_KEYS=kid_2026_05:<NEW_SECRET>
   MTRACE_AUTH_SIGNING_ACTIVE_KID=kid_2026_05
   ```
   Restart. Ergebnis: ältere Tokens, die noch unter `kid_2026_03`
   ausgegeben wurden, werden **nicht** mehr verifizieren — daher
   das Wartefenster.

**Anti-Pattern**: `MTRACE_AUTH_SIGNING_KEY`/`_KID` einfach
überschreiben (Schritt 3 ohne Schritt 2 vorab), ohne den alten
Key im Verify-Set zu behalten. Resultat: alle ausgegebenen Tokens
werden sofort `auth_signature_invalid`. **Genau das verhindert
das Multi-Key-Schema.**

**Restart-Stabilität** ist im Code unabhängig vom ENV-Schema seit
`0.12.0` getestet (`TestHMACSigner_RestartStableAcrossKeyResolverReinitialization`)
— ein Restart ohne Key-Wechsel invalidiert keine Tokens. Seit
`0.12.5` deckt zusätzlich `TestParseSigningKeysEnv_RotationEndToEnd`
den vollständigen ENV-Parser-→-Resolver-→-Signer-→-Rotation-Pfad ab;
derselbe Test wird vom Lab-Smoke `make smoke-key-rotation` als
reproduzierbarer Operator-Check angeboten.

### 5.4 Issuance-Limiter-Backend

`POST /api/auth/session-tokens` (und alle weiteren Issuance-Pfade,
die `driven.IssuanceRateLimiter.Allow` aufrufen) lassen sich seit
`0.12.5` (Tranche 2, RAK-77) gegen einen geteilten SQLite-State
betreiben. Damit löst der Limiter R-17 für **Single-Host-Multi-
Replica-Setups** auf: zwei API-Instances auf demselben Host, die
sich denselben SQLite-Volume teilen (Compose-`volumes:`,
K8s-`hostPath`), zählen das Token-Bucket gemeinsam — die effektive
Issuance-Quote bleibt damit auf den konfigurierten Wert begrenzt,
unabhängig von der Replica-Zahl.

Auswahl per ENV:

| Wert                            | Verhalten                                                                             |
| ------------------------------- | ------------------------------------------------------------------------------------- |
| `MTRACE_AUTH_ISSUANCE_LIMITER` leer / `memory` | **Default** — In-Process-Token-Bucket (`0.12.0`-Pfad). Misst pro Replica. |
| `MTRACE_AUTH_ISSUANCE_LIMITER=sqlite`          | Opt-in — Shared-State über die `auth_issuance_counters`-Tabelle (Migration V5). Braucht aktive `MTRACE_PERSISTENCE=sqlite`; ohne SQLite-Persistenz hard-failt der API-Start. Single-Host-only. |
| `MTRACE_AUTH_ISSUANCE_LIMITER=redis` (`0.12.6` T7) | Multi-Host-Backend über Redis. Atomare Lua-`EVAL`-Token-Bucket-Operation; teilt sich den Redis-Server mit dem Origin-Limiter (eigener Key-Prefix `mtrace:issuance`). Pflicht-ENV `MTRACE_REDIS_ADDR`; optional `MTRACE_REDIS_AUTH` / `MTRACE_REDIS_DB`. Fail-Mode default fail-closed; opt-in fail-open via `MTRACE_AUTH_ISSUANCE_FAIL_OPEN=1`. |
| `MTRACE_AUTH_ISSUANCE_LIMITER=memcached`       | Bewusst nicht unterstützt — bleibt gemeinsames Folge-Item mit `R-22` (siehe §5.9), falls Operator-Bedarf nach Memcached entsteht. |

**Topologie-Constraint**: der SQLite-Pfad wirkt nur über einen
gemeinsam gemounteten Persistent-Volume. Auf zwei separaten Hosts
mit jeweils eigener SQLite-Datei sieht jeder Host nur sein lokales
Bucket — semantisch identisch zum `memory`-Default. Für echte
Multi-Host-Topologie (Kubernetes über mehrere Nodes, Container-
Plattformen ohne shared storage) liefert ab `0.12.6` Tranche 7
der `redis`-Backend einen Network-State-Bucket.

**Redis-Pfad (`0.12.6` T7 / R-17)**:

- Atomicity: ein einziger `EVAL`-Lua-Script-Aufruf prüft beide
  Buckets (global + project) inkl. Refund bei project-deny. Damit
  ist die Two-Bucket-Logik race-frei auch unter Parallel-Calls aus
  mehreren API-Replicas.
- Bucket-Keys: `mtrace:issuance:global` und
  `mtrace:issuance:project:<projectID>`. TTL Default 24 h
  (halten idle Buckets aus dem Redis-Speicher).
- **Fail-Mode**:
  - Default **fail-closed**: Redis-Outage → Limiter liefert
    `deny`; HTTP-Handler antwortet mit `429 auth_issuance_rate_limited`.
    Damit wird ein Redis-Outage nie zu einer Mint-Welle.
  - Opt-in **fail-open** via `MTRACE_AUTH_ISSUANCE_FAIL_OPEN=1`:
    bei Outage fällt der Limiter auf den lokalen
    In-Memory-Fallback-Bucket zurück. Misst pro Replica —
    explizite Operator-Entscheidung gegen den Multi-Host-Bucket-
    Konsens während des Outage-Fensters.
  - Der `FailOpen`-Schalter gilt **gemeinsam** für Issuance- und
    Origin-Limiter (`§5.9`), damit kein halb-fail-closed-Pfad
    entsteht.

**RAK-74-Scope-Cut** (Lastenheft §13.14) bleibt aktiv: der Limiter
hängt **nicht** vor `/api/ingest/*` — der Ingest-Control-Pfad ist
operator-/CLI-getrieben und nutzt das `0.11.0`-Token-Modell. Limiter-
Wirkung gilt ausschließlich für die Browser-/Telemetrie-Pfade
(`POST /api/auth/session-tokens`, `POST /api/playback-events`).

Atomarität ist über `BEGIN IMMEDIATE` der SQLite-DSN garantiert
(siehe `apps/api/internal/storage` und ADR-0002 §8.3): konkurrente
Allow-Calls über alle Replicas werden serialisiert. Bucket-Einträge
haben eine TTL (Default 24h); opportunistisches Cleanup räumt
veraltete Project-Buckets während des Hot-Paths auf, ohne einen
dedizierten Hintergrundjob.

Reproduzierbarer Lab-Smoke: `make smoke-issuance-replica` (opt-in,
nicht in `make gates`). Der Smoke nutzt zwei `*sql.DB`-Verbindungen
auf dieselbe SQLite-Datei, verbraucht das Project-Bucket auf
Instance A und prüft, dass Instance B den nächsten Allow als
„denied" sieht.

**Multi-Host-Smoke** (`0.12.6` T7): `make smoke-issuance-multi-host`
(opt-in). Startet einen `miniredis`-Mock und zwei
`RedisIssuanceRateLimiter`-Adapter-Instances (analog zwei API-
Replicas auf verschiedenen Hosts) gegen denselben Mock-Server;
verifiziert Cross-Instance-Sharing, den Refund-Pfad bei project-
deny und beide Fail-Modi (fail-closed Default und fail-open
opt-in).

### 5.5 Signing-Key-Backend (Secret-Source)

Wo das Signing-Key-Material **herkommt**, ist seit `0.12.5`
(Tranche 3, RAK-79) ein eigener Driven-Port
(`hexagon/port/driven/auth_secret_backend.go`). Der Boot-Pfad
wählt das Backend per ENV; alles andere ist Adapter-Detail.

| `MTRACE_AUTH_SECRET_BACKEND` | Adapter                                                                 | Konfigurations-Quelle                                                                |
| ---------------------------- | ----------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| leer / `env` (Default)       | `EnvSecretBackend` — Backwards-Compat zu `0.12.0`/`0.12.5` Tranche 1     | `MTRACE_AUTH_SIGNING_KEYS` + `MTRACE_AUTH_SIGNING_ACTIVE_KID` (oder Single-Key)      |
| `vault`                      | `VaultSecretBackend` — Skelett gegen Vault KV-v2                        | `MTRACE_AUTH_VAULT_ADDR/_TOKEN/_PATH`                                                |
| jeder andere Wert (`kms`, …) | Boot-Validator failt mit „not supported"                                | —                                                                                    |

**Vault-Adapter-Skelett (Code-Pfad in `0.12.5`, RAK-79):**

- Eigener minimaler HTTP-Client gegen `/v1/<mount>/data/<path>` —
  ohne `hashicorp/vault/api`-Dependency. Eine produktive Anbindung
  kann ihn 1:1 durch einen `hashicorp/vault/api`-Adapter ersetzen,
  ohne den `AuthSecretBackend`-Port zu ändern.
- Authentication: Token only (`X-Vault-Token`). AppRole, AWS-IAM-
  Auth und Kubernetes-Service-Account-Auth bleiben Folge-Item für
  die produktive Anbindung.
- Pflicht-ENV:
  - `MTRACE_AUTH_VAULT_ADDR` (z. B. `http://127.0.0.1:8200`)
  - `MTRACE_AUTH_VAULT_TOKEN`
  - `MTRACE_AUTH_VAULT_PATH` — KV-v2-Pfad inkl. `data/`-Marker
    (z. B. `secret/data/m-trace/signing`).
- Optionale ENV-Var-Aliase: `MTRACE_AUTH_VAULT_KEYS_FIELD`
  (Default `keys`), `MTRACE_AUTH_VAULT_ACTIVE_KID_FIELD`
  (Default `active_kid`).
- Secret-Format im Vault-Pfad: dieselben zwei Felder, die der
  ENV-Backend aus `MTRACE_AUTH_SIGNING_KEYS` / `_ACTIVE_KID` liest
  — `kid_a:<base64>,kid_b:<base64>` für `keys`, plain string für
  `active_kid`. Beide Backends teilen sich denselben Parser
  (`ParseSigningKeysEnv`).

**Lifecycle (`0.12.5`):**

- **Boot-Time-Load**: das Backend wird beim API-Start einmal
  aufgerufen; danach hält der `MultiKeySigningResolver` das
  Material in-memory.
- **Kein periodischer Refresh**, kein TTL-Caching im Adapter —
  Schlüsselwechsel passieren per Operator-Restart (gleich wie
  beim `env`-Pfad, siehe §5.3.1).
- **Fail-closed**: ein nicht erreichbares Backend, ein
  HTTP-Fehler-Status, ein leeres `keys`-Feld oder ungültiges
  Material liefert beim Boot einen klaren Fehler — die API
  startet nicht. Es gibt **kein** stillen Fallback auf einen
  Lab-Default für externe Backends; nur der ENV-Pfad hat das
  `MTRACE_AUTH_LAB_DEFAULT=1`-Opt-in.

**Lab-Setup (Vault dev-Server):**

```bash
# Vault im Dev-Mode starten (Root-Token aus stdout merken).
vault server -dev -dev-root-token-id=root-dev &
export VAULT_ADDR=http://127.0.0.1:8200

# Secret-Felder schreiben.
vault kv put -mount=secret m-trace/signing \
  keys="kid_a:$(openssl rand -base64 32 | tr '+/' '-_' | tr -d '=')" \
  active_kid="kid_a"

# API starten:
export MTRACE_AUTH_SECRET_BACKEND=vault
export MTRACE_AUTH_VAULT_ADDR=$VAULT_ADDR
export MTRACE_AUTH_VAULT_TOKEN=root-dev
export MTRACE_AUTH_VAULT_PATH=secret/data/m-trace/signing
```

**Resttrigger** für eine vollständig aufgelöste R-20-Auflösung
bleiben offen: erste Operator-Anbindung an produktives Vault
oder KMS, Compliance-Audit (PCI/SOC2). Die Skelett-Lieferung in
`0.12.5` deckt nur den Driven-Port und einen Lab-Pfad — siehe
R-20 im Risiken-Backlog.

### 5.6 Browser-Ingest-Policy

Bis `0.12.0` war `/api/ingest/*` strikt operator-/CLI-only:
der RAK-74-Scope-Cut hielt jeden Browser-Konsumenten heraus.
`0.12.5` Tranche 4 (RAK-80, R-21) hebt diesen Scope-Cut
**kontrolliert** auf — pro Project per
`domain.BrowserIngestPolicy`.

**Schema-Erweiterung** im `domain.ProjectPolicy`:

| Feld            | Typ        | Bedeutung                                                                                       |
| --------------- | ---------- | ----------------------------------------------------------------------------------------------- |
| `Enabled`       | `bool`     | Master-Switch. `false` (Default) → RAK-74-Scope-Cut bleibt strikt. `true` → Browser-Pfad offen. |
| `CORSAllowlist` | `[]string` | Browser-Origins, die durchgelassen werden (genauer String-Match).                               |
| `CSRFRequired`  | `bool`     | Wenn `true`, müssen POSTs einen nicht-leeren `X-MTrace-CSRF`-Header tragen.                     |
| `OriginPin`     | `string`   | Defense-in-Depth: wenn gesetzt, muss `Origin` exakt diesem Wert entsprechen (nicht nur in Allowlist). |

**Preflight-Verhalten** für `OPTIONS /api/ingest/*`:

- **Keine aktivierte Policy für irgendein Project**: Preflight läuft
  über die globale konservative Allowlist (`dashboardPreflightHandler`)
  — RAK-74-Scope-Cut bleibt strikt.
- **Aktivierte Policy + Origin in CORSAllowlist eines Projects**:
  `204` + `Access-Control-Allow-Origin: <origin>` plus
  `Access-Control-Allow-Methods: POST, OPTIONS` und
  `Access-Control-Allow-Headers: Content-Type, X-MTrace-Token, X-MTrace-CSRF`.
- **Aktivierte Policy aber Origin nicht in Allowlist**: `204` ohne
  Allow-Origin (kein Enumerations-Leak); `mtrace_cors_preflight_refused_total`
  wird inkrementiert.

**POST-Enforcement** auf `/api/ingest/*` (Middleware, hängt sich vor
die bestehenden Handler):

1. **Kein `X-MTrace-Token`-Header**: Middleware tut nichts; der
   Handler liefert sein heutiges `auth_token_*`-Verhalten.
2. **Token resolved + Policy.Enabled=false**: Pfad wie heute
   (Backwards-Compat).
3. **Token resolved + Enabled=true**:
   - Origin **muss** in `CORSAllowlist` stehen — sonst
     `403 ingest_browser_origin_not_allowed`.
   - Wenn `OriginPin != ""`, muss Origin exakt dem Pin entsprechen
     — sonst `403 ingest_browser_origin_pin_mismatch`. Ein
     fehlender `Origin`-Header bei gesetztem Pin gilt ebenfalls als
     Mismatch.
   - Wenn `CSRFRequired=true`, muss `X-MTrace-CSRF` nicht-leer sein
     — sonst `403 ingest_browser_csrf_missing`. **Hinweis**: das
     Skelett prüft nur Header-Anwesenheit; eine produktive Anti-
     CSRF-Token-Bibliothek mit signierten/zeitlich begrenzten
     Tokens ist Folge-Item (siehe R-21-Mitigation im Backlog).

**Beispiel-Konfiguration (Operator-Setup, in-memory Resolver):**

```go
policies := map[string]domain.ProjectPolicy{
    "tenant-a": {
        ProjectID:      "tenant-a",
        AllowedOrigins: []string{"https://app.tenant-a.example.com"},
        BrowserIngest: domain.BrowserIngestPolicy{
            Enabled:       true,
            CORSAllowlist: []string{"https://app.tenant-a.example.com"},
            CSRFRequired:  true,
            OriginPin:     "https://app.tenant-a.example.com",
        },
    },
}
```

Reproduzierbarer Lab-Smoke: `make smoke-browser-ingest` (opt-in,
nicht in `make gates`). Der Smoke fährt alle sechs Preflight-/POST-
Pfade durch — inklusive Origin-Pin-Mismatch und CSRF-Missing-Fall.

### 5.7 MediaMTX-Auth-Bridge

Mit `0.12.5` Tranche 5 (RAK-81, R-14) gibt es einen optionalen
HTTP-Endpoint, der MediaMTX' `externalAuth`-Hook bedient:
`POST /api/ingest/auth-hook`. Skelett-Adapter — bewusst ohne
eigene Token-Issuance, weil die existierenden Stream-Keys aus
`/api/ingest/streams` schon den nötigen kryptographischen Anker
liefern.

**MediaMTX-Config** (Auszug `mediamtx.yml`):

```yaml
authMethod: http
authHTTPAddress: http://m-trace-api:8080/api/ingest/auth-hook
authHTTPExclude:
  - action: api
  - action: metrics
```

**Wire-Vertrag**:

- Methode: `POST`, `Content-Type: application/x-www-form-urlencoded`.
- Form-Body (MediaMTX-Standard): `user`, `password`, `action`,
  `path`, optional `ip`, `protocol`, `id`, `query`.
- Mapping (m-trace-Konvention):
  - `user`     → Project-ID
  - `password` → Stream-Key Klartext (einmaliger Wert aus
    `POST /api/ingest/streams` bzw. `…/rotate-key`)
  - `path`     → Stream-ID
  - `action`   → nur `publish` ist erlaubt; `read`/`api`/`metrics`
    liefern `403`.
- Response: `200` = allow, `403` = deny, `400` = ungültiger
  Content-Type, `405` = falsche HTTP-Methode. Bodies sind leer
  bzw. enthalten nur den Plain-Text-Fehlergrund.

**Sicherheitsprofil**:

- Trust-Boundary zwischen MediaMTX und m-trace API. Der Endpoint
  hat **selbst keine Project-Token-Auth** — Netzwerk-Isolation
  ist Operator-Verantwortung (Compose-internal-Netz, K8s-
  `ClusterIP`, Reverse-Proxy mit IP-Allowlist).
- Audit-Log markiert jeden Deny-Pfad mit Project-/Stream-ID und
  Reason-Code (`action_not_supported`/`missing_field`/`invalid_key`/
  `validate_error`); Klartext-Material wird **niemals** geloggt.
- Idempotent — die zugrundeliegende `ValidateKey`-Operation ist
  side-effect-frei, ein replay-er MediaMTX-Hook führt zu keinem
  doppelten Effekt.

**Was nicht im Skelett enthalten ist** (Folge-Item nach `0.12.5`):

- Read-Auth (Player-Authentifizierung über MediaMTX-`read`-Action).
- Eigener Publish-Token-Lifecycle (`POST /api/ingest/publish-tokens`),
  separat zu den Stream-Keys.
- Mutual-TLS zwischen MediaMTX und m-trace (heute via
  Compose-Netz-Isolation, in Production via Reverse-Proxy mit
  mTLS-Terminierung lösbar).

Reproduzierbarer Lab-Smoke: `make smoke-mediamtx-auth` (opt-in).
Wickelt alle sieben Wire-Pfade (allow, deny-invalid-key,
deny-read-action, deny-missing-field, bad-content-type, GET 405,
validate-error) als End-to-End-Test über `httptest.Server` ab —
keine MediaMTX-Container nötig. Echte Compose-Variante mit einem
MediaMTX-Container, der gegen die laufende m-trace-API
authentifiziert, bleibt Folge-Item.

### 5.8 Outbound-Webhook für Stream-Lifecycle

Mit `0.12.5` Tranche 5 (RAK-82, R-16) liefert m-trace eine
optionale Webhook-Zustellung der `0.11.0`-Lifecycle-Events
(`stream_started`/`stream_ended`) an externe Konsumenten — der
heutige lokale Pfad
(`POST /api/ingest/hooks/stream-{started,ended}`) bleibt
unverändert, plus Outbound-POST.

**Konfiguration** (Boot-ENV):

| ENV                                 | Pflicht | Bedeutung                                                                  |
| ----------------------------------- | ------- | -------------------------------------------------------------------------- |
| `MTRACE_OUTBOUND_WEBHOOK_URL`       | nein    | Zustelladresse. Leer/unset → Adapter deaktiviert (No-Op, kein Outbound).   |
| `MTRACE_OUTBOUND_WEBHOOK_SECRET`    | nein    | HMAC-Secret für die Payload-Signatur. Empfohlen ≥ 32 Byte Entropie.        |

**Wire-Vertrag** (siehe `apps/api/adapters/driven/webhooks/http_dispatcher.go`):

- Methode: `POST`, `Content-Type: application/json`.
- Body (JSON):
  ```json
  {
    "event_id":     "evt_…",
    "type":         "stream_started" | "stream_ended",
    "project_id":   "<id>",
    "stream_id":    "<id>",
    "observed_at":  "<RFC3339Nano>",
    "source":       "local-smoke" | "mediamtx-hook",
    "connection_id":"…",
    "reason":       "…"
  }
  ```
- Header `X-MTrace-Signature: sha256=<hex>` — HMAC-SHA-256 über
  den exakten Body, signiert mit `MTRACE_OUTBOUND_WEBHOOK_SECRET`.
  Der Konsument muss die Signatur prüfen.
- Header `X-MTrace-Timestamp` (RFC3339Nano) für Replay-Schutz.
- Erfolg: HTTP-Status `200 ≤ s < 300`; alles andere zählt als
  fehlgeschlagen.

**Retry-Schema**:

- Bis zu **3 Versuche** (`MaxAttempts`).
- Exponential-Backoff: `100ms`, `200ms`, `400ms` (`BaseBackoff`,
  Multiplikator 2).
- Per-Versuch-Timeout: `10s` (`RequestTimeout`).
- Nach Erschöpfung: Dead-Letter — der Adapter loggt
  `outbound webhook dead-letter` mit `event_id`/`endpoint`/
  `last_error`. **Wichtig**: der Lifecycle-Pfad failed nicht;
  `POST /api/ingest/hooks/stream-…` antwortet weiterhin `202`,
  selbst wenn der Outbound-Konsument nicht erreichbar ist.

**Sicherheitsprofil**:

- Klartext-Stream-Keys werden **niemals** in der Payload
  mitgeschickt — nur das `key_fingerprint`-Äquivalent ist im
  Datenbank-Event sichtbar, taucht aber im Wire-Body bewusst
  nicht auf (Lifecycle-Events haben kein Key-Material).
- HMAC-Signatur über den exakten Body schützt vor Manipulation
  im Transit; der Konsument lehnt Bodies ohne korrekte Signatur
  ab.
- Replay-Schutz: `event_id` ist opak und eindeutig — der
  Konsument kann sie deduplizieren.

**Was nicht im Skelett enthalten ist** (Folge-Item):

- Project-spezifische Webhook-URLs (heute global eine URL pro
  m-trace-Instanz). Multi-Tenant-Setups würden eine
  Project-Policy-Erweiterung brauchen.
- Persistente Dead-Letter-Queue. Heute landet ein erschöpftes
  Event ausschließlich im Log; eine Wiederzustellung nach
  Operator-Trigger wäre Folge-Item.
- Konfigurierbare `MaxAttempts`/`BaseBackoff`/`RequestTimeout`
  via ENV (heute hartcodiert, weil das Skelett eine konservative
  Defaults-Topologie nutzt).

Reproduzierbarer Lab-Smoke: `make smoke-outbound-webhook` (opt-in).
Wickelt sieben Pfade (disabled-noop, happy path, HMAC-Match,
Retry-success, Dead-Letter-Exhaustion, Body-Shape, Context-Cancel)
gegen einen `httptest.Server`-Mock-Konsumenten.

### 5.9 Origin-/IP-Rate-Limiter (`0.12.6` / R-22)

`0.12.6` Tranche 6 ergänzt einen **Defense-in-Depth-Limiter** vor
`POST /api/auth/session-tokens` und `POST /api/playback-events`, der
Bursts aus einer einzelnen Client-Quelle bremst **bevor** der
project-skopierte Issuance- oder Event-Counter überhaupt geprüft wird.
Anders als der Issuance-Limiter (`§5.4`, project-skopiert) misst dieser
Limiter pro **Client-IP** (oder Origin-Hash).

**ENV-Selektor** `MTRACE_ORIGIN_RATE_LIMITER`:

| Wert | Verhalten |
|---|---|
| `disabled` (Default) / leer | Kein Limiter; Pfad ist 1:1 wie vor `0.12.6`. |
| `memory` | In-Process-Token-Bucket pro Key (`r.RemoteAddr` oder XFF-letzter-Hop). Capacity 20, Refill 5 / s; misst pro Replica. |
| `redis` (`0.12.6` T7) | Multi-Host-Backend über Redis. Atomare Lua-`EVAL`-Token-Bucket-Operation; gleicher Redis-Server wie der Issuance-Limiter, eigener Key-Prefix `mtrace:origin`. Pflicht-ENV `MTRACE_REDIS_ADDR`; Fail-Mode-Schalter `MTRACE_AUTH_ISSUANCE_FAIL_OPEN` (gilt gemeinsam für beide Limiter — siehe §5.4). TTL 10 Minuten (kürzlebiger als Issuance-Buckets, weil pro Key/IP). |
| `sqlite` | **nicht unterstützt** — Origin-Limits über Hosts hinweg brauchen ein Network-Backend; SQLite-Volume produziert false-negative-Limits. |
| `memcached` | Folge-Item gemeinsam mit dem Issuance-Limiter, falls Operator-Bedarf nach Memcached entsteht. |

**`X-Forwarded-For`-Trust** (`MTRACE_TRUST_FORWARDED_FOR`): Setzt der
Operator `1`/`true`/`yes`, nutzt der Limiter das **letzte (rechteste)**
Element der XFF-Header-Liste als Client-IP. Das ist nur dann korrekt,
wenn der Reverse-Proxy:

1. Den XFF-Header beim Eintritt **strippt** (sonst kann ein Client
   beliebige XFF-Werte vorgeben und die Limit-Buckets stündlich
   switchen — Spoofing-Surface).
2. Genau **einen** Hop hinzufügt (sonst zeigt das rechte XFF-Element
   auf einen internen Hop statt auf den Client).

Ohne den Opt-in fällt der Limiter auf `r.RemoteAddr` zurück, was bei
direktem Client→Server-Setup korrekt ist, aber bei Reverse-Proxy auf
den Proxy zeigt (alle Clients teilen sich dann einen Bucket — der
Limiter wirkt wie ein Single-Source-Throttle und ist effektiv nutzlos).

**Reject-Body** bei Limit-Verletzung:

```json
{"error":"origin_rate_limited"}
```

HTTP-Status: `429 Too Many Requests`. Reihenfolge gegenüber dem
Issuance-Limiter: **Origin-Limit zuerst**; wenn dieser durchgelassen
hat, prüft der Issuance-Limiter (`§5.4`) seinen project-skopierten
Bucket. Operator-Pfade, die nur den Issuance-Pfad messen wollten,
müssen den Origin-Limiter deaktiviert lassen (Default).

**Reproduzierbarer Lab-Smoke**: `make smoke-origin-rate-limit`
(opt-in). Drei aufeinanderfolgende Token-Aufrufe → erwartete
201/201/429 mit `origin_rate_limited`-Body.

---

## 6. Datenschutz / GDPR

`0.12.0` erweitert die persistierten Auth-Metadaten **nicht** um
PII-Felder. Konkret:

- **IP-/User-Agent-Speicherung wird nicht ausgeweitet** — die in
  `0.4.0` etablierte Cardinality-Politik (`spec/telemetry-model.md`
  §3.1) gilt unverändert. Der HTTP-Adapter liest `client_ip` für die
  Rate-Limit-Dimension F-110, persistiert ihn aber nicht.
- **Session-Token-Claims sind pseudonym**: `iss`, `sub` (`project_id`),
  `aud`, `iat`/`nbf`/`exp` (Unix-Sekunden), `jti` (`token_id`), plus
  optional `session_id` und `origin`. Keine User-IDs, keine E-Mail-
  Adressen, keine Browser-Fingerprints.
- **Project-Token-Persistenz** speichert nur `key_hash`,
  `fingerprint` und Lifecycle-Felder — niemals Klartext.
- **Logs/Metriken/Traces** zeigen ausschließlich `token_id` oder
  Fingerprints. Klartext-Tokens leaken in keiner Code-Pfad.
- **Issuance-Audit**: `0.12.0` führt keinen dedizierten
  Per-Token-Issuance-Log oder -Metric. Rate-Limit-Befunde landen als
  `429 auth_issuance_rate_limited`-Wire-Antwort an den Aufrufer;
  CORS-Preflight-Refusals sind über den Counter
  `mtrace_cors_preflight_refused_total{path}` sichtbar.

Folge-Doku zur GDPR-Klassifikation (Welche Auth-Daten sind unter
welchem Rechtsgrund verarbeitet?) bleibt — falls erforderlich —
Aufgabe des Operators und liegt außerhalb dieses Repos.

---

## 7. Verweise

- [`spec/backend-api-contract.md`](../../spec/backend-api-contract.md)
  §3.9 — normativer Wire-Vertrag (Auth-Matrix, Header-Priorität,
  9-stufige Fehlerpräzedenz, CORS-Preflight, Project-Policies).
- [`spec/lastenheft.md`](../../spec/lastenheft.md) §13.14 — RAK-71..
  RAK-76.
- [`docs/planning/done/plan-0.12.0.md`](../planning/done/plan-0.12.0.md)
  §0 — Plan-Scope, Architektur, Threat Model.
- [`docs/user/local-development.md`](./local-development.md) §2.7.3 —
  Operator-Migrationspfad `demo-token` → rotierbare Generation.
- [`docs/planning/in-progress/risks-backlog.md`](../planning/in-progress/risks-backlog.md)
  R-14, R-17, R-18 — Auth-bezogene Folge-Risiken mit
  Triggerschwellen.
