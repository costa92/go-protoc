package apiserver

import (
	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
	"github.com/costa92/go-protoc/v2/pkg/server" // Import for custom codecs
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/kratos/v2/transport/http/pprof"
	"github.com/go-kratos/swagger-api/openapiv2"
	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (c *ServerConfig) NewHTTPServer() *http.Server {
	opts := []http.ServerOption{
		http.Middleware(c.middlewares...),
		// Add filter options to the middleware chain.
		http.Filter(handlers.CORS(
			handlers.AllowedHeaders([]string{
				"X-Requested-With",
				"Content-Type",
				"Authorization",
				"X-Idempotent-ID",
			}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
			handlers.AllowedOrigins([]string{"*"}),
		)),
		// Apply custom response and error encoders
		http.ResponseEncoder(server.EncodeResponseFunc),
		http.ErrorEncoder(server.EncodeErrorFunc),
	}
	if c.cfg.HTTPOptions.Network != "" {
		opts = append(opts, http.Network(c.cfg.HTTPOptions.Network))
	}
	if c.cfg.HTTPOptions.Timeout != 0 {
		opts = append(opts, http.Timeout(c.cfg.HTTPOptions.Timeout))
	}
	if c.cfg.HTTPOptions.Addr != "" {
		opts = append(opts, http.Address(c.cfg.HTTPOptions.Addr))
	}
	if c.cfg.TLSOptions.UseTLS {
		opts = append(opts, http.TLSConfig(c.cfg.TLSOptions.MustTLSConfig()))
	}

	// Create and return the server instance.
	srv := http.NewServer(opts...)
	h := openapiv2.NewHandler()
	srv.HandlePrefix("/openapi/", h)
	srv.Handle("/metrics", promhttp.Handler())
	srv.Handle("", pprof.NewHandler())

	v1.RegisterApiServerHTTPServer(srv, c.handler)
	return srv
}
