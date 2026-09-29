# Slice 035: commit-msg-Muster an MR-003-Familien (R-32)

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle — Closure-Bedingung
ist die DoD dieses Slices.

**Bezug:** [R-32](../../risks-backlog.md#r-32) ·
[MR-003](../../../../harness/conventions.md#mr-003) · Durchsetzungsschicht-
Adoption `8856c40` · Review
[2026-09-29-slice-034.md, F-2](../../../reviews/2026-09-29-slice-034.md).

**Autor:** Owner-Auftrag „weitermachen" nach slice-034-Closure. **Datum:** 2026-09-29.

---

## 1. Ziel und Abgrenzung

**Ziel:** Der commit-msg-Wächter (`tools/harness/commit-msg-traceability.sh`)
prüft die Kennungs-Menge nach
[MR-003](../../../../harness/conventions.md#mr-003) — die m-trace-Familien
`F-*`, `NF-*`, `MVP-*`, `AK-*`, `RAK-*`, `R-*` neben
`ADR-NNNN`/`MR-NNN`/`slice-<NNN>`; `LH-*` (aih-Fremdfamilie) geht nicht mehr
durch. Die Kennungs-Menge in den Commands (`.claude/commands/*.md`) folgt
demselben Abgleich.

**Ausdrücklich NICHT in diesem Slice:**

- **Aktivierung** (`make hooks-install`) — bleibt Owner-Entscheid; dieser
  Slice räumt nur die Bedingung weg, die
  [R-32](../../risks-backlog.md#r-32) „vor Aktivierung" setzt.
- **d-check-`ids`-Konfiguration** — sie prüft bereits die Familien nach
  [MR-003](../../../../harness/conventions.md#mr-003); nur die Commands-Texte
  nennen noch die Emission-Menge.
- **Stop-Hook/Command-Guard/Gate-Nachweis** — weiter offen, eigenes Design
  (slice-034 §1); berührt dieser Slice nicht.

## 2. Definition of Done

- [ ] Muster in `commit-msg-traceability.sh` erkennt alle Familien nach
      [MR-003](../../../../harness/conventions.md#mr-003), `LH-*` geht nicht
      mehr durch; Wortgrenzen verhindern Substring-Über-Matches.
- [ ] `selbstpruefung.sh`-Marker (MSG_ROT/MSG_GRUEN) tragen m-trace-konforme
      Messages; die Selbstprüfung läuft gegen einen Wegwerf-Klon grün
      (rot fällt, grün geht durch).
- [ ] Commands nennen die korrekte Kennungs-Menge (kein `LH-` mehr);
      [R-32](../../risks-backlog.md#r-32) im Register aufgelöst (§1.2, mit
      Auflösungs-Befund).
- [ ] `make docs-check` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` (frischer Kontext).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag; §6-Risiken mit Ausgang.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tools/harness/commit-msg-traceability.sh` | update | Muster (Z. 60) + Kopf-Kommentar (Z. 7) auf die Familien nach [MR-003](../../../../harness/conventions.md#mr-003), `LH-` streichen |
| `tools/harness/selbstpruefung.sh` | update | MSG_ROT/MSG_GRUEN-Defaults auf m-trace-konforme Beispiele |
| `.claude/commands/implement-slice.md` | update | Commit-Kennung (Z. 43), Berichten (Z. 70), Herkunfts-Formen (Z. 147), Doc-Gate (Z. 33) |
| `.claude/commands/plan-welle.md`, `close-welle.md` | update | Doc-Gate-Bullets (LH- entfällt aus der Kennungs-Liste) |
| `docs/plan/planning/risks-backlog.md` | update | [R-32](../../risks-backlog.md#r-32) nach §1.2 (aufgelöst), Anker bleibt |

## 4. Trigger

- **Start (`next` → `in-progress`):** sofort — Owner-Auftrag „weitermachen"
  nach slice-034-Closure.
- **Rückführungen:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + Selbstprüfungs-Lauf grün (Wegwerf-Klon) + Closure-Notiz +
`git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Substring-Über-Match** der neuen Familien: `R-[0-9]+` matcht ohne
  Wortgrenze in `ADR-0001` (harmlos — ADR ist valide Familie), aber auch in
  Fremd-Substrings wie `FOR-1234`; `NF-` enthält `F-`. Gegenmaßnahme:
  `\b`-Anker an den Alternations-Anfang und Reihenfolge (NF- vor F-).
  **Ausgang:** wird im Lauf belegt (Negativ-Proben).
- **Selbstprüfung kloniert das Repo** (Wegwerf-Verzeichnis, zwei Commit-
  Versuche) — läuft netzlos, verlangt aber den sauberen Arbeitsbaum.
  **Ausgang:** wird im Lauf belegt.

## 7. Closure-Notiz

*(füllt bei Closure)*

## 8. Sub-Area-Prüfungen und Modus-Begründung

**Vorgelagert — Sub-Area-Wahl:** Commit-Traceability (`COMMIT`, conventions
§Modus-Deklaration) berührt mit dem Wächter genau eine Sub-Area; Schwelle
(Kennungs-Zusage in AGENTS.md §5.1 + [MR-003](../../../../harness/conventions.md#mr-003)
+ [R-32](../../risks-backlog.md#r-32)) erfüllt.

**Vorgelagert — offene Beobachtungen sichten:** slice-034-Closure meldete
„keine Beobachtung angefallen"; Register trägt keine BEO-Klasse zur
Commit-Traceability.

**Modus-Begründung:** Werkzeug-Anpassung am adoptierten Emission-Stand
(Brownfield): die Form „Emission byte-identisch" gilt hier nicht — der
Wächter wurde in `8856c40` ohne Byte-Identitäts-DoD adoptiert, und
[R-32](../../risks-backlog.md#r-32) dokumentiert die Anpassung als
[MR-003](../../../../harness/conventions.md#mr-003)-Konvergenz. Keine
Gate-Senkung — der Wächter schärft (Fremdfamilie `LH-*` fällt durch). Kein ADR.
