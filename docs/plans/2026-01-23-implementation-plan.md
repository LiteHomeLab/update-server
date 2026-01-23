# Update Server 自动化测试实施计划

**创建日期**: 2026-01-23
**状态**: 待执行
**设计文档**: 2026-01-23-automated-e2e-test-design.md

## 概述

本实施计划将设计文档中的测试方案转换为具体的、可执行的任务列表。每个任务都是独立且可验证的。

---

## 阶段 1: 基础设施搭建

### 任务 1.1: 创建测试目录结构

**目标**: 按照设计文档创建测试目录结构

**步骤**:
1. 创建 `tests/` 主目录
2. 创建 `tests/e2e/` 目录（Playwright E2E 测试）
3. 创建 `tests/integration/` 目录（API 集成测试）
4. 创建 `tests/helpers/` 目录（测试辅助工具）
5. 创建 `tests/config/` 目录（测试配置）
6. 创建 `tmp/` 目录（临时测试文件）

**验证**:
- 所有目录创建成功
- 目录结构与设计文档一致

**预计时间**: 5 分钟

---

### 任务 1.2: 实现测试服务器辅助工具

**目标**: 创建可复用的测试服务器管理工具

**文件**: `tests/helpers/server.go`

**步骤**:
1. 创建 `tests/helpers/server.go` 文件
2. 实现 `TestServer` 结构体，包含：
   - `Router *gin.Engine`
   - `DB *gorm.DB`
   - `URL string`
   - `TempDir string`
   - `TokenService *service.TokenService`
3. 实现 `setupTestServer(t *testing.T) *TestServer` 函数：
   - 创建临时目录
   - 初始化内存数据库
   - 自动迁移所有表模型
   - 设置 Gin 路由（包含所有 API 端点）
   - 返回测试服务器实例
4. 实现 `setupTestServerWithAdmin(t *testing.T) *TestServer` 函数：
   - 调用 `setupTestServer`
   - 创建测试管理员账户
   - 返回包含管理员的服务器实例
5. 实现 `setupTestServerWithProgram(t *testing.T) *TestServer` 函数：
   - 调用 `setupTestServerWithAdmin`
   - 创建测试程序
   - 返回包含程序的服务器实例
6. 实现 `(srv *TestServer) Close()` 方法：
   - 关闭数据库连接
   - 清理临时目录

**验证**:
- 文件创建成功
- 代码编译通过
- 基本功能覆盖所需测试场景

**预计时间**: 20 分钟

---

### 任务 1.3: 实现测试数据 Fixtures

**目标**: 创建测试数据和工具函数

**文件**: `tests/helpers/fixtures.go`

**步骤**:
1. 创建 `tests/helpers/fixtures.go` 文件
2. 实现 `createTestProgram(t *testing.T, srv *TestServer, name, description string) string` 函数
3. 实现 `createTestVersion(t *testing.T, srv *TestServer, programID, channel, version string)` 函数
4. 实现 `createTestZip(t *testing.T, programName, version string) string` 函数：
   - 创建临时 ZIP 文件
   - 包含测试内容
   - 返回文件路径
5. 实现 `getAdminToken(t *testing.T, srv *TestServer) string` 函数
6. 实现 `getProgramTokens(t *testing.T, srv *TestServer, programID string) (upload, download string)` 函数

**验证**:
- 文件创建成功
- 代码编译通过

**预计时间**: 15 分钟

---

### 任务 1.4: 实现自定义断言

**目标**: 创建可复用的断言函数

**文件**: `tests/helpers/assertions.go`

**步骤**:
1. 创建 `tests/helpers/assertions.go` 文件
2. 实现 `assertFileExists(t *testing.T, path string)` 函数
3. 实现 `assertFileContains(t *testing.T, path, content string)` 函数
4. 实现 `assertZipStructure(t *testing.T, zipPath string, expectedFiles []string)` 函数
5. 实现 `unzipFile(t *testing.T, zipPath, destDir string) error` 函数
6. 实现 `compareFileHash(t *testing.T, file1, file2 string)` 函数

**验证**:
- 文件创建成功
- 代码编译通过

**预计时间**: 15 分钟

---

### 任务 1.5: 配置 Playwright

**目标**: 安装和配置 Playwright 测试框架

**步骤**:
1. 安装 Playwright Go SDK:
   ```bash
   go get github.com/playwright-community/playwright-go
   ```
2. 安装 Playwright 浏览器:
   ```bash
   go run github.com/playwright-community/playwright-go/cmd/install@latest
   ```
3. 创建 `tests/e2e/common_test.go` 文件
4. 实现 Playwright 测试辅助函数：
   - `setupPlaywright(t *testing.T) (*playwright.Playwright, *playwright.Browser, *playwright.Page)`
   - `teardownPlaywright(pw *playwright.Playwright, browser *playwright.Browser)`
5. 创建简单的首页测试验证 Playwright 可用

**验证**:
- Playwright 安装成功
- 测试可以启动浏览器
- 可以访问测试服务器

**预计时间**: 15 分钟

---

## 阶段 2: 集成测试实现

### 任务 2.1: 实现 API 基础测试

**目标**: 实现 API 端点的集成测试

**文件**: `tests/integration/api_test.go`

**步骤**:
1. 创建 `tests/integration/api_test.go` 文件
2. 实现 `TestHealthCheck` 函数
3. 实现 `TestAdminLoginAPI` 函数（测试有效/无效凭据）
4. 实现 `TestGetStats` 函数
5. 每个测试使用 `setupTestServerWithAdmin` 创建测试环境
6. 使用 `defer srv.Close()` 清理资源

**验证**:
- 所有测试通过
- 覆盖主要 API 端点

**预计时间**: 20 分钟

---

### 任务 2.2: 实现程序管理 API 测试

**目标**: 实现程序创建和管理的 API 测试

**文件**: `tests/integration/program_test.go`

**步骤**:
1. 创建 `tests/integration/program_test.go` 文件
2. 实现 `TestCreateProgram` 函数：
   - 创建程序
   - 验证响应（HTTP 201）
   - 验证数据库记录
   - 验证自动生成的 Token 和加密密钥
3. 实现 `TestListPrograms` 函数
4. 实现 `TestGetProgramDetail` 函数
5. 实现 `TestDeleteProgram` 函数（软删除）

**验证**:
- 所有测试通过
- 验证 Token 和密钥生成正确

**预计时间**: 25 分钟

---

### 任务 2.3: 实现版本管理 API 测试

**目标**: 实现版本上传和下载的 API 测试

**文件**: `tests/integration/version_test.go`

**步骤**:
1. 创建 `tests/integration/version_test.go` 文件
2. 实现 `TestVersionUploadAndDownload` 函数：
   - 创建测试 ZIP 文件
   - 使用 upload_token 上传
   - 验证文件存储
   - 验证数据库记录
   - 使用 download_token 下载
   - 验证文件内容
   - 验证下载计数
3. 实现 `TestGetLatestVersion` 函数
4. 实现 `TestVersionList` 函数（频道过滤）
5. 实现 `TestDeleteVersion` 函数

**验证**:
- 所有测试通过
- 文件存储路径正确
- SHA256 哈希计算正确

**预计时间**: 30 分钟

---

### 任务 2.4: 实现客户端打包测试

**目标**: 测试服务器生成客户端包的功能

**文件**: `tests/integration/client_test.go`

**步骤**:
1. 创建 `tests/integration/client_test.go` 文件
2. 实现 `TestClientPackaging` 函数：
   - 测试发布客户端包生成
   - 测试更新客户端包生成
3. 实现包内容验证：
   - 解压 ZIP 文件
   - 验证必需文件存在（exe, config.yaml, README.md, version.txt）
   - 验证 config.yaml 包含正确的 token 和密钥
4. 验证可执行文件完整

**验证**:
- 所有测试通过
- 生成的包结构正确
- 配置文件内容正确

**预计时间**: 25 分钟

---

### 任务 2.5: 实现发布流程闭环测试（核心）

**目标**: 完整测试从创建程序到使用客户端上传版本的闭环

**文件**: `tests/integration/loop_test.go`

**步骤**:
1. 创建 `tests/integration/loop_test.go` 文件
2. 实现 `TestPublishLoop` 函数，完整流程：
   - 步骤 1-2: 创建程序
   - 步骤 3: 下载发布客户端包
   - 步骤 4: 解压客户端包
   - 步骤 5: 验证 config.yaml 配置
   - 步骤 6: 准备测试版本 ZIP
   - 步骤 7: **使用下载的客户端上传版本** (exec.Command)
   - 步骤 8: 验证 API 响应
   - 步骤 9: 验证文件存储
   - 步骤 10: 验证后台统计
3. 添加详细的日志输出
4. 验证所有验收标准

**验证**:
- 闭环测试完整通过
- 客户端工具可以正常工作
- 后台可以看到上传的版本

**预计时间**: 40 分钟

---

### 任务 2.6: 实现更新流程闭环测试（核心）

**目标**: 完整测试从上传版本到客户端下载更新的闭环

**文件**: `tests/integration/loop_test.go` (继续)

**步骤**:
1. 在 `loop_test.go` 中实现 `TestUpdateLoop` 函数，完整流程：
   - 步骤 1-2: 创建程序并上传版本
   - 步骤 3: 下载更新客户端包
   - 步骤 4: 解压客户端包
   - 步骤 5: **使用客户端检查更新** (exec.Command)
   - 步骤 6: 验证 API 检查结果
   - 步骤 7: **使用客户端下载更新** (exec.Command)
   - 步骤 8: 验证下载的文件
2. 添加详细的日志输出
3. 验证所有验收标准

**验证**:
- 闭环测试完整通过
- 客户端检查和下载功能正常
- 文件完整性验证通过
- 下载计数正确

**预计时间**: 40 分钟

---

## 阶段 3: E2E 测试实现

### 任务 3.1: 实现首次运行引导 E2E 测试

**目标**: 使用 Playwright 测试首次运行初始化流程

**文件**: `tests/e2e/setup_test.go`

**步骤**:
1. 创建 `tests/e2e/setup_test.go` 文件
2. 实现 `TestFirstRunSetupFlow` 函数：
   - 启动未初始化的测试服务器
   - 使用 Playwright 访问根路径，验证重定向到 `/setup`
   - 填写管理员表单
   - 提交并验证成功
   - 验证可以登录
3. 实现边界条件测试（可选）

**验证**:
- E2E 测试通过
- 初始化流程完整验证

**预计时间**: 30 分钟

---

### 任务 3.2: 实现管理后台 E2E 测试

**目标**: 测试管理后台的 UI 交互

**文件**: `tests/e2e/admin_test.go`

**步骤**:
1. 创建 `tests/e2e/admin_test.go` 文件
2. 实现 `TestAdminLoginUI` 函数
3. 实现 `TestProgramManagementUI` 函数：
   - 登录后台
   - 创建程序
   - 查看程序列表
   - 查看程序详情
4. 实现 `TestVersionManagementUI` 函数：
   - 进入程序详情
   - 上传版本
   - 查看版本列表

**验证**:
- 所有 UI 测试通过
- 交互流程正确

**预计时间**: 35 分钟

---

### 任务 3.3: 实现混合 UI+API 闭环测试

**目标**: 结合 Playwright UI 和 API 命令行工具的完整闭环测试

**文件**: `tests/e2e/loop_test.go`

**步骤**:
1. 创建 `tests/e2e/loop_test.go` 文件
2. 实现 `TestPublishLoopWithUI` 函数：
   - 通过 UI 创建程序
   - 通过 API 下载客户端包
   - 使用命令行客户端上传版本
   - 通过 UI 验证版本显示
3. 添加详细的步骤日志
4. 截图保存关键步骤

**验证**:
- 混合测试通过
- UI 和 API 集成正确

**预计时间**: 45 分钟

---

## 阶段 4: 构建和脚本

### 任务 4.1: 增强构建脚本

**目标**: 更新 build-all.bat 以创建完整的发布包

**文件**: `build-all.bat`

**步骤**:
1. 备份现有的 `build-all.bat`
2. 根据设计文档更新脚本：
   - 添加版本号参数支持
   - 添加创建发布目录的步骤
   - 添加复制客户端文件到 data/clients 的步骤
   - 添加创建 scripts/ 目录的步骤
   - 添加创建 ZIP 归档的步骤
3. 确保所有路径正确
4. 测试脚本运行

**验证**:
- 脚本执行成功
- 发布目录结构正确
- ZIP 文件创建成功

**预计时间**: 20 分钟

---

### 任务 4.2: 创建测试运行脚本

**目标**: 创建便捷的测试运行脚本

**文件**: `scripts/test-e2e.bat`

**步骤**:
1. 创建 `scripts/` 目录（如果不存在）
2. 创建 `scripts/test-e2e.bat` 文件
3. 实现以下功能：
   - 构建 test binaries（可选）
   - 安装 Playwright 浏览器
   - 运行 E2E 测试
   - 运行集成测试
   - 显示测试结果摘要
4. 添加错误处理和友好的输出

**验证**:
- 脚本可以运行所有测试
- 错误时正确退出
- 输出清晰

**预计时间**: 15 分钟

---

### 任务 4.3: 创建 CI/CD 配置（可选）

**目标**: 为 GitHub Actions 创建 CI 配置

**文件**: `.github/workflows/test.yml`

**步骤**:
1. 创建 `.github/workflows/` 目录
2. 创建 `test.yml` 文件
3. 配置以下 jobs：
   - unit-tests: 运行单元测试
   - integration-tests: 运行集成测试
   - e2e-tests: 运行 E2E 测试
4. 配置 Windows 运行环境

**验证**:
- 配置文件格式正确
- 可以在 GitHub Actions 中运行

**预计时间**: 15 分钟

---

## 验收标准

### 完整的测试套件

当所有任务完成后，应该能够运行：

```bash
# 运行所有测试
go test -v ./tests/...

# 运行集成测试
go test -v ./tests/integration/...

# 运行 E2E 测试
go test -v -tags=e2e ./tests/e2e/...

# 使用脚本运行
scripts\test-e2e.bat
```

### 关键验证点

1. **发布流程闭环**: `TestPublishLoop` 完整通过
2. **更新流程闭环**: `TestUpdateLoop` 完整通过
3. **E2E 测试**: Playwright 测试可以运行浏览器
4. **构建脚本**: 可以创建完整的发布包

---

## 风险和注意事项

1. **Playwright 浏览器安装**: 需要网络连接下载浏览器
2. **Windows 路径**: 注意 Windows 路径分隔符
3. **端口冲突**: 测试服务器使用固定端口可能冲突
4. **临时文件清理**: 确保测试后正确清理临时文件
5. **客户端工具**: 闭环测试需要 update-publisher.exe 和 update-client.exe 存在

---

## 时间估算

| 阶段 | 任务数 | 预计时间 |
|-----|--------|---------|
| 阶段 1: 基础设施 | 5 | 70 分钟 |
| 阶段 2: 集成测试 | 6 | 180 分钟 |
| 阶段 3: E2E 测试 | 3 | 110 分钟 |
| 阶段 4: 构建和脚本 | 3 | 50 分钟 |
| **总计** | **17** | **410 分钟 (~7 小时)** |

---

## 执行顺序建议

1. 先完成阶段 1（基础设施），确保测试框架可用
2. 再完成阶段 2（集成测试），验证核心功能
3. 然后完成阶段 3（E2E 测试），验证 UI 流程
4. 最后完成阶段 4（构建和脚本），准备发布
