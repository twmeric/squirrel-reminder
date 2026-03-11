"""
用户画像分析器测试
"""

import pytest
from datetime import datetime, timedelta
from unittest.mock import AsyncMock, MagicMock

from src.profiler import UserProfiler, StayPoint, ClusterInfo
from src.models import OccupationType, UserProfile


@pytest.fixture
def mock_tidb():
    """模拟TiDB客户端"""
    return AsyncMock()


@pytest.fixture
def mock_redis():
    """模拟Redis客户端"""
    return AsyncMock()


@pytest.fixture
def profiler(mock_tidb, mock_redis):
    """创建测试用的Profiler实例"""
    return UserProfiler(mock_tidb, mock_redis)


@pytest.fixture
def sample_staypoints():
    """生成测试用的停留点"""
    base_time = datetime.now() - timedelta(days=30)
    
    # 家位置 (科技园附近) - 主要在夜间
    home_points = []
    for i in range(30):
        # 每晚22:00-06:00在家
        night_start = base_time + timedelta(days=i, hours=22)
        home_points.append(StayPoint(
            lat=22.5431,
            lng=113.9589,
            start_time=night_start,
            end_time=night_start + timedelta(hours=8),
            duration_minutes=480,
            grid_id="grid_2254_11395"
        ))
    
    # 工作位置 (CBD) - 主要在工作日白天
    work_points = []
    for i in range(30):
        if i % 7 < 5:  # 工作日
            day_start = base_time + timedelta(days=i, hours=9)
            work_points.append(StayPoint(
                lat=22.5372,
                lng=114.0654,
                start_time=day_start,
                end_time=day_start + timedelta(hours=9),
                duration_minutes=540,
                grid_id="grid_2253_11406"
            ))
    
    return home_points + work_points


@pytest.mark.asyncio
async def test_analyze_with_sufficient_data(profiler, mock_tidb, sample_staypoints):
    """测试有足够数据时的分析"""
    # 设置mock返回值
    mock_tidb.fetchall.return_value = [
        {
            "center_lat": sp.lat,
            "center_lng": sp.lng,
            "start_time": sp.start_time,
            "end_time": sp.end_time,
            "duration": sp.duration_minutes,
            "grid_id": sp.grid_id
        }
        for sp in sample_staypoints
    ]
    
    profile = await profiler.analyze("user_001", 30)
    
    assert isinstance(profile, UserProfile)
    assert profile.user_id == "user_001"
    assert profile.home is not None
    assert profile.work is not None
    assert profile.occupation == OccupationType.COMMUTER


@pytest.mark.asyncio
async def test_analyze_with_insufficient_data(profiler, mock_tidb):
    """测试数据不足时的分析"""
    mock_tidb.fetchall.return_value = [
        {
            "center_lat": 22.5,
            "center_lng": 113.9,
            "start_time": datetime.now(),
            "end_time": datetime.now() + timedelta(minutes=30),
            "duration": 30,
            "grid_id": "grid_1"
        }
    ]
    
    profile = await profiler.analyze("user_002", 30)
    
    assert profile.total_staypoints < 5
    assert profile.home is None
    assert profile.work is None
    assert profile.occupation == OccupationType.UNKNOWN


@pytest.mark.asyncio
async def test_detect_home(profiler, sample_staypoints):
    """测试家位置检测"""
    clusters = profiler._cluster_locations(sample_staypoints)
    home = profiler._detect_home(clusters, sample_staypoints)
    
    assert home is not None
    assert home["grid_id"] == "grid_2254_11395"
    assert home["confidence"] > 0.7


@pytest.mark.asyncio
async def test_detect_work(profiler, sample_staypoints):
    """测试工作位置检测"""
    clusters = profiler._cluster_locations(sample_staypoints)
    home = profiler._detect_home(clusters, sample_staypoints)
    work = profiler._detect_work(clusters, sample_staypoints, home)
    
    assert work is not None
    assert work["grid_id"] == "grid_2253_11406"
    assert work["weekday_visits"] > work.get("weekend_visits", 0)


def test_calculate_regularity(profiler, sample_staypoints):
    """测试规律性计算"""
    regularity = profiler._calculate_regularity(sample_staypoints)
    
    assert 0 <= regularity <= 1
    # 测试数据很规律，应该>0.6
    assert regularity > 0.6


def test_calculate_mobility(profiler, sample_staypoints):
    """测试活跃度计算"""
    mobility = profiler._calculate_mobility(sample_staypoints)
    
    assert 0 <= mobility <= 1


def test_infer_occupation_commuter(profiler):
    """测试上班族识别"""
    occupation = profiler._infer_occupation(
        regularity=0.8,
        mobility=0.4,
        home={"grid_id": "home"},
        work={"grid_id": "work"}
    )
    
    assert occupation == OccupationType.COMMUTER


def test_infer_occupation_field_worker(profiler):
    """测试外勤人员识别"""
    occupation = profiler._infer_occupation(
        regularity=0.3,
        mobility=0.8,
        home=None,
        work=None
    )
    
    assert occupation == OccupationType.FIELD_WORKER


def test_haversine_distance(profiler):
    """测试距离计算"""
    # 科技园到深大，大约1公里
    dist = profiler._haversine(22.5431, 113.9589, 22.5268, 113.9800)
    assert 1000 < dist < 3000


@pytest.mark.asyncio
async def test_get_commute_pattern(profiler, mock_tidb):
    """测试通勤模式获取"""
    mock_tidb.fetchall.return_value = [
        {
            "departure_time": datetime.now() - timedelta(days=i),
            "duration_minutes": 35,
            "departure_hour": 8,
            "start_station_id": "L1_S18",
            "end_station_id": "L1_S08"
        }
        for i in range(1, 20)
    ]
    
    pattern = await profiler.get_commute_pattern("user_001", 30)
    
    assert pattern.is_commuter
    assert pattern.avg_commute_minutes > 0
    assert pattern.regularity_score > 0


class TestClusterLocations:
    """测试聚类功能"""
    
    def test_cluster_clear_groups(self, profiler):
        """测试清晰的聚类"""
        # 创建两组明显分离的点
        points = []
        # 家位置
        for i in range(10):
            points.append(StayPoint(
                lat=22.5431 + (i % 3 - 1) * 0.0001,
                lng=113.9589 + (i % 3 - 1) * 0.0001,
                start_time=datetime.now() - timedelta(days=i, hours=12),
                end_time=datetime.now() - timedelta(days=i, hours=8),
                duration_minutes=240,
                grid_id=f"home_{i}"
            ))
        # 工作位置
        for i in range(10):
            points.append(StayPoint(
                lat=22.5372 + (i % 3 - 1) * 0.0001,
                lng=114.0654 + (i % 3 - 1) * 0.0001,
                start_time=datetime.now() - timedelta(days=i),
                end_time=datetime.now() - timedelta(days=i, hours=-8),
                duration_minutes=480,
                grid_id=f"work_{i}"
            ))
        
        clusters = profiler._cluster_locations(points)
        assert len(clusters) >= 2


class TestPerformance:
    """性能测试"""
    
    @pytest.mark.benchmark
    def test_analyze_performance(self, profiler, sample_staypoints):
        """测试分析性能 - 应能在100ms内处理1000个点"""
        import time
        
        # 生成大量数据
        large_dataset = []
        for i in range(1000):
            large_dataset.append(StayPoint(
                lat=22.5 + (i % 100) * 0.001,
                lng=113.9 + (i % 100) * 0.001,
                start_time=datetime.now() - timedelta(days=i//33, hours=i%24),
                end_time=datetime.now() - timedelta(days=i//33, hours=(i%24)-1),
                duration_minutes=60,
                grid_id=f"grid_{i}"
            ))
        
        start = time.time()
        clusters = profiler._cluster_locations(large_dataset)
        duration = time.time() - start
        
        assert duration < 1.0  # 1秒内完成
