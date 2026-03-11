@echo off
chcp 65001 >nul
echo 🐿️ 松鼠提醒 APK 构建工具
echo ==========================
echo.

echo 📋 检查 Java...
java -version >nul 2>&1
if errorlevel 1 (
    echo ❌ Java 未安装
    echo.
    echo 请安装 Java JDK 17:
    echo https://adoptium.net/
    echo.
    pause
    exit /b 1
)

echo ✅ Java 已安装
echo.

cd /d "%~dp0android"

echo 🔨 构建 APK...
echo.

if not exist "gradlew.bat" (
    echo ❌ 未找到 Gradle Wrapper
    echo 请确保您在 squirrel-mobile 目录中
    pause
    exit /b 1
)

call gradlew.bat assembleDebug --no-daemon

if errorlevel 1 (
    echo.
    echo ❌ 构建失败
    pause
    exit /b 1
)

echo.
echo 🎉 APK 构建成功！
echo.
echo 📦 APK 位置:
echo    %~dp0android\app\build\outputs\apk\debug\app-debug.apk
echo.
echo 📱 安装命令:
echo    adb install "%~dp0android\app\build\outputs\apk\debug\app-debug.apk"
echo.

pause
