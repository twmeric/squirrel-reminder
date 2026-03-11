package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Manager 通知管理器
type Manager struct {
	pushURL string
}

// NewManager 创建通知管理器
func NewManager() *Manager {
	return &Manager{
		pushURL: "https://api.push service.com/v1/send",
	}
}

// Notification 通知内容
type Notification struct {
	UserID  string `json:"user_id"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Level   string `json:"level"` // urgent, normal, low
	Data    map[string]string `json:"data"`
}

// SendPush 发送推送通知
func (m *Manager) SendPush(n *Notification) error {
	payload := map[string]interface{}{
		"to":      n.UserID,
		"title":   n.Title,
		"body":    n.Body,
		"priority": m.getPriority(n.Level),
		"data":    n.Data,
	}

	jsonData, _ := json.Marshal(payload)
	
	// 这里实际会调用第三方推送服务
	// resp, err := http.Post(m.pushURL, "application/json", bytes.NewBuffer(jsonData))
	
	// 模拟发送成功
	fmt.Printf("[m02] Push sent to %s: %s\n", n.UserID, n.Title)
	_ = jsonData
	return nil
}

// SendSMS 发送短信（L3紧急告警）
func (m *Manager) SendSMS(userID, message string) error {
	// 实际会调用短信服务商
	fmt.Printf("[m02] SMS sent to %s: %s\n", userID, message)
	return nil
}

// ScheduleReminder 定时提醒
func (m *Manager) ScheduleReminder(userID, content string, at time.Time) {
	// 使用定时任务调度
	go func() {
		time.Sleep(time.Until(at))
		m.SendPush(&Notification{
			UserID: userID,
			Title:  "提醒",
			Body:   content,
			Level:  "normal",
		})
	}()
}

// HandleHTTP HTTP 处理器
func (m *Manager) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var notif Notification
	if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := m.SendPush(&notif); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

func (m *Manager) getPriority(level string) string {
	switch level {
	case "urgent":
		return "high"
	case "normal":
		return "default"
	default:
		return "low"
	}
}
