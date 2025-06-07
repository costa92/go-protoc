# Go-Protoc 开发指南

## 环境要求

### 基础环境

- Go 1.20+
- Protocol Buffers 编译器 (protoc)
- Make
- Git

### 开发工具

- 推荐使用 VS Code 或 GoLand
- Go 语言插件
- Protocol Buffers 插件
- Git 工具

## 开发环境设置

1. 安装 Go

```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.20.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.20.linux-amd64.tar.gz
```

2. 安装 Protocol Buffers

```bash
# macOS
brew install protobuf

# Linux
sudo apt-get install protobuf-compiler
```

3. 克隆项目

```bash
git clone https://github.com/costa92/go-protoc.git
cd go-protoc
```

4. 安装依赖

```bash
go mod download
```

## 开发流程

### 1. 分支管理

- main: 主分支，用于发布
- develop: 开发分支，用于集成功能
- feature/*: 功能分支，用于开发新功能
- bugfix/*: 修复分支，用于修复问题
- release/*: 发布分支，用于版本发布

### 2. 开发步骤

1. 从 develop 分支创建功能分支

```bash
git checkout develop
git checkout -b feature/your-feature
```

2. 开发功能
3. 提交代码
4. 创建合并请求
5. 代码审查
6. 合并到 develop 分支

## 代码规范

### 1. 目录结构规范

- cmd/: 存放主要的应用程序入口
- pkg/: 存放可以被外部应用程序使用的代码
- internal/: 存放私有应用程序和库代码
- configs/: 存放配置文件模板或默认配置
- docs/: 存放设计和用户文档
- test/: 存放额外的外部测试应用程序和测试数据

### 2. 命名规范

- 包名：使用小写单词，不使用下划线或混合大小写
- 文件名：使用小写，可以使用下划线分隔单词
- 函数名：使用驼峰命名法
- 常量名：使用全大写，使用下划线分隔单词
- 变量名：使用驼峰命名法

### 3. 代码风格

- 使用 gofmt 格式化代码
- 遵循 Go 官方代码规范
- 添加适当的注释
- 错误处理要明确
- 避免过长的函数

## 测试规范

### 1. 单元测试

- 测试文件命名：*_test.go
- 测试函数命名：Test[被测试函数名]
- 使用 Go 标准测试框架
- 保持测试简单明了
- 测试覆盖率要求：>80%

### 2. 基准测试

- 基准测试函数命名：Benchmark[被测试函数名]
- 关注性能指标
- 与历史数据比较

### 3. 集成测试

- 测试完整功能流程
- 模拟真实环境
- 测试异常情况

## 版本控制

### 1. 提交信息规范

格式：

```
<type>(<scope>): <subject>

<body>

<footer>
```

类型（type）：

- feat: 新功能
- fix: 修复
- docs: 文档
- style: 格式
- refactor: 重构
- test: 测试
- chore: 构建过程或辅助工具的变动

### 2. 语义化版本

版本格式：主版本号.次版本号.修订号

- 主版本号：做了不兼容的 API 修改
- 次版本号：做了向下兼容的功能性新增
- 修订号：做了向下兼容的问题修正

## CI/CD 流程

### 1. 持续集成

- 代码提交触发构建
- 运行单元测试
- 运行代码检查
- 生成测试报告

### 2. 持续部署

- 自动化构建
- 自动化测试
- 自动化部署
- 环境隔离

## 常见问题

### 1. 编译问题

- 检查 Go 版本
- 检查 protoc 版本
- 检查依赖是否完整
- 检查环境变量设置

### 2. 依赖问题

- 使用 go mod tidy 清理依赖
- 检查 go.mod 文件
- 确保依赖版本兼容
- 使用 go mod vendor（如需要）

### 3. 测试问题

- 检查测试环境
- 查看测试日志
- 使用调试工具
- 隔离测试用例
