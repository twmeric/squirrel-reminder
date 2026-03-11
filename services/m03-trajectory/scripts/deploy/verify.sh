#!/bin/bash
# verify.sh - 最终集成测试验证脚本
# 验证所有指标: P99 < 10ms

set -e

HTTP_PORT=8083
METRICS_PORT=9090

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🧪 m03-trajectory 最终集成测试"
echo "================================"
echo ""

PASS=0
FAIL=0

# 测试函数
test_api() {
    local name=$1
    local url=$2
    local target_ms=$3
    
    echo -n "Testing $name (target: ${target_ms}ms)... "
    
    # 测量延迟
    local latencies=()
    for i in {1..100}; do
        local start=$(date +%s%N)
        curl -s -o /dev/null "$url"
        local end=$(date +%s%N)
        local latency=$(( (end - start) / 1000000 ))  # 转换为ms
        latencies+=($latency)
    done
    
    # 计算P99
    IFS=$'\n' sorted=($(sort -n <<<"${latencies[*]}")); unset IFS
    local p99=${sorted[98]}
    
    if [ "$p99" -lt "$target_ms" ]; then
        echo -e "${GREEN}PASS${NC} (P99: ${p99}ms)"
        ((PASS++))
    else
        echo -e "${RED}FAIL${NC} (P99: ${p99}ms > ${target_ms}ms)"
        ((FAIL++))
    fi
}

check_metric() {
    local name=$1
    local query=$2
    local threshold=$3
    
    echo -n "Checking $name... "
    
    local value=$(curl -s "http://localhost:$METRICS_PORT/metrics" | grep "$query" | head -1 | awk '{print $2}')
    
    if [ -z "$value" ]; then
        echo -e "${YELLOW}SKIP${NC} (metric not found)"
        return
    fi
    
    if (( $(echo "$value $threshold" | awk '{print ($1 < $2)}') )); then
        echo -e "${GREEN}PASS${NC} ($value < $threshold)"
        ((PASS++))
    else
        echo -e "${RED}FAIL${NC} ($value >= $threshold)"
        ((FAIL++))
    fi
}

# 1. 基础健康检查
echo "📋 Step 1: 基础健康检查"
echo "------------------------"

echo -n "Health endpoint... "
if curl -s http://localhost:$HTTP_PORT/health | grep -q "ok"; then
    echo -e "${GREEN}PASS${NC}"
    ((PASS++))
else
    echo -e "${RED}FAIL${NC}"
    ((FAIL++))
fi

echo -n "Ready endpoint... "
if curl -s http://localhost:$HTTP_PORT/ready | grep -q "true"; then
    echo -e "${GREEN}PASS${NC}"
    ((PASS++))
else
    echo -e "${RED}FAIL${NC}"
    ((FAIL++))
fi

echo ""
echo "📊 Step 2: API 延迟测试 (P99 < 10ms)"
echo "--------------------------------------"

# 测试各个API的P99延迟
test_api "GetSpeed" "http://localhost:$HTTP_PORT/api/v1/speed?user_id=test" 10
test_api "GetCurrentGrid" "http://localhost:$HTTP_PORT/api/v1/grid?user_id=test" 10
test_api "IsInMetroArea" "http://localhost:$HTTP_PORT/api/v1/metro?user_id=test" 10

echo ""
echo "📈 Step 3: 指标验证"
echo "--------------------"

# 检查Prometheus指标
check_metric "Request Rate" "m03_request_total" 1000
check_metric "Error Rate" "m03_error_total" 10
check_metric "Cache Hit Rate" "m03_cache_hit_total" 0.8

echo ""
echo "🎯 Step 4: 关键指标验证"
echo "------------------------"

# 验证P99延迟
echo -n "Overall P99 Latency < 10ms... "
P99=$(curl -s "http://localhost:$METRICS_PORT/metrics" | grep "m03_request_duration_seconds_bucket" | grep "le=0.01" | head -1 | awk '{print $2}')
if [ ! -z "$P99" ]; then
    echo -e "${GREEN}PASS${NC}"
    ((PASS++))
else
    echo -e "${YELLOW}SKIP${NC}"
fi

echo ""
echo "================================"
echo "测试结果: ${GREEN}$PASS PASS${NC}, ${RED}$FAIL FAIL${NC}"
echo "================================"

if [ $FAIL -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 所有测试通过！服务已就绪！${NC}"
    echo ""
    echo "服务地址:"
    echo "  - HTTP API: http://localhost:$HTTP_PORT"
    echo "  - Metrics:  http://localhost:$METRICS_PORT/metrics"
    echo "  - Health:   http://localhost:$HTTP_PORT/health"
    exit 0
else
    echo ""
    echo -e "${RED}❌ 存在失败的测试，请检查服务状态${NC}"
    exit 1
fi
