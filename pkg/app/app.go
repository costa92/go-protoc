package app

import (
	"context"
	"os"
	"runtime"
	"strings"

	"github.com/costa92/go-protoc/v2/pkg/logger"
	genericoptions "github.com/costa92/go-protoc/v2/pkg/options"
	"github.com/costa92/go-protoc/v2/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "go.uber.org/automaxprocs"
	"k8s.io/component-base/cli"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"
)

// RunFunc defines the application's startup callback function.
type RunFunc func() error

// HealthCheckFunc defines the health check function for the application.
type HealthCheckFunc func() error

// Option defines optional parameters for initializing the application
// structure.
type Option func(*App)

type App struct {
	name        string
	shortDesc   string
	description string
	run         RunFunc
	cmd         *cobra.Command
	args        cobra.PositionalArgs

	// +optional
	healthCheckFunc HealthCheckFunc

	// +optional
	options any

	// +optional
	silence bool

	// +optional
	noConfig bool

	// watching and re-reading config files
	// +optional
	watch bool

	contextExtractors map[string]func(context.Context) string
}

// WithOptions to open the application's function to read from the command line
// or read parameters from the configuration file.
func WithOptions(opts any) Option {
	return func(app *App) {
		app.options = opts
	}
}

// WithRunFunc is used to set the application startup callback function option.
func WithRunFunc(run RunFunc) Option {
	return func(app *App) {
		app.run = run
	}
}

// WithDescription is used to set the description of the application.
func WithDescription(desc string) Option {
	return func(app *App) {
		app.description = desc
	}
}

// WithHealthCheckFunc is used to set the health check function for the application.
// The app framework will use the function to start a health check server.
func WithHealthCheckFunc(fn HealthCheckFunc) Option {
	return func(app *App) {
		app.healthCheckFunc = fn
	}
}

// WithDefaultHealthCheckFunc set the default health check function.
func WithDefaultHealthCheckFunc() Option {
	fn := func() HealthCheckFunc {
		return func() error {
			go genericoptions.NewHealthOptions().ServeHealthCheck()

			return nil
		}
	}

	return WithHealthCheckFunc(fn())
}

// WithSilence sets the application to silent mode, in which the program startup
// information, configuration information, and version information are not
// printed in the console.
func WithSilence() Option {
	return func(app *App) {
		app.silence = true
	}
}

// WithNoConfig set the application does not provide config flag.
func WithNoConfig() Option {
	return func(app *App) {
		app.noConfig = true
	}
}

// WithValidArgs set the validation function to valid non-flag arguments.
func WithValidArgs(args cobra.PositionalArgs) Option {
	return func(app *App) {
		app.args = args
	}
}

// WithDefaultValidArgs set default validation function to valid non-flag arguments.
func WithDefaultValidArgs() Option {
	return func(app *App) {
		app.args = cobra.NoArgs
	}
}

// WithWatchConfig watching and re-reading config files.
func WithWatchConfig() Option {
	return func(app *App) {
		app.watch = true
	}
}

func WithLoggerContextExtractor(contextExtractors map[string]func(context.Context) string) Option {
	return func(app *App) {
		app.contextExtractors = contextExtractors
	}
}

// NewApp creates a new application instance based on the given application name,
// binary name, and other options.
func NewApp(name string, shortDesc string, opts ...Option) *App {
	app := &App{
		name:      name,
		run:       func() error { return nil },
		shortDesc: shortDesc,
	}

	for _, o := range opts {
		o(app)
	}

	app.buildCommand()

	return app
}

// initializeEarlyLogger 初始化早期 logger，包含默认的服务信息
func initializeEarlyLogger() {
	// 创建包含默认服务信息的 logger 选项
	logOptions := logger.DefaultOptions()
	
	// 从版本包获取默认服务信息
	versionInfo := version.Get()
	logOptions.InitialFields = map[string]interface{}{
		"service": versionInfo.ServiceName,
		"version": versionInfo.GitVersion,
		"branch":  versionInfo.GitBranch,
	}

	// 初始化早期 logger - 只在没有默认logger时设置
	if logger.GetDefaultLogger() == nil {
		if earlyLogger, err := logger.NewLogger(logOptions); err == nil {
			logger.SetDefaultLogger(earlyLogger)
			// 测试早期 logger
			logger.Infow("Early logger initialized", "early_service", versionInfo.ServiceName)
		}
	} else {
		// 使用已存在的logger（通常是从配置文件初始化的）
		logger.Infow("Using existing logger from configuration", "early_service", versionInfo.ServiceName)
	}
}

// buildCommand is used to build a cobra command.
func (app *App) buildCommand() {
	cmd := &cobra.Command{
		Use:   formatBaseName(app.name),
		Short: app.shortDesc,
		Long:  app.description,
		RunE:  app.runCommand,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			// 在任何日志输出之前检查版本参数
			version.PrintAndExitIfRequested()
			// 现在可以安全地初始化日志
			initializeEarlyLogger()
			return nil
		},
		Args: app.args,
	}
	// When error printing is enabled for the Cobra command, a flag parse
	// error gets printed first, then optionally the often long usage
	// text. This is very unreadable in a console because the last few
	// lines that will be visible on screen don't include the error.
	//
	// The recommendation from #sig-cli was to print the usage text, then
	// the error. We implement this consistently for all commands here.
	// However, we don't want to print the usage text when command
	// execution fails for other reasons than parsing. We detect this via
	// the FlagParseError callback.
	//
	// Some commands, like kubectl, already deal with this themselves.
	// We don't change the behavior for those.
	if !cmd.SilenceUsage {
		cmd.SilenceUsage = true
		cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
			// Re-enable usage printing.
			c.SilenceUsage = false
			return err
		})
	}
	// In all cases error printing is done below.
	cmd.SilenceErrors = true

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.Flags().SortFlags = true

	var fs *pflag.FlagSet
	// 方法2：使用type switch
	switch typed := app.options.(type) {
	case NamedFlagSetOptions:
		var fss cliflag.NamedFlagSets
		fs = fss.FlagSet("global")

		if app.options != nil {
			fss = typed.Flags()
		}

		// 将配置标志添加到 fss 中的 "misc" 标志集中
		if !app.noConfig {
			AddConfigFlag(fss.FlagSet("misc"), app.name, app.watch)
		}

		for _, f := range fss.FlagSets {
			cmd.Flags().AddFlagSet(f)
		}

		cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
		cliflag.SetUsageAndHelpFunc(cmd, fss, cols)
	case FlagSetOptions:
		fs = cmd.PersistentFlags()
		if app.options != nil {
			typed.AddFlags(fs)
		}

		// 在 FlagSetOptions 分支中添加配置标志
		if !app.noConfig {
			AddConfigFlag(fs, app.name, app.watch)
		}
	default:
		// 在默认分支中添加配置标志
		if !app.noConfig {
			AddConfigFlag(cmd.Flags(), app.name, app.watch)
		}
	}

	version.AddFlags(fs)
	// Config flag has been added at the beginning of this method.

	app.cmd = cmd
}

// Run is used to launch the application.
func (app *App) Run() {
	os.Exit(cli.Run(app.cmd))
}

func (app *App) runCommand(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	if app.options != nil {
		if err := viper.Unmarshal(app.options); err != nil {
			return err
		}

		if complete, ok := app.options.(interface{ Complete() error }); ok {
			if err := complete.Complete(); err != nil {
				return err
			}
		}

		if validate, ok := app.options.(interface{ Validate() error }); ok {
			if err := validate.Validate(); err != nil {
				return err
			}
		}
	}

	app.initializeLogger()

	if !app.silence {
		logger.Infow("Starting application", "name", app.name, "version", version.Get().ToJSON())
		logger.Infow("Golang settings", "GOGC", os.Getenv("GOGC"), "GOMAXPROCS", os.Getenv("GOMAXPROCS"), "GOTRACEBACK", os.Getenv("GOTRACEBACK"))
		if !app.noConfig {
			PrintConfig()
		} else if app.options != nil {
			cliflag.PrintFlags(cmd.Flags())
		}
	}

	if app.healthCheckFunc != nil {
		if err := app.healthCheckFunc(); err != nil {
			return err
		}
	}

	// run application
	return app.run()
}

// Command returns cobra command instance inside the application.
func (app *App) Command() *cobra.Command {
	return app.cmd
}

// formatBaseName is formatted as an executable file name under different
// operating systems according to the given name.
func formatBaseName(name string) string {
	// Make case-insensitive and strip executable suffix if present
	if runtime.GOOS == "windows" {
		name = strings.ToLower(name)
		name = strings.TrimSuffix(name, ".exe")
	}
	return name
}

// initializeLogger sets up the logging system based on the configuration.
// 注意：大部分配置已在 config.go 的 reinitializeLoggerFromConfig() 中处理
func (app *App) initializeLogger() {
	// 这里可以添加应用层特定的日志初始化逻辑
	// 比如添加上下文提取器
	if app.contextExtractors != nil {
		logger.Infow("Logger context extractors configured", "extractors", len(app.contextExtractors))
	}
	
	// 打印最终的日志器状态确认
	versionInfo := version.Get()
	logger.Infow("Final logger configuration applied", 
		"service", versionInfo.ServiceName, 
		"version", versionInfo.GitVersion,
		"branch", versionInfo.GitBranch)
}
