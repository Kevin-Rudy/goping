//go:build darwin

package pinger

import (
	"errors"
	"os"

	"github.com/Kevin-Rudy/goping/pkg/core"
)

// darwinCapability macOS平台能力实现
type darwinCapability struct{}

// hasPrivilegedAccess 检查macOS root权限
func (d *darwinCapability) hasPrivilegedAccess() bool {
	return checkDarwinRoot()
}

// createPrivilegedPinger 创建特权模式pinger（使用raw socket）
func (d *darwinCapability) createPrivilegedPinger(targets []string, config *Config) (core.DataSource, error) {
	return newPrivilegedPinger(targets, config)
}

// createUnprivilegedPinger macOS非特权模式直接报错要求sudo
func (d *darwinCapability) createUnprivilegedPinger(targets []string, config *Config) (core.DataSource, error) {
	return nil, errors.New("macOS需要root权限才能进行ping操作，请使用sudo运行")
}

// checkDarwinRoot 检查macOS系统的root权限
func checkDarwinRoot() bool {
	return os.Geteuid() == 0
}

// getPlatformCapability 获取macOS平台的能力实现
func getPlatformCapability() platformCapability {
	return &darwinCapability{}
}
