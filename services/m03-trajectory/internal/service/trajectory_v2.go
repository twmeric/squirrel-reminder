// trajectory_v2.go - v1.2 优化版轨迹服务实现
// 集成真实算法和TiDB存储

package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/squirrel-awake/m03-trajectory/internal/algorithm"
	"github.com/squirrel-awake/m03-trajectory/internal/storage"
	pb "github.com/squirrel-awake/m03-trajectory/proto"
)

// TrajectoryServiceV2 v1.2 轨迹服务
type TrajectoryServiceV2 struct {
	pb.UnimplementedTrajectoryServiceServer

	mu           sync.RWMutex
	stayDetector *algorithm.StayPointDetector
	speedSmoother *algorithm.SpeedSmoother
	driftFilter  *algorithm.GPSDriftFilter
	storage      *storage.TiDBStorage
	
	// 内存缓存（热点数据）
	userPoints   map[string][]algorithm.GPSPoint
	metroLines   map[string]*MetroLine
	ready        bool
}

// NewTrajectoryServiceV2 创建服务实例
func NewTrajectoryServiceV2(store *storage.TiDBStorage) *TrajectoryServiceV2 {
	svc := &TrajectoryServiceV2{
		stayDetector:  algorithm.NewStayPointDetector(),
		speedSmoother: algorithm.NewSpeedSmoother(),
		driftFilter:   algorithm.NewGPSDriftFilter(),
		storage:       store,
		userPoints:    make(map[string][]algorithm.GPSPoint),
		metroLines:    make(map[string]*MetroLine),
		ready:         true,
	}

	// 初始化地铁数据
	svc.initMetroData()
	
	// 启动后台任务
	go svc.cleanupTask()
	go svc.detectStayPointsTask()

	return svc
}

// Ready 检查服务就绪状态
func (s *TrajectoryServiceV2) Ready() bool {
	return s.ready && s.storage != nil
}

// ===== gRPC 接口实现 =====

// GetCurrentGrid 获取当前网格（P99 < 5ms）
func (s *TrajectoryServiceV2) GetCurrentGrid(ctx context.Context, req *pb.GetCurrentGridRequest) (*pb.GetCurrentGridResponse, error) {
	points := s.getCachedPoints(req.UserId)
	if len(points) == 0 {
		return &pb.GetCurrentGridResponse{GridId: ""}, nil
	}

	latest := points[len(points)-1]
	gridId := latLngToGrid(latest.Lat, latest.Lng)

	return &pb.GetCurrentGridResponse{GridId: gridId}, nil
}

// GetSpeed 获取平滑速度（P99 < 5ms）
func (s *TrajectoryServiceV2) GetSpeed(ctx context.Context, req *pb.GetSpeedRequest) (*pb.GetSpeedResponse, error) {
	points := s.getCachedPoints(req.UserId)
	
	speed := s.speedSmoother.GetCurrentSpeed(points)
	
	return &pb.GetSpeedResponse{SpeedKmh: speed}, nil
}

// IsInMetroArea 判断是否地铁区域（P99 < 10ms）
func (s *TrajectoryServiceV2) IsInMetroArea(ctx context.Context, req *pb.IsInMetroAreaRequest) (*pb.IsInMetroAreaResponse, error) {
	points := s.getCachedPoints(req.UserId)
	if len(points) == 0 {
		return &pb.IsInMetroAreaResponse{InMetroArea: false}, nil
	}

	latest := points[len(points)-1]
	
	// 检查是否在任一站点150米内
	for _, line := range s.metroLines {
		for _, station := range line.Stations {
			dist := haversineDistance(latest.Lat, latest.Lng, station.Lat, station.Lng)
			if dist <= 150 {
				return &pb.IsInMetroAreaResponse{
					InMetroArea:       true,
					NearestStationId: station.ID,
				}, nil
			}
		}
	}

	return &pb.IsInMetroAreaResponse{InMetroArea: false}, nil
}

// GetRecentStops 获取最近停留点（P99 < 10ms）
func (s *TrajectoryServiceV2) GetRecentStops(ctx context.Context, req *pb.GetRecentStopsRequest) (*pb.GetRecentStopsResponse, error) {
	// 优先从数据库查询
	stops, err := s.storage.QueryRecentStops(ctx, req.UserId, int(req.Days))
	if err != nil {
		log.Printf("QueryRecentStops error: %v", err)
		return &pb.GetRecentStopsResponse{Stops: []*pb.StayPoint{}}, nil
	}

	// 转换为protobuf格式
	pbStops := make([]*pb.StayPoint, len(stops))
	for i, s := range stops {
		pbStops[i] = &pb.StayPoint{
			Id:          s.ID,
			CenterLat:   s.CenterLat,
			CenterLng:   s.CenterLng,
			ArriveTime:  s.ArriveTime,
			LeaveTime:   s.LeaveTime,
			Duration:    s.Duration,
			ClusterSize: s.ClusterSize,
		}
	}

	return &pb.GetRecentStopsResponse{Stops: pbStops}, nil
}

// GetTrajectory 获取轨迹片段
func (s *TrajectoryServiceV2) GetTrajectory(ctx context.Context, req *pb.GetTrajectoryRequest) (*pb.GetTrajectoryResponse, error) {
	points, err := s.storage.QueryRecentTrajectory(ctx, req.UserId, req.StartTime, req.EndTime)
	if err != nil {
		log.Printf("QueryRecentTrajectory error: %v", err)
		return &pb.GetTrajectoryResponse{Points: []*pb.GPSPoint{}}, nil
	}

	pbPoints := make([]*pb.GPSPoint, len(points))
	for i, p := range points {
		pbPoints[i] = &pb.GPSPoint{
			Timestamp: p.Timestamp,
			Latitude:  p.Lat,
			Longitude: p.Lng,
			Accuracy:  p.Accuracy,
		}
	}

	return &pb.GetTrajectoryResponse{Points: pbPoints}, nil
}

// GetNextTransferStation 获取下一个换乘站（P99 < 50ms）
func (s *TrajectoryServiceV2) GetNextTransferStation(ctx context.Context, req *pb.GetNextTransferStationRequest) (*pb.GetNextTransferStationResponse, error) {
	points := s.getCachedPoints(req.UserId)
	if len(points) < 5 {
		return &pb.GetNextTransferStationResponse{HasResult: false}, nil
	}

	// 过滤漂移点
	filtered := s.driftFilter.Filter(points)
	if len(filtered) < 5 {
		filtered = points
	}

	latest := filtered[len(filtered)-1]

	// 查找最近线路和站点
	for _, line := range s.metroLines {
		for i, station := range line.Stations {
			dist := haversineDistance(latest.Lat, latest.Lng, station.Lat, station.Lng)
			if dist <= 500 { // 在500米内
				// 计算预计到站时间
				speed := s.speedSmoother.GetCurrentSpeed(filtered)
				etaSeconds := int32(180)
				if speed > 0 {
					etaSeconds = int32(dist / (speed / 3.6))
				}

				// 下一站信息
				nextStationName := ""
				if i+1 < len(line.Stations) {
					nextStationName = line.Stations[i+1].Name
				}

				// 计算到换乘站还需几站
				stationsToTransfer := s.countStationsToTransfer(line, i)

				stationInfo := &pb.StationInfo{
					StationId:               station.ID,
					StationName:             station.Name,
					LineName:                line.Name,
					LineColor:               line.Color,
					Direction:               "往四惠东方向",
					DirectionCode:           "east",
					EstimatedArrivalSeconds: etaSeconds,
					Confidence:              0.85,
					DistanceMeters:          int32(dist),
					IsTransfer:              station.IsTransfer,
					StationsToTransfer:      int32(stationsToTransfer),
					NextStationName:         nextStationName,
				}

				return &pb.GetNextTransferStationResponse{
					Station:   stationInfo,
					HasResult: true,
				}, nil
			}
		}
	}

	return &pb.GetNextTransferStationResponse{HasResult: false}, nil
}

// ReportGPSPoint 上报GPS点（P99 < 20ms）
func (s *TrajectoryServiceV2) ReportGPSPoint(ctx context.Context, req *pb.ReportGPSPointRequest) (*pb.ReportGPSPointResponse, error) {
	// 转换为算法格式
	point := algorithm.GPSPoint{
		Timestamp: req.Point.Timestamp,
		Lat:       req.Point.Latitude,
		Lng:       req.Point.Longitude,
		Accuracy:  req.Point.Accuracy,
	}

	// 更新内存缓存
	s.updateCache(req.UserId, point)

	// 异步写入数据库（批量）
	go s.asyncWriteToDB(req.UserId, point)

	return &pb.ReportGPSPointResponse{Success: true, Message: "OK"}, nil
}

// ===== 内部方法 =====

func (s *TrajectoryServiceV2) getCachedPoints(userID string) []algorithm.GPSPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userPoints[userID]
}

func (s *TrajectoryServiceV2) updateCache(userID string, point algorithm.GPSPoint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.userPoints[userID] = append(s.userPoints[userID], point)

	// 限制缓存大小（保留最近100个点）
	if len(s.userPoints[userID]) > 100 {
		s.userPoints[userID] = s.userPoints[userID][len(s.userPoints[userID])-100:]
	}
}

func (s *TrajectoryServiceV2) asyncWriteToDB(userID string, point algorithm.GPSPoint) {
	// TODO: 实现批量写入队列
	record := storage.GPSLogRecord{
		UserID:    userID,
		Timestamp: point.Timestamp,
		Lat:       point.Lat,
		Lng:       point.Lng,
		Accuracy:  point.Accuracy,
		Source:    "gps",
	}
	
	ctx := context.Background()
	if err := s.storage.BatchInsertGPSLogs(ctx, []storage.GPSLogRecord{record}); err != nil {
		log.Printf("Failed to insert GPS log: %v", err)
	}
}

func (s *TrajectoryServiceV2) countStationsToTransfer(line *MetroLine, currentIdx int) int {
	count := 0
	for i := currentIdx; i < len(line.Stations); i++ {
		if line.Stations[i].IsTransfer {
			return count
		}
		count++
	}
	return -1
}

// cleanupTask 定期清理过期数据
func (s *TrajectoryServiceV2) cleanupTask() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		n, err := s.storage.CleanupOldGPSLogs(ctx)
		if err != nil {
			log.Printf("Cleanup old GPS logs failed: %v", err)
		} else {
			log.Printf("Cleaned up %d old GPS logs", n)
		}
	}
}

// detectStayPointsTask 定期检测停留点
func (s *TrajectoryServiceV2) detectStayPointsTask() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for userID, points := range s.userPoints {
			if len(points) < 5 {
				continue
			}

			stays := s.stayDetector.Detect(points)
			if len(stays) > 0 {
				ctx := context.Background()
				for _, stay := range stays {
					if err := s.storage.InsertStayPoint(ctx, userID, &stay); err != nil {
						log.Printf("Insert stay point failed: %v", err)
					}
				}
			}
		}
		s.mu.Unlock()
	}
}

// ===== 辅助函数 =====

func latLngToGrid(lat, lng float64) string {
	gridSize := 0.0045
	latIdx := int(math.Floor(lat / gridSize))
	lngIdx := int(math.Floor(lng / gridSize))
	return fmt.Sprintf("GRID_%d_%d", latIdx, lngIdx)
}

func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// MetroLine 地铁线路（同前）
type MetroLine struct {
	ID       string
	Name     string
	Color    string
	Stations []*Station
}

type Station struct {
	ID         string
	Name       string
	Lat        float64
	Lng        float64
	IsTransfer bool
}

func (s *TrajectoryServiceV2) initMetroData() {
	s.metroLines["line_001"] = &MetroLine{
		ID:    "line_001",
		Name:  "1号线",
		Color: "#A4343A",
		Stations: []*Station{
			{ID: "s001_001", Name: "苹果园站", Lat: 39.905, Lng: 116.17, IsTransfer: false},
			{ID: "s001_002", Name: "古城站", Lat: 39.906, Lng: 116.19, IsTransfer: false},
			{ID: "s001_003", Name: "八角游乐园站", Lat: 39.907, Lng: 116.21, IsTransfer: false},
			{ID: "s001_004", Name: "八宝山站", Lat: 39.908, Lng: 116.23, IsTransfer: false},
			{ID: "s001_005", Name: "玉泉路站", Lat: 39.909, Lng: 116.25, IsTransfer: false},
			{ID: "s001_006", Name: "五棵松站", Lat: 39.910, Lng: 116.27, IsTransfer: false},
			{ID: "s001_007", Name: "万寿路站", Lat: 39.911, Lng: 116.29, IsTransfer: false},
			{ID: "s001_008", Name: "公主坟站", Lat: 39.912, Lng: 116.31, IsTransfer: true},
			{ID: "s001_009", Name: "军事博物馆站", Lat: 39.913, Lng: 116.33, IsTransfer: true},
			{ID: "s001_010", Name: "木樨地站", Lat: 39.914, Lng: 116.35, IsTransfer: false},
			{ID: "s001_011", Name: "南礼士路站", Lat: 39.915, Lng: 116.36, IsTransfer: false},
			{ID: "s001_012", Name: "复兴门站", Lat: 39.916, Lng: 116.37, IsTransfer: true},
			{ID: "s001_013", Name: "西单站", Lat: 39.917, Lng: 116.38, IsTransfer: true},
			{ID: "s001_014", Name: "天安门西站", Lat: 39.918, Lng: 116.40, IsTransfer: false},
			{ID: "s001_015", Name: "天安门东站", Lat: 39.919, Lng: 116.41, IsTransfer: false},
			{ID: "s001_016", Name: "王府井站", Lat: 39.920, Lng: 116.42, IsTransfer: true},
			{ID: "s001_017", Name: "东单站", Lat: 39.921, Lng: 116.43, IsTransfer: true},
			{ID: "s001_018", Name: "建国门站", Lat: 39.922, Lng: 116.44, IsTransfer: true},
			{ID: "s001_019", Name: "永安里站", Lat: 39.923, Lng: 116.46, IsTransfer: false},
			{ID: "s001_020", Name: "国贸站", Lat: 39.924, Lng: 116.47, IsTransfer: true},
			{ID: "s001_021", Name: "大望路站", Lat: 39.925, Lng: 116.48, IsTransfer: true},
			{ID: "s001_022", Name: "四惠站", Lat: 39.926, Lng: 116.50, IsTransfer: false},
			{ID: "s001_023", Name: "四惠东站", Lat: 39.927, Lng: 116.52, IsTransfer: false},
		},
	}
}
