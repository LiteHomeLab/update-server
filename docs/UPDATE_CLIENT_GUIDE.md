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

## 集成示例

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
