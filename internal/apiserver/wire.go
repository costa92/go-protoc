//go:build wireinject
// +build wireinject

package apiserver

//go:generate go run github.com/google/wire/cmd/wire
import (
	"github.com/costa92/go-protoc/v2/internal/apiserver/biz"
	"github.com/costa92/go-protoc/v2/internal/apiserver/handler"
	"github.com/costa92/go-protoc/v2/internal/apiserver/pkg/validation"
	"github.com/costa92/go-protoc/v2/internal/apiserver/store"
	"github.com/costa92/go-protoc/v2/internal/pkg/middleware/validate"
	"github.com/costa92/go-protoc/v2/pkg/db"
	"github.com/costa92/go-protoc/v2/pkg/options" // For genericoptions.JWTOptions
	"github.com/costa92/go-protoc/v2/pkg/server"
	genericvalidation "github.com/costa92/go-protoc/v2/pkg/validation"
	"github.com/google/wire"
)

func InitializeWebServer(done <-chan struct{}, cfg *Config, mysqlOpts *db.MySQLOptions, jwtOpts *options.JWTOptions) (server.Server, error) {
	wire.Build(
		// provideJWTOptions is no longer needed as jwtOpts is a direct parameter
		NewMiddlewares, // NewMiddlewares now expects jwtOpts, which Wire will pass from InitializeWebServer's params
		ProvideKratosAppConfig,
		ProvideKratosLogger,
		ProvideRegistrar,
		store.ProviderSet,
		biz.ProviderSet,
		db.ProviderSet,
		handler.ProviderSet,
		wire.NewSet(
			validation.ProviderSet,
			genericvalidation.NewValidator,
			wire.Bind(new(validate.RequestValidator), new(*genericvalidation.Validator)),
		),
		wire.Struct(new(ServerConfig), "*"), // * 表示注入全部字段
		NewWebServer,
	)
	return nil, nil
}
