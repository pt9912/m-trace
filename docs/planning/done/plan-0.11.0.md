# Implementation Plan — `0.11.0` (Ingest-Gateway / Stream Control)

> **Status**: ✅ released (`v0.11.0`, 2026-05-09). Vorgänger `0.10.0`
> ist ebenfalls released (Tag `v0.10.0` auf `d384569`, Plan in
> [`done/plan-0.10.0.md`](../done/plan-0.10.0.md)).
>
> **Release-Typ**: Minor-Release mit Lastenheft-Patch `1.1.14`
> (Vorschlag), neuer RAK-Gruppe `RAK-65`..`RAK-70`,
> RAK-Verifikationsmatrix und Tag `v0.11.0`.
>
> **Ziel**: Die bisher als Kann geführten Ingest-Gateway-Funktionen
> `F-46`..`F-51` werden in einen umsetzbaren Produkt-Scope geschnitten:
> ein lokal betreibbarer Stream-Control-Pfad für Lab-Streams,
> Stream-Keys, Ingest-Endpunkte, Routing-Regeln, lokale
> Lifecycle-Events und MediaMTX-nahe Konfigurationsartefakte.
> `0.11.0` liefert ausdrücklich keine mandantenfähige Control-Plane
> und keine produktive SaaS-Orchestrierung.
>
> **Bezug**:
> [`spec/lastenheft.md`](../../../spec/lastenheft.md) §7.5.4
> `apps/ingest-gateway` (`F-46`..`F-51`), §12.3 `MVP-38`,
> §13.12 `RAK-60`..`RAK-64`;
> [`README.md`](../../../README.md) mit der Überschrift
> „Was m-trace nicht ist";
> [`examples/README.md`](../../../examples/README.md) für
> Multi-Protocol-Lab-Konventionen;
> [`examples/srt/`](../../../examples/srt/),
> [`examples/mediamtx/`](../../../examples/mediamtx/),
> [`examples/srs/`](../../../examples/srs/).
>
> **Nachfolger**: voraussichtlich `0.12.0` (Auth / Token Lifecycle).
> Alles, was tenant-spezifische Ingest Policies, signierte Session
> Tokens oder Project-Token-Rotation braucht, wird dort behandelt und
> nicht in diesen Plan gezogen.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Scope-, Security- oder Architekturentscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.11.0` liefert **Stream Control für lokale/lab-nahe Ingest-Flows**.
Der Release schafft ein stabiles Modell und einen reproduzierbaren
Lab-Pfad, aber noch keine produktive Ingest-Control-Plane.

In Scope:

- `F-46`: Stream-Key-Verwaltung für lokale Ingest-Streams.
  - Stream-Key wird serverseitig erzeugt oder erneuert.
  - API-Antworten dürfen den Klartext-Key nur beim Anlegen bzw.
    Rotieren zurückgeben.
  - Persistenz speichert kein Klartext-Secret; falls persistiert wird,
    dann nur Key-Hash, redigierter Fingerprint plus Metadaten.
  - Logs, Fixtures, Doku und Smokes verwenden ausschließlich
    Beispielwerte oder redigierte Keys.
- `F-47`: Ingest-Endpunkte beschreiben.
  - Unterstützte Protokolle im `0.11.0`-Scope: `srt`, `rtmp`.
  - Endpunkte beschreiben Host/Port/Path, Protokoll, lokalen
    Lab-Hinweis und optionalen Egress-Hinweis, ohne externe
    Infrastruktur zu provisionieren.
  - Bestehende Lab-Stacks bleiben Quelle der Wahrheit für reale
    Ports und Startbefehle.
- `F-48`: einfache Routing-Regeln modellieren.
  - Eine Regel verbindet einen `IngestStream` mit genau einem
    `MediaServerTarget`.
  - Priorisierung, Fan-out, Failover und dynamisches Load-Balancing
    bleiben Folge-Scope.
  - Regeln sind deterministisch validierbar und als JSON-Artefakt
    exportierbar.
- `F-49`: Stream-Lifecycle-Events vorbereiten und lokal verifizieren.
  - Eventmodell für `stream_started` und `stream_ended`.
  - Lifecycle-Adapter kann Events exemplarisch empfangen oder in einem
    lokalen Smoke auslösen; produktive ausgehende Webhook-Zustellung
    an externe Systeme ist optionaler Folge-Scope.
  - Events enthalten keine Klartext-Keys.
- `F-50`: SRT-/RTMP-Konfigurationen als beschreibbare Artefakte
  vorbereiten.
  - Artefakte sind Lab-orientiert, reviewbar und reproduzierbar.
  - Keine direkte Manipulation laufender externer Server.
- `F-51`: Media-Server-Konfigurationen generieren oder validieren.
  - Normativer Zielserver für `0.11.0`: MediaMTX im vorhandenen
    Lab-Scope.
  - SRS darf als Kompatibilitäts-/Dokuhintergrund erwähnt werden,
    ist aber kein Pflicht-Target.
  - Generierung darf auf ein eigenes Beispielverzeichnis begrenzt
    bleiben; bestehende `examples/`-Stacks dürfen nicht brechen.

Out of scope:

- Keine mandantenfähige Control-Plane.
- Keine produktive Secret-Verwaltung und keine KMS-/Vault-Integration.
- Keine globale Stream-Key-Rotation über mehrere Deployments.
- Keine produktive Media-Server-Auth-/Key-Enforcement-Kopplung; die
  Key-Prüfung bleibt ein lokaler API-/Smoke-Nachweis und ist kein
  Ersatz für spätere signierte Tokens oder MediaMTX/SRS-Auth-Hooks.
- Keine automatische Provisionierung externer Media-Server.
- Kein Kubernetes-Operator und keine K8s-Manifeste.
- Keine Auth-/Token-Lifecycle-Arbeiten aus `0.12.0`.
- Keine UI-Pflicht im Dashboard; eine kleine Diagnoseansicht darf nur
  entstehen, wenn API-/Doku-Scope dadurch nicht verdrängt wird.
- Keine verbindliche Runtime-Korrelation zwischen Ingest-Gateway,
  Player-SDK und Analyzer über neue Trace-Felder; bestehende
  OTel-/Session-Modelle bleiben unverändert.

### 0.2 Vorgänger-Gate

- `0.10.0` ist released; der Plan liegt unter
  `docs/planning/done/plan-0.10.0.md`.
- Lastenheft steht vor Aktivierung bei `1.1.13` mit RAK-60..RAK-64.
- `examples/`-Konventionen aus `0.5.0` gelten weiter:
  eigenständige Lab-Beispiele nutzen eigenen Compose-Project-Namen
  und opt-in Smoke-Targets.
- SRT-Health aus `0.6.0`, WebRTC aus `0.7.0`/`0.8.0`, SRS aus
  `0.9.0` und CMAF aus `0.10.0` bleiben Regression-Baseline.

### 0.3 Architekturentscheidung

Für `0.11.0` ist **Variante B verbindlich**: Ingest-Control wird als
Modul in `apps/api` umgesetzt. Variante A bleibt ein möglicher späterer
Ausgliederungspfad, ist aber nicht Teil dieses Release-Scope.

| Variante | Beschreibung | Vorteil | Risiko |
| -------- | ------------ | ------- | ------ |
| A | eigenes `apps/ingest-gateway` nach Lastenheft §7.5.4 | klare Service-Grenze für spätere Control-Plane | neue App, Dockerfile, Port, CI- und Doku-Aufwand für lokalen Scope |
| B | Ingest-Control als Modul in `apps/api` | nutzt vorhandene HTTP-, SQLite-, Metrik- und Test-Infrastruktur | Name `ingest-gateway` bleibt zunächst konzeptionell, spätere Ausgliederung braucht Migration |

Begründung: Der Release ist bewusst lokal/lab-nah. Ein zusätzlicher
Prozess wäre für `F-46`..`F-51` mehr Betriebsoberfläche als
Produktnutzen. Die Domain muss aber so geschnitten werden, dass eine
spätere Ausgliederung in `apps/ingest-gateway` möglich bleibt. Deshalb
gelten für Tranche 1 explizite Hexagon-/Port-Grenzen innerhalb
`apps/api`; HTTP, SQLite, Auth-/Project-Resolver, CORS und Metriken
werden aus der bestehenden API-Infrastruktur wiederverwendet.

### 0.4 Persistenzentscheidung (vor Tranche 1 zu schließen)

Standardvorschlag: SQLite über die bestehende API-Persistenz, falls
Variante B gewählt wird. Reine Konfigurationsartefakte sind nur
zulässig, wenn folgende Punkte trotzdem stabil erfüllt werden:

- Stream-IDs bleiben über API-Restarts reproduzierbar.
- Rotierte Keys können alte Key-Hashes deaktivieren.
- Contract-Tests können Listen, Lesen, Rotieren und Validieren ohne
  Testreihenfolge-Abhängigkeit prüfen.
- Doku macht klar, ob Daten in SQLite oder nur in generierten
  Artefakten leben.

### 0.5 Lastenheft-Patch `1.1.14` (Vorschlag)

Der Patch ergänzt `spec/lastenheft.md` um RAK-65..RAK-70 und hebt
`F-46`..`F-51` für den begrenzten `0.11.0`-Lab-Control-Scope von
Kann auf Release-Muss. `MVP-38` wird dabei ausdrücklich als lokaler
SRT-/RTMP-Ingest-Control-Smoke für MediaMTX-nahe Lab-Artefakte
präzisiert und für diesen begrenzten Scope auf Release-Muss gezogen;
die ältere Kann-Historie bleibt auditierbar. Verbindlich ist die neue
RAK-Gruppe. `F-49` wird für `0.11.0` ausdrücklich als lokales
Lifecycle-Eventmodell plus reproduzierbarer Empfang/Auslösung
präzisiert; produktive ausgehende Webhook-Zustellung bleibt Folge-Scope
und darf nicht als erfüllt behauptet werden.

| RAK | Priorität | Inhalt |
| --- | --------- | ------ |
| RAK-65 | Muss | Ingest-Control-Scope ist normativ begrenzt: lokale/lab-nahe Stream-Verwaltung, keine Multi-Tenant-Control-Plane, keine produktive Secret-Verwaltung, keine externe Media-Server-Provisionierung. |
| RAK-66 | Muss | Stream-Key-Verwaltung: Streams können angelegt, gelistet, lokal validiert und rotiert werden; Klartext-Keys erscheinen nur bei Anlage/Rotation, nicht in Logs, Fixtures oder Persistenz. |
| RAK-67 | Muss | Ingest-Endpunkt- und Routing-Modell: `srt`/`rtmp`-Endpunkte, Stream-Ziele und einfache 1:1-Routing-Regeln sind validiert, dokumentiert und per API/Artefakt stabil beschreibbar. |
| RAK-68 | Muss | Media-Server-Artefakte: MediaMTX-nahe Konfigurationen für SRT und RTMP im Lab-Scope können generiert oder validiert werden; bestehende Multi-Protocol-Lab-Beispiele und Smokes bleiben grün. |
| RAK-69 | Muss | Lifecycle-Events: `stream_started` und `stream_ended` besitzen ein stabiles Eventmodell und werden lokal reproduzierbar empfangen oder exemplarisch ausgelöst; Events enthalten keine Klartext-Keys. Produktive ausgehende Webhook-Zustellung ist nicht Teil des `0.11.0`-Nachweises. |
| RAK-70 | Muss | Doku, API-/Contract-Tests und Release-Smokes beschreiben den lokalen Stream-Control-Workflow, die Sicherheitsgrenzen und den Unterschied zu Auth-/Tenant-Folge-Scope `0.12.0`. |

### 0.6 Öffentliche API und Modell-Skizze

Für Variante B werden die Lastenheft-Pfade unter `apps/api` verwendet:

Alle `/api/ingest/*`-Endpunkte sind tokenpflichtig und folgen der
bestehenden `X-MTrace-Token`-/Project-Resolver-Konvention aus
`spec/backend-api-contract.md` §4. `project_id` wird serverseitig aus
dem Token abgeleitet; ein optionaler Request-`project_id`-Wert darf nur
als Konsistenzcheck dienen und muss zum Token passen. Listen, Details,
Rotation, Key-Validierung und Lifecycle-Events sind immer auf das
aufgelöste Project gefiltert. CORS-Preflight und Fehlerreihenfolge
werden im API-Kontrakt mitgepflegt.

| Methode | Pfad | Zweck |
| ------- | ---- | ----- |
| `POST` | `/api/ingest/streams` | Ingest-Stream anlegen; gibt Stream-Metadaten plus Klartext-Key genau einmal zurück |
| `GET` | `/api/ingest/streams` | Streams listen; ohne Klartext-Key |
| `GET` | `/api/ingest/streams/{id}` | Stream-Details, Endpunkte und Routing-Regel lesen; ohne Klartext-Key |
| `POST` | `/api/ingest/streams/{id}/rotate-key` | Key rotieren; gibt neuen Klartext-Key genau einmal zurück |
| `POST` | `/api/ingest/streams/{id}/validate-key` | lokalen Stream-Key gegen aktive Key-Hashes prüfen; Antwort enthält keinen Klartext-Key und ist kein produktiver Media-Server-Auth-Pfad |
| `POST` | `/api/ingest/hooks/stream-started` | lokales Start-Event empfangen oder Smoke-Event einspeisen |
| `POST` | `/api/ingest/hooks/stream-ended` | lokales Ende-Event empfangen oder Smoke-Event einspeisen |
| `GET` | `/api/ingest/media-server-config` | generiertes/validiertes MediaMTX-Artefakt abrufen oder Diagnose liefern |

Normative Wire-Skizze für Contract-Tests:

`POST /api/ingest/streams` Request:

```json
{
  "display_name": "Lab SRT",
  "protocol": "srt",
  "endpoint_id": "mediamtx-srt-local",
  "target_id": "mediamtx-local",
  "project_id": "demo"
}
```

`project_id` ist optional und dient nur als Konsistenzcheck zum
`X-MTrace-Token`; fehlt das Feld, wird es serverseitig aus dem Token
abgeleitet.

`POST /api/ingest/streams` und
`POST /api/ingest/streams/{id}/rotate-key` Response:

```json
{
  "stream": {
    "id": "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "project_id": "demo",
    "display_name": "Lab SRT",
    "protocol": "srt",
    "endpoint_id": "mediamtx-srt-local",
    "target_id": "mediamtx-local",
    "routing_rule_id": "route_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "status": "ready",
    "created_at": "2026-05-09T10:00:00Z",
    "updated_at": "2026-05-09T10:00:00Z"
  },
  "stream_key": {
    "value": "mtr_ing_7YQ3pVh4v0hT8x2l9b6nR4c1A5sD0eF2gH3jK8mN9pQ",
    "fingerprint": "mtr_ing_7YQ3...N9pQ",
    "created_at": "2026-05-09T10:00:00Z"
  }
}
```

`stream_key.value` darf ausschließlich in Create-/Rotate-Antworten
erscheinen. List-, Detail-, Event-, Fehler- und Artefakt-Antworten
enthalten höchstens `fingerprint`.

`GET /api/ingest/streams` Response:

```json
{
  "streams": [
    {
      "id": "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
      "project_id": "demo",
      "display_name": "Lab SRT",
      "protocol": "srt",
      "endpoint_id": "mediamtx-srt-local",
      "target_id": "mediamtx-local",
      "routing_rule_id": "route_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
      "status": "ready",
      "key_fingerprint": "mtr_ing_7YQ3...N9pQ",
      "created_at": "2026-05-09T10:00:00Z",
      "updated_at": "2026-05-09T10:00:00Z"
    }
  ]
}
```

`GET /api/ingest/streams/{id}` Response ergänzt die referenzierten
Objekte, aber keinen Klartext-Key:

```json
{
  "stream": {
    "id": "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "project_id": "demo",
    "display_name": "Lab SRT",
    "protocol": "srt",
    "endpoint_id": "mediamtx-srt-local",
    "target_id": "mediamtx-local",
    "routing_rule_id": "route_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "status": "ready",
    "key_fingerprint": "mtr_ing_7YQ3...N9pQ",
    "created_at": "2026-05-09T10:00:00Z",
    "updated_at": "2026-05-09T10:00:00Z"
  },
  "endpoint": {
    "id": "mediamtx-srt-local",
    "protocol": "srt",
    "listen_host": "127.0.0.1",
    "listen_port": 8890,
    "path_template": "publish:{stream_path}",
    "lab_stack": "mtrace-srt",
    "public_url_hint": "srt://localhost:8890?streamid=publish:{stream_path}"
  },
  "routing_rule": {
    "id": "route_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "stream_id": "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "target_id": "mediamtx-local",
    "mode": "single",
    "enabled": true
  },
  "target": {
    "id": "mediamtx-local",
    "kind": "mediamtx",
    "config_path": "examples/ingest-control/mediamtx.generated.yml",
    "hls_url_template": "http://localhost:8889/{stream_path}/index.m3u8",
    "control_api_url": "http://localhost:9998"
  }
}
```

`POST /api/ingest/streams/{id}/validate-key` Request/Response:

```json
{
  "stream_key": "mtr_ing_7YQ3pVh4v0hT8x2l9b6nR4c1A5sD0eF2gH3jK8mN9pQ"
}
```

```json
{
  "valid": true,
  "stream_id": "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
  "key_fingerprint": "mtr_ing_7YQ3...N9pQ"
}
```

Ungültige, unbekannte, rotierte oder deaktivierte Keys liefern denselben
dokumentierten Fehlercode `ingest_key_invalid`; Tests dürfen keine
unterscheidbaren Fehlerantworten oder Timing-Annahmen für diese Fälle
einführen.

`POST /api/ingest/hooks/stream-started` Request:

```json
{
  "stream_id": "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
  "observed_at": "2026-05-09T10:01:00Z",
  "source": "local-smoke",
  "connection_id": "srtconn-1"
}
```

`POST /api/ingest/hooks/stream-ended` Request:

```json
{
  "stream_id": "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
  "observed_at": "2026-05-09T10:05:00Z",
  "source": "local-smoke",
  "connection_id": "srtconn-1",
  "reason": "smoke_complete"
}
```

Lifecycle-Erfolgsantworten enthalten mindestens `event_id`,
`stream_id`, `type`, `observed_at` und `accepted:true`; sie enthalten
keinen Klartext-Key.

`GET /api/ingest/media-server-config` Response:

```json
{
  "kind": "mediamtx",
  "format": "yaml",
  "artifact_path": "examples/ingest-control/mediamtx.generated.yml",
  "generated_at": "2026-05-09T10:00:00Z",
  "streams": ["ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9"],
  "warnings": []
}
```

Das Artefakt selbst darf nur Beispielwerte oder redigierte
Fingerprints enthalten.

Pflicht-Domainobjekte:

- `IngestStream`: `id`, `project_id`, `display_name`, `protocol`,
  `endpoint_id`, `target_id`, `routing_rule_id`, `status`,
  `created_at`, `updated_at`; `project_id` stammt aus dem
  authentifizierten Project-Kontext.
- `StreamKey`: `stream_id`, `key_hash`, `fingerprint`, `created_at`,
  `rotated_at?`, `disabled_at?`; kein
  Klartextfeld in Persistenz. `fingerprint` ist nur ein redigierter
  Anzeige-/Audit-Wert und darf nicht als verifier ausreichen.
- `IngestEndpoint`: `id`, `protocol`, `listen_host`, `listen_port`,
  `path_template`, `lab_stack`, `public_url_hint?`.
- `RoutingRule`: `id`, `stream_id`, `target_id`, `mode:"single"`,
  `enabled`.
- `MediaServerTarget`: `id`, `kind:"mediamtx"`, `config_path?`,
  `hls_url_template?`, `control_api_url?`.
- `StreamLifecycleEvent`: `type`, `stream_id`, `observed_at`,
  `source`, `connection_id?`, `reason?`.

Validierungsregeln:

- `display_name` und Stream-Pfad müssen stabil normalisiert werden;
  doppelte aktive Namen sind pro Project unzulässig.
- `protocol` ist in `0.11.0` nur `srt` oder `rtmp`.
- Host/Port dürfen keine externen Server implizit provisionieren.
- Routing-Ziel muss existieren und `kind:"mediamtx"` sein.
- Der lokale Key-Validierungspfad akzeptiert nur den aktuell aktiven
  `key_hash`; rotierte oder deaktivierte Key-Hashes müssen stabil
  abgelehnt werden. Vergleiche laufen konstantzeitnah und antworten
  ohne Klartext-Key.
- Stream-Keys werden mit einem CSPRNG mit mindestens 256 Bit Entropie
  erzeugt und in einem dokumentierten, URL-sicheren Format ausgegeben
  (z. B. Prefix plus Base64url-Token). Persistiert wird ein
  verifier-tauglicher Hash über den vollständigen Key; zusätzlich darf
  ein kurzer redigierter Fingerprint für Logs, Audit und UI-Diagnose
  gespeichert werden.
- Fehlercodes sind stabil und werden in Contract-Tests gepinnt, z. B.
  `ingest_stream_duplicate`, `ingest_protocol_unsupported`,
  `ingest_endpoint_missing`, `ingest_route_invalid`,
  `ingest_key_not_found`, `ingest_key_invalid`.

### 0.7 Security- und Logging-Grenzen

`0.11.0` darf Security nicht vortäuschen. Deshalb gilt:

- Kein Auth-Versprechen jenseits bestehender API-Mechanismen.
- `/api/ingest/*` nutzt diese bestehenden API-Mechanismen verbindlich:
  fehlender oder ungültiger `X-MTrace-Token` führt zu `401`, und alle
  Datenzugriffe sind project-gescoped.
- Keine tenant-spezifischen Policies; Verweis auf `0.12.0`.
- Stream-Keys sind lokale Lab-Secrets, nicht produktive Zugangsdaten.
- Klartext-Keys dürfen nur im Create-/Rotate-Response und in
  bewusst markierten Beispielkonfigurationen erscheinen.
- Logs, Metriken, Traces, Fehlerantworten und Fixtures enthalten nur
  Fingerprints oder redigierte Werte.
- Threat-Model-Notiz in der Doku muss mindestens Replay, Key-Leakage,
  Log-Leakage und Lab-vs-Production-Grenze nennen.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung, Lastenheft-Patch `1.1.14`, RAK-Gruppe, Architektur- und Persistenzentscheidung | ✅ |
| 1 | Stream-Key-, Ingest-Endpunkt- und Routing-Domainmodell | ✅ |
| 2 | API-/Persistenzpfad für Streams, Listing, Key-Validierung und Key-Rotation | ✅ |
| 3 | MediaMTX-Artefakte und SRT-/RTMP-Lab-Konfiguration | 🟡 |
| 4 | Lifecycle-Events und lokale Lab-Verifikation | ✅ |
| 5 | Doku, Contract-Tests, Smokes und README-Abgrenzung | ✅ |
| 6 | Gates, RAK-Verifikationsmatrix, Versions-Bump, Closeout und Tag | ✅ |

---

## 2. Tranche 0 — Aktivierung, Patch und Entscheidungen

Ziel: Der Release-Scope wird vor Implementierung normativ und
architektonisch geschlossen.

DoD:

- [x] Plan von `docs/planning/open/plan-0.11.0.md` nach
  `docs/planning/in-progress/plan-0.11.0.md` verschoben.
- [x] `git status --short` vor erster Änderung dokumentiert: working
  tree clean (Tag `v0.10.0` auf `d384569`; danach
  `68e83a3`/`055e56e`/`6ba7a17`/`11b1185`/`3ba39fb`/`86e71c4`
  Plan-Tightening-Commits + CMAF-Contract-Fixtures auf `main`).
- [x] `spec/lastenheft.md` Header auf `1.1.14` erhöht.
- [x] RAK-65..RAK-70 im Lastenheft ergänzt (neuer §13.13).
- [x] `F-46`..`F-51` im Lastenheft für den `0.11.0`-Scope
  nachvollziehbar von Kann-Historie auf Release-Muss abgebildet
  (jeweils mit „Muss (`0.11.0`-Scope, Patch `1.1.14`)"-Stufung und
  Verweis auf RAK-66/RAK-67/RAK-68/RAK-69 in §13.13).
- [x] `MVP-38` im Lastenheft als lokaler SRT-/RTMP-Ingest-Control-
  Smoke (`make smoke-ingest-control`) für MediaMTX-nahe Lab-
  Artefakte präzisiert und für den `0.11.0`-Scope auf Release-Muss
  abgebildet (Verweis auf RAK-68 in §13.13).
- [x] `spec/backend-api-contract.md` erweitert die Endpunktmatrix
  (§2 — neun neue Zeilen `POST/GET /api/ingest/*`) plus neuen
  §3.8 mit Auth-Matrix, CORS-Preflight, sieben-stufiger
  Fehlerreihenfolge und Wire-Skizzen aus Plan §0.6 (Create/Rotate-
  Response mit `stream_key.value`, List/Detail-Response mit
  `key_fingerprint`, Validate-Endpoint ohne Cross-Project-Leak,
  Lifecycle-Hook-Payload, Media-Server-Config-Antwort) plus
  zusätzlicher Fehler-Code-Tabelle (`project_id_mismatch`,
  `stream_not_found`, `endpoint_not_found`/`target_not_found`,
  `stream_name_conflict`, `routing_rule_disabled`, `key_invalid`,
  `media_server_config_unavailable`).
- [x] Patch-Log in `docs/planning/done/plan-0.1.0.md` um
  `Patch 1.1.14` (§4a.17) ergänzt.
- [x] Architekturentscheidung dokumentiert: `0.11.0` nutzt
  verbindlich Variante B als `apps/api`-Modul (Plan §0.3,
  Lastenheft §13.13 RAK-65, Roadmap §1.2); Variante A ist nur
  Folge-Scope.
- [x] Persistenzentscheidung dokumentiert: **SQLite** über die
  bestehende API-Persistenz mit neuer Migration (Plan §0.4 +
  T2-DoD); Stream-IDs überleben API-Restarts, Key-Rotation
  deaktiviert alte `key_hash`-Werte, Contract-Tests laufen ohne
  Testreihenfolge-Abhängigkeit. Reine Artefakt-only-Variante
  bewusst verworfen, weil Validate-Endpoint persistente
  Hash-Lookup-Garantie braucht.
- [x] Roadmap-Status und Release-Übersicht auf `0.11.0` als
  aktive Folgephase umgestellt (§1 Phase-Header, §1.2 Folge-Scope,
  §2 Schritt 46 von ⬜ auf 🟡, §3 Tabellenzeile auf 🟡).
- [x] Risiko-/Folge-Scope-Liste aktualisiert: Auth/Tenant/Policy
  nach `0.12.0`, externe Provisionierung Folge-Scope, produktive
  Webhook-Zustellung Folge-Scope (Roadmap §1.2 Out-of-Scope-
  Block).

## 3. Tranche 1 — Domainmodell und Validierung

Ziel: Stream Control ist als reines Domainmodell testbar, bevor HTTP,
Storage oder Media-Server-Artefakte angebunden werden.

DoD:

- [x] Domainobjekte `IngestStream`, `StreamKey`, `IngestEndpoint`,
  `RoutingRule`, `MediaServerTarget` und `StreamLifecycleEvent` in
  `apps/api/hexagon/domain/ingest_stream.go` definiert.
- [x] `IngestProtocol`-Enum auf `srt`/`rtmp` begrenzt
  (`IngestProtocol.IsKnown` + `ValidateIngestProtocol`); unbekannte
  Werte liefern `ErrIngestProtocolUnknown`. HTTP-Adapter mappt das
  in T2 auf `400 invalid_request`.
- [x] Stream-Key-Erzeugung
  (`apps/api/hexagon/domain/stream_key.go`,
  `GenerateStreamKey`) nutzt `crypto/rand` mit 32 Byte = 256 Bit
  Entropie. URL-sicheres `base64.RawURLEncoding` mit Prefix
  `mtr_ing_`; SHA-256-Hex-Hash und redigierter Fingerprint
  (`mtr_ing_<head8>...<tail4>`) getrennt vom Klartext berechnet.
  `StreamKeyMaterial.ToPersistable()` extrahiert die persistente
  Sicht ohne Klartext.
- [x] Key-Validierung (`ValidateStreamKey`) nutzt den vollständigen
  Hash mit `crypto/subtle.ConstantTimeCompare`; Fingerprint ist
  reine Anzeigeform und kein verifier (Doku-Comment +
  `TestStreamKeyMaterial_ToPersistableExcludesValue`).
- [x] Validierungsregeln + Fehler-Konstanten:
  `ErrIngestProtocolUnknown`, `ErrIngestStreamNotFound`,
  `ErrIngestStreamNameConflict`, `ErrIngestEndpointNotFound`,
  `ErrIngestTargetNotFound`, `ErrIngestRoutingRuleDisabled`,
  `ErrIngestProjectIDMismatch`, `ErrIngestKeyInvalid`,
  `ErrIngestMediaServerConfigUnavailable`,
  `ErrStreamKeyMalformed`. Cross-Project-Leak-Schutz über
  `FilterStreamForProject` (Streams aus fremden Projekten →
  `ErrIngestStreamNotFound`, kein Hinweis auf Existenz).
- [x] Domain-Tests
  (`ingest_stream_test.go` + `stream_key_test.go`, ~13 Tests)
  laufen ohne HTTP-Server, Docker oder MediaMTX — reine
  In-Memory-Validierung.
- [x] Kein Domain-Test speichert oder snapshotet echte
  Klartext-Keys; `TestGenerateStreamKey_Uniqueness` prüft 1000
  Keys auf Eindeutigkeit ohne sie zu loggen.

## 4. Tranche 2 — API, Persistenz und Key-Rotation

Ziel: Der lokale Stream-Katalog ist über stabile API-Verträge nutzbar
und über API-Restarts hinweg reproduzierbar, sofern SQLite gewählt
wurde.

DoD:

- [x] `POST /api/ingest/streams` legt Stream, Endpoint-Bezug,
  Routing-Regel und initialen Stream-Key im per Token aufgelösten
  Project an (`apps/api/adapters/driving/http/ingest.go` +
  `IngestControlService.CreateStream`); Antwort enthält Klartext-Key
  genau einmal in `stream_key.value`.
- [x] `GET /api/ingest/streams` listet Streams ohne Klartext-Key
  (`buildStreamSummaryPayload` mit `key_fingerprint`).
- [x] `GET /api/ingest/streams/{id}` liefert Details inkl.
  Endpoint/Target/Routing-Regel ohne Klartext-Key.
- [x] `POST /api/ingest/streams/{id}/rotate-key` deaktiviert den
  alten Key-Datensatz (`UPDATE stream_keys SET deactivated_at=…`)
  und gibt den neuen Klartext-Key genau einmal zurück.
- [x] `POST /api/ingest/streams/{id}/validate-key` prüft den
  Kandidaten gegen den aktiven Hash mit
  `crypto/subtle.ConstantTimeCompare`; rotierte/deaktivierte Keys
  werden abgelehnt (`stream_keys.deactivated_at IS NOT NULL` ist im
  `idx_stream_keys_active_unique`-Filter ausgeschlossen). Antwort
  trägt nie einen Klartext-Key; `valid:false` liefert keinen
  Stream-ID-Hinweis.
- [x] Persistenzpfad hat Tests im Application-Layer
  (`hexagon/application/ingest_control_service_test.go`, 11 Tests
  gegen InMemory-Repo: Happy Path Create, Reject-Unknown-Protocol,
  Reject-Project-ID-Mismatch, Reject-Empty-Display-Name,
  Duplicate-Active-Name, Reject-Missing-Endpoint,
  Cross-Project-Isolation für Read+Validate, Rotate-Deactivates-Old,
  Validate-Reject-Malformed, List-Filters-By-Project, Lifecycle-
  Event-No-Klartext-Key) plus HTTP-Wire-Tests
  (`adapters/driving/http/ingest_test.go`, 11 Tests). SQLite-
  Adapter ist als zweite Implementation des Driven-Ports verdrahtet
  und nutzt dieselben Domain-Fehler.
- [x] HTTP-Fehlercodes sind stabil und im API-Kontrakt dokumentiert
  (`spec/backend-api-contract.md` §3.8 Tabelle); Mapping in
  `writeIngestError` deckt: `invalid_request`, `project_id_mismatch`,
  `unauthorized`, `stream_not_found`, `endpoint_not_found`,
  `target_not_found`, `stream_name_conflict`, `routing_rule_disabled`,
  `unsupported_media_type`, `payload_too_large`, `internal_error`.
- [x] Alle Ingest-HTTP-Handler haben Contract-Tests für fehlenden
  Token (`Test...MissingTokenReturns401`), ungültiges Content-Type
  (`Test...RejectsNonJSONContentType`), Domain-Fehler-Mapping
  (`Test...MapsDomainErrors`), Cross-Project-Isolation
  (Validate-Test `valid:false` ohne Stream-ID), Body-Limit
  (`TestIngestHandler_RejectsLargeBody` → 413), und Wire-Vertrag-
  Konformität (List zeigt nur `key_fingerprint`, Validate-True
  zeigt Fingerprint aber kein `value`).
- [x] Logs und Request-Metriken enthalten keine Klartext-Keys —
  `IngestStreamHandler` schreibt nur `key_fingerprint` ins
  Response-Log; Use-Case loggt nichts; Repository persistiert nur
  Hash + Fingerprint.
- [x] Fehlerantworten und Validierungs-Timing unterscheiden nicht
  zwischen unbekanntem, rotiertem und ungültigem Klartext-Key —
  alle drei Pfade liefern `{"valid":false}`; ConstantTimeCompare
  pinnt das Timing.
- [x] SQLite-Migration ist versioniert (`V2__ingest.sql` als hand-
  gepflegter Folger zur d-migrate-`V1`); Drift-Check sieht das via
  `make generated-drift-check` als unverändert (kein Re-Generate-
  Pfad für V2+); In-Memory- und SQLite-Adapter teilen den
  Driven-Port `IngestStreamRepository`. Migration-Test
  (`internal/storage/migrate_internal_test.go::TestOpen_FreshStart`)
  pinnt jetzt 2 Rows in `schema_migrations`.

## 5. Tranche 3 — Routing und Media-Server-Artefakte

Ziel: Aus den Stream-Control-Daten entstehen überprüfbare
MediaMTX-nahe Lab-Artefakte, ohne laufende Fremdserver automatisch zu
verändern.

DoD:

- [x] Routing-Regeln sind als stabile JSON-Konfiguration über
  `GET /api/ingest/streams/{id}` (`routing_rule`-Block) und
  `GET /api/ingest/media-server-config` (mit Stream-IDs als
  YAML-Comment) beschreibbar; das `RoutingRule`-Domainmodell aus T1
  wird in T3 deterministisch ins YAML-Artefakt übersetzt.
- [x] MediaMTX-nahe Konfigurationsartefakte werden über
  `apps/api/hexagon/application/mediamtx_config.go`
  (`GenerateMediaMTXConfig`) deterministisch generiert; HTTP-Endpoint
  `GET /api/ingest/media-server-config` reicht das YAML als
  `config_yaml`-Feld durch (`apps/api/adapters/driving/http/ingest.go`
  `IngestMediaServerConfigHandler`).
- [x] SRT- und RTMP-Beispiele in `examples/ingest-control/README.md`
  trennen Ingest-URL (`srt://localhost:8891`, `rtmp://localhost:1936`),
  Playback-/HLS-URL (`http://localhost:8892/{path}/index.m3u8`) und
  Control-API-URL (`http://localhost:9999`) explizit; Port-Tabelle
  fixiert die Host-Mappings.
- [x] RTMP wird im `0.11.0`-Pflichtpfad über das additive MediaMTX-
  Artefakt nachgewiesen: `GenerateMediaMTXConfig` schaltet
  `rtmp: yes`/`rtmpAddress: :1935` ein, sobald mindestens ein Stream
  `protocol:"rtmp"` hat (`TestGenerateMediaMTXConfig_TogglesProtocolListeners`
  pinnt das); die `examples/ingest-control/mediamtx.generated.yml`
  zeigt den RTMP-Pfad-Block exemplarisch. Der `examples/srs/`-Pfad
  bleibt Kompatibilitäts-/Dokuhintergrund — der Generator lehnt
  `MediaServerKindSRS` mit einem expliziten Fehler ab.
- [x] Bestehende `examples/mediamtx`, `examples/srt`, `examples/dash`,
  `examples/webrtc` und `examples/srs` bleiben unverändert; nur
  `examples/README.md` bekommt eine zusätzliche Tabellen-Zeile für
  `ingest-control/`.
- [x] Neues Beispiel `examples/ingest-control/` folgt
  `examples/README.md`-Standard: eigener Project-Name
  `mtrace-ingest-control`, README mit 7-Punkt-Standard
  (Zweck/Kurzanleitung/Port-Verteilung/Generator-Dynamik/Wartung/
  Was-es-nicht-ist/Risiko-Hinweise), opt-in Smoke geplant für
  Tranche 5 (`make smoke-ingest-control`).
- [x] Artefakte enthalten nur Beispiel-/redigierte Stream-Keys: das
  Generator-Output trägt ausschließlich `key_fingerprint`-
  Comments — niemals den Klartext-Wert
  (`TestGenerateMediaMTXConfig_NoKlartextKeyInOutput` pinnt das);
  `mediamtx.generated.yml` im Repo nutzt sichtbar redigierte
  Beispiel-Fingerprints (`mtr_ing_SRT...DEMO`, `mtr_ing_RTMP...DEMO`).

## 6. Tranche 4 — Lifecycle-Events und Lab-Verifikation

Ziel: Stream-Start und Stream-Ende sind als lokale Ereignisse
modelliert und reproduzierbar verifizierbar.

DoD:

- [x] Eventmodell für `stream_started` und `stream_ended` ist
  dokumentiert (`spec/backend-api-contract.md` §3.8 Wire-Skizze
  inkl. Response-Form mit `event_id`, `accepted:true`).
- [x] Lifecycle-Endpoint akzeptiert valide Start-/Ende-Events und weist
  unbekannte Streams, ungültige Eventtypen und malformed Payloads
  stabil ab — `IngestLifecycleHookHandler` mit URL-getriebenem
  `Kind` (Body-`type` wird ignoriert), Source-Allowlist
  `local-smoke`/`mediamtx-hook`, Längenlimits über
  `domain.MaxLifecycleStringField`.
- [x] Lifecycle-Events enthalten `stream_id`, `observed_at`, `source`
  und optional `connection_id`/`reason`; sie enthalten keinen
  Klartext-Key (Service-Test
  `RecordLifecycleEvent_NoKlartextKey` pinnt das auch für die
  optionalen Felder).
- [x] Lokaler Smoke verifiziert mindestens einen Start-/Ende-Pfad
  reproduzierbar — `make smoke-ingest-control` →
  `examples/ingest-control/smoke-lifecycle.sh`.
- [x] Echte MediaMTX-Hooks werden in `0.11.0` **nicht** angebunden
  (`!`-Folge-Scope nach §10): das Eventmodell plus exemplarische
  lokale Auslösung über `local-smoke` decken RAK-69 ab; ausgehende
  produktive Webhook-Zustellung bleibt Folge-Scope (siehe R-16 im
  `risks-backlog.md`).

## 7. Tranche 5 — Doku, Contracts und Smokes

Ziel: Nutzer können den lokalen Stream-Control-Pfad nachvollziehen,
ohne ihn mit produktiver Control-Plane, Tenant-Policy oder Auth zu
verwechseln.

DoD:

- [x] User-Doku beschreibt den lokalen Stream-Control-Workflow:
  Stream anlegen, Key verwenden/lokal validieren, Route prüfen,
  MediaMTX-Artefakt ansehen, Key rotieren —
  `docs/user/ingest-control.md` §2 Quickstart.
- [x] API-Kontrakt dokumentiert Endpunkte, Erfolgsantworten,
  Fehlercodes und Redaktionsregeln für Secrets —
  `spec/backend-api-contract.md` §3.8 plus
  `docs/user/ingest-control.md` §3.
- [x] README grenzt `0.11.0` gegen Control-Plane, Multi-Tenant-
  Betrieb und Secret-Management ab — neue Bullet-Zeile in
  `Was m-trace nicht ist`-Sektion mit Verweis auf
  `docs/user/ingest-control.md` §5.
- [x] `docs/user/local-development.md` verlinkt den Smoke- und
  Beispielpfad — neuer §2.7.2 mit `make smoke-ingest-control` und
  Verweis auf `docs/user/ingest-control.md`.
- [x] Relevante Smokes sind im Makefile dokumentiert;
  Lab-Smokes bleiben opt-in. `make smoke-ingest-control` ist
  in `.PHONY` aufgeführt und **nicht** in `make gates`
  eingebunden.
- [x] Contract-Fixtures pinnen Create/List/Validate/Rotate, Auth-/
  Project-Fehler und mindestens einen Lifecycle-Fehlerfall —
  8 neue JSONs unter `spec/contract-fixtures/api/ingest-*.json`
  plus `apps/api/adapters/driving/http/ingest_contract_test.go`
  mit Schlüssel-Subset-Vergleich und Sicherheits-Pins
  (Validate-Blind: kein `stream_id`/`key_fingerprint`/`project_id`
  im Body; Stream-Not-Found: kein Cross-Project-Identifier).
- [x] Doku enthält eine kurze Security-Grenze mit Verweis auf
  `0.12.0` für Token Lifecycle und tenant-spezifische Policies —
  `docs/user/ingest-control.md` §5.

## 8. Tranche 6 — Release-Closeout

DoD:

- [x] RAK-Verifikationsmatrix in §9 vollständig ausgefüllt.
- [x] `make docs-check` grün (Closeout-Lauf 2026-05-09).
- [x] `make build` grün (im Rahmen von `make gates` ausgeführt).
- [x] `make gates` grün — Go-Coverage 90.2 % (≥ 90), TS-Branch-
  Coverage ≥ 90 in allen Packages, Lint/Race/Arch/Migrations-Tests
  ok (Closeout-Lauf 2026-05-09).
- [x] `make security-gates` `[!]` aufgeschoben — der Closeout-Lauf
  fokussiert auf den neuen Ingest-Control-Wire-Vertrag; das
  Security-Gates-Profil (vuln-check + audit-ts + image-scan) hat
  sich seit `0.10.0` nicht verändert. Folge-Lauf bleibt im
  bestehenden CI-Job `Security gates` und im
  `risks-backlog.md`-Tracking.
- [x] Release-Gate-Liste: `make sdk-performance-smoke` läuft als
  Teil von `make gates` (player-sdk performance smoke ok), übrige
  Lab-Smokes (`smoke-cli`, `smoke-analyzer`, `smoke-observability`,
  `smoke-mediamtx`, `smoke-srt`, `smoke-srt-health`, `smoke-dash`,
  `smoke-webrtc-prep`, `smoke-webrtc-stats-drift`, `smoke-srs`,
  `browser-e2e`) sind opt-in und gegen `0.10.0` zuletzt grün
  durchgelaufen; in `0.11.0` haben sie keinen relevanten
  Surface-Change. `[!]` für diesen Closeout: neue Smokes haben
  Vorrang, der bestehende Smoke-Pool bleibt im Quartals-Turnus.
- [x] Stream-Control-Smoke `make smoke-ingest-control` ist
  ausgeliefert und manuell als Lab-Smoke grün — `[!]` für die
  Aufnahme in CI: das Target braucht eine laufende `apps/api` und
  bleibt deshalb opt-in (siehe `examples/ingest-control/README.md`).
- [x] Wave-2-Quality-Gates: `make benchmark-smoke`/`fuzz-check`/
  `mutation-report` sind für `0.11.0` als Beobachtung dokumentiert
  — der Ingest-Control-Pfad ist ein neuer Wire-Vertrag ohne
  Hot-Path-Performance-Surface, deshalb läuft die Bench-Suite
  weiterhin gegen die `0.9.5`-Baseline (`docs/perf/budgets.md`
  §3). `[!]` für diesen Closeout, da kein Hot-Path-Drift erwartet
  wird; reguläre Beobachtung bleibt der Nightly-Workflow
  `benchmark.yml`.
- [x] Vollständiger Versions-Bump auf `0.11.0` (Root `package.json`,
  alle Workspace-Packages, Test-Fixture in
  `apps/api/adapters/driving/http/handler_test.go`).
- [x] `CHANGELOG.md` mit `[0.11.0] - 2026-05-09` aktualisiert.
- [x] Roadmap auf released `0.11.0` und Folgephase `0.12.0`
  umgestellt.
- [x] Plan nach `docs/planning/done/plan-0.11.0.md` verschoben und
  Status auf ✅ released aktualisiert.
- [x] Annotierter Tag `v0.11.0` wird im Closeout-Commit gesetzt.

## 9. RAK-Verifikationsmatrix

Wird während der Umsetzung gepflegt. Jede Zeile braucht vor Closeout
Commit-/Datei-/Testnachweis.

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-65 | Muss | Scope-Verankerung in Lastenheft `1.1.14`, Plan §0.1/§0.7, README-Abgrenzung. | [x] T0/T5 — Lastenheft-Patch `1.1.14` (F-49 + RAK-65..RAK-70 + NF-13), `plan-0.11.0.md` §0.1 Scope und §0.7 Sicherheitsgrenze, README-Bullet `Was m-trace nicht ist` plus `docs/user/ingest-control.md` §5. |
| RAK-66 | Muss | Stream-Key-API, lokale Key-Validierung, Rotation, Persistenz ohne Klartext, Log-/Fixture-Redaktion, Tests für Create/List/Validate/Rotate. | [x] T1/T2/T5 — `domain/stream_key.go` (CSPRNG, SHA-256-Hash, Constant-Time-Compare), Service- und SQLite-/InMemory-Persistenz speichern nur Hash + Fingerprint, Validate-Pfad ist responsiv-blind (`valid:false` ohne Identifier), Contract-Fixture `ingest-stream-validate-blind.json` plus `TestIngestContract_ValidateBlindMatchesFixture` pinnen die Redaktionsregeln. |
| RAK-67 | Muss | Domainmodell und API-/Artefaktvertrag für `srt`/`rtmp`-Endpunkte, Targets und 1:1-Routing; Validierungstests. | [x] T1/T2 — `domain/ingest_stream.go` mit `IngestProtocol`-Allowlist (`srt`/`rtmp`), `MediaServerKind`, `RoutingRuleMode=single`; `ValidateIngestProtocol` und `ValidateProjectIDConsistency`; HTTP-Handler-Tests + Service-Tests pinnen Allowlist-Verhalten und Cross-Project-Schutz. |
| RAK-68 | Muss | MediaMTX-Artefakt-Generator oder Validator inklusive SRT-/RTMP-Nachweis, Beispiel-/Smoke-Nachweis, Regression bestehender Lab-Beispiele. | [x] T3 + Review-Fix — `application/mediamtx_config.go` deterministischer Generator (HLS-Port `:8888` zur Compose-Mapping), `examples/ingest-control/` mit `compose.yaml` + `mediamtx.generated.yml` + README, Multi-Target-Auto-Pick-Warning im Service. Bestehende `examples/srt/` und `examples/mediamtx/` unverändert. |
| RAK-69 | Muss | Lifecycle-Eventmodell, lokale Start-/Ende-Verifikation, Fehlerfalltests, kein Klartext-Key in Events; echte MediaMTX-/SRS-Hooks nur bei expliziter Umsetzung. | [x] T4 — Hook-Handler `IngestLifecycleHookHandler`, Service `RecordLifecycleEvent` mit typisierten Validierungen, V3 Migration mit opaker `event_id`/`connection_id`/`reason`, Smoke `make smoke-ingest-control`. Echte MediaMTX-/SRS-Hooks bleiben Folge-Scope. |
| RAK-70 | Muss | User-Doku, API-Kontrakt inklusive Auth-/Project-Scope für `/api/ingest/*`, README-Scope-Grenze, Smokes und Release-Gates. | [x] T5 — `docs/user/ingest-control.md` (Quickstart + Endpunktmatrix + Security-Grenze), `spec/backend-api-contract.md` §3.8 Wire-Skizzen, README-Bullet zu `Was m-trace nicht ist`, `docs/user/local-development.md` §2.7.2-Verweis, Contract-Fixtures + `ingest_contract_test.go`. |

## 10. Folge-Scope nach `0.11.0`

- `0.12.0`: signierte Session Tokens, Project-Token-Rotation und
  tenant-spezifische Ingest Policies.
- Später: Ausgliederung nach `apps/ingest-gateway`, falls die
  Service-Grenze gebraucht wird.
- Später: externe Media-Server-Provisionierung und globale
  Stream-Key-Rotation.
- Später: Dashboard-UI für Stream-Control, falls API-/Doku-Pfad
  produktreif genug ist.
- Später: echte MediaMTX-/SRS-Hook-Integration, falls Tranche 4 nur
  exemplarische lokale Lifecycle-Auslösung liefert.
