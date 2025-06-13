package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

type MetricsOptions struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	Path    string `json:"path" mapstructure:"path"`
}

func NewMetricsOptions() *MetricsOptions {
	return &MetricsOptions{
		Enabled: true,       // 默认启用指标收集
		Path:    "/metrics", // 默认指标路径
	}
}

func (o *MetricsOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "metrics.enabled", o.Enabled, "启用指标收集")
	fs.StringVar(&o.Path, "metrics.path", o.Path, "指标收集的HTTP路径")
}

func (o *MetricsOptions) Validate() []error {
	var errs []error
	if o.Path == "" {
		errs = append(errs, fmt.Errorf("metrics path cannot be empty"))
	}
	return errs
}

func (o *MetricsOptions) Complete() error {
	return nil
}
