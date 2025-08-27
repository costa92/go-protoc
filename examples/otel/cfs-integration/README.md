# OTEL 日志写入腾讯云 CFS 集成示例

这个示例展示如何配置 OpenTelemetry Collector 将日志直接写入腾讯云云文件存储 (CFS)，实现持久化的日志存储和管理。

## 🌟 功能特点

- **直接写入CFS**: 日志直接写入腾讯云CFS，无需中间存储
- **自动分区**: 按日期和服务名自动组织CFS目录结构
- **NFS v4.0**: 使用NFS v4.0协议确保高性能和稳定性
- **容错处理**: 本地缓存机制应对网络中断
- **元数据增强**: 自动添加CFS相关元数据信息

## 📋 前置要求

### 腾讯云CFS配置

1. **创建CFS文件系统**

   ```bash
   # 在腾讯云控制台或通过API创建CFS文件系统
   # 获取文件系统ID和挂载点IP
   ```

2. **获取必要信息**
   - CFS文件系统ID: `cfs-xxxxxxxx`
   - 挂载点IP地址: `10.x.x.x`
   - 地域信息: `ap-beijing`、`ap-shanghai` 等

### 环境准备

```bash
# 设置必要的环境变量
export CFS_FILESYSTEM_ID="cfs-xxxxxxxx"    # CFS文件系统ID
export CFS_IP="10.0.0.100"                 # CFS挂载点IP
export TENCENT_REGION="ap-beijing"          # 腾讯云地域
export CFS_MOUNT_POINT="/mnt/cfs-logs"     # 本地挂载点
```

## 🚀 快速开始

### 1. 安装CFS集成

```bash
# 运行CFS集成安装脚本
./scripts/installation/otel-cfs-setup.sh install
```

这个命令会自动执行：

- 验证CFS配置参数
- 安装NFS客户端工具
- 挂载CFS文件系统
- 配置自动挂载（fstab）
- 生成CFS专用的OTEL配置
- 启动OTEL Collector容器

### 2. 验证安装

```bash
# 检查CFS集成状态
./scripts/installation/otel-cfs-setup.sh status

# 查看详细信息
./scripts/installation/otel-cfs-setup.sh info
```

### 3. 测试日志写入

```bash
# 测试日志写入CFS功能
./scripts/installation/otel-cfs-setup.sh test
```

## 📊 CFS目录结构

日志在CFS中的组织结构：

```
/mnt/cfs-logs/
├── logs/                           # 按服务组织的日志
│   └── apiserver/                  # 服务名目录
│       └── 2024/01/15/            # 按日期分区
│           └── app.jsonl          # JSON Lines格式日志
└── otel-logs/                      # OTEL Collector原始日志
    └── 2024/01/15/                # 按日期分区
        └── otel-collector-xxx.jsonl
```

## ⚙️ 配置详解

### CFS专用配置文件

OTEL Collector使用专门的CFS配置：`config-tencent-cfs.yaml`

关键配置项：

```yaml
exporters:
  # CFS文件导出器
  file/cfs_organized:
    path: "${CFS_MOUNT_POINT}/logs/${PROJ_SERVICE_NAME}/${CFS_DATE_PARTITION}/app.jsonl"
    rotation:
      max_megabytes: 50      # 50MB文件轮转（CFS优化）
      max_days: 90           # 90天保留期
      max_backups: 200       # 最大备份文件数
      localtime: true        # 使用本地时间
```

### 性能优化配置

```yaml
processors:
  batch:
    timeout: 2s              # 较长超时适应网络存储
    send_batch_size: 500     # 大批次提高网络效率
    send_batch_max_size: 1000

  memory_limiter:
    limit_mib: 512          # 更高内存限制用于CFS缓冲
    spike_limit_mib: 128
```

## 🔧 运维操作

### 日常管理命令

```bash
# 检查CFS挂载状态
mountpoint /mnt/cfs-logs

# 查看CFS使用情况
df -h /mnt/cfs-logs

# 检查OTEL容器日志
docker logs proj-otel-cfs-collector

# 手动重新挂载CFS
sudo umount /mnt/cfs-logs
sudo mount -t nfs -o vers=4.0,proto=tcp,fsc ${CFS_IP}:/ /mnt/cfs-logs
```

### 监控和告警

```bash
# 检查日志写入性能
curl http://127.0.0.1:8888/metrics | grep -i file

# 监控CFS相关指标
curl http://127.0.0.1:8889/metrics | grep cfs
```

### 故障排除

#### CFS挂载失败

```bash
# 检查网络连通性
ping ${CFS_IP}
telnet ${CFS_IP} 2049

# 检查NFS服务
showmount -e ${CFS_IP}

# 查看挂载日志
dmesg | grep nfs
journalctl -u rpc-statd
```

#### 日志写入失败

```bash
# 检查CFS权限
ls -la /mnt/cfs-logs/

# 查看OTEL错误日志
docker logs proj-otel-cfs-collector | grep -i error

# 测试手动写入
echo "test" > /mnt/cfs-logs/test.txt
```

## 📈 性能考虑

### CFS性能限制

- **吞吐量**: 根据CFS规格，通常100-300MB/s
- **IOPS**: 通常1000-3000 IOPS
- **延迟**: 网络存储延迟1-10ms

### 优化建议

1. **批量写入**: 使用较大的批处理大小
2. **文件轮转**: 适当的文件大小避免CFS性能下降
3. **本地缓冲**: 配置本地缓冲应对网络中断
4. **压缩传输**: 启用gzip压缩减少网络带宽

## 🔒 安全配置

### CFS访问控制

```bash
# 设置CFS目录权限
sudo chmod 755 /mnt/cfs-logs
sudo chown -R $(id -u):$(id -g) /mnt/cfs-logs/logs
```

### 容器安全

```bash
# 使用只读挂载本地日志目录
-v "${PROJ_ROOT_DIR}/logs:/opt/logs:ro"

# 限制容器资源
--memory=1g --cpu-quota=50000
```

## 🌐 环境变量参考

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `CFS_FILESYSTEM_ID` | 无 | **必需** CFS文件系统ID |
| `CFS_IP` | 无 | **必需** CFS挂载点IP地址 |
| `CFS_MOUNT_POINT` | `/mnt/cfs-logs` | CFS本地挂载点 |
| `TENCENT_REGION` | `ap-beijing` | 腾讯云地域 |
| `CFS_INSTANCE_ID` | 自动生成 | OTEL Collector实例ID |
| `PROJ_SERVICE_NAME` | `apiserver` | 服务名称 |
| `PROJ_ENVIRONMENT` | `development` | 部署环境 |

## 📚 相关文档

- [腾讯云CFS文档](https://cloud.tencent.com/document/product/582)
- [OpenTelemetry Collector配置](https://opentelemetry.io/docs/collector/configuration/)
- [NFS客户端配置](https://linux.die.net/man/5/nfs)

## 💡 最佳实践

1. **定期监控**: 监控CFS使用量和性能指标
2. **日志轮转**: 配置合适的日志轮转策略避免大文件
3. **备份策略**: 重要日志考虑多地备份
4. **成本控制**: 监控CFS存储成本和流量费用
5. **网络优化**: 使用同地域CFS减少网络延迟

## 🆘 技术支持

如果遇到问题：

1. 查看 [故障排除指南](../../../docs/guides/troubleshooting/)
2. 检查容器日志：`docker logs proj-otel-cfs-collector`
3. 验证CFS网络连通性和权限配置
4. 联系运维团队获取CFS访问权限
