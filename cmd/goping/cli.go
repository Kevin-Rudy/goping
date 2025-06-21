package main

import (
	"fmt"
	"time"

	"github.com/Kevin-Rudy/goping/pkg/pinger"
	"github.com/urfave/cli/v2"
)

// createCliApp 创建CLI应用实例
func createCliApp() *cli.App {
	app := &cli.App{
		Name:    AppName,
		Version: AppVersion,
		Usage:   AppDesc,
		Flags:   createCliFlags(),
		Action:  runApp,
		Before: func(c *cli.Context) error {
			// 显示启动信息
			fmt.Printf("正在启动 %s v%s...\n", AppName, AppVersion)
			return nil
		},
		ArgsUsage: "<目标主机...>",
	}

	// 添加版本子命令
	app.Commands = createCommands()

	return app
}

// createCliFlags 创建CLI参数定义
func createCliFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "4",
			Usage: "使用IPv4进行域名解析（默认）",
			Value: true,
		},
		&cli.BoolFlag{
			Name:  "6",
			Usage: "使用IPv6进行域名解析",
		},
		&cli.DurationFlag{
			Name:    "watch-interval",
			Aliases: []string{"n"},
			Value:   200 * time.Millisecond,
			Usage:   "ping间隔时间 (例如: 100ms, 1s)",
		},
		&cli.DurationFlag{
			Name:    "timeout",
			Aliases: []string{"t"},
			Value:   3 * time.Second,
			Usage:   "ping超时时间 (例如: 3s, 1000ms)",
		},
		&cli.IntFlag{
			Name:    "buffer",
			Aliases: []string{"b"},
			Value:   150,
			Usage:   "TUI图表历史缓冲区大小",
		},
		&cli.DurationFlag{
			Name:    "refresh-rate",
			Aliases: []string{"r"},
			Value:   200 * time.Millisecond,
			Usage:   "UI刷新频率 (例如: 100ms, 500ms)",
		},
		&cli.IntFlag{
			Name:  "chart-width",
			Value: 20,
			Usage: "最小图表宽度",
		},
		&cli.IntFlag{
			Name:  "chart-height",
			Value: 5,
			Usage: "最小图表高度",
		},
		&cli.Float64Flag{
			Name:  "ceiling",
			Value: 100.0,
			Usage: "图表默认上限值 (ms)",
		},
		&cli.DurationFlag{
			Name:  "timeout-threshold",
			Value: 0, // 0表示自动计算
			Usage: "超时判定阈值，0表示自动计算 (例如: 5s)",
		},
		&cli.Float64Flag{
			Name:  "timeout-buffer-ratio",
			Value: 1.2,
			Usage: "超时缓冲比例，TUI超时 = Pinger超时 * 此比例",
		},
	}
}

// createCommands 创建子命令
func createCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "显示详细版本信息",
			Action: func(c *cli.Context) error {
				fmt.Printf("%s v%s\n", AppName, AppVersion)
				fmt.Printf("描述: %s\n", AppDesc)
				fmt.Printf("系统: %s\n", pinger.GetOSName())
				fmt.Printf("实现: %s\n", pinger.GetImplementationType())
				return nil
			},
		},
	}
}
