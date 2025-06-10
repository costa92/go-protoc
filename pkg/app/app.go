package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/costa92/go-protoc/pkg/options"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Server 是服务器接口，定义了启动和停止方法
type Server interface {
	// Start 启动服务器并阻塞直到上下文取消
	Start(ctx context.Context) error
	// Stop 优雅地关闭服务器
	Stop(ctx context.Context) error
}

// Option 是用于配置 App 的函数式选项。
type Option func(*App)

// RunFunc 定义了应用启动时要执行的函数的签名。
type RunFunc func(basename string) error

// App 是一个命令行应用的核心结构。
type App struct {
	basename    string
	name        string
	description string
	options     options.CliOptions // 关键改动：依赖于接口而非具体实现
	runFunc     RunFunc            // 应用启动时要执行的真正业务逻辑
	silence     bool
	commands    []*cobra.Command
	args        cobra.PositionalArgs
	cmd         *cobra.Command
}

// WithOptions 设置应用的命令行选项。
func WithOptions(opts options.CliOptions) Option {
	return func(a *App) {
		a.options = opts
	}
}

// WithRunFunc 设置应用启动时执行的函数。
func WithRunFunc(run RunFunc) Option {
	return func(a *App) {
		a.runFunc = run
	}
}

// WithDescription 设置应用的描述。
func WithDescription(desc string) Option {
	return func(a *App) {
		a.description = desc
	}
}

// NewApp 创建一个新的 App 实例
// NewApp 创建一个新的 App 实例。
func NewApp(name string, basename string, opts ...Option) *App {
	a := &App{
		name:     name,
		basename: basename,
	}

	for _, o := range opts {
		o(a)
	}

	a.buildCommand()

	return a
}

func (a *App) buildCommand() {
	cmd := cobra.Command{
		Use:   a.basename,
		Short: a.name,
		Long:  a.description,
		// 在执行 RunE 之前，完成配置的加载和绑定
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return a.initConfig()
		},
		RunE: a.run,
	}
	cmd.SilenceUsage = true

	// 添加全局标志
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.apiserver.yaml or ./config.yaml or ./apiserver.yaml)")

	// 添加来自 Options 的标志
	if a.options != nil {
		a.options.AddFlags(cmd.PersistentFlags())
	}

	a.cmd = &cmd
}

// Run 启动应用。
func (a *App) Run() {
	if err := a.cmd.Execute(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func (a *App) run(cmd *cobra.Command, args []string) error {
	if a.options != nil {
		if errs := a.options.Validate(); len(errs) > 0 {
			return errs[0]
		}
	}

	if a.runFunc != nil {
		return a.runFunc(a.basename)
	}

	return nil
}

var cfgFile string

func (a *App) initConfig() error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("$HOME")
		viper.SetConfigName(a.basename)
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix(strings.ToUpper(a.basename))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// 将 viper 加载的配置反序列化到我们的 Options 结构体中
	if a.options != nil {
		if err := viper.Unmarshal(a.options); err != nil {
			return err
		}
	}

	// 监控配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		if err := viper.Unmarshal(a.options); err != nil {
			fmt.Printf("unmarshal config error: %s\n", err)
		}
	})

	return nil
}
