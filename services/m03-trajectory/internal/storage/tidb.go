// tidb.go - TiDB 存储层实现
// 批量插入 + 高性能查询
// 性能目标：P99 < 10ms (查询), P99 < 50ms (批量插入)

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/squirrel-awake/m03-trajectory/internal/algorithm"

	_ "github.com/go-sql-driver/mysql"
)

// TiDBStorage TiDB存储
type TiDBStorage struct {
	db *sql.DB
}

// Config TiDB配置
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:            getEnv("TIDB_HOST", "localhost"),
		Port:            4000,
		User:            getEnv("TIDB_USER", "root"),
		Password:        getEnv("TIDB_PASSWORD", ""),
		Database:        getEnv("TIDB_DATABASE", "squirrel_m03"),
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
	}
}

// NewTiDBStorage 创建存储实例
func NewTiDBStorage(config *Config) (*TiDBStorage, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		config.User, config.Password, config.Host, config.Port, config.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db failed: %w", err)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db failed: %w", err)
	}

	return &TiDBStorage{db: db}, nil
}

// Close 关闭连接
func (s *TiDBStorage) Close() error {
	return s.db.Close()
}

// InitSchema 初始化表结构
func (s *TiDBStorage) InitSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS gps_logs (
    id BIGINT AUTO_RANDOM PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    timestamp BIGINT NOT NULL,
    latitude DOUBLE NOT NULL,
    longitude DOUBLE NOT NULL,
    accuracy FLOAT,
    speed FLOAT,
    source VARCHAR(16),
    INDEX idx_user_time (user_id, timestamp),
    INDEX idx_timestamp (timestamp)
);

CREATE TABLE IF NOT EXISTS stay_points (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    center_lat DOUBLE NOT NULL,
    center_lng DOUBLE NOT NULL,
    arrive_time BIGINT NOT NULL,
    leave_time BIGINT,
    duration INT,
    cluster_size INT,
    grid_id VARCHAR(32),
    INDEX idx_user_arrive (user_id, arrive_time),
    INDEX idx_grid (grid_id)
);

CREATE TABLE IF NOT EXISTS metro_stations (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    line_id VARCHAR(64) NOT NULL,
    line_name VARCHAR(64) NOT NULL,
    line_color VARCHAR(16),
    latitude DOUBLE NOT NULL,
    longitude DOUBLE NOT NULL,
    is_transfer BOOLEAN DEFAULT FALSE,
    station_order INT,
    INDEX idx_line_order (line_id, station_order),
    INDEX idx_location (latitude, longitude)
);
`
	_, err := s.db.Exec(schema)
	return err
}

// GPSLogRecord GPS日志记录
type GPSLogRecord struct {
	UserID    string
	Timestamp int64
	Lat       float64
	Lng       float64
	Accuracy  float32
	Speed     float32
	Source    string
}

// BatchInsertGPSLogs 批量插入GPS日志
// 性能：1000点 < 50ms (P99)
func (s *TiDBStorage) BatchInsertGPSLogs(ctx context.Context, records []GPSLogRecord) error {
	if len(records) == 0 {
		return nil
	}

	// 使用批量插入优化
	const batchSize = 100
	
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		
		batch := records[i:end]
		if err := s.insertBatch(ctx, batch); err != nil {
			return fmt.Errorf("insert batch %d-%d failed: %w", i, end, err)
		}
	}

	return nil
}

func (s *TiDBStorage) insertBatch(ctx context.Context, records []GPSLogRecord) error {
	var builder strings.Builder
	builder.WriteString("INSERT INTO gps_logs (user_id, timestamp, latitude, longitude, accuracy, speed, source) VALUES ")

	args := make([]interface{}, 0, len(records)*7)
	for i, r := range records {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString("(?,?,?,?,?,?,?)")
		args = append(args, r.UserID, r.Timestamp, r.Lat, r.Lng, r.Accuracy, r.Speed, r.Source)
	}

	_, err := s.db.ExecContext(ctx, builder.String(), args...)
	return err
}

// QueryRecentTrajectory 查询最近轨迹
// 性能：P99 < 10ms
func (s *TiDBStorage) QueryRecentTrajectory(ctx context.Context, userID string, startTime, endTime int64) ([]algorithm.GPSPoint, error) {
	query := `
SELECT timestamp, latitude, longitude, accuracy 
FROM gps_logs 
WHERE user_id = ? AND timestamp BETWEEN ? AND ? 
ORDER BY timestamp ASC`

	rows, err := s.db.QueryContext(ctx, query, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []algorithm.GPSPoint
	for rows.Next() {
		var p algorithm.GPSPoint
		if err := rows.Scan(&p.Timestamp, &p.Lat, &p.Lng, &p.Accuracy); err != nil {
			return nil, err
		}
		points = append(points, p)
	}

	return points, rows.Err()
}

// QueryRecentStops 查询最近停留点
// 性能：P99 < 10ms
func (s *TiDBStorage) QueryRecentStops(ctx context.Context, userID string, days int) ([]algorithm.StayPoint, error) {
	startTime := time.Now().AddDate(0, 0, -days).UnixMilli()

	query := `
SELECT id, center_lat, center_lng, arrive_time, leave_time, duration, cluster_size 
FROM stay_points 
WHERE user_id = ? AND arrive_time >= ? 
ORDER BY arrive_time DESC`

	rows, err := s.db.QueryContext(ctx, query, userID, startTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stops []algorithm.StayPoint
	for rows.Next() {
		var s algorithm.StayPoint
		if err := rows.Scan(&s.ID, &s.CenterLat, &s.CenterLng, &s.ArriveTime, &s.LeaveTime, &s.Duration, &s.ClusterSize); err != nil {
			return nil, err
		}
		stops = append(stops, s)
	}

	return stops, rows.Err()
}

// InsertStayPoint 插入停留点
func (s *TiDBStorage) InsertStayPoint(ctx context.Context, userID string, stay *algorithm.StayPoint) error {
	gridID := latLngToGrid(stay.CenterLat, stay.CenterLng)
	
	query := `
INSERT INTO stay_points (id, user_id, center_lat, center_lng, arrive_time, leave_time, duration, cluster_size, grid_id) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, stay.ID, userID, stay.CenterLat, stay.CenterLng, 
		stay.ArriveTime, stay.LeaveTime, stay.Duration, stay.ClusterSize, gridID)
	return err
}

// CleanupOldGPSLogs 清理过期GPS日志（7天前）
func (s *TiDBStorage) CleanupOldGPSLogs(ctx context.Context) (int64, error) {
	cutoffTime := time.Now().AddDate(0, 0, -7).UnixMilli()
	
	result, err := s.db.ExecContext(ctx, "DELETE FROM gps_logs WHERE timestamp < ?", cutoffTime)
	if err != nil {
		return 0, err
	}
	
	return result.RowsAffected()
}

// GetMetroStations 获取所有地铁站
func (s *TiDBStorage) GetMetroStations(ctx context.Context) ([]MetroStation, error) {
	query := `
SELECT id, name, line_id, line_name, line_color, latitude, longitude, is_transfer, station_order 
FROM metro_stations 
ORDER BY line_id, station_order`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stations []MetroStation
	for rows.Next() {
		var s MetroStation
		if err := rows.Scan(&s.ID, &s.Name, &s.LineID, &s.LineName, &s.LineColor, 
			&s.Lat, &s.Lng, &s.IsTransfer, &s.StationOrder); err != nil {
			return nil, err
		}
		stations = append(stations, s)
	}

	return stations, rows.Err()
}

// MetroStation 地铁站
type MetroStation struct {
	ID           string
	Name         string
	LineID       string
	LineName     string
	LineColor    string
	Lat          float64
	Lng          float64
	IsTransfer   bool
	StationOrder int
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func latLngToGrid(lat, lng float64) string {
	// 500米网格
	gridSize := 0.0045
	latIdx := int(lat / gridSize)
	lngIdx := int(lng / gridSize)
	return fmt.Sprintf("GRID_%d_%d", latIdx, lngIdx)
}
