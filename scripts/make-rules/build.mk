##@ Build

# ==============================================================================
# Build System for Binary and Docker Builds
# ==============================================================================

# Build configuration
BUILD_DIR ?= $(PROJ_ROOT_DIR)/bin
DOCKER_BUILD_DIR ?= $(PROJ_ROOT_DIR)/build
DOCKER_REGISTRY ?= docker.io
DOCKER_NAMESPACE ?= costa92
IMAGE_NAME ?= go-protoc
PLATFORM ?= linux/amd64,linux/arm64

# Version and build information
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_VERSION := $(shell git describe --tags --always --match='v*')
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_BRANCH := $(shell git branch --show-current || git rev-parse --abbrev-ref HEAD || echo "unknown")
GIT_TREE_STATE := $(shell if [ -z "`git status --porcelain`" ]; then echo clean; else echo dirty; fi)

# Service configuration
SERVICE_NAME ?= apiserver

# If there are uncommitted changes, append -dirty to version
ifneq (,$(shell git status --porcelain 2>/dev/null))
    GIT_VERSION := $(GIT_VERSION)-dirty
endif

# Build ldflags for version injection
VERSION_PKG := $(PRJ_SRC_PATH)/pkg/version
LDFLAGS := -X '$(VERSION_PKG).serviceName=$(SERVICE_NAME)' \
           -X '$(VERSION_PKG).gitVersion=$(GIT_VERSION)' \
           -X '$(VERSION_PKG).gitCommit=$(GIT_COMMIT)' \
           -X '$(VERSION_PKG).gitBranch=$(GIT_BRANCH)' \
           -X '$(VERSION_PKG).gitTreeState=$(GIT_TREE_STATE)' \
           -X '$(VERSION_PKG).buildDate=$(BUILD_DATE)'

# Build flags
BUILD_FLAGS := -ldflags "$(LDFLAGS)" -trimpath

# Docker image tag
ifeq ($(GIT_TREE_STATE),clean)
    IMAGE_TAG ?= $(GIT_VERSION)
else
    IMAGE_TAG ?= $(GIT_VERSION)
endif

# ==============================================================================
# Binary Build Targets
# ==============================================================================

.PHONY: build
build: build.apiserver ## Build all binaries.

.PHONY: build.apiserver
build.apiserver: ## Build the API server binary.
	@echo "===========> Building $(SERVICE_NAME) binary"
	@echo "Service: $(SERVICE_NAME)"
	@echo "Version: $(GIT_VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@echo "Branch: $(GIT_BRANCH)"
	@echo "Tree State: $(GIT_TREE_STATE)"
	@echo "Build Date: $(BUILD_DATE)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(SERVICE_NAME) $(PRJ_SRC_PATH)/cmd/$(SERVICE_NAME)
	@echo "===========> Build completed: $(BUILD_DIR)/$(SERVICE_NAME)"

.PHONY: build.linux
build.linux: ## Build Linux binaries.
	@echo "===========> Building Linux binaries"
	GOOS=linux GOARCH=amd64 $(MAKE) build.apiserver

.PHONY: build.darwin
build.darwin: ## Build macOS binaries.
	@echo "===========> Building macOS binaries"
	GOOS=darwin GOARCH=amd64 $(MAKE) build.apiserver

.PHONY: build.windows
build.windows: ## Build Windows binaries.
	@echo "===========> Building Windows binaries"
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(SERVICE_NAME).exe $(PRJ_SRC_PATH)/cmd/$(SERVICE_NAME)

.PHONY: build.multiarch
build.multiarch: ## Build binaries for multiple architectures.
	@echo "===========> Building multi-architecture binaries"
	@mkdir -p $(BUILD_DIR)/linux-amd64 $(BUILD_DIR)/linux-arm64 $(BUILD_DIR)/darwin-amd64 $(BUILD_DIR)/darwin-arm64
	GOOS=linux GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/linux-amd64/$(SERVICE_NAME) $(PRJ_SRC_PATH)/cmd/$(SERVICE_NAME)
	GOOS=linux GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/linux-arm64/$(SERVICE_NAME) $(PRJ_SRC_PATH)/cmd/$(SERVICE_NAME)
	GOOS=darwin GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/darwin-amd64/$(SERVICE_NAME) $(PRJ_SRC_PATH)/cmd/$(SERVICE_NAME)
	GOOS=darwin GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/darwin-arm64/$(SERVICE_NAME) $(PRJ_SRC_PATH)/cmd/$(SERVICE_NAME)
	@echo "===========> Multi-architecture builds completed"

# ==============================================================================
# Docker Build Targets
# ==============================================================================

.PHONY: docker
docker: docker.build ## Build Docker image.

.PHONY: docker.build
docker.build: ## Build Docker image.
	@echo "===========> Building Docker image"
	@echo "Image: $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG)"
	@$(MAKE) docker.prepare
	docker build \
		--build-arg SERVICE_NAME="$(SERVICE_NAME)" \
		--build-arg VERSION="$(GIT_VERSION)" \
		--build-arg COMMIT="$(GIT_COMMIT)" \
		--build-arg BRANCH="$(GIT_BRANCH)" \
		--build-arg TREE_STATE="$(GIT_TREE_STATE)" \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		--platform $(PLATFORM) \
		-t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE_NAME):latest \
		$(DOCKER_BUILD_DIR)
	@echo "===========> Docker image built successfully"

.PHONY: docker.buildx
docker.buildx: ## Build multi-platform Docker image using buildx.
	@echo "===========> Building multi-platform Docker image"
	@$(MAKE) docker.prepare
	docker buildx build \
		--build-arg SERVICE_NAME="$(SERVICE_NAME)" \
		--build-arg VERSION="$(GIT_VERSION)" \
		--build-arg COMMIT="$(GIT_COMMIT)" \
		--build-arg BRANCH="$(GIT_BRANCH)" \
		--build-arg TREE_STATE="$(GIT_TREE_STATE)" \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		--platform $(PLATFORM) \
		-t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE_NAME):latest \
		--push \
		$(DOCKER_BUILD_DIR)

.PHONY: docker.prepare
docker.prepare: ## Prepare Docker build context.
	@echo "===========> Preparing Docker build context"
	@mkdir -p $(DOCKER_BUILD_DIR)
	@cp -f $(PROJ_ROOT_DIR)/Dockerfile $(DOCKER_BUILD_DIR)/ 2>/dev/null || $(MAKE) docker.dockerfile
	@cp -rf $(PROJ_ROOT_DIR)/{cmd,internal,pkg,configs,go.mod,go.sum} $(DOCKER_BUILD_DIR)/
	@echo "===========> Docker build context prepared"

.PHONY: docker.dockerfile
docker.dockerfile: ## Generate Dockerfile if it doesn't exist.
	@echo "===========> Generating Dockerfile"
	@mkdir -p $(DOCKER_BUILD_DIR)
	@$(PROJ_ROOT_DIR)/scripts/generate-dockerfile.sh $(DOCKER_BUILD_DIR)
	@echo "===========> Dockerfile generated"

.PHONY: docker.push
docker.push: ## Push Docker image to registry.
	@echo "===========> Pushing Docker image"
	docker push $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE_NAME):latest
	@echo "===========> Docker image pushed successfully"

.PHONY: docker.run
docker.run: ## Run Docker container.
	@echo "===========> Running Docker container"
	docker run --rm -it \
		-p 8080:8080 \
		-p 9090:9090 \
		--name $(IMAGE_NAME)-dev \
		$(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG)

.PHONY: docker.run.daemon
docker.run.daemon: ## Run Docker container in daemon mode.
	@echo "===========> Running Docker container in daemon mode"
	docker run -d \
		-p 8080:8080 \
		-p 9090:9090 \
		--name $(IMAGE_NAME) \
		--restart unless-stopped \
		$(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG)
	@echo "===========> Docker container started: $(IMAGE_NAME)"

.PHONY: docker.stop
docker.stop: ## Stop Docker container.
	@echo "===========> Stopping Docker container"
	@docker stop $(IMAGE_NAME) 2>/dev/null || true
	@docker rm $(IMAGE_NAME) 2>/dev/null || true
	@echo "===========> Docker container stopped"

.PHONY: docker.logs
docker.logs: ## Show Docker container logs.
	docker logs -f $(IMAGE_NAME)

.PHONY: docker.clean
docker.clean: docker.stop ## Clean Docker images and containers.
	@echo "===========> Cleaning Docker images"
	@docker rmi $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true
	@docker rmi $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE_NAME):latest 2>/dev/null || true
	@docker system prune -f
	@echo "===========> Docker cleanup completed"

# ==============================================================================
# Development and Testing Targets
# ==============================================================================

.PHONY: build.debug
build.debug: ## Build debug binary with race detection and debug symbols.
	@echo "===========> Building debug binary"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -race -gcflags="all=-N -l" $(BUILD_FLAGS) -o $(BUILD_DIR)/$(SERVICE_NAME)-debug $(PRJ_SRC_PATH)/cmd/$(SERVICE_NAME)
	@echo "===========> Debug build completed: $(BUILD_DIR)/$(SERVICE_NAME)-debug"

.PHONY: build.test
build.test: ## Build and run tests.
	@echo "===========> Running build tests"
	$(GO) test -race -v ./...
	@echo "===========> Build tests completed"

.PHONY: build.version
build.version: ## Show build version information.
	@echo "===========> Build Version Information"
	@echo "Service Name:   $(SERVICE_NAME)"
	@echo "Git Version:    $(GIT_VERSION)"
	@echo "Git Commit:     $(GIT_COMMIT)"
	@echo "Git Branch:     $(GIT_BRANCH)"
	@echo "Git Tree State: $(GIT_TREE_STATE)"
	@echo "Build Date:     $(BUILD_DATE)"
	@echo "Go Version:     $(shell go version)"
	@echo "Platform:       $(shell go env GOOS)/$(shell go env GOARCH)"
	@echo "LDFLAGS:        $(LDFLAGS)"

.PHONY: build.verify
build.verify: build.apiserver ## Build and verify the binary.
	@echo "===========> Verifying build"
	@$(BUILD_DIR)/$(SERVICE_NAME) --version
	@$(BUILD_DIR)/$(SERVICE_NAME) --version=raw
	@echo "===========> Build verification completed"

# ==============================================================================
# Clean Targets
# ==============================================================================

.PHONY: build.clean
build.clean: ## Clean build artifacts.
	@echo "===========> Cleaning build artifacts"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DOCKER_BUILD_DIR)
	@echo "===========> Build cleanup completed"

.PHONY: build.clean.all
build.clean.all: build.clean docker.clean ## Clean all build artifacts and Docker images.
	@echo "===========> All build artifacts cleaned"

# ==============================================================================
# Help and Information
# ==============================================================================

.PHONY: build.help
build.help: ## Show build system help.
	@echo "Build System Commands:"
	@echo ""
	@echo "Binary Builds:"
	@echo "  build                 - Build all binaries"
	@echo "  build.apiserver       - Build API server binary"
	@echo "  build.linux           - Build Linux binaries"
	@echo "  build.darwin          - Build macOS binaries"  
	@echo "  build.windows         - Build Windows binaries"
	@echo "  build.multiarch       - Build multi-architecture binaries"
	@echo "  build.debug           - Build debug binary"
	@echo ""
	@echo "Docker Builds:"
	@echo "  docker.build          - Build Docker image"
	@echo "  docker.buildx         - Build multi-platform Docker image"
	@echo "  docker.push           - Push Docker image"
	@echo "  docker.run            - Run Docker container"
	@echo "  docker.run.daemon     - Run Docker container in daemon mode"
	@echo "  docker.stop           - Stop Docker container"
	@echo "  docker.logs           - Show Docker container logs"
	@echo ""
	@echo "Utilities:"
	@echo "  build.version         - Show version information"
	@echo "  build.verify          - Build and verify binary"
	@echo "  build.test            - Run build tests"
	@echo "  build.clean           - Clean build artifacts"
	@echo "  build.clean.all       - Clean all artifacts"
	@echo ""
	@echo "Environment Variables:"
	@echo "  SERVICE_NAME          - Service name (default: apiserver)"
	@echo "  DOCKER_REGISTRY       - Docker registry (default: docker.io)"
	@echo "  DOCKER_NAMESPACE      - Docker namespace (default: costa92)"
	@echo "  IMAGE_NAME            - Image name (default: go-protoc)"
	@echo "  IMAGE_TAG             - Image tag (default: git version)"
	@echo "  PLATFORM              - Build platforms (default: linux/amd64,linux/arm64)"
	@echo ""
	@echo "Version Information:"
	@echo "  The build system automatically injects the following:"
	@echo "  - Service Name        - From SERVICE_NAME variable"
	@echo "  - Git Version         - From git describe --tags"
	@echo "  - Git Commit          - From git rev-parse HEAD"
	@echo "  - Git Branch          - From git branch --show-current"
	@echo "  - Git Tree State      - clean/dirty based on git status"
	@echo "  - Build Date          - ISO8601 timestamp"