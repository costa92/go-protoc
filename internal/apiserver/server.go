package apiserver

import (
	"context"
	"os"

	"github.com/costa92/go-protoc/v2/internal/pkg/middleware/logging"
	"github.com/costa92/go-protoc/v2/internal/pkg/middleware/tracing"
	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
	"github.com/costa92/go-protoc/v2/pkg/core"
	"github.com/costa92/go-protoc/v2/pkg/db"
	genericoptions "github.com/costa92/go-protoc/v2/pkg/options"
	"github.com/costa92/go-protoc/v2/pkg/server"
	"github.com/costa92/go-protoc/v2/pkg/version"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/registry"

	krtlog "github.com/go-kratos/kratos/v2/log"
)

var (
	// Name is the name of the compiled software.
	Name = "apiserver"

	ID, _ = os.Hostname()

	Version = version.Get().String()
)

type Config struct {
	GRPCOptions  *genericoptions.GRPCOptions
	HTTPOptions  *genericoptions.HTTPOptions
	TLSOptions   *genericoptions.TLSOptions
	MySQLOptions *genericoptions.MySQLOptions
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

	var mysqlOptions db.MySQLOptions
	_ = core.Copy(&mysqlOptions, cfg.MySQLOptions)

	srv, err := InitializeWebServer(ctx.Done(), cfg, &mysqlOptions)
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
	return server.KratosAppConfig{
		ID:        ID,
		Name:      Name,
		Version:   Version,
		Metadata:  map[string]string{},
		Registrar: registrar,
	}
}

func NewMiddlewares(logger krtlog.Logger) []middleware.Middleware {
	return []middleware.Middleware{
		tracing.Server(),
		logging.Server(logger),
	}
}

func ProvideKratosLogger() krtlog.Logger {
	return server.NewKratosLogger(ID, Name, Version)
}

func ProvideRegistrar() registry.Registrar {
	return nil // 返回空注册器，如果不需要服务注册
}
