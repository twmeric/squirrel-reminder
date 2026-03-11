package detector

import (
	"log"
	"time"
)

// BatteryOptimizer 电池优化器
type BatteryOptimizer struct {
	currentInterval time.Duration
	minInterval     time.Duration
	maxInterval     time.Duration
	lastState       string
}

// NewBatteryOptimizer 创建电池优化器
func NewBatteryOptimizer() *BatteryOptimizer {
	return &BatteryOptimizer{
		currentInterval: 10 * time.Second,
		minInterval:     5 * time.Second,
		maxInterval:     60 * time.Second,
		lastState:       "stationary",
	}
}

// GetPollingInterval 获取当前轮询间隔
func (bo *BatteryOptimizer) GetPollingInterval() time.Duration {
	return bo.currentInterval
}

// UpdateBasedOnState 根据状态更新轮询间隔
func (bo *BatteryOptimizer) UpdateBasedOnState(state string) time.Duration {
	switch state {
	case "in_transit":
		bo.currentInterval = bo.minInterval
	case "walking":
		bo.currentInterval = 15 * time.Second
	case "stationary":
		bo.currentInterval = bo.maxInterval
	}
	
	if state != bo.lastState {
		log.Printf("[BatteryOptimizer] State: %s -> %s, interval: %v", 
			bo.lastState, state, bo.currentInterval)
		bo.lastState = state
	}
	
	return bo.currentInterval
}
