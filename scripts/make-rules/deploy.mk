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


# ==============================================================================
# Internal installation methods
# ==============================================================================
.PHONY: _install.redis
_install.redis:  ## Install Redis for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::redis::install

.PHONY: _uninstall.redis
_uninstall.redis: ## Uninstall Redis for deployment.
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh proj::redis::uninstall