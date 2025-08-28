package apiserver

import (
	"context"
	"os"

	"github.com/costa92/go-protoc/v2/internal/apiserver/pkg/locales"
	i18nmw "github.com/costa92/go-protoc/v2/internal/pkg/middleware/i18n"
	"github.com/costa92/go-protoc/v2/internal/pkg/middleware/logging"
	"github.com/costa92/go-protoc/v2/internal/pkg/middleware/tracing"
	"github.com/costa92/go-protoc/v2/internal/pkg/middleware/validate"
	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
	"github.com/costa92/go-protoc/v2/pkg/core"
	"github.com/costa92/go-protoc/v2/pkg/db"
	"github.com/costa92/go-protoc/v2/pkg/i18n"
	"github.com/costa92/go-protoc/v2/pkg/logger"
	"github.com/costa92/go-protoc/v2/pkg/middleware/authn" // JWT Auth Middleware
	genericoptions "github.com/costa92/go-protoc/v2/pkg/options"
	"github.com/costa92/go-protoc/v2/pkg/server"
	"github.com/costa92/go-protoc/v2/pkg/version"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/registry"
	"golang.org/x/text/language"

	krtlog "github.com/go-kratos/kratos/v2/log"
)

var (
	// Name is the name of the compiled software.
	Name = "apiserver"

	ID, _ = os.Hostname()

	Version = version.Get().String()
)

type Config struct {
	GRPCOptions        *genericoptions.GRPCOptions
	HTTPOptions        *genericoptions.HTTPOptions
	TLSOptions         *genericoptions.TLSOptions
	MySQLOptions       *genericoptions.MySQLOptions
	RedisOptions       *genericoptions.RedisOptions       // Added Redis Options
	JWTOptions         *genericoptions.JWTOptions         // Added JWT Options
	JaegerOptions      *genericoptions.JaegerOptions      // Added Jaeger Options
	SentryOptions      *genericoptions.SentryOptions      // Added Sentry Options
	MetricsOptions     *genericoptions.MetricsOptions     // Added Metrics Options (K8s style)
	PoolMonitorOptions *genericoptions.PoolMonitorOptions // Added Pool Monitor Options
	LogOptions         *logger.LogsOptions                // Added Log Options
}

type Server struct {
	srv server.Server
}

type ServerConfig struct {
	cfg         *Config
	appConfig   server.KratosAppConfig
	handler     v1.ApiServerServer
	middlewares []middleware.Middleware
}

func (cfg *Config) NewServer(ctx context.Context) (*Server, error) {
	// 记录服务版本信息
	versionInfo := version.Get()
	logger.Infow("Initializing API server",
		"service", versionInfo.ServiceName,
		"version", versionInfo.GitVersion,
		"branch", versionInfo.GitBranch,
		"commit", versionInfo.GitCommit[:8],
		"build_date", versionInfo.BuildDate,
		"go_version", versionInfo.GoVersion,
	)

	// 日志库已在应用程序启动时初始化，包含context extractors配置
	// 这里只记录配置信息
	if cfg.LogOptions != nil {
		logger.Infow("Logger configuration", "level", cfg.LogOptions.Level, "format", cfg.LogOptions.Format)
		logger.Debugw("Debug logging enabled", "caller_skip", 2, "enable_color", cfg.LogOptions.EnableColor)
	}

	// 在启动服务前验证所有连接
	logger.Infow("Validating service connections before startup...")

	// 验证MySQL连接
	if cfg.MySQLOptions != nil {
		logger.Infow("Testing MySQL connection...")
		if err := cfg.MySQLOptions.TestConnection(); err != nil {
			logger.Errorw("MySQL connection test failed", "error", err)
			return nil, err
		}
		logger.Infow("MySQL connection test passed")
	}

	// 验证Redis连接
	if cfg.RedisOptions != nil {
		logger.Infow("Testing Redis connection...")
		if err := cfg.RedisOptions.TestConnection(); err != nil {
			logger.Errorw("Redis connection test failed", "error", err)
			return nil, err
		}
		logger.Infow("Redis connection test passed")
	}

	// 验证Jaeger连接
	if cfg.JaegerOptions != nil {
		logger.Infow("Testing Jaeger connection...")
		if err := cfg.JaegerOptions.TestConnection(); err != nil {
			logger.Errorw("Jaeger connection test failed", "error", err)
			return nil, err
		}
		logger.Infow("Jaeger connection test passed")
	}

	logger.Infow("All service connections validated successfully")

	if err := cfg.JaegerOptions.SetTracerProvider(Name); err != nil {
		return nil, err
	}

	var mysqlOptions db.MySQLOptions
	_ = core.Copy(&mysqlOptions, cfg.MySQLOptions)

	// Pass cfg.JWTOptions as the fourth argument
	srv, err := InitializeWebServer(ctx.Done(), cfg, &mysqlOptions, cfg.JWTOptions)
	if err != nil {
		return nil, err
	}
	return &Server{srv: srv}, nil
}

func (s *Server) Run(ctx context.Context) error {
	return server.Serve(ctx, s.srv)
}

func NewWhiteListMatcher() selector.MatchFunc {
	whitelist := make(map[string]struct{})
	return func(ctx context.Context, operation string) bool {
		if _, ok := whitelist[operation]; ok {
			return false
		}
		return true
	}
}

func NewWebServer(serverConfig *ServerConfig) (server.Server, error) {
	grpcsrv := serverConfig.NewGRPCServer()
	httpsrv := serverConfig.NewHTTPServer()
	return server.NewKratosServer(serverConfig.appConfig, grpcsrv, httpsrv)
}

func ProvideKratosAppConfig(registrar registry.Registrar) server.KratosAppConfig {
	// 使用版本系统获取统一的服务信息
	versionInfo := version.Get()
	serviceName := versionInfo.ServiceName
	if serviceName == "" {
		serviceName = Name // 后备到硬编码常量
	}
	serviceVersion := versionInfo.GitVersion
	if serviceVersion == "" {
		serviceVersion = Version // 后备到硬编码常量
	}
	
	return server.KratosAppConfig{
		ID:        ID,
		Name:      serviceName,      // 使用版本系统的服务名
		Version:   serviceVersion,   // 使用版本系统的版本
		Metadata:  map[string]string{},
		Registrar: registrar,
	}
}

func NewMiddlewares(logger krtlog.Logger, val validate.RequestValidator, jwtOpts *genericoptions.JWTOptions, jaegerOpts *genericoptions.JaegerOptions) []middleware.Middleware {
	// 优化: 调整中间件顺序，按性能影响和过滤效果排序
	middlewares := make([]middleware.Middleware, 0, 5) // 预分配容量

	// 1. 验证中间件: 最轻量级，能快速拒绝格式错误的请求
	middlewares = append(middlewares, validate.Validator(val))

	// 2. 认证中间件: 过滤未授权请求，避免后续处理开销
	middlewares = append(middlewares, authn.ServerJWTAuth(jwtOpts))

	// 3. 追踪中间件: 仅在启用时加载，只对通过认证的请求启动追踪
	if jaegerOpts != nil && jaegerOpts.Enabled {
		middlewares = append(middlewares, tracing.Server())
	}

	// 4. 国际化中间件: 为有效请求提供本地化支持
	middlewares = append(middlewares, i18nmw.Translator(i18n.WithLanguage(language.English), i18n.WithFS(locales.Locales)))

	// 5. 日志中间件: 最后记录，只记录完整的请求处理流程
	middlewares = append(middlewares, logging.Server(logger))

	return middlewares
}

func ProvideKratosLogger() krtlog.Logger {
	return server.NewKratosLogger(ID, Name, Version)
}

func ProvideRegistrar() registry.Registrar {
	return nil // 返回空注册器，如果不需要服务注册
}
