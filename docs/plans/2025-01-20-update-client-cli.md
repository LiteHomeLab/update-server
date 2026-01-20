# Update Client CLI Tool Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a standalone CLI tool (`update-client.exe`) that any application can integrate to check for and download updates from the update server.

**Architecture:** Command-line tool with two commands (`check` and `download`), reads configuration from YAML file, outputs JSON for machine parsing, supports automatic decryption of encrypted downloads.

**Tech Stack:** Go 1.23, Cobra CLI framework, YAML config, AES decryption for encrypted packages.

---

## Overview

This implementation creates a new standalone CLI tool that can be distributed with any application. The tool communicates with the update server via HTTP APIs, handles encrypted downloads, and returns structured JSON results for easy integration.

**Key Design Decisions:**
- Configuration via `update-config.yaml` in the same directory as the exe
- Two commands: `check` (compare versions) and `download` (fetch update)
- Dual output modes: human-friendly text (default) and JSON (`--json` flag)
- Automatic decryption when encryption key is present in config
- Server-side pre-configuration: the update server's "Download Update Client" feature generates ready-to-use config files

---

## Task 1: Create CLI Entry Point with Cobra

**Files:**
- Create: `clients/go/client/main.go`

**Step 1: Write the basic CLI structure**

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	jsonOutput bool
)

var rootCmd = &cobra.Command{
	Use:   "update-client",
	Short: "Update Client - Check and download updates from update server",
	Long:  `A command-line tool for checking and downloading application updates from the update server.`,
}

var checkCmd = &cobra.Command{
	Use:   "check [--current-version VERSION]",
	Short: "Check for updates",
	Long:  `Check if a new version is available on the update server.`,
	RunE:  runCheck,
}

var downloadCmd = &cobra.Command{
	Use:   "download --version VERSION [--output PATH]",
	Short: "Download an update",
	Long:  `Download a specific version from the update server.`,
	RunE:  runDownload,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "update-config.yaml", "config file")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output JSON format")

	downloadCmd.Flags().String("output", "", "output file path")
	downloadCmd.Flags().String("version", "", "version to download")
	downloadCmd.MarkFlagRequired("version")

	rootCmd.AddCommand(checkCmd, downloadCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Implementation will be added in Task 3
	fmt.Println("check command - to be implemented")
	return nil
}

func runDownload(cmd *cobra.Command, args []string) error {
	// Implementation will be added in Task 4
	fmt.Println("download command - to be implemented")
	return nil
}
```

**Step 2: Update go.mod with cobra dependency**

Run: `cd clients/go/client && go get github.com/spf13/cobra@latest`
Expected: Dependency added to go.mod and go.sum

**Step 3: Verify it compiles**

Run: `cd clients/go/client && go build -o ../../bin/update-client.exe .`
Expected: Binary created at `bin/update-client.exe`

**Step 4: Test basic commands**

Run: `bin/update-client.exe --help`
Expected: Help text displayed

Run: `bin/update-client.exe check`
Expected: "check command - to be implemented"

**Step 5: Commit**

```bash
git add clients/go/client/main.go clients/go/client/go.mod clients/go/client/go.sum
git commit -m "feat: add update-client CLI entry point with cobra"
```

---

## Task 2: Implement Config File Loading

**Files:**
- Modify: `clients/go/client/config.go`
- Create: `clients/go/client/config_test.go`

**Step 1: Write the config structure and loader**

```go
package client

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 客户端配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Program  ProgramConfig  `yaml:"program"`
	Auth     AuthConfig     `yaml:"auth"`
	Download DownloadConfig `yaml:"download"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	URL     string `yaml:"url"`
	Timeout int    `yaml:"timeout"`
}

type ProgramConfig struct {
	ID             string `yaml:"id"`
	CurrentVersion string `yaml:"current_version"`
}

type AuthConfig struct {
	Token         string `yaml:"token"`
	EncryptionKey string `yaml:"encryption_key"`
}

type DownloadConfig struct {
	SavePath    string `yaml:"save_path"`
	Naming      string `yaml:"naming"`       // version | date | simple
	Keep        int    `yaml:"keep"`         // for date mode
	AutoVerify  bool   `yaml:"auto_verify"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			URL:     "http://localhost:8080",
			Timeout: 30,
		},
		Program: ProgramConfig{
			ID: "",
		},
		Download: DownloadConfig{
			SavePath:   "./updates",
			Naming:     "version",
			Keep:       3,
			AutoVerify: true,
		},
		Logging: LoggingConfig{
			Level: "info",
			File:  "update-client.log",
		},
	}
}

// LoadConfig loads configuration from YAML file
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set timeout
	if cfg.Server.Timeout > 0 {
		cfg.Timeout = time.Duration(cfg.Server.Timeout) * time.Second
	}

	return cfg, nil
}

// GetTimeout returns the timeout duration
func (c *Config) GetTimeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 30 * time.Second
}
```

**Step 2: Write tests for config loading**

Create `clients/go/client/config_test.go`:

```go
package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Default(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.Server.URL != "http://localhost:8080" {
		t.Errorf("Expected default URL, got %s", cfg.Server.URL)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  url: "http://test-server:9000"
  timeout: 60
program:
  id: "test-app"
  current_version: "1.0.0"
auth:
  token: "test-token"
download:
  save_path: "./test-updates"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.Server.URL != "http://test-server:9000" {
		t.Errorf("Expected test-server URL, got %s", cfg.Server.URL)
	}

	if cfg.Program.ID != "test-app" {
		t.Errorf("Expected test-app, got %s", cfg.Program.ID)
	}
}
```

**Step 3: Run tests to verify they pass**

Run: `cd clients/go/client && go test -v ./...`
Expected: All tests pass

**Step 4: Update go.mod with yaml dependency**

Run: `cd clients/go/client && go get gopkg.in/yaml.v3`
Expected: Dependency added

**Step 5: Commit**

```bash
git add clients/go/client/config.go clients/go/client/config_test.go
git commit -m "feat: add YAML config loading for update-client"
```

---

## Task 3: Implement Check Command

**Files:**
- Modify: `clients/go/client/main.go`
- Modify: `clients/go/client/checker.go` (existing, enhance)

**Step 1: Update checker.go to support JSON output**

```go
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// UpdateChecker 更新检查器
type UpdateChecker struct {
	config     *Config
	httpClient *http.Client
	jsonOutput bool
}

// NewUpdateChecker 创建更新检查器
func NewUpdateChecker(config *Config, jsonOutput bool) *UpdateChecker {
	return &UpdateChecker{
		config:     config,
		jsonOutput: jsonOutput,
		httpClient: &http.Client{
			Timeout: config.GetTimeout(),
		},
	}
}

// CheckResult 检查结果
type CheckResult struct {
	HasUpdate      bool   `json:"hasUpdate"`
	CurrentVersion string `json:"currentVersion,omitempty"`
	LatestVersion  string `json:"latestVersion"`
	DownloadURL    string `json:"downloadUrl,omitempty"`
	FileSize       int64  `json:"fileSize,omitempty"`
	ReleaseNotes   string `json:"releaseNotes,omitempty"`
	PublishDate    string `json:"publishDate,omitempty"`
	Mandatory      bool   `json:"mandatory"`
}

// Check 检查更新并输出结果
func (c *UpdateChecker) Check(currentVersion string) error {
	info, err := c.CheckUpdate(currentVersion)
	if err != nil {
		return c.outputError(err)
	}

	return c.outputResult(info)
}

// CheckUpdate 检查是否有新版本（internal method）
func (c *UpdateChecker) CheckUpdate(currentVersion string) (*UpdateInfo, error) {
	url := fmt.Sprintf("%s/api/programs/%s/versions/latest?channel=%s",
		c.config.Server.URL, c.config.Program.ID, "stable") // TODO: support channel

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, &UpdateError{
			Code:    "NETWORK_ERROR",
			Message: fmt.Sprintf("Failed to connect to server: %v", err),
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &UpdateError{
			Code:    "NO_VERSION",
			Message: "No version found for this program",
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &UpdateError{
			Code:    "SERVER_ERROR",
			Message: fmt.Sprintf("Server returned status %d", resp.StatusCode),
		}
	}

	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, &UpdateError{
			Code:    "PARSE_ERROR",
			Message: "Failed to parse response",
			Err:     err,
		}
	}

	return &info, nil
}

// outputResult 输出检查结果
func (c *UpdateChecker) outputResult(info *UpdateInfo) error {
	if c.jsonOutput {
		result := &CheckResult{
			HasUpdate:     true, // TODO: compare versions
			LatestVersion: info.Version,
			FileSize:      info.FileSize,
			ReleaseNotes:  info.ReleaseNotes,
			PublishDate:   info.PublishDate.Format(time.RFC3339),
			Mandatory:     info.Mandatory,
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	// Human-readable output
	fmt.Printf("✓ Connected to update server\n")
	fmt.Printf("  Latest version: %s\n", info.Version)
	fmt.Printf("\n  Version details:\n")
	fmt.Printf("    Size: %.1f MB\n", float64(info.FileSize)/1024/1024)
	fmt.Printf("    Published: %s\n", info.PublishDate.Format("2006-01-02"))
	fmt.Printf("    Mandatory: %s\n", map[bool]string{true: "Yes", false: "No"}[info.Mandatory])
	if info.ReleaseNotes != "" {
		fmt.Printf("    Release notes:\n      %s\n", info.ReleaseNotes)
	}

	return nil
}

// outputError 输出错误信息
func (c *UpdateChecker) outputError(err error) error {
	if c.jsonOutput {
		result := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	// Human-readable error
	fmt.Printf("✗ Failed to check for updates\n")
	if ue, ok := err.(*UpdateError); ok {
		fmt.Printf("  Error: %s\n", ue.Message)
	} else {
		fmt.Printf("  Error: %v\n", err)
	}

	return err
}
```

**Step 2: Update main.go to use the checker**

Replace the `runCheck` function:

```go
func runCheck(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := client.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current version from flag or config
	currentVersion, _ := cmd.Flags().GetString("current-version")
	if currentVersion == "" {
		currentVersion = cfg.Program.CurrentVersion
	}

	// Create checker and run
	checker := client.NewUpdateChecker(cfg, jsonOutput)
	return checker.Check(currentVersion)
}
```

Add the flag to init():

```go
func init() {
	// ... existing flags ...

	checkCmd.Flags().String("current-version", "", "current version (overrides config)")

	// ... rest of init ...
}
```

**Step 3: Test the check command**

Run: `bin/update-client.exe check`
Expected: Either connection error (if server not running) or version info

**Step 4: Commit**

```bash
git add clients/go/client/main.go clients/go/client/checker.go
git commit -m "feat: implement check command with JSON output"
```

---

## Task 4: Implement Download Command

**Files:**
- Modify: `clients/go/client/main.go`
- Modify: `clients/go/client/downloader.go` (existing, enhance)

**Step 1: Update downloader.go to support JSON output**

Add to `downloader.go`:

```go
// DownloadResult 下载结果
type DownloadResult struct {
	Success  bool   `json:"success"`
	File     string `json:"file"`
	FileSize int64  `json:"fileSize"`
	Verified bool   `json:"verified"`
	Decrypted bool  `json:"decrypted"`
}

// DownloadWithOutput 下载更新并输出结果
func (c *UpdateChecker) DownloadWithOutput(version string, outputPath string) error {
	if outputPath == "" {
		outputPath = c.generateOutputPath(version)
	}

	// Download
	if err := c.DownloadUpdate(version, outputPath, c.progressCallback); err != nil {
		return c.outputError(err)
	}

	// Verify
	info, err := c.CheckUpdate("")
	if err == nil {
		verified, _ := c.VerifyFile(outputPath, info.FileHash)
		if !verified {
			return c.outputError(fmt.Errorf("verification failed"))
		}
	}

	// Decrypt if key is available
	decrypted := false
	if c.config.Auth.EncryptionKey != "" {
		decryptor, err := NewDecryptor(c.config.Auth.EncryptionKey)
		if err == nil {
			decryptedPath := outputPath
			if err := decryptor.DecryptFile(outputPath, decryptedPath); err == nil {
				decrypted = true
			}
		}
	}

	return c.outputDownloadResult(outputPath, decrypted)
}

func (c *UpdateChecker) generateOutputPath(version string) string {
	baseName := "app"
	if c.config.Program.ID != "" {
		baseName = c.config.Program.ID
	}

	switch c.config.Download.Naming {
	case "version":
		return fmt.Sprintf("%s/%s-v%s.zip", c.config.Download.SavePath, baseName, version)
	case "date":
		return fmt.Sprintf("%s/%s-%s.zip", c.config.Download.SavePath, baseName, time.Now().Format("2006-01-02"))
	default:
		return fmt.Sprintf("%s/%s.zip", c.config.Download.SavePath, baseName)
	}
}

func (c *UpdateChecker) progressCallback(progress DownloadProgress) {
	if c.jsonOutput {
		return // No progress in JSON mode
	}

	fmt.Printf("\r  Progress: [%-20s] %.1f%% (%.1f/%.1f MB) - %.1f MB/s",
		strings.Repeat("=", int(progress.Percentage/5)),
		progress.Percentage,
		float64(progress.Downloaded)/1024/1024,
		float64(progress.Total)/1024/1024,
		progress.Speed/1024/1024)
}

func (c *UpdateChecker) outputDownloadResult(filePath string, decrypted bool) error {
	if c.jsonOutput {
		result := &DownloadResult{
			Success:   true,
			File:      filePath,
			Verified:  true,
			Decrypted: decrypted,
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	fmt.Println() // New line after progress
	fmt.Printf("✓ Download completed: %s\n", filePath)
	fmt.Printf("✓ Verified: SHA256 matches\n")
	if decrypted {
		fmt.Printf("✓ Decrypted: file ready to use\n")
	}

	return nil
}
```

**Step 2: Update main.go to use the downloader**

Replace the `runDownload` function:

```go
func runDownload(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := client.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get version flag
	version, _ := cmd.Flags().GetString("version")
	outputPath, _ := cmd.Flags().GetString("output")

	// Create checker and run
	checker := client.NewUpdateChecker(cfg, jsonOutput)
	return checker.DownloadWithOutput(version, outputPath)
}
```

**Step 3: Commit**

```bash
git add clients/go/client/main.go clients/go/client/downloader.go
git commit -m "feat: implement download command with progress and verification"
```

---

## Task 5: Implement Decryption

**Files:**
- Create: `clients/go/client/decryptor.go`
- Create: `clients/go/client/decryptor_test.go`

**Step 1: Write the decryptor implementation**

```go
package client

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// Decryptor 文件解密器
type Decryptor struct {
	key []byte
}

// NewDecryptor 创建解密器
func NewDecryptor(base64Key string) (*Decryptor, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: expected 32 bytes, got %d", len(key))
	}

	return &Decryptor{key: key}, nil
}

// DecryptFile 解密文件（CTR 模式）
func (d *Decryptor) DecryptFile(srcPath, dstPath string) error {
	// 读取加密文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// 创建输出文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// 创建 AES 解密器
	block, err := aes.NewCipher(d.key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// 读取 IV（前 16 字节）
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(srcFile, iv); err != nil {
		return fmt.Errorf("failed to read IV: %w", err)
	}

	// 创建流解密器
	stream := cipher.NewCTR(block, iv)
	reader := &cipher.StreamReader{S: stream, R: srcFile}

	// 解密并写入
	if _, err := io.Copy(dstFile, reader); err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	return nil
}
```

**Step 2: Write unit tests**

```go
package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDecryptor_ValidKey(t *testing.T) {
	// 32-byte key encoded in base64
	key := "dGVzdGtleWZvcmVuY3J5cHRpb25zdGVzdGtleWZvcmVuY3J5cHRpb24="
	_, err := NewDecryptor(key)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestNewDecryptor_InvalidKey(t *testing.T) {
	_, err := NewDecryptor("invalid-base64!!!")
	if err == nil {
		t.Fatal("Expected error for invalid base64")
	}
}

func TestNewDecryptor_WrongLength(t *testing.T) {
	// Too short
	shortKey := base64.StdEncoding.EncodeToString([]byte("short"))
	_, err := NewDecryptor(shortKey)
	if err == nil {
		t.Fatal("Expected error for wrong key length")
	}
}
```

**Step 3: Run tests**

Run: `cd clients/go/client && go test -v ./...`
Expected: All tests pass

**Step 4: Commit**

```bash
git add clients/go/client/decryptor.go clients/go/client/decryptor_test.go
git commit -m "feat: add file decryption support for encrypted updates"
```

---

## Task 6: Update Build Script

**Files:**
- Modify: `build-all.bat`

**Step 1: Add update-client build step**

Find the section that builds update-admin and add:

```batch
echo [4/5] Building Update Client...
cd /d "%SCRIPT_DIR%clients\go\client"
go build -o "%CLIENT_OUTPUT_DIR%\update-client.exe" .
if errorlevel 1 (
    echo ERROR: Failed to build update-client.exe
    goto :error
)
echo SUCCESS: Built update-client.exe
echo.

echo [5/5] Copying client executables to server data directory...
set "SERVER_CLIENT_DIR=%SCRIPT_DIR%data\clients"
if not exist "%SERVER_CLIENT_DIR%" mkdir "%SERVER_CLIENT_DIR%"

REM Copy publish client
copy /Y "%CLIENT_OUTPUT_DIR%\update-admin.exe" "%SERVER_CLIENT_DIR%\publish-client.exe" >nul
echo Copied: publish-client.exe

REM Copy update client
copy /Y "%CLIENT_OUTPUT_DIR%\update-client.exe" "%SERVER_CLIENT_DIR%\update-client.exe" >nul
echo Copied: update-client.exe
```

Update the final output message:

```batch
echo Output files:
echo   - Server: %OUTPUT_DIR%\update-server.exe
echo   - Publish Client: %CLIENT_OUTPUT_DIR%\update-admin.exe
echo   - Update Client: %CLIENT_OUTPUT_DIR%\update-client.exe
```

**Step 2: Test the build script**

Run: `build-all.bat`
Expected: All three binaries built successfully

**Step 3: Commit**

```bash
git add build-all.bat
git commit -m "build: add update-client to build-all script"
```

---

## Task 7: Update Server Client Packager

**Files:**
- Modify: `internal/service/client_packager.go`

**Step 1: Update GenerateUpdateClient to include the exe**

Find the section that copies the update client and update:

```go
// 复制更新客户端可执行文件（如果存在）
updateClientSrc := filepath.Join("./data/clients", "update-client.exe")
if _, err := os.Stat(updateClientSrc); err == nil {
	updateClientDst := filepath.Join(tempDir, "update-client.exe")
	if err := copyFile(updateClientSrc, updateClientDst); err != nil {
		// Log warning but don't fail - exe might not be built yet
		fmt.Printf("Warning: failed to copy update-client.exe: %v\n", err)
	}
}
```

**Step 2: Test the download functionality**

1. Start the server
2. Create a program
3. Download the update client package
4. Verify it contains `update-client.exe`

**Step 3: Commit**

```bash
git add internal/service/client_packager.go
git commit -m "fix: include update-client.exe in client packages"
```

---

## Task 8: Remove Unused SDKs

**Files:**
- Delete: `clients/csharp/`
- Delete: `clients/python/`

**Step 1: Remove C# SDK**

Run: `rm -rf clients/csharp/`

**Step 2: Remove Python SDK**

Run: `rm -rf clients/python/`

**Step 3: Commit**

```bash
git add -A
git commit -m "refactor: remove C# and Python SDKs, using Go CLI only"
```

---

## Task 9: Update Documentation

**Files:**
- Modify: `docs/README.md`
- Create: `docs/UPDATE_CLIENT_GUIDE.md`

**Step 1: Update main README**

Update the client download section to reflect the new CLI tool.

**Step 2: Create integration guide** (see separate document below)

**Step 3: Commit**

```bash
git add docs/README.md docs/UPDATE_CLIENT_GUIDE.md
git commit -m "docs: add update-client integration guide"
```

---

## Task 10: Final Testing

**Step 1: End-to-end test**

1. Run `build-all.bat`
2. Start server
3. Create a test program
4. Download and extract the update client package
5. Run `update-client.exe check`
6. Run `update-client.exe download --version X.Y.Z`
7. Verify the downloaded file

**Step 2: Test JSON output**

```bash
update-client.exe check --json
update-client.exe download --version 1.0.0 --json
```

**Step 3: Test with real application**

Integrate with a sample application and verify the full workflow.

**Step 4: Final commit**

```bash
git commit -m "test: verify update-client CLI functionality"
```

---

## Implementation Order

Execute tasks in order: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9 → 10

Each task should be completed and committed before moving to the next.

## Dependencies

- Go 1.23+
- Internet connection for downloading dependencies
- Update server running for testing

## Testing Strategy

1. Unit tests for config loading, decryption
2. Integration tests with running update server
3. Manual testing of CLI commands
4. End-to-end testing with sample application
