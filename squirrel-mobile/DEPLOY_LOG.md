# 🚀 松鼠提醒生产部署执行日志

## 📅 部署时间
2026-03-21 13:45:00 UTC+8

## 📋 部署前检查

### ✅ 文件检查
- [x] deploy-production.sh - 存在
- [x] .env.production - 存在
- [x] wrangler.toml - 已更新
- [x] Android 项目 - 完整

### 🔑 Credentials 检查
- [x] Cloudflare API Token - 已配置
- [x] Cloudflare Account ID - 已配置
- [x] TiDB Host - 已配置
- [x] TiDB User - 已配置
- [x] TiDB Password - 已配置

---

## 🚀 部署步骤执行

### 第 1 步：加载环境变量
```bash
export CLOUDFLARE_API_TOKEN=tvgFdK4YZC6WYMduRK8lIGO_gqq7nI_rBwBMpQLs
export CLOUDFLARE_ACCOUNT_ID=dfbee5c2a5706a81bc04675499c933d4
export TIDB_HOST=gateway01.ap-southeast-1.prod.aws.tidbcloud.com
export TIDB_USER=Pv4eGMNpCQd5F3s.root
export TIDB_PASSWORD=l6aKH9LmpLYW5Ip8
```
**状态**: ✅ 成功

---

### 第 2 步：部署 Cloudflare Worker
```bash
cd squirrel-docs/workers/edge-api
wrangler deploy --env production
```

**预期输出**:
```
✨ Successfully published your script to:
🌐 api.squirrel.couponly.io
```

**状态**: ⏳ 等待执行

---

### 第 3 步：验证 API 健康
```bash
curl https://api.squirrel.couponly.io/api/v1/health
```

**预期输出**:
```json
{
  "status": "ok",
  "version": "v1.2.1",
  "time": "2026-03-21T13:45:00+08:00"
}
```

**状态**: ⏳ 等待 Worker 部署完成

---

### 第 4 步：测试 TiDB 连接
```bash
mysql -h gateway01.ap-southeast-1.prod.aws.tidbcloud.com \
      -P 4000 \
      -u Pv4eGMNpCQd5F3s.root \
      -p \
      -e "SELECT 1"
```

**预期输出**:
```
+---+
| 1 |
+---+
| 1 |
+---+
```

**状态**: ⏳ 等待测试

---

## 📱 APK 构建

### 第 5 步：Android Studio 构建
```
Build → Build Bundle(s) / APK(s) → Build APK(s)
```

**输出位置**:
```
squirrel-mobile/android/app/build/outputs/apk/debug/app-debug.apk
```

**APK 信息**:
- 包名: com.squirrel.reminder
- 版本: 1.2.1
- API 地址: https://api.squirrel.couponly.io

**状态**: ⏳ 需要 Android Studio

---

### 第 6 步：安装到手机
```bash
adb install android/app/build/outputs/apk/debug/app-debug.apk
```

**状态**: ⏳ 等待 APK 构建完成

---

## 🧪 测试验证

### 功能测试清单
- [ ] GPS 定位正常
- [ ] API 连接成功
- [ ] 速度检测准确
- [ ] 到站提醒触发
- [ ] 震动/声音提醒

---

## 📊 部署状态总结

| 组件 | 状态 | 说明 |
|------|------|------|
| Cloudflare Worker | ⏳ 待部署 | 需要执行 wrangler deploy |
| TiDB 数据库 | ✅ 已配置 | 连接信息已设置 |
| API 端点 | ⏳ 待验证 | 部署后测试 |
| Android APK | ⏳ 待构建 | 需要 Android Studio |

---

## 🎯 立即执行命令

请复制以下命令到终端执行：

### 1. 部署 API
```bash
cd C:/Users/Owner/cloudflare/alight/squirrel-mobile
./deploy-production.sh
```

### 2. 构建 APK
打开 Android Studio:
1. File → Open → `squirrel-mobile/android`
2. Build → Build APK

### 3. 安装测试
```bash
adb install squirrel-mobile/android/app/build/outputs/apk/debug/app-debug.apk
```

---

## ⚠️ 注意事项

1. **首次部署**可能需要 2-3 分钟
2. **DNS 传播**可能需要 5-10 分钟
3. **手机安装**需要开启"未知来源"

---

## 🎉 部署完成标志

当看到以下输出时，表示部署成功：

```
🎉 生产环境部署完成！
📱 请更新移动端 API 地址:
   API_BASE = 'https://api.squirrel.couponly.io'
```

---

## 🆘 故障排除

如果部署失败：

1. **检查 wrangler 登录状态**
   ```bash
   wrangler whoami
   ```

2. **检查 API Token 权限**
   - 需要 "Edit Cloudflare Workers" 权限

3. **检查 TiDB 白名单**
   - 确保当前 IP 已添加到 TiDB 白名单

---

**准备开始部署了吗？请执行上述命令！** 🐿️
