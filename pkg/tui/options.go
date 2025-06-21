// Package tui 选项模式支持
package tui

import (
	"time"
)

// Option TUI配置选项函数类型
type Option func(*Config)

// WithRefreshInterval 设置UI刷新间隔
func WithRefreshInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.RefreshInterval = interval
	}
}

// WithTimeGridInterval 设置时间网格维护间隔
func WithTimeGridInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.TimeGridInterval = interval
	}
}

// WithTimeoutBufferRatio 设置超时缓冲比例
func WithTimeoutBufferRatio(ratio float64) Option {
	return func(c *Config) {
		c.TimeoutBufferRatio = ratio
	}
}

// WithChartSize 设置图表尺寸
func WithChartSize(width, height int) Option {
	return func(c *Config) {
		c.MinChartWidth = width
		c.MinChartHeight = height
	}
}

// WithHistorySize 设置历史缓冲区大小
func WithHistorySize(size int) Option {
	return func(c *Config) {
		c.MaxHistorySize = size
	}
}

// WithDefaultCeiling 设置默认天花板值
func WithDefaultCeiling(ceiling float64) Option {
	return func(c *Config) {
		c.DefaultCeiling = ceiling
	}
}

// WithValueBufferRatio 设置值缓冲比例
func WithValueBufferRatio(ratio float64) Option {
	return func(c *Config) {
		c.ValueBufferRatio = ratio
	}
}

// NewConfigWithOptions 使用选项模式创建TUI配置
func NewConfigWithOptions(opts ...Option) *Config {
	config := DefaultConfig()

	// 应用所有选项
	for _, opt := range opts {
		opt(config)
	}

	return config
}
