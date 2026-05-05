# Implementation Plan — `0.6.0` (SRT Health View)

> **Status**: ⬜ offen. Dieser Plan ist vorbereitet, aber noch nicht
> aktive Phase. `0.5.0` (Multi-Protocol Lab) muss vorher abgeschlossen
> und released sein; insbesondere muss das SRT-Beispiel aus
> [`plan-0.5.0.md`](../in-progress/plan-0.5.0.md) Tranche 3 als
> reproduzierbarer Lab-Pfad vorliegen.
>
> **Bezug**: [Lastenheft `1.1.8`](../../../spec/lastenheft.md) §4.3
> (SRT als späterer starker Hebel), §7.8 (lokales Streaming-Lab), §7.9
> (Observability), §7.10 (Telemetry/Cardinality), §8.3 NF-11
> (SRT-Ingest-Metriken), §13.8 (RAK-41..RAK-46);
> [Roadmap](../in-progress/roadmap.md) §3/§4;
> [Architektur](../../../spec/architecture.md);
> [Telemetry-Model](../../../spec/telemetry-model.md);
> [API-Kontrakt](../../../spec/backend-api-contract.md);
> [Risiken-Backlog](./risks-backlog.md) R-2.
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
| Prometheus | Nur bounded Aggregate sind erlaubt. Per-Verbindung-/Per-Stream-Details gehen in SQLite und/oder OTel-Spans/Events, nicht in Prometheus-Labels. Rohmetriken aus MediaMTX oder anderen Quellen werden nicht in den Projekt-Prometheus gescraped, solange sie Labels wie `id`, `path`, `remoteAddr`, `state`, Connection-ID oder IP enthalten; alternativ muss der Smoke diese Source-Targets mit derselben Forbidden-Label-Policy prüfen. | `stream_id`, `session_id`, URL, Token, IP oder Connection-ID als Prometheus-Label. |
| Dashboard | Eine eigene SRT-Health-Ansicht zeigt aktuelle Verbindung, Verlauf und Warnzustände. | Vollständige Media-Server-Konsole oder Stream-Key-Verwaltung. |
| Lab | `0.6.0` baut auf `examples/srt/` aus `0.5.0` auf und härtet es für Health-Nachweise. | Neues paralleles SRT-Lab ohne Bezug zum bestehenden Beispiel. |
| Fehlerbilder | Dokumentation erklärt typische SRT-Probleme anhand der gelieferten Metriken. | Allgemeines SRT-Lehrbuch oder produktive Netzwerk-Tuning-Anleitung. |
| Erweiterte SRT-Signale | `0.6.0` priorisiert die RAK-43-Pflichtwerte RTT, Packet Loss, Retransmissions und Bandbreite. Send-/Receive-Buffer, Verbindungsstabilität, Link Health und Failover-Zustände aus Lastenheft §4.3 werden als Deferred-Liste geführt, sofern sie nicht ohne Zusatzrisiko aus der gewählten Quelle mitfallen. | RAK-43 durch nicht geforderte Zusatzwerte erweitern und damit den Release-Scope aufblasen. |

### 0.2 Risiko-Triage aus dem Backlog

| Risiko | Entscheidung für `0.6.0` | Trigger / Nachweis |
| ------ | ------------------------ | ------------------ |
| R-2 CGO/SRT-Bindings | Vor jeder Code-Integration entscheiden. Bevorzugt wird ein Sidecar-/Import-Pfad, der `apps/api` CGO-frei hält. Falls nur CGO sinnvoll ist, ist eine ADR Pflicht. | Tranche 1 liefert ADR oder expliziten "kein CGO in API"-Beschluss mit Smoke-Nachweis. |
| R-5 Time-Skew-Persistenz | Nicht Teil der SRT-Health-Pflicht, außer SRT-Metriken brauchen Client-/Server-Zeitvergleich im Dashboard. | Wenn Health-Events Zeitversatz bewerten müssen, wird eine additive Sub-Tranche ergänzt oder R-5 separat aktiviert. |
| R-7 Session-List-N+1 | Beobachten. SRT-Health darf Session-Listen nicht durch zusätzliche N+1-Reads verschlechtern. | Wenn SRT-Health in `GET /api/stream-sessions` eingebettet wird und p95 >= 200 ms reproduzierbar wird, Bulk-Read-Port vor Dashboard-Integration liefern. |
| R-10 Sampling-Vollständigkeit | Nicht Teil von `0.6.0`, solange SRT-Metriken unabhängig von Player-Sampling laufen. | Aktivieren, falls Health-Ansicht Player-Event-Vollständigkeit als Diagnosevoraussetzung behauptet. |

### 0.3 Sequenzierung und harte Gates

`0.6.0` hat eine höhere Fehlentscheidungsgefahr als die vorherigen
Lab-Tranchen, weil die Metrikquelle Runtime, Cardinality und Dashboard-
Vertrag gleichzeitig beeinflusst. Daher gelten diese Reihenfolgen:

1. Tranche 1 blockiert jede produktive Code-Integration, die eine
   SRT-Metrikquelle in `apps/api`, Prometheus oder Dashboard verdrahtet.
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
| 0 | Vorgänger-Gate und Scope-Festlegung | ⬜ |
| 1 | SRT-Metrikquelle und Binding-Entscheidung (R-2, RAK-42) | ⬜ |
| 2 | SRT-Testsetup zum Health-Lab härten (RAK-41) | ⬜ |
| 3 | SRT-Health-Datenmodell, Storage und OTel-Vertrag (RAK-42, RAK-46) | ⬜ |
| 4 | API-Read-Pfad und Health-Bewertung (RAK-43) | ⬜ |
| 5 | Dashboard-SRT-Health-Ansicht (RAK-43, RAK-44) | ⬜ |
| 6 | Fehlerbild-Dokumentation und Operator-Guide (RAK-45) | ⬜ |
| 7 | Smokes, Gates und Release-Closeout (RAK-41..RAK-46) | ⬜ |

---

## 1a. Tranche 0 — Vorgänger-Gate und Scope-Festlegung

Bezug: Roadmap §1/§3/§4; `plan-0.5.0.md` §4/§7; Risiken-Backlog R-2.

Ziel: Vor Implementierung ist klar, dass `0.6.0` auf einem
funktionierenden SRT-Lab aufsetzt und nur SRT-Health liefert, nicht
Ingest-Orchestrierung oder produktive SRT-Betriebsverwaltung.

DoD:

- [ ] `plan-0.5.0.md` ist vollständig abgeschlossen, nach
  `docs/planning/done/` verschoben und mit finalen Commit-Hashes
  versehen.
- [ ] Annotierter Release-Tag `v0.5.0` existiert und GitHub Actions
  `Build` ist für den Release-Commit grün.
- [ ] `examples/srt/README.md` beschreibt den finalen `0.5.0`-SRT-Pfad
  inklusive Start, Verifikation, Stop/Reset, Troubleshooting und
  bekannter Grenzen.
- [ ] `make smoke-srt` ist auf dem `0.5.0`-Release-Stand grün und als
  Baseline für `0.6.0` dokumentiert.
- [ ] Roadmap §1.2 und §3 markieren `0.6.0` als aktive Phase und
  verweisen auf diesen Plan.
- [ ] R-2 ist für den Startzustand präzisiert: bekannte Metrikquellen,
  Runtime-Auswirkungen und offene ADR-Frage sind im Risiken-Backlog
  aktuell.
- [ ] Scope-Grenze ist in README/User-Doku sichtbar: `0.6.0` liefert
  SRT-Health, nicht Stream-Key-Verwaltung, SRT-Auth, Multi-Publisher-
  Routing oder Kubernetes.

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
| MediaMTX-/Server-API | SRT-Verbindungsdaten vom lokalen Media-Server lesen. | Bevorzugt, wenn RTT/Loss/Retransmissions/Bandbreite vollständig und stabil verfügbar sind. |
| Sidecar-Exporter | Separater Container sammelt SRT-Stats und liefert HTTP/JSON oder OTLP an `apps/api`. | Bevorzugt, wenn `apps/api` CGO-frei bleiben soll und die Quelle trotzdem vollständig ist. |
| Log-/CLI-Import | Lab-Smoke oder Sidecar normalisiert bekannte SRT-Tool-Ausgabe. | Nur akzeptabel, wenn deterministisch testbar und nicht fragil gegen lokalisierte Logtexte. |
| Direktes libsrt-Binding | `apps/api` liest SRT-Stats direkt über Binding. | Nur mit ADR und bewusst akzeptierter Runtime-/Image-Konsequenz. |

Harte Auswahlkriterien:

| Kriterium | Muss erfüllt sein |
| --------- | ----------------- |
| Vollständigkeit | RTT, Packet Loss, Retransmissions und Bandbreite sind alle verfügbar und semantisch erklärbar. |
| Reproduzierbarkeit | Fixtures und Smoke laufen ohne Internet und ohne manuelle SRT-Tools auf dem Host. |
| Runtime-Grenze | `apps/api` bleibt CGO-frei oder die Runtime-Änderung ist per ADR akzeptiert. |
| Cardinality | Rohdaten gelangen nicht ungefiltert mit Per-Verbindung-/IP-Labels in Prometheus. |
| Freshness | Quelle liefert `observed_at` oder ein äquivalentes Sample-Zeitfenster, sodass stale Daten erkennbar sind. |
| Failure-Mode | Quelle hat unterscheidbare Fehler für "nicht erreichbar", "keine Verbindung" und "unvollständige Rohdaten". |

DoD:

- [ ] Eine Metrikquelle ist verbindlich gewählt und in einer kurzen ADR
  oder in diesem Plan mit Entscheidungsbegründung dokumentiert.
- [ ] Die Entscheidung benennt explizit, ob `apps/api` CGO-frei bleibt.
- [ ] Falls CGO oder eine andere Runtime-Änderung nötig ist, existiert
  eine accepted ADR "SRT-Binding-Stack" und `risks-backlog.md` R-2 ist
  entsprechend aktualisiert.
- [ ] Die gewählte Quelle kann RTT, Packet Loss, Retransmissions und
  Bandbreite vollständig liefern. Fehlt einer dieser vier RAK-43-Werte,
  blockiert Tranche 1 (`[!]`) und es braucht entweder eine andere Quelle
  oder einen Lastenheft-Patch; eine bloß dokumentierte Mapping-Grenze
  reicht für RAK-43 nicht.
- [ ] Falls die gewählte Quelle eigene Prometheus-Metriken anbietet
  (z. B. MediaMTX-SRT-Metrics), werden diese Rohmetriken nicht in den
  Projekt-Prometheus gescraped, solange sie verbotene Per-Verbindung-,
  Pfad-, IP- oder Connection-ID-Labels tragen. Wenn sie doch gescraped
  werden müssen, ist ein Smoke-Gate Pflicht, das Source-Metriken und
  `mtrace_*`-Metriken mit derselben Forbidden-Label-Policy prüft.
- [ ] Ein minimales Fixture für rohe SRT-Metrikdaten liegt unter
  `spec/contract-fixtures/` oder einem komponentennahen `testdata/`
  Verzeichnis.
- [ ] Für jeden Rohwert ist Einheit und Semantik festgelegt
  (z. B. Millisekunden, Bytes/s oder Bits/s, absolute Counter vs.
  Intervallwert).
- [ ] Für Counter-Quellen ist festgelegt, wie daraus Dashboard-Werte
  berechnet werden: Roh-Counter anzeigen, Intervallrate ableiten oder
  beides liefern. Das Intervall und Reset-Verhalten sind dokumentiert.
- [ ] Freshness-Semantik ist entschieden: Sample-Intervall, stale-
  Schwelle und Verhalten bei fehlendem Zeitstempel.
- [ ] Fehlerklassen der Quelle sind normalisiert:
  `source_unavailable`, `no_active_connection`, `partial_sample`,
  `parse_error` oder äquivalente stabile Codes.
- [ ] Metrikquelle und Fixture sind ohne externen Netzwerkzugriff in CI
  testbar.
- [ ] RAK-42 und RAK-46 sind nicht allein durch diese Tranche erfüllt,
  sondern nur vorbereitet; die Verifikationsmatrix bleibt bis Tranche 7
  offen.

---

## 3. Tranche 2 — SRT-Testsetup zum Health-Lab härten

Bezug: RAK-41; `examples/srt/`; `make smoke-srt`; Lastenheft §7.8
F-82..F-84.

Ziel: Das SRT-Beispiel wird vom reinen Protokoll-Smoke zum Health-
Testsetup erweitert. Der Smoke erzeugt reproduzierbare Metrikdaten und
kann mindestens einen Normalzustand sicher nachweisen.

DoD:

- [ ] `examples/srt/compose.yaml` oder der finale `0.5.0`-Startpfad
  enthält alle Container, Ports und Umgebungsvariablen, die für
  Health-Metriken nötig sind.
- [ ] Der SRT-Publisher erzeugt einen deterministischen Teststream mit
  begrenzter Laufzeit oder klarer Stop-Bedingung.
- [ ] Das Setup liefert neben der Media-Ausspielung auch eine
  erreichbare Metrikquelle aus Tranche 1.
- [ ] Das Lab benennt die Datenflussrichtung eindeutig:
  Publisher → SRT-Receiver/Media-Server → Metrikquelle → `apps/api`
  → Dashboard.
- [ ] `make smoke-srt` wird erweitert oder ein neues
  `make smoke-srt-health` wird ergänzt; der Befehl prüft Publish,
  Ausspielung und Metrikabruf.
- [ ] Der Smoke prüft mindestens den gesunden Fall: Verbindung aktiv,
  RTT vorhanden, Packet-Loss-Signal vorhanden, Retransmission-Signal
  vorhanden und Bandbreite > 0 oder ein quellspezifisch korrekt
  normalisierter Wert.
- [ ] Smoke-Waits sind bounded und liefern Diagnoseausgabe aus
  Metrikquelle, Media-Server und Publisher.
- [ ] Smoke-Fehler sind kategorisiert: Publish fehlgeschlagen,
  Ausspielung fehlt, Metrikquelle fehlt, Sample unvollständig,
  API-Import fehlgeschlagen.
- [ ] Stop/Reset räumt nur das `mtrace-srt`-Compose-Projekt auf und
  greift nicht in Core-Lab-Volumes ein.
- [ ] `examples/srt/README.md` beschreibt den Health-Erweiterungspfad
  ohne die normale `0.5.0`-SRT-Verifikation zu ersetzen.

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
| `observed_at` | Serverzeit der Messung | timestamp |
| `rtt_ms` | Round-trip time | number |
| `packet_loss_total` oder `packet_loss_rate` | Verlustsignal laut Quelle | counter oder ratio, Quelle entscheidet |
| `retransmissions_total` | Retransmission-Counter | counter |
| `bandwidth_bps` | geschätzte oder gemessene Bandbreite | bits/s |
| `sample_window_ms` | Zeitfenster für aus Countern abgeleitete Raten, falls relevant | integer, optional |
| `source_status` | Status der Metrikquelle | enum: `ok`, `unavailable`, `partial`, `stale` |
| `health_state` | `healthy`, `degraded`, `critical`, `unknown` | enum |

Deferred gegenüber Lastenheft §4.3, sofern die gewählte Quelle sie nicht
ohne zusätzliche Runtime- oder Cardinality-Risiken liefert:

| Signal | Entscheidung für `0.6.0` |
| ------ | ------------------------ |
| Send-/Receive-Buffer | Optionales Zusatzfeld, kein RAK-43-Blocker. |
| Verbindungsstabilität | Kann aus Health-Verlauf abgeleitet werden, aber kein eigener Pflichtwert. |
| Link Health | Wird durch `health_state` repräsentiert; detaillierter Link-Score bleibt Folge-Scope. |
| Failover-Zustände | Deferred, weil `0.6.0` kein Multi-Path-/Failover-Lab liefert. |

DoD:

- [ ] `spec/telemetry-model.md` beschreibt SRT-Health-Metriken,
  Einheiten, OTel-Namen/Attribute und Cardinality-Grenzen.
- [ ] `spec/backend-api-contract.md` beschreibt den Read-Vertrag für
  SRT-Health oder verweist auf einen eigenen neuen Abschnitt.
- [ ] Domain-/Application-Port für SRT-Health existiert in `apps/api`
  ohne Import auf konkrete Metrikquelle.
- [ ] Driven-Adapter importiert oder normalisiert Rohmetriken aus der in
  Tranche 1 gewählten Quelle.
- [ ] SQLite- oder anderer lokaler Persistenzpfad speichert aktuelle und
  historische Health-Snapshots restart-stabil, falls die Dashboard-
  Ansicht Verlauf zeigen soll.
- [ ] Retention-Grenze ist entschieden: unbegrenzt wie bestehende
  lokale SQLite-Demo-Daten oder bounded Snapshot-Historie mit
  dokumentiertem Reset-/Prune-Pfad.
- [ ] Schema-Migration ist idempotent und mit Restart-/Migrationstests
  abgedeckt.
- [ ] Dedupe-/Upsert-Regel ist festgelegt: Ein Sample ist eindeutig über
  Quelle, Stream/Connection, `observed_at` und ggf. Sample-Sequenz.
- [ ] OTel-Export ist kompatibel mit dem bestehenden Telemetry-Port und
  vermeidet forbidden Prometheus-Labels.
- [ ] Prometheus erhält höchstens bounded Aggregate, z. B. Anzahl
  Health-Samples nach `health_state`; keine `stream_id`,
  `connection_id`, URL, IP oder Token als Label.
- [ ] Rohmetriken der Quelle werden entweder gar nicht in den
  Projekt-Prometheus gescraped oder durch ein Smoke-Gate auf dieselbe
  Forbidden-Label-Liste geprüft wie `mtrace_*`-Metriken.
- [ ] Tests pinnen Einheiten- und Mapping-Verhalten anhand der Fixtures
  aus Tranche 1.
- [ ] `scripts/smoke-observability.sh` oder ein passender neuer Smoke
  prüft, dass neue `mtrace_*`-Metriken und, falls gescraped, Source-
  Metriken keine verbotenen Labels tragen.

---

## 5. Tranche 4 — API-Read-Pfad und Health-Bewertung

Bezug: RAK-43; Lastenheft §7.3 F-21, §7.5.1 F-32/F-34;
API-Kontrakt; Dashboard-Lese-Pfade.

Ziel: Dashboard und Nutzer können SRT-Health über stabile API-
Endpunkte lesen. Der Server berechnet einen einfachen Health-Zustand
aus RTT, Packet Loss, Retransmissions und Bandbreite, ohne die Rohwerte
zu verstecken.

DoD:

- [ ] API-Endpunkt(e) für SRT-Health sind festgelegt und dokumentiert,
  z. B. `GET /api/srt/health` und optional
  `GET /api/srt/health/{stream_id}`.
- [ ] Read-Responses enthalten mindestens RTT, Packet Loss,
  Retransmissions, Bandbreite, `observed_at`, `health_state` und eine
  Quellen-/Freshness-Angabe.
- [ ] `health_state`-Schwellen sind dokumentiert und testbar; `unknown`
  ist der definierte Zustand bei fehlender oder stale Metrikquelle.
- [ ] API-Response trennt Rohwerte, abgeleitete Werte und Bewertung:
  `metrics`, `derived`, `health_state`, `source_status` oder eine
  gleichwertige Struktur.
- [ ] Freshness ist im Response sichtbar, z. B. `observed_at`,
  `sample_age_ms`, `stale_after_ms` und `source_status`.
- [ ] CORS-/Auth-Verhalten folgt den bestehenden Dashboard-Read-Pfaden
  und ist im API-Kontrakt beschrieben.
- [ ] Fehlerfälle sind stabil: Metrikquelle nicht erreichbar,
  unvollständige Rohdaten, stale Daten, ungültige Stream-ID.
- [ ] Unit-/Handler-Tests decken Normalfall, degraded/critical,
  unknown/stale und Auth/CORS ab.
- [ ] API-Read-Pfad fügt keine N+1-Erweiterung zu bestehenden
  Session-Listen hinzu; falls Integration in `GET /api/stream-sessions`
  nötig ist, existiert ein Bulk-Read-Port oder ein begründeter
  separater Endpoint.
- [ ] Pagination oder Limitierung für historische Samples ist definiert;
  unbeschränkte Zeitreihen-Antworten sind nicht zulässig.
- [ ] OpenAPI-/Contract-Fixtures oder Snapshot-Tests pinnen den
  Response-Shape.

---

## 6. Tranche 5 — Dashboard-SRT-Health-Ansicht

Bezug: RAK-43, RAK-44; Lastenheft §7.4 F-25..F-27; bestehende
Dashboard-Session-/Status-Ansichten.

Ziel: Das Dashboard enthält eine nutzbare SRT-Health-Ansicht, die den
aktuellen Zustand schnell lesbar macht und die Rohwerte für Diagnose
zeigt. Sie bleibt operativ schlicht und passt zur bestehenden
Dashboard-Struktur.

DoD:

- [ ] Neue Route oder Tab für SRT-Health ist in der Dashboard-Navigation
  erreichbar.
- [ ] Ansicht zeigt pro SRT-Stream oder Verbindung mindestens
  `health_state`, RTT, Packet Loss, Retransmissions, Bandbreite,
  letzte Aktualisierung und Quelle/Freshness.
- [ ] Werte sind mit Einheiten und Zeitbezug sichtbar: RTT in ms,
  Bandbreite in bit/s oder Mbit/s, Loss/Retransmission als Counter oder
  Rate gemäß API-Vertrag.
- [ ] Warnzustände unterscheiden `degraded`, `critical`, `unknown` und
  normalen Zustand visuell und textlich eindeutig.
- [ ] Verlauf oder Mini-Timeline ist vorhanden, wenn Tranche 3
  historische Snapshots persistiert; andernfalls ist die Ansicht klar
  als aktueller Snapshot gekennzeichnet.
- [ ] Loading-, Empty-, Error- und Stale-Zustände sind implementiert und
  getestet.
- [ ] Dashboard ruft nur dokumentierte API-Endpunkte auf und dupliziert
  keine Health-Schwellenlogik, außer reine UI-Formatierung.
- [ ] Stale-Daten werden nicht als gesunder Zustand angezeigt; die UI
  muss Freshness/Quelle sichtbar machen.
- [ ] Tests decken Rendering der vier Pflichtmetriken, Health-Zustände
  und API-Fehler ab.
- [ ] Browser-E2E oder ein gezielter Dashboard-Smoke validiert die
  Ansicht gegen das lokale SRT-Health-Lab.

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
| Retransmission-Spirale | Retransmissions steigen dauerhaft, Bandbreite kann sinken oder instabil werden. |
| Bandbreitenengpass | `bandwidth_bps` unter erwarteter Streamrate oder stark schwankend. |
| Keine Verbindung | Health `unknown` oder `critical`, stale/freshness-Hinweis. |
| Metrikquelle stale | Ausspielung kann noch laufen, aber Health-Samples sind veraltet. |

Nicht-Pflichtwerte aus Lastenheft §4.3 werden in der Doku als
Folge-Scope markiert, solange sie nicht geliefert werden:
Send-/Receive-Buffer, detaillierte Verbindungsstabilität, separater Link-
Health-Score und Failover-Zustände.

DoD:

- [ ] `docs/user/srt-health.md` oder eine passende User-Doku-Seite
  erklärt SRT-Health-View, Metriken, Einheiten, Health-Zustände und
  Grenzen.
- [ ] `examples/srt/README.md` verlinkt auf die Health-Doku und enthält
  kurze Smoke-/Troubleshooting-Hinweise.
- [ ] `docs/user/local-development.md` nennt den SRT-Health-Startpfad
  und die relevanten URLs/Make-Targets.
- [ ] `docs/user/releasing.md` nennt zusätzliche `0.6.0`-Smokes und
  manuelle Prüfungen.
- [ ] Dokumentation erklärt die typischen Fehlerbilder aus der Tabelle
  mit konkreten Messwerten oder Zustandsbeispielen.
- [ ] Dokumentation erklärt Counter-vs.-Rate-Semantik, Sample-Fenster
  und stale/unknown-Zustände, damit Operatoren Werte nicht als
  absolute Momentaufnahme missverstehen.
- [ ] Dokumentation enthält eine Deferred-Liste für Send-/Receive-
  Buffer, Verbindungsstabilität, Link Health und Failover-Zustände,
  falls diese Signale nicht in `0.6.0` ausgeliefert werden.
- [ ] Dokumentation benennt Datenschutz-/Cardinality-Grenzen: keine
  per-Verbindung-Identifier in Prometheus-Labels, keine Secrets in
  Metriken.
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
  Verbindung oder stale Sample.
- [ ] Observability-Smoke ist grün und weist keine forbidden Labels auf
  neuen `mtrace_*`-Metriken und, falls gescraped, Source-Metriken nach.
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
| RAK-43 | Muss | API und Dashboard zeigen RTT, Packet Loss, Retransmissions und Bandbreite | [ ] |
| RAK-44 | Muss | Dashboard-Route/Tab "SRT Health" mit Zuständen, Fehler-/Stale-Handling und Tests | [ ] |
| RAK-45 | Muss | User-Doku erklärt typische SRT-Fehlerbilder anhand der gelieferten Metriken | [ ] |
| RAK-46 | Muss | Telemetry-Model/API-Kontrakt beschreiben OTel-kompatibles Modell; Observability-Smoke prüft `mtrace_*` und ggf. Source-Metriken auf verbotene Labels | [ ] |

### 8.2 Release-Closeout-Protokoll

Dieser Abschnitt wird während Tranche 7 gefüllt, damit Smoke-Nachweise
nicht ad hoc in Commit-Bodies oder Release-Notes verschwinden.

| Prüfung | Befehl / Nachweis | Datum | Ergebnis | Notiz |
| ------- | ----------------- | ----- | -------- | ----- |
| `make gates` | — | — | [ ] | — |
| Basis-SRT-Smoke | `make smoke-srt` | — | [ ] | RAK-41 |
| SRT-Health-Smoke | `make smoke-srt-health` oder erweiterter `make smoke-srt` | — | [ ] | RAK-42/43 |
| Observability/Cardinality | `make smoke-observability` oder äquivalenter neuer Smoke, inklusive Source-Metriken falls gescraped | — | [ ] | RAK-46 |
| Dashboard-SRT-Health | Browser-E2E oder gezielter Dashboard-Smoke | — | [ ] | RAK-44 |
| Docs-Gate | `make docs-check` oder Teil von `make gates` | — | [ ] | RAK-45 |
