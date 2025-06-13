package server

import (
	"github.com/costa92/go-protoc/pkg/app"
	"github.com/gorilla/mux"
)

func NewHTTPServer(cfg *Config, middlewares ...mux.MiddlewareFunc) *app.HTTPServer {
	return app.NewHTTPServer(cfg.HTTPOptions.Addr, middlewares...)
}
