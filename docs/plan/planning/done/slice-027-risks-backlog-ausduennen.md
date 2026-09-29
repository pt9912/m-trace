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
echt offenen Risiken ([R-9](../risks-backlog.md#r-9), [R-12](../risks-backlog.md#r-12), [R-30](../risks-backlog.md#r-30) sowie [R-13](../risks-backlog.md#r-13) als Zeiger auf die
MR-006-Registry), gelöste Zeilen (🟢) wandern nach §1.2 in Kurzform, der
Chronik-lastige Header-Stand-Block (0.12.x–0.19.0-Releases) schrumpft auf
Zustand (Hard Rule 3.7). Die operationale Security-Cohort-Lage lebt bereits
in der MR-006-Registry (`vulnignore.yaml`) — der [R-13](../risks-backlog.md#r-13)-Zeiger beendet die
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

- [x] §1.1 enthält nur offene Risiken ([R-9](../risks-backlog.md#r-9), [R-12](../risks-backlog.md#r-12), [R-30](../risks-backlog.md#r-30)) + [R-13](../risks-backlog.md#r-13) als
      Kompakt-Zeiger auf die MR-006-Registry; 🟢-Zeilen in §1.2 als Kurzform
      (Kennung | Kurzform | Auflösung | Verweis), Anker (`<a id="r-N">`)
      vollständig erhalten (31/31 gegen HEAD gegengeprüft).
- [x] Header-Stand-Block: Zustand statt Release-Chronik (Hard Rule 3.7).
- [x] `make docs-check` grün (ids/REQLINK: alle R-Verweise lösen weiter).
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `docs/plan/planning/risks-backlog.md` | refactor | Header schrumpfen; §1.1/§1.2 umsorten; [R-13](../risks-backlog.md#r-13) auf MR-006-Registry verdünnen |

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

Das Register schrumpfte von 287 auf 100 Zeilen: §1.1 trägt nur noch die
echt offenen Risiken (R-9 K8s-Smoke, R-12 WebRTC-Drift-Detector, R-30
SSE-Backfill-Skip) plus R-13 als Kompakt-Zeiger auf die MR-006-Registry —
die operationale Security-Cohort-Lage (reason/expires/scope je CVE) liegt
dort exklusiv, die Doppelbuchung im Backlog ist beendet. 14 🟢-Zeilen
wanderten nach §1.2 in Kurzform (Anker 31/31 erhalten, Verbatim-Check je
Zeile gegen HEAD), der 0.12.x–0.19.0-Release-Chronik-Header auf
Zustandsform.

**Was hat funktioniert:** Der Verbatim-Komparator (21 erhaltene Zeilen
zeilenweise gegen HEAD, Anker-Menge gegen alt) fing drei stille
Substanzverluste der Neuübernahme, bevor sie committet wurden: einen
verhunzten Commit-Hash (R-12), einen Tippfehler (R-31 „unauflösbar") und
drei abgerissene Verweis-Zellen (OS-4, OS-5, R-18).

**Was ging anders als geplant:** Der `ids`-Sensor meldet die R-Kennungen
des Slice-Plans selbst (`id-unlinked`) — die Linkpflicht gilt ab
`anlegen` in `in-progress/`, nicht erst in `done/` (das exempt ist). Die
Vorgänger-Slices entgingen dem nur, weil der Gate erst nach dem `git mv`
lief.

**Steering-Loop-Eintrag:** Guide geschärft: Slice-Pläne, die R-/RAK-
Kennungen nennen, verlinken sie ab Anlegen als `../risks-backlog.md#r-N`
bzw. `../../../spec/lastenheft.md#<anker>` — `make docs-check` gehört in
den anlegen-Commit, nicht erst in die Closure. (Gezählt, nicht verkörpert
— Erstvorkommen.)

**Beobachtungs-Register (`../observations/`):** keine Beobachtung angefallen.

**Folge-Slices:** keine. Nächste Suppression-Schwellen: 2026-10-08
(CVE-2026-53615), 2026-10-29 (acl/attr/gzip), 2026-11-02 (perl-Kohorte +
R-13-Strukturauflösung).

**Risiken aus §6:** beide entfallen — Anker-Bruch durch docs-check
ausgeschlossen, Informationsverlust durch git-Historie + verlinkte
Plans/Reports abgedeckt.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Reine Register-Pflege (Brownfield): die Datei prä-adoptioniert, die
Wartungsregeln in §2 der Datei selbst tragen die Kurzform-Umsiedlung. Kein
ADR — kein Gate gesenkt, keine Kennung gelöscht.
