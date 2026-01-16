# Go Update Client SDK

DocuFiller 更新服务器的 Go 客户端 SDK。

## 安装

```bash
go get github.com/LiteHomeLab/update-client
```

## 使用示例

### 检查更新

```go
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
        fmt.Printf("Error: %v\n", err)
        return
    }

    if info == nil {
        fmt.Println("No update available")
        return
    }

    fmt.Printf("New version available: %s\n", info.Version)
}
```

### 下载更新

```go
callback := func(p client.DownloadProgress) {
    fmt.Printf("Downloaded: %d/%d (%.1f%%)\n",
        p.Downloaded, p.Total, p.Percentage)
}

err := checker.DownloadUpdate(info.Version, "./update.zip", callback)
if err != nil {
    fmt.Printf("Download failed: %v\n", err)
    return
}

// 验证文件
valid, _ := checker.VerifyFile("./update.zip", info.FileHash)
if !valid {
    fmt.Println("File verification failed")
    return
}

fmt.Println("Update downloaded and verified!")
```

## 管理工具

```bash
# 上传新版本
./update-admin upload --program-id myapp --channel stable --version 1.0.0 \
  --file ./myapp-1.0.0.zip --notes "Initial release"

# 列出版本
./update-admin list --program-id myapp --channel stable

# 删除版本
./update-admin delete --program-id myapp --channel stable --version 1.0.0
```
