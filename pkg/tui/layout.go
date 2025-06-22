// Package tui 布局管理模块
package tui

import (
	"fmt"

	"github.com/Kevin-Rudy/goping/pkg/core"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// setupUI 设置用户界面布局
func (t *TUI) setupUI() {
	// 设置图表属性
	t.chart.SetWordWrap(false)
	t.chart.SetDynamicColors(true)
	t.chart.SetText("[yellow]正在初始化，等待数据...[white]")

	// 创建主垂直布局
	t.flex = tview.NewFlex()
	t.flex.SetDirection(tview.FlexRow)

	// 添加初始的等待信息行
	waitingInfo := tview.NewTextView()
	waitingInfo.SetText("[green]GoPing 已启动[white] - [yellow]正在连接目标...[white]")
	waitingInfo.SetDynamicColors(true)
	waitingInfo.SetTextAlign(tview.AlignCenter)

	// 立即添加等待信息和图表到布局中
	t.flex.AddItem(waitingInfo, 1, 0, false)
	t.flex.AddItem(t.chart, 0, 1, false)

	t.app.SetRoot(t.flex, true)
}

// rebuildUI 重建UI布局
func (t *TUI) rebuildUI() {
	if t.testMode {
		t.updateIdentifiersForTest()
		return
	}

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

	// 清空主布局
	t.flex.Clear()
	t.rowFlexes = make([]*tview.Flex, 0, len(t.identifiers)+1) // +1 为表头行

	if len(t.identifiers) == 0 {
		// 没有数据时，显示等待界面
		waitingInfo := tview.NewTextView()
		waitingInfo.SetText("[green]GoPing 已启动[white] - [yellow]正在连接目标...[white]")
		waitingInfo.SetDynamicColors(true)
		waitingInfo.SetTextAlign(tview.AlignCenter)

		t.flex.AddItem(waitingInfo, 1, 0, false)
		t.flex.AddItem(t.chart, 0, 1, false)
		return
	}

	// 预计算：收集所有可能的 Summary keys
	summaryKeysSet := make(map[string]bool)
	for _, identifier := range t.identifiers {
		if stats, exists := t.statsData[identifier]; exists {
			for key := range stats.Summary {
				summaryKeysSet[key] = true
			}
		}
	}

	// 按预定义顺序排列统计项
	predefinedOrder := []string{"t/o", "丢包率", "发送/接收", "平均延迟", "最小延迟", "最大延迟"}
	var summaryKeys []string
	for _, key := range predefinedOrder {
		if summaryKeysSet[key] {
			summaryKeys = append(summaryKeys, key)
		}
	}

	// 构建完整的表头：目标 + 所有Summary keys
	headers := []string{"目标"}
	headers = append(headers, summaryKeys...)
	t.headers = headers

	// 创建表头行
	headerFlex := t.createHeaderRow(summaryKeys)
	t.flex.AddItem(headerFlex, 1, 0, false)
	t.rowFlexes = append(t.rowFlexes, headerFlex)

	// 为每个数据源创建一行
	for _, identifier := range t.identifiers {
		stats := t.statsData[identifier]
		rowFlex := t.createDataRow(identifier, stats, summaryKeys)
		t.flex.AddItem(rowFlex, 1, 0, false)
		t.rowFlexes = append(t.rowFlexes, rowFlex)
	}

	// 最后添加图表，占据所有剩余空间
	t.flex.AddItem(t.chart, 0, 1, false)

	// 确保选择状态正确（注意现在索引需要+1，因为有表头行）
	if t.selectedRow >= len(t.identifiers) {
		t.selectedRow = len(t.identifiers) - 1
	}
	t.updateSelection()
}

// createHeaderRow 创建表头行
func (t *TUI) createHeaderRow(summaryKeys []string) *tview.Flex {
	headerFlex := tview.NewFlex()
	headerFlex.SetDirection(tview.FlexColumn)

	// 添加表头的目标列
	targetHeaderText := tview.NewTextView()
	targetHeaderText.SetText(fmt.Sprintf("[yellow]%-20s[white]", "目标"))
	targetHeaderText.SetDynamicColors(true)
	targetHeaderText.SetTextAlign(tview.AlignLeft)
	headerFlex.AddItem(targetHeaderText, 0, 2, false) // 给目标列更多空间

	// 添加表头的数据列
	for _, header := range summaryKeys {
		headerText := tview.NewTextView()
		headerText.SetText(fmt.Sprintf("[yellow]%8s[white]", header))
		headerText.SetDynamicColors(true)
		headerText.SetTextAlign(tview.AlignCenter)
		headerFlex.AddItem(headerText, 0, 1, false)
	}

	return headerFlex
}

// createDataRow 创建数据行
func (t *TUI) createDataRow(identifier string, stats *core.Stats, summaryKeys []string) *tview.Flex {
	rowFlex := tview.NewFlex()
	rowFlex.SetDirection(tview.FlexColumn)

	// 确定目标的颜色（与图表一致，使用统一的颜色分配函数）
	color := t.getTargetColor(identifier)

	// 第一列：目标标识符（带颜色）
	targetText := tview.NewTextView()
	targetText.SetText(fmt.Sprintf("%s%-20s[white]", color, identifier))
	targetText.SetDynamicColors(true)
	targetText.SetTextAlign(tview.AlignLeft)
	rowFlex.AddItem(targetText, 0, 2, false) // 给目标名称更多空间

	// 其他列：严格按照预计算的 summaryKeys 顺序填充数据
	for _, key := range summaryKeys {
		value := "N/A" // 默认值
		if stats.Summary != nil {
			if val, exists := stats.Summary[key]; exists && val != "" {
				value = val
			}
		}

		dataText := tview.NewTextView()
		dataText.SetText(fmt.Sprintf("%8s", value))
		dataText.SetTextAlign(tview.AlignCenter)
		dataText.SetTextColor(tcell.ColorWhite)
		rowFlex.AddItem(dataText, 0, 1, false)
	}

	return rowFlex
}

// updateChart 更新图表显示
func (t *TUI) updateChart() {
	if t.testMode || t.chart == nil {
		return
	}

	t.statsMu.RLock()
	defer t.statsMu.RUnlock()

	if len(t.identifiers) == 0 {
		t.chart.SetText("没有数据")
		return
	}

	// 获取图表视图的实际可绘制尺寸
	_, _, width, height := t.chart.GetInnerRect()

	// 确保有合理的最小尺寸
	if width < 20 {
		width = 80
	}
	if height < 10 {
		height = 15
	}

	var chartText string

	if t.selectedRow == -1 {
		// 全选状态：显示所有目标的折线图
		chartText = t.drawMultiTargetChart(width, height)
	} else if t.selectedRow >= 0 && t.selectedRow < len(t.identifiers) {
		// 单选状态：显示选中目标的折线图
		identifier := t.identifiers[t.selectedRow]
		chartText = t.drawSingleTargetChart(identifier, width, height)
	}

	t.chart.SetText(chartText)
}

// safeUIUpdate 安全地执行UI更新操作
func (t *TUI) safeUIUpdate(updateFunc func()) {
	defer func() {
		if r := recover(); r != nil {
			// 如果应用已经停止，忽略panic
		}
	}()
	t.app.QueueUpdateDraw(updateFunc)
}
