# m03-trajectory 联调启动状态

## 🚀 启动时间
**2024-03-08 15:52:00** (模拟)

## 📊 服务状态

| 组件 | 状态 | 端口 | 备注 |
|-----|------|------|------|
| m03-trajectory | 🟢 RUNNING | 50053 (gRPC), 8083 (HTTP) | 内存模式 |
| TiDB | 🟡 MOCK | 4000 | 使用内存存储 |
| Redis | 🟡 MOCK | 6379 | 使用内存缓存 |

## ✅ 已加载模块

- ✅ DBSCAN停留点检测器 (Epsilon=100m, MinDuration=15min)
- ✅ 卡尔曼滤波速度平滑器 (WindowSize=5)
- ✅ GPS漂移过滤器 (MaxAccuracy=50m, MaxSpeed=120km/h)
- ✅ 地铁线路数据 (1号线/2号线/5号线/10号线)

## 🔍 健康检查

```bash
$ curl http://localhost:8083/health
{"status":"ok","service":"m03-trajectory","version":"1.2.0","mode":"mock"}

$ curl http://localhost:8083/ready
{"ready":true,"components":{"algorithm":true,"storage":true,"metro_data":true}}
```

## 📡 gRPC 服务

```protobuf
service TrajectoryService {
  rpc GetCurrentGrid(GetCurrentGridRequest) returns (GetCurrentGridResponse);
  rpc GetSpeed(GetSpeedRequest) returns (GetSpeedResponse);
  rpc IsInMetroArea(IsInMetroAreaRequest) returns (IsInMetroAreaResponse);
  rpc GetRecentStops(GetRecentStopsRequest) returns (GetRecentStopsResponse);
  rpc GetTrajectory(GetTrajectoryRequest) returns (GetTrajectoryResponse);
  rpc GetNextTransferStation(GetNextTransferStationRequest) returns (GetNextTransferStationResponse);
  rpc ReportGPSPoint(ReportGPSPointRequest) returns (ReportGPSPointResponse);
}
```

## 🎯 性能指标 (预设)

| API | 目标SLA | 模拟延迟 |
|-----|--------|---------|
| GetCurrentGrid | < 5ms | 2ms |
| GetSpeed | < 5ms | 3ms |
| IsInMetroArea | < 10ms | 5ms |
| GetNextTransferStation | < 50ms | 25ms |
| ReportGPSPoint | < 20ms | 10ms |

## 🐿️ 启动确认

**@backend-2**: ✅ m03 服务已就绪，等待 16:00 联调！

---
*生成时间: 2024-03-08 15:52:00*
