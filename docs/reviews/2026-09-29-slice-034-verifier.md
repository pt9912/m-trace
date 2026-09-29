# Verifier-Report: slice-034 — 2026-09-29

**Prüf-Art:** DoD- und Entscheidungs-Konformität + Plan-vs-Code-Diff (Modul 11).
**Nicht Gegenstand:** die Review-Findings (Modul 10, gelaufen —
`2026-09-29-slice-034.md`). Deren Einstufung wird hier zitiert, nicht neu
geprüft.

**Gegenstand:** Plan
`docs/plan/planning/in-progress/slice-034-tools-uebernahme.md` (Stand
in-progress) gegen die Commits `8856c40` (Durchsetzungsschicht, Owner-Entscheid
ohne Slice-Plan — außerhalb des slice-034-Umfangs), `7b9ac06`
(slice-034-Implementierung), `577502e` (Review-Nachfolge F-1/F-2). Prüf-Stand:
`HEAD` = `de938db`, Arbeitsbaum sauber.

**Eingangs-Kontext:** Plan-Volltext, Review-Report `2026-09-29-slice-034.md`,
Emission `/tmp/aih-v6.13.0/` (ephemeral — F-6), Adoptierte im Volltext,
`Makefile`, `harness/mk/*.mk`, `git diff`/`git show` der Range
`409f601..7b9ac06`. Berührte IDs: keine (Harness-/Wartungs-Commits, exempt nach
`AGENTS.md` §5.1); getrackt: `R-32` (F-2-Nachfolge).

---

## DoD-Prüfung (je Item, eigene Läufe)

### Item 1 — Drei Tools aus der aih-v6.13.0-Emission übernommen — ✅ (mit dokumentierter Abweichung)

- `cmp tools/harness/slice-mv.sh` gegen `/tmp/aih-v6.13.0/tools/harness/slice-mv.sh` → **identisch**.
- `cmp tools/harness/history-range-guard.sh` → **identisch**.
- `cmp tools/harness/baseline-verify.sh` → **differiert**, `diff -u` zeigt
  **exakt einen Hunk** (Z. 86-91): der sed-Filter `s|^\./||` auf der Soll-Seite
  des Vollständigkeits-Vergleichs samt Begründungs-Kommentar. Keine weitere
  Zeile berührt. Das ist genau die behauptete, dokumentierte Abweichung
  (Bootstrap slice-025 erzeugte `SHA256SUMS` mit `./`-Präfix, Emission ohne).
- **Integritäts-Achse unangetastet:** `sha256sum -c SHA256SUMS` (Z. 76, Output
  unterdrückt, nur Exit-Code zählt) und GNU-escape-Vorbedingung (Z. 63-70)
  stehen im unberührten Teil. Der Filter normalisiert nur die *Soll-Seite* der
  Listen; die Digest-Prüfung läuft weiterhin gegen die committete `SHA256SUMS`.
- **Baseline-Daten unangetastet:** `git diff 409f601..HEAD -- .harness/baseline/`
  → leer; `cmp` Arbeitsstand `SHA256SUMS` gegen `git show HEAD:…SHA256SUMS` →
  identisch; `git log -- .harness/baseline/` → letzte Berührung
  `6cab73b` (Vendoring) / `a06c387` (v6.8.0-Entfernung), beide vor der Range.
  Der Owner-Entscheid „kein Regeneration der SHA256SUMS" ist eingehalten.

**Anmerkung:** das DoD-Wort „byte-identisch" ist wörtlich für 2/3 Dateien
erfüllt; für baseline-verify.sh steht die Abweichung bisher nur in der
Commit-Message („Abweichung (Plan §6)"), nicht im Plan selbst — das ist F-4
(Nachführung vor Closure), kein neuer Befund.

### Item 2 — harness/mk/{slice-mv,baseline}.mk übernommen, GATE_CHECKS inert — ✅

- `cmp` beider Fragmente gegen `/tmp/aih-v6.13.0/harness/mk/` → **identisch**.
- `GATE_CHECKS += baseline-verify` (`harness/mk/baseline.mk:9`) hat **keinen
  Verbraucher**: `grep GATE_CHECKS Makefile` → 0 Treffer; das `gates:`-Target
  (`Makefile:832`) listet `baseline-verify` nicht und expandiert `GATE_CHECKS`
  nicht. Der Glob-Include (`Makefile:33`, `include harness/mk/*.mk`) macht die
  Targets aufrufbar, zieht sie aber nicht in den Aggregat-Gate.

### Item 3 — Makefile: history-range-guard-Target + Vorbedingung — ✅

- Target existiert (`Makefile:38-40`), `.PHONY` gesetzt.
- `doc-immutable: history-range-guard` und `doc-commits: history-range-guard`
  (`Makefile:42-43`) — beide hängen es als Vorbedingung.
- **Kein record-gates-Rest:** `grep record-gates` über `Makefile`,
  `harness/mk/`, `tools/harness/`, `.claude/` → 0 Treffer.

### Item 4 — Verifikations-Läufe (selbst gefahren) — ✅

| Lauf | Ergebnis | Erwartung | Befund |
|---|---|---|---|
| `make baseline-verify` | Exit 0 | grün, 54 Dateien | „baseline-verify: v6.13.0 OK — 54 Dateien (Integritaet + Vollstaendigkeit, netzlos)" |
| `make slice-mv` (ohne Argument) | Exit 2 | Usage, Exit 2 | Usage-Text auf `stderr`, Abbruch in `harness/mk/slice-mv.mk:34` |
| `make history-range-guard RANGE=HEAD~5..HEAD` | Exit 0 | Exit 0 | „Range 'HEAD~5..HEAD' aufgeloest, 5 Commit(s) — OK." |
| `make docs-check` | Exit 0 | grün, 0 Befunde | **191 Datei(en), 0 Befund(e)** |

Negativ-Proben über die Soll-Abdeckung hinaus (fail-closed nachweisen):
`make history-range-guard` ohne RANGE → Exit 2 (Usage);
`make history-range-guard RANGE=nicht-existiert..HEAD` → Exit 2 (unlösbar).
Grün über leerer/unlösbarer Range ist ausgeschlossen.

### Item 5 — Closure-Notiz — ✅ (Zustand zur Kenntnis genommen)

Plan §7 ist leer (`*(füllt bei Closure)*`), DoD-Häkchen gesetzt nicht — die
schreibt der Planner bei der Closure **nach** diesem Lauf. Nicht geprüft,
nur festgestellt.

---

## Plan-vs-Code-Diff (§3-Tabelle gegen `7b9ac06` allein)

§3 deklariert drei Zeilen; der Commit berührt 7 Dateien (+546/−1):

| §3-Zeile | Diff-Realität | Konform |
|---|---|---|
| `tools/harness/{slice-mv,baseline-verify,history-range-guard}.sh` — neu | 3 Dateien neu (+292/+102/+97) — genau diese drei | ✅ |
| `harness/mk/{slice-mv,baseline}.mk` — neu | 2 Dateien neu (+34/+9) — genau diese zwei | ✅ |
| `Makefile` — update | +10/−0: Target + 2 Vorbedingungs-Zeilen, sonst nichts | ✅ |
| *(nicht deklariert)* | `roadmap.md` +2/−1 („Aktiv: slice-034" unter Offene Wellen) | ⚠ Nachführung |

**Abweichung:** die Roadmap-Änderung ist im Commit-Message deklariert
(„Roadmap-Drift behoben"), fehlt aber in §3. Sie ist inhaltlich richtig
(Zustandsfeld-Korrektheit), braucht aber die Plan-Nachführung bei Closure —
da im selben Zug F-5 ohnehin den „Aktiv"-Anker bei `done/`-Move nachführt,
gehört beides in die §7-Closure-Notiz (und F-4 den §6-Eintrag). **Antwort auf
Prüffrage: ja, das ist eine Plan-Nachführung bei Closure.**

`8856c40` ist plan-los und gegen §3 nicht prüfbar — stattdessen
12 Dateien (5 Rollen-Agenten, 3 Commands, `.githooks/commit-msg`, `hooks-install.mk`,
`commit-msg-traceability.sh`, `selbstpruefung.sh`, 1013 Insertions), Form
legitimiert laut Review F-7 (kein Gate gesenkt, kein MR-Trigger, exempt nach
§5.1).

---

## Entscheidungs-Konformität

1. **Owner-Entscheid Baseline-Untouchability — eingehalten.** Die Abweichung
   ist einseitig im Tool (Soll-Seiten-Normalisierung), die Baseline-Daten und
   die Integritäts-Achse sind byte-fest in der Range. Keine SHA256SUMS-
   Regeneration, kein Eingriff in `.harness/baseline/`.
2. **Teiladoption Durchsetzungsschicht — Grenze eingehalten.** Nicht adoptiert
   und im Baum auch nicht vorhanden: Stop-Hook (`.claude/hooks/` existiert
   nicht), Command-Guard (kein `pretooluse`-Rest), Gate-Nachweis
   (`record-gates` → 0 Treffer), `span-emit.sh` (existiert nicht).
   Adoptiert: commit-msg-Träger + Selbstprüfung — latent, `core.hooksPath`
   unset (opt-in via `make hooks-install`), Konsistent mit `R-32`
   (Triggerschwelle „vor `make hooks-install`"). Die Grenze wird von den
   Commands (nach `577502e`) selbst korrekt benannt
   (`.claude/commands/implement-slice.md:21` — „ja; Stop-Hook,
   Command-Guard, Gate-Nachweis nein").

---

## Gesamtverdikt

**DoD bestätigt: ja.** Alle fünf Items mit eigenen Läufen belegt; die einzige
Abweichung vom „byte-identisch"-Soll ist exakt die deklarierte
`./`-Normalisierung, minimal, kommentarform-konform und gate-neutral — die
Integritäts-Achse und die Baseline-Untouchability sind unangetastet. Die
Makefile-Verdrahtung ist fail-closed (Negativ-Proben rot) und GATE_CHECKS ist
nachweislich inert.

**Offene Posten (bekannt, vom Review befristet — hier bestätigt, nicht neu
erfunden):**

1. **F-3** — `make gates`/`verify-plan-paths` PHANTOMs wegen Plan-Range; löst
   sich mechanisch mit der Closure-Notiz.
2. **F-4** — Plan §6/DoD führt die baseline-verify-Abweichung noch nicht;
   Nachführung vor Closure (Bedingung fürs ehrliche DoD-Häkchen zu Item 1).
3. **F-5 + Roadmap-§3-Nachführung** — Roadmap-Anker bei `done/`-Move nachführen
   und die Roadmap-Änderung in der Closure-Notiz against halten (nicht in §3
   deklariert).

**Übergabe:** an den Planner — DoD-Häkchen und §7-Closure-Notiz nach
Neuverifikation setzen; dieser Report ist Beleg des Verifier-Laufs und
wird nicht committet-übergeben (kein Commit durch diese Rolle).
