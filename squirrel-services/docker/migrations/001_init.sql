-- ============================================
-- Squirrel Database Schema v1.2.0
-- Initial Migration
-- ============================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS squirrel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE squirrel;

-- ============================================
-- 位置数据表
-- ============================================

-- 原始位置点
CREATE TABLE IF NOT EXISTS locations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    timestamp DATETIME NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    accuracy FLOAT,
    speed FLOAT,
    provider VARCHAR(20),
    grid_id VARCHAR(32),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_time (user_id, timestamp),
    INDEX idx_grid (grid_id),
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 停留点表
CREATE TABLE IF NOT EXISTS staypoints (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    center_lat DECIMAL(10, 8) NOT NULL,
    center_lng DECIMAL(11, 8) NOT NULL,
    grid_id VARCHAR(32) NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    duration_minutes INT,
    point_count INT,
    nearest_station_id VARCHAR(32),
    distance_to_station INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_time (user_id, start_time),
    INDEX idx_grid (grid_id),
    INDEX idx_station (nearest_station_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================
-- 用户画像表
-- ============================================

-- 用户家位置历史
CREATE TABLE IF NOT EXISTS user_home_locations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    grid_id VARCHAR(32) NOT NULL,
    lat DECIMAL(10, 8) NOT NULL,
    lng DECIMAL(11, 8) NOT NULL,
    confidence FLOAT,
    consecutive_days INT DEFAULT 0,
    detected_at DATETIME NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user (user_id),
    INDEX idx_user_time (user_id, detected_at),
    UNIQUE KEY uk_user_grid (user_id, grid_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 通勤记录
CREATE TABLE IF NOT EXISTS commute_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    departure_time DATETIME NOT NULL,
    arrival_time DATETIME,
    duration_minutes INT,
    start_station_id VARCHAR(32),
    end_station_id VARCHAR(32),
    is_valid BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_departure (user_id, departure_time),
    INDEX idx_route (start_station_id, end_station_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 用户每日统计
CREATE TABLE IF NOT EXISTS user_daily_stats (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    date DATE NOT NULL,
    is_weekday BOOLEAN,
    total_points INT DEFAULT 0,
    commute_count INT DEFAULT 0,
    home_hours FLOAT,
    work_hours FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_date (user_id, date),
    UNIQUE KEY uk_user_date (user_id, date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================
-- 生活事件表
-- ============================================

CREATE TABLE IF NOT EXISTS life_events (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    event_type VARCHAR(32) NOT NULL,
    confidence VARCHAR(16) NOT NULL,
    detected_at DATETIME NOT NULL,
    event_date DATE,
    details JSON,
    evidence JSON,
    is_notified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user (user_id),
    INDEX idx_user_type (user_id, event_type),
    INDEX idx_detected_at (detected_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================
-- 地铁站点数据
-- ============================================

CREATE TABLE IF NOT EXISTS metro_stations (
    id VARCHAR(32) PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    line_name VARCHAR(32) NOT NULL,
    line_id VARCHAR(16) NOT NULL,
    lat DECIMAL(10, 8) NOT NULL,
    lng DECIMAL(11, 8) NOT NULL,
    is_transfer BOOLEAN DEFAULT FALSE,
    transfer_to JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_line (line_id),
    INDEX idx_location (lat, lng)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================
-- 配置和元数据
-- ============================================

CREATE TABLE IF NOT EXISTS system_config (
    config_key VARCHAR(64) PRIMARY KEY,
    config_value TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 插入初始配置
INSERT INTO system_config (config_key, config_value) VALUES
('schema_version', '1.2.0'),
('speed_threshold_kmh', '20')
ON DUPLICATE KEY UPDATE config_value = VALUES(config_value);

-- ============================================
-- 视图
-- ============================================

CREATE OR REPLACE VIEW v_active_users AS
SELECT 
    user_id,
    COUNT(DISTINCT date) as active_days,
    SUM(total_points) as total_points
FROM user_daily_stats
WHERE date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
GROUP BY user_id
HAVING active_days >= 10;
