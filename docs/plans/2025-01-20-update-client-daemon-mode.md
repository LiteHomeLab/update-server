# Update Client Daemon Mode Design

> **设计日期**: 2025-01-20
> **版本**: 1.0
> **状态**: 设计完成

## 概述

为 update-client 添加 `--daemon` 模式，允许调用程序通过 HTTP API 获取下载进度，解决命令行调用无法实时获取进度的问题。

## 架构

### 运行模式

**默认模式（前台）**：
- 直接运行下载，进度输出到 stdout
- 适合命令行用户直接使用
- 下载完成后自动退出

**Daemon 模式（后台）**：
- 启动 HTTP 服务器暴露状态
- 通过 HTTP API 获取进度
- 等待调用者主动关闭
- 监控父进程，父进程退出时自动终止

### 调用流程

```
主程序
  │
  ├── 1. 启动: update-client.exe download --daemon --port 19876 --version 1.2.0
  │
  ├── 2. 定期轮询: GET http://localhost:19876/status
  │     ← 返回: {"state": "downloading", "progress": {...}}
  │
  ├── 3. 检测完成: GET http://localhost:19876/status
  │     ← 返回: {"state": "completed", "file": "..."}
  │
  └── 4. 关闭: POST http://localhost:19876/shutdown
  │
download 进程: 检测父进程存活，父进程退出时自动终止
```

### 状态转换

```
idle → downloading → completed → (等待 shutdown)
  ↓
error
```

## HTTP API 设计

### GET /status

返回当前下载状态：

```json
{
  "state": "downloading",        // idle | downloading | completed | error
  "version": "1.2.0",
  "file": "./updates/app-v1.2.0.zip",
  "progress": {
    "downloaded": 52428800,     // 已下载字节数
    "total": 104857600,          // 总字节数
    "percentage": 50.0,           // 下载百分比
    "speed": 8912896              // 速度 (bytes/s)
  },
  "error": ""                     // 错误信息（仅在 state=error 时）
}
```

**状态说明：**
- `idle`: 下载尚未开始
- `downloading`: 正在下载
- `completed`: 下载完成
- `error`: 发生错误

### POST /shutdown

关闭下载进程和 HTTP 服务器：

请求：
```json
{
  "reason": "download_completed"  // 可选：关闭原因
}
```

响应：
```json
{
  "success": true,
  "message": "Server shutting down"
}
```

**错误响应：**

端口冲突：
```json
{
  "success": false,
  "error": "port_already_in_use",
  "message": "Port 19876 is already in use. Try a different port."
}
```

重复关闭：
```json
{
  "success": false,
  "error": "already_shutting_down",
  "message": "Already shutting down"
}
```

服务器已关闭：
- 连接被拒绝（TCP 错误）

## 命令行参数

### 新增参数

| 参数 | 说明 | 必填 |
|------|------|------|
| `--daemon` | 启用 daemon 模式（HTTP 服务器） | 是 |
| `--port` | HTTP 服务器端口 | 是 |
| `--version` | 要下载的版本 | 是 |
| `--output` | 输出文件路径（可选） | 否 |

### 行为差异

| 特性 | 默认模式 | Daemon 模式 |
|------|----------|-------------|
| 进度输出 | stdout 实时显示进度条 | 无输出（通过 /status 获取） |
| 退出时机 | 下载完成后自动退出 | 父进程退出或 /shutdown 命令 |
| 父进程检测 | 不检测 | 每 5 秒检测一次 |

### 使用示例

```bash
# 默认模式（前台）
update-client.exe download --version 1.2.0

# Daemon 模式（后台）
update-client.exe download --daemon --port 19876 --version 1.2.0

# 指定输出路径
update-client.exe download --daemon --port 19876 --version 1.2.0 --output "C:\updates\app.zip"
```

## 实现细节

### HTTP 服务器

使用标准 `net/http` 库，实现轻量级 HTTP 服务器：

```go
type DaemonServer struct {
    port       int
    server     *http.Server
    downloader *DownloadManager
    done       chan struct{}
}

func (d *DaemonServer) Start() error {
    mux := http.NewServeMux()
    mux.HandleFunc("/status", d.handleStatus)
    mux.HandleFunc("/shutdown", d.handleShutdown)

    d.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", d.port),
        Handler: mux,
    }

    return d.server.ListenAndServe()
}
```

### 父进程监控

跨平台父进程检测：

```go
func (d *DaemonServer) monitorParentProcess() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if !d.isParentAlive() {
                log.Println("Parent process died, shutting down")
                d.shutdown()
                return
            }
        case <-d.done:
            return
        }
    }
}
```

### 下载状态管理

线程安全的状态管理：

```go
type DownloadState struct {
    State       string        `json:"state"`
    Version     string        `json:"version"`
    File        string        `json:"file"`
    Progress    *ProgressInfo `json:"progress,omitempty"`
    Error       string        `json:"error,omitempty"`
    mu          sync.RWMutex
}

type ProgressInfo struct {
    Downloaded int64   `json:"downloaded"`
    Total      int64   `json:"total"`
    Percentage float64 `json:"percentage"`
    Speed      int64   `json:"speed"`
}
```

## 错误处理

### 端口冲突

启动时检测端口：
- 端口可用 → 正常启动
- 端口被占用 → 立即退出，返回错误码 1

### 下载失败

- 下载失败时状态变为 `error`
- `/status` 返回错误详情
- HTTP 服务器保持运行
- 调用者获取错误后发送 `/shutdown`

### 重复关闭

- 第一个 `/shutdown`：开始关闭流程
- 第二个 `/shutdown`：返回 `already_shutting_down`
- 关闭完成后：停止接受新连接

### 超时保护

调用者可以通过以下方式处理超时：
1. 发送 `/shutdown` 强制终止
2. 依赖父进程检测机制自动终止（5-10 秒后检测到父进程死亡）

## 集成示例

### C# / WPF

```csharp
public class UpdateDownloader : IDisposable
{
    private Process _process;
    private HttpClient _httpClient;

    public async Task<bool> StartDownloadAsync(string version, string outputPath)
    {
        int port = GetAvailablePort(); // 19876-19880

        var psi = new ProcessStartInfo {
            FileName = "update-client.exe",
            Arguments = $"download --daemon --port {port} --version {version} --output \"{outputPath}\"",
            UseShellExecute = false,
            CreateNoWindow = true
        };

        _process = Process.Start(psi);
        await Task.Delay(1000); // 等待 HTTP 服务器启动

        _httpClient = new HttpClient {
            BaseAddress = new Uri($"http://localhost:{port}")
        };
        return true;
    }

    public async Task<DownloadStatus> GetStatusAsync()
    {
        var response = await _httpClient.GetAsync("/status");
        var json = await response.Content.ReadAsStringAsync();
        return JsonSerializer.Deserialize<DownloadStatus>(json);
    }

    public async Task ShutdownAsync()
    {
        await _httpClient.PostAsync("/shutdown", null);
        _process?.WaitForExit(5000);
        _process?.Close();
    }

    public void Dispose()
    {
        ShutdownAsync().Wait();
    }
}
```

### Go

```go
type UpdateDownloader struct {
    cmd    *exec.Cmd
    client *http.Client
    port   int
}

func (d *UpdateDownloader) Start(version, outputPath string) error {
    d.port = getAvailablePort()
    d.cmd = exec.Command("update-client.exe",
        "download", "--daemon",
        "--port", strconv.Itoa(d.port),
        "--version", version,
        "--output", outputPath,
    )

    if err := d.cmd.Start(); err != nil {
        return err
    }

    time.Sleep(time.Second) // 等待 HTTP 服务器启动

    d.client = &http.Client{
        Timeout: 5 * time.Second,
    }

    return nil
}

func (d *UpdateDownloader) GetStatus() (*DownloadStatus, error) {
    resp, err := d.client.Get(fmt.Sprintf("http://localhost:%d/status", d.port))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var status DownloadStatus
    if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
        return nil, err
    }

    return &status, nil
}

func (d *UpdateDownloader) Shutdown() error {
    resp, err := d.client.Post(fmt.Sprintf("http://localhost:%d/shutdown", d.port), "application/json", nil)
    if err != nil {
        return err
    }
    resp.Body.Close()

    d.cmd.Wait()
    return nil
}
```

### Python

```python
import requests
import subprocess
import time

class UpdateDownloader:
    def __init__(self):
        self.process = None
        self.port = None

    def start(self, version, output_path):
        self.port = self.get_available_port()

        cmd = [
            "update-client.exe",
            "download", "--daemon",
            "--port", str(self.port),
            "--version", version,
            "--output", output_path
        ]

        self.process = subprocess.Popen(cmd)
        time.sleep(1)  # 等待 HTTP 服务器启动

    def get_status(self):
        resp = requests.get(f"http://localhost:{self.port}/status")
        return resp.json()

    def shutdown(self):
        requests.post(f"http://localhost:{self.port}/shutdown")
        self.process.wait()
        self.process = None
```

## 端口分配策略

### 推荐策略

调用者负责分配端口：

1. **固定范围**：19876-19880（5个端口）
2. **动态检测**：尝试连接端口，检查是否可用
3. **失败重试**：端口被占用时尝试下一个

### C# 实现

```csharp
private int GetAvailablePort()
{
    for (int port = 19876; port <= 19880; port++) {
        if (IsPortAvailable(port))
            return port;
    }
    throw new Exception("No available port in range 19876-19880");
}

private bool IsPortAvailable(int port)
{
    try {
        var listener = new TcpListener(port);
        listener.Start();
        listener.Stop();
        return true;
    }
    catch (SocketException)
    {
        return false;
    }
}
```

## 使用流程

### 完整集成流程

```
1. 启动下载
   update-client.exe download --daemon --port 19876 --version 1.2.0
   └─> 输出: ✓ Daemon mode started on port 19876

2. 轮询状态
   while (true) {
       status = GET http://localhost:19876/status

       switch (status.state) {
           case "downloading":
               更新 UI 进度条 (status.progress.percentage)
               break
           case "completed":
               file = status.file
               使用文件进行安装
               break
           case "error":
               显示错误 (status.error)
               break
       }

       if (下载完成或错误) break
       sleep(1000)
   }

3. 清理
   POST http://localhost:19876/shutdown
   └─> 输出: ✓ Server stopped
```

### 错误处理

```
启动失败 → 尝试下一个端口
下载失败 → 通过 /status 获取错误 → shutdown → 重试
进程崩溃 → 父进程检测自动清理（5秒内）
```

## 配置文件支持

### 可选配置（update-config.yaml）

```yaml
# daemon 模式配置
daemon:
  port_range: "19876-19880"       # 端口范围
  parent_check_interval: 5        # 父进程检测间隔（秒）
  shutdown_timeout: 5             # shutdown 等待时间（秒）
```

## 注意事项

1. **端口冲突**：确保端口未被占用，或使用端口分配逻辑
2. **进程清理**：确保调用程序在异常情况下也能发送 shutdown
3. **文件锁定**：下载完成后文件才可用，下载中不要访问
4. **超时处理**：设置合理的 HTTP 请求超时（建议 5 秒）
5. **重试机制**：下载失败时可重新启动下载进程

## 未来扩展

可能的功能增强（未包含在初始实现）：

- WebSocket 支持（实时推送进度）
- 多下载队列管理
- 断点续传支持
- 下载历史记录
