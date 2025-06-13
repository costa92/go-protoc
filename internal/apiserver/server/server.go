package server

import (
	"github.com/costa92/go-protoc/pkg/app"
	"github.com/google/wire"
	"github.com/gorilla/mux"
)

// ProviderSet defines a wire provider set.
var ProviderSet = wire.NewSet(NewServer, NewGRPCServer, NewHTTPServer)

func NewServer(httpServer *app.HTTPServer, grpcServer *app.GRPCServer) *[]app.Server {
	return &[]app.Server{httpServer, grpcServer}
}

func NewMiddlewares(cfg *Config) *[]mux.MiddlewareFunc {
	return &[]mux.MiddlewareFunc{}
}
