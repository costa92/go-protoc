
# --- Variáveis ---
# Ferramentas
PROTOC ?= protoc
GO ?= go

API_DIR = pkg/api
BINARY_PATH = ./bin/$(BINARY_NAME)

PROTOC_GEN_GO := $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(shell go env GOPATH)/bin/protoc-gen-go-grpc
PROTOC_GEN_GRPC_GATEWAY := $(shell go env GOPATH)/bin/protoc-gen-grpc-gateway
GOOGLEAPIS := $(shell go env GOPATH)/pkg/mod/github.com/googleapis/googleapis@*/

PROTO_DIRS := pkg/api/helloworld/v1 pkg/api/helloworld/v2
PROTO_FILES := $(foreach dir,$(PROTO_DIRS),$(wildcard $(dir)/*.proto))

.PHONY: all proto clean

all: proto

proto:
	$(PROTOC) -I. \
		-Ithird_party/ \
		--go_out . --go_opt paths=source_relative \
		--go-grpc_out . --go-grpc_opt paths=source_relative \
		--grpc-gateway_out . --grpc-gateway_opt paths=source_relative \
		$(PROTO_FILES)

swagger:
	@echo ">> Gerando arquivos JSON do OpenAPIv2 (Swagger)..."
	@# Verifica se 'protoc-gen-openapiv2' está instalado, instala se não estiver.
	@command -v protoc-gen-openapiv2 >/dev/null 2>&1 || \
		(echo "   'protoc-gen-openapiv2' não encontrado, instalando..."; \
		go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest)
	$(PROTOC) -I . \
		-Ithird_party/ \
		--openapiv2_out=. \
		--openapiv2_opt=logtostderr=true \
		$(PROTO_FILES)
	@echo "Arquivos Swagger gerados com sucesso."

clean:
	@echo ">> Limpando arquivos gerados e binários..."
	rm -f $(BINARY_PATH)
	find $(API_DIR) -name "*.pb.go" -exec rm -f {} +
	find $(API_DIR) -name "*.pb.gw.go" -exec rm -f {} +
	find $(API_DIR) -name "*.swagger.json" -exec rm -f {} +
