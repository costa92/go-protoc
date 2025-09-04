# Docker服务数据卷删除标准化文档

本文档描述了所有Docker服务停止脚本的统一数据卷删除处理标准。

## 🎯 设计目标

1. **统一体验**: 所有服务使用相同的数据卷删除接口和提示
2. **多种触发方式**: 支持环境变量、命令行参数等多种方式
3. **安全第一**: 默认保留数据，删除需明确确认
4. **Make友好**: 支持通过Make命令的环境变量传递参数
5. **非交互支持**: 支持CI/CD等自动化环境

## 📋 标准化功能特性

### ✅ 支持的触发方式

#### 1. 环境变量方式（推荐）

```bash
# 交互式删除确认
REMOVE_DATA=true make docker.<service>.stop

# 强制删除（非交互式）
REMOVE_DATA=true FORCE_DELETE=true make docker.<service>.stop
```

#### 2. 命令行参数方式

```bash
# 直接调用生成的脚本
bash /path/to/_generated/docker-scripts/<service>/docker-stop.sh --remove-data
bash /path/to/_generated/docker-scripts/<service>/docker-stop.sh --force
```

### ✅ 交互模式处理

#### 交互式终端

- 显示详细的数据删除警告
- 列出具体会被删除的数据类型
- 要求用户明确输入 `y/N` 确认（默认N）
- 显示删除结果或保留状态

#### 非交互式环境

- 检测到非交互式环境时
- 需要额外的 `FORCE_DELETE=true` 环境变量才能执行删除
- 防止在CI/CD中意外删除数据

### ✅ 统一的用户界面

#### 基本停止（数据保留）

```
💡 提示: Redis数据卷已保留
   数据卷: proj-redis-data

   删除数据的方式：
     环境变量: REMOVE_DATA=true make docker.redis.stop
     直接调用: bash <脚本路径> --remove-data
     强制删除: REMOVE_DATA=true FORCE_DELETE=true make docker.redis.stop
```

#### 删除确认提示

```
⚠️  警告：将删除Redis数据卷！这将永久删除所有缓存数据！
包括：
  - 所有键值对数据
  - 持久化RDB快照
  - AOF操作日志
  - 缓存统计信息

确认删除数据卷？(y/N):
```

## 🔧 实现标准

### 代码模板结构

每个服务的 `docker-stop.sh.tpl` 文件都包含以下标准化部分：

```bash
# === 统一数据卷删除处理 ===
SERVICE_NAME="ServiceName"
DATA_VOLUMES="volume1,volume2"  # 逗号分隔的数据卷列表
DATA_DESCRIPTION="  - 数据类型1
  - 数据类型2
  - 数据类型3"

# 支持多种数据删除触发方式
should_remove_data=false

# 方式1: 环境变量 (推荐用于Make命令)
if [[ "${REMOVE_DATA:-}" == "true" ]]; then
    should_remove_data=true
fi

# 方式2: 命令行参数
for arg in "$@"; do
    case $arg in
        --remove-data|--force)
            should_remove_data=true
            break
            ;;
    esac
done

# 统一的删除逻辑...
```

### 必需变量定义

每个服务必须定义以下变量：

- `SERVICE_NAME`: 显示名称（如 "Redis", "Nacos", "Kafka"）
- `DATA_VOLUMES`: 数据卷列表，逗号分隔
- `DATA_DESCRIPTION`: 多行数据描述，每行以 "  - " 开头

## 📊 已标准化的服务列表

| 服务 | 状态 | 数据卷数量 | 特殊处理 |
|------|------|-----------|----------|
| Redis | ✅ 完成 | 1 | 无 |
| Nacos | ✅ 完成 | 2 | 无 |
| Kafka | ✅ 完成 | 2 | 包含Zookeeper依赖 |
| MariaDB | ✅ 完成 | 1 | 无 |
| MongoDB | ✅ 完成 | 3 | 多卷配置 |
| MySQL | ✅ 完成 | 1 | 无 |
| etcd | ✅ 完成 | 1 | 无 |
| VictoriaLogs | ✅ 完成 | 1 | 无 |
| OpenTelemetry Collector | ✅ 完成 | 1 | 无 |
| Zookeeper | ✅ 完成 | 2 | 无 |
| Prometheus | 🔄 需更新 | 1 | 使用不同参数模式 |

## 🛠️ 维护指南

### 添加新服务时

1. 在停止脚本末尾添加标准化数据卷删除处理部分
2. 定义服务特定的 `SERVICE_NAME`, `DATA_VOLUMES`, `DATA_DESCRIPTION`
3. 测试所有触发方式：基本停止、环境变量删除、强制删除
4. 更新本文档的服务列表

### 修改现有服务时

1. 保持标准化结构不变
2. 只修改服务特定的变量定义
3. 确保向后兼容性

## 🧪 测试验证

### 基本功能测试

```bash
# 1. 基本停止（应显示保留提示）
make docker.redis.stop

# 2. 环境变量删除（应显示删除确认）
REMOVE_DATA=true make docker.redis.stop

# 3. 强制删除（非交互式）
REMOVE_DATA=true FORCE_DELETE=true make docker.redis.stop
```

### 命令行测试

```bash
# 直接调用脚本测试
echo "N" | bash /path/to/docker-stop.sh --remove-data
echo "y" | bash /path/to/docker-stop.sh --remove-data
```

## 📈 优势总结

1. **用户体验一致**: 所有服务使用相同的交互模式
2. **Make集成**: 完美支持Make命令的环境变量传递
3. **安全性**: 多重确认机制，防止意外删除
4. **自动化友好**: 支持CI/CD等非交互式环境
5. **维护性**: 标准化代码结构，易于维护和扩展

---

> **注意**: 此标准化方案已在Redis服务上完整验证，其他服务正在逐步更新中。
