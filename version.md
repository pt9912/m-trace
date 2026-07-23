# m-trace — Release-Register

> Kanonischer, **auflösender** Link-Ziel-Ort für Erwähnungen **eigener**
> m-trace-Releases — etwa [die jeweils aktuelle Version](#aktuell). **Nur die
> aktuelle Version** trägt einen expliziten HTML-Anker `#X.Y.Z` (wörtlich, mit
> Punkten — der Heading-/Tabellen-Slug verschluckt sie). Beim Release **wandert**
> der Anker zur neuen aktuellen Version; die bisherige Zeile verliert ihn —
> dadurch *bricht* jeder feste Pin auf eine veraltete Version (`anchor-missing`),
> und ein vergessener Bump fällt auf (das ist der Zweck dieses Registers).
>
> **Versions-Form:** Die Image-/Paket-Version ist `X.Y.Z` (ohne `v`); der
> zugehörige Git-Tag ist `vX.Y.Z`. Das Modul `versions` (`.d-check.yml`) prüft
> gepinnte `ghcr.io/pt9912/m-trace-*`-Image-Verweise gegen die
> `#aktuell`-Version.
>
> **Kein Duplikat** der Detail-Changes — die stehen im
> [CHANGELOG](CHANGELOG.md). Hier nur Versions-Koordinaten (Version, Datum, Tag).
> Bootstrapped in der v3.5.0-Migration W7 (2026-07-23).

## Aktuell

Aktuelle Version: [`0.25.0`](#0.25.0) — 2026-07-13
(Tag [`v0.25.0`](https://github.com/pt9912/m-trace/releases/tag/v0.25.0)).

Aus anderen Dokumenten stabil referenzierbar als `version.md#aktuell` (zeigt
immer hierher, nie auf eine feste Nummer). Pro Release sind genau diese Zeile
**und** eine neue `## Verlauf`-Zeile nachzuziehen **und der `<a id>`-Anker auf
die neue Version zu verschieben** (die bisherige Zeile verliert ihn) — siehe
[`docs/user/releasing.md`](docs/user/releasing.md).

## Aktuelle Runtime-Images

Die drei auf GHCR veröffentlichten Runtime-Images der aktuellen Version:

- `ghcr.io/pt9912/m-trace-api:0.25.0`
- `ghcr.io/pt9912/m-trace-dashboard:0.25.0`
- `ghcr.io/pt9912/m-trace-analyzer-service:0.25.0`

Die deploy/k8s-Manifeste tragen dieselben Tags; `make k8s-validate` erzwingt,
dass sie mit der Root-`package.json`-Version übereinstimmen (der maschinell
durchgesetzte Versions-Guard).

## Verlauf

| Version | Datum | Tag |
|---|---|---|
| `0.25.0` <a id="0.25.0"></a> | 2026-07-13 | [`v0.25.0`](https://github.com/pt9912/m-trace/releases/tag/v0.25.0) |
| `0.23.0` | 2026-07-11 | [`v0.23.0`](https://github.com/pt9912/m-trace/releases/tag/v0.23.0) |
| `0.22.4` | 2026-06-23 | [`v0.22.4`](https://github.com/pt9912/m-trace/releases/tag/v0.22.4) |
| `0.22.3` | 2026-06-16 | [`v0.22.3`](https://github.com/pt9912/m-trace/releases/tag/v0.22.3) |
| `0.22.2` | 2026-06-03 | [`v0.22.2`](https://github.com/pt9912/m-trace/releases/tag/v0.22.2) |
| `0.22.1` | 2026-05-17 | [`v0.22.1`](https://github.com/pt9912/m-trace/releases/tag/v0.22.1) |
| `0.22.0` | 2026-05-13 | [`v0.22.0`](https://github.com/pt9912/m-trace/releases/tag/v0.22.0) |
| `0.21.0` | 2026-05-13 | [`v0.21.0`](https://github.com/pt9912/m-trace/releases/tag/v0.21.0) |
| `0.20.0` | 2026-05-13 | [`v0.20.0`](https://github.com/pt9912/m-trace/releases/tag/v0.20.0) |
| `0.18.0` | 2026-05-13 | [`v0.18.0`](https://github.com/pt9912/m-trace/releases/tag/v0.18.0) |
| `0.17.0` | 2026-05-13 | [`v0.17.0`](https://github.com/pt9912/m-trace/releases/tag/v0.17.0) |
| `0.16.0` | 2026-05-12 | [`v0.16.0`](https://github.com/pt9912/m-trace/releases/tag/v0.16.0) |
| `0.15.0` | 2026-05-12 | [`v0.15.0`](https://github.com/pt9912/m-trace/releases/tag/v0.15.0) |
| `0.14.0` | 2026-05-12 | [`v0.14.0`](https://github.com/pt9912/m-trace/releases/tag/v0.14.0) |
| `0.13.0` | 2026-05-12 | [`v0.13.0`](https://github.com/pt9912/m-trace/releases/tag/v0.13.0) |
| `0.12.6` | 2026-05-12 | [`v0.12.6`](https://github.com/pt9912/m-trace/releases/tag/v0.12.6) |
| `0.12.5` | 2026-05-11 | [`v0.12.5`](https://github.com/pt9912/m-trace/releases/tag/v0.12.5) |
| `0.12.1` | 2026-05-10 | [`v0.12.1`](https://github.com/pt9912/m-trace/releases/tag/v0.12.1) |
| `0.12.0` | 2026-05-10 | [`v0.12.0`](https://github.com/pt9912/m-trace/releases/tag/v0.12.0) |
| `0.11.0` | 2026-05-09 | [`v0.11.0`](https://github.com/pt9912/m-trace/releases/tag/v0.11.0) |
| `0.10.0` | 2026-05-09 | [`v0.10.0`](https://github.com/pt9912/m-trace/releases/tag/v0.10.0) |
| `0.9.6` | 2026-05-08 | [`v0.9.6`](https://github.com/pt9912/m-trace/releases/tag/v0.9.6) |
| `0.9.5` | 2026-05-07 | [`v0.9.5`](https://github.com/pt9912/m-trace/releases/tag/v0.9.5) |
| `0.9.1` | 2026-05-07 | [`v0.9.1`](https://github.com/pt9912/m-trace/releases/tag/v0.9.1) |
| `0.9.0` | 2026-05-07 | [`v0.9.0`](https://github.com/pt9912/m-trace/releases/tag/v0.9.0) |
| `0.8.5` | 2026-05-07 | [`v0.8.5`](https://github.com/pt9912/m-trace/releases/tag/v0.8.5) |
| `0.8.0` | 2026-05-06 | [`v0.8.0`](https://github.com/pt9912/m-trace/releases/tag/v0.8.0) |
| `0.7.0` | 2026-05-06 | [`v0.7.0`](https://github.com/pt9912/m-trace/releases/tag/v0.7.0) |
| `0.6.0` | 2026-05-05 | [`v0.6.0`](https://github.com/pt9912/m-trace/releases/tag/v0.6.0) |
| `0.5.0` | 2026-05-05 | [`v0.5.0`](https://github.com/pt9912/m-trace/releases/tag/v0.5.0) |
| `0.4.0` | 2026-05-05 | [`v0.4.0`](https://github.com/pt9912/m-trace/releases/tag/v0.4.0) |
| `0.3.0` | 2026-05-01 | [`v0.3.0`](https://github.com/pt9912/m-trace/releases/tag/v0.3.0) |
| `0.2.0` | 2026-04-30 | [`v0.2.0`](https://github.com/pt9912/m-trace/releases/tag/v0.2.0) |
| `0.1.2` | 2026-04-30 | [`v0.1.2`](https://github.com/pt9912/m-trace/releases/tag/v0.1.2) |
| `0.1.1` | 2026-04-30 | [`v0.1.1`](https://github.com/pt9912/m-trace/releases/tag/v0.1.1) |
| `0.1.0` | 2026-04-30 | [`v0.1.0`](https://github.com/pt9912/m-trace/releases/tag/v0.1.0) |
