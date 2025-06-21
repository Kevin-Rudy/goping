//go:build windows

// Package pinger - Windows非特权模式实现
// 使用Icmp.dll系统调用，适用于Windows系统
package pinger

import (
	"math"
	"net"
	"syscall"
	"time"
	"unsafe"

	"github.com/Kevin-Rudy/goping/pkg/core"
	"golang.org/x/sys/windows"
)

var (
	// 加载Icmp.dll库
	icmpDLL = windows.NewLazyDLL("Icmp.dll")

	// 获取函数地址
	icmpCreateFile  = icmpDLL.NewProc("IcmpCreateFile")
	icmpCloseHandle = icmpDLL.NewProc("IcmpCloseHandle")
	icmpSendEcho    = icmpDLL.NewProc("IcmpSendEcho")
)

// ICMP_ECHO_REPLY Windows ICMP回复结构体
type ICMP_ECHO_REPLY struct {
	Address       uint32
	Status        uint32
	RoundTripTime uint32
	DataSize      uint16
	Reserved      uint16
	Data          uintptr
	Options       ICMP_OPTIONS
}

// ICMP_OPTIONS Windows ICMP选项结构体
type ICMP_OPTIONS struct {
	Ttl         uint8
	Tos         uint8
	Flags       uint8
	OptionsSize uint8
	OptionsData uintptr
}

// windowsPinger Windows非特权模式的ping实现
type windowsPinger struct {
	*basePinger
	icmpHandle syscall.Handle // ICMP句柄
}

// newWindowsPinger 创建Windows非特权模式的pinger实例
func newWindowsPinger(targets []string, config *Config) (core.DataSource, error) {
	p := &windowsPinger{
		basePinger: newBasePinger(targets, config),
	}

	// 创建ICMP句柄
	ret, _, err := icmpCreateFile.Call()
	if ret == 0 || ret == uintptr(syscall.InvalidHandle) {
		return nil, err
	}

	p.icmpHandle = syscall.Handle(ret)
	return p, nil
}

// Start 实现core.DataSource接口，启动ping操作
func (p *windowsPinger) Start() {
	p.setRunning(true)

	// 为每个目标启动一个goroutine
	for _, target := range p.targets {
		p.wg.Add(1)
		go p.pingTarget(target)
	}
}

// pingTarget 对单个目标进行ping操作
func (p *windowsPinger) pingTarget(target string) {
	defer p.wg.Done()

	// 解析目标地址
	dst, err := net.ResolveIPAddr(p.config.GetIPProtocol(), target)
	if err != nil {
		return
	}

	// 将IP地址转换为32位整数（网络字节序）
	ip := dst.IP.To4()
	destAddr := uint32(ip[0]) | (uint32(ip[1]) << 8) | (uint32(ip[2]) << 16) | (uint32(ip[3]) << 24)

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.sendPing(destAddr, target)
		}
	}
}

// sendPing 发送单个ping包
func (p *windowsPinger) sendPing(destAddr uint32, target string) {
	// 准备发送数据
	sendData := []byte("goping")

	// 准备接收缓冲区
	// 需要足够大的缓冲区来存储ICMP_ECHO_REPLY结构和数据
	replySize := unsafe.Sizeof(ICMP_ECHO_REPLY{}) + uintptr(len(sendData)) + 8
	replyBuffer := make([]byte, replySize)

	// 设置超时（毫秒）
	timeoutMs := uint32(p.config.Timeout.Milliseconds())

	// 记录发送时间
	sendTime := time.Now()

	// 调用IcmpSendEcho
	ret, _, _ := icmpSendEcho.Call(
		uintptr(p.icmpHandle),                    // ICMP句柄
		uintptr(destAddr),                        // 目标IP地址
		uintptr(unsafe.Pointer(&sendData[0])),    // 发送数据
		uintptr(len(sendData)),                   // 发送数据长度
		0,                                        // ICMP选项（NULL）
		uintptr(unsafe.Pointer(&replyBuffer[0])), // 接收缓冲区
		uintptr(len(replyBuffer)),                // 接收缓冲区大小
		uintptr(timeoutMs),                       // 超时时间（毫秒）
	)

	receiveTime := time.Now()

	if ret == 0 {
		// 请求失败或超时 - 发送NaN作为延迟
		p.sendPingResultWithTime(target, math.NaN(), sendTime, receiveTime)
		return
	}

	// 解析回复
	reply := (*ICMP_ECHO_REPLY)(unsafe.Pointer(&replyBuffer[0]))

	if reply.Status == 0 { // IP_SUCCESS
		// 成功收到回复
		// 使用Windows API返回的往返时间，或计算时间差
		var rtt time.Duration
		if reply.RoundTripTime > 0 {
			rtt = time.Duration(reply.RoundTripTime) * time.Millisecond
		} else {
			rtt = receiveTime.Sub(sendTime)
		}

		// 发送延迟结果（转换为毫秒）
		latencyMs := float64(rtt.Nanoseconds()) / 1e6
		p.sendPingResultWithTime(target, latencyMs, sendTime, receiveTime)
	} else {
		// 回复有错误状态 - 发送NaN作为延迟
		p.sendPingResultWithTime(target, math.NaN(), sendTime, receiveTime)
	}
}

// Stop 停止Windows模式的pinger
func (p *windowsPinger) Stop() {
	// 调用基础的停止方法
	p.basePinger.Stop()

	// 关闭ICMP句柄
	if p.icmpHandle != syscall.InvalidHandle {
		icmpCloseHandle.Call(uintptr(p.icmpHandle))
		p.icmpHandle = syscall.InvalidHandle
	}
}

// checkWindowsAdmin 检查是否具有Windows管理员权限
func checkWindowsAdmin() bool {
	var sid *windows.SID

	// 获取管理员组的SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	// 获取当前进程的token
	token := windows.Token(0)

	// 检查是否是管理员组成员
	isMember, err := token.IsMember(sid)
	if err != nil {
		return false
	}

	return isMember
}
