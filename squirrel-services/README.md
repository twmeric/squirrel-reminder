# Squirrel Services

Squirrel项目的核心服务实现 - Hybrid Architecture中的代码仓库。

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                     Squirrel Services                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐ │
│  │   m03        │    │   m04        │    │  Future      │ │
│  │ Trajectory   │◄──►│ Insight      │    │  Services    │ │
│  │ Service      │    │ Engine       │    │              │ │
│  │ (Go/gRPC)    │    │ (Python/HTTP)│    │              │ │
│  └──────┬───────┘    └──────┬───────┘    └──────────────┘ │
│         │                   │                               │
│         └───────────────────┘                               │
│              TiDB (Analytics)                               │
│              Redis (Cache)                                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────┐
              │   Cloudflare Edge API     │
              │   (JavaScript/WASM)       │
              └───────────────────────────┘
```

## 服务说明

### m03-trajectory (轨迹处理服务)

**语言**: Go 1.21
**协议**: gRPC + HTTP/REST
**端口**: 50053(gRPC), 8083(HTTP)

**功能**:
- 实时GPS批量处理
- 停留点检测 (DBSCAN算法)
- 最近站点匹配 (KD-tree)
- 速度计算与平滑

**性能指标**:
- ProcessBatch P99: <50ms (100 points)
- GetSpeed P99: <10ms
- 内存占用: <200MB per instance

### m04-insight-engine (用户洞察引擎)

**语言**: Python 3.11
**框架**: FastAPI
**端口**: 8084

**功能**:
- 用户画像分析 (Home/Work检测)
- 通勤模式识别
- 生活事件检测 (换工作、搬家、失业等)
- 批量分析API

**性能指标**:
- 单用户分析: <200ms
- 批量100用户: <3s
- 准确率: Home 94%, Work 90%

## 快速开始

### 本地开发

```bash
# 克隆仓库
git clone https://github.com/squirrelawake/squirrel-services.git
cd squirrel-services

# 复制环境配置
cp docker/.env.example docker/.env
# 编辑 .env 填入你的配置

# 启动所有服务
docker-compose -f docker/docker-compose.yml up -d

# 验证服务
./scripts/deploy.sh local
```

### 构建镜像

```bash
# 构建M03
docker build -f docker/Dockerfile.m03 -t squirrel/m03:latest .

# 构建M04
docker build -f docker/Dockerfile.m04 -t squirrel/m04:latest .
```

### 运行测试

```bash
# 运行所有测试
make test

# M03 Go测试
cd m03-trajectory && go test -v ./...

# M04 Python测试
cd m04-insight-engine && pytest tests/ -v
```

## 部署

### 本地 (Docker Compose)

```bash
./scripts/deploy.sh local
```

### 腾讯云TKE (Kubernetes)

```bash
# 配置kubectl访问TKE
export TENCENT_SECRET_ID=your_secret_id
export TENCENT_SECRET_KEY=your_secret_key

# 部署
./scripts/deploy.sh tencent
```

### 数据库迁移

```bash
# 本地开发
mysql -h localhost -P 4000 -u root < docker/migrations/001_init.sql

# 生产环境
./scripts/migrate.sh production
```

## API文档

### m03-trajectory

**gRPC服务**:
- `TrajectoryProcessor.ProcessBatch` - 批量处理位置
- `LocationService.GetSpeed` - 获取当前速度
- `LocationService.GetNearestStation` - 获取最近站点

**HTTP端点**:
- `GET /health` - 健康检查
- `GET /ready` - 就绪检查
- `GET /metrics` - Prometheus指标

### m04-insight-engine

**REST API**:
- `GET /api/v1/users/{user_id}/profile` - 获取用户画像
- `GET /api/v1/users/{user_id}/events` - 获取生活事件
- `POST /api/v1/batch/analyze` - 批量分析用户
- `GET /health` - 健康检查
- `GET /ready` - 就绪检查
- `GET /metrics` - Prometheus指标

完整API文档: [docs/API.md](docs/API.md)

## 监控与告警

### 访问监控界面

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686

### 关键指标

| 指标 | 告警阈值 | 严重程度 |
|-----|---------|---------|
| Request Latency P99 | > 100ms | Warning |
| Error Rate | > 1% | Critical |
| CPU Usage | > 80% | Warning |
| Memory Usage | > 85% | Warning |

## 贡献指南

1. Fork本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建Pull Request

### 代码规范

- Go: 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- Python: 使用 Black 格式化, 遵循 PEP 8

## 许可证

MIT License - 详见 [LICENSE](LICENSE)

## 相关仓库

- [squirrel-docs](https://github.com/squirrelawake/squirrel-docs) - 文档和架构知识库
- [squirrel-mobile](https://github.com/squirrelawake/squirrel-mobile) - 移动端App
- [squirrel-edge](https://github.com/squirrelawake/squirrel-edge) - Cloudflare Edge API

## 联系方式

- 项目主页: https://squirrel.couponly.io
- 文档站点: https://docs.squirrel.couponly.io
- 问题反馈: https://github.com/squirrelawake/squirrel-services/issues
