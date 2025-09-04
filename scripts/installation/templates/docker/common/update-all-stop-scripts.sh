#!/usr/bin/env bash

# 批量更新所有Docker服务的停止脚本，统一数据卷删除处理方式
# 使用方式: bash update-all-stop-scripts.sh

set -eEuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATES_DIR="$(dirname "$SCRIPT_DIR")"

echo "🔄 开始统一所有Docker服务的数据卷删除处理方式..."

# 定义服务及其对应的数据卷和描述
declare -A SERVICES=(
    ["etcd"]="SERVICE_NAME=\"etcd\"
DATA_VOLUMES=\"\${PROJ_PREFIX}-etcd-data\"
DATA_DESCRIPTION=\"  - 服务配置键值对
  - 集群成员信息
  - 租约和监听器
  - 认证用户和角色
  - WAL日志文件\""

    ["kafka"]="SERVICE_NAME=\"Kafka\"
DATA_VOLUMES=\"\${PROJ_PREFIX}-kafka-data,\${PROJ_PREFIX}-kafka-logs\"
DATA_DESCRIPTION=\"  - 所有Topic数据
  - 消息日志
  - 分区数据
  - 索引文件
  - 事务状态日志\""

    ["mariadb"]="SERVICE_NAME=\"MariaDB\"
DATA_VOLUMES=\"proj-mariadb-data\"
DATA_DESCRIPTION=\"  - 所有数据库和表
  - 用户账户和权限
  - 索引和触发器
  - 存储过程和函数
  - 二进制日志\""

    ["mongodb"]="SERVICE_NAME=\"MongoDB\"
DATA_VOLUMES=\"\${PROJ_PREFIX}-mongodb-data,\${PROJ_PREFIX}-mongodb-config,\${PROJ_PREFIX}-mongodb-logs\"
DATA_DESCRIPTION=\"  - 所有数据库和集合
  - 索引和分片信息
  - 用户账户和权限
  - 复制集配置
  - 操作日志文件\""

    ["mysql"]="SERVICE_NAME=\"MySQL\"
DATA_VOLUMES=\"\${PROJ_PREFIX}-mysql-data\"
DATA_DESCRIPTION=\"  - 所有数据库和表
  - 用户账户和权限
  - 索引和触发器
  - 存储过程和函数
  - 二进制日志和慢查询日志\""

    ["nacos"]="SERVICE_NAME=\"Nacos\"
DATA_VOLUMES=\"\${PROJ_PREFIX}-nacos-data,\${PROJ_PREFIX}-nacos-logs\"
DATA_DESCRIPTION=\"  - 服务注册信息
  - 配置管理数据
  - 命名空间配置
  - 用户和权限数据\""

    ["otelcol"]="SERVICE_NAME=\"OpenTelemetry Collector\"
DATA_VOLUMES=\"\${PROJ_PREFIX}-otelcol-data\"
DATA_DESCRIPTION=\"  - 临时缓存数据
  - 处理器状态
  - 导出器队列
  - 配置文件备份
  - 统计和诊断数据\""

    ["victorialogs"]="SERVICE_NAME=\"VictoriaLogs\"
DATA_VOLUMES=\"\${PROJ_PREFIX}-victorialogs-data\"
DATA_DESCRIPTION=\"  - 所有应用程序日志
  - 系统日志记录
  - 结构化日志数据
  - 索引和查询缓存
  - 日志流元数据\""

    ["zookeeper"]="SERVICE_NAME=\"Zookeeper\"
DATA_VOLUMES=\"\${PROJ_PREFIX}-zookeeper-data,\${PROJ_PREFIX}-zookeeper-logs\"
DATA_DESCRIPTION=\"  - Zookeeper集群数据
  - 事务日志
  - 快照数据
  - 配置信息\""
)

# 统一的数据卷删除处理模板
read -r -d '' VOLUME_DELETION_TEMPLATE << 'EOF' || true
# === 统一数据卷删除处理 ===
SERVICE_NAME_PLACEHOLDER
DATA_VOLUMES_PLACEHOLDER
DATA_DESCRIPTION_PLACEHOLDER

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

if [ "$should_remove_data" = true ]; then
    echo ""
    echo "⚠️  警告：将删除${SERVICE_NAME}数据卷！这将永久删除所有数据！"
    echo "包括："
    echo "$DATA_DESCRIPTION"
    echo ""
    
    # 交互式确认
    if [ -t 0 ]; then  # 检查是否为交互式终端
        read -p "确认删除数据卷？(y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            IFS=',' read -ra VOLUME_ARRAY <<< "$DATA_VOLUMES"
            deleted_volumes=()
            for volume in "${VOLUME_ARRAY[@]}"; do
                volume=$(echo "$volume" | xargs)  # 去除空格
                if docker volume rm "$volume" 2>/dev/null; then
                    deleted_volumes+=("$volume")
                fi
            done
            if [ ${#deleted_volumes[@]} -gt 0 ]; then
                echo "${SERVICE_NAME}数据卷已删除: ${deleted_volumes[*]}"
                echo "所有${SERVICE_NAME}数据已被永久删除"
            else
                echo "未找到要删除的数据卷"
            fi
        else
            echo "数据卷保留: $DATA_VOLUMES"
        fi
    else
        # 非交互式环境，需要环境变量确认
        if [[ "${FORCE_DELETE:-}" == "true" ]]; then
            IFS=',' read -ra VOLUME_ARRAY <<< "$DATA_VOLUMES"
            for volume in "${VOLUME_ARRAY[@]}"; do
                volume=$(echo "$volume" | xargs)
                docker volume rm "$volume" 2>/dev/null || true
            done
            echo "${SERVICE_NAME}数据卷已强制删除: $DATA_VOLUMES"
        else
            echo "⚠️  非交互式环境，需要设置 FORCE_DELETE=true 进行强制删除"
            echo "数据卷保留: $DATA_VOLUMES"
        fi
    fi
else
    echo ""
    echo "💡 提示: ${SERVICE_NAME}数据卷已保留"
    IFS=',' read -ra VOLUME_ARRAY <<< "$DATA_VOLUMES"
    for volume in "${VOLUME_ARRAY[@]}"; do
        volume=$(echo "$volume" | xargs)
        echo "   数据卷: $volume"
    done
    echo ""
    echo "   删除数据的方式："
    echo "     环境变量: REMOVE_DATA=true make docker.SERVICE_LOWER.stop"
    echo "     直接调用: bash <脚本路径> --remove-data"
    echo "     强制删除: REMOVE_DATA=true FORCE_DELETE=true make docker.SERVICE_LOWER.stop"
fi
EOF

# 处理每个服务
for service in "${!SERVICES[@]}"; do
    service_file="$TEMPLATES_DIR/$service/docker-stop.sh.tpl"
    
    if [[ ! -f "$service_file" ]]; then
        echo "⚠️  跳过 $service: 文件不存在 $service_file"
        continue
    fi
    
    echo "🔧 处理服务: $service"
    
    # 创建备份
    cp "$service_file" "$service_file.backup.$(date +%Y%m%d_%H%M%S)"
    
    # 获取服务配置
    service_config="${SERVICES[$service]}"
    
    # 创建临时模板
    temp_template="$VOLUME_DELETION_TEMPLATE"
    
    # 替换占位符
    temp_template="${temp_template//SERVICE_NAME_PLACEHOLDER/$service_config}"
    temp_template="${temp_template//DATA_VOLUMES_PLACEHOLDER/}"
    temp_template="${temp_template//DATA_DESCRIPTION_PLACEHOLDER/}"
    temp_template="${temp_template//SERVICE_LOWER/${service}}"
    
    # 找到现有的数据卷删除部分并替换
    # 使用 awk 来精确替换
    awk -v template="$temp_template" -v config="$service_config" '
    BEGIN { 
        in_volume_section = 0
        printed_new = 0
        split(config, config_lines, "\n")
        for (i in config_lines) {
            if (config_lines[i] ~ /SERVICE_NAME=/) {
                service_name = config_lines[i]
                gsub(/SERVICE_NAME="/, "", service_name)
                gsub(/".*/, "", service_name)
            }
            if (config_lines[i] ~ /DATA_VOLUMES=/) {
                data_volumes = config_lines[i]
            }
            if (config_lines[i] ~ /DATA_DESCRIPTION=/) {
                data_description = config_lines[i]
                # 收集多行描述
                for (j = i+1; j in config_lines; j++) {
                    if (config_lines[j] ~ /^  -/) {
                        data_description = data_description "\n" config_lines[j]
                    } else {
                        break
                    }
                }
            }
        }
    }
    /^# (可选：删除数据卷|=== 统一数据卷删除处理)/ { 
        if (!printed_new) {
            print "# === 统一数据卷删除处理 ==="
            print service_name
            print data_volumes  
            print data_description
            print ""
            # 插入统一模板的其余部分
            print "# 支持多种数据删除触发方式"
            print "should_remove_data=false"
            print ""
            print "# 方式1: 环境变量 (推荐用于Make命令)"
            print "if [[ \"${REMOVE_DATA:-}\" == \"true\" ]]; then"
            print "    should_remove_data=true"
            print "fi"
            print ""
            print "# 方式2: 命令行参数"  
            print "for arg in \"$@\"; do"
            print "    case $arg in"
            print "        --remove-data|--force)"
            print "            should_remove_data=true"
            print "            break"
            print "            ;;"
            print "    esac"
            print "done"
            print ""
            print "if [ \"$should_remove_data\" = true ]; then"
            print "    echo \"\""
            print "    echo \"⚠️  警告：将删除${SERVICE_NAME}数据卷！这将永久删除所有数据！\""
            print "    echo \"包括：\""
            print "    echo \"$DATA_DESCRIPTION\""
            print "    echo \"\""
            print "    "
            print "    # 交互式确认"
            print "    if [ -t 0 ]; then  # 检查是否为交互式终端"
            print "        read -p \"确认删除数据卷？(y/N): \" -n 1 -r"
            print "        echo"
            print "        if [[ $REPLY =~ ^[Yy]$ ]]; then"
            print "            IFS=',' read -ra VOLUME_ARRAY <<< \"$DATA_VOLUMES\""
            print "            deleted_volumes=()"
            print "            for volume in \"${VOLUME_ARRAY[@]}\"; do"
            print "                volume=$(echo \"$volume\" | xargs)  # 去除空格"
            print "                if docker volume rm \"$volume\" 2>/dev/null; then"
            print "                    deleted_volumes+=(\"$volume\")"
            print "                fi"
            print "            done"
            print "            if [ ${#deleted_volumes[@]} -gt 0 ]; then"
            print "                echo \"${SERVICE_NAME}数据卷已删除: ${deleted_volumes[*]}\""
            print "                echo \"所有${SERVICE_NAME}数据已被永久删除\""
            print "            else"
            print "                echo \"未找到要删除的数据卷\""
            print "            fi"
            print "        else"
            print "            echo \"数据卷保留: $DATA_VOLUMES\""
            print "        fi"
            print "    else"
            print "        # 非交互式环境，需要环境变量确认"
            print "        if [[ \"${FORCE_DELETE:-}\" == \"true\" ]]; then"
            print "            IFS=',' read -ra VOLUME_ARRAY <<< \"$DATA_VOLUMES\""
            print "            for volume in \"${VOLUME_ARRAY[@]}\"; do"
            print "                volume=$(echo \"$volume\" | xargs)"
            print "                docker volume rm \"$volume\" 2>/dev/null || true"
            print "            done"
            print "            echo \"${SERVICE_NAME}数据卷已强制删除: $DATA_VOLUMES\""
            print "        else"
            print "            echo \"⚠️  非交互式环境，需要设置 FORCE_DELETE=true 进行强制删除\""
            print "            echo \"数据卷保留: $DATA_VOLUMES\""
            print "        fi"
            print "    fi"
            print "else"
            print "    echo \"\""
            print "    echo \"💡 提示: ${SERVICE_NAME}数据卷已保留\""
            print "    IFS=',' read -ra VOLUME_ARRAY <<< \"$DATA_VOLUMES\""
            print "    for volume in \"${VOLUME_ARRAY[@]}\"; do"
            print "        volume=$(echo \"$volume\" | xargs)"
            print "        echo \"   数据卷: $volume\""
            print "    done"
            print "    echo \"\""
            print "    echo \"   删除数据的方式：\""
            printf "    echo \"     环境变量: REMOVE_DATA=true make docker.%s.stop\"\n", "'$service'"
            print "    echo \"     直接调用: bash <脚本路径> --remove-data\""
            printf "    echo \"     强制删除: REMOVE_DATA=true FORCE_DELETE=true make docker.%s.stop\"\n", "'$service'"
            print "fi"
            printed_new = 1
        }
        in_volume_section = 1
        next
    }
    in_volume_section && /^fi$/ && printed_new { 
        in_volume_section = 0
        next
    }
    !in_volume_section { print }
    ' "$service_file" > "$service_file.tmp"
    
    # 替换原文件
    mv "$service_file.tmp" "$service_file"
    
    echo "✅ $service 更新完成"
done

echo ""
echo "🎉 所有服务的数据卷删除处理已统一完成！"
echo ""
echo "📋 统一的使用方式："
echo "   1. 基本停止 (保留数据): make docker.<service>.stop"
echo "   2. 删除数据 (交互确认): REMOVE_DATA=true make docker.<service>.stop"  
echo "   3. 强制删除 (非交互): REMOVE_DATA=true FORCE_DELETE=true make docker.<service>.stop"
echo "   4. 命令行方式: bash <script> --remove-data"
echo ""
echo "💡 备份文件已保存，可在需要时恢复"