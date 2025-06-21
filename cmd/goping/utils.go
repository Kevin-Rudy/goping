package main

import (
	"fmt"

	"github.com/Kevin-Rudy/goping/pkg/pinger"
)

// 程序信息常量
const (
	AppName    = "goping"
	AppVersion = "0.1.0"
	AppDesc    = "智能适配运行权限的可视化多目标PING工具"
)

// showSystemInfo 显示系统环境和配置信息
func showSystemInfo() {
	fmt.Println("\n系统信息:")
	fmt.Printf("  操作系统: %s\n", pinger.GetOSName())
	fmt.Printf("  权限状态: %s\n", pinger.GetPrivilegeStatus())
	fmt.Printf("  实现方式: %s\n", pinger.GetImplementationType())
}

// printUsageInstructions 显示TUI操作说明
func printUsageInstructions() {
	fmt.Println("操作说明:")
	fmt.Println("  ↑/↓ 方向键  - 导航选择目标")
	fmt.Println("  在边界继续按方向键 - 切换到全选模式")
	fmt.Println("  q 或 Ctrl+C - 退出程序")
	fmt.Println("========================================")
}
