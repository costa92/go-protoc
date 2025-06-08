# 项目完善规划

## 1. 配置管理 (Configuration)

### 当前状态

- 服务地址和端口等关键配置是硬编码在代码中的（例如 ":8090", ":8091"）

### 缺失环节

- 缺少一个灵活的配置管理机制，导致在不同环境（开发、测试、生产）部署时需要修改代码

### 规划建议

引入一个支持从文件和环境变量加载配置的库（如 Viper），将所有可变配置集中管理

### 具体步骤

1. 在 `configs/` 目录下创建一个默认的配置文件，例如 `config.yaml`
2. 定义一个与配置文件结构对应的 Go struct
3. 在 `cmd/apiserver/main.go` 中，使用配置库加载配置，替换所有硬编码的值
4. 支持通过环境变量覆盖配置文件中的值，以适应容器化部署

## 2. 可观测性 (Observability) 的深化

### 当前状态

- 已经集成了结构化日志和链路追踪

### 缺失环节

- **Metrics (指标)**：缺少关键的业务和系统指标，例如 QPS、请求延迟、错误率等。没有指标就无法进行有效的监控和告警
- **日志与追踪的关联**：当前的日志信息中没有包含 TraceID，这使得在分布式系统中根据一条日志追溯完整的调用链变得困难

### 规划建议

- 引入 Prometheus 作为指标监控解决方案，它是云原生领域的事实标准
- 改造日志中间件，从请求的 context 中提取 TraceID 并添加到每一条日志中

### 具体步骤

1. 添加 `prometheus/client_golang` 依赖
2. 在 HTTP 服务中添加一个 `/metrics` 路由，用于暴露 Prometheus 指标
3. 创建并注册自定义指标（例如，使用 Counter 和 Histogram）
4. 修改 `pkg/middleware/http/logging.go` 和 `pkg/middleware/grpc/logging.go`，从 context 中提取 TraceID 并作为 zap 的一个字段进行记录

## 3. 代码质量与健壮性

### 当前状态

- 项目结构清晰，但缺少自动化保障
- 有一个基础的测试文件 `app_test.go`

### 缺失环节

- **单元测试**：核心业务逻辑（如 `internal/helloworld/service`）缺少单元测试，无法保证代码质量和方便地进行重构
- **静态代码检查 (Linter)**：没有配置 Linter，容易引入潜在的 bug 和不规范的代码风格

### 规划建议

- 为关键的业务逻辑和服务编写单元测试
- 引入 golangci-lint 作为代码检查工具，并配置规则

### 具体步骤

1. 为 `internal/helloworld/service` 中的 SayHello 等方法编写 `_test.go` 文件和测试用例
2. 在项目根目录添加 `.golangci.yml` 配置文件
3. 在本地开发和 CI/CD 流程中集成 `golangci-lint run` 命令

## 4. API 强化

### 当前状态

- 使用 Protobuf 定义 API，并通过 gRPC-Gateway 提供 HTTP 接口

### 缺失环节

- **请求参数校验**：目前所有请求的参数校验都需要在业务逻辑中手动编写，这既繁琐又容易出错

### 规划建议

引入 protoc-gen-validate 插件，它允许您直接在 .proto 文件中通过注解定义参数的校验规则

### 具体步骤

1. 安装 protoc-gen-validate
2. 在 .proto 文件中为字段添加校验规则（例如 `[(validate.rules).string.min_len = 1]`）
3. 重新生成 `*.pb.go` 和 `*.pb.validate.go` 文件
4. 在 gRPC 拦截器或 HTTP 中间件中，添加一个校验环节，自动对所有请求进行校验

## 5. 构建与部署

### 当前状态

- 项目布局中包含了 `build/` 和 `deployments/` 目录，但它们目前是空的
- 项目没有提供容器化部署的方案

### 缺失环节

- **容器化构建 (Dockerfile)** 和 **持续集成 (CI)**

### 规划建议

- 提供一个用于生产环境的、经过优化的多阶段 Dockerfile
- 建立一个基础的 CI 流水线（例如使用 GitHub Actions），自动化执行代码检查、测试和构建

### 具体步骤

1. 在项目根目录编写 Dockerfile，使用多阶段构建来减小最终镜像的体积
2. 编写 `.dockerignore` 文件，排除不必要的文件
3. 在 `.github/workflows/` 目录下创建一个 CI 配置文件（如 `go.yml`），定义构建、检查和测试的作业
