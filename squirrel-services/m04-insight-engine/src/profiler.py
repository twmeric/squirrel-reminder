"""
用户画像分析器
实现Home/Work检测、通勤模式分析、职业推断等算法
"""

import logging
from dataclasses import dataclass
from datetime import datetime, timedelta, time
from typing import Dict, List, Optional, Tuple

import numpy as np
from sklearn.cluster import DBSCAN

from src.db.tidb_client import TiDBClient
from src.db.redis_client import RedisClient
from src.models import UserProfile, LocationProfile, CommutePattern, OccupationType

logger = logging.getLogger(__name__)


@dataclass
class StayPoint:
    """停留点数据类"""
    lat: float
    lng: float
    start_time: datetime
    end_time: datetime
    duration_minutes: int
    grid_id: str


@dataclass
class ClusterInfo:
    """聚类信息"""
    center_lat: float
    center_lng: float
    point_count: int
    night_visits: int
    day_visits: int
    weekday_visits: int
    weekend_visits: int
    avg_duration: float


class UserProfiler:
    """用户画像分析器"""
    
    # 算法参数
    HOME_NIGHT_START = time(22, 0)  # 22:00
    HOME_NIGHT_END = time(6, 0)     # 06:00
    WORK_START = time(9, 0)         # 09:00
    WORK_END = time(18, 0)          # 18:00
    
    DBSCAN_EPS = 200  # 200米
    DBSCAN_MIN_SAMPLES = 3
    
    def __init__(self, tidb: TiDBClient, redis: RedisClient):
        self.tidb = tidb
        self.redis = redis
    
    async def analyze(self, user_id: str, days: int = 30) -> UserProfile:
        """分析用户画像"""
        logger.info(f"Analyzing user {user_id} for last {days} days")
        
        # 获取停留点数据
        staypoints = await self._get_staypoints(user_id, days)
        
        if len(staypoints) < 5:
            logger.warning(f"Insufficient data for user {user_id}")
            return self._create_empty_profile(user_id)
        
        # 空间聚类
        clusters = self._cluster_locations(staypoints)
        
        # 检测Home/Work
        home_cluster = self._detect_home(clusters, staypoints)
        work_cluster = self._detect_work(clusters, staypoints, home_cluster)
        
        # 计算各项指标
        regularity = self._calculate_regularity(staypoints)
        mobility = self._calculate_mobility(staypoints)
        occupation = self._infer_occupation(regularity, mobility, home_cluster, work_cluster)
        
        # 获取通勤模式
        commute_pattern = await self.get_commute_pattern(user_id, days)
        
        # 构建用户画像
        profile = UserProfile(
            user_id=user_id,
            analysis_date=datetime.now(),
            analysis_days=days,
            home=LocationProfile(
                grid_id=home_cluster.get("grid_id", "") if home_cluster else "",
                lat=home_cluster.get("lat", 0.0) if home_cluster else 0.0,
                lng=home_cluster.get("lng", 0.0) if home_cluster else 0.0,
                confidence=home_cluster.get("confidence", 0.0) if home_cluster else 0.0,
                detection_method="night_clustering"
            ) if home_cluster else None,
            work=LocationProfile(
                grid_id=work_cluster.get("grid_id", "") if work_cluster else "",
                lat=work_cluster.get("lat", 0.0) if work_cluster else 0.0,
                lng=work_cluster.get("lng", 0.0) if work_cluster else 0.0,
                confidence=work_cluster.get("confidence", 0.0) if work_cluster else 0.0,
                detection_method="weekday_clustering"
            ) if work_cluster else None,
            regularity_score=regularity,
            mobility_score=mobility,
            occupation=occupation,
            commute_pattern=commute_pattern,
            total_staypoints=len(staypoints),
            unique_locations=len(clusters)
        )
        
        return profile
    
    async def get_commute_pattern(self, user_id: str, days: int = 30) -> CommutePattern:
        """获取通勤模式"""
        # 获取通勤记录
        commutes = await self._get_commutes(user_id, days)
        
        if len(commutes) < 3:
            return CommutePattern(is_commuter=False)
        
        # 计算平均通勤时间和标准差
        durations = [c["duration_minutes"] for c in commutes]
        departure_times = [c["departure_hour"] for c in commutes]
        
        avg_duration = np.mean(durations)
        std_duration = np.std(durations)
        avg_departure = np.mean(departure_times)
        
        # 计算规律性 (变异系数的倒数)
        cv_duration = std_duration / avg_duration if avg_duration > 0 else 1.0
        regularity = max(0, 1 - cv_duration)
        
        # 检测是否有固定路线
        route_consistency = self._calculate_route_consistency(commutes)
        
        return CommutePattern(
            is_commuter=True,
            avg_commute_minutes=float(avg_duration),
            regularity_score=float(regularity),
            route_consistency=float(route_consistency),
            avg_departure_hour=float(avg_departure),
            commute_count=len(commutes)
        )
    
    async def _get_staypoints(self, user_id: str, days: int) -> List[StayPoint]:
        """从数据库获取停留点"""
        query = """
            SELECT center_lat, center_lng, start_time, end_time, 
                   TIMESTAMPDIFF(MINUTE, start_time, end_time) as duration,
                   grid_id
            FROM staypoints
            WHERE user_id = %s 
              AND start_time >= DATE_SUB(NOW(), INTERVAL %s DAY)
            ORDER BY start_time
        """
        
        rows = await self.tidb.fetchall(query, (user_id, days))
        
        staypoints = []
        for row in rows:
            staypoints.append(StayPoint(
                lat=row["center_lat"],
                lng=row["center_lng"],
                start_time=row["start_time"],
                end_time=row["end_time"],
                duration_minutes=row["duration"],
                grid_id=row["grid_id"]
            ))
        
        return staypoints
    
    async def _get_commutes(self, user_id: str, days: int) -> List[Dict]:
        """获取通勤记录"""
        query = """
            SELECT departure_time, arrival_time, 
                   TIMESTAMPDIFF(MINUTE, departure_time, arrival_time) as duration,
                   HOUR(departure_time) as departure_hour,
                   start_station_id, end_station_id
            FROM commute_records
            WHERE user_id = %s
              AND departure_time >= DATE_SUB(NOW(), INTERVAL %s DAY)
              AND is_valid = 1
            ORDER BY departure_time
        """
        
        rows = await self.tidb.fetchall(query, (user_id, days))
        return [dict(row) for row in rows]
    
    def _cluster_locations(self, staypoints: List[StayPoint]) -> List[ClusterInfo]:
        """使用DBSCAN聚类位置"""
        if len(staypoints) < 3:
            return []
        
        # 提取坐标
        coords = np.array([[sp.lat, sp.lng] for sp in staypoints])
        
        # DBSCAN聚类
        clustering = DBSCAN(
            eps=self.DBSCAN_EPS / 111000,  # 米转度 (约)
            min_samples=self.DBSCAN_MIN_SAMPLES
        ).fit(coords)
        
        labels = clustering.labels_
        unique_labels = set(labels)
        
        clusters = []
        for label in unique_labels:
            if label == -1:  # 噪声点
                continue
            
            cluster_points = [sp for sp, l in zip(staypoints, labels) if l == label]
            
            # 统计
            night_visits = sum(1 for sp in cluster_points 
                             if self._is_night_time(sp.start_time))
            day_visits = len(cluster_points) - night_visits
            weekday_visits = sum(1 for sp in cluster_points 
                               if sp.start_time.weekday() < 5)
            weekend_visits = len(cluster_points) - weekday_visits
            avg_duration = np.mean([sp.duration_minutes for sp in cluster_points])
            
            # 计算中心
            center_lat = np.mean([sp.lat for sp in cluster_points])
            center_lng = np.mean([sp.lng for sp in cluster_points])
            
            clusters.append(ClusterInfo(
                center_lat=center_lat,
                center_lng=center_lng,
                point_count=len(cluster_points),
                night_visits=night_visits,
                day_visits=day_visits,
                weekday_visits=weekday_visits,
                weekend_visits=weekend_visits,
                avg_duration=avg_duration
            ))
        
        return clusters
    
    def _detect_home(self, clusters: List[ClusterInfo], staypoints: List[StayPoint]) -> Optional[Dict]:
        """检测家位置 - 基于夜间停留"""
        if not clusters:
            return None
        
        # 给每个聚类打分
        scores = []
        for cluster in clusters:
            score = 0.0
            
            # 夜间停留权重最高
            night_ratio = cluster.night_visits / cluster.point_count if cluster.point_count > 0 else 0
            score += night_ratio * 0.5
            
            # 停留时间长加分
            if cluster.avg_duration > 360:  # > 6小时
                score += 0.2
            elif cluster.avg_duration > 180:  # > 3小时
                score += 0.1
            
            # 访问频率
            if cluster.point_count >= 10:
                score += 0.15
            elif cluster.point_count >= 5:
                score += 0.1
            
            # 周末访问
            weekend_ratio = cluster.weekend_visits / cluster.point_count if cluster.point_count > 0 else 0
            score += weekend_ratio * 0.05
            
            scores.append((cluster, score))
        
        # 选择得分最高的
        scores.sort(key=lambda x: x[1], reverse=True)
        best_cluster, best_score = scores[0]
        
        # 找到对应的grid_id
        grid_id = self._find_nearest_grid(staypoints, best_cluster.center_lat, best_cluster.center_lng)
        
        return {
            "lat": best_cluster.center_lat,
            "lng": best_cluster.center_lng,
            "grid_id": grid_id,
            "confidence": min(1.0, best_score),
            "night_visits": best_cluster.night_visits,
            "total_visits": best_cluster.point_count
        }
    
    def _detect_work(self, clusters: List[ClusterInfo], staypoints: List[StayPoint], 
                     home: Optional[Dict]) -> Optional[Dict]:
        """检测工作位置 - 基于工作日白天停留"""
        if not clusters:
            return None
        
        # 排除home
        home_coords = None
        if home:
            home_coords = (home["lat"], home["lng"])
        
        scores = []
        for cluster in clusters:
            # 排除与home太近的
            if home_coords:
                dist = self._haversine(home_coords[0], home_coords[1], 
                                      cluster.center_lat, cluster.center_lng)
                if dist < 500:  # < 500米认为是同一个地方
                    continue
            
            score = 0.0
            
            # 工作日访问权重
            weekday_ratio = cluster.weekday_visits / cluster.point_count if cluster.point_count > 0 else 0
            score += weekday_ratio * 0.4
            
            # 白天停留
            day_ratio = cluster.day_visits / cluster.point_count if cluster.point_count > 0 else 0
            score += day_ratio * 0.2
            
            # 停留时间
            if 120 <= cluster.avg_duration <= 600:  # 2-10小时
                score += 0.25
            
            # 访问频率
            if cluster.point_count >= 10:
                score += 0.15
            
            scores.append((cluster, score))
        
        if not scores:
            return None
        
        scores.sort(key=lambda x: x[1], reverse=True)
        best_cluster, best_score = scores[0]
        
        # 找到对应的grid_id
        grid_id = self._find_nearest_grid(staypoints, best_cluster.center_lat, best_cluster.center_lng)
        
        return {
            "lat": best_cluster.center_lat,
            "lng": best_cluster.center_lng,
            "grid_id": grid_id,
            "confidence": min(1.0, best_score),
            "weekday_visits": best_cluster.weekday_visits,
            "total_visits": best_cluster.point_count
        }
    
    def _calculate_regularity(self, staypoints: List[StayPoint]) -> float:
        """计算规律性分数 (0-1)"""
        if len(staypoints) < 7:
            return 0.5
        
        # 计算每天访问的位置数
        daily_locations = {}
        for sp in staypoints:
            date = sp.start_time.date()
            if date not in daily_locations:
                daily_locations[date] = set()
            daily_locations[date].add(sp.grid_id)
        
        if len(daily_locations) < 2:
            return 0.5
        
        location_counts = [len(s) for s in daily_locations.values()]
        cv = np.std(location_counts) / np.mean(location_counts) if np.mean(location_counts) > 0 else 1.0
        
        # 变异系数越低，规律性越高
        regularity = max(0, 1 - cv)
        return float(regularity)
    
    def _calculate_mobility(self, staypoints: List[StayPoint]) -> float:
        """计算活跃度分数 (0-1)"""
        if len(staypoints) < 3:
            return 0.0
        
        # 计算日均位置数
        daily_locations = {}
        for sp in staypoints:
            date = sp.start_time.date()
            if date not in daily_locations:
                daily_locations[date] = set()
            daily_locations[date].add(sp.grid_id)
        
        avg_locations = np.mean([len(s) for s in daily_locations.values()])
        
        # 归一化到0-1 (假设正常人日均3-10个位置)
        mobility = min(1.0, avg_locations / 8.0)
        return float(mobility)
    
    def _infer_occupation(self, regularity: float, mobility: float, 
                          home: Optional[Dict], work: Optional[Dict]) -> OccupationType:
        """推断职业类型"""
        
        # 有固定工作地 + 高规律性 = 上班族
        if work and regularity > 0.6:
            return OccupationType.COMMUTER
        
        # 无固定工作地 + 高移动性 = 外勤
        if not work and mobility > 0.6:
            return OccupationType.FIELD_WORKER
        
        # 低规律性 + 低移动性 = 自由职业/居家
        if regularity < 0.4 and mobility < 0.4:
            return OccupationType.FREELANCER
        
        # 默认
        return OccupationType.OTHER
    
    def _calculate_route_consistency(self, commutes: List[Dict]) -> float:
        """计算路线一致性"""
        if len(commutes) < 2:
            return 0.0
        
        # 统计最常走路线占比
        route_counts = {}
        for c in commutes:
            route = (c.get("start_station_id"), c.get("end_station_id"))
            route_counts[route] = route_counts.get(route, 0) + 1
        
        if not route_counts:
            return 0.0
        
        max_count = max(route_counts.values())
        consistency = max_count / len(commutes)
        
        return float(consistency)
    
    def _is_night_time(self, dt: datetime) -> bool:
        """判断是否为夜间"""
        t = dt.time()
        return t >= self.HOME_NIGHT_START or t <= self.HOME_NIGHT_END
    
    def _find_nearest_grid(self, staypoints: List[StayPoint], lat: float, lng: float) -> str:
        """找到最近的grid_id"""
        min_dist = float('inf')
        best_grid = ""
        
        for sp in staypoints:
            dist = self._haversine(lat, lng, sp.lat, sp.lng)
            if dist < min_dist:
                min_dist = dist
                best_grid = sp.grid_id
        
        return best_grid
    
    def _haversine(self, lat1: float, lng1: float, lat2: float, lng2: float) -> float:
        """计算两点间距离(米)"""
        from math import radians, sin, cos, sqrt, atan2
        
        R = 6371000  # 地球半径(米)
        
        phi1 = radians(lat1)
        phi2 = radians(lat2)
        dphi = radians(lat2 - lat1)
        dlambda = radians(lng2 - lng1)
        
        a = sin(dphi / 2) ** 2 + cos(phi1) * cos(phi2) * sin(dlambda / 2) ** 2
        c = 2 * atan2(sqrt(a), sqrt(1 - a))
        
        return R * c
    
    def _create_empty_profile(self, user_id: str) -> UserProfile:
        """创建空画像"""
        return UserProfile(
            user_id=user_id,
            analysis_date=datetime.now(),
            analysis_days=0,
            regularity_score=0.0,
            mobility_score=0.0,
            occupation=OccupationType.UNKNOWN,
            total_staypoints=0,
            unique_locations=0
        )
