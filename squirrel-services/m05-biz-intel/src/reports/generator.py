"""
报表生成器
"""

import os
from datetime import datetime
from typing import Dict

import pandas as pd
from jinja2 import Template


class ReportGenerator:
    """报表生成器"""
    
    def __init__(self):
        self.output_dir = "reports"
        os.makedirs(self.output_dir, exist_ok=True)
    
    async def generate(self, report_type: str, params: Dict) -> str:
        """生成报表"""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        
        if report_type == "daily_operation":
            return await self._generate_daily_report(params, timestamp)
        elif report_type == "weekly_insight":
            return await self._generate_weekly_report(params, timestamp)
        elif report_type == "monthly_business":
            return await self._generate_monthly_report(params, timestamp)
        else:
            return await self._generate_custom_report(params, timestamp)
    
    async def _generate_daily_report(self, params: Dict, timestamp: str) -> str:
        """生成日常运营报表"""
        date = params.get("date", datetime.now().strftime("%Y-%m-%d"))
        
        # 模拟数据
        data = {
            "日期": [date],
            "总客流": [3200000],
            "早高峰客流": [850000],
            "晚高峰客流": [920000],
            "平均行程时间": [32],
            "活跃用户数": [850000],
            "新用户数": [3200],
            "收入估算": ["¥128,000"]
        }
        
        df = pd.DataFrame(data)
        filepath = f"{self.output_dir}/daily_report_{timestamp}.xlsx"
        df.to_excel(filepath, index=False)
        
        return filepath
    
    async def _generate_weekly_report(self, params: Dict, timestamp: str) -> str:
        """生成周度洞察报告"""
        # 模拟一周数据
        data = {
            "日期": pd.date_range(end=datetime.now(), periods=7, freq='D').strftime("%Y-%m-%d").tolist(),
            "日客流": [3100000, 3200000, 3150000, 3300000, 3400000, 2800000, 2600000],
            "通勤族": [2100000, 2200000, 2150000, 2250000, 2300000, 1200000, 1100000],
            "周末达人": [150000, 160000, 155000, 170000, 180000, 850000, 920000]
        }
        
        df = pd.DataFrame(data)
        filepath = f"{self.output_dir}/weekly_report_{timestamp}.xlsx"
        
        with pd.ExcelWriter(filepath, engine='openpyxl') as writer:
            df.to_excel(writer, sheet_name='客流趋势', index=False)
            
            # 添加汇总页
            summary = pd.DataFrame({
                "指标": ["周总客流", "日均客流", "最高单日", "最低单日", "周末占比"],
                "数值": ["2155万", "308万", "340万", "260万", "26.8%"]
            })
            summary.to_excel(writer, sheet_name='汇总', index=False)
        
        return filepath
    
    async def _generate_monthly_report(self, params: Dict, timestamp: str) -> str:
        """生成月度商业报告"""
        filepath = f"{self.output_dir}/monthly_report_{timestamp}.xlsx"
        
        # 创建多个sheet
        with pd.ExcelWriter(filepath, engine='openpyxl') as writer:
            # 客流分析
            traffic = pd.DataFrame({
                "线路": ["1号线", "2号线", "3号线", "4号线", "5号线"],
                "月客流(万)": [850, 720, 680, 520, 480],
                "环比增长": ["+3.2%", "+1.5%", "+2.8%", "-0.5%", "+4.1%"]
            })
            traffic.to_excel(writer, sheet_name='线路客流', index=False)
            
            # 用户画像
            users = pd.DataFrame({
                "群体": ["通勤族", "外勤人员", "学生", "其他"],
                "人数(万)": [65, 15, 12, 8],
                "占比": ["65%", "15%", "12%", "8%"],
                "活跃度": ["高", "中", "高", "低"]
            })
            users.to_excel(writer, sheet_name='用户画像', index=False)
        
        return filepath
    
    async def _generate_custom_report(self, params: Dict, timestamp: str) -> str:
        """生成自定义报表"""
        filepath = f"{self.output_dir}/custom_report_{timestamp}.xlsx"
        
        df = pd.DataFrame(params.get("data", []))
        df.to_excel(filepath, index=False)
        
        return filepath
    
    def generate_html_summary(self, report_data: Dict) -> str:
        """生成HTML摘要"""
        template = Template("""
        <!DOCTYPE html>
        <html>
        <head><title>报表摘要</title></head>
        <body>
            <h1>{{ title }}</h1>
            <p>生成时间: {{ timestamp }}</p>
            <hr>
            <h2>关键指标</h2>
            <ul>
            {% for key, value in metrics.items() %}
                <li><strong>{{ key }}:</strong> {{ value }}</li>
            {% endfor %}
            </ul>
        </body>
        </html>
        """)
        
        return template.render(
            title=report_data.get("title", "数据报表"),
            timestamp=datetime.now().isoformat(),
            metrics=report_data.get("metrics", {})
        )
