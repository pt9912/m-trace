# Extra Quality Gates

> **Status**: Master-Backlog fuer die sechs vorgeschlagenen Quality-
> Gates. Aktiv konkretisiert in zwei Folge-Plaenen:
>
> - [`plan-0.8.5.md`](../done/plan-0.8.5.md) — **Wave 1**: §3.1 govulncheck
>   + Container-Scan, §3.4 Generated-Artifact-Drift-Gate
>   (deterministisch, schnell, PR-blockierend; vor `0.9.0`).
> - [`plan-0.9.5.md`](../done/plan-0.9.5.md) — **Wave 2**: §3.2 Benchmark-
>   Smoke, §3.3 Nightly-`benchstat`, §3.5 Fuzzing/Property,
>   §3.6 Mutation Testing (statistisch + langlaufend; nach `0.9.0`).
>
> Dieses Dokument bleibt als Backlog/Master-Uebersicht erhalten und
> ist die Quelle fuer DoD-Wortlaute, Implementierungsleitplanken und
> die Benchmarking-Policy. Neue Quality-Gate-Vorschlaege landen
> zuerst hier, bevor sie in einem Folge-Plan konkretisiert werden.

## 1. Zweck

m-trace hat bereits harte Gates fuer Tests, Coverage, SOLID-nahe
Lint-Regeln, Architekturgrenzen, Schema-Validierung, Contract-Drift,
SDK-Packaging, SDK-Performance und mehrere Smoke-Pfade. Diese Notiz
beschreibt die naechste sinnvolle Gate-Schicht:

- Security-Risiken frueher sichtbar machen.
- Performance-Regressionen messbar machen, ohne PRs durch Runner-
  Rauschen unzuverlaessig zu blockieren.
- Generierte Artefakte synchron halten.
- Testqualitaet an kritischen Stellen ueber Coverage hinaus pruefen.

Nicht-Ziel: jedes Gate sofort in `make gates` aufnehmen. Einige
Pruefungen sind bewusst Nightly- oder Release-Kandidaten.

## 2. Gate-Matrix

| Prio | Gate | Zielpfad | Blockierend | Hauptnutzen |
| ---: | ---- | -------- | ----------- | ----------- |
| 1 | `govulncheck` + Container-Scan | PR + Release | ja | bekannte CVEs frueh stoppen |
| 2 | Benchmark-Smoke mit festen Budgets | PR | ja | grobe Performance-Regressionen stoppen |
| 3 | Nightly `benchstat`-Regressionen | Nightly + Release | ja fuer Release | statistisch belastbare Trends erkennen |
| 4 | Generated-Artifact-Drift-Gate | PR | ja | Quellen und generierte Dateien synchron halten |
| 5 | Selektives Fuzzing / Property Tests | PR kurz, Nightly lang | ja fuer Seed-Corpus | Parser-/Validation-Robustheit erhoehen |
| 6 | Mutation Testing auf kritischen Modulen | Nightly / manuell | nein initial | Teststaerke sichtbar machen |
| 7 | WebRTC-Ton-Smoke (1-kHz-FFT/Goertzel) | Nightly + lokal opt-in | nein | manuellen 1-kHz-Hoercheck (releasing.md §2.3) automatisieren |
| 8 | Last-/Soak-Smoke (k6 + Readback-Reconciliation) | Nightly (SLO) + on-demand (Soak) + lokal opt-in | nein | Lab-Lastfaehigkeit (NF-20/22/23), kein stiller Verlust, ADR-0005-Trigger-Evidenz |

## 3. Priorisierung

### 3.1 `govulncheck` + Container-Scan

**Entscheidung:** zuerst einfuehren. Der Nutzen ist hoch, die
Interpretation klar und das Gate passt gut zum bestehenden
Docker-/CI-Modell.

Scope:

- Go-Abhaengigkeiten unter `apps/api` mit `govulncheck ./...`.
- Runtime-Images fuer API, Dashboard und Analyzer-Service mit einem
  Image-Scanner wie Trivy oder Grype.
- Scan-Policy fuer Container: `CRITICAL` und `HIGH` blockieren;
  `MEDIUM` wird berichtet, aber nicht sofort blockierend.

Erste Ziel-Targets:

- `make vuln-check`
- `make image-scan`
- spaeter optional: `make security-gates`

DoD:

- Tool-Versionen sind gepinnt oder reproduzierbar bezogen.
- CI gibt maschinenlesbare Scan-Artefakte aus.
- False-Positive-/Ignore-Regeln liegen versioniert mit Begruendung vor.
- `make image-scan` baut die zu scannenden Runtime-Images im selben
  Lauf oder konsumiert eindeutig benannte Image-Tags aus einem
  vorangegangenen CI-Step. Dashboard und Analyzer-Service duerfen
  nicht implizit als vorhanden angenommen werden, weil `make build`
  aktuell nur API-Image plus TypeScript-Artefakte baut.
- Wenn Security-Gates PR-blockierend werden, ist der GitHub-Actions-
  Workflow explizit erweitert oder auf ein zentrales Target umgestellt,
  das diese Gates enthaelt. Ein neues Make-Target allein reicht nicht.

### 3.2 Benchmark-Smoke fuer API und Stream-Analyzer

**Entscheidung:** als Budget-Smoke in den PR-Pfad aufnehmen, aber
nicht als Mikrobenchmark-Baselinevergleich. PRs sollen nur bei klaren
Budget-Verletzungen scheitern.

API-Kandidaten:

- Event-Ingest: `RegisterPlaybackEventBatch` mit typischer Batch-
  Groesse und maximal erlaubter Batch-Groesse.
- SQLite-Ingest: persistierter Event-Batch inklusive Sequenzvergabe.
- Session-Listing/Pagination: typische Seiten und grosse Session-
  Menge.
- Cursor: Encode/Decode und Fehlerpfade.

Stream-Analyzer-Kandidaten:

- HLS Master-Manifest klein/gross.
- HLS Media-Manifest mit vielen Segmenten.
- DASH-MPD, sobald DASH-Analyse produktiv ist.
- SSRF-/URL-Klassifizierung mit typischen erlaubten und geblockten
  Eingaben.

Policy:

- Budgets sind absolute Obergrenzen, nicht Vergleich gegen letzten
  Commit.
- Budgets starten grosszuegig und werden erst nach Messhistorie
  geschaerft.
- Budget-Smokes messen Hot Paths mit synthetischen, repo-lokalen
  Fixtures; keine Netzwerkabhaengigkeit.
- PR-Budgets gelten zunaechst fuer GitHub Actions auf `ubuntu-24.04`.
  Jeder Lauf druckt mindestens Runner-OS, CPU-Modell und relevante
  Runtime-Versionen, damit Budget-Failures einordenbar bleiben.
- Neue oder geaenderte Budgets laufen erst fuer mehrere gruene CI-
  Beobachtungslaeufe nicht-blockierend mit, bevor sie PRs blockieren.

Erste Ziel-Targets:

- `make api-benchmark-smoke`
- `make analyzer-benchmark-smoke`
- `make benchmark-smoke`

DoD:

- Jeder Smoke druckt Laufzeit, Allokationen oder Durchsatz lesbar aus.
- Budgetverletzung erzeugt eine eindeutige Fehlermeldung mit Ist/Soll.
- Die Fixtures sind stabil und versioniert.
- Wenn `make benchmark-smoke` PR-blockierend wird, ist der
  GitHub-Actions-Workflow explizit erweitert oder auf ein zentrales
  Target umgestellt, das den Smoke enthaelt.

### 3.3 Nightly `benchstat`-Regressionen

**Entscheidung:** nicht in den PR-Critical-Path. Benchmarks mit
Baselinevergleich brauchen Wiederholungen und stabile Auswertung.

Scope:

- Go-Benchmarks mit `go test -bench=. -benchmem -count=N`.
- Auswertung mit `benchstat` gegen eine gespeicherte Baseline.
- Initial nur API-Benchmarks; TypeScript-Benchmarks erst nach
  stabilen Budget-Smokes.

Regression Policy:

- Nightly markiert auffaellige Regressionen als CI-Failure.
- Release-Gate blockiert bei bestaetigter Regression.
- Einzelne laute Benchmarks duerfen quarantined werden, muessen aber
  im Plan vermerkt bleiben.

DoD:

- Baseline-Herkunft ist dokumentiert.
- `benchstat`-Output wird als CI-Artefakt gespeichert.
- Schwellen fuer Blocker sind explizit, z. B. >15 % Regression bei
  stabil signifikantem Ergebnis.

### 3.4 Generated-Artifact-Drift-Gate

**Entscheidung:** frueh einfuehren. Dieses Gate ist deterministisch und
passt gut zu PRs.

Kandidaten:

- Schema-DDL aus `apps/api/internal/storage/schema.yaml`.
- Contract-Fixtures aus `spec/contract-fixtures`.
- Public-API-Snapshots der TypeScript-Pakete.
- `contracts/sdk-compat.json` und versionsfuehrende Fixtures.

Bevorzugtes Muster:

1. Generierungs-/Sync-Targets ausfuehren.
2. `git diff --exit-code` auf die erwarteten Pfade.
3. Bei Diff mit klarer Meldung abbrechen.

Erste Ziel-Targets:

- `make generated-drift-check`
- optional paketweise Untertargets fuer Schema, Fixtures und
  Public-API.

DoD:

- Das Gate laeuft ohne Netzwerk.
- Es prueft nur deterministische Artefakte.
- Der Fehlertext nennt den konkreten Regenerierungsbefehl.
- Wenn `make generated-drift-check` PR-blockierend wird, ist der
  GitHub-Actions-Workflow explizit erweitert oder auf ein zentrales
  Target umgestellt, das den Check enthaelt.

### 3.5 Selektives Fuzzing / Property Tests

**Entscheidung:** selektiv, nicht repo-weit. Ziel sind Parser,
Decoder und Validierungslogik, bei denen Edge-Cases realistisch sind.

Go-Kandidaten:

- Cursor Encode/Decode.
- HTTP-Validation von Playback-Event-Batches.
- Event-Metadaten und Redaction.
- SRT-Health-Mapping.

TypeScript-Kandidaten:

- HLS Attribute Parser.
- HLS Master-/Media-Parser.
- SSRF-/URL-Klassifizierung.
- Result-Stability-Invarianten.

Policy:

- PR-Gate: kurzer Seed-Corpus-Lauf oder Property Tests mit begrenztem
  Case-Budget.
- Nightly: laengerer Fuzz-Lauf mit Artefakt fuer gefundene Crasher.
- Crasher werden als Regression-Fixtures eingecheckt.

DoD:

- Fuzz-Ziele sind klein und deterministisch.
- Seeds liegen im Repo.
- Gefundene Minimalfaelle werden in normale Unit-Tests ueberfuehrt.

### 3.6 Mutation Testing auf kritischen Modulen

**Entscheidung:** spaeter und nur punktuell. Mutation Testing ist
nuetzlich, aber teuer und anfangs oft noisy.

Kandidaten:

- Event-Validierung.
- Cursor-Logik.
- HLS-Parser; DASH-Parser erst, sobald DASH-Analyse produktiv ist.
- SRT-Health-Mapping.
- Security-relevante URL-/SSRF-Pruefung.

Policy:

- Kein repo-weites Gate.
- Start als Nightly- oder manueller Report.
- Erst blockierend machen, wenn Mutant-Survival-Rate stabil und
  False-Positive-Rate niedrig ist.

DoD:

- Kritische Module sind explizit allowlisted.
- Ergebnis wird als Trend verfolgt, nicht nur als Einzelwert.
- Threshold-Senkungen oder Scope-Aenderungen sind begruendungspflichtig.

### 3.7 Nightly-Security-Audit-Mirror

**Entscheidung:** Pendant zu §3.1, aber zeitbasiert. Die Push-/PR-
Gates aus §3.1 (`govulncheck`, `pnpm audit`, `trivy image scan`)
fangen nur Advisories, die zum Zeitpunkt eines Pushes bereits
veroeffentlicht waren. Neu publizierte Advisories zwischen zwei
Pushes bleiben unerkannt, bis jemand pusht. Ausloeser fuer die
Aufnahme: GHSA-77vg-94rm-hx3p (`devalue` DoS) wurde vier Tage nach
Release 0.22.0 publiziert und hat den naechsten Push gebrochen,
ohne dass die Vulnerability vorher sichtbar war.

Scope:

- Identischer Stack wie §3.1: `make vuln-check`, `make audit-ts`,
  `make image-scan`.
- Tagesgenaue Cadence (Cron 01:57 Europe/Berlin, gestaffelt mit
  den anderen Nightlies in der 01:xx-Stunde).
- Bei einem Finding wird automatisch ein GitHub-Issue eroeffnet
  (`security,audit,plan-0.8.5`) mit den letzten 40 Zeilen jedes
  Checks plus Verweis aufs Artefakt.

Policy:

- Drei Steps mit `continue-on-error: true`; das Issue bekommt alle
  Outcomes auf einmal, statt drei separate Issues bei
  Kombinations-Findings.
- Run faellt am Ende explizit, damit die Workflow-Liste in der
  GitHub-UI rot leuchtet, solange das Issue offen ist.
- Manueller `workflow_dispatch`-Trigger fuer ad-hoc-Reruns nach
  einem grossen Advisory-Push.

DoD:

- Cron-Slot kollidiert nicht mit den existierenden Nightlies.
- Issue-Body enthaelt einen Reaction-Block, der die drei Fix-
  Pfade (Go-Bump / `pnpm.overrides` / Base-Image-Bump) jeweils mit
  einem konkreten Re-Verifikations-Target benennt.
- Artefakt-Retention deckt mindestens 30 Tage, damit auch nicht
  sofort bearbeitete Findings noch reproduzierbar sind.

### 3.8 WebRTC-Ton-Smoke (1-kHz-FFT)

**Entscheidung:** Automatisiert die *eng definierte* manuelle
Release-Abnahme „bestätigen, dass ein 1-kHz-Sinuston hörbar abspielt"
aus [`releasing.md`](../../../user/releasing.md) §2.3. Komplementaer zum
WebRTC-`getStats()`-Drift-Smoke (RAK-56, releasing.md §2.4.1), der nur
`bytesReceived>0` (Medien fliessen), nicht die Tonqualitaet prueft.
Plan: [`plan-0.22.4-webrtc-tone-smoke.md`](../done/plan-0.22.4-webrtc-tone-smoke.md).

Scope:

- `scripts/check-tone.mjs`: dependency-freier Goertzel-Einzel-Bin-DFT;
  Verdict ueber den Energie-Anteil (Parseval) des Ziel-Bands an der
  Gesamtenergie — Reinton ~0.5, abwesende Frequenz/Rauschen ~0.
- `make smoke-webrtc-tone` (`scripts/smoke-webrtc-tone.sh`): faehrt das
  `mtrace-webrtc`-Lab hoch, zieht den RTSP-Egress per ffmpeg im
  Lab-Netz (host-seitig ist nur WHEP erreichbar) und pipet das PCM an
  den Detektor.
- Nightly-Schritt in
  [`webrtc-drift.yml`](../../../../.github/workflows/webrtc-drift.yml)
  neben dem Drift-Smoke.

Policy:

- **Nicht-blockierend**: opt-in lokal, im Nightly `continue-on-error`
  — das WebRTC-Lab ist unter Last flaky (siehe
  [`plan-0.22.3-webrtc-drift`](../done/plan-0.22.3-webrtc-drift.md)
  §2), ein Ton-Failure soll weder PRs blocken noch den Drift-Befund
  maskieren. Nicht in `make gates`.
- Ersetzt **nicht** die perzeptuelle Operator-Abnahme („klingt/sieht
  das Demo im echten Browser richtig"); deckt nur den maschinell
  praezise pruefbaren „1-kHz-Ton vorhanden und dominant"-Teil ab.

DoD:

- Detektor deterministisch ohne Lab testbar (synthetische
  ffmpeg-Signale: Reinton PASS, falsche Frequenz/Stille/Rauschen FAIL).
- Smoke live gegen `mtrace-webrtc` verifiziert (Energie-Anteil ~0.5
  durch die reale Opus-Pipeline).
- Nightly-Schritt laeuft `continue-on-error` neben dem Drift-Schritt.

### 3.9 Last-/Soak-Smoke

**Entscheidung:** Die Hot-Path-Mikrobenchmarks (§3.2) messen Funktionen
isoliert gegen Budgets — nicht die Ingest→Persistenz→Read-Kette unter
echter Parallelität (NF-20/NF-22/NF-23). Der Last-Smoke schließt diese
Lücke als separater opt-in-Pfad (vgl.
[`docs/perf/budgets.md`](../../../perf/budgets.md) §7). Plan:
[`plan-0.22.5-load-smoke`](../done/plan-0.22.5-load-smoke.md).

Scope:

- `make smoke-load` (`scripts/smoke-load.sh` + `scripts/load/playback-events.k6.js`):
  k6 gegen `/api/playback-events`, dann **Readback-Reconciliation** gegen
  die echte `playback_events`-Tabelle (`events[]`-Array, nicht
  `event_count`) — `persisted >= accepted` = kein stiller Verlust.
- `MODE=capacity` (Limit angehoben, echte Kapazität) / `MODE=contract`
  (100/s, Limiter-Check); `LOAD_PROFILE=closed` (`--vus`, Decke finden)
  / `LOAD_PROFILE=open` (`make smoke-load-slo`, constant-arrival-rate,
  runner-stabile SLO: p95 < Budget + dropped_iterations).
- `make smoke-soak` (`RETENTION_PROBE=1`): Read-Retention-p95 gegen 2 s
  ([ADR-0005](../../adr/0005-production-ops-backends.md) Trigger #3) —
  belastbar erst ab ≥ 10 Mio Events (Nightly-Soak ~Stunden).
- Nightly: [`load-smoke.yml`](../../../../.github/workflows/load-smoke.yml)
  fährt die SLO (open-loop) nightlich; der Soak läuft on-demand
  (`workflow_dispatch`, lange DURATION).

Policy:

- **Nicht-blockierend**, NICHT in `make gates` (lastabhängig flaky). Das
  Verdikt liest sich aus Job-Summary + Artefakt (Step-Outcome,
  k6-Summary, Reconciliation), nicht aus der Job-Farbe.
- Harte Smoke-Gates: kein stiller Verlust (Readback), Fehlerquote
  (Status ≠ 202/429, 429-frei gerechnet) ≤ `MAX_ERROR_PCT`, im
  open-Profil zusätzlich der k6-p95-/dropped-Threshold. Ein Lesefehler im
  Readback bricht INCONCLUSIVE (Exit 3), nie als Verlust.

DoD:

- Beide Modi + beide Profile lokal grün; Readback gegen die echte
  Tabelle, nicht `event_count`.
- Nightly-SLO-Schritt `continue-on-error`, Report als Artefakt + Summary.
- Soak liefert die ADR-0005-Trigger-#3-Evidenz (erst ≥ 10 Mio belastbar).

## 4. Benchmarking-Policy

Benchmarking bleibt zweigeteilt:

- **PR-Gate:** robuste Budget-Smokes mit festen, grosszuegigen
  Schwellwerten.
- **Nightly/Release:** statistische Regressionserkennung mit mehreren
  Wiederholungen und Baseline-Vergleich.

Ein PR sollte nicht daran scheitern, dass ein Mikrobenchmark auf einem
Shared Runner leicht schwankt. Blockierend im PR-Pfad sind nur klare
Budget-Verletzungen. Statistische Trends gehoeren in Nightly- und
Release-Gates.

## 5. Vorgeschlagene Einfuehrungsreihenfolge

1. Security-Gates (`govulncheck`, Container-Scan) als eigene Targets
   einfuehren und danach in CI verdrahten.
2. Generated-Artifact-Drift-Gate einfuehren, weil es deterministisch
   und schnell ist.
3. Benchmark-Smoke fuer API und Stream-Analyzer mit konservativen
   Budgets ergaenzen.
4. Nightly-Benchmark-Workflow mit `benchstat` und Baseline-Artefakt
   aufsetzen.
5. Property-/Fuzz-Ziele fuer Cursor, Parser und URL-Klassifizierung
   aufnehmen.
6. Mutation Testing als nicht-blockierenden Nightly-Report fuer wenige
   kritische Module starten.

## 6. Offene Entscheidungen

Wave 1 (`plan-0.8.5.md` Tranche 0):

- ✅ Container-Scanner: **Trivy** (`plan-0.8.5.md` §0.4 Default,
  Tranche-0-Commit `836452f`).
- ✅ Security-Gates separat: **`make security-gates`** als Wrapper
  von `vuln-check` + `audit-ts` + `image-scan`; eigener CI-Job
  `security` parallel zu `build` (`plan-0.8.5.md` Tranche 1,
  Commit `927555a`).

Wave 2 (`plan-0.9.5.md` Tranche 0):

- ✅ Benchmark-Baseline: **Git-Branch `benchmark-baseline`**
  (Default-Empfehlung; deterministische Historie + kein Retention-
  Limit; `plan-0.9.5.md` §1a Tranche-0-Commit).
- ✅ Initiale Performance-Budgets: in
  [`docs/perf/budgets.md`](../../../perf/budgets.md) dokumentiert —
  Architektur-basiert großzügig, Tranche-1-Beobachtungsphase
  schärft nach realen Messungen (`plan-0.9.5.md` §1a
  Tranche-0-Commit).
- ✅ `benchstat`-Regressions-Schwelle: **> 15 % auf statistisch
  signifikantem Benchmark** als Issue-Auto-Erstellung; **>=15 %
  bestätigt = Release-Block** (Tranche 2 implementiert das in
  `.github/workflows/benchmark.yml`; `extra-gates.md` §3.3 ist
  normativer Anker).
- ✅ Benchmark-Smoke-PR-Blockierung: nach fünf grünen
  `benchmark-observation.yml`-Runs (`25592982776`, `25621106187`,
  `25643426077`, `25704811721`, `25769811661`) ist
  `make benchmark-smoke` in `make gates` aufgenommen
  ([`plan-0.22.0`](../done/plan-0.22.0.md)). Der Nightly-Workflow läuft ohne
  `continue-on-error`.
- 🟡 Mutation-PR-Blockierung: weiter deferred. Der TS-Workflow-Scope
  wurde in `0.22.0` von `@npm9912/player-sdk` auf
  `@pt9912/player-sdk` korrigiert; erst drei reale Reports mit
  >70 % Mutation-Score pro Modul lösen die Blockierungsentscheidung
  aus.
