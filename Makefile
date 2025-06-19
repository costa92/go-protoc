# Build all by default, even if it's not first
.DEFAULT_GOAL := help


# ==============================================================================
# Includes

include scripts/make-rules/common.mk # make sure include common.mk at the first include line
include scripts/make-rules/all.mk

# 生成 Wire 代码
.PHONY: wire
wire:
	cd internal/apiserver && wire


.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: run-api
run-api:
	cd cmd/apiserver && go run main.go

.PHONY: proto
proto:
	buf generate

.PHONY: build
build:
	go build -o bin/apiserver cmd/apiserver/main.go

.PHONY: install-tools
install-tools: ## Install CI-related tools. Install all tools by specifying `A=1`.
	$(MAKE) install.ci
	if [[ "$(A)" == 1 ]]; then                                             \
		$(MAKE) _install.other ;                                            \
	fi


.PHONY: apidiff
apidiff: tools.verify.go-apidiff ## Run the go-apidiff to verify any API differences compared with origin/master.
	@go-apidiff master --compare-imports --print-compatible --repo-path=.

.PHONY: tidy
tidy:
	@$(GO) mod tidy



.PHONY: targets
targets: Makefile ## Show all Sub-makefile targets.
	@for mk in `echo $(MAKEFILE_LIST) | sed 's/Makefile //g'`; do echo -e \\n\\033[35m$$mk\\033[0m; awk -F':.*##' '/^[0-9A-Za-z._-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^\$$\([0-9A-Za-z_-]+\):.*?##/ { gsub("_","-", $$1); printf "  \033[36m%-45s\033[0m %s\n", tolower(substr($$1, 3, length($$1)-7)), $$2 }' $$mk;done;
# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php
.PHONY: help
help: Makefile ## Display this help info.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<TARGETS> <OPTIONS>\033[0m\n\n\033[35mTargets:\033[0m\n"} /^[0-9A-Za-z._-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^\$$\([0-9A-Za-z_-]+\):.*?##/ { gsub("_","-", $$1); printf "  \033[36m%-45s\033[0m %s\n", tolower(substr($$1, 3, length($$1)-7)), $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' Makefile #$(MAKEFILE_LIST)
	@echo -e "$$USAGE_OPTIONS"

