# Update Server 使用指南

Update Server 是一个通用的应用程序自动更新服务器，支持多程序管理、Web 管理界面、端到端加密等功能。

## 快速开始

### 一、部署服务器

1. **下载并解压**
   ```
   解压 update-server.zip 到服务器目录
   ```

2. **运行服务器**
   ```bash
   ./update-server.exe
   ```

3. **完成初始化向导**
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
   - 解压 `程序名-publish-client.zip`
   - 包含：`publish-client.exe` + `publish-config.yaml` + `README.md`

2. **配置发布客户端**

   编辑 `publish-config.yaml`：
   ```yaml
   server: "http://your-server:8080"        # 服务器地址
   programId: "docufiller"                  # 程序ID
   uploadToken: "ul_xxxxxxxxxxxxx"          # Upload Token
   encryption:
     enabled: true
     key: "base64编码的密钥"                # Encryption Key
   file: "./app-v1.0.0.zip"                 # 要发布的文件
   version: "1.0.0"                         # 版本号
   channel: "stable"                        # stable 或 beta
   changelog: "更新说明"                    # 发布说明
   ```

3. **执行发布**
   ```bash
   publish-client.exe
   ```

4. **验证发布**
   - 在Web管理后台的「版本列表」中查看新版本
   - 版本信息包括：版本号、发布时间、下载次数、文件大小等

### 四、集成更新功能

1. **下载更新客户端**
   - 在程序详情页点击「下载更新端」
   - 解压 `程序名-update-client.zip`
   - 包含：`update-client.exe` + `update-config.yaml` + `README.md`

2. **配置更新客户端**

   编辑 `update-config.yaml`：
   ```yaml
   server: "http://your-server:8080"        # 服务器地址
   programId: "docufiller"                  # 程序ID
   downloadToken: "dl_xxxxxxxxxxxxx"        # Download Token
   encryption:
     enabled: true
     key: "base64编码的密钥"                # Encryption Key
   check:
     channel: "stable"                       # 检查的渠道
     autoDownload: true                      # 是否自动下载
   download:
     outputPath: "./updates"                 # 下载目录
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
