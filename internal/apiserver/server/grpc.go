package server

import (
	"github.com/costa92/go-protoc/pkg/app"
	"google.golang.org/grpc"
)

func NewGRPCServer(cfg *Config, opts ...grpc.ServerOption) *app.GRPCServer {
	return app.NewGRPCServer(cfg.GRPCOptions.Addr, opts...)
}
