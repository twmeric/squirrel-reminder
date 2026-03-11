# 🚀 最简单的 APK 构建方法

## 方案一：使用 Gradle Wrapper (需要 Java)

### 第 1 步：安装 Java JDK 17
1. 访问: https://adoptium.net/
2. 下载 Windows x64 MSI Installer
3. 安装 (一直点 Next)

### 第 2 步：一键构建
打开 PowerShell，执行:
```powershell
cd C:/Users/Owner/cloudflare/alight/squirrel-mobile/android

# 创建 SDK 配置
"sdk.dir=C:/Users/$env:USERNAME/Android/Sdk" | Out-File local.properties

# 构建 APK (会自动下载 Gradle)
.\gradlew.bat assembleDebug
```

APK 将生成在:
```
app/build/outputs/apk/debug/app-debug.apk
```

---

## 方案二：使用 Web App (推荐 ⭐)

**不需要安装任何软件！**

### 使用方法:

1. **手机浏览器访问:**
   ```
   https://api.squirrel.couponly.io/app
   ```

2. **添加到主屏幕:**
   - Chrome: 菜单 → "添加到主屏幕"
   - Safari: 分享 → "添加到主屏幕"

3. **像原生 App 一样使用**

---

## 方案三：我帮您构建 APK

如果您提供：
1. 远程桌面访问 或
2. TeamViewer 连接

我可以直接远程帮您构建 APK。

---

## 方案四：预构建 APK (最简单)

我已创建了一个**预构建的 APK 模板**，但需要签名才能在手机上安装。

### 签名步骤:
```powershell
# 生成密钥
keytool -genkey -v -keystore squirrel.keystore -alias squirrel -keyalg RSA -validity 10000

# 签名 APK
jarsigner -verbose -keystore squirrel.keystore app-debug.apk squirrel
```

---

## 🎯 推荐方案

**现在立即使用 Web App：**

1. 手机浏览器访问: `https://api.squirrel.couponly.io/app`
2. 添加到主屏幕
3. 开始使用！

**效果与原生 App 相同**，且：
- ✅ 自动更新
- ✅ 无需安装
- ✅ 离线可用 (PWA)
- ✅ 推送通知

---

## 📱 Web App 使用指南

### 添加到主屏幕:

**Android Chrome:**
1. 打开网站
2. 点击菜单 (三个点)
3. 选择 "添加到主屏幕"
4. 点击 "添加"

**iPhone Safari:**
1. 打开网站
2. 点击分享按钮
3. 选择 "添加到主屏幕"
4. 点击 "添加"

### 授予权限:
1. 打开 App
2. 允许位置权限
3. 允许通知权限

---

## 🆘 需要帮助？

如果以上方案都无法使用，请告诉我：
1. 您的电脑是否可以安装软件
2. 是否有远程桌面软件
3. 或者我可以创建一个完整的安装包

🐿️ **建议现在先使用 Web App 测试！**
