"""
生活事件检测器测试
"""

import pytest
from datetime import datetime, timedelta
from unittest.mock import AsyncMock, MagicMock

from src.event_detector import (
    JobChangeDetector, MoveDetector, UnemploymentDetector,
    RetirementDetector, LifeEventDetector
)
from src.models import EventType, EventConfidence


@pytest.fixture
def mock_tidb():
    return AsyncMock()


@pytest.fixture
def mock_redis():
    return AsyncMock()


@pytest.fixture
def detector(mock_tidb, mock_redis):
    return LifeEventDetector(mock_tidb, mock_redis)


class TestJobChangeDetector:
    """测试换工作检测"""
    
    @pytest.fixture
    def job_detector(self):
        return JobChangeDetector()
    
    @pytest.mark.asyncio
    async def test_detect_job_change_with_location_change(self, job_detector):
        """测试检测到工作地点变化"""
        data = {
            "work_history": [
                {
                    "detected_at": datetime.now() - timedelta(days=60),
                    "grid_id": "grid_old",
                    "lat": 22.5431,
                    "lng": 113.9589,
                    "confidence": 0.9
                },
                {
                    "detected_at": datetime.now() - timedelta(days=10),
                    "grid_id": "grid_new",
                    "lat": 22.5372,
                    "lng": 114.0654,
                    "confidence": 0.85
                }
            ],
            "commute_history": [
                {"duration_minutes": 30, "departure_hour": 8},
                {"duration_minutes": 30, "departure_hour": 8},
                {"duration_minutes": 45, "departure_hour": 8},  # 新通勤时间
                {"duration_minutes": 45, "departure_hour": 8},
            ]
        }
        
        event = await job_detector.detect("user_001", data)
        
        assert event is not None
        assert event.event_type == EventType.JOB_CHANGE
        assert event.confidence in [EventConfidence.MEDIUM, EventConfidence.HIGH]
    
    @pytest.mark.asyncio
    async def test_no_job_change(self, job_detector):
        """测试无工作变化"""
        data = {
            "work_history": [
                {
                    "detected_at": datetime.now() - timedelta(days=30),
                    "grid_id": "grid_same",
                    "lat": 22.5431,
                    "lng": 113.9589
                }
            ]
        }
        
        event = await job_detector.detect("user_002", data)
        assert event is None


class TestMoveDetector:
    """测试搬家检测"""
    
    @pytest.fixture
    def move_detector(self):
        return MoveDetector()
    
    @pytest.mark.asyncio
    async def test_detect_move(self, move_detector):
        """测试检测到搬家"""
        data = {
            "home_history": [
                {
                    "detected_at": datetime.now() - timedelta(days=60),
                    "grid_id": "old_home",
                    "lat": 22.5431,
                    "lng": 113.9589,
                    "consecutive_days": 30
                },
                {
                    "detected_at": datetime.now() - timedelta(days=10),
                    "grid_id": "new_home",
                    "lat": 22.5202,
                    "lng": 113.9266,
                    "consecutive_days": 10
                }
            ]
        }
        
        event = await move_detector.detect("user_003", data)
        
        assert event is not None
        assert event.event_type == EventType.MOVE
        assert event.confidence == EventConfidence.HIGH
    
    @pytest.mark.asyncio
    async def test_short_stay_not_move(self, move_detector):
        """测试短期停留不算搬家"""
        data = {
            "home_history": [
                {
                    "detected_at": datetime.now() - timedelta(days=60),
                    "grid_id": "home",
                    "lat": 22.5431,
                    "lng": 113.9589,
                    "consecutive_days": 30
                },
                {
                    "detected_at": datetime.now() - timedelta(days=2),
                    "grid_id": "temp",
                    "lat": 22.5202,
                    "lng": 113.9266,
                    "consecutive_days": 2  # 太短
                }
            ]
        }
        
        event = await move_detector.detect("user_004", data)
        assert event is None


class TestUnemploymentDetector:
    """测试失业检测"""
    
    @pytest.fixture
    def unemployment_detector(self):
        return UnemploymentDetector()
    
    @pytest.mark.asyncio
    async def test_detect_unemployment(self, unemployment_detector):
        """测试检测到失业"""
        # 生成连续无通勤的工作日
        daily_stats = []
        for i in range(15):
            daily_stats.append({
                "is_weekday": i % 7 < 5,
                "commute_count": 0,  # 无通勤
                "home_hours": 10     # 长时间在家
            })
        
        data = {
            "daily_stats": daily_stats,
            "recent_commutes": [
                {"duration_minutes": 30}  # 只有一条旧记录
            ],
            "current_work_location": {"grid_id": "work"}
        }
        
        event = await unemployment_detector.detect("user_005", data)
        
        assert event is not None
        assert event.event_type == EventType.UNEMPLOYMENT
        assert event.confidence == EventConfidence.HIGH
    
    @pytest.mark.asyncio
    async def test_vacation_not_unemployment(self, unemployment_detector):
        """测试假期不算失业"""
        daily_stats = []
        for i in range(5):  # 只有5天，不够长
            daily_stats.append({
                "is_weekday": True,
                "commute_count": 0,
                "home_hours": 10
            })
        
        data = {
            "daily_stats": daily_stats,
            "recent_commutes": []
        }
        
        event = await unemployment_detector.detect("user_006", data)
        assert event is None


class TestRetirementDetector:
    """测试退休检测"""
    
    @pytest.fixture
    def retirement_detector(self):
        return RetirementDetector()
    
    @pytest.mark.asyncio
    async def test_detect_retirement(self, retirement_detector):
        """测试检测到退休"""
        # 生成4周无通勤、非高峰活动
        daily_stats = []
        for i in range(30):
            daily_stats.append({
                "is_weekday": i % 7 < 5,
                "commute_count": 0,
                "is_off_peak_activity": True  # 非高峰活动
            })
        
        data = {
            "estimated_age": 60,
            "daily_stats": daily_stats,
            "recent_commutes": []
        }
        
        event = await retirement_detector.detect("user_007", data)
        
        assert event is not None
        assert event.event_type == EventType.RETIREMENT
        assert event.confidence == EventConfidence.HIGH
    
    @pytest.mark.asyncio
    async def test_young_not_retirement(self, retirement_detector):
        """测试年轻人不会误判"""
        data = {
            "estimated_age": 30,  # 太年轻
            "daily_stats": [],
            "recent_commutes": []
        }
        
        event = await retirement_detector.detect("user_008", data)
        assert event is None


class TestLifeEventDetector:
    """测试主检测器"""
    
    @pytest.mark.asyncio
    async def test_detect_multiple_events(self, detector, mock_tidb):
        """测试同时检测多种事件"""
        # 设置mock数据
        mock_tidb.fetchall.side_effect = [
            # work_history
            [
                {"detected_at": datetime.now() - timedelta(days=30), "grid_id": "work1", "lat": 22.5, "lng": 113.9, "confidence": 0.9},
                {"detected_at": datetime.now() - timedelta(days=5), "grid_id": "work2", "lat": 22.6, "lng": 114.0, "confidence": 0.85}
            ],
            # home_history
            [
                {"detected_at": datetime.now() - timedelta(days=60), "grid_id": "home1", "lat": 22.54, "lng": 113.95, "consecutive_days": 30},
                {"detected_at": datetime.now() - timedelta(days=10), "grid_id": "home2", "lat": 22.52, "lng": 113.92, "consecutive_days": 10}
            ],
            # commute_history
            [
                {"duration_minutes": 30, "departure_time": datetime.now() - timedelta(days=40)}
            ],
            # daily_stats
            []
        ]
        mock_tidb.fetchone.return_value = {"avg_start_hour": 8, "avg_night_activities": 1}
        
        events = await detector.detect("user_009", 30)
        
        # 应该检测到换工作和搬家
        event_types = [e.event_type for e in events]
        assert EventType.JOB_CHANGE in event_types
        assert EventType.MOVE in event_types
    
    @pytest.mark.asyncio
    async def test_detect_no_events(self, detector, mock_tidb):
        """测试无事件情况"""
        mock_tidb.fetchall.side_effect = [
            [],  # work_history
            [],  # home_history
            [],  # commute_history
            []   # daily_stats
        ]
        mock_tidb.fetchone.return_value = None
        
        events = await detector.detect("user_010", 30)
        
        assert len(events) == 0


class TestPerformance:
    """性能测试"""
    
    @pytest.mark.asyncio
    async def test_detection_performance(self, detector, mock_tidb):
        """测试检测性能 - 应能在500ms内完成"""
        import time
        
        # 生成大量历史数据
        work_history = [
            {
                "detected_at": datetime.now() - timedelta(days=i*30),
                "grid_id": f"work_{i}",
                "lat": 22.5 + i*0.01,
                "lng": 113.9 + i*0.01
            }
            for i in range(10)
        ]
        
        mock_tidb.fetchall.side_effect = [
            work_history,
            [],
            [],
            []
        ]
        mock_tidb.fetchone.return_value = None
        
        start = time.time()
        events = await detector.detect("user_perf", 90)
        duration = time.time() - start
        
        assert duration < 0.5  # 500ms内完成
