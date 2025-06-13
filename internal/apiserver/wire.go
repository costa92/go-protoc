//go:build wireinject
// +build wireinject

package apiserver

import (
	"context"

	"github.com/google/wire"

	"github.com/costa92/go-protoc/internal/apiserver/server"
	"github.com/costa92/go-protoc/internal/apiserver/service"
	"github.com/costa92/go-protoc/pkg/logger"
	genericoptions "github.com/costa92/go-protoc/pkg/options"
)

// ProvideServerName 提供服务器名称
func ProvideServerName() string {
	return "apiserver"
}

// ProvideConfig 提供服务器配置
func ProvideConfig() (*Config, error) {
	return &Config{
		GRPCOptions: genericoptions.NewGRPCOptions(),
		HTTPOptions: genericoptions.NewHTTPOptions(),
	}, nil
}

// 将所有提供者集合组合到一起
var allProviderSets = wire.NewSet(
	ProvideServerName,
	ProvideConfig,
	NewServer,
	server.ProviderSet,
	logger.ProviderSet,
	service.ProviderSet,
)

// InitializeServer 初始化 API 服务器
func InitializeServer(ctx context.Context) (*Server, error) {
	wire.Build(allProviderSets)
	return &Server{}, nil
}
