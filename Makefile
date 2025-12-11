.PHONY: help build run test test-unit test-integration localstack-up localstack-down seed clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o bin/payment-api cmd/api/main.go

run: ## Run the application
	@echo "Starting Payment API with LocalStack configuration..."
	@export USE_LOCALSTACK=true && \
	export AWS_REGION=us-east-1 && \
	export AWS_ENDPOINT=http://localhost:4566 && \
	export PORT=8080 && \
	go run cmd/api/main.go

localstack-up: ## Start LocalStack
	docker-compose up -d
	@echo "Waiting for LocalStack to be ready..."
	@sleep 5

localstack-down: ## Stop LocalStack
	docker-compose down

seed: ## Seed initial data (run AFTER init-db)
	@echo "🌱 Seeding initial data..."
	chmod +x scripts/seed_data.sh
	./scripts/seed_data.sh

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	go test -v -count=1 ./tests/unit/...

test-integration: ## Run integration tests
	go test -v -count=1 ./tests/integration/...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf localstack-data/

setup: localstack-up ## Setup development environment
	@echo "Installing dependencies..."
	go mod download
	@echo "Waiting for LocalStack..."
	@sleep 5
	@echo "Setup complete!"

init-db: ## Initialize database tables (run once after localstack-up)
	@echo "📊 Initializing database tables..."
	@export USE_LOCALSTACK=true && \
	export AWS_REGION=us-east-1 && \
	export AWS_ENDPOINT=http://localhost:4566 && \
	go run scripts/init_tables.go

dev: ## Start full development environment
	@echo "🚀 Setting up development environment..."
	@echo "Step 1: Starting LocalStack..."
	@make localstack-up
	@echo ""
	@echo "Step 2: Initializing database..."
	@make init-db
	@echo ""
	@echo "Step 3: Seeding data..."
	@make seed
	@echo ""
	@echo "✅ Setup complete! Now run 'make run' to start the API"

