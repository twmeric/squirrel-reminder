"""
Redis缓存客户端
"""

import json
import logging
from typing import Any, Optional, Union

import redis.asyncio as redis

logger = logging.getLogger(__name__)


class RedisClient:
    """异步Redis客户端"""
    
    def __init__(self, host: str = "localhost", port: int = 6379, db: int = 0):
        self.host = host
        self.port = port
        self.db = db
        self.client: Optional[redis.Redis] = None
        self._connected = False
    
    async def connect(self) -> None:
        """连接Redis"""
        try:
            self.client = redis.Redis(
                host=self.host,
                port=self.port,
                db=self.db,
                decode_responses=True
            )
            await self.client.ping()
            self._connected = True
            logger.info(f"Connected to Redis at {self.host}:{self.port}")
        except Exception as e:
            logger.error(f"Failed to connect to Redis: {e}")
            raise
    
    async def close(self) -> None:
        """关闭连接"""
        if self.client:
            await self.client.close()
            self._connected = False
            logger.info("Redis connection closed")
    
    def is_connected(self) -> bool:
        return self._connected
    
    async def get(self, key: str) -> Optional[str]:
        """获取值"""
        return await self.client.get(key)
    
    async def set(self, key: str, value: Union[str, int, float], ex: Optional[int] = None) -> bool:
        """设置值"""
        return await self.client.set(key, value, ex=ex)
    
    async def setex(self, key: str, seconds: int, value: str) -> bool:
        """设置值并指定过期时间"""
        return await self.client.setex(key, seconds, value)
    
    async def delete(self, key: str) -> int:
        """删除键"""
        return await self.client.delete(key)
    
    async def exists(self, key: str) -> bool:
        """检查键是否存在"""
        return await self.client.exists(key) > 0
    
    async def get_json(self, key: str) -> Optional[Any]:
        """获取JSON对象"""
        value = await self.get(key)
        if value:
            return json.loads(value)
        return None
    
    async def set_json(self, key: str, value: Any, ex: Optional[int] = None) -> bool:
        """存储JSON对象"""
        return await self.set(key, json.dumps(value), ex)
    
    async def increment(self, key: str, amount: int = 1) -> int:
        """自增"""
        return await self.client.incrby(key, amount)
    
    async def expire(self, key: str, seconds: int) -> bool:
        """设置过期时间"""
        return await self.client.expire(key, seconds)
    
    async def ttl(self, key: str) -> int:
        """获取剩余过期时间"""
        return await self.client.ttl(key)
