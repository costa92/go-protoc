package options

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/pflag"
)

var _ IOptions = (*SentryOptions)(nil)

// SentryOptions defines options for Sentry error reporting.
type SentryOptions struct {
	// DSN is the Sentry Data Source Name (connection string)
	DSN string `json:"dsn,omitempty" mapstructure:"dsn"`

	// Environment specifies the environment this application is running in (e.g., production, staging, development)
	Environment string `json:"environment,omitempty" mapstructure:"environment"`

	// Release identifies the version of your application
	Release string `json:"release,omitempty" mapstructure:"release"`

	// ServerName will be used as the server_name attribute
	ServerName string `json:"server-name,omitempty" mapstructure:"server-name"`

	// SampleRate configures the sample rate for error events, in the range [0.0, 1.0]
	SampleRate float64 `json:"sample-rate,omitempty" mapstructure:"sample-rate"`

	// TracesSampleRate configures the sample rate for performance monitoring (traces), in the range [0.0, 1.0]
	TracesSampleRate float64 `json:"traces-sample-rate,omitempty" mapstructure:"traces-sample-rate"`

	// ProfilesSampleRate configures the sample rate for profiling, in the range [0.0, 1.0]
	ProfilesSampleRate float64 `json:"profiles-sample-rate,omitempty" mapstructure:"profiles-sample-rate"`

	// Debug enables debug mode for detailed logging
	Debug bool `json:"debug,omitempty" mapstructure:"debug"`

	// AttachStacktrace configures whether to attach stack traces to pure capture message calls
	AttachStacktrace bool `json:"attach-stacktrace,omitempty" mapstructure:"attach-stacktrace"`

	// SendDefaultPII configures whether to send default PII (personally identifiable information)
	SendDefaultPII bool `json:"send-default-pii,omitempty" mapstructure:"send-default-pii"`

	// FlushTimeout is the timeout for flushing events before shutdown
	FlushTimeout time.Duration `json:"flush-timeout,omitempty" mapstructure:"flush-timeout"`

	// IgnoreErrors is a list of error messages to ignore
	IgnoreErrors []string `json:"ignore-errors,omitempty" mapstructure:"ignore-errors"`

	// BeforeSend is executed before sending an event to Sentry (not configurable via flags)
	BeforeSend interface{} `json:"-" mapstructure:"-"`

	// BeforeBreadcrumb is executed before adding a breadcrumb (not configurable via flags)
	BeforeBreadcrumb interface{} `json:"-" mapstructure:"-"`

	// Tags are key-value pairs that will be attached to every event
	Tags map[string]string `json:"tags,omitempty" mapstructure:"tags"`

	// MaxBreadcrumbs is the maximum number of breadcrumbs that should be captured
	MaxBreadcrumbs int `json:"max-breadcrumbs,omitempty" mapstructure:"max-breadcrumbs"`

	// EnableTracing enables or disables performance monitoring
	EnableTracing bool `json:"enable-tracing,omitempty" mapstructure:"enable-tracing"`

	// EnableProfiling enables or disables profiling
	EnableProfiling bool `json:"enable-profiling,omitempty" mapstructure:"enable-profiling"`

	// Enabled controls whether Sentry is enabled or disabled
	Enabled bool `json:"enabled,omitempty" mapstructure:"enabled"`
}

// NewSentryOptions creates a new SentryOptions instance with default values.
func NewSentryOptions() *SentryOptions {
	return &SentryOptions{
		Environment:        "development",
		SampleRate:         1.0,
		TracesSampleRate:   0.1, // 10% for performance monitoring
		ProfilesSampleRate: 0.1, // 10% for profiling
		Debug:              false,
		AttachStacktrace:   true,
		SendDefaultPII:     false,
		FlushTimeout:       5 * time.Second,
		IgnoreErrors:       []string{},
		Tags:               make(map[string]string),
		MaxBreadcrumbs:     100,
		EnableTracing:      true,
		EnableProfiling:    false, // Disabled by default due to performance impact
		Enabled:            false, // Disabled by default for security
	}
}

// Validate verifies the Sentry configuration.
func (o *SentryOptions) Validate() []error {
	var errs []error

	// If Sentry is not enabled, skip validation
	if !o.Enabled {
		return errs
	}

	// DSN is required when enabled
	if o.DSN == "" {
		errs = append(errs, fmt.Errorf("sentry DSN is required when enabled"))
	} else {
		// Validate DSN format
		if _, err := url.Parse(o.DSN); err != nil {
			errs = append(errs, fmt.Errorf("invalid sentry DSN format: %w", err))
		}
	}

	// Validate sample rates are in valid range [0.0, 1.0]
	if o.SampleRate < 0.0 || o.SampleRate > 1.0 {
		errs = append(errs, fmt.Errorf("sentry sample rate must be between 0.0 and 1.0, got %f", o.SampleRate))
	}

	if o.TracesSampleRate < 0.0 || o.TracesSampleRate > 1.0 {
		errs = append(errs, fmt.Errorf("sentry traces sample rate must be between 0.0 and 1.0, got %f", o.TracesSampleRate))
	}

	if o.ProfilesSampleRate < 0.0 || o.ProfilesSampleRate > 1.0 {
		errs = append(errs, fmt.Errorf("sentry profiles sample rate must be between 0.0 and 1.0, got %f", o.ProfilesSampleRate))
	}

	// Validate environment is not empty
	if o.Environment == "" {
		errs = append(errs, fmt.Errorf("sentry environment cannot be empty"))
	}

	// Validate MaxBreadcrumbs is positive
	if o.MaxBreadcrumbs < 0 {
		errs = append(errs, fmt.Errorf("sentry max breadcrumbs must be non-negative, got %d", o.MaxBreadcrumbs))
	}

	// Validate FlushTimeout is positive
	if o.FlushTimeout < 0 {
		errs = append(errs, fmt.Errorf("sentry flush timeout must be non-negative, got %v", o.FlushTimeout))
	}

	return errs
}

// AddFlags adds Sentry-related command line flags.
func (o *SentryOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.BoolVar(&o.Enabled, "sentry.enabled", o.Enabled,
		"Enable Sentry error reporting and performance monitoring.")

	fs.StringVar(&o.DSN, "sentry.dsn", o.DSN,
		"Sentry Data Source Name (DSN) for error reporting. Required when Sentry is enabled.")

	fs.StringVar(&o.Environment, "sentry.environment", o.Environment,
		"Environment name for Sentry (e.g., development, staging, production).")

	fs.StringVar(&o.Release, "sentry.release", o.Release,
		"Release version for Sentry. Helps track errors across deployments.")

	fs.StringVar(&o.ServerName, "sentry.server-name", o.ServerName,
		"Server name that will be used as the server_name attribute in Sentry.")

	fs.Float64Var(&o.SampleRate, "sentry.sample-rate", o.SampleRate,
		"Sample rate for error events (0.0 to 1.0). 1.0 captures all errors.")

	fs.Float64Var(&o.TracesSampleRate, "sentry.traces-sample-rate", o.TracesSampleRate,
		"Sample rate for performance monitoring traces (0.0 to 1.0).")

	fs.Float64Var(&o.ProfilesSampleRate, "sentry.profiles-sample-rate", o.ProfilesSampleRate,
		"Sample rate for profiling (0.0 to 1.0). Note: Profiling has performance impact.")

	fs.BoolVar(&o.Debug, "sentry.debug", o.Debug,
		"Enable Sentry debug mode for detailed logging.")

	fs.BoolVar(&o.AttachStacktrace, "sentry.attach-stacktrace", o.AttachStacktrace,
		"Attach stack traces to pure capture message calls.")

	fs.BoolVar(&o.SendDefaultPII, "sentry.send-default-pii", o.SendDefaultPII,
		"Send default PII (personally identifiable information). Use with caution.")

	fs.DurationVar(&o.FlushTimeout, "sentry.flush-timeout", o.FlushTimeout,
		"Timeout for flushing events before shutdown.")

	fs.StringSliceVar(&o.IgnoreErrors, "sentry.ignore-errors", o.IgnoreErrors,
		"List of error messages to ignore (substring match).")

	fs.IntVar(&o.MaxBreadcrumbs, "sentry.max-breadcrumbs", o.MaxBreadcrumbs,
		"Maximum number of breadcrumbs to capture.")

	fs.BoolVar(&o.EnableTracing, "sentry.enable-tracing", o.EnableTracing,
		"Enable Sentry performance monitoring (tracing).")

	fs.BoolVar(&o.EnableProfiling, "sentry.enable-profiling", o.EnableProfiling,
		"Enable Sentry profiling. Note: This has performance impact.")
}

// IsSentryEnabled returns whether Sentry is enabled and properly configured.
func (o *SentryOptions) IsSentryEnabled() bool {
	return o.Enabled && o.DSN != ""
}

// IsTracingEnabled returns whether performance monitoring is enabled.
func (o *SentryOptions) IsTracingEnabled() bool {
	return o.IsSentryEnabled() && o.EnableTracing && o.TracesSampleRate > 0
}

// IsProfilingEnabled returns whether profiling is enabled.
func (o *SentryOptions) IsProfilingEnabled() bool {
	return o.IsSentryEnabled() && o.EnableProfiling && o.ProfilesSampleRate > 0
}

// GetTagsSlice returns tags as a slice of key=value strings for command line usage.
func (o *SentryOptions) GetTagsSlice() []string {
	var tags []string
	for key, value := range o.Tags {
		tags = append(tags, fmt.Sprintf("%s=%s", key, value))
	}
	return tags
}

// SetTagsFromSlice sets tags from a slice of key=value strings.
func (o *SentryOptions) SetTagsFromSlice(tags []string) error {
	if o.Tags == nil {
		o.Tags = make(map[string]string)
	}

	for _, tag := range tags {
		// Split on first '=' to handle values that contain '='
		parts := []string{}
		if equalIndex := fmt.Sprintf("%s", tag); len(equalIndex) > 0 {
			for i, char := range tag {
				if char == '=' {
					parts = []string{tag[:i], tag[i+1:]}
					break
				}
			}
		}

		if len(parts) != 2 {
			return fmt.Errorf("invalid tag format '%s', expected 'key=value'", tag)
		}

		key, value := parts[0], parts[1]
		if key == "" {
			return fmt.Errorf("tag key cannot be empty in '%s'", tag)
		}

		o.Tags[key] = value
	}

	return nil
}
