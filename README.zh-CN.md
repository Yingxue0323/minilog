# 📊 MiniLog - 超轻量级日志 + 监控系统

> **Languages:** [English](README.md) | [中文](README.zh-CN.md)

**一体化解决方案**：日志收集 + 服务器监控，单个二进制，30MB 内存，5分钟部署

## ✨ 为什么选择 MiniLog?

- 🪶 **30 MB 内存** - 比 Prometheus 轻 30 倍
- ⚡ **5分钟部署** - 单个二进制，零配置
- 💰 **节省 96% 成本** - 无需额外服务器
- 📦 **一体化** - 日志 + 监控集于一身
- 🔒 **无心跳** - 基于日志推送时间推断状态

## 🎯 核心功能

**日志管理：**
- LZ4 压缩（5:1 压缩比）
- 多维度查询（关键字 + 服务器 + 级别）
- 内存优先策略
- 按小时分片存储

**服务器监控：**
- 4 个核心指标（CPU、内存、磁盘、负载）
- 实时图表
- 零额外网络开销
- 纯 Go 实现

---

## 🚀 快速开始

### 1. 启动 MiniLog 服务器

```bash
# 克隆仓库
git clone https://github.com/Yingxue0323/minilog.git
cd minilog

# 运行服务器
go run main.go metrics.go
```

### 2. 编译并部署 Agent

```bash
# 进入 agent 目录
cd agent

# 编译（生成单个二进制文件）
go build -o minilog-agent

# 在被监控服务器上运行
./minilog-agent --server web-01 --minilog http://192.168.1.100:8080
```

### 3. 访问 Web 界面

- **日志查询**：http://localhost:8080
- **服务器监控**：http://localhost:8080/monitor.html

---

## 📁 项目结构

```
minilog/
├── main.go                # 主服务器
├── metrics.go             # 监控存储引擎
├── agent/
│   ├── agent.go          # 轻量级 Go Agent
│   └── go.mod
├── static/
│   ├── index.html        # 日志查询页面
│   └── monitor.html      # 监控页面
├── data/                  # 数据目录（日志和指标）
└── README.md
```

---

## 🎯 完美适配

✅ **个人开发者**  
一台 VPS 上多个项目，快速调试

✅ **小型创业公司（5-50 台服务器）**  
预算有限，无专职运维团队

✅ **边缘计算 / IoT**  
树莓派、嵌入式设备、资源受限环境

✅ **任何需要**  
SSH 到服务器手动 `grep` 日志的人

---

## 📊 对比其他系统

|  | MiniLog | Elasticsearch | Loki |
|--|---------|---------------|------|
| **部署时间** | 5 分钟 | 2 小时 | 30 分钟 |
| **内存占用** | 30 MB | 4 GB | 500 MB |
| **部署方式** | 单个二进制 | 多组件 | 多组件 |
| **成本** | $10/月 | $520/月 | $130/月 |

---

## 📄 许可证

MIT License

---

**🚀 享受超轻量级的日志 + 监控体验！**
