package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/squirrel-awake/m03-trajectory/internal/service"
	pb "github.com/squirrel-awake/m03-trajectory/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 配置
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50053"
	}
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8083"
	}

	// 初始化存储
	store, err := storage.NewTiDBStorage(storage.DefaultConfig())
	if err != nil {
		log.Printf("⚠️  TiDB connection failed, using mock storage: %v", err)
		store = nil
	} else {
		// 初始化表结构
		if err := store.InitSchema(); err != nil {
			log.Printf("⚠️  Failed to init schema: %v", err)
		}
		defer store.Close()
		log.Println("✅ TiDB connected")
	}

	// 初始化服务
	var trajectorySvc pb.TrajectoryServiceServer
	if store != nil {
		trajectorySvc = service.NewTrajectoryServiceV2(store)
		log.Println("✅ Using TrajectoryServiceV2 (with TiDB)")
	} else {
		trajectorySvc = service.NewTrajectoryService()
		log.Println("✅ Using TrajectoryService (mock)")
	}

	// 启动 gRPC 服务
	grpcServer := grpc.NewServer()
	pb.RegisterTrajectoryServiceServer(grpcServer, trajectorySvc)
	reflection.Register(grpcServer)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen gRPC: %v", err)
	}

	go func() {
		log.Printf("🚀 m03-trajectory gRPC server starting on port %s", grpcPort)
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// 启动 HTTP 服务（健康检查）
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"m03-trajectory","version":"1.1.1"}`))
	})
	httpMux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// 检查依赖（TiDB, Redis）
		ready := trajectorySvc.Ready()
		if ready {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ready":true}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"ready":false}`))
		}
	})

	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: httpMux,
	}

	go func() {
		log.Printf("🚀 m03-trajectory HTTP server starting on port %s", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⚠️  Shutting down m03-trajectory server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	httpServer.Shutdown(ctx)
	grpcServer.GracefulStop()

	log.Println("✅ m03-trajectory server stopped")
}
