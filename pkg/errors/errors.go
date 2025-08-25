package errors

import (
	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

// Package errors 提供业务相关的预定义错误类型和构建器函数
//
// 本包基于 errorsx 包构建，提供了用户、认证、通用业务等领域的标准错误定义。
// 使用两层架构设计：
//   - errorsx: 错误处理引擎（框架层）
//   - errors: 业务错误目录（应用层）

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
