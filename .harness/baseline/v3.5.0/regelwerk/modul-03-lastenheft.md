## Modul 3 — Lastenheft und Spezifikation

<!-- Quelle: [01-spec-und-architektur/modul-03-lastenheft.md](https://github.com/pt9912/ai-harness-course/blob/v3.5.0/kurs/de/01-spec-und-architektur/modul-03-lastenheft.md) -->

### Harness-Einordnung (Modul 3)

Spec = *inferential feedforward* (siehe
[`grundlagen/klassifikation.md`](grundlagen-klassifikation.md)).
Sie ist die billigste Kontrolle: Was die Spec sauber ausschließt, kommt
im Review nicht mehr vor.

### Kernidee (Modul 3)

Ein Agent ist ein extrem buchstabengetreuer Praktikant. Was nicht in der
Spec steht, existiert für ihn nicht — Lopopolos Maxime: *"Was der Agent
nicht im Kontext erreicht, existiert für ihn nicht."* Was zweideutig in der Spec
steht, wird auf die für dich ungünstigste Weise interpretiert.

**Grenze der Metapher.** Die Praktikant-Metapher trägt nur die
*Buchstabentreue*. Anders als ein echter Praktikant **vergisst** der
Agent zwischen den Aufgaben — was nicht im Kontext steht, war für ihn
nie da (siehe Glossar in
[`grundlagen/konventionen.md#kernbegriffe`](grundlagen-konventionen.md#kernbegriffe):
LLM ist *stateless*). Wer die Metapher zu weit treibt, erwartet
"Mitlernen" — und plant Reviews, als würden sie *einmal* erklärt
ausreichen. Sie reichen nicht. Jeder Lauf beginnt bei Null.

### Regeln gegen typische Fehlannahmen (Modul 3)

- Happy Path widerlegt nur die These "es funktioniert gar nicht". Boundary und Negative widerlegen die stillen Annahmen, *die ein Agent am liebsten als selbstverständlich behandelt*.
- Im Gegenteil: ein Satz "das System *darf nicht* …" spart später drei Reviews. Negativ ist genauso präzise wie positiv.
- Nein, Performance gehört in den nichtfunktionalen Block der Spec (oder in `spec/spezifikation.md`, wenn stratifiziert). Der ADR begründet, *wie* man die Schwelle einhält.
- Was nicht explizit ausgeschlossen ist, baut der Agent plausibel mit. Das ist die häufigste Quelle für "wir hatten das nie gefordert"-PRs.
- Falsch. Lopopolos Maxime *"Was der Agent nicht im Kontext erreicht, existiert für ihn nicht"* ist ein Plädoyer *für* Kontext-Verfügbarkeit — und sagt damit, dass Spec und Prompt *unterschiedliche* Lebenszyklen haben: Spec wird *gepflegt* (Versions-Geschichte, Bezüge, Audit), Prompt wird *für einen Lauf zusammengestellt*. Was im Prompt steht, aber nicht in der Spec, gilt nur für *diesen* Lauf — der nächste Agent sieht es nicht. Das Muster (Spec sagte *speichert*, Agent baute PostgreSQL) wäre mit einem Mega-Prompt nicht besser geworden — der Prompt würde im nächsten Lauf vergessen.

### Ziel-Form: Akzeptanzkriterium

Anforderungen leben im Lastenheft
([`templates/spec/lastenheft.template.md`](../templates/spec/lastenheft.template.md)).
Ein funktionales Kriterium trägt eine `<PREFIX>-FA-NNN`-ID und dann drei Pfade
im Given/When/Then-Stil — **Happy · Boundary · Negative** — plus einen
**Out-of-Scope**-Block. Vagen Satz zuerst auf Mehrdeutigkeiten prüfen (*was
genau · welche Felder · welcher Speicherort*), bevor die Pfade formuliert
werden; das Negative (`darf nicht …`) spart die spätere Review.

### Spec-Stratifizierung — Drei Schichten (Modul 3)

Reifere Specs zerfallen in drei Schichten mit eigener Precedence:
`lastenheft.md` (vertragliches *Was*) › `spezifikation.md` (präzisiertes
*Wie genau*) › `architektur.md` (strukturelles *Wodurch*). Konfliktregel:
**Lastenheft sticht Spezifikation sticht Architektur** — die untere
Schicht darf *präzisieren*, nie *erweitern*. Vollform (Straten-Klassen,
Referenz-Richtung, `check-references`-Gate) in
[`konventionen.md` §Spec-Stratifizierung](grundlagen-konventionen.md#spec-stratifizierung).
Vorlagen: [`spec/`-Templates](../templates/spec/).

