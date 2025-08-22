package routine

import "errors"

var (
	// ErrTooManyGoroutines 超过最大goroutine数量限制
	ErrTooManyGoroutines = errors.New("too many goroutines running")

	// ErrShutdownTimeout 关闭超时
	ErrShutdownTimeout = errors.New("shutdown timeout")

	// ErrPoolFull 池已满
	ErrPoolFull = errors.New("pool is full")

	// ErrSubmitTimeout 提交任务超时
	ErrSubmitTimeout = errors.New("submit task timeout")
)
