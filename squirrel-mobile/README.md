# 🐿️ 松鼠提醒 Mobile MVP

立即可以試用的移動端 MVP！

## 🚀 快速开始

### 方式一：Docker（推荐）

```bash
# 1. 进入目录
cd squirrel-mobile

# 2. 启动服务
docker-compose up -d

# 3. 等待 10 秒后访问
# 手机或电脑浏览器打开: http://localhost:8080
```

### 方式二：本地运行

```bash
# 1. 启动 Go 服务器
go run server.go

# 2. 浏览器访问
# http://localhost:8080
```

## 📱 如何使用

### 1. 访问应用
- 电脑：打开 `http://localhost:8080`
- 手机：确保手机和电脑在同一 WiFi，然后访问 `http://<电脑IP>:8080`

### 2. 设置路线
- 输入起点站（如：羅湖）
- 输入终点站（如：福田）
- 点击"🚀 開始追蹤"

### 3. 授予权限
- 允许位置权限（始终允许）
- 允许通知权限

### 4. 开始使用
- 上地铁后点击"🚀 開始追蹤"
- App 会在后台追踪位置
- 接近到站时会震动+声音提醒

## 🧪 测试功能

| 功能 | 状态 | 说明 |
|------|------|------|
| GPS 定位 | ✅ | 实时获取位置 |
| 速度检测 | ✅ | 自动判断是否在地鐵上 |
| 到站提醒 | ✅ | 震动 + 声音 + 视觉提醒 |
| 路线设置 | ✅ | 支持自定义起终点 |
| 后台运行 | ⚠️ | PWA 版本支持 |

## 🌐 API 端点

```
GET  /api/v1/health      - 健康检查
POST /api/v1/location    - 上报位置
GET  /api/v1/route       - 获取路线列表
```

## 🔧 故障排除

### 无法访问 localhost:8080
```bash
# 检查服务是否运行
docker-compose ps

# 查看日志
docker-compose logs squirrel-mobile
```

### 手机无法访问
```bash
# 1. 确保手机和电脑在同一 WiFi
# 2. 找到电脑 IP
ifconfig  # Mac/Linux
ipconfig  # Windows

# 3. 手机访问 http://<电脑IP>:8080
```

### GPS 定位不准
- 确保在户外或靠近窗户
- 检查浏览器位置权限是否开启

## 📝 生产部署

当前是 MVP 版本，如需生产部署：
1. 配置 Cloudflare Workers
2. 配置 TiDB Cloud
3. 配置腾讯云 TKE
4. 使用 HTTPS 域名

详细配置请看项目根目录的 DEPLOYMENT_STATUS.md

## 🎉 开始使用

```bash
# 一键启动
docker-compose up -d

# 然后打开浏览器访问 http://localhost:8080
```

祝您測試順利！🚇
