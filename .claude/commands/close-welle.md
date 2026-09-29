# Welle schließen (Harness)

Argument: $ARGUMENTS

Dieser Command führt die **Planner**-Rolle für die **Wellen-Closure** (Modul 6). Eine Welle schließt
**nicht** durch einen einzelnen Slice-Übergang, sondern durch einen geordneten Ablauf, der alle ihre
Slices bündelt — **sechs Schritte, jeder hinterlässt einen Beleg, keiner ein Datum.** Erst wenn alle
sechs Belege vorliegen, ist die Welle *auditierbar* geschlossen.

**Die Welle-Plan-Datei wandert bei Closure per `git mv` nach `done/`** (neben ihre Results-Notiz) —
der Zustand ist die **Verzeichnis-Position, kein `Status:`-Feld** (wie beim Slice). Die Closure
erzeugt **zusätzlich** eine *separate* Results-Notiz (`done/<welle-id>-results.md`).

Kanonische Quellen (vendored Regelwerk, `.harness/baseline/<tag>/regelwerk/`): Modul 6 (Roadmap /
Wellen-Closure), Modul 7 (Carveouts), Modul 5 (Lifecycle). Bei Konflikt gilt der Kurs.

## Repo-lokale Adaptionen, die du beachten MUSST (ANPASSEN an dein Repo)

<!-- ANPASSEN: die Adaptionen DEINES Repos gegenüber der Baseline (dein
     `harness/conventions.md`, „MR-Block"). Die closure-relevanten aus der
     emittierten Schicht stehen unten; ergänze/streiche nach deinem Repo. -->

- **Docker-only + Gate-Nachweis/Stop-Hook.** Nur `make`-Targets; `make gates` endet mit
  `record-gates`. Jede Inhaltsänderung nach einem Gate-Lauf (inkl. Commit) macht den Stempel ungültig →
  nach dem Wave-Self-Close-Commit `make gates` grün bestätigen.
- **Strenges Doc-Gate.** `LH-`/`ADR-`/`MR-`-Kennungen in gescannten `.md` als klickbare Anker-Links —
  die Results-Notiz und die Roadmap werden gescannt.
- **Neue Artefakte per `cp` aus den vendored Templates** — die Results-Notiz entsteht per `cp` aus
  `.harness/baseline/<tag>/templates/docs/plan/planning/welle-results.template.md` und wird danach
  ausgefüllt; die Welle-Datei bleibt Quelle der Plan-Struktur für alles, was das Template offenlässt.
- **Commit via Message-Datei** (`git commit -F <datei>`).

## Vorbedingung: Kontext lesen

1. `CLAUDE.md` (falls vorhanden), `harness/README.md`, `AGENTS.md`, `harness/conventions.md`,
   **Modul 6** on-demand und die Welle-Plan-Datei (`docs/plan/planning/<welle-id>.md`, v. a. §3
   Closure-Kriterien) lesen.

## Die sechs Schritte (Modul 6 — jeder mit Beleg, keiner mit Datum)

2. **Schritt 1 — Trigger prüfen.** Alle Slices der Welle liegen in `done/`; `make gates` grün; die
   welle-spezifischen Closure-Kriterien aus der Welle-Datei §3 sind erfüllt (z. B. ein benannter Smoke).
   Das ist die **beobachtbare** Bedingung, nicht der Kalendertag. Fehlt ein Beleg (ein Slice nicht
   `done`, ein Gate rot), **schließt die Welle nicht** — kein halbfertiges `done/`. Erzeuge die Belege
   **real** (Gate-Ausgabe, Smoke-Lauf), behaupte sie nicht.
3. **Schritt 2 — Trigger-Audit der Welle.** Drei Artefaktklassen tragen einen Trigger, alle drei
   werden geprüft: **Carveout** (Modul 7) → aufgelöst · verlängert (mit Folge-Slice) · permanent ·
   **bootstrap-aware Gate** (Modul 13) → Stufe hochschalten, oder Carveout eröffnen, wenn die neue
   Schwelle rot ist · **Entscheidung/ADR** (Modul 4) → Re-Evaluierungs-Trigger bestätigen oder
   Folge-Entscheidung mit `supersedes`. Die Welle darf *mit* dokumentiertem Carveout schließen —
   **nie** mit einem stillen roten Gate, einer stehengebliebenen Reifestufe oder einer Entscheidung,
   deren Re-Evaluierungs-Bedingung längst eintrat. Keine fälligen Trigger → je Klasse eine belegte
   „0 offen"-Feststellung, kein Auslassen. **Ein Trigger ohne Wächter ist eine Absichtserklärung mit
   Verfallsdatum.**
4. **Schritt 3 — Closure-Notiz `done/<welle-id>-results.md` schreiben** (per `cp` aus dem vendored
   `welle-results.template.md`, dann füllen). Hält fest, *was gelernt wurde*: geliefert · was
   funktionierte · was anders lief · **Steering-Loop-Einträge** (geschärfte Regel / neuer Sensor /
   benannte Spec-Lücke) · Zeiger aufs Beobachtungs-Register · Folge-Slices · Verifikation (die Belege
   aus Schritt 1). **Ohne Lerneintrag ist die Welle nicht „fertig", nur „weg".**
   **Der Lese-Schritt des Beobachtungs-Registers gehört hierher** (`docs/plan/planning/observations/README.md`,
   Modul 6): jeder Eintrag, dessen `evidence/` **≥ 3** Dateien führt, wandert in die
   Steering-Loop-Einträge und wird zur **verkörperten Regel** mit Herkunfts-Anker
   (`seit welle-<NN>`). Der Ausgang steht danach in seiner `state.md`, das Verzeichnis bleibt liegen;
   *gestrichen* trägt dieselbe `state.md`, mit Begründung — still löschen macht die Beobachtung
   ununterscheidbar von einer, die es nie gab.
   Erreicht kein Eintrag 3×, ist *„kein Eintrag über der Schwelle"* die Feststellung, die in die
   Results-Notiz gehört — Auslassen ist keine Antwort. Was **unter** 3× steht, liest diese Closure
   **nicht**; dafür ist der Sichtungs-Schritt der Slice-Planung zuständig (`/plan-welle`).
   **Zugleich gehört die Welle-Plan-Datei per `git mv` nach `done/`** — wegen der repo-lokalen Hard
   Rule 3.3 (Move ≠ Inhalt) als **eigener reiner Move-Commit** (s. Schritt 5). Der Move bricht die
   Inbound-Links (Roadmap + die Welle-Verweise der Slices) **und** die eigenen `../`-Links der Datei
   (jetzt eine Ebene tiefer) → im selben Zug reconcilen, bis `docs-check` grün ist.
   **Zum Schluss die drei Paarungen prüfen** — erst jetzt, weil sie die gerade entstandenen Einträge
   prüfen: (a) *Anker* — wo ein Steering-Loop-Eintrag das Feld `liegt in <Zielort>` trägt, existiert
   der Zielort und trägt `seit welle-<NN>` bzw. `seit slice-<NNN>`; (b) *Folge-Slice* — jeder genannte
   Folge-Slice existiert als Datei irgendwo im Planning-Lifecycle, nicht nur in `open/`;
   (c) *Register* — jede genannte Kennung `BEO-<KUERZEL>/<slug>` existiert als Verzeichnis im
   Register, und jedes Verzeichnis trägt ein nicht leeres `evidence/`. Rot heißt in allen drei
   Fällen: etwas wurde versprochen und nicht angelegt.
5. **Schritt 4 — Zeitdokumente der Welle archivieren.** Ihre Slice-Dateien, ihr Plan und die
   Review-Reports dieser Slices wandern nach `done/<welle-id>/archiv.zip`; an ihrer Stelle bleiben
   gekürzte Stubs — per `cp` aus den vendored Vorlagen `archiv-stub-slice.template.md` bzw.
   `archiv-stub-welle.template.md`; ein Stub trägt keine Abschnittsüberschriften, Review-Reports
   bekommen keinen. Die **Ergebnisnotiz bleibt vollständig und flach**. Eingesammelt wird nach der
   Welle, nicht nach dem Verzeichnis: die Slices, deren `Welle:` diese Welle nennt, **und** die
   wellenlosen seit der letzten Closure; Slices einer noch offenen Welle bleiben liegen.
   **Die Operation gehört in ein Werkzeug, nicht in Handarbeit** — dass das Archiv vollständig ist,
   bezeugt allein der Archivierungs-Commit, und der Move bricht die Verweise auf die bewegten
   Dateien. Liegt der Träger im Repo, führt `make archive-welle WELLE=<welle-id>` ihn; fehlt er,
   sagt das Kommando das selbst.
   <!-- ANPASSEN: der Weg zum Träger ist die repo-spezifische Stelle; nenne hier den deines Repos.
        Der Ziel-NAME `archive-welle` ist es nicht. Er kommt aus einem tool-eigenen Fragment, das
        jeder Bootstrap kanonisch neu schreibt — ein umbenanntes Ziel hält darum nicht, und diese
        Anleitung bleibt auf einem Namen stehen, den `make` nach dem nächsten Lauf nicht mehr kennt.
        Der Aufruf endet dann laut statt still; er endet nicht erfolgreich. -->
   Hat dein Repo das Werkzeug nicht, ist die Bedingung nicht eingetreten; **das** gehört als
   Feststellung in die Results-Notiz, nicht in einen Handlauf. Wellen, die vor der Einführung
   schlossen, müssen nicht nachgerüstet werden.
   **Vor der ersten Archivierung den Geltungsbereich deiner Sensoren prüfen:** was auf `done/*.md`
   keilt, sieht die Stubs eine Ebene tiefer nicht mehr und bleibt grün, ohne noch etwas zu prüfen.
   Prüfe ebenso, ob die Link-/ID-Pflichten deines Doku-Gates im Stub gelten — er trägt Kennungen.
6. **Schritt 5 — Wave-Self-Close-Commit + Move.** Der **Self-Close-Commit** (Inhalt) trägt: die
   Results-Notiz + die Welle-Datei §7 (Verweis auf die Results-Notiz; **kein `Status:`-Feld** — der
   Zustand ist die Position) + die Roadmap-Fortschreibung (Schritt 6). **Danach** der reine
   **`git mv`-Commit** der Welle-Plan-Datei nach `done/` und der **Link-Reconciliation-Commit** (Schritt 3):
   Hard Rule 3.3 trennt Move und Inhalt, daher mehrere Commits statt des einen Baseline-Self-Close-Commits
   — die Bewegung bleibt beobachtbar (ein zusammenhängender Zug). Commits via `-F`.
7. **Schritt 6 — Roadmap fortschreiben** (`in-progress/roadmap.md`, im selben Commit): die Welle
   bekommt ihre Zeile in *Abgeschlossene Wellen* (mit Zeiger auf die Results-Notiz), ihr Zeiger
   verlässt *Offene Wellen*. **Befördert wird niemand** — welche Wellen offen sind, sagen die flachen
   Welle-Dateien; die Zeilen unter *Nächste Wellen* bleiben stehen, bis ihre Welle eröffnet wird. Den
   zugehörigen Meilenstein auf *erreicht* setzen, falls die Welle ihn erfüllt; löste ein Trigger eine
   Umplanung aus, bekommt *Historische Trigger-Verschiebungen* ihren Eintrag — eine Schließung ist
   keine Umplanung und gehört nicht dorthin.

## Abschluss

8. `make gates` grün nach dem Commit bestätigen (der Stop-Hook-Stempel muss auf den aktuellen Tree
   passen). Erst wenn **alle sechs Belege** vorliegen — Trigger · Carveout-Audit · Results-Notiz ·
   Archivierung (oder ihre nicht eingetretene Start-Bedingung) · Self-Close-Commit ·
   fortgeschriebene Roadmap — ist die Welle auditierbar geschlossen.

**Merke (Modul 6):** Datum ist *Output*, nie *Trigger*. Wer die Welle am Kalendertag schließt, kappt
halbfertige Slices und produziert genau die Auditierbarkeits-Lücke, die der Harness verhindert.

Gates nicht überspringen. Keine Erfolgsmeldung ohne Command-Ausgabe.
