# Slice 022: Docker-Harness-Audit Teil 3 — hermetische Fuzz-/Mutation-Stages mit Schreib-Rückweg-Export

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](../welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 9c von 11 — Docker-Harness-Audit, Teil 3 von 3).

**Bezug:** `apps/api/Dockerfile` (neue Stage `fuzz`), `apps/api/Makefile`
(`fuzz-check`, `mutation-report`), Regelwerk `modul-14-docker-harness.md`
§Besitz der Belege eines containerisierten Gates, §Der Prüflauf ist
hermetisch — kein Mount.

**Autor:** Nachgang Templates-Diff v3.5.1→v6.8.0, Owner-Entscheidung
„volle Rearchitektur" (AskUserQuestion, 2026-09-13) — Teil 3 von 3, nach
`slice-020`/`slice-021`. **Datum:** 2026-09-13.

---

## 1. Ziel

`fuzz-check` und `mutation-report` von Bind-Mount auf hermetische
Dockerfile-Stages umstellen — **inklusive** Schreib-Rückweg (Crash-
Reproduktionen unter `testdata/fuzz/`, Mutation-Report unter
`.tmp/mutation/`) über `docker create`/`docker cp`/`docker rm` statt
Bind-Mount, damit die Artefakte dem aufrufenden Host-User gehören, nicht
`root` (Modul 14 §Besitz der Belege).

## 2. Definition of Done

- [ ] **`apps/api/Dockerfile`**: neue Stage `FROM deps AS fuzz` (`COPY .
      .`, kein `RUN` des Fuzz-Laufs selbst — analog `benchmark`/`vuln`
      aus `slice-021`). **Keine** eigene `mutation`-Stage nötig — sie
      teilt sich `deps` direkt mit `gremlins` als Laufzeit-`go install`
      im Makefile-Aufruf (Tool-Version ändert sich unabhängig vom
      Quellstand; ein eigener Stage-Cache brächte hier keinen Vorteil,
      da `gremlins unleash` ohnehin bei jedem Lauf frisch mutiert).
- [ ] **`apps/api/Makefile` `fuzz-check`**: `docker build --target fuzz`,
      dann `docker create` + `docker start -a` (streamt stdout, kein
      `--rm`) + `docker cp <cid>:/src/testdata/fuzz testdata/fuzz` +
      `docker rm` — Crash-Reproduktionen kommen dem Host-User gehörend
      zurück, nicht `root:root`.
- [ ] **`apps/api/Makefile` `mutation-report`**: gleiches Muster
      (`docker build --target fuzz` wiederverwendet — `mkdir -p
      .tmp/mutation` bleibt Host-seitig vor dem Lauf, `docker cp
      <cid>:/src/.tmp/mutation .tmp/` nach Lauf-Ende) statt direktem
      Bind-Mount-Schreibzugriff.
- [ ] **Verifikation:** `make fuzz-check` (in `apps/api/`, alle sechs
      `Fuzz*`-Targets, Default-`FUZZTIME=30s`) läuft durch, `ls -l
      testdata/fuzz` (falls Crashes) zeigt Host-User-Eigentümerschaft.
      `make mutation-report` läuft durch, `ls -l .tmp/mutation/`
      zeigt Host-User-Eigentümerschaft (Root-Ownership-Risiko aus
      `slice-021` §6 damit aufgelöst).
- [ ] `make gates` grün (beide Targets sind opt-in, nicht Teil von
      `make gates` selbst — Verifikation lässt `make gates` unberührt
      grün bleiben).
- [ ] `make docs-check` grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `apps/api/Dockerfile` | neue Stage `fuzz` | hermetische Basis für Fuzz-Lauf |
| `apps/api/Makefile` | `fuzz-check` umgeschrieben | create/start/cp/rm statt Bind-Mount |
| `apps/api/Makefile` | `mutation-report` umgeschrieben | create/start/cp/rm statt Bind-Mount |

## 4. Trigger

- **`in-progress`:** nach `slice-021` (Muster `benchmark`/`vuln`-Stage
  etabliert, `fuzz` folgt demselben Ansatz).
- **Rückführung:** falls `docker cp` aus einem gestoppten (nicht `--rm`)
  Container in diesem Docker/BuildKit-Stand unerwartetes Verhalten zeigt
  (z. B. Owner-Mapping bei Rootless-Docker) — dann Root-Ownership-Fix
  als eigenen, kleineren Folge-Slice abspalten, Bind-Mount-zu-Build-Stage-
  Umstellung trotzdem behalten.

## 5. Closure-Trigger

DoD vollständig + `make gates` grün + `make docs-check` grün +
Closure-Notiz + `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **`docker cp` aus gestopptem statt laufendem Container** — Muster ist
  Standard-Docker (kein Rootless-/Podman-Sonderfall in m-traces
  CI-Umgebung bekannt), aber erste Verwendung in diesem Repo; bei
  Auffälligkeiten Rückführung siehe §4.
- **Kein ADR nötig.** Reine Build-Infrastruktur-Härtung.
- Nach diesem Slice ist Tranche 9 (Docker-Harness-Audit) vollständig
  abgeschlossen (9a/9b/9c alle `done`).

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Dockerfiles + Makefiles (Werkzeug/Prozess)

Reine Build-Infrastruktur-Härtung, kein Anwendungscode geändert. Ohne
ADR — keine Architekturentscheidung, kein Gate gesenkt.
