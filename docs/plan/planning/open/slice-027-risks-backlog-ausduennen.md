# Slice 027: risks-backlog ausdünnen — §1.1 auf offene Risiken, Duplikate raus

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle — Closure-Bedingung
ist die DoD dieses Slices, keine Wellen-Grenze.

**Bezug:** `docs/plan/planning/risks-backlog.md`, `harness/conventions.md`
(MR-003 R-Familie, REQLINK), `.security/vulnignore.yaml` (MR-006-Registry,
nur Referenzziel).

**Autor:** Owner-Frage „brauchen wir das nach Regelwerk" + Antwort. **Datum:** 2026-09-29.

---

## 1. Ziel und Abgrenzung

**Ziel:** Das Risiko-Register auf seinen Kern reduzieren: §1.1 hält nur die
echt offenen Risiken (R-9, R-12, R-30 sowie R-13 als Zeiger auf die
MR-006-Registry), gelöste Zeilen (🟢) wandern nach §1.2 in Kurzform, der
Chronik-lastige Header-Stand-Block (0.12.x–0.19.0-Releases) schrumpft auf
Zustand (Hard Rule 3.7). Die operationale Security-Cohort-Lage lebt bereits
in der MR-006-Registry (`vulnignore.yaml`) — der R-13-Zeiger beendet die
Doppelbuchung; **keine** CO-Dateien (MR-006 hat die CO-Kaskade explizit
verworfen).

**Ausdrücklich NICHT in diesem Slice:**

- **Neue R-Einträge oder Trigger-Änderungen** — Bestand bleibt inhaltlich
  (Kennungen, Trigger, Auflösungen), nur Ort und Kompaktheit ändern sich.
- **CO-Dateien für den Security-Cluster** — MR-006 entschieden; das
  generische `docs/plan/carveouts/` bleibt reserviert.
- **Historische Referenzen reparieren** (`done/`-Pläne, Wartungs-Pfad-Drift
  `docs/planning/` vs. `docs/plan/planning/` in §2) — bleiben, wie in
  `done/` üblich.
- **Änderungen an `.security/vulnignore.yaml`** — reine Referenzdatei hier.

## 2. Definition of Done

- [ ] §1.1 enthält nur offene Risiken (R-9, R-12, R-30) + R-13 als
      Kompakt-Zeiger auf die MR-006-Registry; 🟢-Zeilen in §1.2 als Kurzform
      (Kennung | Kurzform | Auflösung | Verweis), Anker (`<a id="r-N">`)
      vollständig erhalten.
- [ ] Header-Stand-Block: Zustand statt Release-Chronik (Hard Rule 3.7).
- [ ] `make docs-check` grün (ids/REQLINK: alle R-Verweise lösen weiter).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `docs/plan/planning/risks-backlog.md` | refactor | Header schrumpfen; §1.1/§1.2 umsorten; R-13 auf MR-006-Registry verdünnen |

## 4. Trigger

- **Start (`next` → `in-progress`):** sofort — Owner-Auftrag.
- **Rückführungen:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **Anker-Bruch bei `spec/**`-R-Verweisen** — **Ausgang:** entfallen
  (Anker `<a id="r-N">` bleiben an den Zeilen hängen; docs-check prüft).
- **Informationsverlust bei Kurzform** — **Ausgang:** entfallen — die
  ausführlichen Mitigations-Beschreibungen bleiben in git-Historie und in
  den verlinkten Plans/Reports; die Wartungsregel verlangt genau die
  Kurzform.

## 7. Closure-Notiz

*(füllt bei Closure)*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Reine Register-Pflege (Brownfield): die Datei prä-adoptioniert, die
Wartungsregeln in §2 der Datei selbst tragen die Kurzform-Umsiedlung. Kein
ADR — kein Gate gesenkt, keine Kennung gelöscht.
