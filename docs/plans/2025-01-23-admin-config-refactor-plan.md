# 管理后台配置重构实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 移除服务器引导系统，将管理员凭据迁移到 config.yaml 配置文件，使用 Session Cookie 管理登录状态。

**Architecture:** 保留现有 Session 认证逻辑，将数据源从数据库改为配置文件。移除所有 Setup 相关代码和路由，简化启动流程。

**Tech Stack:** Go, Gin, gin-contrib/sessions, GORM, SQLite

---

## Task 1: 添加 Session 依赖

**Files:**
- Modify: `go.mod`

**Step 1: 添加 gin-contrib/sessions 依赖**

```bash
go get github.com/gin-contrib/sessions
go get github.com/gin-contrib/sessions/cookie
go mod tidy
```

**Step 2: 验证依赖安装成功**

Run: `go mod verify`
Expected: 无错误输出

**Step 3: 提交**

```bash
git add go.mod go.sum
git commit -m "deps: 添加 gin-contrib/sessions 依赖"
```

---

## Task 2: 更新配置结构体

**Files:**
- Modify: `internal/config/config.go`

**Step 1: 读取当前配置文件**

```bash
cat internal/config/config.go
```

**Step 2: 添加 AdminConfig 结构体**

在 `internal/config/config.go` 中找到 `Config` 结构体定义，添加 `Admin` 字段：

```go
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Storage  StorageConfig  `yaml:"storage"`
    API      APIConfig      `yaml:"api"`
    Logger   LoggerConfig   `yaml:"logger"`
    Crypto   CryptoConfig   `yaml:"crypto"`
    Admin    AdminConfig    `yaml:"admin"`  // 新增
}

// AdminConfig 管理员配置
type AdminConfig struct {
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}
```

**Step 3: 验证编译**

Run: `go build ./cmd/update-server`
Expected: 编译成功

**Step 4: 提交**

```bash
git add internal/config/config.go
git commit -m "feat(config): 添加 AdminConfig 结构体"
```

---

## Task 3: 更新 config.yaml

**Files:**
- Modify: `config.yaml`

**Step 1: 添加 admin 配置段**

在 `config.yaml` 文件末尾添加：

```yaml
admin:
  username: "admin"
  password: "change-this-password-in-production"
```

**Step 2: 验证配置加载**

Run: `go run cmd/update-server/main.go`
Expected: 服务器正常启动，无配置错误

**Step 3: 停止服务器**

按 Ctrl+C 停止服务器

**Step 4: 提交**

```bash
git add config.yaml
git commit -m "feat(config): 添加管理员凭据配置"
```

---

## Task 4: 重写 AuthHandler

**Files:**
- Modify: `internal/handler/auth.go`

**Step 1: 备份原文件**

```bash
cp internal/handler/auth.go internal/handler/auth.go.bak
```

**Step 2: 完全重写 AuthHandler**

将 `internal/handler/auth.go` 内容替换为：

```go
package handler

import (
    "net/http"
    "docufiller-update-server/internal/config"

    "github.com/gin-contrib/sessions"
    "github.com/gin-gonic/gin"
)

type AuthHandler struct {
    cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
    return &AuthHandler{cfg: cfg}
}

// Login 处理登录请求
func (h *AuthHandler) Login(c *gin.Context) {
    var req struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
        return
    }

    // 从配置验证凭据
    if req.Username != h.cfg.Admin.Username || req.Password != h.cfg.Admin.Password {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
        return
    }

    // 设置 Session
    session := sessions.Default(c)
    session.Set("authenticated", true)
    session.Set("username", req.Username)
    if err := session.Save(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}

// Logout 处理登出请求
func (h *AuthHandler) Logout(c *gin.Context) {
    session := sessions.Default(c)
    session.Clear()
    session.Save()
    c.JSON(http.StatusOK, gin.H{"success": true})
}
```

**Step 3: 验证编译**

Run: `go build ./cmd/update-server`
Expected: 编译成功

**Step 4: 删除备份文件**

```bash
rm internal/handler/auth.go.bak
```

**Step 5: 提交**

```bash
git add internal/handler/auth.go
git commit -m "refactor(auth): 从配置文件验证管理员凭据"
```

---

## Task 5: 更新 AdminHandler 和 AuthMiddleware

**Files:**
- Modify: `internal/handler/admin.go`

**Step 1: 更新 AuthMiddleware**

在 `internal/handler/admin.go` 中找到 `AuthMiddleware` 函数，替换为：

```go
// AuthMiddleware 管理后台认证中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)

        authenticated := session.Get("authenticated")
        if authenticated != true {
            // 未登录，返回 401 或重定向到登录页
            if c.Request.Header.Get("Content-Type") == "application/json" {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
            } else {
                c.Redirect(http.StatusFound, "/admin/login")
            }
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**Step 2: 更新 AdminHandler 构造函数**

在 `internal/handler/admin.go` 中找到 `NewAdminHandler` 函数，移除 `setupService` 参数：

```go
func NewAdminHandler(
    programService *service.ProgramService,
    versionService *service.VersionService,
    tokenService *service.TokenService,
    clientPackagerService *service.ClientPackager,
) *AdminHandler {
    return &AdminHandler{
        programService:      programService,
        versionService:     versionService,
        tokenService:       tokenService,
        clientPackagerService: clientPackagerService,
    }
}
```

**Step 3: 移除 AdminHandler 结构体中的 setupService 字段**

在 `internal/handler/admin.go` 中找到 `AdminHandler` 结构体定义，移除 `setupService` 字段：

```go
type AdminHandler struct {
    programService      *service.ProgramService
    versionService     *service.VersionService
    tokenService       *service.TokenService
    clientPackagerService *service.ClientPackager
}
```

**Step 4: 验证编译**

Run: `go build ./cmd/update-server`
Expected: 编译成功

**Step 5: 提交**

```bash
git add internal/handler/admin.go
git commit -m "refactor(admin): 更新 AuthMiddleware 使用 Session，移除 SetupService 依赖"
```

---

## Task 6: 更新 main.go 路由

**Files:**
- Modify: `cmd/update-server/main.go`

**Step 1: 添加 Session 中间件导入**

在 `cmd/update-server/main.go` 的 import 部分添加：

```go
import (
    "fmt"
    "net/http"
    "strings"

    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    "github.com/gin-gonic/gin"
    // ... 其他导入保持不变 ...
)
```

**Step 2: 添加 Session 中间件**

在 `cmd/update-server/main.go` 中找到 `r := gin.Default()` 之后，添加：

```go
    // 添加 Session 中间件
    store := cookie.NewStore([]byte(cfg.Crypto.MasterKey))
    r.Use(sessions.Sessions("admin-session", store))
```

**Step 3: 更新 AuthHandler 初始化**

找到 `authHandler := handler.NewAuthHandler(db)`，替换为：

```go
authHandler := handler.NewAuthHandler(cfg)
```

**Step 4: 移除 SetupService 初始化**

找到 `setupService := service.NewSetupService(db)`，删除这行。

**Step 5: 移除 SetupHandler 初始化**

找到 `setupHandler := handler.NewSetupHandler(setupService)`，删除这行。

**Step 6: 更新 AdminHandler 初始化**

找到 `adminHandler := handler.NewAdminHandler(...)`，移除 `setupService` 参数：

```go
adminHandler := handler.NewAdminHandler(
    programService,
    versionService,
    tokenSvc,
    clientPackagerService,
)
```

**Step 7: 移除初始化检查和条件路由**

找到 `initialized, err := setupService.IsInitialized()` 及相关的 `if !initialized` 块，删除整个条件块。替换为简单的根路由：

```go
    // 根路径直接重定向到管理后台
    r.GET("/", func(c *gin.Context) {
        c.Redirect(http.StatusFound, "/admin")
    })
```

**Step 8: 删除 Setup 路由组**

找到 `setupGroup := r.Group("/setup")` 整个块，删除。

**Step 9: 简化 Admin 路由组**

找到 `adminGroup := r.Group("/admin")`，移除条件中间件，始终添加认证：

```go
    // Admin 页面路由 - 需要认证
    adminGroup := r.Group("/admin")
    adminGroup.Use(handler.AuthMiddleware())
    {
        adminGroup.GET("", func(c *gin.Context) {
            c.HTML(http.StatusOK, "admin.html", gin.H{
                "title": "管理后台",
            })
        })

        adminGroup.POST("/logout", authHandler.Logout)
    }
```

**Step 10: 删除 Setup API 路由组**

找到 `setupAPI := r.Group("/api/setup")` 整个块，删除。

**Step 11: 移除 adminGroup 内的动态初始化检查**

在 `adminGroup.GET("")` 中，移除 `isInitialized` 检查逻辑。

**Step 12: 验证编译**

Run: `go build ./cmd/update-server`
Expected: 编译成功

**Step 13: 测试启动**

Run: `go run cmd/update-server/main.go`
Expected: 服务器正常启动在 8083 端口

**Step 14: 停止服务器**

按 Ctrl+C 停止服务器

**Step 15: 提交**

```bash
git add cmd/update-server/main.go
git commit -m "refactor(main): 移除 Setup 路由，添加 Session 中间件，简化启动流程"
```

---

## Task 7: 移除数据库迁移中的 AdminUser

**Files:**
- Modify: `internal/database/database.go`

**Step 1: 读取数据库迁移文件**

```bash
cat internal/database/database.go
```

**Step 2: 移除 AdminUser 迁移**

在 `AutoMigrate` 函数中，移除 `&models.AdminUser{}`：

```go
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.Program{},
        &models.Version{},
        &models.Token{},
    )
}
```

**Step 3: 验证编译**

Run: `go build ./cmd/update-server`
Expected: 编译成功

**Step 4: 提交**

```bash
git add internal/database/database.go
git commit -m "refactor(database): 移除 AdminUser 表迁移"
```

---

## Task 8: 删除 Setup 相关文件

**Files:**
- Delete: `internal/service/setup.go`
- Delete: `internal/handler/setup.go`
- Delete: `internal/models/admin_user.go`
- Delete: `web/setup.html`
- Delete: `cmd/create-admin/main.go`

**Step 1: 删除 SetupService**

```bash
rm internal/service/setup.go
```

**Step 2: 删除 SetupHandler**

```bash
rm internal/handler/setup.go
```

**Step 3: 删除 AdminUser 模型**

```bash
rm internal/models/admin_user.go
```

**Step 4: 删除 Setup 页面**

```bash
rm web/setup.html
```

**Step 5: 删除创建管理员工具**

```bash
rm cmd/create-admin/main.go
```

**Step 6: 验证编译**

Run: `go build ./cmd/update-server`
Expected: 编译成功

**Step 7: 提交**

```bash
git add -A
git commit -m "refactor: 删除所有 Setup 相关文件和代码"
```

---

## Task 9: 手动测试

**Step 1: 启动服务器**

```bash
go run cmd/update-server/main.go
```

Expected: 服务器正常启动，无 Setup 相关日志

**Step 2: 测试根路径重定向**

```bash
curl -I http://localhost:8083/
```

Expected: `HTTP/1.1 302 Found` 和 `Location: /admin`

**Step 3: 测试未登录访问管理后台 API**

```bash
curl http://localhost:8083/api/admin/programs
```

Expected: `{"error":"未登录"}` 或重定向到登录页

**Step 4: 测试登录**

```bash
curl -X POST http://localhost:8083/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"change-this-password-in-production"}' \
  -c cookies.txt
```

Expected: `{"success":true}`

**Step 5: 测试登录后访问 API**

```bash
curl http://localhost:8083/api/admin/programs -b cookies.txt
```

Expected: 返回程序列表（可能为空）

**Step 6: 测试错误密码**

```bash
curl -X POST http://localhost:8083/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"wrong"}'
```

Expected: `{"error":"用户名或密码错误"}`

**Step 7: 测试健康检查 API（无需认证）**

```bash
curl http://localhost:8083/api/health
```

Expected: `{"status":"ok"}`

**Step 8: 停止服务器**

按 Ctrl+C 停止服务器

**Step 9: 清理测试文件**

```bash
rm cookies.txt
```

---

## Task 10: 更新 web/admin.js 中的 API 调用

**Files:**
- Modify: `web/admin.js`

**Step 1: 检查 admin.js 是否需要更新**

```bash
cat web/admin.js | grep -i setup
```

**Step 2: 如果有 Setup 相关代码，删除它们**

如果没有 setup 相关内容，跳过此步骤

**Step 3: 验证**

检查 admin.js 中的 API 调用路径是否正确

**Step 4: 提交（如果有更改）**

```bash
git add web/admin.js
git commit -m "refactor(web): 移除 Setup 相关代码"
```

---

## Task 11: 清理数据库（可选）

**注意**: 此步骤会删除现有数据库中的 admin_users 表，请先备份数据库

**Step 1: 备份数据库（可选）**

```bash
cp data/versions.db data/versions.db.backup
```

**Step 2: 删除数据库文件，让系统重建**

```bash
rm data/versions.db
```

**Step 3: 启动服务器，自动创建新数据库**

```bash
go run cmd/update-server/main.go
```

Expected: 服务器正常启动，自动创建不包含 admin_users 表的数据库

**Step 4: 停止服务器**

按 Ctrl+C 停止服务器

**Step 5: 验证数据库表结构**

```bash
sqlite3 data/versions.db ".schema"
```

Expected: 只有 `programs`, `versions`, `tokens` 表，没有 `admin_users` 表

**Step 6: 提交（如果需要）**

数据库文件通常不提交到 git

---

## Task 12: 更新文档

**Files:**
- Modify: `CLAUDE.md`

**Step 1: 更新 CLAUDE.md 中的配置说明**

在 CLAUDE.md 的配置文件部分添加 admin 配置说明：

```yaml
admin:
  username: "admin"              # 管理后台用户名
  password: "change-this-password-in-production"  # 管理后台密码（生产环境必须修改）
```

**Step 2: 更新 API 端点参考**

移除 `/setup` 和 `/api/setup/*` 相关的端点说明

**Step 3: 提交**

```bash
git add CLAUDE.md
git commit -m "docs: 更新配置说明和 API 文档"
```

---

## Task 13: 最终验证

**Step 1: 完整构建测试**

```bash
make build
```

Expected: 生成 `./bin/docufiller-update-server.exe`

**Step 2: 运行构建的二进制文件**

```bash
./bin/docufiller-update-server.exe
```

Expected: 服务器正常启动

**Step 3: 浏览器测试**

打开浏览器访问 `http://localhost:8083`，应该自动重定向到登录页面

**Step 4: 登录测试**

使用 config.yaml 中的用户名密码登录

**Step 5: 停止服务器**

按 Ctrl+C 停止服务器

**Step 6: 最终提交**

```bash
git add -A
git commit -m "chore: 完成管理后台配置重构"
```

---

## 验收标准

完成所有任务后，应该满足以下条件：

1. ✅ `config.yaml` 中包含 `admin.username` 和 `admin.password` 配置
2. ✅ 服务器启动时无需初始化检查
3. ✅ 访问根路径直接重定向到 `/admin`
4. ✅ 未登录访问 `/admin` 重定向到 `/admin/login`
5. ✅ 使用配置文件中的用户名密码可以登录
6. ✅ 登录后可以访问管理后台 API
7. ✅ 数据库中没有 `admin_users` 表
8. ✅ 所有 Setup 相关代码已删除
9. ✅ 公开 API (`/api/health`, `/api/programs/*`) 正常工作
10. ✅ 服务器可以正常构建和运行

---

## 回滚计划

如果出现问题，可以使用以下命令回滚：

```bash
git log --oneline  # 查看提交历史
git reset --hard <commit-before-refactor>  # 回滚到重构之前的提交
```

建议保留重构前的提交 ID 作为备份。
