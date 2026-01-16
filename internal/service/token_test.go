package service

import (
	"docufiller-update-server/internal/models"
	"testing"
	"time"

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

func TestTokenService_RevokeToken(t *testing.T) {
	db := setupTestDB(t)
	tokenSvc := NewTokenService(db)

	token, tokenValue, _ := tokenSvc.GenerateToken("test-program", "upload", "admin")

	// 验证 Token 有效
	_, err := tokenSvc.ValidateToken(tokenValue)
	if err != nil {
		t.Fatalf("Token should be valid before revocation: %v", err)
	}

	// 撤销 Token
	err = tokenSvc.RevokeToken(token.TokenID)
	if err != nil {
		t.Fatalf("RevokeToken failed: %v", err)
	}

	// 验证 Token 已失效
	_, err = tokenSvc.ValidateToken(tokenValue)
	if err == nil {
		t.Error("Token should be invalid after revocation")
	}
}

func TestTokenService_ValidateInvalidToken(t *testing.T) {
	db := setupTestDB(t)
	tokenSvc := NewTokenService(db)

	// 测试无效 Token
	_, err := tokenSvc.ValidateToken("invalid-token-value")
	if err == nil {
		t.Error("Should return error for invalid token")
	}

	if err.Error() != "invalid token" {
		t.Errorf("Error message mismatch: got %s, want 'invalid token'", err.Error())
	}
}

func TestTokenService_ExpiredToken(t *testing.T) {
	db := setupTestDB(t)
	tokenSvc := NewTokenService(db)

	// 创建一个已过期的 Token
	expiredTime := time.Now().Add(-1 * time.Hour)
	token, tokenValue, _ := tokenSvc.GenerateToken("test-program", "upload", "admin")

	// 手动设置过期时间
	db.Model(&token).Update("expires_at", expiredTime)

	// 验证过期的 Token
	_, err := tokenSvc.ValidateToken(tokenValue)
	if err == nil {
		t.Error("Should return error for expired token")
	}

	if err.Error() != "token expired" {
		t.Errorf("Error message mismatch: got %s, want 'token expired'", err.Error())
	}
}
