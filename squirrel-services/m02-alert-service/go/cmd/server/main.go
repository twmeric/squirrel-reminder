package main

import (
	"log"
	"net/http"
	"os"

	"squirrel-m02/internal/alerter"
	"squirrel-m02/internal/notifier"
	"squirrel-m02/internal/websocket"
)

func main() {
	// 初始化各模块
	alertEngine := alerter.NewEngine()
	wsHub := websocket.NewHub()
	notif := notifier.NewManager()

	// 启动 WebSocket Hub
	go wsHub.Run()

	// 启动告警引擎
	go alertEngine.Start()

	// HTTP 路由
	http.HandleFunc("/ws", wsHub.HandleWebSocket)
	http.HandleFunc("/api/v1/alert", alertEngine.HandleHTTP)
	http.HandleFunc("/api/v1/notify", notif.HandleHTTP)
	http.HandleFunc("/health", healthHandler)

	port := os.Getenv("M02_PORT")
	if port == "" {
		port = ":8081"
	}

	log.Printf("[m02] Alert Service starting on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"m02-alert-service"}`))
}
