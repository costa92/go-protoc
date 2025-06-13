package app

import "context"

// Server 定义了所有服务（如 gRPC, HTTP）必须实现的通用接口。
type Server interface {
	// Start 负责启动服务。它应该是非阻塞的。
	// context 用于控制服务的生命周期。
	Start(ctx context.Context) error

	// Stop 负责优雅地停止服务。
	Stop(ctx context.Context) error
}
