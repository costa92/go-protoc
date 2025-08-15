package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/db"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func ExampleTracingDemo() {
	// 初始化 OpenTelemetry
	if err := initTracing(); err != nil {
		log.Fatal("Failed to initialize tracing:", err)
	}

	// 测试数据库追踪
	testDatabaseTracing()
}

func initTracing() error {
	ctx := context.Background()

	// 创建 OTLP gRPC 导出器
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("http://localhost:14250"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	// 创建资源
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("tracing-test"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建 TracerProvider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tp)

	return nil
}

func testDatabaseTracing() {
	ctx := context.Background()

	// 创建根 span
	tracer := otel.Tracer("tracing-test")
	ctx, span := tracer.Start(ctx, "test-database-operations")
	defer span.End()

	fmt.Println("Testing database tracing...")

	// 测试 MySQL 追踪
	testMySQL(ctx)

	// 测试 Redis 追踪
	testRedis(ctx)

	fmt.Println("Database tracing test completed. Check Jaeger UI at http://localhost:16686")
}

func testMySQL(ctx context.Context) {
	fmt.Println("Testing MySQL tracing...")

	// 创建 MySQL 连接配置
	mysqlOpts := &db.MySQLOptions{
		Addr:                  "127.0.0.1:3306",
		Username:              "root",
		Password:              "123456",
		Database:              "krm",
		MaxIdleConnections:    10,
		MaxOpenConnections:    100,
		MaxConnectionLifeTime: time.Hour,
	}

	// 创建 MySQL 连接
	mysqlDB, err := db.NewMySQL(mysqlOpts)
	if err != nil {
		fmt.Printf("MySQL connection failed (expected in demo): %v\n", err)
		return
	}
	defer func() {
		if sqlDB, err := mysqlDB.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// 执行一个简单的查询来生成 trace
	ctx, span := otel.Tracer("mysql-test").Start(ctx, "test-mysql-query")
	defer span.End()

	// 使用上下文执行查询
	result := mysqlDB.WithContext(ctx).Raw("SELECT 1 as test")
	if result.Error != nil {
		fmt.Printf("MySQL query failed (expected in demo): %v\n", result.Error)
	} else {
		fmt.Println("MySQL query executed successfully")
	}
}

func testRedis(ctx context.Context) {
	fmt.Println("Testing Redis tracing...")

	// 创建 Redis 连接配置
	redisOpts := &db.RedisOptions{
		Addr:     "127.0.0.1:6379",
		Password: "",
		Database: 0,
	}

	// 创建 Redis 连接
	redisClient, err := db.NewRedis(redisOpts)
	if err != nil {
		fmt.Printf("Redis connection failed (expected in demo): %v\n", err)
		return
	}
	defer redisClient.Close()

	// 执行一些 Redis 操作来生成 trace
	ctx, span := otel.Tracer("redis-test").Start(ctx, "test-redis-operations")
	defer span.End()

	// SET 操作
	err = redisClient.Set(ctx, "test:key", "test-value", time.Minute).Err()
	if err != nil {
		fmt.Printf("Redis SET failed (expected in demo): %v\n", err)
	} else {
		fmt.Println("Redis SET executed successfully")
	}

	// GET 操作
	val, err := redisClient.Get(ctx, "test:key").Result()
	if err != nil {
		fmt.Printf("Redis GET failed (expected in demo): %v\n", err)
	} else {
		fmt.Printf("Redis GET executed successfully, value: %s\n", val)
	}
}
