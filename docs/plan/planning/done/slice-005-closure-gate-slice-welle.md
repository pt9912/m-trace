# Slice 005: Closure-Gate greift auf slice/welle (ADR-0010-Loch schließen)

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle (Werkzeug-Fix).

**Bezug:** ADR-0010 (Closure-Note-Pflicht), `harness/conventions.md` (MR-007
slice/welle-Form), `slice-004` (Befund beim Verifizieren).

**Autor:** Harness-Migration. **Datum:** 2026-07-23.

---

## 1. Ziel

Das in ADR-0010 verankerte Closure-Note-Gate (`scripts/check_closure_notes.py`,
`make verify-closure-notes`) prüft die **kanonische MR-007-Form** (`slice-*`,
`welle-*`) tatsächlich — heute ist es dort ein No-Op. Zwei Defekte, in
`slice-004` beim Verifizieren entdeckt:

1. **Glob** `plan-*.md` → `slice-*`/`welle-*` werden nie erfasst.
2. **Heading-Matcher** greift die **erste** Überschrift mit „Closure" — im
   slice/welle-Template ist das `## 5. Closure-Trigger` (Boilerplate) statt
   `## 7. Closure-Notiz`.

Folge: die Notizen von `slice-001..004` (inhaltlich echt) werden nie validiert.

## 2. Definition of Done

- [x] `check_closure_notes.py` erfasst `plan-*` **und** `slice-*`/`welle-*`.
- [x] Der Matcher wählt die tatsächliche `Closure-Notiz`-Sektion (bevorzugt
      `Closure-Not…`, Fallback = erste `Closure`-Überschrift für Bestand).
- [x] Kein Regress: der Migrationsplan (`plan-harness-v3.5.0-migration.md`)
      bleibt gültig.
- [x] `welle-01`-Backfill: `welle-01-requirement-link-konvergenz.md` §7 auf
      ≥2 Sätze. (`welle-01-results.md` passt bereits — seine `## Verifikation
      (Closure-Trigger)`-Sektion greift über den Fallback; kein Grandfather nötig.)
- [x] `make gates` grün (inkl. `verify-closure-notes`, das jetzt `slice-001..004`
      mitzählt).
- [x] Closure-Notiz.

## 3. Plan (vor Code)

**A) Discovery.** Statt eines einzelnen `--glob` iteriert der Checker über die
drei Familien-Präfixe (`plan-`, `slice-`, `welle-`) in `done/`. `--glob` bleibt
als Override erhalten (Rückwärtskompatibilität).

**B) Heading-Matcher.** Alle `Closure`-Überschriften sammeln; bevorzugt die, die
`Closure-Not…` (Notiz/Note) trägt; sonst die erste (Alt-Verhalten). Die Sektion
läuft wie bisher bis zur nächsten gleich-/höherrangigen Überschrift.

**C) Backfill.** `welle-01-requirement-link-konvergenz.md` §7-Note auf ≥2 Sätze
heben. `welle-01-results.md` braucht nichts — seine Verifikations-Sektion greift
über den Heading-Fallback.

| Datei / Komponente | Änderungs-Art |
|---|---|
| `scripts/check_closure_notes.py` | Discovery + Heading-Matcher |
| `docs/plan/planning/done/welle-01-requirement-link-konvergenz.md` | §7-Note erweitern |

## 4. Trigger

- **`in-progress`:** slice-004 done (Befund steht). Erfüllt.
- **Rückführung:** rein Werkzeug + Doku; kein Produktcode berührt.

## 5. Closure-Trigger

DoD grün + `make gates` grün + Closure-Notiz; `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Alt-Bestand:** die 41 grandfatherten Pläne bleiben exempt; die Discovery-
  Erweiterung erfasst nur neue Familien, senkt also nichts ab.
- **Heading-Heuristik:** der Fallback auf die erste `Closure`-Überschrift hält
  Pläne gültig, die nur eine generische `Closure`-Sektion führen.

## 7. Closure-Notiz

Der Fix hat zwei Ursachen getrennt behandelt: die **Discovery** iteriert jetzt über
die drei Familien-Präfixe (`plan-`/`slice-`/`welle-`) statt eines einzelnen
`plan-*`-Globs, und der **Heading-Matcher** bevorzugt die `Closure-Not…`-Sektion,
fällt aber auf die erste `Closure`-Überschrift zurück — so blieb der Migrationsplan
(nur `## 7. Closure-Notiz`) unverändert gültig, und `welle-01-results.md` passt über
seine Verifikations-Sektion ohne Grandfather. Nach dem Fix stieg die geprüfte Menge
von 1 auf 7 Pläne; einziges echtes Loch war die Platzhalter-Note in
`welle-01-requirement-link-konvergenz.md` (nur HTML-Kommentar), die nun eine echte
Zwei-Slice-Zusammenfassung trägt. **Lehre:** ein frisch akzeptiertes Gate gegen die
*aktuelle* Namenskonvention gegenprüfen — ADR-0010 kam mit dem `plan-*`-Glob aus der
Bump-Ära, während MR-007 längst auf slice/welle umgestellt hatte; das Gate lief grün
und maß trotzdem nichts. Der `--glob`-Override bleibt für Sonderläufe erhalten.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Closure-Gate (Werkzeug)

- **Modus:** Brownfield (bestehendes Gate mit Deckungslücke).
- **Konventionen-Dichte:** hoch — ADR-0010 + MR-007 sind die Regel.
- **Phase-Reife:** frisch akzeptiertes Gate; die Lücke ist ein Erst-Fix.
- **Evidenz-/Diskrepanz-Risiko:** niedrig — reine Discovery-/Matcher-Logik,
  mit Vorher/Nachher-Lauf gegen `slice-001..004` verifizierbar.
- **Reconciliation-Aufwand:** dieser Slice.
