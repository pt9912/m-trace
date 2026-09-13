# Slice 012: Harness-Baseline v6.8.0 vendoren

**Lifecycle:** Zustand = Verzeichnis. **Welle:** [`welle-02`](../welle-02-regelwerk-v6.8.0-migration.md)
(Tranche 1 von 11).

**Bezug:** `.harness/baseline/`, `harness/conventions.md` §Baseline (nur der
Baseline-Verweis, nicht die MR-Struktur — die ist Tranche 2).

**Autor:** Nachgang Regelwerk-Recherche v3.5.1→v6.8.0. **Datum:** 2026-09-13.

---

## 1. Ziel

Die vendorte Regelwerk-Baseline von `v3.5.1` (Kurs-Welle 33) auf `v6.8.0`
(Kurs-Welle 135) heben — reines Vendoring + der Baseline-Zeiger in
`harness/conventions.md` §Baseline (Stand, Pfad, Digest). **Keine** sonstige
inhaltliche Anpassung von `AGENTS.md`/`harness/conventions.md` (die
MR-Container→Index-Umstellung und der AGENTS.md-Rewrite folgen in
Tranche 2/3). Damit steht der neue Regelwerk-Wortlaut netzlos +
integritäts-geprüft im Repo, bevor irgendeine andere Tranche ihn zitiert.

## 2. Definition of Done

- [ ] `.harness/baseline/v6.8.0/{regelwerk,templates}/` + `SHA256SUMS`
      committet (Assets aus dem GitHub-Release `v6.8.0` von
      `pt9912/ai-harness-course`, Digest lokal gegen `SHA256SUMS` verifiziert
      vor dem Commit — nicht nur „hat entpackt").
- [ ] `harness/conventions.md` §Baseline zeigt auf `v6.8.0` (Stand, Pfad,
      SHA256-Digest — analog zum bisherigen `v3.5.1`-Absatz).
- [ ] `.harness/baseline/v3.5.1/` bleibt als historischer Stand erhalten
      (kein Löschen — Audit-Trail, analog wie `v3.5.0` neben `v3.5.1` liegt).
- [ ] `make docs-check` grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.harness/baseline/v6.8.0/regelwerk/**` | neu | aus `lab-regelwerk.zip` entpackt |
| `.harness/baseline/v6.8.0/templates/**` | neu | aus `lab-regelwerk.zip` entpackt |
| `.harness/baseline/v6.8.0/SHA256SUMS` | neu | Release-Asset, unverändert übernommen |
| `harness/conventions.md` §Baseline | update | Baseline-Zeiger `v3.5.1` → `v6.8.0` (nur dieser Absatz, nicht die MR-Struktur) |

**Bereits erledigt (Recherche vor Schnitt, 2026-09-13):** `lab-regelwerk.zip`
+ `SHA256SUMS` von
[`github.com/pt9912/ai-harness-course/releases/download/v6.8.0/`](https://github.com/pt9912/ai-harness-course/releases/download/v6.8.0/)
geladen, `sha256sum -c SHA256SUMS` → `OK`. Struktur-Diff gegen
`.harness/baseline/v3.5.1/` bereits gezogen (Grundlage für den gesamten
`welle-02`-Tranchen-Schnitt): 17 Module (00–16) unverändert nummeriert, aber
Modul 3/4-Grenze neu gezogen (Architektur wandert von Modul 4 zu Modul 3) und
die einzelne `grundlagen-konventionen.md` ist zu sechs Themen-Dateien
aufgeteilt (`grundlagen-begriffe.md`, `-source-precedence.md`,
`-referenz-richtung.md`, `-harness-dateien.md`, `-bootstrap.md`,
`-traceability.md`).

**Offen für die Implementierung:** Den exakten Bestand von
`.harness/baseline/v3.5.0/` (heute schon neben `v3.5.1` vorhanden) als
Präzedenzfall für „wie viele alte Baselines bleiben liegen" heranziehen —
vermutlich bleibt auch `v3.5.1` unbefristet stehen (Audit-Trail), aber das
ist beim Commit zu bestätigen, kein Automatismus.

## 4. Trigger

- **`in-progress`:** jederzeit — reines Vendoring, kein Abhängigkeits-Trigger,
  keine andere Tranche muss vorher fertig sein.
- **Rückführung:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **Diese Tranche allein hat keinen sichtbaren Effekt.** `AGENTS.md`,
  `harness/conventions.md` und `.d-check.yml` zitieren nach diesem Slice
  weiterhin den alten (v3.5.1-)Wortlaut — das ist bewusst so (additiv-zuerst,
  siehe `welle-02` §7) und kein unvollständiger Slice.
- **Kein ADR nötig.** Reines Vendoring einer neueren Baseline-Version, keine
  Architekturentscheidung, kein gesenktes Gate — analog zu `slice-008`
  (v3.5.0 → v3.5.1), nur eine Version weiter.

## 7. Closure-Notiz (nach `done/`)

`.harness/baseline/v6.8.0/{regelwerk,templates}/` + `SHA256SUMS` (54 Dateien)
committet, `harness/conventions.md` §Baseline auf `v6.8.0` umgestellt (Stand,
Pfad, Release-Digest, Vendoring-Datum + Slice-/Welle-Verweis). `v3.5.1`
bleibt unangetastet als Audit-Referenzform liegen, wie in §3 vorgesehen.

**Ein Stolperstein beim Erzeugen von `SHA256SUMS`, der Erwähnung wert ist:**
Der naive Einzeiler
`find . -type f | sed … | xargs sha256sum > SHA256SUMS` erzeugte **zweimal
hintereinander** eine Datei, die sich selbst nicht bestätigt
(`sha256sum -c` meldete `SHA256SUMS: GESCHEITERT`) — die Shell legt die
Ausgabedatei für die Umleitung an, sobald die Pipeline startet, und `find`
lief offenbar oft genug parallel dazu, um die (noch leere) Datei selbst
aufzulisten. Behoben, indem die Dateiliste zuerst vollständig in eine
Shell-Variable eingefangen wurde (`FILES=$(find … )`), bevor `sha256sum`
überhaupt lief — danach 54 von 54 Dateien `OK`, `SHA256SUMS` selbst korrekt
nicht in sich selbst enthalten (wie im `v3.5.1`-Vorbild).

**Verifikation:** `sha256sum -c SHA256SUMS` — 54/54 `OK`, Exit 0.
`make docs-check` grün (nach Nachzug der Folge-Referenzen in `roadmap.md`,
`welle-02` und dieser Slice-Datei selbst — alle drei zeigten zwischenzeitlich
noch auf den `open/`-Vorstand). `make gates` grün.

**Steering-Loop-Lerneintrag:** Ein Shell-Einzeiler, der eine Datei erzeugt
UND deren eigenen Inhalt in derselben Pipeline aufzählt, ist eine
Race-Condition-Falle — unabhängig davon, wie harmlos er aussieht. Für
künftige Baseline-Vendorings: die Datei-Liste immer erst vollständig
einfangen (Variable oder temporäre Datei), dann erst die Zieldatei
schreiben.

**Folge-Slices:** keine unmittelbaren — Tranche 2 (`slice-013`) ist bereits
geschnitten und kann jetzt gegen den frisch vendorten `v6.8.0`-Wortlaut
umgesetzt werden.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Harness-Baseline (Werkzeug/Prozess)

Reines Vendoring + Zeiger-Update, kein Produktcode, keine Spec, kein
Requirement berührt. Ohne direkten ADR-Bezug (die Baseline-Wahl selbst ist
kein architektonischer Trade-off, sondern ein Versions-Bump).
