# ==============================================================================
#  Makefile helper functions for tools
#
# Specify tools category.
CODE_GENERATOR_TOOLS = client-gen conversion-gen deepcopy-gen defaulter-gen informer-gen lister-gen prerelease-lifecycle-gen \
                      register-gen applyconfiguration-gen go-to-protobuf

# code-generator is a makefile target not a real tool.
CI_WORKFLOW_TOOLS := code-generator golangci-lint goimports wire

# ==============================================================================
# Tools
#
.PHONY: tools.install.%
tools.install.%: ## Install a specified tool.
	@echo "===========> Installing $*"
	@$(MAKE) _install.$*

.PHONY: tools.verify.%
tools.verify.%: ## Verify a specified tool.
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi

# ==============================================================================

.PHONY: _install.ci
_install.ci: $(addprefix tools.install., $(CI_WORKFLOW_TOOLS)) ## Install necessary tools used by CI/CD workflow.

# 安装 code-generator 工具
.PHONY: _install.code-generator
_install.code-generator: ## Install all code generator tools.
	@$(MAKE) install-code-generator

# 安装 wire 工具
.PHONY: _install.wire
_install.wire: ## Install wire.
	@$(GO) install github.com/google/wire/cmd/wire@$(WIRE_VERSION)

# 安装 golangci-lint 工具
.PHONY: _install.golangci-lint
_install.golangci-lint: ## Install golangci-lint.
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

# 安装 goimports 工具
.PHONY: _install.goimports
_install.goimports: ## Install goimports.
	@$(GO) install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)

# 安装 kratos 工具
.PHONY: _install.kratos
_install.kratos: _install.grpc ## Install kratos toolkit, includes multiple protoc plugins.
	@$(GO) install github.com/joelanford/go-apidiff@$(GO_APIDIFF_VERSION)
	@$(GO) install github.com/envoyproxy/protoc-gen-validate@$(PROTOC_GEN_VALIDATE_VERSION)
	@$(GO) install github.com/google/gnostic/cmd/protoc-gen-openapi@$(PROTOC_GEN_OPENAPI_VERSION)
	@$(GO) install github.com/go-kratos/kratos/cmd/kratos/v2@$(KRATOS_VERSION)
	@$(GO) install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@$(KRATOS_VERSION)
	@$(GO) install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@$(KRATOS_VERSION)
	@$(SCRIPTS_DIR)/add-completion.sh kratos bash

# 安装 grpcurl 工具
.PHONY: _install.grpcurl
_install.grpcurl:
	@$(GO) install github.com/fullstorydev/grpcurl/cmd/grpcurl@$(GRPCURL_VERSION)

# 安装 logcheck 工具
.PHONY: _install.logcheck
_install.logcheck:
	@$(GO) install sigs.k8s.io/logtools/logcheck@$(LOGCHECK_VERSION)

# 安装 protoc-gen-deepcopy 工具
.PHONY: _install.protoc-gen-deepcopy
_install.protoc-gen-deepcopy:
	@$(GO) install github.com/protobuf-tools/protoc-gen-deepcopy@latest

# 安装 protoc-gen-go-json 工具
.PHONY: _install.protoc-gen-go-json
_install.protoc-gen-go-json:
	@$(GO) install github.com/mfridman/protoc-gen-go-json@latest

# 安装 go-mod-upgrade 工具
.PHONY: _install.go-mod-upgrade
_install.go-mod-upgrade:
	@$(GO) install github.com/oligot/go-mod-upgrade@latest

.PHONY: install-code-generator
install-code-generator:
	@for tool in $(CODE_GENERATOR_TOOLS); do \
		echo "===========> Installing $$tool"; \
		$(GO) install k8s.io/code-generator/cmd/$$tool@$(CODE_GENERATOR_VERSION); \
	done

.PHONY: install-tools
install-tools: _install.ci ## Install all tools used by CI/CD workflow.