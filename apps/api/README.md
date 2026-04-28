# m-trace API — Kotlin/Micronaut-Spike-Prototyp

Backend-Spike gemäß `docs/plan-spike.md` §6.3 und
`docs/spike/0001-backend-stack.md` §12.3.

## Workflow (Docker-only)

```bash
make test     # docker build --target test
make lint     # docker build --target lint  (Soll, detekt)
make build    # docker build --target runtime
make run      # docker run -p 8080:8080
```

Lokales JDK / Gradle ist **nicht erforderlich** — das Build-Image
`gradle:8.12-jdk21` bringt die Toolchain mit. Es wird **kein**
`gradle-wrapper.jar` versioniert (Plan §12.2).

## Verzeichnisstruktur

`hexagon/` und `adapters/` liegen direkt unter `apps/api/`,
flach und sichtbar (Plan §12.2). Custom Gradle `srcDirs` mappen
diese auf den Kotlin-Compiler.

```text
apps/api/
├── hexagon/
│   ├── domain/
│   ├── port/{driving,driven}/
│   └── application/
├── adapters/
│   ├── driving/http/
│   │   └── Application.kt
│   └── driven/{persistence,telemetry,metrics}/
├── resources/
├── test/
├── build.gradle.kts
├── gradle.properties
├── detekt.yml
├── Dockerfile
└── Makefile
```

## Status

Bootstrap. `/api/health` liefert `200 OK` über Micronaut.
Domain, Use Case, Tests und Pflicht-Endpunkte folgen in
nachfolgenden Commits.
