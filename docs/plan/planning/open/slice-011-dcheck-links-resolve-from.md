# Slice 011: d-check `links.resolve-from` für ortsfeste Verweise im Planning-Lifecycle aktivieren

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle (Werkzeug/Harness).

**Bezug:** `.d-check.yml`, `docs/plan/planning/{open,next,in-progress,done}/`,
d-check-Modul `links` (opt-in-Fähigkeit `resolve-from`, seit v0.60.0, CR des
Konsumenten a-check). Ausgangspunkt: der d-check-Versions-Bump
v0.51.1 → v0.75.0.

**Autor:** Nachgang d-check-Versions-Bump. **Datum:** 2026-09-13.

---

## 1. Ziel

Planning-Artefakte wandern per `git mv` zwischen
`open/ → next/ → in-progress/ → done/` (Planning-Lifecycle, MR-007). Ein
relativer Verweis in einem Slice- oder Wellen-Dokument (z. B. auf eine ADR,
eine Spec-Datei oder ein Geschwister-Dokument) ist am **aktuellen** Ort grün,
kann aber beim nächsten `git mv` ins Leere zeigen, wenn die relative
Pfadtiefe sich zwischen den vier Verzeichnissen unterscheidet. `links`
(bereits aktiv in `.d-check.yml`) prüft heute nur, ob ein Ziel am **Ist-Ort**
existiert — nicht, ob es das an **jedem** Ort der Wander-Gruppe täte. Die
opt-in-Fähigkeit `resolve-from` schließt genau diese Lücke: sie meldet
`link-position-dependent`, **bevor** der nächste Move den Verweis bricht.

## 2. Definition of Done

- [ ] `.d-check.yml` führt einen `links.resolve-from`-Block mit `dirs: [
      docs/plan/planning/open, docs/plan/planning/next,
      docs/plan/planning/in-progress ]` und `fixed-dirs: [
      docs/plan/planning/done ]`.
- [ ] `make docs-check` läuft grün gegen den aktuellen Bestand (Positiv-Probe,
      bereits vorab verifiziert — siehe §3).
- [ ] Negativ-Probe dokumentiert: ein Testdokument mit einem nur ortsabhängig
      auflösenden relativen Verweis (unterschiedliche Pfadtiefe je
      Lifecycle-Verzeichnis) erzeugt nachweislich `link-position-dependent`.
- [ ] `make gates` bleibt grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.d-check.yml` | update | `links.resolve-from`-Block ergänzen (additiv zum bestehenden `links`-Modul, das schon aktiv ist) |

**Bereits geklärt (Recherche vor Schnitt, 2026-09-13):** Die Konfiguration ist
exakt die im Benutzerhandbuch als Lifecycle-Anlassfall genannte Form. Vier
Feinheiten sind vorab gemessen:

1. **Quellen sind nur Dateien in den `dirs`-Verzeichnissen selbst** (kein
   rekursiver Treffer auf Unterverzeichnisse) — passt zu m-traces flacher
   Ablage je Lifecycle-Stufe.
2. **`fixed-dirs` (hier `done/`) zählt als gültiger Auflösungsort, wird aber
   selbst nicht als Quelle geprüft** — ein bereits abgeschlossenes Dokument
   in `done/` löst also keinen eigenen Befund aus, dient aber als Ziel, gegen
   das Verweise aus `open/`/`next/`/`in-progress/` aufgelöst werden müssen.
3. **Fail-closed nur, wenn ALLE `dirs`-Orte fehlen** — ein einzelnes leeres
   Lifecycle-Verzeichnis (z. B. `open/` heute, siehe Bestand) meldet
   **nichts**, weil git leere Verzeichnisse nicht überträgt und ein frischer
   Klon das legitim so vorfindet.
4. **Bewusst ortsgebundene Verweise** (falls je nötig) laufen über das
   bestehende, bereits in `.d-check.yml` genutzte `ignore-refs`-Ventil — kein
   neuer Mechanismus.

**Verifikation (Probe-Lauf gegen den echten Baum, `d-check:v0.75.0`, voller
Modul-Satz inkl. `links.resolve-from` wie oben konfiguriert):** `126
Datei(en) geprüft, 0 Befund(e)` — keine bestehende Referenz in
`in-progress/roadmap.md` (aktuell die einzige Datei mit Inhalt in den drei
`dirs`-Verzeichnissen; `open/` und `next/` sind bis auf `next/README.md`
leer) ist heute schon ortsabhängig.

**Offen für die Implementierung:** Da der heutige Bestand in
`open/`/`next/`/`in-progress/` fast leer ist (dieser Slice selbst sowie
`slice-010` sind die ersten `open/`-Bewohner), ist der Positiv-Lauf wenig
aussagekräftig für sich allein — die Negativ-Probe (Punkt 2 der DoD) trägt
hier das eigentliche Gewicht des Nachweises, analog zur Lehre aus
`slice-009`: ein Gate ist erst dann etwas wert, wenn es nachweislich auch rot
werden kann.

## 4. Trigger

- **`in-progress`:** jederzeit — reine Werkzeug-/Config-Änderung, kein
  Produktcode betroffen, kein Abhängigkeits-Trigger. Sinnvollerweise **nach**
  `slice-010`, weil beide denselben Verzeichnisbereich (`docs/plan/planning/`)
  und dieselbe `.d-check.yml` anfassen — nicht zwingend, aber
  konflikt-ärmer nacheinander als parallel.
- **Rückführung:** keine erwartet — reine additive Config, geringes Risiko.

## 5. Closure-Trigger

DoD vollständig + `make gates` grün + Positiv- und Negativ-Probe belegt +
Closure-Notiz + `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Geringe Schärfe am heutigen Bestand.** Mit `open/`/`next/` fast leer
  entfaltet das Gate seinen Wert erst, sobald echte Slices durch den
  Lifecycle wandern — dieser Slice selbst (und `slice-010`) sind die ersten
  Nutznießer/Testfälle in `open/`.
- **Kein ADR nötig.** Reine Gate-Ergänzung, keine gesenkte Prüfung, keine
  Architekturentscheidung — passt additiv in den bestehenden `links`-Block.
- **Wechselwirkung mit `ignore-refs`:** falls ein künftiger Slice bewusst
  einen ortsgebundenen Verweis braucht (etwa ein Beispiel-Pfad, der nur am
  Ist-Ort Sinn ergibt), ist das bestehende `ignore-refs`-Ventil der richtige
  Ort dafür — nicht ein neuer, modul-lokaler Schlüssel.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

### Sub-Area: d-check-Konfiguration (Werkzeug)

Reine Config-/Harness-Änderung, kein Produktcode, keine Spec, kein
Requirement berührt. Ohne Welle (Modul 5 „Wartung/Architektur"), ohne ADR —
es wird kein Gate gesenkt, sondern eines zusätzlich scharf geschaltet.
