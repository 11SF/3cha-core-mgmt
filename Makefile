SHELL := /bin/bash
.DEFAULT_GOAL := run

BINARY_NAME := portal-backend

.PHONY: run
run: ## run the application
	go run .

.PHONY: build
build: ## build binary
	go build -ldflags "-X main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo dev)" -o $(BINARY_NAME) .

.PHONY: mod
mod: ## tidy go modules
	go mod tidy

.PHONY: test
test: ## run tests
	go test -v -race ./...

.PHONY: lint
lint: ## run linter
	golangci-lint run --fix

.PHONY: docker-up
docker-up: ## start postgres with docker compose
	docker compose up -d

.PHONY: docker-down
docker-down: ## stop docker compose
	docker compose down
