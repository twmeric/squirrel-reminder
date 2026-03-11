# v1.2.1 灰度发布方案

**版本**: v1.2.1  
**发布人**: @backend-2  
**DDL**: 3/22 12:00  
**目标**: 零停机，平滑升级

---

## 1. 发布概览

### 版本信息

| 项 | 值 |
|---|----|
| 版本号 | v1.2.1 |
| 上一个版本 | v1.2.0 |
| 主要特性 | SIMD优化 + 并行化 + 多级缓存 |
| 预期提升 | QPS +50%, 延迟 -40%, 命中率 96% |

### 发布策略

**灰度发布**: 10% → 50% → 100%
**发布窗口**: 3/22 20:00 - 3/23 08:00 (低峰期)

---

## 2. 灰度阶段

### Phase 1: 10% 灰度 (3/22 20:00-22:00)

**目标**: 验证稳定性，观察错误率

```yaml
# Kubernetes灰度配置
canary:
  replicas: 3          # 总10个副本，3个新版本
  weight: 10           # 10%流量
  headers:
    canary: "always"   # 指定用户测试
```

**验收标准**:
- [ ] 错误率 < 0.1%
- [ ] P99延迟 < 7ms
- [ ] 无panic/crash
- [ ] 监控指标正常

**回滚条件**:
- 错误率 > 0.5%
- P99延迟 > 15ms
- 任何Critical告警

---

### Phase 2: 50% 灰度 (3/22 22:00-00:00)

**目标**: 验证性能提升

```yaml
canary:
  replicas: 5          # 5个新版本
  weight: 50           # 50%流量
```

**验收标准**:
- [ ] QPS > 2,500
- [ ] 命中率 > 95%
- [ ] CPU使用率 < 70%
- [ ] 内存使用率 < 80%

---

### Phase 3: 100% 全量 (3/23 00:00-08:00)

**目标**: 完全切换

```yaml
canary:
  replicas: 10         # 全部新版本
  weight: 100          # 100%流量
```

**验收标准**:
- [ ] 所有指标达标
- [ ] 监控24小时无异常
- [ ] 用户反馈正常

---

## 3. 发布前检查清单

### 3.1 代码检查

- [x] 代码审查通过
- [x] 单元测试覆盖率 >85%
- [x] 集成测试通过
- [x] Race detector无警告
- [x] 性能基准测试通过

### 3.2 配置检查

- [x] 配置文件更新
- [x] 环境变量确认
- [x] 数据库迁移脚本准备
- [x] 降级开关测试

### 3.3 监控检查

- [x] 告警规则配置
- [x] Dashboard更新
- [x] 日志收集正常
- [x] 链路追踪正常

### 3.4 回滚准备

- [x] v1.2.0镜像备份
- [x] 回滚脚本准备
- [x] 数据库回滚方案
- [x] 降级开关验证

---

## 4. 发布执行

### 4.1 发布脚本

```bash
#!/bin/bash
# deploy-canary.sh

VERSION="v1.2.1"
PHASE=$1

echo "🚀 开始灰度发布: $VERSION Phase $PHASE"

case $PHASE in
  10)
    echo "Phase 1: 10% 灰度"
    kubectl set image deployment/m03-trajectory \
      m03-trajectory=squirrel/m03-trajectory:$VERSION
    kubectl patch service m03-trajectory \
      -p '{"spec":{"trafficPolicy":{"weight":10}}}'
    ;;
  50)
    echo "Phase 2: 50% 灰度"
    kubectl patch service m03-trajectory \
      -p '{"spec":{"trafficPolicy":{"weight":50}}}'
    ;;
  100)
    echo "Phase 3: 100% 全量"
    kubectl patch service m03-trajectory \
      -p '{"spec":{"trafficPolicy":{"weight":100}}}'
    ;;
  rollback)
    echo "⚠️  执行回滚"
    kubectl rollout undo deployment/m03-trajectory
    ;;
esac

echo "✅ Phase $PHASE 完成"
```

### 4.2 监控命令

```bash
# 实时监控
kubectl logs -f deployment/m03-trajectory -c m03-trajectory

# 指标检查
curl http://localhost:9090/metrics | grep m03_request_total
curl http://localhost:9090/metrics | grep m03_request_duration_seconds

# 错误率检查
watch 'curl -s http://localhost:9090/metrics | grep m03_error_total'
```

---

## 5. 回滚方案

### 5.1 自动回滚触发条件

| 指标 | 阈值 | 动作 |
|-----|------|------|
| 错误率 | > 0.5% | 自动回滚到v1.2.0 |
| P99延迟 | > 15ms | 自动回滚到v1.2.0 |
| Pod重启 | > 3次/10min | 自动回滚到v1.2.0 |

### 5.2 手动回滚

```bash
# 一键回滚
./scripts/deploy-canary.sh rollback

# 或者
kubectl rollout undo deployment/m03-trajectory

# 验证回滚
kubectl rollout status deployment/m03-trajectory
```

### 5.3 降级开关

```yaml
# 紧急降级配置
feature_flags:
  enable_fast_smooth: false    # 降级到卡尔曼滤波
  enable_parallel: false       # 降级到串行处理
  enable_multilevel_cache: false # 降级到Redis单级缓存
```

---

## 6. 发布后验证

### 6.1 功能验证

```bash
# 健康检查
curl http://localhost:8083/health

# 速度计算测试
curl "http://localhost:8083/api/v1/speed?user_id=test"

# 缓存命中率检查
curl http://localhost:9090/metrics | grep cache_hit_rate
```

### 6.2 性能验证

| 指标 | 目标 | 验证命令 |
|-----|------|---------|
| QPS | > 2,500 | `wrk -t10 -c100 -d30s http://localhost:8083/api/v1/speed` |
| P99延迟 | < 7ms | `curl http://localhost:9090/metrics` |
| 命中率 | > 95% | `curl http://localhost:9090/metrics` |

---

## 7. 风险与应对

| 风险 | 概率 | 影响 | 应对 |
|-----|------|------|------|
| 发布失败 | 低 | 高 | 自动回滚+监控告警 |
| 性能不达标 | 低 | 中 | 功能开关降级 |
| 数据不一致 | 极低 | 高 | 数据库事务+回滚脚本 |
| 监控盲区 | 中 | 中 | 增加临时监控点 |

---

## 8. 沟通计划

| 时间 | 事项 | 通知人 |
|-----|------|--------|
| 发布前2h | 发布预告 | @architect-lead @oncall |
| Phase 1完成 | 10%结果 | @architect-lead |
| Phase 2完成 | 50%结果 | @architect-lead |
| 发布完成 | 全量确认 | 全团队 |
| 发布后24h | 总结报告 | @architect-lead |

---

## 9. 发布后总结模板

```markdown
# v1.2.1 发布总结

**发布时间**: 3/23 08:00
**发布结果**: ✅ 成功

## 关键指标

| 指标 | 发布前 | 发布后 | 提升 |
|-----|--------|--------|------|
| QPS | 1,847 | 2,823 | +53% |
| P99延迟 | 8.5ms | 5.1ms | -40% |
| 命中率 | 85.4% | 96.3% | +11% |

## 问题记录

- 无

## 经验总结

- 灰度发布策略有效
- 监控告警及时
- 回滚方案准备充分
```

---

> 🐿️ **松鼠提醒项目 | 后端开发-2 @backend-2**  
> **准备时间**: 2024-03-18  
> **发布时间**: 3/22 20:00
