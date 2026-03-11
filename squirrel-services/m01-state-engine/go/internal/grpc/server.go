package grpc

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "squirrel-m01/proto"
)

// StateServer 状态检测服务器
type StateServer struct {
	pb.UnimplementedStateDetectionServer
}

// NewStateServer 创建状态服务器
func NewStateServer() *StateServer {
	return &StateServer{}
}

// DetectState 检测状态
func (s *StateServer) DetectState(ctx context.Context, req *pb.StateRequest) (*pb.StateResponse, error) {
	log.Printf("[m01] DetectState for user: %s", req.UserId)
	
	confidence := 0.85
	targetState := "stationary"
	triggered := false
	
	if req.Speed > 20 {
		targetState = "in_transit"
		triggered = req.ScreenOn && req.Handheld
	} else if req.Speed > 5 {
		targetState = "walking"
	}
	
	return &pb.StateResponse{
		Status:     targetState,
		Confidence: confidence,
		Triggered:  triggered,
	}, nil
}

// StartServer 启动gRPC服务器
func StartServer(port string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	
	grpcServer := grpc.NewServer()
	pb.RegisterStateDetectionServer(grpcServer, NewStateServer())
	
	log.Printf("[m01] gRPC server on %s", port)
	return grpcServer.Serve(lis)
}
