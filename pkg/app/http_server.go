package app

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	helloworldv1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"

	helloworldv2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
)

type HTTPServerOption func(*HTTPServer) func(*http.Server)

type HTTPServer struct {
	Addr string
}

// WithAddr sets the address to listen on
func WithAddr(addr string) HTTPServerOption {
	return func(s *HTTPServer) func(*http.Server) {
		return func(hs *http.Server) {
			hs.Addr = addr
		}
	}
}

// NewHTTPServer creates a new HTTPServer with the given options
func NewHTTPServer(options ...HTTPServerOption) *HTTPServer {
	server := &HTTPServer{}
	for _, option := range options {
		option(server)
	}

	return server
}

// RunHTTPServer runs the HTTP server
func (s *HTTPServer) RunHTTPServer(v1srv helloworldv1.GreeterServer, v2srv helloworldv2.GreeterServer) (*http.Server, error) {
	server := &http.Server{
		Addr: s.Addr,
	}
	ctx := context.Background()
	mux := runtime.NewServeMux()
	if err := helloworldv1.RegisterGreeterHandlerServer(ctx, mux, v1srv); err != nil {
		return nil, err
	}
	if err := helloworldv2.RegisterGreeterHandlerServer(ctx, mux, v2srv); err != nil {
		return nil, err
	}
	pprofMux := http.NewServeMux()
	setPprofMux(pprofMux)
	pprofMux.Handle("/", mux)
	server.Handler = pprofMux

	return server, nil
}
