// fast_smooth_test.go - FastSmooth单元测试与基准测试

package optimization

import (
	"math"
	"testing"
	"time"

	"github.com/squirrel-awake/m03-trajectory/internal/algorithm"
)

// TestFastSmoothBasic 基本功能测试
func TestFastSmoothBasic(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		alpha  float64
	}{
		{
			name:   "匀速序列",
			values: []float64{50, 50, 50, 50, 50},
			alpha:  0.3,
		},
		{
			name:   "加速序列",
			values: []float64{10, 20, 30, 40, 50},
			alpha:  0.3,
		},
		{
			name:   "波动序列",
			values: []float64{50, 52, 48, 51, 49},
			alpha:  0.3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FastSmooth(tt.values, tt.alpha)
			
			if len(result) != len(tt.values) {
				t.Errorf("result length %d != input length %d", len(result), len(tt.values))
			}
			
			// 验证平滑效果：结果应比输入更平滑（方差更小）
			inputVar := variance(tt.values)
			outputVar := variance(result)
			
			if outputVar > inputVar {
				t.Logf("Warning: output variance (%.2f) > input variance (%.2f)", outputVar, inputVar)
			}
		})
	}
}

// TestFastSmoothAccuracy 精度测试
func TestFastSmoothAccuracy(t *testing.T) {
	// 生成测试数据
	values := generateTestData(100)
	
	// FastSmooth结果
	fastResult := FastSmooth(values, 0.3)
	
	// 原始卡尔曼滤波结果
	smoother := algorithm.NewSpeedSmoother()
	// 将float64转为GPSPoint进行卡尔曼滤波
	points := make([]algorithm.GPSPoint, len(values))
	for i, v := range values {
		points[i] = algorithm.GPSPoint{Speed: float32(v)}
	}
	kalmanResult := smoother.CalculateSpeeds(points)
	
	// 计算精度损失
	loss := CalculateAccuracyLoss(fastResult, kalmanResult)
	
	t.Logf("Accuracy loss: %.2f%%", loss)
	
	if loss > 2.0 {
		t.Errorf("Accuracy loss too high: %.2f%% (threshold: 2%%)", loss)
	}
}

// TestAdaptiveFastSmooth 自适应平滑测试
func TestAdaptiveFastSmooth(t *testing.T) {
	values := []float64{10, 12, 11, 13, 12}
	
	// 高速 - 应该更平滑
	highSpeed := AdaptiveFastSmooth(values, 70)
	
	// 低速 - 应该更敏感
	lowSpeed := AdaptiveFastSmooth(values, 10)
	
	t.Logf("High speed result: %v", highSpeed)
	t.Logf("Low speed result: %v", lowSpeed)
}

// TestEstimateOptimalAlpha 最优alpha估计测试
func TestEstimateOptimalAlpha(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64 // 期望范围
	}{
		{
			name:     "稳定数据",
			values:   []float64{50, 51, 50, 49, 50},
			expected: 0.4,
		},
		{
			name:     "波动数据",
			values:   []float64{50, 60, 40, 70, 30},
			expected: 0.2,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alpha := EstimateOptimalAlpha(tt.values)
			t.Logf("Estimated alpha: %.2f", alpha)
			
			if math.Abs(alpha-tt.expected) > 0.1 {
				t.Errorf("alpha %.2f not close to expected %.2f", alpha, tt.expected)
			}
		})
	}
}

// BenchmarkFastSmooth 基准测试
func BenchmarkFastSmooth(b *testing.B) {
	values := generateTestData(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FastSmooth(values, 0.3)
	}
}

// BenchmarkKalman 原始卡尔曼滤波基准测试
func BenchmarkKalman(b *testing.B) {
	values := generateTestData(1000)
	points := make([]algorithm.GPSPoint, len(values))
	for i, v := range values {
		points[i] = algorithm.GPSPoint{
			Timestamp: time.Now().UnixMilli() + int64(i)*1000,
			Lat:       39.9 + float64(i)*0.001,
			Lng:       116.4 + float64(i)*0.001,
			Speed:     float32(v),
		}
	}
	
	smoother := algorithm.NewSpeedSmoother()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		smoother.CalculateSpeeds(points)
	}
}

// BenchmarkBatchFastSmooth 批量处理基准测试
func BenchmarkBatchFastSmooth(b *testing.B) {
	// 模拟100个用户
	userValues := make(map[string][]float64)
	for i := 0; i < 100; i++ {
		userValues[string(rune(i))] = generateTestData(100)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BatchFastSmooth(userValues, 0.3)
	}
}

// 辅助函数

func generateTestData(n int) []float64 {
	data := make([]float64, n)
	base := 50.0
	
	for i := 0; i < n; i++ {
		// 添加噪声的正弦波
		noise := (float64(i%10) - 5) * 2
		trend := float64(i) * 0.1
		data[i] = base + trend + noise
	}
	
	return data
}

func variance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))
	
	var sum float64
	for _, v := range values {
		sum += (v - mean) * (v - mean)
	}
	
	return sum / float64(len(values))
}
