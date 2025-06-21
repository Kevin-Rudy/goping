package main

import (
	"fmt"

	"github.com/Kevin-Rudy/goping/pkg/pinger"
	"github.com/Kevin-Rudy/goping/pkg/tui"
	"github.com/urfave/cli/v2"
)

// runApp 主要应用逻辑处理函数
func runApp(c *cli.Context) error {
	// 验证命令行参数
	targets := c.Args().Slice()
	if len(targets) == 0 {
		return cli.Exit("错误: 必须指定至少一个要ping的目标地址\n使用方法: goping <目标主机...>", 1)
	}

	// IP版本冲突检查
	explicitIPv4 := c.IsSet("4")
	ipv6 := c.Bool("6")
	if explicitIPv4 && ipv6 {
		return cli.Exit("错误: -4 和 -6 选项不能同时使用", 1)
	}

	// 构建配置
	appConfig := buildConfigFromCLI(c)

	// 验证配置
	if err := validateConfig(appConfig); err != nil {
		return cli.Exit(fmt.Sprintf("配置验证失败: %v", err), 1)
	}

	// 显示运行配置
	printRunningConfig(appConfig)

	// 显示系统环境信息
	showSystemInfo()

	fmt.Println("\n正在初始化ping引擎...")

	// 创建Pinger实例 - 使用配置而不是单独的参数
	pingerInstance, err := pinger.NewPinger(appConfig.Targets, appConfig.PingerConfig)
	if err != nil {
		return cli.Exit(fmt.Sprintf("无法创建ping引擎: %v", err), 1)
	}

	fmt.Println("ping引擎初始化成功")
	fmt.Println("\n正在启动TUI界面...")

	// 显示使用说明
	printUsageInstructions()

	// 创建并启动TUI实例 - 使用新的签名
	tuiInstance := tui.NewTUI(pingerInstance, appConfig.Targets, appConfig.TUIConfig, appConfig.PingerConfig)

	// 启动TUI界面 - 这会阻塞直到用户退出
	if err := tuiInstance.Run(); err != nil {
		return cli.Exit(fmt.Sprintf("TUI运行出错: %v", err), 1)
	}

	fmt.Println("\n程序已退出")
	return nil
}

// printRunningConfig 打印运行配置信息
func printRunningConfig(config *AppConfig) {
	fmt.Printf("目标地址: %v\n", config.Targets)
	fmt.Printf("ping间隔: %v\n", config.PingerConfig.Interval)
	fmt.Printf("ping超时: %v\n", config.PingerConfig.Timeout)
	fmt.Printf("缓冲区大小: %d\n", config.TUIConfig.MaxHistorySize)
}
