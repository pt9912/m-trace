# AGENTS.md — Briefing für AI-Coding-Agenten

> **Template-Hinweis.** Diese Datei ist eine Vorlage für die
> Repo-Root-`AGENTS.md`. Kopiere nach `AGENTS.md` deines Repos,
> ersetze `<Platzhalter>` und lösche diesen Block. AGENTS.md *trägt
> Hard Rules und Pointer auf kanonische Quellen*, sie *dupliziert deren
> Inhalt nicht* — sonst entsteht Drift (siehe
> [Kurs Modul 9](https://github.com/pt9912/ai-harness-course/blob/v3.5.1/kurs/de/03-agenten/modul-09-implementierung.md)).
> **Pointer-Artefakt:** verweist auf andere kanonische Quellen — zuletzt
> füllen bzw. re-syncen, sobald die Ziele stehen; veraltete
> `(folgt)`/Klartext-Verweise fängt kein Linter (Reviewer-Sache).

---

## 1. Was diese Datei ist

Onboarding-Briefing für jede AI-Session, die in diesem Repo Code oder
Dokumentation ändert. Sie verweist auf die kanonischen Quellen und
formuliert die Hard Rules, die der Implementation-Agent immer
einhalten muss.

**Bei Konflikt zwischen dieser Datei und einer kanonischen Quelle gilt
die kanonische Quelle** (Source Precedence — siehe
`harness/README.md`).

Strukturregeln (ID-Schemata, Verzeichniskonvention, Adaptionen ggü.
Baseline, Modus-Deklarationen pro Sub-Area, Zusatzklassen für
Sensors-Bindung) leben in
[`harness/conventions.md`](harness/conventions.md).

Das **Regelwerk der adoptierten Baseline** ist die **präsente,
nachschlagbare Vertiefung** zu diesem Briefing: ein self-navigierbares
**Modul-Bundle** (`README.md` = Index). Beim Bootstrap wird das
self-contained Release-ZIP
(<https://github.com/pt9912/ai-harness-course/releases/download/v3.5.1/lab-regelwerk.zip>)
**committet vendored** unter `.harness/baseline/<tag>/{regelwerk,templates}/`
(Regelwerk *und* Templates parallel, netzlos materialisiert samt `SHA256SUMS`
— Vorgehen siehe
[Kurs Modul 2 §Bootstrap](https://github.com/pt9912/ai-harness-course/blob/v3.5.1/kurs/de/01-spec-und-architektur/modul-02-harness-bootstrap.md);
Quelle/Stand in [`harness/conventions.md`](harness/conventions.md) §Baseline).

Die verkörperte Form (dieses Briefing, die Konventionen, deine
ausgefüllten Artefakte) **führt**; das Regelwerk wird **pro Entscheidung
nachgeschlagen, deren
operative Detailtiefe das Briefing nicht trägt** — Trigger-Klassen,
Sub-Area-Qualifikation, Carveout-vs-Reconciliation, Modus-Diagnose. Dabei
**nur den benötigten Abschnitt** laden (README ist der Index), **nicht das
ganze Regelwerk im Kontext halten**. Breiterer Pflicht-Blick bleibt bei:
Bootstrap, Änderung an [`harness/conventions.md`](harness/conventions.md)
(Adaptionen `MR-<NNN>`, Source-Precedence, ID-Schema), Drift-Audit gegen die
Baseline. Derivativ: bei Konflikt gelten die kanonischen Quellen.

Die **Skelett-Vorlagen** der Baseline liegen **vendored** unter
`.harness/baseline/<tag>/templates/` (aus demselben Baseline-Bundle) und
tragen zwei Rollen: als **Referenz-Form**, auf die das Regelwerk mit
`../templates/…` als „Ziel-Form" verweist (netzlos, weil parallel zu
`regelwerk/` vendored), und als Vorlage, die beim Anlegen neuer Artefakte
(ADR, Slice, Welle, Carveout, Review-Report) **kopiert und ausgefüllt** wird
statt frei zu formulieren.

## 2. Kanonische Quellen (Source Precedence)

In dieser Reihenfolge:

1. [`spec/lastenheft.md`](spec/lastenheft.md) — vertraglich abnahmebindend.
2. [`spec/spezifikation.md`](spec/spezifikation.md) — technisch verbindlich, fortschreibbar. *(Optionales 3. Spec-Stratum — siehe Spec-Stratifizierung. Repos mit 2 Straten löschen diese Zeile und nummerieren die Ränge neu.)*
3. [`spec/architecture.md`](spec/architecture.md) — Komponenten- und Sequenzsicht.
4. [`docs/plan/adr/`](docs/plan/adr/) — ADR-Verzeichnis und -Index.
5. [`docs/plan/planning/in-progress/roadmap.md`](docs/plan/planning/in-progress/roadmap.md) — aktuelle Welle.
6. [`README.md`](README.md) — Projekt-Überblick.
7. **AGENTS.md (diese Datei).**
8. [`harness/README.md`](harness/README.md) — Harness-Einstieg.

## 3. Harte Regeln

<!--
Eigene Hard Rules ergänzen, basierend auf der Repo-Klasse. Beispiele
zur Inspiration:
-->

### 3.1 Docker-only

<!-- Wenn das Repo Docker-only ist (typisch für Multi-Toolchain-Repos): -->

Kein lokales <venv/SDK/Toolchain-Install>. Alles läuft über `make`
(das Docker nutzt). Host braucht nur Docker und GNU `make`.

**Falsch:** <z.B. `pip install ...`>
**Richtig:** <z.B. `make test`>

**Begründung:** Toolchain-Reproduzierbarkeit + Supply-Chain-Defense.

### 3.2 Suppression-Verbot

<!--
Pro Sprache eine Variante. Beispiele:
- Python: # noqa
- Go: //nolint
- C#: #pragma warning disable, [SuppressMessage]
- Kotlin: @Suppress
- Java: @SuppressWarnings
-->

Inline-Suppression bricht das `<suppression>-gate`. Ausnahmen leben in
<zentraler Konfigurations-Datei> mit Begründung.

### 3.3 git mv + Inhaltsänderung = zwei Commits

Wenn eine Datei verschoben **und** der Inhalt umgeschrieben wird:

1. `git mv source target` → eigener Commit (reiner Move, Git erkennt R-Rename).
2. Inhalt umschreiben → zweiter Commit.

**Begründung:** Sonst fällt die Rename-Detection unter die 50%-
Similarity-Schwelle und `git log --follow` wird unzuverlässig.

### 3.4 Architektur ist sprach- und meilensteinfrei

`spec/architecture.md` referenziert ADRs und Modul-Pfade, aber **keine**
Wellen, Slices, Commit-Hashes oder Closure-Daten. Die zeitliche Schicht
lebt in `docs/plan/planning/` und den späteren Closure-Notizen.

### 3.5 ADRs sind nach `Accepted` immutable

Eine ADR mit Status `Accepted` wird nicht inhaltlich überschrieben.
Korrekturen entstehen als neue ADR mit `Supersedes ADR-NNNN`.

### 3.6 Gates dürfen nicht ohne ADR gelockert werden

Jede Schwellen-Senkung (Coverage, Linter-Strenge, Architekturregel)
ist ein ADR, kein PR-Kommentar.

<!--
Repo-spezifische Hard Rules ergänzen, z.B. für Safety/Control:
- "Optimierer darf nie direkt aufs Gerät schreiben."
- "Protokoll-Adapter dürfen keine Marktentscheidungen enthalten."
- "Produktion-Profile müssen fail-closed sein."
-->

## 4. Quality Gates

<!--
Nur Befehle aufzählen, die im Makefile *existieren*. Halluzinierte
Gates sind die häufigste Form von Harness-Lüge (siehe Modul 13).
-->

| Target | Zweck |
|---|---|
| `make lint` | <…> |
| `make test` | <…> |
| `make arch-check` | <…> |
| `make coverage-gate` | <…> |
| `make gates` | alle inneren Gates (mandatory vor PR) |
| `make ci` | CI-äquivalent (gates + zusätzliche) |
| `make fullbuild` | volle Closure (vor Welle-Merge) |

## 5. Dokumentations-Regeln

- Requirement- und Architektur-IDs müssen in PRs/Commits referenziert
  sein. Vergeben werden IDs beim Spec-/ADR-Schreiben nach dem in
  `harness/conventions.md` deklarierten ID-Schema (Default:
  `<PREFIX>-FA-<NN>` / `<PREFIX>-QA-<NN>` aus dem Lastenheft, ADR-Nummern
  über den ADR-Index) — nie ad hoc im PR.
- Neue ADRs müssen den ADR-Index aktualisieren.
- Roadmap/Status-Geschichte lebt in `docs/plan/planning/`, nicht in `spec/architecture.md`.
- Quality-Gate-Definitionen leben in <`docs/user/quality.md` oder Äquivalent>.

## 6. Minimal Agent Workflow

Pro Slice:

1. `harness/README.md` lesen.
2. Relevante kanonische Quelle lesen (Source Precedence beachten).
3. Betroffene Requirement-/ADR-IDs identifizieren.
4. Kleinste sinnvolle Änderung planen.
5. Engsten nützlichen Sensor laufen lassen.
6. Repo-weiten Gate-Lauf vor Handoff (`make gates`).
7. Doku/Indizes aktualisieren, falls ein öffentlicher Vertrag berührt.
8. Ausgeführte Sensors und verbleibende Risiken berichten — keine Erfolgsmeldung ohne Gate-Ausführung.
