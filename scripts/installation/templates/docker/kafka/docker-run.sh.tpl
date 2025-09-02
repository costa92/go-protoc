#!/usr/bin/env bash

# Kafka Stack启动脚本
# Project: ${PROJ_NAME:-go-protoc}
# Service: Kafka Stack (先启动Zookeeper，再启动Kafka)
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 颜色定义
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

echo -e "${BLUE}🚀 启动Kafka技术栈 (Zookeeper + Kafka)${NC}"
echo "=================================="

# 获取脚本所在目录  
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 第一步：启动Zookeeper
echo -e "${BLUE}📋 第1步: 启动Zookeeper服务${NC}"

if [[ -f "$SCRIPT_DIR/zookeeper-run.sh" ]]; then
    chmod +x "$SCRIPT_DIR/zookeeper-run.sh"
    if "$SCRIPT_DIR/zookeeper-run.sh"; then
        echo -e "${GREEN}✅ Zookeeper启动完成${NC}"
    else
        echo -e "${RED}❌ Zookeeper启动失败${NC}"
        exit 1
    fi
else
    echo -e "${RED}❌ 找不到Zookeeper启动脚本: $SCRIPT_DIR/zookeeper-run.sh${NC}"
    exit 1
fi

# 第二步：启动Kafka
echo ""
echo -e "${BLUE}📋 第2步: 启动Kafka服务${NC}"

if [[ -f "$SCRIPT_DIR/kafka-only-run.sh" ]]; then
    chmod +x "$SCRIPT_DIR/kafka-only-run.sh"
    if "$SCRIPT_DIR/kafka-only-run.sh"; then
        echo -e "${GREEN}✅ Kafka启动完成${NC}"
    else
        echo -e "${RED}❌ Kafka启动失败${NC}"
        exit 1
    fi
else
    echo -e "${RED}❌ 找不到Kafka启动脚本: $SCRIPT_DIR/kafka-only-run.sh${NC}"
    exit 1
fi

# 第三步：验证整体服务状态
echo ""
echo -e "${BLUE}📋 第3步: 验证服务状态${NC}"

# 检查容器状态
echo "容器状态:"
docker ps --filter name="${PROJ_PREFIX}-zookeeper" --filter name="${PROJ_PREFIX}-kafka" \
    --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo -e "${GREEN}🎉 Kafka技术栈启动完成!${NC}"
echo "=================================="
echo -e "${BLUE}📊 服务信息:${NC}"
echo "  🔗 Zookeeper: localhost:${PROJ_ZOOKEEPER_PORT:-2181}"  
echo "  🔗 Kafka: localhost:${PROJ_KAFKA_PORT:-9092}"
echo "  📊 JMX监控: localhost:9999"
echo "  🌐 Docker网络: ${PROJ_NETWORK_NAME:-proj-network}"
echo ""
echo -e "${BLUE}🛠️  快速测试命令:${NC}"
echo "  # 创建Topic"
echo "  docker exec ${PROJ_PREFIX}-kafka /bin/kafka-topics --create --topic test --bootstrap-server localhost:9092"
echo ""  
echo "  # 生产消息"
echo "  docker exec -it ${PROJ_PREFIX}-kafka /bin/kafka-console-producer --topic test --bootstrap-server localhost:9092"
echo ""
echo "  # 消费消息"  
echo "  docker exec -it ${PROJ_PREFIX}-kafka /bin/kafka-console-consumer --topic test --bootstrap-server localhost:9092 --from-beginning"