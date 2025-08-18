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
	"github.com/costa92/go-protoc/v2/pkg/logger"
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
			logger.Warnf("Failed to start metrics collection: %v", err)
		}

		ctx := genericapiserver.SetupSignalContext()

		// Build the server using the configuration
		server, err := cfg.NewServer(ctx)
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}

		// Run the server with signal context for graceful shutdown
		return server.Run(ctx)
	}
}

// startMetricsCollection 启动连接池监控
func startMetricsCollection(cfg *apiserver.Config) error {
	if cfg.PoolMonitorOptions != nil && cfg.PoolMonitorOptions.Enabled {
		logger.Infow("Connection pool metrics collection is enabled through dependency injection")
		logger.Infow("Metrics will be automatically collected by the PoolMonitor when database connections are established")

		if cfg.PoolMonitorOptions.Database.Enabled {
			logger.Infow("Database pool monitoring enabled",
				"slow_query_threshold", cfg.PoolMonitorOptions.Database.SlowQuery,
				"collect_interval", cfg.PoolMonitorOptions.CollectInterval)
		}

		if cfg.PoolMonitorOptions.Redis.Enabled {
			logger.Infow("Redis pool monitoring enabled",
				"collect_interval", cfg.PoolMonitorOptions.CollectInterval)
		}
	} else {
		logger.Infow("Connection pool metrics collection is disabled")
	}

	return nil
}
