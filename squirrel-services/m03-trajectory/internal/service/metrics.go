package service

import (
	"sync"
	"time"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	mu              sync.RWMutex
	processTimes    []time.Duration
	requestCount    int64
	errorCount      int64
	avgProcessTime  time.Duration
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		processTimes: make([]time.Duration, 0, 1000),
	}
}

func (m *MetricsCollector) RecordProcessingTime(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.processTimes = append(m.processTimes, d)
	m.requestCount++
	
	// 只保留最近1000条
	if len(m.processTimes) > 1000 {
		m.processTimes = m.processTimes[len(m.processTimes)-1000:]
	}
	
	// 更新平均
	var total time.Duration
	for _, t := range m.processTimes {
		total += t
	}
	m.avgProcessTime = total / time.Duration(len(m.processTimes))
}

func (m *MetricsCollector) RecordError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCount++
}

func (m *MetricsCollector) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return Stats{
		AvgProcessTime: m.avgProcessTime,
		RequestCount:   m.requestCount,
		ErrorCount:     m.errorCount,
	}
}

type Stats struct {
	AvgProcessTime time.Duration
	RequestCount   int64
	ErrorCount     int64
}
