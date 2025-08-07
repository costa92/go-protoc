# go-protoc 开发文档

> 该文档由 `make docs.dev` 自动生成，更新时间: {{.GenerateTime}}

## 📋 项目概览

- **项目名称**: go-protoc
- **架构**: Clean Architecture + Kratos v2
- **语言**: Go 1.24.0+
- **API**: Protocol Buffers + gRPC + HTTP REST
- **模块路径**: github.com/costa92/go-protoc/v2

## 🚀 快速开始

### 环境要求
- Go 1.24.0+ (安装命令: `make tools.install.go`)
- Protocol Buffers 编译器 (安装命令: `make tools.install.buf`)
- Docker & Docker Compose (用于开发环境)

### 本地开发
```bash
# 1. 安装开发工具
make install-tools

# 2. 启动依赖服务
make run-redis

# 3. 生成代码
make generate

# 4. 运行API服务器
make run-api
```

## 🏗️ 项目结构

```
{{.ModulePath}}/
├── cmd/apiserver/          # 应用入口
├── internal/
│   └── apiserver/
│       ├── handler/        # API处理层
│       ├── biz/           # 业务逻辑层
│       └── store/         # 数据访问层
├── pkg/                   # 公共包
├── api/                   # 自动生成的API文档
├── configs/               # 配置文件
└── docs/                  # 开发文档
```

## 🔧 开发命令

### 核心命令
```bash
make help                        # 查看所有可用命令
make run-api                     # 运行开发服务器
make build                       # 构建二进制文件
make test                        # 运行测试套件
```

### 代码生成
```bash
make generate                    # 生成protobuf相关代码
make wire                        # 重新生成依赖注入
make fmt                         # 格式化代码并排序imports
```

### 依赖管理
```bash
make tidy                        # 清理和整理go.mod
make tools.install.%             # 安装特定工具
make install-tools               # 安装所有开发工具
```

## 📁 关键目录

| 目录 | 用途 | 日常开发重点 |
|---|---|---|
| `cmd/apiserver/` | 程序入口 | 启动配置和main函数 |
| `internal/apiserver/` | 业务核心 | 添加新功能的主要工作区 |
| `pkg/api/` | API定义 | protobuf文件和生成代码 |
| `pkg/errorsx/` | 错误处理 | 自定义错误码和国际化 |
| `pkg/db/` | 数据库 | 数据库连接和迁移 |
| `configs/` | 配置 | 环境配置文件 |

## 🔍 调试与监控

### 健康检查
- **健康状态**: `GET /health`
- **性能指标**: `GET /metrics` (Prometheus格式)
- **调用链路**: Jaeger集成 (http://localhost:16686)

### 开发调试
```bash
# 查看所有可用环境变量
go run cmd/apiserver/main.go -h

# 指定配置文件
make run-api CONFIG=configs/apiserver_local.yaml

# 启用调试模式
DEBUG=true make run-api
```

## 📊 测试

### 单元测试
```bash
go test ./...                    # 运行所有测试
go test -v ./pkg/errorsx/       # 针对特定包测试
go test -run TestUser           # 运行特定测试
go test ./... -bench=.          # 基准测试
```

### 集成测试
```bash
# 启动测试环境
make run-redis

# 运行集成测试
go test -tags=integration ./...
```

## 📝 添加新功能流程

### 添加新API端点的完整流程

1. **定义API** (`pkg/api/apiserver/v1/*.proto`)
   ```protobuf
   service ApiServer {
     rpc GetUser(GetUserRequest) returns (GetUserResponse) {}
   }
   ```

2. **生成代码**
   ```bash
   make generate
   ```

3. **实现处理逻辑** (`internal/apiserver/handler/`)

4. **注册依赖** 运行 `make wire`

5. **测试验证**

### 添加新错误码

1. **在proto中定义**: `errors.proto`
2. **重新生成**: `make generate`
3. **使用**: `errorsx.NewCode("MyError")`

## 🌐 国际化

- **错误消息**: 支持中英文
- **语言检测**: 基于Accept-Language头
- **添加新语言**: 编辑 `pkg/i18n/translations/`

## 📚 进阶配置

### 数据库配置
```yaml
# configs/apiserver.yaml
mysql:
  host: "localhost"
  port: 3306
  database: "myapp"
  username: "user"
  password: "password"
```

### Redis配置
```yaml
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
```

## 🔗 相关资源

- **GitHub**: {{.GitHub}}
- **API文档**: [OpenAPI]({{.OpenAPIDoc}})
- **错误码文档**: [Error Codes]({{.ErrorCodesDoc}})
- **架构设计**: [Architecture](./guide/zh-CN/errors-architecture.md)

---

*此文档由 `make docs.dev` 自动生成，运行该命令可更新此文档