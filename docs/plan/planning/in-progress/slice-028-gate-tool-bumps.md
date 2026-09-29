# Slice 028: Gate-Tool-Bumps — a-check v0.20.0, d-check v0.79.0

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle — Closure-Bedingung
ist die DoD dieses Slices, keine Wellen-Grenze.

**Bezug:** `a-check.mk`, `d-check.mk`, `Makefile` (Version-Kommentar),
`.a-check.yml`, `.d-check.yml` (nur falls Befunde Konfig-Anpassungen erfordern).

**Autor:** Owner-Anfrage „sind a-check und d-check aktuell" + Antwort. **Datum:** 2026-09-29.

---

## 1. Ziel und Abgrenzung

**Ziel:** Die beiden digest-gepinnten Gate-Tools auf den aktuellen Release-
Stand heben — a-check `v0.19.0` → `v0.20.0` (BREAKING: Port-Glob-
Richtungssegment schaltet `port-locality` nicht mehr still ab; advisory
Hinweis für Port-Globs, die den App-Baum nicht erreichen) und d-check
`v0.75.0` → `v0.79.0` (`links` erkennt Ziel hinter Zeilenumbruch,
Link-Referenz-Definitionen werden unabhängig geprüft, neues opt-in-Modul
`file`) — und die dadurch ggf. neu sichtbaren Befunde triagieren.

**Ausdrücklich NICHT in diesem Slice:**

- **Neue Module scharfschalten** (`file`-Ratchet für AGENTS.md) — opt-in,
  eigener Schritt, wenn die Zeilen-Marge gemessen ist.
- **Konfig-Optimierung** über die Triage hinaus — `.a-check.yml`/`.d-check.yml`
  ändern sich nur, wenn ein neuer Befund eine echte Lücke zeigt.

## 2. Definition of Done

- [ ] `a-check.mk` Digest auf v0.20.0 (Release-Notes:
      `sha256:e8208764…`), Kommentar-Version nachgezogen.
- [ ] `d-check.mk` auf v0.79.0 (DCHECK_IMAGE + DCHECK_DIGEST
      `sha256:b4b8756b…`), Makefile-§-Kommentar (Zeile 19) nachgezogen.
- [ ] `make arch-check` gelaufen; neue Befunde (a-check-Breaking-Change)
      triagiert.
- [ ] `make docs-check` gelaufen; neue Befunde (Zeilenumbruch-Links,
      Referenz-Definitionen) triagiert.
- [ ] `make gates`-Relevanz: beide Gates sind PR-blockierend im CI — grün
      lokal vor Closure.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `a-check.mk` | update | Digest v0.20.0 |
| `d-check.mk` | update | Version + Digest v0.79.0 |
| `Makefile` | update | Versionskommentar Zeile 19 |
| `.a-check.yml` / `.d-check.yml` | update falls nötig | nur bei neuen Befunden mit Konfig-Lücke |

## 4. Trigger

- **Start (`next` → `in-progress`):** sofort — Owner-Auftrag.
- **Rückführungen:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + beide Gates lokal grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **a-check-Breaking-Change meldet nach**, was vorher still verschwieg —
  **Ausgang:** offen bis Triage; ein Befund ist ein Fund, kein Defekt des
  Bumps.
- **d-check 0.79 meldet tote Referenz-Definitionen** in Doku, die v0.75
  übersprungen hat — **Ausgang:** offen bis Triage.

## 7. Closure-Notiz

*(füllt bei Closure)*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Werkzeug-Bump auf digest-gepinnten Pfad (Greenfield-Muster, Präzedenz
d-migrate-Pins). Kein Gate gesenkt — Versionen vorwärts.
