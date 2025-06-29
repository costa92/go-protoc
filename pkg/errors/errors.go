package errors

import (
	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

// 为了向后兼容，保留原有的错误类型定义
// 但建议使用新的 errorsx 包中的错误处理机制

// MinerStatusError 定义了矿工状态相关的错误类型.
// Deprecated: 建议使用 errorsx.ErrorX 替代
type MinerStatusError string

// Error 实现了 error 接口.
func (e MinerStatusError) Error() string {
	return string(e)
}

// MinerSetStatusError 定义了矿工集状态相关的错误类型.
// Deprecated: 建议使用 errorsx.ErrorX 替代
type MinerSetStatusError string

// Error 实现了 error 接口.
func (e MinerSetStatusError) Error() string {
	return string(e)
}

// 定义矿工相关的错误常量.
// Deprecated: 建议使用新的错误定义
const (
	// InvalidConfigurationMinerError 表示矿工配置无效的错误.
	InvalidConfigurationMinerError MinerStatusError = "InvalidConfiguration"

	// UnsupportedChangeMinerError 表示不支持的矿工变更操作错误.
	UnsupportedChangeMinerError MinerStatusError = "UnsupportedChange"

	// InsufficientResourcesMinerError 表示矿工资源不足的错误.
	InsufficientResourcesMinerError MinerStatusError = "InsufficientResources"

	// CreateMinerError 表示创建矿工失败的错误.
	CreateMinerError MinerStatusError = "CreateError"

	// UpdateMinerError 表示更新矿工失败的错误.
	UpdateMinerError MinerStatusError = "UpdateError"

	// DeleteMinerError 表示删除矿工失败的错误.
	DeleteMinerError MinerStatusError = "DeleteError"

	// JoinClusterTimeoutMinerError 表示矿工加入集群超时的错误.
	JoinClusterTimeoutMinerError MinerStatusError = "JoinClusterTimeoutError"

	// InvalidConfigurationMinerSetError 表示矿工集配置无效的错误.
	InvalidConfigurationMinerSetError MinerSetStatusError = "InvalidConfiguration"
)

// 新的矿工相关错误定义，使用 errorsx 包
var (
	// ErrMinerInvalidConfiguration 矿工配置无效
	ErrMinerInvalidConfiguration = errorsx.New(400, "MINER_INVALID_CONFIGURATION", "Invalid miner configuration").WithI18nKey("errors.miner.invalid_configuration")
	
	// ErrMinerUnsupportedChange 不支持的矿工变更操作
	ErrMinerUnsupportedChange = errorsx.New(400, "MINER_UNSUPPORTED_CHANGE", "Unsupported miner change operation").WithI18nKey("errors.miner.unsupported_change")
	
	// ErrMinerInsufficientResources 矿工资源不足
	ErrMinerInsufficientResources = errorsx.New(400, "MINER_INSUFFICIENT_RESOURCES", "Insufficient miner resources").WithI18nKey("errors.miner.insufficient_resources")
	
	// ErrMinerCreateFailed 创建矿工失败
	ErrMinerCreateFailed = errorsx.New(500, "MINER_CREATE_FAILED", "Failed to create miner").WithI18nKey("errors.miner.create_failed")
	
	// ErrMinerUpdateFailed 更新矿工失败
	ErrMinerUpdateFailed = errorsx.New(500, "MINER_UPDATE_FAILED", "Failed to update miner").WithI18nKey("errors.miner.update_failed")
	
	// ErrMinerDeleteFailed 删除矿工失败
	ErrMinerDeleteFailed = errorsx.New(500, "MINER_DELETE_FAILED", "Failed to delete miner").WithI18nKey("errors.miner.delete_failed")
	
	// ErrMinerJoinClusterTimeout 矿工加入集群超时
	ErrMinerJoinClusterTimeout = errorsx.New(408, "MINER_JOIN_CLUSTER_TIMEOUT", "Miner join cluster timeout").WithI18nKey("errors.miner.join_cluster_timeout")
	
	// ErrMinerNotFound 矿工不存在
	ErrMinerNotFound = errorsx.New(404, "MINER_NOT_FOUND", "Miner not found").WithI18nKey("errors.miner.not_found")
	
	// ErrMinerAlreadyExists 矿工已存在
	ErrMinerAlreadyExists = errorsx.New(409, "MINER_ALREADY_EXISTS", "Miner already exists").WithI18nKey("errors.miner.already_exists")
)

// 矿工错误构建器函数

// NewMinerInvalidConfigurationError 创建矿工配置无效错误
func NewMinerInvalidConfigurationError(minerID, reason string) *errorsx.ErrorX {
	return ErrMinerInvalidConfiguration.
		AddMetadata("miner_id", minerID).
		AddMetadata("reason", reason)
}

// NewMinerNotFoundError 创建矿工不存在错误
func NewMinerNotFoundError(minerID string) *errorsx.ErrorX {
	return ErrMinerNotFound.AddMetadata("miner_id", minerID)
}

// NewMinerCreateFailedError 创建矿工创建失败错误
func NewMinerCreateFailedError(minerID string, err error) *errorsx.ErrorX {
	return ErrMinerCreateFailed.
		AddMetadata("miner_id", minerID).
		WithCause(err)
}

// NewMinerJoinClusterTimeoutError 创建矿工加入集群超时错误
func NewMinerJoinClusterTimeoutError(minerID, clusterID string, timeout int) *errorsx.ErrorX {
	return ErrMinerJoinClusterTimeout.
		AddMetadata("miner_id", minerID).
		AddMetadata("cluster_id", clusterID).
		AddMetadata("timeout_seconds", timeout)
}
