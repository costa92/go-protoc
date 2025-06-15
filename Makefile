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


.PHONY: run-api
run-api:
	cd cmd/apiserver && go run main.go

.PHONY: proto
proto:
	buf generate

.PHONY: build
build:
	go build -o bin/apiserver cmd/apiserver/main.go


.PHONY: apidiff
apidiff: tools.verify.go-apidiff ## Run the go-apidiff to verify any API differences compared with origin/master.
	@go-apidiff master --compare-imports --print-compatible --repo-path=.

.PHONY: tidy
tidy:
	@$(GO) mod tidy