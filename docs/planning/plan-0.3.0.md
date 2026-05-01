# Implementation Plan — `0.3.0` (Stream Analyzer)

> **Status**: 🟡 in Arbeit. Beginnt nach Release `v0.2.0` (Publizierbares Player SDK).  
> **Bezug**: [Lastenheft `1.1.7`](../../spec/lastenheft.md) §7.7 (Stream Analyzer), §13.5 (RAK-22..RAK-28), §18 (MVP-DoD-Anteil); [Roadmap](./roadmap.md) §2/§3; [Architektur (Zielbild)](../../spec/architecture.md); [API-Kontrakt](../../spec/backend-api-contract.md); [Risiken-Backlog](./risks-backlog.md); [ADR 0002 Persistenz-Store](../adr/0002-persistence-store.md).
> **Vorgänger-Gate (Stand zum `0.3.0`-Start)**:
>
> - [`plan-0.2.0.md`](./plan-0.2.0.md) ist vollständig (`[x]`) und `v0.2.0` ist veröffentlicht.
> - GitHub Actions `Build` ist für den Release-Commit `v0.2.0` grün.
> - Offene Persistenzfragen aus OE-3 blockieren `0.3.0` nur, wenn Analyseergebnisse durable gespeichert werden müssen.
>
> **Nachfolger**: `plan-0.4.0.md` (Erweiterte Trace-Korrelation).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog [`plan-0.1.0.md`](./plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz (siehe `roadmap.md` §7.1).
- 🟡 in Arbeit.

Neue Lastenheft-Patches während `0.3.0` landen weiterhin zentral in `plan-0.1.0.md` Tranche 0c, weil sie projektweit gelten.

---

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
|---|---|---|
| 0 | Vorgänger-Gate-Verifikation | ✅ |
| 1 | Stream-Analyzer-Paket und Public API | ✅ |
| 2 | HLS-Manifest laden und Playlist-Typ erkennen | ✅ |
| 3 | Master-Playlist-Analyse | ✅ |
| 4 | Media-Playlist-Analyse und Segment-Dauerprüfung | ✅ |
| 5 | JSON-Ergebnisformat und Dokumentation | ✅ |
| 6 | API-Anbindung über StreamAnalyzer-Port | ✅ |
| 7 | CLI-Grundlage | ✅ |
| 7.5 | Tranche-6-Folge-Issues härten | ✅ |
| 8 | Release-Akzeptanzkriterien `0.3.0` | ✅ |

---

## 1a. Tranche 0 — Vorgänger-Gate-Verifikation

Konvertiert die narrative Vorgänger-Gate-Beschreibung aus §0 in prüfbare DoD-Items.

DoD:

- [x] `plan-0.2.0.md` ist vollständig (`[x]`), inklusive Release-Akzeptanzkriterien (`a2cb66d`).
- [x] Annotierter Release-Tag `v0.2.0` existiert und zeigt auf den Release-Commit (`a2cb66d`).
- [x] GitHub-Release `v0.2.0` existiert und verweist auf die Changelog-Release-Notes (`a2cb66d`).
- [x] GitHub Actions `Build` ist für den Release-Commit grün (`a2cb66d`).
- [x] `docs/planning/roadmap.md` §2 markiert Schritt 19 als abgeschlossen und führt die Folgearbeiten für `0.3.0` (`070b231`).
- [x] Persistenzstatus ist geprüft: ADR 0002 bleibt nicht-blockierend, solange Analyseergebnisse nicht durable gespeichert werden (`37f20ce`).

---

## 2. Tranche 1 — Stream-Analyzer-Paket und Public API

Bezug: Lastenheft §7.7; RAK-22..RAK-26; MVP-33.

Ziel: `packages/stream-analyzer` wird als TypeScript-Paket mit stabiler Public API angelegt. Das Paket ist für Backend-Integration und CLI nutzbar, ohne Dashboard- oder API-Interna zu importieren.

DoD:

- [x] Workspace-Paket `packages/stream-analyzer` ist angelegt und in `pnpm-workspace.yaml` enthalten (`5aa3f34`).
- [x] `package.json` enthält eindeutige Paketmetadaten, `type`, `main`, `module`, `types`, `exports`, `files`, `license`, `repository.directory` und CI-taugliche Scripts (`5aa3f34`).
- [x] `packages/stream-analyzer/package.json.version` wird für `0.3.0` mit dem Repo-Release versioniert; abweichende unabhängige Versionierung wäre vor Umsetzung als explizite Versionierungsregel zu dokumentieren (`5aa3f34`).
- [x] Analyzer-Version wird aus genau einer Quelle abgeleitet: `packages/stream-analyzer/package.json`, exportierte Version und die im JSON-Ergebnis gesendete Analyzer-Version sind synchron und getestet (`5aa3f34`).
- [x] Public API exportiert mindestens eine Analysefunktion für HLS-Manifeste und klar typisierte Ergebnis-/Fehlertypen (`5aa3f34`).
- [x] API-/Adapter-Kontrakt für Backend-Nutzung ist vor Parser-Implementierung entschieden: Manifest-Input, optionale URL/Base-URL, Analyse-Resultat und Fehlerform sind so modelliert, dass der Go-`StreamAnalyzer`-Port nicht auf Playback-Event-Batches festgelegt bleibt (`5aa3f34`).
- [x] Go-Port-/Use-Case-Zielsignatur ist früh umgesetzt oder als Kompatibilitätsadapter vorbereitet, bevor Parser-Implementierung beginnt; der bestehende `AnalyzeBatch(ctx, []domain.PlaybackEvent) error`-Slot darf nicht bis Tranche 6 als einziger API-Integrationsvertrag stehen bleiben (`5aa3f34`).
- [x] Interne Parser-Module bleiben intern; dokumentierte Konsumenten importieren nur über den Package-Entry-Point (`5aa3f34`).
- [x] F-73 ist vorbereitet: Parser- und Ergebnisgrenzen sind so geschnitten, dass DASH-/CMAF-Analyse später als eigener Analyzer-Typ ergänzt werden kann; Nicht-HLS bleibt in `0.3.0` explizit dokumentiert out of scope (`5aa3f34`).
- [x] TypeScript-Build erzeugt ESM, CJS und Type-Definitionen oder dokumentiert bewusst, warum nur ein Format unterstützt wird (`5aa3f34`).
- [x] Unit-Test-Setup ist vorhanden und läuft über `pnpm --filter <stream-analyzer-paket> run test` (`5aa3f34`).
- [x] Coverage-Scope für `packages/stream-analyzer/src/` ist definiert; Zielschwelle ist mindestens 90 % für Statements, Lines, Functions und Branches (`5aa3f34`).
- [x] Root-Gates `make test`, `make lint`, `make coverage-gate` und `make build` beziehen das neue Workspace-Paket ein (`5aa3f34`).
- [x] `docs/user/stream-analyzer.md` beschreibt Public API, unterstützten Scope, bekannte Grenzen und Beispiele (`5aa3f34`).

---

## 3. Tranche 2 — HLS-Manifest laden und Playlist-Typ erkennen

Bezug: RAK-22, RAK-23, RAK-24; F-68, F-69, F-74, F-75.

Ziel: Der Analyzer kann HLS-Manifeste aus kontrollierten Eingaben laden und zwischen Master Playlist und Media Playlist unterscheiden.

DoD:

- [x] Analyse kann mit Manifest-Text als Input laufen; Netzwerkabruf ist ein separater, austauschbarer Input-Pfad (`6b96d4e`).
- [x] HTTP-/Fetch-Laden unterstützt Timeout, maximale Manifest-Größe und klare Fehler für Netzwerk-, Statuscode- und Content-Type-Probleme (`6b96d4e`).
- [x] URL-Laden hat verbindliche SSRF-Schutzregeln: nur `http`/`https`, keine Credentials in URLs, Redirect-Limit, gleiche Sicherheitsprüfung nach jedem Redirect, Block für localhost, private/link-local/loopback/reservierte IP-Bereiche, dokumentiertes DNS-Rebinding-Verhalten und Größenlimit auch nach Redirects (`6b96d4e`).
- [x] SSRF-Schutz ist testpflichtig: Tests decken Credentials in URLs, localhost/private/link-local/loopback/reservierte IPs, Redirect auf verbotene Ziele, Redirect-Limit, dokumentierte DNS-Rebinding-Entscheidung und Größenlimit nach Redirect ab (`6b96d4e`).
- [x] Parser erkennt HLS-Grundstruktur und lehnt nicht-HLS-Text mit einem strukturierten Fehler ab (`6b96d4e`).
- [x] Master Playlist wird anhand HLS-Tags erkannt und getestet (`6b96d4e`).
- [x] Media Playlist wird anhand HLS-Tags erkannt und getestet (`6b96d4e`).
- [x] Ambige oder gemischte Playlists liefern ein definiertes Ergebnis oder einen dokumentierten Fehler (`6b96d4e`).
- [x] Fixtures decken gültige Master Playlists, gültige Media Playlists, leere Dateien, nicht-HLS-Inhalt und malformed HLS ab (`6b96d4e`).
- [x] Der Parser arbeitet deterministisch ohne echte Netzwerkabhängigkeit in Unit-Tests (`6b96d4e`).

---

## 4. Tranche 3 — Master-Playlist-Analyse

Bezug: RAK-23, RAK-26; F-76, F-81.

Ziel: Master Playlists liefern ein strukturiertes Analyseergebnis mit Varianten und Renditions.

DoD:

- [x] Varianten aus `#EXT-X-STREAM-INF` werden extrahiert: URI, Bandbreite, optionale Auflösung, Codecs und Frame-Rate, soweit vorhanden (`18857b5`).
- [x] Renditions aus `#EXT-X-MEDIA` werden extrahiert: Typ, Group-ID, Name, Sprache, URI, Default-/Autoselect-Flags, soweit vorhanden (`18857b5`).
- [x] Relative URIs werden optional gegen eine Base-URL auflösbar gemacht; die Roh-URI bleibt im Ergebnis erhalten (`18857b5`).
- [x] Fehlende optionale Attribute führen nicht zum Analyseabbruch (`18857b5`).
- [x] Offensichtliche Inkonsistenzen werden als Findings ausgegeben, z. B. fehlende Variant-URI oder unvollständige Rendition-Referenzen (`18857b5`).
- [x] JSON-Ergebnis für Master Playlists ist snapshot- oder schema-nah getestet (`18857b5`).

---

## 5. Tranche 4 — Media-Playlist-Analyse und Segment-Dauerprüfung

Bezug: RAK-24, RAK-25, RAK-26; F-70, F-71, F-77, F-78, F-79, F-80, F-81.

Ziel: Media Playlists liefern Segmentdaten, Dauerstatistiken, einfache Inkonsistenz-Findings und eine Live-Latenz-Schätzung.

DoD:

- [x] Segmente aus `#EXTINF` werden mit URI und Dauer extrahiert (`1a9cc27`).
- [x] `#EXT-X-TARGETDURATION` wird ausgewertet; Segment-Dauer-Findings prüfen, ob gerundete Segmentdauern die Target-Duration verletzen (`1a9cc27`).
- [x] Segment-Anzahl wird bestimmt (`1a9cc27`).
- [x] Durchschnittliche Segment-Dauer wird berechnet (`1a9cc27`).
- [x] Segment-Dauerabweichungen werden gegen eine dokumentierte Toleranz geprüft (`1a9cc27`).
- [x] Findings unterscheiden mindestens `info`, `warning` und `error` (`1a9cc27`).
- [x] Live-/VOD-Erkennung ist definiert, z. B. anhand `#EXT-X-ENDLIST` (`1a9cc27`).
- [x] Einfache Live-Latenz-Schätzung ist dokumentiert und getestet (`1a9cc27`).
- [x] Malformed Segment-Einträge liefern strukturierte Findings, statt unkontrolliert zu crashen (`1a9cc27`).
- [x] JSON-Ergebnis für Media Playlists ist snapshot- oder schema-nah getestet (`1a9cc27`).

---

## 6. Tranche 5 — JSON-Ergebnisformat und Dokumentation

Bezug: RAK-26; F-81; `docs/user/stream-analyzer.md`.

Ziel: Analyzer-Ergebnisse sind stabil serialisierbar, für API und CLI geeignet und dokumentiert.

DoD:

- [x] JSON-Shape ist als TypeScript-Typ und Dokumentationsbeispiel festgelegt; das Format bleibt um weitere Analyzer-Typen wie DASH/CMAF erweiterbar (`a2c597a`).
- [x] Ergebnis enthält mindestens Analyzer-Version aus der Paketversion, Input-Metadaten, Playlist-Typ, Summary, Findings und typspezifische Details (`a2c597a`).
- [x] Fehler-JSON ist klar vom Erfolgs-JSON unterscheidbar (`a2c597a`).
- [x] Stabilitätsregel ist dokumentiert: additive Änderungen sind erlaubt; breaking Changes benötigen Changelog und Plan-/Doku-Update (`a2c597a`).
- [x] JSON-Serialisierung ist getestet und enthält keine nicht-deterministischen Felder ohne explizite Normalisierung (`a2c597a`).
- [x] `docs/user/stream-analyzer.md` enthält Beispiele für Master Playlist, Media Playlist und Fehlerfall (`a2c597a`).

---

## 7. Tranche 6 — API-Anbindung über StreamAnalyzer-Port

Bezug: RAK-27; F-22, F-33; Architektur §5/§8.

Ziel: `apps/api` kann den Analyzer nutzen, ohne die hexagonalen Grenzen zu verletzen.

DoD:

- [x] Bestehender Go-Port `hexagon/port/driven.StreamAnalyzer` wird überprüft und bei Bedarf so angepasst, dass Analyseaufrufe fachlich sinnvoll modelliert sind (`5aa3f34`, `579e7cc`).
- [x] API-Integration nutzt einen Adapter an der Driven-Seite; Domain und Application Layer importieren keine Node-/TypeScript-Implementierungsdetails (`579e7cc`).
- [x] Bevorzugter Integrationsmodus für `0.3.0` ist ein interner Analyzer-HTTP-Service, damit das distroless-Go-API-Image keinen Node-/CLI-Runtime-Stack enthalten muss. Abweichungen müssen die Runtime- und Security-Folgen im Plan dokumentieren (`579e7cc`).
- [x] Falls Node-Analyzer aus Go heraus nicht sinnvoll direkt nutzbar ist, ist der Integrationsmodus explizit entschieden: bevorzugt HTTP-intern, alternativ separater Prozess oder CLI-Adapter. API-Nutzbarkeit selbst ist release-blocking und darf nicht deferred werden (`579e7cc`).
- [x] Docker-/Compose-/Runtime-Wiring ist umgesetzt und getestet: Analyzer-Service oder Adapter ist in lokaler Entwicklung und CI verfügbar, `apps/api` kann ihn erreichen, und ein Smoke-Test prüft den Analyzer-API-Pfad im laufenden Stack (`579e7cc`).
- [x] API-Endpunkt oder Use-Case-Pfad für Analyse ist definiert und getestet (`579e7cc`).
- [x] Backend-API-Kontrakt wird synchron aktualisiert oder eine Nachfolgedoku als neue Contract-Quelle benannt; Pfad, Methode, Request-/Response-Shape, Fehlerabbildung und Pflichttests für den Analyzer-API-Pfad sind dokumentiert (`579e7cc`).
- [x] Fehlerabbildung von Analyzer-Fehlern auf HTTP-Status/Problem-Shape ist dokumentiert (`579e7cc`).
- [!] Metriken/Logs für Analyseaufrufe sind minimal vorhanden oder bewusst deferred — bewusst deferred auf 0.3.x: HTTPStreamAnalyzer und AnalyzeHandler loggen Fehler über den existierenden slog-Pfad; Prometheus-Counter/Histogram bleiben Tranche-0.3.x-Folge-Issue (kein Release-Blocker für Tranche 6).
- [x] Architekturcheck bleibt grün (`579e7cc`).

---

## 8. Tranche 7 — CLI-Grundlage

Bezug: RAK-28; MVP-34.

Ziel: Eine einfache CLI kann lokale oder per URL geladene HLS-Manifeste analysieren und JSON ausgeben.

DoD:

- [x] CLI-Einstieg ist im Analyzer-Paket definiert, z. B. `bin` in `package.json` (CLI-Hash folgt mit dem Implementierungs-Commit).
- [x] Der Lastenheft-Aufruf `pnpm m-trace check <url>` funktioniert exakt als Smoke-Kriterium (Wurzel-Skript leitet auf den gebauten CLI-Bundle, `make smoke-cli` deckt es ab).
- [x] CLI akzeptiert mindestens Datei-Pfad oder URL als Input.
- [x] CLI-URL-Input ist verpflichtend und nutzt dieselben SSRF-Schutzregeln wie der Analyzer-Loader; lokale Datei-Analyse ist zusätzlich erlaubt, ersetzt aber das URL-Ziel nicht.
- [x] CLI gibt Analyseergebnis auf stdout als JSON aus.
- [x] Fehler werden mit sinnvollem Exit-Code und maschinenlesbarem Fehler-JSON (analysis-error → stdout JSON, exit 1) oder klar dokumentiertem stderr (I/O-Fehler → stderr-Hinweis, exit 1; Usage-Fehler → stderr + exit 2) ausgegeben.
- [x] CLI-Tests decken Datei-Input, URL-/Fetch-Mock, Erfolg und Fehlerfall ab (`tests/cli.test.ts`, 16 Cases).
- [x] Lokaler Smoke-Befehl ist in `docs/user/local-development.md` dokumentiert.

---

## 9. Tranche 7.5 — Tranche-6-Folge-Issues härten

Bezug: Code-Review zu Tranche 6 (Hash `579e7cc`) und zu den Tranche-6-
Review-Fixes (`d894556`); siehe `spec/lastenheft.md` F-22/F-33 für den
Analyzer-Pfad und §13.5 für RAK-27.

Ziel: die vier nicht-release-blockierenden Punkte aus dem Tranche-6-
Review schließen, bevor `0.3.0` ins Release-Akzeptanz-Gate (Tranche 8)
geht. Damit gehen wir mit weniger Tech-Debt und besserer Operator-
Sicht in den Release-Schnitt.

DoD:

- [x] Cross-Process-Vertragstest TS↔Go: `spec/contract-fixtures/analyzer/{success-master,error-fetch-blocked}.json` als Quelle der Wahrheit; TS-Test in `packages/stream-analyzer/tests/contract.test.ts` pinnt `analyzeHlsManifest`-Output gegen Spec und prüft Drift gegen die Go-`testdata`-Kopien byte-genau; Go-Test in `apps/api/adapters/driven/streamanalyzer/contract_test.go` parst die Kopien via `go:embed` durch `parseSuccessResponse`/`parseDomainError` (`a622ae8`).
- [x] Prometheus-Counter `mtrace_analyze_requests_total{outcome,code}` ergänzt; `AnalyzeHandler` ruft ihn auf jedem Erfolgs- und Fehlerpfad inkl. der frühen Eingabe-Validierungen. Coverage-Gate grün; Label-Domäne beschränkt (`outcome` ∈ {ok, error}, `code` aus der bekannten Fehler-/Erfolgsdomäne) (`a622ae8`).
- [x] `analyzer-service` respektiert `ANALYZER_ALLOW_PRIVATE_NETWORKS=true|1|yes|on` und reicht das neue `FetchOptions.allowPrivateNetworks`-Flag pro Aufruf an den Loader weiter; Default bleibt: SSRF-IP-Block aktiv. Schema-/Credentials-/Größen-/Redirect-Regeln bleiben unangetastet; Doku in `docs/user/stream-analyzer.md` §6 ergänzt (`a622ae8`).
- [x] `apps/analyzer-service/Dockerfile` baut ohne zweiten `pnpm install`-Schritt — `pnpm deploy --prod --legacy /deploy` erzeugt ein selbsttragendes Bundle; Runtime-Stage übernimmt es per `COPY`. Image-Größe ~155 MB, kein Wachstum gegenüber 0.3.0-Baseline; `make smoke-analyzer` grün (`a622ae8`).
- [x] Plan-/Doku-Aktualisierung: §9-DoD-Items mit Hash, `CHANGELOG.md`-Unreleased-Eintrag mit Counter-Name, Flag-Name, Dockerfile-Notiz und Contract-Test-Hinweis (`a622ae8`).
- [x] `make gates` und `make smoke-analyzer` bleiben grün (`a622ae8`); CI-Run auf `main` wird nach dem Push verifiziert.

---

## 10. Tranche 8 — Release-Akzeptanzkriterien `0.3.0`

Bezug: RAK-22..RAK-28; `docs/user/releasing.md`.

DoD:

- [x] **RAK-22** HLS Manifest kann geladen werden — URL-Loader (`6b96d4e`) plus SSRF-Härtung (`5614684`).
- [x] **RAK-23** Master Playlist kann erkannt werden — Klassifikator (`6b96d4e`) und Master-Detail-Parser (`18857b5`, `42ee1d8`).
- [x] **RAK-24** Media Playlist kann erkannt werden — Klassifikator (`6b96d4e`) und Media-Detail-Parser (`1a9cc27`, `759acdc`).
- [x] **RAK-25** Segment-Dauern werden geprüft — TARGETDURATION-Verletzung und Outlier-Findings in Tranche 4 (`1a9cc27`, `759acdc`).
- [x] **RAK-26** Ergebnis wird als JSON ausgegeben — Schema-Stabilisierung mit diskriminierter Union (`a2c597a`, `ecb3cdb`).
- [x] **RAK-27** API kann Analyzer nutzen — `POST /api/analyze` über internen analyzer-service (`579e7cc`, `d894556`); Tranche-7.5-Härtung in `a622ae8`/`ba337db`.
- [x] **RAK-28** CLI-Grundlage existiert — `pnpm m-trace check <url-or-file>` (`7466c02`, `5817c62`).
- [x] Versionen sind konsistent: Root- und Workspace-Pakete tragen `0.3.0` (`@npm9912/stream-analyzer`, `@npm9912/analyzer-service`, `@npm9912/player-sdk` synchronisiert via `check-sdk-compat`); CHANGELOG-Abschnitt `[0.3.0] - 2026-05-01`; Release-Tag `v0.3.0`; im Result emittierte `analyzerVersion` wird aus `packages/stream-analyzer/package.json` abgeleitet und durch `tests/version.test.ts` und `tests/contract.test.ts` gepinnt.
- [x] `docs/user/stream-analyzer.md` (Status: 0.3.0 — veröffentlicht), `docs/user/local-development.md` (CLI- und Smoke-Beispiele), `docs/user/quality.md` (Coverage-Gate inkl. analyzer-service) und `README.md` („Enthalten seit v0.3.0", Status-Zeile) beschreiben den tatsächlichen Analyzer-Lieferstand.
- [x] `CHANGELOG.md` enthält den Versionsabschnitt `[0.3.0] - 2026-05-01` mit Auflistung aller Tranche-1..7.5-Liefermerkmale.
- [x] Release-Gates grün: `make gates` (deckt `test` + `lint` + `coverage-gate` + `arch-check` ab), `make sdk-performance-smoke`, `make build`, `make smoke-cli` (CLI-Smoke gegen lokales Fixture und SSRF-Negativ-URL).
- [x] `make browser-e2e` ist nicht durch Analyzer-Änderungen betroffen (kein Touch an Player-SDK-Wire-Format, Dashboard-Routen, MediaMTX-Topologie); bleibt manuelles Release-Gate gemäß DoD-Wortlaut.
- [x] Release-Prozess aus `docs/user/releasing.md` ist durchgeführt: Release-Commit `c8d4238` existiert, annotierter Tag `v0.3.0` zeigt darauf, GitHub-Release wird nach `git push` mit `gh release create` erzeugt (Release-Notes aus dem `[0.3.0]`-CHANGELOG-Block).
- [x] OE-3/Persistenz ist explizit nicht-blockierend deferred: `0.3.0` speichert Analyseergebnisse nicht durable; jeder `POST /api/analyze`-Aufruf ist stateless. Damit bleibt der ADR-0002-Draft (`docs/adr/0002-persistence-store.md`) für `0.4.0` offen, ohne `0.3.0` zu blockieren (siehe Roadmap §5 OE-3).

---

## 11. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`, Commit-Hash anhängen.
- Neue Findings in der `0.3.0`-Phase landen entweder in dieser Datei oder in `risks-backlog.md`.
- Lastenheft-Patches während `0.3.0` werden in `plan-0.1.0.md` Tranche 0c als neue §4a.x-Einträge ergänzt.
- Beim Release-Bump `0.3.0` → `0.4.0`: dieses Dokument als historisch archivieren; Lieferstand wandert dokumentarisch nach `CHANGELOG.md`.
