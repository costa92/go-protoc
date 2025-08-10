##@ Deploy

.PHONY: deploy.install.redis
deploy.install.redis: ##  Install redis
	@$(PROJ_ROOT_DIR)/scripts/installation/install.sh
	proj::redis::install