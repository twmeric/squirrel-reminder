# 🐿️ 松鼠提醒 - 部署状态与运行指南

> 更新时间: 2026-03-21

## 📊 基础设施状态

### 生产环境 (Production)

| 服务 | 提供商 | 区域 | 状态 | 说明 |
|------|--------|------|------|------|
| **Cloudflare Workers** | Cloudflare | 全球 | ⚠️ **待配置** | wrangler.toml 已创建，KV/R2 需手动配置 |
| **TiDB Cloud** | PingCAP | 新加坡 | ⚠️ **待创建** | 需要注册 TiDB Cloud 并创建集群 |
| **腾讯云 TKE** | 腾讯云 | 香港 | ⚠️ **待创建** | K8s 集群配置已准备 |
| **Redis** | 腾讯云 | 香港 | ⚠️ **待创建** | 可与 TKE 同区域 |

### 本地开发环境 (Local Dev)

| 服务 | 方式 | 状态 | 说明 |
|------|------|------|------|
| **TiDB** | Docker | ✅ **可用** | `docker-compose up tidb` 即可启动 |
| **Redis** | Docker | ✅ **可用** | 已包含在 docker-compose 中 |
| **后端服务** | Docker | ✅ **可用** | M03/M04 可通过 Docker 启动 |

---

## 🚀 立即运行（本地开发版）

### 步骤 1: 启动基础设施

```bash
# 进入服务目录
cd squirrel-services

# 启动 TiDB + Redis + 监控
docker-compose up -d tidb redis prometheus grafana jaeger

# 等待 30 秒让数据库启动
sleep 30
```

**启动后访问**:
- TiDB: `localhost:4000` (MySQL 协议，用户名 root，密码空)
- Redis: `localhost:6379`
- Grafana: `http://localhost:3000` (admin/admin)
- Prometheus: `http://localhost:9090`
- Jaeger: `http://localhost:16686`

### 步骤 2: 初始化数据库

```bash
# 运行数据库迁移
# 将 TIDB_DSN 设置为本地 TiDB
cd services/m02-alert-service/go
TIDB_DSN="root@tcp(localhost:4000)/squirrel" go run migrations/migrate.go
```

### 步骤 3: 启动后端服务

**方式 A: Docker（推荐）**
```bash
# 在 squirrel-services 目录
docker-compose up -d m03-trajectory m04-insight

# 查看日志
docker-compose logs -f m03-trajectory
```

**方式 B: 本地运行（开发调试）**
```bash
# 终端 1: 启动 M03
export TIDB_DSN="root@tcp(localhost:4000)/squirrel"
export REDIS_HOST=localhost
cd services/m03-trajectory
go run cmd/server/main.go

# 终端 2: 启动 M04
export TIDB_DSN="root@tcp(localhost:4000)/squirrel"
export REDIS_HOST=localhost
export M03_ENDPOINT=localhost:50053
cd services/m04-insight-engine
go run src/api/main.go
```

### 步骤 4: 验证服务

```bash
# 检查 M03 健康
curl http://localhost:8083/health

# 检查 M04 健康  
curl http://localhost:8084/health

# 测试轨迹处理
curl -X POST http://localhost:8083/api/v1/trajectory \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test123", "points": [{"lat": 22.3, "lng": 114.1, "timestamp": "2026-03-21T10:00:00Z"}]}'
```

---

## ☁️ 生产环境部署

### Cloudflare 配置（边缘层）

```bash
cd squirrel-docs/workers/edge-api

# 1. 登录 Cloudflare
npx wrangler login

# 2. 创建 KV 命名空间
npx wrangler kv:namespace create "CACHE"
# 记录下返回的 ID

# 3. 创建 R2 Bucket
npx wrangler r2 bucket create squirrel-data

# 4. 更新 wrangler.toml，填入上面获取的 ID
# [[env.production.kv_namespaces]]
# binding = "CACHE"
# id = "your-kv-namespace-id"

# 5. 部署
npx wrangler deploy --env production
```

### TiDB Cloud 配置（数据层）

```bash
# 1. 访问 https://tidbcloud.com 注册账号
# 2. 创建 Serverless Tier 集群（免费）
# 3. 获取连接字符串，格式如下：
TIDB_DSN="user:password@tcp(gateway.XXX.aws.tidbcloud.com:4000)/squirrel?tls=true"

# 4. 在连接前需要配置白名单（你的服务器 IP）
```

### 腾讯云 TKE 配置（计算层）

```bash
# 1. 登录腾讯云控制台
# 2. 创建 TKE 标准集群（香港区域）
# 3. 配置 kubectl
# 4. 部署服务
cd squirrel-services/k8s
kubectl apply -f namespace.yaml
kubectl apply -f m03-deployment.yaml
kubectl apply -f m04-deployment.yaml
```

---

## 📱 移动端测试

### 方案 1: Web 版本（最快）

```bash
# 启动服务后，直接用浏览器访问
# 电脑端:
open http://localhost:8080

# 手机端（同一 WiFi 下）:
# 找到电脑 IP: ifconfig/ipconfig
# 手机浏览器访问: http://<电脑IP>:8080
```

### 方案 2: 构建 APK

目前项目中**没有现成的移动端代码**，需要：

1. **检查是否有单独的移动仓库**
   ```bash
   # 询问项目维护者
   # 移动端可能在:
   # - squirrel-mobile (React Native)
   # - squirrel-flutter (Flutter)
   # - squirrel-android (原生)
   ```

2. **或使用 PWA 方案**
   ```bash
   # 在 Web 端添加 manifest.json 和 service worker
   # 手机浏览器"添加到主屏幕"即可像 App 使用
   ```

---

## ⚠️ 常见问题

### Q: TiDB 启动失败？
```bash
# 检查端口占用
lsof -i :4000

# 清理旧数据重新启动
docker-compose down -v
docker-compose up -d tidb
```

### Q: 服务连接不上 TiDB？
```bash
# 检查 DSN 格式
# 本地开发: root@tcp(localhost:4000)/squirrel
# TiDB Cloud: user@tcp(host:4000)/db?tls=true

# 测试连接
mysql -h localhost -P 4000 -u root
```

### Q: Cloudflare Workers 部署失败？
```bash
# 确保已登录
npx wrangler whoami

# 检查 account_id
# 在 wrangler.toml 或 ~/.wrangler/config.toml 中配置
```

---

## 🎯 总结

| 环境 | 状态 | 操作 |
|------|------|------|
| **本地开发** | ✅ 可用 | `docker-compose up` 即可运行 |
| **生产 Cloudflare** | ⚠️ 需配置 | 需创建 KV/R2 并部署 |
| **生产 TiDB** | ⚠️ 需创建 | 需注册 TiDB Cloud |
| **移动端 APK** | ❌ 暂无 | 需要单独的移动仓库 |

---

**💡 建议**: 先运行本地 Docker 版本测试功能，生产环境配置需要您提供 Cloudflare 和 TiDB 账号。

🐿️ **有任何问题请告诉我！**
