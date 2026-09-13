# Slice 016: Roadmap auf das v6.8.0-Template-Format heben (Tranche 8)

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](../welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 8 von 10).

**Bezug:** `docs/plan/planning/in-progress/roadmap.md`, `.d-check.yml`
(`planning`-Block), `.harness/baseline/v6.8.0/templates/docs/plan/planning/roadmap.template.md`.

**Autor:** Owner-Entscheidung 2026-09-13 (welle-02 §8 Punkt 2). **Datum:** 2026-09-13.

---

## 1. Ziel

Die in `welle-02` §8 offen gehaltene Owner-Entscheidung ist gefallen: m-trace
übernimmt das **neue** Roadmap-Template-Format vollständig, nicht nur den
Mechanismus (den wir beim Anlegen von `welle-02` bereits über
`planning.waves.mode: many` umgesetzt hatten). Drei konkrete Format-Deltas
aus dem direkten `roadmap.template.md`-Diff (v3.5.1 → v6.8.0):

1. Abschnitt „## Aktuelle Welle" → „## Offene Wellen".
2. Die Kopfzeile „**Status:** Aktiv. **Letzte Änderung:** …" entfällt
   ersatzlos (im neuen Template nicht mehr vorgesehen).
3. Meilensteine-Status trägt „erreicht YYYY-MM-DD" **plus auflösbarem
   Beleg-Anker** statt bloßem „erreicht".

## 2. Definition of Done

- [ ] `roadmap.md`: Kopfzeile `**Status:** Aktiv. **Letzte Änderung:** …`
      entfernt.
- [ ] `roadmap.md`: `## Aktuelle Welle` → `## Offene Wellen` umbenannt,
      Absatztext unverändert (er beschreibt bereits korrekt den
      Mehr-Wellen-Zustand: Marker + `welle-02`-Zeiger zusammen).
- [ ] `roadmap.md`: Meilensteine-Tabelle — jede „erreicht"-Zeile bekommt ihr
      Datum **plus** einen auflösbaren Beleg-Anker (Git-Tag-Referenz über
      lokal auflösbaren Link, wo verfügbar; sonst der nächstliegende
      Plan-/Migrations-Record).
- [ ] `.d-check.yml`: `planning`-Block bekommt `heading: "## Offene Wellen"`
      (expliziter Override, da d-checks Default `## Aktuelle Welle` lautet).
- [ ] `make docs-check` grün (Positiv-Probe: `wave-drift`/`planning-drift`
      bleiben bei 0, trotz Heading-Wechsel).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `docs/plan/planning/in-progress/roadmap.md` | update | drei Format-Deltas oben |
| `.d-check.yml` | update | `heading:`-Override für den neuen Abschnittsnamen |

**Beleg-Anker je Meilenstein-Zeile** (recherchiert vor Umsetzung):
- `0.25.0 released` → `CHANGELOG.md#0250---2026-07-13` (lokal auflösbar,
  Tag `v0.25.0` existiert zusätzlich als Git-Referenz, aber CHANGELOG ist
  das netzlos auflösbare Artefakt).
- `v3.5.0-Harness-Migration abgeschlossen` → bereits verlinkt
  (`../done/plan-harness-v3.5.0-migration.md`) — zählt schon als Anker,
  nur das Status-Wort selbst bekommt das Datum ergänzt.
- `0.25.1 released` → `CHANGELOG.md#0251---2026-08-21` (kein dedizierter
  Plan-Record, war „ohne Welle"/Wartung).

## 4. Trigger

- **`in-progress`:** sofort — Owner-Entscheidung ist gefallen, keine weitere
  Abhängigkeit.
- **Rückführung:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **Kein ADR nötig.** Reine Format-/Terminologie-Nachführung des
  Roadmap-Dokuments gegen die neue Baseline-Form, keine
  Architekturentscheidung, kein gesenktes Gate.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Roadmap-Format (Werkzeug/Prozess)

Reine Format-Nachführung eines Planungsdokuments, kein Produktcode berührt.
Ohne ADR — kosmetische/terminologische Angleichung, keine funktionale
Änderung (der Mehr-Wellen-Mechanismus stand bereits vor diesem Slice).
