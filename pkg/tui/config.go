// Package tui 配置定义
package tui

import (
	"errors"
	"time"
)

// Config TUI组件的配置结构
type Config struct {
	RefreshInterval    time.Duration // UI刷新间隔
	TimeGridInterval   time.Duration // 时间网格维护间隔
	TimeoutBufferRatio float64       // 超时缓冲比例，TUI超时 = Pinger超时 * 此比例
	MinChartWidth      int           // 最小图表宽度
	MinChartHeight     int           // 最小图表高度
	MaxHistorySize     int           // 历史缓冲区大小
	ValueBufferRatio   float64       // 值缓冲比例
	MaxChartSize       int           // 最大图表尺寸（防止极端值）
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		RefreshInterval:    200 * time.Millisecond, // 默认200ms刷新
		TimeGridInterval:   200 * time.Millisecond, // 默认200ms维护间隔
		TimeoutBufferRatio: 1.2,                    // TUI超时是Pinger超时的1.2倍
		MinChartWidth:      20,                     // 最小图表宽度
		MinChartHeight:     5,                      // 最小图表高度
		MaxHistorySize:     150,                    // 默认150个历史点
		ValueBufferRatio:   0.1,                    // 10%缓冲
		MaxChartSize:       1000,                   // 最大图表尺寸
	}
}

// GetTimeoutThreshold 计算TUI超时阈值
// 基于pinger的超时时间和缓冲比例计算
func (c *Config) GetTimeoutThreshold(pingerTimeout time.Duration) time.Duration {
	return time.Duration(float64(pingerTimeout) * c.TimeoutBufferRatio)
}

// Validate 验证配置的合理性
func (c *Config) Validate() error {
	if c.RefreshInterval <= 0 {
		return errors.New("UI刷新间隔必须大于0")
	}

	if c.RefreshInterval < 10*time.Millisecond {
		return errors.New("UI刷新间隔不能小于10ms")
	}

	if c.TimeGridInterval <= 0 {
		return errors.New("时间网格维护间隔必须大于0")
	}

	if c.TimeoutBufferRatio <= 0 {
		return errors.New("超时缓冲比例必须大于0")
	}

	if c.TimeoutBufferRatio < 1.0 {
		return errors.New("超时缓冲比例不能小于1.0")
	}

	if c.MinChartWidth <= 0 {
		return errors.New("最小图表宽度必须大于0")
	}

	if c.MinChartHeight <= 0 {
		return errors.New("最小图表高度必须大于0")
	}

	if c.MaxHistorySize <= 0 {
		return errors.New("历史缓冲区大小必须大于0")
	}

	if c.MaxHistorySize < 10 {
		return errors.New("历史缓冲区大小不能小于10")
	}

	if c.MaxHistorySize > 1000 {
		return errors.New("历史缓冲区大小不能超过1000")
	}

	if c.ValueBufferRatio < 0 {
		return errors.New("值缓冲比例不能为负数")
	}

	if c.MaxChartSize <= 0 {
		return errors.New("最大图表尺寸必须大于0")
	}

	return nil
}
