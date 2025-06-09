
# code-generator is a makefile target not a real tool.
CI_WORKFLOW_TOOLS := code-generator golangci-lint goimports wire

# unused tools in this project: gentool
OTHER_TOOLS := mockgen uplift git-chglog addlicense kratos kind go-apidiff gotests \
			   cfssl go-gitlint kustomize kafkactl kube-linter kubeconform kubectl \
			   helm-docs db2struct gentool air swagger license gothanks kubebuilder \
			   go-junit-report controller-gen

.PHONY: tools.install
tools.install: install.ci _install.other tools.print-manual-tool ## Install all tools.

.PHONY: tools.install.%
tools.install.%: ## Install a specified tool.
	@echo "===========> Installing $*"
	@$(MAKE) _install.$*

.PHONY: tools.verify.%
tools.verify.%: ## Verify a specified tool.
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi

.PHONY: _install.ci
_install.ci: $(addprefix tools.install., $(CI_WORKFLOW_TOOLS)) ## Install necessary tools used by CI/CD workflow.

.PHONY: _install.other
_install.other: $(addprefix tools.install., $(OTHER_TOOLS))


.PHONY: _install.swagger
_install.swagger:
	@$(GO) install github.com/go-swagger/go-swagger/cmd/swagger@$(GO_SWAGGER_VERSION)


.PHONY: _install.golangci-lint
_install.golangci-lint: ## Install golangci-lint.
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@$(SCRIPTS_DIR)/add-completion.sh golangci-lint bash

.PHONY: _install.protoc-gen-validate
_install.protoc-gen-validate: ## Install protoc-gen-validate.
	@$(GO) install github.com/envoyproxy/protoc-gen-validate@$(PROTOC_GEN_VALIDATE_VERSION)