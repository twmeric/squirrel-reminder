package optimizer

import (
	"time"

	pb "github.com/squirrelawake/m01-state-engine/proto"
)

// BatteryOptimizer manages detection intervals based on battery level
type BatteryOptimizer struct {
	// 基础间隔配置（毫秒）
	baseIntervals map[pb.DetectionStatus]time.Duration
}

// NewBatteryOptimizer creates a new optimizer instance
func NewBatteryOptimizer() *BatteryOptimizer {
	return &BatteryOptimizer{
		baseIntervals: map[pb.DetectionStatus]time.Duration{
			pb.DetectionStatus_IN_TRANSIT: 5 * time.Second,
			pb.DetectionStatus_WALKING:    8 * time.Second,
			pb.DetectionStatus_STATIONARY: 10 * time.Second,
		},
	}
}

// CalculateInterval calculates the optimal detection interval
func (bo *BatteryOptimizer) CalculateInterval(
	batteryLevel float64,
	currentState pb.DetectionStatus,
) time.Duration {
	
	// 获取基础间隔
	base := bo.baseIntervals[currentState]
	if base == 0 {
		base = 5 * time.Second
	}
	
	// 根据电量调整
	switch {
	case batteryLevel > 0.5: // 高电量
		return base
	case batteryLevel > 0.2: // 中电量
		return base * 2
	default: // 低电量
		return base * 3
	}
}

// GetSensorStrategy returns which sensors to use based on battery
func (bo *BatteryOptimizer) GetSensorStrategy(batteryLevel float64) SensorStrategy {
	switch {
	case batteryLevel < 0.1: // 极低电量
		return SensorStrategy{
			UseGPS:           false,
			UseAccelerometer: true,
			UseGyroscope:     false,
			BatchSize:        10,
		}
	case batteryLevel < 0.2: // 低电量
		return SensorStrategy{
			UseGPS:           true,
			UseAccelerometer: true,
			UseGyroscope:     false,
			BatchSize:        5,
		}
	default: // 正常电量
		return SensorStrategy{
			UseGPS:           true,
			UseAccelerometer: true,
			UseGyroscope:     true,
			BatchSize:        1,
		}
	}
}

// SensorStrategy defines which sensors to use
type SensorStrategy struct {
	UseGPS           bool
	UseAccelerometer bool
	UseGyroscope     bool
	BatchSize        int
}

// EstimateBatteryDrain estimates battery consumption per hour
func (bo *BatteryOptimizer) EstimateBatteryDrain(
	interval time.Duration,
	strategy SensorStrategy,
) float64 {
	// 简化的电量消耗估算（%/小时）
	baseDrain := 0.5 // 基础消耗
	
	if strategy.UseGPS {
		baseDrain += 2.0
	}
	if strategy.UseAccelerometer {
		baseDrain += 0.5
	}
	if strategy.UseGyroscope {
		baseDrain += 0.3
	}
	
	// 根据频率调整
	frequency := float64(time.Hour) / float64(interval)
	return baseDrain * frequency * 0.1
}
