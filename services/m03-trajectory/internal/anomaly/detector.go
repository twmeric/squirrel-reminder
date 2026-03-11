// detector.go - 异常检测引擎

package anomaly

import (
	"math"
	"sync"
	"time"
)

// Detector 异常检测器
type Detector struct {
	mu         sync.RWMutex
	metrics    map[string]*MetricWindow
	windowSize int
	threshold  float64
}

// MetricWindow 指标滑动窗口
type MetricWindow struct {
	Name       string
	Values     []float64
	Timestamps []time.Time
	MaxSize    int
}

// AnomalyResult 异常检测结果
type AnomalyResult struct {
	Metric    string
	Value     float64
	Expected  float64
	Score     float64
	IsAnomaly bool
	Reason    string
}

// NewDetector 创建异常检测器
func NewDetector(windowSize int) *Detector {
	return &Detector{
		metrics:    make(map[string]*MetricWindow),
		windowSize: windowSize,
		threshold:  0.95,
	}
}

// AddMetric 添加指标值
func (d *Detector) AddMetric(name string, value float64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	window, exists := d.metrics[name]
	if !exists {
		window = &MetricWindow{
			Name:    name,
			MaxSize: d.windowSize,
		}
		d.metrics[name] = window
	}

	window.Values = append(window.Values, value)
	window.Timestamps = append(window.Timestamps, time.Now())

	if len(window.Values) > window.MaxSize {
		window.Values = window.Values[1:]
		window.Timestamps = window.Timestamps[1:]
	}
}

// Detect 检测异常
func (d *Detector) Detect(name string, currentValue float64) *AnomalyResult {
	d.mu.RLock()
	defer d.mu.RUnlock()

	window, exists := d.metrics[name]
	if !exists || len(window.Values) < 10 {
		return &AnomalyResult{
			Metric:    name,
			Value:     currentValue,
			IsAnomaly: false,
			Reason:    "insufficient data",
		}
	}

	mean, std := calculateStats(window.Values)
	zScore := math.Abs(currentValue-mean) / std
	anomalyScore := math.Min(zScore/3.0, 1.0)
	isAnomaly := anomalyScore > d.threshold

	result := &AnomalyResult{
		Metric:    name,
		Value:     currentValue,
		Expected:  mean,
		Score:     anomalyScore,
		IsAnomaly: isAnomaly,
	}

	if isAnomaly {
		if currentValue > mean {
			result.Reason = "value significantly higher than expected"
		} else {
			result.Reason = "value significantly lower than expected"
		}
	}

	return result
}

func calculateStats(values []float64) (mean, std float64) {
	if len(values) == 0 {
		return 0, 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	var variance float64
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	std = math.Sqrt(variance / float64(len(values)))

	if std == 0 {
		std = 0.001
	}

	return mean, std
}
