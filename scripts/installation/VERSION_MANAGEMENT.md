# 版本管理说明

## 统一版本管理实现

本项目已实现统一的版本管理机制，所有第三方组件的版本都在 `scripts/installation/versions.sh` 中集中管理。

## 架构变更

### 之前的版本管理
- 版本信息分散在各个安装脚本中
- 部分版本在 `manifests/env/env.dev` 中定义
- 存在版本不一致和重复定义问题

### 现在的版本管理
- **统一配置文件**: `scripts/installation/versions.sh`
- **自动加载机制**: 通过 `common.sh` 自动加载到所有安装脚本
- **向后兼容**: 保持原有环境变量名称不变

## 版本配置结构

```bash
# 基础设施组件版本
export REDIS_VERSION=7.2.4
export MARIADB_VERSION=11.2.2
export MONGODB_VERSION=7.0.5
export ETCD_VERSION=v3.5.12
export KAFKA_VERSION=3.6.1

# 可观测性栈版本
export JAEGER_VERSION=1.52.0
export PROMETHEUS_VERSION=2.48.1
export GRAFANA_VERSION=10.2.4
export ALERTMANAGER_VERSION=0.26.0
export OTELCOL_VERSION=0.91.0

# 日志管理版本
export VICTORIALOGS_VERSION=v0.5.2-victorialogs

# 工具版本
export DOCKER_COMPOSE_VERSION=v2.29.7
```

## 使用方式

### 查看所有版本
```bash
./scripts/installation/versions.sh show
```

### 验证版本格式
```bash
./scripts/installation/versions.sh validate
```

### 在脚本中使用
```bash
# 自动通过 common.sh 加载
source scripts/installation/common.sh
echo "使用 Prometheus 版本: $PROJ_PROMETHEUS_VERSION"
```

## 版本升级步骤

1. **修改版本文件**
   ```bash
   # 编辑 scripts/installation/versions.sh
   export PROMETHEUS_VERSION=2.50.0  # 更新到新版本
   ```

2. **验证版本格式**
   ```bash
   ./scripts/installation/versions.sh validate
   ```

3. **查看更新结果**
   ```bash
   ./scripts/installation/versions.sh show
   ```

4. **测试相关安装脚本**
   ```bash
   # 测试 Prometheus 安装脚本是否正常
   source scripts/installation/prometheus.sh
   echo "Prometheus版本: $PROJ_PROMETHEUS_VERSION"
   ```

## 已更新的安装脚本

以下脚本已更新为使用统一版本管理：

- ✅ `prometheus.sh` - Prometheus 监控
- ✅ `grafana.sh` - Grafana 仪表板
- ✅ `alertmanager.sh` - AlertManager 告警
- ✅ `jaeger.sh` - Jaeger 链路追踪
- ✅ `etcd.sh` - etcd 分布式存储
- ✅ `otelcol.sh` - OpenTelemetry Collector
- ✅ `victorialogs.sh` - VictoriaLogs 日志
- ✅ `docker-compose.sh` - Docker Compose

## 版本兼容性

为了向后兼容，保持了原有的环境变量名称：

```bash
# 新版本变量 -> 原版本变量
PROMETHEUS_VERSION -> PROJ_PROMETHEUS_VERSION
GRAFANA_VERSION -> PROJ_GRAFANA_VERSION
# ... 等等
```

## 版本管理最佳实践

1. **统一升级**: 升级组件版本时，只需修改 `versions.sh` 文件
2. **批量测试**: 使用 `versions.sh validate` 验证所有版本格式
3. **版本跟踪**: 在 Git 提交中清楚标注版本变更
4. **兼容性测试**: 升级后测试相关组件的兼容性

## 故障排除

### 版本未加载
```bash
# 确保 common.sh 正确加载版本文件
source scripts/installation/common.sh
echo $PROJ_PROMETHEUS_VERSION
```

### 版本格式错误
```bash
# 使用验证功能检查
./scripts/installation/versions.sh validate
```

### 环境变量冲突
```bash
# 检查是否有其他地方设置了相同变量
env | grep VERSION
```

## 维护说明

- 定期检查组件官方版本更新
- 升级前测试兼容性
- 保持版本文档同步更新
- 遵循语义化版本规范