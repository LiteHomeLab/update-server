# Update Server 架构说明

## 系统概述

Update Server 是一个通用的应用程序自动更新系统，采用三层架构设计，支持多程序管理、Web管理界面和端到端加密。

```
┌─────────────────────────────────────────────────────────────┐
│                    update-server                            │
│              (中央更新服务器 - Go + Gin)                     │
│                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ Web 管理界面 │  │   API 服务   │  │  文件存储    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
                           ↑                ↓
                    管理员使用           存储加密文件
                           │
        ┌──────────────────┼──────────────────┐
        ↓                  ↓                  ↓
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│ Program A     │  │ Program B     │  │ Program C     │
│               │  │               │  │               │
│ ┌───────────┐ │  │ ┌───────────┐ │  │ ┌───────────┐ │
│ │publish-cli│ │  │ │publish-cli│ │  │ │publish-cli│ │
│ └───────────┘ │  │ └───────────┘ │  │ └───────────┘ │
│ ┌───────────┐ │  │ ┌───────────┐ │  │ ┌───────────┐ │
│ │update-cli │ │  │ │update-cli │ │  │ │update-cli │ │
│ └───────────┘ │  │ └───────────┘ │  │ └───────────┘ │
└───────────────┘  └───────────────┘  └───────────────┘
```

## 组件说明

### 1. update-server（中央服务器）

**职责**：
- 托管所有程序的更新包
- 提供Web管理界面
- 管理程序、版本、Token、密钥
- 处理上传和下载请求

**技术栈**：
- Go 1.23+
- Gin Web框架
- GORM ORM
- SQLite数据库
- 嵌入式前端（纯JavaScript/CSS）

### 2. publish-client（发布端）

**职责**：
- 上传新版本到服务器
- 上传前使用密钥加密文件
- 支持命令行配置

**分发方式**：
- 从Web管理后台下载
- 预配置的ZIP包（包含exe + config.yaml + README.md）

### 3. update-client（更新端）

**职责**：
- 检查是否有新版本
- 下载更新包
- 解密下载的文件
- 返回更新信息

**分发方式**：
- 从Web管理后台下载
- 预配置的ZIP包（包含exe + config.yaml + README.md）

## 认证机制

### Token 类型

| Token 类型 | 用途 | 权限 |
|-----------|------|------|
| Upload Token | 上传特定程序的版本 | 上传/删除版本 |
| Download Token | 下载特定程序的版本 | 下载文件 |
| Admin Token | Web管理后台认证 | 所有权限 |

### Token 生成流程

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ 生成随机Token值 │ →  │ 计算SHA256哈希  │ →  │ 存入数据库      │
│ 64位十六进制     │    │ 作为Token ID    │    │ 只存储哈希      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Token 认证流程

```
客户端请求
    │
    ├─ HTTP Header: Authorization: Bearer <token>
    ↓
服务器提取Token
    │
    ├─ 计算 SHA256(Token)
    ↓
查询数据库
    │
    ├─ SELECT * FROM tokens WHERE token_id = <hash>
    ↓
验证权限
    │
    ├─ 检查 Token 类型和程序ID
    ↓
允许/拒绝访问
```

### 安全设计

- 数据库只存储Token的SHA256哈希，不存储原始Token
- Token可通过Web界面重新生成
- 每个程序有独立的Upload和Download Token

## 加密机制

### 端到端加密流程

```
┌─────────────┐                ┌─────────────┐                ┌─────────────┐
│publish-cli  │                │   服务器     │                │update-cli  │
└──────┬──────┘                └──────┬──────┘                └──────┬──────┘
       │                              │                              │
       │ 1. 读取Encryption Key        │                              │
       │    (从配置文件)              │                              │
       │                              │                              │
       │ 2. AES-256-GCM加密文件       │                              │
       │                              │                              │
       │ 3. 上传加密文件 ────────────▶ │ 存储加密文件                │
       │                              │ (无法解密)                   │
       │                              │                              │
       │                              │ 4. 请求下载 ──────────────▶ │
       │                              │                              │
       │                              │ 5. 返回加密文件 ───────────▶ │
       │                              │                              │
       │                              │                              │ 6. 读取Encryption Key
       │                              │                              │    (从配置文件)
       │                              │                              │
       │                              │                              │ 7. AES-256-GCM解密
       │                              │                              │
       │                              │                              │ 8. 得到原始文件
```

### 加密算法

- **算法**：AES-256-GCM
- **密钥长度**：32字节（256位）
- **密钥生成**：每个程序创建时自动生成
- **密钥存储**：Base64编码后存储在数据库
- **密钥分发**：通过Web界面下载预配置的客户端工具

### 密钥管理

```yaml
# 程序创建时自动生成
encryption:
  enabled: true
  key: "base64编码的32字节密钥"  # 自动生成，写入配置文件
```

- 每个程序有独立的加密密钥
- 密钥可在Web界面重新生成
- 服务器只存储加密文件，无法解密

## 数据模型

### programs 表
```sql
CREATE TABLE programs (
  id INTEGER PRIMARY KEY,
  program_id TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### encryption_keys 表
```sql
CREATE TABLE encryption_keys (
  id INTEGER PRIMARY KEY,
  program_id TEXT UNIQUE NOT NULL,
  key_data TEXT NOT NULL,  -- Base64编码的密钥
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (program_id) REFERENCES programs(program_id)
);
```

### tokens 表
```sql
CREATE TABLE tokens (
  id INTEGER PRIMARY KEY,
  token_id TEXT UNIQUE NOT NULL,      -- SHA256哈希
  token_value TEXT NOT NULL,          -- 原始Token（可选）
  program_id TEXT,
  token_type TEXT NOT NULL,           -- 'upload', 'download', 'admin'
  is_active BOOLEAN DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (program_id) REFERENCES programs(program_id)
);
```

### versions 表
```sql
CREATE TABLE versions (
  id INTEGER PRIMARY KEY,
  program_id TEXT NOT NULL,
  version TEXT NOT NULL,
  channel TEXT NOT NULL,              -- 'stable' or 'beta'
  file_path TEXT NOT NULL,
  file_size INTEGER,
  file_hash TEXT,
  changelog TEXT,
  download_count INTEGER DEFAULT 0,
  mandatory BOOLEAN DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (program_id) REFERENCES programs(program_id),
  UNIQUE(program_id, version, channel)
);
```

### admin_users 表
```sql
CREATE TABLE admin_users (
  id INTEGER PRIMARY KEY,
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,        -- bcrypt哈希
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Web 管理界面

### 架构

- **技术**：纯JavaScript + CSS，无框架依赖
- **部署**：通过Go embed.FS嵌入到二进制文件
- **通信**：RESTful API + Cookie认证

### 页面结构

| 页面 | 路径 | 功能 |
|------|------|------|
| 初始化向导 | /setup | 首次启动配置 |
| 登录页面 | /login | 管理员认证 |
| 仪表盘 | / | 系统概览 |
| 程序管理 | /programs | 程序CRUD |
| 程序详情 | /programs/:id | Token/版本/下载管理 |
| 系统设置 | /settings | 服务器配置 |

### 客户端工具打包流程

```
┌─────────────────────┐
│ Web管理后台请求下载  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ 读取服务器配置       │
│ - ServerURL         │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ 从数据库读取        │
│ - ProgramID         │
│ - Token             │
│ - Encryption Key    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ 生成配置文件         │
│ publish-config.yaml │
│ 或 update-config.yaml│
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ 打包ZIP             │
│ - xxx.exe           │
│ - xxx-config.yaml   │
│ - README.md         │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ 返回ZIP下载          │
└─────────────────────┘
```

## API 端点

### 公开端点
- `GET /api/health` - 健康检查
- `GET /api/programs/{id}/versions/latest` - 获取最新版本
- `GET /api/programs/{id}/versions` - 获取版本列表

### 认证端点（Token）
- `POST /api/programs/{id}/versions` - 上传版本（Upload Token）
- `DELETE /api/programs/{id}/versions/{channel}/{version}` - 删除版本（Upload Token）
- `GET /api/programs/{id}/download/{channel}/{version}` - 下载文件（Download Token）

### 管理端点（Web登录）
- `POST /api/programs` - 创建程序
- `GET /api/programs` - 程序列表
- `DELETE /api/programs/{id}` - 删除程序
- `POST /api/programs/{id}/tokens/regenerate` - 重新生成Token
- `GET /api/programs/{id}/clients/download` - 下载客户端工具

## 配置文件

### 服务器端 config.yaml

```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  path: "./data/versions.db"

storage:
  basePath: "./data/packages"
  maxFileSize: 536870912  # 512MB

serverUrl: "https://update.example.com"  # 客户端连接地址
adminInitialized: false                  # 初始化状态
clientsDirectory: "./clients"            # 客户端工具位置

logger:
  level: "info"
  output: "both"
  filePath: "./logs/server.log"
  maxSize: 10485760
  maxBackups: 5
  maxAge: 30
  compress: true
```

### 发布端 publish-config.yaml

```yaml
server: "https://update.example.com"
programId: "docufiller"
uploadToken: "ul_xxxxxxxxxxxxx"
encryption:
  enabled: true
  key: "base64密钥"
file: "./app-v1.0.0.zip"
version: "1.0.0"
channel: "stable"
changelog: "更新说明"
```

### 更新端 update-config.yaml

```yaml
server: "https://update.example.com"
programId: "docufiller"
downloadToken: "dl_xxxxxxxxxxxxx"
encryption:
  enabled: true
  key: "base64密钥"
check:
  channel: "stable"
  autoDownload: true
download:
  outputPath: "./updates"
```

## 部署架构

### 单服务器部署

```
┌─────────────────────────────────────┐
│         用户局域网                   │
│                                     │
│  ┌─────────────────────────────┐   │
│  │   update-server.exe         │   │
│  │   - Web UI (端口8080)       │   │
│  │   - API 服务                │   │
│  │   - SQLite 数据库           │   │
│  │   - 文件存储                │   │
│  └─────────────────────────────┘   │
│                                     │
│  管理员: http://server:8080         │
│  客户端: http://server:8080/api     │
└─────────────────────────────────────┘
```

### 生产环境建议

1. **反向代理**：使用Nginx配置HTTPS
2. **数据库**：考虑升级到PostgreSQL/MySQL
3. **文件存储**：考虑使用对象存储（MinIO/S3）
4. **监控**：配置日志监控和告警
5. **备份**：定期备份数据库和文件存储

---

**最后更新**: 2026-01-20
