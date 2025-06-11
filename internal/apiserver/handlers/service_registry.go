package handlers

import (
	"fmt"
	"sync"

	"github.com/costa92/go-protoc/pkg/app"
	"github.com/costa92/go-protoc/pkg/logger"
)

// ServiceRegistration 是服务注册的接口
type ServiceRegistration interface {
	// Name 返回服务的名称
	Name() string
	// Register 注册服务到 gRPC 和 HTTP 服务器
	Register(grpcServer *app.GRPCServer, httpServer *app.HTTPServer) error
}

// ServiceRegistry 是服务注册表
type ServiceRegistry struct {
	mu       sync.RWMutex
	services map[string]ServiceRegistration
	logger   logger.Logger
}

// NewServiceRegistry 创建一个新的服务注册表
func NewServiceRegistry(logger logger.Logger) *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]ServiceRegistration),
		logger:   logger,
	}
}

// Register 注册一个服务
func (r *ServiceRegistry) Register(service ServiceRegistration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := service.Name()
	if _, exists := r.services[name]; exists {
		r.logger.Warnw("服务已注册，将被覆盖", "name", name)
	}

	r.services[name] = service
	r.logger.Infow("服务已注册", "name", name)
}

// Get 获取指定名称的服务
func (r *ServiceRegistry) Get(name string) (ServiceRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	if !exists {
		return nil, fmt.Errorf("服务 %s 未注册", name)
	}

	return service, nil
}

// GetAll 获取所有已注册的服务
func (r *ServiceRegistry) GetAll() []ServiceRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ServiceRegistration, 0, len(r.services))
	for _, service := range r.services {
		result = append(result, service)
	}

	return result
}

// RegisterAll 注册所有服务到 gRPC 和 HTTP 服务器
func (r *ServiceRegistry) RegisterAll(grpcServer *app.GRPCServer, httpServer *app.HTTPServer) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, service := range r.services {
		r.logger.Infow("正在注册服务", "name", name)
		if err := service.Register(grpcServer, httpServer); err != nil {
			r.logger.Errorw("注册服务失败", "name", name, "error", err)
			return err
		}
	}

	return nil
}
