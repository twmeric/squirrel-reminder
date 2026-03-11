package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/squirrelawake/m01-state-engine/internal/server"
	pb "github.com/squirrelawake/m01-state-engine/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultPort = ":50051"
)

func main() {
	port := os.Getenv("M01_PORT")
	if port == "" {
		port = defaultPort
	}

	// 创建gRPC服务器
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute,
			MaxConnectionAge:      30 * time.Minute,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
	)

	// 注册服务
	stateServer := server.NewStateDetectionServer()
	pb.RegisterStateDetectionServiceServer(grpcServer, stateServer)

	// 健康检查服务
	pb.RegisterHealthServiceServer(grpcServer, server.NewHealthServer())

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	log.Printf("m01-state-engine starting on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
