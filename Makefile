BINARY     := bin/bot
CMD        := ./cmd/bot
COMPOSE    := docker compose --env-file .env -f deployments/docker-compose.yml
GOOSE      := goose -dir db/migrations postgres "$(DATABASE_URL)"
SQLC       := cd db && sqlc generate

.PHONY: help up down migrate migrate-down migrate-status sqlc build build-check run test test-unit test-integration lint vet vuln tidy clean smoke-test rollback

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

build-check: ## Verify all packages compile and pass go vet
	go build ./...
	go vet ./...

run: ## Run the bot locally (requires .env to be sourced)
	go run $(CMD)

test: ## Run all tests (unit + integration)
	go test -v -race -timeout 120s -p 2 ./...

test-unit: ## Run unit tests only (skips integration tests)
	go test -v -race -short ./...

test-integration: ## Run integration tests only (requires DATABASE_URL and REDIS_URL)
	go test -v -race -tags integration -timeout 120s -p 2 ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

vet: ## Run go vet on all packages
	go vet ./...

vuln: ## Run govulncheck (install: go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...

tidy: ## Tidy go module dependencies
	go mod tidy

clean: ## Remove build artifacts
	rm -rf bin/

smoke-test: ## Run post-deploy smoke test on server (requires DEPLOY_USER, DEPLOY_HOST)
	ssh $(DEPLOY_USER)@$(DEPLOY_HOST) 'bash -s' < deployments/smoke-test.sh

rollback: ## Restore previous binary on server and restart service (requires DEPLOY_USER, DEPLOY_HOST)
	ssh $(DEPLOY_USER)@$(DEPLOY_HOST) \
		'cp /opt/moneytracker/bot.prev /opt/moneytracker/bot && \
		 sudo systemctl restart moneytracker'

.DEFAULT_GOAL := help
