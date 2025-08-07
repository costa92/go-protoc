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

.PHONY: tools.install.ci
tools.install.ci: $(addprefix tools.install., $(CI_WORKFLOW_TOOLS)) ## Install CI workflow tools (wire, golangci-lint, etc).

# 安装 code-generator 工具
.PHONY: tools.install.code-generator
tools.install.code-generator: ## Install Kubernetes code generators (client-gen, etc).

# 安装 wire 工具
.PHONY: tools.install.wire
tools.install.wire: ## Install Google Wire dependency injection generator.

# 安装 golangci-lint 工具
.PHONY: tools.install.golangci-lint
tools.install.golangci-lint: ## Install Go static analysis runner.

# 安装 goimports 工具
.PHONY: tools.install.goimports
tools.install.goimports: ## Install Go import organization tool.

# 安装 kratos 工具
.PHONY: tools.install.kratos
tools.install.kratos: tools.install.grpc ## Install Kratos toolkit and protoc plugins.
	@$(GO) install github.com/joelanford/go-apidiff@$(GO_APIDIFF_VERSION)
	@$(GO) install github.com/envoyproxy/protoc-gen-validate@$(PROTOC_GEN_VALIDATE_VERSION)
	@$(GO) install github.com/google/gnostic/cmd/protoc-gen-openapi@$(PROTOC_GEN_OPENAPI_VERSION)
	@$(GO) install github.com/go-kratos/kratos/cmd/kratos/v2@$(KRATOS_VERSION)
	@$(GO) install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@$(KRATOS_VERSION)
	@$(GO) install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@$(KRATOS_VERSION)
	@$(SCRIPTS_DIR)/add-completion.sh kratos bash

# 安装 grpcurl 工具
.PHONY: tools.install.grpcurl
tools.install.grpcurl: ## Install gRPC command line testing tool.

# 安装 logcheck 工具
.PHONY: tools.install.logcheck
tools.install.logcheck: ## Install Kubernetes logging lint tool.

# 安装 protoc-gen-deepcopy 工具
.PHONY: tools.install.protoc-gen-deepcopy
tools.install.protoc-gen-deepcopy: ## Install protobuf deep copy generator.

# 安装 protoc-gen-go-json 工具
.PHONY: tools.install.protoc-gen-go-json
tools.install.protoc-gen-go-json: ## Install protobuf JSON generation plugin.

# 安装 go-mod-upgrade 工具
.PHONY: tools.install.go-mod-upgrade
tools.install.go-mod-upgrade: ## Install Go module upgrade helper.

.PHONY: tools.install.grpc
tools.install.grpc: ## Install gRPC tools and protoc-plugins.

# ==============================================================================
# Internal installation methods (called by tools.install.* targets)
#
.PHONY: _install.code-generator
_install.code-generator:
	@$(MAKE) install-code-generator

.PHONY: _install.wire
_install.wire:
	@$(GO) install github.com/google/wire/cmd/wire@$(WIRE_VERSION)

.PHONY: _install.golangci-lint
_install.golangci-lint:
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: _install.goimports
_install.goimports:
	@$(GO) install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)

.PHONY: _install.grpc
_install.grpc:
	@$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

.PHONY: _install.kratos
_install.kratos: _install.grpc
	@$(GO) install github.com/joelanford/go-apidiff@$(GO_APIDIFF_VERSION)
	@$(GO) install github.com/envoyproxy/protoc-gen-validate@$(PROTOC_GEN_VALIDATE_VERSION)
	@$(GO) install github.com/google/gnostic/cmd/protoc-gen-openapi@$(PROTOC_GEN_OPENAPI_VERSION)
	@$(GO) install github.com/go-kratos/kratos/cmd/kratos/v2@$(KRATOS_VERSION)
	@$(GO) install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@$(KRATOS_VERSION)
	@$(GO) install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@$(KRATOS_VERSION)
	@$(SCRIPTS_DIR)/add-completion.sh kratos bash

.PHONY: _install.grpcurl
_install.grpcurl:
	@$(GO) install github.com/fullstorydev/grpcurl/cmd/grpcurl@$(GRPCURL_VERSION)

.PHONY: _install.logcheck
_install.logcheck:
	@$(GO) install sigs.k8s.io/logtools/logcheck@$(LOGCHECK_VERSION)

.PHONY: _install.protoc-gen-deepcopy
_install.protoc-gen-deepcopy:
	@$(GO) install github.com/protobuf-tools/protoc-gen-deepcopy@latest

.PHONY: _install.protoc-gen-go-json
_install.protoc-gen-go-json:
	@$(GO) install github.com/mfridman/protoc-gen-go-json@latest

.PHONY: _install.go-mod-upgrade
_install.go-mod-upgrade:
	@$(GO) install github.com/oligot/go-mod-upgrade@latest

.PHONY: _install.buf
_install.buf:
	@$(GO) install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)

.PHONY: install-code-generator
install-code-generator:
	@for tool in $(CODE_GENERATOR_TOOLS); do \
		echo "===========> Installing $$tool"; \
		$(GO) install k8s.io/code-generator/cmd/$$tool@$(CODE_GENERATOR_VERSION); \
	done

.PHONY: install-tools
install-tools: tools.install.ci ## Install all CI workflow tools (wire, golangci-lint, goimports, code-generator).


.PHONY: tools.install.kratos
tools.install.kratos: tools.install.grpc ## Install Kratos toolkit and protoc plugins.

.PHONY: tools.install.buf
tools.install.buf: ## Install Protocol Buffers compiler (buf).