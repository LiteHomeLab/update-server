package models

import (
	"time"

	"gorm.io/gorm"
)

// Program 程序元数据模型
type Program struct {
	ID          uint           `gorm:"primaryKey"`
	ProgramID   string         `gorm:"uniqueIndex;size:50;not null" json:"programId"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	IconURL     string         `gorm:"size:255" json:"iconUrl"`
	IsActive    bool           `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Program) TableName() string {
	return "programs"
}
