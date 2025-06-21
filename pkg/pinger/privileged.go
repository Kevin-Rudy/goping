// Package pinger - 特权模式实现
// 使用原始套接字，需要管理员/root权限，但支持所有操作系统
package pinger

import (
	"math"
	"net"
	"os"
	"time"

	"github.com/Kevin-Rudy/goping/pkg/core"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// privilegedPinger 特权模式的ping实现
type privilegedPinger struct {
	*basePinger
}

// newPrivilegedPinger 创建特权模式的pinger实例
func newPrivilegedPinger(targets []string, config *Config) (core.DataSource, error) {
	p := &privilegedPinger{
		basePinger: newBasePinger(targets, config),
	}
	return p, nil
}

// Start 实现core.DataSource接口，启动ping操作
func (p *privilegedPinger) Start() {
	p.setRunning(true)

	// 为每个目标启动一个goroutine
	for _, target := range p.targets {
		p.wg.Add(1)
		go p.pingTarget(target)
	}
}

// pingTarget 对单个目标进行ping操作
func (p *privilegedPinger) pingTarget(target string) {
	defer p.wg.Done()

	// 解析目标地址（地址已在NewPinger中预验证，此处失败属于临时网络问题）
	dst, err := net.ResolveIPAddr(p.config.GetIPProtocol(), target)
	if err != nil {
		// 发送错误结果，但不退出（可能是临时DNS问题）
		p.sendPingResult(target, math.NaN())
		return
	}

	// 创建原始套接字
	protocol := "ip4:icmp"
	if p.config.IPVersion == 6 {
		protocol = "ip6:ipv6-icmp"
	}
	conn, err := net.Dial(protocol, dst.String())
	if err != nil {
		// 连接失败，发送错误结果
		p.sendPingResult(target, math.NaN())
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	seq := 0

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			seq++
			p.sendPing(conn, target, seq)
		}
	}
}

// sendPing 发送单个ping包
func (p *privilegedPinger) sendPing(conn net.Conn, target string, seq int) {
	// 创建ICMP包
	icmpPacket := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  seq,
			Data: []byte("goping"),
		},
	}

	// 序列化ICMP包
	data, err := icmpPacket.Marshal(nil)
	if err != nil {
		p.sendPingResult(target, math.NaN())
		return
	}

	// 记录发送时间
	startTime := time.Now()

	// 设置超时
	conn.SetDeadline(time.Now().Add(p.config.Timeout))

	// 发送ICMP包
	_, err = conn.Write(data)
	if err != nil {
		p.sendPingResult(target, math.NaN())
		return
	}

	// 读取回复
	reply := make([]byte, 1500)
	n, err := conn.Read(reply)
	if err != nil {
		p.sendPingResult(target, math.NaN())
		return
	}

	// 计算往返时间
	rtt := time.Since(startTime)

	// 解析ICMP回复
	replyMsg, err := icmp.ParseMessage(int(ipv4.ICMPTypeEchoReply), reply[:n])
	if err != nil {
		p.sendPingResult(target, math.NaN())
		return
	}

	// 验证这是我们的回复
	if echo, ok := replyMsg.Body.(*icmp.Echo); ok {
		if echo.ID == (os.Getpid()&0xffff) && echo.Seq == seq {
			// 成功收到回复，发送延迟结果（转换为毫秒）
			latencyMs := float64(rtt.Nanoseconds()) / 1e6
			p.sendPingResult(target, latencyMs)
			return
		}
	}

	// 如果验证失败，记录为超时
	p.sendPingResult(target, math.NaN())
}

// Stop 停止特权模式的pinger
func (p *privilegedPinger) Stop() {
	// 调用基础的停止方法
	p.basePinger.Stop()
}
