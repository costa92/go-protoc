package logger

import (
	"context"
	"sync"
	"sync/atomic"
)

// DynamicLogger 支持动态配置的 Logger 包装器
type DynamicLogger struct {
	mu     sync.RWMutex
	logger atomic.Pointer[Logger] // 存储 Logger 接口指针
	opts   *LogsOptions
}

// NewDynamicLogger 创建支持动态配置的 logger
func NewDynamicLogger(opts *LogsOptions) (*DynamicLogger, error) {
	// 创建基础logger时需要临时禁用动态配置，避免无限递归
	baseOpts := opts.clone()
	baseOpts.Dynamic = false

	logger, err := NewLogger(baseOpts)
	if err != nil {
		return nil, err
	}

	dl := &DynamicLogger{
		opts: opts.clone(),
	}
	dl.logger.Store(&logger)

	return dl, nil
}

// UpdateConfig 动态更新日志配置
func (d *DynamicLogger) UpdateConfig(opts *LogsOptions) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 创建新logger时需要临时禁用动态配置
	baseOpts := opts.clone()
	baseOpts.Dynamic = false

	logger, err := NewLogger(baseOpts)
	if err != nil {
		return err
	}

	d.opts = opts.clone()
	d.logger.Store(&logger)
	return nil
}

// UpdateLevel 动态更新日志级别
func (d *DynamicLogger) UpdateLevel(level string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.opts.Level = level

	// 创建新logger时需要临时禁用动态配置
	baseOpts := d.opts.clone()
	baseOpts.Dynamic = false

	logger, err := NewLogger(baseOpts)
	if err != nil {
		return err
	}

	d.logger.Store(&logger)
	return nil
}

// GetOptions 获取当前配置（只读）
func (d *DynamicLogger) GetOptions() LogsOptions {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return *d.opts.clone()
}

// 实现 Logger 接口的所有方法，委托给内部 logger

func (d *DynamicLogger) Debug(args ...interface{}) {
	(*d.logger.Load()).Debug(args...)
}

func (d *DynamicLogger) Info(args ...interface{}) {
	(*d.logger.Load()).Info(args...)
}

func (d *DynamicLogger) Warn(args ...interface{}) {
	(*d.logger.Load()).Warn(args...)
}

func (d *DynamicLogger) Error(args ...interface{}) {
	(*d.logger.Load()).Error(args...)
}

func (d *DynamicLogger) Fatal(args ...interface{}) {
	(*d.logger.Load()).Fatal(args...)
}

func (d *DynamicLogger) Debugf(template string, args ...interface{}) {
	(*d.logger.Load()).Debugf(template, args...)
}

func (d *DynamicLogger) Infof(template string, args ...interface{}) {
	(*d.logger.Load()).Infof(template, args...)
}

func (d *DynamicLogger) Warnf(template string, args ...interface{}) {
	(*d.logger.Load()).Warnf(template, args...)
}

func (d *DynamicLogger) Errorf(template string, args ...interface{}) {
	(*d.logger.Load()).Errorf(template, args...)
}

func (d *DynamicLogger) Fatalf(template string, args ...interface{}) {
	(*d.logger.Load()).Fatalf(template, args...)
}

func (d *DynamicLogger) Debugw(msg string, keysAndValues ...interface{}) {
	(*d.logger.Load()).Debugw(msg, keysAndValues...)
}

func (d *DynamicLogger) Infow(msg string, keysAndValues ...interface{}) {
	(*d.logger.Load()).Infow(msg, keysAndValues...)
}

func (d *DynamicLogger) Warnw(msg string, keysAndValues ...interface{}) {
	(*d.logger.Load()).Warnw(msg, keysAndValues...)
}

func (d *DynamicLogger) Errorw(msg string, keysAndValues ...interface{}) {
	(*d.logger.Load()).Errorw(msg, keysAndValues...)
}

func (d *DynamicLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	(*d.logger.Load()).Fatalw(msg, keysAndValues...)
}

func (d *DynamicLogger) With(keyValues ...interface{}) Logger {
	return (*d.logger.Load()).With(keyValues...)
}

func (d *DynamicLogger) WithCtx(ctx context.Context, keyValues ...interface{}) Logger {
	return (*d.logger.Load()).WithCtx(ctx, keyValues...)
}

func (d *DynamicLogger) WithCallerSkip(skip int) Logger {
	return (*d.logger.Load()).WithCallerSkip(skip)
}

func (d *DynamicLogger) SetLevel(level Level) {
	(*d.logger.Load()).SetLevel(level)
}

// 确保 DynamicLogger 实现了 Logger 接口
var _ Logger = (*DynamicLogger)(nil)
