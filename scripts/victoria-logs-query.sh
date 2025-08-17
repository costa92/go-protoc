#!/usr/bin/env bash

# VictoriaLogs 查询脚本
# 此脚本演示如何查询 VictoriaLogs 中的日志数据

set -o errexit
set -o nounset
set -o pipefail

# 配置
VICTORIALOGS_URL="http://127.0.0.1:9428"
QUERY_URL="${VICTORIALOGS_URL}/select/logsql/query"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 帮助函数
show_help() {
    cat << EOF
VictoriaLogs 查询工具

用法: $0 [查询选项]

查询选项:
  all           显示所有日志
  recent        显示最近 5 条日志
  info          显示 info 级别日志
  server        显示服务器相关日志
  count         显示日志总数
  stats         显示日志统计信息
  "custom_query" 自定义查询

示例:
  $0 all                    # 显示所有日志
  $0 recent                 # 显示最近日志
  $0 "level:info"          # 查询 info 级别日志
  $0 "_msg:*MySQL*"        # 查询包含 MySQL 的日志

EOF
}

# 执行查询
execute_query() {
    local query="$1"
    echo -e "${BLUE}执行查询: ${query}${NC}"
    echo -e "${YELLOW}结果:${NC}"
    curl -s "${QUERY_URL}" -d "query=${query}" | while IFS= read -r line; do
        echo "  $line"
    done
    echo ""
}

# 统计查询
count_query() {
    local query="$1"
    curl -s "${QUERY_URL}" -d "query=${query}" | wc -l
}

# 主逻辑
main() {
    local command=${1:-help}

    case "$command" in
        help|--help|-h)
            show_help
            ;;
        all)
            echo -e "${GREEN}=== 所有日志 ===${NC}"
            execute_query "*"
            ;;
        recent)
            echo -e "${GREEN}=== 最近 5 条日志 ===${NC}"
            curl -s "${QUERY_URL}" -d "query=*" | tail -5 | while IFS= read -r line; do
                echo "  $line"
            done
            echo ""
            ;;
        info)
            echo -e "${GREEN}=== Info 级别日志 ===${NC}"
            execute_query "level:info"
            ;;
        server)
            echo -e "${GREEN}=== 服务器相关日志 ===${NC}"
            execute_query "_msg:*server*"
            ;;
        count)
            echo -e "${GREEN}=== 日志统计 ===${NC}"
            total=$(count_query "*")
            info_count=$(count_query "level:info")
            server_count=$(count_query "_msg:*server*")
            
            echo -e "  总日志数: ${YELLOW}${total}${NC}"
            echo -e "  Info 级别: ${YELLOW}${info_count}${NC}"
            echo -e "  服务器相关: ${YELLOW}${server_count}${NC}"
            echo ""
            ;;
        stats)
            echo -e "${GREEN}=== 详细统计 ===${NC}"
            
            # 按级别统计
            echo -e "${BLUE}按日志级别统计:${NC}"
            for level in debug info warn error; do
                count=$(count_query "level:${level}")
                echo -e "  ${level}: ${YELLOW}${count} 条${NC}"
            done
            
            echo ""
            echo -e "${BLUE}按消息类型统计:${NC}"
            mysql_count=$(count_query "_msg:*MySQL*")
            redis_count=$(count_query "_msg:*Redis*")
            jaeger_count=$(count_query "_msg:*Jaeger*")
            grpc_count=$(count_query "_msg:*gRPC*")
            http_count=$(count_query "_msg:*HTTP*")
            
            echo -e "  MySQL: ${YELLOW}${mysql_count} 条${NC}"
            echo -e "  Redis: ${YELLOW}${redis_count} 条${NC}"
            echo -e "  Jaeger: ${YELLOW}${jaeger_count} 条${NC}"
            echo -e "  gRPC: ${YELLOW}${grpc_count} 条${NC}"
            echo -e "  HTTP: ${YELLOW}${http_count} 条${NC}"
            echo ""
            ;;
        *)
            echo -e "${GREEN}=== 自定义查询 ===${NC}"
            execute_query "$command"
            ;;
    esac
}

# 检查 VictoriaLogs 连通性
if ! curl -s "${VICTORIALOGS_URL}/health" > /dev/null; then
    echo -e "${RED}错误: VictoriaLogs 服务不可访问${NC}"
    echo "请确保 VictoriaLogs 正在运行: make deploy.install.docker.victorialogs"
    exit 1
fi

main "$@"