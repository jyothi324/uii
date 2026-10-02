# Load .env variables (optional include: absent on a fresh clone until `make setup-all` creates one)
-include .env
export

# Re-read DB_PASSWORD via shell so # in the value isn't treated as a Make comment
DB_PASSWORD := $(shell grep -E '^DB_PASSWORD=' .env | cut -d'=' -f2- | sed 's/^"//;s/"$$//' | sed "s/^'//;s/'$$//")

# Space-separated tenant names — each becomes database tenant_<name>
# Override at the command line: make setup-all TENANTS="foo bar baz"
TENANTS ?= app tampa miami acergranite

COVERAGE_THRESHOLD := 75

.PHONY: all setup setup-all db-create db-migrate db-seed run clean coverage covfunc covhtml check

all: setup

setup: db-create db-migrate db-seed
	@echo "✅ Setup complete"

# Interactive: pick a profile (local/dev/prod), confirm its DB connection details, then run
# `setup` for every tenant in TENANTS against that profile. See .claude/skills/run/SKILL.md.
setup-all:
	@profile=""; \
	while [ -z "$$profile" ]; do \
		echo ""; \
		echo "Select environment to set up:"; \
		echo "  1) local"; \
		echo "  2) dev"; \
		echo "  3) prod"; \
		read -p "Enter choice [1-3]: " choice; \
		case "$$choice" in \
			1) profile="local" ;; \
			2) profile="dev" ;; \
			3) profile="prod" ;; \
			*) echo "❌ Invalid choice, try again." ;; \
		esac; \
	done; \
	env_file=".env.$$profile"; \
	if [ ! -f "$$env_file" ]; then \
		echo "❌ $$env_file not found in $$(pwd) — create it first."; \
		exit 1; \
	fi; \
	db_host=$$(grep -E '^DB_HOST=' "$$env_file" | cut -d'=' -f2-); \
	db_port=$$(grep -E '^DB_PORT=' "$$env_file" | cut -d'=' -f2-); \
	db_user=$$(grep -E '^DB_USER=' "$$env_file" | cut -d'=' -f2-); \
	db_name=$$(grep -E '^DB_NAME=' "$$env_file" | cut -d'=' -f2-); \
	db_ssl=$$(grep -E '^DB_SSL_MODE=' "$$env_file" | cut -d'=' -f2-); \
	db_pass_raw=$$(grep -E '^DB_PASSWORD=' "$$env_file" | cut -d'=' -f2-); \
	if [ -n "$$db_pass_raw" ]; then db_pass_display="(set, hidden)"; else db_pass_display="(empty)"; fi; \
	echo ""; \
	echo "==> Profile:     $$profile   ($$env_file)"; \
	echo "==> DB_HOST:     $$db_host"; \
	echo "==> DB_PORT:     $$db_port"; \
	echo "==> DB_USER:     $$db_user"; \
	echo "==> DB_NAME:     $$db_name   (base name — tenant DBs are tenant_<name>)"; \
	echo "==> DB_SSL_MODE: $$db_ssl"; \
	echo "==> DB_PASSWORD: $$db_pass_display"; \
	echo "==> Tenants:     $(TENANTS)"; \
	echo ""; \
	read -p "Type 'ok' to run setup for '$$profile' with the values above: " confirm; \
	if [ "$$confirm" != "ok" ]; then \
		echo "Aborted — no changes made."; \
		exit 1; \
	fi; \
	if [ "$$profile" = "prod" ]; then \
		read -p "This is PRODUCTION. Type 'prod' again to confirm: " confirm2; \
		if [ "$$confirm2" != "prod" ]; then \
			echo "Aborted — no changes made."; \
			exit 1; \
		fi; \
	fi; \
	cp "$$env_file" .env; \
	echo "✅ Copied $$env_file -> .env"; \
	for t in $(TENANTS); do \
		echo ""; \
		echo "==> Setting up: tenant_$$t"; \
		$(MAKE) --no-print-directory setup DB_NAME=tenant_$$t || exit 1; \
	done; \
	echo ""; \
	echo "✅ All $(words $(TENANTS)) tenant(s) set up for profile '$$profile'"

db-create:
	@if PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -tAc \
		"SELECT 1 FROM pg_database WHERE datname='$(DB_NAME)'" 2>/dev/null | grep -q 1; then \
		echo "✅ Database '$(DB_NAME)' already exists"; \
	else \
		echo "🚧 Creating database '$(DB_NAME)'..."; \
		PGPASSWORD=$(DB_PASSWORD) createdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME); \
		echo "✅ Database created"; \
	fi

db-migrate:
	@echo "🚀 Running GORM migrations..."
	go run cmd/main.go migrate-only

db-seed:
	@if [ -n "$(SQL_FILE)" ] && [ -f "$(SQL_FILE)" ]; then \
		echo "🌱 Inserting seed data from $(SQL_FILE)..."; \
		PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f $(SQL_FILE); \
		echo "✅ Seed complete"; \
	else \
		echo "⏭️  No seed file found, skipping"; \
	fi

run:
	@echo "🟢 Starting application..."
	go run cmd/main.go

clean:
	@echo "🧹 Cleaning up..."
	go clean
	@rm -f coverage.out

coverage:
	go test ./test/... -coverpkg=./handlers -coverprofile=coverage.out

covfunc: coverage
	go tool cover -func=coverage.out

covhtml: coverage
	go tool cover -html=coverage.out

check:
	go test ./test/... -coverpkg=./handlers -coverprofile=coverage.out
	@coverage=$$(go tool cover -func=coverage.out | tail -1 | awk '{gsub(/%/,"",$$3); print $$3}'); \
	echo "Total coverage: $$coverage%"; \
	if [ $$(printf "%.0f\n" $$coverage) -lt $(COVERAGE_THRESHOLD) ]; then \
		echo "❌ Coverage below $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "✅ Coverage check passed"; \
	fi
