// redis.go - Redis缓存层实现

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/squirrel-awake/m03-trajectory/internal/algorithm"
)

// RedisCache Redis缓存
type RedisCache struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(addr, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: 20,
	})

	return &RedisCache{
		client: client,
		prefix: "m03:",
		ttl:    5 * time.Minute,
	}
}

// Ping 检查连接
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// GetTrajectory 获取轨迹缓存
func (c *RedisCache) GetTrajectory(ctx context.Context, userID string, start, end int64) ([]algorithm.GPSPoint, error) {
	key := fmt.Sprintf("%strajectory:%s:%d:%d", c.prefix, userID, start, end)

	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var points []algorithm.GPSPoint
	if err := json.Unmarshal([]byte(data), &points); err != nil {
		return nil, err
	}

	return points, nil
}

// SetTrajectory 设置轨迹缓存
func (c *RedisCache) SetTrajectory(ctx context.Context, userID string, start, end int64, points []algorithm.GPSPoint) error {
	key := fmt.Sprintf("%strajectory:%s:%d:%d", c.prefix, userID, start, end)

	data, err := json.Marshal(points)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

// GetSpeed 获取速度缓存
func (c *RedisCache) GetSpeed(ctx context.Context, userID string) (float64, error) {
	key := fmt.Sprintf("%sspeed:%s", c.prefix, userID)

	val, err := c.client.Get(ctx, key).Float64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// SetSpeed 设置速度缓存
func (c *RedisCache) SetSpeed(ctx context.Context, userID string, speed float64) error {
	key := fmt.Sprintf("%sspeed:%s", c.prefix, userID)
	return c.client.Set(ctx, key, speed, 30*time.Second).Err()
}

// DeleteUserCache 删除用户缓存
func (c *RedisCache) DeleteUserCache(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("%s*:%s:*", c.prefix, userID)

	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

// Close 关闭连接
func (c *RedisCache) Close() error {
	return c.client.Close()
}
