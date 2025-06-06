package grpc

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// UnaryLoggingInterceptor 创建一个 gRPC 一元拦截器
func UnaryLoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// 获取元数据
		md, _ := metadata.FromIncomingContext(ctx)
		// 获取对等方信息
		peer, _ := peer.FromContext(ctx)

		// 处理请求
		resp, err := handler(ctx, req)

		// 记录请求信息
		logger.Info("grpc unary request",
			zap.String("method", info.FullMethod),
			zap.Any("metadata", md),
			zap.String("peer_address", peer.Addr.String()),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)

		return resp, err
	}
}

// StreamLoggingInterceptor 创建一个 gRPC 流式拦截器
func StreamLoggingInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// 获取元数据
		md, _ := metadata.FromIncomingContext(ss.Context())
		// 获取对等方信息
		peer, _ := peer.FromContext(ss.Context())

		// 处理流
		err := handler(srv, ss)

		// 记录请求信息
		logger.Info("grpc stream request",
			zap.String("method", info.FullMethod),
			zap.Any("metadata", md),
			zap.String("peer_address", peer.Addr.String()),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)

		return err
	}
}

// UnaryRecoveryInterceptor 创建一个 gRPC 一元恢复拦截器
func UnaryRecoveryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("grpc panic recovered",
					zap.Any("error", r),
					zap.String("method", info.FullMethod),
				)
				err = status.Error(codes.Internal, "Internal Server Error")
			}
		}()
		return handler(ctx, req)
	}
}

// StreamRecoveryInterceptor 创建一个 gRPC 流式恢复拦截器
func StreamRecoveryInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("grpc panic recovered",
					zap.Any("error", r),
					zap.String("method", info.FullMethod),
				)
				err = status.Error(codes.Internal, "Internal Server Error")
			}
		}()
		return handler(srv, ss)
	}
}
