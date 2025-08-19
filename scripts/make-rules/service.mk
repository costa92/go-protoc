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

##@ Logs Collection System
# Commands for managing the integrated log collection system (OpenTelemetry Collector + VictoriaLogs)

.PHONY: logs-start
logs-start: ## Start the complete log collection system (VictoriaLogs + OTel Collector)
	@echo "🚀 Starting log collection system..."
	@./scripts/installation/logs-collection.sh start

.PHONY: logs-stop  
logs-stop: ## Stop the log collection system
	@echo "🛑 Stopping log collection system..."
	@./scripts/installation/logs-collection.sh stop

.PHONY: logs-restart
logs-restart: ## Restart the log collection system
	@echo "🔄 Restarting log collection system..."
	@./scripts/installation/logs-collection.sh restart

.PHONY: logs-status
logs-status: ## Check log collection system status
	@echo "📊 Checking log collection system status..."
	@./scripts/installation/logs-collection.sh status

.PHONY: logs-info
logs-info: ## Display log collection system information
	@echo "ℹ️  Log collection system information..."
	@./scripts/installation/logs-collection.sh info

.PHONY: logs-test
logs-test: ## Send test logs to the collection system
	@echo "📤 Sending test logs..."
	@./scripts/installation/logs-collection.sh test-log

.PHONY: logs-view
logs-view: ## View logs from collection services
	@echo "📋 Viewing collection service logs..."
	@./scripts/installation/logs-collection.sh logs

.PHONY: logs-otel
logs-otel: ## View OpenTelemetry Collector logs only
	@echo "📋 Viewing OpenTelemetry Collector logs..."
	@./scripts/installation/logs-collection.sh logs otelcol

.PHONY: logs-vl
logs-vl: ## View VictoriaLogs service logs only
	@echo "📋 Viewing VictoriaLogs logs..."
	@./scripts/installation/logs-collection.sh logs victorialogs

##@ OTEL Collector Dynamic Management
# Commands for managing OTEL Collector with automatic project path detection

.PHONY: otel-start
otel-start: ## Start OTEL Collector with dynamic project path detection
	@echo "🚀 Starting OTEL Collector (dynamic path detection)..."
	@./scripts/otelcol-manager.sh start

.PHONY: otel-stop
otel-stop: ## Stop OTEL Collector
	@echo "🛑 Stopping OTEL Collector..."
	@./scripts/otelcol-manager.sh stop

.PHONY: otel-restart
otel-restart: ## Restart OTEL Collector with dynamic path detection
	@echo "🔄 Restarting OTEL Collector..."
	@./scripts/otelcol-manager.sh restart

.PHONY: otel-status
otel-status: ## Show OTEL Collector status and health
	@./scripts/otelcol-manager.sh status

.PHONY: otel-logs
otel-logs: ## View OTEL Collector logs (use make otel-logs ARGS="-f" for real-time)
	@./scripts/otelcol-manager.sh logs $(ARGS)

.PHONY: otel-health
otel-health: ## Check OTEL Collector health status
	@./scripts/otelcol-manager.sh health

.PHONY: otel-forward
otel-forward: ## Forward processed logs to VictoriaLogs
	@echo "📤 Forwarding logs to VictoriaLogs..."
	@./scripts/otelcol-manager.sh forward