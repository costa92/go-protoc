package options

import "github.com/spf13/pflag"

// CliOptions 为命令行应用程序提供了一个配置契约。
// 任何实现了此接口的结构体都可以被 app.App 用来构建应用。
type CliOptions interface {
	// AddFlags 向指定的 FlagSet 添加命令行标志。
	// 这使得每个模块都可以独立定义自己的标志。
	AddFlags(fs *pflag.FlagSet)

	// Validate 检查所有选项字段的合法性，并返回错误列表。
	Validate() []error
}
