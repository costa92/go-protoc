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

// App 是一个命令行应用的核心结构。
type App struct {
	basename    string
	name        string
	description string
	options     any // 关键改动：依赖于接口而非具体实现
	runFunc     RunFunc            // 应用启动时要执行的真正业务逻辑
	silence     bool
	commands    []*cobra.Command
	args        cobra.PositionalArgs
	cmd         *cobra.Command
}

// WithOptions 设置应用的命令行选项。
func WithOptions(opts any) Option {
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
		fmt.Printf("使用指定的配置文件: %s\n", cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("$HOME")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		fmt.Printf("搜索配置文件路径: ., ./configs, $HOME\n")
	}

	viper.SetEnvPrefix(strings.ToUpper(a.basename))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取配置文件失败: %v", err)
		}
		fmt.Printf("未找到配置文件，将使用默认值\n")
	} else {
		fmt.Printf("成功加载配置文件: %s\n", viper.ConfigFileUsed())
	}

	// 将 viper 加载的配置反序列化到我们的 Options 结构体中
	if a.options != nil {
		if err := viper.Unmarshal(a.options); err != nil {
			return fmt.Errorf("解析配置文件失败: %v", err)
		}
		fmt.Printf("配置加载完成，开始验证配置...\n")

		// 验证配置
		if errs := a.options.Validate(); len(errs) > 0 {
			for _, err := range errs {
				fmt.Printf("配置验证错误: %v\n", err)
			}
			if len(errs) > 0 {
				return fmt.Errorf("配置验证失败")
			}
		}
		fmt.Printf("配置验证通过\n")
	}

	// 监控配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("配置文件发生变化: %s\n", e.Name)
		if err := viper.Unmarshal(a.options); err != nil {
			fmt.Printf("重新加载配置失败: %v\n", err)
		} else {
			fmt.Printf("配置已重新加载\n")
		}
	})

	return nil
}
