# Update Server 自动化端到端测试方案设计

**日期**: 2026-01-23
**作者**: Claude Code
**状态**: 设计草稿

## 1. 概述

本文档描述了 DocuFiller Update Server 的完整自动化端到端（E2E）测试方案，覆盖从服务器首次运行引导到客户端包发布的完整流程。

### 1.1 测试目标

- 验证首次运行初始化流程的正确性
- 验证管理员登录和会话管理
- 验证程序创建和配置功能
- 验证版本发布、查询、下载的完整流程
- 验证客户端打包和分发的功能
- 验证服务器发布包的完整性和可部署性

### 1.2 技术方案

采用**混合测试方式**：

| 测试类型 | 工具 | 用途 |
|---------|------|------|
| Web UI 测试 | Playwright | 管理后台交互、初始化向导 |
| API 测试 | Go testing + httptest | REST API 验证 |
| 集成测试 | Go testing | 完整业务流程验证 |
| 文件系统测试 | Go testing | 文件存储、打包验证 |

## 2. 测试架构

### 2.1 测试目录结构

```
update-server/
├── tests/
│   ├── e2e/                    # Playwright E2E 测试
│   │   ├── setup_test.go       # 初始化流程测试
│   │   ├── admin_test.go       # 管理后台测试
│   │   ├── program_test.go     # 程序管理测试
│   │   └── e2e_test.go         # 完整流程测试
│   ├── integration/            # API 集成测试
│   │   ├── api_test.go         # API 端点测试
│   │   ├── client_test.go      # 客户端打包测试
│   │   └── package_test.go     # 发布包测试
│   └── helpers/                # 测试辅助工具
│       ├── server.go           # 测试服务器管理
│       ├── fixtures.go         # 测试数据
│       └── assertions.go       # 自定义断言
├── scripts/
│   └── test-e2e.bat            # E2E 测试运行脚本
└── docs/
    └── plans/
        └── 2026-01-23-automated-e2e-test-design.md
```

### 2.2 测试环境配置

```yaml
# tests/config/test-config.yaml
test:
  server:
    port: 18080                # 测试服务器端口
    host: "127.0.0.1"
  database:
    path: ":memory:"           # 使用内存数据库
  storage:
    basePath: "./tmp/test-packages"
  timeout:
    serverStartup: 10s         # 服务器启动超时
    request: 30s               # API 请求超时
  playwright:
    headless: true
    slowMo: 50ms
    screenshotDir: "./tmp/screenshots"
```

## 3. 测试场景设计

### 3.1 首次运行引导流程测试

**测试文件**: `tests/e2e/setup_test.go`

#### 场景描述

当服务器首次启动（数据库为空）时，系统应引导用户完成初始化。

#### 测试步骤

| 步骤 | 操作 | 验证点 |
|-----|------|--------|
| 1 | 启动测试服务器 | 服务器正常启动，监听测试端口 |
| 2 | 访问 `http://localhost:18080/` | 自动重定向到 `/setup` |
| 3 | 检查 Setup 页面元素 | 存在初始化表单、用户名输入框、密码输入框 |
| 4 | 填写管理员信息 | 用户名: admin, 密码: Test@123 |
| 5 | 提交表单 | 调用 `POST /api/setup/initialize` |
| 6 | 验证响应 | 返回 200，success: true |
| 7 | 验证数据库 | `admin_users` 表存在管理员记录 |
| 8 | 验证初始化状态 | `initialized` 标记为 true |
| 9 | 再次访问 `/` | 重定向到 `/admin/login` |
| 10 | 访问 `/setup` | 返回 403 Forbidden |

#### 边界条件测试

| 场景 | 输入 | 预期结果 |
|-----|------|----------|
| 密码不匹配 | 密码和确认密码不同 | 表单验证错误 |
| 弱密码 | 密码少于 8 位 | 密码强度提示 |
| 重复初始化 | 已初始化后访问 setup | 403 Forbidden |
| 用户名已存在 | 尝试创建同名管理员 | 错误提示 |

#### 代码框架

```go
// tests/e2e/setup_test.go
package e2e

import (
    "testing"
    "github.com/playwright-community/playwright-go"
)

func TestFirstRunSetupFlow(t *testing.T) {
    // Setup: 启动测试服务器
    srv := setupTestServer(t)
    defer srv.Close()

    // 启动 Playwright
    pw, err := playwright.Run()
    if err != nil {
        t.Fatalf("could not start playwright: %v", err)
    }
    defer pw.Stop()

    browser, err := pw.Chromium.Launch(playwright.BrowserLaunchOptions{
        Headless: playwright.Bool(true),
    })
    if err != nil {
        t.Fatalf("could not launch browser: %v", err)
    }
    defer browser.Close()

    context, err := browser.NewContext()
    if err != nil {
        t.Fatalf("could not create context: %v", err)
    }
    defer context.Close()

    page, err := context.NewPage()
    if err != nil {
        t.Fatalf("could not create page: %v", err)
    }

    // 测试步骤 1: 访问根路径，验证重定向到 setup
    _, err = page.Goto(srv.URL + "/")
    if err != nil {
        t.Fatalf("could not goto: %v", err)
    }

    if page.URL() != srv.URL+"/setup" {
        t.Errorf("expected redirect to /setup, got %s", page.URL())
    }

    // 测试步骤 2: 填写管理员表单
    err = page.Fill("#username", "testadmin")
    if err != nil {
        t.Fatalf("could not fill username: %v", err)
    }

    err = page.Fill("#password", "Test@123")
    if err != nil {
        t.Fatalf("could not fill password: %v", err)
    }

    err = page.Fill("#confirmPassword", "Test@123")
    if err != nil {
        t.Fatalf("could not fill confirm password: %v", err)
    }

    // 测试步骤 3: 提交表单
    err = page.Click("#submit-btn")
    if err != nil {
        t.Fatalf("could not click submit: %v", err)
    }

    // 测试步骤 4: 验证成功后重定向到登录页
    page.WaitForURL(srv.URL + "/admin/login")

    // 测试步骤 5: 验证可以使用新账户登录
    err = page.Fill("#login-username", "testadmin")
    if err != nil {
        t.Fatalf("could not fill login username: %v", err)
    }

    err = page.Fill("#login-password", "Test@123")
    if err != nil {
        t.Fatalf("could not fill login password: %v", err)
    }

    err = page.Click("#login-btn")
    if err != nil {
        t.Fatalf("could not click login: %v", err)
    }

    // 验证登录成功，重定向到管理后台
    page.WaitForURL(srv.URL + "/admin")
}
```

### 3.2 管理员登录测试

**测试文件**: `tests/e2e/admin_test.go`

#### 场景描述

验证管理员登录、会话管理和登出功能。

#### 测试步骤

| 步骤 | 操作 | 验证点 |
|-----|------|--------|
| 1 | 访问 `/admin/login` | 显示登录表单 |
| 2 | 输入错误密码 | 显示错误消息 |
| 3 | 输入正确凭据 | 重定向到 `/admin` |
| 4 | 检查 session cookie | cookie 已设置 |
| 5 | 刷新页面 | 保持登录状态 |
| 6 | 访问需要认证的 API | 返回 200 |
| 7 | 点击登出 | 清除 session，重定向到登录页 |
| 8 | 访问受保护页面 | 重定向到登录页 |

#### API 测试

```go
// tests/integration/api_test.go
func TestAdminLoginAPI(t *testing.T) {
    srv := setupTestServerWithAdmin(t, "admin", "password123")
    defer srv.Close()

    tests := []struct {
        name       string
        username   string
        password   string
        wantStatus int
        wantToken  bool
    }{
        {
            name:       "valid credentials",
            username:   "admin",
            password:   "password123",
            wantStatus: http.StatusOK,
            wantToken:  true,
        },
        {
            name:       "invalid username",
            username:   "nonexistent",
            password:   "password123",
            wantStatus: http.StatusUnauthorized,
            wantToken:  false,
        },
        {
            name:       "invalid password",
            username:   "admin",
            password:   "wrongpassword",
            wantStatus: http.StatusUnauthorized,
            wantToken:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            body := fmt.Sprintf(`{"username":"%s","password":"%s"}`, tt.username, tt.password)
            req := httptest.NewRequest("POST", srv.URL+"/api/admin/login", strings.NewReader(body))
            req.Header.Set("Content-Type", "application/json")

            w := httptest.NewRecorder()
            srv.Router.ServeHTTP(w, req)

            assert.Equal(t, tt.wantStatus, w.Code)

            if tt.wantToken {
                var response map[string]interface{}
                json.Unmarshal(w.Body.Bytes(), &response)
                assert.NotEmpty(t, response["token"])
            }
        })
    }
}
```

### 3.3 程序创建测试

**测试文件**: `tests/e2e/program_test.go` + `tests/integration/program_test.go`

#### 场景描述

验证管理员可以创建、查看和删除应用程序。

#### Playwright 测试步骤

| 步骤 | 操作 | 验证点 |
|-----|------|--------|
| 1 | 登录管理后台 | 成功登录 |
| 2 | 点击"新建程序" | 显示创建表单 |
| 3 | 填写程序信息 | 名称: DocuFiller, 描述: 测试程序 |
| 4 | 提交表单 | 程序创建成功 |
| 5 | 检查程序列表 | 新程序出现在列表中 |
| 6 | 点击程序详情 | 显示程序信息和配置选项 |

#### API 测试

```go
// tests/integration/program_test.go
func TestCreateProgram(t *testing.T) {
    srv := setupTestServerWithAdmin(t)
    defer srv.Close()

    // 获取管理员 token
    token := getAdminToken(t, srv)

    // 创建程序
    payload := map[string]interface{}{
        "name":        "TestApp",
        "description": "Test application",
    }
    body, _ := json.Marshal(payload)

    req := httptest.NewRequest("POST", srv.URL+"/api/admin/programs", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    w := httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    programID := response["program_id"].(string)
    assert.NotEmpty(t, programID)

    // 验证数据库
    var program models.Program
    srv.DB.Where("program_id = ?", programID).First(&program)
    assert.Equal(t, "TestApp", program.Name)

    // 验证自动生成的 token
    uploadToken, _ := srv.TokenService.GetToken(programID, "upload", "system")
    assert.NotEmpty(t, uploadToken.TokenValue)

    downloadToken, _ := srv.TokenService.GetToken(programID, "download", "system")
    assert.NotEmpty(t, downloadToken.TokenValue)

    // 验证加密密钥
    assert.NotEmpty(t, program.EncryptionKey)
    assert.Len(t, program.EncryptionKey, 32) // 32 字节
}

func TestListPrograms(t *testing.T) {
    srv := setupTestServerWithAdmin(t)
    defer srv.Close()

    token := getAdminToken(t, srv)

    // 创建多个程序
    createProgram(t, srv, token, "App1", "Description 1")
    createProgram(t, srv, token, "App2", "Description 2")

    // 查询列表
    req := httptest.NewRequest("GET", srv.URL+"/api/admin/programs", nil)
    req.Header.Set("Authorization", "Bearer "+token)

    w := httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response []map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    assert.GreaterOrEqual(t, len(response), 2)
}
```

### 3.4 版本发布和更新测试

**测试文件**: `tests/integration/version_test.go`

#### 场景描述

验证完整的版本发布、查询、下载流程。

#### 测试步骤

| 步骤 | 操作 | 验证点 |
|-----|------|--------|
| 1 | 创建测试程序 | 程序创建成功 |
| 2 | 准备测试 ZIP 包 | 创建测试文件 |
| 3 | 上传版本（stable/1.0.0） | API 返回 200 |
| 4 | 验证文件存储 | 文件在正确路径 |
| 5 | 验证数据库记录 | 版本信息正确 |
| 6 | 查询最新版本 | 返回 1.0.0 |
| 7 | 上传版本（stable/1.1.0） | API 返回 200 |
| 8 | 查询最新版本 | 返回 1.1.0 |
| 9 | 使用 download_token 下载 | 文件内容正确 |
| 10 | 验证下载计数 | 计数增加 |

#### API 测试

```go
// tests/integration/version_test.go
func TestVersionUploadAndDownload(t *testing.T) {
    srv := setupTestServerWithProgram(t)
    defer srv.Close()

    programID := srv.TestProgramID
    uploadToken := srv.TestUploadToken
    downloadToken := srv.TestDownloadToken

    // 创建测试 ZIP 文件
    testContent := []byte("test application package content")
    testFile := createTestZip(t, testContent)

    // 创建 multipart form
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    part, _ := writer.CreateFormFile("file", "testapp-1.0.0.zip")
    io.Copy(part, testFile)

    writer.WriteField("channel", "stable")
    writer.WriteField("version", "1.0.0")
    writer.WriteField("notes", "First release")
    writer.WriteField("mandatory", "false")
    writer.Close()

    // 上传版本
    req := httptest.NewRequest("POST", srv.URL+"/api/programs/"+programID+"/versions", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("Authorization", "Bearer "+uploadToken)

    w := httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "Version uploaded successfully", response["message"])

    // 验证数据库记录
    var version models.Version
    srv.DB.Where("program_id = ? AND version = ?", programID, "1.0.0").First(&version)
    assert.Equal(t, "stable", version.Channel)
    assert.Equal(t, int64(len(testContent)), version.FileSize)
    assert.NotEmpty(t, version.FileHash)

    // 验证文件存储
    expectedPath := filepath.Join(srv.StorageBasePath, programID, "stable", "1.0.0", "testapp-1.0.0.zip")
    assert.FileExists(t, expectedPath)

    // 查询最新版本
    req = httptest.NewRequest("GET", srv.URL+"/api/programs/"+programID+"/versions/latest?channel=stable", nil)
    w = httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "1.0.0", response["version"])

    // 使用 download_token 下载
    req = httptest.NewRequest("GET", srv.URL+"/api/programs/"+programID+"/download/stable/1.0.0", nil)
    req.Header.Set("Authorization", "Bearer "+downloadToken)
    w = httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Equal(t, testContent, w.Body.Bytes())

    // 验证下载计数
    srv.DB.Where("program_id = ? AND version = ?", programID, "1.0.0").First(&version)
    assert.Equal(t, 1, version.DownloadCount)
}
```

#### Playwright UI 测试

```go
// tests/e2e/version_test.go
func TestVersionUploadUI(t *testing.T) {
    // Setup: 启动服务器、登录、创建程序
    srv := setupTestServer(t)
    defer srv.Close()

    pw, _ := playwright.Run()
    defer pw.Stop()

    browser, _ := pw.Chromium.Launch(playwright.BrowserLaunchOptions{Headless: playwright.Bool(false)})
    context, _ := browser.NewContext()
    page, _ := context.NewPage()
    defer browser.Close()

    // 登录
    loginAsAdmin(t, page, srv.URL)

    // 进入程序详情页
    _, _ = page.Goto(srv.URL + "/admin/program/" + srv.TestProgramID)

    // 点击上传版本
    _ = page.Click("#upload-version-btn")

    // 填写版本信息
    _ = page.SelectOption("#channel", "stable")
    _ = page.Fill("#version", "1.0.0")
    _ = page.Fill("#notes", "First release")

    // 选择文件
    testFile := createTestZipFile(t)
    _ = page.setInputFiles("#file", testFile)

    // 提交
    _ = page.Click("#submit-version-btn")

    // 等待成功消息
    _ = page.WaitForSelector(".alert-success")

    // 验证版本出现在列表中
    _ = page.WaitForSelector(".version-list-item:has-text('1.0.0')")
}
```

### 3.5 客户端打包测试

**测试文件**: `tests/integration/client_test.go`

#### 场景描述

验证服务器可以为每个程序生成独立的发布客户端和更新客户端包。

#### 测试步骤

| 步骤 | 操作 | 验证点 |
|-----|------|--------|
| 1 | 创建测试程序 | 程序创建成功 |
| 2 | 请求发布客户端包 | 返回 ZIP 文件 |
| 3 | 解压包内容 | 包含必需文件 |
| 4 | 验证 config.yaml | 配置正确 |
| 5 | 请求更新客户端包 | 返回 ZIP 文件 |
| 6 | 解压并验证 | 内容完整 |
| 7 | 测试发布客户端 | 可以上传版本 |
| 8 | 测试更新客户端 | 可以查询更新 |

#### API 测试

```go
// tests/integration/client_test.go
func TestClientPackaging(t *testing.T) {
    srv := setupTestServerWithProgram(t)
    defer srv.Close()

    programID := srv.TestProgramID

    t.Run("Publish Client Package", func(t *testing.T) {
        req := httptest.NewRequest("GET", srv.URL+"/api/admin/programs/"+programID+"/client/publish", nil)
        req.Header.Set("Authorization", "Bearer "+srv.AdminToken)

        w := httptest.NewRecorder()
        srv.Router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
        assert.Equal(t, "application/zip", w.Header().Get("Content-Type"))

        // 保存 ZIP 文件
        zipPath := filepath.Join(t.TempDir(), "publish-client.zip")
        os.WriteFile(zipPath, w.Body.Bytes(), 0644)

        // 解压验证
        extractDir := filepath.Join(t.TempDir(), "extracted")
        unzip(t, zipPath, extractDir)

        // 验证文件存在
        assert.FileExists(t, filepath.Join(extractDir, "update-admin.exe"))
        assert.FileExists(t, filepath.Join(extractDir, "config.yaml"))
        assert.FileExists(t, filepath.Join(extractDir, "README.md"))
        assert.FileExists(t, filepath.Join(extractDir, "version.txt"))

        // 验证 config.yaml 内容
        configContent, _ := os.ReadFile(filepath.Join(extractDir, "config.yaml"))
        assert.Contains(t, string(configContent), "server:")
        assert.Contains(t, string(configContent), "url:")
        assert.Contains(t, string(configContent), "auth:")
        assert.Contains(t, string(configContent), srv.TestUploadToken)
        assert.Contains(t, string(configContent), srv.TestProgram.EncryptionKey)
    })

    t.Run("Update Client Package", func(t *testing.T) {
        req := httptest.NewRequest("GET", srv.URL+"/api/admin/programs/"+programID+"/client/update", nil)
        req.Header.Set("Authorization", "Bearer "+srv.AdminToken)

        w := httptest.NewRecorder()
        srv.Router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)

        zipPath := filepath.Join(t.TempDir(), "update-client.zip")
        os.WriteFile(zipPath, w.Body.Bytes(), 0644)

        extractDir := filepath.Join(t.TempDir(), "update-extracted")
        unzip(t, zipPath, extractDir)

        assert.FileExists(t, filepath.Join(extractDir, "update-client.exe"))
        assert.FileExists(t, filepath.Join(extractDir, "config.yaml"))

        // 验证配置包含 download_token
        configContent, _ := os.ReadFile(filepath.Join(extractDir, "config.yaml"))
        assert.Contains(t, string(configContent), srv.TestDownloadToken)
    })
}

func TestClientPackageE2E(t *testing.T) {
    srv := setupTestServerWithProgram(t)
    defer srv.Close()

    // 下载发布客户端
    req := httptest.NewRequest("GET", srv.URL+"/api/admin/programs/"+srv.TestProgramID+"/client/publish", nil)
    req.Header.Set("Authorization", "Bearer "+srv.AdminToken)

    w := httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    // 解压到临时目录
    clientDir := filepath.Join(t.TempDir(), "client")
    os.MkdirAll(clientDir, 0755)

    zipPath := filepath.Join(t.TempDir(), "client.zip")
    os.WriteFile(zipPath, w.Body.Bytes(), 0644)
    unzip(t, zipPath, clientDir)

    // 使用包中的配置运行 update-publisher（模拟）
    // 这里可以启动子进程运行实际的 update-publisher.exe
    // 并验证它可以使用内置配置上传版本
}
```

### 3.6 完整闭环测试（核心场景）

**测试文件**: `tests/integration/loop_test.go` + `tests/e2e/loop_test.go`

#### 场景概述

这是最核心的端到端测试，模拟真实用户的完整使用流程。测试必须形成完整的闭环：

1. **发布流程闭环**: 创建程序 → 下载发布客户端包 → 使用客户端上传版本 → 在后台验证
2. **更新流程闭环**: 创建程序 → 上传版本 → 下载更新客户端包 → 使用客户端检查和下载更新 → 验证

#### 3.6.1 发布流程闭环测试

**测试流程图**:

```
┌─────────────────────────────────────────────────────────────────────┐
│                      发布流程完整闭环                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐     │
│  │ 1. 创建   │ -> │ 2. 下载   │ -> │ 3. 解压   │ -> │ 4. 使用   │     │
│  │   程序    │    │发布客户端包│    │客户端包  │    │客户端上传 │     │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘     │
│                                                                     │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐                      │
│  │ 5. 后台   │ <- │ 6. 验证   │ <- │ 7. 检查   │                      │
│  │   验证    │    │ 版本信息  │    │  API     │                      │
│  └──────────┘    └──────────┘    └──────────┘                      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**详细测试步骤**:

| 步骤 | 操作 | 验证点 | 方法 |
|-----|------|--------|------|
| 1 | 管理员登录后台 | 成功登录 | Playwright |
| 2 | 创建新程序 | 程序创建成功，获得 program_id | Playwright + API |
| 3 | 下载发布客户端包 | 获得 ZIP 文件 | API |
| 4 | 解压客户端包到临时目录 | 包含 update-admin.exe 和 config.yaml | Go |
| 5 | 验证 config.yaml 配置 | 包含正确的 server URL、upload_token、encryption_key | Go |
| 6 | 准备测试用的版本 ZIP 包 | 创建测试文件 | Go |
| 7 | **使用下载的 update-admin.exe 上传版本** | 命令执行成功 | exec.Command |
| 8 | 验证上传 API 响应 | 返回 200，上传成功 | Go |
| 9 | **在管理后台查看版本列表** | 新版本出现在列表中 | Playwright |
| 10 | 验证版本详细信息 | 版本号、文件大小、发布时间正确 | Playwright + API |
| 11 | 验证文件存储在服务器 | 文件存在于 data/packages/{programId}/stable/1.0.0/ | Go |
| 12 | **闭环完成** | 从创建到上传到验证，完整流程通过 | - |

**代码实现**:

```go
// tests/integration/loop_test.go
package integration

import (
    "archive/zip"
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/http/httptest"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
    "time"

    "docufiller-update-server/internal/models"
    "github.com/stretchr/testify/assert"
)

// TestPublishLoop 测试完整的发布流程闭环
// 流程: 创建程序 -> 下载发布客户端 -> 使用客户端上传版本 -> 后台验证
func TestPublishLoop(t *testing.T) {
    // Setup: 启动测试服务器并创建管理员
    srv := setupTestServerWithAdmin(t)
    defer srv.Close()

    adminToken := getAdminToken(t, srv)

    // ========== 步骤 1-2: 创建程序 ==========
    t.Log("Step 1-2: Creating program...")

    programName := fmt.Sprintf("test-app-%d", time.Now().Unix())
    createPayload := map[string]interface{}{
        "name":        programName,
        "description": "Test application for loop testing",
    }
    body, _ := json.Marshal(createPayload)

    req := httptest.NewRequest("POST", srv.URL+"/api/admin/programs", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+adminToken)

    w := httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)

    var createResp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &createResp)
    programID := createResp["program_id"].(string)
    assert.NotEmpty(t, programID)

    t.Logf("Program created with ID: %s", programID)

    // ========== 步骤 3: 下载发布客户端包 ==========
    t.Log("Step 3: Downloading publish client package...")

    req = httptest.NewRequest("GET", srv.URL+"/api/admin/programs/"+programID+"/client/publish", nil)
    req.Header.Set("Authorization", "Bearer "+adminToken)

    w = httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Equal(t, "application/zip", w.Header().Get("Content-Type"))

    // 保存 ZIP 包
    clientZipPath := filepath.Join(t.TempDir(), "publish-client.zip")
    err := os.WriteFile(clientZipPath, w.Body.Bytes(), 0644)
    assert.NoError(t, err)

    t.Logf("Publish client package saved to: %s", clientZipPath)

    // ========== 步骤 4: 解压客户端包 ==========
    t.Log("Step 4: Extracting client package...")

    clientDir := filepath.Join(t.TempDir(), "client")
    err = os.MkdirAll(clientDir, 0755)
    assert.NoError(t, err)

    err = unzipFile(clientZipPath, clientDir)
    assert.NoError(t, err)

    // 验证文件存在
    clientExePath := filepath.Join(clientDir, "update-admin.exe")
    configPath := filepath.Join(clientDir, "config.yaml")

    assert.FileExists(t, clientExePath)
    assert.FileExists(t, configPath)
    assert.FileExists(t, filepath.Join(clientDir, "README.md"))

    t.Logf("Client package extracted to: %s", clientDir)

    // ========== 步骤 5: 验证 config.yaml 配置 ==========
    t.Log("Step 5: Validating config.yaml...")

    configContent, err := os.ReadFile(configPath)
    assert.NoError(t, err)

    configStr := string(configContent)

    // 验证服务器 URL（应该是测试服务器的 URL）
    assert.Contains(t, configStr, "server:")
    assert.Contains(t, configStr, "url:")

    // 获取程序的 upload_token
    var program models.Program
    srv.DB.Where("program_id = ?", programID).First(&program)
    uploadToken, _ := srv.TokenService.GetToken(programID, "upload", "system")

    // 验证配置中包含正确的 token
    assert.Contains(t, configStr, uploadToken.TokenValue)
    // 验证配置中包含加密密钥
    assert.Contains(t, configStr, program.EncryptionKey)

    t.Logf("Config validation passed - token: %s..., key length: %d",
        uploadToken.TokenValue[:10], len(program.EncryptionKey))

    // ========== 步骤 6: 准备测试用的版本 ZIP 包 ==========
    t.Log("Step 6: Creating test version package...")

    versionZipPath := createTestVersionZip(t, programName, "1.0.0")

    t.Logf("Test version package created: %s", versionZipPath)

    // ========== 步骤 7: 使用下载的客户端上传版本 ==========
    t.Log("Step 7: Uploading version using downloaded client...")

    // 使用 update-admin.exe 上传版本
    // 命令格式: update-admin.exe upload --program-id {id} --channel stable --version 1.0.0 --file {path}
    clientWorkDir := t.TempDir()
    uploadCmd := exec.Command(clientExePath,
        "upload",
        "--server", srv.URL,
        "--token", uploadToken.TokenValue,
        "--program-id", programID,
        "--channel", "stable",
        "--version", "1.0.0",
        "--file", versionZipPath,
        "--notes", "First release via client loop test",
    )
    uploadCmd.Dir = clientWorkDir

    output, err := uploadCmd.CombinedOutput()
    t.Logf("Client upload output:\n%s", string(output))

    assert.NoError(t, err, "Client upload should succeed")

    // ========== 步骤 8: 验证 API 响应 ==========
    t.Log("Step 8: Verifying upload via API...")

    // 查询版本列表
    req = httptest.NewRequest("GET", srv.URL+"/api/admin/programs/"+programID+"/versions", nil)
    req.Header.Set("Authorization", "Bearer "+adminToken)

    w = httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var versions []map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &versions)

    assert.Greater(t, len(versions), 0, "Should have at least one version")

    // 查找刚上传的版本
    var uploadedVersion map[string]interface{}
    for _, v := range versions {
        if v["version"] == "1.0.0" && v["channel"] == "stable" {
            uploadedVersion = v
            break
        }
    }

    assert.NotNil(t, uploadedVersion, "Uploaded version should be in the list")
    assert.Equal(t, "1.0.0", uploadedVersion["version"])
    assert.Equal(t, "stable", uploadedVersion["channel"])
    assert.Equal(t, "First release via client loop test", uploadedVersion["release_notes"])

    t.Logf("Version verified via API: %+v", uploadedVersion)

    // ========== 步骤 9: 验证文件存储 ==========
    t.Log("Step 9: Verifying file storage on server...")

    var versionRecord models.Version
    srv.DB.Where("program_id = ? AND version = ? AND channel = ?", programID, "1.0.0", "stable").First(&versionRecord)

    assert.NotEmpty(t, versionRecord.FilePath)
    assert.NotEmpty(t, versionRecord.FileName)
    assert.NotEmpty(t, versionRecord.FileHash)
    assert.Greater(t, versionRecord.FileSize, int64(0))

    // 验证文件存在
    assert.FileExists(t, versionRecord.FilePath)

    t.Logf("File storage verified - path: %s, size: %d, hash: %s",
        versionRecord.FilePath, versionRecord.FileSize, versionRecord.FileHash)

    // ========== 步骤 10: 后台 UI 验证 ==========
    t.Log("Step 11: Verifying in admin backend UI...")

    // 这里可以用 Playwright 验证后台显示
    // 或者通过 API 验证统计信息
    req = httptest.NewRequest("GET", srv.URL+"/api/admin/stats", nil)
    req.Header.Set("Authorization", "Bearer "+adminToken)

    w = httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var stats map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &stats)

    totalVersions := int(stats["total_versions"].(float64))
    assert.GreaterOrEqual(t, totalVersions, 1)

    t.Logf("Backend stats verified - total versions: %d", totalVersions)

    // ========== 闭环完成 ==========
    t.Log("✓ Publish loop completed successfully!")
    t.Logf("Program: %s (%s)", programName, programID)
    t.Logf("Version: 1.0.0 (stable)")
    t.Logf("File: %s", versionRecord.FileName)
    t.Logf("Size: %d bytes", versionRecord.FileSize)
}
```

#### 3.6.2 更新流程闭环测试

**测试流程图**:

```
┌─────────────────────────────────────────────────────────────────────┐
│                      更新流程完整闭环                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐     │
│  │ 1. 创建   │ -> │ 2. 手动   │ -> │ 3. 下载   │ -> │ 4. 解压   │     │
│  │   程序    │    │上传版本  │    │更新客户端 │    │客户端包  │     │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘     │
│                                                                     │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐     │
│  │ 5. 使用   │ -> │ 6. 验证   │ -> │ 7. 使用   │ -> │ 8. 验证   │     │
│  │客户端检查 │    │检查结果  │    │客户端下载 │    │下载文件  │     │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**代码实现**:

```go
// TestUpdateLoop 测试完整的更新流程闭环
// 流程: 创建程序 -> 上传版本 -> 下载更新客户端 -> 使用客户端检查和下载更新 -> 验证
func TestUpdateLoop(t *testing.T) {
    // Setup: 启动测试服务器并创建管理员
    srv := setupTestServerWithAdmin(t)
    defer srv.Close()

    adminToken := getAdminToken(t, srv)

    // ========== 步骤 1-2: 创建程序并上传初始版本 ==========
    t.Log("Step 1-2: Creating program and uploading initial version...")

    programName := fmt.Sprintf("update-test-app-%d", time.Now().Unix())

    // 创建程序
    createPayload := map[string]interface{}{
        "name":        programName,
        "description": "Test application for update loop testing",
    }
    body, _ := json.Marshal(createPayload)

    req := httptest.NewRequest("POST", srv.URL+"/api/admin/programs", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+adminToken)

    w := httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)

    var createResp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &createResp)
    programID := createResp["program_id"].(string)

    // 获取 upload_token 并上传版本
    var program models.Program
    srv.DB.Where("program_id = ?", programID).First(&program)
    uploadToken, _ := srv.TokenService.GetToken(programID, "upload", "system")

    versionZipPath := createTestVersionZip(t, programName, "1.0.0")

    // 使用 API 上传版本
    uploadBody := &bytes.Buffer{}
    writer := multipart.NewWriter(uploadBody)

    part, _ := writer.CreateFormFile("file", filepath.Base(versionZipPath))
    fileContent, _ := os.ReadFile(versionZipPath)
    part.Write(fileContent)

    writer.WriteField("channel", "stable")
    writer.WriteField("version", "1.0.0")
    writer.WriteField("notes", "Initial release")
    writer.WriteField("mandatory", "false")
    writer.Close()

    req = httptest.NewRequest("POST", srv.URL+"/api/programs/"+programID+"/versions", uploadBody)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("Authorization", "Bearer "+uploadToken.TokenValue)

    w = httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    t.Logf("Program %s created with version 1.0.0", programID)

    // ========== 步骤 3: 下载更新客户端包 ==========
    t.Log("Step 3: Downloading update client package...")

    req = httptest.NewRequest("GET", srv.URL+"/api/admin/programs/"+programID+"/client/update", nil)
    req.Header.Set("Authorization", "Bearer "+adminToken)

    w = httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Equal(t, "application/zip", w.Header().Get("Content-Type"))

    clientZipPath := filepath.Join(t.TempDir(), "update-client.zip")
    os.WriteFile(clientZipPath, w.Body.Bytes(), 0644)

    t.Logf("Update client package saved: %s", clientZipPath)

    // ========== 步骤 4: 解压客户端包 ==========
    t.Log("Step 4: Extracting update client package...")

    clientDir := filepath.Join(t.TempDir(), "update-client")
    os.MkdirAll(clientDir, 0755)
    unzipFile(clientZipPath, clientDir)

    clientExePath := filepath.Join(clientDir, "update-client.exe")
    configPath := filepath.Join(clientDir, "config.yaml")

    assert.FileExists(t, clientExePath)
    assert.FileExists(t, configPath)

    t.Logf("Update client extracted to: %s", clientDir)

    // ========== 步骤 5: 使用客户端检查更新 ==========
    t.Log("Step 5: Checking for updates using client...")

    downloadToken, _ := srv.TokenService.GetToken(programID, "download", "system")

    // 使用 update-client.exe 检查更新
    checkCmd := exec.Command(clientExePath,
        "check",
        "--server", srv.URL,
        "--token", downloadToken.TokenValue,
        "--program-id", programID,
        "--channel", "stable",
    )

    checkOutput, err := checkCmd.CombinedOutput()
    t.Logf("Check command output:\n%s", string(checkOutput))

    assert.NoError(t, err, "Check command should succeed")

    // 解析输出，应该显示有新版本 1.0.0
    assert.Contains(t, string(checkOutput), "1.0.0")

    // ========== 步骤 6: 验证 API 检查结果 ==========
    t.Log("Step 6: Verifying check result via API...")

    req = httptest.NewRequest("GET", srv.URL+"/api/programs/"+programID+"/versions/latest?channel=stable", nil)
    w = httptest.NewRecorder()
    srv.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var latestResp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &latestResp)

    assert.Equal(t, "1.0.0", latestResp["version"])
    assert.Equal(t, "stable", latestResp["channel"])

    t.Logf("Latest version verified: %+v", latestResp)

    // ========== 步骤 7: 使用客户端下载更新 ==========
    t.Log("Step 7: Downloading update using client...")

    downloadDir := filepath.Join(t.TempDir(), "downloads")
    os.MkdirAll(downloadDir, 0755)

    downloadCmd := exec.Command(clientExePath,
        "download",
        "--server", srv.URL,
        "--token", downloadToken.TokenValue,
        "--program-id", programID,
        "--channel", "stable",
        "--version", "1.0.0",
        "--output", downloadDir,
    )

    downloadOutput, err := downloadCmd.CombinedOutput()
    t.Logf("Download command output:\n%s", string(downloadOutput))

    assert.NoError(t, err, "Download command should succeed")

    // ========== 步骤 8: 验证下载的文件 ==========
    t.Log("Step 8: Verifying downloaded file...")

    downloadedFilePath := filepath.Join(downloadDir, fmt.Sprintf("%s-1.0.0-stable.zip", programName))

    assert.FileExists(t, downloadedFilePath, "Downloaded file should exist")

    // 验证文件大小
    fileInfo, _ := os.Stat(downloadedFilePath)
    assert.Equal(t, int64(len(fileContent)), fileInfo.Size())

    // 验证文件内容（SHA256）
    downloadedContent, _ := os.ReadFile(downloadedFilePath)
    assert.Equal(t, fileContent, downloadedContent, "Downloaded content should match")

    // 验证下载计数增加
    var versionRecord models.Version
    srv.DB.Where("program_id = ? AND version = ? AND channel = ?", programID, "1.0.0", "stable").First(&versionRecord)
    assert.Equal(t, 1, versionRecord.DownloadCount, "Download count should be incremented")

    t.Logf("Download verified - file: %s, size: %d, count: %d",
        downloadedFilePath, fileInfo.Size(), versionRecord.DownloadCount)

    // ========== 闭环完成 ==========
    t.Log("✓ Update loop completed successfully!")
    t.Logf("Program: %s (%s)", programName, programID)
    t.Logf("Latest version: 1.0.0 (stable)")
    t.Logf("Downloaded to: %s", downloadedFilePath)
}
```

#### 3.6.3 混合 UI+API 闭环测试

**测试文件**: `tests/e2e/loop_test.go`

使用 Playwright 验证后台 UI，同时使用 API 和命令行工具执行操作：

```go
// tests/e2e/loop_test.go
func TestPublishLoopWithUI(t *testing.T) {
    // Setup
    srv := setupTestServer(t)
    defer srv.Close()

    pw, _ := playwright.Run()
    defer pw.Stop()

    browser, _ := pw.Chromium.Launch(playwright.BrowserLaunchOptions{Headless: playwright.Bool(false)})
    context, _ := browser.NewContext()
    page, _ := context.NewPage()
    defer browser.Close()

    programID := ""
    programName := ""

    // ========== 步骤 1-2: 通过 UI 创建程序 ==========
    t.Log("Creating program via UI...")

    loginAsAdmin(t, page, srv.URL)
    page.Click("#new-program-btn")

    programName = fmt.Sprintf("ui-loop-test-%d", time.Now().Unix())
    page.Fill("#program-name", programName)
    page.Fill("#program-description", "UI loop test program")
    page.Click("#submit-program-btn")

    page.WaitForSelector(".alert-success")
    page.WaitForURL("*admin/program/*")

    // 从 URL 获取 program_id
    currentURL := page.URL()
    parts := strings.Split(currentURL, "/")
    programID = parts[len(parts)-1]

    t.Logf("Program created via UI: %s (%s)", programName, programID)

    // ========== 步骤 3-4: 通过 API 下载并解压客户端 ==========
    t.Log("Downloading publish client via API...")

    adminToken := getAdminToken(t, srv)
    clientZipPath := downloadPublishClient(t, srv.URL, adminToken, programID)

    clientDir := filepath.Join(t.TempDir(), "client")
    os.MkdirAll(clientDir, 0755)
    unzipFile(clientZipPath, clientDir)

    // ========== 步骤 5-6: 使用客户端上传版本 ==========
    t.Log("Uploading version via client...")

    versionZipPath := createTestVersionZip(t, programName, "1.0.0")
    clientExePath := filepath.Join(clientDir, "update-admin.exe")

    var program models.Program
    srv.DB.Where("program_id = ?", programID).First(&program)
    uploadToken, _ := srv.TokenService.GetToken(programID, "upload", "system")

    uploadCmd := exec.Command(clientExePath,
        "upload",
        "--server", srv.URL,
        "--token", uploadToken.TokenValue,
        "--program-id", programID,
        "--channel", "stable",
        "--version", "1.0.0",
        "--file", versionZipPath,
    )

    output, _ := uploadCmd.CombinedOutput()
    t.Logf("Upload output: %s", string(output))

    // ========== 步骤 7-8: 在 UI 中验证 ==========
    t.Log("Verifying in UI...")

    page.Reload()
    page.WaitForSelector(".version-list-item")

    // 验证版本出现在列表中
    versionItems := page.Locator(".version-list-item")
    count, _ := versionItems.Count()
    assert.Greater(t, count, 0, "Should have at least one version")

    // 验证版本信息
    page.WaitForSelector(".version-list-item:has-text('1.0.0')")
    page.WaitForSelector(".version-list-item:has-text('stable')")

    // 点击版本详情
    page.Click(".version-list-item:has-text('1.0.0') .detail-btn")
    page.WaitForSelector(".version-detail")

    // 验证详细信息
    versionDetail := page.Locator(".version-detail")
    detailText, _ := detailDetail.TextContent()
    assert.Contains(t, detailText, "1.0.0")
    assert.Contains(t, detailText, "stable")

    t.Log("✓ UI loop test completed!")
}
```

#### 闭环测试验收标准

| 场景 | 验收标准 |
|-----|---------|
| 发布流程闭环 | 1. 程序创建成功<br>2. 客户端包下载成功<br>3. 配置文件包含正确的 token 和密钥<br>4. 客户端可执行文件完整<br>5. 使用客户端上传成功<br>6. 版本在后台可见<br>7. 文件存储在正确路径<br>8. 所有元数据正确 |
| 更新流程闭环 | 1. 版本已上传<br>2. 客户端包下载成功<br>3. 配置包含 download_token<br>4. 客户端检查更新成功<br>5. API 返回正确版本<br>6. 客户端下载成功<br>7. 下载的文件完整<br>8. 下载计数正确增加 |

### 3.7 发布包完整性测试

**测试文件**: `tests/integration/package_test.go`

#### 场景描述

验证服务器发布包包含所有必要文件，可以独立部署运行。

#### 测试步骤

| 步骤 | 操作 | 验证点 |
|-----|------|--------|
| 1 | 运行构建脚本 | 构建成功 |
| 2 | 检查发布目录 | 所有文件存在 |
| 3 | 复制到临时目录 | 模拟部署 |
| 4 | 修改配置 | 更改端口和数据库路径 |
| 5 | 启动服务器 | 成功启动 |
| 6 | 验证初始化流程 | 可以完成首次设置 |
| 7 | 验证客户端文件 | data/clients 文件可用 |
| 8 | 执行闭环测试 | 完整的发布和更新流程 |
| 9 | 停止并清理 | 无残留文件 |

#### 测试代码

```go
// tests/integration/package_test.go
func TestDistributionPackage(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping package test in short mode")
    }

    // 1. 检查发布包文件
    distDir := filepath.Join("..", "dist", "update-server-1.0.0")

    requiredFiles := []string{
        "update-server.exe",
        "config.yaml",
        "README.md",
        "data/clients/update-publisher.exe",
        "data/clients/update-client.exe",
    }

    for _, file := range requiredFiles {
        path := filepath.Join(distDir, file)
        if _, err := os.Stat(path); os.IsNotExist(err) {
            t.Errorf("required file missing: %s", file)
        }
    }

    // 2. 模拟部署
    deployDir := t.TempDir()
    copyDir(t, distDir, deployDir)

    // 3. 修改配置
    configPath := filepath.Join(deployDir, "config.yaml")
    configContent, _ := os.ReadFile(configPath)
    configContent = bytes.ReplaceAll(configContent, []byte("port: 8080"), []byte("port: 18081"))
    configContent = bytes.ReplaceAll(configContent, []byte("./data/versions.db"), []byte(filepath.Join(deployDir, "test.db")))
    os.WriteFile(configPath, configContent, 0644)

    // 4. 启动服务器
    cmd := exec.Command(filepath.Join(deployDir, "update-server.exe"))
    cmd.Dir = deployDir
    if err := cmd.Start(); err != nil {
        t.Fatalf("failed to start server: %v", err)
    }
    defer cmd.Process.Kill()

    // 等待服务器启动
    time.Sleep(2 * time.Second)

    // 5. 验证服务器响应
    resp, err := http.Get("http://localhost:18081/api/health")
    if err != nil {
        t.Fatalf("server not responding: %v", err)
    }
    defer resp.Body.Close()
    assert.Equal(t, http.StatusOK, resp.StatusCode)

    // 6. 验证初始化流程
    // ... Playwright 测试初始化

    // 7. 验证客户端文件可下载
    // ... API 测试
}
```

## 4. 发布目录结构

### 4.1 服务器发布包结构

```
update-server-v1.0.0/
├── update-server.exe              # 服务器可执行文件
├── config.yaml                     # 配置文件模板
├── config.production.yaml          # 生产环境配置模板（可选）
├── README.md                       # 部署和使用说明
├── LICENSE                         # 许可证
├── data/                           # 数据目录（首次运行创建）
│   ├── clients/                    # 客户端文件
│   │   ├── update-publisher.exe
│   │   ├── update-client.exe
│   │   ├── update-publisher.usage.txt
│   │   └── update-client.config.yaml
│   ├── packages/                   # 版本包存储（自动创建）
│   │   └── {programId}/
│   │       ├── stable/
│   │       └── beta/
│   └── versions.db                 # SQLite 数据库（自动创建）
├── logs/                           # 日志目录（自动创建）
└── scripts/                        # 辅助脚本（可选）
    ├── install-service.bat         # Windows 服务安装
    ├── start.bat                   # 启动脚本
    └── create-admin.bat            # 管理员创建工具
```

### 4.2 客户端包结构（下载后）

**发布客户端包** (`{programName}-client-publish.zip`):

```
docufiller-client-publish/
├── update-admin.exe                # 发布工具（或 update-publisher.exe）
├── config.yaml                     # 预配置
│   # server.url: "http://your-server:8080"
│   # program.id: "docufiller"
│   # auth.token: "{upload_token}"
│   # auth.encryption_key: "{encryption_key}"
├── README.md                       # 使用说明
└── version.txt                     # 版本信息
```

**更新客户端包** (`{programName}-client-update.zip`):

```
docufiller-client-update/
├── update-client.exe               # 更新客户端
├── config.yaml                     # 预配置（使用 download_token）
├── README.md                       # 使用说明
└── version.txt                     # 版本信息
```

## 5. 构建和打包脚本

### 5.1 增强的 build-all.bat

```batch
@echo off
REM Update Server All-in-One Build and Package Script
REM This script builds all components and creates distribution package.

setlocal enabledelayedexpansion

set "VERSION=%~1"
if "%VERSION%"=="" set "VERSION=1.0.0"

set "SCRIPT_DIR=%~dp0"
set "OUTPUT_DIR=%SCRIPT_DIR%bin"
set "DIST_DIR=%SCRIPT_DIR%dist\update-server-%VERSION%"
set "CLIENT_OUTPUT_DIR=%OUTPUT_DIR%\clients"
set "SERVER_CLIENT_DIR=%SCRIPT_DIR%data\clients"

echo ========================================
echo   Update Server Builder v%VERSION%
echo ========================================
echo.

echo [1/7] Cleaning output directories...
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%" 2>nul
if exist "%DIST_DIR%" rmdir /s /q "%DIST_DIR%" 2>nul
if exist "%SERVER_CLIENT_DIR%" rmdir /s /q "%SERVER_CLIENT_DIR%" 2>nul

mkdir "%OUTPUT_DIR%"
mkdir "%CLIENT_OUTPUT_DIR%"
mkdir "%DIST_DIR%"
mkdir "%DIST_DIR%\data\clients"
mkdir "%DIST_DIR%\scripts"
echo Done.
echo.

echo [2/7] Building Update Server...
cd /d "%SCRIPT_DIR%cmd\update-server"
go build -ldflags "-X main.version=%VERSION%" -o "%OUTPUT_DIR%\update-server.exe" .
if errorlevel 1 goto :error
echo SUCCESS: Built update-server.exe
echo.

echo [3/7] Building Update Publisher...
cd /d "%SCRIPT_DIR%cmd\update-publisher"
go build -ldflags "-X main.version=%VERSION%" -o "%CLIENT_OUTPUT_DIR%\update-publisher.exe" .
if errorlevel 1 goto :error
echo SUCCESS: Built update-publisher.exe
echo.

echo [4/7] Building Update Client...
cd /d "%SCRIPT_DIR%cmd\update-client"
go build -ldflags "-X main.version=%VERSION%" -o "%CLIENT_OUTPUT_DIR%\update-client.exe" .
if errorlevel 1 goto :error
echo SUCCESS: Built update-client.exe
echo.

echo [5/7] Copying client executables to server data directory...
if not exist "%SERVER_CLIENT_DIR%" mkdir "%SERVER_CLIENT_DIR%"

copy /Y "%CLIENT_OUTPUT_DIR%\update-publisher.exe" "%SERVER_CLIENT_DIR%\update-publisher.exe" >nul
copy /Y "%CLIENT_OUTPUT_DIR%\update-client.exe" "%SERVER_CLIENT_DIR%\update-client.exe" >nul
copy /Y "%SCRIPT_DIR%cmd\update-publisher\update-publisher.usage.txt" "%SERVER_CLIENT_DIR%\" >nul
copy /Y "%SCRIPT_DIR%cmd\update-client\update-client.config.yaml" "%SERVER_CLIENT_DIR%\" >nul

copy /Y "%CLIENT_OUTPUT_DIR%\update-publisher.exe" "%DIST_DIR%\data\clients\" >nul
copy /Y "%CLIENT_OUTPUT_DIR%\update-client.exe" "%DIST_DIR%\data\clients\" >nul
echo Done.
echo.

echo [6/7] Creating distribution package...
copy /Y "%OUTPUT_DIR%\update-server.exe" "%DIST_DIR%\" >nul
copy /Y "%SCRIPT_DIR%config.yaml" "%DIST_DIR%\config.template.yaml" >nul
copy /Y "%SCRIPT_DIR%README.md" "%DIST_DIR%\" >nul 2>nul
copy /Y "%SCRIPT_DIR%LICENSE" "%DIST_DIR%\" >nul 2>nul

REM Create deployment scripts
echo @echo off > "%DIST_DIR%\scripts\start.bat"
echo start "" "%~dp0..\update-server.exe" >> "%DIST_DIR%\scripts\start.bat"

echo @echo off > "%DIST_DIR%\scripts\create-admin.bat"
echo "%~dp0..\update-server.exe" create-admin >> "%DIST_DIR%\scripts\create-admin.bat"
echo Done.
echo.

echo [7/7] Creating ZIP archive...
powershell -Command "Compress-Archive -Path '%DIST_DIR%' -DestinationPath '%SCRIPT_DIR%dist\update-server-%VERSION%.zip' -Force"
if errorlevel 1 goto :error
echo SUCCESS: Created dist\update-server-%VERSION%.zip
echo.

echo ========================================
echo   Build Completed Successfully!
echo ========================================
echo.
echo Distribution: %DIST_DIR%
echo Archive: dist\update-server-%VERSION%.zip
echo.
goto :end

:error
echo.
echo ========================================
echo   Build Failed!
echo ========================================
echo.
pause

:end
endlocal
pause
```

### 5.2 E2E 测试运行脚本

**scripts/test-e2e.bat**:

```batch
@echo off
REM Run End-to-End Tests

setlocal enabledelayedexpansion

echo ========================================
echo   E2E Test Runner
echo ========================================
echo.

echo [1/4] Building test binaries...
go build -o "./bin/test-server.exe" "./cmd/test-server"
if errorlevel 1 (
    echo ERROR: Failed to build test server
    goto :error
)
echo Done.
echo.

echo [2/4] Installing Playwright...
go run github.com/playwright-community/playwright-go/cmd/install@latest
if errorlevel 1 (
    echo WARNING: Playwright install failed, browser tests may fail
)
echo.

echo [3/4] Running E2E tests...
go test -v -tags=e2e ./tests/e2e/... -timeout 10m
if errorlevel 1 (
    echo ERROR: E2E tests failed
    goto :error
)
echo Done.
echo.

echo [4/4] Running integration tests...
go test -v ./tests/integration/... -timeout 5m
if errorlevel 1 (
    echo ERROR: Integration tests failed
    goto :error
)
echo Done.
echo.

echo ========================================
echo   All Tests Passed!
echo ========================================
echo.
goto :end

:error
echo.
echo ========================================
echo   Tests Failed!
echo ========================================
echo.
exit /b 1

:end
endlocal
```

## 6. 测试执行计划

### 6.1 测试优先级

| 优先级 | 测试套件 | 执行频率 | 耗时 |
|-------|---------|---------|------|
| P0 | 首次运行引导 | 每次构建 | 1 分钟 |
| P0 | 登录认证 | 每次构建 | 30 秒 |
| P0 | API 端点 | 每次构建 | 2 分钟 |
| P1 | 版本上传下载 | 每次 PR | 3 分钟 |
| P1 | 客户端打包 | 每次 PR | 2 分钟 |
| P2 | E2E 完整流程 | 每日构建 | 10 分钟 |
| P2 | 发布包验证 | 发布前 | 5 分钟 |

### 6.2 CI/CD 集成

```yaml
# .github/workflows/test.yml
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  unit-tests:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run unit tests
        run: go test -v ./...

  integration-tests:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run integration tests
        run: go test -v ./tests/integration/...

  e2e-tests:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Install Playwright
        run: go run github.com/playwright-community/playwright-go/cmd/install@latest
      - name: Run E2E tests
        run: go test -v -tags=e2e ./tests/e2e/...

  package-validation:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Build distribution package
        run: .\build-all.bat %GITHUB_REF_NAME%
      - name: Validate package
        run: go test -v ./tests/integration/package_test.go
```

## 7. 总结

本测试方案设计了一个完整的自动化测试体系，涵盖：

1. **首次运行引导流程** - 验证新用户初始化体验
2. **管理员认证** - 登录、会话管理、权限控制
3. **程序管理** - 创建、配置应用程序
4. **版本发布** - 上传、查询、下载完整流程
5. **客户端打包** - 发布和更新客户端的生成和验证
6. **完整闭环测试**（核心） - 真实场景的端到端验证
7. **发布包验证** - 确保部署包的完整性

### 核心价值：闭环测试

**完整的闭环测试是本方案的精华**，它验证了真实的用户使用场景：

#### 发布流程闭环
```
创建程序 → 下载发布客户端包 → 解压验证配置 → 使用客户端上传版本 → 后台验证版本
```

#### 更新流程闭环
```
创建程序并上传版本 → 下载更新客户端包 → 使用客户端检查更新 → 使用客户端下载 → 验证文件完整性
```

#### 闭环测试的价值

| 价值点 | 说明 |
|-------|------|
| **真实场景验证** | 模拟用户完整的操作流程，而非孤立的功能点 |
| **客户端工具验证** | 确保下载的客户端包确实可以正常工作 |
| **配置正确性** | 验证自动生成的配置包含正确的 token 和密钥 |
| **端到端可见性** | 从创建到发布再到更新的完整链路可见 |
| **集成验证** | 验证各组件之间的集成是否正确 |
| **发布信心** | 通过闭环测试可以确信发布包可以实际部署使用 |

通过 Playwright + Go 测试的混合方式，结合完整的闭环测试，可以在保证测试覆盖率的同时，提供真实的用户场景验证，确保系统可以真正投入使用。

### 后续实施步骤

1. 实现 Playwright 测试框架
2. 扩展现有集成测试
3. 创建测试辅助工具和 fixtures
4. 实现 build-all.bat 增强版
5. 配置 CI/CD 流水线
6. 编写测试使用文档
