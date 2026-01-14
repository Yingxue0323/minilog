package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

// 日志条目（与 MiniLog 服务器保持一致）
type LogEntry struct {
	Timestamp string   `json:"timestamp"`
	Level     string   `json:"level"`
	Server    string   `json:"server"`
	Message   string   `json:"message"`
	Metrics   *Metrics `json:"metrics,omitempty"`
}

// 轻量级监控指标
type Metrics struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskPercent   float64 `json:"disk_percent"`
	LoadAvg       float64 `json:"load_avg"`
}

// Agent 配置
type Agent struct {
	serverName string
	minilogURL string
	interval   int
}

func main() {
	// 命令行参数
	serverName := flag.String("server", "", "服务器名称（默认使用主机名）")
	minilogURL := flag.String("minilog", "http://localhost:8080", "MiniLog 服务器地址")
	interval := flag.Int("interval", 30, "采集间隔（秒）")
	flag.Parse()

	// 创建 Agent
	agent := &Agent{
		serverName: *serverName,
		minilogURL: *minilogURL,
		interval:   *interval,
	}

	// 如果未指定服务器名称，使用主机名
	if agent.serverName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal("无法获取主机名:", err)
		}
		agent.serverName = hostname
	}

	fmt.Println("🚀 MiniLog Agent 启动")
	fmt.Println("📡 服务器名称:", agent.serverName)
	fmt.Println("🌐 MiniLog URL:", agent.minilogURL)
	fmt.Println("⏱  采集间隔:", agent.interval, "秒")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("📊 开始采集监控数据...")
	fmt.Println()

	// 主循环
	agent.run()
}

func (a *Agent) run() {
	ticker := time.NewTicker(time.Duration(a.interval) * time.Second)
	defer ticker.Stop()

	// 立即采集一次
	a.collectAndSend()

	// 定期采集
	for range ticker.C {
		a.collectAndSend()
	}
}

func (a *Agent) collectAndSend() {
	metrics, err := a.collectMetrics()
	if err != nil {
		log.Printf("❌ 采集失败: %v\n", err)
		return
	}

	if err := a.sendToMiniLog(metrics); err != nil {
		log.Printf("⚠️  推送失败: %v\n", err)
		return
	}

	// 成功输出
	fmt.Printf("✅ [%s] CPU: %5.1f%% | 内存: %5.1f%% | 磁盘: %5.1f%% | 负载: %.2f\n",
		time.Now().Format("15:04:05"),
		metrics.CPUPercent,
		metrics.MemoryPercent,
		metrics.DiskPercent,
		metrics.LoadAvg,
	)
}

func (a *Agent) collectMetrics() (*Metrics, error) {
	metrics := &Metrics{}

	// 1. CPU 使用率（采集 1 秒）
	cpuPercents, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("获取 CPU 失败: %w", err)
	}
	if len(cpuPercents) > 0 {
		metrics.CPUPercent = round(cpuPercents[0], 2)
	}

	// 2. 内存使用率
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("获取内存失败: %w", err)
	}
	metrics.MemoryPercent = round(memInfo.UsedPercent, 2)

	// 3. 磁盘使用率
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("获取磁盘失败: %w", err)
	}
	metrics.DiskPercent = round(diskInfo.UsedPercent, 2)

	// 4. 系统负载（1分钟）
	if runtime.GOOS != "windows" { // Windows 不支持 load average
		loadInfo, err := load.Avg()
		if err == nil {
			metrics.LoadAvg = round(loadInfo.Load1, 2)
		}
	}

	return metrics, nil
}

func (a *Agent) sendToMiniLog(metrics *Metrics) error {
	// 构造日志条目
	entry := LogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "METRICS",
		Server:    a.serverName,
		Message:   fmt.Sprintf("系统指标 - CPU: %.1f%% | 内存: %.1f%%", metrics.CPUPercent, metrics.MemoryPercent),
		Metrics:   metrics,
	}

	// JSON 序列化
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	// HTTP 请求
	url := a.minilogURL + "/api/logs"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误: %d", resp.StatusCode)
	}

	return nil
}

// 辅助函数：四舍五入
func round(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
