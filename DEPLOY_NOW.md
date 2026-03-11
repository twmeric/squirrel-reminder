# 🚀 立即部署 - 执行命令

## ⚡ 请立即执行以下命令

### 第 1 步：打开 PowerShell/Terminal

```powershell
# 进入项目目录
cd C:/Users/Owner/cloudflare/alight
```

---

### 第 2 步：安装 Wrangler (如未安装)

```powershell
npm install -g wrangler
```

---

### 第 3 步：登录 Cloudflare

```powershell
$env:CLOUDFLARE_API_TOKEN="tvgFdK4YZC6WYMduRK8lIGO_gqq7nI_rBwBMpQLs"
echo $env:CLOUDFLARE_API_TOKEN | wrangler login
```

---

### 第 4 步：部署 Worker

```powershell
cd C:/Users/Owner/cloudflare/alight/squirrel-docs/workers/edge-api
wrangler deploy --env production
```

**成功标志**:
```
✨ Successfully published your script to:
🌐 api.squirrel.couponly.io
```

---

### 第 5 步：验证部署

```powershell
curl https://api.squirrel.couponly.io/api/v1/health
```

**预期返回**:
```json
{"status":"ok","version":"v1.2.1"}
```

---

### 第 6 步：打开 Android Studio 构建 APK

1. 启动 **Android Studio**
2. 点击 **File → Open**
3. 选择文件夹: `C:/Users/Owner/cloudflare/alight/squirrel-mobile/android`
4. 等待 Gradle 同步完成 (首次需要 3-5 分钟)
5. 点击菜单: **Build → Build Bundle(s) / APK(s) → Build APK(s)**
6. 等待构建完成

**APK 位置**:
```
C:/Users/Owner/cloudflare/alight/squirrel-mobile/android/app/build/outputs/apk/debug/app-debug.apk
```

---

### 第 7 步：安装到手机

**方式 A: 通过 USB**
```powershell
adb install "C:/Users/Owner/cloudflare/alight/squirrel-mobile/android/app/build/outputs/apk/debug/app-debug.apk"
```

**方式 B: 手动安装**
1. 将 APK 文件复制到手机
2. 在手机上点击安装
3. 允许"未知来源"安装

---

### 第 8 步：测试使用

1. 打开 **松鼠提醒** App
2. 授予 **位置权限** (始终允许)
3. 设置路线: 起点 → 终点
4. 点击 **"开始追踪"**
5. 上地铁测试！

---

## ✅ 成功标志

当您看到以下画面时，表示全部成功：

### API 部署成功
```json
{
  "status": "ok",
  "version": "v1.2.1",
  "time": "2026-03-21T..."
}
```

### App 界面
```
┌─────────────────────┐
│ 🐿️ 松鼠提醒 v1.2.1  │
├─────────────────────┤
│ 📡 GPS 定位         │
│    追蹤中 ✅        │
├─────────────────────┤
│ ☁️ API 連接         │
│    已連接 ✅        │
└─────────────────────┘
```

---

## 🆘 遇到问题？

### 问题 1: wrangler 命令找不到
```powershell
# 解决方案
npm install -g wrangler
```

### 问题 2: 部署失败
```powershell
# 检查是否登录
wrangler whoami

# 重新登录
wrangler login
```

### 问题 3: API 返回 403
等待 5-10 分钟，DNS 传播需要时间。

### 问题 4: APK 安装失败
- 开启手机"允许安装未知应用"
- 检查 APK 是否完整

---

## 🎉 准备开始！

**第一步：复制以下命令立即执行**

```powershell
cd C:/Users/Owner/cloudflare/alight/squirrel-docs/workers/edge-api
wrangler deploy --env production
```

**然后告诉我执行结果！** 🐿️
