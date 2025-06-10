package log

import (
	"sync"

	"go.uber.org/zap"
)

// Logger 定义了项目所需的日志接口。
// 它是 go-kit log.Logger 和 zap.SugaredLogger 的一个子集。
type Logger interface {
	// Debugf 使用 fmt.Sprintf 格式记录一条调试级别的消息。
	Debugf(format string, args ...interface{})
	// Infof 使用 fmt.Sprintf 格式记录一条信息级别的消息。
	Infof(format string, args ...interface{})
	// Warnf 使用 fmt.Sprintf 格式记录一条警告级别的消息。
	Warnf(format string, args ...interface{})
	// Errorf 使用 fmt.Sprintf 格式记录一条错误级别的消息。
	Errorf(format string, args ...interface{})
	// Panicf 使用 fmt.Sprintf 格式记录一条 panic 级别的消息，然后 panic。
	Panicf(format string, args ...interface{})
	// Fatalf 使用 fmt.Sprintf 格式记录一条致命级别的消息，然后调用 os.Exit。
	Fatalf(format string, args ...interface{})

	// Debugw 记录一条调试级别的结构化消息。
	Debugw(msg string, keysAndValues ...interface{})
	// Infow 记录一条信息级别的结构化消息。
	Infow(msg string, keysAndValues ...interface{})
	// Warnw 记录一条警告级别的结构化消息。
	Warnw(msg string, keysAndValues ...interface{})
	// Errorw 记录一条错误级别的结构化消息。
	Errorw(msg string, keysAndValues ...interface{})
	// Panicw 记录一条 panic 级别的结构化消息，然后 panic。
	Panicw(msg string, keysAndValues ...interface{})
	// Fatalw 记录一条致命级别的结构化消息，然后调用 os.Exit。
	Fatalw(msg string, keysAndValues ...interface{})

	// WithValues 添加一些上下文键值对到日志记录器。
	WithValues(keysAndValues ...interface{}) Logger
	// Sync 调用底层 Core 的 Sync 方法，刷新所有缓冲的日志条目。
	// 应用程序应尽量调用 Sync，在退出前刷新日志。
	Sync()
}

var (
	// mu 保护对全局日志记录器的访问
	mu sync.Mutex
	// std 是全局日志记录器
	std Logger
)

// init 在包初始化时设置一个默认的 failsafe 日志记录器。
func init() {
	// 这个 failsafe logger 在 Init() 被调用前使用。
	// 它保证了在主 logger 初始化失败时，日志功能依然可用。
	zapLogger, _ := zap.NewProduction() // zap.NewProduction() 不会返回错误
	std = &SugaredLogger{SugaredLogger: zapLogger.Sugar()}
}

// Init 使用给定的选项初始化全局日志记录器。
// 它会替换掉 failsafe 日志记录器。
func Init(opts *LogOptions) (err error) {
	mu.Lock()
	defer mu.Unlock()

	// 在替换前，同步旧的日志记录器。
	std.Sync()

	logger, err := NewZapLogger(opts)
	if err != nil {
		return err
	}
	std = logger

	return nil
}

// L 返回全局日志记录器。
func L() Logger {
	mu.Lock()
	defer mu.Unlock()
	return std
}

// Sync 同步全局日志记录器。
func Sync() {
	mu.Lock()
	defer mu.Unlock()
	std.Sync()
}

// Debugf 使用全局日志记录器记录一条调试级别的消息。
func Debugf(format string, args ...interface{}) {
	std.Debugf(format, args...)
}

// Infof 使用全局日志记录器记录一条信息级别的消息。
func Infof(format string, args ...interface{}) {
	std.Infof(format, args...)
}

// Warnf 使用全局日志记录器记录一条警告级别的消息。
func Warnf(format string, args ...interface{}) {
	std.Warnf(format, args...)
}

// Errorf 使用全局日志记录器记录一条错误级别的消息。
func Errorf(format string, args ...interface{}) {
	std.Errorf(format, args...)
}

// Panicf 使用全局日志记录器记录一条 panic 级别的消息。
func Panicf(format string, args ...interface{}) {
	std.Panicf(format, args...)
}

// Fatalf 使用全局日志记录器记录一条致命级别的消息。
func Fatalf(format string, args ...interface{}) {
	std.Fatalf(format, args...)
}

// Debugw 使用全局日志记录器记录一条调试级别的结构化消息。
func Debugw(msg string, keysAndValues ...interface{}) {
	std.Debugw(msg, keysAndValues...)
}

// Infow 使用全局日志记录器记录一条信息级别的结构化消息。
func Infow(msg string, keysAndValues ...interface{}) {
	std.Infow(msg, keysAndValues...)
}

// Warnw 使用全局日志记录器记录一条警告级别的结构化消息。
func Warnw(msg string, keysAndValues ...interface{}) {
	std.Warnw(msg, keysAndValues...)
}

// Errorw 使用全局日志记录器记录一条错误级别的结构化消息。
func Errorw(msg string, keysAndValues ...interface{}) {
	std.Errorw(msg, keysAndValues...)
}

// Panicw 使用全局日志记录器记录一条 panic 级别的结构化消息。
func Panicw(msg string, keysAndValues ...interface{}) {
	std.Panicw(msg, keysAndValues...)
}

// Fatalw 使用全局日志记录器记录一条致命级别的结构化消息。
func Fatalw(msg string, keysAndValues ...interface{}) {
	std.Fatalw(msg, keysAndValues...)
}

// WithValues 返回一个包含额外上下文的全局日志记录器。
func WithValues(keysAndValues ...interface{}) Logger {
	mu.Lock()
	defer mu.Unlock()
	return std.WithValues(keysAndValues...)
}
