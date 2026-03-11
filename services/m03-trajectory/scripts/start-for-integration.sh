#!/bin/bash
# start-for-integration.sh - 联调启动脚本

set -e

echo "🚀 启动 m03-trajectory 联调环境"
echo "================================"
echo ""

cd "$(dirname "$0")/.."

# 1. 启动基础设施
echo "📦 步骤1: 启动 TiDB + Redis"
docker-compose up -d tidb redis

# 2. 等待服务就绪
echo ""
echo "⏳ 步骤2: 等待数据库就绪 (约10秒)..."
sleep 10

# 3. 启动 m03 服务
echo ""
echo "🖥️  步骤3: 启动 m03 服务"
go run cmd/server/main.go &
M03_PID=$!
echo "  m03 PID: $M03_PID"

# 4. 等待服务就绪
echo ""
echo "⏳ 步骤4: 等待 m03 服务就绪 (约5秒)..."
sleep 5

# 5. 最终检查
echo ""
echo "🔍 步骤5: 执行联调前检查"
./scripts/pre-integration-check.sh

if [ $? -eq 0 ]; then
    echo ""
    echo "================================"
    echo "🎉 m03 联调环境启动成功！"
    echo ""
    echo "服务信息:"
    echo "  - HTTP: http://localhost:8083/health"
    echo "  - gRPC: localhost:50053"
    echo ""
    echo "停止命令: kill $M03_PID"
    echo ""
    echo "🐿️  准备好参加 16:00 联调会议！"
else
    echo ""
    echo "❌ 启动失败"
    exit 1
fi
