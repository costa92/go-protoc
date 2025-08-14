package db

import (
	"context"
	"fmt"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/trace"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	ottrace "go.opentelemetry.io/otel/trace"
)

const (
	RedisTracerName = "redis"
)

// NewTracedRedisClient creates a new Redis client with OpenTelemetry tracing enabled.
func NewTracedRedisClient(opts *redis.Options, tracingOpts ...redisotel.TracingOption) (*redis.Client, error) {
	client := redis.NewClient(opts)
	
	// Enable tracing using the official instrumentation
	if err := redisotel.InstrumentTracing(client, tracingOpts...); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to instrument Redis tracing: %w", err)
	}
	
	return client, nil
}

// NewTracedRedisClusterClient creates a new Redis cluster client with OpenTelemetry tracing enabled.
func NewTracedRedisClusterClient(opts *redis.ClusterOptions, tracingOpts ...redisotel.TracingOption) (*redis.ClusterClient, error) {
	client := redis.NewClusterClient(opts)
	
	// Enable tracing using the official instrumentation
	if err := redisotel.InstrumentTracing(client, tracingOpts...); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to instrument Redis cluster tracing: %w", err)
	}
	
	return client, nil
}

// InstrumentRedisTracing enables OpenTelemetry tracing on an existing Redis client.
func InstrumentRedisTracing(client redis.UniversalClient, opts ...redisotel.TracingOption) error {
	return redisotel.InstrumentTracing(client, opts...)
}

// RedisWithContext creates a Redis operation with context for tracing.
func RedisWithContext(ctx context.Context, operation string, fn func(context.Context) error) error {
	tracer := trace.GetTracer(RedisTracerName)
	ctx, span := tracer.Start(ctx, fmt.Sprintf("redis.%s", operation), trace.WithSpanKind(ottrace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("db.operation", operation),
	)

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	span.SetAttributes(
		attribute.Int64("redis.duration", duration.Microseconds()),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}

// TracedRedisWrapper wraps a Redis client with additional tracing utilities.
type TracedRedisWrapper struct {
	*redis.Client
}

// NewTracedRedisWrapper creates a new traced Redis client wrapper.
func NewTracedRedisWrapper(client *redis.Client) *TracedRedisWrapper {
	return &TracedRedisWrapper{Client: client}
}

// WithTracing wraps a Redis operation with additional tracing context.
func (c *TracedRedisWrapper) WithTracing(ctx context.Context, operation string) context.Context {
	tracer := trace.GetTracer(RedisTracerName)
	ctx, _ = tracer.Start(ctx, fmt.Sprintf("redis.%s", operation), trace.WithSpanKind(ottrace.SpanKindClient))
	return ctx
}
