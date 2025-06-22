#!/bin/bash

set -e

# 清理旧的生成文件
rm -rf api/apiserver/v1/*.pb.go
rm -rf api/apiserver/v1/*.pb.gw.go
rm -rf api/apiserver/v1/*.pb.validate.go
rm -rf api/openapi/*.swagger.json

# 生成 protobuf 代码，只处理 apiserver 目录
echo "Generating protobuf code and documentation for apiserver..."
buf generate --path pkg/api/apiserver

# 格式化生成的代码
go fmt ./...

echo "Code generation completed!"
echo "Generated files:"
echo "  - Go code: pkg/api/apiserver/"
echo "  - HTML documentation: docs/index.html"
echo "  - Markdown documentation: docs/README.md"
echo ""
echo "To view the HTML documentation, open docs/index.html in your browser"
