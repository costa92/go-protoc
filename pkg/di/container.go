package di

import (
	"fmt"
	"reflect"
	"sync"
)

// Container 是一个简单的依赖注入容器
type Container struct {
	mu           sync.RWMutex
	services     map[reflect.Type]interface{}
	factoryFuncs map[reflect.Type]interface{}
}

// NewContainer 创建一个新的依赖注入容器
func NewContainer() *Container {
	return &Container{
		services:     make(map[reflect.Type]interface{}),
		factoryFuncs: make(map[reflect.Type]interface{}),
	}
}

// Register 注册一个服务实例
func (c *Container) Register(service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	t := reflect.TypeOf(service)
	c.services[t] = service
}

// RegisterInterface 注册一个接口及其实现
func (c *Container) RegisterInterface(iface interface{}, impl interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	t := reflect.TypeOf(iface)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	c.services[t] = impl
}

// RegisterFactory 注册一个工厂函数，用于延迟创建服务实例
func (c *Container) RegisterFactory(factoryFunc interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	t := reflect.TypeOf(factoryFunc)
	if t.Kind() != reflect.Func {
		panic("factory must be a function")
	}

	// 工厂函数的返回类型作为服务类型
	returnType := t.Out(0)
	c.factoryFuncs[returnType] = factoryFunc
}

// Resolve 解析一个服务实例
func (c *Container) Resolve(servicePtr interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ptrValue := reflect.ValueOf(servicePtr)
	if ptrValue.Kind() != reflect.Ptr {
		return fmt.Errorf("servicePtr must be a pointer")
	}

	// 获取指针指向的类型
	elemType := ptrValue.Elem().Type()

	// 尝试从服务实例映射中获取
	if service, ok := c.services[elemType]; ok {
		ptrValue.Elem().Set(reflect.ValueOf(service))
		return nil
	}

	// 尝试从工厂函数映射中获取
	if factoryFunc, ok := c.factoryFuncs[elemType]; ok {
		factoryValue := reflect.ValueOf(factoryFunc)

		// 调用工厂函数创建实例
		args := []reflect.Value{}
		results := factoryValue.Call(args)

		if len(results) > 0 {
			ptrValue.Elem().Set(results[0])
			return nil
		}
	}

	return fmt.Errorf("service %s not registered", elemType.String())
}

// MustResolve 解析一个服务实例，如果失败则 panic
func (c *Container) MustResolve(servicePtr interface{}) {
	if err := c.Resolve(servicePtr); err != nil {
		panic(err)
	}
}
