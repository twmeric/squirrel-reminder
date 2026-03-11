"""
数据模型定义
"""

from datetime import datetime
from enum import Enum
from typing import Dict, List, Optional, Any

from pydantic import BaseModel, Field


class OccupationType(str, Enum):
    """职业类型"""
    COMMUTER = "commuter"        # 上班族
    FIELD_WORKER = "field_worker" # 外勤人员
    STUDENT = "student"          # 学生
    FREELANCER = "freelancer"    # 自由职业
    RETIRED = "retired"          # 退休
    UNEMPLOYED = "unemployed"    # 失业
    OTHER = "other"              # 其他
    UNKNOWN = "unknown"          # 未知


class EventType(str, Enum):
    """事件类型"""
    JOB_CHANGE = "job_change"          # 换工作
    MOVE = "move"                      # 搬家
    UNEMPLOYMENT = "unemployment"      # 失业
    RETIREMENT = "retirement"          # 退休
    NEW_JOB = "new_job"                # 新工作
    COMMUTE_CHANGE = "commute_change"  # 通勤变化
    SCHEDULE_CHANGE = "schedule_change" # 作息变化


class EventConfidence(str, Enum):
    """事件置信度"""
    LOW = "low"        # 低 (0.3-0.5)
    MEDIUM = "medium"  # 中 (0.5-0.8)
    HIGH = "high"      # 高 (0.8-1.0)
    
    @property
    def value(self) -> float:
        values = {"low": 0.4, "medium": 0.65, "high": 0.9}
        return values.get(self.value, 0.5)


class LocationProfile(BaseModel):
    """位置画像"""
    grid_id: str = Field(..., description="网格ID")
    lat: float = Field(..., description="纬度")
    lng: float = Field(..., description="经度")
    confidence: float = Field(..., ge=0, le=1, description="置信度")
    detection_method: str = Field(..., description="检测方法")
    first_seen: Optional[datetime] = None
    last_seen: Optional[datetime] = None
    visit_count: Optional[int] = None


class CommutePattern(BaseModel):
    """通勤模式"""
    is_commuter: bool = Field(default=False, description="是否为通勤者")
    avg_commute_minutes: Optional[float] = None
    regularity_score: Optional[float] = Field(None, ge=0, le=1)
    route_consistency: Optional[float] = Field(None, ge=0, le=1)
    avg_departure_hour: Optional[float] = None
    commute_count: int = 0
    typical_route: Optional[str] = None
    alternative_routes: List[str] = Field(default_factory=list)


class UserProfile(BaseModel):
    """用户画像"""
    user_id: str = Field(..., description="用户ID")
    analysis_date: datetime = Field(..., description="分析日期")
    analysis_days: int = Field(..., description="分析天数")
    
    home: Optional[LocationProfile] = Field(None, description="家位置")
    work: Optional[LocationProfile] = Field(None, description="工作位置")
    
    regularity_score: float = Field(..., ge=0, le=1, description="规律性分数")
    mobility_score: float = Field(..., ge=0, le=1, description="活跃度分数")
    occupation: OccupationType = Field(..., description="职业类型")
    
    commute_pattern: Optional[CommutePattern] = None
    
    total_staypoints: int = 0
    unique_locations: int = 0
    
    # 标签
    tags: List[str] = Field(default_factory=list)
    
    class Config:
        json_encoders = {
            datetime: lambda v: v.isoformat()
        }


class LifeEvent(BaseModel):
    """生活事件"""
    user_id: str
    event_type: EventType
    confidence: EventConfidence
    detected_at: datetime
    event_date: Optional[datetime] = None
    details: Dict[str, Any] = Field(default_factory=dict)
    evidence: List[str] = Field(default_factory=list)
    is_notified: bool = False
    notification_level: str = "L1"  # L1=info, L2=alert, L3=urgent
    
    class Config:
        json_encoders = {
            datetime: lambda v: v.isoformat()
        }


class UserSegment(BaseModel):
    """用户分群"""
    segment_id: str
    name: str
    description: str
    criteria: Dict[str, Any]
    user_count: int = 0
    avg_regularity: float = 0.0
    avg_mobility: float = 0.0
    top_occupations: Dict[str, int] = Field(default_factory=dict)


class DailyStats(BaseModel):
    """每日统计"""
    user_id: str
    date: datetime
    total_points: int = 0
    staypoint_count: int = 0
    unique_grids: int = 0
    total_distance_km: float = 0.0
    max_speed: float = 0.0
    transit_duration_min: int = 0
    
    # 活动特征
    home_hours: float = 0.0
    work_hours: float = 0.0
    commute_count: int = 0
    is_weekday: bool = True


class BatchAnalysisRequest(BaseModel):
    """批量分析请求"""
    user_ids: List[str] = Field(..., max_length=100)
    days: int = Field(default=30, ge=7, le=90)


class BatchAnalysisResponse(BaseModel):
    """批量分析响应"""
    total: int
    success: int
    failed: int = 0
    results: Dict[str, UserProfile]
    errors: Dict[str, str] = Field(default_factory=dict)
