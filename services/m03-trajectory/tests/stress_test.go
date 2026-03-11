// stress_test.go - 压力测试

package main

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/squirrel-awake/m03-trajectory/internal/algorithm"
)

// TestStress_100ConcurrentUsers 100并发用户测试
func TestStress_100ConcurrentUsers(t *testing.T) {
	log.Println("🚀 开始100并发用户压力测试")

	concurrentUsers := 100
	requestsPerUser := 1000

	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64

	start := time.Now()

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			for j := 0; j < requestsPerUser; j++ {
				if err := simulateUserRequest(fmt.Sprintf("user_%d", userID)); err != nil {
					errorCount++
				} else {
					successCount++
				}
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(start)
	totalRequests := concurrentUsers * requestsPerUser

	t.Logf("总请求: %d, 成功: %d, 失败: %d", totalRequests, successCount, errorCount)
	t.Logf("成功率: %.2f%%, QPS: %.2f",
		float64(successCount)/float64(totalRequests)*100,
		float64(totalRequests)/totalTime.Seconds())
}

// TestMemoryLeak 内存泄漏检查
func TestMemoryLeak(t *testing.T) {
	log.Println("🧠 内存泄漏检查")

	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	for i := 0; i < 10000; i++ {
		simulateHeavyOperation()
	}

	runtime.GC()
	time.Sleep(time.Second)

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	heapGrowth := int64(m2.HeapAlloc) - int64(m1.HeapAlloc)
	t.Logf("堆内存增长: %d KB", heapGrowth/1024)

	if heapGrowth > 100*1024*1024 {
		t.Errorf("可能内存泄漏: %d MB", heapGrowth/1024/1024)
	}
}

func simulateUserRequest(userID string) error {
	points := generateRandomPoints(10)
	detector := algorithm.NewStayPointDetector()
	_ = detector.Detect(points)
	return nil
}

func simulateHeavyOperation() {
	points := generateRandomPoints(100)
	detector := algorithm.NewStayPointDetector()
	_ = detector.Detect(points)
}

func generateRandomPoints(n int) []algorithm.GPSPoint {
	points := make([]algorithm.GPSPoint, n)
	now := time.Now().UnixMilli()
	for i := 0; i < n; i++ {
		points[i] = algorithm.GPSPoint{
			Timestamp: now - int64((n-i)*1000),
			Lat:       39.9 + float64(i)*0.001,
			Lng:       116.4 + float64(i)*0.001,
			Accuracy:  10,
		}
	}
	return points
}
