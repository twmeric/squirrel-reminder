// worker_pool.go - 算法并行化工作池
// v1.2.1 核心优化：多核CPU并行处理

package optimization

import (
	"context"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/squirrel-awake/m03-trajectory/internal/algorithm"
)

// Job 处理任务
type Job struct {
	UserID   string
	Points   []algorithm.GPSPoint
	ResultCh chan<- Result
}

// Result 处理结果
type Result struct {
	UserID    string
	Speed     float64
	StayPoints []algorithm.StayPoint
	Err       error
	Duration  time.Duration
}

// WorkerPool 工作池
type WorkerPool struct {
	numWorkers int
	jobQueue   chan Job
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	
	// 依赖
	stayDetector  *algorithm.StayPointDetector
	speedSmoother *algorithm.SpeedSmoother
}

// NewWorkerPool 创建工作池
func NewWorkerPool(numWorkers int) *WorkerPool {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WorkerPool{
		numWorkers:    numWorkers,
		jobQueue:      make(chan Job, numWorkers*4),
		ctx:           ctx,
		cancel:        cancel,
		stayDetector:  algorithm.NewStayPointDetector(),
		speedSmoother: algorithm.NewSpeedSmoother(),
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	log.Printf("[WorkerPool] Starting %d workers", wp.numWorkers)
	
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	log.Println("[WorkerPool] Stopping...")
	wp.cancel()
	wp.wg.Wait()
	log.Println("[WorkerPool] Stopped")
}

// Submit 提交任务
func (wp *WorkerPool) Submit(job Job) bool {
	select {
	case wp.jobQueue <- job:
		return true
	case <-time.After(5 * time.Second):
		log.Printf("[WorkerPool] Submit timeout for user %s", job.UserID)
		return false
	}
}

// worker 工作协程
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	
	log.Printf("[WorkerPool] Worker %d started", id)
	
	for {
		select {
		case job := <-wp.jobQueue:
			start := time.Now()
			result := wp.processJob(job)
			result.Duration = time.Since(start)
			
			// 发送结果
			select {
			case job.ResultCh <- result:
			case <-time.After(10 * time.Second):
				log.Printf("[WorkerPool] Result send timeout for user %s", job.UserID)
			}
			
		case <-wp.ctx.Done():
			log.Printf("[WorkerPool] Worker %d stopping", id)
			return
		}
	}
}

// processJob 处理单个任务
func (wp *WorkerPool) processJob(job Job) Result {
	result := Result{
		UserID: job.UserID,
	}
	
	// 1. 使用FastSmooth计算速度 (SIMD优化)
	if len(job.Points) >= 2 {
		// 提取速度值
		instantSpeeds := make([]float64, len(job.Points))
		instantSpeeds[0] = 0
		
		for i := 1; i < len(job.Points); i++ {
			dist := haversineDistance(
				job.Points[i-1].Lat, job.Points[i-1].Lng,
				job.Points[i].Lat, job.Points[i].Lng,
			)
			timeDiff := float64(job.Points[i].Timestamp-job.Points[i-1].Timestamp) / 1000.0
			if timeDiff > 0 {
				instantSpeeds[i] = (dist / timeDiff) * 3.6
			}
		}
		
		// 使用FastSmooth平滑
		smoothed := FastSmooth(instantSpeeds, 0.3)
		result.Speed = smoothed[len(smoothed)-1]
	}
	
	// 2. 检测停留点
	result.StayPoints = wp.stayDetector.Detect(job.Points)
	
	return result
}

// BatchProcess 批量处理
func (wp *WorkerPool) BatchProcess(users []UserData) []Result {
	results := make([]Result, len(users))
	resultCh := make(chan Result, len(users))
	
	// 提交所有任务
	for i, user := range users {
		job := Job{
			UserID:   user.ID,
			Points:   user.Points,
			ResultCh: resultCh,
		}
		
		if !wp.Submit(job) {
			// 提交失败，记录错误
			results[i] = Result{
				UserID: user.ID,
				Err:    ErrJobSubmitTimeout,
			}
		}
	}
	
	// 收集结果
	for i := 0; i < len(users); i++ {
		select {
		case result := <-resultCh:
			// 找到对应位置
			for j, user := range users {
				if user.ID == result.UserID {
					results[j] = result
					break
				}
			}
		case <-time.After(30 * time.Second):
			log.Println("[WorkerPool] BatchProcess timeout")
			break
		}
	}
	
	return results
}

// UserData 用户数据
type UserData struct {
	ID     string
	Points []algorithm.GPSPoint
}

var ErrJobSubmitTimeout = Error("job submit timeout")

type Error string

func (e Error) Error() string {
	return string(e)
}

// GetStats 获取工作池统计
func (wp *WorkerPool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"num_workers": wp.numWorkers,
		"queue_size":  len(wp.jobQueue),
		"queue_cap":   cap(wp.jobQueue),
	}
}

// haversineDistance 计算距离（简化版）
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000
	
	lat1Rad := lat1 * 3.14159 / 180
	lat2Rad := lat2 * 3.14159 / 180
	deltaLat := (lat2 - lat1) * 3.14159 / 180
	deltaLng := (lng2 - lng1) * 3.14159 / 180
	
	a := sin2(deltaLat/2) + cos(lat1Rad)*cos(lat2Rad)*sin2(deltaLng/2)
	c := 2 * atan2(sqrt(a), sqrt(1-a))
	
	return R * c
}

func sin2(x float64) float64 {
	s := sin(x)
	return s * s
}

func cos(x float64) float64 {
	return 1 - x*x/2 + x*x*x*x/24 // 泰勒展开近似
}

func sin(x float64) float64 {
	return x - x*x*x/6 + x*x*x*x*x/120 // 泰勒展开近似
}

func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func atan2(y, x float64) float64 {
	if x > 0 {
		return atan(y/x)
	}
	if x < 0 && y >= 0 {
		return atan(y/x) + 3.14159
	}
	if x < 0 && y < 0 {
		return atan(y/x) - 3.14159
	}
	if x == 0 && y > 0 {
		return 3.14159 / 2
	}
	if x == 0 && y < 0 {
		return -3.14159 / 2
	}
	return 0
}

func atan(x float64) float64 {
	return x - x*x*x/3 + x*x*x*x*x/5 // 泰勒展开近似
}
