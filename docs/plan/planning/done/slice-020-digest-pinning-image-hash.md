# Slice 020: Docker-Harness-Audit Teil 1 — Base-Image-Digest-Pinning + `harness/image-hash.txt`

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 9a von 11 — Docker-Harness-Audit Teil 1 von 2).

**Bezug:** `apps/api/Dockerfile`, `apps/dashboard/Dockerfile`,
`apps/analyzer-service/Dockerfile`, `apps/api/Makefile`, `Makefile`
(`image-build`), neu `harness/image-hash.txt`, `harness/README.md`,
Regelwerk `modul-14-docker-harness.md` §Multi-Stage-Build: die operativen
Disziplinen.

**Autor:** Nachgang Templates-Diff v3.5.1→v6.8.0, Owner-Entscheidung
„volle Rearchitektur" (AskUserQuestion, 2026-09-13) — dieser Slice ist der
erste, risikoarme Teil. **Datum:** 2026-09-13.

---

## 1. Ziel

Alle Base-Images der drei App-Dockerfiles und der Ad-hoc-`docker
run`-Aufrufe in `apps/api/Makefile` per `@sha256:`-Digest pinnen (statt
floatender Tags), und `harness/image-hash.txt` als Beleg-Artefakt für die
drei Runtime-Images einführen.

**Ausdrücklich NICHT in diesem Slice:**
- Die Bind-Mount-Prüfläufe (`benchmark-smoke`, `fuzz-check`,
  `mutation-report` in `apps/api/Makefile`) auf hermetische Dockerfile-
  Gate-Stages umstellen — größerer, separat zu verifizierender Umbau,
  siehe Tranche 9b.
- Die `docker.sock`-Mounts in `vuln-check`/`image-scan` — andere
  Mount-Klasse (Zugriff auf den Daemon zum Scannen bereits gebauter
  Images, kein Quellbaum-Mount); Einordnung folgt in Tranche 9b.
- Das Spannungsfeld `apt-get upgrade` (Sicherheits-Frische, bewusst
  gewählt) vs. Digest-Pinning (Reproduzierbarkeit) — beide Ziele bleiben
  nebeneinander bestehen, dokumentiert im Closure-Note, kein Zielkonflikt
  dieses Slices.

## 2. Definition of Done

- [x] **`apps/api/Dockerfile`**: `golang:1.26.6`, `golangci/golangci-lint:v2.12.1-alpine`,
      `gcr.io/distroless/static-debian12:nonroot` je mit `@sha256:…`-Digest
      (per `docker buildx imagetools inspect <image:tag>` ermittelt,
      2026-09-13) — Tag bleibt als Kommentar neben dem Digest für
      Menschen-Lesbarkeit.
- [x] **`apps/dashboard/Dockerfile`, `apps/analyzer-service/Dockerfile`**:
      beide `node:22-trixie-slim`-Stages (build + runtime) je mit
      `@sha256:…`-Digest.
- [x] **`apps/api/Makefile`**: die drei Ad-hoc-`docker run … golang:1.26.6`-
      Aufrufe (`benchmark-smoke`, `fuzz-check`, `mutation-report`) **und**
      `Makefile` (root) `vuln-check` (vierter, beim Umsetzen gefundener
      Fall) auf denselben Digest gezogen — reine String-Änderung, **keine**
      Struktur-/Mount-Änderung (die kommt in Tranche 9b).
- [x] **`harness/image-hash.txt`** neu: `make image-build` erweitert um
      `--metadata-file`, Digest je Service per `jq` extrahiert
      (`api`/`dashboard`/`analyzer-service`, je eine Zeile
      `<service>: sha256:…`).
- [x] **`harness/README.md`** referenziert `harness/image-hash.txt` (§Guides
      oder §Sensors, je nach Passung).
- [x] **Verifikation:** `make image-build VER=<test-tag>` läuft durch,
      erzeugt korrektes `harness/image-hash.txt`; `make gates` läuft
      vollständig grün (Digest-Pin ändert den Image-*Inhalt* nicht, nur
      seine Adressierung — bereits lokal gecachte Digests bestätigt
      identisch mit den bisherigen Tags).
- [x] `make docs-check` grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `apps/api/Dockerfile` | 3 `FROM`-Zeilen pinnen | Digest statt Tag |
| `apps/dashboard/Dockerfile` | 2 `FROM`-Zeilen pinnen | Digest statt Tag |
| `apps/analyzer-service/Dockerfile` | 2 `FROM`-Zeilen pinnen | Digest statt Tag |
| `apps/api/Makefile` | 3 `docker run`-Zeilen pinnen | Digest statt Tag, konsistent mit Dockerfile |
| `Makefile` (`image-build`) | `--metadata-file` + `jq`-Extraktion ergänzen | `harness/image-hash.txt` erzeugen |
| `harness/README.md` | Zeiger ergänzen | Beleg-Artefakt referenzierbar machen |

**Bereits geklärt (2026-09-13):** Digests via
`docker buildx imagetools inspect` ermittelt und gegen lokal gecachte
Images verifiziert (identisch, kein Überraschungs-Drift):
`node:22-trixie-slim` → `7b8a0c89c54499bee567618f96578e1a12a800f062fbdbfd1fb6a443fa6f6284`,
`golang:1.26.6` → `0d1d3a794be25f809dd2cb3160d8c73276c4056a9f8242a138e908ddeee7b6b6`,
`golangci/golangci-lint:v2.12.1-alpine` → `25cd0933d1a4ced61e907aabac044cfc0182ab067c09c25d568e15750e9695aa`,
`gcr.io/distroless/static-debian12:nonroot` → `afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab`.

## 4. Trigger

- **`in-progress`:** sofort, keine Abhängigkeiten außer Tranche 1
  (Vendoring, bereits `done`).
- **Rückführung:** falls `--metadata-file`-Wiring im lokalen Docker/
  BuildKit-Stand nicht sauber funktioniert (Buildx-Version-Abhängigkeit) —
  dann Digest-Pinning ohne `image-hash.txt` als reduzierter Slice
  abschließen, `image-hash.txt` als Folge-Punkt vermerken.

## 5. Closure-Trigger

DoD vollständig + `make gates` grün + `make docs-check` grün +
Closure-Notiz + `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Digest-Drift bei künftigen Base-Image-Updates.** Ein Digest-Pin
  friert die Toolchain ein — Updates werden künftig ein bewusster Commit
  (Digest-Zeile anheben), nicht mehr automatisch. Das ist der Zweck der
  Regel, kein Nebeneffekt.
- **Spannung `apt-get upgrade` vs. Digest-Pinning** (siehe §1) — bewusst
  als Nebeneinander dokumentiert, kein Widerspruch, der aufgelöst werden
  müsste.
- **Kein ADR nötig.** Reine Reproduzierbarkeits-Härtung, keine
  Architekturentscheidung, kein Gate gesenkt.

## 7. Closure-Notiz (nach `done/`)

Alle vier Base-Images (`golang:1.26.6`, `golangci/golangci-lint:v2.12.1-alpine`,
`gcr.io/distroless/static-debian12:nonroot`, `node:22-trixie-slim`) in den
drei App-Dockerfiles per `@sha256:`-Digest gepinnt. Bei den Ad-hoc-
`docker run`-Aufrufen in `apps/api/Makefile` fand sich beim Umsetzen ein
**vierter** Fall über die geplanten drei hinaus: `vuln-check` im
root-`Makefile` nutzte denselben unpgepinnten `golang:1.26.6` — mit
gepinnt.

`make image-build` erweitert um `--metadata-file` je Service +
`jq`-Extraktion nach `harness/image-hash.txt`. Ein `jq`-Quoting-Fehler
beim ersten Versuch (`.\"containerimage.digest\"` — Make doppelt-
escapte die Backslashes) auf die robustere `.["containerimage.digest"]`-
Syntax korrigiert, mit einem echten `make image-build VER=…`-Testlauf
verifiziert (drei Digests korrekt geschrieben, Test-Images danach
gelöscht). `harness/README.md` §Werkzeuge — kein Gate bekam die Zeile.

**Verifikation:** vollständiger `make gates`-Lauf grün (api-race, ts-test,
lint, coverage-gate, arch-check, schema-validate, generated-drift-check,
schema-generate-postgres-check, sdk-pack-smoke, sdk-performance-smoke,
benchmark-smoke) — Digest-Pinning ändert den Image-Inhalt nicht, nur
seine Adressierung, bestätigt durch identische lokale Digests vor der
Änderung. `make docs-check` — 0 Befunde (nach Marker-Rücksetzung).

**Steering-Loop-Lerneintrag:** Ein systematischer `grep` über beide
Makefiles nach dem Image-Namen (statt nur die im Plan vorab benannten
Stellen zu patchen) fand den vierten, ungeplanten Fall (`vuln-check`).
Dieselbe Lektion wie in `slice-018` (Vorab-Schätzung vs. Audit-Befund):
ein `grep -rn "<image>"` über den gesamten Suchraum ist billiger als das
Risiko, eine Stelle zu übersehen, die dieselbe Klasse Drift trägt wie die
gerade gefixten.

**Folge-Slices:** Tranche 9b (hermetische Gate-Stages für `apps/api`
Bench/Fuzz/Mutation, Security-Scan-Mount-Abgrenzung) — noch nicht
geschnitten.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Dockerfiles + Makefiles (Werkzeug/Prozess)

Reine Build-Infrastruktur-Härtung, kein Anwendungscode geändert. Ohne
ADR — keine Architekturentscheidung, kein Gate gesenkt (im Gegenteil:
Reproduzierbarkeit wird schärfer).
