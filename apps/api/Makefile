# m-trace API spike — Kotlin/Micronaut variant
# Docker-only workflow per docs/plan-spike.md §14.11.

IMAGE := m-trace-api-spike

.PHONY: test lint build deps compile run clean

test:
	docker build --target test -t $(IMAGE):micronaut-test .

lint:
	docker build --target lint -t $(IMAGE):micronaut-lint .

build:
	docker build --target runtime -t $(IMAGE):micronaut .

deps:
	docker build --no-cache --target deps -t $(IMAGE):micronaut-deps .

compile:
	docker build --target compile -t $(IMAGE):micronaut-compile .

run: build
	docker run --rm -p 8080:8080 --name m-trace-api-micronaut $(IMAGE):micronaut

clean:
	-docker rmi $(IMAGE):micronaut $(IMAGE):micronaut-test $(IMAGE):micronaut-lint $(IMAGE):micronaut-compile $(IMAGE):micronaut-deps
