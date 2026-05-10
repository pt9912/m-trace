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
2. **Signing-Key** per Env: `MTRACE_AUTH_SIGNING_KID=<kid>` und
   `MTRACE_AUTH_SIGNING_KEY=<base64url-encoded ≥32 bytes>`.
   Ohne explizit gesetzten Key plus `MTRACE_AUTH_LAB_DEFAULT=1`-
   Opt-in failt der API-Start. Lab-Beispiel:
   ```bash
   export MTRACE_AUTH_SIGNING_KID="lab-2026-05"
   export MTRACE_AUTH_SIGNING_KEY="$(openssl rand -base64 48 | tr '+/' '-_' | tr -d '=')"
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
signierten Tokens abgelaufen sind. Im aktuellen Spike-Setup hält der
`StaticSigningKeyResolver` einen einzelnen aktiven Key. Multi-Key-
Konfiguration über Env (z. B.
`MTRACE_AUTH_SIGNING_KEYS=kid_a:base64,kid_b:base64`) und Operator-
Runbook für die Rotation sind als R-18 im Risiken-Backlog geführt
(Triggerschwelle: erstes Rotation-Event in Lab/Staging).

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
