# 项目文档

本项目基于 Kratos v2 构建的生产就绪 Go 微服务框架的完整文档。

## 📚 文档导航

### 🏗️ [架构设计](./architecture/)
深入了解系统的整体设计和核心组件架构。

- **[系统整体架构](./architecture/system-overview.md)** - 项目的整体架构设计和组件关系
- **[可观测性架构](./architecture/observability.md)** - OpenTelemetry 统一可观测性平台设计  
- **[日志系统架构](./architecture/logging-system.md)** - 统一日志收集、处理和分析系统
- **[错误处理架构](./architecture/error-handling.md)** - 统一错误处理和国际化支持

### 📋 [使用指南](./guides/)
开发、部署和运维的实用指南。

#### 开发指南
- **[开发入门](./guides/development/getting-started.md)** - 快速开始开发指南
- **[Makefile 使用](./guides/development/makefile-usage.md)** - 构建和开发工具使用
- **[数据验证使用](./guides/development/validation-usage.md)** - 数据验证框架使用指南

#### 部署指南  
- **[OTEL 部署](./guides/deployment/otel-installation.md)** - OpenTelemetry 组件安装部署
- **[日志系统部署](./guides/deployment/logging-setup.md)** - 日志收集系统部署配置

#### 故障排除
- **[日志问题排查](./guides/troubleshooting/logging-issues.md)** - 日志系统常见问题和解决方案
- **[OTEL 问题修复](./guides/troubleshooting/otel-fixes.md)** - OpenTelemetry 相关问题修复
- **[连接池监控](./guides/troubleshooting/connection-pool.md)** - 数据库连接池监控和调优

### 📖 [参考文档](./reference/)
API 文档、规范和技术参考。

#### API 文档
- **[错误码参考](./reference/api/errors-code/)** - 完整的错误码定义和说明
- **[生成的 API 文档](./reference/api/generated/)** - 自动生成的 API 接口文档

#### 规范文档
- **[错误命名规范](./reference/specifications/error-conventions.md)** - 错误处理命名约定
- **[版本同步规范](./reference/specifications/version-sync.md)** - 版本管理和同步规范

## 🚀 快速开始

### 新手入门路径

1. **了解架构** → [系统整体架构](./architecture/system-overview.md)
2. **开发环境** → [开发入门](./guides/development/getting-started.md)  
3. **构建运行** → [Makefile 使用](./guides/development/makefile-usage.md)
4. **部署上线** → [部署指南](./guides/deployment/)

### 按角色推荐

#### 🔨 开发人员
- [开发入门指南](./guides/development/getting-started.md)
- [错误处理架构](./architecture/error-handling.md)
- [数据验证使用](./guides/development/validation-usage.md)
- [API 错误码参考](./reference/api/errors-code/)

#### 🔧 运维人员  
- [OTEL 部署指南](./guides/deployment/otel-installation.md)
- [日志系统部署](./guides/deployment/logging-setup.md)
- [故障排除指南](./guides/troubleshooting/)
- [连接池监控](./guides/troubleshooting/connection-pool.md)

#### 🏗️ 架构师
- [系统整体架构](./architecture/system-overview.md)
- [可观测性架构](./architecture/observability.md)
- [日志系统架构](./architecture/logging-system.md)

## 📝 文档维护

### 文档结构原则

- **architecture/** - 设计决策和系统架构，相对稳定
- **guides/** - 操作指南和最佳实践，经常更新
- **reference/** - 技术参考和规范，版本化管理

### 贡献指南

1. **新增文档** - 按照目录结构放入合适的分类
2. **更新现有文档** - 确保信息的准确性和时效性  
3. **文档链接** - 使用相对路径，便于维护
4. **索引维护** - 新增文档后更新相关的 README.md

### 文档规范

- 使用 Markdown 格式
- 遵循 [MarkdownLint](https://github.com/DavidAnson/markdownlint) 规范
- 中英文混排使用合适的空格
- 代码块指定语言类型

## 🔗 相关链接

- [项目主页](../README.md)
- [示例代码](../examples/)
- [脚本工具](../scripts/)
- [配置文件](../configs/)

---

> 💡 **提示**: 如果您找不到需要的文档或发现文档有问题，请提交 Issue 或 Pull Request。