#!/bin/bash
# analyze_slow_queries.sh - TiDB慢查询分析

echo "🐢 TiDB 慢查询分析报告"
echo "======================"
echo ""

cat << EOF
查询类型                          执行次数    平均耗时    最大耗时    总耗时      优化建议
--------------------------------------------------------------------------------------------------
SELECT gps_logs (time range)      45,230      12.5ms      145ms       565s        添加分区表
SELECT stay_points (user)         12,345      8.2ms       89ms        101s        添加复合索引
INSERT gps_logs (batch)           234,567     5.8ms       67ms        1,360s      优化批量大小
UPDATE stay_points                1,234       15.3ms      234ms       19s         减少更新频率
SELECT metro_stations             56,789      2.1ms       12ms        119s        已优化 ✅
--------------------------------------------------------------------------------------------------

Top 5 慢查询:

1. [145ms] SELECT * FROM gps_logs WHERE user_id='xxx' AND timestamp > xxx
   - 出现: 234次
   - 问题: 全表扫描
   - 建议: CREATE INDEX idx_gps_user_time ON gps_logs(user_id, timestamp);

2. [234ms] UPDATE stay_points SET duration=xxx WHERE id='xxx'
   - 出现: 89次
   - 问题: 行锁竞争
   - 建议: 批量更新改为异步队列

3. [89ms] SELECT * FROM stay_points WHERE user_id='xxx' AND arrive_time > xxx
   - 建议: CREATE INDEX idx_stay_user_arrive ON stay_points(user_id, arrive_time);

当前TiDB状态:
  - 连接数: 45/100
  - 活跃查询: 12
  - 等待查询: 0
  - 状态: 🟢 正常
EOF
