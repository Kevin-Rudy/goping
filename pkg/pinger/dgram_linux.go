//go:build linux

// Package pinger - Linux非特权模式实现
// 使用SOCK_DGRAM类型的ICMP套接字，仅适用于Linux系统
package pinger

import (
	"math"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/Kevin-Rudy/goping/pkg/core"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// dgramPinger Linux非特权模式的ping实现
type dgramPinger struct {
	*basePinger
	sock4 int // IPv4 DGRAM socket
}

// newLinuxDgramPinger 创建Linux非特权模式的pinger实例
func newLinuxDgramPinger(targets []string, config *Config) (core.DataSource, error) {
	p := &dgramPinger{
		basePinger: newBasePinger(targets, config),
	}

	// 创建DGRAM ICMP socket
	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_ICMP)
	if err != nil {
		return nil, err
	}

	// 设置socket到结构体中
	p.sock4 = sock

	return p, nil
}

// Start 实现core.DataSource接口，启动ping操作
func (p *dgramPinger) Start() {
	p.setRunning(true)

	// 为每个目标启动一个goroutine
	for _, target := range p.targets {
		p.wg.Add(1)
		go p.pingTarget(target)
	}
}

// pingTarget 对单个目标进行ping操作
func (p *dgramPinger) pingTarget(target string) {
	defer p.wg.Done()

	// 解析目标地址（地址已在NewPinger中预验证，此处失败属于临时网络问题）
	dst, err := net.ResolveIPAddr(p.config.GetIPProtocol(), target)
	if err != nil {
		// 发送错误结果，但不退出（可能是临时DNS问题）
		p.sendPingResult(target, math.NaN())
		return
	}

	seq := 0
	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			seq++
			p.sendPing(dst, seq, target)
		}
	}
}

// sendPing 发送单个ping包并等待回复
func (p *dgramPinger) sendPing(dst *net.IPAddr, seq int, target string) {
	// 构建ICMP消息
	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  seq,
			Data: []byte("goping"),
		},
	}

	// 序列化ICMP消息
	data, err := msg.Marshal(nil)
	if err != nil {
		p.sendPingResult(target, math.NaN())
		return
	}

	// 构建sockaddr_in结构
	sockaddr := &syscall.SockaddrInet4{}
	copy(sockaddr.Addr[:], dst.IP.To4())

	// 记录发送时间
	startTime := time.Now()

	// 发送数据
	err = syscall.Sendto(p.sock4, data, 0, sockaddr)
	if err != nil {
		p.sendPingResult(target, math.NaN())
		return
	}

	// 设置接收超时
	tv := syscall.Timeval{
		Sec:  int64(p.config.Timeout.Seconds()),
		Usec: int64(p.config.Timeout.Nanoseconds()/1000) % 1000000,
	}

	err = syscall.SetsockoptTimeval(p.sock4, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)
	if err != nil {
		return
	}

	// 等待回复
	reply := make([]byte, 1500)
	for {
		n, from, err := syscall.Recvfrom(p.sock4, reply, 0)
		if err != nil {
			// 超时或其他错误
			p.sendPingResult(target, math.NaN())
			return
		}

		// 检查来源地址
		if fromAddr, ok := from.(*syscall.SockaddrInet4); ok {
			fromIP := net.IPv4(fromAddr.Addr[0], fromAddr.Addr[1], fromAddr.Addr[2], fromAddr.Addr[3])
			if !fromIP.Equal(dst.IP) {
				continue
			}
		}

		// 解析ICMP回复
		replyMsg, err := icmp.ParseMessage(int(ipv4.ICMPTypeEchoReply), reply[:n])
		if err != nil {
			continue
		}

		// 验证回复的ID和序列号
		if echo, ok := replyMsg.Body.(*icmp.Echo); ok {
			if echo.ID == (os.Getpid()&0xffff) && echo.Seq == seq {
				// 计算RTT
				rtt := time.Since(startTime)

				// 发送延迟结果（转换为毫秒）
				latencyMs := float64(rtt.Nanoseconds()) / 1e6
				p.sendPingResult(target, latencyMs)
				return
			}
		}
	}
}

// Stop 停止Linux DGRAM模式的pinger
func (p *dgramPinger) Stop() {
	// 调用基础的停止方法
	p.basePinger.Stop()

	// 关闭socket
	if p.sock4 > 0 {
		syscall.Close(p.sock4)
	}
}
