# Slice 015: `harness/README.md` — Sensors-Tabelle + `harness/sensors/`-Auslagerung, Leseordnung

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](../welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 4 von 10).

**Bezug:** `harness/README.md`, neu `harness/sensors/<target>.md` (bei Bedarf),
`AGENTS.md` (Tranche 3, Abhängigkeit — die dortige Gate-Tabellen-Entfernung
setzt voraus, dass hier der alleinige Index steht),
`.harness/baseline/v6.8.0/templates/harness/README.template.md`,
`.harness/baseline/v6.8.0/templates/harness/sensors/gate.template.md`,
Regelwerk `grundlagen-harness-dateien.md` §Sensors.

**Autor:** Nachgang Templates-Diff v3.5.1→v6.8.0. **Datum:** 2026-09-13.

---

## 1. Ziel

`harness/README.md` gegen das `v6.8.0`-Template nachziehen: die Sensors-Tabelle
bekommt eine dritte Spalte (`Bindung`), eine zweite Tabelle für
Nicht-Gate-Werkzeuge kommt dazu, Gates mit mehr als einem Satz Vertrag wandern
als Prosa nach `harness/sensors/<target>.md`, und eine neue „Leseordnung"-Sektion
schließt die Datei ab. Dies wird nach dem Umbau der **alleinige** Gate-Index
des Repos (Tranche 3 entfernt die Duplikat-Tabelle aus `AGENTS.md`).

## 2. Definition of Done

- [ ] **Sensors-Tabelle auf drei Spalten** `Target | Vertrag | Bindung`
      (m-traces aktuelle Tabelle hat nur `Target | Prüft` — **keine**
      Bindung-Spalte). Bindung nennt eine der vier kanonischen Klassen
      (ADR-Bindung, Carveout-Bindung, Kalibrierungs-Bindung,
      Reproduzierbarkeits-Bindung) **oder** eine repo-lokal deklarierte
      Zusatzklasse — m-trace hat bereits `## Sensor-Bindungsklassen` in
      `harness/conventions.md` (Requirement-Bindung `F-*`/`NF-*`/…,
      ADR-Bindung, Reproduzierbarkeits-Bindung über Image-Digests): diese
      drei Klassen jetzt **in die Tabellenzellen** ziehen, statt als separate
      Prosa-Sektion daneben zu stehen.
- [ ] **Target-Zelle trägt den nackten Namen, kein Argument.** Ein Aufruf wie
      `SLICE=<id>` gehört in eine Nachbarspalte/Prosa, nicht in die Code-Span
      des Targets — sonst meldet ein künftiger `targets`-Sensor (Tranche 10)
      `gate-undocumented`, als gäbe es die Zeile nicht (gemessen an d-check
      v0.74.1, im Regelwerk zitiert). Ein **verlinktes** Target
      (`[make X](sensors/X.md)`) ist davon nicht betroffen.
- [ ] **Kein Lauf-Status** (grün/rot) in der Tabelle — Lauf-Wahrheit lebt in
      CI. Ein strukturell rotes Gate gehört als Carveout nach
      `docs/plan/carveouts/` (heute leer — kein Handlungsbedarf, nur als
      Regel dokumentieren).
- [ ] **Zweite Tabelle „Werkzeuge — kein Gate"** für Targets, die der Agent
      braucht, aber die nichts über den Repo-Zustand urteilen (Mover,
      Messungen, Vorschau-Läufe) — Spalten `Target | Tut was | Bindung`,
      Bindung-Zelle trägt `kein Gate` **in der Zeile selbst**.
- [ ] **Ein Gate je Datei, sobald sein Vertrag mehr als einen Satz braucht**:
      Prosa nach `harness/sensors/<target>.md` (kopiert aus
      `harness/sensors/gate.template.md`), Target-Zelle wird zum Link
      **darauf**. Sensor-Dateien tragen **kein Status-/Datumsfeld** (anders
      als MR-Dateien aus `slice-013`) und **kein `sensors/done/`** — ein
      retiriertes Gate verschwindet ersatzlos (Zeile + Datei), es wandert
      nicht.
- [ ] **Wichtige Referenz-Regel für künftige einfrierende Artefakte** (Review-
      Reports, Closure-Notizen, Accepted-ADRs, geschlossene Slices — **auch
      unsere eigenen**, ab jetzt): Sie zitieren ein Gate über sein
      `make <target>`-Token, **nicht** über den Pfad zu seiner Sensor-Datei —
      die Datei kann verschwinden/wandern, das Token bleibt die stabile
      Adresse. Diese Regel selbst wird **nicht** rückwirkend auf bestehende
      Closure-Notizen (`slice-009`…`012`) angewendet — nur als Praxis ab
      diesem Slice.
- [ ] **Neue Sektion `## Leseordnung`** am Dateiende: drei bis fünf geordnete
      Zeiger, was ein neuer Mensch zuerst liest (z. B. `AGENTS.md` §Hard
      Rules → `spec/lastenheft.md` → `harness/conventions.md` bei Bedarf) —
      **keine** vollständige Liste, das wäre keine Ordnung.
- [ ] **`Guides`-Tabelle bekommt eine `.harness/skills/reviewer.md`-Zeile**
      (falls dieser Skill in m-trace existiert — prüfen, ggf. aus Template
      übernehmen, das ist ggf. ein eigener Nachzug, kein Blocker dieses
      Slices).
- [ ] **Expliziter Ein-Index-Hinweis**: „DIES IST DER EINZIGE GATE-INDEX" (oder
      sinngemäß) direkt über der Sensors-Tabelle, damit die Regel nicht nur
      in `AGENTS.md` steht.
- [ ] `make docs-check` grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/README.md` | umschreiben | Sensors-Tabelle 3-spaltig + zweite Tabelle + Leseordnung |
| `harness/conventions.md` §Zusatzklassen-Deklaration für Sensors-Bindung | ggf. kürzen | Inhalt zieht in die Tabellenzellen, Sektion wird ggf. zum reinen Zeiger oder entfällt |
| `harness/sensors/<target>.md` (0–N Dateien) | neu, nur bei Bedarf | erst anlegen, wenn ein konkretes Gate mehr als einen Satz Vertrag braucht — kein Vorab-Schnitt für alle Targets |

**Bereits geklärt (Recherche vor Schnitt, 2026-09-13):** Direkter Diff
`harness/README.template.md` v3.5.1 → v6.8.0 sowie das Regelwerk-Kapitel
`grundlagen-harness-dateien.md` §Sensors gelesen (nicht nur Fork-
Zusammenfassung). Die Sensor-Datei-Auslagerung ist **Default: Tabellenzeile**,
nicht Datei — anders als der MR-Block aus `slice-013`, wo jeder Eintrag
Pflichtfelder trägt. Für m-traces heutiges einziges Gate-Bündel
(`make docs-check`, `make gates` etc.) genügt vermutlich die Tabellenzeile;
ob ein Target (z. B. `make gates` selbst, mit seiner komplexen
Zusammensetzung aus Sub-Gates) eine eigene `sensors/`-Datei braucht, ist beim
Umsetzen zu entscheiden, nicht hier vorwegzunehmen.

**Offen für die Implementierung:**
1. Welche m-trace-Targets brauchen tatsächlich eine eigene
   `harness/sensors/<target>.md`? Kandidat: `make gates` (bündelt viele
   Sub-Checks, der Vertrag ist mehr als ein Satz) — beim Schreiben der
   Tabelle entscheiden, nicht vorab festlegen.
2. Existiert `.harness/skills/reviewer.md` in m-trace bereits (aus
   `slice-007`s Review-Harness)? Falls ja, nur die Guides-Zeile ergänzen;
   falls nein, das ist ein separater Nachzug (nicht Teil dieses Slices).

## 4. Trigger

- **`in-progress`:** nach `slice-013` (Konventions-Index, da die
  Bindungsklassen-Sektion dorthin bezieht) **und** parallel zu/nach
  `slice-014` (AGENTS.md-Rewrite entfernt die Duplikat-Tabelle — beide
  Slices sollten in kurzer Folge laufen, damit das Repo nicht zwischenzeitlich
  zwei widersprüchliche Gate-Indizes zeigt oder keinen).
- **Rückführung:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + `AGENTS.md` (aus `slice-014`)
verweist konsistent auf diese Sektion + Closure-Notiz + `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Reihenfolge-Falle mit Tranche 3:** Wird `slice-014` (Gate-Tabelle aus
  AGENTS.md entfernen) vor diesem Slice abgeschlossen, gibt es
  zwischenzeitlich **keinen** vollständigen Gate-Index im Repo. Beide sollten
  im selben Arbeitsgang oder in enger Folge laufen.
- **Sensor-Datei-Umfang unklar bis zur Umsetzung** (§3, offene Frage 1) —
  bewusst nicht vorab entschieden, um keine leeren/unnötigen Dateien zu
  produzieren (Regelwerk: „ein Gate je Datei, sobald sein Vertrag mehr als
  einen Satz braucht" — nicht: jedes Gate bekommt eine Datei).
- **Kein ADR nötig.** Reine Struktur-/Index-Nachführung, keine
  Architekturentscheidung, kein gesenktes Gate — ein Gate wird schärfer
  indiziert, nicht geschwächt.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

### Sub-Area: harness/README.md (Werkzeug/Prozess)

Reine Index-/Struktur-Nachführung des Root-Einstiegsdokuments, kein
Produktcode berührt. Ohne ADR — kein Gate wird gesenkt, nur schärfer
indiziert und mit einer Bindung-Spalte versehen, die vorher fehlte.
