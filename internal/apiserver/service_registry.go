package apiserver

import (
	"sync"

	"github.com/costa92/go-protoc/pkg/app"
)

// APIGroupInstaller 定义了用于安装 API 组的接口
type APIGroupInstaller interface {
	// Install 将 API 组的路由安装到给定的服务中
	Install(grpcServer *app.GRPCServer, httpServer *app.HTTPServer) error
}

var (
	// apiGroups 存储所有注册的 API 组安装器
	apiGroups []APIGroupInstaller
	// mutex 保护并发访问注册表
	mutex sync.Mutex
)

// RegisterAPIGroup 注册一个 API 组安装器
// 这个函数是线程安全的，应当在服务初始化时调用
func RegisterAPIGroup(installer APIGroupInstaller) {
	mutex.Lock()
	defer mutex.Unlock()
	apiGroups = append(apiGroups, installer)
}

// GetAPIGroups 返回所有注册的 API 组安装器
// 这个函数返回一个副本，以防止修改原始切片
func GetAPIGroups() []APIGroupInstaller {
	mutex.Lock()
	defer mutex.Unlock()
	result := make([]APIGroupInstaller, len(apiGroups))
	copy(result, apiGroups)
	return result
}
