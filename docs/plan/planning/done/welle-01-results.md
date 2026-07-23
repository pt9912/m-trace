# Welle 01 — Ergebnis: Requirement-Link-Konvergenz

**Abschluss:** 2026-07-23. **Welle-Plan:** [`welle-01-requirement-link-konvergenz.md`](welle-01-requirement-link-konvergenz.md).

## Geliefert

`ids` ist **repo-weit** scharf: jede nackte Requirement-Kennung in aktiver Doku
zeigt als verankerter Link auf ihre Definition.

- **[`slice-001`](slice-001-requirement-anker-ids.md):** 372 Lastenheft-Anker,
  213 Links in `spec/**`, `ids` auf den Spec-Straten.
- **[`slice-002`](slice-002-requirement-links-restdocs.md):** 31 R-Anker im
  Risiko-Register, 173 Links über 21 Dateien, `ids` repo-weit + R-Familie.
- `conventions.md` §Requirement-Link-Konvergenz voll graduiert (Modus-Zeile
  „Requirement-Links" → Greenfield). Erste kanonische Welle/Slices nach der
  v3.5.0-Migration (MR-007).

## Verifikation (Closure-Trigger)

- Beide Slices in `done/`. `make gates` grün (`ids`/`anchors`/`links`/`matrix`).
- `make doc-trace` ok (372 Requirements erkannt); die kombinierte
  `trace.slices.file-pattern` (`plan-`/`slice-`/`welle-`) erkennt jetzt beide
  Formen (plan-Erkennung byte-identisch verifiziert).

## Steering-Loop-Einträge (was gelernt wurde)

1. **Advisory-Tools separat prüfen.** Die Anker zuerst in der Kennungs-Zelle
   brachen `--trace`/`doc-complete` (0 Anforderungen erkannt) — `make gates` sah
   es **nicht**, weil die RTM advisory ist. Nach Doku-Struktur-Edits die
   Nicht-Gate-Sensoren separat laufen lassen. → Sensor-Guide ergänzt in
   `conventions.md` §Requirement-Link-Konvergenz.
2. **Erst Handbuch + empirischer Lauf, dann Urteil.** „Vakuum/geht nicht/passt
   nicht" fiel mehrfach bei einem `--repair`/`--trace`-Lauf oder Handbuch-Blick.
   Bei konfigurierbaren Werkzeugen (d-check-Modul) immer erst verifizieren.
3. **ID-Präfix-Kollisionen & Referenz-Richtung.** Neue ID-Familien mit `\b`
   gegen Präfix-Über-Matches absichern (`R-` in `MR-`/`ADR-`); die
   `matrix`-Richtung (kein Vertrag → Planning) vor dem Verlinken prüfen.

## Folge-Slices

Keine offenen. Die Sub-Area „Requirement-Links" ist Greenfield; künftige
Kennungs-Mentions sind per `ids` in `make gates` linkpflichtig.
