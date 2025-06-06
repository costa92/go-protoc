package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// 测试用的 API 组
type testAPIGroup struct{}

func (t *testAPIGroup) Install(router *mux.Router) error {
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test ok"))
	})
	return nil
}

func (t *testAPIGroup) RegisterGRPC(srv *grpc.Server) error {
	healthpb.RegisterHealthServer(srv, health.NewServer())
	return nil
}

// 测试中间件
func testHTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "test")
		next.ServeHTTP(w, r)
	})
}

// 测试拦截器
func testGRPCInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

// go test -timeout 30s -run ^TestApp$ ./pkg/app -v -count=1
// TestApp 测试应用功能
func TestApp(t *testing.T) {
	// 创建日志器
	logger, _ := zap.NewDevelopment()

	// 创建应用实例
	app := NewApp(":18090", ":18091", logger,
		WithHTTPMiddlewares(testHTTPMiddleware),
		WithGRPCUnaryInterceptors(testGRPCInterceptor),
	)

	// 安装测试 API 组
	if err := app.InstallAPIGroup(&testAPIGroup{}); err != nil {
		t.Fatalf("Failed to install API group: %v", err)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动应用
	errChan := make(chan error, 1)
	go func() {
		if err := app.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 测试 HTTP 服务器
	t.Run("HTTP Server", func(t *testing.T) {
		resp, err := http.Get("http://localhost:18090/api/test")
		if err != nil {
			t.Fatalf("Failed to make HTTP request: %v", err)
		}
		defer resp.Body.Close()

		// 验证中间件
		if resp.Header.Get("X-Test") != "test" {
			t.Error("HTTP middleware not working")
		}

		// 验证响应
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "test ok" {
			t.Errorf("Unexpected response: %s", string(body))
		}
	})

	// 测试 gRPC 服务器
	t.Run("gRPC Server", func(t *testing.T) {
		// 创建 gRPC 连接
		conn, err := grpc.Dial("localhost:18091",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			t.Fatalf("Failed to connect to gRPC server: %v", err)
		}
		defer conn.Close()

		// 创建健康检查客户端
		client := healthpb.NewHealthClient(conn)
		resp, err := client.Check(context.Background(), &healthpb.HealthCheckRequest{})
		if err != nil {
			t.Fatalf("Health check failed: %v", err)
		}
		if resp.Status != healthpb.HealthCheckResponse_SERVING {
			t.Error("Unexpected health status")
		}
	})

	// 测试优雅关闭
	t.Run("Graceful Shutdown", func(t *testing.T) {
		// 触发关闭
		cancel()

		// 等待关闭完成
		shutdownTimeout := time.After(3 * time.Second)
		select {
		case err := <-errChan:
			if err != nil {
				t.Errorf("Error during shutdown: %v", err)
			}
		case <-shutdownTimeout:
			// 验证服务器已关闭
			httpClient := http.Client{Timeout: time.Second}
			_, err := httpClient.Get("http://localhost:18090/api/test")
			if err == nil {
				t.Error("HTTP server still running after shutdown")
			}

			_, err = grpc.Dial("localhost:18091",
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithBlock(),
				grpc.WithTimeout(time.Second),
			)
			if err == nil {
				t.Error("gRPC server still running after shutdown")
			}
		}
	})
}

// TestAppOptions 测试选项功能
func TestAppOptions(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// 测试多个中间件
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-1", "test1")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-2", "test2")
			next.ServeHTTP(w, r)
		})
	}

	// 创建应用实例
	app := NewApp(":18092", ":18093", logger,
		WithHTTPMiddlewares(middleware1, middleware2),
	)

	// 安装测试路由
	router := app.httpServer.Router()
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})

	// 启动服务器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go app.Start(ctx)

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 测试中间件链
	resp, err := http.Get("http://localhost:18092/api/test")
	if err != nil {
		t.Fatalf("Failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// 验证中间件顺序
	if resp.Header.Get("X-Test-1") != "test1" {
		t.Error("First middleware not working")
	}
	if resp.Header.Get("X-Test-2") != "test2" {
		t.Error("Second middleware not working")
	}
}
