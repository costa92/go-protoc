# 实施计划 - Makefile统一入口安装脚本重构

## 阶段1: 基础设施和平台支持

- [ ] 1. 创建标准化工具库目录结构
  - 创建 `scripts/installation/lib/` 目录
  - 设置正确的文件权限和执行位
  - 创建基础的 `common_lib.sh` 用于库函数加载
  - _需求: 6.1, 6.2, 6.3, 6.4_

- [ ] 2. 实现平台检测工具库
  - 编写 `scripts/installation/lib/platform.sh`
  - 实现 `proj::platform::detect_os()` 函数检测Ubuntu/macOS/不支持平台
  - 实现 `proj::platform::has_docker()` 函数检测Docker可用性
  - 实现 `proj::platform::get_package_manager()` 函数检测包管理器类型
  - 编写单元测试验证平台检测逻辑正确性
  - _需求: 5.1, 5.2_

- [ ] 3. 实现Docker标准化工具库
  - 编写 `scripts/installation/lib/docker_helper.sh`
  - 实现 `proj::docker::run_service()` 统一容器启动函数
  - 实现 `proj::docker::cleanup_container()` 容器清理函数
  - 实现 `proj::docker::check_container_status()` 容器状态检查函数
  - 添加标准化容器命名和网络配置逻辑
  - _需求: 2.1, 2.2, 2.3, 2.4_

- [ ] 4. 实现宿主机安装工具库
  - 编写 `scripts/installation/lib/native_helper.sh`
  - 实现 `proj::native::install_service()` 统一宿主机安装函数
  - 实现 `proj::native::start_service()` 和 `proj::native::stop_service()` 服务管理函数
  - 实现 `proj::native::setup_service_user()` 系统用户和目录创建函数
  - 添加平台路由逻辑调用对应适配器
  - _需求: 7.1, 7.2, 7.3_

- [ ] 5. 实现Ubuntu平台适配器
  - 编写 `scripts/installation/lib/ubuntu_adapter.sh`
  - 实现 `proj::ubuntu::install_package()` APT包安装函数
  - 实现 `proj::ubuntu::install_service()` Ubuntu服务安装函数
  - 实现systemd服务管理函数 (start/stop/status)
  - 实现 `proj::ubuntu::setup_user()` Ubuntu用户和目录管理
  - 实现 `proj::ubuntu::create_systemd_service()` systemd服务文件生成
  - _需求: 5.3, 5.4, 8.1, 8.2_

- [ ] 6. 实现macOS平台适配器
  - 编写 `scripts/installation/lib/macos_adapter.sh`
  - 实现 `proj::macos::install_package()` Homebrew包安装函数
  - 实现 `proj::macos::install_service()` macOS服务安装函数
  - 实现launchd服务管理函数 (start/stop/status)
  - 实现 `proj::macos::setup_user()` macOS用户目录管理
  - 实现 `proj::macos::create_launchd_service()` launchd服务文件生成
  - _需求: 5.3, 5.4, 8.1, 8.2_

- [ ] 7. 扩展版本管理系统
  - 修改 `scripts/installation/versions.sh` 添加平台配置变量
  - 添加配置目录变量 (PROJ_*_CONFIG_DIR)
  - 添加服务偏好配置 (INSTALL_PREFERENCES)
  - 添加平台特定的包名和公式映射
  - 验证所有现有变量保持向后兼容
  - _需求: 3.1, 3.2, 7.4_

- [ ] 8. 创建多平台模板目录结构
  - 创建 `scripts/installation/templates/` 主目录
  - 为每个服务创建 docker/ubuntu/macos 子目录
  - 创建通用模板文件 (systemd-template.service.tpl, launchd-template.plist.tpl)
  - 设置模板文件的正确权限
  - _需求: 3.3, 7.2_

## 阶段2: 双模式核心服务迁移

- [ ] 9. 实现配置管理器
  - 编写 `scripts/installation/lib/config_manager.sh`
  - 实现 `proj::config::render_service_template()` 多平台模板渲染函数
  - 实现 `proj::config::inject_service_vars()` 服务发现变量注入
  - 实现 `proj::config::ensure_config_dir()` 配置目录创建函数
  - 支持根据平台和安装方式选择正确的模板
  - _需求: 3.1, 3.2, 3.3, 3.4_

- [ ] 10. 迁移Redis服务支持双模式
  - 修改 `scripts/installation/redis.sh` 添加平台检测和路由逻辑
  - 实现 `proj::redis::install()` 智能安装方式选择函数
  - 实现 `proj::redis::docker::install()` Docker安装函数（保持兼容）
  - 实现 `proj::redis::native::install()` 宿主机安装函数
  - 实现 `proj::redis::ubuntu::install()` Ubuntu特定安装函数
  - 实现 `proj::redis::macos::install()` macOS特定安装函数
  - _需求: 1.1, 1.2, 1.3, 2.1, 2.2, 5.1, 5.2, 7.1_

- [ ] 11. 创建Redis多平台配置模板
  - 创建 `scripts/installation/templates/redis/docker/redis.conf.tpl`
  - 创建 `scripts/installation/templates/redis/ubuntu/redis.conf.tpl`
  - 创建 `scripts/installation/templates/redis/ubuntu/redis.service.tpl`
  - 创建 `scripts/installation/templates/redis/macos/redis.conf.tpl`
  - 创建 `scripts/installation/templates/redis/macos/com.proj.redis.plist.tpl`
  - 验证模板变量替换正确性
  - _需求: 3.3, 3.4_

- [ ] 12. 迁移OTEL Collector服务支持双模式
  - 修改 `scripts/installation/otelcol.sh` 添加多平台支持
  - 实现智能安装方式选择和平台特定安装函数
  - 处理OTEL Collector的复杂配置依赖和服务发现
  - 实现不同平台的二进制下载和安装逻辑
  - 验证所有平台下的配置文件生成正确性
  - _需求: 1.1, 1.2, 2.1, 2.2, 3.1, 3.3, 5.3_

- [ ] 13. 创建OTEL Collector多平台配置模板
  - 创建完整的OTEL Collector多平台模板集合
  - 实现服务发现变量注入 (VictoriaLogs端点等)
  - 处理不同平台下的路径和权限差异
  - 验证配置文件在不同环境下的正确性
  - _需求: 3.2, 3.3, 3.4, 5.3_

- [ ] 14. 更新Makefile路由支持新命令格式
  - 修改 `scripts/make-rules/deploy.mk` 添加新的命令模式支持
  - 实现 `deploy.install.native.%` 宿主机安装路由
  - 实现 `deploy.install.native.ubuntu.%` Ubuntu特定路由
  - 实现 `deploy.install.native.macos.%` macOS特定路由
  - 保持所有现有命令的完全向后兼容性
  - _需求: 1.1, 1.3, 1.4_

## 阶段3: 批量操作和跨平台测试

- [ ] 15. 实现健康检查系统
  - 编写 `scripts/installation/lib/health_check.sh`
  - 实现 `proj::health::check_service()` 通用健康检查函数
  - 实现 `proj::health::check_http()` HTTP端点检查函数
  - 实现 `proj::health::check_tcp()` TCP端口检查函数
  - 实现服务特定检查函数 (Redis ping, OTEL状态等)
  - 添加Docker容器和宿主机服务的健康检查支持
  - _需求: 4.1, 4.2, 4.3, 4.4_

- [ ] 16. 实现重试机制系统
  - 编写 `scripts/installation/lib/retry.sh`
  - 实现 `proj::retry::with_backoff()` 指数退避重试函数
  - 实现 `proj::retry::linear()` 线性重试函数
  - 集成重试机制到健康检查和服务启动流程
  - 添加可配置的重试参数和错误类型分类
  - _需求: 4.2, 4.3, 8.3, 8.4_

- [ ] 17. 实现服务组批量操作功能
  - 扩展服务组定义支持混合安装方式
  - 实现 `deploy.install.database` 批量数据库服务安装
  - 实现 `deploy.install.docker.database` 批量Docker安装
  - 实现 `deploy.install.native.database` 批量宿主机安装
  - 添加依赖关系检查和正确的启动顺序
  - _需求: 5.1, 5.2, 5.3, 5.4, 7.3_

- [ ] 18. 编写跨平台单元测试套件
  - 创建 `tests/unit/` 目录结构
  - 编写平台检测函数的单元测试 (bats框架)
  - 编写Docker Helper功能的单元测试
  - 编写配置管理器的单元测试
  - 编写Ubuntu和macOS适配器的单元测试
  - 实现测试环境的Mock和Stub机制
  - _需求: 8.1, 8.2, 8.3, 8.4_

- [ ] 19. 编写跨平台集成测试套件
  - 创建 `tests/integration/` 目录结构
  - 编写Redis服务完整生命周期测试 (Docker + 宿主机)
  - 编写OTEL Collector部署和配置测试
  - 编写批量操作的端到端测试
  - 编写平台切换和迁移测试
  - 创建CI/CD流水线配置支持多平台测试
  - _需求: 1.1, 1.2, 1.3, 2.1, 2.2, 4.1, 5.1_

- [ ] 20. 性能基准测试和优化
  - 创建不同安装方式的性能对比测试
  - 测量服务启动时间、健康检查响应时间等关键指标
  - 对比Docker vs 宿主机安装的资源使用情况
  - 优化脚本执行效率和并发处理能力
  - 生成性能基准报告和优化建议
  - _需求: 7.4_

## 阶段4: 文档、模板和最佳实践

- [ ] 21. 创建新服务添加模板和工具
  - 创建 `scripts/installation/template-service.sh` 标准服务脚本模板
  - 实现服务生成工具脚本自动创建新服务骨架
  - 创建多平台配置模板生成工具
  - 编写服务元数据配置指南
  - 验证使用模板添加新服务的完整流程
  - _需求: 7.1, 7.2, 7.3, 7.4_

- [ ] 22. 更新项目文档和使用指南
  - 更新 `CLAUDE.md` 中的服务管理命令说明
  - 编写平台选择和安装方式决策指南
  - 创建故障排除和常见问题解答文档
  - 编写服务迁移指南 (Docker ↔ 宿主机)
  - 更新命令参考和最佳实践文档
  - _需求: 1.4, 8.4_

- [ ] 23. 编写跨平台最佳实践指南
  - 创建不同平台下的性能和安全最佳实践文档
  - 编写服务配置优化指南
  - 创建混合环境部署策略文档
  - 编写平台特定的运维和监控指南
  - 包含实际案例和配置示例
  - _需求: 8.1, 8.2, 8.3_

- [ ] 24. 完整的向后兼容性验证
  - 创建兼容性测试套件验证所有现有命令正常工作
  - 测试现有脚本调用方式的兼容性
  - 验证现有配置变量和环境变量的兼容性
  - 测试从现有系统到新系统的平滑迁移
  - 创建回滚机制和应急预案
  - _需求: 1.3, 1.4_

- [ ] 25. 团队培训和知识转移
  - 准备团队培训材料和演示环境
  - 编写开发者快速上手指南
  - 创建常见操作的视频教程或演示脚本
  - 组织代码审查和知识分享会议
  - 建立新系统的运维支持和问题反馈机制
  - _需求: 7.4, 8.4_

这个实施计划将项目分解为25个具体的编程和测试任务，按4个阶段递进实施，确保每个阶段都有明确的交付物和验证标准。每个任务都直接对应到需求文档中的具体验收标准，保证实施的完整性和正确性。