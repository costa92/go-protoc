package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/costa92/go-protoc/pkg/app"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// 测试用的 API 组
type testAPIGroup struct{}

func (t *testAPIGroup) Install(grpcServer *grpc.Server, httpServer *app.HTTPServer) {
	// 注册 HTTP 路由
	router := httpServer.Router()
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("test ok"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// 注册 gRPC 服务
	healthpb.RegisterHealthServer(grpcServer, health.NewServer())
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

// TestApp 测试应用功能
func TestApp(t *testing.T) {
	// 创建应用实例
	app := app.NewApp(":18090", ":18091",
		app.WithHTTPMiddlewares(testHTTPMiddleware),
		app.WithGRPCUnaryInterceptors(testGRPCInterceptor),
	)

	// 安装测试 API 组
	app.InstallAPIGroup(&testAPIGroup{})

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 启动应用
	errChan := make(chan error, 1)
	go func() {
		if err := app.Start(ctx); err != nil && err != context.Canceled {
			errChan <- err
		}
		close(errChan)
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 测试 HTTP 服务器
	t.Run("HTTP Server", func(t *testing.T) {
		resp, err := http.Get("http://localhost:18090/test")
		if err != nil {
			t.Fatalf("Failed to make HTTP request: %v", err)
		}
		defer resp.Body.Close()

		// 验证中间件
		if resp.Header.Get("X-Test") != "test" {
			t.Error("HTTP middleware not working")
		}

		// 验证响应
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		if string(body) != "test ok" {
			t.Errorf("Unexpected response: %s", string(body))
		}
	})

	// 测试 gRPC 服务器
	t.Run("gRPC Server", func(t *testing.T) {
		// 创建 gRPC 连接
		dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Second)
		defer dialCancel()

		// 创建 gRPC 连接
		conn, err := grpc.DialContext(dialCtx, "localhost:18091",
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

		// 等待服务器关闭
		select {
		case err, ok := <-errChan:
			if ok && err != nil {
				t.Errorf("Error during shutdown: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Server shutdown timeout")
		}

		// 验证服务器已关闭
		time.Sleep(time.Second)

		// 验证 HTTP 服务器已关闭
		httpClient := http.Client{Timeout: time.Second}
		if _, err := httpClient.Get("http://localhost:18090/test"); err == nil {
			t.Error("HTTP server still running after shutdown")
		}

		// 验证 gRPC 服务器已关闭
		dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Second)
		defer dialCancel()
		if conn, err := grpc.DialContext(dialCtx, "localhost:18091",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		); err == nil {
			conn.Close()
			t.Error("gRPC server still running after shutdown")
		}
	})
}

// TestAppOptions 测试选项功能
func TestAppOptions(t *testing.T) {
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
	app := app.NewApp(":18092", ":18093",
		app.WithHTTPMiddlewares(middleware1, middleware2),
	)

	// 安装测试路由
	router := app.GetHTTPServer().Router()
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "ok")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// 启动服务器
	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)
	go func() {
		if err := app.Start(ctx); err != nil && err != context.Canceled {
			errChan <- err
		}
		close(errChan)
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 测试中间件链
	resp, err := http.Get("http://localhost:18092/test")
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

	// 关闭服务器
	cancel()

	// 等待服务器关闭
	select {
	case err, ok := <-errChan:
		if ok && err != nil {
			t.Errorf("Error during shutdown: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server shutdown timeout")
	}
}
