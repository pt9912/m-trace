# syntax=docker/dockerfile:1.7

# ---------------------------------------------------------------------------
# m-trace API spike — Kotlin/Micronaut variant
#
# Docker-only workflow per docs/plan-spike.md §14.11. The repository ships
# no Gradle wrapper jar; the gradle:8.12-jdk21 build image brings Gradle.
# All build/test/lint steps run via:
#   docker build --target <stage> -t m-trace-api-spike:micronaut-<stage> .
#
# Stages:
#   deps     - resolves Gradle dependencies (cache layer)
#   compile  - compiles classes (fast feedback)
#   lint     - runs detekt (Soll, plan §14.9)
#   test     - runs `gradle test`
#   build    - produces the install distribution
#   runtime  - eclipse-temurin JRE running the install distribution
# ---------------------------------------------------------------------------

# ---- deps ------------------------------------------------------------------
FROM gradle:8.12-jdk21 AS deps

WORKDIR /src
COPY --chown=gradle:gradle settings.gradle.kts build.gradle.kts gradle.properties detekt.yml ./

RUN gradle --no-daemon resolveAllDependencies

# ---- compile ---------------------------------------------------------------
FROM deps AS compile

COPY --chown=gradle:gradle . .
RUN gradle --no-daemon classes

# ---- lint ------------------------------------------------------------------
FROM compile AS lint

RUN gradle --no-daemon detekt

# ---- test ------------------------------------------------------------------
FROM compile AS test

RUN gradle --no-daemon test

# ---- build -----------------------------------------------------------------
FROM compile AS build

RUN gradle --no-daemon installDist

# ---- runtime ---------------------------------------------------------------
FROM eclipse-temurin:21-jre-alpine AS runtime

RUN addgroup -S app && adduser -S app -G app

COPY --from=build --chown=app:app /src/build/install/api /opt/api

USER app
EXPOSE 8080
ENTRYPOINT ["/opt/api/bin/api"]
