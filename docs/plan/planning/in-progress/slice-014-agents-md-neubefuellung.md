# Slice 014: `AGENTS.md` gegen das v6.8.0-Template neu befüllen

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](../welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 3 von 10).

**Bezug:** `AGENTS.md`,
`.harness/baseline/v6.8.0/templates/AGENTS.template.md`, Regelwerk-Module 2
(Bootstrap), 3 (Spec), 8 (Agentenrollen), 9 (Implementierung), 13
(Quality-Gates).

**Autor:** Nachgang Templates-Diff v3.5.1→v6.8.0. **Datum:** 2026-09-13.

---

## 1. Ziel

`AGENTS.md` komplett aus dem `v6.8.0`-Template neu befüllen (kein
Absatz-für-Absatz-Patch — laut Templates-Fork zu viele Querbezüge zwischen
den geänderten Abschnitten). Direkter Diff `AGENTS.template.md` v3.5.1 gegen
v6.8.0 gezogen (nicht nur Fork-Zusammenfassung übernommen); die Deltas unten
sind entlang dieses Diffs konkret, nicht nacherzählt.

## 2. Definition of Done

- [ ] **Source-Precedence-Liste auf 9 Ränge** (statt 8): neuer Rang 6
      `docs/user/*` (optional — falls im Repo nicht vorhanden, entlinkt mit
      `<!-- d-check:ignore -->`, m-trace **hat** `docs/user/`, also real
      verlinken). Rang 5 (`roadmap.md`) heißt jetzt „Wellen-Sequenz" statt
      „aktuelle Welle" — **nur** nachziehen, wenn Tranche 8
      (Roadmap-Terminologie) parallel oder vorher entschieden ist, sonst
      Terminologie-Bruch zwischen AGENTS.md und roadmap.md.
- [ ] **Rang 2 (`spec/spezifikation.md`)-Kommentar „optionales 3. Spec-Stratum"
      entfernt** — die drei Straten sind im neuen Kanon **Default**, nicht
      Zusatz (Kehrtwende gegenüber v3.5.1: **2-Straten wäre jetzt die
      MR-pflichtige Abweichung**, nicht 3-Straten). m-trace führt bereits
      drei Straten (Contract/Technical/View) — hier **kein** Widerspruch,
      nur die MR-001-Referenz im alten Kommentar entfällt (m-trace hat
      ohnehin nie `MR-001` für diesen Zweck genutzt, das ist bei uns eine
      andere Nummer/Sache — siehe `harness/conventions.md`).
- [ ] **§3.3 (git-mv-Regel) bekommt zweite Variante:** Regelfall bleibt
      `git mv` zuerst, Inhalt danach — **außer** beim Übergang nach `done/`:
      dort **zuerst** der Inhalt (DoD-Häkchen, Closure-Notiz), **dann** der
      reine `git mv` (die Notiz ist Bedingung für `done/`, nicht seine
      Folge). **m-trace hat das in `slice-010`/`011`/`012` bereits genau so
      gemacht** (Closure-Notiz geschrieben, dann `git mv` nach `done/`) —
      dieser Punkt kodifiziert nur nachträglich gelebte Praxis.
- [ ] **§3.4 (Architektur-Sicht) geändert — das ist der einzige echte
      Verhaltens-Bruch dieser Tranche:** `spec/architecture.md` darf jetzt
      Pfade zu **Code-Modulen** referenzieren (neu erlaubt), aber
      **keine ADR-Bezüge mehr** (neu **verboten** — vorher erlaubt!). Welche
      ADR eine Aussage verbindlich macht, deklariert künftig die ADR selbst
      über ein `Schärft:`-Feld, nicht umgekehrt. **Vor dem Umsetzen:**
      `grep -n "ADR-[0-9]" spec/architecture.md` — jeder Treffer braucht eine
      Entscheidung (Verweis entfernen + Bindung ggf. in die betroffene ADR
      verlagern, oder als bewusste Abweichung dokumentieren).
- [ ] **Neues §3.7 „Ein Kommentar beschreibt, was da ist"** — Fünf-Klassen-Test
      (Zusage · Kopplung · Abgrenzung · Rang-Zeiger · Grenze), Indikativ statt
      Konjunktiv, keine Verweise auf entfernten Code. Gilt auch für
      Zustandsfelder (Roadmap-/Register-/Meilenstein-Status-Zellen: Zustand +
      auflösbarer Beleg-Anker, keine Chronik). **Deckt sich mit CLAUDE.md**s
      bereits gelebter Kommentar-Policy dieses Repos — reine Kodifizierung,
      kein Verhaltenswechsel für uns.
- [ ] **§4 Quality-Gates-Tabelle vollständig entfernt.** m-traces `AGENTS.md`
      führt aktuell **noch eine volle Gate-Tabelle** — das ist nach neuem
      Kanon ein Duplikat-Verstoß (der Gate-Index lebt exklusiv in
      `harness/README.md` §Sensors, Tranche 4). §4 wird auf Regel + Zeiger
      gekürzt: „Kein Target nennen, das im Makefile nicht existiert — auch
      nicht in Prosa."
- [ ] **ID-Referenz-Absatz präzisiert:** Nur Anforderungs-IDs und ADR-Nummern
      gehören in Commit-/PR-Referenzen; neue Struktur-IDs `SPEC-<NNN>`/
      `ARC-<NNN>` (falls über Tranche 5/7 eingeführt) adressieren nur
      *innerhalb* der Spec und gehören **nicht** in Commit-Messages.
- [ ] **§6 Workflow, Schritt 8 wird expliziter Rollenwechsel:** Bericht →
      Handoff an Reviewer (`.harness/skills/reviewer.md`) → Verifier, kein
      Self-Review. Deckt sich mit dem seit `slice-007` gelebten
      Review-Report-Prozess — kodifiziert nur, was schon passiert.
- [ ] `make docs-check` grün, `harness/conventions.md`-Verweise
      (Tranche 2 muss vorher fertig sein) konsistent.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `AGENTS.md` | komplett neu befüllt | siehe DoD, zu viele Querbezüge für Patch |
| `spec/architecture.md` | ggf. update | nur falls `grep -n "ADR-[0-9]"` Treffer liefert (§3.4-Bruch) |

**Bereits geklärt (Recherche vor Schnitt, 2026-09-13):** Direkter Diff
`AGENTS.template.md` v3.5.1 → v6.8.0 gezogen. Die oben gelisteten Punkte sind
die **vollständige** Delta-Menge des Templates (keine Auswahl) — ergänzt um
die Einschätzung, welche für m-trace einen echten Verhaltenswechsel bedeuten
(§3.4, §4) und welche nur nachträglich kodifizieren, was schon gelebte
Praxis ist (§3.3-Variante-2, §3.7, §6-Schritt-8).

**Offen für die Implementierung:**
1. Rang-5-Umbenennung („Wellen-Sequenz") hängt an Tranche 8s
   Owner-Entscheidung (`welle-02` §8) — bei Umsetzung prüfen, ob die
   Entscheidung schon gefallen ist, sonst mit dem *bisherigen* Wortlaut
   („aktuelle Welle") arbeiten und einen Folge-Punkt vermerken.
2. `spec/architecture.md`-ADR-Referenzen: Umfang erst beim `grep` bekannt,
   nicht vorab geschätzt.

## 4. Trigger

- **`in-progress`:** nach `slice-013` (Konventions-Index muss stehen, AGENTS.md
  verweist darauf).
- **Rückführung `in-progress` → `open`:** falls der `spec/architecture.md`-Audit
  (§3.4-Bruch) einen größeren Umbau offenlegt, der eine eigene
  Owner-Entscheidung braucht, statt in dieser Tranche mitzulaufen.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + `spec/architecture.md`-Audit
dokumentiert (auch wenn 0 Treffer) + Closure-Notiz + `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **§3.4-Bruch ist der einzige Punkt mit echtem Korrektur-Risiko.** Anders
  als die übrigen Deltas (kodifizieren bereits gelebte Praxis) verlangt
  dieser eine tatsächliche Prüfung von `spec/architecture.md` — Umfang
  unbekannt bis zum `grep`.
- **Terminologie-Kopplung an Tranche 8.** Ein AGENTS.md-Rewrite vor der
  Roadmap-Entscheidung riskiert einen internen Wortlaut-Bruch (AGENTS.md
  sagt „Wellen-Sequenz", roadmap.md noch „aktuelle Welle"). Sequenzierung
  beachten (§4 oben).
- **Kein ADR nötig für den Rewrite selbst** — reine Nachführung des
  Root-Dokuments gegen die neue Baseline-Form, keine neue
  Architekturentscheidung. Ein ADR wäre nur nötig, falls der
  `architecture.md`-Audit eine tatsächliche ADR-Bindungs-Änderung
  (`Schärft:`-Feld) an einer **Accepted**-ADR verlangt — das würde die neue
  ADR selbst tragen, nicht dieser Slice.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

### Sub-Area: AGENTS.md (Werkzeug/Prozess)

Root-Dokument-Nachführung gegen die neue Baseline-Form, kein Produktcode
direkt berührt (außer ggf. `spec/architecture.md`-Referenzen). Ohne Welle
(Modul 5 „ohne Welle" — hier: Teil von `welle-02`), ADR nur falls der Audit
eine ADR-Bindungsänderung verlangt (siehe §6).
