package storage

import (
	"context"
	"database/sql"
	"time"
)

// HealthChecker 健康检查器
type HealthChecker struct {
	db *sql.DB
}

func NewHealthChecker(db *sql.DB) *HealthChecker {
	return &HealthChecker{db: db}
}

// Check 执行健康检查
func (h *HealthChecker) Check(ctx context.Context) map[string]string {
	results := make(map[string]string)
	
	// 检查数据库连接
	if err := h.checkDatabase(ctx); err != nil {
		results["database"] = "unhealthy: " + err.Error()
	} else {
		results["database"] = "healthy"
	}
	
	return results
}

func (h *HealthChecker) checkDatabase(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	return h.db.PingContext(ctx)
}

// IsHealthy 判断是否健康
func (h *HealthChecker) IsHealthy(ctx context.Context) bool {
	results := h.Check(ctx)
	for _, status := range results {
		if status != "healthy" {
			return false
		}
	}
	return true
}
