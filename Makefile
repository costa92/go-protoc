# Build all by default, even if it's not first
.DEFAULT_GOAL := help

# ==============================================================================
# Includes

include scripts/make-rules/common.mk # make sure include common.mk at the first include line
include scripts/make-rules/all.mk

# --- Variáveis ---
# Ferramentas
GO ?= go

API_DIR = pkg/api
BINARY_PATH = ./bin/$(BINARY_NAME)

# 处理 proto 文件路径
PROTO_DIRS := pkg/api/user/v1
PROTO_FILES := $(foreach dir,$(PROTO_DIRS),$(wildcard $(dir)/*.proto))

.PHONY: all proto clean

all: proto

proto: ## Generate protobuf code
	@echo ">> Generating protobuf code..."
	@if ! which buf > /dev/null; then \
		echo "buf is not installed. Installing..."; \
		go install github.com/bufbuild/buf/cmd/buf@latest; \
	fi
	buf generate

.PHONY: swagger
swagger: ## Generate swagger documentation using buf
	buf generate

.PHONY: swagger.serve
serve-swagger: ## Serve swagger spec and docs at 65534.
	@$(MAKE) swagger.serve

clean:
	@echo ">> Limpando arquivos gerados e binários..."
	rm -f $(BINARY_PATH)
	find $(API_DIR) -name "*.pb.go" -exec rm -f {} +
	find $(API_DIR) -name "*.pb.gw.go" -exec rm -f {} +
	find $(API_DIR) -name "*.swagger.json" -exec rm -f {} +
	find $(API_DIR) -name "*.validate.pb.go" -exec rm -f {} +

.PHONY: install-tools
install-tools: ## Install CI-related tools. Install all tools by specifying `A=1`.
	$(MAKE) install.ci
	if [[ "$(A)" == 1 ]]; then                                             \
		$(MAKE) _install.other ;                                            \
	fi

.PHONY: install-protoc-gen-validate
install-protoc-gen-validate: ## Install protoc-gen-validate.
	@$(MAKE) _install.protoc-gen-validate

.PHONY: test
test: ## 运行单元测试
	@echo ">> 运行单元测试"
	@go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test ## 生成测试覆盖率报告
	@echo ">> 生成测试覆盖率报告"
	@go tool cover -html=coverage.out

.PHONY: lint
lint: ## 运行代码质量检查
	@echo ">> 运行代码质量检查"
	@golangci-lint run ./...

.PHONY: fmt
fmt: ## 格式化代码
	@echo ">> 格式化代码"
	@gofmt -s -w .
	@goimports -w .

.PHONY: vet
vet: ## 代码静态检查
	@echo ">> 代码静态检查"
	@go vet ./...

.PHONY: mod-tidy
mod-tidy: ## 整理依赖
	@echo ">> 整理Go模块依赖"
	@go mod tidy

# 添加新的命令
.PHONY: run-api
run-api: ## 运行 API 服务器
	@echo ">> 启动 API 服务器"
	@go run cmd/apiserver/main.go

.PHONY: gen-swagger-docs
gen-swagger-docs: ## 生成 Swagger 文档
	@echo ">> 生成 Swagger 文档"
	@go run cmd/gen-swaggertype-docs/swagger_type_docs.go -s $(TYPE_SRC) -f $(FUNC_DEST)

# 生成 Wire 代码
.PHONY: wire
wire:
	cd internal/apiserver && wire

