# 🚀 GoPing - 基于时间戳对齐的智能多目标PING工具

GoPing是一个用Go语言开发的现代化网络延迟监控工具，提供**基于时间戳的精确数据对齐**、**Braille字符高精度图表**和**智能权限适配**功能。

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Go Version](https://img.shields.io/badge/go-1.23+-blue)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)

## 📝 项目声明

本项目是一个**学习型项目**，主要目的包括：

- **架构设计实践**：探索和实践现代Go项目的分层架构设计
- **Go语言学习**：深入理解Go语言的并发模型、接口设计和跨平台开发
- **AI辅助开发探索**：实践AI工具在软件开发中的应用模式

### 开发模式说明
- **设计主导**：整体架构、技术选型和模块划分由作者规划设计
- **AI辅助实现**：具体代码实现主要通过Cursor AI辅助完成
- **学习导向**：重点关注设计思路和架构模式，而非生产级别的完善程度

### 使用建议
- 本项目适合作为Go语言学习和架构设计的参考案例
- **不建议直接用于生产环境**，未经充分的测试和优化
- 欢迎用于学习交流，但请了解其实验性质

如果您需要生产级别的网络监控工具，建议考虑更成熟的开源项目。

## ✨ 核心特性

### 🎯 智能权限适配
- **自动检测运行权限**：根据当前用户权限自动选择最佳实现方式
- **多平台支持**：
  - **Windows**: 管理员模式使用Raw Socket，普通用户模式使用Windows ICMP API
  - **Linux**: Root权限使用Raw Socket，普通用户使用DGRAM Socket
  - **macOS**: 支持Raw Socket（需要sudo权限）

### 📊 高精度可视化界面
- **Braille字符图表**：采用Unicode盲文字符实现8x4像素密度的平滑曲线显示
- **时间戳精确对齐**：基于发送时间戳的亚秒级时间窗口对齐，确保多目标数据同步显示
- **动态时间窗口**：智能维护历史数据的时间网格，支持实时滚动和缓冲区管理
- **交互式多目标导航**：方向键在目标间切换，支持单目标详细视图和全局对比模式

### ⚡ 高性能架构设计
- **三层模块化架构**：核心接口层 + Ping引擎层 + TUI界面层完全解耦
- **异步数据流处理**：基于Go通道的生产者-消费者模式，支持高频监控
- **可配置缓冲策略**：
  - 数据通道缓冲区：防止高频ping时的数据丢失
  - 历史记录缓冲区：可配置的滚动窗口大小
  - UI刷新缓冲：独立的渲染频率控制
- **智能超时管理**：TUI层超时阈值与Ping层超时的分离设计

## 🛠️ 安装方法

### 预编译二进制文件
```bash
# 下载对应平台的二进制文件
# Windows用户可直接运行 goping.exe
```

### 从源码编译
```bash
# 克隆仓库
git clone https://github.com/Kevin-Rudy/goping.git
cd goping

# 编译
go build -o goping cmd/goping/*.go

# 或使用 go install
go install github.com/Kevin-Rudy/goping/cmd/goping@latest
```

## 🚀 使用方法

### 基本用法
```bash
# ping单个目标
goping google.com

# ping多个目标
goping google.com baidu.com 8.8.8.8

# 使用IPv6
goping -6 google.com

# 自定义ping间隔（默认200ms）
goping --watch-interval 100ms google.com
goping -n 100ms google.com  # 简写形式
```

### 高级配置
```bash
# 完整配置示例
goping \
  --watch-interval 50ms \     # ping间隔50ms
  --timeout 2s \              # ping超时2秒
  --buffer 300 \              # 历史缓冲区300个数据点
  --refresh-rate 100ms \      # UI刷新频率100ms
  --chart-width 30 \          # 最小图表宽度
  --chart-height 10 \         # 最小图表高度
  --ceiling 200.0 \           # 图表默认上限200ms
  --timeout-buffer-ratio 1.5 \ # TUI超时是ping超时的1.5倍
  google.com baidu.com

# 高频监控模式（需要足够权限）
goping -n 10ms --refresh-rate 50ms 8.8.8.8
```

### 系统信息查看
```bash
# 显示详细版本和系统信息
goping version

# 查看帮助信息
goping --help
```

## 📋 命令行参数详解

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `-4` | | `true` | 使用IPv4进行域名解析（默认） |
| `-6` | | `false` | 使用IPv6进行域名解析 |
| `--watch-interval` | `-n` | `200ms` | ping间隔时间 |
| `--timeout` | `-t` | `3s` | ping超时时间 |
| `--buffer` | `-b` | `150` | TUI图表历史缓冲区大小 |
| `--refresh-rate` | `-r` | `200ms` | UI刷新频率 |
| `--chart-width` | | `20` | 最小图表宽度 |
| `--chart-height` | | `5` | 最小图表高度 |
| `--ceiling` | | `100.0` | 图表默认上限值（ms） |
| `--timeout-threshold` | | `0` | 超时判定阈值，0表示自动计算 |
| `--timeout-buffer-ratio` | | `1.2` | 超时缓冲比例（TUI超时 = Ping超时 × 此比例） |

### 交互式操作
运行后在TUI界面中：
- `↑/↓` 方向键：在目标间导航
- 在边界继续按方向键：切换到全选模式
- `q` 或 `Ctrl+C`：退出程序

## 🔧 技术架构

### 模块化三层架构
```
cmd/goping/          # 应用层 - CLI参数解析和程序入口
├── main.go          # 程序入口
├── app.go           # 应用逻辑控制器
├── cli.go           # 命令行接口定义
├── config.go        # 配置聚合和验证
└── utils.go         # 工具函数和版本信息

pkg/core/            # 核心接口层 - 定义标准接口和数据结构
├── types.go         # 核心数据结构和接口定义
└── types_test.go    # 核心类型测试

pkg/pinger/          # Ping引擎层 - 跨平台ping实现
├── pinger.go        # 主要pinger逻辑
├── config.go        # pinger配置管理
├── capability.go    # 平台能力接口定义
├── capability_*.go  # 各平台能力实现
├── privileged.go    # 特权模式raw socket实现
├── dgram_linux.go   # Linux非特权DGRAM实现
└── windows.go       # Windows API实现

pkg/tui/             # TUI界面层 - 终端用户界面
├── tui.go           # TUI主控制器
├── config.go        # TUI配置管理
├── chart_renderer.go # Braille字符图表渲染
├── data_processor.go # 数据处理和统计
├── time_manager.go  # 时间窗口管理
├── layout.go        # 界面布局管理
└── interaction.go   # 用户交互处理
```

### 关键设计模式

#### 1. 数据源接口抽象
```go
type DataSource interface {
    DataStream() <-chan PingResult  // 实时数据流
    Start()                         // 启动数据收集
    Stop()                          // 停止并清理资源
}
```

#### 2. 时间戳对齐机制
- 所有ping结果携带精确的发送时间戳
- TUI层基于时间戳进行数据对齐和窗口管理
- 支持亚秒级的时间网格维护

#### 3. 平台权限适配策略
```
Windows:
├── 管理员权限 → Raw Socket (ICMP)
└── 普通用户 → Windows ICMP API

Linux:
├── Root权限 → Raw Socket (ICMP)  
└── 普通用户 → DGRAM Socket (需要 net.ipv4.ping_group_range)

macOS:
└── Root权限 → Raw Socket (ICMP)
```

## 📦 依赖项目

- [tview](https://github.com/rivo/tview) `v0.0.0-20250501113434` - 终端UI框架
- [tcell](https://github.com/gdamore/tcell/v2) `v2.8.1` - 终端控制库
- [cli](https://github.com/urfave/cli/v2) `v2.27.7` - 命令行参数解析
- [golang.org/x/net](https://golang.org/x/net) `v0.29.0` - 扩展网络库
- [golang.org/x/sys](https://golang.org/x/sys) `v0.29.0` - 系统调用接口

## 🎛️ 性能调优建议

### 高频监控配置
```bash
# 10ms间隔高频ping (需要管理员权限)
goping -n 10ms --refresh-rate 50ms --buffer 500 8.8.8.8

# 适合网络质量监控的配置
goping -n 100ms --timeout 5s --buffer 300 target.com
```

### 多目标监控优化
```bash
# 大量目标时的推荐配置
goping \
  --buffer 200 \              # 增大缓冲区
  --refresh-rate 500ms \      # 降低刷新频率减少CPU占用
  --timeout-buffer-ratio 2.0 \# 增大超时缓冲避免误判
  host1 host2 host3 ... host10
```

### 资源受限环境
```bash
# 低资源消耗配置
goping -n 1s --refresh-rate 1s --buffer 60 --chart-height 3 target.com
```

## 🤝 参与贡献

欢迎提交Issue和Pull Request！

### 开发环境要求
- Go 1.23.1+
- 支持的操作系统：Windows 10+, Linux, macOS

### 本地开发
```bash
# 克隆项目
git clone https://github.com/Kevin-Rudy/goping.git
cd goping

# 安装依赖
go mod download

# 运行测试
go test ./...

# 本地构建
go build -o goping cmd/goping/*.go

# 开发模式运行
go run cmd/goping/*.go google.com
```

### 项目结构说明
- 新增功能请遵循三层架构设计
- 平台相关代码请使用构建标签分离
- TUI相关修改请确保时间对齐逻辑的正确性
- 添加新的配置项时请同时更新验证逻辑

## 📄 许可证

本项目基于 MIT 许可证开源 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

灵感来源于 [gping](https://github.com/orf/gping) 项目，使用Go语言重新实现并增加了智能权限适配功能。

