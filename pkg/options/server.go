package options

import (
	"github.com/spf13/pflag"
)

// ServerOptions 包含所有服务器相关选项
type ServerOptions struct {
	HTTP *HTTPOptions `json:"http" mapstructure:"http"`
	GRPC *GRPCOptions `json:"grpc" mapstructure:"grpc"`
}

// NewServerOptions 创建一个带有默认值的 ServerOptions
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		HTTP: NewHTTPOptions(),
		GRPC: NewGRPCOptions(),
	}
}

// Validate 验证所有服务器选项
func (o *ServerOptions) Validate() []error {
	var errs []error

	// 验证 HTTP 选项
	if o.HTTP != nil {
		errs = append(errs, o.HTTP.Validate()...)
	}

	// 验证 gRPC 选项
	if o.GRPC != nil {
		errs = append(errs, o.GRPC.Validate()...)
	}

	return errs
}

// AddFlags 添加所有服务器相关的命令行标志
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	if o.HTTP != nil {
		o.HTTP.AddFlags(fs)
	}

	if o.GRPC != nil {
		o.GRPC.AddFlags(fs)
	}
}
