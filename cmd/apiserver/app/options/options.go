package options

import (
	"github.com/costa92/go-protoc/internal/apiserver"
	genericoptions "github.com/costa92/go-protoc/pkg/options"
	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

type ServerOptions struct {
	GRPCOptions *genericoptions.GRPCOptions `json:"grpc_options" mapstructure:"grpc_options"`
	HTTPOptions *genericoptions.HTTPOptions `json:"http_options" mapstructure:"http_options"`
}

// NewServerOptions 创建一个新的 ServerOptions 实例，包含 GRPC 和 HTTP 的默认选项。
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		GRPCOptions: genericoptions.NewGRPCOptions(),
		HTTPOptions: genericoptions.NewHTTPOptions(),
	}
}

// AddFlags 向指定的 Command 添加服务器相关的标志。
func (o *ServerOptions) AddFlags(cmd *cobra.Command) {
	o.GRPCOptions.AddFlags(cmd.Flags())
	o.HTTPOptions.AddFlags(cmd.Flags())
}

// Validate 校验服务器选项的合法性。
func (o *ServerOptions) Validate() error {
	var errs []error
	errs = append(errs, o.GRPCOptions.Validate()...)
	errs = append(errs, o.HTTPOptions.Validate()...)
	return utilerrors.NewAggregate(errs)
}

// Config 返回一个 apiserver.Config 实例，包含 GRPC 和 HTTP 的配置。
func (o *ServerOptions) Config() (*apiserver.Config, error) {
	return &apiserver.Config{
		GRPCOptions: o.GRPCOptions,
		HTTPOptions: o.HTTPOptions,
	}, nil
}
