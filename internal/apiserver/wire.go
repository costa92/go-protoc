//go:build wireinject
// +build wireinject

package apiserver

import (
	"github.com/google/wire"

	"github.com/costa92/go-protoc/internal/apiserver/server"
	"github.com/costa92/go-protoc/internal/apiserver/service"
	"github.com/costa92/go-protoc/pkg/logger"
)

// ProvideServerName 提供服务器名称
func ProvideServerName() string {
	return "apiserver"
}

// 将所有提供者集合组合到一起
var allProviderSets = wire.NewSet(
	ProvideServerName,
	server.ProviderSet,
	logger.ProviderSet,
	service.ProviderSet,
)

// InitializeServer 初始化 API 服务器
func InitializeServer(<-chan struct, *Config) (*Server, error) {
	wire.Build(allProviderSets)
	return &Server{}, nil
}
