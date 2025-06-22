# 代码风格和约定

## 项目约定
1. 使用 Protocol Buffers 定义 API
2. 遵循 Go 标准项目布局
   - cmd/: 主程序入口
   - internal/: 私有代码
   - pkg/: 可重用包

## 命名约定
1. 包名使用小写
2. 接口名通常以 -er 结尾
3. 错误变量以 Err 或 err 开头
4. 配置结构体以 Options 结尾

## 代码组织
1. 使用 Wire 进行依赖注入
2. 业务逻辑放在 internal/biz 目录
3. 数据访问层放在 internal/store 目录
4. HTTP/gRPC 处理器放在 internal/handler 目录

## 错误处理
1. 使用自定义的错误包 pkg/errors
2. 错误信息应该提供足够的上下文
3. 使用 wrap 方式传递错误

## 配置管理
1. 使用 Viper 处理配置文件
2. 配置文件使用 YAML 格式
3. 支持环境变量覆盖

## 日志规范
1. 使用结构化日志（Zap）
2. 包含必要的上下文信息
3. 合理使用日志级别

## 测试规范
1. 单元测试文件以 _test.go 结尾
2. 使用 testify 包进行断言
3. 模拟外部依赖进行测试