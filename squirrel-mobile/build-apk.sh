#!/bin/bash

# 构建 APK 脚本
# 需要安装: Node.js, Cordova

echo "🐿️ 松鼠提醒 APK 构建脚本"
echo "=========================="

# 检查依赖
if ! command -v node &> /dev/null; then
    echo "❌ 请先安装 Node.js: https://nodejs.org/"
    exit 1
fi

# 安装 Cordova（如果没有）
if ! command -v cordova &> /dev/null; then
    echo "📦 安装 Cordova..."
    npm install -g cordova
fi

# 创建 Cordova 项目
if [ ! -d "cordova-app" ]; then
    echo "📁 创建 Cordova 项目..."
    cordova create cordova-app com.squirrel.reminder "松鼠提醒"
fi

# 复制 Web 文件
echo "📋 复制 Web 文件..."
cp index.html cordova-app/www/
cp server.go cordova-app/www/

# 修改 index.html 添加 Cordova 脚本
sed -i 's/<\/head>/<script src="cordova.js"><\/script><\/head>/' cordova-app/www/index.html

# 添加 Android 平台
cd cordova-app
if [ ! -d "platforms/android" ]; then
    echo "📱 添加 Android 平台..."
    cordova platform add android
fi

# 构建 APK
echo "🔨 构建 APK..."
cordova build android --release

echo ""
echo "✅ APK 构建完成！"
echo "📦 输出路径: cordova-app/platforms/android/app/build/outputs/apk/release/app-release-unsigned.apk"
echo ""
echo "⚠️  注意: 这是未签名的 APK，正式使用需要签名"
