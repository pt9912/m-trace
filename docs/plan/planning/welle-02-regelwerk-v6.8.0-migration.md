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
bevor überhaupt Arbeit beginnt.

| Slice | Titel | Bezug (Regelwerk-Modul) |
|---|---|---|
| [`slice-012`](open/slice-012-harness-baseline-v6.8.0-vendoring.md) | Baseline v6.8.0 vendoren, `harness/README.md`-Zeiger umstellen | Grundlagen, Modul 2 (Bootstrap) |
| Tranche 2 (noch nicht geschnitten) | `harness/conventions.md` → Index + `harness/conventions/MR-<NNN>-*.md` | Grundlagen (MR-Datei-Form) |
| Tranche 3 (noch nicht geschnitten) | `AGENTS.md` komplett neu befüllen | Modul 8, 9, 13 (Templates) |
| Tranche 4 (noch nicht geschnitten) | `harness/README.md` nachziehen (Leseordnung, Reviewer-Zeile, Sensors-Alleinstellung) | Templates |
| Tranche 5 (noch nicht geschnitten) | MR für m-traces Multi-Datei-Technical-Schicht deklarieren | Modul 3 (Spec) |
| Tranche 6 (noch nicht geschnitten) | ADR-Re-Evaluierungs-Trigger-Audit (ADR-0001..0011) | Modul 4 (ADRs) |
| Tranche 7 (noch nicht geschnitten) | Slice-Template + Beobachtungs-Register (`planning.observations`) | Modul 5 (Planning-Harness) |
| Tranche 8 (noch nicht geschnitten) | Roadmap-Abschnitt „Aktuelle Welle" → „Offene Wellen" (Owner-Entscheidung) | Modul 6 (Roadmap) |
| Tranche 9 (noch nicht geschnitten) | Docker-Harness-Audit (hermetische Build-/Test-Stages) | Modul 14 |
| Tranche 10 (noch nicht geschnitten) | d-check-Sensoren `targets`/`vcs`/`reviews` aktivieren | Templates (`.d-check.yml`), Modul 10 |

**Bewusst nicht in dieser Welle:** Modul 12 (Replay-Evaluierung) — m-trace hat
keinen nicht-deterministischen Modell-Kern, aspirational bis zu einem
ML-/Scoring-Feature (§6).

## 5. Abhängigkeiten

- Tranche 2 (Konventions-Index) blockiert Tranche 10s `vcs`-Sensor (der prüft
  MR-Datei-Immutability — ohne Einzeldateien nichts zu prüfen).
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
  §8 vorgemerkt, nicht automatisch in Tranche 10 mit übernommen.

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
dann bereits migrierten Stand, T10 hängt zusätzlich an T2.

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
2. **Roadmap-Terminologie (Tranche 8).** „Aktuelle Welle" (Singleton,
   `planning.waves.mode: one`, gerade in `slice-010` bewusst so gewählt) vs.
   „Offene Wellen" (Liste, `mode: many`, aktueller Kanon-Vorschlag). m-trace
   fährt bisher nie mehr als eine Welle gleichzeitig — der Umstieg ist
   **kosmetisch/terminologisch**, keine funktionale Notwendigkeit. **Frage:**
   Umbenennen (und `.d-check.yml`s `planning`-Block + `heading:`-Override
   nachziehen) oder bei „Aktuelle Welle" + `mode: one` bleiben und die
   Abweichung in Tranche 2 als weitere MR-Adaption festhalten?
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
