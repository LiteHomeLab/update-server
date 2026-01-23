# Update Server 使用指南

Update Server 是一个通用的应用程序自动更新服务器，支持多程序管理、Web 管理界面、端到端加密等功能。

## 快速开始

### 一、部署服务器

1. **编译所有组件**

   运行一键编译脚本：
   ```bash
   build-all.bat
   ```

   这将编译：
   - `bin/update-server.exe` - 服务器主程序
   - `bin/clients/update-admin.exe` - 发布客户端（用于上传版本）

2. **服务器部署目录结构**

   ```
   update-server/
   ├── update-server.exe          # 服务器主程序
   ├── config.yaml                 # 服务器配置文件（自动生成）
   ├── data/
   │   ├── versions.db            # 数据库文件（自动生成）
   │   ├── packages/              # 版本包存储目录（自动生成）
   │   │   ├── stable/            # 稳定版
   │   │   └── beta/              # 测试版
   │   └── clients/               # 客户端工具目录（需手动创建）
   │       ├── publish-client.exe # 发布客户端（从 bin/clients/ 复制）
   │       └── update-client.exe  # 更新客户端（可选，暂未实现）
   └── logs/                      # 日志目录（自动生成）
   ```

   **重要**：
   - 运行服务器前，确保创建 `data/clients/` 目录
   - 将编译好的 `bin/clients/update-admin.exe` 复制到 `data/clients/publish-client.exe`

3. **运行服务器**
   ```bash
   ./bin/update-server.exe
   ```

4. **完成初始化向导**
   - 浏览器会自动打开 `http://localhost:8080/setup`
   - 按照向导完成配置：
     - 步骤1：基本配置（服务器名称、端口、数据目录）
     - 步骤2：管理员账号（用户名、密码、服务器URL）
     - 步骤3：确认配置
   - 完成后自动进入管理后台

### 二、创建程序

1. 登录管理后台
   ```
   访问 http://localhost:8080
   使用初始化时设置的管理员账号登录
   ```

2. 创建新程序
   - 点击「程序管理」→「创建新程序」
   - 填写程序信息：
     - Program ID：唯一标识符（如：`docufiller`、`myapp`）
     - 名称：程序显示名称
     - 描述：程序说明（可选）
   - 点击创建，系统会自动生成：
     - Upload Token（上传Token）
     - Download Token（下载Token）
     - Encryption Key（加密密钥）

3. 记录重要信息
   - 在程序详情页可以查看所有生成的密钥和Token
   - 这些信息将在配置客户端工具时使用

### 三、发布更新

1. **下载发布客户端**
   - 在程序详情页点击「下载发布端」
   - 解压 `程序名-client-publish.zip`
   - 包含：
     - `update-admin.exe` - 发布客户端命令行工具
     - `config.yaml` - 预配置的配置文件（已包含 Token 和密钥）
     - `README.md` - 使用说明

2. **配置发布客户端**

   配置文件 `config.yaml` 已自动生成，包含必要的服务器地址、Token 和密钥。
   如需修改，可编辑配置文件：

   ```yaml
   server:
     url: "http://your-server:8080"
     timeout: 30

   program:
     id: "docufiller"

   auth:
     token: "ul_xxxxxxxxxxxxx"       # Upload Token（已预配置）
     encryption_key: "base64..."     # Encryption Key（已预配置）
   ```

3. **执行发布**

   使用命令行上传版本：

   ```bash
   update-admin.exe upload --channel stable --version 1.0.0 --file ./app.zip --notes "发布说明"
   ```

   或使用配置文件：

   ```bash
   update-admin.exe --server http://your-server:8080 --token ul_xxx upload --program-id docufiller --channel stable --version 1.0.0 --file ./app.zip
   ```

   支持的命令：
   - `upload` - 上传新版本
   - `list` - 列出版本
   - `delete` - 删除版本

4. **验证发布**
   - 在Web管理后台的「版本列表」中查看新版本
   - 版本信息包括：版本号、发布时间、下载次数、文件大小等

### 四、集成更新功能

1. **下载更新客户端**
   - 在程序详情页点击「下载更新端」
   - 解压 `程序名-client-update.zip`
   - 包含：
     - `update-client.exe` - 更新客户端工具（可选，如已实现）
     - `config.yaml` - 预配置的配置文件
     - `README.md` - 使用说明

   **注意**：当前版本更新客户端为可选组件，如未实现，下载的包将仅包含配置文件。

2. **配置更新客户端**

   配置文件 `config.yaml` 已自动生成：

   ```yaml
   server:
     url: "http://your-server:8080"
     timeout: 30

   program:
     id: "docufiller"

   auth:
     token: "dl_xxxxxxxxxxxxx"       # Download Token（已预配置）
     encryption_key: "base64..."     # Encryption Key（已预配置）
   ```

3. **检查更新**
   ```bash
   update-client.exe --check
   ```

4. **处理更新结果**

   返回JSON格式：
   ```json
   {
     "hasUpdate": true,
     "latestVersion": "1.2.0",
     "downloadUrl": "http://server:8080/api/download/...",
     "fileSize": 52428800,
     "changelog": "更新说明",
     "mandatory": false
   }
   ```

   如果 `hasUpdate` 为 `true`，则：
   - 提示用户有新版本
   - 显示更新内容
   - 询问是否下载更新

5. **下载并安装更新**
   ```bash
   # 自动下载（配置了 autoDownload: true）
   update-client.exe --check
   # 下载完成后文件在 ./updates 目录

   # 手动下载
   update-client.exe --download
   ```

## Web 管理后台功能

### 仪表盘
- 系统概览统计
- 最近活动记录

### 程序管理
- 查看所有程序列表
- 创建新程序
- 查看程序详情
- 删除程序

### 程序详情页
- **基本信息**：Program ID、创建时间、下载统计
- **Token管理**：查看/重新生成 Upload Token、Download Token
- **加密密钥**：查看/重新生成 Encryption Key
- **版本列表**：查看所有版本、下载统计、删除版本
- **客户端工具**：一键下载配置好的发布端和更新端

### 系统设置
- 修改服务器配置
- 查看系统日志

## 常见问题

### Q: 如何修改服务器端口？
A: 在初始化向导中设置，或编辑 `config.yaml` 中的 `server.port` 配置项，然后重启服务器。

### Q: 如何配置客户端访问的服务器地址？
A: 编辑 `config.yaml` 中的 `serverUrl` 配置项，设置为服务器对外暴露的完整 URL（例如：`http://192.168.1.100:8083`）。客户端下载的配置包将使用此地址。

### Q: 忘记管理员密码怎么办？
A: 停止服务器，删除 `data/versions.db` 数据库文件，重新启动服务器并完成初始化向导。

### Q: Token 或密钥泄露怎么办？
A: 在程序详情页点击「重新生成」按钮，然后重新下载客户端工具配置。

### Q: 如何备份数据？
A: 定期备份以下内容：
- `data/versions.db` 数据库文件
- `data/packages/` 存储的文件
- `config.yaml` 配置文件

### Q: 支持多台服务器部署吗？
A: 支持，但需要共享数据库和文件存储，或使用负载均衡器。

## 技术支持

- GitHub Issues: https://github.com/LiteHomeLab/update-server/issues
- 架构说明: [ARCHITECTURE.md](ARCHITECTURE.md)

---

**最后更新**: 2026-01-20
