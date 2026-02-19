
APP_NAME=career
APP_PORT=8080

DB_HOST=career_db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=1234
DB_NAME=career
DB_DSN=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# --- Tools ---
MIGRATE=$(shell which migrate 2>/dev/null)



up: ## Start containers (local development with build)
	@echo "🚀 Starting Docker Compose (local build)..."
	docker compose up -d --build

up-prod: ## Start containers (production mode - uses registry image)
	@echo "🚀 Starting Docker Compose (production)..."
	docker compose up -d

down: ## Stop containers
	@echo "🛑 Stopping Docker Compose..."
	docker compose down

destroy: ## Remove everything (containers + volumes)
	@echo "🔥 Removing all Docker data..."
	docker compose down -v --remove-orphans

logs: ## Show logs for app
	@docker logs -f career_app

ps: ## Show running containers
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"


migrate-up: ## Run all migrations
	@if [ -z "$(MIGRATE)" ]; then echo "❌ migrate not installed. Install it with: go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; exit 1; fi
	migrate -path=migrations -database "$(DB_DSN)" up

migrate-down: ## Rollback last migration
	migrate -path=migrations -database "$(DB_DSN)" down 1

migrate-drop: ## Drop all tables
	migrate -path=migrations -database "$(DB_DSN)" drop

migrate-create: ## Create new migration file
	@read -p "Migration nomi: " name; \
	migrate create -ext sql -dir migrations $$name

build: ## Build the app binary
	@echo "🏗️ Building Go app..."
	go build -o bin/$(APP_NAME) ./app/main.go

run: ## Run the app locally
	@echo "▶️ Running app locally..."
	go run ./app/main.go

lint: ## Run linter
	golangci-lint run ./...

clean: ## Remove build artifacts
	rm -rf bin
	@echo "🧹 Clean done!"
