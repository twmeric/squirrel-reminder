package alerter

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// Alert 告警定义
type Alert struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Level     string    `json:"level"` // L1, L2, L3
	Type      string    `json:"type"`  // job_change, move, unemployment
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Engine 告警引擎
type Engine struct {
	alerts     map[string]*Alert
	mu         sync.RWMutex
	subscribers map[string][]chan *Alert
}

// NewEngine 创建告警引擎
func NewEngine() *Engine {
	return &Engine{
		alerts:      make(map[string]*Alert),
		subscribers: make(map[string][]chan *Alert),
	}
}

// Start 启动引擎
func (e *Engine) Start() {
	log.Println("[m02] AlertEngine started")
	// 定期清理过期告警
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		e.cleanupExpired()
	}
}

// Trigger 触发告警
func (e *Engine) Trigger(alert *Alert) {
	e.mu.Lock()
	defer e.mu.Unlock()

	alert.ID = generateID()
	alert.CreatedAt = time.Now()
	alert.ExpiresAt = time.Now().Add(24 * time.Hour)

	e.alerts[alert.ID] = alert

	// 通知订阅者
	e.notifySubscribers(alert)

	log.Printf("[m02] Alert triggered: %s for user %s (level: %s)", 
		alert.Type, alert.UserID, alert.Level)
}

// GetAlerts 获取用户告警
func (e *Engine) GetAlerts(userID string) []*Alert {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*Alert
	for _, alert := range e.alerts {
		if alert.UserID == userID && time.Now().Before(alert.ExpiresAt) {
			result = append(result, alert)
		}
	}
	return result
}

// Subscribe 订阅告警
func (e *Engine) Subscribe(userID string) chan *Alert {
	ch := make(chan *Alert, 10)
	e.mu.Lock()
	e.subscribers[userID] = append(e.subscribers[userID], ch)
	e.mu.Unlock()
	return ch
}

func (e *Engine) notifySubscribers(alert *Alert) {
	subs, ok := e.subscribers[alert.UserID]
	if !ok {
		return
	}
	
	for _, ch := range subs {
		select {
		case ch <- alert:
		default:
		}
	}
}

func (e *Engine) cleanupExpired() {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	for id, alert := range e.alerts {
		if now.After(alert.ExpiresAt) {
			delete(e.alerts, id)
		}
	}
}

// HandleHTTP HTTP 处理器
func (e *Engine) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var alert Alert
		if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		e.Trigger(&alert)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(alert)

	case http.MethodGet:
		userID := r.URL.Query().Get("user_id")
		alerts := e.GetAlerts(userID)
		json.NewEncoder(w).Encode(alerts)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func generateID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
