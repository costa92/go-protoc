# Prometheus 模块化配置系统

## 📋 概述

这是一个完全模块化的Prometheus配置管理系统，将复杂的配置逻辑拆分为独立的、可复用的模块。

## 🏗️ 目录结构

```
prometheus/
├── docker-run.sh.tpl              # 原始启动脚本（单体版本）
├── docker-run-modular.sh.tpl      # 新的模块化启动脚本
├── docker-status.sh.tpl           # 状态检查脚本
├── docker-stop.sh.tpl             # 停止脚本
├── prometheus.yml.tpl             # 原始配置模板
├── prometheus-base.yml.tpl        # 基础配置模板（模块化版本）
├── exporters/                     # 📦 Exporter管理模块
│   ├── common.sh.tpl                 # 通用exporter函数
│   ├── redis-exporter.sh.tpl         # Redis exporter逻辑
│   └── mysql-exporter.sh.tpl         # MySQL exporter逻辑
├── scrape_configs/                # 📊 抓取配置模块
│   ├── config-manager.sh.tpl         # 配置管理器
│   ├── application.yml.tpl           # 应用服务配置
│   ├── redis.yml.tpl                 # Redis监控配置
│   └── mysql.yml.tpl                 # MySQL监控配置
└── rule_files/                    # ⚠️ 告警规则模块
    ├── application-rules.yml.tpl     # 应用服务告警规则
    ├── redis-rules.yml.tpl           # Redis告警规则
    └── mysql-rules.yml.tpl           # MySQL告警规则
```

## 🚀 使用方法

### 方法1: 使用模块化脚本（推荐）

```bash
# 使用新的模块化脚本
/path/to/docker-run-modular.sh.tpl

# 或者创建一个新的make命令
make docker.prometheus.start.modular
```

### 方法2: 逐步迁移现有脚本

模块化组件可以逐步集成到现有的 `docker-run.sh.tpl` 中。

## 🎯 模块功能说明

### 1. Exporter管理模块 (`exporters/`)

**功能**: 自动检测服务并启动对应的exporter
- 自动检测运行中的服务（Redis、MySQL等）
- 自动启动对应的exporter容器
- 统一的健康检查和状态管理
- 支持多种数据库exporter

**核心函数**:
- `detect_and_start_all_exporters()` - 检测所有服务
- `redis_exporter_detect_and_start()` - Redis检测和启动
- `mysql_exporter_detect_and_start()` - MySQL检测和启动
- `check_all_exporters_health()` - 健康检查

### 2. 抓取配置模块 (`scrape_configs/`)

**功能**: 根据检测到的服务动态生成抓取配置
- 模块化的scrape配置文件
- 动态组合配置
- 标准化的标签和配置参数

**配置文件**:
- `application.yml.tpl` - Prometheus自监控 + API服务
- `redis.yml.tpl` - Redis指标抓取配置
- `mysql.yml.tpl` - MySQL指标抓取配置

### 3. 告警规则模块 (`rule_files/`)

**功能**: 提供开箱即用的告警规则
- 服务可用性监控
- 性能指标告警
- 资源使用率监控
- 分级告警（critical/warning/info）

**规则文件**:
- `application-rules.yml.tpl` - API服务 + Prometheus告警
- `redis-rules.yml.tpl` - Redis性能和可用性告警
- `mysql-rules.yml.tpl` - MySQL性能和连接告警

## ⚙️ 配置自定义

### 添加新的服务监控

1. **添加Exporter模块** (`exporters/new-service-exporter.sh.tpl`)
```bash
new_service_exporter_detect_and_start() {
    # 检测逻辑
    # 启动exporter
    # 设置环境变量
}
```

2. **添加抓取配置** (`scrape_configs/new-service.yml.tpl`)
```yaml
  - job_name: 'new-service'
    static_configs:
      - targets: ['${PROJ_PREFIX}-new-service-exporter:9999']
```

3. **添加告警规则** (`rule_files/new-service-rules.yml.tpl`)
```yaml
groups:
  - name: new-service.rules
    rules:
    - alert: NewServiceDown
      expr: new_service_up == 0
```

4. **更新通用模块**
在 `exporters/common.sh.tpl` 中添加:
```bash
# 加载新服务模块
source "${exporter_dir}/new-service-exporter.sh.tpl"

# 在detect_and_start_all_exporters()中添加
new_service_exporter_detect_and_start
```

### 自定义告警阈值

编辑对应的规则文件，修改表达式和阈值：
```yaml
# 例如：修改Redis内存告警阈值从80%到90%
- alert: RedisMemoryHigh
  expr: (redis_memory_used_bytes / redis_memory_max_bytes) * 100 > 90
```

## 🔧 环境变量

模块系统使用以下环境变量进行通信：

```bash
# 监控启用状态
ENABLE_REDIS_MONITORING=true/false
ENABLE_MYSQL_MONITORING=true/false

# 检测到的服务名称  
DETECTED_MYSQL_SERVICE=proj-mariadb

# 项目配置
PROJ_PREFIX=proj
PROJ_NETWORK_NAME=proj-network
PROJ_PROMETHEUS_CONFIG_DIR=/path/to/config
```

## 🔄 迁移指南

### 从单体脚本迁移到模块化

1. **备份现有配置**
```bash
cp docker-run.sh.tpl docker-run.sh.tpl.backup
```

2. **测试模块化脚本**
```bash
./docker-run-modular.sh.tpl
```

3. **验证功能**
- 检查 http://localhost:9090/targets
- 验证所有exporter正常运行
- 确认告警规则加载正确

4. **逐步迁移**
可以选择将模块函数逐步引入现有脚本，或完全切换到模块化版本。

## ✅ 优势

1. **可维护性**: 每个组件职责单一，易于维护和调试
2. **可扩展性**: 添加新服务监控只需要添加对应模块
3. **可复用性**: 模块可以在不同项目中复用
4. **一致性**: 标准化的配置模式和命名约定
5. **灵活性**: 支持条件性启用/禁用监控组件

## 🧪 测试

```bash
# 测试Redis监控
docker start proj-redis
./docker-run-modular.sh.tpl
curl http://localhost:9090/api/v1/targets | grep redis

# 测试MySQL监控
docker start proj-mariadb  
./docker-run-modular.sh.tpl
curl http://localhost:9090/api/v1/targets | grep mysql

# 测试告警规则
curl http://localhost:9090/api/v1/rules
```