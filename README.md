# MoneyTracker Bot

A personal finance Telegram bot built in Go. Track your income and expenses, view balance and category statistics — all without leaving Telegram.

## Features

- **Add expenses** with category and optional note (multi-step flow)
- **Add income** with category and optional note (multi-step flow)
- **Balance** — instant income / expense / net summary
- **Transaction history** — last 10 entries with category emoji
- **Stats by category** — for today, this week, this month, or last month
- **Redis FSM** — conversation state survives bot restarts (30-minute TTL)
- **Auto-registration** — users are registered on first message, no `/start` required

## Tech Stack

| Layer       | Library                    |
|-------------|----------------------------|
| Bot         | `go-telegram/bot`          |
| DB driver   | `jackc/pgx/v5`             |
| SQL codegen | `sqlc`                     |
| Migrations  | `pressly/goose/v3`         |
| Cache / FSM | `go-redis/redis/v9`        |
| Config      | `caarlos0/env/v11`         |
| Logging     | `log/slog` (stdlib)        |
| Tests       | `testify` + `testcontainers-go` |

## Requirements

- Go 1.23+
- Docker & Docker Compose (for local development)
- PostgreSQL 16
- Redis 7
- A Telegram Bot Token from [@BotFather](https://t.me/BotFather)

## Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/horexdev/money-tracker.git
cd money-tracker

# 2. Copy and fill in environment variables
cp .env.example .env
# Edit .env with your BOT_TOKEN, DATABASE_URL, REDIS_URL

# 3. Start PostgreSQL and Redis locally
make up

# 4. Apply database migrations
make migrate

# 5. Run the bot
make run
```

## Environment Variables

| Variable        | Required | Default         | Description                        |
|-----------------|----------|-----------------|------------------------------------|
| `BOT_TOKEN`     | yes      | —               | Telegram bot token from @BotFather |
| `DATABASE_URL`  | yes      | —               | PostgreSQL DSN                     |
| `REDIS_URL`     | yes      | —               | Redis URL (`redis://host:port`)    |
| `LOG_LEVEL`     | no       | `info`          | `debug` / `info` / `warn` / `error` |
| `MIGRATIONS_DIR`| no       | `db/migrations` | Path to goose migration files      |

## Bot Commands

| Command        | Description                          |
|----------------|--------------------------------------|
| `/start`       | Show welcome message and command list |
| `/help`        | Same as `/start`                     |
| `/addexpense`  | Record an expense (multi-step)       |
| `/addincome`   | Record income (multi-step)           |
| `/balance`     | Show current balance                 |
| `/history`     | Show last 10 transactions            |
| `/stats`       | View stats by category and period    |
| `/cancel`      | Cancel any active flow               |

## Development

```bash
make up            # Start Docker services
make migrate       # Apply goose migrations
make sqlc          # Regenerate sqlc query code (after editing db/queries/)
make build         # Build binary to bin/bot
make run           # Run bot locally
make test          # Run all tests (unit + integration)
make test-unit     # Run only unit tests (no Docker needed)
make test-cover    # Run unit tests with coverage and per-package summary
make web-test      # Run Mini App test suite (vitest)
make lint          # Run golangci-lint
make tidy          # go mod tidy
make down          # Stop Docker services
```

## Project Structure

```
├── cmd/bot/           Entry point — wires all components and starts the bot
├── internal/
│   ├── config/        Environment-based configuration (caarlos0/env)
│   ├── domain/        Core business models and domain errors
│   ├── fsm/           Redis-backed FSM for conversation state
│   ├── handler/       Telegram update handlers and middleware
│   ├── repository/    Database access layer (pgx + sqlc)
│   └── service/       Business logic (user, transaction, stats)
├── pkg/money/         Decimal-safe money arithmetic helpers
├── db/
│   ├── migrations/    Goose SQL migration files
│   ├── queries/       sqlc SQL query definitions
│   └── sqlc.yaml      sqlc code generation config
└── deployments/       Docker Compose and systemd service file
```

## Architecture

The project follows a layered clean architecture:

```
Handler → Service → Repository → Database
                ↓
             FSM Store (Redis)
```

- **Handlers** parse Telegram updates and delegate to services.
- **Services** contain business rules and orchestrate repositories.
- **Repositories** wrap `sqlc`-generated code and translate between DB and domain types.
- **FSM Store** persists multi-step conversation state in Redis with a sliding 30-minute TTL.
- Money amounts are stored as **integer cents** (e.g. $12.50 → `1250`) to avoid float precision issues.

## Deployment

### Server Setup (one-time)

```bash
# On the server
sudo useradd -r -s /bin/false -d /opt/moneytracker moneytracker
sudo mkdir -p /opt/moneytracker
sudo chown moneytracker:moneytracker /opt/moneytracker

# Install goose for running migrations during deploys
go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy and enable the systemd service
sudo cp deployments/moneytracker.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable moneytracker

# Create /opt/moneytracker/.env with production values
```

### GitHub Actions CI/CD

Push to `main` triggers:

1. **CI** (`.github/workflows/ci.yml`) — runs tests with Postgres and Redis service containers.
2. **CD** (`.github/workflows/cd.yml`) — builds a Linux amd64 binary, copies it to the server via SSH, runs migrations, and restarts the systemd service.

Required GitHub repository secrets:

| Secret           | Description                         |
|------------------|-------------------------------------|
| `DEPLOY_SSH_KEY` | Private SSH key for the deploy user |
| `DEPLOY_HOST`    | Server IP or hostname               |
| `DEPLOY_USER`    | SSH username on the server          |
| `BOT_TOKEN`      | Telegram bot token                  |
| `DATABASE_URL`   | Production PostgreSQL DSN           |
| `REDIS_URL`      | Production Redis URL                |

## Testing

```bash
# Backend
make test-unit         # unit tests (no external dependencies)
make test-cover        # unit tests + coverage profile (coverage.out)
make test-integration  # integration tests (requires DATABASE_URL + REDIS_URL)
make test              # everything (unit + integration)

# Frontend (Mini App)
cd web && npm run test            # one-shot vitest run
cd web && npm run test:watch      # watch mode
cd web && npm run test:coverage   # vitest + v8 coverage report into web/coverage/
make web-test                     # shorthand for `cd web && npm run test`
```

Integration tests are tagged with `//go:build integration` and require a running PostgreSQL + Redis. CI provisions both via service containers; locally use `make up` to start them.

Frontend tests use [Vitest](https://vitest.dev/) + [@testing-library/react](https://testing-library.com/docs/react-testing-library/intro/) with a jsdom environment. The shared setup lives in `web/src/test/`:

- `setup.ts` — global mocks (`window.Telegram`, `matchMedia`, `@tma.js/sdk-react`, `framer-motion`).
- `render.tsx` — `renderWithProviders` helper that wires `QueryClientProvider`, `MemoryRouter`, and an in-memory i18next instance.

### Local CI gate (run before opening a PR)

In strict order — stop on the first failure:

```bash
make lint
make vuln
make test-unit
make build-check
make test-cover
cd web && npm ci && npm run lint && npm run build && npm run test:coverage
```

## License

MIT
