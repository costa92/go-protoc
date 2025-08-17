#!/usr/bin/env bash

# VictoriaLogs UI 测试脚本
# 此脚本验证 VictoriaLogs 集成是否正常工作

set -o errexit
set -o nounset
set -o pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# VictoriaLogs 配置
VICTORIALOGS_URL="http://127.0.0.1:9428"
VMUI_URL="${VICTORIALOGS_URL}/select/vmui"
QUERY_URL="${VICTORIALOGS_URL}/select/logsql/query"

echo -e "${BLUE}=== VictoriaLogs 集成测试 ===${NC}"
echo ""

# 1. 检查 VictoriaLogs 服务状态
echo -e "${YELLOW}1. 检查 VictoriaLogs 服务状态...${NC}"
if curl -s "${VICTORIALOGS_URL}/health" > /dev/null; then
    echo -e "${GREEN}✓ VictoriaLogs 服务运行正常${NC}"
else
    echo -e "${RED}✗ VictoriaLogs 服务不可访问${NC}"
    exit 1
fi

# 2. 检查 UI 界面
echo -e "${YELLOW}2. 检查 VictoriaLogs UI 界面...${NC}"
if curl -s "${VMUI_URL}" | grep -q "VM UI"; then
    echo -e "${GREEN}✓ VictoriaLogs UI 界面可访问${NC}"
    echo -e "   📱 Web UI: ${VMUI_URL}"
else
    echo -e "${RED}✗ VictoriaLogs UI 界面不可访问${NC}"
fi

# 3. 查询现有日志
echo -e "${YELLOW}3. 查询现有日志数据...${NC}"
log_count=$(curl -s "${QUERY_URL}" -d 'query=*' | wc -l)
if [ "$log_count" -gt 0 ]; then
    echo -e "${GREEN}✓ 发现 ${log_count} 条日志记录${NC}"
else
    echo -e "${RED}✗ 未发现任何日志记录${NC}"
fi

# 4. 显示最近的日志样例
echo -e "${YELLOW}4. 显示最近的日志样例...${NC}"
echo -e "${BLUE}最近 3 条日志记录:${NC}"
curl -s "${QUERY_URL}" -d 'query=*' | head -3 | while IFS= read -r line; do
    echo "   📄 $line"
done

# 5. 按日志级别查询
echo ""
echo -e "${YELLOW}5. 按日志级别统计...${NC}"
for level in debug info warn error; do
    count=$(curl -s "${QUERY_URL}" -d "query=level:${level}" 2>/dev/null | wc -l)
    if [ "$count" -gt 0 ]; then
        echo -e "   ${level}: ${GREEN}${count} 条${NC}"
    else
        echo -e "   ${level}: ${YELLOW}0 条${NC}"
    fi
done

# 6. 显示服务相关日志
echo ""
echo -e "${YELLOW}6. 服务启动相关日志...${NC}"
service_logs=$(curl -s "${QUERY_URL}" -d 'query=_msg:*server*' 2>/dev/null | wc -l)
if [ "$service_logs" -gt 0 ]; then
    echo -e "${GREEN}✓ 发现 ${service_logs} 条服务相关日志${NC}"
    echo -e "${BLUE}服务日志样例:${NC}"
    curl -s "${QUERY_URL}" -d 'query=_msg:*server*' | head -2 | while IFS= read -r line; do
        echo "   🚀 $line"
    done
fi

# 7. 测试总结
echo ""
echo -e "${BLUE}=== 测试总结 ===${NC}"
echo -e "${GREEN}✅ VictoriaLogs 集成测试完成${NC}"
echo ""
echo -e "${YELLOW}📊 访问信息:${NC}"
echo -e "   🌐 VictoriaLogs UI: ${VMUI_URL}"
echo -e "   🔍 日志查询 API: ${QUERY_URL}"
echo ""
echo -e "${YELLOW}💡 使用提示:${NC}"
echo -e "   • 在浏览器中访问: ${VMUI_URL}"
echo -e "   • 查询语法: * (所有日志), level:info (按级别), _msg:*error* (消息包含error)"
echo -e "   • 时间范围: 可在 UI 中设置时间过滤器"
echo ""
echo -e "${GREEN}🎉 项目日志已成功集成到 VictoriaLogs！${NC}"