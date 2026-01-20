# Daemon Mode 快速使用指南

## 什么是 Daemon 模式？

Daemon 模式允许 update-client 在后台运行一个 HTTP 服务器，这样可以：
- 从主应用程序控制下载过程
- 实时查询下载状态
- 优雅地停止下载
- 监控父进程生命周期

## 快速开始

### 1. 构建客户端

```bash
# 使用构建脚本
build-client.bat

# 或手动构建
cd cmd/update-client
go build -o ../../bin/update-client.exe .
```

### 2. 启动 Daemon 模式

```bash
update-client.exe download --daemon --port 19876 --version 1.0.5
```

**参数说明**:
- `--daemon`: 启用 daemon 模式（必须）
- `--port`: HTTP 服务器端口（必须）
- `--version`: 要下载的版本（必须）

**预期输出**:
```
2026/01/20 17:46:00 Daemon server started on port 19876
```

### 3. 查询状态

在另一个终端或使用 HTTP 客户端：

```bash
curl http://localhost:19876/status
```

**响应**:
```json
{
  "state": "downloading",
  "version": "1.0.5"
}
```

**状态值**:
- `idle`: 未开始下载
- `downloading`: 正在下载
- `completed`: 下载完成
- `error`: 下载失败

### 4. 停止服务器

```bash
curl -X POST http://localhost:19876/shutdown
```

**响应**:
```json
{
  "success": true,
  "message": "Server shutting down"
}
```

## API 端点

### GET /status

查询下载状态。

**响应**:
```json
{
  "state": "downloading",
  "version": "1.0.5",
  "error": ""
}
```

### POST /shutdown

优雅地停止服务器。

**响应**:
```json
{
  "success": true,
  "message": "Server shutting down"
}
```

## 配置文件

创建 `update-config.yaml`:

```yaml
server:
  url: http://localhost:8080
  timeout: 30

program:
  id: docufiller
  current_version: 0.0.1

download:
  save_path: ./downloads
  naming: version
  keep: 3
  auto_verify: true

logging:
  level: info
  file: update-client.log
```

## 常见问题

### Q: 端口已被占用怎么办？

A: 使用不同的端口号：
```bash
update-client.exe download --daemon --port 19877 --version 1.0.5
```

### Q: 如何在后台运行？

A: Windows 下使用 start 命令：
```cmd
start /B update-client.exe download --daemon --port 19876 --version 1.0.5
```

### Q: 如何查看详细日志？

A: 检查日志文件：
```bash
tail -f update-client.log
```

### Q: 服务器没有响应怎么办？

A: 检查防火墙设置或使用 `netstat -ano | grep 19876` 检查端口状态。

## 故障排查

### 问题：无法启动

**症状**: 端口冲突

**解决**:
```bash
# 查找占用端口的进程
netstat -ano | findstr :19876

# 终止进程（PID）
taskkill /PID <pid> /F
```

### 问题：下载失败

**症状**: state 为 "error"

**解决**:
1. 检查服务器 URL 是否正确
2. 检查版本是否存在
3. 检查网络连接
4. 查看日志文件

### 问题：服务器无法关闭

**症状**: shutdown 请求无响应

**解决**:
```bash
# 强制终止进程
taskkill /F /IM update-client.exe
```

## 示例用法

### PowerShell 示例

```powershell
# 启动 daemon
Start-Process -FilePath "update-client.exe" -ArgumentList "download --daemon --port 19876 --version 1.0.5"

# 检查状态
Invoke-WebRequest -Uri "http://localhost:19876/status" | Select-Object -ExpandProperty Content

# 停止服务器
Invoke-WebRequest -Uri "http://localhost:19876/shutdown" -Method Post
```

### C# 示例

```csharp
using System.Net.Http;

// 启动进程
var process = Process.Start("update-client.exe", "download --daemon --port 19876 --version 1.0.5");

// 检查状态
using var client = new HttpClient();
var response = await client.GetStringAsync("http://localhost:19876/status");

// 停止服务器
await client.PostAsync("http://localhost:19876/shutdown", null);
```

## 性能建议

1. **端口选择**: 使用 1024-49151 之间的注册端口
2. **超时设置**: 根据网络环境调整 server.timeout
3. **日志级别**: 生产环境使用 "info" 或 "warn"
4. **文件清理**: 定期清理下载目录

## 安全建议

1. **防火墙**: 只允许本地访问（127.0.0.1）
2. **认证**: 如需远程访问，添加认证机制
3. **HTTPS**: 生产环境使用 TLS
4. **输入验证**: 验证所有输入参数

## 更多信息

- 详细测试报告: `DAEMON_TEST_REPORT.md`
- 架构文档: `docs/DAEMON_ARCHITECTURE.md`
- API 文档: `docs/API.md`
