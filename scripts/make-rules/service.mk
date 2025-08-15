##@ Project Code Operations
# The following commands are used for project code development operations.

.PHONY: dev-setup
dev-setup: ## Setup complete development environment: tools + generate
	@echo "🚀 Setting up development environment..."
	@$(MAKE) install-tools
	@$(MAKE) generate
	@echo "✅ Development environment ready!"

.PHONY: dev-clean
dev-clean: ## Clean development environment and cleanup
	@echo "🧹 Cleaning development environment..."
	@docker system prune -f
	@echo "✅ Development environment cleaned!"

.PHONY: dev-watch
dev-watch: ## Development mode with file watching and auto-rebuild
	@echo "👀 Starting development with file watching..."
	@$(MAKE) generate
	@air || (echo "Air not found, installing..."; go install github.com/cosmtrek/air@latest && air)

.PHONY: dev-quick
dev-quick: ## Quick development start (skip tool installation)
	@echo "⚡ Quick development start..."
	@$(MAKE) generate
	@$(MAKE) run-api

.PHONY: dev-test
dev-test: ## Run full test pipeline with coverage
	@echo "🧪 Running full test pipeline..."
	@go test -v -race ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

.PHONY: dev-bench
dev-bench: ## Run benchmarks in development environment
	@echo "📈 Running benchmarks..."
	@go test -v -bench=. -benchmem ./...