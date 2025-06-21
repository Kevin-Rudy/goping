// Package core 定义了监控框架的核心接口和数据结构
// 这些接口保证了TUI与具体执行器的完全解耦
package core

import (
	"math"
	"time"
)

// PingResult 表示单次ping操作的原子结果
// 用于在数据源和TUI之间传递单次ping的结果
type PingResult struct {
	Identifier  string    // 目标标识符（如IP地址、域名等）
	Latency     float64   // 延迟(ms)。超时或失败时为 math.NaN()
	SendTime    time.Time // ping发送时间，用于时间对齐
	ReceiveTime time.Time // ping接收时间，用于精确计算延迟
}

// PointStatus 表示数据点的状态
type PointStatus int

const (
	PointPending      PointStatus = iota // 正在等待响应
	PointSuccess                         // 成功收到响应
	PointTimeout                         // 超时
	PointInterpolated                    // 插值点
)

// DataPoint 表示带时间戳和状态的数据点
type DataPoint struct {
	Timestamp time.Time   // 数据点的时间戳（基于发送时间）
	Value     float64     // 延迟值，NaN表示超时或待定
	Status    PointStatus // 数据点状态
}

// Stats 表示来自监控数据源的统计数据
// 由TUI层管理，区分用于显示的近期历史和用于统计的全局累加器
type Stats struct {
	// Identifier 数据源的唯一标识符（如IP地址、URL等）
	Identifier string

	// --- 用于图表显示的近期历史 ---
	History []DataPoint // 由TUI管理的、有长度上限的滚动缓冲区，支持时间对齐

	// --- 用于表格统计的全局累加器 ---
	PacketsSent int // 总发包数
	PacketsRecv int // 总收包数

	// Welford's Online Algorithm 所需的累加器
	WelfordCount int64   // Welford算法的样本计数
	WelfordMean  float64 // Welford算法的均值
	WelfordM2    float64 // Welford算法的M2值

	// 全局最大/最小值
	MinLatency float64 // 全局最小延迟
	MaxLatency float64 // 全局最大延迟

	// --- 用于最终显示的格式化数据 ---
	Summary map[string]string // 汇总统计信息的键值对映射
}

// NewStats 创建一个新的Stats实例
func NewStats(identifier string) *Stats {
	return &Stats{
		Identifier:   identifier,
		History:      make([]DataPoint, 0),
		PacketsSent:  0,
		PacketsRecv:  0,
		WelfordCount: 0,
		WelfordMean:  0.0,
		WelfordM2:    0.0,
		MinLatency:   math.Inf(1),  // 初始化为正无穷
		MaxLatency:   math.Inf(-1), // 初始化为负无穷
		Summary:      make(map[string]string),
	}
}

// DataSource 定义了数据源的标准接口
// 任何监控执行器（如Pinger、下载速度测试器等）都应该实现这个接口
type DataSource interface {
	// DataStream 返回一个只读通道，用于接收实时的ping结果
	// 实现者应该在独立的goroutine中持续发送PingResult数据到这个通道
	DataStream() <-chan PingResult

	// Start 启动数据收集
	// 这个方法应该是非阻塞的，实际的数据收集工作在后台goroutine中进行
	Start()

	// Stop 停止数据收集并清理资源
	// 调用此方法后，DataStream()返回的通道应该被关闭
	// 所有相关的goroutine应该优雅地退出
	Stop()
}
