package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type LocationRequest struct {
	UserID    string  `json:"user_id"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Speed     float64 `json:"speed"`
	Timestamp string  `json:"timestamp"`
}

type AlertResponse struct {
	ShouldAlert bool   `json:"should_alert"`
	Message     string `json:"message"`
	NextStation string `json:"next_station"`
	StopsRemaining int `json:"stops_remaining"`
}

func main() {
	// 静态文件服务
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)
	
	// API 端点
	http.HandleFunc("/api/v1/location", handleLocation)
	http.HandleFunc("/api/v1/health", handleHealth)
	http.HandleFunc("/api/v1/route", handleRoute)
	
	fmt.Println("🐿️ 松鼠提醒 Mobile MVP 服务器启动...")
	fmt.Println("📱 访问: http://localhost:8080")
	fmt.Println("📊 API: http://localhost:8080/api/v1/health")
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req LocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// 模拟到站检测逻辑
	speed := req.Speed * 3.6 // 转换为 km/h
	response := AlertResponse{
		ShouldAlert:     speed > 20,
		Message:         "地鐵行駛中",
		NextStation:     "福田站",
		StopsRemaining:  2,
	}
	
	// 如果接近到站（模拟）
	if req.Lat > 22.4 {
		response.ShouldAlert = true
		response.Message = "⚠️ 即將到站！請準備下車"
		response.StopsRemaining = 0
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"version": "v1.2.1-mvp",
		"time": time.Now().Format(time.RFC3339),
	})
}

func handleRoute(w http.ResponseWriter, r *http.Request) {
	routes := []map[string]interface{}{
		{"id": "1", "name": "羅湖 → 福田", "stops": ["羅湖", "老街", "大劇院", "科學館", "華強路", "崗廈", "會展中心", "購物公園", "福田"]},
		{"id": "2", "name": "福田 → 羅湖", "stops": ["福田", "購物公園", "會展中心", "崗廈", "華強路", "科學館", "大劇院", "老街", "羅湖"]},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routes)
}
