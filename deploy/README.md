# `deploy/` — Deployment-Artefakte

Diese Struktur ist der Anker für reproduzierbare Deployment-
Artefakte (`F-7`). Sie ist **nicht** die laufende Lab-Umgebung —
das ist weiterhin
[`docker-compose.yml`](../docker-compose.yml) im Repo-Root, das
über `make dev` und `make dev-observability` aus
[`docs/user/local-development.md`](../docs/user/local-development.md)
gesteuert wird.

## Status (Stand `0.13.0`)

| Pfad | Status | Anmerkung |
| ---- | ------ | --------- |
| `deploy/compose/` | Reserviert | Aktuell leer; künftige Compose-Snippets, die nicht zum Lab-Default gehören (z. B. CI- oder Stand-alone-Deployments), landen hier. Der Lab-Default bleibt `docker-compose.yml` im Repo-Root. |
| `deploy/docker/` | Reserviert | Reserviert für Image-Build- bzw. Image-Veröffentlichungs-Artefakte, sobald Container-Images veröffentlicht werden (siehe [`docs/user/releasing.md`](../docs/user/releasing.md) — Container-Image-Veröffentlichung ist deferred). |
| `deploy/k8s/` | Beispiel | Optionale Kubernetes-Manifeste als `0.13.0`-Seed für `MVP-42`/`NF-18`. m-trace ist **nicht** für Production-K8s ausgelegt; die Manifeste sind ausdrücklich **kein** Production-Ready-Stand. |

## Was hier nicht hingehört

- Geheime Tokens, Credentials oder private URLs — die gehören in
  ein operatives Secret-Management außerhalb des Repos.
- Spike- oder Beispielcode, der besser unter
  [`examples/`](../examples/) aufgehoben ist.

## Bezug

- [`spec/lastenheft.md`](../spec/lastenheft.md) `F-7` (Pflicht-
  Repo-Struktur), `NF-18`/`MVP-42` (Kubernetes-Status).
- [`docs/user/local-development.md`](../docs/user/local-development.md)
  für den Lab-Default über `docker-compose.yml`.
- [`docs/user/releasing.md`](../docs/user/releasing.md) für den
  Release- und Tag-Prozess.
