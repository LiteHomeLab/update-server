# Update Client 使用指南

本文档说明如何将 `update-client.exe` 集成到你的应用程序中，实现自动更新功能。

## 概述

`update-client.exe` 是一个独立的命令行工具，可以与任何编程语言的应用程序集成。它负责：
- 检查服务器上是否有新版本
- 下载更新包
- 验证文件完整性（SHA256）
- 自动解密加密的更新包

## 目录结构

应用程序的安装目录应包含：

```
your-app/
├── YourApp.exe              # 你的主程序
├── update-client.exe         # 更新客户端工具
└── update-config.yaml        # 配置文件
```

## 配置文件

`update-config.yaml` 示例：

```yaml
# Update Server Configuration
server:
  url: "http://your-server:8080"              # 服务器地址
  timeout: 30                                  # 请求超时（秒）

program:
  id: "your-app-id"                            # 程序 ID（服务器端分配）
  current_version: "1.0.0"                     # 当前版本号

auth:
  token: "dl_xxxxxxxxxxxxx"                    # Download Token（服务器端分配）
  encryption_key: "base64编码的密钥"           # 加密密钥（如启用）

download:
  save_path: "./updates"                       # 下载保存目录
  naming: "version"                            # 命名方式: version | date | simple
  keep: 3                                      # 保留最近 N 个（仅 date 模式）
  auto_verify: true                            # 自动验证 SHA256

logging:
  level: "info"                                # 日志级别
  file: "update-client.log"                    # 日志文件
```

### 配置说明

| 配置项 | 说明 | 必填 |
|--------|------|------|
| `server.url` | 更新服务器地址 | 是 |
| `server.timeout` | 请求超时时间（秒） | 否 |
| `program.id` | 程序 ID（在服务器创建程序时分配） | 是 |
| `program.current_version` | 当前版本号 | 否 |
| `auth.token` | Download Token | 是 |
| `auth.encryption_key` | 加密密钥（如服务器启用加密） | 条件 |
| `download.save_path` | 下载目录 | 否 |
| `download.naming` | 文件命名方式 | 否 |
| `download.keep` | 保留文件数量 | 否 |

### 命名方式

- `version`: `app-v1.2.0.zip`（推荐，保留所有版本）
- `date`: `app-2024-01-25.zip`（按日期命名）
- `simple`: `app.zip`（总是覆盖）

## 命令使用

### 1. 检查更新

```bash
update-client.exe check [--current-version VERSION] [--json]
```

**参数：**
- `--current-version`: 当前版本号（覆盖配置文件）
- `--json`: 输出 JSON 格式

**默认输出（人类可读）：**
```
✓ Connected to update server
  Latest version: 1.2.0

  Version details:
    Size: 52.4 MB
    Published: 2024-01-25
    Mandatory: No
    Release notes:
      - Bug fixes and improvements
```

**JSON 输出（程序解析）：**
```json
{
  "hasUpdate": true,
  "currentVersion": "1.0.0",
  "latestVersion": "1.2.0",
  "downloadUrl": "/api/download/your-app/stable/1.2.0",
  "fileSize": 52428800,
  "releaseNotes": "Bug fixes and improvements...",
  "publishDate": "2024-01-25T10:30:00Z",
  "mandatory": false
}
```

### 2. 下载更新

```bash
update-client.exe download --version VERSION [--output PATH] [--json]
```

**参数：**
- `--version`: 要下载的版本号（必填）
- `--output`: 输出文件路径（覆盖配置）
- `--json`: 输出 JSON 格式

**默认输出：**
```
✓ Starting download: app-v1.2.0.zip
  Size: 52.4 MB

  Progress: [====================] 100% (52.4/52.4 MB) - 8.5 MB/s

✓ Download completed: ./updates/app-v1.2.0.zip
✓ Verified: SHA256 matches
✓ Decrypted: file ready to use
```

**JSON 输出：**
```json
{
  "success": true,
  "file": "C:\\path\\to\\updates\\app-v1.2.0.zip",
  "fileSize": 52428800,
  "verified": true,
  "decrypted": true
}
```

## Daemon 模式（后台进度监控）

当需要实时监控下载进度时，使用 `--daemon` 参数启动独立的 HTTP 服务器。

### 启动 Daemon 模式

```bash
update-client.exe download --daemon --port 19876 --version 1.2.0
```

输出：
```
✓ Daemon mode started on port 19876
✓ Downloading: app-v1.2.0.zip
✓ Monitoring parent process (PID: 12345)
```

### HTTP API

**GET /status** - 获取下载状态

```json
{
  "state": "downloading",        // idle | downloading | completed | error
  "version": "1.2.0",
  "file": "./updates/app-v1.2.0.zip",
  "progress": {
    "downloaded": 52428800,
    "total": 104857600,
    "percentage": 50.0,
    "speed": 8912896
  },
  "error": ""
}
```

**POST /shutdown** - 关闭服务器

```bash
curl -X POST http://localhost:19876/shutdown
```

响应：
```json
{
  "success": true,
  "message": "Server shutting down"
}
```

### 完整使用流程

1. **启动下载进程**（指定可用端口）
2. **定期轮询状态**（每秒一次）
3. **下载完成后获取文件路径**
4. **发送关闭命令**

### 重要说明

- **端口分配**：调用者负责管理端口（建议范围 19876-19880）
- **父进程监控**：daemon 模式下每 5 秒检测父进程，父进程退出时自动终止
- **错误处理**：下载失败时状态变为 `error`，调用者需主动发送 shutdown
- **超时保护**：若下载卡住，可发送 shutdown 强制终止

## Daemon 模式集成示例

Daemon 模式适用于需要实时显示下载进度的场景。以下是各语言的完整集成示例。

### C# / WPF 应用 - Daemon 模式

```csharp
using System;
using System.Diagnostics;
using System.Net;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

public class UpdateDownloader : IDisposable
{
    private Process _process;
    private HttpClient _httpClient;
    private int _port;
    private readonly string _updateClientPath;

    public UpdateDownloader()
    {
        _updateClientPath = Path.Combine(
            AppDomain.CurrentDomain.BaseDirectory,
            "update-client.exe"
        );
    }

    // 启动下载（Daemon 模式）
    public async Task<bool> StartDownloadAsync(string version, string outputPath)
    {
        _port = GetAvailablePort();

        var psi = new ProcessStartInfo
        {
            FileName = _updateClientPath,
            Arguments = $"download --daemon --port {_port} --version {version} --output \"{outputPath}\"",
            UseShellExecute = false,
            CreateNoWindow = true,
            RedirectStandardOutput = true,
            RedirectStandardError = true
        };

        try
        {
            _process = Process.Start(psi);
            await Task.Delay(1000); // 等待 HTTP 服务器启动

            _httpClient = new HttpClient
            {
                BaseAddress = new Uri($"http://localhost:{_port}"),
                Timeout = TimeSpan.FromSeconds(5)
            };

            // 验证服务器是否启动成功
            var status = await GetStatusAsync();
            return status != null;
        }
        catch
        {
            return false;
        }
    }

    // 获取下载状态
    public async Task<DownloadStatus?> GetStatusAsync()
    {
        try
        {
            var response = await _httpClient.GetAsync("/status");
            if (!response.IsSuccessStatusCode)
                return null;

            var json = await response.Content.ReadAsStringAsync();
            return JsonSerializer.Deserialize<DownloadStatus>(json);
        }
        catch
        {
            return null;
        }
    }

    // 监控下载进度（带回调）
    public async Task<DownloadStatus?> MonitorDownloadAsync(
        Action<DownloadProgress> onProgress,
        CancellationToken cancellationToken = default)
    {
        while (!cancellationToken.IsCancellationRequested)
        {
            var status = await GetStatusAsync();
            if (status == null)
                return null;

            switch (status.State)
            {
                case "downloading":
                    onProgress?.Invoke(status.Progress);
                    break;

                case "completed":
                    return status;

                case "error":
                    return status;
            }

            await Task.Delay(1000, cancellationToken);
        }

        return null;
    }

    // 关闭下载进程
    public async Task ShutdownAsync()
    {
        try
        {
            if (_httpClient != null)
            {
                await _httpClient.PostAsync("/shutdown", null);
            }
        }
        catch { }

        if (_process != null && !_process.HasExited)
        {
            _process.WaitForExit(5000);
            _process.Close();
        }
    }

    // 获取可用端口（19876-19880）
    private int GetAvailablePort()
    {
        for (int port = 19876; port <= 19880; port++)
        {
            if (IsPortAvailable(port))
                return port;
        }
        throw new Exception("No available port in range 19876-19880");
    }

    private bool IsPortAvailable(int port)
    {
        try
        {
            var listener = new System.Net.Sockets.TcpListener(IPAddress.Loopback, port);
            listener.Start();
            listener.Stop();
            return true;
        }
        catch
        {
            return false;
        }
    }

    public void Dispose()
    {
        ShutdownAsync().Wait();
        _httpClient?.Dispose();
        _process?.Dispose();
    }
}

// 数据模型
public class DownloadStatus
{
    public string State { get; set; }        // idle | downloading | completed | error
    public string Version { get; set; }
    public string File { get; set; }
    public DownloadProgress Progress { get; set; }
    public string Error { get; set; }
}

public class DownloadProgress
{
    public long Downloaded { get; set; }
    public long Total { get; set; }
    public double Percentage { get; set; }
    public long Speed { get; set; }
}

// 使用示例
public async Task DownloadUpdateWithProgress()
{
    var downloader = new UpdateDownloader();

    try
    {
        // 启动下载
        if (!await downloader.StartDownloadAsync("1.2.0", "./updates/app.zip"))
        {
            MessageBox.Show("Failed to start download");
            return;
        }

        // 监控进度
        var result = await downloader.MonitorDownloadAsync(progress =>
        {
            // 更新 UI 进度条
            Dispatcher.Invoke(() =>
            {
                progressBar.Value = progress.Percentage;
                speedText.Text = FormatSpeed(progress.Speed);
            });
        });

        if (result?.State == "completed")
        {
            MessageBox.Show($"Download completed: {result.File}");
            // 继续安装流程
        }
        else if (result?.State == "error")
        {
            MessageBox.Show($"Download failed: {result.Error}");
        }
    }
    finally
    {
        await downloader.ShutdownAsync();
    }
}
```

### Go 应用 - Daemon 模式

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

type UpdateDownloader struct {
	cmd    *exec.Cmd
	client *http.Client
	port   int
}

type DownloadStatus struct {
	State    string        `json:"state"`    // idle | downloading | completed | error
	Version  string        `json:"version"`
	File     string        `json:"file"`
	Progress *ProgressInfo `json:"progress,omitempty"`
	Error    string        `json:"error,omitempty"`
}

type ProgressInfo struct {
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Percentage float64 `json:"percentage"`
	Speed      int64   `json:"speed"`
}

// 启动下载
func (d *UpdateDownloader) Start(version, outputPath string) error {
	d.port = d.getAvailablePort()

	d.cmd = exec.Command(
		"update-client.exe",
		"download", "--daemon",
		"--port", strconv.Itoa(d.port),
		"--version", version,
		"--output", outputPath,
	)

	if err := d.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start download process: %w", err)
	}

	// 等待 HTTP 服务器启动
	time.Sleep(time.Second)

	d.client = &http.Client{
		Timeout: 5 * time.Second,
	}

	return nil
}

// 获取状态
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

// 监控下载
func (d *UpdateDownloader) Monitor(callback func(*ProgressInfo)) (*DownloadStatus, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		status, err := d.GetStatus()
		if err != nil {
			return nil, err
		}

		switch status.State {
		case "downloading":
			if status.Progress != nil {
				callback(status.Progress)
			}
		case "completed", "error":
			return status, nil
		}
	}

	return nil, nil
}

// 关闭
func (d *UpdateDownloader) Shutdown() error {
	resp, err := d.client.Post(
		fmt.Sprintf("http://localhost:%d/shutdown", d.port),
		"application/json",
		bytes.NewBuffer(nil),
	)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return d.cmd.Wait()
}

// 获取可用端口
func (d *UpdateDownloader) getAvailablePort() int {
	for port := 19876; port <= 19880; port++ {
		if d.isPortAvailable(port) {
			return port
		}
	}
	panic("no available port in range 19876-19880")
}

func (d *UpdateDownloader) isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// 使用示例
func main() {
	downloader := &UpdateDownloader{}

	// 启动下载
	if err := downloader.Start("1.2.0", "./updates/app.zip"); err != nil {
		fmt.Printf("Failed to start: %v\n", err)
		return
	}
	defer downloader.Shutdown()

	// 监控进度
	status, err := downloader.Monitor(func(p *ProgressInfo) {
		fmt.Printf("Progress: %.1f%% (%.2f MB/s)\n",
			p.Percentage, float64(p.Speed)/(1024*1024))
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if status.State == "completed" {
		fmt.Printf("Download completed: %s\n", status.File)
	} else if status.State == "error" {
		fmt.Printf("Download failed: %s\n", status.Error)
	}
}
```

### Python 应用 - Daemon 模式

```python
import requests
import subprocess
import socket
import time
from typing import Optional, Callable
from dataclasses import dataclass

@dataclass
class DownloadProgress:
    downloaded: int
    total: int
    percentage: float
    speed: int

@dataclass
class DownloadStatus:
    state: str           # idle | downloading | completed | error
    version: str
    file: str
    progress: Optional[DownloadProgress]
    error: str

class UpdateDownloader:
    def __init__(self):
        self.process: Optional[subprocess.Popen] = None
        self.port: Optional[int] = None
        self.base_url: Optional[str] = None

    def start(self, version: str, output_path: str) -> bool:
        """启动下载进程"""
        self.port = self._get_available_port()

        cmd = [
            "update-client.exe",
            "download", "--daemon",
            "--port", str(self.port),
            "--version", version,
            "--output", output_path
        ]

        try:
            self.process = subprocess.Popen(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                creationflags=subprocess.CREATE_NO_WINDOW
            )

            self.base_url = f"http://localhost:{self.port}"
            time.sleep(1)  # 等待 HTTP 服务器启动
            return True
        except Exception as e:
            print(f"Failed to start: {e}")
            return False

    def get_status(self) -> Optional[DownloadStatus]:
        """获取下载状态"""
        try:
            resp = requests.get(f"{self.base_url}/status", timeout=5)
            data = resp.json()

            progress = None
            if data.get("progress"):
                p = data["progress"]
                progress = DownloadProgress(**p)

            return DownloadStatus(
                state=data["state"],
                version=data["version"],
                file=data["file"],
                progress=progress,
                error=data.get("error", "")
            )
        except Exception as e:
            print(f"Failed to get status: {e}")
            return None

    def monitor(self, callback: Callable[[DownloadProgress], None]) -> Optional[DownloadStatus]:
        """监控下载进度"""
        while True:
            status = self.get_status()
            if status is None:
                return None

            if status.state == "downloading":
                if status.progress:
                    callback(status.progress)
            elif status.state in ("completed", "error"):
                return status

            time.sleep(1)

    def shutdown(self):
        """关闭下载进程"""
        try:
            requests.post(f"{self.base_url}/shutdown", timeout=5)
        except:
            pass

        if self.process:
            self.process.wait(timeout=5)
            self.process = None

    def _get_available_port(self) -> int:
        """获取可用端口"""
        for port in range(19876, 19881):
            if self._is_port_available(port):
                return port
        raise Exception("No available port in range 19876-19880")

    def _is_port_available(self, port: int) -> bool:
        """检查端口是否可用"""
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try:
            sock.bind(("127.0.0.1", port))
            sock.close()
            return True
        except:
            return False

    def __del__(self):
        self.shutdown()


# 使用示例
def main():
    downloader = UpdateDownloader()

    # 启动下载
    if not downloader.start("1.2.0", "./updates/app.zip"):
        print("Failed to start download")
        return

    # 进度回调函数
    def on_progress(progress: DownloadProgress):
        print(f"Progress: {progress.percentage:.1f}% "
              f"({progress.downloaded}/{progress.total} bytes) "
              f"({progress.speed / 1024 / 1024:.2f} MB/s)")

    try:
        # 监控下载
        status = downloader.monitor(callback=on_progress)

        if status.state == "completed":
            print(f"\nDownload completed: {status.file}")
        elif status.state == "error":
            print(f"\nDownload failed: {status.error}")
    finally:
        downloader.shutdown()

if __name__ == "__main__":
    main()
```

## 常规模式集成示例

以下是不使用 Daemon 模式的简单集成方式（适合不需要实时进度的场景）。

### C# / WPF 应用

```csharp
using System;
using System.Diagnostics;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

public class UpdateService
{
    private readonly string _updateClientPath;
    private readonly string _currentVersion;

    public UpdateService(string currentVersion)
    {
        _updateClientPath = Path.Combine(
            AppDomain.CurrentDomain.BaseDirectory,
            "update-client.exe"
        );
        _currentVersion = currentVersion;
    }

    // 检查更新
    public async Task<UpdateInfo?> CheckForUpdatesAsync()
    {
        var result = await RunCommandAsync("check", "--current-version", _currentVersion, "--json");

        if (result.ExitCode != 0)
            return null;

        return JsonSerializer.Deserialize<UpdateInfo>(result.Output);
    }

    // 下载更新
    public async Task<string?> DownloadUpdateAsync(string version)
    {
        var result = await RunCommandAsync("download", "--version", version, "--json");

        if (result.ExitCode != 0)
            return null;

        var downloadResult = JsonSerializer.Deserialize<DownloadResult>(result.Output);
        return downloadResult?.File;
    }

    private async Task<CommandResult> RunCommandAsync(params string[] args)
    {
        var psi = new ProcessStartInfo
        {
            FileName = _updateClientPath,
            Arguments = string.Join(" ", args),
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            UseShellExecute = false,
            CreateNoWindow = true
        };

        using var process = Process.Start(psi);
        await process.WaitForExitAsync();

        var output = await process.StandardOutput.ReadToEndAsync();

        return new CommandResult
        {
            ExitCode = process.ExitCode,
            Output = output
        };
    }

    private class CommandResult
    {
        public int ExitCode { get; set; }
        public string Output { get; set; }
    }
}

// 数据模型
public class UpdateInfo
{
    public bool HasUpdate { get; set; }
    public string CurrentVersion { get; set; }
    public string LatestVersion { get; set; }
    public long FileSize { get; set; }
    public string ReleaseNotes { get; set; }
    public bool Mandatory { get; set; }
}

public class DownloadResult
{
    public bool Success { get; set; }
    public string File { get; set; }
    public bool Verified { get; set; }
    public bool Decrypted { get; set; }
}
```

**使用示例：**

```csharp
var updateService = new UpdateService("1.0.0");

// 检查更新
var updateInfo = await updateService.CheckForUpdatesAsync();
if (updateInfo?.HasUpdate == true)
{
    // 提示用户
    var result = MessageBox.Show(
        $"发现新版本 {updateInfo.LatestVersion}，是否下载？\n\n{updateInfo.ReleaseNotes}",
        "更新可用",
        MessageBoxButton.YesNo
    );

    if (result == MessageBoxResult.Yes)
    {
        // 下载更新
        var filePath = await updateService.DownloadUpdateAsync(updateInfo.LatestVersion);
        if (filePath != null)
        {
            // 处理下载的文件（解压、安装等）
            InstallUpdate(filePath);
        }
    }
}
```

### Go 应用

```go
package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type UpdateInfo struct {
    HasUpdate     bool   `json:"hasUpdate"`
    LatestVersion string `json:"latestVersion"`
    FileSize      int64  `json:"fileSize"`
    ReleaseNotes  string `json:"releaseNotes"`
}

type DownloadResult struct {
    Success   bool   `json:"success"`
    File      string `json:"file"`
    Verified  bool   `json:"verified"`
    Decrypted bool   `json:"decrypted"`
}

func CheckUpdate() (*UpdateInfo, error) {
    out, err := exec.Command("./update-client.exe", "check", "--json").Output()
    if err != nil {
        return nil, err
    }

    var info UpdateInfo
    if err := json.Unmarshal(out, &info); err != nil {
        return nil, err
    }

    return &info, nil
}

func DownloadUpdate(version string) (*DownloadResult, error) {
    out, err := exec.Command("./update-client.exe", "download", "--version", version, "--json").Output()
    if err != nil {
        return nil, err
    }

    var result DownloadResult
    if err := json.Unmarshal(out, &result); err != nil {
        return nil, err
    }

    return &result, nil
}
```

### Python 应用

```python
import json
import subprocess
from pathlib import Path

def check_update():
    result = subprocess.run(
        ["./update-client.exe", "check", "--json"],
        capture_output=True,
        text=True
    )

    if result.returncode != 0:
        return None

    return json.loads(result.stdout)

def download_update(version):
    result = subprocess.run(
        ["./update-client.exe", "download", "--version", version, "--json"],
        capture_output=True,
        text=True
    )

    if result.returncode != 0:
        return None

    return json.loads(result.stdout)

# 使用示例
update_info = check_update()
if update_info and update_info.get("hasUpdate"):
    print(f"New version available: {update_info['latestVersion']}")

    download_result = download_update(update_info['latestVersion'])
    if download_result and download_result.get("success"):
        file_path = download_result["file"]
        print(f"Downloaded to: {file_path}")
```

### Java 应用

```java
import com.fasterxml.jackson.databind.ObjectMapper;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.util.ArrayList;
import java.util.List;

public class UpdateService {
    private final String updateClientPath;

    public UpdateService() {
        this.updateClientPath = "./update-client.exe";
    }

    public UpdateInfo checkUpdate() throws Exception {
        List<String> command = new ArrayList<>();
        command.add(updateClientPath);
        command.add("check");
        command.add("--json");

        ProcessBuilder pb = new ProcessBuilder(command);
        pb.redirectErrorStream(true);
        Process process = pb.start();

        BufferedReader reader = new BufferedReader(
            new InputStreamReader(process.getInputStream())
        );

        StringBuilder output = new StringBuilder();
        String line;
        while ((line = reader.readLine()) != null) {
            output.append(line);
        }

        process.waitFor();

        ObjectMapper mapper = new ObjectMapper();
        return mapper.readValue(output.toString(), UpdateInfo.class);
    }

    public DownloadResult downloadUpdate(String version) throws Exception {
        // 类似实现
        // ...
    }
}
```

## 错误处理

### 退出码

- `0`: 成功
- `1`: 错误（网络、文件、验证失败等）
- `2`: 配置错误

### 常见错误

**网络错误：**
```
✗ Failed to connect to server
  Server: http://localhost:8080
  Error: Connection refused
```

**版本不存在：**
```
✗ Download failed
  Version: 1.2.0
  Error: Version not found on server
```

**验证失败：**
```
✗ Verification failed
  Expected: abc123...
  Actual: def456...
```

## 获取配置

配置文件中的关键信息（Token 和密钥）从 Update Server 的 Web 管理界面获取：

1. 登录管理后台
2. 进入程序详情页
3. 点击「下载更新端」
4. 解压下载的 zip 包
5. 将 `update-client.exe` 和 `update-config.yaml` 复制到你的应用目录

下载的配置文件已预先填充：
- 正确的服务器地址
- 程序 ID
- Download Token
- 加密密钥（如启用）

## 部署清单

发布应用时，确保包含以下文件：

- [ ] `update-client.exe` - 更新客户端工具
- [ ] `update-config.yaml` - 配置文件
- [ ] `updates/` 目录（可选，用于存放下载的更新包）

## 最佳实践

1. **后台检查**：在应用启动时或定期在后台检查更新
2. **用户提示**：发现更新时，提示用户并显示更新内容
3. **断点续传**：下载失败时，支持重试（工具已内置重试机制）
4. **版本回滚**：保留旧版本文件，以便需要时回滚
5. **签名验证**：对于生产环境，建议验证下载文件的数字签名

## 故障排查

**问题**：`update-client.exe` 无法运行
- **解决**：确保 .exe 和配置文件在同一目录

**问题**：检查更新时返回网络错误
- **解决**：检查 `server.url` 是否正确，服务器是否可访问

**问题**：下载的文件无法使用
- **解决**：检查配置文件中的 `encryption_key` 是否正确

## 支持

- GitHub Issues: https://github.com/LiteHomeLab/update-server/issues
- 架构文档: [ARCHITECTURE.md](ARCHITECTURE.md)
