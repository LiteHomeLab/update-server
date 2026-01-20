package models

import (
	"time"

	"gorm.io/gorm"
)

type EncryptionKey struct {
	ID        uint           `gorm:"primaryKey"`
	ProgramID string         `gorm:"uniqueIndex;size:50;not null"`
	KeyData   string         `gorm:"size:255;not null"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (EncryptionKey) TableName() string {
	return "encryption_keys"
}
