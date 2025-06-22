//go:build wireinject
// +build wireinject

package apiserver

//go:generate go run github.com/google/wire/cmd/wire
import (
	"github.com/costa92/go-protoc/v2/internal/apiserver/biz"
	"github.com/costa92/go-protoc/v2/internal/apiserver/handler"
	"github.com/costa92/go-protoc/v2/internal/apiserver/store"
	"github.com/costa92/go-protoc/v2/pkg/db"
	"github.com/costa92/go-protoc/v2/pkg/server"
	"github.com/costa92/go-protoc/v2/pkg/validation"
	"github.com/google/wire"
)

func InitializeWebServer(<-chan struct{}, *Config, *db.MySQLOptions) (server.Server, error) {
	wire.Build(
		NewMiddlewares,
		ProvideKratosAppConfig,
		ProvideKratosLogger,
		ProvideRegistrar,
		store.ProviderSet,
		biz.ProviderSet,
		db.ProviderSet,
		handler.ProviderSet,
		validation.ProviderSet,
		wire.Struct(new(ServerConfig), "*"), // * 表示注入全部字段
		NewWebServer,
	)
	return nil, nil
}
