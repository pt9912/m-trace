# Slice 003: `arch-check` von `check-architecture.sh` auf a-check umstellen

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle (Werkzeug-/Gate-Arbeit).

**Bezug:** `spec/architecture.md` §3.4 (Hexagon-Boundaries), MR-007 (Slice-Form).

**Autor:** Harness-Migration. **Datum:** 2026-07-23.

---

## 1. Ziel

Das Go-Architektur-Gate von `apps/api/scripts/check-architecture.sh` (bespoke
Bash + `go list`) auf **a-check** (`ghcr.io/pt9912/a-check`, digest-gepinnt)
umstellen — **verhaltensgleich** (dieselben Hexagon-Regeln, `make gates` grün),
aber als deklaratives, sprach-erweiterbares Gate der d-check-Werkzeug-Familie.

## 2. Definition of Done

- [ ] `.a-check.yml` (Repo-Wurzel) bildet die 4 `check-architecture.sh`-Regeln ab
      (domain/port/application/adapters + `tech`-Patterns); Scan von der
      Repo-Wurzel, `.security/**`/`node_modules`/`dist` via `exclude` geprunt.
- [ ] `a-check.mk` (digest-gepinnt v0.14.0) eingecheckt; `make arch-check` ruft
      a-check statt der Bash.
- [ ] `check-architecture.sh` + der `apps/api`-`arch-check`-Docker-Wrapper entfernt.
- [ ] `spec/architecture.md` §3.4 + Gate-Tabelle-Verweise auf a-check nachgezogen.
- [ ] `make arch-check` + `make gates` grün.
- [ ] Closure-Notiz.

## 3. Plan (vor Code)

**Verhaltensgleich (3a):** die `adapters→application`-Kante bleibt in `.a-check.yml`
(check-architecture.sh prüft `adapters→application` nicht). Der striktere
Driving-Port-Boundary (Kante entfernen + Handler-Refactor) ist **`slice-004`**.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.a-check.yml` | neu | deklarative Hexagon-Regeln (verifiziert: a-check v0.14.0 → 0 Befunde) |
| `a-check.mk` | neu | `A_CHECK_IMAGE` (digest `f1b8ff5e…`, v0.14.0) + `a-check`/`a-check-graph`-Targets |
| `Makefile` (root) | update | `include a-check.mk`; `arch-check` → a-check (Repo-Wurzel-Scan) |
| `apps/api/Makefile` | update | `arch-check`-Target entfernen |
| `apps/api/scripts/check-architecture.sh` | löschen | durch a-check abgelöst |
| `spec/architecture.md` §3.4 | update | Verweis `check-architecture.sh` → a-check |

**Bonus (kein Regress):** a-check erzwingt die OTel-Zwei-Pfade-Regel (§3.4:
`go.opentelemetry.io` nur in `telemetry`+`driving/http`) und `database/sql`-Scope
präziser, als die Bash-Regex es tat (die verbot Tech nur *im Hexagon*, erlaubte
sie in *jedem* Adapter). Verifiziert grün.

## 4. Trigger

- **`in-progress`:** a-check v0.14.0 verfügbar (`exclude` prunt den Walk → Repo-
  Wurzel-Scan trotz root-eigenem `.security/.trivy-cache`). Erfüllt.
- **Rückführung:** unerwartete a-check-Falschbefunde → `next` (Config schärfen).

## 5. Closure-Trigger

DoD grün + `make gates` grün + Closure-Notiz; `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Digest-Pflege:** ein weiteres GHCR-Image im Nightly-Audit-Radar (wie d-check).
- **`--print-mk`-Digest lag:** a-checks `--print-mk` bettet noch den v0.13.0-Digest
  ein; `a.check.mk` wird explizit auf den verifizierten v0.14.0-Digest gepinnt.
- **Heuristik-Grenze:** a-check liest Importe text-heuristisch — für m-traces
  Go-Code im Spike **null** Falschtreffer; `markers.ignore_symbols` steht bereit.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Architektur-Gate (Tooling)

- **Modus:** Greenfield (Werkzeug-Tausch, kein Bestandscode berührt außer Gate-Wiring).
- **Konventionen-Dichte:** mittel — `spec/architecture.md` §3.4 ist die autoritative
  Regel-Quelle; a-check-Config bildet sie ab.
- **Phase-Reife:** reif; das Gate existiert, nur die Implementierung wechselt.
- **Evidenz-/Diskrepanz-Risiko:** niedrig — verhaltensgleich, Spike-verifiziert
  (0 Befunde = check-architecture.sh; Gegenprobe `net/http` in Domain → Befund).
- **Reconciliation-Aufwand:** dieser Slice; der striktere Boundary = `slice-004`.
