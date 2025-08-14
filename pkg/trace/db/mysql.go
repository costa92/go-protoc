package db

import (
	"context"
	"fmt"

	gormOtel "gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/gorm"
	ottrace "go.opentelemetry.io/otel/trace"
)

// NewMySQLPlugin creates a new MySQL tracing plugin using GORM's official OpenTelemetry plugin.
// This is the recommended way to add tracing to GORM operations.
func NewMySQLPlugin(opts ...gormOtel.Option) gorm.Plugin {
	return gormOtel.NewPlugin(opts...)
}

// NewMySQLPluginWithOptions creates a MySQL tracing plugin with custom options.
func NewMySQLPluginWithOptions() gorm.Plugin {
	return gormOtel.NewPlugin(
		// 配置追踪选项
		gormOtel.WithDBSystem("mysql"),
		// 注意：新版API可能有所不同，具体选项需要查看文档
	)
}

// WithContext returns a new DB instance with the given context.
// This is useful for propagating trace context.
func WithContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.WithContext(ctx)
}

// StartTransaction starts a traced database transaction.
// 使用 GORM 的官方插件，事务操作会自动被追踪
func StartTransaction(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.WithContext(ctx).Begin()
}

// CommitTransaction commits a database transaction.
func CommitTransaction(tx *gorm.DB) error {
	return tx.Commit().Error
}

// RollbackTransaction rolls back a database transaction.
func RollbackTransaction(tx *gorm.DB) error {
	return tx.Rollback().Error
}

// ExecuteInTransaction executes a function within a database transaction with automatic tracing.
func ExecuteInTransaction(ctx context.Context, db *gorm.DB, fn func(*gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// TracedQuery wraps a GORM query with additional context for better tracing.
func TracedQuery(ctx context.Context, db *gorm.DB, operation string, fn func(*gorm.DB) error) error {
	// GORM 的官方插件会自动处理追踪，我们只需要确保上下文传递
	db = db.WithContext(ctx)
	
	// 如果需要，可以添加额外的 span 属性
	if span := ottrace.SpanFromContext(ctx); span.IsRecording() {
		span.SetName(fmt.Sprintf("mysql.%s", operation))
	}
	
	return fn(db)
}
