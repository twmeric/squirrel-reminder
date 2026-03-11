// speed_optimizer.go - GetSpeed性能优化
// 目标: P99 11.5ms → 8ms

package optimization

import (
	"sync"
	"time"

	"github.com/squirrel-awake/m03-trajectory/internal/algorithm"
)

// SpeedOptimizer 速度计算优化器
type SpeedOptimizer struct {
	cache      map[string]*speedCacheEntry
	cacheMutex sync.RWMutex
	cacheTTL   time.Duration
	pointPool  sync.Pool
}

type speedCacheEntry struct {
	speed     float64
	timestamp time.Time
	pointHash uint64
}

// NewSpeedOptimizer 创建优化器
func NewSpeedOptimizer() *SpeedOptimizer {
	return &SpeedOptimizer{
		cache:    make(map[string]*speedCacheEntry),
		cacheTTL: 100 * time.Millisecond,
		pointPool: sync.Pool{
			New: func() interface{} {
				return make([]algorithm.GPSPoint, 0, 128)
			},
		},
	}
}

// GetSpeedOptimized 优化版速度获取
func (o *SpeedOptimizer) GetSpeedOptimized(userID string, points []algorithm.GPSPoint, smoother *algorithm.SpeedSmoother) float64 {
	if len(points) == 0 {
		return 0
	}

	pointHash := o.hashPoints(points)
	if cached := o.getFromCache(userID, pointHash); cached >= 0 {
		return cached
	}

	pooledPoints := o.pointPool.Get().([]algorithm.GPSPoint)
	defer o.pointPool.Put(pooledPoints[:0])

	pooledPoints = append(pooledPoints[:0], points...)
	speed := smoother.GetCurrentSpeed(pooledPoints)
	o.setCache(userID, pointHash, speed)

	return speed
}

func (o *SpeedOptimizer) getFromCache(userID string, pointHash uint64) float64 {
	o.cacheMutex.RLock()
	defer o.cacheMutex.RUnlock()

	entry, exists := o.cache[userID]
	if !exists || time.Since(entry.timestamp) > o.cacheTTL || entry.pointHash != pointHash {
		return -1
	}
	return entry.speed
}

func (o *SpeedOptimizer) setCache(userID string, pointHash uint64, speed float64) {
	o.cacheMutex.Lock()
	defer o.cacheMutex.Unlock()
	o.cache[userID] = &speedCacheEntry{speed: speed, timestamp: time.Now(), pointHash: pointHash}
}

func (o *SpeedOptimizer) hashPoints(points []algorithm.GPSPoint) uint64 {
	if len(points) == 0 {
		return 0
	}
	var hash uint64
	start := len(points) - 3
	if start < 0 {
		start = 0
	}
	for i, p := range points[start:] {
		hash += uint64(int64(p.Timestamp)^int64(p.Lat*1000000)^int64(p.Lng*1000000)) << uint(i*8)
	}
	return hash
}
