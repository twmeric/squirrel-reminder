package detector

import (
	"math"
)

// AccelerometerData represents 3-axis accelerometer reading
type AccelerometerData struct {
	X, Y, Z float64
}

// HandheldThresholds for detection
type HandheldThresholds struct {
	HandheldMin float64 // 手持最小方差
	PocketMax   float64 // 口袋最大方差
	DesktopXY   float64 // 桌面XY轴阈值
}

var defaultThresholds = HandheldThresholds{
	HandheldMin: 0.15,
	PocketMax:   0.10,
	DesktopXY:   0.05,
}

// AnalyzeHandheld analyzes accelerometer data to determine if phone is handheld
// Returns confidence score 0.0-1.0
func AnalyzeHandheld(readings []AccelerometerData) float64 {
	if len(readings) < 3 {
		return 0.5 // 数据不足，返回中等置信度
	}
	
	// 计算各轴方差
	varX := calculateVariance(extractAxis(readings, 'X'))
	varY := calculateVariance(extractAxis(readings, 'Y'))
	varZ := calculateVariance(extractAxis(readings, 'Z'))
	
	// 平均方差
	avgVariance := (varX + varY + varZ) / 3
	
	// 检测模式
	switch {
	case isHandheld(avgVariance, varZ):
		return 0.9
	case isPocket(avgVariance):
		return 0.3
	case isDesktop(readings, varX, varY):
		return 0.1
	default:
		return 0.5
	}
}

// isHandheld detects handheld mode by variance
func isHandheld(avgVariance, zVariance float64) bool {
	return avgVariance > defaultThresholds.HandheldMin && zVariance > 0.1
}

// isPocket detects pocket mode (lower variance)
func isPocket(avgVariance float64) bool {
	return avgVariance >= 0.05 && avgVariance <= defaultThresholds.PocketMax
}

// isDesktop detects stationary on desk
func isDesktop(readings []AccelerometerData, varX, varY float64) bool {
	if varX > defaultThresholds.DesktopXY || varY > defaultThresholds.DesktopXY {
		return false
	}
	
	// 检查Z轴是否接近重力加速度
	zMean := mean(extractAxis(readings, 'Z'))
	return math.Abs(zMean-9.8) < 0.5
}

// Helper functions

func extractAxis(readings []AccelerometerData, axis rune) []float64 {
	result := make([]float64, len(readings))
	for i, r := range readings {
		switch axis {
		case 'X':
			result[i] = r.X
		case 'Y':
			result[i] = r.Y
		case 'Z':
			result[i] = r.Z
		}
	}
	return result
}

func calculateVariance(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	
	m := mean(data)
	variance := 0.0
	for _, v := range data {
		diff := v - m
		variance += diff * diff
	}
	return variance / float64(len(data))
}

func mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}
