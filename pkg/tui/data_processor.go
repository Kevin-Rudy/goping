// Package tui 数据处理模块
package tui

import (
	"fmt"
	"math"
	"time"

	"github.com/Kevin-Rudy/goping/pkg/core"
)

// updateStatsWithTime 使用时间对齐更新统计信息
func (t *TUI) updateStatsWithTime(result core.PingResult) {
	t.statsMu.Lock()
	defer t.statsMu.Unlock()

	// 根据result.Identifier获取或创建core.Stats实例
	stats := t.getOrCreateStats(result.Identifier)

	// 创建数据点
	dataPoint := core.DataPoint{
		Timestamp: result.SendTime,
		Value:     result.Latency,
		Status:    core.PointSuccess,
	}

	// 如果延迟为NaN，表示超时
	if math.IsNaN(result.Latency) {
		dataPoint.Status = core.PointTimeout
	}

	// 插入数据点（按时间戳排序插入）
	t.insertDataPointByTime(stats, dataPoint)

	// 更新全局统计
	stats.PacketsSent++
	if !math.IsNaN(result.Latency) {
		stats.PacketsRecv++
		t.updateWelfordAccumulator(stats, result.Latency)

		// 更新最大最小值
		if result.Latency < stats.MinLatency {
			stats.MinLatency = result.Latency
		}
		if result.Latency > stats.MaxLatency {
			stats.MaxLatency = result.Latency
		}
	}

	// 维护历史缓冲区大小
	t.dequeueOutOfWindow(stats)

	// 更新汇总信息
	t.updateSummary(stats)
}

// insertDataPointByTime 按时间戳插入数据点到历史记录中
func (t *TUI) insertDataPointByTime(stats *core.Stats, newPoint core.DataPoint) {
	// 如果历史记录为空，直接插入
	if len(stats.History) == 0 {
		stats.History = append(stats.History, newPoint)
		return
	}

	lastPoint := stats.History[len(stats.History)-1]

	// 检查是否需要从超时状态插值到正常状态
	if lastPoint.Status == core.PointTimeout {
		switch newPoint.Status {
		case core.PointSuccess:
			t.interpolateFromTimeout(stats, lastPoint, newPoint)
		case core.PointTimeout:
			t.fillWithNaN(stats, lastPoint, newPoint)
		}
	}
	stats.History = append(stats.History, newPoint)
}

// interpolateFromTimeout 从超时状态插值到正常状态
func (t *TUI) interpolateFromTimeout(stats *core.Stats, lastPoint, newPoint core.DataPoint) {
	timeDiff := newPoint.Timestamp.Sub(lastPoint.Timestamp)
	expectedInterval := t.tuiConfig.TimeGridInterval
	steps := int(timeDiff / expectedInterval)

	t.interpolateFromTimeoutToNormal(stats, lastPoint, newPoint, steps)
	
}

// fillWithNaN 用NaN填充时间间隔
func (t *TUI) fillWithNaN(stats *core.Stats, lastPoint, newPoint core.DataPoint) {
	timeDiff := newPoint.Timestamp.Sub(lastPoint.Timestamp)
	expectedInterval := t.tuiConfig.TimeGridInterval
	steps := int(timeDiff / expectedInterval)

	for i := 1; i < steps; i++ {
		interpolatedTime := lastPoint.Timestamp.Add(time.Duration(i) * expectedInterval)
		interpolatedPoint := core.DataPoint{
			Timestamp: interpolatedTime,
			Value:     math.NaN(),
			Status:    core.PointInterpolated,
		}
		stats.History = append(stats.History, interpolatedPoint)
	}
}

// interpolateFromTimeoutToNormal 从超时状态插值到正常状态
func (t *TUI) interpolateFromTimeoutToNormal(stats *core.Stats, lastPoint, newPoint core.DataPoint, steps int) {
	stepDuration := t.tuiConfig.TimeGridInterval
	ceilingValue := t.getCurrentCeilingValue(stats)
	valueDiff := (ceilingValue - newPoint.Value) / float64(steps)

	for i := 1; i < steps; i++ {
		interpolatedTime := lastPoint.Timestamp.Add(time.Duration(i) * stepDuration)
		// 从超时逐渐过渡到正常值
		interpolatedValue := ceilingValue - float64(i)*valueDiff

		interpolatedPoint := core.DataPoint{
			Timestamp: interpolatedTime,
			Value:     interpolatedValue,
			Status:    core.PointInterpolated,
		}
		stats.History = append(stats.History, interpolatedPoint)
	}
}

// getCurrentCeilingValue 获取当前的天花板值
func (t *TUI) getCurrentCeilingValue(stats *core.Stats) float64 {
	// 从历史数据中计算动态天花板
	var maxValue float64 = 0
	for _, point := range stats.History {
		if !math.IsNaN(point.Value) && !math.IsInf(point.Value, 0) {
			if point.Value > maxValue {
				maxValue = point.Value
			}
		}
	}
	return maxValue
}

// dequeueOutOfWindow 移除窗口外的数据点
func (t *TUI) dequeueOutOfWindow(stats *core.Stats) {
	if len(stats.History) > t.tuiConfig.MaxHistorySize {
		// 移除最老的数据点
		stats.History = stats.History[len(stats.History)-t.tuiConfig.MaxHistorySize:]
	}
}

// maintainTimeGrid 维护时间网格
func (t *TUI) maintainTimeGrid() {
	now := time.Now()
	t.statsMu.Lock()
	defer t.statsMu.Unlock()

	for _, stats := range t.statsData {
		t.processPendingTimeouts(stats, now)
	}
}

// processPendingTimeouts 处理待定的超时
func (t *TUI) processPendingTimeouts(stats *core.Stats, now time.Time) {
	// 检查历史记录中是否有pending状态的点需要转换为超时
	for i := len(stats.History) - 1; i >= 0; i-- {
		point := &stats.History[i]
		if point.Status == core.PointPending {
			if now.Sub(point.Timestamp) > t.timeoutThreshold {
				point.Status = core.PointTimeout
				point.Value = math.NaN()
			}
		}
	}
}

// updateWelfordAccumulator 使用Welford在线算法更新统计累加器
func (t *TUI) updateWelfordAccumulator(stats *core.Stats, newValue float64) {
	stats.WelfordCount++
	delta := newValue - stats.WelfordMean
	stats.WelfordMean += delta / float64(stats.WelfordCount)
	delta2 := newValue - stats.WelfordMean
	stats.WelfordM2 += delta * delta2
}

// updateSummary 更新汇总统计信息
func (t *TUI) updateSummary(stats *core.Stats) {
	summary := make(map[string]string)

	// 超时次数
	timeouts := stats.PacketsSent - stats.PacketsRecv
	summary["t/o"] = fmt.Sprintf("%d", timeouts)

	// 丢包率
	var lossRate float64
	if stats.PacketsSent > 0 {
		lossRate = float64(stats.PacketsSent-stats.PacketsRecv) / float64(stats.PacketsSent) * 100
	}
	summary["丢包率"] = fmt.Sprintf("%.1f%%", lossRate)

	// 发送/接收合并显示
	summary["发送/接收"] = fmt.Sprintf("%d/%d", stats.PacketsSent, stats.PacketsRecv)

	// 延迟统计
	if stats.PacketsRecv > 0 {
		summary["平均延迟"] = formatLatency(stats.WelfordMean)
		summary["最小延迟"] = formatLatency(stats.MinLatency)
		summary["最大延迟"] = formatLatency(stats.MaxLatency)
	} else {
		summary["平均延迟"] = "N/A"
		summary["最小延迟"] = "N/A"
		summary["最大延迟"] = "N/A"
	}

	stats.Summary = summary
}
