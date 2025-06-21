//go:build linux

package pinger

import (
	"net"
	"os"

	"github.com/Kevin-Rudy/goping/pkg/core"
)

// linuxCapability Linux平台能力实现
type linuxCapability struct{}

// hasPrivilegedAccess 检查Linux权限（CAP_NET_RAW或root）
func (l *linuxCapability) hasPrivilegedAccess() bool {
	return checkLinuxCapNetRaw()
}

// createPrivilegedPinger 创建特权模式pinger（使用raw socket）
func (l *linuxCapability) createPrivilegedPinger(targets []string, config *Config) (core.DataSource, error) {
	return newPrivilegedPinger(targets, config)
}

// createUnprivilegedPinger 创建Linux DGRAM pinger
func (l *linuxCapability) createUnprivilegedPinger(targets []string, config *Config) (core.DataSource, error) {
	return newLinuxDgramPinger(targets, config)
}

// checkLinuxCapNetRaw 检查Linux系统的CAP_NET_RAW权限或root权限
func checkLinuxCapNetRaw() bool {
	// 首先检查是否为root用户
	if os.Geteuid() == 0 {
		return true
	}

	// 尝试创建原始套接字来检测CAP_NET_RAW权限
	conn, err := net.Dial("ip4:icmp", "127.0.0.1")
	if err != nil {
		return false
	}

	// 使用defer确保连接一定会被关闭，避免资源泄露
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			// 记录关闭错误，但不影响权限检测结果
			// 在实际部署中可以考虑添加日志记录
		}
	}()

	return true
}

// getPlatformCapability 获取Linux平台的能力实现
func getPlatformCapability() platformCapability {
	return &linuxCapability{}
}
