# 🐿️ 松鼠提醒 - 快速试用指引

> 版本: v1.2.1-dev | 更新时间: 2026-03-21

## 📱 当前可用版本

| 版本 | 状态 | 可用性 |
|------|------|--------|
| v1.2.0 MVP | 已发布 (3/15) | ✅ 可试用 |
| v1.2.1 | 开发中 (92%) | 🔄 3/29发布 |

---

## 🚀 快速启动（开发者模式）

### 方式一：本地启动后端服务

```bash
# 1. 进入服务目录
cd squirrel-services

# 2. 启动 M01 状态引擎
cd m01-state-engine/go
go run cmd/server/main.go
# 服务将运行在 :8080

# 3. 启动 M02 告警服务（新终端）
cd m02-alert-service/go
go run cmd/server/main.go
# 服务将运行在 :8081

# 4. 启动 M03 轨迹处理（新终端）
cd m03-trajectory
go run cmd/server/main.go
# 服务将运行在 :8082
```

### 方式二：Docker 一键启动

```bash
# 在项目根目录
docker-compose up -d

# 服务将启动：
# - API Gateway: http://localhost:8000
# - State Engine: http://localhost:8080
# - Alert Service: http://localhost:8081
# - Trajectory Processor: http://localhost:8082
```

---

## 🧪 日常地铁测试指引

### 测试准备

1. **安装测试 App**
   - iOS: TestFlight 邀请（需联系团队）
   - Android: APK 下载 `releases/v1.2.0-mvp.apk`
   - 或直接使用 Web 版本: `https://test.squirrel.couponly.io`

2. **注册测试账号**
   ```
   测试账号: test@example.com
   密码: Test1234!
   ```

3. **配置常用路线**
   - 打开 App → "我的路线" → 添加
   - 输入起点站、终点站
   - 设置提醒距离（建议：到站前 2-3 个站）

### 日常测试步骤

| 时间 | 操作 | 预期结果 |
|------|------|---------|
| 进站前 | 打开 App，选择今日路线 | GPS 开始追踪 |
| 进站后 | 确认"已上车" | 状态变为"乘车中" |
| 行驶中 | 手机放口袋，正常使用 | 后台静默追踪 |
| 接近目标站 | - | 震动 + 声音提醒 |
| 到站后 | 确认"已下车" | 记录完成 |

### 测试反馈

测试后请在以下渠道反馈：
- 📧 邮件: `feedback@squirrel.couponly.io`
- 💬 企业微信: SquirrelTest 群
- 🐛 GitHub Issue: `squirrel-services/issues`

反馈内容：
```
日期: 2026-03-21
路线: 罗湖 → 福田
手机: iPhone 15 / Android 14
问题: [如有]
建议: [如有]
```

---

## 📊 当前功能状态

| 功能 | v1.2.0 | v1.2.1 (3/29) | 备注 |
|------|--------|---------------|------|
| GPS 轨迹追踪 | ✅ | ✅ 更精准 | v1.2.1 +SIMD优化 |
| 到站提醒 | ✅ | ✅ 更快 | v1.2.1 P99 5ms |
| 智能识别上下车 | ✅ | ✅ 更准确 | v1.2.1 96.2%命中率 |
| 多路线管理 | ✅ | ✅ | - |
| 职业识别 | ✅ | ✅ 增强 | v1.2.1 52.45%自由职业识别 |
| 离线模式 | ⚠️ | ✅ | v1.2.1 新增 |
| 省电优化 | ✅ | ✅ 更优 | v1.2.1 多级缓存 |

---

## 🔧 故障排除

### 常见问题

**Q: App 定位不准？**
```
检查: 设置 → 隐私 → 位置服务 → 松鼠提醒 → "始终"
建议: iPhone 用户关闭"精确位置"可省电
```

**Q: 提醒延迟？**
```
可能原因:
1. 网络信号差（地铁隧道）
2. GPS 信号弱
3. 后台被杀

解决方案:
1. 开启"到站前 3 站提醒"
2. 关闭省电模式
3. 设置白名单
```

**Q: 想测试但不想真的坐地铁？**
```
使用模拟器:
# 连接本地服务
curl http://localhost:8080/api/v1/simulate \
  -X POST \
  -d '{"route": "罗湖-福田", "speed": 60}'
```

---

## 📅 正式发布计划

| 版本 | 日期 | 渠道 |
|------|------|------|
| v1.2.0 MVP | 已发布 | 内测 |
| v1.2.1 | 3/29 | 内测 +
| v1.2.2 | 4/12 | 公测 |
| v1.3.0 | 4/26 | 正式版 |

---

## 📞 联系方式

| 渠道 | 链接/地址 |
|------|----------|
| 测试文档 | https://main.squirrel-docs.pages.dev |
| API 文档 | https://api.squirrel.couponly.io/docs |
| 反馈邮箱 | feedback@squirrel.couponly.io |
| 紧急联系 | oncall@squirrel.couponly.io |

---

🐿️ **祝您测试顺利，再也不坐过站！**
