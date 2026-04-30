# Implementation Plan — `0.2.0` (Publizierbares Player SDK)

> **Status**: 🟡 in Arbeit. Beginnt nach Abschluss von `0.1.2` (Observability-Stack).  
> **Bezug**: [Lastenheft `1.1.7`](./lastenheft.md) §13.4 (RAK-11..RAK-21), §18 (MVP-DoD-Anteil); [Roadmap](./roadmap.md) §2/§3; [Architektur (Zielbild)](./architecture.md); [Telemetry-Modell](./telemetry-model.md); [API-Kontrakt](./spike/backend-api-contract.md); [Risiken-Backlog](./risks-backlog.md).
> **Vorgänger-Gate (Stand zum `0.2.0`-Start)**:
>
> - [`plan-0.1.2.md`](./plan-0.1.2.md) muss vollständig (`[x]`) sein, inklusive Release-Akzeptanzkriterien `0.1.2` (§4).
> - [`plan-0.1.1.md`](./plan-0.1.1.md) Tranche 1 (Player-SDK) muss vollständig (`[x]`) sein; `0.2.0` stabilisiert dieses Paket, statt ein zweites SDK anzulegen.
> - `v0.1.2` muss als Release-Tag existieren; offene neue Arbeit beginnt danach in diesem Plan.
>
> **Nachfolger**: `plan-0.3.0.md` (Stream Analyzer).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog [`plan-0.1.0.md`](./plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz (siehe `roadmap.md` §7.1).
- 🟡 in Arbeit.

Neue Lastenheft-Patches während `0.2.0` landen weiterhin zentral in `plan-0.1.0.md` Tranche 0c, weil sie projektweit gelten.

---

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
|---|---|---|
| 0 | Vorgänger-Gate-Verifikation | ✅ |
| 1 | SDK-Paketierung und Public API | ✅ |
| 2 | Event-Schema-Versionierung und CI-Kompatibilitätscheck | ✅ |
| 3 | Adapter-/Transport-Tests und Runtime-Grenzen | ⬜ |
| 3a | Node-Coverage-Gates für Player-SDK; Dashboard-Entscheidung | ⬜ |
| 4 | OTel-Transport-Option, Performance-Budget und Browser-Matrix | ⬜ |
| 5 | Demo-Integrationsdoku und Release-Akzeptanzkriterien `0.2.0` | ⬜ |
| 6 | OE-3/Persistenz-Folge-ADR vorbereiten | ⬜ |

---

## 1a. Tranche 0 — Vorgänger-Gate-Verifikation

Konvertiert die narrative Vorgänger-Gate-Beschreibung aus §0 in prüfbare DoD-Items.

DoD:

- [x] `plan-0.1.2.md` Tranche 1 (Observability Pflicht-Anteile) abgeschlossen (`beabcd5`, `7acda47`).
- [x] `plan-0.1.2.md` Tranche 2 (observability-Compose-Profil) abgeschlossen (`beabcd5`).
- [x] `plan-0.1.2.md` Tranche 3 (RAK-9, RAK-10, übergreifende DoD) abgeschlossen (`79be435`, `7acda47`, `3d3a25e`).
- [x] `CHANGELOG.md` enthält einen datierten Eintrag für `0.1.2` (`c76457c`).
- [x] Annotierter Release-Tag `v0.1.2` existiert (`c76457c`).
- [x] `plan-0.1.1.md` Tranche 1 beschreibt den bestehenden Player-SDK-Lieferstand vollständig genug, um `0.2.0` darauf aufzubauen (`bae4a2a`, `cf07fda`, `55ccac4`).

---

## 2. Tranche 1 — SDK-Paketierung und Public API

Bezug: RAK-11, RAK-12.

Ziel: `packages/player-sdk` wird von einem Workspace-Paket zu einem lokal installierbaren, klar dokumentierten npm-Paket mit stabiler Public API.

DoD:

- [x] Paketnamen-Migration gemäß [`docs/migrate-package-name.md`](./migrate-package-name.md) durchgeführt: OE-8 neu entschieden (`@npm9912/player-sdk` ab `0.2.0`), alle unter §2.1 gelisteten Stellen umgestellt, Such-Gate aus `docs/migrate-package-name.md` §4 liefert nur noch die dort erlaubten historischen Treffer (`d367720`).
- [x] `packages/player-sdk/package.json` enthält publish-fähige Metadaten: `name`, `version`, `description`, `license`, `repository`, `files`, `exports`, `types`, ESM/CJS/Browser-Build-Einstiege (`819ee17`, `d367720`).
- [x] `packages/player-sdk/package.json` setzt `version` auf `0.2.0`; Pack-/Install-Smoke-Test prüft, dass das erzeugte npm-Paket ebenfalls Version `0.2.0` meldet (`819ee17`).
- [x] Root-`package.json` setzt `version` auf `0.2.0`; Repo-Release-Tag, Root-Metadaten und SDK-Paketversion sind konsistent oder eine bewusst abweichende Versionierungsregel ist dokumentiert (`819ee17`).
- [x] `packages/player-sdk/package.json` ist nicht mehr als privates Paket blockiert: `private` ist entfernt oder `false`; bei öffentlichem Scoped Package ist `publishConfig.access` passend gesetzt (`819ee17`).
- [x] `pnpm --filter @npm9912/player-sdk pack` erzeugt ein installierbares Tarball-Artefakt (`819ee17`, `d367720`).
- [x] Publish-Dry-Run läuft ohne Paketierungsfehler, z. B. `pnpm --filter @npm9912/player-sdk publish --dry-run` oder ein äquivalenter `npm publish --dry-run` gegen das erzeugte Tarball (`819ee17`, `d367720`).
- [x] Lokaler Install-Smoke-Test installiert das gepackte SDK in ein temporäres Beispielprojekt und importiert ESM sowie CJS erfolgreich (`819ee17`, `d367720`).
- [x] Tarball-Content-Check prüft, dass ESM-, CJS-, Type-Definition- und Browser/IIFE-Build-Artefakte im gepackten Paket enthalten sind (`819ee17`).
- [x] Browser-Einstieg ist stabil auffindbar und verifiziert: `package.json` definiert entweder `browser`, einen Conditional Export oder einen dokumentierten CDN-/IIFE-Pfad, und der Smoke-Test nutzt genau diesen Einstieg (`819ee17`).
- [x] Browser-Load-Smoke lädt den Browser/IIFE-Build aus dem installierten Paket und prüft, dass der globale SDK-Einstieg nutzbar ist (`819ee17`).
- [x] Browser-Bundle bleibt als expliziter Build-Ausgang vorhanden und ist in der Paketdoku beschrieben (`819ee17`).
- [x] Public API wird in `packages/player-sdk/README.md` dokumentiert: Einstieg, Tracker-Lifecycle, Konfiguration, Events, hls.js-Adapter, Transport-Optionen, Fehlerverhalten (`819ee17`, `d367720`).
- [x] Projektdokument `docs/player-sdk.md` wird angelegt und beschreibt die Player-SDK-Nutzung gemäß Lastenheft §7.6; damit wird der SDK-Anteil der Pflichtdokument-Liste aus §7.12 erfüllt. `packages/player-sdk/README.md` darf darauf verweisen, ersetzt es aber nicht (`819ee17`, `d367720`).
- [x] Public API-Surface wird technisch abgesichert, z. B. durch TypeScript-API-Snapshot oder Export-Snapshot-Test (`819ee17`, `d367720`).
- [x] Interne Module bleiben intern: dokumentierte Imports laufen nur über den Package-Entry-Point, nicht über tiefe Pfade (`819ee17`).

---

## 3. Tranche 2 — Event-Schema-Versionierung und CI-Kompatibilitätscheck

Bezug: RAK-13, RAK-21; `docs/telemetry-model.md`.

Ziel: SDK-Version und Event-Schema-Version sind explizit gekoppelt, damit API, SDK und Dokumentation bei Schema-Änderungen nicht auseinanderlaufen.

DoD:

- [x] Event-Schema-Version wird im SDK als stabile Konstante exportiert (`d367720`).
- [x] SDK-Version wird aus genau einer Quelle abgeleitet: `packages/player-sdk/package.json`, exportierte SDK-Version und die im Wire-Event gesendete `sdk.version` sind synchron (`d367720`).
- [x] SDK sendet die Schema-Version im definierten Wire-Format aus `docs/telemetry-model.md` (`d367720`).
- [x] SDK-Tests verhindern Regressionen bei `sdk.version` und Schema-Version im erzeugten Event-Payload (`d367720`).
- [x] `docs/telemetry-model.md` beschreibt die aktuelle Schema-Version, Kompatibilitätsregeln und erlaubte additive/breaking Changes (`d367720`).
- [x] Lebende Doku referenziert die Contract-Artefakte oder ist mit ihnen synchronisiert; historische Spike-Snapshots bleiben gemäß [`docs/migrate-package-name.md`](./migrate-package-name.md) §2.2 unverändert (`d367720`).
- [x] API-Tests prüfen, dass die aktuell unterstützte Schema-Version akzeptiert wird (`d367720`).
- [x] SDK-Tests prüfen, dass jeder erzeugte Batch die aktuelle Schema-Version trägt (`d367720`).
- [x] Maschinenlesbare Contract-Artefakte werden angelegt, z. B. `contracts/event-schema.json` und `contracts/sdk-compat.json`; sie sind Source of Truth für Schema-Version und SDK↔Schema-Kompatibilität (`d367720`).
- [x] Contract-Ownership und Update-Regel sind dokumentiert: Änderungen an Schema-Version, SDK-Version oder API-`SupportedSchemaVersion` müssen die Contract-Artefakte im selben Commit aktualisieren (`d367720`).
- [x] CI-Kompatibilitätscheck vergleicht SDK-Konstante, Contract-Artefakte und API-Erwartung; technische API-Quelle ist `SupportedSchemaVersion` in `apps/api/hexagon/application/register_playback_event_batch.go` (`d367720`).
- [x] Markdown-Doku (`docs/telemetry-model.md`, `docs/player-sdk.md`) referenziert die Contract-Artefakte; CI parst nicht freitextliche Markdown-Prosa als Source of Truth (`d367720`).
- [x] RAK-21 wird über die Contract-Artefakte geprüft: `packages/player-sdk/package.json.version`, exportierte SDK-Version, ausgesendetes `sdk.version`, ausgesendete `schema_version` und API-`SupportedSchemaVersion` müssen gemeinsam konsistent sein (`d367720`).
- [x] CI-Kompatibilitätscheck ist an ein verbindliches Gate angebunden: entweder Bestandteil von `make test`/`make lint` oder eigenes Root-Target mit explizitem GitHub-Actions-Step (`d367720`).
- [x] `CHANGELOG.md`-Konvention für Event-Schema-Änderungen ist dokumentiert (`d367720`).

---

## 4. Tranche 3 — Adapter-/Transport-Tests und Runtime-Grenzen

Bezug: RAK-14, RAK-15, RAK-17.

Ziel: hls.js-Adapter, HTTP-Transport sowie Batching/Sampling/Retry-Grenzen sind eigenständig testbar und entsprechen den API-Limits.

DoD:

- [ ] hls.js-Adapter-Tests decken Manifest-/Level-/Fragment-/Error-/Rebuffer-nahe Ereignisse ab.
- [ ] hls.js-Adapter-Tests nutzen kontrollierte Test-Doubles oder eine reproduzierbare Browser-Testumgebung; reale Netzinstabilität ist kein Test-Orakel.
- [ ] HTTP-Transport-Tests decken erfolgreiche Sends, 4xx/5xx, Timeout/Abort und Netzwerkfehler ab.
- [ ] HTTP-Transport-Verhalten für `429 Too Many Requests` ist explizit festgelegt und getestet: `Retry-After` aus dem API-Kontrakt wird als Cooldown respektiert; vor Ablauf wird nicht erneut gesendet. Fehlt `Retry-After`, greift die dokumentierte Backoff-/Drop-Regel.
- [ ] Retry-Mindestverhalten ist festgelegt und getestet: Netzwerkfehler, Timeout und `5xx` sind retrybar; `400`, `401`, `403` und andere nicht-transiente `4xx` werden nicht retried; `429` folgt der gesonderten `Retry-After`-Regel.
- [ ] Retry-Grenzen sind dokumentiert und technisch erzwungen: maximale Versuche, Backoff-Regel mit Obergrenze, Abbruch-/Drop-Verhalten und Queue-Limit.
- [ ] Batching-Grenze bleibt hart bei maximal 100 Events pro Request; größere Queues werden gesplittet.
- [ ] Payload-Byte-Grenze aus dem API-Kontrakt ist technisch erzwungen: SDK-Batches bleiben unter 256 KB Request-Body; große Queues werden nach Event-Anzahl und Byte-Größe gesplittet oder dokumentiert gedroppt.
- [ ] `413 Payload Too Large` ist getestet und wird nicht blind retried; Drop-/Fehlerverhalten ist dokumentiert.
- [ ] Sampling-Semantik ist deterministisch testbar und dokumentiert: Sampling-Einheit (Event oder Session), Behandlung von `session_ended` und Zählweise von `sequence_number` bei verworfenen Events sind eindeutig festgelegt.
- [ ] `destroy()`/Session-Ende-Verhalten bleibt getestet, inklusive Flush- und Drop-Fällen.
- [ ] Root-Gates `make test` und `make lint` decken die neuen SDK-Tests ab.

### 4a. Arbeitspaket 3a — Coverage-Konfiguration für Node-Workspaces

Bezug: `docs/quality.md` §3; RAK-14, RAK-15, RAK-17.

Ziel: Die in `0.2.0` stabilisierten Node-Workspaces bekommen
reproduzierbare Coverage-Gates. Dashboard-Coverage wird nicht
implizit behauptet, sondern als eigener Scope unter `apps/dashboard/src/`
behandelt.

DoD:

- [ ] Coverage-Provider für Vitest ist im Workspace festgelegt und
  reproduzierbar gepinnt, z. B. `@vitest/coverage-v8` passend zur
  Vitest-Version.
- [ ] `packages/player-sdk/package.json` enthält ein Coverage-Script,
  z. B. `test:coverage`, das ohne Watch-Modus läuft und CI-taugliche
  Artefakte erzeugt.
- [ ] Player-SDK-Coverage-Scope ist auf produktiven Code unter
  `packages/player-sdk/src/` begrenzt; `tests/`, `dist/` und
  `scripts/` sind ausgeschlossen.
- [ ] Player-SDK-Threshold ist verbindlich definiert und begründet.
  Der Startwert darf niedriger als das API-Ziel sein, muss aber eine
  Erhöhungsperspektive enthalten; Senkungen nach Einführung sind wie
  beim API-Gate begründungspflichtig.
- [ ] Root-Target-Strategie ist entschieden und umgesetzt: entweder
  `make coverage-gate` umfasst API plus Player-SDK, oder ein eigenes
  Node-Coverage-Target wird angelegt und in CI/Release-Gates explizit
  genannt. Stillschweigende API-only-Semantik ist nicht zulässig.
- [ ] `.github/workflows/build.yml` führt das neue Coverage-Gate aus
  oder dokumentiert bewusst, warum es bis zum Folge-Release lokal
  bleibt.
- [ ] `docs/quality.md`, `docs/local-development.md` und
  `docs/releasing.md` beschreiben die tatsächlichen Coverage-Kommandos
  und Artefakte für `packages/player-sdk`.
- [ ] Für `apps/dashboard` ist eine Entscheidung dokumentiert:
  entweder Unit-/Component-Test-Setup plus Coverage-Scope
  `apps/dashboard/src/` wird eingeführt, oder Dashboard-Coverage wird
  mit Begründung und Folge-Release explizit deferred.

---

## 5. Tranche 4 — OTel-Transport-Option, Performance-Budget und Browser-Matrix

Bezug: RAK-16, RAK-18, RAK-19.

Ziel: Das SDK hat klare Laufzeitgrenzen, eine dokumentierte Browser-Support-Matrix und eine belastbare Entscheidung zum experimentellen OTel-Transport.

DoD:

- [ ] OTel-Transport-Option erfüllt RAK-16: vorbereitet oder experimentell nutzbar. Da RAK-16 ein Soll-Kriterium ist, darf es nur über das unten definierte Deferral-DoD nicht-blockierend verschoben werden.
- [ ] „Vorbereitet" ist konkret nachweisbar: stabiler Transport-Port/Interface, dokumentierter Opt-in-Konfigurationspfad, keine Pflicht-Abhängigkeit im Default-Bundle und ein Test/Compile-Check, der die spätere OTel-Transport-Implementierung anschließbar hält.
- [ ] Falls experimentell nutzbar: OTel-Transport hängt nicht am Default-Bundle und ist als separater opt-in Pfad dokumentiert.
- [ ] Falls RAK-16 nicht mindestens als „vorbereitet" erfüllt wird: nicht-blockierendes Deferral-DoD ist vollständig erfüllt — Begründung dokumentiert, Folge-Release benannt, `risks-backlog.md` oder Roadmap aktualisiert, Release Notes nennen RAK-16 als nicht enthaltenes Soll-Kriterium.
- [ ] Performance-Budget übernimmt die normativen Lastenheft-Grenzen: SDK-Bundle < 30 KB gzip ohne hls.js, < 5 ms Verarbeitungszeit pro Event, kein synchrones Netzwerk im Hot Path, Playback darf bei Telemetriefehlern nicht abbrechen.
- [ ] Performance-Smoke misst mindestens Bundle-Größe, synthetische Event-Verarbeitungszeit und Queue-/Retry-Grenzen reproduzierbar.
- [ ] Browser-Support-Matrix ist dokumentiert und übernimmt mindestens die Lastenheft-Einstufung: Chrome Desktop `supported`, Firefox Desktop `supported`, Safari Desktop `documented limitation`; weitere Browser werden als `supported`, `documented limitation` oder `out of scope` klassifiziert.
- [ ] Browser-E2E-Gates bleiben für alle als `supported` markierten Browser grün; Browser mit `documented limitation` brauchen eine begründete Einschränkung, aber kein vollständiges Gate.

---

## 6. Tranche 5 — Demo-Integrationsdoku und Release-Akzeptanzkriterien `0.2.0`

Bezug: RAK-11..RAK-21; RAK-20.

DoD:

- [ ] Dashboard-Route `/demo` ist als Beispielintegration dokumentiert: SDK-Konfiguration, API-URL, Token/Project, hls.js-Anbindung, erwartete Events.
- [ ] `README.md` verlinkt auf die SDK-Doku und beschreibt den lokalen SDK-Install-/Demo-Pfad.
- [ ] `docs/local-development.md` beschreibt, wie das gepackte SDK lokal gegen Dashboard/API getestet wird.
- [ ] **RAK-11** SDK ist als npm-Paket baubar und lokal installierbar; Paketversion ist `0.2.0`.
- [ ] **RAK-12** Public API ist dokumentiert.
- [ ] **RAK-13** Event-Schema ist versioniert.
- [ ] **RAK-14** hls.js-Adapter ist getestet.
- [ ] **RAK-15** HTTP-Transport ist getestet.
- [ ] **RAK-16 (Soll)** OTel-Transport ist vorbereitet oder experimentell nutzbar; alternativ ist das nicht-blockierende Deferral-DoD aus Tranche 4 erfüllt.
- [ ] **RAK-17** SDK unterstützt Batching, Sampling und Retry-Grenzen.
- [ ] **RAK-18** SDK hält das definierte Performance-Budget ein.
- [ ] **RAK-19** Browser-Support-Matrix ist dokumentiert.
- [ ] **RAK-20** Beispielintegration in der Dashboard-Route `/demo` ist dokumentiert.
- [ ] **RAK-21** Kompatibilität zwischen SDK-Version und Event-Schema wird in CI geprüft; der Check nutzt ein maschinenlesbares Compat-Fixture oder einen Snapshot mit SDK-Version, `sdk.version`, `schema_version` und API-`SupportedSchemaVersion`.
- [ ] `CHANGELOG.md` enthält Eintrag für `0.2.0`.
- [ ] Release-Prozess aus `docs/releasing.md` ist durchgeführt: Release-Commit existiert, annotierter Tag `v0.2.0` ist erstellt und das Release-Artefakt ist nachvollziehbar.
- [ ] Tranche 6 ist abgeschlossen oder explizit nicht-blockierend deferred: OE-3/Persistenz-Vorbereitung ist entweder dokumentiert oder mit Begründung, Folge-Release und Roadmap-/Risiken-Verweis verschoben.
- [ ] Arbeitspaket 3a ist abgeschlossen oder explizit deferred; bei Deferral ist dokumentiert, welche Coverage-Gates `0.2.0` tatsächlich erzwingt.
- [ ] Release-Gates laufen grün: `make test`, `make lint`, `make coverage-gate`, `make arch-check`, `make build`, `make browser-e2e` sowie ein eventuell separates Node-Coverage-Target aus Arbeitspaket 3a.
- [ ] CI-Gate-Abgrenzung ist dokumentiert: `make test`, `make lint`, `make coverage-gate`, `make arch-check` und `make build` laufen in GitHub Actions; der SDK↔Schema-Kompatibilitätscheck und ein eventuell separates Node-Coverage-Gate hängen an einem dieser Gates oder an eigenen Actions-Steps; `make browser-e2e` bleibt entweder bewusst manuelles Release-Gate oder wird als eigener Actions-Job ergänzt.

---

## 7. Tranche 6 — OE-3/Persistenz-Folge-ADR vorbereiten

Bezug: Roadmap §4/§5, Architektur §11; OE-3, MVP-16.

Diese Tranche ist vorbereitend für die nächste Persistenzentscheidung. Sie blockiert `0.2.0` nur, wenn bei der SDK-Stabilisierung ein API-/Storage-Vertrag geändert werden muss.

DoD:

- [ ] Entscheidungsvorlage für In-Memory → SQLite/PostgreSQL erstellt oder als Folge-ADR-Entwurf begonnen.
- [ ] Persistenz-Anforderungen aus SDK-/Schema-Sicht gesammelt: Event-Schema-Version, Cursor-Stabilität, Session-Ende, Retention.
- [ ] Roadmap/Risiken-Backlog verweisen auf die konkrete ADR-Arbeit.
- [ ] Falls kein ADR-Abschluss in `0.2.0` erfolgt: nicht-blockierende Begründung im Plan dokumentiert.

---

## 8. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`, Commit-Hash anhängen.
- Neue Findings in der `0.2.0`-Phase landen entweder in dieser Datei oder in `risks-backlog.md`.
- Lastenheft-Patches während `0.2.0` werden in `plan-0.1.0.md` Tranche 0c als neue §4a.x-Einträge ergänzt.
- Beim Release-Bump `0.2.0` → `0.3.0`: dieses Dokument als historisch archivieren; Lieferstand wandert dokumentarisch nach `CHANGELOG.md`.
