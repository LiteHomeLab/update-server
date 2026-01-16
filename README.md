# DocuFiller 更新服务器

用于 DocuFiller WPF 应用程序的自动更新系统后端服务。

## 📋 简介

DocuFiller 更新服务器是一个基于 Go 语言开发的 RESTful API 服务器，为 DocuFiller 客户端提供自动更新功能。服务器支持多渠道版本发布（稳定版/测试版）、版本管理、文件存储和下载统计等功能。

### 核心特性

- 🔐 **安全的 API 认证** - Bearer Token 保护敏感操作
- 📦 **多渠道支持** - 同时支持 stable（稳定版）和 beta（测试版）发布渠道
- 🗄️ **SQLite 数据库** - 轻量级数据库，易于部署和维护
- 📝 **完整的日志系统** - 基于 WQGroup/logger 的结构化日志
- 📊 **下载统计** - 记录每个版本的下载次数
- 🔒 **文件完整性验证** - SHA256 哈希确保文件未被篡改
- 🚀 **高性能** - 基于 Gin 框架的高性能 HTTP 服务

## 🛠️ 技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.23+ |
| Web 框架 | Gin | v1.11.0 |
| ORM | GORM | v1.31.1 |
| 数据库 | SQLite | v1.6.0 |
| 日志库 | WQGroup/logger | v0.0.16 |
| 配置管理 | YAML | v3.0.1 |

## 📁 项目结构

```
docufiller-update-server/
├── main.go                          # 程序入口
├── go.mod                           # 依赖管理
├── go.sum                           # 依赖锁定
├── config.yaml                      # 配置文件
├── Makefile                         # 构建脚本
├── internal/
│   ├── config/
│   │   └── config.go                # 配置加载
│   ├── models/
│   │   └── version.go               # GORM 数据模型
│   ├── database/
│   │   └── gorm.go                  # GORM 初始化
│   ├── handler/
│   │   └── version.go               # API 处理器
│   ├── service/
│   │   ├── version.go               # 版本业务逻辑
│   │   └── storage.go               # 文件存储服务
│   ├── middleware/
│   │   └── auth.go                  # 认证中间件
│   └── logger/
│       └── logger.go                # 日志初始化
├── data/
│   ├── versions.db                  # SQLite 数据库
│   └── packages/                    # 安装包存储
│       ├── stable/                  # 稳定版存储
│       └── beta/                    # 测试版存储
└── logs/                            # 日志文件
```

## 🚀 快速开始

### 前置要求

- Go 1.23 或更高版本
- Windows/Linux/macOS 操作系统

### 安装

1. **克隆仓库**
   ```bash
   git clone https://github.com/allanpk716/docx_replacer.git
   cd docx_replacer/docufiller-update-server
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **配置服务器**

   编辑 `config.yaml` 文件：

   ```yaml
   server:
     port: 8080              # 服务端口
     host: "0.0.0.0"         # 监听地址

   database:
     path: "./data/versions.db"  # 数据库路径

   storage:
     basePath: "./data/packages"  # 文件存储路径
     maxFileSize: 536870912       # 最大文件大小 (512MB)

   api:
     uploadToken: "change-this-token-in-production"  # 上传认证令牌（必须修改！）
     corsEnable: true                       # 是否启用 CORS

   logger:
     level: "info"          # 日志级别: trace, debug, info, warn, error
     output: "both"         # 输出方式: console, file, both
     filePath: "./logs/server.log"
     maxSize: 10485760     # 日志文件最大大小 (10MB)
     maxBackups: 5         # 保留的日志文件数量
     maxAge: 30            # 日志文件保留天数
     compress: true        # 是否压缩旧日志
   ```

4. **运行服务器**

   开发环境：
   ```bash
   go run main.go
   ```

   生产环境：
   ```bash
   make build
   ./bin/docufiller-update-server
   ```

5. **验证服务**

   ```bash
   curl http://localhost:8080/api/health
   # 预期输出: {"status":"ok"}
   ```

## 📡 API 文档

### 健康检查

```
GET /api/health
```

**响应示例:**
```json
{
  "status": "ok"
}
```

### 获取最新版本

```
GET /api/version/latest?channel={channel}
```

**参数:**
- `channel` (必需): 发布渠道，`stable` 或 `beta`

**响应示例:**
```json
{
  "version": "1.2.0",
  "channel": "stable",
  "fileName": "docufiller-1.2.0.zip",
  "fileSize": 52428800,
  "fileHash": "a1b2c3d4e5f6...",
  "releaseNotes": "修复了xxx问题",
  "publishDate": "2025-01-15T10:00:00Z",
  "mandatory": false,
  "downloadCount": 42
}
```

### 获取版本列表

```
GET /api/version/list?channel={channel}
```

**参数:**
- `channel` (可选): 筛选特定渠道，不提供则返回所有版本

**响应示例:**
```json
{
  "versions": [
    {
      "version": "1.2.0",
      "channel": "stable",
      "fileName": "docufiller-1.2.0.zip",
      "fileSize": 52428800,
      "publishDate": "2025-01-15T10:00:00Z"
    }
  ],
  "total": 1
}
```

### 上传新版本

```
POST /api/version/upload
```

**认证:** Bearer Token (必需)

**请求参数 (multipart/form-data):**
- `channel` (必需): 发布渠道 (`stable` 或 `beta`)
- `version` (必需): 版本号 (如 `1.2.0`)
- `file` (必需): 安装包文件 (.zip)
- `mandatory` (可选): 是否强制更新，默认 `false`
- `releaseNotes` (可选): 发布说明

**请求示例:**
```bash
curl -X POST "http://localhost:8080/api/version/upload" \
  -H "Authorization: Bearer your-secret-token" \
  -F "channel=stable" \
  -F "version=1.2.0" \
  -F "file=@docufiller-1.2.0.zip" \
  -F "mandatory=false" \
  -F "releaseNotes=修复了xxx问题"
```

**响应示例:**
```json
{
  "message": "Version uploaded successfully",
  "version": {
    "version": "1.2.0",
    "channel": "stable",
    "fileName": "docufiller-1.2.0.zip",
    "filePath": "./data/packages/stable/1.2.0/docufiller-1.2.0.zip",
    "fileSize": 52428800,
    "fileHash": "a1b2c3d4e5f6...",
    "downloadCount": 0
  }
}
```

### 下载安装包

```
GET /api/download/{channel}/{version}
```

**参数:**
- `channel`: 发布渠道
- `version`: 版本号

**响应:** 直接返回文件流

### 删除版本

```
DELETE /api/version/{channel}/{version}
```

**认证:** Bearer Token (必需)

**响应示例:**
```json
{
  "message": "Version deleted successfully"
}
```

## 🔧 构建和部署

### 构建

```bash
make build
```

这将编译服务器并输出到 `./bin/docufiller-update-server`

### Windows 服务部署

使用 [NSSM](https://nssm.cc/) 将服务器注册为 Windows 服务：

```batch
nssm install DocuFillerUpdateServer "C:\path\to\docufiller-update-server.exe"
nssm set DocuFillerUpdateServer AppDirectory "C:\path\to\docufiller-update-server"
nssm set DocuFillerUpdateServer AppEnvironmentExtra "CONFIG_FILE=C:\path\to\config.yaml"
nssm start DocuFillerUpdateServer
```

### Linux systemd 服务部署

创建 `/etc/systemd/system/docufiller-update-server.service`:

```ini
[Unit]
Description=DocuFiller Update Server
After=network.target

[Service]
Type=simple
User=docufiller
WorkingDirectory=/opt/docufiller-update-server
ExecStart=/opt/docufiller-update-server/docufiller-update-server
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

启用并启动服务：

```bash
sudo systemctl enable docufiller-update-server
sudo systemctl start docufiller-update-server
sudo systemctl status docufiller-update-server
```

## 🔐 安全建议

### 生产环境部署前

1. **修改默认 Token**

   在 `config.yaml` 中设置强密码：
   ```yaml
   api:
     uploadToken: "use-a-strong-random-password-here"
   ```

2. **使用 HTTPS**

   在服务器前配置反向代理（如 Nginx）并启用 SSL/TLS

3. **限制访问**

   配置防火墙规则，限制对上传端点的访问

4. **定期备份**

   备份 `data/versions.db` 数据库文件

5. **监控日志**

   定期检查 `logs/` 目录中的日志文件

## 📊 监控和维护

### 查看日志

日志文件位于 `logs/` 目录：
- `server-YYYY-MM-DD.log` - 当日日志
- `server-YYYY-MM-DD.log.gz` - 历史压缩日志

### 数据库查询

使用 SQLite 客户端查询版本信息：

```bash
sqlite3 data/versions.db
```

示例查询：

```sql
-- 查看所有版本
SELECT version, channel, publish_date, download_count
FROM versions
ORDER BY publish_date DESC;

-- 查看下载统计
SELECT channel, COUNT(*) as count, SUM(download_count) as total_downloads
FROM versions
GROUP BY channel;
```

### 存储管理

定期清理旧版本以释放存储空间：

```bash
# 删除超过 90 天的旧版本
# (需要手动实现或编写脚本)
```

## 🧪 开发

### 运行测试

```bash
go test ./...
```

### 代码格式化

```bash
go fmt ./...
gofmt -w .
```

### 添加新依赖

```bash
go get github.com/some/package
go mod tidy
```

## ❓ 常见问题

### Q: 如何修改服务器端口？

A: 编辑 `config.yaml` 中的 `server.port` 配置项。

### Q: 数据库文件在哪里？

A: 默认位于 `./data/versions.db`，可在 `config.yaml` 中修改路径。

### Q: 如何重置上传 Token？

A: 编辑 `config.yaml` 中的 `api.uploadToken` 并重启服务器。

### Q: 上传文件大小有限制吗？

A: 默认限制为 512MB，可在 `config.yaml` 的 `storage.maxFileSize` 中修改。

### Q: 支持并发上传吗？

A: 支持，服务器使用 SQLite 的并发事务处理。

### Q: 日志文件会无限增长吗？

A: 不会，日志系统会自动轮转、压缩和清理旧日志（根据 `logger.maxBackups` 和 `logger.maxAge` 配置）。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](../LICENSE) 文件。

## 🔗 相关链接

- [DocuFiller 主项目](https://github.com/allanpk716/docx_replacer)
- [系统设计文档](../docs/plans/2025-01-15-auto-update-system-design.md)
- [实施计划](../docs/plans/2025-01-15-auto-update-implementation.md)

## 📧 联系方式

- 作者: Allan
- 项目主页: https://github.com/allanpk716/docx_replacer

---

**最后更新:** 2025-01-15
