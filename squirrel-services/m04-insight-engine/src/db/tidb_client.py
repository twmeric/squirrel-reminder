"""
TiDB数据库客户端
"""

import logging
from contextlib import asynccontextmanager
from typing import Any, Dict, List, Optional

import aiomysql

logger = logging.getLogger(__name__)


class TiDBClient:
    """异步TiDB客户端"""
    
    def __init__(self, dsn: str):
        self.dsn = dsn
        self.pool: Optional[aiomysql.Pool] = None
        self._connected = False
    
    async def connect(self) -> None:
        """建立连接池"""
        try:
            self.pool = await aiomysql.create_pool(
                host=self._parse_host(),
                port=self._parse_port(),
                user=self._parse_user(),
                password=self._parse_password(),
                db=self._parse_db(),
                charset='utf8mb4',
                autocommit=True,
                minsize=5,
                maxsize=20
            )
            self._connected = True
            logger.info("TiDB connection pool created")
        except Exception as e:
            logger.error(f"Failed to connect to TiDB: {e}")
            raise
    
    async def close(self) -> None:
        """关闭连接池"""
        if self.pool:
            self.pool.close()
            await self.pool.wait_closed()
            self._connected = False
            logger.info("TiDB connection pool closed")
    
    def is_connected(self) -> bool:
        return self._connected
    
    @asynccontextmanager
    async def acquire(self):
        """获取连接上下文"""
        async with self.pool.acquire() as conn:
            async with conn.cursor(aiomysql.DictCursor) as cur:
                yield cur
    
    async def fetchone(self, query: str, params: tuple = ()) -> Optional[Dict]:
        """查询单条记录"""
        async with self.acquire() as cur:
            await cur.execute(query, params)
            return await cur.fetchone()
    
    async def fetchall(self, query: str, params: tuple = ()) -> List[Dict]:
        """查询多条记录"""
        async with self.acquire() as cur:
            await cur.execute(query, params)
            return await cur.fetchall()
    
    async def execute(self, query: str, params: tuple = ()) -> int:
        """执行SQL"""
        async with self.acquire() as cur:
            return await cur.execute(query, params)
    
    async def executemany(self, query: str, params: List[tuple]) -> int:
        """批量执行"""
        async with self.acquire() as cur:
            return await cur.executemany(query, params)
    
    # DSN解析辅助方法
    def _parse_host(self) -> str:
        # 简化解析，实际应该使用更健壮的解析
        import re
        match = re.search(r'@tcp\(([^:]+)', self.dsn)
        return match.group(1) if match else 'localhost'
    
    def _parse_port(self) -> int:
        import re
        match = re.search(r'@tcp\([^:]+:(\d+)', self.dsn)
        return int(match.group(1)) if match else 4000
    
    def _parse_user(self) -> str:
        import re
        match = re.search(r'^([^:]+):', self.dsn)
        return match.group(1) if match else 'root'
    
    def _parse_password(self) -> str:
        import re
        match = re.search(r'^[^:]+:([^@]+)@', self.dsn)
        return match.group(1) if match else ''
    
    def _parse_db(self) -> str:
        import re
        match = re.search(r'/([^?]+)', self.dsn)
        return match.group(1) if match else 'squirrel'
