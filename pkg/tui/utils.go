// Package tui 工具函数和辅助类型
package tui

import (
	"fmt"
	"math"

	"github.com/Kevin-Rudy/goping/pkg/core"
)

// formatLatency 提供自适应的延迟格式化
func formatLatency(latency float64) string {
	if math.IsNaN(latency) {
		return "N/A"
	}

	if latency < 1.0 {
		// 小于1ms，显示为微秒
		return fmt.Sprintf("%.0fµs", latency*1000)
	} else if latency < 1000.0 {
		// 1ms到1000ms之间，显示为毫秒
		return fmt.Sprintf("%.1fms", latency)
	} else {
		// 大于等于1000ms，显示为秒
		return fmt.Sprintf("%.2fs", latency/1000)
	}
}

// getTargetColor 根据目标标识符获取对应的颜色（统一颜色分配逻辑）
func (t *TUI) getTargetColor(identifier string) string {
	// 重新设计的颜色序列，使用更兼容、更鲜明的颜色
	colorSequence := []string{
		"[green]",       // 绿色 - 通常显示良好
		"[yellow]",      // 黄色 - 通常显示良好
		"[blue]",        // 蓝色 - 通常显示良好
		"[red]",         // 红色 - 通常显示良好
		"[lightgreen]",  // 浅绿色
		"[lightblue]",   // 浅蓝色
		"[lightyellow]", // 浅黄色
		"[lightred]",    // 浅红色
		"[white]",       // 白色 - 在深色背景下醒目
		"[gray]",        // 灰色
		"[darkgreen]",   // 深绿色
		"[darkblue]",    // 深蓝色
		"[darkyellow]",  // 深黄色
		"[darkred]",     // 深红色
	}

	// 基于预定义的targets顺序分配颜色，确保颜色稳定
	for i, target := range t.targets {
		if target == identifier {
			return colorSequence[i%len(colorSequence)]
		}
	}

	// 如果没找到，返回白色作为默认值
	return "[white]"
}

// abs 返回整数的绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// updateIdentifiersForTest 在测试模式下更新标识符列表（不操作图形组件）
func (t *TUI) updateIdentifiersForTest() {
	t.statsMu.RLock()
	defer t.statsMu.RUnlock()

	// 使用预定义的targets顺序，而不是动态排序
	// 只包含已有数据的targets
	var identifiers []string
	for _, target := range t.targets {
		if _, exists := t.statsData[target]; exists {
			identifiers = append(identifiers, target)
		}
	}
	t.identifiers = identifiers

	// 确保选择状态正确
	if t.selectedRow >= len(t.identifiers) {
		t.selectedRow = len(t.identifiers) - 1
	}
}

// getOrCreateStats 获取或创建统计数据结构
func (t *TUI) getOrCreateStats(identifier string) *core.Stats {
	stats, exists := t.statsData[identifier]
	if !exists {
		stats = core.NewStats(identifier)
		t.statsData[identifier] = stats
	}
	return stats
}
