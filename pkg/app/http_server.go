package app

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	helloworldv1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"

	helloworldv2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
)

func RunHTTPServer(httpAddr string, v1srv helloworldv1.GreeterServer, v2srv helloworldv2.GreeterServer) *http.Server {
	ctx := context.Background()
	mux := runtime.NewServeMux()
	if err := helloworldv1.RegisterGreeterHandlerServer(ctx, mux, v1srv); err != nil {
		log.Fatalf("failed to register v1 http handler: %v", err)
	}
	if err := helloworldv2.RegisterGreeterHandlerServer(ctx, mux, v2srv); err != nil {
		log.Fatalf("failed to register v2 http handler: %v", err)
	}
	pprofMux := http.NewServeMux()
	// 注册 pprof 路由
	setPprofMux(pprofMux)
	// 注册其他路由
	pprofMux.Handle("/", mux)
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: pprofMux,
	}
	go func() {
		log.Printf("HTTP server listening at %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to serve http: %v", err)
		}
	}()
	return httpServer
}
