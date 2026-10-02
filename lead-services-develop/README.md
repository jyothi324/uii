# 🎯 Lead Services

> A backend REST API server built with **Go**, **Gin**, and **GORM** — managing lead data with automated database setup, migrations, and coverage-gated unit tests.

---

## 🛠️ Tech Stack

| Layer | Technology |
|-------|-----------|
| **Language** | Go |
| **Web Framework** | [Gin](https://github.com/gin-gonic/gin) — fast HTTP router & middleware |
| **ORM** | [GORM](https://gorm.io) — migrations, queries, associations |
| **Live Reload** | [Air](https://github.com/air-verse/air) — hot reload on file changes |
| **Task Runner** | GNU Make — standardized dev/test/db commands |
| **Database** | PostgreSQL |
| **Schema** | `schema.sql` — raw SQL schema reference |

---

## 🗂️ Project Structure

```
lead-services/
├── .github/                 # GitHub Actions workflows / CI config
├── .vscode/                 # VS Code workspace settings
│
├── cmd/                     # Entry point — main.go (run & migrate-only modes)
├── config/                  # Loads .env, DB connection setup
├── db/                      # Migration scripts & seed SQL files
│
├── handlers/                # Gin route handlers (controllers)
├── middleware/              # Auth, logging, CORS middleware
├── mocks/                   # Mock implementations for unit testing
├── models/                  # GORM model structs & table definitions
├── routes/                  # Route registration
├── services/                # Business logic layer
├── store/                   # Data access layer (DB queries via GORM)
├── test/                    # Unit & integration tests
├── tmp/                     # Air build output (git-ignored)
├── utils/                   # Shared helper functions
│
├── .air.toml                # Air live-reload config
├── .env                     # Environment variables (git-ignored)
├── .env.example             # Environment variable template (committed)
├── .gitignore
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile                 # All dev/test/db commands
├── schema.sql               # Raw SQL schema reference
└── README.md
```

---

## 🌍 Environment Configuration

All environment-specific values live in `.env`. A safe template is committed as `.env.example` — **never commit `.env` with real secrets**.

```bash
cp .env.example .env
```

### `.env.example`

```env
# Server
APP_PORT=8080
APP_ENV=local          # local | qa | production

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=lead_db
```

The `Makefile` automatically loads and exports all variables from `.env`:

```makefile
include .env
export
```

So any `make` command has full access to `$(DB_HOST)`, `$(DB_NAME)`, etc. — no need to pass them manually.

---

## 🚀 Setup

Run the full setup in one command — creates the database if it doesn't exist, runs GORM migrations, and seeds data:

```bash
make setup
```

### What it does step by step

| Step | Command | Description |
|------|---------|-------------|
| 1 | `make db-create` | Checks if DB exists via `pg_database`; creates it only if missing |
| 2 | `make db-migrate` | Runs `go run cmd/main.go migrate-only` to apply GORM migrations |
| 3 | `make db-seed` | *(commented out by default)* Runs `db/setup.sql` seed file |

> **`db-create` is idempotent** — safe to re-run. It checks `pg_database` first and skips creation if the DB already exists.

---

## ▶️ Run

### Development — with live reload

```bash
air
```

Air watches all `.go` files and rebuilds automatically on save. Config lives in `.air.toml`.

### Without live reload

```bash
make run
# or directly:
go run cmd/main.go
```

### Run migrations only (no server start)

```bash
go run cmd/main.go migrate-only
```

### Build binary

```bash
go build -o bin/lead-services ./cmd
./bin/lead-services
```

---

## 🧪 Testing

Tests live in the `test/` folder and target the `handlers` package for coverage measurement.

### Quick reference

| Command | Description |
|---------|-------------|
| `make check` | Run tests + enforce ≥60% coverage threshold |
| `make coverage` | Run tests and generate `coverage.out` |
| `make covfunc` | Print per-function coverage % to terminal |
| `make covhtml` | Open visual HTML coverage report in browser |

---

### `make check` — coverage gate

Runs tests and **fails the build** if total coverage drops below **60%**:

```bash
make check
```

```
Total coverage: 74.3%
✅ Coverage check passed
```

```
Total coverage: 52.1%
❌ Coverage below 60%
```

> The threshold is set via `COVERAGE_THRESHOLD := 60` in the `Makefile`. Raise it as coverage improves.

---

### `make coverage` — generate report file

```bash
make coverage
# runs: go test ./test/... -coverpkg=./handlers -coverprofile=coverage.out
```

---

### `make covfunc` — per-function breakdown in terminal

```bash
make covfunc
# runs: go tool cover -func=coverage.out
```

Output example:

```
handlers/leads.go:25:   GetAll        85.7%
handlers/leads.go:61:   GetByID       100.0%
handlers/leads.go:90:   Create        66.7%
total:                  (statements)  74.3%
```

---

### `make covhtml` — visual line-by-line report

```bash
make covhtml
# runs: go tool cover -html=coverage.out
```

Opens `coverage.out` as an interactive HTML page in your default browser — green lines are covered, red lines are not.

---

### Run tests directly

```bash
# Run all tests in test/
go test ./test/...

# Verbose output
go test ./test/... -v

# Run a specific test by name
go test ./test/... -run TestLeadHandler_GetAll

# With race condition detector
go test ./test/... -race
```

---

### Example unit test (Gin handler + mock pattern)

```go
// test/leads_test.go
package test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestGetAllLeads_Success(t *testing.T) {
    mockService := new(mocks.MockLeadService)
    mockService.On("GetAll").Return([]models.Lead{
        {ID: 1, Name: "Alice"},
        {ID: 2, Name: "Bob"},
    }, nil)

    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.GET("/leads", handlers.GetAll(mockService))

    req := httptest.NewRequest(http.MethodGet, "/leads", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    mockService.AssertExpectations(t)
}
```

---

## 🐳 Docker

```bash
# Build image
docker build -t lead-services .

# Run container
docker run -d -p 8080:8080 \
  --env-file .env \
  --name lead-services-container \
  lead-services

# Stop and remove
docker stop lead-services-container && docker rm lead-services-container
```

---

## 📋 All Make Commands

| Command | Description |
|---------|-------------|
| `make setup` | Full setup: create DB + migrate + seed |
| `make db-create` | Create PostgreSQL database (skips if already exists) |
| `make db-migrate` | Run GORM auto-migrations |
| `make run` | Start the server (`go run cmd/main.go`) |
| `make clean` | Run `go clean` |
| `make coverage` | Run tests → generate `coverage.out` |
| `make covfunc` | Print per-function coverage % to terminal |
| `make covhtml` | Open visual HTML coverage report |
| `make check` | Run tests + fail if coverage < 60% |

---

## 🔌 Key Libraries

| Package | Purpose |
|---------|---------|
| `github.com/gin-gonic/gin` | HTTP framework — routing, middleware, JSON binding |
| `gorm.io/gorm` | ORM — model definitions, auto-migrate, queries |
| `gorm.io/driver/postgres` | GORM PostgreSQL driver |
| `github.com/air-verse/air` | Live reload — rebuilds on `.go` file changes |
| `github.com/stretchr/testify` | Test assertions (`assert`) and mocking |
| `github.com/joho/godotenv` | Loads `.env` into environment on startup |

---

## 🤝 Contributing

1. Fork this repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Write tests for your changes in `test/`
4. Run `make check` — ensure coverage stays above 60%
5. Commit: `git commit -m 'add: your feature'`
6. Push and open a Pull Request

---

> _Built for quality. Tested for confidence._
