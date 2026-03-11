// prometheus.go - Prometheus监控埋点

package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector Prometheus指标收集器
type MetricsCollector struct {
	// 延迟直方图
	requestDuration *prometheus.HistogramVec

	// 计数器
	requestTotal    *prometheus.CounterVec
	errorTotal      *prometheus.CounterVec
	cacheHitTotal   prometheus.Counter
	cacheMissTotal  prometheus.Counter

	// 仪表盘
	activeConnections prometheus.Gauge
	queueSize         prometheus.Gauge
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{
		// 请求延迟直方图
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "m03_request_duration_seconds",
				Help:    "Request duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
			},
			[]string{"method", "endpoint"},
		),

		// 请求总数
		requestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "m03_request_total",
				Help: "Total number of requests",
			},
			[]string{"method", "endpoint", "status"},
		),

		// 错误总数
		errorTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "m03_error_total",
				Help: "Total number of errors",
			},
			[]string{"method", "endpoint", "error_type"},
		),

		// 缓存命中
		cacheHitTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "m03_cache_hit_total",
				Help: "Total number of cache hits",
			},
		),

		// 缓存未命中
		cacheMissTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "m03_cache_miss_total",
				Help: "Total number of cache misses",
			},
		),

		// 活跃连接数
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "m03_active_connections",
				Help: "Number of active connections",
			},
		),

		// 队列大小
		queueSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "m03_queue_size",
				Help: "Current queue size",
			},
		),
	}

	// 注册所有指标
	prometheus.MustRegister(mc.requestDuration)
	prometheus.MustRegister(mc.requestTotal)
	prometheus.MustRegister(mc.errorTotal)
	prometheus.MustRegister(mc.cacheHitTotal)
	prometheus.MustRegister(mc.cacheMissTotal)
	prometheus.MustRegister(mc.activeConnections)
	prometheus.MustRegister(mc.queueSize)

	return mc
}

// RecordRequest 记录请求
func (mc *MetricsCollector) RecordRequest(method, endpoint string, duration time.Duration, err error) {
	mc.requestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())

	status := "success"
	if err != nil {
		status = "error"
		mc.errorTotal.WithLabelValues(method, endpoint, "unknown").Inc()
	}

	mc.requestTotal.WithLabelValues(method, endpoint, status).Inc()
}

// RecordCacheHit 记录缓存命中
func (mc *MetricsCollector) RecordCacheHit() {
	mc.cacheHitTotal.Inc()
}

// RecordCacheMiss 记录缓存未命中
func (mc *MetricsCollector) RecordCacheMiss() {
	mc.cacheMissTotal.Inc()
}

// SetActiveConnections 设置活跃连接数
func (mc *MetricsCollector) SetActiveConnections(n float64) {
	mc.activeConnections.Set(n)
}

// SetQueueSize 设置队列大小
func (mc *MetricsCollector) SetQueueSize(n float64) {
	mc.queueSize.Set(n)
}

// Handler 返回Prometheus HTTP handler
func (mc *MetricsCollector) Handler() http.Handler {
	return promhttp.Handler()
}

// StartMetricsServer 启动metrics服务器
func (mc *MetricsCollector) StartMetricsServer(port int) {
	http.Handle("/metrics", mc.Handler())
	addr := fmt.Sprintf(":%d", port)
	go http.ListenAndServe(addr, nil)
}

// TraceSpan 链路追踪跨度
type TraceSpan struct {
	Name      string
	StartTime time.Time
	Tags      map[string]string
}

// NewTraceSpan 创建追踪跨度
func NewTraceSpan(name string) *TraceSpan {
	return &TraceSpan{
		Name:      name,
		StartTime: time.Now(),
		Tags:      make(map[string]string),
	}
}

// Finish 完成追踪
func (ts *TraceSpan) Finish() time.Duration {
	return time.Since(ts.StartTime)
}

// SetTag 设置标签
func (ts *TraceSpan) SetTag(key, value string) {
	ts.Tags[key] = value
}
