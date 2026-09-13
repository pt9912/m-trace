# Welle 02: Harness-Regelwerk-Baseline v3.5.1 → v6.8.0

**Lifecycle:** Die aktive Welle liegt flach unter `docs/plan/planning/`; bei
Closure wandert diese Datei per `git mv` nach `done/` (neben ihre
`welle-02-results.md`). Der Zustand ist die Verzeichnis-Position — kein
Status-Feld.

**Zielmeilenstein:** kein Meilenstein-Bezug (Harness/Prozess-Arbeit, keine
User-Surface).

**Verantwortlich:** Owner (m-trace). **Datum:** 2026-09-13.

---

## 1. Welle-Ziel

Die vendorte, integritäts-geprüfte Regelwerk-Baseline
(`.harness/baseline/v3.5.1/`) durch die aktuelle `v6.8.0`-Fassung ersetzen und
die Root-Artefakte (`AGENTS.md`, `harness/conventions.md`, `harness/README.md`,
`.d-check.yml`) so weit nachziehen, dass sie **ohne Divergenz-Steuer** zum neuen
Kanon passen — dieselbe Zielformulierung wie bei der v3.0→v3.5.0-Migration
(`done/plan-harness-v3.5.0-migration.md` §1), nur einen Kanon-Sprung weiter.

**Größenordnung, damit niemand überrascht wird:** Zwischen v3.5.1 (Kurs-Welle 33)
und v6.8.0 (Kurs-Welle 135) liegen 102 Kurs-Wellen. Das ist kein Patch-Re-Vendor
wie `slice-008` (v3.5.0 → v3.5.1) — mehrere Modul-Grenzen wurden neu gezogen
(Architektur wandert von Modul 4 zu Modul 3), eine komplette Datei-Klasse wird
zum Index (`harness/conventions.md`), und mindestens zwei Stellen verlangen eine
**Owner-Entscheidung**, keine mechanische Übernahme (§8).

**Ausgangsrecherche:** Vier parallele Recherche-Durchgänge gegen die
SHA256-verifizierten `v6.8.0`-Release-Assets
(`lab-regelwerk.zip`/`SHA256SUMS` von
[`github.com/pt9912/ai-harness-course/releases/tag/v6.8.0`](https://github.com/pt9912/ai-harness-course/releases/tag/v6.8.0))
gegen den vendorten `v3.5.1`-Bestand, plus Templates-Abgleich gegen die
aktuellen Root-Dateien. Ergebnisse sind in die Tranchen unten eingearbeitet;
Rohbefunde nicht separat archiviert (Fork-Ergebnisse dieser Session).

**Nicht-normativer Hinweis für die Tranchen-Umsetzung:** Ein anderes,
maschinenlokales Repo hat dieselbe Baseline-Linie bereits **inkrementell**
durchlaufen (v3.1.0 → … → v6.7.2, in mehreren Re-Baseline-Schritten statt
einem großen Sprung) und liegt live vor — u. a. mit `harness/conventions/`
als Index + `MR-<NNN>-<beschreibender-titel>.md`-Einzeldateien,
`harness/sensors/<target>.md`-Vertiefungsseiten und einem
„ÜBERHOLT: … → Nachfolger"-Kopfmarker-Muster für abgelöste Adaptionen (statt
Löschen). Wer Tranche 2/4 umsetzt, findet dort ein durchgespieltes Beispiel
der Zielform — als Orientierung, nicht als zu kopierende Quelle (anderes
Repo, andere Adaptions-Historie).

## 2. Trigger (Welle startet)

- Owner bestätigt den Tranchen-Schnitt (§4) und entscheidet die offenen Punkte
  in §8 — mindestens für Tranche 1–3, die übrigen Tranchen können ihre
  Entscheidung bei Bedarf verzögern (Reihenfolge ist additiv-zuerst, s. §7).

## 3. Closure-Trigger (Welle schließt)

- Alle Tranchen-Slices `done` (oder bewusst als eigenständige Folge-Welle
  abgespalten, siehe §6).
- `.harness/baseline/v6.8.0/` vendort + `SHA256SUMS` verifiziert.
- `AGENTS.md`, `harness/conventions.md` (+ `harness/conventions/`),
  `harness/README.md` gegen die `v6.8.0`-Templates abgeglichen, keine
  Duplikat-Gate-Tabelle mehr außerhalb `harness/README.md` §Sensors.
- `make gates` grün.
- Closure-Notiz in `welle-02-results.md`.

## 4. Slices in dieser Welle

Die Reihenfolge ist **additiv-zuerst, Struktur-Umbau später** — derselbe
Grundsatz wie in der v3.5.0-Migration (ADR-0009 Variante C). Nur
**Tranche 1** ist als konkreter Slice geschnitten (`open/`); die übrigen sind
hier als Tranchen mit Kennung, Titel und Kern-Aussage vorgemerkt und werden
geschnitten, sobald sie an der Reihe sind (§7) — ein Vorab-Schnitt aller zehn
Tranchen würde mehrere davon an ungeklärten Owner-Entscheidungen (§8) aufhängen,
bevor überhaupt Arbeit beginnt. Tranche 2–4 sind mittlerweile geschnitten (alle
drei Templates — `AGENTS.template.md`, `harness/conventions.template.md`,
`harness/README.template.md` — sind direkt diffgeprüft, nicht nur aus der
Fork-Recherche übernommen; Details je Slice-Text).

| Slice | Titel | Bezug (Regelwerk-Modul) |
|---|---|---|
| [`slice-012`](done/slice-012-harness-baseline-v6.8.0-vendoring.md) ✅ | Baseline v6.8.0 vendoren, `harness/conventions.md`-Zeiger umstellen | Grundlagen, Modul 2 (Bootstrap) |
| [`slice-013`](done/slice-013-conventions-md-zu-index.md) ✅ | `harness/conventions.md` → Index + `harness/conventions/MR-<NNN>-*.md` | Grundlagen (MR-Datei-Form) |
| [`slice-014`](done/slice-014-agents-md-neubefuellung.md) ✅ | `AGENTS.md` komplett neu befüllen | Modul 3, 8, 9, 13 (Templates) |
| [`slice-015`](done/slice-015-harness-readme-sensoren-auslagerung.md) ✅ | `harness/README.md`: Sensors-Tabelle 3-spaltig + `harness/sensors/<target>.md`-Auslagerung + Leseordnung | Grundlagen (Sensors-Auslagerung, Templates) |
| [`slice-017`](done/slice-017-spec-straten-mr.md) ✅ | MR für m-traces Multi-Datei-Technical-Schicht deklarieren | Modul 3 (Spec) |
| [`slice-018`](done/slice-018-adr-trigger-grandfathering.md) ✅ | ADR-Re-Evaluierungs-Trigger-Audit (ADR-0001..0008 grandfathered, `MR-009`) | Modul 4 (ADRs) |
| [`slice-019`](done/slice-019-beobachtungs-register.md) ✅ | Beobachtungs-Register etablieren (`docs/plan/planning/observations/`, Kürzel-Spalte) — mechanische `d-check`-Durchsetzung deferred, `planning.observations` existiert in `v0.75.0` (aktueller Pin) noch nicht | Modul 5/6 (Planning-Harness, Roadmap) |
| [`slice-016`](done/slice-016-roadmap-offene-wellen-format.md) ✅ | Roadmap „Aktuelle Welle" → „Offene Wellen", Status-Zeile entfernt, Meilenstein-Anker | Modul 6 (Roadmap) |
| [`slice-020`](done/slice-020-digest-pinning-image-hash.md) ✅ | Docker-Harness-Audit Teil 1: Base-Image-Digest-Pinning + `harness/image-hash.txt` | Modul 14 |
| [`slice-021`](done/slice-021-hermetic-benchmark-mount-abgrenzung.md) ✅ | Docker-Harness-Audit Teil 2: hermetische `benchmark-smoke`-Stage, Security-Scan-Mount-Abgrenzung (MR) | Modul 14 |
| [`slice-022`](done/slice-022-hermetic-fuzz-mutation-export.md) ✅ | Docker-Harness-Audit Teil 3: hermetische Gate-Stages für `apps/api` Fuzz/Mutation (Schreib-Rückweg-Export, root-Ownership-Risiko bei `mutation-report`) | Modul 14 |
| [`slice-023`](done/slice-023-dcheck-reviews-modul.md) ✅ | d-check-Sensor `reviews` aktivieren (`vcs`/`commits` bereits aktiv, `targets` zurückgestellt — s. §8) | Templates (`.d-check.yml`), Modul 10 |
| [`slice-024`](done/slice-024-review-harness-templates.md) ✅ | Review-Harness-Templates nachziehen (Findings-Tabelle + `Klasse`-Spalte, Zitier-Form-Disziplin, neue Reviewer-Skill-Fundklassen) | Modul 10, Templates (`docs/reviews/review-report.template.md`, `.harness/skills/reviewer.template.md`) |

**Bewusst nicht in dieser Welle:** Modul 12 (Replay-Evaluierung) — m-trace hat
keinen nicht-deterministischen Modell-Kern, aspirational bis zu einem
ML-/Scoring-Feature (§6).

## 5. Abhängigkeiten

- Tranche 2 (Konventions-Index) blockiert Tranche 10s `vcs`-Sensor (der prüft
  MR-Datei-Immutability — ohne Einzeldateien nichts zu prüfen).
- Tranche 7 (Beobachtungs-Register) blockiert Tranche 11s `Klasse`-Spalte im
  Review-Report (der Steering-Loop-Zähler braucht das Register als Ziel).
- Tranche 1 (Vendoring) blockiert alle übrigen Tranchen (sie zitieren durchweg
  den neuen Regelwerk-Wortlaut).
- Tranche 6 (ADR-Trigger-Audit) berührt **keine** bestehende ADR inhaltlich
  (ADRs sind nach `Accepted` immutable, AGENTS.md §3.5) — das Trigger-Feld wird
  über eine neue MR-Adaption geführt (Registry-Ansatz, analog `MR-002`), nicht
  über ADR-Edits.
- Wird von nichts blockiert — kann parallel zu Produkt-Wellen laufen (reines
  Harness/Prozess-Repo-Territorium, kein Produktcode betroffen).

## 6. Out-of-Scope für diese Welle

- **Modul 12 (Replay-Evaluierung).** Kein nicht-deterministischer Modell-Kern
  in m-trace heute — Roadmap-Kandidat, kein Slice.
- **Modul 15 (Observability) inhaltlich.** Nur redaktionelle Präzisierung
  (Meta-Observability des Agenten, nicht Produkt-Observability) — kein
  Handlungsbedarf, kein Slice.
- **ID-Schema-Umstellung** (`LH-FA-NN.a`/`SPEC-<NNN>` statt `F-`/`NF-`/`RAK-`)
  — bewusst getrennt von Tranche 5 gehalten: Tranche 5 deklariert nur die
  **MR-Abweichung** (mehrere Technical-Dateien statt einer `spezifikation.md`),
  ein ID-Schema-Wechsel wäre ein eigenständiger, deutlich größerer Slice mit
  eigenem Migrations-Aufwand über `spec/lastenheft.md` (RTM, `matrix`-Modul,
  `ids`-Modul-Patterns) und gehört — falls überhaupt gewollt — in eine eigene
  Folge-Welle, nicht hierher.
- **Review-Report-Archivierung ändern** (`done/slice-<Kennung>-archiv.zip`
  statt loser `docs/reviews/`-Dateien, Modul 10). Verhaltensänderung
  gegenüber der seit `slice-007` gelebten Praxis — als Owner-Entscheidung in
  §8 vorgemerkt. **Davon getrennt zu sehen** (und **in** Tranche 11
  enthalten, weil unabhängig von der Archivierungsfrage nötig): die
  Format-Nachführung von `docs/reviews/review-report.template.md` und
  `.harness/skills/reviewer.template.md` selbst — direkter Diff gezogen
  (2026-09-13, auf Nachfrage). Konkrete Deltas: die Findings-Tabelle bekommt
  echte Markdown-Tabellenform (`ID | Kategorie | Befund | Quelle | Pfad |
  Verifizierbar | Klasse`) statt Bullet-Listen, die neue `Klasse`-Spalte ist
  der Übergabepunkt in den Steering-Loop-Zähler (Beobachtungs-Register aus
  Tranche 7 — **Tranche 11 hängt deshalb an Tranche 7**); Zitier-Form-Disziplin
  „Kennung, nicht Adresse" (`slice-<Kennung>` statt Lifecycle-Pfad, `make
  <target>` statt Sensor-Datei-Link, Baseline-Stand als `vX.Y.Z` ·
  `regelwerk/<datei>.md` §<Abschnitt> in Inline-Code statt als Link — friert
  ein, verrottet nicht bei Baseline-Bumps); zwei neue HIGH-Fundklassen für
  `.harness/skills/reviewer.md` („Norm nur im Template-Kommentar", „Zustandsfeld
  trägt Chronik" — deckt sich mit AGENTS.md §3.7 aus Tranche 3).
  `.harness/skills/closure-note-reviewer.template.md` diffte dagegen nur
  Platzhalter-Generalisierung (`ADR-0011`/`check_closure_notes.py` → generische
  Verweise) — kein Handlungsbedarf für m-trace, unsere Fassung ist längst
  konkret ausgefüllt.

## 7. Tranchen-Sequenzbegründung

**T1 (Vendoring) zuerst, additiv/netzlos** — genau wie `W1` in der
v3.5.0-Migration: ändert nur `.harness/baseline/` + einen Zeiger in
`harness/README.md`, bricht nichts Bestehendes. **T2–T4 (Konventions-Index,
AGENTS.md, README) als zusammenhängender Root-Umbau** — sie zitieren einander
(README verweist auf die Konventions-Datei-Form, AGENTS.md auf beide) und
sollten in kurzer Folge laufen, damit das Repo nicht längere Zeit auf einem
inkonsistenten Zwischenstand steht. **T5/T6 (Spec-Straten-MR,
ADR-Trigger-Audit) sind unabhängig voneinander und von T2–T4** — reine
Registry-Ergänzungen, können parallel oder in beliebiger Reihenfolge laufen.
**T7 (Planning-Harness/Beobachtungs-Register) nach T2** — das neue
`harness/conventions/`-Layout ist die natürliche Ablage-Form, an der sich das
Beobachtungs-Register orientiert. **T8 (Roadmap-Terminologie) bewusst spät**,
weil sie eine Owner-Entscheidung mit Rückwirkung auf `slice-010`s frisch
gesetzte `.d-check.yml`-Konfiguration hat (§8). **T9 (Docker-Harness-Audit)
und T10 (d-check-Sensoren) zuletzt** — beide sind Absicherung/Härtung auf dem
dann bereits migrierten Stand, T10 hängt zusätzlich an T2. **T11
(Review-Harness-Templates) nach T7** — die neue `Klasse`-Spalte im
Review-Report braucht das Beobachtungs-Register als Ziel; unabhängig davon
kann T11 parallel zu T9/T10 laufen.

## 8. Offene Owner-Entscheidungen

Diese Welle enthält **echte Entscheidungen**, keine reinen
Mach-Fragen — bewusst hier gesammelt statt in einzelnen Slices versteckt:

1. **Spec-Straten-Form (Tranche 5) — entschieden (Owner, 2026-09-13):**
   MR-Adaption deklarieren, die m-traces Vier-Datei-Technical-Schicht
   (`backend-api-contract.md`, `browser-support.md`, `player-sdk.md`,
   `telemetry-model.md` statt einer `spezifikation.md`) als bewusste
   Abweichung festhält — kein Refactor der Spec-Dateien selbst. Umsetzung:
   [`slice-017`](done/slice-017-spec-straten-mr.md).
2. **Roadmap-Terminologie (Tranche 8) — entschieden und umgesetzt
   (`slice-016`, ergänzt beim Anlegen von `slice-015`):** Umbenennung auf
   `## Offene Wellen`, `heading:` in `.d-check.yml` mitgezogen,
   `**Status:** Aktiv. **Letzte Änderung:** …`-Kopfzeile entfernt,
   Meilensteine-Status-Spalte trägt `erreicht YYYY-MM-DD` plus
   auflösbarem Beleg-Anker. Keine offene Frage mehr.
3. **Review-Report-Archivierung (Modul 10, bewusst Out-of-Scope in §6).** Der
   neue Kanon archiviert Review-Reports bei Slice-Closure vollständig
   (`done/slice-<Kennung>-archiv.zip`) statt sie lose in `docs/reviews/`
   liegen zu lassen — eine Verhaltensänderung gegenüber der seit `slice-007`
   gelebten, erst kürzlich „scharf geschalteten" Praxis. **Frage:** Umstellen
   (eigene Folge-Welle) oder bei der aktuellen, bewährten Form bleiben und die
   Abweichung dokumentieren?
4. **ADR-Trigger-Grandfathering (Tranche 6) — entschieden (Owner,
   2026-09-13): pauschal grandfathern.** Korrigierter Befund beim Schneiden
   von `slice-018`: **nicht** alle elf, sondern nur ADR-0001..0008 fehlt der
   `## Re-Evaluierungs-Trigger`-Abschnitt — ADR-0009/-0010/-0011 haben ihn
   bereits (entstanden nach dessen Einführung in den `v3.5.0`-Templates).
   Umsetzung: [`slice-018`](done/slice-018-adr-trigger-grandfathering.md)
   (`MR-009`, analog `MR-002`).
5. **`targets`-Sensor zurückgestellt (Tranche 10) — entschieden (Owner,
   2026-09-13).** Ein Test gegen `harness/README.md` als Autoritäts-Doku
   ergab 90 `gate-undocumented`-Funde: m-trace hat ~90 Makefile-Targets
   (Smoke-Suiten, Release-/Image-Plumbing, Dev-Helfer), `harness/README.md`
   listet bewusst nur 17 (Gates + Werkzeuge). Eine vollständige
   `exempt-targets`-Liste für alle Nicht-Gate-Targets wäre selbst ein
   größerer, laufend zu pflegender Aufwand — zurückgestellt als eigener
   Folge-Slice, wenn Bedarf entsteht. `vcs`/`commits` sind bereits aktiv
   (Future-only-Sensoren, `.d-check.yml`); `reviews` läuft in
   [`slice-023`](done/slice-023-dcheck-reviews-modul.md).

## 9. Closure-Notiz

<!-- Erst nach Welle-Abschluss füllen. Verweis auf welle-02-results.md. -->
