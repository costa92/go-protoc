package di

// Provider 定义了一个依赖提供者的接口
type Provider interface {
	// Register 向容器注册所有依赖
	Register(container *Container)
}

// ProviderFunc 是一个实现 Provider 接口的函数类型
type ProviderFunc func(container *Container)

// Register 实现 Provider 接口
func (f ProviderFunc) Register(container *Container) {
	f(container)
}

// Providers 是一组 Provider 的集合
type Providers []Provider

// Register 注册所有提供者
func (ps Providers) Register(container *Container) {
	for _, p := range ps {
		p.Register(container)
	}
}

// NewProviders 创建一个新的 Providers 实例
func NewProviders(providers ...Provider) Providers {
	return providers
}
