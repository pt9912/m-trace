# Slice 035: commit-msg-Muster an MR-003-Familien (R-32)

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle — Closure-Bedingung
ist die DoD dieses Slices.

**Bezug:** [R-32](../risks-backlog.md#r-32) ·
[MR-003](../../../../harness/conventions.md#mr-003) · Durchsetzungsschicht-
Adoption `8856c40` · Review
[2026-09-29-slice-034.md, Finding 2](../../../reviews/2026-09-29-slice-034.md).

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
  [R-32](../risks-backlog.md#r-32) „vor Aktivierung" setzt.
- **d-check-`ids`-Konfiguration** — sie prüft bereits die Familien nach
  [MR-003](../../../../harness/conventions.md#mr-003); nur die Commands-Texte
  nennen noch die Emission-Menge.
- **Stop-Hook/Command-Guard/Gate-Nachweis** — weiter offen, eigenes Design
  (slice-034 §1); berührt dieser Slice nicht.

## 2. Definition of Done

- [x] Muster in `commit-msg-traceability.sh` erkennt alle Familien nach
      [MR-003](../../../../harness/conventions.md#mr-003), `LH-*` geht nicht
      mehr durch; Wortgrenzen verhindern Substring-Über-Matches.
- [x] `selbstpruefung.sh`-Marker (MSG_ROT/MSG_GRUEN) tragen m-trace-konforme
      Messages; die Selbstprüfung läuft gegen einen Wegwerf-Klon grün
      (rot fällt, grün geht durch).
- [x] Commands nennen die korrekte Kennungs-Menge (kein `LH-` mehr);
      [R-32](../risks-backlog.md#r-32) im Register aufgelöst (§1.2, mit
      Auflösungs-Befund).
- [x] `make docs-check` grün.
- [x] Review durchgeführt, Report unter `docs/reviews/` (frischer Kontext).
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag; §6-Risiken mit Ausgang.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tools/harness/commit-msg-traceability.sh` | update | Muster (Z. 60) + Kopf-Kommentar (Z. 7) auf die Familien nach [MR-003](../../../../harness/conventions.md#mr-003), `LH-` streichen |
| `tools/harness/selbstpruefung.sh` | update | MSG_ROT/MSG_GRUEN-Defaults auf m-trace-konforme Beispiele |
| `.claude/commands/implement-slice.md` | update | Commit-Kennung (Z. 43), Berichten (Z. 70), Herkunfts-Formen (Z. 147), Doc-Gate (Z. 33) |
| `.claude/commands/plan-welle.md`, `close-welle.md` | update | Doc-Gate-Bullets (LH- entfällt aus der Kennungs-Liste) |
| `docs/plan/planning/risks-backlog.md` | update | [R-32](../risks-backlog.md#r-32) nach §1.2 (aufgelöst), Anker bleibt |

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
  **Ausgang:** belegt — Negativ-Proben (`FOR-1234`, `XF-1`, `LH-FA-01`)
  fallen, Familien-Positivproben gehen durch (Review- und Verifier-Batterie,
  [Verifier-Report](../../../reviews/2026-10-22-slice-035-verifier.md)). Die
  Proben sind nicht committet: siehe Finding 3 unten.
- **Selbstprüfung kloniert das Repo** (Wegwerf-Verzeichnis, zwei Commit-
  Versuche) — läuft netzlos, verlangt aber den sauberen Arbeitsbaum.
  **Ausgang:** belegt — Verifier-Neulauf im Wegwerf-Klon: rot fällt (Exit 1,
  HEAD unverändert), grün geht durch (Gate-Override `make docs-check`, siehe
  Review-Finding 6).
- **Review-Finding 3 (LOW): `\b`-Grenzproben nicht committet** — die
  Selbstprüfung deckt sie nicht ab (grüne Message matcht auch ohne
  Wortgrenze, rote trägt keine Kennung). **Ausgang:** dokumentierte
  Entscheidung, kein neuer R-Eintrag — der Wächter ist unaktiviert (Opt-in),
  die Aussage wurde von Review und Verifier mit je eigener Batterie
  unabhängig reproduziert. Wiederaufnahme-Trigger: vor `make hooks-install`
  oder bei der nächsten Änderung an der `patterns=`-Zeile — dann die
  Proben-Batterie (Familien, `LH-FA-01`, `FOR-1234`, `XF-1`) in
  `tools/harness/selbstpruefung.sh` aufnehmen.
- **Review-Finding 5 (INFO): Ziffernlängen-Präfix-Match** (`MR-0003`, `ADR-00011`,
  `slice-035x` gehen durch). **Ausgang:** akzeptiert — Erbe der Emission,
  gedeckt von der Wächter-GRENZE „Anwesenheit einer Kennung, nicht ihre
  Wahrheit"; keine Aktion.
- **Roadmap-Aktiv-Zeile (`bfefb6d`) war in §3/DoD nicht deklariert**
  (Review-Finding 1, 2. Auftreten nach slice-034). **Ausgang:** eingetreten, mit
  der Closure zum Ruhe-Marker zurückgeführt.

## 7. Closure-Notiz

**Datum:** 2026-10-22. **Belege:** Review
[2026-09-29-slice-035.md](../../../reviews/2026-09-29-slice-035.md) (0 HIGH /
0 MEDIUM / 3 LOW / 3 INFO, nicht merge-blockierend; Finding 2 in `bbed415`
behoben); Verifier
[2026-10-22-slice-035-verifier.md](../../../reviews/2026-10-22-slice-035-verifier.md)
(bestanden, bedingt — Bedingung: diese Closure-Arbeit; eigene Läufe:
`make docs-check` 0 Befunde, `make lint-variante-b` grün, 15 Wächter-Proben,
Selbstprüfung im Wegwerf-Klon grün). `make verify-closure-notes` und
`make verify-plan-paths` greifen erst nach dem Move bzw. dem Closure-Edit;
sie sind vor dem Move nicht aussagekräftig.

- **Geliefert:** Wächter-Muster auf die Familien nach
  [MR-003](../../../../harness/conventions.md#mr-003) (`LH-*` fällt,
  `\b`-Wortgrenzen), `MSG_GRUEN` m-trace-konform, Commands ohne `LH-`,
  [R-32](../risks-backlog.md#r-32) in §1.2 aufgelöst. Aktivierung bleibt
  Owner-Entscheid.
- **Validator:** n/a — interne Wartung, kein End-Nutzer-Wert;
  ausdrücklich übersprungen statt still.
- **Risiken:** [R-32](../risks-backlog.md#r-32) aufgelöst; Finding 3 als
  dokumentierte Entscheidung mit Wiederaufnahme-Trigger (§6), Finding 5 akzeptiert.
  Kein neuer R-Eintrag.
- **Folge-Slice:** keiner geschnitten; die Proben-Batterie in der
  Selbstprüfung ist an den Trigger „vor `make hooks-install`" gebunden
  (§6, Finding 3). Stop-Hook/Command-Guard/Gate-Nachweis bleiben offen (§1).
- **Architektur-Beobachtung:** keine — `spec/architecture.md` unberührt.
  Rest-Nennung `LH-` in `.claude/agents/architect.md` liegt außerhalb des
  Scopes (Verifier, DoD 3) und bleibt unbearbeitet.
- **Beobachtungs-Register:** keine Beobachtung angefallen.
- **Plan-vs-Diff:** §3 stimmt mit dem realen Diff überein (alle sechs Pfade
  berührt); Mehr: `roadmap.md` (Aktiv-Zeile, nun zurückgeführt) und der
  Review-Report.
- **Steering-Loop-Lerneintrag (geschärfte Regel, `· seit slice-035`):**
  Eine Muster-Zusage (Wortgrenze, Familienmenge) braucht ihre Grenzproben
  als committetes Artefakt, nicht nur als Zählung in Commit-Message oder
  Register-Zelle — sonst ist der Beleg nach dem Lauf nicht
  rekonstruierbar (Review-Finding 3; Verifier musste neu proben). Zweiter Teil:
  Setzt ein Slice eine Roadmap-Aktiv-Zeile, nennt der Plan die
  Rückführung beim Closure als Nachführung (Review-Finding 1, 2. Auftreten der
  Klasse „Zustandsfeld ohne Closure-Nachführung" nach slice-034).

## 8. Sub-Area-Prüfungen und Modus-Begründung

**Vorgelagert — Sub-Area-Wahl:** Commit-Traceability (`COMMIT`, conventions
§Modus-Deklaration) berührt mit dem Wächter genau eine Sub-Area; Schwelle
(Kennungs-Zusage in AGENTS.md §5.1 + [MR-003](../../../../harness/conventions.md#mr-003)
+ [R-32](../risks-backlog.md#r-32)) erfüllt.

**Vorgelagert — offene Beobachtungen sichten:** slice-034-Closure meldete
„keine Beobachtung angefallen"; Register trägt keine BEO-Klasse zur
Commit-Traceability.

**Modus-Begründung:** Werkzeug-Anpassung am adoptierten Emission-Stand
(Brownfield): die Form „Emission byte-identisch" gilt hier nicht — der
Wächter wurde in `8856c40` ohne Byte-Identitäts-DoD adoptiert, und
[R-32](../risks-backlog.md#r-32) dokumentiert die Anpassung als
[MR-003](../../../../harness/conventions.md#mr-003)-Konvergenz. Keine
Gate-Senkung — der Wächter schärft (Fremdfamilie `LH-*` fällt durch). Kein ADR.
