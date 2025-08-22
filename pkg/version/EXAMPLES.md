# Version 包使用示例

本文档提供了 `pkg/version` 包的详细使用示例，涵盖了各种实际应用场景和最佳实践。

## 目录

- [基础使用示例](#基础使用示例)
- [构建系统集成](#构建系统集成)
- [Web 服务集成](#web-服务集成)
- [CLI 应用集成](#cli-应用集成)
- [日志和监控集成](#日志和监控集成)
- [动态版本管理](#动态版本管理)
- [容器化部署](#容器化部署)
- [CI/CD 集成](#cicd-集成)

## 基础使用示例

### 1. 简单版本查询

```go
package main

import (
    "fmt"
    "github.com/costa92/go-protoc/v2/pkg/version"
)

func main() {
    // 获取版本信息
    info := version.Get()
    
    // 不同格式的输出
    fmt.Printf("简化版本: %s\n", info.String())
    fmt.Printf("JSON格式: %s\n", info.ToJSON())
    fmt.Printf("详细信息:\n%s\n", info.Text())
    
    // 访问具体字段
    fmt.Printf("服务名: %s\n", info.ServiceName)
    fmt.Printf("Git分支: %s\n", info.GitBranch)
    fmt.Printf("构建时间: %s\n", info.BuildDate)
    
    // 运行时环境信息
    fmt.Printf("Go版本: %s\n", info.GoVersion)
    fmt.Printf("平台: %s\n", info.Platform)
}
```

**输出示例**：
```
简化版本: v1.2.3
JSON格式: {"gitVersion":"v1.2.3","gitCommit":"abc1234567890","gitTreeState":"clean","gitBranch":"main","buildDate":"2024-01-15T10:30:45Z","serviceName":"myservice","goVersion":"go1.21.0","compiler":"gc","platform":"linux/amd64"}
详细信息:
    serviceName: myservice
     gitVersion: v1.2.3
      gitCommit: abc1234567890
     gitBranch: main
  gitTreeState: clean
     buildDate: 2024-01-15T10:30:45Z
     goVersion: go1.21.0
      compiler: gc
      platform: linux/amd64

服务名: myservice
Git分支: main
构建时间: 2024-01-15T10:30:45Z
Go版本: go1.21.0
平台: linux/amd64
```

### 2. 版本信息校验

```go
package main

import (
    "fmt"
    "log"
    "strings"
    
    "github.com/costa92/go-protoc/v2/pkg/version"
)

func validateVersion() error {
    info := version.Get()
    
    // 检查版本格式
    if !strings.HasPrefix(info.GitVersion, "v") {
        return fmt.Errorf("版本号格式错误: %s", info.GitVersion)
    }
    
    // 检查构建状态
    if info.GitTreeState == "dirty" {
        log.Printf("警告: 构建时存在未提交的修改")
    }
    
    // 检查分支
    if info.GitBranch == "unknown" {
        log.Printf("警告: 无法确定Git分支")
    }
    
    // 检查构建时间
    if info.BuildDate == "1970-01-01T00:00:00Z" {
        return fmt.Errorf("构建时间未正确设置")
    }
    
    return nil
}

func main() {
    if err := validateVersion(); err != nil {
        log.Fatalf("版本验证失败: %v", err)
    }
    
    fmt.Println("版本信息验证通过")
}
```

## 构建系统集成

### 1. Makefile 集成

```makefile
# Makefile

# 项目信息
PROJECT_NAME := myservice
VERSION_PKG := github.com/costa92/go-protoc/v2/pkg/version

# Git 信息获取
GIT_VERSION := $(shell git describe --tags --always --dirty)
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_BRANCH := $(shell git branch --show-current)
GIT_TREE_STATE := $(shell if [ -n "$$(git status --porcelain)" ]; then echo "dirty"; else echo "clean"; fi)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# 构建标志
LDFLAGS := -w -s \
    -X '$(VERSION_PKG).serviceName=$(PROJECT_NAME)' \
    -X '$(VERSION_PKG).gitVersion=$(GIT_VERSION)' \
    -X '$(VERSION_PKG).gitCommit=$(GIT_COMMIT)' \
    -X '$(VERSION_PKG).gitBranch=$(GIT_BRANCH)' \
    -X '$(VERSION_PKG).gitTreeState=$(GIT_TREE_STATE)' \
    -X '$(VERSION_PKG).buildDate=$(BUILD_DATE)'

# 构建目标
.PHONY: build
build:
	@echo "Building $(PROJECT_NAME) with version info..."
	@echo "Version: $(GIT_VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@echo "Branch: $(GIT_BRANCH)" 
	@echo "Date: $(BUILD_DATE)"
	go build -ldflags "$(LDFLAGS)" -o bin/$(PROJECT_NAME) ./cmd/$(PROJECT_NAME)

# 显示版本信息
.PHONY: version
version:
	@echo "Project: $(PROJECT_NAME)"
	@echo "Version: $(GIT_VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@echo "Branch: $(GIT_BRANCH)"
	@echo "Tree State: $(GIT_TREE_STATE)"
	@echo "Build Date: $(BUILD_DATE)"

# 多平台构建
.PHONY: build-all
build-all:
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ]; then ext=".exe"; else ext=""; fi; \
			echo "Building $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch go build \
				-ldflags "$(LDFLAGS)" \
				-o bin/$(PROJECT_NAME)-$$os-$$arch$$ext \
				./cmd/$(PROJECT_NAME); \
		done; \
	done
```

### 2. 构建脚本

```bash
#!/bin/bash
# build.sh - 高级构建脚本

set -e

# 配置
PROJECT_NAME=${PROJECT_NAME:-"myservice"}
OUTPUT_DIR=${OUTPUT_DIR:-"bin"}
VERSION_PKG="github.com/costa92/go-protoc/v2/pkg/version"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Git 仓库
check_git() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "当前目录不是Git仓库"
        exit 1
    fi
}

# 获取版本信息
get_version_info() {
    log_info "获取版本信息..."
    
    GIT_VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "unknown")
    GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    
    # 处理分支信息
    GIT_BRANCH=$(git branch --show-current 2>/dev/null)
    if [ -z "$GIT_BRANCH" ]; then
        # Detached HEAD 状态
        GIT_BRANCH=$(git describe --contains --all HEAD 2>/dev/null || echo "detached")
    fi
    
    # 检查工作树状态
    if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
        GIT_TREE_STATE="dirty"
        log_warn "工作树包含未提交的修改"
    else
        GIT_TREE_STATE="clean"
    fi
    
    BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
    
    log_info "版本信息:"
    log_info "  服务名: $PROJECT_NAME"
    log_info "  版本: $GIT_VERSION"
    log_info "  提交: ${GIT_COMMIT:0:8}"
    log_info "  分支: $GIT_BRANCH"
    log_info "  状态: $GIT_TREE_STATE"
    log_info "  构建时间: $BUILD_DATE"
}

# 构建二进制文件
build_binary() {
    local os=${1:-"linux"}
    local arch=${2:-"amd64"}
    local output_name="$PROJECT_NAME"
    
    if [ "$os" = "windows" ]; then
        output_name="$output_name.exe"
    fi
    
    local output_path="$OUTPUT_DIR/$output_name"
    if [ "$os" != "linux" ] || [ "$arch" != "amd64" ]; then
        output_path="$OUTPUT_DIR/$PROJECT_NAME-$os-$arch"
        if [ "$os" = "windows" ]; then
            output_path="$output_path.exe"
        fi
    fi
    
    log_info "构建 $os/$arch -> $output_path"
    
    GOOS=$os GOARCH=$arch go build \
        -ldflags "-w -s \
            -X '${VERSION_PKG}.serviceName=${PROJECT_NAME}' \
            -X '${VERSION_PKG}.gitVersion=${GIT_VERSION}' \
            -X '${VERSION_PKG}.gitCommit=${GIT_COMMIT}' \
            -X '${VERSION_PKG}.gitBranch=${GIT_BRANCH}' \
            -X '${VERSION_PKG}.gitTreeState=${GIT_TREE_STATE}' \
            -X '${VERSION_PKG}.buildDate=${BUILD_DATE}'" \
        -o "$output_path" \
        ./cmd/$PROJECT_NAME
    
    if [ -f "$output_path" ]; then
        log_info "构建成功: $output_path"
        
        # 显示文件大小
        if command -v du >/dev/null 2>&1; then
            size=$(du -h "$output_path" | cut -f1)
            log_info "文件大小: $size"
        fi
    else
        log_error "构建失败: $output_path"
        exit 1
    fi
}

# 主函数
main() {
    log_info "开始构建 $PROJECT_NAME..."
    
    check_git
    get_version_info
    
    mkdir -p "$OUTPUT_DIR"
    
    # 根据参数决定构建目标
    case "${1:-single}" in
        "single")
            build_binary
            ;;
        "multi")
            for os in linux darwin windows; do
                for arch in amd64 arm64; do
                    build_binary "$os" "$arch"
                done
            done
            ;;
        "linux")
            build_binary "linux" "amd64"
            build_binary "linux" "arm64"
            ;;
        "darwin")
            build_binary "darwin" "amd64"
            build_binary "darwin" "arm64"
            ;;
        "windows")
            build_binary "windows" "amd64"
            build_binary "windows" "arm64"
            ;;
        *)
            log_error "不支持的构建目标: $1"
            log_info "支持的目标: single, multi, linux, darwin, windows"
            exit 1
            ;;
    esac
    
    log_info "构建完成!"
}

# 检查参数
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "用法: $0 [目标]"
    echo ""
    echo "目标:"
    echo "  single   - 构建当前平台 (默认)"
    echo "  multi    - 构建所有平台"
    echo "  linux    - 构建 Linux 平台"
    echo "  darwin   - 构建 macOS 平台"
    echo "  windows  - 构建 Windows 平台"
    echo ""
    echo "环境变量:"
    echo "  PROJECT_NAME  - 项目名称 (默认: myservice)"
    echo "  OUTPUT_DIR    - 输出目录 (默认: bin)"
    exit 0
fi

main "$@"
```

## Web 服务集成

### 1. HTTP API 版本端点

```go
package main

import (
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/gorilla/mux"
    "github.com/costa92/go-protoc/v2/pkg/version"
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

// VersionResponse API 响应结构
type VersionResponse struct {
    Version   version.Info `json:"version"`
    Timestamp string       `json:"timestamp"`
    Status    string       `json:"status"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
    Status    string       `json:"status"`
    Version   string       `json:"version"`
    Service   string       `json:"service"`
    Timestamp string       `json:"timestamp"`
    Uptime    string       `json:"uptime"`
}

var (
    startTime = time.Now()
    versionInfo = version.Get()
)

// 版本信息处理器
func versionHandler(w http.ResponseWriter, r *http.Request) {
    response := VersionResponse{
        Version:   versionInfo,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        Status:    "ok",
    }
    
    w.Header().Set("Content-Type", "application/json")
    
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Errorw("Failed to encode version response", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    
    logger.Debugw("Version info requested",
        "remote_addr", r.RemoteAddr,
        "user_agent", r.UserAgent(),
    )
}

// 健康检查处理器
func healthHandler(w http.ResponseWriter, r *http.Request) {
    uptime := time.Since(startTime)
    
    response := HealthResponse{
        Status:    "healthy",
        Version:   versionInfo.GitVersion,
        Service:   versionInfo.ServiceName,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        Uptime:    uptime.String(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// 简化版本处理器
func versionTextHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(versionInfo.String()))
}

// 详细版本处理器
func versionDetailHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(versionInfo.Text()))
}

func setupRoutes() *mux.Router {
    r := mux.NewRouter()
    
    // API 路由
    api := r.PathPrefix("/api/v1").Subrouter()
    api.HandleFunc("/version", versionHandler).Methods("GET")
    api.HandleFunc("/health", healthHandler).Methods("GET")
    
    // 简化路由
    r.HandleFunc("/version", versionTextHandler).Methods("GET")
    r.HandleFunc("/version/detail", versionDetailHandler).Methods("GET")
    
    return r
}

func main() {
    // 启动时记录版本信息
    logger.Infow("Starting web service",
        "service", versionInfo.ServiceName,
        "version", versionInfo.GitVersion,
        "branch", versionInfo.GitBranch,
        "commit", versionInfo.GitCommit[:8],
        "build_date", versionInfo.BuildDate,
    )
    
    router := setupRoutes()
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }
    
    logger.Infow("Server listening", "addr", server.Addr)
    if err := server.ListenAndServe(); err != nil {
        logger.Fatalw("Server failed", "error", err)
    }
}
```

### 2. 中间件集成

```go
package middleware

import (
    "net/http"
    "time"
    
    "github.com/costa92/go-protoc/v2/pkg/version"
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

// VersionMiddleware 添加版本信息到响应头和日志上下文
func VersionMiddleware(next http.Handler) http.Handler {
    info := version.Get()
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 添加版本信息到响应头
        w.Header().Set("X-Service-Name", info.ServiceName)
        w.Header().Set("X-Service-Version", info.GitVersion)
        w.Header().Set("X-Build-Date", info.BuildDate)
        
        // 在请求上下文中添加版本信息
        ctx := r.Context()
        
        // 为该请求创建带版本信息的日志记录器
        reqLogger := logger.WithCtx(ctx,
            "service", info.ServiceName,
            "version", info.GitVersion,
            "branch", info.GitBranch,
            "request_id", generateRequestID(),
        )
        
        start := time.Now()
        
        // 记录请求开始
        reqLogger.Infow("Request started",
            "method", r.Method,
            "path", r.URL.Path,
            "remote_addr", r.RemoteAddr,
        )
        
        next.ServeHTTP(w, r)
        
        // 记录请求完成
        reqLogger.Infow("Request completed",
            "method", r.Method,
            "path", r.URL.Path,
            "duration", time.Since(start),
        )
    })
}

func generateRequestID() string {
    // 生成请求ID的简单实现
    return fmt.Sprintf("%d", time.Now().UnixNano())
}
```

## CLI 应用集成

### 1. Cobra CLI 集成

```go
package cmd

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    "github.com/costa92/go-protoc/v2/pkg/version"
)

var (
    // 版本输出格式
    versionFormat string
    
    // 版本命令
    versionCmd = &cobra.Command{
        Use:   "version",
        Short: "显示版本信息",
        Long:  `显示应用程序的详细版本信息，包括Git版本、提交信息、构建时间等。`,
        Run:   runVersionCommand,
    }
)

func init() {
    // 添加标志
    versionCmd.Flags().StringVarP(&versionFormat, "output", "o", "text", 
        "输出格式 (text|json|yaml)")
}

func runVersionCommand(cmd *cobra.Command, args []string) {
    info := version.Get()
    
    switch versionFormat {
    case "json":
        fmt.Println(info.ToJSON())
    case "yaml":
        // 如果实现了 YAML 格式
        if yamlOutput, err := info.ToYAML(); err == nil {
            fmt.Println(yamlOutput)
        } else {
            fmt.Fprintf(os.Stderr, "YAML输出失败: %v\n", err)
            os.Exit(1)
        }
    case "text", "":
        fmt.Println(info.Text())
    default:
        fmt.Fprintf(os.Stderr, "不支持的输出格式: %s\n", versionFormat)
        fmt.Fprintf(os.Stderr, "支持的格式: text, json, yaml\n")
        os.Exit(1)
    }
}

// GetVersionCmd 返回版本命令
func GetVersionCmd() *cobra.Command {
    return versionCmd
}
```

### 2. 主命令集成

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    "github.com/costa92/go-protoc/v2/pkg/version"
    "github.com/costa92/go-protoc/v2/pkg/logger"
    "github.com/yourproject/cmd"
)

var (
    cfgFile string
    verbose bool
)

// 根命令
var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "我的应用程序",
    Long:  `一个基于Go构建的示例应用程序`,
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // 设置日志级别
        if verbose {
            logger.SetLevel(logger.DebugLevel)
        }
        
        // 记录启动信息
        info := version.Get()
        logger.Infow("Application starting",
            "service", info.ServiceName,
            "version", info.GitVersion,
            "branch", info.GitBranch,
            "commit", info.GitCommit[:8],
        )
    },
}

func init() {
    // 全局标志
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", 
        "配置文件路径 (默认搜索 $HOME/.myapp.yaml)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, 
        "启用详细输出")
    
    // 添加版本命令
    rootCmd.AddCommand(cmd.GetVersionCmd())
    
    // 添加其他子命令
    rootCmd.AddCommand(cmd.GetServerCmd())
    rootCmd.AddCommand(cmd.GetMigrateCmd())
}

func main() {
    // 设置版本信息 (可选，用于 cobra 的内置版本处理)
    info := version.Get()
    rootCmd.Version = info.GitVersion
    
    // 自定义版本模板
    rootCmd.SetVersionTemplate(fmt.Sprintf(`%s
服务名: %s
版本: %s
分支: %s  
提交: %s
构建时间: %s
Go版本: %s
平台: %s
`,
        info.GitVersion,
        info.ServiceName,
        info.GitVersion,
        info.GitBranch,
        info.GitCommit[:8],
        info.BuildDate,
        info.GoVersion,
        info.Platform,
    ))
    
    if err := rootCmd.Execute(); err != nil {
        logger.Errorw("Command execution failed", "error", err)
        os.Exit(1)
    }
}
```

## 日志和监控集成

### 1. 结构化日志集成

```go
package main

import (
    "context"
    "time"
    
    "github.com/costa92/go-protoc/v2/pkg/version"
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

// 服务上下文
type ServiceContext struct {
    Version   version.Info
    StartTime time.Time
    Logger    logger.Logger
}

func NewServiceContext() *ServiceContext {
    info := version.Get()
    
    // 创建带版本信息的日志记录器
    serviceLogger := logger.With(
        "service", info.ServiceName,
        "version", info.GitVersion,
        "branch", info.GitBranch,
        "commit", info.GitCommit[:8],
        "build_date", info.BuildDate,
    )
    
    return &ServiceContext{
        Version:   info,
        StartTime: time.Now(),
        Logger:    serviceLogger,
    }
}

func (sc *ServiceContext) LogStartup() {
    sc.Logger.Infow("Service startup",
        "go_version", sc.Version.GoVersion,
        "platform", sc.Version.Platform,
        "git_tree_state", sc.Version.GitTreeState,
    )
    
    // 版本兼容性检查
    if sc.Version.GitTreeState == "dirty" {
        sc.Logger.Warnw("Service built from modified source",
            "warning", "production_deployment_not_recommended",
        )
    }
}

func (sc *ServiceContext) LogShutdown() {
    uptime := time.Since(sc.StartTime)
    
    sc.Logger.Infow("Service shutdown",
        "uptime", uptime.String(),
        "uptime_seconds", uptime.Seconds(),
    )
}

// 业务操作日志示例
func (sc *ServiceContext) ProcessRequest(ctx context.Context, requestID string) {
    // 创建请求特定的日志记录器
    reqLogger := sc.Logger.WithCtx(ctx, "request_id", requestID)
    
    reqLogger.Infow("Processing request")
    
    // 模拟处理
    time.Sleep(100 * time.Millisecond)
    
    reqLogger.Infow("Request processed successfully",
        "duration_ms", 100,
        "status", "success",
    )
}

func main() {
    ctx := context.Background()
    sc := NewServiceContext()
    
    // 记录启动
    sc.LogStartup()
    
    // 模拟请求处理
    for i := 0; i < 3; i++ {
        sc.ProcessRequest(ctx, fmt.Sprintf("req-%d", i+1))
    }
    
    // 记录关闭
    sc.LogShutdown()
}
```

### 2. Prometheus 指标集成

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/costa92/go-protoc/v2/pkg/version"
)

var (
    // 版本信息指标
    versionInfo = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "application_version_info",
            Help: "应用程序版本信息",
        },
        []string{"service", "version", "branch", "commit", "build_date", "go_version", "platform"},
    )
    
    // 启动时间指标
    startTime = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "application_start_time_seconds",
            Help: "应用程序启动时间（Unix时间戳）",
        },
    )
    
    // HTTP请求指标（包含版本标签）
    httpRequests = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "HTTP请求总数",
        },
        []string{"service", "version", "method", "endpoint", "status_code"},
    )
    
    // 应用信息
    info = version.Get()
)

func init() {
    // 注册版本信息指标
    versionInfo.WithLabelValues(
        info.ServiceName,
        info.GitVersion,
        info.GitBranch,
        info.GitCommit[:8],
        info.BuildDate,
        info.GoVersion,
        info.Platform,
    ).Set(1)
    
    // 设置启动时间
    startTime.SetToCurrentTime()
}

// RecordHTTPRequest 记录HTTP请求指标
func RecordHTTPRequest(method, endpoint, statusCode string) {
    httpRequests.WithLabelValues(
        info.ServiceName,
        info.GitVersion,
        method,
        endpoint,
        statusCode,
    ).Inc()
}

// 应用特定指标
type AppMetrics struct {
    ProcessedItems prometheus.Counter
    ErrorsTotal    prometheus.Counter
}

func NewAppMetrics() *AppMetrics {
    return &AppMetrics{
        ProcessedItems: promauto.NewCounter(prometheus.CounterOpts{
            Name:        "app_processed_items_total",
            Help:        "处理的项目总数",
            ConstLabels: prometheus.Labels{
                "service": info.ServiceName,
                "version": info.GitVersion,
            },
        }),
        
        ErrorsTotal: promauto.NewCounter(prometheus.CounterOpts{
            Name:        "app_errors_total", 
            Help:        "错误总数",
            ConstLabels: prometheus.Labels{
                "service": info.ServiceName,
                "version": info.GitVersion,
            },
        }),
    }
}
```

## 动态版本管理

### 1. 环境变量动态版本

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/costa92/go-protoc/v2/pkg/version"
)

func setupDynamicVersion() {
    // 检查环境变量中的动态版本设置
    if dynamicVersion := os.Getenv("DYNAMIC_VERSION"); dynamicVersion != "" {
        log.Printf("检测到动态版本设置: %s", dynamicVersion)
        
        if err := version.SetDynamicVersion(dynamicVersion); err != nil {
            log.Fatalf("设置动态版本失败: %v", err)
        }
        
        log.Printf("动态版本设置成功: %s", dynamicVersion)
    }
    
    // 检查版本覆盖（用于A/B测试等场景）
    if versionOverride := os.Getenv("VERSION_OVERRIDE"); versionOverride != "" {
        log.Printf("检测到版本覆盖: %s", versionOverride)
        
        // 注意：这里应该有额外的验证逻辑
        if err := validateVersionOverride(versionOverride); err != nil {
            log.Fatalf("版本覆盖验证失败: %v", err)
        }
        
        if err := version.SetDynamicVersion(versionOverride); err != nil {
            log.Fatalf("设置版本覆盖失败: %v", err)
        }
        
        log.Printf("版本覆盖设置成功: %s", versionOverride)
    }
}

func validateVersionOverride(override string) error {
    // 实现版本覆盖的验证逻辑
    // 例如：检查是否在允许的版本列表中
    allowedVersions := []string{"v1.0.0", "v1.0.1", "v1.1.0"}
    
    for _, allowed := range allowedVersions {
        if override == allowed {
            return nil
        }
    }
    
    return fmt.Errorf("版本 %s 不在允许列表中", override)
}

func main() {
    // 设置动态版本
    setupDynamicVersion()
    
    // 显示最终版本信息
    info := version.Get()
    fmt.Printf("运行版本: %s\n", info.GitVersion)
    fmt.Printf("服务名: %s\n", info.ServiceName)
}
```

### 2. 配置文件动态版本

```go
package config

import (
    "fmt"
    "gopkg.in/yaml.v3"
    "io/ioutil"
    
    "github.com/costa92/go-protoc/v2/pkg/version"
)

// Config 应用配置
type Config struct {
    Service ServiceConfig `yaml:"service"`
    // ... 其他配置
}

// ServiceConfig 服务配置
type ServiceConfig struct {
    Name            string `yaml:"name"`
    DynamicVersion  string `yaml:"dynamic_version,omitempty"`
    VersionOverride string `yaml:"version_override,omitempty"`
}

// LoadConfig 加载配置文件
func LoadConfig(filename string) (*Config, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("读取配置文件失败: %w", err)
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("解析配置文件失败: %w", err)
    }
    
    // 应用动态版本设置
    if err := applyVersionSettings(&config); err != nil {
        return nil, fmt.Errorf("应用版本设置失败: %w", err)
    }
    
    return &config, nil
}

func applyVersionSettings(config *Config) error {
    // 优先级：version_override > dynamic_version
    versionToSet := ""
    
    if config.Service.VersionOverride != "" {
        versionToSet = config.Service.VersionOverride
    } else if config.Service.DynamicVersion != "" {
        versionToSet = config.Service.DynamicVersion
    }
    
    if versionToSet != "" {
        if err := version.SetDynamicVersion(versionToSet); err != nil {
            return fmt.Errorf("设置动态版本 %s 失败: %w", versionToSet, err)
        }
        
        fmt.Printf("从配置文件应用动态版本: %s\n", versionToSet)
    }
    
    return nil
}
```

## 容器化部署

### 1. Docker 多阶段构建

```dockerfile
# Dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装必要工具
RUN apk add --no-cache git

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建参数
ARG SERVICE_NAME=myservice
ARG VERSION
ARG COMMIT
ARG BRANCH
ARG BUILD_DATE
ARG TREE_STATE=clean

# 版本包路径
ARG VERSION_PKG=github.com/costa92/go-protoc/v2/pkg/version

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s \
        -X '${VERSION_PKG}.serviceName=${SERVICE_NAME}' \
        -X '${VERSION_PKG}.gitVersion=${VERSION}' \
        -X '${VERSION_PKG}.gitCommit=${COMMIT}' \
        -X '${VERSION_PKG}.gitBranch=${BRANCH}' \
        -X '${VERSION_PKG}.gitTreeState=${TREE_STATE}' \
        -X '${VERSION_PKG}.buildDate=${BUILD_DATE}'" \
    -o myservice ./cmd/myservice

# 运行阶段
FROM alpine:latest

# 安装CA证书
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/myservice .

# 重新声明构建参数（运行时标签需要）
ARG VERSION
ARG COMMIT
ARG BRANCH
ARG BUILD_DATE

# 添加标签
LABEL \
    org.opencontainers.image.title="MyService" \
    org.opencontainers.image.description="示例微服务" \
    org.opencontainers.image.version="${VERSION}" \
    org.opencontainers.image.revision="${COMMIT}" \
    org.opencontainers.image.created="${BUILD_DATE}" \
    org.opencontainers.image.source="https://github.com/yourorg/yourrepo" \
    service.version="${VERSION}" \
    service.commit="${COMMIT}" \
    service.branch="${BRANCH}"

# 创建非root用户
RUN adduser -D -s /bin/sh appuser
USER appuser

EXPOSE 8080

CMD ["./myservice"]
```

### 2. Docker Compose 配置

```yaml
# docker-compose.yml
version: '3.8'

services:
  myservice:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        SERVICE_NAME: myservice
        VERSION: ${VERSION:-unknown}
        COMMIT: ${COMMIT:-unknown}
        BRANCH: ${BRANCH:-unknown}
        BUILD_DATE: ${BUILD_DATE:-1970-01-01T00:00:00Z}
    ports:
      - "8080:8080"
    environment:
      - SERVICE_ENV=docker
      - DYNAMIC_VERSION=${DYNAMIC_VERSION:-}
    labels:
      - "service.name=myservice"
      - "service.version=${VERSION:-unknown}"
      - "service.commit=${COMMIT:-unknown}"
    healthcheck:
      test: ["CMD", "./myservice", "--version"]
      interval: 30s
      timeout: 5s
      retries: 3
    restart: unless-stopped
```

### 3. 构建脚本

```bash
#!/bin/bash
# docker-build.sh

set -e

# 获取版本信息
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse HEAD)
BRANCH=$(git branch --show-current)
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
TREE_STATE=$(if [ -n "$(git status --porcelain)" ]; then echo "dirty"; else echo "clean"; fi)

echo "构建Docker镜像..."
echo "版本: $VERSION"
echo "提交: ${COMMIT:0:8}"
echo "分支: $BRANCH"
echo "构建时间: $BUILD_DATE"

# 构建镜像
docker build \
    --build-arg SERVICE_NAME=myservice \
    --build-arg VERSION="$VERSION" \
    --build-arg COMMIT="$COMMIT" \
    --build-arg BRANCH="$BRANCH" \
    --build-arg BUILD_DATE="$BUILD_DATE" \
    --build-arg TREE_STATE="$TREE_STATE" \
    -t myservice:"$VERSION" \
    -t myservice:latest \
    .

echo "构建完成!"
echo "镜像标签: myservice:$VERSION, myservice:latest"

# 可选：推送到镜像仓库
if [ "$1" = "--push" ]; then
    echo "推送镜像到仓库..."
    docker push myservice:"$VERSION"
    docker push myservice:latest
fi
```

## CI/CD 集成

### 1. GitHub Actions

```yaml
# .github/workflows/build.yml
name: Build and Release

on:
  push:
    branches: [ main, develop ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

env:
  SERVICE_NAME: myservice
  VERSION_PKG: github.com/costa92/go-protoc/v2/pkg/version

jobs:
  build:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
      with:
        fetch-depth: 0  # 获取完整历史用于版本标记
    
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Get version info
      id: version
      run: |
        VERSION=$(git describe --tags --always --dirty)
        COMMIT=$(git rev-parse HEAD)
        BRANCH=${GITHUB_REF#refs/heads/}
        BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
        TREE_STATE=$(if [ -n "$(git status --porcelain)" ]; then echo "dirty"; else echo "clean"; fi)
        
        echo "VERSION=$VERSION" >> $GITHUB_OUTPUT
        echo "COMMIT=$COMMIT" >> $GITHUB_OUTPUT  
        echo "BRANCH=$BRANCH" >> $GITHUB_OUTPUT
        echo "BUILD_DATE=$BUILD_DATE" >> $GITHUB_OUTPUT
        echo "TREE_STATE=$TREE_STATE" >> $GITHUB_OUTPUT
        
        echo "Version: $VERSION"
        echo "Commit: ${COMMIT:0:8}"
        echo "Branch: $BRANCH"
    
    - name: Build binary
      env:
        VERSION: ${{ steps.version.outputs.VERSION }}
        COMMIT: ${{ steps.version.outputs.COMMIT }}
        BRANCH: ${{ steps.version.outputs.BRANCH }}
        BUILD_DATE: ${{ steps.version.outputs.BUILD_DATE }}
        TREE_STATE: ${{ steps.version.outputs.TREE_STATE }}
      run: |
        go build -ldflags "\
          -w -s \
          -X '${VERSION_PKG}.serviceName=${SERVICE_NAME}' \
          -X '${VERSION_PKG}.gitVersion=${VERSION}' \
          -X '${VERSION_PKG}.gitCommit=${COMMIT}' \
          -X '${VERSION_PKG}.gitBranch=${BRANCH}' \
          -X '${VERSION_PKG}.gitTreeState=${TREE_STATE}' \
          -X '${VERSION_PKG}.buildDate=${BUILD_DATE}' \
        " -o bin/${SERVICE_NAME} ./cmd/${SERVICE_NAME}
    
    - name: Test version info
      run: |
        ./bin/${SERVICE_NAME} --version
        echo "---"
        ./bin/${SERVICE_NAME} version --output json
    
    - name: Run tests
      run: go test -v ./...
    
    - name: Build multi-platform
      if: startsWith(github.ref, 'refs/tags/v')
      env:
        VERSION: ${{ steps.version.outputs.VERSION }}
        COMMIT: ${{ steps.version.outputs.COMMIT }}
        BRANCH: ${{ steps.version.outputs.BRANCH }}
        BUILD_DATE: ${{ steps.version.outputs.BUILD_DATE }}
        TREE_STATE: ${{ steps.version.outputs.TREE_STATE }}
      run: |
        mkdir -p dist
        
        for os in linux darwin windows; do
          for arch in amd64 arm64; do
            ext=""
            if [ "$os" = "windows" ]; then ext=".exe"; fi
            
            echo "Building $os/$arch..."
            
            GOOS=$os GOARCH=$arch go build -ldflags "\
              -w -s \
              -X '${VERSION_PKG}.serviceName=${SERVICE_NAME}' \
              -X '${VERSION_PKG}.gitVersion=${VERSION}' \
              -X '${VERSION_PKG}.gitCommit=${COMMIT}' \
              -X '${VERSION_PKG}.gitBranch=${BRANCH}' \
              -X '${VERSION_PKG}.gitTreeState=${TREE_STATE}' \
              -X '${VERSION_PKG}.buildDate=${BUILD_DATE}' \
            " -o dist/${SERVICE_NAME}-${os}-${arch}${ext} ./cmd/${SERVICE_NAME}
          done
        done
    
    - name: Create release
      if: startsWith(github.ref, 'refs/tags/v')
      uses: softprops/action-gh-release@v1
      with:
        files: dist/*
        generate_release_notes: true
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  docker:
    runs-on: ubuntu-latest
    needs: build
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
      with:
        fetch-depth: 0
    
    - name: Setup Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Get version info
      id: version
      run: |
        VERSION=$(git describe --tags --always --dirty)
        COMMIT=$(git rev-parse HEAD)
        BRANCH=${GITHUB_REF#refs/heads/}
        BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
        
        echo "VERSION=$VERSION" >> $GITHUB_OUTPUT
        echo "COMMIT=$COMMIT" >> $GITHUB_OUTPUT
        echo "BRANCH=$BRANCH" >> $GITHUB_OUTPUT
        echo "BUILD_DATE=$BUILD_DATE" >> $GITHUB_OUTPUT
    
    - name: Login to registry
      if: github.event_name != 'pull_request'
      uses: docker/login-action@v2
      with:
        registry: ghcr.io
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Build and push
      uses: docker/build-push-action@v4
      with:
        context: .
        push: ${{ github.event_name != 'pull_request' }}
        tags: |
          ghcr.io/${{ github.repository }}:${{ steps.version.outputs.VERSION }}
          ghcr.io/${{ github.repository }}:latest
        build-args: |
          SERVICE_NAME=${{ env.SERVICE_NAME }}
          VERSION=${{ steps.version.outputs.VERSION }}
          COMMIT=${{ steps.version.outputs.COMMIT }}
          BRANCH=${{ steps.version.outputs.BRANCH }}
          BUILD_DATE=${{ steps.version.outputs.BUILD_DATE }}
```

这些示例展示了 `pkg/version` 包在各种实际场景中的完整应用，从简单的版本查询到复杂的 CI/CD 集成，提供了全面的最佳实践指导。