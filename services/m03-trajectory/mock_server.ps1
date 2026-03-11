# mock_server.ps1 - m03 服务模拟（演示用）
# 实际环境使用: go run cmd/server/main.go

param(
    [int]$HttpPort = 8083,
    [int]$GrpcPort = 50053
)

Write-Host "🚀 启动 m03-trajectory 模拟服务" -ForegroundColor Green
Write-Host "================================" -ForegroundColor Green
Write-Host ""
Write-Host "HTTP 端口: $HttpPort" -ForegroundColor Cyan
Write-Host "gRPC 端口: $GrpcPort (模拟)" -ForegroundColor Cyan
Write-Host ""

# 模拟服务启动延迟
Write-Host "⏳ 初始化算法模块..." -ForegroundColor Yellow
Start-Sleep -Milliseconds 500
Write-Host "  ✅ DBSCAN停留点检测器就绪" -ForegroundColor Green

Write-Host "⏳ 初始化卡尔曼滤波器..." -ForegroundColor Yellow
Start-Sleep -Milliseconds 300
Write-Host "  ✅ 速度平滑模块就绪" -ForegroundColor Green

Write-Host "⏳ 连接 TiDB..." -ForegroundColor Yellow
Start-Sleep -Milliseconds 200
Write-Host "  ⚠️  TiDB 连接失败，使用内存存储" -ForegroundColor Yellow
Write-Host "  ✅ 内存缓存已启用" -ForegroundColor Green

Write-Host "⏳ 加载地铁线路数据..." -ForegroundColor Yellow
Start-Sleep -Milliseconds 200
Write-Host "  ✅ 1号线/2号线/5号线/10号线数据已加载" -ForegroundColor Green
Write-Host ""

# 创建 HTTP 监听器
$listener = New-Object System.Net.HttpListener
$listener.Prefixes.Add("http://localhost:$HttpPort/")

Write-Host "🔍 启动健康检查服务..." -ForegroundColor Yellow

try {
    $listener.Start()
    Write-Host "  ✅ HTTP 服务已启动 http://localhost:$HttpPort" -ForegroundColor Green
    Write-Host ""
    Write-Host "================================" -ForegroundColor Green
    Write-Host "🎉 m03-trajectory 模拟服务启动成功！" -ForegroundColor Green
    Write-Host ""
    Write-Host "可用端点:" -ForegroundColor Cyan
    Write-Host "  GET  /health       -> 健康检查" -ForegroundColor White
    Write-Host "  GET  /ready        -> 就绪检查" -ForegroundColor White
    Write-Host "  POST /v1/report    -> GPS上报" -ForegroundColor White
    Write-Host ""
    Write-Host "gRPC 服务 (模拟):" -ForegroundColor Cyan
    Write-Host "  localhost:$GrpcPort" -ForegroundColor White
    Write-Host ""
    Write-Host "按 Ctrl+C 停止服务" -ForegroundColor Yellow
    Write-Host ""

    while ($listener.IsListening) {
        $context = $listener.GetContext()
        $request = $context.Request
        $response = $context.Response
        
        $path = $request.Url.LocalPath
        $method = $request.HttpMethod
        
        Write-Host "[$method] $path" -ForegroundColor Gray
        
        switch ($path) {
            "/health" {
                $content = '{"status":"ok","service":"m03-trajectory","version":"1.2.0","mode":"mock"}'
                $buffer = [System.Text.Encoding]::UTF8.GetBytes($content)
                $response.ContentType = "application/json"
                $response.ContentLength64 = $buffer.Length
                $response.OutputStream.Write($buffer, 0, $buffer.Length)
            }
            
            "/ready" {
                $content = '{"ready":true,"components":{"algorithm":true,"storage":true,"metro_data":true}}'
                $buffer = [System.Text.Encoding]::UTF8.GetBytes($content)
                $response.ContentType = "application/json"
                $response.ContentLength64 = $buffer.Length
                $response.OutputStream.Write($buffer, 0, $buffer.Length)
            }
            "/v1/report" {
                $content = '{"success":true,"message":"GPS point received"}'
                $buffer = [System.Text.Encoding]::UTF8.GetBytes($content)
                $response.ContentType = "application/json"
                $response.ContentLength64 = $buffer.Length
                $response.OutputStream.Write($buffer, 0, $buffer.Length)
            }
            default {
                $response.StatusCode = 404
                $content = '{"error":"not found"}'
                $buffer = [System.Text.Encoding]::UTF8.GetBytes($content)
                $response.ContentLength64 = $buffer.Length
                $response.OutputStream.Write($buffer, 0, $buffer.Length)
            }
        }
        
        $response.OutputStream.Close()
    }
}
catch {
    Write-Host "❌ 服务启动失败: $_" -ForegroundColor Red
}
finally {
    $listener.Stop()
    Write-Host ""
    Write-Host "🛑 服务已停止" -ForegroundColor Yellow
}
