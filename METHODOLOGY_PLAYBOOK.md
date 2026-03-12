# 🎯 松鼠 Reminder 开发方法论 Playbook

> 从 0 到 1 构建跨平台应用的完整思维框架
> 
> 适用场景：大数据项目、跨平台应用、CI/CD 流程设计

---

## 一、项目架构思维

### 1.1 分层架构设计（Hybrid Architecture）

```
┌─────────────────────────────────────────┐
│  表现层 (Presentation)                  │  
│  ├─ Android Native (WebView)           │  ← 权限管理、系统接口
│  └─ Web PWA (HTML/CSS/JS)              │  ← 业务逻辑、UI渲染
├─────────────────────────────────────────┤
│  服务层 (Service Layer)                 │
│  └─ Cloudflare Workers (Edge API)      │  ← 无服务器、全球部署
├─────────────────────────────────────────┤
│  数据层 (Data Layer)                    │
│  └─ TiDB Cloud + R2 Storage            │  ← 分布式、可扩展
└─────────────────────────────────────────┘
```

**核心原则**：
- **关注点分离**：Native 管权限，Web 管业务，Cloud 管数据
- **渐进增强**：MVP → Production → Optimization
- **云原生**：利用 Serverless 降低运维成本

---

## 二、问题驱动开发（PDD - Problem Driven Development）

### 2.1 错误处理五步曲

本项目共经历 **12 次构建失败**，总结出的排查方法论：

```
Step 1: 精确定位（Localization）
    └─ 查看 GitHub Actions 日志，找到第一处 FAILED
    
Step 2: 根因分类（Categorization）
    ├─ 依赖问题（Dependency）→ SDK版本不匹配
    ├─ 配置问题（Configuration）→ 缺少gradle.properties
    ├─ 资源问题（Resource）→ 图标文件缺失
    └─ 权限问题（Permission）→ GPS授权
    
Step 3: 最小修复（Minimal Fix）
    └─ 只改必要的文件，避免过度修复
    
Step 4: 验证假设（Validation）
    └─ 重新触发构建，验证是否解决
    
Step 5: 知识沉淀（Documentation）
    └─ 记录错误模式和解决方案
```

### 2.2 本项目错误模式库

| 错误 | 根因 | 修复 | 预防 |
|-----|------|------|------|
| `gradlew: No such file` | 缺少Gradle wrapper | 创建gradle-wrapper.properties | 确保完整上传所有文件 |
| `compileSdk 34 not supported` | 依赖库版本冲突 | 统一SDK版本为33或升级Plugin | 锁定依赖版本 |
| `AndroidX property not enabled` | 缺少gradle.properties | 添加android.useAndroidX=true | 提供完整的配置文件 |
| `ic_launcher not found` | 资源文件缺失 | 创建mipmap图标 | 资源完整性检查 |
| `GPS Timeout expired` | 定位超时设置过短 | timeout:10000→30000 | 合理的超时配置 |

---

## 三、CI/CD 设计哲学

### 3.1 GitHub Actions 最佳实践

```yaml
# 核心设计原则：幂等性 + 可重现性

jobs:
  build:
    runs-on: ubuntu-latest  # 固定环境
    
    steps:
      # 1. 代码检出（确定性的起点）
      - uses: actions/checkout@v4
      
      # 2. 环境固定（版本锁定）
      - uses: actions/setup-java@v4
        with:
          java-version: '17'  # 固定JDK版本
          distribution: 'temurin'
      
      # 3. 缓存策略（加速构建）
      - uses: gradle/actions/setup-gradle@v3
        with:
          gradle-version: '8.2'  # 锁定构建工具版本
      
      # 4. 构建步骤（单一职责）
      - name: Build APK
        run: gradle assembleDebug
      
      # 5. 产物管理（可追溯）
      - uses: actions/upload-artifact@v4
        with:
          name: squirrel-reminder-apk
          path: app/build/outputs/apk/debug/app-debug.apk
```

### 3.2 版本控制策略

```
主分支（master/main）
    │
    ├─ 功能开发分支（feature/gps-optimization）
    │      └─ 小步提交（micro-commits）
    │      └─ 频繁 rebase（保持线性历史）
    │
    └─ 修复分支（hotfix/gps-timeout）
           └─ 原子提交（atomic commits）
           └─ 清晰提交信息：
               "Fix: 延长GPS超时时间 10s→30s（解决室内定位失败）"
```

---

## 四、跨平台开发思维

### 4.1 WebView 混合开发模式

**适用场景**：
- ✅ 快速迭代（无需应用商店审核）
- ✅ 跨平台复用（一套代码，多端运行）
- ✅ 动态更新（HTML/CSS/JS 可热更新）

**关键边界**：
| 能力 | Web 层 | Native 层 | 协作方式 |
|-----|--------|----------|---------|
| UI渲染 | ✅ HTML/CSS | ❌ | WebView 渲染 |
| GPS定位 | ❌ | ✅ 权限+硬件 | JS Bridge 调用 |
| 本地存储 | ✅ LocalStorage | ✅ SQLite | 分层存储 |
| 后台运行 | ❌ | ✅ Service | 需要Native实现 |
| 推送通知 | ❌ | ✅ FCM/APNs | 需额外开发 |

### 4.2 资源优化策略

**"无感"设计的三层境界**：

```
Level 1: 显性配置（当前）
    用户操作：打开App → 选择起点 → 选择终点 → 点击开始
    资源消耗：前台GPS持续运行

Level 2: 半自动（改进）
    用户操作：打开App → 一键确认智能推荐
    资源消耗：按需启动GPS

Level 3: 完全无感（理想）
    用户操作：无需操作（系统自动检测进入地铁站）
    资源消耗：地理围栏触发，后台Service运行
```

---

## 五、大数据项目可借鉴的思维

### 5.1 边缘计算设计

本项目采用 **Cloudflare Workers** 作为边缘节点：

```
用户手机（香港）
    ↓
最近的 Cloudflare 节点（香港）
    ↓
边缘处理（API响应 < 50ms）
    ↓
需要时 → TiDB Cloud（新加坡）
```

**大数据应用**：
- 实时数据预处理（在边缘过滤无效数据）
- 降低中心服务器压力
- GDPR/数据主权合规（数据不离境）

### 5.2 流式处理思维

```javascript
// GPS 数据流处理（可复用于物联网传感器）
navigator.geolocation.watchPosition(
    (pos) => {
        // 1. 数据清洗（过滤异常值）
        if (pos.speed > 120) return; // 地铁不可能120km/h
        
        // 2. 实时计算（速度、方向、距离）
        const distance = calculateDistance(pos, destination);
        
        // 3. 阈值触发（事件驱动）
        if (distance < PROXIMITY_THRESHOLD) {
            triggerAlert();
        }
        
        // 4. 批上报（减少API调用）
        buffer.push(pos);
        if (buffer.length >= BATCH_SIZE) flush();
    }
);
```

---

## 六、调试与监控方法论

### 6.1 日志分级体系

```javascript
// 生产环境日志规范
const LOG_LEVELS = {
    ERROR:   ['GPS错误', 'API失败'],      // 立即告警
    WARNING: ['定位超时', '缓存失效'],    // 监控关注
    INFO:    ['开始追踪', '到站提醒'],    // 用户行为分析
    DEBUG:   ['坐标更新', '速度计算']     // 仅开发模式
};

// 日志输出规范（便于ELK/ Datadog分析）
[2026-03-11 15:30:45] [INFO] [GPS] lat:22.3 lng:114.2 speed:45km/h
[2026-03-11 15:31:12] [ERROR] [GPS] Timeout: 信号弱, 建议: 移动至窗边
```

### 6.2 故障排查决策树

```
App 无法定位？
    │
    ├─ 检查权限 → 设置 > 应用 > 松鼠提醒 > 位置权限
    │
    ├─ 检查信号 → 到室外/窗边测试
    │
    ├─ 检查设置 → 省电模式可能关闭GPS
    │
    └─ 检查版本 → 旧版本可能有bug
```

---

## 七、团队协作规范

### 7.1 文档先行（Documentation First）

```
项目启动
    │
    ├─ README.md（快速开始）
    ├─ ARCHITECTURE.md（架构决策记录 ADR）
    ├─ DEPLOYMENT.md（部署指南）
    └─ METHODOLOGY.md（本 Playbook）
```

### 7.2 知识传承清单

**交接时必须包含**：
1. **凭证清单**（API Keys、数据库连接串）
2. **环境变量**（.env 文件模板）
3. **已知问题**（当前版本限制）
4. **回滚方案**（上一个稳定版本）

---

## 八、未来扩展路线图

### 8.1 技术债务清单

| 优先级 | 问题 | 解决方案 | 预估工时 |
|-------|------|---------|---------|
| P0 | GPS后台运行 | 添加Foreground Service | 4h |
| P1 | 智能路径预测 | ML模型训练 | 16h |
| P1 | 离线模式 | Service Worker缓存 | 8h |
| P2 | 多语言支持 | i18n国际化 | 4h |

### 8.2 可复用组件

**已开发的可复用模块**：
- ✅ WebView + Native Bridge 模板
- ✅ GitHub Actions Android 构建流程
- ✅ Cloudflare Workers API 模板
- ✅ GPS追踪核心算法

---

## 九、关键决策记录（ADR）

### ADR-001: 为什么选择 WebView 而不是原生 Android？
**决策**：使用 WebView 混合开发
**原因**：
1. 快速迭代（无需Google Play审核）
2. 一套代码可扩展为iOS版本
3. Web技术栈更易招聘
**权衡**：性能略低于原生，但可接受

### ADR-002: 为什么使用 Cloudflare 而不是 AWS/Azure？
**决策**：使用 Cloudflare Workers + R2
**原因**：
1. 免费额度 generous
2. 香港节点延迟低
3. 无需管理服务器
**权衡**：vendor lock-in风险

---

## 十、总结：工程师的思维升级

### 10.1 从"功能实现"到"系统设计"
```
初级：实现GPS定位
中级：处理GPS各种错误情况
高级：设计"无感"体验，自动优化资源消耗
```

### 10.2 从"本地开发"到"云端交付"
```
传统：本地编译 → 手动安装
现代：Git push → CI/CD自动构建 → 自动部署 → 监控告警
```

### 10.3 从"解决当前问题"到"预防未来问题"
```
点：修复GPS timeout
线：优化所有网络请求的超时配置
面：建立完整的错误处理框架
体：形成可复用的方法论（本Playbook）
```

---

## 附录：快速检查清单

**发布前检查**：
- [ ] 所有依赖版本锁定
- [ ] 敏感信息未提交到Git
- [ ] 生产环境API地址正确
- [ ] 错误处理覆盖所有边界情况
- [ ] 日志不包含敏感信息
- [ ] 权限申请符合最小原则

**监控指标**：
- GPS定位成功率 > 95%
- API响应时间 < 200ms
- 崩溃率 < 0.1%
- 电池消耗 < 5%/小时

---

> 文档版本：v1.0
> 最后更新：2026-03-11
> 作者：松鼠 Reminder 团队
> 许可：内部使用，可分享经验
