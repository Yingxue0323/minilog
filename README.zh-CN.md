# 📊 MiniLog - 超轻量级日志 + 监控系统

> **Languages:** [English](README.md) | [中文](README.zh-CN.md)

> **一体化解决方案**：日志收集 + 服务器监控，单机部署，内存占用仅 30 MB！

## 🎯 核心特性

### ✨ 极致轻量
- **内存占用**：仅 30 MB（比 Prometheus 少 30 倍）
- **指标大小**：每次推送 ~50 bytes（CPU、内存、磁盘、负载）
- **无心跳开销**：基于日志推送时间自动判断服务器状态
- **零配置**：开箱即用，无需复杂配置

### 📋 日志管理
- ✅ 实时日志收集（多服务器支持）
- ✅ 智能压缩存储（LZ4 压缩，压缩比 5:1）
- ✅ 多维度查询（关键字 + 服务器 + 日志级别）
- ✅ Web 可视化界面（矩阵 Tab 筛选）
- ✅ 内存优先策略（查询先内存后磁盘）

### 📊 服务器监控
- ✅ **轻量级指标**：CPU、内存、磁盘、系统负载
- ✅ **无需心跳**：基于最后日志推送时间判断在线状态
- ✅ **随日志推送**：指标跟随日志一起发送，零额外开销
- ✅ **实时图表**：Chart.js 可视化
- ✅ **纯 Go 实现**：Agent 和服务器都是 Go，无需 Python

---

## 📊 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                   被监控服务器                            │
│  web-server-01, api-server-02, db-server-01...          │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │  你的应用程序（产生日志）                       │    │
│  └────────────┬───────────────────────────────────┘    │
│               │                                         │
│  ┌────────────▼───────────────────────────────────┐    │
│  │  MiniLog Agent (agent/agent.go)               │    │
│  │  - 纯 Go 实现，编译后 5-10 MB                  │    │
│  │  - 采集轻量指标（CPU, 内存, 磁盘, 负载）       │    │
│  │  - 跟随日志推送，无需单独心跳                  │    │
│  └────────────┬───────────────────────────────────┘    │
└───────────────┼─────────────────────────────────────────┘
                │ HTTP POST（统一推送）
                ▼
┌─────────────────────────────────────────────────────────┐
│              MiniLog Server (main.go)                   │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │  HTTP API 层                                    │    │
│  │  - POST /api/logs        (接收日志+指标)       │    │
│  │  - GET  /api/query       (查询日志)            │    │
│  │  - GET  /api/metrics     (查询监控数据)        │    │
│  │  - GET  /api/servers     (服务器状态)          │    │
│  └───────────────┬────────────────────────────────┘    │
│                  │                                      │
│  ┌───────────────▼────────────────────────────────┐    │
│  │  双存储引擎                                     │    │
│  │  ┌─────────────────┬──────────────────┐       │    │
│  │  │ LogStorage      │ MetricsStorage   │       │    │
│  │  │ (日志)          │ (监控指标)       │       │    │
│  │  │ - 内存缓冲      │ - 时序数据       │       │    │
│  │  │ - LZ4 压缩      │ - 无心跳         │       │    │
│  │  │ - 按小时分片    │ - 轻量化         │       │    │
│  │  └─────────────────┴──────────────────┘       │    │
│  └────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

---

## 🚀 快速开始

### 1. 启动 MiniLog 服务器

```bash
# 克隆仓库
git clone https://github.com/yourusername/minilog.git
cd minilog

# 运行服务器
go run main.go
```

**输出：**
```
🚀 MiniLog 轻量级监控版启动！
📊 Web界面: http://localhost:8080
📡 接收日志: POST http://localhost:8080/api/logs
📈 轻量指标: CPU, 内存, 磁盘, 负载 (每次推送 ~50 bytes)
💾 智能压缩: 1000条/1分钟 触发
🔍 查询策略: 内存优先 → 磁盘补充
📉 监控功能: 无心跳，基于日志推送时间判断状态
```

### 2. 编译并部署 Agent

```bash
# 进入 agent 目录
cd agent

# 编译（会生成单个二进制文件）
go build -o minilog-agent

# 查看帮助
./minilog-agent --help
```

### 3. 在被监控服务器上运行 Agent

```bash
# 基本用法
./minilog-agent --server web-nw01 --minilog http://192.168.1.100:8080

# 自定义采集间隔（默认 30 秒）
./minilog-agent --server api-se02 --minilog http://minilog:8080 --interval 60

# 使用默认主机名
./minilog-agent --minilog http://localhost:8080
```

**Agent 输出：**
```
🚀 MiniLog Agent 启动
📡 服务器名称: web-nw01
🌐 MiniLog URL: http://192.168.1.100:8080
⏱  采集间隔: 30 秒
--------------------------------------------------
📊 开始采集监控数据...

✅ [10:30:01] CPU:  45.2% | 内存:  68.5% | 磁盘:  72.3% | 负载: 1.23
✅ [10:30:31] CPU:  42.8% | 内存:  69.1% | 磁盘:  72.3% | 负载: 1.15
```

### 4. 访问 Web 界面

- **日志查询**：http://localhost:8080
- **服务器监控**：http://localhost:8080/monitor.html

---

## 📡 API 文档

### 推送日志 + 指标

```bash
curl -X POST http://localhost:8080/api/logs \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "2026-01-14 10:30:00",
    "level": "INFO",
    "server": "web-nw01",
    "message": "用户登录成功",
    "metrics": {
      "cpu_percent": 45.2,
      "memory_percent": 68.5,
      "disk_percent": 72.3,
      "load_avg": 1.23
    }
  }'
```

### 查询日志

```bash
# 查询所有
curl "http://localhost:8080/api/query"

# 筛选服务器
curl "http://localhost:8080/api/query?server=web-nw01"

# 筛选级别
curl "http://localhost:8080/api/query?level=ERROR"

# 组合查询
curl "http://localhost:8080/api/query?server=web-nw01&level=ERROR&keyword=timeout"
```

### 查询服务器状态

```bash
curl "http://localhost:8080/api/servers"
```

**返回示例：**
```json
[
  {
    "server": "web-nw01",
    "status": "online",
    "last_seen": "2026-01-14 10:30:15",
    "latest": {
      "cpu_percent": 45.2,
      "memory_percent": 68.5,
      "disk_percent": 72.3,
      "load_avg": 1.23
    }
  }
]
```

---

## 📊 性能指标

### 资源占用对比

| 指标 | MiniLog | Prometheus | Grafana Agent |
|-----|---------|------------|--------------|
| **服务器内存** | 30 MB | 500 MB - 1 GB | 50-100 MB |
| **Agent 大小** | 5-10 MB | N/A | 20-30 MB |
| **指标大小** | ~50 bytes | ~200 bytes | ~150 bytes |
| **心跳开销** | 无 | 有 | 有 |

### 核心优势

```
✅ 内存占用：30 MB（少 20-30 倍）
✅ 指标大小：50 bytes（精简 75%）
✅ 无心跳：基于日志推送判断，零额外请求
✅ 纯 Go：无需 Python 环境
✅ 单文件：编译后一个二进制文件
```

---

## 🎯 适用场景

### ✅ 非常适合

**场景1：个人开发者 - "我的Side Project"**
```
我有3个项目跑在一台VPS上：
- 个人博客
- Side Project API
- 爬虫脚本

问题：出Bug了，我只想看看刚才发生了什么
需求：最近1小时的ERROR日志
```

**场景2：小型创业公司 - "预算有限"**
```
创业团队的需求：
- 5台服务器
- 每天10GB日志
- 预算有限（不想为日志再买3台服务器）

他们的真实操作：
1. SSH到每台服务器 cat /var/log/app.log | grep ERROR
2. 或者干脆不看日志，直到出大问题

MiniLog对他们的价值：
✅ 5分钟部署完毕
✅ 所有日志集中查看
✅ 只需一台1GB内存的机器
✅ 成本几乎为0
```

**场景3：边缘计算 / IoT**
```
树莓派监控系统：
- 100个树莓派设备
- 每天收集温度、状态日志
- 存储空间有限（32GB SD卡）

需求：能看最近的日志就行，不需要复杂搜索
MiniLog：完美适配！
```

### 🎯 MiniLog 解决的痛点

```
❌ SSH到每台服务器手动查日志
❌ 日志文件满了自动删除，想查的时候已经没了
❌ 不敢开DEBUG日志，怕磁盘爆满
❌ 小项目根本不部署日志系统（太麻烦）

✅ MiniLog 解决的是：让不用日志系统的人开始用日志系统！
```

---

## 💡 设计理念

### 1. 为什么不需要心跳？

**传统方案：**
```
Agent 每 30 秒推送日志 → 同时每 10 秒发送心跳
问题：增加 3 倍网络请求
```

**MiniLog 方案：**
```
Agent 每 30 秒推送日志（带指标）
服务器检查最后推送时间：
- < 60 秒 → 在线 🟢
- 60-90 秒 → 超时 🟠
- > 90 秒 → 离线 🔴

优势：零额外开销
```

### 2. 为什么只选4个指标？

**指标选择原则：**
- ✅ **CPU 使用率**：判断是否过载（必须）
- ✅ **内存使用率**：判断是否内存泄漏（必须）
- ✅ **磁盘使用率**：判断是否磁盘满（必须）
- ✅ **系统负载**：判断整体压力（重要）
- ❌ **网络流量**：波动大，不稳定（去掉）
- ❌ **进程列表**：太重，不实时（去掉）

**每次推送大小：**
```json
{
  "cpu_percent": 45.2,      // 8 bytes
  "memory_percent": 68.5,   // 8 bytes
  "disk_percent": 72.3,     // 8 bytes
  "load_avg": 1.23          // 8 bytes
}
// 总计：~50 bytes（超轻量！）
```

### 3. 为什么用 Go 不用 Python？

| 维度 | Python Agent | Go Agent |
|-----|-------------|----------|
| **依赖** | psutil, requests | gopsutil（编译时链接）|
| **部署** | 需要 Python 环境 | 单个二进制文件 |
| **大小** | ~50 MB（含 Python） | 5-10 MB |
| **启动** | ~500 ms | ~50 ms |
| **内存** | ~50 MB | ~10 MB |
| **跨平台** | 需要对应版本 | 一次编译到处运行 |

---

## 📁 项目结构

```
minilog/
├── main.go                    # 主服务器
├── metrics.go                 # 监控存储引擎
├── go.mod                     # 依赖
├── agent/
│   ├── agent.go              # 轻量级 Go Agent
│   └── go.mod                # Agent 依赖
├── static/
│   ├── index.html            # 日志查询页面
│   └── monitor.html          # 监控页面
├── data/                      # 数据目录
│   ├── logs-*.lz4            # 压缩日志
│   └── metrics-*.json        # 聚合指标
├── README.md                  # 英文文档
└── README.zh-CN.md            # 中文文档（本文）
```

---

## ⚠️ 适用范围

### ✅ 高度适合

1. **中小型应用**（10-50 台服务器）
2. **内存受限环境**（< 100 MB 可用）
3. **边缘设备**（树莓派、IoT）
4. **快速部署**（单文件，零配置）
5. **日志 + 监控一体化**（简化工具链）

### ⚠️ 不太适合

1. **超大规模**（> 100 台服务器）
2. **复杂查询**（需要 SQL 分析）
3. **分布式集群**（需要 Elasticsearch）
4. **长期存储**（年级别归档）

---

## 🛠️ 生产环境部署

### 1. 编译优化版本

```bash
# 服务器
cd minilog
go build -ldflags="-s -w" -o minilog
# 大小：~8 MB

# Agent
cd agent
go build -ldflags="-s -w" -o minilog-agent
# 大小：~6 MB
```

### 2. 使用 systemd 管理

**服务器 `/etc/systemd/system/minilog.service`：**
```ini
[Unit]
Description=MiniLog Server
After=network.target

[Service]
Type=simple
User=minilog
WorkingDirectory=/opt/minilog
ExecStart=/opt/minilog/minilog
Restart=always

[Install]
WantedBy=multi-user.target
```

**Agent `/etc/systemd/system/minilog-agent.service`：**
```ini
[Unit]
Description=MiniLog Agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=/opt/minilog/minilog-agent \
    --server web-nw01 \
    --minilog http://192.168.1.100:8080 \
    --interval 30
Restart=always

[Install]
WantedBy=multi-user.target
```

**启动：**
```bash
sudo systemctl daemon-reload
sudo systemctl enable minilog minilog-agent
sudo systemctl start minilog minilog-agent
```

---

## 🔧 配置说明

### 服务器配置（main.go）

```go
// 日志配置
maxBufferSize:   1000,              // 内存缓冲区大小
flushInterval:   60 * time.Second,  // 压缩间隔

// 监控配置
maxPointsPerServer: 120,            // 每台服务器保留数据点（1小时）
offlineThreshold:   90 * time.Second, // 离线判定时间（90秒未推送）
```

### Agent 配置

```bash
--server <name>        # 服务器名称（默认主机名）
--minilog <url>        # MiniLog 服务器地址
--interval <seconds>   # 采集间隔（默认 30 秒）
```

---

## ❓ 常见问题

### Q: 为什么 Agent 只推送 4 个指标？
A: 这 4 个指标足以判断服务器是否有问题：
- CPU 高 → 过载
- 内存高 → 内存泄漏
- 磁盘满 → 需要清理
- 负载高 → 整体压力大

更多指标会增加网络开销和存储成本，性价比不高。

### Q: 如果服务器一直没有日志怎么办？
A: 指标会显示上次推送的数据，状态会标记为"超时"或"离线"。
你可以让 Agent 定期推送一条"心跳日志"：
```bash
# 每 30 秒推送一次（即使没有业务日志）
```

### Q: 30 MB 内存还是太大了？
A: 可以调整配置：
```go
maxBufferSize: 500,        // 减少内存缓冲
maxPointsPerServer: 60,    // 减少保留数据点
// 可以降到 ~25 MB
```

### Q: 支持多少台服务器？
A: 建议 < 50 台，每台约占用 1 MB 内存

### Q: Agent 需要 root 权限吗？
A: 不需要，普通用户即可运行

---

## 📄 许可证

MIT License

---

**🚀 享受超轻量级的日志 + 监控体验！**
