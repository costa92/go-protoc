# 文档重组计划

## 📊 现状分析

### 当前文档统计

- **总文档数量**: 19个 markdown 文件
- **主要问题**:
  - 文档分散，缺乏清晰分类
  - 存在内容重复（如日志架构相关文档）
  - 命名不一致（大小写混合）
  - 缺乏统一的文档索引

### 识别的重复内容

| 主题类别 | 文档文件 | 重复问题 |
|---------|---------|---------|
| **日志架构** | `logging-architecture.md`<br>`logging-architecture-diagram.md` | 内容重叠，一个是图表，一个是详细说明 |
| **OTEL架构** | `otel-architecture-design.md`<br>`otel-installation-guide.md` | 架构设计与安装指南分离，但有部分架构描述重复 |
| **错误处理** | `ERROR_HANDLING_ARCHITECTURE.md`<br>`ERROR_NAMING_CONVENTIONS.md` | 架构与命名规范应该整合 |

## 🎯 新的分类结构设计

```
docs/
├── README.md                           # 文档总索引（新建）
├──
├── 📋 architecture/                    # 架构设计文档
│   ├── README.md                       # 架构文档索引
│   ├── system-overview.md              # 系统整体架构（新建）
│   ├── error-handling.md               # 错误处理架构（合并）
│   ├── logging-system.md               # 日志系统架构（合并）
│   └── observability.md                # 可观测性架构（OTEL相关合并）
│
├── 📚 guides/                          # 使用指南
│   ├── README.md                       # 指南索引
│   ├── development/                    # 开发指南
│   │   ├── getting-started.md          # 开发入门（基于DEVELOPMENT.md）
│   │   ├── makefile-usage.md           # Makefile使用（合并中英文）
│   │   └── validation-usage.md         # 数据验证使用
│   ├── deployment/                     # 部署指南
│   │   ├── otel-installation.md        # OTEL安装部署
│   │   └── logging-setup.md            # 日志系统部署
│   └── troubleshooting/                # 故障排除
│       ├── logging-issues.md           # 日志相关问题
│       ├── otel-fixes.md               # OTEL相关修复
│       └── connection-pool.md          # 连接池监控
│
├── 📖 reference/                       # 参考文档
│   ├── README.md                       # 参考文档索引
│   ├── api/                            # API 文档
│   │   ├── errors-code/                # 错误码文档（移动现有）
│   │   └── generated/                  # 自动生成的API文档
│   └── specifications/                 # 规范文档
│       ├── error-conventions.md        # 错误命名规范
│       └── version-sync.md             # 版本同步规范
│
└── 🗂️ legacy/                         # 历史文档（保留但不推荐）
    ├── README.md                       # 说明这些是历史文档
    └── [原有的临时/修复类文档]
```

## 🔄 文档合并和重组策略

### 1. 架构文档整合

#### `architecture/observability.md` (新建)

**合并来源**:

- `otel-architecture-design.md` (主要内容)
- `otel-installation-guide.md` (安装部分移到guides/)
- `logging-architecture.md` (可观测性相关部分)

**整合原则**:

- 统一的架构视角
- 从 Logs、Metrics、Traces 三个维度组织
- 包含 Agent-Collector-SaaS 完整架构

#### `architecture/logging-system.md` (新建)

**合并来源**:

- `logging-architecture.md` (主要内容)
- `logging-architecture-diagram.md` (图表部分)
- `logging-simple-diagram.md` (简化图表)

**整合原则**:

- 图文结合，架构图嵌入文档中
- 统一的系统视角
- 包含技术选型和设计原则

#### `architecture/error-handling.md` (新建)

**合并来源**:

- `ERROR_HANDLING_ARCHITECTURE.md` (架构设计)
- `ERROR_NAMING_CONVENTIONS.md` (命名规范部分)

### 2. 指南文档重组

#### `guides/development/makefile-usage.md` (新建)

**合并来源**:

- `MAKEFILE_DESIGN.md`
- `MAKEFILE_DESIGN.zh-CN.md`

**整合方式**:

- 统一为英文文档，重要概念提供中文说明
- 按功能分组组织内容

#### `guides/troubleshooting/` (新建目录)

**包含内容**:

- `logging-issues.md` (基于 `victorialogs-msg-field-fix.md`)
- `otel-fixes.md` (基于 `otelcol-loki-exporter-fix.md`)
- `connection-pool.md` (基于 `CONNECTION_POOL_MONITORING.md`)

### 3. 参考文档清理

#### 保留现有结构

- `guide/zh-CN/` - 保持现有的中文指南结构
- `generated/`, `html/`, `json/`, `markdown/` - 保持生成文档结构

#### 规范文档整理

- 将错误命名规范等移入 `reference/specifications/`

## 📝 实施计划

### 阶段一：创建新目录结构

1. 创建新的目录结构
2. 创建各级 README.md 索引文件

### 阶段二：内容合并和重写

1. 合并重复的架构文档
2. 重组指南类文档
3. 清理和分类参考文档

### 阶段三：链接更新和验证

1. 更新文档间的相互引用
2. 更新项目主 README 中的文档链接
3. 验证文档完整性

### 阶段四：历史文档处理

1. 将不再需要的文档移入 legacy/
2. 添加重定向说明

## 🎯 预期效果

### 用户体验改善

- **清晰导航**: 通过分层索引快速找到所需文档
- **减少重复**: 合并重复内容，信息更准确
- **逻辑分组**: 按用户需求分组（架构理解、开发指南、故障排除）

### 维护效率提升

- **统一标准**: 统一的文档命名和组织规范
- **更新简化**: 单一信息源，减少多处维护
- **扩展性**: 清晰的目录结构便于添加新文档

### 信息质量提升

- **内容去重**: 消除冗余和矛盾信息
- **结构优化**: 逻辑更清晰的信息组织
- **索引完善**: 完整的文档导航体系
