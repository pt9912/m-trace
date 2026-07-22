# Harness-Konventionen

## Zweck

Diese Datei deklariert m-trace-spezifische Strukturregeln und Adaptionen
gegenüber der adoptierten Harness-Baseline. Sie supplementiert die Baseline,
ohne sie zu kopieren.

## Baseline

m-trace adoptiert die ai-harness-course-**v3.5.0**-Baseline (2026-07-19),
vendored netzlos unter `.harness/baseline/v3.5.0/regelwerk/` (17 Module +
3 Grundlagen) und `.harness/baseline/v3.5.0/templates/` (die Referenz-Ziel-
Formen). Das Release-Archiv `lab-regelwerk.zip` trägt sha256
`123e3383261102e6be6465e1f4bade08a474c00edc4fff89f5c4b11bd640f8ff`; die
Per-Datei-Integrität ist in `.harness/baseline/v3.5.0/SHA256SUMS` gepinnt und
wird mit `sha256sum -c` verifiziert. Vendored 2026-07-22 gemäß
[ADR-0009](../docs/adr/0009-harness-baseline-v3.5.0.md).

Dies löst den vorherigen Commit-Pin ab (nur `grundlagen-konventionen.md`, bei
ai-harness-course `d2f60da`, abgerufen 2026-07-14): das gesamte Regelwerk und
der vollständige Template-Satz sind jetzt präsent und integritäts-geprüft,
sodass die `../templates/…`-Referenz-Formen lokal auflösen, statt gegen einen
sich bewegenden Kurs-Head zu driften.

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

## Adaptionen

### MR-001 - Repository-Pfade

- **Datum:** 2026-07-14
- **Scope:** ADR- und Planning-Verzeichnisse
- **Baseline-Unterschied:** m-trace nutzt `docs/adr/` statt `docs/plan/adr/`
  und `docs/planning/` statt `docs/plan/planning/`.
- **Grund:** Etabliertes öffentliches Repository-Layout mit umfangreichen
  stabilen Links.
- **Auflösungs-Trigger:** Permanent, sofern nicht eine separat reviewte
  Repository-Layout-Migration freigegeben wird.

### MR-002 - Accepted-ADR-Grandfathering

- **Datum:** 2026-07-14
- **Scope:** `docs/adr/0001-*.md` bis `docs/adr/0007-*.md`
- **Baseline-Unterschied:** Diese akzeptierten Brownfield-Records enthalten
  historische Plan-Provenienz außerhalb eines ausgewiesenen History-Abschnitts.
- **Grund:** Akzeptierte ADRs sind unter der adoptierten Baseline immutable und
  werden nicht allein zum Nachrüsten der Konvention umgeschrieben.
- **Auflösungs-Trigger:** Permanente historische Ausnahme. Neue ADRs erhalten
  keine Ausnahme; künftige Accepted-ADR-Änderungen werden vom
  ADR-Immutabilitäts-Sensor blockiert.

### MR-003 - Requirement-ID-Familien

- **Datum:** 2026-07-14
- **Scope:** Contract, Pläne, Commits und Reviews
- **Baseline-Unterschied:** m-trace datiert vor der `LH-*`-Beispielfamilie und
  nutzt `F-*`, `NF-*`, `MVP-*`, `AK-*`, `RAK-*` und `R-*`.
- **Grund:** Die Kennungen sind Teil des etablierten Contracts und der
  Release-Historie.
- **Auflösungs-Trigger:** Permanent. Neue Requirement-Familien müssen hier vor
  Nutzung deklariert werden.

### MR-004 - WSL-Host-Pfad-Beispiele

- **Datum:** 2026-07-14
- **Scope:** Drei WSL-Troubleshooting-Beispiele in
  `docs/user/local-development.md`
- **Baseline-Unterschied:** Der `hostpaths`-Sensor lehnt host-lokale absolute
  Pfade normalerweise ab.
- **Grund:** Diese Pfade sind der Gegenstand der Operator-Guidance und können
  nicht durch repository-relative Pfade ersetzt werden, ohne die Diagnose zu
  verlieren.
- **Auflösungs-Trigger:** Permanent, solange WSL2 unterstützt wird. Der
  `hostpaths`-Sensor gated daher `/Users` und `/Development`; `/mnt` und
  `/home` liegen bewusst außerhalb seines konfigurierten Präfix-Satzes.

## Sensor-Bindungsklassen

m-trace nutzt derzeit Requirement-Bindung (`F-*`, `NF-*`, `MVP-*`, `AK-*`,
`RAK-*`, `R-*`), ADR-Bindung (`ADR-NNNN`) und Reproduzierbarkeits-Bindung über
immutable Image-Digests.

## Modi

| Sub-Area | Modus | Graduierungs-Bedingung |
|---|---|---|
| Spec-Referenz-Richtung | Greenfield | Durchgesetzt von `make docs-check`; keine offenen Reconciliation-Befunde |
| Bestehende akzeptierte ADRs | Brownfield, grandfathered | Historische Dateien bleiben immutable; jede neue ADR folgt der Baseline |
| Commit-Traceability | Greenfield für neue Pull Requests | PR-Bereiche bestehen `make docs-commits`; Vor-Adoptions-Historie bleibt unverändert |
| Requirement-Coverage | Brownfield, observable | Jedes geforderte Requirement hat einen Slice oder kuratierten Coverage-Verweis und `make doc-complete` besteht |
| Requirement-Links | Brownfield | Stabile Per-Requirement-Anker existieren und der ID-Link-Gate kann ohne mehrdeutige Root-Links aktiviert werden |

## Requirement-Coverage-Konvergenz

d-check v0.43.0 liest die bestehenden `Kennung`/`Prioritaet`/`Anforderung`- und
`Akzeptanzkriterium`-Tabellen nativ. `make doc-trace` ist der Advisory-Sensor.
Die historische `RAK-51`-Redefinition nutzt die explizite
`duplicate-ids: last`-Policy, weil die spätere Zeile ihre Modalität von Kann auf
Muss anhebt.

`make doc-complete` ist nicht in CI gebunden, solange geforderte Requirements
ohne Slice- oder kuratierten Coverage-Verweis verbleiben. Die Graduierung
erfordert das Triagieren dieser Einträge, das Hinzufügen wahrheitsgemäßer
Aufwärts-Coverage-Verweise oder expliziter kuratierter Coverage und das
Erreichen eines bestehenden Gates ohne Schwächung der Modalitäts-Policy.

## Requirement-Link-Konvergenz

Das d-check-`ids`-Modul ist noch nicht aktiviert. Requirements leben derzeit in
Markdown-Tabellenzeilen, sodass eine automatische Reparatur 340 nackte
Kennungen in Links auf den Anfang von `spec/lastenheft.md` verwandeln würde,
nicht auf die einzelne Definition. Das erzeugt link-förmige Mehrdeutigkeit und
wird nicht als Konvergenz akzeptiert.

Die Graduierung erfordert diese Schritte:

1. Jedem normativen Requirement einen stabilen, direkt adressierbaren
   Definitions-Anker geben, ohne seine Kennung oder vertragliche Formulierung zu
   ändern.
2. Verifizieren, dass d-check jede konfigurierte ID-Familie auf diese Definition
   auflöst, nicht bloß auf die enthaltende Datei.
3. Aufwärts-Verweise in aktiver technischer, ADR-, Planning- und Anwender-
   Dokumentation migrieren; historische Release-Records bleiben unverändert, wo
   gefordert.
4. `ids` erst aktivieren, nachdem sein Advisory-Lauf kein mehrdeutiges Ziel
   meldet und die resultierenden Links `links` und `anchors` bestehen.
