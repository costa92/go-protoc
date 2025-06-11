package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// GRPCOptions 定义了 gRPC 服务的配置。
type GRPCOptions struct {
	Enabled bool          `json:"enabled" mapstructure:"enabled"`
	Addr    string        `json:"addr"    mapstructure:"addr"`
	Timeout time.Duration `json:"timeout" mapstructure:"timeout"`
}

// NewGRPCOptions 创建带有默认值的 GRPCOptions。
func NewGRPCOptions() *GRPCOptions {
	return &GRPCOptions{
		Enabled: true, // gRPC 服务默认启用
		Addr:    "127.0.0.1:9091",
		Timeout: 10 * time.Second,
	}
}

// Validate 校验 gRPC 选项。
func (o *GRPCOptions) Validate() []error {
	var errs []error
	if o.Addr == "" && o.Enabled {
		errs = append(errs, fmt.Errorf("gRPC address cannot be empty when gRPC server is enabled"))
	}
	return errs
}

// AddFlags 向指定的 FlagSet 添加 gRPC 相关的标志。
func (o *GRPCOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "grpc.enabled", o.Enabled, "Enable gRPC server.")
	fs.StringVar(&o.Addr, "grpc.addr", o.Addr, "gRPC server address")
	fs.DurationVar(&o.Timeout, "grpc.timeout", o.Timeout, "gRPC server timeout")
}
