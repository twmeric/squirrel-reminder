// staypoint.go - DBSCAN 停留点检测算法实现
// 时间阈值：15分钟，距离阈值：100米
// 性能目标：1000点 < 100ms (P99)

package algorithm

import (
	"math"
	"sort"
	"time"
)

// GPSPoint GPS轨迹点
type GPSPoint struct {
	Timestamp int64   // Unix timestamp (ms)
	Lat       float64 // 纬度
	Lng       float64 // 经度
	Accuracy  float32 // GPS精度（米）
}

// StayPoint 停留点
type StayPoint struct {
	ID          string  // 唯一ID
	CenterLat   float64 // 中心纬度
	CenterLng   float64 // 中心经度
	ArriveTime  int64   // 到达时间 (ms)
	LeaveTime   int64   // 离开时间 (ms)
	Duration    int32   // 停留时长（秒）
	ClusterSize int32   // 聚类点数
}

// StayPointDetector 停留点检测器
type StayPointDetector struct {
	Epsilon     float64 // 聚类半径（米）
	MinDuration int64   // 最短停留时间（毫秒）
	MinPoints   int     // 最少点数
}

// NewStayPointDetector 创建检测器
func NewStayPointDetector() *StayPointDetector {
	return &StayPointDetector{
		Epsilon:     100.0,              // 100米
		MinDuration: 15 * 60 * 1000,     // 15分钟 = 900000毫秒
		MinPoints:   5,                  // 最少5个点
	}
}

// Detect 检测停留点
// 性能：O(n log n)，n为轨迹点数
func (d *StayPointDetector) Detect(points []GPSPoint) []StayPoint {
	if len(points) < d.MinPoints {
		return nil
	}

	// 按时间排序
	sorted := make([]GPSPoint, len(points))
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp < sorted[j].Timestamp
	})

	var stays []StayPoint
	i := 0
	n := len(sorted)

	for i < n {
		// 寻找潜在停留区域
		cluster, nextIdx := d.findCluster(sorted, i)
		
		if len(cluster) >= d.MinPoints {
			// 检查时间阈值
			duration := cluster[len(cluster)-1].Timestamp - cluster[0].Timestamp
			if duration >= d.MinDuration {
				stay := d.createStayPoint(cluster)
				stays = append(stays, stay)
				i = nextIdx
				continue
			}
		}
		
		i++
	}

	// 合并相邻停留点
	stays = d.mergeNearbyStays(stays)

	return stays
}

// findCluster 从起始位置查找一个簇
func (d *StayPointDetector) findCluster(points []GPSPoint, startIdx int) ([]GPSPoint, int) {
	center := points[startIdx]
	cluster := []GPSPoint{center}
	
	i := startIdx + 1
	for i < len(points) {
		dist := haversineDistance(center.Lat, center.Lng, points[i].Lat, points[i].Lng)
		if dist > d.Epsilon {
			break
		}
		cluster = append(cluster, points[i])
		i++
	}
	
	return cluster, i
}

// createStayPoint 从簇创建停留点
func (d *StayPointDetector) createStayPoint(cluster []GPSPoint) StayPoint {
	centerLat, centerLng := d.computeWeightedCenter(cluster)
	
	return StayPoint{
		ID:          generateStayPointID(cluster[0].Timestamp),
		CenterLat:   centerLat,
		CenterLng:   centerLng,
		ArriveTime:  cluster[0].Timestamp,
		LeaveTime:   cluster[len(cluster)-1].Timestamp,
		Duration:    int32((cluster[len(cluster)-1].Timestamp - cluster[0].Timestamp) / 1000),
		ClusterSize: int32(len(cluster)),
	}
}

// computeWeightedCenter 计算加权中心（权重 = 1/accuracy）
func (d *StayPointDetector) computeWeightedCenter(points []GPSPoint) (float64, float64) {
	var totalWeight float64
	var weightedLat, weightedLng float64

	for _, p := range points {
		weight := 1.0
		if p.Accuracy > 0 {
			weight = 1.0 / float64(p.Accuracy)
		}
		weightedLat += p.Lat * weight
		weightedLng += p.Lng * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		// 等权重平均
		for _, p := range points {
			weightedLat += p.Lat
			weightedLng += p.Lng
		}
		return weightedLat / float64(len(points)), weightedLng / float64(len(points))
	}

	return weightedLat / totalWeight, weightedLng / totalWeight
}

// mergeNearbyStays 合并距离相近的停留点（<50米）
func (d *StayPointDetector) mergeNearbyStays(stays []StayPoint) []StayPoint {
	if len(stays) <= 1 {
		return stays
	}

	const mergeThreshold = 50.0 // 50米
	
	var merged []StayPoint
	used := make(map[int]bool)

	for i := 0; i < len(stays); i++ {
		if used[i] {
			continue
		}

		current := stays[i]
		mergeList := []StayPoint{current}

		for j := i + 1; j < len(stays); j++ {
			if used[j] {
				continue
			}

			dist := haversineDistance(
				current.CenterLat, current.CenterLng,
				stays[j].CenterLat, stays[j].CenterLng,
			)

			if dist < mergeThreshold {
				mergeList = append(mergeList, stays[j])
				used[j] = true
			}
		}

		if len(mergeList) > 1 {
			merged = append(merged, d.mergeStayPoints(mergeList))
		} else {
			merged = append(merged, current)
		}
	}

	return merged
}

// mergeStayPoints 合并多个停留点
func (d *StayPointDetector) mergeStayPoints(stays []StayPoint) StayPoint {
	var totalLat, totalLng float64
	var minArrive, maxLeave int64
	var totalDuration int32
	var totalSize int32

	minArrive = stays[0].ArriveTime
	maxLeave = stays[0].LeaveTime

	for _, s := range stays {
		totalLat += s.CenterLat
		totalLng += s.CenterLng
		
		if s.ArriveTime < minArrive {
			minArrive = s.ArriveTime
		}
		if s.LeaveTime > maxLeave {
			maxLeave = s.LeaveTime
		}
		totalDuration += s.Duration
		totalSize += s.ClusterSize
	}

	n := float64(len(stays))
	return StayPoint{
		ID:          stays[0].ID,
		CenterLat:   totalLat / n,
		CenterLng:   totalLng / n,
		ArriveTime:  minArrive,
		LeaveTime:   maxLeave,
		Duration:    totalDuration,
		ClusterSize: totalSize,
	}
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

// generateStayPointID 生成停留点ID
func generateStayPointID(timestamp int64) string {
	return "sp_" + time.UnixMilli(timestamp).Format("20060102_150405")
}

// BenchmarkDetect 性能基准测试
func BenchmarkDetect(points []GPSPoint) (time.Duration, []StayPoint) {
	start := time.Now()
	detector := NewStayPointDetector()
	stays := detector.Detect(points)
	elapsed := time.Since(start)
	return elapsed, stays
}
