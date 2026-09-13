# `make gates` — bündelt alle inneren Quality-Gates in einem Lauf

## Vertrag

Was verletzt wäre, wenn `make gates` rot wird: mindestens eines der
gebündelten Gates urteilt negativ über den aktuellen Repo-Zustand. Das
Bündel (`Makefile`-Definition von `gates`): `api-race`, `ts-test`, `lint`
(`api-lint` + `ts-lint`), `coverage-gate` (`api-coverage-gate` +
`ts-coverage-gate`), `arch-check`, `schema-validate`,
`generated-drift-check`, `schema-generate-postgres-check`,
`sdk-pack-smoke`, `sdk-performance-smoke`, `benchmark-smoke`, `docs-check`,
`lint-variante-b`, `verify-closure-notes`.

## Grenze — was das Grün nicht abdeckt

1. **`security-gates`** (`vuln-check`, `audit-ts`, `image-scan`,
   `image-start-check`) — läuft als separater CI-Job, nicht in diesem
   Bündel; permanent, nicht heilbar durch Aufnahme (Security-Läufe brauchen
   Netzwerk/Registry-Zugriff, `gates` ist netzlos).
2. **`docs-immutable`/`docs-commits`** — Bereichs-Parameter (`STAGED=1`/
   `RANGE=…`) sind für einen Pull-Request-Kontext gedacht, nicht für einen
   einzelnen Arbeitsbaum-Snapshot; laufen separat in CI.
3. **`release-guard`/`release-gate`/`k8s-validate`/`devcontainer-validate`**
   — Release- und Deploy-spezifische Prüfungen, nicht Teil der laufenden
   Entwicklungs-Gates; permanent getrennt.
4. **`smoke-*`-Targets** (Docker-Compose-Stack-Smokes) — laufen einzeln in
   CI-Jobs, nicht gebündelt in `gates` (Laufzeit/Ressourcen).

**Wie groß der Ausschnitt ist, sagt das Kommando, nicht diese Datei:**
`grep -n '^gates:' Makefile`. Eine eingefrorene Liste stünde hier falsch,
sobald jemand das Bündel ändert.

## Ausgabe und Ausgänge

| Exit | Bedeutung |
|---|---|
| 0 | alle gebündelten Gates grün |
| ≠ 0 | das erste fehlschlagende Sub-Target bricht den `make`-Lauf ab (Standard-`make`-Semantik, keine Sonderbehandlung) |

## Sperren

- Docker-Daemon erreichbar — alle Sub-Gates laufen containerisiert
  (Docker-only, siehe `AGENTS.md` §3.1) → ohne laufenden Daemon bricht der
  erste Sub-Target ab.

## Bindung

—
