#!/usr/bin/env python3
"""
生成模拟用户数据用于测试
生成通勤者、外勤人员等不同类型用户的轨迹数据
"""

import json
import random
import argparse
from datetime import datetime, timedelta
from typing import List, Dict
import numpy as np


class UserDataGenerator:
    """用户数据生成器"""
    
    # 深圳地铁站点坐标
    STATIONS = {
        "science_park": (22.5431, 113.9589),
        "shenzhen_university": (22.5268, 113.9800),
        "world_window": (22.5342, 113.9789),
        "futian": (22.5432, 114.0578),
        "baoan_center": (22.5202, 113.9266),
        "high_tech": (22.5279, 113.9889),
    }
    
    def __init__(self, seed: int = 42):
        random.seed(seed)
        np.random.seed(seed)
    
    def generate_commuter(self, user_id: str, days: int = 30) -> Dict:
        """生成上班族数据"""
        home = self.STATIONS["science_park"]
        work = self.STATIONS["futian"]
        
        locations = []
        
        for day in range(days):
            date = datetime.now() - timedelta(days=days-day)
            
            # 跳过周末
            if date.weekday() >= 5:
                # 周末在家附近活动
                locations.extend(self._generate_home_day(date, home, noise=0.5))
                continue
            
            # 早上从家出发 (7:00-8:00)
            morning_start = date.replace(hour=7, minute=random.randint(0, 59))
            locations.extend(self._generate_commute(
                morning_start, home, work, 
                duration_minutes=random.randint(30, 45)
            ))
            
            # 工作时间 (9:00-18:00)
            work_start = date.replace(hour=9)
            work_end = date.replace(hour=18)
            locations.extend(self._generate_stay(
                work_start, work_end, work, noise=0.1
            ))
            
            # 晚上回家 (18:00-19:00)
            evening_start = date.replace(hour=18, minute=random.randint(0, 30))
            locations.extend(self._generate_commute(
                evening_start, work, home,
                duration_minutes=random.randint(30, 45)
            ))
            
            # 晚上在家 (19:00-24:00)
            night_start = date.replace(hour=19)
            night_end = date.replace(hour=23, minute=59)
            locations.extend(self._generate_stay(
                night_start, night_end, home, noise=0.05
            ))
        
        return {
            "user_id": user_id,
            "type": "commuter",
            "home": home,
            "work": work,
            "locations": locations
        }
    
    def generate_field_worker(self, user_id: str, days: int = 30) -> Dict:
        """生成外勤人员数据 - 移动性高，无固定工作地点"""
        home = self.STATIONS["baoan_center"]
        
        locations = []
        
        for day in range(days):
            date = datetime.now() - timedelta(days=days-day)
            
            # 早上在家
            morning = date.replace(hour=8)
            locations.extend(self._generate_stay(
                morning, morning + timedelta(hours=1), home, noise=0.1
            ))
            
            # 白天随机移动 (访问多个地点)
            current_time = date.replace(hour=9)
            end_time = date.replace(hour=18)
            
            while current_time < end_time:
                # 随机选择一个站点
                station = random.choice(list(self.STATIONS.values()))
                
                # 移动到该地点
                travel_duration = random.randint(20, 40)
                locations.extend(self._generate_commute(
                    current_time, 
                    (locations[-1]["lat"], locations[-1]["lng"]) if locations else home,
                    station,
                    travel_duration
                ))
                
                current_time += timedelta(minutes=travel_duration)
                
                # 停留一段时间
                stay_duration = random.randint(30, 120)
                locations.extend(self._generate_stay(
                    current_time,
                    current_time + timedelta(minutes=stay_duration),
                    station,
                    noise=0.2
                ))
                
                current_time += timedelta(minutes=stay_duration)
        
        return {
            "user_id": user_id,
            "type": "field_worker",
            "home": home,
            "locations": locations
        }
    
    def generate_unemployed(self, user_id: str, days: int = 30) -> Dict:
        """生成失业人员数据 - 无通勤，长时间在家"""
        home = self.STATIONS["high_tech"]
        
        locations = []
        
        # 前15天正常通勤
        for day in range(15):
            date = datetime.now() - timedelta(days=days-day)
            if date.weekday() < 5:
                # 简化的通勤模式
                locations.extend(self._generate_commute_day(date, home))
            else:
                locations.extend(self._generate_home_day(date, home))
        
        # 后15天无通勤，长时间在家
        for day in range(15, days):
            date = datetime.now() - timedelta(days=days-day)
            locations.extend(self._generate_home_day(date, home, noise=0.1))
        
        return {
            "user_id": user_id,
            "type": "unemployed",
            "home": home,
            "unemployment_start_day": 15,
            "locations": locations
        }
    
    def _generate_commute(self, start_time: datetime, 
                         start: tuple, end: tuple,
                         duration_minutes: int) -> List[Dict]:
        """生成通勤轨迹"""
        points = []
        steps = max(3, duration_minutes // 2)  # 每2分钟一个点
        
        for i in range(steps):
            progress = i / (steps - 1)
            lat = start[0] + (end[0] - start[0]) * progress
            lng = start[1] + (end[1] - start[1]) * progress
            
            # 添加噪声
            lat += random.gauss(0, 0.0001)
            lng += random.gauss(0, 0.0001)
            
            points.append({
                "timestamp": (start_time + timedelta(minutes=i*2)).isoformat(),
                "lat": round(lat, 6),
                "lng": round(lng, 6),
                "speed": random.uniform(25, 45),  # 地铁速度
                "accuracy": random.uniform(5, 20)
            })
        
        return points
    
    def _generate_stay(self, start_time: datetime, end_time: datetime,
                      center: tuple, noise: float = 0.1) -> List[Dict]:
        """生成停留轨迹"""
        points = []
        duration = (end_time - start_time).total_seconds() / 60
        steps = max(2, int(duration / 5))  # 每5分钟一个点
        
        for i in range(steps):
            time_point = start_time + timedelta(minutes=i*5)
            points.append({
                "timestamp": time_point.isoformat(),
                "lat": round(center[0] + random.gauss(0, noise * 0.001), 6),
                "lng": round(center[1] + random.gauss(0, noise * 0.001), 6),
                "speed": 0,
                "accuracy": random.uniform(3, 15)
            })
        
        return points
    
    def _generate_home_day(self, date: datetime, home: tuple, noise: float = 0.05) -> List[Dict]:
        """生成在家的一天"""
        start = date.replace(hour=0)
        end = date.replace(hour=23, minute=59)
        return self._generate_stay(start, end, home, noise)
    
    def _generate_commute_day(self, date: datetime, home: tuple) -> List[Dict]:
        """生成通勤的一天 (简化)"""
        # 简化的实现
        work = (22.5432, 114.0578)
        
        morning = date.replace(hour=8)
        return self._generate_commute(morning, home, work, 40)
    
    def generate_batch(self, count: int, user_type: str = "mixed") -> List[Dict]:
        """生成批量用户数据"""
        users = []
        
        for i in range(count):
            user_id = f"sim_user_{i:04d}"
            
            if user_type == "mixed":
                # 随机选择类型
                weights = [0.6, 0.3, 0.1]  # 60%通勤, 30%外勤, 10%失业
                gen_func = random.choices([
                    self.generate_commuter,
                    self.generate_field_worker,
                    self.generate_unemployed
                ], weights=weights)[0]
                users.append(gen_func(user_id))
            elif user_type == "commuter":
                users.append(self.generate_commuter(user_id))
            elif user_type == "field_worker":
                users.append(self.generate_field_worker(user_id))
            elif user_type == "unemployed":
                users.append(self.generate_unemployed(user_id))
        
        return users


def main():
    parser = argparse.ArgumentParser(description="Generate test user data")
    parser.add_argument("--count", type=int, default=10, help="Number of users")
    parser.add_argument("--type", default="mixed", 
                       choices=["mixed", "commuter", "field_worker", "unemployed"],
                       help="User type")
    parser.add_argument("--days", type=int, default=30, help="Days of data")
    parser.add_argument("--output", default="test_data.json", help="Output file")
    parser.add_argument("--seed", type=int, default=42, help="Random seed")
    
    args = parser.parse_args()
    
    generator = UserDataGenerator(seed=args.seed)
    users = generator.generate_batch(args.count, args.type)
    
    with open(args.output, "w") as f:
        json.dump(users, f, indent=2)
    
    print(f"Generated {args.count} users ({args.type}) with {args.days} days of data")
    print(f"Saved to {args.output}")
    
    # 打印统计
    total_locations = sum(len(u["locations"]) for u in users)
    print(f"Total location points: {total_locations}")
    print(f"Average per user: {total_locations // args.count}")


if __name__ == "__main__":
    main()
