package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	krtlog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// AsyncLogger 异步日志记录器
type AsyncLogger struct {
	logger  krtlog.Logger
	logChan chan logEntry
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

type logEntry struct {
	ctx     context.Context
	level   krtlog.Level
	keyvals []interface{}
}

func NewAsyncLogger(logger krtlog.Logger, bufferSize int) *AsyncLogger {
	ctx, cancel := context.WithCancel(context.Background())
	al := &AsyncLogger{
		logger:  logger,
		logChan: make(chan logEntry, bufferSize),
		ctx:     ctx,
		cancel:  cancel,
	}
	al.start()
	return al
}

func (al *AsyncLogger) start() {
	al.wg.Add(1)
	go func() {
		defer al.wg.Done()
		for {
			select {
			case entry := <-al.logChan:
				_ = al.logger.Log(entry.level, entry.keyvals...)
			case <-al.ctx.Done():
				// 处理剩余日志
				for len(al.logChan) > 0 {
					entry := <-al.logChan
					_ = al.logger.Log(entry.level, entry.keyvals...)
				}
				return
			}
		}
	}()
}

func (al *AsyncLogger) LogAsync(ctx context.Context, level krtlog.Level, keyvals ...interface{}) {
	entry := logEntry{
		ctx:     ctx,
		level:   level,
		keyvals: keyvals,
	}
	select {
	case al.logChan <- entry:
		// 日志成功发送到队列
	default:
		// 队列满时直接记录，避免阻塞
		_ = al.logger.Log(level, keyvals...)
	}
}

func (al *AsyncLogger) Close() {
	al.cancel()
	al.wg.Wait()
	close(al.logChan)
}

var defaultAsyncLogger *AsyncLogger
var once sync.Once

func getAsyncLogger(logger krtlog.Logger) *AsyncLogger {
	once.Do(func() {
		defaultAsyncLogger = NewAsyncLogger(logger, 1000) // 1000个日志条目的缓冲区
	})
	return defaultAsyncLogger
}

func Server(logger krtlog.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, rq any) (reply any, err error) {
			var (
				code      int32
				reason    string
				kind      string
				operation string
			)
			startTime := time.Now()
			if tr, ok := transport.FromServerContext(ctx); ok {
				kind = tr.Kind().String()
				operation = tr.Operation()
			}
			reply, err = handler(ctx, rq)
			if se := errors.FromError(err); se != nil {
				code = se.Code
				reason = se.Reason
			}

			// 优化: 只对错误请求或高级别日志进行详细记录
			latency := time.Since(startTime).Seconds()
			if err != nil || latency > 0.01 { // 只记录错误请求或慢请求(>500ms)
				level, stack := extractError(err)
				// 使用异步日志记录，避免阻塞请求处理
				asyncLogger := getAsyncLogger(logger)
				asyncLogger.LogAsync(ctx, level,
					"kind", "server",
					"component", kind,
					"operation", operation,
					"args", extractArgs(rq),
					"code", code,
					"reason", reason,
					"stack", stack,
					"latency", latency,
				)
			}
			return reply, err
		}
	}
}

// Client is a client logging middleware.
func Client(logger krtlog.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, rq any) (reply any, err error) {
			var (
				code      int32
				reason    string
				kind      string
				operation string
			)
			startTime := time.Now()
			if tr, ok := transport.FromClientContext(ctx); ok {
				kind = tr.Kind().String()
				operation = tr.Operation()
			}
			reply, err = handler(ctx, rq)
			if se := errors.FromError(err); se != nil {
				code = se.Code
				reason = se.Reason
			}
			level, stack := extractError(err)
			_ = logger.Log(level,
				"kind", "client",
				"component", kind,
				"operation", operation,
				"args", extractArgs(rq),
				"code", code,
				"reason", reason,
				"stack", stack,
				"latency", time.Since(startTime).Seconds(),
			)
			return reply, err
		}
	}
}

// extractArgs returns the string of the rq.
func extractArgs(rq any) string {
	if stringer, ok := rq.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", rq)
}

// extractError returns the string of the error.
func extractError(err error) (krtlog.Level, string) {
	if err != nil {
		return krtlog.LevelError, fmt.Sprintf("%+v", err)
	}
	return krtlog.LevelInfo, ""
}
