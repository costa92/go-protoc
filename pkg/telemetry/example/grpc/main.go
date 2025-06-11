package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/costa92/go-protoc/pkg/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	// 这里应该导入您的 gRPC 生成的代码，这里用一个假设的导入路径作为示例
	// pb "github.com/costa92/go-protoc/pkg/api/v1"
)

// 简化起见，这里定义一个示意性的服务接口
type GreeterService interface {
	SayHello(ctx context.Context, request *HelloRequest) (*HelloResponse, error)
}

// 简化的请求和响应结构
type HelloRequest struct {
	Name string
}

type HelloResponse struct {
	Message string
}

// 服务实现
type grpcServer struct {
	// UnimplementedGreeterServer pb.UnimplementedGreeterServer
}

// SayHello 实现 GreeterService 接口
func (s *grpcServer) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "名称不能为空")
	}
	return &HelloResponse{Message: "Hello " + req.Name}, nil
}

// 这是一个健康检查服务的示例实现
func (s *grpcServer) Check(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func main() {
	// 初始化 tracer
	endpoint := os.Getenv("OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4317" // 默认 OTLP 端点
	}

	shutdown, err := telemetry.InitTracer("grpc-server", endpoint)
	if err != nil {
		log.Fatalf("初始化 tracer 失败: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			log.Fatalf("关闭 tracer 失败: %v", err)
		}
	}()

	// 创建 gRPC 服务器，并应用 tracing 拦截器
	s := grpc.NewServer(
		grpc.UnaryInterceptor(telemetry.UnaryServerInterceptor()),
	)

	// 注册您的 gRPC 服务
	// pb.RegisterGreeterServer(s, &grpcServer{})

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("无法监听端口: %v", err)
	}

	// 在单独的 goroutine 中启动服务器
	go func() {
		log.Printf("gRPC 服务器启动在 %s", lis.Addr().String())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("服务器正在关闭...")

	// 优雅停止 gRPC 服务器
	s.GracefulStop()

	log.Println("服务器已关闭")
}
