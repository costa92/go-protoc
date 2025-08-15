# ==============================================================================
# Database Management Commands
# This makefile contains commands for managing MySQL database

##@ Database

.PHONY: db-setup
db-setup: ## Setup database (start MySQL + run migrations)
	@echo "Setting up database..."
	$(MAKE) db-start
	@echo "Waiting for MySQL to be ready..."
	@sleep 30
	$(MAKE) db-migrate
	@echo "Database setup completed!"

.PHONY: db-start
db-start: ## Start MySQL database using Docker Compose
	@echo "Starting MySQL database..."
	@cd deployments/mysql && docker-compose up -d
	@echo "MySQL is starting... Please wait for it to be ready."

.PHONY: db-stop
db-stop: ## Stop MySQL database
	@echo "Stopping MySQL database..."
	@cd deployments/mysql && docker-compose down

.PHONY: db-restart
db-restart: ## Restart MySQL database
	@echo "Restarting MySQL database..."
	$(MAKE) db-stop
	$(MAKE) db-start

.PHONY: db-logs
db-logs: ## Show MySQL database logs
	@cd deployments/mysql && docker-compose logs -f mysql

.PHONY: db-status
db-status: ## Check MySQL database status
	@echo "Checking MySQL status..."
	@cd deployments/mysql && docker-compose ps

.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	@if ! command -v mysql >/dev/null 2>&1; then \
		echo "MySQL client not found. Please install MySQL client first."; \
		echo "On macOS: brew install mysql-client"; \
		echo "On Ubuntu: sudo apt-get install mysql-client"; \
		exit 1; \
	fi
	@echo "Applying migration: 001_create_users_table.sql"
	@mysql -h 127.0.0.1 -P 3306 -u root -p'proj(#)666' onex < deployments/mysql/migrations/001_create_users_table.sql
	@echo "Database migrations completed successfully!"

.PHONY: db-migrate-dry
db-migrate-dry: ## Show what migrations would be applied (dry run)
	@echo "=== Database Migration Plan ==="
	@echo "The following SQL will be executed:"
	@echo "File: deployments/mysql/migrations/001_create_users_table.sql"
	@echo ""
	@cat deployments/mysql/migrations/001_create_users_table.sql

.PHONY: db-connect
db-connect: ## Connect to MySQL database using CLI
	@echo "Connecting to MySQL database..."
	@mysql -h 127.0.0.1 -P 3306 -u root -p'proj(#)666'

.PHONY: db-backup
db-backup: ## Create database backup
	@echo "Creating database backup..."
	@BACKUP_FILE="backup_onex_$(shell date +%Y%m%d_%H%M%S).sql"; \
	mysqldump -h 127.0.0.1 -P 3306 -u root -p'proj(#)666' \
		--single-transaction --routines --triggers onex > "$$BACKUP_FILE" && \
	echo "Database backup created: $$BACKUP_FILE"

.PHONY: db-restore
db-restore: ## Restore database from backup (Usage: make db-restore BACKUP_FILE=backup.sql)
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "Please specify BACKUP_FILE. Usage: make db-restore BACKUP_FILE=backup.sql"; \
		exit 1; \
	fi
	@if [ ! -f "$(BACKUP_FILE)" ]; then \
		echo "Backup file $(BACKUP_FILE) not found!"; \
		exit 1; \
	fi
	@echo "Restoring database from $(BACKUP_FILE)..."
	@mysql -h 127.0.0.1 -P 3307 -u root -p'proj(#)666' onex < "$(BACKUP_FILE)"
	@echo "Database restore completed!"

.PHONY: db-reset
db-reset: ## Reset database (WARNING: This will delete all data!)
	@echo "⚠️  WARNING: This will delete all data in the database!"
	@echo "Are you sure you want to continue? [y/N]"
	@read -r confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		echo "Resetting database..."; \
		mysql -h 127.0.0.1 -P 3307 -u root -p'proj(#)666' -e "DROP DATABASE IF EXISTS onex; CREATE DATABASE onex CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"; \
		$(MAKE) db-migrate; \
		echo "Database reset completed!"; \
	else \
		echo "Operation cancelled."; \
	fi

.PHONY: db-clean
db-clean: ## Remove MySQL container and volumes (WARNING: All data will be lost!)
	@echo "⚠️  WARNING: This will remove all MySQL data permanently!"
	@echo "Are you sure you want to continue? [y/N]"
	@read -r confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		echo "Removing MySQL container and volumes..."; \
		cd deployments/mysql && docker-compose down -v; \
		docker volume rm go-protoc-mysql-data 2>/dev/null || true; \
		echo "MySQL cleanup completed!"; \
	else \
		echo "Operation cancelled."; \
	fi

.PHONY: db-create-migration
db-create-migration: ## Create a new migration file (Usage: make db-create-migration NAME=add_user_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Please specify migration NAME. Usage: make db-create-migration NAME=add_user_table"; \
		exit 1; \
	fi
	@MIGRATION_NUM=$$(ls deployments/mysql/migrations/ | grep -E '^[0-9]+_' | wc -l | tr -d ' '); \
	MIGRATION_NUM=$$(printf "%03d" $$((MIGRATION_NUM + 1))); \
	MIGRATION_FILE="deployments/mysql/migrations/$${MIGRATION_NUM}_$(NAME).sql"; \
	echo "-- Migration: $${MIGRATION_NUM}_$(NAME).sql" > "$$MIGRATION_FILE"; \
	echo "-- Description: $(NAME)" >> "$$MIGRATION_FILE"; \
	echo "-- Created: $$(date +%Y-%m-%d)" >> "$$MIGRATION_FILE"; \
	echo "" >> "$$MIGRATION_FILE"; \
	echo "-- Add your SQL statements here" >> "$$MIGRATION_FILE"; \
	echo "" >> "$$MIGRATION_FILE"; \
	echo "Migration file created: $$MIGRATION_FILE"

.PHONY: db-help
db-help: ## Show database commands help
	@echo "=== Database Management Commands ==="
	@echo ""
	@echo "Setup Commands:"
	@echo "  make db-setup         - Complete database setup (start + migrate)"
	@echo "  make db-start         - Start MySQL container"
	@echo "  make db-stop          - Stop MySQL container"
	@echo "  make db-restart       - Restart MySQL container"
	@echo ""
	@echo "Migration Commands:"
	@echo "  make db-migrate       - Run all pending migrations"
	@echo "  make db-migrate-dry   - Show migration plan (dry run)"
	@echo "  make db-create-migration NAME=migration_name"
	@echo ""
	@echo "Database Operations:"
	@echo "  make db-connect       - Connect via MySQL CLI"
	@echo "  make db-shell         - Open MySQL shell in container"
	@echo "  make db-backup        - Create database backup"
	@echo "  make db-restore BACKUP_FILE=file.sql"
	@echo ""
	@echo "Maintenance Commands:"
	@echo "  make db-status        - Check container status"
	@echo "  make db-logs          - Show container logs"
	@echo "  make db-reset         - Reset database (delete all data)"
	@echo "  make db-clean         - Remove container and volumes"
	@echo ""
	@echo "Configuration:"
	@echo "  Database: onex"
	@echo "  Host: 127.0.0.1:3307"
	@echo "  Username: root"
	@echo "  Password: proj(#)666"