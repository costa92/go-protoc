# Version 包设计文档

## 设计目标

### 核心问题

在微服务架构中，版本信息管理面临以下挑战：

1. **版本信息分散**：版本号、构建时间、Git 信息等分散在不同地方
2. **构建时注入困难**：需要在构建时动态注入实际的版本信息
3. **多格式输出需求**：不同场景需要不同的版本信息展示格式
4. **运行时查询需求**：服务需要能够快速查询和报告自身版本信息
5. **动态版本管理**：特殊部署场景需要运行时修改版本信息

### 设计目标

1. **统一版本信息管理**：提供单一的版本信息来源
2. **构建时自动注入**：通过构建系统自动注入版本信息
3. **多格式支持**：支持简化、JSON、表格等多种输出格式
4. **高性能访问**：版本信息查询应该是高效的
5. **并发安全**：支持并发环境下的安全访问
6. **易于集成**：与现有构建系统和应用程序无缝集成

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Version Package                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ Static Vars │ │Dynamic Mgmt │ │  CLI Flags  │           │
│  │             │ │             │ │             │           │
│  │ gitVersion  │ │atomic.Value │ │ pflag.Flag  │           │
│  │ buildDate   │ │ validation  │ │ version     │           │
│  │ gitCommit   │ │ semver      │ │ handling    │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
│         │               │               │                   │
│         └───────────────┼───────────────┘                   │
│                         │                                   │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Core Info Aggregator                       │ │
│  │                                                         │ │
│  │  func Get() Info {                                      │ │
│  │    // 聚合静态变量 + 动态版本 + 运行时信息              │ │
│  │    return Info{ ... }                                   │ │
│  │  }                                                      │ │
│  └─────────────────────────────────────────────────────────┘ │
│                         │                                   │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Output Formatters                          │ │
│  │                                                         │ │
│  │  String() → "v1.0.0"                                   │ │
│  │  ToJSON() → {"gitVersion":"v1.0.0",...}               │ │  
│  │  Text()   → Table format                               │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   CLI App   │    │  Web API    │    │  Logging    │
│             │    │             │    │             │
│ --version   │    │ /version    │    │ structured  │
│ --help      │    │ /health     │    │ logging     │
└─────────────┘    └─────────────┘    └─────────────┘
```

### 数据流设计

```
构建时 (Build Time)
┌────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Git Repo     │    │  Build System    │    │  Binary File    │
│                │───▶│                  │───▶│                 │
│ git describe   │    │ -ldflags -X      │    │ Embedded        │
│ git rev-parse  │    │ variable=value   │    │ Variables       │
│ git branch     │    │                  │    │                 │
└────────────────┘    └──────────────────┘    └─────────────────┘

运行时 (Runtime)  
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Binary File    │    │  Version Package │    │   Application   │
│                 │───▶│                  │───▶│                 │
│ Static Variables│    │ Get() Function   │    │ Log/API/CLI     │
│ Runtime Info    │    │ Info Struct      │    │ Output          │
│ Dynamic Version │    │ Formatters       │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 核心组件设计

### 1. 静态变量系统

**设计思路**：使用包级变量作为版本信息的载体，通过 ldflags 在构建时注入实际值。

```go
var (
    // 默认值设计有特殊考虑
    gitVersion   = "v0.0.0-master+$Format:%h$"  // Git 格式占位符
    buildDate    = "1970-01-01T00:00:00Z"       // Unix 纪元，表示未设置
    gitCommit    = "$Format:%H$"                // Git 完整格式占位符
    gitTreeState = ""                           // 空字符串，构建时填充
    gitBranch    = "unknown"                    // 明确的未知状态
    serviceName  = "apiserver"                  // 合理的默认服务名
)
```

**设计考虑**：
- **占位符设计**：使用 Git 的 `$Format:` 占位符，便于识别未注入的变量
- **默认值合理性**：所有默认值都有明确的语义，便于调试
- **构建时覆盖**：ldflags 可以完全覆盖这些默认值

### 2. Info 结构体设计

**设计思路**：将所有版本相关信息统一到一个结构体中，支持 JSON 序列化。

```go
type Info struct {
    GitVersion   string `json:"gitVersion"`   // 主版本信息
    GitCommit    string `json:"gitCommit"`    // 具体提交
    GitTreeState string `json:"gitTreeState"` // 构建时仓库状态
    GitBranch    string `json:"gitBranch"`    // 分支信息
    BuildDate    string `json:"buildDate"`    // ISO8601 格式
    ServiceName  string `json:"serviceName"`  // 服务标识
    GoVersion    string `json:"goVersion"`    // 运行时获取
    Compiler     string `json:"compiler"`     // 运行时获取
    Platform     string `json:"platform"`     // 运行时获取
}
```

**设计考虑**：
- **JSON 兼容**：所有字段都支持 JSON 序列化，便于 API 返回
- **信息完整性**：包含构建时信息和运行时信息
- **标准格式**：使用标准的时间格式和平台标识

### 3. 动态版本管理

**设计思路**：使用 `atomic.Value` 实现线程安全的动态版本设置，支持运行时版本覆盖。

```go
var dynamicGitVersion atomic.Value

func SetDynamicVersion(version string) error {
    // 1. 严格验证
    if err := ValidateDynamicVersion(version); err != nil {
        return err
    }
    
    // 2. 原子性存储
    dynamicGitVersion.Store(version)
    return nil
}

func getEffectiveVersion() string {
    // 优先返回动态版本
    if val := dynamicGitVersion.Load(); val != nil {
        return val.(string)
    }
    return gitVersion  // 回退到静态版本
}
```

**设计考虑**：
- **并发安全**：使用 `atomic.Value` 确保并发访问安全
- **验证机制**：严格的语义版本验证，防止无效版本
- **优雅降级**：动态版本失败时回退到静态版本

### 4. 语义版本验证

**设计思路**：确保动态设置的版本符合语义版本规范，并与默认版本兼容。

```go
func validateDynamicVersion(dynamicVersion, defaultVersion string) error {
    // 1. 基础验证
    if len(dynamicVersion) == 0 {
        return fmt.Errorf("version must not be empty")
    }
    
    // 2. 语义版本解析
    vRuntime, err := utilversion.ParseSemantic(dynamicVersion)
    if err != nil {
        return err
    }
    
    // 3. 兼容性检查
    vDefault, err := parseDefaultVersion(defaultVersion)
    if err != nil {
        return err
    }
    
    // 4. 主版本兼容性
    if !isCompatible(vRuntime, vDefault) {
        return fmt.Errorf("version compatibility check failed")
    }
    
    return nil
}
```

**设计考虑**：
- **格式验证**：严格的语义版本格式检查
- **兼容性保证**：确保动态版本与默认版本兼容
- **特殊处理**：处理开发版本等特殊格式

### 5. 多格式输出系统

**设计思路**：提供多种输出格式，满足不同使用场景的需求。

```go
// 简化输出 - 快速查看
func (info Info) String() string {
    return info.GitVersion
}

// JSON 输出 - API 和结构化日志
func (info Info) ToJSON() string {
    data, _ := json.Marshal(info)
    return string(data)
}

// 表格输出 - 人类友好的详细信息
func (info Info) Text() string {
    table := uitable.New()
    // 配置表格格式
    table.RightAlign(0)
    table.MaxColWidth = 80
    table.Separator = " "
    
    // 添加所有字段
    table.AddRow("serviceName:", info.ServiceName)
    table.AddRow("gitVersion:", info.GitVersion)
    // ...
    
    return table.String()
}
```

**设计考虑**：
- **用途分离**：不同格式服务于不同的使用场景
- **格式一致性**：所有格式都基于同一个 Info 结构体
- **性能优化**：JSON 序列化忽略错误，因为结构体总是有效的

### 6. 命令行集成

**设计思路**：提供标准的命令行版本查询支持，集成到应用程序的启动流程中。

```go
type versionValue int

const (
    VersionNotSet versionValue = 0
    VersionEnabled versionValue = 1
    VersionRaw versionValue = 2
)

func PrintAndExitIfRequested() {
    switch *versionFlag {
    case VersionRaw:
        fmt.Printf("%s\n", Get().Text())    // 详细信息
        os.Exit(0)
    case VersionEnabled:
        fmt.Printf("%s\n", Get().String())  // 简化信息
        os.Exit(0)
    }
}
```

**设计考虑**：
- **标准行为**：符合 Unix 工具的 `--version` 标志习惯
- **多种模式**：支持简化和详细两种输出模式
- **早期退出**：版本查询后立即退出，避免应用启动

## 实现细节

### 1. 构建时注入机制

**ldflags 变量映射**：
```bash
# 变量路径规则
github.com/costa92/go-protoc/v2/pkg/version.variableName

# 实际注入命令
go build -ldflags "
  -X 'github.com/costa92/go-protoc/v2/pkg/version.serviceName=myservice'
  -X 'github.com/costa92/go-protoc/v2/pkg/version.gitVersion=$(git describe --tags)'
  -X 'github.com/costa92/go-protoc/v2/pkg/version.gitCommit=$(git rev-parse HEAD)'
  -X 'github.com/costa92/go-protoc/v2/pkg/version.gitBranch=$(git branch --show-current)'
  -X 'github.com/costa92/go-protoc/v2/pkg/version.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'
"
```

**Git 信息获取脚本**：
```bash
#!/bin/bash
# 获取版本信息的标准脚本

# Git 版本 (优先使用 tag，回退到 commit)
GIT_VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "unknown")

# Git 提交 SHA
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")

# Git 分支 (处理 detached HEAD 状态)
GIT_BRANCH=$(git branch --show-current 2>/dev/null || \
           git describe --contains --all HEAD 2>/dev/null || \
           echo "unknown")

# Git 树状态 (检查是否有未提交的修改)
if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
    GIT_TREE_STATE="dirty"
else
    GIT_TREE_STATE="clean"
fi

# ISO8601 格式的构建时间
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
```

### 2. 性能优化考虑

**内存分配优化**：
```go
// 避免重复的字符串拼接
func (info Info) Platform() string {
    // 使用 sync.Once 缓存结果
    once.Do(func() {
        platformCache = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
    })
    return platformCache
}

// 预分配字符串构建器
func (info Info) ToJSON() string {
    // 预估 JSON 大小，减少内存重分配
    var buf bytes.Buffer
    buf.Grow(256) // 预分配缓冲区
    
    enc := json.NewEncoder(&buf)
    enc.Encode(info)
    
    return strings.TrimSpace(buf.String())
}
```

**访问路径优化**：
```go
// 热路径优化
func Get() Info {
    // 避免重复的运行时调用
    goVersion := runtime.Version()
    compiler := runtime.Compiler  
    platform := runtime.GOOS + "/" + runtime.GOARCH
    
    return Info{
        ServiceName:  serviceName,
        GitVersion:   getEffectiveVersion(), // 处理动态版本
        GitCommit:    gitCommit,
        GitBranch:    gitBranch,
        GitTreeState: gitTreeState,
        BuildDate:    buildDate,
        GoVersion:    goVersion,
        Compiler:     compiler,
        Platform:     platform,
    }
}
```

### 3. 错误处理策略

**优雅降级**：
```go
func getEffectiveVersion() string {
    // 动态版本优先
    if val := dynamicGitVersion.Load(); val != nil {
        if version, ok := val.(string); ok && version != "" {
            return version
        }
    }
    
    // 回退到静态版本
    if gitVersion != "" {
        return gitVersion
    }
    
    // 最终回退
    return "unknown"
}
```

**验证错误处理**：
```go
func SetDynamicVersion(version string) error {
    if err := ValidateDynamicVersion(version); err != nil {
        return fmt.Errorf("invalid dynamic version %q: %w", version, err)
    }
    
    dynamicGitVersion.Store(version)
    return nil
}
```

## 扩展点设计

### 1. 输出格式扩展

```go
// 接口定义
type Formatter interface {
    Format(info Info) (string, error)
}

// 实现示例
type YAMLFormatter struct{}

func (f YAMLFormatter) Format(info Info) (string, error) {
    return yaml.Marshal(info)
}

// 注册机制
var formatters = map[string]Formatter{
    "json":  JSONFormatter{},
    "yaml":  YAMLFormatter{},
    "table": TableFormatter{},
}
```

### 2. 版本比较扩展

```go
// 版本比较接口
type Comparable interface {
    Compare(other string) (int, error)
    IsPrerelease() bool
    IsCompatible(other string) bool
}

// 实现
func (info Info) Compare(other string) (int, error) {
    return semver.Compare(info.GitVersion, other)
}
```

### 3. 环境感知扩展

```go
// 环境信息收集
type EnvironmentProvider interface {
    GetEnvironmentInfo() map[string]string
}

// K8s 环境提供者
type KubernetesProvider struct{}

func (k KubernetesProvider) GetEnvironmentInfo() map[string]string {
    return map[string]string{
        "namespace": os.Getenv("K8S_NAMESPACE"),
        "pod":       os.Getenv("HOSTNAME"),
        "node":      os.Getenv("K8S_NODE_NAME"),
    }
}
```

## 测试策略

### 1. 单元测试覆盖

```go
func TestVersionInjection(t *testing.T) {
    tests := []struct {
        name     string
        ldflags  map[string]string
        expected Info
    }{
        {
            name: "complete_version_info",
            ldflags: map[string]string{
                "gitVersion": "v1.0.0",
                "gitCommit":  "abc123",
                "gitBranch":  "main",
            },
            expected: Info{
                GitVersion: "v1.0.0",
                GitCommit:  "abc123", 
                GitBranch:  "main",
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 模拟 ldflags 注入
            simulateLdflags(tt.ldflags)
            
            info := Get()
            assert.Equal(t, tt.expected.GitVersion, info.GitVersion)
            assert.Equal(t, tt.expected.GitCommit, info.GitCommit)
            assert.Equal(t, tt.expected.GitBranch, info.GitBranch)
        })
    }
}
```

### 2. 集成测试

```go
func TestBuildIntegration(t *testing.T) {
    // 真实构建测试
    cmd := exec.Command("go", "build", 
        "-ldflags", "-X pkg/version.gitVersion=test-v1.0.0",
        "./cmd/test-app")
    
    err := cmd.Run()
    assert.NoError(t, err)
    
    // 执行构建结果
    out, err := exec.Command("./test-app", "--version").Output()
    assert.NoError(t, err)
    assert.Contains(t, string(out), "test-v1.0.0")
}
```

### 3. 并发安全测试

```go
func TestConcurrentAccess(t *testing.T) {
    const numGoroutines = 100
    
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // 并发读取
            info := Get()
            if info.GitVersion == "" {
                errors <- fmt.Errorf("empty version in goroutine %d", id)
                return
            }
            
            // 并发设置动态版本
            version := fmt.Sprintf("v1.0.%d", id)
            if err := SetDynamicVersion(version); err != nil {
                errors <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        t.Error(err)
    }
}
```

## 性能基准

### 1. 基础性能测试

```go
func BenchmarkGet(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = Get()
    }
}

func BenchmarkToJSON(b *testing.B) {
    info := Get()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _ = info.ToJSON()
    }
}

func BenchmarkDynamicVersion(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = SetDynamicVersion(fmt.Sprintf("v1.0.%d", i))
    }
}
```

### 2. 内存分配测试

```go
func BenchmarkMemoryAllocation(b *testing.B) {
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        info := Get()
        _ = info.ToJSON()
        _ = info.Text()
    }
}
```

这个设计文档展现了 version 包的完整技术架构和实现思路，为开发者提供了深入理解和扩展的基础。通过这种模块化和扩展性的设计，version 包能够满足各种复杂的版本管理需求。