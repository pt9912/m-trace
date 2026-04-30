# Implementation Plan — `0.3.0` (Stream Analyzer)

> **Status**: 🟡 in Arbeit. Beginnt nach Release `v0.2.0` (Publizierbares Player SDK).  
> **Bezug**: [Lastenheft `1.1.7`](./lastenheft.md) §7.7 (Stream Analyzer), §13.5 (RAK-22..RAK-28), §18 (MVP-DoD-Anteil); [Roadmap](./roadmap.md) §2/§3; [Architektur (Zielbild)](./architecture.md); [API-Kontrakt](./spike/backend-api-contract.md); [Risiken-Backlog](./risks-backlog.md); [ADR 0002 Persistenz-Store](./adr/0002-persistence-store.md).
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
| 1 | Stream-Analyzer-Paket und Public API | ⬜ |
| 2 | HLS-Manifest laden und Playlist-Typ erkennen | ⬜ |
| 3 | Master-Playlist-Analyse | ⬜ |
| 4 | Media-Playlist-Analyse und Segment-Dauerprüfung | ⬜ |
| 5 | JSON-Ergebnisformat und Dokumentation | ⬜ |
| 6 | API-Anbindung über StreamAnalyzer-Port | ⬜ |
| 7 | CLI-Grundlage | ⬜ |
| 8 | Release-Akzeptanzkriterien `0.3.0` | ⬜ |

---

## 1a. Tranche 0 — Vorgänger-Gate-Verifikation

Konvertiert die narrative Vorgänger-Gate-Beschreibung aus §0 in prüfbare DoD-Items.

DoD:

- [x] `plan-0.2.0.md` ist vollständig (`[x]`), inklusive Release-Akzeptanzkriterien (`a2cb66d`).
- [x] Annotierter Release-Tag `v0.2.0` existiert und zeigt auf den Release-Commit (`a2cb66d`).
- [x] GitHub-Release `v0.2.0` existiert und verweist auf die Changelog-Release-Notes (`a2cb66d`).
- [x] GitHub Actions `Build` ist für den Release-Commit grün (`a2cb66d`).
- [x] `docs/roadmap.md` §2 markiert Schritt 19 als abgeschlossen und führt die Folgearbeiten für `0.3.0` (`070b231`).
- [x] Persistenzstatus ist geprüft: ADR 0002 bleibt nicht-blockierend, solange Analyseergebnisse nicht durable gespeichert werden (`37f20ce`).

---

## 2. Tranche 1 — Stream-Analyzer-Paket und Public API

Bezug: Lastenheft §7.7; RAK-22..RAK-26; MVP-33.

Ziel: `packages/stream-analyzer` wird als TypeScript-Paket mit stabiler Public API angelegt. Das Paket ist für Backend-Integration und CLI nutzbar, ohne Dashboard- oder API-Interna zu importieren.

DoD:

- [ ] Workspace-Paket `packages/stream-analyzer` ist angelegt und in `pnpm-workspace.yaml` enthalten.
- [ ] `package.json` enthält eindeutige Paketmetadaten, `type`, `main`, `module`, `types`, `exports`, `files`, `license`, `repository.directory` und CI-taugliche Scripts.
- [ ] Public API exportiert mindestens eine Analysefunktion für HLS-Manifeste und klar typisierte Ergebnis-/Fehlertypen.
- [ ] Interne Parser-Module bleiben intern; dokumentierte Konsumenten importieren nur über den Package-Entry-Point.
- [ ] TypeScript-Build erzeugt ESM, CJS und Type-Definitionen oder dokumentiert bewusst, warum nur ein Format unterstützt wird.
- [ ] Unit-Test-Setup ist vorhanden und läuft über `pnpm --filter <stream-analyzer-paket> run test`.
- [ ] Coverage-Scope für `packages/stream-analyzer/src/` ist definiert; Zielschwelle ist mindestens 90 % für Statements, Lines, Functions und Branches.
- [ ] Root-Gates `make test`, `make lint`, `make coverage-gate` und `make build` beziehen das neue Workspace-Paket ein.
- [ ] `docs/stream-analyzer.md` beschreibt Public API, unterstützten Scope, bekannte Grenzen und Beispiele.

---

## 3. Tranche 2 — HLS-Manifest laden und Playlist-Typ erkennen

Bezug: RAK-22, RAK-23, RAK-24; F-68, F-69, F-74, F-75.

Ziel: Der Analyzer kann HLS-Manifeste aus kontrollierten Eingaben laden und zwischen Master Playlist und Media Playlist unterscheiden.

DoD:

- [ ] Analyse kann mit Manifest-Text als Input laufen; Netzwerkabruf ist ein separater, austauschbarer Input-Pfad.
- [ ] HTTP-/Fetch-Laden unterstützt Timeout, maximale Manifest-Größe und klare Fehler für Netzwerk-, Statuscode- und Content-Type-Probleme.
- [ ] Parser erkennt HLS-Grundstruktur und lehnt nicht-HLS-Text mit einem strukturierten Fehler ab.
- [ ] Master Playlist wird anhand HLS-Tags erkannt und getestet.
- [ ] Media Playlist wird anhand HLS-Tags erkannt und getestet.
- [ ] Ambige oder gemischte Playlists liefern ein definiertes Ergebnis oder einen dokumentierten Fehler.
- [ ] Fixtures decken gültige Master Playlists, gültige Media Playlists, leere Dateien, nicht-HLS-Inhalt und malformed HLS ab.
- [ ] Der Parser arbeitet deterministisch ohne echte Netzwerkabhängigkeit in Unit-Tests.

---

## 4. Tranche 3 — Master-Playlist-Analyse

Bezug: RAK-23, RAK-26; F-76, F-81.

Ziel: Master Playlists liefern ein strukturiertes Analyseergebnis mit Varianten und Renditions.

DoD:

- [ ] Varianten aus `#EXT-X-STREAM-INF` werden extrahiert: URI, Bandbreite, optionale Auflösung, Codecs und Frame-Rate, soweit vorhanden.
- [ ] Renditions aus `#EXT-X-MEDIA` werden extrahiert: Typ, Group-ID, Name, Sprache, URI, Default-/Autoselect-Flags, soweit vorhanden.
- [ ] Relative URIs werden optional gegen eine Base-URL auflösbar gemacht; die Roh-URI bleibt im Ergebnis erhalten.
- [ ] Fehlende optionale Attribute führen nicht zum Analyseabbruch.
- [ ] Offensichtliche Inkonsistenzen werden als Findings ausgegeben, z. B. fehlende Variant-URI oder unvollständige Rendition-Referenzen.
- [ ] JSON-Ergebnis für Master Playlists ist snapshot- oder schema-nah getestet.

---

## 5. Tranche 4 — Media-Playlist-Analyse und Segment-Dauerprüfung

Bezug: RAK-24, RAK-25, RAK-26; F-70, F-71, F-77, F-78, F-79, F-80, F-81.

Ziel: Media Playlists liefern Segmentdaten, Dauerstatistiken, einfache Inkonsistenz-Findings und eine Live-Latenz-Schätzung.

DoD:

- [ ] Segmente aus `#EXTINF` werden mit URI und Dauer extrahiert.
- [ ] Segment-Anzahl wird bestimmt.
- [ ] Durchschnittliche Segment-Dauer wird berechnet.
- [ ] Segment-Dauerabweichungen werden gegen eine dokumentierte Toleranz geprüft.
- [ ] Findings unterscheiden mindestens `info`, `warning` und `error`.
- [ ] Live-/VOD-Erkennung ist definiert, z. B. anhand `#EXT-X-ENDLIST`.
- [ ] Einfache Live-Latenz-Schätzung ist dokumentiert und getestet oder bewusst als nicht erfülltes RAK-/F-Item markiert.
- [ ] Malformed Segment-Einträge liefern strukturierte Findings, statt unkontrolliert zu crashen.
- [ ] JSON-Ergebnis für Media Playlists ist snapshot- oder schema-nah getestet.

---

## 6. Tranche 5 — JSON-Ergebnisformat und Dokumentation

Bezug: RAK-26; F-81; `docs/stream-analyzer.md`.

Ziel: Analyzer-Ergebnisse sind stabil serialisierbar, für API und CLI geeignet und dokumentiert.

DoD:

- [ ] JSON-Shape ist als TypeScript-Typ und Dokumentationsbeispiel festgelegt.
- [ ] Ergebnis enthält mindestens Analyzer-Version, Input-Metadaten, Playlist-Typ, Summary, Findings und typspezifische Details.
- [ ] Fehler-JSON ist klar vom Erfolgs-JSON unterscheidbar.
- [ ] Stabilitätsregel ist dokumentiert: additive Änderungen sind erlaubt; breaking Changes benötigen Changelog und Plan-/Doku-Update.
- [ ] JSON-Serialisierung ist getestet und enthält keine nicht-deterministischen Felder ohne explizite Normalisierung.
- [ ] `docs/stream-analyzer.md` enthält Beispiele für Master Playlist, Media Playlist und Fehlerfall.

---

## 7. Tranche 6 — API-Anbindung über StreamAnalyzer-Port

Bezug: RAK-27; F-22, F-33; Architektur §5/§8.

Ziel: `apps/api` kann den Analyzer nutzen, ohne die hexagonalen Grenzen zu verletzen.

DoD:

- [ ] Bestehender Go-Port `hexagon/port/driven.StreamAnalyzer` wird überprüft und bei Bedarf so angepasst, dass Analyseaufrufe fachlich sinnvoll modelliert sind.
- [ ] API-Integration nutzt einen Adapter an der Driven-Seite; Domain und Application Layer importieren keine Node-/TypeScript-Implementierungsdetails.
- [ ] Falls Node-Analyzer aus Go heraus nicht sinnvoll direkt nutzbar ist, ist der Integrationsmodus explizit entschieden: CLI-Adapter, separater Prozess, HTTP-intern oder bewusst deferred.
- [ ] API-Endpunkt oder Use-Case-Pfad für Analyse ist definiert und getestet.
- [ ] Fehlerabbildung von Analyzer-Fehlern auf HTTP-Status/Problem-Shape ist dokumentiert.
- [ ] Metriken/Logs für Analyseaufrufe sind minimal vorhanden oder bewusst deferred.
- [ ] Architekturcheck bleibt grün.

---

## 8. Tranche 7 — CLI-Grundlage

Bezug: RAK-28; MVP-34.

Ziel: Eine einfache CLI kann lokale oder per URL geladene HLS-Manifeste analysieren und JSON ausgeben.

DoD:

- [ ] CLI-Einstieg ist im Analyzer-Paket definiert, z. B. `bin` in `package.json`.
- [ ] CLI akzeptiert mindestens Datei-Pfad oder URL als Input.
- [ ] CLI gibt Analyseergebnis auf stdout als JSON aus.
- [ ] Fehler werden mit sinnvollem Exit-Code und maschinenlesbarem Fehler-JSON oder klar dokumentiertem stderr ausgegeben.
- [ ] CLI-Tests decken Datei-Input, URL-/Fetch-Mock, Erfolg und Fehlerfall ab.
- [ ] Lokaler Smoke-Befehl ist in `docs/local-development.md` dokumentiert.

---

## 9. Tranche 8 — Release-Akzeptanzkriterien `0.3.0`

Bezug: RAK-22..RAK-28; `docs/releasing.md`.

DoD:

- [ ] **RAK-22** HLS Manifest kann geladen werden.
- [ ] **RAK-23** Master Playlist kann erkannt werden.
- [ ] **RAK-24** Media Playlist kann erkannt werden.
- [ ] **RAK-25** Segment-Dauern werden geprüft.
- [ ] **RAK-26** Ergebnis wird als JSON ausgegeben.
- [ ] **RAK-27** API kann Analyzer nutzen.
- [ ] **RAK-28** CLI-Grundlage existiert.
- [ ] `docs/stream-analyzer.md`, `docs/local-development.md`, `docs/quality.md` und `README.md` beschreiben den tatsächlichen Analyzer-Lieferstand.
- [ ] `CHANGELOG.md` enthält Eintrag für `0.3.0`.
- [ ] Release-Gates laufen grün: `make test`, `make lint`, `make coverage-gate`, `make arch-check`, `make build` und Analyzer-spezifische Smoke-/CLI-Gates, sofern eingeführt.
- [ ] Falls `make browser-e2e` nicht durch Analyzer-Änderungen betroffen ist, bleibt es ein manuelles Release-Gate; bei Dashboard-/Demo-Auswirkungen läuft es grün.
- [ ] Release-Prozess aus `docs/releasing.md` ist durchgeführt: Release-Commit existiert, annotierter Tag `v0.3.0` ist erstellt und das Release-Artefakt ist nachvollziehbar.
- [ ] OE-3/Persistenz ist entschieden oder explizit nicht-blockierend deferred, falls Analyseergebnisse nicht durable gespeichert werden.

---

## 10. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`, Commit-Hash anhängen.
- Neue Findings in der `0.3.0`-Phase landen entweder in dieser Datei oder in `risks-backlog.md`.
- Lastenheft-Patches während `0.3.0` werden in `plan-0.1.0.md` Tranche 0c als neue §4a.x-Einträge ergänzt.
- Beim Release-Bump `0.3.0` → `0.4.0`: dieses Dokument als historisch archivieren; Lieferstand wandert dokumentarisch nach `CHANGELOG.md`.
