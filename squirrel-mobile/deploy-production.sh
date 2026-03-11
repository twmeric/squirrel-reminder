#!/bin/bash

# 松鼠提醒生产环境部署脚本
# 使用已保存的 credentials

set -e

echo "🐿️ 松鼠提醒生产环境部署"
echo "=========================="

# 加载环境变量
export $(grep -v '^#' ../.env.production | xargs)

# 1. 部署 Cloudflare Worker
echo "📡 部署 Cloudflare Worker..."
cd ../squirrel-docs/workers/edge-api

# 安装 wrangler (如果未安装)
if ! command -v wrangler &> /dev/null; then
    npm install -g wrangler
fi

# 登录 Cloudflare
echo "$CLOUDFLARE_API_TOKEN" | wrangler login

# 部署 Worker
wrangler deploy --env production

echo "✅ Cloudflare Worker 部署完成"
echo "   API 地址: https://$API_DOMAIN"

# 2. 测试 API
echo ""
echo "🧪 测试 API..."
curl -s "https://$API_DOMAIN/api/v1/health" | jq .

echo ""
echo "🎉 生产环境部署完成！"
echo ""
echo "📱 请更新移动端 API 地址:"
echo "   API_BASE = 'https://$API_DOMAIN'"
