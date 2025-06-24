package options

import (
	"github.com/costa92/go-protoc/v2/internal/apiserver"
	"github.com/costa92/go-protoc/v2/pkg/app"
	"github.com/costa92/go-protoc/v2/pkg/log"
	genericoptions "github.com/costa92/go-protoc/v2/pkg/options"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/util/feature"
	cliflag "k8s.io/component-base/cli/flag"
)

const (
	UserAgent = "apiserver"
)

type ServerOptions struct {
	GRPCOptions  *genericoptions.GRPCOptions  `json:"grpc" mapstructure:"grpc"`
	HTTPOptions  *genericoptions.HTTPOptions  `json:"http" mapstructure:"http"`
	MySQLOptions *genericoptions.MySQLOptions `json:"mysql" mapstructure:"mysql"`
	TLSOptions   *genericoptions.TLSOptions   `json:"tls" mapstructure:"tls"`
	JWTOptions   *genericoptions.JWTOptions   `json:"jwt" mapstructure:"jwt"` // Added JWT Options
	Log          *log.Options                 `json:"log" mapstructure:"log"`
	FeatureGates map[string]bool              `json:"feature-gates"`
}

// Ensure ServerOptions implements the app.NamedFlagSetOptions interface.
var _ app.NamedFlagSetOptions = (*ServerOptions)(nil)

func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		GRPCOptions:  genericoptions.NewGRPCOptions(),
		HTTPOptions:  genericoptions.NewHTTPOptions(),
		TLSOptions:   genericoptions.NewTLSOptions(),
		MySQLOptions: genericoptions.NewMySQLOptions(),
		JWTOptions:   genericoptions.NewJWTOptions(), // Initialize JWT Options
		Log:          log.NewOptions(),
	}
}

func (o *ServerOptions) Flags() (fss cliflag.NamedFlagSets) {
	o.GRPCOptions.AddFlags(fss.FlagSet("grpc"))
	o.HTTPOptions.AddFlags(fss.FlagSet("http"))
	o.MySQLOptions.AddFlags(fss.FlagSet("mysql"))
	o.JWTOptions.AddFlags(fss.FlagSet("jwt")) // Add JWT flags
	o.Log.AddFlags(fss.FlagSet("log"))

	fs := fss.FlagSet("misc")
	// client.AddFlags(fs)
	feature.DefaultMutableFeatureGate.AddFlag(fs)
	return fss

}

func (o *ServerOptions) Validate() error {
	errs := []error{}
	errs = append(errs, o.GRPCOptions.Validate()...)
	errs = append(errs, o.HTTPOptions.Validate()...)
	errs = append(errs, o.MySQLOptions.Validate()...)
	errs = append(errs, o.JWTOptions.Validate()...) // Validate JWT Options
	errs = append(errs, o.Log.Validate()...)

	return utilerrors.NewAggregate(errs)
}

func (o *ServerOptions) Complete() error {
	_ = feature.DefaultMutableFeatureGate.SetFromMap(o.FeatureGates)
	return nil
}

func (o *ServerOptions) Config() (*apiserver.Config, error) {
	if err := o.Validate(); err != nil { // It's good practice to validate before returning config
		return nil, err
	}
	return &apiserver.Config{
		GRPCOptions:  o.GRPCOptions,
		HTTPOptions:  o.HTTPOptions,
		TLSOptions:   o.TLSOptions,
		MySQLOptions: o.MySQLOptions,
		JWTOptions:   o.JWTOptions, // Pass JWT Options to apiserver.Config
	}, nil
}
