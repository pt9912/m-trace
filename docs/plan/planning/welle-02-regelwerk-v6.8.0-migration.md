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
| [`slice-014`](open/slice-014-agents-md-neubefuellung.md) | `AGENTS.md` komplett neu befüllen | Modul 3, 8, 9, 13 (Templates) |
| [`slice-015`](open/slice-015-harness-readme-sensoren-auslagerung.md) | `harness/README.md`: Sensors-Tabelle 3-spaltig + `harness/sensors/<target>.md`-Auslagerung + Leseordnung | Grundlagen (Sensors-Auslagerung, Templates) |
| Tranche 5 (noch nicht geschnitten) | MR für m-traces Multi-Datei-Technical-Schicht deklarieren | Modul 3 (Spec) |
| Tranche 6 (noch nicht geschnitten) | ADR-Re-Evaluierungs-Trigger-Audit (ADR-0001..0011) | Modul 4 (ADRs) |
| Tranche 7 (noch nicht geschnitten) | Slice-Template + Beobachtungs-Register (`planning.observations`) | Modul 5 (Planning-Harness) |
| Tranche 8 (noch nicht geschnitten) | Roadmap-Abschnitt „Aktuelle Welle" → „Offene Wellen" (Owner-Entscheidung) | Modul 6 (Roadmap) |
| [`slice-016`](done/slice-016-roadmap-offene-wellen-format.md) ✅ | Roadmap „Aktuelle Welle" → „Offene Wellen", Status-Zeile entfernt, Meilenstein-Anker | Modul 6 (Roadmap) |
| Tranche 9 (noch nicht geschnitten) | Docker-Harness-Audit (hermetische Build-/Test-Stages) | Modul 14 |
| Tranche 10 (noch nicht geschnitten) | d-check-Sensoren `targets`/`vcs`/`reviews` aktivieren | Templates (`.d-check.yml`), Modul 10 |
| Tranche 11 (noch nicht geschnitten) | Review-Harness-Templates nachziehen (Findings-Tabelle + `Klasse`-Spalte, Zitier-Form-Disziplin, zwei neue Reviewer-Skill-Fundklassen) | Modul 10, Templates (`docs/reviews/review-report.template.md`, `.harness/skills/reviewer.template.md`) |

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

1. **Spec-Straten-Form (Tranche 5).** Der neue Kanon verlangt zwingend alle
   drei Straten (`lastenheft.md` › `spezifikation.md` › `architektur.md`).
   m-trace führt vier getrennte „Technical"-Dateien
   (`backend-api-contract.md`, `browser-support.md`, `player-sdk.md`,
   `telemetry-model.md`) statt einer `spezifikation.md`. **Frage:** Eine
   MR-Adaption deklarieren, die diese Vier-Datei-Form als bewusste Abweichung
   festhält (kein Refactor der Spec-Dateien selbst) — oder die vier Dateien
   tatsächlich zu einer `spezifikation.md` zusammenführen? Letzteres wäre ein
   erheblich größerer, produktnaher Eingriff (RTM, Cross-Refs, `matrix`-Modul)
   und sprengt diese Welle.
2. **Roadmap-Terminologie (Tranche 8) — Mechanik bereits umgestellt, nur der
   Name ist noch offen.** Direkter Diff `roadmap.template.md` v3.5.1 gegen
   v6.8.0 gezogen (2026-09-13, beim Anlegen dieses Slices): der neue Kanon
   nennt den Abschnitt „Offene Wellen" und führt ihn als **Liste** von
   Wellen-Zeigern (`[<welle-id>](../<welle-id>.md)` je offener Wellen-Datei)
   **plus** den Ruhe-Marker **zusätzlich**, wenn `in-progress/` keinen Slice
   trägt — beides gleichzeitig ist dort ausdrücklich der Normalfall direkt
   nach einer Wellen-Eröffnung. Genau das haben wir in `.d-check.yml`
   (`planning.waves.mode: many`) und `roadmap.md` (Marker + Wellen-Zeiger im
   selben Abschnitt) beim Anlegen von `welle-02` bereits umgesetzt — nicht
   als Workaround, sondern weil es exakt der Kanon-Vorgabe entspricht.
   **Die verbleibende, echte Frage ist nur noch der Abschnitts-**Name**:**
   `## Aktuelle Welle` (heutiger Wortlaut, `heading:`-Default in
   `.d-check.yml` bereits implizit darauf gesetzt) beibehalten und als
   MR-Adaption dokumentieren, oder auf `## Offene Wellen` umbenennen (dann
   `heading:` in `.d-check.yml` explizit mitziehen)? Zwei weitere,
   unabhängig entscheidbare Format-Deltas aus demselben Template-Diff, die
   bei Gelegenheit dieser Tranche mitlaufen könnten: die
   `**Status:** Aktiv. **Letzte Änderung:** …`-Kopfzeile entfällt im neuen
   Template ersatzlos, und die Meilensteine-Status-Spalte trägt künftig
   „erreicht YYYY-MM-DD" **plus auflösbarem Beleg-Anker** (Tag, Workflow-Lauf,
   Ergebnis-Notiz) statt bloßem „erreicht"/„offen".
3. **Review-Report-Archivierung (Modul 10, bewusst Out-of-Scope in §6).** Der
   neue Kanon archiviert Review-Reports bei Slice-Closure vollständig
   (`done/slice-<Kennung>-archiv.zip`) statt sie lose in `docs/reviews/`
   liegen zu lassen — eine Verhaltensänderung gegenüber der seit `slice-007`
   gelebten, erst kürzlich „scharf geschalteten" Praxis. **Frage:** Umstellen
   (eigene Folge-Welle) oder bei der aktuellen, bewährten Form bleiben und die
   Abweichung dokumentieren?
4. **ADR-Trigger-Grandfathering (Tranche 6).** ADR-0001..0011 haben kein
   Re-Evaluierungs-Trigger-Feld. **Frage:** Analog `MR-002`
   (ADR-Pfad-Grandfathering) alle Bestands-ADRs pauschal grandfathern und nur
   neue ADRs zum Trigger-Feld verpflichten — oder rückwirkend je ADR ein
   Trigger nachtragen (bedeutet: Ergänzung in einer neuen, nachgelagerten
   Registry-Datei, **nicht** Edit der immutablen ADR-Bodies selbst,
   AGENTS.md §3.5)?

## 9. Closure-Notiz

<!-- Erst nach Welle-Abschluss füllen. Verweis auf welle-02-results.md. -->
