package algorithm

import (
	"encoding/json"
	"math"
	"os"
	"sort"
)

// Station 地铁站点
type Station struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	LineName    string  `json:"line_name"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	IsTransfer  bool    `json:"is_transfer"`
	TransferTo  []string `json:"transfer_to"`
	StationsToNext int  `json:"stations_to_next"` // 到下一站的站数
}

type StationMatcher struct {
	stations []Station
	kdtree   *KDTree
}

type KDTree struct {
	root *KDNode
}

type KDNode struct {
	point    Station
	left     *KDNode
	right    *KDNode
	axis     int
}

func NewStationMatcher() *StationMatcher {
	return &StationMatcher{
		stations: loadStations(),
	}
}

func (sm *StationMatcher) BuildKDTree() {
	sm.kdtree = BuildKDTree(sm.stations, 0)
}

func (sm *StationMatcher) FindNearestStation(lat, lng float64) (*Station, float64) {
	if sm.kdtree == nil {
		sm.BuildKDTree()
	}

	// KD-tree最近邻搜索
	best := &Best{dist: math.MaxFloat64}
	sm.kdtree.nearest(sm.kdtree.root, lat, lng, best)
	
	if best.station != nil {
		return best.station, math.Sqrt(best.dist)
	}
	return nil, -1
}

func (sm *StationMatcher) FindStationsWithinRadius(lat, lng float64, radius float64) []*Station {
	var result []*Station
	for i := range sm.stations {
		dist := haversineDistance(lat, lng, sm.stations[i].Lat, sm.stations[i].Lng)
		if dist <= radius {
			result = append(result, &sm.stations[i])
		}
	}
	
	// 按距离排序
	sort.Slice(result, func(i, j int) bool {
		di := haversineDistance(lat, lng, result[i].Lat, result[i].Lng)
		dj := haversineDistance(lat, lng, result[j].Lat, result[j].Lng)
		return di < dj
	})
	
	return result
}

// KDTree实现
func BuildKDTree(stations []Station, depth int) *KDTree {
	if len(stations) == 0 {
		return &KDTree{}
	}

	axis := depth % 2 // 0 for latitude, 1 for longitude
	
	// 按轴排序
	sort.Slice(stations, func(i, j int) bool {
		if axis == 0 {
			return stations[i].Lat < stations[j].Lat
		}
		return stations[i].Lng < stations[j].Lng
	})

	mid := len(stations) / 2

	return &KDTree{
		root: &KDNode{
			point: stations[mid],
			left:  buildKDNode(stations[:mid], depth+1),
			right: buildKDNode(stations[mid+1:], depth+1),
			axis:  axis,
		},
	}
}

func buildKDNode(stations []Station, depth int) *KDNode {
	if len(stations) == 0 {
		return nil
	}

	axis := depth % 2
	sort.Slice(stations, func(i, j int) bool {
		if axis == 0 {
			return stations[i].Lat < stations[j].Lat
		}
		return stations[i].Lng < stations[j].Lng
	})

	mid := len(stations) / 2

	return &KDNode{
		point: stations[mid],
		left:  buildKDNode(stations[:mid], depth+1),
		right: buildKDNode(stations[mid+1:], depth+1),
		axis:  axis,
	}
}

type Best struct {
	station *Station
	dist    float64
}

func (kdt *KDTree) nearest(node *KDNode, lat, lng float64, best *Best) {
	if node == nil {
		return
	}

	// 计算当前点距离
	d := squaredDist(lat, lng, node.point.Lat, node.point.Lng)
	if d < best.dist {
		best.dist = d
		station := node.point
		best.station = &station
	}

	// 选择搜索分支
	var near, far *KDNode
	var diff float64
	
	if node.axis == 0 {
		diff = lat - node.point.Lat
	} else {
		diff = lng - node.point.Lng
	}

	if diff < 0 {
		near, far = node.left, node.right
	} else {
		near, far = node.right, node.left
	}

	// 先搜索近分支
	kdt.nearest(near, lat, lng, best)

	// 如果超平面距离小于当前最优，搜索远分支
	if diff*diff < best.dist {
		kdt.nearest(far, lat, lng, best)
	}
}

func squaredDist(lat1, lng1, lat2, lng2 float64) float64 {
	// 简化版，实际应该用haversine
	dlat := lat1 - lat2
	dlng := lng1 - lng2
	return dlat*dlat + dlng*dlng
}

func loadStations() []Station {
	// 尝试加载文件
	data, err := os.ReadFile("/data/metro_stations.json")
	if err != nil {
		// 返回默认数据
		return defaultStations()
	}

	var stations []Station
	if err := json.Unmarshal(data, &stations); err != nil {
		return defaultStations()
	}
	return stations
}

func defaultStations() []Station {
	return []Station{
		{ID: "L1_S01", Name: "科技园", LineName: "1号线", Lat: 22.5431, Lng: 113.9589, StationsToNext: 1},
		{ID: "L1_S02", Name: "深大", LineName: "1号线", Lat: 22.5321, Lng: 113.9512, IsTransfer: true, TransferTo: []string{"L2_S01"}, StationsToNext: 2},
		{ID: "L1_S03", Name: "世界之窗", LineName: "1号线", Lat: 22.5342, Lng: 113.9789, StationsToNext: 1},
		{ID: "L2_S01", Name: "深大北", LineName: "2号线", Lat: 22.5330, Lng: 113.9500, IsTransfer: true, TransferTo: []string{"L1_S02"}, StationsToNext: 1},
		{ID: "L2_S02", Name: "科苑", LineName: "2号线", Lat: 22.5280, Lng: 113.9450, StationsToNext: 2},
	}
}
