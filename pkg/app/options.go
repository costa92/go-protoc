package app

import (
	"net"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

// Options 包含了服务器的配置选项。
type Options struct {
	httpAddr               string
	grpcAddr               string
	httpMiddlewares        []mux.MiddlewareFunc
	grpcUnaryInterceptors  []grpc.UnaryServerInterceptor
	grpcStreamInterceptors []grpc.StreamServerInterceptor
	grpcOptions            []grpc.ServerOption
	grpcListener           net.Listener
}

// NewOptions 创建一个带有默认值的新 Options 对象。
func NewOptions() *Options {
	return &Options{
		httpAddr:               ":8080",
		grpcAddr:               ":9090",
		httpMiddlewares:        make([]mux.MiddlewareFunc, 0),
		grpcUnaryInterceptors:  make([]grpc.UnaryServerInterceptor, 0),
		grpcStreamInterceptors: make([]grpc.StreamServerInterceptor, 0),
		grpcOptions:            make([]grpc.ServerOption, 0),
	}
}

// ServerOption 定义了用于配置服务器的函数类型。
type ServerOption func(*Options)

// WithHTTPMiddlewares 添加 HTTP 中间件。
func WithHTTPMiddlewares(mws ...mux.MiddlewareFunc) ServerOption {
	return func(o *Options) {
		o.httpMiddlewares = append(o.httpMiddlewares, mws...)
	}
}

// WithGRPCUnaryInterceptors 添加 gRPC 一元拦截器。
func WithGRPCUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) ServerOption {
	return func(o *Options) {
		o.grpcUnaryInterceptors = append(o.grpcUnaryInterceptors, interceptors...)
	}
}

// WithGRPCStreamInterceptors 添加 gRPC 流拦截器。
func WithGRPCStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) ServerOption {
	return func(o *Options) {
		o.grpcStreamInterceptors = append(o.grpcStreamInterceptors, interceptors...)
	}
}

// WithGRPCOptions 添加 gRPC 服务器选项。
func WithGRPCOptions(opts ...grpc.ServerOption) ServerOption {
	return func(o *Options) {
		o.grpcOptions = append(o.grpcOptions, opts...)
	}
}

// WithGRPCListener 设置 gRPC 服务器的监听器。
func WithGRPCListener(lis net.Listener) ServerOption {
	return func(o *Options) {
		o.grpcListener = lis
	}
}
