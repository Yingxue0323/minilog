package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pierrec/lz4/v4"
)

type LogEntry struct {
	Timestamp string   `json:"timestamp"`
	Level     string   `json:"level"`
	Server    string   `json:"server"`
	Message   string   `json:"message"`
	Metrics   *Metrics `json:"metrics,omitempty"` // 可选的监控指标
}

// 日志存储引擎（核心）
type LogStorage struct {
	// 内存缓冲区（最新的日志，未压缩）
	memoryBuffer []LogEntry
	bufferMu     sync.RWMutex
	
	// 配置参数
	maxBufferSize   int           // 最大缓冲条数
	maxBufferMemory int64         // 最大缓冲内存（字节）
	flushInterval   time.Duration // 刷盘间隔
	dataDir         string
	
	// 统计信息
	stats struct {
		TotalReceived   int64
		TotalCompressed int64
		CompressionRatio float64
	}
}

func NewLogStorage(dataDir string) *LogStorage {
	os.MkdirAll(dataDir, 0755)
	
	storage := &LogStorage{
		memoryBuffer:    make([]LogEntry, 0, 1000),
		maxBufferSize:   1000,                 // 攒够1000条就压缩
		maxBufferMemory: 10 * 1024 * 1024,     // 或者超过10MB就压缩
		flushInterval:   60 * time.Second,     // 或者超过60秒就压缩
		dataDir:         dataDir,
	}
	
	// 启动后台定时压缩任务
	go storage.backgroundFlusher()
	
	return storage
}

// 接收日志（实时写入内存）
func (s *LogStorage) Append(log LogEntry) {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()
	
	// 添加到内存缓冲
	s.memoryBuffer = append(s.memoryBuffer, log)
	s.stats.TotalReceived++
	
	// 检查是否需要立即压缩（条件触发）
	if len(s.memoryBuffer) >= s.maxBufferSize {
		go s.flushToDisk() // 异步压缩，不阻塞接收
	}
}

// 后台定时任务（定时压缩）
func (s *LogStorage) backgroundFlusher() {
	ticker := time.NewTicker(s.flushInterval)
	for range ticker.C {
		s.flushToDisk()
	}
}

// 压缩并写入磁盘
func (s *LogStorage) flushToDisk() {
	s.bufferMu.Lock()
	
	// 如果缓冲区为空，直接返回
	if len(s.memoryBuffer) == 0 {
		s.bufferMu.Unlock()
		return
	}
	
	// 取出缓冲区数据（快速释放锁）
	logsToCompress := make([]LogEntry, len(s.memoryBuffer))
	copy(logsToCompress, s.memoryBuffer)
	s.memoryBuffer = s.memoryBuffer[:0] // 清空缓冲区
	
	s.bufferMu.Unlock()
	
	// 下面的操作不持有锁，不影响新日志写入
	
	// 1. 序列化为文本
	var plainText bytes.Buffer
	for _, log := range logsToCompress {
		line := fmt.Sprintf("[%s] [%s] [%s] %s\n",
			log.Timestamp, log.Level, log.Server, log.Message)
		plainText.WriteString(line)
	}
	
	// 2. LZ4压缩
	var compressed bytes.Buffer
	writer := lz4.NewWriter(&compressed)
	writer.Write(plainText.Bytes())
	writer.Close()
	
	// 3. 写入文件（按小时分片）
	hour := time.Now().Format("2006-01-02-15")
	filename := fmt.Sprintf("%s/logs-%s.lz4", s.dataDir, hour)
	
	f, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	
	// 写入分隔符（方便后续分块读取）
	separator := fmt.Sprintf("===CHUNK_%d===\n", time.Now().Unix())
	f.WriteString(separator)
	f.Write(compressed.Bytes())
	f.Close()
	
	// 4. 更新统计
	s.stats.TotalCompressed += int64(len(logsToCompress))
	originalSize := plainText.Len()
	compressedSize := compressed.Len()
	ratio := float64(originalSize) / float64(compressedSize)
	s.stats.CompressionRatio = ratio
	
	fmt.Printf("💾 [Compressed] %d logs | %d B → %d B | Ratio %.1f:1 | File: %s\n",
		len(logsToCompress), originalSize, compressedSize, ratio, filename)
}

// 查询日志（内存 + 磁盘）支持多维度筛选
func (s *LogStorage) Query(keyword, server, level string, limit int) []LogEntry {
	results := make([]LogEntry, 0)
	keywordLower := strings.ToLower(keyword)
	serverLower := strings.ToLower(server)
	levelLower := strings.ToLower(level)
	
	// 1. 先查内存（最新的未压缩数据）
	s.bufferMu.RLock()
	for i := len(s.memoryBuffer) - 1; i >= 0 && len(results) < limit; i-- {
		log := s.memoryBuffer[i]
		if s.matchLogWithFilters(log, keywordLower, serverLower, levelLower) {
			results = append(results, log)
		}
	}
	s.bufferMu.RUnlock()
	
	// 如果内存中已经够了，直接返回
	if len(results) >= limit {
		return results
	}
	
	// 2. 再查磁盘（压缩的历史数据）
	diskResults := s.queryDisk(keywordLower, serverLower, levelLower, limit-len(results))
	results = append(results, diskResults...)
	
	return results
}

// 多维度匹配（支持关键字、服务器、级别筛选）
func (s *LogStorage) matchLogWithFilters(log LogEntry, keyword, server, level string) bool {
	// 关键字匹配
	if keyword != "" {
		matchKeyword := strings.Contains(strings.ToLower(log.Message), keyword) ||
			strings.Contains(strings.ToLower(log.Level), keyword) ||
			strings.Contains(strings.ToLower(log.Server), keyword)
		if !matchKeyword {
			return false
		}
	}
	
	// 服务器匹配
	if server != "" && strings.ToLower(log.Server) != server {
		return false
	}
	
	// 级别匹配
	if level != "" && strings.ToLower(log.Level) != level {
		return false
	}
	
	return true
}

func (s *LogStorage) queryDisk(keyword, server, level string, limit int) []LogEntry {
	results := make([]LogEntry, 0)
	
	// 读取当前小时的压缩文件
	hour := time.Now().Format("2006-01-02-15")
	filename := fmt.Sprintf("%s/logs-%s.lz4", s.dataDir, hour)
	
	data, err := os.ReadFile(filename)
	if err != nil {
		return results
	}
	
	// 分块解压（按===CHUNK===分隔）
	chunks := bytes.Split(data, []byte("===CHUNK_"))
	
	for _, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}
		
		// 跳过时间戳行
		idx := bytes.Index(chunk, []byte("\n"))
		if idx == -1 {
			continue
		}
		chunk = chunk[idx+1:]
		
		// 解压
		reader := lz4.NewReader(bytes.NewReader(chunk))
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			continue
		}
		
		// 解析日志行
		lines := strings.Split(string(decompressed), "\n")
		for i := len(lines) - 1; i >= 0 && len(results) < limit; i-- {
			line := lines[i]
			if line == "" {
				continue
			}
			
			// 解析日志
			log := parseLogLine(line)
			
			// 多维度筛选
			if s.matchLogWithFilters(log, keyword, server, level) {
				results = append(results, log)
			}
		}
		
		if len(results) >= limit {
			break
		}
	}
	
	return results
}

func parseLogLine(line string) LogEntry {
	// 简单解析 [时间] [级别] [服务器] 消息
	// 生产环境应该更健壮
	return LogEntry{
		Timestamp: extractBracket(line, 0),
		Level:     extractBracket(line, 1),
		Server:    extractBracket(line, 2),
		Message:   line,
	}
}

func extractBracket(s string, index int) string {
	count := 0
	start := -1
	for i, c := range s {
		if c == '[' {
			if count == index {
				start = i + 1
			}
			count++
		} else if c == ']' && start != -1 {
			return s[start:i]
		}
	}
	return ""
}

// 获取统计信息
func (s *LogStorage) GetStats() map[string]interface{} {
	s.bufferMu.RLock()
	defer s.bufferMu.RUnlock()
	
	// 统计所有不同的服务器
	servers := make(map[string]bool)
	for _, log := range s.memoryBuffer {
		if log.Server != "" {
			servers[log.Server] = true
		}
	}
	
	serverList := make([]string, 0, len(servers))
	for server := range servers {
		serverList = append(serverList, server)
	}
	
	return map[string]interface{}{
		"total_received":    s.stats.TotalReceived,
		"total_compressed":  s.stats.TotalCompressed,
		"in_memory":         len(s.memoryBuffer),
		"compression_ratio": fmt.Sprintf("%.1f:1", s.stats.CompressionRatio),
		"servers":           serverList,
	}
}

func main() {
	storage := NewLogStorage("data")
	metricsStorage := NewMetricsStorage("data", 120) // 每台服务器保留120个数据点（1小时）
	
	// API: 接收日志（实时写入内存）
	http.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "只接受POST", http.StatusMethodNotAllowed)
			return
		}
		
		body, _ := io.ReadAll(r.Body)
		var log LogEntry
		
		if err := json.Unmarshal(body, &log); err != nil {
			log = LogEntry{
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				Message:   string(body),
			}
		}
		
		if log.Timestamp == "" {
			log.Timestamp = time.Now().Format("2006-01-02 15:04:05")
		}
		
		// 实时追加日志到内存
		storage.Append(log)
		
		// 如果包含监控指标，存储到 metricsStorage
		// 任何带 server 的日志都会更新服务器状态（基于最后推送时间）
		if log.Metrics != nil && log.Server != "" {
			metricsEntry := MetricsEntry{
				Timestamp: log.Timestamp,
				Server:    log.Server,
				Metrics:   *log.Metrics,
			}
			metricsStorage.Append(metricsEntry)
		}
		
		fmt.Fprintf(w, "✓ Received")
	})
	
	// API: 查询日志（内存+磁盘，支持多维度筛选）
	http.HandleFunc("/api/query", func(w http.ResponseWriter, r *http.Request) {
		keyword := r.URL.Query().Get("keyword")
		server := r.URL.Query().Get("server")
		level := r.URL.Query().Get("level")
		
		// 查询最新的1000条（内存+磁盘）
		results := storage.Query(keyword, server, level, 1000)
		
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		
		for i := len(results) - 1; i >= 0; i-- {
			log := results[i]
			fmt.Fprintf(w, "[%s] [%s] [%s] %s\n",
				log.Timestamp, log.Level, log.Server, log.Message)
		}
	})
	
	// API: 统计信息
	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// 合并日志和监控统计
		logStats := storage.GetStats()
		metricsStats := metricsStorage.GetStats()
		
		combined := make(map[string]interface{})
		for k, v := range logStats {
			combined[k] = v
		}
		for k, v := range metricsStats {
			combined[k] = v
		}
		
		json.NewEncoder(w).Encode(combined)
	})
	
	// API: 查询监控数据
	http.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		server := r.URL.Query().Get("server")
		metricName := r.URL.Query().Get("metric")
		
		// 默认返回最近 120 个点（1小时）
		limit := 120
		
		w.Header().Set("Content-Type", "application/json")
		
		if server == "" {
			// 返回所有服务器的最新数据
			results := metricsStorage.Query("", metricName, 1)
			json.NewEncoder(w).Encode(results)
		} else {
			// 返回指定服务器的时序数据
			results := metricsStorage.Query(server, metricName, limit)
			json.NewEncoder(w).Encode(results)
		}
	})
	
	// API: 服务器状态
	http.HandleFunc("/api/servers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		servers := metricsStorage.GetServerStatus()
		json.NewEncoder(w).Encode(servers)
	})
	
	// API: 服务器聚合统计
	http.HandleFunc("/api/metrics/summary", func(w http.ResponseWriter, r *http.Request) {
		server := r.URL.Query().Get("server")
		
		w.Header().Set("Content-Type", "application/json")
		
		if server == "" {
			// 返回所有服务器的摘要
			servers := metricsStorage.GetServerStatus()
			summaries := make([]map[string]interface{}, 0)
			for _, s := range servers {
				if summary := metricsStorage.GetAggregatedStats(s.Server); summary != nil {
					summaries = append(summaries, summary)
				}
			}
			json.NewEncoder(w).Encode(summaries)
		} else {
			// 返回指定服务器的摘要
			summary := metricsStorage.GetAggregatedStats(server)
			json.NewEncoder(w).Encode(summary)
		}
	})
	
	// 静态文件服务（前端页面）
	http.Handle("/", http.FileServer(http.Dir("static")))
	
	fmt.Println("🚀 MiniLog Lightweight Monitoring Version Started!")
	fmt.Println("📊 Web UI: http://localhost:8080")
	fmt.Println("📡 Receive Logs: POST http://localhost:8080/api/logs")
	fmt.Println("📈 Lightweight Metrics: CPU, Memory, Disk, Load (~50 bytes per push)")
	fmt.Println("💾 Smart Compression: Triggers at 1000 logs or 1 minute")
	fmt.Println("🔍 Query Strategy: Memory first → Disk fallback")
	fmt.Println("📉 Monitoring: No heartbeat, status based on log push time")
	http.ListenAndServe(":8080", nil)
}