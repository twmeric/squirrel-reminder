# 🐿️ 松鼠提醒 - APK 构建完整指南

## ✅ Cloudflare Worker 已部署成功！

**API 地址**: `https://api.squirrel.couponly.io`

---

## 🎯 方案选择

根据您的情况，选择以下方案之一：

---

## 方案 A: GitHub 自动构建 (推荐 ⭐⭐⭐)

**最简单，无需安装任何软件！**

### 步骤:

1. **创建 GitHub 账号**
   - 访问 https://github.com/signup

2. **创建新仓库**
   ```
   仓库名: squirrel-reminder
   公开/私有: 都可以
   ```

3. **上传代码**
   ```bash
   cd C:/Users/Owner/cloudflare/alight
   git init
   git add .
   git commit -m "Initial commit"
   git remote add origin https://github.com/您的用户名/squirrel-reminder.git
   git push -u origin main
   ```

4. **自动构建**
   - 推送后 GitHub 会自动构建 APK
   - 等待 5-10 分钟
   - 在 Actions 标签页下载 APK

---

## 方案 B: 一键脚本 (需要 Java)

### 步骤 1: 安装 Java JDK 17
1. 访问 https://adoptium.net/
2. 下载 **Windows x64 MSI Installer**
3. 双击安装，一直点 **Next**

### 步骤 2: 运行构建脚本
```powershell
cd C:/Users/Owner/cloudflare/alight/squirrel-mobile
./build-apk-simple.bat
```

**APK 将生成在**:
```
squirrel-mobile/android/app/build/outputs/apk/debug/app-debug.apk
```

---

## 方案 C: Web App (无需安装 ⭐⭐)

**直接在手机上使用！**

### 使用方法:

1. **手机浏览器访问**:
   ```
   https://api.squirrel.couponly.io/app
   ```

2. **添加到主屏幕**:
   - **Android Chrome**: 菜单 → "添加到主屏幕"
   - **iPhone Safari**: 分享 → "添加到主屏幕"

3. **像原生 App 一样使用**

---

## 方案 D: Docker 构建 (需要 Docker)

```powershell
cd C:/Users/Owner/cloudflare/alight/squirrel-mobile
docker run --rm -v ${PWD}:/project -w /project/android \
  runmymind/docker-android-sdk \
  ./gradlew assembleDebug
```

---

## 📱 安装 APK

构建完成后，安装到手机：

### 方式 1: USB 连接
```powershell
adb install app-debug.apk
```

### 方式 2: 手动安装
1. 将 APK 文件复制到手机
2. 在手机上找到 APK 文件
3. 点击安装
4. 允许"未知来源"安装

---

## 🎯 推荐选择

| 方案 | 难度 | 时间 | 推荐度 |
|------|------|------|--------|
| **A. GitHub 自动构建** | ⭐ 简单 | 10分钟 | ⭐⭐⭐ |
| **B. 一键脚本** | ⭐⭐ 中等 | 5分钟 | ⭐⭐ |
| **C. Web App** | ⭐ 最简单 | 1分钟 | ⭐⭐ |
| **D. Docker** | ⭐⭐⭐ 复杂 | 10分钟 | ⭐ |

---

## 🚀 立即开始

### 最快方案 - Web App (1分钟)

**现在立即用手机访问**:
```
https://api.squirrel.couponly.io/app
```

然后添加到主屏幕即可使用！

---

### 最佳方案 - GitHub 自动构建 (10分钟)

1. 注册 GitHub 账号
2. 创建仓库
3. 上传代码
4. 等待自动构建
5. 下载 APK

---

## 🆘 需要帮助？

如果您：
- 不会使用 GitHub
- 无法安装 Java
- 没有 Docker

请告诉我，我可以：
1. **远程协助** (TeamViewer/AnyDesk)
2. **创建预签名 APK**
3. **优化 Web App 体验**

---

## ✅ 检查清单

部署状态:
- [x] Cloudflare Worker 部署
- [x] API 端点配置
- [x] TiDB 数据库连接
- [ ] APK 构建
- [ ] 安装测试

---

## 📞 下一步

**请选择您想使用的方案，我立即指导您操作！**

🐿️ **推荐先试试 Web App，如果满意再构建原生 APK！**
