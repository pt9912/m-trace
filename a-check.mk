# a-check.mk — Architektur-Gate via a-check (digest-gepinnt), included im Root-Makefile.
# Basiert auf `a-check --print-mk`; Digest explizit auf den verifizierten
# v0.14.0-Stand gepinnt (a-checks --print-mk bettet noch v0.13.0 ein).
A_CHECK_IMAGE ?= ghcr.io/pt9912/a-check@sha256:f1b8ff5e9e9ab2007d2ba88527c97f070a30fb9fe08da78b20f4be6c6b5505ac

.PHONY: a-check a-check-graph
a-check: ## Architektur: Hexagon-Regeln via a-check (netzlos, read-only).
	docker run --rm --network none -v "$(CURDIR)":/src:ro $(A_CHECK_IMAGE) /src

a-check-graph: ## Architektur-Graph (Mermaid) aus .a-check.yml auf stdout (read-only, kein Scan).
	docker run --rm --network none -v "$(CURDIR)":/src:ro $(A_CHECK_IMAGE) --print-graph /src
