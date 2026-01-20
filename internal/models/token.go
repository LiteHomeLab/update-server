package models

import (
	"time"

	"gorm.io/gorm"
)

// Token 权限管理模型
type Token struct {
	ID          uint           `gorm:"primaryKey"`
	TokenID     string         `gorm:"uniqueIndex;size:64;not null" json:"tokenId"`
	TokenValue  string         `gorm:"size:128;not null" json:"-"` // 不 JSON 序列化，存储实际的 Token 值
	ProgramID   string         `gorm:"index;size:50;not null" json:"programId"`
	TokenType   string         `gorm:"size:20;not null" json:"tokenType"` // admin, upload, download
	CreatedBy   string         `gorm:"size:100" json:"createdBy"`
	ExpiresAt   *time.Time     `json:"expiresAt"`
	IsActive    bool           `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time      `json:"createdAt"`
	LastUsedAt  *time.Time     `json:"lastUsedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Token) TableName() string {
	return "tokens"
}
