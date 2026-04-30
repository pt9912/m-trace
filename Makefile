COMPOSE ?= docker compose

.PHONY: dev stop smoke test lint build coverage-gate arch-check

dev:
	$(COMPOSE) up --build

stop:
	$(COMPOSE) down

smoke:
	bash scripts/smoke-0.1.0.sh

test:
	$(MAKE) -C apps/api test

lint:
	$(MAKE) -C apps/api lint

build:
	$(MAKE) -C apps/api build

coverage-gate:
	$(MAKE) -C apps/api coverage-gate $(if $(THRESHOLD),THRESHOLD="$(THRESHOLD)")

arch-check:
	$(MAKE) -C apps/api arch-check
