package main

import (
	"fmt"

	"github.com/Kevin-Rudy/goping/pkg/pinger"
	"github.com/Kevin-Rudy/goping/pkg/tui"
	"github.com/urfave/cli/v2"
)

// AppConfig 应用层配置聚合
type AppConfig struct {
	PingerConfig *pinger.Config
	TUIConfig    *tui.Config
	Targets      []string
}

// buildConfigFromCLI 从命令行参数构建配置
func buildConfigFromCLI(c *cli.Context) *AppConfig {
	// 构建 pinger 配置
	pingerConfig := pinger.DefaultConfig()
	if c.Bool("6") {
		pingerConfig.IPVersion = 6
	}
	if c.IsSet("watch-interval") {
		pingerConfig.Interval = c.Duration("watch-interval")
	}
	if c.IsSet("timeout") {
		pingerConfig.Timeout = c.Duration("timeout")
	}

	// 构建 TUI 配置
	tuiConfig := tui.DefaultConfig()
	if c.IsSet("buffer") {
		tuiConfig.MaxHistorySize = c.Int("buffer")
	}
	if c.IsSet("refresh-rate") {
		tuiConfig.RefreshInterval = c.Duration("refresh-rate")
	}
	if c.IsSet("chart-width") {
		tuiConfig.MinChartWidth = c.Int("chart-width")
	}
	if c.IsSet("chart-height") {
		tuiConfig.MinChartHeight = c.Int("chart-height")
	}
	if c.IsSet("timeout-buffer-ratio") {
		tuiConfig.TimeoutBufferRatio = c.Float64("timeout-buffer-ratio")
	}

	return &AppConfig{
		PingerConfig: pingerConfig,
		TUIConfig:    tuiConfig,
		Targets:      c.Args().Slice(),
	}
}

// validateConfig 验证配置的合理性
func validateConfig(config *AppConfig) error {
	// 验证 pinger 配置
	if err := config.PingerConfig.Validate(); err != nil {
		return fmt.Errorf("pinger配置错误: %v", err)
	}

	// 验证 TUI 配置
	if err := config.TUIConfig.Validate(); err != nil {
		return fmt.Errorf("tui配置错误: %v", err)
	}

	return nil
}
