package server

import (
	"log"

	"github.com/costa92/go-protoc/pkg/app"
	"github.com/costa92/go-protoc/pkg/options"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
)

// NewGRPCServer 的签名已更新
func NewGRPCServer(opts *options.GRPCOptions, srv *service.APIService, logger *log.Logger) *app.Server {
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		logging.UnaryServerInterceptor(logger),
		validate.UnaryServerInterceptor(),
	}

	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
	}

	s := grpc.NewServer(grpcOpts...)
	v1.RegisterGreeterServer(s, srv)

	return app.NewGRPCServer(opts.Addr, s)
}
