package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// UnaryServerInterceptor 创建一个用于跟踪 gRPC 请求的服务端拦截器
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 从 metadata 中提取 trace 上下文
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			propagator := otel.GetTextMapPropagator()
			ctx = propagator.Extract(ctx, metadataCarrier(md))
		}

		// 创建新的 span
		spanName := info.FullMethod
		ctx, span := Tracer.Start(
			ctx,
			spanName,
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
			),
		)
		defer span.End()

		// 记录开始时间
		startTime := time.Now()

		// 执行原始 handler
		resp, err := handler(ctx, req)

		// 记录请求时间
		duration := time.Since(startTime)
		span.SetAttributes(
			attribute.Float64("rpc.duration_ms", float64(duration.Milliseconds())),
		)

		// 如果发生错误，在 span 中记录
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(
				attribute.String("rpc.error_code", s.Code().String()),
				attribute.String("rpc.error_message", s.Message()),
			)
		}

		return resp, err
	}
}

// UnaryClientInterceptor 创建一个用于跟踪 gRPC 客户端请求的拦截器
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// 创建新的客户端 span
		spanName := fmt.Sprintf("Client %s", method)
		ctx, span := Tracer.Start(
			ctx,
			spanName,
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", method),
				attribute.String("rpc.service", cc.Target()),
			),
		)
		defer span.End()

		// 注入 trace 上下文到 metadata
		propagator := otel.GetTextMapPropagator()
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		carrier := metadataCarrier(md)
		propagator.Inject(ctx, carrier)
		ctx = metadata.NewOutgoingContext(ctx, metadata.MD(carrier))

		// 记录开始时间
		startTime := time.Now()

		// 调用原始方法
		err := invoker(ctx, method, req, reply, cc, opts...)

		// 记录请求时间
		duration := time.Since(startTime)
		span.SetAttributes(
			attribute.Float64("rpc.duration_ms", float64(duration.Milliseconds())),
		)

		// 如果发生错误，在 span 中记录
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(
				attribute.String("rpc.error_code", s.Code().String()),
				attribute.String("rpc.error_message", s.Message()),
			)
		}

		return err
	}
}

// metadataCarrier 实现 TextMapCarrier 接口
type metadataCarrier metadata.MD

// Get 返回给定键的值
func (mc metadataCarrier) Get(key string) string {
	values := metadata.MD(mc).Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

// Set 设置给定键的值
func (mc metadataCarrier) Set(key, value string) {
	metadata.MD(mc).Set(key, value)
}

// Keys 返回所有键
func (mc metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc))
	for k := range metadata.MD(mc) {
		keys = append(keys, k)
	}
	return keys
}
