// connection_pool.go - TiDB连接池优化

package optimization

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// OptimizedTiDBConfig 优化的TiDB配置
type OptimizedTiDBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	QueryTimeout    time.Duration
	ConnectTimeout  time.Duration
}

// DefaultOptimizedConfig 返回优化的默认配置
func DefaultOptimizedConfig() *OptimizedTiDBConfig {
	return &OptimizedTiDBConfig{
		Host:            "localhost",
		Port:            4000,
		User:            "root",
		Password:        "",
		Database:        "squirrel_m03",
		MaxOpenConns:    100,
		MaxIdleConns:    20,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
		QueryTimeout:    500 * time.Millisecond,
		ConnectTimeout:  5 * time.Second,
	}
}

// OptimizedTiDBStorage 优化的TiDB存储
type OptimizedTiDBStorage struct {
	db     *sql.DB
	config *OptimizedTiDBConfig
}

// NewOptimizedTiDBStorage 创建优化的存储实例
func NewOptimizedTiDBStorage(config *OptimizedTiDBConfig) (*OptimizedTiDBStorage, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&timeout=%s&readTimeout=%s&writeTimeout=%s",
		config.User, config.Password, config.Host, config.Port, config.Database,
		config.ConnectTimeout, config.QueryTimeout, config.QueryTimeout)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	storage := &OptimizedTiDBStorage{db: db, config: config}
	go storage.monitorConnections()
	return storage, nil
}

func (s *OptimizedTiDBStorage) monitorConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := s.db.Stats()
		log.Printf("[TiDB连接池] Open:%d InUse:%d Idle:%d Wait:%d",
			stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)
	}
}

// GetStats 获取连接池统计
func (s *OptimizedTiDBStorage) GetStats() sql.DBStats {
	return s.db.Stats()
}
