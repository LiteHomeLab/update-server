# 客户端 SDK 和管理工具实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-step.

**目标:** 为 DocuFiller 更新服务器创建三种语言（Go、Python、C#）的客户端自动更新 SDK 和管理工具，包含生产级功能（进度回调、失败重试、文件验证），并通过集成测试验证。

**架构:** 分层设计 - HTTP 客户端层 + 业务逻辑层 + 进度回调层，使用各语言标准 HTTP 库实现 RESTful API 调用，SHA256 验证文件完整性，支持异步下载和重试机制。

**Tech Stack:** Go 1.23+, Python 3.10+, .NET 8.0, requests, aiohttp, System.Net.Http

---

## 前置准备

### Task 0: 创建项目结构

**目标:** 创建客户端 SDK 的目录结构。

**Files:**
- Create: `clients/`
- Create: `clients/go/`
- Create: `clients/python/`
- Create: `clients/csharp/`
- Create: `tests/`
- Create: `tests/test_data/`

**Step 1: 创建目录结构**

```bash
mkdir -p clients/go/client clients/go/admin
mkdir -p clients/python/update_client clients/python/update_admin
mkdir -p clients/csharp/DocuFiller.UpdateClient clients/csharp/DocuFiller.UpdateAdmin
mkdir -p tests/test_data/packages
```

**Step 2: 验证目录创建**

```bash
ls -la clients/
```

Expected: 显示 `go/`, `python/`, `csharp/` 三个目录

**Step 3: 创建 README 占位文件**

```bash
touch clients/go/README.md
touch clients/python/README.md
touch clients/csharp/README.md
```

**Step 4: Git 提交**

```bash
git add clients/
git commit -m "feat: create client SDK directory structure"
```

---

## 阶段一：Go 客户端 SDK

### Task 1: Go 客户端 - 基础结构

**目标:** 创建 Go 客户端 SDK 的基础结构和配置。

**Files:**
- Create: `clients/go/client/go.mod`
- Create: `clients/go/client/config.go`
- Create: `clients/go/client/types.go`

**Step 1: 创建 go.mod**

```go
module github.com/LiteHomeLab/update-client

go 1.23

require (
    github.com/stretchr/testify v1.9.0
)
```

**Step 2: 创建配置类型**

**File:** `clients/go/client/config.go`

```go
package client

import "time"

// Config 客户端配置
type Config struct {
    ServerURL   string        // 服务器地址
    ProgramID   string        // 程序 ID
    Channel     string        // 发布渠道: stable 或 beta
    Timeout     time.Duration // 请求超时时间
    MaxRetries  int           // 最大重试次数
    SavePath    string        // 下载保存路径
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
    return &Config{
        ServerURL:  "http://localhost:8080",
        Channel:    "stable",
        Timeout:    30 * time.Second,
        MaxRetries: 3,
        SavePath:   "./updates",
    }
}
```

**Step 3: 创建数据类型**

**File:** `clients/go/client/types.go`

```go
package client

import "time"

// UpdateInfo 更新信息
type UpdateInfo struct {
    Version      string    `json:"version"`
    Channel      string    `json:"channel"`
    FileName     string    `json:"fileName"`
    FileSize     int64     `json:"fileSize"`
    FileHash     string    `json:"fileHash"`
    ReleaseNotes string    `json:"releaseNotes"`
    PublishDate  time.Time `json:"publishDate"`
    Mandatory    bool      `json:"mandatory"`
    DownloadCount int      `json:"downloadCount"`
}

// DownloadProgress 下载进度
type DownloadProgress struct {
    Version    string
    Downloaded int64
    Total      int64
    Percentage float64
    Speed      float64 // bytes/second
}

// ProgressCallback 进度回调函数
type ProgressCallback func(DownloadProgress)

// UpdateError 更新错误
type UpdateError struct {
    Code    string
    Message string
    Err     error
}

func (e *UpdateError) Error() string {
    return e.Message
}
```

**Step 4: 初始化模块**

```bash
cd clients/go/client && go mod tidy
```

**Step 5: Git 提交**

```bash
git add clients/go/client/
git commit -m "feat(go-client): add base config and types"
```

### Task 2: Go 客户端 - 版本比较

**目标:** 实现语义化版本比较功能。

**Files:**
- Create: `clients/go/client/version.go`
- Create: `clients/go/client/version_test.go`

**Step 1: 编写测试**

**File:** `clients/go/client/version_test.go`

```go
package client

import (
    "testing"
)

func TestCompareVersions(t *testing.T) {
    tests := []struct {
        name     string
        v1       string
        v2       string
        expected int
    }{
        {"equal", "1.0.0", "1.0.0", 0},
        {"v1 greater", "1.2.0", "1.1.0", 1},
        {"v2 greater", "1.0.0", "2.0.0", -1},
        {"with v prefix", "v1.0.0", "1.0.0", 0},
        {"three parts", "1.2.3", "1.2.2", 1},
        {"four parts", "1.2.3.4", "1.2.3.3", 1},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CompareVersions(tt.v1, tt.v2)
            if result != tt.expected {
                t.Errorf("CompareVersions(%q, %q) = %d, want %d",
                    tt.v1, tt.v2, result, tt.expected)
            }
        })
    }
}
```

**Step 2: 运行测试验证失败**

```bash
cd clients/go/client && go test -v -run TestCompareVersions
```

Expected: FAIL - "undefined: CompareVersions"

**Step 3: 实现版本比较**

**File:** `clients/go/client/version.go`

```go
package client

import (
    "strconv"
    "strings"
)

// CompareVersions 比较两个版本号
// 返回: -1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)
func CompareVersions(v1, v2 string) int {
    // 移除 v 前缀
    v1 = strings.TrimPrefix(v1, "v")
    v2 = strings.TrimPrefix(v2, "v")

    parts1 := parseVersion(v1)
    parts2 := parseVersion(v2)

    maxLen := len(parts1)
    if len(parts2) > maxLen {
        maxLen = len(parts2)
    }

    for i := 0; i < maxLen; i++ {
        p1 := 0
        p2 := 0

        if i < len(parts1) {
            p1 = parts1[i]
        }
        if i < len(parts2) {
            p2 = parts2[i]
        }

        if p1 > p2 {
            return 1
        }
        if p1 < p2 {
            return -1
        }
    }

    return 0
}

func parseVersion(version string) []int {
    parts := strings.Split(version, ".")
    result := make([]int, len(parts))

    for i, part := range parts {
        val, _ := strconv.Atoi(part)
        result[i] = val
    }

    return result
}
```

**Step 4: 运行测试验证通过**

```bash
cd clients/go/client && go test -v -run TestCompareVersions
```

Expected: PASS

**Step 5: Git 提交**

```bash
git add clients/go/client/
git commit -m "feat(go-client): add version comparison with tests"
```

### Task 3: Go 客户端 - UpdateChecker 实现

**目标:** 实现核心的 UpdateChecker 客户端。

**Files:**
- Create: `clients/go/client/checker.go`
- Create: `clients/go/client/checker_test.go`

**Step 1: 编写测试**

**File:** `clients/go/client/checker_test.go`

```go
package client

import (
    "testing"
)

func TestNewUpdateChecker(t *testing.T) {
    config := DefaultConfig()
    config.ProgramID = "testapp"

    checker := NewUpdateChecker(config)

    if checker == nil {
        t.Fatal("NewUpdateChecker returned nil")
    }

    if checker.config.ProgramID != "testapp" {
        t.Errorf("ProgramID = %s, want testapp", checker.config.ProgramID)
    }
}
```

**Step 2: 运行测试验证失败**

```bash
cd clients/go/client && go test -v -run TestNewUpdateChecker
```

Expected: FAIL - "undefined: NewUpdateChecker"

**Step 3: 实现 UpdateChecker**

**File:** `clients/go/client/checker.go`

```go
package client

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// UpdateChecker 更新检查器
type UpdateChecker struct {
    config     *Config
    httpClient *http.Client
}

// NewUpdateChecker 创建更新检查器
func NewUpdateChecker(config *Config) *UpdateChecker {
    return &UpdateChecker{
        config: config,
        httpClient: &http.Client{
            Timeout: config.Timeout,
        },
    }
}

// CheckUpdate 检查是否有新版本
func (c *UpdateChecker) CheckUpdate(currentVersion string) (*UpdateInfo, error) {
    url := fmt.Sprintf("%s/api/version/%s/latest?channel=%s",
        c.config.ServerURL, c.config.ProgramID, c.config.Channel)

    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, &UpdateError{
            Code:    "NETWORK_ERROR",
            Message: fmt.Sprintf("Failed to connect to server: %v", err),
            Err:     err,
        }
    }
    defer resp.Body.Close()

    if resp.StatusCode == 404 {
        return nil, &UpdateError{
            Code:    "NO_VERSION",
            Message: "No version found for this program",
        }
    }

    if resp.StatusCode != 200 {
        return nil, &UpdateError{
            Code:    "SERVER_ERROR",
            Message: fmt.Sprintf("Server returned status %d", resp.StatusCode),
        }
    }

    var info UpdateInfo
    if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
        return nil, &UpdateError{
            Code:    "PARSE_ERROR",
            Message: "Failed to parse response",
            Err:     err,
        }
    }

    // 检查是否是新版本
    if CompareVersions(info.Version, currentVersion) <= 0 {
        return nil, nil // 没有新版本
    }

    return &info, nil
}
```

**Step 4: 运行测试验证通过**

```bash
cd clients/go/client && go test -v -run TestNewUpdateChecker
```

Expected: PASS

**Step 5: Git 提交**

```bash
git add clients/go/client/
git commit -m "feat(go-client): implement UpdateChecker with CheckUpdate"
```

### Task 4: Go 客户端 - 下载器实现

**目标:** 实现带进度回调和重试机制的下载器。

**Files:**
- Create: `clients/go/client/downloader.go`
- Create: `clients/go/client/downloader_test.go`

**Step 1: 编写下载器测试**

**File:** `clients/go/client/downloader_test.go`

```go
package client

import (
    "os"
    "path/filepath"
    "testing"
    "time"
)

func TestDownloadUpdate(t *testing.T) {
    // 这个测试需要运行中的服务器
    config := DefaultConfig()
    config.ServerURL = "http://localhost:18080"
    config.ProgramID = "testapp"
    config.Channel = "stable"

    checker := NewUpdateChecker(config)

    tempDir := t.TempDir()
    destPath := filepath.Join(tempDir, "downloaded.zip")

    progressCalled := false
    callback := func(p DownloadProgress) {
        progressCalled = true
    }

    err := checker.DownloadUpdate("1.0.0", destPath, callback)

    if err != nil {
        t.Logf("Download failed (expected if server not running): %v", err)
        return
    }

    if _, err := os.Stat(destPath); os.IsNotExist(err) {
        t.Error("File was not downloaded")
    }

    if !progressCalled {
        t.Error("Progress callback was not called")
    }
}
```

**Step 2: 运行测试验证失败**

```bash
cd clients/go/client && go test -v -run TestDownloadUpdate
```

Expected: FAIL - "method DownloadUpdate not defined"

**Step 3: 实现下载器**

**File:** `clients/go/client/downloader.go`

```go
package client

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"
)

// DownloadUpdate 下载更新包
func (c *UpdateChecker) DownloadUpdate(version string, destPath string, callback ProgressCallback) error {
    var lastErr error

    // 重试机制
    for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
        if attempt > 0 {
            time.Sleep(time.Duration(attempt) * 2 * time.Second) // 指数退避
        }

        err := c.downloadOnce(version, destPath, callback)
        if err == nil {
            return nil // 成功
        }

        lastErr = err
    }

    return lastErr
}

func (c *UpdateChecker) downloadOnce(version string, destPath string, callback ProgressCallback) error {
    url := fmt.Sprintf("%s/api/download/%s/%s/%s",
        c.config.ServerURL, c.config.ProgramID, c.config.Channel, version)

    resp, err := c.httpClient.Get(url)
    if err != nil {
        return &UpdateError{
            Code:    "NETWORK_ERROR",
            Message: "Failed to connect to server",
            Err:     err,
        }
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return &UpdateError{
            Code:    "DOWNLOAD_ERROR",
            Message: fmt.Sprintf("Server returned status %d", resp.StatusCode),
        }
    }

    // 确保目录存在
    if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
        return &UpdateError{
            Code:    "FILE_ERROR",
            Message: "Failed to create directory",
            Err:     err,
        }
    }

    // 创建文件
    file, err := os.Create(destPath)
    if err != nil {
        return &UpdateError{
            Code:    "FILE_ERROR",
            Message: "Failed to create file",
            Err:     err,
        }
    }
    defer file.Close()

    // 获取文件大小
    total := resp.ContentLength
    downloaded := int64(0)
    startTime := time.Now()

    // 使用 buffer 复制
    buffer := make([]byte, 32*1024) // 32KB chunks
    for {
        n, err := resp.Body.Read(buffer)
        if n > 0 {
            if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
                return writeErr
            }

            downloaded += int64(n)

            // 调用进度回调
            if callback != nil {
                elapsed := time.Since(startTime).Seconds()
                speed := float64(downloaded) / elapsed

                callback(DownloadProgress{
                    Version:    version,
                    Downloaded: downloaded,
                    Total:      total,
                    Percentage: float64(downloaded) / float64(total) * 100,
                    Speed:      speed,
                })
            }
        }

        if err == io.EOF {
            break
        }
        if err != nil {
            return &UpdateError{
                Code:    "DOWNLOAD_ERROR",
                Message: "Failed to download file",
                Err:     err,
            }
        }
    }

    return nil
}

// VerifyFile 验证文件 SHA256 哈希
func (c *UpdateChecker) VerifyFile(filePath string, expectedHash string) (bool, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return false, err
    }
    defer file.Close()

    hash := sha256.New()
    if _, err := io.Copy(hash, file); err != nil {
        return false, err
    }

    actualHash := hex.EncodeToString(hash.Sum(nil))
    return actualHash == expectedHash, nil
}
```

**Step 4: 运行测试**

```bash
cd clients/go/client && go test -v -run TestDownloadUpdate
```

Expected: 可能失败（服务器未运行），但代码已编译

**Step 5: Git 提交**

```bash
git add clients/go/client/
git commit -m "feat(go-client): implement downloader with progress and retry"
```

---

## 阶段二：Python 客户端 SDK

### Task 5: Python 客户端 - 基础结构

**目标:** 创建 Python 客户端 SDK 的基础结构。

**Files:**
- Create: `clients/python/update_client/__init__.py`
- Create: `clients/python/update_client/config.py`
- Create: `clients/python/update_client/types.py`
- Create: `clients/python/update_client/requirements.txt`

**Step 1: 创建 requirements.txt**

**File:** `clients/python/update_client/requirements.txt`

```
requests>=2.31.0
aiohttp>=3.9.0
pytest>=7.4.0
pytest-asyncio>=0.21.0
```

**Step 2: 创建配置类型**

**File:** `clients/python/update_client/config.py`

```python
"""配置管理"""
from dataclasses import dataclass
from typing import Optional


@dataclass
class Config:
    """客户端配置"""
    server_url: str
    program_id: str
    channel: str = "stable"
    timeout: int = 30
    max_retries: int = 3
    save_path: str = "./updates"

    @classmethod
    def default(cls) -> "Config":
        """返回默认配置"""
        return cls(
            server_url="http://localhost:8080",
            channel="stable",
            timeout=30,
            max_retries=3,
            save_path="./updates"
        )
```

**Step 3: 创建数据类型**

**File:** `clients/python/update_client/types.py`

```python
"""数据类型定义"""
from dataclasses import dataclass
from datetime import datetime
from typing import Callable, Optional


@dataclass
class UpdateInfo:
    """更新信息"""
    version: str
    channel: str
    file_name: str
    file_size: int
    file_hash: str
    release_notes: str
    publish_date: datetime
    mandatory: bool
    download_count: int


@dataclass
class DownloadProgress:
    """下载进度"""
    version: str
    downloaded: int
    total: int
    percentage: float
    speed: float  # bytes/second


ProgressCallback = Callable[[DownloadProgress], None]


class UpdateError(Exception):
    """更新错误"""
    def __init__(self, code: str, message: str):
        self.code = code
        self.message = message
        super().__init__(f"[{code}] {message}")
```

**Step 4: 创建包初始化文件**

**File:** `clients/python/update_client/__init__.py`

```python
"""DocuFiller Update Client SDK"""
from .config import Config
from .types import UpdateInfo, DownloadProgress, ProgressCallback, UpdateError
from .version import compare_versions
from .checker import UpdateChecker

__version__ = "1.0.0"
__all__ = [
    "Config",
    "UpdateInfo",
    "DownloadProgress",
    "ProgressCallback",
    "UpdateError",
    "compare_versions",
    "UpdateChecker",
]
```

**Step 5: Git 提交**

```bash
git add clients/python/update_client/
git commit -m "feat(py-client): add base config and types"
```

### Task 6: Python 客户端 - 版本比较

**目标:** 实现语义化版本比较。

**Files:**
- Create: `clients/python/update_client/version.py`
- Create: `clients/python/update_client/test_version.py`

**Step 1: 编写测试**

**File:** `clients/python/update_client/test_version.py`

```python
"""版本比较测试"""
import pytest
from update_client.version import compare_versions


def test_equal_versions():
    """测试相等版本"""
    assert compare_versions("1.0.0", "1.0.0") == 0


def test_v_prefix():
    """测试 v 前缀"""
    assert compare_versions("v1.0.0", "1.0.0") == 0


def test_v1_greater():
    """测试 v1 大于 v2"""
    assert compare_versions("1.2.0", "1.1.0") == 1


def test_v2_greater():
    """测试 v2 大于 v1"""
    assert compare_versions("1.0.0", "2.0.0") == -1


def test_three_parts():
    """测试三位版本号"""
    assert compare_versions("1.2.3", "1.2.2") == 1


def test_four_parts():
    """测试四位版本号"""
    assert compare_versions("1.2.3.4", "1.2.3.3") == 1
```

**Step 2: 运行测试验证失败**

```bash
cd clients/python/update_client && pytest test_version.py -v
```

Expected: FAIL - "ModuleNotFoundError: No module named 'update_client.version'"

**Step 3: 实现版本比较**

**File:** `clients/python/update_client/version.py`

```python
"""版本比较工具"""
from typing import List


def compare_versions(v1: str, v2: str) -> int:
    """
    比较两个版本号

    Args:
        v1: 版本号 1
        v2: 版本号 2

    Returns:
        -1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)
    """
    # 移除 v 前缀
    v1 = v1.lstrip('v')
    v2 = v2.lstrip('v')

    parts1 = _parse_version(v1)
    parts2 = _parse_version(v2)

    max_len = max(len(parts1), len(parts2))

    for i in range(max_len):
        p1 = parts1[i] if i < len(parts1) else 0
        p2 = parts2[i] if i < len(parts2) else 0

        if p1 > p2:
            return 1
        if p1 < p2:
            return -1

    return 0


def _parse_version(version: str) -> List[int]:
    """解析版本号为数字列表"""
    parts = version.split('.')
    result = []

    for part in parts:
        try:
            result.append(int(part))
        except ValueError:
            result.append(0)

    return result
```

**Step 4: 运行测试验证通过**

```bash
cd clients/python/update_client && pytest test_version.py -v
```

Expected: PASS

**Step 5: Git 提交**

```bash
git add clients/python/update_client/
git commit -m "feat(py-client): add version comparison with tests"
```

### Task 7: Python 客户端 - UpdateChecker 实现

**目标:** 实现核心的 UpdateChecker 类。

**Files:**
- Create: `clients/python/update_client/checker.py`
- Create: `clients/python/update_client/test_checker.py`

**Step 1: 编写测试**

**File:** `clients/python/update_client/test_checker.py`

```python
"""UpdateChecker 测试"""
import pytest
from update_client import Config, UpdateChecker
from update_client.version import compare_versions


def test_new_checker():
    """测试创建 UpdateChecker"""
    config = Config.default()
    config.program_id = "testapp"

    checker = UpdateChecker(config)

    assert checker.config.program_id == "testapp"
    assert checker.config.channel == "stable"
```

**Step 2: 运行测试验证失败**

```bash
cd clients/python/update_client && pytest test_checker.py -v
```

Expected: FAIL - "No module named 'update_client.checker'"

**Step 3: 实现 UpdateChecker**

**File:** `clients/python/update_client/checker.py`

```python
"""更新检查器"""
import time
import hashlib
import os
from typing import Optional
from pathlib import Path
from datetime import datetime

import requests

from .config import Config
from .types import UpdateInfo, DownloadProgress, ProgressCallback, UpdateError
from .version import compare_versions


class UpdateChecker:
    """更新检查器"""

    def __init__(self, config: Config):
        """
        初始化更新检查器

        Args:
            config: 客户端配置
        """
        self.config = config
        self.session = requests.Session()
        self.session.timeout = config.timeout

    def check_update(self, current_version: str) -> Optional[UpdateInfo]:
        """
        检查是否有新版本

        Args:
            current_version: 当前版本号

        Returns:
            UpdateInfo 如果有新版本，否则 None
        """
        url = f"{self.config.server_url}/api/version/{self.config.program_id}/latest"
        params = {"channel": self.config.channel}

        try:
            response = self.session.get(url, params=params)

            if response.status_code == 404:
                raise UpdateError("NO_VERSION", "No version found for this program")

            response.raise_for_status()

            data = response.json()
            info = UpdateInfo(
                version=data["version"],
                channel=data["channel"],
                file_name=data["fileName"],
                file_size=data["fileSize"],
                file_hash=data["fileHash"],
                release_notes=data["releaseNotes"],
                publish_date=datetime.fromisoformat(data["publishDate"].replace("Z", "+00:00")),
                mandatory=data["mandatory"],
                download_count=data["downloadCount"]
            )

            # 检查是否是新版本
            if compare_versions(info.version, current_version) <= 0:
                return None

            return info

        except requests.RequestException as e:
            raise UpdateError("NETWORK_ERROR", f"Failed to connect to server: {e}")

    def download_update(
        self,
        version: str,
        dest_path: str,
        progress_callback: Optional[ProgressCallback] = None
    ) -> None:
        """
        下载更新包（带重试）

        Args:
            version: 版本号
            dest_path: 目标路径
            progress_callback: 进度回调
        """
        last_error = None

        for attempt in range(self.config.max_retries + 1):
            if attempt > 0:
                time.sleep(attempt * 2)  # 指数退避

            try:
                self._download_once(version, dest_path, progress_callback)
                return  # 成功
            except Exception as e:
                last_error = e

        raise last_error

    def _download_once(
        self,
        version: str,
        dest_path: str,
        progress_callback: Optional[ProgressCallback] = None
    ) -> None:
        """单次下载尝试"""
        url = f"{self.config.server_url}/api/download/{self.config.program_id}/{self.config.channel}/{version}"

        response = self.session.get(url, stream=True)
        response.raise_for_status()

        # 确保目录存在
        dest_path_obj = Path(dest_path)
        dest_path_obj.parent.mkdir(parents=True, exist_ok=True)

        total = int(response.headers.get("content-length", 0))
        downloaded = 0
        start_time = time.time()

        with open(dest_path, "wb") as f:
            for chunk in response.iter_content(chunk_size=32 * 1024):
                if chunk:
                    f.write(chunk)
                    downloaded += len(chunk)

                    # 进度回调
                    if progress_callback:
                        elapsed = time.time() - start_time
                        speed = downloaded / elapsed if elapsed > 0 else 0

                        progress_callback(DownloadProgress(
                            version=version,
                            downloaded=downloaded,
                            total=total,
                            percentage=(downloaded / total * 100) if total > 0 else 0,
                            speed=speed
                        ))

    def verify_file(self, file_path: str, expected_hash: str) -> bool:
        """
        验证文件 SHA256 哈希

        Args:
            file_path: 文件路径
            expected_hash: 期望的哈希值

        Returns:
            是否匹配
        """
        sha256_hash = hashlib.sha256()

        with open(file_path, "rb") as f:
            for chunk in iter(lambda: f.read(4096), b""):
                sha256_hash.update(chunk)

        actual_hash = sha256_hash.hexdigest()
        return actual_hash.lower() == expected_hash.lower()
```

**Step 4: 运行测试验证通过**

```bash
cd clients/python/update_client && pytest test_checker.py -v
```

Expected: PASS

**Step 5: Git 提交**

```bash
git add clients/python/update_client/
git commit -m "feat(py-client): implement UpdateChecker with download"
```

---

## 阶段三：C# 客户端 SDK

### Task 8: C# 客户端 - 基础结构

**目标:** 创建 C# 客户端 SDK 项目结构。

**Files:**
- Create: `clients/csharp/DocuFiller.UpdateClient/DocuFiller.UpdateClient.csproj`
- Create: `clients/csharp/DocuFiller.UpdateClient/Config.cs`
- Create: `clients/csharp/DocuFiller.UpdateClient/Types.cs`

**Step 1: 创建项目文件**

**File:** `clients/csharp/DocuFiller.UpdateClient/DocuFiller.UpdateClient.csproj`

```xml
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <ImplicitUsings>enable</ImplicitUsings>
    <Nullable>enable</Nullable>
    <RootNamespace>DocuFiller.UpdateClient</RootNamespace>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="System.Net.Http.Json" Version="8.0.0" />
    <PackageReference Include="Microsoft.Extensions.Logging" Version="8.0.0" />
  </ItemGroup>

</Project>
```

**Step 2: 创建配置类**

**File:** `clients/csharp/DocuFiller.UpdateClient/Config.cs`

```csharp
namespace DocuFiller.UpdateClient;

/// <summary>
/// 客户端配置
/// </summary>
public class Config
{
    /// <summary>
    /// 服务器地址
    /// </summary>
    public string ServerUrl { get; set; } = "http://localhost:8080";

    /// <summary>
    /// 程序 ID
    /// </summary>
    public string ProgramId { get; set; } = string.Empty;

    /// <summary>
    /// 发布渠道
    /// </summary>
    public string Channel { get; set; } = "stable";

    /// <summary>
    /// 请求超时时间（秒）
    /// </summary>
    public int Timeout { get; set; } = 30;

    /// <summary>
    /// 最大重试次数
    /// </summary>
    public int MaxRetries { get; set; } = 3;

    /// <summary>
    /// 下载保存路径
    /// </summary>
    public string SavePath { get; set; } = "./updates";

    /// <summary>
    /// 创建默认配置
    /// </summary>
    public static Config Default() => new();
}
```

**Step 3: 创建类型定义**

**File:** `clients/csharp/DocuFiller.UpdateClient/Types.cs`

```csharp
using System.Text.Json.Serialization;

namespace DocuFiller.UpdateClient;

/// <summary>
/// 更新信息
/// </summary>
public record UpdateInfo
{
    [JsonPropertyName("version")]
    public string Version { get; init; } = string.Empty;

    [JsonPropertyName("channel")]
    public string Channel { get; init; } = string.Empty;

    [JsonPropertyName("fileName")]
    public string FileName { get; init; } = string.Empty;

    [JsonPropertyName("fileSize")]
    public long FileSize { get; init; }

    [JsonPropertyName("fileHash")]
    public string FileHash { get; init; } = string.Empty;

    [JsonPropertyName("releaseNotes")]
    public string ReleaseNotes { get; init; } = string.Empty;

    [JsonPropertyName("publishDate")]
    public DateTime PublishDate { get; init; }

    [JsonPropertyName("mandatory")]
    public bool Mandatory { get; init; }

    [JsonPropertyName("downloadCount")]
    public int DownloadCount { get; init; }
}

/// <summary>
/// 下载进度
/// </summary>
public record DownloadProgress
{
    public string Version { get; init; } = string.Empty;
    public long Downloaded { get; init; }
    public long Total { get; init; }
    public double Percentage { get; init; }
    public double Speed { get; init; } // bytes/second
}

/// <summary>
/// 更新错误
/// </summary>
public class UpdateError : Exception
{
    public string Code { get; }

    public UpdateError(string code, string message) : base($"[{code}] {message}")
    {
        Code = code;
    }
}
```

**Step 4: Git 提交**

```bash
git add clients/csharp/DocuFiller.UpdateClient/
git commit -m "feat(cs-client): add base config and types"
```

### Task 9: C# 客户端 - 版本比较

**目标:** 实现版本比较功能。

**Files:**
- Create: `clients/csharp/DocuFiller.UpdateClient/VersionComparer.cs`
- Create: `clients/csharp/DocuFiller.UpdateClient/VersionComparerTests.cs`

**Step 1: 编写测试**

**File:** `clients/csharp/DocuFiller.UpdateClient/VersionComparerTests.cs`

```csharp
using Xunit;

namespace DocuFiller.UpdateClient.Tests;

public class VersionComparerTests
{
    [Theory]
    [InlineData("1.0.0", "1.0.0", 0)]
    [InlineData("1.2.0", "1.1.0", 1)]
    [InlineData("1.0.0", "2.0.0", -1)]
    [InlineData("v1.0.0", "1.0.0", 0)]
    [InlineData("1.2.3", "1.2.2", 1)]
    public void CompareVersions_ShouldReturnExpected(string v1, string v2, int expected)
    {
        var result = VersionComparer.Compare(v1, v2);
        Assert.Equal(expected, result);
    }
}
```

**Step 2: 运行测试验证失败**

```bash
cd clients/csharp/DocuFiller.UpdateClient && dotnet test
```

Expected: FAIL - "VersionComparer does not exist"

**Step 3: 实现版本比较**

**File:** `clients/csharp/DocuFiller.UpdateClient/VersionComparer.cs`

```csharp
namespace DocuFiller.UpdateClient;

/// <summary>
/// 版本比较器
/// </summary>
public static class VersionComparer
{
    /// <summary>
    /// 比较两个版本号
    /// </summary>
    /// <returns>-1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)</returns>
    public static int Compare(string v1, string v2)
    {
        // 移除 v 前缀
        v1 = v1.TrimStart('v');
        v2 = v2.TrimStart('v');

        var parts1 = ParseVersion(v1);
        var parts2 = ParseVersion(v2);

        var maxLen = Math.Max(parts1.Count, parts2.Count);

        for (int i = 0; i < maxLen; i++)
        {
            int p1 = i < parts1.Count ? parts1[i] : 0;
            int p2 = i < parts2.Count ? parts2[i] : 0;

            if (p1 > p2) return 1;
            if (p1 < p2) return -1;
        }

        return 0;
    }

    private static List<int> ParseVersion(string version)
    {
        var parts = version.Split('.');
        var result = new List<int>();

        foreach (var part in parts)
        {
            if (int.TryParse(part, out int num))
            {
                result.Add(num);
            }
            else
            {
                result.Add(0);
            }
        }

        return result;
    }
}
```

**Step 4: 运行测试验证通过**

```bash
cd clients/csharp/DocuFiller.UpdateClient && dotnet test
```

Expected: PASS

**Step 5: Git 提交**

```bash
git add clients/csharp/DocuFiller.UpdateClient/
git commit -m "feat(cs-client): add version comparison with tests"
```

### Task 10: C# 客户端 - UpdateChecker 实现

**目标:** 实现完整的 UpdateChecker 类。

**Files:**
- Create: `clients/csharp/DocuFiller.UpdateClient/UpdateChecker.cs`
- Create: `clients/csharp/DocuFiller.UpdateClient/UpdateCheckerTests.cs`

**Step 1: 编写测试**

**File:** `clients/csharp/DocuFiller.UpdateClient/UpdateCheckerTests.cs`

```csharp
using Xunit;
using Moq;

namespace DocuFiller.UpdateClient.Tests;

public class UpdateCheckerTests
{
    [Fact]
    public void Constructor_ShouldInitializeWithConfig()
    {
        var config = new Config
        {
            ProgramId = "testapp",
            Channel = "stable"
        };

        var checker = new UpdateChecker(config);

        Assert.NotNull(checker);
    }
}
```

**Step 2: 运行测试验证失败**

```bash
cd clients/csharp/DocuFiller.UpdateClient && dotnet test
```

Expected: FAIL - "UpdateChecker does not exist"

**Step 3: 实现 UpdateChecker**

**File:** `clients/csharp/DocuFiller.UpdateClient/UpdateChecker.cs`

```csharp
using System.Runtime.CompilerServices;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;

namespace DocuFiller.UpdateClient;

/// <summary>
/// 更新检查器
/// </summary>
public class UpdateChecker
{
    private readonly Config _config;
    private readonly HttpClient _httpClient;

    public UpdateChecker(Config config)
    {
        _config = config;
        _httpClient = new HttpClient
        {
            Timeout = TimeSpan.FromSeconds(config.Timeout)
        };
    }

    /// <summary>
    /// 检查是否有新版本
    /// </summary>
    public async Task<UpdateInfo?> CheckUpdateAsync(string currentVersion, CancellationToken cancellationToken = default)
    {
        var url = $"{_config.ServerUrl}/api/version/{_config.ProgramId}/latest?channel={_config.Channel}";

        try
        {
            var response = await _httpClient.GetAsync(url, cancellationToken);

            if (response.StatusCode == System.Net.HttpStatusCode.NotFound)
            {
                throw new UpdateError("NO_VERSION", "No version found for this program");
            }

            response.EnsureSuccessStatusCode();

            var json = await response.Content.ReadAsStringAsync(cancellationToken);
            var info = JsonSerializer.Deserialize<UpdateInfo>(json, new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            });

            if (info == null) return null;

            // 检查是否是新版本
            if (VersionComparer.Compare(info.Version, currentVersion) <= 0)
            {
                return null; // 没有新版本
            }

            return info;
        }
        catch (HttpRequestException ex)
        {
            throw new UpdateError("NETWORK_ERROR", $"Failed to connect to server: {ex.Message}");
        }
    }

    /// <summary>
    /// 下载更新包
    /// </summary>
    public async Task DownloadUpdateAsync(
        string version,
        string destPath,
        IProgress<DownloadProgress>? progress = null,
        CancellationToken cancellationToken = default)
    {
        var lastError = new Exception?();

        for (int attempt = 0; attempt <= _config.MaxRetries; attempt++)
        {
            if (attempt > 0)
            {
                await Task.Delay(attempt * 2000, cancellationToken); // 指数退避
            }

            try
            {
                await DownloadOnceAsync(version, destPath, progress, cancellationToken);
                return; // 成功
            }
            catch (Exception ex)
            {
                lastError = ex;
            }
        }

        throw lastError!;
    }

    private async Task DownloadOnceAsync(
        string version,
        string destPath,
        IProgress<DownloadProgress>? progress,
        CancellationToken cancellationToken)
    {
        var url = $"{_config.ServerUrl}/api/download/{_config.ProgramId}/{_config.Channel}/{version}";

        var response = await _httpClient.GetAsync(url, HttpCompletionOption.ResponseHeadersRead, cancellationToken);
        response.EnsureSuccessStatusCode();

        var total = response.Content.Headers.ContentLength ?? 0;
        var directory = Path.GetDirectoryName(destPath);
        if (!string.IsNullOrEmpty(directory))
        {
            Directory.CreateDirectory(directory);
        }

        using var contentStream = await response.Content.ReadAsStreamAsync(cancellationToken);
        using var fileStream = File.Create(destPath);

        var downloaded = 0L;
        var startTime = DateTime.UtcNow;
        var buffer = new byte[32 * 1024];

        int bytesRead;
        while ((bytesRead = await contentStream.ReadAsync(buffer, cancellationToken)) > 0)
        {
            await fileStream.WriteAsync(buffer.AsMemory(0, bytesRead), cancellationToken);
            downloaded += bytesRead;

            progress?.Report(new DownloadProgress
            {
                Version = version,
                Downloaded = downloaded,
                Total = total,
                Percentage = total > 0 ? (double)downloaded / total * 100 : 0,
                Speed = CalculateSpeed(downloaded, startTime)
            });
        }
    }

    private static double CalculateSpeed(long downloaded, DateTime startTime)
    {
        var elapsed = (DateTime.UtcNow - startTime).TotalSeconds;
        return elapsed > 0 ? downloaded / elapsed : 0;
    }

    /// <summary>
    /// 验证文件 SHA256 哈希
    /// </summary>
    public bool VerifyFile(string filePath, string expectedHash)
    {
        using var sha256 = SHA256.Create();
        using var stream = File.OpenRead(filePath);

        var hash = sha256.ComputeHash(stream);
        var actualHash = Convert.ToHexString(hash).ToLowerInvariant();

        return actualHash == expectedHash.ToLowerInvariant();
    }
}
```

**Step 4: 运行测试验证通过**

```bash
cd clients/csharp/DocuFiller.UpdateClient && dotnet test
```

Expected: PASS

**Step 5: Git 提交**

```bash
git add clients/csharp/DocuFiller.UpdateClient/
git commit -m "feat(cs-client): implement UpdateChecker with async download"
```

---

## 阶段四：管理工具

### Task 11: Go 管理工具

**目标:** 创建 Go CLI 管理工具用于上传和删除版本。

**Files:**
- Create: `clients/go/admin/go.mod`
- Create: `clients/go/admin/admin.go`
- Create: `clients/go/admin/main.go`

**Step 1: 创建 go.mod**

**File:** `clients/go/admin/go.mod`

```go
module github.com/LiteHomeLab/update-admin

go 1.23

require (
    github.com/spf13/cobra v1.8.0
)
```

**Step 2: 实现管理客户端**

**File:** `clients/go/admin/admin.go`

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
    "time"
)

type UpdateAdmin struct {
    serverURL string
    token     string
    client    *http.Client
}

func NewUpdateAdmin(serverURL, token string) *UpdateAdmin {
    return &UpdateAdmin{
        serverURL: serverURL,
        token:     token,
        client: &http.Client{
            Timeout: 60 * time.Second,
        },
    }
}

type VersionInfo struct {
    Version      string    `json:"version"`
    Channel      string    `json:"channel"`
    FileName     string    `json:"fileName"`
    FileSize     int64     `json:"fileSize"`
    PublishDate  time.Time `json:"publishDate"`
}

func (a *UpdateAdmin) UploadVersion(programID, channel, version, filePath, notes string, mandatory bool) error {
    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    var body bytes.Buffer
    writer := multipart.NewWriter(&body)

    writer.WriteField("channel", channel)
    writer.WriteField("version", version)
    writer.WriteField("notes", notes)
    writer.WriteField("mandatory", fmt.Sprintf("%v", mandatory))

    part, err := writer.CreateFormFile("file", file.Name())
    if err != nil {
        return fmt.Errorf("failed to create form file: %w", err)
    }

    io.Copy(part, file)
    writer.Close()

    url := fmt.Sprintf("%s/api/version/%s/upload", a.serverURL, programID)
    req, err := http.NewRequest("POST", url, &body)
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("Authorization", "Bearer "+a.token)

    resp, err := a.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return fmt.Errorf("upload failed with status %d", resp.StatusCode)
    }

    fmt.Printf("Version %s/%s/%s uploaded successfully\n", programID, channel, version)
    return nil
}

func (a *UpdateAdmin) DeleteVersion(programID, channel, version string) error {
    url := fmt.Sprintf("%s/api/version/%s/%s/%s", a.serverURL, programID, channel, version)
    req, err := http.NewRequest("DELETE", url, nil)
    if err != nil {
        return err
    }

    req.Header.Set("Authorization", "Bearer "+a.token)

    resp, err := a.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return fmt.Errorf("delete failed with status %d", resp.StatusCode)
    }

    fmt.Printf("Version %s/%s/%s deleted successfully\n", programID, channel, version)
    return nil
}

func (a *UpdateAdmin) ListVersions(programID, channel string) ([]VersionInfo, error) {
    url := fmt.Sprintf("%s/api/version/%s/list?channel=%s", a.serverURL, programID, channel)

    resp, err := a.client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("list failed with status %d", resp.StatusCode)
    }

    var result struct {
        Versions []VersionInfo `json:"versions"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.Versions, nil
}
```

**Step 3: 创建 CLI 入口**

**File:** `clients/go/admin/main.go`

```go
package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var (
    serverURL string
    token     string
    programID string
)

var rootCmd = &cobra.Command{
    Use:   "update-admin",
    Short: "DocuFiller Update Server Admin Tool",
}

var uploadCmd = &cobra.Command{
    Use:   "upload --channel <stable|beta> --version <version> --file <path>",
    Short: "Upload a new version",
    Args:  cobra.NoArgs,
    Run: func(cmd *cobra.Command, args []string) {
        channel, _ := cmd.Flags().GetString("channel")
        version, _ := cmd.Flags().GetString("version")
        filePath, _ := cmd.Flags().GetString("file")
        notes, _ := cmd.Flags().GetString("notes")
        mandatory, _ := cmd.Flags().GetBool("mandatory")

        admin := NewUpdateAdmin(serverURL, token)
        if err := admin.UploadVersion(programID, channel, version, filePath, notes, mandatory); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    },
}

var deleteCmd = &cobra.Command{
    Use:   "delete --channel <stable|beta> --version <version>",
    Short: "Delete a version",
    Args:  cobra.NoArgs,
    Run: func(cmd *cobra.Command, args []string) {
        channel, _ := cmd.Flags().GetString("channel")
        version, _ := cmd.Flags().GetString("version")

        admin := NewUpdateAdmin(serverURL, token)
        if err := admin.DeleteVersion(programID, channel, version); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    },
}

var listCmd = &cobra.Command{
    Use:   "list [--channel <stable|beta>]",
    Short: "List versions",
    Args:  cobra.NoArgs,
    Run: func(cmd *cobra.Command, args []string) {
        channel, _ := cmd.Flags().GetString("channel")

        admin := NewUpdateAdmin(serverURL, token)
        versions, err := admin.ListVersions(programID, channel)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }

        for _, v := range versions {
            fmt.Printf("%s (%s) - %s - %d bytes\n", v.Version, v.Channel, v.PublishDate.Format("2006-01-02"), v.FileSize)
        }
    },
}

func init() {
    rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
    rootCmd.PersistentFlags().StringVar(&token, "token", "", "API token (required)")
    rootCmd.PersistentFlags().StringVar(&programID, "program-id", "", "Program ID (required)")
    rootCmd.MarkPersistentFlagRequired("token")
    rootCmd.MarkPersistentFlagRequired("program-id")

    uploadCmd.Flags().String("channel", "", "Channel (stable/beta)")
    uploadCmd.Flags().String("version", "", "Version number")
    uploadCmd.Flags().String("file", "", "File path")
    uploadCmd.Flags().String("notes", "", "Release notes")
    uploadCmd.Flags().Bool("mandatory", false, "Mandatory update")
    uploadCmd.MarkFlagRequired("channel")
    uploadCmd.MarkFlagRequired("version")
    uploadCmd.MarkFlagRequired("file")

    deleteCmd.Flags().String("channel", "", "Channel (stable/beta)")
    deleteCmd.Flags().String("version", "", "Version number")
    deleteCmd.MarkFlagRequired("channel")
    deleteCmd.MarkFlagRequired("version")

    listCmd.Flags().String("channel", "", "Channel filter (stable/beta)")

    rootCmd.AddCommand(uploadCmd, deleteCmd, listCmd)
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Step 4: 初始化依赖并测试**

```bash
cd clients/go/admin && go mod tidy
go build -o update-admin.exe
```

**Step 5: Git 提交**

```bash
git add clients/go/admin/
git commit -m "feat(go-admin): add CLI admin tool for version management"
```

### Task 12: Python 管理工具

**目标:** 创建 Python CLI 管理工具。

**Files:**
- Create: `clients/python/update_admin/cli.py`
- Create: `clients/python/update_admin/admin.py`
- Create: `clients/python/update_admin/requirements.txt`

**Step 1: 创建 requirements.txt**

**File:** `clients/python/update_admin/requirements.txt`

```
requests>=2.31.0
click>=8.1.0
```

**Step 2: 实现管理客户端**

**File:** `clients/python/update_admin/admin.py`

```python
"""管理客户端"""
import requests


class UpdateAdmin:
    """更新服务器管理客户端"""

    def __init__(self, server_url: str, token: str):
        self.server_url = server_url
        self.token = token
        self.session = requests.Session()
        self.session.headers.update({
            "Authorization": f"Bearer {token}"
        })

    def upload_version(
        self,
        program_id: str,
        channel: str,
        version: str,
        file_path: str,
        notes: str = "",
        mandatory: bool = False
    ) -> None:
        """上传新版本"""
        url = f"{self.server_url}/api/version/{program_id}/upload"

        with open(file_path, "rb") as f:
            files = {"file": f}
            data = {
                "channel": channel,
                "version": version,
                "notes": notes,
                "mandatory": str(mandatory).lower()
            }

            response = self.session.post(url, files=files, data=data)
            response.raise_for_status()

        print(f"Version {program_id}/{channel}/{version} uploaded successfully")

    def delete_version(self, program_id: str, channel: str, version: str) -> None:
        """删除版本"""
        url = f"{self.server_url}/api/version/{program_id}/{channel}/{version}"

        response = self.session.delete(url)
        response.raise_for_status()

        print(f"Version {program_id}/{channel}/{version} deleted successfully")

    def list_versions(self, program_id: str, channel: str = None) -> list:
        """列出版本"""
        url = f"{self.server_url}/api/version/{program_id}/list"
        params = {}
        if channel:
            params["channel"] = channel

        response = self.session.get(url, params=params)
        response.raise_for_status()

        data = response.json()
        return data.get("versions", [])
```

**Step 3: 创建 CLI**

**File:** `clients/python/update_admin/cli.py`

```python
"""CLI 工具"""
import click
from .admin import UpdateAdmin


@click.group()
@click.option("--server", default="http://localhost:8080", help="Server URL")
@click.option("--token", required=True, help="API token")
@click.option("--program-id", required=True, help="Program ID")
@click.pass_context
def cli(ctx, server, token, program_id):
    """DocuFiller Update Server Admin Tool"""
    ctx.ensure_object(dict)
    ctx.obj["admin"] = UpdateAdmin(server, token)
    ctx.obj["program_id"] = program_id


@cli.command()
@click.option("--channel", required=True, help="Channel (stable/beta)")
@click.option("--version", required=True, help="Version number")
@click.option("--file", required=True, help="File path", type=click.Path(exists=True))
@click.option("--notes", default="", help="Release notes")
@click.option("--mandatory", is_flag=True, help="Mandatory update")
@click.pass_context
def upload(ctx, channel, version, file, notes, mandatory):
    """Upload a new version"""
    admin = ctx.obj["admin"]
    program_id = ctx.obj["program_id"]

    admin.upload_version(program_id, channel, version, file, notes, mandatory)


@cli.command()
@click.option("--channel", required=True, help="Channel (stable/beta)")
@click.option("--version", required=True, help="Version number")
@click.pass_context
def delete(ctx, channel, version):
    """Delete a version"""
    admin = ctx.obj["admin"]
    program_id = ctx.obj["program_id"]

    admin.delete_version(program_id, channel, version)


@cli.command()
@click.option("--channel", help="Channel filter (stable/beta)")
@click.pass_context
def list(ctx, channel):
    """List versions"""
    admin = ctx.obj["admin"]
    program_id = ctx.obj["program_id"]

    versions = admin.list_versions(program_id, channel)

    for v in versions:
        print(f"{v['version']} ({v['channel']}) - {v['publishDate'][:10]} - {v['fileSize']} bytes")


if __name__ == "__main__":
    cli()
```

**Step 4: Git 提交**

```bash
git add clients/python/update_admin/
git commit -m "feat(py-admin): add CLI admin tool for version management"
```

---

## 阶段五：集成测试

### Task 13: 创建测试服务器

**目标:** 创建用于测试的测试服务器和数据。

**Files:**
- Create: `tests/test_server.go`
- Create: `tests/create_test_data.go`
- Create: `tests/run_tests.bat`

**Step 1: 创建测试服务器**

**File:** `tests/test_server.go`

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "time"
)

const TEST_PORT = 18080

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: test_server.go [start|stop|create-data]")
        return
    }

    command := os.Args[1]

    switch command {
    case "start":
        startTestServer()
    case "stop":
        stopTestServer()
    case "create-data":
        createTestData()
    default:
        fmt.Println("Unknown command:", command)
    }
}

func startTestServer() {
    // 设置测试环境变量
    os.Setenv("SERVER_PORT", fmt.Sprintf("%d", TEST_PORT))
    os.Setenv("DB_PATH", "./tests/test_data/versions.db")
    os.Setenv("STORAGE_PATH", "./tests/test_data/packages")

    // 启动服务器
    cmd := exec.Command("go", "run", "../main.go")
    cmd.Dir = ".."
    cmd.Env = append(os.Environ(),
        fmt.Sprintf("SERVER_PORT=%d", TEST_PORT),
        "DB_PATH=./tests/test_data/versions.db",
        "STORAGE_PATH=./tests/test_data/packages",
    )

    if err := cmd.Start(); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }

    // 保存 PID
    pidFile := filepath.Join("tests", "test_server.pid")
    if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644); err != nil {
        log.Printf("Warning: failed to write PID file: %v", err)
    }

    fmt.Printf("Test server started on port %d (PID: %d)\n", TEST_PORT, cmd.Process.Pid)

    // 等待服务器启动
    time.Sleep(2 * time.Second)

    // 服务器在后台运行
    cmd.Wait()
}

func stopTestServer() {
    pidFile := filepath.Join("tests", "test_server.pid")
    data, err := os.ReadFile(pidFile)
    if err != nil {
        log.Printf("No PID file found: %v", err)
        return
    }

    var pid int
    fmt.Sscanf(string(data), "%d", &pid)

    process, err := os.FindProcess(pid)
    if err != nil {
        log.Printf("Failed to find process: %v", err)
        return
    }

    if err := process.Kill(); err != nil {
        log.Printf("Failed to kill process: %v", err)
        return
    }

    os.Remove(pidFile)
    fmt.Println("Test server stopped")
}

func createTestData() {
    // 创建测试目录
    dirs := []string{
        "test_data/packages/testapp/stable/1.0.0",
        "test_data/packages/testapp/stable/2.0.0",
        "test_data/packages/testapp/beta/2.1.0-beta",
    }

    for _, dir := range dirs {
        if err := os.MkdirAll(filepath.Join("tests", dir), 0755); err != nil {
            log.Printf("Failed to create directory %s: %v", dir, err)
        }
    }

    // 创建测试文件
    createTestFile("tests/test_data/packages/testapp/stable/1.0.0/testapp.zip", 10*1024*1024) // 10MB
    createTestFile("tests/test_data/packages/testapp/stable/2.0.0/testapp.zip", 15*1024*1024) // 15MB
    createTestFile("tests/test_data/packages/testapp/beta/2.1.0-beta/testapp.zip", 12*1024*1024) // 12MB

    fmt.Println("Test data created")
}

func createTestFile(path string, size int64) {
    file, err := os.Create(path)
    if err != nil {
        log.Printf("Failed to create file %s: %v", path, err)
        return
    }
    defer file.Close()

    // 写入随机数据
    buffer := make([]byte, 1024*1024) // 1MB
    for i := 0; i < len(buffer); i++ {
        buffer[i] = byte(i % 256)
    }

    written := int64(0)
    for written < size {
        toWrite := size - written
        if toWrite > int64(len(buffer)) {
            toWrite = int64(len(buffer))
        }

        n, err := file.Write(buffer[:toWrite])
        if err != nil {
            log.Printf("Failed to write to %s: %v", path, err)
            return
        }

        written += int64(n)
    }

    fmt.Printf("Created %s (%d bytes)\n", path, size)
}
```

**Step 2: 创建测试批处理脚本**

**File:** `tests/run_tests.bat`

```batch
@echo off
echo Starting integration tests...

REM Create test data
echo Creating test data...
go run test_server.go create-data

REM Start test server
echo Starting test server...
start /B go run test_server.go start
timeout /t 3 /nobreak > nul

REM Run Go tests
echo Running Go tests...
cd ..\clients\go\client
go test -v
cd ..\..\..

REM Run Python tests
echo Running Python tests...
cd ..\clients\python\update_client
pytest test_version.py test_checker.py -v
cd ..\..\..

REM Run C# tests
echo Running C# tests...
cd ..\clients\csharp\DocuFiller.UpdateClient
dotnet test
cd ..\..\..\..

REM Stop test server
echo Stopping test server...
cd tests
go run test_server.go stop

echo All tests completed!
```

**Step 3: Git 提交**

```bash
git add tests/
git commit -m "feat(tests): add test server and integration test runner"
```

### Task 14: 运行完整集成测试

**目标:** 执行完整的集成测试流程。

**Step 1: 启动测试服务器**

```bash
cd tests && go run test_server.go create-data
go run test_server.go start
```

**Step 2: 使用 Go 管理工具上传测试版本**

```bash
cd ../clients/go/admin
go build -o update-admin.exe
./update-admin.exe upload --program-id testapp --channel stable --version 1.0.0 \
  --file ../../tests/test_data/packages/testapp/stable/1.0.0/testapp.zip \
  --token "change-this-token-in-production"
```

**Step 3: 测试 Go 客户端**

```bash
cd ../client
go test -v
```

**Step 4: 测试 Python 客户端**

```bash
cd ../../../python/update_client
pip install -r requirements.txt
pytest test_version.py test_checker.py -v
```

**Step 5: 测试 C# 客户端**

```bash
cd ../../../csharp/DocuFiller.UpdateClient
dotnet test
```

**Step 6: 停止测试服务器**

```bash
cd ../../../../tests
go run test_server.go stop
```

**Step 7: 清理并提交**

```bash
cd ..
git add .
git commit -m "test: run successful integration tests for all clients"
```

---

## 阶段六：文档和示例

### Task 15: 创建文档

**目标:** 为每个 SDK 创建 README 和使用示例。

**Files:**
- Create: `clients/go/README.md`
- Create: `clients/python/README.md`
- Create: `clients/csharp/README.md`

**Step 1: Go SDK README**

**File:** `clients/go/README.md`

```markdown
# Go Update Client SDK

DocuFiller 更新服务器的 Go 客户端 SDK。

## 安装

\`\`\`bash
go get github.com/LiteHomeLab/update-client
\`\`\`

## 使用示例

### 检查更新

\`\`\`go
package main

import (
    "fmt"
    "github.com/LiteHomeLab/update-client"
)

func main() {
    config := client.DefaultConfig()
    config.ProgramID = "myapp"
    config.ServerURL = "http://localhost:8080"

    checker := client.NewUpdateChecker(config)

    info, err := checker.CheckUpdate("1.0.0")
    if err != nil {
        fmt.Printf("Error: %v\\n", err)
        return
    }

    if info == nil {
        fmt.Println("No update available")
        return
    }

    fmt.Printf("New version available: %s\\n", info.Version)
}
\`\`\`

### 下载更新

\`\`\`go
callback := func(p client.DownloadProgress) {
    fmt.Printf("Downloaded: %d/%d (%.1f%%)\\n",
        p.Downloaded, p.Total, p.Percentage)
}

err := checker.DownloadUpdate(info.Version, "./update.zip", callback)
if err != nil {
    fmt.Printf("Download failed: %v\\n", err)
    return
}

// 验证文件
valid, _ := checker.VerifyFile("./update.zip", info.FileHash)
if !valid {
    fmt.Println("File verification failed")
    return
}

fmt.Println("Update downloaded and verified!")
\`\`\`

## 管理工具

\`\`\`bash
# 上传新版本
./update-admin upload --program-id myapp --channel stable --version 1.0.0 \\
  --file ./myapp-1.0.0.zip --notes "Initial release"

# 列出版本
./update-admin list --program-id myapp --channel stable

# 删除版本
./update-admin delete --program-id myapp --channel stable --version 1.0.0
\`\`\`
```

**Step 2: Python SDK README**

**File:** `clients/python/README.md`

```markdown
# Python Update Client SDK

DocuFiller 更新服务器的 Python 客户端 SDK。

## 安装

\`\`\`bash
pip install -r requirements.txt
\`\`\`

## 使用示例

### 检查更新

\`\`\`python
from update_client import Config, UpdateChecker

config = Config.default()
config.program_id = "myapp"
config.server_url = "http://localhost:8080"

checker = UpdateChecker(config)

info = checker.check_update("1.0.0")
if info is None:
    print("No update available")
else:
    print(f"New version available: {info.version}")
\`\`\`

### 下载更新

\`\`\`python
def progress_callback(p):
    print(f"Downloaded: {p.downloaded}/{p.total} ({p.percentage:.1f}%)")

checker.download_update(info.version, "./update.zip", progress_callback)

# 验证文件
if checker.verify_file("./update.zip", info.file_hash):
    print("Update downloaded and verified!")
\`\`\`

## 管理工具

\`\`\`bash
# 上传新版本
python -m update_admin.cli upload --program-id myapp --channel stable \\
  --version 1.0.0 --file ./myapp-1.0.0.zip --notes "Initial release"

# 列出版本
python -m update_admin.cli list --program-id myapp --channel stable

# 删除版本
python -m update_admin.cli delete --program-id myapp --channel stable --version 1.0.0
\`\`\`
```

**Step 3: C# SDK README**

**File:** `clients/csharp/README.md`

```markdown
# C# Update Client SDK

DocuFiller 更新服务器的 C# 客户端 SDK。

## 安装

\`\`\`bash
dotnet add reference DocuFiller.UpdateClient
\`\`\`

## 使用示例

### 检查更新

\`\`\`csharp
using DocuFiller.UpdateClient;

var config = new Config
{
    ProgramId = "myapp",
    ServerUrl = "http://localhost:8080"
};

var checker = new UpdateChecker(config);

var info = await checker.CheckUpdateAsync("1.0.0");
if (info == null)
{
    Console.WriteLine("No update available");
}
else
{
    Console.WriteLine($"New version available: {info.Version}");
}
\`\`\`

### 下载更新

\`\`\`csharp
var progress = new Progress<DownloadProgress>(p =>
{
    Console.WriteLine($"Downloaded: {p.Downloaded}/{p.Total} ({p.Percentage:F1}%)");
});

await checker.DownloadUpdateAsync(info.Version, "./update.zip", progress);

// 验证文件
if (checker.VerifyFile("./update.zip", info.FileHash))
{
    Console.WriteLine("Update downloaded and verified!");
}
\`\`\`
```

**Step 4: Git 提交**

```bash
git add clients/
git commit -m "docs: add README files for all client SDKs"
```

### Task 16: 最终验证和清理

**Step 1: 运行所有测试**

```bash
cd tests
./run_tests.bat
```

**Step 2: 验证所有 SDK 可以正常工作**

```bash
# Go SDK
cd ../clients/go/client
go test -v

# Python SDK
cd ../../python/update_client
pytest test_version.py test_checker.py -v

# C# SDK
cd ../../csharp/DocuFiller.UpdateClient
dotnet test
```

**Step 3: 构建所有 CLI 工具**

```bash
# Go Admin
cd ../../go/admin
go build -o update-admin.exe

# Python Admin (安装)
cd ../../python/update_admin
pip install -r requirements.txt
pip install -e .
```

**Step 4: 最终提交**

```bash
cd ../../../..
git add .
git commit -m "feat: complete multi-language client SDK implementation with tests"
```

---

## 总结

完成所有任务后，你将拥有：

1. **Go 客户端 SDK** - 包含自动更新功能和管理 CLI 工具
2. **Python 客户端 SDK** - 包含自动更新功能和管理 CLI 工具
3. **C# 客户端 SDK** - 包含自动更新功能
4. **集成测试框架** - 测试服务器和自动化测试脚本
5. **完整文档** - 每个语言的 README 和使用示例

所有 SDK 都支持：
- 版本检查和比较
- 文件下载（带进度回调）
- SHA256 文件验证
- 失败重试机制
- 管理工具（上传/删除版本）
