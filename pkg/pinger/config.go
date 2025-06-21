// Package pinger 配置定义
package pinger

import (
	"errors"
	"fmt"
	"net"
	"time"
)

// Config pinger组件的配置结构
type Config struct {
	IPVersion  int           // IP版本，4或6
	Interval   time.Duration // ping间隔时间
	Timeout    time.Duration // ping超时时间
	BufferSize int           // 数据通道缓冲区大小
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		IPVersion:  4,                      // 默认IPv4
		Interval:   200 * time.Millisecond, // 默认200ms间隔
		Timeout:    3 * time.Second,        // 默认3秒超时
		BufferSize: 100,                    // 默认100缓冲区大小
	}
}

// GetIPProtocol 获取IP协议字符串，用于网络操作
func (c *Config) GetIPProtocol() string {
	if c.IPVersion == 6 {
		return "ip6"
	}
	return "ip4"
}

// ValidateTargets 验证目标地址是否符合当前IP版本配置
func (c *Config) ValidateTargets(targets []string) error {
	protocol := c.GetIPProtocol()

	for _, target := range targets {
		if target == "" {
			return errors.New("目标地址不能为空")
		}

		_, err := net.ResolveIPAddr(protocol, target)
		if err != nil {
			return fmt.Errorf("无法将 '%s' 解析为IPv%d地址: %v", target, c.IPVersion, err)
		}
	}
	return nil
}

// Validate 验证配置的合理性
func (c *Config) Validate() error {
	if c.IPVersion != 4 && c.IPVersion != 6 {
		return errors.New("IP版本必须是4或6")
	}

	if c.Interval <= 0 {
		return errors.New("ping间隔必须大于0")
	}

	if c.Interval < 10*time.Millisecond {
		return errors.New("ping间隔不能小于10ms")
	}

	if c.Timeout <= 0 {
		return errors.New("超时时间必须大于0")
	}

	if c.Timeout < 100*time.Millisecond {
		return errors.New("超时时间不能小于100ms")
	}

	if c.BufferSize <= 0 {
		return errors.New("缓冲区大小必须大于0")
	}

	return nil
}
