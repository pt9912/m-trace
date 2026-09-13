# MR-006 — Security-Gate-Carveout-Registry

- **Datum:** 2026-07-23
- **Geltungsbereich:** `image-scan`/`vuln-check`-Gate; OS-CVE-Ausnahmen der
  `node:22-trixie-slim`-Base (`mtrace-dashboard`, `mtrace-analyzer-service`),
  geführt in `.security/vulnignore.yaml`
- **Ersetzt-Baseline-Regel:** [`modul-07-carveouts.md` §Ziel-Form:
  Carveout](../../.harness/baseline/v6.8.0/regelwerk/modul-07-carveouts.md#ziel-form-carveout)
  — dort ist die Dateikonvention **ein** `docs/plan/carveouts/CO-<NNN>-*.md`
  je einzelner, temporärer Gate-Senkung.
- **Adaption:** m-trace senkt den Security-Gate für einen **Cluster**
  transitiver OS-CVEs (kein Runtime-Pfad, oft ohne Upstream-Fix) über die
  domänenspezifische Registry `.security/vulnignore.yaml` (per-CVE `reason`
  + `expires` + `scope`, deterministisch nach `.trivyignore` gerendert,
  Nightly-Audit-Re-Eval) statt über ein `CO-<NNN>`-File je CVE.
- **Begründung:** Modul-7-Werkzeug-Wahl Frage 1 (Granularität): ein Cluster
  im selben Geltungsbereich ist eine BF-Sub-Area-Markierung, **keine**
  Carveout-Kaskade (ein CO-File je CVE ist der explizit gewarnte
  Anti-Pattern). Die vorhandene Registry ist reicher als das generische
  CO-Template und die Single Source of Truth; ein CO-File je CVE würde sie
  duplizieren.
- **Auflösungs-Trigger:** Je CVE der eigene `expires`/Upstream-Fix-Trigger in
  `vulnignore.yaml`; als Sub-Area permanent, solange die `trixie-slim`-Base
  transitive OS-CVEs ohne Runtime-Exponierung trägt. Das generische
  `docs/plan/carveouts/` bleibt für künftige einzelne, nicht-Security
  Gate-Senkungen reserviert.
