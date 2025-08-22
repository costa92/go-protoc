// Package version 提供了全面的版本信息管理系统，支持构建时版本注入、运行时版本查询以及多种输出格式。
//
// 核心特性：
//
//   - 构建时版本注入：通过 ldflags 在构建时注入版本信息
//   - 多维度信息：Git 版本、提交、分支、构建时间、运行环境等
//   - 多种输出格式：简化、JSON、表格格式
//   - 动态版本管理：运行时可以动态设置版本信息
//   - 命令行集成：内置的 --version 标志支持
//
// 基础使用：
//
//	info := version.Get()
//	fmt.Printf("Version: %s\n", info.String())           // 简化输出
//	fmt.Printf("JSON: %s\n", info.ToJSON())              // JSON 输出
//	fmt.Printf("Details:\n%s\n", info.Text())            // 详细表格
//
// 构建时版本注入：
//
//	go build -ldflags "
//	    -X 'github.com/costa92/go-protoc/v2/pkg/version.serviceName=myservice'
//	    -X 'github.com/costa92/go-protoc/v2/pkg/version.gitVersion=v1.0.0'
//	    -X 'github.com/costa92/go-protoc/v2/pkg/version.gitCommit=abc12345'
//	    -X 'github.com/costa92/go-protoc/v2/pkg/version.gitBranch=main'
//	    -X 'github.com/costa92/go-protoc/v2/pkg/version.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'
//	" ./cmd/myservice
//
// 动态版本管理：
//
//	// 运行时设置版本信息
//	if err := version.SetDynamicVersion("v1.2.3-hotfix.1"); err != nil {
//	    log.Fatal("Invalid version: ", err)
//	}
//
// 命令行集成：
//
//	version.AddFlags(pflag.CommandLine)
//	pflag.Parse()
//	version.PrintAndExitIfRequested()  // 处理 --version 标志
//
// 版本信息结构：
//
// Info 结构体包含了完整的版本和构建信息：
//   - GitVersion: Git 版本标签
//   - GitCommit: Git 提交 SHA
//   - GitTreeState: Git 仓库状态 (clean/dirty)
//   - GitBranch: Git 分支名
//   - BuildDate: ISO8601 格式构建时间
//   - ServiceName: 服务名称
//   - GoVersion: Go 运行时版本
//   - Compiler: Go 编译器
//   - Platform: 操作系统/架构
//
// 详细的使用指南和最佳实践请参考 README.md 文件。
package version // import "github.com/costa92/go-protoc/v2/pkg/version"
