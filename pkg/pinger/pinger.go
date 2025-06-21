// Package pinger 实现了core.DataSource接口，提供ping功能
// 根据操作系统和用户权限自动选择最合适的底层实现
package pinger

import (
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/Kevin-Rudy/goping/pkg/core"
)

// basePinger 定义了所有pinger实现的基本结构
type basePinger struct {
	targets   []string             // ping目标列表
	config    *Config              // 配置信息（替代原来的全局变量）
	dataChan  chan core.PingResult // 数据输出通道
	stopChan  chan struct{}        // 停止信号通道
	wg        sync.WaitGroup       // 等待组，用于优雅关闭
	running   bool                 // 运行状态
	runningMu sync.RWMutex         // 保护running状态的锁
}

// newBasePinger 创建基础pinger结构
func newBasePinger(targets []string, config *Config) *basePinger {
	return &basePinger{
		targets:  targets,
		config:   config,
		dataChan: make(chan core.PingResult, config.BufferSize), // 使用配置的缓冲区大小
		stopChan: make(chan struct{}),
	}
}

// DataStream 实现core.DataSource接口
func (bp *basePinger) DataStream() <-chan core.PingResult {
	return bp.dataChan
}

// Stop 实现core.DataSource接口
func (bp *basePinger) Stop() {
	bp.runningMu.Lock()
	if !bp.running {
		bp.runningMu.Unlock()
		return
	}
	bp.running = false
	bp.runningMu.Unlock()

	// 发送停止信号
	close(bp.stopChan)

	// 等待所有goroutine结束
	bp.wg.Wait()

	// 关闭数据通道
	close(bp.dataChan)
}

// isRunning 检查是否正在运行
func (bp *basePinger) isRunning() bool {
	bp.runningMu.RLock()
	defer bp.runningMu.RUnlock()
	return bp.running
}

// setRunning 设置运行状态
func (bp *basePinger) setRunning(running bool) {
	bp.runningMu.Lock()
	defer bp.runningMu.Unlock()
	bp.running = running
}

// sendPingResult 发送ping结果到数据通道
func (bp *basePinger) sendPingResult(target string, latency float64) {
	bp.sendPingResultWithTime(target, latency, time.Now(), time.Now())
}

// sendPingResultWithTime 发送带时间戳的ping结果到数据通道
func (bp *basePinger) sendPingResultWithTime(target string, latency float64, sendTime, receiveTime time.Time) {
	if !bp.isRunning() {
		return
	}

	result := core.PingResult{
		Identifier:  target,
		Latency:     latency,
		SendTime:    sendTime,
		ReceiveTime: receiveTime,
	}

	select {
	case bp.dataChan <- result:
		// 成功发送
	case <-bp.stopChan:
		// 停止信号，不再发送
		return
	default:
		// 通道满了，丢弃这个数据点
		// 在高频ping场景中这是可以接受的
	}
}

// NewPinger 创建新的Pinger实例
func NewPinger(targets []string, config *Config) (core.DataSource, error) {
	if len(targets) == 0 {
		return nil, errors.New("必须指定至少一个目标")
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 验证目标地址
	if err := config.ValidateTargets(targets); err != nil {
		return nil, err
	}

	// 获取当前平台的能力实现
	platform := getPlatformCapability()

	// 优先尝试特权模式（所有平台统一用raw socket）
	if platform.hasPrivilegedAccess() {
		return platform.createPrivilegedPinger(targets, config)
	}

	// 降级到非特权模式（各平台不同的实现）
	return platform.createUnprivilegedPinger(targets, config)
}

// GetSystemInfo 获取完整的系统信息
// 返回操作系统名称、权限状态和实现类型
func GetSystemInfo() (osName, privilegeStatus, implementationType string) {
	// 获取操作系统名称
	switch runtime.GOOS {
	case "windows":
		osName = "Windows"
	case "linux":
		osName = "Linux"
	case "darwin":
		osName = "macOS"
	default:
		osName = runtime.GOOS
	}

	// 获取当前平台能力并检查权限状态
	platform := getPlatformCapability()
	hasPriv := platform.hasPrivilegedAccess()

	switch runtime.GOOS {
	case "windows":
		if hasPriv {
			privilegeStatus = "管理员模式 (Raw Socket)"
			implementationType = "Raw Socket"
		} else {
			privilegeStatus = "普通用户模式 (Windows API)"
			implementationType = "Windows ICMP API"
		}
	case "linux":
		if hasPriv {
			privilegeStatus = "特权模式 (Raw Socket)"
			implementationType = "Linux Raw Socket"
		} else {
			privilegeStatus = "非特权模式 (DGRAM Socket)"
			implementationType = "Linux DGRAM Socket"
		}
	case "darwin":
		if hasPriv {
			privilegeStatus = "特权模式 (Root权限)"
			implementationType = "macOS Raw Socket"
		} else {
			privilegeStatus = "权限不足 (需要sudo)"
			implementationType = "macOS Raw Socket (未启用)"
		}
	default:
		if hasPriv {
			privilegeStatus = "特权模式"
			implementationType = "通用Raw Socket"
		} else {
			privilegeStatus = "权限不足"
			implementationType = "通用Raw Socket (需要提权)"
		}
	}

	return
}

// GetOSName 获取操作系统名称
func GetOSName() string {
	osName, _, _ := GetSystemInfo()
	return osName
}

// GetPrivilegeStatus 获取权限状态描述
func GetPrivilegeStatus() string {
	_, privilegeStatus, _ := GetSystemInfo()
	return privilegeStatus
}

// GetImplementationType 获取ping实现类型描述
func GetImplementationType() string {
	_, _, implementationType := GetSystemInfo()
	return implementationType
}

// HasPrivilegedAccess 检查是否有特权访问能力
func HasPrivilegedAccess() bool {
	platform := getPlatformCapability()
	return platform.hasPrivilegedAccess()
}
