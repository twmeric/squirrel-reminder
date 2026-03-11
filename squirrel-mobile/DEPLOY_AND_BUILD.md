# 🐿️ 松鼠提醒 - 生产部署与 APK 构建

## ✅ Credentials 已找到！

在 `squirrel-docs/.env` 文件中找到了您之前提供的 credentials：

- ✅ Cloudflare API Token
- ✅ Cloudflare Account ID
- ✅ TiDB Cloud 连接信息
- ✅ 腾讯云 Secret

---

## 🚀 第一步：部署生产环境

### 方式 A: 自动部署 (推荐)

```bash
cd squirrel-mobile
chmod +x deploy-production.sh
./deploy-production.sh
```

### 方式 B: 手动部署

```bash
# 1. 进入 Worker 目录
cd squirrel-docs/workers/edge-api

# 2. 安装 wrangler
npm install -g wrangler

# 3. 使用已保存的 credentials 登录
wrangler login

# 4. 部署到生产环境
wrangler deploy --env production

# 5. 验证部署
curl https://api.squirrel.couponly.io/api/v1/health
```

---

## 📱 第二步：构建生产版 APK

### 准备工作

1. **下载 Android Studio**
   - https://developer.android.com/studio

2. **打开项目**
   - File → Open → 选择 `squirrel-mobile/android`

3. **等待同步**
   - 首次打开会下载 Gradle 和依赖

### 构建 APK

#### Debug 版本 (测试用)

```
Build → Build Bundle(s) / APK(s) → Build APK(s)
```

输出位置:
```
android/app/build/outputs/apk/debug/app-debug.apk
```

#### Release 版本 (发布用)

```
Build → Generate Signed Bundle / APK → APK
```

需要创建签名密钥:
```bash
keytool -genkey -v -keystore squirrel-reminder.keystore -alias squirrel -keyalg RSA -keysize 2048 -validity 10000
```

---

## 📂 生产环境文件

| 文件 | 说明 |
|------|------|
| `.env.production` | 生产环境配置 |
| `index.production.html` | 生产版 Web 界面 |
| `android/app/src/main/assets/index.html` | 已更新为生产版 |

---

## 🔧 生产环境特性

### API 地址
```javascript
const API_BASE = 'https://api.squirrel.couponly.io';
```

### 功能
- ✅ 连接生产 API
- ✅ TiDB 数据库存储
- ✅ 云端轨迹处理
- ✅ 离线模式支持

---

## 📲 安装 APK

### 方式 1: ADB 安装
```bash
adb install android/app/build/outputs/apk/debug/app-debug.apk
```

### 方式 2: 手动安装
1. 将 APK 传输到手机
2. 允许"未知来源"安装
3. 点击安装

---

## ✅ 检查清单

### 生产部署
- [ ] Cloudflare Worker 部署
- [ ] API 健康检查通过
- [ ] TiDB 连接正常

### APK 构建
- [ ] Android Studio 打开项目
- [ ] Gradle 同步完成
- [ ] APK 构建成功
- [ ] 安装到手机

### 功能测试
- [ ] GPS 定位正常
- [ ] API 连接成功
- [ ] 到站提醒工作
- [ ] 震动/声音提醒

---

## 🎯 立即行动

### 1. 部署生产环境
```bash
cd squirrel-mobile
./deploy-production.sh
```

### 2. 构建 APK
- 打开 Android Studio
- 导入 `squirrel-mobile/android`
- Build → Build APK

### 3. 测试使用
- 安装 APK
- 设置路线
- 坐地铁测试

---

## 🆘 故障排除

### API 连接失败
```bash
# 检查 API 状态
curl https://api.squirrel.couponly.io/api/v1/health

# 如果失败，检查:
# 1. Cloudflare Worker 是否部署
# 2. DNS 是否正确配置
```

### TiDB 连接失败
```bash
# 测试 TiDB 连接
mysql -h gateway01.ap-southeast-1.prod.aws.tidbcloud.com \
      -P 4000 \
      -u Pv4eGMNpCQd5F3s.root \
      -p \
      -e "SELECT 1"
```

---

## 📞 支持

遇到问题？检查：
1. `BUILD_APK.md` - 详细构建指南
2. Cloudflare Dashboard - Worker 状态
3. TiDB Cloud Console - 数据库状态

🐿️ **准备开始部署了吗？**
