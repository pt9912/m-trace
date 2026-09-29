# Slice implementieren (Harness)

Argument: $ARGUMENTS

Dieser Command führt die **Implementer**-Rolle (Modul 9) für *einen* Slice — innerhalb der
Rollen-Sequenz Planner → Architect → Implementer → Reviewer → Verifier → Validator →
Planner-Closure (Modul 8). **Rollen-Trennung ist Kontext-Trennung:** die nachgelagerten Rollen
(Review, Verifikation, Validation, Closure) laufen in **frischem Kontext** (Subagent / geleerter
Kontext), nie im Kontext, der den Code schrieb — sonst wiederholt sich derselbe blinde Fleck.
Keine Rolle springt rückwärts ohne Übergabe-Artefakt (Findings · Folge-ADR · Carveout, Modul 8).

Kanonische Quellen (vendored Regelwerk, `.harness/baseline/<tag>/regelwerk/`): Modul 9
(Implementierung), Modul 5 (Lifecycle), Modul 8 (Rollen), Modul 10 (Review), Modul 11
(Verifikation).

## Repo-lokale Adaptionen, die du beachten MUSST (ANPASSEN an dein Repo)

<!-- ANPASSEN: Dieser Block listet die Adaptionen DEINES Repos gegenüber der Baseline
     (dein `harness/conventions.md`, „MR-Block"). Der Bootstrap hat eine
     Durchsetzungsschicht emittiert (Stop-Hook, Command-Guard, Gate-Nachweis,
     Doc-Gate); die daraus folgenden, workflow-relevanten Adaptionen stehen unten.
     Ergänze/streiche nach deinem Repo. -->

Über das Regelwerk hinaus trägt dein Repo lokale Adaptionen gegenüber der Baseline. Lies den
Adaptions-Block („MR-Block") in `harness/conventions.md`; die workflow-relevanten (aus der
emittierten Durchsetzungsschicht):

- **Docker-only, kein Host-Toolchain.** Jeder Gate und jedes Tool läuft in einem gepinnten
  Docker-Image; der emittierte PreToolUse-Guard (`.claude/hooks/pretooluse-command-guard.sh`) blockt
  die Host-Toolchain deiner Sprache (und prüft Sub-Shell-Strings). Rufe nie einen Host-Toolchain auf
  — nur die `make`-Targets.
- **Gate-Nachweis + Stop-Hook.** `make gates` endet mit `record-gates`, das einen Content-Hash des
  Working Tree stempelt; der Stop-Hook verweigert den Abschluss, solange der aktuelle Tree nicht
  passt. **Jede Inhaltsänderung nach einem Gate-Lauf — inklusive jedes Commits und jedes `git mv`
  — macht den Stempel ungültig: `make gates` erneut laufen.** Ein Commit/Move ohne frischen
  Gate-Lauf lässt den Stop-Hook rot.
- **Strenges Doc-Gate (d-check).** Jede `LH-`/`ADR-`/`MR-`-Kennung in einer gescannten `.md` muss
  ein klickbarer Anker-Link sein (link-policy: always) — ein bares Kennungs-Token bricht
  `docs-check` (`id-unlinked`). `codepaths` verlangt, dass Pfade in Inline-Code existieren: eine
  *geplante* Datei braucht einen Inline-`d-check:ignore`-Marker, eine *bewusst entfernte* gehört in
  `ignore-refs`. Spec verweist nie abwärts auf ADR/Slice; ein Verweis auf eine superseded ADR nur
  via Inline-Code + `d-check:ignore`. `docs/reviews/**` ist ausgenommen (Zeitdokumente).
- **Neue Artefakte per `cp` aus den vendored Templates** (`.harness/baseline/<tag>/templates/…`),
  dann ausfüllen — keine handgeschriebenen oder repo-gepflegten Template-Kopien.
- **Commit via Message-Datei** (`git commit -F <datei>`): der Guard scannt den Command-String,
  also nie eine Commit-Message inline, die ein geblocktes Tool-Token enthält.
- **Commit-Kennung.** Eine Commit-Message ohne Kennung (`ADR-NNNN`, `LH-XX-NN`, `MR-NNN`,
  `slice-N`) weist der git-eigene Hook `.githooks/commit-msg` ab, sobald er aktiviert ist. Er liegt
  versioniert im Repo und **reist mit dem Klon, seine Aktivierung nicht**: `make hooks-install`
  setzt `core.hooksPath` und ist der eine Schritt dazwischen; `git commit --no-verify` umgeht ihn.
  **An diesem Pfad ist das Werkzeug ein Gast** — der Name ist von `git` fixiert und das Verzeichnis
  gehört dem Repo: der Bootstrap legt seinen Träger **nur ab, wo der Pfad frei ist**. Führt dieses
  Repo dort schon einen eigenen, bleibt er unberührt, und der Bootstrap sagt es; die mitgelieferte
  Prüfung `tools/harness/commit-msg-traceability.sh` liegt in beiden Fällen daneben (sie wird bei
  jedem Lauf kanonisch neu geschrieben) und kann aus dem eigenen Träger aufgerufen werden.
  Geprüft wird die **Anwesenheit** einer Kennung, nicht ihre Wahrheit. Was er **nicht** erreicht:
  die zweite Hälfte der Traceability-Zusage — ein Doku-Update bei berührtem öffentlichem Vertrag —
  ist von einem Commit-Wächter nicht mechanisch prüfbar und bleibt deine Arbeit. Er **erreicht**
  dagegen jede Commit-Klasse, die `git` erzeugt — auch die der Repo-Werkzeuge (`make slice-mv`,
  `make archive-welle`): trägt eine solche Message keine Kennung aus der Menge — bei einer benannten
  Kennung trifft kein Muster —, fällt sie.

## Kontext lesen (Modul 9, Schritte 1–3)

1. `CLAUDE.md` lesen (falls dein Repo eines führt — das agentseitige Briefing).
2. `harness/README.md` lesen.
3. `AGENTS.md` lesen.
4. `harness/conventions.md` lesen.
5. Den Regelwerk-Index (`.harness/baseline/<tag>/regelwerk/README.md`) und das aufgabenrelevante
   Modul **on-demand** lesen (Source Precedence, committet vendored Baseline). Nicht den ganzen
   Baum laden.
6. Die als Argument übergebene Slice-Datei lesen.
7. Alle referenzierten ADRs und Anforderungen lesen.
8. Berichten: Slice-ID · LH-IDs · ADR-IDs · betroffene Komponenten · zu laufende Gates.

## Nach in-progress eintreten (Modul 5 Lifecycle + Modul 8 Übergabe)

9. Der Implementer erhält den Slice **in `in-progress/`** (Planner→Implementer-Übergabe,
   Modul 8; `next → in-progress` = „Implementer beginnt", Modul 5). Liegt er noch in `open/`,
   zuerst dorthin verschieben (`open → next → in-progress`); `open → next` setzt dabei das
   Kopf-Feld `Verantwortlich:`. **Jeder dieser Übergänge läuft über
   `make slice-mv SLICE=slice-<Kennung> TO=<next|in-progress>`** — das Werkzeug bewegt die Datei
   per `git mv` und committet den **reinen Move** sofort als eigenen Commit (Hard Rule 3.3), und
   es zieht danach die **Verweise** nach: **eingehend** jede Präfix-Form auf die bewegte Datei,
   repo-weit, dazu präfixlose Links aus den Geschwistern im Ausgangsverzeichnis, und **ausgehend**
   die präfixlosen Ziele innerhalb der bewegten Datei selbst. Fielen
   Verweise an, committet es sie als **zweiten**, vom Move getrennten Commit; sonst bleibt es beim
   einen Move-Commit. Der Move landet damit **auf dem Hauptzweig, vor der Arbeit** — der Branch
   entsteht danach. Reist er erst im PR mit, ist der Zustand zweigelokal, und `in-progress/`
   bleibt für alle anderen leer, bis die Arbeit fertig ist.
   <!-- ANPASSEN: der Weg zum Werkzeug ist die repo-spezifische Stelle; nenne hier den deines
        Repos. Der Ziel-NAME `slice-mv` ist es nicht — er kommt aus einem tool-eigenen Fragment,
        das jeder Bootstrap kanonisch neu schreibt; ein umbenanntes Ziel hält darum nicht, und
        diese Anleitung bliebe auf einem Namen stehen, den `make` nach dem nächsten Lauf nicht
        mehr kennt. Das Fragment bricht dann laut ab, statt still nichts zu tun. -->
   **Was das Werkzeug nicht kann, steht in seinem Kopf, und der Rest ist deine Handarbeit:** es
   zieht Pfade nach, keine Zustandsätze, und einen präfixlosen **eingehenden** Verweis erkennt es nur
   als Markdown-Link in den Geschwistern, die flach im Ausgangsverzeichnis liegen. `make docs-check`
   nach dem Move zeigt, was
   stehen blieb — zieh es nach, bevor der nächste Schritt startet.
10. WIP-Limit = 1 pro Implementer (Modul 5): kein paralleles `in-progress/`.
11. Lifecycle-Rücksprungkanten (Modul 5), falls sich der Slice als falsch erweist: zu groß →
    `in-progress → next` (zurück zur Zerlegung); blockiert → `in-progress → open` (Carveout,
    Modul 7). Zurückführen ist Disziplin, kein Scheitern.

## Plan vor Code (Modul 9, Schritt 4 — nicht optional)

12. **Den Ist-Zustand gegen den Slice-Plan messen, bevor du editierst** (`grep`/`diff`, nicht
    `edit`) — Geschwister-Slices lassen Pläne altern (gelöschte Pfade, verschobene
    Lifecycle-Dateien). Drift zuerst abgleichen; keinen veralteten Plan blind abarbeiten.
13. Die kleinste sinnvolle Änderung gegen die DoD planen. Erst planen, dann coden.

## Implementieren und gaten (Modul 9, Schritte 5–6)

14. Die kleinste sinnvolle Änderung implementieren.
15. Zuerst den engsten nützlichen Gate laufen lassen (z. B. eine Testdatei / ein Gate).
16. `make gates` laufen lassen.

**Plan-Defekt-Rücksprungkanten (Modul 9):** ein roter Sensor (15) oder rotes Gate (16) führt
zurück zum **Plan** (13) — den Plan verfeinern, nicht den Kontext neu lesen. Ein Rücksprung zu
Schritt 1 signalisiert einen Kontext-Defekt. Ein struktureller Fehlschnitt (zu groß / blockiert)
ist eine Lifecycle-Rücksprungkante (11).

## Pre-completion-Checkliste (Modul 9, Schritt 8 — letzte Handlung der Implementer-Rolle)

17. Doku, ADR-Index und README aktualisieren, falls ein öffentlicher Vertrag berührt ist.
18. Die Pre-completion-Checkliste laufen: die DoD Punkt für Punkt **behaupten** und die
    **Sensor-Belege** anhängen — `make gates` **und die Nicht-Gate-Sensoren, die den Slice
    betreffen** (die dein Repo führt — z. B. ein Mutations-Sensor, wenn Wächter neu/geändert sind;
    ein Emit-/Integrations-Smoke, wenn der betroffene Pfad berührt ist). Modul 11 verlangt genau
    hier den Lauf: *„der Implementer-Agent läuft `make verify-*` **selbst** vor der
    ‚fertig'-Meldung"* — ein Sensor, der erst zur Wellen-Closure feuert, ist pro Slice keiner.
    **Ein nicht gelaufener Sensor ist ein Befund, kein Formfehler:** ihn wegzulassen ist eine
    Aussage („betrifft diesen Slice nicht"), die begründet werden muss. Kein Gate erzwingt das —
    der Stop-Hook deckt nur `make gates`. Das ist die *Behauptung* der Implementer-Rolle und die
    *Eingabe* des Verifiers — **nicht** das finale DoD-Urteil (Modul 11: „Behauptung ohne
    Bestätigung ist die häufigste Verifier-Lücke"; eine DoD-Verletzung ist eine Verifier-only-Klasse,
    unsichtbar für Review und Tests). Ausgeführte Sensors + Restrisiken berichten.
19. **Zu jedem neuen oder geänderten Wächter die rot färbende Mutation benennen**
    (`AGENTS.md` §3.6). Ein grüner Gate-Lauf belegt nur, dass nichts *bricht* — nicht, dass
    der Wächter greift. Pro Zusage also: *welche Änderung am geprüften Code müsste diesen
    Test rot machen, und wurde sie einmal gesehen?* Wo die Antwort dauerhaft interessant
    ist, gehört sie in den Mutations-Sensor deines Repos (falls vorhanden); wo sie einmalig ist, in
    den Bericht. **Keine Antwort ist ein Befund**, kein Formfehler — die Klasse „Zusage greift
    weiter als Abdeckung" ist in der Praxis teuer erkauft.
20. **Jeden in diesem Lauf neu geschriebenen oder geänderten Kommentar gegen `AGENTS.md` §3.7
    prüfen** (Code, Konfiguration, Skripte). Die Probe: beschreibt der Satz den **Ist-Zustand**
    (indikativ, auflösbar), oder trägt er eine Slice-Nummer als Begründung, ein „(… , entschieden)"
    ohne Anker-Form, oder einen Konjunktiv über eine verworfene Alternative bzw. eine noch nicht
    existierende künftige Änderung (**„sobald Slice X das tut …"**)? Herkunft steht nur als **ein**
    auflösbares Feld in den dort genannten Formen (`LH-*`, `ADR-*`, `· seit welle-<NN>`, wellenlos
    `· seit slice-<NNN>`) — alles andere ist Zustand, keine Chronik, und wird vor der Übergabe
    umformuliert statt mitgeschleift.

Hier endet die Implementation. Die übrigen Rollen laufen in **getrennten Kontexten** (Modul 8).

## Übergaben an nachgelagerte Rollen (Modul 8 → 10 → 11)

21. **→ Reviewer (Code-Review, Modul 10):** den Diff + Plan-Verweis an einen **unabhängigen**
    Reviewer übergeben (`.harness/skills/reviewer.md`, frischer Kontext — kein Selbst-Review). Er
    kategorisiert Findings (HIGH/MEDIUM/LOW/INFO) in einen Report unter `docs/reviews/` und prüft
    den Diff gegen **Plan + ADR + Hard Rules** (nicht die DoD). HIGH/MEDIUM auflösen; ein HIGH mit
    Rollen-Konflikt folgt Modul 8 §Konflikt-Pfad (Sequenz mit Übergabe-Artefakten, nie
    „herabstufen, weil der Implementer widerspricht").
22. **→ Verifier (Modul 11):** in getrenntem Kontext die DoD-/Spec-Behauptung und den
    Plan-vs-Code-Diff **bestätigen**, dazu ADR-Konformität. Das fängt, was Tests übersehen und der
    Reviewer nicht sieht (DoD-Verletzung).
23. **→ Validator (Modul 8):** falls der Slice End-Nutzer-Wert liefert, gegen den realen Bedarf
    validieren („das Richtige bauen"). Meist n/a bei interner Wartung — dann explizit sagen statt
    still überspringen.

## Closure — Planner-Rolle (Modul 8 + Modul 5)

24. Erst wenn der Review konform **und** die Verifikation die DoD bestätigt hat, schließt der
    **Planner**: die Closure-Notiz mit einem **Steering-Loop-Eintrag** schreiben (geschärfte Regel ·
    neuer Sensor · benannte Spec-Lücke — Modul 5: der `→ done`-Übergang verlangt einen Lerneintrag,
    nicht nur grüne Gates), dann den Slice `in-progress → done` verschieben —
    **`make slice-mv SLICE=slice-<Kennung> TO=done`**, derselbe Aufruf wie in Schritt 9: reiner
    Move als eigener Commit, Verweis-Nachzug getrennt davon (Hard Rule 3.3). Ein rotes Gate erreicht `done/` **nur** mit
    dokumentiertem Carveout (Modul 7), nie als stilles Rot. **Jedes offene Risiko aus dem Slice-Plan
    bekommt dabei genau einen von drei Ausgängen** (Modul 5): *eingetreten* → Carveout oder
    Folge-Slice mit ID · *entfallen* → gestrichen **mit Begründung** · *weiter offen* → wandert ins
    Beobachtungs-Register (Schritt 25). Ein Slice geht nicht nach `done/`, während ein Risiko ohne
    Ausgang dasteht.
25. **Das Beobachtungs-Register fortschreiben** (`docs/plan/planning/observations/`, Modul 6) —
    der **Schreib**-Schritt, und er hängt an der Closure, nicht an der Implementation. Für jede
    Beobachtung aus der Closure-Notiz: führt das Register die Klasse schon, dann die vorhandene
    Kennung `BEO-<KUERZEL>/<slug>` **zitieren** und eine weitere Datei in ihrem `evidence/` anlegen
    — wer neu formuliert, spaltet eine Klasse in zwei Pfade, und keiner der beiden erreicht je 3×.
    Sonst ein neues Verzeichnis `BEO-<KUERZEL>/<slug>/` mit `observation.md` und `state.md` anlegen
    — Kürzel aus der Modus-Deklaration nachschlagen, nicht erfinden; das Register ist zugleich die
    Vergabestelle für den `<slug>`-Teil. Der Beleg ist **formgebunden**: `evidence/slice-<NNN>.md`,
    kein Freitext, eine Datei je Auftreten. Geschrieben wird er **vor** dem Move aus Schritt 24 — die
    Slice-Datei liegt dann noch nicht in `done/`, und das ist richtig so, weil Move und Inhalt
    getrennt committen (Hard Rule 3.3). Der Zähler wird **nicht gesetzt**, er ist die Zahl der
    Evidence-Dateien und **folgt** aus ihnen. **Bei null Beobachtungen** bleibt die Ablage
    unverändert und die Closure-Notiz trägt den Satz *keine Beobachtung angefallen*: das Auslassen
    ist keine Antwort. Erreicht ein Eintrag **mit diesem Slice** 3× (die Zahl seiner
    Evidence-Dateien), wandert er in die Steering-Loop-Einträge der laufenden Welle-Closure
    (`/close-welle`); läuft keine Welle, löst die Slice-Closure den Lese-Schritt selbst aus, und der
    Herkunfts-Anker lautet dann `seit slice-<NNN>` statt `seit welle-<NN>`.

Gates nicht überspringen. Keine Erfolgsmeldung ohne Command-Ausgabe.
