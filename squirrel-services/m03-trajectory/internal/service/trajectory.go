package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/squirrelawake/m03-trajectory/internal/algorithm"
	"github.com/squirrelawake/m03-trajectory/internal/storage"
	pb "github.com/squirrelawake/m03-trajectory/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TrajectoryService struct {
	pb.UnimplementedTrajectoryProcessorServer
	pb.UnimplementedLocationServiceServer
	db          *sql.DB
	staypointAlg *algorithm.StayPointDetector
	matcher     *algorithm.StationMatcher
	metrics     *MetricsCollector
}

func NewTrajectoryService(db *sql.DB) *TrajectoryService {
	return &TrajectoryService{
		db:           db,
		staypointAlg: algorithm.NewStayPointDetector(),
		matcher:      algorithm.NewStationMatcher(),
		metrics:      NewMetricsCollector(),
	}
}

func (s *TrajectoryService) ProcessBatch(ctx context.Context, req *pb.BatchRequest) (*pb.BatchResponse, error) {
	start := time.Now()
	
	// 1. 插入原始位置点
	if err := s.insertLocationBatch(ctx, req.UserId, req.Locations); err != nil {
		return nil, fmt.Errorf("insert batch failed: %w", err)
	}

	// 2. 检测停留点
	staypoints := s.detectStaypoints(req.UserId, req.Locations)
	
	// 3. 匹配最近站点
	for _, sp := range staypoints {
		station, distance := s.matcher.FindNearestStation(sp.CenterLat, sp.CenterLng)
		if station != nil && distance < 500 { // 500m阈值
			sp.NearestStationId = station.ID
			sp.DistanceToStation = int32(distance)
		}
	}

	// 4. 更新用户轨迹摘要
	if err := s.updateTrajectorySummary(ctx, req.UserId, req.Locations, staypoints); err != nil {
		return nil, err
	}

	// 记录指标
	duration := time.Since(start)
	s.metrics.RecordProcessingTime(duration)

	return &pb.BatchResponse{
		ProcessedCount: int32(len(req.Locations)),
		StaypointCount: int32(len(staypoints)),
		ProcessTimeMs:  int64(duration.Milliseconds()),
	}, nil
}

func (s *TrajectoryService) GetTrajectory(ctx context.Context, req *pb.TrajectoryRequest) (*pb.TrajectoryResponse, error) {
	query := `
		SELECT timestamp, latitude, longitude, speed, accuracy, provider
		FROM locations
		WHERE user_id = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp
	`
	
	rows, err := s.db.QueryContext(ctx, query, req.UserId, 
		req.StartTime.AsTime(), req.EndTime.AsTime())
	if err != nil {
		return nil, fmt.Errorf("query trajectory failed: %w", err)
	}
	defer rows.Close()

	var locations []*pb.Location
	for rows.Next() {
		var loc pb.Location
		var ts time.Time
		err := rows.Scan(&ts, &loc.Latitude, &loc.Longitude, &loc.Speed, &loc.Accuracy, &loc.Provider)
		if err != nil {
			continue
		}
		loc.Timestamp = timestamppb.New(ts)
		locations = append(locations, &loc)
	}

	return &pb.TrajectoryResponse{
		UserId:    req.UserId,
		Locations: locations,
		Count:     int32(len(locations)),
	}, nil
}

func (s *TrajectoryService) GetSpeed(ctx context.Context, req *pb.SpeedRequest) (*pb.SpeedResponse, error) {
	// 获取最近两个点计算速度
	query := `
		SELECT latitude, longitude, timestamp, speed
		FROM locations
		WHERE user_id = ?
		ORDER BY timestamp DESC
		LIMIT 2
	`
	
	rows, err := s.db.QueryContext(ctx, query, req.UserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []struct {
		lat, lng float64
		ts       time.Time
		speed    float64
	}
	
	for rows.Next() {
		var p struct {
			lat, lng float64
			ts       time.Time
			speed    float64
		}
		rows.Scan(&p.lat, &p.lng, &p.ts, &p.speed)
		points = append(points, p)
	}

	if len(points) < 2 {
		return &pb.SpeedResponse{
			UserId:      req.UserId,
			SpeedKmh:    0,
			IsMoving:    false,
			Confidence:  0.5,
		}, nil
	}

	// 计算速度
	distance := haversine(points[1].lat, points[1].lng, points[0].lat, points[0].lng)
	duration := points[0].ts.Sub(points[1].ts).Hours()
	
	var speedKmh float64
	if duration > 0 {
		speedKmh = distance / duration
	}

	// 使用DB速度作为补充
	if points[0].speed > 0 {
		speedKmh = (speedKmh + points[0].speed) / 2 // 融合
	}

	return &pb.SpeedResponse{
		UserId:       req.UserId,
		SpeedKmh:     speedKmh,
		IsMoving:     speedKmh > 5,
		IsTransit:    speedKmh > 20,
		Confidence:   calculateSpeedConfidence(speedKmh, points[0].ts.Sub(points[1].ts)),
		LastLocation: &pb.Location{
			Latitude:  points[0].lat,
			Longitude: points[0].lng,
			Speed:     speedKmh,
		},
	}, nil
}

func (s *TrajectoryService) GetNearestStation(ctx context.Context, req *pb.NearestStationRequest) (*pb.NearestStationResponse, error) {
	station, distance := s.matcher.FindNearestStation(req.Latitude, req.Longitude)
	if station == nil {
		return nil, fmt.Errorf("no stations found")
	}

	return &pb.NearestStationResponse{
		StationId:   station.ID,
		Name:        station.Name,
		LineName:    station.LineName,
		Latitude:    station.Lat,
		Longitude:   station.Lng,
		Distance:    int32(distance),
		IsTransfer:  station.IsTransfer,
		TransferTo:  station.TransferTo,
	}, nil
}

func (s *TrajectoryService) insertLocationBatch(ctx context.Context, userID string, locations []*pb.Location) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO locations (user_id, timestamp, latitude, longitude, accuracy, speed, provider, grid_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, loc := range locations {
		gridID := generateGridID(loc.Latitude, loc.Longitude, 100)
		_, err := stmt.ExecContext(ctx, userID, loc.Timestamp.AsTime(),
			loc.Latitude, loc.Longitude, loc.Accuracy, loc.Speed, loc.Provider, gridID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *TrajectoryService) detectStaypoints(userID string, locations []*pb.Location) []*pb.Staypoint {
	if len(locations) < 5 {
		return nil
	}

	// 转换为算法需要的格式
	points := make([]algorithm.GPSPoint, len(locations))
	for i, loc := range locations {
		points[i] = algorithm.GPSPoint{
			Lat: loc.Latitude,
			Lng: loc.Longitude,
			Ts:  loc.Timestamp.AsTime(),
		}
	}

	// 调用DBSCAN检测停留点
	clusters := s.staypointAlg.Detect(points)
	
	var staypoints []*pb.Staypoint
	for _, cluster := range clusters {
		center := cluster.Center()
		staypoints = append(staypoints, &pb.Staypoint{
			UserId:        userID,
			CenterLat:     center.Lat,
			CenterLng:     center.Lng,
			StartTime:     timestamppb.New(cluster.StartTime()),
			EndTime:       timestamppb.New(cluster.EndTime()),
			Duration:      int32(cluster.Duration().Minutes()),
			PointCount:    int32(cluster.PointCount()),
		})
	}

	return staypoints
}

func (s *TrajectoryService) updateTrajectorySummary(ctx context.Context, userID string, 
	locations []*pb.Location, staypoints []*pb.Staypoint) error {
	
	summary := storage.TrajectorySummary{
		UserID:         userID,
		Date:           time.Now().Format("2006-01-02"),
		TotalPoints:    len(locations),
		StaypointCount: len(staypoints),
		UpdatedAt:      time.Now(),
	}

	// 序列化停留点
	if len(staypoints) > 0 {
		staypointJSON, _ := json.Marshal(staypoints)
		summary.StaypointsJSON = string(staypointJSON)
	}

	return storage.SaveTrajectorySummary(ctx, s.db, summary)
}

func generateGridID(lat, lng float64, precision int) string {
	x := int(math.Floor(lat * float64(precision)))
	y := int(math.Floor(lng * float64(precision)))
	return fmt.Sprintf("grid_%d_%d", x, y)
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // 地球半径km
	
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func calculateSpeedConfidence(speed float64, duration time.Duration) float32 {
	confidence := 0.5
	
	// 基于速度范围
	if speed < 5 {
		confidence += 0.2 // 静止状态更确定
	} else if speed > 30 {
		confidence += 0.15 // 高速也较确定
	}
	
	// 基于采样间隔
	if duration > 30*time.Second && duration < 2*time.Minute {
		confidence += 0.2
	} else if duration > 5*time.Minute {
		confidence -= 0.1 // 间隔太长不确定性增加
	}
	
	if confidence > 1.0 {
		confidence = 1.0
	}
	return float32(confidence)
}
