package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	pb "github.com/squirrel-awake/m03-trajectory/proto"
)

// TrajectoryService 实现轨迹处理服务
type TrajectoryService struct {
	pb.UnimplementedTrajectoryServiceServer

	mu          sync.RWMutex
	userPoints  map[string][]*pb.GPSPoint // 用户最近GPS点（内存缓存）
	metroLines  map[string]*MetroLine     // 地铁线路数据
	ready       bool
}

// MetroLine 地铁线路
type MetroLine struct {
	ID        string
	Name      string
	Color     string
	Stations  []*Station
}

// Station 站点
type Station struct {
	ID        string
	Name      string
	Lat       float64
	Lng       float64
	IsTransfer bool
}

// NewTrajectoryService 创建服务实例
func NewTrajectoryService() *TrajectoryService {
	svc := &TrajectoryService{
		userPoints: make(map[string][]*pb.GPSPoint),
		metroLines: make(map[string]*MetroLine),
		ready:      true,
	}

	// 初始化地铁数据（实际应从数据库加载）
	svc.initMetroData()

	return svc
}

// Ready 检查服务就绪状态
func (s *TrajectoryService) Ready() bool {
	return s.ready
}

// ===== 接口实现 =====

// GetCurrentGrid 获取当前网格（脱敏坐标）
func (s *TrajectoryService) GetCurrentGrid(ctx context.Context, req *pb.GetCurrentGridRequest) (*pb.GetCurrentGridResponse, error) {
	points := s.getUserPoints(req.UserId)
	if len(points) == 0 {
		return &pb.GetCurrentGridResponse{GridId: ""}, nil
	}

	// 取最新点
	latest := points[len(points)-1]
	
	// 网格化（500m精度）
	gridId := latLngToGrid(latest.Latitude, latest.Longitude)
	
	return &pb.GetCurrentGridResponse{GridId: gridId}, nil
}

// GetSpeed 获取平滑速度
func (s *TrajectoryService) GetSpeed(ctx context.Context, req *pb.GetSpeedRequest) (*pb.GetSpeedResponse, error) {
	points := s.getUserPoints(req.UserId)
	if len(points) < 2 {
		return &pb.GetSpeedResponse{SpeedKmh: 0}, nil
	}

	// 计算平滑速度（简化版，实际应使用卡尔曼滤波）
	speed := s.calculateSmoothSpeed(points)
	
	return &pb.GetSpeedResponse{SpeedKmh: speed}, nil
}

// IsInMetroArea 判断是否地铁区域
func (s *TrajectoryService) IsInMetroArea(ctx context.Context, req *pb.IsInMetroAreaRequest) (*pb.IsInMetroAreaResponse, error) {
	points := s.getUserPoints(req.UserId)
	if len(points) == 0 {
		return &pb.IsInMetroAreaResponse{InMetroArea: false}, nil
	}

	latest := points[len(points)-1]
	
	// 检查是否在任一站点150米内
	for _, line := range s.metroLines {
		for _, station := range line.Stations {
			dist := haversineDistance(latest.Latitude, latest.Longitude, station.Lat, station.Lng)
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

// GetRecentStops 获取最近停留点
func (s *TrajectoryService) GetRecentStops(ctx context.Context, req *pb.GetRecentStopsRequest) (*pb.GetRecentStopsResponse, error) {
	// TODO: 从 TiDB 查询
	// 简化实现：返回空列表
	return &pb.GetRecentStopsResponse{Stops: []*pb.StayPoint{}}, nil
}

// GetTrajectory 获取轨迹片段
func (s *TrajectoryService) GetTrajectory(ctx context.Context, req *pb.GetTrajectoryRequest) (*pb.GetTrajectoryResponse, error) {
	points := s.getUserPoints(req.UserId)
	
	// 过滤时间范围
	var result []*pb.GPSPoint
	for _, p := range points {
		if p.Timestamp >= req.StartTime && p.Timestamp <= req.EndTime {
			result = append(result, p)
		}
	}

	return &pb.GetTrajectoryResponse{Points: result}, nil
}

// GetNextTransferStation 获取下一个换乘站预测（BLK-001 已补充字段）
func (s *TrajectoryService) GetNextTransferStation(ctx context.Context, req *pb.GetNextTransferStationRequest) (*pb.GetNextTransferStationResponse, error) {
	points := s.getUserPoints(req.UserId)
	if len(points) < 5 {
		return &pb.GetNextTransferStationResponse{HasResult: false}, nil
	}

	// 简化实现：模拟返回数据
	latest := points[len(points)-1]
	
	// 查找最近线路和站点
	for _, line := range s.metroLines {
		for i, station := range line.Stations {
			dist := haversineDistance(latest.Latitude, latest.Longitude, station.Lat, station.Lng)
			if dist <= 500 { // 在500米内
				// 计算预计到站时间（简化）
				speed := s.calculateSmoothSpeed(points)
				var etaSeconds int32 = 180
				if speed > 0 {
					etaSeconds = int32(dist / (speed / 3.6))
				}

				// 查找下一站
				nextStationName := ""
				if i+1 < len(line.Stations) {
					nextStationName = line.Stations[i+1].Name
				}

				// 计算到换乘站还需几站
				stationsToTransfer := s.countStationsToTransfer(line, i)

				stationInfo := &pb.StationInfo{
					StationId:          station.ID,
					StationName:        station.Name,
					LineName:           line.Name,
					LineColor:          line.Color,
					Direction:          "往四惠东方向",
					DirectionCode:      "east",
					EstimatedArrivalSeconds: etaSeconds,  // BLK-001: 单位为秒
					Confidence:         0.85,
					DistanceMeters:     int32(dist),
					IsTransfer:         station.IsTransfer,  // BLK-001 新增
					StationsToTransfer: int32(stationsToTransfer), // BLK-001 新增
					NextStationName:    nextStationName,     // BLK-001 新增
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

// ReportGPSPoint 上报GPS点
func (s *TrajectoryService) ReportGPSPoint(ctx context.Context, req *pb.ReportGPSPointRequest) (*pb.ReportGPSPointResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 添加到用户点缓存
	s.userPoints[req.UserId] = append(s.userPoints[req.UserId], req.Point)
	
	// 限制缓存大小（保留最近100个点）
	if len(s.userPoints[req.UserId]) > 100 {
		s.userPoints[req.UserId] = s.userPoints[req.UserId][len(s.userPoints[req.UserId])-100:]
	}

	log.Printf("📍 Received GPS point from user %s: (%.4f, %.4f)", 
		req.UserId, req.Point.Latitude, req.Point.Longitude)

	return &pb.ReportGPSPointResponse{Success: true, Message: "OK"}, nil
}

// ===== 辅助方法 =====

func (s *TrajectoryService) getUserPoints(userId string) []*pb.GPSPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userPoints[userId]
}

func (s *TrajectoryService) calculateSmoothSpeed(points []*pb.GPSPoint) float64 {
	if len(points) < 2 {
		return 0
	}

	// 简化版：计算最近两个点的速度
	latest := points[len(points)-1]
	prev := points[len(points)-2]

	dist := haversineDistance(latest.Latitude, latest.Longitude, prev.Latitude, prev.Longitude)
	timeDiff := float64(latest.Timestamp-prev.Timestamp) / 1000.0 // 秒

	if timeDiff > 0 {
		return (dist / timeDiff) * 3.6 // km/h
	}
	return 0
}

func (s *TrajectoryService) countStationsToTransfer(line *MetroLine, currentIdx int) int {
	// 计算从当前站到下一个换乘站的站数
	count := 0
	for i := currentIdx; i < len(line.Stations); i++ {
		if line.Stations[i].IsTransfer {
			return count
		}
		count++
	}
	return -1 // 无后续换乘站
}

// latLngToGrid 将经纬度转换为网格ID
func latLngToGrid(lat, lng float64) string {
	// 500米网格
	gridSize := 0.0045 // 约500米
	latIdx := int(math.Floor(lat / gridSize))
	lngIdx := int(math.Floor(lng / gridSize))
	return fmt.Sprintf("GRID_%d_%d", latIdx, lngIdx)
}

// haversineDistance 计算两点间距离（米）
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000 // 地球半径（米）

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

// initMetroData 初始化地铁数据
func (s *TrajectoryService) initMetroData() {
	// 模拟数据：1号线
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
			{ID: "s001_008", Name: "公主坟站", Lat: 39.912, Lng: 116.31, IsTransfer: true},  // 换乘10号线
			{ID: "s001_009", Name: "军事博物馆站", Lat: 39.913, Lng: 116.33, IsTransfer: true}, // 换乘9号线
			{ID: "s001_010", Name: "木樨地站", Lat: 39.914, Lng: 116.35, IsTransfer: false},
			{ID: "s001_011", Name: "南礼士路站", Lat: 39.915, Lng: 116.36, IsTransfer: false},
			{ID: "s001_012", Name: "复兴门站", Lat: 39.916, Lng: 116.37, IsTransfer: true},  // 换乘2号线
			{ID: "s001_013", Name: "西单站", Lat: 39.917, Lng: 116.38, IsTransfer: true},    // 换乘4号线
			{ID: "s001_014", Name: "天安门西站", Lat: 39.918, Lng: 116.40, IsTransfer: false},
			{ID: "s001_015", Name: "天安门东站", Lat: 39.919, Lng: 116.41, IsTransfer: false},
			{ID: "s001_016", Name: "王府井站", Lat: 39.920, Lng: 116.42, IsTransfer: true},  // 换乘8号线
			{ID: "s001_017", Name: "东单站", Lat: 39.921, Lng: 116.43, IsTransfer: true},    // 换乘5号线
			{ID: "s001_018", Name: "建国门站", Lat: 39.922, Lng: 116.44, IsTransfer: true},  // 换乘2号线
			{ID: "s001_019", Name: "永安里站", Lat: 39.923, Lng: 116.46, IsTransfer: false},
			{ID: "s001_020", Name: "国贸站", Lat: 39.924, Lng: 116.47, IsTransfer: true},    // 换乘10号线
			{ID: "s001_021", Name: "大望路站", Lat: 39.925, Lng: 116.48, IsTransfer: true},  // 换乘14号线
			{ID: "s001_022", Name: "四惠站", Lat: 39.926, Lng: 116.50, IsTransfer: false},
			{ID: "s001_023", Name: "四惠东站", Lat: 39.927, Lng: 116.52, IsTransfer: false},
		},
	}
}
