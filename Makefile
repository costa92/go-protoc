
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

.PHONY: run
run: