package main

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"
)

// 监控指标数据结构（轻量化版本）
type Metrics struct {
	CPUPercent    float64 `json:"cpu_percent"`     // CPU 使用率 %
	MemoryPercent float64 `json:"memory_percent"`  // 内存使用率 %
	DiskPercent   float64 `json:"disk_percent"`    // 磁盘使用率 %
	LoadAvg       float64 `json:"load_avg"`        // 系统负载（1分钟）
}

// 监控数据点（带时间戳）
type MetricsEntry struct {
	Timestamp string  `json:"timestamp"`
	Server    string  `json:"server"`
	Metrics   Metrics `json:"metrics"`
}

// 服务器状态
type ServerStatus struct {
	Server     string       `json:"server"`
	Status     string       `json:"status"` // online, offline, timeout
	LastSeen   string       `json:"last_seen"`
	Latest     *Metrics     `json:"latest,omitempty"`
	Uptime     string       `json:"uptime,omitempty"`
}

// 监控存储引擎
type MetricsStorage struct {
	// 内存中保留最近的数据点（每台服务器最多 120 个点 = 1小时，30秒间隔）
	recentMetrics map[string][]MetricsEntry // server -> []metrics
	metricsMu     sync.RWMutex
	
	// 服务器状态表
	serverStatus map[string]*ServerStatus
	statusMu     sync.RWMutex
	
	// 配置
	maxPointsPerServer int           // 每台服务器最多保留多少个数据点
	offlineThreshold   time.Duration // 多久未收到数据算离线
	dataDir            string
	
	// 统计
	stats struct {
		TotalMetricsReceived int64
		ActiveServers        int
	}
}

// 创建监控存储引擎
func NewMetricsStorage(dataDir string, maxPoints int) *MetricsStorage {
	storage := &MetricsStorage{
		recentMetrics:      make(map[string][]MetricsEntry),
		serverStatus:       make(map[string]*ServerStatus),
		maxPointsPerServer: maxPoints,
		offlineThreshold:   90 * time.Second, // 90秒未推送视为离线
		dataDir:            dataDir,
	}
	
	// 启动后台任务：定期持久化数据（每小时）
	go storage.persistMetrics()
	
	return storage
}

// 接收监控数据
func (m *MetricsStorage) Append(entry MetricsEntry) {
	m.metricsMu.Lock()
	defer m.metricsMu.Unlock()
	
	server := entry.Server
	if server == "" {
		return
	}
	
	// 添加到内存
	if _, exists := m.recentMetrics[server]; !exists {
		m.recentMetrics[server] = make([]MetricsEntry, 0, m.maxPointsPerServer)
	}
	
	m.recentMetrics[server] = append(m.recentMetrics[server], entry)
	
	// 保持固定长度（滚动窗口）
	if len(m.recentMetrics[server]) > m.maxPointsPerServer {
		m.recentMetrics[server] = m.recentMetrics[server][1:]
	}
	
	m.stats.TotalMetricsReceived++
	
	// 更新服务器状态
	m.updateServerStatus(server, entry)
}

// 更新服务器状态（基于最后推送时间，无需心跳）
func (m *MetricsStorage) updateServerStatus(server string, entry MetricsEntry) {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()
	
	now := time.Now().Format("2006-01-02 15:04:05")
	
	if status, exists := m.serverStatus[server]; exists {
		status.LastSeen = now
		status.Latest = &entry.Metrics
	} else {
		m.serverStatus[server] = &ServerStatus{
			Server:   server,
			Status:   "online",
			LastSeen: now,
			Latest:   &entry.Metrics,
		}
	}
}

// 计算服务器状态（按需调用，无后台任务）
func (m *MetricsStorage) calculateServerStatus(server string) string {
	m.statusMu.RLock()
	status, exists := m.serverStatus[server]
	m.statusMu.RUnlock()
	
	if !exists {
		return "unknown"
	}
	
	lastSeen, err := time.Parse("2006-01-02 15:04:05", status.LastSeen)
	if err != nil {
		return "unknown"
	}
	
	elapsed := time.Since(lastSeen)
	if elapsed > m.offlineThreshold {
		return "offline"
	} else if elapsed > 60*time.Second {
		return "timeout"
	}
	
	return "online"
}

// 查询指定服务器的监控数据
func (m *MetricsStorage) Query(server string, metricName string, limit int) []MetricsEntry {
	m.metricsMu.RLock()
	defer m.metricsMu.RUnlock()
	
	if server == "" {
		// 返回所有服务器的最新数据点
		result := make([]MetricsEntry, 0)
		for _, entries := range m.recentMetrics {
			if len(entries) > 0 {
				result = append(result, entries[len(entries)-1])
			}
		}
		return result
	}
	
	entries, exists := m.recentMetrics[server]
	if !exists || len(entries) == 0 {
		return []MetricsEntry{}
	}
	
	// 返回最近 N 个点
	start := 0
	if len(entries) > limit {
		start = len(entries) - limit
	}
	
	result := make([]MetricsEntry, len(entries)-start)
	copy(result, entries[start:])
	
	return result
}

// 获取所有服务器状态（实时计算状态）
func (m *MetricsStorage) GetServerStatus() []ServerStatus {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()
	
	result := make([]ServerStatus, 0, len(m.serverStatus))
	for server, status := range m.serverStatus {
		// 实时计算状态
		currentStatus := m.calculateServerStatus(server)
		
		statusCopy := *status
		statusCopy.Status = currentStatus
		result = append(result, statusCopy)
	}
	
	// 按服务器名排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Server < result[j].Server
	})
	
	return result
}

// 获取统计信息
func (m *MetricsStorage) GetStats() map[string]interface{} {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()
	
	activeCount := 0
	for _, status := range m.serverStatus {
		if status.Status == "online" {
			activeCount++
		}
	}
	
	m.stats.ActiveServers = activeCount
	
	return map[string]interface{}{
		"total_metrics_received": m.stats.TotalMetricsReceived,
		"active_servers":         m.stats.ActiveServers,
		"total_servers":          len(m.serverStatus),
	}
}

// 持久化指标数据（每小时保存一次聚合数据）
func (m *MetricsStorage) persistMetrics() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		m.metricsMu.RLock()
		
		// 聚合每台服务器的最近数据
		aggregated := make(map[string]interface{})
		for server, entries := range m.recentMetrics {
			if len(entries) == 0 {
				continue
			}
			
		// 计算平均值
		var sumCPU, sumMemory, sumDisk float64
		var sumLoad float64
		for _, entry := range entries {
			sumCPU += entry.Metrics.CPUPercent
			sumMemory += entry.Metrics.MemoryPercent
			sumDisk += entry.Metrics.DiskPercent
			sumLoad += entry.Metrics.LoadAvg
		}
		
		count := float64(len(entries))
		aggregated[server] = map[string]interface{}{
			"avg_cpu":    sumCPU / count,
			"avg_memory": sumMemory / count,
			"avg_disk":   sumDisk / count,
			"avg_load":   sumLoad / count,
			"samples":    len(entries),
		}
		}
		
		m.metricsMu.RUnlock()
		
		// 保存到文件
		if len(aggregated) > 0 {
			hour := time.Now().Format("2006-01-02-15")
			filename := m.dataDir + "/metrics-" + hour + ".json"
			data, _ := json.MarshalIndent(aggregated, "", "  ")
			os.WriteFile(filename, data, 0644)
		}
	}
}

// 获取指定服务器的聚合统计
func (m *MetricsStorage) GetAggregatedStats(server string) map[string]interface{} {
	m.metricsMu.RLock()
	defer m.metricsMu.RUnlock()
	
	entries, exists := m.recentMetrics[server]
	if !exists || len(entries) == 0 {
		return nil
	}
	
	// 计算统计值
	var sumCPU, sumMemory, sumDisk, sumLoad float64
	var maxCPU, maxMemory, minCPU, minMemory float64
	
	maxCPU = entries[0].Metrics.CPUPercent
	minCPU = entries[0].Metrics.CPUPercent
	maxMemory = entries[0].Metrics.MemoryPercent
	minMemory = entries[0].Metrics.MemoryPercent
	
	for _, entry := range entries {
		sumCPU += entry.Metrics.CPUPercent
		sumMemory += entry.Metrics.MemoryPercent
		sumDisk += entry.Metrics.DiskPercent
		sumLoad += entry.Metrics.LoadAvg
		
		if entry.Metrics.CPUPercent > maxCPU {
			maxCPU = entry.Metrics.CPUPercent
		}
		if entry.Metrics.CPUPercent < minCPU {
			minCPU = entry.Metrics.CPUPercent
		}
		if entry.Metrics.MemoryPercent > maxMemory {
			maxMemory = entry.Metrics.MemoryPercent
		}
		if entry.Metrics.MemoryPercent < minMemory {
			minMemory = entry.Metrics.MemoryPercent
		}
	}
	
	count := float64(len(entries))
	latest := entries[len(entries)-1].Metrics
	
	return map[string]interface{}{
		"server": server,
		"latest": latest,
		"avg": map[string]float64{
			"cpu":    sumCPU / count,
			"memory": sumMemory / count,
			"disk":   sumDisk / count,
			"load":   sumLoad / count,
		},
		"max": map[string]float64{
			"cpu":    maxCPU,
			"memory": maxMemory,
		},
		"min": map[string]float64{
			"cpu":    minCPU,
			"memory": minMemory,
		},
		"samples": len(entries),
	}
}
