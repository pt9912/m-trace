# Welle 01: Requirement-Link-Konvergenz

**Lifecycle:** Die aktive Welle liegt flach unter `docs/plan/planning/`; bei
Closure `git mv` nach `done/` (neben `welle-01-results.md`). Zustand = Verzeichnis-
Position; ob *aktuell* oder *geplant*, sagt die [Roadmap](../in-progress/roadmap.md).

**Zielmeilenstein:** kein Meilenstein-Bezug (interne Konvergenz-/Harness-Arbeit).

**Verantwortlich:** Harness-Migration. **Datum:** 2026-07-23.

---

## 1. Welle-Ziel

Das opt-in-Modul `ids` repo-weit scharfschalten — jede nackte Requirement-Kennung
im Fließtext zeigt als **verankerter** Link auf ihre Definition. Damit graduiert
`conventions.md` §Requirement-Link-Konvergenz von „noch nicht aktiviert" auf
durchgesetzt. Erste kanonische Welle nach der v3.5.0-Migration (parallel dazu,
dass `slice-001` der erste Slice der neuen Form ist).

## 2. Trigger (Welle startet)

- v3.5.0-Migration W1–W7 abgeschlossen (`ids` als W7-Advisory-Ergebnis hierher
  ausgelagert).
- MR-007 (Kanon-Slice/Welle-Form) deklariert.

## 3. Closure-Trigger (Welle schließt)

- Alle Slices dieser Welle in `done/`.
- `ids` in `make gates` **repo-weit** aktiv (minus dokumentierte Exempts:
  immutable ADRs, `done/`, `CHANGELOG.md`), keine Falschbefunde.
- `conventions.md` §Requirement-Link-Konvergenz + Modus-Zeile „Requirement-Links"
  auf durchgesetzt graduiert.
- Closure-Notiz `welle-01-results.md`.

## 4. Slices in dieser Welle

Zustand jedes Slice = sein Lifecycle-Verzeichnis (nicht hier gespiegelt).

| Slice | Titel | Scope |
|---|---|---|
| [`slice-001`](slice-001-requirement-anker-ids.md) | Requirement-Anker + `ids` auf Spec-Straten | `spec/**`, F/NF/MVP/AK/RAK, 372 Anker |
| `slice-002` (geschnitten bei slice-001-Closure) | `ids`-Scope-Ausweitung: `R-`-Familie + risks-backlog, docs/user, examples, Planning-`in-progress/` | Rest der aktiven Doku |

## 5. Abhängigkeiten

- `slice-002` wird von `slice-001` blockiert (teilt die Anker-Infrastruktur +
  das `ids`-Config-Muster; erst Spec-Straten grün, dann Scope-Ausweitung).
- Keine nachfolgende Welle hängt an dieser (interne Konvergenz).

## 6. Out-of-Scope für diese Welle

- **Immutable Accepted-ADRs** (`docs/plan/adr/**`): Body-Edit = Immutabilitätsbruch
  (MR-002, analog W5) → `ids` nimmt ADRs dauerhaft aus, kein Link-Retrofit.
- Historischer Bestand (`done/`, `CHANGELOG.md`).
- `citations`/`sources`-Module (keine Marker in m-trace; nicht Teil dieser Welle).

## 7. Closure-Notiz

Die Welle konvergierte die Requirement-Verlinkung in zwei Slices: `slice-001`
baute die Anker-Infrastruktur und verankerte die Spec-Straten, `slice-002` weitete
das auf die R-Familie und die übrige Doku aus. Am Ende waren nackte
Requirement-Kennungen durchgängig verankerte Links, und `ids`/`anchors`/`links`
erzwingen das seither maschinell in `make gates`. Der detaillierte Nachweis
(RTM-Zählung, Vorher/Nachher je Modul) liegt in `welle-01-results.md`; die
Steering-Lehre — advisory RTM (`--trace`) nach Doku-Struktur-Edits separat prüfen —
ist dort und in der `slice-001`-Note festgehalten.
