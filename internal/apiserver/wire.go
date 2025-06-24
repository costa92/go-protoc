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

// provideJWTOptions extracts JWTOptions from the main Config.
func provideJWTOptions(cfg *Config) *options.JWTOptions {
	return cfg.JWTOptions
}

func InitializeWebServer(<-chan struct{}, *Config, *db.MySQLOptions) (server.Server, error) {
	wire.Build(
		provideJWTOptions, // Provide JWTOptions
		NewMiddlewares,
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
