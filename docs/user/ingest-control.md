# Ingest-Control (lokaler Stream-Control-Pfad)

`0.11.0` ergänzt das m-trace-Lab um einen **lokalen Stream-Control-
Pfad**: Streams werden über `apps/api` angelegt, jeder Stream
bekommt einen kryptografisch sicheren Stream-Key, ein
deterministisches MediaMTX-Konfigurations-Artefakt fällt für den
Lab-Stack ab, und Lifecycle-Hooks lassen sich lokal einspeisen.

> Bezug: [`spec/lastenheft.md`](../../spec/lastenheft.md) §1.1.14,
> RAK-65..RAK-70 / NF-13;
> [`docs/planning/in-progress/plan-0.11.0.md`](../planning/done/plan-0.11.0.md);
> [`spec/backend-api-contract.md`](../../spec/backend-api-contract.md) §3.8;
> [`examples/ingest-control/README.md`](../../examples/ingest-control/README.md).

## 1. Lieferumfang `0.11.0`

- **HTTP-API** `/api/ingest/streams` (Create/List), `/api/ingest/streams/{id}`,
  `/api/ingest/streams/{id}/rotate-key`, `/api/ingest/streams/{id}/validate-key`,
  `/api/ingest/media-server-config`, `/api/ingest/hooks/stream-{started,ended}`.
- **Stream-Keys** mit 256-Bit-Entropie (CSPRNG `crypto/rand`),
  URL-safer Base64-Repräsentation mit Prefix `mtr_ing_`. Persistiert
  werden ausschließlich SHA-256-Hash plus `key_fingerprint`
  (Anzeige-Wert). Klartext gibt es **genau einmal** im
  Create-/Rotate-Response.
- **Persistenz** in SQLite (`apps/api/internal/storage/migrations/V2__ingest.sql`
  + `V3__ingest_lifecycle_extras.sql`). InMemory-Adapter ist Default
  für Tests; SQLite läuft im Compose-Stack.
- **MediaMTX-Artefakt-Generator** liefert deterministisches YAML mit
  Pfad-Block, sanitized `display_name` und `key_fingerprint` —
  Klartext-Keys erscheinen nie.
- **Lifecycle-Hooks** für `stream_started` / `stream_ended`. Source-
  Allowlist `local-smoke`/`mediamtx-hook`. Server-generierte
  `event_id` (Prefix `evt_`).
- **Lab-Smoke** [`make smoke-ingest-control`](../../examples/ingest-control/smoke-lifecycle.sh)
  führt Stream-Anlage + Start- + Ende-Hook reproduzierbar aus.

Was `0.11.0` **nicht** liefert, steht in §6 dieser Datei und in
[`plan-0.11.0.md`](../planning/done/plan-0.11.0.md) §0.1.

## 2. Quickstart

### 2.1 API + SQLite hochfahren

```bash
make dev
```

Der Compose-Stack startet u. a. `api` mit SQLite-Volume; die
Ingest-Control-Routen sind aktiv, sobald die Migrations V1..V3
durchgelaufen sind.

### 2.2 Stream anlegen

```bash
curl -sS -X POST http://localhost:8080/api/ingest/streams \
  -H "X-MTrace-Token: demo-token" \
  -H "Content-Type: application/json" \
  -d '{
        "display_name": "Lab SRT",
        "protocol": "srt",
        "endpoint_id": "ep-srt",
        "target_id": "tgt-mediamtx"
      }'
```

Antwort `201 Created` enthält Stream-Metadaten plus
`stream_key.value` — den Klartext-Wert zeigt die API **genau einmal**.
Aufruferseitig direkt in den Publisher (FFmpeg, OBS, …) übernehmen,
serverseitig gibt es ihn nie wieder.

### 2.3 Key lokal validieren

```bash
curl -sS -X POST http://localhost:8080/api/ingest/streams/ing_<id>/validate-key \
  -H "X-MTrace-Token: demo-token" \
  -H "Content-Type: application/json" \
  -d '{"stream_key":"mtr_ing_..."}'
```

`200 {"valid":true,"stream_id":"ing_...","key_fingerprint":"..."}` bei
Match, sonst `200 {"valid":false}` ohne Detail-Hinweise. Dieses
Endpoint ist **kein** produktiver Auth-Pfad — Cross-Project-
Existenzprüfungen werden bewusst nicht ausgeliefert.

### 2.4 Route prüfen / MediaMTX-Artefakt anzeigen

```bash
curl -sS http://localhost:8080/api/ingest/media-server-config \
  -H "X-MTrace-Token: demo-token" | jq -r '.config_yaml'
```

Antwort enthält `target_id`, `kind`, `config_path`, `config_yaml`
und `warnings`. Existieren mehrere Targets im Project ohne expliziten
`?target_id=`-Filter, listet `warnings` die übrigen Target-IDs.

### 2.5 Key rotieren

```bash
curl -sS -X POST http://localhost:8080/api/ingest/streams/ing_<id>/rotate-key \
  -H "X-MTrace-Token: demo-token" -H "Content-Type: application/json" -d '{}'
```

Antwort `200` enthält den **neuen** Klartext-Key (einmalig); der
alte wird in Persistenz auf `deactivated_at = now()` gesetzt und
schlägt ab dem Zeitpunkt im Validate-Pfad fehl.

### 2.6 Lifecycle-Hook einspeisen

```bash
make smoke-ingest-control
```

Der Smoke ist die kanonische Verifikation: Stream anlegen,
`stream-started` einspeisen, `stream-ended` einspeisen, in beiden
Fällen `202 accepted:true`. Default-API-URL `http://localhost:8080`,
Default-Token `demo-token`; beides via `MTRACE_API_URL` und
`MTRACE_API_TOKEN` überschreibbar.

## 3. API-Kontrakt

Vollständige Wire-Skizze, Auth-Matrix, Fehlercodes und Redaktionsregeln
in [`spec/backend-api-contract.md`](../../spec/backend-api-contract.md)
§3.8. Die wichtigsten Antworten und ihre Statuscodes:

| Endpoint | Erfolg | Wichtige Fehlercodes |
| --- | --- | --- |
| `POST /api/ingest/streams` | `201` | `400 invalid_request`, `401 unauthorized`, `409 stream_name_conflict`, `404 endpoint_not_found`/`target_not_found` |
| `GET /api/ingest/streams` | `200` | `401` |
| `GET /api/ingest/streams/{id}` | `200` | `401`, `404 stream_not_found` (auch für Cross-Project) |
| `POST /api/ingest/streams/{id}/rotate-key` | `200` | `401`, `404`, `409 routing_rule_disabled` |
| `POST /api/ingest/streams/{id}/validate-key` | `200` (immer) | `401` |
| `GET /api/ingest/media-server-config` | `200` | `401`, `404 target_not_found`, `503 media_server_config_unavailable` |
| `POST /api/ingest/hooks/stream-{started,ended}` | `202` | `400`, `401`, `404`, `409` |

### Redaktionsregeln (RAK-66 / NF-13)

- **Klartext-Stream-Keys** erscheinen ausschließlich im Create-/
  Rotate-Response. Persistenz, Logs, Lifecycle-Events,
  MediaMTX-Artefakt führen nur Hash und Fingerprint.
- **Lifecycle-Events** tragen `key_fingerprint` (≠ Schlüssel),
  `connection_id` und `reason` sind dokumentarisch (≤ 256 Zeichen).
- **Validate-Antwort** ist responsiv-blind: `valid:false` enthält
  niemals einen `stream_id`-Hinweis.

## 4. Lab-Smokes

Alle Lab-Smokes sind opt-in und **nicht** Teil von `make gates`:

| Target | Zweck |
| --- | --- |
| `make smoke-ingest-control` | Lifecycle-Hooks gegen lokale API (RAK-69) |
| `make smoke-srt` / `make smoke-srt-health` | Bestehende SRT-/HLS-Lab-Smokes (RAK-37/RAK-43) |
| `make smoke-mediamtx` | MediaMTX-Core-Lab (RAK-36) |

Das Generator-Beispiel in `examples/ingest-control/` läuft **ohne**
laufende API: `mediamtx.generated.yml` ist commit-stabil und kann
direkt in einen lokalen MediaMTX-Stack gemountet werden
(siehe [`examples/ingest-control/README.md`](../../examples/ingest-control/README.md)).

## 5. Security-Grenze

Was `0.11.0` **bewusst nicht** ist:

- **Keine produktive Control-Plane.** Es gibt keine Multi-Tenant-
  Token-Lifecycle, keine signierten Sessions, keine IP-Allowlist,
  kein Replay-Schutz auf den Hooks. Tokens kommen aus dem statischen
  Project-Resolver für Lab-Setups.
- **Keine ausgehende Webhook-Zustellung.** Der Hook-Endpoint
  empfängt Events von außen (lokal/Lab); m-trace ruft selbst keine
  externen Webhooks auf.
- **Keine globale Stream-Key-Rotation.** Rotation ist pro Stream;
  Project-/Tenant-weite Rotationsstrategien sind Folge-Scope.
- **Kein produktives Secret-Management.** Stream-Keys werden in
  SQLite gehashed gespeichert, aber die Lab-Compose hat kein
  Secrets-Backend. Für Production-Setups gehört der Key-Store hinter
  einen externen KMS-/Vault-Pfad.

Diese Grenzen lösen `0.12.0` (signierte Session-Tokens,
Project-Token-Rotation, tenant-spezifische Ingest-Policies) und
spätere Phasen (externe Media-Server-Provisionierung,
Dashboard-UI für Stream-Control) auf — siehe
[`docs/planning/in-progress/roadmap.md`](../planning/in-progress/roadmap.md)
und [`plan-0.11.0.md`](../planning/done/plan-0.11.0.md) §10.

## 6. Was `0.11.0` nicht liefert

- Externe MediaMTX-Hook-Integration (Lifecycle-Events kommen aus
  dem Lab-Smoke oder einem manuellen Curl, nicht aus MediaMTX
  selbst — Folge-Scope laut R-16 in
  [`risks-backlog.md`](../planning/in-progress/risks-backlog.md)).
- Provisionierung der MediaMTX-Control-API. Das generierte YAML
  wird als Datei ausgespielt; MediaMTX liest es per Volume-Mount.
- Dashboard-UI für Stream-Control. Der CRUD-Pfad ist API-only.
- Cross-Project-Lookups oder Existenz-Hinweise im Validate- bzw.
  Stream-Detail-Endpoint (deliberate `404`/`valid:false`).
