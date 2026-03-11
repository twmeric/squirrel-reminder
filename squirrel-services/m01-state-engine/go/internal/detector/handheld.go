package detector

import (
	"math"
)

// HandheldDetector 手持检测器
type HandheldDetector struct {
	varianceThreshold float64
}

// NewHandheldDetector 创建手持检测器
func NewHandheldDetector() *HandheldDetector {
	return &HandheldDetector{
		varianceThreshold: 0.15, // 加速度方差阈值
	}
}

// IsHandheld 判断是否手持
func (hd *HandheldDetector) IsHandheld(accelX, accelY, accelZ []float64) bool {
	if len(accelX) < 10 {
		return false
	}
	
	varX := calculateVariance(accelX)
	varY := calculateVariance(accelY)
	varZ := calculateVariance(accelZ)
	
	totalVar := (varX + varY + varZ) / 3
	return totalVar > hd.varianceThreshold
}

func calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	
	var variance float64
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	
	return variance / float64(len(values))
}

func calculateStdDev(values []float64) float64 {
	return math.Sqrt(calculateVariance(values))
}
