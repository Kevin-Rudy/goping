package core

import (
	"math"
	"testing"
	"time"
)

// TestNewStats 测试NewStats构造函数
func TestNewStats(t *testing.T) {
	stats := NewStats("test.com")

	if stats.Identifier != "test.com" {
		t.Errorf("Expected identifier 'test.com', got '%s'", stats.Identifier)
	}

	if len(stats.History) != 0 {
		t.Errorf("Expected 0 history values, got %d", len(stats.History))
	}

	if stats.PacketsSent != 0 {
		t.Errorf("Expected 0 packets sent, got %d", stats.PacketsSent)
	}

	if stats.PacketsRecv != 0 {
		t.Errorf("Expected 0 packets received, got %d", stats.PacketsRecv)
	}

	if stats.WelfordCount != 0 {
		t.Errorf("Expected 0 Welford count, got %d", stats.WelfordCount)
	}

	if !math.IsInf(stats.MinLatency, 1) {
		t.Errorf("Expected MinLatency to be +Inf, got %f", stats.MinLatency)
	}

	if !math.IsInf(stats.MaxLatency, -1) {
		t.Errorf("Expected MaxLatency to be -Inf, got %f", stats.MaxLatency)
	}
}

// TestPingResult 测试PingResult结构体
func TestPingResult(t *testing.T) {
	// 测试正常延迟
	result1 := PingResult{
		Identifier: "test.com",
		Latency:    12.5,
	}

	if result1.Identifier != "test.com" {
		t.Errorf("Expected identifier 'test.com', got '%s'", result1.Identifier)
	}

	if result1.Latency != 12.5 {
		t.Errorf("Expected latency 12.5, got %f", result1.Latency)
	}

	// 测试超时（NaN）
	result2 := PingResult{
		Identifier: "timeout.com",
		Latency:    math.NaN(),
	}

	if result2.Identifier != "timeout.com" {
		t.Errorf("Expected identifier 'timeout.com', got '%s'", result2.Identifier)
	}

	if !math.IsNaN(result2.Latency) {
		t.Errorf("Expected latency to be NaN, got %f", result2.Latency)
	}
}

// mockDataSource 模拟数据源，用于测试
type mockDataSource struct {
	dataChan chan PingResult
	started  bool
	stopped  bool
}

func newMockDataSource() *mockDataSource {
	return &mockDataSource{
		dataChan: make(chan PingResult, 10),
	}
}

func (m *mockDataSource) DataStream() <-chan PingResult {
	return m.dataChan
}

func (m *mockDataSource) Start() {
	m.started = true
	go func() {
		for i := 0; i < 3; i++ {
			if m.stopped {
				break
			}
			// 发送正常延迟结果
			m.dataChan <- PingResult{
				Identifier: "mock.test",
				Latency:    float64(i + 1),
			}
			time.Sleep(10 * time.Millisecond)
		}
		// 发送一个超时结果
		if !m.stopped {
			m.dataChan <- PingResult{
				Identifier: "mock.test",
				Latency:    math.NaN(),
			}
		}
		close(m.dataChan)
	}()
}

func (m *mockDataSource) Stop() {
	m.stopped = true
}

// TestDataSourceInterface 测试DataSource接口
func TestDataSourceInterface(t *testing.T) {
	mock := newMockDataSource()

	// 测试初始状态
	if mock.started {
		t.Error("DataSource should not be started initially")
	}

	// 测试启动
	mock.Start()
	if !mock.started {
		t.Error("DataSource should be started after Start() call")
	}

	// 测试数据流
	dataChan := mock.DataStream()
	if dataChan == nil {
		t.Error("DataStream() should return a valid channel")
	}

	// 接收一些数据
	receivedCount := 0
	normalCount := 0
	timeoutCount := 0
	timeout := time.After(200 * time.Millisecond)

	for {
		select {
		case result, ok := <-dataChan:
			if !ok {
				// 通道关闭，跳出循环
				goto done
			}
			receivedCount++
			if result.Identifier != "mock.test" {
				t.Errorf("Expected identifier 'mock.test', got '%s'", result.Identifier)
			}
			if math.IsNaN(result.Latency) {
				timeoutCount++
			} else {
				normalCount++
			}
		case <-timeout:
			goto done
		}
	}

done:
	if receivedCount == 0 {
		t.Error("Should receive at least one result from DataStream")
	}

	if normalCount == 0 {
		t.Error("Should receive at least one normal ping result")
	}

	if timeoutCount == 0 {
		t.Error("Should receive at least one timeout result")
	}

	// 测试停止
	mock.Stop()
	if !mock.stopped {
		t.Error("DataSource should be stopped after Stop() call")
	}
}
