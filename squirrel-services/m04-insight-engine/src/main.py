#!/usr/bin/env python3
"""
Squirrel M04 Insight Engine - 用户洞察服务
提供用户画像分析、生活事件检测、通勤模式识别等功能
"""

import asyncio
import logging
import os
from contextlib import asynccontextmanager
from datetime import datetime, timedelta
from typing import AsyncGenerator, Dict, List, Optional

import grpc
import uvicorn
from fastapi import FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware
from prometheus_client import Counter, Histogram, make_asgi_app

from src.api import router
from src.config import Settings, get_settings
from src.db.tidb_client import TiDBClient
from src.db.redis_client import RedisClient
from src.models import LifeEvent, UserProfile
from src.profiler import UserProfiler
from src.event_detector import LifeEventDetector
from src.services.m03_client import M03Client

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Prometheus指标
REQUEST_COUNT = Counter('m04_requests_total', 'Total requests', ['method', 'endpoint'])
PROCESSING_TIME = Histogram('m04_processing_seconds', 'Processing time')
ACTIVE_USERS = Counter('m04_active_users_total', 'Active users analyzed')


class InsightEngine:
    """洞察引擎主类"""
    
    def __init__(self):
        self.settings: Optional[Settings] = None
        self.tidb: Optional[TiDBClient] = None
        self.redis: Optional[RedisClient] = None
        self.m03_client: Optional[M03Client] = None
        self.profiler: Optional[UserProfiler] = None
        self.event_detector: Optional[LifeEventDetector] = None
        self._shutdown_event: Optional[asyncio.Event] = None
    
    async def initialize(self) -> None:
        """初始化所有组件"""
        logger.info("Initializing M04 Insight Engine...")
        
        self.settings = get_settings()
        self._shutdown_event = asyncio.Event()
        
        # 初始化数据库连接
        self.tidb = TiDBClient(self.settings.TIDB_DSN)
        await self.tidb.connect()
        logger.info("Connected to TiDB")
        
        # 初始化Redis缓存
        self.redis = RedisClient(
            host=self.settings.REDIS_HOST,
            port=self.settings.REDIS_PORT,
            db=self.settings.REDIS_DB
        )
        await self.redis.connect()
        logger.info("Connected to Redis")
        
        # 初始化M03客户端
        self.m03_client = M03Client(self.settings.M03_ENDPOINT)
        await self.m03_client.connect()
        logger.info("Connected to M03 Trajectory Service")
        
        # 初始化分析器
        self.profiler = UserProfiler(self.tidb, self.redis)
        self.event_detector = LifeEventDetector(self.tidb, self.redis)
        
        logger.info("M04 Insight Engine initialized successfully")
    
    async def shutdown(self) -> None:
        """优雅关闭"""
        logger.info("Shutting down M04 Insight Engine...")
        
        if self.m03_client:
            await self.m03_client.close()
        if self.tidb:
            await self.tidb.close()
        if self.redis:
            await self.redis.close()
        
        logger.info("M04 Insight Engine shutdown complete")
    
    async def analyze_user(self, user_id: str, days: int = 30) -> UserProfile:
        """分析单个用户"""
        REQUEST_COUNT.labels(method='analyze', endpoint='user').inc()
        
        with PROCESSING_TIME.time():
            profile = await self.profiler.analyze(user_id, days)
            ACTIVE_USERS.inc()
            
            # 缓存结果
            cache_key = f"profile:{user_id}"
            await self.redis.setex(cache_key, 3600, profile.model_dump_json())
            
            return profile
    
    async def detect_life_events(self, user_id: str, days: int = 30) -> List[LifeEvent]:
        """检测用户生活事件"""
        REQUEST_COUNT.labels(method='detect', endpoint='events').inc()
        
        with PROCESSING_TIME.time():
            events = await self.event_detector.detect(user_id, days)
            return events
    
    async def get_commute_pattern(self, user_id: str, days: int = 30) -> Dict:
        """获取通勤模式"""
        REQUEST_COUNT.labels(method='get', endpoint='commute_pattern').inc()
        
        with PROCESSING_TIME.time():
            pattern = await self.profiler.get_commute_pattern(user_id, days)
            return pattern
    
    async def batch_analyze(self, user_ids: List[str], days: int = 30) -> Dict[str, UserProfile]:
        """批量分析用户"""
        REQUEST_COUNT.labels(method='analyze', endpoint='batch').inc()
        
        results = {}
        # 并发分析
        tasks = [self.analyze_user(uid, days) for uid in user_ids]
        profiles = await asyncio.gather(*tasks, return_exceptions=True)
        
        for uid, profile in zip(user_ids, profiles):
            if isinstance(profile, Exception):
                logger.error(f"Failed to analyze user {uid}: {profile}")
                continue
            results[uid] = profile
        
        return results


# 全局引擎实例
engine = InsightEngine()


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator:
    """应用生命周期管理"""
    await engine.initialize()
    yield
    await engine.shutdown()


# 创建FastAPI应用
app = FastAPI(
    title="Squirrel M04 Insight Engine",
    description="用户洞察分析服务 - 通勤模式识别、生活事件检测、用户画像",
    version="1.2.0",
    lifespan=lifespan
)

# CORS中间件
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 注册路由
app.include_router(router, prefix="/api/v1")

# Prometheus指标端点
metrics_app = make_asgi_app()
app.mount("/metrics", metrics_app)


@app.get("/health")
async def health_check():
    """健康检查端点"""
    return {
        "status": "healthy",
        "service": "m04-insight-engine",
        "version": "1.2.0",
        "timestamp": datetime.now().isoformat()
    }


@app.get("/ready")
async def readiness_check():
    """就绪检查端点"""
    checks = {
        "tidb": engine.tidb.is_connected() if engine.tidb else False,
        "redis": engine.redis.is_connected() if engine.redis else False,
        "m03": engine.m03_client.is_connected() if engine.m03_client else False,
    }
    
    all_ready = all(checks.values())
    
    if all_ready:
        return {"status": "ready", "checks": checks}
    else:
        raise HTTPException(status_code=503, detail={"status": "not ready", "checks": checks})


@app.get("/api/v1/users/{user_id}/profile")
async def get_user_profile(user_id: str, days: int = Query(default=30, ge=7, le=90)):
    """获取用户画像"""
    try:
        profile = await engine.analyze_user(user_id, days)
        return profile.model_dump()
    except Exception as e:
        logger.error(f"Error analyzing user {user_id}: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/v1/users/{user_id}/events")
async def get_user_events(user_id: str, days: int = Query(default=30, ge=7, le=90)):
    """获取用户生活事件"""
    try:
        events = await engine.detect_life_events(user_id, days)
        return {"user_id": user_id, "events": [e.model_dump() for e in events]}
    except Exception as e:
        logger.error(f"Error detecting events for user {user_id}: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/v1/batch/analyze")
async def batch_analyze(user_ids: List[str], days: int = Query(default=30, ge=7, le=90)):
    """批量分析用户"""
    if len(user_ids) > 100:
        raise HTTPException(status_code=400, detail="Maximum 100 users per batch")
    
    try:
        results = await engine.batch_analyze(user_ids, days)
        return {
            "total": len(user_ids),
            "success": len(results),
            "results": {uid: profile.model_dump() for uid, profile in results.items()}
        }
    except Exception as e:
        logger.error(f"Error in batch analysis: {e}")
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    port = int(os.getenv("M04_PORT", "8084"))
    uvicorn.run(app, host="0.0.0.0", port=port)
