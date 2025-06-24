# go-protoc

## 安装 tools

```makefile
make tools.install.*
```

## 重命名项目模块路径

如果您需要更改项目的 Go 模块路径 (例如，从 `github.com/old/project` 到 `github.com/new/project`)，可以使用 `rename-project` Make 目标。

**使用方法:**

```bash
make rename-project OLD_PATH=<current_module_path> NEW_PATH=<new_module_path>
```

**参数:**

*   `OLD_PATH`: 当前项目的 Go 模块路径。
*   `NEW_PATH`: 您希望使用的新 Go 模块路径。

**示例:**

假设您想将项目模块路径从 `github.com/costa92/go-protoc/v2` 更改为 `github.com/costa92/go-protoc/v3`，您可以运行：

```bash
make rename-project OLD_PATH=github.com/costa92/go-protoc/v2 NEW_PATH=github.com/costa92/go-protoc/v3
```

**注意:** 此命令会修改项目中的多个文件，包括 `go.mod`, `*.go` 文件, `*.sh` 文件等。执行后，请仔细检查更改并运行 `go mod tidy`。
