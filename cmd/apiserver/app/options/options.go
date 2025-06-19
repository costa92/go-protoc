package options

import (
	"github.com/costa92/go-protoc/v2/internal/apiserver"
	"github.com/costa92/go-protoc/v2/pkg/app"
	"github.com/costa92/go-protoc/v2/pkg/log"
	genericoptions "github.com/costa92/go-protoc/v2/pkg/options"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
)

const (
	UserAgent = "apiserver"
)

type ServerOptions struct {
	GRPCOptions *genericoptions.GRPCOptions `json:"grpc" mapstructure:"grpc"`
	HTTPOptions *genericoptions.HTTPOptions `json:"http" mapstructure:"http"`
	TLSOptions  *genericoptions.TLSOptions  `json:"tls" mapstructure:"tls"`
	Log         *log.Options                `json:"log" mapstructure:"log"`
}

// Ensure ServerOptions implements the app.NamedFlagSetOptions interface.
var _ app.NamedFlagSetOptions = (*ServerOptions)(nil)

func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		GRPCOptions: genericoptions.NewGRPCOptions(),
		HTTPOptions: genericoptions.NewHTTPOptions(),
		TLSOptions:  genericoptions.NewTLSOptions(),
		Log:         log.NewOptions(),
	}
}

func (o *ServerOptions) Flags() (fss cliflag.NamedFlagSets) {
	o.GRPCOptions.AddFlags(fss.FlagSet("grpc"))
	o.HTTPOptions.AddFlags(fss.FlagSet("http"))
	o.Log.AddFlags(fss.FlagSet("log"))

	// fs := fss.FlagSet("misc")
	// client.AddFlags(fs)

	return fss

}

func (o *ServerOptions) Validate() error {
	errs := []error{}
	errs = append(errs, o.GRPCOptions.Validate()...)
	errs = append(errs, o.HTTPOptions.Validate()...)
	errs = append(errs, o.Log.Validate()...)

	return utilerrors.NewAggregate(errs)
}

func (o *ServerOptions) Complete() error {
	return nil
}

func (o *ServerOptions) Config() (*apiserver.Config, error) {
	return &apiserver.Config{
		GRPCOptions: o.GRPCOptions,
		HTTPOptions: o.HTTPOptions,
		TLSOptions:  o.TLSOptions,
	}, nil
}
