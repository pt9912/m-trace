# Slice 010: d-check `planning.waves` gegen die Wellen-Register-Invariante aktivieren

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle (Werkzeug/Harness).

**Bezug:** `.d-check.yml`, `docs/plan/planning/in-progress/roadmap.md`,
d-check-Modul `planning` (dritte Fähigkeit, `planning.waves.dir`, seit v0.59.0,
Kardinalitäts-Modus `planning.waves.mode` seit v0.62.0). Ausgangspunkt: der
d-check-Versions-Bump v0.51.1 → v0.75.0.

**Autor:** Nachgang d-check-Versions-Bump. **Datum:** 2026-09-13.

---

## 1. Ziel

`.d-check.yml` aktiviert bereits die Basis-Fähigkeit des Moduls `planning`
nicht — weder die Slice/Marker-Invariante noch die Wellen-Register-Invariante
sind heute konfiguriert. Dieser Slice schaltet die **dritte Fähigkeit**
(`planning.waves`) zu: sie hält die Roadmap-Aussagen über Wellen — „Aktuelle
Welle" nennt X, „Abgeschlossene Wellen" listet Y mit Ergebnisnotiz Z — gegen
die tatsächlich vorhandenen flachen Wellendokumente. Bisher ist diese
Konsistenz reine Konvention (MR-007) ohne Gate; ein falsch verlinkter oder
vergessener Eintrag in der Roadmap-Tabelle fiele erst einem Menschen auf.

## 2. Definition of Done

- [ ] `.d-check.yml` führt einen `planning`-Block mit `roadmap:
      docs/plan/planning/in-progress/roadmap.md` und `waves.dir:
      docs/plan/planning`.
- [ ] `planning` ist in `modules:` aktiviert (additiv zur bestehenden Liste).
- [ ] `make docs-check` (bzw. das dedizierte `make doc-planning`-Target aus
      `d-check.mk`) läuft grün gegen den aktuellen Bestand (Positiv-Probe).
- [ ] Negativ-Probe dokumentiert: eine bewusst kaputte Register-Zeile oder ein
      fehlendes `waves.dir` erzeugt nachweislich `wave-drift` /
      `wave-results-missing` (Belegpflicht analog `slice-009`).
- [ ] `make gates` bleibt grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.d-check.yml` | update | `planning`-Block ergänzen, `planning` in `modules:` aufnehmen |
| `docs/plan/planning/in-progress/roadmap.md` | ggf. update | Nur falls die Positiv-Probe eine bisher unbemerkte Diskrepanz findet (siehe §6) |

**Bereits geklärt (Recherche vor Schnitt, 2026-09-13):** `waves.dir` erwartet
**eine** Wurzel und liest deren **eigenes** `done/`-Unterverzeichnis für
abgeschlossene Wellen-Dokumente + Ergebnisnotizen mit — kein rekursiver Scan
des ganzen Baums, sondern exakt die Konvention `<dir>` (aktiv, flach) +
`<dir>/done` (abgeschlossen), die MR-007 für m-trace bereits beschreibt
(„Wellen `welle-<NN>.md` flach → `done/` neben `welle-<NN>-results.md`").
`waves.dir: docs/plan/planning` passt damit **ohne Anpassung** auf den
bestehenden Bestand — empirisch verifiziert (siehe unten), nicht nur aus der
Doku hergeleitet.

**Verifikation (Probe-Lauf gegen den echten Baum, `d-check:v0.75.0`,
`--enable planning`, `waves.dir: docs/plan/planning`):** `126 Datei(en)
geprüft, 0 Befund(e)` — die vorhandene Zeile für `welle-01` in der
„Abgeschlossene Wellen"-Tabelle findet ihre `welle-01-results.md` in `done/`,
die „Nächste Wellen"-Vorschau-Zeile trägt einen Namen statt einer
`welle-<n>`-Kennung (zählt laut Spezifikation nicht als Vorschau-Kennung), und
es liegt kein aktives, flaches Wellendokument vor — konsistent mit dem
Ruhe-Marker „Keine aktive Welle." im Aktiv-Abschnitt.

**Negativ-Probe (bereits einmal ausgeführt, zur Reproduktion vorgemerkt):**
`waves.dir` auf ein Verzeichnis **ohne** eigenes `done/`-Unterverzeichnis
gezeigt (`docs/plan/planning/next`) meldet sofort
`wave-drift: Wellen-Verzeichnis docs/plan/planning/next/done fehlt oder ist
unlesbar (fail-closed)`, Exit 1 — belegt, dass das Modul nicht blind grün
läuft, sobald der Pfad nicht mehr passt.

**Offen für die Implementierung:** `planning.waves.mode` (Default `one`)
genügt für m-traces Ein-Wellen-Betrieb; `many` wäre nur nötig, würde m-trace je
mehrere Wellen gleichzeitig offen führen (aktuell nicht der Fall, kein
Änderungsbedarf hier).

## 4. Trigger

- **`in-progress`:** jederzeit — reine Werkzeug-/Config-Änderung, kein
  Produktcode betroffen, kein Abhängigkeits-Trigger.
- **Rückführung `in-progress` → `open`:** falls die Positiv-Probe eine
  bestehende Roadmap-Abweichung aufdeckt, die erst eine Owner-Entscheidung
  braucht (z. B. eine Umbenennung der Register-Tabellen-Form) — dann zurück,
  bis die Entscheidung steht.

## 5. Closure-Trigger

DoD vollständig + `make gates` grün + Positiv- und Negativ-Probe belegt +
Closure-Notiz + `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Zukünftige Produkt-Folgewelle als erster echter Test.** Der heutige
  Bestand hat keine aktive Welle — das Gate ist bis zur nächsten
  geschnittenen Welle rein defensiv (verhindert künftige Drift, prüft aber
  keinen echten Wellen-Zyklus). Der erste scharfe Test kommt erst mit
  `welle-02` (siehe Roadmap „Produkt-Folgewelle, noch ungeschnitten").
- **`slice-glob`/Marker-Basisfähigkeit bleibt bewusst außen vor.** Dieser
  Slice aktiviert nur die dritte Fähigkeit (`waves`), nicht die erste
  (Marker ⟺ `slice-*` im Roadmap-Verzeichnis). Eine gemeinsame Aktivierung
  wäre ein größerer Schnitt (eigener Slice), weil sie eine andere Grundmenge
  prüft (`in-progress/`-Inhalt statt `docs/plan/planning/`-Inhalt) und einen
  eigenen Negativ-Probe-Aufwand hätte.
- **Kein ADR nötig.** Reine Gate-Ergänzung, keine gesenkte Prüfung, keine
  Architekturentscheidung — passt in `modules:` additiv.

## 7. Closure-Notiz (nach `done/`)

`.d-check.yml` führt jetzt `planning` in `modules:` und einen `planning`-Block
mit `roadmap: docs/plan/planning/in-progress/roadmap.md` und
`waves.dir: docs/plan/planning`. `waves.dir` passte wie in §3 vorab
recherchiert ohne Anpassung — keine Überraschung in der Umsetzung selbst.

**Eine Annahme aus §3 war unpräzise und wurde in der Umsetzung korrigiert:**
„Die erste Fähigkeit (Marker/`slice-glob`) ist bewusst nicht geschnitten" —
das ist so nicht trennbar. Sobald `planning.roadmap` gesetzt ist, ist die
Marker-Invariante **immer** aktiv; nur `waves`/`closure`/`observations`
darunter sind selbst noch einmal opt-in. Der erste Positiv-Lauf bestätigte
das sofort und unmittelbar handfest: `planning-drift`, weil dieser Slice
selbst (als `slice-010-*.md`) zum Zeitpunkt des Laufs in
`docs/plan/planning/in-progress/` lag, während die Roadmap noch „Keine aktive
Welle" sagte — ein **echter**, korrekt erkannter Zustand, kein Fehlalarm. Der
Fund verschwindet von selbst, sobald der Slice (dieser Commit) nach `done/`
wandert und `in-progress/` wieder leer ist — verifiziert unten.

**Negativ-Probe** (`waves.dir` auf ein Verzeichnis ohne eigenes
`done/`-Unterverzeichnis gezeigt, `docs/plan/planning/next`):
`wave-drift: Wellen-Verzeichnis docs/plan/planning/next/done fehlt oder ist
unlesbar (fail-closed)`, Exit 1 — reproduziert wie in §3 vorab belegt.

**Verifikation nach dem Move nach `done/`:** `make docs-check` — 0 Befunde
(kein `planning-drift`, kein `wave-drift`). `make gates` grün.

**Steering-Loop-Lerneintrag:** Bei zusammengesetzten Opt-in-Modulen (Fähigkeit
N baut auf einer nicht abschaltbaren Basis-Fähigkeit auf) vor dem Schnitt
knapp gegen den echten Baum probieren, **nicht nur** die Ziel-Fähigkeit
isoliert — der Nebenbefund (Marker-Check) wäre sonst erst beim ersten
`make gates`-Lauf aufgefallen, nicht schon in der Planungsphase. Für künftige
d-check-Modul-Slices: immer den vollen `--enable <modul>`-Lauf gegen den
Ist-Zustand vorab fahren, nicht nur den dokumentierten Einzel-Schlüssel
gedanklich isolieren.

**Folge-Slices:** keine unmittelbaren — die erste Fähigkeit (Marker) ist jetzt
faktisch mit-aktiv und deckt bereits Zukünftiges ab. Ein Folge-Slice für
`planning.waves.mode: many` ist erst nötig, wenn m-trace je mehrere Wellen
parallel offen führt (heute nicht der Fall).

## 8. Sub-Area-Modus-Begründung

### Sub-Area: d-check-Konfiguration (Werkzeug)

Reine Config-/Harness-Änderung, kein Produktcode, keine Spec, kein
Requirement berührt. Ohne Welle (Modul 5 „Wartung/Architektur"), ohne ADR —
es wird kein Gate gesenkt, sondern eines zusätzlich scharf geschaltet.
