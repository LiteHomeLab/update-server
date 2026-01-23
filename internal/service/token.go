package service

import (
	"crypto/rand"
	"crypto/sha256"
	"docufiller-update-server/internal/models"
	"encoding/hex"
	"errors"
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
	return s.GenerateTokenWithDB(s.db, programID, tokenType, createdBy)
}

// GenerateTokenWithDB 使用指定数据库实例生成新 Token（支持事务上下文）
func (s *TokenService) GenerateTokenWithDB(db *gorm.DB, programID, tokenType, createdBy string) (*models.Token, string, error) {
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
		TokenValue: tokenValue,
		ProgramID:  programID,
		TokenType:  tokenType,
		CreatedBy:  createdBy,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	if err := db.Create(token).Error; err != nil {
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

// RegenerateToken 重新生成指定类型的 Token
func (s *TokenService) RegenerateToken(programID, tokenType, createdBy string) (*models.Token, string, error) {
	// 先撤销旧的 Token
	if err := s.RevokeTokenByType(programID, tokenType); err != nil {
		return nil, "", err
	}

	// 生成新 Token
	return s.GenerateToken(programID, tokenType, createdBy)
}

// RevokeTokenByType 根据类型撤销 Token
func (s *TokenService) RevokeTokenByType(programID, tokenType string) error {
	return s.db.Model(&models.Token{}).
		Where("program_id = ? AND token_type = ?", programID, tokenType).
		Update("is_active", false).Error
}

// GetToken 获取指定程序的 Token
func (s *TokenService) GetToken(programID, tokenType, createdBy string) (*models.Token, error) {
	var token models.Token
	err := s.db.Where("program_id = ? AND token_type = ? AND is_active = ?", programID, tokenType, true).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有找到 Token，创建一个新的
			newToken, _, err := s.GenerateToken(programID, tokenType, createdBy)
			if err != nil {
				return nil, err
			}
			return newToken, nil
		}
		return nil, err
	}
	return &token, nil
}

func (s *TokenService) updateLastUsed(tokenID uint) {
	s.db.Model(&models.Token{}).
		Where("id = ?", tokenID).
		Update("last_used_at", time.Now())
}
