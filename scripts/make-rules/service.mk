##@ Service
# The following commands are used to manage dependent services.

# Database Services
.PHONY: run-redis
run-redis: ## Run Redis service using docker-compose.
	@./scripts/installation/service.sh start redis

.PHONY: stop-redis
stop-redis: ## Stop and remove Redis service.
	@./scripts/installation/service.sh stop redis

.PHONY: run-mariadb
run-mariadb: ## Run MariaDB service using docker-compose.
	@./scripts/installation/service.sh start mariadb

.PHONY: stop-mariadb
stop-mariadb: ## Stop and remove MariaDB service.
	@./scripts/installation/service.sh stop mariadb

.PHONY: run-mongodb
run-mongodb: ## Run MongoDB service using docker-compose.
	@./scripts/installation/service.sh start mongodb

.PHONY: stop-mongodb
stop-mongodb: ## Stop and remove MongoDB service.
	@./scripts/installation/service.sh stop mongodb

# Messaging Services
.PHONY: run-kafka
run-kafka: ## Run Kafka service using docker-compose.
	@./scripts/installation/service.sh start kafka

.PHONY: stop-kafka
stop-kafka: ## Stop and remove Kafka service.
	@./scripts/installation/service.sh stop kafka

# Distributed Services
.PHONY: run-etcd
run-etcd: ## Run etcd service using installation script.
	@./scripts/installation/service.sh start etcd

.PHONY: stop-etcd
stop-etcd: ## Stop and remove etcd service.
	@./scripts/installation/service.sh stop etcd

# Observability Services
.PHONY: run-jaeger
run-jaeger: ## Run Jaeger service using docker-compose.
	@./scripts/installation/service.sh start jaeger

.PHONY: stop-jaeger
stop-jaeger: ## Stop and remove Jaeger service.
	@./scripts/installation/service.sh stop jaeger

.PHONY: run-prometheus
run-prometheus: ## Run Prometheus service using installation script.
	@./scripts/installation/service.sh start prometheus

.PHONY: stop-prometheus
stop-prometheus: ## Stop and remove Prometheus service.
	@./scripts/installation/service.sh stop prometheus

.PHONY: run-grafana
run-grafana: ## Run Grafana service using installation script.
	@./scripts/installation/service.sh start grafana

.PHONY: stop-grafana
stop-grafana: ## Stop and remove Grafana service.
	@./scripts/installation/service.sh stop grafana

.PHONY: run-alertmanager
run-alertmanager: ## Run AlertManager service using installation script.
	@./scripts/installation/service.sh start alertmanager

.PHONY: stop-alertmanager
stop-alertmanager: ## Stop and remove AlertManager service.
	@./scripts/installation/service.sh stop alertmanager

.PHONY: run-otelcol
run-otelcol: ## Run OpenTelemetry Collector service using installation script.
	@./scripts/installation/service.sh start otelcol

.PHONY: stop-otelcol
stop-otelcol: ## Stop and remove OpenTelemetry Collector service.
	@./scripts/installation/service.sh stop otelcol

.PHONY: run-victorialogs
run-victorialogs: ## Run VictoriaLogs service using installation script.
	@./scripts/installation/service.sh start victorialogs

.PHONY: stop-victorialogs
stop-victorialogs: ## Stop and remove VictoriaLogs service.
	@./scripts/installation/service.sh stop victorialogs

# Service Groups
.PHONY: start-all
start-all: ## Start all dependent services.
	@./scripts/installation/service.sh start all

.PHONY: stop-all
stop-all: ## Stop and remove all dependent services.
	@./scripts/installation/service.sh stop all

.PHONY: restart-all
restart-all: ## Restart all dependent services.
	@./scripts/installation/service.sh restart all

.PHONY: start-database
start-database: ## Start all database services (redis, mariadb, mongodb).
	@./scripts/installation/service.sh start database

.PHONY: stop-database
stop-database: ## Stop all database services.
	@./scripts/installation/service.sh stop database

.PHONY: start-observability
start-observability: ## Start all observability services.
	@./scripts/installation/service.sh start observability

.PHONY: stop-observability
stop-observability: ## Stop all observability services.
	@./scripts/installation/service.sh stop observability

# Service Status and Logs
.PHONY: status-all
status-all: ## Check status of all services.
	@./scripts/installation/service.sh status all

.PHONY: logs-all
logs-all: ## Show logs for all services (docker-compose services only).
	@echo "Note: Logs command only works for docker-compose services"
	@./scripts/installation/service.sh logs redis || true
	@./scripts/installation/service.sh logs kafka || true
	@./scripts/installation/service.sh logs jaeger || true
	@./scripts/installation/service.sh logs mariadb || true
	@./scripts/installation/service.sh logs mongodb || true

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