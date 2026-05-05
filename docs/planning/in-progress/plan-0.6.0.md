# Implementation Plan — `0.6.0` (SRT Health View)

> **Status**: 🟡 aktiv. `0.5.0` (Multi-Protocol Lab) ist released
> (Tag `v0.5.0` auf `a56dc0b`, CI-Run 25364250989 grün); Vorgängerplan
> ist nach [`../done/plan-0.5.0.md`](../done/plan-0.5.0.md) archiviert.
> Tranche 0 (Vorgänger-Gate) ist abgeschlossen; Tranche 1 (SRT-
> Metrikquelle und Binding-Entscheidung) ist die nächste Arbeitsstufe.
>
> **Bezug**: [Lastenheft `1.1.9`](../../../spec/lastenheft.md) §4.3
> (SRT als späterer starker Hebel), §7.8 (lokales Streaming-Lab), §7.9
> (Observability), §7.10 (Telemetry/Cardinality), §8.3 NF-11
> (SRT-Ingest-Metriken), §13.8 (RAK-41..RAK-46);
> [Roadmap](../in-progress/roadmap.md) §3/§4;
> [Architektur](../../../spec/architecture.md);
> [Telemetry-Model](../../../spec/telemetry-model.md);
> [API-Kontrakt](../../../spec/backend-api-contract.md);
> [Risiken-Backlog](../open/risks-backlog.md) R-2.
>
> **Vorgänger-Gate (Soll zum `0.6.0`-Start)**:
>
> - `0.5.0` ist released, `plan-0.5.0.md` ist nach
>   `docs/planning/done/` verschoben und RAK-36..RAK-40 sind
>   verifiziert.
> - `examples/srt/` enthält ein lokales, smoke-testbares SRT-Beispiel.
> - `make smoke-srt` ist grün und validiert nicht nur offene Ports,
>   sondern eine echte SRT-Publish- und Ausspielungsstrecke.
> - R-2 ist vor Code-Start neu bewertet: Die Metrikquelle darf das
>   `apps/api`-Runtime-Image nicht implizit von `distroless/static` auf
>   eine CGO-/glibc-Variante zwingen, ohne dass eine ADR die Konsequenz
>   dokumentiert.
> - Roadmap §1/§3 ist auf `0.6.0` als aktive Phase umgestellt.
>
> **Nachfolger**: offen.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog
[`plan-0.1.0.md`](../done/plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz oder offene
  Entscheidung.
- 🟡 in Arbeit.

Neue Lastenheft-Patches während `0.6.0` landen weiterhin zentral in
`plan-0.1.0.md` Tranche 0c, weil sie projektweit gelten.

Scope-Grenze: `0.6.0` liefert eine **lokale SRT-Health-Ansicht** mit
erfassten oder importierten SRT-Verbindungsmetriken, API-/Storage-
Vertrag, Dashboard-Sicht und Dokumentation. Es liefert keine produktive
Ingest-Verwaltung, keine Multi-Tenant-SRT-Orchestrierung, keine
Kubernetes-Deployment-Story und keine WebRTC-Metriken.

### 0.1 Vorab-Entscheidungen für den Scope-Cut

Diese Entscheidungen verhindern, dass die SRT-Health-Phase zu einer
vollständigen Media-Server-Verwaltung wächst.

| Thema | Entscheidung für `0.6.0` | Nicht-Ziel |
| ----- | ------------------------ | ---------- |
| Metrikquelle | Eine konkrete, lokal reproduzierbare Quelle wird in Tranche 1 festgelegt: MediaMTX-/SRT-Stats-API, Sidecar-Exporter oder kontrollierter Log-/CLI-Import. | Direkte libsrt-CGO-Bindings in `apps/api`, solange R-2 nicht per ADR entschieden ist. |
| Datenmodell | SRT-Health wird als getrenntes Verbindungs-/Ingest-Signal modelliert und OTel-kompatibel beschrieben. | SRT-Werte als Player-Playback-Events vortäuschen. |
| Prometheus | Nur bounded Aggregate sind erlaubt. Per-Verbindung-/Per-Stream-Details gehen in SQLite und/oder OTel-Spans/Events, nicht in Prometheus-Labels. Rohmetriken aus MediaMTX oder anderen Quellen werden **nicht** in den Projekt-Prometheus gescraped; sie werden über Adapter/Sidecar gelesen und kontrolliert normalisiert. | `stream_id`, `session_id`, URL, Token, IP, Connection-ID oder Source-Labels wie `id`, `path`, `remoteAddr`, `state` als Prometheus-Label. |
| Dashboard | Eine eigene SRT-Health-Ansicht zeigt aktuelle Verbindung, Freshness, Warnzustände und die letzten Samples als kurzer Verlauf. Ein Snapshot-only-Abschluss ist für `0.6.0` nicht ausreichend. | Vollständige Media-Server-Konsole oder Stream-Key-Verwaltung. |
| Lab | `0.6.0` baut auf `examples/srt/` aus `0.5.0` auf und härtet es für Health-Nachweise. | Neues paralleles SRT-Lab ohne Bezug zum bestehenden Beispiel. |
| Fehlerbilder | Dokumentation erklärt typische SRT-Probleme anhand der gelieferten Metriken. | Allgemeines SRT-Lehrbuch oder produktive Netzwerk-Tuning-Anleitung. |
| Erweiterte SRT-Signale | `0.6.0` priorisiert die RAK-43-Pflichtwerte RTT, Packet Loss, Retransmissions und verfügbare Bandbreite. Send-/Receive-Buffer, Verbindungsstabilität, Link Health und Failover-Zustände aus Lastenheft §4.3 werden als Deferred-Liste geführt, sofern sie nicht ohne Zusatzrisiko aus der gewählten Quelle mitfallen. | RAK-43 durch nicht geforderte Zusatzwerte erweitern und damit den Release-Scope aufblasen. |

### 0.2 Risiko-Triage aus dem Backlog

| Risiko | Entscheidung für `0.6.0` | Trigger / Nachweis |
| ------ | ------------------------ | ------------------ |
| R-2 CGO/SRT-Bindings | **Aufgelöst durch Sub-1.3**: Option 1 (MediaMTX-API über HTTP) gewählt. `apps/api` bleibt CGO-frei; `distroless-static`-Pattern unangefasst. Folge-ADR „SRT-Binding-Stack" entfällt. | Sub-1.2 Probe liefert vier Pflichtwerte; §2.4/§2.5 dokumentiert; risks-backlog R-2 nach §1.2 verschoben. |
| R-5 Time-Skew-Persistenz | Nicht Teil der SRT-Health-Pflicht, außer SRT-Metriken brauchen Client-/Server-Zeitvergleich im Dashboard. | Wenn Health-Events Zeitversatz bewerten müssen, wird eine additive Sub-Tranche ergänzt oder R-5 separat aktiviert. |
| R-7 Session-List-N+1 | Beobachten. SRT-Health darf Session-Listen nicht durch zusätzliche N+1-Reads verschlechtern. | Wenn SRT-Health in `GET /api/stream-sessions` eingebettet wird und p95 >= 200 ms reproduzierbar wird, Bulk-Read-Port vor Dashboard-Integration liefern. |
| R-10 Sampling-Vollständigkeit | Nicht Teil von `0.6.0`, solange SRT-Metriken unabhängig von Player-Sampling laufen. | Aktivieren, falls Health-Ansicht Player-Event-Vollständigkeit als Diagnosevoraussetzung behauptet. |

### 0.3 Sequenzierung und harte Gates

`0.6.0` hat eine höhere Fehlentscheidungsgefahr als die vorherigen
Lab-Tranchen, weil die Metrikquelle Runtime, Cardinality und Dashboard-
Vertrag gleichzeitig beeinflusst. Daher gelten diese Reihenfolgen:

1. Tranche 1 blockiert jede produktive Code-Integration, die eine
   SRT-Metrikquelle in `apps/api` oder Dashboard verdrahtet. Projekt-
   Prometheus-Scraping von Source-Rohmetriken bleibt generell außerhalb
   des `0.6.0`-Scopes.
2. Tranche 2 darf nur gegen die in Tranche 1 gewählte Quelle härten;
   parallele zweite SRT-Testpfade sind nicht Teil des Plans.
3. Tranche 3 muss Telemetry-Model, API-Kontrakt und Storage-Shape
   festlegen, bevor Tranche 4 HTTP-Handler oder Dashboard-Client-Code
   darauf aufbauen.
4. Tranche 5 darf keine eigene Health-Bewertung implementieren; die
   Bewertung kommt aus Tranche 4 oder ist als reine UI-Formatierung
   dokumentiert.
5. Tranche 7 darf RAK-43 erst schließen, wenn die vier Pflichtwerte aus
   derselben End-to-End-Quelle in API und Dashboard sichtbar sind.

---

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Vorgänger-Gate und Scope-Festlegung | ✅ |
| 1 | SRT-Metrikquelle und Binding-Entscheidung (R-2, RAK-42) | ✅ Quellen-Entscheidung (Sub-1.1–1.4); zwei DoD-Items in Tranche 2/3 verlagert (`required_bandwidth_bps`, formaler API-Pull-Vertrag) |
| 2 | SRT-Testsetup zum Health-Lab härten (RAK-41) | ✅ |
| 3 | SRT-Health-Datenmodell, Storage und OTel-Vertrag (RAK-42, RAK-46) | ✅ (Sub-3.1..3.7 alle ✅) |
| 4 | API-Read-Pfad und Health-Bewertung (RAK-43) | ✅ |
| 5 | Dashboard-SRT-Health-Ansicht (RAK-43, RAK-44) | 🟡 (UI ✅, Browser-E2E → Tranche 7) |
| 6 | Fehlerbild-Dokumentation und Operator-Guide (RAK-45) | ✅ |
| 7 | Smokes, Gates und Release-Closeout (RAK-41..RAK-46) | ⬜ |

---

## 1a. Tranche 0 — Vorgänger-Gate und Scope-Festlegung

Bezug: Roadmap §1/§3/§4; `plan-0.5.0.md` §4/§7; Risiken-Backlog R-2.

Ziel: Vor Implementierung ist klar, dass `0.6.0` auf einem
funktionierenden SRT-Lab aufsetzt und nur SRT-Health liefert, nicht
Ingest-Orchestrierung oder produktive SRT-Betriebsverwaltung.

DoD:

- [x] `plan-0.5.0.md` ist vollständig abgeschlossen, nach
  `docs/planning/done/` verschoben und mit finalen Commit-Hashes
  versehen (T6c-Closeout `486324e`).
- [x] Annotierter Release-Tag `v0.5.0` existiert und GitHub Actions
  `Build` ist für den Release-Commit grün (`v0.5.0` auf `a56dc0b`,
  CI-Run 25364250989).
- [x] `examples/srt/README.md` beschreibt den finalen `0.5.0`-SRT-Pfad
  inklusive Start, Verifikation, Stop/Reset, Troubleshooting und
  bekannter Grenzen (Lieferstand aus `plan-0.5.0` Tranche 3 RAK-37).
- [x] `make smoke-srt` ist auf dem `0.5.0`-Release-Stand grün und als
  Baseline für `0.6.0` dokumentiert (`scripts/smoke-srt.sh`,
  `examples/srt/README.md` „Verifikation").
- [x] Roadmap §1.2 und §3 markieren `0.6.0` als aktive Phase und
  verweisen auf diesen Plan (Tranche 0 dieses Plans).
- [x] R-2 ist für den Startzustand präzisiert: bekannte Metrikquellen,
  Runtime-Auswirkungen und offene ADR-Frage sind im Risiken-Backlog
  aktuell ([`risks-backlog.md`](../open/risks-backlog.md) R-2
  verweist auf §2-Optionentabelle).
- [x] Scope-Grenze ist in README/User-Doku sichtbar: `0.6.0` liefert
  SRT-Health, nicht Stream-Key-Verwaltung, SRT-Auth, Multi-Publisher-
  Routing oder Kubernetes (`README.md` Roadmap-Block + §0
  Scope-Grenze dieses Plans).

---

## 2. Tranche 1 — SRT-Metrikquelle und Binding-Entscheidung

Bezug: RAK-42, RAK-46; Lastenheft §4.3; Roadmap §4
SRT-Binding-Stack; Risiken-Backlog R-2; ADR-0001 §8.

Ziel: Vor produktivem Code ist entschieden, **woher** die SRT-
Verbindungsmetriken kommen und welche Runtime-Konsequenzen das hat.
Diese Tranche darf Doku-/ADR-lastig sein; sie verhindert eine
versehentliche CGO-Entscheidung im Implementierungsdiff.

Zu bewertende Quellen:

| Option | Idee | Bewertungskriterium |
| ------ | ---- | ------------------- |
| MediaMTX-/Server-API | SRT-Verbindungsdaten vom lokalen Media-Server lesen. | Bevorzugt, wenn RTT/Loss/Retransmissions/verfügbare Bandbreite vollständig und stabil verfügbar sind. |
| Sidecar-Exporter | Separater Container sammelt SRT-Stats und liefert sie über einen expliziten API-Import- oder API-Pull-Vertrag an `apps/api`; OTLP ist nur Zusatzexport. | Bevorzugt, wenn `apps/api` CGO-frei bleiben soll und die Quelle trotzdem vollständig ist. Reines OTLP zu Collector/Tempo erfüllt den Dashboard-Read-Pfad nicht. |
| Log-/CLI-Import | Lab-Smoke oder Sidecar normalisiert bekannte SRT-Tool-Ausgabe. | Nur akzeptabel, wenn deterministisch testbar und nicht fragil gegen lokalisierte Logtexte. |
| Direktes libsrt-Binding | `apps/api` liest SRT-Stats direkt über Binding. | Nur mit ADR und bewusst akzeptierter Runtime-/Image-Konsequenz. |

Harte Auswahlkriterien:

| Kriterium | Muss erfüllt sein |
| --------- | ----------------- |
| Vollständigkeit | RTT, Packet Loss, Retransmissions und verfügbare Bandbreite sind alle verfügbar und semantisch erklärbar. |
| Reproduzierbarkeit | Fixtures und Smoke laufen ohne Internet und ohne manuelle SRT-Tools auf dem Host. |
| Runtime-Grenze | `apps/api` bleibt CGO-frei oder die Runtime-Änderung ist per ADR akzeptiert. |
| Cardinality | Source-Rohmetriken werden nicht vom Projekt-Prometheus gescraped; nur m-trace-normalisierte bounded Aggregate dürfen exportiert werden. |
| Freshness | Quelle liefert `source_observed_at` oder ein äquivalentes quellen-nahes Freshness-Signal wie Sample-Window, Generation-ID oder monotone Source-Sequenz; Importzeit allein darf Freshness nicht beweisen. |
| Failure-Mode | Quelle hat unterscheidbare Fehler für "nicht erreichbar", "keine Verbindung" und "unvollständige Rohdaten". |
| Probe-Fähigkeit | Ein minimaler Source-Probe kann eine Rohantwort gegen Fixture/Parser prüfen, ohne `apps/api` oder Dashboard zu starten. |

### 2.1 Evaluations-Stand vor Probe (Sub-Tranche 1.1)

Bewertung der vier Optionen gegen die sieben Kriterien anhand des
aktuellen Wissensstands (vor Lab-Probe). Legende: ✅ erfüllt /
🟡 erfüllbar, aber Aufwand · ❓ unbekannt, Probe nötig · ❌ nicht
erfüllbar.

| Kriterium | (1) MediaMTX-API | (2) Sidecar-Exporter | (3) Log-/CLI-Import | (4) libsrt-Binding |
| --------- | :--------------: | :------------------: | :-----------------: | :----------------: |
| Vollständigkeit | ❓ | ✅ | 🟡 | ✅ |
| Reproduzierbarkeit | ✅ | 🟡 | ❌ | 🟡 |
| Runtime-Grenze | ✅ | ✅ | ✅ | ❌ |
| Cardinality | ✅ | ✅ | ✅ | ✅ |
| Freshness | 🟡 | ✅ | 🟡 | ✅ |
| Failure-Mode | ✅ | ✅ | 🟡 | ✅ |
| Probe-Fähigkeit | ✅ | 🟡 | ✅ | 🟡 |

Begründungen:

- **(1) MediaMTX-API** — Reproduzierbar (lokales Compose, Standard-
  HTTP-API auf `:9997`), Runtime-Grenze ist trivial (HTTP-Client in
  Go), Cardinality kontrollierbar (Adapter normalisiert, Projekt-
  Prometheus scrapt nichts). Probe ist `curl` plus Auth-Token.
  **Offene Hauptfrage**: liefert `/v3/srtconns/list` die vier RAK-43-
  Pflichtwerte (RTT, Packet Loss, Retransmissions, verfügbare
  Bandbreite) als first-class Felder oder nur Verbindungs-Metadaten
  plus Bytes-Counter? Das entscheidet Sub-Tranche 1.2 (Lab-Probe).
  Freshness: kein expliziter Source-Sample-Timestamp im API-Schema
  bekannt — Adapter muss `collected_at` setzen plus optional ein
  monotones Sample-Window aus `created`/`bytesReceived`-Δ ableiten.
- **(2) Sidecar-Exporter** — Vollständigkeit ist Implementierungs-
  entscheidung (eigener Container kann libsrt-Stats voll nutzen);
  apps/api bleibt CGO-frei, das CGO ist im Sidecar isoliert. Freshness
  ist Sidecar-kontrolliert (eigener `source_observed_at`). Reproduzier-
  barkeit/Probe-Aufwand höher als bei (1), weil das Sidecar-Image erst
  gebaut/gepinnt werden muss. **Bevorzugter Fallback**, falls (1) die
  vier Pflichtwerte nicht trägt.
- **(3) Log-/CLI-Import** — Plan §2 Optionentabelle qualifiziert das
  selbst: „nur akzeptabel, wenn deterministisch testbar und nicht
  fragil gegen lokalisierte Logtexte". MediaMTX-Logs sind text-
  formattiert und versionsabhängig; `srt-live-transmit --stats` hat
  CSV-Format, aber externe Tool-Abhängigkeit auf dem Lab-Host.
  Reproduzierbarkeit reißt damit. **Effektives Plan-Veto** — keine
  ernsthafte Option für `0.6.0`.
- **(4) libsrt-Binding** — CGO-Pflicht reißt distroless-static
  (R-2). Voll erfüllt sonst alle Kriterien. **Notausgang** mit
  accepted ADR „SRT-Binding-Stack" und R-2-Update auf konkrete
  Runtime-Konsequenz.

### 2.2 Empfehlung als Probe-Default

**Bevorzugter Pfad**: Option 1 (MediaMTX-API). Bestätigt durch
Reproduzierbarkeit, Runtime-Grenze, Cardinality, Probe-Fähigkeit. Die
einzige offene Hard-Kriterium-Frage ist Vollständigkeit — RAK-43
verlangt RTT, Packet Loss, Retransmissions, verfügbare Bandbreite.

**Fallback-Reihenfolge** bei negativem Probe-Ergebnis aus Sub-1.2:
1. Option 1 + ergänzende MediaMTX-Konfig oder neuere MediaMTX-Version,
   falls die fehlenden Felder dort exponiert werden.
2. Option 2 (Sidecar-Exporter) — eigener Container liest libsrt direkt
   und liefert über expliziten HTTP-Pull-Vertrag an `apps/api`.
3. Option 4 (libsrt in `apps/api`) nur, falls 1+2 unmöglich; mit
   accepted ADR und R-2-Update auf gewähltes Runtime-Profil.

Option 3 (Log/CLI) bleibt aus dem Rennen (Plan-Veto).

### 2.3 Probe-Plan für Sub-Tranche 1.2

Ziel: anhand von zwei aufeinanderfolgenden API-Antworten aus dem
Lab nachweisen, ob Option 1 die vier RAK-43-Pflichtwerte trägt.

| Schritt | Aktion | Erwartung |
| ------- | ------ | --------- |
| P1 | `examples/srt/`-Stack starten (`docker compose -p mtrace-srt -f examples/srt/compose.yaml up -d --build`); FFmpeg-Loop publiziert auf `srt://:8890?streamid=publish:srt-test`. | MediaMTX-Container `Healthy`, FFmpeg-Container running. |
| P2 | MediaMTX-Auth temporär für Probe konfigurieren: `mediamtx.yml` `authInternalUsers` mit Lab-Token oder API-Auth-Override; **außerhalb** des smoke-srt-Pfads, damit die `0.5.0`-Baseline grün bleibt. | API auf `:9997` mit `Authorization`-Header erreichbar. |
| P3 | `curl -sS http://localhost:9997/v3/srtconns/list` während Publisher läuft; Response speichern. | JSON mit `items[]`-Array, mindestens ein Eintrag mit `state: "publish"` (oder analog). |
| P4 | Schema-Inspektion: enthält jedes `items[]`-Element ein RTT-Feld, ein Loss-Feld, ein Retransmissions-Feld, ein Bandbreiten-Feld? Felder-Namen und Einheiten dokumentieren. | Vier Pflichtwerte vorhanden ODER negativer Befund. |
| P5 | Zweiten Request nach 5 s. | `bytesReceived` höher; falls vorhanden, `pktRetransTotal`/`pktLossTotal` monoton oder dokumentierter Reset. |
| P6 | Response-Snapshot in `spec/contract-fixtures/srt/mediamtx-srtconns-list.json` speichern (anonymisiert: keine echten IPs als Lab-Daten). | Fixture für späteren Adapter-Parser. |

Ergebnis-Pfade aus Sub-1.2:

- **Positiv** (alle vier Pflichtwerte vorhanden): Sub-1.3 dokumentiert
  Option 1 als verbindliche Wahl im Plan; Adapter-Skizze + Einheiten/
  Counter-vs-Rate-Mapping; Sub-1.4 setzt R-2 auf „aufgelöst durch
  CGO-freie HTTP-Quelle".
- **Negativ** (mind. ein Pflichtwert fehlt): Sub-1.3 prüft Option 2
  (Sidecar) oder Lastenheft-Patch §4.3, dann ggf. Option 4 mit ADR.

### 2.4 Probe-Befund (Sub-Tranche 1.2, 2026-05-05)

Probe ausgeführt auf `examples/srt/`-Stack (separate Compose-Probe-
Variante mit auth-Override; Smoke-Baseline `examples/srt/mediamtx.yml`
unangefasst). MediaMTX `bluenviron/mediamtx:1` (1.x-Linie), API auf
`localhost:9998` → `:9997`-intern. Zwei Snapshots im Abstand von 5 s.

**Befund: positiv — alle vier RAK-43-Pflichtwerte vorhanden.**

| RAK-43-Wert | MediaMTX-Feld | Typ / Einheit | Probe-Beispielwert |
| ----------- | ------------- | ------------- | ------------------ |
| RTT | `msRTT` | Snapshot, Millisekunden | `0.365` ms |
| Packet Loss (counter) | `packetsReceivedLoss` | Counter, kumulativ | `0` |
| Packet Loss (rate) | `packetsReceivedLossRate` | Rate, Snapshot | `0` |
| Retransmissions (sent) | `packetsRetrans` | Counter, kumulativ | `0` |
| Retransmissions (recv) | `packetsReceivedRetrans` | Counter, kumulativ | `0` |
| Verfügbare Bandbreite | `mbpsLinkCapacity` | Snapshot, Mbps | `4352.2` Mbps |
| Tatsächliche Empfangsrate | `mbpsReceiveRate` | Snapshot, Mbps | `1.14` Mbps |

Δt-Validierung (5 s): `bytesReceived` 36.4 M → 37.2 M (monoton +793 KB),
`packetsReceived` 30 136 → 30 797 (monoton +661). `msRTT` und
`mbpsReceiveRate` variieren als Snapshot-Werte (nicht kumulativ),
`packetsReceivedLoss`/`packetsReceivedRetrans` bleiben `0` im
gesunden Lab.

**Mapping-Festlegungen für den späteren Adapter:**

- `available_bandwidth_bps = mbpsLinkCapacity × 1_000_000`. Achtung:
  `mbpsLinkCapacity` ist die SRT-eigene **Schätzung** der maximalen
  Linkkapazität, nicht der konfigurierte Maximalwert (`mbpsMaxBW = -1`
  bedeutet „unlimitiert" und ist getrennt zu behandeln). Lab-Werte
  liegen wegen Loopback-Bandbreite in der Größenordnung mehrerer
  Gbps und sind kein realistischer „verfügbarer"-Wert; Plan §4
  (Tranche 3) muss klären, ob diese Schätzung als Health-Indikator
  in nicht-localhost-Netzen tragfähig ist (Folge-Punkt für
  Tranche 3).
- `throughput_bps = mbpsReceiveRate × 1_000_000` (optional; erfüllt
  RAK-43 nicht allein, siehe §0.1 Tabelle „Erweiterte SRT-Signale").
- Loss-Modell: Quelle liefert beide Formen (Counter + Rate). Adapter
  speichert Counter (`packets_received_loss_total`); Rate ist
  abgeleitet und im Read-Pfad zusätzlich erlaubt.
- Retransmissions: zwei separate Counter (`Retrans` Sender-seitig,
  `ReceivedRetrans` Empfänger-seitig). Für Health-Bewertung relevant
  ist `packetsReceivedRetrans` (was der Receiver tatsächlich nochmal
  bekommt); `packetsRetrans` ist 0 für reine Publish-Verbindungen.

**Freshness-Strategie**: MediaMTX liefert **keinen** expliziten
`source_observed_at`. Adapter setzt `collected_at` zum Polling-
Zeitpunkt; als Source-Sample-Window nutzt er `bytesReceived`-Δ
zwischen aufeinanderfolgenden Polls (monoton steigend bestätigt).
Stale-Erkennung: identischer `bytesReceived` zwischen zwei Polls
plus Verbindung weiterhin im `state: "publish"` zeigt einen
stagnierenden Stream — `source_status: stale_sample` ist das richtige
Mapping. Verbindung verschwindet aus `items[]` →
`source_status: no_active_connection`.

**Fehlerklassen-Mapping**:

| Beobachtung | `source_status` | `source_error_code` |
| ----------- | --------------- | ------------------- |
| HTTP `200 OK`, Item für `path=srt-test` vorhanden, alle Pflichtfelder gesetzt | `ok` | `none` |
| HTTP `200 OK`, `items[]` leer | `no_active_connection` | `no_active_connection` |
| HTTP `200 OK`, Item gefunden, einzelne Pflichtfelder fehlen / non-numeric | `partial` | `partial_sample` |
| HTTP `200 OK`, identisches `bytesReceived` über N Polls trotz `state: "publish"` | `stale` | `stale_sample` |
| HTTP `401`/`403` | `unavailable` | `source_unavailable` |
| HTTP `5xx` oder Connection refused | `unavailable` | `source_unavailable` |
| JSON-Parse-Fehler / Schema-Drift | `unavailable` | `parse_error` |

**Fixture**: [`spec/contract-fixtures/srt/mediamtx-srtconns-list.json`](../../../spec/contract-fixtures/srt/mediamtx-srtconns-list.json)
(anonymisiert: `id`, `remoteAddr`; reale Probe-Werte aus 2026-05-05).

**Auth-Konsequenz für `0.6.0`**: MediaMTX 1.14+ Default ist auth-
pflichtig. Der Adapter braucht einen Auth-Mechanismus (z. B.
`Authorization: Basic ...` mit Lab-Token in `mediamtx.yml`
`authInternalUsers`). Tranche 2 (SRT-Testsetup härten) muss
entscheiden, ob `examples/srt/mediamtx.yml` im `0.6.0`-Stand einen
expliziten Probe-/Health-User bekommt — dieser User braucht
mindestens `action: api`-Permission. Smoke-srt selbst nutzt die API
nicht und bleibt auth-frei für den HLS-Pfad.

**Offene Folge-Punkte für Tranche 3 / 4**:

1. `mbpsLinkCapacity`-Health-Schwelle: in Loopback-Netzen produziert
   die SRT-Schätzung Werte im Gbps-Bereich, die nicht als
   „verfügbare Bandbreite" für Health interpretiert werden können.
   Tranche 3 muss `required_bandwidth_bps` (z. B. 1.5 Mbps für den
   Lab-Stream plus Sicherheitsmarge) gegen `mbpsLinkCapacity`
   bewerten und/oder zu `mbpsReceiveRate`-basierte Heuristik
   wechseln, wenn die Linkkapazitäts-Schätzung nicht aussagekräftig
   ist.
2. `mbpsLinkCapacity = -1` ist kein Probe-Wert hier (`mbpsMaxBW`
   liefert `-1`); Tranche 3/4 muss prüfen, was MediaMTX bei
   getrennter / unbekannter Verbindung liefert.
3. Erweiterte Lastenheft-§4.3-Signale aus dem Probe verfügbar:
   `msReceiveBuf`, `bytesReceiveBuf`, `packetsReceiveBuf` (Receive-
   Buffer-Status), `outboundFramesDiscarded`, `packetsReorderTolerance`.
   §0.1-Tabelle „Erweiterte SRT-Signale" entscheidet, welche davon
   ohne Zusatzrisiko mitfallen.

### 2.5 Verbindliche Wahl (Sub-Tranche 1.3, 2026-05-05)

**Entschieden: Option 1 — MediaMTX-/Server-API über HTTP.**

Begründung anhand der Probe aus §2.4:

- Vollständigkeit der vier RAK-43-Pflichtwerte direkt nachgewiesen
  (`msRTT`, `packetsReceivedLoss`, `packetsRetrans`/`packetsReceivedRetrans`,
  `mbpsLinkCapacity`).
- `apps/api` bleibt CGO-frei: HTTP-Client gegen `:9997`/`:9998`,
  `distroless-static`-Pattern unangefasst.
- Reproduzierbar im lokalen Compose-Lab ohne Internet, ohne Sidecar-
  Build, ohne externe SRT-Tools.
- Cardinality kontrolliert über den Adapter (Source-Rohmetriken
  werden nicht vom Projekt-Prometheus gescraped; per-Verbindung-
  Felder gehen in SQLite und/oder OTel-Spans, nicht in Prometheus-
  Labels — Plan §0.1).
- Fixture als Adapter-Parser-Baseline ist abgelegt
  ([`spec/contract-fixtures/srt/mediamtx-srtconns-list.json`](../../../spec/contract-fixtures/srt/mediamtx-srtconns-list.json)).

**Konsequenz für R-2** (Risiken-Backlog): aufgelöst durch CGO-freie
HTTP-Quelle. Eintrag wandert mit Sub-Tranche 1.4 von §1.1 (Aktiv) nach
§1.2 (Historisch) mit Verweis auf diesen Plan-Block.

**Keine ADR „SRT-Binding-Stack" nötig.** Diese ADR war als Folge-ADR
für libsrt-CGO-Bindings gedacht; mit der CGO-freien Wahl ist sie
obsolet. Roadmap §4 wird im 0.6.0-Closeout entsprechend aktualisiert.

**Adapter-Skizze** (Detail in Tranche 3 / §4):

- Driven-Port `driven.SrtSource` mit `SnapshotConnections(ctx) ([]SrtConnectionSample, error)`.
- HTTP-Client-Adapter `adapters/driven/srt/mediamtxclient` ruft
  `GET {base}/v3/srtconns/list` mit `Authorization: Basic ...` und
  parst die Response gegen das in §2.4 dokumentierte Mapping.
- Keine direkte CGO-/libsrt-Abhängigkeit in `apps/api`.
- Tranche 3 formalisiert das als API-Pull-Vertrag in
  `spec/architecture.md` und `spec/backend-api-contract.md`.

DoD:

- [x] Eine Metrikquelle ist verbindlich gewählt und in einer kurzen ADR
  oder in diesem Plan mit Entscheidungsbegründung dokumentiert
  (Option 1 MediaMTX-API; Begründung §2.4 Probe-Befund + §2.5
  Verbindliche Wahl).
- [x] Die Entscheidung benennt explizit, ob `apps/api` CGO-frei bleibt
  (CGO-frei; HTTP-Client-Adapter; §2.5).
- [-] Falls CGO oder eine andere Runtime-Änderung nötig ist, existiert
  eine accepted ADR "SRT-Binding-Stack" und `risks-backlog.md` R-2 ist
  entsprechend aktualisiert. **N/A** — keine CGO-Wahl. R-2 wird in
  Sub-1.4 als aufgelöst nach §1.2 verschoben.
- [x] Die gewählte Quelle kann RTT, Packet Loss, Retransmissions und
  **verfügbare Bandbreite** vollständig liefern (Probe §2.4
  Befund-Tabelle).
- [x] Falls die gewählte Quelle eigene Prometheus-Metriken anbietet,
  werden diese Rohmetriken nicht in den Projekt-Prometheus gescraped
  (Plan §0.1; Tranche 7 Smoke verifiziert das per
  `scripts/smoke-observability.sh`-Erweiterung).
- [x] Ein minimales Fixture für rohe SRT-Metrikdaten liegt unter
  `spec/contract-fixtures/srt/mediamtx-srtconns-list.json`.
- [x] Für jeden Rohwert ist Einheit und Semantik festgelegt
  (§2.4 Befund-Tabelle: ms, Mbps × 1_000_000 → bps, Counter
  kumulativ, Snapshot-Werte).
- [ ] Quelle oder Lab-Konfiguration liefert einen erwarteten
  Bandbreitenbedarf (`required_bandwidth_bps` oder äquivalente
  Streamrate plus Sicherheitsmarge). **→ Tranche 2/3** (Lab-Stream
  ist 1 Mbps Video plus 96 kbps Audio = ~1.1 Mbps Nutzdaten;
  Sicherheitsmarge in Tranche 3 entscheidet die finale Schwelle).
- [x] Für Counter-Quellen ist festgelegt, wie daraus Dashboard-Werte
  berechnet werden (§2.4 Mapping-Festlegungen: Counter speichern,
  Rate optional ableiten; `bytesReceived`-Δ als Sample-Window).
- [x] Freshness-Semantik ist entschieden (§2.4 „Freshness-
  Strategie"-Block: kein Source-Timestamp im API-Schema; Adapter
  setzt `collected_at`; Sample-Window via `bytesReceived`-Δ als
  Source-Sequence-Surrogat; Stale = identischer `bytesReceived`
  trotz `state: publish`).
- [x] Fehlerklassen der Quelle sind normalisiert (§2.4 „Fehlerklassen-
  Mapping"-Tabelle: `ok`, `no_active_connection`, `partial`,
  `stale`, `unavailable` × `none`/`no_active_connection`/
  `partial_sample`/`stale_sample`/`source_unavailable`/`parse_error`).
- [x] Metrikquelle und Fixture sind ohne externen Netzwerkzugriff in CI
  testbar (Probe lief lokal in Docker-Compose ohne Internet; Fixture
  ist eine reine JSON-Datei).
- [ ] Jede Quelle hat einen expliziten API-Import- oder API-Pull-Vertrag
  für `apps/api`. **→ Tranche 3** (formaler Vertrag in
  `spec/architecture.md` + `spec/backend-api-contract.md`; §2.5
  liefert die Adapter-Skizze als Eingang).
- [x] Ein minimaler Source-Probe-Nachweis existiert (Sub-1.2:
  zwei `curl`-Snapshots gegen Lab-Stack; Schema-Inspektion gegen
  Fixture; alle vier Pflichtwerte plus Fehlerklassen-Mapping
  abgedeckt).
- [x] RAK-42 und RAK-46 sind nicht allein durch diese Tranche erfüllt,
  sondern nur vorbereitet; die Verifikationsmatrix bleibt bis Tranche 7
  offen.

**Übergang in Folgetranchen**: zwei DoD-Items bleiben offen und sind
explizit in Tranche 2/3 verlagert — `required_bandwidth_bps`-Schwelle
und formaler API-Pull-Vertrag. Sub-Tranche 1.4 (R-2 schließen) ist
der letzte Schritt von Tranche 1.

---

## 3. Tranche 2 — SRT-Testsetup zum Health-Lab härten

Bezug: RAK-41; `examples/srt/`; `make smoke-srt`; Lastenheft §7.8
F-82..F-84.

Ziel: Das SRT-Beispiel wird vom reinen Protokoll-Smoke zum Health-
Testsetup erweitert. Der Smoke erzeugt reproduzierbare Metrikdaten und
kann mindestens einen Normalzustand sicher nachweisen.

DoD:

- [x] `examples/srt/compose.yaml` oder der finale `0.5.0`-Startpfad
  enthält alle Container, Ports und Umgebungsvariablen, die für
  Health-Metriken nötig sind (Compose unverändert; `mediamtx.yml`
  um `authInternalUsers`-Block für `action: api`/`metrics` erweitert,
  damit der Health-Smoke `/v3/srtconns/list` lesen darf).
- [x] Der SRT-Publisher erzeugt einen deterministischen Teststream
  (`testsrc2 + sine` aus `ffmpeg-srt-loop.sh`); Stop-Bedingung ist
  Container-Down (Endlos-Loop wird per `docker compose down`
  beendet).
- [x] Das Setup liefert neben der Media-Ausspielung auch eine
  erreichbare Metrikquelle aus Tranche 1 (`http://localhost:9998/v3/srtconns/list`).
- [x] Das Lab benennt die Datenflussrichtung eindeutig
  (`examples/srt/README.md` „Verifikation" → Datenfluss-Block für
  den Health-Pfad: `srt-publisher → mediamtx :8890/udp → /v3/srtconns/list :9997 → host :9998 → smoke-srt-health`).
- [x] `make smoke-srt-health` ergänzt; prüft Publish, Ausspielung und
  Metrikabruf (`scripts/smoke-srt-health.sh`).
- [x] Der Smoke prüft den gesunden Fall (live verifiziert
  2026-05-05, beide Smokes grün): `state=publish`, `msRTT >= 0`,
  `packetsReceivedLoss >= 0`, `packetsReceivedRetrans >= 0`,
  `mbpsLinkCapacity > 0`. Caveat zu Linkkapazität (Loopback liefert
  Gbps-Werte, kein realistischer „verfügbarer"-Wert) ist in §2.4
  Folge-Punkt 1 für Tranche 3 dokumentiert.
- [x] Smoke-Waits sind bounded (`WAIT_SECONDS=45`) und liefern bei
  Fehler Diagnose-Hinweise auf `docker compose logs mediamtx` und
  `srt-publisher`.
- [x] Smoke-Fehler kategorisiert: HLS-Manifest unreachable,
  MediaMTX-API unreachable / leerer Body (Auth-Hinweis), JSON-Parse-
  Fehler, Item-Filter (`path`/`state`) negativ, Pflichtfeld
  fehlt/falsch typisiert. API-Import-Fehler sind weiterhin erst ab
  Tranche 4/7 Scope.
- [x] Stop/Reset räumt nur das `mtrace-srt`-Compose-Projekt auf
  (`docker compose -p mtrace-srt down` im Cleanup-Trap; Smoke
  nutzt keine fremden Project-Namen oder Volumes).
- [x] `examples/srt/README.md` beschreibt den Health-Erweiterungspfad
  ohne die normale `0.5.0`-SRT-Verifikation zu ersetzen (eigener
  „make smoke-srt-health"-Block; Baseline `make smoke-srt` bleibt
  als RAK-37-Nachweis dokumentiert).

---

## 4. Tranche 3 — SRT-Health-Datenmodell, Storage und OTel-Vertrag

Bezug: RAK-42, RAK-46; Telemetry-Model §3; API-Kontrakt §7;
ADR-0002; Lastenheft §7.10 F-95..F-105.

Ziel: SRT-Metriken haben ein eigenes, OTel-kompatibles Modell mit
klarer Cardinality-Grenze. Per-Verbindung-Details sind nicht Prometheus-
Labels, sondern durable Read-Daten und/oder OTel-Span-/Event-Attribute.

Vorgeschlagenes Mindestmodell:

| Feld | Bedeutung | Einheit / Typ |
| ---- | --------- | ------------- |
| `project_id` | kontrollierter Project-Kontext | string, bounded durch Project-Resolver |
| `stream_id` | lokaler SRT-Lab-Stream oder Ingest-Name | string, nicht als Prometheus-Label |
| `connection_id` | Quellseitige Verbindung oder normalisierte ID | string, nicht als Prometheus-Label |
| `source_observed_at` | Zeitpunkt, zu dem die Quelle den SRT-Zustand gemessen hat | timestamp, nullable nur wenn Quelle keine Zeit liefert |
| `source_sequence` | monotone Source-Sample-Sequenz oder Generation-ID, falls Quelle keine Sample-Zeit liefert | integer/string, optional aber Pflicht als Freshness-Ersatz ohne `source_observed_at` |
| `collected_at` | Zeitpunkt, zu dem m-trace/Sidecar die Quelle gelesen hat | timestamp |
| `ingested_at` | Zeitpunkt, zu dem `apps/api` das normalisierte Sample persistiert hat | timestamp |
| `rtt_ms` | Round-trip time | number |
| `packet_loss_total` oder `packet_loss_rate` | Verlustsignal laut Quelle | counter oder ratio, Quelle entscheidet |
| `retransmissions_total` | Retransmission-Counter | counter |
| `available_bandwidth_bps` | verfügbare Link-Bandbreite laut Quelle; nicht bloßer Stream-Durchsatz | bits/s |
| `throughput_bps` | tatsächlich beobachteter Stream-Durchsatz, falls Quelle ihn zusätzlich liefert | bits/s, optional; erfüllt RAK-43 nicht allein |
| `required_bandwidth_bps` | erwarteter Bandbreitenbedarf des Lab-Streams oder der Stream-Konfiguration | bits/s |
| `sample_window_ms` | Zeitfenster für aus Countern abgeleitete Raten, falls relevant | integer, optional |
| `source_status` | Status der Metrikquelle | enum: `ok`, `unavailable`, `partial`, `stale`, `no_active_connection` |
| `source_error_code` | stabile Fehlerklasse bei nicht-`ok`-Status | enum: `source_unavailable`, `no_active_connection`, `partial_sample`, `parse_error`, `stale_sample`, optional `none` |
| `connection_state` | SRT-Verbindungszustand getrennt vom Quellenstatus | enum: `connected`, `no_active_connection`, `unknown` |
| `health_state` | `healthy`, `degraded`, `critical`, `unknown` | enum |

Deferred gegenüber Lastenheft §4.3, sofern die gewählte Quelle sie nicht
ohne zusätzliche Runtime- oder Cardinality-Risiken liefert:

| Signal | Entscheidung für `0.6.0` |
| ------ | ------------------------ |
| Send-/Receive-Buffer | Optionales Zusatzfeld, kein RAK-43-Blocker. |
| Verbindungsstabilität | Kann aus Health-Verlauf abgeleitet werden, aber kein eigener Pflichtwert. |
| Link Health | Wird durch `health_state` repräsentiert; detaillierter Link-Score bleibt Folge-Scope. |
| Failover-Zustände | Deferred, weil `0.6.0` kein Multi-Path-/Failover-Lab liefert. |

### 4.0 Sub-Tranchen-Aufteilung

Tranche 3 ist groß (~21 DoD-Items über Spec, Domain, Storage, Adapter,
Collector, OTel, Tests). Aufteilung in sieben Sub-Tranchen:

| Sub | Inhalt | Form | Status |
| --- | ------ | ---- | ------ |
| 3.1 | Spec-Block: `telemetry-model.md` §3.1/§3.2/§7, `backend-api-contract.md` §7/§7a/§10.6, `architecture.md` §3.3/§3.4/§5.4 | Doku | ✅ (siehe §4.1 unten) |
| 3.2 | Domain-Modell + Driven-Ports (`SrtSource`, `SrtHealthRepository`); Application-Use-Case `SrtHealthCollector` mit Health-Bewertung; Sentinel-Compile-Checks | Code, Hexagon | ✅ |
| 3.3 | SQLite-Schema `srt_health_samples`, Migration im Apply-Runner, Idempotenz-/Restart-Tests; SQLite-Adapter implementiert `SrtHealthRepository` | Code, Storage | ✅ |
| 3.4 | HTTP-Client-Adapter `adapters/driven/srt/mediamtxclient` gegen Fixture aus Sub-1.2 | Code, Adapter | ✅ |
| 3.5 | Collector-Goroutine in `cmd/api`-Setup mit Polling, Backoff, Shutdown; transaktionale Persistenz | Code, Application | ✅ |
| 3.6 | OTel-Span `mtrace.srt.health.collect` + Prometheus bounded Aggregate (`mtrace_srt_health_*`) | Code, Telemetry | ✅ |
| 3.7 | Smoke-/Integrationstest mit zwei Samples; `scripts/smoke-observability.sh` erweitert um SRT-Allowlist-Prüfung | Tests, Smoke | ✅ |

Sub-3.1 ist abgeschlossen; Sub-3.2 ist die nächste Arbeitsstufe.

### 4.1 Spec-Block (Sub-3.1, ✅ 2026-05-05)

DoD:

- [x] `spec/telemetry-model.md` beschreibt SRT-Health-Metriken,
  Einheiten, OTel-Namen/Attribute und Cardinality-Grenzen (neue §7
  mit Sub-Sektionen 7.1–7.9: Datenmodell, deferred Signale, Counter-
  vs-Rate, Health-Bewertung, Source-Status, Freshness, Cardinality,
  OTel, Datenschutz).
- [x] `spec/telemetry-model.md` §3.1 und §3.2 erweitern die
  Prometheus-Label-Allowlist explizit um `health_state` und
  `source_status`; gleichzeitig SRT-Source-Labels (`id`, `path`,
  `remoteAddr`, `state`, `connection_id`, `stream_id`) explizit in
  §3.1 als verboten. `spec/backend-api-contract.md` §7 referenziert
  beide Erweiterungen mit `mtrace_srt_health_*`-Aggregat-Liste.
- [x] `spec/backend-api-contract.md` beschreibt den Read-Vertrag für
  SRT-Health (neue §7a: Endpoints, Response-Struktur mit
  `metrics`/`derived`/`freshness`-Block, Pagination, Fehlerverhalten,
  Pflichttest-Anker) plus Persistenz-Vertrag (neue §10.6: Tabellen-
  schema, Dedupe-/Upsert-Regel, Retention, Migration).
- [x] `spec/architecture.md` beschreibt den neuen SRT-Health-Datenfluss
  (§3.3 Driven-Ports `SrtSource`/`SrtHealthRepository`; §3.4
  Adapter-Tabelle ergänzt um `mediamtxclient` und `sqlite/srt_health/`;
  §5.4 Datenfluss-Diagramm mit Polling-Modell, Backoff, Shutdown,
  Auth-Pfad, Cardinality-Vertrag).
### 4.2 Domain + Application + Adapter + Storage (Sub-3.2..3.5)

Sub-3.2 abgeschlossen 2026-05-05; Sub-3.3..3.5 offen.

DoD:

- [x] Domain-/Application-Port für SRT-Health existiert in `apps/api`
  ohne Import auf konkrete Metrikquelle. Sub-3.2: `domain/srt_health.go`
  (Sample-/Enum-Typen), `port/driven/srt_source.go`,
  `port/driven/srt_health_repository.go`, Application-Use-Case
  `SrtHealthCollector` mit reiner Bewertungsfunktion `Evaluate`.
  Mocks in Test-File belegen Port-Compile-Time-Compliance via
  Sentinel-Checks; `make arch-check` grün.
- [x] Driven-Adapter importiert oder normalisiert Rohmetriken aus der in
  Tranche 1 gewählten Quelle. Sub-3.4: `adapters/driven/srt/mediamtxclient/`
  (`http.go`, `mapping.go`) implementiert `SrtSource` via HTTP-
  Client gegen `/v3/srtconns/list`; mappt MediaMTX-Felder gemäß
  spec/telemetry-model.md §7.1 (`mbpsLinkCapacity × 1_000_000`,
  `bytesReceived` als Source-Sequence-Surrogat,
  `state ∈ {publish,read}` → `connected`). Drei Sentinel-Fehler
  (`ErrSourceUnauthorized`, `ErrSourceUnavailable`,
  `ErrSourceParseError`) für Sub-3.5-Klassifikation. Tests gegen
  go:embed-Fixture aus Sub-1.2 plus 9 weitere Tests
  (Auth, 401, 403, 5xx, Body-Parse-Drift, Empty-Items,
  Unknown-State, Missing-Bandwidth, Response-Too-Large,
  Context-Cancel).
- [x] Collector-/Import-Use-Case ist implementiert und getestet:
  Poll-Intervall, Start/Stop-Verhalten, Konfiguration,
  Fehlerpropagation, Backoff/Retry-Grenzen und Shutdown-Verhalten sind
  dokumentiert und über Tests abgesichert. Sub-3.5: `Run(ctx)` auf
  `SrtHealthCollector` mit `pollInterval`/`maxBackoff`-Optionen
  (`WithPollInterval`/`WithMaxBackoff`/`WithLogger`); exponentielles
  Backoff bei Source-Fehlern bis `DefaultSrtHealthMaxBackoff = 60s`.
  ENV-Konfig in `cmd/api/main.go`: `MTRACE_SRT_SOURCE_URL` (opt-in),
  `MTRACE_SRT_PROJECT_ID` (Default `demo`),
  `MTRACE_SRT_SOURCE_USER`/`_PASS`,
  `MTRACE_SRT_POLL_INTERVAL_SECONDS`. Collector nur aktiv mit
  SQLite (InMemory wird bewusst übersprungen).
- [x] Collector nutzt den expliziten API-Import-/Pull-Vertrag aus
  Tranche 1: HTTP-Client gegen MediaMTX `/v3/srtconns/list` über
  `mediamtxclient`-Adapter. Kein OTLP-only-Pfad — der Adapter ist
  rein Pull-basiert, OTel kommt erst in Sub-3.6 als Zusatz-Export.
- [x] Collector persistiert Samples transaktional: `SrtHealthRepository.Append`
  (Sub-3.3) öffnet eine SQLite-Transaktion (`BEGIN IMMEDIATE`), führt
  Dedupe-Lookup + INSERT pro Sample aus und committet — ein Fehler
  rollt die ganze Charge zurück. OTel-Export ist nicht Pflicht-
  bestandteil von `Append`; das Wiring folgt in Sub-3.6 als
  best-effort nach Commit.
- [x] Collector-/Import-Test weist mindestens zwei aufeinanderfolgende
  Samples mit steigendem `source_sequence` nach
  (`TestRun_AppendsTwoConsecutiveSamples`); MediaMTX-Quelle liefert
  keinen `source_observed_at`, `bytesReceived`-Surrogat funktioniert
  als monotones Sample-Window. Stale-Bewertung greift in
  `TestEvaluate_Stale` und `TestCollect_StaleViaPreviousLookup`
  (Sub-3.2): identischer `source_sequence` über `StaleAfterMillis`
  → `unknown` / `stale_sample`. Backoff- und Shutdown-Verhalten in
  `TestRun_BackoffOnSourceError` / `TestRun_ShutdownOnCancel`.
- [x] SQLite- oder anderer lokaler Persistenzpfad speichert aktuelle und
  historische Health-Snapshots restart-stabil; der Dashboard-Verlauf ist
  `0.6.0`-Pflicht. Sub-3.3: `apps/api/adapters/driven/persistence/sqlite/srt_health_repository.go`
  + `srt_health_scan.go`; `TestSrtHealth_RestartPreservesData`
  weist Close + Re-Open + LatestByStream/HistoryByStream nach.
- [x] Retention-Grenze ist entschieden: unbegrenzt wie bestehende
  lokale SQLite-Demo-Daten — `backend-api-contract.md` §10.6 + V5-
  Migration-Header dokumentieren das. Bounded Snapshot-Historie
  mit Reset-/Prune-Pfad bleibt Folge-Scope.
- [x] Schema-Migration ist idempotent und mit Restart-/Migrationstests
  abgedeckt. V5 läuft via `internal/storage/migrate.go`-Apply-Runner;
  `TestOpen_FreshStart` pinnt `schema_migrations rows = 5`;
  `TestOpen_ReRunIsNoop` deckt Re-Open auf bestehender DB ab.
  `make schema-validate` grün (5 Tabellen, 53 Spalten, 7 Indices,
  10 Constraints).
- [x] Dedupe-/Upsert-Regel ist festgelegt: ein Sample ist eindeutig über
  `(project_id, stream_id, connection_id, COALESCE(source_observed_at, source_sequence))`.
  Adapter macht Vorab-Lookup auf `idx_srt_health_samples_dedupe`,
  überspringt vorhandene Einträge. `TestSrtHealth_DedupeSkipsIdenticalKey`
  pinnt das Verhalten — `collected_at` allein erzeugt **keine**
  zusätzliche Row.

### 4.3 Telemetry + Tests (Sub-3.6..3.7)

DoD (offen, Sub-3.6..3.7):

- [x] OTel-Export ist kompatibel mit dem bestehenden Telemetry-Port und
  vermeidet forbidden Prometheus-Labels. Sub-3.6:
  `Telemetry.SrtSampleRecorded` nimmt `SrtSampleAttrs` (StreamID,
  ConnectionID, HealthState, SourceStatus, RTT, AvailableBandwidth)
  und erzeugt einen Span `mtrace.srt.health.collect` mit
  `mtrace.srt.*`-Attributen — Per-Verbindung-Identifier wandern in
  den Span (sample-basiert), nie in Prometheus-Labels.
- [x] Prometheus erhält höchstens bounded Aggregate. Sub-3.6:
  `mtrace_srt_health_samples_total{health_state}`,
  `mtrace_srt_health_collector_runs_total{source_status}`,
  `mtrace_srt_health_collector_errors_total{source_error_code}` —
  Werte aus den Domain-Enums, mit `_unknown`-Fallback als
  Cardinality-Defense-in-Depth. PrometheusPublisher implementiert
  die drei `SrtHealthSampleAccepted`/`SrtCollectorRun`/
  `SrtCollectorError`-Methoden auf dem MetricsPublisher-Port.
- [x] Neue `mtrace_srt_*`-Metriken werden allowlist-basiert geprüft:
  erlaubte Labels sind ausschließlich `__name__`, `instance`, `job`
  und die in `spec/telemetry-model.md` §3.2 /
  `spec/backend-api-contract.md` §7 neu erlaubten bounded Labels
  (`health_state`, `source_status`, `source_error_code`). Sub-3.7:
  `scripts/smoke-observability.sh` macht eine pro-Metrik-Allowlist-
  Prüfung mit explizitem `allowedByMetric`-Map; jede unerwartete
  Label-Spalte (insbesondere `stream_id`, `connection_id`,
  MediaMTX-`id`/`path`/`remoteAddr`/`state`) ist release-blockierend.
  Konditional: wenn keine `mtrace_srt_health_*`-Serien existieren
  (Collector deaktiviert), wird der Check übersprungen.
- [x] Rohmetriken der Quelle werden nicht in den Projekt-Prometheus
  gescraped. Sub-3.7: smoke-observability.sh fragt
  `/api/v1/targets` ab und sucht nach `mediamtx`/`srt`-Mustern in
  Job/Instance/Scrape-URL — Treffer ist release-blockierend.
- [x] Tests pinnen Einheiten- und Mapping-Verhalten anhand der Fixtures
  aus Tranche 1. Sub-3.4: `mediamtxclient`-Tests gegen go:embed-
  Fixture (`mediamtx-srtconns-list.json`) prüfen
  Bandbreiten-/Counter-/RTT-Mapping numerisch.
- [x] Smoke- oder Integrationstest weist nach, dass der Collector im Lab
  mindestens zwei aufeinanderfolgende Samples importiert und persistiert.
  Sub-3.7: `apps/api/adapters/driven/persistence/sqlite/srt_health_collector_integration_test.go`
  verdrahtet realen SQLite-Storage mit Mock-Source und der
  Collector-Run-Loop; assertert ≥2 persistierte Rows mit
  fortschreitender SourceSequence und `HealthState=healthy`.
- [x] `scripts/smoke-observability.sh` oder ein passender neuer Smoke
  prüft neue `mtrace_srt_*`-Metriken per Label-Allowlist und weist nach,
  dass Source-Rohmetriken nicht als Prometheus-Targets im Projekt-Stack
  konfiguriert sind (Sub-3.7, siehe oben).

---

## 5. Tranche 4 — API-Read-Pfad und Health-Bewertung

Bezug: RAK-43; Lastenheft §7.3 F-21, §7.5.1 F-32/F-34;
API-Kontrakt; Dashboard-Lese-Pfade.

Ziel: Dashboard und Nutzer können SRT-Health über stabile API-
Endpunkte lesen. Der Server berechnet einen einfachen Health-Zustand
aus RTT, Packet Loss, Retransmissions und verfügbarer Bandbreite, ohne
die Rohwerte zu verstecken.

### 5.0 Sub-Tranchen

| Sub | Inhalt | Form | Status |
| --- | ------ | ---- | ------ |
| 4.1 | Application-Query-Service `SrtHealthQueryService` (LatestByStream + HistoryByStream) mit derived/freshness-Ableitung | Code, Application | ✅ |
| 4.2 | HTTP-Handler `SrtHealthListHandler` + `SrtHealthGetHandler` + Router-Wiring; Token-Auth + CORS analog Dashboard-Read | Code, Adapter | ✅ |
| 4.3 | Contract-Fixture `spec/contract-fixtures/api/srt-health-detail.json` + go:embed-Snapshot-Test | Tests, Fixture | ✅ |
| 4.4 | Plan-Closeout | Doku | 🟡 |

### 5.1 DoD

- [x] API-Endpunkt(e) für SRT-Health sind festgelegt und dokumentiert:
  `GET /api/srt/health` und `GET /api/srt/health/{stream_id}`
  (`adapters/driving/http/srt_health_handlers.go`; spec
  `backend-api-contract.md` §7a).
- [x] Read-Responses enthalten alle Pflichtfelder. Sub-4.2:
  `srtHealthWireItem` mit `metrics`/`derived`/`freshness`-Block
  und Top-Level `health_state`/`source_status`/`source_error_code`/
  `connection_state`. Tests gegen Contract-Fixture
  `spec/contract-fixtures/api/srt-health-detail.json` bestätigen
  das Schema.
- [x] `health_state`-Schwellen sind in `application.DefaultThresholds()`
  dokumentiert und über Tabellen-Tests in
  `srt_health_collector_test.go` getestet (RTT 100/250 ms, Loss
  1 %/5 %, Headroom-Faktor 1.5, StaleAfter 15 s).
- [x] Bandbreiten-Health vergleicht `available_bandwidth_bps` gegen
  `required_bandwidth_bps` (Sub-3.2 `evaluateBandwidthHealth`):
  ohne Schwelle bleibt Bandbreite ohne Bewertung, keine
  `degraded`/`critical`-Eskalation.
- [x] API-Response trennt Rohwerte, abgeleitete Werte und Bewertung:
  `metrics` (rtt, loss, retrans, available_bandwidth, throughput,
  required_bandwidth), `derived` (bandwidth_headroom_factor),
  Top-Level Bewertungsfelder.
- [x] Freshness ist sichtbar (`source_observed_at`, `source_sequence`,
  `collected_at`, `ingested_at`, `sample_age_ms`, `stale_after_ms`).
  `sample_age_ms` wird vom Query-Service als Zeit seit
  `IngestedAt` zum Lesezeitpunkt berechnet — kombiniert mit der
  Stale-Schwelle aus `DefaultThresholds`.
- [x] CORS-/Auth-Verhalten folgt Dashboard-Read-Pfaden:
  `resolveProjectFromToken` für Auth, `dashboardPreflightHandler`
  für OPTIONS — analog `/api/stream-sessions`.
- [x] Fehlerfälle stabil: `404 stream_unknown`, `400
  samples_limit_invalid`, `401` ohne Token, `500` bei Repo-Fehler;
  alle in `srt_health_handlers_test.go` abgedeckt.
- [x] API-/Import-Fehler sind kategorisiert: `source_unavailable`,
  `no_active_connection`, `partial_sample`, `parse_error`,
  `stale_sample` (Sub-3.2 EvaluateInput / Sub-3.6
  classifySourceErrorCode).
- [x] Unit-/Handler-Tests decken Normalfall, 404, 401, 400, 500 plus
  Schema-Snapshot ab. Health-State-Tests (degraded/critical/
  unknown/stale) bleiben in der Application-Test-Suite (Sub-3.2),
  weil dort die reine Bewertungsfunktion lebt.
- [x] API-Read-Pfad fügt keine N+1-Erweiterung zu bestehenden
  Session-Listen hinzu: SRT-Health läuft auf eigenen Endpoints
  (`/api/srt/health`), keine Integration in
  `GET /api/stream-sessions`.
- [x] Pagination/Limitierung für historische Samples ist definiert:
  `samples_limit` (default 100, max 1000) im Detail-Endpoint;
  Cursor-Pagination ist als Folge-Item dokumentiert (spec §7a.3
  beschrieben, Sub-3.3 SQLite-Adapter gibt explizit
  ErrNotImplemented bis Folge-Tranche).
- [x] OpenAPI-/Contract-Fixtures oder Snapshot-Tests pinnen den
  Response-Shape: `spec/contract-fixtures/api/srt-health-detail.json`
  + `TestSrtHealthDetail_SchemaMatchesFixture` mit
  Schlüssel-Subset-Vergleich.

---

## 6. Tranche 5 — Dashboard-SRT-Health-Ansicht

Bezug: RAK-43, RAK-44; Lastenheft §7.4 F-25..F-27; bestehende
Dashboard-Session-/Status-Ansichten.

Ziel: Das Dashboard enthält eine nutzbare SRT-Health-Ansicht, die den
aktuellen Zustand schnell lesbar macht und die Rohwerte für Diagnose
zeigt. Sie bleibt operativ schlicht und passt zur bestehenden
Dashboard-Struktur.

### 6.0 Sub-Tranchen

| Sub | Inhalt | Form | Status |
| --- | ------ | ---- | ------ |
| 5.1 | API-Helper `listSrtHealth` / `getSrtHealthDetail` mit Wire-Format-Typen aus spec §7a.2 | Code, Frontend-Library | ✅ |
| 5.2 | Route `/srt-health` mit Tabelle pro Stream (Health-Badge, vier Pflichtmetriken, Freshness, Stale-Hinweis) | Code, SvelteKit-Route | ✅ |
| 5.3 | Detail-Route `/srt-health/[stream_id]` mit aktuellem Sample und Mini-Timeline der letzten 50 Samples | Code, SvelteKit-Route | ✅ |
| 5.4 | Sidebar-Navigation, vitest-Tests (List/Detail/Polling), Plan-Closeout | Tests, Doku | ✅ |

### 6.1 DoD

- [x] Neue Route für SRT-Health in der Sidebar-Navigation
  (`apps/dashboard/src/routes/+layout.svelte`); Tab „SRT health"
  zwischen Errors und Status.
- [x] Ansicht zeigt pro SRT-Stream `health_state` (Pill mit Stale-
  Variante), `rtt_ms`, `packet_loss_total`, `retransmissions_total`,
  `available_bandwidth_bps`, `required_bandwidth_bps` (Detail),
  letzte Aktualisierung mit Sekunden-Alter, Source-Status und
  Source-Error-Code.
- [x] Werte mit Einheiten: `rtt_ms` als ms, `available_bandwidth_bps`
  über `formatBandwidthMbps` als Mbit/s mit drei Nachkommastellen,
  `throughput_bps` separat im Detail; Loss/Retrans als absolute
  Counter (Rate optional via `packet_loss_rate`).
- [x] Warnzustände: CSS-Pill-Klassen `pill healthy`/`degraded`/
  `critical`/`unknown` plus separater `pill stale`-Variant; bei
  Stale-Drift wird der Pill-Text um `(stale)` ergänzt; Source-Status
  ≠ `ok` zeigt eine Source-Hint-Spalte mit Title-Tooltip.
- [x] Mini-Timeline: Detail-Route `/srt-health/[stream_id]` zeigt
  History-Tabelle mit den letzten 50 Samples (samples_limit=50);
  reine Snapshot-Ansicht ist explizit nicht implementiert.
- [x] Loading-/Empty-/Error-/Stale-Zustände implementiert: leere
  Liste → Hinweis „Collector may be disabled"; `notFound` →
  „has no persisted health samples"; sonstige Fehler → `<p
  class="error">{error}</p>`; Stale via `isSrtSampleStale` (entweder
  `source_status = stale` oder `sample_age_ms > stale_after_ms`).
- [x] Dashboard ruft nur die zwei dokumentierten Endpoints
  (`GET /api/srt/health[/{stream_id}]`); keine UI-Schwellenlogik
  (Health-State kommt vom Server via `application.Evaluate`).
- [x] Stale-Daten werden nicht als gesunder Zustand angezeigt:
  `pillClass(item)` priorisiert `stale` über `health_state`; das
  Stale-CSS hat eigene Farbe (gelb-orange) statt grün.
- [x] Tests decken Rendering der vier Pflichtmetriken, Health-Zustände,
  API-Fehler und Polling ab — `apps/dashboard/tests/srt-health.test.ts`
  (API-Helper) plus `srt-health-pages.test.ts` (List + Detail mit
  13 Tests: empty, healthy, degraded, stale, source-status-non-ok,
  error, 404, polling, refresh-button, missing-stream-id,
  source_observed_at-Variante).
- [ ] Browser-E2E oder ein gezielter Dashboard-Smoke validiert die
  Ansicht gegen das lokale SRT-Health-Lab. **→ Tranche 7 Closeout**
  (E2E gegen Lab braucht aktiven Collector mit MediaMTX; das wird
  beim Release-Smoke gesetzt, nicht in Tranche 5 isoliert).

---

## 7. Tranche 6 — Fehlerbild-Dokumentation und Operator-Guide

Bezug: RAK-45; `docs/user/local-development.md`; `examples/srt/README.md`;
`docs/user/releasing.md`.

Ziel: Nutzer können typische SRT-Probleme aus den gelieferten
Metriken ableiten. Die Doku bleibt praktisch und verweist auf konkrete
Lab-Befehle statt abstrakter Protokolltheorie.

Mindestens zu erklärende Fehlerbilder:

| Fehlerbild | Erwartetes Signal |
| ---------- | ----------------- |
| Hohe RTT | `rtt_ms` steigt, Health wird degraded/critical je nach Schwelle. |
| Paketverlust | Loss-Signal steigt, Retransmissions können folgen. |
| Retransmission-Spirale | Retransmissions steigen dauerhaft, verfügbare Bandbreite kann sinken oder instabil werden. |
| Bandbreitenengpass | `available_bandwidth_bps` liegt unter erwarteter Streamrate oder schwankt stark; optionaler `throughput_bps` allein beweist keinen Engpass. |
| Keine Verbindung | Health `unknown` oder `critical`, stale/freshness-Hinweis. |
| Metrikquelle stale | Ausspielung kann noch laufen, aber Health-Samples sind veraltet. |

Nicht-Pflichtwerte aus Lastenheft §4.3 werden in der Doku als
Folge-Scope markiert, solange sie nicht geliefert werden:
Send-/Receive-Buffer, detaillierte Verbindungsstabilität, separater Link-
Health-Score und Failover-Zustände.

DoD:

- [x] `docs/user/srt-health.md` ist die zentrale User-Doku. Zwölf
  Sektionen: Lieferumfang, Quickstart (Smoke + Collector + Read +
  Dashboard), Datenfluss-Diagramm, Metriken-Tabelle inkl. MediaMTX-
  Mapping, Health-Zustände mit Default-Schwellen, Freshness/Stale-
  Erkennung, Source-Status-Tabelle, Fehlerbilder, Cardinality-
  Vertrag, Operator-Quickref, Deferred-Liste, Querverweise.
- [x] `examples/srt/README.md` verlinkt auf
  `docs/user/srt-health.md` (Operator-Doku-Hinweis im Zweck-Abschnitt).
- [x] `docs/user/local-development.md` §2.7 Tabelle erweitert um
  `make smoke-srt-health`-Hinweis; neuer §2.7.1 SRT-Health-View
  mit Quickref und Verlinkung auf `srt-health.md`.
- [x] `docs/user/releasing.md` §2 Smoke-Liste um `make smoke-srt-health`
  ergänzt; neuer §2.1 manuelle `0.6.0`-Prüfungen (Lab + Collector-
  ENV + Dashboard-Route + Stale-Pfad-Test).
- [x] Fehlerbilder-Tabelle in §8 srt-health.md mit acht Mustern
  (RTT, Loss, Retransmissions, Bandbreite, Keine Verbindung, Stale,
  API blockiert, Schema-Drift) plus konkreter Wertebild-Beschreibung.
- [x] §4.1 (Counter vs Rate) und §6 (Freshness) erklären Sample-
  Window, Reset-Verhalten, `bytesReceived`-Surrogat und
  Stale-Erkennung.
- [x] §4.2 (Bandbreite richtig lesen) trennt `available` vs
  `throughput`, dokumentiert Loopback-Gbps-Caveat und das
  Verhalten ohne `required_bandwidth_bps`.
- [x] §11 Deferred/Out-of-Scope listet Buffer-Felder, Verbindungs-
  stabilität, Link-Health-Score, Failover, Cursor-Pagination.
- [x] §9 Cardinality-/Datenschutz-Block benennt verbotene Labels
  und ENV-/Geheimnis-Pflicht für Auth-Credentials.
- [x] Alle neuen Doku-Links laufen durch `make docs-check`.
- [ ] Alle neuen Doku-Links laufen durch `scripts/verify-doc-refs.sh`
  oder `make gates`.

---

## 8. Tranche 7 — Smokes, Gates und Release-Closeout

Bezug: RAK-41..RAK-46; `make gates`; `docs/user/releasing.md`;
`CHANGELOG.md`.

Ziel: Release-Akzeptanz prüft SRT-Health explizit und reproduzierbar.
Die normale Qualitäts-Pipeline bleibt grün; zusätzliche Smokes sind
opt-in, aber für den Release-Closeout verpflichtend.

DoD:

- [ ] `make gates` ist grün.
- [ ] `make smoke-srt` bleibt grün und validiert den Basis-SRT-Pfad aus
  `0.5.0`.
- [ ] `make smoke-srt-health` oder der erweiterte `make smoke-srt`
  validiert Metrikabruf, API-Read-Pfad und mindestens eine Dashboard-
  oder API-Verifikation der vier Pflichtwerte.
- [ ] Health-Smoke prüft neben dem gesunden Fall mindestens einen
  definierten Fehlerpfad: fehlende Metrikquelle, keine aktive
  Verbindung, stale Sample oder API-Importfehler.
- [ ] Observability-Smoke ist grün und weist keine forbidden Labels auf
  neuen `mtrace_*`-Metriken nach. Zusätzlich weist er nach, dass
  Source-Rohmetriken nicht als Projekt-Prometheus-Targets gescraped
  werden.
- [ ] Dashboard-Test/E2E für die SRT-Health-Ansicht ist grün.
- [ ] RAK-Verifikationsmatrix §8.1 ist vollständig ausgefüllt.
- [ ] Release-Closeout-Protokoll §8.2 enthält Befehle, Datum, Ergebnis
  und relevante Notizen.
- [ ] Versionen, `CHANGELOG.md`, README, Roadmap und Release-Notes sind
  für `0.6.0` aktualisiert.
- [ ] `plan-0.6.0.md` ist nach Abschluss nach `docs/planning/done/`
  verschoben; Roadmap verweist danach auf den finalen Pfad.

### 8.1 RAK-Verifikationsmatrix

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-41 | Muss | SRT-Testsetup aus `examples/srt/` plus Health-Smoke | [ ] |
| RAK-42 | Muss | Metrikquelle importiert/erfasst SRT-Verbindungsmetriken; Tests/Fixture pinnen Mapping | [ ] |
| RAK-43 | Muss | API und Dashboard zeigen RTT, Packet Loss, Retransmissions und verfügbare Bandbreite (`available_bandwidth_bps`) | [ ] |
| RAK-44 | Muss | Dashboard-Route/Tab "SRT Health" mit Zuständen, kurzem Verlauf, Fehler-/Stale-Handling und Tests | [ ] |
| RAK-45 | Muss | User-Doku erklärt typische SRT-Fehlerbilder anhand der gelieferten Metriken | [ ] |
| RAK-46 | Muss | Telemetry-Model/API-Kontrakt beschreiben OTel-kompatibles Modell; Observability-Smoke prüft `mtrace_*`-Labels und dass keine Source-Rohmetriken als Projekt-Prometheus-Targets gescraped werden | [ ] |

### 8.2 Release-Closeout-Protokoll

Dieser Abschnitt wird während Tranche 7 gefüllt, damit Smoke-Nachweise
nicht ad hoc in Commit-Bodies oder Release-Notes verschwinden.

| Prüfung | Befehl / Nachweis | Datum | Ergebnis | Notiz |
| ------- | ----------------- | ----- | -------- | ----- |
| `make gates` | — | — | [ ] | — |
| Basis-SRT-Smoke | `make smoke-srt` | — | [ ] | RAK-41 |
| SRT-Health-Smoke | `make smoke-srt-health` oder erweiterter `make smoke-srt` | — | [ ] | RAK-42/43 |
| Observability/Cardinality | `make smoke-observability` oder äquivalenter neuer Smoke, inklusive Nachweis "keine Source-Rohmetriken als Projekt-Prometheus-Targets" | — | [ ] | RAK-46 |
| Dashboard-SRT-Health | Browser-E2E oder gezielter Dashboard-Smoke | — | [ ] | RAK-44 |
| Docs-Gate | `make docs-check` oder Teil von `make gates` | — | [ ] | RAK-45 |
