// Package tui 交互控制模块
package tui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// 导航事件频率控制 - 包级私有变量
var (
	navigationEventCounter   int                       // 事件计数器
	navigationEventThreshold = 5                       // 5次事件后休息
	navigationRestDuration   = 100 * time.Millisecond // 休息1秒
	isNavigationResting      bool                      // 是否在休息状态
	lastNavigationEventTime  time.Time                 // 最后一次事件时间
)

// shouldHandleNavigationEvent 判断是否应该处理导航事件
func shouldHandleNavigationEvent() bool {
	now := time.Now()

	// 如果正在休息中，检查是否休息够了
	if isNavigationResting {
		if now.Sub(lastNavigationEventTime) >= navigationRestDuration {
			// 休息够了，重置状态
			isNavigationResting = false
			navigationEventCounter = 0
			return true
		}
		// 还在休息，忽略事件
		return false
	}

	// 不在休息状态，可以处理
	return true
}

// recordNavigationEvent 记录导航事件
func recordNavigationEvent() {
	navigationEventCounter++
	lastNavigationEventTime = time.Now()

	// 检查是否达到阈值
	if navigationEventCounter >= navigationEventThreshold {
		isNavigationResting = true
	}
}

// setupKeyBindings 设置键盘绑定
func (t *TUI) setupKeyBindings() {
	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			t.Stop()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				t.Stop()
				return nil
			}
		case tcell.KeyUp:
			// 添加频率控制检查
			if shouldHandleNavigationEvent() {
				t.navigateUp()
				recordNavigationEvent()
			}
			return nil
		case tcell.KeyDown:
			// 添加频率控制检查
			if shouldHandleNavigationEvent() {
				t.navigateDown()
				recordNavigationEvent()
			}
			return nil
		}
		return event
	})
}

// navigateUp 向上导航
func (t *TUI) navigateUp() {
	if len(t.identifiers) == 0 {
		return
	}

	if t.selectedRow == -1 {
		// 从全选状态按上键，选择最后一个条目
		t.selectedRow = len(t.identifiers) - 1
	} else if t.selectedRow > 0 {
		// 向上移动到上一个条目
		t.selectedRow--
	} else {
		// 在第一个条目时按上键，返回全选状态
		t.selectedRow = -1
	}

	if !t.testMode {
		t.updateSelection()
		t.updateChart()
	}
}

// navigateDown 向下导航
func (t *TUI) navigateDown() {
	if len(t.identifiers) == 0 {
		return
	}

	if t.selectedRow == -1 {
		// 从全选状态按下键，选择第一个条目
		t.selectedRow = 0
	} else if t.selectedRow < len(t.identifiers)-1 {
		// 向下移动到下一个条目
		t.selectedRow++
	} else {
		// 在最后一个条目时按下键，返回全选状态
		t.selectedRow = -1
	}

	if !t.testMode {
		t.updateSelection()
		t.updateChart()
	}
}

// updateSelection 更新行选择状态
func (t *TUI) updateSelection() {
	if t.testMode || len(t.rowFlexes) == 0 {
		return
	}

	// 重置所有行的背景色
	for i, rowFlex := range t.rowFlexes {
		if i == 0 {
			// 表头行，保持默认背景色
			for j := 0; j < rowFlex.GetItemCount(); j++ {
				if textView, ok := rowFlex.GetItem(j).(*tview.TextView); ok {
					textView.SetBackgroundColor(tcell.ColorDefault)
				}
			}
		} else if t.selectedRow == i-1 { // -1 因为表头行占用了索引0
			// 高亮选中的数据行的所有子项
			for j := 0; j < rowFlex.GetItemCount(); j++ {
				if textView, ok := rowFlex.GetItem(j).(*tview.TextView); ok {
					textView.SetBackgroundColor(tcell.ColorDarkCyan)
				}
			}
		} else {
			// 重置未选中数据行的所有子项背景色
			for j := 0; j < rowFlex.GetItemCount(); j++ {
				if textView, ok := rowFlex.GetItem(j).(*tview.TextView); ok {
					textView.SetBackgroundColor(tcell.ColorDefault)
				}
			}
		}
	}
}
