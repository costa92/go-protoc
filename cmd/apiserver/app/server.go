package app

import (
	"context"
	"fmt"

	"github.com/costa92/go-protoc/v2/cmd/apiserver/app/options"
	"github.com/costa92/go-protoc/v2/internal/apiserver"
	_ "github.com/costa92/go-protoc/v2/internal/apiserver" // Import for error mapper registration
	"github.com/costa92/go-protoc/v2/internal/pkg/contextx"
	"github.com/costa92/go-protoc/v2/internal/pkg/known"
	"github.com/costa92/go-protoc/v2/pkg/app"
	"github.com/costa92/go-protoc/v2/pkg/db"
	"github.com/costa92/go-protoc/v2/pkg/log"
	genericapiserver "k8s.io/apiserver/pkg/server"
)

const commandDesc = `The apiserver server is used to manage users, keys, fees, etc.`

func NewApp() *app.App {
	opts := options.NewServerOptions()
	application := app.NewApp(
		apiserver.Name,
		"Launch a go-protoc apiserver server",
		app.WithDescription(commandDesc),
		app.WithOptions(opts),
		app.WithDefaultValidArgs(),
		app.WithRunFunc(run(opts)),
		// app.WithNoConfig(),
		app.WithLoggerContextExtractor(map[string]func(context.Context) string{
			known.XTraceID: contextx.TraceID,
			known.XUserID:  contextx.UserID,
		}),
	)

	return application
}

func run(opts *options.ServerOptions) app.RunFunc {
	return func() error {
		// Load the configuration options
		cfg, err := opts.Config()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// 启动连接池监控
		if err := startMetricsCollection(cfg); err != nil {
			log.Warnf("Failed to start metrics collection: %v", err)
		}

		ctx := genericapiserver.SetupSignalContext()

		// Build the server using the configuration
		server, err := cfg.NewServer(ctx)
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}

		// 在服务停止时清理监控资源
		defer func() {
			if err := db.ShutdownGlobalMonitoring(); err != nil {
				log.Warnf("Failed to stop metrics collection: %v", err)
			}
			log.Infow("Metrics collection stopped")
		}()

		// Run the server with signal context for graceful shutdown
		return server.Run(ctx)
	}
}

// startMetricsCollection 启动连接池监控
func startMetricsCollection(cfg *apiserver.Config) error {
	// 检查是否启用了任何监控
	mysqlEnabled := cfg.MySQLOptions != nil && cfg.MySQLOptions.EnableMetrics
	// 可以在这里添加对其他数据源的检查

	if !mysqlEnabled {
		log.Infow("Connection pool metrics collection is disabled")
		return nil
	}

	log.Infow("Starting connection pool metrics collection...")

	// 启动全局监控
	config := db.NewMetricsConfig()
	config.Enabled = true
	config.Database.Enabled = true
	config.Database.Name = "default"
	
	if err := db.InitializeGlobalMonitoring(config); err != nil {
		return fmt.Errorf("failed to start global metrics collection: %w", err)
	}

	collector := db.GetGlobalOptimizedCollector()
	status := collector.GetStatus()
	log.Infow("Metrics collection started", "interval", status["collect_interval"])

	return nil
}
