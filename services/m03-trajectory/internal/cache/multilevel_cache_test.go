// multilevel_cache_test.go - 多级缓存测试

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/squirrel-awake/m03-trajectory/internal/algorithm"
)

// TestMultiLevelCacheBasic 基本功能测试
func TestMultiLevelCacheBasic(t *testing.T) {
	// 使用mock缓存
	mc := &MultiLevelCache{
		L1: NewLocalCache(100 * time.Millisecond),
		L2: NewMockRedisCache(),
		L3: NewMockDiskCache(),
		hitCounter:  &CacheCounter{counts: make(map[string]int64)},
		missCounter: &CacheCounter{counts: make(map[string]int64)},
	}
	
	ctx := context.Background()
	userID := "test_user"
	
	// 测试Set/Get Speed
	speed := 45.5
	mc.SetSpeed(ctx, userID, speed)
	
	// 应该从L1获取
	got, ok := mc.GetSpeed(ctx, userID)
	if !ok {
		t.Fatal("Expected to get speed from cache")
	}
	if got != speed {
		t.Errorf("Expected speed %f, got %f", speed, got)
	}
	
	t.Logf("L1 hit rate: %.2f%%", mc.GetHitRate())
}

// TestMultiLevelCacheHitRate 命中率测试
func TestMultiLevelCacheHitRate(t *testing.T) {
	mc := &MultiLevelCache{
		L1: NewLocalCache(100 * time.Millisecond),
		L2: NewMockRedisCache(),
		L3: NewMockDiskCache(),
		hitCounter:  &CacheCounter{counts: make(map[string]int64)},
		missCounter: &CacheCounter{counts: make(map[string]int64)},
	}
	
	ctx := context.Background()
	
	// 模拟1000次请求，预填充800次
	for i := 0; i < 800; i++ {
		userID := string(rune(i))
		mc.SetSpeed(ctx, userID, float64(i))
	}
	
	// 查询1000次 (800命中，200未命中)
	for i := 0; i < 1000; i++ {
		userID := string(rune(i % 1000))
		mc.GetSpeed(ctx, userID)
	}
	
	hitRate := mc.GetHitRate()
	t.Logf("Cache hit rate: %.2f%%", hitRate)
	
	// 预期命中率约80%
	if hitRate < 70 {
		t.Errorf("Hit rate too low: %.2f%%", hitRate)
	}
}

// TestLocalCacheTTL TTL测试
func TestLocalCacheTTL(t *testing.T) {
	lc := NewLocalCache(100 * time.Millisecond)
	
	// 设置值
	lc.Set("key1", "value1", 50*time.Millisecond)
	
	// 立即获取 - 应该命中
	if _, ok := lc.Get("key1"); !ok {
		t.Error("Expected to get value immediately")
	}
	
	// 等待过期
	time.Sleep(100 * time.Millisecond)
	
	// 再次获取 - 应该未命中
	if _, ok := lc.Get("key1"); ok {
		t.Error("Expected value to expire")
	}
}

// TestLocalCacheEviction 淘汰测试
func TestLocalCacheEviction(t *testing.T) {
	lc := NewLocalCache(100 * time.Millisecond)
	lc.maxSize = 5 // 设置小容量便于测试
	
	// 添加6个元素，触发淘汰
	for i := 0; i < 6; i++ {
		lc.Set(string(rune(i)), i, time.Minute)
	}
	
	// 验证大小不超过5
	if lc.Size() > 5 {
		t.Errorf("Cache size %d > max 5", lc.Size())
	}
}

// BenchmarkMultiLevelCache 基准测试
func BenchmarkMultiLevelCache(b *testing.B) {
	mc := &MultiLevelCache{
		L1: NewLocalCache(100 * time.Millisecond),
		L2: NewMockRedisCache(),
		L3: NewMockDiskCache(),
		hitCounter:  &CacheCounter{counts: make(map[string]int64)},
		missCounter: &CacheCounter{counts: make(map[string]int64)},
	}
	
	ctx := context.Background()
	
	// 预填充
	for i := 0; i < 1000; i++ {
		mc.SetSpeed(ctx, string(rune(i)), float64(i))
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			userID := string(rune(i % 1000))
			mc.GetSpeed(ctx, userID)
			i++
		}
	})
}

// Mock 实现

type MockRedisCache struct {
	data map[string][]byte
}

func NewMockRedisCache() *MockRedisCache {
	return &MockRedisCache{data: make(map[string][]byte)}
}

func (m *MockRedisCache) GetSpeed(ctx context.Context, userID string) (float64, bool) {
	return 0, false
}

func (m *MockRedisCache) SetSpeed(ctx context.Context, userID string, speed float64) error {
	return nil
}

func (m *MockRedisCache) GetTrajectory(ctx context.Context, userID string, start, end int64) ([]algorithm.GPSPoint, bool) {
	return nil, false
}

func (m *MockRedisCache) SetTrajectory(ctx context.Context, userID string, start, end int64, points []algorithm.GPSPoint) error {
	return nil
}

func (m *MockRedisCache) GetStayPoints(ctx context.Context, userID string, days int) ([]algorithm.StayPoint, bool) {
	return nil, false
}

func (m *MockRedisCache) SetStayPoints(ctx context.Context, userID string, days int, stays []algorithm.StayPoint) error {
	return nil
}

func (m *MockRedisCache) DeleteUserCache(ctx context.Context, userID string) error {
	return nil
}

type MockDiskCache struct{}

func NewMockDiskCache() *MockDiskCache {
	return &MockDiskCache{}
}

func (m *MockDiskCache) GetSpeed(userID string) (float64, bool) {
	return 0, false
}

func (m *MockDiskCache) SetSpeed(userID string, speed float64) {}

func (m *MockDiskCache) GetTrajectory(userID string, start, end int64) ([]algorithm.GPSPoint, bool) {
	return nil, false
}

func (m *MockDiskCache) SetTrajectory(userID string, start, end int64, points []algorithm.GPSPoint) {}
