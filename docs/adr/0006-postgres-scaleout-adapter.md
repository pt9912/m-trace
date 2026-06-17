# 0006 — Postgres-Runtime-Adapter für Production-Scale-out

> **Status**: Accepted
> **Datum**: 2026-06-17
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)
> **Bezug**: `spec/lastenheft.md` RAK-91 (Reaktivierung von „defer" auf
> „proceed, optional"), NF-20/NF-22/NF-23; [ADR-0005](0005-production-ops-backends.md)
> (amendet die Postgres-Vertagung); `docs/planning/in-progress/risks-backlog.md`
> R-26; Ausführung in `docs/planning/open/plan-0.23.0-postgres-scaleout.md`.

## Kontext

[ADR-0005](0005-production-ops-backends.md) hat den Postgres-Runtime-
Adapter unter RAK-91 **deferred** — mit drei messbaren Reaktivierungs-
Triggern: (1) ≥ 2 API-Replicas auf demselben Store, (2) RPO ≤ 15 min /
RTO ≤ 30 min verbindlich, (3) Retention-Queries über > 10 Mio Events mit
p95 < 2 s.

Der Load-/Soak-Smoke (`plan-0.22.5-load-smoke`, abgeschlossen) hat die
**vertikale Lab-Lastfähigkeit** der Ingest→Persistenz→Read-Kette belegt:
eine SQLite-Single-Instance hält im 4h-Soak 3842 ev/s ohne stillen
Verlust (55,3 Mio Events) und liefert keyset-indizierte Reads im
ms-Bereich. Trigger #3 ist damit **gemessen nicht ausgelöst**.

Was **fehlt** (festgehalten als R-26): **horizontale Production-Scale-
out-Evidenz** — Multi-Tenant-Isolation unter echter Last und vor allem
Multi-Replica-Betrieb (≥ 2 API-Instanzen auf einem geteilten Store).
Letzteres ist mit SQLite **strukturell nicht erbringbar**: SQLite ist
file-/single-writer-gebunden und nicht mehrprozess-/mehrhost-writer-safe.
Scale-out setzt zwingend einen netzwerkfähigen Store voraus — also genau
den per ADR-0005 zurückgestellten Postgres-Adapter.

## Entscheidung

Der **Postgres-Runtime-Adapter wird reaktiviert** und als **optionaler,
nicht-default** Persistenz-Adapter umgesetzt (RAK-91: „defer" → „proceed,
optional"):

- SQLite **bleibt der Default** (`MTRACE_PERSISTENCE=sqlite`) und der
  lokale Standard-Store. Postgres ist opt-in über
  `MTRACE_PERSISTENCE=postgres` + DSN. Keine versteckte
  Pflichtabhängigkeit in der lokalen Standardumgebung.
- Das Postgres-Schema entsteht durch **Portage der Migrationshistorie
  V1–V7** nach Postgres-Dialekt (bzw. `schema.yaml`-Vollausbau +
  Generierung) — `schema.yaml` deckt nur die V1-Baseline (5 Tabellen) und
  ist **kein** fertiger One-Shot-Anker; `d-migrate --target postgres` ist
  vorab nachzuweisen (heute nur `--target sqlite` belegt). Der hartkodierte
  `driverName = "sqlite"` wird parametrisiert.
- Der Adapter implementiert die sechs Driven-Ports des SQLite-Adapters.
  Korrektheit: die adapter-agnostische Contract-Suite
  (`apps/api/adapters/driven/persistence/contract`) deckt **drei** Ports
  (Sessions/Events/Sequencer), die anderen drei brauchen portierte
  Postgres-Tests. **Ausnahme von „kein Rewrite": der `ingest_sequencer`** —
  heute ein In-Process-RAM-Counter (`SELECT MAX(...)`-Seed + `atomic.Add`),
  der über N Replicas identische Werte vergibt (PK-Kollisionen); er muss
  DB-autoritativ werden (`nextval`/`IDENTITY` + `RETURNING`) — ein
  Port-Redesign, getrackt als **R-28**.
- Eine **Multi-Replica-Harness** (≥ 2 API-Instanzen hinter einem
  Load-Balancer, ein Postgres) plus ein **Scale-out-Lasttest** liefern
  die fehlende R-26-Evidenz: horizontale Durchsatz-Skalierung, kein
  Verlust/keine Duplikate über Replicas, `ingest_sequence`-Integrität
  unter Concurrent-Writern.
- Die Ausführung erfolgt **tranchenweise** über
  `plan-0.23.0-postgres-scaleout` mit eigenem Gate je Tranche; SQLite-
  Pfad bleibt während der gesamten Umsetzung unverändert grün.

## Begründung

**Trigger-Re-Evaluation (2026-06-17): kein harter ADR-0005-Trigger ist
gefeuert.** Trigger #1 (≥ 2 Replicas) hat kein konkretes Produktiv-
Requirement; Trigger #2 (RPO/RTO) ist nicht gestellt; Trigger #3
(Retention) ist gemessen *nicht* ausgelöst. Die Reaktivierung ist daher
eine **bewusst proaktive** Entscheidung, nicht eine Trigger-Reaktion —
aus drei Gründen:

1. **Evidenz statt Behauptung.** Die einzige verbleibende, unbelegte
   Architektur-Achse ist die horizontale Skalierung. „Single-Instance
   hält viel, Scale-out ist Postgres-Gebiet" ist heute eine *Behauptung*;
   der Adapter + Multi-Replica-Lasttest macht daraus einen *Nachweis*.
2. **De-Risking des künftigen Triggers.** Wenn Trigger #1 real wird (ein
   Betreiber braucht ≥ 2 Replicas), ist der Pfad dann erprobt und nicht
   ein Notfall-Umbau unter Druck.
3. **Die Architektur ist teilweise vorbereitet** — der Review
   (2026-06-17) hat hier zwei zu optimistische Annahmen korrigiert:
   hexagonale Driven-Ports und eine (für drei der sechs Ports) adapter-
   agnostische Contract-Suite sind echt da. **Aber**: `schema.yaml` ist
   **kein** fertiger Anker (nur V1-Baseline; das volle Schema V1–V7 ist
   Tranche-1-Portage), und der **`ingest_sequencer` ist doch ein Rewrite**
   — heute ein In-Process-RAM-Counter, der über Replicas PK-Kollisionen
   erzeugt und DB-autoritativ werden muss (**R-28**). „Kein Rewrite" gilt
   für fünf Ports, nicht für den Sequencer.

   **Zudem zwei Single-Instance-Blöcke, nicht einer.** (i) Der
   Persistenz-Store (dieses ADR). (ii) Der
   **Per-Projekt-Ingest-Limiter** ist in-process
   (`ratelimit.NewTokenBucketRateLimiter`, `main.go`) und hat **keine**
   Redis-Variante — Multi-Host-Redis-Backends existieren nur für den
   Origin- (R-22) und den Issuance-Limiter (R-17), **nicht** für den
   Ingest-Limiter aus R-26. Hinter N Replicas wäre die effektive
   Per-Projekt-Decke `N × Capacity` und die Fairness nicht repliken-
   übergreifend. Für den **Durchsatz-/Verlust-Scale-out-Nachweis (R-26 c,
   dieses ADR)** genügt der Store-Wechsel; ein echter **Multi-Tenant-
   Fairness-Nachweis (R-26 b)** braucht zusätzlich einen shared (Redis)
   Ingest-Limiter und wird in der Multi-Tenant-Arbeit gescopt, nicht hier.

**Ehrliche Kosten.** Ein zweiter Persistenz-Adapter heißt: doppelte
CI-Matrix (Contract-Suite gegen SQLite *und* Postgres), Postgres als
zusätzliche Test-Infrastruktur, und Wartung zweier Stores bei jeder
Schema-Migration. Diese Kosten werden in Kauf genommen, weil sie an einen
optionalen Pfad gebunden bleiben (Default unverändert SQLite) und der
Nachweis-Wert die laufende Last rechtfertigt.

## Nicht Entschieden / Grenzen

- **SQLite bleibt Default** und wird nicht ersetzt. Kein Zwang zu
  Postgres in der lokalen Standardumgebung.
- **Keine automatische SQLite→Postgres-Datenmigration** bestehender
  Läufe. Postgres startet mit frischem, aus `schema.yaml` generiertem
  Store; eine Datenübernahme wäre ein eigener Folge-Scope.
- **Kein Cloud-Managed-Postgres-Zwang**, keine Backup-/PITR-/Replikations-
  Topologie als Pflicht — die Harness belegt Scale-out, nicht eine
  konkrete Betriebs-Topologie. RPO/RTO (Trigger #2) bleibt separat.
- **Analytics-Backends (RAK-92)** bleiben unverändert deferred;
  **Kubernetes (RAK-93/NF-18)** bleibt unverändert optionaler Beispielpfad.
- **Multi-Tenant** (R-26 Teil b) ist orthogonal: das N-Projekt-Seeding +
  Token-Fan-out kann auf SQLite *und* Postgres laufen und ist nicht an
  diese Entscheidung gebunden.

## What Ändert Sich

- Neuer optionaler Adapter `apps/api/adapters/driven/persistence/postgres`.
- `apps/api/internal/storage`: `driverName`/DSN parametrisiert, Postgres-
  Migrations-Target aus `schema.yaml`.
- `MTRACE_PERSISTENCE=postgres` + DSN-ENV in `main.go`.
- Multi-Replica-Compose-Profil + Scale-out-Lasttest (Erweiterung von
  `scripts/smoke-load.sh`).
- Lastenheft-Patch: RAK-91-Status „defer" → „proceed, optional".

## What Bleibt Unverändert

- SQLite ist lokaler Default und Standard-Lab-Store.
- Compose bleibt die primäre Lab-Umgebung; Single-Instance ist der
  Default-Betrieb.
- Die ADR-0005-Trigger-Logik bleibt gültig; dieses ADR löst nur den
  Postgres-Teil aus „nicht entschieden" und dokumentiert die proaktive
  Reaktivierung.
- Tags/Releases bleiben human-approval-pflichtig.
