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
	GRPCOptions  *genericoptions.GRPCOptions
	HTTPOptions  *genericoptions.HTTPOptions
	TLSOptions   *genericoptions.TLSOptions
	MySQLOptions *genericoptions.MySQLOptions
	JWTOptions   *genericoptions.JWTOptions // Added JWT Options
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
	return server.KratosAppConfig{
		ID:        ID,
		Name:      Name,
		Version:   Version,
		Metadata:  map[string]string{},
		Registrar: registrar,
	}
}

func NewMiddlewares(logger krtlog.Logger, val validate.RequestValidator, jwtOpts *genericoptions.JWTOptions) []middleware.Middleware {
	return []middleware.Middleware{
		logging.Server(logger), // Logging early
		tracing.Server(),       // Tracing
		i18nmw.Translator(i18n.WithLanguage(language.English), i18n.WithFS(locales.Locales)), // i18n
		authn.ServerJWTAuth(jwtOpts), // JWT Authentication
		validate.Validator(val),      // Validation after auth
	}
}

func ProvideKratosLogger() krtlog.Logger {
	return server.NewKratosLogger(ID, Name, Version)
}

func ProvideRegistrar() registry.Registrar {
	return nil // 返回空注册器，如果不需要服务注册
}
