# Slice 006: Planning-Layout — Nicht-Slices flach aus `in-progress/`

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle (Wartung/Layout).

**Bezug:** `templates/docs/plan/planning/README.template.md` (§„Lifecycle-Verzeichnisse
sind slice-reserviert"), MR-007 (Slice/Welle-Form), MR-005 (Risiko-Register),
`docs/reviews/README.md` (Review-Report-Klasse). Auslöser: Planning-Layout-Audit
2026-07-23.

**Autor:** Harness-Migration. **Datum:** 2026-07-23.

---

## 1. Ziel

Das slice-reservierte Lifecycle-Verzeichnis `docs/plan/planning/in-progress/`
enthält nur noch **Slices** + die kanonisch vorgesehene `roadmap.md`. Alle
Nicht-Slice-Artefakte wandern an ihre kanon-konsistente Heimat — **mit
vorhandenen Mustern, keine neue Artefaktklasse** (der Kanon *schweigt* über
Nicht-Slice-Artefakte, er verbietet sie nicht; der Welle-Plan liegt bereits
**flach** in `planning/` — dasselbe Muster trägt Register).

Kein Regelwerk-Change, keine Rückwärtskante: die einzige explizit berührte Regel
ist „Lifecycle-Verzeichnisse sind slice-reserviert" — eine reine
Platzierungs-Hygiene lokal in m-trace.

## 2. Definition of Done

- [x] `in-progress/` enthält nur Slices + `roadmap.md`.
- [x] `risks-backlog.md`, `extra-gates.md` → **flach in `planning/`**.
- [x] `r13-trivy-rereview-2026-05-13.md` → `docs/reviews/2026-05-13-r13-trivy-rereview.md` (Kanon-Review-Klasse).
- [x] `risks-backlog-werkzeug-triage.md` → **D1: flach `planning/`** neben dem Register (Owner-Go 2026-07-23).
- [x] `conventions.md` (0.1.0-Waise, 0 Referrer) → **D2: `docs/spike/`** archiviert (Owner-Go 2026-07-23).
- [x] Alle Referrer repathed (46 MD skriptet + 3 Nicht-MD-Pfade); `make docs-check` grün.
- [x] MR-005 auf **flache-Platzierung-Strukturregel** zurückgebaut (keine Abweichung, keine Klassen-Sanktion).
- [x] `.d-check.yml:35` (`ids`-Target für `R-*`) auf den flachen Pfad gezogen.
- [x] `make gates` grün. Closure-Notiz.

## 3. Plan (vor Code)

| Datei | Von | Nach | Referrer |
|---|---|---|---|
| `risks-backlog.md` | `planning/in-progress/` | `planning/` (flach) | 49 + Auswärts-Links (Tiefe −1) |
| `extra-gates.md` | `planning/in-progress/` | `planning/` (flach) | 14 + Auswärts-Links |
| `r13-trivy-rereview-2026-05-13.md` | `planning/in-progress/` | `docs/reviews/2026-05-13-r13-trivy-rereview.md` | 2 |
| `risks-backlog-werkzeug-triage.md` | `planning/in-progress/` | **[D1]** `planning/` flach | 5 |
| `conventions.md` | `planning/in-progress/` | **[D2]** `docs/spike/` | 0 |

**Repath-Verfahren:** skript-getriebene Link-Umschreibung (wie W5), danach
`make docs-check` als vollständiges Link-/Anchor-Netz. `roadmap.md` **bleibt** in
`in-progress/` (kanonisch die eine erlaubte Nicht-Slice-Datei dort).

**MR-005-Rückbau:** Scope + Text so kürzen, dass MR-005 nur noch die *Existenz*
des Registers als repo-spezifische Praxis vermerkt (falls überhaupt nötig) —
nicht mehr „Kanon kennt kein Äquivalent" als Klassen-Sanktion. Der Auflösungs-
Trigger (W6-Roadmap-Kandidat-Aufnahme) bleibt als Sachverhalt.

## 4. Trigger

- **`in-progress`:** Audit-Befund steht (slice-005-Nachgang). Erfüllt.
- **Rückführung:** rein Doku/Layout; kein Produktcode. Bei zu viel Churn pro
  Datei → in Teil-Slices (Register vs. Records) splitten.

## 5. Closure-Trigger

DoD grün + `make gates` grün + Closure-Notiz; `git mv` des Slice nach `done/`.

## 6. Risiken und offene Punkte

- **Link-Churn:** `risks-backlog.md` hat 49 Referrer + Dutzende Auswärts-Links
  (72 KB) → größter Fehler-Vektor. `docs-check` fängt jeden gebrochenen Link;
  Vorgehen identisch zum bewährten W5-Move (281 Links skriptet, grün).
- **`.d-check.yml`:** falls `matrix`/`trace`/`codepaths` explizite
  `in-progress/`-Pfade der bewegten Dateien führen, mitziehen.
- **Zwei offene Entscheidungen (D1/D2):** im Owner-Go unten.

## 7. Closure-Notiz

Die Umsetzung war rein mechanisch, der Erkenntnisweg war es nicht. Fünf
Nicht-Slice-Dateien verließen das slice-reservierte `in-progress/`: zwei stehende
Register (`risks-backlog.md`, `extra-gates.md`) + der Triage-Companion **flach nach
`planning/`** (Welle-Flach-Präzedenz), der datierte Security-Re-Review nach
`docs/reviews/` (Kanon-Review-Klasse, Modul 10 — das Verzeichnis war angelegt, aber
leer), die 0.1.0-Waise `conventions.md` nach `docs/spike/`. `in-progress/` hält jetzt
nur Slices + `roadmap.md`. Repath: 46 MD-Dateien skriptet (Algorithmus in
Absolut-Pfad-Raum, behandelt In- und Auswärts-Links inkl. Ziel-und-Referrer-beide-
bewegt), 3 funktionale Nicht-MD-Pfade manuell (`.d-check.yml:35` `ids`-Target,
`add_requirement_anchors.py`, `open-security-audit-issue.sh`); `docs-check` grün.

**Was die eigentliche Lehre war (Steering-Loop, drei Owner-Korrekturen):** (1) Ich
begründete die Ist-Platzierung mit „bewusst als Companion" — erfunden; die Provenienz
zeigte einen mechanischen W5-Move ([[feedback_decision_attribution.md]]). (2) Ich rahmte
das Kanon-Schweigen als **Lücke**, die per MR zu füllen sei — der Kurs-Autor korrigierte:
*Schweigen ≠ Ausschluss*, und der MR-Block ist für **Abweichungen**, nicht zum Stopfen
von Regelwerk-Lücken; „keine Heimat vorgesehen" ≠ „Heimat ausgeschlossen". (3) Ich übersah,
dass der **Welle-Plan bereits flach in `planning/`** liegt — das Muster für Nicht-Slice-
Artefakte lag die ganze Zeit vor, ich musste es nur verallgemeinern statt eine neue Klasse
zu erfinden. **Kern-Lehre: erst prüfen, was Kanon + Repo bereits vorsehen** (Taxonomie,
`docs/reviews/`, Welle-Flach), dann platzieren — nie Struktur erfinden, die die Lehre nicht
motiviert. MR-005 entsprechend zur reinen Strukturregel zurückgebaut.

**Sekundärbefund (offen, eigener Folge-Slice):** `docs/reviews/` war leer, obwohl m-trace
viele Reviews fuhr — das Review-Harness ist eingerichtet, aber ungenutzt (Reviews landeten
ad-hoc in Commit-Messages/Memory). Anwendungslücke, kein Regelwerk-Thema.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Planning-Layout

- **Modus:** Brownfield (Bestandslayout mit Fehlplatzierung im Lifecycle-Dir).
- **Konventionen-Dichte:** hoch — README.template „slice-reserviert" + MR-005/007.
- **Phase-Reife:** reifer Bestand; nur die Platzierung war regelwidrig.
- **Evidenz-/Diskrepanz-Risiko:** niedrig-mittel — mechanischer Move + Repath,
  `docs-check` verifiziert vollständig.
- **Reconciliation-Aufwand:** dieser Slice.
