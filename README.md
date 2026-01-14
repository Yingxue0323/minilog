<h1 align="center">📊 MiniLog - Ultra-Lightweight Log + Monitoring System</h1>

<p align="center">
  <strong>Languages:</strong>
  <a href="README.md">English</a> |
  <a href="README.zh-CN.md">中文</a>
</p>

**All-in-one solution:** Log collection + Server monitoring, single binary, 30MB memory, 5-minute setup
<p align="center">
  <img width="1000" height="770" alt="Screen Shot 2026-01-14 at 16 30 48" src="https://github.com/user-attachments/assets/c879f11f-db9f-435c-ba44-a77fe553a588" />
  <img src="https://github.com/user-attachments/assets/d99e070b-cd33-4b03-8387-19ba0607b762" width="779" />
</p>

## ✨ Why MiniLog?

- 🪶 **30 MB Memory** - 30x lighter than Prometheus
- ⚡ **5-Minute Setup** - Single binary, zero configuration
- 📦 **All-in-One** - Logs + Monitoring in one system
- 🔒 **No Heartbeat** - Status inferred from log push time

## 🎯 Core Features

**Log Management:**
- LZ4 compression (5:1 ratio)
- Multi-dimensional queries (keyword + server + level)
- Memory-first strategy
- Hourly log sharding

**Server Monitoring:**
- 4 essential metrics (CPU, Memory, Disk, Load)
- Real-time charts
- No extra network overhead
- Pure Go implementation

---

## 🚀 Quick Start

### 1. Start MiniLog Server

```bash
# Clone repository
git clone https://github.com/Yingxue0323/minilog.git
cd minilog

# Run server
go run main.go metrics.go
```

### 2. Compile and Deploy Agent

```bash
# Enter agent directory
cd agent

# Compile (generates single binary)
go build -o minilog-agent

# Run on monitored servers
./minilog-agent --server web-01 --minilog http://192.168.1.100:8080
```

---

## 📁 Project Structure

```
minilog/
├── main.go                # Main server
├── metrics.go             # Monitoring storage engine
├── agent/
│   ├── agent.go          # Lightweight Go Agent
│   └── go.mod
├── static/
│   ├── index.html        # Log query page
│   └── monitor.html      # Monitoring page
├── data/                  # Data directory (logs & metrics)
└── README.md
```

---

## 🎯 Perfect For

✅ **Individual Developers**  
Multiple projects on one VPS, quick debugging

✅ **Small Startups (5-50 servers)**  
Limited budget, no dedicated DevOps team

✅ **Edge Computing / IoT**  
Raspberry Pi, embedded devices, limited resources

✅ **Anyone Who**  
SSH into servers to manually `grep` logs

---

## 📊 vs Other Systems

|  | MiniLog | Elasticsearch | Loki |
|--|---------|---------------|------|
| **Setup** | 5 min | 2 hours | 30 min |
| **Memory** | 30 MB | 4 GB | 500 MB |
| **Deployment** | Single binary | Multi-component | Multi-component |

---

## 📄 License

MIT License

---

**🚀 Enjoy the ultra-lightweight log + monitoring experience!**
