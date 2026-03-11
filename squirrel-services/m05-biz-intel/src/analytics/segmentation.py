"""
用户分群分析
"""

from typing import Dict, List, Optional


class UserSegmentation:
    """用户分群服务"""
    
    # 预定义分群
    PREDEFINED_SEGMENTS = [
        {
            "id": "high_value_commuters",
            "name": "高价值通勤族",
            "description": "通勤规律、消费能力强、活跃度高",
            "criteria": {"regularity": ">0.8", "mobility": "<3", "days_active": ">20"},
            "estimated_size": 450000,
            "characteristics": ["早8点出行", "单程30-45分钟", "周末低频使用"]
        },
        {
            "id": "field_workers",
            "name": "外勤人员",
            "description": "移动性高、路线多变、全天候活跃",
            "criteria": {"mobility": ">4", "regularity": "<0.5"},
            "estimated_size": 180000,
            "characteristics": ["多地点停留", "非高峰出行", "路线不固定"]
        },
        {
            "id": "students",
            "name": "学生群体",
            "description": "年轻用户、价格敏感、特定时段活跃",
            "criteria": {"age_estimate": "<25", "commute_pattern": "student"},
            "estimated_size": 150000,
            "characteristics": ["早晚高峰", "周末休闲出行", "换乘多"]
        },
        {
            "id": "night_owls",
            "name": "夜猫子",
            "description": "夜间活跃、作息不规律",
            "criteria": {"night_activity_ratio": ">0.3"},
            "estimated_size": 80000,
            "characteristics": ["22点后出行", "周末夜生活", "非工作日补觉"]
        },
        {
            "id": "weekend_warriors",
            "name": "周末达人",
            "description": "工作日低频、周末高频",
            "criteria": {"weekend_ratio": ">0.6"},
            "estimated_size": 120000,
            "characteristics": ["周末购物", "休闲娱乐", "城市探索"]
        }
    ]
    
    def get_all_segments(self) -> List[Dict]:
        """获取所有分群"""
        return self.PREDEFINED_SEGMENTS
    
    def get_segment_users(self, segment_id: str, limit: int = 100, 
                         offset: int = 0) -> List[Dict]:
        """获取分群用户列表（模拟）"""
        # 实际应从TiDB查询
        segment = next((s for s in self.PREDEFINED_SEGMENTS if s["id"] == segment_id), None)
        if not segment:
            return []
        
        users = []
        for i in range(min(limit, 100)):
            users.append({
                "user_id": f"user_{segment_id}_{offset + i:05d}",
                "score": 0.8 + (i % 20) / 100,
                "first_seen": "2024-01-01",
                "last_active": "2024-03-08"
            })
        return users
    
    def create_segment(self, criteria: Dict) -> Dict:
        """创建新分群"""
        segment_id = f"custom_{criteria.get('name', 'segment').lower().replace(' ', '_')}"
        
        new_segment = {
            "id": segment_id,
            "name": criteria.get("name", "自定义分群"),
            "description": criteria.get("description", ""),
            "criteria": criteria.get("filters", {}),
            "created_at": "2024-03-08T00:00:00",
            "status": "active"
        }
        
        return new_segment
    
    def analyze_segment_overlap(self, segment_ids: List[str]) -> Dict:
        """分析分群重叠"""
        # 实际应计算交集
        return {
            "segments": segment_ids,
            "overlap_matrix": [[1.0, 0.15, 0.05],
                              [0.15, 1.0, 0.10],
                              [0.05, 0.10, 1.0]],
            "total_unique_users": 750000
        }
