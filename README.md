# go-protoc

本项目采用 [Golang 标准项目布局](https://github.com/golang-standards/project-layout)。

## 目录结构

```
├── cmd/            # 各个主程序（可执行文件）入口
├── pkg/            # 可被外部项目引用的库代码
├── internal/       # 仅限本项目内部使用的代码
├── api/            # API 定义（如 Protobuf、OpenAPI 等）
├── configs/        # 配置文件
├── scripts/        # 各类运维脚本
├── build/          # 打包与持续集成相关文件
├── deployments/    # 部署相关文件（如 Docker、K8s）
├── test/           # 额外的外部测试代码
├── go.mod
├── go.sum
└── README.md
```

## 说明
- `cmd/`：每个子目录对应一个可编译的主程序。
- `pkg/`：对外可用的库代码。
- `internal/`：仅限本项目内部使用的代码。
- `api/`：API 协议文件。
- `configs/`：配置文件。
- `scripts/`：脚本和工具。
- `build/`：持续集成和打包相关。
- `deployments/`：部署相关文件。
- `test/`：外部测试代码。 