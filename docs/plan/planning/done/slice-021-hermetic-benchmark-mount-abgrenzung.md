# Slice 021: Docker-Harness-Audit Teil 2 — hermetische `benchmark-smoke`/`vuln-check`-Stages, Mount-Abgrenzung

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 9b von 11 — Docker-Harness-Audit, Teil 2 von 3).

**Bezug:** `apps/api/Dockerfile` (neue Stages `benchmark`, `vuln`),
`apps/api/Makefile` (`benchmark-smoke`), `Makefile` (`vuln-check`),
Regelwerk `modul-14-docker-harness.md` §Der Prüflauf ist hermetisch —
kein Mount.

**Autor:** Nachgang Templates-Diff v3.5.1→v6.8.0, Owner-Entscheidung
„volle Rearchitektur" (AskUserQuestion, 2026-09-13) — Teil 2 von 3, nach
`slice-020` (Digest-Pinning). **Datum:** 2026-09-13.

---

## 1. Ziel

`benchmark-smoke` (Teil von `make gates`) und `vuln-check` (Teil von
`make security-gates`) von Bind-Mount (`-v $(CURDIR):/src`) auf
hermetische Dockerfile-Stages umstellen — Quellen per `COPY` gebacken,
kein Mount, klarer stdout-Rückweg für die nachgelagerten Parser
(`check-bench-budgets.mjs`).

**Ausdrücklich NICHT in diesem Slice:**
- `fuzz-check`/`mutation-report` — beide schreiben Ergebnisse **in den
  Baum zurück** (Crash-Reproduktionen bzw. JSON-Report), was einen
  Export-Stage-Umbau (`FROM scratch` + `--output type=local,dest=…`,
  analog der bestehenden `coverage-export`-Stage) braucht, keinen reinen
  Stage-Wechsel. Eigener Folge-Slice, Tranche 9c.
- `image-scan` (Trivy) — mountet `docker.sock` + Trivy-Cache, **kein**
  Quellbaum-Mount. Andere Mount-Klasse (Zugriff auf den Daemon, um
  bereits gebaute Images zu scannen), von der „kein Mount"-Regel für
  Prüfläufe gegen den **Arbeitsbaum** nicht erfasst — Audit-Befund, keine
  Adaption, kein MR nötig (Regelwerk-Schweigen zu dieser Mount-Klasse ist
  erlaubend, keine Lücke).

## 2. Definition of Done

- [x] **`apps/api/Dockerfile`**: neue Stage `FROM deps AS benchmark` (nur
      `COPY . .`, **kein** `RUN` des Bench selbst — der eigentliche Lauf
      passiert bei `docker run` gegen das gebaute Image, sonst wäre das
      Bench-Ergebnis im BuildKit-Log präfixiert und für
      `check-bench-budgets.mjs`s verankerten `^Benchmark`-Regex nicht
      mehr parsebar).
- [x] **`apps/api/Makefile` `benchmark-smoke`**: `docker build --target
      benchmark -t <lokaler-tag>:benchmark apps/api` (Digest/Layer-Cache
      trägt Frische) gefolgt von `docker run --rm <tag>:benchmark go
      test …` — kein Mount, sauberer stdout.
- [x] **`apps/api/Dockerfile`**: neue Stage `FROM deps AS vuln` (`COPY .
      .`, `RUN go install golang.org/x/vuln/cmd/govulncheck@…` als
      Build-Schritt — das Tool selbst ist ein legitimer Build-Bestandteil,
      **nicht** der eigentliche Scan).
- [x] **`Makefile` (root) `vuln-check`**: `docker build --target vuln -t
      <tag>:vuln apps/api` gefolgt von `docker run --rm <tag>:vuln
      govulncheck ./...` — kein Mount.
- [x] **Verifikation:** `make benchmark-smoke` (in `apps/api/`) und `make
      vuln-check` (root) laufen durch mit identischem
      Ausgabe-Format (Budget-Parser bzw. govulncheck-Report unverändert
      lesbar); `make gates` vollständig grün.
- [x] **Audit-Notiz dokumentiert**: `image-scan`s `docker.sock`-Mount als
      andere Mount-Klasse eingeordnet (§1) — kein Code-Fix, nur Befund.
- [x] `make docs-check` grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `apps/api/Dockerfile` | 2 neue Stages (`benchmark`, `vuln`) | hermetische Basis für beide Targets |
| `apps/api/Makefile` | `benchmark-smoke` umgeschrieben | build+run statt Bind-Mount |
| `Makefile` (root) | `vuln-check` umgeschrieben | build+run statt Bind-Mount |

## 4. Trigger

- **`in-progress`:** nach `slice-020` (Digest-Pinning muss stehen, die
  neuen Stages erben von `deps`, das bereits gepinnt ist).
- **Rückführung:** falls das Ausgabe-Format nach der Umstellung nicht
  mehr zum bestehenden Parser passt (unwahrscheinlich, da `docker run`
  gegen ein bereits gebautes Image dieselbe Shell-Umgebung liefert wie
  der bisherige Bind-Mount-Lauf) — dann Rückbau, kein Parser-Umbau in
  diesem Slice.

## 5. Closure-Trigger

DoD vollständig + `make gates` grün + `make docs-check` grün +
Closure-Notiz + `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Root-Ownership-Risiko bei `mutation-report`** (nicht Teil dieses
  Slices, siehe §1) — `gremlins unleash --output=.tmp/mutation/…json`
  schreibt als root direkt in den (heute noch) bind-gemounteten Baum;
  ungeklärt bis Tranche 9c. — **Ausgang:** weiter offen → Tranche 9c.
- **Kein ADR nötig.** Reine Build-Infrastruktur-Härtung.

## 7. Closure-Notiz (nach `done/`)

Zwei neue hermetische Stages in `apps/api/Dockerfile` (`benchmark`,
`vuln`), beide `FROM deps` mit `COPY . .` — **kein** `RUN` des eigentlichen
Prüflaufs in der Stage selbst: Der Lauf passiert bei `docker run` gegen
das gebaute Image, mit sauberem, unpräfigiertem stdout (verifiziert
gegen echte Ausgabe — siehe unten). `benchmark-smoke`
(`apps/api/Makefile`) und `vuln-check` (root-`Makefile`) auf
`docker build --target … && docker run --rm …` umgestellt, Bind-Mount
entfernt. `apps/api/Makefile`s `clean`-Target um die neuen (und den
zuvor schon fehlenden `go-race`-) Image-Tags ergänzt.

**Audit-Befund `image-scan`:** `docker.sock` + Trivy-Cache-Mounts bleiben
unverändert — andere Mount-Klasse (Daemon-Zugriff zum Scannen bereits
gebauter Images, kein Quellbaum-Mount), von der „kein Mount"-Regel für
Prüfläufe gegen den Arbeitsbaum nicht erfasst. Kein MR nötig
(Regelwerk-Schweigen zu dieser Mount-Klasse ist erlaubend).

**Verifikation:** `make benchmark-smoke` (in `apps/api/`) direkt
getestet — sauberer `go test`-Output ohne BuildKit-Präfix. Vollständiger
`make api-benchmark-smoke`-Pfad (root) getestet: `check-bench-budgets.mjs`
parst den Output unverändert korrekt (6 Go-Benches, alle unter Budget).
`make vuln-check` getestet: `govulncheck` läuft, „No vulnerabilities
found." Vollständiger `make gates`-Lauf grün. Test-Image-Tags
(`go-benchmark`, `go-vuln`) nach Verifikation gelöscht. `make docs-check`
— 0 Befunde (nach Marker-Rücksetzung).

**Steering-Loop-Lerneintrag:** Der naive erste Ansatz („Prüflauf als
`RUN`-Schritt in der Stage, wie bei `test`/`race`") hätte den
Bench-Output mit BuildKit-Zeilenpräfixen (`#N <timestamp> `)
zurückgegeben — der verankerte `^Benchmark`-Regex in
`check-bench-budgets.mjs` hätte dann keine Zeile mehr gematcht, und der
Fehler wäre erst beim tatsächlichen `make gates`-Lauf sichtbar geworden,
nicht beim Docker-Build selbst (der wäre grün geblieben). Das
Build-vs-Run-Split (Stage bäckt nur die Quellen ein, der eigentliche Lauf
passiert bei `docker run`) ist der Unterschied zwischen einem Gate, das
grün wird, weil sein Build durchläuft, und einem, dessen Output noch
maschinenlesbar ist.

**Folge-Slices:** Tranche 9c (`fuzz-check`/`mutation-report`, Export-Stage
für Schreib-Rückweg) — noch nicht geschnitten.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Dockerfiles + Makefiles (Werkzeug/Prozess)

Reine Build-Infrastruktur-Härtung, kein Anwendungscode geändert. Ohne
ADR — keine Architekturentscheidung, kein Gate gesenkt.
