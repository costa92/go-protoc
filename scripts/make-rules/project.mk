##@ Project
# The following commands are project-specific tasks.

.PHONY: wire
wire: ## Generate wire dependency injection code.
	cd internal/apiserver && wire

.PHONY: fmt
fmt: ## Format go source code.
	go fmt ./...

.PHONY: run-api
run-api: ## Run the API server.
	go run cmd/apiserver/main.go --config=configs/apiserver.yaml

.PHONY: kill-ports
kill-ports: ## Kill processes using ports 8080 and 9090.
	@echo "🔪 Killing processes using ports 8080 and 9090..."
	@lsof -i :8080 | grep LISTEN | awk '{print $$2}' | xargs kill -9 2>/dev/null || true
	@lsof -i :9090 | grep LISTEN | awk '{print $$2}' | xargs kill -9 2>/dev/null || true
	@echo "✅ Ports cleaned!"

.PHONY: clean-run
clean-run: ## Clean ports and run the API server.
	@$(MAKE) kill-ports
	@echo "🚀 Starting clean API server..."
	@$(MAKE) run-api

.PHONY: generate
generate: ## Generate code from protobuf definitions.
	buf generate

# Note: Build commands have been moved to build.mk for enhanced functionality
# Use 'make build' for standard builds or 'make build.help' for all build options

.PHONY: apidiff
apidiff: tools.verify.go-apidiff ## Run the go-apidiff to verify any API differences compared with origin/master.
	@go-apidiff master --compare-imports --print-compatible --repo-path=.

.PHONY: tidy
tidy: ## Tidy go module dependencies.
	@$(GO) mod tidy

.PHONY: lint
lint: tools.verify.golangci-lint ## Run golangci-lint to check code quality.
	@golangci-lint run

.PHONY: lint-fix
lint-fix: tools.verify.golangci-lint ## Run golangci-lint and automatically fix issues where possible.
	@golangci-lint run --fix

.PHONY: lint-fast
lint-fast: tools.verify.golangci-lint ## Run golangci-lint with reduced linters for quick feedback.
	@golangci-lint run --disable=gocritic,gosec,staticcheck

.PHONY: lint-new
lint-new: tools.verify.golangci-lint ## Run golangci-lint only on new or changed files.
	@golangci-lint run --new-from-rev=HEAD~1

.PHONY: lint-diff
lint-diff: tools.verify.golangci-lint ## Run golangci-lint on files changed in current branch compared to master.
	@golangci-lint run --new-from-rev=origin/master
