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
        if echo ruok | nc localhost "${ZOOKEEPER_PORT}" 2>/dev/null | grep -q imok; then
            echo "✅ Zookeeper服务正常 (ruok -> imok)"
        else
            echo "❌ Zookeeper服务异常"
        fi
    else
        echo "❌ Zookeeper容器未运行"
    fi
    
    # 检查Kafka服务状态
    echo ""
    echo "=== Kafka服务健康检查 ==="
    # 检测平台以使用正确的命令
    if [[ "$(uname)" == "Darwin" ]]; then
        # macOS 使用 Confluent 镜像命令
        if docker exec "${CONTAINER_NAME}" kafka-topics --bootstrap-server localhost:9092 --list >/dev/null 2>&1; then
            echo "✅ Kafka服务正常 (可以连接到bootstrap服务器)"
        else
            echo "❌ Kafka服务异常 (无法连接到bootstrap服务器)"
        fi
    else
        # Linux 使用官方镜像命令
        if docker exec "${CONTAINER_NAME}" kafka-topics.sh --bootstrap-server localhost:9092 --list >/dev/null 2>&1; then
            echo "✅ Kafka服务正常 (可以连接到bootstrap服务器)"
        else
            echo "❌ Kafka服务异常 (无法连接到bootstrap服务器)"
        fi
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
    if [[ "$(uname)" == "Darwin" ]]; then
        docker exec "${CONTAINER_NAME}" kafka-topics --bootstrap-server localhost:9092 --list 2>/dev/null || echo "无法获取Topics列表"
    else
        docker exec "${CONTAINER_NAME}" kafka-topics.sh --bootstrap-server localhost:9092 --list 2>/dev/null || echo "无法获取Topics列表"
    fi
    
    # 显示Broker信息
    echo ""
    echo "=== Broker信息 ==="
    if [[ "$(uname)" == "Darwin" ]]; then
        docker exec "${CONTAINER_NAME}" kafka-broker-api-versions --bootstrap-server localhost:9092 2>/dev/null | head -5 || echo "无法获取Broker信息"
    else
        docker exec "${CONTAINER_NAME}" kafka-broker-api-versions.sh --bootstrap-server localhost:9092 2>/dev/null | head -5 || echo "无法获取Broker信息"
    fi
    
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
    if [[ "$(uname)" == "Darwin" ]]; then
        echo "📝 创建Topic: docker exec ${CONTAINER_NAME} kafka-topics --create --topic <topic-name> --bootstrap-server localhost:9092"
        echo "📋 列出Topics: docker exec ${CONTAINER_NAME} kafka-topics --list --bootstrap-server localhost:9092"
        echo "📤 生产消息: docker exec -it ${CONTAINER_NAME} kafka-console-producer --topic <topic-name> --bootstrap-server localhost:9092"
        echo "📥 消费消息: docker exec -it ${CONTAINER_NAME} kafka-console-consumer --topic <topic-name> --bootstrap-server localhost:9092 --from-beginning"
    else
        echo "📝 创建Topic: docker exec ${CONTAINER_NAME} kafka-topics.sh --create --topic <topic-name> --bootstrap-server localhost:9092"
        echo "📋 列出Topics: docker exec ${CONTAINER_NAME} kafka-topics.sh --list --bootstrap-server localhost:9092"
        echo "📤 生产消息: docker exec -it ${CONTAINER_NAME} kafka-console-producer.sh --topic <topic-name> --bootstrap-server localhost:9092"
        echo "📥 消费消息: docker exec -it ${CONTAINER_NAME} kafka-console-consumer.sh --topic <topic-name> --bootstrap-server localhost:9092 --from-beginning"
    fi
    
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