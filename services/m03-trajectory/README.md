# m03-trajectory 轨迹处理服务 (v1.2)

## 服务概述

m03 轨迹处理服务是松鼠提醒项目的核心数据处理模块，负责：
- GPS 轨迹清洗与存储
- 停留点检测（DBSCAN算法）
- 速度平滑计算（卡尔曼滤波）
- 换乘站预测

## 性能目标

| 指标 | 目标 | 状态 |
|-----|------|------|
| 停留点检测 | P99 < 100ms | 🟡 待验证 |
| 速度平滑 | P99 < 10ms | 🟡 待验证 |
| GPS过滤 | P99 < 10ms | 🟡 待验证 |
| 数据查询 | P99 < 10ms | 🟡 待验证 |

## 项目结构

```
services/m03-trajectory/
├── cmd/server/           # 服务入口
│   └── main.go
├── internal/
│   ├── algorithm/        # 核心算法
│   │   ├── staypoint.go      # DBSCAN停留点检测
│   │   ├── staypoint_test.go # 算法测试
│   │   └── speed.go          # 卡尔曼滤波速度平滑
│   ├── service/          # 业务服务
│   │   ├── trajectory.go     # v1.0 基础实现
│   │   └── trajectory_v2.go  # v1.2 优化实现
│   ├── storage/          # 存储层
│   │   └── tidb.go           # TiDB操作
│   └── performance/      # 性能测试
│       └── benchmark.go
├── proto/                # gRPC定义
│   └── trajectory.proto
├── integration_test/     # 集成测试
│   ├── test_cases.go
│   └── report_template.md
├── docker-compose.yml    # 本地开发环境
├── Dockerfile
└── README.md
```

## 核心算法

### 1. 停留点检测 (DBSCAN变体)

```go
detector := algorithm.NewStayPointDetector()
stays := detector.Detect(points)
```

**参数：**
- Epsilon: 100米（聚类半径）
- MinDuration: 15分钟（最短停留时间）
- MinPoints: 5（最少GPS点数）

**复杂度：** O(n log n)

### 2. 速度平滑 (卡尔曼滤波 + 滑动窗口)

```go
smoother := algorithm.NewSpeedSmoother()
speeds := smoother.CalculateSpeeds(points)
```

**流程：**
1. 计算瞬时速度
2. 3-sigma 异常值过滤
3. 卡尔曼滤波平滑
4. 加权滑动窗口平均

### 3. GPS漂移过滤

```go
filter := algorithm.NewGPSDriftFilter()
filtered := filter.Filter(points)
```

**规则：**
- 精度 > 50米：过滤
- 速度 > 120km/h：过滤
- 5秒内跳跃 > 500米：过滤

## 快速开始

### 启动服务

```bash
cd services/m03-trajectory

# 启动依赖（TiDB + Redis）
docker-compose up -d tidb redis

# 等待TiDB就绪
sleep 10

# 启动服务
go run cmd/server/main.go
```

### 运行测试

```bash
# 单元测试
go test ./internal/algorithm/... -v

# 性能基准测试
go test -bench=. ./internal/algorithm/...

# 集成测试
go run integration_test/test_cases.go
```

### 性能基准测试

```bash
go test -bench=BenchmarkDetect -benchmem ./internal/algorithm/
```

## API 接口

### gRPC 服务

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

### HTTP 健康检查

```bash
curl http://localhost:8083/health
# {"status":"ok","service":"m03-trajectory","version":"1.2.0"}
```

## 配置

环境变量：

| 变量 | 默认值 | 说明 |
|-----|-------|------|
| TIDB_HOST | localhost | TiDB地址 |
| TIDB_PORT | 4000 | TiDB端口 |
| TIDB_USER | root | TiDB用户名 |
| TIDB_PASSWORD | - | TiDB密码 |
| TIDB_DATABASE | squirrel_m03 | 数据库名 |
| GRPC_PORT | 50053 | gRPC端口 |
| HTTP_PORT | 8083 | HTTP端口 |

## 开发计划

### v1.2 (当前)
- [x] DBSCAN停留点检测算法
- [x] 卡尔曼滤波速度平滑
- [x] TiDB存储层实现
- [ ] 性能优化（P99 < 100ms）
- [ ] 联调测试

### v1.3 (计划中)
- [ ] Redis缓存层
- [ ] 批量写入优化
- [ ] 空间索引（R-tree）

## 联调状态

| 模块 | 状态 | 负责人 |
|-----|------|-------|
| m01 状态感知 | ⏳ 等待联调 | @backend-1 |
| m03 轨迹处理 | ✅ 就绪 | @backend-2 |

---

🐿️ **松鼠提醒项目 | 后端开发-2 @backend-2**
