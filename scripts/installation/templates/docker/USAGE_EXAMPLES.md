# Docker服务数据卷删除使用示例

## 🚀 快速使用指南

### 1️⃣ 基本停止（保留数据）

```bash
make docker.redis.stop
make docker.nacos.stop
make docker.kafka.stop
```

**输出示例：**
```
Redis服务已完全停止

💡 提示: Redis数据卷已保留
   数据卷: proj-redis-data

   删除数据的方式：
     环境变量: REMOVE_DATA=true make docker.redis.stop
     直接调用: bash <脚本路径> --remove-data
     强制删除: REMOVE_DATA=true FORCE_DELETE=true make docker.redis.stop
```

### 2️⃣ 交互式数据删除

```bash
REMOVE_DATA=true make docker.redis.stop
```

**交互过程：**
```
⚠️  警告：将删除Redis数据卷！这将永久删除所有缓存数据！
包括：
  - 所有键值对数据
  - 持久化RDB快照
  - AOF操作日志
  - 缓存统计信息

确认删除数据卷？(y/N): █
```

**选择 'N' 或直接回车：**
```
数据卷保留: proj-redis-data
```

**选择 'y'：**
```
proj-redis-data
Redis数据卷已删除: proj-redis-data
所有Redis数据已被永久删除
```

### 3️⃣ 强制删除（自动化环境）

适用于CI/CD、脚本自动化等非交互环境：

```bash
REMOVE_DATA=true FORCE_DELETE=true make docker.redis.stop
```

**输出：**
```
⚠️  警告：将删除Redis数据卷！这将永久删除所有缓存数据！
包括：
  - 所有键值对数据
  - 持久化RDB快照
  - AOF操作日志
  - 缓存统计信息

Redis数据卷已强制删除: proj-redis-data
```

### 4️⃣ 命令行参数方式

```bash
# 交互确认
bash /path/to/_generated/docker-scripts/redis/docker-stop.sh --remove-data

# 或使用 --force 参数
bash /path/to/_generated/docker-scripts/redis/docker-stop.sh --force
```

## 🔍 服务特定示例

### Nacos (多数据卷)

```bash
# 基本停止
make docker.nacos.stop
```

**保留提示：**
```
💡 提示: Nacos数据卷已保留
   数据卷: proj-nacos-data
   数据卷: proj-nacos-logs

   删除数据的方式：
     环境变量: REMOVE_DATA=true make docker.nacos.stop
     直接调用: bash <脚本路径> --remove-data
```

**删除确认：**
```
⚠️  警告：将删除Nacos数据卷！这将永久删除所有配置和数据！
包括：
  - 服务注册信息
  - 配置管理数据
  - 命名空间配置
  - 用户和权限数据

确认删除数据卷？(y/N):
```

### Kafka (包含依赖)

Kafka停止时会自动处理Zookeeper依赖：

```bash
REMOVE_DATA=true make docker.kafka.stop
```

**处理过程：**
```
⚠️  警告：将删除Kafka数据卷！这将永久删除所有数据！
包括：
  - 所有Topic数据
  - 消息日志
  - 分区数据
  - 索引文件
  - 事务状态日志

确认删除数据卷？(y/N): y

Kafka数据卷已删除: proj-kafka-data, proj-kafka-logs
所有Kafka数据已被永久删除

🔄 正在停止Zookeeper依赖...
✅ Zookeeper容器已停止: proj-zookeeper
✅ Zookeeper容器已删除: proj-zookeeper
```

## 🔧 故障排除

### 非交互环境缺少FORCE_DELETE

**错误输出：**
```
⚠️  非交互式环境，需要设置 FORCE_DELETE=true 进行强制删除
数据卷保留: proj-redis-data
```

**解决方案：**
```bash
REMOVE_DATA=true FORCE_DELETE=true make docker.redis.stop
```

### 数据卷不存在

**输出：**
```
未找到要删除的数据卷
```

这是正常情况，说明数据卷已经被删除或从未创建。

### Make命令参数问题

❌ **错误方式：**
```bash
make docker.redis.stop --remove-data  # Make不支持此方式
```

✅ **正确方式：**
```bash
REMOVE_DATA=true make docker.redis.stop  # 使用环境变量
```

## 🛡️ 安全提醒

1. **默认安全**: 不带参数的停止命令永远不会删除数据
2. **明确确认**: 删除数据需要明确的环境变量或参数设置
3. **详细警告**: 每种服务都会详细列出将要删除的数据类型
4. **可恢复性**: 数据卷删除后无法恢复，请谨慎操作

## 📊 支持的服务列表

| 服务 | 基本停止 | 环境变量删除 | 强制删除 | 多数据卷 |
|------|---------|-------------|----------|----------|
| Redis | ✅ | ✅ | ✅ | - |
| Nacos | ✅ | ✅ | ✅ | ✅ |
| Kafka | ✅ | ✅ | ✅ | ✅ |
| MariaDB | ✅ | ✅ | ✅ | - |
| MongoDB | ✅ | ✅ | ✅ | ✅ |
| MySQL | ✅ | ✅ | ✅ | - |
| etcd | ✅ | ✅ | ✅ | - |
| VictoriaLogs | ✅ | ✅ | ✅ | - |
| OpenTelemetry Collector | ✅ | ✅ | ✅ | - |
| Zookeeper | ✅ | ✅ | ✅ | ✅ |

---

💡 **提示**: 建议在生产环境中使用前先在测试环境验证数据删除流程。