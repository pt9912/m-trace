# W4 — Werkzeug-Triage des Risiko-Registers

> **Stand:** 2026-07-23. Ergebnis der Welle **W4** des
> [v3.5.0-Migrationsplans](done/plan-harness-v3.5.0-migration.md). Triagiert die
> aktiven Einträge aus [`risks-backlog.md`](risks-backlog.md) gegen die
> v3.5.0-Werkzeug-Triade (Modul 7 §Werkzeug-Wahl bei Diskrepanz):
> **Carveout · ADR · Roadmap-Kandidat**. Additiv/netzlos — Verifikation
> `make docs-check`. Keine Statusänderung an bestehenden Risiken.

## 1. Frage

Der v3.5.0-Kanon trennt drei Werkzeuge: **Carveout** (einzelne, temporäre
Gate-Senkung mit messbarem Auflösungs-Trigger), **ADR** (permanente
Architekturentscheidung) und **Roadmap-Kandidat** (aufgeschobene Arbeit mit
Re-Eval-Trigger — kein Gate wird gesenkt). m-traces `risks-backlog.md` datiert
vor dieser Adoption. W4 fragt für jeden **aktiven offenen** Eintrag: welches der
drei Werkzeuge trägt ihn tatsächlich?

Modul 7 stellt zwei sequenzielle Fragen — **Granularität vor Temporalität**:

1. **Cluster oder einzelne Diskrepanz?** Cluster im selben Geltungsbereich →
   **BF-Sub-Area-Markierung** (Modus-Deklaration, kein Carveout-je-Fall).
2. **Trigger ernst erreichbar?** Ja → **Carveout**; Nein → **permanent → ADR**.
   Wird kein Gate gesenkt, ist der Eintrag gar keine Diskrepanz, sondern
   aufgeschobene Arbeit → **Roadmap-Kandidat**.

## 2. Menge

Trigger-Status der Aktiv-Tabelle (`§1.1`) zum Triage-Zeitpunkt:

- **Offen (⬜):** [`R-9`](risks-backlog.md#r-9), [`R-12`](risks-backlog.md#r-12), [`R-13`](risks-backlog.md#r-13), [`R-28`](risks-backlog.md#r-28), [`R-30`](risks-backlog.md#r-30) — triagiert unten.
- **Trigger eingetreten / gelöst, Backlog-Status stale:** [`R-23`](risks-backlog.md#r-23) (OpenSSL
  `CVE-2026-45447` ist seit `df4d070` **nicht mehr** in `.security/vulnignore.yaml`
  — der `expires`/Base-Update-Trigger ist eingetreten), [`R-31`](risks-backlog.md#r-31) (TS-Mutation-Gate,
  im Eintrag selbst als **GELÖST** vermerkt). Beide sind keine Triage-Kandidaten;
  ihre Überführung nach `§1.2` ist Register-Hygiene (Folge-Punkt, §5).
- Die **🟢 aufgelösten** Zeilen ([`R-5`](risks-backlog.md#r-5)/[`R-7`](risks-backlog.md#r-7)/[`R-10`](risks-backlog.md#r-10)/[`R-11`](risks-backlog.md#r-11)/[`R-15`](risks-backlog.md#r-15)/[`R-17`](risks-backlog.md#r-17)/[`R-20`](risks-backlog.md#r-20)/[`R-22`](risks-backlog.md#r-22)/[`R-24`](risks-backlog.md#r-24)/
  [`R-25`](risks-backlog.md#r-25)/[`R-26`](risks-backlog.md#r-26)/[`R-27`](risks-backlog.md#r-27)/[`R-29`](risks-backlog.md#r-29)) sind erledigt und kein Werkzeug-Kandidat.

## 3. Triage

| R-N | Wesen | Werkzeug | Begründung |
|---|---|---|---|
| **[`R-13`](risks-backlog.md#r-13)** | Security-Gate-Suppression: drei `trixie-slim`-OS-CVEs (`image-scan`), kein Runtime-Pfad, ohne Upstream-Fix | **BF-Sub-Area-Markierung** (kein CO-`NNN`) | Modul-7-Frage 1: Teil eines **Clusters** vieler OS-CVE-Suppressions im selben Geltungsbereich (perl-base, sqlite3, acl, attr, gzip, util-linux, …). Ein CO-File je CVE wäre die explizit gewarnte Carveout-Kaskade. Bereits carveout-verwaltet in `.security/vulnignore.yaml` (reicher als das CO-Template) → [MR-006](../../../harness/conventions.md). |
| **[`R-9`](risks-backlog.md#r-9)** | K8s-Smoke-Label-Allowlist noch Compose-Lab-spezifisch | **Roadmap-Kandidat** | Kein Gate heute gesenkt (die K8s-Smoke-Stage existiert nicht). Forward-Trigger „bei K8s-Smoke-Gate". Aufgeschobene Härtung, keine Diskrepanz. |
| **[`R-12`](risks-backlog.md#r-12)** | WebRTC-`getStats()`-Drift zwischen Browser-Majors | **Roadmap-Kandidat** (überwacht) | Kein Gate gesenkt; ein **aktiver** Sensor (`webrtc-drift.yml`) detektiert die Drift. Re-Eval bei Browser-Major-Bump. Überwachtes Risiko, kein Carveout. |
| **[`R-28`](risks-backlog.md#r-28)** | Multi-Replica-`ingest_sequencer` (RAM-Counter) | **Roadmap-Kandidat** | Ingenieurarbeit mit Trigger (`plan-0.23.0` Tranche 2). Kein Gate-Exception — ein fehlendes Feature, kein gesenkter Check. |
| **[`R-30`](risks-backlog.md#r-30)** | Postgres-SSE-Backfill [`R-27`](risks-backlog.md#r-27)-analoge Skip-Exposition | **Roadmap-Kandidat** | dito — aufgeschobene Bewertung mit Trigger, kein Gate gesenkt. |

**ADR:** keiner der aktiven offenen Einträge ist eine **permanente
Architekturentscheidung** — die architektonisch-permanenten Risiken ([`R-1`](risks-backlog.md#r-1)..[`R-4`](risks-backlog.md#r-4))
sind bereits per ADR-0001..0003 aufgelöst; die Out-of-Scope-Stände (OS-1..OS-5,
[`RAK-71`](../../../spec/lastenheft.md#rak-71)) sind normativ im Lastenheft, nicht in einer ADR.

## 4. Ergebnis

1. **`docs/plan/carveouts/` besteht** (Index + Verweis aufs vendored Template),
   ist aber **ehrlich leer**: es gibt keine einzelne, generische, temporäre
   Gate-Senkung, die dort zu materialisieren wäre.
2. **Der Security-Gate-Suppression-Cluster** ([`R-13`](risks-backlog.md#r-13) + die vielen nicht-R-N
   Nightly-Audit-Suppressions) ist eine **BF-Sub-Area-Markierung**, geführt in
   der bestehenden Registry `.security/vulnignore.yaml` — deklariert als
   **[MR-006](../../../harness/conventions.md)** plus Modus-Zeile. **Nicht** in
   CO-Files dupliziert (Single Source of Truth, Anti-Kaskade).
3. **Die Roadmap-Kandidat-Klasse** ([`R-9`](risks-backlog.md#r-9), [`R-12`](risks-backlog.md#r-12), [`R-28`](risks-backlog.md#r-28), [`R-30`](risks-backlog.md#r-30) und die Register-Praxis
   generell) bleibt im Risiko-Register. Dass m-trace ein RAK-gekoppeltes
   Roadmap-Discovery-Register **ohne** direktes Kanon-Äquivalent führt, ist als
   **[MR-005](../../../harness/conventions.md)** deklariert.

## 5. Folge-Punkte

- **Register-Hygiene:** [`R-23`](risks-backlog.md#r-23) und [`R-31`](risks-backlog.md#r-31) tragen einen stale `⬜`-Status trotz
  eingetretenem Trigger → Überführung in `§1.2` (aufgelöst). Bewusst **nicht** in
  W4 (W4 triagiert Werkzeug-Zuordnung, nicht Auflösungs-Status).
- **Roadmap-Kandidat-Aufnahme (W6):** sobald die Wellen/Slices-Roadmap-Form steht,
  prüfen, ob [`R-9`](risks-backlog.md#r-9)/[`R-12`](risks-backlog.md#r-12)/[`R-28`](risks-backlog.md#r-28)/[`R-30`](risks-backlog.md#r-30) als forward-looking Kandidaten dorthin wandern
  oder im Register bleiben (MR-005 hält den Zustand offen).
