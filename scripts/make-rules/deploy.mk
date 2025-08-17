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

##@ MySQL Service
# ==============================================================================
# MySQL installation methods
# ==============================================================================
.PHONY: _install.mysql
_install.mysql:  ## Install MySQL for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/mysql.sh proj::mysql::install

.PHONY: _uninstall.mysql
_uninstall.mysql: ## Uninstall MySQL for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/mysql.sh proj::mysql::uninstall

.PHONY: _install.docker.mysql
_install.docker.mysql: ## Install docker install MySQL
	@$(PROJ_ROOT_DIR)/scripts/installation/mysql.sh proj::mysql::docker::install

.PHONY: _uninstall.docker.mysql
_uninstall.docker.mysql: ## Uninstall docker install MySQL
	@$(PROJ_ROOT_DIR)/scripts/installation/mysql.sh proj::mysql::docker::uninstall

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



##@ Sentry Service
# ==============================================================================
# Sentry installation methods
# ==============================================================================
.PHONY: _install.sentry
_install.sentry:  ## Install Sentry for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::sentry::install

.PHONY: _uninstall.sentry
_uninstall.sentry: ## Uninstall Sentry for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::sentry::uninstall

.PHONY: _install.docker.sentry
_install.docker.sentry: ## Install docker install Sentry
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::sentry::docker::install

.PHONY: _uninstall.docker.sentry
_uninstall.docker.sentry: ## Uninstall docker install Sentry
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::sentry::docker::uninstall


##@ Victoria Suite (Complete Stack)
# ==============================================================================
# Victoria Suite installation methods (unified stack)
# ==============================================================================
.PHONY: _install.victoria
_install.victoria:  ## Install Victoria Suite for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh install

.PHONY: _uninstall.victoria
_uninstall.victoria: ## Uninstall Victoria Suite for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh uninstall

.PHONY: _install.docker.victoria
_install.docker.victoria: ## Install docker install Victoria Suite
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh docker.install

.PHONY: _uninstall.docker.victoria
_uninstall.docker.victoria: ## Uninstall docker install Victoria Suite
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh docker.uninstall

.PHONY: deploy.status.victoria
deploy.status.victoria: ## Check Victoria Suite status
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh status

.PHONY: deploy.info.victoria
deploy.info.victoria: ## Show Victoria Suite information
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh info

.PHONY: deploy.install.all.victoria
deploy.install.all.victoria: ## Install all Victoria components individually using Docker
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh install.all

.PHONY: deploy.uninstall.all.victoria
deploy.uninstall.all.victoria: ## Uninstall all Victoria components
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh uninstall.all

##@ Victoria Individual Components  
# ==============================================================================
# Individual component installation methods (via unified entry point)
# ==============================================================================
.PHONY: _install.victoriametrics
_install.victoriametrics:  ## Install VictoriaMetrics for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh victoriametrics.install

.PHONY: _uninstall.victoriametrics
_uninstall.victoriametrics: ## Uninstall VictoriaMetrics for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh victoriametrics.uninstall

.PHONY: _install.docker.victoriametrics
_install.docker.victoriametrics: ## Install docker install VictoriaMetrics
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh victoriametrics.docker.install

.PHONY: _uninstall.docker.victoriametrics
_uninstall.docker.victoriametrics: ## Uninstall docker install VictoriaMetrics
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh victoriametrics.docker.uninstall

.PHONY: deploy.status.victoriametrics
deploy.status.victoriametrics: ## Check VictoriaMetrics status
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh victoriametrics.status

.PHONY: deploy.info.victoriametrics
deploy.info.victoriametrics: ## Show VictoriaMetrics information
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh victoriametrics.info

.PHONY: _install.victorialogs
_install.victorialogs:  ## Install VictoriaLogs for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh victorialogs.install

.PHONY: _uninstall.victorialogs
_uninstall.victorialogs: ## Uninstall VictoriaLogs for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh victorialogs.uninstall

.PHONY: _install.docker.victorialogs
_install.docker.victorialogs: ## Install docker install VictoriaLogs
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh victorialogs.docker.install

.PHONY: _uninstall.docker.victorialogs
_uninstall.docker.victorialogs: ## Uninstall docker install VictoriaLogs
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh victorialogs.docker.uninstall

.PHONY: deploy.status.victorialogs
deploy.status.victorialogs: ## Check VictoriaLogs status
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh victorialogs.status

.PHONY: deploy.info.victorialogs
deploy.info.victorialogs: ## Show VictoriaLogs information
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh victorialogs.info

.PHONY: _install.vmagent
_install.vmagent:  ## Install vmagent for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh vmagent.install

.PHONY: _uninstall.vmagent
_uninstall.vmagent: ## Uninstall vmagent for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh vmagent.uninstall

.PHONY: _install.docker.vmagent
_install.docker.vmagent: ## Install docker install vmagent
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh vmagent.docker.install

.PHONY: _uninstall.docker.vmagent
_uninstall.docker.vmagent: ## Uninstall docker install vmagent
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh vmagent.docker.uninstall

.PHONY: deploy.status.vmagent
deploy.status.vmagent: ## Check vmagent status
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh vmagent.status

.PHONY: deploy.info.vmagent
deploy.info.vmagent: ## Show vmagent information
	@$(PROJ_ROOT_DIR)/scripts/installation/victoria.sh vmagent.info