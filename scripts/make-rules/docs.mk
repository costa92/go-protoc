##@ Documentation
# ==============================================================================
# Development documentation generation and management

.PHONY: docs.dev
docs.dev: ## Generate development documentation 生成开发文档到 docs/DEVELOPMENT.md
	@echo "===========> Generating development documentation..."
	@mkdir -p docs
	@$(MAKE) _docs.generate-dev

.PHONY: docs.api
docs.api: ## Generate API documentation from protobuf
	@echo "===========> Generating API documentation..."
	@$(MAKE) _docs.generate-api

.PHONY: docs.update
docs.update: docs.api docs.dev ## Update all documentation (API + dev docs)
	@echo "===========> All documentation updated successfully"

.PHONY: docs.serve
docs.serve: ## Serve generated documentation locally
	@echo "===========> Starting documentation server..."
	@$(MAKE) _docs.serve

.PHONY: docs.clean
docs.clean: ## Clean all generated documentation
	@echo "===========> Cleaning generated documentation..."
	@rm -f docs/DEVELOPMENT.md
	@rm -rf docs/generated
	@echo "Documentation cleaned"

.PHONY: _docs.generate-dev
_docs.generate-dev: ## Internal: Generate development documentation
	@scripts/generate-dev-docs.sh

.PHONY: _docs.generate-api
_docs.generate-api: ## Internal: Generate API documentation
	@mkdir -p docs/generated
	@if command -v protoc-gen-doc >/dev/null 2>&1; then \
		echo "Generating API documentation..."; \
		buf generate --template buf.gen.docs.yaml; \
	else \
		echo "protoc-gen-doc not found, installing..."; \
		$(MAKE) tools.install.protoc-gen-doc; \
		buf generate --template buf.gen.docs.yaml; \
	fi
	@echo "API documentation updated"

.PHONY: _docs.serve
_docs.serve: ## Internal: Serve documentation
	@if command -v http-server >/dev/null 2>&1; then \
		http-server docs -p 8081; \
	else \
		echo "Installing http-server..."; \
		npm install -g http-server; \
		http-server docs -p 8081; \
	fi