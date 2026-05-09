# Implementation Plan — `0.11.0` (Ingest-Gateway / Stream Control)

> **Status**: ⬜ open — noch nicht aktiviert. Dieser Plan darf erst nach
> explizitem Move nach `docs/planning/in-progress/` umgesetzt werden.
> Vorgänger ist `0.10.0` (CMAF-Analyse / NF-13; Tag `v0.10.0`);
> Aktivierung erst nach dessen Release-Closeout.
>
> **Release-Typ**: Minor-Release mit Lastenheft-Patch `1.1.14`
> (Vorschlag), neuer RAK-Gruppe `RAK-65`..`RAK-70`,
> RAK-Verifikationsmatrix und Tag `v0.11.0`.
>
> **Ziel**: Die bisher als Kann geführten Ingest-Gateway-Funktionen
> `F-46`..`F-51` werden in einen umsetzbaren Produkt-Scope geschnitten:
> ein lokal betreibbarer Stream-Control-Pfad für Lab-Streams,
> Stream-Keys, Ingest-Endpunkte, Routing-Regeln, Webhook-
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
    dann nur Hash/Fingerprint plus Metadaten.
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
  - Webhook-Adapter kann Events exemplarisch empfangen oder in einem
    lokalen Smoke auslösen; produktive Zustellung an externe Systeme
    ist optionaler Folge-Scope.
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

### 0.3 Architekturentscheidung (vor Tranche 1 zu schließen)

Vor Implementierung muss eine von zwei Varianten verbindlich gewählt
und im Plan dokumentiert werden.

| Variante | Beschreibung | Vorteil | Risiko |
| -------- | ------------ | ------- | ------ |
| A | eigenes `apps/ingest-gateway` nach Lastenheft §7.5.4 | klare Service-Grenze für spätere Control-Plane | neue App, Dockerfile, Port, CI- und Doku-Aufwand für lokalen Scope |
| B | Ingest-Control als Modul in `apps/api` | nutzt vorhandene HTTP-, SQLite-, Metrik- und Test-Infrastruktur | Name `ingest-gateway` bleibt zunächst konzeptionell, spätere Ausgliederung braucht Migration |

Vorschlag für `0.11.0`: **Variante B**, sofern Tranche 0 keine
zwingende Service-Grenze findet. Begründung: Der Release ist bewusst
lokal/lab-nah. Ein zusätzlicher Prozess wäre für `F-46`..`F-51` noch
mehr Betriebsoberfläche als Produktnutzen. Die Domain soll aber so
geschnitten werden, dass eine spätere Ausgliederung in
`apps/ingest-gateway` möglich bleibt.

### 0.4 Persistenzentscheidung (vor Tranche 1 zu schließen)

Standardvorschlag: SQLite über die bestehende API-Persistenz, falls
Variante B gewählt wird. Reine Konfigurationsartefakte sind nur
zulässig, wenn folgende Punkte trotzdem stabil erfüllt werden:

- Stream-IDs bleiben über API-Restarts reproduzierbar.
- Rotierte Keys können alte Fingerprints deaktivieren.
- Contract-Tests können Listen, Lesen, Rotieren und Validieren ohne
  Testreihenfolge-Abhängigkeit prüfen.
- Doku macht klar, ob Daten in SQLite oder nur in generierten
  Artefakten leben.

### 0.5 Lastenheft-Patch `1.1.14` (Vorschlag)

Der Patch ergänzt `spec/lastenheft.md` um RAK-65..RAK-70 und hebt
`F-46`..`F-51` für den begrenzten `0.11.0`-Lab-Control-Scope von
Kann auf Release-Muss. Die Kann-Historie bleibt auditierbar; verbindlich
ist die neue RAK-Gruppe.

| RAK | Priorität | Inhalt |
| --- | --------- | ------ |
| RAK-65 | Muss | Ingest-Control-Scope ist normativ begrenzt: lokale/lab-nahe Stream-Verwaltung, keine Multi-Tenant-Control-Plane, keine produktive Secret-Verwaltung, keine externe Media-Server-Provisionierung. |
| RAK-66 | Muss | Stream-Key-Verwaltung: Streams können angelegt, gelistet und rotiert werden; Klartext-Keys erscheinen nur bei Anlage/Rotation, nicht in Logs, Fixtures oder Persistenz. |
| RAK-67 | Muss | Ingest-Endpunkt- und Routing-Modell: `srt`/`rtmp`-Endpunkte, Stream-Ziele und einfache 1:1-Routing-Regeln sind validiert, dokumentiert und per API/Artefakt stabil beschreibbar. |
| RAK-68 | Muss | Media-Server-Artefakte: MediaMTX-nahe Konfigurationen für den Lab-Scope können generiert oder validiert werden; bestehende Multi-Protocol-Lab-Beispiele und Smokes bleiben grün. |
| RAK-69 | Muss | Lifecycle-Webhooks: `stream_started` und `stream_ended` besitzen ein stabiles Eventmodell und werden lokal reproduzierbar empfangen oder exemplarisch ausgelöst; Events enthalten keine Klartext-Keys. |
| RAK-70 | Muss | Doku, API-/Contract-Tests und Release-Smokes beschreiben den lokalen Stream-Control-Workflow, die Sicherheitsgrenzen und den Unterschied zu Auth-/Tenant-Folge-Scope `0.12.0`. |

### 0.6 Öffentliche API und Modell-Skizze

Die finale Pfadwahl hängt von der Architekturentscheidung ab. Für
Variante B werden die Lastenheft-Pfade unter `apps/api` verwendet:

| Methode | Pfad | Zweck |
| ------- | ---- | ----- |
| `POST` | `/api/ingest/streams` | Ingest-Stream anlegen; gibt Stream-Metadaten plus Klartext-Key genau einmal zurück |
| `GET` | `/api/ingest/streams` | Streams listen; ohne Klartext-Key |
| `GET` | `/api/ingest/streams/{id}` | Stream-Details, Endpunkte und Routing-Regel lesen; ohne Klartext-Key |
| `POST` | `/api/ingest/streams/{id}/rotate-key` | Key rotieren; gibt neuen Klartext-Key genau einmal zurück |
| `POST` | `/api/ingest/hooks/stream-started` | lokalen Start-Webhook empfangen oder Smoke-Event einspeisen |
| `POST` | `/api/ingest/hooks/stream-ended` | lokalen Ende-Webhook empfangen oder Smoke-Event einspeisen |
| `GET` | `/api/ingest/media-server-config` | generiertes/validiertes MediaMTX-Artefakt abrufen oder Diagnose liefern |

Pflicht-Domainobjekte:

- `IngestStream`: `id`, `project_id`, `display_name`, `protocol`,
  `endpoint_id`, `target_id`, `routing_rule_id`, `status`,
  `created_at`, `updated_at`.
- `StreamKey`: `stream_id`, `fingerprint`, `created_at`,
  `rotated_at?`, `disabled_at?`; kein Klartextfeld in Persistenz.
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
- Rotierte oder deaktivierte Key-Fingerprints akzeptiert kein
  Validierungspfad mehr.
- Fehlercodes sind stabil und werden in Contract-Tests gepinnt, z. B.
  `ingest_stream_duplicate`, `ingest_protocol_unsupported`,
  `ingest_endpoint_missing`, `ingest_route_invalid`,
  `ingest_key_not_found`.

### 0.7 Security- und Logging-Grenzen

`0.11.0` darf Security nicht vortäuschen. Deshalb gilt:

- Kein Auth-Versprechen jenseits bestehender API-Mechanismen.
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
| 0 | Plan-Aktivierung, Lastenheft-Patch `1.1.14`, RAK-Gruppe, Architektur- und Persistenzentscheidung | ⬜ |
| 1 | Stream-Key-, Ingest-Endpunkt- und Routing-Domainmodell | ⬜ |
| 2 | API-/Persistenzpfad für Streams, Listing und Key-Rotation | ⬜ |
| 3 | MediaMTX-Artefakte und SRT-/RTMP-Lab-Konfiguration | ⬜ |
| 4 | Lifecycle-Webhooks und lokale Lab-Verifikation | ⬜ |
| 5 | Doku, Contract-Tests, Smokes und README-Abgrenzung | ⬜ |
| 6 | Gates, RAK-Verifikationsmatrix, Versions-Bump, Closeout und Tag | ⬜ |

---

## 2. Tranche 0 — Aktivierung, Patch und Entscheidungen

Ziel: Der Release-Scope wird vor Implementierung normativ und
architektonisch geschlossen.

DoD:

- [ ] Plan von `docs/planning/open/plan-0.11.0.md` nach
  `docs/planning/in-progress/plan-0.11.0.md` verschoben.
- [ ] `git status --short` vor erster Änderung dokumentiert.
- [ ] `spec/lastenheft.md` Header auf `1.1.14` erhöht.
- [ ] RAK-65..RAK-70 im Lastenheft ergänzt.
- [ ] `F-46`..`F-51` im Lastenheft für den `0.11.0`-Scope
  nachvollziehbar von Kann-Historie auf Release-Muss abgebildet.
- [ ] Patch-Log in `docs/planning/done/plan-0.1.0.md` um
  `Patch 1.1.14` ergänzt.
- [ ] Architekturentscheidung dokumentiert: Variante A
  `apps/ingest-gateway` oder Variante B `apps/api`-Modul.
- [ ] Persistenzentscheidung dokumentiert: SQLite oder rein
  artefaktbasiert, inklusive Migrations-/Testfolge.
- [ ] Roadmap-Status und Release-Übersicht auf `0.11.0` als aktive
  Folgephase umgestellt.
- [ ] Risiko-/Folge-Scope-Liste aktualisiert: Auth/Tenant/Policy nach
  `0.12.0`, externe Provisionierung offen.

## 3. Tranche 1 — Domainmodell und Validierung

Ziel: Stream Control ist als reines Domainmodell testbar, bevor HTTP,
Storage oder Media-Server-Artefakte angebunden werden.

DoD:

- [ ] Domainobjekte `IngestStream`, `StreamKey`, `IngestEndpoint`,
  `RoutingRule`, `MediaServerTarget` und `StreamLifecycleEvent`
  definiert.
- [ ] Protocol-Enum ist auf `srt` und `rtmp` begrenzt; unbekannte
  Werte liefern stabilen Fehlercode.
- [ ] Stream-Key-Erzeugung erzeugt ausreichend zufällige lokale Keys
  und berechnet Fingerprint/Hash getrennt vom Klartext.
- [ ] Validierungsregeln decken ungültige Keys, doppelte aktive
  Stream-Namen, fehlende Endpunkte, fehlende Targets und deaktivierte
  Routing-Regeln ab.
- [ ] Domain-Tests laufen ohne HTTP-Server, Docker oder MediaMTX.
- [ ] Kein Domain-Test speichert oder snapshotet echte Klartext-Keys.

## 4. Tranche 2 — API, Persistenz und Key-Rotation

Ziel: Der lokale Stream-Katalog ist über stabile API-Verträge nutzbar
und über API-Restarts hinweg reproduzierbar, sofern SQLite gewählt
wurde.

DoD:

- [ ] `POST /api/ingest/streams` legt Stream, Endpoint-Bezug,
  Routing-Regel und initialen Stream-Key an.
- [ ] `GET /api/ingest/streams` listet Streams ohne Klartext-Key.
- [ ] `GET /api/ingest/streams/{id}` liefert Details ohne
  Klartext-Key.
- [ ] `POST /api/ingest/streams/{id}/rotate-key` deaktiviert den
  alten Fingerprint und gibt den neuen Klartext-Key genau einmal
  zurück.
- [ ] Persistenz- oder Artefaktpfad hat Contract-Tests für Create,
  List, Detail, Rotation, Duplicate, Missing und Invalid.
- [ ] HTTP-Fehlercodes sind stabil und im API-Kontrakt dokumentiert.
- [ ] Logs und Request-Metriken enthalten keine Klartext-Keys.
- [ ] Falls SQLite genutzt wird: Migration ist versioniert, Drift-Check
  grün und In-Memory-/SQLite-Adapter teilen Contract-Tests.

## 5. Tranche 3 — Routing und Media-Server-Artefakte

Ziel: Aus den Stream-Control-Daten entstehen überprüfbare
MediaMTX-nahe Lab-Artefakte, ohne laufende Fremdserver automatisch zu
verändern.

DoD:

- [ ] Routing-Regeln sind als stabile JSON-Konfiguration beschreibbar.
- [ ] MediaMTX-nahe Konfigurationsartefakte können generiert oder
  validiert werden.
- [ ] SRT- und RTMP-Beispiele enthalten klare Endpoint-/URL-Templates
  und trennen Ingest-URL, Playback-/HLS-URL und Control-API-URL.
- [ ] Bestehende `examples/mediamtx`, `examples/srt` und
  `examples/srs` bleiben unverändert nutzbar oder werden nur additiv
  dokumentiert.
- [ ] Falls ein neues Beispiel entsteht, folgt es
  `examples/README.md`: eigener Project-Name, eigene README-
  Mindeststruktur, opt-in Smoke.
- [ ] Artefakte enthalten nur Beispiel- oder redigierte Stream-Keys.

## 6. Tranche 4 — Lifecycle-Webhooks und Lab-Verifikation

Ziel: Stream-Start und Stream-Ende sind als lokale Ereignisse
modelliert und reproduzierbar verifizierbar.

DoD:

- [ ] Eventmodell für `stream_started` und `stream_ended` ist
  dokumentiert.
- [ ] Webhook-Endpoint akzeptiert valide Start-/Ende-Events und weist
  unbekannte Streams, ungültige Eventtypen und malformed Payloads
  stabil ab.
- [ ] Lifecycle-Events enthalten `stream_id`, `observed_at`, `source`
  und optional `connection_id`/`reason`; sie enthalten keinen
  Klartext-Key.
- [ ] Lokaler Smoke verifiziert mindestens einen Start-/Ende-Pfad
  reproduzierbar.
- [ ] Falls echte MediaMTX-Hooks in `0.11.0` nicht angebunden werden,
  ist die Entscheidung als `[!]`-Folge-Scope mit RAK-69-Nachweis
  dokumentiert: Eventmodell + exemplarische lokale Auslösung genügen.

## 7. Tranche 5 — Doku, Contracts und Smokes

Ziel: Nutzer können den lokalen Stream-Control-Pfad nachvollziehen,
ohne ihn mit produktiver Control-Plane, Tenant-Policy oder Auth zu
verwechseln.

DoD:

- [ ] User-Doku beschreibt den lokalen Stream-Control-Workflow:
  Stream anlegen, Key verwenden, Route prüfen, MediaMTX-Artefakt
  ansehen, Key rotieren.
- [ ] API-Kontrakt dokumentiert Endpunkte, Erfolgsantworten,
  Fehlercodes und Redaktionsregeln für Secrets.
- [ ] README grenzt `0.11.0` gegen Control-Plane, Multi-Tenant-
  Betrieb und Secret-Management ab.
- [ ] `docs/user/local-development.md` oder ein neues User-Dokument
  verlinkt den Smoke- und Beispielpfad.
- [ ] Relevante Smokes sind im Makefile dokumentiert; Lab-Smokes
  bleiben opt-in und werden nicht ungeprüft in `make gates` gezogen.
- [ ] Contract-Fixtures oder API-Snapshots pinnen Create/List/Rotate
  und mindestens einen Webhook-Fehlerfall.
- [ ] Doku enthält eine kurze Security-Grenze mit Verweis auf
  `0.12.0` für Token Lifecycle und tenant-spezifische Policies.

## 8. Tranche 6 — Release-Closeout

DoD:

- [ ] RAK-Verifikationsmatrix in §9 vollständig ausgefüllt.
- [ ] `make docs-check` grün.
- [ ] `make build` grün.
- [ ] `make gates` grün.
- [ ] `make security-gates` grün oder CI-Job `Security gates` grün
  dokumentiert.
- [ ] Relevante opt-in Lab-Smokes dokumentiert; mindestens der neue
  Stream-Control-Smoke grün oder begründet `[!]`.
- [ ] Wave-2-Quality-Gates vor dem Tag geprüft.
- [ ] Vollständiger Versions-Bump auf `0.11.0`.
- [ ] `CHANGELOG.md` mit `[0.11.0] - YYYY-MM-DD` aktualisiert.
- [ ] Roadmap auf released `0.11.0` und Folgephase `0.12.0`
  umgestellt.
- [ ] Plan nach `docs/planning/done/plan-0.11.0.md` verschoben und
  Status auf ✅ released aktualisiert.
- [ ] Annotierter Tag `v0.11.0` erstellt.

## 9. RAK-Verifikationsmatrix

Wird während der Umsetzung gepflegt. Jede Zeile braucht vor Closeout
Commit-/Datei-/Testnachweis.

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-65 | Muss | Scope-Verankerung in Lastenheft `1.1.14`, Plan §0.1/§0.7, README-Abgrenzung. | [ ] |
| RAK-66 | Muss | Stream-Key-API, Rotation, Persistenz ohne Klartext, Log-/Fixture-Redaktion, Tests für Create/List/Rotate. | [ ] |
| RAK-67 | Muss | Domainmodell und API-/Artefaktvertrag für `srt`/`rtmp`-Endpunkte, Targets und 1:1-Routing; Validierungstests. | [ ] |
| RAK-68 | Muss | MediaMTX-Artefakt-Generator oder Validator, Beispiel-/Smoke-Nachweis, Regression bestehender Lab-Beispiele. | [ ] |
| RAK-69 | Muss | Webhook-Eventmodell, lokale Start-/Ende-Verifikation, Fehlerfalltests, kein Klartext-Key in Events. | [ ] |
| RAK-70 | Muss | User-Doku, API-Kontrakt, README-Scope-Grenze, Smokes und Release-Gates. | [ ] |

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
  exemplarische lokale Webhook-Auslösung liefert.
