package algorithm

import (
	"math"
	"testing"
)

func TestKalmanFilter_Update(t *testing.T) {
	kf := NewKalmanFilter(0, 0.1, 1.0)

	// 模拟带噪声的测量值
	trueValue := 30.0
	measurements := []float64{
		trueValue + 5,
		trueValue - 3,
		trueValue + 2,
		trueValue - 1,
		trueValue + 0.5,
	}

	var lastEstimate float64
	for _, m := range measurements {
		lastEstimate = kf.Update(m)
	}

	// 最终估计应接近真实值
	if math.Abs(lastEstimate-trueValue) > 3 {
		t.Errorf("Expected estimate close to %v, got %v", trueValue, lastEstimate)
	}
}

func TestSpeedSmoother(t *testing.T) {
	ss := NewSpeedSmoother()

	// 测试正常速度
	speeds := []float64{30, 32, 28, 31, 29, 30.5, 29.5}
	for _, s := range speeds {
		ss.Smooth(s)
	}

	// 测试零速度
	result := ss.Smooth(0)
	if result < 0 {
		t.Errorf("Speed should not be negative, got %v", result)
	}
}

func TestMultiPointSmoother(t *testing.T) {
	mps := NewMultiPointSmoother(5)

	values := []float64{10, 20, 30, 25, 15, 18, 22}
	for _, v := range values {
		mps.Add(v)
	}

	avg := mps.average()
	if avg <= 0 {
		t.Error("Average should be positive")
	}

	median := mps.Median()
	if median <= 0 {
		t.Error("Median should be positive")
	}
}

func TestSpeedValidator(t *testing.T) {
	sv := NewSpeedValidator()

	// 测试正常速度
	speed, valid := sv.Validate(50, 1000)
	if !valid {
		t.Error("Normal speed should be valid")
	}
	if speed != 50 {
		t.Errorf("Speed should be 50, got %v", speed)
	}

	// 测试超速
	speed, valid = sv.Validate(150, 2000)
	if valid {
		t.Error("Excessive speed should be invalid")
	}
	if speed > 120 {
		t.Error("Speed should be clamped to max")
	}
}

func BenchmarkKalmanFilter_Update(b *testing.B) {
	kf := NewKalmanFilter(0, 0.1, 1.0)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kf.Update(30 + float64(i%10))
	}
}
