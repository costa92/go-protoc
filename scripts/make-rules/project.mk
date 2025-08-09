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
