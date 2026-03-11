# 松鼠提醒 APK 构建指南

## 📱 前提条件

构建 APK 需要以下工具：

1. **Java JDK 17** 或更高版本
   - 下载: https://adoptium.net/
   
2. **Android Studio** 或 **Android SDK Command Line Tools**
   - 下载: https://developer.android.com/studio

3. **Gradle** (可选，项目已包含 Wrapper)

---

## 🚀 快速构建

### 方法 1: 使用 Android Studio (推荐)

1. 打开 Android Studio
2. 选择 `File` → `Open`
3. 选择 `squirrel-mobile/android` 文件夹
4. 等待 Gradle 同步完成
5. 点击 `Build` → `Build Bundle(s) / APK(s)` → `Build APK(s)`
6. APK 将生成在:
   ```
   android/app/build/outputs/apk/debug/app-debug.apk
   ```

### 方法 2: 使用命令行

```bash
# 进入 Android 项目目录
cd squirrel-mobile/android

# 确保 gradlew 有执行权限 (Linux/Mac)
chmod +x gradlew

# 构建 Debug APK
./gradlew assembleDebug

# 或构建 Release APK (需要签名配置)
./gradlew assembleRelease
```

---

## 📦 APK 输出位置

构建完成后，APK 文件位于：

| 类型 | 路径 |
|------|------|
| Debug | `android/app/build/outputs/apk/debug/app-debug.apk` |
| Release | `android/app/build/outputs/apk/release/app-release-unsigned.apk` |

---

## 🔧 签名 APK (Release)

要发布到应用商店，需要签名：

```bash
# 生成密钥库
keytool -genkey -v -keystore squirrel-reminder.keystore -alias squirrel -keyalg RSA -keysize 2048 -validity 10000

# 签名 APK
jarsigner -verbose -sigalg SHA1withRSA -digestalg SHA1 -keystore squirrel-reminder.keystore app-release-unsigned.apk squirrel

# 优化 APK
zipalign -v 4 app-release-unsigned.apk squirrel-reminder.apk
```

---

## 📋 项目结构

```
squirrel-mobile/
├── android/                    # Android 项目
│   ├── app/
│   │   ├── src/main/
│   │   │   ├── java/com/squirrel/reminder/
│   │   │   │   └── MainActivity.kt    # 主活动
│   │   │   ├── res/                   # 资源文件
│   │   │   ├── assets/
│   │   │   │   └── index.html         # Web 界面
│   │   │   └── AndroidManifest.xml    # 应用配置
│   │   └── build.gradle               # 应用构建配置
│   ├── build.gradle                   # 项目构建配置
│   └── settings.gradle
├── index.html                  # Web 界面 (源文件)
└── BUILD_APK.md               # 本文件
```

---

## ⚠️ 常见问题

### 1. Gradle 同步失败
```bash
# 清除缓存
./gradlew clean

# 重新同步
./gradlew build
```

### 2. 缺少 SDK
在 `local.properties` 文件中添加：
```properties
sdk.dir=C:\\Users\\YourName\\AppData\\Local\\Android\\Sdk
```

### 3. 权限问题 (Linux/Mac)
```bash
chmod +x android/gradlew
```

---

## 🎯 安装到手机

构建完成后，安装 APK：

```bash
# 通过 ADB 安装
adb install android/app/build/outputs/apk/debug/app-debug.apk

# 或通过 USB 传输到手机后安装
```

**注意**: 需要在手机设置中允许"安装未知来源应用"

---

## 📞 需要帮助？

如果遇到问题：
1. 检查 Android Studio 和 SDK 是否正确安装
2. 查看 `Build` 窗口的错误信息
3. 确保 `local.properties` 中的 SDK 路径正确

🐿️ **松鼠提醒团队**
