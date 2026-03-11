"""
M03 Trajectory Service gRPC客户端
"""

import logging
from typing import Optional

import grpc

from src.proto import m03_pb2, m03_pb2_grpc

logger = logging.getLogger(__name__)


class M03Client:
    """M03服务客户端"""
    
    def __init__(self, endpoint: str = "localhost:50053"):
        self.endpoint = endpoint
        self.channel: Optional[grpc.aio.Channel] = None
        self.stub: Optional[m03_pb2_grpc.TrajectoryProcessorStub] = None
        self._connected = False
    
    async def connect(self) -> None:
        """建立连接"""
        try:
            self.channel = grpc.aio.insecure_channel(self.endpoint)
            self.stub = m03_pb2_grpc.TrajectoryProcessorStub(self.channel)
            
            # 测试连接
            await self.stub.GetHealth(m03_pb2.HealthRequest())
            self._connected = True
            logger.info(f"Connected to M03 at {self.endpoint}")
        except Exception as e:
            logger.error(f"Failed to connect to M03: {e}")
            raise
    
    async def close(self) -> None:
        """关闭连接"""
        if self.channel:
            await self.channel.close()
            self._connected = False
            logger.info("M03 connection closed")
    
    def is_connected(self) -> bool:
        return self._connected
    
    async def get_trajectory(self, user_id: str, start_time: int, end_time: int) -> m03_pb2.TrajectoryResponse:
        """获取用户轨迹"""
        from google.protobuf.timestamp_pb2 import Timestamp
        
        request = m03_pb2.TrajectoryRequest(
            user_id=user_id,
            start_time=Timestamp(seconds=start_time),
            end_time=Timestamp(seconds=end_time)
        )
        return await self.stub.GetTrajectory(request)
    
    async def get_speed(self, user_id: str) -> m03_pb2.SpeedResponse:
        """获取当前速度"""
        request = m03_pb2.SpeedRequest(user_id=user_id)
        return await self.stub.GetSpeed(request)
    
    async def get_nearest_station(self, lat: float, lng: float) -> m03_pb2.NearestStationResponse:
        """获取最近站点"""
        request = m03_pb2.NearestStationRequest(latitude=lat, longitude=lng)
        return await self.stub.GetNearestStation(request)
    
    async def process_batch(self, user_id: str, locations: list) -> m03_pb2.BatchResponse:
        """批量处理位置"""
        from google.protobuf.timestamp_pb2 import Timestamp
        from datetime import datetime
        
        proto_locations = []
        for loc in locations:
            proto_loc = m03_pb2.Location(
                latitude=loc["lat"],
                longitude=loc["lng"],
                timestamp=Timestamp(seconds=int(loc["timestamp"])),
                accuracy=loc.get("accuracy", 0),
                speed=loc.get("speed", 0),
                provider=loc.get("provider", "unknown")
            )
            proto_locations.append(proto_loc)
        
        request = m03_pb2.BatchRequest(
            user_id=user_id,
            locations=proto_locations
        )
        return await self.stub.ProcessBatch(request)
