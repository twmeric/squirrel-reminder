package algorithm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStayPointDetector_Detect(t *testing.T) {
	spd := NewStayPointDetector()

	tests := []struct {
		name     string
		points   []GPSPoint
		expected int
	}{
		{
			name:     "empty input",
			points:   []GPSPoint{},
			expected: 0,
		},
		{
			name:     "single point",
			points:   []GPSPoint{{Lat: 22.5, Lng: 113.9, Ts: time.Now()}},
			expected: 0,
		},
		{
			name: "clear staypoint - cluster at home",
			points: generateClusterPoints(
				22.5431, 113.9589,  // 中心
				20,                 // 20个点
				time.Hour,          // 1小时持续时间
			),
			expected: 1,
		},
		{
			name: "two staypoints with movement",
			points: concatPoints(
				generateClusterPoints(22.5431, 113.9589, 20, time.Hour),
				generateClusterPoints(22.5321, 113.9512, 20, time.Hour*2),
			),
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := spd.Detect(tt.points)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestStayPoint_Center(t *testing.T) {
	sp := &StayPoint{
		Points: []GPSPoint{
			{Lat: 22.0, Lng: 113.0},
			{Lat: 22.1, Lng: 113.1},
			{Lat: 22.2, Lng: 113.2},
		},
	}

	center := sp.Center()
	assert.InDelta(t, 22.1, center.Lat, 0.01)
	assert.InDelta(t, 113.1, center.Lng, 0.01)
}

func TestStayPoint_Duration(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Hour * 2)

	sp := &StayPoint{
		Points: []GPSPoint{
			{Lat: 22.0, Lng: 113.0, Ts: start},
			{Lat: 22.0, Lng: 113.0, Ts: end},
		},
	}

	duration := sp.Duration()
	assert.InDelta(t, time.Hour*2, duration, time.Second)
}

func BenchmarkStayPointDetector_Detect(b *testing.B) {
	spd := NewStayPointDetector()
	points := generateClusterPoints(22.5431, 113.9589, 1000, time.Hour*2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spd.Detect(points)
	}
}

// 辅助函数
func generateClusterPoints(centerLat, centerLng float64, count int, duration time.Duration) []GPSPoint {
	points := make([]GPSPoint, count)
	startTime := time.Now()
	interval := duration / time.Duration(count)

	for i := 0; i < count; i++ {
		// 在中心点周围生成一些噪声
		lat := centerLat + (float64(i%10)-5)*0.0001
		lng := centerLng + (float64(i%7)-3)*0.0001
		points[i] = GPSPoint{
			Lat: lat,
			Lng: lng,
			Ts:  startTime.Add(interval * time.Duration(i)),
		}
	}
	return points
}

func concatPoints(slices ...[]GPSPoint) []GPSPoint {
	var result []GPSPoint
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}
