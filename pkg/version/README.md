# Version 包文档

`pkg/version` 包提供了一个全面的版本信息管理系统，支持构建时版本注入、运行时版本查询以及多种输出格式。这个包设计用于 Go 微服务架构，提供统一的版本管理和显示功能。

## 目录

- [设计理念](#设计理念)
- [架构设计](#架构设计)
- [核心特性](#核心特性)
- [实现方式](#实现方式)
- [使用指南](#使用指南)
- [最佳实践](#最佳实践)
- [扩展机制](#扩展机制)

## 设计理念

### 1. 构建时注入，运行时获取

版本信息在构建时通过 Go 的 `-ldflags` 机制注入，运行时可以快速获取，避免了版本信息的硬编码问题：

```go
// 构建时注入的变量
var gitVersion = "v0.0.0-master+$Format:%h$"  // 默认值，构建时覆盖
var buildDate = "1970-01-01T00:00:00Z"        // 默认值，构建时覆盖
```

### 2. 多维度版本信息

不仅包含传统的版本号，还包含完整的构建环境信息：

- **Git 信息**: 版本、提交 ID、分支、仓库状态
- **构建信息**: 构建时间、服务名称
- **运行环境**: Go 版本、编译器、平台信息

### 3. 多种输出格式

支持不同场景的版本信息展示需求：

- **简化格式**: `String()` - 仅版本号，适用于快速查看
- **JSON 格式**: `ToJSON()` - 结构化数据，适用于 API 和日志
- **表格格式**: `Text()` - 详细表格，适用于人类阅读

### 4. 动态版本管理

支持运行时动态设置版本信息，用于特殊的部署场景：

```go
// 运行时动态设置版本
err := SetDynamicVersion("v1.2.3-hotfix.1")
```

## 架构设计

### 核心组件关系

```
构建系统 (ldflags) → 版本变量 → Info 结构体 → 多种输出格式
                         ↓              ↓
                    动态版本管理    命令行支持
```

### 文件组织

```
pkg/version/
├── doc.go          # 包文档和导入声明
├── version.go      # 核心版本信息结构和功能
├── dynamic.go      # 动态版本设置功能
├── flag.go         # 命令行标志支持
└── version_test.go # 完整的单元测试
```

### 数据流转

```
1. 构建时 ldflags 注入 → 包级变量
2. Get() 函数 → 创建 Info 结构体
3. Info 方法 → 多种格式输出
4. 动态版本 → atomic.Value 存储 → 覆盖静态版本
```

## 核心特性

### 1. 版本信息结构

```go
type Info struct {
    GitVersion   string `json:"gitVersion"`   // Git 版本标签
    GitCommit    string `json:"gitCommit"`    // Git 提交 SHA
    GitTreeState string `json:"gitTreeState"` // Git 仓库状态 (clean/dirty)
    GitBranch    string `json:"gitBranch"`    // Git 分支名
    BuildDate    string `json:"buildDate"`    // ISO8601 格式构建时间
    ServiceName  string `json:"serviceName"`  // 服务名称
    GoVersion    string `json:"goVersion"`    // Go 运行时版本
    Compiler     string `json:"compiler"`     // Go 编译器
    Platform     string `json:"platform"`     // 操作系统/架构
}
```

### 2. 构建时版本注入

通过 ldflags 在构建时注入实际的版本信息：

```bash
go build -ldflags "
    -X 'github.com/costa92/go-protoc/v2/pkg/version.serviceName=myservice'
    -X 'github.com/costa92/go-protoc/v2/pkg/version.gitVersion=v1.0.0'
    -X 'github.com/costa92/go-protoc/v2/pkg/version.gitCommit=abc12345'
    -X 'github.com/costa92/go-protoc/v2/pkg/version.gitBranch=main'
    -X 'github.com/costa92/go-protoc/v2/pkg/version.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'
" ./cmd/myservice
```

### 3. 动态版本管理

支持运行时修改版本信息，用于容器环境或特殊部署场景：

```go
// 验证并设置动态版本
if err := SetDynamicVersion("v1.2.3-hotfix.1"); err != nil {
    log.Fatal("Invalid dynamic version: ", err)
}

// 动态版本会覆盖构建时的静态版本
info := Get() // 返回动态设置的版本信息
```

### 4. 命令行标志支持

内置的命令行标志支持，集成到标准的 flag 解析流程：

```go
import "github.com/costa92/go-protoc/v2/pkg/version"

func main() {
    // 注册版本标志
    version.AddFlags(pflag.CommandLine)
    pflag.Parse()

    // 检查并处理版本请求
    version.PrintAndExitIfRequested()

    // 应用程序继续运行...
}
```

## 实现方式

### 1. 版本变量的注入机制

```go
// 包级变量，默认值会被构建时 ldflags 覆盖
var (
    gitVersion   = "v0.0.0-master+$Format:%h$"  // Git 版本
    buildDate    = "1970-01-01T00:00:00Z"       // 构建时间
    gitCommit    = "$Format:%H$"                // Git 提交
    gitTreeState = ""                           // Git 状态
    gitBranch    = "unknown"                    // Git 分支
    serviceName  = "apiserver"                  // 服务名称
)
```

这些变量使用了特殊的默认值：
- `$Format:%h$` 和 `$Format:%H$` 是 Git 导出时的占位符
- `1970-01-01T00:00:00Z` 是 Unix 纪元时间，表示未设置
- 构建系统会在编译时替换这些默认值

### 2. 信息聚合机制

`Get()` 函数是核心入口，负责聚合所有版本信息：

```go
func Get() Info {
    return Info{
        ServiceName:  serviceName,
        GitVersion:   getEffectiveVersion(),  // 考虑动态版本
        GitCommit:    gitCommit,
        GitBranch:    gitBranch,
        GitTreeState: gitTreeState,
        BuildDate:    buildDate,
        GoVersion:    runtime.Version(),      // 运行时获取
        Compiler:     runtime.Compiler,      // 运行时获取
        Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH), // 运行时获取
    }
}
```

### 3. 动态版本的原子性实现

使用 `atomic.Value` 确保动态版本的并发安全：

```go
var dynamicGitVersion atomic.Value

func init() {
    // 初始化为静态版本
    dynamicGitVersion.Store(gitVersion)
}

func SetDynamicVersion(version string) error {
    if err := ValidateDynamicVersion(version); err != nil {
        return err
    }
    dynamicGitVersion.Store(version)  // 原子性存储
    return nil
}

func getEffectiveVersion() string {
    if val := dynamicGitVersion.Load(); val != nil {
        return val.(string)
    }
    return gitVersion
}
```

### 4. 语义版本验证

动态版本设置包含严格的语义版本验证：

```go
func ValidateDynamicVersion(dynamicVersion string) error {
    // 1. 非空验证
    if len(dynamicVersion) == 0 {
        return fmt.Errorf("version must not be empty")
    }

    // 2. 语义版本格式验证
    vRuntime, err := utilversion.ParseSemantic(dynamicVersion)
    if err != nil {
        return err
    }

    // 3. 主版本兼容性验证
    vDefault, _ := utilversion.ParseSemantic(defaultVersion)
    if vRuntime.Major() != vDefault.Major() ||
       vRuntime.Minor() != vDefault.Minor() ||
       vRuntime.Patch() != vDefault.Patch() {
        return fmt.Errorf("version mismatch")
    }

    return nil
}
```

### 5. 多格式输出实现

#### 简化输出 (String)
```go
func (info Info) String() string {
    return info.GitVersion  // 仅返回版本号
}
```

#### JSON 输出 (ToJSON)
```go
func (info Info) ToJSON() string {
    s, _ := json.Marshal(info)
    return string(s)
}
```

#### 表格输出 (Text)
```go
func (info Info) Text() string {
    table := uitable.New()
    table.RightAlign(0)
    table.MaxColWidth = 80
    table.Separator = " "

    // 添加所有字段到表格
    table.AddRow("serviceName:", info.ServiceName)
    table.AddRow("gitVersion:", info.GitVersion)
    // ... 其他字段

    return table.String()
}
```

## 使用指南

### 1. 基础使用

```go
package main

import (
    "fmt"
    "github.com/costa92/go-protoc/v2/pkg/version"
)

func main() {
    // 获取版本信息
    info := version.Get()

    // 简单输出
    fmt.Printf("Version: %s\n", info.String())

    // JSON 输出
    fmt.Printf("JSON: %s\n", info.ToJSON())

    // 详细表格输出
    fmt.Printf("Details:\n%s\n", info.Text())
}
```

### 2. 集成到 Web 服务

```go
package main

import (
    "encoding/json"
    "net/http"

    "github.com/costa92/go-protoc/v2/pkg/version"
)

func versionHandler(w http.ResponseWriter, r *http.Request) {
    info := version.Get()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(info)
}

func main() {
    http.HandleFunc("/version", versionHandler)

    // 启动时打印版本信息
     .Printf("Starting service %s version %s\n",
        version.Get().ServiceName,
        version.Get().String())

    http.ListenAndServe(":8080", nil)
}
```

### 3. 命令行集成

```go
package main

import (
    "flag"
    "fmt"

    "github.com/costa92/go-protoc/v2/pkg/version"
    "github.com/spf13/pflag"
)

func main() {
    // 添加版本标志
    version.AddFlags(pflag.CommandLine)

    // 解析命令行参数
    pflag.Parse()

    // 如果请求版本信息，打印并退出
    version.PrintAndExitIfRequested()

    // 应用程序逻辑
    fmt.Println("Application running...")
}
```

支持的命令行用法：
```bash
# 简单版本输出
./myapp --version

# 详细版本输出
./myapp --version=raw

# 布尔标志形式
./myapp --version=true
./myapp --version=false
```

### 4. 动态版本管理

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/costa92/go-protoc/v2/pkg/version"
)

func main() {
    // 检查环境变量中的动态版本
    if dynamicVer := os.Getenv("DYNAMIC_VERSION"); dynamicVer != "" {
        if err := version.SetDynamicVersion(dynamicVer); err != nil {
            log.Fatalf("Failed to set dynamic version: %v", err)
        }
        fmt.Printf("Dynamic version set to: %s\n", dynamicVer)
    }

    // 正常使用版本信息
    info := version.Get()
    fmt.Printf("Running version: %s\n", info.String())
}
```

### 5. 日志集成

```go
package main

import (
    "github.com/costa92/go-protoc/v2/pkg/logger"
    "github.com/costa92/go-protoc/v2/pkg/version"
)

func main() {
    // 在日志中记录版本信息
    info := version.Get()
    logger.Infow("Application starting",
        "service", info.ServiceName,
        "version", info.GitVersion,
        "branch", info.GitBranch,
        "commit", info.GitCommit[:8],
        "build_date", info.BuildDate,
    )

    // 应用程序逻辑...
}
```

## 最佳实践

### 1. 构建系统集成

**Makefile 示例**：
```makefile
# 版本信息变量
SERVICE_NAME ?= myservice
VERSION_PKG = github.com/costa92/go-protoc/v2/pkg/version

# Git 信息获取
GIT_VERSION := $(shell git describe --tags --always --dirty)
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_BRANCH := $(shell git branch --show-current)
GIT_TREE_STATE := $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# 构建标志
LDFLAGS := -X '$(VERSION_PKG).serviceName=$(SERVICE_NAME)' \
           -X '$(VERSION_PKG).gitVersion=$(GIT_VERSION)' \
           -X '$(VERSION_PKG).gitCommit=$(GIT_COMMIT)' \
           -X '$(VERSION_PKG).gitBranch=$(GIT_BRANCH)' \
           -X '$(VERSION_PKG).gitTreeState=$(GIT_TREE_STATE)' \
           -X '$(VERSION_PKG).buildDate=$(BUILD_DATE)'

# 构建目标
build:
	go build -ldflags "$(LDFLAGS)" ./cmd/myservice
```

### 2. CI/CD 集成

**GitHub Actions 示例**：
```yaml
name: Build and Release

on:
  push:
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
      with:
        fetch-depth: 0  # 获取完整 Git 历史

    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21

    - name: Build with version info
      env:
        SERVICE_NAME: myservice
      run: |
        VERSION=$(git describe --tags --always --dirty)
        COMMIT=$(git rev-parse HEAD)
        BRANCH=$(git branch --show-current)
        DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

        go build -ldflags "
          -X 'github.com/costa92/go-protoc/v2/pkg/version.serviceName=${SERVICE_NAME}'
          -X 'github.com/costa92/go-protoc/v2/pkg/version.gitVersion=${VERSION}'
          -X 'github.com/costa92/go-protoc/v2/pkg/version.gitCommit=${COMMIT}'
          -X 'github.com/costa92/go-protoc/v2/pkg/version.gitBranch=${BRANCH}'
          -X 'github.com/costa92/go-protoc/v2/pkg/version.buildDate=${DATE}'
        " ./cmd/myservice
```

### 3. 容器化部署

**Dockerfile 示例**：
```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

# 构建参数
ARG SERVICE_NAME=myservice
ARG VERSION
ARG COMMIT
ARG BRANCH
ARG BUILD_DATE

# 构建带版本信息的二进制文件
RUN go build -ldflags "\
    -X 'github.com/costa92/go-protoc/v2/pkg/version.serviceName=${SERVICE_NAME}' \
    -X 'github.com/costa92/go-protoc/v2/pkg/version.gitVersion=${VERSION}' \
    -X 'github.com/costa92/go-protoc/v2/pkg/version.gitCommit=${COMMIT}' \
    -X 'github.com/costa92/go-protoc/v2/pkg/version.gitBranch=${BRANCH}' \
    -X 'github.com/costa92/go-protoc/v2/pkg/version.buildDate=${BUILD_DATE}' \
    " -o myservice ./cmd/myservice

# 运行阶段
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/myservice /usr/local/bin/

# 设置版本标签
LABEL version="${VERSION}"
LABEL commit="${COMMIT}"
LABEL branch="${BRANCH}"

ENTRYPOINT ["myservice"]
```

### 4. 监控集成

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/costa92/go-protoc/v2/pkg/version"
)

func init() {
    // 注册版本信息作为 Prometheus 指标
    info := version.Get()

    versionGauge := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "application_info",
            Help: "Application version information",
        },
        []string{"service", "version", "branch", "commit"},
    )

    versionGauge.WithLabelValues(
        info.ServiceName,
        info.GitVersion,
        info.GitBranch,
        info.GitCommit[:8],
    ).Set(1)

    prometheus.MustRegister(versionGauge)
}
```

### 5. 版本兼容性检查

```go
package main

import (
    "fmt"
    "strings"

    "github.com/costa92/go-protoc/v2/pkg/version"
)

func checkCompatibility() error {
    info := version.Get()

    // 检查最低版本要求
    if strings.HasPrefix(info.GitVersion, "v0.") {
        return fmt.Errorf("development version detected: %s", info.GitVersion)
    }

    // 检查分支兼容性
    if info.GitBranch == "experimental" {
        fmt.Printf("Warning: running experimental branch\n")
    }

    // 检查构建状态
    if info.GitTreeState == "dirty" {
        fmt.Printf("Warning: built from modified source\n")
    }

    return nil
}
```

## 扩展机制

### 1. 自定义输出格式

```go
package version

import (
    "fmt"
    "gopkg.in/yaml.v3"
)

// YAML 输出格式扩展
func (info Info) ToYAML() (string, error) {
    data, err := yaml.Marshal(info)
    if err != nil {
        return "", err
    }
    return string(data), nil
}

// 简化的服务信息输出
func (info Info) ServiceInfo() string {
    return fmt.Sprintf("%s:%s@%s",
        info.ServiceName,
        info.GitVersion,
        info.GitBranch)
}
```

### 2. 版本比较功能

```go
package version

import (
    "github.com/Masterminds/semver/v3"
)

// 版本比较扩展
func (info Info) Compare(other string) (int, error) {
    v1, err := semver.NewVersion(info.GitVersion)
    if err != nil {
        return 0, err
    }

    v2, err := semver.NewVersion(other)
    if err != nil {
        return 0, err
    }

    return v1.Compare(v2), nil
}

// 检查是否为预发布版本
func (info Info) IsPrerelease() bool {
    v, err := semver.NewVersion(info.GitVersion)
    if err != nil {
        return false
    }
    return v.Prerelease() != ""
}
```

### 3. 环境特定的版本处理

```go
package version

import (
    "os"
    "fmt"
)

// 环境感知的版本信息
func GetWithEnvironment() Info {
    info := Get()

    // 添加环境信息
    if env := os.Getenv("ENVIRONMENT"); env != "" {
        info.ServiceName = fmt.Sprintf("%s-%s", info.ServiceName, env)
    }

    // 添加部署信息
    if deployment := os.Getenv("DEPLOYMENT_ID"); deployment != "" {
        info.GitVersion = fmt.Sprintf("%s+%s", info.GitVersion, deployment)
    }

    return info
}
```

这个版本包通过精心设计的架构提供了全面的版本管理功能，支持从开发到生产的完整生命周期。它的模块化设计使得可以根据具体需求进行扩展和定制，同时保持了简单易用的 API 接口。