# 更新服务器配置与维护指南

本文档说明如何配置、部署和维护 DocuFiller 更新服务器。

## 目录

- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [维护操作](#维护操作)
- [故障排查](#故障排查)

---

## 快速开始

### 1. 编译服务器

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\LiteHomeLab\update-server
go build -o bin/update-server.exe main.go
```

### 2. 生成 Token

```bash
go run cmd/gen-token/main.go
```

输出示例：
```
Admin Token: db2d387ff07aed70562da78115a45edd2821740ebd3233e9dac4cb163eec67cc
Token ID: 1022d95b8439843d2e385fa56b7b3ec90b2a36ab0a3486a98033fade8b782652
```

**重要**：保存此 Token，后续配置需要使用。

### 3. 创建程序记录

```bash
curl -X POST http://localhost:8080/api/programs \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "programId": "docufiller",
    "name": "DocuFiller",
    "description": "Word document template filling tool"
  }'
```

### 4. 启动服务器

```bash
.\bin\update-server.exe
```

服务器将在 `http://localhost:8080` 启动。

---

## 配置说明

### config.yaml 配置文件

```yaml
# 服务器配置
server:
  port: 8080              # 监听端口
  host: "0.0.0.0"         # 监听地址 (0.0.0.0 表示所有网卡)

# 数据库配置
database:
  path: "./data/versions.db"  # SQLite 数据库文件路径

# 存储配置
storage:
  basePath: "./data/packages"  # 文件存储根目录
  maxFileSize: 536870912       # 最大文件大小 (512MB)

# API 配置
api:
  uploadToken: "change-this-token-in-production"  # ⚠️ 已弃用，使用数据库 Token
  corsEnable: true                         # 是否启用 CORS

# 日志配置
logger:
  level: "info"          # 日志级别: debug, info, warn, error
  output: "both"         # 输出目标: console, file, both
  filePath: "./logs/server.log"
  maxSize: 10485760      # 单个日志文件最大 10MB
  maxBackups: 5          # 保留 5 个备份
  maxAge: 30             # 保留 30 天
  compress: true         # 压缩旧日志

# 加密配置
crypto:
  masterKey: "change-this-to-a-secure-32-byte-key-in-production"  # 主加密密钥
```

### 重要配置项说明

#### 端口配置

局域网部署时，保持默认即可：
```yaml
server:
  port: 8080
  host: "0.0.0.0"  # 允许外部访问
```

#### Token 管理

**重要**：`api.uploadToken` 配置项已弃用！

使用数据库 Token 系统：
- Admin Token：拥有所有权限
- 程序 Token：仅能管理特定程序

生成 Token：
```bash
# 生成 Admin Token
go run cmd/gen-token/main.go
```

---

## 维护操作

### 1. 版本管理

#### 查看版本列表

```bash
.\bin\upload-admin.exe list --program-id docufiller --channel stable \
  --server http://localhost:8080 \
  --token <ADMIN_TOKEN>
```

#### 删除版本

```bash
.\bin\upload-admin.exe delete --program-id docufiller --version 1.0.0 --channel stable \
  --server http://localhost:8080 \
  --token <ADMIN_TOKEN>
```

#### 手动上传版本

```bash
.\bin\upload-admin.exe upload \
  --program-id docufiller \
  --channel stable \
  --version 1.0.0 \
  --file path\to\package.zip \
  --server http://localhost:8080 \
  --token <ADMIN_TOKEN> \
  --notes "Release notes here" \
  --mandatory=false
```

### 2. 数据库维护

#### 清理软删除记录

如果发现删除后无法重新上传相同版本，运行清理脚本：

```bash
go run scripts/cleanup-soft-deleted.go
```

输出示例：
```
Successfully cleaned up 3 soft-deleted version records
```

#### 备份数据库

```bash
# 备份
copy data\versions.db data\versions.db.backup

# 恢复
copy data\versions.db.backup data\versions.db
```

### 3. 日志管理

#### 查看日志

```bash
# 实时查看（Windows PowerShell）
Get-Content .\logs\server.log -Wait -Tail 50

# 或使用文本编辑器打开
notepad .\logs\server.log
```

#### 日志级别调整

开发时使用 debug 级别：
```yaml
logger:
  level: "debug"
```

生产环境使用 info 级别：
```yaml
logger:
  level: "info"
```

### 4. 文件存储管理

#### 查看存储占用

```bash
# Windows PowerShell
Get-ChildItem -Path .\data\packages -Recurse | Measure-Object -Property Length -Sum
```

#### 清理旧版本文件

```bash
# 删除超过 90 天的版本（需要手动编写脚本或使用 upload-admin 逐个删除）
```

---

## 故障排查

### 问题 1：端口被占用

**错误信息**：
```
listen tcp 0.0.0.0:8080: bind: Only one usage of each socket address
```

**解决方案**：

方式 1：终止占用进程
```bash
# 查找进程
netstat -ano | findstr ":8080"

# 终止进程
taskkill /F /PID <进程ID>
```

方式 2：更换端口
```yaml
# 编辑 config.yaml
server:
  port: 8081  # 改为其他端口
```

### 问题 2：Token 无效

**错误信息**：
```
{"error":"invalid token"}
```

**解决方案**：

1. 检查 Token 是否正确复制
2. 重新生成 Token：
```bash
go run cmd/gen-token/main.go
```

### 问题 3：唯一约束冲突

**错误信息**：
```
UNIQUE constraint failed: versions.program_id, versions.version, versions.channel
```

**解决方案**：

运行清理脚本删除软删除记录：
```bash
go run scripts/cleanup-soft-deleted.go
```

### 问题 4：文件上传失败

**可能原因**：

1. **文件太大**：检查 `config.yaml` 中的 `maxFileSize` 配置
2. **磁盘空间不足**：检查磁盘空间
3. **权限问题**：确保进程有写入 `data/packages` 的权限

**解决方案**：

```bash
# 检查磁盘空间
wmic logicaldisk get size,freespace,caption

# 检查目录权限
icacls .\data\packages
```

### 问题 5：数据库锁定

**错误信息**：
```
database is locked
```

**解决方案**：

1. 确保只有一个服务器实例在运行
2. 重启服务器
3. 如果问题持续，删除数据库文件的锁：
```bash
# 停止服务器
taskkill /F /IM update-server.exe

# 删除锁文件（如果存在）
del .\data\versions.db-shm
del .\data\versions.db-wal

# 重新启动
.\bin\update-server.exe
```

---

## 性能优化

### 1. 数据库优化

SQLite 默认配置适用于中小规模（<1000 版本）。如需更大规模：

考虑迁移到 MySQL 或 PostgreSQL（需要修改 `internal/database/` 代码）

### 2. 文件存储优化

对于大型文件存储：
- 使用 SSD 存储
- 定期清理旧版本
- 考虑使用对象存储（如 MinIO）

### 3. 网络优化

局域网部署时，确保：
- 使用千兆网络
- 服务器和客户端在同一网段
- 防火墙允许端口访问

---

## 安全建议

1. **保护 Token**
   - 不要将 Token 提交到版本控制
   - 定期轮换 Token
   - 使用环境变量存储 Token

2. **网络安全**
   - 仅在局域网使用
   - 配置防火墙规则
   - 考虑使用 HTTPS（需要反向代理）

3. **文件验证**
   - 服务器会验证文件 SHA256 哈希
   - 客户端应验证下载文件的完整性

4. **访问控制**
   - 限制 upload-admin.exe 的访问
   - 记录所有上传/删除操作

---

## 附录

### A. 目录结构

```
update-server/
├── bin/
│   ├── update-server.exe       # 服务器可执行文件
│   └── upload-admin.exe        # 管理工具
├── data/
│   ├── versions.db             # 版本数据库
│   └── packages/               # 文件存储
│       └── docufiller/
│           ├── stable/
│           │   └── 1.0.0/
│           │       └── docufiller-1.0.0.zip
│           └── beta/
├── logs/
│   └── server.log              # 服务器日志
├── config.yaml                 # 配置文件
├── internal/
│   ├── handler/                # HTTP 处理器
│   ├── service/                # 业务逻辑
│   │   └── version.go         # 版本服务（硬删除修复）
│   ├── middleware/             # 中间件
│   └── models/                 # 数据模型
└── scripts/
    ├── cleanup-soft-deleted.go # 清理脚本
    └── migrate.go              # 数据库迁移
```

### B. 端口占用查询

```bash
# Windows 查询端口占用
netstat -ano | findstr ":8080"

# 查看具体进程
tasklist | findstr "<进程ID>"
```

### C. 服务注册（可选）

将更新服务器注册为 Windows 服务：

```bash
# 使用 NSSM 工具
# 下载：https://nssm.cc/download

nssm install DocuFillerUpdateServer "C:\Path\to\update-server.exe"
nssm set DocuFillerUpdateServer AppDirectory "C:\Path\to\update-server"
nssm start DocuFillerUpdateServer
```

---

**文档版本**：1.0
**最后更新**：2026-01-20
**维护者**：Claude Code
