// Package pinger 选项模式支持
package pinger

import (
	"time"

	"github.com/Kevin-Rudy/goping/pkg/core"
)

// Option 配置选项函数类型
type Option func(*Config)

// WithIPVersion 设置IP版本
func WithIPVersion(version int) Option {
	return func(c *Config) {
		c.IPVersion = version
	}
}

// WithInterval 设置ping间隔
func WithInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.Interval = interval
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithBufferSize 设置缓冲区大小
func WithBufferSize(size int) Option {
	return func(c *Config) {
		c.BufferSize = size
	}
}

// NewPingerWithOptions 使用选项模式创建Pinger
func NewPingerWithOptions(targets []string, opts ...Option) (core.DataSource, error) {
	config := DefaultConfig()

	// 应用所有选项
	for _, opt := range opts {
		opt(config)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 验证目标地址
	if err := config.ValidateTargets(targets); err != nil {
		return nil, err
	}

	// 调用已经重构的NewPinger函数
	return NewPinger(targets, config)
}
