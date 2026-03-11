package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	pb "github.com/squirrelawake/m03-trajectory/proto"
)

// TrajectorySummary 用户轨迹摘要
type TrajectorySummary struct {
	UserID         string    `json:"user_id"`
	Date           string    `json:"date"`
	TotalPoints    int       `json:"total_points"`
	StaypointCount int       `json:"staypoint_count"`
	StaypointsJSON string    `json:"staypoints_json"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func SaveTrajectorySummary(ctx context.Context, db *sql.DB, summary TrajectorySummary) error {
	query := `
		INSERT INTO trajectory_summaries 
		(user_id, date, total_points, staypoint_count, staypoints_json, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		total_points = VALUES(total_points),
		staypoint_count = VALUES(staypoint_count),
		staypoints_json = VALUES(staypoints_json),
		updated_at = VALUES(updated_at)
	`
	
	_, err := db.ExecContext(ctx, query,
		summary.UserID, summary.Date, summary.TotalPoints,
		summary.StaypointCount, summary.StaypointsJSON, summary.UpdatedAt)
	
	return err
}

func GetTrajectorySummary(ctx context.Context, db *sql.DB, userID, date string) (*TrajectorySummary, error) {
	query := `
		SELECT user_id, date, total_points, staypoint_count, staypoints_json, updated_at
		FROM trajectory_summaries
		WHERE user_id = ? AND date = ?
	`
	
	var summary TrajectorySummary
	err := db.QueryRowContext(ctx, query, userID, date).Scan(
		&summary.UserID, &summary.Date, &summary.TotalPoints,
		&summary.StaypointCount, &summary.StaypointsJSON, &summary.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &summary, nil
}

func GetUserDailyStats(ctx context.Context, db *sql.DB, userID string, startDate, endDate string) ([]*pb.DailyStats, error) {
	query := `
		SELECT date, total_points, staypoint_count
		FROM trajectory_summaries
		WHERE user_id = ? AND date BETWEEN ? AND ?
		ORDER BY date
	`
	
	rows, err := db.QueryContext(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var stats []*pb.DailyStats
	for rows.Next() {
		var s pb.DailyStats
		var dateStr string
		if err := rows.Scan(&dateStr, &s.TotalPoints, &s.StaypointCount); err != nil {
			continue
		}
		
		// 解析日期
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			s.Date = timestamppb.New(t)
		}
		s.UserId = userID
		stats = append(stats, &s)
	}
	
	return stats, nil
}

// DailyMetrics 日常指标
type DailyMetrics struct {
	UserID          string  `json:"user_id"`
	Date            string  `json:"date"`
	TotalDistance   float64 `json:"total_distance_km"`   // 总行程(公里)
	MaxSpeed        float64 `json:"max_speed_kmh"`       // 最高速度
	AvgSpeed        float64 `json:"avg_speed_kmh"`       // 平均速度
	TransitDuration int     `json:"transit_duration_min"` // 地铁时间
	UniqueStations  int     `json:"unique_stations"`     // 访问站点数
}

func SaveDailyMetrics(ctx context.Context, db *sql.DB, metrics DailyMetrics) error {
	query := `
		INSERT INTO daily_metrics
		(user_id, date, total_distance, max_speed, avg_speed, transit_duration, unique_stations)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		total_distance = VALUES(total_distance),
		max_speed = VALUES(max_speed),
		avg_speed = VALUES(avg_speed),
		transit_duration = VALUES(transit_duration),
		unique_stations = VALUES(unique_stations)
	`
	
	_, err := db.ExecContext(ctx, query,
		metrics.UserID, metrics.Date, metrics.TotalDistance,
		metrics.MaxSpeed, metrics.AvgSpeed, metrics.TransitDuration, metrics.UniqueStations)
	
	return err
}

func CalculateDailyMetrics(ctx context.Context, db *sql.DB, userID string, date string) (*DailyMetrics, error) {
	// 查询当日所有位置点
	query := `
		SELECT latitude, longitude, speed
		FROM locations
		WHERE user_id = ? AND DATE(timestamp) = ?
		ORDER BY timestamp
	`
	
	rows, err := db.QueryContext(ctx, query, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var points []struct {
		lat, lng, speed float64
	}
	
	for rows.Next() {
		var p struct {
			lat, lng, speed float64
		}
		if err := rows.Scan(&p.lat, &p.lng, &p.speed); err != nil {
			continue
		}
		points = append(points, p)
	}
	
	if len(points) < 2 {
		return nil, fmt.Errorf("insufficient data")
	}
	
	// 计算指标
	metrics := &DailyMetrics{
		UserID: userID,
		Date:   date,
	}
	
	for i := 1; i < len(points); i++ {
		dist := haversineDistance(points[i-1].lat, points[i-1].lng, points[i].lat, points[i].lng)
		metrics.TotalDistance += dist / 1000 // 转公里
		
		if points[i].speed > metrics.MaxSpeed {
			metrics.MaxSpeed = points[i].speed
		}
	}
	
	if len(points) > 0 {
		var totalSpeed float64
		for _, p := range points {
			totalSpeed += p.speed
		}
		metrics.AvgSpeed = totalSpeed / float64(len(points))
	}
	
	return metrics, nil
}
