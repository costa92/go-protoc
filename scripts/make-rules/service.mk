##@ Service
# The following commands are used to manage dependent services.

.PHONY: run-jaeger
run-jaeger: ## Run Jaeger service using docker-compose.
	@./scripts/installation/service.sh start jaeger

.PHONY: stop-jaeger
stop-jaeger: ## Stop and remove Jaeger service.
	@./scripts/installation/service.sh stop jaeger

.PHONY: run-redis
run-redis: ## Run Redis service using docker-compose.
	@./scripts/installation/service.sh start redis

.PHONY: stop-redis
stop-redis: ## Stop and remove Redis service.
	@./scripts/installation/service.sh stop redis

.PHONY: run-kafka
run-kafka: ## Run Kafka service using docker-compose.
	@./scripts/installation/service.sh start kafka

.PHONY: stop-kafka
stop-kafka: ## Stop and remove Kafka service.
	@./scripts/installation/service.sh stop kafka

.PHONY: start-all
start-all: ## Start all dependent services.
	@./scripts/installation/service.sh start all

.PHONY: stop-all
stop-all: ## Stop and remove all dependent services.
	@./scripts/installation/service.sh stop all

##@ Development Pipeline
# The following commands provide a complete development pipeline workflow.

.PHONY: dev-setup
dev-setup: ## Setup complete development environment: tools + services + generate
	@echo "🚀 Setting up development environment..."
	@$(MAKE) install-tools
	@$(MAKE) start-all
	@$(MAKE) generate
	@echo "✅ Development environment ready!"

.PHONY: dev-clean
dev-clean: ## Clean development environment: stop services and cleanup
	@echo "🧹 Cleaning development environment..."
	@$(MAKE) stop-all
	@docker system prune -f
	@echo "✅ Development environment cleaned!"

.PHONY: dev-restart
dev-restart: dev-clean dev-setup ## Fresh restart of complete development environment

.PHONY: dev-watch
dev-watch: ## Development mode with file watching and auto-rebuild
	@echo "👀 Starting development with file watching..."
	@$(MAKE) start-all
	@$(MAKE) generate
	@air || (echo "Air not found, installing..."; go install github.com/cosmtrek/air@latest && air)

.PHONY: dev-quick
dev-quick: ## Quick development start (skip tool installation)
	@echo "⚡ Quick development start..."
	@$(MAKE) run-redis
	@$(MAKE) generate
	@$(MAKE) run-api

.PHONY: dev-test
dev-test: ## Run full test pipeline with coverage
	@echo "🧪 Running full test pipeline..."
	@$(MAKE) start-all
	@go test -v -race ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

.PHONY: dev-bench
dev-bench: ## Run benchmarks in development environment
	@echo "📈 Running benchmarks..."
	@$(MAKE) start-all
	@go test -v -bench=. -benchmem ./...