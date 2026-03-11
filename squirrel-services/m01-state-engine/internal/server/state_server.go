package server

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/squirrelawake/m01-state-engine/internal/detector"
	"github.com/squirrelawake/m01-state-engine/internal/optimizer"
	pb "github.com/squirrelawake/m01-state-engine/proto"
	m03pb "github.com/squirrelawake/shared/proto/m03"
)

// StateDetectionServer implements the state detection service
type StateDetectionServer struct {
	pb.UnimplementedStateDetectionServiceServer
	
	stateMachine     *detector.StateMachine
	batteryOptimizer *optimizer.BatteryOptimizer
	m03Client        m03pb.TrajectoryProcessorClient
	
	// 用户状态缓存
	userStates sync.Map // map[string]*UserStateContext
}

// UserStateContext maintains per-user detection context
type UserStateContext struct {
	UserID           string
	CurrentState     pb.DetectionStatus
	History          []pb.StateTransition
	LastUpdateTime   time.Time
	ConsecutiveCount int
}

// NewStateDetectionServer creates a new server instance
func NewStateDetectionServer() *StateDetectionServer {
	return &StateDetectionServer{
		stateMachine:     detector.NewStateMachine(),
		batteryOptimizer: optimizer.NewBatteryOptimizer(),
	}
}

// DetectState is the core RPC method
func (s *StateDetectionServer) DetectState(ctx context.Context, req *pb.StateDetectionRequest) (*pb.StateDetectionResult, error) {
	startTime := time.Now()
	
	// 获取或创建用户上下文
	userCtx := s.getOrCreateUserContext(req.UserId)
	
	// 电池优化：计算检测间隔
	interval := s.batteryOptimizer.CalculateInterval(req.BatteryLevel, userCtx.CurrentState)
	
	// 手持检测
	handheldConfidence := detector.AnalyzeHandheld(req.AccelerometerData)
	
	// 状态机判断
	targetState := s.stateMachine.DetermineTargetState(
		req.SpeedKmh,
		req.ScreenOn,
		handheldConfidence,
		req.IsInMetroArea,
	)
	
	// 防抖处理
	shouldTrigger := s.stateMachine.ShouldTransition(
		userCtx.CurrentState,
		targetState,
		userCtx.ConsecutiveCount,
	)
	
	if shouldTrigger {
		userCtx.CurrentState = targetState
		userCtx.ConsecutiveCount = 0
		
		// 记录状态转换
		transition := pb.StateTransition{
			FromState:   userCtx.CurrentState,
			ToState:     targetState,
			Timestamp:   time.Now().UnixMilli(),
			TriggerReason: s.getTriggerReason(req),
		}
		userCtx.History = append(userCtx.History, transition)
	} else {
		userCtx.ConsecutiveCount++
	}
	
	userCtx.LastUpdateTime = time.Now()
	
	// 如果进入交通工具模式，查询m03获取轨迹信息
	var nextCheckInfo *pb.NextCheckInfo
	if targetState == pb.DetectionStatus_IN_TRANSIT && s.m03Client != nil {
		nextCheckInfo = s.queryTrajectoryInfo(ctx, req)
	}
	
	latency := time.Since(startTime).Milliseconds()
	
	return &pb.StateDetectionResult{
		Status:            targetState,
		Confidence:        s.calculateConfidence(req, handheldConfidence),
		Triggered:         shouldTrigger,
		NextCheckInterval: int64(interval.Milliseconds()),
		NextCheckInfo:     nextCheckInfo,
		LatencyMs:         latency,
	}, nil
}

// GetCurrentStatus returns the current status for a user
func (s *StateDetectionServer) GetCurrentStatus(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	userCtx, exists := s.userStates.Load(req.UserId)
	if !exists {
		return &pb.StatusResponse{
			Exists: false,
		}, nil
	}
	
	ctx := userCtx.(*UserStateContext)
	return &pb.StatusResponse{
		Exists:       true,
		CurrentState: ctx.CurrentState,
		History:      ctx.History,
		LastUpdate:   ctx.LastUpdateTime.UnixMilli(),
	}, nil
}

// SubscribeToChanges streams state changes
func (s *StateDetectionServer) SubscribeToChanges(req *pb.SubscribeRequest, stream pb.StateDetectionService_SubscribeToChangesServer) error {
	// TODO: Implement streaming with proper context handling
	return nil
}

// Helper methods

func (s *StateDetectionServer) getOrCreateUserContext(userID string) *UserStateContext {
	if val, ok := s.userStates.Load(userID); ok {
		return val.(*UserStateContext)
	}
	
	ctx := &UserStateContext{
		UserID:         userID,
		CurrentState:   pb.DetectionStatus_STATIONARY,
		History:        make([]pb.StateTransition, 0),
		LastUpdateTime: time.Now(),
	}
	s.userStates.Store(userID, ctx)
	return ctx
}

func (s *StateDetectionServer) calculateConfidence(req *pb.StateDetectionRequest, handheldConfidence float64) float64 {
	// 多因子置信度计算
	confidence := 0.0
	
	// 速度因子 (30%)
	if req.SpeedKmh > 20 {
		confidence += 0.3
	} else if req.SpeedKmh > 10 {
		confidence += 0.15
	}
	
	// 屏幕因子 (20%)
	if req.ScreenOn {
		confidence += 0.2
	}
	
	// 手持因子 (30%)
	confidence += handheldConfidence * 0.3
	
	// 地理围栏因子 (20%)
	if req.IsInMetroArea {
		confidence += 0.2
	}
	
	return confidence
}

func (s *StateDetectionServer) getTriggerReason(req *pb.StateDetectionRequest) string {
	if req.SpeedKmh > 20 && req.ScreenOn && req.IsInMetroArea {
		return "FOCUSED_TRANSIT_MODE"
	}
	return "STATE_CHANGE"
}

func (s *StateDetectionServer) queryTrajectoryInfo(ctx context.Context, req *pb.StateDetectionRequest) *pb.NextCheckInfo {
	// 调用m03服务获取轨迹信息
	// 实际实现需要连接m03 gRPC客户端
	return &pb.NextCheckInfo{
		ShouldCheckTrajectory: true,
		RecommendedIntervalMs: 5000,
	}
}
