package options

import (
	"github.com/costa92/go-protoc/pkg/log"
	generOptions "github.com/costa92/go-protoc/pkg/options"
	"github.com/spf13/pflag"
)

// Options 是 apiserver 的顶层选项结构。
type Options struct {
	GRPCOptions *generOptions.GRPCOptions `json:"grpc" mapstructure:"grpc"`
	HTTPOptions *generOptions.HTTPOptions `json:"http" mapstructure:"http"`
	Log         *log.LogOptions           `json:"log"  mapstructure:"log"`
}

// NewOptions 创建一个带有完整默认值的顶层 Options。
func NewOptions() *Options {
	return &Options{
		GRPCOptions: generOptions.NewGRPCOptions(),
		HTTPOptions: generOptions.NewHTTPOptions(),
		Log:         log.NewLogOptions(),
	}
}

// AddFlags 将所有组件的标志添加到指定的 FlagSet。
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.GRPCOptions.AddFlags(fs)
	o.HTTPOptions.AddFlags(fs)
	o.Log.AddFlags(fs)
}

// Validate 校验所有选项。
func (o *Options) Validate() []error {
	var errs []error
	errs = append(errs, o.GRPCOptions.Validate()...)
	errs = append(errs, o.HTTPOptions.Validate()...)
	errs = append(errs, o.Log.Validate()...)
	return errs
}
