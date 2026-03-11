package detector

import (
	"sync"

	pb "github.com/squirrelawake/m01-state-engine/proto"
)

// StateMachine manages state transitions with debouncing
type StateMachine struct {
	mu sync.RWMutex
	
	// 防抖配置：每个状态需要的连续确认次数
	requiredConsecutive map[pb.DetectionStatus]int
}

// NewStateMachine creates a new state machine instance
func NewStateMachine() *StateMachine {
	return &StateMachine{
		requiredConsecutive: map[pb.DetectionStatus]int{
			pb.DetectionStatus_IN_TRANSIT: 3,  // 交通工具模式需要3次确认（防抖）
			pb.DetectionStatus_WALKING:    2,  // 步行需要2次确认
			pb.DetectionStatus_STATIONARY: 1,  // 静止只需1次确认
		},
	}
}

// DetermineTargetState determines the target state based on inputs
func (sm *StateMachine) DetermineTargetState(
	speedKmh float64,
	screenOn bool,
	handheldConfidence float64,
	isInMetroArea bool,
) pb.DetectionStatus {
	
	// 优先级1：交通工具模式（最严格条件）
	if speedKmh > 20 && screenOn && handheldConfidence > 0.6 && isInMetroArea {
		return pb.DetectionStatus_IN_TRANSIT
	}
	
	// 优先级2：步行模式
	if speedKmh > 5 && speedKmh <= 20 {
		return pb.DetectionStatus_WALKING
	}
	
	// 默认：静止
	return pb.DetectionStatus_STATIONARY
}

// ShouldTransition applies debouncing logic
func (sm *StateMachine) ShouldTransition(
	currentState pb.DetectionStatus,
	targetState pb.DetectionStatus,
	consecutiveCount int,
) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	// 如果状态没有变化，不需要转换
	if currentState == targetState {
		return false
	}
	
	// 获取目标状态所需的连续确认次数
	required := sm.requiredConsecutive[targetState]
	if required <= 0 {
		required = 1
	}
	
	// 检查是否达到防抖阈值
	return consecutiveCount >= required
}

// GetRequiredConsecutive returns the required count for a state
func (sm *StateMachine) GetRequiredConsecutive(state pb.DetectionStatus) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	if count, ok := sm.requiredConsecutive[state]; ok {
		return count
	}
	return 1
}

// StatePriority returns the priority of a state (higher = more specific)
func (sm *StateMachine) StatePriority(state pb.DetectionStatus) int {
	switch state {
	case pb.DetectionStatus_IN_TRANSIT:
		return 3
	case pb.DetectionStatus_WALKING:
		return 2
	case pb.DetectionStatus_STATIONARY:
		return 1
	default:
		return 0
	}
}
