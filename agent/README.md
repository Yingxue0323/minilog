# MiniLog Agent

Lightweight monitoring agent written in Go.

## Features

- ✅ Pure Go implementation
- ✅ Single binary (~6 MB compiled)
- ✅ Cross-platform (Linux, macOS, Windows)
- ✅ Ultra-lightweight metrics (~50 bytes per push)

## Build

```bash
cd agent
go build -ldflags="-s -w" -o minilog-agent
```

## Usage

```bash
# Basic usage
./minilog-agent --server web-01 --minilog http://192.168.1.100:8080

# Custom interval
./minilog-agent --server api-02 --minilog http://minilog:8080 --interval 60

# Use hostname as server name
./minilog-agent --minilog http://localhost:8080
```

## Metrics Collected

- **CPU Usage** (%)
- **Memory Usage** (%)
- **Disk Usage** (%)
- **System Load** (1-minute average)

Total size: ~50 bytes per push

## Dependencies

- github.com/shirou/gopsutil/v3

## License

MIT
