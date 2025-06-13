package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Option 是用于配置 App 的函数式选项。
type Option func(*App)

// RunFunc 定义了应用启动时要执行的函数的签名。
type RunFunc func() error

// CliOptions 定义了命令行选项接口
type CliOptions interface {
	Validate() error
	AddFlags(*cobra.Command)
}

// App 是一个命令行应用的核心结构。
type App struct {
	basename    string
	name        string
	description string
	options     CliOptions // 使用接口而非 any
	runFunc     RunFunc    // 应用启动时要执行的真正业务逻辑
	silence     bool
	commands    []*cobra.Command
	args        cobra.PositionalArgs
	cmd         *cobra.Command
}

// WithOptions 设置应用的命令行选项。
func WithOptions(opts CliOptions) Option {
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
		a.options.AddFlags(&cmd)
	}

	a.cmd = &cmd
}

// Run 启动应用。
func (a *App) Run() error {
	if err := a.cmd.Execute(); err != nil {
		return err
	}

	return nil
}

// run 是实际的应用运行函数。
func (a *App) run(cmd *cobra.Command, args []string) error {
	// 验证选项
	if a.options != nil {
		if err := a.options.Validate(); err != nil {
			return err
		}
	}

	// 执行业务逻辑
	if a.runFunc != nil {
		return a.runFunc()
	}

	return nil
}

var cfgFile string

// initConfig 初始化配置。
func (a *App) initConfig() error {
	if cfgFile != "" {
		// 使用命令行标志指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 在当前目录和用户主目录中查找配置文件
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		// 添加配置文件的搜索路径
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigName(fmt.Sprintf(".%s", a.basename))
		viper.SetConfigType("yaml")
	}

	// 读取环境变量
	viper.AutomaticEnv()
	viper.SetEnvPrefix(strings.ToUpper(a.basename))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// 如果找到配置文件，读取它
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 监听配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("配置文件已更改: %s\n", e.Name)
		// 验证新的配置
		if a.options != nil {
			if err := a.options.Validate(); err != nil {
				fmt.Printf("新配置验证失败: %v\n", err)
			}
		}
	})

	return nil
}
