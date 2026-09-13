# a-check.mk — Architektur-Gate via a-check (digest-gepinnt), included im Root-Makefile.
# Basiert auf `a-check --print-mk`; Digest explizit auf den verifizierten
# v0.19.0-Stand gepinnt (a-checks --print-mk bettet noch den Vorgänger-Digest ein).
A_CHECK_IMAGE ?= ghcr.io/pt9912/a-check@sha256:34d3dfb50e44d99ea735186a35e1040589c4681dcfa2a51ed0f2aaea718cdd2d

# Container-Runtime ueber eine Indirektion, damit ein Repo mit podman/nerdctl
# oder einem docker-Wrapper nicht die Haelfte seiner Targets anders faehrt als
# die andere (slice-082).
#
# REIHENFOLGE ZAEHLT: `?=` setzt nur, wenn DOCKER noch nicht belegt ist.
# Wer eine eigene Runtime nutzt, definiert sie VOR dem `include` — oder
# hart (`DOCKER = podman`). Ein `DOCKER ?= podman` NACH dem
# include greift nicht mehr, weil dieses Fragment die Variable dann schon
# gesetzt hat.
DOCKER ?= docker

.PHONY: a-check a-check-graph
a-check: ## Architektur: Hexagon-Regeln via a-check (netzlos, read-only).
	$(DOCKER) run --rm --network none -v "$(CURDIR)":/src:ro $(A_CHECK_IMAGE) /src

a-check-graph: ## Architektur-Graph (Mermaid) aus .a-check.yml auf stdout (read-only, kein Scan).
	$(DOCKER) run --rm --network none -v "$(CURDIR)":/src:ro $(A_CHECK_IMAGE) --print-graph /src
