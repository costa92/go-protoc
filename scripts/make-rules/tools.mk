##@ Tools
# ==============================================================================
#  Makefile helper functions for tools
#
# Specify tools category.
CODE_GENERATOR_TOOLS = client-gen conversion-gen deepcopy-gen defaulter-gen informer-gen lister-gen prerelease-lifecycle-gen \
                      register-gen applyconfiguration-gen go-to-protobuf

# code-generator is a makefile target not a real tool.
CI_WORKFLOW_TOOLS := code-generator golangci-lint goimports wire

# ==============================================================================
# Tools (Consistent naming with tools.install.* prefix)
#
.PHONY: tools.install.%
tools.install.%: ## Install a specified tool.
	@echo "===========> Installing $*"
	@$(MAKE) _install.$*

.PHONY: tools.verify.%
tools.verify.%: ## Check if a tool is installed, install if missing.
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi

# ==============================================================================
# Internal installation methods (called by tools.install.* targets)
#
.PHONY: _install.code-generator
_install.code-generator: ## Install Kubernetes code generators (client-gen, deepcopy-gen, etc.)
	@$(MAKE) install-code-generator

.PHONY: _install.wire
_install.wire: ## Install Google Wire dependency injection tool
	@$(GO) install github.com/google/wire/cmd/wire@$(WIRE_VERSION)

.PHONY: _install.golangci-lint
_install.golangci-lint: ## Install golangci-lint Go linting tool
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: _install.goimports
_install.goimports: ## Install goimports Go import formatting tool
	@$(GO) install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)

.PHONY: _install.grpc
_install.grpc: ## Install protoc-gen-go and protoc-gen-go-grpc for gRPC development
	@$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

.PHONY: _install.kratos
_install.kratos: _install.grpc ## Install Kratos framework and related protoc plugins
	@$(GO) install github.com/joelanford/go-apidiff@$(GO_APIDIFF_VERSION)
	@$(GO) install github.com/envoyproxy/protoc-gen-validate@$(PROTOC_GEN_VALIDATE_VERSION)
	@$(GO) install github.com/google/gnostic/cmd/protoc-gen-openapi@$(PROTOC_GEN_OPENAPI_VERSION)
	@$(GO) install github.com/go-kratos/kratos/cmd/kratos/v2@$(KRATOS_VERSION)
	@$(GO) install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@$(KRATOS_VERSION)
	@$(GO) install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@$(KRATOS_VERSION)
	@$(SCRIPTS_DIR)/add-completion.sh kratos bash

.PHONY: _install.grpcurl
_install.grpcurl: ## Install grpcurl tool for gRPC API testing
	@$(GO) install github.com/fullstorydev/grpcurl/cmd/grpcurl@$(GRPCURL_VERSION)

.PHONY: _install.logcheck
_install.logcheck: ## Install logcheck for Kubernetes log formatting validation
	@$(GO) install sigs.k8s.io/logtools/logcheck@$(LOGCHECK_VERSION)

.PHONY: _install.protoc-gen-deepcopy
_install.protoc-gen-deepcopy: ## Install protoc-gen-deepcopy for generating deepcopy methods
	@$(GO) install github.com/protobuf-tools/protoc-gen-deepcopy@latest

.PHONY: _install.protoc-gen-go-json
_install.protoc-gen-go-json: ## Install protoc-gen-go-json for JSON marshaling/unmarshaling
	@$(GO) install github.com/mfridman/protoc-gen-go-json@latest

.PHONY: _install.go-mod-upgrade
_install.go-mod-upgrade: ## Install go-mod-upgrade for API
	@$(GO) install github.com/oligot/go-mod-upgrade@latest

.PHONY: _install.buf
_install.buf: ## Install buf for API
	@$(GO) install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)

.PHONY: install-code-generator
install-code-generator:
	@for tool in $(CODE_GENERATOR_TOOLS); do \
		echo "===========> Installing $$tool"; \
		$(GO) install k8s.io/code-generator/cmd/$$tool@$(CODE_GENERATOR_VERSION); \
	done


.PHONY: _install.protoc-gen-doc
_install.protoc-gen-doc: ## Install protoc-gen-doc for API documentation
	@echo "Installing protoc-gen-doc..."
	@$(GO) install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest

.PHONY: _install.docker-compose
_install.docker-compose: ## Install Docker Compose for container orchestration
	@DOCKER_COMPOSE_VERSION=$(DOCKER_COMPOSE_VERSION) $(SCRIPTS_DIR)/installation/docker-compose.sh

