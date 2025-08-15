# Options Package

配置选项包，提供统一的配置管理和命令行参数处理，支持多种数据库、中间件和服务组件的配置。

## 📋 目录

- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [支持的组件](#支持的组件)
- [配置选项](#配置选项)
- [API参考](#api参考)
- [最佳实践](#最佳实践)
- [配置示例](#配置示例)

## 🎯 功能特性

### 核心特性
- 🔧 **统一接口**: 标准化的配置选项接口
- 🚩 **命令行支持**: 自动生成命令行参数
- ✅ **配置验证**: 内置配置验证和默认值
- 🏗️ **工厂模式**: 支持各种组件的工厂创建
- 📊 **指标集成**: 内置监控和追踪配置
- 🔌 **插件化**: 模块化的配置组件

### 支持的组件
- **数据库**: MySQL、PostgreSQL、MongoDB、Redis
- **服务**: HTTP、gRPC、TLS安全服务
- **中间件**: JWT认证、健康检查、指标监控
- **基础设施**: Consul、etcd、Kafka、Jaeger

## 🚀 快速开始

### 基础使用

```go
package main

import (
    "flag"
    "github.com/costa92/go-protoc/v2/pkg/options"
    "github.com/spf13/pflag"
)

func main() {
    // 1. 创建配置选项
    mysqlOpts := options.NewMySQLOptions()
    redisOpts := options.NewRedisOptions()
    httpOpts := options.NewHTTPOptions()

    // 2. 创建命令行标志
    fs := pflag.NewFlagSet("myapp", pflag.ExitOnError)
    
    // 3. 添加配置到命令行
    mysqlOpts.AddFlags(fs, "storage")
    redisOpts.AddFlags(fs, "cache")
    httpOpts.AddFlags(fs, "server")

    // 4. 解析命令行参数
    fs.Parse(os.Args[1:])

    // 5. 验证配置
    if errs := mysqlOpts.Validate(); len(errs) > 0 {
        for _, err := range errs {
            log.Error("MySQL配置错误", err)
        }
        os.Exit(1)
    }

    // 6. 创建组件
    db, err := mysqlOpts.NewDB()
    if err != nil {
        log.Fatal("创建数据库连接失败", err)
    }

    redisClient, err := redisOpts.NewClient()
    if err != nil {
        log.Fatal("创建Redis客户端失败", err)
    }

    httpServer, err := httpOpts.NewServer()
    if err != nil {
        log.Fatal("创建HTTP服务器失败", err)
    }

    // 7. 启动服务
    log.Info("服务启动成功")
}
```

### 配置文件支持

```go
package main

import (
    "github.com/costa92/go-protoc/v2/pkg/options"
    "github.com/spf13/viper"
)

func main() {
    // 1. 设置Viper配置
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")
    
    if err := viper.ReadInConfig(); err != nil {
        log.Fatal("读取配置文件失败", err)
    }

    // 2. 创建配置选项
    mysqlOpts := options.NewMySQLOptions()
    
    // 3. 从配置文件绑定
    viper.Unmarshal(mysqlOpts)
    
    // 4. 验证和使用
    if errs := mysqlOpts.Validate(); len(errs) > 0 {
        log.Fatal("配置验证失败", errs)
    }

    db, err := mysqlOpts.NewDB()
    if err != nil {
        log.Fatal("创建数据库失败", err)
    }
}
```

### 带监控的组件创建

```go
func createMonitoredComponents() {
    // MySQL配置（启用监控）
    mysqlOpts := options.NewMySQLOptions()
    mysqlOpts.EnableMetrics = true
    mysqlOpts.MetricsName = "main_db"
    mysqlOpts.EnableTrace = true

    // 创建带监控的数据库
    monitoredDB, err := mysqlOpts.NewMonitoredDB()
    if err != nil {
        log.Fatal("创建监控数据库失败", err)
    }

    // Redis配置（启用监控）
    redisOpts := options.NewRedisOptions()
    redisOpts.EnableMetrics = true
    redisOpts.MetricsName = "main_cache"

    // 创建带监控的Redis客户端
    monitoredRedis, err := redisOpts.NewMonitoredClient()
    if err != nil {
        log.Fatal("创建监控Redis失败", err)
    }

    log.Info("监控组件创建成功")
}
```

## 🔧 支持的组件

### 数据库选项

#### MySQL选项
```go
type MySQLOptions struct {
    Addr                  string        // 数据库地址
    Username              string        // 用户名
    Password              string        // 密码
    Database              string        // 数据库名
    MaxIdleConnections    int           // 最大空闲连接数
    MaxOpenConnections    int           // 最大打开连接数
    MaxConnectionLifeTime time.Duration // 连接最大生存时间
    LogLevel              int           // 日志级别
    EnableTrace           bool          // 启用追踪
    EnableMetrics         bool          // 启用指标
    MetricsName           string        // 指标名称
}

// 创建方法
func NewMySQLOptions() *MySQLOptions
func (o *MySQLOptions) NewDB() (*gorm.DB, error)
func (o *MySQLOptions) NewMonitoredDB() (*db.MonitoredDB, error)
func (o *MySQLOptions) DSN() string
```

#### Redis选项
```go
type RedisOptions struct {
    Addr          string        // Redis地址
    Password      string        // 密码
    DB            int           // 数据库索引
    PoolSize      int           // 连接池大小
    MinIdleConns  int           // 最小空闲连接数
    MaxRetries    int           // 最大重试次数
    DialTimeout   time.Duration // 连接超时
    ReadTimeout   time.Duration // 读取超时
    WriteTimeout  time.Duration // 写入超时
    EnableMetrics bool          // 启用指标
    MetricsName   string        // 指标名称
}

// 创建方法
func NewRedisOptions() *RedisOptions
func (o *RedisOptions) NewClient() (*redis.Client, error)
func (o *RedisOptions) NewMonitoredClient() (*db.MonitoredRedis, error)
```

#### MongoDB选项
```go
type MongoOptions struct {
    URI               string        // MongoDB URI
    Database          string        // 数据库名
    Username          string        // 用户名
    Password          string        // 密码
    MaxPoolSize       uint64        // 最大连接池大小
    MinPoolSize       uint64        // 最小连接池大小
    MaxConnIdleTime   time.Duration // 连接最大空闲时间
    ConnectTimeout    time.Duration // 连接超时
    ServerSelectionTimeout time.Duration // 服务器选择超时
    EnableMetrics     bool          // 启用指标
    MetricsName       string        // 指标名称
}
```

### 服务选项

#### HTTP选项
```go
type HTTPOptions struct {
    BindAddress         string        // 绑定地址
    BindPort            int           // 绑定端口
    ReadTimeout         time.Duration // 读取超时
    WriteTimeout        time.Duration // 写入超时
    IdleTimeout         time.Duration // 空闲超时
    MaxHeaderBytes      int           // 最大请求头大小
    EnableMetrics       bool          // 启用指标
    MetricsPath         string        // 指标路径
    EnableCORS          bool          // 启用CORS
    CORSAllowedOrigins  []string      // CORS允许的源
}

// 创建方法
func NewHTTPOptions() *HTTPOptions
func (o *HTTPOptions) NewServer() (*http.Server, error)
```

#### gRPC选项
```go
type GRPCOptions struct {
    BindAddress              string        // 绑定地址
    BindPort                 int           // 绑定端口
    MaxReceiveMessageSize    int           // 最大接收消息大小
    MaxSendMessageSize       int           // 最大发送消息大小
    MaxConcurrentStreams     uint32        // 最大并发流数
    ConnectionTimeout        time.Duration // 连接超时
    EnableReflection         bool          // 启用反射
    EnableMetrics            bool          // 启用指标
    EnableTrace              bool          // 启用追踪
}

// 创建方法
func NewGRPCOptions() *GRPCOptions
func (o *GRPCOptions) NewServer() (*grpc.Server, error)
```

### 安全选项

#### TLS选项
```go
type TLSOptions struct {
    CertFile      string   // 证书文件
    KeyFile       string   // 密钥文件
    CAFile        string   // CA文件
    ServerName    string   // 服务器名称
    InsecureSkipVerify bool // 跳过证书验证
    MinVersion    string   // 最小TLS版本
    MaxVersion    string   // 最大TLS版本
    CipherSuites  []string // 加密套件
}

// 创建方法
func NewTLSOptions() *TLSOptions
func (o *TLSOptions) NewTLSConfig() (*tls.Config, error)
```

#### JWT选项
```go
type JWTOptions struct {
    SecretKey        string        // 密钥
    Algorithm        string        // 算法
    TokenExpiration  time.Duration // Token过期时间
    RefreshExpiration time.Duration // 刷新Token过期时间
    Issuer           string        // 签发者
    Audience         string        // 受众
}

// 创建方法
func NewJWTOptions() *JWTOptions
func (o *JWTOptions) NewAuthenticator() (*jwt.Authenticator, error)
```

### 中间件选项

#### 健康检查选项
```go
type HealthOptions struct {
    CheckInterval    time.Duration     // 检查间隔
    Timeout          time.Duration     // 超时时间
    FailureThreshold int               // 失败阈值
    SuccessThreshold int               // 成功阈值
    EnableMetrics    bool              // 启用指标
    Checks           []HealthCheckConfig // 检查配置
}

type HealthCheckConfig struct {
    Name     string        // 检查名称
    Type     string        // 检查类型
    Endpoint string        // 检查端点
    Timeout  time.Duration // 超时时间
}
```

#### 指标选项
```go
type MetricsOptions struct {
    Enabled      bool          // 启用指标
    Path         string        // 指标路径
    Port         int           // 指标端口
    Namespace    string        // 命名空间
    Subsystem    string        // 子系统
    EnableCPU    bool          // 启用CPU指标
    EnableMemory bool          // 启用内存指标
    EnableGC     bool          // 启用GC指标
}
```

### 基础设施选项

#### Consul选项
```go
type ConsulOptions struct {
    Address    string        // Consul地址
    Scheme     string        // 协议方案
    Datacenter string        // 数据中心
    Token      string        // 访问令牌
    Timeout    time.Duration // 超时时间
}
```

#### etcd选项
```go
type EtcdOptions struct {
    Endpoints   []string      // etcd端点
    Username    string        // 用户名
    Password    string        // 密码
    DialTimeout time.Duration // 连接超时
    AutoSyncInterval time.Duration // 自动同步间隔
    EnableTLS   bool          // 启用TLS
}
```

#### Kafka选项
```go
type KafkaOptions struct {
    Brokers         []string      // Broker地址
    ClientID        string        // 客户端ID
    Version         string        // Kafka版本
    EnableSASL      bool          // 启用SASL
    SASLMechanism   string        // SASL机制
    SASLUsername    string        // SASL用户名
    SASLPassword    string        // SASL密码
    EnableTLS       bool          // 启用TLS
}
```

#### Jaeger选项
```go
type JaegerOptions struct {
    ServiceName     string  // 服务名称
    RPCMetrics      bool    // RPC指标
    Tags            string  // 标签
    SamplerType     string  // 采样器类型
    SamplerParam    float64 // 采样器参数
    LocalAgentHostPort string // 本地代理地址
    CollectorEndpoint  string // 收集器端点
}
```

## 📚 API参考

### 核心接口

```go
// IOptions 基础配置接口
type IOptions interface {
    // Validate 验证配置
    Validate() []error
    
    // AddFlags 添加命令行参数
    AddFlags(fs *pflag.FlagSet, prefixes ...string)
}
```

### 配置选项创建

```go
// 数据库
func NewMySQLOptions() *MySQLOptions
func NewPostgreSQLOptions() *PostgreSQLOptions
func NewMongoOptions() *MongoOptions
func NewRedisOptions() *RedisOptions

// 服务
func NewHTTPOptions() *HTTPOptions
func NewGRPCOptions() *GRPCOptions
func NewTLSOptions() *TLSOptions

// 中间件
func NewJWTOptions() *JWTOptions
func NewHealthOptions() *HealthOptions
func NewMetricsOptions() *MetricsOptions

// 基础设施
func NewConsulOptions() *ConsulOptions
func NewEtcdOptions() *EtcdOptions
func NewKafkaOptions() *KafkaOptions
func NewJaegerOptions() *JaegerOptions
```

### 工厂方法

```go
// 数据库连接
func (o *MySQLOptions) NewDB() (*gorm.DB, error)
func (o *MySQLOptions) NewMonitoredDB() (*db.MonitoredDB, error)
func (o *RedisOptions) NewClient() (*redis.Client, error)
func (o *RedisOptions) NewMonitoredClient() (*db.MonitoredRedis, error)

// 服务器
func (o *HTTPOptions) NewServer() (*http.Server, error)
func (o *GRPCOptions) NewServer() (*grpc.Server, error)

// 安全
func (o *TLSOptions) NewTLSConfig() (*tls.Config, error)
func (o *JWTOptions) NewAuthenticator() (*jwt.Authenticator, error)

// 客户端
func (o *ConsulOptions) NewClient() (*consul.Client, error)
func (o *EtcdOptions) NewClient() (*clientv3.Client, error)
func (o *KafkaOptions) NewProducer() (sarama.SyncProducer, error)
func (o *KafkaOptions) NewConsumer() (sarama.Consumer, error)
```

## 🏆 最佳实践

### 1. 配置结构组织

```go
// 应用配置结构
type AppConfig struct {
    // 数据库配置
    MySQL    *options.MySQLOptions    `mapstructure:"mysql"`
    Redis    *options.RedisOptions    `mapstructure:"redis"`
    
    // 服务配置
    HTTP     *options.HTTPOptions     `mapstructure:"http"`
    GRPC     *options.GRPCOptions     `mapstructure:"grpc"`
    
    // 安全配置
    TLS      *options.TLSOptions      `mapstructure:"tls"`
    JWT      *options.JWTOptions      `mapstructure:"jwt"`
    
    // 中间件配置
    Health   *options.HealthOptions   `mapstructure:"health"`
    Metrics  *options.MetricsOptions  `mapstructure:"metrics"`
}

// 创建默认配置
func NewAppConfig() *AppConfig {
    return &AppConfig{
        MySQL:   options.NewMySQLOptions(),
        Redis:   options.NewRedisOptions(),
        HTTP:    options.NewHTTPOptions(),
        GRPC:    options.NewGRPCOptions(),
        TLS:     options.NewTLSOptions(),
        JWT:     options.NewJWTOptions(),
        Health:  options.NewHealthOptions(),
        Metrics: options.NewMetricsOptions(),
    }
}
```

### 2. 命令行参数处理

```go
func setupCommandLine() *AppConfig {
    config := NewAppConfig()
    
    // 创建根命令
    rootCmd := &cobra.Command{
        Use:   "myapp",
        Short: "My Application",
        Run: func(cmd *cobra.Command, args []string) {
            runApp(config)
        },
    }
    
    // 添加所有配置的命令行参数
    fs := rootCmd.Flags()
    config.MySQL.AddFlags(fs, "storage")
    config.Redis.AddFlags(fs, "cache")
    config.HTTP.AddFlags(fs, "server")
    config.GRPC.AddFlags(fs, "grpc")
    config.TLS.AddFlags(fs, "tls")
    config.JWT.AddFlags(fs, "auth")
    config.Health.AddFlags(fs, "health")
    config.Metrics.AddFlags(fs, "metrics")
    
    return config
}
```

### 3. 配置验证

```go
func validateConfig(config *AppConfig) error {
    var allErrors []error
    
    // 验证各个组件配置
    if errs := config.MySQL.Validate(); len(errs) > 0 {
        allErrors = append(allErrors, errs...)
    }
    
    if errs := config.Redis.Validate(); len(errs) > 0 {
        allErrors = append(allErrors, errs...)
    }
    
    if errs := config.HTTP.Validate(); len(errs) > 0 {
        allErrors = append(allErrors, errs...)
    }
    
    // 交叉验证
    if config.TLS.CertFile != "" && config.HTTP.BindPort == 80 {
        allErrors = append(allErrors, 
            fmt.Errorf("TLS配置不应该使用HTTP端口80"))
    }
    
    if len(allErrors) > 0 {
        return fmt.Errorf("配置验证失败: %v", allErrors)
    }
    
    return nil
}
```

### 4. 环境特定配置

```go
// 开发环境配置
func NewDevelopmentConfig() *AppConfig {
    config := NewAppConfig()
    
    // 开发数据库配置
    config.MySQL.Addr = "localhost:3306"
    config.MySQL.Database = "myapp_dev"
    config.MySQL.LogLevel = 4 // 详细日志
    config.MySQL.EnableMetrics = true
    config.MySQL.EnableTrace = true
    
    // 开发Redis配置
    config.Redis.Addr = "localhost:6379"
    config.Redis.DB = 1 // 使用不同的数据库
    
    // 开发HTTP配置
    config.HTTP.BindPort = 8080
    config.HTTP.EnableMetrics = true
    
    return config
}

// 生产环境配置
func NewProductionConfig() *AppConfig {
    config := NewAppConfig()
    
    // 生产数据库配置
    config.MySQL.MaxIdleConnections = 50
    config.MySQL.MaxOpenConnections = 200
    config.MySQL.LogLevel = 1 // 静默日志
    config.MySQL.EnableMetrics = true
    config.MySQL.EnableTrace = false // 生产环境可能关闭追踪
    
    // 生产HTTP配置
    config.HTTP.BindPort = 443
    config.HTTP.ReadTimeout = 30 * time.Second
    config.HTTP.WriteTimeout = 30 * time.Second
    
    // 启用TLS
    config.TLS.CertFile = "/etc/ssl/certs/app.crt"
    config.TLS.KeyFile = "/etc/ssl/private/app.key"
    
    return config
}
```

### 5. 动态配置重载

```go
func setupConfigWatcher(config *AppConfig) {
    viper.WatchConfig()
    viper.OnConfigChange(func(e fsnotify.Event) {
        log.Info("配置文件发生变化", "file", e.Name)
        
        // 重新加载配置
        newConfig := NewAppConfig()
        if err := viper.Unmarshal(newConfig); err != nil {
            log.Error("重新加载配置失败", err)
            return
        }
        
        // 验证新配置
        if err := validateConfig(newConfig); err != nil {
            log.Error("新配置验证失败", err)
            return
        }
        
        // 应用新配置（需要根据具体情况实现）
        applyNewConfig(config, newConfig)
        
        log.Info("配置重载成功")
    })
}

func applyNewConfig(oldConfig, newConfig *AppConfig) {
    // 比较配置变化并应用
    if oldConfig.MySQL.LogLevel != newConfig.MySQL.LogLevel {
        // 更新日志级别
        updateDatabaseLogLevel(newConfig.MySQL.LogLevel)
    }
    
    if oldConfig.HTTP.ReadTimeout != newConfig.HTTP.ReadTimeout {
        // 更新HTTP超时
        updateHTTPTimeout(newConfig.HTTP.ReadTimeout)
    }
    
    // 更新配置引用
    *oldConfig = *newConfig
}
```

## 📄 配置示例

### YAML配置文件

```yaml
# config.yaml
mysql:
  addr: "localhost:3306"
  username: "myapp"
  password: "password123"
  database: "myapp"
  max-idle-connections: 10
  max-open-connections: 100
  max-connection-life-time: "1h"
  log-level: 1
  enable-metrics: true
  metrics-name: "main_db"
  enable-trace: true

redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  pool-size: 100
  min-idle-conns: 10
  max-retries: 3
  dial-timeout: "5s"
  read-timeout: "3s"
  write-timeout: "3s"
  enable-metrics: true
  metrics-name: "main_cache"

http:
  bind-address: "0.0.0.0"
  bind-port: 8080
  read-timeout: "30s"
  write-timeout: "30s"
  idle-timeout: "120s"
  max-header-bytes: 1048576
  enable-metrics: true
  metrics-path: "/metrics"
  enable-cors: true
  cors-allowed-origins: ["*"]

grpc:
  bind-address: "0.0.0.0"
  bind-port: 9090
  max-receive-message-size: 4194304
  max-send-message-size: 4194304
  max-concurrent-streams: 1000
  connection-timeout: "5s"
  enable-reflection: true
  enable-metrics: true
  enable-trace: true

tls:
  cert-file: "/etc/ssl/certs/app.crt"
  key-file: "/etc/ssl/private/app.key"
  ca-file: "/etc/ssl/certs/ca.crt"
  server-name: "myapp.example.com"
  insecure-skip-verify: false
  min-version: "1.2"
  max-version: "1.3"

jwt:
  secret-key: "your-secret-key"
  algorithm: "HS256"
  token-expiration: "24h"
  refresh-expiration: "168h"
  issuer: "myapp"
  audience: "users"

health:
  check-interval: "30s"
  timeout: "10s"
  failure-threshold: 3
  success-threshold: 1
  enable-metrics: true
  checks:
    - name: "database"
      type: "mysql"
      endpoint: "localhost:3306"
      timeout: "5s"
    - name: "cache"
      type: "redis"
      endpoint: "localhost:6379"
      timeout: "3s"

metrics:
  enabled: true
  path: "/metrics"
  port: 9090
  namespace: "myapp"
  subsystem: "api"
  enable-cpu: true
  enable-memory: true
  enable-gc: true

consul:
  address: "localhost:8500"
  scheme: "http"
  datacenter: "dc1"
  token: ""
  timeout: "10s"

etcd:
  endpoints: ["localhost:2379"]
  username: ""
  password: ""
  dial-timeout: "5s"
  auto-sync-interval: "30s"
  enable-tls: false

kafka:
  brokers: ["localhost:9092"]
  client-id: "myapp"
  version: "2.8.0"
  enable-sasl: false
  sasl-mechanism: "PLAIN"
  sasl-username: ""
  sasl-password: ""
  enable-tls: false

jaeger:
  service-name: "myapp"
  rpc-metrics: true
  tags: "env=development,version=1.0.0"
  sampler-type: "const"
  sampler-param: 1.0
  local-agent-host-port: "localhost:6832"
  collector-endpoint: "http://localhost:14268/api/traces"
```

### 环境变量配置

```bash
# 数据库配置
export MYSQL_ADDR="localhost:3306"
export MYSQL_USERNAME="myapp"
export MYSQL_PASSWORD="password123"
export MYSQL_DATABASE="myapp"
export MYSQL_ENABLE_METRICS="true"
export MYSQL_ENABLE_TRACE="true"

# Redis配置
export REDIS_ADDR="localhost:6379"
export REDIS_PASSWORD=""
export REDIS_DB="0"
export REDIS_ENABLE_METRICS="true"

# HTTP服务配置
export HTTP_BIND_ADDRESS="0.0.0.0"
export HTTP_BIND_PORT="8080"
export HTTP_ENABLE_METRICS="true"

# gRPC服务配置
export GRPC_BIND_ADDRESS="0.0.0.0"
export GRPC_BIND_PORT="9090"
export GRPC_ENABLE_METRICS="true"
export GRPC_ENABLE_TRACE="true"
```

---

## 🤝 贡献指南

1. **新组件支持**: 添加新的配置选项组件
2. **功能改进**: 改进现有配置选项的功能
3. **文档完善**: 补充使用文档和示例
4. **测试用例**: 添加配置验证和使用测试

## 📄 许可证

本项目采用MIT许可证，详见LICENSE文件。

---

**版本**: v2.0.0  
**最后更新**: 2025-01-15  
**维护者**: Go-Protoc Team