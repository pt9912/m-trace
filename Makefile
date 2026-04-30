COMPOSE ?= docker compose

.PHONY: dev stop smoke

dev:
	$(COMPOSE) up --build

stop:
	$(COMPOSE) down

smoke:
	bash scripts/smoke-0.1.0.sh
