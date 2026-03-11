"""
测试分析模块
"""

import pytest
from datetime import datetime, timedelta

from src.analytics.aggregator import DataAggregator
from src.analytics.segmentation import UserSegmentation


class TestDataAggregator:
    """测试数据聚合器"""
    
    @pytest.fixture
    def aggregator(self):
        return DataAggregator()
    
    def test_analyze_footfall(self, aggregator):
        """测试客流分析"""
        result = aggregator.analyze_footfall(
            area="grid_123",
            start_date="2024-03-01",
            end_date="2024-03-07",
            granularity="daily"
        )
        
        assert result["area"] == "grid_123"
        assert "total_visits" in result
        assert "trend" in result
        assert len(result["trend"]) == 7
    
    def test_analyze_commute_patterns(self, aggregator):
        """测试通勤模式分析"""
        result = aggregator.analyze_commute_patterns(
            line_id="L1",
            hour=8
        )
        
        assert result["line_id"] == "L1"
        assert len(result["patterns"]) == 1
        assert result["patterns"][0]["hour"] == 8
    
    def test_get_city_overview(self, aggregator):
        """测试城市概览"""
        result = aggregator.get_city_overview()
        
        assert "metrics" in result
        assert "today_stats" in result
        assert "demographics" in result


class TestUserSegmentation:
    """测试用户分群"""
    
    @pytest.fixture
    def segmentation(self):
        return UserSegmentation()
    
    def test_get_all_segments(self, segmentation):
        """测试获取所有分群"""
        segments = segmentation.get_all_segments()
        
        assert len(segments) == 5
        assert segments[0]["id"] == "high_value_commuters"
    
    def test_get_segment_users(self, segmentation):
        """测试获取分群用户"""
        users = segmentation.get_segment_users(
            "high_value_commuters",
            limit=10
        )
        
        assert len(users) == 10
    
    def test_create_segment(self, segmentation):
        """测试创建分群"""
        criteria = {
            "name": "测试分群",
            "description": "用于测试",
            "filters": {"age": ">30"}
        }
        
        segment = segmentation.create_segment(criteria)
        
        assert segment["name"] == "测试分群"
        assert segment["status"] == "active"
