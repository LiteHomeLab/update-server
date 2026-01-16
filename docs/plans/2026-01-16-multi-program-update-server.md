# 多程序更新服务器实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将单程序 DocuFiller 更新服务升级为支持 20+ 程序的多程序自动更新平台，支持分级 Token 权限管理和 API 层加密传输。

**架构:** 采用分层架构（API 网关层 → 路由层 → 业务层 → 数据层），使用 SQLite 存储，Gin 框架实现 RESTful API，AES-256-GCM 实现 API 层加密。

**Tech Stack:** Go 1.21+, Gin, GORM, SQLite, AES-256-GCM, HKDF

---

## 前置准备

### Task 0: 环境准备

**目标:** 确保 Go 环境配置正确，依赖已安装。

**Step 1: 检查 Go 版本**

```bash
go version
```

Expected: `go version go1.21.x windows/amd64` 或更高版本

**Step 2: 安装依赖**

```bash
go mod tidy
go mod download
```

Expected: 无错误输出

**Step 3: 验证项目可以运行**

```bash
go run main.go
```

Expected: 服务在 8080 端口启动

**Step 4: 停止服务并创建开发分支**

```bash
git checkout -b feature/multi-program-support
```

---

## 阶段一：数据库层

### Task 1: 创建 Program 模型

**目标:** 实现程序元数据模型和数据库表。

**Files:**
- Create: `internal/models/program.go`

**Step 1: 编写 Program 模型**

```go
package models

import (
	"time"
	"gorm.io/gorm"
)

type Program struct {
	ID          uint      `gorm:"primaryKey"`
	ProgramID   string    `gorm:"uniqueIndex;size:50;not null"`
	Name        string    `gorm:"size:100;not null"`
	Description string    `gorm:"size:500"`
	IconURL     string    `gorm:"size:255"`
	IsActive    bool      `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
```

**Step 2: 更新数据库迁移**

**File:** `internal/database/database.go`

在 `AutoMigrate` 中添加 `&models.Program{}`:

```go
func NewGORM(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	err = db.AutoMigrate(
		&models.Version{},
		&models.Program{},
	)
	return db, err
}
```

**Step 3: 运行测试验证表创建**

```bash
go run main.go
```

Expected: 服务启动，`data/versions.db` 中包含 `programs` 表

**Step 4: 验证表结构**

```bash
sqlite3 data/versions.db ".schema programs"
```

Expected: 显示 programs 表的 CREATE 语句

**Step 5: 提交**

```bash
git add internal/models/program.go internal/database/database.go
git commit -m "feat: 添加 Program 模型支持多程序管理"
```

---

### Task 2: 创建 Token 模型

**目标:** 实现权限管理模型。

**Files:**
- Create: `internal/models/token.go`

**Step 1: 编写 Token 模型**

```go
package models

import (
	"time"
	"gorm.io/gorm"
)

type Token struct {
	ID         uint      `gorm:"primaryKey"`
	TokenID    string    `gorm:"uniqueIndex;size:64;not null"`
	ProgramID  string    `gorm:"index;size:50;not null"`
	TokenType  string    `gorm:"size:20;not null"` // admin, upload, download
	CreatedBy  string    `gorm:"size:100"`
	ExpiresAt  *time.Time
	IsActive   bool      `gorm:"default:true"`
	CreatedAt  time.Time
	LastUsedAt *time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// TableName 指定表名
func (Token) TableName() string {
	return "tokens"
}
```

**Step 2: 更新数据库迁移**

**File:** `internal/database/database.go`

```go
err = db.AutoMigrate(
	&models.Version{},
	&models.Program{},
	&models.Token{},
)
```

**Step 3: 运行测试**

```bash
go run main.go
```

Expected: 服务启动，`tokens` 表创建成功

**Step 4: 提交**

```bash
git add internal/models/token.go internal/database/database.go
git commit -m "feat: 添加 Token 模型支持权限管理"
```

---

### Task 3: 扩展 Version 模型

**目标:** 为 Version 模型添加 ProgramID 字段，支持多程序。

**Files:**
- Modify: `internal/models/version.go`

**Step 1: 读取现有模型**

```bash
cat internal/models/version.go
```

**Step 2: 添加 ProgramID 字段**

**File:** `internal/models/version.go`

在现有字段后添加:

```go
type Version struct {
	ID           uint      `gorm:"primaryKey"`
	ProgramID    string    `gorm:"index;size:50;not null"` // 新增
	Version      string    `gorm:"size:20;not null"`
	Channel      string    `gorm:"size:20;index;not null"`
	// ... 其余字段保持不变
}
```

**Step 3: 添加复合唯一索引**

在 Version 结构体后添加:

```go
// BeforeCreate GORM hook
func (v *Version) BeforeCreate(tx *gorm.DB) error {
	// 确保 ProgramID 有默认值
	if v.ProgramID == "" {
		v.ProgramID = "docufiller" // 向后兼容
	}
	return nil
}
```

**Step 4: 运行测试验证迁移**

```bash
go run main.go
```

Expected: 服务启动，versions 表添加 program_id 列

**Step 5: 提交**

```bash
git add internal/models/version.go
git commit -m "feat: Version 模型添加 ProgramID 字段支持多程序"
```

---

## 阶段二：加密服务

### Task 4: 实现加密服务核心

**目标:** 实现 AES-256-GCM 加密解密功能。

**Files:**
- Create: `internal/service/crypto.go`

**Step 1: 编写加密服务**

```go
package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

type CryptoService struct {
	masterKey []byte
}

type EncryptedData struct {
	Encrypted bool   `json:"encrypted"`
	Algorithm string `json:"algorithm"`
	IV        string `json:"iv"`
	Ciphertext string `json:"ciphertext"`
	Tag       string `json:"tag"`
}

func NewCryptoService(masterKey string) *CryptoService {
	return &CryptoService{
		masterKey: []byte(masterKey),
	}
}

// DeriveKey 使用 HKDF 从 masterKey 派生程序专用密钥
func (s *CryptoService) DeriveKey(programID string) ([]byte, error) {
	salt := []byte("update-server-salt-" + programID)
	hash := sha256.New

	hkdf := hkdf.New(hash, s.masterKey, salt, nil)

	key := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(hkdf, key); err != nil {
		return nil, err
	}

	return key, nil
}

// Encrypt 加密数据
func (s *CryptoService) Encrypt(plaintext []byte, programID string) (*EncryptedData, error) {
	key, err := s.DeriveKey(programID)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// 分离 nonce 和 ciphertext
	iv := base64.StdEncoding.EncodeToString(nonce)
	tagAndCiphertext := ciphertext[gcm.NonceSize():]

	return &EncryptedData{
		Encrypted:  true,
		Algorithm:  "AES-256-GCM",
		IV:         iv,
		Ciphertext: base64.StdEncoding.EncodeToString(tagAndCiphertext),
	}, nil
}

// Decrypt 解密数据
func (s *CryptoService) Decrypt(data *EncryptedData, programID string) ([]byte, error) {
	if !data.Encrypted || data.Algorithm != "AES-256-GCM" {
		return nil, errors.New("invalid encrypted data format")
	}

	key, err := s.DeriveKey(programID)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	iv, err := base64.StdEncoding.DecodeString(data.IV)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(data.Ciphertext)
	if err != nil {
		return nil, err
	}

	// 重组 nonce 和 ciphertext
	combined := append(iv, ciphertext...)

	plaintext, err := gcm.Open(nil, combined[:gcm.NonceSize()], combined[gcm.NonceSize():], nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
```

**Step 2: 添加 golang.org/x/crypto 依赖**

```bash
go get golang.org/x/crypto
```

**Step 3: 创建加密测试**

**File:** `internal/service/crypto_test.go`

```go
package service

import (
	"testing"
)

func TestCryptoService_EncryptDecrypt(t *testing.T) {
	masterKey := "test-master-key-32-bytes-long-!!"
	cryptoSvc := NewCryptoService(masterKey)

	plaintext := []byte("Hello, World!")
	programID := "test-program"

	// 测试加密
	encrypted, err := cryptoSvc.Encrypt(plaintext, programID)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if !encrypted.Encrypted {
		t.Error("Encrypted flag should be true")
	}

	// 测试解密
	decrypted, err := cryptoSvc.Decrypt(encrypted, programID)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text mismatch: got %s, want %s", decrypted, plaintext)
	}
}
```

**Step 4: 运行测试**

```bash
go test ./internal/service/crypto_test.go -v
```

Expected: 所有测试通过

**Step 5: 提交**

```bash
git add internal/service/crypto.go internal/service/crypto_test.go go.mod go.sum
git commit -m "feat: 实现 AES-256-GCM 加密服务"
```

---

### Task 5: 创建加密中间件

**目标:** 实现 Gin 中间件处理加密请求/响应。

**Files:**
- Create: `internal/middleware/crypto.go`

**Step 1: 编写加密中间件**

```go
package middleware

import (
	"bytes"
	"docufiller-update-server/internal/service"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

type CryptoMiddleware struct {
	cryptoSvc *service.CryptoService
}

func NewCryptoMiddleware(cryptoSvc *service.CryptoService) *CryptoMiddleware {
	return &CryptoMiddleware{
		cryptoSvc: cryptoSvc,
	}
}

func (m *CryptoMiddleware) Process() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 检查是否为加密请求
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body.Close()
		}

		// 尝试解析为加密格式
		var encryptedData service.EncryptedData
		isEncrypted := json.Unmarshal(bodyBytes, &encryptedData) == nil && encryptedData.Encrypted

		programID := c.Param("programId")
		if programID == "" {
			programID = c.Query("programId")
		}

		// 2. 如果是加密请求，解密
		if isEncrypted && programID != "" {
			plaintext, err := m.cryptoSvc.Decrypt(&encryptedData, programID)
			if err != nil {
				c.JSON(400, gin.H{"error": "decryption failed"})
				c.Abort()
				return
			}

			// 替换请求体
			c.Request.Body = io.NopCloser(bytes.NewReader(plaintext))
		}

		// 3. 继续处理请求
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			cryptoSvc:      m.cryptoSvc,
			programID:      programID,
			shouldEncrypt:  isEncrypted,
		}
		c.Writer = writer

		c.Next()
	}
}

type responseWriter struct {
	gin.ResponseWriter
	cryptoSvc     *service.CryptoService
	programID     string
	shouldEncrypt bool
}

func (w *responseWriter) Write(data []byte) (int, error) {
	if !w.shouldEncrypt || w.programID == "" {
		return w.ResponseWriter.Write(data)
	}

	// 加密响应
	encrypted, err := w.cryptoSvc.Encrypt(data, w.programID)
	if err != nil {
		// 加密失败，返回原始数据
		return w.ResponseWriter.Write(data)
	}

	jsonData, _ := json.Marshal(encrypted)
	return w.ResponseWriter.Write(jsonData)
}
```

**Step 2: 注册中间件**

**File:** `main.go`

在路由设置中添加:

```go
cryptoSvc := service.NewCryptoService(config.Crypto.MasterKey)
cryptoMiddleware := middleware.NewCryptoMiddleware(cryptoSvc)

router.Use(cryptoMiddleware.Process())
```

**Step 3: 更新配置文件**

**File:** `config.yaml`

添加:

```yaml
crypto:
  masterKey: "change-this-to-a-secure-32-byte-key"
```

**Step 4: 更新配置模型**

**File:** `internal/config/config.go`

```go
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	API     APIConfig     `yaml:"api"`
	Storage StorageConfig `yaml:"storage"`
	Logger  LoggerConfig  `yaml:"logger"`
	Crypto  CryptoConfig  `yaml:"crypto"` // 新增
}

type CryptoConfig struct {
	MasterKey string `yaml:"masterKey"`
}
```

**Step 5: 提交**

```bash
git add internal/middleware/crypto.go main.go internal/config/config.go config.yaml
git commit -m "feat: 添加加密中间件支持 API 层加密"
```

---

## 阶段三：认证与授权

### Task 6: 创建 Token 管理服务

**目标:** 实现 Token 的生成、验证、撤销功能。

**Files:**
- Create: `internal/service/token.go`

**Step 1: 编写 Token 服务**

```go
package service

import (
	"crypto/rand"
	"crypto/sha256"
	"docufiller-update-server/internal/models"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TokenService struct {
	db *gorm.DB
}

func NewTokenService(db *gorm.DB) *TokenService {
	return &TokenService{db: db}
}

// GenerateToken 生成新 Token
func (s *TokenService) GenerateToken(programID, tokenType, createdBy string) (*models.Token, string, error) {
	// 生成随机 Token 值
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, "", err
	}
	tokenValue := hex.EncodeToString(randomBytes)

	// 计算哈希
	hash := sha256.Sum256([]byte(tokenValue))
	tokenID := hex.EncodeToString(hash[:])

	token := &models.Token{
		TokenID:    tokenID,
		ProgramID:  programID,
		TokenType:  tokenType,
		CreatedBy:  createdBy,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	if err := s.db.Create(token).Error; err != nil {
		return nil, "", err
	}

	return token, tokenValue, nil
}

// ValidateToken 验证 Token
func (s *TokenService) ValidateToken(tokenValue string) (*models.Token, error) {
	hash := sha256.Sum256([]byte(tokenValue))
	tokenID := hex.EncodeToString(hash[:])

	var token models.Token
	err := s.db.Where("token_id = ? AND is_active = ?", tokenID, true).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid token")
		}
		return nil, err
	}

	// 检查过期时间
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	// 更新最后使用时间
	go s.updateLastUsed(token.ID)

	return &token, nil
}

// HasPermission 检查权限
func (s *TokenService) HasPermission(token *models.Token, requiredType, programID string) bool {
	// Admin Token 拥有所有权限
	if token.TokenType == "admin" {
		return true
	}

	// 检查 Token 类型
	if token.TokenType != requiredType {
		return false
	}

	// 检查程序权限
	if token.ProgramID != "*" && token.ProgramID != programID {
		return false
	}

	return true
}

// RevokeToken 撤销 Token
func (s *TokenService) RevokeToken(tokenID string) error {
	return s.db.Model(&models.Token{}).
		Where("token_id = ?", tokenID).
		Update("is_active", false).Error
}

func (s *TokenService) updateLastUsed(tokenID uint) {
	s.db.Model(&models.Token{}).
		Where("id = ?", tokenID).
		Update("last_used_at", time.Now())
}
```

**Step 2: 创建 Token 测试**

**File:** `internal/service/token_test.go`

```go
package service

import (
	"docufiller-update-server/internal/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	db.AutoMigrate(&models.Token{}, &models.Program{})
	return db
}

func TestTokenService_GenerateAndValidate(t *testing.T) {
	db := setupTestDB(t)
	tokenSvc := NewTokenService(db)

	token, tokenValue, err := tokenSvc.GenerateToken("test-program", "upload", "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token.TokenType != "upload" {
		t.Errorf("TokenType mismatch: got %s, want upload", token.TokenType)
	}

	// 验证 Token
	validated, err := tokenSvc.ValidateToken(tokenValue)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if validated.TokenID != token.TokenID {
		t.Error("TokenID mismatch")
	}
}

func TestTokenService_HasPermission(t *testing.T) {
	db := setupTestDB(t)
	tokenSvc := NewTokenService(db)

	// 测试 Admin Token
	adminToken, _, _ := tokenSvc.GenerateToken("*", "admin", "system")
	if !tokenSvc.HasPermission(adminToken, "upload", "any-program") {
		t.Error("Admin token should have all permissions")
	}

	// 测试 Program Token
	programToken, _, _ := tokenSvc.GenerateToken("my-app", "upload", "admin")
	if !tokenSvc.HasPermission(programToken, "upload", "my-app") {
		t.Error("Program token should have access to own program")
	}

	if tokenSvc.HasPermission(programToken, "upload", "other-app") {
		t.Error("Program token should not have access to other programs")
	}
}
```

**Step 3: 运行测试**

```bash
go test ./internal/service/token_test.go -v
```

Expected: 所有测试通过

**Step 4: 提交**

```bash
git add internal/service/token.go internal/service/token_test.go
git commit -m "feat: 实现 Token 管理服务"
```

---

### Task 7: 创建认证中间件

**目标:** 实现 JWT/Bearer Token 认证中间件。

**Files:**
- Create: `internal/middleware/auth.go`

**Step 1: 编写认证中间件**

```go
package middleware

import (
	"docufiller-update-server/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokenSvc *service.TokenService
}

func NewAuthMiddleware(tokenSvc *service.TokenService) *AuthMiddleware {
	return &AuthMiddleware{tokenSvc: tokenSvc}
}

// RequireAuth 需要认证（任意有效 Token）
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return m.requireAuth("")
}

// RequireAdmin 需要管理员 Token
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.requireAuth("admin")
}

// RequireUpload 需要上传权限（Admin 或 Upload Token）
func (m *AuthMiddleware) RequireUpload() gin.HandlerFunc {
	return m.requireAuthWithProgram("upload")
}

// RequireDownload 需要下载权限
func (m *AuthMiddleware) RequireDownload() gin.HandlerFunc {
	return m.requireAuthWithProgram("download")
}

func (m *AuthMiddleware) requireAuth(requiredType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.JSON(401, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		tokenRecord, err := m.tokenSvc.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// 检查 Token 类型
		if requiredType != "" &&
		   tokenRecord.TokenType != requiredType &&
		   tokenRecord.TokenType != "admin" {
			c.JSON(403, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Set("token", tokenRecord)
		c.Next()
	}
}

func (m *AuthMiddleware) requireAuthWithProgram(requiredType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.JSON(401, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		tokenRecord, err := m.tokenSvc.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		programID := c.Param("programId")
		if !m.tokenSvc.HasPermission(tokenRecord, requiredType, programID) {
			c.JSON(403, gin.H{"error": "program access denied"})
			c.Abort()
			return
		}

		c.Set("token", tokenRecord)
		c.Next()
	}
}

func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// OptionalAuth 可选认证（支持匿名访问）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token != "" {
			if tokenRecord, err := m.tokenSvc.ValidateToken(token); err == nil {
				c.Set("token", tokenRecord)
			}
		}
		c.Next()
	}
}
```

**Step 2: 在 main.go 中注册中间件**

**File:** `main.go`

```go
tokenSvc := service.NewTokenService(db)
authMiddleware := middleware.NewAuthMiddleware(tokenSvc)

// 公开路由
public := router.Group("/api")
{
	public.GET("/health", handlers.HealthCheck)
	public.GET("/programs/:programId/versions/latest", handlers.GetLatestVersion)
	public.GET("/programs/:programId/versions", handlers.GetVersionList)
}

// 认证路由 - 下载
download := router.Group("/api")
download.Use(authMiddleware.RequireDownload())
{
	download.GET("/programs/:programId/download/:channel/:version", handlers.DownloadFile)
}

// 认证路由 - 上传
upload := router.Group("/api")
upload.Use(authMiddleware.RequireUpload())
{
	upload.POST("/programs/:programId/versions", handlers.UploadVersion)
	upload.DELETE("/programs/:programId/versions/:version", handlers.DeleteVersion)
}

// 管理员路由
admin := router.Group("/api")
admin.Use(authMiddleware.RequireAdmin())
{
	admin.POST("/programs", handlers.CreateProgram)
	admin.GET("/programs", handlers.ListPrograms)
	admin.POST("/tokens", handlers.CreateToken)
}
```

**Step 3: 提交**

```bash
git add internal/middleware/auth.go main.go
git commit -m "feat: 添加认证中间件支持 Token 验证"
```

---

## 阶段四：业务逻辑层

### Task 8: 创建程序管理服务

**目标:** 实现程序的 CRUD 操作。

**Files:**
- Create: `internal/service/program.go`

**Step 1: 编写程序服务**

```go
package service

import (
	"docufiller-update-server/internal/models"
	"errors"

	"gorm.io/gorm"
)

type ProgramService struct {
	db *gorm.DB
}

func NewProgramService(db *gorm.DB) *ProgramService {
	return &ProgramService{db: db}
}

// CreateProgram 创建程序
func (s *ProgramService) CreateProgram(program *models.Program) error {
	return s.db.Create(program).Error
}

// GetProgramByID 获取程序
func (s *ProgramService) GetProgramByID(programID string) (*models.Program, error) {
	var program models.Program
	err := s.db.Where("program_id = ?", programID).First(&program).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("program not found")
		}
		return nil, err
	}
	return &program, nil
}

// ListPrograms 列出所有程序
func (s *ProgramService) ListPrograms() ([]models.Program, error) {
	var programs []models.Program
	err := s.db.Find(&programs).Error
	return programs, err
}

// UpdateProgram 更新程序
func (s *ProgramService) UpdateProgram(program *models.Program) error {
	return s.db.Save(program).Error
}

// DeleteProgram 删除程序（软删除）
func (s *ProgramService) DeleteProgram(programID string) error {
	return s.db.Where("program_id = ?", programID).Delete(&models.Program{}).Error
}
```

**Step 2: 创建程序 Handler**

**File:** `internal/handler/program.go`

```go
package handler

import (
	"docufiller-update-server/internal/models"
	"docufiller-update-server/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProgramHandler struct {
	programSvc *service.ProgramService
}

func NewProgramHandler(programSvc *service.ProgramService) *ProgramHandler {
	return &ProgramHandler{programSvc: programSvc}
}

// CreateProgram 创建程序
func (h *ProgramHandler) CreateProgram(c *gin.Context) {
	var req struct {
		ProgramID   string `json:"programId" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IconURL     string `json:"iconUrl"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	program := &models.Program{
		ProgramID:   req.ProgramID,
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconURL,
		IsActive:    true,
	}

	if err := h.programSvc.CreateProgram(program); err != nil {
		c.JSON(500, gin.H{"error": "failed to create program"})
		return
	}

	c.JSON(http.StatusOK, program)
}

// ListPrograms 列出所有程序
func (h *ProgramHandler) ListPrograms(c *gin.Context) {
	programs, err := h.programSvc.ListPrograms()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list programs"})
		return
	}

	c.JSON(http.StatusOK, programs)
}

// GetProgram 获取程序详情
func (h *ProgramHandler) GetProgram(c *gin.Context) {
	programID := c.Param("programId")

	program, err := h.programSvc.GetProgramByID(programID)
	if err != nil {
		c.JSON(404, gin.H{"error": "program not found"})
		return
	}

	c.JSON(http.StatusOK, program)
}
```

**Step 3: 更新路由**

**File:** `main.go`

注册程序管理路由:

```go
programHandler := handler.NewProgramHandler(service.NewProgramService(db))

admin.POST("/programs", programHandler.CreateProgram)
admin.GET("/programs", programHandler.ListPrograms)
admin.GET("/programs/:programId", programHandler.GetProgram)
```

**Step 4: 提交**

```bash
git add internal/service/program.go internal/handler/program.go main.go
git commit -m "feat: 添加程序管理服务和 Handler"
```

---

### Task 9: 更新版本服务支持多程序

**目标:** 扩展现有 VersionService 以支持多程序。

**Files:**
- Modify: `internal/service/version.go`
- Modify: `internal/handler/version.go`

**Step 1: 更新存储服务**

**File:** `internal/service/storage.go`

更新 SaveFile 方法签名:

```go
func (s *StorageService) SaveFile(programID, channel, version string, file io.Reader) (string, int64, string, error) {
	// 创建目录: data/packages/{programID}/{channel}/{version}/
	dir := filepath.Join(s.basePath, programID, channel, version)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", 0, "", err
	}

	// 计算哈希
	hash := sha256.New()
	tee := io.TeeReader(file, hash)

	// 保存文件
	fileName := fmt.Sprintf("%s-%s.zip", programID, version)
	filePath := filepath.Join(dir, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		return "", 0, "", err
	}
	defer f.Close()

	size, err := io.Copy(f, tee)
	if err != nil {
		return "", 0, "", err
	}

	fileHash := hex.EncodeToString(hash.Sum(nil))

	return fileName, size, fileHash, nil
}

func (s *StorageService) GetFilePath(programID, channel, version string) string {
	return filepath.Join(s.basePath, programID, channel, version)
}

func (s *StorageService) DeleteFile(programID, channel, version string) error {
	dir := filepath.Join(s.basePath, programID, channel, version)
	return os.RemoveAll(dir)
}
```

**Step 2: 更新版本服务**

**File:** `internal/service/version.go`

更新所有方法签名，添加 programID 参数:

```go
func (s *VersionService) GetLatestVersion(programID, channel string) (*models.Version, error) {
	var version models.Version
	err := s.db.Where("program_id = ? AND channel = ?", programID, channel).
		Order("publish_date DESC").
		First(&version).Error
	return &version, err
}

func (s *VersionService) CreateVersion(version *models.Version) error {
	// 确保 ProgramID 不为空
	if version.ProgramID == "" {
		version.ProgramID = "docufiller"
	}
	return s.db.Create(version).Error
}

func (s *VersionService) GetVersionList(programID, channel string) ([]models.Version, error) {
	var versions []models.Version
	query := s.db.Where("program_id = ?", programID)
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	err := query.Order("publish_date DESC").Find(&versions).Error
	return versions, err
}
```

**Step 3: 更新版本 Handler**

**File:** `internal/handler/version.go`

更新所有处理函数:

```go
// GetLatestVersion 获取最新版本
func (h *VersionHandler) GetLatestVersion(c *gin.Context) {
	programID := c.Param("programId")
	channel := c.DefaultQuery("channel", "stable")

	version, err := h.versionSvc.GetLatestVersion(programID, channel)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "No version found"})
		} else {
			c.JSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(200, version)
}

// GetVersionList 获取版本列表
func (h *VersionHandler) GetVersionList(c *gin.Context) {
	programID := c.Param("programId")
	channel := c.Query("channel")

	versions, err := h.versionSvc.GetVersionList(programID, channel)
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(200, versions)
}

// UploadVersion 上传新版本
func (h *VersionHandler) UploadVersion(c *gin.Context) {
	programID := c.Param("programId")
	channel := c.PostForm("channel")
	version := c.PostForm("version")
	notes := c.PostForm("notes")
	mandatory, _ := strconv.ParseBool(c.PostForm("mandatory"))

	if channel == "" || version == "" {
		c.JSON(400, gin.H{"error": "channel and version are required"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to process file"})
		return
	}
	defer file.Close()

	fileName, fileSize, fileHash, err := h.versionSvc.GetStorageService().SaveFile(programID, channel, version, file)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	v := &models.Version{
		ProgramID:    programID,
		Version:      version,
		Channel:      channel,
		FileName:     fileName,
		FilePath:     filepath.Join("./data/packages", programID, channel, version),
		FileSize:     fileSize,
		FileHash:     fileHash,
		ReleaseNotes: notes,
		PublishDate:  time.Now(),
		Mandatory:    mandatory,
	}

	if err := h.versionSvc.CreateVersion(v); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Version uploaded successfully", "version": v})
}

// DownloadFile 下载文件
func (h *VersionHandler) DownloadFile(c *gin.Context) {
	programID := c.Param("programId")
	channel := c.Param("channel")
	version := c.Param("version")

	v, err := h.versionSvc.GetVersion(programID, channel, version)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "Version not found"})
		} else {
			c.JSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}

	filePath := h.versionSvc.GetStorageService().GetFilePath(programID, channel, version)
	c.File(filePath)

	go h.versionSvc.IncrementDownloadCount(v.ID)
}
```

**Step 4: 更新路由**

**File:** `main.go`

更新所有路由路径，添加 programId 参数:

```go
// 公开路由
public.GET("/programs/:programId/versions/latest", versionHandler.GetLatestVersion)
public.GET("/programs/:programId/versions", versionHandler.GetVersionList)

// 下载路由
download.GET("/programs/:programId/download/:channel/:version", versionHandler.DownloadFile)

// 上传路由
upload.POST("/programs/:programId/versions", versionHandler.UploadVersion)
upload.DELETE("/programs/:programId/versions/:version", versionHandler.DeleteVersion)
```

**Step 5: 提交**

```bash
git add internal/service/storage.go internal/service/version.go internal/handler/version.go main.go
git commit -m "feat: 更新版本服务支持多程序"
```

---

## 阶段五：向后兼容

### Task 10: 实现向后兼容路由

**目标:** 保留旧 API 端点，映射到 DocuFiller 程序。

**Files:**
- Modify: `main.go`

**Step 1: 添加兼容路由**

**File:** `main.go`

在路由设置中添加:

```go
// 向后兼容路由 - 映射到 docufiller
legacy := router.Group("/api/version")
{
	legacy.GET("/latest", func(c *gin.Context) {
		c.Request.URL.Path = "/api/programs/docufiller/versions/latest"
		router.HandleContext(c)
	})
	legacy.GET("/list", func(c *gin.Context) {
		c.Request.URL.Path = "/api/programs/docufiller/versions"
		router.HandleContext(c)
	})
	legacy.POST("/upload", func(c *gin.Context) {
		c.Request.URL.Path = "/api/programs/docufiller/versions"
		router.HandleContext(c)
	})
}

// 兼容下载路由
router.GET("/api/download/:channel/:version", func(c *gin.Context) {
	c.Request.URL.Path = "/api/programs/docufiller/download/" + c.Param("channel") + "/" + c.Param("version")
	router.HandleContext(c)
})
```

**Step 2: 添加弃用警告中间件**

**File:** `internal/middleware/deprecation.go`

```go
package middleware

import "github.com/gin-gonic/gin"

// DeprecationWarning 添加弃用警告头
func DeprecationWarning() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-API-Deprecation", "This API is deprecated. Use /api/programs/* instead.")
		c.Next()
	}
}
```

**Step 3: 应用弃用警告**

**File:** `main.go`

```go
deprecationMiddleware := middleware.NewDeprecationWarning()

legacy.Use(deprecationMiddleware)
router.GET("/api/download/:channel/:version", deprecationMiddleware.Process(), func(c *gin.Context) {
	// ... 重定向逻辑
})
```

**Step 4: 测试兼容性**

```bash
# 测试旧 API
curl http://localhost:8080/api/version/latest?channel=stable

# 测试新 API
curl http://localhost:8080/api/programs/docufiller/versions/latest?channel=stable
```

Expected: 两者返回相同结果，旧 API 返回 `X-API-Deprecation` 头

**Step 5: 提交**

```bash
git add main.go internal/middleware/deprecation.go
git commit -m "feat: 添加向后兼容路由支持旧 API"
```

---

## 阶段六：数据迁移

### Task 11: 创建数据迁移工具

**目标:** 迁移现有数据到新的多程序结构。

**Files:**
- Create: `scripts/migrate.go`
- Create: `scripts/migrate-storage.sh`

**Step 1: 编写数据迁移脚本**

**File:** `scripts/migrate.go`

```go
package main

import (
	"docufiller-update-server/internal/config"
	"docufiller-update-server/internal/database"
	"docufiller-update-server/internal/models"
	"docufiller-update-server/internal/logger"
	"fmt"
	"os"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化日志
	logger.Init(cfg.Logger)

	// 连接数据库
	db, err := database.NewGORM(cfg.Database.Path)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting migration...")

	// 1. 创建 docufiller 程序记录
	program := &models.Program{
		ProgramID:   "docufiller",
		Name:        "DocuFiller",
		Description: "文档填充工具",
		IsActive:    true,
	}
	if err := db.Create(program).Error; err != nil {
		fmt.Printf("Failed to create program: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Created docufiller program")

	// 2. 更新现有版本记录
	result := db.Model(&models.Version{}).
		Where("program_id IS NULL OR program_id = ''").
		Update("program_id", "docufiller")
	if result.Error != nil {
		fmt.Printf("Failed to update versions: %v\n", result.Error)
		os.Exit(1)
	}
	fmt.Printf("✓ Updated %d version records\n", result.RowsAffected)

	// 3. 生成初始 Token
	tokenSvc := service.NewTokenService(db)

	uploadToken, uploadValue, _ := tokenSvc.GenerateToken("docufiller", "upload", "migration")
	fmt.Printf("✓ Generated upload token: %s\n", uploadValue)

	downloadToken, downloadValue, _ := tokenSvc.GenerateToken("docufiller", "download", "migration")
	fmt.Printf("✓ Generated download token: %s\n", downloadValue)

	fmt.Println("\nMigration completed successfully!")
	fmt.Println("\nIMPORTANT: Save these tokens securely:")
	fmt.Printf("Upload Token: %s\n", uploadValue)
	fmt.Printf("Download Token: %s\n", downloadValue)
}
```

**Step 2: 编写存储目录迁移脚本**

**File:** `scripts/migrate-storage.sh`

```bash
#!/bin/bash

echo "Starting storage migration..."

OLD_BASE="./data/packages"
NEW_BASE="./data/packages"

# 创建新目录结构
echo "Creating new directory structure..."
mkdir -p "$NEW_BASE/docufiller/stable"
mkdir -p "$NEW_BASE/docufiller/beta"

# 迁移 stable 版本
if [ -d "$OLD_BASE/stable" ]; then
    echo "Migrating stable versions..."
    find "$OLD_BASE/stable" -type f -print0 | while IFS= read -r -d '' file; do
        filename=$(basename "$file")
        version="${filename%.*}"
        mkdir -p "$NEW_BASE/docufiller/stable/$version"
        mv "$file" "$NEW_BASE/docufiller/stable/$version/"
    done
    rmdir "$OLD_BASE/stable" 2>/dev/null
    echo "✓ Migrated stable versions"
fi

# 迁移 beta 版本
if [ -d "$OLD_BASE/beta" ]; then
    echo "Migrating beta versions..."
    find "$OLD_BASE/beta" -type f -print0 | while IFS= read -r -d '' file; do
        filename=$(basename "$file")
        version="${filename%.*}"
        mkdir -p "$NEW_BASE/docufiller/beta/$version"
        mv "$file" "$NEW_BASE/docufiller/beta/$version/"
    done
    rmdir "$OLD_BASE/beta" 2>/dev/null
    echo "✓ Migrated beta versions"
fi

echo "Storage migration completed!"
```

**Step 3: 运行数据迁移**

```bash
go run scripts/migrate.go
```

Expected: 输出迁移步骤和生成的 Token

**Step 4: 运行存储迁移**

```bash
bash scripts/migrate-storage.sh
```

Expected: 文件目录结构更新

**Step 5: 验证迁移**

```bash
# 检查数据库
sqlite3 data/versions.db "SELECT program_id, COUNT(*) FROM versions GROUP BY program_id;"

# 检查文件结构
ls -R data/packages/docufiller/
```

Expected: 显示 docufiller 程序的版本统计

**Step 6: 提交**

```bash
git add scripts/migrate.go scripts/migrate-storage.sh
git commit -m "feat: 添加数据迁移脚本"
```

---

## 阶段七：测试

### Task 12: 集成测试

**目标:** 端到端测试所有功能。

**Files:**
- Create: `tests/integration_test.go`

**Step 1: 编写集成测试**

**File:** `tests/integration_test.go`

```go
package tests

import (
	"bytes"
	"docufiller-update-server/internal/config"
	"docufiller-update-server/internal/database"
	"docufiller-update-server/internal/handler"
	"docufiller-update-server/internal/logger"
	"docufiller-update-server/internal/middleware"
	"docufiller-update-server/internal/models"
	"docufiller-update-server/internal/service"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

type TestServer struct {
	router      *gin.Engine
	db          *gorm.DB
	adminToken  string
	uploadToken string
	downloadToken string
}

func setupTestServer(t *testing.T) *TestServer {
	gin.SetMode(gin.TestMode)

	// 内存数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	db.AutoMigrate(&models.Program{}, &models.Version{}, &models.Token{})

	// 初始化服务
	tokenSvc := service.NewTokenService(db)
	cryptoSvc := service.NewCryptoService("test-master-key-32-bytes-long!!")
	storageSvc := service.NewStorageService("./test-data")
	versionSvc := service.NewVersionService(db, storageSvc)
	programSvc := service.NewProgramService(db)

	// 生成测试 Token
	_, adminToken, _ := tokenSvc.GenerateToken("*", "admin", "test")
	_, uploadToken, _ := tokenSvc.GenerateToken("testapp", "upload", "admin")
	_, downloadToken, _ := tokenSvc.GenerateToken("testapp", "download", "admin")

	// 创建测试程序
	programSvc.CreateProgram(&models.Program{
		ProgramID: "testapp",
		Name:      "Test Application",
		IsActive:  true,
	})

	// 设置路由
	router := gin.New()
	auth := middleware.NewAuthMiddleware(tokenSvc)
	crypto := middleware.NewCryptoMiddleware(cryptoSvc)

	versionHandler := handler.NewVersionHandler(db)
	programHandler := handler.NewProgramHandler(programSvc)

	// 公开路由
	public := router.Group("/api")
	{
		public.GET("/programs/:programId/versions/latest", versionHandler.GetLatestVersion)
	}

	// 认证路由
	upload := router.Group("/api")
	upload.Use(auth.RequireUpload())
	{
		upload.POST("/programs/:programId/versions", versionHandler.UploadVersion)
	}

	download := router.Group("/api")
	download.Use(auth.RequireDownload())
	{
		download.GET("/programs/:programId/download/:channel/:version", versionHandler.DownloadFile)
	}

	// 管理员路由
	admin := router.Group("/api")
	admin.Use(auth.RequireAdmin())
	{
		admin.POST("/programs", programHandler.CreateProgram)
	}

	return &TestServer{
		router:        router,
		db:            db,
		adminToken:    adminToken,
		uploadToken:   uploadToken,
		downloadToken: downloadToken,
	}
}

func TestIntegration_ProgramCRUD(t *testing.T) {
	server := setupTestServer(t)

	// 创建程序
	body := `{"programId":"newapp","name":"New App","description":"Test app"}`
	req := httptest.NewRequest("POST", "/api/programs", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+server.adminToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 查询程序
	req = httptest.NewRequest("GET", "/api/programs/newapp", nil)
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestIntegration_VersionUpload(t *testing.T) {
	server := setupTestServer(t)

	// 创建测试文件
	content := []byte("test file content")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("channel", "stable")
	writer.WriteField("version", "1.0.0")
	writer.WriteField("notes", "Test version")

	part, _ := writer.CreateFormFile("file", "test.zip")
	io.WriteString(part, string(content))

	writer.Close()

	// 上传版本
	req := httptest.NewRequest("POST", "/api/programs/testapp/versions", body)
	req.Header.Set("Authorization", "Bearer "+server.uploadToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// 查询最新版本
	req = httptest.NewRequest("GET", "/api/programs/testapp/versions/latest?channel=stable", nil)
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var version models.Version
	json.Unmarshal(w.Body.Bytes(), &version)

	if version.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", version.Version)
	}
}

func TestIntegration_TokenPermissions(t *testing.T) {
	server := setupTestServer(t)

	// 测试上传 Token 不能下载
	body := `{"channel":"stable","version":"1.0.0"}`
	req := httptest.NewRequest("GET", "/api/programs/testapp/download/stable/1.0.0", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+server.uploadToken)

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}
```

**Step 2: 运行集成测试**

```bash
go test ./tests/integration_test.go -v
```

Expected: 所有测试通过

**Step 3: 清理测试数据**

```bash
rm -rf test-data
```

**Step 4: 提交**

```bash
git add tests/integration_test.go
git commit -m "test: 添加集成测试"
```

---

## 阶段八：文档

### Task 13: 编写 API 文档

**目标:** 创建完整的 API 使用文档。

**Files:**
- Create: `docs/api-documentation.md`

**Step 1: 编写 API 文档**

**File:** `docs/api-documentation.md`

```markdown
# 多程序更新服务器 API 文档

## 概述

本文档描述了多程序自动更新服务器的 RESTful API。

## 认证

所有非公开 API 需要使用 Bearer Token 认证:

\`\`\`
Authorization: Bearer <your-token>
\`\`\`

## API 端点

### 程序管理

#### 创建程序

\`\`\`
POST /api/programs
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "programId": "myapp",
  "name": "My Application",
  "description": "应用描述",
  "iconUrl": "https://example.com/icon.png"
}
\`\`\`

#### 列出程序

\`\`\`
GET /api/programs
Authorization: Bearer <admin-token>
\`\`\`

#### 获取程序详情

\`\`\`
GET /api/programs/:programId
\`\`\`

### 版本管理

#### 获取最新版本

\`\`\`
GET /api/programs/:programId/versions/latest?channel=stable
\`\`\`

响应:
\`\`\`json
{
  "id": 1,
  "programId": "myapp",
  "version": "1.0.0",
  "channel": "stable",
  "fileName": "myapp-1.0.0.zip",
  "fileSize": 1024000,
  "fileHash": "abc123...",
  "releaseNotes": "更新内容",
  "publishDate": "2026-01-16T10:00:00Z",
  "mandatory": false,
  "downloadCount": 42
}
\`\`\`

#### 上传版本

\`\`\`
POST /api/programs/:programId/versions
Authorization: Bearer <upload-token>
Content-Type: multipart/form-data

channel: stable
version: 1.0.0
notes: 更新说明
mandatory: false
file: <文件>
\`\`\`

#### 下载版本

\`\`\`
GET /api/programs/:programId/download/:channel/:version
Authorization: Bearer <download-token>
\`\`\`

### Token 管理

#### 生成 Token

\`\`\`
POST /api/tokens
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "programId": "myapp",
  "tokenType": "upload",
  "createdBy": "admin"
}
\`\`\`

#### 列出 Token

\`\`\`
GET /api/tokens
Authorization: Bearer <admin-token>
\`\`\`

#### 撤销 Token

\`\`\`
DELETE /api/tokens/:id
Authorization: Bearer <admin-token>
\`\`\`

## 加密

支持 API 层加密，请求格式:

\`\`\`json
{
  "encrypted": true,
  "algorithm": "AES-256-GCM",
  "iv": "base64编码的IV",
  "ciphertext": "base64编码的密文"
}
\`\`\`

## 错误响应

\`\`\`json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "错误描述",
    "requestId": "req-abc123"
  }
}
\`\`\`

## 向后兼容

旧 API 端点仍然可用，但会返回弃用警告:

- \`GET /api/version/latest\` → \`GET /api/programs/docufiller/versions/latest\`
- \`POST /api/version/upload\` → \`POST /api/programs/docufiller/versions\`
- \`GET /api/download/:channel/:version\` → 保持不变
\`\`\`

---

## 完成清单

在部署前，确保:

- [ ] 所有测试通过
- [ ] 数据迁移成功
- [ ] 存储目录迁移成功
- [ ] 新旧 API 均可访问
- [ ] 加密功能正常
- [ ] Token 权限控制正确
- [ ] 日志记录正常
- [ ] 配置文件更新（MasterKey）
- [ ] 生产环境 Token 已生成并安全存储

---

**实施计划完成！**

使用 \`superpowers:executing-plans\` 技能按步骤执行此计划。
