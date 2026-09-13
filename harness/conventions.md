# Harness-Konventionen

## Zweck

Diese Datei deklariert die repo-lokalen Strukturregeln von m-trace gegenüber
der adoptierten Harness-Baseline. Sie ist der Default-Ort für:

- **Adaptionen** ggü. der Baseline (mit Begründung und Auflösungs-Trigger).
- **ID-Schema-Deklaration** — welches Präfix-Schema m-trace nutzt
  ([MR-003](#mr-003), da m-trace keinen `MR-000`-Adoptionseintrag führt —
  die Baseline-Adoption selbst steht in §Baseline).
- **Zusatzklassen-Deklarationen** für repo-spezifische Bindung-Klassen in der
  Sensors-Tabelle, die über die vier kanonischen hinausgehen (ADR, Carveout,
  Schwelle, Reproduzierbarkeit).
- **Modus-Deklarationen** pro Sub-Area (Greenfield / Brownfield / Hybrid)
  inklusive Konvergenz-Auftrag bei BF.

Bei Konflikt zwischen dieser Datei und einer kanonischen Quelle gilt die
kanonische Quelle (Source Precedence). Diese Datei ist konformitätsbringend
für Form-Fragen, nicht autoritativ über Inhalt.

## Baseline

- **Konvention:** AI-Harness-Kurs (ai-harness-course)
- **Stand:** `v6.8.0`
- **Datum der Adoption:** 2026-09-13. Frühere Stände bleiben unter
  `.harness/baseline/<tag>/` liegen (Audit-Referenzform, nicht der aktive
  Stand).

**Was das Feld `Stand:` trägt:** den adoptierten Stand als **Version**, nie
als Datum — das Datum steht in der eigenen Zeile. Ein `versions`-Sensor, der
dieses Feld gegen Baseline-Pins im Repo hält, läuft in m-trace (noch) nicht
(`grep -n '^modules:' .d-check.yml` führt kein `versions`); das Feld steht
als Ziel-Form, nicht als bewachte Zusage (`welle-02` Tranche 10 evaluiert
das).

## Adoptierte Konventions-Quellen

- **Extern (Lehrmaterial, kanonisch):**
  <https://github.com/pt9912/ai-harness-course/tree/v6.8.0/kurs/de> — auf den
  Tag `v6.8.0` gepinnt, nicht `main`-floating.
- **Vendored Baseline (Regelwerk + Templates):** aus dem self-contained
  Release-Asset
  <https://github.com/pt9912/ai-harness-course/releases/download/v6.8.0/lab-regelwerk.zip>
  nach `.harness/baseline/v6.8.0/{regelwerk,templates}/` entpackt (netzlos,
  `SHA256SUMS`, per `sha256sum -c` verifiziert — 54 Dateien, Digest des
  Release-Assets: `2c55e6d1b821ae15ff73f5a9b3dc2269843db0ffcf9845a4bd0df2cfebbdc6c7`).
  Regelwerks-Stand laut `regelwerk/README.md`: Kurs-Welle 135, 2026-09-13.
- **In-Repo (verkörperte Form):** [`AGENTS.md`](../AGENTS.md),
  [`harness/README.md`](README.md) und diese Datei — kopiert-und-ausgefüllt
  aus den vendorten `.harness/baseline/v6.8.0/templates/`; bei Konflikt gilt
  das Lehrmaterial.

## Spec-Straten

| Stratum | Dateien | Rolle |
|---|---|---|
| Contract | `spec/lastenheft.md` | Bindende Anforderungen und Akzeptanzkriterien |
| Technical | `spec/backend-api-contract.md`, `spec/browser-support.md`, `spec/player-sdk.md`, `spec/telemetry-model.md` | Bindendes technisches Detail, das den Contract verfeinert |
| View | `spec/architecture.md` | Abgeleitete Komponenten-, Abhängigkeits- und Datenfluss-Sicht |

Die normative Referenz-Stabilität ist `Contract > Technical > View > ADR >
Planning`. Verweise zeigen aufwärts. Spec-Dokumente nutzen keine ADR- oder
Planning-Artefakte als normative Quelle. Abwärts-Provenienz ist nur in
ausgewiesenen History-Abschnitten erlaubt.

## Adaptions-Block

Regeln dieser Sektion: Diese Datei trägt den **Index**, nicht die Einträge.
Jede Adaption ist eine eigene Datei unter `harness/conventions/`, kopiert aus
`harness/conventions/MR-NNN-titel.template.md` der vendored Baseline; ist ihr
Auflösungs-Trigger eingetreten, wandert sie per `git mv` nach
`conventions/done/`. Der Zustand ist die Verzeichnis-Position, kein
Status-Feld. Der Grund für den Schnitt: Was hier steht, liest **jeder**
Agentenlauf — aufgelöste Adaptionen gehören nicht in diesen Pfad
(Baseline-Regelwerk `grundlagen-harness-dateien.md`
§harness/conventions.md als Konventionsspeicher).

### MR-000 — Baseline-Aussage

Bleibt hier: Sie ist keine Adaption, sondern die Adoptions-Erklärung, und
sie gilt für jeden Lauf.

- **Datum:** 2026-06-13 (Erstadoption, damals Commit-Pin)
- **Geltungsbereich:** gesamtes Repo
- **Ersetzt-Baseline-Regel:** — *(keine; dieser Eintrag ist die
  Adoptions-Erklärung, keine Adaption)*
- **Adaption:** keine inhaltlichen Adaptionen ggü. Baseline-Default für
  Verzeichniskonvention, Lifecycle-Regeln und Carveout-Disziplin. Das
  ID-Schema weicht ab — siehe [MR-003](#mr-003) (`F-*`/`NF-*`/`MVP-*`/
  `AK-*`/`RAK-*`/`R-*` statt `<PREFIX>-FA-*`/`<PREFIX>-QA-*`); m-trace zählt
  ohne Bereichssegment (ein schreibender Mensch + Agent, kein
  Mehr-Schreiber-Betrieb).
- **Begründung:** Initial-Setzung. Spätere Adaptionen werden als `MR-<NNN>`
  nachgetragen.
- **Auflösungs-Trigger:** permanent.

### Aktive Adaptionen

| MR | Titel | Geltungsbereich | Ersetzt-Baseline-Regel |
|---|---|---|---|
| [002](conventions/MR-002-accepted-adr-grandfathering.md) <a id="mr-002"></a> | Accepted-ADR-Grandfathering | `docs/plan/adr/0001-*.md`–`0007-*.md` | `modul-04-adrs.md` §Hard Rule für Accepted-ADRs |
| [003](conventions/MR-003-requirement-id-familien.md) <a id="mr-003"></a> | Requirement-ID-Familien | Contract, Pläne, Commits, Reviews | `grundlagen-source-precedence.md` §ID-Schema als Klammer |
| [006](conventions/MR-006-security-gate-carveout-registry.md) <a id="mr-006"></a> | Security-Gate-Carveout-Registry | `image-scan`/`vuln-check`-Gate | `modul-07-carveouts.md` §Ziel-Form: Carveout |
| [007](conventions/MR-007-planning-artefakt-form.md) <a id="mr-007"></a> | Planning-Artefakt-Form (Slice/Welle vs. `plan-<version>`) | `docs/plan/planning/` | `modul-05-planning-harness.md` §Ziel-Form: Slice |
| [008](conventions/MR-008-spec-straten-multi-datei.md) <a id="mr-008"></a> | Spec-Straten: Technical-Schicht auf vier Dateien | `spec/backend-api-contract.md`, `spec/browser-support.md`, `spec/player-sdk.md`, `spec/telemetry-model.md` | `modul-03-spec.md` §Ziel-Form: Spezifikation |

### Aufgelöste Adaptionen

| MR | aufgelöst durch |
|---|---|
| [001](conventions/done/MR-001-repository-pfade.md) <a id="mr-001"></a> | slice-006 (v3.5.0-Migration W5 — kein Nachfolger-MR, Auflösung durch Slice-Arbeit) |

**`MR-004` und `MR-005` sind keine Adaptionen in diesem Mechanismus:** beide
ersetzen keine Baseline-Regel (Pflichtfeld oben). `MR-004`
(WSL-Host-Pfad-Beispiele) steht als Kommentar bei `hostpaths:` in
`.d-check.yml`. `MR-005` (Nicht-Slice-Register: `risks-backlog.md`/
`extra-gates.md` liegen flach in `planning/`, wie der Welle-Plan) entfällt
ersatzlos — die Ablage ist selbsterklärend und kein Gate hängt an ihrer
Begründung.

## Zusatzklassen-Deklaration für Sensors-Bindung

Die vier kanonischen Bindung-Klassen (ADR, Carveout, Schwelle,
Reproduzierbarkeits-Bindung — Letztere nutzt m-trace über immutable
Image-Digests) sind ohne Deklaration legitim. Eine Zusatzklasse darüber
hinaus:

| Klasse | Form | Bedeutung | Beispiel |
|---|---|---|---|
| Requirement-Bindung | `F-*`, `NF-*`, `MVP-*`, `AK-*`, `RAK-*`, `R-*` | Gate prüft eine bestimmte Anforderung/ein Risiko aus dem Lastenheft bzw. Risiko-Register | `RAK-74` für den Per-Projekt-Limiter-Bucket |

## Modus-Deklaration pro Sub-Area

Ohne **Kürzel**-Spalte: m-trace zählt ADR-/Slice-/Welle-Kennungen ohne
Bereichssegment (ein schreibender Mensch + Agent, kein Mehr-Schreiber-Betrieb
— siehe `MR-000`).

| Sub-Area | Modus | Graduierungs-Bedingung |
|---|---|---|
| Spec-Referenz-Richtung | Greenfield | Durchgesetzt von `make docs-check`; keine offenen Reconciliation-Befunde |
| Bestehende akzeptierte ADRs | Brownfield, grandfathered | Historische Dateien bleiben immutable; jede neue ADR folgt der Baseline |
| Commit-Traceability | Greenfield für neue Pull Requests | PR-Bereiche bestehen `make docs-commits`; Vor-Adoptions-Historie bleibt unverändert |
| Requirement-Coverage | Brownfield, observable | Jedes geforderte Requirement hat einen Slice oder kuratierten Coverage-Verweis und `make doc-complete` besteht |
| Requirement-Links | Greenfield | `ids` repo-weit über alle aktiven Doc-Dirs aktiv (welle-01: slice-001 Spec-Straten, slice-002 Rest + R-Familie); verankerte Links gg. inline-`<a id>`-Anker. Durchgesetzt in `make gates`. Exempt: immutable ADRs, `done/`, Root-Übersicht; R-Familie in `spec/**` (matrix-Richtung) |
| Security-Gate-Suppressions (`image-scan`) | Brownfield, observable | Jede Suppression trägt Begründung + `expires` + Scope in `.security/vulnignore.yaml`; Nightly-Audit re-evaluiert; aufgelöst, sobald die `trixie-slim`-Base keine transitiven OS-CVEs ohne Runtime-Pfad mehr trägt (MR-006) |

## Requirement-Coverage-Konvergenz

d-check liest die bestehenden `Kennung`/`Prioritaet`/`Anforderung`- und
`Akzeptanzkriterium`-Tabellen nativ (lebender Pin in `d-check.mk`). `make
doc-trace` ist der Advisory-Sensor.
Die historische `RAK-51`-Redefinition nutzt die explizite
`duplicate-ids: last`-Policy, weil die spätere Zeile ihre Modalität von Kann auf
Muss anhebt.

`make doc-complete` ist nicht in CI gebunden, solange geforderte Requirements
ohne Slice- oder kuratierten Coverage-Verweis verbleiben. Die Graduierung
erfordert das Triagieren dieser Einträge, das Hinzufügen wahrheitsgemäßer
Aufwärts-Coverage-Verweise oder expliziter kuratierter Coverage und das
Erreichen eines bestehenden Gates ohne Schwächung der Modalitäts-Policy.

## Requirement-Link-Konvergenz

Das d-check-`ids`-Modul ist **repo-weit aktiv** (welle-01, `make gates`) über alle
aktiven Doc-Dirs. Requirements leben in Markdown-Tabellenzeilen; damit d-check auf
die *einzelne Definition* auflöst statt auf den Datei-Anfang (link-förmige
Mehrdeutigkeit, die m-trace nicht als Konvergenz akzeptiert), trägt jede
Definitionszeile einen inline-Anker `<a id="<kennung-klein>"></a>` **in der
letzten Zelle** (die Kennungs-Zelle bleibt rein — sonst passt sie nicht
„vollständig" aufs RTM-`id-pattern` und `--trace`/`doc-complete` erkennt 0
Anforderungen), und Mentions sind verankerte Links `[ID](…#slug)`.

**Klarstellung (slice-001):** d-check selbst verlangt den Anker *nicht* — sein
`--repair` erzeugt datei-level Links; der Anker-Zwang ist m-traces Qualitäts-
Policy (Verweis muss auf die Definition zeigen). Der Weg zu den Ankern folgt der
Handbuch-„Brownfield-Migration für tabellarische Lastenhefte" (inline-Anker statt
Heading-Umbau, Tabellenform bleibt).

**Stand der Graduierung — abgeschlossen (welle-01):**

1. ✅ `slice-001` — Spec-Straten (`spec/**`, F/NF/MVP/AK/RAK), 372 Lastenheft-
   Anker, 213 Mentions verankert.
2. ✅ `slice-002` — Rest der aktiven Doku (docs/user, examples, docs/perf,
   Planning-`in-progress/`, carveouts) + `R-`-Familie (31 Anker in
   `risks-backlog.md`). `\b`-Wortgrenze verhindert Über-Matches (`R-` in
   `MR-`/`ADR-`).
3. **Dauerhaft außerhalb** (`ids`-Scope/`exempt-paths`): immutable Accepted-ADRs
   (`docs/plan/adr/**`, MR-002), `done/`, `CHANGELOG`, Root-Übersichts-Docs; die
   `R-`-Familie zusätzlich in `spec/**` (der Vertrag verweist nicht abwärts aufs
   Risiko-Register — `matrix`-Richtung).
