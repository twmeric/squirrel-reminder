"""
配置管理
"""

import os
from functools import lru_cache
from typing import Optional

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """应用配置"""
    
    # 服务配置
    M04_PORT: int = 8084
    ENVIRONMENT: str = "development"
    LOG_LEVEL: str = "INFO"
    
    # 数据库配置
    TIDB_DSN: str = "root@tcp(localhost:4000)/squirrel?parseTime=true"
    
    # Redis配置
    REDIS_HOST: str = "localhost"
    REDIS_PORT: int = 6379
    REDIS_DB: int = 0
    REDIS_PASSWORD: Optional[str] = None
    
    # M03服务配置
    M03_ENDPOINT: str = "localhost:50053"
    
    # 算法参数
    HOME_NIGHT_START_HOUR: int = 22
    HOME_NIGHT_END_HOUR: int = 6
    WORK_START_HOUR: int = 9
    WORK_END_HOUR: int = 18
    DBSCAN_EPS_METERS: float = 200.0
    DBSCAN_MIN_SAMPLES: int = 3
    
    # 缓存配置
    PROFILE_CACHE_TTL: int = 3600  # 1小时
    EVENT_CACHE_TTL: int = 1800    # 30分钟
    
    # 批处理配置
    BATCH_SIZE: int = 100
    BATCH_TIMEOUT: int = 30
    
    # 监控配置
    METRICS_ENABLED: bool = True
    TRACING_ENABLED: bool = True
    
    class Config:
        env_file = ".env"
        case_sensitive = True


@lru_cache()
def get_settings() -> Settings:
    """获取配置实例 (单例)"""
    return Settings()


def reload_settings() -> Settings:
    """重新加载配置"""
    get_settings.cache_clear()
    return get_settings()
