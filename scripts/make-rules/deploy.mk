##@ Deploy

# ==============================================================================
# Deploy (Consistent naming with tools.install.* prefix)
# ==============================================================================

.PHONY: deploy.install.%
deploy.install.%: ## Install a specified tool.
	@echo "===========> Installing $*"
	@$(MAKE) _install.$*


.PHONY: deploy.uninstall.%
deploy.uninstall.%: ## Uninstall a specified tool.
	@echo "===========> Uninstall talling $*"
	@$(MAKE) _uninstall.$*


##@ Redis Service
# ==============================================================================
# Redis installation methods
# ==============================================================================
.PHONY: _install.redis
_install.redis:  ## Install Redis for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::redis::install

.PHONY: _uninstall.redis
_uninstall.redis: ## Uninstall Redis for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::redis::uninstall

.PHONY: _install.docker.redis
_install.docker.redis: ## Install docker install redis
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::redis::docker::install


.PHONY: _uninstall.docker.mariadb
_uninstall.docker.redis: ## Uninstall docker install redis
	@$(PROJ_ROOT_DIR)/scripts/installation/mariadb.sh proj::mariadb::docker::uninstall


##@ MariaDB Service
# ==============================================================================
# MariaDB installation methods
# ==============================================================================
.PHONY: _install.mariadb
_install.mariadb:  ## Install Mariadb for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::mariadb::install

.PHONY: _uninstall.mariadb
_uninstall.mariadb: ## Uninstall Mariadb for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::mariadb::uninstall

.PHONY: _install.docker.mariadb
_install.docker.mariadb: ## Install docker install mariadb
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::mariadb::docker::install


.PHONY: _uninstall.docker.mariadb
_uninstall.docker.mariadb: ## Uninstall docker install Mariadb
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::mariadb::docker::uninstall

##@ MongoDB Service
# ==============================================================================
# MongoDB installation methods
# ==============================================================================
.PHONY: _install.mongo
_install.mongo:  ## Install mongo
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::mongo::install

.PHONY: _uninstall.mongo
_uninstall.mongo: ## Uninstall mongo
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::mongo::uninstall

.PHONY: _install.docker.mongo
_install.docker.mongo: ## Install docker install mongo
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::mongo::docker::install


.PHONY: _uninstall.docker.mongo
_uninstall.docker.mongo: ## Uninstall docker install mongo
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::mongo::docker::uninstall


##@ Kafka Service
# ==============================================================================
# Kafka installation methods
# ==============================================================================
.PHONY: _install.docker.kafka
_install.docker.kafka: ## Install docker install kafka
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::kafka::docker::install

.PHONY: _uninstall.docker.kafka
_uninstall.docker.kafka: ## Uninstall docker install kafka
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::kafka::docker::uninstall


##@ etcd Service
# ==============================================================================
# etcd installation methods
# ==============================================================================
.PHONY: _install.etcd
_install.etcd:  ## Install etcd for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::etcd::install

.PHONY: _uninstall.etcd
_uninstall.etcd: ## Uninstall etcd for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::etcd::uninstall

.PHONY: _install.docker.etcd
_install.docker.etcd: ## Install docker install etcd
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::etcd::docker::install

.PHONY: _uninstall.docker.etcd
_uninstall.docker.etcd: ## Uninstall docker install etcd
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::etcd::docker::uninstall


##@ Jaeger Service
# ==============================================================================
# Jaeger installation methods
# ==============================================================================
.PHONY: _install.jaeger
_install.jaeger:  ## Install Jaeger for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::jaeger::install

.PHONY: _uninstall.jaeger
_uninstall.jaeger: ## Uninstall Jaeger for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::jaeger::uninstall

.PHONY: _install.docker.jaeger
_install.docker.jaeger: ## Install docker install Jaeger
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::jaeger::docker::install

.PHONY: _uninstall.docker.jaeger
_uninstall.docker.jaeger: ## Uninstall docker install Jaeger
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::jaeger::docker::uninstall


##@ Grafana Service
# ==============================================================================
# Grafana installation methods
# ==============================================================================
.PHONY: _install.grafana
_install.grafana:  ## Install Grafana for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::grafana::install

.PHONY: _uninstall.grafana
_uninstall.grafana: ## Uninstall Grafana for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::grafana::uninstall

.PHONY: _install.docker.grafana
_install.docker.grafana: ## Install docker install Grafana
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::grafana::docker::install

.PHONY: _uninstall.docker.grafana
_uninstall.docker.grafana: ## Uninstall docker install Grafana
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::grafana::docker::uninstall


##@ Prometheus Service
# ==============================================================================
# Prometheus installation methods
# ==============================================================================
.PHONY: _install.prometheus
_install.prometheus:  ## Install Prometheus for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::prometheus::install

.PHONY: _uninstall.prometheus
_uninstall.prometheus: ## Uninstall Prometheus for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::prometheus::uninstall

.PHONY: _install.docker.prometheus
_install.docker.prometheus: ## Install docker install Prometheus
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::prometheus::docker::install

.PHONY: _uninstall.docker.prometheus
_uninstall.docker.prometheus: ## Uninstall docker install Prometheus
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::prometheus::docker::uninstall


##@ AlertManager Service
# ==============================================================================
# AlertManager installation methods
# ==============================================================================
.PHONY: _install.alertmanager
_install.alertmanager:  ## Install AlertManager for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::alertmanager::install

.PHONY: _uninstall.alertmanager
_uninstall.alertmanager: ## Uninstall AlertManager for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::alertmanager::uninstall

.PHONY: _install.docker.alertmanager
_install.docker.alertmanager: ## Install docker install AlertManager
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::alertmanager::docker::install

.PHONY: _uninstall.docker.alertmanager
_uninstall.docker.alertmanager: ## Uninstall docker install AlertManager
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::alertmanager::docker::uninstall


##@ OpenTelemetry Collector Service
# ==============================================================================
# OpenTelemetry Collector installation methods
# ==============================================================================
.PHONY: _install.otelcol
_install.otelcol:  ## Install OpenTelemetry Collector for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::otelcol::install

.PHONY: _uninstall.otelcol
_uninstall.otelcol: ## Uninstall OpenTelemetry Collector for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::otelcol::uninstall

.PHONY: _install.docker.otelcol
_install.docker.otelcol: ## Install docker install OpenTelemetry Collector
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::otelcol::docker::install

.PHONY: _uninstall.docker.otelcol
_uninstall.docker.otelcol: ## Uninstall docker install OpenTelemetry Collector
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::otelcol::docker::uninstall