// Package tui 时间管理模块
package tui

import (
	"time"
)

// getTimeWindow 获取当前的时间窗口
func (t *TUI) getTimeWindow() (start, end time.Time) {
	now := time.Now()
	elapsed := now.Sub(t.startTime)
	windowDuration := time.Duration(t.tuiConfig.MaxHistorySize) * t.tuiConfig.TimeGridInterval

	if elapsed < windowDuration {
		// 填充阶段：固定窗口，从启动时间开始
		return t.startTime, t.startTime.Add(windowDuration)
	} else {
		// 滚动阶段：跟随当前时间的移动窗口
		return now.Add(-windowDuration), now
	}
}

// timestampToX 将时间戳转换为X坐标
func (t *TUI) timestampToX(timestamp time.Time, windowStart, windowEnd time.Time, chartWidth int) int {
	windowDuration := windowEnd.Sub(windowStart)
	if windowDuration == 0 {
		return 0
	}

	offset := timestamp.Sub(windowStart)
	if offset < 0 {
		return -1 // 在窗口左边界外
	}
	if offset > windowDuration {
		return chartWidth // 在窗口右边界外
	}

	// 将时间偏移转换为X坐标
	x := int(float64(offset) / float64(windowDuration) * float64(chartWidth))
	return x
}
