package app

import (
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

// ServerOption 定义服务器选项函数类型
type ServerOption func(interface{})

// WithGRPCUnaryInterceptors 添加 gRPC 一元拦截器
func WithGRPCUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) ServerOption {
	return func(srv interface{}) {
		if s, ok := srv.(*GRPCServer); ok {
			s.unaryInterceptors = append(s.unaryInterceptors, interceptors...)
		}
	}
}

// WithGRPCStreamInterceptors 添加 gRPC 流式拦截器
func WithGRPCStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) ServerOption {
	return func(srv interface{}) {
		if s, ok := srv.(*GRPCServer); ok {
			s.streamInterceptors = append(s.streamInterceptors, interceptors...)
		}
	}
}

// WithHTTPMiddlewares 添加 HTTP 中间件
func WithHTTPMiddlewares(middlewares ...mux.MiddlewareFunc) ServerOption {
	return func(srv interface{}) {
		if s, ok := srv.(*HTTPServer); ok {
			s.middlewares = append(s.middlewares, middlewares...)
		}
	}
}

// WithGRPCOptions 添加 gRPC 服务器选项
func WithGRPCOptions(opts ...grpc.ServerOption) ServerOption {
	return func(srv interface{}) {
		if s, ok := srv.(*GRPCServer); ok {
			s.options = append(s.options, opts...)
		}
	}
}
