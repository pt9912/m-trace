# Slice 030: AGENTS.md §5 auf v6.13.0-Template-Form (Index-Tabelle)

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle — Closure-Bedingung
ist die DoD dieses Slices, keine Wellen-Grenze.

**Bezug:** `AGENTS.md` §5, `harness/rules/` (neu),
`.harness/baseline/v6.13.0/templates/AGENTS.template.md`.

**Autor:** Owner-Hinweis „§5 Dokumentations-Regeln wurde nicht nach dem
neuen Template angepasst". **Datum:** 2026-09-29.

---

## 1. Ziel und Abgrenzung

**Ziel:** AGENTS.md §5 (Dokumentations-Regeln) von der Bullet-Liste auf die
v6.13.0-Template-Form heben: Index-Tabelle (`| # | Regel | Datei |`), kurze
Regeln vollständig in der Tabelle, die zwei Langregeln (Closure-Noten,
Review-Läufe) als Kurzform-Zeiger mit Volltext nach `harness/rules/
{closure-notes,reviews}.md` — dem Schnitt-Prinzip gegen
„Guide-Datei-Wildwuchs" (`v6.13.0` · `regelwerk/modul-09-implementierung.md`
§Hard Rules; Form ist Wahl, hier vom Owner angeordnet).

**Ausdrücklich NICHT in diesem Slice:**

- **Inhaltliche Regel-Änderungen** — alle acht m-trace-Regeln bleiben
  inhaltlich erhalten (RB-Reihe entfällt im ID-Schema-Bezug: Lastenheft
  trägt keine Randbedingungen, MR-003 gilt).
- **`file`-Modul-Ratchet für AGENTS.md** — opt-in, wenn gemessen.

## 2. Definition of Done

- [ ] `AGENTS.md` §5 als Index-Tabelle nach dem v6.13.0-Template (8 Zeilen,
      alle bisherigen Regeln erhalten).
- [ ] `harness/rules/closure-notes.md` + `harness/rules/reviews.md` tragen
      die Volltexte der zwei Langregeln (links vom neuen Pfad aus gültig).
- [ ] `make docs-check` grün (163+ Dateien, keine toten Links).
- [ ] Closure-Notiz.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `AGENTS.md` | update | §5 Bullet-Liste → Index-Tabelle |
| `harness/rules/closure-notes.md` | neu | Volltext Closure-Noten-Regel |
| `harness/rules/reviews.md` | neu | Volltext Review-Läufe-Regel |

## 4. Trigger

- **Start (`next` → `in-progress`):** sofort — Owner-Auftrag.
- **Rückführungen:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + `make docs-check` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **Tote Links beim Volltext-Umzug** (relative Pfade ändern sich von
  Repo-Wurzel auf `harness/rules/`) — **Ausgang:** offen bis docs-check.
- **„Plan-Angabe vs. realer Diff"** (BEO-PLAN) — **Ausgang:** entfallen —
  Plan §3 nennt genau die drei Dateien; der Real-Diff wird vor Closure
  dagegen geprüft (AGENTS.md §6, Schritt 4 — in diesem Slice verkörpert).

## 7. Closure-Notiz

*(füllt bei Closure)*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Embodied-Form-Nachführung (Brownfield): AGENTS.md gegen das v6.13.0-Template
— Append-only für Inhalte, Form-Follow-up hier. Kein ADR — Form-Änderung,
kein Gate gesenkt.
