COMPOSE ?= docker compose

.PHONY: dev dev-observability stop smoke smoke-observability smoke-rak10-console seed-rak9 browser-e2e test lint build coverage-gate arch-check sdk-performance-smoke

dev:
	$(COMPOSE) up --build

dev-observability:
	OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317 OTEL_EXPORTER_OTLP_PROTOCOL=grpc OTEL_TRACES_EXPORTER=otlp OTEL_METRICS_EXPORTER=otlp $(COMPOSE) --profile observability up --build

stop:
	$(COMPOSE) --profile observability down

smoke:
	bash scripts/smoke-0.1.1.sh

smoke-observability:
	bash scripts/smoke-observability.sh

smoke-rak10-console:
	OTEL_TRACES_EXPORTER=console $(COMPOSE) up -d --build api
	bash scripts/smoke-rak10-console.sh

seed-rak9:
	bash scripts/seed-rak9.sh

browser-e2e:
	bash scripts/test-browser-e2e.sh

test:
	$(MAKE) -C apps/api test
	pnpm run test

lint:
	$(MAKE) -C apps/api lint
	pnpm run lint

build:
	$(MAKE) -C apps/api build
	pnpm run build

coverage-gate:
	$(MAKE) -C apps/api coverage-gate $(if $(THRESHOLD),THRESHOLD="$(THRESHOLD)")
	pnpm --filter @npm9912/player-sdk run test:coverage
	pnpm --filter @npm9912/m-trace-dashboard run test:coverage
	pnpm --filter @npm9912/stream-analyzer run test:coverage

arch-check:
	$(MAKE) -C apps/api arch-check

sdk-performance-smoke:
	pnpm --filter @npm9912/player-sdk run performance:smoke
