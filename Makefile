PROTOC_GEN_GO := $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(shell go env GOPATH)/bin/protoc-gen-go-grpc
PROTOC_GEN_GRPC_GATEWAY := $(shell go env GOPATH)/bin/protoc-gen-grpc-gateway
GOOGLEAPIS := $(shell go env GOPATH)/pkg/mod/github.com/googleapis/googleapis@*/

PROTO_DIRS := pkg/api/helloworld/v1 pkg/api/helloworld/v2
PROTO_FILES := $(foreach dir,$(PROTO_DIRS),$(wildcard $(dir)/*.proto))

.PHONY: all proto clean

all: proto

proto:
	protoc -I. \
		-Ipkg/third-party/proto \
		--go_out . --go_opt paths=source_relative \
		--go-grpc_out . --go-grpc_opt paths=source_relative \
		--grpc-gateway_out . --grpc-gateway_opt paths=source_relative \
		$(PROTO_FILES)

clean:
	for dir in $(PROTO_DIRS); do rm -f $$dir/*.pb.go $$dir/*_grpc.pb.go $$dir/*_gw.go; done