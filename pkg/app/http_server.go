package app

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type HTTPServerOption func(*HTTPServer) func(*http.Server)

type HTTPServer struct {
	Addr             string
	gatewayRegisters []GatewayRegisterFunc
	logger           *zap.Logger
}

// WithAddr sets the address to listen on
func WithAddr(addr string) HTTPServerOption {
	return func(s *HTTPServer) func(*http.Server) {
		s.Addr = addr // 设置 HTTPServer 的 Addr 字段
		return func(hs *http.Server) {
			hs.Addr = addr
		}
	}
}

// WithHTTPServerLogger sets the logger for the HTTP server
func WithHTTPServerLogger(logger *zap.Logger) HTTPServerOption {
	return func(s *HTTPServer) func(*http.Server) {
		s.logger = logger
		return func(hs *http.Server) {}
	}
}

func WithHTTPServerGatewayRegisterFuncs(gatewayRegisterFuncs ...GatewayRegisterFunc) HTTPServerOption {
	return func(s *HTTPServer) func(*http.Server) {
		return func(hs *http.Server) {
			s.gatewayRegisters = append(s.gatewayRegisters, gatewayRegisterFuncs...)
		}
	}
}

// NewHTTPServer creates a new HTTPServer with the given options
func NewHTTPServer(options ...HTTPServerOption) *HTTPServer {
	server := &HTTPServer{
		logger: zap.NewNop(), // 设置一个默认的 no-op logger
	}
	for _, option := range options {
		option(server)
	}

	return server
}

// RunHTTPServer runs the HTTP server
func (s *HTTPServer) RunHTTPServer() (*http.Server, error) {
	if s.logger == nil {
		s.logger = zap.NewNop()
	}

	server := &http.Server{
		Addr: s.Addr,
	}
	s.logger.Info("starting HTTP server", zap.String("addr", s.Addr))
	ctx := context.Background()
	// 创建 gRPC-Gateway 的 Mux
	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
		// 可以在这里添加其他的 gateway 选项，例如自定义错误处理器
	)
	// 连接到 gRPC 服务器的拨号选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	// 注册 gRPC-Gateway 处理器
	for _, gatewayRegisterFunc := range s.gatewayRegisters {
		if err := gatewayRegisterFunc(ctx, gwMux, "localhost:8100", opts); err != nil {
			s.logger.Error("failed to register gRPC-Gateway", zap.Error(err))
			return nil, err
		}
	}
	router := mux.NewRouter()

	// 挂载 gRPC-Gateway
	// 注意: 这种方式会将所有未被其他路由匹配的请求都转发给 gwMux。
	// 如果你的 proto 文件中的 http rule 路径各不相同（如 /v1/... /v2/...），这种方式是可行的。
	router.PathPrefix("/").Handler(gwMux)
	pprofMux := http.NewServeMux()
	setPprofMux(pprofMux)
	pprofMux.Handle("/", router)
	server.Handler = pprofMux

	return server, nil
}
