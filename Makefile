# Build all by default, even if it's not first
.DEFAULT_GOAL := help


# ==============================================================================
# Includes

# Include common variables and functions.
include scripts/make-rules/common.mk

# Include all other makefiles.
include scripts/make-rules/all.mk

# ==============================================================================

.PHONY: install-tools
install-tools: ## Install CI-related tools. Install all tools by specifying `A=1`.
	$(MAKE) install.ci
	if [[ "$(A)" == 1 ]]; then \
		$(MAKE) _install.other ; \
	fi

.PHONY: targets
targets: Makefile ## Show all Sub-makefile targets.
	@for mk in `echo $(MAKEFILE_LIST) | sed 's/Makefile //g'`; do \
		if grep -q -E ':.*##' "$$mk"; then \
			printf '\n\033[35m%s\033[0m\n' "$$mk"; \
			category=$$(grep -m 1 '^##@' "$$mk" 2>/dev/null | sed 's/^##@ //' | tr '[:upper:]' '[:lower:]'); \
			awk -F':.*##' -v category="$$category" -f scripts/targets.awk "$$mk"; \
		fi; \
	done

.PHONY: list-mk
list-mk: ## List all included makefiles.
	@printf "Included makefiles:\n"
	@for mk in $(MAKEFILE_LIST); do \
		printf "  - %s\n" "$$mk"; \
	done
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
	@awk -f scripts/help.awk $(MAKEFILE_LIST)


.PHONY: rename-project
rename-project: ## Rename the project module path. Usage: make rename-project OLD_PATH=... NEW_PATH=...
	@echo "Renaming project from $(OLD_PATH) to $(NEW_PATH)..."
	@# MacOS sed requires a backup extension, Linux sed does not.
	@# The `sed -i.bak` command works on both, creating a .bak file on MacOS
	@# and a file with literal '.bak' extension on Linux. We remove these backups.
	@find . -type f \( -name "*.go" -o -name "go.mod" -o -name "go.sum" -o -name "*.sh" -o -name "*.yaml" -o -name "*.md" \)
		-not -path "./.git/*"
		-not -path "./.cursor/*"
		-not -path "./.serena/*"
		-not -path "./vendor/*"
		-print0 | xargs -0 sed -i.bak 's|$(OLD_PATH)|$(NEW_PATH)|g'
	@find . -type f -name "*.bak" -delete
	@echo "Project renaming complete. Please review the changes."
	@echo "You might need to run 'go mod tidy' or similar commands."