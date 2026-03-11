// benchmark.go - 性能基准测试与优化
// 目标：P99 < 100ms

package performance

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/squirrel-awake/m03-trajectory/internal/algorithm"
	"github.com/squirrel-awake/m03-trajectory/internal/storage"
)

// BenchmarkResult 基准测试结果
type BenchmarkResult struct {
	Operation   string
	Count       int
	AvgLatency  time.Duration
	P50Latency  time.Duration
	P95Latency  time.Duration
	P99Latency  time.Duration
	MaxLatency  time.Duration
	Throughput  float64 // ops/sec
	Status      string  // PASS / FAIL
}

// RunAllBenchmarks 运行所有基准测试
func RunAllBenchmarks() []BenchmarkResult {
	var results []BenchmarkResult

	results = append(results, BenchmarkStayPointDetection())
	results = append(results, BenchmarkSpeedSmoothing())
	results = append(results, BenchmarkGPSFilter())

	return results
}

// BenchmarkStayPointDetection 停留点检测性能测试
func BenchmarkStayPointDetection() BenchmarkResult {
	points := generateTestPoints(1000)
	detector := algorithm.NewStayPointDetector()

	// 预热
	for i := 0; i < 10; i++ {
		detector.Detect(points)
	}

	// 正式测试
	var times []time.Duration
	for i := 0; i < 100; i++ {
		start := time.Now()
		detector.Detect(points)
		times = append(times, time.Since(start))
	}

	result := calculateMetrics(times, "StayPointDetection")
	result.Status = checkSLA(result.P99Latency, 100*time.Millisecond)
	return result
}

// BenchmarkSpeedSmoothing 速度平滑性能测试
func BenchmarkSpeedSmoothing() BenchmarkResult {
	points := generateTestPoints(1000)
	smoother := algorithm.NewSpeedSmoother()

	var times []time.Duration
	for i := 0; i < 100; i++ {
		start := time.Now()
		smoother.CalculateSpeeds(points)
		times = append(times, time.Since(start))
	}

	result := calculateMetrics(times, "SpeedSmoothing")
	result.Status = checkSLA(result.P99Latency, 10*time.Millisecond)
	return result
}

// BenchmarkGPSFilter GPS漂移过滤性能测试
func BenchmarkGPSFilter() BenchmarkResult {
	points := generateTestPoints(1000)
	filter := algorithm.NewGPSDriftFilter()

	var times []time.Duration
	for i := 0; i < 100; i++ {
		start := time.Now()
		filter.Filter(points)
		times = append(times, time.Since(start))
	}

	result := calculateMetrics(times, "GPSFilter")
	result.Status = checkSLA(result.P99Latency, 10*time.Millisecond)
	return result
}

// calculateMetrics 计算性能指标
func calculateMetrics(times []time.Duration, operation string) BenchmarkResult {
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})

	n := len(times)
	var sum time.Duration
	for _, t := range times {
		sum += t
	}

	avg := sum / time.Duration(n)
	p50 := times[n*50/100]
	p95 := times[n*95/100]
	p99 := times[n*99/100]
	max := times[n-1]

	throughput := float64(n) / sum.Seconds()

	return BenchmarkResult{
		Operation:  operation,
		Count:      n,
		AvgLatency: avg,
		P50Latency: p50,
		P95Latency: p95,
		P99Latency: p99,
		MaxLatency: max,
		Throughput: throughput,
	}
}

// checkSLA 检查是否满足SLA
func checkSLA(p99 time.Duration, target time.Duration) string {
	if p99 <= target {
		return "PASS"
	}
	return "FAIL"
}

// PrintResults 打印结果
func PrintResults(results []BenchmarkResult) {
	fmt.Println("\n📊 Performance Benchmark Results")
	fmt.Println("================================================================")
	fmt.Printf("%-20s %8s %10s %10s %10s %8s\n",
		"Operation", "Count", "Avg(ms)", "P99(ms)", "Target", "Status")
	fmt.Println("----------------------------------------------------------------")

	for _, r := range results {
		target := "-"
		switch r.Operation {
		case "StayPointDetection":
			target = "100"
		case "SpeedSmoothing", "GPSFilter":
			target = "10"
		}

		fmt.Printf("%-20s %8d %10.2f %10.2f %10s %8s\n",
			r.Operation,
			r.Count,
			float64(r.AvgLatency)/float64(time.Millisecond),
			float64(r.P99Latency)/float64(time.Millisecond),
			target,
			r.Status,
		)
	}
	fmt.Println("================================================================")

	allPass := true
	for _, r := range results {
		if r.Status != "PASS" {
			allPass = false
			break
		}
	}

	if allPass {
		fmt.Println("✅ All benchmarks PASSED (P99 < SLA)")
	} else {
		fmt.Println("❌ Some benchmarks FAILED (need optimization)")
	}
}

// 辅助函数

func generateTestPoints(n int) []algorithm.GPSPoint {
	points := make([]algorithm.GPSPoint, n)
	startTime := time.Now().Add(-time.Hour).UnixMilli()
	interval := int64(3600*1000) / int64(n)

	lat, lng := 39.9, 116.4

	for i := 0; i < n; i++ {
		lat += (float64(i%10) - 5) * 0.0001
		lng += (float64(i%10) - 5) * 0.0001

		points[i] = algorithm.GPSPoint{
			Timestamp: startTime + int64(i)*interval,
			Lat:       lat,
			Lng:       lng,
			Accuracy:  10,
		}
	}

	return points
}

func generateTestRecords(n int) []storage.GPSLogRecord {
	records := make([]storage.GPSLogRecord, n)
	now := time.Now()

	for i := 0; i < n; i++ {
		records[i] = storage.GPSLogRecord{
			UserID:    "test_user",
			Timestamp: now.Add(-time.Duration(i) * time.Second).UnixMilli(),
			Lat:       39.9 + float64(i)*0.0001,
			Lng:       116.4 + float64(i)*0.0001,
			Accuracy:  10,
			Speed:     5,
			Source:    "gps",
		}
	}

	return records
}
