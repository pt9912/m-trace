# Welle 02 — Harness-Regelwerk-Baseline v3.5.1 → v6.8.0 — Closure-Notiz

**Welle:** welle-02
**Abschluss:** 2026-09-13
**Verantwortlich:** Owner (m-trace)

## Was wurde geliefert?

Alle elf Tranchen abgeschlossen, additiv-zuerst (Tranche 1) vor
Struktur-Umbau (2–4), dann unabhängige Registry-Ergänzungen (5, 6),
Mechanismus-Etablierung (7), Härtung (9, in drei Teilen), Sensor-
Aktivierung (10) und Prozess-Nachführung (11):

- `slice-012`: `.harness/baseline/v6.8.0/` vendort, `SHA256SUMS`
  verifiziert.
- `slice-013`: `harness/conventions.md` → Index +
  `harness/conventions/MR-<NNN>-*.md`-Einzeldateien.
- `slice-014`: `AGENTS.md` komplett gegen das Template neu befüllt (9-Rang
  Source-Precedence, §3.7 Kommentar-Disziplin, §3.4-ADR-Verbot in
  `spec/architecture.md` — Audit 0 Treffer).
- `slice-015`: `harness/README.md` als alleiniger Gate-Index (14 Gates +
  3 Werkzeuge), `harness/sensors/gates.md`.
- `slice-016`: Roadmap „Aktuelle Welle" → „Offene Wellen".
- `slice-017`: `MR-008` (Spec-Straten-Vier-Datei-Form).
- `slice-018`: `MR-009` (ADR-0001..0008 Trigger-Grandfathering — Audit
  ergab acht statt der geschätzten elf ADRs).
- `slice-019`: Beobachtungs-Register etabliert
  (`docs/plan/planning/observations/`, Kürzel-Spalte in
  `harness/conventions.md`), `docs/plan/planning/README.md` neu (fehlte
  komplett).
- `slice-020`–`022`: Docker-Harness-Audit — Digest-Pinning aller
  Base-Images + `harness/image-hash.txt` (020), hermetische
  `benchmark-smoke`/`vuln-check`-Stages (021), hermetische
  `fuzz-check`/`mutation-report`-Stages mit Schreib-Rückweg-Export (022).
- `slice-023`: d-check-Modul `reviews` aktiviert.
- `slice-024`: `.harness/skills/reviewer.md` gegen das Template
  nachgezogen (zwei neue HIGH-Klassen, `klasse`-Feld, Zitier-Form-
  Disziplin).

Nebenbei (nicht als eigene Tranche geplant, aber direkt im Zuge der
Migration gefunden und behoben): `.harness/baseline/v3.5.0/` entfernt
(überholtes Audit-Material), `docs/plan/adr/README.md` neu angelegt
(fehlte komplett), `docs/plan/carveouts/README.md` von stale
v3.5.1-/Kurs-Referenzen bereinigt, `docs/plan/planning/in-progress/roadmap.md`
vollständig gegen `modul-06-roadmap.md` nachgezogen (Zitatsätze,
Liste+Marker-Form, „Slices ohne Welle"-Sektion gestrichen — neue
Baseline-Regel, die es in v3.5.1 noch nicht gab).

## Was hat funktioniert?

- **Additiv-zuerst** hat getragen: Tranche 1 (Vendoring) brach nichts
  Bestehendes, die Root-Artefakte liefen erst danach um.
- **Owner-Entscheidungen explizit sammeln statt versteckt entscheiden**
  (welle-02 §8) — sechs echte Entscheidungspunkte über die Welle verteilt,
  jede einzeln mit `AskUserQuestion` geklärt, bevor der jeweilige Slice
  geschnitten wurde. Kein Fall, in dem eine Annahme sich später als falsch
  herausstellte.
- **Digest-Pinning + hermetische Build-Stages** (`slice-020`–`022`) sind
  jetzt ein wiederverwendbares Muster (`FROM deps AS <stage>` + `COPY .
  .`, kein `RUN` des Prüflaufs selbst; `docker create`/`start -a`/`cp`/`rm`
  für Schreib-Rückweg) — nicht nur eine einmalige Reparatur.
- **Isoliert testen vor dem Einbau** (d-check-Module gegen den Bestand,
  Docker-Digests gegen lokale Caches) hat in jedem Fall Überraschungen
  vor dem Commit abgefangen, nicht danach.

## Was ging anders als geplant?

- **Tranche 9 (Docker-Harness-Audit)** war als eine Tranche geplant,
  wurde aber in drei Slices gesplittet (`020`/`021`/`022`) — der
  ursprüngliche Umfang („volle Rearchitektur") war größer als eine
  einzelne, sicher verifizierbare Änderung trägt. Jeder Teil einzeln
  gegen `make gates` verifiziert statt eines einzigen großen,
  schwerer zu prüfenden Commits.
- **Tranche 6s Geltungsbereich** war eine Vorab-Schätzung
  („ADR-0001..0011"), die sich beim Audit als zu groß herausstellte
  (tatsächlich nur 0001..0008) — dieselbe Lektion wie in mehreren
  anderen Slices dieser Welle: ein `grep`/Audit vor dem Schreiben der
  MR-Datei ersetzt keine Schätzung im Plantext.
- **Tranche 10s `targets`-Sensor** wurde zurückgestellt (§8 Punkt 5) —
  der Testlauf zeigte 90 `gate-undocumented`-Funde gegen m-traces
  ~90-Target-Makefile, eine vollständige `exempt-targets`-Pflege wäre ein
  eigener, größerer Aufwand gewesen.
- **`planning.observations`** (im ursprünglichen Tranche-7-Titel
  genannt) existiert im aktuell gepinnten `d-check v0.75.0` noch nicht —
  das Register (`slice-019`) ist reine Ablage-Konvention, mechanische
  Durchsetzung bleibt ein Folge-Punkt für einen künftigen Tool-Upgrade.

## Steering-Loop-Einträge

Keine — das Beobachtungs-Register wurde in dieser Welle selbst erst
etabliert (`slice-019`); es liegen noch keine Beobachtungen vor, die 3×
erreicht haben könnten.

## Beobachtungs-Register (Zeiger)

Der Zähler steht in
[`observations/README.md`](../observations/README.md) (m-traces Ablageform,
abweichend vom Kurs-Vorbild `observations.md` als Einzeldatei — m-trace
folgt der Baseline-Regelwerk-Form `docs/plan/planning/observations/BEO-<KUERZEL>/<slug>/`
direkt, siehe `slice-019`). Aktuell: keine offenen Beobachtungen.

## Folge-Slices

Keine unmittelbar angelegten Folge-Slices. Offen, aber bewusst nicht
geschnitten (kein akuter Bedarf):

- `targets`-Sensor (welle-02 §8 Punkt 5) — bei Bedarf eigener Slice mit
  `exempt-targets`-Pflege.
- Review-Report-Archivierung (`done/slice-<Kennung>-archiv.zip` statt
  loser `docs/reviews/`-Dateien, welle-02 §8 Punkt 3) — offene
  Owner-Frage, nicht entschieden in dieser Welle.
- Fuzz-/Mutation-Docker-Muster als Vorlage für künftige
  Schreib-Rückweg-Gates (kein Slice, nur Präzedenzfall).

## Verifikation (Closure-Trigger)

- `make gates` vollständig grün (letzter Lauf: alle Sub-Gates —
  `api-race`, `ts-test`, `lint`, `coverage-gate`, `arch-check`,
  `schema-validate`, `generated-drift-check`,
  `schema-generate-postgres-check`, `sdk-pack-smoke`,
  `sdk-performance-smoke`, `benchmark-smoke`, `docs-check`,
  `lint-variante-b`, `verify-closure-notes` — 0 Befunde).
- `make docs-check` — 0 Befunde (154 Dateien geprüft, Marker-Zustand
  „Keine aktive Welle" korrekt).
- `.harness/baseline/v6.8.0/SHA256SUMS` verifiziert (`slice-012`).
- `harness/image-hash.txt` erzeugt und verifiziert (`slice-020`) —
  Digest je Runtime-Image (`api`/`dashboard`/`analyzer-service`).
- Kein Modul 12 (Replay-Evaluierung) — kein Golden-Set-Lauf anwendbar,
  m-trace hat keinen nicht-deterministischen Modell-Kern (welle-02 §6).
