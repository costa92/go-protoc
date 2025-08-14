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
	"gorm.io/gorm"
)

// ProvideGormDB provides GORM database connection using options
func ProvideGormDB(cfg *Config) (*gorm.DB, error) {
	return cfg.MySQLOptions.NewDB()
}

func InitializeWebServer(done <-chan struct{}, cfg *Config, mysqlOpts *db.MySQLOptions, jwtOpts *options.JWTOptions) (server.Server, error) {
	wire.Build(
		// Database providers using options
		ProvideGormDB,
		
		// Middleware and server components
		NewMiddlewares,
		ProvideKratosAppConfig,
		ProvideKratosLogger,
		ProvideRegistrar,
		store.ProviderSet,
		biz.ProviderSet,
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
