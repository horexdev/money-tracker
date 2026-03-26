BINARY     := bin/bot
CMD        := ./cmd/bot
COMPOSE    := docker compose -f deployments/docker-compose.yml
GOOSE      := goose -dir db/migrations postgres "$(DATABASE_URL)"
SQLC       := cd db && sqlc generate

.PHONY: help up down migrate migrate-down migrate-status sqlc build run test test-unit lint tidy clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Start local Docker services (postgres + redis)
	$(COMPOSE) up -d

down: ## Stop local Docker services
	$(COMPOSE) down

migrate: ## Apply all pending goose migrations
	$(GOOSE) up

migrate-down: ## Roll back the last migration
	$(GOOSE) down

migrate-status: ## Show migration status
	$(GOOSE) status

sqlc: ## Regenerate sqlc query code
	$(SQLC)

build: ## Build the bot binary to bin/bot
	@mkdir -p bin
	go build -o $(BINARY) $(CMD)

run: ## Run the bot locally (requires .env to be sourced)
	go run $(CMD)

test: ## Run all tests (unit + integration)
	go test -v -race -timeout 120s -p 2 ./...

test-unit: ## Run unit tests only (skips integration tests)
	go test -v -race -short ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

tidy: ## Tidy go module dependencies
	go mod tidy

clean: ## Remove build artifacts
	rm -rf bin/

.DEFAULT_GOAL := help
