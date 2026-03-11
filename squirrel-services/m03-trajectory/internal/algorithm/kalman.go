package algorithm

import (
	"math"
)

// KalmanFilter 卡尔曼滤波器 - 用于GPS速度平滑
type KalmanFilter struct {
	// 状态
	x float64 // 估计值
	p float64 // 估计误差协方差
	
	// 参数
	q float64 // 过程噪声
	r float64 // 测量噪声
}

// NewKalmanFilter 创建卡尔曼滤波器
func NewKalmanFilter(initialValue, processNoise, measurementNoise float64) *KalmanFilter {
	return &KalmanFilter{
		x: initialValue,
		p: 1.0,  // 初始不确定性
		q: processNoise,
		r: measurementNoise,
	}
}

// Update 更新滤波器
func (kf *KalmanFilter) Update(measurement float64) float64 {
	// 预测步骤
	kf.p = kf.p + kf.q
	
	// 更新步骤
	k := kf.p / (kf.p + kf.r)  // 卡尔曼增益
	kf.x = kf.x + k*(measurement-kf.x)
	kf.p = (1 - k) * kf.p
	
	return kf.x
}

// SpeedSmoother 速度平滑器
type SpeedSmoother struct {
	kf *KalmanFilter
}

// NewSpeedSmoother 创建速度平滑器
func NewSpeedSmoother() *SpeedSmoother {
	return &SpeedSmoother{
		kf: NewKalmanFilter(0, 0.1, 1.0), // q=0.1(过程噪声小), r=1.0(测量噪声大)
	}
}

// Smooth 平滑速度值
func (ss *SpeedSmoother) Speed(speed float64) float64 {
	// 对于速度，使用对数尺度可能更合适
	if speed <= 0 {
		return ss.kf.Update(0)
	}
	return ss.kf.Update(speed)
}

// MultiPointSmoother 多点平滑 - 使用滑动窗口
type MultiPointSmoother struct {
	window []float64
	size   int
}

// NewMultiPointSmoother 创建多点平滑器
func NewMultiPointSmoother(windowSize int) *MultiPointSmoother {
	return &MultiPointSmoother{
		window: make([]float64, 0, windowSize),
		size:   windowSize,
	}
}

// Add 添加新值并获取平滑后的值
func (mps *MultiPointSmoother) Add(value float64) float64 {
	mps.window = append(mps.window, value)
	
	if len(mps.window) > mps.size {
		mps.window = mps.window[1:]
	}
	
	return mps.average()
}

func (mps *MultiPointSmoother) average() float64 {
	if len(mps.window) == 0 {
		return 0
	}
	
	var sum float64
	for _, v := range mps.window {
		sum += v
	}
	return sum / float64(len(mps.window))
}

// Median 中值滤波
func (mps *MultiPointSmoother) Median() float64 {
	if len(mps.window) == 0 {
		return 0
	}
	
	// 复制并排序
	sorted := make([]float64, len(mps.window))
	copy(sorted, mps.window)
	
	// 简单冒泡排序
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// SpeedValidator 速度验证器
type SpeedValidator struct {
	maxSpeed        float64 // 最大合理速度
	maxAcceleration float64 // 最大合理加速度
	lastSpeed       float64
	lastTime        int64
}

// NewSpeedValidator 创建速度验证器
func NewSpeedValidator() *SpeedValidator {
	return &SpeedValidator{
		maxSpeed:        120, // 120 km/h
		maxAcceleration: 10,  // 10 km/h/s
	}
}

// Validate 验证速度值
func (sv *SpeedValidator) Validate(speed float64, timestamp int64) (float64, bool) {
	// 检查最大值
	if speed > sv.maxSpeed {
		return sv.maxSpeed, false
	}
	
	// 如果是第一次，直接接受
	if sv.lastTime == 0 {
		sv.lastSpeed = speed
		sv.lastTime = timestamp
		return speed, true
	}
	
	// 计算时间差
	dt := float64(timestamp-sv.lastTime) / 1000.0 // 转换为秒
	if dt <= 0 {
		return speed, true
	}
	
	// 计算加速度
	acceleration := math.Abs(speed-sv.lastSpeed) / dt
	
	// 检查加速度
	if acceleration > sv.maxAcceleration {
		// 限制加速度
		if speed > sv.lastSpeed {
			speed = sv.lastSpeed + sv.maxAcceleration*dt
		} else {
			speed = sv.lastSpeed - sv.maxAcceleration*dt
		}
		return speed, false
	}
	
	sv.lastSpeed = speed
	sv.lastTime = timestamp
	return speed, true
}
