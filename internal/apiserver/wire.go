//go:build wireinject
// +build wireinject

package apiserver

//go:generate go run github.com/google/wire/cmd/wire
import (
	"github.com/costa92/go-protoc/v2/internal/apiserver/handler"
	"github.com/costa92/go-protoc/v2/pkg/server"
	"github.com/google/wire"
)

func InitializeWebServer(<-chan struct{}, *Config) (server.Server, error) {
	wire.Build(
		handler.ProviderSet,
		NewMiddlewares,
		ProvideKratosAppConfig,
		ProvideKratosLogger,
		ProvideRegistrar,
		wire.Struct(new(ServerConfig), "*"), // * 表示注入全部字段
		NewWebServer,
	)
	return nil, nil
}
