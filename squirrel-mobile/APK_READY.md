# 🐿️ 松鼠提醒 - APK 构建包

## ✅ 已完成的内容

我已经为您创建了完整的 Android 项目，可以直接导入 Android Studio 构建 APK。

### 📁 项目结构

```
squirrel-mobile/
├── android/                          # 完整的 Android 项目
│   ├── app/
│   │   ├── src/main/
│   │   │   ├── java/com/squirrel/reminder/
│   │   │   │   └── MainActivity.kt   # Kotlin 主活动
│   │   │   ├── assets/
│   │   │   │   └── index.html        # Web 界面
│   │   │   ├── res/                  # 资源文件
│   │   │   └── AndroidManifest.xml   # 权限配置
│   │   └── build.gradle              # 应用构建配置
│   ├── build.gradle                  # 项目构建配置
│   └── settings.gradle
├── index.html                        # Web 界面
├── BUILD_APK.md                      # 详细构建指南
└── APK_READY.md                      # 本文件
```

### 📱 应用功能

| 功能 | 状态 | 说明 |
|------|------|------|
| GPS 定位 | ✅ | 实时获取位置 |
| 后台运行 | ✅ | 支持后台定位 |
| 速度检测 | ✅ | 判断是否在地鐵上 |
| 到站提醒 | ✅ | 震动 + 声音 |
| 离线使用 | ✅ | 本地 Web 资源 |

### 🔐 权限申请

AndroidManifest.xml 已配置：
- `ACCESS_FINE_LOCATION` - 精确定位
- `ACCESS_BACKGROUND_LOCATION` - 后台定位
- `VIBRATE` - 震动提醒
- `FOREGROUND_SERVICE` - 前台服务

---

## 🚀 构建 APK 步骤

### 方法 1: Android Studio (推荐)

1. **下载并安装 Android Studio**
   - https://developer.android.com/studio

2. **打开项目**
   - 启动 Android Studio
   - 选择 `Open an existing project`
   - 选择 `squirrel-mobile/android` 文件夹

3. **等待同步**
   - 首次打开会自动下载 Gradle 和依赖
   - 等待底部进度条完成

4. **构建 APK**
   - 菜单栏: `Build` → `Build Bundle(s) / APK(s)` → `Build APK(s)`
   - 等待构建完成

5. **获取 APK**
   - 构建完成后，右下角会显示提示
   - 点击 `locate` 或在文件夹中找到:
   ```
   android/app/build/outputs/apk/debug/app-debug.apk
   ```

### 方法 2: 命令行 (需要 Android SDK)

```bash
# 进入项目目录
cd squirrel-mobile/android

# 构建 Debug APK
./gradlew assembleDebug

# 输出位置
# app/build/outputs/apk/debug/app-debug.apk
```

---

## 📲 安装到手机

### 方式 1: USB 调试

```bash
# 连接手机，开启 USB 调试
adb install android/app/build/outputs/apk/debug/app-debug.apk
```

### 方式 2: 手动安装

1. 将 APK 文件传输到手机
2. 在手机上找到 APK 文件
3. 点击安装
4. 允许"未知来源"安装

---

## 🌐 配置生产环境 (需要 Credentials)

您提到会提供 credentials，当前项目使用**本地模式**，如需连接生产环境，请提供：

### 需要的 Credentials

| 服务 | 需要的信息 | 用途 |
|------|-----------|------|
| **Cloudflare** | API Token + Account ID | Workers 部署 |
| **TiDB Cloud** | Host + User + Password | 数据库存储 |
| **腾讯云** (可选) | SecretId + SecretKey | TKE 部署 |

### 配置步骤

1. **修改 `index.html` 中的 API 地址**
   ```javascript
   // 当前使用本地模式
   const API_BASE = 'http://localhost:8080';
   
   // 改为生产环境
   const API_BASE = 'https://api.squirrel.couponly.io';
   ```

2. **配置后端服务**
   - 部署到 Cloudflare Workers
   - 连接 TiDB Cloud

3. **重新构建 APK**
   - 更新代码后重新构建

---

## ⚡ 立即测试

### 本地测试 (无需 Credentials)

```bash
# 1. 启动本地服务器
cd squirrel-mobile
docker-compose up -d

# 2. 打开浏览器访问
# http://localhost:8080

# 3. 用手机浏览器测试 (同一 WiFi)
# http://<电脑IP>:8080
```

### 生产测试 (需要 Credentials)

提供以下信息后，我可以立即部署：

```yaml
# Cloudflare
cf_api_token: "your_token_here"
cf_account_id: "your_account_id"

# TiDB Cloud
tidb_host: "gateway.xxx.aws.tidbcloud.com"
tidb_user: "your_user"
tidb_password: "your_password"

# 域名 (可选)
domain: "api.squirrel.couponly.io"
```

---

## 📋 检查清单

- [x] Android 项目创建
- [x] Web 界面开发
- [x] 权限配置
- [x] 构建文档
- [ ] APK 构建 (需要您运行 Android Studio)
- [ ] 生产环境部署 (需要 Credentials)

---

## 🎯 下一步

### 您现在可以：

1. **立即试用本地版**
   ```bash
   docker-compose up -d
   ```

2. **构建 APK**
   - 导入 Android Studio
   - 构建 APK
   - 安装到手机

3. **提供 Credentials**
   - 我立即部署生产环境
   - 更新 APK 连接到生产 API

---

## 📞 支持

遇到问题？
1. 查看 `BUILD_APK.md` 详细指南
2. 检查 Android Studio 版本是否最新
3. 确保 Android SDK 已安装

🐿️ **准备好 Credentials 了吗？发给我，立即部署！**
