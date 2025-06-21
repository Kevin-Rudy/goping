//go:build windows

package pinger

import (
	"github.com/Kevin-Rudy/goping/pkg/core"
)

// windowsCapability Windows平台能力实现
type windowsCapability struct{}

// hasPrivilegedAccess 检查Windows管理员权限
func (w *windowsCapability) hasPrivilegedAccess() bool {
	return checkWindowsAdmin()
}

// createPrivilegedPinger 创建特权模式pinger（使用raw socket）
func (w *windowsCapability) createPrivilegedPinger(targets []string, config *Config) (core.DataSource, error) {
	return newPrivilegedPinger(targets, config)
}

// createUnprivilegedPinger 创建Windows API pinger
func (w *windowsCapability) createUnprivilegedPinger(targets []string, config *Config) (core.DataSource, error) {
	return newWindowsPinger(targets, config)
}

// getPlatformCapability 获取Windows平台的能力实现
func getPlatformCapability() platformCapability {
	return &windowsCapability{}
}
