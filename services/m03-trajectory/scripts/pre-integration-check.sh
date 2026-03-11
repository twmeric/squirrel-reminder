#!/bin/bash
# pre-integration-check.sh - 联调前检查脚本

echo "🐿️  m03-trajectory 联调前检查"
echo "================================"
echo ""

CHECK_PASS=0
CHECK_FAIL=0

# 检查端口
check_port() {
    local name=$1
    local port=$2
    
    echo "检查 $name ($port) ..."
    
    if netstat -an | grep -q ":$port "; then
        echo "  ✅ 监听中"
        ((CHECK_PASS++))
    else
        echo "  ❌ 未监听"
        ((CHECK_FAIL++))
    fi
}

# 检查HTTP
check_http() {
    echo "检查 m03 HTTP (8083) ..."
    
    response=$(curl -s http://localhost:8083/health 2>/dev/null)
    
    if echo "$response" | grep -q "ok"; then
        echo "  ✅ 健康检查通过"
        ((CHECK_PASS++))
    else
        echo "  ❌ 健康检查失败"
        ((CHECK_FAIL++))
    fi
}

echo "1️⃣  检查依赖服务"
echo "----------------"
check_port "TiDB-SQL" 4000
check_port "TiDB-Status" 10080
check_port "Redis" 6379

echo ""
echo "2️⃣  检查 m03 服务"
echo "----------------"
check_http
check_port "m03-gRPC" 50053

echo ""
echo "================================"
echo "检查结果: $CHECK_PASS 通过, $CHECK_FAIL 失败"
echo ""

if [ $CHECK_FAIL -eq 0 ]; then
    echo "🎉 所有检查通过，可以开始联调！"
    exit 0
else
    echo "⚠️  存在失败的检查项"
    echo ""
    echo "修复命令:"
    echo "  docker-compose up -d tidb redis"
    echo "  go run cmd/server/main.go"
    exit 1
fi
