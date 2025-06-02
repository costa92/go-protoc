package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	helloworldv1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"
	helloworldv2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type greeterV1Server struct {
	helloworldv1.UnimplementedGreeterServer
}

// post /v1/hello
func (s *greeterV1Server) SayHello(ctx context.Context, req *helloworldv1.HelloRequest) (*helloworldv1.HelloReply, error) {
	return &helloworldv1.HelloReply{Message: "V1: Hello, " + req.GetName()}, nil
}

// get /v1/hello/{name}
func (s *greeterV1Server) SayHelloAgain(ctx context.Context, req *helloworldv1.HelloRequest) (*helloworldv1.HelloReply, error) {
	return &helloworldv1.HelloReply{Message: "V1: Hello, " + req.GetName()}, nil
}

type greeterV2Server struct {
	helloworldv2.UnimplementedGreeterServer
}

func (s *greeterV2Server) SayHello(ctx context.Context, req *helloworldv2.HelloRequest) (*helloworldv2.HelloReply, error) {
	return &helloworldv2.HelloReply{Message: "V2: Hello, " + req.GetName()}, nil
}

func (s *greeterV2Server) SayHelloAgain(ctx context.Context, req *helloworldv2.HelloRequest) (*helloworldv2.HelloReply, error) {
	return &helloworldv2.HelloReply{Message: "V2: Hello, " + req.GetName()}, nil
}

func runGRPCServer(grpcAddr string, v1srv helloworldv1.GreeterServer, v2srv helloworldv2.GreeterServer) *grpc.Server {
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	helloworldv1.RegisterGreeterServer(grpcServer, v1srv)
	helloworldv2.RegisterGreeterServer(grpcServer, v2srv)
	// 注册反射服务
	// reflection.Register(grpcServer)
	go func() {
		log.Printf("gRPC server listening at %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return grpcServer
}

func runHTTPServer(httpAddr string, v1srv helloworldv1.GreeterServer, v2srv helloworldv2.GreeterServer) *http.Server {
	ctx := context.Background()
	mux := runtime.NewServeMux()
	if err := helloworldv1.RegisterGreeterHandlerServer(ctx, mux, v1srv); err != nil {
		log.Fatalf("failed to register v1 http handler: %v", err)
	}
	if err := helloworldv2.RegisterGreeterHandlerServer(ctx, mux, v2srv); err != nil {
		log.Fatalf("failed to register v2 http handler: %v", err)
	}

	// 修正 pprof 路由注册顺序，pprof 路由优先
	pprofMux := http.NewServeMux()
	pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
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

func main() {
	var grpcAddr = flag.String("grpc", ":8100", "gRPC listen address")
	var httpAddr = flag.String("http", ":8080", "HTTP listen address")
	flag.Parse()

	v1srv := &greeterV1Server{}
	v2srv := &greeterV2Server{}
	grpcServer := runGRPCServer(*grpcAddr, v1srv, v2srv)
	httpServer := runHTTPServer(*httpAddr, v1srv, v2srv)

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")
	grpcServer.GracefulStop()
	httpServer.Shutdown(context.Background())
	log.Println("Servers stopped.")
}
