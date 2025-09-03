##@ Docker Template System
# ==============================================================================
# New Docker template-based service management (no docker-compose dependency)
# ==============================================================================

# Docker template testing variables
DOCKER_TEST_SERVICES := redis mysql mariadb etcd otelcol victorialogs prometheus
# 动态环境选择：支持 PROJ_ENVIRONMENT 变量控制环境配置
# 使用方式：PROJ_ENVIRONMENT=test make docker.redis.start
DOCKER_ENV_FILE = $(PROJ_ROOT_DIR)/manifests/env/env.$(or $(PROJ_ENVIRONMENT),dev)
# 注意：禁止引用 scripts/installation/versions.sh，只使用 manifests/env 配置
DOCKER_TEMPLATE_LIB_LOADER = source $(DOCKER_ENV_FILE) && \
							 source $(PROJ_ROOT_DIR)/scripts/installation/common.sh && \
							 source $(PROJ_ROOT_DIR)/scripts/installation/lib/common_lib.sh

##@ Docker Template - Individual Services
# ==============================================================================
# Individual service management using Docker templates
# ==============================================================================

.PHONY: docker.%.start
docker.%.start: ## Start a service using Docker templates (e.g., make docker.redis.start)
	@echo "===========> Starting $* service using Docker template"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && proj::docker::run_service_from_template "$*"

.PHONY: docker.%.stop
docker.%.stop: ## Stop a service using Docker templates (e.g., make docker.redis.stop)
	@echo "===========> Stopping $* service using Docker template"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && proj::docker::stop_service_from_template "$*"

.PHONY: docker.%.status
docker.%.status: ## Check service status using Docker templates (e.g., make docker.redis.status)
	@echo "===========> Checking $* service status using Docker template"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && proj::docker::check_service_status_from_template "$*"

.PHONY: docker.%.restart
docker.%.restart: ## Restart a service using Docker templates (e.g., make docker.redis.restart)
	@echo "===========> Restarting $* service using Docker template"
	@$(MAKE) docker.$*.stop
	@sleep 2
	@$(MAKE) docker.$*.start

.PHONY: docker.%.cleanup
docker.%.cleanup: ## Clean up service Docker scripts (e.g., make docker.redis.cleanup)
	@echo "===========> Cleaning up $* Docker template artifacts"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && proj::docker::cleanup_generated_scripts "$*"

##@ Docker Template - Batch Operations
# ==============================================================================
# Batch operations for multiple services
# ==============================================================================

.PHONY: docker.test-services.start
docker.test-services.start: ## Start all test services using Docker templates
	@echo "===========> Starting all test services using Docker templates"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && proj::docker::manage_services "start" $(DOCKER_TEST_SERVICES)

.PHONY: docker.test-services.stop
docker.test-services.stop: ## Stop all test services using Docker templates
	@echo "===========> Stopping all test services using Docker templates"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && proj::docker::manage_services "stop" $(DOCKER_TEST_SERVICES)

.PHONY: docker.test-services.status
docker.test-services.status: ## Check status of all test services using Docker templates
	@echo "===========> Checking status of all test services using Docker templates"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && proj::docker::manage_services "status" $(DOCKER_TEST_SERVICES)

.PHONY: docker.test-services.restart
docker.test-services.restart: ## Restart all test services using Docker templates
	@echo "===========> Restarting all test services using Docker templates"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && proj::docker::manage_services "restart" $(DOCKER_TEST_SERVICES)

##@ Docker Template - Network & Infrastructure
# ==============================================================================
# Network and infrastructure management
# ==============================================================================

.PHONY: docker.network.create
docker.network.create: ## Create project Docker network
	@echo "===========> Creating project Docker network"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && proj::docker::ensure_project_network

.PHONY: docker.network.info
docker.network.info: ## Show project Docker network information
	@echo "===========> Docker network information"
	@docker network ls | grep proj-network || echo "Project network not found"
	@docker network inspect proj-network 2>/dev/null | jq -r '.[] | {Name: .Name, Driver: .Driver, Containers: .Containers}' || true

.PHONY: docker.volumes.info
docker.volumes.info: ## Show project Docker volumes
	@echo "===========> Docker volumes information"
	@docker volume ls | grep proj- || echo "No project volumes found"

.PHONY: docker.containers.info
docker.containers.info: ## Show all project containers
	@echo "===========> Project containers information"
	@docker ps -a --filter name=proj- --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}" || echo "No project containers found"

##@ Docker Template - Development & Testing
# ==============================================================================
# Development and testing utilities
# ==============================================================================

.PHONY: docker.test.system
docker.test.system: ## Run Docker template system tests
	@echo "===========> Running Docker template system tests"
	@$(PROJ_ROOT_DIR)/scripts/installation/test-docker-templates.sh

.PHONY: docker.templates.list
docker.templates.list: ## List all available Docker templates
	@echo "===========> Available Docker templates"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && proj::template::list_templates docker

.PHONY: docker.scripts.list
docker.scripts.list: ## List all generated Docker scripts
	@echo "===========> Generated Docker scripts"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && proj::docker::list_generated_scripts

.PHONY: docker.scripts.cleanup-all
docker.scripts.cleanup-all: ## Clean up all generated Docker scripts
	@echo "===========> Cleaning up all generated Docker scripts"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && proj::docker::cleanup_generated_scripts

.PHONY: docker.test.redis-full
docker.test.redis-full: ## Full Redis test cycle (start -> status -> stop -> cleanup)
	@echo "===========> Running full Redis test cycle"
	@$(MAKE) docker.redis.start
	@sleep 5
	@$(MAKE) docker.redis.status
	@sleep 2
	@$(MAKE) docker.redis.stop
	@$(MAKE) docker.redis.cleanup
	@echo "===========> Redis test cycle completed successfully"

.PHONY: docker.test.mysql-full
docker.test.mysql-full: ## Full MySQL test cycle (start -> status -> stop -> cleanup)
	@echo "===========> Running full MySQL test cycle"
	@$(MAKE) docker.mysql.start
	@sleep 10
	@$(MAKE) docker.mysql.status
	@sleep 2
	@$(MAKE) docker.mysql.stop
	@$(MAKE) docker.mysql.cleanup
	@echo "===========> MySQL test cycle completed successfully"

.PHONY: docker.test.mariadb-full
docker.test.mariadb-full: ## Full MariaDB test cycle (start -> status -> stop -> cleanup)
	@echo "===========> Running full MariaDB test cycle"
	@$(MAKE) docker.mariadb.start
	@sleep 10
	@$(MAKE) docker.mariadb.status
	@sleep 2
	@$(MAKE) docker.mariadb.stop
	@$(MAKE) docker.mariadb.cleanup
	@echo "===========> MariaDB test cycle completed successfully"

##@ Docker Template - Specific Service Helpers
# ==============================================================================
# Service-specific helper commands
# ==============================================================================

.PHONY: docker.redis.connect
docker.redis.connect: ## Connect to Redis using redis-cli
	@echo "===========> Connecting to Redis"
	@docker exec -it proj-redis redis-cli

.PHONY: docker.mysql.connect
docker.mysql.connect: ## Connect to MySQL using mysql client
	@echo "===========> Connecting to MySQL"
	@docker exec -it proj-mysql mysql -u root -p

.PHONY: docker.mariadb.connect
docker.mariadb.connect: ## Connect to MariaDB using mariadb client
	@echo "===========> Connecting to MariaDB"
	@docker exec -it proj-mariadb mariadb -u root -p

.PHONY: docker.mysql.logs
docker.mysql.logs: ## Show MySQL container logs
	@echo "===========> MySQL container logs"
	@docker logs proj-mysql --tail 50

.PHONY: docker.mariadb.logs
docker.mariadb.logs: ## Show MariaDB container logs
	@echo "===========> MariaDB container logs"
	@docker logs proj-mariadb --tail 50

.PHONY: docker.redis.logs
docker.redis.logs: ## Show Redis container logs
	@echo "===========> Redis container logs"
	@docker logs proj-redis --tail 50

.PHONY: docker.victorialogs.ui
docker.victorialogs.ui: ## Open VictoriaLogs Web UI
	@echo "===========> Opening VictoriaLogs Web UI"
	@echo "VictoriaLogs Web UI: http://localhost:9428/select/vmui/"
	@open "http://localhost:9428/select/vmui/" 2>/dev/null || echo "Please open http://localhost:9428/select/vmui/ in your browser"

##@ Docker Template - Environment Management
# ==============================================================================
# Environment and configuration management
# ==============================================================================

.PHONY: docker.env.check
docker.env.check: ## Check Docker environment and prerequisites
	@echo "===========> Checking Docker environment"
	@command -v docker >/dev/null 2>&1 || (echo "❌ Docker not found" && exit 1)
	@docker info >/dev/null 2>&1 || (echo "❌ Docker daemon not running" && exit 1)
	@echo "✅ Docker is available and running"
	@docker --version
	@echo ""
	@echo "===========> Checking template system"
	@ls -la $(PROJ_ROOT_DIR)/scripts/installation/templates/docker/ | head -5
	@echo "✅ Template system is ready"

.PHONY: docker.env.versions
docker.env.versions: ## Show version information for all Docker services
	@echo "===========> Service versions from manifests/env configuration"
	@echo "使用环境配置文件: $(DOCKER_ENV_FILE)"
	@source $(DOCKER_ENV_FILE) && echo "✅ 环境配置已加载: $$PROJ_ENVIRONMENT (端口前缀: $$PROJ_ACCESS_PORT_PREFIX)" && \
	 echo "=== 第三方组件版本信息 ===" && \
	 echo "数据库组件:" && \
	 echo "  Redis:        $$REDIS_VERSION" && \
	 echo "  MariaDB:      $$MARIADB_VERSION" && \
	 echo "  MySQL:        $$MYSQL_VERSION" && \
	 echo "  MongoDB:      $$MONGODB_VERSION" && \
	 echo "分布式系统:" && \
	 echo "  etcd:         $$ETCD_VERSION" && \
	 echo "  Kafka:        $$KAFKA_VERSION" && \
	 echo "  Zookeeper:    $$ZOOKEEPER_VERSION" && \
	 echo "  Nacos:        $$NACOS_VERSION" && \
	 echo "监控与可观测性:" && \
	 echo "  Jaeger:       $$JAEGER_VERSION" && \
	 echo "  Prometheus:   $$PROMETHEUS_VERSION" && \
	 echo "  Grafana:      $$GRAFANA_VERSION" && \
	 echo "  AlertManager: $$ALERTMANAGER_VERSION" && \
	 echo "  OTEL Collector: $$OTEL_COLLECTOR_VERSION" && \
	 echo "  Pyroscope:    $$PYROSCOPE_VERSION" && \
	 echo "  VictoriaLogs: $$VICTORIALOGS_VERSION" && \
	 echo "  VictoriaMetrics: $$VICTORIAMETRICS_VERSION"

.PHONY: docker.debug.redis
docker.debug.redis: ## Debug Redis Docker template generation
	@echo "===========> Debugging Redis Docker template generation"
	@$(DOCKER_TEMPLATE_LIB_LOADER) && \
	 proj::docker::generate_service_scripts "redis" "/tmp/redis-debug" && \
	 echo "Generated scripts in /tmp/redis-debug:" && \
	 ls -la /tmp/redis-debug/ && \
	 echo "" && \
	 echo "Generated docker-run.sh content (first 20 lines):" && \
	 head -20 /tmp/redis-debug/docker-run.sh

##@ Docker Template - Help
# ==============================================================================
# Help and documentation
# ==============================================================================

.PHONY: docker.help.templates
docker.help.templates: ## Show help for Docker template system
	@echo "=== Docker Template System Help ==="
	@echo ""
	@echo "🐳 Individual Service Commands:"
	@echo "  make docker.redis.start      - Start Redis container"
	@echo "  make docker.redis.stop       - Stop Redis container"
	@echo "  make docker.redis.status     - Check Redis status"
	@echo "  make docker.redis.restart    - Restart Redis container"
	@echo "  make docker.redis.cleanup    - Clean up Redis scripts"
	@echo ""
	@echo "🌍 Environment Support:"
	@echo "  make docker.redis.start                     - Development environment (default)"
	@echo "  PROJ_ENVIRONMENT=test make docker.redis.start    - Test environment"
	@echo "  PROJ_ENVIRONMENT=prod make docker.redis.start    - Production environment"
	@echo "  make docker.env.versions                    - Show current environment versions"
	@echo ""
	@echo "📦 Batch Operations:"
	@echo "  make docker.test-services.start   - Start all test services"
	@echo "  make docker.test-services.stop    - Stop all test services"
	@echo "  make docker.test-services.status  - Check all services status"
	@echo ""
	@echo "🔧 Development & Testing:"
	@echo "  make docker.test.redis-full   - Full Redis test cycle"
	@echo "  make docker.test.mysql-full   - Full MySQL test cycle"
	@echo "  make docker.test.mariadb-full - Full MariaDB test cycle"
	@echo "  make docker.env.check         - Check Docker environment"
	@echo ""
	@echo "🌐 Network & Infrastructure:"
	@echo "  make docker.network.create    - Create project network"
	@echo "  make docker.containers.info   - Show project containers"
	@echo "  make docker.volumes.info      - Show project volumes"
	@echo ""
	@echo "📋 Available Services: $(DOCKER_TEST_SERVICES)"
	@echo "📁 Environment Files: manifests/env/env.{dev,test,prod}"
	@echo ""
	@echo "For more information, see: scripts/installation/templates/DOCKER_USAGE.md"