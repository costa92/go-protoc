# 统一日志收集系统架构图

## 🏗️ 整体架构概览

```mermaid
graph TB
    subgraph "应用层"
        A1[Go应用服务]
        A2[其他微服务]
        A3[第三方服务]
    end
    
    subgraph "日志生成"
        L1[Zap Logger<br/>JSON格式]
        L2[其他Logger<br/>JSON格式]
        L3[File Logger<br/>JSON格式]
        
        A1 --> L1
        A2 --> L2
        A3 --> L3
    end
    
    subgraph "数据传输"
        O1[OTLP Export<br/>gRPC/HTTP]
        F1[File Write<br/>本地文件]
        
        L1 --> O1
        L1 --> F1
        L2 --> F1
        L3 --> F1
    end
    
    subgraph "OpenTelemetry Collector"
        subgraph "接收器"
            R1[OTLP Receiver<br/>:4327/:4328]
            R2[FileLogs Receiver<br/>100ms轮询]
            R3[Prometheus Receiver<br/>指标收集]
        end
        
        subgraph "处理器"
            P1[Memory Limiter<br/>512MB限制]
            P2[Resource Processor<br/>元数据增强]
            P3[Batch Processor<br/>批处理优化]
            
            subgraph "🔧 字段映射处理"
                PM[JSON Parser<br/>解析JSON日志]
                PF[Field Mapping<br/>msg → _msg]
                PE[Metadata Enhancement<br/>添加标识信息]
                
                PM --> PF
                PF --> PE
            end
        end
        
        subgraph "导出器"
            E1[Logging Export<br/>控制台调试]
            E2[VictoriaLogs Export<br/>:9428 OTLP]
            E3[File Export<br/>本地备份]
        end
        
        O1 --> R1
        F1 --> R2
        R1 --> P1
        R2 --> PM
        R3 --> P1
        P1 --> P2
        P2 --> P3
        PE --> P3
        P3 --> E1
        P3 --> E2
        P3 --> E3
    end
    
    subgraph "存储层"
        S1[控制台输出<br/>开发调试]
        S2[VictoriaLogs<br/>TSDB存储]
        S3[本地文件<br/>灾备存储]
        
        E1 --> S1
        E2 --> S2
        E3 --> S3
    end
    
    subgraph "查询分析层"
        Q1[LogSQL查询<br/>强大语法]
        Q2[Web UI界面<br/>:9428/vmui]
        Q3[Grafana集成<br/>可视化仪表板]
        Q4[API接口<br/>程序化访问]
        
        S2 --> Q1
        S2 --> Q2
        S2 --> Q3
        S2 --> Q4
    end

    classDef appLayer fill:#e1f5fe
    classDef collectLayer fill:#f3e5f5
    classDef processLayer fill:#fff3e0
    classDef storageLayer fill:#e8f5e8
    classDef queryLayer fill:#fce4ec
    
    class A1,A2,A3,L1,L2,L3 appLayer
    class R1,R2,R3,O1,F1 collectLayer
    class P1,P2,P3,PM,PF,PE processLayer
    class E1,E2,E3,S1,S2,S3 storageLayer
    class Q1,Q2,Q3,Q4 queryLayer
```

## 🔄 数据流转时序图

```mermaid
sequenceDiagram
    participant App as 应用程序
    participant OTLP as OTLP Export
    participant File as 文件系统
    participant OTC as OTEL Collector
    participant VL as VictoriaLogs
    participant UI as 查询界面
    
    Note over App, UI: 日志生成与处理流程
    
    App->>OTLP: 1. 实时推送日志 (gRPC/HTTP)
    App->>File: 2. 写入本地文件 (backup)
    
    OTLP->>OTC: 3. OTLP数据传输
    Note over File, OTC: 4. 文件监控 (100ms轮询)
    File-->>OTC: 5. 检测到文件变化
    
    rect rgb(255, 240, 245)
        Note over OTC: 🔧 关键处理步骤
        OTC->>OTC: 6. JSON解析
        OTC->>OTC: 7. 字段映射 (msg→_msg)
        OTC->>OTC: 8. 元数据增强
        OTC->>OTC: 9. 批处理优化
    end
    
    OTC->>VL: 10. 发送到VictoriaLogs
    VL->>VL: 11. 存储到TSDB
    
    UI->>VL: 12. LogSQL查询请求
    VL->>UI: 13. 返回查询结果
    
    Note over App, UI: 端到端延迟: ~100-200ms
```

## 🔧 字段映射修复详解

```mermaid
flowchart LR
    subgraph "修复前 ❌"
        A1[JSON日志<br/>{"msg":"hello"}]
        A2[JSON Parser<br/>解析字段]
        A3[错误映射<br/>msg → body.msg]
        A4[VictoriaLogs<br/>missing _msg field]
        
        A1 --> A2 --> A3 --> A4
    end
    
    subgraph "修复后 ✅"
        B1[JSON日志<br/>{"msg":"hello"}]
        B2[JSON Parser<br/>解析字段]
        B3[正确映射<br/>msg → _msg]
        B4[VictoriaLogs<br/>_msg: hello]
        
        B1 --> B2 --> B3 --> B4
    end
    
    style A3 fill:#ffcdd2
    style A4 fill:#ffcdd2
    style B3 fill:#c8e6c9
    style B4 fill:#c8e6c9
```

## 📊 性能指标与监控

### 系统性能特性

| 指标项 | 目标值 | 当前表现 |
|--------|--------|----------|
| **日志收集延迟** | < 200ms | ~100ms |
| **文件轮询间隔** | 100ms | 100ms ✅ |
| **批处理延迟** | < 1s | 500ms ✅ |
| **内存使用** | < 512MB | ~200MB ✅ |
| **CPU使用率** | < 50% | ~15% ✅ |
| **磁盘IO** | < 100MB/s | ~20MB/s ✅ |

### 监控端点

```mermaid
graph LR
    subgraph "监控端点"
        M1[健康检查<br/>:13133/health]
        M2[指标监控<br/>:8888/metrics]
        M3[性能分析<br/>:1777/pprof]
        M4[调试界面<br/>:55679/debug]
    end
    
    subgraph "查询端点"
        Q1[VictoriaLogs UI<br/>:9428/vmui]
        Q2[LogSQL API<br/>:9428/select/logsql/query]
        Q3[健康检查<br/>:9428/health]
    end
    
    subgraph "管理命令"
        C1[服务状态<br/>service.sh status all]
        C2[容器日志<br/>docker logs proj-otelcol]
        C3[配置验证<br/>docker inspect proj-otelcol]
    end
```

## 🔍 故障排查流程图

```mermaid
flowchart TD
    S[发现问题] --> D1{日志无法查询?}
    
    D1 -->|是| C1[检查VictoriaLogs状态]
    C1 --> C2{_msg字段缺失?}
    C2 -->|是| F1[🔧 修复字段映射<br/>msg → _msg]
    C2 -->|否| C3[检查网络连接]
    
    D1 -->|否| D2{收集延迟高?}
    D2 -->|是| C4[优化批处理配置]
    D2 -->|否| D3{内存使用高?}
    D3 -->|是| C5[调整内存限制]
    D3 -->|否| C6[检查其他组件]
    
    F1 --> V1[重启OTEL Collector]
    V1 --> V2[验证修复效果]
    V2 --> E[问题解决]
    
    C3 --> E
    C4 --> E
    C5 --> E
    C6 --> E
    
    style F1 fill:#c8e6c9
    style V2 fill:#c8e6c9
    style E fill:#4caf50,color:#fff
```

## 🚀 部署架构对比

| 部署方式 | 优势 | 适用场景 | 配置复杂度 |
|----------|------|----------|------------|
| **Docker容器** | 隔离性好，易扩展 | 开发测试，小规模生产 | ⭐⭐⭐ |
| **Linux原生** | 性能优异，资源利用率高 | 大规模生产环境 | ⭐⭐⭐⭐ |
| **Kubernetes** | 高可用，自动扩缩容 | 云原生，大规模集群 | ⭐⭐⭐⭐⭐ |

## 📈 扩展性考虑

```mermaid
graph TB
    subgraph "水平扩展"
        H1[多个OTEL Collector实例]
        H2[VictoriaLogs集群]
        H3[负载均衡]
    end
    
    subgraph "垂直扩展"
        V1[增加CPU/内存]
        V2[优化批处理大小]
        V3[调整轮询间隔]
    end
    
    subgraph "存储扩展"
        S1[数据分片]
        S2[冷热数据分离]
        S3[自动清理策略]
    end
```

---

## 📚 相关文档链接

- **配置指南**: `scripts/installation/otelcol/README.md`
- **开发文档**: `docs/logging-development.md`
- **快速开始**: `docs/logging-quick-start.md`
- **故障排除**: `docs/victorialogs-msg-field-fix.md`