# 🐿️ 松鼠提醒 - 部署就绪

## ✅ Credentials 已找到！

找到了您之前提供的 credentials，保存在 `squirrel-docs/.env` 文件中。

已为您创建：
- ✅ 完整 Android 项目 (Kotlin)
- ✅ 生产环境配置
- ✅ 部署脚本
- ✅ 生产版 Web 界面

---

## 🚀 快速开始 (3 步)

### 第 1 步：部署生产 API

```bash
cd squirrel-mobile
./deploy-production.sh
```

或手动：
```bash
cd squirrel-docs/workers/edge-api
wrangler deploy --env production
```

### 第 2 步：构建 APK

```bash
# 1. 打开 Android Studio
# 2. File → Open → squirrel-mobile/android
# 3. Build → Build APK
# 4. 获取 app-debug.apk
```

### 第 3 步：安装测试

```bash
adb install android/app/build/outputs/apk/debug/app-debug.apk
```

---

## 📋 项目结构

```
squirrel-mobile/
├── android/                    # 完整 Android 项目
│   └── app/src/main/assets/index.html  (已配置生产API)
├── index.production.html       # 生产版 Web
├── deploy-production.sh        # 部署脚本
└── DEPLOY_AND_BUILD.md         # 详细指南
```

---

## 🎯 API 地址

```
生产环境: https://api.squirrel.couponly.io
```

---

## 🔑 Credentials 位置

```
squirrel-docs/.env  (原始)
.env.production     (生产配置)
```

---

## ⚠️ 安全提醒

**请勿将 `.env` 文件提交到 Git！**

已添加到 `.gitignore`:
```
.env
.env.production
*.keystore
```

---

## 🎉 立即开始

运行以下命令开始部署：

```bash
cd squirrel-mobile
./deploy-production.sh
```

然后构建 APK，安装到手机即可测试！

🐿️ **一切准备就绪！**
