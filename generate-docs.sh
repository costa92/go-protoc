#!/bin/bash

# 生成 protobuf 文档
echo "Generating protobuf documentation..."

# 使用专门的文档配置文件
buf generate --template buf.gen.docs.yaml --path pkg/api/apiserver

echo "Documentation generation completed!"
echo "Generated documentation files:"
echo "  - HTML: docs/index.html"
echo "  - Markdown: docs/README.md"
echo "  - JSON: docs/docs.json"
echo "  - DocBook XML: docs/docs.xml"
echo ""
echo "To view the HTML documentation, run:"
echo "  open docs/index.html"