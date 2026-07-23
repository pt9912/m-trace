# Claude Code — Einstieg für m-trace

Dieses Repo folgt dem AI-Harness-Prozess (ai-harness-course **v3.5.0**, vendored
unter `.harness/baseline/v3.5.0/`). Diese Datei ist **derivativ**: der
maßgebliche Einstieg ist [`AGENTS.md`](AGENTS.md); bei Konflikt gelten die dort
genannten kanonischen Quellen ([`harness/README.md`](harness/README.md) §Source
Precedence).

## Vor jeder Änderung an Code oder Dokumentation lesen

1. [`harness/README.md`](harness/README.md) — Harness-Einstieg, Source Precedence.
2. [`AGENTS.md`](AGENTS.md) — Briefing + Hard Rules (§3) + Gate-Tabelle.
3. [`harness/conventions.md`](harness/conventions.md) — ID-Schemata, Adaptionen
   (`MR-001`…), Modi, Sensor-Bindung.
4. Der **aktive Slice / die aktive Welle** unter
   [`docs/plan/planning/`](docs/plan/planning/) (`in-progress/`); die aktuelle
   Welle nennt [`docs/plan/planning/in-progress/roadmap.md`](docs/plan/planning/in-progress/roadmap.md).
5. Die referenzierten **ADRs** unter [`docs/plan/adr/`](docs/plan/adr/).
6. Die referenzierten **Anforderungen** unter [`spec/`](spec/) (`Contract >
   Technical > View`, Details in `AGENTS.md` §2).

Das Regelwerk-Bundle wird **pro Entscheidung** nachgeschlagen (Index:
`.harness/baseline/v3.5.0/regelwerk/README.md`), **nicht** als Ganzes in den
Kontext geladen.

## Nicht verhandelbar (voll in `AGENTS.md` §3)

- **Docker-only.** Checks/Gates nur über `make`-Targets — nie `go`, `pnpm`,
  `docker build`, `pip`, `apt` direkt. Host-Deps nur via `make host-deps`.
- **Abschluss nur mit echter Gate-Ausgabe.** Vor „fertig": `make gates`
  (CI-äquivalent). Bei reinen Markdown-Edits unter `spec/`, `docs/`, `examples/`,
  `README` reicht `make docs-check`.
- **Kennungs-Linkpflicht.** Nackte Requirement-Kennungen (`F-`/`NF-`/`MVP-`/
  `AK-`/`RAK-`/`R-`) im Fließtext müssen verankerte Links sein — sonst bricht
  `ids` in `make gates` (seit `welle-01`).
- **Immutable ADRs** nach `Accepted`; **keine Gate-Senkung ohne ADR**; **keine
  Inline-Suppressions** (Vulnerability-Ausnahmen nur in `.security/vulnignore.yaml`).
- **Neue Arbeit** als `slice-<NNN>` / `welle-<NN>` in Kanon-Form (`MR-007`);
  der Bestand `plan-<version>.md` ist grandfathered.

Kein Erfolg ohne echte Gate-Ausgabe; bei Quellen-Konflikt den Konflikt melden und
der höherrangigen Quelle folgen.
