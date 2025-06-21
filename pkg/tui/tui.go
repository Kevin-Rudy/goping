// Package tui 提供基于时间戳的终端用户界面组件
// 支持实时数据可视化和多目标监控
package tui

import (
	"sync"
	"time"

	"github.com/Kevin-Rudy/goping/pkg/core"
	"github.com/Kevin-Rudy/goping/pkg/pinger"
	"github.com/rivo/tview"
)

// TUI 主界面结构
type TUI struct {
	app        *tview.Application
	rowFlexes  []*tview.Flex
	chart      *tview.TextView
	flex       *tview.Flex
	dataSource core.DataSource

	// 配置信息
	tuiConfig        *Config       // TUI配置
	timeoutThreshold time.Duration // 计算得出的超时阈值

	// 数据存储
	statsData map[string]*core.Stats
	statsMu   sync.RWMutex

	// 界面状态
	selectedRow int
	headers     []string
	identifiers []string
	targets     []string // 保存命令行输入的目标顺序

	// 控制
	stopChan chan struct{}
	doneChan chan struct{}

	// 测试模式标志
	testMode bool

	// 时间管理
	startTime time.Time // 程序启动时间，用于时间窗口计算
}

// NewTUI 创建新的TUI实例
func NewTUI(dataSource core.DataSource, targets []string, tuiConfig *Config, pingerConfig *pinger.Config) *TUI {
	tui := &TUI{
		app:              tview.NewApplication(),
		chart:            tview.NewTextView(),
		dataSource:       dataSource,
		targets:          targets,
		tuiConfig:        tuiConfig,
		timeoutThreshold: tuiConfig.GetTimeoutThreshold(pingerConfig.Timeout),
		statsData:        make(map[string]*core.Stats),
		stopChan:         make(chan struct{}),
		doneChan:         make(chan struct{}),
		testMode:         false,
		selectedRow:      -1,         // 默认全选状态
		startTime:        time.Now(), // 记录程序启动时间
	}

	tui.setupUI()
	tui.setupKeyBindings()

	return tui
}

// NewTUIForTest 创建用于测试的TUI实例（不初始化图形组件）
func NewTUIForTest(dataSource core.DataSource, targets []string, tuiConfig *Config, pingerConfig *pinger.Config) *TUI {
	return &TUI{
		app:              tview.NewApplication(), // 创建一个应用实例，但不会运行
		dataSource:       dataSource,
		targets:          targets,
		tuiConfig:        tuiConfig,
		timeoutThreshold: tuiConfig.GetTimeoutThreshold(pingerConfig.Timeout),
		statsData:        make(map[string]*core.Stats),
		stopChan:         make(chan struct{}),
		doneChan:         make(chan struct{}),
		testMode:         true,
		selectedRow:      -1,         // 默认全选状态
		startTime:        time.Now(), // 记录程序启动时间
	}
}

// Run 启动TUI界面
func (t *TUI) Run() error {
	// 启动数据源
	t.dataSource.Start()

	// 启动数据处理goroutine
	go t.processData()

	// 运行应用
	err := t.app.Run()

	// 确保清理工作完成
	<-t.doneChan

	return err
}

// Stop 停止TUI界面
func (t *TUI) Stop() {
	// 先发送停止信号，让processData退出
	select {
	case <-t.stopChan:
		// stopChan已经关闭，避免重复关闭
	default:
		close(t.stopChan)
	}

	// 停止数据源
	t.dataSource.Stop()

	// 停止应用
	t.app.Stop()
}

// processData 处理来自数据源的数据 - 实现时间驱动渲染
func (t *TUI) processData() {
	defer close(t.doneChan)

	dataChan := t.dataSource.DataStream()
	uiTicker := time.NewTicker(t.tuiConfig.RefreshInterval)
	defer uiTicker.Stop()

	timeGridTicker := time.NewTicker(t.tuiConfig.TimeGridInterval)
	defer timeGridTicker.Stop()

	// 初始UI刷新
	t.forceInitialDraw()

	for {
		select {
		case result, ok := <-dataChan:
			if !ok {
				return
			}
			t.handleDataUpdate(result)

		case <-uiTicker.C:
			t.handleUIRefresh()

		case <-timeGridTicker.C:
			t.handleTimeGridMaintenance()

		case <-t.stopChan:
			return
		}
	}
}

// forceInitialDraw 强制初始绘制
func (t *TUI) forceInitialDraw() {
	if !t.testMode && t.app != nil {
		t.app.QueueUpdateDraw(func() {
			// 强制初始绘制
		})
	}
}

// handleDataUpdate 处理数据更新
func (t *TUI) handleDataUpdate(result core.PingResult) {
	t.updateStatsWithTime(result)
}

// handleUIRefresh 处理UI刷新
func (t *TUI) handleUIRefresh() {
	if !t.testMode && t.app != nil {
		t.safeUIUpdate(func() {
			t.rebuildUI()
			t.updateChart()
		})
	}
}

// handleTimeGridMaintenance 处理时间网格维护
func (t *TUI) handleTimeGridMaintenance() {
	t.maintainTimeGrid()
}
