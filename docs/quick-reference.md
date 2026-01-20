# 密钥与 Token 系统快速参考

## 一张图看懂所有密钥和 Token

```mermaid
graph TB
    subgraph ConfigLayer["【配置文件层】config.yaml"]
        MK["✅ crypto.masterKey<br/>用途: 文件加密<br/>状态: 可选功能<br/>默认: 32字节密钥"]
        UT["❌ api.uploadToken<br/>状态: 已弃用!<br/>操作: 忽略即可"]
    end

    subgraph CryptoService["🔐 加密服务 (CryptoService)"]
        HKDF["MasterKey ──HKDF──▶ 程序专用密钥 ──▶ 加密/解密文件<br/>注: 这是可选功能，用于保护存储的文件"]
    end

    subgraph TokenGen["【数据库 Token 层】gen-token 工具 ← 这是实际用于认证的！"]
        GT["go run cmd/gen-token/main.go"]
        Step1["1. 生成 64 位随机 Token<br/>db2d387ff07aed70562da78115a45edd..."]
        Step2["2. 计算 SHA256 Token → Token ID<br/>1022d95b8439843d2e385fa56b7b3ec..."]
        Step3["3. 存入数据库 只存哈希，不存原值"]
        TokenOut["返回: db2d387ff07aed70562da78115a45edd2821740ebd3233e9dac4cb163eec67cc"]
    end

    subgraph Database["📊 数据库 versions.db"]
        DB["tokens 表:<br/>token_id | token值 | token_type | program_id | 使用场景<br/>────────────────────────────────────────<br/>哈希值 | 不存储 | admin | NULL | 👑 管理所有程序<br/>哈希值 | 不存储 | upload | docufiller | 📤 上传版本<br/>哈希值 | 不存储 | download | docufiller | 📥 下载文件<br/><br/>⚠️ 安全: 数据库中只存储 Token 的 SHA256 哈希，不存储原始 Token！"]
    end

    subgraph ClientConfig["【使用层】release-config.bat 发布配置"]
        RC["UPDATE_SERVER_URL=http://172.18.200.47:58100<br/>UPDATE_TOKEN=db2d387ff07aed70562da78115a45edd2821740ebd3233e9dac4cb163eec67cc<br/>↑ 这是从 gen-token 生成的 Admin Token"]
    end

    subgraph PublishFlow["🚀 发布流程"]
        Release["release.bat"]
        UploadAdmin["upload-admin"]
        API["POST /api/programs/docufiller/versions<br/><br/>请求头:<br/>Authorization: Bearer db2d387f...<br/><br/>服务器处理:<br/>1. 提取 Token<br/>2. 计算 SHA256 Token<br/>3. 在数据库中查找匹配的哈希<br/>4. 检查 Token 类型和权限<br/>5. 允许或拒绝请求"]
    end

    subgraph DownloadFlow["【客户端下载层】DocuFiller 客户端"]
        DL["1. 请求: GET /api/programs/docufiller/versions/latest?channel=stable 公开接口，无需 Token<br/><br/>2. 请求: GET /api/programs/docufiller/download/stable/1.0.0 需要 Download Token<br/><br/>3. 下载并验证文件 SHA256"]
    end

    MK --> CryptoService
    GT --> Step1
    Step1 --> Step2
    Step2 --> Step3
    Step3 --> TokenOut
    TokenOut --> Database
    TokenOut --> ClientConfig
    ClientConfig --> Release
    Release -->|"使用 Token<br/>HTTP Authorization<br/>Bearer <TOKEN>"| UploadAdmin
    UploadAdmin --> API

    classDef deprecated fill:#f99,stroke:#f00,stroke-width:2px,color:#000
    classDef active fill:#9f9,stroke:#090,stroke-width:2px,color:#000
    classDef crypto fill:#ff9,stroke:#990,stroke-width:2px,color:#000
    classDef db fill:#99f,stroke:#009,stroke-width:2px,color:#000

    class UT deprecated
    class MK,GT,Step1,Step2,Step3,TokenOut,RC,Release,UploadAdmin,DL active
    class CryptoService,HKDF crypto
    class DB db
```

---

## 三步配置快速指南

### 第 1 步：生成 Admin Token（只需执行一次）

```bash
# 在服务器目录执行
cd C:\WorkSpace\Go2Hell\src\github.com\LiteHomeLab\update-server
go run cmd/gen-token/main.go

# 输出示例:
# Admin Token: db2d387ff07aed70562da78115a45edd2821740ebd3233e9dac4cb163eec67cc
#              ↑ 复制这个完整的 Token 字符串
```

### 第 2 步：配置客户端（开发者机器）

```bat
# 编辑 DocuFiller 项目中的
# scripts\config\release-config.bat

set UPDATE_SERVER_URL=http://172.18.200.47:58100
set UPDATE_TOKEN=db2d387ff07aed70562da78115a45edd2821740ebd3233e9dac4cb163eec67cc
set UPLOAD_ADMIN_PATH=C:\WorkSpace\Go2Hell\src\github.com\LiteHomeLab\update-server\bin\upload-admin.exe
```

### 第 3 步：发布版本

```bash
# 1. 更新 DocuFiller.csproj 版本号
# 2. 创建 Git 标签
git tag v1.0.0
git push origin v1.0.0

# 3. 执行发布
scripts\release.bat
```

---

## 密钥/Token 一览表

| 名称 | 位置 | 用途 | 状态 | 是否必需 |
|------|------|------|------|----------|
| **MasterKey** | config.yaml → crypto.masterKey | 文件加密 | 使用中 | ❌ 可选 |
| **uploadToken** | config.yaml → api.uploadToken | ~~认证~~ | ❌ 已弃用 | ❌ 不使用 |
| **Admin Token** | 数据库 | 管理所有操作 | 使用中 | ✅ 必需 |
| **Upload Token** | 数据库 | 上传特定程序 | 使用中 | ⚠️ 可选 |
| **Download Token** | 数据库 | 下载特定程序 | 使用中 | ⚠️ 可选 |

---

## 认证流程简化图

```mermaid
sequenceDiagram
    participant C as 客户端
    participant M as AuthMiddleware
    participant DB as 数据库

    C->>M: Bearer: db2d387f...<br/>HTTP Request
    M->>M: 提取 Token
    M->>M: 计算 SHA256(Token)<br/>= 1022d95b8439...
    M->>DB: 查询数据库:<br/>SELECT * FROM tokens<br/>WHERE token_id = '1022d95b...'<br/>AND is_active = true

    alt 找到匹配的 Token
        DB-->>M: 返回 Token 信息
        M->>M: 检查权限
        M-->>C: ✅ 允许访问
    else 未找到或未激活
        DB-->>M: 无结果
        M-->>C: ❌ 401 Unauthorized
    end
```

---

**文档版本**：1.0
**最后更新**：2026-01-20
