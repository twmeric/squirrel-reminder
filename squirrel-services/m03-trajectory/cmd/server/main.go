package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/squirrelawake/m03-trajectory/internal/service"
	pb "github.com/squirrelawake/m03-trajectory/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultPort = ":50053"
	defaultDBDSN = "user:password@tcp(localhost:4000)/squirrel?parseTime=true"
)

func main() {
	port := os.Getenv("M03_PORT")
	if port == "" {
		port = defaultPort
	}

	dbDSN := os.Getenv("TIDB_DSN")
	if dbDSN == "" {
		dbDSN = defaultDBDSN
	}

	// 连接TiDB
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to TiDB successfully")

	// 创建gRPC服务器
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 15 * time.Minute,
			MaxConnectionAge:  30 * time.Minute,
			Time:              5 * time.Second,
			Timeout:           1 * time.Second,
		}),
	)

	// 注册服务
	trajectoryService := service.NewTrajectoryService(db)
	pb.RegisterTrajectoryProcessorServer(grpcServer, trajectoryService)
	pb.RegisterLocationServiceServer(grpcServer, trajectoryService)
	pb.RegisterHealthServiceServer(grpcServer, service.NewHealthServer())

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	log.Printf("m03-trajectory starting on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
