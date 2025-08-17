# vmagent 配置文件说明

## 📋 概述

vmagent 支持两种安装模式，每种模式使用不同的配置文件来适应其网络环境：

## 📁 配置文件结构

```sh
scripts/installation/victoria/
├── vmagent.yml         # 本地/宿主机安装配置
└── vmagent-docker.yml  # Docker 容器安装配置
```

## 🔧 配置差异对比

### 1. 本地安装模式 (`vmagent.yml`)

**使用场景**：在宿主机上直接安装 vmagent 二进制文件

**目标地址**：

```yaml
scrape_configs:
  - job_name: 'victoria-metrics'
    static_configs:
      - targets: ['127.0.0.1:8428']  # localhost 地址

  - job_name: 'victoria-logs'
    static_configs:
      - targets: ['127.0.0.1:9428']  # localhost 地址

  - job_name: 'vmagent'
    static_configs:
      - targets: ['127.0.0.1:8429']  # 自监控

  - job_name: 'go-protoc-api'
    static_configs:
      - targets: ['127.0.0.1:8080']  # API 服务器
```

### 2. Docker 安装模式 (`vmagent-docker.yml`)

**使用场景**：在 Docker 容器中运行 vmagent

**目标地址**：

```yaml
scrape_configs:
  - job_name: 'victoria-metrics'
    static_configs:
      - targets: ['proj-victoriametrics:8428']  # 容器名称

  - job_name: 'victoria-logs'
    static_configs:
      - targets: ['proj-victorialogs:9428']  # 容器名称

  - job_name: 'vmagent'
    static_configs:
      - targets: ['proj-vmagent:8429']  # 自监控容器名

  - job_name: 'go-protoc-api'
    static_configs:
      - targets: ['host.docker.internal:8080']  # 宿主机服务
```

## 🚀 自动选择机制

### 本地安装

```bash
# 调用本地安装会生成 vmagent.yml
./scripts/installation/victoria.sh vmagent.install

# 生成的配置文件
scripts/installation/victoria/vmagent.yml  # 使用 localhost 地址
```

### Docker 安装

```bash
# 调用 Docker 安装会生成 vmagent-docker.yml
./scripts/installation/victoria.sh vmagent.docker.install

# 生成的配置文件
scripts/installation/victoria/vmagent-docker.yml  # 使用容器名称
```

## 🌐 网络架构说明

### 本地网络模式

```sh
宿主机 (127.0.0.1)
├─ VictoriaMetrics  :8428
├─ VictoriaLogs     :9428
├─ vmagent          :8429
└─ API Server       :8080
```

### Docker 网络模式

```sh
Docker Network: proj (172.18.0.0/16)
├─ proj-victoriametrics  :8428
├─ proj-victorialogs     :9428
└─ proj-vmagent          :8429
   └─ 访问: host.docker.internal:8080 (API服务在宿主机)
```

## ⚙️ 配置文件使用

### systemd 服务 (本地安装)

```bash
# 服务文件路径
/etc/systemd/system/vmagent.service

# 使用的配置文件
-promscrape.config=/path/to/victoria/vmagent.yml
```

### Docker 容器 (Docker 安装)

```bash
# 容器挂载
-v /path/to/victoria/vmagent-docker.yml:/etc/vmagent.yml

# 容器内使用
-promscrape.config=/etc/vmagent.yml
```

## 🔍 验证配置

### 查看当前目标

```bash
# 检查 vmagent 目标状态
curl http://127.0.0.1:8429/targets

# 所有目标应显示为 "up" 状态
```

### 配置文件位置

```bash
# 查看配置文件
ls -la scripts/installation/victoria/vmagent*.yml

# 对比配置差异
diff scripts/installation/victoria/vmagent.yml \
     scripts/installation/victoria/vmagent-docker.yml
```

## 📊 监控指标

### 成功的配置应显示

- ✅ **victoria-metrics** (1/1 up)
- ✅ **victoria-logs** (1/1 up)
- ✅ **vmagent** (1/1 up)
- ✅ **go-protoc-api** (1/1 up)

### 失败指标 (需要检查配置)

- ❌ **connection refused** - 网络地址错误
- ❌ **timeout** - 服务未运行
- ❌ **not found** - 端点不存在

## 🛠️ 故障排除

### 容器网络问题

```bash
# 检查容器网络
docker network inspect proj

# 验证容器间连通性
docker exec proj-vmagent ping proj-victoriametrics
```

### 配置文件问题

```bash
# 验证配置语法
yamllint scripts/installation/victoria/vmagent-docker.yml

# 检查目标可达性
curl http://127.0.0.1:8428/metrics  # VictoriaMetrics
curl http://127.0.0.1:9428/metrics  # VictoriaLogs
```

## 📝 最佳实践

1. **开发环境**：使用 Docker 模式便于快速部署
2. **生产环境**：根据架构选择适合的安装模式
3. **配置管理**：两个配置文件都应纳入版本控制
4. **监控验证**：定期检查目标状态确保正常采集

---

💡 **提示**：无论选择哪种模式，vmagent 都会自动使用相应的配置文件，无需手动干预。
