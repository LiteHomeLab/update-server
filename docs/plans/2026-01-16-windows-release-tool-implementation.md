# Windows 发布工具实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** 创建一个独立的 Windows 命令行发布工具 `update-admin.exe`，支持环境变量配置、上传进度显示、SHA256 校验和自动验证。

**Architecture:** 基于 Cobra 框架的命令行工具，配置采用环境变量 + 参数覆盖策略，HTTP 客户端与服务器 API 交互，支持文件上传、下载验证和版本管理。

**Tech Stack:** Go 1.23, Cobra (CLI), net/http, crypto/sha256, testing (标准库)

---

## Task 0: 创建目录结构和 go.mod

**Files:**
- Create: `clients/go/tool/go.mod`
- Create: `clients/go/tool/config.go`
- Create: `clients/go/tool/utils.go`
- Create: `clients/go/tool/admin.go`
- Create: `clients/go/tool/main.go`
- Test: `clients/go/tool/admin_test.go`
- Test: `clients/go/tool/utils_test.go`

**Step 1: 创建目录结构**

Run:
```bash
cd clients/go && mkdir -p tool
```

**Step 2: 创建 go.mod**

```go
module github.com/LiteHomeLab/update-admin

go 1.23

require github.com/spf13/cobra v1.8.0
```

**Step 3: 初始化模块**

Run:
```bash
cd clients/go/tool && go mod tidy
```

**Step 4: 创建空文件**

Run:
```bash
cd clients/go/tool && touch config.go utils.go admin.go main.go admin_test.go utils_test.go
```

**Step 5: 提交**

Run:
```bash
git add clients/go/tool/go.mod clients/go/tool/*.go
git commit -m "feat(tool): create directory structure and go.mod"
```

---

## Task 1: 实现配置管理 (config.go)

**Files:**
- Modify: `clients/go/tool/config.go`
- Test: `clients/go/tool/config_test.go`

**Step 1: 编写配置加载测试**

```go
package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// 设置环境变量
	os.Setenv("UPDATE_SERVER_URL", "http://test-server:8080")
	os.Setenv("UPDATE_TOKEN", "test-token")

	defer os.Unsetenv("UPDATE_SERVER_URL")
	defer os.Unsetenv("UPDATE_TOKEN")

	cfg := LoadConfig("http://override:9090", "override-token", "myapp")

	if cfg.ServerURL != "http://override:9090" {
		t.Errorf("Expected override URL, got %s", cfg.ServerURL)
	}
	if cfg.Token != "override-token" {
		t.Errorf("Expected override token, got %s", cfg.Token)
	}
	if cfg.ProgramID != "myapp" {
		t.Errorf("Expected program ID, got %s", cfg.ProgramID)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("UPDATE_SERVER_URL", "http://env-server:8080")
	os.Setenv("UPDATE_TOKEN", "env-token")

	defer os.Unsetenv("UPDATE_SERVER_URL")
	defer os.Unsetenv("UPDATE_TOKEN")

	cfg := LoadConfig("", "", "myapp")

	if cfg.ServerURL != "http://env-server:8080" {
		t.Errorf("Expected env URL, got %s", cfg.ServerURL)
	}
	if cfg.Token != "env-token" {
		t.Errorf("Expected env token, got %s", cfg.Token)
	}
}
```

**Step 2: 运行测试验证失败**

Run: `cd clients/go/tool && go test -v -run TestLoadConfig`
Expected: FAIL with "undefined: LoadConfig"

**Step 3: 实现配置加载**

```go
package main

import (
	"os"
)

type Config struct {
	ServerURL string
	Token     string
	ProgramID string
}

func LoadConfig(serverURL, token, programID string) *Config {
	cfg := &Config{
		ServerURL: serverURL,
		Token:     token,
		ProgramID: programID,
	}

	// 从环境变量读取默认值
	if cfg.ServerURL == "" {
		cfg.ServerURL = os.Getenv("UPDATE_SERVER_URL")
	}
	if cfg.Token == "" {
		cfg.Token = os.Getenv("UPDATE_TOKEN")
	}

	// ProgramID 必须通过参数指定，不使用环境变量

	return cfg
}
```

**Step 4: 运行测试验证通过**

Run: `cd clients/go/tool && go test -v -run TestLoadConfig`
Expected: PASS

**Step 5: 提交**

Run:
```bash
git add clients/go/tool/config.go clients/go/tool/config_test.go
git commit -m "feat(tool): add config management with env var support"
```

---

## Task 2: 实现 SHA256 计算工具 (utils.go)

**Files:**
- Modify: `clients/go/tool/utils.go`
- Test: `clients/go/tool/utils_test.go`

**Step 1: 编写 SHA256 测试**

```go
package main

import (
	"testing"
)

func TestCalculateSHA256(t *testing.T) {
	// 创建临时测试文件
	content := []byte("test content for sha256")
	tmpfile := "/tmp/test-sha256.bin"
	if err := os.WriteFile(tmpfile, content, 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile)

	hash, err := CalculateSHA256(tmpfile)
	if err != nil {
		t.Fatalf("Failed to calculate SHA256: %v", err)
	}

	// 已知的正确哈希值
	expected := "dfe60fc7c66f6e8e5531a687992a7bc642db798ddb0eac76931bde5bbd77e951"
	if hash != expected {
		t.Errorf("Expected hash %s, got %s", expected, hash)
	}
}

func TestCalculateSHA256NonExistentFile(t *testing.T) {
	_, err := CalculateSHA256("/tmp/nonexistent-file-12345.bin")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}
```

**Step 2: 运行测试验证失败**

Run: `cd clients/go/tool && go test -v -run TestCalculateSHA256`
Expected: FAIL with "undefined: CalculateSHA256"

**Step 3: 实现 SHA256 计算**

```go
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
)

func CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := hash.Seek(0, 0); err != nil {
		return "", err
	}

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
```

**Step 4: 修复 import 和测试**

更新 utils.go 添加缺失的 import：

```go
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
```

**Step 5: 运行测试验证通过**

Run: `cd clients/go/tool && go test -v -run TestCalculateSHA256`
Expected: PASS

**Step 6: 提交**

Run:
```bash
git add clients/go/tool/utils.go clients/go/tool/utils_test.go
git commit -m "feat(tool): add SHA256 calculation utility"
```

---

## Task 3: 实现 UpdateAdmin 核心 API 调用 (admin.go - 第1部分)

**Files:**
- Modify: `clients/go/tool/admin.go`
- Test: `clients/go/tool/admin_test.go`

**Step 1: 编写基础结构和构造函数测试**

```go
package main

import (
	"testing"
	"time"
)

func TestNewUpdateAdmin(t *testing.T) {
	admin := NewUpdateAdmin("http://localhost:8080", "test-token")

	if admin == nil {
		t.Fatal("Expected non-nil admin")
	}
	if admin.serverURL != "http://localhost:8080" {
		t.Errorf("Expected server URL, got %s", admin.serverURL)
	}
	if admin.token != "test-token" {
		t.Errorf("Expected token, got %s", admin.token)
	}
	if admin.client == nil {
		t.Error("Expected non-nil HTTP client")
	}
	if admin.client.Timeout != 60*time.Second {
		t.Errorf("Expected 60s timeout, got %v", admin.client.Timeout)
	}
}
```

**Step 2: 运行测试验证失败**

Run: `cd clients/go/tool && go test -v -run TestNewUpdateAdmin`
Expected: FAIL with "undefined: NewUpdateAdmin"

**Step 3: 实现基础结构和构造函数**

```go
package main

import (
	"net/http"
	"time"
)

type UpdateAdmin struct {
	serverURL string
	token     string
	client    *http.Client
}

func NewUpdateAdmin(serverURL, token string) *UpdateAdmin {
	return &UpdateAdmin{
		serverURL: serverURL,
		token:     token,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}
```

**Step 4: 运行测试验证通过**

Run: `cd clients/go/tool && go test -v -run TestNewUpdateAdmin`
Expected: PASS

**Step 5: 提交**

Run:
```bash
git add clients/go/tool/admin.go clients/go/tool/admin_test.go
git commit -m "feat(tool): add UpdateAdmin base structure"
```

---

## Task 4: 实现列表功能 (admin.go - ListVersions)

**Files:**
- Modify: `clients/go/tool/admin.go`
- Test: `clients/go/tool/admin_test.go`

**Step 1: 编写列表功能测试**

```go
func TestListVersions(t *testing.T) {
	// 这个测试需要实际的测试服务器
	// 先测试 URL 构造是否正确
	admin := NewUpdateAdmin("http://localhost:8080", "test-token")

	// 测试 URL 构造
	expectedURL := "http://localhost:8080/api/programs/myapp/versions?channel=stable"
	// 我们需要通过反射或日志验证 URL，这里先创建函数
}
```

**Step 2: 实现 ListVersions**

在 admin.go 中添加：

```go
type VersionInfo struct {
	Version     string    `json:"version"`
	Channel     string    `json:"channel"`
	FileName    string    `json:"fileName"`
	FileSize    int64     `json:"fileSize"`
	FileHash    string    `json:"fileHash"`
	ReleaseNotes string   `json:"releaseNotes"`
	PublishDate  time.Time `json:"publishDate"`
	Mandatory    bool      `json:"mandatory"`
}

func (a *UpdateAdmin) ListVersions(programID, channel string) ([]VersionInfo, error) {
	url := fmt.Sprintf("%s/api/programs/%s/versions?channel=%s", a.serverURL, programID, channel)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("list failed with status %d", resp.StatusCode)
	}

	var versions []VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, err
	}

	return versions, nil
}
```

**Step 3: 添加缺失的 import**

更新 admin.go 的 imports：

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)
```

**Step 4: 提交**

Run:
```bash
git add clients/go/tool/admin.go
git commit -m "feat(tool): add ListVersions method"
```

---

## Task 5: 实现删除功能 (admin.go - DeleteVersion)

**Files:**
- Modify: `clients/go/tool/admin.go`

**Step 1: 实现 DeleteVersion**

在 admin.go 中添加：

```go
func (a *UpdateAdmin) DeleteVersion(programID, version string) error {
	url := fmt.Sprintf("%s/api/programs/%s/versions/%s", a.serverURL, programID, version)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("delete failed with status %d", resp.StatusCode)
	}

	return nil
}
```

**Step 2: 提交**

Run:
```bash
git add clients/go/tool/admin.go
git commit -m "feat(tool): add DeleteVersion method"
```

---

## Task 6: 实现上传进度跟踪 (admin.go - UploadVersion 基础)

**Files:**
- Modify: `clients/go/tool/admin.go`

**Step 1: 定义进度回调类型**

在 admin.go 中添加：

```go
type UploadProgress struct {
	Uploaded  int64
	Total     int64
	Percentage float64
}

type ProgressCallback func(UploadProgress)
```

**Step 2: 实现带进度的上传函数**

在 admin.go 中添加：

```go
func (a *UpdateAdmin) UploadVersion(programID, channel, version, filePath, notes string, mandatory bool, callback ProgressCallback) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// 创建 multipart writer
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// 添加表单字段
	writer.WriteField("channel", channel)
	writer.WriteField("version", version)
	writer.WriteField("notes", notes)
	writer.WriteField("mandatory", fmt.Sprintf("%v", mandatory))

	// 创建文件字段
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	// 使用进度跟踪器包装 writer
	progressWriter := &progressWriter{
		writer:    part,
		total:     fileSize,
		callback:  callback,
	}
	progressWriter.writer = part

	// 复制文件数据
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	writer.Close()

	// 构造请求
	url := fmt.Sprintf("%s/api/programs/%s/versions", a.serverURL, programID)
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+a.token)

	// 发送请求
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}

	return nil
}

type progressWriter struct {
	writer   io.Writer
	total    int64
	written  int64
	callback ProgressCallback
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.written += int64(n)
	if pw.callback != nil {
		pw.callback(UploadProgress{
			Uploaded:    pw.written,
			Total:       pw.total,
			Percentage: float64(pw.written) / float64(pw.total) * 100,
		})
	}

	return n, nil
}
```

**Step 3: 添加缺失的 import**

更新 admin.go 的 imports：

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)
```

**Step 4: 提交**

Run:
```bash
git add clients/go/tool/admin.go
git commit -m "feat(tool): add UploadVersion with progress tracking"
```

---

## Task 7: 实现上传后验证功能 (admin.go - VerifyUpload)

**Files:**
- Modify: `clients/go/tool/admin.go`

**Step 1: 实现验证函数**

在 admin.go 中添加：

```go
func (a *UpdateAdmin) VerifyUpload(programID, channel, version, expectedHash string) error {
	// 获取版本详情
	url := fmt.Sprintf("%s/api/programs/%s/versions/%s/%s", a.serverURL, programID, channel, version)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("verification failed: status %d", resp.StatusCode)
	}

	var versionInfo VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&versionInfo); err != nil {
		return err
	}

	// 验证 SHA256
	if versionInfo.FileHash != expectedHash {
		return fmt.Errorf("SHA256 mismatch: expected %s, got %s", expectedHash, versionInfo.FileHash)
	}

	return nil
}
```

**Step 2: 提交**

Run:
```bash
git add clients/go/tool/admin.go
git commit -m "feat(tool): add VerifyUpload method"
```

---

## Task 8: 实现完整的上传流程 (admin.go - UploadVersionWithVerify)

**Files:**
- Modify: `clients/go/tool/admin.go`

**Step 1: 实现完整上传流程**

在 admin.go 中添加：

```go
func (a *UpdateAdmin) UploadVersionWithVerify(programID, channel, version, filePath, notes string, mandatory bool, callback ProgressCallback) error {
	// 1. 计算本地文件 SHA256
	localHash, err := CalculateSHA256(filePath)
	if err != nil {
		return fmt.Errorf("failed to calculate SHA256: %w", err)
	}

	// 2. 上传文件
	if err := a.UploadVersion(programID, channel, version, filePath, notes, mandatory, callback); err != nil {
		return err
	}

	// 3. 验证上传
	if err := a.VerifyUpload(programID, channel, version, localHash); err != nil {
		return fmt.Errorf("upload verification failed: %w", err)
	}

	return nil
}
```

**Step 2: 提交**

Run:
```bash
git add clients/go/tool/admin.go
git commit -m "feat(tool): add UploadVersionWithVerify with SHA256 verification"
```

---

## Task 9: 实现 CLI 主程序 (main.go)

**Files:**
- Modify: `clients/go/tool/main.go`

**Step 1: 实现 main.go**

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	cfgServerURL   string
	cfgToken       string
	cfgProgramID   string
)

var rootCmd = &cobra.Command{
	Use:   "update-admin",
	Short: "DocuFiller Update Server Admin Tool",
	Long:  `A command-line tool for managing program versions on the update server.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgServerURL, "server", "", "Server URL (overrides UPDATE_SERVER_URL env var)")
	rootCmd.PersistentFlags().StringVar(&cfgToken, "token", "", "API token (overrides UPDATE_TOKEN env var)")
	rootCmd.PersistentFlags().StringVar(&cfgProgramID, "program-id", "", "Program ID (required)")
	rootCmd.MarkPersistentFlagRequired("program-id")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 2: 提交**

Run:
```bash
git add clients/go/tool/main.go
git commit -m "feat(tool): add CLI base structure"
```

---

## Task 10: 实现 upload 命令

**Files:**
- Modify: `clients/go/tool/main.go`

**Step 1: 实现 upload 命令**

在 main.go 中添加（在 init() 之前）：

```go
var (
	uploadChannel   string
	uploadVersion   string
	uploadFile      string
	uploadNotes     string
	uploadMandatory bool
)

var uploadCmd = &cobra.Command{
	Use:   "upload --channel <stable|beta> --version <version> --file <path>",
	Short: "Upload a new version",
	Long:  `Upload a new version to the update server with progress display and automatic verification.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 加载配置
		cfg := LoadConfig(cfgServerURL, cfgToken, cfgProgramID)

		// 验证必需参数
		if uploadChannel == "" || uploadVersion == "" || uploadFile == "" {
			return fmt.Errorf("--channel, --version, and --file are required")
		}

		// 验证文件存在
		if _, err := os.Stat(uploadFile); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", uploadFile)
		}

		// 创建 admin 客户端
		admin := NewUpdateAdmin(cfg.ServerURL, cfg.Token)

		// 定义进度回调
		progressCallback := func(p UploadProgress) {
			fmt.Printf("\rProgress: %.1f%% (%d/%d bytes)", p.Percentage, p.Uploaded, p.Total)
		}

		fmt.Printf("Uploading %s/%s/%s...\n", cfgProgramID, uploadChannel, uploadVersion)

		// 执行上传
		if err := admin.UploadVersionWithVerify(cfg.ProgramID, uploadChannel, uploadVersion, uploadFile, uploadNotes, uploadMandatory, progressCallback); err != nil {
			fmt.Println() // 换行
			return err
		}

		fmt.Println("\n✓ Upload successful!")
		return nil
	},
}

func init() {
	uploadCmd.Flags().StringVar(&uploadChannel, "channel", "", "Channel (stable/beta)")
	uploadCmd.Flags().StringVar(&uploadVersion, "version", "", "Version number")
	uploadCmd.Flags().StringVar(&uploadFile, "file", "", "File path")
	uploadCmd.Flags().StringVar(&uploadNotes, "notes", "", "Release notes")
	uploadCmd.Flags().BoolVar(&uploadMandatory, "mandatory", false, "Mandatory update")

	uploadCmd.MarkFlagRequired("channel")
	uploadCmd.MarkFlagRequired("version")
	uploadCmd.MarkFlagRequired("file")

	rootCmd.AddCommand(uploadCmd)
}
```

**Step 2: 提交**

Run:
```bash
git add clients/go/tool/main.go
git commit -m "feat(tool): add upload command with progress display"
```

---

## Task 11: 实现 list 命令

**Files:**
- Modify: `clients/go/tool/main.go`

**Step 1: 实现 list 命令**

在 main.go 中添加：

```go
var (
	listChannel string
)

var listCmd = &cobra.Command{
	Use:   "list [--channel <stable|beta>]",
	Short: "List versions",
	Long:  `List all versions for a program, optionally filtered by channel.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := LoadConfig(cfgServerURL, cfgToken, cfgProgramID)
		admin := NewUpdateAdmin(cfg.ServerURL, cfg.Token)

		versions, err := admin.ListVersions(cfg.ProgramID, listChannel)
		if err != nil {
			return err
		}

		if len(versions) == 0 {
			fmt.Println("No versions found")
			return nil
		}

		// 打印表格头部
		fmt.Println("Version\tChannel\tSize\t\tDate\t\tMandatory")
		fmt.Println("-------\t-------\t----\t\t----\t\t---------")

		for _, v := range versions {
			sizeMB := float64(v.FileSize) / 1024 / 1024
			date := v.PublishDate.Format("2006-01-02")
			mandatory := "No"
			if v.Mandatory {
				mandatory = "Yes"
			}
			fmt.Printf("%s\t%s\t%.2f MB\t%s\t%s\n", v.Version, v.Channel, sizeMB, date, mandatory)
		}

		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listChannel, "channel", "", "Channel filter (stable/beta)")

	rootCmd.AddCommand(listCmd)
}
```

**Step 2: 提交**

Run:
```bash
git add clients/go/tool/main.go
git commit -m "feat(tool): add list command with table output"
```

---

## Task 12: 实现 delete 命令

**Files:**
- Modify: `clients/go/tool/main.go`

**Step 1: 实现 delete 命令**

在 main.go 中添加：

```go
var (
	deleteVersion string
)

var deleteCmd = &cobra.Command{
	Use:   "delete --version <version>",
	Short: "Delete a version",
	Long:  `Delete a specific version from the update server.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := LoadConfig(cfgServerURL, cfgToken, cfgProgramID)

		// 验证必需参数
		if deleteVersion == "" {
			return fmt.Errorf("--version is required")
		}

		admin := NewUpdateAdmin(cfg.ServerURL, cfg.Token)

		fmt.Printf("Deleting %s/%s...\n", cfgProgramID, deleteVersion)

		if err := admin.DeleteVersion(cfg.ProgramID, deleteVersion); err != nil {
			return err
		}

		fmt.Println("✓ Delete successful!")
		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVar(&deleteVersion, "version", "", "Version number")
	deleteCmd.MarkFlagRequired("version")

	rootCmd.AddCommand(deleteCmd)
}
```

**Step 2: 提交**

Run:
```bash
git add clients/go/tool/main.go
git commit -m "feat(tool): add delete command"
```

---

## Task 13: 编译并测试工具

**Files:**
- Create: `clients/go/tool/README.md`

**Step 1: 编译 Windows 可执行文件**

Run:
```bash
cd clients/go/tool && go build -ldflags "-s -w" -o update-admin.exe .
```

**Step 2: 验证可执行文件生成**

Run:
```bash
cd clients/go/tool && ls -lh update-admin.exe
```

Expected: 文件大小约 3-5 MB

**Step 3: 测试帮助信息**

Run:
```bash
cd clients/go/tool && ./update-admin.exe --help
```

Expected: 显示命令帮助

**Step 4: 提交**

Run:
```bash
git add clients/go/tool/update-admin.exe
git commit -m "build(tool): compile Windows executable"
```

---

## Task 14: 启动测试服务器

**Step 1: 启动服务器**

Run:
```bash
cd /c/WorkSpace/Go2Hell/src/github.com/LiteHomeLab/update-server && go run main.go
```

Background: 运行在后台，使用 port 8080

**Step 2: 验证服务器运行**

Run:
```bash
curl http://localhost:8080/api/health
```

Expected: `{"status":"ok"}`

---

## Task 15: 集成测试 - 完整流程

**Step 1: 设置环境变量**

Run:
```bash
export UPDATE_SERVER_URL="http://localhost:8080"
export UPDATE_TOKEN="test-token"
```

**Step 2: 创建测试程序和 token**

Run:
```bash
# 创建测试程序
curl -X POST http://localhost:8080/api/programs \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{"programId":"testapp","name":"Test App","description":"Test application"}'

# 创建测试 token
curl -X POST http://localhost:8080/api/tokens \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{"name":"test-upload","permissions":["upload"],"programId":"testapp"}'
```

**Step 3: 创建测试文件**

Run:
```bash
dd if=/dev/zero of=/tmp/test-v1.0.0.zip bs=1M count=10
```

**Step 4: 测试上传**

Run:
```bash
cd clients/go/tool && ./update-admin.exe upload --program-id testapp --channel stable --version 1.0.0 --file /tmp/test-v1.0.0.zip --notes "Initial release"
```

Expected: 显示进度并成功

**Step 5: 测试列表**

Run:
```bash
cd clients/go/tool && ./update-admin.exe list --program-id testapp --channel stable
```

Expected: 显示版本 1.0.0

**Step 6: 测试删除**

Run:
```bash
cd clients/go/tool && ./update-admin.exe delete --program-id testapp --version 1.0.0
```

Expected: 删除成功

---

## Task 16: 编写使用文档

**Files:**
- Modify: `clients/go/tool/README.md`

**Step 1: 编写 README.md**

```markdown
# Update Admin Tool

用于 DocuFiller 更新服务器的命令行管理工具。

## 编译

```bash
go build -ldflags "-s -w" -o update-admin.exe .
```

## 配置

### 环境变量（推荐）

```cmd
setx UPDATE_SERVER_URL "http://your-server:8080"
setx UPDATE_TOKEN "your-api-token"
```

### 命令行参数

- `--server url`：服务器地址（覆盖环境变量）
- `--token value`：认证令牌（覆盖环境变量）
- `--program-id id`：程序标识符（**必须指定**）

## 使用示例

### 上传新版本

```cmd
update-admin.exe upload --program-id myapp --channel stable --version 1.0.0 --file myapp.zip --notes "Initial release"
```

### 列出版本

```cmd
update-admin.exe list --program-id myapp --channel stable
```

### 删除版本

```cmd
update-admin.exe delete --program-id myapp --version 1.0.0
```

## 命令参数

### upload 命令

| 参数 | 说明 |
|------|------|
| `--channel` | 发布通道（stable/beta） |
| `--version` | 版本号 |
| `--file` | 文件路径 |
| `--notes` | 发布说明（可选） |
| `--mandatory` | 强制更新标记（可选） |

### list 命令

| 参数 | 说明 |
|------|------|
| `--channel` | 通道过滤（可选） |

### delete 命令

| 参数 | 说明 |
|------|------|
| `--version` | 版本号 |

## CI/CD 集成

### GitHub Actions 示例

```yaml
- name: Upload Release
  run: |
    ./update-admin.exe upload \
      --program-id myapp \
      --channel stable \
      --version ${{ github.ref_name }} \
      --file ./dist/myapp.zip \
      --notes "Release ${{ github.ref_name }}"
  env:
    UPDATE_SERVER_URL: ${{ secrets.UPDATE_SERVER_URL }}
    UPDATE_TOKEN: ${{ secrets.UPDATE_TOKEN }}
```
```

**Step 2: 提交文档**

Run:
```bash
git add clients/go/tool/README.md
git commit -m "docs(tool): add usage documentation"
```

---

## Task 17: 最终测试和验证

**Step 1: 停止测试服务器**

如果测试服务器仍在运行，停止它。

**Step 2: 清理测试数据**

Run:
```bash
rm -rf /tmp/test-*.zip
```

**Step 3: 检查所有代码**

Run:
```bash
cd clients/go/tool && go fmt ./...
go vet ./...
```

**Step 4: 运行所有测试**

Run:
```bash
cd clients/go/tool && go test -v ./...
```

Expected: 所有测试通过

**Step 5: 最终提交**

Run:
```bash
git add -A
git commit -m "feat(tool): complete Windows release tool implementation"
```

---

## 任务完成检查清单

- [ ] 配置管理支持环境变量 + 参数覆盖
- [ ] SHA256 文件校验功能
- [ ] 上传进度实时显示
- [ ] 上传后自动验证
- [ ] 列表命令带表格输出
- [ ] 删除命令
- [ ] 编译生成独立的 update-admin.exe
- [ ] 完整的使用文档
- [ ] 集成测试通过

## 预期产物

1. `clients/go/tool/update-admin.exe` - 独立可执行文件
2. `clients/go/tool/README.md` - 使用文档
3. 完整的测试覆盖

## 后续步骤

完成实现后，可以：
1. 将 `update-admin.exe` 复制到其他项目使用
2. 在 CI/CD 流程中集成
3. 根据需要添加更多功能（如批量操作）
