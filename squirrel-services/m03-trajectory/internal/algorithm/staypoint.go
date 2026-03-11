package algorithm

import (
	"math"
	"sort"
	"time"
)

// GPSPoint 原始GPS点
type GPSPoint struct {
	Lat float64
	Lng float64
	Ts  time.Time
}

// StayPoint 停留点
type StayPoint struct {
	Points []GPSPoint
}

// DBSCAN参数
const (
	DefaultEpsilon     = 100.0       // 100米
	DefaultMinPoints   = 5           // 最少5个点
	DefaultMinDuration = time.Minute * 15 // 至少15分钟
)

type StayPointDetector struct {
	Epsilon     float64       // 聚类半径(米)
	MinPoints   int           // 最小点数
	MinDuration time.Duration // 最小持续时间
}

func NewStayPointDetector() *StayPointDetector {
	return &StayPointDetector{
		Epsilon:     DefaultEpsilon,
		MinPoints:   DefaultMinPoints,
		MinDuration: DefaultMinDuration,
	}
}

func (spd *StayPointDetector) Detect(points []GPSPoint) []*StayPoint {
	if len(points) < spd.MinPoints {
		return nil
	}

	// 按时间排序
	sort.Slice(points, func(i, j int) bool {
		return points[i].Ts.Before(points[j].Ts)
	})

	visited := make([]bool, len(points))
	var stayPoints []*StayPoint

	for i := range points {
		if visited[i] {
			continue
		}

		// 找到所有邻居
		neighbors := spd.regionQuery(points, i)
		if len(neighbors) < spd.MinPoints {
			visited[i] = true
			continue
		}

		// 扩展聚类
		cluster := spd.expandCluster(points, neighbors, visited)
		
		// 检查持续时间
		if cluster.Duration() >= spd.MinDuration {
			stayPoints = append(stayPoints, cluster)
		}
	}

	return stayPoints
}

func (spd *StayPointDetector) regionQuery(points []GPSPoint, centerIdx int) []int {
	var neighbors []int
	center := points[centerIdx]

	for i, p := range points {
		if haversineDistance(center.Lat, center.Lng, p.Lat, p.Lng) <= spd.Epsilon {
			neighbors = append(neighbors, i)
		}
	}

	return neighbors
}

func (spd *StayPointDetector) expandCluster(points []GPSPoint, neighbors []int, visited []bool) *StayPoint {
	cluster := &StayPoint{}

	for i := 0; i < len(neighbors); i++ {
		idx := neighbors[i]
		if visited[idx] {
			continue
		}

		visited[idx] = true
		cluster.Points = append(cluster.Points, points[idx])

		// 递归查找更多邻居
		newNeighbors := spd.regionQuery(points, idx)
		if len(newNeighbors) >= spd.MinPoints {
			neighbors = append(neighbors, newNeighbors...)
		}
	}

	return cluster
}

// Center 计算停留点中心
func (sp *StayPoint) Center() GPSPoint {
	if len(sp.Points) == 0 {
		return GPSPoint{}
	}

	var sumLat, sumLng float64
	for _, p := range sp.Points {
		sumLat += p.Lat
		sumLng += p.Lng
	}

	return GPSPoint{
		Lat: sumLat / float64(len(sp.Points)),
		Lng: sumLng / float64(len(sp.Points)),
	}
}

func (sp *StayPoint) StartTime() time.Time {
	if len(sp.Points) == 0 {
		return time.Time{}
	}
	return sp.Points[0].Ts
}

func (sp *StayPoint) EndTime() time.Time {
	if len(sp.Points) == 0 {
		return time.Time{}
	}
	return sp.Points[len(sp.Points)-1].Ts
}

func (sp *StayPoint) Duration() time.Duration {
	return sp.EndTime().Sub(sp.StartTime())
}

func (sp *StayPoint) PointCount() int {
	return len(sp.Points)
}

// haversineDistance 计算两点间距离(米)
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000 // 地球半径(米)

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
