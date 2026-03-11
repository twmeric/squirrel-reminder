// worker_pool_test.go - WorkerPool单元测试

package optimization

import (
	"testing"
	"time"

	"github.com/squirrel-awake/m03-trajectory/internal/algorithm"
)

// TestWorkerPoolBasic 基本功能测试
func TestWorkerPoolBasic(t *testing.T) {
	wp := NewWorkerPool(4)
	wp.Start()
	defer wp.Stop()
	
	// 提交单个任务
	resultCh := make(chan Result, 1)
	job := Job{
		UserID:   "test_user_1",
		Points:   generateTestPoints(10),
		ResultCh: resultCh,
	}
	
	if !wp.Submit(job) {
		t.Fatal("Failed to submit job")
	}
	
	// 等待结果
	select {
	case result := <-resultCh:
		if result.UserID != "test_user_1" {
			t.Errorf("Expected user test_user_1, got %s", result.UserID)
		}
		if result.Err != nil {
			t.Errorf("Unexpected error: %v", result.Err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for result")
	}
}

// TestWorkerPoolBatch 批量处理测试
func TestWorkerPoolBatch(t *testing.T) {
	wp := NewWorkerPool(4)
	wp.Start()
	defer wp.Stop()
	
	// 准备批量数据
	users := []UserData{
		{ID: "user_1", Points: generateTestPoints(10)},
		{ID: "user_2", Points: generateTestPoints(10)},
		{ID: "user_3", Points: generateTestPoints(10)},
		{ID: "user_4", Points: generateTestPoints(10)},
	}
	
	// 批量处理
	results := wp.BatchProcess(users)
	
	// 验证结果
	if len(results) != len(users) {
		t.Errorf("Expected %d results, got %d", len(users), len(results))
	}
	
	for i, result := range results {
		if result.UserID != users[i].ID {
			t.Errorf("Result %d: expected user %s, got %s", i, users[i].ID, result.UserID)
		}
	}
}

// TestWorkerPoolConcurrency 并发测试
func TestWorkerPoolConcurrency(t *testing.T) {
	wp := NewWorkerPool(8)
	wp.Start()
	defer wp.Stop()
	
	numJobs := 100
	resultCh := make(chan Result, numJobs)
	
	// 并发提交任务
	for i := 0; i < numJobs; i++ {
		go func(id int) {
			job := Job{
				UserID:   string(rune(id)),
				Points:   generateTestPoints(5),
				ResultCh: resultCh,
			}
			wp.Submit(job)
		}(i)
	}
	
	// 收集结果
	collected := 0
	timeout := time.After(10 * time.Second)
	
	for collected < numJobs {
		select {
		case <-resultCh:
			collected++
		case <-timeout:
			t.Fatalf("Timeout: only collected %d/%d results", collected, numJobs)
		}
	}
	
	t.Logf("Successfully processed %d concurrent jobs", collected)
}

// TestWorkerPoolGracefulShutdown 优雅关闭测试
func TestWorkerPoolGracefulShutdown(t *testing.T) {
	wp := NewWorkerPool(4)
	wp.Start()
	
	// 提交一些任务
	for i := 0; i < 10; i++ {
		resultCh := make(chan Result, 1)
		job := Job{
			UserID:   string(rune(i)),
			Points:   generateTestPoints(5),
			ResultCh: resultCh,
		}
		wp.Submit(job)
	}
	
	// 优雅关闭
	done := make(chan bool)
	go func() {
		wp.Stop()
		done <- true
	}()
	
	select {
	case <-done:
		t.Log("Graceful shutdown completed")
	case <-time.After(5 * time.Second):
		t.Fatal("Shutdown timeout")
	}
}

// TestWorkerPoolStats 统计信息测试
func TestWorkerPoolStats(t *testing.T) {
	wp := NewWorkerPool(4)
	
	stats := wp.GetStats()
	
	if stats["num_workers"] != 4 {
		t.Errorf("Expected 4 workers, got %v", stats["num_workers"])
	}
	
	if stats["queue_cap"] != 16 { // 4*4
		t.Errorf("Expected queue cap 16, got %v", stats["queue_cap"])
	}
}

// BenchmarkWorkerPool WorkerPool基准测试
func BenchmarkWorkerPool(b *testing.B) {
	wp := NewWorkerPool(4)
	wp.Start()
	defer wp.Stop()
	
	users := make([]UserData, 100)
	for i := 0; i < 100; i++ {
		users[i] = UserData{
			ID:     string(rune(i)),
			Points: generateTestPoints(10),
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wp.BatchProcess(users)
	}
}

// BenchmarkSerial 串行处理基准测试
func BenchmarkSerial(b *testing.B) {
	detector := algorithm.NewStayPointDetector()
	smoother := algorithm.NewSpeedSmoother()
	
	users := make([]UserData, 100)
	for i := 0; i < 100; i++ {
		users[i] = UserData{
			ID:     string(rune(i)),
			Points: generateTestPoints(10),
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, user := range users {
			_ = detector.Detect(user.Points)
			_ = smoother.CalculateSpeeds(user.Points)
		}
	}
}

// 辅助函数
func generateTestPoints(n int) []algorithm.GPSPoint {
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
