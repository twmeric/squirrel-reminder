# Squirrel Services API Documentation

## Overview

This document describes the APIs exposed by Squirrel Services (m03-trajectory and m04-insight-engine).

## m03-trajectory

### gRPC Services

#### TrajectoryProcessor

**ProcessBatch**
- Request: `BatchRequest`
  - user_id: string
  - locations: []Location
  - device_type: string
- Response: `BatchResponse`
  - processed_count: int32
  - staypoint_count: int32
  - process_time_ms: int64

#### LocationService

**GetSpeed**
- Request: `SpeedRequest` (user_id, include_history)
- Response: `SpeedResponse` (speed_kmh, is_moving, is_transit, confidence)

**GetNearestStation**
- Request: `NearestStationRequest` (lat, lng, max_distance)
- Response: `NearestStationResponse` (station_id, name, line_name, distance)

### HTTP Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| GET | /ready | Readiness check |
| GET | /metrics | Prometheus metrics |

## m04-insight-engine

### REST API

#### Get User Profile

```
GET /api/v1/users/{user_id}/profile?days={days}
```

**Response:**
```json
{
  "user_id": "user_001",
  "home": {"grid_id": "...", "confidence": 0.94},
  "work": {"grid_id": "...", "confidence": 0.88},
  "regularity_score": 0.82,
  "mobility_score": 0.45,
  "occupation": "commuter"
}
```

#### Get User Events

```
GET /api/v1/users/{user_id}/events
```

#### Batch Analyze

```
POST /api/v1/batch/analyze
```

**Request:** `{"user_ids": ["id1", "id2"]}`
