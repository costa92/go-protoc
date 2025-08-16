# ==============================================================================
# Database Management Commands
# This makefile contains commands for managing MySQL database



##@ Database

.PHONY: db-setup
db-setup: ## Setup database (start MySQL + run migrations)
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::setup

.PHONY: db-start
db-start: ## Start MySQL database using Docker Compose
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::start

.PHONY: db-stop
db-stop: ## Stop MySQL database
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::stop

.PHONY: db-restart
db-restart: ## Restart MySQL database
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::restart

.PHONY: db-logs
db-logs: ## Show MySQL database logs
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::logs

.PHONY: db-status
db-status: ## Check MySQL database status
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::status

.PHONY: db-migrate
db-migrate: ## Run database migrations
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::migrate

.PHONY: db-migrate-dry
db-migrate-dry: ## Show what migrations would be applied (dry run)
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::migrate_dry

.PHONY: db-connect
db-connect: ## Connect to MySQL database using CLI
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::connect

.PHONY: db-shell
db-shell: ## Open MySQL shell in container
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::shell

.PHONY: db-backup
db-backup: ## Create database backup
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::backup

.PHONY: db-restore
db-restore: ## Restore database from backup (Usage: make db-restore BACKUP_FILE=backup.sql)
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::restore "$(BACKUP_FILE)"

.PHONY: db-reset
db-reset: ## Reset database (WARNING: This will delete all data!)
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::reset

.PHONY: db-clean
db-clean: ## Remove MySQL container and volumes (WARNING: All data will be lost!)
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::clean

.PHONY: db-create-migration
db-create-migration: ## Create a new migration file (Usage: make db-create-migration NAME=add_user_table)
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::create_migration "$(NAME)"

.PHONY: db-help
db-help: ## Show database commands help
	@$(PROJ_ROOT_DIR)/scripts/database/mysql.sh proj::mysql::help