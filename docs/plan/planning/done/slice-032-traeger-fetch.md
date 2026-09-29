# Slice 032: traeger-fetch übernehmen — verifizierter Fetch des ai-harness-init-Trägers

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle — Closure-Bedingung
ist die DoD dieses Slices, keine Wellen-Grenze.

**Bezug:** `scripts/traeger-fetch.sh` (neu), `Makefile` (Target + Pins),
`.gitignore`, Quelle: ai-harness-Scaffold `/tmp/aih-v6.13.0/tools/harness/
traeger-fetch.sh` (ai-harness-init v0.2.5-Emission).

**Autor:** Owner-Auftrag „wir brauchen auch traeger-fetch.sh". **Datum:** 2026-09-29.

---

## 1. Ziel und Abgrenzung

**Ziel:** Den fail-closed Fetch des `ai-harness-init`-Trägers übernehmen —
Transport im digest-gepinntem curl-Bild, Digest-Verifizierung vor der Ablage
(Pin-Kanal Makefile vs. Release-SHA256SUMS), Plattform-Matrix laut-abrechend.
Pin: `TRAEGER_TAG=v0.2.5` + Plattform-Digest `linux/amd64` im Makefile.

**Ausdrücklich NICHT in diesem Slice:**

- **Ausführen des Trägers** (Emit-/Refresh-Läufe von ai-harness-init gegen
  m-trace) — der Träger wird abgelegt und verifiziert; sein Einsatz ist
  eine eigene Entscheidung.
- **Die übrigen aih-Tools** (slice-mv, baseline-verify, history-range-guard,
  commit-msg-Hook, Selbstprüfung) — separate Slices, je eigene Abwägung.
- **aih-Makefile-Aggregator** (`harness/mk/*.mk` + GATE_CHECKS) — m-traces
  Makefile ist etabliert; das Tool hängt als plain Target ein.

## 2. Definition of Done

- [x] `scripts/traeger-fetch.sh` übernommen — byte-identisch zur
      v6.13.0-Emission (Diff gegen `/tmp/aih-v6.13.0`).
- [x] `Makefile`: Glob-Include `include harness/mk/*.mk` (aih-Vorlage) +
      `harness/mk/traeger.mk` mit Pins `TRAEGER_TAG=v0.2.5` +
      `TRAEGER_CARRIER=.harness/state/bin/ai-harness-init`.
- [x] `.gitignore`: `.harness/state/` (der Träger ist fetched State, kein
      Source).
- [x] `make traeger-fetch` gelaufen: Träger liegt unter
      `.harness/state/bin/ai-harness-init`, Digest = Release-Digest
      (`c6a6a171…`, Manifest-Kanal SHA256SUMS), Exit 0.
- [x] `make docs-check` grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `scripts/traeger-fetch.sh` | neu | aus dem aih-Scaffold übernommen (v6.13.0-Emission), Herkunft im Header |
| `Makefile` | update | Target + Pins |
| `.gitignore` | update | `.harness/state/` |

## 4. Trigger

- **Start (`next` → `in-progress`):** sofort — Owner-Auftrag.
- **Rückführungen:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + `make traeger-fetch` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **Einsatz-Semantik des Trägers** (Emit gegen m-traces verkörperte Form) —
  **Ausgang:** entfallen — dieser Slice legt nur den verifizierten Träger
  ab; Emit-Läufe sind eigene Entscheidung.
- **Plattform-Kopplung** — **Ausgang:** entfallen — nur linux/amd64 gepinnt
  (docker-only Repo); weitere Plattformen brachen laut ab (fail-closed
  Teil-Pin-Kopplung des Skripts).

## 7. Closure-Notiz

Der Träger `ai-harness-init v0.2.5` (linux/amd64) liegt digest-verifiziert
unter `.harness/state/bin/` (gitignored State); der Fetch läuft im
digest-gepinntem curl-Bild und verifiziert gegen die SHA256SUMS desselben
Releases (Manifest-Kanal). Die Verdrahtung folgt der aih-Architektur:
Fragment `harness/mk/traeger.mk` (Pin + Target, „EIN KOMMANDO, KEIN GATE")
+ Glob-Include im Root-Makefile — künftige Fragmente greifen automatisch
(Owner-Hinweis: „es wird noch mehr mk's geben").

**Was hat funktioniert:** Byte-identische Übernahme per Diff gegen die
Emission statt Abschreiben — zwei Verfremdungen im ersten Write (Tippfehler
`carier_abs`, duplizierter Block) wurden so vor dem Commit gefangen.

**Was ging anders als geplannt:** Die Pins sind weniger als ursprünglich
vorgesehen — statt `TRAEGER_SHA256_LINUX_AMD64` im Makefile trägt das
Fragment nur den Release-Tag, und die Verifizierung läuft über den
Manifest-Kanal (SHA256SUMS desselben Releases). Das ist die
emittierte-Form-Vorgabe des Fragments („das Fragment führt nur den
Release-Tag"); der Digest-Pin wäre die Dogfood-Hälfte gewesen, die das
Fragment laut Kopf entfallen lässt.

**Steering-Loop-Eintrag:** Guide geschärft: `harness/mk/*.mk`-Fragments
sind die Emission des Init-Tools — Root-Makefile bindet per Glob ein,
kein Inlining von Target-Logik ins Root-Makefile. — liegt in
`Makefile` (Include-Kommentar). Auslöser: `BEO-PLAN`-Nachbarschaft,
Erstvorkommen.

**Beobachtungs-Register (`../observations/`):** keine Beobachtung
angefallen.

**Folge-Slices:** keine.

**Risiken aus §6:** beide entfallen — der Träger ist abgelegt, nicht
ausgeführt (Emit gegen m-trace bleibt eigene Entscheidung); Plattform-
Kopplung entfällt (Manifest-Kanal statt Einzeldigest-Pins).

Werkzeug-Übernahme (Brownfield): das Skript ist aus der aih-Emission
geprüft (Review-Report slice-031-Nachlauf); die Anpassung beschränkt sich
auf Pins und Ablageort. Kein ADR — kein Gate gesenkt.
