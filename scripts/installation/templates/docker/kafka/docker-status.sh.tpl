#!/usr/bin/env bash

# Kafka Docker状态检查脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Kafka ${KAFKA_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="kafka"
readonly CONTAINER_NAME="${PROJ_PREFIX}-kafka"
readonly SERVICE_PORT="${PROJ_KAFKA_PORT:-9092}"
readonly ZOOKEEPER_PORT="${PROJ_ZOOKEEPER_PORT:-2181}"

# 平台检测和工具路径选择
if [[ "$(uname)" == "Darwin" ]]; then
    # macOS - 使用 Confluent 镜像工具路径
    readonly KAFKA_TOOLS_PATH="/bin"
else
    # Linux - 使用 Bitnami 镜像工具路径
    readonly KAFKA_TOOLS_PATH="/opt/bitnami/kafka/bin"
fi

# 检查容器状态
echo "=== Kafka容器状态检查 ==="
if docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -q "${CONTAINER_NAME}"; then
    echo "✅ Kafka容器正在运行"
    docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    
    # 检查Zookeeper依赖
    echo ""
    echo "=== Zookeeper依赖检查 ==="
    if docker ps --filter name="${PROJ_PREFIX}-zookeeper" --format "{{.Names}}" | grep -q "${PROJ_PREFIX}-zookeeper"; then
        echo "✅ Zookeeper容器正在运行"
        # 检查 Zookeeper 端口连通性，如果能连接则认为服务正常
        if nc -z localhost "${ZOOKEEPER_PORT}" 2>/dev/null; then
            echo "✅ Zookeeper服务正常 (端口 ${ZOOKEEPER_PORT} 可访问)"
        else
            echo "❌ Zookeeper服务异常 (端口 ${ZOOKEEPER_PORT} 不可访问)"
        fi
    else
        echo "❌ Zookeeper容器未运行"
    fi
    
    # 检查Kafka服务状态
    echo ""
    echo "=== Kafka服务健康检查 ==="
    # 使用平台相关的工具路径
    if docker exec "${CONTAINER_NAME}" ${KAFKA_TOOLS_PATH}/kafka-topics --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER} --list >/dev/null 2>&1; then
        echo "✅ Kafka服务正常 (可以连接到bootstrap服务器)"
    else
        echo "❌ Kafka服务异常 (无法连接到bootstrap服务器)"
    fi
    
    # 检查服务端口
    echo ""
    echo "=== 端口连通性 ==="
    if nc -z 127.0.0.1 "${SERVICE_PORT}" 2>/dev/null; then
        echo "✅ Kafka端口 ${SERVICE_PORT} 可访问"
    else
        echo "❌ Kafka端口 ${SERVICE_PORT} 不可访问"
    fi
    
    if nc -z 127.0.0.1 29092 2>/dev/null; then
        echo "✅ Kafka内部端口 29092 可访问"
    else
        echo "❌ Kafka内部端口 29092 不可访问"
    fi
    
    if nc -z 127.0.0.1 9999 2>/dev/null; then
        echo "✅ JMX监控端口 9999 可访问"
    else
        echo "❌ JMX监控端口 9999 不可访问"
    fi
    
    # 显示Topics列表
    echo ""
    echo "=== Topics列表 ==="
    docker exec "${CONTAINER_NAME}" ${KAFKA_TOOLS_PATH}/kafka-topics --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER} --list 2>/dev/null || echo "无法获取Topics列表"
    
    # 显示Broker信息
    echo ""
    echo "=== Broker信息 ==="
    docker exec "${CONTAINER_NAME}" ${KAFKA_TOOLS_PATH}/kafka-broker-api-versions --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER} 2>/dev/null | head -5 || echo "无法获取Broker信息"
    
    # 显示服务信息
    echo ""
    echo "=== 服务信息 ==="
    echo "🔗 Bootstrap服务器: localhost:${SERVICE_PORT}"
    echo "🔌 内部通信端口: localhost:29092"
    echo "📊 JMX监控端口: localhost:9999"
    echo "🗂️  Zookeeper连接: proj-zookeeper:${ZOOKEEPER_PORT}"
    
    # 显示常用命令
    echo ""
    echo "=== 常用管理命令 ==="
    echo "📝 创建Topic: docker exec ${CONTAINER_NAME} ${KAFKA_TOOLS_PATH}/kafka-topics --create --topic <topic-name> --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER}"
    echo "📋 列出Topics: docker exec ${CONTAINER_NAME} ${KAFKA_TOOLS_PATH}/kafka-topics --list --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER}"
    echo "📤 生产消息: docker exec -it ${CONTAINER_NAME} ${KAFKA_TOOLS_PATH}/kafka-console-producer --topic <topic-name> --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER}"
    echo "📥 消费消息: docker exec -it ${CONTAINER_NAME} ${KAFKA_TOOLS_PATH}/kafka-console-consumer --topic <topic-name> --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER} --from-beginning"
    
    # 显示最近日志
    echo ""
    echo "=== 最近日志 (最后10行) ==="
    docker logs "${CONTAINER_NAME}" --tail 10
    
else
    echo "❌ Kafka容器未运行"
    echo ""
    echo "启动顺序:"
    echo "  1. 启动Zookeeper: make docker.zookeeper.start"
    echo "  2. 启动Kafka: make docker.kafka.start"
    echo ""
    echo "或使用脚本:"
    echo "  ./scripts/installation/templates/docker/zookeeper/docker-run.sh"
    echo "  ./scripts/installation/templates/docker/kafka/docker-run.sh"
fi

# 检查数据卷
echo ""
echo "=== 数据卷状态 ==="
docker volume ls | grep -E "(${PROJ_PREFIX}-kafka-data|${PROJ_PREFIX}-kafka-logs)" || echo "未找到Kafka数据卷"