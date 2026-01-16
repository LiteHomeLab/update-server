# Windows 发布工具设计文档

## 概述

一个基于 Go 的命令行发布工具，编译为独立的 Windows 可执行文件 `update-admin.exe`，用于向 DocuFiller 更新服务器上传和管理程序版本。

## 架构

### 模块结构

```
update-admin/
├── main.go          # 命令行接口（CLI）
├── admin.go         # 核心业务逻辑
├── config.go        # 配置管理
├── utils.go         # 工具函数（SHA256、进度）
└── go.mod           # 依赖管理
```

### 技术栈

- **框架**：Cobra（命令行参数解析）
- **HTTP**：net/http 标准库
- **加密**：crypto/sha256（文件校验）
- **编译**：Go 1.23+，支持交叉编译

## 核心功能

### 1. 上传命令（upload）

完整功能实现，包含以下步骤：

#### 参数验证
- 检查必需参数：`--channel`、`--version`、`--file`
- 验证文件存在性
- 检查文件大小（不超过服务器 maxFileSize 限制）

#### 文件预处理
- 计算本地文件 SHA256 哈希值

#### 上传过程
- 构造 multipart/form-data 请求
- 实时显示上传进度百分比
- 发送到 `/api/programs/{programId}/versions/upload`

#### 自动验证
- 上传成功后，下载文件验证可访问性
- 比对服务器返回的 SHA256 与本地哈希值

#### 结果报告
```
✓ 上传成功: myapp 1.0.0 stable
  文件: myapp.zip (15.2 MB)
  SHA256: a1b2c3d4...
  校验: 通过
```

### 2. 列表命令（list）

以表格形式展示版本信息：
```
Version    Channel    File Size    Publish Date
1.0.0      stable     15.2 MB      2026-01-16
0.9.0      stable     14.8 MB      2026-01-10
```

### 3. 删除命令（delete）

发送 DELETE 请求到 `/api/programs/{programId}/versions/{channel}/{version}`，显示确认信息。

## 配置管理

### 环境变量（全局配置）

```cmd
setx UPDATE_SERVER_URL "http://release-server:8080"
setx UPDATE_TOKEN "your-production-token"
```

- `UPDATE_SERVER_URL`：服务器地址
- `UPDATE_TOKEN`：API 认证令牌（Bearer Token）

### 命令行参数（每次指定或覆盖）

| 参数 | 说明 | 是否必需 |
|------|------|----------|
| `--program-id` | 程序标识符 | **必需** |
| `--server` | 服务器地址 | 可选（覆盖环境变量） |
| `--token` | 认证令牌 | 可选（覆盖环境变量） |
| `--channel` | 发布通道（stable/beta） | 上传/删除必需 |
| `--version` | 版本号 | 上传/删除必需 |
| `--file` | 文件路径 | 上传必需 |
| `--notes` | 发布说明 | 可选 |
| `--mandatory` | 强制更新标记 | 可选（默认 false） |

**注意**：`--program-id` 不使用环境变量，必须每次指定，因为同一台机器可能发布多个程序。

### 使用示例

```cmd
# 配置全局服务器地址和令牌
setx UPDATE_SERVER_URL "http://localhost:8080"
setx UPDATE_TOKEN "test-token-123"

# 发布不同程序
update-admin.exe upload --program-id myapp --channel stable --version 1.0.0 --file myapp.zip
update-admin.exe upload --program-id another-app --channel beta --version 2.0.0 --file another.zip

# 临时切换到测试服务器
update-admin.exe upload --server http://test-server:8080 --program-id myapp --channel beta --version 1.1.0 --file myapp.zip
```

## 编译和分发

### 编译命令

```bash
# 标准 Windows 64位编译
go build -ldflags "-s -w" -o update-admin.exe .

# 使用 UPX 压缩（可选）
upx --best --lzma update-admin.exe
```

### 分发方式

最终产物为单个 `update-admin.exe` 文件（约 3-5 MB），无需外部依赖。

用户可以：
- 直接复制到任何 Windows 机器使用
- 放在项目源码仓库的 `tools/` 目录
- 放入系统 PATH（如 `C:\Tools\`），全局使用

## 错误处理

所有命令遵循统一的错误处理原则：

1. 显示清晰的错误信息（包括 HTTP 状态码）
2. 以非零退出码退出，便于脚本集成
3. 常见错误提示解决方案

示例：
```
Error: upload failed with status 401 (Unauthorized)
Hint: Check your UPDATE_TOKEN environment variable or --token parameter
```

## 测试策略

### 单元测试

- SHA256 计算正确性
- 配置加载和优先级
- 进度回调触发

### 集成测试

1. 启动本地测试服务器
2. 执行完整流程：上传 → 验证 → 列表 → 删除
3. 测试错误场景（无效 token、文件过大等）

## 文档

### README.md

- 快速开始
- 编译步骤
- 基本使用示例

### USAGE.md

- 完整命令和参数说明
- 环境变量配置模板
- CI/CD 集成示例

### FAQ.md

- 上传失败排查
- SHA256 校验失败处理
- 多项目管理最佳实践

## API 依赖

工具与服务器 API 交互：

| 端点 | 方法 | 用途 |
|------|------|------|
| `/api/programs/{programId}/versions/upload` | POST | 上传版本 |
| `/api/programs/{programId}/versions?channel={channel}` | GET | 列出版本 |
| `/api/programs/{programId}/versions/{channel}/{version}` | DELETE | 删除版本 |
| `/api/download/{programId}/{channel}/{version}` | GET | 下载验证 |

## 安全考虑

1. **Token 保护**：避免在命令行历史中暴露 token（优先使用环境变量）
2. **HTTPS 支持**：服务器地址应使用 HTTPS（生产环境）
3. **文件验证**：上传后自动验证文件完整性和可访问性
