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
		switch s := srv.(type) {
		case *GRPCServer:
			s.unaryInterceptors = append(s.unaryInterceptors, interceptors...)
		case *App:
			s.grpcUnaryInterceptors = append(s.grpcUnaryInterceptors, interceptors...)
		}
	}
}

// WithGRPCStreamInterceptors 添加 gRPC 流式拦截器
func WithGRPCStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) ServerOption {
	return func(srv interface{}) {
		switch s := srv.(type) {
		case *GRPCServer:
			s.streamInterceptors = append(s.streamInterceptors, interceptors...)
		case *App:
			s.grpcStreamInterceptors = append(s.grpcStreamInterceptors, interceptors...)
		}
	}
}

// WithHTTPMiddlewares 添加 HTTP 中间件
func WithHTTPMiddlewares(middlewares ...mux.MiddlewareFunc) ServerOption {
	return func(srv interface{}) {
		switch s := srv.(type) {
		case *HTTPServer:
			s.middlewares = append(s.middlewares, middlewares...)
			// 立即应用中间件到路由器
			for _, middleware := range middlewares {
				s.router.Use(middleware)
			}
		case *App:
			s.httpMiddlewares = append(s.httpMiddlewares, middlewares...)
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
