"""
Pytest fixtures and configuration
"""

import pytest
from unittest.mock import AsyncMock


@pytest.fixture
async def mock_tidb():
    """Mock TiDB client"""
    client = AsyncMock()
    client.is_connected.return_value = True
    return client


@pytest.fixture
async def mock_redis():
    """Mock Redis client"""
    client = AsyncMock()
    client.is_connected.return_value = True
    return client


@pytest.fixture
def sample_user_data():
    """Sample user data for testing"""
    return {
        "user_id": "test_user_001",
        "home_grid": "grid_2254_11395",
        "work_grid": "grid_2253_11406",
        "locations": [
            {"lat": 22.5431, "lng": 113.9589, "timestamp": 1704067200},
            {"lat": 22.5268, "lng": 113.9800, "timestamp": 1704070800},
        ]
    }
