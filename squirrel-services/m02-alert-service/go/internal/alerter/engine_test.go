package alerter

import (
	"testing"
	"time"
)

func TestEngine_Trigger(t *testing.T) {
	engine := NewEngine()
	
	alert := &Alert{
		UserID:  "user_001",
		Level:   "L2",
		Type:    "job_change",
		Title:   "检测到工作变化",
		Message: "您的工作地点发生了变化",
	}
	
	engine.Trigger(alert)
	
	// 验证告警已存储
	alerts := engine.GetAlerts("user_001")
	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}
	
	if alerts[0].Type != "job_change" {
		t.Errorf("Expected type job_change, got %s", alerts[0].Type)
	}
}

func TestEngine_GetAlerts(t *testing.T) {
	engine := NewEngine()
	
	// 添加测试告警
	engine.Trigger(&Alert{
		UserID: "user_002",
		Level:  "L1",
		Type:   "move",
	})
	
	engine.Trigger(&Alert{
		UserID: "user_002",
		Level:  "L3",
		Type:   "unemployment",
	})
	
	alerts := engine.GetAlerts("user_002")
	if len(alerts) != 2 {
		t.Errorf("Expected 2 alerts, got %d", len(alerts))
	}
}

func TestEngine_Subscribe(t *testing.T) {
	engine := NewEngine()
	
	ch := engine.Subscribe("user_003")
	if ch == nil {
		t.Error("Expected non-nil channel")
	}
	
	// 触发告警并验证收到通知
	go engine.Trigger(&Alert{
		UserID: "user_003",
		Level:  "L2",
		Type:   "test",
	})
	
	select {
	case alert := <-ch:
		if alert.UserID != "user_003" {
			t.Errorf("Expected user_003, got %s", alert.UserID)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for alert")
	}
}

func TestEngine_CleanupExpired(t *testing.T) {
	engine := NewEngine()
	
	// 添加即将过期的告警
	oldAlert := &Alert{
		ID:        "old_001",
		UserID:    "user_004",
		Level:     "L1",
		Type:      "old",
		ExpiresAt: time.Now().Add(-time.Hour), // 已过期
	}
	
	engine.mu.Lock()
	engine.alerts["old_001"] = oldAlert
	engine.mu.Unlock()
	
	// 执行清理
	engine.cleanupExpired()
	
	// 验证已清理
	engine.mu.RLock()
	_, exists := engine.alerts["old_001"]
	engine.mu.RUnlock()
	
	if exists {
		t.Error("Expired alert should be cleaned up")
	}
}

func BenchmarkEngine_Trigger(b *testing.B) {
	engine := NewEngine()
	
	for i := 0; i < b.N; i++ {
		engine.Trigger(&Alert{
			UserID: "bench_user",
			Level:  "L1",
			Type:   "benchmark",
		})
	}
}
