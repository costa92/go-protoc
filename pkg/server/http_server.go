package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"

	"github.com/costa92/go-protoc/v2/pkg/logger"
	genericoptions "github.com/costa92/go-protoc/v2/pkg/options"
	"github.com/costa92/go-protoc/v2/pkg/version"
)

// HTTPServer 代表一个 HTTP 服务器.
type HTTPServer struct {
	srv *http.Server
}

// NewHTTPServer 创建一个新的 HTTP 服务器实例.
func NewHTTPServer(httpOptions *genericoptions.HTTPOptions, tlsOptions *genericoptions.TLSOptions, handler http.Handler) *HTTPServer {
	var tlsConfig *tls.Config
	if tlsOptions != nil && tlsOptions.UseTLS {
		tlsConfig = tlsOptions.MustTLSConfig()
	}

	return &HTTPServer{
		srv: &http.Server{
			Addr:      httpOptions.Addr,
			Handler:   handler,
			TLSConfig: tlsConfig,
		},
	}
}

// RunOrDie 启动 HTTP 服务器并在出错时记录致命错误.
func (s *HTTPServer) RunOrDie() {
	versionInfo := version.Get()
	logger.Infow("Starting HTTP server",
		"protocol", protocolName(s.srv),
		"addr", s.srv.Addr,
		"service", versionInfo.ServiceName,
		"version", versionInfo.GitVersion,
		"branch", versionInfo.GitBranch,
		"commit", versionInfo.GitCommit[:8], // 显示短commit hash
		"build_date", versionInfo.BuildDate,
	)

	// 默认启动 HTTP 服务器
	serveFn := func() error { return s.srv.ListenAndServe() }
	if s.srv.TLSConfig != nil {
		serveFn = func() error { return s.srv.ListenAndServeTLS("", "") }
	}

	if err := serveFn(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalw("Failed to start HTTP(s) server", "err", err,
			"service", versionInfo.ServiceName,
			"protocol", protocolName(s.srv),
			"addr", s.srv.Addr,
		)
	}
}

// GracefulStop 优雅地关闭 HTTP 服务器.
func (s *HTTPServer) GracefulStop(ctx context.Context) {
	versionInfo := version.Get()
	logger.Infow("Gracefully stopping HTTP server",
		"service", versionInfo.ServiceName,
		"protocol", protocolName(s.srv),
		"addr", s.srv.Addr,
	)
	if err := s.srv.Shutdown(ctx); err != nil {
		logger.Errorw("HTTP server forced to shutdown",
			"error", err,
			"service", versionInfo.ServiceName,
			"protocol", protocolName(s.srv),
			"addr", s.srv.Addr,
		)
	} else {
		logger.Infow("HTTP server stopped gracefully",
			"service", versionInfo.ServiceName,
			"protocol", protocolName(s.srv),
		)
	}
}
