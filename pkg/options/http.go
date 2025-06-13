package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// HTTPOptions 定义了 HTTP 服务的配置。
type HTTPOptions struct {
	Enabled bool          `json:"enabled" mapstructure:"enabled"`
	Addr    string        `json:"addr"    mapstructure:"addr"`
	Timeout time.Duration `json:"timeout" mapstructure:"timeout"`
}

// NewHTTPOptions 创建带有默认值的 HTTPOptions。
func NewHTTPOptions() *HTTPOptions {
	return &HTTPOptions{
		Enabled: true, // HTTP 服务默认启用
		Addr:    "127.0.0.1:8081",
		Timeout: 10 * time.Second,
	}
}

// Validate 校验 HTTP 选项。
func (o *HTTPOptions) Validate() []error {
	var errs []error
	if o.Addr == "" && o.Enabled {
		errs = append(errs, fmt.Errorf("HTTP address cannot be empty when HTTP server is enabled"))
	}
	return errs
}

// AddFlags 向指定的 FlagSet 添加 HTTP 相关的标志。
func (o *HTTPOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "http.enabled", o.Enabled, "Enable HTTP server.")
	fs.StringVar(&o.Addr, "http.addr", o.Addr, "HTTP server address")
	fs.DurationVar(&o.Timeout, "http.timeout", o.Timeout, "HTTP server timeout")
}
