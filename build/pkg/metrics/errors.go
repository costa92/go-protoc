package metrics

import "errors"

// 定义错误类型
var (
	// ErrCollectorNotFound 收集器未找到
	ErrCollectorNotFound = errors.New("collector not found")

	// ErrCollectorDisabled 收集器已禁用
	ErrCollectorDisabled = errors.New("collector is disabled")

	// ErrUnsupportedCollector 不支持的收集器类型
	ErrUnsupportedCollector = errors.New("unsupported collector type")

	// ErrUnsupportedTarget 不支持的目标类型
	ErrUnsupportedTarget = errors.New("unsupported target type")

	// ErrNilTarget 目标为空
	ErrNilTarget = errors.New("target is nil")

	// ErrMetricNotFound 指标未找到
	ErrMetricNotFound = errors.New("metric not found")

	// ErrUnsupportedMetricType 不支持的指标类型
	ErrUnsupportedMetricType = errors.New("unsupported metric type")

	// ErrInvalidConfig 无效配置
	ErrInvalidConfig = errors.New("invalid config")

	// ErrManagerNotInitialized 管理器未初始化
	ErrManagerNotInitialized = errors.New("manager not initialized")

	// ErrRegistryNotSet 注册器未设置
	ErrRegistryNotSet = errors.New("registry not set")

	// ErrDuplicateRegistration 重复注册
	ErrDuplicateRegistration = errors.New("duplicate registration")

	// ErrHealthCheckFailed 健康检查失败
	ErrHealthCheckFailed = errors.New("health check failed")
)
