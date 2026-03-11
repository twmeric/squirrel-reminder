// staypoint_test.go - 停留点检测算法测试

package algorithm

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// TestDetectStayPoints_Basic 基础功能测试
func TestDetectStayPoints_Basic(t *testing.T) {
	detector := NewStayPointDetector()
	
	// 构造轨迹：50个点在100米内停留20分钟
	points := generateStayTrajectory(39.9042, 116.4074, 50, 20*60, 30)
	
	stays := detector.Detect(points)
	
	if len(stays) != 1 {
		t.Errorf("Expected 1 stay point, got %d", len(stays))
	}
	
	if stays[0].Duration < 900 {
		t.Errorf("Expected duration >= 900s, got %d", stays[0].Duration)
	}
	
	t.Logf("✅ Detected stay: %v", stays[0])
}

// TestDetectStayPoints_DurationThreshold 时间阈值测试
func TestDetectStayPoints_DurationThreshold(t *testing.T) {
	detector := NewStayPointDetector()
	
	// 10分钟（不足15分钟阈值）
	points := generateStayTrajectory(39.9042, 116.4074, 20, 10*60, 30)
	
	stays := detector.Detect(points)
	
	if len(stays) != 0 {
		t.Errorf("Expected 0 stay points (duration < 15min), got %d", len(stays))
	}
	
	t.Log("✅ Duration threshold check passed")
}

// TestDetectStayPoints_MinPoints 最少点数测试
func TestDetectStayPoints_MinPoints(t *testing.T) {
	detector := NewStayPointDetector()
	
	// 3个点（不足5个）
	points := generateStayTrajectory(39.9042, 116.4074, 3, 20*60, 30)
	
	stays := detector.Detect(points)
	
	if len(stays) != 0 {
		t.Errorf("Expected 0 stay points (points < 5), got %d", len(stays))
	}
	
	t.Log("✅ Min points check passed")
}

// TestDetectStayPoints_Multiple 多个停留点测试
func TestDetectStayPoints_Multiple(t *testing.T) {
	detector := NewStayPointDetector()
	
	var allPoints []GPSPoint
	
	// 第一个停留点（国贸）
	points1 := generateStayTrajectory(39.9240, 116.4690, 50, 20*60, 30)
	allPoints = append(allPoints, points1...)
	
	// 移动段
	movePoints := generateMoveTrajectory(39.9240, 116.4690, 39.9220, 116.4400, 10)
	allPoints = append(allPoints, movePoints...)
	
	// 第二个停留点（建国门）
	points2 := generateStayTrajectory(39.9220, 116.4400, 50, 20*60, 30)
	allPoints = append(allPoints, points2...)
	
	stays := detector.Detect(allPoints)
	
	if len(stays) != 2 {
		t.Errorf("Expected 2 stay points, got %d", len(stays))
	}
	
	t.Logf("✅ Detected %d stay points", len(stays))
}

// BenchmarkDetect_1000Points 性能基准测试 - 1000点
func BenchmarkDetect_1000Points(b *testing.B) {
	points := generateRandomTrajectory(1000)
	detector := NewStayPointDetector()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.Detect(points)
	}
}

// BenchmarkDetect_10000Points 性能基准测试 - 10000点
func BenchmarkDetect_10000Points(b *testing.B) {
	points := generateRandomTrajectory(10000)
	detector := NewStayPointDetector()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.Detect(points)
	}
}

// TestPerformance_1000Points P99 < 100ms 验证
func TestPerformance_1000Points(t *testing.T) {
	points := generateRandomTrajectory(1000)
	detector := NewStayPointDetector()
	
	// 运行100次取P99
	var times []time.Duration
	for i := 0; i < 100; i++ {
		start := time.Now()
		detector.Detect(points)
		times = append(times, time.Since(start))
	}
	
	// 计算P99
	p99 := calculateP99(times)
	
	if p99 > 100*time.Millisecond {
		t.Errorf("P99 latency %v exceeds 100ms", p99)
	}
	
	t.Logf("✅ P99 latency: %v (target: < 100ms)", p99)
}

// 辅助函数

func generateStayTrajectory(centerLat, centerLng float64, numPoints int, durationSec int, radiusMeters float64) []GPSPoint {
	points := make([]GPSPoint, numPoints)
	startTime := time.Now().Add(-time.Duration(durationSec) * time.Second).UnixMilli()
	interval := int64(durationSec*1000) / int64(numPoints)
	
	for i := 0; i < numPoints; i++ {
		// 在半径内随机偏移
		offsetLat := (rand.Float64() - 0.5) * 2 * radiusMeters / 111000.0
		offsetLng := (rand.Float64() - 0.5) * 2 * radiusMeters / (111000.0 * cos(centerLat))
		
		points[i] = GPSPoint{
			Timestamp: startTime + int64(i)*interval,
			Lat:       centerLat + offsetLat,
			Lng:       centerLng + offsetLng,
			Accuracy:  10 + float32(rand.Intn(20)),
		}
	}
	
	return points
}

func generateMoveTrajectory(fromLat, fromLng, toLat, toLng float64, numPoints int) []GPSPoint {
	points := make([]GPSPoint, numPoints)
	startTime := time.Now().UnixMilli()
	interval := int64(5000) // 5秒间隔
	
	for i := 0; i < numPoints; i++ {
		ratio := float64(i) / float64(numPoints-1)
		points[i] = GPSPoint{
			Timestamp: startTime + int64(i)*interval,
			Lat:       fromLat + (toLat-fromLat)*ratio,
			Lng:       fromLng + (toLng-fromLng)*ratio,
			Accuracy:  10,
		}
	}
	
	return points
}

func generateRandomTrajectory(numPoints int) []GPSPoint {
	points := make([]GPSPoint, numPoints)
	startTime := time.Now().Add(-time.Hour).UnixMilli()
	interval := int64(3600*1000) / int64(numPoints)
	
	lat, lng := 39.9, 116.4
	
	for i := 0; i < numPoints; i++ {
		// 随机移动
		lat += (rand.Float64() - 0.5) * 0.001
		lng += (rand.Float64() - 0.5) * 0.001
		
		points[i] = GPSPoint{
			Timestamp: startTime + int64(i)*interval,
			Lat:       lat,
			Lng:       lng,
			Accuracy:  10 + float32(rand.Intn(20)),
		}
	}
	
	return points
}

func calculateP99(times []time.Duration) time.Duration {
	// 简单排序取99分位
	sorted := make([]time.Duration, len(times))
	copy(sorted, times)
	
	// 冒泡排序（小数据量）
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	idx := int(float64(len(sorted)) * 0.99)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func cos(deg float64) float64 {
	// 简化版，实际应使用 math.Cos(math.Pi/180*deg)
	return 0.8 // 近似值，北京纬度
}

// ExampleDetect 使用示例
func ExampleStayPointDetector_Detect() {
	detector := NewStayPointDetector()
	
	// 模拟轨迹数据
	points := []GPSPoint{
		{Timestamp: 1, Lat: 39.9042, Lng: 116.4074, Accuracy: 10},
		{Timestamp: 2, Lat: 39.9043, Lng: 116.4075, Accuracy: 12},
		{Timestamp: 3, Lat: 39.9041, Lng: 116.4073, Accuracy: 8},
		// ... 更多点
	}
	
	stays := detector.Detect(points)
	
	for _, stay := range stays {
		fmt.Printf("Stay at (%.4f, %.4f) for %d seconds\n", 
			stay.CenterLat, stay.CenterLng, stay.Duration)
	}
}
