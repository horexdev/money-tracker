BINARY     := bin/bot
API_BINARY := bin/api
CMD        := ./cmd/bot
API_CMD    := ./cmd/api
COMPOSE    := docker compose --env-file .env -f deployments/docker-compose.yml
GOOSE_BIN  := $(shell which goose 2>/dev/null || echo "$(shell go env GOPATH)/bin/goose")
DATABASE_URL := $(shell grep -v '^\#' .env | grep '^DATABASE_URL=' | cut -d'=' -f2-)
GOOSE      := $(GOOSE_BIN) -dir db/migrations postgres "$(DATABASE_URL)"
SQLC       := cd db && sqlc generate

.PHONY: help up down migrate migrate-down migrate-status sqlc build build-api build-check run run-api web-dev web-build web-test test test-unit test-integration test-cover lint vet vuln tidy clean smoke-test rollback backup backup-status

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

build-api: ## Build the API server binary to bin/api
	@mkdir -p bin
	go build -o $(API_BINARY) $(API_CMD)

build-check: ## Verify all packages compile and pass go vet
	go build ./...
	go vet ./...

run: ## Run the bot locally
	@set -a && . ./.env && set +a && go run $(CMD)

run-api: ## Run the API server locally
	@set -a && . ./.env && set +a && go run $(API_CMD)

web-dev: ## Start Mini App dev server (proxies /api → localhost:8080)
	cd web && npm run dev

web-build: ## Build Mini App for production into web/dist/
	cd web && npm run build

web-test: ## Run Mini App test suite (vitest)
	cd web && npm run test

test: ## Run all tests (unit + integration)
	go test -v -race -timeout 120s -p 2 ./...

test-unit: ## Run unit tests only (skips integration tests)
	go test -v -race -short ./...

test-integration: ## Run integration tests only (requires DATABASE_URL and REDIS_URL)
	go test -v -race -tags integration -timeout 120s -p 2 ./...

test-cover: ## Run unit tests with coverage and print per-package summary
	go test -race -short -timeout 60s -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

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

backup: ## Trigger manual backup on server (requires DEPLOY_USER, DEPLOY_HOST)
	ssh $(DEPLOY_USER)@$(DEPLOY_HOST) 'sudo systemctl start moneytracker-backup.service'

backup-status: ## Show backup timer status and recent logs on server (requires DEPLOY_USER, DEPLOY_HOST)
	ssh $(DEPLOY_USER)@$(DEPLOY_HOST) \
		'systemctl list-timers moneytracker-backup.timer --no-pager; \
		 echo; journalctl -u moneytracker-backup.service --no-pager -n 30'

.DEFAULT_GOAL := help
