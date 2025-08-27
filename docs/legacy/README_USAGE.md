# 文档生成使用说明

本项目使用 buf 和 protoc-gen-doc 插件来生成 protobuf 文档。

## 生成的文档格式

- **HTML**: `docs/index.html` - 可在浏览器中查看的交互式文档
- **Markdown**: `docs/README.md` - 适合在 GitHub 等平台显示的文档
- **JSON**: `docs/docs.json` - 机器可读的 JSON 格式文档
- **DocBook XML**: `docs/docs.xml` - 可用于进一步处理的 XML 格式

## 使用方法

### 1. 生成代码和文档（推荐）
```bash
./generate.sh
```
这会同时生成 Go 代码和基本文档（HTML 和 Markdown）。

### 2. 只生成文档
```bash
./generate-docs.sh
```
这会生成所有格式的文档（HTML、Markdown、JSON、XML）。

### 3. 手动生成
```bash
# 生成代码和基本文档
buf generate --path pkg/api/apiserver

# 只生成文档
buf generate --template buf.gen.docs.yaml --path pkg/api/apiserver
```

## 查看文档

### HTML 文档
在浏览器中打开 `docs/index.html`：
```bash
open docs/index.html
```

### Markdown 文档
直接查看 `docs/README.md` 文件，或在支持 Markdown 的编辑器中打开。

## 自定义文档

如需自定义文档格式，可以修改：
- `buf.gen.yaml` - 主要的生成配置
- `buf.gen.docs.yaml` - 专门的文档生成配置

支持的文档格式：
- `html` - HTML 格式
- `markdown` - Markdown 格式
- `json` - JSON 格式
- `docbook` - DocBook XML 格式

## 依赖

确保已安装以下工具：
- `buf` - Protocol Buffer 工具
- `protoc-gen-doc` - 文档生成插件

安装 protoc-gen-doc：
```bash
go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
```