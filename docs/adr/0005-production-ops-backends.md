# 0005 — Production- und Ops-Backends als optionale Seeds

> **Status**: Accepted  
> **Datum**: 2026-05-12  
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)  
> **Bezug**: `spec/lastenheft.md` NF-18, MVP-40..MVP-44,
> RAK-91..RAK-95; `docs/planning/done/plan-0.13.0.md`;
> `docs/planning/in-progress/risks-backlog.md` R-9.

## Kontext

`0.13.0` bewertet Production-/Ops-nahe Folgepunkte. Die bestehende
lokale Standardumgebung soll stabil bleiben: SQLite ist der Default-
Store, Compose ist das Lab, und Release-Veröffentlichung braucht eine
explizite menschliche Freigabe.

## Entscheidung

`0.13.0` liefert einen Decision-and-Seed-Scope:

- PostgreSQL wird nicht als Laufzeitadapter eingeführt. Der bestehende
  neutrale `schema.yaml`-Pfad bleibt der Migrationsanker; ein Postgres-
  Adapter wird erst bei Multi-Instance-, Recovery- oder Retention-
  Triggern umgesetzt.
- ClickHouse, VictoriaMetrics und Mimir werden nicht als Pflichtbackend
  eingeführt. Der Bedarf bleibt POC-gesteuert, bis Query-Workloads und
  Datenvolumen die Zusatzkomplexität rechtfertigen.
- Kubernetes bleibt ein optionaler Beispielpfad unter `deploy/k8s/`.
  Die Manifeste sind keine Production-Ready-Zusage.
- Der Devcontainer wird als reproduzierbare Entwicklungs-Seed-
  Konfiguration geliefert.
- Release-Automatisierung startet mit einem lokalen Guard-Script. Tags
  und Veröffentlichungen bleiben manuell freigabepflichtig.

## Begründung

Die aktuellen Hot-Path- und Operator-Anforderungen rechtfertigen keine
neue Pflichtabhängigkeit in der lokalen Standardumgebung. SQLite deckt
Restart-Durability und lokale Diagnose ab. Ein vollwertiger Postgres-
Adapter würde Migrations-, Replay-, Backup- und Dual-Read-Pfade
berühren, ohne dass `0.13.0` bereits konkrete SLA- oder Multi-Replica-
Schwellen liefert.

Analytics-Backends sind erst sinnvoll, wenn die Queries nicht mehr durch
die bestehenden API-/Prometheus-Pfade abgedeckt werden. Ein früher
Pflichtpfad würde Betriebskosten und Failure-Modes erhöhen.

Kubernetes-Artefakte sind als Beispiel hilfreich, aber R-9 verhindert,
dass eine K8s-Smoke-Stage stillschweigend die Compose-spezifische
Observability-Allowlist erweitert.

## Nicht Entschieden

- Kein Postgres-Runtime-Adapter und keine automatische SQLite-zu-
  Postgres-Migration.
- Kein ClickHouse-, VictoriaMetrics- oder Mimir-POC im Release-Default.
- Kein Production-Ingress, TLS, Secret-Management, HPA oder Network-
  Policy-Scope für Kubernetes.
- Keine automatische GitHub-Release-Erstellung ohne human approval.

## Trigger

Postgres wird reaktiviert, wenn mindestens einer dieser Trigger
erreicht wird:

- zwei oder mehr API-Replicas müssen denselben Store ohne shared-volume
  SQLite betreiben;
- Recovery-Ziel `RPO <= 15 min` oder `RTO <= 30 min` wird verbindlich;
- Retention-Queries über mehr als 10 Millionen Events brauchen stabile
  p95-Latenz unter 2 Sekunden.

Analytics-POC wird reaktiviert, wenn mindestens einer dieser Trigger
erreicht wird:

- aggregierte Playback-/SRT-/WebRTC-Abfragen über mehr als 50 Millionen
  Events pro Tag werden produktnah benötigt;
- Prometheus-Metriken reichen für geforderte Ad-hoc-Analysen nicht mehr;
- Kosten-/Operativvergleich benennt einen klaren Owner und ein
  Abbruchdatum von maximal 30 Tagen.

## What Aendert Sich

- `deploy/k8s/` enthält optionale Beispielmanifeste.
- `.devcontainer/devcontainer.json` pinnt Go, Node und Docker-outside-
  of-Docker für reproduzierbare Entwicklung.
- `scripts/release-guard.sh` prüft Freigabe, Version, Tag-Existenz und
  Release-Anker vor Tag/Publish.

## What Bleibt Unveraendert

- SQLite bleibt lokaler Default.
- Compose bleibt die primäre Lab-Umgebung.
- K8s, Postgres und Analytics-Backends sind keine lokalen
  Pflichtabhängigkeiten.
- Tags und GitHub-Releases werden nicht ohne explizite Freigabe
  veröffentlicht.
