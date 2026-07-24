# 0011 — Regelwerk-Baseline-Bump v3.5.0 → v3.5.1 (nicht-struktureller Re-Vendor)

> **Status**: **Accepted** (2026-07-24)
> **Datum**: 2026-07-24 (Proposed), 2026-07-24 (Accepted)
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)
> **Bezug**: [ADR-0009](0009-harness-baseline-v3.5.0.md) (strukturelle
> v3.5.0-Adoption; dessen Re-Evaluierungs-Trigger „Baseline-Re-Vendoring als
> eigener Review, kein stiller Auto-Bump" ist die Grundlage dieses ADR);
> [`harness/conventions.md`](../../harness/conventions.md) §Baseline. Prozess-/
> Harness-ADR ohne Spec-Stratum-Schärfung.

## Kontext

ai-harness-course hat **v3.5.1** veröffentlicht (Kurs-Welle 33, 2026-07-23;
Asset `lab-regelwerk.zip`, sha256
`7268a8e6f36476c98d5cf0547d16deacec70fcddcf23df38f87d029e967cb10d`). ADR-0009
pinnt die Baseline explizit auf **v3.5.0** (Archiv-sha256 `123e3383…`) und ist
nach `Accepted` immutable; der Pin lässt sich also nicht durch Edit von ADR-0009
fortschreiben. ADR-0009 nennt genau diesen Fall als Re-Evaluierungs-Trigger:
*„Ein Kurs-Release nach v3.5.0 mit strukturellen Template-/Layout-Änderungen
(dann Baseline-Re-Vendoring als eigener Review, kein stiller Auto-Bump)."*

**Delta-Assessment v3.5.0 → v3.5.1** (verifiziert per Zip-Download + File-Diff
gegen die vendored v3.5.0-Baseline):

1. **Mechanisch, alle Files:** der Bundle-Export stempelt in jeder Datei die
   Quell-URLs `…/blob/v3.5.0/…` → `…/blob/v3.5.1/…`; `regelwerk/README.md`
   hebt `Stand: Welle 32 · 2026-07-19` → `Welle 33 · 2026-07-23`. Keine
   inhaltliche Änderung, kein File hinzugefügt/entfernt (gleiche 43 Dateien).
2. **Substantiell, genau ein File:**
   `templates/docs/plan/planning/README.template.md` korrigiert einen
   Selbstwiderspruch im Block „Slices vs. Wellen". Alt: *„die
   Lifecycle-Verzeichnisse sind slice-reserviert"* (alle vier), während der
   Text zugleich `welle-<id>-results.md` in `done/` ablegte. Neu:
   `open/` → `next/` → `in-progress/` nehmen **ausschließlich Slices** auf;
   `done/` archiviert **zusätzlich Nicht-Slice-Records** (Welle-Closure
   `done/<welle-id>-results.md` **und** aufgelöste Carveouts, Modul 7).

**Einordnung:** Die Korrektur ist eine **Formulierungs-Klarstellung in einem
Template**, keine Layout-/Pfad-Änderung. Der strukturelle Trigger von ADR-0009
(„strukturelle Template-/Layout-Änderungen") ist damit **nicht** ausgelöst; die
allgemeine Doktrin „kein stiller Auto-Bump" gilt aber weiter, daher dieser
eigene Review. Für m-trace ist die Klarstellung **bestätigend**: das Repo hält
`welle-01-results.md` und die grandfatherten `plan-*.md` bereits in `done/` und
führt Nicht-Slice-Register flach in `planning/` (MR-005, slice-006) — v3.5.1
macht genau diese Ablage explizit kanon-konform.

## Entscheidung

> **Entscheidung (Proposed 2026-07-24):** Die aktive Regelwerk-Baseline wird auf
> **v3.5.1** fortgeschrieben. Der Bump ist als **nicht-struktureller Re-Vendor**
> klassifiziert — kein weiterer Layout-/Pfad-Umbau nötig (im Gegensatz zur
> v3.5.0-Adoption, ADR-0009).

Bestandteile:

1. **v3.5.1 vendoren** nach `.harness/baseline/v3.5.1/{regelwerk,templates}/` +
   erzeugtem `SHA256SUMS` (43 Dateien, `sha256sum -c` grün), netzlos committet.
2. **v3.5.0 behalten** (Owner-Entscheidung 2026-07-24): die vendored
   v3.5.0-Baseline bleibt unter `.harness/baseline/v3.5.0/` liegen, damit die
   `.harness/baseline/v3.5.0/…`-Referenzen der historischen `done/`-Records und
   der immutablen ADRs netzlos auflösbar bleiben. Es gibt damit **zwei**
   vendored Baselines; die **aktive** ist v3.5.1.
3. **Lebende Zeiger umhängen:** `harness/conventions.md` §Baseline (Version +
   Archiv-sha + Datum), `AGENTS.md`, `CLAUDE.md`, `docs/plan/carveouts/README.md`,
   `docs/reviews/README.md`, `.harness/skills/reviewer.md` zeigen künftig auf
   `.harness/baseline/v3.5.1/…`. Historische `done/`-Records und die immutablen
   ADRs bleiben unverändert (ihre v3.5.0-Verweise sind Audit-Bestand).
4. **Ausführung als `slice-008`** (Re-Vendor + Repath); dieses ADR ist die
   Aufwärts-Referenz (SDP: Plan → ADR).

## Konsequenzen

**Positiv:**

- Aktiver Pin folgt dem Kurs-Stand (Welle 33); die vendored „Ziel-Form"-Templates
  tragen die korrigierte `done/`-Klarstellung.
- Kein Layout-Churn: da v3.5.1 keine Pfade ändert und v3.5.0 erhalten bleibt,
  brechen keine historischen Links; die Migration ist ein reiner Additions-/
  Repath-Schritt statt eines Struktur-Umbaus.
- Die Template-Klarstellung stützt MR-005/slice-006 (flach-in-`planning/` +
  `done/`-Records) — die frühere „slice-reserviert"-Spannung ist im Kanon aufgelöst.

**Kosten / Grenzen (ehrlich benannt):**

- **Zwei vendored Baselines** = ~124 KB Redundanz. Bewusst akzeptiert
  (Owner 2026-07-24) zugunsten netzloser Auflösung der historischen Referenzform.
  Folge-Trigger: bei einem späteren Bump entscheiden, ob v3.5.0 dann entfällt.
- **Geteilter Pin-Zustand.** Solange beide liegen, ist „welche Baseline gilt"
  nur aus `conventions.md` §Baseline + diesem ADR ersichtlich, nicht aus dem
  bloßen Verzeichnis-Bestand. §Baseline benennt die aktive Version explizit.

## Alternativen

- **A — v3.5.0 ersetzen (nur eine Baseline).** Sauberer, aber die
  `.harness/baseline/v3.5.0/…`-Links der `done/`-Records und immutablen ADRs
  würden brechen (bzw. müssten via `ignore-refs`-Tombstone grandfathert werden).
  **Verworfen** (Owner 2026-07-24): Retention ist billiger als Tombstones.
- **B — Slice ohne ADR (Delta nur in der Closure-Notiz).** Leichter, da
  nicht-strukturell und von der ADR-0009-Doktrin gedeckt. **Verworfen**
  (Owner 2026-07-24): ADR-0009 ist immutable und nennt die exakte Version + sha;
  ein sichtbarer ADR-Trail für den Pin-Wechsel ist die auditierbarere Form.
- **C — Bump vertagen.** Zulässig (kein Blocker, kein erzwungener Repo-Edit).
  **Verworfen** (Owner 2026-07-24): das Delta ist klein und die Klarstellung
  bestätigt den Bestand; kein Grund zu warten.

## Re-Evaluierungs-Trigger

- Ein Kurs-Release nach v3.5.1 mit **strukturellen** Template-/Layout-Änderungen
  (dann erneut Re-Vendoring als eigener Review, wie ADR-0009 es etabliert hat).
- Beim nächsten Bump: entscheiden, ob die dann zwei Versionen alte v3.5.0-Baseline
  entfernt werden kann (sobald keine lebenden `done/`-Referenzen mehr darauf zeigen
  bzw. via Tombstone abgelöst).

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-07-24 | Proposed | ADR-0011 |
| 2026-07-24 | Accepted (Owner-Freigabe) | ADR-0011 |
