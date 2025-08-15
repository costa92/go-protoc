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
	go run cmd/apiserver/main.go

.PHONY: generate
generate: ## Generate code from protobuf definitions.
	buf generate

.PHONY: build
build: ## Build the API server binary.
	go build -o bin/apiserver cmd/apiserver/main.go

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
