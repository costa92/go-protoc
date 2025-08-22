package routine

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Manager 统一的goroutine管理器
type Manager struct {
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	running int64 // 使用原子操作计数

	// 配置
	maxGoroutines int
	timeout       time.Duration
}

// NewManager 创建新的goroutine管理器
func NewManager(maxGoroutines int, timeout time.Duration) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	if maxGoroutines <= 0 {
		maxGoroutines = runtime.NumCPU() * 10 // 默认CPU核数的10倍
	}

	if timeout <= 0 {
		timeout = 30 * time.Second // 默认30秒超时
	}

	return &Manager{
		ctx:           ctx,
		cancel:        cancel,
		maxGoroutines: maxGoroutines,
		timeout:       timeout,
	}
}

// Go 启动一个受管理的goroutine
func (m *Manager) Go(fn func(ctx context.Context)) error {
	// 检查是否超过最大goroutine数
	if current := atomic.LoadInt64(&m.running); current >= int64(m.maxGoroutines) {
		return ErrTooManyGoroutines
	}

	m.wg.Add(1)
	atomic.AddInt64(&m.running, 1)

	go func() {
		defer func() {
			atomic.AddInt64(&m.running, -1)
			m.wg.Done()
			// 恢复panic
			if r := recover(); r != nil {
				// 可以在这里记录panic信息
			}
		}()

		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(m.ctx, m.timeout)
		defer cancel()

		fn(ctx)
	}()

	return nil
}

// GoWithTimeout 启动一个带自定义超时的goroutine
func (m *Manager) GoWithTimeout(timeout time.Duration, fn func(ctx context.Context)) error {
	if current := atomic.LoadInt64(&m.running); current >= int64(m.maxGoroutines) {
		return ErrTooManyGoroutines
	}

	m.wg.Add(1)
	atomic.AddInt64(&m.running, 1)

	go func() {
		defer func() {
			atomic.AddInt64(&m.running, -1)
			m.wg.Done()
			if r := recover(); r != nil {
				// 记录panic信息
			}
		}()

		ctx, cancel := context.WithTimeout(m.ctx, timeout)
		defer cancel()

		fn(ctx)
	}()

	return nil
}

// Worker 启动一个工作goroutine，会一直运行直到上下文取消
func (m *Manager) Worker(fn func(ctx context.Context) error) error {
	if current := atomic.LoadInt64(&m.running); current >= int64(m.maxGoroutines) {
		return ErrTooManyGoroutines
	}

	m.wg.Add(1)
	atomic.AddInt64(&m.running, 1)

	go func() {
		defer func() {
			atomic.AddInt64(&m.running, -1)
			m.wg.Done()
			if r := recover(); r != nil {
				// 记录panic信息
			}
		}()

		for {
			select {
			case <-m.ctx.Done():
				return
			default:
				if err := fn(m.ctx); err != nil {
					// 可以在这里记录错误
					return
				}
			}
		}
	}()

	return nil
}

// Stop 停止所有goroutine
func (m *Manager) Stop() {
	m.cancel()
}

// Wait 等待所有goroutine完成
func (m *Manager) Wait() {
	m.wg.Wait()
}

// StopAndWait 停止并等待所有goroutine完成
func (m *Manager) StopAndWait() {
	m.Stop()
	m.Wait()
}

// StopWithTimeout 带超时的停止
func (m *Manager) StopWithTimeout(timeout time.Duration) error {
	m.Stop()

	done := make(chan struct{})
	go func() {
		m.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return ErrShutdownTimeout
	}
}

// RunningCount 返回当前运行的goroutine数量
func (m *Manager) RunningCount() int64 {
	return atomic.LoadInt64(&m.running)
}

// IsRunning 检查管理器是否还在运行
func (m *Manager) IsRunning() bool {
	select {
	case <-m.ctx.Done():
		return false
	default:
		return true
	}
}

// Pool goroutine池实现
type Pool struct {
	manager   *Manager
	taskCh    chan Task
	workerNum int
}

// Task 任务定义
type Task func(ctx context.Context) error

// NewPool 创建goroutine池
func NewPool(workerNum int, bufferSize int, timeout time.Duration) *Pool {
	return &Pool{
		manager:   NewManager(workerNum*2, timeout), // 允许一些额外的goroutine
		taskCh:    make(chan Task, bufferSize),
		workerNum: workerNum,
	}
}

// Start 启动goroutine池
func (p *Pool) Start() error {
	for i := 0; i < p.workerNum; i++ {
		err := p.manager.Worker(func(ctx context.Context) error {
			select {
			case task := <-p.taskCh:
				return task(ctx)
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Submit 提交任务
func (p *Pool) Submit(task Task) error {
	select {
	case p.taskCh <- task:
		return nil
	default:
		return ErrPoolFull
	}
}

// SubmitWithTimeout 带超时的任务提交
func (p *Pool) SubmitWithTimeout(task Task, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case p.taskCh <- task:
		return nil
	case <-ctx.Done():
		return ErrSubmitTimeout
	}
}

// Stop 停止池
func (p *Pool) Stop() {
	close(p.taskCh)
	p.manager.StopAndWait()
}

// StopWithTimeout 带超时的停止
func (p *Pool) StopWithTimeout(timeout time.Duration) error {
	close(p.taskCh)
	return p.manager.StopWithTimeout(timeout)
}

// 全局默认管理器
var defaultManager = NewManager(0, 0) // 使用默认配置

// Go 使用默认管理器启动goroutine
func Go(fn func(ctx context.Context)) error {
	return defaultManager.Go(fn)
}

// GoWithTimeout 使用默认管理器启动带超时的goroutine
func GoWithTimeout(timeout time.Duration, fn func(ctx context.Context)) error {
	return defaultManager.GoWithTimeout(timeout, fn)
}

// Worker 使用默认管理器启动工作goroutine
func Worker(fn func(ctx context.Context) error) error {
	return defaultManager.Worker(fn)
}

// SetDefaultManager 设置默认管理器
func SetDefaultManager(manager *Manager) {
	defaultManager = manager
}

// GetDefaultManager 获取默认管理器
func GetDefaultManager() *Manager {
	return defaultManager
}

// Shutdown 优雅关闭默认管理器
func Shutdown(timeout time.Duration) error {
	return defaultManager.StopWithTimeout(timeout)
}
