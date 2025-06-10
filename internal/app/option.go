package app

import (
	"github.com/costa92/go-protoc/pkg/app"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Option func(*Server)

// WithTracerProvider sets the tracer provider for the server.
func WithTracerProvider(tp *sdktrace.TracerProvider) Option {
	return func(s *Server) {
		s.tp = tp
	}
}

// WithApplication sets the application for the server.
func WithApplication(app *app.App) Option {
	return func(s *Server) {
		s.app = app
	}
}

// // WithGRPCServer sets the gRPC server for the server.
// func WithGRPCServer(grpcServer *grpc.Server) Option {
// 	return func(s *Server) {
// 		s.grpcServer = grpcServer
// 	}
// }

// // WithGatewayMux sets the gRPC-Gateway ServeMux for the server.
// func WithGatewayMux(gwmux *runtime.ServeMux) Option {
// 	return func(s *Server) {
// 		s.gwmux = gwmux
// 	}
// }
