package grpc

import (
	"context"
	"time"

	"github.com/costa92/go-protoc/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryLoggingInterceptor 是一个 gRPC 一元拦截器，用于记录请求信息
func UnaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		logger.WithValues(
			"method", info.FullMethod,
			"duration", duration,
			"error", err,
		).Infof("gRPC request")

		return resp, err
	}
}

// StreamLoggingInterceptor 是一个 gRPC 流拦截器，用于记录请求信息
func StreamLoggingInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		err := handler(srv, ss)
		duration := time.Since(start)

		logger.WithValues(
			"method", info.FullMethod,
			"duration", duration,
			"error", err,
		).Infof("gRPC stream request")

		return err
	}
}

// UnaryRecoveryInterceptor 是一个 gRPC 一元拦截器，用于从 panic 中恢复
func UnaryRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.WithValues(
					"method", info.FullMethod,
					"panic", r,
				).Errorf("gRPC panic recovered")
				err = status.Errorf(codes.Internal, "panic: %v", r)
			}
		}()
		return handler(ctx, req)
	}
}

// StreamRecoveryInterceptor 是一个 gRPC 流拦截器，用于从 panic 中恢复
func StreamRecoveryInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.WithValues(
					"method", info.FullMethod,
					"panic", r,
				).Errorf("gRPC stream panic recovered")
				err = status.Errorf(codes.Internal, "panic: %v", r)
			}
		}()
		return handler(srv, ss)
	}
}
