package apiserver

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"

	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
)

// NewGRPCServer creates and configures a new gRPC server instance.
func (c *ServerConfig) NewGRPCServer() *grpc.Server {
	opts := []grpc.ServerOption{
		// grpc.WithDiscovery(nil),
		// grpc.WithEndpoint("discovery:///matrix.creation.service.grpc"),
		// Define the middleware chain with variable options.
		grpc.Middleware(c.middlewares...),
	}

	if c.cfg.GRPCOptions.Network != "" {
		opts = append(opts, grpc.Network(c.cfg.GRPCOptions.Network))
	}
	if c.cfg.GRPCOptions.Timeout != 0 {
		opts = append(opts, grpc.Timeout(c.cfg.GRPCOptions.Timeout))
	}
	if c.cfg.GRPCOptions.Addr != "" {
		opts = append(opts, grpc.Address(c.cfg.GRPCOptions.Addr))
	}
	if c.cfg.TLSOptions.UseTLS {
		opts = append(opts, grpc.TLSConfig(c.cfg.TLSOptions.MustTLSConfig()))
	}

	// Create a new gRPC server with the configured options.
	srv := grpc.NewServer(opts...)

	// Register the UserCenter service handler with the gRPC server.
	v1.RegisterApiServerServer(srv, c.handler)

	return srv
}
