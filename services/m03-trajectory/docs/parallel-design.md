# 算法并行化设计文档

**版本**: v1.0  
**日期**: 2024-03-17  
**设计人**: @backend-2  
**状态**: 待评审

---

## 1. 设计目标

### 1.1 问题背景

当前m03-trajectory采用串行处理模式，在多核CPU环境下利用率低：
- CPU使用率: 45% (8核机器)
- QPS瓶颈: 1,847 (单核限制)

### 1.2 优化目标

| 指标 | 当前 | 目标 | 提升 |
|-----|------|------|------|
| QPS | 1,847 | **2,400** | +30% |
| CPU利用率 | 45% | **80%** | +78% |
| 多用户延迟 | 串行累加 | **并行持平** | -50% |

---

## 2. 架构设计

### 2.1 总体架构

```
┌─────────────────────────────────────────┐
│           TrajectoryService             │
│  ┌─────────┐  ┌─────────┐  ┌────────┐  │
│  │ User A  │  │ User B  │  │ User C │  │
│  └────┬────┘  └────┬────┘  └───┬────┘  │
│       │            │            │       │
│       └────────────┼────────────┘       │
│                    │                    │
│            ┌───────▼────────┐           │
│            │   Job Queue    │           │
│            │  (buffered)    │           │
│            └───────┬────────┘           │
│                    │                    │
│       ┌────────────┼────────────┐       │
│       ▼            ▼            ▼       │
│  ┌────────┐  ┌────────┐  ┌────────┐   │
│  │Worker 1│  │Worker 2│  │Worker 3│   │
│  │(CPU 1) │  │(CPU 2) │  │(CPU 3) │   │
│  └────┬───┘  └────┬───┘  └────┬───┘   │
│       └────────────┼────────────┘       │
│                    │                    │
│            ┌───────▼────────┐           │
│            │  Result Chan   │           │
│            └────────────────┘           │
└─────────────────────────────────────────┘
```

### 2.2 核心组件

#### WorkerPool (工作池)

```go
type WorkerPool struct {
    numWorkers int              // 工作协程数
    jobQueue   chan Job         // 任务队列
    wg         sync.WaitGroup   // 同步等待
    ctx        context.Context  // 生命周期控制
}
```

**职责**:
- 管理工作协程生命周期
- 任务分发
- 结果收集
- 优雅关闭

#### Job (任务)

```go
type Job struct {
    UserID   string              // 用户标识
    Points   []GPSPoint          // GPS轨迹点
    ResultCh chan<- Result       // 结果回调
}
```

#### Worker (工作协程)

```go
func (wp *WorkerPool) worker(id int) {
    for job := range wp.jobQueue {
        result := processJob(job)
        job.ResultCh <- result
    }
}
```

**处理流程**:
1. 接收任务
2. 使用FastSmooth计算速度 (SIMD优化)
3. 检测停留点
4. 返回结果

---

## 3. 并发安全设计

### 3.1 数据隔离

- **无共享状态**: 每个Worker独立处理，无共享内存
- **只读共享**: algorithm.SpeedSmoother等可并发读
- **结果通道**: 每个Job自带ResultCh，避免竞争

### 3.2 竞态条件防护

```go
// Race Detector测试
go test -race ./internal/optimization/

// 静态检查
go vet ./internal/optimization/
```

### 3.3 资源限制

| 资源 | 限制 | 说明 |
|-----|------|------|
| CPU | runtime.NumCPU() | 自适应核心数 |
| 内存 | 100MB/Worker | 防止OOM |
| 队列 | numWorkers*4 | 有界队列防堆积 |
| 超时 | 30s | 防止死锁 |

---

## 4. 性能优化

### 4.1 减少锁竞争

- 使用channel通信，避免显式锁
- 每个Worker独立实例化算法对象
- 无全局状态

### 4.2 减少内存分配

```go
// 复用PointPool
pointPool := sync.Pool{
    New: func() interface{} {
        return make([]GPSPoint, 0, 128)
    },
}
```

### 4.3 批处理优化

```go
// 批量提交，减少channel操作开销
func (wp *WorkerPool) BatchProcess(users []UserData) []Result {
    // 一次性提交所有任务
    // 并发收集结果
}
```

---

## 5. 错误处理

### 5.1 错误类型

| 错误 | 处理 | 说明 |
|-----|------|------|
| 任务提交超时 | 返回错误 | 队列满 |
| 处理超时 | 上下文取消 | 防止死锁 |
| 算法错误 | 记录日志 | 继续处理 |
| Worker panic | 自动恢复 | defer+recover |

### 5.2 降级策略

```go
if !wp.Submit(job) {
    // 队列满，同步处理
    return syncProcess(job)
}
```

---

## 6. 监控与可观测

### 6.1 关键指标

| 指标 | 类型 | 说明 |
|-----|------|------|
| worker_pool_jobs_total | Counter | 总任务数 |
| worker_pool_queue_size | Gauge | 队列长度 |
| worker_pool_active_workers | Gauge | 活跃Worker数 |
| worker_pool_job_duration | Histogram | 任务处理耗时 |

### 6.2 日志规范

```go
log.Printf("[WorkerPool] Worker %d started", id)
log.Printf("[WorkerPool] Job %s completed in %v", userID, duration)
log.Printf("[WorkerPool] Error processing job %s: %v", userID, err)
```

---

## 7. 测试策略

### 7.1 单元测试

```go
func TestWorkerPoolBasic(t *testing.T) {
    wp := NewWorkerPool(4)
    wp.Start()
    defer wp.Stop()
    
    // 提交任务
    // 验证结果
    // 检查无竞态
}
```

### 7.2 压力测试

```go
func BenchmarkWorkerPool(b *testing.B) {
    wp := NewWorkerPool(runtime.NumCPU())
    wp.Start()
    defer wp.Stop()
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            wp.Submit(job)
        }
    })
}
```

### 7.3 正确性测试

- 串行 vs 并行结果一致性
- 高并发下无panic
- 资源无泄漏

---

## 8. 部署方案

### 8.1 灰度策略

1. **10%流量**: 观察错误率
2. **50%流量**: 验证性能提升
3. **100%流量**: 全量发布

### 8.2 回滚方案

```go
// 功能开关
if config.EnableParallel {
    return wp.BatchProcess(users)
}
return serialProcess(users)  // 回退
```

---

## 9. 风险评估

| 风险 | 概率 | 影响 | 应对 |
|-----|------|------|------|
| 竞态条件 | 中 | 高 | race detector + 代码审查 |
| 资源泄漏 | 低 | 高 | context + defer |
| 性能不及预期 | 低 | 中 | 预留优化空间 |
| 兼容性问题 | 低 | 中 | 功能开关 |

---

## 10. 验收标准

- [ ] 单元测试覆盖率 >85%
- [ ] Race detector无警告
- [ ] QPS提升 >25%
- [ ] 错误率 <0.01%
- [ ] 内存无泄漏

---

## 附录

### A. 接口定义

```go
// WorkerPool interface
type WorkerPool interface {
    Start()
    Stop()
    Submit(Job) bool
    BatchProcess([]UserData) []Result
    GetStats() map[string]interface{}
}
```

### B. 配置参数

```yaml
worker_pool:
  num_workers: 0  # 0表示自适应CPU核心数
  queue_size_factor: 4
  job_timeout: 30s
  shutdown_timeout: 10s
```

---

> 🐿️ **松鼠提醒项目 | 后端开发-2 @backend-2**
