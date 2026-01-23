# 管理后台配置重构设计

**日期**: 2025-01-23
**作者**: Claude Code
**状态**: 设计完成

## 概述

本文档描述了移除服务器引导系统、将管理员配置迁移到 `config.yaml` 的设计方案。

### 当前问题

- 引导页面 (`/setup`) 增加了系统复杂度
- 管理员凭据存储在数据库中，不便于配置管理
- 初始化流程依赖数据库状态，违反了"配置即代码"原则

### 设计目标

1. 移除所有引导相关代码和页面
2. 管理员凭据通过 `config.yaml` 配置
3. 简化启动流程，无需初始化检查
4. 使用 Session Cookie 管理登录状态

## 配置文件结构

### config.yaml 新增配置段

```yaml
admin:
  username: "admin"
  password: "change-this-password-in-production"
```

### Config 结构体更新

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Storage  StorageConfig
    API      APIConfig
    Logger   LoggerConfig
    Crypto   CryptoConfig
    Admin    AdminConfig  // 新增
}

type AdminConfig struct {
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}
```

## 认证流程

### 登录验证

登录时直接对比配置文件中的用户名和密码，不再查询数据库：

```go
if req.Username != h.cfg.Admin.Username || req.Password != h.cfg.Admin.Password {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
    return
}
```

### Session 管理

使用 `gin-contrib/sessions/cookie` 存储 Session：

```go
store := cookie.NewStore([]byte(cfg.Crypto.MasterKey))
r.Use(sessions.Sessions("admin-session", store))
```

登录成功后设置 Session：

```go
session.Set("authenticated", true)
session.Set("username", req.Username)
session.Save()
```

## 路由简化

### 删除的路由

- `GET /setup` - 引导页面
- `POST /api/setup/initialize` - 初始化 API
- `GET /api/setup/status` - 初始化状态检查

### 简化后的根路由

```go
// 根路径直接重定向到管理后台
r.GET("/", func(c *gin.Context) {
    c.Redirect(http.StatusFound, "/admin")
})
```

### 认证中间件

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        if session.Get("authenticated") != true {
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 代码变更清单

### 修改的文件

| 文件 | 变更 |
|------|------|
| `config.yaml` | 添加 `admin` 配置段 |
| `internal/config/config.go` | 添加 `AdminConfig` 结构体 |
| `internal/handler/auth.go` | 重写为从配置验证 |
| `internal/handler/admin.go` | 更新 `AuthMiddleware()` 和构造函数 |
| `cmd/update-server/main.go` | 添加 Session 中间件，移除 Setup 路由 |
| `internal/database/database.go` | 移除 AdminUser 迁移 |

### 删除的文件

| 文件 | 说明 |
|------|------|
| `internal/service/setup.go` | SetupService |
| `internal/handler/setup.go` | SetupHandler |
| `internal/models/admin_user.go` | AdminUser 模型 |
| `web/setup.html` | 引导页面 |
| `cmd/create-admin/main.go` | 创建管理员工具 |

### 新增依赖

```bash
go get github.com/gin-contrib/sessions
go get github.com/gin-contrib/sessions/cookie
```

## 安全考虑

1. **明文密码**: config.yaml 中的密码为明文，需要通过文件权限保护
2. **Session 密钥**: 使用 `crypto.masterKey` 作为 Session 加密密钥
3. **生产环境**: 部署时必须修改默认密码

## 向后兼容性

- 所有公开 API (`/api/health`, `/api/programs/*`) 保持不变
- 客户端下载和上传 API 不受影响
- 仅影响管理后台的访问方式

## 测试建议

1. 修改 config.yaml 中的用户名密码，验证登录功能
2. 验证未登录时访问 `/admin` 重定向到登录页
3. 验证 Session 过期后需要重新登录
4. 确认删除 Setup 路由后不影响现有 API
