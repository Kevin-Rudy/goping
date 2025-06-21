// Package tui 图表渲染模块
package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Kevin-Rudy/goping/pkg/core"
)

// brailleCell 定义盲文字符的cell结构
type brailleCell struct {
	char  int
	color string
}

// validateChartSize 验证图表尺寸是否合理
func (t *TUI) validateChartSize(width, height int) string {
	if height < t.tuiConfig.MinChartHeight || width < t.tuiConfig.MinChartWidth {
		return "终端尺寸过小"
	}
	if width > t.tuiConfig.MaxChartSize || height > t.tuiConfig.MaxChartSize {
		return "终端尺寸过大"
	}
	return ""
}

// calculateValueRange 计算数据的值范围
func (t *TUI) calculateValueRange(targetDataPoints map[string][]core.DataPoint, windowStart, windowEnd time.Time) (minVal, maxVal, valueRange float64, errMsg string) {
	// 收集窗口内的所有有效数据点
	var allValidValues []float64
	for _, dataPoints := range targetDataPoints {
		for _, point := range dataPoints {
			if point.Timestamp.After(windowStart) && point.Timestamp.Before(windowEnd) {
				if !math.IsNaN(point.Value) && !math.IsInf(point.Value, 0) {
					allValidValues = append(allValidValues, point.Value)
				}
			}
		}
	}

	if len(allValidValues) == 0 {
		return 0, 0, 0, "当前窗口内没有有效数据"
	}

	minVal, maxVal = allValidValues[0], allValidValues[0]
	for _, v := range allValidValues {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// 如果所有值都一样，特殊处理
	if maxVal == minVal {
		maxVal++
		minVal--
	}

	// 采用缓冲算法
	maxVal = maxVal + maxVal*t.tuiConfig.ValueBufferRatio
	minVal = minVal - minVal*t.tuiConfig.ValueBufferRatio
	if minVal < 0 {
		minVal = 0
	}

	valueRange = maxVal - minVal
	if valueRange == 0 {
		valueRange = 1
	}

	return minVal, maxVal, valueRange, ""
}

// drawSingleTargetChart 绘制单目标图表，基于时间戳
func (t *TUI) drawSingleTargetChart(identifier string, width, height int) string {
	targetDataPoints := make(map[string][]core.DataPoint)

	t.statsMu.RLock()
	if stats, exists := t.statsData[identifier]; exists && len(stats.History) > 0 {
		targetDataPoints[identifier] = stats.History
	}
	t.statsMu.RUnlock()

	if len(targetDataPoints) == 0 {
		return "没有数据"
	}

	colors := map[string]string{
		identifier: t.getTargetColor(identifier),
	}

	return t.drawChartWithTimestamps(targetDataPoints, colors, width, height)
}

// drawMultiTargetChart 绘制多目标对比图表，基于时间戳
func (t *TUI) drawMultiTargetChart(width, height int) string {
	allTargetDataPoints := make(map[string][]core.DataPoint)
	colors := make(map[string]string)

	t.statsMu.RLock()
	// 使用排序后的标识符列表，确保颜色分配稳定
	for _, identifier := range t.identifiers {
		if stats, exists := t.statsData[identifier]; exists && len(stats.History) > 0 {
			allTargetDataPoints[identifier] = stats.History
			colors[identifier] = t.getTargetColor(identifier)
		}
	}
	t.statsMu.RUnlock()

	if len(allTargetDataPoints) == 0 {
		return "没有数据"
	}

	return t.drawChartWithTimestamps(allTargetDataPoints, colors, width, height)
}

// drawChartWithTimestamps 基于时间戳绘制图表
func (t *TUI) drawChartWithTimestamps(targetDataPoints map[string][]core.DataPoint, colors map[string]string, width, height int) string {
	// 检查图表尺寸是否合理
	if sizeErr := t.validateChartSize(width, height); sizeErr != "" {
		return sizeErr
	}

	// 获取当前时间窗口
	windowStart, windowEnd := t.getTimeWindow()

	// 计算值范围
	minVal, maxVal, valueRange, err := t.calculateValueRange(targetDataPoints, windowStart, windowEnd)
	if err != "" {
		return err
	}

	// 2. 动态计算Y轴标签宽度
	topLabel := formatLatency(maxVal)
	bottomLabel := formatLatency(minVal)
	maxLabelLen := len(topLabel)
	if len(bottomLabel) > maxLabelLen {
		maxLabelLen = len(bottomLabel)
	}
	yAxisLabelWidth := maxLabelLen + 2 // +2 为│分隔符和右侧空格留出缓冲

	// 3. 准备画布尺寸
	chartBodyHeight := height - 2 // 为X轴和时间戳留出2行空间
	chartWidth := width - yAxisLabelWidth

	// 确保画布尺寸合理
	if chartBodyHeight <= 0 || chartWidth <= 0 {
		return "可绘制区域过小"
	}

	// 4. 创建盲文画布
	canvas := make([][]brailleCell, chartWidth)
	for i := range canvas {
		canvas[i] = make([]brailleCell, chartBodyHeight)
	}

	// 定义盲文点阵的映射关系 (2x4 grid)
	brailleDotMap := [4][2]int{
		{0b00000001, 0b00001000}, // (y:0, x:0), (y:0, x:1)
		{0b00000010, 0b00010000}, // (y:1, x:0), (y:1, x:1)
		{0b00000100, 0b00100000}, // (y:2, x:0), (y:2, x:1)
		{0b01000000, 0b10000000}, // (y:3, x:0), (y:3, x:1)
	}

	// 5. 绘制所有目标，使用稳定的遍历顺序
	for _, targetName := range t.identifiers {
		dataPoints := targetDataPoints[targetName]
		color := colors[targetName]
		if color == "" {
			color = "[white]"
		}

		if len(dataPoints) == 0 {
			continue
		}

		// 按时间戳排序数据点，确保线条连接正确
		// （注意：数据点应该已经是按时间排序的，但为了安全起见）

		var lastValidX, lastValidY int = -1, -1

		for _, point := range dataPoints {
			// 只处理在当前时间窗口内的数据点
			if !point.Timestamp.After(windowStart) || !point.Timestamp.Before(windowEnd) {
				continue
			}

			// 计算X坐标（基于时间戳）
			currX := t.timestampToX(point.Timestamp, windowStart, windowEnd, chartWidth*2) // 使用高分辨率
			if currX < 0 || currX >= chartWidth*2 {
				continue // 跳过窗口外的点
			}

			// 计算Y坐标
			var currY int
			if math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
				// 超时：画到图表顶部（天花板）
				currY = 0
			} else {
				// 正常值：按延迟值计算Y坐标
				normalized := (point.Value - minVal) / valueRange
				if math.IsNaN(normalized) || math.IsInf(normalized, 0) {
					currY = 0 // 异常情况也画到顶部
				} else {
					currY = int((1.0 - normalized) * float64(chartBodyHeight*4-1))
				}
			}

			// 边界检查（高分辨率坐标）
			if currY < 0 {
				currY = 0
			} else if currY >= chartBodyHeight*4 {
				currY = chartBodyHeight*4 - 1
			}

			// 如果上一个有效点存在，绘制连接线
			if lastValidX != -1 && lastValidY != -1 {
				t.drawBrailleLineNew(canvas, lastValidX, lastValidY, currX, currY, chartBodyHeight*4, chartWidth*2, color)
			} else {
				// 如果这是线条的第一个点，直接在画布上标记
				canvasX := currX / 2
				canvasY := currY / 4
				subY := currY % 4
				subX := currX % 2

				if canvasX >= 0 && canvasX < chartWidth && canvasY >= 0 && canvasY < chartBodyHeight {
					canvas[canvasX][canvasY].char |= brailleDotMap[subY][subX]
					canvas[canvasX][canvasY].color = color
				}
			}

			// 更新上一个有效点的坐标
			lastValidX, lastValidY = currX, currY
		}
	}

	// 6. 构建输出字符串
	var lines []string

	// 预先计算Y轴标签位置
	yAxisLabelCount := 5
	if chartBodyHeight < yAxisLabelCount {
		yAxisLabelCount = chartBodyHeight
	}

	// 预先计算所有Y轴标签及其对应的像素行号
	yAxisLabels := make(map[int]string)
	if yAxisLabelCount > 1 {
		for i := 0; i < yAxisLabelCount; i++ {
			// 在数值上均匀分布
			normalized := float64(i) / float64(yAxisLabelCount-1) // 0.0 到 1.0
			value := maxVal - normalized*valueRange               // 从最大值到最小值
			// 计算对应的像素行号
			pixelRow := int(normalized * float64(chartBodyHeight-1))
			yAxisLabels[pixelRow] = formatLatency(value)
		}
	}

	// 绘制Y轴和图表主体
	for i := 0; i < chartBodyHeight; i++ {
		// 从预计算的map中查找Y轴标签
		yLabel := yAxisLabels[i]

		line := fmt.Sprintf("[gray]%*s[white] [gray]│[white]", yAxisLabelWidth-2, yLabel)

		for j := 0; j < chartWidth; j++ {
			cell := canvas[j][i]
			if cell.char == 0 {
				line += " "
			} else {
				// 确保颜色已经是tview格式，不需要再次添加方括号
				line += cell.color + string(rune(0x2800+cell.char)) + "[white]"
			}
		}
		lines = append(lines, line)
	}

	// 7. 绘制X轴
	xAxisLine := fmt.Sprintf("%-*s└%s", yAxisLabelWidth-1, "", strings.Repeat("─", chartWidth))
	lines = append(lines, "[gray]"+xAxisLine+"[white]")

	// X轴时间刻度 - 显示实际的时间窗口
	startTimeStr := windowStart.Format("15:04:05")
	endTimeStr := windowEnd.Format("15:04:05")

	spaceCount := chartWidth - len(startTimeStr) - len(endTimeStr)
	if spaceCount < 1 {
		spaceCount = 1
	}
	timeLine := fmt.Sprintf("%-*s%s%*s%s", yAxisLabelWidth, "", startTimeStr, spaceCount, "", endTimeStr)
	lines = append(lines, "[gray]"+timeLine+"[white]")

	// 保护性检查：确保输出不会超过可用高度，保证X轴总是可见
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

// drawBrailleLineNew 使用布雷森汉姆算法在盲文画布上绘制线段
func (t *TUI) drawBrailleLineNew(canvas [][]brailleCell, x1, y1, x2, y2, maxHeight, chartWidth int, color string) {
	// 定义盲文点阵的映射关系 (2x4 grid)
	brailleDotMap := [4][2]int{
		{0b00000001, 0b00001000}, // (y:0, x:0), (y:0, x:1)
		{0b00000010, 0b00010000}, // (y:1, x:0), (y:1, x:1)
		{0b00000100, 0b00100000}, // (y:2, x:0), (y:2, x:1)
		{0b01000000, 0b10000000}, // (y:3, x:0), (y:3, x:1)
	}

	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	x, y := x1, y1
	for {
		// 在当前位置放置字符
		if y >= 0 && y < maxHeight && x >= 0 && x < chartWidth {
			// 计算子像素位置
			subY := y % 4 // 盲文字符内的垂直位置 (0-3)
			subX := x % 2 // 盲文字符内的水平位置 (0-1)

			// 计算画布坐标（每个盲文字符覆盖2x4个子像素）
			canvasX := x / 2
			canvasY := y / 4

			// 确保画布坐标在有效范围内
			if canvasX >= 0 && canvasX < chartWidth && canvasY >= 0 && canvasY < len(canvas[0]) {
				canvas[canvasX][canvasY].char |= brailleDotMap[subY][subX]
				canvas[canvasX][canvasY].color = color
			}
		}

		// 检查是否到达终点
		if x == x2 && y == y2 {
			break
		}

		// 计算下一个位置
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}
 