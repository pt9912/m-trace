# a-check.mk — Architektur-Gate via a-check (digest-gepinnt), included im Root-Makefile.
# Basiert auf `a-check --print-mk`; Digest explizit auf den verifizierten
# v0.20.0-Stand gepinnt (Release-Notes, a-check-Repo).
A_CHECK_IMAGE ?= ghcr.io/pt9912/a-check@sha256:e8208764b119c606c92f82722813386277a65b12812d23b6107ea7a14dc25da1

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
