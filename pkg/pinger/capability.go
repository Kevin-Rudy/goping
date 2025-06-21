// Package pinger - 平台能力接口定义
// 定义了跨平台的权限检测和pinger创建接口
package pinger

import (
	"github.com/Kevin-Rudy/goping/pkg/core"
)

// platformCapability 定义平台能力接口
// 每个平台实现此接口来提供权限检测和pinger创建功能
type platformCapability interface {
	// hasPrivilegedAccess 检查是否有特权访问能力
	// Windows: 检查管理员权限
	// Linux: 检查CAP_NET_RAW或root权限
	// macOS: 检查root权限
	hasPrivilegedAccess() bool

	// createPrivilegedPinger 创建特权模式pinger
	// 所有平台统一使用raw socket实现
	createPrivilegedPinger(targets []string, config *Config) (core.DataSource, error)

	// createUnprivilegedPinger 创建非特权模式pinger
	// Windows: 使用Windows API
	// Linux: 使用DGRAM socket
	// macOS: 返回错误要求sudo
	createUnprivilegedPinger(targets []string, config *Config) (core.DataSource, error)
}
