#!/bin/bash
# prod_verify.sh - 生产环境部署验证脚本

set -e

PROD_HOST="api.squirrel.couponly.io"
METRICS_URL="https://${PROD_HOST}/metrics"
HEALTH_URL="https://${PROD_HOST}/health"

echo "🚀 生产环境部署验证"
echo "==================="
echo ""
echo "目标环境: ${PROD_HOST}"
echo ""

PASS=0
FAIL=0

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

check_service() {
    local name=$1
    local url=$2
    local expected=$3
    
    echo -n "Checking $name... "
    
    response=$(curl -s -k "$url" 2>/dev/null || echo "ERROR")
    
    if echo "$response" | grep -q "$expected"; then
        echo -e "${GREEN}✅ PASS${NC}"
        ((PASS++))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC}"
        echo "  Expected: $expected"
        echo "  Actual: $response"
        ((FAIL++))
        return 1
    fi
}

echo "Step 1: 服务状态检查"
echo "--------------------"

# 健康检查
check_service "Health Endpoint" "$HEALTH_URL" "ok"

# 就绪检查
check_service "Readiness Check" "$HEALTH_URL/ready" "true"

echo ""
echo "Step 2: 监控指标检查"
echo "--------------------"

echo -n "Fetching metrics... "
metrics=$(curl -s -k "$METRICS_URL" 2>/dev/null || echo "ERROR")

if [ "$metrics" = "ERROR" ]; then
    echo -e "${RED}❌ FAIL${NC} - Cannot fetch metrics"
    ((FAIL++))
else
    echo -e "${GREEN}✅ PASS${NC}"
    ((PASS++))
    
    # 检查关键指标
    echo ""
    echo "关键指标:"
    
    # 请求总数
    req_total=$(echo "$metrics" | grep "m03_request_total" | tail -1 | awk '{print $2}')
    if [ ! -z "$req_total" ]; then
        echo "  m03_request_total: $req_total"
    fi
    
    # 缓存命中率
    cache_hits=$(echo "$metrics" | grep "m03_cache_hit_total" | tail -1 | awk '{print $2}')
    cache_misses=$(echo "$metrics" | grep "m03_cache_miss_total" | tail -1 | awk '{print $2}')
    if [ ! -z "$cache_hits" ] && [ ! -z "$cache_misses" ]; then
        hit_rate=$(echo "scale=2; $cache_hits / ($cache_hits + $cache_misses)" | bc 2>/dev/null || echo "N/A")
        echo "  cache_hit_rate: ${hit_rate}"
    fi
    
    # 活跃连接
    connections=$(echo "$metrics" | grep "m03_active_connections" | tail -1 | awk '{print $2}')
    if [ ! -z "$connections" ]; then
        echo "  active_connections: $connections"
    fi
fi

echo ""
echo "Step 3: 性能指标验证"
echo "--------------------"

# 检查P99延迟
p99_latency=$(echo "$metrics" | grep "m03_request_duration_seconds_bucket" | grep "0.99" | tail -1 | awk '{print $2}')

if [ ! -z "$p99_latency" ]; then
    # 转换为毫秒
    p99_ms=$(echo "scale=1; $p99_latency * 1000" | bc 2>/dev/null || echo "0")
    
    echo -n "P99 Latency (${p99_ms}ms < 10ms)... "
    
    if (( $(echo "$p99_latency < 0.01" | bc -l 2>/dev/null || echo 0) )); then
        echo -e "${GREEN}✅ PASS${NC}"
        ((PASS++))
    else
        echo -e "${RED}❌ FAIL${NC}"
        ((FAIL++))
    fi
else
    echo -e "${YELLOW}⚠️  SKIP${NC} - P99 metric not found"
fi

echo ""
echo "Step 4: 错误率检查"
echo "------------------"

error_count=$(echo "$metrics" | grep "m03_error_total" | tail -1 | awk '{print $2}')

if [ ! -z "$error_count" ]; then
    echo -n "Error count ($error_count < 100)... "
    
    if [ "$error_count" -lt 100 ] 2>/dev/null; then
        echo -e "${GREEN}✅ PASS${NC}"
        ((PASS++))
    else
        echo -e "${RED}❌ FAIL${NC}"
        ((FAIL++))
    fi
else
    echo -e "${YELLOW}⚠️  SKIP${NC} - Error metric not found"
fi

echo ""
echo "==================="
echo "验证结果: $PASS PASS, $FAIL FAIL"
echo "==================="

if [ $FAIL -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 生产环境验证通过！${NC}"
    echo ""
    echo "服务地址:"
    echo "  - API: https://${PROD_HOST}"
    echo "  - Metrics: ${METRICS_URL}"
    echo "  - Health: ${HEALTH_URL}"
    exit 0
else
    echo ""
    echo -e "${RED}❌ 生产环境验证失败，请检查配置${NC}"
    exit 1
fi
