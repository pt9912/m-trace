# Harness-Konventionen

> **Template-Hinweis.** Diese Datei ist eine Vorlage für
> `harness/conventions.md` deines Repos. Kopiere sie nach
> `harness/conventions.md`, ersetze `<Platzhalter>` und lösche
> diesen Block. Pflichtgliederung folgt
> [Kurs Konventionen / harness/conventions.md als Konventionsspeicher](https://github.com/pt9912/ai-harness-course/blob/v3.5.1/kurs/de/grundlagen/konventionen.md#harnessconventionsmd-als-konventionsspeicher).
>
> **Was diese Datei trägt:** repo-lokale Strukturregeln und Adaptionen
> ggü. der adoptierten Harnesskonvention (Baseline). Sie ist
> **Pflicht** (Existenz), die Form (Einzeldatei vs. Verzeichnis,
> ADR-artig vs. Prosa) ist **Wahl**.
>
> **Was diese Datei NICHT trägt:** Kurs- oder Baseline-Konventionstext
> wird nicht dupliziert — Pointer reichen. Sonst drift gegen die Quelle.

---

## Purpose

Diese Datei deklariert die *repo-lokalen* Strukturregeln dieses Repos
gegenüber der adoptierten Harnesskonvention (Baseline). Sie ist der
Default-Ort für:

- **Adaptionen** ggü. der Baseline (mit Begründung und Auflösungs-Trigger).
- **ID-Schema-Deklaration** — welches Präfix-Schema dieses Repo nutzt.
  Der Baseline-Default wird als Teil der `MR-000`-Aussage festgehalten;
  ein abweichendes Präfix oder Schema ist ein eigener `MR`-Eintrag.
- **Zusatzklassen-Deklarationen** für repo-spezifische
  Bindung-Klassen in der Sensors-Tabelle, die über die vier kanonischen
  hinausgehen (ADR, Carveout, Schwelle, Reproduzierbarkeit).
- **Modus-Deklarationen** pro Sub-Area (Greenfield / Brownfield /
  Hybrid) inklusive Konvergenz-Auftrag bei BF.

Bei Konflikt zwischen dieser Datei und einer kanonischen Quelle gilt die
kanonische Quelle (Source Precedence). Diese Datei ist konformitäts-
bringend für *Form*-Fragen, nicht autoritativ über Inhalt.

## Baseline

<!--
Welche Harnesskonvention wird adoptiert? Stand und Datum festhalten,
damit spätere Adaptionen einen Bezugspunkt haben.
-->

- **Konvention:** <Name, z. B. "AI-Harness-Kurs", interner Standard, Industrie-Norm>
- **Stand:** <Datum oder Version, z. B. "Template-Set 2026-06">
- **Datum der Adoption:** <Datum>

## Adoptierte Konventions-Quellen

<!--
Pointer auf die Quellen der Baseline. KEINE Wiederholung des Inhalts —
nur Verweise.

Das Agenten-Regelwerk ist die Quelle, die ein Code-Agent statt des
vollen Lehrmaterials liest (operatives Regelwerk ohne Didaktik). Es
ist derivativ — bei Konflikt gilt das Lehrmaterial.
-->

- **Extern (Lehrmaterial):** <Pfad oder URL>
- **Vendored Baseline (Regelwerk + Templates):** aus dem self-contained
  Release-Asset
  https://github.com/pt9912/ai-harness-course/releases/download/v3.5.1/lab-regelwerk.zip
  nach `.harness/baseline/<tag>/{regelwerk,templates}/` entpackt (netzlos,
  `SHA256SUMS`) — adoptierten Stand notieren (Stand-Zeile in
  `regelwerk/README.md`, z. B. „Kurs-Welle 24 · 2026-07-16"; Wellen-Register:
  CHANGELOG.md im Kurs-Repo); für harte Reproduzierbarkeit das Asset eines Tags
  ziehen statt `latest`. Rollen und Netzlosigkeit: §MR-003.
- **In-Repo (verkörperte Form):** <Pfade zu deinen kopiert-und-ausgefüllten
  Artefakten> — die vendored `.harness/baseline/<tag>/templates/` sind die
  Referenz-Form („Ziel-Form" des Regelwerks); deine eigenen Dateien sind daraus
  kopiert und ausgefüllt.

## Adaptions-Block

<!--
ADR-artige Liste der Abweichungen ggü. Baseline.
Jeder Eintrag mit Pflichtfeldern: ID (MR-<NNN>), Datum, Geltungsbereich,
Adaption, Begründung, Auflösungs-Trigger (oder "permanent").

Disziplin: chronologisch nummeriert, keine nachträglichen
inhaltlichen Änderungen an akzeptierten Einträgen — nur neue Einträge
oder explizite Aufhebungen via neuen MR.

Zum ID-Schema: Hier wird es nur DEKLARIERT — VERGEBEN werden IDs beim
Schreiben der Artefakte: Anforderungs-IDs im Lastenheft
(`<PREFIX>-FA-<NN>` / `<PREFIX>-QA-<NN>`, Schema-Definition in
spec/lastenheft.template.md), Verfeinerungen in der Spezifikation
(`<PREFIX>-FA-<NN>.<Buchstabe>`), ADR-Nummern chronologisch über den
ADR-Index. Hintergrund: Kurs grundlagen/konventionen.md
§ID-Schema als Klammer. Agenten referenzieren IDs nur, sie erfinden
keine (AGENTS.md §5/§6).
-->

### MR-000 — Baseline-Aussage

- **Datum:** <Datum>
- **Geltungsbereich:** gesamtes Repo
- **Adaption:** *keine inhaltlichen Adaptionen ggü. Baseline-Default
  für Verzeichniskonvention, Lifecycle-Regeln, Carveout-Disziplin,
  ID-Schema (`<PREFIX>-FA-*`, `<PREFIX>-QA-*`, `ADR-<NNNN>`, `CO-<NNN>`,
  `slice-<NNN>`, `MR-<NNN>` — Präfix repo-weit festlegen, z. B. `LH`).*
  (Source-Precedence-Adaptionen werden in separaten `MR-<NNN>`
  dokumentiert — siehe Beispiel `MR-001` unten, weil das mitkopierte
  README-Template eine 3-Schichten-Spec-Precedence zeigt.)
- **Begründung:** Initial-Setzung. Spätere Adaptionen werden als
  `MR-<NNN>` nachgetragen.
- **Auflösungs-Trigger:** permanent.

<!-- Beispiel-Eintrag für eine konkrete Adaption — passt zum 3-Schichten-
     Spec-Default des README-Templates. Anpassen, wenn dein Repo das
     anders handhabt, oder entfernen, wenn du eine 2-Schichten-Spec-
     Precedence wählst. -->

### MR-001 — Source Precedence mit eigener Spezifikations-Schicht

- **Datum:** <Datum>
- **Geltungsbereich:** `harness/README.md` §Source precedence
- **Adaption:** Source-Precedence-Tabelle führt `spec/spezifikation.md`
  als eigenen **Rang 2** zwischen Lastenheft (Rang 1) und Architektur
  (Rang 3). Der Kurs-Default
  ([`konventionen.md` §Source Precedence](https://github.com/pt9912/ai-harness-course/blob/v3.5.1/kurs/de/grundlagen/konventionen.md#source-precedence))
  setzt zwei Spec-Ränge (`lastenheft` → `architecture`); dieses Repo
  nutzt drei.
- **Begründung:** Das Repo verwendet die Spec-Stratifizierung
  ([`konventionen.md` §Spec-Stratifizierung](https://github.com/pt9912/ai-harness-course/blob/v3.5.1/kurs/de/grundlagen/konventionen.md#spec-stratifizierung))
  explizit mit drei Spec-Dateien. Damit die Source-Precedence-Tabelle
  die ADR-Schärfungs-Regel ("ADR darf Spezifikation schärfen, nicht
  Lastenheft") strukturell abbildet, muss die Spezifikation als
  eigener Rang sichtbar sein.
- **Auflösungs-Trigger:** permanent.

### MR-003 — Regelwerk und Templates als vendored, nachschlagbare Baseline

- **Datum:** <Datum>
- **Geltungsbereich:** [`AGENTS.md`](../AGENTS.md) §1, [§Baseline](#baseline)
- **Adaption:** Provenienz/Konkretisierung (keine inhaltliche Abweichung vom
  Baseline-Default): Regelwerk *und* Templates der Baseline werden **committet
  vendored** unter `.harness/baseline/<tag>/{regelwerk,templates}/` geführt —
  beim Bootstrap aus dem self-contained `lab-regelwerk.zip` materialisiert
  (Modul 2), netzlos auf jedem Checkout, Integrität über
  `.harness/baseline/<tag>/SHA256SUMS`. Das Regelwerk ist die **präsente,
  nachschlagbare Vertiefung**: pro Entscheidung, deren operative Detailtiefe
  Briefing und Konventionen nicht tragen (Trigger-Klassen,
  Sub-Area-Qualifikation, Carveout-vs-Reconciliation, Modus-Diagnose), wird
  der **relevante Abschnitt** referenziert (README = Index), ohne das ganze
  Regelwerk im Kontext zu halten. Weil `regelwerk/` und `templates/` **parallel**
  vendored liegen, lösen die `../templates/`-Ziel-Form-Verweise des Regelwerks
  **netzlos lokal** auf (Referenz-Form für „so sieht das Artefakt aus").
- **Begründung:** Modul-0-Prinzip — *Per-Lauf-Relevantes gehört verkörpert,
  nicht extern nachgeladen*: Da Regelwerk und Ziel-Formen im Slice-Betrieb
  routinemäßig nachgeschlagen werden (Real-Beleg eines Konsument-Repos:
  d-check), werden sie vendored statt pro Lauf extern gefetcht — das macht den
  Nachschlag netzlos und über den `<tag>` reproduzierbar. Der konkrete Tag und
  das Integritätsmanifest stehen hier als Provenienz, damit Drift gegen die
  adoptierte Baseline prüfbar bleibt. (Die *Kontext*-Hygiene bleibt: nur der
  benötigte Abschnitt, nie das ganze Bundle im Kontext.)
- **Auflösungs-Trigger:** permanent (Provenienz/Baseline-Konformität).

<!-- Weitere konkrete Adaptionen wie folgt: -->

### MR-NNN — <Titel der Adaption>

- **Datum:** <Datum>
- **Geltungsbereich:** <Dateien / Module / Sub-Areas>
- **Adaption:** <was weicht inhaltlich ab>
- **Begründung:** <warum, idealerweise mit Praxis-Bezug>
- **Auflösungs-Trigger:** <Trigger oder "permanent">

## Zusatzklassen-Deklaration für Sensors-Bindung

<!--
Die vier kanonischen Bindung-Klassen der Sensors-Tabelle in
`harness/README.md` (ADR, Carveout, Schwelle, Reproduzierbarkeit) sind
ohne Deklaration legitim.

Repos können weitere Klassen einführen — z. B. Anforderungs-Bindung
(`LH-...`), Compliance-Bindung (Regulatorik-Artikel), Modell-Version-
Bindung (für KI-Evals). Diese müssen hier deklariert werden, sonst sind
sie für Reviewer nicht von Tippfehlern unterscheidbar.

Eine nicht-deklarierte Zusatzklasse in der Sensors-Tabelle ist eine
stille Setzung und damit Harness-Lüge in derselben Klasse wie ein
halluziniertes Gate (Modul 13).
-->

| Klasse | Form | Bedeutung | Beispiel |
|---|---|---|---|
| <z. B. LH-Bindung> | `LH-<...>` | <z. B. Gate prüft eine bestimmte LH-Anforderung> | <z. B. `LH-QA-01` für Determinismus-Gate> |

<!-- Wenn keine Zusatzklassen verwendet werden: Tabelle entfernen oder
"— keine —" eintragen. -->

## Modus-Deklaration pro Sub-Area

<!--
Pro Modul / Verzeichnis / Sub-Area: Modus festlegen.
- Greenfield (GF): Doc führt, Code folgt. Steady-State.
- Brownfield (BF): Code führt, Doc folgt. Übergangsmodus mit
  Konvergenz-Auftrag zu GF. Graduation-Bedingung benennen.
- Hybrid: gemischt pro Sub-Sub-Area.
- Permanent-BF (selten): nur für Code, der absehbar entfernt wird;
  mit Begründung und Folge-Slice analog zu permanentem Carveout.

Eine Sub-Area in BF *ohne* Graduation-Plan ist eine Harness-Lüge:
"permanente Ausnahme als temporär getarnt" (Modul 7 Analogie).
-->

| Sub-Area (Pfad / Modul) | Modus | Begründung | Graduation-Bedingung / Folge-Slice |
|---|---|---|---|
| `*` (Default für gesamtes Repo) | <Greenfield / Brownfield / Hybrid> | <warum> | <Bedingung oder "n/a (GF)" oder "permanent + slice-Ref"> |

## Glossar (optional)

<!--
Repo-spezifische Begriffe, die im Kurs-Glossar nicht stehen.
Nur ergänzen, nicht Kurs-Glossar wiederholen.
-->

| Begriff | Bedeutung |
|---|---|
| <repo-spezifischer Begriff> | <Bedeutung in diesem Repo> |
