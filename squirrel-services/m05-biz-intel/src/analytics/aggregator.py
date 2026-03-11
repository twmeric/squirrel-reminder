"""
数据聚合分析器
"""

from datetime import datetime, timedelta
from typing import Dict, List, Optional

import pandas as pd


class DataAggregator:
    """数据聚合分析器"""
    
    def analyze_footfall(self, area: str, start_date: str, end_date: str, 
                        granularity: str = "daily") -> Dict:
        """客流分析"""
        # 模拟数据 - 实际应从TiDB查询
        dates = pd.date_range(start=start_date, end=end_date, freq='D')
        
        data = []
        for date in dates:
            base = 1000 + (date.weekday() * 200)  # 工作日更多
            data.append({
                "date": date.strftime("%Y-%m-%d"),
                "total": int(base + 100),
                "morning_peak": int(base * 0.3),
                "evening_peak": int(base * 0.35),
                "off_peak": int(base * 0.35),
                "yoy_change": 5.2  # 同比增长
            })
        
        return {
            "area": area,
            "period": f"{start_date} to {end_date}",
            "granularity": granularity,
            "total_visits": sum(d["total"] for d in data),
            "avg_daily": sum(d["total"] for d in data) / len(data),
            "peak_day": max(data, key=lambda x: x["total"]),
            "trend": data
        }
    
    def analyze_commute_patterns(self, line_id: Optional[str] = None,
                                station_id: Optional[str] = None,
                                hour: Optional[int] = None) -> Dict:
        """通勤模式分析"""
        patterns = []
        
        for h in range(24):
            if hour is not None and h != hour:
                continue
                
            # 早高峰 7-9, 晚高峰 17-19
            if 7 <= h <= 9:
                intensity = "high"
                factor = 1.5
            elif 17 <= h <= 19:
                intensity = "high"
                factor = 1.4
            elif 10 <= h <= 16:
                intensity = "medium"
                factor = 1.0
            else:
                intensity = "low"
                factor = 0.3
            
            patterns.append({
                "hour": h,
                "intensity": intensity,
                "avg_passengers": int(1000 * factor),
                "avg_commute_time": 35 if intensity == "high" else 28
            })
        
        return {
            "line_id": line_id,
            "station_id": station_id,
            "patterns": patterns,
            "peak_hours": [7, 8, 17, 18, 19],
            "off_peak_hours": [10, 11, 14, 15]
        }
    
    def get_city_overview(self) -> Dict:
        """城市级概览"""
        return {
            "timestamp": datetime.now().isoformat(),
            "metrics": {
                "total_users": 1250000,
                "active_today": 850000,
                "total_routes": 15,
                "total_stations": 285
            },
            "today_stats": {
                "total_trips": 3200000,
                "avg_trip_duration": 32,
                "peak_hour": "08:00",
                "hot_stations": ["科技园", "深大", "会展中心"]
            },
            "demographics": {
                "commuter_ratio": 0.65,
                "field_worker_ratio": 0.15,
                "student_ratio": 0.12,
                "other_ratio": 0.08
            }
        }
