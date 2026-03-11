"""
生活事件检测器
检测换工作、搬家、失业等生活事件
"""

import logging
from datetime import datetime, timedelta
from enum import Enum
from typing import Dict, List, Optional, Tuple

import numpy as np
from scipy import stats

from src.db.tidb_client import TiDBClient
from src.db.redis_client import RedisClient
from src.models import LifeEvent, EventType, EventConfidence

logger = logging.getLogger(__name__)


class EventDetector:
    """事件检测基类"""
    
    def __init__(self, name: str):
        self.name = name
    
    async def detect(self, user_id: str, data: Dict) -> Optional[LifeEvent]:
        """检测事件 - 子类实现"""
        raise NotImplementedError


class JobChangeDetector(EventDetector):
    """换工作检测器"""
    
    def __init__(self):
        super().__init__("job_change")
    
    async def detect(self, user_id: str, data: Dict) -> Optional[LifeEvent]:
        """检测换工作信号:
        1. 工作地位置变化
        2. 通勤时间显著改变
        3. 通勤路线改变
        """
        work_history = data.get("work_history", [])
        commute_history = data.get("commute_history", [])
        
        if len(work_history) < 2:
            return None
        
        # 检查最近的工作地是否变化
        recent_work = work_history[-1]
        previous_work = work_history[-2]
        
        if self._is_location_change(previous_work, recent_work):
            # 工作地变化，可能是换工作
            return LifeEvent(
                user_id=user_id,
                event_type=EventType.JOB_CHANGE,
                confidence=EventConfidence.HIGH if self._verify_commute_change(commute_history) else EventConfidence.MEDIUM,
                detected_at=datetime.now(),
                event_date=recent_work.get("detected_at"),
                details={
                    "old_location": previous_work,
                    "new_location": recent_work,
                    "reason": "work_location_changed"
                },
                evidence=[
                    f"工作地从 {previous_work.get('grid_id')} 变为 {recent_work.get('grid_id')}",
                    f"通勤路线变化: {self._get_commute_change_desc(commute_history)}"
                ]
            )
        
        return None
    
    def _is_location_change(self, old: Dict, new: Dict) -> bool:
        """判断位置是否变化"""
        if not old or not new:
            return False
        
        old_grid = old.get("grid_id", "")
        new_grid = new.get("grid_id", "")
        
        # 计算距离
        distance = self._haversine(
            old.get("lat", 0), old.get("lng", 0),
            new.get("lat", 0), new.get("lng", 0)
        )
        
        # Grid不同且距离 > 500米
        return old_grid != new_grid and distance > 500
    
    def _verify_commute_change(self, commute_history: List[Dict]) -> bool:
        """验证通勤是否变化"""
        if len(commute_history) < 6:
            return False
        
        # 比较最近3周和前3周的通勤
        recent = commute_history[-3:]
        previous = commute_history[-6:-3]
        
        recent_duration = np.mean([c.get("duration_minutes", 0) for c in recent])
        previous_duration = np.mean([c.get("duration_minutes", 0) for c in previous])
        
        # 通勤时间变化 > 30%
        if previous_duration > 0:
            change_ratio = abs(recent_duration - previous_duration) / previous_duration
            return change_ratio > 0.3
        
        return False
    
    def _get_commute_change_desc(self, commute_history: List[Dict]) -> str:
        """获取通勤变化描述"""
        if len(commute_history) < 2:
            return "数据不足"
        
        recent_duration = commute_history[-1].get("duration_minutes", 0)
        previous_duration = commute_history[-2].get("duration_minutes", 0)
        
        if previous_duration > 0:
            change_pct = (recent_duration - previous_duration) / previous_duration * 100
            return f"{change_pct:+.1f}%"
        return "N/A"
    
    def _haversine(self, lat1: float, lng1: float, lat2: float, lng2: float) -> float:
        """计算两点间距离"""
        from math import radians, sin, cos, sqrt, atan2
        
        R = 6371000
        phi1, phi2 = radians(lat1), radians(lat2)
        dphi = radians(lat2 - lat1)
        dlambda = radians(lng2 - lng1)
        
        a = sin(dphi / 2) ** 2 + cos(phi1) * cos(phi2) * sin(dlambda / 2) ** 2
        return 2 * 6371000 * atan2(sqrt(a), sqrt(1 - a))


class MoveDetector(EventDetector):
    """搬家检测器"""
    
    def __init__(self):
        super().__init__("move")
    
    async def detect(self, user_id: str, data: Dict) -> Optional[LifeEvent]:
        """检测搬家信号:
        1. 夜间停留位置变化
        2. 新的居住模式建立
        """
        home_history = data.get("home_history", [])
        
        if len(home_history) < 2:
            return None
        
        recent_home = home_history[-1]
        previous_home = home_history[-2]
        
        if self._is_location_change(previous_home, recent_home):
            # 验证是否持续在新位置
            if recent_home.get("consecutive_days", 0) >= 7:
                return LifeEvent(
                    user_id=user_id,
                    event_type=EventType.MOVE,
                    confidence=EventConfidence.HIGH,
                    detected_at=datetime.now(),
                    event_date=recent_home.get("detected_at"),
                    details={
                        "old_home": previous_home,
                        "new_home": recent_home,
                        "consecutive_days": recent_home.get("consecutive_days", 0)
                    },
                    evidence=[
                        f"夜间停留位置从 {previous_home.get('grid_id')} 变为 {recent_home.get('grid_id')}",
                        f"新位置已连续停留 {recent_home.get('consecutive_days', 0)} 天"
                    ]
                )
        
        return None
    
    def _is_location_change(self, old: Dict, new: Dict) -> bool:
        """判断位置是否变化"""
        if not old or not new:
            return False
        
        old_grid = old.get("grid_id", "")
        new_grid = new.get("grid_id", "")
        
        # Grid变化即认为搬家
        return old_grid != new_grid and old_grid != "" and new_grid != ""


class UnemploymentDetector(EventDetector):
    """失业检测器"""
    
    def __init__(self):
        super().__init__("unemployment")
        # 阈值参数
        self.NO_COMMUTE_DAYS_THRESHOLD = 10  # 连续无通勤天数
        self.STAY_HOME_HOURS_THRESHOLD = 8   # 工作日在家时长
    
    async def detect(self, user_id: str, data: Dict) -> Optional[LifeEvent]:
        """检测失业信号:
        1. 连续多日无通勤记录
        2. 工作日长时间在家
        3. 原有工作地不再访问
        """
        recent_commutes = data.get("recent_commutes", [])
        daily_stats = data.get("daily_stats", [])
        work_location = data.get("current_work_location")
        
        # 1. 检查通勤缺失
        no_commute_days = self._count_no_commute_days(daily_stats)
        
        if no_commute_days >= self.NO_COMMUTE_DAYS_THRESHOLD:
            # 2. 检查是否长时间在家
            avg_home_hours = self._calculate_weekday_home_hours(daily_stats)
            
            if avg_home_hours > self.STAY_HOME_HOURS_THRESHOLD:
                return LifeEvent(
                    user_id=user_id,
                    event_type=EventType.UNEMPLOYMENT,
                    confidence=EventConfidence.HIGH,
                    detected_at=datetime.now(),
                    event_date=datetime.now() - timedelta(days=no_commute_days // 2),
                    details={
                        "no_commute_days": no_commute_days,
                        "avg_home_hours": avg_home_hours,
                        "last_commute": recent_commutes[-1] if recent_commutes else None
                    },
                    evidence=[
                        f"连续 {no_commute_days} 天无通勤记录",
                        f"工作日平均每天在家 {avg_home_hours:.1f} 小时",
                        "原工作地已多日未访问"
                    ]
                )
        
        return None
    
    def _count_no_commute_days(self, daily_stats: List[Dict]) -> int:
        """统计连续无通勤天数"""
        count = 0
        for stat in reversed(daily_stats):
            if stat.get("is_weekday", False) and stat.get("commute_count", 0) == 0:
                count += 1
            elif stat.get("is_weekday", False):
                break
        return count
    
    def _calculate_weekday_home_hours(self, daily_stats: List[Dict]) -> float:
        """计算工作日平均在家时长"""
        weekday_hours = []
        for stat in daily_stats:
            if stat.get("is_weekday", False):
                hours = stat.get("home_hours", 0)
                weekday_hours.append(hours)
        
        return np.mean(weekday_hours) if weekday_hours else 0


class RetirementDetector(EventDetector):
    """退休检测器"""
    
    def __init__(self):
        super().__init__("retirement")
    
    async def detect(self, user_id: str, data: Dict) -> Optional[LifeEvent]:
        """检测退休信号:
        1. 年龄估计 > 55
        2. 通勤停止
        3. 活动模式改变(非高峰出行)
        """
        estimated_age = data.get("estimated_age", 0)
        
        # 年龄过滤
        if estimated_age < 50:
            return None
        
        # 检查是否停止通勤
        daily_stats = data.get("daily_stats", [])
        recent_commutes = data.get("recent_commutes", [])
        
        no_commute_weeks = self._count_no_commute_weeks(daily_stats)
        
        if no_commute_weeks >= 4:  # 连续4周无通勤
            # 检查非高峰活动模式
            off_peak_ratio = self._calculate_off_peak_ratio(daily_stats)
            
            if off_peak_ratio > 0.7:  # 70%以上活动在非高峰
                return LifeEvent(
                    user_id=user_id,
                    event_type=EventType.RETIREMENT,
                    confidence=EventConfidence.MEDIUM if estimated_age < 60 else EventConfidence.HIGH,
                    detected_at=datetime.now(),
                    event_date=datetime.now() - timedelta(weeks=no_commute_weeks // 2),
                    details={
                        "estimated_age": estimated_age,
                        "no_commute_weeks": no_commute_weeks,
                        "off_peak_ratio": off_peak_ratio
                    },
                    evidence=[
                        f"估计年龄 {estimated_age} 岁",
                        f"连续 {no_commute_weeks} 周无通勤",
                        f"非高峰活动比例 {off_peak_ratio*100:.1f}%"
                    ]
                )
        
        return None
    
    def _count_no_commute_weeks(self, daily_stats: List[Dict]) -> int:
        """统计连续无通勤周数"""
        weeks_no_commute = 0
        current_week_no_commute = True
        
        for stat in reversed(daily_stats):
            if stat.get("commute_count", 0) > 0:
                if current_week_no_commute:
                    weeks_no_commute += 1
                    current_week_no_commute = False
            
            # 按周统计简化处理
            if len(daily_stats) - list(reversed(daily_stats)).index(stat) >= 7:
                break
        
        return weeks_no_commute
    
    def _calculate_off_peak_ratio(self, daily_stats: List[Dict]) -> float:
        """计算非高峰活动比例"""
        total = len(daily_stats)
        off_peak = sum(1 for s in daily_stats if s.get("is_off_peak_activity", False))
        return off_peak / total if total > 0 else 0


class LifeEventDetector:
    """生活事件检测主类"""
    
    def __init__(self, tidb: TiDBClient, redis: RedisClient):
        self.tidb = tidb
        self.redis = redis
        self.detectors: List[EventDetector] = [
            JobChangeDetector(),
            MoveDetector(),
            UnemploymentDetector(),
            RetirementDetector(),
        ]
    
    async def detect(self, user_id: str, days: int = 30) -> List[LifeEvent]:
        """检测用户的生活事件"""
        logger.info(f"Detecting life events for user {user_id}")
        
        # 获取用户数据
        user_data = await self._fetch_user_data(user_id, days)
        
        # 运行所有检测器
        events = []
        for detector in self.detectors:
            try:
                event = await detector.detect(user_id, user_data)
                if event:
                    events.append(event)
                    logger.info(f"Detected {detector.name} for user {user_id}")
            except Exception as e:
                logger.error(f"Detector {detector.name} failed: {e}")
        
        # 按置信度排序
        events.sort(key=lambda e: e.confidence.value, reverse=True)
        
        return events
    
    async def _fetch_user_data(self, user_id: str, days: int) -> Dict:
        """获取用户历史数据用于事件检测"""
        
        # 获取工作地历史
        work_history_query = """
            SELECT detected_at, grid_id, lat, lng, confidence
            FROM user_work_locations
            WHERE user_id = %s
            ORDER BY detected_at DESC
            LIMIT 10
        """
        work_history = await self.tidb.fetchall(work_history_query, (user_id,))
        
        # 获取居住地历史
        home_history_query = """
            SELECT detected_at, grid_id, lat, lng, confidence, consecutive_days
            FROM user_home_locations
            WHERE user_id = %s
            ORDER BY detected_at DESC
            LIMIT 10
        """
        home_history = await self.tidb.fetchall(home_history_query, (user_id,))
        
        # 获取通勤历史
        commute_query = """
            SELECT departure_time, duration_minutes, start_station_id, end_station_id
            FROM commute_records
            WHERE user_id = %s
              AND departure_time >= DATE_SUB(NOW(), INTERVAL %s DAY)
            ORDER BY departure_time
        """
        commute_history = await self.tidb.fetchall(commute_query, (user_id, days * 2))
        
        # 获取每日统计
        daily_stats_query = """
            SELECT date, is_weekday, commute_count, home_hours, 
                   is_off_peak_activity
            FROM user_daily_stats
            WHERE user_id = %s
              AND date >= DATE_SUB(NOW(), INTERVAL %s DAY)
            ORDER BY date
        """
        daily_stats = await self.tidb.fetchall(daily_stats_query, (user_id, days))
        
        return {
            "work_history": work_history,
            "home_history": home_history,
            "commute_history": commute_history,
            "daily_stats": daily_stats,
            "current_work_location": work_history[0] if work_history else None,
            "recent_commutes": commute_history[-30:] if len(commute_history) > 30 else commute_history,
            "estimated_age": await self._estimate_age(user_id)
        }
    
    async def _estimate_age(self, user_id: str) -> int:
        """根据行为模式估计年龄"""
        # 简单启发式：根据作息时间和活动类型估计
        query = """
            SELECT AVG(HOUR(first_activity_time)) as avg_start_hour,
                   AVG(night_activities) as avg_night_activities
            FROM user_daily_stats
            WHERE user_id = %s
              AND date >= DATE_SUB(NOW(), INTERVAL 30 DAY)
        """
        result = await self.tidb.fetchone(query, (user_id,))
        
        if not result:
            return 35  # 默认
        
        avg_start = result.get("avg_start_hour", 8)
        night_act = result.get("avg_night_activities", 0)
        
        # 早起 + 少夜间活动 = 偏年长
        if avg_start < 7 and night_act < 1:
            return 55
        elif avg_start > 9 and night_act > 3:
            return 25
        else:
            return 35
