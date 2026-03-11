"""
M05 Business Intelligence Service
商业智能服务 - 数据聚合、报表生成、洞察分析
"""

import os
from contextlib import asynccontextmanager
from datetime import datetime
from typing import List, Optional

import uvicorn
from fastapi import FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import FileResponse, JSONResponse

from src.analytics.aggregator import DataAggregator
from src.analytics.segmentation import UserSegmentation
from src.reports.generator import ReportGenerator


@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期"""
    # 启动时初始化
    app.state.aggregator = DataAggregator()
    app.state.segmentation = UserSegmentation()
    app.state.report_gen = ReportGenerator()
    yield
    # 关闭时清理


app = FastAPI(
    title="Squirrel M05 Business Intelligence",
    description="商业智能服务 - 客流分析、用户分群、报表生成",
    version="1.0.0",
    lifespan=lifespan
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
async def health():
    return {"status": "healthy", "service": "m05-biz-intel", "time": datetime.now()}


@app.get("/api/v1/analytics/footfall")
async def get_footfall_analysis(
    area: str = Query(..., description="区域grid_id或站点ID"),
    start_date: str = Query(..., description="开始日期 YYYY-MM-DD"),
    end_date: str = Query(..., description="结束日期 YYYY-MM-DD"),
    granularity: str = Query("daily", enum=["hourly", "daily", "weekly"])
):
    """客流分析 API"""
    try:
        result = app.state.aggregator.analyze_footfall(
            area=area,
            start_date=start_date,
            end_date=end_date,
            granularity=granularity
        )
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/v1/analytics/commute-patterns")
async def get_commute_patterns(
    line_id: Optional[str] = Query(None, description="地铁线路ID"),
    station_id: Optional[str] = Query(None, description="站点ID"),
    hour: Optional[int] = Query(None, ge=0, le=23, description="时段")
):
    """通勤模式分析"""
    try:
        result = app.state.aggregator.analyze_commute_patterns(
            line_id=line_id,
            station_id=station_id,
            hour=hour
        )
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/v1/segments")
async def list_segments():
    """获取用户分群列表"""
    segments = app.state.segmentation.get_all_segments()
    return {"segments": segments}


@app.get("/api/v1/segments/{segment_id}/users")
async def get_segment_users(
    segment_id: str,
    limit: int = Query(100, le=1000),
    offset: int = Query(0, ge=0)
):
    """获取分群用户列表"""
    users = app.state.segmentation.get_segment_users(segment_id, limit, offset)
    return {"segment_id": segment_id, "users": users, "total": len(users)}


@app.post("/api/v1/segments/create")
async def create_segment(criteria: dict):
    """创建新用户分群"""
    try:
        segment = app.state.segmentation.create_segment(criteria)
        return segment
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/v1/reports/types")
async def list_report_types():
    """获取可生成的报表类型"""
    return {
        "reports": [
            {"id": "daily_operation", "name": "日常运营报表", "description": "每日客流、收入、异常统计"},
            {"id": "weekly_insight", "name": "周度洞察报告", "description": "周趋势分析、用户行为变化"},
            {"id": "monthly_business", "name": "月度商业报告", "description": "商业指标、用户画像、收入分析"},
            {"id": "custom", "name": "自定义报表", "description": "按需定制的专项分析"}
        ]
    }


@app.post("/api/v1/reports/generate")
async def generate_report(request: dict):
    """生成报表"""
    try:
        report_type = request.get("type", "daily_operation")
        params = request.get("params", {})
        
        filepath = await app.state.report_gen.generate(
            report_type=report_type,
            params=params
        )
        
        return {
            "status": "success",
            "download_url": f"/api/v1/reports/download?file={filepath}",
            "expires_at": "2024-12-31T23:59:59"
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/v1/reports/download")
async def download_report(file: str):
    """下载报表文件"""
    return FileResponse(file, media_type="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")


@app.get("/api/v1/insights/city-overview")
async def get_city_overview():
    """城市级数据概览（B2G/B2B大屏数据）"""
    try:
        overview = app.state.aggregator.get_city_overview()
        return overview
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    port = int(os.getenv("M05_PORT", "8085"))
    uvicorn.run(app, host="0.0.0.0", port=port)
