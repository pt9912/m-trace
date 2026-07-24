## Modul 4 — Architektur und ADRs

<!-- Quelle: [01-spec-und-architektur/modul-04-architektur-adrs.md](https://github.com/pt9912/ai-harness-course/blob/v3.5.1/kurs/de/01-spec-und-architektur/modul-04-architektur-adrs.md) -->

### Mini-Glossar für dieses Modul (Modul 4)

| Begriff | Ein-Satz-Definition | Bild im Kopf |
|---|---|---|
| **MADR** | Markdown-basiertes ADR-Format mit Kopf-Feldern (Status, Datum, Bezug, Supersedes) und Body-Blöcken (Kontext, Optionen mit Trade-offs, Entscheidung, Konsequenzen). | ein Formular, das die Entscheidung zwingt, ihre Belege mitzubringen. |
| **Nygard-Format** | Das ursprüngliche, schlankere ADR-Format nach Michael Nygard: Kontext, Entscheidung, Konsequenzen. | der Urahn von MADR — gleiche Idee, weniger Felder. |
| **superseded** | ADR-Status: Entscheidung ist durch eine *neue* ADR abgelöst — der Bedarf bleibt, die Antwort wechselt. | Schild "ersetzt durch Nr. N" am alten Protokoll. |
| **deprecated** | ADR-Status: Entscheidung entfällt *ersatzlos* — der zugrunde liegende Bedarf existiert nicht mehr. | Akte geschlossen, kein Nachfolger nötig. |
| **Fitness-Function-Werkzeuge** | ArchUnit (Java), dep-cruiser (JS/TS), import-linter (Python) — prüfen Architektur-Aussagen maschinell, z. B. Layer-Importregeln. | der Prüfstand, auf den die ADR-Aussage geschnallt wird. |

### Harness-Einordnung (Modul 4)

ADR = *inferential feedforward* (für den Implementation-Agent) und
gleichzeitig Quelle für *computational feedback* (ArchUnit/Fitness
Functions, wenn die Entscheidung maschinell prüfbar ist). Eine ADR ohne
Fitness Function ist eine Absichtserklärung.

### Kernidee (Modul 4)

Ein ADR ist die einzige Stelle, an der "weil" gegen "ist halt so" gewinnt.
Wenn dein Reviewer-Agent den Grund nicht findet, kann er die Entscheidung
nicht verteidigen.

### Hard Rule (Beispiel aus c-hsm-doc, ADR 0001)

Begriff *Hard Rule* siehe Glossar in
[`grundlagen/konventionen.md`](grundlagen-konventionen.md).

*"Eine ADR mit Status `Accepted` wird nicht inhaltlich überschrieben.
Spätere Korrekturen oder Schärfungen entstehen als neue ADR mit
explizitem Verweis auf die abgelöste oder geschärfte Vorgängerin."*

Wirkung: ADRs sind Geschichtsdokumente, kein Wiki. Reviewer-Agent kann
auf ältere Entscheidungen vertrauen, ohne Versionsstände zu vergleichen.

### Regeln gegen typische Fehlannahmen (Modul 4)

- Nein. ADRs begründen die *Lösung*. Anforderungen begründet die Spec. Wer ADRs zur Spec macht, kann später keine Architektur ohne Lastenheft-Änderung wechseln.
- Hard Rule: Accepted-ADRs werden nicht überschrieben. Folge-ADR mit `supersedes ADR-N`. Sonst kann der Reviewer-Agent nicht auf ältere Entscheidungen vertrauen.
- Eine ADR ohne Fitness Function ist eine Absichtserklärung. Wer architecture fitness im Kopf hat, schreibt parallel den ArchUnit-Test.
- MADR ist ein Format unter mehreren (auch Nygard, Tyree/Akerman). Wichtig ist, dass dein Repo *eines* konsequent benutzt.
- Diagramme sind *eine* Output-Form, nicht die Sache selbst. Architektur in diesem Kurs heißt: *Entscheidungen mit Begründung (ADR), prüfbar gemacht (Fitness Function), versioniert (Accepted-Hard-Rule)*. Ein Diagramm ohne ADRs hinter sich ist Wandtapete; eine ADR ohne Fitness Function ist Absichtserklärung. `spec/architecture.md` ist explizit *diagrammatisch und enthält keine eigenen Anforderungen* (siehe Spec-Stratifizierung in [`grundlagen/konventionen.md#spec-stratifizierung`](grundlagen-konventionen.md#spec-stratifizierung)) — genau weil sonst Bilder anfangen würden, die ADR-Schicht zu ersetzen.
- Eine ADR ohne maschinelle Durchsetzung ist eine *Absichtserklärung*, die der Implementation-Agent freundlich liest und dann ignoriert, wenn ein anderer Pfad "einfacher" wirkt. Eine ADR *mit* Fitness Function ist ein Constraint — die Layering-Regel, die ArchUnit dem Agenten als roten Build entgegenhält. Die Übersetzung (ADR-Satz → Werkzeug → Make-Target → Failure-Beispiel) steht kompakt in [Modul 13 §Fitness Function aus einem ADR-Satz](modul-13-quality-gates.md#adr-zur-fitness-function). Wer das nicht macht, dokumentiert *Hoffnung*.

### Ziel-Form: ADR (MADR)

Die Form liefert die Vorlage
[`templates/docs/plan/adr/NNNN-titel.template.md`](../templates/docs/plan/adr/NNNN-titel.template.md):
Kopf (Status · Datum · Bezug · Supersedes) plus Body (Kontext · Verglichene
Alternativen · Entscheidung · Konsequenz mit Fitness Function). Operative
Regeln zur Form:

- Der Kontext *referenziert* die Anforderung, wiederholt sie nicht.
- Mindestens drei Verglichene Alternativen, jede mit Trade-off.
- Jede Entscheidung mit Architektur-Wirkung bekommt eine Fitness Function —
  sonst ist sie Absichtserklärung.
- `Accepted` wird nie überschrieben — Korrektur = Folge-ADR mit `Supersedes`.
