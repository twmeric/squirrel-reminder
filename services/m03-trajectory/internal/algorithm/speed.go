// speed.go - 速度平滑算法实现
// 卡尔曼滤波 + 滑动窗口去噪
// 性能目标：P99 < 5ms

package algorithm

import (
	"math"
)

// SpeedSmoother 速度平滑器
type SpeedSmoother struct {
	WindowSize int     // 滑动窗口大小
	KalmanQ    float64 // 过程噪声协方差
	KalmanR    float64 // 测量噪声协方差
}

// NewSpeedSmoother 创建速度平滑器
func NewSpeedSmoother() *SpeedSmoother {
	return &SpeedSmoother{
		WindowSize: 5,    // 5点窗口
		KalmanQ:    0.01, // 过程噪声
		KalmanR:    1.0,  // 测量噪声
	}
}

// CalculateSpeeds 计算平滑速度序列
// 输入：轨迹点（需包含时间戳和坐标）
// 输出：平滑后的速度序列（km/h）
func (s *SpeedSmoother) CalculateSpeeds(points []GPSPoint) []float64 {
	if len(points) < 2 {
		return make([]float64, len(points))
	}

	// 步骤1: 计算瞬时速度
	instantSpeeds := s.calculateInstantSpeeds(points)

	// 步骤2: 异常值过滤（3-sigma原则）
	filteredSpeeds := s.filterOutliers(instantSpeeds)

	// 步骤3: 卡尔曼滤波
	kalmanSpeeds := s.kalmanFilter(filteredSpeeds)

	// 步骤4: 滑动窗口平滑
	smoothSpeeds := s.movingAverage(kalmanSpeeds)

	return smoothSpeeds
}

// GetCurrentSpeed 获取当前平滑速度（最新点）
func (s *SpeedSmoother) GetCurrentSpeed(points []GPSPoint) float64 {
	if len(points) == 0 {
		return 0
	}

	speeds := s.CalculateSpeeds(points)
	return speeds[len(speeds)-1]
}

// calculateInstantSpeeds 计算瞬时速度
func (s *SpeedSmoother) calculateInstantSpeeds(points []GPSPoint) []float64 {
	n := len(points)
	speeds := make([]float64, n)
	speeds[0] = 0 // 第一个点速度为0

	for i := 1; i < n; i++ {
		dist := haversineDistance(
			points[i-1].Lat, points[i-1].Lng,
			points[i].Lat, points[i].Lng,
		)
		timeDiff := float64(points[i].Timestamp-points[i-1].Timestamp) / 1000.0 // 秒

		if timeDiff > 0 {
			speedMPS := dist / timeDiff
			speeds[i] = speedMPS * 3.6 // 转换为 km/h
		} else {
			speeds[i] = 0
		}
	}

	return speeds
}

// filterOutliers 使用3-sigma原则过滤异常值
func (s *SpeedSmoother) filterOutliers(speeds []float64) []float64 {
	if len(speeds) < 3 {
		return speeds
	}

	// 计算均值和标准差
	mean := s.mean(speeds)
	std := s.stdDev(speeds, mean)

	if std == 0 {
		return speeds
	}

	// 过滤异常值（替换为均值）
	filtered := make([]float64, len(speeds))
	copy(filtered, speeds)

	threshold := 3.0 * std
	for i, v := range speeds {
		if math.Abs(v-mean) > threshold {
			filtered[i] = mean
		}
	}

	return filtered
}

// kalmanFilter 一维卡尔曼滤波
func (s *SpeedSmoother) kalmanFilter(measurements []float64) []float64 {
	n := len(measurements)
	if n == 0 {
		return nil
	}

	// 初始化
	x := measurements[0] // 状态估计
	p := 1.0             // 估计误差协方差

	filtered := make([]float64, n)
	filtered[0] = x

	for i := 1; i < n; i++ {
		z := measurements[i] // 测量值

		// 预测
		xPred := x
		pPred := p + s.KalmanQ

		// 更新
		k := pPred / (pPred + s.KalmanR) // 卡尔曼增益
		x = xPred + k*(z-xPred)
		p = (1 - k) * pPred

		filtered[i] = x
	}

	return filtered
}

// movingAverage 加权滑动窗口平均
func (s *SpeedSmoother) movingAverage(data []float64) []float64 {
	n := len(data)
	if n == 0 {
		return nil
	}

	result := make([]float64, n)
	halfWindow := s.WindowSize / 2

	for i := 0; i < n; i++ {
		start := max(0, i-halfWindow)
		end := min(n, i+halfWindow+1)

		// 加权平均（中心点权重更高）
		var weightedSum, weightSum float64
		for j := start; j < end; j++ {
			distance := abs(j - i)
			weight := 1.0 / float64(distance+1)
			weightedSum += data[j] * weight
			weightSum += weight
		}

		if weightSum > 0 {
			result[i] = weightedSum / weightSum
		} else {
			result[i] = data[i]
		}
	}

	return result
}

// mean 计算平均值
func (s *SpeedSmoother) mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	var sum float64
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// stdDev 计算标准差
func (s *SpeedSmoother) stdDev(data []float64, mean float64) float64 {
	if len(data) < 2 {
		return 0
	}
	var sum float64
	for _, v := range data {
		diff := v - mean
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(data)))
}

// GPSDriftFilter GPS漂移过滤器
type GPSDriftFilter struct {
	MaxAccuracy  float32 // 最大允许精度（米）
	MaxSpeed     float64 // 最大允许速度（km/h）
	MaxJumpDist  float64 // 最大跳跃距离（米）
	MinJumpTime  float64 // 最小跳跃时间（秒）
}

// NewGPSDriftFilter 创建漂移过滤器
func NewGPSDriftFilter() *GPSDriftFilter {
	return &GPSDriftFilter{
		MaxAccuracy: 50.0,   // 50米
		MaxSpeed:    120.0,  // 120 km/h
		MaxJumpDist: 500.0,  // 500米
		MinJumpTime: 5.0,    // 5秒
	}
}

// Filter 过滤漂移点
func (f *GPSDriftFilter) Filter(points []GPSPoint) []GPSPoint {
	if len(points) == 0 {
		return nil
	}

	filtered := make([]GPSPoint, 0, len(points))
	
	for i, p := range points {
		// 规则1: 精度检查
		if p.Accuracy > f.MaxAccuracy {
			continue
		}

		if i == 0 {
			filtered = append(filtered, p)
			continue
		}

		prev := filtered[len(filtered)-1]
		dist := haversineDistance(prev.Lat, prev.Lng, p.Lat, p.Lng)
		timeDiff := float64(p.Timestamp-prev.Timestamp) / 1000.0 // 秒

		// 规则2: 速度检查
		if timeDiff > 0 {
			speed := (dist / timeDiff) * 3.6 // km/h
			if speed > f.MaxSpeed {
				continue
			}
		}

		// 规则3: 跳跃检查
		if dist > f.MaxJumpDist && timeDiff < f.MinJumpTime {
			continue
		}

		filtered = append(filtered, p)
	}

	return filtered
}

// 辅助函数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
