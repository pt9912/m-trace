# Migration Plan — npm-Paketnamen `@m-trace/*` → `@npm9912/*`

> **Status**: ⬜ geplant, durchzuführen vor dem `0.2.0`-Release.
> **Bezug**: [`plan-0.2.0.md`](./plan-0.2.0.md) Tranche 1 (RAK-11, publizierbares SDK); [`lastenheft.md`](./lastenheft.md) §16.2 (OE-8).
> **Trigger**: Maintainer publishet npm-Pakete bereits unter dem bestehenden Scope `@npm9912`. Eine separate `@m-trace`-Org ist auf npmjs.com nicht reserviert. OE-8 wird damit faktisch wiedereröffnet und neu entschieden, bevor das SDK erstmals veröffentlicht wird.

## 1. Entscheidung

- Neuer Player-SDK-Paketname **ab `0.2.0`**: `@npm9912/player-sdk`.
- Neuer Dashboard-Workspace-Paketname **ab `0.2.0`**: `@npm9912/m-trace-dashboard`.
- Historische Artefakte mit explizitem `0.1.x`-Bezug bleiben **unverändert** (Wire-Fixtures, Smoke-Skripte, abgeschlossene Plan-Dokumente, CHANGELOG-Einträge in `0.1.x`-Sektionen). Begründung: sie dokumentieren den Lieferstand `0.1.0` / `0.1.1` und müssen reproduzierbar bleiben.
- Lebende Spezifikations- und Code-Stellen werden auf den neuen Namen umgestellt.
- Globaler IIFE-Name `MTracePlayerSDK` (tsup `--global-name`) bleibt erhalten — er ist orthogonal zum npm-Paketnamen und kein Branding-Bruch.

## 2. Scope-Abgrenzung

### 2.1 Geändert (Lieferstand `0.2.0`)

| Datei | Zeile(n) | Art der Änderung |
|---|---|---|
| `packages/player-sdk/package.json` | 2 | `name` → `@npm9912/player-sdk` |
| `packages/player-sdk/package.json` | 40 | `pack:smoke`-Tarball-Pfad → `npm9912-player-sdk-0.2.0.tgz` |
| `packages/player-sdk/src/core/tracker.ts` | 6 | `sdk.name`-Konstante |
| `packages/player-sdk/tests/tracker.test.ts` | 46, 54 | Test-Erwartung |
| `packages/player-sdk/scripts/pack-smoke.mjs` | 64 | `require()`-Pfad im CJS-Smoke |
| `packages/player-sdk/README.md` | 1, 9, 19, 62, 113 | Paket-Doku |
| `apps/dashboard/package.json` | 2, 10, 11, 14 | Paketname, Workspace-Dependency + Build-Hooks |
| `apps/dashboard/Dockerfile` | 11, 16 | Build-Filter |
| `apps/dashboard/src/routes/demo/+page.svelte` | 5 | Import |
| `apps/api/adapters/driving/http/handler_test.go` | 39 | Test-Fixture |
| `apps/api/hexagon/application/register_playback_event_batch_test.go` | 162 | Test-Fixture |
| `docs/player-sdk.md` | 4, 14, 24, 55 | Pflichtdoku gem. Lastenheft §7.6 |
| `docs/telemetry-model.md` | 47, 100 | Spec-Beispiel |
| `docs/lastenheft.md` | 1043, 1756 | Beispiel-Payload + OE-8-Tabelleneintrag (§5) |
| `docs/local-development.md` | 93 | Beispiel-Curl: Name **und** `sdk.version` auf `0.2.0` |
| `docs/plan-0.2.0.md` | 64 (neuer DoD-Eintrag), 69, 70 | DoD-Verweis auf diese Migration + `pnpm --filter ...`-Targets (bereits mit DoD-Eintrag normalisiert) |
| `README.md` | 136 | Beispiel-Payload + `sdk.version` auf `0.2.0` |
| `CHANGELOG.md` | (neuer Eintrag in `0.2.0`-Sektion) | OE-8-Neuentscheidung dokumentieren |
| `pnpm-lock.yaml` | 29 | regeneriert über `pnpm install` |

### 2.2 Nicht geändert (historisch, an `0.1.x` gepinnt)

| Datei | Begründung |
|---|---|
| `scripts/smoke-0.1.0.sh` | reproduziert `0.1.0`-Wire-Format inkl. damaligem `sdk.name` |
| `scripts/smoke-0.1.1.sh` | analog `0.1.1` |
| `scripts/seed-rak9.sh` Z. 140 | sendet explizit `sdk.version: 0.1.1` (RAK-9-Verifikation, abgeschlossen) |
| `docs/plan-0.1.0.md` Z. 331 | dokumentiert OE-8-Resolution zum Zeitpunkt `0.1.0` (`bae4a2a`) |
| `docs/plan-0.1.1.md` Z. 90 | analog für `0.1.1` |
| `docs/plan-spike.md` Z. 997 | abgeschlossener Spike |
| `docs/spike/backend-api-contract.md`, `docs/spike/0001-backend-stack.md` | Spike-Snapshots |
| `docs/migrate-package-name.md` | dieses Migrationsprotokoll nennt alten und neuen Namen ausdrücklich |
| `docs/lastenheft.md` §16.2 | OE-8-Neuentscheidung enthält die historische `0.1.x`-Einordnung |
| `CHANGELOG.md` Z. 35 | Eintrag in `0.1.1`-Sektion: „Lastenheft `1.1.5` löst OE-8 auf: `@m-trace/player-sdk`" — historische Wahrheit |

## 3. Durchführungsreihenfolge

1. **Lastenheft + Cross-Cutting-Plan**: OE-8 in `docs/lastenheft.md:1756` neu fassen; in `docs/plan-0.1.0.md` Tranche 0c einen neuen §4a-Eintrag ergänzen, der die Neuentscheidung begründet (Konvention aus `plan-0.2.0.md` §0).
2. **SDK-Paket**: `package.json` (`name`, `pack:smoke`-Pfad), `src/core/tracker.ts`, Tests, `scripts/pack-smoke.mjs`, `README.md`.
3. **Dashboard-Konsument und Workspace-Paket**: `apps/dashboard/package.json`, `Dockerfile`, `routes/demo/+page.svelte`.
4. **Backend-Test-Fixtures**: beide Go-Tests.
5. **Lebende Spec-Doku**: `telemetry-model.md`, `lastenheft.md` (Beispiel-Payload Z. 1043), `player-sdk.md`, `local-development.md`, `README.md`, `plan-0.2.0.md`.
6. **CHANGELOG**: Eintrag in `0.2.0`-Sektion ergänzen — siehe §5.
7. **Lockfile**: `pnpm install` an Repo-Root.
8. **Verifikation** (§4).

## 4. Verifikation

Nach der Umbenennung müssen folgende Befehle grün laufen:

- `pnpm --filter @npm9912/player-sdk run build`
- `pnpm --filter @npm9912/player-sdk run test`
- `pnpm --filter @npm9912/player-sdk run lint` (TypeScript + Boundaries + Public-API-Snapshot)
- `pnpm --filter @npm9912/player-sdk run pack:smoke` — erzeugt `.tmp/player-sdk-pack/npm9912-player-sdk-0.2.0.tgz` und der CJS-Require-Smoke importiert das umbenannte Paket
- `pnpm --filter @npm9912/player-sdk run publish:dry-run` — npm akzeptiert den neuen Scope-Namen formal
- `pnpm --filter @npm9912/m-trace-dashboard run lint` (Workspace-Dependency-Auflösung)
- `make test`, `make lint`, `make build`, `make browser-e2e` an Repo-Root
- Such-Gate: `rg '@m-trace/player-sdk'` darf **nur noch** in den unter §2.2 gelisteten historischen Dateien Treffer liefern.

## 5. Lastenheft-Update (OE-8)

Aktueller Eintrag in `docs/lastenheft.md:1756`:

> `| OE-8 | resolved | npm-Paketname Player-SDK: **`@m-trace/player-sdk`** |`

Neu:

> `| OE-8 | resolved | npm-Paketname Player-SDK: **`@npm9912/player-sdk`** ab `0.2.0`; Lieferstand `0.1.x` wurde unter `@m-trace/player-sdk` ausgeliefert (Lastenheft `1.1.5`, `bae4a2a`), aber nie öffentlich publishet. |`

CHANGELOG-Eintrag in der `0.2.0`-Sektion (Block `Changed`):

> Lastenheft `1.1.7` (oder die zum Release gültige Minor-Version) entscheidet OE-8 neu: Player-SDK wird ab `0.2.0` als `@npm9912/player-sdk` veröffentlicht. Der `0.1.x`-Lieferstand wurde nie öffentlich publishet, daher kein Migrations-Pfad für externe Konsumenten erforderlich.

In `docs/plan-0.1.0.md` Tranche 0c als neuen §4a-Eintrag aufnehmen (cross-cutting Lastenheft-Patch — Konvention aus `plan-0.2.0.md` §0): kurze Begründung („`@m-trace` ist auf npm nicht als Org reserviert, Maintainer publishet unter `@npm9912`") und Verweis auf dieses Migrations-Dokument.

## 6. Risiken und offene Punkte

- **Wire-Vertrag**: `sdk.name` ist laut `telemetry-model.md` §1.2 nur „string, nicht-leer" — die API matcht den Wert nicht gegen eine Whitelist. Backend-seitig sind daher nur Test-Fixtures betroffen; kein Migrationspfad für bereits persistierte In-Memory-Events nötig (auch weil OE-3 noch In-Memory ist).
- **Compat-Fixture (RAK-21, `plan-0.2.0.md` Tranche 2)**: Falls dort `sdk.name` als geprüftes Feld in `contracts/sdk-compat.json` aufgenommen wird, mit dem neuen Namen befüllen. Diese Migration berührt das Fixture noch nicht, weil es zu diesem Zeitpunkt noch nicht existiert.
- **Tarball-Dateiname**: `pnpm pack` benennt Scoped-Tarballs nach dem Schema `<scope>-<name>-<version>.tgz`. Aus `@npm9912/player-sdk@0.2.0` wird `npm9912-player-sdk-0.2.0.tgz`. `package.json` Skript `pack:smoke` und ggf. spätere CI-Artefakt-Pfade müssen das berücksichtigen.
- **Externe Konsumenten**: Es existieren keine. Das Paket war noch nie auf npm publishet, daher keine Migrationskommunikation, keine `deprecate`-Markierung auf der Registry erforderlich.
- **Globaler IIFE-Name**: `MTracePlayerSDK` bleibt. CDN-Konsumenten würden im Browser weiterhin `window.MTracePlayerSDK` nutzen. Falls später ein Branding-Sweep gewünscht ist, separate Entscheidung außerhalb dieses Plans.

## 7. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit Status mit Hash markieren analog `plan-0.2.0.md` §0.
- Nach Abschluss dieses Plans wird das Dokument im `0.2.0`-Release als „durchgeführt" markiert; eine Archivierung erfolgt mit dem Bump auf `0.3.0`.
