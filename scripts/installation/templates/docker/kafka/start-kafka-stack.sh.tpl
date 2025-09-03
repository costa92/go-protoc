#!/usr/bin/env bash

# Kafka Stack统一启动脚本模板
# Project: ${PROJ_NAME:-go-protoc}  
# Service: Kafka Stack (Zookeeper + Kafka)
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 颜色定义
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# 服务配置
readonly ZOOKEEPER_CONTAINER="${PROJ_PREFIX}-zookeeper"
readonly KAFKA_CONTAINER="${PROJ_PREFIX}-kafka"
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

echo -e "${BLUE}🚀 启动Kafka技术栈 (Zookeeper + Kafka)${NC}"
echo "=================================="

# 函数：等待容器启动并健康
wait_for_container() {
    local container_name="$1"
    local check_cmd="$2"
    local service_name="$3"
    local timeout=60
    
    echo -e "${YELLOW}等待 ${service_name} 启动完成...${NC}"
    
    while [ $timeout -gt 0 ]; do
        if eval "$check_cmd" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ ${service_name} 启动成功${NC}"
            return 0
        fi
        sleep 2
        ((timeout-=2))
        echo -n "."
    done
    
    echo -e "${RED}❌ ${service_name} 启动超时${NC}"
    docker logs "$container_name" --tail 10
    return 1
}

# 第一步：启动Zookeeper
echo -e "${BLUE}📋 第1步: 启动Zookeeper服务${NC}"
echo "正在生成Zookeeper启动脚本..."

# 调用模板生成Zookeeper脚本
proj::template::generate_config "kafka" "docker" "zookeeper-run.sh.tpl" "/tmp/zookeeper-run.sh"
chmod +x /tmp/zookeeper-run.sh

# 执行Zookeeper启动脚本
echo "执行Zookeeper启动..."
if /tmp/zookeeper-run.sh; then
    echo -e "${GREEN}✅ Zookeeper启动完成${NC}"
else
    echo -e "${RED}❌ Zookeeper启动失败${NC}"
    exit 1
fi

# 第二步：启动Kafka
echo ""
echo -e "${BLUE}📋 第2步: 启动Kafka服务${NC}"
echo "正在生成Kafka启动脚本..."

# 调用模板生成Kafka脚本  
proj::template::generate_config "kafka" "docker" "kafka-only-run.sh.tpl" "/tmp/kafka-only-run.sh"
chmod +x /tmp/kafka-only-run.sh

# 执行Kafka启动脚本
echo "执行Kafka启动..."
if /tmp/kafka-only-run.sh; then
    echo -e "${GREEN}✅ Kafka启动完成${NC}"
else
    echo -e "${RED}❌ Kafka启动失败${NC}"
    exit 1
fi

# 清理临时脚本
rm -f /tmp/zookeeper-run.sh /tmp/kafka-only-run.sh

# 第三步：验证整体服务状态
echo ""
echo -e "${BLUE}📋 第3步: 验证服务状态${NC}"

# 检查容器状态
echo "容器状态:"
docker ps --filter name="${ZOOKEEPER_CONTAINER}" --filter name="${KAFKA_CONTAINER}" \
    --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 检查网络连接
echo ""
echo "网络连接状态:"
if docker network inspect "${NETWORK_NAME}" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Docker网络 '${NETWORK_NAME}' 正常${NC}"
    echo "网络中的容器:"
    docker network inspect "${NETWORK_NAME}" --format '{{range $key, $value := .Containers}}  - {{$value.Name}} ({{$value.IPv4Address}}){{"\n"}}{{end}}'
else
    echo -e "${RED}❌ Docker网络 '${NETWORK_NAME}' 不存在${NC}"
fi

# 集成测试
echo ""
echo -e "${BLUE}📋 第4步: 集成测试${NC}"

# 平台检测和工具路径设置
if [[ "$(uname)" == "Darwin" ]]; then
    KAFKA_TOOLS_PATH="/bin"
else
    KAFKA_TOOLS_PATH="/opt/bitnami/kafka/bin"
fi

# 测试Kafka Topic操作
echo "测试Kafka功能..."
if docker exec "${KAFKA_CONTAINER}" ${KAFKA_TOOLS_PATH}/kafka-topics --create --topic test-integration --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER} --partitions 1 --replication-factor 1 >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Topic创建成功${NC}"
    
    if docker exec "${KAFKA_CONTAINER}" ${KAFKA_TOOLS_PATH}/kafka-topics --list --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER} 2>/dev/null | grep -q "test-integration"; then
        echo -e "${GREEN}✅ Topic列表查询正常${NC}"
        
        # 清理测试Topic
        docker exec "${KAFKA_CONTAINER}" ${KAFKA_TOOLS_PATH}/kafka-topics --delete --topic test-integration --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER} >/dev/null 2>&1
    else
        echo -e "${YELLOW}⚠️  Topic列表查询异常${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Topic创建失败，但服务可能仍在初始化${NC}"
fi

# 最终状态报告
echo ""
echo -e "${GREEN}🎉 Kafka技术栈启动完成!${NC}"
echo "=================================="
echo -e "${BLUE}📊 服务信息:${NC}"
echo "  🔗 Zookeeper: localhost:${PROJ_ZOOKEEPER_PORT:-2181}"  
echo "  🔗 Kafka: localhost:${PROJ_KAFKA_PORT:-9092}"
echo "  📊 JMX监控: localhost:9999"
echo "  🌐 Docker网络: ${NETWORK_NAME}"
echo ""
echo -e "${BLUE}🛠️  快速测试命令:${NC}"
echo "  # 创建Topic"
echo "  docker exec ${KAFKA_CONTAINER} ${KAFKA_TOOLS_PATH}/kafka-topics --create --topic test --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER}"
echo ""  
echo "  # 生产消息"
echo "  docker exec -it ${KAFKA_CONTAINER} ${KAFKA_TOOLS_PATH}/kafka-console-producer --topic test --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER}"
echo ""
echo "  # 消费消息"  
echo "  docker exec -it ${KAFKA_CONTAINER} ${KAFKA_TOOLS_PATH}/kafka-console-consumer --topic test --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER} --from-beginning"
echo ""
echo "  # 停止服务"
echo "  make docker.kafka.stop"